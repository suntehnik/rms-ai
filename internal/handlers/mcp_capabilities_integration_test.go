package handlers

import (
	"context"
	"testing"

	"product-requirements-management/internal/mcp"
	"product-requirements-management/internal/service"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCapabilitiesIntegration_WithRealHandlers(t *testing.T) {
	// Create mock tool and prompt providers
	mockToolProvider := &MockToolProvider{}
	mockPromptProvider := &MockPromptProvider{}

	// Configure mock expectations
	mockToolProvider.On("HasTools", mock.Anything).Return(true)
	mockToolProvider.On("SupportsListChanged", mock.Anything).Return(true)
	mockPromptProvider.On("HasPrompts", mock.Anything).Return(true)
	mockPromptProvider.On("SupportsListChanged", mock.Anything).Return(true)

	// Create capabilities manager with mock providers
	cm := mcp.NewCapabilitiesManager(mockToolProvider, mockPromptProvider)

	// Generate capabilities
	capabilities, err := cm.GenerateCapabilities(context.Background())

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, capabilities)

	// Verify that capabilities accurately reflect actual server functionality

	// 1. Logging capability is temporary disabled
	// assert.NotNil(t, capabilities.Logging)

	// 2. Tools capability should reflect that we have tools available
	assert.True(t, capabilities.Tools.ListChanged, "Tools should support list changes")

	// 3. Prompts capability should reflect that we have prompts available
	assert.True(t, capabilities.Prompts.ListChanged, "Prompts should support list changes")

	// 4. Resources capability should be configured correctly
	assert.True(t, capabilities.Resources.ListChanged, "Resources should support list changes")
	assert.True(t, capabilities.Resources.Subscribe, "Resources should support subscription")

	// Verify mock expectations
	mockToolProvider.AssertExpectations(t)
	mockPromptProvider.AssertExpectations(t)
}

func TestCapabilitiesIntegration_ConsistencyCheck(t *testing.T) {
	// Create mock providers
	mockToolProvider := &MockToolProvider{}
	mockPromptProvider := &MockPromptProvider{}

	// Configure mock expectations for multiple calls
	mockToolProvider.On("HasTools", mock.Anything).Return(true)
	mockToolProvider.On("SupportsListChanged", mock.Anything).Return(true)
	mockPromptProvider.On("HasPrompts", mock.Anything).Return(true)
	mockPromptProvider.On("SupportsListChanged", mock.Anything).Return(true)

	cm := mcp.NewCapabilitiesManager(mockToolProvider, mockPromptProvider)

	// Generate capabilities multiple times - should be consistent
	capabilities1, err1 := cm.GenerateCapabilities(context.Background())
	capabilities2, err2 := cm.GenerateCapabilities(context.Background())

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, capabilities1.Tools.ListChanged, capabilities2.Tools.ListChanged)
	assert.Equal(t, capabilities1.Prompts.ListChanged, capabilities2.Prompts.ListChanged)
	assert.Equal(t, capabilities1.Resources.ListChanged, capabilities2.Resources.ListChanged)
	assert.Equal(t, capabilities1.Resources.Subscribe, capabilities2.Resources.Subscribe)

	// Verify that HasCapability method is consistent with generated capabilities
	assert.Equal(t, false, cm.HasCapability(context.Background(), "logging")) // Logging capability is not supported
	assert.Equal(t, true, cm.HasCapability(context.Background(), "tools"))
	assert.Equal(t, true, cm.HasCapability(context.Background(), "prompts"))
	assert.Equal(t, true, cm.HasCapability(context.Background(), "resources"))
	assert.Equal(t, false, cm.HasCapability(context.Background(), "unknown"))

	// Verify mock expectations
	mockToolProvider.AssertExpectations(t)
	mockPromptProvider.AssertExpectations(t)
}

func TestInitializeHandler_CapabilitiesIntegration(t *testing.T) {
	// Create mock providers
	mockToolProvider := &MockToolProvider{}
	mockPromptProvider := &MockPromptProvider{}
	mockPromptService := &MockPromptService{}

	// Configure mocks
	mockToolProvider.On("HasTools", mock.Anything).Return(true)
	mockToolProvider.On("SupportsListChanged", mock.Anything).Return(true)
	mockPromptProvider.On("HasPrompts", mock.Anything).Return(true)
	mockPromptProvider.On("SupportsListChanged", mock.Anything).Return(true)
	mockPromptService.On("GetActive", mock.Anything).Return(nil, service.ErrNotFound)

	// Create initialize handler with mock providers
	initHandler := NewInitializeHandler(mockToolProvider, mockPromptProvider, mockPromptService, logrus.New())

	// Test parameters
	params := map[string]interface{}{
		"protocolVersion": "2025-03-26",
		"clientInfo": map[string]interface{}{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}

	// Execute
	result, err := initHandler.HandleInitializeFromParams(context.Background(), params)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)

	initResult, ok := result.(InitializeResult)
	assert.True(t, ok)

	// Verify that capabilities are properly integrated and reflect actual functionality
	capabilities := initResult.Capabilities

	// Check that all required capabilities are present
	// assert.NotNil(t, capabilities.Logging)
	assert.True(t, capabilities.Tools.ListChanged, "Tools capability should indicate list change support")
	assert.True(t, capabilities.Prompts.ListChanged, "Prompts capability should indicate list change support")
	assert.True(t, capabilities.Resources.ListChanged, "Resources capability should indicate list change support")
	assert.True(t, capabilities.Resources.Subscribe, "Resources capability should indicate subscription support")

	// Verify mock expectations
	mockToolProvider.AssertExpectations(t)
	mockPromptProvider.AssertExpectations(t)
	mockPromptService.AssertExpectations(t)
}
