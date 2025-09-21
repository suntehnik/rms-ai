# Design Document

## Overview

This design addresses the synchronization of API documentation with the actual implementation by updating the OpenAPI specification and steering documentation to accurately reflect all implemented endpoints, standardizing response formats, and ensuring comprehensive coverage of the comment system, deletion workflows, and authentication requirements.

## Architecture

### Documentation Structure
```
docs/
├── openapi-v3.yaml           # Complete OpenAPI specification
└── api-client-export.md      # Client implementation guide

.kiro/steering/
└── api-client-export.md      # Steering documentation (updated)
```

### Endpoint Categories to Address

1. **Missing Documented Endpoints**
   - Comprehensive deletion endpoints
   - Entity comment endpoints (all CRUD + inline operations)
   - Additional navigation endpoints


2. **Response Format Standardization**
   - Configuration endpoints alignment
   - Error response consistency
   - Authentication documentation

3. **TypeScript Interface Completeness**
   - Deletion workflow types
   - Comment system types
   - Status management types

## Components and Interfaces

### 1. OpenAPI Specification Updates

#### Missing Endpoint Paths
```yaml
# Comprehensive Deletion Endpoints
/api/v1/{entity_type}/{id}/validate-deletion:
  get:
    summary: Validate entity deletion
    responses:
      '200':
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/DependencyInfo'

/api/v1/{entity_type}/{id}/delete:
  delete:
    summary: Comprehensive entity deletion
    responses:
      '200':
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/DeletionResult'

# Entity Comment Endpoints (for each entity type)
/api/v1/{entity_type}/{id}/comments:
  get:
    summary: Get entity comments
  post:
    summary: Create entity comment

/api/v1/{entity_type}/{id}/comments/inline:
  post:
    summary: Create inline comment

/api/v1/{entity_type}/{id}/comments/inline/visible:
  get:
    summary: Get visible inline comments

/api/v1/{entity_type}/{id}/comments/inline/validate:
  post:
    summary: Validate inline comments


```

#### New Schema Definitions
```yaml
components:
  schemas:
    DependencyInfo:
      type: object
      required: [can_delete, dependencies, warnings]
      properties:
        can_delete:
          type: boolean
        dependencies:
          type: array
          items:
            $ref: '#/components/schemas/DependencyItem'
        warnings:
          type: array
          items:
            type: string

    DependencyItem:
      type: object
      required: [entity_type, entity_id, reference_id, title, dependency_type]
      properties:
        entity_type:
          type: string
        entity_id:
          type: string
        reference_id:
          type: string
        title:
          type: string
        dependency_type:
          type: string

    DeletionResult:
      type: object
      required: [success, deleted_entities, message]
      properties:
        success:
          type: boolean
        deleted_entities:
          type: array
          items:
            $ref: '#/components/schemas/DeletedEntity'
        message:
          type: string

    DeletedEntity:
      type: object
      required: [entity_type, entity_id, reference_id]
      properties:
        entity_type:
          type: string
        entity_id:
          type: string
        reference_id:
          type: string

    CommentListResponse:
      allOf:
        - $ref: '#/components/schemas/ListResponse'
        - type: object
          properties:
            data:
              type: array
              items:
                $ref: '#/components/schemas/Comment'

    InlineCommentValidationRequest:
      type: object
      required: [comments]
      properties:
        comments:
          type: array
          items:
            $ref: '#/components/schemas/InlineCommentPosition'

    InlineCommentPosition:
      type: object
      required: [comment_id, text_position_start, text_position_end]
      properties:
        comment_id:
          type: string
        text_position_start:
          type: integer
        text_position_end:
          type: integer

    HealthCheckResponse:
      type: object
      required: [status]
      properties:
        status:
          type: string
        reason:
          type: string
```

### 2. Authentication Documentation

#### Security Scheme Updates
```yaml
security:
  - BearerAuth: []  # Default for all endpoints

paths:
  /auth/login:
    post:
      security: []  # Public endpoint
  


  /api/v1/config/**:
    # All config endpoints require admin role
    security:
      - BearerAuth: []
    # Add custom extension for role requirement
    x-required-role: Administrator
```

### 3. Response Format Standardization

#### Configuration Endpoint Alignment
All configuration list endpoints will use the standard `ListResponse` format:

```yaml
# Before (inconsistent)
RequirementTypeListResponse:
  properties:
    requirement_types: [...]
    count: number

# After (standardized)
RequirementTypeListResponse:
  allOf:
    - $ref: '#/components/schemas/ListResponse'
    - type: object
      properties:
        data:
          type: array
          items:
            $ref: '#/components/schemas/RequirementType'
```

## Data Models

### Comment System Models

```typescript
// Enhanced comment interfaces
interface CommentThread {
  parent_comment: Comment;
  replies: Comment[];
  total_replies: number;
}

interface InlineCommentContext {
  linked_text: string;
  text_position_start: number;
  text_position_end: number;
  is_valid: boolean;
}

interface CommentResolution {
  is_resolved: boolean;
  resolved_by?: User;
  resolved_at?: string;
  resolution_note?: string;
}
```

### Deletion System Models

```typescript
interface DeletionValidation {
  entity_type: string;
  entity_id: string;
  can_delete: boolean;
  blocking_dependencies: DependencyItem[];
  cascade_dependencies: DependencyItem[];
  warnings: string[];
}

interface DeletionPlan {
  primary_entity: EntityReference;
  cascade_deletions: EntityReference[];
  dependency_updates: DependencyUpdate[];
  estimated_impact: number;
}
```

## Error Handling

### Standardized Error Responses

```yaml
components:
  responses:
    DeletionConflict:
      description: Cannot delete due to dependencies
      content:
        application/json:
          schema:
            allOf:
              - $ref: '#/components/schemas/ErrorResponse'
              - type: object
                properties:
                  error:
                    properties:
                      dependencies:
                        type: array
                        items:
                          $ref: '#/components/schemas/DependencyItem'

    ValidationError:
      description: Request validation failed
      content:
        application/json:
          schema:
            allOf:
              - $ref: '#/components/schemas/ErrorResponse'
              - type: object
                properties:
                  error:
                    properties:
                      validation_errors:
                        type: array
                        items:
                          $ref: '#/components/schemas/ValidationError'
```

## Testing Strategy

### Documentation Validation
1. **OpenAPI Validation**: Use OpenAPI validators to ensure specification correctness
2. **Route Coverage**: Automated tests to verify all routes have documentation
3. **Schema Validation**: Validate that response schemas match actual responses
4. **Authentication Testing**: Verify security requirements are properly enforced

### Implementation Verification
1. **Endpoint Existence**: Test that all documented endpoints exist
2. **Response Format**: Validate actual responses match documented schemas
3. **Authentication Flow**: Test authentication requirements for each endpoint
4. **Error Scenarios**: Verify error responses match documented formats

## Implementation Phases

### Phase 1: Core Documentation Updates
1. Add missing endpoints to OpenAPI specification
2. Define new schemas for deletion and comment systems
3. Update authentication documentation
4. Standardize response formats

### Phase 2: TypeScript Interface Updates
1. Generate TypeScript interfaces from updated OpenAPI spec
2. Add deletion workflow types
3. Enhance comment system types
4. Update client export documentation

### Phase 3: Validation and Testing
1. Implement automated documentation validation
2. Create tests for endpoint coverage
3. Validate response format consistency
4. Test authentication requirements

### Phase 4: Documentation Deployment
1. Update steering documentation
2. Generate client SDKs from updated specification
3. Create migration guide for API consumers
4. Update developer documentation

## Design Decisions

### 1. Comprehensive vs. Incremental Updates
**Decision**: Comprehensive update of all documentation at once
**Rationale**: Ensures consistency and prevents partial documentation states

### 2. Response Format Standardization
**Decision**: Align all list endpoints to use standard `ListResponse` format
**Rationale**: Provides consistency for API consumers and simplifies client implementation

### 3. Authentication Documentation Strategy
**Decision**: Use OpenAPI security schemes with custom extensions for role requirements
**Rationale**: Leverages standard OpenAPI features while providing clear role-based access documentation

### 4. Deletion System Documentation
**Decision**: Document both validation and execution endpoints separately
**Rationale**: Allows clients to implement safe deletion workflows with proper user confirmation

### 5. Comment System Completeness
**Decision**: Document all comment operations including inline functionality
**Rationale**: The comment system is a core feature that needs complete API coverage for proper client implementation