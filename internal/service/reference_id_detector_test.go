package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReferenceIDDetector_DetectPattern(t *testing.T) {
	detector := NewReferenceIDDetector()

	tests := []struct {
		name     string
		query    string
		expected ReferenceIDPattern
	}{
		// Epic patterns
		{
			name:  "valid epic reference ID uppercase",
			query: "EP-001",
			expected: ReferenceIDPattern{
				IsReferenceID: true,
				EntityType:    "epic",
				Number:        "001",
				OriginalQuery: "EP-001",
			},
		},
		{
			name:  "valid epic reference ID lowercase",
			query: "ep-119",
			expected: ReferenceIDPattern{
				IsReferenceID: true,
				EntityType:    "epic",
				Number:        "119",
				OriginalQuery: "ep-119",
			},
		},
		{
			name:  "valid epic reference ID mixed case",
			query: "Ep-042",
			expected: ReferenceIDPattern{
				IsReferenceID: true,
				EntityType:    "epic",
				Number:        "042",
				OriginalQuery: "Ep-042",
			},
		},

		// User Story patterns
		{
			name:  "valid user story reference ID uppercase",
			query: "US-119",
			expected: ReferenceIDPattern{
				IsReferenceID: true,
				EntityType:    "user_story",
				Number:        "119",
				OriginalQuery: "US-119",
			},
		},
		{
			name:  "valid user story reference ID lowercase",
			query: "us-001",
			expected: ReferenceIDPattern{
				IsReferenceID: true,
				EntityType:    "user_story",
				Number:        "001",
				OriginalQuery: "us-001",
			},
		},
		{
			name:  "valid user story reference ID mixed case",
			query: "Us-999",
			expected: ReferenceIDPattern{
				IsReferenceID: true,
				EntityType:    "user_story",
				Number:        "999",
				OriginalQuery: "Us-999",
			},
		},

		// Requirement patterns
		{
			name:  "valid requirement reference ID uppercase",
			query: "REQ-045",
			expected: ReferenceIDPattern{
				IsReferenceID: true,
				EntityType:    "requirement",
				Number:        "045",
				OriginalQuery: "REQ-045",
			},
		},
		{
			name:  "valid requirement reference ID lowercase",
			query: "req-123",
			expected: ReferenceIDPattern{
				IsReferenceID: true,
				EntityType:    "requirement",
				Number:        "123",
				OriginalQuery: "req-123",
			},
		},
		{
			name:  "valid requirement reference ID mixed case",
			query: "Req-007",
			expected: ReferenceIDPattern{
				IsReferenceID: true,
				EntityType:    "requirement",
				Number:        "007",
				OriginalQuery: "Req-007",
			},
		},

		// Acceptance Criteria patterns
		{
			name:  "valid acceptance criteria reference ID uppercase",
			query: "AC-023",
			expected: ReferenceIDPattern{
				IsReferenceID: true,
				EntityType:    "acceptance_criteria",
				Number:        "023",
				OriginalQuery: "AC-023",
			},
		},
		{
			name:  "valid acceptance criteria reference ID lowercase",
			query: "ac-456",
			expected: ReferenceIDPattern{
				IsReferenceID: true,
				EntityType:    "acceptance_criteria",
				Number:        "456",
				OriginalQuery: "ac-456",
			},
		},
		{
			name:  "valid acceptance criteria reference ID mixed case",
			query: "Ac-789",
			expected: ReferenceIDPattern{
				IsReferenceID: true,
				EntityType:    "acceptance_criteria",
				Number:        "789",
				OriginalQuery: "Ac-789",
			},
		},

		// Steering Document patterns
		{
			name:  "valid steering document reference ID uppercase",
			query: "STD-012",
			expected: ReferenceIDPattern{
				IsReferenceID: true,
				EntityType:    "steering_document",
				Number:        "012",
				OriginalQuery: "STD-012",
			},
		},
		{
			name:  "valid steering document reference ID lowercase",
			query: "std-345",
			expected: ReferenceIDPattern{
				IsReferenceID: true,
				EntityType:    "steering_document",
				Number:        "345",
				OriginalQuery: "std-345",
			},
		},
		{
			name:  "valid steering document reference ID mixed case",
			query: "Std-678",
			expected: ReferenceIDPattern{
				IsReferenceID: true,
				EntityType:    "steering_document",
				Number:        "678",
				OriginalQuery: "Std-678",
			},
		},

		// Invalid patterns
		{
			name:  "invalid pattern - no dash",
			query: "EP001",
			expected: ReferenceIDPattern{
				IsReferenceID: false,
				OriginalQuery: "EP001",
			},
		},
		{
			name:  "invalid pattern - no number",
			query: "EP-",
			expected: ReferenceIDPattern{
				IsReferenceID: false,
				OriginalQuery: "EP-",
			},
		},
		{
			name:  "invalid pattern - letters after dash",
			query: "EP-ABC",
			expected: ReferenceIDPattern{
				IsReferenceID: false,
				OriginalQuery: "EP-ABC",
			},
		},
		{
			name:  "invalid pattern - unknown prefix",
			query: "XY-123",
			expected: ReferenceIDPattern{
				IsReferenceID: false,
				OriginalQuery: "XY-123",
			},
		},
		{
			name:  "invalid pattern - extra characters",
			query: "EP-123-extra",
			expected: ReferenceIDPattern{
				IsReferenceID: false,
				OriginalQuery: "EP-123-extra",
			},
		},
		{
			name:  "invalid pattern - spaces",
			query: "EP - 123",
			expected: ReferenceIDPattern{
				IsReferenceID: false,
				OriginalQuery: "EP - 123",
			},
		},
		{
			name:  "invalid pattern - random text",
			query: "random text",
			expected: ReferenceIDPattern{
				IsReferenceID: false,
				OriginalQuery: "random text",
			},
		},
		{
			name:  "empty string",
			query: "",
			expected: ReferenceIDPattern{
				IsReferenceID: false,
				OriginalQuery: "",
			},
		},
		{
			name:  "whitespace only",
			query: "   ",
			expected: ReferenceIDPattern{
				IsReferenceID: false,
				OriginalQuery: "   ",
			},
		},

		// Edge cases with whitespace
		{
			name:  "valid pattern with leading whitespace",
			query: "  US-119",
			expected: ReferenceIDPattern{
				IsReferenceID: true,
				EntityType:    "user_story",
				Number:        "119",
				OriginalQuery: "  US-119",
			},
		},
		{
			name:  "valid pattern with trailing whitespace",
			query: "EP-006  ",
			expected: ReferenceIDPattern{
				IsReferenceID: true,
				EntityType:    "epic",
				Number:        "006",
				OriginalQuery: "EP-006  ",
			},
		},
		{
			name:  "valid pattern with surrounding whitespace",
			query: "  REQ-045  ",
			expected: ReferenceIDPattern{
				IsReferenceID: true,
				EntityType:    "requirement",
				Number:        "045",
				OriginalQuery: "  REQ-045  ",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.DetectPattern(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReferenceIDDetector_IsValidReferenceID(t *testing.T) {
	detector := NewReferenceIDDetector()

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		// Valid reference IDs
		{"valid epic uppercase", "EP-001", true},
		{"valid epic lowercase", "ep-119", true},
		{"valid user story uppercase", "US-119", true},
		{"valid user story lowercase", "us-001", true},
		{"valid requirement uppercase", "REQ-045", true},
		{"valid requirement lowercase", "req-123", true},
		{"valid acceptance criteria uppercase", "AC-023", true},
		{"valid acceptance criteria lowercase", "ac-456", true},
		{"valid steering document uppercase", "STD-012", true},
		{"valid steering document lowercase", "std-345", true},

		// Invalid reference IDs
		{"invalid pattern", "EP001", false},
		{"unknown prefix", "XY-123", false},
		{"random text", "random text", false},
		{"empty string", "", false},
		{"whitespace only", "   ", false},
		{"extra characters", "EP-123-extra", false},
		{"spaces in pattern", "EP - 123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.IsValidReferenceID(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReferenceIDDetector_GetEntityTypeFromReferenceID(t *testing.T) {
	detector := NewReferenceIDDetector()

	tests := []struct {
		name     string
		query    string
		expected string
	}{
		// Valid reference IDs
		{"epic uppercase", "EP-001", "epic"},
		{"epic lowercase", "ep-119", "epic"},
		{"user story uppercase", "US-119", "user_story"},
		{"user story lowercase", "us-001", "user_story"},
		{"requirement uppercase", "REQ-045", "requirement"},
		{"requirement lowercase", "req-123", "requirement"},
		{"acceptance criteria uppercase", "AC-023", "acceptance_criteria"},
		{"acceptance criteria lowercase", "ac-456", "acceptance_criteria"},
		{"steering document uppercase", "STD-012", "steering_document"},
		{"steering document lowercase", "std-345", "steering_document"},

		// Invalid reference IDs should return empty string
		{"invalid pattern", "EP001", ""},
		{"unknown prefix", "XY-123", ""},
		{"random text", "random text", ""},
		{"empty string", "", ""},
		{"whitespace only", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.GetEntityTypeFromReferenceID(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReferenceIDDetector_CaseInsensitiveMatching(t *testing.T) {
	detector := NewReferenceIDDetector()

	// Test that case variations of the same reference ID all match
	testCases := []string{
		"EP-119", "ep-119", "Ep-119", "eP-119",
		"US-119", "us-119", "Us-119", "uS-119",
		"REQ-045", "req-045", "Req-045", "rEq-045", "reQ-045", "REq-045", "rEQ-045",
		"AC-023", "ac-023", "Ac-023", "aC-023",
		"STD-012", "std-012", "Std-012", "sTd-012", "stD-012", "STd-012", "StD-012", "sTD-012",
	}

	for _, testCase := range testCases {
		t.Run("case insensitive: "+testCase, func(t *testing.T) {
			result := detector.DetectPattern(testCase)
			assert.True(t, result.IsReferenceID, "Expected %s to be detected as reference ID", testCase)
			assert.NotEmpty(t, result.EntityType, "Expected entity type to be set for %s", testCase)
			assert.NotEmpty(t, result.Number, "Expected number to be extracted for %s", testCase)
		})
	}
}

func TestNewReferenceIDDetector(t *testing.T) {
	detector := NewReferenceIDDetector()

	// Verify that the detector is properly initialized
	assert.NotNil(t, detector)
	assert.NotNil(t, detector.patterns)
	assert.Len(t, detector.patterns, 5) // Should have 5 entity types

	// Verify all expected patterns are present
	expectedEntityTypes := []string{"epic", "user_story", "requirement", "acceptance_criteria", "steering_document"}
	for _, entityType := range expectedEntityTypes {
		assert.Contains(t, detector.patterns, entityType, "Expected pattern for entity type %s", entityType)
		assert.NotNil(t, detector.patterns[entityType], "Expected non-nil pattern for entity type %s", entityType)
	}
}
