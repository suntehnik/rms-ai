package handlers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"product-requirements-management/internal/jsonrpc"
)

func TestResourceHandler_HandleResourcesRead_InvalidParams(t *testing.T) {
	handler := NewResourceHandler(nil, nil, nil, nil)

	tests := []struct {
		name   string
		params interface{}
	}{
		{
			name:   "invalid parameters format",
			params: "invalid",
		},
		{
			name:   "missing URI parameter",
			params: map[string]interface{}{"other": "value"},
		},
		{
			name:   "empty URI parameter",
			params: map[string]interface{}{"uri": ""},
		},
		{
			name:   "invalid URI format",
			params: map[string]interface{}{"uri": "invalid-uri"},
		},
		{
			name:   "unsupported URI scheme",
			params: map[string]interface{}{"uri": "unsupported://test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.HandleResourcesRead(context.Background(), tt.params)

			assert.Error(t, err)
			assert.Nil(t, result)

			// Verify it's a JSON-RPC error with InvalidParams code
			jsonrpcErr, ok := err.(*jsonrpc.JSONRPCError)
			assert.True(t, ok, "Error should be a JSON-RPC error")
			assert.Equal(t, jsonrpc.InvalidParams, jsonrpcErr.Code)
		})
	}
}

func TestResourceHandler_URIParser(t *testing.T) {
	handler := NewResourceHandler(nil, nil, nil, nil)

	tests := []struct {
		name        string
		uri         string
		expectError bool
	}{
		{
			name:        "valid epic URI",
			uri:         "epic://EP-001",
			expectError: false,
		},
		{
			name:        "valid user story URI",
			uri:         "user-story://US-001",
			expectError: false,
		},
		{
			name:        "valid requirement URI",
			uri:         "requirement://REQ-001",
			expectError: false,
		},
		{
			name:        "valid acceptance criteria URI",
			uri:         "acceptance-criteria://AC-001",
			expectError: false,
		},
		{
			name:        "epic URI with hierarchy sub-path",
			uri:         "epic://EP-001/hierarchy",
			expectError: false,
		},
		{
			name:        "user story URI with requirements sub-path",
			uri:         "user-story://US-001/requirements",
			expectError: false,
		},
		{
			name:        "requirement URI with relationships sub-path",
			uri:         "requirement://REQ-001/relationships",
			expectError: false,
		},
		{
			name:        "invalid reference ID format",
			uri:         "epic://INVALID-001",
			expectError: true,
		},
		{
			name:        "wrong prefix for scheme",
			uri:         "epic://US-001",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURI, err := handler.uriParser.Parse(tt.uri)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, parsedURI)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, parsedURI)
				assert.NotEmpty(t, parsedURI.Scheme)
				assert.NotEmpty(t, parsedURI.ReferenceID)
			}
		})
	}
}

func TestResourceHandler_NewResourceHandler(t *testing.T) {
	handler := NewResourceHandler(nil, nil, nil, nil)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.uriParser)
}
