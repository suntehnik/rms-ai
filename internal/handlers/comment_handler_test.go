package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// MockCommentService is a mock implementation of CommentService
type MockCommentService struct {
	mock.Mock
}

func (m *MockCommentService) CreateComment(req service.CreateCommentRequest) (*service.CommentResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.CommentResponse), args.Error(1)
}

func (m *MockCommentService) GetComment(id uuid.UUID) (*service.CommentResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.CommentResponse), args.Error(1)
}

func (m *MockCommentService) UpdateComment(id uuid.UUID, req service.UpdateCommentRequest) (*service.CommentResponse, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.CommentResponse), args.Error(1)
}

func (m *MockCommentService) DeleteComment(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCommentService) GetCommentsByEntity(entityType models.EntityType, entityID uuid.UUID) ([]service.CommentResponse, error) {
	args := m.Called(entityType, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.CommentResponse), args.Error(1)
}

func (m *MockCommentService) GetThreadedComments(entityType models.EntityType, entityID uuid.UUID) ([]service.CommentResponse, error) {
	args := m.Called(entityType, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.CommentResponse), args.Error(1)
}

func (m *MockCommentService) GetCommentsByStatus(isResolved bool) ([]service.CommentResponse, error) {
	args := m.Called(isResolved)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.CommentResponse), args.Error(1)
}

func (m *MockCommentService) GetInlineComments(entityType models.EntityType, entityID uuid.UUID) ([]service.CommentResponse, error) {
	args := m.Called(entityType, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.CommentResponse), args.Error(1)
}

func (m *MockCommentService) ResolveComment(id uuid.UUID) (*service.CommentResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.CommentResponse), args.Error(1)
}

func (m *MockCommentService) UnresolveComment(id uuid.UUID) (*service.CommentResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.CommentResponse), args.Error(1)
}

func (m *MockCommentService) GetVisibleInlineComments(entityType models.EntityType, entityID uuid.UUID) ([]service.CommentResponse, error) {
	args := m.Called(entityType, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.CommentResponse), args.Error(1)
}

func (m *MockCommentService) ValidateInlineCommentsAfterTextChange(entityType models.EntityType, entityID uuid.UUID, newDescription string) error {
	args := m.Called(entityType, entityID, newDescription)
	return args.Error(0)
}

func (m *MockCommentService) GetCommentReplies(parentID uuid.UUID) ([]service.CommentResponse, error) {
	args := m.Called(parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.CommentResponse), args.Error(1)
}

func setupCommentHandler() (*CommentHandler, *MockCommentService, *auth.Service) {
	mockService := &MockCommentService{}
	handler := NewCommentHandler(mockService)
	authService := auth.NewService("test-secret-key", 24*time.Hour)
	return handler, mockService, authService
}

func setupGinRouter(handler *CommentHandler, authService *auth.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Comment routes
	v1 := router.Group("/api/v1")
	{
		// Direct comment routes (require authentication)
		authenticated := v1.Group("")
		authenticated.Use(authService.Middleware())
		authenticated.Use(authService.RequireCommenter())
		{
			authenticated.GET("/comments/:id", handler.GetComment)
			authenticated.PUT("/comments/:id", handler.UpdateComment)
			authenticated.DELETE("/comments/:id", handler.DeleteComment)
			authenticated.POST("/comments/:id/resolve", handler.ResolveComment)
			authenticated.POST("/comments/:id/unresolve", handler.UnresolveComment)
			authenticated.GET("/comments/status/:status", handler.GetCommentsByStatus)
			authenticated.GET("/comments/:id/replies", handler.GetCommentReplies)
			authenticated.POST("/comments/:id/replies", handler.CreateCommentReply)

			// Entity-specific comment routes (matching actual application routes)
			epics := authenticated.Group("/epics")
			{
				epics.GET("/:id/comments", handler.GetEpicComments)
				epics.POST("/:id/comments", handler.CreateEpicComment)
			}

			userStories := authenticated.Group("/user-stories")
			{
				userStories.GET("/:id/comments", handler.GetUserStoryComments)
				userStories.POST("/:id/comments", handler.CreateUserStoryComment)
			}

			acceptanceCriteria := authenticated.Group("/acceptance-criteria")
			{
				acceptanceCriteria.GET("/:id/comments", handler.GetAcceptanceCriteriaComments)
				acceptanceCriteria.POST("/:id/comments", handler.CreateAcceptanceCriteriaComment)
			}

			requirements := authenticated.Group("/requirements")
			{
				requirements.GET("/:id/comments", handler.GetRequirementComments)
				requirements.POST("/:id/comments", handler.CreateRequirementComment)
			}
		}
	}

	return router
}

// createTestUser creates a test user for authentication
func createTestUser() *models.User {
	return &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}
}

// createTestToken creates a JWT token for testing
func createTestToken(authService *auth.Service, user *models.User) (string, error) {
	return authService.GenerateToken(user)
}

// addAuthHeader adds authentication header to request
func addAuthHeader(req *http.Request, token string) {
	req.Header.Set("Authorization", "Bearer "+token)
}

func TestCreateComment(t *testing.T) {
	handler, mockService, authService := setupCommentHandler()
	router := setupGinRouter(handler, authService)

	// Create test user and token
	testUser := createTestUser()
	token, err := createTestToken(authService, testUser)
	assert.NoError(t, err)

	entityID := uuid.New()
	commentID := uuid.New()

	tests := []struct {
		name           string
		entityType     string
		entityID       string
		requestBody    map[string]interface{}
		mockSetup      func()
		expectedStatus int
		expectedError  string
		useAuth        bool
	}{
		{
			name:       "successful comment creation",
			entityType: "epics",
			entityID:   entityID.String(),
			requestBody: map[string]interface{}{
				"content": "This is a test comment",
			},
			mockSetup: func() {
				expectedReq := service.CreateCommentRequest{
					EntityType: models.EntityTypeEpic,
					EntityID:   entityID,
					AuthorID:   testUser.ID,
					Content:    "This is a test comment",
				}
				expectedResponse := &service.CommentResponse{
					ID:         commentID,
					EntityType: models.EntityTypeEpic,
					EntityID:   entityID,
					AuthorID:   testUser.ID,
					Content:    "This is a test comment",
					IsResolved: false,
				}
				mockService.On("CreateComment", expectedReq).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusCreated,
			useAuth:        true,
		},
		{
			name:       "unauthorized - no token",
			entityType: "epics",
			entityID:   entityID.String(),
			requestBody: map[string]interface{}{
				"content": "This is a test comment",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization header required",
			useAuth:        false,
		},
		{
			name:       "invalid entity ID",
			entityType: "epics",
			entityID:   "invalid-uuid",
			requestBody: map[string]interface{}{
				"content": "This is a test comment",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid entity ID format",
			useAuth:        true,
		},
		{
			name:       "empty content",
			entityType: "epics",
			entityID:   entityID.String(),
			requestBody: map[string]interface{}{
				"content": "",
			},
			mockSetup: func() {
				mockService.On("CreateComment", mock.AnythingOfType("service.CreateCommentRequest")).
					Return(nil, service.ErrEmptyContent)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Content cannot be empty",
			useAuth:        true,
		},
		{
			name:       "author not found",
			entityType: "epics",
			entityID:   entityID.String(),
			requestBody: map[string]interface{}{
				"content": "This is a test comment",
			},
			mockSetup: func() {
				mockService.On("CreateComment", mock.AnythingOfType("service.CreateCommentRequest")).
					Return(nil, service.ErrCommentAuthorNotFound)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Author not found",
			useAuth:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			tt.mockSetup()

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost,
				fmt.Sprintf("/api/v1/%s/%s/comments", tt.entityType, tt.entityID),
				bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Add auth header if needed
			if tt.useAuth {
				addAuthHeader(req, token)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Assert error message if expected
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetCommentsByEntity(t *testing.T) {
	handler, mockService, authService := setupCommentHandler()
	router := setupGinRouter(handler, authService)

	// Create test user and token
	testUser := createTestUser()
	token, err := createTestToken(authService, testUser)
	assert.NoError(t, err)

	entityID := uuid.New()
	commentID := uuid.New()
	authorID := uuid.New()

	tests := []struct {
		name           string
		entityType     string
		entityID       string
		queryParams    string
		mockSetup      func()
		expectedStatus int
		expectedError  string
		useAuth        bool
	}{
		{
			name:       "successful get comments",
			entityType: "epics",
			entityID:   entityID.String(),
			mockSetup: func() {
				expectedComments := []service.CommentResponse{
					{
						ID:         commentID,
						EntityType: models.EntityTypeEpic,
						EntityID:   entityID,
						AuthorID:   authorID,
						Content:    "Test comment",
						IsResolved: false,
					},
				}
				mockService.On("GetCommentsByEntity", models.EntityTypeEpic, entityID).
					Return(expectedComments, nil)
			},
			expectedStatus: http.StatusOK,
			useAuth:        true,
		},
		{
			name:           "unauthorized - no token",
			entityType:     "epics",
			entityID:       entityID.String(),
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization header required",
			useAuth:        false,
		},
		{
			name:        "get threaded comments",
			entityType:  "epics",
			entityID:    entityID.String(),
			queryParams: "?threaded=true",
			mockSetup: func() {
				expectedComments := []service.CommentResponse{
					{
						ID:         commentID,
						EntityType: models.EntityTypeEpic,
						EntityID:   entityID,
						AuthorID:   authorID,
						Content:    "Test comment",
						IsResolved: false,
						Replies:    []service.CommentResponse{},
					},
				}
				mockService.On("GetThreadedComments", models.EntityTypeEpic, entityID).
					Return(expectedComments, nil)
			},
			expectedStatus: http.StatusOK,
			useAuth:        true,
		},
		{
			name:        "get inline comments",
			entityType:  "epics",
			entityID:    entityID.String(),
			queryParams: "?inline=true",
			mockSetup: func() {
				linkedText := "selected text"
				start := 10
				end := 23
				expectedComments := []service.CommentResponse{
					{
						ID:                commentID,
						EntityType:        models.EntityTypeEpic,
						EntityID:          entityID,
						AuthorID:          authorID,
						Content:           "Inline comment",
						IsResolved:        false,
						LinkedText:        &linkedText,
						TextPositionStart: &start,
						TextPositionEnd:   &end,
						IsInline:          true,
					},
				}
				mockService.On("GetVisibleInlineComments", models.EntityTypeEpic, entityID).
					Return(expectedComments, nil)
			},
			expectedStatus: http.StatusOK,
			useAuth:        true,
		},
		{
			name:           "invalid entity ID",
			entityType:     "epics",
			entityID:       "invalid-uuid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid entity ID format",
			useAuth:        true,
		},
		{
			name:       "entity not found",
			entityType: "epics",
			entityID:   entityID.String(),
			mockSetup: func() {
				mockService.On("GetCommentsByEntity", models.EntityTypeEpic, entityID).
					Return(nil, service.ErrCommentEntityNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Entity not found",
			useAuth:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			tt.mockSetup()

			// Create request
			url := fmt.Sprintf("/api/v1/%s/%s/comments%s", tt.entityType, tt.entityID, tt.queryParams)
			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Add auth header if needed
			if tt.useAuth {
				addAuthHeader(req, token)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Assert error message if expected
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetComment(t *testing.T) {
	handler, mockService, authService := setupCommentHandler()
	router := setupGinRouter(handler, authService)

	// Create test user and token
	testUser := createTestUser()
	token, err := createTestToken(authService, testUser)
	assert.NoError(t, err)

	commentID := uuid.New()
	entityID := uuid.New()
	authorID := uuid.New()

	tests := []struct {
		name           string
		commentID      string
		mockSetup      func()
		expectedStatus int
		expectedError  string
		useAuth        bool
	}{
		{
			name:      "successful get comment",
			commentID: commentID.String(),
			mockSetup: func() {
				expectedComment := &service.CommentResponse{
					ID:         commentID,
					EntityType: models.EntityTypeEpic,
					EntityID:   entityID,
					AuthorID:   authorID,
					Content:    "Test comment",
					IsResolved: false,
				}
				mockService.On("GetComment", commentID).Return(expectedComment, nil)
			},
			expectedStatus: http.StatusOK,
			useAuth:        true,
		},
		{
			name:           "unauthorized - no token",
			commentID:      commentID.String(),
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization header required",
			useAuth:        false,
		},
		{
			name:           "invalid comment ID",
			commentID:      "invalid-uuid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid comment ID format",
			useAuth:        true,
		},
		{
			name:      "comment not found",
			commentID: commentID.String(),
			mockSetup: func() {
				mockService.On("GetComment", commentID).Return(nil, service.ErrCommentNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Comment not found",
			useAuth:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			tt.mockSetup()

			// Create request
			req := httptest.NewRequest(http.MethodGet,
				fmt.Sprintf("/api/v1/comments/%s", tt.commentID), nil)

			// Add auth header if needed
			if tt.useAuth {
				addAuthHeader(req, token)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Assert error message if expected
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestUpdateComment(t *testing.T) {
	handler, mockService, authService := setupCommentHandler()
	router := setupGinRouter(handler, authService)

	// Create test user and token
	testUser := createTestUser()
	token, err := createTestToken(authService, testUser)
	assert.NoError(t, err)

	commentID := uuid.New()
	entityID := uuid.New()
	authorID := uuid.New()

	tests := []struct {
		name           string
		commentID      string
		requestBody    map[string]interface{}
		mockSetup      func()
		expectedStatus int
		expectedError  string
		useAuth        bool
	}{
		{
			name:      "successful update comment",
			commentID: commentID.String(),
			requestBody: map[string]interface{}{
				"content": "Updated comment content",
			},
			mockSetup: func() {
				expectedReq := service.UpdateCommentRequest{
					Content: "Updated comment content",
				}
				expectedResponse := &service.CommentResponse{
					ID:         commentID,
					EntityType: models.EntityTypeEpic,
					EntityID:   entityID,
					AuthorID:   authorID,
					Content:    "Updated comment content",
					IsResolved: false,
				}
				mockService.On("UpdateComment", commentID, expectedReq).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			useAuth:        true,
		},
		{
			name:      "unauthorized - no token",
			commentID: commentID.String(),
			requestBody: map[string]interface{}{
				"content": "Updated comment content",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization header required",
			useAuth:        false,
		},
		{
			name:      "empty content",
			commentID: commentID.String(),
			requestBody: map[string]interface{}{
				"content": "",
			},
			mockSetup: func() {
				mockService.On("UpdateComment", commentID, mock.AnythingOfType("service.UpdateCommentRequest")).
					Return(nil, service.ErrEmptyContent)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Content cannot be empty",
			useAuth:        true,
		},
		{
			name:      "comment not found",
			commentID: commentID.String(),
			requestBody: map[string]interface{}{
				"content": "Updated content",
			},
			mockSetup: func() {
				mockService.On("UpdateComment", commentID, mock.AnythingOfType("service.UpdateCommentRequest")).
					Return(nil, service.ErrCommentNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Comment not found",
			useAuth:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			tt.mockSetup()

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut,
				fmt.Sprintf("/api/v1/comments/%s", tt.commentID),
				bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Add auth header if needed
			if tt.useAuth {
				addAuthHeader(req, token)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Assert error message if expected
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestDeleteComment(t *testing.T) {
	handler, mockService, authService := setupCommentHandler()
	router := setupGinRouter(handler, authService)

	// Create test user and token
	testUser := createTestUser()
	token, err := createTestToken(authService, testUser)
	assert.NoError(t, err)

	commentID := uuid.New()

	tests := []struct {
		name           string
		commentID      string
		mockSetup      func()
		expectedStatus int
		expectedError  string
		useAuth        bool
	}{
		{
			name:      "successful delete comment",
			commentID: commentID.String(),
			mockSetup: func() {
				mockService.On("DeleteComment", commentID).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			useAuth:        true,
		},
		{
			name:           "unauthorized - no token",
			commentID:      commentID.String(),
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization header required",
			useAuth:        false,
		},
		{
			name:           "invalid comment ID",
			commentID:      "invalid-uuid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid comment ID format",
			useAuth:        true,
		},
		{
			name:      "comment not found",
			commentID: commentID.String(),
			mockSetup: func() {
				mockService.On("DeleteComment", commentID).Return(service.ErrCommentNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Comment not found",
			useAuth:        true,
		},
		{
			name:      "comment has replies",
			commentID: commentID.String(),
			mockSetup: func() {
				mockService.On("DeleteComment", commentID).Return(service.ErrCommentHasReplies)
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "Comment has replies and cannot be deleted",
			useAuth:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			tt.mockSetup()

			// Create request
			req := httptest.NewRequest(http.MethodDelete,
				fmt.Sprintf("/api/v1/comments/%s", tt.commentID), nil)

			// Add auth header if needed
			if tt.useAuth {
				addAuthHeader(req, token)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Assert error message if expected
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestResolveComment(t *testing.T) {
	handler, mockService, authService := setupCommentHandler()
	router := setupGinRouter(handler, authService)

	// Create test user and token
	testUser := createTestUser()
	token, err := createTestToken(authService, testUser)
	assert.NoError(t, err)

	commentID := uuid.New()
	entityID := uuid.New()
	authorID := uuid.New()

	tests := []struct {
		name           string
		commentID      string
		mockSetup      func()
		expectedStatus int
		expectedError  string
		useAuth        bool
	}{
		{
			name:      "successful resolve comment",
			commentID: commentID.String(),
			mockSetup: func() {
				expectedResponse := &service.CommentResponse{
					ID:         commentID,
					EntityType: models.EntityTypeEpic,
					EntityID:   entityID,
					AuthorID:   authorID,
					Content:    "Test comment",
					IsResolved: true,
				}
				mockService.On("ResolveComment", commentID).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			useAuth:        true,
		},
		{
			name:           "unauthorized - no token",
			commentID:      commentID.String(),
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization header required",
			useAuth:        false,
		},
		{
			name:           "invalid comment ID",
			commentID:      "invalid-uuid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid comment ID format",
			useAuth:        true,
		},
		{
			name:      "comment not found",
			commentID: commentID.String(),
			mockSetup: func() {
				mockService.On("ResolveComment", commentID).Return(nil, service.ErrCommentNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Comment not found",
			useAuth:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			tt.mockSetup()

			// Create request
			req := httptest.NewRequest(http.MethodPost,
				fmt.Sprintf("/api/v1/comments/%s/resolve", tt.commentID), nil)

			// Add auth header if needed
			if tt.useAuth {
				addAuthHeader(req, token)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Assert error message if expected
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetCommentsByStatus(t *testing.T) {
	handler, mockService, authService := setupCommentHandler()
	router := setupGinRouter(handler, authService)

	// Create test user and token
	testUser := createTestUser()
	token, err := createTestToken(authService, testUser)
	assert.NoError(t, err)

	commentID := uuid.New()
	entityID := uuid.New()
	authorID := uuid.New()

	tests := []struct {
		name           string
		status         string
		mockSetup      func()
		expectedStatus int
		expectedError  string
		useAuth        bool
	}{
		{
			name:   "get resolved comments",
			status: "resolved",
			mockSetup: func() {
				expectedComments := []service.CommentResponse{
					{
						ID:         commentID,
						EntityType: models.EntityTypeEpic,
						EntityID:   entityID,
						AuthorID:   authorID,
						Content:    "Resolved comment",
						IsResolved: true,
					},
				}
				mockService.On("GetCommentsByStatus", true).Return(expectedComments, nil)
			},
			expectedStatus: http.StatusOK,
			useAuth:        true,
		},
		{
			name:   "get unresolved comments",
			status: "unresolved",
			mockSetup: func() {
				expectedComments := []service.CommentResponse{
					{
						ID:         commentID,
						EntityType: models.EntityTypeEpic,
						EntityID:   entityID,
						AuthorID:   authorID,
						Content:    "Unresolved comment",
						IsResolved: false,
					},
				}
				mockService.On("GetCommentsByStatus", false).Return(expectedComments, nil)
			},
			expectedStatus: http.StatusOK,
			useAuth:        true,
		},
		{
			name:           "unauthorized - no token",
			status:         "resolved",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization header required",
			useAuth:        false,
		},
		{
			name:           "invalid status",
			status:         "invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid status. Use 'resolved' or 'unresolved'",
			useAuth:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			tt.mockSetup()

			// Create request
			req := httptest.NewRequest(http.MethodGet,
				fmt.Sprintf("/api/v1/comments/status/%s", tt.status), nil)

			// Add auth header if needed
			if tt.useAuth {
				addAuthHeader(req, token)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Assert error message if expected
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestCommentFiltering(t *testing.T) {
	handler, mockService, authService := setupCommentHandler()
	router := setupGinRouter(handler, authService)

	// Create test user and token
	testUser := createTestUser()
	token, err := createTestToken(authService, testUser)
	assert.NoError(t, err)

	entityID := uuid.New()
	resolvedCommentID := uuid.New()
	unresolvedCommentID := uuid.New()
	authorID := uuid.New()

	tests := []struct {
		name           string
		entityType     string
		entityID       string
		queryParams    string
		mockSetup      func()
		expectedStatus int
		expectedCount  int
		description    string
	}{
		{
			name:        "filter resolved comments only",
			entityType:  "epics",
			entityID:    entityID.String(),
			queryParams: "?status=resolved",
			mockSetup: func() {
				allComments := []service.CommentResponse{
					{
						ID:         resolvedCommentID,
						EntityType: models.EntityTypeEpic,
						EntityID:   entityID,
						AuthorID:   authorID,
						Content:    "Resolved comment",
						IsResolved: true,
					},
					{
						ID:         unresolvedCommentID,
						EntityType: models.EntityTypeEpic,
						EntityID:   entityID,
						AuthorID:   authorID,
						Content:    "Unresolved comment",
						IsResolved: false,
					},
				}
				mockService.On("GetCommentsByEntity", models.EntityTypeEpic, entityID).
					Return(allComments, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			description:    "Should return only resolved comments when status=resolved",
		},
		{
			name:        "filter unresolved comments only",
			entityType:  "epics",
			entityID:    entityID.String(),
			queryParams: "?status=unresolved",
			mockSetup: func() {
				allComments := []service.CommentResponse{
					{
						ID:         resolvedCommentID,
						EntityType: models.EntityTypeEpic,
						EntityID:   entityID,
						AuthorID:   authorID,
						Content:    "Resolved comment",
						IsResolved: true,
					},
					{
						ID:         unresolvedCommentID,
						EntityType: models.EntityTypeEpic,
						EntityID:   entityID,
						AuthorID:   authorID,
						Content:    "Unresolved comment",
						IsResolved: false,
					},
				}
				mockService.On("GetCommentsByEntity", models.EntityTypeEpic, entityID).
					Return(allComments, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			description:    "Should return only unresolved comments when status=unresolved",
		},
		{
			name:        "no status filter returns all comments",
			entityType:  "epics",
			entityID:    entityID.String(),
			queryParams: "",
			mockSetup: func() {
				allComments := []service.CommentResponse{
					{
						ID:         resolvedCommentID,
						EntityType: models.EntityTypeEpic,
						EntityID:   entityID,
						AuthorID:   authorID,
						Content:    "Resolved comment",
						IsResolved: true,
					},
					{
						ID:         unresolvedCommentID,
						EntityType: models.EntityTypeEpic,
						EntityID:   entityID,
						AuthorID:   authorID,
						Content:    "Unresolved comment",
						IsResolved: false,
					},
				}
				mockService.On("GetCommentsByEntity", models.EntityTypeEpic, entityID).
					Return(allComments, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			description:    "Should return all comments when no status filter is applied",
		},
		{
			name:        "invalid status filter returns all comments",
			entityType:  "epics",
			entityID:    entityID.String(),
			queryParams: "?status=invalid",
			mockSetup: func() {
				allComments := []service.CommentResponse{
					{
						ID:         resolvedCommentID,
						EntityType: models.EntityTypeEpic,
						EntityID:   entityID,
						AuthorID:   authorID,
						Content:    "Resolved comment",
						IsResolved: true,
					},
					{
						ID:         unresolvedCommentID,
						EntityType: models.EntityTypeEpic,
						EntityID:   entityID,
						AuthorID:   authorID,
						Content:    "Unresolved comment",
						IsResolved: false,
					},
				}
				mockService.On("GetCommentsByEntity", models.EntityTypeEpic, entityID).
					Return(allComments, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			description:    "Should return all comments when invalid status filter is provided",
		},
		{
			name:        "filter resolved comments with threaded view",
			entityType:  "user-stories",
			entityID:    entityID.String(),
			queryParams: "?threaded=true&status=resolved",
			mockSetup: func() {
				allComments := []service.CommentResponse{
					{
						ID:         resolvedCommentID,
						EntityType: models.EntityTypeUserStory,
						EntityID:   entityID,
						AuthorID:   authorID,
						Content:    "Resolved parent comment",
						IsResolved: true,
						Replies: []service.CommentResponse{
							{
								ID:         uuid.New(),
								EntityType: models.EntityTypeUserStory,
								EntityID:   entityID,
								AuthorID:   authorID,
								Content:    "Resolved reply",
								IsResolved: true,
							},
						},
					},
					{
						ID:         unresolvedCommentID,
						EntityType: models.EntityTypeUserStory,
						EntityID:   entityID,
						AuthorID:   authorID,
						Content:    "Unresolved comment",
						IsResolved: false,
					},
				}
				mockService.On("GetThreadedComments", models.EntityTypeUserStory, entityID).
					Return(allComments, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			description:    "Should filter resolved comments in threaded view",
		},
		{
			name:        "filter resolved inline comments",
			entityType:  "requirements",
			entityID:    entityID.String(),
			queryParams: "?inline=true&status=resolved",
			mockSetup: func() {
				linkedText := "selected text"
				start := 10
				end := 23
				allComments := []service.CommentResponse{
					{
						ID:                resolvedCommentID,
						EntityType:        models.EntityTypeRequirement,
						EntityID:          entityID,
						AuthorID:          authorID,
						Content:           "Resolved inline comment",
						IsResolved:        true,
						LinkedText:        &linkedText,
						TextPositionStart: &start,
						TextPositionEnd:   &end,
						IsInline:          true,
					},
					{
						ID:                unresolvedCommentID,
						EntityType:        models.EntityTypeRequirement,
						EntityID:          entityID,
						AuthorID:          authorID,
						Content:           "Unresolved inline comment",
						IsResolved:        false,
						LinkedText:        &linkedText,
						TextPositionStart: &start,
						TextPositionEnd:   &end,
						IsInline:          true,
					},
				}
				mockService.On("GetVisibleInlineComments", models.EntityTypeRequirement, entityID).
					Return(allComments, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			description:    "Should filter resolved inline comments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			tt.mockSetup()

			// Create request
			url := fmt.Sprintf("/api/v1/%s/%s/comments%s", tt.entityType, tt.entityID, tt.queryParams)
			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Add auth header
			addAuthHeader(req, token)

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)

			// Parse response
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Assert comment count
			comments, ok := response["comments"].([]interface{})
			assert.True(t, ok, "Response should contain comments array")
			assert.Equal(t, tt.expectedCount, len(comments), tt.description)

			// Assert count field
			count, ok := response["count"].(float64)
			assert.True(t, ok, "Response should contain count field")
			assert.Equal(t, float64(tt.expectedCount), count, tt.description)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestUnresolveComment(t *testing.T) {
	handler, mockService, authService := setupCommentHandler()
	router := setupGinRouter(handler, authService)

	// Create test user and token
	testUser := createTestUser()
	token, err := createTestToken(authService, testUser)
	assert.NoError(t, err)

	commentID := uuid.New()
	entityID := uuid.New()
	authorID := uuid.New()

	tests := []struct {
		name           string
		commentID      string
		mockSetup      func()
		expectedStatus int
		expectedError  string
		useAuth        bool
	}{
		{
			name:      "successful unresolve comment",
			commentID: commentID.String(),
			mockSetup: func() {
				expectedResponse := &service.CommentResponse{
					ID:         commentID,
					EntityType: models.EntityTypeEpic,
					EntityID:   entityID,
					AuthorID:   authorID,
					Content:    "Test comment",
					IsResolved: false,
				}
				mockService.On("UnresolveComment", commentID).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			useAuth:        true,
		},
		{
			name:           "unauthorized - no token",
			commentID:      commentID.String(),
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization header required",
			useAuth:        false,
		},
		{
			name:           "invalid comment ID",
			commentID:      "invalid-uuid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid comment ID format",
			useAuth:        true,
		},
		{
			name:      "comment not found",
			commentID: commentID.String(),
			mockSetup: func() {
				mockService.On("UnresolveComment", commentID).Return(nil, service.ErrCommentNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Comment not found",
			useAuth:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			tt.mockSetup()

			// Create request
			req := httptest.NewRequest(http.MethodPost,
				fmt.Sprintf("/api/v1/comments/%s/unresolve", tt.commentID), nil)

			// Add auth header if needed
			if tt.useAuth {
				addAuthHeader(req, token)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Assert error message if expected
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}
