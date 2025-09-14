package docs

// APIOrganization provides comprehensive documentation of API endpoint organization
// This file documents the logical grouping and structure of all API endpoints

// EndpointGroup represents a logical group of related endpoints
type EndpointGroup struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	BasePath    string     `json:"base_path"`
	Endpoints   []Endpoint `json:"endpoints"`
}

// Endpoint represents a single API endpoint with its details
type Endpoint struct {
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Summary     string            `json:"summary"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Parameters  []Parameter       `json:"parameters,omitempty"`
	Responses   map[string]string `json:"responses"`
}

// Parameter represents an endpoint parameter
type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"` // query, path, body, header
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Example     string `json:"example,omitempty"`
}

// GetAPIOrganization returns the complete API organization structure
func GetAPIOrganization() []EndpointGroup {
	return []EndpointGroup{
		{
			Name:        "Epic Management",
			Description: "Endpoints for managing epics - high-level features and initiatives that serve as containers for user stories. Epics provide project-level organization and tracking.",
			BasePath:    "/api/v1/epics",
			Endpoints: []Endpoint{
				{
					Method:      "POST",
					Path:        "/api/v1/epics",
					Summary:     "Create new epic",
					Description: "Create a new epic with title, description, and priority. Epic will be assigned a unique reference ID (e.g., EP-001).",
					Tags:        []string{"epics"},
					Parameters: []Parameter{
						{Name: "epic", In: "body", Type: "object", Required: true, Description: "Epic creation request"},
					},
					Responses: map[string]string{
						"201": "Epic created successfully",
						"400": "Invalid request data",
						"401": "Authentication required",
						"500": "Internal server error",
					},
				},
				{
					Method:      "GET",
					Path:        "/api/v1/epics",
					Summary:     "List epics with filtering",
					Description: "Retrieve paginated list of epics with optional filtering by creator, assignee, status, and priority.",
					Tags:        []string{"epics"},
					Parameters: []Parameter{
						{Name: "limit", In: "query", Type: "integer", Required: false, Description: "Maximum results (1-100)", Example: "25"},
						{Name: "offset", In: "query", Type: "integer", Required: false, Description: "Results to skip", Example: "0"},
						{Name: "creator_id", In: "query", Type: "string", Required: false, Description: "Filter by creator UUID"},
						{Name: "status", In: "query", Type: "string", Required: false, Description: "Filter by status"},
						{Name: "priority", In: "query", Type: "integer", Required: false, Description: "Filter by priority (1-4)"},
					},
					Responses: map[string]string{
						"200": "List of epics with pagination metadata",
						"500": "Internal server error",
					},
				},
				{
					Method:      "GET",
					Path:        "/api/v1/epics/{id}",
					Summary:     "Get epic by ID",
					Description: "Retrieve specific epic by UUID or reference ID (e.g., EP-001).",
					Tags:        []string{"epics"},
					Parameters: []Parameter{
						{Name: "id", In: "path", Type: "string", Required: true, Description: "Epic UUID or reference ID"},
					},
					Responses: map[string]string{
						"200": "Epic details",
						"404": "Epic not found",
						"500": "Internal server error",
					},
				},
				{
					Method:      "PUT",
					Path:        "/api/v1/epics/{id}",
					Summary:     "Update epic",
					Description: "Update epic properties including title, description, priority, status, and assignee.",
					Tags:        []string{"epics"},
					Parameters: []Parameter{
						{Name: "id", In: "path", Type: "string", Required: true, Description: "Epic UUID"},
						{Name: "epic", In: "body", Type: "object", Required: true, Description: "Epic update request"},
					},
					Responses: map[string]string{
						"200": "Epic updated successfully",
						"400": "Invalid request data",
						"404": "Epic not found",
						"500": "Internal server error",
					},
				},
				{
					Method:      "GET",
					Path:        "/api/v1/epics/{id}/user-stories",
					Summary:     "Get epic with user stories",
					Description: "Retrieve epic with all associated user stories in hierarchical structure.",
					Tags:        []string{"epics", "user-stories"},
					Parameters: []Parameter{
						{Name: "id", In: "path", Type: "string", Required: true, Description: "Epic UUID or reference ID"},
					},
					Responses: map[string]string{
						"200": "Epic with nested user stories",
						"404": "Epic not found",
						"500": "Internal server error",
					},
				},
			},
		},
		{
			Name:        "User Story Management",
			Description: "Endpoints for managing user stories within epics. User stories represent feature requirements from the user perspective and serve as containers for acceptance criteria and detailed requirements.",
			BasePath:    "/api/v1/user-stories",
			Endpoints: []Endpoint{
				{
					Method:      "POST",
					Path:        "/api/v1/user-stories",
					Summary:     "Create user story",
					Description: "Create a new user story with title, description, and epic association. Automatically assigned reference ID (e.g., US-001).",
					Tags:        []string{"user-stories"},
					Parameters: []Parameter{
						{Name: "user_story", In: "body", Type: "object", Required: true, Description: "User story creation request"},
					},
					Responses: map[string]string{
						"201": "User story created successfully",
						"400": "Invalid request data",
						"401": "Authentication required",
						"500": "Internal server error",
					},
				},
				{
					Method:      "POST",
					Path:        "/api/v1/epics/{id}/user-stories",
					Summary:     "Create user story in epic",
					Description: "Create a new user story directly within a specific epic context.",
					Tags:        []string{"epics", "user-stories"},
					Parameters: []Parameter{
						{Name: "id", In: "path", Type: "string", Required: true, Description: "Epic UUID or reference ID"},
						{Name: "user_story", In: "body", Type: "object", Required: true, Description: "User story creation request"},
					},
					Responses: map[string]string{
						"201": "User story created in epic",
						"400": "Invalid request data",
						"404": "Epic not found",
						"500": "Internal server error",
					},
				},
				{
					Method:      "GET",
					Path:        "/api/v1/user-stories/{id}/acceptance-criteria",
					Summary:     "Get user story acceptance criteria",
					Description: "Retrieve all acceptance criteria associated with a user story.",
					Tags:        []string{"user-stories", "acceptance-criteria"},
					Parameters: []Parameter{
						{Name: "id", In: "path", Type: "string", Required: true, Description: "User story UUID or reference ID"},
					},
					Responses: map[string]string{
						"200": "List of acceptance criteria",
						"404": "User story not found",
						"500": "Internal server error",
					},
				},
			},
		},
		{
			Name:        "Acceptance Criteria Management",
			Description: "Endpoints for managing acceptance criteria within user stories. Acceptance criteria define testable conditions that must be met for user story completion, typically using EARS format (Easy Approach to Requirements Syntax).",
			BasePath:    "/api/v1/acceptance-criteria",
			Endpoints: []Endpoint{
				{
					Method:      "POST",
					Path:        "/api/v1/user-stories/{id}/acceptance-criteria",
					Summary:     "Create acceptance criteria",
					Description: "Create new acceptance criteria for a user story with testable conditions.",
					Tags:        []string{"acceptance-criteria"},
					Parameters: []Parameter{
						{Name: "id", In: "path", Type: "string", Required: true, Description: "User story UUID or reference ID"},
						{Name: "acceptance_criteria", In: "body", Type: "object", Required: true, Description: "Acceptance criteria creation request"},
					},
					Responses: map[string]string{
						"201": "Acceptance criteria created",
						"400": "Invalid request data",
						"404": "User story not found",
						"500": "Internal server error",
					},
				},
				{
					Method:      "GET",
					Path:        "/api/v1/acceptance-criteria",
					Summary:     "List acceptance criteria",
					Description: "Retrieve paginated list of acceptance criteria with filtering options.",
					Tags:        []string{"acceptance-criteria"},
					Parameters: []Parameter{
						{Name: "user_story_id", In: "query", Type: "string", Required: false, Description: "Filter by user story UUID"},
						{Name: "author_id", In: "query", Type: "string", Required: false, Description: "Filter by author UUID"},
						{Name: "limit", In: "query", Type: "integer", Required: false, Description: "Maximum results"},
						{Name: "offset", In: "query", Type: "integer", Required: false, Description: "Results to skip"},
					},
					Responses: map[string]string{
						"200": "List of acceptance criteria",
						"500": "Internal server error",
					},
				},
			},
		},
		{
			Name:        "Requirement Management",
			Description: "Endpoints for managing detailed technical requirements. Requirements provide specific implementation details and can be linked with various relationship types (depends_on, blocks, relates_to, conflicts_with, derives_from).",
			BasePath:    "/api/v1/requirements",
			Endpoints: []Endpoint{
				{
					Method:      "POST",
					Path:        "/api/v1/requirements",
					Summary:     "Create requirement",
					Description: "Create a new detailed requirement with title, description, type, and relationships.",
					Tags:        []string{"requirements"},
					Parameters: []Parameter{
						{Name: "requirement", In: "body", Type: "object", Required: true, Description: "Requirement creation request"},
					},
					Responses: map[string]string{
						"201": "Requirement created successfully",
						"400": "Invalid request data",
						"401": "Authentication required",
						"500": "Internal server error",
					},
				},
				{
					Method:      "GET",
					Path:        "/api/v1/requirements/{id}/relationships",
					Summary:     "Get requirement relationships",
					Description: "Retrieve all relationships for a requirement including dependencies, blocks, and related requirements.",
					Tags:        []string{"requirements"},
					Parameters: []Parameter{
						{Name: "id", In: "path", Type: "string", Required: true, Description: "Requirement UUID or reference ID"},
					},
					Responses: map[string]string{
						"200": "Requirement with relationships",
						"404": "Requirement not found",
						"500": "Internal server error",
					},
				},
				{
					Method:      "POST",
					Path:        "/api/v1/requirements/relationships",
					Summary:     "Create requirement relationship",
					Description: "Create a relationship between two requirements (depends_on, blocks, relates_to, etc.).",
					Tags:        []string{"requirements"},
					Parameters: []Parameter{
						{Name: "relationship", In: "body", Type: "object", Required: true, Description: "Relationship creation request"},
					},
					Responses: map[string]string{
						"201": "Relationship created successfully",
						"400": "Invalid request data",
						"409": "Circular dependency detected",
						"500": "Internal server error",
					},
				},
			},
		},
		{
			Name:        "Search & Discovery",
			Description: "Endpoints for full-text search across all entities with advanced filtering, sorting, and suggestion capabilities. Provides efficient content discovery and navigation.",
			BasePath:    "/api/v1/search",
			Endpoints: []Endpoint{
				{
					Method:      "GET",
					Path:        "/api/v1/search",
					Summary:     "Full-text search",
					Description: "Search across all entities with advanced filtering by type, status, priority, dates, and more.",
					Tags:        []string{"search"},
					Parameters: []Parameter{
						{Name: "query", In: "query", Type: "string", Required: true, Description: "Search query text"},
						{Name: "entity_types", In: "query", Type: "array", Required: false, Description: "Filter by entity types"},
						{Name: "status", In: "query", Type: "string", Required: false, Description: "Filter by status"},
						{Name: "priority", In: "query", Type: "integer", Required: false, Description: "Filter by priority"},
						{Name: "limit", In: "query", Type: "integer", Required: false, Description: "Maximum results"},
						{Name: "sort_by", In: "query", Type: "string", Required: false, Description: "Sort by field (relevance, created_at, etc.)"},
					},
					Responses: map[string]string{
						"200": "Search results with metadata",
						"400": "Invalid search parameters",
						"500": "Internal server error",
					},
				},
				{
					Method:      "GET",
					Path:        "/api/v1/search/suggestions",
					Summary:     "Get search suggestions",
					Description: "Get search suggestions based on partial query input for autocomplete functionality.",
					Tags:        []string{"search"},
					Parameters: []Parameter{
						{Name: "query", In: "query", Type: "string", Required: true, Description: "Partial search query (minimum 2 characters)"},
						{Name: "limit", In: "query", Type: "integer", Required: false, Description: "Maximum suggestions per category"},
					},
					Responses: map[string]string{
						"200": "Search suggestions grouped by category",
						"400": "Invalid parameters",
						"500": "Internal server error",
					},
				},
			},
		},
		{
			Name:        "Comment System",
			Description: "Endpoints for managing comments and collaboration. Supports both general comments and inline comments with threading, resolution tracking, and entity associations.",
			BasePath:    "/api/v1/comments",
			Endpoints: []Endpoint{
				{
					Method:      "POST",
					Path:        "/api/v1/{entity_type}/{id}/comments",
					Summary:     "Create entity comment",
					Description: "Create a general comment on any entity (epic, user story, requirement, acceptance criteria).",
					Tags:        []string{"comments"},
					Parameters: []Parameter{
						{Name: "entity_type", In: "path", Type: "string", Required: true, Description: "Entity type (epics, user-stories, requirements, acceptance-criteria)"},
						{Name: "id", In: "path", Type: "string", Required: true, Description: "Entity UUID or reference ID"},
						{Name: "comment", In: "body", Type: "object", Required: true, Description: "Comment creation request"},
					},
					Responses: map[string]string{
						"201": "Comment created successfully",
						"400": "Invalid request data",
						"404": "Entity not found",
						"500": "Internal server error",
					},
				},
				{
					Method:      "POST",
					Path:        "/api/v1/{entity_type}/{id}/comments/inline",
					Summary:     "Create inline comment",
					Description: "Create an inline comment linked to specific text position within entity description.",
					Tags:        []string{"comments"},
					Parameters: []Parameter{
						{Name: "entity_type", In: "path", Type: "string", Required: true, Description: "Entity type"},
						{Name: "id", In: "path", Type: "string", Required: true, Description: "Entity UUID or reference ID"},
						{Name: "inline_comment", In: "body", Type: "object", Required: true, Description: "Inline comment with position data"},
					},
					Responses: map[string]string{
						"201": "Inline comment created successfully",
						"400": "Invalid position or request data",
						"404": "Entity not found",
						"500": "Internal server error",
					},
				},
				{
					Method:      "POST",
					Path:        "/api/v1/comments/{id}/resolve",
					Summary:     "Resolve comment",
					Description: "Mark a comment as resolved, indicating the issue or discussion has been addressed.",
					Tags:        []string{"comments"},
					Parameters: []Parameter{
						{Name: "id", In: "path", Type: "string", Required: true, Description: "Comment UUID"},
					},
					Responses: map[string]string{
						"200": "Comment resolved successfully",
						"404": "Comment not found",
						"500": "Internal server error",
					},
				},
			},
		},
		{
			Name:        "Navigation & Hierarchy",
			Description: "Endpoints for retrieving hierarchical structures and entity relationships. Provides navigation capabilities for building tree views and breadcrumb navigation.",
			BasePath:    "/api/v1/hierarchy",
			Endpoints: []Endpoint{
				{
					Method:      "GET",
					Path:        "/api/v1/hierarchy",
					Summary:     "Get complete hierarchy",
					Description: "Retrieve the complete system hierarchy showing all epics with nested user stories and requirements.",
					Tags:        []string{"navigation"},
					Responses: map[string]string{
						"200": "Complete system hierarchy",
						"500": "Internal server error",
					},
				},
				{
					Method:      "GET",
					Path:        "/api/v1/hierarchy/epics/{id}",
					Summary:     "Get epic hierarchy",
					Description: "Retrieve complete hierarchy for a specific epic including all nested user stories, acceptance criteria, and requirements.",
					Tags:        []string{"navigation"},
					Parameters: []Parameter{
						{Name: "id", In: "path", Type: "string", Required: true, Description: "Epic UUID or reference ID"},
					},
					Responses: map[string]string{
						"200": "Epic hierarchy structure",
						"404": "Epic not found",
						"500": "Internal server error",
					},
				},
				{
					Method:      "GET",
					Path:        "/api/v1/hierarchy/path/{entity_type}/{id}",
					Summary:     "Get entity path",
					Description: "Retrieve the hierarchical path from root to a specific entity for breadcrumb navigation.",
					Tags:        []string{"navigation"},
					Parameters: []Parameter{
						{Name: "entity_type", In: "path", Type: "string", Required: true, Description: "Entity type"},
						{Name: "id", In: "path", Type: "string", Required: true, Description: "Entity UUID or reference ID"},
					},
					Responses: map[string]string{
						"200": "Entity hierarchical path",
						"404": "Entity not found",
						"500": "Internal server error",
					},
				},
			},
		},
		{
			Name:        "System Configuration",
			Description: "Administrative endpoints for configuring system behavior including requirement types, relationship types, and status workflows. Requires administrator role.",
			BasePath:    "/api/v1/config",
			Endpoints: []Endpoint{
				{
					Method:      "GET",
					Path:        "/api/v1/config/requirement-types",
					Summary:     "List requirement types",
					Description: "Retrieve all available requirement types for categorizing requirements (functional, non-functional, etc.).",
					Tags:        []string{"configuration"},
					Responses: map[string]string{
						"200": "List of requirement types",
						"401": "Authentication required",
						"403": "Administrator role required",
						"500": "Internal server error",
					},
				},
				{
					Method:      "POST",
					Path:        "/api/v1/config/requirement-types",
					Summary:     "Create requirement type",
					Description: "Create a new requirement type for system-wide use in requirement categorization.",
					Tags:        []string{"configuration"},
					Parameters: []Parameter{
						{Name: "requirement_type", In: "body", Type: "object", Required: true, Description: "Requirement type creation request"},
					},
					Responses: map[string]string{
						"201": "Requirement type created",
						"400": "Invalid request data",
						"401": "Authentication required",
						"403": "Administrator role required",
						"500": "Internal server error",
					},
				},
				{
					Method:      "GET",
					Path:        "/api/v1/config/status-models",
					Summary:     "List status models",
					Description: "Retrieve all status workflow models that define allowed status transitions for different entity types.",
					Tags:        []string{"configuration"},
					Responses: map[string]string{
						"200": "List of status models",
						"401": "Authentication required",
						"403": "Administrator role required",
						"500": "Internal server error",
					},
				},
			},
		},
		{
			Name:        "Deletion Management",
			Description: "Comprehensive deletion endpoints with dependency validation and cascade options. Provides safe deletion with impact analysis to prevent data integrity issues.",
			BasePath:    "/api/deletion",
			Endpoints: []Endpoint{
				{
					Method:      "GET",
					Path:        "/api/v1/{entity_type}/{id}/validate-deletion",
					Summary:     "Validate entity deletion",
					Description: "Check if an entity can be safely deleted and return dependency information showing what would be affected.",
					Tags:        []string{"deletion"},
					Parameters: []Parameter{
						{Name: "entity_type", In: "path", Type: "string", Required: true, Description: "Entity type to validate"},
						{Name: "id", In: "path", Type: "string", Required: true, Description: "Entity UUID or reference ID"},
					},
					Responses: map[string]string{
						"200": "Deletion validation result with dependencies",
						"404": "Entity not found",
						"500": "Internal server error",
					},
				},
				{
					Method:      "DELETE",
					Path:        "/api/v1/{entity_type}/{id}/delete",
					Summary:     "Delete entity with options",
					Description: "Delete an entity with comprehensive validation and optional cascade deletion of dependent entities.",
					Tags:        []string{"deletion"},
					Parameters: []Parameter{
						{Name: "entity_type", In: "path", Type: "string", Required: true, Description: "Entity type to delete"},
						{Name: "id", In: "path", Type: "string", Required: true, Description: "Entity UUID or reference ID"},
						{Name: "options", In: "body", Type: "object", Required: false, Description: "Deletion options including cascade settings"},
					},
					Responses: map[string]string{
						"200": "Entity deleted successfully with impact report",
						"400": "Invalid deletion options",
						"404": "Entity not found",
						"409": "Cannot delete due to dependencies",
						"500": "Internal server error",
					},
				},
			},
		},
		{
			Name:        "Health & Monitoring",
			Description: "System health and monitoring endpoints for service status, database connectivity, and operational metrics. Used for health checks and monitoring integrations.",
			BasePath:    "/health",
			Endpoints: []Endpoint{
				{
					Method:      "GET",
					Path:        "/health",
					Summary:     "Basic health check",
					Description: "Simple health check endpoint returning service status.",
					Tags:        []string{"health"},
					Responses: map[string]string{
						"200": "Service is healthy",
						"503": "Service unavailable",
					},
				},
				{
					Method:      "GET",
					Path:        "/health/deep",
					Summary:     "Deep health check",
					Description: "Comprehensive health check including database connectivity and dependency status.",
					Tags:        []string{"health"},
					Responses: map[string]string{
						"200": "All systems healthy",
						"503": "One or more systems unhealthy",
					},
				},
				{
					Method:      "GET",
					Path:        "/ready",
					Summary:     "Readiness check",
					Description: "Kubernetes readiness probe endpoint indicating if service is ready to accept traffic.",
					Tags:        []string{"health"},
					Responses: map[string]string{
						"200": "Service ready",
						"503": "Service not ready",
					},
				},
				{
					Method:      "GET",
					Path:        "/live",
					Summary:     "Liveness check",
					Description: "Kubernetes liveness probe endpoint indicating if service is alive and should not be restarted.",
					Tags:        []string{"health"},
					Responses: map[string]string{
						"200": "Service alive",
					},
				},
			},
		},
	}
}

// GetEndpointsByTag returns endpoints organized by Swagger tags
func GetEndpointsByTag() map[string][]Endpoint {
	organization := GetAPIOrganization()
	tagMap := make(map[string][]Endpoint)

	for _, group := range organization {
		for _, endpoint := range group.Endpoints {
			for _, tag := range endpoint.Tags {
				tagMap[tag] = append(tagMap[tag], endpoint)
			}
		}
	}

	return tagMap
}

// GetAPIStatistics returns statistics about the API organization
func GetAPIStatistics() map[string]interface{} {
	organization := GetAPIOrganization()

	totalEndpoints := 0
	methodCounts := make(map[string]int)
	tagCounts := make(map[string]int)

	for _, group := range organization {
		totalEndpoints += len(group.Endpoints)

		for _, endpoint := range group.Endpoints {
			methodCounts[endpoint.Method]++

			for _, tag := range endpoint.Tags {
				tagCounts[tag]++
			}
		}
	}

	return map[string]interface{}{
		"total_endpoint_groups": len(organization),
		"total_endpoints":       totalEndpoints,
		"methods":               methodCounts,
		"tags":                  tagCounts,
		"groups": func() []map[string]interface{} {
			var groups []map[string]interface{}
			for _, group := range organization {
				groups = append(groups, map[string]interface{}{
					"name":           group.Name,
					"endpoint_count": len(group.Endpoints),
					"base_path":      group.BasePath,
				})
			}
			return groups
		}(),
	}
}
