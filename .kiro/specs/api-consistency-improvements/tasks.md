# Implementation Plan

- [x] 1. Create standardized ListResponse struct
  - Create `ListResponse` struct with `data`, `total_count`, `limit`, `offset` fields in `internal/handlers/response.go`
  - Define generic `ListResponse[T any]` type for type safety
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 1.8, 1.9, 1.10, 1.11_

- [x] 1.2 Update acceptance criteria handlers to use ListResponse
  - Modify `GetAcceptanceCriteriaByUserStory` method in `internal/handlers/acceptance_criteria_handler.go` to return `ListResponse[models.AcceptanceCriteria]`
  - Modify `GetAcceptanceCriteriaByAuthor` method in `internal/handlers/acceptance_criteria_handler.go` to return `ListResponse[models.AcceptanceCriteria]`
  - _Requirements: 1.1, 1.2_

- [x] 1.3 Update comment handlers to implement proper pagination with ListResponse
  - Modify `GetCommentReplies` method in `internal/handlers/comment_handler.go` to implement limit/offset and return `ListResponse[service.CommentResponse]`
  - Update service layer to support pagination parameters
  - _Requirements: 1.3_

- [x] 1.4 Update config handlers to use ListResponse
  - Modify `ListRequirementTypes` method in `internal/handlers/config_handler.go` to return `ListResponse[models.RequirementType]`
  - Modify `ListRelationshipTypes` method in `internal/handlers/config_handler.go` to return `ListResponse[models.RelationshipType]`
  - Modify `ListStatusModels` method in `internal/handlers/config_handler.go` to return `ListResponse[models.StatusModel]`
  - Modify `ListStatusesByModel` method in `internal/handlers/config_handler.go` to return `ListResponse[models.Status]`
  - Modify `ListStatusTransitionsByModel` method in `internal/handlers/config_handler.go` to return `ListResponse[models.StatusTransition]`
  - _Requirements: 1.4, 1.5, 1.6, 1.7, 1.8_

- [x] 1.5 Update requirement handlers to use ListResponse
  - Modify `GetRelationshipsByRequirement` method in `internal/handlers/requirement_handler.go` to return `ListResponse[models.RequirementRelationship]`
  - Modify `SearchRequirements` method in `internal/handlers/requirement_handler.go` to return `ListResponse[models.Requirement]`
  - _Requirements: 1.9, 1.10_

- [x] 1.6 Update steering document handlers to use ListResponse
  - Modify `GetEpicSteeringDocuments` method in `internal/handlers/steering_document_handler.go` to return `ListResponse[models.SteeringDocument]`
  - _Requirements: 1.11_

- [x] 2. Default preloads

- [x] 2.1 Update epic handlers to include default preloads
  - Modify `ListEpics` method in `internal/handlers/epic_handler.go` to preload Creator and Assignee
  - Update service layer to include preloads by default
  - _Requirements: 2.1_

- [x] 2.2 Update user story handlers to include default preloads
  - Modify `ListUserStories` method in `internal/handlers/user_story_handler.go` to preload Creator, Assignee, and Epic
  - Update service layer to include preloads by default
  - _Requirements: 2.2_

- [x] 2.3 Update requirement handlers to include default preloads
  - Modify `GetRequirement` method in `internal/handlers/requirement_handler.go` to preload Creator, Assignee, UserStory, AcceptanceCriteria, Type
  - Modify `ListRequirements` method in `internal/handlers/requirement_handler.go` to preload all relationships
  - Update service layer to include preloads by default
  - _Requirements: 2.3, 2.4_

- [x] 2.4 Update acceptance criteria handlers to include default preloads
  - Modify `GetAcceptanceCriteria` method in `internal/handlers/acceptance_criteria_handler.go` to preload UserStory and Author
  - Modify `ListAcceptanceCriteria` method in `internal/handlers/acceptance_criteria_handler.go` to preload relationships
  - Update service layer to include preloads by default
  - _Requirements: 2.5, 2.6_

- [x] 2.5 Update PAT handlers to include default preloads
  - Modify `ListPATs` method in `internal/handlers/pat_handler.go` to preload User entity
  - Update service layer to include preloads by default
  - _Requirements: 2.7_

- [ ] 3.1 Consolidate acceptance criteria creation handlers
  - Remove unused `CreateAcceptanceCriteria` handler if it exists
  - Rename `CreateAcceptanceCriteriaInUserStory` to `CreateAcceptanceCriteria` in `internal/handlers/acceptance_criteria_handler.go`
  - Update routing to use consolidated handler
  - _Requirements: 3.1_

- [ ] 3.2 Consolidate requirement creation handlers
  - Rename `CreateRequirementInUserStory` to `CreateRequirement` in `internal/handlers/requirement_handler.go`
  - Update routing to use consolidated handler for nested creation
  - _Requirements: 3.2_

- [ ] 3.3 Consolidate comment creation handlers
  - Replace `CreateEpicComment`, `CreateUserStoryComment`, `CreateAcceptanceCriteriaComment`, `CreateRequirementComment` with single `CreateComment` method
  - Update routing to use parameterized entity type routing
  - Modify existing `CreateComment` method to handle all entity types
  - _Requirements: 3.3, 3.4, 3.5, 3.6_

- [ ] 4.1 Create centralized error message templates
  - Create error message templates in `internal/service/errors.go`
  - Add `NotFound`, `InvalidID`, and `DeletionConflict` message functions
  - Define standard error message formats
  - _Requirements: 4.7_

- [ ] 4.2 Update epic handler error messages
  - Standardize "not found" errors to use "{Entity_Type} not found" format
  - Standardize invalid ID errors to use "Invalid {Entity_Type} ID format" format
  - Standardize deletion conflict errors to use standard format
  - _Requirements: 4.1_

- [ ] 4.3 Update user story handler error messages
  - Standardize "not found" errors to use "{Entity_Type} not found" format
  - Standardize invalid ID errors to use "Invalid {Entity_Type} ID format" format
  - Standardize deletion conflict errors to use standard format
  - _Requirements: 4.2_

- [ ] 4.4 Update acceptance criteria handler error messages
  - Standardize "not found" errors to use "{Entity_Type} not found" format
  - Standardize invalid ID errors to use "Invalid {Entity_Type} ID format" format
  - Standardize deletion conflict errors to use standard format
  - _Requirements: 4.3_

- [ ] 4.5 Update comment handler error messages
  - Standardize "not found" errors to use "{Entity_Type} not found" format
  - Standardize invalid ID errors to use "Invalid {Entity_Type} ID format" format
  - Standardize deletion conflict errors to use standard format
  - _Requirements: 4.4_

- [ ] 4.6 Update requirement handler error messages
  - Standardize invalid ID errors to use "Invalid {Entity_Type} ID format" format
  - Standardize deletion conflict errors to use standard format
  - _Requirements: 4.5_

- [ ] 4.7 Update all handlers to use centralized error messages
  - Replace hardcoded error messages with calls to centralized error templates
  - Ensure consistent error message format across all handlers
  - _Requirements: 4.6_

- [ ] 5.1 Update service methods to support standardized pagination
  - Modify service layer methods to return total count along with data
  - Ensure all list methods support limit/offset parameters
  - Update method signatures to support new response format
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 1.8, 1.9, 1.10, 1.11_

- [ ] 5.2 Update service methods to include default preloads
  - Modify service layer to automatically include relationship preloads
  - Ensure consistent preloading across all entity types
  - Update repository calls to include preload specifications
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7_

- [ ]* 6.1 Update handler tests for new response formats
  - Update existing handler tests to expect new standardized response format
  - Add tests for preloading functionality
  - Add tests for consolidated handlers
  - _Requirements: All requirements_

- [ ]* 6.2 Update integration tests
  - Update integration tests to work with new response formats
  - Test error message consistency
  - Test preloading functionality
  - _Requirements: All requirements_

- [ ]* 6.3 Update API documentation
  - Update OpenAPI specifications to reflect new response formats
  - Update error message documentation
  - Update handler consolidation documentation
  - _Requirements: All requirements_
