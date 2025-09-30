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
8. [Deletion Workflows](#deletion-workflows)
9. [Configuration Management](#configuration-management)
10. [Implementation Notes](#implementation-notes)

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
- `GET /auth/users` - List users with pagination
- `GET /auth/users/:id` - Get user by ID
- `PUT /auth/users/:id` - Update user
- `DELETE /auth/users/:id` - Delete user

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
| GET | `/:id/comments` | Get epic comments |
| POST | `/:id/comments` | Create epic comment |
| POST | `/:id/comments/inline` | Create epic inline comment |
| GET | `/:id/comments/inline/visible` | Get visible epic inline comments |
| POST | `/:id/comments/inline/validate` | Validate epic inline comments |

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
| GET | `/:id/comments` | Get user story comments |
| POST | `/:id/comments` | Create user story comment |
| POST | `/:id/comments/inline` | Create user story inline comment |
| GET | `/:id/comments/inline/visible` | Get visible user story inline comments |
| POST | `/:id/comments/inline/validate` | Validate user story inline comments |

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
| GET | `/:id/comments` | Get acceptance criteria comments |
| POST | `/:id/comments` | Create acceptance criteria comment |
| POST | `/:id/comments/inline` | Create acceptance criteria inline comment |
| GET | `/:id/comments/inline/visible` | Get visible acceptance criteria inline comments |
| POST | `/:id/comments/inline/validate` | Validate acceptance criteria inline comments |

**Query Parameters for List:**
- `user_story_id` (UUID) - Filter by user story
- `author_id` (UUID) - Filter by author
- `order_by` (string) - Sort order
- `limit` (1-100) - Page size
- `offset` (number) - Pagination offset

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
| GET | `/:id/comments` | Get requirement comments |
| POST | `/:id/comments` | Create requirement comment |
| POST | `/:id/comments/inline` | Create requirement inline comment |
| GET | `/:id/comments/inline/visible` | Get visible requirement inline comments |
| POST | `/:id/comments/inline/validate` | Validate requirement inline comments |

**Query Parameters for List:**
- `user_story_id` (UUID) - Filter by user story
- `acceptance_criteria_id` (UUID) - Filter by acceptance criteria
- `type_id` (UUID) - Filter by requirement type
- `creator_id` (UUID) - Filter by creator
- `assignee_id` (UUID) - Filter by assignee
- `status` (RequirementStatus) - Filter by status
- `priority` (1-4) - Filter by priority
- `order_by` (string) - Sort order
- `limit` (1-100) - Page size
- `offset` (number) - Pagination offset
- `include` (string) - Include related data: `user_story,acceptance_criteria,type,creator,assignee,source_relationships,target_relationships,comments`

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
| GET | `/:id` | Get comment by ID |
| PUT | `/:id` | Update comment content |
| DELETE | `/:id` | Delete comment (cascade deletes replies) |
| POST | `/:id/resolve` | Mark comment as resolved |
| POST | `/:id/unresolve` | Mark comment as unresolved |
| GET | `/status/:status` | Get comments by status (resolved/unresolved) |
| GET | `/:id/replies` | Get comment replies with pagination |
| POST | `/:id/replies` | Create reply to comment |

### Entity Comments
Each entity (Epic, User Story, Acceptance Criteria, Requirement) supports comments:

**General Comments:**
- `GET /:entity_type/:id/comments` - Get entity comments with pagination
- `POST /:entity_type/:id/comments` - Create general comment on entity

**Inline Comments:**
- `POST /:entity_type/:id/comments/inline` - Create inline comment linked to specific text
- `GET /:entity_type/:id/comments/inline/visible` - Get visible inline comments
- `POST /:entity_type/:id/comments/inline/validate` - Validate inline comment positions

### Comment Threading
Comments support parent-child relationships for threaded discussions:
- Use `parent_comment_id` when creating replies
- Replies are automatically linked to parent comments
- Deleting a parent comment cascades to all replies

---

## Deletion Workflows

The API provides comprehensive deletion workflows with dependency validation to ensure safe entity removal.

### Deletion Validation Endpoints

Each entity type supports deletion validation:
- `GET /api/v1/epics/:id/validate-deletion`
- `GET /api/v1/user-stories/:id/validate-deletion`
- `GET /api/v1/acceptance-criteria/:id/validate-deletion`
- `GET /api/v1/requirements/:id/validate-deletion`

### Comprehensive Deletion Endpoints

Each entity type supports comprehensive deletion with cascade operations:
- `DELETE /api/v1/epics/:id/delete`
- `DELETE /api/v1/user-stories/:id/delete`
- `DELETE /api/v1/acceptance-criteria/:id/delete`
- `DELETE /api/v1/requirements/:id/delete`

### General Deletion Confirmation

For flexible deletion validation across entity types:
- `GET /api/v1/deletion/confirm?entity_type=epic&id=uuid`

### Deletion Workflow Process

1. **Validate Deletion**: Call validate-deletion endpoint to check dependencies
2. **Review Dependencies**: Present dependency information to user
3. **Confirm Deletion**: If acceptable, call comprehensive deletion endpoint
4. **Handle Results**: Process deletion results and update UI accordingly

### Dependency Types

- `child` - Direct child entities (e.g., User Stories under Epic)
- `reference` - Entities that reference this entity
- `relationship` - Requirement relationships that would be broken

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

**Note**: All configuration endpoints require Administrator role and proper authentication.

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
  updated_at: string;
  
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
  updated_at: string;
  
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
  updated_at: string;
  
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
  updated_at: string;
  
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

interface ValidationResponse {
  valid: boolean;
  errors: string[];
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

// Status Management Request Types
interface CreateStatusModelRequest {
  name: string;
  description?: string;
  entity_type: 'epic' | 'user_story' | 'acceptance_criteria' | 'requirement';
  is_default?: boolean;
}

interface UpdateStatusModelRequest {
  name?: string;
  description?: string;
  is_default?: boolean;
}

interface CreateStatusRequest {
  name: string;
  description?: string;
  color?: string; // Hex color code (e.g., '#FF5733')
  order: number;
  is_initial?: boolean;
  is_final?: boolean;
  status_model_id: string;
}

interface UpdateStatusRequest {
  name?: string;
  description?: string;
  color?: string; // Hex color code (e.g., '#FF5733')
  order?: number;
  is_initial?: boolean;
  is_final?: boolean;
}

interface CreateStatusTransitionRequest {
  name?: string;
  description?: string;
  from_status_id: string;
  to_status_id: string;
  status_model_id: string;
}

interface UpdateStatusTransitionRequest {
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

// Enhanced Deletion Workflow Types
interface DeletionValidation {
  entity_type: string;
  entity_id: string;
  can_delete: boolean;
  blocking_dependencies: DependencyItem[];
  cascade_dependencies: DependencyItem[];
  warnings: string[];
}

interface DeletionPlan {
  primary_entity: EntityReference;
  cascade_deletions: EntityReference[];
  dependency_updates: DependencyUpdate[];
  estimated_impact: number;
}

interface EntityReference {
  entity_type: string;
  entity_id: string;
  reference_id: string;
  title: string;
}

interface DependencyUpdate {
  entity_type: string;
  entity_id: string;
  field: string;
  action: 'nullify' | 'cascade' | 'restrict';
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
  status: 'ok' | 'error';
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

interface AuthenticationErrorResponse extends ErrorResponse {
  error: {
    code: 'AUTHENTICATION_REQUIRED' | 'INSUFFICIENT_PERMISSIONS';
    message: string;
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

### Deletion Workflow Implementation
1. **Pre-deletion Validation**:
   ```typescript
   const validation = await api.get(`/api/v1/epics/${id}/validate-deletion`);
   if (!validation.can_delete) {
     // Show dependencies and warnings to user
     showDeletionConflicts(validation.dependencies, validation.warnings);
     return;
   }
   ```

2. **Comprehensive Deletion**:
   ```typescript
   const result = await api.delete(`/api/v1/epics/${id}/delete`);
   if (result.success) {
     // Update UI to reflect deleted entities
     updateUIAfterDeletion(result.deleted_entities);
   }
   ```

### Comment System Implementation
1. **Inline Comments**:
   ```typescript
   // Create inline comment
   const inlineComment = await api.post(`/api/v1/epics/${id}/comments/inline`, {
     content: "This needs clarification",
     linked_text: "user authentication",
     text_position_start: 45,
     text_position_end: 63
   });
   
   // Validate inline comments after content changes
   const validation = await api.post(`/api/v1/epics/${id}/comments/inline/validate`, {
     comments: [
       {
         comment_id: inlineComment.id,
         text_position_start: 45,
         text_position_end: 63
       }
     ]
   });
   ```

2. **Comment Threading**:
   ```typescript
   // Create reply to comment
   const reply = await api.post(`/api/v1/comments/${parentId}/replies`, {
     content: "I agree with this point"
   });
   
   // Get all replies
   const replies = await api.get(`/api/v1/comments/${parentId}/replies`);
   ```

### Authentication & Authorization
- All endpoints except `/auth/login`, `/ready`, and `/live` require JWT authentication
- Configuration endpoints require Administrator role
- Use `x-required-role` extension in OpenAPI for role-based access control
- Handle 401 (Unauthorized) and 403 (Forbidden) responses appropriately

### Response Format Consistency
All list endpoints now use standardized `ListResponse` format:
```typescript
interface ListResponse<T> {
  data: T[];
  total_count: number;
  limit: number;
  offset: number;
}
```

This ensures consistent pagination handling across all entity types.

This documentation provides a complete foundation for implementing a web UI client for the Product Requirements Management API with full support for deletion workflows, comprehensive comment system, and all documented endpoints.
---


## 📚 Complete Documentation Suite

For the most up-to-date and comprehensive API documentation, visit our generated documentation hub:

### Interactive Documentation
- **[Documentation Hub](../docs/generated/index.html)** - Central access point for all documentation formats
- **[Swagger UI](../docs/generated/swagger-ui.html)** - Interactive API explorer with live testing
- **[Developer Guide](../docs/generated/developer-guide.md)** - Comprehensive integration guide with examples

### Reference Documentation
- **[HTML Documentation](../docs/generated/api-documentation.html)** - Complete API reference in HTML format
- **[Markdown Documentation](../docs/generated/api-documentation.md)** - API reference in Markdown format
- **[TypeScript Interfaces](../docs/generated/api-types.ts)** - Complete TypeScript type definitions
- **[JSON Schema](../docs/generated/api-documentation.json)** - Machine-readable API documentation

### Source Specifications
- **[OpenAPI Specification](../docs/openapi-v3.yaml)** - Complete OpenAPI 3.0.3 specification
- **[Swagger JSON](../docs/swagger.json)** - Generated Swagger documentation

### Quick Access Commands
```bash
# Generate all documentation formats
make docs-generate

# Generate specific formats
make docs-generate-html
make docs-generate-markdown  
make docs-generate-typescript
make docs-generate-json

# Generate interactive Swagger UI
make swagger

# Serve documentation locally
make swagger-serve
```

### Documentation Features
- **Interactive Testing**: Live API exploration with Swagger UI
- **Multiple Formats**: HTML, Markdown, TypeScript, JSON for different use cases
- **Complete Coverage**: All 80+ endpoints documented with examples
- **Type Safety**: Full TypeScript interface definitions
- **Client Examples**: Implementation examples in multiple languages
- **Best Practices**: Security, error handling, and performance guidance
- **Real-time Updates**: Documentation generated from OpenAPI specification