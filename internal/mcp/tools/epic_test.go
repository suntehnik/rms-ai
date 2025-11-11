package tools

import (
	"context"
	"errors"
	"strings"
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

func (m *MockEpicService) GetEpicWithCompleteHierarchy(id uuid.UUID) (*models.Epic, error) {
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

	expected := []string{"create_epic", "update_epic", "list_epics", "epic_hierarchy"}
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

// TestEpicHandler_formatTree tests the formatTree method with complete hierarchy
func TestEpicHandler_formatTree(t *testing.T) {
	handler := NewEpicHandler(nil, nil)

	tests := []struct {
		name     string
		epic     *models.Epic
		expected []string // Expected strings to be present in output
	}{
		{
			name: "complete hierarchy",
			epic: &models.Epic{
				ReferenceID: "EP-001",
				Title:       "Test Epic",
				Status:      models.EpicStatusBacklog,
				Priority:    models.PriorityHigh,
				UserStories: []models.UserStory{
					{
						ReferenceID: "US-001",
						Title:       "User Story 1",
						Status:      models.UserStoryStatusBacklog,
						Priority:    models.PriorityHigh,
						Requirements: []models.Requirement{
							{
								ReferenceID: "REQ-001",
								Title:       "Requirement 1",
								Status:      models.RequirementStatusDraft,
								Priority:    models.PriorityHigh,
							},
						},
						AcceptanceCriteria: []models.AcceptanceCriteria{
							{
								ReferenceID: "AC-001",
								Description: "Acceptance criteria description",
							},
						},
					},
				},
			},
			expected: []string{
				"EP-001 [Backlog] [P2] Test Epic",
				"└─┬ US-001 [Backlog] [P2] User Story 1", // Last (and only) user story uses └─┬
				"├── REQ-001 [Draft] [P2] Requirement 1",
				"└── AC-001 — Acceptance criteria description",
			},
		},
		{
			name: "empty user stories",
			epic: &models.Epic{
				ReferenceID: "EP-002",
				Title:       "Empty Epic",
				Status:      models.EpicStatusBacklog,
				Priority:    models.PriorityMedium,
				UserStories: []models.UserStory{},
			},
			expected: []string{
				"EP-002 [Backlog] [P3] Empty Epic",
				"└── No steering documents or user stories attached",
			},
		},
		{
			name: "user story with no requirements",
			epic: &models.Epic{
				ReferenceID: "EP-003",
				Title:       "Epic with Empty US",
				Status:      models.EpicStatusInProgress,
				Priority:    models.PriorityCritical,
				UserStories: []models.UserStory{
					{
						ReferenceID:        "US-002",
						Title:              "Empty User Story",
						Status:             models.UserStoryStatusBacklog,
						Priority:           models.PriorityLow,
						Requirements:       []models.Requirement{},
						AcceptanceCriteria: []models.AcceptanceCriteria{},
					},
				},
			},
			expected: []string{
				"EP-003 [In Progress] [P1] Epic with Empty US",
				"└─┬ US-002 [Backlog] [P4] Empty User Story",
				"├── No requirements",
				"└── No acceptance criteria",
			},
		},
		{
			name: "multiple user stories",
			epic: &models.Epic{
				ReferenceID: "EP-004",
				Title:       "Multi US Epic",
				Status:      models.EpicStatusBacklog,
				Priority:    models.PriorityHigh,
				UserStories: []models.UserStory{
					{
						ReferenceID: "US-003",
						Title:       "First Story",
						Status:      models.UserStoryStatusBacklog,
						Priority:    models.PriorityHigh,
						Requirements: []models.Requirement{
							{
								ReferenceID: "REQ-002",
								Title:       "Req 1",
								Status:      models.RequirementStatusActive,
								Priority:    models.PriorityHigh,
							},
						},
						AcceptanceCriteria: []models.AcceptanceCriteria{},
					},
					{
						ReferenceID:        "US-004",
						Title:              "Second Story",
						Status:             models.UserStoryStatusInProgress,
						Priority:           models.PriorityMedium,
						Requirements:       []models.Requirement{},
						AcceptanceCriteria: []models.AcceptanceCriteria{},
					},
				},
			},
			expected: []string{
				"EP-004 [Backlog] [P2] Multi US Epic",
				"├─┬ US-003 [Backlog] [P2] First Story",
				"├── REQ-002 [Active] [P2] Req 1",
				"└─┬ US-004 [In Progress] [P3] Second Story",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := handler.formatTree(tt.epic)

			// Verify all expected strings are present
			for _, expected := range tt.expected {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}

			// Verify tree structure characters are present
			assert.Contains(t, output, "│", "Output should contain vertical line character")
		})
	}
}

// TestEpicHandler_truncateDescription tests the truncateDescription method
func TestEpicHandler_truncateDescription(t *testing.T) {
	handler := NewEpicHandler(nil, nil)

	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{
			name:      "short text no truncation",
			input:     "Short description",
			maxLength: 80,
			expected:  "Short description",
		},
		{
			name:      "long text with truncation",
			input:     "This is a very long description that exceeds the maximum length and should be truncated with ellipsis",
			maxLength: 80,
			expected:  "This is a very long description that exceeds the maximum length and should be...",
		},
		{
			name:      "first sentence extraction",
			input:     "First sentence. Second sentence. Third sentence.",
			maxLength: 80,
			expected:  "First sentence",
		},
		{
			name:      "first sentence with truncation",
			input:     "This is a very long first sentence that exceeds the maximum length and should be truncated. Second sentence.",
			maxLength: 80,
			expected:  "This is a very long first sentence that exceeds the maximum length and should...",
		},
		{
			name:      "UTF-8 Cyrillic characters",
			input:     "Просмотр иерархии эпика для быстрой оценки покрытия задач и статуса без ручного открытия каждой сущности",
			maxLength: 80,
			expected:  "Просмотр иерархии эпика для быстрой оценки покрытия задач и статуса без ручно...",
		},
		{
			name:      "UTF-8 emoji characters",
			input:     "Test with emoji 🚀 and other Unicode characters 中文 that should be counted correctly",
			maxLength: 80,
			expected:  "Test with emoji 🚀 and other Unicode characters 中文 that should be counted corr...",
		},
		{
			name:      "UTF-8 emoji with truncation",
			input:     "Test with emoji 🚀🎉🔥 and other Unicode characters 中文日本語한국어 that should be truncated properly when exceeding max length",
			maxLength: 80,
			expected:  "Test with emoji 🚀🎉🔥 and other Unicode characters 中文日本語한국어 that should be trun...",
		},
		{
			name:      "exactly 80 characters",
			input:     "This description is exactly eighty characters long and should not be truncated!",
			maxLength: 80,
			expected:  "This description is exactly eighty characters long and should not be truncated!",
		},
		{
			name:      "empty string",
			input:     "",
			maxLength: 80,
			expected:  "",
		},
		{
			name:      "single word longer than max",
			input:     "Supercalifragilisticexpialidocious",
			maxLength: 20,
			expected:  "Supercalifragilis...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.truncateDescription(tt.input, tt.maxLength)
			assert.Equal(t, tt.expected, result)

			// Verify UTF-8 character count (not byte count)
			runes := []rune(result)
			assert.LessOrEqual(t, len(runes), tt.maxLength, "Result should not exceed max length in characters")
		})
	}
}

// TestEpicHandler_formatTree_Indentation tests proper indentation and tree characters
func TestEpicHandler_formatTree_Indentation(t *testing.T) {
	handler := NewEpicHandler(nil, nil)

	epic := &models.Epic{
		ReferenceID: "EP-005",
		Title:       "Indentation Test",
		Status:      models.EpicStatusBacklog,
		Priority:    models.PriorityHigh,
		UserStories: []models.UserStory{
			{
				ReferenceID: "US-005",
				Title:       "Story 1",
				Status:      models.UserStoryStatusBacklog,
				Priority:    models.PriorityHigh,
				Requirements: []models.Requirement{
					{
						ReferenceID: "REQ-003",
						Title:       "Req 1",
						Status:      models.RequirementStatusDraft,
						Priority:    models.PriorityHigh,
					},
					{
						ReferenceID: "REQ-004",
						Title:       "Req 2",
						Status:      models.RequirementStatusDraft,
						Priority:    models.PriorityMedium,
					},
				},
				AcceptanceCriteria: []models.AcceptanceCriteria{
					{
						ReferenceID: "AC-002",
						Description: "AC 1",
					},
					{
						ReferenceID: "AC-003",
						Description: "AC 2",
					},
				},
			},
			{
				ReferenceID: "US-006",
				Title:       "Story 2",
				Status:      models.UserStoryStatusBacklog,
				Priority:    models.PriorityMedium,
				Requirements: []models.Requirement{
					{
						ReferenceID: "REQ-005",
						Title:       "Req 3",
						Status:      models.RequirementStatusActive,
						Priority:    models.PriorityLow,
					},
				},
				AcceptanceCriteria: []models.AcceptanceCriteria{
					{
						ReferenceID: "AC-004",
						Description: "AC 3",
					},
				},
			},
		},
	}

	output := handler.formatTree(epic)

	// Verify tree structure for first user story (not last)
	assert.Contains(t, output, "├─┬ US-005", "First user story should use ├─┬")
	assert.Contains(t, output, "│ │", "First user story should have proper child indentation")
	assert.Contains(t, output, "│ ├── REQ-003", "Requirements should be indented under user story")
	assert.Contains(t, output, "│ ├── REQ-004", "Multiple requirements should use ├──")

	// Verify tree structure for last user story
	assert.Contains(t, output, "└─┬ US-006", "Last user story should use └─┬")
	assert.Contains(t, output, "  ├── REQ-005", "Last user story requirements should have different indentation")
	assert.Contains(t, output, "  └── AC-004", "Last AC should use └──")

	// Verify acceptance criteria formatting
	assert.Contains(t, output, "│ ├── AC-002", "First AC should use ├──")
	assert.Contains(t, output, "│ └── AC-003", "Last AC in first story should use └──")
}

// TestEpicHandler_formatTree_EmptyStates tests empty state messages
func TestEpicHandler_formatTree_EmptyStates(t *testing.T) {
	handler := NewEpicHandler(nil, nil)

	tests := []struct {
		name     string
		epic     *models.Epic
		expected []string
	}{
		{
			name: "no user stories",
			epic: &models.Epic{
				ReferenceID: "EP-006",
				Title:       "Empty Epic",
				Status:      models.EpicStatusBacklog,
				Priority:    models.PriorityHigh,
				UserStories: []models.UserStory{},
			},
			expected: []string{
				"└── No steering documents or user stories attached",
			},
		},
		{
			name: "user story with no requirements",
			epic: &models.Epic{
				ReferenceID: "EP-007",
				Title:       "Epic",
				Status:      models.EpicStatusBacklog,
				Priority:    models.PriorityHigh,
				UserStories: []models.UserStory{
					{
						ReferenceID:        "US-007",
						Title:              "Story",
						Status:             models.UserStoryStatusBacklog,
						Priority:           models.PriorityHigh,
						Requirements:       []models.Requirement{},
						AcceptanceCriteria: []models.AcceptanceCriteria{},
					},
				},
			},
			expected: []string{
				"├── No requirements",
				"└── No acceptance criteria",
			},
		},
		{
			name: "user story with requirements but no AC",
			epic: &models.Epic{
				ReferenceID: "EP-008",
				Title:       "Epic",
				Status:      models.EpicStatusBacklog,
				Priority:    models.PriorityHigh,
				UserStories: []models.UserStory{
					{
						ReferenceID: "US-008",
						Title:       "Story",
						Status:      models.UserStoryStatusBacklog,
						Priority:    models.PriorityHigh,
						Requirements: []models.Requirement{
							{
								ReferenceID: "REQ-006",
								Title:       "Req",
								Status:      models.RequirementStatusDraft,
								Priority:    models.PriorityHigh,
							},
						},
						AcceptanceCriteria: []models.AcceptanceCriteria{},
					},
				},
			},
			expected: []string{
				"├── REQ-006",
				"└── No acceptance criteria",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := handler.formatTree(tt.epic)

			for _, expected := range tt.expected {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
		})
	}
}

// TestEpicHandler_GetHierarchy tests the GetHierarchy method
func TestEpicHandler_GetHierarchy(t *testing.T) {
	tests := []struct {
		name          string
		args          map[string]interface{}
		setupMocks    func(*MockEpicService)
		expectError   bool
		errorContains string
	}{
		{
			name: "valid reference ID input",
			args: map[string]interface{}{
				"epic": "EP-001",
			},
			setupMocks: func(mockService *MockEpicService) {
				epic := &models.Epic{
					ID:          uuid.New(),
					ReferenceID: "EP-001",
					Title:       "Test Epic",
					Status:      models.EpicStatusBacklog,
					Priority:    models.PriorityHigh,
					UserStories: []models.UserStory{
						{
							ReferenceID: "US-001",
							Title:       "User Story",
							Status:      models.UserStoryStatusBacklog,
							Priority:    models.PriorityHigh,
							Requirements: []models.Requirement{
								{
									ReferenceID: "REQ-001",
									Title:       "Requirement",
									Status:      models.RequirementStatusDraft,
									Priority:    models.PriorityHigh,
								},
							},
							AcceptanceCriteria: []models.AcceptanceCriteria{
								{
									ReferenceID: "AC-001",
									Description: "Acceptance criteria",
								},
							},
						},
					},
				}
				mockService.On("GetEpicByReferenceID", "EP-001").Return(epic, nil).Once()
				mockService.On("GetEpicWithCompleteHierarchy", epic.ID).Return(epic, nil).Once()
			},
			expectError: false,
		},
		{
			name: "valid UUID input",
			args: map[string]interface{}{
				"epic": "550e8400-e29b-41d4-a716-446655440000",
			},
			setupMocks: func(mockService *MockEpicService) {
				epicID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
				epic := &models.Epic{
					ID:          epicID,
					ReferenceID: "EP-002",
					Title:       "Test Epic",
					Status:      models.EpicStatusBacklog,
					Priority:    models.PriorityHigh,
					UserStories: []models.UserStory{},
				}
				mockService.On("GetEpicWithCompleteHierarchy", epicID).Return(epic, nil).Once()
			},
			expectError: false,
		},
		{
			name: "missing epic argument",
			args: map[string]interface{}{},
			setupMocks: func(mockService *MockEpicService) {
				// No mocks needed
			},
			expectError:   true,
			errorContains: "Invalid params",
		},
		{
			name: "empty epic argument",
			args: map[string]interface{}{
				"epic": "",
			},
			setupMocks: func(mockService *MockEpicService) {
				// No mocks needed
			},
			expectError:   true,
			errorContains: "Invalid params",
		},
		{
			name: "invalid reference ID format",
			args: map[string]interface{}{
				"epic": "INVALID-ID",
			},
			setupMocks: func(mockService *MockEpicService) {
				mockService.On("GetEpicByReferenceID", "INVALID-ID").Return(nil, errors.New("not found")).Once()
			},
			expectError:   true,
			errorContains: "Invalid params",
		},
		{
			name: "epic not found",
			args: map[string]interface{}{
				"epic": "EP-999",
			},
			setupMocks: func(mockService *MockEpicService) {
				mockService.On("GetEpicByReferenceID", "EP-999").Return(nil, service.ErrEpicNotFound).Once()
			},
			expectError:   true,
			errorContains: "Invalid params",
		},
		{
			name: "epic not found by UUID",
			args: map[string]interface{}{
				"epic": "550e8400-e29b-41d4-a716-446655440000",
			},
			setupMocks: func(mockService *MockEpicService) {
				epicID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
				mockService.On("GetEpicWithCompleteHierarchy", epicID).Return(nil, service.ErrEpicNotFound).Once()
			},
			expectError:   true,
			errorContains: "Invalid params",
		},
		{
			name: "service error",
			args: map[string]interface{}{
				"epic": "EP-001",
			},
			setupMocks: func(mockService *MockEpicService) {
				epic := &models.Epic{
					ID:          uuid.New(),
					ReferenceID: "EP-001",
				}
				mockService.On("GetEpicByReferenceID", "EP-001").Return(epic, nil).Once()
				mockService.On("GetEpicWithCompleteHierarchy", epic.ID).Return(nil, errors.New("database error")).Once()
			},
			expectError:   true,
			errorContains: "Internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockEpicService{}
			handler := NewEpicHandler(mockService, nil)

			tt.setupMocks(mockService)

			result, err := handler.GetHierarchy(context.Background(), tt.args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify response format
				response, ok := result.(*types.ToolResponse)
				assert.True(t, ok, "Result should be a ToolResponse")
				assert.NotEmpty(t, response.Content, "Response should have content")
				assert.Equal(t, "text", response.Content[0].Type, "First content should be text")
			}

			mockService.AssertExpectations(t)
		})
	}
}

// TestEpicHandler_GetHierarchy_ResponseFormat tests the JSON-RPC response format
func TestEpicHandler_GetHierarchy_ResponseFormat(t *testing.T) {
	mockService := &MockEpicService{}
	handler := NewEpicHandler(mockService, nil)

	epicID := uuid.New()
	epic := &models.Epic{
		ID:          epicID,
		ReferenceID: "EP-001",
		Title:       "Test Epic",
		Status:      models.EpicStatusBacklog,
		Priority:    models.PriorityHigh,
		UserStories: []models.UserStory{
			{
				ReferenceID: "US-001",
				Title:       "User Story",
				Status:      models.UserStoryStatusBacklog,
				Priority:    models.PriorityHigh,
				Requirements: []models.Requirement{
					{
						ReferenceID: "REQ-001",
						Title:       "Requirement",
						Status:      models.RequirementStatusDraft,
						Priority:    models.PriorityHigh,
					},
				},
				AcceptanceCriteria: []models.AcceptanceCriteria{
					{
						ReferenceID: "AC-001",
						Description: "Acceptance criteria description",
					},
				},
			},
		},
	}

	mockService.On("GetEpicByReferenceID", "EP-001").Return(epic, nil).Once()
	mockService.On("GetEpicWithCompleteHierarchy", epicID).Return(epic, nil).Once()

	args := map[string]interface{}{
		"epic": "EP-001",
	}

	result, err := handler.GetHierarchy(context.Background(), args)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify response structure
	response, ok := result.(*types.ToolResponse)
	assert.True(t, ok, "Result should be a ToolResponse")
	assert.NotEmpty(t, response.Content, "Response should have content")

	// Verify first content is text with the tree output
	assert.Equal(t, "text", response.Content[0].Type)
	assert.NotEmpty(t, response.Content[0].Text)

	// Verify tree output contains expected elements
	treeOutput := response.Content[0].Text
	assert.Contains(t, treeOutput, "EP-001")
	assert.Contains(t, treeOutput, "US-001")
	assert.Contains(t, treeOutput, "REQ-001")
	assert.Contains(t, treeOutput, "AC-001")
	assert.Contains(t, treeOutput, "│")
	assert.Contains(t, treeOutput, "├──")
	assert.Contains(t, treeOutput, "└──")

	mockService.AssertExpectations(t)
}

// TestEpicHandler_GetHierarchy_Integration tests integration with existing EpicHandler
func TestEpicHandler_GetHierarchy_Integration(t *testing.T) {
	mockService := &MockEpicService{}
	mockUserService := &MockUserService{}
	handler := NewEpicHandler(mockService, mockUserService)

	// Verify GetHierarchy is in supported tools
	supportedTools := handler.GetSupportedTools()
	assert.Contains(t, supportedTools, "epic_hierarchy", "epic_hierarchy should be in supported tools")

	// Verify HandleTool routes to GetHierarchy
	epicID := uuid.New()
	epic := &models.Epic{
		ID:          epicID,
		ReferenceID: "EP-001",
		Title:       "Test Epic",
		Status:      models.EpicStatusBacklog,
		Priority:    models.PriorityHigh,
		UserStories: []models.UserStory{},
	}

	mockService.On("GetEpicByReferenceID", "EP-001").Return(epic, nil).Once()
	mockService.On("GetEpicWithCompleteHierarchy", epicID).Return(epic, nil).Once()

	args := map[string]interface{}{
		"epic": "EP-001",
	}

	result, err := handler.HandleTool(context.Background(), "epic_hierarchy", args)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify it returns a proper ToolResponse
	response, ok := result.(*types.ToolResponse)
	assert.True(t, ok)
	assert.NotEmpty(t, response.Content)

	mockService.AssertExpectations(t)
}

// TestEpicHandler_formatSteeringDocument tests the formatSteeringDocument method
func TestEpicHandler_formatSteeringDocument(t *testing.T) {
	handler := NewEpicHandler(nil, nil)

	tests := []struct {
		name     string
		std      models.SteeringDocument
		expected []string
	}{
		{
			name: "steering document without description",
			std: models.SteeringDocument{
				ReferenceID: "STD-001",
				Title:       "Technical Architecture Guidelines",
			},
			expected: []string{
				"├── STD-001 Technical Architecture Guidelines",
			},
		},
		{
			name: "steering document with short description",
			std: models.SteeringDocument{
				ReferenceID: "STD-002",
				Title:       "API Design Standards",
				Description: stringPtr("Short description"),
			},
			expected: []string{
				"├── STD-002 API Design Standards",
				"│   Short description",
			},
		},
		{
			name: "steering document with long description",
			std: models.SteeringDocument{
				ReferenceID: "STD-003",
				Title:       "Code Review Standards",
				Description: stringPtr("This is a very long description that exceeds the maximum length and should be truncated with ellipsis to fit within the 80 character limit"),
			},
			expected: []string{
				"├── STD-003 Code Review Standards",
				"│   This is a very long description that exceeds the maximum length and should be...",
			},
		},
		{
			name: "steering document with Cyrillic description",
			std: models.SteeringDocument{
				ReferenceID: "STD-004",
				Title:       "Руководство по архитектуре",
				Description: stringPtr("Это руководство содержит стандарты и практики для разработки архитектуры системы"),
			},
			expected: []string{
				"├── STD-004 Руководство по архитектуре",
				"│   Это руководство содержит стандарты и практики для разработки архитектуры системы",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var builder strings.Builder
			handler.formatSteeringDocument(&builder, tt.std)
			output := builder.String()

			for _, expected := range tt.expected {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
		})
	}
}

// TestEpicHandler_formatTree_WithSteeringDocuments tests formatTree with steering documents
func TestEpicHandler_formatTree_WithSteeringDocuments(t *testing.T) {
	handler := NewEpicHandler(nil, nil)

	tests := []struct {
		name     string
		epic     *models.Epic
		expected []string
	}{
		{
			name: "epic with only steering documents",
			epic: &models.Epic{
				ReferenceID: "EP-010",
				Title:       "Epic with Steering Docs",
				Status:      models.EpicStatusBacklog,
				Priority:    models.PriorityHigh,
				SteeringDocuments: []models.SteeringDocument{
					{
						ReferenceID: "STD-001",
						Title:       "Technical Guidelines",
					},
					{
						ReferenceID: "STD-002",
						Title:       "API Standards",
					},
				},
				UserStories: []models.UserStory{},
			},
			expected: []string{
				"EP-010 [Backlog] [P2] Epic with Steering Docs",
				"├── STD-001 Technical Guidelines",
				"├── STD-002 API Standards",
			},
		},
		{
			name: "epic with steering documents and user stories",
			epic: &models.Epic{
				ReferenceID: "EP-011",
				Title:       "Complete Epic",
				Status:      models.EpicStatusInProgress,
				Priority:    models.PriorityCritical,
				SteeringDocuments: []models.SteeringDocument{
					{
						ReferenceID: "STD-003",
						Title:       "Code Review Standards",
						Description: stringPtr("Standards for code review process"),
					},
				},
				UserStories: []models.UserStory{
					{
						ReferenceID: "US-010",
						Title:       "User Story 1",
						Status:      models.UserStoryStatusBacklog,
						Priority:    models.PriorityHigh,
						Requirements: []models.Requirement{
							{
								ReferenceID: "REQ-010",
								Title:       "Requirement 1",
								Status:      models.RequirementStatusDraft,
								Priority:    models.PriorityHigh,
							},
						},
						AcceptanceCriteria: []models.AcceptanceCriteria{},
					},
				},
			},
			expected: []string{
				"EP-011 [In Progress] [P1] Complete Epic",
				"├── STD-003 Code Review Standards",
				"│   Standards for code review process",
				"│",
				"└─┬ US-010 [Backlog] [P2] User Story 1",
				"├── REQ-010 [Draft] [P2] Requirement 1",
			},
		},
		{
			name: "epic with no steering documents and no user stories",
			epic: &models.Epic{
				ReferenceID:       "EP-012",
				Title:             "Empty Epic",
				Status:            models.EpicStatusBacklog,
				Priority:          models.PriorityMedium,
				SteeringDocuments: []models.SteeringDocument{},
				UserStories:       []models.UserStory{},
			},
			expected: []string{
				"EP-012 [Backlog] [P3] Empty Epic",
				"└── No steering documents or user stories attached",
			},
		},
		{
			name: "epic with multiple steering documents and user stories",
			epic: &models.Epic{
				ReferenceID: "EP-013",
				Title:       "Multi-entity Epic",
				Status:      models.EpicStatusBacklog,
				Priority:    models.PriorityHigh,
				SteeringDocuments: []models.SteeringDocument{
					{
						ReferenceID: "STD-004",
						Title:       "Architecture Guide",
					},
					{
						ReferenceID: "STD-005",
						Title:       "Security Standards",
					},
				},
				UserStories: []models.UserStory{
					{
						ReferenceID:        "US-011",
						Title:              "Story 1",
						Status:             models.UserStoryStatusBacklog,
						Priority:           models.PriorityHigh,
						Requirements:       []models.Requirement{},
						AcceptanceCriteria: []models.AcceptanceCriteria{},
					},
					{
						ReferenceID:        "US-012",
						Title:              "Story 2",
						Status:             models.UserStoryStatusInProgress,
						Priority:           models.PriorityMedium,
						Requirements:       []models.Requirement{},
						AcceptanceCriteria: []models.AcceptanceCriteria{},
					},
				},
			},
			expected: []string{
				"EP-013 [Backlog] [P2] Multi-entity Epic",
				"├── STD-004 Architecture Guide",
				"├── STD-005 Security Standards",
				"│",
				"├─┬ US-011 [Backlog] [P2] Story 1",
				"└─┬ US-012 [In Progress] [P3] Story 2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := handler.formatTree(tt.epic)

			for _, expected := range tt.expected {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
		})
	}
}

// TestEpicHandler_formatTree_SteeringDocumentDescriptionTruncation tests description truncation
func TestEpicHandler_formatTree_SteeringDocumentDescriptionTruncation(t *testing.T) {
	handler := NewEpicHandler(nil, nil)

	tests := []struct {
		name        string
		description string
		maxLength   int
		shouldTrunc bool
	}{
		{
			name:        "short description no truncation",
			description: "Short description",
			maxLength:   80,
			shouldTrunc: false,
		},
		{
			name:        "long description with truncation",
			description: "This is a very long description that exceeds the maximum length and should be truncated with ellipsis to fit within the character limit",
			maxLength:   80,
			shouldTrunc: true,
		},
		{
			name:        "first sentence extraction",
			description: "First sentence. Second sentence. Third sentence.",
			maxLength:   80,
			shouldTrunc: false,
		},
		{
			name:        "Cyrillic text truncation",
			description: "Это очень длинное описание на русском языке, которое превышает максимальную длину и должно быть обрезано с многоточием",
			maxLength:   80,
			shouldTrunc: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			epic := &models.Epic{
				ReferenceID: "EP-014",
				Title:       "Test Epic",
				Status:      models.EpicStatusBacklog,
				Priority:    models.PriorityHigh,
				SteeringDocuments: []models.SteeringDocument{
					{
						ReferenceID: "STD-006",
						Title:       "Test Document",
						Description: &tt.description,
					},
				},
				UserStories: []models.UserStory{},
			}

			output := handler.formatTree(epic)

			if tt.shouldTrunc {
				assert.Contains(t, output, "...", "Output should contain ellipsis for truncated text")
			}

			// Verify the description line doesn't exceed max length
			lines := strings.Split(output, "\n")
			for _, line := range lines {
				if strings.Contains(line, "│   ") {
					// This is a description line
					descPart := strings.TrimPrefix(line, "│   ")
					runes := []rune(descPart)
					assert.LessOrEqual(t, len(runes), tt.maxLength, "Description should not exceed max length")
				}
			}
		})
	}
}
