# Implementation Plan

- [x] 1. Set up database schema and migrations
  - Create migration file for personal_access_tokens table with proper indexes and constraints
  - Add foreign key relationship to users table with CASCADE delete
  - Include JSONB scopes field with default full_access value
  - _Requirements: 1.6, 4.6, 5.2_

- [x] 2. Implement core PAT data model and repository
- [x] 2.1 Create PersonalAccessToken GORM model
  - Define struct with proper GORM tags and JSON serialization rules
  - Implement BeforeCreate hook for any initialization logic
  - Add association with User model
  - _Requirements: 1.6, 4.6_

- [x] 2.2 Implement PAT repository interface and PostgreSQL implementation
  - Create repository interface with CRUD operations and authentication queries
  - Implement PostgreSQL-specific repository with proper error handling
  - Add methods for token lookup by prefix and user management
  - _Requirements: 2.2, 2.3, 3.1, 3.4_

- [x] 2.3 Write unit tests for PAT model and repository
  - Test model validation and serialization rules
  - Test repository CRUD operations with SQLite
  - Test error handling and edge cases
  - _Requirements: 1.6, 2.3, 3.4_

- [x] 3. Implement secure token generation and validation
- [x] 3.1 Create token generator service
  - Implement SecureTokenGenerator using crypto/rand
  - Generate tokens with mcp_pat_ prefix and base64url encoding
  - Ensure cryptographically secure random generation
  - _Requirements: 1.1, 5.1, 5.3_

- [x] 3.2 Implement bcrypt hash service for token storage
  - Create hash service using golang.org/x/crypto/bcrypt
  - Implement secure hashing with appropriate cost factor
  - Add constant-time token comparison methods
  - _Requirements: 1.6, 2.4, 5.2, 5.7_

- [x] 3.3 Write unit tests for token generation and hashing
  - Test token format and entropy requirements
  - Test bcrypt hashing and validation
  - Test security properties and edge cases
  - _Requirements: 1.1, 1.6, 5.1, 5.2_

- [x] 4. Implement PAT service layer
- [x] 4.1 Create PAT service interface and implementation
  - Implement CreatePAT method with validation and token generation
  - Add ListUserPATs method with pagination support
  - Implement RevokePAT method with proper authorization
  - _Requirements: 1.1, 1.2, 1.3, 3.1, 3.3, 3.4_

- [x] 4.2 Implement token validation service for authentication
  - Add ValidateToken method with prefix extraction and hash comparison
  - Implement UpdateLastUsed method for usage tracking
  - Add CleanupExpiredTokens method for maintenance
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.6, 3.5_

- [x] 4.3 Write unit tests for PAT service layer
  - Test token creation workflow with mocked dependencies
  - Test authentication flow and validation logic
  - Test error handling and security edge cases
  - _Requirements: 1.7, 2.5, 2.7, 3.7_

- [x] 5. Implement HTTP API endpoints
- [x] 5.1 Create PAT HTTP handlers
  - Implement POST /api/v1/pats endpoint for token creation
  - Add GET /api/v1/pats endpoint for listing user tokens
  - Implement DELETE /api/v1/pats/{id} endpoint for token revocation
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.7_

- [x] 5.2 Add request/response models and validation
  - Create CreatePATRequest struct with validation tags
  - Implement PATCreateResponse with token field (returned only once)
  - Add proper error response structures
  - _Requirements: 1.2, 1.5, 1.7, 4.4, 4.7_

- [x] 5.3 Write unit tests for PAT handlers
  - Test HTTP request/response handling
  - Test validation and error cases
  - Test authentication requirements for endpoints
  - _Requirements: 4.6, 4.7_

- [-] 6. Implement PAT authentication middleware
- [x] 6.1 Create PAT authentication middleware
  - Extract Bearer tokens from Authorization header
  - Detect PAT tokens by mcp_pat_ prefix
  - Integrate with existing JWT authentication as fallback
  - _Requirements: 2.1, 2.2, 2.3, 2.5_

- [ ] 6.2 Integrate middleware with existing authentication system
  - Add PAT middleware to authentication chain
  - Ensure compatibility with existing JWT authentication
  - Set proper user context for authenticated requests
  - _Requirements: 2.4, 2.6, 4.6_

- [ ] 6.3 Write integration tests for authentication middleware
  - Test PAT authentication flow end-to-end
  - Test fallback to JWT authentication
  - Test error handling and security scenarios
  - _Requirements: 2.5, 2.7_

- [ ] 7. Add PAT routes and integrate with server
- [-] 7.1 Register PAT routes in server configuration
  - Add PAT routes to existing route groups
  - Ensure proper middleware chain for authentication
  - Configure rate limiting for PAT endpoints
  - _Requirements: 4.1, 4.2, 4.3, 4.6, 5.6_

- [ ] 7.2 Update server startup to include PAT services
  - Initialize PAT service and repository in dependency injection
  - Add PAT cleanup job for expired tokens
  - Configure logging for PAT operations
  - _Requirements: 3.5, 5.5, 5.7_

- [ ] 7.3 Write integration tests for complete PAT workflow
  - Test end-to-end token creation and usage
  - Test token management operations
  - Test authentication with real HTTP requests
  - _Requirements: 1.1, 2.1, 3.1, 4.1_

- [ ] 8. Add security enhancements and monitoring
- [ ] 8.1 Implement rate limiting for PAT endpoints
  - Add rate limiting middleware for token creation
  - Implement authentication attempt rate limiting
  - Configure appropriate limits and error responses
  - _Requirements: 5.6_

- [ ] 8.2 Add security logging and monitoring
  - Log PAT creation, usage, and revocation events
  - Implement security event logging without exposing tokens
  - Add metrics for PAT usage and authentication attempts
  - _Requirements: 2.7, 5.5, 5.7_

- [ ] 8.3 Write security tests for PAT implementation
  - Test rate limiting effectiveness
  - Test security logging without token exposure
  - Test timing attack resistance
  - _Requirements: 5.5, 5.6, 5.7_