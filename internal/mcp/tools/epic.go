package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"product-requirements-management/internal/jsonrpc"
	"product-requirements-management/internal/mcp/types"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
	"product-requirements-management/internal/validation"
)

// EpicHandler handles MCP tools for Epic domain operations
type EpicHandler struct {
	epicService     service.EpicService
	userService     service.UserService
	statusValidator validation.StatusValidator
}

// NewEpicHandler creates a new Epic handler instance
func NewEpicHandler(epicService service.EpicService, userService service.UserService) *EpicHandler {
	return &EpicHandler{
		epicService:     epicService,
		userService:     userService,
		statusValidator: validation.NewStatusValidator(),
	}
}

// GetSupportedTools returns the list of tools this handler supports
func (h *EpicHandler) GetSupportedTools() []string {
	return []string{
		"create_epic",
		"update_epic",
		"list_epics",
	}
}

// HandleTool processes a specific tool call for Epic domain operations
func (h *EpicHandler) HandleTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	switch toolName {
	case "list_epics":
		return h.List(ctx, args)
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
		// Check for status validation errors and provide specific error messages
		if statusErr, ok := validation.GetStatusValidationError(err); ok {
			return nil, jsonrpc.NewInvalidParamsError(statusErr.Message)
		}

		// Check for entity not found errors
		if errors.Is(err, service.ErrEpicNotFound) {
			return nil, jsonrpc.NewInvalidParamsError("Epic not found")
		}

		// Check for other validation errors
		if errors.Is(err, service.ErrInvalidPriority) {
			return nil, jsonrpc.NewInvalidParamsError("Invalid priority value")
		}

		if errors.Is(err, service.ErrUserNotFound) {
			return nil, jsonrpc.NewInvalidParamsError("Assignee user not found")
		}

		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to update epic: %v", err))
	}

	message := fmt.Sprintf("Successfully updated epic %s: %s", epic.ReferenceID, epic.Title)
	return types.CreateDataResponse(message, epic), nil
}

// List handles the list_epic tool
func (h *EpicHandler) List(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	filters := service.EpicFilters{}

	if creatorIDStr, ok := getStringArg(args, "creator_id"); ok && strings.TrimSpace(creatorIDStr) != "" {
		creatorUUID, err := uuid.Parse(creatorIDStr)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'creator_id' format")
		}
		filters.CreatorID = &creatorUUID
	}

	if creatorStr, ok := getStringArg(args, "creator"); ok && strings.TrimSpace(creatorStr) != "" {
		creator := strings.TrimSpace(creatorStr)
		var creatorUUID uuid.UUID

		if strings.EqualFold(creator, "me") {
			user, err := getUserFromContext(ctx)
			if err != nil {
				return nil, jsonrpc.NewInvalidParamsError("Creator 'me' could not be resolved for this request")
			}
			creatorUUID = user.ID
		} else {
			if h.userService == nil {
				return nil, jsonrpc.NewInternalError("User service not configured")
			}
			user, err := h.userService.GetByName(creator)
			if err != nil {
				if errors.Is(err, service.ErrUserNotFound) {
					return nil, jsonrpc.NewInvalidParamsError("Creator user not found")
				}
				return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to resolve creator: %v", err))
			}
			creatorUUID = user.ID
		}

		filters.CreatorID = &creatorUUID
	}

	if assigneeIDStr, ok := getStringArg(args, "assignee"); ok && assigneeIDStr != "" {
		assignee := strings.TrimSpace(assigneeIDStr)
		var assigneeUUID uuid.UUID
		if strings.EqualFold(assignee, "me") {
			user, err := getUserFromContext(ctx)
			if err != nil {
				return nil, jsonrpc.NewInvalidParamsError("Creator 'me' could not be resolved for this request")
			}
			assigneeUUID = user.ID
		} else {
			if h.userService == nil {
				return nil, jsonrpc.NewInternalError("User service not configured")
			}
			user, err := h.userService.GetByName(assignee)
			if err != nil {
				if errors.Is(err, service.ErrUserNotFound) {
					return nil, jsonrpc.NewInvalidParamsError("Creator user not found")
				}
				return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to resolve creator: %v", err))
			}
			assigneeUUID = user.ID
		}
		filters.AssigneeID = &assigneeUUID
	}

	if statusStr, ok := getStringArg(args, "status"); ok && statusStr != "" {
		if err := h.statusValidator.ValidateEpicStatus(statusStr); err != nil {
			return nil, jsonrpc.NewInvalidParamsError(err.Error())
		}
		status := models.EpicStatus(statusStr)
		filters.Status = &status
	}

	if priorityVal, ok := getIntArg(args, "priority"); ok {
		if priorityVal < int(models.PriorityCritical) || priorityVal > int(models.PriorityLow) {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'priority' value")
		}
		priority := models.Priority(priorityVal)
		filters.Priority = &priority
	}

	if includeRaw, exists := args["include"]; exists {
		switch v := includeRaw.(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				filters.Include = parseIncludeList(v)
			}
		case []interface{}:
			includes := make([]string, 0, len(v))
			for _, item := range v {
				if str, ok := item.(string); ok && strings.TrimSpace(str) != "" {
					includes = append(includes, strings.TrimSpace(str))
				}
			}
			if len(includes) > 0 {
				filters.Include = includes
			}
		}
	}

	if orderBy, ok := getStringArg(args, "order_by"); ok && orderBy != "" {
		filters.OrderBy = orderBy
	}

	if limit, ok := getIntArg(args, "limit"); ok {
		if limit <= 0 {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'limit' value")
		}
		filters.Limit = limit
	}

	if offset, ok := getIntArg(args, "offset"); ok {
		if offset < 0 {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'offset' value")
		}
		filters.Offset = offset
	}

	epics, totalCount, err := h.epicService.ListEpics(filters)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to list epics: %v", err))
	}

	limitValue := 50
	if filters.Limit > 0 {
		limitValue = filters.Limit
	}

	epicList := make([]map[string]interface{}, 0, len(epics))
	for _, epic := range epics {
		epicData := map[string]interface{}{
			"reference_id":      epic.ReferenceID,
			"title":             epic.Title,
			"status":            epic.Status,
			"priority":          epic.Priority,
			"creator_username":  epic.Creator.Username,
			"assignee_username": epic.Assignee.Username,
		}
		epicList = append(epicList, epicData)
	}

	responseData := map[string]interface{}{
		"epics":       epicList,
		"total_count": totalCount,
		"limit":       limitValue,
		"offset":      filters.Offset,
	}

	message := fmt.Sprintf("Found %d epics (total: %d)", len(epicList), totalCount)
	return types.CreateDataResponse(message, responseData), nil
}

func parseIncludeList(include string) []string {
	parts := strings.Split(include, ",")
	includes := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			includes = append(includes, trimmed)
		}
	}
	return includes
}
