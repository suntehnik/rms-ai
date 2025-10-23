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
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"product-requirements-management/internal/handlers"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/service"
)

type EpicIntegrationTestSuite struct {
	suite.Suite
	db           *gorm.DB
	testDatabase *TestDatabase
	router       *gin.Engine
	epicHandler  *handlers.EpicHandler
	epicService  service.EpicService
	epicRepo     repository.EpicRepository
	userRepo     repository.UserRepository
	testUser     *models.User
	authContext  *TestAuthContext
}

func (suite *EpicIntegrationTestSuite) SetupSuite() {
	// Setup test database with SQL migrations
	suite.testDatabase = SetupTestDatabase(suite.T())
	suite.db = suite.testDatabase.DB

	// Setup repositories
	suite.userRepo = repository.NewUserRepository(suite.db)
	suite.epicRepo = repository.NewEpicRepository(suite.db)

	// Setup services
	suite.epicService = service.NewEpicService(suite.epicRepo, suite.userRepo)

	// Setup handlers
	suite.epicHandler = handlers.NewEpicHandler(suite.epicService)

	// Setup authentication
	suite.authContext = SetupTestAuth(suite.T(), suite.db)

	// Setup router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	v1 := suite.router.Group("/api/v1")
	// Apply authentication middleware to all routes
	v1.Use(suite.authContext.AuthService.Middleware())
	{
		epics := v1.Group("/epics")
		{
			epics.POST("", suite.epicHandler.CreateEpic)
			epics.GET("", suite.epicHandler.ListEpics)
			epics.GET("/:id", suite.epicHandler.GetEpic)
			epics.PUT("/:id", suite.epicHandler.UpdateEpic)
			epics.DELETE("/:id", suite.epicHandler.DeleteEpic)
			epics.GET("/:id/user-stories", suite.epicHandler.GetEpicWithUserStories)
			epics.PATCH("/:id/status", suite.epicHandler.ChangeEpicStatus)
			epics.PATCH("/:id/assign", suite.epicHandler.AssignEpic)
		}
	}
}

func (suite *EpicIntegrationTestSuite) SetupTest() {
	// Clean up database before each test
	suite.db.Exec("DELETE FROM epics")
	suite.db.Exec("DELETE FROM users WHERE username NOT IN ('testuser', 'adminuser')")

	// Use the authenticated test user
	suite.testUser = suite.authContext.TestUser
}

func (suite *EpicIntegrationTestSuite) TearDownSuite() {
	if suite.testDatabase != nil {
		suite.testDatabase.Cleanup(suite.T())
	}
}

func (suite *EpicIntegrationTestSuite) TestCreateEpic() {
	// Test data
	createReq := service.CreateEpicRequest{
		CreatorID:   suite.testUser.ID,
		Priority:    models.PriorityHigh,
		Title:       "Integration Test Epic",
		Description: stringPtr("This is a test epic for integration testing"),
	}

	// Create epic
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/epics", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createdEpic models.Epic
	err := json.Unmarshal(w.Body.Bytes(), &createdEpic)
	suite.Require().NoError(err)

	assert.NotEqual(suite.T(), uuid.Nil, createdEpic.ID)
	assert.Equal(suite.T(), createReq.Title, createdEpic.Title)
	assert.Equal(suite.T(), createReq.Priority, createdEpic.Priority)
	assert.Equal(suite.T(), models.EpicStatusBacklog, createdEpic.Status)
	assert.NotEmpty(suite.T(), createdEpic.ReferenceID)
}

func (suite *EpicIntegrationTestSuite) TestGetEpic() {
	// Create epic first
	epic := &models.Epic{
		ID:          uuid.New(),
		ReferenceID: "EP-001",
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityHigh,
		Status:      models.EpicStatusBacklog,
		Title:       "Test Epic",
		Description: stringPtr("Test epic description"),
	}

	err := suite.epicRepo.Create(epic)
	suite.Require().NoError(err)

	// Get epic by ID
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/epics/%s", epic.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var retrievedEpic models.Epic
	err = json.Unmarshal(w.Body.Bytes(), &retrievedEpic)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), epic.ID, retrievedEpic.ID)
	assert.Equal(suite.T(), epic.Title, retrievedEpic.Title)
	assert.Equal(suite.T(), epic.ReferenceID, retrievedEpic.ReferenceID)
}

func (suite *EpicIntegrationTestSuite) TestGetEpicByReferenceID() {
	// Create epic first
	epic := &models.Epic{
		ID:          uuid.New(),
		ReferenceID: "EP-001",
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityHigh,
		Status:      models.EpicStatusBacklog,
		Title:       "Test Epic",
		Description: stringPtr("Test epic description"),
	}

	err := suite.epicRepo.Create(epic)
	suite.Require().NoError(err)

	// Get epic by reference ID
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/epics/%s", epic.ReferenceID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var retrievedEpic models.Epic
	err = json.Unmarshal(w.Body.Bytes(), &retrievedEpic)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), epic.ID, retrievedEpic.ID)
	assert.Equal(suite.T(), epic.ReferenceID, retrievedEpic.ReferenceID)
}

func (suite *EpicIntegrationTestSuite) TestUpdateEpic() {
	// Create epic first
	epic := &models.Epic{
		ID:          uuid.New(),
		ReferenceID: "EP-001",
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityMedium,
		Status:      models.EpicStatusBacklog,
		Title:       "Original Title",
		Description: stringPtr("Original description"),
	}

	err := suite.epicRepo.Create(epic)
	suite.Require().NoError(err)

	// Update epic
	newTitle := "Updated Title"
	newPriority := models.PriorityHigh
	updateReq := service.UpdateEpicRequest{
		Priority: &newPriority,
		Title:    &newTitle,
	}

	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/epics/%s", epic.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var updatedEpic models.Epic
	err = json.Unmarshal(w.Body.Bytes(), &updatedEpic)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), epic.ID, updatedEpic.ID)
	assert.Equal(suite.T(), "Updated Title", updatedEpic.Title)
	assert.Equal(suite.T(), models.PriorityHigh, updatedEpic.Priority)
}

func (suite *EpicIntegrationTestSuite) TestDeleteEpic() {
	// Create epic first
	epic := &models.Epic{
		ID:          uuid.New(),
		ReferenceID: "EP-001",
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityMedium,
		Status:      models.EpicStatusBacklog,
		Title:       "Epic to Delete",
	}

	err := suite.epicRepo.Create(epic)
	suite.Require().NoError(err)

	// Delete epic
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/epics/%s", epic.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)

	// Verify epic is deleted
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/epics/%s", epic.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *EpicIntegrationTestSuite) TestListEpics() {
	// Create multiple epics
	epics := []*models.Epic{
		{
			ID:          uuid.New(),
			ReferenceID: "EP-001",
			CreatorID:   suite.testUser.ID,
			AssigneeID:  suite.testUser.ID,
			Priority:    models.PriorityHigh,
			Status:      models.EpicStatusBacklog,
			Title:       "Epic 1",
		},
		{
			ID:          uuid.New(),
			ReferenceID: "EP-002",
			CreatorID:   suite.testUser.ID,
			AssigneeID:  suite.testUser.ID,
			Priority:    models.PriorityMedium,
			Status:      models.EpicStatusInProgress,
			Title:       "Epic 2",
		},
	}

	for _, epic := range epics {
		err := suite.epicRepo.Create(epic)
		suite.Require().NoError(err)
	}

	// List all epics
	req, _ := http.NewRequest("GET", "/api/v1/epics", nil)
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response struct {
		Data       []models.Epic `json:"data"`
		TotalCount int64         `json:"total_count"`
		Limit      int           `json:"limit"`
		Offset     int           `json:"offset"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), int64(2), response.TotalCount)
	assert.Len(suite.T(), response.Data, 2)
}

func (suite *EpicIntegrationTestSuite) TestListEpicsWithStatusFilter() {
	// Create multiple epics
	epics := []*models.Epic{
		{
			ID:          uuid.New(),
			ReferenceID: "EP-001",
			CreatorID:   suite.testUser.ID,
			AssigneeID:  suite.testUser.ID,
			Priority:    models.PriorityHigh,
			Status:      models.EpicStatusBacklog,
			Title:       "Epic 1",
		},
		{
			ID:          uuid.New(),
			ReferenceID: "EP-002",
			CreatorID:   suite.testUser.ID,
			AssigneeID:  suite.testUser.ID,
			Priority:    models.PriorityMedium,
			Status:      models.EpicStatusInProgress,
			Title:       "Epic 2",
		},
	}

	for _, epic := range epics {
		err := suite.epicRepo.Create(epic)
		suite.Require().NoError(err)
	}

	// List epics with status filter
	req, _ := http.NewRequest("GET", "/api/v1/epics?status=Backlog", nil)
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response struct {
		Data       []models.Epic `json:"data"`
		TotalCount int64         `json:"total_count"`
		Limit      int           `json:"limit"`
		Offset     int           `json:"offset"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), int64(1), response.TotalCount)
	assert.Len(suite.T(), response.Data, 1)
	assert.Equal(suite.T(), models.EpicStatusBacklog, response.Data[0].Status)
}

func (suite *EpicIntegrationTestSuite) TestChangeEpicStatus() {
	// Create epic first
	epic := &models.Epic{
		ID:          uuid.New(),
		ReferenceID: "EP-001",
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityMedium,
		Status:      models.EpicStatusBacklog,
		Title:       "Epic for Status Change",
	}

	err := suite.epicRepo.Create(epic)
	suite.Require().NoError(err)

	// Change status
	statusReq := map[string]string{
		"status": "In Progress",
	}

	body, _ := json.Marshal(statusReq)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/epics/%s/status", epic.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var updatedEpic models.Epic
	err = json.Unmarshal(w.Body.Bytes(), &updatedEpic)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), epic.ID, updatedEpic.ID)
	assert.Equal(suite.T(), models.EpicStatusInProgress, updatedEpic.Status)
}

func (suite *EpicIntegrationTestSuite) TestAssignEpic() {
	// Create epic first
	epic := &models.Epic{
		ID:          uuid.New(),
		ReferenceID: "EP-001",
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityMedium,
		Status:      models.EpicStatusBacklog,
		Title:       "Epic for Assignment",
	}

	err := suite.epicRepo.Create(epic)
	suite.Require().NoError(err)

	// Assign to admin user
	assignReq := map[string]string{
		"assignee_id": suite.authContext.AdminUser.ID.String(),
	}

	body, _ := json.Marshal(assignReq)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/epics/%s/assign", epic.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var updatedEpic models.Epic
	err = json.Unmarshal(w.Body.Bytes(), &updatedEpic)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), epic.ID, updatedEpic.ID)
	assert.Equal(suite.T(), suite.authContext.AdminUser.ID, updatedEpic.AssigneeID)
}

func (suite *EpicIntegrationTestSuite) TestGetEpicWithUserStories() {
	// Create epic first
	epic := &models.Epic{
		ID:          uuid.New(),
		ReferenceID: "EP-001",
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityMedium,
		Status:      models.EpicStatusBacklog,
		Title:       "Epic with User Stories",
	}

	err := suite.epicRepo.Create(epic)
	suite.Require().NoError(err)

	// Get epic with user stories (should be empty initially)
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/epics/%s/user-stories", epic.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.Epic
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), epic.ID, response.ID)
	assert.Equal(suite.T(), epic.Title, response.Title)
	// UserStories should be empty initially
	if response.UserStories != nil {
		assert.Len(suite.T(), response.UserStories, 0)
	}
}

func (suite *EpicIntegrationTestSuite) TestUnauthorizedAccess() {
	req, _ := http.NewRequest("GET", "/api/v1/epics", nil)
	// Intentionally not setting Authorization header
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *EpicIntegrationTestSuite) TestInvalidToken() {
	req, _ := http.NewRequest("GET", "/api/v1/epics", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *EpicIntegrationTestSuite) TestCreateEpicWithInvalidData() {
	// Test with missing required fields
	createReq := service.CreateEpicRequest{
		// Missing CreatorID, Title
		Priority: models.PriorityHigh,
	}

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/epics", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *EpicIntegrationTestSuite) TestGetNonExistentEpic() {
	nonExistentID := uuid.New()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/epics/%s", nonExistentID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *EpicIntegrationTestSuite) TestUpdateNonExistentEpic() {
	nonExistentID := uuid.New()
	newTitle := "Updated Title"
	updateReq := service.UpdateEpicRequest{
		Title: &newTitle,
	}

	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/epics/%s", nonExistentID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *EpicIntegrationTestSuite) TestDeleteNonExistentEpic() {
	nonExistentID := uuid.New()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/epics/%s", nonExistentID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func TestEpicIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(EpicIntegrationTestSuite))
}
