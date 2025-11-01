# Implementation Plan: MCP Handlers Refactoring

## Overview
This implementation plan converts the monolithic `ToolsHandler` structure into a modular, domain-specific architecture following the Facade Pattern. The refactoring will improve maintainability, testability, and adherence to SOLID principles.

## Task List

- [ ] 1. Create MCP tools package structure and interfaces
  - Create `internal/mcp/tools/` directory structure
  - Define `ToolHandler` interface for domain-specific handlers
  - Create common types and utilities package
  - _Requirements: FR1, FR6_

- [x] 1.1 Create base interface and types
  - Create `internal/mcp/tools/interface.go` with `ToolHandler` interface
  - Create `internal/mcp/types/responses.go` with common response types
  - Create `internal/mcp/tools/common.go` with shared utilities
  - _Requirements: FR1, FR6_

- [x] 1.2 Move tool schemas to MCP package
  - Create `internal/mcp/schemas/tools.go`
  - Move `GetSupportedTools()` function from `mcp_tool_schemas.go`
  - Update imports and references
  - _Requirements: FR5_

- [x] 2. Create Epic domain handler
  - Create `internal/mcp/tools/epic.go` with `EpicHandler` struct
  - Implement `create_epic` and `update_epic` tool methods
  - Move logic from `ToolsHandler.handleCreateEpic` and `ToolsHandler.handleUpdateEpic`
  - Add comprehensive unit tests in `internal/mcp/tools/epic_test.go`
  - _Requirements: FR1, FR2, FR5_

- [x] 2.1 Implement Epic handler methods
  - Implement `NewEpicHandler()` constructor with `service.EpicService` dependency
  - Implement `GetSupportedTools()` method returning epic tool names
  - Implement `HandleTool()` method with routing to specific epic operations
  - _Requirements: FR1, FR2_

- [x] 2.2 Move Epic tool logic
  - Move `handleCreateEpic` logic to `EpicHandler.Create()` method
  - Move `handleUpdateEpic` logic to `EpicHandler.Update()` method
  - Ensure reference ID resolution and error handling are preserved
  - _Requirements: FR5_

- [x] 2.3 Create Epic handler unit tests
  - Test epic creation with valid and invalid parameters
  - Test epic updates with UUID and reference ID resolution
  - Test error handling and validation scenarios
  - _Requirements: NFR2_

- [x] 3. Create User Story domain handler
  - Create `internal/mcp/tools/user_story.go` with `UserStoryHandler` struct
  - Implement `create_user_story` and `update_user_story` tool methods
  - Move logic from `ToolsHandler.handleCreateUserStory` and `ToolsHandler.handleUpdateUserStory`
  - Add comprehensive unit tests in `internal/mcp/tools/user_story_test.go`
  - _Requirements: FR1, FR2, FR5_

- [x] 3.1 Implement User Story handler methods
  - Implement `NewUserStoryHandler()` constructor with `service.UserStoryService` and `service.EpicService` dependencies
  - Implement `GetSupportedTools()` method returning user story tool names
  - Implement `HandleTool()` method with routing to specific user story operations
  - _Requirements: FR1, FR2_

- [x] 3.2 Move User Story tool logic
  - Move `handleCreateUserStory` logic to `UserStoryHandler.Create()` method
  - Move `handleUpdateUserStory` logic to `UserStoryHandler.Update()` method
  - Ensure epic reference ID resolution is preserved
  - _Requirements: FR5_

- [x] 3.3 Create User Story handler unit tests
  - Test user story creation with epic ID resolution
  - Test user story updates with validation
  - Test error handling for invalid epic references
  - _Requirements: NFR2_

- [x] 4. Create Requirement domain handler
  - Create `internal/mcp/tools/requirement.go` with `RequirementHandler` struct
  - Implement `create_requirement`, `update_requirement`, and `create_relationship` tool methods
  - Move logic from `ToolsHandler.handleCreateRequirement`, `ToolsHandler.handleUpdateRequirement`, and `ToolsHandler.handleCreateRelationship`
  - Add comprehensive unit tests in `internal/mcp/tools/requirement_test.go`
  - _Requirements: FR1, FR2, FR5_

- [x] 4.1 Implement Requirement handler methods
  - Implement `NewRequirementHandler()` constructor with `service.RequirementService` and `service.UserStoryService` dependencies
  - Implement `GetSupportedTools()` method returning requirement tool names
  - Implement `HandleTool()` method with routing to requirement operations
  - _Requirements: FR1, FR2_

- [x] 4.2 Move Requirement tool logic
  - Move `handleCreateRequirement` logic to `RequirementHandler.Create()` method
  - Move `handleUpdateRequirement` logic to `RequirementHandler.Update()` method
  - Move `handleCreateRelationship` logic to `RequirementHandler.CreateRelationship()` method
  - _Requirements: FR5_

- [x] 4.3 Create Requirement handler unit tests
  - Test requirement creation with user story and acceptance criteria linking
  - Test requirement updates and relationship creation
  - Test reference ID resolution for source and target requirements
  - _Requirements: NFR2_

- [x] 5. Create Search domain handler
  - Create `internal/mcp/tools/search.go` with `SearchHandler` struct
  - Implement `search_global` and `search_requirements` tool methods
  - Move logic from `ToolsHandler.handleSearchGlobal` and `ToolsHandler.handleSearchRequirements`
  - Add comprehensive unit tests in `internal/mcp/tools/search_test.go`
  - _Requirements: FR1, FR2, FR5_

- [x] 5.1 Implement Search handler methods
  - Implement `NewSearchHandler()` constructor with `service.SearchServiceInterface` dependency
  - Implement `GetSupportedTools()` method returning search tool names
  - Implement `HandleTool()` method with routing to search operations
  - _Requirements: FR1, FR2_

- [x] 5.2 Move Search tool logic
  - Move `handleSearchGlobal` logic to `SearchHandler.Global()` method
  - Move `handleSearchRequirements` logic to `SearchHandler.Requirements()` method
  - Preserve search options and result formatting
  - _Requirements: FR5_

- [x] 5.3 Create Search handler unit tests
  - Test global search with entity type filtering
  - Test requirements-specific search functionality
  - Test pagination and result formatting
  - _Requirements: NFR2_

- [x] 6. Create Steering Document domain handler
  - Create `internal/mcp/tools/steering_document.go` with `SteeringDocumentHandler` struct
  - Implement all steering document tool methods (`list_steering_documents`, `create_steering_document`, `get_steering_document`, `update_steering_document`, `link_steering_to_epic`, `unlink_steering_from_epic`, `get_epic_steering_documents`)
  - Move logic from corresponding `ToolsHandler.handle*` methods
  - Add comprehensive unit tests in `internal/mcp/tools/steering_document_test.go`
  - _Requirements: FR1, FR2, FR5_

- [x] 6.1 Implement Steering Document handler methods
  - Implement `NewSteeringDocumentHandler()` constructor with `service.SteeringDocumentService` and `service.EpicService` dependencies
  - Implement `GetSupportedTools()` method returning steering document tool names
  - Implement `HandleTool()` method with routing to steering document operations
  - _Requirements: FR1, FR2_

- [x] 6.2 Move Steering Document tool logic
  - Move all `handleListSteeringDocuments`, `handleCreateSteeringDocument`, etc. logic to respective methods
  - Ensure epic linking functionality and reference ID resolution are preserved
  - Maintain user permission checks and filtering
  - _Requirements: FR5_

- [x] 6.3 Create Steering Document handler unit tests
  - Test steering document CRUD operations
  - Test epic linking and unlinking functionality
  - Test user permission validation and filtering
  - _Requirements: NFR2_

- [x] 7. Create Prompt domain handler
  - Create `internal/mcp/tools/prompt.go` with `PromptHandler` struct
  - Implement all prompt tool methods (`create_prompt`, `update_prompt`, `delete_prompt`, `activate_prompt`, `list_prompts`, `get_active_prompt`)
  - Move logic from corresponding `ToolsHandler.handle*` methods
  - Add comprehensive unit tests in `internal/mcp/tools/prompt_test.go`
  - _Requirements: FR1, FR2, FR5_

- [x] 7.1 Implement Prompt handler methods
  - Implement `NewPromptHandler()` constructor with `service.PromptService` dependency
  - Implement `GetSupportedTools()` method returning prompt tool names
  - Implement `HandleTool()` method with routing to prompt operations
  - _Requirements: FR1, FR2_

- [x] 7.2 Move Prompt tool logic
  - Move all `handleCreatePrompt`, `handleUpdatePrompt`, etc. logic to respective methods
  - Preserve Administrator role validation and permission checks
  - Maintain reference ID resolution and error handling
  - _Requirements: FR5_

- [x] 7.3 Create Prompt handler unit tests
  - Test prompt CRUD operations with Administrator role validation
  - Test prompt activation and listing functionality
  - Test error handling for insufficient permissions
  - _Requirements: NFR2_

- [x] 8. Create Tools Handler Facade
  - Create `internal/mcp/tools/handler.go` with main `Handler` struct
  - Implement facade pattern with tool routing map
  - Create `NewHandler()` constructor that initializes all domain handlers
  - Implement `HandleToolsCall()` method that routes to appropriate domain handlers
  - _Requirements: FR3, FR4_

- [x] 8.1 Implement facade constructor and routing
  - Create `Handler` struct with all domain handler dependencies
  - Build tool routing map for O(1) lookup performance
  - Implement centralized error handling and logging
  - _Requirements: FR3, FR4_

- [x] 8.2 Implement tool call routing
  - Implement `HandleToolsCall()` method with parameter extraction
  - Route tool calls to appropriate domain handlers using the routing map
  - Maintain backward compatibility with existing API
  - _Requirements: FR4, NFR1_

- [x] 8.3 Create facade integration tests
  - Test tool routing to correct domain handlers
  - Test error handling for unknown tools
  - Test parameter validation and response formatting
  - _Requirements: NFR2_

- [x] 9. Update main MCP handler integration
  - Update `internal/handlers/mcp_handler.go` to use new tools package
  - Replace `ToolsHandler` with `tools.Handler` in `MCPHandler` struct
  - Update constructor to use new `tools.NewHandler()` function
  - Ensure all existing functionality works with new architecture
  - _Requirements: FR3, NFR1_

- [x] 9.1 Update MCP handler constructor
  - Modify `NewMCPHandler()` to create `tools.Handler` instead of `ToolsHandler`
  - Update dependency injection to pass services to new tools handler
  - Ensure proper initialization of all domain handlers
  - _Requirements: FR3_

- [x] 9.2 Update handler registration
  - Update `tools/call` handler registration to use new facade
  - Ensure `tools/list` handler uses new schema location
  - Maintain existing JSON-RPC method signatures
  - _Requirements: NFR1_

- [x] 10. Integration testing and validation
  - Run comprehensive integration tests to ensure API compatibility
  - Validate that all existing MCP tool functionality works correctly
  - Test error handling and response formatting
  - Verify performance is maintained or improved
  - _Requirements: NFR1, NFR2_

- [x] 10.1 Run existing MCP integration tests
  - Execute all existing MCP-related integration tests
  - Verify `mcp_tools_handler_test.go` passes with new architecture
  - Test `mcp_capabilities_integration_test.go` functionality
  - _Requirements: NFR2_

- [x] 10.2 Validate API compatibility
  - Test all 22 MCP tools through JSON-RPC interface
  - Verify request/response formats remain unchanged
  - Test error scenarios and edge cases
  - _Requirements: NFR1_

- [x] 11. Cleanup and file removal
  - Remove old `internal/handlers/mcp_tools_handler.go` file (1,500+ lines)
  - Remove old `internal/handlers/mcp_tool_schemas.go` file (500+ lines)
  - Update imports throughout codebase to use new package locations
  - Clean up any unused dependencies or imports
  - _Requirements: FR5_

- [x] 11.1 Remove legacy files
  - Delete `internal/handlers/mcp_tools_handler.go`
  - Delete `internal/handlers/mcp_tool_schemas.go`
  - Update any remaining references to old handler locations
  - _Requirements: FR5_

- [x] 11.2 Update imports and references
  - Update all imports to use new `internal/mcp/tools` package
  - Update any documentation or comments referencing old structure
  - Ensure no broken imports or references remain
  - _Requirements: FR5_

## Implementation Notes

### Testing Strategy
- Unit tests focus on individual domain handlers with mocked dependencies
- Integration tests validate the facade routing and API compatibility
- Optional test tasks (marked with *) can be skipped for faster MVP delivery
- Existing test coverage must be maintained or improved

### Backward Compatibility
- All JSON-RPC method signatures remain unchanged
- Tool request/response formats are preserved
- Error handling patterns are maintained
- Performance characteristics are preserved or improved

### Architecture Benefits
- **Single Responsibility**: Each handler focuses on one domain
- **Reduced Coupling**: Minimal dependencies per handler  
- **Improved Testability**: Isolated, focused unit tests
- **Better Maintainability**: Smaller, more focused code files
- **Future Extensibility**: Easy to add new domains and tools

### File Organization
```
internal/mcp/
├── tools/
│   ├── handler.go (Facade/Router)
│   ├── interface.go (ToolHandler interface)
│   ├── common.go (shared utilities)
│   ├── epic.go & epic_test.go
│   ├── user_story.go & user_story_test.go
│   ├── requirement.go & requirement_test.go
│   ├── search.go & search_test.go
│   ├── steering_document.go & steering_document_test.go
│   └── prompt.go & prompt_test.go
├── schemas/
│   └── tools.go (tool definitions)
└── types/
    └── responses.go (common response types)
```

This refactoring will result in a net reduction of ~2,000 lines of code while significantly improving maintainability and adherence to SOLID principles.