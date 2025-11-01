package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/config"
	"product-requirements-management/internal/database"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/server/routes"
	"product-requirements-management/internal/service"
)

func TestNavigationIntegration(t *testing.T) {
	// Skip if running in short mode or Docker not available
	skipIfShort(t)
	skipIfNoDocker(t)

	logTestStart(t, "NavigationIntegration")
	defer logTestEnd(t, "NavigationIntegration")

	// Setup test database
	testDB := SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	// Setup authentication
	authCtx := SetupTestAuth(t, testDB.DB)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup database wrapper
	db := &database.DB{
		Postgres: testDB.DB,
	}

	// Setup configuration with JWT secret
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-jwt-secret-key-for-integration-tests",
		},
	}

	// Setup routes
	routes.Setup(router, cfg, db)

	// Create test data
	testData := setupNavigationTestData(t, testDB, authCtx.TestUser)

	// Helper function to create authenticated requests
	makeAuthenticatedRequest := func(method, url string) (*http.Request, *httptest.ResponseRecorder) {
		req, _ := http.NewRequest(method, url, nil)
		req.Header.Set("Authorization", "Bearer "+authCtx.Token)
		w := httptest.NewRecorder()
		return req, w
	}

	t.Run("GetHierarchy", func(t *testing.T) {
		req, w := makeAuthenticatedRequest("GET", "/api/v1/hierarchy")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response service.HierarchyResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Greater(t, len(response.Epics), 0)
		assert.Equal(t, len(response.Epics), response.Count)
	})

	t.Run("GetHierarchyWithExpansion", func(t *testing.T) {
		req, w := makeAuthenticatedRequest("GET", "/api/v1/hierarchy?expand=user_stories,requirements,acceptance_criteria")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response service.HierarchyResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Greater(t, len(response.Epics), 0)

		// Check that user stories are expanded
		if len(response.Epics) > 0 {
			epic := response.Epics[0]
			assert.NotNil(t, epic.UserStories)

			if len(epic.UserStories) > 0 {
				userStory := epic.UserStories[0]
				assert.NotNil(t, userStory.Requirements)
				assert.NotNil(t, userStory.AcceptanceCriteria)
			}
		}
	})

	t.Run("GetEpicHierarchy", func(t *testing.T) {
		epicID := testData.Epic.ID.String()
		req, w := makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/hierarchy/epics/%s", epicID))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response service.EpicHierarchy
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, testData.Epic.ID, response.ID)
		assert.NotNil(t, response.UserStories)
	})

	t.Run("GetEpicHierarchyByReferenceID", func(t *testing.T) {
		req, w := makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/hierarchy/epics/%s", testData.Epic.ReferenceID))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response service.EpicHierarchy
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, testData.Epic.ID, response.ID)
		assert.Equal(t, testData.Epic.ReferenceID, response.ReferenceID)
	})

	t.Run("GetUserStoryHierarchy", func(t *testing.T) {
		userStoryID := testData.UserStory.ID.String()
		req, w := makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/hierarchy/user-stories/%s?expand=requirements,acceptance_criteria", userStoryID))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response service.UserStoryHierarchy
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, testData.UserStory.ID, response.ID)
		assert.NotNil(t, response.Requirements)
		assert.NotNil(t, response.AcceptanceCriteria)
	})

	t.Run("GetEntityPath", func(t *testing.T) {
		requirementID := testData.Requirement.ID.String()
		req, w := makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/hierarchy/path/requirement/%s", requirementID))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string][]service.PathElement
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		path := response["path"]
		assert.Len(t, path, 3) // Epic -> UserStory -> Requirement

		// Check path order (Epic first, Requirement last)
		assert.Equal(t, "epic", path[0].Type)
		assert.Equal(t, "user_story", path[1].Type)
		assert.Equal(t, "requirement", path[2].Type)

		assert.Equal(t, testData.Epic.ID, path[0].ID)
		assert.Equal(t, testData.UserStory.ID, path[1].ID)
		assert.Equal(t, testData.Requirement.ID, path[2].ID)
	})

	t.Run("GetHierarchyWithSorting", func(t *testing.T) {
		req, w := makeAuthenticatedRequest("GET", "/api/v1/hierarchy?order_by=priority&order_dir=asc")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response service.HierarchyResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Greater(t, len(response.Epics), 0)
	})

	t.Run("GetHierarchyWithFiltering", func(t *testing.T) {
		creatorID := testData.User.ID.String()
		req, w := makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/hierarchy?creator_id=%s", creatorID))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response service.HierarchyResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// All epics should be created by the test user
		for _, epic := range response.Epics {
			assert.Equal(t, testData.User.ID, epic.CreatorID)
		}
	})

	t.Run("GetNonExistentEpicHierarchy", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		req, w := makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/hierarchy/epics/%s", nonExistentID))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("GetInvalidEntityTypePath", func(t *testing.T) {
		req, w := makeAuthenticatedRequest("GET", "/api/v1/hierarchy/path/invalid_type/123")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UnauthorizedAccess", func(t *testing.T) {
		// Test without authentication token
		req, _ := http.NewRequest("GET", "/api/v1/hierarchy", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		// Test with invalid token
		req, _ := http.NewRequest("GET", "/api/v1/hierarchy", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// NavigationTestData holds test data for navigation tests
type NavigationTestData struct {
	User               *models.User
	Epic               *models.Epic
	UserStory          *models.UserStory
	AcceptanceCriteria *models.AcceptanceCriteria
	Requirement        *models.Requirement
	RequirementType    *models.RequirementType
}

// setupNavigationTestData creates test data for navigation tests
func setupNavigationTestData(t *testing.T, testDB *TestDatabase, user *models.User) *NavigationTestData {
	// Create repositories
	repos := repository.NewRepositories(testDB.DB, nil)

	// Get or create requirement type
	reqType, err := repos.RequirementType.GetByName("Functional")
	if err != nil {
		// Create if doesn't exist
		reqType = &models.RequirementType{
			ID:          uuid.New(),
			Name:        "TestFunctional",
			Description: stringPtr("Test functional requirement"),
		}
		err = repos.RequirementType.Create(reqType)
		require.NoError(t, err)
	}

	// Create epic
	epic := &models.Epic{
		ID:          uuid.New(),
		ReferenceID: "EP-001",
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityHigh,
		Status:      models.EpicStatusDraft,
		Title:       "Test Epic",
		Description: stringPtr("Test epic description"),
	}
	err = repos.Epic.Create(epic)
	require.NoError(t, err)

	// Create user story
	userStory := &models.UserStory{
		ID:          uuid.New(),
		ReferenceID: "US-001",
		EpicID:      epic.ID,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityMedium,
		Status:      models.UserStoryStatusDraft,
		Title:       "Test User Story",
		Description: stringPtr("As a user, I want to test, so that I can verify functionality"),
	}
	err = repos.UserStory.Create(userStory)
	require.NoError(t, err)

	// Create acceptance criteria
	acceptanceCriteria := &models.AcceptanceCriteria{
		ID:          uuid.New(),
		ReferenceID: "AC-001",
		UserStoryID: userStory.ID,
		AuthorID:    user.ID,
		Description: "WHEN user performs action THEN system SHALL respond",
	}
	err = repos.AcceptanceCriteria.Create(acceptanceCriteria)
	require.NoError(t, err)

	// Create requirement
	requirement := &models.Requirement{
		ID:                   uuid.New(),
		ReferenceID:          "REQ-001",
		UserStoryID:          userStory.ID,
		AcceptanceCriteriaID: &acceptanceCriteria.ID,
		CreatorID:            user.ID,
		AssigneeID:           user.ID,
		Priority:             models.PriorityHigh,
		Status:               models.RequirementStatusDraft,
		TypeID:               reqType.ID,
		Title:                "Test Requirement",
		Description:          stringPtr("Test requirement description"),
	}
	err = repos.Requirement.Create(requirement)
	require.NoError(t, err)

	return &NavigationTestData{
		User:               user,
		Epic:               epic,
		UserStory:          userStory,
		AcceptanceCriteria: acceptanceCriteria,
		Requirement:        requirement,
		RequirementType:    reqType,
	}
}
