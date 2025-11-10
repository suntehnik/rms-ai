# Requirements Document: Auth Session Lifecycle

## Introduction

This document specifies the requirements for implementing session lifecycle management in the Spexus Product Requirements Management System. The feature adds JWT refresh token functionality and explicit logout capabilities to support frontend session management, reducing login friction while maintaining security.

## Glossary

- **Access Token**: Short-lived JWT token used for API authentication (current implementation)
- **Refresh Token**: Long-lived secure token used to obtain new access tokens without re-authentication
- **Session**: A user's authenticated state tracked by refresh token
- **Token Rotation**: Security practice of issuing new refresh tokens on each refresh operation
- **Spexus Frontend**: The web UI client consuming the authentication API
- **Auth Service**: The internal/auth package handling authentication operations
- **Token Store**: Database storage for refresh tokens and session metadata

## Requirements

### Requirement 1: POST /auth/refresh Endpoint

**User Story:** As the Spexus frontend client, I want to refresh expired access tokens using a refresh token, so that users can maintain their session without re-entering credentials.

#### Acceptance Criteria

1. WHEN Spexus frontend sends a valid refresh_token to POST /auth/refresh, THEN the Auth Service SHALL validate the refresh token against the Token Store and return HTTP 200 with a new access token, new refresh token, and expires_at timestamp in ISO-8601 format.

2. WHEN Spexus frontend sends an expired refresh_token to POST /auth/refresh, THEN the Auth Service SHALL return HTTP 401 with error code "REFRESH_TOKEN_EXPIRED" and message "Refresh token has expired" in internal_handlers.ErrorResponse format.

3. WHEN Spexus frontend sends an invalid or revoked refresh_token to POST /auth/refresh, THEN the Auth Service SHALL return HTTP 401 with error code "INVALID_REFRESH_TOKEN" and message "Invalid or revoked refresh token" in internal_handlers.ErrorResponse format.

4. WHEN Spexus frontend exceeds the rate limit for refresh attempts, THEN the Auth Service SHALL return HTTP 429 with error code "RATE_LIMIT_EXCEEDED" and message "Too many refresh attempts" in internal_handlers.ErrorResponse format.

5. WHEN the Auth Service successfully refreshes tokens, THEN the Auth Service SHALL invalidate the old refresh token and store the new refresh token with session metadata including user_id, created_at, expires_at, and last_used_at timestamps.

### Requirement 2: POST /auth/logout Endpoint

**User Story:** As the Spexus frontend client, I want to explicitly log out users and invalidate their sessions, so that administrators, users, and commenters can securely end their authenticated sessions.

#### Acceptance Criteria

1. WHEN Spexus frontend sends a valid refresh_token to POST /auth/logout, THEN the Auth Service SHALL invalidate the refresh token in the Token Store, delete the session record, and return HTTP 204 with no response body.

2. WHEN Spexus frontend sends an already-invalidated refresh_token to POST /auth/logout, THEN the Auth Service SHALL return HTTP 401 with error code "INVALID_REFRESH_TOKEN" and message "Session already logged out" in internal_handlers.ErrorResponse format.

3. WHEN Spexus frontend attempts to use a refresh_token after logout, THEN the Auth Service SHALL return HTTP 401 with error code "SESSION_INVALIDATED" and message "Session has been logged out" in internal_handlers.ErrorResponse format.

4. WHEN the Auth Service successfully logs out a session, THEN the Auth Service SHALL remove all associated refresh tokens for that session from the Token Store.

### Requirement 3: Enhanced POST /auth/login Response

**User Story:** As the Spexus frontend client, I want to receive a refresh token during login, so that I can implement seamless token refresh without requiring users to re-authenticate.

#### Acceptance Criteria

1. WHEN a user successfully authenticates via POST /auth/login, THEN the Auth Service SHALL generate both an access token and a refresh token and include both in the LoginResponse structure.

2. WHEN the Auth Service generates tokens for login, THEN the LoginResponse SHALL include fields: token (access JWT), refresh_token (secure refresh token), expires_at (access token expiration in ISO-8601), and user (UserResponse object).

3. WHEN the Auth Service creates a refresh token, THEN the Auth Service SHALL store the refresh token in the Token Store with fields: id (UUID), user_id (UUID), token_hash (bcrypt hash), created_at, expires_at (30 days from creation), and last_used_at (initially null).

4. WHEN the Auth Service generates a refresh token, THEN the refresh token SHALL be compatible with POST /auth/refresh and POST /auth/logout endpoints.

5. WHEN the Auth Service returns LoginResponse, THEN the response structure SHALL match the internal/auth.LoginResponse type definition and be documented in swagger with request/response examples.

### Requirement 4: Refresh Token Storage and Security

**User Story:** As a system administrator, I want refresh tokens to be securely stored and managed, so that the authentication system maintains security best practices.

#### Acceptance Criteria

1. WHEN the Auth Service stores a refresh token, THEN the Auth Service SHALL hash the token using bcrypt before storing in the database and SHALL NOT store the plain-text token.

2. WHEN the Auth Service validates a refresh token, THEN the Auth Service SHALL use bcrypt comparison against the stored hash.

3. WHEN a refresh token is created, THEN the refresh token SHALL have an expiration time of 30 days from creation.

4. WHEN a refresh token is used successfully, THEN the Auth Service SHALL update the last_used_at timestamp in the Token Store.

5. WHEN the Auth Service detects an expired refresh token, THEN the Auth Service SHALL automatically remove the expired token from the Token Store during validation.

### Requirement 5: Database Migration for Refresh Tokens

**User Story:** As a system administrator, I want a database schema for refresh token storage, so that the system can persist and manage refresh tokens reliably.

#### Acceptance Criteria

1. WHEN the database migration runs, THEN the migration SHALL create a table named "refresh_tokens" with columns: id (UUID primary key), user_id (UUID foreign key to users), token_hash (text not null), created_at (timestamp), expires_at (timestamp), last_used_at (timestamp nullable).

2. WHEN the database migration creates the refresh_tokens table, THEN the migration SHALL create an index on user_id for efficient user session queries.

3. WHEN the database migration creates the refresh_tokens table, THEN the migration SHALL create an index on expires_at for efficient cleanup of expired tokens.

4. WHEN the database migration creates the refresh_tokens table, THEN the migration SHALL set a foreign key constraint on user_id with ON DELETE CASCADE to automatically remove tokens when users are deleted.

5. WHEN the down migration runs, THEN the migration SHALL drop the refresh_tokens table and all associated indexes.

### Requirement 6: API Documentation in Swagger (REQ-088)

**User Story:** As a frontend developer, I want comprehensive API documentation for the new endpoints, so that I can correctly implement the authentication flow.

#### Acceptance Criteria

1. WHEN the backend implements POST /auth/refresh, POST /auth/logout, and enhanced POST /auth/login, THEN swagger annotations SHALL describe all parameters, response fields, error examples, and authentication requirements in the code.

2. WHEN the swagger documentation is generated, THEN docs/swagger.json SHALL include POST /auth/refresh with request schema (refresh_token: string), response schemas (200: RefreshResponse, 401: ErrorResponse, 429: ErrorResponse, 500: ErrorResponse), and example requests/responses.

3. WHEN the swagger documentation is generated, THEN docs/swagger.json SHALL include POST /auth/logout with request schema (refresh_token: string), response schemas (204: no content, 401: ErrorResponse, 500: ErrorResponse), and example requests/responses.

4. WHEN the swagger documentation is generated, THEN docs/swagger.json SHALL update POST /auth/login response schema to include refresh_token field with type string and example value.

5. WHEN the swagger documentation is generated, THEN all error responses SHALL use the internal_handlers.ErrorResponse format with fields: error.code (string) and error.message (string).

6. WHEN the swagger documentation is generated, THEN all authentication endpoints SHALL be tagged with "authentication" and include security requirements where applicable.

7. WHEN the updated swagger documentation is published, THEN docs/swagger.json SHALL reflect all HTTP status codes: 200 (success), 204 (no content), 401 (unauthorized), 429 (rate limit), and 500 (internal error) with appropriate descriptions.

### Requirement 7: Rate Limiting for Refresh Endpoint

**User Story:** As a system administrator, I want rate limiting on the refresh endpoint, so that the system is protected against token refresh abuse.

#### Acceptance Criteria

1. WHEN Spexus frontend makes refresh requests, THEN the Auth Service SHALL limit refresh attempts to 10 requests per minute per user.

2. WHEN the rate limit is exceeded, THEN the Auth Service SHALL return HTTP 429 with error code "RATE_LIMIT_EXCEEDED" and include a Retry-After header with seconds until the limit resets.

3. WHEN the Auth Service tracks rate limits, THEN the Auth Service SHALL use Redis or in-memory storage with automatic expiration after 1 minute.

4. WHEN a successful refresh occurs, THEN the Auth Service SHALL NOT reset the rate limit counter.

5. WHEN the rate limit window expires, THEN the Auth Service SHALL automatically reset the counter to zero.

### Requirement 8: Token Cleanup Background Job

**User Story:** As a system administrator, I want automatic cleanup of expired refresh tokens, so that the database does not accumulate stale session data.

#### Acceptance Criteria

1. WHEN the application starts, THEN the Auth Service SHALL initialize a background job that runs every 24 hours to clean up expired tokens.

2. WHEN the cleanup job runs, THEN the Auth Service SHALL delete all refresh tokens where expires_at is less than the current timestamp.

3. WHEN the cleanup job completes, THEN the Auth Service SHALL log the number of expired tokens removed using the structured logger.

4. WHEN the cleanup job encounters an error, THEN the Auth Service SHALL log the error and continue running on the next scheduled interval.

5. WHEN the application shuts down, THEN the Auth Service SHALL gracefully stop the cleanup job.
