package tools

import (
	"context"
	"fmt"

	"product-requirements-management/internal/jsonrpc"
	"product-requirements-management/internal/mcp/types"
	"product-requirements-management/internal/service"
)

// AcceptanceCriteriaHandler handles MCP tool calls for Acceptance Criteria operations
type AcceptanceCriteriaHandler struct {
	acceptanceCriteriaService service.AcceptanceCriteriaService
	userStoryService          service.UserStoryService
}

// NewAcceptanceCriteriaHandler creates a new AcceptanceCriteriaHandler instance
func NewAcceptanceCriteriaHandler(
	acceptanceCriteriaService service.AcceptanceCriteriaService,
	userStoryService service.UserStoryService,
) *AcceptanceCriteriaHandler {
	return &AcceptanceCriteriaHandler{
		acceptanceCriteriaService: acceptanceCriteriaService,
		userStoryService:          userStoryService,
	}
}

// GetSupportedTools returns the list of tools this handler supports
func (h *AcceptanceCriteriaHandler) GetSupportedTools() []string {
	return []string{
		"create_acceptance_criteria",
	}
}

// HandleTool processes a specific tool call for Acceptance Criteria operations
func (h *AcceptanceCriteriaHandler) HandleTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	switch toolName {
	case "create_acceptance_criteria":
		return h.Create(ctx, args)
	default:
		return nil, jsonrpc.NewMethodNotFoundError(fmt.Sprintf("Unknown tool: %s", toolName))
	}
}

// Create handles the create_acceptance_criteria tool
func (h *AcceptanceCriteriaHandler) Create(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	if err := validateRequiredArgs(args, []string{"user_story_id", "description"}); err != nil {
		return nil, err
	}

	userStoryIDStr, _ := getStringArg(args, "user_story_id")
	if userStoryIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'user_story_id' argument")
	}

	description, _ := getStringArg(args, "description")
	if description == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'description' argument")
	}

	// Validate description field constraints (required, max 50000 characters)
	if len(description) > 50000 {
		return nil, jsonrpc.NewInvalidParamsError("Description exceeds maximum length of 50000 characters")
	}

	// Parse user story ID (UUID or reference ID format)
	userStoryID, err := parseUUIDOrReferenceID(userStoryIDStr, func(refID string) (interface{}, error) {
		return h.userStoryService.GetUserStoryByReferenceID(refID)
	})
	if err != nil {
		return nil, jsonrpc.NewInvalidParamsError("Invalid 'user_story_id': not a valid UUID or reference ID")
	}

	// Create the acceptance criteria using existing AcceptanceCriteriaService
	req := service.CreateAcceptanceCriteriaRequest{
		UserStoryID: userStoryID,
		AuthorID:    user.ID, // Auto-set from authentication context
		Description: description,
	}

	acceptanceCriteria, err := h.acceptanceCriteriaService.CreateAcceptanceCriteria(req)
	if err != nil {
		// Map service errors to appropriate JSON-RPC error codes
		switch err {
		case service.ErrUserStoryNotFound:
			return nil, jsonrpc.NewInvalidParamsError("User story not found")
		case service.ErrUserNotFound:
			return nil, jsonrpc.NewUnauthorizedError("Authentication required")
		default:
			return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to create acceptance criteria: %v", err))
		}
	}

	// Format structured JSON-RPC response with all acceptance criteria fields
	message := fmt.Sprintf("Successfully created acceptance criteria %s", acceptanceCriteria.ReferenceID)
	return types.CreateDataResponse(message, acceptanceCriteria), nil
}
