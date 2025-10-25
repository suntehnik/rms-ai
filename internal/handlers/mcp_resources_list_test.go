package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockResourceService implements service.ResourceService for testing
type MockResourceService struct {
	mock.Mock
}

func (m *MockResourceService) GetResourceList(ctx context.Context) ([]service.ResourceDescriptor, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.ResourceDescriptor), args.Error(1)
}

func TestMCPHandler_ResourcesList_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock resource service
	mockResourceService := new(MockResourceService)

	// Setup expected resources
	expectedResources := []service.ResourceDescriptor{
		{
			URI:         "requirements://epics/EP-001",
			Name:        "Epic: Test Epic",
			Description: "Epic EP-001: Test Epic",
			MimeType:    "application/json",
		},
		{
			URI:         "requirements://search/{query}",
			Name:        "Search Requirements",
			Description: "Search across all epics, user stories, and requirements",
			MimeType:    "application/json",
		},
	}

	mockResourceService.On("GetResourceList", mock.Anything).Return(expectedResources, nil)

	// Create MCP handler with mock resource service
	handler := NewMCPHandler(nil, nil, nil, nil, nil, nil, nil, mockResourceService)

	// Create test request
	requestBody := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "resources/list"
	}`

	req := httptest.NewRequest("POST", "/api/v1/mcp", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Add mock user to context (required for logging)
	c.Set("user", &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	})

	// Process the request
	handler.Process(c)

	// Check status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check JSON-RPC structure
	assert.Equal(t, "2.0", response["jsonrpc"])
	assert.Equal(t, float64(1), response["id"])
	assert.Nil(t, response["error"])

	// Check result structure
	result, ok := response["result"].(map[string]interface{})
	assert.True(t, ok, "Result should be an object")

	resources, ok := result["resources"].([]interface{})
	assert.True(t, ok, "Resources should be an array")
	assert.Len(t, resources, 2, "Should return 2 resources")

	// Check first resource
	firstResource := resources[0].(map[string]interface{})
	assert.Equal(t, "requirements://epics/EP-001", firstResource["uri"])
	assert.Equal(t, "Epic: Test Epic", firstResource["name"])
	assert.Equal(t, "Epic EP-001: Test Epic", firstResource["description"])
	assert.Equal(t, "application/json", firstResource["mimeType"])

	// Check second resource
	secondResource := resources[1].(map[string]interface{})
	assert.Equal(t, "requirements://search/{query}", secondResource["uri"])
	assert.Equal(t, "Search Requirements", secondResource["name"])
	assert.Equal(t, "Search across all epics, user stories, and requirements", secondResource["description"])
	assert.Equal(t, "application/json", secondResource["mimeType"])

	// Verify mock was called
	mockResourceService.AssertExpectations(t)
}

func TestMCPHandler_ResourcesList_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock resource service that returns an error
	mockResourceService := new(MockResourceService)
	mockResourceService.On("GetResourceList", mock.Anything).Return([]service.ResourceDescriptor{}, assert.AnError)

	// Create MCP handler with mock resource service
	handler := NewMCPHandler(nil, nil, nil, nil, nil, nil, nil, mockResourceService)

	// Create test request
	requestBody := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "resources/list"
	}`

	req := httptest.NewRequest("POST", "/api/v1/mcp", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Add mock user to context (required for logging)
	c.Set("user", &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	})

	// Process the request
	handler.Process(c)

	// Check status code (should still be 200 for JSON-RPC errors)
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check JSON-RPC structure
	assert.Equal(t, "2.0", response["jsonrpc"])
	assert.Equal(t, float64(1), response["id"])
	assert.Nil(t, response["result"])

	// Check error structure
	errorObj, ok := response["error"].(map[string]interface{})
	assert.True(t, ok, "Error should be an object")
	assert.NotNil(t, errorObj["code"])
	assert.NotNil(t, errorObj["message"])

	// Verify mock was called
	mockResourceService.AssertExpectations(t)
}

func TestMCPHandler_ResourcesList_Integration(t *testing.T) {
	// Test that resources/list method is properly registered
	gin.SetMode(gin.TestMode)

	// Create MCP handler with nil resource service (just for registration test)
	handler := NewMCPHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	// Check that resources/list method is registered
	methods := handler.processor.GetRegisteredMethods()
	assert.Contains(t, methods, "resources/list", "resources/list method should be registered")

	// Test that resources/list method exists
	assert.True(t, handler.processor.HasMethod("resources/list"), "resources/list method should exist")
}

func TestHandleResourcesList_Direct(t *testing.T) {
	// Test the handleResourcesList method directly
	gin.SetMode(gin.TestMode)

	// Create mock resource service
	mockResourceService := new(MockResourceService)

	expectedResources := []service.ResourceDescriptor{
		{
			URI:      "requirements://epics/EP-001",
			Name:     "Epic: Test Epic",
			MimeType: "application/json",
		},
	}

	mockResourceService.On("GetResourceList", mock.Anything).Return(expectedResources, nil)

	// Create handler
	handler := &MCPHandler{
		resourceService: mockResourceService,
		mcpLogger:       NewMCPLogger(),
	}

	// Create context
	ctx := context.Background()

	// Call the method directly
	result, err := handler.handleResourcesList(ctx, nil)

	// Check result
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	resources, ok := resultMap["resources"].([]service.ResourceDescriptor)
	assert.True(t, ok)
	assert.Len(t, resources, 1)
	assert.Equal(t, "requirements://epics/EP-001", resources[0].URI)

	// Verify mock was called
	mockResourceService.AssertExpectations(t)
}
