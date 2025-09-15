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

// setupTestServer creates a test server with in-memory database
func setupTestServer(t *testing.T) (*gin.Engine, *gorm.DB, func()) {
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

	// Create services
	epicService := service.NewEpicService(epicRepo, userRepo)

	// Create handlers
	epicHandler := handlers.NewEpicHandler(epicService)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup routes
	v1 := router.Group("/api/v1")
	epics := v1.Group("/epics")
	{
		epics.POST("", epicHandler.CreateEpic)
		epics.GET("", epicHandler.ListEpics)
		epics.GET("/:id", epicHandler.GetEpic)
		epics.PUT("/:id", epicHandler.UpdateEpic)
		epics.DELETE("/:id", epicHandler.DeleteEpic)
		epics.GET("/:id/user-stories", epicHandler.GetEpicWithUserStories)
		epics.PATCH("/:id/status", epicHandler.ChangeEpicStatus)
		epics.PATCH("/:id/assign", epicHandler.AssignEpic)
	}

	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return router, db, cleanup
}

func TestEpicIntegration_CreateAndGetEpic(t *testing.T) {
	router, db, cleanup := setupTestServer(t)
	defer cleanup()

	// Create test user
	user := createTestUser(t, db)

	// Test data
	createReq := service.CreateEpicRequest{
		CreatorID:   user.ID,
		Priority:    models.PriorityHigh,
		Title:       "Integration Test Epic",
		Description: stringPtr("This is a test epic for integration testing"),
	}

	// Create epic
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/epics", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createdEpic models.Epic
	err := json.Unmarshal(w.Body.Bytes(), &createdEpic)
	require.NoError(t, err)

	assert.NotEqual(t, uuid.Nil, createdEpic.ID)
	assert.Equal(t, createReq.Title, createdEpic.Title)
	assert.Equal(t, createReq.Priority, createdEpic.Priority)
	assert.Equal(t, models.EpicStatusBacklog, createdEpic.Status)

	// Get epic by ID
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/epics/%s", createdEpic.ID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var retrievedEpic models.Epic
	err = json.Unmarshal(w.Body.Bytes(), &retrievedEpic)
	require.NoError(t, err)

	assert.Equal(t, createdEpic.ID, retrievedEpic.ID)
	assert.Equal(t, createdEpic.Title, retrievedEpic.Title)
}

func TestEpicIntegration_UpdateEpic(t *testing.T) {
	router, db, cleanup := setupTestServer(t)
	defer cleanup()

	// Create test user
	user := createTestUser(t, db)

	// Create epic first
	epic := &models.Epic{
		ID:          uuid.New(),
		ReferenceID: "EP-001",
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityMedium,
		Status:      models.EpicStatusBacklog,
		Title:       "Original Title",
	}

	err := db.Create(epic).Error
	require.NoError(t, err)

	// Update epic
	updateReq := service.UpdateEpicRequest{
		Priority: &[]models.Priority{models.PriorityHigh}[0],
		Title:    &[]string{"Updated Title"}[0],
	}

	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/epics/%s", epic.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedEpic models.Epic
	err = json.Unmarshal(w.Body.Bytes(), &updatedEpic)
	require.NoError(t, err)

	assert.Equal(t, epic.ID, updatedEpic.ID)
	assert.Equal(t, "Updated Title", updatedEpic.Title)
	assert.Equal(t, models.PriorityHigh, updatedEpic.Priority)
}

func TestEpicIntegration_DeleteEpic(t *testing.T) {
	router, db, cleanup := setupTestServer(t)
	defer cleanup()

	// Create test user
	user := createTestUser(t, db)

	// Create epic first
	epic := &models.Epic{
		ID:          uuid.New(),
		ReferenceID: "EP-001",
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityMedium,
		Status:      models.EpicStatusBacklog,
		Title:       "Epic to Delete",
	}

	err := db.Create(epic).Error
	require.NoError(t, err)

	// Delete epic
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/epics/%s", epic.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify epic is deleted
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/epics/%s", epic.ID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestEpicIntegration_ListEpics(t *testing.T) {
	router, db, cleanup := setupTestServer(t)
	defer cleanup()

	// Create test user
	user := createTestUser(t, db)

	// Create multiple epics
	epics := []*models.Epic{
		{
			ID:          uuid.New(),
			ReferenceID: "EP-001",
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityHigh,
			Status:      models.EpicStatusBacklog,
			Title:       "Epic 1",
		},
		{
			ID:          uuid.New(),
			ReferenceID: "EP-002",
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityMedium,
			Status:      models.EpicStatusInProgress,
			Title:       "Epic 2",
		},
	}

	for _, epic := range epics {
		err := db.Create(epic).Error
		require.NoError(t, err)
	}

	// List all epics
	req, _ := http.NewRequest("GET", "/api/v1/epics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Epics []models.Epic `json:"epics"`
		Count int           `json:"count"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, response.Count)
	assert.Len(t, response.Epics, 2)

	// List epics with status filter
	req, _ = http.NewRequest("GET", "/api/v1/epics?status=Backlog", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 1, response.Count)
	assert.Equal(t, models.EpicStatusBacklog, response.Epics[0].Status)
}

func TestEpicIntegration_ChangeEpicStatus(t *testing.T) {
	router, db, cleanup := setupTestServer(t)
	defer cleanup()

	// Create test user
	user := createTestUser(t, db)

	// Create epic first
	epic := &models.Epic{
		ID:          uuid.New(),
		ReferenceID: "EP-001",
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityMedium,
		Status:      models.EpicStatusBacklog,
		Title:       "Epic for Status Change",
	}

	err := db.Create(epic).Error
	require.NoError(t, err)

	// Change status
	statusReq := map[string]string{
		"status": "In Progress",
	}

	body, _ := json.Marshal(statusReq)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/epics/%s/status", epic.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedEpic models.Epic
	err = json.Unmarshal(w.Body.Bytes(), &updatedEpic)
	require.NoError(t, err)

	assert.Equal(t, epic.ID, updatedEpic.ID)
	assert.Equal(t, models.EpicStatusInProgress, updatedEpic.Status)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
