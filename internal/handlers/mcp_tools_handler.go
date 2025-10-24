package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"product-requirements-management/internal/auth"
	"product-requirements-management/internal/jsonrpc"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// ToolsHandler handles MCP tools/call requests for CRUD operations
type ToolsHandler struct {
	epicService             service.EpicService
	userStoryService        service.UserStoryService
	requirementService      service.RequirementService
	searchService           service.SearchServiceInterface
	steeringDocumentService service.SteeringDocumentService
}

// NewToolsHandler creates a new tools handler instance
func NewToolsHandler(
	epicService service.EpicService,
	userStoryService service.UserStoryService,
	requirementService service.RequirementService,
	searchService service.SearchServiceInterface,
	steeringDocumentService service.SteeringDocumentService,
) *ToolsHandler {
	return &ToolsHandler{
		epicService:             epicService,
		userStoryService:        userStoryService,
		requirementService:      requirementService,
		searchService:           searchService,
		steeringDocumentService: steeringDocumentService,
	}
}

// ToolCallRequest represents a tools/call request
type ToolCallRequest struct {
	Name      string                 `json:"name" validate:"required"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResponse represents the response from a tool call
type ToolResponse struct {
	Content []ContentItem `json:"content"`
}

// ContentItem represents a single content item in a tool response
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// getUserFromContext extracts user information from the context
func getUserFromContext(ctx context.Context) (*models.User, error) {
	// Extract Gin context from the context
	ginCtx, ok := ctx.Value("gin_context").(*gin.Context)
	if !ok {
		return nil, fmt.Errorf("gin context not found")
	}

	// Get user from Gin context using auth helper
	user, ok := auth.GetUserFromContext(ginCtx)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}

	return user, nil
}

// HandleToolsCall processes tools/call requests
func (h *ToolsHandler) HandleToolsCall(ctx context.Context, params interface{}) (interface{}, error) {
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

	// Route to appropriate tool handler
	switch toolName {
	case "create_epic":
		return h.handleCreateEpic(ctx, arguments)
	case "update_epic":
		return h.handleUpdateEpic(ctx, arguments)
	case "create_user_story":
		return h.handleCreateUserStory(ctx, arguments)
	case "update_user_story":
		return h.handleUpdateUserStory(ctx, arguments)
	case "create_requirement":
		return h.handleCreateRequirement(ctx, arguments)
	case "update_requirement":
		return h.handleUpdateRequirement(ctx, arguments)
	case "create_relationship":
		return h.handleCreateRelationship(ctx, arguments)
	case "search_global":
		return h.handleSearchGlobal(ctx, arguments)
	case "search_requirements":
		return h.handleSearchRequirements(ctx, arguments)
	case "list_steering_documents":
		return h.handleListSteeringDocuments(ctx, arguments)
	case "create_steering_document":
		return h.handleCreateSteeringDocument(ctx, arguments)
	case "get_steering_document":
		return h.handleGetSteeringDocument(ctx, arguments)
	case "update_steering_document":
		return h.handleUpdateSteeringDocument(ctx, arguments)
	case "link_steering_to_epic":
		return h.handleLinkSteeringToEpic(ctx, arguments)
	case "unlink_steering_from_epic":
		return h.handleUnlinkSteeringFromEpic(ctx, arguments)
	case "get_epic_steering_documents":
		return h.handleGetEpicSteeringDocuments(ctx, arguments)
	default:
		return nil, jsonrpc.NewMethodNotFoundError(fmt.Sprintf("Unknown tool: %s", toolName))
	}
}

// handleCreateEpic handles the create_epic tool
func (h *ToolsHandler) handleCreateEpic(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	title, ok := args["title"].(string)
	if !ok || title == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'title' argument")
	}

	priorityFloat, ok := args["priority"].(float64)
	if !ok {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'priority' argument")
	}
	priority := models.Priority(int(priorityFloat))

	// Optional arguments
	description, _ := args["description"].(string)
	var assigneeID *uuid.UUID
	if assigneeIDStr, ok := args["assignee_id"].(string); ok && assigneeIDStr != "" {
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
		Priority:    priority,
		Title:       title,
		Description: &description,
	}

	epic, err := h.epicService.CreateEpic(req)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to create epic: %v", err))
	}

	return &ToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully created epic %s: %s", epic.ReferenceID, epic.Title),
			},
			{
				Type: "text",
				Text: fmt.Sprintf("Epic data: %+v", epic),
			},
		},
	}, nil
}

// handleUpdateEpic handles the update_epic tool
func (h *ToolsHandler) handleUpdateEpic(_ context.Context, args map[string]interface{}) (interface{}, error) {
	// Validate required arguments
	epicIDStr, ok := args["epic_id"].(string)
	if !ok || epicIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'epic_id' argument")
	}

	var epicID uuid.UUID
	var err error

	// Try to parse as UUID first, then as reference ID
	if epicID, err = uuid.Parse(epicIDStr); err != nil {
		// Try to get by reference ID
		epic, err := h.epicService.GetEpicByReferenceID(epicIDStr)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'epic_id': not a valid UUID or reference ID")
		}
		epicID = epic.ID
	}

	// Build update request from optional arguments
	req := service.UpdateEpicRequest{}

	if title, ok := args["title"].(string); ok && title != "" {
		req.Title = &title
	}

	if description, ok := args["description"].(string); ok {
		req.Description = &description
	}

	if priorityFloat, ok := args["priority"].(float64); ok {
		priority := models.Priority(int(priorityFloat))
		req.Priority = &priority
	}

	if assigneeIDStr, ok := args["assignee_id"].(string); ok {
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

	if status, ok := args["status"].(string); ok && status != "" {
		epicStatus := models.EpicStatus(status)
		req.Status = &epicStatus
	}

	// Update the epic
	epic, err := h.epicService.UpdateEpic(epicID, req)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to update epic: %v", err))
	}

	return &ToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully updated epic %s: %s", epic.ReferenceID, epic.Title),
			},
			{
				Type: "text",
				Text: fmt.Sprintf("Epic data: %+v", epic),
			},
		},
	}, nil
}

// handleCreateUserStory handles the create_user_story tool
func (h *ToolsHandler) handleCreateUserStory(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	title, ok := args["title"].(string)
	if !ok || title == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'title' argument")
	}

	epicIDStr, ok := args["epic_id"].(string)
	if !ok || epicIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'epic_id' argument")
	}

	var epicID uuid.UUID

	// Try to parse as UUID first, then as reference ID
	if epicID, err = uuid.Parse(epicIDStr); err != nil {
		// Try to get by reference ID
		epic, err := h.epicService.GetEpicByReferenceID(epicIDStr)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'epic_id': not a valid UUID or reference ID")
		}
		epicID = epic.ID
	}

	priorityFloat, ok := args["priority"].(float64)
	if !ok {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'priority' argument")
	}
	priority := models.Priority(int(priorityFloat))

	// Optional arguments
	description, _ := args["description"].(string)
	var assigneeID *uuid.UUID
	if assigneeIDStr, ok := args["assignee_id"].(string); ok && assigneeIDStr != "" {
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

	return &ToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully created user story %s: %s", userStory.ReferenceID, userStory.Title),
			},
			{
				Type: "text",
				Text: fmt.Sprintf("User story data: %+v", userStory),
			},
		},
	}, nil
}

// handleUpdateUserStory handles the update_user_story tool
func (h *ToolsHandler) handleUpdateUserStory(_ context.Context, args map[string]interface{}) (interface{}, error) {
	// Validate required arguments
	userStoryIDStr, ok := args["user_story_id"].(string)
	if !ok || userStoryIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'user_story_id' argument")
	}

	var userStoryID uuid.UUID
	var err error

	// Try to parse as UUID first, then as reference ID
	if userStoryID, err = uuid.Parse(userStoryIDStr); err != nil {
		// Try to get by reference ID
		userStory, err := h.userStoryService.GetUserStoryByReferenceID(userStoryIDStr)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'user_story_id': not a valid UUID or reference ID")
		}
		userStoryID = userStory.ID
	}

	// Build update request from optional arguments
	req := service.UpdateUserStoryRequest{}

	if title, ok := args["title"].(string); ok && title != "" {
		req.Title = &title
	}

	if description, ok := args["description"].(string); ok {
		req.Description = &description
	}

	if priorityFloat, ok := args["priority"].(float64); ok {
		priority := models.Priority(int(priorityFloat))
		req.Priority = &priority
	}

	if assigneeIDStr, ok := args["assignee_id"].(string); ok {
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

	return &ToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully updated user story %s: %s", userStory.ReferenceID, userStory.Title),
			},
			{
				Type: "text",
				Text: fmt.Sprintf("User story data: %+v", userStory),
			},
		},
	}, nil
}

// handleCreateRequirement handles the create_requirement tool
func (h *ToolsHandler) handleCreateRequirement(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	title, ok := args["title"].(string)
	if !ok || title == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'title' argument")
	}

	userStoryIDStr, ok := args["user_story_id"].(string)
	if !ok || userStoryIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'user_story_id' argument")
	}

	var userStoryID uuid.UUID

	// Try to parse as UUID first, then as reference ID
	if userStoryID, err = uuid.Parse(userStoryIDStr); err != nil {
		// Try to get by reference ID
		userStory, err := h.userStoryService.GetUserStoryByReferenceID(userStoryIDStr)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'user_story_id': not a valid UUID or reference ID")
		}
		userStoryID = userStory.ID
	}

	typeIDStr, ok := args["type_id"].(string)
	if !ok || typeIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'type_id' argument")
	}

	typeID, err := uuid.Parse(typeIDStr)
	if err != nil {
		return nil, jsonrpc.NewInvalidParamsError("Invalid 'type_id' format")
	}

	priorityFloat, ok := args["priority"].(float64)
	if !ok {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'priority' argument")
	}
	priority := models.Priority(int(priorityFloat))

	// Optional arguments
	description, _ := args["description"].(string)
	var assigneeID *uuid.UUID
	if assigneeIDStr, ok := args["assignee_id"].(string); ok && assigneeIDStr != "" {
		parsed, err := uuid.Parse(assigneeIDStr)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'assignee_id' format")
		}
		assigneeID = &parsed
	}

	var acceptanceCriteriaID *uuid.UUID
	if acIDStr, ok := args["acceptance_criteria_id"].(string); ok && acIDStr != "" {
		parsed, err := uuid.Parse(acIDStr)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'acceptance_criteria_id' format")
		}
		acceptanceCriteriaID = &parsed
	}

	// Create the requirement
	req := service.CreateRequirementRequest{
		UserStoryID:          userStoryID,
		AcceptanceCriteriaID: acceptanceCriteriaID,
		CreatorID:            user.ID,
		AssigneeID:           assigneeID,
		Priority:             priority,
		TypeID:               typeID,
		Title:                title,
		Description:          &description,
	}

	requirement, err := h.requirementService.CreateRequirement(req)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to create requirement: %v", err))
	}

	return &ToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully created requirement %s: %s", requirement.ReferenceID, requirement.Title),
			},
			{
				Type: "text",
				Text: fmt.Sprintf("Requirement data: %+v", requirement),
			},
		},
	}, nil
}

// handleUpdateRequirement handles the update_requirement tool
func (h *ToolsHandler) handleUpdateRequirement(_ context.Context, args map[string]interface{}) (interface{}, error) {
	// Validate required arguments
	requirementIDStr, ok := args["requirement_id"].(string)
	if !ok || requirementIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'requirement_id' argument")
	}

	var requirementID uuid.UUID
	var err error

	// Try to parse as UUID first, then as reference ID
	if requirementID, err = uuid.Parse(requirementIDStr); err != nil {
		// Try to get by reference ID
		requirement, err := h.requirementService.GetRequirementByReferenceID(requirementIDStr)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'requirement_id': not a valid UUID or reference ID")
		}
		requirementID = requirement.ID
	}

	// Build update request from optional arguments
	req := service.UpdateRequirementRequest{}

	if title, ok := args["title"].(string); ok && title != "" {
		req.Title = &title
	}

	if description, ok := args["description"].(string); ok {
		req.Description = &description
	}

	if priorityFloat, ok := args["priority"].(float64); ok {
		priority := models.Priority(int(priorityFloat))
		req.Priority = &priority
	}

	if assigneeIDStr, ok := args["assignee_id"].(string); ok {
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

	return &ToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully updated requirement %s: %s", requirement.ReferenceID, requirement.Title),
			},
			{
				Type: "text",
				Text: fmt.Sprintf("Requirement data: %+v", requirement),
			},
		},
	}, nil
}

// handleCreateRelationship handles the create_relationship tool
func (h *ToolsHandler) handleCreateRelationship(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	sourceIDStr, ok := args["source_requirement_id"].(string)
	if !ok || sourceIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'source_requirement_id' argument")
	}

	targetIDStr, ok := args["target_requirement_id"].(string)
	if !ok || targetIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'target_requirement_id' argument")
	}

	relationshipTypeIDStr, ok := args["relationship_type_id"].(string)
	if !ok || relationshipTypeIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'relationship_type_id' argument")
	}

	// Parse source requirement ID
	var sourceID uuid.UUID
	if parsedID, parseErr := uuid.Parse(sourceIDStr); parseErr != nil {
		// Try to get by reference ID
		requirement, reqErr := h.requirementService.GetRequirementByReferenceID(sourceIDStr)
		if reqErr != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'source_requirement_id': not a valid UUID or reference ID")
		}
		sourceID = requirement.ID
	} else {
		sourceID = parsedID
	}

	// Parse target requirement ID
	var targetID uuid.UUID
	if parsedID, parseErr := uuid.Parse(targetIDStr); parseErr != nil {
		// Try to get by reference ID
		requirement, reqErr := h.requirementService.GetRequirementByReferenceID(targetIDStr)
		if reqErr != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'target_requirement_id': not a valid UUID or reference ID")
		}
		targetID = requirement.ID
	} else {
		targetID = parsedID
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

	return &ToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully created relationship between requirements %s and %s", sourceIDStr, targetIDStr),
			},
			{
				Type: "text",
				Text: fmt.Sprintf("Relationship data: %+v", relationship),
			},
		},
	}, nil
}

// handleSearchGlobal handles the search_global tool
func (h *ToolsHandler) handleSearchGlobal(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Validate required arguments
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'query' argument")
	}

	// Optional arguments
	var entityTypes []string
	if entityTypesInterface, ok := args["entity_types"].([]interface{}); ok {
		for _, et := range entityTypesInterface {
			if etStr, ok := et.(string); ok {
				entityTypes = append(entityTypes, etStr)
			}
		}
	}

	limit := 50 // Default limit
	if limitFloat, ok := args["limit"].(float64); ok {
		limit = int(limitFloat)
	}

	offset := 0 // Default offset
	if offsetFloat, ok := args["offset"].(float64); ok {
		offset = int(offsetFloat)
	}

	// Create search options with entity types
	searchOptions := service.SearchOptions{
		Query:       query,
		EntityTypes: entityTypes,
		Limit:       limit,
		Offset:      offset,
	}

	// Perform the search using the search service
	response, err := h.searchService.Search(ctx, searchOptions)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Search failed: %v", err))
	}

	// Convert search results to JSON string for MCP compatibility
	searchData := map[string]interface{}{
		"results":      response.Results,
		"total_count":  response.Total,
		"query":        query,
		"entity_types": entityTypes,
		"limit":        limit,
		"offset":       offset,
	}

	jsonData, err := json.MarshalIndent(searchData, "", "  ")
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to marshal search results: %v", err))
	}

	return &ToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Found %d results for query '%s'", response.Total, query),
			},
			{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}, nil
}

// handleSearchRequirements handles the search_requirements tool
func (h *ToolsHandler) handleSearchRequirements(_ context.Context, args map[string]interface{}) (interface{}, error) {
	// Validate required arguments
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'query' argument")
	}

	// Perform the search using the requirement service
	requirements, err := h.requirementService.SearchRequirements(query)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Requirements search failed: %v", err))
	}

	// Convert requirements search results to JSON string for MCP compatibility
	searchData := map[string]interface{}{
		"requirements": requirements,
		"query":        query,
		"count":        len(requirements),
	}

	jsonData, err := json.MarshalIndent(searchData, "", "  ")
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to marshal requirements search results: %v", err))
	}

	return &ToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Found %d requirements matching query '%s'", len(requirements), query),
			},
			{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}, nil
}

// handleListSteeringDocuments handles the list_steering_documents tool
func (h *ToolsHandler) handleListSteeringDocuments(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Build filters from optional arguments
	filters := service.SteeringDocumentFilters{}

	if creatorIDStr, ok := args["creator_id"].(string); ok && creatorIDStr != "" {
		creatorID, err := uuid.Parse(creatorIDStr)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'creator_id' format")
		}
		filters.CreatorID = &creatorID
	}

	if search, ok := args["search"].(string); ok {
		filters.Search = search
	}

	if orderBy, ok := args["order_by"].(string); ok {
		filters.OrderBy = orderBy
	}

	if limitFloat, ok := args["limit"].(float64); ok {
		filters.Limit = int(limitFloat)
	}

	if offsetFloat, ok := args["offset"].(float64); ok {
		filters.Offset = int(offsetFloat)
	}

	// List steering documents
	docs, total, err := h.steeringDocumentService.ListSteeringDocuments(filters, user)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to list steering documents: %v", err))
	}

	// Convert results to JSON string for MCP compatibility
	listData := map[string]interface{}{
		"steering_documents": docs,
		"total_count":        total,
		"limit":              filters.Limit,
		"offset":             filters.Offset,
	}

	jsonData, err := json.MarshalIndent(listData, "", "  ")
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to marshal steering documents list: %v", err))
	}

	return &ToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Found %d steering documents (total: %d)", len(docs), total),
			},
			{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}, nil
}

// handleCreateSteeringDocument handles the create_steering_document tool
func (h *ToolsHandler) handleCreateSteeringDocument(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	title, ok := args["title"].(string)
	if !ok || title == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'title' argument")
	}

	// Optional arguments
	var description *string
	if desc, ok := args["description"].(string); ok && desc != "" {
		description = &desc
	}

	var epicID *string
	if epicIDStr, ok := args["epic_id"].(string); ok && epicIDStr != "" {
		epicID = &epicIDStr
	}

	// Create the steering document
	req := service.CreateSteeringDocumentRequest{
		Title:       title,
		Description: description,
		EpicID:      epicID,
	}

	doc, err := h.steeringDocumentService.CreateSteeringDocument(req, user)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to create steering document: %v", err))
	}

	// Prepare success message
	successMsg := fmt.Sprintf("Successfully created steering document %s: %s", doc.ReferenceID, doc.Title)
	if epicID != nil {
		successMsg += " and linked to epic"
	}

	return &ToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: successMsg,
			},
			{
				Type: "text",
				Text: fmt.Sprintf("Steering document data: %+v", doc),
			},
		},
	}, nil
}

// handleGetSteeringDocument handles the get_steering_document tool
func (h *ToolsHandler) handleGetSteeringDocument(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	steeringDocIDStr, ok := args["steering_document_id"].(string)
	if !ok || steeringDocIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'steering_document_id' argument")
	}

	var doc *models.SteeringDocument

	// Try to parse as UUID first, then as reference ID
	if steeringDocID, err := uuid.Parse(steeringDocIDStr); err == nil {
		doc, err = h.steeringDocumentService.GetSteeringDocumentByID(steeringDocID, user)
		if err != nil {
			return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get steering document by ID: %v", err))
		}
	} else {
		// Try to get by reference ID
		doc, err = h.steeringDocumentService.GetSteeringDocumentByReferenceID(steeringDocIDStr, user)
		if err != nil {
			return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get steering document by reference ID: %v", err))
		}
	}

	// Convert document to JSON string for MCP compatibility
	jsonData, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to marshal steering document: %v", err))
	}

	return &ToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Retrieved steering document %s: %s", doc.ReferenceID, doc.Title),
			},
			{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}, nil
}

// handleUpdateSteeringDocument handles the update_steering_document tool
func (h *ToolsHandler) handleUpdateSteeringDocument(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	steeringDocIDStr, ok := args["steering_document_id"].(string)
	if !ok || steeringDocIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'steering_document_id' argument")
	}

	var steeringDocID uuid.UUID

	// Try to parse as UUID first, then as reference ID
	if parsedID, err := uuid.Parse(steeringDocIDStr); err == nil {
		steeringDocID = parsedID
	} else {
		// Try to get by reference ID
		doc, err := h.steeringDocumentService.GetSteeringDocumentByReferenceID(steeringDocIDStr, user)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'steering_document_id': not a valid UUID or reference ID")
		}
		steeringDocID = doc.ID
	}

	// Build update request from optional arguments
	req := service.UpdateSteeringDocumentRequest{}

	if title, ok := args["title"].(string); ok && title != "" {
		req.Title = &title
	}

	if description, ok := args["description"].(string); ok {
		req.Description = &description
	}

	// Update the steering document
	doc, err := h.steeringDocumentService.UpdateSteeringDocument(steeringDocID, req, user)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to update steering document: %v", err))
	}

	return &ToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully updated steering document %s: %s", doc.ReferenceID, doc.Title),
			},
			{
				Type: "text",
				Text: fmt.Sprintf("Steering document data: %+v", doc),
			},
		},
	}, nil
}

// handleLinkSteeringToEpic handles the link_steering_to_epic tool
func (h *ToolsHandler) handleLinkSteeringToEpic(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	steeringDocIDStr, ok := args["steering_document_id"].(string)
	if !ok || steeringDocIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'steering_document_id' argument")
	}

	epicIDStr, ok := args["epic_id"].(string)
	if !ok || epicIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'epic_id' argument")
	}

	// Parse steering document ID
	var steeringDocID uuid.UUID
	if parsedID, err := uuid.Parse(steeringDocIDStr); err == nil {
		steeringDocID = parsedID
	} else {
		// Try to get by reference ID
		doc, err := h.steeringDocumentService.GetSteeringDocumentByReferenceID(steeringDocIDStr, user)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'steering_document_id': not a valid UUID or reference ID")
		}
		steeringDocID = doc.ID
	}

	// Parse epic ID
	var epicID uuid.UUID
	if parsedID, err := uuid.Parse(epicIDStr); err == nil {
		epicID = parsedID
	} else {
		// Try to get by reference ID
		epic, err := h.epicService.GetEpicByReferenceID(epicIDStr)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'epic_id': not a valid UUID or reference ID")
		}
		epicID = epic.ID
	}

	// Create the link
	err = h.steeringDocumentService.LinkSteeringDocumentToEpic(steeringDocID, epicID, user)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to link steering document to epic: %v", err))
	}

	return &ToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully linked steering document %s to epic %s", steeringDocIDStr, epicIDStr),
			},
		},
	}, nil
}

// handleUnlinkSteeringFromEpic handles the unlink_steering_from_epic tool
func (h *ToolsHandler) handleUnlinkSteeringFromEpic(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	steeringDocIDStr, ok := args["steering_document_id"].(string)
	if !ok || steeringDocIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'steering_document_id' argument")
	}

	epicIDStr, ok := args["epic_id"].(string)
	if !ok || epicIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'epic_id' argument")
	}

	// Parse steering document ID
	var steeringDocID uuid.UUID
	if parsedID, err := uuid.Parse(steeringDocIDStr); err == nil {
		steeringDocID = parsedID
	} else {
		// Try to get by reference ID
		doc, err := h.steeringDocumentService.GetSteeringDocumentByReferenceID(steeringDocIDStr, user)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'steering_document_id': not a valid UUID or reference ID")
		}
		steeringDocID = doc.ID
	}

	// Parse epic ID
	var epicID uuid.UUID
	if parsedID, err := uuid.Parse(epicIDStr); err == nil {
		epicID = parsedID
	} else {
		// Try to get by reference ID
		epic, err := h.epicService.GetEpicByReferenceID(epicIDStr)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'epic_id': not a valid UUID or reference ID")
		}
		epicID = epic.ID
	}

	// Remove the link
	err = h.steeringDocumentService.UnlinkSteeringDocumentFromEpic(steeringDocID, epicID, user)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to unlink steering document from epic: %v", err))
	}

	return &ToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully unlinked steering document %s from epic %s", steeringDocIDStr, epicIDStr),
			},
		},
	}, nil
}

// handleGetEpicSteeringDocuments handles the get_epic_steering_documents tool
func (h *ToolsHandler) handleGetEpicSteeringDocuments(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	epicIDStr, ok := args["epic_id"].(string)
	if !ok || epicIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'epic_id' argument")
	}

	var epicID uuid.UUID

	// Try to parse as UUID first, then as reference ID
	if parsedID, err := uuid.Parse(epicIDStr); err == nil {
		epicID = parsedID
	} else {
		// Try to get by reference ID
		epic, err := h.epicService.GetEpicByReferenceID(epicIDStr)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'epic_id': not a valid UUID or reference ID")
		}
		epicID = epic.ID
	}

	// Get steering documents for the epic
	docs, err := h.steeringDocumentService.GetSteeringDocumentsByEpicID(epicID, user)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get steering documents for epic: %v", err))
	}

	// Convert results to JSON string for MCP compatibility
	epicData := map[string]interface{}{
		"epic_id":            epicIDStr,
		"steering_documents": docs,
		"count":              len(docs),
	}

	jsonData, err := json.MarshalIndent(epicData, "", "  ")
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to marshal epic steering documents: %v", err))
	}

	return &ToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Found %d steering documents linked to epic %s", len(docs), epicIDStr),
			},
			{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}, nil
}
