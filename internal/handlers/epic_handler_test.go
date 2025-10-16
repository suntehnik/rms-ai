package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// MockEpicService is a mock implementation of EpicService
type MockEpicService struct {
	mock.Mock
}

func (m *MockEpicService) CreateEpic(req service.CreateEpicRequest) (*models.Epic, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MockEpicService) GetEpicByID(id uuid.UUID) (*models.Epic, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MockEpicService) GetEpicByReferenceID(referenceID string) (*models.Epic, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MockEpicService) UpdateEpic(id uuid.UUID, req service.UpdateEpicRequest) (*models.Epic, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MockEpicService) DeleteEpic(id uuid.UUID, force bool) error {
	args := m.Called(id, force)
	return args.Error(0)
}

func (m *MockEpicService) ListEpics(filters service.EpicFilters) ([]models.Epic, int64, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.Epic), args.Get(1).(int64), args.Error(2)
}

func (m *MockEpicService) GetEpicWithUserStories(id uuid.UUID) (*models.Epic, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MockEpicService) ChangeEpicStatus(id uuid.UUID, newStatus models.EpicStatus) (*models.Epic, error) {
	args := m.Called(id, newStatus)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MockEpicService) AssignEpic(id uuid.UUID, assigneeID *uuid.UUID) (*models.Epic, error) {
	args := m.Called(id, assigneeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func setupEpicTestRouter() (*gin.Engine, *auth.Service) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create auth service for testing
	authService := auth.NewService("test-secret", time.Hour)

	return router, authService
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Helper function to create authenticated request
func createAuthenticatedEpicRequest(method, url string, body *bytes.Buffer, authService *auth.Service) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = body
	}
	req, err := http.NewRequest(method, url, bodyReader)
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
func TestEpicHandler_CreateEpic(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockEpicService)
		expectedStatus int
	}{
		{
			name: "successful epic creation",
			requestBody: service.CreateEpicRequest{
				CreatorID:   uuid.New(),
				Priority:    models.PriorityHigh,
				Title:       "Test Epic",
				Description: stringPtr("Test Description"),
			},
			setupMock: func(mockService *MockEpicService) {
				epic := &models.Epic{
					ID:          uuid.New(),
					ReferenceID: "EP-001",
					Title:       "Test Epic",
					Priority:    models.PriorityHigh,
					Status:      models.EpicStatusBacklog,
				}
				mockService.On("CreateEpic", mock.AnythingOfType("service.CreateEpicRequest")).Return(epic, nil)
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockEpicService)
			tt.setupMock(mockService)

			handler := NewEpicHandler(mockService)
			router, authService := setupEpicTestRouter()
			router.Use(authService.Middleware())
			router.POST("/epics", handler.CreateEpic)

			body, _ := json.Marshal(tt.requestBody)
			req, err := createAuthenticatedEpicRequest("POST", "/epics", bytes.NewBuffer(body), authService)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}
func TestEpicHandler_GetEpic(t *testing.T) {
	tests := []struct {
		name           string
		epicID         string
		setupMock      func(*MockEpicService)
		expectedStatus int
	}{
		{
			name:   "successful get epic by UUID",
			epicID: uuid.New().String(),
			setupMock: func(mockService *MockEpicService) {
				epic := &models.Epic{
					ID:          uuid.New(),
					ReferenceID: "EP-001",
					Title:       "Test Epic",
				}
				mockService.On("GetEpicByID", mock.AnythingOfType("uuid.UUID")).Return(epic, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "successful get epic by reference ID",
			epicID: "EP-001",
			setupMock: func(mockService *MockEpicService) {
				epic := &models.Epic{
					ID:          uuid.New(),
					ReferenceID: "EP-001",
					Title:       "Test Epic",
				}
				mockService.On("GetEpicByReferenceID", "EP-001").Return(epic, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "epic not found",
			epicID: uuid.New().String(),
			setupMock: func(mockService *MockEpicService) {
				mockService.On("GetEpicByID", mock.AnythingOfType("uuid.UUID")).Return(nil, service.ErrEpicNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockEpicService)
			tt.setupMock(mockService)

			handler := NewEpicHandler(mockService)
			router, authService := setupEpicTestRouter()
			router.Use(authService.Middleware())
			router.GET("/epics/:id", handler.GetEpic)

			req, err := createAuthenticatedEpicRequest("GET", fmt.Sprintf("/epics/%s", tt.epicID), nil, authService)
			assert.NoError(t, err)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestEpicHandler_ListEpics(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*MockEpicService)
		expectedStatus int
	}{
		{
			name:        "successful list epics",
			queryParams: "",
			setupMock: func(mockService *MockEpicService) {
				epics := []models.Epic{
					{
						ID:          uuid.New(),
						ReferenceID: "EP-001",
						Title:       "Epic 1",
					},
					{
						ID:          uuid.New(),
						ReferenceID: "EP-002",
						Title:       "Epic 2",
					},
				}
				mockService.On("ListEpics", mock.AnythingOfType("service.EpicFilters")).Return(epics, int64(2), nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockEpicService)
			tt.setupMock(mockService)

			handler := NewEpicHandler(mockService)
			router, authService := setupEpicTestRouter()
			router.Use(authService.Middleware())
			router.GET("/epics", handler.ListEpics)

			url := "/epics" + tt.queryParams
			req, err := createAuthenticatedEpicRequest("GET", url, nil, authService)
			assert.NoError(t, err)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

// TestEpicHandler_ListEpics_JSONResponse tests that the ListEpics handler returns
// creator_id and assignee_id fields without including the full creator and assignee objects
// in the JSON response. This ensures the API returns clean, minimal responses with only
// the necessary ID fields.
func TestEpicHandler_ListEpics_JSONResponse(t *testing.T) {
	t.Run("should return creator_id and assignee_id without creator and assignee objects", func(t *testing.T) {
		// Setup
		mockService := new(MockEpicService)

		creatorID := uuid.New()
		assigneeID := uuid.New()

		epics := []models.Epic{
			{
				ID:          uuid.New(),
				ReferenceID: "EP-001",
				CreatorID:   creatorID,
				AssigneeID:  assigneeID,
				Title:       "Test Epic",
				Description: stringPtr("Test Description"),
				Priority:    models.PriorityHigh,
				Status:      models.EpicStatusBacklog,
			},
		}

		mockService.On("ListEpics", mock.AnythingOfType("service.EpicFilters")).Return(epics, int64(1), nil)

		handler := NewEpicHandler(mockService)
		router, authService := setupEpicTestRouter()
		router.Use(authService.Middleware())
		router.GET("/epics", handler.ListEpics)

		// Execute
		req, err := createAuthenticatedEpicRequest("GET", "/epics", nil, authService)
		assert.NoError(t, err)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert HTTP response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse JSON response
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Verify response structure
		assert.Contains(t, response, "data")
		assert.Contains(t, response, "total_count")
		assert.Contains(t, response, "limit")
		assert.Contains(t, response, "offset")

		// Get the epics data
		data, ok := response["data"].([]interface{})
		assert.True(t, ok, "data should be an array")
		assert.Len(t, data, 1, "should have one epic")

		// Get the first epic
		epic, ok := data[0].(map[string]interface{})
		assert.True(t, ok, "epic should be an object")

		// Verify that creator_id and assignee_id are present
		assert.Contains(t, epic, "creator_id", "should contain creator_id field")
		assert.Contains(t, epic, "assignee_id", "should contain assignee_id field")

		// Verify that creator_id and assignee_id have the correct values
		assert.Equal(t, creatorID.String(), epic["creator_id"], "creator_id should match")
		assert.Equal(t, assigneeID.String(), epic["assignee_id"], "assignee_id should match")

		// Verify that creator and assignee objects are NOT present
		assert.NotContains(t, epic, "creator", "should NOT contain creator object")
		assert.NotContains(t, epic, "assignee", "should NOT contain assignee object")

		// Verify other expected fields are present
		assert.Contains(t, epic, "id")
		assert.Contains(t, epic, "reference_id")
		assert.Contains(t, epic, "title")
		assert.Contains(t, epic, "description")
		assert.Contains(t, epic, "priority")
		assert.Contains(t, epic, "status")

		// Verify the values
		assert.Equal(t, "EP-001", epic["reference_id"])
		assert.Equal(t, "Test Epic", epic["title"])
		assert.Equal(t, "Test Description", epic["description"])
		assert.Equal(t, float64(2), epic["priority"]) // JSON unmarshals numbers as float64
		assert.Equal(t, "Backlog", epic["status"])

		mockService.AssertExpectations(t)
	})

	t.Run("should handle multiple epics correctly", func(t *testing.T) {
		// Setup
		mockService := new(MockEpicService)

		creator1ID := uuid.New()
		assignee1ID := uuid.New()
		creator2ID := uuid.New()
		assignee2ID := uuid.New()

		epics := []models.Epic{
			{
				ID:          uuid.New(),
				ReferenceID: "EP-001",
				CreatorID:   creator1ID,
				AssigneeID:  assignee1ID,
				Title:       "Epic 1",
				Priority:    models.PriorityHigh,
				Status:      models.EpicStatusBacklog,
			},
			{
				ID:          uuid.New(),
				ReferenceID: "EP-002",
				CreatorID:   creator2ID,
				AssigneeID:  assignee2ID,
				Title:       "Epic 2",
				Priority:    models.PriorityMedium,
				Status:      models.EpicStatusInProgress,
			},
		}

		mockService.On("ListEpics", mock.AnythingOfType("service.EpicFilters")).Return(epics, int64(2), nil)

		handler := NewEpicHandler(mockService)
		router, authService := setupEpicTestRouter()
		router.Use(authService.Middleware())
		router.GET("/epics", handler.ListEpics)

		// Execute
		req, err := createAuthenticatedEpicRequest("GET", "/epics", nil, authService)
		assert.NoError(t, err)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert HTTP response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse JSON response
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Get the epics data
		data, ok := response["data"].([]interface{})
		assert.True(t, ok, "data should be an array")
		assert.Len(t, data, 2, "should have two epics")

		// Check both epics
		for i, expectedEpic := range epics {
			epic, ok := data[i].(map[string]interface{})
			assert.True(t, ok, "epic should be an object")

			// Verify creator_id and assignee_id are present with correct values
			assert.Equal(t, expectedEpic.CreatorID.String(), epic["creator_id"], "creator_id should match for epic %d", i)
			assert.Equal(t, expectedEpic.AssigneeID.String(), epic["assignee_id"], "assignee_id should match for epic %d", i)

			// Verify creator and assignee objects are NOT present
			assert.NotContains(t, epic, "creator", "epic %d should NOT contain creator object", i)
			assert.NotContains(t, epic, "assignee", "epic %d should NOT contain assignee object", i)
		}

		mockService.AssertExpectations(t)
	})
}

func TestEpicHandler_DeleteEpic(t *testing.T) {
	tests := []struct {
		name           string
		epicID         string
		force          string
		setupMock      func(*MockEpicService)
		expectedStatus int
	}{
		{
			name:   "successful delete",
			epicID: uuid.New().String(),
			force:  "",
			setupMock: func(mockService *MockEpicService) {
				mockService.On("DeleteEpic", mock.AnythingOfType("uuid.UUID"), false).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "epic has user stories",
			epicID: uuid.New().String(),
			force:  "",
			setupMock: func(mockService *MockEpicService) {
				mockService.On("DeleteEpic", mock.AnythingOfType("uuid.UUID"), false).Return(service.ErrEpicHasUserStories)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockEpicService)
			tt.setupMock(mockService)

			handler := NewEpicHandler(mockService)
			router, authService := setupEpicTestRouter()
			router.Use(authService.Middleware())
			router.DELETE("/epics/:id", handler.DeleteEpic)

			url := fmt.Sprintf("/epics/%s", tt.epicID)
			if tt.force != "" {
				url += fmt.Sprintf("?force=%s", tt.force)
			}

			req, err := createAuthenticatedEpicRequest("DELETE", url, nil, authService)
			assert.NoError(t, err)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestEpicHandler_ChangeEpicStatus(t *testing.T) {
	tests := []struct {
		name           string
		epicID         string
		requestBody    interface{}
		setupMock      func(*MockEpicService)
		expectedStatus int
	}{
		{
			name:   "successful status change",
			epicID: uuid.New().String(),
			requestBody: map[string]string{
				"status": "In Progress",
			},
			setupMock: func(mockService *MockEpicService) {
				epic := &models.Epic{
					ID:     uuid.New(),
					Status: models.EpicStatusInProgress,
				}
				mockService.On("ChangeEpicStatus", mock.AnythingOfType("uuid.UUID"), models.EpicStatusInProgress).Return(epic, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockEpicService)
			tt.setupMock(mockService)

			handler := NewEpicHandler(mockService)
			router, authService := setupEpicTestRouter()
			router.Use(authService.Middleware())
			router.PATCH("/epics/:id/status", handler.ChangeEpicStatus)

			body, _ := json.Marshal(tt.requestBody)
			req, err := createAuthenticatedEpicRequest("PATCH", fmt.Sprintf("/epics/%s/status", tt.epicID), bytes.NewBuffer(body), authService)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}
