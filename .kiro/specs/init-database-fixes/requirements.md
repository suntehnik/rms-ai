# Requirements Document

## Introduction

This document describes a set of requirements for fixing critical errors in the initialization service and eliminating architectural anti-patterns in database connection management throughout the project. The goal is to ensure stability, predictability, and proper resource management.

The primary issue is that the initialization service fails when running on a database without tables, as it tries to check for data in non-existent tables. Additionally, various system components (migrator, scripts) independently manage database connection lifecycles, often closing connections prematurely or not closing them at all, leading to "database already closed" errors and resource leaks.

## Requirements

### Requirement 1: Correct Initialization Service Operation on Empty Database

**Problem:** The initialization service fails when starting on a database without tables, as it tries to check for data in non-existent tables.

**User Story:** As a system administrator, I want the initialization service to work correctly with a completely empty database (without tables), so that I can deploy the system from scratch without errors.

#### Acceptance Criteria

1.1. WHEN the initialization service performs a safety check THEN it SHALL correctly handle "table not found" errors (e.g., SQLSTATE 42P01 for PostgreSQL).

1.2. WHEN a "table not found" error is received during table checking THEN the system SHALL consider the record count in that table as zero and continue checking, rather than terminating with an error.

1.3. IF checking all tables shows they are either missing or empty THEN the system SHALL consider the database safe for initialization and proceed to the migration step.

1.4. IF any other database error occurs during checking (connection issues, access rights, etc.) THEN the system SHALL terminate with the corresponding error.

### Requirement 2: Centralized Database Connection Lifecycle Management

**Problem:** Various system components (migrator, scripts) independently manage database connection lifecycles, often closing connections prematurely or not closing them at all. This leads to "database already closed" errors and resource leaks.

**User Story:** As a developer, I want database connection lifecycle to be managed centrally, so that I can avoid errors and resource leaks.

#### Acceptance Criteria

2.1. **Migration Module** (`internal/database/migrate.go`):
- The module SHALL be modified to accept an existing database connection object (*gorm.DB) as an argument.
- The module SHALL NOT open its own connection and SHALL NOT close the connection passed to it.

2.2. **Initialization Service** (`internal/init/service.go`):
- The service SHALL create one database connection during its initialization.
- This connection SHALL be passed to all internal components (Safety Checker, Migrator, Admin Creator).
- Connection closure SHALL happen centrally when the service terminates (as currently implemented through defer service.Close() in cmd/init/main.go).

2.3. **Model Verification Script** (`scripts/verify_models.go`):
- In the main function of this script, deferred closure (defer) of the database connection SHALL be added immediately after its successful creation to guarantee resource cleanup.

2.4. **Repository Usage Example** (`internal/repository/example_usage.go`):
- In the ExampleUsage function, deferred closure (defer) of the database connection SHALL be added to demonstrate the correct resource management pattern.

2.5. **Code Audit**:
- An audit SHALL be conducted of all other code (including tests and benchmarks) for the presence of this anti-pattern. Any component that creates a connection to perform a task must also ensure its guaranteed closure. Any component that receives a connection from outside should not close it.

### Requirement 3: Robust Error Handling and Logging

**User Story:** As a system administrator, I want clear error messages and appropriate logging when database operations fail, so that I can understand what went wrong and take appropriate action.

#### Acceptance Criteria

3.1. WHEN a "table not found" error occurs THEN the system SHALL log this as a debug message and continue processing.

3.2. WHEN other database errors occur THEN the system SHALL provide clear error messages indicating the specific problem.

3.3. WHEN the database contains existing data THEN the system SHALL provide a clear message indicating which tables contain data and exit safely.

3.4. WHEN database connection fails THEN the system SHALL provide connection-specific error information.

### Requirement 4: Backward Compatibility

**User Story:** As a developer, I want the updated system to maintain existing behavior for populated databases, so that existing safety mechanisms and workflows continue to work.

#### Acceptance Criteria

4.1. WHEN running against a database with existing tables and data THEN the system SHALL behave exactly as before the fix.

4.2. WHEN running against a database with empty tables THEN the system SHALL recognize this as safe for initialization.

4.3. WHEN safety checks pass THEN the system SHALL proceed with the same migration and admin user creation process as before.

4.4. WHEN database connections are managed centrally THEN existing functionality SHALL remain unchanged from the user perspective.