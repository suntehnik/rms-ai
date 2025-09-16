package init

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/config"
)

// TestErrorHandlingIntegration tests error handling with real configuration scenarios
func TestErrorHandlingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("configuration_error_scenario", func(t *testing.T) {
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

		// Set minimal required config to get past config.Load()
		os.Setenv("JWT_SECRET", "test-secret")
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_USER", "testuser")
		os.Setenv("DB_NAME", "testdb")

		// Create configuration - should succeed with minimal config
		cfg, err := config.Load()
		require.NoError(t, err)

		// Clear the admin password to trigger validation error in initialization
		os.Unsetenv("DEFAULT_ADMIN_PASSWORD")

		// Try to create initialization service - should fail with config error
		service, err := NewInitService(cfg)
		if err == nil {
			defer service.Close()
			err = service.Initialize()
		}

		require.Error(t, err)

		// Verify it's properly categorized as a configuration error
		exitCode := DetermineExitCode(err)
		assert.Equal(t, ExitConfigError, exitCode)

		// If it's an InitError, verify additional properties
		if initErr, ok := err.(*InitError); ok {
			assert.Equal(t, ErrorTypeConfig, initErr.Type)
			assert.Equal(t, SeverityCritical, initErr.Severity)
			assert.True(t, initErr.IsRecoverable())
			assert.Equal(t, ExitConfigError, initErr.GetExitCode())
			assert.NotEmpty(t, initErr.CorrelationID)
		}
	})

	t.Run("database_error_scenario", func(t *testing.T) {
		// Set up valid environment but invalid database
		os.Setenv("DB_HOST", "nonexistent-host-12345")
		os.Setenv("DB_PORT", "9999")
		os.Setenv("DB_USER", "testuser")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("JWT_SECRET", "test-jwt-secret")
		os.Setenv("DEFAULT_ADMIN_PASSWORD", "testpassword123")

		defer func() {
			os.Unsetenv("DB_HOST")
			os.Unsetenv("DB_PORT")
			os.Unsetenv("DB_USER")
			os.Unsetenv("DB_NAME")
			os.Unsetenv("JWT_SECRET")
			os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
		}()

		// Create configuration
		cfg, err := config.Load()
		require.NoError(t, err)

		// Try to create initialization service - should fail with database error
		service, err := NewInitService(cfg)
		if err == nil {
			defer service.Close()
			err = service.Initialize()
		}

		require.Error(t, err)

		// Verify it's properly categorized as a database error
		exitCode := DetermineExitCode(err)
		assert.Equal(t, ExitDatabaseError, exitCode)
	})

	t.Run("error_reporter_functionality", func(t *testing.T) {
		correlationID := "test-error-reporter-123"
		reporter := NewErrorReporter(correlationID)

		// Create various error types
		configErr := NewConfigError("Missing configuration", nil).
			WithStep("environment_validation").
			WithContext("missing_vars", []string{"DB_HOST"})

		dbErr := NewDatabaseError("Connection failed", nil).
			WithStep("database_connection").
			WithContext("host", "invalid-host")

		safetyErr := NewSafetyError("Database not empty", nil).
			WithStep("safety_check").
			WithContext("user_count", 5)

		// Report errors
		reporter.ReportError(configErr)
		reporter.ReportError(dbErr)
		reporter.ReportError(safetyErr)

		// Verify error collection
		assert.True(t, reporter.HasErrors())
		assert.Len(t, reporter.GetErrors(), 3)

		// Verify all errors have the correlation ID
		for _, err := range reporter.GetErrors() {
			assert.Equal(t, correlationID, err.CorrelationID)
		}

		// Generate comprehensive report
		report := reporter.GenerateReport()
		assert.Equal(t, correlationID, report["correlation_id"])
		assert.Equal(t, 3, report["error_count"])
		assert.NotNil(t, report["most_severe"])
		assert.NotNil(t, report["all_errors"])
		assert.NotNil(t, report["errors_by_type"])

		// Verify errors are grouped by type
		errorsByType := report["errors_by_type"].(map[ErrorType][]*InitError)
		assert.Len(t, errorsByType[ErrorTypeConfig], 1)
		assert.Len(t, errorsByType[ErrorTypeDatabase], 1)
		assert.Len(t, errorsByType[ErrorTypeSafety], 1)
	})

	t.Run("error_context_functionality", func(t *testing.T) {
		ctx := NewErrorContext("test-correlation", "test_step")
		ctx.AddData("operation", "database_connection")
		ctx.AddData("host", "localhost")
		ctx.AddData("port", 5432)
		ctx.AddData("timeout", "30s")

		// Complete the context
		ctx.Complete()

		// Verify context data
		contextMap := ctx.ToMap()
		assert.Equal(t, "test-correlation", contextMap["correlation_id"])
		assert.Equal(t, "test_step", contextMap["step"])
		assert.Equal(t, "database_connection", contextMap["operation"])
		assert.Equal(t, "localhost", contextMap["host"])
		assert.Equal(t, 5432, contextMap["port"])
		assert.Equal(t, "30s", contextMap["timeout"])
		assert.NotNil(t, contextMap["start_time"])
		assert.NotNil(t, contextMap["duration"])

		// Verify duration is set
		assert.True(t, ctx.Duration > 0)
	})

	t.Run("error_wrapping_functionality", func(t *testing.T) {
		// Create original error
		originalErr := assert.AnError

		// Create error context
		ctx := NewErrorContext("test-correlation", "test_step")
		ctx.AddData("additional_info", "test_value")
		ctx.Complete()

		// Wrap the error
		wrappedErr := WrapError(originalErr, ctx)

		require.NotNil(t, wrappedErr)
		assert.Equal(t, "test-correlation", wrappedErr.CorrelationID)
		assert.Equal(t, "test_step", wrappedErr.Step)
		assert.Equal(t, originalErr, wrappedErr.Cause)
		assert.Equal(t, "test_value", wrappedErr.Context["additional_info"])
		assert.NotNil(t, wrappedErr.Context["correlation_id"])
		assert.NotNil(t, wrappedErr.Context["duration"])

		// Verify exit code determination
		exitCode := DetermineExitCode(wrappedErr)
		assert.Equal(t, ExitSystemError, exitCode) // assert.AnError should map to system error
	})
}
