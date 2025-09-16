package init

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/config"
	"product-requirements-management/internal/logger"
)

func TestValidateEnvironmentLogging(t *testing.T) {
	capture := setupTestLogger()
	defer capture.hook.Reset()

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-testing",
		},
		Database: config.DatabaseConfig{
			Host:   "localhost",
			Port:   "5432",
			User:   "test",
			DBName: "test",
		},
	}

	service, err := NewInitService(cfg)
	require.NoError(t, err)

	// Set required environment variable
	os.Setenv("DEFAULT_ADMIN_PASSWORD", "testpassword123")
	defer os.Unsetenv("DEFAULT_ADMIN_PASSWORD")

	ctx := logger.WithInitializationStep(service.ctx, "environment_validation")

	// Test successful validation
	err = service.validateEnvironment(ctx)
	assert.NoError(t, err)

	entries, err := capture.parseLogEntries()
	require.NoError(t, err)

	// Find validation logs
	var startLog, progressLog, completedLog *LogEntry
	for _, entry := range entries {
		switch entry.Action {
		case "start_validation":
			startLog = &entry
		case "validation_progress":
			progressLog = &entry
		case "validation_completed":
			completedLog = &entry
		}
	}

	// Verify start log
	require.NotNil(t, startLog)
	assert.Equal(t, "info", startLog.Level)
	assert.Equal(t, "environment_validation", startLog.Step)
	assert.Contains(t, startLog.Message, "Validating environment configuration")

	// Verify progress log
	require.NotNil(t, progressLog)
	assert.Equal(t, "info", progressLog.Level)
	assert.Equal(t, "environment_validation", progressLog.Step)
	assert.Contains(t, progressLog.Fields, "variables_checked")
	assert.Contains(t, progressLog.Fields, "missing_variables")
	assert.Contains(t, progressLog.Fields, "invalid_variables")

	// Verify completion log
	require.NotNil(t, completedLog)
	assert.Equal(t, "info", completedLog.Level)
	assert.Equal(t, "environment_validation", completedLog.Step)
	assert.Equal(t, "success", completedLog.Status)
	assert.NotEmpty(t, completedLog.Duration)
}

func TestValidateEnvironmentFailureLogging(t *testing.T) {
	capture := setupTestLogger()
	defer capture.hook.Reset()

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "", // Missing JWT secret
		},
		Database: config.DatabaseConfig{
			Host:   "", // Missing host
			Port:   "5432",
			User:   "test",
			DBName: "test",
		},
	}

	service, err := NewInitService(cfg)
	require.NoError(t, err)

	// Don't set DEFAULT_ADMIN_PASSWORD to trigger failure
	os.Unsetenv("DEFAULT_ADMIN_PASSWORD")

	ctx := logger.WithInitializationStep(service.ctx, "environment_validation")

	// Test failed validation
	err = service.validateEnvironment(ctx)
	assert.Error(t, err)

	entries, err := capture.parseLogEntries()
	require.NoError(t, err)

	// Find validation failure log
	var failureLog *LogEntry
	for _, entry := range entries {
		if entry.Action == "validation_failed" {
			failureLog = &entry
			break
		}
	}

	require.NotNil(t, failureLog)
	assert.Equal(t, "error", failureLog.Level)
	assert.Equal(t, "environment_validation", failureLog.Step)
	assert.Contains(t, failureLog.Fields, "missing_variables")

	// Check that missing variables are logged
	missingVars, ok := failureLog.Fields["missing_variables"].([]interface{})
	require.True(t, ok)
	assert.Contains(t, missingVars, "DB_HOST")
	assert.Contains(t, missingVars, "JWT_SECRET")
	assert.Contains(t, missingVars, "DEFAULT_ADMIN_PASSWORD")
}

func TestStepTimingAccuracy(t *testing.T) {
	capture := setupTestLogger()
	defer capture.hook.Reset()

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-testing",
		},
		Database: config.DatabaseConfig{
			Host:   "localhost",
			Port:   "5432",
			User:   "test",
			DBName: "test",
		},
	}

	service, err := NewInitService(cfg)
	require.NoError(t, err)

	os.Setenv("DEFAULT_ADMIN_PASSWORD", "testpassword123")
	defer os.Unsetenv("DEFAULT_ADMIN_PASSWORD")

	ctx := logger.WithInitializationStep(service.ctx, "environment_validation")

	startTime := time.Now()
	err = service.validateEnvironment(ctx)
	endTime := time.Now()
	actualDuration := endTime.Sub(startTime)

	assert.NoError(t, err)

	entries, err := capture.parseLogEntries()
	require.NoError(t, err)

	// Find completion log with duration
	var completedLog *LogEntry
	for _, entry := range entries {
		if entry.Action == "validation_completed" && entry.Duration != "" {
			completedLog = &entry
			break
		}
	}

	require.NotNil(t, completedLog)
	assert.NotEmpty(t, completedLog.Duration)

	// Parse duration and verify it's reasonable
	loggedDuration, err := time.ParseDuration(completedLog.Duration)
	require.NoError(t, err)

	// Duration should be within reasonable bounds (allowing for some overhead)
	assert.True(t, loggedDuration >= actualDuration/2, "Logged duration too small")
	assert.True(t, loggedDuration <= actualDuration*2, "Logged duration too large")
}

func TestCorrelationIDConsistency(t *testing.T) {
	capture := setupTestLogger()
	defer capture.hook.Reset()

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-testing",
		},
		Database: config.DatabaseConfig{
			Host:   "localhost",
			Port:   "5432",
			User:   "test",
			DBName: "test",
		},
	}

	service, err := NewInitService(cfg)
	require.NoError(t, err)

	os.Setenv("DEFAULT_ADMIN_PASSWORD", "testpassword123")
	defer os.Unsetenv("DEFAULT_ADMIN_PASSWORD")

	ctx := logger.WithInitializationStep(service.ctx, "environment_validation")
	err = service.validateEnvironment(ctx)
	assert.NoError(t, err)

	entries, err := capture.parseLogEntries()
	require.NoError(t, err)

	// Verify all entries have the same correlation ID
	var correlationID string
	for i, entry := range entries {
		if entry.CorrelationID != "" {
			if correlationID == "" {
				correlationID = entry.CorrelationID
			} else {
				assert.Equal(t, correlationID, entry.CorrelationID,
					"Entry %d has different correlation ID", i)
			}
		}
	}

	assert.NotEmpty(t, correlationID, "No correlation ID found in log entries")
	assert.Equal(t, service.correlationID, correlationID,
		"Service correlation ID doesn't match logged correlation ID")
}

func TestLogStepFailureDetails(t *testing.T) {
	capture := setupTestLogger()
	defer capture.hook.Reset()

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-testing",
		},
	}

	service, err := NewInitService(cfg)
	require.NoError(t, err)

	stepStart := time.Now()
	time.Sleep(50 * time.Millisecond) // Ensure measurable duration
	testError := assert.AnError

	service.logStepFailure("test-step-failure", stepStart, testError)

	entries, err := capture.parseLogEntries()
	require.NoError(t, err)

	// Find the step failure log
	var failureLog *LogEntry
	for _, entry := range entries {
		if entry.Action == "step_failed" && entry.Step == "test-step-failure" {
			failureLog = &entry
			break
		}
	}

	require.NotNil(t, failureLog)
	assert.Equal(t, "error", failureLog.Level)
	assert.Equal(t, "test-step-failure", failureLog.Step)
	assert.Equal(t, "failed", failureLog.Status)
	assert.NotEmpty(t, failureLog.Duration)
	assert.Equal(t, testError.Error(), failureLog.Error)
	assert.Contains(t, failureLog.Message, "test-step-failure")
	assert.Contains(t, failureLog.Message, "failed")

	// Verify duration is reasonable
	duration, err := time.ParseDuration(failureLog.Duration)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, duration, 40*time.Millisecond)
	assert.LessOrEqual(t, duration, 200*time.Millisecond)
}

func TestInitializationSummaryLogging(t *testing.T) {
	capture := setupTestLogger()
	defer capture.hook.Reset()

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-testing",
		},
		Database: config.DatabaseConfig{
			Host:   "test-host",
			Port:   "5432",
			User:   "test",
			DBName: "test-db",
		},
	}

	service, err := NewInitService(cfg)
	require.NoError(t, err)

	// Create test step summaries
	stepSummaries := []StepSummary{
		{
			Name:      "step1",
			StartTime: time.Now().Add(-2 * time.Minute),
			EndTime:   time.Now().Add(-1 * time.Minute),
			Duration:  1 * time.Minute,
			Status:    "success",
			Details:   map[string]interface{}{"test": "value1"},
		},
		{
			Name:      "step2",
			StartTime: time.Now().Add(-1 * time.Minute),
			EndTime:   time.Now(),
			Duration:  1 * time.Minute,
			Status:    "success",
			Details:   map[string]interface{}{"migrations": 3},
		},
	}

	service.logSuccessAndNextSteps(stepSummaries, "admin", 3)

	entries, err := capture.parseLogEntries()
	require.NoError(t, err)

	// Find the completion summary log
	var summaryLog *LogEntry
	for _, entry := range entries {
		if entry.Action == "initialization_completed" {
			summaryLog = &entry
			break
		}
	}

	require.NotNil(t, summaryLog)
	assert.Equal(t, "info", summaryLog.Level)
	assert.Equal(t, "success", summaryLog.Status)
	assert.Contains(t, summaryLog.Fields, "total_duration")
	assert.NotEmpty(t, summaryLog.Fields["total_duration"])
	assert.Contains(t, summaryLog.Fields, "steps_completed")
	assert.Contains(t, summaryLog.Fields, "admin_username")
	assert.Contains(t, summaryLog.Fields, "migrations_applied")
	assert.Contains(t, summaryLog.Fields, "database_host")
	assert.Contains(t, summaryLog.Fields, "database_name")
	assert.Contains(t, summaryLog.Fields, "summary")

	// Verify field values
	assert.Equal(t, float64(2), summaryLog.Fields["steps_completed"])
	assert.Equal(t, "admin", summaryLog.Fields["admin_username"])
	assert.Equal(t, float64(3), summaryLog.Fields["migrations_applied"])
	assert.Equal(t, "test-host", summaryLog.Fields["database_host"])
	assert.Equal(t, "test-db", summaryLog.Fields["database_name"])

	// Find step detail logs
	var stepDetailLogs []LogEntry
	for _, entry := range entries {
		if entry.Action == "step_detail" {
			stepDetailLogs = append(stepDetailLogs, entry)
		}
	}

	assert.Len(t, stepDetailLogs, 2)
	for i, stepLog := range stepDetailLogs {
		assert.Equal(t, "info", stepLog.Level)
		assert.Contains(t, stepLog.Fields, "step_number")
		assert.Contains(t, stepLog.Fields, "step_name")
		assert.Contains(t, stepLog.Fields, "duration")
		assert.Contains(t, stepLog.Fields, "status")
		assert.Contains(t, stepLog.Fields, "details")
		assert.Equal(t, float64(i+1), stepLog.Fields["step_number"])
	}

	// Find next steps logs
	var nextStepLogs []LogEntry
	for _, entry := range entries {
		if entry.Action == "next_step" {
			nextStepLogs = append(nextStepLogs, entry)
		}
	}

	assert.Len(t, nextStepLogs, 4) // Should have 4 next steps
	for i, nextStepLog := range nextStepLogs {
		assert.Equal(t, "info", nextStepLog.Level)
		assert.Contains(t, nextStepLog.Fields, "step_number")
		assert.Contains(t, nextStepLog.Fields, "instruction")
		assert.Equal(t, float64(i+1), nextStepLog.Fields["step_number"])
	}
}

func TestLogFieldConsistency(t *testing.T) {
	capture := setupTestLogger()
	defer capture.hook.Reset()

	correlationID := logger.NewCorrelationID()
	ctx := logger.WithCorrelationID(context.Background(), correlationID)
	ctx = logger.WithInitializationStep(ctx, "test-step")

	// Test various logging scenarios
	testCases := []struct {
		name   string
		fields map[string]interface{}
		level  string
	}{
		{
			name: "info_with_duration",
			fields: map[string]interface{}{
				"component": "test",
				"action":    "test_action",
				"duration":  "100ms",
				"status":    "success",
			},
			level: "info",
		},
		{
			name: "error_with_context",
			fields: map[string]interface{}{
				"component": "test",
				"action":    "test_error",
				"error":     "test error message",
			},
			level: "error",
		},
		{
			name: "debug_with_details",
			fields: map[string]interface{}{
				"component": "test",
				"action":    "test_debug",
				"details":   map[string]interface{}{"key": "value"},
			},
			level: "debug",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			switch tc.level {
			case "info":
				logger.WithContextAndFields(ctx, tc.fields).Info("Test message")
			case "error":
				logger.WithContextAndFields(ctx, tc.fields).Error("Test message")
			case "debug":
				logger.WithContextAndFields(ctx, tc.fields).Debug("Test message")
			}
		})
	}

	entries, err := capture.parseLogEntries()
	require.NoError(t, err)
	require.Len(t, entries, len(testCases))

	// Verify all entries have consistent context fields
	for i, entry := range entries {
		assert.Equal(t, correlationID, entry.CorrelationID,
			"Entry %d missing correlation ID", i)
		assert.Equal(t, "test-step", entry.Step,
			"Entry %d missing step", i)
		assert.Equal(t, "test", entry.Component,
			"Entry %d missing component", i)
		assert.NotEmpty(t, entry.Action,
			"Entry %d missing action", i)
		assert.Equal(t, testCases[i].level, entry.Level,
			"Entry %d has wrong level", i)
	}
}
