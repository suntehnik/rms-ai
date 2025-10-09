package jsonrpc

import (
	"errors"
	"fmt"
	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/service"
	"strings"
)

// Standard JSON-RPC 2.0 error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// Custom MCP error codes (starting from -32000 as per JSON-RPC spec)
const (
	ResourceNotFound   = -32001
	UnauthorizedAccess = -32002
	ValidationError    = -32003
	ServiceUnavailable = -32004
	RateLimitExceeded  = -32005
)

// ErrorMessages maps error codes to their standard messages
var ErrorMessages = map[int]string{
	ParseError:         "Parse error",
	InvalidRequest:     "Invalid Request",
	MethodNotFound:     "Method not found",
	InvalidParams:      "Invalid params",
	InternalError:      "Internal error",
	ResourceNotFound:   "Resource not found",
	UnauthorizedAccess: "Unauthorized access",
	ValidationError:    "Validation error",
	ServiceUnavailable: "Service unavailable",
	RateLimitExceeded:  "Rate limit exceeded",
}

// NewJSONRPCError creates a new JSON-RPC error
func NewJSONRPCError(code int, message string, data interface{}) *JSONRPCError {
	return &JSONRPCError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// NewStandardError creates a standard JSON-RPC error using predefined messages
func NewStandardError(code int, data interface{}) *JSONRPCError {
	message, exists := ErrorMessages[code]
	if !exists {
		message = "Unknown error"
	}
	return NewJSONRPCError(code, message, data)
}

// ErrorMapper provides methods to map service layer errors to JSON-RPC errors
type ErrorMapper struct{}

// NewErrorMapper creates a new error mapper
func NewErrorMapper() *ErrorMapper {
	return &ErrorMapper{}
}

// MapError maps a Go error to a JSON-RPC error
func (em *ErrorMapper) MapError(err error) *JSONRPCError {
	if err == nil {
		return nil
	}

	// Check if it's already a JSON-RPC error
	var jsonrpcErr *JSONRPCError
	if errors.As(err, &jsonrpcErr) {
		return jsonrpcErr
	}

	// Map specific service layer errors first
	switch {
	// Repository layer errors
	case errors.Is(err, repository.ErrNotFound):
		return NewStandardError(ResourceNotFound, err.Error())
	case errors.Is(err, repository.ErrInvalidID):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, repository.ErrDuplicateKey):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, repository.ErrForeignKey):
		return NewStandardError(ValidationError, err.Error())

	// Service layer errors - Epic
	case isEpicError(err):
		return mapEpicError(err)

	// Service layer errors - User Story
	case isUserStoryError(err):
		return mapUserStoryError(err)

	// Service layer errors - Requirement
	case isRequirementError(err):
		return mapRequirementError(err)

	// Service layer errors - Acceptance Criteria
	case isAcceptanceCriteriaError(err):
		return mapAcceptanceCriteriaError(err)

	// Service layer errors - Configuration
	case isConfigError(err):
		return mapConfigError(err)

	// Service layer errors - PAT Authentication
	case isPATError(err):
		return mapPATError(err)

	// Service layer errors - Comments
	case isCommentError(err):
		return mapCommentError(err)

	// Generic error pattern matching (fallback)
	case isNotFoundError(err):
		return NewStandardError(ResourceNotFound, err.Error())
	case isUnauthorizedError(err):
		return NewStandardError(UnauthorizedAccess, err.Error())
	case isValidationError(err):
		return NewStandardError(ValidationError, err.Error())
	case isServiceUnavailableError(err):
		return NewStandardError(ServiceUnavailable, err.Error())
	case isRateLimitError(err):
		return NewStandardError(RateLimitExceeded, err.Error())
	default:
		return NewStandardError(InternalError, "An unexpected error occurred")
	}
}

// Helper functions to identify error types
// These can be extended to match your service layer error patterns

func isNotFoundError(err error) bool {
	// Check for common "not found" error patterns
	errStr := err.Error()
	return contains(errStr, "not found") ||
		contains(errStr, "does not exist") ||
		contains(errStr, "record not found")
}

func isUnauthorizedError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "unauthorized") ||
		contains(errStr, "access denied") ||
		contains(errStr, "permission denied") ||
		contains(errStr, "forbidden")
}

func isValidationError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "validation") ||
		contains(errStr, "invalid") ||
		contains(errStr, "required") ||
		contains(errStr, "constraint")
}

func isServiceUnavailableError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "service unavailable") ||
		contains(errStr, "connection") ||
		contains(errStr, "timeout") ||
		contains(errStr, "database")
}

func isRateLimitError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "rate limit") ||
		contains(errStr, "too many requests") ||
		contains(errStr, "quota exceeded")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// Common error constructors for convenience

// NewParseError creates a parse error
func NewParseError(data interface{}) *JSONRPCError {
	return NewStandardError(ParseError, data)
}

// NewInvalidRequestError creates an invalid request error
func NewInvalidRequestError(data interface{}) *JSONRPCError {
	return NewStandardError(InvalidRequest, data)
}

// NewMethodNotFoundError creates a method not found error
func NewMethodNotFoundError(method string) *JSONRPCError {
	return NewStandardError(MethodNotFound, fmt.Sprintf("Method '%s' not found", method))
}

// NewInvalidParamsError creates an invalid params error
func NewInvalidParamsError(data interface{}) *JSONRPCError {
	return NewStandardError(InvalidParams, data)
}

// NewInternalError creates an internal error
func NewInternalError(data interface{}) *JSONRPCError {
	return NewStandardError(InternalError, data)
}

// NewResourceNotFoundError creates a resource not found error
func NewResourceNotFoundError(resource string) *JSONRPCError {
	return NewStandardError(ResourceNotFound, fmt.Sprintf("Resource '%s' not found", resource))
}

// NewUnauthorizedError creates an unauthorized access error
func NewUnauthorizedError(data interface{}) *JSONRPCError {
	return NewStandardError(UnauthorizedAccess, data)
}

// NewValidationError creates a validation error
func NewValidationError(data interface{}) *JSONRPCError {
	return NewStandardError(ValidationError, data)
}

// Service-specific error checking and mapping functions

// isEpicError checks if the error is related to Epic operations
func isEpicError(err error) bool {
	epicErrors := []error{
		service.ErrEpicNotFound,
		service.ErrEpicHasUserStories,
		service.ErrInvalidEpicStatus,
		service.ErrInvalidPriority,
		service.ErrUserNotFound,
	}

	for _, epicErr := range epicErrors {
		if errors.Is(err, epicErr) {
			return true
		}
	}
	return false
}

// mapEpicError maps Epic-specific errors to JSON-RPC errors
func mapEpicError(err error) *JSONRPCError {
	switch {
	case errors.Is(err, service.ErrEpicNotFound):
		return NewStandardError(ResourceNotFound, err.Error())
	case errors.Is(err, service.ErrEpicHasUserStories):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, service.ErrInvalidEpicStatus):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, service.ErrInvalidPriority):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, service.ErrUserNotFound):
		return NewStandardError(ResourceNotFound, err.Error())
	default:
		return NewStandardError(InternalError, "Epic operation failed")
	}
}

// isUserStoryError checks if the error is related to UserStory operations
func isUserStoryError(err error) bool {
	userStoryErrors := []error{
		service.ErrUserStoryNotFound,
		service.ErrUserStoryHasRequirements,
		service.ErrInvalidUserStoryStatus,
		service.ErrInvalidUserStoryTemplate,
	}

	for _, userStoryErr := range userStoryErrors {
		if errors.Is(err, userStoryErr) {
			return true
		}
	}
	return false
}

// mapUserStoryError maps UserStory-specific errors to JSON-RPC errors
func mapUserStoryError(err error) *JSONRPCError {
	switch {
	case errors.Is(err, service.ErrUserStoryNotFound):
		return NewStandardError(ResourceNotFound, err.Error())
	case errors.Is(err, service.ErrUserStoryHasRequirements):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, service.ErrInvalidUserStoryStatus):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, service.ErrInvalidUserStoryTemplate):
		return NewStandardError(ValidationError, err.Error())
	default:
		return NewStandardError(InternalError, "User story operation failed")
	}
}

// isRequirementError checks if the error is related to Requirement operations
func isRequirementError(err error) bool {
	requirementErrors := []error{
		service.ErrRequirementNotFound,
		service.ErrRequirementHasRelationships,
		service.ErrInvalidRequirementStatus,
		service.ErrCircularRelationship,
		service.ErrDuplicateRelationship,
	}

	for _, reqErr := range requirementErrors {
		if errors.Is(err, reqErr) {
			return true
		}
	}
	return false
}

// mapRequirementError maps Requirement-specific errors to JSON-RPC errors
func mapRequirementError(err error) *JSONRPCError {
	switch {
	case errors.Is(err, service.ErrRequirementNotFound):
		return NewStandardError(ResourceNotFound, err.Error())
	case errors.Is(err, service.ErrRequirementHasRelationships):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, service.ErrInvalidRequirementStatus):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, service.ErrCircularRelationship):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, service.ErrDuplicateRelationship):
		return NewStandardError(ValidationError, err.Error())
	default:
		return NewStandardError(InternalError, "Requirement operation failed")
	}
}

// isAcceptanceCriteriaError checks if the error is related to AcceptanceCriteria operations
func isAcceptanceCriteriaError(err error) bool {
	acErrors := []error{
		service.ErrAcceptanceCriteriaNotFound,
		service.ErrAcceptanceCriteriaHasRequirements,
		service.ErrUserStoryMustHaveAcceptanceCriteria,
	}

	for _, acErr := range acErrors {
		if errors.Is(err, acErr) {
			return true
		}
	}
	return false
}

// mapAcceptanceCriteriaError maps AcceptanceCriteria-specific errors to JSON-RPC errors
func mapAcceptanceCriteriaError(err error) *JSONRPCError {
	switch {
	case errors.Is(err, service.ErrAcceptanceCriteriaNotFound):
		return NewStandardError(ResourceNotFound, err.Error())
	case errors.Is(err, service.ErrAcceptanceCriteriaHasRequirements):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, service.ErrUserStoryMustHaveAcceptanceCriteria):
		return NewStandardError(ValidationError, err.Error())
	default:
		return NewStandardError(InternalError, "Acceptance criteria operation failed")
	}
}

// isConfigError checks if the error is related to Configuration operations
func isConfigError(err error) bool {
	configErrors := []error{
		service.ErrRequirementTypeNotFound,
		service.ErrRelationshipTypeNotFound,
		service.ErrInvalidStatusTransition,
		service.ErrRequirementTypeNameExists,
		service.ErrRequirementTypeHasRequirements,
		service.ErrRelationshipTypeNameExists,
		service.ErrRelationshipTypeHasRelationships,
		service.ErrStatusModelNameExists,
		service.ErrStatusModelNotFound,
		service.ErrStatusNotFound,
		service.ErrStatusTransitionNotFound,
		service.ErrStatusNameExists,
		service.ErrTransitionExists,
		service.ErrInvalidEntityType,
	}

	for _, configErr := range configErrors {
		if errors.Is(err, configErr) {
			return true
		}
	}
	return false
}

// mapConfigError maps Configuration-specific errors to JSON-RPC errors
func mapConfigError(err error) *JSONRPCError {
	switch {
	case errors.Is(err, service.ErrRequirementTypeNotFound):
		return NewStandardError(ResourceNotFound, err.Error())
	case errors.Is(err, service.ErrRelationshipTypeNotFound):
		return NewStandardError(ResourceNotFound, err.Error())
	case errors.Is(err, service.ErrStatusModelNotFound):
		return NewStandardError(ResourceNotFound, err.Error())
	case errors.Is(err, service.ErrStatusNotFound):
		return NewStandardError(ResourceNotFound, err.Error())
	case errors.Is(err, service.ErrStatusTransitionNotFound):
		return NewStandardError(ResourceNotFound, err.Error())
	case errors.Is(err, service.ErrInvalidStatusTransition):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, service.ErrRequirementTypeNameExists):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, service.ErrRequirementTypeHasRequirements):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, service.ErrRelationshipTypeNameExists):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, service.ErrRelationshipTypeHasRelationships):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, service.ErrStatusModelNameExists):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, service.ErrStatusNameExists):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, service.ErrTransitionExists):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, service.ErrInvalidEntityType):
		return NewStandardError(ValidationError, err.Error())
	default:
		return NewStandardError(InternalError, "Configuration operation failed")
	}
}

// isPATError checks if the error is related to PAT authentication
func isPATError(err error) bool {
	patErrors := []error{
		service.ErrPATNotFound,
		service.ErrPATExpired,
		service.ErrPATInvalidToken,
		service.ErrPATInvalidPrefix,
		service.ErrPATDuplicateName,
		service.ErrPATUserNotFound,
		service.ErrPATUnauthorized,
		service.ErrPATInvalidScopes,
		service.ErrPATTokenHashMismatch,
	}

	for _, patErr := range patErrors {
		if errors.Is(err, patErr) {
			return true
		}
	}
	return false
}

// mapPATError maps PAT-specific errors to JSON-RPC errors
func mapPATError(err error) *JSONRPCError {
	switch {
	case errors.Is(err, service.ErrPATNotFound):
		return NewStandardError(UnauthorizedAccess, "PAT token not found")
	case errors.Is(err, service.ErrPATExpired):
		return NewStandardError(UnauthorizedAccess, "PAT token expired")
	case errors.Is(err, service.ErrPATInvalidToken):
		return NewStandardError(UnauthorizedAccess, "Invalid PAT token")
	case errors.Is(err, service.ErrPATInvalidPrefix):
		return NewStandardError(UnauthorizedAccess, "Invalid PAT token format")
	case errors.Is(err, service.ErrPATUserNotFound):
		return NewStandardError(UnauthorizedAccess, "User not found")
	case errors.Is(err, service.ErrPATUnauthorized):
		return NewStandardError(UnauthorizedAccess, "Unauthorized access")
	case errors.Is(err, service.ErrPATTokenHashMismatch):
		return NewStandardError(UnauthorizedAccess, "Invalid PAT token")
	case errors.Is(err, service.ErrPATDuplicateName):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, service.ErrPATInvalidScopes):
		return NewStandardError(ValidationError, err.Error())
	default:
		return NewStandardError(InternalError, "PAT authentication failed")
	}
}

// isCommentError checks if the error is related to Comment operations
func isCommentError(err error) bool {
	commentErrors := []error{
		service.ErrCommentNotFound,
		service.ErrCommentHasReplies,
		service.ErrCommentInvalidEntityType,
		service.ErrCommentEntityNotFound,
		service.ErrCommentAuthorNotFound,
		service.ErrParentCommentNotFound,
	}

	for _, commentErr := range commentErrors {
		if errors.Is(err, commentErr) {
			return true
		}
	}
	return false
}

// mapCommentError maps Comment-specific errors to JSON-RPC errors
func mapCommentError(err error) *JSONRPCError {
	switch {
	case errors.Is(err, service.ErrCommentNotFound):
		return NewStandardError(ResourceNotFound, err.Error())
	case errors.Is(err, service.ErrCommentEntityNotFound):
		return NewStandardError(ResourceNotFound, err.Error())
	case errors.Is(err, service.ErrCommentAuthorNotFound):
		return NewStandardError(ResourceNotFound, err.Error())
	case errors.Is(err, service.ErrParentCommentNotFound):
		return NewStandardError(ResourceNotFound, err.Error())
	case errors.Is(err, service.ErrCommentHasReplies):
		return NewStandardError(ValidationError, err.Error())
	case errors.Is(err, service.ErrCommentInvalidEntityType):
		return NewStandardError(ValidationError, err.Error())
	default:
		return NewStandardError(InternalError, "Comment operation failed")
	}
}
