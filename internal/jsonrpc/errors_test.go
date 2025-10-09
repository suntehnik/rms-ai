package jsonrpc

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/service"
)

func TestNewJSONRPCError(t *testing.T) {
	code := InvalidRequest
	message := "Invalid request"
	data := "test data"

	err := NewJSONRPCError(code, message, data)

	assert.Equal(t, code, err.Code)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, data, err.Data)
}

func TestNewStandardError(t *testing.T) {
	tests := []struct {
		name        string
		code        int
		data        interface{}
		expectedMsg string
	}{
		{
			name:        "parse error",
			code:        ParseError,
			data:        "test data",
			expectedMsg: "Parse error",
		},
		{
			name:        "invalid request",
			code:        InvalidRequest,
			data:        "test data",
			expectedMsg: "Invalid Request",
		},
		{
			name:        "method not found",
			code:        MethodNotFound,
			data:        "test data",
			expectedMsg: "Method not found",
		},
		{
			name:        "invalid params",
			code:        InvalidParams,
			data:        "test data",
			expectedMsg: "Invalid params",
		},
		{
			name:        "internal error",
			code:        InternalError,
			data:        "test data",
			expectedMsg: "Internal error",
		},
		{
			name:        "resource not found",
			code:        ResourceNotFound,
			data:        "test data",
			expectedMsg: "Resource not found",
		},
		{
			name:        "unauthorized access",
			code:        UnauthorizedAccess,
			data:        "test data",
			expectedMsg: "Unauthorized access",
		},
		{
			name:        "validation error",
			code:        ValidationError,
			data:        "test data",
			expectedMsg: "Validation error",
		},
		{
			name:        "service unavailable",
			code:        ServiceUnavailable,
			data:        "test data",
			expectedMsg: "Service unavailable",
		},
		{
			name:        "rate limit exceeded",
			code:        RateLimitExceeded,
			data:        "test data",
			expectedMsg: "Rate limit exceeded",
		},
		{
			name:        "unknown error code",
			code:        -99999,
			data:        "test data",
			expectedMsg: "Unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewStandardError(tt.code, tt.data)
			assert.Equal(t, tt.code, err.Code)
			assert.Equal(t, tt.expectedMsg, err.Message)
			assert.Equal(t, tt.data, err.Data)
		})
	}
}

func TestErrorMapper_MapError(t *testing.T) {
	mapper := NewErrorMapper()

	tests := []struct {
		name         string
		inputError   error
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "nil error",
			inputError:   nil,
			expectedCode: 0,
		},
		{
			name:         "already JSON-RPC error",
			inputError:   NewJSONRPCError(InvalidRequest, "Invalid request", nil),
			expectedCode: InvalidRequest,
			expectedMsg:  "Invalid request",
		},
		{
			name:         "repository not found error",
			inputError:   repository.ErrNotFound,
			expectedCode: ResourceNotFound,
			expectedMsg:  "Resource not found",
		},
		{
			name:         "repository invalid ID error",
			inputError:   repository.ErrInvalidID,
			expectedCode: ValidationError,
			expectedMsg:  "Validation error",
		},
		{
			name:         "repository duplicate key error",
			inputError:   repository.ErrDuplicateKey,
			expectedCode: ValidationError,
			expectedMsg:  "Validation error",
		},
		{
			name:         "repository foreign key error",
			inputError:   repository.ErrForeignKey,
			expectedCode: ValidationError,
			expectedMsg:  "Validation error",
		},
		{
			name:         "epic not found error",
			inputError:   service.ErrEpicNotFound,
			expectedCode: ResourceNotFound,
			expectedMsg:  "Resource not found",
		},
		{
			name:         "epic has user stories error",
			inputError:   service.ErrEpicHasUserStories,
			expectedCode: ValidationError,
			expectedMsg:  "Validation error",
		},
		{
			name:         "invalid epic status error",
			inputError:   service.ErrInvalidEpicStatus,
			expectedCode: ValidationError,
			expectedMsg:  "Validation error",
		},
		{
			name:         "invalid priority error",
			inputError:   service.ErrInvalidPriority,
			expectedCode: ValidationError,
			expectedMsg:  "Validation error",
		},
		{
			name:         "user not found error",
			inputError:   service.ErrUserNotFound,
			expectedCode: ResourceNotFound,
			expectedMsg:  "Resource not found",
		},
		{
			name:         "user story not found error",
			inputError:   service.ErrUserStoryNotFound,
			expectedCode: ResourceNotFound,
			expectedMsg:  "Resource not found",
		},
		{
			name:         "user story has requirements error",
			inputError:   service.ErrUserStoryHasRequirements,
			expectedCode: ValidationError,
			expectedMsg:  "Validation error",
		},
		{
			name:         "invalid user story status error",
			inputError:   service.ErrInvalidUserStoryStatus,
			expectedCode: ValidationError,
			expectedMsg:  "Validation error",
		},
		{
			name:         "requirement not found error",
			inputError:   service.ErrRequirementNotFound,
			expectedCode: ResourceNotFound,
			expectedMsg:  "Resource not found",
		},
		{
			name:         "requirement has relationships error",
			inputError:   service.ErrRequirementHasRelationships,
			expectedCode: ValidationError,
			expectedMsg:  "Validation error",
		},
		{
			name:         "invalid requirement status error",
			inputError:   service.ErrInvalidRequirementStatus,
			expectedCode: ValidationError,
			expectedMsg:  "Validation error",
		},
		{
			name:         "circular relationship error",
			inputError:   service.ErrCircularRelationship,
			expectedCode: ValidationError,
			expectedMsg:  "Validation error",
		},
		{
			name:         "duplicate relationship error",
			inputError:   service.ErrDuplicateRelationship,
			expectedCode: ValidationError,
			expectedMsg:  "Validation error",
		},
		{
			name:         "acceptance criteria not found error",
			inputError:   service.ErrAcceptanceCriteriaNotFound,
			expectedCode: ResourceNotFound,
			expectedMsg:  "Resource not found",
		},
		{
			name:         "acceptance criteria has requirements error",
			inputError:   service.ErrAcceptanceCriteriaHasRequirements,
			expectedCode: ValidationError,
			expectedMsg:  "Validation error",
		},
		{
			name:         "PAT not found error",
			inputError:   service.ErrPATNotFound,
			expectedCode: UnauthorizedAccess,
			expectedMsg:  "Unauthorized access",
		},
		{
			name:         "PAT expired error",
			inputError:   service.ErrPATExpired,
			expectedCode: UnauthorizedAccess,
			expectedMsg:  "Unauthorized access",
		},
		{
			name:         "PAT invalid token error",
			inputError:   service.ErrPATInvalidToken,
			expectedCode: UnauthorizedAccess,
			expectedMsg:  "Unauthorized access",
		},
		{
			name:         "PAT unauthorized error",
			inputError:   service.ErrPATUnauthorized,
			expectedCode: UnauthorizedAccess,
			expectedMsg:  "Unauthorized access",
		},
		{
			name:         "comment not found error",
			inputError:   service.ErrCommentNotFound,
			expectedCode: ResourceNotFound,
			expectedMsg:  "Resource not found",
		},
		{
			name:         "comment has replies error",
			inputError:   service.ErrCommentHasReplies,
			expectedCode: ValidationError,
			expectedMsg:  "Validation error",
		},
		{
			name:         "generic not found error",
			inputError:   errors.New("entity not found"),
			expectedCode: ResourceNotFound,
			expectedMsg:  "Resource not found",
		},
		{
			name:         "generic unauthorized error",
			inputError:   errors.New("access denied"),
			expectedCode: UnauthorizedAccess,
			expectedMsg:  "Unauthorized access",
		},
		{
			name:         "generic validation error",
			inputError:   errors.New("validation failed"),
			expectedCode: ValidationError,
			expectedMsg:  "Validation error",
		},
		{
			name:         "generic service unavailable error",
			inputError:   errors.New("database connection failed"),
			expectedCode: ServiceUnavailable,
			expectedMsg:  "Service unavailable",
		},
		{
			name:         "generic rate limit error",
			inputError:   errors.New("rate limit exceeded"),
			expectedCode: RateLimitExceeded,
			expectedMsg:  "Rate limit exceeded",
		},
		{
			name:         "unknown error",
			inputError:   errors.New("some unknown error"),
			expectedCode: InternalError,
			expectedMsg:  "Internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapError(tt.inputError)

			if tt.inputError == nil {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.expectedMsg, result.Message)
		})
	}
}

func TestErrorConstructors(t *testing.T) {
	t.Run("NewParseError", func(t *testing.T) {
		err := NewParseError("test data")
		assert.Equal(t, ParseError, err.Code)
		assert.Equal(t, "Parse error", err.Message)
		assert.Equal(t, "test data", err.Data)
	})

	t.Run("NewInvalidRequestError", func(t *testing.T) {
		err := NewInvalidRequestError("test data")
		assert.Equal(t, InvalidRequest, err.Code)
		assert.Equal(t, "Invalid Request", err.Message)
		assert.Equal(t, "test data", err.Data)
	})

	t.Run("NewMethodNotFoundError", func(t *testing.T) {
		err := NewMethodNotFoundError("test_method")
		assert.Equal(t, MethodNotFound, err.Code)
		assert.Equal(t, "Method not found", err.Message)
		assert.Equal(t, "Method 'test_method' not found", err.Data)
	})

	t.Run("NewInvalidParamsError", func(t *testing.T) {
		err := NewInvalidParamsError("test data")
		assert.Equal(t, InvalidParams, err.Code)
		assert.Equal(t, "Invalid params", err.Message)
		assert.Equal(t, "test data", err.Data)
	})

	t.Run("NewInternalError", func(t *testing.T) {
		err := NewInternalError("test data")
		assert.Equal(t, InternalError, err.Code)
		assert.Equal(t, "Internal error", err.Message)
		assert.Equal(t, "test data", err.Data)
	})

	t.Run("NewResourceNotFoundError", func(t *testing.T) {
		err := NewResourceNotFoundError("test_resource")
		assert.Equal(t, ResourceNotFound, err.Code)
		assert.Equal(t, "Resource not found", err.Message)
		assert.Equal(t, "Resource 'test_resource' not found", err.Data)
	})

	t.Run("NewUnauthorizedError", func(t *testing.T) {
		err := NewUnauthorizedError("test data")
		assert.Equal(t, UnauthorizedAccess, err.Code)
		assert.Equal(t, "Unauthorized access", err.Message)
		assert.Equal(t, "test data", err.Data)
	})

	t.Run("NewValidationError", func(t *testing.T) {
		err := NewValidationError("test data")
		assert.Equal(t, ValidationError, err.Code)
		assert.Equal(t, "Validation error", err.Message)
		assert.Equal(t, "test data", err.Data)
	})
}

func TestErrorPatternMatching(t *testing.T) {
	tests := []struct {
		name     string
		error    error
		function func(error) bool
		expected bool
	}{
		{
			name:     "isNotFoundError - positive case",
			error:    errors.New("entity not found"),
			function: isNotFoundError,
			expected: true,
		},
		{
			name:     "isNotFoundError - does not exist case",
			error:    errors.New("record does not exist"),
			function: isNotFoundError,
			expected: true,
		},
		{
			name:     "isNotFoundError - record not found case",
			error:    errors.New("record not found"),
			function: isNotFoundError,
			expected: true,
		},
		{
			name:     "isNotFoundError - negative case",
			error:    errors.New("some other error"),
			function: isNotFoundError,
			expected: false,
		},
		{
			name:     "isUnauthorizedError - unauthorized case",
			error:    errors.New("unauthorized access"),
			function: isUnauthorizedError,
			expected: true,
		},
		{
			name:     "isUnauthorizedError - access denied case",
			error:    errors.New("access denied"),
			function: isUnauthorizedError,
			expected: true,
		},
		{
			name:     "isUnauthorizedError - permission denied case",
			error:    errors.New("permission denied"),
			function: isUnauthorizedError,
			expected: true,
		},
		{
			name:     "isUnauthorizedError - forbidden case",
			error:    errors.New("forbidden operation"),
			function: isUnauthorizedError,
			expected: true,
		},
		{
			name:     "isUnauthorizedError - negative case",
			error:    errors.New("some other error"),
			function: isUnauthorizedError,
			expected: false,
		},
		{
			name:     "isValidationError - validation case",
			error:    errors.New("validation failed"),
			function: isValidationError,
			expected: true,
		},
		{
			name:     "isValidationError - invalid case",
			error:    errors.New("invalid input"),
			function: isValidationError,
			expected: true,
		},
		{
			name:     "isValidationError - required case",
			error:    errors.New("field is required"),
			function: isValidationError,
			expected: true,
		},
		{
			name:     "isValidationError - constraint case",
			error:    errors.New("constraint violation"),
			function: isValidationError,
			expected: true,
		},
		{
			name:     "isValidationError - negative case",
			error:    errors.New("some other error"),
			function: isValidationError,
			expected: false,
		},
		{
			name:     "isServiceUnavailableError - service unavailable case",
			error:    errors.New("service unavailable"),
			function: isServiceUnavailableError,
			expected: true,
		},
		{
			name:     "isServiceUnavailableError - connection case",
			error:    errors.New("connection failed"),
			function: isServiceUnavailableError,
			expected: true,
		},
		{
			name:     "isServiceUnavailableError - timeout case",
			error:    errors.New("request timeout"),
			function: isServiceUnavailableError,
			expected: true,
		},
		{
			name:     "isServiceUnavailableError - database case",
			error:    errors.New("database error"),
			function: isServiceUnavailableError,
			expected: true,
		},
		{
			name:     "isServiceUnavailableError - negative case",
			error:    errors.New("some other error"),
			function: isServiceUnavailableError,
			expected: false,
		},
		{
			name:     "isRateLimitError - rate limit case",
			error:    errors.New("rate limit exceeded"),
			function: isRateLimitError,
			expected: true,
		},
		{
			name:     "isRateLimitError - too many requests case",
			error:    errors.New("too many requests"),
			function: isRateLimitError,
			expected: true,
		},
		{
			name:     "isRateLimitError - quota exceeded case",
			error:    errors.New("quota exceeded"),
			function: isRateLimitError,
			expected: true,
		},
		{
			name:     "isRateLimitError - negative case",
			error:    errors.New("some other error"),
			function: isRateLimitError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function(tt.error)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestServiceSpecificErrorChecking(t *testing.T) {
	tests := []struct {
		name     string
		error    error
		function func(error) bool
		expected bool
	}{
		{
			name:     "isEpicError - positive case",
			error:    service.ErrEpicNotFound,
			function: isEpicError,
			expected: true,
		},
		{
			name:     "isEpicError - negative case",
			error:    service.ErrUserStoryNotFound,
			function: isEpicError,
			expected: false,
		},
		{
			name:     "isUserStoryError - positive case",
			error:    service.ErrUserStoryNotFound,
			function: isUserStoryError,
			expected: true,
		},
		{
			name:     "isUserStoryError - negative case",
			error:    service.ErrEpicNotFound,
			function: isUserStoryError,
			expected: false,
		},
		{
			name:     "isRequirementError - positive case",
			error:    service.ErrRequirementNotFound,
			function: isRequirementError,
			expected: true,
		},
		{
			name:     "isRequirementError - negative case",
			error:    service.ErrEpicNotFound,
			function: isRequirementError,
			expected: false,
		},
		{
			name:     "isAcceptanceCriteriaError - positive case",
			error:    service.ErrAcceptanceCriteriaNotFound,
			function: isAcceptanceCriteriaError,
			expected: true,
		},
		{
			name:     "isAcceptanceCriteriaError - negative case",
			error:    service.ErrEpicNotFound,
			function: isAcceptanceCriteriaError,
			expected: false,
		},
		{
			name:     "isPATError - positive case",
			error:    service.ErrPATNotFound,
			function: isPATError,
			expected: true,
		},
		{
			name:     "isPATError - negative case",
			error:    service.ErrEpicNotFound,
			function: isPATError,
			expected: false,
		},
		{
			name:     "isCommentError - positive case",
			error:    service.ErrCommentNotFound,
			function: isCommentError,
			expected: true,
		},
		{
			name:     "isCommentError - negative case",
			error:    service.ErrEpicNotFound,
			function: isCommentError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function(tt.error)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestServiceSpecificErrorMapping(t *testing.T) {
	tests := []struct {
		name         string
		error        error
		function     func(error) *JSONRPCError
		expectedCode int
	}{
		{
			name:         "mapEpicError - epic not found",
			error:        service.ErrEpicNotFound,
			function:     mapEpicError,
			expectedCode: ResourceNotFound,
		},
		{
			name:         "mapEpicError - epic has user stories",
			error:        service.ErrEpicHasUserStories,
			function:     mapEpicError,
			expectedCode: ValidationError,
		},
		{
			name:         "mapEpicError - invalid epic status",
			error:        service.ErrInvalidEpicStatus,
			function:     mapEpicError,
			expectedCode: ValidationError,
		},
		{
			name:         "mapEpicError - invalid priority",
			error:        service.ErrInvalidPriority,
			function:     mapEpicError,
			expectedCode: ValidationError,
		},
		{
			name:         "mapEpicError - user not found",
			error:        service.ErrUserNotFound,
			function:     mapEpicError,
			expectedCode: ResourceNotFound,
		},
		{
			name:         "mapEpicError - unknown epic error",
			error:        errors.New("unknown epic error"),
			function:     mapEpicError,
			expectedCode: InternalError,
		},
		{
			name:         "mapUserStoryError - user story not found",
			error:        service.ErrUserStoryNotFound,
			function:     mapUserStoryError,
			expectedCode: ResourceNotFound,
		},
		{
			name:         "mapUserStoryError - user story has requirements",
			error:        service.ErrUserStoryHasRequirements,
			function:     mapUserStoryError,
			expectedCode: ValidationError,
		},
		{
			name:         "mapUserStoryError - invalid user story status",
			error:        service.ErrInvalidUserStoryStatus,
			function:     mapUserStoryError,
			expectedCode: ValidationError,
		},
		{
			name:         "mapUserStoryError - unknown user story error",
			error:        errors.New("unknown user story error"),
			function:     mapUserStoryError,
			expectedCode: InternalError,
		},
		{
			name:         "mapRequirementError - requirement not found",
			error:        service.ErrRequirementNotFound,
			function:     mapRequirementError,
			expectedCode: ResourceNotFound,
		},
		{
			name:         "mapRequirementError - requirement has relationships",
			error:        service.ErrRequirementHasRelationships,
			function:     mapRequirementError,
			expectedCode: ValidationError,
		},
		{
			name:         "mapRequirementError - invalid requirement status",
			error:        service.ErrInvalidRequirementStatus,
			function:     mapRequirementError,
			expectedCode: ValidationError,
		},
		{
			name:         "mapRequirementError - circular relationship",
			error:        service.ErrCircularRelationship,
			function:     mapRequirementError,
			expectedCode: ValidationError,
		},
		{
			name:         "mapRequirementError - duplicate relationship",
			error:        service.ErrDuplicateRelationship,
			function:     mapRequirementError,
			expectedCode: ValidationError,
		},
		{
			name:         "mapRequirementError - unknown requirement error",
			error:        errors.New("unknown requirement error"),
			function:     mapRequirementError,
			expectedCode: InternalError,
		},
		{
			name:         "mapAcceptanceCriteriaError - acceptance criteria not found",
			error:        service.ErrAcceptanceCriteriaNotFound,
			function:     mapAcceptanceCriteriaError,
			expectedCode: ResourceNotFound,
		},
		{
			name:         "mapAcceptanceCriteriaError - acceptance criteria has requirements",
			error:        service.ErrAcceptanceCriteriaHasRequirements,
			function:     mapAcceptanceCriteriaError,
			expectedCode: ValidationError,
		},
		{
			name:         "mapAcceptanceCriteriaError - unknown acceptance criteria error",
			error:        errors.New("unknown acceptance criteria error"),
			function:     mapAcceptanceCriteriaError,
			expectedCode: InternalError,
		},
		{
			name:         "mapPATError - PAT not found",
			error:        service.ErrPATNotFound,
			function:     mapPATError,
			expectedCode: UnauthorizedAccess,
		},
		{
			name:         "mapPATError - PAT expired",
			error:        service.ErrPATExpired,
			function:     mapPATError,
			expectedCode: UnauthorizedAccess,
		},
		{
			name:         "mapPATError - PAT invalid token",
			error:        service.ErrPATInvalidToken,
			function:     mapPATError,
			expectedCode: UnauthorizedAccess,
		},
		{
			name:         "mapPATError - PAT unauthorized",
			error:        service.ErrPATUnauthorized,
			function:     mapPATError,
			expectedCode: UnauthorizedAccess,
		},
		{
			name:         "mapPATError - PAT duplicate name",
			error:        service.ErrPATDuplicateName,
			function:     mapPATError,
			expectedCode: ValidationError,
		},
		{
			name:         "mapPATError - PAT invalid scopes",
			error:        service.ErrPATInvalidScopes,
			function:     mapPATError,
			expectedCode: ValidationError,
		},
		{
			name:         "mapPATError - unknown PAT error",
			error:        errors.New("unknown PAT error"),
			function:     mapPATError,
			expectedCode: InternalError,
		},
		{
			name:         "mapCommentError - comment not found",
			error:        service.ErrCommentNotFound,
			function:     mapCommentError,
			expectedCode: ResourceNotFound,
		},
		{
			name:         "mapCommentError - comment has replies",
			error:        service.ErrCommentHasReplies,
			function:     mapCommentError,
			expectedCode: ValidationError,
		},
		{
			name:         "mapCommentError - comment invalid entity type",
			error:        service.ErrCommentInvalidEntityType,
			function:     mapCommentError,
			expectedCode: ValidationError,
		},
		{
			name:         "mapCommentError - unknown comment error",
			error:        errors.New("unknown comment error"),
			function:     mapCommentError,
			expectedCode: InternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function(tt.error)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedCode, result.Code)
		})
	}
}

func TestErrorMessages(t *testing.T) {
	// Test that all error codes have corresponding messages
	expectedCodes := []int{
		ParseError,
		InvalidRequest,
		MethodNotFound,
		InvalidParams,
		InternalError,
		ResourceNotFound,
		UnauthorizedAccess,
		ValidationError,
		ServiceUnavailable,
		RateLimitExceeded,
	}

	for _, code := range expectedCodes {
		message, exists := ErrorMessages[code]
		assert.True(t, exists, "Error code %d should have a message", code)
		assert.NotEmpty(t, message, "Error message for code %d should not be empty", code)
	}
}

func TestContainsHelper(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "exact match",
			s:        "test string",
			substr:   "test string",
			expected: true,
		},
		{
			name:     "substring match",
			s:        "test string",
			substr:   "test",
			expected: true,
		},
		{
			name:     "case insensitive match",
			s:        "Test String",
			substr:   "test",
			expected: true,
		},
		{
			name:     "no match",
			s:        "test string",
			substr:   "xyz",
			expected: false,
		},
		{
			name:     "empty substring",
			s:        "test string",
			substr:   "",
			expected: true,
		},
		{
			name:     "empty string",
			s:        "",
			substr:   "test",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}
