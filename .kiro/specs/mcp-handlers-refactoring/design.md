# MCP Handlers Refactoring Design Document

## Overview

This document outlines the design for refactoring the monolithic `ToolsHandler` structure into a more modular, maintainable architecture that follows SOLID principles and improves code organization.

## Architecture

### Current State Analysis

The current implementation has several architectural issues:

1. **Monolithic Structure**: The `ToolsHandler` struct contains 1,500+ lines of code handling all MCP tool operations
2. **Excessive Dependencies**: The handler depends on 6 different services simultaneously
3. **Single Responsibility Violation**: One struct handles operations for all domain entities (Epics, User Stories, Requirements, etc.)
4. **Poor Testability**: Large methods with multiple responsibilities are difficult to test in isolation
5. **Tight Coupling**: All tool operations are tightly coupled within a single file
6. **Mixed Concerns**: MCP-specific logic is mixed with general HTTP handlers

### Target Architecture

The refactored architecture will implement a **Facade Pattern** with domain-specific handlers organized in a dedicated MCP package:

```
internal/
├── handlers/
│   ├── mcp_handler.go (Main MCP Entry Point)
│   ├── mcp_resource_handler.go
│   ├── mcp_prompts_handler.go
│   └── mcp_initialize.go
└── mcp/
    ├── tools/
    │   ├── handler.go (Facade/Router)
    │   ├── epic.go
    │   ├── user_story.go
    │   ├── requirement.go
    │   ├── search.go
    │   ├── steering_document.go
    │   ├── prompt.go
    │   ├── acceptance_criteria.go (future)
    │   └── common.go (shared utilities)
    ├── schemas/
    │   └── tools.go (tool definitions)
    └── types/
        └── responses.go (common response types)
```

### Package Organization Benefits

1. **Clear Separation**: MCP-specific logic is isolated from general HTTP handlers
2. **Domain Grouping**: All MCP tools are organized by domain within the `mcp/tools` package
3. **Shared Resources**: Common utilities and types are easily accessible within the MCP package
4. **Future Extensibility**: New MCP features (resources, prompts, etc.) can be added to the MCP package
5. **Import Clarity**: Clear distinction between `internal/handlers` (HTTP) and `internal/mcp` (MCP protocol)

## Components and Interfaces

### 1. Domain-Specific Handler Interface

**File**: `internal/mcp/tools/interface.go`
**Package**: `package tools`

```go
package tools

// ToolHandler defines the interface for domain-specific MCP tool handlers
type ToolHandler interface {
    // GetSupportedTools returns the list of tools this handler supports
    GetSupportedTools() []string
    
    // HandleTool processes a specific tool call for this domain
    HandleTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error)
}
```

### 2. MCP Tools Package Structure

#### Epic Handler
**File**: `internal/mcp/tools/epic.go`
**Package**: `package tools`
**Dependencies**: `service.EpicService`
**Supported Tools**:
- `create_epic`
- `update_epic`

```go
package tools

type EpicHandler struct {
    epicService service.EpicService
}

func NewEpicHandler(epicService service.EpicService) *EpicHandler {
    return &EpicHandler{
        epicService: epicService,
    }
}
```

#### User Story Handler
**File**: `internal/mcp/tools/user_story.go`
**Package**: `package tools`
**Dependencies**: `service.UserStoryService`, `service.EpicService` (for reference ID resolution)
**Supported Tools**:
- `create_user_story`
- `update_user_story`

#### Requirement Handler
**File**: `internal/mcp/tools/requirement.go`
**Package**: `package tools`
**Dependencies**: `service.RequirementService`, `service.UserStoryService` (for reference ID resolution)
**Supported Tools**:
- `create_requirement`
- `update_requirement`
- `create_relationship`

#### Search Handler
**File**: `internal/mcp/tools/search.go`
**Package**: `package tools`
**Dependencies**: `service.SearchServiceInterface`
**Supported Tools**:
- `search_global`
- `search_requirements`

#### Steering Document Handler
**File**: `internal/mcp/tools/steering_document.go`
**Package**: `package tools`
**Dependencies**: `service.SteeringDocumentService`, `service.EpicService` (for reference ID resolution)
**Supported Tools**:
- `list_steering_documents`
- `create_steering_document`
- `get_steering_document`
- `update_steering_document`
- `link_steering_to_epic`
- `unlink_steering_from_epic`
- `get_epic_steering_documents`

#### Prompt Handler
**File**: `internal/mcp/tools/prompt.go`
**Package**: `package tools`
**Dependencies**: `service.PromptService`
**Supported Tools**:
- `create_prompt`
- `update_prompt`
- `delete_prompt`
- `activate_prompt`
- `list_prompts`
- `get_active_prompt`

### 3. Tools Handler Facade

**File**: `internal/mcp/tools/handler.go`
**Package**: `package tools`

The main `Handler` will be a lightweight facade that:

1. **Routes tool calls** to appropriate domain handlers
2. **Maintains backward compatibility** with existing API
3. **Provides centralized error handling** and logging
4. **Manages handler lifecycle** and dependencies

```go
package tools

type Handler struct {
    // Domain-specific handlers
    epicHandler             *EpicHandler
    userStoryHandler        *UserStoryHandler
    requirementHandler      *RequirementHandler
    searchHandler           *SearchHandler
    steeringDocumentHandler *SteeringDocumentHandler
    promptHandler           *PromptHandler
    
    // Tool routing map for efficient lookup
    toolRoutes map[string]ToolHandler
}

// NewHandler creates a new MCP tools handler with all domain handlers
func NewHandler(
    epicService service.EpicService,
    userStoryService service.UserStoryService,
    requirementService service.RequirementService,
    searchService service.SearchServiceInterface,
    steeringDocumentService service.SteeringDocumentService,
    promptService *service.PromptService,
) *Handler {
    // Initialize domain handlers
    epicHandler := NewEpicHandler(epicService)
    userStoryHandler := NewUserStoryHandler(userStoryService, epicService)
    requirementHandler := NewRequirementHandler(requirementService, userStoryService)
    searchHandler := NewSearchHandler(searchService)
    steeringDocumentHandler := NewSteeringDocumentHandler(steeringDocumentService, epicService)
    promptHandler := NewPromptHandler(promptService)
    
    // Create tool routing map
    toolRoutes := map[string]ToolHandler{
        "create_epic":                  epicHandler,
        "update_epic":                  epicHandler,
        "create_user_story":            userStoryHandler,
        "update_user_story":            userStoryHandler,
        "create_requirement":           requirementHandler,
        "update_requirement":           requirementHandler,
        "create_relationship":          requirementHandler,
        "search_global":                searchHandler,
        "search_requirements":          searchHandler,
        "list_steering_documents":      steeringDocumentHandler,
        "create_steering_document":     steeringDocumentHandler,
        "get_steering_document":        steeringDocumentHandler,
        "update_steering_document":     steeringDocumentHandler,
        "link_steering_to_epic":        steeringDocumentHandler,
        "unlink_steering_from_epic":    steeringDocumentHandler,
        "get_epic_steering_documents":  steeringDocumentHandler,
        "create_prompt":                promptHandler,
        "update_prompt":                promptHandler,
        "delete_prompt":                promptHandler,
        "activate_prompt":              promptHandler,
        "list_prompts":                 promptHandler,
        "get_active_prompt":            promptHandler,
    }
    
    return &Handler{
        epicHandler:             epicHandler,
        userStoryHandler:        userStoryHandler,
        requirementHandler:      requirementHandler,
        searchHandler:           searchHandler,
        steeringDocumentHandler: steeringDocumentHandler,
        promptHandler:           promptHandler,
        toolRoutes:              toolRoutes,
    }
}

// HandleToolsCall processes tools/call requests by routing to appropriate domain handler
func (h *Handler) HandleToolsCall(ctx context.Context, params interface{}) (interface{}, error) {
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

    // Route to appropriate domain handler
    handler, exists := h.toolRoutes[toolName]
    if !exists {
        return nil, jsonrpc.NewMethodNotFoundError(fmt.Sprintf("Unknown tool: %s", toolName))
    }

    return handler.HandleTool(ctx, toolName, arguments)
}
```

### 4. Common Utilities and Types

#### Shared Helper Functions
**File**: `internal/mcp/tools/common.go`
**Package**: `package tools`

```go
package tools

// Common utilities shared across MCP tool handlers
func getUserFromContext(ctx context.Context) (*models.User, error)
func parseUUIDOrReferenceID(idStr string, getByRefFunc func(string) (interface{}, error)) (uuid.UUID, error)
func validateRequiredArgs(args map[string]interface{}, required []string) error
```

#### Response Types
**File**: `internal/mcp/types/responses.go`
**Package**: `package types`

```go
package types

// ToolResponse represents the response from a tool call
type ToolResponse struct {
    Content []ContentItem `json:"content"`
}

// ContentItem represents a single content item in a tool response
type ContentItem struct {
    Type string `json:"type"`
    Text string `json:"text"`
}

// CreateToolResponse creates a standard tool response with message and optional data
func CreateToolResponse(message string, data interface{}) *ToolResponse {
    content := []ContentItem{
        {
            Type: "text",
            Text: message,
        },
    }
    
    if data != nil {
        if jsonData, err := json.MarshalIndent(data, "", "  "); err == nil {
            content = append(content, ContentItem{
                Type: "text",
                Text: string(jsonData),
            })
        }
    }
    
    return &ToolResponse{Content: content}
}
```

#### Tool Schemas
**File**: `internal/mcp/schemas/tools.go`
**Package**: `package schemas`

Move the existing `GetSupportedTools()` function and related tool definitions to this dedicated package.

#### Error Handling
Centralized error handling patterns:
- Input validation errors → `jsonrpc.NewInvalidParamsError`
- Service layer errors → `jsonrpc.NewInternalError`
- Authentication errors → Custom JSON-RPC error codes
- Not found errors → Specific error codes

## Data Models

### Tool Response Structure
The existing `ToolResponse` and `ContentItem` structures will remain unchanged to maintain API compatibility:

```go
type ToolResponse struct {
    Content []ContentItem `json:"content"`
}

type ContentItem struct {
    Type string `json:"type"`
    Text string `json:"text"`
}
```

### Handler Registration
Tool routing will be implemented using a map-based approach for O(1) lookup performance, centralized in the main tools handler facade.

## Error Handling

### Centralized Error Mapping
Each handler will use consistent error handling patterns:

1. **Input Validation**: Validate required parameters and formats
2. **Service Layer Errors**: Map service errors to appropriate JSON-RPC errors
3. **Authentication**: Validate user context and permissions
4. **Reference ID Resolution**: Handle UUID vs reference ID parsing consistently

### Error Response Format
All handlers will return errors in the standard JSON-RPC format:

```go
// Invalid parameters
jsonrpc.NewInvalidParamsError("Missing or invalid 'title' argument")

// Internal service errors  
jsonrpc.NewInternalError(fmt.Sprintf("Failed to create epic: %v", err))

// Custom application errors
jsonrpc.NewJSONRPCError(-32002, "Insufficient permissions: Administrator role required", nil)
```

## Testing Strategy

### Unit Testing Approach

1. **Individual Handler Testing**: Each domain handler will have comprehensive unit tests
2. **Mock Dependencies**: Use interfaces to mock service dependencies
3. **Isolated Testing**: Test each handler independently without external dependencies
4. **Error Scenario Coverage**: Test all error paths and edge cases

### Test Structure
```
internal/mcp/tools/
├── epic_test.go
├── user_story_test.go
├── requirement_test.go
├── search_test.go
├── steering_document_test.go
├── prompt_test.go
├── handler_test.go (integration tests)
└── common_test.go (utility tests)
```

### Test Coverage Goals
- **Unit Tests**: 90%+ coverage for each individual handler
- **Integration Tests**: Maintain existing test coverage for `ToolsHandler` facade
- **Error Handling**: 100% coverage for error paths and validation

## Implementation Strategy

### Phase 1: Create MCP Package Structure
1. Create the new `internal/mcp/` package structure
2. Create `internal/mcp/types/responses.go` with common response types
3. Move tool schemas to `internal/mcp/schemas/tools.go`
4. Create `internal/mcp/tools/common.go` with shared utilities

### Phase 2: Create Domain Handlers
1. Create individual handler files in `internal/mcp/tools/` with their specific dependencies
2. Implement the `ToolHandler` interface for each domain
3. Move existing logic from `ToolsHandler` methods to domain handlers
4. Add comprehensive unit tests for each handler

### Phase 3: Create Tools Handler Facade
1. Create `internal/mcp/tools/handler.go` as the main facade
2. Implement tool routing using the handler map
3. Update the main `HandleToolsCall` to delegate to appropriate handlers
4. Maintain backward compatibility with existing API

### Phase 4: Update Main MCP Handler
1. Update `internal/handlers/mcp_handler.go` to use the new tools package
2. Replace the old `ToolsHandler` with the new `tools.Handler`
3. Update imports and dependencies
4. Ensure all existing functionality works

### Phase 5: Integration and Validation
1. Run comprehensive integration tests
2. Validate API compatibility
3. Performance testing to ensure no regression
4. Update documentation and examples
5. Remove old `mcp_tools_handler.go` and `mcp_tool_schemas.go` files


## Integration with Existing MCP Handler

### Updated MCP Handler Structure

The existing `internal/handlers/mcp_handler.go` will be updated to use the new tools package:

```go
// Updated MCPHandler structure
type MCPHandler struct {
    processor         *jsonrpc.Processor
    resourceHandler   *ResourceHandler
    toolsHandler      *tools.Handler  // Updated to use new tools package
    promptsHandler    *PromptsHandler
    initializeHandler *InitializeHandler
    mcpLogger         *MCPLogger
    errorMapper       *jsonrpc.ErrorMapper
    resourceService   service.ResourceService
}

// Updated constructor
func NewMCPHandler(
    epicService service.EpicService,
    userStoryService service.UserStoryService,
    requirementService service.RequirementService,
    acceptanceCriteriaService service.AcceptanceCriteriaService,
    searchService service.SearchServiceInterface,
    steeringDocumentService service.SteeringDocumentService,
    promptService *service.PromptService,
    resourceService service.ResourceService,
    requirementTypeRepo repository.RequirementTypeRepository,
) *MCPHandler {
    // ... existing initialization ...
    
    // Use new tools handler
    toolsHandler := tools.NewHandler(
        epicService,
        userStoryService,
        requirementService,
        searchService,
        steeringDocumentService,
        promptService,
    )
    
    // ... rest of initialization ...
}
```

### File Removal Plan

After successful migration, the following files will be removed:
- `internal/handlers/mcp_tools_handler.go` (1,500+ lines)
- `internal/handlers/mcp_tool_schemas.go` (500+ lines)

This will result in a net reduction of ~2,000 lines of code while improving maintainability.

## Benefits

### Code Quality Improvements
1. **Single Responsibility**: Each handler focuses on one domain
2. **Reduced Coupling**: Minimal dependencies per handler
3. **Improved Testability**: Isolated, focused unit tests
4. **Better Maintainability**: Smaller, more focused code files

### Development Experience
1. **Easier Debugging**: Isolated domain logic
2. **Faster Development**: Clear separation of concerns
3. **Reduced Merge Conflicts**: Multiple developers can work on different domains
4. **Clearer Code Reviews**: Smaller, focused changes

### Future Extensibility
1. **New Domains**: Easy to add new domain handlers
2. **Tool Addition**: Simple to add new tools to existing domains
3. **Service Evolution**: Handlers can evolve independently
4. **Testing Strategy**: Consistent testing patterns across domains

