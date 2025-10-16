# Product Requirements Management API

Comprehensive API for managing product requirements through hierarchical structure of Epics → User Stories → Requirements. 
Features include full-text search, comment system, relationship mapping, and configurable workflows.


**Version:** 1.0.0

## Table of Contents

- [Acceptance Criteria](# )
- [Authentication](# )
- [Comments](# )
- [Configuration](# )
- [Deletion](# )
- [Epics](# )
- [Health](# )
- [Navigation](# )
- [Requirements](# )
- [Search](# )
- [Steering Documents](# )
- [User Management](# )
- [User Stories](# )


## Base URLs

- **Development server**: http://localhost:8080
- **Production server**: https://api.requirements.example.com


## Authentication

This API uses JWT Bearer token authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your_jwt_token>
```

## Acceptance Criteria

### POST /api/v1/acceptance-criteria/{id}/comments/inline

Create acceptance criteria inline comment





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Inline comment created |




---

### POST /api/v1/acceptance-criteria/{id}/comments/inline/validate

Validate acceptance criteria inline comments





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Validation result |




---

### GET /api/v1/acceptance-criteria/{id}

Get acceptance criteria by ID







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Acceptance criteria details |




---

### PUT /api/v1/acceptance-criteria/{id}

Update acceptance criteria





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Acceptance criteria updated |




---

### DELETE /api/v1/acceptance-criteria/{id}

Delete acceptance criteria







#### Responses

| Status Code | Description |
|-------------|-------------|
| 204 | Acceptance criteria deleted |




---

### DELETE /api/v1/acceptance-criteria/{id}/delete

Comprehensive acceptance criteria deletion

Delete acceptance criteria with all dependencies and cascade operations






#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Deletion completed |
| 409 | Cannot delete due to dependencies |




---

### GET /api/v1/acceptance-criteria/{id}/comments/inline/visible

Get visible acceptance criteria inline comments







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of visible inline comments |




---

### GET /api/v1/acceptance-criteria

List acceptance criteria



#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
| user_story_id | query | map[format:uuid type:string] | No |  |
| author_id | query | map[format:uuid type:string] | No |  |
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of acceptance criteria |




---

### GET /api/v1/acceptance-criteria/{id}/comments

Get acceptance criteria comments



#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of acceptance criteria comments |




---

### POST /api/v1/acceptance-criteria/{id}/comments

Create acceptance criteria comment





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Comment created |




---

### GET /api/v1/acceptance-criteria/{id}/validate-deletion

Validate acceptance criteria deletion

Check if acceptance criteria can be deleted and get dependency information






#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Deletion validation result |
| 404 |  |




---

## Authentication

### GET /auth/users

List users (Admin only)

Get list of all users


#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of users |
| 401 |  |
| 403 |  |




---

### POST /auth/users

Create user (Admin only)

Create a new user account




#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | User created successfully |
| 400 |  |
| 401 |  |
| 403 |  |




---

### POST /auth/change-password

Change user password

Change the password for the currently authenticated user




#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Password changed successfully |
| 400 | Invalid request |


#### Security

- **BearerAuth**: No specific scopes


---

### GET /auth/profile

Get current user profile

Get the profile information of the currently authenticated user






#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | User profile |


#### Security

- **BearerAuth**: No specific scopes


---

### POST /auth/login

User login

Authenticate user and receive JWT token




#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Login successful |
| 401 | Invalid credentials |




---

### GET /auth/users/{id}

Get user by ID (Admin only)

Get user details by ID






#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | User details |
| 401 |  |
| 403 |  |
| 404 |  |




---

### PUT /auth/users/{id}

Update user (Admin only)

Update user information




#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | User updated successfully |
| 400 |  |
| 401 |  |
| 403 |  |
| 404 |  |




---

### DELETE /auth/users/{id}

Delete user (Admin only)

Delete user account






#### Responses

| Status Code | Description |
|-------------|-------------|
| 204 | User deleted successfully |
| 401 |  |
| 403 |  |
| 404 |  |




---

## Comments

### POST /api/v1/comments/{id}/unresolve

Unresolve comment

Mark a previously resolved comment as unresolved, reopening the discussion






#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Comment unresolved successfully |
| 404 |  |




---

### GET /api/v1/comments/status/{status}

Get comments by status

Retrieve comments filtered by their resolution status across all entities


#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
| status | path | map[enum:[resolved unresolved] type:string] | **Yes** | Filter comments by resolution status |
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of comments filtered by status |
| 400 |  |




---

### GET /api/v1/comments/{id}/replies

Get comment replies

Retrieve all direct replies to a specific comment with pagination support. 
Returns replies in chronological order (oldest first) to maintain conversation flow. 
Each reply includes author information and metadata for building threaded comment interfaces.

**Threading Behavior:**
- Only returns direct replies (depth = 1 from parent)
- For nested replies, call this endpoint recursively with each reply's ID
- Replies inherit the same entity context as their parent comment

**Use Cases:**
- Building threaded comment interfaces
- Loading conversation threads on demand
- Implementing expandable comment sections



#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of comment replies retrieved successfully |
| 400 | Invalid parent comment ID format |
| 404 | Parent comment not found |
| 500 |  |




---

### POST /api/v1/comments/{id}/replies

Create comment reply

Create a new reply to an existing comment, automatically inheriting the parent's entity context for threaded discussions.

**Automatic Context Inheritance:**
- Entity type and ID are inherited from parent comment
- Parent-child relationship is automatically established
- Reply depth is calculated based on parent's depth

**Threading Rules:**
- Replies can be nested to any depth
- Each reply maintains reference to its direct parent
- All replies in a thread share the same entity context

**Required Fields:**
- Only `content` and `author_id` are required
- Entity context is inherited automatically
- Parent relationship is established via URL parameter





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Reply created successfully with parent-child relationship established |
| 400 | Invalid request - bad parent ID format, missing required fields, or empty content |
| 404 | Parent comment not found |
| 500 |  |




---

### GET /api/v1/epics/{id}/comments/inline/visible

Get visible epic inline comments







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of visible inline comments |




---

### POST /api/v1/acceptance-criteria/{id}/comments/inline

Create acceptance criteria inline comment





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Inline comment created |




---

### POST /api/v1/epics/{id}/comments/inline/validate

Validate epic inline comments

Validate that inline comment positions are still valid against the current epic content




#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Validation completed successfully |
| 400 |  |
| 404 |  |




---

### GET /api/v1/user-stories/{id}/comments/inline/visible

Get visible user story inline comments







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of visible inline comments |




---

### POST /api/v1/comments/{id}/resolve

Resolve comment

Mark a comment as resolved, indicating that the issue or question has been addressed






#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Comment resolved successfully |
| 404 |  |




---

### GET /api/v1/requirements/{id}/comments/inline/visible

Get visible requirement inline comments







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of visible inline comments |




---

### POST /api/v1/epics/{id}/comments/inline

Create epic inline comment





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Inline comment created |




---

### POST /api/v1/user-stories/{id}/comments/inline

Create user story inline comment





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Inline comment created |




---

### POST /api/v1/user-stories/{id}/comments/inline/validate

Validate user story inline comments





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Validation result |




---

### POST /api/v1/acceptance-criteria/{id}/comments/inline/validate

Validate acceptance criteria inline comments





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Validation result |




---

### GET /api/v1/requirements/{id}/comments

Get requirement comments



#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of requirement comments |




---

### POST /api/v1/requirements/{id}/comments

Create requirement comment





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Comment created |




---

### GET /api/v1/epics/{id}/comments

Get epic comments

Retrieve all comments associated with a specific epic, including both general and inline comments


#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of epic comments retrieved successfully |
| 404 |  |




---

### POST /api/v1/epics/{id}/comments

Create epic comment

Create a new general comment on an epic




#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Comment created successfully |
| 400 |  |
| 404 |  |




---

### GET /api/v1/acceptance-criteria/{id}/comments/inline/visible

Get visible acceptance criteria inline comments







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of visible inline comments |




---

### POST /api/v1/requirements/{id}/comments/inline/validate

Validate requirement inline comments





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Validation result |




---

### GET /api/v1/comments/{id}

Get comment by ID

Retrieve a specific comment with its details and optional populated fields






#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Comment details retrieved successfully |
| 404 |  |




---

### PUT /api/v1/comments/{id}

Update comment

Update the content of an existing comment




#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Comment updated successfully |
| 400 |  |
| 404 |  |




---

### DELETE /api/v1/comments/{id}

Delete comment

Delete a comment and all its replies (cascade deletion)






#### Responses

| Status Code | Description |
|-------------|-------------|
| 204 | Comment deleted successfully |
| 404 |  |




---

### GET /api/v1/acceptance-criteria/{id}/comments

Get acceptance criteria comments



#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of acceptance criteria comments |




---

### POST /api/v1/acceptance-criteria/{id}/comments

Create acceptance criteria comment





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Comment created |




---

### POST /api/v1/requirements/{id}/comments/inline

Create requirement inline comment





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Inline comment created |




---

### GET /api/v1/user-stories/{id}/comments

Get user story comments



#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of user story comments |




---

### POST /api/v1/user-stories/{id}/comments

Create user story comment





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Comment created |




---

## Configuration

### POST /api/v1/config/status-transitions

Create status transition





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Status transition created |


#### Security

- **BearerAuth**: No specific scopes


---

### GET /api/v1/config/requirement-types

List requirement types







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of requirement types |
| 401 |  |
| 403 |  |


#### Security

- **BearerAuth**: No specific scopes


---

### POST /api/v1/config/requirement-types

Create requirement type





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Requirement type created |
| 400 |  |
| 401 |  |
| 403 |  |


#### Security

- **BearerAuth**: No specific scopes


---

### GET /api/v1/config/status-models

List status models







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of status models |


#### Security

- **BearerAuth**: No specific scopes


---

### POST /api/v1/config/status-models

Create status model





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Status model created |


#### Security

- **BearerAuth**: No specific scopes


---

### POST /api/v1/config/statuses

Create status





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Status created |


#### Security

- **BearerAuth**: No specific scopes


---

### GET /api/v1/config/status-models/{id}

Get status model by ID







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Status model details |
| 404 |  |


#### Security

- **BearerAuth**: No specific scopes


---

### PUT /api/v1/config/status-models/{id}

Update status model





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Status model updated |


#### Security

- **BearerAuth**: No specific scopes


---

### DELETE /api/v1/config/status-models/{id}

Delete status model







#### Responses

| Status Code | Description |
|-------------|-------------|
| 204 | Status model deleted |


#### Security

- **BearerAuth**: No specific scopes


---

### GET /api/v1/config/statuses/{id}

Get status by ID







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Status details |
| 404 |  |


#### Security

- **BearerAuth**: No specific scopes


---

### PUT /api/v1/config/statuses/{id}

Update status





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Status updated |


#### Security

- **BearerAuth**: No specific scopes


---

### DELETE /api/v1/config/statuses/{id}

Delete status







#### Responses

| Status Code | Description |
|-------------|-------------|
| 204 | Status deleted |


#### Security

- **BearerAuth**: No specific scopes


---

### GET /api/v1/config/relationship-types

List relationship types







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of relationship types |


#### Security

- **BearerAuth**: No specific scopes


---

### POST /api/v1/config/relationship-types

Create relationship type





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Relationship type created |


#### Security

- **BearerAuth**: No specific scopes


---

### GET /api/v1/config/relationship-types/{id}

Get relationship type by ID







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Relationship type details |
| 404 |  |


#### Security

- **BearerAuth**: No specific scopes


---

### PUT /api/v1/config/relationship-types/{id}

Update relationship type





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Relationship type updated |


#### Security

- **BearerAuth**: No specific scopes


---

### DELETE /api/v1/config/relationship-types/{id}

Delete relationship type







#### Responses

| Status Code | Description |
|-------------|-------------|
| 204 | Relationship type deleted |


#### Security

- **BearerAuth**: No specific scopes


---

### GET /api/v1/config/status-models/{id}/statuses

List statuses by model







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of statuses |


#### Security

- **BearerAuth**: No specific scopes


---

### GET /api/v1/config/status-models/{id}/transitions

List status transitions by model







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of status transitions |


#### Security

- **BearerAuth**: No specific scopes


---

### GET /api/v1/config/status-transitions/{id}

Get status transition by ID







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Status transition details |
| 404 |  |


#### Security

- **BearerAuth**: No specific scopes


---

### PUT /api/v1/config/status-transitions/{id}

Update status transition





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Status transition updated |


#### Security

- **BearerAuth**: No specific scopes


---

### DELETE /api/v1/config/status-transitions/{id}

Delete status transition







#### Responses

| Status Code | Description |
|-------------|-------------|
| 204 | Status transition deleted |


#### Security

- **BearerAuth**: No specific scopes


---

### GET /api/v1/config/requirement-types/{id}

Get requirement type by ID







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Requirement type details |
| 404 |  |


#### Security

- **BearerAuth**: No specific scopes


---

### PUT /api/v1/config/requirement-types/{id}

Update requirement type





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Requirement type updated |


#### Security

- **BearerAuth**: No specific scopes


---

### DELETE /api/v1/config/requirement-types/{id}

Delete requirement type







#### Responses

| Status Code | Description |
|-------------|-------------|
| 204 | Requirement type deleted |


#### Security

- **BearerAuth**: No specific scopes


---

### GET /api/v1/config/status-models/default/{entity_type}

Get default status model for entity type



#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
| entity_type | path | map[$ref:#/components/schemas/EntityType] | **Yes** |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Default status model |


#### Security

- **BearerAuth**: No specific scopes


---

## Deletion

### GET /api/v1/deletion/confirm

Get deletion confirmation

Get deletion validation information for any entity type using query parameters


#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
| entity_type | query | map[enum:[epic user_story acceptance_criteria requirement] type:string] | **Yes** | Type of entity to validate deletion for |
| id | query | map[format:uuid type:string] | **Yes** | Entity ID to validate deletion for |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Deletion validation result |
| 400 | Bad request - missing or invalid parameters |
| 404 | Entity not found |
| 500 | Internal server error |




---

## Epics

### GET /api/v1/epics/{id}

Get epic by ID







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Epic details |
| 404 |  |




---

### PUT /api/v1/epics/{id}

Update epic





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Epic updated |




---

### DELETE /api/v1/epics/{id}

Delete epic







#### Responses

| Status Code | Description |
|-------------|-------------|
| 204 | Epic deleted |
| 409 | Cannot delete due to dependencies |




---

### GET /api/v1/epics/{id}/comments/inline/visible

Get visible epic inline comments







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of visible inline comments |




---

### POST /api/v1/epics/{id}/comments/inline/validate

Validate epic inline comments

Validate that inline comment positions are still valid against the current epic content




#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Validation completed successfully |
| 400 |  |
| 404 |  |




---

### PATCH /api/v1/epics/{id}/assign

Assign epic to user





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Epic assigned |




---

### GET /api/v1/epics

List epics



#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |
| status | query | map[$ref:#/components/schemas/EpicStatus] | No |  |
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of epics |
| 401 |  |




---

### POST /api/v1/epics

Create epic





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Epic created |
| 400 |  |
| 401 |  |




---

### DELETE /api/v1/epics/{id}/delete

Comprehensive epic deletion

Delete epic with all dependencies and cascade operations






#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Deletion completed |
| 409 | Cannot delete due to dependencies |




---

### POST /api/v1/epics/{id}/comments/inline

Create epic inline comment





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Inline comment created |




---

### GET /api/v1/epics/{id}/validate-deletion

Validate epic deletion

Check if epic can be deleted and get dependency information






#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Deletion validation result |
| 404 |  |




---

### POST /api/v1/epics/{epic_id}/steering-documents/{doc_id}

Link a steering document to an epic

Create a link between a steering document and an epic. Both entities must exist. Administrators can link any document, Users can only link their own documents.






#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Successfully linked steering document to epic |
| 400 |  |
| 401 |  |
| 403 |  |
| 404 |  |
| 409 | Link already exists |
| 500 |  |


#### Security

- **BearerAuth**: No specific scopes


---

### DELETE /api/v1/epics/{epic_id}/steering-documents/{doc_id}

Unlink a steering document from an epic

Remove the link between a steering document and an epic. Administrators can unlink any document, Users can only unlink their own documents.






#### Responses

| Status Code | Description |
|-------------|-------------|
| 204 | Successfully unlinked steering document from epic |
| 400 |  |
| 401 |  |
| 403 |  |
| 404 |  |
| 500 |  |


#### Security

- **BearerAuth**: No specific scopes


---

### PATCH /api/v1/epics/{id}/status

Change epic status





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Status changed |




---

### GET /api/v1/epics/{id}/steering-documents

Get steering documents linked to an epic

Retrieve all steering documents that are linked to a specific epic. Returns an array of steering documents associated with the epic.






#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Successfully retrieved steering documents for epic |
| 400 |  |
| 401 |  |
| 404 |  |
| 500 |  |


#### Security

- **BearerAuth**: No specific scopes


---

### GET /api/v1/epics/{id}/user-stories

Get epic with user stories







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Epic with user stories |




---

### POST /api/v1/epics/{id}/user-stories

Create user story in epic





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | User story created |




---

### GET /api/v1/epics/{id}/comments

Get epic comments

Retrieve all comments associated with a specific epic, including both general and inline comments


#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of epic comments retrieved successfully |
| 404 |  |




---

### POST /api/v1/epics/{id}/comments

Create epic comment

Create a new general comment on an epic




#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Comment created successfully |
| 400 |  |
| 404 |  |




---

## Health

### GET /live

Liveness check

Check if the application is alive






#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Application is alive |




---

### GET /ready

Readiness check

Check if the application is ready to serve requests






#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Application is ready |
| 503 | Application is not ready |




---

## Navigation

### GET /api/v1/hierarchy/epics/{id}

Get epic hierarchy



#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
| id | path | map[type:string] | **Yes** | Epic ID (UUID or reference ID) |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Epic hierarchy |




---

### GET /api/v1/hierarchy

Get full hierarchy







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Complete hierarchy tree |




---

### GET /api/v1/hierarchy/path/{entity_type}/{id}

Get entity breadcrumb path



#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
| entity_type | path | map[$ref:#/components/schemas/EntityType] | **Yes** |  |
| id | path | map[type:string] | **Yes** |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Entity path |




---

### GET /api/v1/hierarchy/user-stories/{id}

Get user story hierarchy



#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
| id | path | map[type:string] | **Yes** | User story ID (UUID or reference ID) |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | User story hierarchy |




---

## Requirements

### POST /api/v1/requirements/relationships

Create requirement relationship





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Relationship created |




---

### DELETE /api/v1/requirements/{id}/delete

Comprehensive requirement deletion

Delete requirement with all dependencies and cascade operations






#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Deletion completed |
| 409 | Cannot delete due to dependencies |




---

### GET /api/v1/requirements/{id}/comments/inline/visible

Get visible requirement inline comments







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of visible inline comments |




---

### GET /api/v1/requirements/search

Search requirements



#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
| q | query | map[type:string] | **Yes** | Search query |
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Search results |




---

### GET /api/v1/requirements/{id}/relationships

Get requirement with relationships







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Requirement with relationships |




---

### DELETE /api/v1/requirement-relationships/{id}

Delete requirement relationship



#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
| id | path | map[format:uuid type:string] | **Yes** |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 204 | Relationship deleted |
| 404 |  |




---

### PATCH /api/v1/requirements/{id}/status

Change requirement status





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Status changed |




---

### GET /api/v1/requirements/{id}/validate-deletion

Validate requirement deletion

Check if requirement can be deleted and get dependency information






#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Deletion validation result |
| 404 |  |




---

### GET /api/v1/requirements/{id}/comments

Get requirement comments



#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of requirement comments |




---

### POST /api/v1/requirements/{id}/comments

Create requirement comment





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Comment created |




---

### POST /api/v1/requirements/{id}/comments/inline/validate

Validate requirement inline comments





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Validation result |




---

### POST /api/v1/requirements/{id}/comments/inline

Create requirement inline comment





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Inline comment created |




---

### GET /api/v1/requirements/{id}

Get requirement by ID







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Requirement details |




---

### PUT /api/v1/requirements/{id}

Update requirement





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Requirement updated |




---

### DELETE /api/v1/requirements/{id}

Delete requirement







#### Responses

| Status Code | Description |
|-------------|-------------|
| 204 | Requirement deleted |




---

### GET /api/v1/requirements

List requirements



#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
| user_story_id | query | map[format:uuid type:string] | No |  |
| acceptance_criteria_id | query | map[format:uuid type:string] | No |  |
| type_id | query | map[format:uuid type:string] | No |  |
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |
| status | query | map[$ref:#/components/schemas/RequirementStatus] | No |  |
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of requirements |




---

### POST /api/v1/requirements

Create requirement





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Requirement created |




---

### PATCH /api/v1/requirements/{id}/assign

Assign requirement to user





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Requirement assigned |




---

## Search

### GET /api/v1/search

Global search



#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
| q | query | map[type:string] | **Yes** | Search query |
| entity_types | query | map[type:string] | No | Comma-separated entity types to search |
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Search results |
| 401 |  |




---

### GET /api/v1/search/suggestions

Get search suggestions



#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
| query | query | map[minLength:2 type:string] | **Yes** | Partial search query |
| limit | query | map[default:10 maximum:50 minimum:1 type:integer] | No | Maximum suggestions per category |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Search suggestions |




---

## Steering Documents

### GET /api/v1/steering-documents/{id}

Get a steering document by ID or reference ID

Retrieve a single steering document by its UUID or reference ID (e.g., STD-001). Supports both formats for flexible access. Requires authentication.






#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Steering document found successfully |
| 401 |  |
| 404 |  |
| 500 |  |


#### Security

- **BearerAuth**: No specific scopes


---

### PUT /api/v1/steering-documents/{id}

Update an existing steering document

Update a steering document's properties. Only provided fields will be updated. Administrators can update any document, Users can only update their own documents.




#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Steering document updated successfully |
| 400 |  |
| 401 |  |
| 403 |  |
| 404 |  |
| 500 |  |


#### Security

- **BearerAuth**: No specific scopes


---

### DELETE /api/v1/steering-documents/{id}

Delete a steering document

Delete a steering document by UUID or reference ID. Administrators can delete any document, Users can only delete their own documents.






#### Responses

| Status Code | Description |
|-------------|-------------|
| 204 | Steering document deleted successfully |
| 401 |  |
| 403 |  |
| 404 |  |
| 500 |  |


#### Security

- **BearerAuth**: No specific scopes


---

### GET /api/v1/steering-documents

List steering documents with filtering and pagination

Retrieve a list of steering documents with optional filtering by creator and search query. Supports pagination and custom ordering. Requires authentication.


#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
| creator_id | query | map[format:uuid type:string] | No | Filter by creator UUID |
| search | query | map[type:string] | No | Search query for full-text search in title and description |
| order_by | query | map[type:string] | No | Order results by field |
| limit | query | map[default:50 maximum:100 minimum:1 type:integer] | No | Maximum number of results to return |
| offset | query | map[default:0 minimum:0 type:integer] | No | Number of results to skip for pagination |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of steering documents with count |
| 401 |  |
| 500 |  |


#### Security

- **BearerAuth**: No specific scopes


---

### POST /api/v1/steering-documents

Create a new steering document

Create a new steering document with the provided details. The steering document will be assigned a unique reference ID (STD-XXX format). Requires User or Administrator role.




#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Successfully created steering document |
| 400 |  |
| 401 |  |
| 403 |  |
| 500 |  |


#### Security

- **BearerAuth**: No specific scopes


---

## User Management

### GET /auth/users

List users (Admin only)

Get list of all users


#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of users |
| 401 |  |
| 403 |  |




---

### POST /auth/users

Create user (Admin only)

Create a new user account




#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | User created successfully |
| 400 |  |
| 401 |  |
| 403 |  |




---

### GET /auth/users/{id}

Get user by ID (Admin only)

Get user details by ID






#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | User details |
| 401 |  |
| 403 |  |
| 404 |  |




---

### PUT /auth/users/{id}

Update user (Admin only)

Update user information




#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | User updated successfully |
| 400 |  |
| 401 |  |
| 403 |  |
| 404 |  |




---

### DELETE /auth/users/{id}

Delete user (Admin only)

Delete user account






#### Responses

| Status Code | Description |
|-------------|-------------|
| 204 | User deleted successfully |
| 401 |  |
| 403 |  |
| 404 |  |




---

## User Stories

### GET /api/v1/user-stories/{id}/comments/inline/visible

Get visible user story inline comments







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of visible inline comments |




---

### PATCH /api/v1/user-stories/{id}/status

Change user story status





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Status changed |




---

### DELETE /api/v1/user-stories/{id}/delete

Comprehensive user story deletion

Delete user story with all dependencies and cascade operations






#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Deletion completed |
| 409 | Cannot delete due to dependencies |




---

### POST /api/v1/user-stories/{id}/comments/inline

Create user story inline comment





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Inline comment created |




---

### POST /api/v1/user-stories/{id}/comments/inline/validate

Validate user story inline comments





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Validation result |




---

### GET /api/v1/user-stories/{id}/acceptance-criteria

Get user story acceptance criteria







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of acceptance criteria |




---

### POST /api/v1/user-stories/{id}/acceptance-criteria

Create acceptance criteria in user story





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Acceptance criteria created |




---

### PATCH /api/v1/user-stories/{id}/assign

Assign user story to user





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | User story assigned |




---

### GET /api/v1/user-stories

List user stories



#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
| epic_id | query | map[format:uuid type:string] | No |  |
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |
| status | query | map[$ref:#/components/schemas/UserStoryStatus] | No |  |
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of user stories |




---

### POST /api/v1/user-stories

Create user story





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | User story created |




---

### GET /api/v1/user-stories/{id}/validate-deletion

Validate user story deletion

Check if user story can be deleted and get dependency information






#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Deletion validation result |
| 404 |  |




---

### GET /api/v1/user-stories/{id}/requirements

Get user story requirements







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | User story with requirements |




---

### POST /api/v1/user-stories/{id}/requirements

Create requirement in user story





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Requirement created |




---

### GET /api/v1/user-stories/{id}

Get user story by ID







#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | User story details |




---

### PUT /api/v1/user-stories/{id}

Update user story





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | User story updated |




---

### DELETE /api/v1/user-stories/{id}

Delete user story







#### Responses

| Status Code | Description |
|-------------|-------------|
| 204 | User story deleted |




---

### GET /api/v1/user-stories/{id}/comments

Get user story comments



#### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
|  |  | <no value> | No |  |
|  |  | <no value> | No |  |




#### Responses

| Status Code | Description |
|-------------|-------------|
| 200 | List of user story comments |




---

### POST /api/v1/user-stories/{id}/comments

Create user story comment





#### Request Body



**Required:** Yes


#### Responses

| Status Code | Description |
|-------------|-------------|
| 201 | Comment created |




---



## Error Handling

### Standard Error Response Format

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message"
  }
}
```

### Common Error Codes

- **VALIDATION_ERROR**: Request validation failed
- **AUTHENTICATION_REQUIRED**: JWT token required
- **INSUFFICIENT_PERMISSIONS**: User lacks required permissions
- **ENTITY_NOT_FOUND**: Requested entity doesn't exist
- **DELETION_CONFLICT**: Entity has dependencies preventing deletion
- **INTERNAL_ERROR**: Server-side error

---

*Generated from OpenAPI specification version 1.0.0*