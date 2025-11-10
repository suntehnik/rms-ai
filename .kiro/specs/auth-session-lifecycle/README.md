# Auth Session Lifecycle - Spec Summary

## Overview

This spec implements JWT refresh token functionality and explicit logout capabilities for the Spexus Product Requirements Management System, based on Epic EP-019 and User Story US-057.

## Requirements Covered

- **REQ-085**: POST /auth/refresh endpoint for token renewal
- **REQ-086**: POST /auth/logout endpoint for session termination
- **REQ-087**: Enhanced POST /auth/login response with refresh_token
- **REQ-088**: Comprehensive Swagger documentation for all endpoints

## Key Features

### Core Functionality
1. **Token Refresh** - Seamless token renewal without re-authentication
2. **Explicit Logout** - Secure session termination
3. **Token Rotation** - Automatic refresh token rotation for security
4. **Rate Limiting** - 10 requests/minute to prevent abuse

### Security
- Bcrypt hashing for refresh token storage
- 30-day refresh token expiration
- 24-hour access token expiration
- Automatic cleanup of expired tokens (optional)

### Database
- New `refresh_tokens` table with proper indexes
- Foreign key constraints with CASCADE delete
- Efficient queries for token validation

## Implementation Structure

### Required Tasks (Core MVP)
1. Database schema and migrations
2. Data models (RefreshToken)
3. Repository layer (CRUD operations)
4. Service layer (token generation, validation, revocation)
5. Handler layer (Login, Refresh, Logout endpoints)
6. Rate limiting middleware
7. Routing configuration
8. API documentation (Swagger)

### Optional Tasks (Can be added later)
- Unit tests for all layers
- Integration tests for auth flows
- Background cleanup job for expired tokens
- Comprehensive documentation

## Quick Start

To begin implementation:

1. Review `requirements.md` for detailed acceptance criteria
2. Review `design.md` for architecture and component details
3. Follow `tasks.md` sequentially for implementation
4. Start with Task 1 (Database Schema and Migration)

## Files Structure

```
.kiro/specs/auth-session-lifecycle/
├── README.md           # This file
├── requirements.md     # Detailed requirements with EARS patterns
├── design.md          # Architecture and component design
└── tasks.md           # Implementation task list
```

## Next Steps

1. Get approval for requirements, design, and tasks
2. Begin implementation starting with database migrations
3. Follow the task list sequentially
4. Test each component as it's implemented
5. Generate and verify Swagger documentation

## Related Documentation

- Epic: EP-019 "[BE] Session lifecycle endpoints"
- User Story: US-057 "Refresh и logout эндпоинты для фронта"
- Requirements: REQ-085, REQ-086, REQ-087, REQ-088

## Notes

- The implementation is backward compatible with existing clients
- Existing `/auth/login` endpoint is enhanced, not replaced
- New endpoints are additive and don't break existing functionality
- Background cleanup job is optional and can be added later if needed
