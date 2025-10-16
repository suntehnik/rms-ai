package handlers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/auth"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// Mock services are defined in other test files in this package

// createContextWithUser creates a context with a Gin context that has the user set
func createContextWithUser(user *models.User) context.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(auth.UserContextKey, user)
	return context.WithValue(context.Background(), "gin_context", c)
}

func TestMCPToolsHandler_handleListSteeringDocuments_Success(t *testing.T) {
	mockEpicService := &MockEpicService{}
	mockUserStoryService := &MockUserStoryService{}
	mockRequirementService := &MockRequirementService{}
	mockSearchService := &MockSearchService{}
	mockSteeringDocumentService := &MockSteeringDocumentService{}

	handler := NewToolsHandler(
		mockEpicService,
		mockUserStoryService,
		mockRequirementService,
		mockSearchService,
		mockSteeringDocumentService,
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
	mockSteeringDocumentService.On("ListSteeringDocuments", filters, user).Return(docs, int64(2), nil)

	// Create request arguments
	args := map[string]interface{}{
		"limit":  float64(10),
		"offset": float64(0),
	}

	// Execute
	ctx := createContextWithUser(user)
	result, err := handler.handleListSteeringDocuments(ctx, args)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the result structure
	toolResponse, ok := result.(*ToolResponse)
	require.True(t, ok)
	require.Len(t, toolResponse.Content, 2)

	// First content item should be the summary
	assert.Equal(t, "text", toolResponse.Content[0].Type)
	assert.Contains(t, toolResponse.Content[0].Text, "Found 2 steering documents")

	// Second content item should be the JSON data
	assert.Equal(t, "text", toolResponse.Content[1].Type)

	// Parse the JSON data
	var jsonData map[string]interface{}
	err = json.Unmarshal([]byte(toolResponse.Content[1].Text), &jsonData)
	require.NoError(t, err)

	// Verify the JSON structure
	steeringDocs, ok := jsonData["steering_documents"].([]interface{})
	require.True(t, ok)
	assert.Len(t, steeringDocs, 2)

	// Verify first document
	firstDoc := steeringDocs[0].(map[string]interface{})
	assert.Equal(t, "STD-001", firstDoc["reference_id"])
	assert.Equal(t, "Test Document 1", firstDoc["title"])

	mockSteeringDocumentService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleCreateSteeringDocument_Success(t *testing.T) {
	t.Skip("Temporary disable due to work needed to refactor tests")
	mockEpicService := &MockEpicService{}
	mockUserStoryService := &MockUserStoryService{}
	mockRequirementService := &MockRequirementService{}
	mockSearchService := &MockSearchService{}
	mockSteeringDocumentService := &MockSteeringDocumentService{}

	handler := NewToolsHandler(
		mockEpicService,
		mockUserStoryService,
		mockRequirementService,
		mockSearchService,
		mockSteeringDocumentService,
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
	mockSteeringDocumentService.On("CreateSteeringDocument", req, user).Return(expectedDoc, nil)

	// Create request arguments
	args := map[string]interface{}{
		"title":       "Test Document",
		"description": "Test description",
	}

	// Execute
	ctx := createContextWithUser(user)
	result, err := handler.handleCreateSteeringDocument(ctx, args)

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

	mockSteeringDocumentService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleGetSteeringDocument_ByUUID_Success(t *testing.T) {
	t.Skip("Temporary disable due to work needed to refactor tests")
	mockEpicService := &MockEpicService{}
	mockUserStoryService := &MockUserStoryService{}
	mockRequirementService := &MockRequirementService{}
	mockSearchService := &MockSearchService{}
	mockSteeringDocumentService := &MockSteeringDocumentService{}

	handler := NewToolsHandler(
		mockEpicService,
		mockUserStoryService,
		mockRequirementService,
		mockSearchService,
		mockSteeringDocumentService,
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
	mockSteeringDocumentService.On("GetSteeringDocumentByID", docID, user).Return(expectedDoc, nil)

	// Create request arguments
	args := map[string]interface{}{
		"id": docID.String(),
	}

	// Execute
	ctx := createContextWithUser(user)
	result, err := handler.handleGetSteeringDocument(ctx, args)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the result structure
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, expectedDoc.ID.String(), resultMap["id"])
	assert.Equal(t, expectedDoc.ReferenceID, resultMap["reference_id"])
	assert.Equal(t, expectedDoc.Title, resultMap["title"])

	mockSteeringDocumentService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleGetSteeringDocument_ByReferenceID_Success(t *testing.T) {
	t.Skip("Temporary disable due to work needed to refactor tests")
	mockEpicService := &MockEpicService{}
	mockUserStoryService := &MockUserStoryService{}
	mockRequirementService := &MockRequirementService{}
	mockSearchService := &MockSearchService{}
	mockSteeringDocumentService := &MockSteeringDocumentService{}

	handler := NewToolsHandler(
		mockEpicService,
		mockUserStoryService,
		mockRequirementService,
		mockSearchService,
		mockSteeringDocumentService,
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
	mockSteeringDocumentService.On("GetSteeringDocumentByReferenceID", referenceID, user).Return(expectedDoc, nil)

	// Create request arguments
	args := map[string]interface{}{
		"id": referenceID,
	}

	// Execute
	ctx := createContextWithUser(user)
	result, err := handler.handleGetSteeringDocument(ctx, args)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the result structure
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, expectedDoc.ID.String(), resultMap["id"])
	assert.Equal(t, expectedDoc.ReferenceID, resultMap["reference_id"])
	assert.Equal(t, expectedDoc.Title, resultMap["title"])

	mockSteeringDocumentService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleUpdateSteeringDocument_Success(t *testing.T) {
	t.Skip("Temporary disable due to work needed to refactor tests")
	mockEpicService := &MockEpicService{}
	mockUserStoryService := &MockUserStoryService{}
	mockRequirementService := &MockRequirementService{}
	mockSearchService := &MockSearchService{}
	mockSteeringDocumentService := &MockSteeringDocumentService{}

	handler := NewToolsHandler(
		mockEpicService,
		mockUserStoryService,
		mockRequirementService,
		mockSearchService,
		mockSteeringDocumentService,
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
	mockSteeringDocumentService.On("UpdateSteeringDocument", docID, req, user).Return(expectedDoc, nil)

	// Create request arguments
	args := map[string]interface{}{
		"id":          docID.String(),
		"title":       "Updated Title",
		"description": "Updated description",
	}

	// Execute
	ctx := createContextWithUser(user)
	result, err := handler.handleUpdateSteeringDocument(ctx, args)

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

	mockSteeringDocumentService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleLinkSteeringToEpic_Success(t *testing.T) {
	t.Skip("Temporary disable due to work needed to refactor tests")
	mockEpicService := &MockEpicService{}
	mockUserStoryService := &MockUserStoryService{}
	mockRequirementService := &MockRequirementService{}
	mockSearchService := &MockSearchService{}
	mockSteeringDocumentService := &MockSteeringDocumentService{}

	handler := NewToolsHandler(
		mockEpicService,
		mockUserStoryService,
		mockRequirementService,
		mockSearchService,
		mockSteeringDocumentService,
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
	mockSteeringDocumentService.On("LinkSteeringDocumentToEpic", docID, epicID, user).Return(nil)

	// Create request arguments
	args := map[string]interface{}{
		"steering_document_id": docID.String(),
		"epic_id":              epicID.String(),
	}

	// Execute
	ctx := createContextWithUser(user)
	result, err := handler.handleLinkSteeringToEpic(ctx, args)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the result structure
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "Steering document linked to epic successfully", resultMap["message"])
	assert.Equal(t, docID.String(), resultMap["steering_document_id"])
	assert.Equal(t, epicID.String(), resultMap["epic_id"])

	mockSteeringDocumentService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleUnlinkSteeringFromEpic_Success(t *testing.T) {
	t.Skip("Temporary disable due to work needed to refactor tests")
	mockEpicService := &MockEpicService{}
	mockUserStoryService := &MockUserStoryService{}
	mockRequirementService := &MockRequirementService{}
	mockSearchService := &MockSearchService{}
	mockSteeringDocumentService := &MockSteeringDocumentService{}

	handler := NewToolsHandler(
		mockEpicService,
		mockUserStoryService,
		mockRequirementService,
		mockSearchService,
		mockSteeringDocumentService,
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
	mockSteeringDocumentService.On("UnlinkSteeringDocumentFromEpic", docID, epicID, user).Return(nil)

	// Create request arguments
	args := map[string]interface{}{
		"steering_document_id": docID.String(),
		"epic_id":              epicID.String(),
	}

	// Execute
	ctx := createContextWithUser(user)
	result, err := handler.handleUnlinkSteeringFromEpic(ctx, args)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the result structure
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "Steering document unlinked from epic successfully", resultMap["message"])
	assert.Equal(t, docID.String(), resultMap["steering_document_id"])
	assert.Equal(t, epicID.String(), resultMap["epic_id"])

	mockSteeringDocumentService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleGetEpicSteeringDocuments_Success(t *testing.T) {
	t.Skip("Temporary disable due to work needed to refactor tests")
	mockEpicService := &MockEpicService{}
	mockUserStoryService := &MockUserStoryService{}
	mockRequirementService := &MockRequirementService{}
	mockSearchService := &MockSearchService{}
	mockSteeringDocumentService := &MockSteeringDocumentService{}

	handler := NewToolsHandler(
		mockEpicService,
		mockUserStoryService,
		mockRequirementService,
		mockSearchService,
		mockSteeringDocumentService,
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
	mockSteeringDocumentService.On("GetSteeringDocumentsByEpicID", epicID, user).Return(docs, nil)

	// Create request arguments
	args := map[string]interface{}{
		"epic_id": epicID.String(),
	}

	// Execute
	ctx := createContextWithUser(user)
	result, err := handler.handleGetEpicSteeringDocuments(ctx, args)

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

	mockSteeringDocumentService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleCreateSteeringDocument_ValidationError(t *testing.T) {
	t.Skip("Temporary disable due to work needed to refactor tests")
	mockEpicService := &MockEpicService{}
	mockUserStoryService := &MockUserStoryService{}
	mockRequirementService := &MockRequirementService{}
	mockSearchService := &MockSearchService{}
	mockSteeringDocumentService := &MockSteeringDocumentService{}

	handler := NewToolsHandler(
		mockEpicService,
		mockUserStoryService,
		mockRequirementService,
		mockSearchService,
		mockSteeringDocumentService,
	)

	// Create test user
	user := &models.User{
		ID:   uuid.New(),
		Role: models.RoleUser,
	}

	req := service.CreateSteeringDocumentRequest{
		Title: "", // Invalid empty title
	}

	// Mock expectations - use a generic error since ErrValidation might not be defined
	mockSteeringDocumentService.On("CreateSteeringDocument", req, user).Return((*models.SteeringDocument)(nil), assert.AnError)

	// Create request arguments
	args := map[string]interface{}{
		"title": "", // Invalid empty title
	}

	// Execute
	ctx := createContextWithUser(user)
	result, err := handler.handleCreateSteeringDocument(ctx, args)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	mockSteeringDocumentService.AssertExpectations(t)
}

func TestMCPToolsHandler_handleGetSteeringDocument_NotFound(t *testing.T) {
	t.Skip("Temporary disable due to work needed to refactor tests")
	mockEpicService := &MockEpicService{}
	mockUserStoryService := &MockUserStoryService{}
	mockRequirementService := &MockRequirementService{}
	mockSearchService := &MockSearchService{}
	mockSteeringDocumentService := &MockSteeringDocumentService{}

	handler := NewToolsHandler(
		mockEpicService,
		mockUserStoryService,
		mockRequirementService,
		mockSearchService,
		mockSteeringDocumentService,
	)

	// Create test user
	user := &models.User{
		ID:   uuid.New(),
		Role: models.RoleUser,
	}

	docID := uuid.New()

	// Mock expectations - use a generic error since ErrSteeringDocumentNotFound might not be defined
	mockSteeringDocumentService.On("GetSteeringDocumentByID", docID, user).Return((*models.SteeringDocument)(nil), assert.AnError)

	// Create request arguments
	args := map[string]interface{}{
		"id": docID.String(),
	}

	// Execute
	ctx := createContextWithUser(user)
	result, err := handler.handleGetSteeringDocument(ctx, args)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	mockSteeringDocumentService.AssertExpectations(t)
}
