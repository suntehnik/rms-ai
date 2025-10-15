package handlers

// ToolDefinition represents a tool definition with its schema
type ToolDefinition struct {
	Name        string      `json:"name"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// GetSupportedTools returns all supported MCP tools with their JSON schemas
func GetSupportedTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "create_epic",
			Title:       "Create Epic",
			Description: "Create a new epic in the requirements management system. The creator is automatically set to the authenticated user.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Title of the epic (required, max 500 characters)",
						"maxLength":   500,
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Detailed description of the epic (optional, max 50000 characters)",
						"maxLength":   50000,
					},
					"priority": map[string]interface{}{
						"type":        "integer",
						"description": "Priority level (1=Critical, 2=High, 3=Medium, 4=Low)",
						"minimum":     1,
						"maximum":     4,
					},

					"assignee_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the user to assign the epic to (optional)",
						"format":      "uuid",
					},
				},
				"required": []string{"title", "priority"},
			},
		},
		{
			Name:        "update_epic",
			Title:       "Update Epic",
			Description: "Update an existing epic in the requirements management system",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"epic_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID or reference ID (EP-XXX) of the epic to update",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "New title of the epic (optional, max 500 characters)",
						"maxLength":   500,
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "New description of the epic (optional, max 50000 characters)",
						"maxLength":   50000,
					},
					"priority": map[string]interface{}{
						"type":        "integer",
						"description": "New priority level (1=Critical, 2=High, 3=Medium, 4=Low)",
						"minimum":     1,
						"maximum":     4,
					},
					"assignee_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the user to assign the epic to (empty string to unassign)",
						"format":      "uuid",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "New status of the epic (Backlog, Draft, In Progress, Done, Cancelled)",
						"enum":        []string{"Backlog", "Draft", "In Progress", "Done", "Cancelled"},
					},
				},
				"required": []string{"epic_id"},
			},
		},
		{
			Name:        "create_user_story",
			Title:       "Create User Story",
			Description: "Create a new user story within an epic. The creator is automatically set to the authenticated user.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Title of the user story (required, max 500 characters)",
						"maxLength":   500,
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Description of the user story, preferably in format 'As [role], I want [function], so that [goal]' (optional, max 50000 characters)",
						"maxLength":   50000,
					},
					"priority": map[string]interface{}{
						"type":        "integer",
						"description": "Priority level (1=Critical, 2=High, 3=Medium, 4=Low)",
						"minimum":     1,
						"maximum":     4,
					},
					"epic_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID or reference ID (EP-XXX) of the parent epic",
					},

					"assignee_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the user to assign the user story to (optional)",
						"format":      "uuid",
					},
				},
				"required": []string{"title", "epic_id", "priority"},
			},
		},
		{
			Name:        "update_user_story",
			Title:       "Update User Story",
			Description: "Update an existing user story",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"user_story_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID or reference ID (US-XXX) of the user story to update",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "New title of the user story (optional, max 500 characters)",
						"maxLength":   500,
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "New description of the user story (optional, max 50000 characters)",
						"maxLength":   50000,
					},
					"priority": map[string]interface{}{
						"type":        "integer",
						"description": "New priority level (1=Critical, 2=High, 3=Medium, 4=Low)",
						"minimum":     1,
						"maximum":     4,
					},
					"assignee_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the user to assign the user story to (empty string to unassign)",
						"format":      "uuid",
					},
				},
				"required": []string{"user_story_id"},
			},
		},
		{
			Name:        "create_requirement",
			Title:       "Create Requirement",
			Description: "Create a new requirement within a user story. The creator is automatically set to the authenticated user.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Title of the requirement (required, max 500 characters)",
						"maxLength":   500,
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Detailed description of the requirement (optional, max 50000 characters)",
						"maxLength":   50000,
					},
					"priority": map[string]interface{}{
						"type":        "integer",
						"description": "Priority level (1=Critical, 2=High, 3=Medium, 4=Low)",
						"minimum":     1,
						"maximum":     4,
					},
					"user_story_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID or reference ID (US-XXX) of the parent user story",
					},
					"acceptance_criteria_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the linked acceptance criteria (optional)",
						"format":      "uuid",
					},
					"type_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the requirement type (Functional, Non-Functional, etc.)",
						"format":      "uuid",
					},

					"assignee_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the user to assign the requirement to (optional)",
						"format":      "uuid",
					},
				},
				"required": []string{"title", "user_story_id", "type_id", "priority"},
			},
		},
		{
			Name:        "update_requirement",
			Title:       "Update Requirement",
			Description: "Update an existing requirement",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"requirement_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID or reference ID (REQ-XXX) of the requirement to update",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "New title of the requirement (optional, max 500 characters)",
						"maxLength":   500,
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "New description of the requirement (optional, max 50000 characters)",
						"maxLength":   50000,
					},
					"priority": map[string]interface{}{
						"type":        "integer",
						"description": "New priority level (1=Critical, 2=High, 3=Medium, 4=Low)",
						"minimum":     1,
						"maximum":     4,
					},
					"assignee_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the user to assign the requirement to (empty string to unassign)",
						"format":      "uuid",
					},
				},
				"required": []string{"requirement_id"},
			},
		},
		{
			Name:        "create_relationship",
			Title:       "Create Requirement Relationship",
			Description: "Create a relationship between two requirements. The creator is automatically set to the authenticated user.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"source_requirement_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID or reference ID (REQ-XXX) of the source requirement",
					},
					"target_requirement_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID or reference ID (REQ-XXX) of the target requirement",
					},
					"relationship_type_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the relationship type (depends_on, blocks, relates_to, etc.)",
						"format":      "uuid",
					},
				},
				"required": []string{"source_requirement_id", "target_requirement_id", "relationship_type_id"},
			},
		},
		{
			Name:        "search_global",
			Title:       "Global Search",
			Description: "Search across all entities in the requirements management system",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query string",
						"minLength":   1,
					},
					"entity_types": map[string]interface{}{
						"type":        "array",
						"description": "Entity types to search (optional, defaults to all)",
						"items": map[string]interface{}{
							"type": "string",
							"enum": []string{"epic", "user_story", "acceptance_criteria", "requirement"},
						},
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of results to return (default: 50, max: 100)",
						"minimum":     1,
						"maximum":     100,
						"default":     50,
					},
					"offset": map[string]interface{}{
						"type":        "integer",
						"description": "Number of results to skip for pagination (default: 0)",
						"minimum":     0,
						"default":     0,
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "search_requirements",
			Title:       "Search Requirements",
			Description: "Search specifically within requirements",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query string for requirements",
						"minLength":   1,
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "list_steering_documents",
			Title:       "List Steering Documents",
			Description: "List steering documents with optional filtering and pagination",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"creator_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by creator UUID (optional)",
						"format":      "uuid",
					},
					"search": map[string]interface{}{
						"type":        "string",
						"description": "Search query for full-text search in title and description (optional)",
					},
					"order_by": map[string]interface{}{
						"type":        "string",
						"description": "Order results by field and direction (optional, default: 'created_at DESC')",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of results to return (optional, default: 50, max: 100)",
						"minimum":     1,
						"maximum":     100,
						"default":     50,
					},
					"offset": map[string]interface{}{
						"type":        "integer",
						"description": "Number of results to skip for pagination (optional, default: 0)",
						"minimum":     0,
						"default":     0,
					},
				},
				"required": []string{},
			},
		},
		{
			Name:        "create_steering_document",
			Title:       "Create Steering Document",
			Description: "Create a new steering document in the requirements management system. The creator is automatically set to the authenticated user.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Title of the steering document (required, max 500 characters)",
						"maxLength":   500,
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Detailed description of the steering document content (optional, max 50000 characters)",
						"maxLength":   50000,
					},
				},
				"required": []string{"title"},
			},
		},
		{
			Name:        "get_steering_document",
			Title:       "Get Steering Document",
			Description: "Get a steering document by UUID or reference ID (STD-XXX)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"steering_document_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID or reference ID (STD-XXX) of the steering document to retrieve",
					},
				},
				"required": []string{"steering_document_id"},
			},
		},
		{
			Name:        "update_steering_document",
			Title:       "Update Steering Document",
			Description: "Update an existing steering document",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"steering_document_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID or reference ID (STD-XXX) of the steering document to update",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "New title of the steering document (optional, max 500 characters)",
						"maxLength":   500,
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "New description of the steering document (optional, max 50000 characters)",
						"maxLength":   50000,
					},
				},
				"required": []string{"steering_document_id"},
			},
		},
		{
			Name:        "link_steering_to_epic",
			Title:       "Link Steering Document to Epic",
			Description: "Create a link between a steering document and an epic. Both entities can be specified by UUID or reference ID.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"steering_document_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID or reference ID (STD-XXX) of the steering document",
					},
					"epic_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID or reference ID (EP-XXX) of the epic",
					},
				},
				"required": []string{"steering_document_id", "epic_id"},
			},
		},
		{
			Name:        "unlink_steering_from_epic",
			Title:       "Unlink Steering Document from Epic",
			Description: "Remove the link between a steering document and an epic. Both entities can be specified by UUID or reference ID.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"steering_document_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID or reference ID (STD-XXX) of the steering document",
					},
					"epic_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID or reference ID (EP-XXX) of the epic",
					},
				},
				"required": []string{"steering_document_id", "epic_id"},
			},
		},
		{
			Name:        "get_epic_steering_documents",
			Title:       "Get Epic Steering Documents",
			Description: "Get all steering documents linked to a specific epic. Epic can be specified by UUID or reference ID.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"epic_id": map[string]interface{}{
						"type":        "string",
						"description": "UUID or reference ID (EP-XXX) of the epic",
					},
				},
				"required": []string{"epic_id"},
			},
		},
	}
}

// GetToolByName returns a tool definition by name
func GetToolByName(name string) *ToolDefinition {
	tools := GetSupportedTools()
	for _, tool := range tools {
		if tool.Name == name {
			return &tool
		}
	}
	return nil
}

// GetToolNames returns a list of all supported tool names
func GetToolNames() []string {
	tools := GetSupportedTools()
	names := make([]string, len(tools))
	for i, tool := range tools {
		names[i] = tool.Name
	}
	return names
}
