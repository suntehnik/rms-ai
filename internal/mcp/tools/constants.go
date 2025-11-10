package tools

// Tool name constants for MCP server
const (
	// Epic tools
	ToolCreateEpic = "create_epic"
	ToolUpdateEpic = "update_epic"
	ToolListEpics  = "list_epics"

	// User Story tools
	ToolCreateUserStory          = "create_user_story"
	ToolUpdateUserStory          = "update_user_story"
	ToolGetUserStoryRequirements = "get_user_story_requirements"

	// Requirement tools
	ToolCreateRequirement  = "create_requirement"
	ToolUpdateRequirement  = "update_requirement"
	ToolCreateRelationship = "create_relationship"

	// Acceptance Criteria tools
	ToolCreateAcceptanceCriteria = "create_acceptance_criteria"

	// Search tools
	ToolSearchGlobal       = "search_global"
	ToolSearchRequirements = "search_requirements"

	// Steering Document tools
	ToolListSteeringDocuments    = "list_steering_documents"
	ToolCreateSteeringDocument   = "create_steering_document"
	ToolGetSteeringDocument      = "get_steering_document"
	ToolUpdateSteeringDocument   = "update_steering_document"
	ToolLinkSteeringToEpic       = "link_steering_to_epic"
	ToolUnlinkSteeringFromEpic   = "unlink_steering_from_epic"
	ToolGetEpicSteeringDocuments = "get_epic_steering_documents"

	// Prompt tools
	ToolCreatePrompt    = "create_prompt"
	ToolUpdatePrompt    = "update_prompt"
	ToolDeletePrompt    = "delete_prompt"
	ToolActivatePrompt  = "activate_prompt"
	ToolListPrompts     = "list_prompts"
	ToolGetActivePrompt = "get_active_prompt"
)
