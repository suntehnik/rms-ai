package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewURIParser(t *testing.T) {
	parser := NewURIParser()

	assert.NotNil(t, parser)
	assert.NotNil(t, parser.referenceIDPattern)
	assert.NotNil(t, parser.schemePrefixMap)
	assert.Len(t, parser.schemePrefixMap, 4)
}

func TestURIParser_Parse(t *testing.T) {
	parser := NewURIParser()

	tests := []struct {
		name     string
		uri      string
		expected *ParsedURI
		hasError bool
	}{
		// Valid basic URIs
		{
			name: "valid epic URI",
			uri:  "epic://EP-001",
			expected: &ParsedURI{
				Scheme:      "epic",
				ReferenceID: "EP-001",
				SubPath:     "",
				Parameters:  map[string]string{},
			},
		},
		{
			name: "valid user story URI",
			uri:  "user-story://US-042",
			expected: &ParsedURI{
				Scheme:      "user-story",
				ReferenceID: "US-042",
				SubPath:     "",
				Parameters:  map[string]string{},
			},
		},
		{
			name: "valid requirement URI",
			uri:  "requirement://REQ-123",
			expected: &ParsedURI{
				Scheme:      "requirement",
				ReferenceID: "REQ-123",
				SubPath:     "",
				Parameters:  map[string]string{},
			},
		},
		{
			name: "valid acceptance criteria URI",
			uri:  "acceptance-criteria://AC-005",
			expected: &ParsedURI{
				Scheme:      "acceptance-criteria",
				ReferenceID: "AC-005",
				SubPath:     "",
				Parameters:  map[string]string{},
			},
		},
		// Valid URIs with sub-paths
		{
			name: "epic with hierarchy sub-path",
			uri:  "epic://EP-001/hierarchy",
			expected: &ParsedURI{
				Scheme:      "epic",
				ReferenceID: "EP-001",
				SubPath:     "hierarchy",
				Parameters:  map[string]string{},
			},
		},
		{
			name: "user story with requirements sub-path",
			uri:  "user-story://US-001/requirements",
			expected: &ParsedURI{
				Scheme:      "user-story",
				ReferenceID: "US-001",
				SubPath:     "requirements",
				Parameters:  map[string]string{},
			},
		},
		{
			name: "requirement with relationships sub-path",
			uri:  "requirement://REQ-123/relationships",
			expected: &ParsedURI{
				Scheme:      "requirement",
				ReferenceID: "REQ-123",
				SubPath:     "relationships",
				Parameters:  map[string]string{},
			},
		},
		// Valid URIs with parameters
		{
			name: "epic with query parameters",
			uri:  "epic://EP-001?include=user-stories&format=json",
			expected: &ParsedURI{
				Scheme:      "epic",
				ReferenceID: "EP-001",
				SubPath:     "",
				Parameters: map[string]string{
					"include": "user-stories",
					"format":  "json",
				},
			},
		},
		{
			name: "user story with sub-path and parameters",
			uri:  "user-story://US-001/requirements?status=active",
			expected: &ParsedURI{
				Scheme:      "user-story",
				ReferenceID: "US-001",
				SubPath:     "requirements",
				Parameters: map[string]string{
					"status": "active",
				},
			},
		},
		// Invalid cases - empty URI
		{
			name:     "empty URI",
			uri:      "",
			hasError: true,
		},
		// Invalid cases - malformed URIs
		{
			name:     "malformed URI",
			uri:      "not-a-valid-uri",
			hasError: true,
		},
		// Invalid cases - unsupported schemes
		{
			name:     "unsupported scheme",
			uri:      "invalid://EP-001",
			hasError: true,
		},
		{
			name:     "missing scheme",
			uri:      "EP-001",
			hasError: true,
		},
		// Invalid cases - missing reference ID
		{
			name:     "missing reference ID",
			uri:      "epic://",
			hasError: true,
		},
		// Invalid cases - invalid reference ID format
		{
			name:     "invalid reference ID format - no dash",
			uri:      "epic://EP001",
			hasError: true,
		},
		{
			name:     "invalid reference ID format - no number",
			uri:      "epic://EP-",
			hasError: true,
		},
		{
			name:     "invalid reference ID format - letters after dash",
			uri:      "epic://EP-ABC",
			hasError: true,
		},
		// Invalid cases - scheme/prefix mismatch
		{
			name:     "epic scheme with user story prefix",
			uri:      "epic://US-001",
			hasError: true,
		},
		{
			name:     "user story scheme with epic prefix",
			uri:      "user-story://EP-001",
			hasError: true,
		},
		{
			name:     "requirement scheme with acceptance criteria prefix",
			uri:      "requirement://AC-001",
			hasError: true,
		},
		{
			name:     "acceptance criteria scheme with requirement prefix",
			uri:      "acceptance-criteria://REQ-001",
			hasError: true,
		},
		// Edge cases - large numbers
		{
			name: "large reference ID number",
			uri:  "epic://EP-999999",
			expected: &ParsedURI{
				Scheme:      "epic",
				ReferenceID: "EP-999999",
				SubPath:     "",
				Parameters:  map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.uri)

			if tt.hasError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.Scheme, result.Scheme)
				assert.Equal(t, tt.expected.ReferenceID, result.ReferenceID)
				assert.Equal(t, tt.expected.SubPath, result.SubPath)
				assert.Equal(t, tt.expected.Parameters, result.Parameters)
			}
		})
	}
}

func TestURIParser_isValidScheme(t *testing.T) {
	parser := NewURIParser()

	tests := []struct {
		scheme   string
		expected bool
	}{
		{"epic", true},
		{"user-story", true},
		{"requirement", true},
		{"acceptance-criteria", true},
		{"invalid", false},
		{"", false},
		{"Epic", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.scheme, func(t *testing.T) {
			result := parser.isValidScheme(tt.scheme)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestURIParser_isValidReferenceID(t *testing.T) {
	parser := NewURIParser()

	tests := []struct {
		referenceID string
		expected    bool
	}{
		// Valid formats
		{"EP-001", true},
		{"US-042", true},
		{"REQ-123", true},
		{"AC-005", true},
		{"EP-999999", true},
		// Invalid formats
		{"EP001", false},   // missing dash
		{"EP-", false},     // missing number
		{"EP-ABC", false},  // letters after dash
		{"ep-001", false},  // lowercase prefix
		{"EP_001", false},  // underscore instead of dash
		{"EP-01A", false},  // mixed alphanumeric
		{"", false},        // empty
		{"001", false},     // no prefix
		{"E-001", false},   // single letter prefix
		{"EPP-001", false}, // three letter prefix
	}

	for _, tt := range tests {
		t.Run(tt.referenceID, func(t *testing.T) {
			result := parser.isValidReferenceID(tt.referenceID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestURIParser_validateSchemeAndReferenceID(t *testing.T) {
	parser := NewURIParser()

	tests := []struct {
		name        string
		scheme      string
		referenceID string
		hasError    bool
	}{
		// Valid combinations
		{"epic with EP prefix", "epic", "EP-001", false},
		{"user-story with US prefix", "user-story", "US-001", false},
		{"requirement with REQ prefix", "requirement", "REQ-001", false},
		{"acceptance-criteria with AC prefix", "acceptance-criteria", "AC-001", false},
		// Invalid combinations
		{"epic with US prefix", "epic", "US-001", true},
		{"user-story with EP prefix", "user-story", "EP-001", true},
		{"requirement with AC prefix", "requirement", "AC-001", true},
		{"acceptance-criteria with REQ prefix", "acceptance-criteria", "REQ-001", true},
		// Invalid scheme
		{"invalid scheme", "invalid", "EP-001", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.validateSchemeAndReferenceID(tt.scheme, tt.referenceID)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestURIParser_GetSupportedSchemes(t *testing.T) {
	parser := NewURIParser()

	schemes := parser.GetSupportedSchemes()

	assert.Len(t, schemes, 4)
	assert.Contains(t, schemes, "epic")
	assert.Contains(t, schemes, "user-story")
	assert.Contains(t, schemes, "requirement")
	assert.Contains(t, schemes, "acceptance-criteria")
}

func TestURIParser_GetExpectedPrefix(t *testing.T) {
	parser := NewURIParser()

	tests := []struct {
		scheme         string
		expectedPrefix string
		hasError       bool
	}{
		{"epic", "EP", false},
		{"user-story", "US", false},
		{"requirement", "REQ", false},
		{"acceptance-criteria", "AC", false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.scheme, func(t *testing.T) {
			prefix, err := parser.GetExpectedPrefix(tt.scheme)

			if tt.hasError {
				assert.Error(t, err)
				assert.Empty(t, prefix)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPrefix, prefix)
			}
		})
	}
}

func TestURIParser_IsSubPathSupported(t *testing.T) {
	parser := NewURIParser()

	tests := []struct {
		scheme   string
		subPath  string
		expected bool
	}{
		// Epic sub-paths
		{"epic", "hierarchy", true},
		{"epic", "user-stories", true},
		{"epic", "invalid", false},
		// User story sub-paths
		{"user-story", "requirements", true},
		{"user-story", "acceptance-criteria", true},
		{"user-story", "invalid", false},
		// Requirement sub-paths
		{"requirement", "relationships", true},
		{"requirement", "invalid", false},
		// Acceptance criteria sub-paths (none supported)
		{"acceptance-criteria", "anything", false},
		// Invalid scheme
		{"invalid", "anything", false},
	}

	for _, tt := range tests {
		t.Run(tt.scheme+"_"+tt.subPath, func(t *testing.T) {
			result := parser.IsSubPathSupported(tt.scheme, tt.subPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestURIParser_BuildURI(t *testing.T) {
	parser := NewURIParser()

	tests := []struct {
		name        string
		scheme      string
		referenceID string
		subPath     string
		parameters  map[string]string
		expected    string
		hasError    bool
	}{
		// Valid basic URIs
		{
			name:        "basic epic URI",
			scheme:      "epic",
			referenceID: "EP-001",
			expected:    "epic://EP-001",
		},
		{
			name:        "basic user story URI",
			scheme:      "user-story",
			referenceID: "US-001",
			expected:    "user-story://US-001",
		},
		// URIs with sub-paths
		{
			name:        "epic with sub-path",
			scheme:      "epic",
			referenceID: "EP-001",
			subPath:     "hierarchy",
			expected:    "epic://EP-001/hierarchy",
		},
		// URIs with parameters
		{
			name:        "epic with parameters",
			scheme:      "epic",
			referenceID: "EP-001",
			parameters: map[string]string{
				"include": "user-stories",
				"format":  "json",
			},
			expected: "epic://EP-001?format=json&include=user-stories",
		},
		// URIs with sub-path and parameters
		{
			name:        "epic with sub-path and parameters",
			scheme:      "epic",
			referenceID: "EP-001",
			subPath:     "hierarchy",
			parameters: map[string]string{
				"depth": "2",
			},
			expected: "epic://EP-001/hierarchy?depth=2",
		},
		// Invalid cases
		{
			name:        "invalid scheme",
			scheme:      "invalid",
			referenceID: "EP-001",
			hasError:    true,
		},
		{
			name:        "invalid reference ID",
			scheme:      "epic",
			referenceID: "invalid",
			hasError:    true,
		},
		{
			name:        "scheme/prefix mismatch",
			scheme:      "epic",
			referenceID: "US-001",
			hasError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.BuildURI(tt.scheme, tt.referenceID, tt.subPath, tt.parameters)

			if tt.hasError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				// For URIs with parameters, we need to parse and compare since parameter order may vary
				if len(tt.parameters) > 0 {
					parsedResult, parseErr := parser.Parse(result)
					require.NoError(t, parseErr)
					parsedExpected, parseErr := parser.Parse(tt.expected)
					require.NoError(t, parseErr)

					assert.Equal(t, parsedExpected.Scheme, parsedResult.Scheme)
					assert.Equal(t, parsedExpected.ReferenceID, parsedResult.ReferenceID)
					assert.Equal(t, parsedExpected.SubPath, parsedResult.SubPath)
					assert.Equal(t, parsedExpected.Parameters, parsedResult.Parameters)
				} else {
					assert.Equal(t, tt.expected, result)
				}
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkURIParser_Parse(b *testing.B) {
	parser := NewURIParser()
	uri := "epic://EP-001/hierarchy?include=user-stories&format=json"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(uri)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkURIParser_BuildURI(b *testing.B) {
	parser := NewURIParser()
	parameters := map[string]string{
		"include": "user-stories",
		"format":  "json",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.BuildURI("epic", "EP-001", "hierarchy", parameters)
		if err != nil {
			b.Fatal(err)
		}
	}
}
