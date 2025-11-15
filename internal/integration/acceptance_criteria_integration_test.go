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

type AcceptanceCriteriaIntegrationTestSuite struct {
	suite.Suite
	db                        *gorm.DB
	testDatabase              *TestDatabase
	router                    *gin.Engine
	acceptanceCriteriaHandler *handlers.AcceptanceCriteriaHandler
	acceptanceCriteriaService service.AcceptanceCriteriaService
	userStoryService          service.UserStoryService
	acceptanceCriteriaRepo    repository.AcceptanceCriteriaRepository
	userStoryRepo             repository.UserStoryRepository
	epicRepo                  repository.EpicRepository
	userRepo                  repository.UserRepository
	testUser                  *models.User
	testEpic                  *models.Epic
	testUserStory             *models.UserStory
	authContext               *TestAuthContext
}

func (suite *AcceptanceCriteriaIntegrationTestSuite) SetupSuite() {
	// Setup test database with SQL migrations
	suite.testDatabase = SetupTestDatabase(suite.T())
	suite.db = suite.testDatabase.DB

	// Setup repositories
	suite.userRepo = repository.NewUserRepository(suite.db)
	suite.epicRepo = repository.NewEpicRepository(suite.db)
	suite.userStoryRepo = repository.NewUserStoryRepository(suite.db, nil)
	suite.acceptanceCriteriaRepo = repository.NewAcceptanceCriteriaRepository(suite.db)

	// Setup services
	suite.acceptanceCriteriaService = service.NewAcceptanceCriteriaService(suite.acceptanceCriteriaRepo, suite.userStoryRepo, suite.userRepo)
	suite.userStoryService = service.NewUserStoryService(suite.userStoryRepo, suite.epicRepo, suite.userRepo)
	// Setup handlers
	suite.acceptanceCriteriaHandler = handlers.NewAcceptanceCriteriaHandler(suite.acceptanceCriteriaService, suite.userStoryService)

	// Setup authentication
	suite.authContext = SetupTestAuth(suite.T(), suite.db)

	// Setup router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	v1 := suite.router.Group("/api/v1")
	// Apply authentication middleware to all routes
	v1.Use(suite.authContext.AuthService.Middleware())
	{
		// User Story routes
		userStories := v1.Group("/user-stories")
		{
			userStories.POST("/:id/acceptance-criteria", suite.acceptanceCriteriaHandler.CreateAcceptanceCriteria)
		}

		// Acceptance Criteria routes
		acceptanceCriteria := v1.Group("/acceptance-criteria")
		{
			acceptanceCriteria.GET("", suite.acceptanceCriteriaHandler.ListAcceptanceCriteria)
			acceptanceCriteria.GET("/:id", suite.acceptanceCriteriaHandler.GetAcceptanceCriteria)
			acceptanceCriteria.PUT("/:id", suite.acceptanceCriteriaHandler.UpdateAcceptanceCriteria)
			acceptanceCriteria.DELETE("/:id", suite.acceptanceCriteriaHandler.DeleteAcceptanceCriteria)
		}
	}
}

func (suite *AcceptanceCriteriaIntegrationTestSuite) SetupTest() {
	// Clean up database before each test
	suite.db.Exec("DELETE FROM acceptance_criteria")
	suite.db.Exec("DELETE FROM user_stories")
	suite.db.Exec("DELETE FROM epics")
	suite.db.Exec("DELETE FROM users WHERE username NOT IN ('testuser', 'adminuser')")

	// Use the authenticated test user
	suite.testUser = suite.authContext.TestUser

	// Create test epic
	suite.testEpic = &models.Epic{
		ID:          uuid.New(),
		ReferenceID: "EP-001",
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityHigh,
		Status:      models.EpicStatusBacklog,
		Title:       "Test Epic",
		Description: stringPtr("Epic description"),
	}
	err := suite.epicRepo.Create(suite.testEpic)
	suite.Require().NoError(err)

	// Create test user story
	description := "As a user, I want to test, so that I can verify"
	suite.testUserStory = &models.UserStory{
		ID:          uuid.New(),
		ReferenceID: "US-001",
		EpicID:      suite.testEpic.ID,
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityHigh,
		Status:      models.UserStoryStatusBacklog,
		Title:       "Test User Story",
		Description: &description,
	}
	err = suite.userStoryRepo.Create(suite.testUserStory)
	suite.Require().NoError(err)
}

func (suite *AcceptanceCriteriaIntegrationTestSuite) TearDownSuite() {
	if suite.testDatabase != nil {
		suite.testDatabase.Cleanup(suite.T())
	}
}

func (suite *AcceptanceCriteriaIntegrationTestSuite) TestCreateAcceptanceCriteria() {
	requestBody := map[string]any{
		"author_id":   suite.testUser.ID.String(),
		"description": "WHEN user clicks submit THEN system SHALL validate the form",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/user-stories/%s/acceptance-criteria", suite.testUserStory.ID.String()), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response models.AcceptanceCriteria
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	assert.NotEqual(suite.T(), uuid.Nil, response.ID)
	assert.NotEmpty(suite.T(), response.ReferenceID)
	assert.Equal(suite.T(), suite.testUserStory.ID, response.UserStoryID)
	assert.Equal(suite.T(), suite.testUser.ID, response.AuthorID)
	assert.Equal(suite.T(), "WHEN user clicks submit THEN system SHALL validate the form", response.Description)
	assert.False(suite.T(), response.CreatedAt.IsZero())
	assert.False(suite.T(), response.UpdatedAt.IsZero())
}

func (suite *AcceptanceCriteriaIntegrationTestSuite) TestGetAcceptanceCriteriaByID() {
	// Create acceptance criteria for the test
	acceptanceCriteria := suite.createTestAcceptanceCriteria("WHEN user submits form THEN system SHALL validate all required fields")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/acceptance-criteria/%s", acceptanceCriteria.ID.String()), nil)
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.AcceptanceCriteria
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), acceptanceCriteria.ID, response.ID)
	assert.Equal(suite.T(), acceptanceCriteria.ReferenceID, response.ReferenceID)
	assert.Equal(suite.T(), acceptanceCriteria.Description, response.Description)
}

func (suite *AcceptanceCriteriaIntegrationTestSuite) TestGetAcceptanceCriteriaByReferenceID() {
	// Create acceptance criteria for the test
	acceptanceCriteria := suite.createTestAcceptanceCriteria("WHEN user submits form THEN system SHALL validate all required fields")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/acceptance-criteria/%s", acceptanceCriteria.ReferenceID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.AcceptanceCriteria
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), acceptanceCriteria.ID, response.ID)
	assert.Equal(suite.T(), acceptanceCriteria.ReferenceID, response.ReferenceID)
}

func (suite *AcceptanceCriteriaIntegrationTestSuite) TestUpdateAcceptanceCriteria() {
	// Create acceptance criteria for the test
	acceptanceCriteria := suite.createTestAcceptanceCriteria("WHEN user submits form THEN system SHALL validate all required fields")

	requestBody := map[string]any{
		"description": "WHEN user submits form THEN system SHALL validate all required fields - updated",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/acceptance-criteria/%s", acceptanceCriteria.ID.String()), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.AcceptanceCriteria
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), acceptanceCriteria.ID, response.ID)
	assert.Equal(suite.T(), "WHEN user submits form THEN system SHALL validate all required fields - updated", response.Description)
	assert.True(suite.T(), response.UpdatedAt.After(response.CreatedAt))
}

func (suite *AcceptanceCriteriaIntegrationTestSuite) TestListAcceptanceCriteria() {
	// Create test acceptance criteria
	suite.createTestAcceptanceCriteria("WHEN user submits form THEN system SHALL validate all required fields")
	suite.createTestAcceptanceCriteria("WHEN user clicks cancel THEN system SHALL discard changes")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/acceptance-criteria", nil)
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), float64(2), response["total_count"])
	assert.Equal(suite.T(), float64(50), response["limit"])
	assert.Equal(suite.T(), float64(0), response["offset"])
	assert.NotNil(suite.T(), response["data"])

	criteria := response["data"].([]any)
	assert.Len(suite.T(), criteria, 2)
}

func (suite *AcceptanceCriteriaIntegrationTestSuite) TestListAcceptanceCriteriaWithUserStoryFilter() {
	// Create test acceptance criteria
	suite.createTestAcceptanceCriteria("WHEN user submits form THEN system SHALL validate all required fields")
	suite.createTestAcceptanceCriteria("WHEN user clicks cancel THEN system SHALL discard changes")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/acceptance-criteria?user_story_id=%s", suite.testUserStory.ID.String()), nil)
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), float64(2), response["total_count"])
	assert.Equal(suite.T(), float64(50), response["limit"])
	assert.Equal(suite.T(), float64(0), response["offset"])
	assert.NotNil(suite.T(), response["data"])

	criteria := response["data"].([]any)
	assert.Len(suite.T(), criteria, 2)
}

func (suite *AcceptanceCriteriaIntegrationTestSuite) TestDeleteAcceptanceCriteriaPreventLastOne() {
	// Create two acceptance criteria
	criteria1 := suite.createTestAcceptanceCriteria("WHEN user submits form THEN system SHALL validate all required fields")
	criteria2 := suite.createTestAcceptanceCriteria("WHEN user clicks cancel THEN system SHALL discard changes")

	// Delete the first one
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/acceptance-criteria/%s", criteria1.ID.String()), nil)
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)

	// Now try to delete the last one - should fail
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/acceptance-criteria/%s", criteria2.ID.String()), nil)
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusConflict, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)
	assert.Contains(suite.T(), response["error"], "must have at least one acceptance criteria")
}

func (suite *AcceptanceCriteriaIntegrationTestSuite) TestForceDeleteLastAcceptanceCriteria() {
	// Create one acceptance criteria
	criteria := suite.createTestAcceptanceCriteria("WHEN user submits form THEN system SHALL validate all required fields")

	// Force delete the last one
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/acceptance-criteria/%s?force=true", criteria.ID.String()), nil)
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)

	// Verify it's deleted
	var count int64
	err := suite.db.Model(&models.AcceptanceCriteria{}).Where("user_story_id = ?", suite.testUserStory.ID).Count(&count).Error
	suite.Require().NoError(err)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *AcceptanceCriteriaIntegrationTestSuite) TestCreateWithInvalidUserStoryID() {
	requestBody := map[string]any{
		"author_id":   suite.testUser.ID.String(),
		"description": "WHEN user clicks submit THEN system SHALL validate the form",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user-stories/invalid-uuid/acceptance-criteria", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *AcceptanceCriteriaIntegrationTestSuite) TestCreateWithNonExistentUserStory() {
	nonExistentID := uuid.New()
	requestBody := map[string]any{
		"author_id":   suite.testUser.ID.String(),
		"description": "WHEN user clicks submit THEN system SHALL validate the form",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/user-stories/%s/acceptance-criteria", nonExistentID.String()), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)
	assert.Contains(suite.T(), response["error"], "User story not found")
}

func (suite *AcceptanceCriteriaIntegrationTestSuite) TestGetNonExistentAcceptanceCriteria() {
	nonExistentID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/acceptance-criteria/%s", nonExistentID.String()), nil)
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)
	assert.Contains(suite.T(), response["error"], "Acceptance criteria not found")
}

func (suite *AcceptanceCriteriaIntegrationTestSuite) TestUpdateNonExistentAcceptanceCriteria() {
	nonExistentID := uuid.New()
	requestBody := map[string]any{
		"description": "Updated description",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/acceptance-criteria/%s", nonExistentID.String()), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)
	assert.Contains(suite.T(), response["error"], "Acceptance criteria not found")
}

func (suite *AcceptanceCriteriaIntegrationTestSuite) TestDeleteNonExistentAcceptanceCriteria() {
	nonExistentID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/acceptance-criteria/%s", nonExistentID.String()), nil)
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)
	assert.Contains(suite.T(), response["error"], "Acceptance criteria not found")
}

// Helper function to create test acceptance criteria
func (suite *AcceptanceCriteriaIntegrationTestSuite) createTestAcceptanceCriteria(description string) *models.AcceptanceCriteria {
	acceptanceCriteria := &models.AcceptanceCriteria{
		ID:          uuid.New(),
		UserStoryID: suite.testUserStory.ID,
		AuthorID:    suite.testUser.ID,
		Description: description,
	}

	err := suite.acceptanceCriteriaRepo.Create(acceptanceCriteria)
	suite.Require().NoError(err)

	return acceptanceCriteria
}

func (suite *AcceptanceCriteriaIntegrationTestSuite) TestUnauthorizedAccess() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/acceptance-criteria", nil)
	// Intentionally not setting Authorization header
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *AcceptanceCriteriaIntegrationTestSuite) TestInvalidToken() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/acceptance-criteria", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// TestAcceptanceCriteriaIntegration runs the acceptance criteria integration test suite
func TestAcceptanceCriteriaIntegration(t *testing.T) {
	suite.Run(t, new(AcceptanceCriteriaIntegrationTestSuite))
}
