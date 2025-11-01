package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"product-requirements-management/internal/mcp/types"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// MockUserStoryService is a mock implementation of service.UserStoryService
type MockUserStoryService struct {
	mock.Mock
}

func (m *MockUserStoryService) CreateUserStory(req service.CreateUserStoryRequest) (*models.UserStory, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryService) UpdateUserStory(id uuid.UUID, req service.UpdateUserStoryRequest) (*models.UserStory, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryService) GetUserStoryByReferenceID(refID string) (*models.UserStory, error) {
	args := m.Called(refID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

// Implement other required methods to satisfy the interface
func (m *MockUserStoryService) GetUserStoryByID(id uuid.UUID) (*models.UserStory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryService) DeleteUserStory(id uuid.UUID, force bool) error {
	args := m.Called(id, force)
	return args.Error(0)
}

func (m *MockUserStoryService) ListUserStories(filters service.UserStoryFilters) ([]models.UserStory, int64, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.UserStory), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserStoryService) GetUserStoryWithAcceptanceCriteria(id uuid.UUID) (*models.UserStory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryService) GetUserStoryWithRequirements(id uuid.UUID) (*models.UserStory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryService) ChangeUserStoryStatus(id uuid.UUID, newStatus models.UserStoryStatus) (*models.UserStory, error) {
	args := m.Called(id, newStatus)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryService) GetUserStoriesByEpic(epicID uuid.UUID) ([]models.UserStory, error) {
	args := m.Called(epicID)
	return args.Get(0).([]models.UserStory), args.Error(1)
}

func (m *MockUserStoryService) AssignUserStory(id uuid.UUID, assigneeID uuid.UUID) (*models.UserStory, error) {
	args := m.Called(id, assigneeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryService) GetUUIDByReferenceID(referenceID string) (uuid.UUID, error) {
	args := m.Called(referenceID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func TestUserStoryHandler_GetSupportedTools(t *testing.T) {
	handler := NewUserStoryHandler(nil, nil)
	tools := handler.GetSupportedTools()

	expected := []string{"create_user_story", "update_user_story"}
	assert.Equal(t, expected, tools)
}

func TestUserStoryHandler_HandleTool(t *testing.T) {
	mockUserStoryService := &MockUserStoryService{}
	mockEpicService := &MockEpicService{}
	handler := NewUserStoryHandler(mockUserStoryService, mockEpicService)

	tests := []struct {
		name        string
		toolName    string
		expectError bool
	}{
		{
			name:        "valid create_user_story tool",
			toolName:    "create_user_story",
			expectError: true, // Will error due to missing context/args, but tool routing works
		},
		{
			name:        "valid update_user_story tool",
			toolName:    "update_user_story",
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

func TestUserStoryHandler_Create_ValidationErrors(t *testing.T) {
	mockUserStoryService := &MockUserStoryService{}
	mockEpicService := &MockEpicService{}
	handler := NewUserStoryHandler(mockUserStoryService, mockEpicService)

	tests := []struct {
		name string
		args map[string]interface{}
	}{
		{
			name: "missing title",
			args: map[string]interface{}{"epic_id": uuid.New().String(), "priority": 1},
		},
		{
			name: "empty title",
			args: map[string]interface{}{"title": "", "epic_id": uuid.New().String(), "priority": 1},
		},
		{
			name: "missing epic_id",
			args: map[string]interface{}{"title": "Test User Story", "priority": 1},
		},
		{
			name: "empty epic_id",
			args: map[string]interface{}{"title": "Test User Story", "epic_id": "", "priority": 1},
		},
		{
			name: "missing priority",
			args: map[string]interface{}{"title": "Test User Story", "epic_id": uuid.New().String()},
		},
		{
			name: "invalid assignee_id format",
			args: map[string]interface{}{"title": "Test User Story", "epic_id": uuid.New().String(), "priority": 1, "assignee_id": "invalid-uuid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Create(context.Background(), tt.args)
			assert.Error(t, err)
		})
	}
}

func TestUserStoryHandler_Update_ValidationErrors(t *testing.T) {
	mockUserStoryService := &MockUserStoryService{}
	mockEpicService := &MockEpicService{}
	handler := NewUserStoryHandler(mockUserStoryService, mockEpicService)

	tests := []struct {
		name string
		args map[string]interface{}
	}{
		{
			name: "missing user_story_id",
			args: map[string]interface{}{"title": "Updated User Story"},
		},
		{
			name: "empty user_story_id",
			args: map[string]interface{}{"user_story_id": "", "title": "Updated User Story"},
		},
		{
			name: "invalid assignee_id format",
			args: map[string]interface{}{"user_story_id": uuid.New().String(), "assignee_id": "invalid-uuid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Update(context.Background(), tt.args)
			assert.Error(t, err)
		})
	}
}

func TestNewUserStoryHandler(t *testing.T) {
	mockUserStoryService := &MockUserStoryService{}
	mockEpicService := &MockEpicService{}
	handler := NewUserStoryHandler(mockUserStoryService, mockEpicService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockUserStoryService, handler.userStoryService)
	assert.Equal(t, mockEpicService, handler.epicService)
}

// TestUserStoryHandler_Create_ValidParameters tests user story creation with valid parameters
func TestUserStoryHandler_Create_ValidParameters(t *testing.T) {
	mockUserStoryService := &MockUserStoryService{}
	mockEpicService := &MockEpicService{}
	handler := NewUserStoryHandler(mockUserStoryService, mockEpicService)

	// Create test user
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	// Create test epic
	epicID := uuid.New()
	epic := &models.Epic{
		ID:          epicID,
		ReferenceID: "EP-001",
		Title:       "Test Epic",
	}

	// Create test user story
	expectedUserStory := &models.UserStory{
		ID:          uuid.New(),
		ReferenceID: "US-001",
		Title:       "Test User Story",
		Description: stringPtr("Test Description"),
		Priority:    models.PriorityHigh,
		Status:      models.UserStoryStatusBacklog,
		EpicID:      epicID,
		CreatorID:   user.ID,
	}

	tests := []struct {
		name     string
		args     map[string]interface{}
		expected *models.UserStory
	}{
		{
			name: "create user story with all parameters",
			args: map[string]interface{}{
				"title":       "Test User Story",
				"description": "Test Description",
				"epic_id":     epicID.String(),
				"priority":    2,
				"assignee_id": uuid.New().String(),
			},
			expected: expectedUserStory,
		},
		{
			name: "create user story with minimal parameters",
			args: map[string]interface{}{
				"title":    "Minimal User Story",
				"epic_id":  epicID.String(),
				"priority": 1,
			},
			expected: expectedUserStory,
		},
		{
			name: "create user story with epic reference ID",
			args: map[string]interface{}{
				"title":    "User Story with Epic Ref",
				"epic_id":  "EP-001",
				"priority": 3,
			},
			expected: expectedUserStory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			if epicRefID, ok := tt.args["epic_id"].(string); ok && epicRefID == "EP-001" {
				mockEpicService.On("GetEpicByReferenceID", "EP-001").Return(epic, nil).Once()
			}
			mockUserStoryService.On("CreateUserStory", mock.AnythingOfType("service.CreateUserStoryRequest")).Return(tt.expected, nil).Once()

			ctx := createContextWithUser(user)
			result, err := handler.Create(ctx, tt.args)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Verify response format
			response, ok := result.(*types.ToolResponse)
			assert.True(t, ok)
			assert.Len(t, response.Content, 2) // Message + data
			assert.Equal(t, "text", response.Content[0].Type)
			assert.Contains(t, response.Content[0].Text, "Successfully created user story")
			assert.Contains(t, response.Content[0].Text, tt.expected.ReferenceID)

			mockUserStoryService.AssertExpectations(t)
			mockEpicService.AssertExpectations(t)
		})
	}
}

// TestUserStoryHandler_Create_EpicReferenceIDResolution tests epic reference ID resolution
func TestUserStoryHandler_Create_EpicReferenceIDResolution(t *testing.T) {
	mockUserStoryService := &MockUserStoryService{}
	mockEpicService := &MockEpicService{}
	handler := NewUserStoryHandler(mockUserStoryService, mockEpicService)

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	epicID := uuid.New()
	epic := &models.Epic{
		ID:          epicID,
		ReferenceID: "EP-001",
		Title:       "Test Epic",
	}

	expectedUserStory := &models.UserStory{
		ID:          uuid.New(),
		ReferenceID: "US-001",
		Title:       "Test User Story",
		EpicID:      epicID,
		CreatorID:   user.ID,
	}

	tests := []struct {
		name        string
		args        map[string]interface{}
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid epic reference ID resolution",
			args: map[string]interface{}{
				"title":    "Test User Story",
				"epic_id":  "EP-001",
				"priority": 1,
			},
			setupMocks: func() {
				mockEpicService.On("GetEpicByReferenceID", "EP-001").Return(epic, nil).Once()
				mockUserStoryService.On("CreateUserStory", mock.AnythingOfType("service.CreateUserStoryRequest")).Return(expectedUserStory, nil).Once()
			},
			expectError: false,
		},
		{
			name: "invalid epic reference ID",
			args: map[string]interface{}{
				"title":    "Test User Story",
				"epic_id":  "EP-999",
				"priority": 1,
			},
			setupMocks: func() {
				mockEpicService.On("GetEpicByReferenceID", "EP-999").Return(nil, errors.New("epic not found")).Once()
			},
			expectError: true,
			errorMsg:    "Invalid params",
		},
		{
			name: "valid epic UUID",
			args: map[string]interface{}{
				"title":    "Test User Story",
				"epic_id":  epicID.String(),
				"priority": 1,
			},
			setupMocks: func() {
				mockUserStoryService.On("CreateUserStory", mock.AnythingOfType("service.CreateUserStoryRequest")).Return(expectedUserStory, nil).Once()
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			ctx := createContextWithUser(user)
			result, err := handler.Create(ctx, tt.args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockUserStoryService.AssertExpectations(t)
			mockEpicService.AssertExpectations(t)
		})
	}
}

// TestUserStoryHandler_Create_ServiceErrors tests user story creation service layer errors
func TestUserStoryHandler_Create_ServiceErrors(t *testing.T) {
	mockUserStoryService := &MockUserStoryService{}
	mockEpicService := &MockEpicService{}
	handler := NewUserStoryHandler(mockUserStoryService, mockEpicService)

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	args := map[string]interface{}{
		"title":    "Test User Story",
		"epic_id":  uuid.New().String(),
		"priority": 1,
	}

	// Setup mock to return error
	mockUserStoryService.On("CreateUserStory", mock.AnythingOfType("service.CreateUserStoryRequest")).Return(nil, errors.New("database error")).Once()

	ctx := createContextWithUser(user)
	_, err := handler.Create(ctx, args)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Internal error")
	mockUserStoryService.AssertExpectations(t)
}

// TestUserStoryHandler_Update_ValidParameters tests user story updates with valid parameters
func TestUserStoryHandler_Update_ValidParameters(t *testing.T) {
	mockUserStoryService := &MockUserStoryService{}
	mockEpicService := &MockEpicService{}
	handler := NewUserStoryHandler(mockUserStoryService, mockEpicService)

	userStoryID := uuid.New()
	expectedUserStory := &models.UserStory{
		ID:          userStoryID,
		ReferenceID: "US-001",
		Title:       "Updated User Story",
		Priority:    models.PriorityHigh,
		Status:      models.UserStoryStatusInProgress,
	}

	tests := []struct {
		name     string
		args     map[string]interface{}
		expected *models.UserStory
	}{
		{
			name: "update with UUID",
			args: map[string]interface{}{
				"user_story_id": userStoryID.String(),
				"title":         "Updated User Story",
			},
			expected: expectedUserStory,
		},
		{
			name: "update with all parameters",
			args: map[string]interface{}{
				"user_story_id": userStoryID.String(),
				"title":         "Updated User Story",
				"description":   "Updated Description",
				"priority":      2,
				"assignee_id":   uuid.New().String(),
			},
			expected: expectedUserStory,
		},
		{
			name: "update with empty assignee_id (unassign)",
			args: map[string]interface{}{
				"user_story_id": userStoryID.String(),
				"assignee_id":   "",
			},
			expected: expectedUserStory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			mockUserStoryService.On("UpdateUserStory", userStoryID, mock.AnythingOfType("service.UpdateUserStoryRequest")).Return(tt.expected, nil).Once()

			result, err := handler.Update(context.Background(), tt.args)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Verify response format
			response, ok := result.(*types.ToolResponse)
			assert.True(t, ok)
			assert.Len(t, response.Content, 2) // Message + data
			assert.Equal(t, "text", response.Content[0].Type)
			assert.Contains(t, response.Content[0].Text, "Successfully updated user story")
			assert.Contains(t, response.Content[0].Text, tt.expected.ReferenceID)

			mockUserStoryService.AssertExpectations(t)
		})
	}
}

// TestUserStoryHandler_Update_ReferenceIDResolution tests user story updates with reference ID resolution
func TestUserStoryHandler_Update_ReferenceIDResolution(t *testing.T) {
	mockUserStoryService := &MockUserStoryService{}
	mockEpicService := &MockEpicService{}
	handler := NewUserStoryHandler(mockUserStoryService, mockEpicService)

	userStoryID := uuid.New()
	expectedUserStory := &models.UserStory{
		ID:          userStoryID,
		ReferenceID: "US-001",
		Title:       "Updated User Story",
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
				"user_story_id": "US-001",
				"title":         "Updated User Story",
			},
			setupMocks: func() {
				mockUserStoryService.On("GetUserStoryByReferenceID", "US-001").Return(expectedUserStory, nil).Once()
				mockUserStoryService.On("UpdateUserStory", userStoryID, mock.AnythingOfType("service.UpdateUserStoryRequest")).Return(expectedUserStory, nil).Once()
			},
			expectError: false,
		},
		{
			name: "invalid reference ID",
			args: map[string]interface{}{
				"user_story_id": "US-999",
				"title":         "Updated User Story",
			},
			setupMocks: func() {
				mockUserStoryService.On("GetUserStoryByReferenceID", "US-999").Return(nil, errors.New("user story not found")).Once()
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

			mockUserStoryService.AssertExpectations(t)
		})
	}
}

// TestUserStoryHandler_Update_ServiceErrors tests user story update service layer errors
func TestUserStoryHandler_Update_ServiceErrors(t *testing.T) {
	mockUserStoryService := &MockUserStoryService{}
	mockEpicService := &MockEpicService{}
	handler := NewUserStoryHandler(mockUserStoryService, mockEpicService)

	userStoryID := uuid.New()
	args := map[string]interface{}{
		"user_story_id": userStoryID.String(),
		"title":         "Updated User Story",
	}

	// Setup mock to return error
	mockUserStoryService.On("UpdateUserStory", userStoryID, mock.AnythingOfType("service.UpdateUserStoryRequest")).Return(nil, errors.New("database error")).Once()

	_, err := handler.Update(context.Background(), args)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Internal error")
	mockUserStoryService.AssertExpectations(t)
}

// TestUserStoryHandler_HandleTool_ErrorHandling tests error handling in HandleTool method
func TestUserStoryHandler_HandleTool_ErrorHandling(t *testing.T) {
	mockUserStoryService := &MockUserStoryService{}
	mockEpicService := &MockEpicService{}
	handler := NewUserStoryHandler(mockUserStoryService, mockEpicService)

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
			name:        "create_user_story with invalid args",
			toolName:    "create_user_story",
			args:        map[string]interface{}{},
			expectError: "Internal error",
		},
		{
			name:        "update_user_story with invalid args",
			toolName:    "update_user_story",
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
