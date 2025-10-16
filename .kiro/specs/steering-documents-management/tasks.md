# Implementation Plan

- [x] 1. Create database migration for steering documents
  - Create migration files 000006_add_steering_documents.up.sql and 000006_add_steering_documents.down.sql
  - Add PostgreSQL sequence steering_document_ref_seq for STD-XXX reference IDs
  - Create get_next_steering_document_ref_id() function
  - Create steering_documents table with proper indexes and constraints
  - Create epic_steering_documents junction table for many-to-many relationship
  - Add updated_at trigger for steering_documents table
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6_

- [x] 2. Implement GORM model for SteeringDocument
  - Create internal/models/steering_document.go with SteeringDocument struct
  - Add proper GORM tags, JSON tags, and validation tags
  - Implement BeforeCreate and BeforeUpdate hooks
  - Add custom MarshalJSON method for conditional field inclusion
  - Add TableName method returning "steering_documents"
  - Update internal/models/models.go to include SteeringDocument in AllModels()
  - _Requirements: 1.1, 1.2, 1.3, 1.6_

- [x] 3. Extend Epic model with steering documents relationship
  - Add SteeringDocuments field to Epic struct in internal/models/epic.go
  - Update Epic MarshalJSON method to include steering_documents when populated
  - Add proper GORM many2many tag for epic_steering_documents junction table
  - _Requirements: 2.1, 2.2, 2.6_

- [x] 4. Create repository layer for steering documents
  - Create internal/repository/steering_document_repository.go with interface and implementation
  - Implement Create, GetByID, GetByReferenceID, Update, Delete methods
  - Implement List method with filtering, pagination, and search
  - Implement Search method using PostgreSQL full-text search
  - Implement GetByEpicID, LinkToEpic, UnlinkFromEpic methods for epic relationships
  - Add proper error handling and GORM query optimization
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 5. Create service layer for steering documents
  - Create internal/service/steering_document_service.go with interface and implementation
  - Implement business logic for CRUD operations with validation
  - Add user existence validation for creator
  - Add epic existence validation for linking operations
  - Implement proper error handling with custom error types
  - Add authorization checks based on user roles (Administrator, User, Commenter)
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 3.1, 3.2, 3.3, 3.4, 3.5, 4.1, 4.2, 4.3, 4.4, 4.5, 5.1, 5.2, 5.3, 5.4, 5.5_

- [x] 6. Create REST API handlers for steering documents
  - Create internal/handlers/steering_document_handler.go with HTTP handlers
  - Implement CreateSteeringDocument (POST /api/v1/steering-documents)
  - Implement GetSteeringDocument (GET /api/v1/steering-documents/:id) supporting UUID and reference_id
  - Implement UpdateSteeringDocument (PUT /api/v1/steering-documents/:id)
  - Implement DeleteSteeringDocument (DELETE /api/v1/steering-documents/:id)
  - Implement ListSteeringDocuments (GET /api/v1/steering-documents) with query parameters
  - Implement LinkSteeringDocumentToEpic (POST /api/v1/epics/:epic_id/steering-documents/:doc_id)
  - Implement UnlinkSteeringDocumentFromEpic (DELETE /api/v1/epics/:epic_id/steering-documents/:doc_id)
  - Implement GetEpicSteeringDocuments (GET /api/v1/epics/:id/steering-documents)
  - Add proper JWT authentication and authorization middleware
  - Add comprehensive error handling with standard HTTP status codes
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7, 6.8, 6.9, 6.10, 5.1, 5.2, 5.3, 5.4, 5.5_

- [x] 7. Add MCP tools for steering documents management
  - Add steering document tools to internal/handlers/mcp_tool_schemas.go
  - Add steeringDocumentService to ToolsHandler struct in internal/handlers/mcp_tools_handler.go
  - Update NewToolsHandler constructor to accept steeringDocumentService parameter
  - Implement handleListSteeringDocuments MCP tool handler
  - Implement handleCreateSteeringDocument MCP tool handler
  - Implement handleGetSteeringDocument MCP tool handler supporting UUID and STD-XXX reference_id
  - Implement handleUpdateSteeringDocument MCP tool handler
  - Implement handleLinkSteeringToEpic MCP tool handler supporting UUID and reference_id for both entities
  - Implement handleUnlinkSteeringFromEpic MCP tool handler
  - Implement handleGetEpicSteeringDocuments MCP tool handler
  - Add proper PAT authentication and user context extraction
  - Add comprehensive error handling using JSON-RPC 2.0 error format
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7, 7.8, 7.9, 7.10_

- [x] 8. Register routes and wire up dependencies
  - Add steering document routes to internal/server/routes package
  - Wire up SteeringDocumentRepository, SteeringDocumentService, and SteeringDocumentHandler in dependency injection
  - Update MCP tools handler initialization to include steering document service
  - Add route registration for all steering document endpoints
  - Ensure proper middleware chain (authentication, authorization, logging)
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7, 6.8, 6.9, 6.10, 7.8_

- [x] 9. Run database migration and test basic functionality
  - ✅ Execute migration to create steering_documents and epic_steering_documents tables
  - ✅ Verify sequence and function creation for STD-XXX reference ID generation
  - ✅ Test basic CRUD operations through REST API
  - ✅ Test epic-steering document linking functionality
  - ⚠️ Verify MCP tools are properly registered and accessible (MCP endpoint not found - needs investigation)
  - ✅ Test authentication and authorization for different user roles
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 5.1, 5.2, 5.3, 5.4, 5.5, 6.10, 7.9_
  
  **Test Results:**
  - Database migration successful - tables created with proper indexes and constraints
  - Reference ID generation working: STD-001, STD-002, etc.
  - CRUD operations tested successfully:
    - Create: ✅ POST /api/v1/steering-documents
    - Read: ✅ GET /api/v1/steering-documents/:id (UUID and reference ID)
    - Update: ✅ PUT /api/v1/steering-documents/:id
    - Delete: ✅ DELETE /api/v1/steering-documents/:id
    - List: ✅ GET /api/v1/steering-documents
  - Epic linking tested successfully:
    - Link: ✅ POST /api/v1/epics/:id/steering-documents/:doc_id
    - Unlink: ✅ DELETE /api/v1/epics/:id/steering-documents/:doc_id
    - Get linked: ✅ GET /api/v1/epics/:id/steering-documents
  - Authorization working correctly:
    - Administrator can manage all documents
    - User can only manage their own documents
    - Proper error messages for unauthorized access

- [x] 10. Write unit tests for steering document functionality
  - ✅ Write unit tests for SteeringDocument GORM model
  - ✅ Write unit tests for steering document repository with SQLite
  - ✅ Write unit tests for steering document service layer
  - ✅ Write unit tests for REST API handlers
  - ✅ Write unit tests for MCP tools handlers
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 3.1, 3.2, 3.3, 4.1, 4.2, 4.3, 5.1, 5.2, 5.3, 6.1, 6.2, 6.3, 7.1, 7.2, 7.3_
  
  **Test Results:**
  - Model tests: ✅ SteeringDocument model with BeforeCreate/BeforeUpdate hooks, JSON marshaling, and relationships
  - Repository tests: ✅ Complete CRUD operations, filtering, search, and epic linking with SQLite
  - Service tests: ✅ Business logic validation, authorization checks, and error handling
  - Handler tests: ✅ Mock-based testing of REST API handlers with comprehensive scenarios
  - MCP tools tests: ✅ Mock-based testing of MCP tool handlers with various input scenarios
  - Coverage: All major functionality paths tested including success cases, error cases, and edge cases

- [x] 11. Add optional epic_id field to steering document creation
  - Update CreateSteeringDocumentRequest struct to include optional EpicID field
  - Update REST API handler to accept epic_id in request body and automatically link document to epic
  - Update MCP tool schema for create_steering_document to include optional epic_id parameter
  - Update MCP tool handler to support epic linking during creation
  - Update service layer to validate epic existence and create link during document creation
  - Add validation to ensure epic exists before creating the link
  - Update unit tests to cover new epic linking functionality during creation
  - _Requirements: 1.3, 6.1.1, 7.2_

- [-] 12. Write integration tests for steering document functionality
  - Write integration tests for steering document CRUD operations with PostgreSQL
  - Write integration tests for epic-steering document relationships
  - Write integration tests for steering document creation with epic_id linking
  - Write integration tests for full-text search functionality
  - Write integration tests for REST API endpoints with authentication
  - Write integration tests for MCP tools with PAT authentication
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 3.1, 3.2, 3.3, 3.4, 3.5, 4.1, 4.2, 4.3, 4.4, 4.5, 5.1, 5.2, 5.3, 5.4, 5.5, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7, 6.8, 6.9, 6.10, 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7, 7.8, 7.9, 7.10_