package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockToolProvider for testing
type MockToolProvider struct {
	mock.Mock
}

func (m *MockToolProvider) HasTools(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func (m *MockToolProvider) SupportsListChanged(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

// MockPromptProvider for testing
type MockPromptProvider struct {
	mock.Mock
}

func (m *MockPromptProvider) HasPrompts(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func (m *MockPromptProvider) SupportsListChanged(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func TestCapabilitiesManager_GenerateCapabilities_WithProviders(t *testing.T) {
	// Setup mocks
	mockToolProvider := &MockToolProvider{}
	mockPromptProvider := &MockPromptProvider{}

	// Configure mock expectations
	mockToolProvider.On("HasTools", mock.Anything).Return(true)
	mockToolProvider.On("SupportsListChanged", mock.Anything).Return(true)
	mockPromptProvider.On("HasPrompts", mock.Anything).Return(true)
	mockPromptProvider.On("SupportsListChanged", mock.Anything).Return(true)

	// Create capabilities manager
	cm := NewCapabilitiesManager(mockToolProvider, mockPromptProvider)

	// Generate capabilities
	capabilities, err := cm.GenerateCapabilities(context.Background())

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, capabilities)

	// Check logging capability (always present)
	// temporary disabled
	// assert.NotNil(t, capabilities.Logging)

	// Check prompts capability
	assert.True(t, capabilities.Prompts.ListChanged)

	// Check resources capability
	assert.True(t, capabilities.Resources.ListChanged)
	assert.True(t, capabilities.Resources.Subscribe)

	// Check tools capability
	assert.True(t, capabilities.Tools.ListChanged)

	// Verify mock expectations
	mockToolProvider.AssertExpectations(t)
	mockPromptProvider.AssertExpectations(t)
}

func TestCapabilitiesManager_GenerateCapabilities_WithoutProviders(t *testing.T) {
	// Create capabilities manager with nil providers
	cm := NewCapabilitiesManager(nil, nil)

	// Generate capabilities
	capabilities, err := cm.GenerateCapabilities(context.Background())

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, capabilities)

	// Check that default capabilities are provided even without providers
	// logging capability is temporary disabled
	// assert.NotNil(t, capabilities.Logging)
	assert.True(t, capabilities.Prompts.ListChanged)
	assert.True(t, capabilities.Resources.ListChanged)
	assert.True(t, capabilities.Resources.Subscribe)
	assert.True(t, capabilities.Tools.ListChanged)
}

func TestCapabilitiesManager_HasCapability(t *testing.T) {
	// Setup mocks
	mockToolProvider := &MockToolProvider{}
	mockPromptProvider := &MockPromptProvider{}

	// Configure mock expectations
	mockToolProvider.On("HasTools", mock.Anything).Return(true)
	mockPromptProvider.On("HasPrompts", mock.Anything).Return(true)

	// Create capabilities manager
	cm := NewCapabilitiesManager(mockToolProvider, mockPromptProvider)

	// Test capability checks
	assert.False(t, cm.HasCapability(context.Background(), "logging")) // Logging capability is not supported
	assert.True(t, cm.HasCapability(context.Background(), "prompts"))
	assert.True(t, cm.HasCapability(context.Background(), "resources"))
	assert.True(t, cm.HasCapability(context.Background(), "tools"))
	assert.False(t, cm.HasCapability(context.Background(), "unknown"))

	// Verify mock expectations
	mockToolProvider.AssertExpectations(t)
	mockPromptProvider.AssertExpectations(t)
}

func TestCapabilitiesManager_UpdateCapabilities(t *testing.T) {
	// Setup mocks
	mockToolProvider := &MockToolProvider{}
	mockPromptProvider := &MockPromptProvider{}

	// Configure mock expectations
	mockToolProvider.On("HasTools", mock.Anything).Return(true)
	mockToolProvider.On("SupportsListChanged", mock.Anything).Return(true)
	mockPromptProvider.On("HasPrompts", mock.Anything).Return(true)
	mockPromptProvider.On("SupportsListChanged", mock.Anything).Return(true)

	// Create capabilities manager
	cm := NewCapabilitiesManager(mockToolProvider, mockPromptProvider)

	// Update capabilities
	capabilities, err := cm.UpdateCapabilities(context.Background())

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, capabilities)

	// Should be the same as GenerateCapabilities for now
	assert.True(t, capabilities.Prompts.ListChanged)
	assert.True(t, capabilities.Resources.ListChanged)
	assert.True(t, capabilities.Resources.Subscribe)
	assert.True(t, capabilities.Tools.ListChanged)

	// Verify mock expectations
	mockToolProvider.AssertExpectations(t)
	mockPromptProvider.AssertExpectations(t)
}
