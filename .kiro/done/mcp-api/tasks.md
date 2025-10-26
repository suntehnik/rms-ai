# Implementation Plan

- [x] 1. Set up JSON-RPC infrastructure and core components
  - Create JSON-RPC message structures and validation
  - Implement JSON-RPC 2.0 compliant request/response handling
  - Set up error code mapping and standardized error responses
  - _Requirements: 3.1, 3.2, 3.4, 3.5, 3.6, 3.7, 3.8, 3.9_

- [x] 2. Implement URI parser for reference ID handling
  - Create URI parser with reference ID validation
  - Implement scheme-to-prefix validation (epic->EP, user-story->US, etc.)
  - Add support for sub-paths (hierarchy, requirements, relationships)
  - _Requirements: 6.1, 6.8_

- [x] 3. Create MCP endpoint and request routing infrastructure
  - Implement POST /api/v1/mcp endpoint with PAT authentication
  - Set up method routing to appropriate handlers
  - Integrate with existing middleware (authentication, logging, CORS)
  - _Requirements: 1.1, 1.2, 1.3, 1.6, 2.1, 2.2, 2.3, 2.9, 2.10_

- [x] 4. Implement Resources handler for entity data access
  - Create ResourceHandler with service layer integration
  - Implement resources/read method with reference ID resolution
  - Support entity retrieval for epics, user stories, requirements, acceptance criteria
  - Handle sub-paths for hierarchical data (epic hierarchy, user story requirements, etc.)
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7, 6.8_

- [x] 5. Implement Tools handler for CRUD operations
  - Create ToolsHandler with comprehensive tool definitions
  - Implement CRUD tools: create_epic, update_epic, create_user_story, update_user_story
  - Implement CRUD tools: create_requirement, update_requirement, create_relationship
  - Add JSON Schema validation for tool arguments
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.8, 7.9_

- [x] 6. Implement search and analytics tools
  - Add search_global tool for cross-entity search
  - Add search_requirements tool with filtering capabilities
  - Integrate with existing SearchService for backend queries
  - Format search results as structured tool responses
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.7, 7.8, 7.9_

- [x] 7. Implement System Prompts management with database storage
  - Create database migration for prompts table with reference IDs (PROMPT-001, etc.)
  - Create Prompt model in internal/models with GORM annotations and reference ID generation
  - Create PromptService in internal/service with CRUD operations and active prompt management
  - Implement REST API endpoints for prompts management (/api/v1/prompts/*)
  - Create PromptsHandler for MCP with prompts/list and prompts/get methods
  - Add MCP tools for prompt management (create_prompt, update_prompt, activate_prompt, etc.)
  - Add resources/read support for "requirements://prompts" and "requirements://prompts/active"
  - Ensure only one prompt can be active at any time with database constraints
  - Add administrator role validation for CRUD operations
  - _Requirements: 8.1-8.12, 9.1-9.12, 10.1-10.12_

- [x] 8. Add comprehensive error handling and logging
  - Implement error mapping from service layer to JSON-RPC error codes
  - Add structured logging with request correlation IDs
  - Implement PAT token redaction in logs for security
  - Add audit logging for all MCP operations with user attribution
  - _Requirements: 4.4, 4.5, 4.6, 4.7, 4.8, 5.1, 5.2, 5.3, 5.8, 5.9_

- [x] 9. Create comprehensive test suite
  - Write unit tests for JSON-RPC message handling and validation
  - Write unit tests for URI parser with reference ID validation
  - Write unit tests for all handlers (Resources, Tools, Prompts)
  - Write integration tests for full request flow with PAT authentication
  - _Requirements: All requirements need test coverage_

- [x] 9.1 Write unit tests for JSON-RPC infrastructure
  - Test JSON-RPC 2.0 message parsing and validation
  - Test error code mapping and response formatting
  - Test request routing to appropriate handlers
  - _Requirements: 3.1, 3.2, 3.4, 3.5, 3.6, 3.7, 3.8, 3.9_

- [x] 9.2 Write unit tests for URI parser
  - Test reference ID validation and scheme matching
  - Test sub-path parsing for hierarchical resources
  - Test error cases for invalid URI formats
  - _Requirements: 6.1, 6.8_

- [ ]* 9.3 Write unit tests for handler implementations
  - Test ResourceHandler entity retrieval and formatting including prompts resources
  - Test ToolsHandler CRUD operations and search functionality including prompt management tools
  - Test PromptsHandler prompts/list method returning prompt descriptors from database
  - Test PromptsHandler prompts/get method with database lookup and validation
  - Test PromptService CRUD operations and active prompt management
  - Test REST API endpoints for prompts with administrator role validation
  - Test MCP tools for prompt management with proper authorization
  - Test database constraints ensuring only one active prompt
  - _Requirements: 6.1-6.8, 7.1-7.9, 8.1-8.12, 9.1-9.12, 10.1-10.12_

- [x] 9.4 Write integration tests for full request flow
  - Test complete flow from HTTP request to service layer
  - Test PAT authentication and authorization
  - Test error handling and logging integration
  - _Requirements: 1.1-1.11, 2.1-2.10, 4.1-4.8, 5.1-5.9_

- [ ]* 10. Add performance optimizations and monitoring
  - Implement response caching for read-only operations
  - Add request/response time monitoring by method type
  - Add error rate tracking by JSON-RPC error code
  - Optimize concurrent request handling
  - _Requirements: 5.1, 5.2, 5.3, 5.9_

- [x] 11. Create documentation and examples
  - Document all supported URI schemes and patterns
  - Document all available tools with JSON Schema examples
  - Document all available prompts with usage scenarios
  - Create troubleshooting guide for common errors
  - _Requirements: All requirements need documentation_

- [ ]* 12. Integration testing and deployment preparation
  - Test with existing PAT authentication infrastructure
  - Verify integration with existing service layer methods
  - Test error handling with existing middleware stack
  - Validate logging integration with existing audit systems
  - _Requirements: 1.2, 1.3, 1.4, 1.5, 2.4, 2.5, 4.1, 4.2, 4.3, 4.8, 5.8_