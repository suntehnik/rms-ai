package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusValidationError(t *testing.T) {
	t.Run("create status validation error", func(t *testing.T) {
		validStatuses := []string{"Draft", "Active", "Obsolete"}
		err := NewStatusValidationError("requirement", "InvalidStatus", validStatuses)

		assert.NotNil(t, err)
		assert.Equal(t, "requirement", err.EntityType)
		assert.Equal(t, "InvalidStatus", err.ProvidedValue)
		assert.Equal(t, validStatuses, err.ValidStatuses)
		assert.Contains(t, err.Message, "Invalid status 'InvalidStatus' for requirement")
		assert.Contains(t, err.Message, "Draft, Active, Obsolete")
	})

	t.Run("create status validation error for empty value", func(t *testing.T) {
		validStatuses := []string{"Backlog", "Draft", "In Progress", "Done", "Cancelled"}
		err := NewStatusValidationError("epic", "", validStatuses)

		assert.NotNil(t, err)
		assert.Equal(t, "epic", err.EntityType)
		assert.Equal(t, "", err.ProvidedValue)
		assert.Equal(t, validStatuses, err.ValidStatuses)
		assert.Contains(t, err.Message, "Status is required for epic")
		assert.Contains(t, err.Message, "Backlog, Draft, In Progress, Done, Cancelled")
	})

	t.Run("error interface implementation", func(t *testing.T) {
		validStatuses := []string{"Draft", "Active"}
		err := NewStatusValidationError("requirement", "BadStatus", validStatuses)

		// Test that it implements error interface
		var errorInterface error = err
		assert.NotNil(t, errorInterface)
		assert.Equal(t, err.Message, errorInterface.Error())
	})
}

func TestIsStatusValidationError(t *testing.T) {
	t.Run("is status validation error", func(t *testing.T) {
		validStatuses := []string{"Draft", "Active"}
		err := NewStatusValidationError("requirement", "BadStatus", validStatuses)

		assert.True(t, IsStatusValidationError(err))
	})

	t.Run("is not status validation error", func(t *testing.T) {
		err := assert.AnError

		assert.False(t, IsStatusValidationError(err))
	})
}

func TestGetStatusValidationError(t *testing.T) {
	t.Run("get status validation error", func(t *testing.T) {
		validStatuses := []string{"Draft", "Active"}
		originalErr := NewStatusValidationError("requirement", "BadStatus", validStatuses)

		statusErr, ok := GetStatusValidationError(originalErr)

		assert.True(t, ok)
		assert.Equal(t, originalErr, statusErr)
	})

	t.Run("get non-status validation error", func(t *testing.T) {
		err := assert.AnError

		statusErr, ok := GetStatusValidationError(err)

		assert.False(t, ok)
		assert.Nil(t, statusErr)
	})
}

func TestStructuredValidationError(t *testing.T) {
	t.Run("create structured validation error", func(t *testing.T) {
		err := NewStructuredValidationError(StatusValidationType, InvalidStatusCode, "Test message")

		assert.NotNil(t, err)
		assert.Equal(t, StatusValidationType, err.Type)
		assert.Equal(t, InvalidStatusCode, err.Code)
		assert.Equal(t, "Test message", err.Message)
		assert.Nil(t, err.Details)
		assert.Empty(t, err.EntityType)
		assert.Empty(t, err.EntityID)
		assert.Empty(t, err.Field)
	})

	t.Run("structured error with details", func(t *testing.T) {
		details := map[string]interface{}{
			"provided_value": "InvalidStatus",
			"valid_statuses": []string{"Draft", "Active"},
		}

		err := NewStructuredValidationError(StatusValidationType, InvalidStatusCode, "Test message").
			WithDetails(details)

		assert.Equal(t, details, err.Details)
	})

	t.Run("structured error with entity", func(t *testing.T) {
		err := NewStructuredValidationError(EntityNotFoundType, EntityNotFoundCode, "Test message").
			WithEntity("requirement", "123e4567-e89b-12d3-a456-426614174000")

		assert.Equal(t, "requirement", err.EntityType)
		assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", err.EntityID)
	})

	t.Run("structured error with field", func(t *testing.T) {
		err := NewStructuredValidationError(GeneralValidationType, RequiredFieldCode, "Test message").
			WithField("status")

		assert.Equal(t, "status", err.Field)
	})

	t.Run("method chaining", func(t *testing.T) {
		details := map[string]string{"key": "value"}

		err := NewStructuredValidationError(StatusValidationType, InvalidStatusCode, "Test message").
			WithDetails(details).
			WithEntity("epic", "test-id").
			WithField("status")

		assert.Equal(t, StatusValidationType, err.Type)
		assert.Equal(t, InvalidStatusCode, err.Code)
		assert.Equal(t, "Test message", err.Message)
		assert.Equal(t, details, err.Details)
		assert.Equal(t, "epic", err.EntityType)
		assert.Equal(t, "test-id", err.EntityID)
		assert.Equal(t, "status", err.Field)
	})

	t.Run("error interface implementation", func(t *testing.T) {
		err := NewStructuredValidationError(StatusValidationType, InvalidStatusCode, "Test message")

		// Test that it implements error interface
		var errorInterface error = err
		assert.NotNil(t, errorInterface)
		assert.Equal(t, "Test message", errorInterface.Error())
	})
}

func TestNewInvalidStatusError(t *testing.T) {
	validStatuses := []string{"Draft", "Active", "Obsolete"}
	err := NewInvalidStatusError("requirement", "BadStatus", validStatuses)

	assert.NotNil(t, err)
	assert.Equal(t, StatusValidationType, err.Type)
	assert.Equal(t, InvalidStatusCode, err.Code)
	assert.Contains(t, err.Message, "Invalid status 'BadStatus' for requirement")
	assert.Equal(t, "status", err.Field)

	// Check details
	details, ok := err.Details.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "BadStatus", details["provided_value"])
	assert.Equal(t, validStatuses, details["valid_statuses"])
}

func TestNewEntityNotFoundError(t *testing.T) {
	err := NewEntityNotFoundError("user_story", "US-001")

	assert.NotNil(t, err)
	assert.Equal(t, EntityNotFoundType, err.Type)
	assert.Equal(t, EntityNotFoundCode, err.Code)
	assert.Contains(t, err.Message, "User Story with ID 'US-001' not found")
	assert.Equal(t, "user_story", err.EntityType)
	assert.Equal(t, "US-001", err.EntityID)
}

func TestNewRequiredFieldError(t *testing.T) {
	err := NewRequiredFieldError("status", "epic")

	assert.NotNil(t, err)
	assert.Equal(t, GeneralValidationType, err.Type)
	assert.Equal(t, RequiredFieldCode, err.Code)
	assert.Contains(t, err.Message, "Field 'status' is required for epic")
	assert.Equal(t, "status", err.Field)
}

func TestErrorFormatter(t *testing.T) {
	formatter := NewErrorFormatter()
	assert.NotNil(t, formatter)

	t.Run("format status error with value", func(t *testing.T) {
		validStatuses := []string{"Draft", "Active", "Obsolete"}
		message := formatter.FormatStatusError("requirement", "BadStatus", validStatuses)

		expected := "Invalid status 'BadStatus' for requirement. Valid options: 'Draft', 'Active', or 'Obsolete'"
		assert.Equal(t, expected, message)
	})

	t.Run("format status error without value", func(t *testing.T) {
		validStatuses := []string{"Backlog", "Draft"}
		message := formatter.FormatStatusError("epic", "", validStatuses)

		expected := "Status is required for epic. Valid options: 'Backlog' or 'Draft'"
		assert.Equal(t, expected, message)
	})

	t.Run("format entity not found error", func(t *testing.T) {
		message := formatter.FormatEntityNotFoundError("user_story", "US-001")

		expected := "User Story with ID 'US-001' not found"
		assert.Equal(t, expected, message)
	})
}

func TestFormatStatusList(t *testing.T) {
	formatter := NewErrorFormatter()

	tests := []struct {
		name     string
		statuses []string
		expected string
	}{
		{
			name:     "empty list",
			statuses: []string{},
			expected: "none available",
		},
		{
			name:     "single status",
			statuses: []string{"Draft"},
			expected: "'Draft'",
		},
		{
			name:     "two statuses",
			statuses: []string{"Draft", "Active"},
			expected: "'Draft' or 'Active'",
		},
		{
			name:     "three statuses",
			statuses: []string{"Draft", "Active", "Obsolete"},
			expected: "'Draft', 'Active', or 'Obsolete'",
		},
		{
			name:     "five statuses",
			statuses: []string{"Backlog", "Draft", "In Progress", "Done", "Cancelled"},
			expected: "'Backlog', 'Draft', 'In Progress', 'Done', or 'Cancelled'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.formatStatusList(tt.statuses)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatEntityType(t *testing.T) {
	formatter := NewErrorFormatter()

	tests := []struct {
		name       string
		entityType string
		expected   string
	}{
		{
			name:       "single word",
			entityType: "epic",
			expected:   "Epic",
		},
		{
			name:       "snake_case",
			entityType: "user_story",
			expected:   "User Story",
		},
		{
			name:       "multiple underscores",
			entityType: "acceptance_criteria",
			expected:   "Acceptance Criteria",
		},
		{
			name:       "already formatted",
			entityType: "requirement",
			expected:   "Requirement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.formatEntityType(tt.entityType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
