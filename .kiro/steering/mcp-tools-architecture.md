# MCP Tools Architecture Guide

## Tool Implementation Pattern

MCP tools are grouped by domain in handlers located in `internal/mcp/tools/`:

### Handler Organization
- **EpicHandler** (`epic.go`) - all epic-related tools: `create_epic`, `update_epic`, `list_epics`, `epic_hierarchy`
- **UserStoryHandler** - all user story tools
- **RequirementHandler** - all requirement tools
- Each handler has `GetSupportedTools()` and `HandleTool()` methods
- **Rule**: Add new tools as methods to existing handlers, NOT as new handlers

### Adding a New Tool to Existing Handler

Example: Adding `epic_hierarchy` tool to `EpicHandler`:

```go
// 1. Add constant in tools package
const ToolEpicHierarchy = "epic_hierarchy"

// 2. Update GetSupportedTools() to include new tool
func (h *EpicHandler) GetSupportedTools() []string {
    return []string{
        ToolCreateEpic,
        ToolUpdateEpic,
        ToolListEpics,
        ToolEpicHierarchy, // NEW
    }
}

// 3. Add case in HandleTool() switch
func (h *EpicHandler) HandleTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
    switch toolName {
    case ToolListEpics:
        return h.List(ctx, args)
    case ToolCreateEpic:
        return h.Create(ctx, args)
    case ToolUpdateEpic:
        return h.Update(ctx, args)
    case ToolEpicHierarchy: // NEW
        return h.GetHierarchy(ctx, args)
    default:
        return nil, jsonrpc.NewMethodNotFoundError(fmt.Sprintf("Unknown Epic tool: %s", toolName))
    }
}

// 4. Implement method
func (h *EpicHandler) GetHierarchy(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    // Implementation here
}
```

### Tool Schema Registration

Add tool definition to `internal/mcp/schemas/tools.go`:

```go
{
    Name:        "epic_hierarchy",
    Title:       "View Epic Hierarchy",
    Description: "Display the complete hierarchical structure...",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "epic": map[string]interface{}{
                "type":        "string",
                "description": "Epic reference ID (e.g., EP-001) or UUID",
            },
        },
        "required": []string{"epic"},
    },
}
```

## Layered Architecture for MCP Tools

### Service Layer (`internal/service/`)

**Responsibilities**:
- Business logic and data retrieval ONLY
- Returns domain models (`Epic`, `UserStory`, `Requirement`, etc.)
- No formatting, no presentation logic
- Reusable across different interfaces (API, MCP, CLI)

**Example**:
```go
// Service returns domain model
func (s *epicService) GetEpicWithCompleteHierarchy(id uuid.UUID) (*models.Epic, error) {
    epic, err := s.epicRepo.GetCompleteHierarchy(id)
    if err != nil {
        if errors.Is(err, repository.ErrNotFound) {
            return nil, ErrEpicNotFound
        }
        return nil, fmt.Errorf("failed to get epic hierarchy: %w", err)
    }
    return epic, nil // Returns Epic model, no formatting
}
```

### Tool Layer (`internal/mcp/tools/`)

**Responsibilities**:
- Presentation logic and formatting
- Converts domain models to user-facing output
- Handles MCP-specific response format
- Input validation and parameter parsing

**Example**:
```go
// Tool formats the output
func (h *EpicHandler) GetHierarchy(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    // 1. Parse and validate input
    epicID, err := parseUUIDOrReferenceID(epicIDStr, ...)
    
    // 2. Get data from service
    epic, err := h.epicService.GetEpicWithCompleteHierarchy(epicID)
    
    // 3. Format output (presentation logic)
    treeOutput := h.formatTree(epic)
    
    // 4. Return MCP response
    return types.CreateDataResponse(treeOutput, nil), nil
}

// Private formatting methods in tool handler
func (h *EpicHandler) formatTree(epic *models.Epic) string {
    // ASCII tree formatting logic here
}
```

### Golden Rule

**If it's about HOW data looks → Tool Layer**
**If it's about WHAT data to get → Service Layer**

Examples:
- ✅ Service: `GetEpicWithCompleteHierarchy()` - retrieves data
- ✅ Tool: `formatTree()` - formats as ASCII tree
- ❌ Service: `GetEpicAsASCIITree()` - mixing concerns
- ❌ Tool: `GetEpicFromDatabase()` - business logic in presentation

## Reference ID and UUID Support

All MCP tools should accept both UUID and Reference ID formats using the helper from `common.go`:

```go
// Use existing helper
epicID, err := parseUUIDOrReferenceID(epicIDStr, func(refID string) (interface{}, error) {
    return h.epicService.GetEpicByReferenceID(refID)
})
if err != nil {
    return nil, jsonrpc.NewInvalidParamsError("Invalid epic ID: not a valid UUID or reference ID")
}
```

**How it works**:
1. Tries to parse as UUID first
2. If fails, treats as reference ID (EP-XXX, US-XXX, REQ-XXX, etc.)
3. Calls provided function to resolve reference ID to entity
4. Returns UUID for further processing

**Supported formats**:
- UUID: `62d1be3e-17ee-4303-a650-e0ca8be7d9df`
- Reference ID: `EP-021`, `US-064`, `REQ-089`, `AC-021`

## Error Handling Pattern

Follow existing patterns from other tools:

```go
func (h *EpicHandler) GetHierarchy(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    // Validation errors
    if !ok || epicIDStr == "" {
        return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'epic' argument")
    }
    
    // Not found errors
    if errors.Is(err, service.ErrEpicNotFound) {
        return nil, jsonrpc.NewInvalidParamsError("Epic not found")
    }
    
    // Internal errors (don't expose details)
    return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to retrieve hierarchy: %v", err))
}
```

## Testing Strategy

### Unit Tests
- Test each tool method independently
- Mock service layer dependencies
- Test error handling paths
- Test input validation

### Integration Tests
- Test complete flow through MCP handler
- Use real database (PostgreSQL with testcontainers)
- Verify JSON-RPC response format
- Test with actual data

Example test structure:
```go
func TestEpicHandler_GetHierarchy(t *testing.T) {
    // Setup
    handler := NewEpicHandler(mockEpicService, mockUserService)
    
    // Test cases
    t.Run("valid reference ID", func(t *testing.T) { ... })
    t.Run("valid UUID", func(t *testing.T) { ... })
    t.Run("epic not found", func(t *testing.T) { ... })
    t.Run("invalid format", func(t *testing.T) { ... })
}
```
