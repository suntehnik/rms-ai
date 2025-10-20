# Implementation Plan: Field Renaming from `last_modified` to `updated_at`

- [x] 1. Create database migration files
  - Create new migration file `000004_rename_last_modified_to_updated_at.up.sql` with column rename statements for all four tables (epics, user_stories, acceptance_criteria, requirements)
  - Create corresponding down migration file `000004_rename_last_modified_to_updated_at.down.sql` for rollback capability
  - Include trigger rename statements for improved naming consistency
  - _Requirements: Database schema consistency, rollback capability_

- [x] 2. Update Epic model structure and methods
  - Rename `LastModified` field to `UpdatedAt` in Epic struct with updated JSON and database tags
  - Update `BeforeUpdate` method to set `UpdatedAt` field instead of `LastModified`
  - Update custom `MarshalJSON` method to use `updated_at` key in result map
  - _Requirements: Go model consistency, JSON serialization compatibility_

- [x] 3. Update UserStory model structure and methods
  - Rename `LastModified` field to `UpdatedAt` in UserStory struct with updated JSON and database tags
  - Update `BeforeUpdate` method to set `UpdatedAt` field instead of `LastModified`
  - Update any custom JSON marshaling logic if present
  - _Requirements: Go model consistency, JSON serialization compatibility_

- [x] 4. Update AcceptanceCriteria model structure and methods
  - Rename `LastModified` field to `UpdatedAt` in AcceptanceCriteria struct with updated JSON and database tags
  - Update `BeforeUpdate` method to set `UpdatedAt` field instead of `LastModified`
  - Update any custom JSON marshaling logic if present
  - _Requirements: Go model consistency, JSON serialization compatibility_

- [x] 5. Update Requirement model structure and methods
  - Rename `LastModified` field to `UpdatedAt` in Requirement struct with updated JSON and database tags
  - Update `BeforeUpdate` method to set `UpdatedAt` field instead of `LastModified`
  - Update any custom JSON marshaling logic if present
  - _Requirements: Go model consistency, JSON serialization compatibility_

- [x] 6. Update repository layer queries and methods
  - Search and update all repository files for references to `LastModified` field in queries, ordering, and filtering
  - Update any raw SQL queries that reference `last_modified` column name
  - Update ORDER BY clauses and WHERE conditions that use the timestamp field
  - _Requirements: Database query compatibility, repository layer functionality_

- [x] 7. Update service layer business logic
  - Search and update all service files for references to `LastModified` field
  - Update any timestamp comparison logic to use `UpdatedAt` field
  - Update any caching keys or logic that incorporates the field name
  - _Requirements: Service layer functionality, business logic integrity_

- [x] 8. Update API handlers and response serialization
  - Search and update all handler files for references to `LastModified` field
  - Verify that API responses will serialize with `updated_at` key after model changes
  - Update any manual JSON construction that includes the timestamp field
  - _Requirements: API response consistency, handler functionality_

- [x] 9. Update OpenAPI specification schemas
  - Update Epic schema in `docs/openapi-v3.yaml` to use `updated_at` property instead of `last_modified`
  - Update UserStory schema to use `updated_at` property instead of `last_modified`
  - Update AcceptanceCriteria schema to use `updated_at` property instead of `last_modified`
  - Update Requirement schema to use `updated_at` property instead of `last_modified`
  - Update required field lists to reference `updated_at` instead of `last_modified`
  - _Requirements: API documentation accuracy, schema consistency_

- [x] 10. Update all test files for field name changes
  - Search and update all test files (`*_test.go`) for references to `LastModified` field
  - Update test assertions and expectations to use `UpdatedAt` field
  - Update JSON test data and expectations to use `updated_at` key
  - Update any test helper functions that reference the timestamp field
  - _Requirements: Test suite functionality, validation coverage_

- [ ] 11. Regenerate API documentation artifacts
  - Run documentation generation scripts to update all generated documentation files
  - Verify that generated HTML, Markdown, and TypeScript files reflect the field name changes
  - Update any manually maintained documentation that references the old field name
  - _Requirements: Documentation consistency, client integration support_

- [ ] 12. Run comprehensive test suite and fix any remaining issues
  - Execute `make test-fast` to run unit tests and identify any missed field references
  - Execute `make test-integration` to test database integration with new field names
  - Execute `make test-e2e` to validate complete API functionality
  - Fix any compilation errors or test failures discovered during testing
  - _Requirements: System functionality validation, regression prevention_