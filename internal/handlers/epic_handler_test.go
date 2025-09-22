package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
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
			router := setupTestRouter()
			router.POST("/epics", handler.CreateEpic)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/epics", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

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
			router := setupTestRouter()
			router.GET("/epics/:id", handler.GetEpic)

			req, _ := http.NewRequest("GET", fmt.Sprintf("/epics/%s", tt.epicID), nil)
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
			router := setupTestRouter()
			router.GET("/epics", handler.ListEpics)

			url := "/epics" + tt.queryParams
			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
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
			router := setupTestRouter()
			router.DELETE("/epics/:id", handler.DeleteEpic)

			url := fmt.Sprintf("/epics/%s", tt.epicID)
			if tt.force != "" {
				url += fmt.Sprintf("?force=%s", tt.force)
			}

			req, _ := http.NewRequest("DELETE", url, nil)
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
			router := setupTestRouter()
			router.PATCH("/epics/:id/status", handler.ChangeEpicStatus)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("PATCH", fmt.Sprintf("/epics/%s/status", tt.epicID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}
