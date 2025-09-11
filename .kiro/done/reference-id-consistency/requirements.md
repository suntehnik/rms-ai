# Requirements Document

## Introduction

This feature addresses the inconsistency in ReferenceID generation strategies across different entities in the product requirements management system. After analyzing the latest codebase, I found that:

- **Epic**: Uses PostgreSQL advisory locks with UUID fallback (sophisticated approach)
- **UserStory**: Uses PostgreSQL advisory locks with UUID fallback (sophisticated approach) 
- **Requirement**: Uses simple count-based approach in BeforeCreate hook (basic approach)
- **AcceptanceCriteria**: Uses simple count-based approach in BeforeCreate hook (basic approach)

The current implementation mixes production-grade concurrency handling with simple approaches that are vulnerable to race conditions. Additionally, the system needs different strategies for production (PostgreSQL with advisory locks) versus unit testing (SQLite with simple counting). This inconsistency can lead to race conditions, duplicate reference IDs, and maintenance issues, especially under concurrent load.

## Requirements

### Requirement 1

**User Story:** As a developer, I want a production-grade reference ID generator for all non-unit-test scenarios, so that the system uses consistent, thread-safe generation in production, integration, and e2e tests.

#### Acceptance Criteria

1. WHEN the system runs in production THEN it SHALL use a PostgreSQL-optimized generator with advisory locks
2. WHEN the system runs integration tests THEN it SHALL use the same PostgreSQL-optimized generator
3. WHEN the system runs e2e tests THEN it SHALL use the same PostgreSQL-optimized generator  
4. WHEN advisory lock acquisition fails THEN the system SHALL fall back to UUID-based reference IDs
5. IF ReferenceID is already set THEN the system SHALL NOT override the existing value

### Requirement 2

**User Story:** As a system administrator, I want reference ID generation to be thread-safe and handle concurrent operations, so that duplicate reference IDs are never created.

#### Acceptance Criteria

1. WHEN multiple entities are created concurrently THEN each entity SHALL receive a unique ReferenceID
2. WHEN advisory lock is not available THEN the system SHALL use UUID-based fallback to ensure uniqueness
3. WHEN database errors occur during reference ID generation THEN the system SHALL handle them gracefully
4. IF reference ID conflicts occur THEN the system SHALL retry with a different ID

### Requirement 3

**User Story:** As a developer, I want the reference ID generation logic to be centralized with separate implementations for different environments, so that it's easier to maintain and test.

#### Acceptance Criteria

1. WHEN reference ID generation is needed THEN the system SHALL use a centralized generator interface
2. WHEN different entity types need reference IDs THEN each type SHALL have its own prefix (EP-, US-, REQ-, AC-)
3. WHEN the generation logic changes THEN it SHALL only need to be updated in the respective generator implementation
4. WHEN new entities with reference IDs are added THEN they SHALL easily integrate with both generator types
5. IF the system switches database types THEN the appropriate generator SHALL be selected automatically

### Requirement 4

**User Story:** As a system architect, I want static selection of reference ID generators rather than dynamic selection, so that the system behavior is predictable and there's no runtime overhead for generator selection.

#### Acceptance Criteria

1. WHEN production code is compiled THEN it SHALL statically use the PostgreSQL-optimized generator
2. WHEN integration tests are compiled THEN they SHALL statically use the PostgreSQL-optimized generator
3. WHEN e2e tests are compiled THEN they SHALL statically use the PostgreSQL-optimized generator
4. WHEN unit tests are compiled THEN they SHALL statically use a separate simple test generator
5. IF generator selection is needed THEN it SHALL be determined at compile time, not runtime

### Requirement 5

**User Story:** As a system architect, I want unit tests to use a separate, simple reference ID generator that is isolated from production code, so that unit tests are fast and don't depend on PostgreSQL features.

#### Acceptance Criteria

1. WHEN unit tests need reference ID generation THEN they SHALL use a simple test generator located in test directories
2. WHEN production code is built THEN the test generator SHALL NOT be included or accessible
3. WHEN unit tests run THEN they SHALL use simple counting without advisory locks for speed
4. WHEN the test generator is implemented THEN it SHALL be completely separate from production generator code
5. IF unit tests need to mock reference ID behavior THEN the test generator SHALL provide predictable, sequential IDs

### Requirement 6

**User Story:** As a developer, I want proper documentation of both reference ID generation strategies, so that future maintainers understand when and how each is used.

#### Acceptance Criteria

1. WHEN developers need to understand reference ID generation THEN comprehensive documentation SHALL be available for both generators
2. WHEN new team members join THEN they SHALL have clear guidance on which generator is used in which scenario
3. WHEN troubleshooting reference ID issues THEN diagnostic information SHALL be available for both production and test scenarios
4. IF either generation strategy changes THEN documentation SHALL be updated accordingly
5. WHEN writing new tests THEN clear guidance SHALL be available on which generator to use