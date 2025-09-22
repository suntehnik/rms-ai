# Implementation Plan

- [x] 1. Update OpenAPI specification with missing endpoints
  - Add comprehensive deletion endpoints for all entity types
  - Add entity comment endpoints (CRUD and inline operations) for all entities
  - Add missing navigation hierarchy endpoints
  - Define new schemas for deletion and comment system responses
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 5.1, 5.2, 5.3_

- [x] 2. Standardize response formats in OpenAPI specification
  - Update configuration endpoint responses to use standard ListResponse format
  - Ensure all list endpoints follow consistent response structure
  - Standardize error response formats across all endpoints
  - _Requirements: 2.1, 2.2, 2.3, 6.4_

- [x] 3. Document authentication and authorization requirements
  - Add security requirements to all protected endpoints
  - Mark admin-only endpoints with appropriate security schemes
  - Document public endpoints (login, etc.) with no security requirements
  - Add custom extensions for role-based access requirements
  - _Requirements: 3.1, 3.2, 3.3, 3.4_

- [x] 4. Add deletion system schemas and endpoints
  - Define DependencyInfo schema for deletion validation responses
  - Define DeletionResult schema for deletion operation responses
  - Define DependencyItem and DeletedEntity supporting schemas
  - Add validate-deletion endpoints for all entity types
  - Add comprehensive delete endpoints for all entity types
  - Add general deletion confirmation endpoint
  - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [x] 5. Add comprehensive comment system documentation
  - Define CommentListResponse schema using standard ListResponse format
  - Define InlineCommentValidationRequest and supporting schemas
  - Add general comment endpoints (get, update, delete, resolve, unresolve)
  - Add comment reply endpoints (get replies, create reply)
  - Add comment status filtering endpoints
  - _Requirements: 4.1, 4.4_

- [x] 6. Add entity-specific comment endpoints
  - Add comment endpoints for Epic entities (CRUD, inline, validation)
  - Add comment endpoints for UserStory entities (CRUD, inline, validation)
  - Add comment endpoints for AcceptanceCriteria entities (CRUD, inline, validation)
  - Add comment endpoints for Requirement entities (CRUD, inline, validation)
  - _Requirements: 4.2, 4.3_

- [x] 7. Update TypeScript interfaces in steering documentation
  - Add deletion workflow TypeScript interfaces (DependencyInfo, DeletionResult, etc.)
  - Add enhanced comment system TypeScript interfaces
  - Add inline comment validation TypeScript interfaces
  - Update existing interfaces to match standardized response formats
  - _Requirements: 7.1, 7.2, 7.3, 7.4_

- [x] 8. Validate OpenAPI specification completeness
  - Verify all routes from routes.go have corresponding OpenAPI documentation
  - Ensure all documented endpoints have proper request/response schemas
  - Validate that parameter definitions match implementation
  - Check that all entity types are covered consistently
  - _Requirements: 6.1, 6.2, 6.3_

- [x] 9. Update steering documentation with complete API reference
  - Sync the steering api-client-export.md with updated OpenAPI specification
  - Add comprehensive endpoint tables for all missing functionality
  - Update TypeScript interface examples with new types
  - Add implementation notes for deletion workflows and comment system
  - _Requirements: 1.1, 4.1, 5.1, 7.1_

- [x] 10. Generate updated API documentation from OpenAPI specification
  - Generate HTML documentation from updated OpenAPI specification
  - Create interactive API documentation with request/response examples
  - Generate client SDK documentation with TypeScript interfaces
  - Update developer documentation with new endpoints and workflows
  - Generate comprehensive API documentation in MD format suitable for another models and coding agents
  - _Requirements: 2.5, 4.5, 7.1_

- [x] 11. Fix epic handler implementation vs documentation discrepancies
  - [x] 11.1 Fix list epics response format to use standard ListResponse schema
    - Update epic handler ListEpics method to return {data, total_count, limit, offset} instead of {epics, count}
    - Ensure response matches EpicListResponse schema from OpenAPI specification
  - [x] 11.2 Add missing creator_id field to CreateEpicRequest schema in OpenAPI
    - Update CreateEpicRequest schema to include required creator_id field
    - Ensure schema matches the actual service.CreateEpicRequest structure
  - [x] 11.3 Standardize error response formats in epic handler
    - Update all error responses to use standard ErrorResponse format with {error: {code, message}}
    - Remove custom error formats like {error, details} structure
  - [x] 11.4 Fix HTTP 204 response to not include JSON body
    - Update DeleteEpic method to return proper 204 No Content without response body
    - Ensure compliance with HTTP standards
  - [x] 11.5 Align assignment request schemas
    - Decide between AssignmentRequest (nullable assignee_id) vs AssignEpicRequest (required assignee_id)
    - Update either OpenAPI schema or service implementation for consistency
  - [x] 11.6 Implement or document missing epic endpoints
    - Either implement missing endpoints (user-stories creation, deletion validation, comments) in epic handler
    - Or remove undocumented endpoints from OpenAPI specification
    - Ensure 1:1 mapping between documented and implemented endpoints
  - _Requirements: 2.1, 2.2, 2.3, 6.1, 6.2, 6.3, 6.4_

- [x] 12. Create validation tests for documentation accuracy
  - Write tests to verify OpenAPI spec matches actual route implementations
  - Create schema validation tests for response formats
  - Add tests to ensure authentication requirements are properly documented
  - Implement automated checks for documentation completeness
  - _Requirements: 6.1, 6.2, 6.3, 6.4_