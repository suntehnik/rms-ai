# Design Document

## Overview

This design document outlines the implementation approach for adding status management capabilities to MCP (Model Context Protocol) tools. The solution extends existing MCP update tools with optional status parameters while maintaining backward compatibility and providing consistent validation across all entity types.

## Architecture

### Current MCP Tools Architecture

The existing MCP tools are implemented in the `internal/server/routes/mcp.go` file and use the following pattern:

1. **Tool Registration**: Tools are registered with JSON schemas defining their parameters
2. **Parameter Validation**: Input parameters are validated against the schema
3. **Business Logic**: Core operations are performed using service layer methods
4. **Response Formation**: Results are formatted and returned to the MCP client

### Enhanced Architecture for Status Management

The enhanced architecture maintains the same pattern while adding status management capabilities:

```
MCP Client Request
       ↓
Tool Parameter Validation (Enhanced with Status)
       ↓
Status Validation Layer (New)
       ↓
Service Layer Operations (Enhanced)
       ↓
Database Updates (Enhanced)
       ↓
Response Formation (Enhanced)
       ↓
MCP Client Response
```

## Components and Interfaces

### 1. MCP Tool Schema Updates

**Enhanced Tool Schemas:**
- `update_epic`: Add optional `status` parameter with enum validation
- `update_user_story`: Add optional `status` parameter with enum validation  
- `update_requirement`: Add optional `status` parameter with enum validation

**Schema Structure:**
```json
{
  "type": "object",
  "properties": {
    "epic_id": {"type": "string"},
    "title": {"type": "string"},
    "description": {"type": "string"},
    "priority": {"type": "integer", "minimum": 1, "maximum": 4},
    "assignee_id": {"type": "string"},
    "status": {
      "type": "string",
      "enum": ["Backlog", "Draft", "In Progress", "Done", "Cancelled"]
    }
  },
  "required": ["epic_id"]
}
```

### 2. Status Validation Component

**Purpose**: Centralized status validation logic for all entity types

**Interface:**
```go
type StatusValidator interface {
    ValidateEpicStatus(status string) error
    ValidateUserStoryStatus(status string) error
    ValidateRequirementStatus(status string) error
}
```

**Implementation:**
- Validate status values against allowed enums
- Provide consistent error messages
- Support case-insensitive validation
- Return structured validation errors

### 3. Service Layer Enhancements

**Enhanced Service Methods:**
```go
// Epic Service
func (s *EpicService) UpdateEpic(ctx context.Context, id string, updates EpicUpdateRequest) (*Epic, error)

// User Story Service  
func (s *UserStoryService) UpdateUserStory(ctx context.Context, id string, updates UserStoryUpdateRequest) (*UserStory, error)

// Requirement Service
func (s *RequirementService) UpdateRequirement(ctx context.Context, id string, updates RequirementUpdateRequest) (*Requirement, error)
```

**Update Request Structures:**
```go
type EpicUpdateRequest struct {
    Title       *string `json:"title,omitempty"`
    Description *string `json:"description,omitempty"`
    Priority    *int    `json:"priority,omitempty"`
    AssigneeID  *string `json:"assignee_id,omitempty"`
    Status      *string `json:"status,omitempty"`
}
```

### 4. Database Layer Updates

**GORM Model Updates:**
- No schema changes required (status fields already exist)
- Enhanced update operations to handle status changes
- Automatic timestamp updates for modified entities

**Update Operations:**
```go
// Example update operation
func (r *EpicRepository) Update(ctx context.Context, id string, updates map[string]interface{}) (*Epic, error) {
    var epic Epic
    result := r.db.WithContext(ctx).Model(&epic).Where("id = ?", id).Updates(updates)
    if result.Error != nil {
        return nil, result.Error
    }
    return r.GetByID(ctx, id)
}
```

## Data Models

### Status Enums

**Epic Status Values:**
- `Backlog`: Initial state for new epics
- `Draft`: Epic is being planned and refined
- `In Progress`: Epic is actively being worked on
- `Done`: Epic is completed
- `Cancelled`: Epic is cancelled and will not be completed

**User Story Status Values:**
- `Backlog`: Initial state for new user stories
- `Draft`: User story is being refined
- `In Progress`: User story is actively being developed
- `Done`: User story is completed
- `Cancelled`: User story is cancelled

**Requirement Status Values:**
- `Draft`: Initial state for new requirements
- `Active`: Requirement is approved and active
- `Obsolete`: Requirement is no longer valid

### Request/Response Models

**MCP Tool Request:**
```json
{
  "user_story_id": "US-048",
  "status": "In Progress"
}
```

**MCP Tool Response:**
```json
{
  "id": "ad1cab8e-9e3d-4093-b248-f3fe204ed24e",
  "reference_id": "US-048",
  "title": "User Story Title",
  "status": "In Progress",
  "updated_at": "2025-11-02T22:30:00Z",
  ...
}
```

## Error Handling

### Validation Errors

**Invalid Status Error:**
```json
{
  "error": "Invalid status 'InvalidStatus' for user story. Valid statuses are: Backlog, Draft, In Progress, Done, Cancelled"
}
```

**Entity Not Found Error:**
```json
{
  "error": "User story with ID 'US-999' not found"
}
```

**General Validation Error:**
```json
{
  "error": "Validation failed: status is required when provided and must be one of the allowed values"
}
```

### Error Handling Strategy

1. **Schema Validation**: JSON schema validates basic parameter types and required fields
2. **Business Validation**: Custom validation for status values and entity existence
3. **Database Errors**: Handle database constraint violations and connection issues
4. **Consistent Formatting**: All errors follow the same response format
5. **Helpful Messages**: Error messages include valid options and clear guidance

## Testing Strategy

### Unit Tests

**Status Validation Tests:**
- Test valid status values for each entity type
- Test invalid status values and error messages
- Test case-insensitive status validation
- Test empty and null status handling

**Service Layer Tests:**
- Test status updates with valid values
- Test status updates with invalid values
- Test backward compatibility (updates without status)
- Test database transaction handling

**MCP Tool Tests:**
- Test tool parameter validation
- Test tool execution with status parameters
- Test tool responses include updated status
- Test error handling and response formatting

### Integration Tests

**End-to-End Status Management:**
- Test complete status change workflows
- Test status changes through MCP tool interface
- Test database persistence of status changes
- Test response consistency across entity types

**Backward Compatibility Tests:**
- Test existing MCP tool functionality remains unchanged
- Test mixed parameter updates (status + other fields)
- Test existing client integrations continue to work

### Performance Tests

**Status Update Performance:**
- Measure status update operation latency
- Test concurrent status updates
- Validate no performance regression for existing operations
- Test database query optimization for status updates

## Implementation Plan

### Phase 1: Core Infrastructure
1. Update MCP tool schemas with status parameters
2. Implement status validation component
3. Enhance service layer methods for status handling
4. Add comprehensive unit tests

### Phase 2: MCP Tool Integration
1. Update MCP tool handlers to process status parameters
2. Integrate status validation into tool execution
3. Enhance error handling and response formatting
4. Add integration tests for MCP tools

### Phase 3: Testing and Validation
1. Comprehensive testing across all entity types
2. Backward compatibility validation
3. Performance testing and optimization
4. Documentation updates

### Phase 4: Deployment and Monitoring
1. Deploy to staging environment
2. Validate functionality with real MCP clients
3. Monitor performance and error rates
4. Production deployment with rollback plan

## Security Considerations

### Authorization
- Maintain existing authorization checks for entity updates
- Ensure status changes respect user permissions
- Validate user has rights to modify specific entities

### Input Validation
- Strict validation of status parameter values
- Protection against injection attacks through parameter validation
- Sanitization of all input parameters

### Audit Trail
- Log all status changes with user identification
- Maintain audit trail for compliance requirements
- Track status change history for debugging

## Performance Considerations

### Database Optimization
- Use existing indexes for entity lookups
- Minimize database queries per status update
- Leverage GORM's optimized update operations

### Caching Strategy
- Maintain existing caching mechanisms
- Consider caching valid status values
- Invalidate relevant caches on status updates

### Response Optimization
- Return only necessary entity data in responses
- Use efficient JSON serialization
- Minimize response payload size

## Monitoring and Observability

### Metrics
- Track status update operation counts by entity type
- Monitor status update success/failure rates
- Measure status update operation latency

### Logging
- Log all status change attempts with outcomes
- Include entity IDs and user context in logs
- Log validation failures with details

### Alerting
- Alert on high error rates for status updates
- Monitor for unusual status change patterns
- Alert on performance degradation