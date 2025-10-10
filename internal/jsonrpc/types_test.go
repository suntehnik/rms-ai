package jsonrpc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONRPCRequest_IsNotification(t *testing.T) {
	tests := []struct {
		name     string
		request  JSONRPCRequest
		expected bool
	}{
		{
			name: "request with ID is not notification",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      &JSONRPCID{Value: 1},
				Method:  "test",
			},
			expected: false,
		},
		{
			name: "request without ID is notification",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      nil,
				Method:  "test",
			},
			expected: true,
		},
		{
			name: "request with null ID is notification",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      &JSONRPCID{Value: nil},
				Method:  "test",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.request.IsNotification())
		})
	}
}

func TestNewSuccessResponse(t *testing.T) {
	result := map[string]interface{}{"status": "ok"}
	response := NewSuccessResponse(1, result)

	assert.Equal(t, JSONRPCVersion, response.JSONRPC)
	assert.Equal(t, 1, response.ID.GetValue())
	assert.Equal(t, result, response.Result)
	assert.Nil(t, response.Error)
}

func TestNewErrorResponse(t *testing.T) {
	err := NewJSONRPCError(InvalidRequest, "Invalid request", "test data")
	response := NewErrorResponse(1, err)

	assert.Equal(t, JSONRPCVersion, response.JSONRPC)
	assert.Equal(t, 1, response.ID.GetValue())
	assert.Nil(t, response.Result)
	assert.Equal(t, err, response.Error)
}

func TestParseRequest(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		expectError bool
		errorCode   int
	}{
		{
			name:        "valid request",
			data:        `{"jsonrpc":"2.0","id":1,"method":"test","params":{"key":"value"}}`,
			expectError: false,
		},
		{
			name:        "valid request without params",
			data:        `{"jsonrpc":"2.0","id":1,"method":"test"}`,
			expectError: false,
		},
		{
			name:        "invalid JSON",
			data:        `{"jsonrpc":"2.0","id":1,"method":"test"`,
			expectError: true,
			errorCode:   ParseError,
		},
		{
			name:        "missing jsonrpc field",
			data:        `{"id":1,"method":"test"}`,
			expectError: true,
			errorCode:   InvalidRequest,
		},
		{
			name:        "wrong jsonrpc version",
			data:        `{"jsonrpc":"1.0","id":1,"method":"test"}`,
			expectError: true,
			errorCode:   InvalidRequest,
		},
		{
			name:        "missing method field",
			data:        `{"jsonrpc":"2.0","id":1}`,
			expectError: true,
			errorCode:   InvalidRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := ParseRequest([]byte(tt.data))

			if tt.expectError {
				require.Error(t, err)
				jsonrpcErr, ok := err.(*JSONRPCError)
				require.True(t, ok)
				assert.Equal(t, tt.errorCode, jsonrpcErr.Code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, JSONRPCVersion, request.JSONRPC)
				assert.Equal(t, "test", request.Method)
			}
		})
	}
}

func TestParseNotification(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		expectError bool
		errorCode   int
	}{
		{
			name:        "valid notification",
			data:        `{"jsonrpc":"2.0","method":"test","params":{"key":"value"}}`,
			expectError: false,
		},
		{
			name:        "valid notification without params",
			data:        `{"jsonrpc":"2.0","method":"test"}`,
			expectError: false,
		},
		{
			name:        "invalid JSON",
			data:        `{"jsonrpc":"2.0","method":"test"`,
			expectError: true,
			errorCode:   ParseError,
		},
		{
			name:        "missing jsonrpc field",
			data:        `{"method":"test"}`,
			expectError: true,
			errorCode:   InvalidRequest,
		},
		{
			name:        "wrong jsonrpc version",
			data:        `{"jsonrpc":"1.0","method":"test"}`,
			expectError: true,
			errorCode:   InvalidRequest,
		},
		{
			name:        "missing method field",
			data:        `{"jsonrpc":"2.0"}`,
			expectError: true,
			errorCode:   InvalidRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notification, err := ParseNotification([]byte(tt.data))

			if tt.expectError {
				require.Error(t, err)
				jsonrpcErr, ok := err.(*JSONRPCError)
				require.True(t, ok)
				assert.Equal(t, tt.errorCode, jsonrpcErr.Code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, JSONRPCVersion, notification.JSONRPC)
				assert.Equal(t, "test", notification.Method)
			}
		})
	}
}

func TestJSONRPCError_Error(t *testing.T) {
	err := NewJSONRPCError(InvalidRequest, "Invalid request", "test data")
	expected := "JSON-RPC error -32600: Invalid request"
	assert.Equal(t, expected, err.Error())
}

func TestJSONSerialization(t *testing.T) {
	t.Run("serialize request", func(t *testing.T) {
		request := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      &JSONRPCID{Value: 1},
			Method:  "test",
			Params:  map[string]interface{}{"key": "value"},
		}

		data, err := json.Marshal(request)
		require.NoError(t, err)

		var parsed JSONRPCRequest
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		assert.Equal(t, request.JSONRPC, parsed.JSONRPC)
		assert.Equal(t, 1, parsed.ID.GetValue())
		assert.Equal(t, request.Method, parsed.Method)
	})

	t.Run("serialize response", func(t *testing.T) {
		response := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      &JSONRPCID{Value: 1},
			Result:  map[string]interface{}{"status": "ok"},
		}

		data, err := json.Marshal(response)
		require.NoError(t, err)

		var parsed JSONRPCResponse
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		assert.Equal(t, response.JSONRPC, parsed.JSONRPC)
		assert.Equal(t, 1, parsed.ID.GetValue())
		assert.Equal(t, response.Result, parsed.Result)
		assert.Nil(t, parsed.Error)
	})

	t.Run("serialize error response", func(t *testing.T) {
		response := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      &JSONRPCID{Value: 1},
			Error: &JSONRPCError{
				Code:    InvalidRequest,
				Message: "Invalid request",
				Data:    "test data",
			},
		}

		data, err := json.Marshal(response)
		require.NoError(t, err)

		var parsed JSONRPCResponse
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		assert.Equal(t, response.JSONRPC, parsed.JSONRPC)
		assert.Equal(t, 1, parsed.ID.GetValue())
		assert.Nil(t, parsed.Result)
		require.NotNil(t, parsed.Error)
		assert.Equal(t, InvalidRequest, parsed.Error.Code)
		assert.Equal(t, "Invalid request", parsed.Error.Message)
	})
}
