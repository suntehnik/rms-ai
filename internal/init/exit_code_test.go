package init

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExitCodeScenarios tests various error scenarios and their corresponding exit codes
func TestExitCodeScenarios(t *testing.T) {
	tests := []struct {
		name         string
		setupError   func() error
		expectedCode int
		description  string
	}{
		{
			name: "configuration_missing_db_host",
			setupError: func() error {
				return NewConfigError("Missing required environment variable: DB_HOST",
					errors.New("DB_HOST not set"))
			},
			expectedCode: ExitConfigError,
			description:  "Missing database host configuration should return config error code",
		},
		{
			name: "configuration_missing_jwt_secret",
			setupError: func() error {
				return NewConfigError("Missing required environment variable: JWT_SECRET",
					errors.New("JWT_SECRET not set"))
			},
			expectedCode: ExitConfigError,
			description:  "Missing JWT secret should return config error code",
		},
		{
			name: "configuration_invalid_admin_password",
			setupError: func() error {
				return NewConfigError("Invalid admin password: must be at least 8 characters",
					errors.New("password too short"))
			},
			expectedCode: ExitConfigError,
			description:  "Invalid admin password should return config error code",
		},
		{
			name: "database_connection_failed",
			setupError: func() error {
				return NewDatabaseError("Failed to connect to database",
					errors.New("connection refused"))
			},
			expectedCode: ExitDatabaseError,
			description:  "Database connection failure should return database error code",
		},
		{
			name: "database_ping_failed",
			setupError: func() error {
				return NewDatabaseError("Database ping failed",
					errors.New("context deadline exceeded"))
			},
			expectedCode: ExitDatabaseError,
			description:  "Database ping failure should return database error code",
		},
		{
			name: "database_permission_denied",
			setupError: func() error {
				return NewDatabaseError("Database permission denied",
					errors.New("role does not have permission"))
			},
			expectedCode: ExitDatabaseError,
			description:  "Database permission error should return database error code",
		},
		{
			name: "safety_database_not_empty",
			setupError: func() error {
				return NewSafetyError("Database contains existing data",
					errors.New("found 5 users, 3 epics"))
			},
			expectedCode: ExitSafetyError,
			description:  "Non-empty database should return safety error code",
		},
		{
			name: "safety_existing_admin_user",
			setupError: func() error {
				return NewSafetyError("Admin user already exists",
					errors.New("user with username 'admin' found"))
			},
			expectedCode: ExitSafetyError,
			description:  "Existing admin user should return safety error code",
		},
		{
			name: "migration_file_not_found",
			setupError: func() error {
				return NewMigrationError("Migration file not found",
					errors.New("no such file or directory"))
			},
			expectedCode: ExitMigrationError,
			description:  "Missing migration file should return migration error code",
		},
		{
			name: "migration_syntax_error",
			setupError: func() error {
				return NewMigrationError("Migration syntax error",
					errors.New("syntax error at line 5"))
			},
			expectedCode: ExitMigrationError,
			description:  "Migration syntax error should return migration error code",
		},
		{
			name: "migration_dirty_state",
			setupError: func() error {
				return NewMigrationError("Database in dirty state",
					errors.New("migration version 3 is dirty"))
			},
			expectedCode: ExitMigrationError,
			description:  "Dirty migration state should return migration error code",
		},
		{
			name: "user_creation_password_hash_failed",
			setupError: func() error {
				return NewCreationError("Failed to hash admin password",
					errors.New("bcrypt cost too high"))
			},
			expectedCode: ExitUserCreationError,
			description:  "Password hashing failure should return user creation error code",
		},
		{
			name: "user_creation_database_constraint",
			setupError: func() error {
				return NewCreationError("Failed to create admin user",
					errors.New("duplicate key value violates unique constraint"))
			},
			expectedCode: ExitUserCreationError,
			description:  "Database constraint violation should return user creation error code",
		},
		{
			name: "user_creation_transaction_failed",
			setupError: func() error {
				return NewCreationError("Transaction failed during user creation",
					errors.New("transaction rolled back"))
			},
			expectedCode: ExitUserCreationError,
			description:  "Transaction failure should return user creation error code",
		},
		{
			name: "system_out_of_memory",
			setupError: func() error {
				return NewSystemError("System out of memory",
					errors.New("cannot allocate memory"))
			},
			expectedCode: ExitSystemError,
			description:  "Out of memory should return system error code",
		},
		{
			name: "system_disk_full",
			setupError: func() error {
				return NewSystemError("Disk full",
					errors.New("no space left on device"))
			},
			expectedCode: ExitSystemError,
			description:  "Disk full should return system error code",
		},
		{
			name: "system_permission_denied",
			setupError: func() error {
				return NewSystemError("Permission denied",
					errors.New("operation not permitted"))
			},
			expectedCode: ExitSystemError,
			description:  "System permission denied should return system error code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setupError()
			exitCode := DetermineExitCode(err)

			assert.Equal(t, tt.expectedCode, exitCode, tt.description)

			// Verify the error is properly typed
			if initErr, ok := err.(*InitError); ok {
				assert.Equal(t, tt.expectedCode, initErr.GetExitCode())
			}
		})
	}
}

// TestExitCodeMapping tests the mapping between error types and exit codes
func TestExitCodeMapping(t *testing.T) {
	mappings := map[ErrorType]int{
		ErrorTypeConfig:    ExitConfigError,
		ErrorTypeDatabase:  ExitDatabaseError,
		ErrorTypeSafety:    ExitSafetyError,
		ErrorTypeMigration: ExitMigrationError,
		ErrorTypeCreation:  ExitUserCreationError,
		ErrorTypeSystem:    ExitSystemError,
	}

	for errorType, expectedCode := range mappings {
		t.Run(string(errorType), func(t *testing.T) {
			var err *InitError

			switch errorType {
			case ErrorTypeConfig:
				err = NewConfigError("Test config error", nil)
			case ErrorTypeDatabase:
				err = NewDatabaseError("Test database error", nil)
			case ErrorTypeSafety:
				err = NewSafetyError("Test safety error", nil)
			case ErrorTypeMigration:
				err = NewMigrationError("Test migration error", nil)
			case ErrorTypeCreation:
				err = NewCreationError("Test creation error", nil)
			case ErrorTypeSystem:
				err = NewSystemError("Test system error", nil)
			}

			require.NotNil(t, err)
			assert.Equal(t, expectedCode, err.GetExitCode())
			assert.Equal(t, expectedCode, DetermineExitCode(err))
		})
	}
}

// TestStringBasedExitCodeDetection tests exit code detection for non-InitError types
func TestStringBasedExitCodeDetection(t *testing.T) {
	tests := []struct {
		name         string
		errorMessage string
		expectedCode int
	}{
		{
			name:         "configuration_keyword",
			errorMessage: "configuration validation failed",
			expectedCode: ExitConfigError,
		},
		{
			name:         "environment_keyword",
			errorMessage: "environment variable missing",
			expectedCode: ExitConfigError,
		},
		{
			name:         "missing_keyword",
			errorMessage: "missing required parameter",
			expectedCode: ExitConfigError,
		},
		{
			name:         "invalid_keyword",
			errorMessage: "invalid configuration value",
			expectedCode: ExitConfigError,
		},
		{
			name:         "database_keyword",
			errorMessage: "database connection error",
			expectedCode: ExitDatabaseError,
		},
		{
			name:         "connection_keyword",
			errorMessage: "connection to server failed",
			expectedCode: ExitDatabaseError,
		},
		{
			name:         "postgres_keyword",
			errorMessage: "postgres server unavailable",
			expectedCode: ExitDatabaseError,
		},
		{
			name:         "safety_keyword",
			errorMessage: "safety check failed",
			expectedCode: ExitSafetyError,
		},
		{
			name:         "not_empty_keyword",
			errorMessage: "database not empty",
			expectedCode: ExitSafetyError,
		},
		{
			name:         "existing_data_keyword",
			errorMessage: "existing data found",
			expectedCode: ExitSafetyError,
		},
		{
			name:         "migration_keyword",
			errorMessage: "migration execution failed",
			expectedCode: ExitMigrationError,
		},
		{
			name:         "schema_keyword",
			errorMessage: "schema validation error",
			expectedCode: ExitMigrationError,
		},
		{
			name:         "user_keyword",
			errorMessage: "user creation failed",
			expectedCode: ExitUserCreationError,
		},
		{
			name:         "admin_keyword",
			errorMessage: "admin setup error",
			expectedCode: ExitUserCreationError,
		},
		{
			name:         "password_keyword",
			errorMessage: "password hashing failed",
			expectedCode: ExitUserCreationError,
		},
		{
			name:         "unknown_error",
			errorMessage: "something went wrong",
			expectedCode: ExitSystemError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errorMessage)
			exitCode := DetermineExitCode(err)

			assert.Equal(t, tt.expectedCode, exitCode)
		})
	}
}

// TestExitCodeConstants verifies the exit code constants have expected values
func TestExitCodeConstants(t *testing.T) {
	expectedCodes := map[string]int{
		"ExitSuccess":           0,
		"ExitConfigError":       1,
		"ExitDatabaseError":     2,
		"ExitSafetyError":       3,
		"ExitMigrationError":    4,
		"ExitUserCreationError": 5,
		"ExitSystemError":       10,
	}

	actualCodes := map[string]int{
		"ExitSuccess":           ExitSuccess,
		"ExitConfigError":       ExitConfigError,
		"ExitDatabaseError":     ExitDatabaseError,
		"ExitSafetyError":       ExitSafetyError,
		"ExitMigrationError":    ExitMigrationError,
		"ExitUserCreationError": ExitUserCreationError,
		"ExitSystemError":       ExitSystemError,
	}

	for name, expectedCode := range expectedCodes {
		t.Run(name, func(t *testing.T) {
			actualCode := actualCodes[name]
			assert.Equal(t, expectedCode, actualCode,
				"Exit code constant %s should have value %d", name, expectedCode)
		})
	}
}

// TestErrorRecoverability tests the recoverability flags for different error types
func TestErrorRecoverability(t *testing.T) {
	tests := []struct {
		name        string
		createError func() *InitError
		recoverable bool
	}{
		{
			name:        "config_error_recoverable",
			createError: func() *InitError { return NewConfigError("Config error", nil) },
			recoverable: true,
		},
		{
			name:        "database_error_recoverable",
			createError: func() *InitError { return NewDatabaseError("Database error", nil) },
			recoverable: true,
		},
		{
			name:        "safety_error_not_recoverable",
			createError: func() *InitError { return NewSafetyError("Safety error", nil) },
			recoverable: false,
		},
		{
			name:        "migration_error_recoverable",
			createError: func() *InitError { return NewMigrationError("Migration error", nil) },
			recoverable: true,
		},
		{
			name:        "creation_error_recoverable",
			createError: func() *InitError { return NewCreationError("Creation error", nil) },
			recoverable: true,
		},
		{
			name:        "system_error_not_recoverable",
			createError: func() *InitError { return NewSystemError("System error", nil) },
			recoverable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.createError()
			assert.Equal(t, tt.recoverable, err.IsRecoverable())
			assert.Equal(t, tt.recoverable, err.Recoverable)
		})
	}
}

// TestErrorSeverityLevels tests that all error types have appropriate severity levels
func TestErrorSeverityLevels(t *testing.T) {
	tests := []struct {
		name        string
		createError func() *InitError
		severity    ErrorSeverity
	}{
		{
			name:        "config_error_critical",
			createError: func() *InitError { return NewConfigError("Config error", nil) },
			severity:    SeverityCritical,
		},
		{
			name:        "database_error_critical",
			createError: func() *InitError { return NewDatabaseError("Database error", nil) },
			severity:    SeverityCritical,
		},
		{
			name:        "safety_error_critical",
			createError: func() *InitError { return NewSafetyError("Safety error", nil) },
			severity:    SeverityCritical,
		},
		{
			name:        "migration_error_critical",
			createError: func() *InitError { return NewMigrationError("Migration error", nil) },
			severity:    SeverityCritical,
		},
		{
			name:        "creation_error_critical",
			createError: func() *InitError { return NewCreationError("Creation error", nil) },
			severity:    SeverityCritical,
		},
		{
			name:        "system_error_critical",
			createError: func() *InitError { return NewSystemError("System error", nil) },
			severity:    SeverityCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.createError()
			assert.Equal(t, tt.severity, err.Severity)
		})
	}
}

// TestErrorContextIntegration tests error context integration with InitError
func TestErrorContextIntegration(t *testing.T) {
	correlationID := "test-correlation-123"
	step := "test_step"

	ctx := NewErrorContext(correlationID, step)
	ctx.AddData("key1", "value1")
	ctx.AddData("key2", 42)
	ctx.Complete()

	originalErr := errors.New("original error")
	wrappedErr := WrapError(originalErr, ctx)

	require.NotNil(t, wrappedErr)
	assert.Equal(t, correlationID, wrappedErr.CorrelationID)
	assert.Equal(t, step, wrappedErr.Step)
	assert.Equal(t, "value1", wrappedErr.Context["key1"])
	assert.Equal(t, 42, wrappedErr.Context["key2"])
	assert.NotNil(t, wrappedErr.Context["correlation_id"])
	assert.NotNil(t, wrappedErr.Context["duration"])
}

// MockInitService provides a mock implementation for testing error scenarios
type MockInitService struct {
	shouldFailAt string
	failureError error
}

func (m *MockInitService) simulateInitialization() error {
	steps := []string{
		"environment_validation",
		"database_connection",
		"database_health_check",
		"safety_check",
		"migration_execution",
		"admin_user_creation",
	}

	for _, step := range steps {
		if step == m.shouldFailAt {
			return m.failureError
		}
	}

	return nil
}

// TestInitializationFailureScenarios tests various initialization failure scenarios
func TestInitializationFailureScenarios(t *testing.T) {
	tests := []struct {
		name         string
		failureStep  string
		failureError error
		expectedCode int
	}{
		{
			name:         "environment_validation_failure",
			failureStep:  "environment_validation",
			failureError: NewConfigError("Missing DB_HOST", errors.New("env var not set")),
			expectedCode: ExitConfigError,
		},
		{
			name:         "database_connection_failure",
			failureStep:  "database_connection",
			failureError: NewDatabaseError("Connection failed", errors.New("connection refused")),
			expectedCode: ExitDatabaseError,
		},
		{
			name:         "safety_check_failure",
			failureStep:  "safety_check",
			failureError: NewSafetyError("Database not empty", errors.New("found existing data")),
			expectedCode: ExitSafetyError,
		},
		{
			name:         "migration_failure",
			failureStep:  "migration_execution",
			failureError: NewMigrationError("Migration failed", errors.New("syntax error")),
			expectedCode: ExitMigrationError,
		},
		{
			name:         "admin_creation_failure",
			failureStep:  "admin_user_creation",
			failureError: NewCreationError("User creation failed", errors.New("constraint violation")),
			expectedCode: ExitUserCreationError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockInitService{
				shouldFailAt: tt.failureStep,
				failureError: tt.failureError,
			}

			err := mockService.simulateInitialization()
			require.Error(t, err)

			exitCode := DetermineExitCode(err)
			assert.Equal(t, tt.expectedCode, exitCode)
		})
	}
}

// TestEnvironmentVariableValidation tests environment variable validation scenarios
func TestEnvironmentVariableValidation(t *testing.T) {
	// Save original environment
	originalEnv := map[string]string{
		"DB_HOST":                os.Getenv("DB_HOST"),
		"DB_USER":                os.Getenv("DB_USER"),
		"DB_NAME":                os.Getenv("DB_NAME"),
		"JWT_SECRET":             os.Getenv("JWT_SECRET"),
		"DEFAULT_ADMIN_PASSWORD": os.Getenv("DEFAULT_ADMIN_PASSWORD"),
	}

	// Restore environment after test
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	tests := []struct {
		name         string
		setupEnv     func()
		expectedCode int
		description  string
	}{
		{
			name: "missing_db_host",
			setupEnv: func() {
				os.Unsetenv("DB_HOST")
				os.Setenv("DB_USER", "test")
				os.Setenv("DB_NAME", "test")
				os.Setenv("JWT_SECRET", "test-secret")
				os.Setenv("DEFAULT_ADMIN_PASSWORD", "testpassword")
			},
			expectedCode: ExitConfigError,
			description:  "Missing DB_HOST should cause configuration error",
		},
		{
			name: "missing_jwt_secret",
			setupEnv: func() {
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_USER", "test")
				os.Setenv("DB_NAME", "test")
				os.Unsetenv("JWT_SECRET")
				os.Setenv("DEFAULT_ADMIN_PASSWORD", "testpassword")
			},
			expectedCode: ExitConfigError,
			description:  "Missing JWT_SECRET should cause configuration error",
		},
		{
			name: "missing_admin_password",
			setupEnv: func() {
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_USER", "test")
				os.Setenv("DB_NAME", "test")
				os.Setenv("JWT_SECRET", "test-secret")
				os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
			},
			expectedCode: ExitConfigError,
			description:  "Missing DEFAULT_ADMIN_PASSWORD should cause configuration error",
		},
		{
			name: "short_admin_password",
			setupEnv: func() {
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_USER", "test")
				os.Setenv("DB_NAME", "test")
				os.Setenv("JWT_SECRET", "test-secret")
				os.Setenv("DEFAULT_ADMIN_PASSWORD", "short") // Less than 8 characters
			},
			expectedCode: ExitConfigError,
			description:  "Short admin password should cause configuration error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()

			// Simulate environment validation error
			var err error
			switch tt.name {
			case "missing_db_host":
				err = NewConfigError("Missing required environment variables: [DB_HOST]",
					errors.New("DB_HOST not set"))
			case "missing_jwt_secret":
				err = NewConfigError("Missing required environment variables: [JWT_SECRET]",
					errors.New("JWT_SECRET not set"))
			case "missing_admin_password":
				err = NewConfigError("Missing required environment variables: [DEFAULT_ADMIN_PASSWORD]",
					errors.New("DEFAULT_ADMIN_PASSWORD not set"))
			case "short_admin_password":
				err = NewConfigError("DEFAULT_ADMIN_PASSWORD must be at least 8 characters long",
					errors.New("password too short"))
			}

			exitCode := DetermineExitCode(err)
			assert.Equal(t, tt.expectedCode, exitCode, tt.description)
		})
	}
}
