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
	"gorm.io/gorm"

	"product-requirements-management/internal/handlers"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/service"
)

// setupAcceptanceCriteriaTestServer creates a test server with in-memory database
func setupAcceptanceCriteriaTestServer(t *testing.T) (*gin.Engine, *gorm.DB, func()) {
	// Create in-memory SQLite database
	testDatabase := SetupTestDatabase(t)
	db := testDatabase.DB

	// Auto-migrate models
	err := models.AutoMigrate(db)
	require.NoError(t, err)

	// Seed default data
	err = models.SeedDefaultData(db)
	require.NoError(t, err)

	// Create repositories
	epicRepo := repository.NewEpicRepository(db)
	userRepo := repository.NewUserRepository(db)
	userStoryRepo := repository.NewUserStoryRepository(db)
	acceptanceCriteriaRepo := repository.NewAcceptanceCriteriaRepository(db)

	// Create services
	epicService := service.NewEpicService(epicRepo, userRepo)
	userStoryService := service.NewUserStoryService(userStoryRepo, epicRepo, userRepo)
	acceptanceCriteriaService := service.NewAcceptanceCriteriaService(acceptanceCriteriaRepo, userStoryRepo, userRepo)

	// Create handlers
	epicHandler := handlers.NewEpicHandler(epicService)
	userStoryHandler := handlers.NewUserStoryHandler(userStoryService)
	acceptanceCriteriaHandler := handlers.NewAcceptanceCriteriaHandler(acceptanceCriteriaService)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/api/v1")
	{
		// Epic routes
		epics := v1.Group("/epics")
		{
			epics.POST("", epicHandler.CreateEpic)
			epics.GET("/:id", epicHandler.GetEpic)
			epics.POST("/:id/user-stories", userStoryHandler.CreateUserStoryInEpic)
		}

		// User Story routes
		userStories := v1.Group("/user-stories")
		{
			userStories.POST("", userStoryHandler.CreateUserStory)
			userStories.GET("/:id", userStoryHandler.GetUserStory)
			userStories.POST("/:id/acceptance-criteria", acceptanceCriteriaHandler.CreateAcceptanceCriteria)
			userStories.GET("/:id/acceptance-criteria", acceptanceCriteriaHandler.GetAcceptanceCriteriaByUserStory)
		}

		// Acceptance Criteria routes
		acceptanceCriteria := v1.Group("/acceptance-criteria")
		{
			acceptanceCriteria.GET("", acceptanceCriteriaHandler.ListAcceptanceCriteria)
			acceptanceCriteria.GET("/:id", acceptanceCriteriaHandler.GetAcceptanceCriteria)
			acceptanceCriteria.PUT("/:id", acceptanceCriteriaHandler.UpdateAcceptanceCriteria)
			acceptanceCriteria.DELETE("/:id", acceptanceCriteriaHandler.DeleteAcceptanceCriteria)
		}
	}

	cleanup := func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}

	return router, db, cleanup
}

// createAcceptanceCriteriaTestUser creates a test user in the database
func createAcceptanceCriteriaTestUser(t *testing.T, db *gorm.DB) *models.User {
	user := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         models.RoleUser,
	}

	err := db.Create(user).Error
	require.NoError(t, err)

	return user
}

// createAcceptanceCriteriaTestEpic creates a test epic in the database
func createAcceptanceCriteriaTestEpic(t *testing.T, db *gorm.DB, creatorID uuid.UUID) *models.Epic {
	epic := &models.Epic{
		ID:          uuid.New(),
		CreatorID:   creatorID,
		AssigneeID:  creatorID,
		Priority:    models.PriorityHigh,
		Status:      models.EpicStatusBacklog,
		Title:       "Test Epic",
		Description: stringPtr("Epic description"),
	}

	err := db.Create(epic).Error
	require.NoError(t, err)

	return epic
}

// createAcceptanceCriteriaTestUserStory creates a test user story in the database
func createAcceptanceCriteriaTestUserStory(t *testing.T, db *gorm.DB, epicID, creatorID uuid.UUID) *models.UserStory {
	description := "As a user, I want to test, so that I can verify"
	userStory := &models.UserStory{
		ID:          uuid.New(),
		EpicID:      epicID,
		CreatorID:   creatorID,
		AssigneeID:  creatorID,
		Priority:    models.PriorityHigh,
		Status:      models.UserStoryStatusBacklog,
		Title:       "Test User Story",
		Description: &description,
	}

	err := db.Create(userStory).Error
	require.NoError(t, err)

	return userStory
}

func TestAcceptanceCriteriaIntegration(t *testing.T) {
	// Setup test environment
	router, db, cleanup := setupAcceptanceCriteriaTestServer(t)
	defer cleanup()

	// Create test user
	user := createAcceptanceCriteriaTestUser(t, db)

	// Create test epic
	epic := createAcceptanceCriteriaTestEpic(t, db, user.ID)

	// Create test user story
	userStory := createAcceptanceCriteriaTestUserStory(t, db, epic.ID, user.ID)

	t.Run("Create Acceptance Criteria", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"author_id":   user.ID.String(),
			"description": "WHEN user clicks submit THEN system SHALL validate the form",
		}

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/user-stories/%s/acceptance-criteria", userStory.ID.String()), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.AcceptanceCriteria
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEqual(t, uuid.Nil, response.ID)
		assert.NotEmpty(t, response.ReferenceID)
		assert.Equal(t, userStory.ID, response.UserStoryID)
		assert.Equal(t, user.ID, response.AuthorID)
		assert.Equal(t, "WHEN user clicks submit THEN system SHALL validate the form", response.Description)
		assert.False(t, response.CreatedAt.IsZero())
		assert.False(t, response.LastModified.IsZero())
	})

	// Create acceptance criteria for further tests
	acceptanceCriteria := createTestAcceptanceCriteria(t, db, userStory.ID, user.ID, "WHEN user submits form THEN system SHALL validate all required fields")

	t.Run("Get Acceptance Criteria by ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/acceptance-criteria/%s", acceptanceCriteria.ID.String()), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.AcceptanceCriteria
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, acceptanceCriteria.ID, response.ID)
		assert.Equal(t, acceptanceCriteria.ReferenceID, response.ReferenceID)
		assert.Equal(t, acceptanceCriteria.Description, response.Description)
	})

	t.Run("Get Acceptance Criteria by Reference ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/acceptance-criteria/%s", acceptanceCriteria.ReferenceID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.AcceptanceCriteria
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, acceptanceCriteria.ID, response.ID)
		assert.Equal(t, acceptanceCriteria.ReferenceID, response.ReferenceID)
	})

	t.Run("Update Acceptance Criteria", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"description": "WHEN user submits form THEN system SHALL validate all required fields - updated",
		}

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/acceptance-criteria/%s", acceptanceCriteria.ID.String()), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.AcceptanceCriteria
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, acceptanceCriteria.ID, response.ID)
		assert.Equal(t, "WHEN user submits form THEN system SHALL validate all required fields - updated", response.Description)
		assert.True(t, response.LastModified.After(response.CreatedAt))
	})

	t.Run("List Acceptance Criteria", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/acceptance-criteria", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Should have 2 acceptance criteria (one from Create test, one from createTestAcceptanceCriteria)
		assert.Equal(t, float64(2), response["count"])

		criteria := response["acceptance_criteria"].([]interface{})
		assert.Len(t, criteria, 2)
	})

	t.Run("List Acceptance Criteria with User Story Filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/acceptance-criteria?user_story_id=%s", userStory.ID.String()), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(2), response["count"])

		criteria := response["acceptance_criteria"].([]interface{})
		assert.Len(t, criteria, 2)
	})

	t.Run("Get Acceptance Criteria by User Story", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/user-stories/%s/acceptance-criteria", userStory.ID.String()), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(2), response["count"])

		criteria := response["acceptance_criteria"].([]interface{})
		assert.Len(t, criteria, 2)
	})

	t.Run("Delete Acceptance Criteria - Prevent Last One", func(t *testing.T) {
		// First, delete one acceptance criteria to leave only one
		var allCriteria []models.AcceptanceCriteria
		err := db.Where("user_story_id = ?", userStory.ID).Find(&allCriteria).Error
		require.NoError(t, err)
		require.Len(t, allCriteria, 2)

		// Delete the first one
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/acceptance-criteria/%s", allCriteria[0].ID.String()), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Now try to delete the last one - should fail
		req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/acceptance-criteria/%s", allCriteria[1].ID.String()), nil)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "must have at least one acceptance criteria")
	})

	t.Run("Force Delete Last Acceptance Criteria", func(t *testing.T) {
		// Get the remaining acceptance criteria
		var remainingCriteria models.AcceptanceCriteria
		err := db.Where("user_story_id = ?", userStory.ID).First(&remainingCriteria).Error
		require.NoError(t, err)

		// Force delete the last one
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/acceptance-criteria/%s?force=true", remainingCriteria.ID.String()), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify it's deleted
		var count int64
		err = db.Model(&models.AcceptanceCriteria{}).Where("user_story_id = ?", userStory.ID).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("Error Cases", func(t *testing.T) {
		t.Run("Create with Invalid User Story ID", func(t *testing.T) {
			requestBody := map[string]interface{}{
				"author_id":   user.ID.String(),
				"description": "WHEN user clicks submit THEN system SHALL validate the form",
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/user-stories/invalid-uuid/acceptance-criteria", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("Create with Non-existent User Story", func(t *testing.T) {
			nonExistentID := uuid.New()
			requestBody := map[string]interface{}{
				"author_id":   user.ID.String(),
				"description": "WHEN user clicks submit THEN system SHALL validate the form",
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/user-stories/%s/acceptance-criteria", nonExistentID.String()), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Contains(t, response["error"], "User story not found")
		})

		t.Run("Get Non-existent Acceptance Criteria", func(t *testing.T) {
			nonExistentID := uuid.New()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/acceptance-criteria/%s", nonExistentID.String()), nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Contains(t, response["error"], "Acceptance criteria not found")
		})

		t.Run("Update Non-existent Acceptance Criteria", func(t *testing.T) {
			nonExistentID := uuid.New()
			requestBody := map[string]interface{}{
				"description": "Updated description",
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/acceptance-criteria/%s", nonExistentID.String()), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Contains(t, response["error"], "Acceptance criteria not found")
		})

		t.Run("Delete Non-existent Acceptance Criteria", func(t *testing.T) {
			nonExistentID := uuid.New()
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/acceptance-criteria/%s", nonExistentID.String()), nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Contains(t, response["error"], "Acceptance criteria not found")
		})
	})
}

// Helper function to create test acceptance criteria
func createTestAcceptanceCriteria(t *testing.T, db *gorm.DB, userStoryID, authorID uuid.UUID, description string) *models.AcceptanceCriteria {
	acceptanceCriteria := &models.AcceptanceCriteria{
		ID:          uuid.New(),
		UserStoryID: userStoryID,
		AuthorID:    authorID,
		Description: description,
	}

	err := db.Create(acceptanceCriteria).Error
	require.NoError(t, err)

	return acceptanceCriteria
}
