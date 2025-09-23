package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// MockUserStoryService is a mock implementation of UserStoryService
type MockUserStoryService struct {
	mock.Mock
}

func (m *MockUserStoryService) CreateUserStory(req service.CreateUserStoryRequest) (*models.UserStory, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryService) GetUserStoryByID(id uuid.UUID) (*models.UserStory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryService) GetUserStoryByReferenceID(referenceID string) (*models.UserStory, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryService) UpdateUserStory(id uuid.UUID, req service.UpdateUserStoryRequest) (*models.UserStory, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryService) DeleteUserStory(id uuid.UUID, force bool) error {
	args := m.Called(id, force)
	return args.Error(0)
}

func (m *MockUserStoryService) ListUserStories(filters service.UserStoryFilters) ([]models.UserStory, int64, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.UserStory), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserStoryService) GetUserStoryWithAcceptanceCriteria(id uuid.UUID) (*models.UserStory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryService) GetUserStoryWithRequirements(id uuid.UUID) (*models.UserStory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryService) GetUserStoriesByEpic(epicID uuid.UUID) ([]models.UserStory, error) {
	args := m.Called(epicID)
	return args.Get(0).([]models.UserStory), args.Error(1)
}

func (m *MockUserStoryService) ChangeUserStoryStatus(id uuid.UUID, newStatus models.UserStoryStatus) (*models.UserStory, error) {
	args := m.Called(id, newStatus)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryService) AssignUserStory(id uuid.UUID, assigneeID uuid.UUID) (*models.UserStory, error) {
	args := m.Called(id, assigneeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func setupUserStoryRouter(handler *UserStoryHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/api/v1")
	{
		v1.POST("/user-stories", handler.CreateUserStory)
		v1.POST("/epics/:id/user-stories", handler.CreateUserStoryInEpic)
		v1.GET("/user-stories/:id", handler.GetUserStory)
		v1.PUT("/user-stories/:id", handler.UpdateUserStory)
		v1.DELETE("/user-stories/:id", handler.DeleteUserStory)
		v1.GET("/user-stories", handler.ListUserStories)
		v1.GET("/user-stories/:id/acceptance-criteria", handler.GetUserStoryWithAcceptanceCriteria)
		v1.GET("/user-stories/:id/requirements", handler.GetUserStoryWithRequirements)
		v1.PATCH("/user-stories/:id/status", handler.ChangeUserStoryStatus)
		v1.PATCH("/user-stories/:id/assign", handler.AssignUserStory)
	}

	return router
}

func TestUserStoryHandler_CreateUserStory(t *testing.T) {
	mockService := new(MockUserStoryService)
	handler := NewUserStoryHandler(mockService)
	router := setupUserStoryRouter(handler)

	t.Run("successful creation", func(t *testing.T) {
		epicID := uuid.New()
		creatorID := uuid.New()
		description := "As a user, I want to login, so that I can access my account"

		reqBody := service.CreateUserStoryRequest{
			EpicID:      epicID,
			CreatorID:   creatorID,
			Priority:    models.PriorityHigh,
			Title:       "User Login",
			Description: &description,
		}

		expectedUserStory := &models.UserStory{
			ID:          uuid.New(),
			EpicID:      epicID,
			CreatorID:   creatorID,
			AssigneeID:  creatorID,
			Priority:    models.PriorityHigh,
			Status:      models.UserStoryStatusBacklog,
			Title:       "User Login",
			Description: &description,
		}

		mockService.On("CreateUserStory", reqBody).Return(expectedUserStory, nil)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/user-stories", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.UserStory
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedUserStory.ID, response.ID)
		assert.Equal(t, expectedUserStory.Title, response.Title)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/user-stories", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("epic not found", func(t *testing.T) {
		reqBody := service.CreateUserStoryRequest{
			EpicID:    uuid.New(),
			CreatorID: uuid.New(),
			Priority:  models.PriorityMedium,
			Title:     "Test User Story",
		}

		mockService.On("CreateUserStory", reqBody).Return(nil, service.ErrEpicNotFound)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/user-stories", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid user story template", func(t *testing.T) {
		reqBody := service.CreateUserStoryRequest{
			EpicID:    uuid.New(),
			CreatorID: uuid.New(),
			Priority:  models.PriorityMedium,
			Title:     "Test User Story",
		}

		mockService.On("CreateUserStory", reqBody).Return(nil, service.ErrInvalidUserStoryTemplate)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/user-stories", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "template")

		mockService.AssertExpectations(t)
	})
}

func TestUserStoryHandler_CreateUserStoryInEpic(t *testing.T) {
	mockService := new(MockUserStoryService)
	handler := NewUserStoryHandler(mockService)
	router := setupUserStoryRouter(handler)

	t.Run("successful creation in epic", func(t *testing.T) {
		epicID := uuid.New()
		creatorID := uuid.New()
		description := "As a user, I want to login, so that I can access my account"

		reqBody := service.CreateUserStoryRequest{
			CreatorID:   creatorID,
			Priority:    models.PriorityHigh,
			Title:       "User Login",
			Description: &description,
		}

		expectedReq := reqBody
		expectedReq.EpicID = epicID

		expectedUserStory := &models.UserStory{
			ID:          uuid.New(),
			EpicID:      epicID,
			CreatorID:   creatorID,
			AssigneeID:  creatorID,
			Priority:    models.PriorityHigh,
			Status:      models.UserStoryStatusBacklog,
			Title:       "User Login",
			Description: &description,
		}

		mockService.On("CreateUserStory", mock.MatchedBy(func(req service.CreateUserStoryRequest) bool {
			return req.EpicID == epicID && req.CreatorID == creatorID && req.Title == "User Login"
		})).Return(expectedUserStory, nil)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/epics/"+epicID.String()+"/user-stories", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Logf("Response body: %s", w.Body.String())
		}
		assert.Equal(t, http.StatusCreated, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid epic ID format", func(t *testing.T) {
		reqBody := service.CreateUserStoryRequest{
			CreatorID: uuid.New(),
			Priority:  models.PriorityHigh,
			Title:     "User Login",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/epics/invalid-uuid/user-stories", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUserStoryHandler_GetUserStory(t *testing.T) {
	mockService := new(MockUserStoryService)
	handler := NewUserStoryHandler(mockService)
	router := setupUserStoryRouter(handler)

	t.Run("successful retrieval by UUID", func(t *testing.T) {
		userStoryID := uuid.New()
		expectedUserStory := &models.UserStory{
			ID:    userStoryID,
			Title: "Test User Story",
		}

		mockService.On("GetUserStoryByID", userStoryID).Return(expectedUserStory, nil)

		req, _ := http.NewRequest("GET", "/api/v1/user-stories/"+userStoryID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.UserStory
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedUserStory.ID, response.ID)

		mockService.AssertExpectations(t)
	})

	t.Run("successful retrieval by reference ID", func(t *testing.T) {
		referenceID := "US-001"
		expectedUserStory := &models.UserStory{
			ReferenceID: referenceID,
			Title:       "Test User Story",
		}

		mockService.On("GetUserStoryByReferenceID", referenceID).Return(expectedUserStory, nil)

		req, _ := http.NewRequest("GET", "/api/v1/user-stories/"+referenceID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("user story not found", func(t *testing.T) {
		userStoryID := uuid.New()

		mockService.On("GetUserStoryByID", userStoryID).Return(nil, service.ErrUserStoryNotFound)

		req, _ := http.NewRequest("GET", "/api/v1/user-stories/"+userStoryID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestUserStoryHandler_UpdateUserStory(t *testing.T) {
	mockService := new(MockUserStoryService)
	handler := NewUserStoryHandler(mockService)
	router := setupUserStoryRouter(handler)

	t.Run("successful update", func(t *testing.T) {
		userStoryID := uuid.New()
		assigneeID := uuid.New()
		newTitle := "Updated User Story"
		newDescription := "As an admin, I want to manage users, so that I can control access"

		reqBody := service.UpdateUserStoryRequest{
			AssigneeID:  &assigneeID,
			Title:       &newTitle,
			Description: &newDescription,
		}

		expectedUserStory := &models.UserStory{
			ID:          userStoryID,
			AssigneeID:  assigneeID,
			Title:       newTitle,
			Description: &newDescription,
		}

		mockService.On("UpdateUserStory", userStoryID, reqBody).Return(expectedUserStory, nil)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/api/v1/user-stories/"+userStoryID.String(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid UUID format", func(t *testing.T) {
		reqBody := service.UpdateUserStoryRequest{}

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/api/v1/user-stories/invalid-uuid", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("user story not found", func(t *testing.T) {
		userStoryID := uuid.New()
		reqBody := service.UpdateUserStoryRequest{}

		mockService.On("UpdateUserStory", userStoryID, reqBody).Return(nil, service.ErrUserStoryNotFound)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/api/v1/user-stories/"+userStoryID.String(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestUserStoryHandler_DeleteUserStory(t *testing.T) {
	mockService := new(MockUserStoryService)
	handler := NewUserStoryHandler(mockService)
	router := setupUserStoryRouter(handler)

	t.Run("successful deletion", func(t *testing.T) {
		userStoryID := uuid.New()

		mockService.On("DeleteUserStory", userStoryID, false).Return(nil)

		req, _ := http.NewRequest("DELETE", "/api/v1/user-stories/"+userStoryID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("force deletion", func(t *testing.T) {
		userStoryID := uuid.New()

		mockService.On("DeleteUserStory", userStoryID, true).Return(nil)

		req, _ := http.NewRequest("DELETE", "/api/v1/user-stories/"+userStoryID.String()+"?force=true", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("deletion blocked by requirements", func(t *testing.T) {
		userStoryID := uuid.New()

		mockService.On("DeleteUserStory", userStoryID, false).Return(service.ErrUserStoryHasRequirements)

		req, _ := http.NewRequest("DELETE", "/api/v1/user-stories/"+userStoryID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["hint"], "force=true")

		mockService.AssertExpectations(t)
	})

	t.Run("invalid UUID format", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/v1/user-stories/invalid-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUserStoryHandler_ListUserStories(t *testing.T) {
	mockService := new(MockUserStoryService)
	handler := NewUserStoryHandler(mockService)
	router := setupUserStoryRouter(handler)

	t.Run("successful listing with filters", func(t *testing.T) {
		epicID := uuid.New()
		expectedUserStories := []models.UserStory{
			{ID: uuid.New(), Title: "User Story 1"},
			{ID: uuid.New(), Title: "User Story 2"},
		}

		mockService.On("ListUserStories", mock.MatchedBy(func(filters service.UserStoryFilters) bool {
			return filters.EpicID != nil && *filters.EpicID == epicID
		})).Return(expectedUserStories, int64(2), nil)

		req, _ := http.NewRequest("GET", "/api/v1/user-stories?epic_id="+epicID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["total_count"])
		assert.Equal(t, float64(50), response["limit"])
		assert.Equal(t, float64(0), response["offset"])
		assert.NotNil(t, response["data"])

		mockService.AssertExpectations(t)
	})

	t.Run("successful listing without filters", func(t *testing.T) {
		expectedUserStories := []models.UserStory{
			{ID: uuid.New(), Title: "User Story 1"},
		}

		mockService.On("ListUserStories", mock.AnythingOfType("service.UserStoryFilters")).Return(expectedUserStories, int64(1), nil)

		req, _ := http.NewRequest("GET", "/api/v1/user-stories", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), response["total_count"])
		assert.Equal(t, float64(50), response["limit"])
		assert.Equal(t, float64(0), response["offset"])
		assert.NotNil(t, response["data"])

		mockService.AssertExpectations(t)
	})
}

func TestUserStoryHandler_ChangeUserStoryStatus(t *testing.T) {
	mockService := new(MockUserStoryService)
	handler := NewUserStoryHandler(mockService)
	router := setupUserStoryRouter(handler)

	t.Run("successful status change", func(t *testing.T) {
		userStoryID := uuid.New()
		newStatus := models.UserStoryStatusInProgress

		reqBody := map[string]interface{}{
			"status": newStatus,
		}

		expectedUserStory := &models.UserStory{
			ID:     userStoryID,
			Status: newStatus,
		}

		mockService.On("ChangeUserStoryStatus", userStoryID, newStatus).Return(expectedUserStory, nil)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PATCH", "/api/v1/user-stories/"+userStoryID.String()+"/status", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid status transition", func(t *testing.T) {
		userStoryID := uuid.New()
		newStatus := models.UserStoryStatusDone

		reqBody := map[string]interface{}{
			"status": newStatus,
		}

		mockService.On("ChangeUserStoryStatus", userStoryID, newStatus).Return(nil, service.ErrInvalidStatusTransition)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PATCH", "/api/v1/user-stories/"+userStoryID.String()+"/status", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockService.AssertExpectations(t)
	})
}
