package init

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitError_Error(t *testing.T) {
	tests := []struct {
		name     string
		initErr  *InitError
		expected string
	}{
		{
			name: "error with cause",
			initErr: &InitError{
				Type:     ErrorTypeConfig,
				Severity: SeverityCritical,
				Message:  "Configuration validation failed",
				Cause:    errors.New("missing DB_HOST"),
				Step:     "environment_validation",
			},
			expected: "[configuration:critical] Step 'environment_validation': Configuration validation failed (caused by: missing DB_HOST)",
		},
		{
			name: "error without cause",
			initErr: &InitError{
				Type:     ErrorTypeDatabase,
				Severity: SeverityHigh,
				Message:  "Connection timeout",
			},
			expected: "[database:high] Connection timeout",
		},
		{
			name: "error without step",
			initErr: &InitError{
				Type:     ErrorTypeSafety,
				Severity: SeverityCritical,
				Message:  "Database not empty",
				Cause:    errors.New("found 5 users"),
			},
			expected: "[safety:critical] Database not empty (caused by: found 5 users)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.initErr.Error())
		})
	}
}

func TestInitError_JSON(t *testing.T) {
	initErr := &InitError{
		Type:          ErrorTypeConfig,
		Severity:      SeverityCritical,
		Message:       "Test error",
		Timestamp:     time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		CorrelationID: "test-123",
		Step:          "test_step",
		Recoverable:   true,
		ExitCode:      ExitConfigError,
		Context: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}

	jsonStr := initErr.JSON()

	// Verify it's valid JSON
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &parsed)
	require.NoError(t, err)

	// Verify key fields
	assert.Equal(t, "configuration", parsed["type"])
	assert.Equal(t, "critical", parsed["severity"])
	assert.Equal(t, "Test error", parsed["message"])
	assert.Equal(t, "test-123", parsed["correlation_id"])
	assert.Equal(t, "test_step", parsed["step"])
	assert.Equal(t, true, parsed["recoverable"])
	assert.Equal(t, float64(ExitConfigError), parsed["exit_code"])
}

func TestInitError_WithContext(t *testing.T) {
	initErr := NewConfigError("Test error", nil)

	result := initErr.WithContext("key1", "value1").WithContext("key2", 42)

	assert.Equal(t, "value1", result.Context["key1"])
	assert.Equal(t, 42, result.Context["key2"])
	assert.Same(t, initErr, result) // Should return same instance
}

func TestInitError_WithCorrelationID(t *testing.T) {
	initErr := NewConfigError("Test error", nil)

	result := initErr.WithCorrelationID("test-correlation-123")

	assert.Equal(t, "test-correlation-123", result.CorrelationID)
	assert.Same(t, initErr, result) // Should return same instance
}

func TestInitError_WithStep(t *testing.T) {
	initErr := NewConfigError("Test error", nil)

	result := initErr.WithStep("test_step")

	assert.Equal(t, "test_step", result.Step)
	assert.Same(t, initErr, result) // Should return same instance
}

func TestNewConfigError(t *testing.T) {
	cause := errors.New("underlying error")
	initErr := NewConfigError("Configuration failed", cause)

	assert.Equal(t, ErrorTypeConfig, initErr.Type)
	assert.Equal(t, SeverityCritical, initErr.Severity)
	assert.Equal(t, "Configuration failed", initErr.Message)
	assert.Equal(t, cause, initErr.Cause)
	assert.True(t, initErr.Recoverable)
	assert.Equal(t, ExitConfigError, initErr.ExitCode)
	assert.NotNil(t, initErr.Context)
	assert.WithinDuration(t, time.Now(), initErr.Timestamp, time.Second)
}

func TestNewDatabaseError(t *testing.T) {
	cause := errors.New("connection failed")
	initErr := NewDatabaseError("Database error", cause)

	assert.Equal(t, ErrorTypeDatabase, initErr.Type)
	assert.Equal(t, SeverityCritical, initErr.Severity)
	assert.Equal(t, "Database error", initErr.Message)
	assert.Equal(t, cause, initErr.Cause)
	assert.True(t, initErr.Recoverable)
	assert.Equal(t, ExitDatabaseError, initErr.ExitCode)
}

func TestNewSafetyError(t *testing.T) {
	cause := errors.New("database not empty")
	initErr := NewSafetyError("Safety check failed", cause)

	assert.Equal(t, ErrorTypeSafety, initErr.Type)
	assert.Equal(t, SeverityCritical, initErr.Severity)
	assert.Equal(t, "Safety check failed", initErr.Message)
	assert.Equal(t, cause, initErr.Cause)
	assert.False(t, initErr.Recoverable) // Safety errors are not recoverable
	assert.Equal(t, ExitSafetyError, initErr.ExitCode)
}

func TestNewMigrationError(t *testing.T) {
	cause := errors.New("migration failed")
	initErr := NewMigrationError("Migration error", cause)

	assert.Equal(t, ErrorTypeMigration, initErr.Type)
	assert.Equal(t, SeverityCritical, initErr.Severity)
	assert.Equal(t, "Migration error", initErr.Message)
	assert.Equal(t, cause, initErr.Cause)
	assert.True(t, initErr.Recoverable)
	assert.Equal(t, ExitMigrationError, initErr.ExitCode)
}

func TestNewCreationError(t *testing.T) {
	cause := errors.New("user creation failed")
	initErr := NewCreationError("Creation error", cause)

	assert.Equal(t, ErrorTypeCreation, initErr.Type)
	assert.Equal(t, SeverityCritical, initErr.Severity)
	assert.Equal(t, "Creation error", initErr.Message)
	assert.Equal(t, cause, initErr.Cause)
	assert.True(t, initErr.Recoverable)
	assert.Equal(t, ExitUserCreationError, initErr.ExitCode)
}

func TestNewSystemError(t *testing.T) {
	cause := errors.New("system failure")
	initErr := NewSystemError("System error", cause)

	assert.Equal(t, ErrorTypeSystem, initErr.Type)
	assert.Equal(t, SeverityCritical, initErr.Severity)
	assert.Equal(t, "System error", initErr.Message)
	assert.Equal(t, cause, initErr.Cause)
	assert.False(t, initErr.Recoverable) // System errors are not recoverable
	assert.Equal(t, ExitSystemError, initErr.ExitCode)
}

func TestErrorContext_NewErrorContext(t *testing.T) {
	ctx := NewErrorContext("test-correlation", "test_step")

	assert.Equal(t, "test-correlation", ctx.CorrelationID)
	assert.Equal(t, "test_step", ctx.Step)
	assert.WithinDuration(t, time.Now(), ctx.StartTime, time.Second)
	assert.NotNil(t, ctx.Data)
	assert.Empty(t, ctx.Data)
	assert.Zero(t, ctx.Duration)
}

func TestErrorContext_AddData(t *testing.T) {
	ctx := NewErrorContext("test-correlation", "test_step")

	result := ctx.AddData("key1", "value1").AddData("key2", 42)

	assert.Equal(t, "value1", ctx.Data["key1"])
	assert.Equal(t, 42, ctx.Data["key2"])
	assert.Same(t, ctx, result) // Should return same instance
}

func TestErrorContext_Complete(t *testing.T) {
	ctx := NewErrorContext("test-correlation", "test_step")

	// Wait a small amount to ensure duration > 0
	time.Sleep(time.Millisecond)

	result := ctx.Complete()

	assert.True(t, ctx.Duration > 0)
	assert.Same(t, ctx, result) // Should return same instance
}

func TestErrorContext_ToMap(t *testing.T) {
	ctx := NewErrorContext("test-correlation", "test_step")
	ctx.AddData("key1", "value1")
	ctx.AddData("key2", 42)
	ctx.Complete()

	result := ctx.ToMap()

	assert.Equal(t, "test-correlation", result["correlation_id"])
	assert.Equal(t, "test_step", result["step"])
	assert.NotNil(t, result["start_time"])
	assert.NotNil(t, result["duration"])
	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, 42, result["key2"])
}

func TestWrapError_NilError(t *testing.T) {
	result := WrapError(nil, nil)
	assert.Nil(t, result)
}

func TestWrapError_ExistingInitError(t *testing.T) {
	originalErr := NewConfigError("Original error", nil)
	ctx := NewErrorContext("test-correlation", "test_step")
	ctx.AddData("key1", "value1")

	result := WrapError(originalErr, ctx)

	assert.Same(t, originalErr, result)
	assert.Equal(t, "test-correlation", result.CorrelationID)
	assert.Equal(t, "test_step", result.Step)
	assert.Equal(t, "value1", result.Context["key1"])
}

func TestWrapError_RegularError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedType ErrorType
		expectedCode int
	}{
		{
			name:         "configuration error",
			err:          errors.New("missing configuration variable"),
			expectedType: ErrorTypeConfig,
			expectedCode: ExitConfigError,
		},
		{
			name:         "database error",
			err:          errors.New("database connection failed"),
			expectedType: ErrorTypeDatabase,
			expectedCode: ExitDatabaseError,
		},
		{
			name:         "safety error",
			err:          errors.New("database not empty"),
			expectedType: ErrorTypeSafety,
			expectedCode: ExitSafetyError,
		},
		{
			name:         "migration error",
			err:          errors.New("migration failed"),
			expectedType: ErrorTypeMigration,
			expectedCode: ExitMigrationError,
		},
		{
			name:         "user creation error",
			err:          errors.New("admin user creation failed"),
			expectedType: ErrorTypeCreation,
			expectedCode: ExitUserCreationError,
		},
		{
			name:         "system error",
			err:          errors.New("unknown system failure"),
			expectedType: ErrorTypeSystem,
			expectedCode: ExitSystemError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewErrorContext("test-correlation", "test_step")

			result := WrapError(tt.err, ctx)

			assert.Equal(t, tt.expectedType, result.Type)
			assert.Equal(t, tt.expectedCode, result.ExitCode)
			assert.Equal(t, tt.err, result.Cause)
			assert.Equal(t, "test-correlation", result.CorrelationID)
			assert.Equal(t, "test_step", result.Step)
		})
	}
}

func TestDetermineExitCode(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{
			name:         "nil error",
			err:          nil,
			expectedCode: ExitSuccess,
		},
		{
			name:         "InitError",
			err:          NewConfigError("Config error", nil),
			expectedCode: ExitConfigError,
		},
		{
			name:         "configuration error string",
			err:          errors.New("missing configuration"),
			expectedCode: ExitConfigError,
		},
		{
			name:         "database error string",
			err:          errors.New("database connection failed"),
			expectedCode: ExitDatabaseError,
		},
		{
			name:         "safety error string",
			err:          errors.New("safety check failed - not empty"),
			expectedCode: ExitSafetyError,
		},
		{
			name:         "migration error string",
			err:          errors.New("migration execution failed"),
			expectedCode: ExitMigrationError,
		},
		{
			name:         "user creation error string",
			err:          errors.New("admin user creation failed"),
			expectedCode: ExitUserCreationError,
		},
		{
			name:         "unknown error",
			err:          errors.New("some unknown error"),
			expectedCode: ExitSystemError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineExitCode(tt.err)
			assert.Equal(t, tt.expectedCode, result)
		})
	}
}

func TestErrorReporter_NewErrorReporter(t *testing.T) {
	reporter := NewErrorReporter("test-correlation")

	assert.Equal(t, "test-correlation", reporter.correlationID)
	assert.Empty(t, reporter.errors)
	assert.False(t, reporter.HasErrors())
}

func TestErrorReporter_ReportError(t *testing.T) {
	reporter := NewErrorReporter("test-correlation")
	err1 := NewConfigError("Error 1", nil)
	err2 := NewDatabaseError("Error 2", nil)

	reporter.ReportError(err1)
	reporter.ReportError(err2)
	reporter.ReportError(nil) // Should be ignored

	assert.True(t, reporter.HasErrors())
	assert.Len(t, reporter.GetErrors(), 2)
	assert.Equal(t, "test-correlation", err1.CorrelationID)
	assert.Equal(t, "test-correlation", err2.CorrelationID)
}

func TestErrorReporter_GetMostSevereError(t *testing.T) {
	reporter := NewErrorReporter("test-correlation")

	// Test with no errors
	assert.Nil(t, reporter.GetMostSevereError())

	// Add errors with different severities
	err1 := NewConfigError("Error 1", nil)
	err1.Severity = SeverityLow

	err2 := NewDatabaseError("Error 2", nil)
	err2.Severity = SeverityCritical

	err3 := NewMigrationError("Error 3", nil)
	err3.Severity = SeverityMedium

	reporter.ReportError(err1)
	reporter.ReportError(err2)
	reporter.ReportError(err3)

	mostSevere := reporter.GetMostSevereError()
	assert.Equal(t, err2, mostSevere)
	assert.Equal(t, SeverityCritical, mostSevere.Severity)
}

func TestErrorReporter_GenerateReport(t *testing.T) {
	reporter := NewErrorReporter("test-correlation")

	// Test empty report
	report := reporter.GenerateReport()
	assert.Equal(t, "test-correlation", report["correlation_id"])
	assert.Equal(t, 0, report["error_count"])
	assert.NotNil(t, report["timestamp"])

	// Add some errors
	err1 := NewConfigError("Config error", nil)
	err2 := NewDatabaseError("DB error", nil)
	err3 := NewConfigError("Another config error", nil)

	reporter.ReportError(err1)
	reporter.ReportError(err2)
	reporter.ReportError(err3)

	report = reporter.GenerateReport()
	assert.Equal(t, 3, report["error_count"])
	assert.NotNil(t, report["most_severe"])
	assert.NotNil(t, report["all_errors"])
	assert.NotNil(t, report["errors_by_type"])

	// Check errors by type grouping
	errorsByType := report["errors_by_type"].(map[ErrorType][]*InitError)
	assert.Len(t, errorsByType[ErrorTypeConfig], 2)
	assert.Len(t, errorsByType[ErrorTypeDatabase], 1)
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name       string
		str        string
		substrings []string
		expected   bool
	}{
		{
			name:       "contains one substring",
			str:        "database connection failed",
			substrings: []string{"database", "config"},
			expected:   true,
		},
		{
			name:       "contains multiple substrings",
			str:        "database configuration error",
			substrings: []string{"database", "config"},
			expected:   true,
		},
		{
			name:       "contains no substrings",
			str:        "unknown error occurred",
			substrings: []string{"database", "config"},
			expected:   false,
		},
		{
			name:       "empty string",
			str:        "",
			substrings: []string{"database", "config"},
			expected:   false,
		},
		{
			name:       "empty substrings",
			str:        "some error",
			substrings: []string{},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsAny(tt.str, tt.substrings...)
			assert.Equal(t, tt.expected, result)
		})
	}
}
