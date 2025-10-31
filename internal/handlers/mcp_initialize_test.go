package handlers

import (
	"context"
	"testing"

	"product-requirements-management/internal/models"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockToolProvider implements mcp.ToolProvider for testing
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

// MockPromptProvider implements mcp.PromptProvider for testing
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

// MockPromptService implements mcp.PromptServiceInterface for testing
type MockPromptService struct {
	mock.Mock
}

func (m *MockPromptService) GetActive(ctx context.Context) (*models.Prompt, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}

func TestInitializeHandler_HandleInitializeFromParams_Success(t *testing.T) {
	// Setup mocks
	mockToolProvider := &MockToolProvider{}
	mockPromptProvider := &MockPromptProvider{}
	mockPromptService := &MockPromptService{}
	logger := logrus.New()

	// Configure mock expectations
	mockToolProvider.On("HasTools", mock.Anything).Return(true)
	mockToolProvider.On("SupportsListChanged", mock.Anything).Return(true)
	mockPromptProvider.On("HasPrompts", mock.Anything).Return(true)
	mockPromptProvider.On("SupportsListChanged", mock.Anything).Return(true)
	mockPromptService.On("GetActive", mock.Anything).Return(nil, nil) // No active prompt

	// Create handler
	handler := NewInitializeHandler(mockToolProvider, mockPromptProvider, mockPromptService, logger)

	// Test parameters
	params := map[string]interface{}{
		"protocolVersion": "2025-03-26",
		"clientInfo": map[string]interface{}{
			"name":    "test-client",
			"version": "1.0.0",
		},
		"capabilities": map[string]interface{}{
			"elicitation": map[string]interface{}{},
		},
	}

	// Execute
	result, err := handler.HandleInitializeFromParams(context.Background(), params)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)

	initResult, ok := result.(InitializeResult)
	assert.True(t, ok)

	// Check protocol version
	assert.Equal(t, "2025-03-26", initResult.ProtocolVersion)

	// Check server info
	assert.Equal(t, "spexus mcp", initResult.ServerInfo.Name)
	assert.Equal(t, "MCP server for requirements management system", initResult.ServerInfo.Title)
	assert.Equal(t, "1.0.0", initResult.ServerInfo.Version)

	// Check capabilities
	assert.NotNil(t, initResult.Capabilities)

	// Check instructions (should be empty since no active prompt)
	assert.Equal(t, "", initResult.Instructions)

	// Verify mock expectations
	mockToolProvider.AssertExpectations(t)
	mockPromptProvider.AssertExpectations(t)
	mockPromptService.AssertExpectations(t)
}

func TestInitializeHandler_HandleInitializeFromParams_InvalidProtocolVersion(t *testing.T) {
	// Setup mocks
	mockToolProvider := &MockToolProvider{}
	mockPromptProvider := &MockPromptProvider{}
	mockPromptService := &MockPromptService{}
	logger := logrus.New()

	// Create handler
	handler := NewInitializeHandler(mockToolProvider, mockPromptProvider, mockPromptService, logger)

	// Test parameters with invalid protocol version
	params := map[string]interface{}{
		"protocolVersion": "2024-01-01", // Invalid version
		"clientInfo": map[string]interface{}{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}

	// Execute
	result, err := handler.HandleInitializeFromParams(context.Background(), params)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, result)

	// Check error type
	initErr, ok := err.(*InitializeError)
	assert.True(t, ok)
	assert.Equal(t, InvalidProtocolVersion, initErr.Code)
}

func TestInitializeHandler_HandleInitializeFromParams_MissingClientInfo(t *testing.T) {
	// Setup mocks
	mockToolProvider := &MockToolProvider{}
	mockPromptProvider := &MockPromptProvider{}
	mockPromptService := &MockPromptService{}
	logger := logrus.New()

	// Create handler
	handler := NewInitializeHandler(mockToolProvider, mockPromptProvider, mockPromptService, logger)

	// Test parameters without client info
	params := map[string]interface{}{
		"protocolVersion": "2025-03-26",
		// Missing clientInfo
	}

	// Execute
	result, err := handler.HandleInitializeFromParams(context.Background(), params)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, result)

	// Check error type
	initErr, ok := err.(*InitializeError)
	assert.True(t, ok)
	assert.Equal(t, MalformedRequest, initErr.Code)
}

func TestValidateInitializeRequest_Success(t *testing.T) {
	requestJSON := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "initialize",
		"params": {
			"protocolVersion": "2025-03-26",
			"clientInfo": {
				"name": "test-client",
				"version": "1.0.0"
			},
			"capabilities": {
				"elicitation": {}
			}
		}
	}`

	req, err := ValidateInitializeRequest([]byte(requestJSON))

	assert.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, "2.0", req.JSONRPC)
	assert.Equal(t, "initialize", req.Method)
	assert.Equal(t, "2025-03-26", req.Params.ProtocolVersion)
	assert.Equal(t, "test-client", req.Params.ClientInfo.Name)
	assert.Equal(t, "1.0.0", req.Params.ClientInfo.Version)
	assert.NotNil(t, req.Params.Capabilities.Elicitation)
}

func TestValidateInitializeRequest_InvalidJSON(t *testing.T) {
	requestJSON := `{invalid json}`

	req, err := ValidateInitializeRequest([]byte(requestJSON))

	assert.Error(t, err)
	assert.Nil(t, req)
}

func TestValidateInitializeRequest_WrongMethod(t *testing.T) {
	requestJSON := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "ping",
		"params": {}
	}`

	req, err := ValidateInitializeRequest([]byte(requestJSON))

	assert.Error(t, err)
	assert.Nil(t, req)
}

func TestInitializeHandler_HandleInitialize_Success(t *testing.T) {
	// Setup mocks
	mockToolProvider := &MockToolProvider{}
	mockPromptProvider := &MockPromptProvider{}
	mockPromptService := &MockPromptService{}
	logger := logrus.New()

	// Configure mock expectations
	mockToolProvider.On("HasTools", mock.Anything).Return(true)
	mockToolProvider.On("SupportsListChanged", mock.Anything).Return(true)
	mockPromptProvider.On("HasPrompts", mock.Anything).Return(true)
	mockPromptProvider.On("SupportsListChanged", mock.Anything).Return(true)
	mockPromptService.On("GetActive", mock.Anything).Return(nil, nil) // No active prompt

	// Create handler
	handler := NewInitializeHandler(mockToolProvider, mockPromptProvider, mockPromptService, logger)

	// Create initialize request
	request := &InitializeRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: InitializeParams{
			ProtocolVersion: "2025-03-26",
			ClientInfo: ClientInfo{
				Name:    "test-client",
				Version: "1.0.0",
			},
			Capabilities: ClientCapabilities{
				Elicitation: &ElicitationCapability{},
			},
		},
	}

	// Execute
	response, err := handler.HandleInitialize(context.Background(), request)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, response)

	// Check JSON-RPC response structure
	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, 1, response.ID.GetValue())
	assert.Nil(t, response.Error)
	assert.NotNil(t, response.Result)

	// Check initialize result
	initResult, ok := response.Result.(InitializeResult)
	assert.True(t, ok)

	// Check protocol version
	assert.Equal(t, "2025-03-26", initResult.ProtocolVersion)

	// Check server info
	assert.Equal(t, "spexus mcp", initResult.ServerInfo.Name)
	assert.Equal(t, "MCP server for requirements management system", initResult.ServerInfo.Title)
	assert.Equal(t, "1.0.0", initResult.ServerInfo.Version)

	// Check capabilities
	assert.NotNil(t, initResult.Capabilities)

	// Check instructions (should be empty since no active prompt)
	assert.Equal(t, "", initResult.Instructions)

	// Verify mock expectations
	mockToolProvider.AssertExpectations(t)
	mockPromptProvider.AssertExpectations(t)
	mockPromptService.AssertExpectations(t)
}

func TestInitializeHandler_HandleInitialize_InvalidProtocolVersion(t *testing.T) {
	// Setup mocks
	mockToolProvider := &MockToolProvider{}
	mockPromptProvider := &MockPromptProvider{}
	mockPromptService := &MockPromptService{}
	logger := logrus.New()

	// Create handler
	handler := NewInitializeHandler(mockToolProvider, mockPromptProvider, mockPromptService, logger)

	// Create initialize request with invalid protocol version
	request := &InitializeRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: InitializeParams{
			ProtocolVersion: "2024-01-01", // Invalid version
			ClientInfo: ClientInfo{
				Name:    "test-client",
				Version: "1.0.0",
			},
		},
	}

	// Execute
	response, err := handler.HandleInitialize(context.Background(), request)

	// Verify
	assert.NoError(t, err) // No error returned, but error response created
	assert.NotNil(t, response)

	// Check JSON-RPC error response structure
	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, 1, response.ID.GetValue())
	assert.Nil(t, response.Result)
	assert.NotNil(t, response.Error)

	// Check error details
	assert.Contains(t, response.Error.Message, "Unsupported protocol version")
}

func TestProcessInitializeRequest_Success(t *testing.T) {
	// Setup mocks
	mockToolProvider := &MockToolProvider{}
	mockPromptProvider := &MockPromptProvider{}
	mockPromptService := &MockPromptService{}
	logger := logrus.New()

	// Configure mock expectations
	mockToolProvider.On("HasTools", mock.Anything).Return(true)
	mockToolProvider.On("SupportsListChanged", mock.Anything).Return(true)
	mockPromptProvider.On("HasPrompts", mock.Anything).Return(true)
	mockPromptProvider.On("SupportsListChanged", mock.Anything).Return(true)
	mockPromptService.On("GetActive", mock.Anything).Return(nil, nil) // No active prompt

	// Create handler
	handler := NewInitializeHandler(mockToolProvider, mockPromptProvider, mockPromptService, logger)

	requestJSON := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "initialize",
		"params": {
			"protocolVersion": "2025-03-26",
			"clientInfo": {
				"name": "test-client",
				"version": "1.0.0"
			},
			"capabilities": {
				"elicitation": {}
			}
		}
	}`

	// Execute
	response, err := handler.ProcessInitializeRequest(context.Background(), []byte(requestJSON))

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, response)

	// Check JSON-RPC response structure
	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, 1, response.ID.GetValue())
	assert.Nil(t, response.Error)
	assert.NotNil(t, response.Result)

	// Verify mock expectations
	mockToolProvider.AssertExpectations(t)
	mockPromptProvider.AssertExpectations(t)
	mockPromptService.AssertExpectations(t)
}

func TestProcessInitializeRequest_InvalidJSON(t *testing.T) {
	// Setup mocks
	mockToolProvider := &MockToolProvider{}
	mockPromptProvider := &MockPromptProvider{}
	mockPromptService := &MockPromptService{}
	logger := logrus.New()

	// Create handler
	handler := NewInitializeHandler(mockToolProvider, mockPromptProvider, mockPromptService, logger)

	requestJSON := `{invalid json}`

	// Execute
	response, err := handler.ProcessInitializeRequest(context.Background(), []byte(requestJSON))

	// Verify
	assert.NoError(t, err) // No error returned, but error response created
	assert.NotNil(t, response)

	// Check JSON-RPC error response structure
	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Nil(t, response.ID) // No valid ID for malformed JSON
	assert.Nil(t, response.Result)
	assert.NotNil(t, response.Error)

	// Check error details
	assert.Contains(t, response.Error.Message, "Invalid JSON format")
}

func TestValidateProtocolVersion_InvalidFormat(t *testing.T) {
	handler := &InitializeHandler{}

	testCases := []struct {
		name    string
		version string
	}{
		{"empty", ""},
		{"too short", "2025-01"},
		{"too long", "2025-01-011"},
		{"wrong format", "2025/01/01"},
		{"non-numeric year", "abcd-01-01"},
		{"non-numeric month", "2025-ab-01"},
		{"non-numeric day", "2025-01-ab"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := handler.validateProtocolVersion(tc.version)
			assert.Error(t, err)

			initErr, ok := err.(*InitializeError)
			assert.True(t, ok)
			assert.Equal(t, InvalidProtocolVersion, initErr.Code)
		})
	}
}

// Tests for Requirement 3: System instructions through initialize method

func TestInitializeHandler_WithActiveSystemPrompt(t *testing.T) {
	// Setup mocks
	mockToolProvider := &MockToolProvider{}
	mockPromptProvider := &MockPromptProvider{}
	mockPromptService := &MockPromptService{}
	logger := logrus.New()

	// Configure mock expectations
	mockToolProvider.On("HasTools", mock.Anything).Return(true)
	mockToolProvider.On("SupportsListChanged", mock.Anything).Return(true)
	mockPromptProvider.On("HasPrompts", mock.Anything).Return(true)
	mockPromptProvider.On("SupportsListChanged", mock.Anything).Return(true)

	// Mock active prompt with comprehensive system instructions
	activePrompt := &models.Prompt{
		Name:    "system-prompt",
		Title:   "System Instructions",
		Content: "You are an AI assistant for the spexus requirements management system. Available tools: create_epic, search_requirements. Resource access: Use spexus:// URIs for entities.",
	}
	mockPromptService.On("GetActive", mock.Anything).Return(activePrompt, nil)

	// Create handler
	handler := NewInitializeHandler(mockToolProvider, mockPromptProvider, mockPromptService, logger)

	// Test parameters
	params := map[string]interface{}{
		"protocolVersion": "2025-03-26",
		"clientInfo": map[string]interface{}{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}

	// Execute
	result, err := handler.HandleInitializeFromParams(context.Background(), params)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)

	initResult, ok := result.(InitializeResult)
	assert.True(t, ok)

	// Check that instructions field contains the active prompt content
	assert.Equal(t, activePrompt.Content, initResult.Instructions)
	assert.Contains(t, initResult.Instructions, "spexus requirements management system")
	assert.Contains(t, initResult.Instructions, "Available tools")
	assert.Contains(t, initResult.Instructions, "Resource access")

	// Verify mock expectations
	mockToolProvider.AssertExpectations(t)
	mockPromptProvider.AssertExpectations(t)
	mockPromptService.AssertExpectations(t)
}

func TestInitializeHandler_InstructionsContentStructure(t *testing.T) {
	// Setup mocks
	mockToolProvider := &MockToolProvider{}
	mockPromptProvider := &MockPromptProvider{}
	mockPromptService := &MockPromptService{}
	logger := logrus.New()

	// Configure mock expectations
	mockToolProvider.On("HasTools", mock.Anything).Return(true)
	mockToolProvider.On("SupportsListChanged", mock.Anything).Return(true)
	mockPromptProvider.On("HasPrompts", mock.Anything).Return(true)
	mockPromptProvider.On("SupportsListChanged", mock.Anything).Return(true)

	// Mock active prompt with structured content including tool guidance and resource patterns
	structuredPrompt := &models.Prompt{
		Name:  "structured-prompt",
		Title: "Structured System Instructions",
		Content: `You are an AI assistant for the spexus requirements management system.

AVAILABLE TOOLS:
- create_epic: Create new epics in the system
- search_requirements: Search through requirements
- update_user_story: Modify existing user stories

RESOURCE ACCESS PATTERNS:
- Use spexus://epic/{id} for epic resources
- Use spexus://requirement/{id} for requirement resources
- Subscribe to resource changes for real-time updates

USAGE GUIDANCE:
- Always validate input before creating entities
- Use search before creating duplicates
- Follow hierarchical structure: Epic -> User Story -> Requirement`,
	}
	mockPromptService.On("GetActive", mock.Anything).Return(structuredPrompt, nil)

	// Create handler
	handler := NewInitializeHandler(mockToolProvider, mockPromptProvider, mockPromptService, logger)

	// Test parameters
	params := map[string]interface{}{
		"protocolVersion": "2025-03-26",
		"clientInfo": map[string]interface{}{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}

	// Execute
	result, err := handler.HandleInitializeFromParams(context.Background(), params)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)

	initResult, ok := result.(InitializeResult)
	assert.True(t, ok)

	// Verify instructions field contains comprehensive system prompt content
	instructions := initResult.Instructions
	assert.Equal(t, structuredPrompt.Content, instructions)

	// Verify guidance on available tools and their usage
	assert.Contains(t, instructions, "AVAILABLE TOOLS")
	assert.Contains(t, instructions, "create_epic")
	assert.Contains(t, instructions, "search_requirements")
	assert.Contains(t, instructions, "update_user_story")

	// Verify information about resource access patterns
	assert.Contains(t, instructions, "RESOURCE ACCESS PATTERNS")
	assert.Contains(t, instructions, "spexus://epic/{id}")
	assert.Contains(t, instructions, "spexus://requirement/{id}")
	assert.Contains(t, instructions, "Subscribe to resource changes")

	// Verify usage guidance
	assert.Contains(t, instructions, "USAGE GUIDANCE")
	assert.Contains(t, instructions, "validate input")
	assert.Contains(t, instructions, "hierarchical structure")

	// Verify mock expectations
	mockToolProvider.AssertExpectations(t)
	mockPromptProvider.AssertExpectations(t)
	mockPromptService.AssertExpectations(t)
}

func TestInitializeHandler_SystemPromptServiceError(t *testing.T) {
	// Setup mocks
	mockToolProvider := &MockToolProvider{}
	mockPromptProvider := &MockPromptProvider{}
	mockPromptService := &MockPromptService{}
	logger := logrus.New()

	// Configure mock expectations
	mockToolProvider.On("HasTools", mock.Anything).Return(true)
	mockToolProvider.On("SupportsListChanged", mock.Anything).Return(true)
	mockPromptProvider.On("HasPrompts", mock.Anything).Return(true)
	mockPromptProvider.On("SupportsListChanged", mock.Anything).Return(true)

	// Mock system prompt service error
	mockPromptService.On("GetActive", mock.Anything).Return(nil, assert.AnError)

	// Create handler
	handler := NewInitializeHandler(mockToolProvider, mockPromptProvider, mockPromptService, logger)

	// Test parameters
	params := map[string]interface{}{
		"protocolVersion": "2025-03-26",
		"clientInfo": map[string]interface{}{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}

	// Execute
	result, err := handler.HandleInitializeFromParams(context.Background(), params)

	// Verify - should not fail, but use empty instructions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	initResult, ok := result.(InitializeResult)
	assert.True(t, ok)

	// Check that instructions field is empty when system prompt service fails
	assert.Equal(t, "", initResult.Instructions)

	// Verify other fields are still correct
	assert.Equal(t, "2025-03-26", initResult.ProtocolVersion)
	assert.Equal(t, "spexus mcp", initResult.ServerInfo.Name)

	// Verify mock expectations
	mockToolProvider.AssertExpectations(t)
	mockPromptProvider.AssertExpectations(t)
	mockPromptService.AssertExpectations(t)
}

func TestInitializeHandler_InstructionsUpdateOnCapabilityChange(t *testing.T) {
	// Test that demonstrates instructions can change based on different system prompts
	// This simulates the scenario where system capabilities change and instructions are updated

	// First scenario: Tools available
	t.Run("with_tools_available", func(t *testing.T) {
		// Setup mocks for first scenario
		mockToolProvider1 := &MockToolProvider{}
		mockPromptProvider1 := &MockPromptProvider{}
		mockPromptService1 := &MockPromptService{}
		logger := logrus.New()

		// Configure mock expectations (tools available)
		mockToolProvider1.On("HasTools", mock.Anything).Return(true)
		mockToolProvider1.On("SupportsListChanged", mock.Anything).Return(true)
		mockPromptProvider1.On("HasPrompts", mock.Anything).Return(true)
		mockPromptProvider1.On("SupportsListChanged", mock.Anything).Return(true)

		// Mock prompt with tool-specific instructions
		toolsPrompt := &models.Prompt{
			Name:    "tools-available-prompt",
			Title:   "Instructions with Tools",
			Content: "System has tools available: create_epic, search_requirements. Use these tools to manage requirements.",
		}
		mockPromptService1.On("GetActive", mock.Anything).Return(toolsPrompt, nil)

		// Create handler for first scenario
		handler1 := NewInitializeHandler(mockToolProvider1, mockPromptProvider1, mockPromptService1, logger)

		params := map[string]interface{}{
			"protocolVersion": "2025-03-26",
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		}

		result1, err1 := handler1.HandleInitializeFromParams(context.Background(), params)
		assert.NoError(t, err1)
		initResult1, ok := result1.(InitializeResult)
		assert.True(t, ok)
		assert.Equal(t, toolsPrompt.Content, initResult1.Instructions)
		assert.Contains(t, initResult1.Instructions, "tools available")

		// Verify mock expectations
		mockToolProvider1.AssertExpectations(t)
		mockPromptProvider1.AssertExpectations(t)
		mockPromptService1.AssertExpectations(t)
	})

	// Second scenario: Limited capabilities
	t.Run("with_limited_capabilities", func(t *testing.T) {
		// Setup mocks for second scenario
		mockToolProvider2 := &MockToolProvider{}
		mockPromptProvider2 := &MockPromptProvider{}
		mockPromptService2 := &MockPromptService{}
		logger := logrus.New()

		// Configure mock expectations (limited capabilities)
		// When HasTools returns false, SupportsListChanged is not called
		mockToolProvider2.On("HasTools", mock.Anything).Return(false)
		mockPromptProvider2.On("HasPrompts", mock.Anything).Return(true)
		mockPromptProvider2.On("SupportsListChanged", mock.Anything).Return(true)

		// Mock prompt with limited capability instructions
		limitedPrompt := &models.Prompt{
			Name:    "limited-capabilities-prompt",
			Title:   "Instructions with Limited Capabilities",
			Content: "System has limited capabilities - no tools available. You can only access resources and provide information.",
		}
		mockPromptService2.On("GetActive", mock.Anything).Return(limitedPrompt, nil)

		// Create handler for second scenario
		handler2 := NewInitializeHandler(mockToolProvider2, mockPromptProvider2, mockPromptService2, logger)

		params := map[string]interface{}{
			"protocolVersion": "2025-03-26",
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		}

		result2, err2 := handler2.HandleInitializeFromParams(context.Background(), params)
		assert.NoError(t, err2)
		initResult2, ok := result2.(InitializeResult)
		assert.True(t, ok)
		assert.Equal(t, limitedPrompt.Content, initResult2.Instructions)
		assert.Contains(t, initResult2.Instructions, "limited capabilities")
		assert.Contains(t, initResult2.Instructions, "no tools available")

		// Verify mock expectations
		mockToolProvider2.AssertExpectations(t)
		mockPromptProvider2.AssertExpectations(t)
		mockPromptService2.AssertExpectations(t)
	})
}

func TestInitializeHandler_EmptyInstructionsWhenNoActivePrompt(t *testing.T) {
	// Setup mocks
	mockToolProvider := &MockToolProvider{}
	mockPromptProvider := &MockPromptProvider{}
	mockPromptService := &MockPromptService{}
	logger := logrus.New()

	// Configure mock expectations
	mockToolProvider.On("HasTools", mock.Anything).Return(true)
	mockToolProvider.On("SupportsListChanged", mock.Anything).Return(true)
	mockPromptProvider.On("HasPrompts", mock.Anything).Return(true)
	mockPromptProvider.On("SupportsListChanged", mock.Anything).Return(true)

	// Mock no active prompt
	mockPromptService.On("GetActive", mock.Anything).Return(nil, nil)

	// Create handler
	handler := NewInitializeHandler(mockToolProvider, mockPromptProvider, mockPromptService, logger)

	// Test parameters
	params := map[string]interface{}{
		"protocolVersion": "2025-03-26",
		"clientInfo": map[string]interface{}{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}

	// Execute
	result, err := handler.HandleInitializeFromParams(context.Background(), params)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)

	initResult, ok := result.(InitializeResult)
	assert.True(t, ok)

	// Check that instructions field is empty string when no active prompt
	assert.Equal(t, "", initResult.Instructions)

	// Verify other fields are still correct
	assert.Equal(t, "2025-03-26", initResult.ProtocolVersion)
	assert.NotNil(t, initResult.Capabilities)

	// Verify mock expectations
	mockToolProvider.AssertExpectations(t)
	mockPromptProvider.AssertExpectations(t)
	mockPromptService.AssertExpectations(t)
}
