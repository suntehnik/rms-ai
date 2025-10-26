# Requirements Document

## Introduction

This document outlines the requirements for implementing a JSON-RPC API handler in the Product Requirements Management System. The API handler will process Model Context Protocol (MCP) requests from MCP servers, providing access to requirements management capabilities through a standardized JSON-RPC 2.0 interface. The system will use existing PAT authentication infrastructure for secure access and integrate with the current service layer architecture.

## Requirements

### Requirement 1: JSON-RPC API Endpoint Implementation

**User Story:** As a Backend API, I want to provide a JSON-RPC endpoint that processes MCP requests, so that external MCP servers can access requirements management capabilities.

#### Acceptance Criteria

1. WHEN implementing the handler THEN the system SHALL provide POST /api/v1/mcp endpoint
2. WHEN receiving requests THEN the endpoint SHALL require Authorization header with Bearer PAT
3. WHEN processing requests THEN the system SHALL validate PAT using existing authentication infrastructure
4. WHEN PAT is invalid THEN the endpoint SHALL return 401 Unauthorized
5. WHEN PAT is valid THEN the system SHALL identify the user and execute operations on their behalf
6. WHEN request body contains "method" field THEN the system SHALL route to appropriate handler
7. WHEN handling "resources/read" THEN the system SHALL fetch entity by URI and return data
8. WHEN handling "tools/call" THEN the system SHALL execute business logic and return results
9. WHEN handling "prompts/get" THEN the system SHALL gather context and return formatted prompt
10. WHEN operations fail THEN the system SHALL return structured error responses
11. IF request is malformed THEN the system SHALL return 400 Bad Request

### Requirement 2: JSON-RPC Request Routing

**User Story:** As a Backend API, I want to route JSON-RPC requests to appropriate handlers, so that different types of MCP operations are processed correctly.

#### Acceptance Criteria

1. WHEN receiving request with method "resources/read" THEN the system SHALL route to resource handler
2. WHEN receiving request with method "tools/call" THEN the system SHALL route to tool handler
3. WHEN receiving request with method "prompts/get" THEN the system SHALL route to prompt handler
4. WHEN routing to handlers THEN the system SHALL pass authenticated user context
5. WHEN handlers execute THEN they SHALL use existing service layer (EpicService, UserStoryService, etc.)
6. WHEN mapping resources THEN the system SHALL parse URI scheme (epic://, user-story://, etc.)
7. WHEN mapping tools THEN the system SHALL map tool names to service method calls
8. WHEN mapping prompts THEN the system SHALL load required entities and format prompt text
9. IF method is not recognized THEN the system SHALL return error with unsupported method
10. IF routing fails THEN the system SHALL log error and return internal server error

### Requirement 3: JSON-RPC 2.0 Message Handling

**User Story:** As a JSON-RPC API handler, I want to handle JSON-RPC 2.0 messages correctly, so that I maintain protocol compliance and interoperability.

#### Acceptance Criteria

1. WHEN parsing messages THEN the API handler SHALL validate presence of "jsonrpc": "2.0"
2. WHEN processing requests THEN the system SHALL validate presence of "id", "method", and "params" fields
3. WHEN processing notifications THEN the system SHALL handle messages without "id" field
4. WHEN sending responses THEN the system SHALL include "jsonrpc": "2.0" and matching "id"
5. WHEN operations succeed THEN the system SHALL return response with "result" field
6. WHEN operations fail THEN the system SHALL return response with "error" field containing code and message
7. WHEN implementing error codes THEN the system SHALL use standard JSON-RPC codes (-32700 parse error, -32600 invalid request, -32601 method not found, -32602 invalid params)
8. WHEN implementing custom errors THEN the system SHALL use codes starting from -32000
9. IF message is not valid JSON-RPC THEN the system SHALL return parse error

### Requirement 4: Authentication and Authorization

**User Story:** As a system administrator, I want JSON-RPC requests to be properly authenticated and authorized, so that security is maintained.

#### Acceptance Criteria

1. WHEN receiving requests THEN the API handler SHALL require PAT in Authorization header
2. WHEN validating PAT THEN the system SHALL use existing PAT authentication middleware
3. WHEN operations execute THEN they SHALL run with permissions of the PAT owner
4. WHEN user lacks permissions THEN the system SHALL return 403 Forbidden
5. WHEN audit logging THEN the system SHALL log all JSON-RPC operations with user attribution
6. WHEN PAT is expired THEN the system SHALL reject request and return authentication error
7. WHEN rate limiting THEN the system SHALL apply existing rate limits to JSON-RPC endpoints
8. IF authentication fails THEN the system SHALL log attempt for security monitoring

### Requirement 5: Error Handling and Logging

**User Story:** As a developer, I want comprehensive error handling and logging for the JSON-RPC API, so that I can debug issues and monitor system health.

#### Acceptance Criteria

1. WHEN errors occur THEN the API handler SHALL log them with appropriate severity levels
2. WHEN logging THEN the system SHALL include timestamp, request ID, and error context
3. WHEN PAT tokens appear in logs THEN the system SHALL redact them for security
4. WHEN service layer calls fail THEN the system SHALL map errors to appropriate JSON-RPC error codes
5. WHEN validation fails THEN errors SHALL include specific field information
6. WHEN errors are user-facing THEN messages SHALL be clear and actionable
7. WHEN internal errors occur THEN the system SHALL log full stack traces
8. WHEN JSON-RPC parsing fails THEN the system SHALL return appropriate parse error
9. IF database operations fail THEN the system SHALL return internal error with correlation ID

### Requirement 6: Resources Handler Implementation

**User Story:** As an MCP client, I want to access contextual data about epics, user stories, and requirements through resources, so that I can provide informed assistance to users.

#### Acceptance Criteria

1. WHEN handling "resources/read" method THEN the system SHALL parse URI scheme (epic://, user-story://, etc.)
2. WHEN URI contains entity ID THEN the system SHALL fetch entity using appropriate service
3. WHEN entity exists THEN the system SHALL return formatted resource content
4. WHEN entity not found THEN the system SHALL return JSON-RPC error with code -32002
5. WHEN supporting hierarchies THEN the system SHALL handle URIs like "epic://{id}/hierarchy"
6. WHEN formatting response THEN the system SHALL include uri, name, description, and mimeType
7. WHEN resource access fails THEN the system SHALL return appropriate error code
8. IF user lacks read permissions THEN the system SHALL return 403 Forbidden

### Requirement 7: Tools Handler Implementation

**User Story:** As an MCP client, I want to execute tools for creating, updating, and managing requirements artifacts, so that I can assist users with content management.

#### Acceptance Criteria

1. WHEN handling "tools/call" method THEN the system SHALL validate tool name exists
2. WHEN tool arguments provided THEN the system SHALL validate against expected schema
3. WHEN executing CRUD tools THEN the system SHALL use appropriate service methods
4. WHEN tool execution succeeds THEN the system SHALL return structured result with content array
5. WHEN supporting tools THEN the system SHALL provide "create_epic", "update_epic", "create_user_story", "update_user_story"
6. WHEN supporting tools THEN the system SHALL provide "create_requirement", "update_requirement", "create_relationship"
7. WHEN supporting search THEN the system SHALL provide "search_global" and "search_requirements" tools
8. WHEN validation fails THEN the system SHALL return JSON-RPC error with details
9. IF user lacks permissions THEN the system SHALL return 403 Forbidden

### Requirement 8: System Prompts Management with Database Storage

**User Story:** As an administrator, I want to manage system prompts stored in database with CRUD operations, so that I can configure AI assistant behavior and ensure only one prompt is active at any time.

#### Acceptance Criteria

1. WHEN system stores prompts THEN they SHALL be persisted in database with reference IDs (PROMPT-001, PROMPT-002, etc.)
2. WHEN managing prompts THEN only one prompt SHALL be marked as active at any time
3. WHEN MCP client initializes THEN the system SHALL automatically provide active system prompt
4. WHEN handling "resources/read" with URI "requirements://prompts" THEN the system SHALL return collection of all prompts
5. WHEN handling "resources/read" with URI "requirements://prompts/active" THEN the system SHALL return currently active prompt
6. WHEN handling "prompts/list" method THEN the system SHALL return array of available prompt descriptors from database
7. WHEN handling "prompts/get" method THEN the system SHALL validate prompt name exists in database and return full definition
8. WHEN administrator creates prompt THEN the system SHALL provide CRUD operations through both REST API and MCP tools
9. WHEN administrator updates prompt THEN the system SHALL validate prompt content structure
10. WHEN administrator activates prompt THEN the system SHALL deactivate current active prompt and activate selected one
11. WHEN formatting prompt response THEN the system SHALL include name, description, and content as simple text
12. IF user lacks administrator permissions for CRUD operations THEN the system SHALL return 403 Forbidden

### Requirement 9: REST API for System Prompts Management

**User Story:** As an administrator, I want to manage system prompts through REST API endpoints, so that I can perform CRUD operations and activate prompts through web interface.

#### Acceptance Criteria

1. WHEN implementing REST API THEN the system SHALL provide POST /api/v1/prompts endpoint for creating prompts
2. WHEN implementing REST API THEN the system SHALL provide GET /api/v1/prompts endpoint for listing prompts with pagination
3. WHEN implementing REST API THEN the system SHALL provide GET /api/v1/prompts/:id endpoint for retrieving prompt by ID or reference ID
4. WHEN implementing REST API THEN the system SHALL provide PUT /api/v1/prompts/:id endpoint for updating prompts
5. WHEN implementing REST API THEN the system SHALL provide DELETE /api/v1/prompts/:id endpoint for deleting prompts
6. WHEN implementing REST API THEN the system SHALL provide PATCH /api/v1/prompts/:id/activate endpoint for activating prompts
7. WHEN implementing REST API THEN the system SHALL provide GET /api/v1/prompts/active endpoint for retrieving active prompt
8. WHEN REST API operations execute THEN they SHALL require Administrator role for all operations except GET
9. WHEN activating prompt through REST API THEN the system SHALL ensure only one prompt is active
10. WHEN validating prompt data THEN the system SHALL validate content field as plain text
11. IF non-administrator attempts CRUD operations THEN the system SHALL return 403 Forbidden
12. IF prompt not found THEN the system SHALL return 404 Not Found

### Requirement 10: MCP Tools for System Prompts Management

**User Story:** As an MCP client with administrator privileges, I want to manage system prompts through MCP tools, so that I can perform CRUD operations and activate prompts through MCP protocol.

#### Acceptance Criteria

1. WHEN supporting MCP tools THEN the system SHALL provide "create_prompt" tool for creating new prompts
2. WHEN supporting MCP tools THEN the system SHALL provide "update_prompt" tool for updating existing prompts
3. WHEN supporting MCP tools THEN the system SHALL provide "delete_prompt" tool for deleting prompts
4. WHEN supporting MCP tools THEN the system SHALL provide "activate_prompt" tool for activating prompts
5. WHEN supporting MCP tools THEN the system SHALL provide "list_prompts" tool for listing all prompts
6. WHEN supporting MCP tools THEN the system SHALL provide "get_active_prompt" tool for retrieving active prompt
7. WHEN MCP tools execute THEN they SHALL require Administrator role for CRUD operations
8. WHEN validating tool arguments THEN the system SHALL validate prompt content as plain text
9. WHEN activating prompt through MCP tool THEN the system SHALL ensure only one prompt is active
10. WHEN tool execution succeeds THEN the system SHALL return structured result with content array
11. IF non-administrator attempts CRUD tools THEN the system SHALL return JSON-RPC error with code -32002
12. IF prompt not found THEN the system SHALL return JSON-RPC error with appropriate message
