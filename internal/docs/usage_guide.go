package docs

// UsageGuide provides comprehensive documentation for common API usage patterns
// This file contains practical examples and workflows for API consumers

// WorkflowGuide represents a complete workflow with multiple steps
type WorkflowGuide struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Steps       []WorkflowStep    `json:"steps"`
	Examples    []WorkflowExample `json:"examples"`
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	Step        int                    `json:"step"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Method      string                 `json:"method"`
	Endpoint    string                 `json:"endpoint"`
	Headers     map[string]string      `json:"headers,omitempty"`
	Body        map[string]interface{} `json:"body,omitempty"`
	Response    map[string]interface{} `json:"response,omitempty"`
	Notes       []string               `json:"notes,omitempty"`
}

// WorkflowExample provides a complete example of a workflow execution
type WorkflowExample struct {
	Scenario    string                `json:"scenario"`
	Description string                `json:"description"`
	Steps       []WorkflowExampleStep `json:"steps"`
}

// WorkflowExampleStep provides actual request/response examples
type WorkflowExampleStep struct {
	Step     int                    `json:"step"`
	Request  map[string]interface{} `json:"request"`
	Response map[string]interface{} `json:"response"`
}

// GetUsageGuides returns comprehensive usage guides for common workflows
func GetUsageGuides() []WorkflowGuide {
	return []WorkflowGuide{
		{
			Name:        "Complete Epic Creation Workflow",
			Description: "End-to-end workflow for creating a complete epic with user stories, acceptance criteria, and requirements",
			Steps: []WorkflowStep{
				{
					Step:        1,
					Title:       "Create Epic",
					Description: "Create the main epic container that will hold all user stories",
					Method:      "POST",
					Endpoint:    "/api/v1/epics",
					Headers: map[string]string{
						"Authorization": "Bearer {jwt_token}",
						"Content-Type":  "application/json",
					},
					Body: map[string]interface{}{
						"title":       "User Authentication System",
						"description": "Implement comprehensive user authentication and authorization system with JWT tokens, role-based access control, and secure session management",
						"priority":    1,
						"creator_id":  "{user_uuid}",
					},
					Response: map[string]interface{}{
						"id":           "{epic_uuid}",
						"reference_id": "EP-001",
						"title":        "User Authentication System",
						"status":       "backlog",
						"created_at":   "2023-01-15T10:30:00Z",
					},
					Notes: []string{
						"Epic will be automatically assigned a reference ID (e.g., EP-001)",
						"Default status is 'backlog' unless specified",
						"Creator ID should be the authenticated user's UUID",
					},
				},
				{
					Step:        2,
					Title:       "Create User Stories",
					Description: "Add user stories to the epic to define feature requirements from user perspective",
					Method:      "POST",
					Endpoint:    "/api/v1/epics/{epic_id}/user-stories",
					Headers: map[string]string{
						"Authorization": "Bearer {jwt_token}",
						"Content-Type":  "application/json",
					},
					Body: map[string]interface{}{
						"title":       "User Login with JWT Token",
						"description": "As a user, I want to log in with my credentials and receive a JWT token, so that I can access protected resources securely",
						"priority":    1,
						"creator_id":  "{user_uuid}",
					},
					Response: map[string]interface{}{
						"id":           "{user_story_uuid}",
						"reference_id": "US-001",
						"title":        "User Login with JWT Token",
						"epic_id":      "{epic_uuid}",
						"status":       "backlog",
					},
					Notes: []string{
						"User story is automatically linked to the epic",
						"Reference ID is auto-generated (US-001, US-002, etc.)",
						"Use user story format: 'As a [role], I want [feature], so that [benefit]'",
					},
				},
				{
					Step:        3,
					Title:       "Add Acceptance Criteria",
					Description: "Define testable conditions for the user story using EARS format",
					Method:      "POST",
					Endpoint:    "/api/v1/user-stories/{user_story_id}/acceptance-criteria",
					Headers: map[string]string{
						"Authorization": "Bearer {jwt_token}",
						"Content-Type":  "application/json",
					},
					Body: map[string]interface{}{
						"description": "WHEN user enters valid credentials THEN system SHALL authenticate user and return JWT token with 1-hour expiration",
						"author_id":   "{user_uuid}",
					},
					Response: map[string]interface{}{
						"id":            "{acceptance_criteria_uuid}",
						"reference_id":  "AC-001",
						"user_story_id": "{user_story_uuid}",
						"description":   "WHEN user enters valid credentials THEN system SHALL authenticate user and return JWT token with 1-hour expiration",
					},
					Notes: []string{
						"Use EARS format: WHEN [condition] THEN system SHALL [response]",
						"Each acceptance criteria gets a unique reference ID (AC-001, AC-002, etc.)",
						"Multiple acceptance criteria can be added to one user story",
					},
				},
				{
					Step:        4,
					Title:       "Create Detailed Requirements",
					Description: "Add technical requirements that specify implementation details",
					Method:      "POST",
					Endpoint:    "/api/v1/user-stories/{user_story_id}/requirements",
					Headers: map[string]string{
						"Authorization": "Bearer {jwt_token}",
						"Content-Type":  "application/json",
					},
					Body: map[string]interface{}{
						"title":                  "JWT Token Generation",
						"description":            "System must generate JWT tokens using HS256 algorithm with configurable secret key and expiration time",
						"priority":               1,
						"requirement_type_id":    "{functional_requirement_type_id}",
						"acceptance_criteria_id": "{acceptance_criteria_uuid}",
						"creator_id":             "{user_uuid}",
					},
					Response: map[string]interface{}{
						"id":                     "{requirement_uuid}",
						"reference_id":           "REQ-001",
						"title":                  "JWT Token Generation",
						"user_story_id":          "{user_story_uuid}",
						"acceptance_criteria_id": "{acceptance_criteria_uuid}",
						"status":                 "draft",
					},
					Notes: []string{
						"Requirements can be linked to specific acceptance criteria",
						"Requirement type ID determines categorization (functional, non-functional, etc.)",
						"Reference ID is auto-generated (REQ-001, REQ-002, etc.)",
					},
				},
				{
					Step:        5,
					Title:       "Create Requirement Relationships",
					Description: "Link requirements with dependencies and relationships",
					Method:      "POST",
					Endpoint:    "/api/v1/requirements/relationships",
					Headers: map[string]string{
						"Authorization": "Bearer {jwt_token}",
						"Content-Type":  "application/json",
					},
					Body: map[string]interface{}{
						"source_requirement_id": "{jwt_requirement_uuid}",
						"target_requirement_id": "{auth_requirement_uuid}",
						"relationship_type_id":  "{depends_on_type_id}",
						"description":           "JWT generation depends on user authentication being implemented first",
					},
					Response: map[string]interface{}{
						"id":                    "{relationship_uuid}",
						"source_requirement_id": "{jwt_requirement_uuid}",
						"target_requirement_id": "{auth_requirement_uuid}",
						"relationship_type":     "depends_on",
					},
					Notes: []string{
						"Common relationship types: depends_on, blocks, relates_to, conflicts_with, derives_from",
						"System prevents circular dependencies",
						"Relationships help with impact analysis and planning",
					},
				},
			},
			Examples: []WorkflowExample{
				{
					Scenario:    "Authentication System Epic",
					Description: "Complete example of creating an authentication system epic with all components",
					Steps: []WorkflowExampleStep{
						{
							Step: 1,
							Request: map[string]interface{}{
								"method":   "POST",
								"endpoint": "/api/v1/epics",
								"body": map[string]interface{}{
									"title":       "User Authentication System",
									"description": "Implement comprehensive user authentication and authorization system",
									"priority":    1,
									"creator_id":  "123e4567-e89b-12d3-a456-426614174001",
								},
							},
							Response: map[string]interface{}{
								"id":           "123e4567-e89b-12d3-a456-426614174000",
								"reference_id": "EP-001",
								"title":        "User Authentication System",
								"status":       "backlog",
								"created_at":   "2023-01-15T10:30:00Z",
							},
						},
					},
				},
			},
		},
		{
			Name:        "Search and Discovery Workflow",
			Description: "How to effectively search and discover content across the system",
			Steps: []WorkflowStep{
				{
					Step:        1,
					Title:       "Basic Search",
					Description: "Perform a basic full-text search across all entities",
					Method:      "GET",
					Endpoint:    "/api/v1/search?query=authentication&limit=20",
					Headers: map[string]string{
						"Authorization": "Bearer {jwt_token}",
					},
					Response: map[string]interface{}{
						"results": []map[string]interface{}{
							{
								"entity_type":  "epic",
								"id":           "{epic_uuid}",
								"reference_id": "EP-001",
								"title":        "User Authentication System",
								"relevance":    0.95,
							},
						},
						"total":  1,
						"limit":  20,
						"offset": 0,
					},
					Notes: []string{
						"Search query is case-insensitive",
						"Results are ranked by relevance",
						"Searches across titles, descriptions, and content",
					},
				},
				{
					Step:        2,
					Title:       "Filtered Search",
					Description: "Search with entity type and status filtering",
					Method:      "GET",
					Endpoint:    "/api/v1/search?query=login&entity_types=user_stories,requirements&status=in_progress",
					Headers: map[string]string{
						"Authorization": "Bearer {jwt_token}",
					},
					Response: map[string]interface{}{
						"results": []map[string]interface{}{
							{
								"entity_type":  "user_story",
								"id":           "{user_story_uuid}",
								"reference_id": "US-001",
								"title":        "User Login with JWT Token",
								"status":       "in_progress",
								"relevance":    0.88,
							},
						},
						"filters_applied": map[string]interface{}{
							"entity_types": []string{"user_stories", "requirements"},
							"status":       "in_progress",
						},
					},
					Notes: []string{
						"Multiple entity types can be specified",
						"Filters reduce result set before ranking",
						"Available entity types: epics, user_stories, acceptance_criteria, requirements, comments",
					},
				},
				{
					Step:        3,
					Title:       "Search Suggestions",
					Description: "Get search suggestions for autocomplete functionality",
					Method:      "GET",
					Endpoint:    "/api/v1/search/suggestions?query=auth&limit=5",
					Headers: map[string]string{
						"Authorization": "Bearer {jwt_token}",
					},
					Response: map[string]interface{}{
						"suggestions": map[string]interface{}{
							"titles": []string{
								"User Authentication System",
								"Authentication Service",
								"Authorization Framework",
							},
							"reference_ids": []string{
								"EP-001",
								"US-001",
								"REQ-001",
							},
							"statuses": []string{
								"in_progress",
								"ready",
							},
						},
					},
					Notes: []string{
						"Minimum 2 characters required for suggestions",
						"Suggestions grouped by category",
						"Useful for building autocomplete interfaces",
					},
				},
			},
		},
		{
			Name:        "Comment and Collaboration Workflow",
			Description: "How to use the comment system for collaboration and feedback",
			Steps: []WorkflowStep{
				{
					Step:        1,
					Title:       "Create General Comment",
					Description: "Add a general comment to any entity for discussion",
					Method:      "POST",
					Endpoint:    "/api/v1/epics/{epic_id}/comments",
					Headers: map[string]string{
						"Authorization": "Bearer {jwt_token}",
						"Content-Type":  "application/json",
					},
					Body: map[string]interface{}{
						"content":   "This epic looks comprehensive. Should we consider adding two-factor authentication as well?",
						"author_id": "{user_uuid}",
					},
					Response: map[string]interface{}{
						"id":          "{comment_uuid}",
						"entity_type": "epic",
						"entity_id":   "{epic_uuid}",
						"content":     "This epic looks comprehensive. Should we consider adding two-factor authentication as well?",
						"status":      "open",
						"created_at":  "2023-01-18T14:30:00Z",
					},
					Notes: []string{
						"Comments can be added to any entity type",
						"Default status is 'open'",
						"Comments support threading with replies",
					},
				},
				{
					Step:        2,
					Title:       "Create Inline Comment",
					Description: "Add an inline comment linked to specific text position",
					Method:      "POST",
					Endpoint:    "/api/v1/user-stories/{user_story_id}/comments/inline",
					Headers: map[string]string{
						"Authorization": "Bearer {jwt_token}",
						"Content-Type":  "application/json",
					},
					Body: map[string]interface{}{
						"content":         "Consider adding password strength requirements here",
						"author_id":       "{user_uuid}",
						"is_inline":       true,
						"inline_position": 45,
						"linked_text":     "credentials",
					},
					Response: map[string]interface{}{
						"id":              "{inline_comment_uuid}",
						"entity_type":     "user_story",
						"entity_id":       "{user_story_uuid}",
						"content":         "Consider adding password strength requirements here",
						"is_inline":       true,
						"inline_position": 45,
						"linked_text":     "credentials",
						"status":          "open",
					},
					Notes: []string{
						"Inline position refers to character position in description",
						"Linked text shows what text the comment refers to",
						"Inline comments are visually highlighted in UI",
					},
				},
				{
					Step:        3,
					Title:       "Resolve Comment",
					Description: "Mark a comment as resolved when issue is addressed",
					Method:      "POST",
					Endpoint:    "/api/v1/comments/{comment_id}/resolve",
					Headers: map[string]string{
						"Authorization": "Bearer {jwt_token}",
					},
					Response: map[string]interface{}{
						"id":                  "{comment_uuid}",
						"status":              "resolved",
						"resolved_at":         "2023-01-21T10:00:00Z",
						"resolved_by_user_id": "{user_uuid}",
					},
					Notes: []string{
						"Only comment author or entity owner can resolve",
						"Resolved comments can be unresolve if needed",
						"Resolution tracking helps with issue management",
					},
				},
			},
		},
		{
			Name:        "Navigation and Hierarchy Workflow",
			Description: "How to navigate the hierarchical structure and retrieve entity relationships",
			Steps: []WorkflowStep{
				{
					Step:        1,
					Title:       "Get Complete Hierarchy",
					Description: "Retrieve the complete system hierarchy for overview",
					Method:      "GET",
					Endpoint:    "/api/v1/hierarchy",
					Headers: map[string]string{
						"Authorization": "Bearer {jwt_token}",
					},
					Response: map[string]interface{}{
						"epics": []map[string]interface{}{
							{
								"id":           "{epic_uuid}",
								"reference_id": "EP-001",
								"title":        "User Authentication System",
								"user_stories": []map[string]interface{}{
									{
										"id":           "{user_story_uuid}",
										"reference_id": "US-001",
										"title":        "User Login with JWT Token",
									},
								},
							},
						},
					},
					Notes: []string{
						"Returns nested structure of all entities",
						"Useful for building tree views",
						"May be large for systems with many entities",
					},
				},
				{
					Step:        2,
					Title:       "Get Epic Hierarchy",
					Description: "Retrieve complete hierarchy for a specific epic",
					Method:      "GET",
					Endpoint:    "/api/v1/hierarchy/epics/{epic_id}",
					Headers: map[string]string{
						"Authorization": "Bearer {jwt_token}",
					},
					Response: map[string]interface{}{
						"epic": map[string]interface{}{
							"id":           "{epic_uuid}",
							"reference_id": "EP-001",
							"title":        "User Authentication System",
							"user_stories": []map[string]interface{}{
								{
									"id":                  "{user_story_uuid}",
									"reference_id":        "US-001",
									"title":               "User Login with JWT Token",
									"acceptance_criteria": []map[string]interface{}{},
									"requirements":        []map[string]interface{}{},
								},
							},
						},
					},
					Notes: []string{
						"Returns complete nested structure for one epic",
						"Includes all user stories, acceptance criteria, and requirements",
						"More efficient than getting complete hierarchy",
					},
				},
				{
					Step:        3,
					Title:       "Get Entity Path",
					Description: "Get hierarchical path to entity for breadcrumb navigation",
					Method:      "GET",
					Endpoint:    "/api/v1/hierarchy/path/requirements/{requirement_id}",
					Headers: map[string]string{
						"Authorization": "Bearer {jwt_token}",
					},
					Response: map[string]interface{}{
						"path": []map[string]interface{}{
							{
								"entity_type":  "epic",
								"id":           "{epic_uuid}",
								"reference_id": "EP-001",
								"title":        "User Authentication System",
							},
							{
								"entity_type":  "user_story",
								"id":           "{user_story_uuid}",
								"reference_id": "US-001",
								"title":        "User Login with JWT Token",
							},
							{
								"entity_type":  "requirement",
								"id":           "{requirement_uuid}",
								"reference_id": "REQ-001",
								"title":        "JWT Token Generation",
							},
						},
					},
					Notes: []string{
						"Returns path from root to specific entity",
						"Useful for breadcrumb navigation",
						"Shows hierarchical relationship chain",
					},
				},
			},
		},
	}
}

// GetQuickStartGuide returns a quick start guide for new API users
func GetQuickStartGuide() map[string]interface{} {
	return map[string]interface{}{
		"title":       "Quick Start Guide",
		"description": "Get started with the Product Requirements Management API in 5 minutes",
		"steps": []map[string]interface{}{
			{
				"step":        1,
				"title":       "Authentication",
				"description": "Obtain a JWT token for API access",
				"details": []string{
					"Contact your administrator for API credentials",
					"Use the authentication endpoint to get a JWT token",
					"Include the token in all requests: 'Authorization: Bearer {token}'",
					"Tokens expire after 1 hour and must be refreshed",
				},
			},
			{
				"step":        2,
				"title":       "Create Your First Epic",
				"description": "Create a high-level feature container",
				"example": map[string]interface{}{
					"method":   "POST",
					"endpoint": "/api/v1/epics",
					"body": map[string]interface{}{
						"title":       "My First Epic",
						"description": "A sample epic to get started",
						"priority":    1,
						"creator_id":  "{your_user_id}",
					},
				},
			},
			{
				"step":        3,
				"title":       "Add a User Story",
				"description": "Define a feature requirement within your epic",
				"example": map[string]interface{}{
					"method":   "POST",
					"endpoint": "/api/v1/epics/{epic_id}/user-stories",
					"body": map[string]interface{}{
						"title":       "As a user, I want to see a welcome message",
						"description": "Display a personalized welcome message on login",
						"priority":    1,
						"creator_id":  "{your_user_id}",
					},
				},
			},
			{
				"step":        4,
				"title":       "Search Your Content",
				"description": "Find your created content using search",
				"example": map[string]interface{}{
					"method":   "GET",
					"endpoint": "/api/v1/search?query=welcome&entity_types=user_stories",
				},
			},
			{
				"step":        5,
				"title":       "Explore the Hierarchy",
				"description": "View your epic with all nested content",
				"example": map[string]interface{}{
					"method":   "GET",
					"endpoint": "/api/v1/hierarchy/epics/{epic_id}",
				},
			},
		},
		"next_steps": []string{
			"Add acceptance criteria to define testable conditions",
			"Create detailed requirements with relationships",
			"Use comments for collaboration and feedback",
			"Explore advanced search and filtering options",
			"Set up status workflows for your team",
		},
		"resources": map[string]string{
			"api_documentation":  "/swagger/index.html",
			"postman_collection": "/api/v1/docs/postman",
			"example_workflows":  "/api/v1/docs/workflows",
		},
	}
}
