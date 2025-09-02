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

func setupInlineCommentIntegrationTest(t *testing.T) (*gin.Engine, *gorm.DB, func()) {
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
	
	// Setup routes for inline comments
	v1 := router.Group("/api/v1")
	{
		epics := v1.Group("/epics")
		{
			epics.POST("/:id/comments", commentHandler.CreateEpicComment)
			epics.POST("/:id/comments/inline", commentHandler.CreateEpicInlineComment)
			epics.GET("/:id/comments/inline/visible", commentHandler.GetEpicVisibleInlineComments)
			epics.POST("/:id/comments/inline/validate", commentHandler.ValidateEpicInlineComments)
			epics.GET("/:id/comments", commentHandler.GetEpicComments)
		}
		
		userStories := v1.Group("/user-stories")
		{
			userStories.POST("/:id/comments/inline", commentHandler.CreateUserStoryInlineComment)
			userStories.GET("/:id/comments/inline/visible", commentHandler.GetUserStoryVisibleInlineComments)
			userStories.POST("/:id/comments/inline/validate", commentHandler.ValidateUserStoryInlineComments)
		}
		
		acceptanceCriteria := v1.Group("/acceptance-criteria")
		{
			acceptanceCriteria.POST("/:id/comments/inline", commentHandler.CreateAcceptanceCriteriaInlineComment)
			acceptanceCriteria.GET("/:id/comments/inline/visible", commentHandler.GetAcceptanceCriteriaVisibleInlineComments)
			acceptanceCriteria.POST("/:id/comments/inline/validate", commentHandler.ValidateAcceptanceCriteriaInlineComments)
		}
		
		requirements := v1.Group("/requirements")
		{
			requirements.POST("/:id/comments/inline", commentHandler.CreateRequirementInlineComment)
			requirements.GET("/:id/comments/inline/visible", commentHandler.GetRequirementVisibleInlineComments)
			requirements.POST("/:id/comments/inline/validate", commentHandler.ValidateRequirementInlineComments)
		}
	}
	
	cleanup := func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}
	
	return router, db, cleanup
}



func createInlineTestEpic(t *testing.T, db *gorm.DB, creator *models.User) *models.Epic {
	description := "This is a test epic description for inline comments."
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

func createInlineTestUserStory(t *testing.T, db *gorm.DB, epicID, creatorID uuid.UUID) *models.UserStory {
	description := "As a user, I want to test inline comments, so that I can verify functionality."
	userStory := &models.UserStory{
		ID:          uuid.New(),
		ReferenceID: fmt.Sprintf("US-%d", generateSequenceNumber()),
		EpicID:      epicID,
		CreatorID:   creatorID,
		AssigneeID:  creatorID,
		Priority:    models.PriorityMedium,
		Status:      models.UserStoryStatusBacklog,
		Title:       "Test User Story",
		Description: &description,
	}
	
	err := db.Create(userStory).Error
	require.NoError(t, err)
	
	return userStory
}

func createInlineTestAcceptanceCriteria(t *testing.T, db *gorm.DB, userStoryID, authorID uuid.UUID) *models.AcceptanceCriteria {
	acceptanceCriteria := &models.AcceptanceCriteria{
		ID:          uuid.New(),
		ReferenceID: fmt.Sprintf("AC-%d", generateSequenceNumber()),
		UserStoryID: userStoryID,
		AuthorID:    authorID,
		Description: "WHEN user creates inline comment THEN system SHALL save it properly",
	}
	
	err := db.Create(acceptanceCriteria).Error
	require.NoError(t, err)
	
	return acceptanceCriteria
}

func createInlineTestRequirement(t *testing.T, db *gorm.DB, userStoryID uuid.UUID, creatorID uuid.UUID) *models.Requirement {
	// First get a requirement type
	var reqType models.RequirementType
	err := db.First(&reqType).Error
	require.NoError(t, err)
	
	description := "This requirement tests inline comment functionality."
	requirement := &models.Requirement{
		ID:          uuid.New(),
		ReferenceID: fmt.Sprintf("REQ-%d", generateSequenceNumber()),
		UserStoryID: userStoryID,
		CreatorID:   creatorID,
		AssigneeID:  creatorID,
		Priority:    models.PriorityMedium,
		Status:      models.RequirementStatusDraft,
		TypeID:      reqType.ID,
		Title:       "Test Requirement",
		Description: &description,
	}
	
	err = db.Create(requirement).Error
	require.NoError(t, err)
	
	return requirement
}

func TestInlineCommentIntegration(t *testing.T) {
	// Setup test environment
	router, db, cleanup := setupInlineCommentIntegrationTest(t)
	defer cleanup()

	// Create test user
	user := createTestUser(t, db)

	// Create test epic with description
	epic := createInlineTestEpic(t, db, user)

	t.Run("CreateInlineComment", func(t *testing.T) {
		// Test creating an inline comment
		req := service.CreateCommentRequest{
			AuthorID:          user.ID,
			Content:           "This is an inline comment",
			LinkedText:        inlineStringPtr("test epic"),
			TextPositionStart: inlineIntPtr(10),
			TextPositionEnd:   inlineIntPtr(19),
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/epics/%s/comments/inline", epic.ID), bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		if w.Code != http.StatusCreated {
			t.Logf("Response body: %s", w.Body.String())
		}
		assert.Equal(t, http.StatusCreated, w.Code)

		var response service.CommentResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, user.ID, response.AuthorID)
		assert.Equal(t, "This is an inline comment", response.Content)
		assert.Equal(t, "test epic", *response.LinkedText)
		assert.Equal(t, 10, *response.TextPositionStart)
		assert.Equal(t, 19, *response.TextPositionEnd)
		assert.True(t, response.IsInline)
	})

	t.Run("CreateInlineCommentWithInvalidTextFragment", func(t *testing.T) {
		// Test creating an inline comment with text that doesn't match
		req := service.CreateCommentRequest{
			AuthorID:          user.ID,
			Content:           "This is an inline comment",
			LinkedText:        inlineStringPtr("wrong text"),
			TextPositionStart: inlineIntPtr(10),
			TextPositionEnd:   inlineIntPtr(20),
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/epics/%s/comments/inline", epic.ID), bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response["details"].(string), "text fragment validation failed")
	})

	t.Run("CreateInlineCommentWithInvalidPositions", func(t *testing.T) {
		// Test creating an inline comment with invalid positions
		req := service.CreateCommentRequest{
			AuthorID:          user.ID,
			Content:           "This is an inline comment",
			LinkedText:        inlineStringPtr("test"),
			TextPositionStart: inlineIntPtr(100), // Beyond text length
			TextPositionEnd:   inlineIntPtr(104),
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/epics/%s/comments/inline", epic.ID), bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GetVisibleInlineComments", func(t *testing.T) {
		// First create a valid inline comment
		// Epic description: "This is a test epic description for inline comments."
		// "description" is at positions 20-30
		req := service.CreateCommentRequest{
			AuthorID:          user.ID,
			Content:           "Visible inline comment",
			LinkedText:        inlineStringPtr("description"),
			TextPositionStart: inlineIntPtr(20),
			TextPositionEnd:   inlineIntPtr(31),
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/epics/%s/comments/inline", epic.ID), bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, httpReq)
		if w.Code != http.StatusCreated {
			t.Logf("Create comment response body: %s", w.Body.String())
		}
		require.Equal(t, http.StatusCreated, w.Code)

		// Now get visible inline comments
		w = httptest.NewRecorder()
		httpReq, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/epics/%s/comments/inline/visible", epic.ID), nil)
		router.ServeHTTP(w, httpReq)

		if w.Code != http.StatusOK {
			t.Logf("Response body: %s", w.Body.String())
		}
		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Comments []service.CommentResponse `json:"comments"`
			Count    int                       `json:"count"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Greater(t, response.Count, 0)
		assert.True(t, response.Comments[0].IsInline)
	})

	t.Run("ValidateInlineCommentsAfterTextChange", func(t *testing.T) {
		// First create an inline comment
		req := service.CreateCommentRequest{
			AuthorID:          user.ID,
			Content:           "Comment to be hidden",
			LinkedText:        inlineStringPtr("inline comments"),
			TextPositionStart: inlineIntPtr(36),
			TextPositionEnd:   inlineIntPtr(51),
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/epics/%s/comments/inline", epic.ID), bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, httpReq)
		if w.Code != http.StatusCreated {
			t.Logf("Create comment response body: %s", w.Body.String())
		}
		require.Equal(t, http.StatusCreated, w.Code)

		// Get the comment ID from response
		var createResponse service.CommentResponse
		err := json.Unmarshal(w.Body.Bytes(), &createResponse)
		require.NoError(t, err)

		// Now validate with changed text
		validateReq := struct {
			NewDescription string `json:"new_description"`
		}{
			NewDescription: "This is a completely different description without the original text.",
		}

		body, _ = json.Marshal(validateReq)
		w = httptest.NewRecorder()
		httpReq, _ = http.NewRequest("POST", fmt.Sprintf("/api/v1/epics/%s/comments/inline/validate", epic.ID), bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify the comment is now hidden by checking visible inline comments
		w = httptest.NewRecorder()
		httpReq, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/epics/%s/comments/inline/visible", epic.ID), nil)
		router.ServeHTTP(w, httpReq)

		var visibleResponse struct {
			Comments []service.CommentResponse `json:"comments"`
			Count    int                       `json:"count"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &visibleResponse)
		require.NoError(t, err)

		// The comment should be hidden, so it shouldn't appear in visible comments
		for _, comment := range visibleResponse.Comments {
			assert.NotEqual(t, createResponse.ID, comment.ID, "Hidden comment should not appear in visible comments")
		}
	})

	t.Run("InlineCommentFiltering", func(t *testing.T) {
		// Create both inline and general comments
		
		// Create a general comment
		generalReq := service.CreateCommentRequest{
			AuthorID: user.ID,
			Content:  "This is a general comment",
		}

		body, _ := json.Marshal(generalReq)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/epics/%s/comments", epic.ID), bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, httpReq)
		if w.Code != http.StatusCreated {
			t.Logf("General comment response body: %s", w.Body.String())
		}
		require.Equal(t, http.StatusCreated, w.Code)

		// Create an inline comment
		inlineReq := service.CreateCommentRequest{
			AuthorID:          user.ID,
			Content:           "This is an inline comment",
			LinkedText:        inlineStringPtr("test"),
			TextPositionStart: inlineIntPtr(10),
			TextPositionEnd:   inlineIntPtr(14),
		}

		body, _ = json.Marshal(inlineReq)
		w = httptest.NewRecorder()
		httpReq, _ = http.NewRequest("POST", fmt.Sprintf("/api/v1/epics/%s/comments/inline", epic.ID), bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, httpReq)
		if w.Code != http.StatusCreated {
			t.Logf("Inline comment response body: %s", w.Body.String())
		}
		require.Equal(t, http.StatusCreated, w.Code)

		// Test filtering for inline comments only
		w = httptest.NewRecorder()
		httpReq, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/epics/%s/comments?inline=true", epic.ID), nil)
		router.ServeHTTP(w, httpReq)

		if w.Code != http.StatusOK {
			t.Logf("Get inline comments response body: %s", w.Body.String())
		}
		assert.Equal(t, http.StatusOK, w.Code)

		var inlineResponse struct {
			Comments []service.CommentResponse `json:"comments"`
			Count    int                       `json:"count"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &inlineResponse)
		require.NoError(t, err)

		// All returned comments should be inline comments
		for _, comment := range inlineResponse.Comments {
			assert.True(t, comment.IsInline, "All comments should be inline when filtering for inline=true")
		}

		// Test getting all comments (should include both)
		w = httptest.NewRecorder()
		httpReq, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/epics/%s/comments", epic.ID), nil)
		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var allResponse struct {
			Comments []service.CommentResponse `json:"comments"`
			Count    int                       `json:"count"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &allResponse)
		require.NoError(t, err)

		// Should have both inline and general comments
		assert.Greater(t, allResponse.Count, inlineResponse.Count)
	})

	t.Run("InlineCommentOnDifferentEntityTypes", func(t *testing.T) {
		// Create user story
		userStory := createInlineTestUserStory(t, db, epic.ID, user.ID)

		// Create acceptance criteria
		acceptanceCriteria := createInlineTestAcceptanceCriteria(t, db, userStory.ID, user.ID)

		// Create requirement
		requirement := createInlineTestRequirement(t, db, userStory.ID, user.ID)

		// Test inline comments on each entity type
		entities := []struct {
			entityType string
			entityID   uuid.UUID
			text       string
			start      int
			end        int
		}{
			{"user-stories", userStory.ID, "test", 21, 25},
			{"acceptance-criteria", acceptanceCriteria.ID, "user", 5, 9},
			{"requirements", requirement.ID, "requirement", 5, 16},
		}

		for _, entity := range entities {
			t.Run(fmt.Sprintf("InlineComment_%s", entity.entityType), func(t *testing.T) {
				req := service.CreateCommentRequest{
					AuthorID:          user.ID,
					Content:           fmt.Sprintf("Inline comment on %s", entity.entityType),
					LinkedText:        inlineStringPtr(entity.text),
					TextPositionStart: inlineIntPtr(entity.start),
					TextPositionEnd:   inlineIntPtr(entity.end),
				}

				body, _ := json.Marshal(req)
				w := httptest.NewRecorder()
				httpReq, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/%s/%s/comments/inline", entity.entityType, entity.entityID), bytes.NewBuffer(body))
				httpReq.Header.Set("Content-Type", "application/json")

				router.ServeHTTP(w, httpReq)

				if w.Code != http.StatusCreated {
					t.Logf("Entity %s response body: %s", entity.entityType, w.Body.String())
				}
				assert.Equal(t, http.StatusCreated, w.Code)

				var response service.CommentResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.True(t, response.IsInline)
				if response.LinkedText != nil {
					assert.Equal(t, entity.text, *response.LinkedText)
				}
			})
		}
	})

	t.Run("InlineCommentValidationEdgeCases", func(t *testing.T) {
		testCases := []struct {
			name          string
			linkedText    *string
			start         *int
			end           *int
			expectedError bool
		}{
			{
				name:          "MissingLinkedText",
				linkedText:    nil,
				start:         inlineIntPtr(0),
				end:           inlineIntPtr(5),
				expectedError: true,
			},
			{
				name:          "MissingStart",
				linkedText:    inlineStringPtr("test"),
				start:         nil,
				end:           inlineIntPtr(5),
				expectedError: true,
			},
			{
				name:          "MissingEnd",
				linkedText:    inlineStringPtr("test"),
				start:         inlineIntPtr(0),
				end:           nil,
				expectedError: true,
			},
			{
				name:          "EmptyLinkedText",
				linkedText:    inlineStringPtr(""),
				start:         inlineIntPtr(0),
				end:           inlineIntPtr(5),
				expectedError: true,
			},
			{
				name:          "NegativeStart",
				linkedText:    inlineStringPtr("test"),
				start:         inlineIntPtr(-1),
				end:           inlineIntPtr(5),
				expectedError: true,
			},
			{
				name:          "StartGreaterThanEnd",
				linkedText:    inlineStringPtr("test"),
				start:         inlineIntPtr(10),
				end:           inlineIntPtr(5),
				expectedError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := service.CreateCommentRequest{
					AuthorID:          user.ID,
					Content:           "Test comment",
					LinkedText:        tc.linkedText,
					TextPositionStart: tc.start,
					TextPositionEnd:   tc.end,
				}

				body, _ := json.Marshal(req)
				w := httptest.NewRecorder()
				httpReq, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/epics/%s/comments/inline", epic.ID), bytes.NewBuffer(body))
				httpReq.Header.Set("Content-Type", "application/json")

				router.ServeHTTP(w, httpReq)

				if tc.expectedError {
					assert.NotEqual(t, http.StatusCreated, w.Code)
				} else {
					assert.Equal(t, http.StatusCreated, w.Code)
				}
			})
		}
	})
}

// Helper functions
func inlineStringPtr(s string) *string {
	return &s
}

func inlineIntPtr(i int) *int {
	return &i
}

