package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// Mock types for MCP tools tests
type MockUserStoryRepository struct{ mock.Mock }
type MockAcceptanceCriteriaRepository struct{ mock.Mock }
type MockRequirementRepository struct{ mock.Mock }
type MockRequirementRelationshipRepository struct{ mock.Mock }

func TestMCPToolsHandler_handleListSteeringDocuments_Success(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockAcceptanceCriteriaRepo := &MockAcceptanceCriteriaRepository{}
	mockRequirementRepo := &MockRequirementRepository{}
	mockRelationshipRepo := &MockRequirementRelationshipRepository{}

	handler := NewToolsHandler(
		mockUserRepo,
		mockEpicRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockRequirementRepo,
		mockRelationshipRepo,
		mockService,
	)

	// Create test user
	user := &models.User{
		ID:   uuid.New(),
		Role: models.RoleUser,
	}

	// Create test data
	docs := []models.SteeringDocument{
		{
			ID:          uuid.New(),
			ReferenceID: "STD-001",
			Title:       "Test Document 1",
			CreatorID:   user.ID,
		},
		{
			ID:          uuid.New(),
			ReferenceID: "STD-002",
			Title:       "Test Document 2",
			CreatorID:   user.ID,
		},
	}

	filters := service.SteeringDocumentFilters{
		Limit:  10,
		Offset: 0,
	}

	// Mock expectations
	mockService.On("ListSteeringDocuments", filters, user).Return(docs, int64(2), nil)

	// Create request arguments
	args := map[string]interface{}{
		"limit":  10,
		"offset": 0,
	}

	// Execute
	result, err := handler.handleListSteeringDocuments(context.Background(), args, user)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the result structure
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	data, hasData := resultMap["data"]
	assert.True(t, hasData)

	dataArray, ok := data.([]interface{})
	require.True(t, ok)
	assert.Len(t, dataArray, 2)

	// Verify first document
	firstDoc := dataArray[0].(map[string]interface{})
	assert.Equal(t, "STD-001", firstDoc["reference_id"])
	assert.Equal(t, "Test Document 1", firstDoc["title"])

	mockService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleCreateSteeringDocument_Success(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockAcceptanceCriteriaRepo := &MockAcceptanceCriteriaRepository{}
	mockRequirementRepo := &MockRequirementRepository{}
	mockRelationshipRepo := &MockRequirementRelationshipRepository{}

	handler := NewToolsHandler(
		mockUserRepo,
		mockEpicRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockRequirementRepo,
		mockRelationshipRepo,
		mockService,
	)

	// Create test user
	user := &models.User{
		ID:   uuid.New(),
		Role: models.RoleUser,
	}

	// Create expected document
	expectedDoc := &models.SteeringDocument{
		ID:          uuid.New(),
		ReferenceID: "STD-001",
		Title:       "Test Document",
		Description: stringPtr("Test description"),
		CreatorID:   user.ID,
	}

	req := service.CreateSteeringDocumentRequest{
		Title:       "Test Document",
		Description: stringPtr("Test description"),
	}

	// Mock expectations
	mockService.On("CreateSteeringDocument", req, user).Return(expectedDoc, nil)

	// Create request arguments
	args := map[string]interface{}{
		"title":       "Test Document",
		"description": "Test description",
	}

	// Execute
	result, err := handler.handleCreateSteeringDocument(context.Background(), args, user)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the result structure
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, expectedDoc.ID.String(), resultMap["id"])
	assert.Equal(t, expectedDoc.ReferenceID, resultMap["reference_id"])
	assert.Equal(t, expectedDoc.Title, resultMap["title"])
	assert.Equal(t, *expectedDoc.Description, resultMap["description"])

	mockService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleGetSteeringDocument_ByUUID_Success(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockAcceptanceCriteriaRepo := &MockAcceptanceCriteriaRepository{}
	mockRequirementRepo := &MockRequirementRepository{}
	mockRelationshipRepo := &MockRequirementRelationshipRepository{}

	handler := NewToolsHandler(
		mockUserRepo,
		mockEpicRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockRequirementRepo,
		mockRelationshipRepo,
		mockService,
	)

	// Create test user
	user := &models.User{
		ID:   uuid.New(),
		Role: models.RoleUser,
	}

	// Create test document
	docID := uuid.New()
	expectedDoc := &models.SteeringDocument{
		ID:          docID,
		ReferenceID: "STD-001",
		Title:       "Test Document",
		CreatorID:   user.ID,
	}

	// Mock expectations
	mockService.On("GetSteeringDocumentByID", docID, user).Return(expectedDoc, nil)

	// Create request arguments
	args := map[string]interface{}{
		"id": docID.String(),
	}

	// Execute
	result, err := handler.handleGetSteeringDocument(context.Background(), args, user)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the result structure
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, expectedDoc.ID.String(), resultMap["id"])
	assert.Equal(t, expectedDoc.ReferenceID, resultMap["reference_id"])
	assert.Equal(t, expectedDoc.Title, resultMap["title"])

	mockService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleGetSteeringDocument_ByReferenceID_Success(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockAcceptanceCriteriaRepo := &MockAcceptanceCriteriaRepository{}
	mockRequirementRepo := &MockRequirementRepository{}
	mockRelationshipRepo := &MockRequirementRelationshipRepository{}

	handler := NewToolsHandler(
		mockUserRepo,
		mockEpicRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockRequirementRepo,
		mockRelationshipRepo,
		mockService,
	)

	// Create test user
	user := &models.User{
		ID:   uuid.New(),
		Role: models.RoleUser,
	}

	// Create test document
	referenceID := "STD-001"
	expectedDoc := &models.SteeringDocument{
		ID:          uuid.New(),
		ReferenceID: referenceID,
		Title:       "Test Document",
		CreatorID:   user.ID,
	}

	// Mock expectations
	mockService.On("GetSteeringDocumentByReferenceID", referenceID, user).Return(expectedDoc, nil)

	// Create request arguments
	args := map[string]interface{}{
		"id": referenceID,
	}

	// Execute
	result, err := handler.handleGetSteeringDocument(context.Background(), args, user)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the result structure
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, expectedDoc.ID.String(), resultMap["id"])
	assert.Equal(t, expectedDoc.ReferenceID, resultMap["reference_id"])
	assert.Equal(t, expectedDoc.Title, resultMap["title"])

	mockService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleUpdateSteeringDocument_Success(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockAcceptanceCriteriaRepo := &MockAcceptanceCriteriaRepository{}
	mockRequirementRepo := &MockRequirementRepository{}
	mockRelationshipRepo := &MockRequirementRelationshipRepository{}

	handler := NewToolsHandler(
		mockUserRepo,
		mockEpicRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockRequirementRepo,
		mockRelationshipRepo,
		mockService,
	)

	// Create test user
	user := &models.User{
		ID:   uuid.New(),
		Role: models.RoleUser,
	}

	// Create test data
	docID := uuid.New()
	req := service.UpdateSteeringDocumentRequest{
		Title:       stringPtr("Updated Title"),
		Description: stringPtr("Updated description"),
	}
	expectedDoc := &models.SteeringDocument{
		ID:          docID,
		ReferenceID: "STD-001",
		Title:       *req.Title,
		Description: req.Description,
		CreatorID:   user.ID,
	}

	// Mock expectations
	mockService.On("UpdateSteeringDocument", docID, req, user).Return(expectedDoc, nil)

	// Create request arguments
	args := map[string]interface{}{
		"id":          docID.String(),
		"title":       "Updated Title",
		"description": "Updated description",
	}

	// Execute
	result, err := handler.handleUpdateSteeringDocument(context.Background(), args, user)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the result structure
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, expectedDoc.ID.String(), resultMap["id"])
	assert.Equal(t, expectedDoc.ReferenceID, resultMap["reference_id"])
	assert.Equal(t, expectedDoc.Title, resultMap["title"])
	assert.Equal(t, *expectedDoc.Description, resultMap["description"])

	mockService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleLinkSteeringToEpic_Success(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockAcceptanceCriteriaRepo := &MockAcceptanceCriteriaRepository{}
	mockRequirementRepo := &MockRequirementRepository{}
	mockRelationshipRepo := &MockRequirementRelationshipRepository{}

	handler := NewToolsHandler(
		mockUserRepo,
		mockEpicRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockRequirementRepo,
		mockRelationshipRepo,
		mockService,
	)

	// Create test user
	user := &models.User{
		ID:   uuid.New(),
		Role: models.RoleUser,
	}

	// Create test data
	docID := uuid.New()
	epicID := uuid.New()

	// Mock expectations
	mockService.On("LinkSteeringDocumentToEpic", docID, epicID, user).Return(nil)

	// Create request arguments
	args := map[string]interface{}{
		"steering_document_id": docID.String(),
		"epic_id":              epicID.String(),
	}

	// Execute
	result, err := handler.handleLinkSteeringToEpic(context.Background(), args, user)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the result structure
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "Steering document linked to epic successfully", resultMap["message"])
	assert.Equal(t, docID.String(), resultMap["steering_document_id"])
	assert.Equal(t, epicID.String(), resultMap["epic_id"])

	mockService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleUnlinkSteeringFromEpic_Success(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockAcceptanceCriteriaRepo := &MockAcceptanceCriteriaRepository{}
	mockRequirementRepo := &MockRequirementRepository{}
	mockRelationshipRepo := &MockRequirementRelationshipRepository{}

	handler := NewToolsHandler(
		mockUserRepo,
		mockEpicRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockRequirementRepo,
		mockRelationshipRepo,
		mockService,
	)

	// Create test user
	user := &models.User{
		ID:   uuid.New(),
		Role: models.RoleUser,
	}

	// Create test data
	docID := uuid.New()
	epicID := uuid.New()

	// Mock expectations
	mockService.On("UnlinkSteeringDocumentFromEpic", docID, epicID, user).Return(nil)

	// Create request arguments
	args := map[string]interface{}{
		"steering_document_id": docID.String(),
		"epic_id":              epicID.String(),
	}

	// Execute
	result, err := handler.handleUnlinkSteeringFromEpic(context.Background(), args, user)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the result structure
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "Steering document unlinked from epic successfully", resultMap["message"])
	assert.Equal(t, docID.String(), resultMap["steering_document_id"])
	assert.Equal(t, epicID.String(), resultMap["epic_id"])

	mockService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleGetEpicSteeringDocuments_Success(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockAcceptanceCriteriaRepo := &MockAcceptanceCriteriaRepository{}
	mockRequirementRepo := &MockRequirementRepository{}
	mockRelationshipRepo := &MockRequirementRelationshipRepository{}

	handler := NewToolsHandler(
		mockUserRepo,
		mockEpicRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockRequirementRepo,
		mockRelationshipRepo,
		mockService,
	)

	// Create test user
	user := &models.User{
		ID:   uuid.New(),
		Role: models.RoleUser,
	}

	// Create test data
	epicID := uuid.New()
	docs := []models.SteeringDocument{
		{
			ID:          uuid.New(),
			ReferenceID: "STD-001",
			Title:       "Test Document 1",
			CreatorID:   user.ID,
		},
		{
			ID:          uuid.New(),
			ReferenceID: "STD-002",
			Title:       "Test Document 2",
			CreatorID:   user.ID,
		},
	}

	// Mock expectations
	mockService.On("GetSteeringDocumentsByEpicID", epicID, user).Return(docs, nil)

	// Create request arguments
	args := map[string]interface{}{
		"epic_id": epicID.String(),
	}

	// Execute
	result, err := handler.handleGetEpicSteeringDocuments(context.Background(), args, user)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the result structure
	resultArray, ok := result.([]interface{})
	require.True(t, ok)
	assert.Len(t, resultArray, 2)

	// Verify first document
	firstDoc := resultArray[0].(map[string]interface{})
	assert.Equal(t, "STD-001", firstDoc["reference_id"])
	assert.Equal(t, "Test Document 1", firstDoc["title"])

	mockService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleCreateSteeringDocument_ValidationError(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockAcceptanceCriteriaRepo := &MockAcceptanceCriteriaRepository{}
	mockRequirementRepo := &MockRequirementRepository{}
	mockRelationshipRepo := &MockRequirementRelationshipRepository{}

	handler := NewToolsHandler(
		mockUserRepo,
		mockEpicRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockRequirementRepo,
		mockRelationshipRepo,
		mockService,
	)

	// Create test user
	user := &models.User{
		ID:   uuid.New(),
		Role: models.RoleUser,
	}

	req := service.CreateSteeringDocumentRequest{
		Title: "", // Invalid empty title
	}

	// Mock expectations
	mockService.On("CreateSteeringDocument", req, user).Return((*models.SteeringDocument)(nil), service.ErrValidation)

	// Create request arguments
	args := map[string]interface{}{
		"title": "", // Invalid empty title
	}

	// Execute
	result, err := handler.handleCreateSteeringDocument(context.Background(), args, user)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, service.ErrValidation, err)

	mockService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleGetSteeringDocument_NotFound(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockAcceptanceCriteriaRepo := &MockAcceptanceCriteriaRepository{}
	mockRequirementRepo := &MockRequirementRepository{}
	mockRelationshipRepo := &MockRequirementRelationshipRepository{}

	handler := NewToolsHandler(
		mockUserRepo,
		mockEpicRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockRequirementRepo,
		mockRelationshipRepo,
		mockService,
	)

	// Create test user
	user := &models.User{
		ID:   uuid.New(),
		Role: models.RoleUser,
	}

	docID := uuid.New()

	// Mock expectations
	mockService.On("GetSteeringDocumentByID", docID, user).Return((*models.SteeringDocument)(nil), service.ErrSteeringDocumentNotFound)

	// Create request arguments
	args := map[string]interface{}{
		"id": docID.String(),
	}

	// Execute
	result, err := handler.handleGetSteeringDocument(context.Background(), args, user)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, service.ErrSteeringDocumentNotFound, err)

	mockService.AssertExpectations(t)
}

// Helper function to convert interface{} to JSON and back for testing
func toJSONAndBack(t *testing.T, input interface{}) map[string]interface{} {
	jsonData, err := json.Marshal(input)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	require.NoError(t, err)

	return result
}
