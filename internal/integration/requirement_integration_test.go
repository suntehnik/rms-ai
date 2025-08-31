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
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"product-requirements-management/internal/handlers"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/service"
)

type RequirementIntegrationTestSuite struct {
	suite.Suite
	db                              *gorm.DB
	router                          *gin.Engine
	requirementHandler              *handlers.RequirementHandler
	requirementService              service.RequirementService
	requirementRepo                 repository.RequirementRepository
	requirementTypeRepo             repository.RequirementTypeRepository
	relationshipTypeRepo            repository.RelationshipTypeRepository
	requirementRelationshipRepo     repository.RequirementRelationshipRepository
	userStoryRepo                   repository.UserStoryRepository
	acceptanceCriteriaRepo          repository.AcceptanceCriteriaRepository
	epicRepo                        repository.EpicRepository
	userRepo                        repository.UserRepository
	testUser                        *models.User
	testEpic                        *models.Epic
	testUserStory                   *models.UserStory
	testAcceptanceCriteria          *models.AcceptanceCriteria
	testRequirementType             *models.RequirementType
	testRelationshipType            *models.RelationshipType
}

func (suite *RequirementIntegrationTestSuite) SetupSuite() {
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
		&models.RequirementType{},
		&models.RelationshipType{},
		&models.RequirementRelationship{},
		&models.Comment{},
	)
	suite.Require().NoError(err)

	suite.db = db

	// Setup repositories
	suite.userRepo = repository.NewUserRepository(db)
	suite.epicRepo = repository.NewEpicRepository(db)
	suite.userStoryRepo = repository.NewUserStoryRepository(db)
	suite.acceptanceCriteriaRepo = repository.NewAcceptanceCriteriaRepository(db)
	suite.requirementRepo = repository.NewRequirementRepository(db)
	suite.requirementTypeRepo = repository.NewRequirementTypeRepository(db)
	suite.relationshipTypeRepo = repository.NewRelationshipTypeRepository(db)
	suite.requirementRelationshipRepo = repository.NewRequirementRelationshipRepository(db)

	// Setup services
	suite.requirementService = service.NewRequirementService(
		suite.requirementRepo,
		suite.requirementTypeRepo,
		suite.relationshipTypeRepo,
		suite.requirementRelationshipRepo,
		suite.userStoryRepo,
		suite.acceptanceCriteriaRepo,
		suite.userRepo,
	)

	// Setup handlers
	suite.requirementHandler = handlers.NewRequirementHandler(suite.requirementService)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	
	// Setup routes
	v1 := suite.router.Group("/api/v1")
	{
		v1.POST("/requirements", suite.requirementHandler.CreateRequirement)
		v1.GET("/requirements/:id", suite.requirementHandler.GetRequirement)
		v1.PUT("/requirements/:id", suite.requirementHandler.UpdateRequirement)
		v1.DELETE("/requirements/:id", suite.requirementHandler.DeleteRequirement)
		v1.GET("/requirements", suite.requirementHandler.ListRequirements)
		v1.GET("/requirements/:id/relationships", suite.requirementHandler.GetRequirementWithRelationships)
		v1.PATCH("/requirements/:id/status", suite.requirementHandler.ChangeRequirementStatus)
		v1.PATCH("/requirements/:id/assign", suite.requirementHandler.AssignRequirement)
		v1.POST("/requirements/relationships", suite.requirementHandler.CreateRelationship)
		v1.DELETE("/requirement-relationships/:id", suite.requirementHandler.DeleteRelationship)
		v1.GET("/requirements/search", suite.requirementHandler.SearchRequirements)
		v1.POST("/user-stories/:id/requirements", suite.requirementHandler.CreateRequirementInUserStory)
	}
}

func (suite *RequirementIntegrationTestSuite) SetupTest() {
	// Clean up database before each test
	suite.db.Exec("DELETE FROM requirement_relationships")
	suite.db.Exec("DELETE FROM requirements")
	suite.db.Exec("DELETE FROM acceptance_criteria")
	suite.db.Exec("DELETE FROM user_stories")
	suite.db.Exec("DELETE FROM epics")
	suite.db.Exec("DELETE FROM users")
	suite.db.Exec("DELETE FROM requirement_types")
	suite.db.Exec("DELETE FROM relationship_types")

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
		ID:         uuid.New(),
		CreatorID:  suite.testUser.ID,
		AssigneeID: suite.testUser.ID,
		Priority:   models.PriorityHigh,
		Status:     models.EpicStatusBacklog,
		Title:      "Test Epic",
	}
	err = suite.epicRepo.Create(suite.testEpic)
	suite.Require().NoError(err)

	// Create test user story
	description := "As a user, I want to test, so that I can verify functionality"
	suite.testUserStory = &models.UserStory{
		ID:          uuid.New(),
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

	// Create test acceptance criteria
	suite.testAcceptanceCriteria = &models.AcceptanceCriteria{
		ID:          uuid.New(),
		UserStoryID: suite.testUserStory.ID,
		AuthorID:    suite.testUser.ID,
		Description: "WHEN user performs action THEN system SHALL respond",
	}
	err = suite.acceptanceCriteriaRepo.Create(suite.testAcceptanceCriteria)
	suite.Require().NoError(err)

	// Create test requirement type
	suite.testRequirementType = &models.RequirementType{
		ID:          uuid.New(),
		Name:        "Functional",
		Description: stringPtr("Functional requirements"),
	}
	err = suite.requirementTypeRepo.Create(suite.testRequirementType)
	suite.Require().NoError(err)

	// Create test relationship type
	suite.testRelationshipType = &models.RelationshipType{
		ID:          uuid.New(),
		Name:        "depends_on",
		Description: stringPtr("Dependency relationship"),
	}
	err = suite.relationshipTypeRepo.Create(suite.testRelationshipType)
	suite.Require().NoError(err)
}

func (suite *RequirementIntegrationTestSuite) TestCreateRequirement() {
	reqBody := service.CreateRequirementRequest{
		UserStoryID:          suite.testUserStory.ID,
		AcceptanceCriteriaID: &suite.testAcceptanceCriteria.ID,
		CreatorID:            suite.testUser.ID,
		Priority:             models.PriorityHigh,
		TypeID:               suite.testRequirementType.ID,
		Title:                "Test Requirement",
		Description:          stringPtr("Test requirement description"),
	}

	jsonBody, err := json.Marshal(reqBody)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", "/api/v1/requirements", bytes.NewBuffer(jsonBody))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusCreated, w.Code)

	var response models.Requirement
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	suite.Equal(reqBody.UserStoryID, response.UserStoryID)
	suite.Equal(reqBody.CreatorID, response.CreatorID)
	suite.Equal(reqBody.CreatorID, response.AssigneeID) // Should default to creator
	suite.Equal(reqBody.Priority, response.Priority)
	suite.Equal(models.RequirementStatusDraft, response.Status) // Should default to Draft
	suite.Equal(reqBody.TypeID, response.TypeID)
	suite.Equal(reqBody.Title, response.Title)
	suite.NotEmpty(response.ReferenceID)
}

func (suite *RequirementIntegrationTestSuite) TestCreateRequirementInUserStory() {
	reqBody := service.CreateRequirementRequest{
		CreatorID:   suite.testUser.ID,
		Priority:    models.PriorityMedium,
		TypeID:      suite.testRequirementType.ID,
		Title:       "Test Requirement in User Story",
		Description: stringPtr("Test requirement description"),
	}

	jsonBody, err := json.Marshal(reqBody)
	suite.Require().NoError(err)

	url := fmt.Sprintf("/api/v1/user-stories/%s/requirements", suite.testUserStory.ID.String())
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		suite.T().Logf("Response body: %s", w.Body.String())
	}
	suite.Equal(http.StatusCreated, w.Code)

	var response models.Requirement
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	suite.Equal(suite.testUserStory.ID, response.UserStoryID)
	suite.Equal(reqBody.CreatorID, response.CreatorID)
	suite.Equal(reqBody.Priority, response.Priority)
	suite.Equal(reqBody.Title, response.Title)
}

func (suite *RequirementIntegrationTestSuite) TestGetRequirement() {
	// Create a requirement first
	requirement := &models.Requirement{
		ID:          uuid.New(),
		ReferenceID: "REQ-001",
		UserStoryID: suite.testUserStory.ID,
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityHigh,
		Status:      models.RequirementStatusDraft,
		TypeID:      suite.testRequirementType.ID,
		Title:       "Test Requirement",
	}
	err := suite.requirementRepo.Create(requirement)
	suite.Require().NoError(err)

	// Test get by UUID
	req, err := http.NewRequest("GET", "/api/v1/requirements/"+requirement.ID.String(), nil)
	suite.Require().NoError(err)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response models.Requirement
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	suite.Equal(requirement.ID, response.ID)
	suite.Equal(requirement.Title, response.Title)
}

func (suite *RequirementIntegrationTestSuite) TestUpdateRequirement() {
	// Create a requirement first
	requirement := &models.Requirement{
		ID:          uuid.New(),
		ReferenceID: "REQ-001",
		UserStoryID: suite.testUserStory.ID,
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityHigh,
		Status:      models.RequirementStatusDraft,
		TypeID:      suite.testRequirementType.ID,
		Title:       "Original Title",
	}
	err := suite.requirementRepo.Create(requirement)
	suite.Require().NoError(err)

	// Update request
	newTitle := "Updated Title"
	newPriority := models.PriorityLow
	newStatus := models.RequirementStatusActive
	
	updateReq := service.UpdateRequirementRequest{
		Title:    &newTitle,
		Priority: &newPriority,
		Status:   &newStatus,
	}

	jsonBody, err := json.Marshal(updateReq)
	suite.Require().NoError(err)

	req, err := http.NewRequest("PUT", "/api/v1/requirements/"+requirement.ID.String(), bytes.NewBuffer(jsonBody))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response models.Requirement
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	suite.Equal(newTitle, response.Title)
	suite.Equal(newPriority, response.Priority)
	suite.Equal(newStatus, response.Status)
}

func (suite *RequirementIntegrationTestSuite) TestDeleteRequirement() {
	// Create a requirement first
	requirement := &models.Requirement{
		ID:          uuid.New(),
		ReferenceID: "REQ-001",
		UserStoryID: suite.testUserStory.ID,
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityHigh,
		Status:      models.RequirementStatusDraft,
		TypeID:      suite.testRequirementType.ID,
		Title:       "Test Requirement",
	}
	err := suite.requirementRepo.Create(requirement)
	suite.Require().NoError(err)

	// Delete request
	req, err := http.NewRequest("DELETE", "/api/v1/requirements/"+requirement.ID.String(), nil)
	suite.Require().NoError(err)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusNoContent, w.Code)

	// Verify requirement is deleted
	_, err = suite.requirementRepo.GetByID(requirement.ID)
	suite.Error(err)
}

func (suite *RequirementIntegrationTestSuite) TestCreateRelationship() {
	// Create two requirements
	requirement1 := &models.Requirement{
		ID:          uuid.New(),
		ReferenceID: "REQ-001",
		UserStoryID: suite.testUserStory.ID,
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityHigh,
		Status:      models.RequirementStatusDraft,
		TypeID:      suite.testRequirementType.ID,
		Title:       "Requirement 1",
	}
	err := suite.requirementRepo.Create(requirement1)
	suite.Require().NoError(err)

	requirement2 := &models.Requirement{
		ID:          uuid.New(),
		ReferenceID: "REQ-002",
		UserStoryID: suite.testUserStory.ID,
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityHigh,
		Status:      models.RequirementStatusDraft,
		TypeID:      suite.testRequirementType.ID,
		Title:       "Requirement 2",
	}
	err = suite.requirementRepo.Create(requirement2)
	suite.Require().NoError(err)

	// Create relationship
	relationshipReq := service.CreateRelationshipRequest{
		SourceRequirementID: requirement1.ID,
		TargetRequirementID: requirement2.ID,
		RelationshipTypeID:  suite.testRelationshipType.ID,
		CreatedBy:           suite.testUser.ID,
	}

	jsonBody, err := json.Marshal(relationshipReq)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", "/api/v1/requirements/relationships", bytes.NewBuffer(jsonBody))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusCreated, w.Code)

	var response models.RequirementRelationship
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	suite.Equal(requirement1.ID, response.SourceRequirementID)
	suite.Equal(requirement2.ID, response.TargetRequirementID)
	suite.Equal(suite.testRelationshipType.ID, response.RelationshipTypeID)
	suite.Equal(suite.testUser.ID, response.CreatedBy)
}

func (suite *RequirementIntegrationTestSuite) TestListRequirements() {
	// Create multiple requirements
	for i := 0; i < 3; i++ {
		requirement := &models.Requirement{
			ID:          uuid.New(),
			ReferenceID: fmt.Sprintf("REQ-%03d", i+1),
			UserStoryID: suite.testUserStory.ID,
			CreatorID:   suite.testUser.ID,
			AssigneeID:  suite.testUser.ID,
			Priority:    models.PriorityHigh,
			Status:      models.RequirementStatusDraft,
			TypeID:      suite.testRequirementType.ID,
			Title:       fmt.Sprintf("Test Requirement %d", i+1),
		}
		err := suite.requirementRepo.Create(requirement)
		suite.Require().NoError(err)
	}

	// List requirements
	req, err := http.NewRequest("GET", "/api/v1/requirements", nil)
	suite.Require().NoError(err)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	suite.Equal(float64(3), response["count"])
	requirements := response["requirements"].([]interface{})
	suite.Len(requirements, 3)
}

func (suite *RequirementIntegrationTestSuite) TestSearchRequirements() {
	// Create requirements with different titles
	requirement1 := &models.Requirement{
		ID:          uuid.New(),
		ReferenceID: "REQ-001",
		UserStoryID: suite.testUserStory.ID,
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityHigh,
		Status:      models.RequirementStatusDraft,
		TypeID:      suite.testRequirementType.ID,
		Title:       "Login functionality requirement",
	}
	err := suite.requirementRepo.Create(requirement1)
	suite.Require().NoError(err)

	requirement2 := &models.Requirement{
		ID:          uuid.New(),
		ReferenceID: "REQ-002",
		UserStoryID: suite.testUserStory.ID,
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityHigh,
		Status:      models.RequirementStatusDraft,
		TypeID:      suite.testRequirementType.ID,
		Title:       "Database connection requirement",
	}
	err = suite.requirementRepo.Create(requirement2)
	suite.Require().NoError(err)

	// Search for "login"
	req, err := http.NewRequest("GET", "/api/v1/requirements/search?q=login", nil)
	suite.Require().NoError(err)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	suite.Equal("login", response["query"])
	requirements := response["requirements"].([]interface{})
	suite.Len(requirements, 1)
}

func (suite *RequirementIntegrationTestSuite) TestChangeRequirementStatus() {
	// Create a requirement first
	requirement := &models.Requirement{
		ID:          uuid.New(),
		ReferenceID: "REQ-001",
		UserStoryID: suite.testUserStory.ID,
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    models.PriorityHigh,
		Status:      models.RequirementStatusDraft,
		TypeID:      suite.testRequirementType.ID,
		Title:       "Test Requirement",
	}
	err := suite.requirementRepo.Create(requirement)
	suite.Require().NoError(err)

	// Change status
	statusReq := map[string]string{
		"status": string(models.RequirementStatusActive),
	}

	jsonBody, err := json.Marshal(statusReq)
	suite.Require().NoError(err)

	req, err := http.NewRequest("PATCH", "/api/v1/requirements/"+requirement.ID.String()+"/status", bytes.NewBuffer(jsonBody))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response models.Requirement
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	suite.Equal(models.RequirementStatusActive, response.Status)
}

func TestRequirementIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(RequirementIntegrationTestSuite))
}

