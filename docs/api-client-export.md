# Product Requirements Management API - Client Implementation Guide

## Overview

This document provides a comprehensive API reference for implementing a web UI client for the Product Requirements Management System. The API follows RESTful conventions and uses JSON for data exchange.

**Base URL**: `http://localhost:8080`
**API Version**: v1
**Authentication**: JWT Bearer tokens

## Table of Contents

1. [Authentication](#authentication)
2. [Core Entities](#core-entities)
3. [API Endpoints](#api-endpoints)
4. [TypeScript Interfaces](#typescript-interfaces)
5. [Error Handling](#error-handling)
6. [Search & Navigation](#search--navigation)
7. [Comments System](#comments-system)
8. [Configuration Management](#configuration-management)

---

## Authentication

### JWT Token Authentication
All API endpoints (except login, refresh, and health checks) require a JWT token in the Authorization header:
```
Authorization: Bearer <jwt_token>
```

### Session Management with Refresh Tokens

The authentication system uses JWT access tokens for API requests and refresh tokens for maintaining sessions without re-authentication. Access tokens are short-lived (configurable, typically 15-60 minutes), while refresh tokens are long-lived (30 days).

**Authentication Flow:**
1. Login with credentials to receive both access token and refresh token
2. Use access token for API requests
3. When access token expires, use refresh token to obtain new tokens
4. Logout to invalidate refresh token and end session

### Auth Endpoints

#### POST /auth/login
Authenticate user and receive JWT access token and refresh token.

```typescript
interface LoginRequest {
  username: string;
  password: string;
}

interface LoginResponse {
  token: string;              // JWT access token for API requests
  refresh_token: string;      // Refresh token for obtaining new access tokens
  expires_at: string;         // Access token expiration timestamp (ISO-8601)
  user: UserResponse;         // Authenticated user information
}
```

**Example Request:**
```typescript
const response = await fetch('/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    username: 'john_doe',
    password: 'password123'
  })
});

const data: LoginResponse = await response.json();
// Store both tokens securely
localStorage.setItem('access_token', data.token);
localStorage.setItem('refresh_token', data.refresh_token);
```

**Response Codes:**
- `200` - Successful authentication
- `400` - Invalid request format
- `401` - Invalid credentials
- `500` - Internal server error

#### POST /auth/refresh
Refresh an expired access token using a refresh token. This endpoint implements token rotation - each refresh returns a new access token AND a new refresh token, invalidating the old refresh token.

```typescript
interface RefreshRequest {
  refresh_token: string;      // Current refresh token
}

interface RefreshResponse {
  token: string;              // New JWT access token
  refresh_token: string;      // New refresh token (token rotation)
  expires_at: string;         // New access token expiration timestamp (ISO-8601)
}
```

**Example Request:**
```typescript
const refreshToken = localStorage.getItem('refresh_token');

const response = await fetch('/auth/refresh', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    refresh_token: refreshToken
  })
});

if (response.ok) {
  const data: RefreshResponse = await response.json();
  // Update stored tokens
  localStorage.setItem('access_token', data.token);
  localStorage.setItem('refresh_token', data.refresh_token);
} else if (response.status === 401) {
  // Refresh token expired or invalid - redirect to login
  window.location.href = '/login';
}
```

**Response Codes:**
- `200` - Successfully refreshed tokens
- `400` - Invalid request format (ErrorResponse)
- `401` - Invalid or expired refresh token (ErrorResponse)
- `429` - Too many refresh attempts (ErrorResponse with Retry-After header)
- `500` - Internal server error (ErrorResponse)

**Error Response Format:**
```typescript
interface ErrorResponse {
  error: {
    code: string;           // Error code for programmatic handling
    message: string;        // Human-readable error message
  };
}
```

**Error Codes:**
- `VALIDATION_ERROR` - Request validation failed
- `REFRESH_TOKEN_EXPIRED` - Refresh token has expired
- `INVALID_REFRESH_TOKEN` - Invalid or revoked refresh token
- `RATE_LIMIT_EXCEEDED` - Too many refresh attempts (includes Retry-After header)
- `INTERNAL_ERROR` - Server-side error

#### POST /auth/logout
Logout user and invalidate refresh token to end the session.

```typescript
interface LogoutRequest {
  refresh_token: string;      // Refresh token to invalidate
}
```

**Example Request:**
```typescript
const refreshToken = localStorage.getItem('refresh_token');

const response = await fetch('/auth/logout', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    refresh_token: refreshToken
  })
});

if (response.status === 204) {
  // Successfully logged out - clear stored tokens
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
  window.location.href = '/login';
}
```

**Response Codes:**
- `204` - Successfully logged out (no content)
- `400` - Invalid request format (ErrorResponse)
- `401` - Invalid refresh token (ErrorResponse)
- `500` - Internal server error (ErrorResponse)

#### GET /auth/profile
Get current user profile (requires authentication)

#### POST /auth/change-password
```typescript
interface ChangePasswordRequest {
  current_password: string;
  new_password: string; // minimum 8 characters
}
```

#### User Management (Admin Only)
- `POST /auth/users` - Create user
- `GET /auth/users` - List users
- `GET /auth/users/:id` - Get user
- `PUT /auth/users/:id` - Update user
- `DELETE /auth/users/:id` - Delete user

### Client Integration Examples

#### Automatic Token Refresh
Implement automatic token refresh when access token expires:

```typescript
class ApiClient {
  private accessToken: string | null = null;
  private refreshToken: string | null = null;

  constructor() {
    this.accessToken = localStorage.getItem('access_token');
    this.refreshToken = localStorage.getItem('refresh_token');
  }

  async request(url: string, options: RequestInit = {}): Promise<Response> {
    // Add access token to request
    const headers = {
      ...options.headers,
      'Authorization': `Bearer ${this.accessToken}`
    };

    let response = await fetch(url, { ...options, headers });

    // If 401, try to refresh token
    if (response.status === 401 && this.refreshToken) {
      const refreshed = await this.refreshAccessToken();
      
      if (refreshed) {
        // Retry original request with new token
        headers['Authorization'] = `Bearer ${this.accessToken}`;
        response = await fetch(url, { ...options, headers });
      } else {
        // Refresh failed - redirect to login
        window.location.href = '/login';
      }
    }

    return response;
  }

  async refreshAccessToken(): Promise<boolean> {
    try {
      const response = await fetch('/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: this.refreshToken })
      });

      if (response.ok) {
        const data: RefreshResponse = await response.json();
        this.accessToken = data.token;
        this.refreshToken = data.refresh_token;
        localStorage.setItem('access_token', data.token);
        localStorage.setItem('refresh_token', data.refresh_token);
        return true;
      }
    } catch (error) {
      console.error('Token refresh failed:', error);
    }

    return false;
  }

  async login(username: string, password: string): Promise<boolean> {
    try {
      const response = await fetch('/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      });

      if (response.ok) {
        const data: LoginResponse = await response.json();
        this.accessToken = data.token;
        this.refreshToken = data.refresh_token;
        localStorage.setItem('access_token', data.token);
        localStorage.setItem('refresh_token', data.refresh_token);
        return true;
      }
    } catch (error) {
      console.error('Login failed:', error);
    }

    return false;
  }

  async logout(): Promise<void> {
    if (this.refreshToken) {
      try {
        await fetch('/auth/logout', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: this.refreshToken })
        });
      } catch (error) {
        console.error('Logout request failed:', error);
      }
    }

    // Clear tokens regardless of API response
    this.accessToken = null;
    this.refreshToken = null;
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
  }
}

// Usage
const api = new ApiClient();

// Login
await api.login('john_doe', 'password123');

// Make authenticated requests
const response = await api.request('/api/v1/epics');

// Logout
await api.logout();
```

#### Proactive Token Refresh
Refresh token before it expires to avoid interruptions:

```typescript
class TokenManager {
  private refreshTimer: number | null = null;

  startAutoRefresh(expiresAt: string) {
    // Refresh 5 minutes before expiration
    const expiresAtMs = new Date(expiresAt).getTime();
    const refreshAtMs = expiresAtMs - (5 * 60 * 1000);
    const delayMs = refreshAtMs - Date.now();

    if (delayMs > 0) {
      this.refreshTimer = window.setTimeout(() => {
        this.refreshToken();
      }, delayMs);
    }
  }

  stopAutoRefresh() {
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer);
      this.refreshTimer = null;
    }
  }

  async refreshToken() {
    const refreshToken = localStorage.getItem('refresh_token');
    if (!refreshToken) return;

    try {
      const response = await fetch('/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken })
      });

      if (response.ok) {
        const data: RefreshResponse = await response.json();
        localStorage.setItem('access_token', data.token);
        localStorage.setItem('refresh_token', data.refresh_token);
        
        // Schedule next refresh
        this.startAutoRefresh(data.expires_at);
      } else {
        // Refresh failed - redirect to login
        window.location.href = '/login';
      }
    } catch (error) {
      console.error('Token refresh failed:', error);
      window.location.href = '/login';
    }
  }
}
```

### Health Check Endpoints (Public)
- `GET /ready` - Readiness check (no authentication required)
- `GET /live` - Liveness check (no authentication required)

---

## Core Entities

### Hierarchical Structure
```
Epic (EP-001)
├── User Story (US-001)
│   ├── Acceptance Criteria (AC-001)
│   └── Requirements (REQ-001)
└── User Story (US-002)
    └── Requirements (REQ-002)
```

### Entity Status Workflows

#### Epic Status
- `Backlog` → `Draft` → `In Progress` → `Done`
- `Cancelled` (from any state)

#### User Story Status  
- `Backlog` → `Draft` → `In Progress` → `Done`
- `Cancelled` (from any state)

#### Requirement Status
- `Draft` → `Active` → `Obsolete`

### Priority Levels
- `1` - Critical (highest urgency)
- `2` - High (important)
- `3` - Medium (normal)
- `4` - Low (can be deferred)

---

## API Endpoints

### Epics (`/api/v1/epics`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/` | Create epic |
| GET | `/` | List epics with filtering |
| GET | `/:id` | Get epic by ID/reference |
| PUT | `/:id` | Update epic |
| DELETE | `/:id` | Delete epic |
| GET | `/:id/user-stories` | Get epic with user stories |
| POST | `/:id/user-stories` | Create user story in epic |
| PATCH | `/:id/status` | Change epic status |
| PATCH | `/:id/assign` | Assign epic to user |
| GET | `/:id/validate-deletion` | Validate deletion |
| DELETE | `/:id/delete` | Comprehensive deletion |

**Query Parameters for List:**
- `creator_id` (UUID) - Filter by creator
- `assignee_id` (UUID) - Filter by assignee  
- `status` (EpicStatus) - Filter by status
- `priority` (1-4) - Filter by priority
- `order_by` (string) - Sort order
- `limit` (1-100) - Page size
- `offset` (number) - Pagination offset
- `include` (string) - Include related data: `creator,assignee,user_stories,comments`

### User Stories (`/api/v1/user-stories`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/` | Create user story |
| GET | `/` | List user stories with filtering |
| GET | `/:id` | Get user story by ID/reference |
| PUT | `/:id` | Update user story |
| DELETE | `/:id` | Delete user story |
| GET | `/:id/acceptance-criteria` | Get acceptance criteria |
| POST | `/:id/acceptance-criteria` | Create acceptance criteria |
| GET | `/:id/requirements` | Get requirements |
| POST | `/:id/requirements` | Create requirement |
| PATCH | `/:id/status` | Change status |
| PATCH | `/:id/assign` | Assign to user |
| GET | `/:id/validate-deletion` | Validate deletion |
| DELETE | `/:id/delete` | Comprehensive deletion |

**Query Parameters for List:**
- `epic_id` (UUID) - Filter by epic
- `creator_id` (UUID) - Filter by creator
- `assignee_id` (UUID) - Filter by assignee
- `status` (UserStoryStatus) - Filter by status
- `priority` (1-4) - Filter by priority
- `order_by` (string) - Sort order
- `limit` (1-100) - Page size
- `offset` (number) - Pagination offset
- `include` (string) - Include related data: `epic,creator,assignee,acceptance_criteria,requirements,comments`

### Acceptance Criteria (`/api/v1/acceptance-criteria`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | List acceptance criteria |
| GET | `/:id` | Get acceptance criteria |
| PUT | `/:id` | Update acceptance criteria |
| DELETE | `/:id` | Delete acceptance criteria |
| GET | `/:id/validate-deletion` | Validate deletion |
| DELETE | `/:id/delete` | Comprehensive deletion |

### Requirements (`/api/v1/requirements`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/` | Create requirement |
| GET | `/` | List requirements |
| GET | `/search` | Search requirements |
| GET | `/:id` | Get requirement |
| PUT | `/:id` | Update requirement |
| DELETE | `/:id` | Delete requirement |
| GET | `/:id/relationships` | Get with relationships |
| PATCH | `/:id/status` | Change status |
| PATCH | `/:id/assign` | Assign to user |
| POST | `/relationships` | Create relationship |
| GET | `/:id/validate-deletion` | Validate deletion |
| DELETE | `/:id/delete` | Comprehensive deletion |

### Requirement Relationships (`/api/v1/requirement-relationships`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| DELETE | `/:id` | Delete relationship |

---

## Search & Navigation

### Search (`/api/v1/search`)

#### GET /api/v1/search
Global search across all entities

**Query Parameters:**
- `q` (string, required) - Search query
- `entity_types` (string) - Comma-separated: `epic,user_story,acceptance_criteria,requirement`
- `limit` (1-100) - Results per page
- `offset` (number) - Pagination offset

#### GET /api/v1/search/suggestions
Get search suggestions for autocomplete

**Query Parameters:**
- `query` (string, required) - Partial query (min 2 chars)
- `limit` (1-50) - Max suggestions per category

```typescript
interface SearchSuggestionsResponse {
  titles: string[];
  reference_ids: string[];
  statuses: string[];
}
```

### Hierarchy & Navigation (`/api/v1/hierarchy`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | Get full hierarchy tree |
| GET | `/epics/:id` | Get epic hierarchy |
| GET | `/user-stories/:id` | Get user story hierarchy |
| GET | `/path/:entity_type/:id` | Get entity breadcrumb path |

---

## Comments System

### Comment Endpoints (`/api/v1/comments`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/:id` | Get comment |
| PUT | `/:id` | Update comment |
| DELETE | `/:id` | Delete comment |
| POST | `/:id/resolve` | Resolve comment |
| POST | `/:id/unresolve` | Unresolve comment |
| GET | `/status/:status` | Get comments by status |
| GET | `/:id/replies` | Get comment replies |
| POST | `/:id/replies` | Create reply |

### Entity Comments
Each entity supports comments:

**General Comments:**
- `GET /:entity_type/:id/comments` - Get entity comments
- `POST /:entity_type/:id/comments` - Create comment

**Inline Comments:**
- `POST /:entity_type/:id/comments/inline` - Create inline comment
- `GET /:entity_type/:id/comments/inline/visible` - Get visible inline comments
- `POST /:entity_type/:id/comments/inline/validate` - Validate inline comments

---

## Configuration Management

### Requirement Types (`/api/v1/config/requirement-types`)
- `POST /` - Create requirement type
- `GET /` - List requirement types
- `GET /:id` - Get requirement type
- `PUT /:id` - Update requirement type
- `DELETE /:id` - Delete requirement type

### Relationship Types (`/api/v1/config/relationship-types`)
- `POST /` - Create relationship type
- `GET /` - List relationship types
- `GET /:id` - Get relationship type
- `PUT /:id` - Update relationship type
- `DELETE /:id` - Delete relationship type

### Status Models (`/api/v1/config/status-models`)
- `POST /` - Create status model
- `GET /` - List status models
- `GET /:id` - Get status model
- `PUT /:id` - Update status model
- `DELETE /:id` - Delete status model
- `GET /default/:entity_type` - Get default status model
- `GET /:id/statuses` - List statuses by model
- `GET /:id/transitions` - List transitions by model

### Statuses (`/api/v1/config/statuses`)
- `POST /` - Create status
- `GET /:id` - Get status
- `PUT /:id` - Update status
- `DELETE /:id` - Delete status

### Status Transitions (`/api/v1/config/status-transitions`)
- `POST /` - Create status transition
- `GET /:id` - Get status transition
- `PUT /:id` - Update status transition
- `DELETE /:id` - Delete status transition

---

## TypeScript Interfaces

### Core Entity Types

```typescript
// User Types
interface User {
  id: string;
  username: string;
  email: string;
  role: 'Administrator' | 'User' | 'Commenter';
  created_at: string;
  updated_at: string;
}

interface UserResponse {
  id: string;
  username: string;
  email: string;
  role: 'Administrator' | 'User' | 'Commenter';
  created_at: string;
  updated_at: string;
}

// Authentication Types
interface LoginRequest {
  username: string;
  password: string;
}

interface LoginResponse {
  token: string;              // JWT access token for API requests
  refresh_token: string;      // Refresh token for obtaining new access tokens
  expires_at: string;         // Access token expiration timestamp (ISO-8601)
  user: UserResponse;         // Authenticated user information
}

interface RefreshRequest {
  refresh_token: string;      // Current refresh token
}

interface RefreshResponse {
  token: string;              // New JWT access token
  refresh_token: string;      // New refresh token (token rotation)
  expires_at: string;         // New access token expiration timestamp (ISO-8601)
}

interface LogoutRequest {
  refresh_token: string;      // Refresh token to invalidate
}

interface ErrorResponse {
  error: {
    code: string;             // Error code for programmatic handling
    message: string;          // Human-readable error message
  };
}

// Epic Types
interface Epic {
  id: string;
  reference_id: string;
  title: string;
  description?: string;
  status: 'Backlog' | 'Draft' | 'In Progress' | 'Done' | 'Cancelled';
  priority: 1 | 2 | 3 | 4;
  creator_id: string;
  assignee_id?: string;
  created_at: string;
  last_modified: string;
  
  // Optional populated fields
  creator?: User;
  assignee?: User;
  user_stories?: UserStory[];
  comments?: Comment[];
}

interface CreateEpicRequest {
  title: string;
  description?: string;
  priority: 1 | 2 | 3 | 4;
  assignee_id?: string;
}

interface UpdateEpicRequest {
  title?: string;
  description?: string;
  priority?: 1 | 2 | 3 | 4;
  assignee_id?: string;
}

// User Story Types
interface UserStory {
  id: string;
  reference_id: string;
  title: string;
  description?: string;
  status: 'Backlog' | 'Draft' | 'In Progress' | 'Done' | 'Cancelled';
  priority: 1 | 2 | 3 | 4;
  epic_id: string;
  creator_id: string;
  assignee_id?: string;
  created_at: string;
  last_modified: string;
  
  // Optional populated fields
  epic?: Epic;
  creator?: User;
  assignee?: User;
  acceptance_criteria?: AcceptanceCriteria[];
  requirements?: Requirement[];
  comments?: Comment[];
}

interface CreateUserStoryRequest {
  title: string;
  description?: string;
  priority: 1 | 2 | 3 | 4;
  epic_id: string;
  assignee_id?: string;
}

interface UpdateUserStoryRequest {
  title?: string;
  description?: string;
  priority?: 1 | 2 | 3 | 4;
  assignee_id?: string;
}

// Acceptance Criteria Types
interface AcceptanceCriteria {
  id: string;
  reference_id: string;
  description: string;
  user_story_id: string;
  author_id: string;
  created_at: string;
  last_modified: string;
  
  // Optional populated fields
  user_story?: UserStory;
  author?: User;
  requirements?: Requirement[];
  comments?: Comment[];
}

interface CreateAcceptanceCriteriaRequest {
  description: string;
  user_story_id: string;
}

interface UpdateAcceptanceCriteriaRequest {
  description?: string;
}

// Requirement Types
interface Requirement {
  id: string;
  reference_id: string;
  title: string;
  description?: string;
  status: 'Draft' | 'Active' | 'Obsolete';
  priority: 1 | 2 | 3 | 4;
  user_story_id: string;
  acceptance_criteria_id?: string;
  type_id: string;
  creator_id: string;
  assignee_id?: string;
  created_at: string;
  last_modified: string;
  
  // Optional populated fields
  user_story?: UserStory;
  acceptance_criteria?: AcceptanceCriteria;
  type?: RequirementType;
  creator?: User;
  assignee?: User;
  source_relationships?: RequirementRelationship[];
  target_relationships?: RequirementRelationship[];
  comments?: Comment[];
}

interface CreateRequirementRequest {
  title: string;
  description?: string;
  priority: 1 | 2 | 3 | 4;
  user_story_id: string;
  acceptance_criteria_id?: string;
  type_id: string;
  assignee_id?: string;
}

interface UpdateRequirementRequest {
  title?: string;
  description?: string;
  priority?: 1 | 2 | 3 | 4;
  assignee_id?: string;
}

// Comment Types
interface Comment {
  id: string;
  content: string;
  entity_type: 'epic' | 'user_story' | 'acceptance_criteria' | 'requirement';
  entity_id: string;
  author_id: string;
  parent_comment_id?: string;
  is_resolved: boolean;
  linked_text?: string;
  text_position_start?: number;
  text_position_end?: number;
  created_at: string;
  updated_at: string;
  
  // Optional populated fields
  author?: User;
  parent_comment?: Comment;
  replies?: Comment[];
}

interface CreateCommentRequest {
  content: string;
  parent_comment_id?: string;
}

interface CreateInlineCommentRequest {
  content: string;
  linked_text: string;
  text_position_start: number;
  text_position_end: number;
}

// Configuration Types
interface RequirementType {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

interface RelationshipType {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

interface RequirementRelationship {
  id: string;
  source_requirement_id: string;
  target_requirement_id: string;
  relationship_type_id: string;
  created_by: string;
  created_at: string;
  
  // Optional populated fields
  source_requirement?: Requirement;
  target_requirement?: Requirement;
  relationship_type?: RelationshipType;
  creator?: User;
}

interface CreateRelationshipRequest {
  source_requirement_id: string;
  target_requirement_id: string;
  relationship_type_id: string;
}

// Status Management Types
interface StatusModel {
  id: string;
  name: string;
  description?: string;
  entity_type: 'epic' | 'user_story' | 'acceptance_criteria' | 'requirement';
  is_default: boolean;
  created_at: string;
  updated_at: string;
  
  // Optional populated fields
  statuses?: Status[];
  transitions?: StatusTransition[];
}

interface Status {
  id: string;
  name: string;
  description?: string;
  color?: string;
  order: number;
  is_initial: boolean;
  is_final: boolean;
  status_model_id: string;
  created_at: string;
  updated_at: string;
  
  // Optional populated fields
  status_model?: StatusModel;
  from_transitions?: StatusTransition[];
  to_transitions?: StatusTransition[];
}

interface StatusTransition {
  id: string;
  name?: string;
  description?: string;
  from_status_id: string;
  to_status_id: string;
  status_model_id: string;
  created_at: string;
  updated_at: string;
  
  // Optional populated fields
  from_status?: Status;
  to_status?: Status;
  status_model?: StatusModel;
}

// Search Types
interface SearchResult {
  entity_type: 'epic' | 'user_story' | 'acceptance_criteria' | 'requirement';
  entity_id: string;
  reference_id: string;
  title: string;
  description?: string;
  highlight?: string;
  rank: number;
}

interface SearchResponse {
  results: SearchResult[];
  total_count: number;
  query: string;
  entity_types: string[];
  limit: number;
  offset: number;
}

// Hierarchy Types
interface HierarchyNode {
  entity_type: 'epic' | 'user_story' | 'acceptance_criteria' | 'requirement';
  entity_id: string;
  reference_id: string;
  title: string;
  status: string;
  children?: HierarchyNode[];
}

interface EntityPath {
  entity_type: 'epic' | 'user_story' | 'acceptance_criteria' | 'requirement';
  entity_id: string;
  reference_id: string;
  title: string;
}

// Deletion Types
interface DependencyInfo {
  can_delete: boolean;
  dependencies: {
    entity_type: string;
    entity_id: string;
    reference_id: string;
    title: string;
    dependency_type: string;
  }[];
  warnings: string[];
}

interface DeletionResult {
  success: boolean;
  deleted_entities: {
    entity_type: string;
    entity_id: string;
    reference_id: string;
  }[];
  message: string;
}

// List Response Types
interface ListResponse<T> {
  data: T[];
  total_count: number;
  limit: number;
  offset: number;
}

// Status Change Types
interface StatusChangeRequest {
  status: string;
}

interface AssignmentRequest {
  assignee_id?: string; // null to unassign
}
```

### API Response Types

```typescript
// Standard API Response
interface ApiResponse<T = any> {
  data?: T;
  error?: {
    code: string;
    message: string;
  };
}

// List responses
interface EpicListResponse extends ListResponse<Epic> {}
interface UserStoryListResponse extends ListResponse<UserStory> {}
interface AcceptanceCriteriaListResponse extends ListResponse<AcceptanceCriteria> {}
interface RequirementListResponse extends ListResponse<Requirement> {}
interface CommentListResponse extends ListResponse<Comment> {}

// Configuration list responses
interface RequirementTypeListResponse extends ListResponse<RequirementType> {}
interface RelationshipTypeListResponse extends ListResponse<RelationshipType> {}
interface StatusModelListResponse extends ListResponse<StatusModel> {}
interface StatusListResponse extends ListResponse<Status> {}
interface StatusTransitionListResponse extends ListResponse<StatusTransition> {}
```

---

## Error Handling

### HTTP Status Codes
- `200` - Success
- `201` - Created
- `400` - Bad Request (validation errors)
- `401` - Unauthorized (missing/invalid token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `409` - Conflict (deletion conflicts)
- `500` - Internal Server Error

### Error Response Format
```typescript
interface ErrorResponse {
  error: {
    code: string;
    message: string;
  };
}
```

### Common Error Codes
- `VALIDATION_ERROR` - Request validation failed
- `AUTHENTICATION_REQUIRED` - JWT token required
- `INSUFFICIENT_PERMISSIONS` - User lacks required permissions
- `ENTITY_NOT_FOUND` - Requested entity doesn't exist
- `DELETION_CONFLICT` - Entity has dependencies preventing deletion
- `INTERNAL_ERROR` - Server-side error

---

## Implementation Notes

### Authentication Flow
1. POST to `/auth/login` with credentials
2. Store JWT token from response
3. Include token in Authorization header for all subsequent requests
4. Handle 401 responses by redirecting to login

### Pagination
Most list endpoints support pagination:
- Use `limit` and `offset` parameters
- Default limit is usually 50, maximum 100
- Response includes `total_count` for pagination UI

### Including Related Data
Many endpoints support `include` parameter to populate related entities:
- `?include=creator,assignee` - Include user objects
- `?include=user_stories` - Include child entities
- `?include=comments` - Include comments

### Reference ID Support
Most endpoints accept either UUID or reference ID (e.g., "EP-001") in path parameters.

### Search Implementation
- Use `/api/v1/search/suggestions` for autocomplete
- Use `/api/v1/search` for full search results
- Support entity type filtering for scoped searches

### Real-time Updates
Consider implementing WebSocket connections for real-time updates to comments and status changes.

### Caching Strategy
- Cache configuration data (requirement types, relationship types)
- Cache user information
- Implement cache invalidation for entity updates

This documentation provides a complete foundation for implementing a web UI client for the Product Requirements Management API.