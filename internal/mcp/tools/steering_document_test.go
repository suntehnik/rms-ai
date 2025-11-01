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

// MockSteeringDocumentService is a mock implementation of service.SteeringDocumentService
type MockSteeringDocumentService struct {
	mock.Mock
}

func (m *MockSteeringDocumentService) CreateSteeringDocument(req service.CreateSteeringDocumentRequest, currentUser *models.User) (*models.SteeringDocument, error) {
	args := m.Called(req, currentUser)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SteeringDocument), args.Error(1)
}

func (m *MockSteeringDocumentService) GetSteeringDocumentByID(id uuid.UUID, currentUser *models.User) (*models.SteeringDocument, error) {
	args := m.Called(id, currentUser)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SteeringDocument), args.Error(1)
}

func (m *MockSteeringDocumentService) GetSteeringDocumentByReferenceID(referenceID string, currentUser *models.User) (*models.SteeringDocument, error) {
	args := m.Called(referenceID, currentUser)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SteeringDocument), args.Error(1)
}

func (m *MockSteeringDocumentService) UpdateSteeringDocument(id uuid.UUID, req service.UpdateSteeringDocumentRequest, currentUser *models.User) (*models.SteeringDocument, error) {
	args := m.Called(id, req, currentUser)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SteeringDocument), args.Error(1)
}

func (m *MockSteeringDocumentService) DeleteSteeringDocument(id uuid.UUID, currentUser *models.User) error {
	args := m.Called(id, currentUser)
	return args.Error(0)
}

func (m *MockSteeringDocumentService) ListSteeringDocuments(filters service.SteeringDocumentFilters, currentUser *models.User) ([]models.SteeringDocument, int64, error) {
	args := m.Called(filters, currentUser)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.SteeringDocument), args.Get(1).(int64), args.Error(2)
}

func (m *MockSteeringDocumentService) SearchSteeringDocuments(query string, currentUser *models.User) ([]models.SteeringDocument, error) {
	args := m.Called(query, currentUser)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SteeringDocument), args.Error(1)
}

func (m *MockSteeringDocumentService) GetSteeringDocumentsByEpicID(epicID uuid.UUID, currentUser *models.User) ([]models.SteeringDocument, error) {
	args := m.Called(epicID, currentUser)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SteeringDocument), args.Error(1)
}

func (m *MockSteeringDocumentService) GetSteeringDocumentsByEpicIDWithPagination(epicID uuid.UUID, limit, offset int, currentUser *models.User) ([]models.SteeringDocument, int64, error) {
	args := m.Called(epicID, limit, offset, currentUser)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.SteeringDocument), args.Get(1).(int64), args.Error(2)
}

func (m *MockSteeringDocumentService) LinkSteeringDocumentToEpic(steeringDocumentID, epicID uuid.UUID, currentUser *models.User) error {
	args := m.Called(steeringDocumentID, epicID, currentUser)
	return args.Error(0)
}

func (m *MockSteeringDocumentService) UnlinkSteeringDocumentFromEpic(steeringDocumentID, epicID uuid.UUID, currentUser *models.User) error {
	args := m.Called(steeringDocumentID, epicID, currentUser)
	return args.Error(0)
}

// createTestContext creates a test context with a mock user
func createTestContext() context.Context {
	gin.SetMode(gin.TestMode)
	ginCtx, _ := gin.CreateTestContext(nil)

	// Create a test user
	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	// Set user in gin context
	ginCtx.Set("user", testUser)

	// Create context with gin context
	ctx := context.WithValue(context.Background(), "gin_context", ginCtx)
	return ctx
}

func TestSteeringDocumentHandler_GetSupportedTools(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)

	tools := handler.GetSupportedTools()

	expectedTools := []string{
		"list_steering_documents",
		"create_steering_document",
		"get_steering_document",
		"update_steering_document",
		"link_steering_to_epic",
		"unlink_steering_from_epic",
		"get_epic_steering_documents",
	}

	assert.Equal(t, expectedTools, tools)
}

func TestSteeringDocumentHandler_HandleTool_UnknownTool(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	result, err := handler.HandleTool(ctx, "unknown_tool", map[string]interface{}{})

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Method not found")
}

func TestSteeringDocumentHandler_CreateSteeringDocument_Success(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	// Test data
	description := "Test Description"
	testDoc := &models.SteeringDocument{
		ID:          uuid.New(),
		ReferenceID: "STD-001",
		Title:       "Test Document",
		Description: &description,
	}

	args := map[string]interface{}{
		"title":       "Test Document",
		"description": "Test Description",
	}

	// Mock expectations
	mockSteeringService.On("CreateSteeringDocument", mock.AnythingOfType("service.CreateSteeringDocumentRequest"), mock.AnythingOfType("*models.User")).Return(testDoc, nil)

	// Execute
	result, err := handler.CreateSteeringDocument(ctx, args)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	response, ok := result.(*types.ToolResponse)
	assert.True(t, ok)
	assert.Len(t, response.Content, 2)
	assert.Contains(t, response.Content[0].Text, "Successfully created steering document STD-001")

	mockSteeringService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_CreateSteeringDocument_MissingTitle(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	args := map[string]interface{}{
		"description": "Test Description",
	}

	// Execute
	result, err := handler.CreateSteeringDocument(ctx, args)

	// Assertions
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid params")
}

func TestSteeringDocumentHandler_GetSteeringDocument_ByUUID_Success(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	// Test data
	testID := uuid.New()
	testDoc := &models.SteeringDocument{
		ID:          testID,
		ReferenceID: "STD-001",
		Title:       "Test Document",
	}

	args := map[string]interface{}{
		"steering_document_id": testID.String(),
	}

	// Mock expectations
	mockSteeringService.On("GetSteeringDocumentByID", testID, mock.AnythingOfType("*models.User")).Return(testDoc, nil)

	// Execute
	result, err := handler.GetSteeringDocument(ctx, args)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	response, ok := result.(*types.ToolResponse)
	assert.True(t, ok)
	assert.Len(t, response.Content, 2)
	assert.Contains(t, response.Content[0].Text, "Retrieved steering document STD-001")

	mockSteeringService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_GetSteeringDocument_ByReferenceID_Success(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	// Test data
	testDoc := &models.SteeringDocument{
		ID:          uuid.New(),
		ReferenceID: "STD-001",
		Title:       "Test Document",
	}

	args := map[string]interface{}{
		"steering_document_id": "STD-001",
	}

	// Mock expectations
	mockSteeringService.On("GetSteeringDocumentByReferenceID", "STD-001", mock.AnythingOfType("*models.User")).Return(testDoc, nil)

	// Execute
	result, err := handler.GetSteeringDocument(ctx, args)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	response, ok := result.(*types.ToolResponse)
	assert.True(t, ok)
	assert.Contains(t, response.Content[0].Text, "Retrieved steering document STD-001")

	mockSteeringService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_UpdateSteeringDocument_Success(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	// Test data
	testID := uuid.New()
	testDoc := &models.SteeringDocument{
		ID:          testID,
		ReferenceID: "STD-001",
		Title:       "Updated Document",
	}

	args := map[string]interface{}{
		"steering_document_id": testID.String(),
		"title":                "Updated Document",
	}

	// Mock expectations
	mockSteeringService.On("UpdateSteeringDocument", testID, mock.AnythingOfType("service.UpdateSteeringDocumentRequest"), mock.AnythingOfType("*models.User")).Return(testDoc, nil)

	// Execute
	result, err := handler.UpdateSteeringDocument(ctx, args)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	response, ok := result.(*types.ToolResponse)
	assert.True(t, ok)
	assert.Contains(t, response.Content[0].Text, "Successfully updated steering document STD-001")

	mockSteeringService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_LinkSteeringToEpic_Success(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	// Test data
	steeringDocID := uuid.New()
	epicID := uuid.New()

	args := map[string]interface{}{
		"steering_document_id": steeringDocID.String(),
		"epic_id":              epicID.String(),
	}

	// Mock expectations
	mockSteeringService.On("LinkSteeringDocumentToEpic", steeringDocID, epicID, mock.AnythingOfType("*models.User")).Return(nil)

	// Execute
	result, err := handler.LinkSteeringToEpic(ctx, args)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	response, ok := result.(*types.ToolResponse)
	assert.True(t, ok)
	assert.Contains(t, response.Content[0].Text, "Successfully linked steering document")

	mockSteeringService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_LinkSteeringToEpic_WithReferenceIDs_Success(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	// Test data
	steeringDocID := uuid.New()
	epicID := uuid.New()

	testDoc := &models.SteeringDocument{
		ID:          steeringDocID,
		ReferenceID: "STD-001",
	}

	testEpic := &models.Epic{
		ID:          epicID,
		ReferenceID: "EP-001",
	}

	args := map[string]interface{}{
		"steering_document_id": "STD-001",
		"epic_id":              "EP-001",
	}

	// Mock expectations
	mockSteeringService.On("GetSteeringDocumentByReferenceID", "STD-001", mock.AnythingOfType("*models.User")).Return(testDoc, nil)
	mockEpicService.On("GetEpicByReferenceID", "EP-001").Return(testEpic, nil)
	mockSteeringService.On("LinkSteeringDocumentToEpic", steeringDocID, epicID, mock.AnythingOfType("*models.User")).Return(nil)

	// Execute
	result, err := handler.LinkSteeringToEpic(ctx, args)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	response, ok := result.(*types.ToolResponse)
	assert.True(t, ok)
	assert.Contains(t, response.Content[0].Text, "Successfully linked steering document STD-001 to epic EP-001")

	mockSteeringService.AssertExpectations(t)
	mockEpicService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_UnlinkSteeringFromEpic_Success(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	// Test data
	steeringDocID := uuid.New()
	epicID := uuid.New()

	args := map[string]interface{}{
		"steering_document_id": steeringDocID.String(),
		"epic_id":              epicID.String(),
	}

	// Mock expectations
	mockSteeringService.On("UnlinkSteeringDocumentFromEpic", steeringDocID, epicID, mock.AnythingOfType("*models.User")).Return(nil)

	// Execute
	result, err := handler.UnlinkSteeringFromEpic(ctx, args)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	response, ok := result.(*types.ToolResponse)
	assert.True(t, ok)
	assert.Contains(t, response.Content[0].Text, "Successfully unlinked steering document")

	mockSteeringService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_GetEpicSteeringDocuments_Success(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	// Test data
	epicID := uuid.New()
	testDocs := []models.SteeringDocument{
		{
			ID:          uuid.New(),
			ReferenceID: "STD-001",
			Title:       "Document 1",
		},
		{
			ID:          uuid.New(),
			ReferenceID: "STD-002",
			Title:       "Document 2",
		},
	}

	args := map[string]interface{}{
		"epic_id": epicID.String(),
	}

	// Mock expectations
	mockSteeringService.On("GetSteeringDocumentsByEpicID", epicID, mock.AnythingOfType("*models.User")).Return(testDocs, nil)

	// Execute
	result, err := handler.GetEpicSteeringDocuments(ctx, args)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	response, ok := result.(*types.ToolResponse)
	assert.True(t, ok)
	assert.Contains(t, response.Content[0].Text, "Found 2 steering documents linked to epic")

	mockSteeringService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_ListSteeringDocuments_Success(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	// Test data
	testDocs := []models.SteeringDocument{
		{
			ID:          uuid.New(),
			ReferenceID: "STD-001",
			Title:       "Document 1",
		},
	}

	args := map[string]interface{}{
		"limit":  float64(10),
		"offset": float64(0),
	}

	// Mock expectations
	mockSteeringService.On("ListSteeringDocuments", mock.AnythingOfType("service.SteeringDocumentFilters"), mock.AnythingOfType("*models.User")).Return(testDocs, int64(1), nil)

	// Execute
	result, err := handler.ListSteeringDocuments(ctx, args)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	response, ok := result.(*types.ToolResponse)
	assert.True(t, ok)
	assert.Contains(t, response.Content[0].Text, "Found 1 steering documents (total: 1)")

	mockSteeringService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_ServiceError_Handling(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	args := map[string]interface{}{
		"title": "Test Document",
	}

	// Mock service error
	mockSteeringService.On("CreateSteeringDocument", mock.AnythingOfType("service.CreateSteeringDocumentRequest"), mock.AnythingOfType("*models.User")).Return(nil, errors.New("service error"))

	// Execute
	result, err := handler.CreateSteeringDocument(ctx, args)

	// Assertions
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Internal error")

	mockSteeringService.AssertExpectations(t)
}

// Additional test cases for comprehensive coverage

func TestSteeringDocumentHandler_GetSteeringDocument_MissingID(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	args := map[string]interface{}{}

	// Execute
	result, err := handler.GetSteeringDocument(ctx, args)

	// Assertions
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid params")
}

func TestSteeringDocumentHandler_UpdateSteeringDocument_InvalidID(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	args := map[string]interface{}{
		"steering_document_id": "INVALID-ID",
		"title":                "Updated Title",
	}

	// Mock service error for invalid reference ID
	mockSteeringService.On("GetSteeringDocumentByReferenceID", "INVALID-ID", mock.AnythingOfType("*models.User")).Return(nil, errors.New("not found"))

	// Execute
	result, err := handler.UpdateSteeringDocument(ctx, args)

	// Assertions
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid params")

	mockSteeringService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_LinkSteeringToEpic_MissingSteeringDocID(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	args := map[string]interface{}{
		"epic_id": uuid.New().String(),
	}

	// Execute
	result, err := handler.LinkSteeringToEpic(ctx, args)

	// Assertions
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid params")
}

func TestSteeringDocumentHandler_LinkSteeringToEpic_MissingEpicID(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	args := map[string]interface{}{
		"steering_document_id": uuid.New().String(),
	}

	// Execute
	result, err := handler.LinkSteeringToEpic(ctx, args)

	// Assertions
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid params")
}

func TestSteeringDocumentHandler_LinkSteeringToEpic_InvalidEpicID(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	steeringDocID := uuid.New()
	args := map[string]interface{}{
		"steering_document_id": steeringDocID.String(),
		"epic_id":              "INVALID-EPIC",
	}

	// Mock service error for invalid epic reference ID
	mockEpicService.On("GetEpicByReferenceID", "INVALID-EPIC").Return(nil, errors.New("epic not found"))

	// Execute
	result, err := handler.LinkSteeringToEpic(ctx, args)

	// Assertions
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid params")

	mockEpicService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_UnlinkSteeringFromEpic_InvalidSteeringDocID(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	epicID := uuid.New()
	args := map[string]interface{}{
		"steering_document_id": "INVALID-DOC",
		"epic_id":              epicID.String(),
	}

	// Mock service error for invalid steering document reference ID
	mockSteeringService.On("GetSteeringDocumentByReferenceID", "INVALID-DOC", mock.AnythingOfType("*models.User")).Return(nil, errors.New("document not found"))

	// Execute
	result, err := handler.UnlinkSteeringFromEpic(ctx, args)

	// Assertions
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid params")

	mockSteeringService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_GetEpicSteeringDocuments_ByReferenceID_Success(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	// Test data
	epicID := uuid.New()
	testEpic := &models.Epic{
		ID:          epicID,
		ReferenceID: "EP-001",
	}

	testDocs := []models.SteeringDocument{
		{
			ID:          uuid.New(),
			ReferenceID: "STD-001",
			Title:       "Document 1",
		},
	}

	args := map[string]interface{}{
		"epic_id": "EP-001",
	}

	// Mock expectations
	mockEpicService.On("GetEpicByReferenceID", "EP-001").Return(testEpic, nil)
	mockSteeringService.On("GetSteeringDocumentsByEpicID", epicID, mock.AnythingOfType("*models.User")).Return(testDocs, nil)

	// Execute
	result, err := handler.GetEpicSteeringDocuments(ctx, args)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	response, ok := result.(*types.ToolResponse)
	assert.True(t, ok)
	assert.Contains(t, response.Content[0].Text, "Found 1 steering documents linked to epic EP-001")

	mockEpicService.AssertExpectations(t)
	mockSteeringService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_GetEpicSteeringDocuments_InvalidEpicID(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	args := map[string]interface{}{
		"epic_id": "INVALID-EPIC",
	}

	// Mock service error for invalid epic reference ID
	mockEpicService.On("GetEpicByReferenceID", "INVALID-EPIC").Return(nil, errors.New("epic not found"))

	// Execute
	result, err := handler.GetEpicSteeringDocuments(ctx, args)

	// Assertions
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid params")

	mockEpicService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_ListSteeringDocuments_WithFilters(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	// Test data
	creatorID := uuid.New()
	testDocs := []models.SteeringDocument{
		{
			ID:          uuid.New(),
			ReferenceID: "STD-001",
			Title:       "Document 1",
		},
	}

	args := map[string]interface{}{
		"creator_id": creatorID.String(),
		"search":     "test search",
		"order_by":   "created_at DESC",
		"limit":      float64(5),
		"offset":     float64(10),
	}

	// Mock expectations - verify filters are passed correctly
	mockSteeringService.On("ListSteeringDocuments", mock.MatchedBy(func(filters service.SteeringDocumentFilters) bool {
		return filters.CreatorID != nil && *filters.CreatorID == creatorID &&
			filters.Search == "test search" &&
			filters.OrderBy == "created_at DESC" &&
			filters.Limit == 5 &&
			filters.Offset == 10
	}), mock.AnythingOfType("*models.User")).Return(testDocs, int64(1), nil)

	// Execute
	result, err := handler.ListSteeringDocuments(ctx, args)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	response, ok := result.(*types.ToolResponse)
	assert.True(t, ok)
	assert.Contains(t, response.Content[0].Text, "Found 1 steering documents (total: 1)")

	mockSteeringService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_ListSteeringDocuments_InvalidCreatorID(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	args := map[string]interface{}{
		"creator_id": "invalid-uuid",
	}

	// Execute
	result, err := handler.ListSteeringDocuments(ctx, args)

	// Assertions
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid params")
}

func TestSteeringDocumentHandler_CreateSteeringDocument_WithEpicLink(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	// Test data
	epicID := "EP-001"
	description := "Test Description"
	testDoc := &models.SteeringDocument{
		ID:          uuid.New(),
		ReferenceID: "STD-001",
		Title:       "Test Document",
		Description: &description,
	}

	args := map[string]interface{}{
		"title":       "Test Document",
		"description": "Test Description",
		"epic_id":     epicID,
	}

	// Mock expectations
	mockSteeringService.On("CreateSteeringDocument", mock.MatchedBy(func(req service.CreateSteeringDocumentRequest) bool {
		return req.Title == "Test Document" &&
			req.Description != nil && *req.Description == "Test Description" &&
			req.EpicID != nil && *req.EpicID == epicID
	}), mock.AnythingOfType("*models.User")).Return(testDoc, nil)

	// Execute
	result, err := handler.CreateSteeringDocument(ctx, args)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	response, ok := result.(*types.ToolResponse)
	assert.True(t, ok)
	assert.Contains(t, response.Content[0].Text, "Successfully created steering document STD-001")
	assert.Contains(t, response.Content[0].Text, "and linked to epic")

	mockSteeringService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_UpdateSteeringDocument_ByReferenceID(t *testing.T) {
	mockSteeringService := &MockSteeringDocumentService{}
	mockEpicService := &MockEpicService{}

	handler := NewSteeringDocumentHandler(mockSteeringService, mockEpicService)
	ctx := createTestContext()

	// Test data
	testID := uuid.New()
	testDoc := &models.SteeringDocument{
		ID:          testID,
		ReferenceID: "STD-001",
	}

	updatedDoc := &models.SteeringDocument{
		ID:          testID,
		ReferenceID: "STD-001",
		Title:       "Updated Document",
	}

	args := map[string]interface{}{
		"steering_document_id": "STD-001",
		"title":                "Updated Document",
	}

	// Mock expectations
	mockSteeringService.On("GetSteeringDocumentByReferenceID", "STD-001", mock.AnythingOfType("*models.User")).Return(testDoc, nil)
	mockSteeringService.On("UpdateSteeringDocument", testID, mock.AnythingOfType("service.UpdateSteeringDocumentRequest"), mock.AnythingOfType("*models.User")).Return(updatedDoc, nil)

	// Execute
	result, err := handler.UpdateSteeringDocument(ctx, args)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	response, ok := result.(*types.ToolResponse)
	assert.True(t, ok)
	assert.Contains(t, response.Content[0].Text, "Successfully updated steering document STD-001")

	mockSteeringService.AssertExpectations(t)
}
