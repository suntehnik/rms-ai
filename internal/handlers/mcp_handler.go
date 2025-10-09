package handlers

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"product-requirements-management/internal/jsonrpc"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"

	"github.com/gin-gonic/gin"
)

// MCPHandler handles MCP (Model Context Protocol) requests
type MCPHandler struct {
	processor       *jsonrpc.Processor
	resourceHandler *ResourceHandler
	toolsHandler    *ToolsHandler
	mcpLogger       *MCPLogger
	errorMapper     *jsonrpc.ErrorMapper
}

// NewMCPHandler creates a new MCP handler instance
func NewMCPHandler(
	epicService service.EpicService,
	userStoryService service.UserStoryService,
	requirementService service.RequirementService,
	acceptanceCriteriaService service.AcceptanceCriteriaService,
	searchService service.SearchServiceInterface,
) *MCPHandler {
	processor := jsonrpc.NewProcessor()
	resourceHandler := NewResourceHandler(epicService, userStoryService, requirementService, acceptanceCriteriaService)
	toolsHandler := NewToolsHandler(epicService, userStoryService, requirementService, searchService)
	mcpLogger := NewMCPLogger()
	errorMapper := jsonrpc.NewErrorMapper()

	// Create handler instance
	handler := &MCPHandler{
		processor:       processor,
		resourceHandler: resourceHandler,
		toolsHandler:    toolsHandler,
		mcpLogger:       mcpLogger,
		errorMapper:     errorMapper,
	}

	// Register MCP methods with enhanced error handling
	processor.RegisterHandler("initialize", handler.wrapHandler("initialize", handleInitialize))
	processor.RegisterHandler("tools/list", handler.wrapHandler("tools/list", handleToolsList))
	processor.RegisterHandler("tools/call", handler.wrapHandler("tools/call", toolsHandler.HandleToolsCall))
	processor.RegisterHandler("resources/read", handler.wrapHandler("resources/read", resourceHandler.HandleResourcesRead))

	return handler
}

// Process handles MCP protocol requests
// @Summary Process MCP request
// @Description Process a Model Context Protocol request using JSON-RPC 2.0
// @Tags mcp
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "MCP request processed successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/mcp [post]
func (h *MCPHandler) Process(c *gin.Context) {
	startTime := time.Now()

	// Extract user information from context
	user := h.mcpLogger.GetUserFromGinContext(c)

	// Create context with correlation ID
	ctx := h.mcpLogger.WithCorrelationID(c.Request.Context())
	ctx = context.WithValue(ctx, "gin_context", c)

	// Log security event for MCP access
	h.mcpLogger.LogSecurityEvent(ctx, "mcp_access_attempt", map[string]interface{}{
		"client_ip":  c.ClientIP(),
		"user_agent": c.GetHeader("User-Agent"),
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
	})

	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.mcpLogger.LogError(ctx, "request_parsing", err, user)

		jsonrpcErr := jsonrpc.NewParseError("Failed to read request body")
		c.JSON(http.StatusBadRequest, jsonrpcErr)
		return
	}

	// Process the JSON-RPC request
	responseData, err := h.processor.ProcessRequest(ctx, body)
	duration := time.Since(startTime)

	if err != nil {
		h.mcpLogger.LogError(ctx, "request_processing", err, user)
		h.mcpLogger.LogResponse(ctx, "unknown", false, duration, user)

		// Map the error to appropriate JSON-RPC error
		jsonrpcErr := h.errorMapper.MapError(err)
		c.JSON(http.StatusInternalServerError, jsonrpcErr)
		return
	}

	// Log successful processing
	h.mcpLogger.LogResponse(ctx, "unknown", true, duration, user)

	// Log performance metrics
	h.mcpLogger.LogPerformanceMetrics(ctx, "mcp_request", duration, user, map[string]interface{}{
		"request_size_bytes":  len(body),
		"response_size_bytes": len(responseData),
	})

	// If responseData is nil, it was a notification (no response expected)
	if responseData == nil {
		c.Status(http.StatusNoContent)
		return
	}

	// Return the JSON-RPC response
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", responseData)
}

// MCP method handlers

// handleInitialize handles the MCP initialize method
func handleInitialize(ctx context.Context, params interface{}) (interface{}, error) {
	// Basic MCP initialization response
	return map[string]interface{}{
		"protocolVersion": "2025-03-26",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": true,
			},
			"resources": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "product-requirements-mcp-server",
			"version": "1.0.0",
		},
	}, nil
}

// handleToolsList handles the tools/list method
func handleToolsList(ctx context.Context, params interface{}) (interface{}, error) {
	// Return comprehensive tool definitions
	tools := GetSupportedTools()
	return map[string]interface{}{
		"tools": tools,
	}, nil
}

// wrapHandler wraps a JSON-RPC handler with enhanced logging and error handling
func (h *MCPHandler) wrapHandler(method string, handler func(context.Context, interface{}) (interface{}, error)) func(context.Context, interface{}) (interface{}, error) {
	return func(ctx context.Context, params interface{}) (interface{}, error) {
		startTime := time.Now()

		// Extract user from context
		var user *models.User
		if ginCtx, ok := ctx.Value("gin_context").(*gin.Context); ok {
			user = h.mcpLogger.GetUserFromGinContext(ginCtx)
		}

		// Log the request
		h.mcpLogger.LogRequest(ctx, method, params, user)

		// Execute the handler
		result, err := handler(ctx, params)
		duration := time.Since(startTime)

		if err != nil {
			// Log the error
			h.mcpLogger.LogError(ctx, method, err, user)
			h.mcpLogger.LogResponse(ctx, method, false, duration, user)

			// Map service layer error to JSON-RPC error
			return nil, h.errorMapper.MapError(err)
		}

		// Log successful response
		h.mcpLogger.LogResponse(ctx, method, true, duration, user)

		// Log audit event for operations that modify data
		if h.isModifyingOperation(method) {
			h.logAuditEventForOperation(ctx, method, params, result, user)
		}

		// Log performance metrics for slow operations
		if duration > 100*time.Millisecond {
			h.mcpLogger.LogPerformanceMetrics(ctx, method, duration, user, map[string]interface{}{
				"slow_operation": true,
			})
		}

		return result, nil
	}
}

// isModifyingOperation checks if the method modifies data
func (h *MCPHandler) isModifyingOperation(method string) bool {
	modifyingMethods := []string{
		"tools/call", // Tool calls can modify data
	}

	for _, modMethod := range modifyingMethods {
		if method == modMethod {
			return true
		}
	}
	return false
}

// logAuditEventForOperation logs audit events for data-modifying operations
func (h *MCPHandler) logAuditEventForOperation(ctx context.Context, method string, params interface{}, result interface{}, user *models.User) {
	switch method {
	case "tools/call":
		h.logToolCallAuditEvent(ctx, params, result, user)
	}
}

// logToolCallAuditEvent logs audit events for tool calls
func (h *MCPHandler) logToolCallAuditEvent(ctx context.Context, params interface{}, result interface{}, user *models.User) {
	if paramsMap, ok := params.(map[string]interface{}); ok {
		if toolName, ok := paramsMap["name"].(string); ok {
			details := map[string]interface{}{
				"tool_name": toolName,
			}

			// Add arguments if present
			if args, ok := paramsMap["arguments"].(map[string]interface{}); ok {
				details["arguments"] = args
			}

			// Determine resource type and ID from tool name and arguments
			resourceType, resourceID := h.extractResourceInfoFromToolCall(toolName, paramsMap)

			h.mcpLogger.LogAuditEvent(ctx, "tool_call", resourceType, resourceID, user, details)
		}
	}
}

// extractResourceInfoFromToolCall extracts resource information from tool call parameters
func (h *MCPHandler) extractResourceInfoFromToolCall(toolName string, params map[string]interface{}) (string, string) {
	// Extract resource type and ID based on tool name
	switch {
	case strings.Contains(toolName, "epic"):
		return "epic", h.extractIDFromParams(params, []string{"epic_id", "id"})
	case strings.Contains(toolName, "user_story"):
		return "user_story", h.extractIDFromParams(params, []string{"user_story_id", "id"})
	case strings.Contains(toolName, "requirement"):
		return "requirement", h.extractIDFromParams(params, []string{"requirement_id", "id"})
	case strings.Contains(toolName, "search"):
		return "search", ""
	default:
		return "unknown", ""
	}
}

// extractIDFromParams extracts ID from parameters using multiple possible key names
func (h *MCPHandler) extractIDFromParams(params map[string]interface{}, possibleKeys []string) string {
	if args, ok := params["arguments"].(map[string]interface{}); ok {
		for _, key := range possibleKeys {
			if id, ok := args[key].(string); ok {
				return id
			}
		}
	}
	return ""
}
