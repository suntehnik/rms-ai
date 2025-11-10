# Implementation Plan: Auth Session Lifecycle

## Task Overview

This implementation plan breaks down the auth session lifecycle feature into discrete, manageable coding tasks. Each task builds incrementally on previous tasks and references specific requirements from the requirements document.

---

## 1. Database Schema and Migration

 - [x] 1. Database Schema and Migration
  - [x] 1.1 Create up migration for refresh_tokens table
    - Create file `migrations/000009_add_refresh_tokens.up.sql`
    - Define refresh_tokens table with columns: id, user_id, token_hash, created_at, expires_at, last_used_at
    - Add foreign key constraint on user_id with ON DELETE CASCADE
    - Create indexes on user_id, expires_at, and token_hash
    - _Requirements: 5.1, 5.2, 5.3, 5.4_

  - [x] 1.2 Create down migration for refresh_tokens table
    - Create file `migrations/000009_add_refresh_tokens.down.sql`
    - Drop refresh_tokens table and all associated indexes
    - _Requirements: 5.5_

  - [x] 1.3 Run and verify migrations
    - Execute `make migrate-up` to apply migrations
    - Verify table structure in PostgreSQL
    - Test rollback with `make migrate-down`
    - _Requirements: 5.1, 5.2, 5.3, 5.4_

---

## 2. Data Models

 - [x] 2. Data Models
  - [x] 2.1 Create RefreshToken model
    - Create file `internal/models/refresh_token.go`
    - Define RefreshToken struct with GORM tags
    - Implement BeforeCreate hook for UUID generation
    - Implement TableName method
    - Implement IsExpired helper method
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

  - [x] 2.2 Write unit tests for RefreshToken model
    - Create file `internal/models/refresh_token_test.go`
    - Test UUID generation in BeforeCreate
    - Test IsExpired method with various timestamps
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

---

## 3. Repository Layer

 - [x] 3. Repository Layer
  - [x] 3.1 Define RefreshTokenRepository interface
    - Update file `internal/repository/interfaces.go`
    - Add RefreshTokenRepository interface with methods: Create, FindByTokenHash, FindByUserID, Update, Delete, DeleteByUserID, DeleteExpired
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

  - [x] 3.2 Implement RefreshTokenRepository
    - Create file `internal/repository/refresh_token_repository.go`
    - Implement all interface methods with proper context handling
    - Use GORM for database operations
    - Implement efficient queries with indexes
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

  - [x] 3.3 Update Repositories struct
    - Update file `internal/repository/repository.go`
    - Add RefreshToken field to Repositories struct
    - Initialize RefreshTokenRepository in NewRepositories function
    - _Requirements: 4.1_

  - [x] 3.4 Write unit tests for RefreshTokenRepository
    - Create file `internal/repository/refresh_token_repository_test.go`
    - Test Create with valid data
    - Test FindByTokenHash with existing and non-existing tokens
    - Test FindByUserID with multiple tokens
    - Test Update with modified fields
    - Test Delete operations
    - Test DeleteExpired with mixed expired/valid tokens
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

---

## 4. Service Layer - Token Generation and Validation

 - [x] 4. Service Layer - Token Generation and Validation
  - [x] 4.1 Update Auth Service struct
    - Update file `internal/auth/service.go`
    - Add refreshTokenRepo field to Service struct
    - Add refreshTokenExpiry field (30 days)
    - Update NewService constructor to accept RefreshTokenRepository
    - _Requirements: 1.1, 1.2, 1.3, 2.1, 3.1, 4.1_

  - [x] 4.2 Implement GenerateRefreshToken method
    - Update file `internal/auth/service.go`
    - Generate 32 bytes of cryptographically secure random data
    - Encode token to base64 URL-safe string
    - Hash token with bcrypt before storage
    - Create RefreshToken record with 30-day expiration
    - Store in database via repository
    - Return plain-text token to caller
    - _Requirements: 1.1, 3.1, 3.2, 4.1, 4.2, 4.3_

  - [x] 4.3 Implement ValidateRefreshToken method
    - Update file `internal/auth/service.go`
    - Query all refresh tokens from database
    - Compare provided token against stored hashes using bcrypt
    - Check token expiration and delete if expired
    - Update last_used_at timestamp on valid token
    - Generate new refresh token (token rotation)
    - Delete old refresh token
    - Return user and new refresh token
    - _Requirements: 1.1, 1.2, 1.3, 1.5, 4.2, 4.4, 4.5_

  - [x] 4.4 Implement RevokeRefreshToken method
    - Update file `internal/auth/service.go`
    - Find refresh token by comparing hashes
    - Delete token from database
    - Return appropriate error if token not found
    - _Requirements: 2.1, 2.2, 2.3, 2.4_

  - [x] 4.5 Implement CleanupExpiredTokens method
    - Update file `internal/auth/service.go`
    - Call repository DeleteExpired method
    - Return count of deleted tokens
    - _Requirements: 4.5, 8.2_

  - [x] 4.6 Write unit tests for Auth Service token methods
    - Create or update file `internal/auth/service_test.go`
    - Test GenerateRefreshToken success and error cases
    - Test ValidateRefreshToken with valid, expired, and invalid tokens
    - Test token rotation (old token invalid after refresh)
    - Test RevokeRefreshToken success and error cases
    - Test CleanupExpiredTokens with mixed data
    - _Requirements: 1.1, 1.2, 1.3, 2.1, 2.2, 2.3, 4.1, 4.2, 4.3, 4.4, 4.5_

---

## 5. Handler Layer - Request/Response Types

 - [x] 5. Handler Layer - Request/Response Types
  - [x] 5.1 Update LoginResponse struct
    - Update file `internal/auth/handlers.go`
    - Add RefreshToken field to LoginResponse struct
    - Add swagger annotation for refresh_token field
    - Update example values in swagger comments
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 6.4_

  - [x] 5.2 Create RefreshRequest and RefreshResponse structs
    - Update file `internal/auth/handlers.go`
    - Define RefreshRequest with refresh_token field and validation tags
    - Define RefreshResponse with token, refresh_token, and expires_at fields
    - Add swagger annotations with examples
    - _Requirements: 1.1, 1.2, 1.3, 6.2_

  - [x] 5.3 Create LogoutRequest struct
    - Update file `internal/auth/handlers.go`
    - Define LogoutRequest with refresh_token field and validation tags
    - Add swagger annotations with examples
    - _Requirements: 2.1, 2.2, 6.3_

---

## 6. Handler Layer - Endpoint Implementation
- [x] 6. Handler Layer - Endpoint Implementation
  - [x] 6.1 Update Login handler to include refresh token
    - Update file `internal/auth/handlers.go`
    - Call service.GenerateRefreshToken after successful authentication
    - Include refresh_token in LoginResponse
    - Handle refresh token generation errors
    - Update swagger annotations
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 6.1, 6.4, 6.7_

  - [x] 6.2 Implement RefreshToken handler
    - Update file `internal/auth/handlers.go`
    - Bind and validate RefreshRequest
    - Call service.ValidateRefreshToken
    - Generate new access token
    - Return RefreshResponse with new tokens
    - Handle errors with appropriate HTTP status codes (401, 429, 500)
    - Use internal_handlers.ErrorResponse format for errors
    - Add swagger annotations
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 6.2, 6.4, 6.5, 6.7_

  - [x] 6.3 Implement Logout handler
    - Update file `internal/auth/handlers.go`
    - Bind and validate LogoutRequest
    - Call service.RevokeRefreshToken
    - Return 204 No Content on success
    - Handle errors with appropriate HTTP status codes (401, 500)
    - Use internal_handlers.ErrorResponse format for errors
    - Add swagger annotations
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 6.3, 6.4, 6.5, 6.7_

  - [x] 6.4 Write unit tests for auth handlers
    - Create or update file `internal/auth/handlers_test.go`
    - Test Login handler returns refresh_token
    - Test RefreshToken handler success case
    - Test RefreshToken handler with expired token (401)
    - Test RefreshToken handler with invalid token (401)
    - Test Logout handler success case (204)
    - Test Logout handler with invalid token (401)
    - Verify error response format matches internal_handlers.ErrorResponse
    - _Requirements: 1.1, 1.2, 1.3, 2.1, 2.2, 3.1, 3.2, 3.3, 6.1, 6.2, 6.3, 6.4, 6.5_

---

## 7. Rate Limiting
- [ ]* 7. Rate Limiting
  - [ ]* 7.1 Implement rate limiter
    - Create file `internal/auth/rate_limiter.go`
    - Implement rateLimiter struct with in-memory storage
    - Implement newRateLimiter constructor with configurable limit and window
    - Implement allow method to check and update rate limits
    - Implement cleanup goroutine to remove expired entries
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_

  - [ ]* 7.2 Implement RefreshRateLimitMiddleware
    - Update file `internal/auth/rate_limiter.go`
    - Create middleware function that uses rateLimiter
    - Use client IP as rate limit key
    - Return 429 with Retry-After header when limit exceeded
    - Use internal_handlers.ErrorResponse format for errors
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 6.4, 6.5_

  - [ ]* 7.3 Write unit tests for rate limiter
    - Create file `internal/auth/rate_limiter_test.go`
    - Test rate limiter allows requests within limit
    - Test rate limiter blocks requests exceeding limit
    - Test rate limiter resets after window expires
    - Test cleanup goroutine removes expired entries
    - Test middleware returns 429 with correct headers
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_

---

## 8. Background Cleanup Job

- [ ]* 8. Background Cleanup Job
- [ ]* 8.1 Implement CleanupService
  - Create file `internal/auth/cleanup.go`
  - Define CleanupService struct with authService, logger, ticker, and done channel
  - Implement NewCleanupService constructor
  - Implement Start method with 24-hour ticker
  - Implement Stop method for graceful shutdown
  - Implement runCleanup method that calls service.CleanupExpiredTokens
  - Add structured logging for cleanup operations
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

- [ ]* 8.2 Integrate CleanupService in server
  - Update file `cmd/server/main.go`
  - Initialize CleanupService after authService
  - Call Start method on application startup
  - Call Stop method in defer for graceful shutdown
  - _Requirements: 8.1, 8.2, 8.5_

- [ ]* 8.3 Write unit tests for CleanupService
  - Create file `internal/auth/cleanup_test.go`
  - Test Start method initializes ticker and goroutine
  - Test Stop method gracefully stops cleanup job
  - Test runCleanup method calls service correctly
  - Test cleanup runs on schedule
  - Test error handling and logging
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

---

## 9. Routing Configuration

 - [x] 9. Routing Configuration
  - [x] 9.1 Add new auth routes
    - Update file `internal/server/routes/routes.go`
    - Add POST /auth/refresh route with RefreshRateLimitMiddleware
    - Add POST /auth/logout route
    - Ensure routes are properly ordered in authGroup
    - _Requirements: 1.1, 2.1, 6.2, 6.3, 7.1_

  - [x] 9.2 Update service initialization
    - Update file `internal/server/routes/routes.go`
    - Pass RefreshTokenRepository to authService constructor
    - Verify all dependencies are properly injected
    - _Requirements: 4.1_

---

## 10. API Documentation

- [x] 10. API Documentation
  - [x] 10.1 Generate swagger documentation
    - Run `make swagger` to generate docs/swagger.json
    - Verify POST /auth/login includes refresh_token in response
    - Verify POST /auth/refresh endpoint is documented
    - Verify POST /auth/logout endpoint is documented
    - Verify all error responses use ErrorResponse format
    - Verify all HTTP status codes are documented (200, 204, 401, 429, 500)
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7_

  - [x] 10.2 Update API client documentation
    - Update file `docs/api-client-export.md`
    - Add POST /auth/refresh endpoint documentation
    - Add POST /auth/logout endpoint documentation
    - Update POST /auth/login response to include refresh_token
    - Add TypeScript interfaces for new request/response types
    - Add client integration examples
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.7_

---

## 11. Integration Testing

 - [ ] 11. Integration Testing
  - [ ]* 11.1 Write end-to-end auth flow test
    - Create file `internal/integration/auth_session_test.go`
    - Test complete flow: login → use access token → refresh → logout
    - Verify tokens work correctly at each step
    - Verify old refresh token is invalid after refresh
    - Verify tokens are invalid after logout
    - _Requirements: 1.1, 1.2, 1.3, 1.5, 2.1, 2.2, 2.3, 2.4, 3.1, 3.2, 3.3, 3.4, 3.5_

  - [ ]* 11.2 Write token rotation test
    - Update file `internal/integration/auth_session_test.go`
    - Test that old refresh token becomes invalid after refresh
    - Test that new refresh token works correctly
    - Verify token rotation prevents replay attacks
    - _Requirements: 1.5, 4.4_

  - [ ]* 11.3 Write rate limiting integration test
    - Update file `internal/integration/auth_session_test.go`
    - Test that refresh endpoint enforces rate limits
    - Test that 429 response includes Retry-After header
    - Test that rate limit resets after window
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_

  - [ ]* 11.4 Write cleanup job integration test
    - Update file `internal/integration/auth_session_test.go`
    - Create expired and valid refresh tokens
    - Run cleanup job
    - Verify only expired tokens are deleted
    - Verify valid tokens remain
    - _Requirements: 8.1, 8.2, 8.3, 8.4_

---

## 12. Documentation and Finalization

- [ ] 12. Documentation and Finalization
  - [ ] 12.1 Update README
    - Update file `README.md`
    - Add section on authentication flow with refresh tokens
    - Document token expiration times
    - Add examples of login, refresh, and logout
    - _Requirements: All_

  - [ ] 12.2 Create migration guide
      - Create file `docs/auth-migration-guide.md`
      - Document changes to /auth/login response
      - Provide client code examples for token refresh
      - Provide client code examples for logout
      - Document error handling best practices
      - _Requirements: All_

    - [x] 12.3 Verify all requirements are met
      - Review requirements document
      - Verify each requirement has corresponding implementation
      - Verify all acceptance criteria are satisfied
      - Run full test suite
      - _Requirements: All_
