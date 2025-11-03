package validation

import (
	"fmt"
	"strings"
)

// StatusValidationError represents a structured error for status validation failures
type StatusValidationError struct {
	EntityType    string   `json:"entity_type"`
	ProvidedValue string   `json:"provided_value"`
	ValidStatuses []string `json:"valid_statuses"`
	Message       string   `json:"message"`
}

// Error implements the error interface
func (e *StatusValidationError) Error() string {
	return e.Message
}

func (e *StatusValidationError) Unwrap() string {
	return e.Message
}

// NewStatusValidationError creates a new StatusValidationError with consistent formatting
func NewStatusValidationError(entityType, providedValue string, validStatuses []string) *StatusValidationError {
	var message string

	if providedValue == "" {
		message = fmt.Sprintf("Status is required for %s. Valid statuses are: %s",
			entityType,
			strings.Join(validStatuses, ", "))
	} else {
		message = fmt.Sprintf("Invalid status '%s' for %s. Valid statuses are: %s",
			providedValue,
			entityType,
			strings.Join(validStatuses, ", "))
	}

	return &StatusValidationError{
		EntityType:    entityType,
		ProvidedValue: providedValue,
		ValidStatuses: validStatuses,
		Message:       message,
	}
}

// IsStatusValidationError checks if an error is a StatusValidationError
func IsStatusValidationError(err error) bool {
	_, ok := err.(*StatusValidationError)
	return ok
}

// GetStatusValidationError extracts StatusValidationError from an error
func GetStatusValidationError(err error) (*StatusValidationError, bool) {
	statusErr, ok := err.(*StatusValidationError)
	return statusErr, ok
}

// ValidationErrorType represents different types of validation errors
type ValidationErrorType string

const (
	// StatusValidationType represents status validation errors
	StatusValidationType ValidationErrorType = "status_validation"
	// EntityNotFoundType represents entity not found errors
	EntityNotFoundType ValidationErrorType = "entity_not_found"
	// GeneralValidationType represents general validation errors
	GeneralValidationType ValidationErrorType = "general_validation"
)

// StructuredValidationError provides a consistent error structure across the system
type StructuredValidationError struct {
	Type       ValidationErrorType `json:"type"`
	Code       string              `json:"code"`
	Message    string              `json:"message"`
	Details    interface{}         `json:"details,omitempty"`
	EntityType string              `json:"entity_type,omitempty"`
	EntityID   string              `json:"entity_id,omitempty"`
	Field      string              `json:"field,omitempty"`
}

// Error implements the error interface
func (e *StructuredValidationError) Error() string {
	return e.Message
}

// NewStructuredValidationError creates a new StructuredValidationError
func NewStructuredValidationError(errorType ValidationErrorType, code, message string) *StructuredValidationError {
	return &StructuredValidationError{
		Type:    errorType,
		Code:    code,
		Message: message,
	}
}

// WithDetails adds details to the structured error
func (e *StructuredValidationError) WithDetails(details interface{}) *StructuredValidationError {
	e.Details = details
	return e
}

// WithEntity adds entity information to the structured error
func (e *StructuredValidationError) WithEntity(entityType, entityID string) *StructuredValidationError {
	e.EntityType = entityType
	e.EntityID = entityID
	return e
}

// WithField adds field information to the structured error
func (e *StructuredValidationError) WithField(field string) *StructuredValidationError {
	e.Field = field
	return e
}

// Common error codes for status validation
const (
	// Status validation error codes
	InvalidStatusCode    = "INVALID_STATUS"
	MissingStatusCode    = "MISSING_STATUS"
	StatusTransitionCode = "INVALID_STATUS_TRANSITION"

	// Entity error codes
	EntityNotFoundCode = "ENTITY_NOT_FOUND"

	// General validation error codes
	ValidationFailedCode = "VALIDATION_FAILED"
	RequiredFieldCode    = "REQUIRED_FIELD"
)

// Helper functions for creating common validation errors

// NewInvalidStatusError creates a structured error for invalid status values
func NewInvalidStatusError(entityType, providedValue string, validStatuses []string) *StructuredValidationError {
	statusErr := NewStatusValidationError(entityType, providedValue, validStatuses)

	return NewStructuredValidationError(
		StatusValidationType,
		InvalidStatusCode,
		statusErr.Message,
	).WithDetails(map[string]interface{}{
		"provided_value": providedValue,
		"valid_statuses": validStatuses,
	}).WithField("status")
}

// NewEntityNotFoundError creates a structured error for entity not found cases
func NewEntityNotFoundError(entityType, entityID string) *StructuredValidationError {
	message := fmt.Sprintf("%s with ID '%s' not found",
		strings.Title(strings.ReplaceAll(entityType, "_", " ")),
		entityID)

	return NewStructuredValidationError(
		EntityNotFoundType,
		EntityNotFoundCode,
		message,
	).WithEntity(entityType, entityID)
}

// NewRequiredFieldError creates a structured error for missing required fields
func NewRequiredFieldError(field, entityType string) *StructuredValidationError {
	message := fmt.Sprintf("Field '%s' is required for %s", field, entityType)

	return NewStructuredValidationError(
		GeneralValidationType,
		RequiredFieldCode,
		message,
	).WithField(field)
}

// ErrorFormatter provides consistent error message formatting
type ErrorFormatter struct{}

// NewErrorFormatter creates a new ErrorFormatter instance
func NewErrorFormatter() *ErrorFormatter {
	return &ErrorFormatter{}
}

// FormatStatusError formats status validation errors with helpful context
func (f *ErrorFormatter) FormatStatusError(entityType, providedValue string, validStatuses []string) string {
	if providedValue == "" {
		return fmt.Sprintf("Status is required for %s. Valid options: %s",
			entityType,
			f.formatStatusList(validStatuses))
	}

	return fmt.Sprintf("Invalid status '%s' for %s. Valid options: %s",
		providedValue,
		entityType,
		f.formatStatusList(validStatuses))
}

// FormatEntityNotFoundError formats entity not found errors
func (f *ErrorFormatter) FormatEntityNotFoundError(entityType, entityID string) string {
	return fmt.Sprintf("%s with ID '%s' not found",
		f.formatEntityType(entityType),
		entityID)
}

// formatStatusList formats the list of valid statuses for display
func (f *ErrorFormatter) formatStatusList(statuses []string) string {
	if len(statuses) == 0 {
		return "none available"
	}

	if len(statuses) == 1 {
		return fmt.Sprintf("'%s'", statuses[0])
	}

	if len(statuses) == 2 {
		return fmt.Sprintf("'%s' or '%s'", statuses[0], statuses[1])
	}

	// For more than 2 statuses, use comma separation with "or" before the last item
	formatted := make([]string, len(statuses))
	for i, status := range statuses {
		formatted[i] = fmt.Sprintf("'%s'", status)
	}

	lastIndex := len(formatted) - 1
	return strings.Join(formatted[:lastIndex], ", ") + ", or " + formatted[lastIndex]
}

// formatEntityType formats entity type names for display
func (f *ErrorFormatter) formatEntityType(entityType string) string {
	// Convert snake_case to Title Case
	words := strings.Split(entityType, "_")
	for i, word := range words {
		words[i] = strings.Title(word)
	}
	return strings.Join(words, " ")
}
