package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"product-requirements-management/internal/auth"
	"product-requirements-management/internal/jsonrpc"
	"product-requirements-management/internal/logger"
	"product-requirements-management/internal/mcp/schemas"
	"product-requirements-management/internal/mcp/tools"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// MCPHandler handles MCP (Model Context Protocol) requests
type MCPHandler struct {
	processor         *jsonrpc.Processor
	resourceHandler   *ResourceHandler
	toolsHandler      *tools.Handler
	promptsHandler    *PromptsHandler
	initializeHandler *InitializeHandler
	mcpLogger         *MCPLogger
	errorMapper       *jsonrpc.ErrorMapper
	resourceService   service.ResourceService
}

// NewMCPHandler creates a new MCP handler instance
func NewMCPHandler(
	epicService service.EpicService,
	userStoryService service.UserStoryService,
	requirementService service.RequirementService,
	acceptanceCriteriaService service.AcceptanceCriteriaService,
	searchService service.SearchServiceInterface,
	steeringDocumentService service.SteeringDocumentService,
	promptService *service.PromptService,
	resourceService service.ResourceService,
	requirementTypeRepo repository.RequirementTypeRepository,
) *MCPHandler {
	processor := jsonrpc.NewProcessor()
	resourceHandler := NewResourceHandler(epicService, userStoryService, requirementService, acceptanceCriteriaService, promptService, requirementTypeRepo)
	toolsHandler := tools.NewHandler(epicService, userStoryService, requirementService, acceptanceCriteriaService, searchService, steeringDocumentService, promptService)
	promptsHandler := NewPromptsHandler(promptService, epicService, userStoryService, requirementService, acceptanceCriteriaService, logger.Logger)
	initializeHandler := NewInitializeHandler(toolsHandler, promptsHandler, promptService, logger.Logger)
	mcpLogger := NewMCPLogger()
	errorMapper := jsonrpc.NewErrorMapper()

	// Set the MCP logger for the JSON-RPC processor to enable request/response body logging
	processor.SetLogger(mcpLogger)

	// Create handler instance
	handler := &MCPHandler{
		processor:         processor,
		resourceHandler:   resourceHandler,
		toolsHandler:      toolsHandler,
		promptsHandler:    promptsHandler,
		initializeHandler: initializeHandler,
		mcpLogger:         mcpLogger,
		errorMapper:       errorMapper,
		resourceService:   resourceService,
	}

	// Register MCP methods with enhanced error handling
	processor.RegisterHandler("initialize", handler.wrapHandler("initialize", handler.initializeHandler.HandleInitializeFromParams))
	processor.RegisterHandler("ping", handler.wrapHandler("ping", handlePing))
	processor.RegisterHandler("tools/list", handler.wrapHandler("tools/list", handleToolsList))
	processor.RegisterHandler("tools/call", handler.wrapHandler("tools/call", toolsHandler.HandleToolsCall))
	processor.RegisterHandler("resources/list", handler.wrapHandler("resources/list", handler.handleResourcesList))
	processor.RegisterHandler("resources/read", handler.wrapHandler("resources/read", resourceHandler.HandleResourcesRead))
	processor.RegisterHandler("prompts/list", handler.wrapHandler("prompts/list", promptsHandler.HandlePromptsList))
	processor.RegisterHandler("prompts/get", handler.wrapHandler("prompts/get", promptsHandler.HandlePromptsGet))

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

		// Log the error response body even for read errors
		if errorBody, marshalErr := json.Marshal(jsonrpcErr); marshalErr == nil {
			h.mcpLogger.LogResponseBody(ctx, "unknown", errorBody, user)
		}

		c.JSON(http.StatusBadRequest, jsonrpcErr)
		return
	}

	// Extract method from request for better logging
	method := h.extractMethodFromRequestBody(body)

	// Log the request body for debugging (always log, regardless of success/failure)
	h.mcpLogger.LogRequestBody(ctx, method, body, user)

	// Process the JSON-RPC request
	responseData, err := h.processor.ProcessRequest(ctx, body)
	duration := time.Since(startTime)

	// Check if the response contains an error (JSON-RPC processor returns responseData even for errors)
	isError := h.isErrorResponse(responseData)

	if err != nil {
		h.mcpLogger.LogError(ctx, "request_processing", err, user)
		h.mcpLogger.LogResponse(ctx, method, false, duration, user)

		// Map the error to appropriate JSON-RPC error
		jsonrpcErr := h.errorMapper.MapError(err)

		// Log the error response body
		if errorBody, marshalErr := json.Marshal(jsonrpcErr); marshalErr == nil {
			h.mcpLogger.LogResponseBody(ctx, method, errorBody, user)
		}

		c.JSON(http.StatusInternalServerError, jsonrpcErr)
		return
	}

	// Log processing result (success or error based on response content)
	h.mcpLogger.LogResponse(ctx, method, !isError, duration, user)

	// Always log the response body for debugging (both success and error responses)
	if len(responseData) > 0 {
		h.mcpLogger.LogResponseBody(ctx, method, responseData, user)
	}

	// Log performance metrics
	h.mcpLogger.LogPerformanceMetrics(ctx, "mcp_request", duration, user, map[string]interface{}{
		"request_size_bytes":  len(body),
		"response_size_bytes": len(responseData),
	})

	// If responseData is nil or empty, it was a notification (no response expected)
	if len(responseData) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	// Return the JSON-RPC response
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", responseData)
}

// MCP method handlers

// handlePing handles the MCP ping method for connectivity testing
func handlePing(ctx context.Context, params interface{}) (interface{}, error) {
	// Ping method can optionally accept parameters but typically returns empty object
	// According to MCP spec, ping is used to test connectivity and server responsiveness
	return map[string]interface{}{}, nil
}

// handleToolsList handles the tools/list method
func handleToolsList(ctx context.Context, params interface{}) (interface{}, error) {
	// Return comprehensive tool definitions
	tools := schemas.GetSupportedTools()
	return map[string]interface{}{
		"tools": tools,
	}, nil
}

// handleResourcesList handles the resources/list method with comprehensive error handling
func (h *MCPHandler) handleResourcesList(ctx context.Context, params interface{}) (interface{}, error) {
	// Add timeout handling for resource list operations
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Extract correlation ID for logging
	correlationID := logger.GetCorrelationID(ctx)

	// Extract user from context for logging and authentication validation
	var user *models.User
	if ginCtx, ok := ctx.Value("gin_context").(*gin.Context); ok {
		user = h.mcpLogger.GetUserFromGinContext(ginCtx)
	}

	// Validate authentication - user should be present for MCP operations
	// Note: For direct method calls (testing), user may be nil
	if user == nil {
		// Check if this is a direct call (no gin context) - allow for testing
		if _, hasGinCtx := ctx.Value("gin_context").(*gin.Context); hasGinCtx {
			h.mcpLogger.LogSecurityEvent(ctx, "resources_list_unauthorized_access", map[string]interface{}{
				"method":         "resources/list",
				"correlation_id": correlationID,
				"error":          "no authenticated user found",
			})
			return nil, h.errorMapper.MapError(errors.New("authentication required"))
		}
		// For direct calls without gin context, proceed without user (testing scenario)
	}

	// Log request with method name and correlation ID
	h.mcpLogger.LogRequest(ctx, "resources/list", params, user)

	// Get resource list from service with timeout context
	resources, err := h.resourceService.GetResourceList(timeoutCtx)
	if err != nil {
		// Handle different types of errors with appropriate logging and mapping
		return h.handleResourcesListError(ctx, err, user, correlationID)
	}

	// Validate resource list is not nil (defensive programming)
	if resources == nil {
		h.mcpLogger.LogError(ctx, "resources/list", errors.New("resource service returned nil resources"), user)
		return nil, h.errorMapper.MapError(errors.New("internal error: invalid resource data"))
	}

	// Create JSON-RPC 2.0 compliant response
	result := map[string]interface{}{
		"resources": resources,
	}

	// Log successful response with resource count
	h.mcpLogger.logger.WithFields(logrus.Fields{
		"method":         "resources/list",
		"correlation_id": correlationID,
		"resource_count": len(resources),
		"user_id":        getUserID(user),
		"component":      "mcp_handler",
		"operation":      "resources_list_success",
	}).Info("Resources list retrieved successfully")

	return result, nil
}

// handleResourcesListError handles different types of errors for resources/list method
func (h *MCPHandler) handleResourcesListError(ctx context.Context, err error, user *models.User, correlationID string) (interface{}, error) {
	// Log error with proper context but without exposing internal details
	h.mcpLogger.LogError(ctx, "resources/list", err, user)

	// Determine error type and create appropriate JSON-RPC error response
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		// Timeout error
		h.mcpLogger.logger.WithFields(logrus.Fields{
			"method":         "resources/list",
			"correlation_id": correlationID,
			"user_id":        getUserID(user),
			"component":      "mcp_handler",
			"operation":      "resources_list_timeout",
			"error_type":     "timeout",
		}).Warn("Resources list operation timed out")

		return nil, h.errorMapper.MapError(errors.New("operation timed out"))

	case errors.Is(err, context.Canceled):
		// Context canceled
		h.mcpLogger.logger.WithFields(logrus.Fields{
			"method":         "resources/list",
			"correlation_id": correlationID,
			"user_id":        getUserID(user),
			"component":      "mcp_handler",
			"operation":      "resources_list_canceled",
			"error_type":     "canceled",
		}).Info("Resources list operation was canceled")

		return nil, h.errorMapper.MapError(errors.New("operation was canceled"))

	case isDatabaseError(err):
		// Database connectivity or query errors
		h.mcpLogger.logger.WithFields(logrus.Fields{
			"method":         "resources/list",
			"correlation_id": correlationID,
			"user_id":        getUserID(user),
			"component":      "mcp_handler",
			"operation":      "resources_list_database_error",
			"error_type":     "database",
		}).Error("Database error during resources list retrieval")

		// Don't expose internal database details to client
		return nil, h.errorMapper.MapError(errors.New("service temporarily unavailable"))

	case isAuthenticationError(err):
		// Authentication/authorization errors
		h.mcpLogger.LogSecurityEvent(ctx, "resources_list_auth_error", map[string]interface{}{
			"method":         "resources/list",
			"correlation_id": correlationID,
			"user_id":        getUserID(user),
			"error_type":     "authentication",
		})

		return nil, h.errorMapper.MapError(err)

	case isValidationError(err):
		// Validation errors (e.g., invalid parameters)
		h.mcpLogger.logger.WithFields(logrus.Fields{
			"method":         "resources/list",
			"correlation_id": correlationID,
			"user_id":        getUserID(user),
			"component":      "mcp_handler",
			"operation":      "resources_list_validation_error",
			"error_type":     "validation",
		}).Warn("Validation error during resources list retrieval")

		return nil, h.errorMapper.MapError(err)

	default:
		// Generic internal errors - don't expose details
		h.mcpLogger.logger.WithFields(logrus.Fields{
			"method":         "resources/list",
			"correlation_id": correlationID,
			"user_id":        getUserID(user),
			"component":      "mcp_handler",
			"operation":      "resources_list_internal_error",
			"error_type":     "internal",
		}).Error("Internal error during resources list retrieval")

		// Return generic error without exposing internal details
		return nil, h.errorMapper.MapError(errors.New("internal server error"))
	}
}

// isDatabaseError checks if the error is related to database operations
func isDatabaseError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	databaseKeywords := []string{
		"database", "connection", "sql", "postgres", "gorm",
		"driver", "network", "timeout", "connection refused",
		"connection reset", "broken pipe", "no such host",
	}

	for _, keyword := range databaseKeywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}

	return false
}

// isAuthenticationError checks if the error is related to authentication
func isAuthenticationError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific authentication errors
	authErrors := []error{
		auth.ErrInvalidCredentials, auth.ErrInvalidToken, auth.ErrTokenExpired, auth.ErrInsufficientRole,
	}

	for _, authErr := range authErrors {
		if errors.Is(err, authErr) {
			return true
		}
	}

	// Check for authentication-related error messages
	errStr := strings.ToLower(err.Error())
	authKeywords := []string{
		"unauthorized", "authentication", "token", "forbidden",
		"access denied", "permission denied", "invalid credentials",
	}

	for _, keyword := range authKeywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}

	return false
}

// isValidationError checks if the error is related to validation
func isValidationError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	validationKeywords := []string{
		"validation", "invalid", "required", "constraint",
		"format", "parse", "decode", "unmarshal",
	}

	for _, keyword := range validationKeywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}

	return false
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
func (h *MCPHandler) logToolCallAuditEvent(ctx context.Context, params any, _ any, user *models.User) {
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
	case strings.Contains(toolName, "steering"):
		return "steering_document", h.extractIDFromParams(params, []string{"steering_document_id", "id"})
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

// getUserID safely extracts user ID from user object for logging
func getUserID(user *models.User) string {
	if user == nil {
		return "anonymous"
	}
	return user.ID.String()
}

// extractMethodFromRequestBody extracts the method name from JSON-RPC request body
func (h *MCPHandler) extractMethodFromRequestBody(body []byte) string {
	var request struct {
		Method string `json:"method"`
	}

	if err := json.Unmarshal(body, &request); err != nil {
		return "unknown"
	}

	if request.Method == "" {
		return "unknown"
	}

	return request.Method
}

// isErrorResponse checks if the JSON-RPC response contains an error
func (h *MCPHandler) isErrorResponse(responseData []byte) bool {
	if len(responseData) == 0 {
		return false
	}

	var response struct {
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(responseData, &response); err != nil {
		return false
	}

	return response.Error != nil
}
