package tools

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"product-requirements-management/internal/jsonrpc"
	"product-requirements-management/internal/mcp/types"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// UserStoryHandler handles MCP tool calls for User Story operations
type UserStoryHandler struct {
	userStoryService service.UserStoryService
	epicService      service.EpicService
}

// NewUserStoryHandler creates a new UserStoryHandler instance
func NewUserStoryHandler(userStoryService service.UserStoryService, epicService service.EpicService) *UserStoryHandler {
	return &UserStoryHandler{
		userStoryService: userStoryService,
		epicService:      epicService,
	}
}

// GetSupportedTools returns the list of tools this handler supports
func (h *UserStoryHandler) GetSupportedTools() []string {
	return []string{
		"create_user_story",
		"update_user_story",
	}
}

// HandleTool processes a specific tool call for User Story operations
func (h *UserStoryHandler) HandleTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	switch toolName {
	case "create_user_story":
		return h.Create(ctx, args)
	case "update_user_story":
		return h.Update(ctx, args)
	default:
		return nil, jsonrpc.NewMethodNotFoundError(fmt.Sprintf("Unknown tool: %s", toolName))
	}
}

// Create handles the create_user_story tool
func (h *UserStoryHandler) Create(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	if err := validateRequiredArgs(args, []string{"title", "epic_id", "priority"}); err != nil {
		return nil, err
	}

	title, _ := getStringArg(args, "title")
	if title == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'title' argument")
	}

	epicIDStr, _ := getStringArg(args, "epic_id")
	if epicIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'epic_id' argument")
	}

	// Parse epic ID (UUID or reference ID)
	epicID, err := parseUUIDOrReferenceID(epicIDStr, func(refID string) (interface{}, error) {
		return h.epicService.GetEpicByReferenceID(refID)
	})
	if err != nil {
		return nil, jsonrpc.NewInvalidParamsError("Invalid 'epic_id': not a valid UUID or reference ID")
	}

	priorityInt, ok := getIntArg(args, "priority")
	if !ok {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'priority' argument")
	}
	priority := models.Priority(priorityInt)

	// Optional arguments
	description, _ := getStringArg(args, "description")
	var assigneeID *uuid.UUID
	if assigneeIDStr, ok := getStringArg(args, "assignee_id"); ok && assigneeIDStr != "" {
		parsed, err := uuid.Parse(assigneeIDStr)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'assignee_id' format")
		}
		assigneeID = &parsed
	}

	// Create the user story
	req := service.CreateUserStoryRequest{
		EpicID:      epicID,
		CreatorID:   user.ID,
		AssigneeID:  assigneeID,
		Priority:    priority,
		Title:       title,
		Description: &description,
	}

	userStory, err := h.userStoryService.CreateUserStory(req)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to create user story: %v", err))
	}

	message := fmt.Sprintf("Successfully created user story %s: %s", userStory.ReferenceID, userStory.Title)
	return types.CreateDataResponse(message, userStory), nil
}

// Update handles the update_user_story tool
func (h *UserStoryHandler) Update(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Validate required arguments
	if err := validateRequiredArgs(args, []string{"user_story_id"}); err != nil {
		return nil, err
	}

	userStoryIDStr, _ := getStringArg(args, "user_story_id")
	if userStoryIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'user_story_id' argument")
	}

	// Parse user story ID (UUID or reference ID)
	userStoryID, err := parseUUIDOrReferenceID(userStoryIDStr, func(refID string) (interface{}, error) {
		return h.userStoryService.GetUserStoryByReferenceID(refID)
	})
	if err != nil {
		return nil, jsonrpc.NewInvalidParamsError("Invalid 'user_story_id': not a valid UUID or reference ID")
	}

	// Build update request from optional arguments
	req := service.UpdateUserStoryRequest{}

	if title, ok := getStringArg(args, "title"); ok && title != "" {
		req.Title = &title
	}

	if description, ok := getStringArg(args, "description"); ok {
		req.Description = &description
	}

	if priorityInt, ok := getIntArg(args, "priority"); ok {
		priority := models.Priority(priorityInt)
		req.Priority = &priority
	}

	if assigneeIDStr, ok := getStringArg(args, "assignee_id"); ok {
		if assigneeIDStr == "" {
			// Empty string means unassign
			req.AssigneeID = nil
		} else {
			assigneeID, err := uuid.Parse(assigneeIDStr)
			if err != nil {
				return nil, jsonrpc.NewInvalidParamsError("Invalid 'assignee_id' format")
			}
			req.AssigneeID = &assigneeID
		}
	}

	// Update the user story
	userStory, err := h.userStoryService.UpdateUserStory(userStoryID, req)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to update user story: %v", err))
	}

	message := fmt.Sprintf("Successfully updated user story %s: %s", userStory.ReferenceID, userStory.Title)
	return types.CreateDataResponse(message, userStory), nil
}
