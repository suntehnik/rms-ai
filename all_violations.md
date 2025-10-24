# All Violations Report

This document outlines architectural and API inconsistencies found in the codebase.

## 1. Inconsistent API Endpoint Naming

**Violation:** The naming convention for creating nested resources is inconsistent.

**Examples:**

*   For acceptance criteria, there is a `CreateAcceptanceCriteria` handler that is not directly exposed as an endpoint, and a `CreateAcceptanceCriteriaInUserStory` handler for the nested route. This is confusing.
*   For requirements, there is `CreateRequirement` and `CreateRequirementInUserStory`.
*   In `internal/handlers/comment_handler.go`, there are multiple functions for creating comments on different entities (`CreateEpicComment`, `CreateUserStoryComment`, etc.). A single `CreateComment` handler that takes the entity type and ID from the path would be more consistent.

**Recommendation:** Standardize on a single, consistent naming convention for all handlers. For nested resources, the handler name should clearly indicate the parent resource, for example `CreateUserStoryAcceptanceCriteria`.

## 2. Inconsistent List Response Format

**Violation:** Several API endpoints that return lists of entities do not adhere to the standard paginated response format (`{ "data": [], "total_count": 0, "limit": 0, "offset": 0 }`). Instead, they return a simple JSON array or a custom object.

**Principle Violated:** Consistency in API design. All list endpoints should follow the same structure to provide a predictable and easy-to-use API for clients.

**Examples:**

*   **`GET /api/v1/user-stories/{id}/acceptance-criteria`**: Returns `{"acceptance_criteria": [], "count": 0}`.
    *   **File:** `internal/handlers/acceptance_criteria_handler.go`
    *   **Function:** `GetAcceptanceCriteriaByUserStory`
*   **`GET /api/v1/users/{id}/acceptance-criteria`**: Returns `{"acceptance_criteria": [], "count": 0}`.
    *   **File:** `internal/handlers/acceptance_criteria_handler.go`
    *   **Function:** `GetAcceptanceCriteriaByAuthor`
*   **`GET /api/v1/comments/{id}/replies`**: Returns `{"data": [], "total_count": 0, "limit": 0, "offset": 0}` but the `limit` and `offset` are not implemented and always return the full list.
    *   **File:** `internal/handlers/comment_handler.go`
    *   **Function:** `GetCommentReplies`
*   **`GET /api/v1/epics/{id}/steering-documents`**: Returns a simple JSON array.
    *   **File:** `internal/handlers/steering_document_handler.go`
    *   **Function:** `GetEpicSteeringDocuments`
*   **`GET /api/v1/config/requirement-types`**: Returns `{"requirement_types": [], "count": 0}`.
    *   **File:** `internal/handlers/config_handler.go`
    *   **Function:** `ListRequirementTypes`
*   **`GET /api/v1/config/relationship-types`**: Returns `{"relationship_types": [], "count": 0}`.
    *   **File:** `internal/handlers/config_handler.go`
    *   **Function:** `ListRelationshipTypes`
*   **`GET /api/v1/config/status-models`**: Returns `{"status_models": [], "count": 0}`.
    *   **File:** `internal/handlers/config_handler.go`
    *   **Function:** `ListStatusModels`
*   **`GET /api/v1/config/status-models/:id/statuses`**: Returns `{"statuses": [], "count": 0}`.
    *   **File:** `internal/handlers/config_handler.go`
    *   **Function:** `ListStatusesByModel`
*   **`GET /api/v1/config/status-models/:id/transitions`**: Returns `{"transitions": [], "count": 0}`.
    *   **File:** `internal/handlers/config_handler.go`
    *   **Function:** `ListStatusTransitionsByModel`
*   **`GET /api/v1/requirements/:id/relationships`**: Returns `{"relationships": [], "count": 0}`.
    *   **File:** `internal/handlers/requirement_handler.go`
    *   **Function:** `GetRelationshipsByRequirement`
*   **`GET /api/v1/requirements/search`**: Returns `{"requirements": [], "count": 0, "query": "..."}`.
    *   **File:** `internal/handlers/requirement_handler.go`
    *   **Function:** `SearchRequirements`

**Recommendation:** Refactor these endpoints to return the standard paginated response object. This will involve updating the corresponding service and repository methods to support pagination.

## 3. Lack of Default Inclusion of Nested Entities

**Violation:** The API does not consistently return full objects for nested to-one relationships (e.g., `creator`, `assignee`, `epic`). These should be included by default to provide a better developer experience and reduce the number of API calls.

**Examples:**

*   **`GET /api/v1/epics`**: The `ListEpics` handler does not preload `Creator` and `Assignee` by default.
    *   **File:** `internal/handlers/epic_handler.go`
    *   **Function:** `ListEpics`
*   **`GET /api/v1/user-stories`**: The `ListUserStories` handler does not preload `Creator`, `Assignee`, and `Epic` by default.
    *   **File:** `internal/handlers/user_story_handler.go`
    *   **Function:** `ListUserStories`
*   **`GET /api/v1/requirements/:id`**: The `GetRequirement` handler does not preload `Creator`, `Assignee`, `UserStory`, `AcceptanceCriteria`, and `Type` by default.
    *   **File:** `internal/handlers/requirement_handler.go`
    *   **Function:** `GetRequirement`
*   **`GET /api/v1/requirements`**: The `ListRequirements` handler does not preload `Creator`, `Assignee`, `UserStory`, `AcceptanceCriteria`, and `Type` by default.
    *   **File:** `internal/handlers/requirement_handler.go`
    *   **Function:** `ListRequirements`
*   **`GET /api/v1/acceptance-criteria/:id`**: The `GetAcceptanceCriteria` handler does not preload `UserStory` and `Author` by default.
    *   **File:** `internal/handlers/acceptance_criteria_handler.go`
    *   **Function:** `GetAcceptanceCriteria`
*   **`GET /api/v1/acceptance-criteria`**: The `ListAcceptanceCriteria` handler does not preload `UserStory` and `Author` by default.
    *   **File:** `internal/handlers/acceptance_criteria_handler.go`
    *   **Function:** `ListAcceptanceCriteria`
*   **`GET /api/v1/pats`**: The `ListPATs` handler does not preload the `User` by default.
    *   **File:** `internal/handlers/pat_handler.go`
    *   **Function:** `ListPATs`

**Recommendation:** Update the corresponding service and repository methods to always preload these nested to-one relationships. The `include` parameter should be reserved for including to-many relationships (e.g., `user_stories`, `comments`) or other optional data.

## 4. Inconsistent Error Messages

**Violation:** Error messages for similar types of errors are inconsistent across different handlers.

**Examples:**

*   **Not Found Errors:**
    *   `epic_handler.go`: "Epic not found", "Creator or assignee not found"
    *   `user_story_handler.go`: "User story not found"
    *   `acceptance_criteria_handler.go`: "Acceptance criteria not found", "User story not found", "Author not found"
    *   `comment_handler.go`: "Comment not found", "Parent comment not found"
    *   `config_handler.go`: "Requirement type not found", "Relationship type not found", "Status model not found", "Status not found", "Status transition not found"
    *   `pat_handler.go`: "PAT not found"
    *   `steering_document_handler.go`: "Steering document not found"

*   **Invalid ID Errors:**
    *   `epic_handler.go`: "Invalid epic ID format"
    *   `user_story_handler.go`: "Invalid user story ID format"
    *   `acceptance_criteria_handler.go`: "Invalid acceptance criteria ID format"
    *   `comment_handler.go`: "Invalid entity ID format", "Invalid comment ID format"

*   **Deletion Conflict Errors:**
    *   `epic_handler.go`: "Epic has associated user stories and cannot be deleted. Use force=true to delete with dependencies"
    *   `user_story_handler.go`: "User story has associated requirements and cannot be deleted"
    *   `acceptance_criteria_handler.go`: "Acceptance criteria has associated requirements and cannot be deleted"
    *   `requirement_handler.go`: "Requirement has associated relationships and cannot be deleted"

**Recommendation:** Use a more generic and consistent error messaging strategy. For example, a single error message like "Cannot delete {entity_type} due to dependencies. Use force=true to override." could be used, with `{entity_type}` being replaced with the actual entity type. A set of standard error messages for common errors like "Not Found" and "Invalid ID" should be defined and used across all handlers.

---

## Execution Plan

This section outlines the steps to fix the violations identified in this report.

### 1. Fix Inconsistent List Response Format

The goal of this section is to refactor all list-based endpoints to return a consistent, paginated response in the format: `{ "data": [], "total_count": 0, "limit": 0, "offset": 0 }`.

#### Step 1.1: Refactor `acceptance_criteria_handler.go`

1.  **File to modify:** `internal/handlers/acceptance_criteria_handler.go`
2.  **Function to modify:** `GetAcceptanceCriteriaByUserStory`
    *   **Current behavior:** Returns `{"acceptance_criteria": [], "count": 0}`.
    *   **Target behavior:** Should return the standard paginated response.
    *   **Action:**
        *   Update the `acceptanceCriteriaService.GetAcceptanceCriteriaByUserStory` to support pagination and return a total count.
        *   Update the handler to accept `limit` and `offset` query parameters.
        *   Update the handler to return the standard paginated response object.
3.  **Function to modify:** `GetAcceptanceCriteriaByAuthor`
    *   **Current behavior:** Returns `{"acceptance_criteria": [], "count": 0}`.
    *   **Target behavior:** Should return the standard paginated response.
    *   **Action:**
        *   Update the `acceptanceCriteriaService.GetAcceptanceCriteriaByAuthor` to support pagination and return a total count.
        *   Update the handler to accept `limit` and `offset` query parameters.
        *   Update the handler to return the standard paginated response object.

#### Step 1.2: Refactor `comment_handler.go`

1.  **File to modify:** `internal/handlers/comment_handler.go`
2.  **Function to modify:** `GetCommentReplies`
    *   **Current behavior:** Returns a paginated response, but pagination is not implemented.
    *   **Target behavior:** Should correctly implement pagination.
    *   **Action:**
        *   Update the `commentService.GetCommentReplies` to support pagination (`limit` and `offset`) and return a total count.
        *   Update the handler to pass the `limit` and `offset` query parameters to the service.

#### Step 1.3: Refactor `config_handler.go`

1.  **File to modify:** `internal/handlers/config_handler.go`
2.  **Functions to modify:**
    *   `ListRequirementTypes`
    *   `ListRelationshipTypes`
    *   `ListStatusModels`
    *   `ListStatusesByModel`
    *   `ListStatusTransitionsByModel`
    *   **Current behavior:** All these functions return a response with a `count` field and the data in a field named after the entity type (e.g., `requirement_types`).
    *   **Target behavior:** Should return the standard paginated response.
    *   **Action:**
        *   For each function, update the corresponding service method to return a total count.
        *   Update each handler to return the standard paginated response object.

#### Step 1.4: Refactor `requirement_handler.go`

1.  **File to modify:** `internal/handlers/requirement_handler.go`
2.  **Function to modify:** `GetRelationshipsByRequirement`
    *   **Current behavior:** Returns `{"relationships": [], "count": 0}`.
    *   **Target behavior:** Should return the standard paginated response.
    *   **Action:**
        *   Update the `requirementService.GetRelationshipsByRequirement` to support pagination and return a total count.
        *   Update the handler to accept `limit` and `offset` query parameters.
        *   Update the handler to return the standard paginated response object.
3.  **Function to modify:** `SearchRequirements`
    *   **Current behavior:** Returns `{"requirements": [], "count": 0, "query": "..."}`.
    *   **Target behavior:** Should return the standard paginated response.
    *   **Action:**
        *   Update the `requirementService.SearchRequirements` to support pagination and return a total count.
        *   Update the handler to accept `limit` and `offset` query parameters.
        *   Update the handler to return the standard paginated response object.

#### Step 1.5: Refactor `steering_document_handler.go`

1.  **File to modify:** `internal/handlers/steering_document_handler.go`
2.  **Function to modify:** `GetEpicSteeringDocuments`
    *   **Current behavior:** Returns a simple JSON array.
    *   **Target behavior:** Should return the standard paginated response.
    *   **Action:**
        *   Update the `steeringDocumentService.GetEpicSteeringDocuments` to support pagination and return a total count.
        *   Update the handler to accept `limit` and `offset` query parameters.
        *   Update the handler to return the standard paginated response object.

### 2. Fix Lack of Default Inclusion of Nested Entities

The goal of this section is to refactor the API to include nested to-one relationships by default.

#### Step 2.1: Update `epic_handler.go`

1.  **File to modify:** `internal/handlers/epic_handler.go`
2.  **Function to modify:** `ListEpics`
    *   **Action:** Update the `epicService.ListEpics` to always preload `Creator` and `Assignee`.

#### Step 2.2: Update `user_story_handler.go`

1.  **File to modify:** `internal/handlers/user_story_handler.go`
2.  **Function to modify:** `ListUserStories`
    *   **Action:** Update the `userStoryService.ListUserStories` to always preload `Creator`, `Assignee`, and `Epic`.

#### Step 2.3: Update `requirement_handler.go`

1.  **File to modify:** `internal/handlers/requirement_handler.go`
2.  **Functions to modify:** `GetRequirement`, `ListRequirements`
    *   **Action:** Update the `requirementService.GetRequirement` and `requirementService.ListRequirements` to always preload `Creator`, `Assignee`, `UserStory`, `AcceptanceCriteria`, and `Type`.

#### Step 2.4: Update `acceptance_criteria_handler.go`

1.  **File to modify:** `internal/handlers/acceptance_criteria_handler.go`
2.  **Functions to modify:** `GetAcceptanceCriteria`, `ListAcceptanceCriteria`
    *   **Action:** Update the `acceptanceCriteriaService.GetAcceptanceCriteria` and `acceptanceCriteriaService.ListAcceptanceCriteria` to always preload `UserStory` and `Author`.

#### Step 2.5: Update `pat_handler.go`

1.  **File to modify:** `internal/handlers/pat_handler.go`
2.  **Function to modify:** `ListPATs`
    *   **Action:** Update the `patService.ListUserPATs` to always preload the `User`.

### 3. Fix Inconsistent API Endpoint Naming

The goal of this section is to standardize the naming of API endpoints and handlers.

#### Step 3.1: Refactor `acceptance_criteria_handler.go`

1.  **File to modify:** `internal/handlers/acceptance_criteria_handler.go`
2.  **Function to rename:** `CreateAcceptanceCriteriaInUserStory` to `CreateAcceptanceCriteria`.
3.  **Action:** The existing `CreateAcceptanceCriteria` function should be renamed to `CreateAcceptanceCriteriaInUserStory` and the logic merged to have a single `CreateAcceptanceCriteria` function that handles the nested creation.

#### Step 3.2: Refactor `requirement_handler.go`

1.  **File to modify:** `internal/handlers/requirement_handler.go`
2.  **Function to rename:** `CreateRequirementInUserStory` to `CreateRequirement`.
3.  **Action:** The existing `CreateRequirement` function should be renamed to `CreateRequirementInUserStory` and the logic merged to have a single `CreateRequirement` function that handles the nested creation.

#### Step 3.3: Refactor `comment_handler.go`

1.  **File to modify:** `internal/handlers/comment_handler.go`
2.  **Functions to merge:** `CreateEpicComment`, `CreateUserStoryComment`, `CreateAcceptanceCriteriaComment`, `CreateRequirementComment` into a single `CreateComment` handler.
3.  **Action:** The `CreateComment` handler should determine the entity type from the URL path and then call the appropriate service method.

### 4. Fix Inconsistent Error Messages (Low Priority)

The goal of this section is to standardize error messages across the API.

#### Step 4.1: Standardize "Not Found" Error Messages

1.  **Action:** Create a set of standard "not found" error messages in the `service` layer.
2.  **Files to modify:** All handler files in `internal/handlers/`.
3.  **Details:**
    *   Replace specific "not found" messages like "Epic not found", "User story not found", etc., with a more generic message, for example, `fmt.Sprintf("%s not found", entityType)`.
    *   Ensure that the error messages are consistent for both direct and related entity lookups.

#### Step 4.2: Standardize "Invalid ID" Error Messages

1.  **Action:** Create a standard "invalid ID" error message.
2.  **Files to modify:** All handler files in `internal/handlers/`.
3.  **Details:**
    *   Replace specific "invalid ID" messages like "Invalid epic ID format", "Invalid user story ID format", etc., with a generic message like `fmt.Sprintf("Invalid %s ID format", entityType)`.

#### Step 4.3: Standardize "Deletion Conflict" Error Messages

1.  **Action:** Create a standard "deletion conflict" error message.
2.  **Files to modify:** `internal/handlers/deletion_handler.go` and other handlers with delete functionality.
3.  **Details:**
    *   Replace specific deletion conflict messages with a generic message like `fmt.Sprintf("Cannot delete %s due to dependencies. Use force=true to override.", entityType)`.