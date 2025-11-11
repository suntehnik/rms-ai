package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"product-requirements-management/internal/handlers"
	"product-requirements-management/internal/mcp/tools"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/service"
)

type EpicHierarchyIntegrationTestSuite struct {
	suite.Suite
	db              *gorm.DB
	testDatabase    *TestDatabase
	router          *gin.Engine
	mcpHandler      *handlers.MCPHandler
	epicHandler     *tools.EpicHandler
	epicService     service.EpicService
	userService     service.UserService
	epicRepo        repository.EpicRepository
	userStoryRepo   repository.UserStoryRepository
	requirementRepo repository.RequirementRepository
	acRepo          repository.AcceptanceCriteriaRepository
	userRepo        repository.UserRepository
	testUser        *models.User
	authContext     *TestAuthContext
}

func (suite *EpicHierarchyIntegrationTestSuite) SetupSuite() {
	// Setup test database with SQL migrations
	suite.testDatabase = SetupTestDatabase(suite.T())
	suite.db = suite.testDatabase.DB

	// Setup repositories
	suite.userRepo = repository.NewUserRepository(suite.db)
	suite.epicRepo = repository.NewEpicRepository(suite.db)
	suite.userStoryRepo = repository.NewUserStoryRepository(suite.db, nil) // nil redis for tests
	suite.requirementRepo = repository.NewRequirementRepository(suite.db)
	suite.acRepo = repository.NewAcceptanceCriteriaRepository(suite.db)

	// Setup services
	suite.epicService = service.NewEpicService(suite.epicRepo, suite.userRepo)
	suite.userService = service.NewUserService(suite.userRepo)

	// Setup MCP tool handlers
	suite.epicHandler = tools.NewEpicHandler(suite.epicService, suite.userService)

	// Setup MCP handler
	suite.mcpHandler = handlers.NewMCPHandler(
		suite.epicService,
		nil, // userStoryService
		nil, // requirementService
		nil, // acceptanceCriteriaService
		nil, // searchService
		nil, // configService
		nil, // commentService
		nil, // promptService
		nil, // resourceService
		nil, // statusValidator
	)

	// Setup authentication
	suite.authContext = SetupTestAuth(suite.T(), suite.db)

	// Setup router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	v1 := suite.router.Group("/api/v1")
	v1.Use(suite.authContext.AuthService.Middleware())
	{
		v1.POST("/mcp", suite.mcpHandler.Process)
	}
}

func (suite *EpicHierarchyIntegrationTestSuite) SetupTest() {
	// Clean up database before each test
	suite.db.Exec("DELETE FROM acceptance_criteria")
	suite.db.Exec("DELETE FROM requirements")
	suite.db.Exec("DELETE FROM user_stories")
	suite.db.Exec("DELETE FROM epics")
	suite.db.Exec("DELETE FROM users WHERE username NOT IN ('testuser', 'adminuser')")

	// Use the authenticated test user
	suite.testUser = suite.authContext.TestUser
}

func (suite *EpicHierarchyIntegrationTestSuite) TearDownSuite() {
	if suite.testDatabase != nil {
		suite.testDatabase.Cleanup(suite.T())
	}
}

// TestEpicHierarchy_CompleteHierarchy tests the epic_hierarchy tool with a full hierarchy
func (suite *EpicHierarchyIntegrationTestSuite) TestEpicHierarchy_CompleteHierarchy() {
	// Create test data: Epic → UserStories → Requirements + AcceptanceCriteria
	epic := suite.createTestEpic("Test Epic for Hierarchy", models.PriorityHigh, models.EpicStatusInProgress)

	// Create first user story with requirements and acceptance criteria
	us1 := suite.createTestUserStory(epic.ID, "User Story 1", models.PriorityHigh, models.UserStoryStatusInProgress)
	req1 := suite.createTestRequirement(us1.ID, "Requirement 1", models.PriorityHigh, models.RequirementStatusActive)
	req2 := suite.createTestRequirement(us1.ID, "Requirement 2", models.PriorityMedium, models.RequirementStatusDraft)
	ac1 := suite.createTestAcceptanceCriteria(us1.ID, "First acceptance criteria for user story 1")
	ac2 := suite.createTestAcceptanceCriteria(us1.ID, "Second acceptance criteria with a very long description that should be truncated to 80 characters maximum")

	// Create second user story with only requirements
	us2 := suite.createTestUserStory(epic.ID, "User Story 2", models.PriorityMedium, models.UserStoryStatusBacklog)
	req3 := suite.createTestRequirement(us2.ID, "Requirement 3", models.PriorityLow, models.RequirementStatusDraft)

	// Create third user story with no requirements or acceptance criteria
	us3 := suite.createTestUserStory(epic.ID, "User Story 3", models.PriorityLow, models.UserStoryStatusBacklog)

	// Call epic_hierarchy tool via MCP handler
	requestBody := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "epic_hierarchy",
			"arguments": {
				"epic": "%s"
			}
		}
	}`, epic.ReferenceID)

	req, _ := http.NewRequest("POST", "/api/v1/mcp", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Verify response
	require.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	// Check JSON-RPC structure
	assert.Equal(suite.T(), "2.0", response["jsonrpc"])
	assert.Equal(suite.T(), float64(1), response["id"])
	assert.Nil(suite.T(), response["error"])

	// Extract result
	result, ok := response["result"].(map[string]interface{})
	require.True(suite.T(), ok, "Result should be an object")

	content, ok := result["content"].([]interface{})
	require.True(suite.T(), ok, "Content should be an array")
	require.Len(suite.T(), content, 1, "Should have one content item")

	contentItem := content[0].(map[string]interface{})
	assert.Equal(suite.T(), "text", contentItem["type"])

	treeOutput := contentItem["text"].(string)

	// Verify output format matches specification
	suite.verifyHierarchyOutput(treeOutput, epic, us1, us2, us3, req1, req2, req3, ac1, ac2)
}

// TestEpicHierarchy_EmptyHierarchy tests the epic_hierarchy tool with an epic that has no user stories
func (suite *EpicHierarchyIntegrationTestSuite) TestEpicHierarchy_EmptyHierarchy() {
	// Create epic with no user stories
	epic := suite.createTestEpic("Empty Epic", models.PriorityMedium, models.EpicStatusBacklog)

	// Call epic_hierarchy tool
	requestBody := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "epic_hierarchy",
			"arguments": {
				"epic": "%s"
			}
		}
	}`, epic.ID.String())

	req, _ := http.NewRequest("POST", "/api/v1/mcp", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Verify response
	require.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	result := response["result"].(map[string]interface{})
	content := result["content"].([]interface{})
	contentItem := content[0].(map[string]interface{})
	treeOutput := contentItem["text"].(string)

	// Verify empty state message
	assert.Contains(suite.T(), treeOutput, epic.ReferenceID)
	assert.Contains(suite.T(), treeOutput, "No steering documents or user stories attached")
}

// TestEpicHierarchy_EpicNotFound tests the epic_hierarchy tool with a non-existent epic
func (suite *EpicHierarchyIntegrationTestSuite) TestEpicHierarchy_EpicNotFound() {
	// Call epic_hierarchy tool with non-existent epic
	requestBody := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "epic_hierarchy",
			"arguments": {
				"epic": "EP-999"
			}
		}
	}`

	req, _ := http.NewRequest("POST", "/api/v1/mcp", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Verify response
	require.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	// Should have error
	assert.NotNil(suite.T(), response["error"])
	errorObj := response["error"].(map[string]interface{})
	// The error message is "Invalid params" because the reference ID doesn't exist
	// and the parseUUIDOrReferenceID function returns an error
	assert.Contains(suite.T(), errorObj["message"], "Invalid params")
}

// TestEpicHierarchy_InvalidReferenceID tests the epic_hierarchy tool with invalid reference ID
func (suite *EpicHierarchyIntegrationTestSuite) TestEpicHierarchy_InvalidReferenceID() {
	// Call epic_hierarchy tool with invalid reference ID
	requestBody := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "epic_hierarchy",
			"arguments": {
				"epic": "INVALID-ID"
			}
		}
	}`

	req, _ := http.NewRequest("POST", "/api/v1/mcp", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Verify response
	require.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	// Should have error
	assert.NotNil(suite.T(), response["error"])
}

// TestEpicHierarchy_UUIDFormat tests the epic_hierarchy tool with UUID format
func (suite *EpicHierarchyIntegrationTestSuite) TestEpicHierarchy_UUIDFormat() {
	// Create test epic
	epic := suite.createTestEpic("Test Epic", models.PriorityHigh, models.EpicStatusBacklog)
	us := suite.createTestUserStory(epic.ID, "Test User Story", models.PriorityHigh, models.UserStoryStatusBacklog)

	// Call epic_hierarchy tool with UUID
	requestBody := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "epic_hierarchy",
			"arguments": {
				"epic": "%s"
			}
		}
	}`, epic.ID.String())

	req, _ := http.NewRequest("POST", "/api/v1/mcp", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authContext.Token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Verify response
	require.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	// Should succeed
	assert.Nil(suite.T(), response["error"])
	result := response["result"].(map[string]interface{})
	content := result["content"].([]interface{})
	contentItem := content[0].(map[string]interface{})
	treeOutput := contentItem["text"].(string)

	// Verify output contains epic and user story
	assert.Contains(suite.T(), treeOutput, epic.ReferenceID)
	assert.Contains(suite.T(), treeOutput, us.ReferenceID)
}

// Helper methods

func (suite *EpicHierarchyIntegrationTestSuite) createTestEpic(title string, priority models.Priority, status models.EpicStatus) *models.Epic {
	epic := &models.Epic{
		ID:          uuid.New(),
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    priority,
		Status:      status,
		Title:       title,
		Description: stringPtr("Test epic description"),
	}

	err := suite.epicRepo.Create(epic)
	require.NoError(suite.T(), err)

	return epic
}

func (suite *EpicHierarchyIntegrationTestSuite) createTestUserStory(epicID uuid.UUID, title string, priority models.Priority, status models.UserStoryStatus) *models.UserStory {
	us := &models.UserStory{
		ID:          uuid.New(),
		EpicID:      epicID,
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		Priority:    priority,
		Status:      status,
		Title:       title,
		Description: stringPtr("Test user story description"),
	}

	err := suite.userStoryRepo.Create(us)
	require.NoError(suite.T(), err)

	return us
}

func (suite *EpicHierarchyIntegrationTestSuite) createTestRequirement(userStoryID uuid.UUID, title string, priority models.Priority, status models.RequirementStatus) *models.Requirement {
	// Get a requirement type
	reqType, err := suite.testDatabase.GetRequirementType("Functional")
	require.NoError(suite.T(), err)

	req := &models.Requirement{
		ID:          uuid.New(),
		UserStoryID: userStoryID,
		CreatorID:   suite.testUser.ID,
		AssigneeID:  suite.testUser.ID,
		TypeID:      reqType.ID,
		Priority:    priority,
		Status:      status,
		Title:       title,
		Description: stringPtr("Test requirement description"),
	}

	err = suite.requirementRepo.Create(req)
	require.NoError(suite.T(), err)

	return req
}

func (suite *EpicHierarchyIntegrationTestSuite) createTestAcceptanceCriteria(userStoryID uuid.UUID, description string) *models.AcceptanceCriteria {
	ac := &models.AcceptanceCriteria{
		ID:          uuid.New(),
		UserStoryID: userStoryID,
		AuthorID:    suite.testUser.ID,
		Description: description,
	}

	err := suite.acRepo.Create(ac)
	require.NoError(suite.T(), err)

	return ac
}

func (suite *EpicHierarchyIntegrationTestSuite) verifyHierarchyOutput(
	output string,
	epic *models.Epic,
	us1, us2, us3 *models.UserStory,
	req1, req2, req3 *models.Requirement,
	ac1, ac2 *models.AcceptanceCriteria,
) {
	// Verify epic is at the root
	assert.Contains(suite.T(), output, epic.ReferenceID)
	assert.Contains(suite.T(), output, epic.Title)
	assert.Contains(suite.T(), output, string(epic.Status))
	assert.Contains(suite.T(), output, fmt.Sprintf("P%d", epic.Priority))

	// Verify user stories are present
	assert.Contains(suite.T(), output, us1.ReferenceID)
	assert.Contains(suite.T(), output, us1.Title)
	assert.Contains(suite.T(), output, us2.ReferenceID)
	assert.Contains(suite.T(), output, us2.Title)
	assert.Contains(suite.T(), output, us3.ReferenceID)
	assert.Contains(suite.T(), output, us3.Title)

	// Verify requirements are present
	assert.Contains(suite.T(), output, req1.ReferenceID)
	assert.Contains(suite.T(), output, req1.Title)
	assert.Contains(suite.T(), output, req2.ReferenceID)
	assert.Contains(suite.T(), output, req2.Title)
	assert.Contains(suite.T(), output, req3.ReferenceID)
	assert.Contains(suite.T(), output, req3.Title)

	// Verify acceptance criteria are present
	assert.Contains(suite.T(), output, ac1.ReferenceID)
	assert.Contains(suite.T(), output, ac2.ReferenceID)

	// Verify AC2 description is truncated (should be max 80 chars)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, ac2.ReferenceID) {
			// Line should not contain the full description
			assert.NotContains(suite.T(), line, "that should be truncated to 80 characters maximum")
			// But should contain truncated version with "..."
			assert.Contains(suite.T(), line, "...")
			break
		}
	}

	// Verify empty state for US3 (no requirements or acceptance criteria)
	assert.Contains(suite.T(), output, "No requirements")
	assert.Contains(suite.T(), output, "No acceptance criteria")

	// Verify tree structure characters are present
	assert.Contains(suite.T(), output, "├")
	assert.Contains(suite.T(), output, "└")
	assert.Contains(suite.T(), output, "│")
}

// Run the test suite
func TestEpicHierarchyIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(EpicHierarchyIntegrationTestSuite))
}
