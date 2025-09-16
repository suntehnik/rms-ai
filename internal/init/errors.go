package init

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ErrorType represents the category of initialization error
type ErrorType string

const (
	ErrorTypeConfig    ErrorType = "configuration"
	ErrorTypeDatabase  ErrorType = "database"
	ErrorTypeSafety    ErrorType = "safety"
	ErrorTypeMigration ErrorType = "migration"
	ErrorTypeCreation  ErrorType = "creation"
	ErrorTypeSystem    ErrorType = "system"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

const (
	SeverityCritical ErrorSeverity = "critical"
	SeverityHigh     ErrorSeverity = "high"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityLow      ErrorSeverity = "low"
)

// InitError represents a comprehensive initialization error with context and metadata
type InitError struct {
	Type          ErrorType              `json:"type"`
	Severity      ErrorSeverity          `json:"severity"`
	Message       string                 `json:"message"`
	Cause         error                  `json:"cause,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Step          string                 `json:"step,omitempty"`
	Recoverable   bool                   `json:"recoverable"`
	ExitCode      int                    `json:"exit_code"`
}

// Error implements the error interface
func (e *InitError) Error() string {
	var parts []string

	// Add type and severity
	parts = append(parts, fmt.Sprintf("[%s:%s]", e.Type, e.Severity))

	// Add step if available
	if e.Step != "" {
		parts = append(parts, fmt.Sprintf("Step '%s':", e.Step))
	}

	// Add main message
	parts = append(parts, e.Message)

	// Add cause if available
	if e.Cause != nil {
		parts = append(parts, fmt.Sprintf("(caused by: %v)", e.Cause))
	}

	return strings.Join(parts, " ")
}

// JSON returns the error as a JSON string for structured logging
func (e *InitError) JSON() string {
	data, _ := json.Marshal(e)
	return string(data)
}

// IsRecoverable returns whether this error might be recoverable with retry or user action
func (e *InitError) IsRecoverable() bool {
	return e.Recoverable
}

// GetExitCode returns the appropriate exit code for this error
func (e *InitError) GetExitCode() int {
	return e.ExitCode
}

// WithContext adds additional context to the error
func (e *InitError) WithContext(key string, value interface{}) *InitError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithCorrelationID adds a correlation ID to the error
func (e *InitError) WithCorrelationID(correlationID string) *InitError {
	e.CorrelationID = correlationID
	return e
}

// WithStep adds the step information to the error
func (e *InitError) WithStep(step string) *InitError {
	e.Step = step
	return e
}

// Exit codes for different failure scenarios
const (
	ExitSuccess           = 0
	ExitConfigError       = 1
	ExitDatabaseError     = 2
	ExitSafetyError       = 3
	ExitMigrationError    = 4
	ExitUserCreationError = 5
	ExitSystemError       = 10
)

// NewConfigError creates a new configuration error
func NewConfigError(message string, cause error) *InitError {
	return &InitError{
		Type:        ErrorTypeConfig,
		Severity:    SeverityCritical,
		Message:     message,
		Cause:       cause,
		Timestamp:   time.Now(),
		Recoverable: true,
		ExitCode:    ExitConfigError,
		Context:     make(map[string]interface{}),
	}
}

// NewDatabaseError creates a new database error
func NewDatabaseError(message string, cause error) *InitError {
	return &InitError{
		Type:        ErrorTypeDatabase,
		Severity:    SeverityCritical,
		Message:     message,
		Cause:       cause,
		Timestamp:   time.Now(),
		Recoverable: true,
		ExitCode:    ExitDatabaseError,
		Context:     make(map[string]interface{}),
	}
}

// NewSafetyError creates a new safety check error
func NewSafetyError(message string, cause error) *InitError {
	return &InitError{
		Type:        ErrorTypeSafety,
		Severity:    SeverityCritical,
		Message:     message,
		Cause:       cause,
		Timestamp:   time.Now(),
		Recoverable: false, // Safety errors are not recoverable - database must be empty
		ExitCode:    ExitSafetyError,
		Context:     make(map[string]interface{}),
	}
}

// NewMigrationError creates a new migration error
func NewMigrationError(message string, cause error) *InitError {
	return &InitError{
		Type:        ErrorTypeMigration,
		Severity:    SeverityCritical,
		Message:     message,
		Cause:       cause,
		Timestamp:   time.Now(),
		Recoverable: true,
		ExitCode:    ExitMigrationError,
		Context:     make(map[string]interface{}),
	}
}

// NewCreationError creates a new user creation error
func NewCreationError(message string, cause error) *InitError {
	return &InitError{
		Type:        ErrorTypeCreation,
		Severity:    SeverityCritical,
		Message:     message,
		Cause:       cause,
		Timestamp:   time.Now(),
		Recoverable: true,
		ExitCode:    ExitUserCreationError,
		Context:     make(map[string]interface{}),
	}
}

// NewSystemError creates a new system error
func NewSystemError(message string, cause error) *InitError {
	return &InitError{
		Type:        ErrorTypeSystem,
		Severity:    SeverityCritical,
		Message:     message,
		Cause:       cause,
		Timestamp:   time.Now(),
		Recoverable: false,
		ExitCode:    ExitSystemError,
		Context:     make(map[string]interface{}),
	}
}

// ErrorContext provides structured context collection for errors
type ErrorContext struct {
	CorrelationID string                 `json:"correlation_id"`
	Step          string                 `json:"step"`
	StartTime     time.Time              `json:"start_time"`
	Duration      time.Duration          `json:"duration,omitempty"`
	Data          map[string]interface{} `json:"data"`
}

// NewErrorContext creates a new error context
func NewErrorContext(correlationID, step string) *ErrorContext {
	return &ErrorContext{
		CorrelationID: correlationID,
		Step:          step,
		StartTime:     time.Now(),
		Data:          make(map[string]interface{}),
	}
}

// AddData adds contextual data to the error context
func (ec *ErrorContext) AddData(key string, value interface{}) *ErrorContext {
	ec.Data[key] = value
	return ec
}

// Complete marks the context as complete and calculates duration
func (ec *ErrorContext) Complete() *ErrorContext {
	ec.Duration = time.Since(ec.StartTime)
	return ec
}

// ToMap converts the error context to a map for logging
func (ec *ErrorContext) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	result["correlation_id"] = ec.CorrelationID
	result["step"] = ec.Step
	result["start_time"] = ec.StartTime
	if ec.Duration > 0 {
		result["duration"] = ec.Duration.String()
	}

	// Add all data fields
	for k, v := range ec.Data {
		result[k] = v
	}

	return result
}

// WrapError wraps an existing error as an InitError with proper type detection
func WrapError(err error, context *ErrorContext) *InitError {
	if err == nil {
		return nil
	}

	// If it's already an InitError, enhance it with context
	if initErr, ok := err.(*InitError); ok {
		if context != nil {
			initErr.CorrelationID = context.CorrelationID
			initErr.Step = context.Step
			for k, v := range context.ToMap() {
				initErr.WithContext(k, v)
			}
		}
		return initErr
	}

	// Determine error type based on error message content
	errStr := strings.ToLower(err.Error())

	var initErr *InitError

	switch {
	case containsAny(errStr, "configuration", "environment", "missing", "invalid", "config"):
		initErr = NewConfigError("Configuration error", err)
	case containsAny(errStr, "database", "connection", "postgres", "sql", "db"):
		initErr = NewDatabaseError("Database error", err)
	case containsAny(errStr, "safety", "not empty", "existing data", "non-empty"):
		initErr = NewSafetyError("Safety check failed", err)
	case containsAny(errStr, "migration", "schema", "migrate"):
		initErr = NewMigrationError("Migration error", err)
	case containsAny(errStr, "user", "admin", "password", "creation", "hash"):
		initErr = NewCreationError("User creation error", err)
	default:
		initErr = NewSystemError("System error", err)
	}

	// Add context if provided
	if context != nil {
		initErr.CorrelationID = context.CorrelationID
		initErr.Step = context.Step
		for k, v := range context.ToMap() {
			initErr.WithContext(k, v)
		}
	}

	return initErr
}

// containsAny checks if the string contains any of the given substrings
func containsAny(str string, substrings ...string) bool {
	for _, substr := range substrings {
		if strings.Contains(str, substr) {
			return true
		}
	}
	return false
}

// DetermineExitCode determines the appropriate exit code based on error type
func DetermineExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}

	// If it's an InitError, use its exit code
	if initErr, ok := err.(*InitError); ok {
		return initErr.GetExitCode()
	}

	// Fallback to string-based detection for non-InitError types
	errStr := strings.ToLower(err.Error())

	switch {
	case containsAny(errStr, "configuration", "environment", "missing", "invalid"):
		return ExitConfigError
	case containsAny(errStr, "database", "connection", "postgres"):
		return ExitDatabaseError
	case containsAny(errStr, "safety", "not empty", "existing data"):
		return ExitSafetyError
	case containsAny(errStr, "migration", "schema"):
		return ExitMigrationError
	case containsAny(errStr, "user", "admin", "password"):
		return ExitUserCreationError
	default:
		return ExitSystemError
	}
}

// ErrorReporter provides structured error reporting capabilities
type ErrorReporter struct {
	correlationID string
	errors        []*InitError
}

// NewErrorReporter creates a new error reporter
func NewErrorReporter(correlationID string) *ErrorReporter {
	return &ErrorReporter{
		correlationID: correlationID,
		errors:        make([]*InitError, 0),
	}
}

// ReportError adds an error to the reporter
func (er *ErrorReporter) ReportError(err *InitError) {
	if err != nil {
		err.CorrelationID = er.correlationID
		er.errors = append(er.errors, err)
	}
}

// GetErrors returns all reported errors
func (er *ErrorReporter) GetErrors() []*InitError {
	return er.errors
}

// HasErrors returns true if any errors have been reported
func (er *ErrorReporter) HasErrors() bool {
	return len(er.errors) > 0
}

// GetMostSevereError returns the error with the highest severity
func (er *ErrorReporter) GetMostSevereError() *InitError {
	if len(er.errors) == 0 {
		return nil
	}

	severityOrder := map[ErrorSeverity]int{
		SeverityCritical: 4,
		SeverityHigh:     3,
		SeverityMedium:   2,
		SeverityLow:      1,
	}

	mostSevere := er.errors[0]
	for _, err := range er.errors[1:] {
		if severityOrder[err.Severity] > severityOrder[mostSevere.Severity] {
			mostSevere = err
		}
	}

	return mostSevere
}

// GenerateReport generates a comprehensive error report
func (er *ErrorReporter) GenerateReport() map[string]interface{} {
	report := map[string]interface{}{
		"correlation_id": er.correlationID,
		"error_count":    len(er.errors),
		"timestamp":      time.Now(),
	}

	if len(er.errors) > 0 {
		report["most_severe"] = er.GetMostSevereError()
		report["all_errors"] = er.errors

		// Group errors by type
		errorsByType := make(map[ErrorType][]*InitError)
		for _, err := range er.errors {
			errorsByType[err.Type] = append(errorsByType[err.Type], err)
		}
		report["errors_by_type"] = errorsByType
	}

	return report
}
