package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// MockAcceptanceCriteriaService is a mock implementation of service.AcceptanceCriteriaService
type MockAcceptanceCriteriaService struct {
	mock.Mock
}

func (m *MockAcceptanceCriteriaService) CreateAcceptanceCriteria(req service.CreateAcceptanceCriteriaRequest) (*models.AcceptanceCriteria, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AcceptanceCriteria), args.Error(1)
}

func (m *MockAcceptanceCriteriaService) GetAcceptanceCriteriaByID(id uuid.UUID) (*models.AcceptanceCriteria, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AcceptanceCriteria), args.Error(1)
}

func (m *MockAcceptanceCriteriaService) GetAcceptanceCriteriaByReferenceID(referenceID string) (*models.AcceptanceCriteria, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AcceptanceCriteria), args.Error(1)
}

func (m *MockAcceptanceCriteriaService) UpdateAcceptanceCriteria(id uuid.UUID, req service.UpdateAcceptanceCriteriaRequest) (*models.AcceptanceCriteria, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AcceptanceCriteria), args.Error(1)
}

func (m *MockAcceptanceCriteriaService) DeleteAcceptanceCriteria(id uuid.UUID, force bool) error {
	args := m.Called(id, force)
	return args.Error(0)
}

func (m *MockAcceptanceCriteriaService) ListAcceptanceCriteria(filters service.AcceptanceCriteriaFilters) ([]models.AcceptanceCriteria, int64, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.AcceptanceCriteria), args.Get(1).(int64), args.Error(2)
}

func (m *MockAcceptanceCriteriaService) GetAcceptanceCriteriaByUserStory(userStoryID uuid.UUID, limit, offset int) ([]models.AcceptanceCriteria, int64, error) {
	args := m.Called(userStoryID, limit, offset)
	return args.Get(0).([]models.AcceptanceCriteria), args.Get(1).(int64), args.Error(2)
}

func (m *MockAcceptanceCriteriaService) GetAcceptanceCriteriaByAuthor(authorID uuid.UUID, limit, offset int) ([]models.AcceptanceCriteria, int64, error) {
	args := m.Called(authorID, limit, offset)
	return args.Get(0).([]models.AcceptanceCriteria), args.Get(1).(int64), args.Error(2)
}

func (m *MockAcceptanceCriteriaService) ValidateUserStoryHasAcceptanceCriteria(userStoryID uuid.UUID) error {
	args := m.Called(userStoryID)
	return args.Error(0)
}

// MockUserStoryServiceForAC is a mock implementation of service.UserStoryService for acceptance criteria tests
type MockUserStoryServiceForAC struct {
	mock.Mock
}

func (m *MockUserStoryServiceForAC) GetUserStoryByReferenceID(referenceID string) (*models.UserStory, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryServiceForAC) GetUUIDByReferenceID(referenceID string) (uuid.UUID, error) {
	args := m.Called(referenceID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

// Implement other required methods to satisfy the interface
func (m *MockUserStoryServiceForAC) CreateUserStory(req service.CreateUserStoryRequest) (*models.UserStory, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryServiceForAC) GetUserStoryByID(id uuid.UUID) (*models.UserStory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryServiceForAC) UpdateUserStory(id uuid.UUID, req service.UpdateUserStoryRequest) (*models.UserStory, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryServiceForAC) DeleteUserStory(id uuid.UUID, force bool) error {
	args := m.Called(id, force)
	return args.Error(0)
}

func (m *MockUserStoryServiceForAC) ListUserStories(filters service.UserStoryFilters) ([]models.UserStory, int64, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.UserStory), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserStoryServiceForAC) GetUserStoryWithAcceptanceCriteria(id uuid.UUID) (*models.UserStory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryServiceForAC) GetUserStoryWithRequirements(id uuid.UUID) (*models.UserStory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryServiceForAC) GetUserStoriesByEpic(epicID uuid.UUID) ([]models.UserStory, error) {
	args := m.Called(epicID)
	return args.Get(0).([]models.UserStory), args.Error(1)
}

func (m *MockUserStoryServiceForAC) ChangeUserStoryStatus(id uuid.UUID, newStatus models.UserStoryStatus) (*models.UserStory, error) {
	args := m.Called(id, newStatus)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryServiceForAC) AssignUserStory(id uuid.UUID, assigneeID uuid.UUID) (*models.UserStory, error) {
	args := m.Called(id, assigneeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

// Helper function to create a context with user for acceptance criteria tests
func createACContextWithUser(user *models.User) context.Context {
	ginCtx := &gin.Context{}
	ginCtx.Set("user", user)

	ctx := context.WithValue(context.Background(), "gin_context", ginCtx)
	return ctx
}

func TestAcceptanceCriteriaHandler_GetSupportedTools(t *testing.T) {
	handler := NewAcceptanceCriteriaHandler(nil, nil)
	tools := handler.GetSupportedTools()

	expected := []string{"create_acceptance_criteria"}
	assert.Equal(t, expected, tools)
}

func TestAcceptanceCriteriaHandler_HandleTool(t *testing.T) {
	mockACService := &MockAcceptanceCriteriaService{}
	mockUSService := &MockUserStoryServiceForAC{}
	handler := NewAcceptanceCriteriaHandler(mockACService, mockUSService)

	// Test valid tool
	ctx := context.Background()
	args := map[string]interface{}{}

	result, err := handler.HandleTool(ctx, "create_acceptance_criteria", args)
	// We expect this to fail due to missing arguments, but the tool should be recognized
	assert.Error(t, err)
	assert.Nil(t, result)

	// Test invalid tool
	result, err = handler.HandleTool(ctx, "invalid_tool", args)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Method not found")
}

func TestNewAcceptanceCriteriaHandler(t *testing.T) {
	mockACService := &MockAcceptanceCriteriaService{}
	mockUSService := &MockUserStoryServiceForAC{}
	handler := NewAcceptanceCriteriaHandler(mockACService, mockUSService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockACService, handler.acceptanceCriteriaService)
	assert.Equal(t, mockUSService, handler.userStoryService)
}

// TestAcceptanceCriteriaHandler_Create_ValidParameters tests acceptance criteria creation with valid parameters
func TestAcceptanceCriteriaHandler_Create_ValidParameters(t *testing.T) {
	mockACService := &MockAcceptanceCriteriaService{}
	mockUSService := &MockUserStoryServiceForAC{}
	handler := NewAcceptanceCriteriaHandler(mockACService, mockUSService)

	// Create test user
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	// Create test user story
	userStoryID := uuid.New()
	userStory := &models.UserStory{
		ID:          userStoryID,
		ReferenceID: "US-001",
		Title:       "Test User Story",
		CreatorID:   user.ID,
	}

	// Create expected acceptance criteria
	expectedAC := &models.AcceptanceCriteria{
		ID:          uuid.New(),
		ReferenceID: "AC-001",
		Description: "Test acceptance criteria description",
		UserStoryID: userStoryID,
		AuthorID:    user.ID,
	}

	tests := []struct {
		name        string
		args        map[string]interface{}
		setupMocks  func()
		expectError bool
	}{
		{
			name: "create acceptance criteria with UUID user story ID",
			args: map[string]interface{}{
				"user_story_id": userStoryID.String(),
				"description":   "Test acceptance criteria description",
			},
			setupMocks: func() {
				mockACService.On("CreateAcceptanceCriteria", mock.MatchedBy(func(req service.CreateAcceptanceCriteriaRequest) bool {
					return req.UserStoryID == userStoryID &&
						req.AuthorID == user.ID &&
						req.Description == "Test acceptance criteria description"
				})).Return(expectedAC, nil)
			},
			expectError: false,
		},
		{
			name: "create acceptance criteria with reference ID user story ID",
			args: map[string]interface{}{
				"user_story_id": "US-001",
				"description":   "Test acceptance criteria description",
			},
			setupMocks: func() {
				mockUSService.On("GetUserStoryByReferenceID", "US-001").Return(userStory, nil)
				mockACService.On("CreateAcceptanceCriteria", mock.MatchedBy(func(req service.CreateAcceptanceCriteriaRequest) bool {
					return req.UserStoryID == userStoryID &&
						req.AuthorID == user.ID &&
						req.Description == "Test acceptance criteria description"
				})).Return(expectedAC, nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockACService.ExpectedCalls = nil
			mockUSService.ExpectedCalls = nil

			// Setup mocks
			tt.setupMocks()

			// Create context with user
			ctx := createACContextWithUser(user)

			// Execute
			result, err := handler.Create(ctx, tt.args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			// Verify mocks
			mockACService.AssertExpectations(t)
			mockUSService.AssertExpectations(t)
		})
	}
}

// TestAcceptanceCriteriaHandler_Create_UserStoryIdentifierValidation tests user story identifier validation
func TestAcceptanceCriteriaHandler_Create_UserStoryIdentifierValidation(t *testing.T) {
	mockACService := &MockAcceptanceCriteriaService{}
	mockUSService := &MockUserStoryServiceForAC{}
	handler := NewAcceptanceCriteriaHandler(mockACService, mockUSService)

	// Create test user
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	tests := []struct {
		name          string
		userStoryID   interface{}
		setupMocks    func()
		expectedError string
	}{
		{
			name:          "missing user_story_id",
			userStoryID:   nil,
			setupMocks:    func() {},
			expectedError: "Invalid params",
		},
		{
			name:          "empty user_story_id",
			userStoryID:   "",
			setupMocks:    func() {},
			expectedError: "Invalid params",
		},
		{
			name:        "invalid UUID format",
			userStoryID: "invalid-uuid",
			setupMocks: func() {
				mockUSService.On("GetUserStoryByReferenceID", "invalid-uuid").Return(nil, service.ErrUserStoryNotFound)
			},
			expectedError: "Invalid params",
		},
		{
			name:        "valid reference ID format but not found",
			userStoryID: "US-999",
			setupMocks: func() {
				mockUSService.On("GetUserStoryByReferenceID", "US-999").Return(nil, service.ErrUserStoryNotFound)
			},
			expectedError: "Invalid params",
		},
		{
			name:        "valid UUID but user story not found",
			userStoryID: uuid.New().String(),
			setupMocks: func() {
				mockACService.On("CreateAcceptanceCriteria", mock.AnythingOfType("service.CreateAcceptanceCriteriaRequest")).Return(nil, service.ErrUserStoryNotFound)
			},
			expectedError: "Invalid params",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockACService.ExpectedCalls = nil
			mockUSService.ExpectedCalls = nil

			// Setup mocks
			tt.setupMocks()

			// Create context with user
			ctx := createACContextWithUser(user)

			// Create args
			args := map[string]interface{}{
				"description": "Test description",
			}
			if tt.userStoryID != nil {
				args["user_story_id"] = tt.userStoryID
			}

			// Execute
			result, err := handler.Create(ctx, args)

			// Verify
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedError)

			// Verify mocks
			mockACService.AssertExpectations(t)
			mockUSService.AssertExpectations(t)
		})
	}
}

// TestAcceptanceCriteriaHandler_Create_DescriptionValidation tests description field validation
func TestAcceptanceCriteriaHandler_Create_DescriptionValidation(t *testing.T) {
	mockACService := &MockAcceptanceCriteriaService{}
	mockUSService := &MockUserStoryServiceForAC{}
	handler := NewAcceptanceCriteriaHandler(mockACService, mockUSService)

	// Create test user
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	userStoryID := uuid.New()

	tests := []struct {
		name          string
		description   interface{}
		expectedError string
	}{
		{
			name:          "missing description",
			description:   nil,
			expectedError: "Invalid params",
		},
		{
			name:          "empty description",
			description:   "",
			expectedError: "Invalid params",
		},
		{
			name:          "description exceeds max length",
			description:   string(make([]byte, 50001)), // 50001 characters
			expectedError: "Invalid params",
		},
		{
			name:          "non-string description",
			description:   123,
			expectedError: "Invalid params",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockACService.ExpectedCalls = nil
			mockUSService.ExpectedCalls = nil

			// Create context with user
			ctx := createACContextWithUser(user)

			// Create args
			args := map[string]interface{}{
				"user_story_id": userStoryID.String(),
			}
			if tt.description != nil {
				args["description"] = tt.description
			}

			// Execute
			result, err := handler.Create(ctx, args)

			// Verify
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedError)

			// Verify mocks
			mockACService.AssertExpectations(t)
			mockUSService.AssertExpectations(t)
		})
	}
}

// TestAcceptanceCriteriaHandler_Create_AuthenticationErrors tests authentication context extraction and error scenarios
func TestAcceptanceCriteriaHandler_Create_AuthenticationErrors(t *testing.T) {
	mockACService := &MockAcceptanceCriteriaService{}
	mockUSService := &MockUserStoryServiceForAC{}
	handler := NewAcceptanceCriteriaHandler(mockACService, mockUSService)

	userStoryID := uuid.New()
	validArgs := map[string]interface{}{
		"user_story_id": userStoryID.String(),
		"description":   "Test description",
	}

	tests := []struct {
		name          string
		context       context.Context
		setupMocks    func()
		expectedError string
	}{
		{
			name:          "missing gin context",
			context:       context.Background(),
			setupMocks:    func() {},
			expectedError: "Internal error",
		},
		{
			name: "gin context without user",
			context: func() context.Context {
				ginCtx := &gin.Context{}
				return context.WithValue(context.Background(), "gin_context", ginCtx)
			}(),
			setupMocks:    func() {},
			expectedError: "Internal error",
		},
		{
			name: "service returns user not found error",
			context: createACContextWithUser(&models.User{
				ID:       uuid.New(),
				Username: "testuser",
				Email:    "test@example.com",
				Role:     models.RoleUser,
			}),
			setupMocks: func() {
				mockACService.On("CreateAcceptanceCriteria", mock.AnythingOfType("service.CreateAcceptanceCriteriaRequest")).Return(nil, service.ErrUserNotFound)
			},
			expectedError: "Unauthorized access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockACService.ExpectedCalls = nil
			mockUSService.ExpectedCalls = nil

			// Setup mocks
			tt.setupMocks()

			// Execute
			result, err := handler.Create(tt.context, validArgs)

			// Verify
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedError)

			// Verify mocks
			mockACService.AssertExpectations(t)
			mockUSService.AssertExpectations(t)
		})
	}
}

// TestAcceptanceCriteriaHandler_Create_ServiceErrors tests various service error scenarios
func TestAcceptanceCriteriaHandler_Create_ServiceErrors(t *testing.T) {
	mockACService := &MockAcceptanceCriteriaService{}
	mockUSService := &MockUserStoryServiceForAC{}
	handler := NewAcceptanceCriteriaHandler(mockACService, mockUSService)

	// Create test user
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	userStoryID := uuid.New()
	validArgs := map[string]interface{}{
		"user_story_id": userStoryID.String(),
		"description":   "Test description",
	}

	tests := []struct {
		name          string
		serviceError  error
		expectedError string
	}{
		{
			name:          "user story not found",
			serviceError:  service.ErrUserStoryNotFound,
			expectedError: "Invalid params",
		},
		{
			name:          "user not found",
			serviceError:  service.ErrUserNotFound,
			expectedError: "Unauthorized access",
		},
		{
			name:          "generic service error",
			serviceError:  errors.New("database connection failed"),
			expectedError: "Internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockACService.ExpectedCalls = nil
			mockUSService.ExpectedCalls = nil

			// Setup mocks
			mockACService.On("CreateAcceptanceCriteria", mock.AnythingOfType("service.CreateAcceptanceCriteriaRequest")).Return(nil, tt.serviceError)

			// Create context with user
			ctx := createACContextWithUser(user)

			// Execute
			result, err := handler.Create(ctx, validArgs)

			// Verify
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedError)

			// Verify mocks
			mockACService.AssertExpectations(t)
			mockUSService.AssertExpectations(t)
		})
	}
}
