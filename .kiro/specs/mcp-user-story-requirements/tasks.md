# Implementation Plan

- [x] 1. Update UserStoryHandler to support get_user_story_requirements tool
  - Add RequirementService dependency to UserStoryHandler struct
  - Update NewUserStoryHandler constructor to accept RequirementService parameter
  - Add "get_user_story_requirements" to GetSupportedTools() method
  - Add GetRequirements case to HandleTool() switch statement
  - _Requirements: REQ-048, REQ-049, REQ-054_

- [x] 2. Implement GetRequirements method in UserStoryHandler
  - Validate required "user_story" argument using existing validateRequiredArgs helper
  - Parse user story ID using existing parseUUIDOrReferenceID helper
  - Handle user story not found errors with proper jsonrpc error responses
  - Call getRequirementsWithRelatedData helper method
  - Format response using formatRequirementsMessage helper
  - Return response using types.CreateSuccessResponse
  - _Requirements: REQ-048, REQ-049, REQ-050, REQ-051_

- [x] 3. Implement getRequirementsWithRelatedData helper method
  - Create RequirementFilters with UserStoryID filter
  - Set Include parameter to "type,creator,assignee" for preloading
  - Set OrderBy to "priority ASC, created_at DESC" for proper sorting
  - Call requirementService.ListRequirements with filters
  - Handle service errors appropriately
  - _Requirements: REQ-050, REQ-053, REQ-056_

- [x] 4. Implement formatRequirementsMessage helper method
  - Handle empty requirements case with appropriate message
  - Format requirements count and user story reference in header
  - Format each requirement with reference_id, title, priority, status, type_name, creator_username
  - Include assignee_username if present
  - Include description if not empty
  - Include created_at timestamp in RFC3339 format
  - _Requirements: REQ-050, REQ-055_

- [x] 5. Update Handler constructor to pass RequirementService
  - Modify NewHandler function in internal/mcp/tools/handler.go
  - Add requirementService parameter to function signature
  - Pass requirementService to NewUserStoryHandler call
  - Update all callers of NewHandler to provide RequirementService
  - _Requirements: REQ-048_

- [ ]* 6. Write unit tests for GetRequirements method
  - Test successful requirements retrieval with valid user story reference
  - Test empty requirements case
  - Test invalid user story reference format
  - Test user story not found scenario
  - Test service error handling
  - Mock RequirementService and UserStoryService dependencies
  - _Requirements: REQ-048, REQ-049, REQ-050, REQ-051_

- [ ]* 7. Write integration tests with PostgreSQL
  - Test complete workflow with real database
  - Test requirements sorting by priority and created_at
  - Test preloading of related entities (type, creator, assignee)
  - Test with various user story and requirement combinations
  - _Requirements: REQ-053, REQ-056_

- [ ]* 8. Write end-to-end MCP protocol tests
  - Test complete MCP tools/call request-response cycle
  - Test JSON-RPC 2.0 protocol compliance
  - Test tool discovery through tools/list
  - Test error responses format compliance
  - _Requirements: REQ-054, REQ-055_