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
All API endpoints (except login) require a JWT token in the Authorization header:
```
Authorization: Bearer <jwt_token>
```

### Auth Endpoints

#### POST /auth/login
```typescript
interface LoginRequest {
  username: string;
  password: string;
}

interface LoginResponse {
  token: string;
  expires_at: string;
  user: UserResponse;
}
```

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
  creator_id: string;
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

interface UpdateCommentRequest {
  content: string;
}

// Inline Comment Validation Types
interface InlineCommentValidationRequest {
  comments: InlineCommentPosition[];
}

interface InlineCommentPosition {
  comment_id: string;
  text_position_start: number;
  text_position_end: number;
}

// Enhanced Comment System Types
interface CommentThread {
  parent_comment: Comment;
  replies: Comment[];
  total_replies: number;
}

interface InlineCommentContext {
  linked_text: string;
  text_position_start: number;
  text_position_end: number;
  is_valid: boolean;
}

interface CommentResolution {
  is_resolved: boolean;
  resolved_by?: User;
  resolved_at?: string;
  resolution_note?: string;
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

// Configuration Management Request Types
interface CreateRequirementTypeRequest {
  name: string;
  description?: string;
}

interface UpdateRequirementTypeRequest {
  name?: string;
  description?: string;
}

interface CreateRelationshipTypeRequest {
  name: string;
  description?: string;
}

interface UpdateRelationshipTypeRequest {
  name?: string;
  description?: string;
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

// Deletion Workflow Types
interface DependencyInfo {
  can_delete: boolean;
  dependencies: DependencyItem[];
  warnings: string[];
}

interface DependencyItem {
  entity_type: string;
  entity_id: string;
  reference_id: string;
  title: string;
  dependency_type: string; // 'child', 'reference', 'relationship'
}

interface DeletionResult {
  success: boolean;
  deleted_entities: DeletedEntity[];
  message: string;
}

interface DeletedEntity {
  entity_type: string;
  entity_id: string;
  reference_id: string;
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

// Health Check Types
interface HealthCheckResponse {
  status: string;
  reason?: string;
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
interface UserListResponse extends ListResponse<User> {}
interface EpicListResponse extends ListResponse<Epic> {}
interface UserStoryListResponse extends ListResponse<UserStory> {}
interface AcceptanceCriteriaListResponse extends ListResponse<AcceptanceCriteria> {}
interface RequirementListResponse extends ListResponse<Requirement> {}
interface CommentListResponse extends ListResponse<Comment> {}

// Configuration list responses (standardized format)
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

### Enhanced Error Types
```typescript
interface ValidationError {
  field: string;
  message: string;
  code: string;
}

interface ValidationErrorResponse extends ErrorResponse {
  error: {
    code: 'VALIDATION_ERROR';
    message: string;
    validation_errors: ValidationError[];
  };
}

interface DeletionConflictResponse extends ErrorResponse {
  error: {
    code: 'DELETION_CONFLICT';
    message: string;
    dependencies: DependencyItem[];
  };
}
```

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