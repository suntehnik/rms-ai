package integration

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
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"product-requirements-management/internal/handlers"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/service"
)

func setupCommentIntegrationTest(t *testing.T) (*gin.Engine, *gorm.DB, func()) {
	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate models
	err = models.AutoMigrate(db)
	require.NoError(t, err)

	// Seed default data
	err = models.SeedDefaultData(db)
	require.NoError(t, err)
	
	// Initialize repositories
	repos := repository.NewRepositories(db)
	
	// Initialize services
	commentService := service.NewCommentService(repos)
	
	// Initialize handlers
	commentHandler := handlers.NewCommentHandler(commentService)
	
	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Setup routes
	v1 := router.Group("/api/v1")
	{
		// Entity comment routes
		v1.POST("/:entityType/:id/comments", commentHandler.CreateComment)
		v1.GET("/:entityType/:id/comments", commentHandler.GetCommentsByEntity)
		
		// Direct comment routes
		v1.GET("/comments/:id", commentHandler.GetComment)
		v1.PUT("/comments/:id", commentHandler.UpdateComment)
		v1.DELETE("/comments/:id", commentHandler.DeleteComment)
		v1.POST("/comments/:id/resolve", commentHandler.ResolveComment)
		v1.POST("/comments/:id/unresolve", commentHandler.UnresolveComment)
		v1.GET("/comments/status/:status", commentHandler.GetCommentsByStatus)
		v1.GET("/comments/:id/replies", commentHandler.GetCommentReplies)
		v1.POST("/comments/:id/replies", commentHandler.CreateCommentReply)
	}
	
	cleanup := func() {
		// Close database connection
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
	
	return router, db, cleanup
}

func createTestUserForComment(t *testing.T, db *gorm.DB) *models.User {
	user := &models.User{
		ID:       uuid.New(),
		Username: fmt.Sprintf("testuser_%s", uuid.New().String()[:8]),
		Email:    fmt.Sprintf("test_%s@example.com", uuid.New().String()[:8]),
		Role:     models.RoleUser,
	}
	
	err := db.Create(user).Error
	require.NoError(t, err)
	
	return user
}

func createTestEpicForComment(t *testing.T, db *gorm.DB, creator *models.User) *models.Epic {
	description := "Test epic description"
	epic := &models.Epic{
		ID:          uuid.New(),
		ReferenceID: fmt.Sprintf("EP-%d", generateSequenceNumber()),
		CreatorID:   creator.ID,
		AssigneeID:  creator.ID,
		Priority:    models.PriorityMedium,
		Status:      models.EpicStatusBacklog,
		Title:       "Test Epic",
		Description: &description,
	}
	
	err := db.Create(epic).Error
	require.NoError(t, err)
	
	return epic
}

func TestCommentIntegration_CreateComment(t *testing.T) {
	router, db, cleanup := setupCommentIntegrationTest(t)
	defer cleanup()

	// Create test data
	user := createTestUserForComment(t, db)
	epic := createTestEpicForComment(t, db, user)

	tests := []struct {
		name           string
		entityType     string
		entityID       string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:       "create general comment",
			entityType: "epic",
			entityID:   epic.ID.String(),
			requestBody: map[string]interface{}{
				"author_id": user.ID.String(),
				"content":   "This is a test comment",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:       "create inline comment",
			entityType: "epic",
			entityID:   epic.ID.String(),
			requestBody: map[string]interface{}{
				"author_id":           user.ID.String(),
				"content":             "This is an inline comment",
				"linked_text":         "selected text",
				"text_position_start": 10,
				"text_position_end":   23,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:       "invalid entity type",
			entityType: "invalid",
			entityID:   epic.ID.String(),
			requestBody: map[string]interface{}{
				"author_id": user.ID.String(),
				"content":   "This is a test comment",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid entity type",
		},
		{
			name:       "entity not found",
			entityType: "epic",
			entityID:   uuid.New().String(),
			requestBody: map[string]interface{}{
				"author_id": user.ID.String(),
				"content":   "This is a test comment",
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Entity not found",
		},
		{
			name:       "author not found",
			entityType: "epic",
			entityID:   epic.ID.String(),
			requestBody: map[string]interface{}{
				"author_id": uuid.New().String(),
				"content":   "This is a test comment",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Author not found",
		},
		{
			name:       "empty content",
			entityType: "epic",
			entityID:   epic.ID.String(),
			requestBody: map[string]interface{}{
				"author_id": user.ID.String(),
				"content":   "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Content cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, 
				fmt.Sprintf("/api/v1/%s/%s/comments", tt.entityType, tt.entityID), 
				bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			} else if tt.expectedStatus == http.StatusCreated {
				var response service.CommentResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, models.EntityType(tt.entityType), response.EntityType)
				assert.Equal(t, tt.requestBody["content"], response.Content)
				assert.False(t, response.IsResolved)
				
				// Check if it's an inline comment
				if linkedText, ok := tt.requestBody["linked_text"]; ok {
					assert.True(t, response.IsInline)
					assert.Equal(t, linkedText, *response.LinkedText)
				} else {
					assert.False(t, response.IsInline)
				}
			}
		})
	}
}

func TestCommentIntegration_GetCommentsByEntity(t *testing.T) {
	router, db, cleanup := setupCommentIntegrationTest(t)
	defer cleanup()

	// Create test data
	user := createTestUserForComment(t, db)
	epic := createTestEpicForComment(t, db, user)

	// Create test comments
	generalComment := &models.Comment{
		ID:         uuid.New(),
		EntityType: models.EntityTypeEpic,
		EntityID:   epic.ID,
		AuthorID:   user.ID,
		Content:    "General comment",
		IsResolved: false,
	}
	
	linkedText := "selected text"
	start := 10
	end := 23
	inlineComment := &models.Comment{
		ID:                uuid.New(),
		EntityType:        models.EntityTypeEpic,
		EntityID:          epic.ID,
		AuthorID:          user.ID,
		Content:           "Inline comment",
		IsResolved:        false,
		LinkedText:        &linkedText,
		TextPositionStart: &start,
		TextPositionEnd:   &end,
	}

	err := db.Create(generalComment).Error
	require.NoError(t, err)
	
	err = db.Create(inlineComment).Error
	require.NoError(t, err)

	tests := []struct {
		name           string
		entityType     string
		entityID       string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "get all comments",
			entityType:     "epic",
			entityID:       epic.ID.String(),
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "get inline comments only",
			entityType:     "epic",
			entityID:       epic.ID.String(),
			queryParams:    "?inline=true",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "get threaded comments",
			entityType:     "epic",
			entityID:       epic.ID.String(),
			queryParams:    "?threaded=true",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "filter by unresolved status",
			entityType:     "epic",
			entityID:       epic.ID.String(),
			queryParams:    "?status=unresolved",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "filter by resolved status",
			entityType:     "epic",
			entityID:       epic.ID.String(),
			queryParams:    "?status=resolved",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "invalid entity type",
			entityType:     "invalid",
			entityID:       epic.ID.String(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "entity not found",
			entityType:     "epic",
			entityID:       uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			url := fmt.Sprintf("/api/v1/%s/%s/comments%s", tt.entityType, tt.entityID, tt.queryParams)
			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				
				comments := response["comments"].([]interface{})
				assert.Equal(t, tt.expectedCount, len(comments))
				assert.Equal(t, float64(tt.expectedCount), response["count"])
			}
		})
	}
}

func TestCommentIntegration_UpdateComment(t *testing.T) {
	router, db, cleanup := setupCommentIntegrationTest(t)
	defer cleanup()

	// Create test data
	user := createTestUserForComment(t, db)
	epic := createTestEpicForComment(t, db, user)

	// Create test comment
	comment := &models.Comment{
		ID:         uuid.New(),
		EntityType: models.EntityTypeEpic,
		EntityID:   epic.ID,
		AuthorID:   user.ID,
		Content:    "Original content",
		IsResolved: false,
	}

	err := db.Create(comment).Error
	require.NoError(t, err)

	tests := []struct {
		name           string
		commentID      string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "successful update",
			commentID: comment.ID.String(),
			requestBody: map[string]interface{}{
				"content": "Updated content",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "empty content",
			commentID: comment.ID.String(),
			requestBody: map[string]interface{}{
				"content": "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Content cannot be empty",
		},
		{
			name:      "comment not found",
			commentID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"content": "Updated content",
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Comment not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, 
				fmt.Sprintf("/api/v1/comments/%s", tt.commentID), 
				bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			} else if tt.expectedStatus == http.StatusOK {
				var response service.CommentResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.requestBody["content"], response.Content)
			}
		})
	}
}

func TestCommentIntegration_ResolveComment(t *testing.T) {
	router, db, cleanup := setupCommentIntegrationTest(t)
	defer cleanup()

	// Create test data
	user := createTestUserForComment(t, db)
	epic := createTestEpicForComment(t, db, user)

	// Create test comment
	comment := &models.Comment{
		ID:         uuid.New(),
		EntityType: models.EntityTypeEpic,
		EntityID:   epic.ID,
		AuthorID:   user.ID,
		Content:    "Test comment",
		IsResolved: false,
	}

	err := db.Create(comment).Error
	require.NoError(t, err)

	// Test resolve comment
	req := httptest.NewRequest(http.MethodPost, 
		fmt.Sprintf("/api/v1/comments/%s/resolve", comment.ID.String()), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response service.CommentResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.IsResolved)

	// Test unresolve comment
	req = httptest.NewRequest(http.MethodPost, 
		fmt.Sprintf("/api/v1/comments/%s/unresolve", comment.ID.String()), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.IsResolved)
}

func TestCommentIntegration_DeleteComment(t *testing.T) {
	router, db, cleanup := setupCommentIntegrationTest(t)
	defer cleanup()

	// Create test data
	user := createTestUserForComment(t, db)
	epic := createTestEpicForComment(t, db, user)

	// Create test comment
	comment := &models.Comment{
		ID:         uuid.New(),
		EntityType: models.EntityTypeEpic,
		EntityID:   epic.ID,
		AuthorID:   user.ID,
		Content:    "Test comment",
		IsResolved: false,
	}

	err := db.Create(comment).Error
	require.NoError(t, err)

	// Test delete comment
	req := httptest.NewRequest(http.MethodDelete, 
		fmt.Sprintf("/api/v1/comments/%s", comment.ID.String()), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify comment is deleted
	var deletedComment models.Comment
	err = db.First(&deletedComment, comment.ID).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestCommentIntegration_GetCommentsByStatus(t *testing.T) {
	router, db, cleanup := setupCommentIntegrationTest(t)
	defer cleanup()

	// Create test data
	user := createTestUserForComment(t, db)
	epic := createTestEpicForComment(t, db, user)

	// Create resolved comment
	resolvedComment := &models.Comment{
		ID:         uuid.New(),
		EntityType: models.EntityTypeEpic,
		EntityID:   epic.ID,
		AuthorID:   user.ID,
		Content:    "Resolved comment",
		IsResolved: true,
	}

	// Create unresolved comment
	unresolvedComment := &models.Comment{
		ID:         uuid.New(),
		EntityType: models.EntityTypeEpic,
		EntityID:   epic.ID,
		AuthorID:   user.ID,
		Content:    "Unresolved comment",
		IsResolved: false,
	}

	err := db.Create(resolvedComment).Error
	require.NoError(t, err)
	
	err = db.Create(unresolvedComment).Error
	require.NoError(t, err)

	tests := []struct {
		name           string
		status         string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "get resolved comments",
			status:         "resolved",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "get unresolved comments",
			status:         "unresolved",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "invalid status",
			status:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, 
				fmt.Sprintf("/api/v1/comments/status/%s", tt.status), nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				
				comments := response["comments"].([]interface{})
				assert.Equal(t, tt.expectedCount, len(comments))
				assert.Equal(t, float64(tt.expectedCount), response["count"])
				assert.Equal(t, tt.status, response["status"])
			}
		})
	}
}

func TestCommentIntegration_CommentThreading(t *testing.T) {
	router, db, cleanup := setupCommentIntegrationTest(t)
	defer cleanup()

	// Create test data
	user := createTestUserForComment(t, db)
	epic := createTestEpicForComment(t, db, user)

	// Create parent comment
	parentComment := &models.Comment{
		ID:         uuid.New(),
		EntityType: models.EntityTypeEpic,
		EntityID:   epic.ID,
		AuthorID:   user.ID,
		Content:    "Parent comment",
		IsResolved: false,
	}

	err := db.Create(parentComment).Error
	require.NoError(t, err)

	// Create reply using API
	replyBody := map[string]interface{}{
		"author_id": user.ID.String(),
		"content":   "Reply to parent comment",
	}

	body, _ := json.Marshal(replyBody)
	req := httptest.NewRequest(http.MethodPost, 
		fmt.Sprintf("/api/v1/comments/%s/replies", parentComment.ID.String()), 
		bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var replyResponse service.CommentResponse
	err = json.Unmarshal(w.Body.Bytes(), &replyResponse)
	assert.NoError(t, err)
	assert.Equal(t, replyBody["content"], replyResponse.Content)
	assert.Equal(t, parentComment.ID, *replyResponse.ParentCommentID)
	assert.True(t, replyResponse.IsReply)
}

// Helper function to generate sequence numbers for testing
var sequenceCounter = 1

func generateSequenceNumber() int {
	sequenceCounter++
	return sequenceCounter
}