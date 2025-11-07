package tools

import (
	"context"
	"fmt"
	"product-requirements-management/internal/jsonrpc"
	"product-requirements-management/internal/service"
)

// Handler is the main facade that routes MCP tool calls to appropriate domain handlers
type Handler struct {
	// Domain-specific handlers
	epicHandler               *EpicHandler
	userStoryHandler          *UserStoryHandler
	requirementHandler        *RequirementHandler
	acceptanceCriteriaHandler *AcceptanceCriteriaHandler
	searchHandler             *SearchHandler
	steeringDocumentHandler   *SteeringDocumentHandler
	promptHandler             *PromptHandler

	// Tool routing map for O(1) lookup performance
	toolRoutes map[string]ToolHandler
}

// NewHandler creates a new MCP tools handler with all domain handlers
func NewHandler(
	epicService service.EpicService,
	userService service.UserService,
	userStoryService service.UserStoryService,
	requirementService service.RequirementService,
	acceptanceCriteriaService service.AcceptanceCriteriaService,
	searchService service.SearchServiceInterface,
	steeringDocumentService service.SteeringDocumentService,
	promptService PromptServiceInterface,
) *Handler {
	// Initialize domain handlers
	epicHandler := NewEpicHandler(epicService, userService)
	userStoryHandler := NewUserStoryHandler(userStoryService, epicService, requirementService)
	requirementHandler := NewRequirementHandler(requirementService, userStoryService)
	acceptanceCriteriaHandler := NewAcceptanceCriteriaHandler(acceptanceCriteriaService, userStoryService)
	searchHandler := NewSearchHandler(searchService, requirementService)
	steeringDocumentHandler := NewSteeringDocumentHandler(steeringDocumentService, epicService)
	promptHandler := NewPromptHandler(promptService)

	// Create tool routing map for efficient lookup
	toolRoutes := make(map[string]ToolHandler)

	// Register Epic tools
	for _, tool := range epicHandler.GetSupportedTools() {
		toolRoutes[tool] = epicHandler
	}

	// Register User Story tools
	for _, tool := range userStoryHandler.GetSupportedTools() {
		toolRoutes[tool] = userStoryHandler
	}

	// Register Requirement tools
	for _, tool := range requirementHandler.GetSupportedTools() {
		toolRoutes[tool] = requirementHandler
	}

	// Register Acceptance Criteria tools
	for _, tool := range acceptanceCriteriaHandler.GetSupportedTools() {
		toolRoutes[tool] = acceptanceCriteriaHandler
	}

	// Register Search tools
	for _, tool := range searchHandler.GetSupportedTools() {
		toolRoutes[tool] = searchHandler
	}

	// Register Steering Document tools
	for _, tool := range steeringDocumentHandler.GetSupportedTools() {
		toolRoutes[tool] = steeringDocumentHandler
	}

	// Register Prompt tools
	for _, tool := range promptHandler.GetSupportedTools() {
		toolRoutes[tool] = promptHandler
	}

	return &Handler{
		epicHandler:               epicHandler,
		userStoryHandler:          userStoryHandler,
		requirementHandler:        requirementHandler,
		acceptanceCriteriaHandler: acceptanceCriteriaHandler,
		searchHandler:             searchHandler,
		steeringDocumentHandler:   steeringDocumentHandler,
		promptHandler:             promptHandler,
		toolRoutes:                toolRoutes,
	}
}

// GetAllSupportedTools returns a list of all tools supported by all domain handlers
func (h *Handler) GetAllSupportedTools() []string {
	var allTools []string

	// Collect tools from all handlers
	allTools = append(allTools, h.epicHandler.GetSupportedTools()...)
	allTools = append(allTools, h.userStoryHandler.GetSupportedTools()...)
	allTools = append(allTools, h.requirementHandler.GetSupportedTools()...)
	allTools = append(allTools, h.searchHandler.GetSupportedTools()...)
	allTools = append(allTools, h.steeringDocumentHandler.GetSupportedTools()...)
	allTools = append(allTools, h.promptHandler.GetSupportedTools()...)

	return allTools
}

// GetToolRoutes returns the tool routing map for debugging/inspection purposes
func (h *Handler) GetToolRoutes() map[string]ToolHandler {
	// Return a copy to prevent external modification
	routes := make(map[string]ToolHandler)
	for tool, handler := range h.toolRoutes {
		routes[tool] = handler
	}
	return routes
}

// HandleToolsCall processes tools/call requests by routing to appropriate domain handler
func (h *Handler) HandleToolsCall(ctx context.Context, params interface{}) (interface{}, error) {
	// Extract parameters
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, jsonrpc.NewInvalidParamsError("Invalid parameters format")
	}

	toolName, ok := paramsMap["name"].(string)
	if !ok {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid tool name")
	}

	arguments, _ := paramsMap["arguments"].(map[string]interface{})
	if arguments == nil {
		arguments = make(map[string]interface{})
	}

	// Route to appropriate domain handler
	handler, exists := h.toolRoutes[toolName]
	if !exists {
		return nil, jsonrpc.NewMethodNotFoundError(fmt.Sprintf("Unknown tool: %s", toolName))
	}

	// Delegate to the domain handler
	return handler.HandleTool(ctx, toolName, arguments)
}

// IsToolSupported checks if a tool is supported by any domain handler
func (h *Handler) IsToolSupported(toolName string) bool {
	_, exists := h.toolRoutes[toolName]
	return exists
}

// GetHandlerForTool returns the handler responsible for a specific tool
func (h *Handler) GetHandlerForTool(toolName string) (ToolHandler, bool) {
	handler, exists := h.toolRoutes[toolName]
	return handler, exists
}

// HasTools implements mcp.ToolProvider interface
func (h *Handler) HasTools(ctx context.Context) bool {
	// Always return true as we have tools available
	return true
}

// SupportsListChanged implements mcp.ToolProvider interface
func (h *Handler) SupportsListChanged(ctx context.Context) bool {
	// Return true to indicate we support list change notifications
	return true
}
