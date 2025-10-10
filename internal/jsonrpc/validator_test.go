package jsonrpc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONRPCID_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    interface{}
		expectError bool
	}{
		{
			name:     "null value",
			input:    "null",
			expected: nil,
		},
		{
			name:     "integer value",
			input:    "123",
			expected: 123,
		},
		{
			name:     "string value",
			input:    `"test-id"`,
			expected: "test-id",
		},
		{
			name:     "float value",
			input:    "123.45",
			expected: 123.45,
		},
		{
			name:        "invalid JSON",
			input:       `{invalid}`,
			expectError: true,
		},
		{
			name:        "array value (invalid)",
			input:       `[1,2,3]`,
			expectError: true,
		},
		{
			name:        "object value (invalid)",
			input:       `{"key":"value"}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id JSONRPCID
			err := id.UnmarshalJSON([]byte(tt.input))

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, id.Value)
			}
		})
	}
}

func TestJSONRPCID_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		id       JSONRPCID
		expected string
	}{
		{
			name:     "null value",
			id:       JSONRPCID{Value: nil},
			expected: "null",
		},
		{
			name:     "integer value",
			id:       JSONRPCID{Value: 123},
			expected: "123",
		},
		{
			name:     "string value",
			id:       JSONRPCID{Value: "test-id"},
			expected: `"test-id"`,
		},
		{
			name:     "float value",
			id:       JSONRPCID{Value: 123.45},
			expected: "123.45",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.id.MarshalJSON()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))
		})
	}
}

func TestJSONRPCID_IsNull(t *testing.T) {
	tests := []struct {
		name     string
		id       *JSONRPCID
		expected bool
	}{
		{
			name:     "nil pointer",
			id:       nil,
			expected: true,
		},
		{
			name:     "nil value",
			id:       &JSONRPCID{Value: nil},
			expected: true,
		},
		{
			name:     "non-nil value",
			id:       &JSONRPCID{Value: 123},
			expected: false,
		},
		{
			name:     "zero value",
			id:       &JSONRPCID{Value: 0},
			expected: false,
		},
		{
			name:     "empty string",
			id:       &JSONRPCID{Value: ""},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.id.IsNull()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJSONRPCID_GetValue(t *testing.T) {
	tests := []struct {
		name     string
		id       *JSONRPCID
		expected interface{}
	}{
		{
			name:     "nil pointer",
			id:       nil,
			expected: nil,
		},
		{
			name:     "nil value",
			id:       &JSONRPCID{Value: nil},
			expected: nil,
		},
		{
			name:     "integer value",
			id:       &JSONRPCID{Value: 123},
			expected: 123,
		},
		{
			name:     "string value",
			id:       &JSONRPCID{Value: "test"},
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.id.GetValue()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJSONRPCRequest_Validation(t *testing.T) {
	tests := []struct {
		name        string
		request     JSONRPCRequest
		expectValid bool
	}{
		{
			name: "valid request with ID",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      &JSONRPCID{Value: 1},
				Method:  "test_method",
				Params:  map[string]interface{}{"key": "value"},
			},
			expectValid: true,
		},
		{
			name: "valid request without params",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      &JSONRPCID{Value: 1},
				Method:  "test_method",
			},
			expectValid: true,
		},
		{
			name: "valid notification (no ID)",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "test_method",
				Params:  map[string]interface{}{"key": "value"},
			},
			expectValid: true,
		},
		{
			name: "invalid jsonrpc version",
			request: JSONRPCRequest{
				JSONRPC: "1.0",
				ID:      &JSONRPCID{Value: 1},
				Method:  "test_method",
			},
			expectValid: false,
		},
		{
			name: "missing method",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      &JSONRPCID{Value: 1},
			},
			expectValid: false,
		},
		{
			name: "empty method",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      &JSONRPCID{Value: 1},
				Method:  "",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation through parsing
			data, err := tt.request.MarshalJSON()
			require.NoError(t, err)

			_, parseErr := ParseRequest(data)

			if tt.expectValid {
				assert.NoError(t, parseErr)
			} else {
				assert.Error(t, parseErr)
			}
		})
	}
}

func TestJSONRPCNotification_Validation(t *testing.T) {
	tests := []struct {
		name         string
		notification JSONRPCNotification
		expectValid  bool
	}{
		{
			name: "valid notification with params",
			notification: JSONRPCNotification{
				JSONRPC: "2.0",
				Method:  "test_method",
				Params:  map[string]interface{}{"key": "value"},
			},
			expectValid: true,
		},
		{
			name: "valid notification without params",
			notification: JSONRPCNotification{
				JSONRPC: "2.0",
				Method:  "test_method",
			},
			expectValid: true,
		},
		{
			name: "invalid jsonrpc version",
			notification: JSONRPCNotification{
				JSONRPC: "1.0",
				Method:  "test_method",
			},
			expectValid: false,
		},
		{
			name: "missing method",
			notification: JSONRPCNotification{
				JSONRPC: "2.0",
			},
			expectValid: false,
		},
		{
			name: "empty method",
			notification: JSONRPCNotification{
				JSONRPC: "2.0",
				Method:  "",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation through parsing
			data, err := tt.notification.MarshalJSON()
			require.NoError(t, err)

			_, parseErr := ParseNotification(data)

			if tt.expectValid {
				assert.NoError(t, parseErr)
			} else {
				assert.Error(t, parseErr)
			}
		})
	}
}

func TestJSONRPCResponse_Validation(t *testing.T) {
	tests := []struct {
		name     string
		response JSONRPCResponse
		isValid  bool
	}{
		{
			name: "valid success response",
			response: JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      &JSONRPCID{Value: 1},
				Result:  map[string]interface{}{"status": "ok"},
			},
			isValid: true,
		},
		{
			name: "valid error response",
			response: JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      &JSONRPCID{Value: 1},
				Error: &JSONRPCError{
					Code:    InvalidRequest,
					Message: "Invalid request",
				},
			},
			isValid: true,
		},
		{
			name: "response with both result and error (invalid)",
			response: JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      &JSONRPCID{Value: 1},
				Result:  map[string]interface{}{"status": "ok"},
				Error: &JSONRPCError{
					Code:    InvalidRequest,
					Message: "Invalid request",
				},
			},
			isValid: false, // Should have either result OR error, not both
		},
		{
			name: "response with neither result nor error (invalid)",
			response: JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      &JSONRPCID{Value: 1},
			},
			isValid: false, // Should have either result OR error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that response can be marshaled/unmarshaled
			data, err := tt.response.MarshalJSON()
			require.NoError(t, err)

			var parsed JSONRPCResponse
			err = parsed.UnmarshalJSON(data)
			require.NoError(t, err)

			// Validate response structure
			hasResult := parsed.Result != nil
			hasError := parsed.Error != nil

			if tt.isValid {
				// Valid responses should have exactly one of result or error
				assert.True(t, hasResult != hasError, "Response should have either result or error, but not both")
			} else {
				// Invalid responses might have both or neither
				assert.False(t, hasResult != hasError, "Invalid response structure detected")
			}
		})
	}
}

func TestJSONRPCError_ErrorInterface(t *testing.T) {
	tests := []struct {
		name     string
		err      JSONRPCError
		expected string
	}{
		{
			name: "standard error",
			err: JSONRPCError{
				Code:    InvalidRequest,
				Message: "Invalid request",
			},
			expected: "JSON-RPC error -32600: Invalid request",
		},
		{
			name: "custom error",
			err: JSONRPCError{
				Code:    ResourceNotFound,
				Message: "Resource not found",
			},
			expected: "JSON-RPC error -32001: Resource not found",
		},
		{
			name: "error with data",
			err: JSONRPCError{
				Code:    ValidationError,
				Message: "Validation failed",
				Data:    "Field 'name' is required",
			},
			expected: "JSON-RPC error -32003: Validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResponseConstructors(t *testing.T) {
	t.Run("NewSuccessResponse with various ID types", func(t *testing.T) {
		tests := []struct {
			name string
			id   interface{}
		}{
			{"integer ID", 123},
			{"string ID", "test-id"},
			{"float ID", 123.45},
			{"nil ID", nil},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := map[string]interface{}{"status": "ok"}
				response := NewSuccessResponse(tt.id, result)

				assert.Equal(t, JSONRPCVersion, response.JSONRPC)
				assert.Equal(t, result, response.Result)
				assert.Nil(t, response.Error)

				if tt.id != nil {
					require.NotNil(t, response.ID)
					assert.Equal(t, tt.id, response.ID.GetValue())
				} else {
					assert.Nil(t, response.ID)
				}
			})
		}
	})

	t.Run("NewErrorResponse with various ID types", func(t *testing.T) {
		tests := []struct {
			name string
			id   interface{}
		}{
			{"integer ID", 123},
			{"string ID", "test-id"},
			{"float ID", 123.45},
			{"nil ID", nil},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := NewJSONRPCError(InvalidRequest, "Invalid request", nil)
				response := NewErrorResponse(tt.id, err)

				assert.Equal(t, JSONRPCVersion, response.JSONRPC)
				assert.Nil(t, response.Result)
				assert.Equal(t, err, response.Error)

				if tt.id != nil {
					require.NotNil(t, response.ID)
					assert.Equal(t, tt.id, response.ID.GetValue())
				} else {
					assert.Nil(t, response.ID)
				}
			})
		}
	})
}

func TestJSONRPCConstants(t *testing.T) {
	// Test that the JSON-RPC version constant is correct
	assert.Equal(t, "2.0", JSONRPCVersion)

	// Test that error codes are within expected ranges
	assert.Equal(t, -32700, ParseError)
	assert.Equal(t, -32600, InvalidRequest)
	assert.Equal(t, -32601, MethodNotFound)
	assert.Equal(t, -32602, InvalidParams)
	assert.Equal(t, -32603, InternalError)

	// Custom error codes should be in the -32000 to -32099 range
	assert.True(t, ResourceNotFound >= -32099 && ResourceNotFound <= -32000)
	assert.True(t, UnauthorizedAccess >= -32099 && UnauthorizedAccess <= -32000)
	assert.True(t, ValidationError >= -32099 && ValidationError <= -32000)
	assert.True(t, ServiceUnavailable >= -32099 && ServiceUnavailable <= -32000)
	assert.True(t, RateLimitExceeded >= -32099 && RateLimitExceeded <= -32000)
}

// Helper methods for testing
func (r JSONRPCRequest) MarshalJSON() ([]byte, error) {
	type Alias JSONRPCRequest
	return json.Marshal((Alias)(r))
}

func (n JSONRPCNotification) MarshalJSON() ([]byte, error) {
	type Alias JSONRPCNotification
	return json.Marshal((Alias)(n))
}

func (r JSONRPCResponse) MarshalJSON() ([]byte, error) {
	type Alias JSONRPCResponse
	return json.Marshal((Alias)(r))
}

func (r *JSONRPCResponse) UnmarshalJSON(data []byte) error {
	type Alias JSONRPCResponse
	return json.Unmarshal(data, (*Alias)(r))
}
