# Security Guide - Product Requirements Management API

## Overview

The Product Requirements Management API implements a comprehensive security model based on JWT (JSON Web Token) authentication and role-based access control (RBAC). This guide provides detailed information about authentication, authorization, security best practices, and error handling.

## Authentication System

### JWT Token Authentication

The API uses JWT tokens for authentication with the following characteristics:

- **Algorithm**: HS256 (HMAC with SHA-256)
- **Token Lifetime**: 1 hour (3600 seconds)
- **Header Format**: `Authorization: Bearer <jwt_token>`
- **Token Claims**: User ID, Username, Role, Expiration, Issued At, Not Before

### Authentication Flow

1. **Login**: POST `/auth/login` with username and password
2. **Token Generation**: Server validates credentials and returns JWT token
3. **Token Usage**: Include token in `Authorization` header for all authenticated requests
4. **Token Expiration**: Tokens expire after 1 hour and require re-authentication

### Example Authentication Request

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "password123"
  }'
```

### Example Authentication Response

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "username": "john_doe",
    "email": "john.doe@example.com",
    "role": "User"
  },
  "expires_at": "2023-01-02T12:30:00Z"
}
```

## Authorization System

### Role-Based Access Control (RBAC)

The system implements three distinct user roles with hierarchical permissions:

#### 1. Administrator Role
- **Permission Level**: 3 (Highest)
- **Capabilities**:
  - Full system access
  - User management (create, edit, delete users)
  - System configuration (requirement types, relationship types, status models)
  - All entity operations (create, edit, delete, view)
  - Comment system (create, edit, resolve, delete)

#### 2. User Role
- **Permission Level**: 2 (Standard)
- **Capabilities**:
  - Entity management (create, edit, delete epics, user stories, requirements, acceptance criteria)
  - View all entities
  - Comment system (create, edit own comments, resolve comments)
  - Cannot manage users or system configuration

#### 3. Commenter Role
- **Permission Level**: 1 (Limited)
- **Capabilities**:
  - View all entities
  - Comment system (create comments, edit own comments, resolve comments)
  - Cannot create, edit, or delete entities
  - Cannot manage users or system configuration

### Permission Matrix

| Operation | Administrator | User | Commenter |
|-----------|---------------|------|-----------|
| Create Entities | ✅ | ✅ | ❌ |
| Edit Entities | ✅ | ✅ | ❌ |
| Delete Entities | ✅ | ✅ | ❌ |
| View Entities | ✅ | ✅ | ✅ |
| Create Comments | ✅ | ✅ | ✅ |
| Edit Own Comments | ✅ | ✅ | ✅ |
| Edit Any Comments | ✅ | ✅ | ❌ |
| Resolve Comments | ✅ | ✅ | ✅ |
| Manage Users | ✅ | ❌ | ❌ |
| System Configuration | ✅ | ❌ | ❌ |

## Security Middleware

### Authentication Middleware Flow

1. **Header Extraction**: Extract `Authorization` header from request
2. **Token Validation**: Validate JWT token signature and structure
3. **Expiration Check**: Verify token has not expired
4. **Claims Extraction**: Extract user information from token
5. **Role Authorization**: Check if user role meets endpoint requirements
6. **Context Storage**: Store user claims in request context for handlers

### Role-Based Middleware

The system provides role-specific middleware functions:

- `RequireAdministrator()`: Requires Administrator role
- `RequireUser()`: Requires User role or higher (User, Administrator)
- `RequireCommenter()`: Requires any authenticated user (Commenter, User, Administrator)

## Error Responses

### Authentication Errors (HTTP 401)

#### Missing Authentication
```json
{
  "error": "Authorization header required",
  "code": "AUTHENTICATION_REQUIRED"
}
```

#### Invalid Token Format
```json
{
  "error": "Bearer token required",
  "code": "INVALID_TOKEN_FORMAT"
}
```

#### Invalid Token
```json
{
  "error": "Invalid token",
  "code": "INVALID_TOKEN"
}
```

#### Expired Token
```json
{
  "error": "Token expired",
  "code": "TOKEN_EXPIRED"
}
```

### Authorization Errors (HTTP 403)

#### Insufficient Permissions
```json
{
  "error": "Insufficient permissions",
  "code": "INSUFFICIENT_PERMISSIONS"
}
```

#### Role Required
```json
{
  "error": "Administrator role required",
  "code": "ADMINISTRATOR_ROLE_REQUIRED"
}
```

## Security Best Practices

### Client-Side Security

#### Token Storage
- **Recommended**: Secure httpOnly cookies or encrypted localStorage
- **Avoid**: Plain localStorage, sessionStorage, URL parameters, local variables
- **Reasoning**: Prevent XSS attacks and token theft

#### Token Transmission
- **Protocol**: HTTPS only in production
- **Header Format**: `Authorization: Bearer <token>`
- **Validation**: Always validate token format before sending

#### Error Handling
- **Expired Token**: Automatically redirect to login or attempt refresh
- **Invalid Token**: Clear stored token and redirect to login
- **Network Errors**: Implement retry with exponential backoff
- **Server Errors**: Show user-friendly error messages

### Server-Side Security

#### Token Generation
- **Algorithm**: HS256 (HMAC with SHA-256)
- **Secret Key**: Use strong, randomly generated secret key
- **Expiration**: Short-lived tokens (1 hour recommended)
- **Claims**: Include minimal necessary user information

#### Token Validation
- **Signature Check**: Always verify token signature
- **Expiration Check**: Reject expired tokens immediately
- **Claims Validation**: Validate all required claims are present
- **Blacklisting**: Optional token revocation mechanism

#### Security Headers
- **CORS**: Configure appropriate CORS policies
- **CSP**: Set Content Security Policy headers
- **HSTS**: Use HTTP Strict Transport Security

## API Endpoint Security

### Public Endpoints (No Authentication Required)
- `POST /auth/login` - User authentication
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /live` - Liveness check

### Authenticated Endpoints (Any Role)
- `GET /api/v1/epics` - List epics
- `GET /api/v1/epics/{id}` - Get epic details
- `GET /api/v1/search` - Search entities
- `GET /auth/profile` - Get current user profile

### User Role Required
- `POST /api/v1/epics` - Create epic
- `PUT /api/v1/epics/{id}` - Update epic
- `DELETE /api/v1/epics/{id}` - Delete epic
- All User Story, Requirement, and Acceptance Criteria CRUD operations

### Administrator Role Required
- `POST /auth/users` - Create user
- `GET /auth/users` - List users
- `PUT /auth/users/{id}` - Update user
- `DELETE /auth/users/{id}` - Delete user
- All configuration endpoints (`/api/v1/config/*`)

## Security Monitoring and Auditing

### Logging Requirements
- Log all authentication attempts with timestamp, IP, and result
- Log all authorization failures with user and attempted operation
- Log token generation, validation failures, and expiration events
- Log suspicious activity patterns

### Monitoring Metrics
- Track successful vs failed authentication rates
- Monitor token usage patterns and expiration rates
- Track distribution of user roles and permission usage
- Monitor 401/403 error rates and patterns

### Security Alerts
- Alert on multiple failed login attempts from same IP
- Alert on attempts to access admin functions without proper role
- Alert on unusual token usage patterns or validation failures
- Alert on access to sensitive configuration endpoints

## Testing Authentication

### Using Swagger UI
1. Click the "Authorize" button in Swagger UI
2. Enter your JWT token in the format: `Bearer <your_token>`
3. Click "Authorize" to apply the token to all requests
4. Test endpoints with the "Try it out" functionality

### Using cURL
```bash
# Get JWT token
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","password":"password123"}' | \
  jq -r '.token')

# Use token in authenticated request
curl -X GET http://localhost:8080/api/v1/epics \
  -H "Authorization: Bearer $TOKEN"
```

### Using Postman
1. Set up environment variable for token
2. Add Authorization header with value `Bearer {{token}}`
3. Use pre-request scripts to automatically refresh tokens

## Troubleshooting

### Common Issues

#### 401 Unauthorized
- Check if Authorization header is present
- Verify token format (must start with "Bearer ")
- Check if token has expired
- Validate token signature

#### 403 Forbidden
- Verify user role meets endpoint requirements
- Check if user account is active
- Confirm role permissions in authorization matrix

#### Token Expiration
- Implement automatic token refresh
- Handle expiration gracefully in client applications
- Provide clear user feedback for expired sessions

### Debug Tips
- Use browser developer tools to inspect request headers
- Check server logs for detailed error messages
- Validate JWT tokens using online JWT debuggers (for development only)
- Test with different user roles to verify permission boundaries

## Security Updates

This security guide should be reviewed and updated whenever:
- New authentication mechanisms are added
- Role permissions are modified
- New security vulnerabilities are discovered
- API endpoints are added or modified
- Security best practices evolve

For the latest security information and updates, refer to the API documentation and security advisories.