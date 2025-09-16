package init

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/config"
	"product-requirements-management/internal/logger"
)

// LogEntry represents a parsed log entry for testing
type LogEntry struct {
	Level         string                 `json:"level"`
	Message       string                 `json:"msg"`
	Time          string                 `json:"time"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Step          string                 `json:"step,omitempty"`
	Component     string                 `json:"component,omitempty"`
	Action        string                 `json:"action,omitempty"`
	Duration      string                 `json:"duration,omitempty"`
	Status        string                 `json:"status,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Fields        map[string]interface{} `json:"-"`
}

// TestLogCapture captures log output for testing
type TestLogCapture struct {
	buffer *bytes.Buffer
	hook   *TestLogHook
}

// TestLogHook captures log entries for testing
type TestLogHook struct {
	entries []logrus.Entry
}

func (h *TestLogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *TestLogHook) Fire(entry *logrus.Entry) error {
	h.entries = append(h.entries, *entry)
	return nil
}

func (h *TestLogHook) GetEntries() []logrus.Entry {
	return h.entries
}

func (h *TestLogHook) Reset() {
	h.entries = nil
}

// setupTestLogger sets up a logger for testing with JSON output
func setupTestLogger() *TestLogCapture {
	buffer := &bytes.Buffer{}
	hook := &TestLogHook{}

	// Configure logger for testing
	testLogger := logrus.New()
	testLogger.SetOutput(buffer)
	testLogger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
	testLogger.SetLevel(logrus.DebugLevel)
	testLogger.AddHook(hook)

	// Replace the global logger for testing
	logger.Logger = testLogger

	return &TestLogCapture{
		buffer: buffer,
		hook:   hook,
	}
}

// parseLogEntries parses JSON log entries from the buffer
func (tc *TestLogCapture) parseLogEntries() ([]LogEntry, error) {
	lines := strings.Split(strings.TrimSpace(tc.buffer.String()), "\n")
	var entries []LogEntry

	for _, line := range lines {
		if line == "" {
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, err
		}

		// Parse additional fields
		var rawEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &rawEntry); err == nil {
			entry.Fields = rawEntry
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func TestCorrelationIDGeneration(t *testing.T) {
	// Test correlation ID generation
	id1 := logger.NewCorrelationID()
	id2 := logger.NewCorrelationID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Len(t, id1, 36) // UUID length
}

func TestContextWithCorrelationID(t *testing.T) {
	ctx := context.Background()
	correlationID := "test-correlation-id"

	// Add correlation ID to context
	ctxWithID := logger.WithCorrelationID(ctx, correlationID)

	// Retrieve correlation ID from context
	retrievedID := logger.GetCorrelationID(ctxWithID)
	assert.Equal(t, correlationID, retrievedID)

	// Test empty context
	emptyID := logger.GetCorrelationID(ctx)
	assert.Empty(t, emptyID)
}

func TestContextWithInitializationStep(t *testing.T) {
	ctx := context.Background()
	step := "test-step"

	// Add step to context
	ctxWithStep := logger.WithInitializationStep(ctx, step)

	// Retrieve step from context
	retrievedStep := logger.GetInitializationStep(ctxWithStep)
	assert.Equal(t, step, retrievedStep)

	// Test empty context
	emptyStep := logger.GetInitializationStep(ctx)
	assert.Empty(t, emptyStep)
}

func TestLoggerWithContext(t *testing.T) {
	capture := setupTestLogger()
	defer capture.hook.Reset()

	correlationID := "test-correlation-123"
	step := "test-step"
	ctx := logger.WithCorrelationID(context.Background(), correlationID)
	ctx = logger.WithInitializationStep(ctx, step)

	// Log with context
	logger.WithContext(ctx).Info("Test message")

	entries, err := capture.parseLogEntries()
	require.NoError(t, err)
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "info", entry.Level)
	assert.Equal(t, "Test message", entry.Message)
	assert.Equal(t, correlationID, entry.CorrelationID)
	assert.Equal(t, step, entry.Step)
}

func TestLoggerWithContextAndFields(t *testing.T) {
	capture := setupTestLogger()
	defer capture.hook.Reset()

	correlationID := "test-correlation-456"
	ctx := logger.WithCorrelationID(context.Background(), correlationID)

	fields := map[string]interface{}{
		"component": "test-component",
		"action":    "test-action",
		"duration":  "100ms",
	}

	// Log with context and fields
	logger.WithContextAndFields(ctx, fields).Info("Test message with fields")

	entries, err := capture.parseLogEntries()
	require.NoError(t, err)
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "info", entry.Level)
	assert.Equal(t, "Test message with fields", entry.Message)
	assert.Equal(t, correlationID, entry.CorrelationID)
	assert.Equal(t, "test-component", entry.Component)
	assert.Equal(t, "test-action", entry.Action)
	assert.Equal(t, "100ms", entry.Duration)
}

func TestInitializationServiceLogging(t *testing.T) {
	capture := setupTestLogger()
	defer capture.hook.Reset()

	// Create a test configuration
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

	// Create initialization service
	service, err := NewInitService(cfg)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Parse log entries
	entries, err := capture.parseLogEntries()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(entries), 2) // At least creation and completion logs

	// Check service creation log
	var creationEntry *LogEntry
	for _, entry := range entries {
		if entry.Action == "create_service" {
			creationEntry = &entry
			break
		}
	}
	require.NotNil(t, creationEntry)
	assert.Equal(t, "info", creationEntry.Level)
	assert.Equal(t, "init_service", creationEntry.Component)
	assert.NotEmpty(t, creationEntry.CorrelationID)

	// Check service created log
	var completedEntry *LogEntry
	for _, entry := range entries {
		if entry.Action == "service_created" {
			completedEntry = &entry
			break
		}
	}
	require.NotNil(t, completedEntry)
	assert.Equal(t, "info", completedEntry.Level)
	assert.Equal(t, "init_service", completedEntry.Component)
	assert.Equal(t, creationEntry.CorrelationID, completedEntry.CorrelationID)
}

func TestStepSummaryCreation(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-testing",
		},
	}

	service, err := NewInitService(cfg)
	require.NoError(t, err)

	startTime := time.Now()
	time.Sleep(10 * time.Millisecond) // Small delay to ensure duration > 0
	endTime := time.Now()

	summary := service.createStepSummary("test-step", startTime, endTime, "success", map[string]interface{}{
		"test_field": "test_value",
	})

	assert.Equal(t, "test-step", summary.Name)
	assert.Equal(t, startTime, summary.StartTime)
	assert.Equal(t, endTime, summary.EndTime)
	assert.Equal(t, "success", summary.Status)
	assert.Greater(t, summary.Duration, time.Duration(0))
	assert.Equal(t, map[string]interface{}{"test_field": "test_value"}, summary.Details)
}

func TestLogStepFailure(t *testing.T) {
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
	time.Sleep(10 * time.Millisecond)
	testError := assert.AnError

	service.logStepFailure("test-step", stepStart, testError)

	entries, err := capture.parseLogEntries()
	require.NoError(t, err)

	// Find the step failure log entry
	var failureEntry *LogEntry
	for _, entry := range entries {
		if entry.Action == "step_failed" {
			failureEntry = &entry
			break
		}
	}

	require.NotNil(t, failureEntry)
	assert.Equal(t, "error", failureEntry.Level)
	assert.Equal(t, "test-step", failureEntry.Step)
	assert.Equal(t, "failed", failureEntry.Status)
	assert.NotEmpty(t, failureEntry.Duration)
	assert.Contains(t, failureEntry.Error, testError.Error())
	assert.Contains(t, failureEntry.Message, "Step 'test-step' failed")
}

func TestInitializationSummaryStructure(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(5 * time.Minute)
	correlationID := "test-correlation-789"

	stepSummaries := []StepSummary{
		{
			Name:      "step1",
			StartTime: startTime,
			EndTime:   startTime.Add(1 * time.Minute),
			Duration:  1 * time.Minute,
			Status:    "success",
			Details:   nil,
		},
		{
			Name:      "step2",
			StartTime: startTime.Add(1 * time.Minute),
			EndTime:   endTime,
			Duration:  4 * time.Minute,
			Status:    "success",
			Details:   map[string]interface{}{"migrations": 5},
		},
	}

	summary := InitializationSummary{
		CorrelationID:     correlationID,
		StartTime:         startTime,
		EndTime:           endTime,
		TotalDuration:     endTime.Sub(startTime),
		StepsCompleted:    stepSummaries,
		AdminUserCreated:  true,
		AdminUsername:     "admin",
		MigrationsApplied: 5,
		DatabaseHost:      "localhost",
		DatabaseName:      "testdb",
	}

	assert.Equal(t, correlationID, summary.CorrelationID)
	assert.Equal(t, 5*time.Minute, summary.TotalDuration)
	assert.Len(t, summary.StepsCompleted, 2)
	assert.True(t, summary.AdminUserCreated)
	assert.Equal(t, "admin", summary.AdminUsername)
	assert.Equal(t, 5, summary.MigrationsApplied)
}

func TestLogFormatValidation(t *testing.T) {
	capture := setupTestLogger()
	defer capture.hook.Reset()

	correlationID := logger.NewCorrelationID()
	ctx := logger.WithCorrelationID(context.Background(), correlationID)
	ctx = logger.WithInitializationStep(ctx, "validation-test")

	// Log various types of messages
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "test",
		"action":    "start",
		"duration":  "100ms",
		"status":    "success",
	}).Info("Test info message")

	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "test",
		"action":    "error",
		"error":     "test error message",
	}).Error("Test error message")

	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "test",
		"action":    "debug",
	}).Debug("Test debug message")

	entries, err := capture.parseLogEntries()
	require.NoError(t, err)
	require.Len(t, entries, 3)

	// Validate each entry has required fields
	for _, entry := range entries {
		assert.NotEmpty(t, entry.Time)
		assert.NotEmpty(t, entry.Level)
		assert.NotEmpty(t, entry.Message)
		assert.Equal(t, correlationID, entry.CorrelationID)
		assert.Equal(t, "validation-test", entry.Step)
		assert.Equal(t, "test", entry.Component)

		// Validate JSON structure
		assert.NotNil(t, entry.Fields)
		assert.Contains(t, entry.Fields, "correlation_id")
		assert.Contains(t, entry.Fields, "step")
		assert.Contains(t, entry.Fields, "component")
		assert.Contains(t, entry.Fields, "action")
	}

	// Check specific log levels
	assert.Equal(t, "info", entries[0].Level)
	assert.Equal(t, "error", entries[1].Level)
	assert.Equal(t, "debug", entries[2].Level)

	// Check error entry has error field
	assert.NotEmpty(t, entries[1].Error)
}
