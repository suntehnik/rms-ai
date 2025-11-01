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

// EpicHandler handles MCP tools for Epic domain operations
type EpicHandler struct {
	epicService service.EpicService
}

// NewEpicHandler creates a new Epic handler instance
func NewEpicHandler(epicService service.EpicService) *EpicHandler {
	return &EpicHandler{
		epicService: epicService,
	}
}

// GetSupportedTools returns the list of tools this handler supports
func (h *EpicHandler) GetSupportedTools() []string {
	return []string{
		"create_epic",
		"update_epic",
	}
}

// HandleTool processes a specific tool call for Epic domain operations
func (h *EpicHandler) HandleTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	switch toolName {
	case "create_epic":
		return h.Create(ctx, args)
	case "update_epic":
		return h.Update(ctx, args)
	default:
		return nil, jsonrpc.NewMethodNotFoundError(fmt.Sprintf("Unknown Epic tool: %s", toolName))
	}
}

// Create handles the create_epic tool
func (h *EpicHandler) Create(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	title, ok := getStringArg(args, "title")
	if !ok || title == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'title' argument")
	}

	priority, ok := getIntArg(args, "priority")
	if !ok {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'priority' argument")
	}

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

	// Create the epic
	req := service.CreateEpicRequest{
		CreatorID:   user.ID,
		AssigneeID:  assigneeID,
		Priority:    models.Priority(priority),
		Title:       title,
		Description: &description,
	}

	epic, err := h.epicService.CreateEpic(req)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to create epic: %v", err))
	}

	message := fmt.Sprintf("Successfully created epic %s: %s", epic.ReferenceID, epic.Title)
	return types.CreateDataResponse(message, epic), nil
}

// Update handles the update_epic tool
func (h *EpicHandler) Update(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Validate required arguments
	epicIDStr, ok := getStringArg(args, "epic_id")
	if !ok || epicIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'epic_id' argument")
	}

	// Parse epic ID (UUID or reference ID)
	epicID, err := parseUUIDOrReferenceID(epicIDStr, func(refID string) (interface{}, error) {
		return h.epicService.GetEpicByReferenceID(refID)
	})
	if err != nil {
		return nil, jsonrpc.NewInvalidParamsError("Invalid 'epic_id': not a valid UUID or reference ID")
	}

	// Build update request from optional arguments
	req := service.UpdateEpicRequest{}

	if title, ok := getStringArg(args, "title"); ok && title != "" {
		req.Title = &title
	}

	if description, ok := getStringArg(args, "description"); ok {
		req.Description = &description
	}

	if priority, ok := getIntArg(args, "priority"); ok {
		priorityValue := models.Priority(priority)
		req.Priority = &priorityValue
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

	if status, ok := getStringArg(args, "status"); ok && status != "" {
		epicStatus := models.EpicStatus(status)
		req.Status = &epicStatus
	}

	// Update the epic
	epic, err := h.epicService.UpdateEpic(epicID, req)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to update epic: %v", err))
	}

	message := fmt.Sprintf("Successfully updated epic %s: %s", epic.ReferenceID, epic.Title)
	return types.CreateDataResponse(message, epic), nil
}
