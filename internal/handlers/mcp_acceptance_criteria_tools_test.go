package handlers

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"product-requirements-management/internal/auth"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

func TestMCPToolsHandler_handleCreateAcceptanceCriteria_Success(t *testing.T) {
	// Create mock services
	mockEpicService := &MockEpicService{}
	mockUserStoryService := &MockUserStoryService{}
	mockRequirementService := &MockRequirementService{}
	mockAcceptanceCriteriaService := &MockAcceptanceCriteriaService{}
	mockSearchService := &MockSearchService{}
	mockSteeringDocumentService := &MockSteeringDocumentService{}

	handler := NewToolsHandler(
		mockEpicService,
		mockUserStoryService,
		mockRequirementService,
		mockAcceptanceCriteriaService,
		mockSearchService,
		mockSteeringDocumentService,
		nil, // promptService
	)

	// Create test user
	user := &models.User{
		ID:   uuid.New(),
		Role: models.RoleUser,
	}

	// Create test user story
	userStoryID := uuid.New()
	userStory := &models.UserStory{
		ID:          userStoryID,
		ReferenceID: "US-001",
		Title:       "Test User Story",
	}

	// Create expected acceptance criteria
	expectedAC := &models.AcceptanceCriteria{
		ID:          uuid.New(),
		ReferenceID: "AC-001",
		UserStoryID: userStoryID,
		AuthorID:    user.ID,
		Description: "WHEN a user enters valid credentials THEN the system SHALL authenticate the user",
	}

	// Set up mock expectations
	mockUserStoryService.On("GetUserStoryByReferenceID", "US-001").Return(userStory, nil)
	mockAcceptanceCriteriaService.On("CreateAcceptanceCriteria", mock.MatchedBy(func(req service.CreateAcceptanceCriteriaRequest) bool {
		return req.UserStoryID == userStoryID &&
			req.AuthorID == user.ID &&
			req.Description == "WHEN a user enters valid credentials THEN the system SHALL authenticate the user"
	})).Return(expectedAC, nil)

	// Create context with user
	ctx := createContextWithUser(user)

	// Prepare arguments
	args := map[string]interface{}{
		"user_story_id": "US-001",
		"description":   "WHEN a user enters valid credentials THEN the system SHALL authenticate the user",
	}

	// Call the handler
	result, err := handler.handleCreateAcceptanceCriteria(ctx, args)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	toolResponse, ok := result.(*ToolResponse)
	assert.True(t, ok)
	assert.Len(t, toolResponse.Content, 2)
	assert.Equal(t, "text", toolResponse.Content[0].Type)
	assert.Contains(t, toolResponse.Content[0].Text, "Successfully created acceptance criteria AC-001")

	// Verify mock expectations
	mockUserStoryService.AssertExpectations(t)
	mockAcceptanceCriteriaService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleCreateAcceptanceCriteria_InvalidUserStoryID(t *testing.T) {
	// Create mock services
	mockEpicService := &MockEpicService{}
	mockUserStoryService := &MockUserStoryService{}
	mockRequirementService := &MockRequirementService{}
	mockAcceptanceCriteriaService := &MockAcceptanceCriteriaService{}
	mockSearchService := &MockSearchService{}
	mockSteeringDocumentService := &MockSteeringDocumentService{}

	handler := NewToolsHandler(
		mockEpicService,
		mockUserStoryService,
		mockRequirementService,
		mockAcceptanceCriteriaService,
		mockSearchService,
		mockSteeringDocumentService,
		nil, // promptService
	)

	// Create test user
	user := &models.User{
		ID:   uuid.New(),
		Role: models.RoleUser,
	}

	// Set up mock expectations for user story not found
	mockUserStoryService.On("GetUserStoryByReferenceID", "US-999").Return(nil, service.ErrUserStoryNotFound)

	// Create context with user
	ctx := createContextWithUser(user)

	// Prepare arguments with invalid user story ID
	args := map[string]interface{}{
		"user_story_id": "US-999",
		"description":   "WHEN a user enters valid credentials THEN the system SHALL authenticate the user",
	}

	// Call the handler
	result, err := handler.handleCreateAcceptanceCriteria(ctx, args)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Invalid params")

	// Verify mock expectations
	mockUserStoryService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleCreateAcceptanceCriteria_MissingDescription(t *testing.T) {
	// Create mock services
	mockEpicService := &MockEpicService{}
	mockUserStoryService := &MockUserStoryService{}
	mockRequirementService := &MockRequirementService{}
	mockAcceptanceCriteriaService := &MockAcceptanceCriteriaService{}
	mockSearchService := &MockSearchService{}
	mockSteeringDocumentService := &MockSteeringDocumentService{}

	handler := NewToolsHandler(
		mockEpicService,
		mockUserStoryService,
		mockRequirementService,
		mockAcceptanceCriteriaService,
		mockSearchService,
		mockSteeringDocumentService,
		nil, // promptService
	)

	// Create test user
	user := &models.User{
		ID:   uuid.New(),
		Role: models.RoleUser,
	}

	// Create context with user
	ctx := createContextWithUser(user)

	// Prepare arguments with missing description
	args := map[string]interface{}{
		"user_story_id": "US-001",
	}

	// Call the handler
	result, err := handler.handleCreateAcceptanceCriteria(ctx, args)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Invalid params")
}

// createContextWithUser creates a context with a Gin context that has the user set
func createContextWithUser(user *models.User) context.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(auth.UserContextKey, user)
	return context.WithValue(context.Background(), "gin_context", c)
}
