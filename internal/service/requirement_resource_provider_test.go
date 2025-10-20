package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"product-requirements-management/internal/models"
)

func TestRequirementResourceProvider_GetResourceDescriptors(t *testing.T) {
	tests := []struct {
		name          string
		requirements  []models.Requirement
		expectedCount int
		expectedError bool
		setupMock     func(*MockRequirementRepository)
	}{
		{
			name: "successful resource generation with requirements",
			requirements: []models.Requirement{
				{
					ID:          uuid.New(),
					ReferenceID: "REQ-001",
					Title:       "User Authentication API",
				},
				{
					ID:          uuid.New(),
					ReferenceID: "REQ-002",
					Title:       "Password Validation Rules",
				},
			},
			expectedCount: 3, // 2 individual requirements + 1 collection resource
			expectedError: false,
			setupMock: func(mockRepo *MockRequirementRepository) {
				mockRepo.On("List", mock.Anything, "created_at ASC", 1000, 0).Return([]models.Requirement{
					{
						ID:          uuid.New(),
						ReferenceID: "REQ-001",
						Title:       "User Authentication API",
					},
					{
						ID:          uuid.New(),
						ReferenceID: "REQ-002",
						Title:       "Password Validation Rules",
					},
				}, nil)
			},
		},
		{
			name:          "successful resource generation with no requirements",
			requirements:  []models.Requirement{},
			expectedCount: 1, // Only collection resource
			expectedError: false,
			setupMock: func(mockRepo *MockRequirementRepository) {
				mockRepo.On("List", mock.Anything, "created_at ASC", 1000, 0).Return([]models.Requirement{}, nil)
			},
		},
		{
			name:          "repository error",
			requirements:  nil,
			expectedCount: 0,
			expectedError: true,
			setupMock: func(mockRepo *MockRequirementRepository) {
				mockRepo.On("List", mock.Anything, "created_at ASC", 1000, 0).Return([]models.Requirement{}, assert.AnError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := &MockRequirementRepository{}
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel) // Reduce log noise in tests

			tt.setupMock(mockRepo)

			provider := NewRequirementResourceProvider(mockRepo, logger)

			// Execute
			ctx := context.Background()
			resources, err := provider.GetResourceDescriptors(ctx)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, resources)
			} else {
				assert.NoError(t, err)
				assert.Len(t, resources, tt.expectedCount)

				// Verify collection resource is always present
				collectionFound := false
				for _, resource := range resources {
					if resource.URI == "requirements://requirements" {
						collectionFound = true
						assert.Equal(t, "All Requirements", resource.Name)
						assert.Equal(t, "Complete list of all requirements in the system", resource.Description)
						assert.Equal(t, "application/json", resource.MimeType)
						break
					}
				}
				assert.True(t, collectionFound, "Collection resource should always be present")

				// Verify individual requirement resources if requirements exist
				if len(tt.requirements) > 0 {
					requirementResourceCount := 0
					for _, resource := range resources {
						if resource.URI != "requirements://requirements" {
							requirementResourceCount++
							assert.Contains(t, resource.URI, "requirements://requirements/")
							assert.Contains(t, resource.Name, "Requirement:")
							assert.Equal(t, "application/json", resource.MimeType)
						}
					}
					assert.Equal(t, len(tt.requirements), requirementResourceCount)
				}
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRequirementResourceProvider_GetProviderName(t *testing.T) {
	// Setup
	mockRepo := &MockRequirementRepository{}
	logger := logrus.New()
	provider := NewRequirementResourceProvider(mockRepo, logger)

	// Execute
	name := provider.GetProviderName()

	// Assert
	assert.Equal(t, "requirement_provider", name)
}

func TestNewRequirementResourceProvider(t *testing.T) {
	// Setup
	mockRepo := &MockRequirementRepository{}
	logger := logrus.New()

	// Execute
	provider := NewRequirementResourceProvider(mockRepo, logger)

	// Assert
	assert.NotNil(t, provider)
	assert.Equal(t, "requirement_provider", provider.GetProviderName())
}

func TestRequirementResourceProvider_GetResourceDescriptors_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockRequirementRepository)
	logger := logrus.New()
	provider := NewRequirementResourceProvider(mockRepo, logger)

	// Mock data
	id1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	id2 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")

	requirements := []models.Requirement{
		{
			ID:          id1,
			ReferenceID: "REQ-001",
			Title:       "User Authentication API",
		},
		{
			ID:          id2,
			ReferenceID: "REQ-002",
			Title:       "Password Validation Rules",
		},
	}

	// Setup expectations
	mockRepo.On("List", mock.Anything, "created_at ASC", 1000, 0).Return(requirements, nil)

	// Execute
	ctx := context.Background()
	resources, err := provider.GetResourceDescriptors(ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, resources, 3) // 2 individual requirements + 1 collection resource

	// Check individual requirement resources
	assert.Equal(t, "requirements://requirements/550e8400-e29b-41d4-a716-446655440001", resources[0].URI)
	assert.Equal(t, "Requirement: User Authentication API", resources[0].Name)
	assert.Equal(t, "Requirement REQ-001: User Authentication API", resources[0].Description)
	assert.Equal(t, "application/json", resources[0].MimeType)

	assert.Equal(t, "requirements://requirements/550e8400-e29b-41d4-a716-446655440002", resources[1].URI)
	assert.Equal(t, "Requirement: Password Validation Rules", resources[1].Name)
	assert.Equal(t, "Requirement REQ-002: Password Validation Rules", resources[1].Description)
	assert.Equal(t, "application/json", resources[1].MimeType)

	// Check collection resource
	assert.Equal(t, "requirements://requirements", resources[2].URI)
	assert.Equal(t, "All Requirements", resources[2].Name)
	assert.Equal(t, "Complete list of all requirements in the system", resources[2].Description)
	assert.Equal(t, "application/json", resources[2].MimeType)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestRequirementResourceProvider_GetResourceDescriptors_DatabaseError(t *testing.T) {
	// Setup
	mockRepo := new(MockRequirementRepository)
	logger := logrus.New()
	provider := NewRequirementResourceProvider(mockRepo, logger)

	// Setup expectations
	mockRepo.On("List", mock.Anything, "created_at ASC", 1000, 0).Return([]models.Requirement{}, assert.AnError)

	// Execute
	ctx := context.Background()
	resources, err := provider.GetResourceDescriptors(ctx)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resources)
	assert.Contains(t, err.Error(), "failed to get requirements")

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestRequirementResourceProvider_GetResourceDescriptors_EmptyResult(t *testing.T) {
	// Setup
	mockRepo := new(MockRequirementRepository)
	logger := logrus.New()
	provider := NewRequirementResourceProvider(mockRepo, logger)

	// Setup expectations - no requirements
	mockRepo.On("List", mock.Anything, "created_at ASC", 1000, 0).Return([]models.Requirement{}, nil)

	// Execute
	ctx := context.Background()
	resources, err := provider.GetResourceDescriptors(ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, resources, 1) // Only collection resource

	// Check collection resource
	assert.Equal(t, "requirements://requirements", resources[0].URI)
	assert.Equal(t, "All Requirements", resources[0].Name)
	assert.Equal(t, "Complete list of all requirements in the system", resources[0].Description)
	assert.Equal(t, "application/json", resources[0].MimeType)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}
