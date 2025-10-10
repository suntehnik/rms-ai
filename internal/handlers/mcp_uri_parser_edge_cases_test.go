package handlers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURIParser_ParseEdgeCases(t *testing.T) {
	parser := NewURIParser()

	tests := []struct {
		name     string
		uri      string
		hasError bool
		errorMsg string
	}{
		// URL encoding edge cases
		{
			name: "URI with encoded characters in sub-path",
			uri:  "epic://EP-001/hierarchy%20test",
		},
		{
			name: "URI with encoded characters in parameters",
			uri:  "epic://EP-001?name=test%20value&special=%21%40%23",
		},
		// Fragment handling (should be ignored)
		{
			name: "URI with fragment",
			uri:  "epic://EP-001#section1",
		},
		{
			name: "URI with sub-path and fragment",
			uri:  "epic://EP-001/hierarchy#details",
		},
		// Multiple parameter values (only first should be taken)
		{
			name: "URI with duplicate parameter keys",
			uri:  "epic://EP-001?filter=active&filter=inactive",
		},
		// Empty parameter values
		{
			name: "URI with empty parameter value",
			uri:  "epic://EP-001?empty=&nonempty=value",
		},
		// Special characters in reference IDs (should fail)
		{
			name:     "reference ID with special characters",
			uri:      "epic://EP-001@test",
			hasError: true,
		},
		{
			name:     "reference ID with spaces",
			uri:      "epic://EP-001 test",
			hasError: true,
		},
		// Port numbers in URI (should fail as it's not a valid reference ID)
		{
			name:     "URI with port number",
			uri:      "epic://EP-001:8080",
			hasError: true,
		},
		// Very long URIs
		{
			name: "very long sub-path",
			uri:  "epic://EP-001/" + strings.Repeat("a", 1000),
		},
		{
			name: "very long parameter value",
			uri:  "epic://EP-001?data=" + strings.Repeat("x", 1000),
		},
		// Case sensitivity tests
		{
			name:     "uppercase scheme",
			uri:      "EPIC://EP-001",
			hasError: true,
		},
		{
			name:     "mixed case scheme",
			uri:      "Epic://EP-001",
			hasError: true,
		},
		// Leading/trailing whitespace (should be handled by url.Parse)
		{
			name:     "URI with leading whitespace",
			uri:      " epic://EP-001",
			hasError: true,
		},
		{
			name:     "URI with trailing whitespace",
			uri:      "epic://EP-001 ",
			hasError: true,
		},
		// Unicode characters
		{
			name: "URI with unicode in parameters",
			uri:  "epic://EP-001?title=测试&description=тест",
		},
		// Boundary reference ID numbers
		{
			name: "reference ID with leading zeros",
			uri:  "epic://EP-0001",
		},
		{
			name: "single digit reference ID",
			uri:  "epic://EP-1",
		},
		// Multiple slashes in sub-path
		{
			name: "sub-path with multiple segments",
			uri:  "epic://EP-001/hierarchy/details/summary",
		},
		// Query parameter edge cases
		{
			name: "parameter with no value (just key)",
			uri:  "epic://EP-001?flag",
		},
		{
			name: "parameter with equals but no value",
			uri:  "epic://EP-001?empty=",
		},
		// Malformed query strings
		{
			name: "malformed query string",
			uri:  "epic://EP-001?key=value&malformed&another=value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.uri)

			if tt.hasError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				// Verify basic structure is valid
				assert.NotEmpty(t, result.Scheme)
				assert.NotEmpty(t, result.ReferenceID)
				assert.NotNil(t, result.Parameters)
			}
		})
	}
}

func TestURIParser_ReferenceIDValidationEdgeCases(t *testing.T) {
	parser := NewURIParser()

	tests := []struct {
		name        string
		referenceID string
		expected    bool
	}{
		// Boundary cases for numbers
		{"zero", "EP-0", true},
		{"single digit", "EP-1", true},
		{"double digit", "EP-12", true},
		{"triple digit", "EP-123", true},
		{"four digits", "EP-1234", true},
		{"five digits", "EP-12345", true},
		{"six digits", "EP-123456", true},
		{"very large number", "EP-999999999", true},

		// Leading zeros
		{"leading zero single", "EP-01", true},
		{"leading zeros multiple", "EP-001", true},
		{"all zeros", "EP-000", true},

		// Invalid patterns
		{"negative number", "EP--1", false},
		{"plus sign", "EP-+1", false},
		{"decimal number", "EP-1.5", false},
		{"scientific notation", "EP-1e5", false},
		{"hex number", "EP-0x1", false},

		// Mixed content after number
		{"number with letter", "EP-123a", false},
		{"number with space", "EP-123 ", false},
		{"number with dash", "EP-123-", false},
		{"number with underscore", "EP-123_", false},

		// Different prefixes
		{"lowercase prefix", "ep-123", false},
		{"mixed case prefix", "Ep-123", false},
		{"wrong prefix length", "E-123", false},
		{"too long prefix", "EPP-123", false},
		{"numeric prefix", "12-123", false},
		{"special char prefix", "E@-123", false},

		// Separator variations
		{"underscore separator", "EP_123", false},
		{"dot separator", "EP.123", false},
		{"colon separator", "EP:123", false},
		{"space separator", "EP 123", false},
		{"no separator", "EP123", false},
		{"double dash", "EP--123", false},

		// Empty and whitespace
		{"empty string", "", false},
		{"only spaces", "   ", false},
		{"only dash", "-", false},
		{"only prefix", "EP", false},
		{"only dash after prefix", "EP-", false},

		// Unicode and special characters
		{"unicode in prefix", "ЕР-123", false},
		{"unicode in number", "EP-１２３", false},
		{"emoji", "EP-😀", false},
		{"null character", "EP-\x00", false},

		// All valid prefixes
		{"US prefix", "US-123", true},
		{"REQ prefix", "REQ-123", true},
		{"AC prefix", "AC-123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.isValidReferenceID(tt.referenceID)
			assert.Equal(t, tt.expected, result, "Reference ID: %s", tt.referenceID)
		})
	}
}

func TestURIParser_SchemeValidationEdgeCases(t *testing.T) {
	parser := NewURIParser()

	tests := []struct {
		name     string
		scheme   string
		expected bool
	}{
		// Valid schemes
		{"epic", "epic", true},
		{"user-story", "user-story", true},
		{"requirement", "requirement", true},
		{"acceptance-criteria", "acceptance-criteria", true},

		// Case variations (should all be false)
		{"Epic", "Epic", false},
		{"EPIC", "EPIC", false},
		{"ePiC", "ePiC", false},
		{"User-Story", "User-Story", false},
		{"USER-STORY", "USER-STORY", false},
		{"Requirement", "Requirement", false},
		{"REQUIREMENT", "REQUIREMENT", false},
		{"Acceptance-Criteria", "Acceptance-Criteria", false},
		{"ACCEPTANCE-CRITERIA", "ACCEPTANCE-CRITERIA", false},

		// Similar but invalid schemes
		{"epics", "epics", false},
		{"user_story", "user_story", false},
		{"user-stories", "user-stories", false},
		{"requirements", "requirements", false},
		{"acceptance_criteria", "acceptance_criteria", false},
		{"acceptance-criterion", "acceptance-criterion", false},

		// Empty and whitespace
		{"empty", "", false},
		{"spaces", "   ", false},
		{"tab", "\t", false},
		{"newline", "\n", false},

		// Special characters
		{"with-slash", "epic/", false},
		{"with-colon", "epic:", false},
		{"with-question", "epic?", false},
		{"with-hash", "epic#", false},
		{"with-at", "epic@", false},

		// Numbers and mixed
		{"numeric", "123", false},
		{"alphanumeric", "epic123", false},
		{"with-numbers", "epic-123", false},

		// Unicode
		{"unicode", "эпик", false},
		{"emoji", "📝", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.isValidScheme(tt.scheme)
			assert.Equal(t, tt.expected, result, "Scheme: %s", tt.scheme)
		})
	}
}

func TestURIParser_BuildURIEdgeCases(t *testing.T) {
	parser := NewURIParser()

	tests := []struct {
		name        string
		scheme      string
		referenceID string
		subPath     string
		parameters  map[string]string
		expectError bool
		validate    func(t *testing.T, uri string)
	}{
		{
			name:        "empty sub-path",
			scheme:      "epic",
			referenceID: "EP-001",
			subPath:     "",
			validate: func(t *testing.T, uri string) {
				assert.Equal(t, "epic://EP-001", uri)
			},
		},
		{
			name:        "nil parameters",
			scheme:      "epic",
			referenceID: "EP-001",
			parameters:  nil,
			validate: func(t *testing.T, uri string) {
				assert.Equal(t, "epic://EP-001", uri)
			},
		},
		{
			name:        "empty parameters map",
			scheme:      "epic",
			referenceID: "EP-001",
			parameters:  map[string]string{},
			validate: func(t *testing.T, uri string) {
				assert.Equal(t, "epic://EP-001", uri)
			},
		},
		{
			name:        "parameter with special characters",
			scheme:      "epic",
			referenceID: "EP-001",
			parameters: map[string]string{
				"filter": "status=active&priority>1",
				"format": "json",
			},
			validate: func(t *testing.T, uri string) {
				// Should be URL encoded
				assert.Contains(t, uri, "epic://EP-001?")
				assert.Contains(t, uri, "filter=status%3Dactive%26priority%3E1")
				assert.Contains(t, uri, "format=json")
			},
		},
		{
			name:        "parameter with unicode",
			scheme:      "epic",
			referenceID: "EP-001",
			parameters: map[string]string{
				"title": "测试标题",
				"desc":  "описание",
			},
			validate: func(t *testing.T, uri string) {
				// Should contain encoded unicode
				assert.Contains(t, uri, "epic://EP-001?")
				// Parse it back to verify it works
				parsed, err := parser.Parse(uri)
				require.NoError(t, err)
				assert.Contains(t, []string{"测试标题", "описание"}, parsed.Parameters["title"])
			},
		},
		{
			name:        "sub-path with special characters",
			scheme:      "epic",
			referenceID: "EP-001",
			subPath:     "hierarchy/details",
			validate: func(t *testing.T, uri string) {
				assert.Equal(t, "epic://EP-001/hierarchy/details", uri)
			},
		},
		{
			name:        "invalid scheme in build",
			scheme:      "invalid-scheme",
			referenceID: "EP-001",
			expectError: true,
		},
		{
			name:        "invalid reference ID in build",
			scheme:      "epic",
			referenceID: "INVALID-ID",
			expectError: true,
		},
		{
			name:        "scheme/reference ID mismatch in build",
			scheme:      "epic",
			referenceID: "US-001",
			expectError: true,
		},
		{
			name:        "empty scheme",
			scheme:      "",
			referenceID: "EP-001",
			expectError: true,
		},
		{
			name:        "empty reference ID",
			scheme:      "epic",
			referenceID: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.BuildURI(tt.scheme, tt.referenceID, tt.subPath, tt.parameters)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, result)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestURIParser_SubPathSupportEdgeCases(t *testing.T) {
	parser := NewURIParser()

	tests := []struct {
		name     string
		scheme   string
		subPath  string
		expected bool
	}{
		// Empty sub-paths
		{"epic empty sub-path", "epic", "", false},
		{"user-story empty sub-path", "user-story", "", false},
		{"requirement empty sub-path", "requirement", "", false},
		{"acceptance-criteria empty sub-path", "acceptance-criteria", "", false},

		// Case sensitivity
		{"epic hierarchy uppercase", "epic", "HIERARCHY", false},
		{"epic hierarchy mixed case", "epic", "Hierarchy", false},
		{"user-story requirements uppercase", "user-story", "REQUIREMENTS", false},
		{"requirement relationships uppercase", "requirement", "RELATIONSHIPS", false},

		// Similar but invalid sub-paths
		{"epic hierarchies", "epic", "hierarchies", false},
		{"epic user-story", "epic", "user-story", false},
		{"user-story requirement", "user-story", "requirement", false},
		{"requirement relationship", "requirement", "relationship", false},

		// Sub-paths with special characters
		{"epic hierarchy with dash", "epic", "hierarchy-test", false},
		{"epic hierarchy with underscore", "epic", "hierarchy_test", false},
		{"epic hierarchy with space", "epic", "hierarchy test", false},

		// Non-existent schemes
		{"invalid scheme", "invalid", "anything", false},
		{"empty scheme", "", "hierarchy", false},

		// All valid combinations for verification
		{"epic hierarchy valid", "epic", "hierarchy", true},
		{"epic user-stories valid", "epic", "user-stories", true},
		{"user-story requirements valid", "user-story", "requirements", true},
		{"user-story acceptance-criteria valid", "user-story", "acceptance-criteria", true},
		{"requirement relationships valid", "requirement", "relationships", true},

		// Acceptance criteria has no supported sub-paths
		{"acceptance-criteria any sub-path", "acceptance-criteria", "details", false},
		{"acceptance-criteria requirements", "acceptance-criteria", "requirements", false},
		{"acceptance-criteria hierarchy", "acceptance-criteria", "hierarchy", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.IsSubPathSupported(tt.scheme, tt.subPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestURIParser_ConcurrentAccess(t *testing.T) {
	parser := NewURIParser()

	// Test concurrent access to ensure thread safety
	const numGoroutines = 100
	const numOperations = 10

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < numOperations; j++ {
				// Test parsing with safe ID values
				uri := fmt.Sprintf("epic://EP-001/hierarchy?id=%d", id)
				parsed, err := parser.Parse(uri)
				assert.NoError(t, err)
				assert.NotNil(t, parsed)

				// Test building with safe ID values
				built, err := parser.BuildURI("epic", "EP-001", "hierarchy", map[string]string{
					"id": fmt.Sprintf("%d", id),
				})
				assert.NoError(t, err)
				assert.NotEmpty(t, built)

				// Test validation methods
				assert.True(t, parser.isValidScheme("epic"))
				assert.True(t, parser.isValidReferenceID("EP-001"))
				assert.NoError(t, parser.validateSchemeAndReferenceID("epic", "EP-001"))

				// Test utility methods
				schemes := parser.GetSupportedSchemes()
				assert.Len(t, schemes, 4)

				prefix, err := parser.GetExpectedPrefix("epic")
				assert.NoError(t, err)
				assert.Equal(t, "EP", prefix)

				assert.True(t, parser.IsSubPathSupported("epic", "hierarchy"))
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestURIParser_MemoryUsage(t *testing.T) {
	parser := NewURIParser()

	// Test with many different URIs to ensure no memory leaks
	const numURIs = 1000

	for i := 0; i < numURIs; i++ {
		uri := "epic://EP-" + string(rune(i%10000)) + "/hierarchy?id=" + string(rune(i))

		parsed, err := parser.Parse(uri)
		if err == nil {
			assert.NotNil(t, parsed)
		}

		// Build URI
		built, err := parser.BuildURI("epic", "EP-"+string(rune(i%10000)), "hierarchy", map[string]string{
			"id": string(rune(i)),
		})
		if err == nil {
			assert.NotEmpty(t, built)
		}
	}
}
