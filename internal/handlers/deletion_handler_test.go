package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"product-requirements-management/internal/service"
)

// Mock DeletionService for testing
type MockDeletionService struct {
	mock.Mock
}

func (m *MockDeletionService) DeleteEpicWithValidation(id uuid.UUID, userID uuid.UUID, force bool) (*service.DeletionResult, error) {
	args := m.Called(id, userID, force)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.DeletionResult), args.Error(1)
}

func (m *MockDeletionService) DeleteUserStoryWithValidation(id uuid.UUID, userID uuid.UUID, force bool) (*service.DeletionResult, error) {
	args := m.Called(id, userID, force)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.DeletionResult), args.Error(1)
}

func (m *MockDeletionService) DeleteAcceptanceCriteriaWithValidation(id uuid.UUID, userID uuid.UUID, force bool) (*service.DeletionResult, error) {
	args := m.Called(id, userID, force)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.DeletionResult), args.Error(1)
}

func (m *MockDeletionService) DeleteRequirementWithValidation(id uuid.UUID, userID uuid.UUID, force bool) (*service.DeletionResult, error) {
	args := m.Called(id, userID, force)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.DeletionResult), args.Error(1)
}

func (m *MockDeletionService) ValidateEpicDeletion(id uuid.UUID) (*service.DependencyInfo, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.DependencyInfo), args.Error(1)
}

func (m *MockDeletionService) ValidateUserStoryDeletion(id uuid.UUID) (*service.DependencyInfo, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.DependencyInfo), args.Error(1)
}

func (m *MockDeletionService) ValidateAcceptanceCriteriaDeletion(id uuid.UUID) (*service.DependencyInfo, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.DependencyInfo), args.Error(1)
}

func (m *MockDeletionService) ValidateRequirementDeletion(id uuid.UUID) (*service.DependencyInfo, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.DependencyInfo), args.Error(1)
}

// Test setup helper
func setupDeletionHandlerTest() (*DeletionHandler, *MockDeletionService, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	
	mockService := &MockDeletionService{}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests
	
	handler := NewDeletionHandler(mockService, logger)
	
	router := gin.New()
	
	// Add middleware to set user_id for authenticated routes
	router.Use(func(c *gin.Context) {
		// Set a test user ID for all requests
		c.Set("user_id", uuid.New())
		c.Next()
	})
	
	// Register routes
	router.GET("/api/epics/:id/validate-deletion", handler.ValidateEpicDeletion)
	router.DELETE("/api/epics/:id/delete", handler.DeleteEpic)
	router.GET("/api/user-stories/:id/validate-deletion", handler.ValidateUserStoryDeletion)
	router.DELETE("/api/user-stories/:id/delete", handler.DeleteUserStory)
	router.GET("/api/acceptance-criteria/:id/validate-deletion", handler.ValidateAcceptanceCriteriaDeletion)
	router.DELETE("/api/acceptance-criteria/:id/delete", handler.DeleteAcceptanceCriteria)
	router.GET("/api/requirements/:id/validate-deletion", handler.ValidateRequirementDeletion)
	router.DELETE("/api/requirements/:id/delete", handler.DeleteRequirement)
	router.GET("/api/deletion/confirm", handler.GetDeletionConfirmation)
	
	return handler, mockService, router
}

// Test ValidateEpicDeletion
func TestValidateEpicDeletion_Success(t *testing.T) {
	_, mockService, router := setupDeletionHandlerTest()
	
	epicID := uuid.New()
	depInfo := &service.DependencyInfo{
		CanDelete:            true,
		Dependencies:         []service.DependencyDetail{},
		CascadeDeleteCount:   0,
		RequiresConfirmation: false,
	}
	
	mockService.On("ValidateEpicDeletion", epicID).Return(depInfo, nil)
	
	req, _ := http.NewRequest("GET", "/api/epics/"+epicID.String()+"/validate-deletion", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response service.DependencyInfo
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.CanDelete)
	assert.Empty(t, response.Dependencies)
	
	mockService.AssertExpectations(t)
}

func TestValidateEpicDeletion_EpicNotFound(t *testing.T) {
	_, mockService, router := setupDeletionHandlerTest()
	
	epicID := uuid.New()
	
	mockService.On("ValidateEpicDeletion", epicID).Return(nil, service.ErrEpicNotFound)
	
	req, _ := http.NewRequest("GET", "/api/epics/"+epicID.String()+"/validate-deletion", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "EPIC_NOT_FOUND", response.Error.Code)
	
	mockService.AssertExpectations(t)
}

func TestValidateEpicDeletion_InvalidID(t *testing.T) {
	_, _, router := setupDeletionHandlerTest()
	
	req, _ := http.NewRequest("GET", "/api/epics/invalid-id/validate-deletion", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "INVALID_ID", response.Error.Code)
}

// Test DeleteEpic
func TestDeleteEpic_Success(t *testing.T) {
	_, mockService, router := setupDeletionHandlerTest()
	
	epicID := uuid.New()
	userID := uuid.New()
	
	// Mock deletion
	deletionResult := &service.DeletionResult{
		EntityType:     "epic",
		EntityID:       epicID,
		ReferenceID:    "EP-001",
		DeletedBy:      userID,
		CascadeDeleted: []service.CascadeDeletedEntity{},
		TransactionID:  "test_transaction",
	}
	mockService.On("DeleteEpicWithValidation", epicID, mock.AnythingOfType("uuid.UUID"), false).Return(deletionResult, nil)
	
	reqBody := DeleteEpicRequest{Force: false}
	jsonBody, _ := json.Marshal(reqBody)
	
	req, _ := http.NewRequest("DELETE", "/api/epics/"+epicID.String()+"/delete", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response service.DeletionResult
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "epic", response.EntityType)
	assert.Equal(t, epicID, response.EntityID)
	assert.Equal(t, "EP-001", response.ReferenceID)
	
	mockService.AssertExpectations(t)
}

func TestDeleteEpic_ValidationFailed(t *testing.T) {
	_, mockService, router := setupDeletionHandlerTest()
	
	epicID := uuid.New()
	
	// Mock deletion failure
	mockService.On("DeleteEpicWithValidation", epicID, mock.AnythingOfType("uuid.UUID"), false).Return(nil, service.ErrDeletionValidationFailed)
	
	reqBody := DeleteEpicRequest{Force: false}
	jsonBody, _ := json.Marshal(reqBody)
	
	req, _ := http.NewRequest("DELETE", "/api/epics/"+epicID.String()+"/delete", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusConflict, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "DELETION_BLOCKED", response.Error.Code)
	
	mockService.AssertExpectations(t)
}

func TestDeleteEpic_ForceDelete(t *testing.T) {
	_, mockService, router := setupDeletionHandlerTest()
	
	epicID := uuid.New()
	userID := uuid.New()
	
	// Mock force deletion success
	deletionResult := &service.DeletionResult{
		EntityType:     "epic",
		EntityID:       epicID,
		ReferenceID:    "EP-001",
		DeletedBy:      userID,
		CascadeDeleted: []service.CascadeDeletedEntity{{EntityType: "user_story", EntityID: uuid.New(), ReferenceID: "US-001"}},
		TransactionID:  "test_transaction",
	}
	mockService.On("DeleteEpicWithValidation", epicID, mock.AnythingOfType("uuid.UUID"), true).Return(deletionResult, nil)
	
	reqBody := DeleteEpicRequest{Force: true}
	jsonBody, _ := json.Marshal(reqBody)
	
	req, _ := http.NewRequest("DELETE", "/api/epics/"+epicID.String()+"/delete", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response service.DeletionResult
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "epic", response.EntityType)
	assert.Len(t, response.CascadeDeleted, 1)
	
	mockService.AssertExpectations(t)
}

// Test ValidateUserStoryDeletion
func TestValidateUserStoryDeletion_Success(t *testing.T) {
	_, mockService, router := setupDeletionHandlerTest()
	
	userStoryID := uuid.New()
	depInfo := &service.DependencyInfo{
		CanDelete:            true,
		Dependencies:         []service.DependencyDetail{},
		CascadeDeleteCount:   0,
		RequiresConfirmation: false,
	}
	
	mockService.On("ValidateUserStoryDeletion", userStoryID).Return(depInfo, nil)
	
	req, _ := http.NewRequest("GET", "/api/user-stories/"+userStoryID.String()+"/validate-deletion", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response service.DependencyInfo
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.CanDelete)
	
	mockService.AssertExpectations(t)
}

// Test DeleteUserStory
func TestDeleteUserStory_Success(t *testing.T) {
	_, mockService, router := setupDeletionHandlerTest()
	
	userStoryID := uuid.New()
	userID := uuid.New()
	
	// Mock deletion
	deletionResult := &service.DeletionResult{
		EntityType:     "user_story",
		EntityID:       userStoryID,
		ReferenceID:    "US-001",
		DeletedBy:      userID,
		CascadeDeleted: []service.CascadeDeletedEntity{},
		TransactionID:  "test_transaction",
	}
	mockService.On("DeleteUserStoryWithValidation", userStoryID, mock.AnythingOfType("uuid.UUID"), false).Return(deletionResult, nil)
	
	reqBody := DeleteUserStoryRequest{Force: false}
	jsonBody, _ := json.Marshal(reqBody)
	
	req, _ := http.NewRequest("DELETE", "/api/user-stories/"+userStoryID.String()+"/delete", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response service.DeletionResult
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "user_story", response.EntityType)
	assert.Equal(t, userStoryID, response.EntityID)
	
	mockService.AssertExpectations(t)
}

// Test ValidateAcceptanceCriteriaDeletion
func TestValidateAcceptanceCriteriaDeletion_Success(t *testing.T) {
	_, mockService, router := setupDeletionHandlerTest()
	
	acceptanceCriteriaID := uuid.New()
	depInfo := &service.DependencyInfo{
		CanDelete:            true,
		Dependencies:         []service.DependencyDetail{},
		CascadeDeleteCount:   0,
		RequiresConfirmation: false,
	}
	
	mockService.On("ValidateAcceptanceCriteriaDeletion", acceptanceCriteriaID).Return(depInfo, nil)
	
	req, _ := http.NewRequest("GET", "/api/acceptance-criteria/"+acceptanceCriteriaID.String()+"/validate-deletion", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response service.DependencyInfo
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.CanDelete)
	
	mockService.AssertExpectations(t)
}

// Test ValidateRequirementDeletion
func TestValidateRequirementDeletion_Success(t *testing.T) {
	_, mockService, router := setupDeletionHandlerTest()
	
	requirementID := uuid.New()
	depInfo := &service.DependencyInfo{
		CanDelete:            true,
		Dependencies:         []service.DependencyDetail{},
		CascadeDeleteCount:   0,
		RequiresConfirmation: false,
	}
	
	mockService.On("ValidateRequirementDeletion", requirementID).Return(depInfo, nil)
	
	req, _ := http.NewRequest("GET", "/api/requirements/"+requirementID.String()+"/validate-deletion", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response service.DependencyInfo
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.CanDelete)
	
	mockService.AssertExpectations(t)
}

// Test GetDeletionConfirmation
func TestGetDeletionConfirmation_Epic_Success(t *testing.T) {
	_, mockService, router := setupDeletionHandlerTest()
	
	epicID := uuid.New()
	depInfo := &service.DependencyInfo{
		CanDelete:            false,
		Dependencies:         []service.DependencyDetail{{EntityType: "user_story"}},
		CascadeDeleteCount:   1,
		RequiresConfirmation: true,
	}
	
	mockService.On("ValidateEpicDeletion", epicID).Return(depInfo, nil)
	
	req, _ := http.NewRequest("GET", "/api/deletion/confirm?entity_type=epic&id="+epicID.String(), nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response service.DependencyInfo
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.CanDelete)
	assert.Len(t, response.Dependencies, 1)
	assert.True(t, response.RequiresConfirmation)
	
	mockService.AssertExpectations(t)
}

func TestGetDeletionConfirmation_MissingParameters(t *testing.T) {
	_, _, router := setupDeletionHandlerTest()
	
	req, _ := http.NewRequest("GET", "/api/deletion/confirm", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "MISSING_PARAMETERS", response.Error.Code)
}

func TestGetDeletionConfirmation_InvalidEntityType(t *testing.T) {
	_, _, router := setupDeletionHandlerTest()
	
	epicID := uuid.New()
	req, _ := http.NewRequest("GET", "/api/deletion/confirm?entity_type=invalid&id="+epicID.String(), nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "INVALID_ENTITY_TYPE", response.Error.Code)
}

func TestGetDeletionConfirmation_InvalidID(t *testing.T) {
	_, _, router := setupDeletionHandlerTest()
	
	req, _ := http.NewRequest("GET", "/api/deletion/confirm?entity_type=epic&id=invalid-id", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "INVALID_ID", response.Error.Code)
}

// Test authentication middleware behavior
func TestDeleteEpic_NoAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockService := &MockDeletionService{}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	handler := NewDeletionHandler(mockService, logger)
	
	router := gin.New()
	// Don't add the auth middleware that sets user_id
	router.DELETE("/api/epics/:id/delete", handler.DeleteEpic)
	
	epicID := uuid.New()
	reqBody := DeleteEpicRequest{Force: false}
	jsonBody, _ := json.Marshal(reqBody)
	
	req, _ := http.NewRequest("DELETE", "/api/epics/"+epicID.String()+"/delete", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "UNAUTHORIZED", response.Error.Code)
}