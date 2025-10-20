package handlers

import (
	"context"
	"regexp"
	"strings"
	"time"

	"product-requirements-management/internal/logger"
	"product-requirements-management/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// MCPLogger provides structured logging for MCP operations with security features
type MCPLogger struct {
	logger *logrus.Logger
}

// NewMCPLogger creates a new MCP logger instance
func NewMCPLogger() *MCPLogger {
	// Ensure logger is initialized
	if logger.Logger == nil {
		// Initialize with a basic logger for testing
		logger.Logger = logrus.New()
	}

	return &MCPLogger{
		logger: logger.Logger,
	}
}

// LogRequest logs an incoming MCP request with correlation ID and user context
func (ml *MCPLogger) LogRequest(ctx context.Context, method string, params interface{}, user *models.User) {
	correlationID := logger.GetCorrelationID(ctx)
	if correlationID == "" {
		correlationID = logger.NewCorrelationID()
	}

	fields := logrus.Fields{
		"correlation_id": correlationID,
		"component":      "mcp_handler",
		"operation":      "request",
		"method":         method,
		"timestamp":      time.Now().UTC(),
	}

	// Add user information if available
	if user != nil {
		fields["user_id"] = user.ID.String()
		fields["username"] = user.Username
		fields["user_role"] = user.Role
	}

	// Add sanitized parameters (remove sensitive data)
	if params != nil {
		fields["params"] = ml.sanitizeParams(params)
	}

	ml.logger.WithFields(fields).Info("Processing MCP request")
}

// LogRequestBody logs the raw JSON-RPC request body for debugging
func (ml *MCPLogger) LogRequestBody(ctx context.Context, method string, requestBody []byte, user *models.User) {
	correlationID := logger.GetCorrelationID(ctx)
	if correlationID == "" {
		correlationID = logger.NewCorrelationID()
	}

	fields := logrus.Fields{
		"correlation_id": correlationID,
		"component":      "mcp_handler",
		"operation":      "request_body",
		"method":         method,
		"timestamp":      time.Now().UTC(),
	}

	// Add user information if available
	if user != nil {
		fields["user_id"] = user.ID.String()
		fields["username"] = user.Username
	}

	// Sanitize the request body before logging
	sanitizedBody := ml.redactSensitiveStrings(string(requestBody))
	fields["request_body"] = sanitizedBody

	ml.logger.WithFields(fields).Info("MCP JSON-RPC request body")
}

// LogResponse logs an MCP response with timing information
func (ml *MCPLogger) LogResponse(ctx context.Context, method string, success bool, duration time.Duration, user *models.User) {
	correlationID := logger.GetCorrelationID(ctx)

	fields := logrus.Fields{
		"correlation_id": correlationID,
		"component":      "mcp_handler",
		"operation":      "response",
		"method":         method,
		"success":        success,
		"duration_ms":    duration.Milliseconds(),
		"timestamp":      time.Now().UTC(),
	}

	// Add user information if available
	if user != nil {
		fields["user_id"] = user.ID.String()
		fields["username"] = user.Username
	}

	if success {
		ml.logger.WithFields(fields).Info("MCP request completed successfully")
	} else {
		ml.logger.WithFields(fields).Warn("MCP request completed with error")
	}
}

// LogResponseBody logs the raw JSON-RPC response body for debugging
func (ml *MCPLogger) LogResponseBody(ctx context.Context, method string, responseBody []byte, user *models.User) {
	correlationID := logger.GetCorrelationID(ctx)

	fields := logrus.Fields{
		"correlation_id": correlationID,
		"component":      "mcp_handler",
		"operation":      "response_body",
		"method":         method,
		"timestamp":      time.Now().UTC(),
	}

	// Add user information if available
	if user != nil {
		fields["user_id"] = user.ID.String()
		fields["username"] = user.Username
	}

	// Sanitize the response body before logging (though responses typically don't contain sensitive data)
	sanitizedBody := ml.redactSensitiveStrings(string(responseBody))
	fields["response_body"] = sanitizedBody

	ml.logger.WithFields(fields).Info("MCP JSON-RPC response body")
}

// LogError logs an error that occurred during MCP processing
func (ml *MCPLogger) LogError(ctx context.Context, method string, err error, user *models.User) {
	correlationID := logger.GetCorrelationID(ctx)

	fields := logrus.Fields{
		"correlation_id": correlationID,
		"component":      "mcp_handler",
		"operation":      "error",
		"method":         method,
		"error":          err.Error(),
		"timestamp":      time.Now().UTC(),
	}

	// Add user information if available
	if user != nil {
		fields["user_id"] = user.ID.String()
		fields["username"] = user.Username
	}

	ml.logger.WithFields(fields).Error("MCP request failed")
}

// LogAuditEvent logs an audit event for MCP operations
func (ml *MCPLogger) LogAuditEvent(ctx context.Context, action string, resourceType string, resourceID string, user *models.User, details map[string]interface{}) {
	correlationID := logger.GetCorrelationID(ctx)

	fields := logrus.Fields{
		"correlation_id": correlationID,
		"component":      "mcp_audit",
		"action":         action,
		"resource_type":  resourceType,
		"resource_id":    resourceID,
		"timestamp":      time.Now().UTC(),
	}

	// Add user information (required for audit logs)
	if user != nil {
		fields["user_id"] = user.ID.String()
		fields["username"] = user.Username
		fields["user_role"] = user.Role
	}

	// Add additional details if provided
	for key, value := range details {
		fields[key] = ml.sanitizeValue(value)
	}

	ml.logger.WithFields(fields).Info("MCP audit event")
}

// LogSecurityEvent logs security-related events (authentication failures, etc.)
func (ml *MCPLogger) LogSecurityEvent(ctx context.Context, event string, details map[string]interface{}) {
	correlationID := logger.GetCorrelationID(ctx)

	fields := logrus.Fields{
		"correlation_id": correlationID,
		"component":      "mcp_security",
		"event":          event,
		"timestamp":      time.Now().UTC(),
	}

	// Add sanitized details
	for key, value := range details {
		fields[key] = ml.sanitizeValue(value)
	}

	ml.logger.WithFields(fields).Warn("MCP security event")
}

// LogPerformanceMetrics logs performance metrics for MCP operations
func (ml *MCPLogger) LogPerformanceMetrics(ctx context.Context, method string, duration time.Duration, user *models.User, metrics map[string]interface{}) {
	correlationID := logger.GetCorrelationID(ctx)

	fields := logrus.Fields{
		"correlation_id": correlationID,
		"component":      "mcp_performance",
		"method":         method,
		"duration_ms":    duration.Milliseconds(),
		"timestamp":      time.Now().UTC(),
	}

	// Add user information if available
	if user != nil {
		fields["user_id"] = user.ID.String()
	}

	// Add performance metrics
	for key, value := range metrics {
		fields[key] = value
	}

	ml.logger.WithFields(fields).Info("MCP performance metrics")
}

// sanitizeParams removes sensitive information from parameters before logging
func (ml *MCPLogger) sanitizeParams(params interface{}) interface{} {
	return ml.sanitizeValue(params)
}

// sanitizeValue recursively sanitizes values to remove sensitive information
func (ml *MCPLogger) sanitizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return ml.redactSensitiveStrings(v)
	case map[string]interface{}:
		sanitized := make(map[string]interface{})
		for key, val := range v {
			if ml.isSensitiveKey(key) {
				sanitized[key] = "[REDACTED]"
			} else {
				sanitized[key] = ml.sanitizeValue(val)
			}
		}
		return sanitized
	case []interface{}:
		sanitized := make([]interface{}, len(v))
		for i, val := range v {
			sanitized[i] = ml.sanitizeValue(val)
		}
		return sanitized
	default:
		return value
	}
}

// redactSensitiveStrings redacts PAT tokens and other sensitive strings
func (ml *MCPLogger) redactSensitiveStrings(s string) string {
	// Handle Bearer PAT tokens specifically first
	bearerPATRegex := regexp.MustCompile(`Bearer\s+mcp_pat_[a-zA-Z0-9_-]+`)
	s = bearerPATRegex.ReplaceAllString(s, "Bearer mcp_pat_[REDACTED]")

	// Handle other Bearer tokens (JWT, etc.) - exclude already redacted ones
	bearerOtherRegex := regexp.MustCompile(`Bearer\s+[a-zA-Z0-9._-]+`)
	s = bearerOtherRegex.ReplaceAllStringFunc(s, func(match string) string {
		// Skip if it's a PAT token (redacted or not)
		if strings.Contains(match, "mcp_pat_") {
			return match
		}
		return "Bearer [REDACTED]"
	})

	// Handle standalone PAT tokens (not after Bearer)
	standalonePatRegex := regexp.MustCompile(`\bmcp_pat_[a-zA-Z0-9_-]+`)
	s = standalonePatRegex.ReplaceAllString(s, "mcp_pat_[REDACTED]")

	// Redact other potential tokens
	tokenRegex := regexp.MustCompile(`(?i)(token|key|secret|password)\s*[:=]\s*[^\s,}]+`)
	s = tokenRegex.ReplaceAllString(s, "$1: [REDACTED]")

	return s
}

// isSensitiveKey checks if a key contains sensitive information
func (ml *MCPLogger) isSensitiveKey(key string) bool {
	sensitiveKeys := []string{
		"token", "password", "secret", "key", "authorization",
		"auth", "credential", "pat", "jwt", "bearer",
	}

	lowerKey := strings.ToLower(key)
	for _, sensitive := range sensitiveKeys {
		if strings.Contains(lowerKey, sensitive) {
			return true
		}
	}
	return false
}

// GetUserFromGinContext extracts user information from Gin context
func (ml *MCPLogger) GetUserFromGinContext(c *gin.Context) *models.User {
	if c == nil {
		return nil
	}

	// Try to get user from context (set by PAT middleware)
	if userVal, exists := c.Get("user"); exists {
		if user, ok := userVal.(*models.User); ok {
			return user
		}
	}

	// Try to get user from claims (JWT authentication)
	if claimsVal, exists := c.Get("claims"); exists {
		if claims, ok := claimsVal.(map[string]interface{}); ok {
			user := &models.User{}
			if userID, ok := claims["user_id"].(string); ok {
				// Parse UUID if needed
				user.Username = userID // Fallback
			}
			if username, ok := claims["username"].(string); ok {
				user.Username = username
			}
			if role, ok := claims["role"].(string); ok {
				user.Role = models.UserRole(role)
			}
			return user
		}
	}

	return nil
}

// WithCorrelationID adds a correlation ID to the context if not present
func (ml *MCPLogger) WithCorrelationID(ctx context.Context) context.Context {
	if logger.GetCorrelationID(ctx) == "" {
		return logger.WithCorrelationID(ctx, logger.NewCorrelationID())
	}
	return ctx
}

// Info implements jsonrpc.Logger interface for info level logging
func (ml *MCPLogger) Info(msg string, fields ...interface{}) {
	logFields := logrus.Fields{}

	// Convert fields to logrus.Fields format
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key := fields[i].(string)
			value := fields[i+1]
			logFields[key] = value
		}
	}

	ml.logger.WithFields(logFields).Info(msg)
}

// Error implements jsonrpc.Logger interface for error level logging
func (ml *MCPLogger) Error(msg string, fields ...interface{}) {
	logFields := logrus.Fields{}

	// Convert fields to logrus.Fields format
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key := fields[i].(string)
			value := fields[i+1]
			logFields[key] = value
		}
	}

	ml.logger.WithFields(logFields).Error(msg)
}

// Debug implements jsonrpc.Logger interface for debug level logging
func (ml *MCPLogger) Debug(msg string, fields ...interface{}) {
	logFields := logrus.Fields{}

	// Convert fields to logrus.Fields format
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key := fields[i].(string)
			value := fields[i+1]
			logFields[key] = value
		}
	}

	ml.logger.WithFields(logFields).Debug(msg)
}
