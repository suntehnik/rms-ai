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

// MockPromptService is a mock implementation of PromptServiceInterface
type MockPromptService struct {
	mock.Mock
}

func (m *MockPromptService) Create(ctx context.Context, req *service.CreatePromptRequest, creatorID uuid.UUID) (*models.Prompt, error) {
	args := m.Called(ctx, req, creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}

func (m *MockPromptService) Update(ctx context.Context, id uuid.UUID, req *service.UpdatePromptRequest) (*models.Prompt, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}

func (m *MockPromptService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPromptService) Activate(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPromptService) GetByReferenceID(ctx context.Context, referenceID string) (*models.Prompt, error) {
	args := m.Called(ctx, referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}

func (m *MockPromptService) GetActive(ctx context.Context) (*models.Prompt, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}

func (m *MockPromptService) List(ctx context.Context, limit, offset int, creatorID *uuid.UUID) ([]*models.Prompt, int64, error) {
	args := m.Called(ctx, limit, offset, creatorID)
	return args.Get(0).([]*models.Prompt), args.Get(1).(int64), args.Error(2)
}

// Implement other required methods to satisfy the interface
func (m *MockPromptService) GetByID(ctx context.Context, id uuid.UUID) (*models.Prompt, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}

func (m *MockPromptService) GetByName(ctx context.Context, name string) (*models.Prompt, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}

func (m *MockPromptService) GetMCPPromptDescriptors(ctx context.Context) ([]*models.MCPPromptDescriptor, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.MCPPromptDescriptor), args.Error(1)
}

func (m *MockPromptService) GetMCPPromptDefinition(ctx context.Context, name string) (*models.MCPPromptDefinition, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MCPPromptDefinition), args.Error(1)
}

func TestPromptHandler_GetSupportedTools(t *testing.T) {
	handler := NewPromptHandler(nil)
	tools := handler.GetSupportedTools()

	expected := []string{
		"create_prompt",
		"update_prompt",
		"delete_prompt",
		"activate_prompt",
		"list_prompts",
		"get_active_prompt",
	}
	assert.Equal(t, expected, tools)
}

func TestPromptHandler_HandleTool(t *testing.T) {
	mockService := &MockPromptService{}
	handler := NewPromptHandler(mockService)

	promptUuid := uuid.New()
	promptDescription := "test description"
	prompt := models.Prompt{
		ID:          promptUuid,
		Name:        "Test Name",
		Title:       "Test title",
		Content:     "test content",
		Description: &promptDescription,
	}

	tests := []struct {
		name        string
		toolName    string
		expectError bool
		arguments   map[string]interface{}
		setupMock   func()
	}{
		{
			name:        "valid create_prompt tool",
			toolName:    "create_prompt",
			expectError: false,
			arguments: map[string]interface{}{
				"name":        prompt.Name,
				"title":       prompt.Title,
				"content":     prompt.Content,
				"description": prompt.Description,
			},
			setupMock: func() {
				mockService.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(&prompt, nil)
			},
		},
		{
			name:        "valid update_prompt tool",
			toolName:    "update_prompt",
			expectError: false,
			arguments: map[string]interface{}{
				"prompt_id":   promptUuid.String(),
				"name":        prompt.Name,
				"title":       prompt.Title,
				"content":     prompt.Content,
				"description": prompt.Description,
			},
			setupMock: func() {
				mockService.On("Update", mock.Anything, prompt.ID, mock.Anything).Return(&prompt, nil)
			},
		},
		{
			name:        "valid delete_prompt tool",
			toolName:    "delete_prompt",
			expectError: false,
			arguments: map[string]interface{}{
				"prompt_id": promptUuid.String(),
			},
			setupMock: func() {
				mockService.On("Delete", mock.Anything, prompt.ID).Return(nil)
			},
		},
		{
			name:        "valid activate_prompt tool",
			toolName:    "activate_prompt",
			expectError: false,
			arguments: map[string]interface{}{
				"prompt_id": promptUuid.String(),
			},
			setupMock: func() {
				mockService.On("Activate", mock.Anything, prompt.ID).Return(nil)
			},
		},
		{
			name:        "valid list_prompts tool",
			toolName:    "list_prompts",
			expectError: false, // List prompts works with empty args
			setupMock: func() {
				mockService.On("List", mock.Anything, 50, 0, (*uuid.UUID)(nil)).Return([]*models.Prompt{}, int64(0), nil).Once()
			},
			arguments: map[string]interface{}{},
		},
		{
			name:        "valid get_active_prompt tool",
			toolName:    "get_active_prompt",
			expectError: true, // Will error due to no active prompt
			setupMock: func() {
				mockService.On("GetActive", mock.Anything).Return(nil, service.ErrNotFound).Once()
			},
		},
		{
			name:        "invalid tool name",
			toolName:    "invalid_tool",
			expectError: true,
			setupMock:   func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			_, err := handler.HandleTool(createContextWithPromptUser(&models.User{
				Username: "admin",
				Role:     models.RoleAdministrator,
			}), tt.toolName, tt.arguments)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestNewPromptHandler(t *testing.T) {
	mockService := &MockPromptService{}
	handler := NewPromptHandler(mockService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.promptService)
}

// Helper function to create a context with user
func createContextWithPromptUser(user *models.User) context.Context {
	ginCtx := &gin.Context{}
	ginCtx.Set("user", user)

	ctx := context.WithValue(context.Background(), "gin_context", ginCtx)
	return ctx
}

// TestPromptHandler_Create_ValidParameters tests prompt creation with valid parameters
func TestPromptHandler_Create_ValidParameters(t *testing.T) {
	mockService := &MockPromptService{}
	handler := NewPromptHandler(mockService)

	// Create test admin user
	adminUser := &models.User{
		ID:       uuid.New(),
		Username: "admin",
		Email:    "admin@example.com",
		Role:     models.RoleAdministrator,
	}

	// Create test prompt
	expectedPrompt := &models.Prompt{
		ID:          uuid.New(),
		ReferenceID: "PROMPT-001",
		Name:        "test-prompt",
		Title:       "Test Prompt",
		Description: stringPtr("Test Description"),
		Content:     "Test content",
		CreatorID:   adminUser.ID,
		IsActive:    false,
	}

	tests := []struct {
		name     string
		args     map[string]interface{}
		expected *models.Prompt
	}{
		{
			name: "create prompt with all parameters",
			args: map[string]interface{}{
				"name":        "test-prompt",
				"title":       "Test Prompt",
				"description": "Test Description",
				"content":     "Test content",
			},
			expected: expectedPrompt,
		},
		{
			name: "create prompt with minimal parameters",
			args: map[string]interface{}{
				"name":    "minimal-prompt",
				"title":   "Minimal Prompt",
				"content": "Minimal content",
			},
			expected: expectedPrompt,
		},
		{
			name: "create prompt with role parameter",
			args: map[string]interface{}{
				"name":    "role-prompt",
				"title":   "Role Prompt",
				"content": "Role content",
				"role":    "assistant",
			},
			expected: expectedPrompt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			mockService.On("Create", mock.Anything, mock.AnythingOfType("*service.CreatePromptRequest"), adminUser.ID).Return(tt.expected, nil).Once()

			ctx := createContextWithPromptUser(adminUser)
			result, err := handler.Create(ctx, tt.args)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Verify response format
			response, ok := result.(*types.ToolResponse)
			assert.True(t, ok)
			assert.Len(t, response.Content, 2) // Message + data
			assert.Equal(t, "text", response.Content[0].Type)
			assert.Contains(t, response.Content[0].Text, "Successfully created prompt")
			assert.Contains(t, response.Content[0].Text, tt.expected.ReferenceID)

			mockService.AssertExpectations(t)
		})
	}
}

// TestPromptHandler_Create_InsufficientPermissions tests prompt creation with insufficient permissions
func TestPromptHandler_Create_InsufficientPermissions(t *testing.T) {
	mockService := &MockPromptService{}
	handler := NewPromptHandler(mockService)

	// Create test non-admin user
	user := &models.User{
		ID:       uuid.New(),
		Username: "user",
		Email:    "user@example.com",
		Role:     models.RoleUser,
	}

	args := map[string]interface{}{
		"name":    "test-prompt",
		"title":   "Test Prompt",
		"content": "Test content",
	}

	ctx := createContextWithPromptUser(user)
	_, err := handler.Create(ctx, args)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Insufficient permissions: Administrator role required")
}

// TestPromptHandler_Create_ValidationErrors tests prompt creation with validation errors
func TestPromptHandler_Create_ValidationErrors(t *testing.T) {
	mockService := &MockPromptService{}
	handler := NewPromptHandler(mockService)

	adminUser := &models.User{
		ID:       uuid.New(),
		Username: "admin",
		Email:    "admin@example.com",
		Role:     models.RoleAdministrator,
	}

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError string
	}{
		{
			name:        "missing name",
			args:        map[string]interface{}{"title": "Test", "content": "Test"},
			expectError: "Invalid params",
		},
		{
			name:        "empty name",
			args:        map[string]interface{}{"name": "", "title": "Test", "content": "Test"},
			expectError: "Invalid params",
		},
		{
			name:        "missing title",
			args:        map[string]interface{}{"name": "test", "content": "Test"},
			expectError: "Invalid params",
		},
		{
			name:        "empty title",
			args:        map[string]interface{}{"name": "test", "title": "", "content": "Test"},
			expectError: "Invalid params",
		},
		{
			name:        "missing content",
			args:        map[string]interface{}{"name": "test", "title": "Test"},
			expectError: "Invalid params",
		},
		{
			name:        "empty content",
			args:        map[string]interface{}{"name": "test", "title": "Test", "content": ""},
			expectError: "Invalid params",
		},
		{
			name:        "invalid role",
			args:        map[string]interface{}{"name": "test", "title": "Test", "content": "Test", "role": "system"},
			expectError: "Invalid params",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createContextWithPromptUser(adminUser)
			_, err := handler.Create(ctx, tt.args)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

// TestPromptHandler_Create_DuplicateEntry tests prompt creation with duplicate name
func TestPromptHandler_Create_DuplicateEntry(t *testing.T) {
	mockService := &MockPromptService{}
	handler := NewPromptHandler(mockService)

	adminUser := &models.User{
		ID:       uuid.New(),
		Username: "admin",
		Email:    "admin@example.com",
		Role:     models.RoleAdministrator,
	}

	args := map[string]interface{}{
		"name":    "duplicate-prompt",
		"title":   "Duplicate Prompt",
		"content": "Test content",
	}

	// Setup mock to return duplicate entry error
	mockService.On("Create", mock.Anything, mock.AnythingOfType("*service.CreatePromptRequest"), adminUser.ID).Return(nil, service.ErrDuplicateEntry).Once()

	ctx := createContextWithPromptUser(adminUser)
	_, err := handler.Create(ctx, args)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Prompt with this name already exists")
	mockService.AssertExpectations(t)
}

// TestPromptHandler_Update_ValidParameters tests prompt updates with valid parameters
func TestPromptHandler_Update_ValidParameters(t *testing.T) {
	mockService := &MockPromptService{}
	handler := NewPromptHandler(mockService)

	adminUser := &models.User{
		ID:       uuid.New(),
		Username: "admin",
		Email:    "admin@example.com",
		Role:     models.RoleAdministrator,
	}

	promptID := uuid.New()
	expectedPrompt := &models.Prompt{
		ID:          promptID,
		ReferenceID: "PROMPT-001",
		Name:        "updated-prompt",
		Title:       "Updated Prompt",
		Content:     "Updated content",
	}

	tests := []struct {
		name     string
		args     map[string]interface{}
		expected *models.Prompt
	}{
		{
			name: "update with UUID",
			args: map[string]interface{}{
				"prompt_id": promptID.String(),
				"title":     "Updated Prompt",
			},
			expected: expectedPrompt,
		},
		{
			name: "update with all parameters",
			args: map[string]interface{}{
				"prompt_id":   promptID.String(),
				"title":       "Updated Prompt",
				"description": "Updated Description",
				"content":     "Updated content",
			},
			expected: expectedPrompt,
		},
		{
			name: "update with role parameter",
			args: map[string]interface{}{
				"prompt_id": promptID.String(),
				"title":     "Updated Prompt",
				"role":      "user",
			},
			expected: expectedPrompt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			mockService.On("Update", mock.Anything, promptID, mock.AnythingOfType("*service.UpdatePromptRequest")).Return(tt.expected, nil).Once()

			ctx := createContextWithPromptUser(adminUser)
			result, err := handler.Update(ctx, tt.args)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Verify response format
			response, ok := result.(*types.ToolResponse)
			assert.True(t, ok)
			assert.Len(t, response.Content, 1) // Message only
			assert.Equal(t, "text", response.Content[0].Type)
			assert.Contains(t, response.Content[0].Text, "Successfully updated prompt")
			assert.Contains(t, response.Content[0].Text, tt.expected.ReferenceID)

			mockService.AssertExpectations(t)
		})
	}
}

// TestPromptHandler_Update_ReferenceIDResolution tests prompt updates with reference ID resolution
func TestPromptHandler_Update_ReferenceIDResolution(t *testing.T) {
	mockService := &MockPromptService{}
	handler := NewPromptHandler(mockService)

	adminUser := &models.User{
		ID:       uuid.New(),
		Username: "admin",
		Email:    "admin@example.com",
		Role:     models.RoleAdministrator,
	}

	promptID := uuid.New()
	expectedPrompt := &models.Prompt{
		ID:          promptID,
		ReferenceID: "PROMPT-001",
		Title:       "Updated Prompt",
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
				"prompt_id": "PROMPT-001",
				"title":     "Updated Prompt",
			},
			setupMocks: func() {
				mockService.On("GetByReferenceID", mock.Anything, "PROMPT-001").Return(expectedPrompt, nil).Once()
				mockService.On("Update", mock.Anything, promptID, mock.AnythingOfType("*service.UpdatePromptRequest")).Return(expectedPrompt, nil).Once()
			},
			expectError: false,
		},
		{
			name: "invalid reference ID",
			args: map[string]interface{}{
				"prompt_id": "PROMPT-999",
				"title":     "Updated Prompt",
			},
			setupMocks: func() {
				mockService.On("GetByReferenceID", mock.Anything, "PROMPT-999").Return(nil, service.ErrNotFound).Once()
			},
			expectError: true,
			errorMsg:    "Prompt not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			ctx := createContextWithPromptUser(adminUser)
			result, err := handler.Update(ctx, tt.args)

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

// TestPromptHandler_Delete_ValidParameters tests prompt deletion with valid parameters
func TestPromptHandler_Delete_ValidParameters(t *testing.T) {
	mockService := &MockPromptService{}
	handler := NewPromptHandler(mockService)

	adminUser := &models.User{
		ID:       uuid.New(),
		Username: "admin",
		Email:    "admin@example.com",
		Role:     models.RoleAdministrator,
	}

	promptID := uuid.New()
	prompt := &models.Prompt{
		ID:          promptID,
		ReferenceID: "PROMPT-001",
	}

	tests := []struct {
		name       string
		args       map[string]interface{}
		setupMocks func()
	}{
		{
			name: "delete with UUID",
			args: map[string]interface{}{
				"prompt_id": promptID.String(),
			},
			setupMocks: func() {
				mockService.On("Delete", mock.Anything, promptID).Return(nil).Once()
			},
		},
		{
			name: "delete with reference ID",
			args: map[string]interface{}{
				"prompt_id": "PROMPT-001",
			},
			setupMocks: func() {
				mockService.On("GetByReferenceID", mock.Anything, "PROMPT-001").Return(prompt, nil).Once()
				mockService.On("Delete", mock.Anything, promptID).Return(nil).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			ctx := createContextWithPromptUser(adminUser)
			result, err := handler.Delete(ctx, tt.args)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Verify response format
			response, ok := result.(*types.ToolResponse)
			assert.True(t, ok)
			assert.Len(t, response.Content, 1) // Message only
			assert.Equal(t, "text", response.Content[0].Type)
			assert.Contains(t, response.Content[0].Text, "Successfully deleted prompt")

			mockService.AssertExpectations(t)
		})
	}
}

// TestPromptHandler_Activate_ValidParameters tests prompt activation with valid parameters
func TestPromptHandler_Activate_ValidParameters(t *testing.T) {
	mockService := &MockPromptService{}
	handler := NewPromptHandler(mockService)

	adminUser := &models.User{
		ID:       uuid.New(),
		Username: "admin",
		Email:    "admin@example.com",
		Role:     models.RoleAdministrator,
	}

	promptID := uuid.New()
	prompt := &models.Prompt{
		ID:          promptID,
		ReferenceID: "PROMPT-001",
	}

	tests := []struct {
		name       string
		args       map[string]interface{}
		setupMocks func()
	}{
		{
			name: "activate with UUID",
			args: map[string]interface{}{
				"prompt_id": promptID.String(),
			},
			setupMocks: func() {
				mockService.On("Activate", mock.Anything, promptID).Return(nil).Once()
			},
		},
		{
			name: "activate with reference ID",
			args: map[string]interface{}{
				"prompt_id": "PROMPT-001",
			},
			setupMocks: func() {
				mockService.On("GetByReferenceID", mock.Anything, "PROMPT-001").Return(prompt, nil).Once()
				mockService.On("Activate", mock.Anything, promptID).Return(nil).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			ctx := createContextWithPromptUser(adminUser)
			result, err := handler.Activate(ctx, tt.args)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Verify response format
			response, ok := result.(*types.ToolResponse)
			assert.True(t, ok)
			assert.Len(t, response.Content, 1) // Message only
			assert.Equal(t, "text", response.Content[0].Type)
			assert.Contains(t, response.Content[0].Text, "Successfully activated prompt")

			mockService.AssertExpectations(t)
		})
	}
}

// TestPromptHandler_List_ValidParameters tests prompt listing with valid parameters
func TestPromptHandler_List_ValidParameters(t *testing.T) {
	mockService := &MockPromptService{}
	handler := NewPromptHandler(mockService)

	prompts := []*models.Prompt{
		{
			ID:          uuid.New(),
			ReferenceID: "PROMPT-001",
			Name:        "prompt1",
			Title:       "Prompt 1",
		},
		{
			ID:          uuid.New(),
			ReferenceID: "PROMPT-002",
			Name:        "prompt2",
			Title:       "Prompt 2",
		},
	}

	tests := []struct {
		name       string
		args       map[string]interface{}
		setupMocks func()
	}{
		{
			name: "list with default parameters",
			args: map[string]interface{}{},
			setupMocks: func() {
				mockService.On("List", mock.Anything, 50, 0, (*uuid.UUID)(nil)).Return(prompts, int64(2), nil).Once()
			},
		},
		{
			name: "list with custom pagination",
			args: map[string]interface{}{
				"limit":  10,
				"offset": 5,
			},
			setupMocks: func() {
				mockService.On("List", mock.Anything, 10, 5, (*uuid.UUID)(nil)).Return(prompts, int64(2), nil).Once()
			},
		},
		{
			name: "list with creator filter",
			args: map[string]interface{}{
				"creator_id": uuid.New().String(),
			},
			setupMocks: func() {
				mockService.On("List", mock.Anything, 50, 0, mock.AnythingOfType("*uuid.UUID")).Return(prompts, int64(2), nil).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			result, err := handler.List(context.Background(), tt.args)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Verify response format
			response, ok := result.(*types.ToolResponse)
			assert.True(t, ok)
			assert.Len(t, response.Content, 2) // Message + data
			assert.Equal(t, "text", response.Content[0].Type)
			assert.Contains(t, response.Content[0].Text, "Found 2 prompts")

			mockService.AssertExpectations(t)
		})
	}
}

// TestPromptHandler_GetActive_ValidParameters tests getting active prompt
func TestPromptHandler_GetActive_ValidParameters(t *testing.T) {
	mockService := &MockPromptService{}
	handler := NewPromptHandler(mockService)

	activePrompt := &models.Prompt{
		ID:          uuid.New(),
		ReferenceID: "PROMPT-001",
		Name:        "active-prompt",
		Title:       "Active Prompt",
		IsActive:    true,
	}

	tests := []struct {
		name       string
		setupMocks func()
		expectErr  bool
		errorMsg   string
	}{
		{
			name: "get active prompt success",
			setupMocks: func() {
				mockService.On("GetActive", mock.Anything).Return(activePrompt, nil).Once()
			},
			expectErr: false,
		},
		{
			name: "no active prompt found",
			setupMocks: func() {
				mockService.On("GetActive", mock.Anything).Return(nil, service.ErrNotFound).Once()
			},
			expectErr: true,
			errorMsg:  "No active prompt found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			result, err := handler.GetActive(context.Background(), map[string]interface{}{})

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify response format
				response, ok := result.(*types.ToolResponse)
				assert.True(t, ok)
				assert.Len(t, response.Content, 2) // Message + data
				assert.Equal(t, "text", response.Content[0].Type)
				assert.Contains(t, response.Content[0].Text, "Active prompt:")
				assert.Contains(t, response.Content[0].Text, activePrompt.Title)
			}

			mockService.AssertExpectations(t)
		})
	}
}

// TestPromptHandler_ContextErrors tests context-related errors
func TestPromptHandler_ContextErrors(t *testing.T) {
	mockService := &MockPromptService{}
	handler := NewPromptHandler(mockService)

	args := map[string]interface{}{
		"name":    "test-prompt",
		"title":   "Test Prompt",
		"content": "Test content",
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

// TestPromptHandler_ServiceErrors tests service layer errors
func TestPromptHandler_ServiceErrors(t *testing.T) {
	mockService := &MockPromptService{}
	handler := NewPromptHandler(mockService)

	adminUser := &models.User{
		ID:       uuid.New(),
		Username: "admin",
		Email:    "admin@example.com",
		Role:     models.RoleAdministrator,
	}

	args := map[string]interface{}{
		"name":    "test-prompt",
		"title":   "Test Prompt",
		"content": "Test content",
	}

	// Setup mock to return error
	mockService.On("Create", mock.Anything, mock.AnythingOfType("*service.CreatePromptRequest"), adminUser.ID).Return(nil, errors.New("database error")).Once()

	ctx := createContextWithPromptUser(adminUser)
	_, err := handler.Create(ctx, args)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Internal error")
	mockService.AssertExpectations(t)
}
