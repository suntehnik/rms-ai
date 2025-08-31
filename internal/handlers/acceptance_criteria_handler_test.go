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

// MockAcceptanceCriteriaService is a mock implementation of AcceptanceCriteriaService
type MockAcceptanceCriteriaService struct {
	mock.Mock
}

func (m *MockAcceptanceCriteriaService) CreateAcceptanceCriteria(req service.CreateAcceptanceCriteriaRequest) (*models.AcceptanceCriteria, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AcceptanceCriteria), args.Error(1)
}

func (m *MockAcceptanceCriteriaService) GetAcceptanceCriteriaByID(id uuid.UUID) (*models.AcceptanceCriteria, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AcceptanceCriteria), args.Error(1)
}

func (m *MockAcceptanceCriteriaService) GetAcceptanceCriteriaByReferenceID(referenceID string) (*models.AcceptanceCriteria, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AcceptanceCriteria), args.Error(1)
}

func (m *MockAcceptanceCriteriaService) UpdateAcceptanceCriteria(id uuid.UUID, req service.UpdateAcceptanceCriteriaRequest) (*models.AcceptanceCriteria, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AcceptanceCriteria), args.Error(1)
}

func (m *MockAcceptanceCriteriaService) DeleteAcceptanceCriteria(id uuid.UUID, force bool) error {
	args := m.Called(id, force)
	return args.Error(0)
}

func (m *MockAcceptanceCriteriaService) ListAcceptanceCriteria(filters service.AcceptanceCriteriaFilters) ([]models.AcceptanceCriteria, error) {
	args := m.Called(filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AcceptanceCriteria), args.Error(1)
}

func (m *MockAcceptanceCriteriaService) GetAcceptanceCriteriaByUserStory(userStoryID uuid.UUID) ([]models.AcceptanceCriteria, error) {
	args := m.Called(userStoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AcceptanceCriteria), args.Error(1)
}

func (m *MockAcceptanceCriteriaService) GetAcceptanceCriteriaByAuthor(authorID uuid.UUID) ([]models.AcceptanceCriteria, error) {
	args := m.Called(authorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AcceptanceCriteria), args.Error(1)
}

func (m *MockAcceptanceCriteriaService) ValidateUserStoryHasAcceptanceCriteria(userStoryID uuid.UUID) error {
	args := m.Called(userStoryID)
	return args.Error(0)
}

func setupAcceptanceCriteriaTestRouter() (*gin.Engine, *MockAcceptanceCriteriaService) {
	gin.SetMode(gin.TestMode)
	
	mockService := new(MockAcceptanceCriteriaService)
	handler := NewAcceptanceCriteriaHandler(mockService)
	
	router := gin.New()
	
	v1 := router.Group("/api/v1")
	{
		v1.POST("/user-stories/:id/acceptance-criteria", handler.CreateAcceptanceCriteria)
		v1.GET("/acceptance-criteria/:id", handler.GetAcceptanceCriteria)
		v1.PUT("/acceptance-criteria/:id", handler.UpdateAcceptanceCriteria)
		v1.DELETE("/acceptance-criteria/:id", handler.DeleteAcceptanceCriteria)
		v1.GET("/acceptance-criteria", handler.ListAcceptanceCriteria)
		v1.GET("/user-stories/:id/acceptance-criteria", handler.GetAcceptanceCriteriaByUserStory)
	}
	
	return router, mockService
}

func TestAcceptanceCriteriaHandler_CreateAcceptanceCriteria(t *testing.T) {
	router, mockService := setupAcceptanceCriteriaTestRouter()

	userStoryID := uuid.New()
	authorID := uuid.New()
	acceptanceCriteriaID := uuid.New()

	tests := []struct {
		name           string
		userStoryID    string
		requestBody    interface{}
		setupMocks     func()
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "successful creation",
			userStoryID: userStoryID.String(),
			requestBody: service.CreateAcceptanceCriteriaRequest{
				AuthorID:    authorID,
				Description: "WHEN user clicks submit THEN system SHALL validate the form",
			},
			setupMocks: func() {
				expectedReq := service.CreateAcceptanceCriteriaRequest{
					UserStoryID: userStoryID,
					AuthorID:    authorID,
					Description: "WHEN user clicks submit THEN system SHALL validate the form",
				}
				expectedResult := &models.AcceptanceCriteria{
					ID:          acceptanceCriteriaID,
					UserStoryID: userStoryID,
					AuthorID:    authorID,
					ReferenceID: "AC-001",
					Description: "WHEN user clicks submit THEN system SHALL validate the form",
				}
				mockService.On("CreateAcceptanceCriteria", expectedReq).Return(expectedResult, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response models.AcceptanceCriteria
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, acceptanceCriteriaID, response.ID)
				assert.Equal(t, "AC-001", response.ReferenceID)
			},
		},
		{
			name:        "invalid user story ID",
			userStoryID: "invalid-uuid",
			requestBody: service.CreateAcceptanceCriteriaRequest{
				AuthorID:    authorID,
				Description: "WHEN user clicks submit THEN system SHALL validate the form",
			},
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid user story ID format")
			},
		},
		{
			name:        "user story not found",
			userStoryID: userStoryID.String(),
			requestBody: service.CreateAcceptanceCriteriaRequest{
				AuthorID:    authorID,
				Description: "WHEN user clicks submit THEN system SHALL validate the form",
			},
			setupMocks: func() {
				expectedReq := service.CreateAcceptanceCriteriaRequest{
					UserStoryID: userStoryID,
					AuthorID:    authorID,
					Description: "WHEN user clicks submit THEN system SHALL validate the form",
				}
				mockService.On("CreateAcceptanceCriteria", expectedReq).Return(nil, service.ErrUserStoryNotFound)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "User story not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			tt.setupMocks()

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/user-stories/"+tt.userStoryID+"/acceptance-criteria", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			mockService.AssertExpectations(t)
		})
	}
}

func TestAcceptanceCriteriaHandler_GetAcceptanceCriteria(t *testing.T) {
	router, mockService := setupAcceptanceCriteriaTestRouter()

	acceptanceCriteriaID := uuid.New()
	expectedAcceptanceCriteria := &models.AcceptanceCriteria{
		ID:          acceptanceCriteriaID,
		ReferenceID: "AC-001",
		Description: "WHEN user clicks submit THEN system SHALL validate the form",
	}

	tests := []struct {
		name           string
		id             string
		setupMocks     func()
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful retrieval by UUID",
			id:   acceptanceCriteriaID.String(),
			setupMocks: func() {
				mockService.On("GetAcceptanceCriteriaByID", acceptanceCriteriaID).Return(expectedAcceptanceCriteria, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response models.AcceptanceCriteria
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, acceptanceCriteriaID, response.ID)
			},
		},
		{
			name: "successful retrieval by reference ID",
			id:   "AC-001",
			setupMocks: func() {
				mockService.On("GetAcceptanceCriteriaByReferenceID", "AC-001").Return(expectedAcceptanceCriteria, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response models.AcceptanceCriteria
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "AC-001", response.ReferenceID)
			},
		},
		{
			name: "acceptance criteria not found",
			id:   acceptanceCriteriaID.String(),
			setupMocks: func() {
				mockService.On("GetAcceptanceCriteriaByID", acceptanceCriteriaID).Return(nil, service.ErrAcceptanceCriteriaNotFound)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Acceptance criteria not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			tt.setupMocks()

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/acceptance-criteria/"+tt.id, nil)
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			mockService.AssertExpectations(t)
		})
	}
}

func TestAcceptanceCriteriaHandler_DeleteAcceptanceCriteria(t *testing.T) {
	router, mockService := setupAcceptanceCriteriaTestRouter()

	acceptanceCriteriaID := uuid.New()

	tests := []struct {
		name           string
		id             string
		force          string
		setupMocks     func()
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:  "successful deletion",
			id:    acceptanceCriteriaID.String(),
			force: "",
			setupMocks: func() {
				mockService.On("DeleteAcceptanceCriteria", acceptanceCriteriaID, false).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Empty(t, w.Body.String())
			},
		},
		{
			name:  "force deletion",
			id:    acceptanceCriteriaID.String(),
			force: "true",
			setupMocks: func() {
				mockService.On("DeleteAcceptanceCriteria", acceptanceCriteriaID, true).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Empty(t, w.Body.String())
			},
		},
		{
			name:  "acceptance criteria not found",
			id:    acceptanceCriteriaID.String(),
			force: "",
			setupMocks: func() {
				mockService.On("DeleteAcceptanceCriteria", acceptanceCriteriaID, false).Return(service.ErrAcceptanceCriteriaNotFound)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Acceptance criteria not found")
			},
		},
		{
			name:  "has requirements without force",
			id:    acceptanceCriteriaID.String(),
			force: "",
			setupMocks: func() {
				mockService.On("DeleteAcceptanceCriteria", acceptanceCriteriaID, false).Return(service.ErrAcceptanceCriteriaHasRequirements)
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "has associated requirements")
				assert.Contains(t, response["hint"], "force=true")
			},
		},
		{
			name:  "last acceptance criteria",
			id:    acceptanceCriteriaID.String(),
			force: "",
			setupMocks: func() {
				mockService.On("DeleteAcceptanceCriteria", acceptanceCriteriaID, false).Return(service.ErrUserStoryMustHaveAcceptanceCriteria)
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "must have at least one acceptance criteria")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			tt.setupMocks()

			// Create request
			url := "/api/v1/acceptance-criteria/" + tt.id
			if tt.force != "" {
				url += "?force=" + tt.force
			}
			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			mockService.AssertExpectations(t)
		})
	}
}

func TestAcceptanceCriteriaHandler_ListAcceptanceCriteria(t *testing.T) {
	router, mockService := setupAcceptanceCriteriaTestRouter()

	userStoryID := uuid.New()
	authorID := uuid.New()
	
	expectedAcceptanceCriteria := []models.AcceptanceCriteria{
		{
			ID:          uuid.New(),
			ReferenceID: "AC-001",
			UserStoryID: userStoryID,
			AuthorID:    authorID,
			Description: "WHEN user clicks submit THEN system SHALL validate the form",
		},
		{
			ID:          uuid.New(),
			ReferenceID: "AC-002",
			UserStoryID: userStoryID,
			AuthorID:    authorID,
			Description: "WHEN validation fails THEN system SHALL display error message",
		},
	}

	tests := []struct {
		name           string
		queryParams    string
		setupMocks     func()
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "successful list without filters",
			queryParams: "",
			setupMocks: func() {
				expectedFilters := service.AcceptanceCriteriaFilters{}
				mockService.On("ListAcceptanceCriteria", expectedFilters).Return(expectedAcceptanceCriteria, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, float64(2), response["count"])
				
				criteria := response["acceptance_criteria"].([]interface{})
				assert.Len(t, criteria, 2)
			},
		},
		{
			name:        "successful list with user story filter",
			queryParams: "?user_story_id=" + userStoryID.String(),
			setupMocks: func() {
				expectedFilters := service.AcceptanceCriteriaFilters{
					UserStoryID: &userStoryID,
				}
				mockService.On("ListAcceptanceCriteria", expectedFilters).Return(expectedAcceptanceCriteria, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, float64(2), response["count"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			tt.setupMocks()

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/acceptance-criteria"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			mockService.AssertExpectations(t)
		})
	}
}