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

func TestEpicResourceProvider_GetResourceDescriptors(t *testing.T) {
	tests := []struct {
		name          string
		epics         []models.Epic
		expectedCount int
		expectedError bool
		setupMock     func(*MockEpicRepository)
	}{
		{
			name: "successful resource generation with epics",
			epics: []models.Epic{
				{
					ID:          uuid.New(),
					ReferenceID: "EP-001",
					Title:       "User Authentication System",
				},
				{
					ID:          uuid.New(),
					ReferenceID: "EP-002",
					Title:       "Payment Processing",
				},
			},
			expectedCount: 3, // 2 individual epics + 1 collection resource
			expectedError: false,
			setupMock: func(mockRepo *MockEpicRepository) {
				mockRepo.On("List", mock.Anything, "created_at ASC", 1000, 0).Return([]models.Epic{
					{
						ID:          uuid.New(),
						ReferenceID: "EP-001",
						Title:       "User Authentication System",
					},
					{
						ID:          uuid.New(),
						ReferenceID: "EP-002",
						Title:       "Payment Processing",
					},
				}, nil)
			},
		},
		{
			name:          "successful resource generation with no epics",
			epics:         []models.Epic{},
			expectedCount: 1, // Only collection resource
			expectedError: false,
			setupMock: func(mockRepo *MockEpicRepository) {
				mockRepo.On("List", mock.Anything, "created_at ASC", 1000, 0).Return([]models.Epic{}, nil)
			},
		},
		{
			name:          "repository error",
			epics:         nil,
			expectedCount: 0,
			expectedError: true,
			setupMock: func(mockRepo *MockEpicRepository) {
				mockRepo.On("List", mock.Anything, "created_at ASC", 1000, 0).Return([]models.Epic{}, assert.AnError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := &MockEpicRepository{}
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel) // Reduce log noise in tests

			tt.setupMock(mockRepo)

			provider := NewEpicResourceProvider(mockRepo, logger)

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
					if resource.URI == "requirements://epics" {
						collectionFound = true
						assert.Equal(t, "All Epics", resource.Name)
						assert.Equal(t, "Complete list of all epics in the system", resource.Description)
						assert.Equal(t, "application/json", resource.MimeType)
						break
					}
				}
				assert.True(t, collectionFound, "Collection resource should always be present")

				// Verify individual epic resources if epics exist
				if len(tt.epics) > 0 {
					epicResourceCount := 0
					for _, resource := range resources {
						if resource.URI != "requirements://epics" {
							epicResourceCount++
							assert.Contains(t, resource.URI, "requirements://epics/")
							assert.Contains(t, resource.Name, "Epic:")
							assert.Equal(t, "application/json", resource.MimeType)
						}
					}
					assert.Equal(t, len(tt.epics), epicResourceCount)
				}
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEpicResourceProvider_GetProviderName(t *testing.T) {
	// Setup
	mockRepo := &MockEpicRepository{}
	logger := logrus.New()
	provider := NewEpicResourceProvider(mockRepo, logger)

	// Execute
	name := provider.GetProviderName()

	// Assert
	assert.Equal(t, "epic_provider", name)
}

func TestNewEpicResourceProvider(t *testing.T) {
	// Setup
	mockRepo := &MockEpicRepository{}
	logger := logrus.New()

	// Execute
	provider := NewEpicResourceProvider(mockRepo, logger)

	// Assert
	assert.NotNil(t, provider)
	assert.Equal(t, "epic_provider", provider.GetProviderName())
}
