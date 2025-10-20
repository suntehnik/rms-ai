package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

func TestResourceHandler_ConvertRequirementsURI(t *testing.T) {
	tests := []struct {
		name          string
		uri           string
		expectedURI   string
		expectedError bool
		setupMock     func(*MockEpicService)
	}{
		{
			name:          "valid epic URI conversion",
			uri:           "requirements://epics/516722c8-4ca6-4cea-b78c-523fc3ea665f",
			expectedURI:   "epic://EP-001",
			expectedError: false,
			setupMock: func(mockService *MockEpicService) {
				epicUUID := uuid.MustParse("516722c8-4ca6-4cea-b78c-523fc3ea665f")
				epic := &models.Epic{
					ID:          epicUUID,
					ReferenceID: "EP-001",
					Title:       "Test Epic",
					Status:      models.EpicStatusBacklog,
					Priority:    2,
				}
				mockService.On("GetEpicByID", epicUUID).Return(epic, nil)
			},
		},
		{
			name:          "valid epic URI conversion with reference ID",
			uri:           "requirements://epics/EP-001",
			expectedURI:   "epic://EP-001",
			expectedError: false,
			setupMock: func(mockService *MockEpicService) {
				epic := &models.Epic{
					ID:          uuid.MustParse("516722c8-4ca6-4cea-b78c-523fc3ea665f"),
					ReferenceID: "EP-001",
					Title:       "Test Epic",
					Status:      models.EpicStatusBacklog,
					Priority:    2,
				}
				mockService.On("GetEpicByReferenceID", "EP-001").Return(epic, nil)
			},
		},
		{
			name:          "invalid identifier format",
			uri:           "requirements://epics/invalid-id",
			expectedURI:   "",
			expectedError: true,
			setupMock:     func(mockService *MockEpicService) {},
		},
		{
			name:          "collection URI supported",
			uri:           "requirements://epics",
			expectedURI:   "requirements://epics",
			expectedError: false,
			setupMock:     func(mockService *MockEpicService) {},
		},
		{
			name:          "unsupported entity type",
			uri:           "requirements://unknown/516722c8-4ca6-4cea-b78c-523fc3ea665f",
			expectedURI:   "",
			expectedError: true,
			setupMock:     func(mockService *MockEpicService) {},
		},
		{
			name:          "epic not found by UUID",
			uri:           "requirements://epics/516722c8-4ca6-4cea-b78c-523fc3ea665f",
			expectedURI:   "",
			expectedError: true,
			setupMock: func(mockService *MockEpicService) {
				epicUUID := uuid.MustParse("516722c8-4ca6-4cea-b78c-523fc3ea665f")
				mockService.On("GetEpicByID", epicUUID).Return(nil, service.ErrEpicNotFound)
			},
		},
		{
			name:          "epic not found by reference ID",
			uri:           "requirements://epics/EP-999",
			expectedURI:   "",
			expectedError: true,
			setupMock: func(mockService *MockEpicService) {
				mockService.On("GetEpicByReferenceID", "EP-999").Return(nil, service.ErrEpicNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockEpicService := new(MockEpicService)
			tt.setupMock(mockEpicService)

			handler := &ResourceHandler{
				epicService:               mockEpicService,
				userStoryService:          nil, // Not needed for these tests
				requirementService:        nil, // Not needed for these tests
				acceptanceCriteriaService: nil, // Not needed for these tests
				uriParser:                 NewURIParser(),
			}

			// Execute
			ctx := context.Background()
			result, err := handler.convertRequirementsURI(ctx, tt.uri)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURI, result)
			}

			// Verify mock expectations
			mockEpicService.AssertExpectations(t)
		})
	}
}

func TestResourceHandler_HandleResourcesRead_WithRequirementsURI(t *testing.T) {
	// Setup
	mockEpicService := new(MockEpicService)
	epicUUID := uuid.MustParse("516722c8-4ca6-4cea-b78c-523fc3ea665f")
	epic := &models.Epic{
		ID:          epicUUID,
		ReferenceID: "EP-001",
		Title:       "Test Epic",
		Status:      models.EpicStatusBacklog,
		Priority:    2,
	}

	// Mock the GetEpicByID call for URI conversion
	mockEpicService.On("GetEpicByID", epicUUID).Return(epic, nil)
	// Mock the GetEpicByReferenceID call for actual resource handling
	mockEpicService.On("GetEpicByReferenceID", "EP-001").Return(epic, nil)

	handler := &ResourceHandler{
		epicService:               mockEpicService,
		userStoryService:          nil,
		requirementService:        nil,
		acceptanceCriteriaService: nil,
		uriParser:                 NewURIParser(),
	}

	// Test parameters with requirements:// URI
	params := map[string]interface{}{
		"uri": "requirements://epics/516722c8-4ca6-4cea-b78c-523fc3ea665f",
	}

	// Execute
	ctx := context.Background()
	result, err := handler.HandleResourcesRead(ctx, params)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the result structure
	resourceResponse, ok := result.(*ResourceResponse)
	assert.True(t, ok, "Result should be a ResourceResponse")
	assert.Len(t, resourceResponse.Contents, 1, "Should have exactly one content item")
	assert.Equal(t, "epic://EP-001", resourceResponse.Contents[0].URI)
	assert.Equal(t, "application/json", resourceResponse.Contents[0].MimeType)
	assert.NotEmpty(t, resourceResponse.Contents[0].Text, "Content text should not be empty")

	// Verify mock expectations
	mockEpicService.AssertExpectations(t)
}

func TestResourceHandler_HandleResourcesRead_WithRequirementsURI_ReferenceID(t *testing.T) {
	// Setup
	mockEpicService := new(MockEpicService)
	epic := &models.Epic{
		ID:          uuid.MustParse("516722c8-4ca6-4cea-b78c-523fc3ea665f"),
		ReferenceID: "EP-001",
		Title:       "Test Epic",
		Status:      models.EpicStatusBacklog,
		Priority:    2,
	}

	// Mock the GetEpicByReferenceID call for URI conversion (first call)
	mockEpicService.On("GetEpicByReferenceID", "EP-001").Return(epic, nil).Once()
	// Mock the GetEpicByReferenceID call for actual resource handling (second call)
	mockEpicService.On("GetEpicByReferenceID", "EP-001").Return(epic, nil).Once()

	handler := &ResourceHandler{
		epicService:               mockEpicService,
		userStoryService:          nil,
		requirementService:        nil,
		acceptanceCriteriaService: nil,
		uriParser:                 NewURIParser(),
	}

	// Test parameters with requirements:// URI using reference ID
	params := map[string]interface{}{
		"uri": "requirements://epics/EP-001",
	}

	// Execute
	ctx := context.Background()
	result, err := handler.HandleResourcesRead(ctx, params)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the result structure
	resourceResponse, ok := result.(*ResourceResponse)
	assert.True(t, ok, "Result should be a ResourceResponse")
	assert.Len(t, resourceResponse.Contents, 1, "Should have exactly one content item")
	assert.Equal(t, "epic://EP-001", resourceResponse.Contents[0].URI)
	assert.Equal(t, "application/json", resourceResponse.Contents[0].MimeType)
	assert.NotEmpty(t, resourceResponse.Contents[0].Text, "Content text should not be empty")

	// Verify mock expectations
	mockEpicService.AssertExpectations(t)
}
