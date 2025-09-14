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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"product-requirements-management/internal/handlers"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/service"
)

type UserStoryIntegrationTestSuite struct {
	suite.Suite
	db               *gorm.DB
	router           *gin.Engine
	userStoryHandler *handlers.UserStoryHandler
	userStoryService service.UserStoryService
	userStoryRepo    repository.UserStoryRepository
	epicRepo         repository.EpicRepository
	userRepo         repository.UserRepository
	testUser         *models.User
	testEpic         *models.Epic
}

func (suite *UserStoryIntegrationTestSuite) SetupSuite() {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.Epic{},
		&models.UserStory{},
		&models.AcceptanceCriteria{},
		&models.Requirement{},
		&models.Comment{},
	)
	suite.Require().NoError(err)

	// Seed default data
	err = models.SeedDefaultData(db)
	suite.Require().NoError(err)

	suite.db = db

	// Setup repositories
	suite.userRepo = repository.NewUserRepository(db)
	suite.epicRepo = repository.NewEpicRepository(db)
	suite.userStoryRepo = repository.NewUserStoryRepository(db)

	// Setup services
	suite.userStoryService = service.NewUserStoryService(suite.userStoryRepo, suite.epicRepo, suite.userRepo)

	// Setup handlers
	suite.userStoryHandler = handlers.NewUserStoryHandler(suite.userStoryService)

	// Setup router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	v1 := suite.router.Group("/api/v1")
	{
		v1.POST("/user-stories", suite.userStoryHandler.CreateUserStory)
		v1.POST("/epics/:id/user-stories", suite.userStoryHandler.CreateUserStoryInEpic)
		v1.GET("/user-stories/:id", suite.userStoryHandler.GetUserStory)
		v1.PUT("/user-stories/:id", suite.userStoryHandler.UpdateUserStory)
		v1.DELETE("/user-stories/:id", suite.userStoryHandler.DeleteUserStory)
		v1.GET("/user-stories", suite.userStoryHandler.ListUserStories)
		v1.GET("/user-stories/:id/acceptance-criteria", suite.userStoryHandler.GetUserStoryWithAcceptanceCriteria)
		v1.GET("/user-stories/:id/requirements", suite.userStoryHandler.GetUserStoryWithRequirements)
		v1.PATCH("/user-stories/:id/status", suite.userStoryHandler.ChangeUserStoryStatus)
		v1.PATCH("/user-stories/:id/assign", suite.userStoryHandler.AssignUserStory)
	}
}

func (suite *UserStoryIntegrationTestSuite) SetupTest() {
	// Clean up database before each test
	suite.db.Exec("DELETE FROM user_stories")
	suite.db.Exec("DELETE FROM epics")
	suite.db.Exec("DELETE FROM users")

	// Create test user
	suite.testUser = &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}
	err := suite.userRepo.Create(suite.testUser)
	suite.Require().NoError(err)

	// Create test epic
	suite.testEpic = &models.Epic{
		ID:          uuid.New(),
		ReferenceID: "EP-001",
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityHigh,
		Status:      models.EpicStatusBacklog,
		Title:       "Test Epic",
		Description: stringPtr("Test epic description"),
	}
	err = suite.epicRepo.Create(suite.testEpic)
	suite.Require().NoError(err)
}

func (suite *UserStoryIntegrationTestSuite) TearDownSuite() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

func (suite *UserStoryIntegrationTestSuite) TestCreateUserStory() {
	description := "As a user, I want to login, so that I can access my account"
	reqBody := service.CreateUserStoryRequest{
		EpicID:      suite.testEpic.ID,
		CreatorID:   suite.testUser.ID,
		Priority:    models.PriorityHigh,
		Title:       "User Login",
		Description: &description,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/user-stories", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response models.UserStory
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.testEpic.ID, response.EpicID)
	assert.Equal(suite.T(), suite.testUser.ID, response.CreatorID)
	assert.Equal(suite.T(), suite.testUser.ID, response.AssigneeID) // Should default to creator
	assert.Equal(suite.T(), models.PriorityHigh, response.Priority)
	assert.Equal(suite.T(), models.UserStoryStatusBacklog, response.Status)
	assert.Equal(suite.T(), "User Login", response.Title)
	assert.Equal(suite.T(), &description, response.Description)
	assert.NotEmpty(suite.T(), response.ReferenceID)
}

func (suite *UserStoryIntegrationTestSuite) TestCreateUserStoryInEpic() {
	description := "As a user, I want to register, so that I can create an account"
	reqBody := service.CreateUserStoryRequest{
		CreatorID:   suite.testUser.ID,
		Priority:    models.PriorityMedium,
		Title:       "User Registration",
		Description: &description,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/epics/%s/user-stories", suite.testEpic.ID.String()), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response models.UserStory
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.testEpic.ID, response.EpicID)
	assert.Equal(suite.T(), "User Registration", response.Title)
}

func (suite *UserStoryIntegrationTestSuite) TestCreateUserStoryWithInvalidTemplate() {
	invalidDescription := "This is not a proper user story template"
	reqBody := service.CreateUserStoryRequest{
		EpicID:      suite.testEpic.ID,
		CreatorID:   suite.testUser.ID,
		Priority:    models.PriorityHigh,
		Title:       "Invalid User Story",
		Description: &invalidDescription,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/user-stories", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "template")
}

func (suite *UserStoryIntegrationTestSuite) TestGetUserStory() {
	// Create a user story first
	description := "As a user, I want to view my profile, so that I can see my information"
	userStory := &models.UserStory{
		ID:          uuid.New(),
		ReferenceID: "US-001",
		EpicID:      suite.testEpic.ID,
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityMedium,
		Status:      models.UserStoryStatusBacklog,
		Title:       "View Profile",
		Description: &description,
	}
	err := suite.userStoryRepo.Create(userStory)
	suite.Require().NoError(err)

	// Test get by UUID
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/user-stories/%s", userStory.ID.String()), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.UserStory
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), userStory.ID, response.ID)
	assert.Equal(suite.T(), userStory.Title, response.Title)

	// Test get by reference ID
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/user-stories/%s", userStory.ReferenceID), nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *UserStoryIntegrationTestSuite) TestUpdateUserStory() {
	// Create a user story first
	description := "As a user, I want to edit my profile, so that I can update my information"
	userStory := &models.UserStory{
		ID:          uuid.New(),
		ReferenceID: "US-002",
		EpicID:      suite.testEpic.ID,
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityMedium,
		Status:      models.UserStoryStatusBacklog,
		Title:       "Edit Profile",
		Description: &description,
	}
	err := suite.userStoryRepo.Create(userStory)
	suite.Require().NoError(err)

	// Update the user story
	newTitle := "Update Profile Information"
	newStatus := models.UserStoryStatusInProgress
	newDescription := "As a user, I want to update my profile, so that I can keep my information current"

	reqBody := service.UpdateUserStoryRequest{
		Title:       &newTitle,
		Status:      &newStatus,
		Description: &newDescription,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/user-stories/%s", userStory.ID.String()), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.UserStory
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newTitle, response.Title)
	assert.Equal(suite.T(), newStatus, response.Status)
	assert.Equal(suite.T(), &newDescription, response.Description)
}

func (suite *UserStoryIntegrationTestSuite) TestUpdateUserStoryWithInvalidTemplate() {
	// Create a user story first
	description := "As a user, I want to edit my profile, so that I can update my information"
	userStory := &models.UserStory{
		ID:          uuid.New(),
		ReferenceID: "US-003",
		EpicID:      suite.testEpic.ID,
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityMedium,
		Status:      models.UserStoryStatusBacklog,
		Title:       "Edit Profile",
		Description: &description,
	}
	err := suite.userStoryRepo.Create(userStory)
	suite.Require().NoError(err)

	// Try to update with invalid template
	invalidDescription := "This is not a proper user story template"
	reqBody := service.UpdateUserStoryRequest{
		Description: &invalidDescription,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/user-stories/%s", userStory.ID.String()), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "template")
}

func (suite *UserStoryIntegrationTestSuite) TestDeleteUserStory() {
	// Create a user story first
	description := "As a user, I want to delete my account, so that I can remove my data"
	userStory := &models.UserStory{
		ID:          uuid.New(),
		ReferenceID: "US-004",
		EpicID:      suite.testEpic.ID,
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityLow,
		Status:      models.UserStoryStatusBacklog,
		Title:       "Delete Account",
		Description: &description,
	}
	err := suite.userStoryRepo.Create(userStory)
	suite.Require().NoError(err)

	// Delete the user story
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/user-stories/%s", userStory.ID.String()), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)

	// Verify it's deleted
	_, err = suite.userStoryRepo.GetByID(userStory.ID)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), repository.ErrNotFound, err)
}

func (suite *UserStoryIntegrationTestSuite) TestListUserStories() {
	// Create multiple user stories
	descriptions := []string{
		"As a user, I want to login, so that I can access my account",
		"As a user, I want to logout, so that I can secure my session",
		"As an admin, I want to manage users, so that I can control access",
	}

	for i, desc := range descriptions {
		userStory := &models.UserStory{
			ID:          uuid.New(),
			ReferenceID: fmt.Sprintf("US-%03d", i+1),
			EpicID:      suite.testEpic.ID,
			CreatorID:   suite.testUser.ID,
			AssigneeID:  suite.testUser.ID,
			Priority:    models.Priority(i%4 + 1),
			Status:      models.UserStoryStatusBacklog,
			Title:       fmt.Sprintf("User Story %d", i+1),
			Description: &desc,
		}
		err := suite.userStoryRepo.Create(userStory)
		suite.Require().NoError(err)
	}

	// Test list all user stories
	req, _ := http.NewRequest("GET", "/api/v1/user-stories", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(3), response["count"])

	// Test list with epic filter
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/user-stories?epic_id=%s", suite.testEpic.ID.String()), nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(3), response["count"])
}

func (suite *UserStoryIntegrationTestSuite) TestChangeUserStoryStatus() {
	// Create a user story first
	description := "As a user, I want to change my password, so that I can secure my account"
	userStory := &models.UserStory{
		ID:          uuid.New(),
		ReferenceID: "US-005",
		EpicID:      suite.testEpic.ID,
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityHigh,
		Status:      models.UserStoryStatusBacklog,
		Title:       "Change Password",
		Description: &description,
	}
	err := suite.userStoryRepo.Create(userStory)
	suite.Require().NoError(err)

	// Change status
	reqBody := map[string]interface{}{
		"status": models.UserStoryStatusInProgress,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/user-stories/%s/status", userStory.ID.String()), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.UserStory
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), models.UserStoryStatusInProgress, response.Status)
}

func (suite *UserStoryIntegrationTestSuite) TestAssignUserStory() {
	// Create another user for assignment
	assignee := &models.User{
		ID:       uuid.New(),
		Username: "assignee",
		Email:    "assignee@example.com",
		Role:     models.RoleUser,
	}
	err := suite.userRepo.Create(assignee)
	suite.Require().NoError(err)

	// Create a user story first
	description := "As a user, I want to reset my password, so that I can recover my account"
	userStory := &models.UserStory{
		ID:          uuid.New(),
		ReferenceID: "US-006",
		EpicID:      suite.testEpic.ID,
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityMedium,
		Status:      models.UserStoryStatusBacklog,
		Title:       "Reset Password",
		Description: &description,
	}
	err = suite.userStoryRepo.Create(userStory)
	suite.Require().NoError(err)

	// Assign to different user
	reqBody := map[string]interface{}{
		"assignee_id": assignee.ID,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/user-stories/%s/assign", userStory.ID.String()), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.UserStory
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), assignee.ID, response.AssigneeID)
}

func TestUserStoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(UserStoryIntegrationTestSuite))
}
