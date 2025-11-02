# Implementation Plan

Convert the feature design into a series of prompts for a code-generation LLM that will implement each step with incremental progress. Make sure that each prompt builds on the previous prompts, and ends with wiring things together. There should be no hanging or orphaned code that isn't integrated into a previous step. Focus ONLY on tasks that involve writing, modifying, or testing code.

## Task List

- [x] 1. Extend UserStoryService with reference ID resolution method
  - Add `GetUUIDByReferenceID(referenceID string) (uuid.UUID, error)` method to UserStoryService interface
  - Implement the method in userStoryService struct with Redis caching support
  - Add corresponding repository method `GetUUIDByReferenceID` for efficient UUID-only queries
  - _Requirements: REQ-036 (user story identifier resolution)_

- [x] 2. Create AcceptanceCriteriaHandler following existing MCP tool patterns
  - [x] 2.1 Create acceptance criteria handler structure
    - Implement `AcceptanceCriteriaHandler` struct following existing handler patterns (UserStoryHandler, EpicHandler)
    - Define `create_acceptance_criteria` tool with JSON schema validation
    - Implement `GetSupportedTools()` and `HandleTool()` methods per ToolHandler interface
    - _Requirements: REQ-036, REQ-037_

  - [x] 2.2 Implement tool execution logic with validation
    - Parse and validate JSON-RPC request parameters (user_story_id, description)
    - Implement user story identifier resolution (UUID or US-XXX reference ID format)
    - Validate description field constraints (required, max 50000 characters)
    - Extract authenticated user ID from context using existing auth patterns
    - _Requirements: REQ-036, REQ-037_

  - [x] 2.3 Implement acceptance criteria creation and response formatting
    - Create acceptance criteria using existing AcceptanceCriteriaService
    - Format structured JSON-RPC response with all acceptance criteria fields
    - Map service errors to appropriate JSON-RPC error codes (-32602 for validation, -32001 for auth)
    - Return success response with auto-generated reference ID and metadata
    - _Requirements: REQ-036, REQ-037_

- [x] 3. Integrate AcceptanceCriteriaHandler with MCP infrastructure
  - [x] 3.1 Register handler in MCP tools handler
    - Add AcceptanceCriteriaHandler to main Handler struct in `internal/mcp/tools/handler.go`
    - Update NewHandler constructor to initialize AcceptanceCriteriaHandler
    - Register acceptance criteria tools in tool routing map
    - Update GetAllSupportedTools to include acceptance criteria tools
    - _Requirements: REQ-036, REQ-037_

  - [x] 3.2 Update MCP server capabilities
    - Ensure acceptance criteria tools are discoverable via tools/list method
    - Verify tool execution works via tools/call method
    - Test integration with existing MCP JSON-RPC infrastructure
    - _Requirements: REQ-036, REQ-037_

- [ ] 4. Write comprehensive tests for acceptance criteria MCP tool
  - [x] 4.1 Write unit tests for AcceptanceCriteriaHandler
    - Test user story identifier validation (UUID and US-XXX reference ID formats)
    - Test description field validation (required field, max 50000 characters)
    - Test error handling for invalid user story identifiers and validation failures
    - Test authentication context extraction and error scenarios
    - _Requirements: REQ-036, REQ-037_

  - [ ]* 4.2 Write integration tests for MCP tool execution
    - Test complete JSON-RPC request/response cycle with real database
    - Test authentication integration with existing PAT middleware
    - Test database persistence and reference ID generation
    - Test error scenarios with proper JSON-RPC error code mapping
    - _Requirements: REQ-036, REQ-037_

  - [ ] 4.3 Write MCP protocol compliance tests
    - Test tool discovery via tools/list endpoint returns acceptance criteria tools
    - Test tool execution via tools/call endpoint with valid and invalid parameters
    - Test JSON-RPC 2.0 format compliance for requests and responses
    - Verify integration with existing MCP server infrastructure
    - _Requirements: REQ-036, REQ-037_