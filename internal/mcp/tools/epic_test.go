package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"product-requirements-management/internal/mcp/types"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// MockEpicService is a mock implementation of service.EpicService
type MockEpicService struct {
	mock.Mock
}

type MockUserService struct {
	mock.Mock
}

func (m *MockEpicService) CreateEpic(req service.CreateEpicRequest) (*models.Epic, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MockEpicService) UpdateEpic(id uuid.UUID, req service.UpdateEpicRequest) (*models.Epic, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MockEpicService) GetEpicByReferenceID(refID string) (*models.Epic, error) {
	args := m.Called(refID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

// Implement other required methods to satisfy the interface
func (m *MockEpicService) GetEpicByID(id uuid.UUID) (*models.Epic, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MockEpicService) DeleteEpic(id uuid.UUID, force bool) error {
	args := m.Called(id, force)
	return args.Error(0)
}

func (m *MockEpicService) ListEpics(filters service.EpicFilters) ([]models.Epic, int64, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.Epic), args.Get(1).(int64), args.Error(2)
}

func (m *MockEpicService) GetEpicWithUserStories(id uuid.UUID) (*models.Epic, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MockEpicService) ChangeEpicStatus(id uuid.UUID, newStatus models.EpicStatus) (*models.Epic, error) {
	args := m.Called(id, newStatus)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MockEpicService) AssignEpic(id uuid.UUID, assigneeID *uuid.UUID) (*models.Epic, error) {
	args := m.Called(id, assigneeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MockUserService) GetByName(name string) (*models.User, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestEpicHandler_GetSupportedTools(t *testing.T) {
	handler := NewEpicHandler(nil, nil)
	tools := handler.GetSupportedTools()

	expected := []string{"create_epic", "update_epic", "list_epics"}
	assert.Equal(t, expected, tools)
}

func TestEpicHandler_HandleTool(t *testing.T) {
	mockService := &MockEpicService{}
	handler := NewEpicHandler(mockService, nil)

	tests := []struct {
		name        string
		toolName    string
		expectError bool
	}{
		{
			name:        "valid create_epic tool",
			toolName:    "create_epic",
			expectError: true, // Will error due to missing context/args, but tool routing works
		},
		{
			name:        "valid update_epic tool",
			toolName:    "update_epic",
			expectError: true, // Will error due to missing context/args, but tool routing works
		},
		{
			name:        "invalid tool name",
			toolName:    "invalid_tool",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.HandleTool(context.Background(), tt.toolName, map[string]interface{}{})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEpicHandler_Create_ValidationErrors(t *testing.T) {
	mockService := &MockEpicService{}
	handler := NewEpicHandler(mockService, nil)

	tests := []struct {
		name string
		args map[string]interface{}
	}{
		{
			name: "missing title",
			args: map[string]interface{}{"priority": 1},
		},
		{
			name: "empty title",
			args: map[string]interface{}{"title": "", "priority": 1},
		},
		{
			name: "missing priority",
			args: map[string]interface{}{"title": "Test Epic"},
		},
		{
			name: "invalid assignee_id format",
			args: map[string]interface{}{"title": "Test Epic", "priority": 1, "assignee_id": "invalid-uuid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Create(context.Background(), tt.args)
			assert.Error(t, err)
		})
	}
}

func TestEpicHandler_Update_ValidationErrors(t *testing.T) {
	mockService := &MockEpicService{}
	handler := NewEpicHandler(mockService, nil)

	tests := []struct {
		name string
		args map[string]interface{}
	}{
		{
			name: "missing epic_id",
			args: map[string]interface{}{"title": "Updated Epic"},
		},
		{
			name: "empty epic_id",
			args: map[string]interface{}{"epic_id": "", "title": "Updated Epic"},
		},
		{
			name: "invalid assignee_id format",
			args: map[string]interface{}{"epic_id": uuid.New().String(), "assignee_id": "invalid-uuid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Update(context.Background(), tt.args)
			assert.Error(t, err)
		})
	}
}

func TestNewEpicHandler(t *testing.T) {
	mockService := &MockEpicService{}
	mockUserService := &MockUserService{}
	handler := NewEpicHandler(mockService, mockUserService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.epicService)
	assert.Equal(t, mockUserService, handler.userService)
}

// Helper function to create a context with user
func createContextWithUser(user *models.User) context.Context {
	ginCtx := &gin.Context{}
	ginCtx.Set("user", user)

	ctx := context.WithValue(context.Background(), "gin_context", ginCtx)
	return ctx
}

// TestEpicHandler_Create_ValidParameters tests epic creation with valid parameters
func TestEpicHandler_Create_ValidParameters(t *testing.T) {
	mockService := &MockEpicService{}
	handler := NewEpicHandler(mockService, nil)

	// Create test user
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	// Create test epic
	expectedEpic := &models.Epic{
		ID:          uuid.New(),
		ReferenceID: "EP-001",
		Title:       "Test Epic",
		Description: stringPtr("Test Description"),
		Priority:    models.PriorityHigh,
		Status:      models.EpicStatusBacklog,
		CreatorID:   user.ID,
	}

	tests := []struct {
		name     string
		args     map[string]interface{}
		expected *models.Epic
	}{
		{
			name: "create epic with all parameters",
			args: map[string]interface{}{
				"title":       "Test Epic",
				"description": "Test Description",
				"priority":    2,
				"assignee_id": uuid.New().String(),
			},
			expected: expectedEpic,
		},
		{
			name: "create epic with minimal parameters",
			args: map[string]interface{}{
				"title":    "Minimal Epic",
				"priority": 1,
			},
			expected: expectedEpic,
		},
		{
			name: "create epic with float priority",
			args: map[string]interface{}{
				"title":    "Float Priority Epic",
				"priority": 3.0,
			},
			expected: expectedEpic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			mockService.On("CreateEpic", mock.AnythingOfType("service.CreateEpicRequest")).Return(tt.expected, nil).Once()

			ctx := createContextWithUser(user)
			result, err := handler.Create(ctx, tt.args)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Verify response format
			response, ok := result.(*types.ToolResponse)
			assert.True(t, ok)
			assert.Len(t, response.Content, 2) // Message + data
			assert.Equal(t, "text", response.Content[0].Type)
			assert.Contains(t, response.Content[0].Text, "Successfully created epic")
			assert.Contains(t, response.Content[0].Text, tt.expected.ReferenceID)

			mockService.AssertExpectations(t)
		})
	}
}

// TestEpicHandler_Create_InvalidParameters tests epic creation with invalid parameters
func TestEpicHandler_Create_InvalidParameters(t *testing.T) {
	mockService := &MockEpicService{}
	handler := NewEpicHandler(mockService, nil)

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError string
	}{
		{
			name:        "missing title",
			args:        map[string]interface{}{"priority": 1},
			expectError: "Invalid params",
		},
		{
			name:        "empty title",
			args:        map[string]interface{}{"title": "", "priority": 1},
			expectError: "Invalid params",
		},
		{
			name:        "missing priority",
			args:        map[string]interface{}{"title": "Test Epic"},
			expectError: "Invalid params",
		},
		{
			name:        "invalid priority type",
			args:        map[string]interface{}{"title": "Test Epic", "priority": "high"},
			expectError: "Invalid params",
		},
		{
			name:        "invalid assignee_id format",
			args:        map[string]interface{}{"title": "Test Epic", "priority": 1, "assignee_id": "invalid-uuid"},
			expectError: "Invalid params",
		},
		{
			name:        "null title",
			args:        map[string]interface{}{"title": nil, "priority": 1},
			expectError: "Invalid params",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createContextWithUser(user)
			_, err := handler.Create(ctx, tt.args)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

// TestEpicHandler_Create_ServiceErrors tests epic creation service layer errors
func TestEpicHandler_Create_ServiceErrors(t *testing.T) {
	mockService := &MockEpicService{}
	handler := NewEpicHandler(mockService, nil)

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	args := map[string]interface{}{
		"title":    "Test Epic",
		"priority": 1,
	}

	// Setup mock to return error
	mockService.On("CreateEpic", mock.AnythingOfType("service.CreateEpicRequest")).Return(nil, errors.New("database error")).Once()

	ctx := createContextWithUser(user)
	_, err := handler.Create(ctx, args)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Internal error")
	mockService.AssertExpectations(t)
}

// TestEpicHandler_Create_ContextErrors tests epic creation context errors
func TestEpicHandler_Create_ContextErrors(t *testing.T) {
	mockService := &MockEpicService{}
	handler := NewEpicHandler(mockService, nil)

	args := map[string]interface{}{
		"title":    "Test Epic",
		"priority": 1,
	}

	tests := []struct {
		name        string
		ctx         context.Context
		expectError string
	}{
		{
			name:        "missing gin context",
			ctx:         context.Background(),
			expectError: "Internal error",
		},
		{
			name:        "gin context without user",
			ctx:         context.WithValue(context.Background(), "gin_context", &gin.Context{}),
			expectError: "Internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Create(tt.ctx, args)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

// TestEpicHandler_Update_ValidParameters tests epic updates with valid parameters
func TestEpicHandler_Update_ValidParameters(t *testing.T) {
	mockService := &MockEpicService{}
	handler := NewEpicHandler(mockService, nil)

	epicID := uuid.New()
	expectedEpic := &models.Epic{
		ID:          epicID,
		ReferenceID: "EP-001",
		Title:       "Updated Epic",
		Priority:    models.PriorityHigh,
		Status:      models.EpicStatusInProgress,
	}

	tests := []struct {
		name     string
		args     map[string]interface{}
		expected *models.Epic
	}{
		{
			name: "update with UUID",
			args: map[string]interface{}{
				"epic_id": epicID.String(),
				"title":   "Updated Epic",
			},
			expected: expectedEpic,
		},
		{
			name: "update with all parameters",
			args: map[string]interface{}{
				"epic_id":     epicID.String(),
				"title":       "Updated Epic",
				"description": "Updated Description",
				"priority":    2,
				"status":      "In Progress",
				"assignee_id": uuid.New().String(),
			},
			expected: expectedEpic,
		},
		{
			name: "update with empty assignee_id (unassign)",
			args: map[string]interface{}{
				"epic_id":     epicID.String(),
				"assignee_id": "",
			},
			expected: expectedEpic,
		},
		{
			name: "update with float priority",
			args: map[string]interface{}{
				"epic_id":  epicID.String(),
				"priority": 3.0,
			},
			expected: expectedEpic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			mockService.On("UpdateEpic", epicID, mock.AnythingOfType("service.UpdateEpicRequest")).Return(tt.expected, nil).Once()

			result, err := handler.Update(context.Background(), tt.args)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Verify response format
			response, ok := result.(*types.ToolResponse)
			assert.True(t, ok)
			assert.Len(t, response.Content, 2) // Message + data
			assert.Equal(t, "text", response.Content[0].Type)
			assert.Contains(t, response.Content[0].Text, "Successfully updated epic")
			assert.Contains(t, response.Content[0].Text, tt.expected.ReferenceID)

			mockService.AssertExpectations(t)
		})
	}
}

// TestEpicHandler_Update_ReferenceIDResolution tests epic updates with reference ID resolution
func TestEpicHandler_Update_ReferenceIDResolution(t *testing.T) {
	mockService := &MockEpicService{}
	handler := NewEpicHandler(mockService, nil)

	epicID := uuid.New()
	expectedEpic := &models.Epic{
		ID:          epicID,
		ReferenceID: "EP-001",
		Title:       "Updated Epic",
	}

	tests := []struct {
		name        string
		args        map[string]interface{}
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid reference ID resolution",
			args: map[string]interface{}{
				"epic_id": "EP-001",
				"title":   "Updated Epic",
			},
			setupMocks: func() {
				mockService.On("GetEpicByReferenceID", "EP-001").Return(expectedEpic, nil).Once()
				mockService.On("UpdateEpic", epicID, mock.AnythingOfType("service.UpdateEpicRequest")).Return(expectedEpic, nil).Once()
			},
			expectError: false,
		},
		{
			name: "invalid reference ID",
			args: map[string]interface{}{
				"epic_id": "EP-999",
				"title":   "Updated Epic",
			},
			setupMocks: func() {
				mockService.On("GetEpicByReferenceID", "EP-999").Return(nil, errors.New("epic not found")).Once()
			},
			expectError: true,
			errorMsg:    "Invalid params",
		},
		{
			name: "invalid ID format",
			args: map[string]interface{}{
				"epic_id": "invalid-id",
				"title":   "Updated Epic",
			},
			setupMocks: func() {
				mockService.On("GetEpicByReferenceID", "invalid-id").Return(nil, errors.New("epic not found")).Once()
			},
			expectError: true,
			errorMsg:    "Invalid params",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			result, err := handler.Update(context.Background(), tt.args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockService.AssertExpectations(t)
		})
	}
}

// TestEpicHandler_Update_InvalidParameters tests epic updates with invalid parameters
func TestEpicHandler_Update_InvalidParameters(t *testing.T) {
	mockService := &MockEpicService{}
	handler := NewEpicHandler(mockService, nil)

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError string
	}{
		{
			name:        "missing epic_id",
			args:        map[string]interface{}{"title": "Updated Epic"},
			expectError: "Invalid params",
		},
		{
			name:        "empty epic_id",
			args:        map[string]interface{}{"epic_id": "", "title": "Updated Epic"},
			expectError: "Invalid params",
		},
		{
			name:        "null epic_id",
			args:        map[string]interface{}{"epic_id": nil, "title": "Updated Epic"},
			expectError: "Invalid params",
		},
		{
			name:        "invalid assignee_id format",
			args:        map[string]interface{}{"epic_id": uuid.New().String(), "assignee_id": "invalid-uuid"},
			expectError: "Invalid params",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Update(context.Background(), tt.args)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

// TestEpicHandler_Update_ServiceErrors tests epic update service layer errors
func TestEpicHandler_Update_ServiceErrors(t *testing.T) {
	mockService := &MockEpicService{}
	handler := NewEpicHandler(mockService, nil)

	epicID := uuid.New()
	args := map[string]interface{}{
		"epic_id": epicID.String(),
		"title":   "Updated Epic",
	}

	// Setup mock to return error
	mockService.On("UpdateEpic", epicID, mock.AnythingOfType("service.UpdateEpicRequest")).Return(nil, errors.New("database error")).Once()

	_, err := handler.Update(context.Background(), args)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Internal error")
	mockService.AssertExpectations(t)
}

// TestEpicHandler_HandleTool_ErrorHandling tests error handling in HandleTool method
func TestEpicHandler_HandleTool_ErrorHandling(t *testing.T) {
	mockService := &MockEpicService{}
	handler := NewEpicHandler(mockService, nil)

	tests := []struct {
		name        string
		toolName    string
		args        map[string]interface{}
		expectError string
	}{
		{
			name:        "unknown tool",
			toolName:    "unknown_tool",
			args:        map[string]interface{}{},
			expectError: "Method not found",
		},
		{
			name:        "create_epic with invalid args",
			toolName:    "create_epic",
			args:        map[string]interface{}{},
			expectError: "Internal error",
		},
		{
			name:        "update_epic with invalid args",
			toolName:    "update_epic",
			args:        map[string]interface{}{},
			expectError: "Invalid params",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.HandleTool(context.Background(), tt.toolName, tt.args)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

// Test response format validation
func TestEpicHandler_ResponseFormat(t *testing.T) {
	// Test that responses follow the expected ToolResponse format
	message := "Test message"
	data := map[string]interface{}{"test": "data"}

	response := types.CreateDataResponse(message, data)

	assert.NotNil(t, response)
	assert.IsType(t, &types.ToolResponse{}, response)
	assert.Len(t, response.Content, 2) // Message + data
	assert.Equal(t, "text", response.Content[0].Type)
	assert.Equal(t, message, response.Content[0].Text)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
