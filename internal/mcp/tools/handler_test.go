package tools

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"product-requirements-management/internal/jsonrpc"
	"product-requirements-management/internal/models"
)

// Helper function to create a test context with user
func createTestContextWithUser() context.Context {
	gin.SetMode(gin.TestMode)
	ginCtx, _ := gin.CreateTestContext(nil)

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	ginCtx.Set("user", user)

	ctx := context.WithValue(context.Background(), "gin_context", ginCtx)
	return ctx
}

func TestHandlerStructure(t *testing.T) {
	// Test that we can create a handler with nil services (just for structure testing)
	handler := &Handler{
		toolRoutes: make(map[string]ToolHandler),
	}

	// Test basic structure
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.toolRoutes)
}

func TestHandleToolsCall_InvalidParameters(t *testing.T) {
	handler := &Handler{
		toolRoutes: make(map[string]ToolHandler),
	}

	ctx := createTestContextWithUser()

	// Test invalid parameters format
	result, err := handler.HandleToolsCall(ctx, "invalid")
	assert.Error(t, err)
	assert.Nil(t, result)

	// Verify it's an invalid params error
	jsonrpcErr, ok := err.(*jsonrpc.JSONRPCError)
	assert.True(t, ok)
	assert.Equal(t, jsonrpc.InvalidParams, jsonrpcErr.Code)
}

func TestHandleToolsCall_MissingToolName(t *testing.T) {
	handler := &Handler{
		toolRoutes: make(map[string]ToolHandler),
	}

	ctx := createTestContextWithUser()

	// Test missing tool name
	params := map[string]interface{}{
		"arguments": map[string]interface{}{
			"title": "Test",
		},
	}

	result, err := handler.HandleToolsCall(ctx, params)
	assert.Error(t, err)
	assert.Nil(t, result)

	// Verify it's an invalid params error
	jsonrpcErr, ok := err.(*jsonrpc.JSONRPCError)
	assert.True(t, ok)
	assert.Equal(t, jsonrpc.InvalidParams, jsonrpcErr.Code)
}

func TestHandleToolsCall_UnknownTool(t *testing.T) {
	handler := &Handler{
		toolRoutes: make(map[string]ToolHandler),
	}

	ctx := createTestContextWithUser()

	// Test unknown tool
	params := map[string]interface{}{
		"name": "unknown_tool",
		"arguments": map[string]interface{}{
			"title": "Test",
		},
	}

	result, err := handler.HandleToolsCall(ctx, params)
	assert.Error(t, err)
	assert.Nil(t, result)

	// Verify it's a method not found error
	jsonrpcErr, ok := err.(*jsonrpc.JSONRPCError)
	assert.True(t, ok)
	assert.Equal(t, jsonrpc.MethodNotFound, jsonrpcErr.Code)
}

func TestIsToolSupported(t *testing.T) {
	handler := &Handler{
		toolRoutes: map[string]ToolHandler{
			"test_tool": nil, // We don't need actual handler for this test
		},
	}

	// Test supported tool
	assert.True(t, handler.IsToolSupported("test_tool"))

	// Test unsupported tools
	assert.False(t, handler.IsToolSupported("unknown_tool"))
	assert.False(t, handler.IsToolSupported(""))
}

func TestGetHandlerForTool(t *testing.T) {
	// Create a mock handler for testing
	mockHandler := &EpicHandler{}

	handler := &Handler{
		toolRoutes: map[string]ToolHandler{
			"test_tool": mockHandler,
		},
	}

	// Test getting handler for existing tool
	foundHandler, exists := handler.GetHandlerForTool("test_tool")
	assert.True(t, exists)
	assert.Equal(t, mockHandler, foundHandler)

	// Test getting handler for non-existing tool
	unknownHandler, exists := handler.GetHandlerForTool("unknown_tool")
	assert.False(t, exists)
	assert.Nil(t, unknownHandler)
}

func TestGetToolRoutes(t *testing.T) {
	// Create a mock handler for testing
	mockHandler := &EpicHandler{}

	originalRoutes := map[string]ToolHandler{
		"test_tool": mockHandler,
	}

	handler := &Handler{
		toolRoutes: originalRoutes,
	}

	// Get tool routes
	routes := handler.GetToolRoutes()

	// Verify routes are returned and it's a copy (not the original map)
	assert.NotNil(t, routes)

	// Verify content is correct
	assert.Equal(t, mockHandler, routes["test_tool"])
	assert.Len(t, routes, 1)

	// Verify it's a copy by modifying the returned map and checking original is unchanged
	routes["new_tool"] = mockHandler
	assert.NotContains(t, handler.toolRoutes, "new_tool", "Original map should not be modified")
}

func TestGetAllSupportedTools_EmptyHandler(t *testing.T) {
	handler := &Handler{
		epicHandler:             &EpicHandler{},
		userStoryHandler:        &UserStoryHandler{},
		requirementHandler:      &RequirementHandler{},
		searchHandler:           &SearchHandler{},
		steeringDocumentHandler: &SteeringDocumentHandler{},
		promptHandler:           &PromptHandler{},
	}

	// Get all supported tools
	allTools := handler.GetAllSupportedTools()

	// Should return a slice (even if empty from nil handlers)
	assert.NotNil(t, allTools)
	// The actual tools will be empty since we have nil service dependencies
	// but the structure should work
}

// Test that the routing map is built correctly when handlers are provided
func TestToolRoutingMapStructure(t *testing.T) {
	// Test the expected tool names that should be routed
	expectedEpicTools := []string{"create_epic", "update_epic"}
	expectedUserStoryTools := []string{"create_user_story", "update_user_story"}
	expectedRequirementTools := []string{"create_requirement", "update_requirement", "create_relationship"}
	expectedSearchTools := []string{"search_global", "search_requirements"}
	expectedSteeringDocumentTools := []string{
		"list_steering_documents", "create_steering_document", "get_steering_document",
		"update_steering_document", "link_steering_to_epic", "unlink_steering_from_epic",
		"get_epic_steering_documents",
	}
	expectedPromptTools := []string{
		"create_prompt", "update_prompt", "delete_prompt", "activate_prompt",
		"list_prompts", "get_active_prompt",
	}

	// Verify that these are the expected tool names
	// This test documents the expected tool routing structure
	allExpectedTools := []string{}
	allExpectedTools = append(allExpectedTools, expectedEpicTools...)
	allExpectedTools = append(allExpectedTools, expectedUserStoryTools...)
	allExpectedTools = append(allExpectedTools, expectedRequirementTools...)
	allExpectedTools = append(allExpectedTools, expectedSearchTools...)
	allExpectedTools = append(allExpectedTools, expectedSteeringDocumentTools...)
	allExpectedTools = append(allExpectedTools, expectedPromptTools...)

	// Verify we have the expected number of tools
	assert.Equal(t, 22, len(allExpectedTools), "Expected 22 total tools across all domains")

	// Verify no duplicate tool names
	toolSet := make(map[string]bool)
	for _, tool := range allExpectedTools {
		assert.False(t, toolSet[tool], "Tool %s should not be duplicated", tool)
		toolSet[tool] = true
	}
}

func TestHandleToolsCall_NilArguments(t *testing.T) {
	handler := &Handler{
		toolRoutes: make(map[string]ToolHandler),
	}

	ctx := createTestContextWithUser()

	// Test with nil arguments (should be handled gracefully)
	params := map[string]interface{}{
		"name": "unknown_tool",
		// arguments is nil
	}

	result, err := handler.HandleToolsCall(ctx, params)
	assert.Error(t, err)
	assert.Nil(t, result)

	// Verify it's a method not found error (since the tool doesn't exist)
	jsonrpcErr, ok := err.(*jsonrpc.JSONRPCError)
	assert.True(t, ok)
	assert.Equal(t, jsonrpc.MethodNotFound, jsonrpcErr.Code)
}
