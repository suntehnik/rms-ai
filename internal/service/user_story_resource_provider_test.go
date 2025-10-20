package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"product-requirements-management/internal/models"
)

func TestUserStoryResourceProvider_GetResourceDescriptors_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockUserStoryRepository)
	logger := logrus.New()
	provider := NewUserStoryResourceProvider(mockRepo, logger)

	// Mock data
	id1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	id2 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")

	userStories := []models.UserStory{
		{
			ID:          id1,
			ReferenceID: "US-001",
			Title:       "User Authentication",
		},
		{
			ID:          id2,
			ReferenceID: "US-002",
			Title:       "User Profile Management",
		},
	}

	// Setup expectations
	mockRepo.On("List", mock.Anything, "created_at ASC", 1000, 0).Return(userStories, nil)

	// Execute
	ctx := context.Background()
	resources, err := provider.GetResourceDescriptors(ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, resources, 5) // 2 user stories * 2 variants (UUID + reference ID) + 1 collection resource

	// Check individual user story resources (UUID variants)
	assert.Equal(t, "requirements://user-stories/550e8400-e29b-41d4-a716-446655440001", resources[0].URI)
	assert.Equal(t, "User Story: User Authentication", resources[0].Name)
	assert.Equal(t, "User Story US-001: User Authentication", resources[0].Description)
	assert.Equal(t, "application/json", resources[0].MimeType)

	// Check individual user story resources (reference ID variants)
	assert.Equal(t, "requirements://user-stories/US-001", resources[1].URI)
	assert.Equal(t, "User Story: User Authentication", resources[1].Name)
	assert.Equal(t, "User Story US-001: User Authentication", resources[1].Description)
	assert.Equal(t, "application/json", resources[1].MimeType)

	assert.Equal(t, "requirements://user-stories/550e8400-e29b-41d4-a716-446655440002", resources[2].URI)
	assert.Equal(t, "User Story: User Profile Management", resources[2].Name)
	assert.Equal(t, "User Story US-002: User Profile Management", resources[2].Description)
	assert.Equal(t, "application/json", resources[2].MimeType)

	assert.Equal(t, "requirements://user-stories/US-002", resources[3].URI)
	assert.Equal(t, "User Story: User Profile Management", resources[3].Name)
	assert.Equal(t, "User Story US-002: User Profile Management", resources[3].Description)
	assert.Equal(t, "application/json", resources[3].MimeType)

	// Check collection resource
	assert.Equal(t, "requirements://user-stories", resources[4].URI)
	assert.Equal(t, "All User Stories", resources[4].Name)
	assert.Equal(t, "Complete list of all user stories in the system", resources[4].Description)
	assert.Equal(t, "application/json", resources[4].MimeType)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestUserStoryResourceProvider_GetResourceDescriptors_DatabaseError(t *testing.T) {
	// Setup
	mockRepo := new(MockUserStoryRepository)
	logger := logrus.New()
	provider := NewUserStoryResourceProvider(mockRepo, logger)

	// Setup expectations
	mockRepo.On("List", mock.Anything, "created_at ASC", 1000, 0).Return([]models.UserStory{}, errors.New("database error"))

	// Execute
	ctx := context.Background()
	resources, err := provider.GetResourceDescriptors(ctx)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resources)
	assert.Contains(t, err.Error(), "failed to get user stories")

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestUserStoryResourceProvider_GetResourceDescriptors_EmptyResult(t *testing.T) {
	// Setup
	mockRepo := new(MockUserStoryRepository)
	logger := logrus.New()
	provider := NewUserStoryResourceProvider(mockRepo, logger)

	// Setup expectations - no user stories
	mockRepo.On("List", mock.Anything, "created_at ASC", 1000, 0).Return([]models.UserStory{}, nil)

	// Execute
	ctx := context.Background()
	resources, err := provider.GetResourceDescriptors(ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, resources, 1) // Only collection resource

	// Check collection resource
	assert.Equal(t, "requirements://user-stories", resources[0].URI)
	assert.Equal(t, "All User Stories", resources[0].Name)
	assert.Equal(t, "Complete list of all user stories in the system", resources[0].Description)
	assert.Equal(t, "application/json", resources[0].MimeType)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestUserStoryResourceProvider_GetProviderName(t *testing.T) {
	// Setup
	mockRepo := new(MockUserStoryRepository)
	logger := logrus.New()
	provider := NewUserStoryResourceProvider(mockRepo, logger)

	// Execute
	name := provider.GetProviderName()

	// Assert
	assert.Equal(t, "user_story_provider", name)
}
