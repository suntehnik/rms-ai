package handlers

import (
	"context"
	"testing"
	"time"

	"product-requirements-management/internal/logger"
	"product-requirements-management/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMCPLogger_RedactSensitiveStrings(t *testing.T) {
	mcpLogger := NewMCPLogger()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "redact PAT token",
			input:    "Authorization: Bearer mcp_pat_abc123def456",
			expected: "Authorization: Bearer mcp_pat_[REDACTED]",
		},
		{
			name:     "redact JWT token",
			input:    "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expected: "Bearer [REDACTED]",
		},
		{
			name:     "redact token in key-value pair",
			input:    "token: secret123",
			expected: "token: [REDACTED]",
		},
		{
			name:     "no redaction needed",
			input:    "This is a normal string",
			expected: "This is a normal string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mcpLogger.redactSensitiveStrings(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMCPLogger_IsSensitiveKey(t *testing.T) {
	mcpLogger := NewMCPLogger()

	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{
			name:     "token key",
			key:      "token",
			expected: true,
		},
		{
			name:     "password key",
			key:      "password",
			expected: true,
		},
		{
			name:     "authorization key",
			key:      "Authorization",
			expected: true,
		},
		{
			name:     "normal key",
			key:      "username",
			expected: false,
		},
		{
			name:     "pat key",
			key:      "pat_token",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mcpLogger.isSensitiveKey(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMCPLogger_SanitizeValue(t *testing.T) {
	mcpLogger := NewMCPLogger()

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "sanitize string with PAT",
			input:    "mcp_pat_secret123",
			expected: "mcp_pat_[REDACTED]",
		},
		{
			name: "sanitize map with sensitive keys",
			input: map[string]interface{}{
				"username": "john",
				"token":    "secret123",
				"data":     "normal",
			},
			expected: map[string]interface{}{
				"username": "john",
				"token":    "[REDACTED]",
				"data":     "normal",
			},
		},
		{
			name:     "sanitize array",
			input:    []interface{}{"normal", "mcp_pat_secret"},
			expected: []interface{}{"normal", "mcp_pat_[REDACTED]"},
		},
		{
			name:     "normal value",
			input:    42,
			expected: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mcpLogger.sanitizeValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMCPLogger_WithCorrelationID(t *testing.T) {
	mcpLogger := NewMCPLogger()

	t.Run("adds correlation ID when not present", func(t *testing.T) {
		ctx := context.Background()
		newCtx := mcpLogger.WithCorrelationID(ctx)

		correlationID := logger.GetCorrelationID(newCtx)
		assert.NotEmpty(t, correlationID)
	})

	t.Run("preserves existing correlation ID", func(t *testing.T) {
		existingID := "existing-id-123"
		ctx := logger.WithCorrelationID(context.Background(), existingID)
		newCtx := mcpLogger.WithCorrelationID(ctx)

		correlationID := logger.GetCorrelationID(newCtx)
		assert.Equal(t, existingID, correlationID)
	})
}

func TestMCPLogger_LogAuditEvent(t *testing.T) {
	mcpLogger := NewMCPLogger()

	// Create a test user
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     models.RoleUser,
	}

	// Create context with correlation ID
	ctx := logger.WithCorrelationID(context.Background(), "test-correlation-id")

	// Test that LogAuditEvent doesn't panic and handles all parameters
	assert.NotPanics(t, func() {
		mcpLogger.LogAuditEvent(ctx, "create", "epic", "EP-001", user, map[string]interface{}{
			"title":    "Test Epic",
			"priority": 1,
		})
	})
}

func TestMCPLogger_LogRequest(t *testing.T) {
	mcpLogger := NewMCPLogger()

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     models.RoleUser,
	}

	ctx := logger.WithCorrelationID(context.Background(), "test-correlation-id")

	params := map[string]interface{}{
		"name": "create_epic",
		"arguments": map[string]interface{}{
			"title": "Test Epic",
			"token": "mcp_pat_secret123", // Should be redacted
		},
	}

	// Test that LogRequest doesn't panic and handles sensitive data
	assert.NotPanics(t, func() {
		mcpLogger.LogRequest(ctx, "tools/call", params, user)
	})
}

func TestMCPLogger_LogResponse(t *testing.T) {
	mcpLogger := NewMCPLogger()

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     models.RoleUser,
	}

	ctx := logger.WithCorrelationID(context.Background(), "test-correlation-id")
	duration := 150 * time.Millisecond

	// Test successful response
	assert.NotPanics(t, func() {
		mcpLogger.LogResponse(ctx, "tools/call", true, duration, user)
	})

	// Test failed response
	assert.NotPanics(t, func() {
		mcpLogger.LogResponse(ctx, "tools/call", false, duration, user)
	})
}

func TestMCPLogger_LogError(t *testing.T) {
	mcpLogger := NewMCPLogger()

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     models.RoleUser,
	}

	ctx := logger.WithCorrelationID(context.Background(), "test-correlation-id")
	testError := assert.AnError

	// Test that LogError doesn't panic
	assert.NotPanics(t, func() {
		mcpLogger.LogError(ctx, "tools/call", testError, user)
	})
}

func TestMCPLogger_LogSecurityEvent(t *testing.T) {
	mcpLogger := NewMCPLogger()

	ctx := logger.WithCorrelationID(context.Background(), "test-correlation-id")

	details := map[string]interface{}{
		"client_ip":  "192.168.1.1",
		"user_agent": "MCP-Client/1.0",
		"token":      "mcp_pat_secret123", // Should be redacted
	}

	// Test that LogSecurityEvent doesn't panic and handles sensitive data
	assert.NotPanics(t, func() {
		mcpLogger.LogSecurityEvent(ctx, "authentication_failure", details)
	})
}

func TestMCPLogger_LogPerformanceMetrics(t *testing.T) {
	mcpLogger := NewMCPLogger()

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     models.RoleUser,
	}

	ctx := logger.WithCorrelationID(context.Background(), "test-correlation-id")
	duration := 250 * time.Millisecond

	metrics := map[string]interface{}{
		"request_size_bytes":  1024,
		"response_size_bytes": 2048,
		"slow_operation":      true,
	}

	// Test that LogPerformanceMetrics doesn't panic
	assert.NotPanics(t, func() {
		mcpLogger.LogPerformanceMetrics(ctx, "tools/call", duration, user, metrics)
	})
}
