package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStatusValidator(t *testing.T) {
	validator := NewStatusValidator()
	assert.NotNil(t, validator)
	assert.Implements(t, (*StatusValidator)(nil), validator)
}

func TestValidateEpicStatus(t *testing.T) {
	validator := NewStatusValidator()

	tests := []struct {
		name          string
		status        string
		expectError   bool
		errorContains string
	}{
		// Valid statuses
		{
			name:        "valid status - Backlog",
			status:      "Backlog",
			expectError: false,
		},
		{
			name:        "valid status - Draft",
			status:      "Draft",
			expectError: false,
		},
		{
			name:        "valid status - In Progress",
			status:      "In Progress",
			expectError: false,
		},
		{
			name:        "valid status - Done",
			status:      "Done",
			expectError: false,
		},
		{
			name:        "valid status - Cancelled",
			status:      "Cancelled",
			expectError: false,
		},
		// Case-insensitive validation
		{
			name:        "case insensitive - backlog",
			status:      "backlog",
			expectError: false,
		},
		{
			name:        "case insensitive - DRAFT",
			status:      "DRAFT",
			expectError: false,
		},
		{
			name:        "case insensitive - in progress",
			status:      "in progress",
			expectError: false,
		},
		{
			name:        "case insensitive - InProgress",
			status:      "InProgress",
			expectError: false,
		},
		{
			name:        "case insensitive - in_progress",
			status:      "in_progress",
			expectError: false,
		},
		{
			name:        "case insensitive - done",
			status:      "done",
			expectError: false,
		},
		{
			name:        "case insensitive - cancelled",
			status:      "cancelled",
			expectError: false,
		},
		{
			name:        "case insensitive - canceled (alternative spelling)",
			status:      "canceled",
			expectError: false,
		},
		// Invalid statuses
		{
			name:          "invalid status",
			status:        "InvalidStatus",
			expectError:   true,
			errorContains: "Invalid status 'InvalidStatus' for epic",
		},
		{
			name:          "empty status",
			status:        "",
			expectError:   true,
			errorContains: "Status is required for epic",
		},
		{
			name:          "whitespace only status",
			status:        "   ",
			expectError:   true,
			errorContains: "Invalid status '   ' for epic",
		},
		{
			name:          "partial match",
			status:        "Back",
			expectError:   true,
			errorContains: "Invalid status 'Back' for epic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateEpicStatus(tt.status)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				// Verify it's a StatusValidationError
				assert.True(t, IsStatusValidationError(err))

				statusErr, ok := GetStatusValidationError(err)
				require.True(t, ok)
				assert.Equal(t, "epic", statusErr.EntityType)
				assert.Contains(t, statusErr.ValidStatuses, "Backlog")
				assert.Contains(t, statusErr.ValidStatuses, "Draft")
				assert.Contains(t, statusErr.ValidStatuses, "In Progress")
				assert.Contains(t, statusErr.ValidStatuses, "Done")
				assert.Contains(t, statusErr.ValidStatuses, "Cancelled")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUserStoryStatus(t *testing.T) {
	validator := NewStatusValidator()

	tests := []struct {
		name          string
		status        string
		expectError   bool
		errorContains string
	}{
		// Valid statuses
		{
			name:        "valid status - Backlog",
			status:      "Backlog",
			expectError: false,
		},
		{
			name:        "valid status - Draft",
			status:      "Draft",
			expectError: false,
		},
		{
			name:        "valid status - In Progress",
			status:      "In Progress",
			expectError: false,
		},
		{
			name:        "valid status - Done",
			status:      "Done",
			expectError: false,
		},
		{
			name:        "valid status - Cancelled",
			status:      "Cancelled",
			expectError: false,
		},
		// Case-insensitive validation
		{
			name:        "case insensitive - backlog",
			status:      "backlog",
			expectError: false,
		},
		{
			name:        "case insensitive - in progress variations",
			status:      "inprogress",
			expectError: false,
		},
		// Invalid statuses
		{
			name:          "invalid status",
			status:        "InvalidStatus",
			expectError:   true,
			errorContains: "Invalid status 'InvalidStatus' for user story",
		},
		{
			name:          "empty status",
			status:        "",
			expectError:   true,
			errorContains: "Status is required for user story",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateUserStoryStatus(tt.status)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				// Verify it's a StatusValidationError
				assert.True(t, IsStatusValidationError(err))

				statusErr, ok := GetStatusValidationError(err)
				require.True(t, ok)
				assert.Equal(t, "user story", statusErr.EntityType)
				assert.Contains(t, statusErr.ValidStatuses, "Backlog")
				assert.Contains(t, statusErr.ValidStatuses, "Draft")
				assert.Contains(t, statusErr.ValidStatuses, "In Progress")
				assert.Contains(t, statusErr.ValidStatuses, "Done")
				assert.Contains(t, statusErr.ValidStatuses, "Cancelled")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRequirementStatus(t *testing.T) {
	validator := NewStatusValidator()

	tests := []struct {
		name          string
		status        string
		expectError   bool
		errorContains string
	}{
		// Valid statuses
		{
			name:        "valid status - Draft",
			status:      "Draft",
			expectError: false,
		},
		{
			name:        "valid status - Active",
			status:      "Active",
			expectError: false,
		},
		{
			name:        "valid status - Obsolete",
			status:      "Obsolete",
			expectError: false,
		},
		// Case-insensitive validation
		{
			name:        "case insensitive - draft",
			status:      "draft",
			expectError: false,
		},
		{
			name:        "case insensitive - ACTIVE",
			status:      "ACTIVE",
			expectError: false,
		},
		{
			name:        "case insensitive - obsolete",
			status:      "obsolete",
			expectError: false,
		},
		// Invalid statuses
		{
			name:          "invalid status",
			status:        "InvalidStatus",
			expectError:   true,
			errorContains: "Invalid status 'InvalidStatus' for requirement",
		},
		{
			name:          "empty status",
			status:        "",
			expectError:   true,
			errorContains: "Status is required for requirement",
		},
		{
			name:          "epic status on requirement",
			status:        "Backlog",
			expectError:   true,
			errorContains: "Invalid status 'Backlog' for requirement",
		},
		{
			name:          "user story status on requirement",
			status:        "In Progress",
			expectError:   true,
			errorContains: "Invalid status 'In Progress' for requirement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateRequirementStatus(tt.status)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				// Verify it's a StatusValidationError
				assert.True(t, IsStatusValidationError(err))

				statusErr, ok := GetStatusValidationError(err)
				require.True(t, ok)
				assert.Equal(t, "requirement", statusErr.EntityType)
				assert.Contains(t, statusErr.ValidStatuses, "Draft")
				assert.Contains(t, statusErr.ValidStatuses, "Active")
				assert.Contains(t, statusErr.ValidStatuses, "Obsolete")
				assert.Len(t, statusErr.ValidStatuses, 3) // Only 3 valid statuses for requirements
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNormalizeStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Exact matches
		{
			name:     "exact match - Backlog",
			input:    "Backlog",
			expected: "Backlog",
		},
		{
			name:     "exact match - In Progress",
			input:    "In Progress",
			expected: "In Progress",
		},
		// Case variations
		{
			name:     "lowercase - backlog",
			input:    "backlog",
			expected: "Backlog",
		},
		{
			name:     "uppercase - DRAFT",
			input:    "DRAFT",
			expected: "Draft",
		},
		{
			name:     "mixed case - DoNe",
			input:    "DoNe",
			expected: "Done",
		},
		// In Progress variations
		{
			name:     "in progress lowercase",
			input:    "in progress",
			expected: "In Progress",
		},
		{
			name:     "inprogress no space",
			input:    "inprogress",
			expected: "In Progress",
		},
		{
			name:     "in_progress underscore",
			input:    "in_progress",
			expected: "In Progress",
		},
		// Cancelled variations
		{
			name:     "cancelled UK spelling",
			input:    "cancelled",
			expected: "Cancelled",
		},
		{
			name:     "canceled US spelling",
			input:    "canceled",
			expected: "Cancelled",
		},
		// Whitespace handling
		{
			name:     "with leading/trailing spaces",
			input:    "  Draft  ",
			expected: "Draft",
		},
		// No normalization rule
		{
			name:     "unknown status",
			input:    "UnknownStatus",
			expected: "UnknownStatus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeStatus(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetValidStatusLists(t *testing.T) {
	t.Run("epic statuses", func(t *testing.T) {
		statuses := GetValidEpicStatuses()
		expected := []string{"Backlog", "Draft", "In Progress", "Done", "Cancelled"}
		assert.Equal(t, expected, statuses)
	})

	t.Run("user story statuses", func(t *testing.T) {
		statuses := GetValidUserStoryStatuses()
		expected := []string{"Backlog", "Draft", "In Progress", "Done", "Cancelled"}
		assert.Equal(t, expected, statuses)
	})

	t.Run("requirement statuses", func(t *testing.T) {
		statuses := GetValidRequirementStatuses()
		expected := []string{"Draft", "Active", "Obsolete"}
		assert.Equal(t, expected, statuses)
	})
}
