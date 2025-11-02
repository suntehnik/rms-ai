package tools

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"product-requirements-management/internal/jsonrpc"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// MockAcceptanceCriteriaService is a mock implementation of AcceptanceCriteriaService
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
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.AcceptanceCriteria), args.Get(1).(int64), args.Error(2)
}

func (m *MockAcceptanceCriteriaService) GetAcceptanceCriteriaByUserStory(userStoryID uuid.UUID, limit, offset int) ([]models.AcceptanceCriteria, int64, error) {
	args := m.Called(userStoryID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.AcceptanceCriteria), args.Get(1).(int64), args.Error(2)
}

func (m *MockAcceptanceCriteriaService) GetAcceptanceCriteriaByAuthor(authorID uuid.UUID, limit, offset int) ([]models.AcceptanceCriteria, int64, error) {
	args := m.Called(authorID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.AcceptanceCriteria), args.Get(1).(int64), args.Error(2)
}

func (m *MockAcceptanceCriteriaService) ValidateUserStoryHasAcceptanceCriteria(userStoryID uuid.UUID) error {
	args := m.Called(userStoryID)
	return args.Error(0)
}

func TestAcceptanceCriteriaHandler_GetSupportedTools(t *testing.T) {
	mockACService := &MockAcceptanceCriteriaService{}
	mockUSService := &MockUserStoryService{}
	handler := NewAcceptanceCriteriaHandler(mockACService, mockUSService)

	tools := handler.GetSupportedTools()

	assert.Equal(t, []string{"create_acceptance_criteria"}, tools)
}

func TestAcceptanceCriteriaHandler_HandleTool(t *testing.T) {
	mockACService := &MockAcceptanceCriteriaService{}
	mockUSService := &MockUserStoryService{}
	handler := NewAcceptanceCriteriaHandler(mockACService, mockUSService)

	tests := []struct {
		name     string
		toolName string
		args     map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "valid create_acceptance_criteria tool",
			toolName: "create_acceptance_criteria",
			args:     map[string]interface{}{"user_story_id": "US-001", "description": "Test description"},
			wantErr:  false,
		},
		{
			name:     "invalid tool name",
			toolName: "invalid_tool",
			args:     map[string]interface{}{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock context with gin context and user
			ctx := context.Background()
			ginCtx := &gin.Context{}
			user := &models.User{ID: uuid.New()}
			ginCtx.Set("user", user)
			ctx = context.WithValue(ctx, "gin_context", ginCtx)

			if tt.toolName == "create_acceptance_criteria" {
				// Mock the user story service call
				userStoryID := uuid.New()
				mockUSService.On("GetUserStoryByReferenceID", "US-001").Return(&models.UserStory{ID: userStoryID}, nil)

				// Mock the acceptance criteria service call
				ac := &models.AcceptanceCriteria{
					ID:          uuid.New(),
					ReferenceID: "AC-001",
					UserStoryID: userStoryID,
					AuthorID:    user.ID,
					Description: "Test description",
				}
				mockACService.On("CreateAcceptanceCriteria", mock.AnythingOfType("service.CreateAcceptanceCriteriaRequest")).Return(ac, nil)
			}

			result, err := handler.HandleTool(ctx, tt.toolName, tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestAcceptanceCriteriaHandler_Create_ValidationErrors(t *testing.T) {
	mockACService := &MockAcceptanceCriteriaService{}
	mockUSService := &MockUserStoryService{}
	handler := NewAcceptanceCriteriaHandler(mockACService, mockUSService)

	tests := []struct {
		name    string
		args    map[string]interface{}
		wantErr string
	}{
		{
			name:    "missing user_story_id",
			args:    map[string]interface{}{"description": "Test description"},
			wantErr: "Missing required argument: user_story_id",
		},
		{
			name:    "empty user_story_id",
			args:    map[string]interface{}{"user_story_id": "", "description": "Test description"},
			wantErr: "Missing or invalid 'user_story_id' argument",
		},
		{
			name:    "missing description",
			args:    map[string]interface{}{"user_story_id": "US-001"},
			wantErr: "Missing required argument: description",
		},
		{
			name:    "empty description",
			args:    map[string]interface{}{"user_story_id": "US-001", "description": ""},
			wantErr: "Missing or invalid 'description' argument",
		},
		{
			name:    "description too long",
			args:    map[string]interface{}{"user_story_id": "US-001", "description": string(make([]byte, 50001))},
			wantErr: "Description exceeds maximum length of 50000 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock context with gin context and user
			ctx := context.Background()
			ginCtx := &gin.Context{}
			user := &models.User{ID: uuid.New()}
			ginCtx.Set("user", user)
			ctx = context.WithValue(ctx, "gin_context", ginCtx)

			result, err := handler.Create(ctx, tt.args)

			assert.Error(t, err)
			assert.Nil(t, result)

			// Check if it's a JSON-RPC error with the expected message
			if jsonrpcErr, ok := err.(*jsonrpc.JSONRPCError); ok {
				// The specific error message might be in the Data field or Message field
				errorText := jsonrpcErr.Message
				if jsonrpcErr.Data != nil {
					if dataStr, ok := jsonrpcErr.Data.(string); ok {
						errorText = dataStr
					}
				}
				assert.Contains(t, errorText, tt.wantErr)
			} else {
				t.Errorf("Expected JSON-RPC error, got %T", err)
			}
		})
	}
}

func TestNewAcceptanceCriteriaHandler(t *testing.T) {
	mockACService := &MockAcceptanceCriteriaService{}
	mockUSService := &MockUserStoryService{}

	handler := NewAcceptanceCriteriaHandler(mockACService, mockUSService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockACService, handler.acceptanceCriteriaService)
	assert.Equal(t, mockUSService, handler.userStoryService)
}
