package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/jsonrpc"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// Mock services for MCP integration testing with unique names
type MCPTestEpicService struct {
	mock.Mock
}

func (m *MCPTestEpicService) CreateEpic(req service.CreateEpicRequest) (*models.Epic, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MCPTestEpicService) GetEpicByReferenceID(referenceID string) (*models.Epic, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MCPTestEpicService) UpdateEpic(id uuid.UUID, req service.UpdateEpicRequest) (*models.Epic, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MCPTestEpicService) GetEpic(id uuid.UUID) (*models.Epic, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MCPTestEpicService) ListEpics(options service.ListEpicsOptions) (*service.EpicListResponse, error) {
	args := m.Called(options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.EpicListResponse), args.Error(1)
}

func (m *MCPTestEpicService) DeleteEpic(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MCPTestEpicService) ChangeEpicStatus(id uuid.UUID, status models.EpicStatus) (*models.Epic, error) {
	args := m.Called(id, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MCPTestEpicService) AssignEpic(id uuid.UUID, assigneeID *uuid.UUID) (*models.Epic, error) {
	args := m.Called(id, assigneeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

type MCPTestUserStoryService struct {
	mock.Mock
}

func (m *MCPTestUserStoryService) CreateUserStory(req service.CreateUserStoryRequest) (*models.UserStory, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MCPTestUserStoryService) GetUserStoryByReferenceID(referenceID string) (*models.UserStory, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MCPTestUserStoryService) UpdateUserStory(id uuid.UUID, req service.UpdateUserStoryRequest) (*models.UserStory, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MCPTestUserStoryService) GetUserStory(id uuid.UUID) (*models.UserStory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MCPTestUserStoryService) ListUserStories(options service.ListUserStoriesOptions) (*service.UserStoryListResponse, error) {
	args := m.Called(options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.UserStoryListResponse), args.Error(1)
}

func (m *MCPTestUserStoryService) DeleteUserStory(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MCPTestUserStoryService) ChangeUserStoryStatus(id uuid.UUID, status models.UserStoryStatus) (*models.UserStory, error) {
	args := m.Called(id, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MCPTestUserStoryService) AssignUserStory(id uuid.UUID, assigneeID *uuid.UUID) (*models.UserStory, error) {
	args := m.Called(id, assigneeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

type MCPTestRequirementService struct {
	mock.Mock
}

func (m *MCPTestRequirementService) CreateRequirement(req service.CreateRequirementRequest) (*models.Requirement, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MCPTestRequirementService) GetRequirementByReferenceID(referenceID string) (*models.Requirement, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MCPTestRequirementService) UpdateRequirement(id uuid.UUID, req service.UpdateRequirementRequest) (*models.Requirement, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MCPTestRequirementService) CreateRelationship(req service.CreateRelationshipRequest) (*models.RequirementRelationship, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RequirementRelationship), args.Error(1)
}

func (m *MCPTestRequirementService) SearchRequirements(query string) ([]*models.Requirement, error) {
	args := m.Called(query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Requirement), args.Error(1)
}

func (m *MCPTestRequirementService) GetRequirement(id uuid.UUID) (*models.Requirement, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MCPTestRequirementService) ListRequirements(options service.ListRequirementsOptions) (*service.RequirementListResponse, error) {
	args := m.Called(options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.RequirementListResponse), args.Error(1)
}

func (m *MCPTestRequirementService) DeleteRequirement(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MCPTestRequirementService) ChangeRequirementStatus(id uuid.UUID, status models.RequirementStatus) (*models.Requirement, error) {
	args := m.Called(id, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MCPTestRequirementService) AssignRequirement(id uuid.UUID, assigneeID *uuid.UUID) (*models.Requirement, error) {
	args := m.Called(id, assigneeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

type MCPTestAcceptanceCriteriaService struct {
	mock.Mock
}

func (m *MCPTestAcceptanceCriteriaService) CreateAcceptanceCriteria(req service.CreateAcceptanceCriteriaRequest) (*models.AcceptanceCriteria, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AcceptanceCriteria), args.Error(1)
}

func (m *MCPTestAcceptanceCriteriaService) GetAcceptanceCriteria(id uuid.UUID) (*models.AcceptanceCriteria, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AcceptanceCriteria), args.Error(1)
}

func (m *MCPTestAcceptanceCriteriaService) UpdateAcceptanceCriteria(id uuid.UUID, req service.UpdateAcceptanceCriteriaRequest) (*models.AcceptanceCriteria, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AcceptanceCriteria), args.Error(1)
}

func (m *MCPTestAcceptanceCriteriaService) ListAcceptanceCriteria(options service.ListAcceptanceCriteriaOptions) (*service.AcceptanceCriteriaListResponse, error) {
	args := m.Called(options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AcceptanceCriteriaListResponse), args.Error(1)
}

func (m *MCPTestAcceptanceCriteriaService) DeleteAcceptanceCriteria(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

type MCPTestSearchService struct {
	mock.Mock
}

func (m *MCPTestSearchService) Search(ctx context.Context, options service.SearchOptions) (*service.SearchResponse, error) {
	args := m.Called(ctx, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.SearchResponse), args.Error(1)
}

func (m *MCPTestSearchService) GetSuggestions(ctx context.Context, query string, limit int) (*service.SearchSuggestionsResponse, error) {
	args := m.Called(ctx, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.SearchSuggestionsResponse), args.Error(1)
}

// Test setup helpers
func setupMCPTestHandler() (*MCPHandler, *MCPTestEpicService, *MCPTestUserStoryService, *MCPTestRequirementService, *MCPTestAcceptanceCriteriaService, *MCPTestSearchService) {
	mockEpicService := &MCPTestEpicService{}
	mockUserStoryService := &MCPTestUserStoryService{}
	mockRequirementService := &MCPTestRequirementService{}
	mockAcceptanceCriteriaService := &MCPTestAcceptanceCriteriaService{}
	mockSearchService := &MCPTestSearchService{}

	handler := NewMCPHandler(
		mockEpicService,
		mockUserStoryService,
		mockRequirementService,
		mockAcceptanceCriteriaService,
		mockSearchService,
	)

	return handler, mockEpicService, mockUserStoryService, mockRequirementService, mockAcceptanceCriteriaService, mockSearchService
}

func setupMCPTestGinContext(user *models.User) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set up user in context (simulating PAT authentication)
	if user != nil {
		c.Set("user", user)
	}

	return c, w
}

func createMCPTestUser() *models.User {
	return &models.User{
		ID:       uuid.New(),
		Username: "mcptestuser",
		Email:    "mcptest@example.com",
		Role:     models.UserRoleUser,
	}
}

func createMCPJSONRPCRequest(id interface{}, method string, params interface{}) map[string]interface{} {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
	}

	if id != nil {
		req["id"] = id
	}

	if params != nil {
		req["params"] = params
	}

	return req
}

func TestMCPHandler_FullRequestFlow_Initialize(t *testing.T) {
	handler, _, _, _, _, _ := setupMCPTestHandler()
	user := createMCPTestUser()
	c, w := setupMCPTestGinContext(user)

	// Create initialize request
	request := createMCPJSONRPCRequest(1, "initialize", map[string]interface{}{
		"protocolVersion": "2025-06-18",
		"capabilities": map[string]interface{}{
			"elicitation": map[string]interface{}{},
		},
		"clientInfo": map[string]interface{}{
			"name":    "test-client",
			"version": "1.0.0",
		},
	})

	requestBody, err := json.Marshal(request)
	require.NoError(t, err)

	c.Request = httptest.NewRequest("POST", "/api/v1/mcp", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Authorization", "Bearer mcp_pat_test_token")

	// Process the request
	handler.Process(c)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response jsonrpc.JSONRPCResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, 1, response.ID.GetValue())
	assert.Nil(t, response.Error)
	assert.NotNil(t, response.Result)

	// Verify initialize response structure
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "2025-06-18", result["protocolVersion"])
	assert.NotNil(t, result["capabilities"])
	assert.NotNil(t, result["serverInfo"])
}

func TestMCPHandler_FullRequestFlow_ToolsList(t *testing.T) {
	handler, _, _, _, _, _ := setupMCPTestHandler()
	user := createMCPTestUser()
	c, w := setupMCPTestGinContext(user)

	// Create tools/list request
	request := createMCPJSONRPCRequest(2, "tools/list", nil)

	requestBody, err := json.Marshal(request)
	require.NoError(t, err)

	c.Request = httptest.NewRequest("POST", "/api/v1/mcp", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Authorization", "Bearer mcp_pat_test_token")

	// Process the request
	handler.Process(c)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response jsonrpc.JSONRPCResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, 2, response.ID.GetValue())
	assert.Nil(t, response.Error)
	assert.NotNil(t, response.Result)

	// Verify tools list response structure
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok)
	tools, ok := result["tools"].([]interface{})
	require.True(t, ok)
	assert.Greater(t, len(tools), 0)

	// Verify first tool structure
	firstTool, ok := tools[0].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, firstTool["name"])
	assert.NotEmpty(t, firstTool["description"])
	assert.NotNil(t, firstTool["inputSchema"])
}

func TestMCPHandler_FullRequestFlow_CreateEpic(t *testing.T) {
	handler, mockEpicService, _, _, _, _ := setupMCPTestHandler()
	user := createMCPTestUser()
	c, w := setupMCPTestGinContext(user)

	// Setup mock expectations
	expectedEpic := &models.Epic{
		ID:          uuid.New(),
		ReferenceID: "EP-001",
		Title:       "Test Epic",
		Description: "Test Description",
		Priority:    models.PriorityHigh,
		CreatorID:   user.ID,
		Status:      models.EpicStatusBacklog,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockEpicService.On("CreateEpic", mock.MatchedBy(func(req service.CreateEpicRequest) bool {
		return req.Title == "Test Epic" && req.CreatorID == user.ID && req.Priority == models.PriorityHigh
	})).Return(expectedEpic, nil)

	// Create tools/call request for create_epic
	request := createMCPJSONRPCRequest(3, "tools/call", map[string]interface{}{
		"name": "create_epic",
		"arguments": map[string]interface{}{
			"title":       "Test Epic",
			"description": "Test Description",
			"priority":    2, // High priority
		},
	})

	requestBody, err := json.Marshal(request)
	require.NoError(t, err)

	c.Request = httptest.NewRequest("POST", "/api/v1/mcp", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Authorization", "Bearer mcp_pat_test_token")

	// Process the request
	handler.Process(c)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response jsonrpc.JSONRPCResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, 3, response.ID.GetValue())
	assert.Nil(t, response.Error)
	assert.NotNil(t, response.Result)

	// Verify tool response structure
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok)
	content, ok := result["content"].([]interface{})
	require.True(t, ok)
	assert.Len(t, content, 2) // Text and data content items

	// Verify mock was called
	mockEpicService.AssertExpectations(t)
}

func TestMCPHandler_FullRequestFlow_ResourcesRead(t *testing.T) {
	handler, mockEpicService, _, _, _, _ := setupMCPTestHandler()
	user := createMCPTestUser()
	c, w := setupMCPTestGinContext(user)

	// Setup mock expectations
	expectedEpic := &models.Epic{
		ID:          uuid.New(),
		ReferenceID: "EP-001",
		Title:       "Test Epic",
		Description: "Test Description",
		Priority:    models.PriorityHigh,
		CreatorID:   user.ID,
		Status:      models.EpicStatusBacklog,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockEpicService.On("GetEpicByReferenceID", "EP-001").Return(expectedEpic, nil)

	// Create resources/read request
	request := createMCPJSONRPCRequest(4, "resources/read", map[string]interface{}{
		"uri": "epic://EP-001",
	})

	requestBody, err := json.Marshal(request)
	require.NoError(t, err)

	c.Request = httptest.NewRequest("POST", "/api/v1/mcp", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Authorization", "Bearer mcp_pat_test_token")

	// Process the request
	handler.Process(c)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response jsonrpc.JSONRPCResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, 4, response.ID.GetValue())
	assert.Nil(t, response.Error)
	assert.NotNil(t, response.Result)

	// Verify resource response structure
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "epic://EP-001", result["uri"])
	assert.NotEmpty(t, result["name"])
	assert.NotEmpty(t, result["mimeType"])
	assert.NotNil(t, result["contents"])

	// Verify mock was called
	mockEpicService.AssertExpectations(t)
}

func TestMCPHandler_FullRequestFlow_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		request        map[string]interface{}
		expectedStatus int
		expectedError  int
	}{
		{
			name: "invalid JSON-RPC version",
			request: map[string]interface{}{
				"jsonrpc": "1.0",
				"id":      1,
				"method":  "initialize",
			},
			expectedStatus: http.StatusOK,
			expectedError:  jsonrpc.InvalidRequest,
		},
		{
			name: "missing method",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
			},
			expectedStatus: http.StatusOK,
			expectedError:  jsonrpc.InvalidRequest,
		},
		{
			name: "unknown method",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"method":  "unknown_method",
			},
			expectedStatus: http.StatusOK,
			expectedError:  jsonrpc.MethodNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _, _, _, _, _ := setupMCPTestHandler()
			user := createMCPTestUser()
			c, w := setupMCPTestGinContext(user)

			requestBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			c.Request = httptest.NewRequest("POST", "/api/v1/mcp", bytes.NewBuffer(requestBody))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer mcp_pat_test_token")

			// Process the request
			handler.Process(c)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response jsonrpc.JSONRPCResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, "2.0", response.JSONRPC)
			assert.Nil(t, response.Result)
			require.NotNil(t, response.Error)
			assert.Equal(t, tt.expectedError, response.Error.Code)
		})
	}
}

func TestMCPHandler_FullRequestFlow_ServiceErrors(t *testing.T) {
	tests := []struct {
		name          string
		serviceError  error
		expectedError int
	}{
		{
			name:          "service not found error",
			serviceError:  service.ErrEpicNotFound,
			expectedError: jsonrpc.ResourceNotFound,
		},
		{
			name:          "service validation error",
			serviceError:  service.ErrInvalidPriority,
			expectedError: jsonrpc.ValidationError,
		},
		{
			name:          "generic service error",
			serviceError:  fmt.Errorf("generic error"),
			expectedError: jsonrpc.InternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockEpicService, _, _, _, _ := setupMCPTestHandler()
			user := createMCPTestUser()
			c, w := setupMCPTestGinContext(user)

			// Setup mock to return error
			mockEpicService.On("CreateEpic", mock.AnythingOfType("service.CreateEpicRequest")).Return(nil, tt.serviceError)

			// Create tools/call request for create_epic
			request := createMCPJSONRPCRequest(1, "tools/call", map[string]interface{}{
				"name": "create_epic",
				"arguments": map[string]interface{}{
					"title":    "Test Epic",
					"priority": 2,
				},
			})

			requestBody, err := json.Marshal(request)
			require.NoError(t, err)

			c.Request = httptest.NewRequest("POST", "/api/v1/mcp", bytes.NewBuffer(requestBody))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer mcp_pat_test_token")

			// Process the request
			handler.Process(c)

			// Verify response
			assert.Equal(t, http.StatusOK, w.Code)

			var response jsonrpc.JSONRPCResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, "2.0", response.JSONRPC)
			assert.Nil(t, response.Result)
			require.NotNil(t, response.Error)
			assert.Equal(t, tt.expectedError, response.Error.Code)

			// Verify mock was called
			mockEpicService.AssertExpectations(t)
		})
	}
}
