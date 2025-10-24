# Requirements Document

## Introduction

This specification addresses architectural and API inconsistencies identified in the Product Requirements Management System codebase. The goal is to standardize API response formats, improve consistency in endpoint naming, ensure proper inclusion of nested entities, and standardize error messaging across all handlers.

## Glossary

- **API_Handler**: Go functions in the handlers package that process HTTP requests and return responses
- **List_Response**: Standardized paginated response format with data, total_count, limit, and offset fields
- **Nested_Entity**: Related database entities that should be preloaded in API responses (e.g., creator, assignee)
- **Entity_Type**: The type of business object (epic, user_story, acceptance_criteria, requirement)
- **Pagination_Parameters**: Query parameters limit and offset used for paginating list responses
- **Service_Layer**: Business logic layer that handlers call to perform operations
- **Repository_Layer**: Data access layer that services call to interact with the database

## Requirements

### Requirement 1

**User Story:** As an API client developer, I want all list endpoints to return consistent paginated responses, so that I can implement uniform pagination logic across my application.

#### Current Inconsistencies

- `GET /api/v1/user-stories/{id}/acceptance-criteria` returns `{"acceptance_criteria": [], "count": 0}`
- `GET /api/v1/config/requirement-types` returns `{"requirement_types": [], "count": 0}`
- `GET /api/v1/epics/{id}/steering-documents` returns a simple JSON array
- `GET /api/v1/comments/{id}/replies` claims pagination but doesn't implement limit/offset
- `GET /api/v1/requirements/search` returns `{"requirements": [], "count": 0, "query": "..."}`

#### Acceptance Criteria

1. WHEN `GET /api/v1/user-stories/{id}/acceptance-criteria` is called, THE `GetAcceptanceCriteriaByUserStory` method in file `internal/handlers/acceptance_criteria_handler.go` SHALL return response in format `{"data": [], "total_count": 0, "limit": 0, "offset": 0}`
2. WHEN `GET /api/v1/users/{id}/acceptance-criteria` is called, THE `GetAcceptanceCriteriaByAuthor` method in file `internal/handlers/acceptance_criteria_handler.go` SHALL return response in format `{"data": [], "total_count": 0, "limit": 0, "offset": 0}`
3. WHEN `GET /api/v1/comments/{id}/replies` is called, THE `GetCommentReplies` method in file `internal/handlers/comment_handler.go` SHALL properly implement limit and offset parameters
4. WHEN `GET /api/v1/config/requirement-types` is called, THE `ListRequirementTypes` method in file `internal/handlers/config_handler.go` SHALL return response in format `{"data": [], "total_count": 0, "limit": 0, "offset": 0}`
5. WHEN `GET /api/v1/config/relationship-types` is called, THE `ListRelationshipTypes` method in file `internal/handlers/config_handler.go` SHALL return response in format `{"data": [], "total_count": 0, "limit": 0, "offset": 0}`
6. WHEN `GET /api/v1/config/status-models` is called, THE `ListStatusModels` method in file `internal/handlers/config_handler.go` SHALL return response in format `{"data": [], "total_count": 0, "limit": 0, "offset": 0}`
7. WHEN `GET /api/v1/config/status-models/:id/statuses` is called, THE `ListStatusesByModel` method in file `internal/handlers/config_handler.go` SHALL return response in format `{"data": [], "total_count": 0, "limit": 0, "offset": 0}`
8. WHEN `GET /api/v1/config/status-models/:id/transitions` is called, THE `ListStatusTransitionsByModel` method in file `internal/handlers/config_handler.go` SHALL return response in format `{"data": [], "total_count": 0, "limit": 0, "offset": 0}`
9. WHEN `GET /api/v1/requirements/:id/relationships` is called, THE `GetRelationshipsByRequirement` method in file `internal/handlers/requirement_handler.go` SHALL return response in format `{"data": [], "total_count": 0, "limit": 0, "offset": 0}`
10. WHEN `GET /api/v1/requirements/search` is called, THE `SearchRequirements` method in file `internal/handlers/requirement_handler.go` SHALL return response in format `{"data": [], "total_count": 0, "limit": 0, "offset": 0}`
11. WHEN `GET /api/v1/epics/{id}/steering-documents` is called, THE `GetEpicSteeringDocuments` method in file `internal/handlers/steering_document_handler.go` SHALL return response in format `{"data": [], "total_count": 0, "limit": 0, "offset": 0}`

### Requirement 2

**User Story:** As an API client developer, I want nested to-one relationships to be included by default in API responses, so that I can reduce the number of API calls needed to display complete information.

#### Current Missing Preloads

- `GET /api/v1/epics` - ListEpics handler missing Creator and Assignee preloads
- `GET /api/v1/user-stories` - ListUserStories handler missing Creator, Assignee, and Epic preloads
- `GET /api/v1/requirements/:id` - GetRequirement handler missing Creator, Assignee, UserStory, AcceptanceCriteria, Type preloads
- `GET /api/v1/requirements` - ListRequirements handler missing all relationship preloads
- `GET /api/v1/acceptance-criteria/:id` - GetAcceptanceCriteria handler missing UserStory and Author preloads
- `GET /api/v1/pats` - ListPATs handler missing User preload

#### Acceptance Criteria

1. WHEN `GET /api/v1/epics` is called, THE `ListEpics` method in file `internal/handlers/epic_handler.go` SHALL preload Creator and Assignee entities in the response
2. WHEN `GET /api/v1/user-stories` is called, THE `ListUserStories` method in file `internal/handlers/user_story_handler.go` SHALL preload Creator, Assignee, and Epic entities in the response
3. WHEN `GET /api/v1/requirements/:id` is called, THE `GetRequirement` method in file `internal/handlers/requirement_handler.go` SHALL preload Creator, Assignee, UserStory, AcceptanceCriteria, and Type entities in the response
4. WHEN `GET /api/v1/requirements` is called, THE `ListRequirements` method in file `internal/handlers/requirement_handler.go` SHALL preload Creator, Assignee, UserStory, AcceptanceCriteria, and Type entities in the response
5. WHEN `GET /api/v1/acceptance-criteria/:id` is called, THE `GetAcceptanceCriteria` method in file `internal/handlers/acceptance_criteria_handler.go` SHALL preload UserStory and Author entities in the response
6. WHEN `GET /api/v1/acceptance-criteria` is called, THE `ListAcceptanceCriteria` method in file `internal/handlers/acceptance_criteria_handler.go` SHALL preload UserStory and Author entities in the response
7. WHEN `GET /api/v1/pats` is called, THE `ListPATs` method in file `internal/handlers/pat_handler.go` SHALL preload User entity in the response

### Requirement 3

**User Story:** As a developer maintaining the API, I want consistent naming conventions for handlers and endpoints, so that the codebase is easier to understand and maintain.

#### Current Naming Inconsistencies

- Acceptance criteria: `CreateAcceptanceCriteria` (unused) and `CreateAcceptanceCriteriaInUserStory` handlers
- Requirements: `CreateRequirement` and `CreateRequirementInUserStory` handlers
- Comments: Multiple handlers `CreateEpicComment`, `CreateUserStoryComment`, `CreateAcceptanceCriteriaComment`, `CreateRequirementComment`
- Handler names don't clearly indicate parent resource relationships

#### Acceptance Criteria

1. WHEN `POST /api/v1/user-stories/{id}/acceptance-criteria` is called, THE `CreateAcceptanceCriteriaInUserStory` method in file `internal/handlers/acceptance_criteria_handler.go` SHALL be renamed to `CreateAcceptanceCriteria` and handle nested creation
2. WHEN `POST /api/v1/user-stories/{id}/requirements` is called, THE `CreateRequirementInUserStory` method in file `internal/handlers/requirement_handler.go` SHALL be renamed to `CreateRequirement` and handle nested creation
3. WHEN `POST /api/v1/epics/{id}/comments` is called, THE `CreateEpicComment` method in file `internal/handlers/comment_handler.go` SHALL be replaced by a single `CreateComment` method
4. WHEN `POST /api/v1/user-stories/{id}/comments` is called, THE `CreateUserStoryComment` method in file `internal/handlers/comment_handler.go` SHALL be replaced by a single `CreateComment` method
5. WHEN `POST /api/v1/acceptance-criteria/{id}/comments` is called, THE `CreateAcceptanceCriteriaComment` method in file `internal/handlers/comment_handler.go` SHALL be replaced by a single `CreateComment` method
6. WHEN `POST /api/v1/requirements/{id}/comments` is called, THE `CreateRequirementComment` method in file `internal/handlers/comment_handler.go` SHALL be replaced by a single `CreateComment` method

### Requirement 4

**User Story:** As an API client developer, I want consistent error messages across all endpoints, so that I can implement uniform error handling in my application.

#### Current Error Message Inconsistencies

**Not Found Errors:**
- epic_handler.go: "Epic not found", "Creator or assignee not found"
- user_story_handler.go: "User story not found"
- acceptance_criteria_handler.go: "Acceptance criteria not found", "User story not found", "Author not found"
- comment_handler.go: "Comment not found", "Parent comment not found"

**Invalid ID Errors:**
- epic_handler.go: "Invalid epic ID format"
- user_story_handler.go: "Invalid user story ID format"
- comment_handler.go: "Invalid entity ID format", "Invalid comment ID format"

**Deletion Conflict Errors:**
- epic_handler.go: "Epic has associated user stories and cannot be deleted. Use force=true to delete with dependencies"
- user_story_handler.go: "User story has associated requirements and cannot be deleted"
- requirement_handler.go: "Requirement has associated relationships and cannot be deleted"

#### Acceptance Criteria

1. WHEN entity not found errors occur in file `internal/handlers/epic_handler.go`, THE error messages SHALL be standardized to format "{Entity_Type} not found"
2. WHEN entity not found errors occur in file `internal/handlers/user_story_handler.go`, THE error messages SHALL be standardized to format "{Entity_Type} not found"
3. WHEN entity not found errors occur in file `internal/handlers/acceptance_criteria_handler.go`, THE error messages SHALL be standardized to format "{Entity_Type} not found"
4. WHEN entity not found errors occur in file `internal/handlers/comment_handler.go`, THE error messages SHALL be standardized to format "{Entity_Type} not found"
5. WHEN invalid ID errors occur in any handler file, THE error messages SHALL be standardized to format "Invalid {Entity_Type} ID format"
6. WHEN deletion conflict errors occur in any handler file, THE error messages SHALL be standardized to format "Cannot delete {Entity_Type} due to dependencies. Use force=true to override."
7. WHEN standardized error messages are needed, THE Service_Layer SHALL provide error message templates in file `internal/service/errors.go`