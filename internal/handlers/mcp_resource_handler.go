package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"product-requirements-management/internal/jsonrpc"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"

	"github.com/google/uuid"
)

// ResourceHandler handles MCP resources/read requests
type ResourceHandler struct {
	epicService               service.EpicService
	userStoryService          service.UserStoryService
	requirementService        service.RequirementService
	acceptanceCriteriaService service.AcceptanceCriteriaService
	promptService             *service.PromptService
	uriParser                 *URIParser
}

// NewResourceHandler creates a new resource handler instance
func NewResourceHandler(
	epicService service.EpicService,
	userStoryService service.UserStoryService,
	requirementService service.RequirementService,
	acceptanceCriteriaService service.AcceptanceCriteriaService,
	promptService *service.PromptService,
) *ResourceHandler {
	return &ResourceHandler{
		epicService:               epicService,
		userStoryService:          userStoryService,
		requirementService:        requirementService,
		acceptanceCriteriaService: acceptanceCriteriaService,
		promptService:             promptService,
		uriParser:                 NewURIParser(),
	}
}

// ResourceReadRequest represents the parameters for resources/read method
type ResourceReadRequest struct {
	URI string `json:"uri" validate:"required"`
}

// ResourceResponse represents the response for resources/read method
type ResourceResponse struct {
	Contents []ResourceContents `json:"contents"`
}

// ResourceContents represents individual resource content according to MCP spec
type ResourceContents struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
}

// HandleResourcesRead handles the resources/read method
func (rh *ResourceHandler) HandleResourcesRead(ctx context.Context, params interface{}) (interface{}, error) {
	// Parse parameters
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, jsonrpc.NewInvalidParamsError("Invalid parameters format")
	}

	uri, ok := paramsMap["uri"].(string)
	if !ok || uri == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid URI parameter")
	}

	// Check if this is a requirements:// URI (from resources/list) and convert it
	if strings.HasPrefix(uri, "requirements://") {
		// Check if it's a collection resource first
		if rh.isCollectionResource(uri) {
			return rh.handleCollectionResource(ctx, uri)
		}

		convertedURI, err := rh.convertRequirementsURI(ctx, uri)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError(fmt.Sprintf("Failed to convert URI: %v", err))
		}
		uri = convertedURI
	}

	// Parse the URI
	parsedURI, err := rh.uriParser.Parse(uri)
	if err != nil {
		return nil, jsonrpc.NewInvalidParamsError(fmt.Sprintf("Invalid URI: %v", err))
	}

	// Handle the resource based on scheme and sub-path
	return rh.handleResourceByScheme(ctx, parsedURI)
}

// handleResourceByScheme routes the request based on URI scheme
func (rh *ResourceHandler) handleResourceByScheme(ctx context.Context, parsedURI *ParsedURI) (interface{}, error) {
	switch parsedURI.Scheme {
	case EpicURIScheme:
		return rh.handleEpicResource(ctx, parsedURI)
	case UserStoryURIScheme:
		return rh.handleUserStoryResource(ctx, parsedURI)
	case RequirementURIScheme:
		return rh.handleRequirementResource(ctx, parsedURI)
	case AcceptanceCriteriaURIScheme:
		return rh.handleAcceptanceCriteriaResource(ctx, parsedURI)
	case PromptURIScheme:
		return rh.handlePromptResource(ctx, parsedURI)
	default:
		return nil, jsonrpc.NewInvalidParamsError(fmt.Sprintf("Unsupported URI scheme: %s", parsedURI.Scheme))
	}
}

// handleEpicResource handles epic:// URIs
func (rh *ResourceHandler) handleEpicResource(_ context.Context, parsedURI *ParsedURI) (interface{}, error) {
	// Get the epic by reference ID
	epic, err := rh.epicService.GetEpicByReferenceID(parsedURI.ReferenceID)
	if err != nil {
		if err == service.ErrEpicNotFound {
			return nil, jsonrpc.NewJSONRPCError(-32002, "Epic not found", nil)
		}
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get epic: %v", err))
	}

	// Handle sub-paths
	switch parsedURI.SubPath {
	case "":
		// Return the epic itself
		return rh.formatEpicResource(parsedURI, epic), nil
	case "hierarchy":
		// Return epic with full hierarchy
		epicWithUserStories, err := rh.epicService.GetEpicWithUserStories(epic.ID)
		if err != nil {
			return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get epic hierarchy: %v", err))
		}
		return rh.formatEpicHierarchyResource(parsedURI, epicWithUserStories), nil
	case "user-stories":
		// Return just the user stories
		userStories, err := rh.userStoryService.GetUserStoriesByEpic(epic.ID)
		if err != nil {
			return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user stories: %v", err))
		}
		return rh.formatUserStoriesResource(parsedURI, userStories), nil
	default:
		return nil, jsonrpc.NewInvalidParamsError(fmt.Sprintf("Unsupported sub-path for epic: %s", parsedURI.SubPath))
	}
}

// handleUserStoryResource handles user-story:// URIs
func (rh *ResourceHandler) handleUserStoryResource(_ context.Context, parsedURI *ParsedURI) (interface{}, error) {
	// Get the user story by reference ID
	userStory, err := rh.userStoryService.GetUserStoryByReferenceID(parsedURI.ReferenceID)
	if err != nil {
		if err == service.ErrUserStoryNotFound {
			return nil, jsonrpc.NewJSONRPCError(-32002, "User story not found", nil)
		}
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user story: %v", err))
	}

	// Handle sub-paths
	switch parsedURI.SubPath {
	case "":
		// Return the user story itself
		return rh.formatUserStoryResource(parsedURI, userStory), nil
	case "requirements":
		// Return requirements for this user story
		requirements, err := rh.requirementService.GetRequirementsByUserStory(userStory.ID)
		if err != nil {
			return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get requirements: %v", err))
		}
		return rh.formatRequirementsResource(parsedURI, requirements), nil
	case "acceptance-criteria":
		// Return acceptance criteria for this user story
		acceptanceCriteria, _, err := rh.acceptanceCriteriaService.GetAcceptanceCriteriaByUserStory(userStory.ID, 100, 0)
		if err != nil {
			return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get acceptance criteria: %v", err))
		}
		return rh.formatAcceptanceCriteriaListResource(parsedURI, acceptanceCriteria), nil
	default:
		return nil, jsonrpc.NewInvalidParamsError(fmt.Sprintf("Unsupported sub-path for user story: %s", parsedURI.SubPath))
	}
}

// handleRequirementResource handles requirement:// URIs
func (rh *ResourceHandler) handleRequirementResource(_ context.Context, parsedURI *ParsedURI) (interface{}, error) {
	// Get the requirement by reference ID
	requirement, err := rh.requirementService.GetRequirementByReferenceID(parsedURI.ReferenceID)
	if err != nil {
		if err == service.ErrRequirementNotFound {
			return nil, jsonrpc.NewJSONRPCError(-32002, "Requirement not found", nil)
		}
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get requirement: %v", err))
	}

	// Handle sub-paths
	switch parsedURI.SubPath {
	case "":
		// Return the requirement itself
		return rh.formatRequirementResource(parsedURI, requirement), nil
	case "relationships":
		// Return requirement with relationships
		requirementWithRelationships, err := rh.requirementService.GetRequirementWithRelationships(requirement.ID)
		if err != nil {
			return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get requirement relationships: %v", err))
		}
		return rh.formatRequirementRelationshipsResource(parsedURI, requirementWithRelationships), nil
	default:
		return nil, jsonrpc.NewInvalidParamsError(fmt.Sprintf("Unsupported sub-path for requirement: %s", parsedURI.SubPath))
	}
}

// handleAcceptanceCriteriaResource handles acceptance-criteria:// URIs
func (rh *ResourceHandler) handleAcceptanceCriteriaResource(_ context.Context, parsedURI *ParsedURI) (interface{}, error) {
	// Get the acceptance criteria by reference ID
	acceptanceCriteria, err := rh.acceptanceCriteriaService.GetAcceptanceCriteriaByReferenceID(parsedURI.ReferenceID)
	if err != nil {
		if err == service.ErrAcceptanceCriteriaNotFound {
			return nil, jsonrpc.NewJSONRPCError(-32002, "Acceptance criteria not found", nil)
		}
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get acceptance criteria: %v", err))
	}

	// Handle sub-paths (currently no sub-paths supported for acceptance criteria)
	switch parsedURI.SubPath {
	case "":
		// Return the acceptance criteria itself
		return rh.formatAcceptanceCriteriaResource(parsedURI, acceptanceCriteria), nil
	default:
		return nil, jsonrpc.NewInvalidParamsError(fmt.Sprintf("Unsupported sub-path for acceptance criteria: %s", parsedURI.SubPath))
	}
}

// Resource formatting methods

// formatEpicResource formats an epic as a resource response
func (rh *ResourceHandler) formatEpicResource(parsedURI *ParsedURI, epic *models.Epic) *ResourceResponse {
	contents := map[string]interface{}{
		"id":           epic.ID,
		"reference_id": epic.ReferenceID,
		"title":        epic.Title,
		"description":  epic.Description,
		"status":       epic.Status,
		"priority":     epic.Priority,
		"creator_id":   epic.CreatorID,
		"assignee_id":  epic.AssigneeID,
		"created_at":   epic.CreatedAt,
		"updated_at":   epic.UpdatedAt,
	}

	contentsJSON, _ := json.Marshal(contents)

	return &ResourceResponse{
		Contents: []ResourceContents{
			{
				URI:      rh.buildURIFromParsed(parsedURI),
				MimeType: "application/json",
				Text:     string(contentsJSON),
			},
		},
	}
}

// formatEpicHierarchyResource formats an epic with its hierarchy as a resource response
func (rh *ResourceHandler) formatEpicHierarchyResource(parsedURI *ParsedURI, epic *models.Epic) *ResourceResponse {
	contents := map[string]interface{}{
		"id":           epic.ID,
		"reference_id": epic.ReferenceID,
		"title":        epic.Title,
		"description":  epic.Description,
		"status":       epic.Status,
		"priority":     epic.Priority,
		"creator_id":   epic.CreatorID,
		"assignee_id":  epic.AssigneeID,
		"created_at":   epic.CreatedAt,
		"updated_at":   epic.UpdatedAt,
		"user_stories": []interface{}{},
	}

	// Add user stories if they exist
	if epic.UserStories != nil {
		userStories := make([]interface{}, len(epic.UserStories))
		for i, us := range epic.UserStories {
			userStories[i] = map[string]interface{}{
				"id":           us.ID,
				"reference_id": us.ReferenceID,
				"title":        us.Title,
				"description":  us.Description,
				"status":       us.Status,
				"priority":     us.Priority,
				"creator_id":   us.CreatorID,
				"assignee_id":  us.AssigneeID,
				"created_at":   us.CreatedAt,
				"updated_at":   us.UpdatedAt,
			}
		}
		contents["user_stories"] = userStories
	}

	contentsJSON, _ := json.Marshal(contents)

	return &ResourceResponse{
		Contents: []ResourceContents{
			{
				URI:      rh.buildURIFromParsed(parsedURI),
				MimeType: "application/json",
				Text:     string(contentsJSON),
			},
		},
	}
}

// formatUserStoriesResource formats a list of user stories as a resource response
func (rh *ResourceHandler) formatUserStoriesResource(parsedURI *ParsedURI, userStories []models.UserStory) *ResourceResponse {
	userStoriesData := make([]interface{}, len(userStories))
	for i, us := range userStories {
		userStoriesData[i] = map[string]interface{}{
			"id":           us.ID,
			"reference_id": us.ReferenceID,
			"title":        us.Title,
			"description":  us.Description,
			"status":       us.Status,
			"priority":     us.Priority,
			"creator_id":   us.CreatorID,
			"assignee_id":  us.AssigneeID,
			"created_at":   us.CreatedAt,
			"updated_at":   us.UpdatedAt,
		}
	}

	contents := map[string]interface{}{
		"user_stories": userStoriesData,
		"count":        len(userStories),
	}

	contentsJSON, _ := json.Marshal(contents)

	return &ResourceResponse{
		Contents: []ResourceContents{
			{
				URI:      rh.buildURIFromParsed(parsedURI),
				MimeType: "application/json",
				Text:     string(contentsJSON),
			},
		},
	}
}

// formatUserStoryResource formats a user story as a resource response
func (rh *ResourceHandler) formatUserStoryResource(parsedURI *ParsedURI, userStory *models.UserStory) *ResourceResponse {
	contents := map[string]interface{}{
		"id":           userStory.ID,
		"reference_id": userStory.ReferenceID,
		"title":        userStory.Title,
		"description":  userStory.Description,
		"status":       userStory.Status,
		"priority":     userStory.Priority,
		"epic_id":      userStory.EpicID,
		"creator_id":   userStory.CreatorID,
		"assignee_id":  userStory.AssigneeID,
		"created_at":   userStory.CreatedAt,
		"updated_at":   userStory.UpdatedAt,
	}

	contentsJSON, _ := json.Marshal(contents)

	return &ResourceResponse{
		Contents: []ResourceContents{
			{
				URI:      rh.buildURIFromParsed(parsedURI),
				MimeType: "application/json",
				Text:     string(contentsJSON),
			},
		},
	}
}

// formatRequirementsResource formats a list of requirements as a resource response
func (rh *ResourceHandler) formatRequirementsResource(parsedURI *ParsedURI, requirements []models.Requirement) *ResourceResponse {
	requirementsData := make([]interface{}, len(requirements))
	for i, req := range requirements {
		requirementsData[i] = map[string]interface{}{
			"id":                     req.ID,
			"reference_id":           req.ReferenceID,
			"title":                  req.Title,
			"description":            req.Description,
			"status":                 req.Status,
			"priority":               req.Priority,
			"user_story_id":          req.UserStoryID,
			"acceptance_criteria_id": req.AcceptanceCriteriaID,
			"type_id":                req.TypeID,
			"creator_id":             req.CreatorID,
			"assignee_id":            req.AssigneeID,
			"created_at":             req.CreatedAt,
			"updated_at":             req.UpdatedAt,
		}
	}

	contents := map[string]interface{}{
		"requirements": requirementsData,
		"count":        len(requirements),
	}

	contentsJSON, _ := json.Marshal(contents)

	return &ResourceResponse{
		Contents: []ResourceContents{
			{
				URI:      rh.buildURIFromParsed(parsedURI),
				MimeType: "application/json",
				Text:     string(contentsJSON),
			},
		},
	}
}

// formatRequirementResource formats a requirement as a resource response
func (rh *ResourceHandler) formatRequirementResource(parsedURI *ParsedURI, requirement *models.Requirement) *ResourceResponse {
	contents := map[string]interface{}{
		"id":                     requirement.ID,
		"reference_id":           requirement.ReferenceID,
		"title":                  requirement.Title,
		"description":            requirement.Description,
		"status":                 requirement.Status,
		"priority":               requirement.Priority,
		"user_story_id":          requirement.UserStoryID,
		"acceptance_criteria_id": requirement.AcceptanceCriteriaID,
		"type_id":                requirement.TypeID,
		"creator_id":             requirement.CreatorID,
		"assignee_id":            requirement.AssigneeID,
		"created_at":             requirement.CreatedAt,
		"updated_at":             requirement.UpdatedAt,
	}

	contentsJSON, _ := json.Marshal(contents)

	return &ResourceResponse{
		Contents: []ResourceContents{
			{
				URI:      rh.buildURIFromParsed(parsedURI),
				MimeType: "application/json",
				Text:     string(contentsJSON),
			},
		},
	}
}

// formatRequirementRelationshipsResource formats a requirement with relationships as a resource response
func (rh *ResourceHandler) formatRequirementRelationshipsResource(parsedURI *ParsedURI, requirement *models.Requirement) *ResourceResponse {
	contents := map[string]interface{}{
		"id":                     requirement.ID,
		"reference_id":           requirement.ReferenceID,
		"title":                  requirement.Title,
		"description":            requirement.Description,
		"status":                 requirement.Status,
		"priority":               requirement.Priority,
		"user_story_id":          requirement.UserStoryID,
		"acceptance_criteria_id": requirement.AcceptanceCriteriaID,
		"type_id":                requirement.TypeID,
		"creator_id":             requirement.CreatorID,
		"assignee_id":            requirement.AssigneeID,
		"created_at":             requirement.CreatedAt,
		"updated_at":             requirement.UpdatedAt,
		"source_relationships":   []interface{}{},
		"target_relationships":   []interface{}{},
	}

	// Add source relationships if they exist
	if requirement.SourceRelationships != nil {
		sourceRels := make([]interface{}, len(requirement.SourceRelationships))
		for i, rel := range requirement.SourceRelationships {
			sourceRels[i] = map[string]interface{}{
				"id":                    rel.ID,
				"source_requirement_id": rel.SourceRequirementID,
				"target_requirement_id": rel.TargetRequirementID,
				"relationship_type_id":  rel.RelationshipTypeID,
				"created_by":            rel.CreatedBy,
				"created_at":            rel.CreatedAt,
			}
		}
		contents["source_relationships"] = sourceRels
	}

	// Add target relationships if they exist
	if requirement.TargetRelationships != nil {
		targetRels := make([]interface{}, len(requirement.TargetRelationships))
		for i, rel := range requirement.TargetRelationships {
			targetRels[i] = map[string]interface{}{
				"id":                    rel.ID,
				"source_requirement_id": rel.SourceRequirementID,
				"target_requirement_id": rel.TargetRequirementID,
				"relationship_type_id":  rel.RelationshipTypeID,
				"created_by":            rel.CreatedBy,
				"created_at":            rel.CreatedAt,
			}
		}
		contents["target_relationships"] = targetRels
	}

	contentsJSON, _ := json.Marshal(contents)

	return &ResourceResponse{
		Contents: []ResourceContents{
			{
				URI:      rh.buildURIFromParsed(parsedURI),
				MimeType: "application/json",
				Text:     string(contentsJSON),
			},
		},
	}
}

// formatAcceptanceCriteriaResource formats acceptance criteria as a resource response
func (rh *ResourceHandler) formatAcceptanceCriteriaResource(parsedURI *ParsedURI, acceptanceCriteria *models.AcceptanceCriteria) *ResourceResponse {
	contents := map[string]interface{}{
		"id":            acceptanceCriteria.ID,
		"reference_id":  acceptanceCriteria.ReferenceID,
		"description":   acceptanceCriteria.Description,
		"user_story_id": acceptanceCriteria.UserStoryID,
		"author_id":     acceptanceCriteria.AuthorID,
		"created_at":    acceptanceCriteria.CreatedAt,
		"updated_at":    acceptanceCriteria.UpdatedAt,
	}

	contentsJSON, _ := json.Marshal(contents)

	return &ResourceResponse{
		Contents: []ResourceContents{
			{
				URI:      rh.buildURIFromParsed(parsedURI),
				MimeType: "application/json",
				Text:     string(contentsJSON),
			},
		},
	}
}

// formatAcceptanceCriteriaListResource formats a list of acceptance criteria as a resource response
func (rh *ResourceHandler) formatAcceptanceCriteriaListResource(parsedURI *ParsedURI, acceptanceCriteriaList []models.AcceptanceCriteria) *ResourceResponse {
	acceptanceCriteriaData := make([]interface{}, len(acceptanceCriteriaList))
	for i, ac := range acceptanceCriteriaList {
		acceptanceCriteriaData[i] = map[string]interface{}{
			"id":            ac.ID,
			"reference_id":  ac.ReferenceID,
			"description":   ac.Description,
			"user_story_id": ac.UserStoryID,
			"author_id":     ac.AuthorID,
			"created_at":    ac.CreatedAt,
			"updated_at":    ac.UpdatedAt,
		}
	}

	contents := map[string]interface{}{
		"acceptance_criteria": acceptanceCriteriaData,
		"count":               len(acceptanceCriteriaList),
	}

	contentsJSON, _ := json.Marshal(contents)

	return &ResourceResponse{
		Contents: []ResourceContents{
			{
				URI:      rh.buildURIFromParsed(parsedURI),
				MimeType: "application/json",
				Text:     string(contentsJSON),
			},
		},
	}
}

// buildURIFromParsed rebuilds a URI string from a ParsedURI
func (rh *ResourceHandler) buildURIFromParsed(parsedURI *ParsedURI) string {
	uri, _ := rh.uriParser.BuildURI(parsedURI.Scheme, parsedURI.ReferenceID, parsedURI.SubPath, parsedURI.Parameters)
	return uri
}

// convertRequirementsURI converts requirements:// URIs to the expected format
// Supports both UUID and reference ID formats:
// - requirements://epics/516722c8-4ca6-4cea-b78c-523fc3ea665f
// - requirements://epics/EP-001
// Also supports collection resources:
// - requirements://epics (returns all epics)
func (rh *ResourceHandler) convertRequirementsURI(_ context.Context, uri string) (string, error) {
	// Parse requirements:// URI format: requirements://epics/{id} or requirements://epics
	parts := strings.Split(strings.TrimPrefix(uri, "requirements://"), "/")
	if len(parts) < 1 {
		return "", fmt.Errorf("invalid requirements URI format")
	}

	entityType := parts[0]

	// Handle collection resources (no ID) - return as-is for collection handling
	if len(parts) == 1 {
		return uri, nil // Return the original URI for collection handling
	}

	if len(parts) != 2 {
		return "", fmt.Errorf("invalid requirements URI format")
	}

	entityID := parts[1]

	// Convert based on entity type
	switch entityType {
	case "epics":
		referenceID, err := rh.getEpicReferenceID(entityID)
		if err != nil {
			return "", fmt.Errorf("epic not found: %v", err)
		}
		return fmt.Sprintf("epic://%s", referenceID), nil

	case "user-stories":
		referenceID, err := rh.getUserStoryReferenceID(entityID)
		if err != nil {
			return "", fmt.Errorf("user story not found: %v", err)
		}
		return fmt.Sprintf("user-story://%s", referenceID), nil

	case "requirements":
		referenceID, err := rh.getRequirementReferenceID(entityID)
		if err != nil {
			return "", fmt.Errorf("requirement not found: %v", err)
		}
		return fmt.Sprintf("requirement://%s", referenceID), nil

	case "acceptance-criteria":
		referenceID, err := rh.getAcceptanceCriteriaReferenceID(entityID)
		if err != nil {
			return "", fmt.Errorf("acceptance criteria not found: %v", err)
		}
		return fmt.Sprintf("acceptance-criteria://%s", referenceID), nil

	default:
		return "", fmt.Errorf("unsupported entity type: %s", entityType)
	}
}

// getEpicReferenceID gets epic reference ID from either UUID or reference ID
func (rh *ResourceHandler) getEpicReferenceID(id string) (string, error) {
	// Try to parse as UUID first
	if epicUUID, err := uuid.Parse(id); err == nil {
		// It's a UUID, get by ID
		epic, err := rh.epicService.GetEpicByID(epicUUID)
		if err != nil {
			return "", err
		}
		return epic.ReferenceID, nil
	}

	// Not a UUID, assume it's a reference ID - validate and return
	if rh.isValidReferenceID(id, "EP") {
		// Verify it exists by trying to get it
		_, err := rh.epicService.GetEpicByReferenceID(id)
		if err != nil {
			return "", err
		}
		return id, nil
	}

	return "", fmt.Errorf("invalid epic identifier: %s", id)
}

// getUserStoryReferenceID gets user story reference ID from either UUID or reference ID
func (rh *ResourceHandler) getUserStoryReferenceID(id string) (string, error) {
	// Try to parse as UUID first
	if userStoryUUID, err := uuid.Parse(id); err == nil {
		// It's a UUID, get by ID
		userStory, err := rh.userStoryService.GetUserStoryByID(userStoryUUID)
		if err != nil {
			return "", err
		}
		return userStory.ReferenceID, nil
	}

	// Not a UUID, assume it's a reference ID - validate and return
	if rh.isValidReferenceID(id, "US") {
		// Verify it exists by trying to get it
		_, err := rh.userStoryService.GetUserStoryByReferenceID(id)
		if err != nil {
			return "", err
		}
		return id, nil
	}

	return "", fmt.Errorf("invalid user story identifier: %s", id)
}

// getRequirementReferenceID gets requirement reference ID from either UUID or reference ID
func (rh *ResourceHandler) getRequirementReferenceID(id string) (string, error) {
	// Try to parse as UUID first
	if requirementUUID, err := uuid.Parse(id); err == nil {
		// It's a UUID, get by ID
		requirement, err := rh.requirementService.GetRequirementByID(requirementUUID)
		if err != nil {
			return "", err
		}
		return requirement.ReferenceID, nil
	}

	// Not a UUID, assume it's a reference ID - validate and return
	if rh.isValidReferenceID(id, "REQ") {
		// Verify it exists by trying to get it
		_, err := rh.requirementService.GetRequirementByReferenceID(id)
		if err != nil {
			return "", err
		}
		return id, nil
	}

	return "", fmt.Errorf("invalid requirement identifier: %s", id)
}

// getAcceptanceCriteriaReferenceID gets acceptance criteria reference ID from either UUID or reference ID
func (rh *ResourceHandler) getAcceptanceCriteriaReferenceID(id string) (string, error) {
	// Try to parse as UUID first
	if acUUID, err := uuid.Parse(id); err == nil {
		// It's a UUID, get by ID
		ac, err := rh.acceptanceCriteriaService.GetAcceptanceCriteriaByID(acUUID)
		if err != nil {
			return "", err
		}
		return ac.ReferenceID, nil
	}

	// Not a UUID, assume it's a reference ID - validate and return
	if rh.isValidReferenceID(id, "AC") {
		// Verify it exists by trying to get it
		_, err := rh.acceptanceCriteriaService.GetAcceptanceCriteriaByReferenceID(id)
		if err != nil {
			return "", err
		}
		return id, nil
	}

	return "", fmt.Errorf("invalid acceptance criteria identifier: %s", id)
}

// isValidReferenceID checks if the ID matches the expected reference ID pattern for the given prefix
func (rh *ResourceHandler) isValidReferenceID(id, expectedPrefix string) bool {
	if !strings.HasPrefix(id, expectedPrefix+"-") {
		return false
	}

	// Use the URI parser's validation logic
	return rh.uriParser.isValidReferenceID(id)
}

// isCollectionResource checks if the URI is a collection resource (no specific ID)
func (rh *ResourceHandler) isCollectionResource(uri string) bool {
	// Parse requirements:// URI format: requirements://epics/{id} or requirements://epics
	parts := strings.Split(strings.TrimPrefix(uri, "requirements://"), "/")

	// Collection resource has only one part (entity type, no ID)
	// Special case: "prompts/active" is also a collection resource
	if len(parts) == 2 && parts[0] == "prompts" && parts[1] == "active" {
		return true
	}

	return len(parts) == 1 && parts[0] != ""
}

// handleCollectionResource handles collection resources like requirements://epics
func (rh *ResourceHandler) handleCollectionResource(ctx context.Context, uri string) (interface{}, error) {
	// Parse the entity type from the URI
	entityType := strings.TrimPrefix(uri, "requirements://")

	// Handle special case for prompts/active
	if entityType == "prompts/active" {
		return rh.handleActivePromptResource(ctx, uri)
	}

	switch entityType {
	case "epics":
		return rh.handleEpicsCollection(ctx, uri)
	case "user-stories":
		return rh.handleUserStoriesCollection(ctx, uri)
	case "requirements":
		return rh.handleRequirementsCollection(ctx, uri)
	case "acceptance-criteria":
		return rh.handleAcceptanceCriteriaCollection(ctx, uri)
	case "prompts":
		return rh.handlePromptsCollection(ctx, uri)
	default:
		return nil, jsonrpc.NewInvalidParamsError(fmt.Sprintf("Unsupported collection type: %s", entityType))
	}
}

// handleEpicsCollection handles requirements://epics collection resource
func (rh *ResourceHandler) handleEpicsCollection(_ context.Context, uri string) (interface{}, error) {
	// Get all epics using the existing ListEpics method with no filters
	epics, _, err := rh.epicService.ListEpics(service.EpicFilters{
		Limit: 1000, // Set a reasonable limit for collection resources
	})
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get epics: %v", err))
	}

	// Format as collection resource
	return rh.formatEpicsCollectionResource(uri, epics), nil
}

// handleUserStoriesCollection handles requirements://user-stories collection resource
func (rh *ResourceHandler) handleUserStoriesCollection(_ context.Context, uri string) (interface{}, error) {
	// Get all user stories using the existing ListUserStories method with no filters
	userStories, _, err := rh.userStoryService.ListUserStories(service.UserStoryFilters{
		Limit: 1000, // Set a reasonable limit for collection resources
	})
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user stories: %v", err))
	}

	// Format as collection resource
	return rh.formatUserStoriesCollectionResource(uri, userStories), nil
}

// handleRequirementsCollection handles requirements://requirements collection resource
func (rh *ResourceHandler) handleRequirementsCollection(_ context.Context, uri string) (interface{}, error) {
	// Get all requirements using the existing ListRequirements method with no filters
	requirements, _, err := rh.requirementService.ListRequirements(service.RequirementFilters{
		Limit: 1000, // Set a reasonable limit for collection resources
	})
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get requirements: %v", err))
	}

	// Format as collection resource
	return rh.formatRequirementsCollectionResource(uri, requirements), nil
}

// handleAcceptanceCriteriaCollection handles requirements://acceptance-criteria collection resource
func (rh *ResourceHandler) handleAcceptanceCriteriaCollection(_ context.Context, uri string) (interface{}, error) {
	// Get all acceptance criteria using the existing ListAcceptanceCriteria method with no filters
	acceptanceCriteria, _, err := rh.acceptanceCriteriaService.ListAcceptanceCriteria(service.AcceptanceCriteriaFilters{
		Limit: 1000, // Set a reasonable limit for collection resources
	})
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get acceptance criteria: %v", err))
	}

	// Format as collection resource
	return rh.formatAcceptanceCriteriaCollectionResource(uri, acceptanceCriteria), nil
}

// Collection formatting methods

// formatEpicsCollectionResource formats a collection of epics as a resource response
func (rh *ResourceHandler) formatEpicsCollectionResource(uri string, epics []models.Epic) *ResourceResponse {
	epicsData := make([]any, len(epics))
	for i, epic := range epics {
		epicsData[i] = map[string]any{
			"id":           epic.ID,
			"reference_id": epic.ReferenceID,
			"title":        epic.Title,
			"description":  epic.Description,
			"status":       epic.Status,
			"priority":     epic.Priority,
			"creator_id":   epic.CreatorID,
			"assignee_id":  epic.AssigneeID,
			"created_at":   epic.CreatedAt,
			"updated_at":   epic.UpdatedAt,
		}
	}

	contents := map[string]any{
		"epics": epicsData,
		"count": len(epics),
	}

	contentsJSON, _ := json.Marshal(contents)

	return &ResourceResponse{
		Contents: []ResourceContents{
			{
				URI:      uri,
				MimeType: "application/json",
				Text:     string(contentsJSON),
			},
		},
	}
}

// formatUserStoriesCollectionResource formats a collection of user stories as a resource response
func (rh *ResourceHandler) formatUserStoriesCollectionResource(uri string, userStories []models.UserStory) *ResourceResponse {
	userStoriesData := make([]any, len(userStories))
	for i, us := range userStories {
		userStoriesData[i] = map[string]any{
			"id":           us.ID,
			"reference_id": us.ReferenceID,
			"title":        us.Title,
			"description":  us.Description,
			"status":       us.Status,
			"priority":     us.Priority,
			"epic_id":      us.EpicID,
			"creator_id":   us.CreatorID,
			"assignee_id":  us.AssigneeID,
			"created_at":   us.CreatedAt,
			"updated_at":   us.UpdatedAt,
		}
	}

	contents := map[string]any{
		"user_stories": userStoriesData,
		"count":        len(userStories),
	}

	contentsJSON, _ := json.Marshal(contents)

	return &ResourceResponse{
		Contents: []ResourceContents{
			{
				URI:      uri,
				MimeType: "application/json",
				Text:     string(contentsJSON),
			},
		},
	}
}

// formatRequirementsCollectionResource formats a collection of requirements as a resource response
func (rh *ResourceHandler) formatRequirementsCollectionResource(uri string, requirements []models.Requirement) *ResourceResponse {
	requirementsData := make([]any, len(requirements))
	for i, req := range requirements {
		requirementsData[i] = map[string]any{
			"id":                     req.ID,
			"reference_id":           req.ReferenceID,
			"title":                  req.Title,
			"description":            req.Description,
			"status":                 req.Status,
			"priority":               req.Priority,
			"user_story_id":          req.UserStoryID,
			"acceptance_criteria_id": req.AcceptanceCriteriaID,
			"type_id":                req.TypeID,
			"creator_id":             req.CreatorID,
			"assignee_id":            req.AssigneeID,
			"created_at":             req.CreatedAt,
			"updated_at":             req.UpdatedAt,
		}
	}

	contents := map[string]any{
		"requirements": requirementsData,
		"count":        len(requirements),
	}

	contentsJSON, _ := json.Marshal(contents)

	return &ResourceResponse{
		Contents: []ResourceContents{
			{
				URI:      uri,
				MimeType: "application/json",
				Text:     string(contentsJSON),
			},
		},
	}
}

// formatAcceptanceCriteriaCollectionResource formats a collection of acceptance criteria as a resource response
func (rh *ResourceHandler) formatAcceptanceCriteriaCollectionResource(uri string, acceptanceCriteria []models.AcceptanceCriteria) *ResourceResponse {
	acceptanceCriteriaData := make([]any, len(acceptanceCriteria))
	for i, ac := range acceptanceCriteria {
		acceptanceCriteriaData[i] = map[string]any{
			"id":            ac.ID,
			"reference_id":  ac.ReferenceID,
			"description":   ac.Description,
			"user_story_id": ac.UserStoryID,
			"author_id":     ac.AuthorID,
			"created_at":    ac.CreatedAt,
			"updated_at":    ac.UpdatedAt,
		}
	}

	contents := map[string]any{
		"acceptance_criteria": acceptanceCriteriaData,
		"count":               len(acceptanceCriteria),
	}

	contentsJSON, _ := json.Marshal(contents)

	return &ResourceResponse{
		Contents: []ResourceContents{
			{
				URI:      uri,
				MimeType: "application/json",
				Text:     string(contentsJSON),
			},
		},
	}
}

// handlePromptResource handles prompt:// URI resources
func (rh *ResourceHandler) handlePromptResource(ctx context.Context, parsedURI *ParsedURI) (interface{}, error) {
	// Get the prompt by reference ID
	prompt, err := rh.promptService.GetByReferenceID(ctx, parsedURI.ReferenceID)
	if err != nil {
		if err == service.ErrNotFound {
			return nil, jsonrpc.NewJSONRPCError(-32002, "Prompt not found", nil)
		}
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get prompt: %v", err))
	}

	// Format the prompt content
	content := map[string]interface{}{
		"id":           prompt.ID,
		"reference_id": prompt.ReferenceID,
		"name":         prompt.Name,
		"title":        prompt.Title,
		"content":      prompt.Content,
		"is_active":    prompt.IsActive,
		"created_at":   prompt.CreatedAt,
		"updated_at":   prompt.UpdatedAt,
	}

	if prompt.Description != nil {
		content["description"] = *prompt.Description
	}

	contentsJSON, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to marshal prompt: %v", err))
	}

	uri := fmt.Sprintf("prompt://%s", parsedURI.ReferenceID)
	return &ResourceResponse{
		Contents: []ResourceContents{
			{
				URI:      uri,
				MimeType: "application/json",
				Text:     string(contentsJSON),
			},
		},
	}, nil
}

// handlePromptsCollection handles requirements://prompts collection resource
func (rh *ResourceHandler) handlePromptsCollection(ctx context.Context, uri string) (interface{}, error) {
	// Get all prompts using the PromptService
	prompts, _, err := rh.promptService.List(ctx, 100, 0, nil)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get prompts: %v", err))
	}

	// Format the collection
	collection := map[string]interface{}{
		"type":        "prompts_collection",
		"total_count": len(prompts),
		"prompts":     prompts,
		"description": "Collection of all system prompts",
	}

	contentsJSON, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to marshal prompts collection: %v", err))
	}

	return &ResourceResponse{
		Contents: []ResourceContents{
			{
				URI:      uri,
				MimeType: "application/json",
				Text:     string(contentsJSON),
			},
		},
	}, nil
}

// handleActivePromptResource handles requirements://prompts/active resource
func (rh *ResourceHandler) handleActivePromptResource(ctx context.Context, uri string) (interface{}, error) {
	// Get the active prompt
	prompt, err := rh.promptService.GetActive(ctx)
	if err != nil {
		if err == service.ErrNotFound {
			// Return empty response if no active prompt
			collection := map[string]interface{}{
				"type":        "active_prompt",
				"active":      false,
				"prompt":      nil,
				"description": "No active system prompt found",
			}

			contentsJSON, err := json.MarshalIndent(collection, "", "  ")
			if err != nil {
				return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to marshal empty active prompt: %v", err))
			}

			return &ResourceResponse{
				Contents: []ResourceContents{
					{
						URI:      uri,
						MimeType: "application/json",
						Text:     string(contentsJSON),
					},
				},
			}, nil
		}
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get active prompt: %v", err))
	}

	// Format the active prompt
	collection := map[string]interface{}{
		"type":        "active_prompt",
		"active":      true,
		"prompt":      prompt,
		"description": fmt.Sprintf("Currently active system prompt: %s", prompt.Title),
	}

	contentsJSON, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to marshal active prompt: %v", err))
	}

	return &ResourceResponse{
		Contents: []ResourceContents{
			{
				URI:      uri,
				MimeType: "application/json",
				Text:     string(contentsJSON),
			},
		},
	}, nil
}
