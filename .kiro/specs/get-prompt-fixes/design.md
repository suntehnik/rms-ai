# Design Document: MCP Prompt Response Compliance

## Overview

This document outlines the design for fixing the MCP `prompts/get` response to comply with the 18 June 2025 MCP Prompt Specification. The current implementation violates the specification by using forbidden role values (`"system"`) and returning content as plain strings instead of structured content chunks.

## Architecture

### Current State Analysis

The current implementation in `internal/service/prompt_service.go` generates MCP prompt definitions with:
- `role: "system"` (violates specification)
- `content` as plain string (violates specification)

```go
// Current problematic implementation
Messages: []models.PromptMessage{
    {
        Role:    "system",  // ❌ Invalid role
        Content: prompt.Content,  // ❌ Should be structured content
    },
},
```

### Target State

The compliant implementation must:
- Use only `"user"` or `"assistant"` roles
- Structure content as an array of typed objects
- Validate prompt data before serialization
- Handle invalid stored data gracefully

## Components and Interfaces

### 1. Enhanced Prompt Message Model

**Design Decision**: Extend the existing `PromptMessage` struct to support structured content while maintaining backward compatibility.

**Rationale**: This approach allows gradual migration and doesn't break existing stored data.

```go
// Enhanced PromptMessage structure
type PromptMessage struct {
    Role    string         `json:"role"`    // "user" or "assistant" only
    Content []ContentChunk `json:"content"` // Array of typed content objects
}

type ContentChunk struct {
    Type string `json:"type"` // "text", "image", etc.
    Text string `json:"text,omitempty"` // For type="text"
    // Future: additional fields for other content types
}
```

### 2. Prompt Message Role Enum

**Design Decision**: Create a strict enum for valid MCP roles with validation.

**Rationale**: Prevents invalid roles from being used and provides clear error messages.

```go
type MCPRole string

const (
    MCPRoleUser      MCPRole = "user"
    MCPRoleAssistant MCPRole = "assistant"
)

func (r MCPRole) IsValid() bool {
    return r == MCPRoleUser || r == MCPRoleAssistant
}
```

### 3. Enhanced Prompt Model

**Design Decision**: Add a `Role` field to the Prompt model to persist the intended role.

**Rationale**: Satisfies FR4 requirement for explicit role persistence and eliminates runtime role inference.

```go
type Prompt struct {
    // ... existing fields ...
    
    // Role specifies the message role for MCP compliance
    Role MCPRole `gorm:"type:varchar(20);not null;default:'assistant'" json:"role"`
    
    // ... existing fields ...
}
```

### 4. Validation Layer

**Design Decision**: Implement validation at the service layer before MCP serialization.

**Rationale**: Provides guard rails (FR3) and prevents malformed responses from reaching clients.

```go
type PromptValidator struct{}

func (v *PromptValidator) ValidateForMCP(prompt *models.Prompt) error {
    if !prompt.Role.IsValid() {
        return fmt.Errorf("invalid role '%s': must be 'user' or 'assistant'", prompt.Role)
    }
    
    if prompt.Content == "" {
        return errors.New("content cannot be empty")
    }
    
    return nil
}
```

## Data Models

### Database Schema Changes

**Migration Strategy**: Add the new `role` field with a default value to maintain backward compatibility.

```sql
-- Migration: Add role field to prompts table
ALTER TABLE prompts 
ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'assistant';

-- Add check constraint for valid roles
ALTER TABLE prompts 
ADD CONSTRAINT check_prompt_role 
CHECK (role IN ('user', 'assistant'));
```

### Content Structure Transformation

**Design Decision**: Transform plain string content into structured content chunks at the service layer.

**Rationale**: Keeps the transformation logic centralized and allows for future content type extensions.

```go
func (ps *PromptService) transformContentToChunks(content string) []ContentChunk {
    return []ContentChunk{
        {
            Type: "text",
            Text: content,
        },
    }
}
```

## Error Handling

### Validation Error Response

**Design Decision**: Return MCP-compliant error responses for validation failures.

**Rationale**: Satisfies FR3 requirement for descriptive error reporting instead of malformed success responses.

```go
func (ps *PromptService) GetMCPPromptDefinition(ctx context.Context, name string) (*models.MCPPromptDefinition, error) {
    prompt, err := ps.GetByName(ctx, name)
    if err != nil {
        return nil, err
    }
    
    // Validate prompt before serialization
    if err := ps.validator.ValidateForMCP(prompt); err != nil {
        ps.logger.WithError(err).WithField("prompt_name", name).Error("Prompt validation failed")
        return nil, fmt.Errorf("invalid prompt data: %w", err)
    }
    
    // ... rest of implementation
}
```

### Error Logging and Observability

**Design Decision**: Add structured logging for validation failures.

**Rationale**: Satisfies non-functional requirement for observability to help operations teams identify issues.

```go
func (ps *PromptService) logValidationFailure(ctx context.Context, prompt *models.Prompt, err error) {
    ps.logger.WithFields(logrus.Fields{
        "prompt_id":    prompt.ID,
        "prompt_name":  prompt.Name,
        "role":         prompt.Role,
        "error":        err.Error(),
        "component":    "prompt_service",
        "operation":    "mcp_validation",
    }).Error("MCP prompt validation failed")
}
```

## Testing Strategy

### Unit Tests

**Coverage Areas**:
1. Role validation with valid and invalid values
2. Content chunk transformation
3. MCP prompt definition generation
4. Error handling for invalid stored data

```go
func TestPromptService_GetMCPPromptDefinition_ValidRole(t *testing.T) {
    // Test with valid "assistant" role
    // Test with valid "user" role
}

func TestPromptService_GetMCPPromptDefinition_InvalidRole(t *testing.T) {
    // Test with invalid "system" role
    // Verify error response instead of malformed success
}
```

### Integration Tests

**Coverage Areas**:
1. End-to-end MCP `prompts/get` request flow
2. Database migration and backward compatibility
3. Error response format compliance

## Migration Strategy

### Backward Compatibility

**Design Decision**: Use a phased migration approach.

**Rationale**: Minimizes disruption to existing functionality while ensuring compliance.

**Phase 1**: Add role field with default value
- All existing prompts get `role: "assistant"`
- New prompts can specify role explicitly

**Phase 2**: Update MCP response generation
- Transform content to structured format
- Add validation layer

**Phase 3**: Optional data cleanup
- Review and update existing prompt roles if needed

### Data Migration Script

```go
func MigrateExistingPrompts(db *gorm.DB) error {
    // Update existing prompts to have explicit role
    return db.Model(&models.Prompt{}).
        Where("role = '' OR role IS NULL").
        Update("role", "assistant").Error
}
```

## Implementation Plan

### Service Layer Changes

1. **Update PromptService.GetMCPPromptDefinition()**:
   - Add validation before serialization
   - Transform content to structured format
   - Use prompt.Role instead of hardcoded "system"

2. **Add validation methods**:
   - Role validation
   - Content validation
   - Error logging

### Model Layer Changes

1. **Update Prompt model**:
   - Add Role field
   - Update validation tags
   - Add migration hooks

2. **Update PromptMessage model**:
   - Change Content from string to []ContentChunk
   - Add ContentChunk struct

### Handler Layer Changes

1. **Update PromptsHandler**:
   - Enhanced error handling
   - Validation error mapping

## Security Considerations

### Input Validation

**Design Decision**: Validate role values at multiple layers (model, service, handler).

**Rationale**: Defense in depth approach prevents invalid data from entering the system.

### Error Information Disclosure

**Design Decision**: Log detailed errors internally but return generic error messages to clients.

**Rationale**: Prevents information leakage while maintaining debuggability.

## Performance Considerations

### Content Transformation Overhead

**Impact**: Minimal - transformation is a simple string-to-struct operation.

**Mitigation**: Consider caching transformed content if performance becomes an issue.

### Database Query Impact

**Impact**: Adding role field has negligible impact on existing queries.

**Optimization**: Role field is indexed for potential future filtering needs.

## Monitoring and Observability

### Metrics to Track

1. **Validation failure rate**: Track how often stored prompts fail MCP validation
2. **Error response rate**: Monitor `prompts/get` error responses
3. **Migration progress**: Track role field population during migration

### Logging Strategy

```go
// Success case
logger.WithFields(logrus.Fields{
    "prompt_name": name,
    "role": prompt.Role,
    "content_chunks": len(contentChunks),
}).Info("MCP prompt definition generated successfully")

// Validation failure case
logger.WithFields(logrus.Fields{
    "prompt_name": name,
    "validation_error": err.Error(),
    "role": prompt.Role,
}).Error("MCP prompt validation failed")
```

## Open Questions Resolution

### Question 1: Additional Content Types

**Decision**: Start with `"text"` type only, design for extensibility.

**Rationale**: Meets current requirements while allowing future expansion for image or other content types.

### Question 2: Invalid Stored Prompts

**Decision**: Fail the request with descriptive error rather than automatic remediation.

**Rationale**: Explicit failure is safer and more transparent than silent data modification. Administrators can fix data issues manually.

## Acceptance Criteria Mapping

### FR1: Valid Roles in Prompt Messages
- ✅ Only "user" or "assistant" roles in responses
- ✅ Validation rejects forbidden roles
- ✅ Satisfies AC-014 role requirements

### FR2: Structured Content Payload
- ✅ Content as array of typed objects
- ✅ Each object has "type" field
- ✅ Text content in "text" field
- ✅ No raw strings in content

### FR3: Guard Rails and Error Reporting
- ✅ Validation before serialization
- ✅ MCP error response for invalid data
- ✅ Structured logging for operations teams

### FR4: Persisted Prompt Message Role Field
- ✅ Explicit role field in Prompt model
- ✅ Valid enum values persisted
- ✅ No runtime role inference needed

This design ensures full compliance with the MCP specification while maintaining backward compatibility and providing robust error handling.