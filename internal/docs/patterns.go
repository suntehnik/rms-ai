package docs

// APIPatterns documents common patterns used throughout the API
// This file provides comprehensive documentation for consistent API usage patterns

// PaginationPattern documents the standard pagination approach used across list endpoints
type PaginationPattern struct {
	// Standard pagination parameters used in query strings
	Limit  int `json:"limit" example:"25" minimum:"1" maximum:"100" description:"Maximum number of results to return"`
	Offset int `json:"offset" example:"0" minimum:"0" description:"Number of results to skip for pagination"`

	// Standard pagination response metadata
	Total int `json:"total" example:"150" description:"Total number of available results"`
	Count int `json:"count" example:"25" description:"Number of results in current response"`
}

// FilteringPattern documents common filtering parameters available across endpoints
type FilteringPattern struct {
	// User-based filtering
	CreatorID  string `json:"creator_id,omitempty" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174001" description:"Filter by entity creator UUID"`
	AssigneeID string `json:"assignee_id,omitempty" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174002" description:"Filter by assigned user UUID"`

	// Status and priority filtering
	Status   string `json:"status,omitempty" example:"in_progress" description:"Filter by entity status (backlog, ready, in_progress, done, cancelled)"`
	Priority int    `json:"priority,omitempty" example:"1" minimum:"1" maximum:"4" description:"Filter by priority level (1=Critical, 2=High, 3=Medium, 4=Low)"`

	// Date range filtering
	CreatedAfter  string `json:"created_after,omitempty" format:"date-time" example:"2023-01-01T00:00:00Z" description:"Filter entities created after this date"`
	CreatedBefore string `json:"created_before,omitempty" format:"date-time" example:"2023-12-31T23:59:59Z" description:"Filter entities created before this date"`
	UpdatedAfter  string `json:"updated_after,omitempty" format:"date-time" example:"2023-01-01T00:00:00Z" description:"Filter entities updated after this date"`
	UpdatedBefore string `json:"updated_before,omitempty" format:"date-time" example:"2023-12-31T23:59:59Z" description:"Filter entities updated before this date"`
}

// SortingPattern documents standard sorting options available across endpoints
type SortingPattern struct {
	OrderBy string `json:"order_by,omitempty" example:"created_at DESC" description:"Sort field and direction. Format: 'field_name ASC|DESC'. Available fields: created_at, updated_at, title, priority, reference_id"`
}

// SearchPattern documents the advanced search capabilities
type SearchPattern struct {
	// Basic search parameters
	Query string `json:"query" example:"authentication system" description:"Search query for full-text search across titles, descriptions, and content"`
	Limit int    `json:"limit,omitempty" example:"20" minimum:"1" maximum:"100" description:"Maximum number of search results"`

	// Entity type filtering
	EntityTypes []string `json:"entity_types,omitempty" example:"epics,user_stories" description:"Filter by entity types: epics, user_stories, acceptance_criteria, requirements, comments"`

	// Search result sorting
	SortBy    string `json:"sort_by,omitempty" example:"relevance" description:"Sort search results by: relevance, created_at, updated_at, title, priority"`
	SortOrder string `json:"sort_order,omitempty" example:"desc" description:"Sort order: asc or desc"`

	// Advanced filtering (inherits from FilteringPattern)
	FilteringPattern
}

// ErrorPattern documents the standard error response structure
type ErrorPattern struct {
	// Standard error response format
	Error   string `json:"error" example:"Validation failed" description:"Human-readable error message"`
	Code    string `json:"code,omitempty" example:"VALIDATION_ERROR" description:"Machine-readable error code"`
	Details string `json:"details,omitempty" example:"Title field is required and cannot be empty" description:"Additional error details and context"`

	// Validation error details
	ValidationErrors []ValidationError `json:"validation_errors,omitempty" description:"Detailed validation errors for form fields"`
}

// ValidationError provides detailed field-level validation errors
type ValidationError struct {
	Field   string `json:"field" example:"title" description:"Field name that failed validation"`
	Message string `json:"message" example:"Title is required and cannot be empty" description:"Validation error message"`
	Code    string `json:"code" example:"REQUIRED_FIELD" description:"Validation error code"`
}

// RelationshipPattern documents how entity relationships work
type RelationshipPattern struct {
	// Hierarchical relationships
	EpicToUserStories      string `json:"epic_to_user_stories" example:"One-to-Many" description:"Epics contain multiple user stories"`
	UserStoryToAcceptance  string `json:"user_story_to_acceptance" example:"One-to-Many" description:"User stories have multiple acceptance criteria"`
	UserStoryToRequirement string `json:"user_story_to_requirements" example:"One-to-Many" description:"User stories contain multiple requirements"`

	// Cross-entity relationships
	RequirementRelationships []string `json:"requirement_relationships" example:"depends_on,blocks,relates_to,conflicts_with,derives_from" description:"Available requirement relationship types"`

	// Comment relationships
	EntityComments string `json:"entity_comments" example:"Polymorphic" description:"Comments can be attached to any entity type (epics, user_stories, requirements, acceptance_criteria)"`
	InlineComments string `json:"inline_comments" example:"Position-based" description:"Inline comments link to specific text positions within entity descriptions"`
}

// AuthenticationPattern documents JWT authentication usage
type AuthenticationPattern struct {
	// Token format
	TokenFormat string `json:"token_format" example:"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." description:"JWT token with Bearer prefix in Authorization header"`
	Expiration  string `json:"expiration" example:"1 hour" description:"Token expiration time"`

	// User roles and permissions
	Roles []RolePermission `json:"roles" description:"Available user roles and their permissions"`
}

// RolePermission documents role-based access control
type RolePermission struct {
	Role        string   `json:"role" example:"administrator" description:"User role name"`
	Description string   `json:"description" example:"Full system access including configuration management" description:"Role description"`
	Permissions []string `json:"permissions" example:"create,read,update,delete,configure" description:"Available permissions for this role"`
}

// StatusWorkflowPattern documents status management across entities
type StatusWorkflowPattern struct {
	// Common status values
	CommonStatuses []string `json:"common_statuses" example:"backlog,ready,in_progress,done,cancelled" description:"Standard status values used across entity types"`

	// Status transitions
	AllowedTransitions map[string][]string `json:"allowed_transitions" example:"{\"backlog\":[\"ready\",\"cancelled\"],\"ready\":[\"in_progress\",\"backlog\"]}" description:"Valid status transitions for workflow management"`

	// Configurable workflows
	CustomWorkflows string `json:"custom_workflows" example:"Configurable via /api/v1/config/status-models" description:"Status workflows can be customized per entity type"`
}

// ReferenceIDPattern documents the human-readable ID system
type ReferenceIDPattern struct {
	// Format patterns
	EpicFormat               string `json:"epic_format" example:"EP-001, EP-002, ..." description:"Epic reference ID format"`
	UserStoryFormat          string `json:"user_story_format" example:"US-001, US-002, ..." description:"User story reference ID format"`
	AcceptanceCriteriaFormat string `json:"acceptance_criteria_format" example:"AC-001, AC-002, ..." description:"Acceptance criteria reference ID format"`
	RequirementFormat        string `json:"requirement_format" example:"REQ-001, REQ-002, ..." description:"Requirement reference ID format"`

	// Usage patterns
	AutoGeneration string `json:"auto_generation" example:"Automatic" description:"Reference IDs are automatically generated on entity creation"`
	Uniqueness     string `json:"uniqueness" example:"Global" description:"Reference IDs are globally unique across the system"`
	Searchable     string `json:"searchable" example:"Yes" description:"Reference IDs can be used in search queries and API endpoints"`
}

// GetAPIPatternDocumentation returns comprehensive API pattern documentation
func GetAPIPatternDocumentation() map[string]interface{} {
	return map[string]interface{}{
		"pagination": map[string]interface{}{
			"limit":  25,
			"offset": 0,
			"total":  150,
			"count":  25,
		},
		"filtering": map[string]interface{}{
			"creator_id":     "123e4567-e89b-12d3-a456-426614174001",
			"assignee_id":    "123e4567-e89b-12d3-a456-426614174002",
			"status":         "in_progress",
			"priority":       1,
			"created_after":  "2023-01-01T00:00:00Z",
			"created_before": "2023-12-31T23:59:59Z",
		},
		"sorting": map[string]interface{}{
			"order_by": "created_at DESC",
		},
		"search": map[string]interface{}{
			"query":        "authentication system",
			"limit":        20,
			"entity_types": []string{"epics", "user_stories"},
			"sort_by":      "relevance",
			"sort_order":   "desc",
		},
		"errors": map[string]interface{}{
			"error":   "Validation failed",
			"code":    "VALIDATION_ERROR",
			"details": "Title field is required and cannot be empty",
			"validation_errors": []map[string]interface{}{
				{
					"field":   "title",
					"message": "Title is required and cannot be empty",
					"code":    "REQUIRED_FIELD",
				},
			},
		},
		"relationships": map[string]interface{}{
			"epic_to_user_stories":      "One-to-Many",
			"user_story_to_acceptance":  "One-to-Many",
			"user_story_to_requirement": "One-to-Many",
			"requirement_relationships": []string{"depends_on", "blocks", "relates_to", "conflicts_with", "derives_from"},
			"entity_comments":           "Polymorphic",
			"inline_comments":           "Position-based",
		},
		"authentication": map[string]interface{}{
			"token_format": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			"expiration":   "1 hour",
			"roles": []map[string]interface{}{
				{
					"role":        "administrator",
					"description": "Full system access including configuration management",
					"permissions": []string{"create", "read", "update", "delete", "configure"},
				},
				{
					"role":        "user",
					"description": "Entity management and assignment capabilities",
					"permissions": []string{"create", "read", "update", "delete"},
				},
				{
					"role":        "commenter",
					"description": "Read access and comment creation only",
					"permissions": []string{"read", "comment"},
				},
			},
		},
		"status_workflow": map[string]interface{}{
			"common_statuses": []string{"backlog", "ready", "in_progress", "done", "cancelled"},
			"allowed_transitions": map[string][]string{
				"backlog":     {"ready", "cancelled"},
				"ready":       {"in_progress", "backlog"},
				"in_progress": {"done", "ready"},
				"done":        {"ready"},
				"cancelled":   {"backlog"},
			},
			"custom_workflows": "Configurable via /api/v1/config/status-models",
		},
		"reference_ids": map[string]interface{}{
			"epic_format":                "EP-001, EP-002, ...",
			"user_story_format":          "US-001, US-002, ...",
			"acceptance_criteria_format": "AC-001, AC-002, ...",
			"requirement_format":         "REQ-001, REQ-002, ...",
			"auto_generation":            "Automatic",
			"uniqueness":                 "Global",
			"searchable":                 "Yes",
		},
	}
}

// GetUsageExamples returns comprehensive usage examples for common API operations
func GetUsageExamples() map[string]interface{} {
	return map[string]interface{}{
		"create_epic_workflow": map[string]interface{}{
			"description": "Complete workflow for creating an epic with user stories and requirements",
			"steps": []map[string]interface{}{
				{
					"step":        1,
					"action":      "Create Epic",
					"method":      "POST",
					"endpoint":    "/api/v1/epics",
					"description": "Create the main epic container",
				},
				{
					"step":        2,
					"action":      "Create User Stories",
					"method":      "POST",
					"endpoint":    "/api/v1/epics/{epic_id}/user-stories",
					"description": "Add user stories to the epic",
				},
				{
					"step":        3,
					"action":      "Add Acceptance Criteria",
					"method":      "POST",
					"endpoint":    "/api/v1/user-stories/{user_story_id}/acceptance-criteria",
					"description": "Define testable conditions for user stories",
				},
				{
					"step":        4,
					"action":      "Create Requirements",
					"method":      "POST",
					"endpoint":    "/api/v1/user-stories/{user_story_id}/requirements",
					"description": "Add detailed technical requirements",
				},
			},
		},
		"search_examples": map[string]interface{}{
			"basic_search": map[string]interface{}{
				"endpoint":    "/api/v1/search?query=authentication&limit=10",
				"description": "Basic full-text search across all entities",
			},
			"filtered_search": map[string]interface{}{
				"endpoint":    "/api/v1/search?query=login&entity_types=user_stories,requirements&status=in_progress&priority=1",
				"description": "Search with entity type and status filtering",
			},
			"date_range_search": map[string]interface{}{
				"endpoint":    "/api/v1/search?query=security&created_after=2023-01-01T00:00:00Z&created_before=2023-12-31T23:59:59Z",
				"description": "Search within specific date range",
			},
		},
		"pagination_examples": map[string]interface{}{
			"first_page": map[string]interface{}{
				"endpoint":    "/api/v1/epics?limit=25&offset=0",
				"description": "Get first page of results",
			},
			"second_page": map[string]interface{}{
				"endpoint":    "/api/v1/epics?limit=25&offset=25",
				"description": "Get second page of results",
			},
			"large_page": map[string]interface{}{
				"endpoint":    "/api/v1/epics?limit=100&offset=0",
				"description": "Get maximum page size (100 items)",
			},
		},
		"relationship_examples": map[string]interface{}{
			"create_dependency": map[string]interface{}{
				"endpoint":    "/api/v1/requirements/relationships",
				"method":      "POST",
				"description": "Create a dependency relationship between requirements",
				"body": map[string]interface{}{
					"source_requirement_id": "123e4567-e89b-12d3-a456-426614174030",
					"target_requirement_id": "123e4567-e89b-12d3-a456-426614174031",
					"relationship_type_id":  "depends_on_type_id",
					"description":           "Authentication must be implemented before authorization",
				},
			},
			"get_hierarchy": map[string]interface{}{
				"endpoint":    "/api/v1/hierarchy/epics/{epic_id}",
				"description": "Get complete hierarchy for an epic including all nested entities",
			},
		},
	}
}
