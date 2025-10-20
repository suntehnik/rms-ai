package handlers

import (
	"context"
	"errors"
	"testing"

	"product-requirements-management/internal/jsonrpc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestHandleResourcesList_ErrorHandling tests the comprehensive error handling
func TestHandleResourcesList_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockResourceService)
		setupContext   func() context.Context
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "timeout_error",
			setupMock: func(mockService *MockResourceService) {
				mockService.On("GetResourceList", mock.Anything).Return(nil, context.DeadlineExceeded)
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedError:  true,
			expectedErrMsg: "Operation timeout",
		},
		{
			name: "context_canceled",
			setupMock: func(mockService *MockResourceService) {
				mockService.On("GetResourceList", mock.Anything).Return(nil, context.Canceled)
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedError:  true,
			expectedErrMsg: "Internal error",
		},
		{
			name: "database_error",
			setupMock: func(mockService *MockResourceService) {
				mockService.On("GetResourceList", mock.Anything).Return(nil, errors.New("database connection failed"))
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedError:  true,
			expectedErrMsg: "Internal error",
		},
		{
			name: "authentication_error",
			setupMock: func(mockService *MockResourceService) {
				mockService.On("GetResourceList", mock.Anything).Return(nil, errors.New("unauthorized access"))
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedError:  true,
			expectedErrMsg: "Unauthorized access",
		},
		{
			name: "validation_error",
			setupMock: func(mockService *MockResourceService) {
				mockService.On("GetResourceList", mock.Anything).Return(nil, errors.New("validation failed"))
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedError:  true,
			expectedErrMsg: "Validation error",
		},
		{
			name: "generic_internal_error",
			setupMock: func(mockService *MockResourceService) {
				mockService.On("GetResourceList", mock.Anything).Return(nil, errors.New("unexpected error"))
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedError:  true,
			expectedErrMsg: "Internal error",
		},
		{
			name: "nil_resources_returned",
			setupMock: func(mockService *MockResourceService) {
				mockService.On("GetResourceList", mock.Anything).Return(nil, nil)
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedError:  true,
			expectedErrMsg: "Validation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock resource service
			mockResourceService := new(MockResourceService)
			tt.setupMock(mockResourceService)

			// Create handler
			handler := &MCPHandler{
				resourceService: mockResourceService,
				mcpLogger:       NewMCPLogger(),
				errorMapper:     jsonrpc.NewErrorMapper(),
			}

			// Setup context
			ctx := tt.setupContext()

			// Call the method
			result, err := handler.handleResourcesList(ctx, nil)

			// Verify expectations
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			// Verify mock was called
			mockResourceService.AssertExpectations(t)
		})
	}
}
