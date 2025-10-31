package tools

import (
	"context"
	"fmt"

	"product-requirements-management/internal/jsonrpc"
	"product-requirements-management/internal/mcp/types"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"

	"github.com/google/uuid"
)

// RequirementHandler handles MCP tools for Requirement domain operations
type RequirementHandler struct {
	requirementService service.RequirementService
	userStoryService   service.UserStoryService
}

// NewRequirementHandler creates a new Requirement handler instance
func NewRequirementHandler(requirementService service.RequirementService, userStoryService service.UserStoryService) *RequirementHandler {
	return &RequirementHandler{
		requirementService: requirementService,
		userStoryService:   userStoryService,
	}
}

// GetSupportedTools returns the list of tools this handler supports
func (h *RequirementHandler) GetSupportedTools() []string {
	return []string{
		"create_requirement",
		"update_requirement",
		"create_relationship",
	}
}

// HandleTool processes a specific tool call for Requirement domain operations
func (h *RequirementHandler) HandleTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	switch toolName {
	case "create_requirement":
		return h.Create(ctx, args)
	case "update_requirement":
		return h.Update(ctx, args)
	case "create_relationship":
		return h.CreateRelationship(ctx, args)
	default:
		return nil, jsonrpc.NewMethodNotFoundError(fmt.Sprintf("Unknown Requirement tool: %s", toolName))
	}
}

// Create handles the create_requirement tool
func (h *RequirementHandler) Create(ctx context.Context, args map[string]interface{}) (interface{}, error) {
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

	userStoryIDStr, ok := getStringArg(args, "user_story_id")
	if !ok || userStoryIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'user_story_id' argument")
	}

	var userStory *models.UserStory
	userStoryUUID, err := uuid.Parse(userStoryIDStr)
	if err == nil {
		userStory, err = h.userStoryService.GetUserStoryByID(userStoryUUID)
	} else {
		userStory, err = h.userStoryService.GetUserStoryByReferenceID(userStoryIDStr)
	}

	if userStory == nil {
		return nil, jsonrpc.NewInvalidParamsError("Invalid 'user_story_id': not a valid UUID or reference ID")
	}

	typeIDStr, ok := getStringArg(args, "type_id")
	if !ok || typeIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'type_id' argument")
	}

	var typeID uuid.UUID
	if typeID, err = uuid.Parse(typeIDStr); err != nil {
		return nil, jsonrpc.NewInvalidParamsError("Invalid 'type_id' format")
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

	var acceptanceCriteriaID *uuid.UUID
	if acIDStr, ok := getStringArg(args, "acceptance_criteria_id"); ok && acIDStr != "" {
		parsed, err := uuid.Parse(acIDStr)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'acceptance_criteria_id' format")
		}
		acceptanceCriteriaID = &parsed
	}

	// Create the requirement
	req := service.CreateRequirementRequest{
		UserStoryID:          userStory.ID,
		AcceptanceCriteriaID: acceptanceCriteriaID,
		CreatorID:            user.ID,
		AssigneeID:           assigneeID,
		Priority:             models.Priority(priority),
		TypeID:               typeID,
		Title:                title,
		Description:          &description,
	}

	requirement, err := h.requirementService.CreateRequirement(req)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to create requirement: %v", err))
	}

	message := fmt.Sprintf("Successfully created requirement %s: %s", requirement.ReferenceID, requirement.Title)
	return types.CreateDataResponse(message, requirement), nil
}

// Update handles the update_requirement tool
func (h *RequirementHandler) Update(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Validate required arguments
	requirementIDStr, ok := getStringArg(args, "requirement_id")
	if !ok || requirementIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'requirement_id' argument")
	}

	// Parse requirement ID (UUID or reference ID)
	requirementID, err := parseUUIDOrReferenceID(requirementIDStr, func(refID string) (interface{}, error) {
		return h.requirementService.GetRequirementByReferenceID(refID)
	})
	if err != nil {
		return nil, jsonrpc.NewInvalidParamsError("Invalid 'requirement_id': not a valid UUID or reference ID")
	}

	// Build update request from optional arguments
	req := service.UpdateRequirementRequest{}

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

	// Update the requirement
	requirement, err := h.requirementService.UpdateRequirement(requirementID, req)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to update requirement: %v", err))
	}

	message := fmt.Sprintf("Successfully updated requirement %s: %s", requirement.ReferenceID, requirement.Title)
	return types.CreateDataResponse(message, requirement), nil
}

// CreateRelationship handles the create_relationship tool
func (h *RequirementHandler) CreateRelationship(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	sourceIDStr, ok := getStringArg(args, "source_requirement_id")
	if !ok || sourceIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'source_requirement_id' argument")
	}

	targetIDStr, ok := getStringArg(args, "target_requirement_id")
	if !ok || targetIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'target_requirement_id' argument")
	}

	relationshipTypeIDStr, ok := getStringArg(args, "relationship_type_id")
	if !ok || relationshipTypeIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'relationship_type_id' argument")
	}

	// Parse source requirement ID (UUID or reference ID)
	sourceID, err := parseUUIDOrReferenceID(sourceIDStr, func(refID string) (interface{}, error) {
		return h.requirementService.GetRequirementByReferenceID(refID)
	})
	if err != nil {
		return nil, jsonrpc.NewInvalidParamsError("Invalid 'source_requirement_id': not a valid UUID or reference ID")
	}

	// Parse target requirement ID (UUID or reference ID)
	targetID, err := parseUUIDOrReferenceID(targetIDStr, func(refID string) (interface{}, error) {
		return h.requirementService.GetRequirementByReferenceID(refID)
	})
	if err != nil {
		return nil, jsonrpc.NewInvalidParamsError("Invalid 'target_requirement_id': not a valid UUID or reference ID")
	}

	// Parse relationship type ID
	relationshipTypeID, err := uuid.Parse(relationshipTypeIDStr)
	if err != nil {
		return nil, jsonrpc.NewInvalidParamsError("Invalid 'relationship_type_id' format")
	}

	// Create the relationship
	req := service.CreateRelationshipRequest{
		SourceRequirementID: sourceID,
		TargetRequirementID: targetID,
		RelationshipTypeID:  relationshipTypeID,
		CreatedBy:           user.ID,
	}

	relationship, err := h.requirementService.CreateRelationship(req)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to create relationship: %v", err))
	}

	message := fmt.Sprintf("Successfully created relationship between requirements %s and %s", sourceIDStr, targetIDStr)
	return types.CreateDataResponse(message, relationship), nil
}
