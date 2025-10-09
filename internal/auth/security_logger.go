package auth

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"product-requirements-management/internal/logger"
)

// ClientInfo contains client information for security logging
type ClientInfo struct {
	IP        string
	UserAgent string
}

// ClientInfoKey is the context key for client information
type ClientInfoKey struct{}

// WithClientInfo adds client information to context
func WithClientInfo(ctx context.Context, clientIP, userAgent string) context.Context {
	return context.WithValue(ctx, ClientInfoKey{}, ClientInfo{
		IP:        clientIP,
		UserAgent: userAgent,
	})
}

// GetClientInfo retrieves client information from context
func GetClientInfo(ctx context.Context) (ClientInfo, bool) {
	info, ok := ctx.Value(ClientInfoKey{}).(ClientInfo)
	return info, ok
}

// SecurityEvent represents different types of security events
type SecurityEvent string

const (
	// PAT Events
	SecurityEventPATCreated        SecurityEvent = "pat_created"
	SecurityEventPATRevoked        SecurityEvent = "pat_revoked"
	SecurityEventPATAuthSuccess    SecurityEvent = "pat_auth_success"
	SecurityEventPATAuthFailure    SecurityEvent = "pat_auth_failure"
	SecurityEventPATExpired        SecurityEvent = "pat_expired"
	SecurityEventPATInvalidFormat  SecurityEvent = "pat_invalid_format"
	SecurityEventPATInvalidPrefix  SecurityEvent = "pat_invalid_prefix"
	SecurityEventPATTokenMismatch  SecurityEvent = "pat_token_mismatch"
	SecurityEventPATUserNotFound   SecurityEvent = "pat_user_not_found"
	SecurityEventPATCleanupExpired SecurityEvent = "pat_cleanup_expired"

	// Authentication Events
	SecurityEventAuthAttempt      SecurityEvent = "auth_attempt"
	SecurityEventAuthSuccess      SecurityEvent = "auth_success"
	SecurityEventAuthFailure      SecurityEvent = "auth_failure"
	SecurityEventAuthMethodSwitch SecurityEvent = "auth_method_switch"
)

// SecurityLogger handles security event logging without exposing sensitive information
type SecurityLogger struct {
	logger *logrus.Logger
}

// NewSecurityLogger creates a new security logger instance
func NewSecurityLogger() *SecurityLogger {
	return &SecurityLogger{
		logger: logger.Logger,
	}
}

// SecurityEventData contains data for security events
type SecurityEventData struct {
	Event       SecurityEvent `json:"event"`
	UserID      *uuid.UUID    `json:"user_id,omitempty"`
	Username    string        `json:"username,omitempty"`
	PATID       *uuid.UUID    `json:"pat_id,omitempty"`
	PATName     string        `json:"pat_name,omitempty"`
	TokenPrefix string        `json:"token_prefix,omitempty"`
	AuthMethod  string        `json:"auth_method,omitempty"`
	ClientIP    string        `json:"client_ip,omitempty"`
	UserAgent   string        `json:"user_agent,omitempty"`
	Reason      string        `json:"reason,omitempty"`
	Count       int           `json:"count,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
}

// LogPATCreated logs PAT creation events
func (sl *SecurityLogger) LogPATCreated(ctx context.Context, userID uuid.UUID, username, patID, patName string, clientIP, userAgent string) {
	patUUID, _ := uuid.Parse(patID)

	data := SecurityEventData{
		Event:     SecurityEventPATCreated,
		UserID:    &userID,
		Username:  username,
		PATID:     &patUUID,
		PATName:   patName,
		ClientIP:  clientIP,
		UserAgent: userAgent,
		Timestamp: time.Now(),
	}

	sl.logSecurityEvent(ctx, data, "Personal Access Token created")
}

// LogPATRevoked logs PAT revocation events
func (sl *SecurityLogger) LogPATRevoked(ctx context.Context, userID uuid.UUID, username, patID, patName string, clientIP, userAgent string) {
	patUUID, _ := uuid.Parse(patID)

	data := SecurityEventData{
		Event:     SecurityEventPATRevoked,
		UserID:    &userID,
		Username:  username,
		PATID:     &patUUID,
		PATName:   patName,
		ClientIP:  clientIP,
		UserAgent: userAgent,
		Timestamp: time.Now(),
	}

	sl.logSecurityEvent(ctx, data, "Personal Access Token revoked")
}

// LogPATAuthSuccess logs successful PAT authentication
func (sl *SecurityLogger) LogPATAuthSuccess(ctx context.Context, userID uuid.UUID, username, patID string, clientIP, userAgent string) {
	patUUID, _ := uuid.Parse(patID)

	data := SecurityEventData{
		Event:       SecurityEventPATAuthSuccess,
		UserID:      &userID,
		Username:    username,
		PATID:       &patUUID,
		TokenPrefix: "mcp_pat_",
		AuthMethod:  "pat",
		ClientIP:    clientIP,
		UserAgent:   userAgent,
		Timestamp:   time.Now(),
	}

	sl.logSecurityEvent(ctx, data, "PAT authentication successful")
}

// LogPATAuthFailure logs failed PAT authentication attempts
func (sl *SecurityLogger) LogPATAuthFailure(ctx context.Context, reason, tokenPrefix, clientIP, userAgent string) {
	data := SecurityEventData{
		Event:       SecurityEventPATAuthFailure,
		TokenPrefix: sl.sanitizeTokenPrefix(tokenPrefix),
		AuthMethod:  "pat",
		ClientIP:    clientIP,
		UserAgent:   userAgent,
		Reason:      reason,
		Timestamp:   time.Now(),
	}

	sl.logSecurityEvent(ctx, data, "PAT authentication failed: "+reason)
}

// LogPATExpired logs when an expired PAT is used
func (sl *SecurityLogger) LogPATExpired(ctx context.Context, userID uuid.UUID, username, patID string, clientIP, userAgent string) {
	patUUID, _ := uuid.Parse(patID)

	data := SecurityEventData{
		Event:       SecurityEventPATExpired,
		UserID:      &userID,
		Username:    username,
		PATID:       &patUUID,
		TokenPrefix: "mcp_pat_",
		AuthMethod:  "pat",
		ClientIP:    clientIP,
		UserAgent:   userAgent,
		Reason:      "token_expired",
		Timestamp:   time.Now(),
	}

	sl.logSecurityEvent(ctx, data, "Expired PAT authentication attempt")
}

// LogPATCleanupExpired logs cleanup of expired tokens
func (sl *SecurityLogger) LogPATCleanupExpired(ctx context.Context, count int) {
	data := SecurityEventData{
		Event:     SecurityEventPATCleanupExpired,
		Count:     count,
		Timestamp: time.Now(),
	}

	sl.logSecurityEvent(ctx, data, "Expired PATs cleaned up")
}

// LogAuthAttempt logs general authentication attempts
func (sl *SecurityLogger) LogAuthAttempt(ctx context.Context, authMethod, clientIP, userAgent string) {
	data := SecurityEventData{
		Event:      SecurityEventAuthAttempt,
		AuthMethod: authMethod,
		ClientIP:   clientIP,
		UserAgent:  userAgent,
		Timestamp:  time.Now(),
	}

	sl.logSecurityEvent(ctx, data, "Authentication attempt")
}

// LogAuthSuccess logs successful authentication
func (sl *SecurityLogger) LogAuthSuccess(ctx context.Context, userID uuid.UUID, username, authMethod, clientIP, userAgent string) {
	data := SecurityEventData{
		Event:      SecurityEventAuthSuccess,
		UserID:     &userID,
		Username:   username,
		AuthMethod: authMethod,
		ClientIP:   clientIP,
		UserAgent:  userAgent,
		Timestamp:  time.Now(),
	}

	sl.logSecurityEvent(ctx, data, "Authentication successful")
}

// LogAuthFailure logs failed authentication attempts
func (sl *SecurityLogger) LogAuthFailure(ctx context.Context, reason, authMethod, clientIP, userAgent string) {
	data := SecurityEventData{
		Event:      SecurityEventAuthFailure,
		AuthMethod: authMethod,
		ClientIP:   clientIP,
		UserAgent:  userAgent,
		Reason:     reason,
		Timestamp:  time.Now(),
	}

	sl.logSecurityEvent(ctx, data, "Authentication failed: "+reason)
}

// LogAuthMethodSwitch logs when authentication method changes within a session
func (sl *SecurityLogger) LogAuthMethodSwitch(ctx context.Context, userID uuid.UUID, username, fromMethod, toMethod, clientIP, userAgent string) {
	data := SecurityEventData{
		Event:      SecurityEventAuthMethodSwitch,
		UserID:     &userID,
		Username:   username,
		AuthMethod: toMethod,
		ClientIP:   clientIP,
		UserAgent:  userAgent,
		Reason:     "switched_from_" + fromMethod,
		Timestamp:  time.Now(),
	}

	sl.logSecurityEvent(ctx, data, "Authentication method switched from "+fromMethod+" to "+toMethod)
}

// logSecurityEvent logs a security event with structured logging
func (sl *SecurityLogger) logSecurityEvent(ctx context.Context, data SecurityEventData, message string) {
	entry := logger.WithContext(ctx).WithFields(logrus.Fields{
		"security_event": data.Event,
		"event_data":     data,
		"component":      "security",
		"category":       "authentication",
	})

	// Add correlation ID if available
	if correlationID := logger.GetCorrelationID(ctx); correlationID != "" {
		entry = entry.WithField("correlation_id", correlationID)
	}

	// Log at appropriate level based on event type
	switch data.Event {
	case SecurityEventPATAuthFailure, SecurityEventAuthFailure, SecurityEventPATExpired:
		entry.Warn(message)
	case SecurityEventPATCreated, SecurityEventPATRevoked, SecurityEventPATCleanupExpired:
		entry.Info(message)
	case SecurityEventPATAuthSuccess, SecurityEventAuthSuccess:
		entry.Info(message)
	default:
		entry.Info(message)
	}
}

// sanitizeTokenPrefix ensures we only log safe token prefix information
func (sl *SecurityLogger) sanitizeTokenPrefix(tokenPrefix string) string {
	// Only log known safe prefixes
	if strings.HasPrefix(tokenPrefix, "mcp_pat_") {
		return "mcp_pat_"
	}
	if strings.HasPrefix(tokenPrefix, "jwt_") {
		return "jwt_"
	}
	// For unknown prefixes, just indicate it's a token
	return "unknown_token_type"
}

// GetSecurityEventFields returns structured fields for security events
func GetSecurityEventFields(event SecurityEvent, userID *uuid.UUID, clientIP, userAgent string) logrus.Fields {
	fields := logrus.Fields{
		"security_event": event,
		"client_ip":      clientIP,
		"user_agent":     userAgent,
		"timestamp":      time.Now(),
	}

	if userID != nil {
		fields["user_id"] = userID.String()
	}

	return fields
}
