# Requirements Document

## Introduction

The e2e tests are failing due to a database constraint violation on the `reference_id` field when creating epics. The issue occurs because GORM is inserting empty strings for the `reference_id` field instead of allowing the database to generate unique reference IDs. This causes duplicate key violations when multiple epics are created in tests. This is a critical bug identified in task 27 of the product-requirements-management spec.

## Requirements

### Requirement 1

**User Story:** As a developer running e2e tests, I want epic creation to work correctly without reference ID conflicts, so that the test suite passes reliably and CI/CD pipelines don't fail.

#### Acceptance Criteria

1. WHEN an epic is created through the API THEN the system SHALL generate a unique reference ID in format "EP-001", "EP-002", etc.
2. WHEN multiple epics are created in sequence THEN each SHALL receive a unique reference ID without conflicts
3. WHEN GORM creates an epic record THEN it SHALL generate reference IDs using application-level logic consistent with other models
4. WHEN the e2e cache invalidation test creates epics THEN it SHALL NOT encounter duplicate key constraint violations

### Requirement 2

**User Story:** As a developer, I want the Epic model to use the same reference ID generation pattern as AcceptanceCriteria and Requirement models, so that the system is consistent and maintainable.

#### Acceptance Criteria

1. WHEN an Epic is created THEN the BeforeCreate hook SHALL generate reference IDs using database count method
2. WHEN the Epic model generates reference IDs THEN it SHALL use the same pattern as AcceptanceCriteria and Requirement models
3. WHEN the Epic model is imported THEN it SHALL include the required fmt import for string formatting
4. WHEN concurrent Epic creation occurs THEN there SHALL be no race conditions in reference ID generation

### Requirement 3

**User Story:** As a developer, I want comprehensive test coverage for Epic reference ID generation, so that this issue doesn't regress in the future.

#### Acceptance Criteria

1. WHEN unit tests run THEN they SHALL verify Epic reference ID generation works correctly
2. WHEN e2e tests run THEN the cache invalidation test SHALL pass without reference ID conflicts
3. WHEN concurrent Epic creation scenarios are tested THEN they SHALL ensure no race conditions exist
4. WHEN the test suite runs THEN all tests SHALL pass without database constraint violations