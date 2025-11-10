package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"product-requirements-management/internal/auth"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// MockRequirementService is a mock implementation of RequirementService
type MockRequirementService struct {
	mock.Mock
}

func (m *MockRequirementService) CreateRequirement(req service.CreateRequirementRequest) (*models.Requirement, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementService) GetRequirementByID(id uuid.UUID) (*models.Requirement, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementService) GetRequirementByReferenceID(referenceID string) (*models.Requirement, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementService) UpdateRequirement(id uuid.UUID, req service.UpdateRequirementRequest) (*models.Requirement, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementService) DeleteRequirement(id uuid.UUID, force bool) error {
	args := m.Called(id, force)
	return args.Error(0)
}

func (m *MockRequirementService) ListRequirements(filters service.RequirementFilters) ([]models.Requirement, int64, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.Requirement), args.Get(1).(int64), args.Error(2)
}

func (m *MockRequirementService) GetRequirementWithRelationships(id uuid.UUID) (*models.Requirement, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementService) GetRequirementsByUserStory(userStoryID uuid.UUID) ([]models.Requirement, error) {
	args := m.Called(userStoryID)
	return args.Get(0).([]models.Requirement), args.Error(1)
}

func (m *MockRequirementService) ChangeRequirementStatus(id uuid.UUID, newStatus models.RequirementStatus) (*models.Requirement, error) {
	args := m.Called(id, newStatus)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementService) AssignRequirement(id uuid.UUID, assigneeID uuid.UUID) (*models.Requirement, error) {
	args := m.Called(id, assigneeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementService) CreateRelationship(req service.CreateRelationshipRequest) (*models.RequirementRelationship, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RequirementRelationship), args.Error(1)
}

func (m *MockRequirementService) DeleteRelationship(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRequirementService) GetRelationshipsByRequirement(requirementID uuid.UUID) ([]models.RequirementRelationship, error) {
	args := m.Called(requirementID)
	return args.Get(0).([]models.RequirementRelationship), args.Error(1)
}

func (m *MockRequirementService) SearchRequirements(searchText string) ([]models.Requirement, error) {
	args := m.Called(searchText)
	return args.Get(0).([]models.Requirement), args.Error(1)
}

func (m *MockRequirementService) GetRelationshipsByRequirementWithPagination(requirementID uuid.UUID, limit, offset int) ([]models.RequirementRelationship, int64, error) {
	args := m.Called(requirementID, limit, offset)
	return args.Get(0).([]models.RequirementRelationship), args.Get(1).(int64), args.Error(2)
}

func (m *MockRequirementService) SearchRequirementsWithPagination(searchText string, limit, offset int) ([]models.Requirement, int64, error) {
	args := m.Called(searchText, limit, offset)
	return args.Get(0).([]models.Requirement), args.Get(1).(int64), args.Error(2)
}

func setupRequirementTestRouter() (*gin.Engine, *MockRequirementService, *auth.Service) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockRequirementService)
	handler := NewRequirementHandler(mockService)

	router := gin.New()

	// Create auth service for testing
	mockRefreshTokenRepo := &mockRefreshTokenRepository{}
	authService := auth.NewService("test-secret", time.Hour, mockRefreshTokenRepo)

	v1 := router.Group("/api/v1")
	v1.Use(authService.Middleware()) // Add auth middleware
	{
		v1.POST("/requirements", handler.CreateRequirement)
		v1.GET("/requirements/:id", handler.GetRequirement)
		v1.PUT("/requirements/:id", handler.UpdateRequirement)
		v1.DELETE("/requirements/:id", handler.DeleteRequirement)
		v1.GET("/requirements", handler.ListRequirements)
		v1.GET("/requirements/search", handler.SearchRequirements)
		v1.GET("/requirements/:id/relationships", handler.GetRequirementWithRelationships)
		v1.POST("/requirements/relationships", handler.CreateRelationship)
		v1.DELETE("/requirement-relationships/:id", handler.DeleteRelationship)
		v1.PATCH("/requirements/:id/status", handler.ChangeRequirementStatus)
		v1.PATCH("/requirements/:id/assign", handler.AssignRequirement)
	}

	return router, mockService, authService
}

// Helper function to create authenticated request
func createAuthenticatedRequirementRequest(method, url string, body io.Reader, authService *auth.Service) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// Create test user and token
	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     models.RoleUser,
	}
	token, err := authService.GenerateToken(testUser)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// Helper function to create authenticated request with specific user
func createAuthenticatedRequirementRequestWithUser(method, url string, body io.Reader, authService *auth.Service, user *models.User) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	token, err := authService.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func TestRequirementHandler_CreateRequirement(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		router, mockService, authService := setupRequirementTestRouter()

		// Create test user
		testUser := &models.User{
			ID:       uuid.New(),
			Username: "testuser",
			Role:     models.RoleUser,
		}

		userStoryID := uuid.New()
		typeID := uuid.New()

		reqBody := service.CreateRequirementRequest{
			UserStoryID: userStoryID,
			CreatorID:   uuid.New(), // This will be overridden by the handler
			Priority:    models.PriorityHigh,
			TypeID:      typeID,
			Title:       "Test Requirement",
		}

		// Expected request with the authenticated user's ID
		expectedReqBody := service.CreateRequirementRequest{
			UserStoryID: userStoryID,
			CreatorID:   testUser.ID, // This is what the handler will set
			Priority:    models.PriorityHigh,
			TypeID:      typeID,
			Title:       "Test Requirement",
		}

		expectedRequirement := &models.Requirement{
			ID:          uuid.New(),
			UserStoryID: userStoryID,
			CreatorID:   testUser.ID,
			AssigneeID:  testUser.ID,
			Priority:    models.PriorityHigh,
			Status:      models.RequirementStatusDraft,
			TypeID:      typeID,
			Title:       "Test Requirement",
		}

		mockService.On("CreateRequirement", expectedReqBody).Return(expectedRequirement, nil)

		// Create authenticated request with specific user
		jsonBody, _ := json.Marshal(reqBody)
		req, err := createAuthenticatedRequirementRequestWithUser("POST", "/api/v1/requirements", bytes.NewBuffer(jsonBody), authService, testUser)
		assert.NoError(t, err)

		// Create response recorder and perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.Requirement
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequirement.ID, response.ID)
		assert.Equal(t, expectedRequirement.Title, response.Title)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		router, _, authService := setupRequirementTestRouter()

		// Create authenticated request with invalid body
		req, err := createAuthenticatedRequirementRequest("POST", "/api/v1/requirements", bytes.NewBuffer([]byte(`{"invalid": "json"}`)), authService)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Invalid request body")
	})

	t.Run("service error - user story not found", func(t *testing.T) {
		router, mockService, authService := setupRequirementTestRouter()

		// Create test user
		testUser := &models.User{
			ID:       uuid.New(),
			Username: "testuser",
			Role:     models.RoleUser,
		}

		reqBody := service.CreateRequirementRequest{
			UserStoryID: uuid.New(),
			CreatorID:   uuid.New(), // This will be overridden
			Priority:    models.PriorityHigh,
			TypeID:      uuid.New(),
			Title:       "Test Requirement",
		}

		// Expected request with the authenticated user's ID
		expectedReqBody := service.CreateRequirementRequest{
			UserStoryID: reqBody.UserStoryID,
			CreatorID:   testUser.ID, // This is what the handler will set
			Priority:    models.PriorityHigh,
			TypeID:      reqBody.TypeID,
			Title:       "Test Requirement",
		}

		mockService.On("CreateRequirement", expectedReqBody).Return(nil, service.ErrUserStoryNotFound)

		jsonBody, _ := json.Marshal(reqBody)
		req, err := createAuthenticatedRequirementRequestWithUser("POST", "/api/v1/requirements", bytes.NewBuffer(jsonBody), authService, testUser)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "User story not found", response["error"])

		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized request", func(t *testing.T) {
		router, _, _ := setupRequirementTestRouter()

		reqBody := service.CreateRequirementRequest{
			UserStoryID: uuid.New(),
			CreatorID:   uuid.New(),
			Priority:    models.PriorityHigh,
			TypeID:      uuid.New(),
			Title:       "Test Requirement",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/requirements", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRequirementHandler_GetRequirement(t *testing.T) {
	t.Run("successful retrieval by UUID", func(t *testing.T) {
		router, mockService, authService := setupRequirementTestRouter()

		requirementID := uuid.New()
		expectedRequirement := &models.Requirement{
			ID:    requirementID,
			Title: "Test Requirement",
		}

		mockService.On("GetRequirementByID", requirementID).Return(expectedRequirement, nil)

		req, err := createAuthenticatedRequirementRequest("GET", "/api/v1/requirements/"+requirementID.String(), nil, authService)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.Requirement
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequirement.ID, response.ID)
		assert.Equal(t, expectedRequirement.Title, response.Title)

		mockService.AssertExpectations(t)
	})

	t.Run("successful retrieval by reference ID", func(t *testing.T) {
		router, mockService, authService := setupRequirementTestRouter()

		referenceID := "REQ-001"
		expectedRequirement := &models.Requirement{
			ID:          uuid.New(),
			ReferenceID: referenceID,
			Title:       "Test Requirement",
		}

		mockService.On("GetRequirementByReferenceID", referenceID).Return(expectedRequirement, nil)

		req, err := createAuthenticatedRequirementRequest("GET", "/api/v1/requirements/"+referenceID, nil, authService)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.Requirement
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequirement.ID, response.ID)
		assert.Equal(t, expectedRequirement.ReferenceID, response.ReferenceID)

		mockService.AssertExpectations(t)
	})

	t.Run("requirement not found", func(t *testing.T) {
		router, mockService, authService := setupRequirementTestRouter()

		requirementID := uuid.New()

		mockService.On("GetRequirementByID", requirementID).Return(nil, service.ErrRequirementNotFound)

		req, err := createAuthenticatedRequirementRequest("GET", "/api/v1/requirements/"+requirementID.String(), nil, authService)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Requirement not found", response["error"])

		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized request", func(t *testing.T) {
		router, _, _ := setupRequirementTestRouter()

		requirementID := uuid.New()

		req, _ := http.NewRequest("GET", "/api/v1/requirements/"+requirementID.String(), nil)
		// No Authorization header

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRequirementHandler_DeleteRequirement(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		router, mockService, authService := setupRequirementTestRouter()

		requirementID := uuid.New()

		mockService.On("DeleteRequirement", requirementID, false).Return(nil)

		req, err := createAuthenticatedRequirementRequest("DELETE", "/api/v1/requirements/"+requirementID.String(), nil, authService)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("force deletion", func(t *testing.T) {
		router, mockService, authService := setupRequirementTestRouter()

		requirementID := uuid.New()

		mockService.On("DeleteRequirement", requirementID, true).Return(nil)

		req, err := createAuthenticatedRequirementRequest("DELETE", "/api/v1/requirements/"+requirementID.String()+"?force=true", nil, authService)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("deletion blocked by relationships", func(t *testing.T) {
		router, mockService, authService := setupRequirementTestRouter()

		requirementID := uuid.New()

		mockService.On("DeleteRequirement", requirementID, false).Return(service.ErrRequirementHasRelationships)

		req, err := createAuthenticatedRequirementRequest("DELETE", "/api/v1/requirements/"+requirementID.String(), nil, authService)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "relationships")
		assert.Contains(t, response["hint"], "force=true")

		mockService.AssertExpectations(t)
	})

	t.Run("invalid UUID format", func(t *testing.T) {
		router, _, authService := setupRequirementTestRouter()

		req, err := createAuthenticatedRequirementRequest("DELETE", "/api/v1/requirements/invalid-uuid", nil, authService)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid requirement ID format", response["error"])
	})

	t.Run("unauthorized request", func(t *testing.T) {
		router, _, _ := setupRequirementTestRouter()

		requirementID := uuid.New()

		req, _ := http.NewRequest("DELETE", "/api/v1/requirements/"+requirementID.String(), nil)
		// No Authorization header

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRequirementHandler_CreateRelationship(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		router, mockService, authService := setupRequirementTestRouter()

		sourceID := uuid.New()
		targetID := uuid.New()
		relationshipTypeID := uuid.New()
		creatorID := uuid.New()

		reqBody := service.CreateRelationshipRequest{
			SourceRequirementID: sourceID,
			TargetRequirementID: targetID,
			RelationshipTypeID:  relationshipTypeID,
			CreatedBy:           creatorID,
		}

		expectedRelationship := &models.RequirementRelationship{
			ID:                  uuid.New(),
			SourceRequirementID: sourceID,
			TargetRequirementID: targetID,
			RelationshipTypeID:  relationshipTypeID,
			CreatedBy:           creatorID,
		}

		mockService.On("CreateRelationship", reqBody).Return(expectedRelationship, nil)

		jsonBody, _ := json.Marshal(reqBody)
		req, err := createAuthenticatedRequirementRequest("POST", "/api/v1/requirements/relationships", bytes.NewBuffer(jsonBody), authService)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.RequirementRelationship
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedRelationship.ID, response.ID)
		assert.Equal(t, expectedRelationship.SourceRequirementID, response.SourceRequirementID)

		mockService.AssertExpectations(t)
	})

	t.Run("circular relationship error", func(t *testing.T) {
		router, mockService, authService := setupRequirementTestRouter()

		reqBody := service.CreateRelationshipRequest{
			SourceRequirementID: uuid.New(),
			TargetRequirementID: uuid.New(),
			RelationshipTypeID:  uuid.New(),
			CreatedBy:           uuid.New(),
		}

		mockService.On("CreateRelationship", reqBody).Return(nil, service.ErrCircularRelationship)

		jsonBody, _ := json.Marshal(reqBody)
		req, err := createAuthenticatedRequirementRequest("POST", "/api/v1/requirements/relationships", bytes.NewBuffer(jsonBody), authService)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "same requirement")

		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized request", func(t *testing.T) {
		router, _, _ := setupRequirementTestRouter()

		reqBody := service.CreateRelationshipRequest{
			SourceRequirementID: uuid.New(),
			TargetRequirementID: uuid.New(),
			RelationshipTypeID:  uuid.New(),
			CreatedBy:           uuid.New(),
		}

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/requirements/relationships", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRequirementHandler_SearchRequirements(t *testing.T) {
	t.Run("successful search", func(t *testing.T) {
		router, mockService, authService := setupRequirementTestRouter()

		searchText := "test requirement"
		expectedRequirements := []models.Requirement{
			{
				ID:    uuid.New(),
				Title: "Test Requirement 1",
			},
			{
				ID:    uuid.New(),
				Title: "Test Requirement 2",
			},
		}

		mockService.On("SearchRequirementsWithPagination", searchText, 50, 0).Return(expectedRequirements, int64(2), nil)

		req, err := createAuthenticatedRequirementRequest("GET", "/api/v1/requirements/search?q="+searchText, nil, authService)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["total_count"])
		assert.Equal(t, float64(50), response["limit"])
		assert.Equal(t, float64(0), response["offset"])

		requirements := response["data"].([]interface{})
		assert.Len(t, requirements, 2)

		mockService.AssertExpectations(t)
	})

	t.Run("missing search query", func(t *testing.T) {
		router, _, authService := setupRequirementTestRouter()

		req, err := createAuthenticatedRequirementRequest("GET", "/api/v1/requirements/search", nil, authService)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Search query parameter 'q' is required")
	})

	t.Run("unauthorized request", func(t *testing.T) {
		router, _, _ := setupRequirementTestRouter()

		req, _ := http.NewRequest("GET", "/api/v1/requirements/search?q=test", nil)
		// No Authorization header

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
