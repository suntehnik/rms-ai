package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

// MockPATService is a mock implementation of PATService
type MockPATService struct {
	mock.Mock
}

func (m *MockPATService) CreatePAT(ctx context.Context, userID uuid.UUID, req service.CreatePATRequest) (*service.PATCreateResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.PATCreateResponse), args.Error(1)
}

func (m *MockPATService) ListUserPATs(ctx context.Context, userID uuid.UUID, limit, offset int) (*service.ListResponse[models.PersonalAccessToken], error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ListResponse[models.PersonalAccessToken]), args.Error(1)
}

func (m *MockPATService) GetPAT(ctx context.Context, patID, userID uuid.UUID) (*models.PersonalAccessToken, error) {
	args := m.Called(ctx, patID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PersonalAccessToken), args.Error(1)
}

func (m *MockPATService) RevokePAT(ctx context.Context, patID, userID uuid.UUID) error {
	args := m.Called(ctx, patID, userID)
	return args.Error(0)
}

func (m *MockPATService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockPATService) UpdateLastUsed(ctx context.Context, patID uuid.UUID) error {
	args := m.Called(ctx, patID)
	return args.Error(0)
}

func (m *MockPATService) CleanupExpiredTokens(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func setupPATTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func setupAuthenticatedRouter(userID uuid.UUID) *gin.Engine {
	router := setupPATTestRouter()

	// Add middleware to set claims for authenticated routes
	router.Use(func(c *gin.Context) {
		// Create claims directly (simulating what the auth middleware would do)
		claims := &auth.Claims{
			UserID:   userID.String(),
			Username: "testuser",
			Role:     models.RoleUser,
		}
		c.Set("claims", claims)
		c.Next()
	})

	return router
}

func TestPATHandler_CreatePAT(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockPATService)
		setupAuth      bool
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful PAT creation",
			requestBody: service.CreatePATRequest{
				Name:      "Test Token",
				ExpiresAt: nil,
				Scopes:    []string{"full_access"},
			},
			setupMock: func(mockService *MockPATService) {
				pat := &models.PersonalAccessToken{
					ID:        uuid.New(),
					UserID:    userID,
					Name:      "Test Token",
					Prefix:    "mcp_pat_",
					Scopes:    `["full_access"]`,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				response := &service.PATCreateResponse{
					Token: "mcp_pat_test_token_secret",
					PAT:   *pat,
				}
				mockService.On("CreatePAT", mock.Anything, userID, mock.AnythingOfType("service.CreatePATRequest")).Return(response, nil)
			},
			setupAuth:      true,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"name": "", // empty name should fail validation
			},
			setupMock:      func(mockService *MockPATService) {},
			setupAuth:      true,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "missing authentication",
			requestBody:    service.CreatePATRequest{Name: "Test Token"},
			setupMock:      func(mockService *MockPATService) {},
			setupAuth:      false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "AUTHENTICATION_REQUIRED",
		},
		{
			name: "duplicate token name",
			requestBody: service.CreatePATRequest{
				Name:   "Duplicate Token",
				Scopes: []string{"full_access"},
			},
			setupMock: func(mockService *MockPATService) {
				mockService.On("CreatePAT", mock.Anything, userID, mock.AnythingOfType("service.CreatePATRequest")).Return(nil, service.ErrPATDuplicateName)
			},
			setupAuth:      true,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "invalid scopes",
			requestBody: service.CreatePATRequest{
				Name:   "Invalid Scopes Token",
				Scopes: []string{"invalid_scope"},
			},
			setupMock: func(mockService *MockPATService) {
				mockService.On("CreatePAT", mock.Anything, userID, mock.AnythingOfType("service.CreatePATRequest")).Return(nil, service.ErrPATInvalidScopes)
			},
			setupAuth:      true,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "user not found",
			requestBody: service.CreatePATRequest{
				Name:   "Test Token",
				Scopes: []string{"full_access"},
			},
			setupMock: func(mockService *MockPATService) {
				mockService.On("CreatePAT", mock.Anything, userID, mock.AnythingOfType("service.CreatePATRequest")).Return(nil, service.ErrPATUserNotFound)
			},
			setupAuth:      true,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "ENTITY_NOT_FOUND",
		},
		{
			name: "internal server error",
			requestBody: service.CreatePATRequest{
				Name:   "Test Token",
				Scopes: []string{"full_access"},
			},
			setupMock: func(mockService *MockPATService) {
				mockService.On("CreatePAT", mock.Anything, userID, mock.AnythingOfType("service.CreatePATRequest")).Return(nil, errors.New("database error"))
			},
			setupAuth:      true,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPATService)
			tt.setupMock(mockService)

			handler := NewPATHandler(mockService)

			var router *gin.Engine
			if tt.setupAuth {
				router = setupAuthenticatedRouter(userID)
			} else {
				router = setupPATTestRouter()
			}

			router.POST("/api/v1/pats", handler.CreatePAT)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/pats", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				errorObj, exists := response["error"].(map[string]interface{})
				assert.True(t, exists)
				assert.Equal(t, tt.expectedError, errorObj["code"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestPATHandler_ListPATs(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*MockPATService)
		setupAuth      bool
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "successful list PATs",
			queryParams: "",
			setupMock: func(mockService *MockPATService) {
				pats := []models.PersonalAccessToken{
					{
						ID:        uuid.New(),
						UserID:    userID,
						Name:      "Token 1",
						Prefix:    "mcp_pat_",
						CreatedAt: time.Now(),
					},
					{
						ID:        uuid.New(),
						UserID:    userID,
						Name:      "Token 2",
						Prefix:    "mcp_pat_",
						CreatedAt: time.Now(),
					},
				}
				response := &service.ListResponse[models.PersonalAccessToken]{
					Data:       pats,
					TotalCount: 2,
					Limit:      50,
					Offset:     0,
				}
				mockService.On("ListUserPATs", mock.Anything, userID, 50, 0).Return(response, nil)
			},
			setupAuth:      true,
			expectedStatus: http.StatusOK,
		},
		{
			name:        "list PATs with pagination",
			queryParams: "?limit=10&offset=20",
			setupMock: func(mockService *MockPATService) {
				response := &service.ListResponse[models.PersonalAccessToken]{
					Data:       []models.PersonalAccessToken{},
					TotalCount: 0,
					Limit:      10,
					Offset:     20,
				}
				mockService.On("ListUserPATs", mock.Anything, userID, 10, 20).Return(response, nil)
			},
			setupAuth:      true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing authentication",
			queryParams:    "",
			setupMock:      func(mockService *MockPATService) {},
			setupAuth:      false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "AUTHENTICATION_REQUIRED",
		},
		{
			name:        "user not found",
			queryParams: "",
			setupMock: func(mockService *MockPATService) {
				mockService.On("ListUserPATs", mock.Anything, userID, 50, 0).Return(nil, service.ErrPATUserNotFound)
			},
			setupAuth:      true,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "ENTITY_NOT_FOUND",
		},
		{
			name:        "internal server error",
			queryParams: "",
			setupMock: func(mockService *MockPATService) {
				mockService.On("ListUserPATs", mock.Anything, userID, 50, 0).Return(nil, errors.New("database error"))
			},
			setupAuth:      true,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPATService)
			tt.setupMock(mockService)

			handler := NewPATHandler(mockService)

			var router *gin.Engine
			if tt.setupAuth {
				router = setupAuthenticatedRouter(userID)
			} else {
				router = setupPATTestRouter()
			}

			router.GET("/api/v1/pats", handler.ListPATs)

			req, _ := http.NewRequest("GET", "/api/v1/pats"+tt.queryParams, nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				errorObj, exists := response["error"].(map[string]interface{})
				assert.True(t, exists)
				assert.Equal(t, tt.expectedError, errorObj["code"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestPATHandler_RevokePAT(t *testing.T) {
	userID := uuid.New()
	patID := uuid.New()

	tests := []struct {
		name           string
		patID          string
		setupMock      func(*MockPATService)
		setupAuth      bool
		expectedStatus int
		expectedError  string
	}{
		{
			name:  "successful PAT revocation",
			patID: patID.String(),
			setupMock: func(mockService *MockPATService) {
				mockService.On("RevokePAT", mock.Anything, patID, userID).Return(nil)
			},
			setupAuth:      true,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid PAT ID format",
			patID:          "invalid-uuid",
			setupMock:      func(mockService *MockPATService) {},
			setupAuth:      true,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "missing authentication",
			patID:          patID.String(),
			setupMock:      func(mockService *MockPATService) {},
			setupAuth:      false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "AUTHENTICATION_REQUIRED",
		},
		{
			name:  "PAT not found",
			patID: patID.String(),
			setupMock: func(mockService *MockPATService) {
				mockService.On("RevokePAT", mock.Anything, patID, userID).Return(service.ErrPATNotFound)
			},
			setupAuth:      true,
			expectedStatus: http.StatusNotFound,
			expectedError:  "ENTITY_NOT_FOUND",
		},
		{
			name:  "unauthorized access to token",
			patID: patID.String(),
			setupMock: func(mockService *MockPATService) {
				mockService.On("RevokePAT", mock.Anything, patID, userID).Return(service.ErrPATUnauthorized)
			},
			setupAuth:      true,
			expectedStatus: http.StatusForbidden,
			expectedError:  "INSUFFICIENT_PERMISSIONS",
		},
		{
			name:  "internal server error",
			patID: patID.String(),
			setupMock: func(mockService *MockPATService) {
				mockService.On("RevokePAT", mock.Anything, patID, userID).Return(errors.New("database error"))
			},
			setupAuth:      true,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPATService)
			tt.setupMock(mockService)

			handler := NewPATHandler(mockService)

			var router *gin.Engine
			if tt.setupAuth {
				router = setupAuthenticatedRouter(userID)
			} else {
				router = setupPATTestRouter()
			}

			router.DELETE("/api/v1/pats/:id", handler.RevokePAT)

			req, _ := http.NewRequest("DELETE", "/api/v1/pats/"+tt.patID, nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				errorObj, exists := response["error"].(map[string]interface{})
				assert.True(t, exists)
				assert.Equal(t, tt.expectedError, errorObj["code"])
			}

			mockService.AssertExpectations(t)
		})
	}
}
