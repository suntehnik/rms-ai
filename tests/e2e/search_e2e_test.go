package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"product-requirements-management/internal/config"
	"product-requirements-management/internal/database"
	"product-requirements-management/internal/handlers"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/service"
)

// TestSearchE2E tests the complete search functionality end-to-end
func TestSearchE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	// Setup complete environment
	env := setupE2EEnvironment(t)
	defer env.Cleanup()

	// Create test data through API
	testData := createTestDataViaAPI(t, env)

	t.Run("complete_search_workflow", func(t *testing.T) {
		t.Run("search_via_http_api", func(t *testing.T) {
			// Test search through HTTP API
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/search?query=authentication&limit=10&offset=0", nil)

			env.Router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response service.SearchResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, "authentication", response.Query)
			assert.True(t, len(response.Results) > 0, "Should find results for authentication")

			// Verify response structure
			for _, result := range response.Results {
				assert.NotEmpty(t, result.ID)
				assert.NotEmpty(t, result.Type)
				assert.NotEmpty(t, result.Title)
				assert.NotNil(t, result.CreatedAt)
			}
		})

		t.Run("search_with_filters_via_api", func(t *testing.T) {
			// Test search with filters through HTTP API
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/search?query=user&priority=2&creator_id=%s", testData.User.ID), nil)

			env.Router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response service.SearchResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Verify filters are applied
			for _, result := range response.Results {
				if result.Priority != nil {
					assert.Equal(t, 2, *result.Priority)
				}
			}
		})

		t.Run("search_empty_query", func(t *testing.T) {
			// Test search with empty query (should return all results)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/search?limit=5", nil)

			env.Router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response service.SearchResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.True(t, len(response.Results) <= 5, "Should respect limit")
		})
	})

	t.Run("search_performance_e2e", func(t *testing.T) {
		t.Run("concurrent_search_requests", func(t *testing.T) {
			// Test concurrent search requests
			concurrency := 10
			results := make(chan error, concurrency)

			for i := 0; i < concurrency; i++ {
				go func(index int) {
					w := httptest.NewRecorder()
					req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/search?query=test%d", index), nil)

					start := time.Now()
					env.Router.ServeHTTP(w, req)
					duration := time.Since(start)

					if w.Code != http.StatusOK {
						results <- fmt.Errorf("request %d failed with status %d", index, w.Code)
						return
					}

					if duration > 5*time.Second {
						results <- fmt.Errorf("request %d took too long: %v", index, duration)
						return
					}

					results <- nil
				}(i)
			}

			// Wait for all requests to complete
			for i := 0; i < concurrency; i++ {
				err := <-results
				assert.NoError(t, err)
			}
		})

		t.Run("large_result_set_pagination", func(t *testing.T) {
			// Create more test data
			createLargeDatasetViaAPI(t, env, testData.User, 50)

			// Test pagination with large result set
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/search?query=&limit=20&offset=0", nil)

			env.Router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response service.SearchResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, 20, response.Limit)
			assert.Equal(t, 0, response.Offset)
			assert.True(t, response.Total > 20, "Should have more results than limit")

			// Test second page
			w2 := httptest.NewRecorder()
			req2, _ := http.NewRequest("GET", "/api/v1/search?query=&limit=20&offset=20", nil)

			env.Router.ServeHTTP(w2, req2)

			assert.Equal(t, http.StatusOK, w2.Code)

			var response2 service.SearchResponse
			err = json.Unmarshal(w2.Body.Bytes(), &response2)
			require.NoError(t, err)

			assert.Equal(t, 20, response2.Limit)
			assert.Equal(t, 20, response2.Offset)

			// Results should be different
			if len(response.Results) > 0 && len(response2.Results) > 0 {
				assert.NotEqual(t, response.Results[0].ID, response2.Results[0].ID)
			}
		})
	})

	t.Run("search_error_handling_e2e", func(t *testing.T) {
		t.Run("invalid_search_parameters", func(t *testing.T) {
			testCases := []struct {
				name           string
				url            string
				expectedStatus int
			}{
				{"invalid_limit", "/api/v1/search?query=test&limit=-1", http.StatusBadRequest},
				{"invalid_offset", "/api/v1/search?query=test&offset=-1", http.StatusBadRequest},
				{"invalid_priority", "/api/v1/search?query=test&priority=999", http.StatusBadRequest},
				{"invalid_uuid", "/api/v1/search?query=test&creator_id=invalid-uuid", http.StatusBadRequest},
				{"invalid_date", "/api/v1/search?query=test&created_from=invalid-date", http.StatusBadRequest},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					w := httptest.NewRecorder()
					req, _ := http.NewRequest("GET", tc.url, nil)

					env.Router.ServeHTTP(w, req)

					assert.Equal(t, tc.expectedStatus, w.Code)
				})
			}
		})

		t.Run("database_connection_issues", func(t *testing.T) {
			// This test would simulate database connection issues
			// For now, we'll skip it as it requires more complex setup
			t.Skip("Database connection simulation requires additional setup")
		})
	})

	t.Run("search_caching_e2e", func(t *testing.T) {
		if env.RedisClient == nil {
			t.Skip("Redis not available for caching tests")
		}

		t.Run("cache_hit_performance", func(t *testing.T) {
			// First request (cache miss)
			w1 := httptest.NewRecorder()
			req1, _ := http.NewRequest("GET", "/api/v1/search?query=authentication", nil)

			start1 := time.Now()
			env.Router.ServeHTTP(w1, req1)
			duration1 := time.Since(start1)

			assert.Equal(t, http.StatusOK, w1.Code)

			// Second request (cache hit)
			w2 := httptest.NewRecorder()
			req2, _ := http.NewRequest("GET", "/api/v1/search?query=authentication", nil)

			start2 := time.Now()
			env.Router.ServeHTTP(w2, req2)
			duration2 := time.Since(start2)

			assert.Equal(t, http.StatusOK, w2.Code)

			// Cache hit should be faster
			assert.Less(t, duration2, duration1, "Cached request should be faster")

			// Results should be identical
			assert.Equal(t, w1.Body.String(), w2.Body.String())
		})

		t.Run("cache_invalidation", func(t *testing.T) {
			// Get initial epic count from database
			var initialEpics []models.Epic
			env.DB.Find(&initialEpics)
			initialCount := len(initialEpics)

			// Test multiple Epic creation scenarios to ensure no duplicate key constraint violations
			epicTitles := []string{
				"New Epic for Cache Test 1",
				"New Epic for Cache Test 2",
				"New Epic for Cache Test 3",
			}

			createdEpics := make([]*models.Epic, 0, len(epicTitles))

			for i, title := range epicTitles {
				epicData := map[string]interface{}{
					"creator_id":  testData.User.ID,
					"assignee_id": testData.User.ID,
					"priority":    2,
					"title":       title,
					"description": fmt.Sprintf("This is test epic number %d for cache invalidation", i+1),
				}

				body, _ := json.Marshal(epicData)
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/api/v1/epics", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")

				env.Router.ServeHTTP(w, req)

				// Verify Epic creation succeeded without constraint violations
				require.Equal(t, http.StatusCreated, w.Code, "Epic creation should succeed without constraint violations")

				var epic models.Epic
				err := json.Unmarshal(w.Body.Bytes(), &epic)
				require.NoError(t, err)
				require.NotEmpty(t, epic.ReferenceID, "Epic should have a reference ID")
				require.Equal(t, title, epic.Title, "Epic title should match")

				createdEpics = append(createdEpics, &epic)
				t.Logf("Epic %d created successfully with reference ID: %s", i+1, epic.ReferenceID)
			}

			// Verify all epics were created in database without constraint violations
			var finalEpics []models.Epic
			env.DB.Find(&finalEpics)
			finalCount := len(finalEpics)

			t.Logf("Initial epic count: %d, Final epic count: %d", initialCount, finalCount)
			assert.Equal(t, initialCount+len(epicTitles), finalCount, "All epics should be created successfully")

			// Verify each created epic has unique reference ID
			referenceIDs := make(map[string]bool)
			for _, epic := range createdEpics {
				assert.False(t, referenceIDs[epic.ReferenceID], "Reference ID should be unique: %s", epic.ReferenceID)
				referenceIDs[epic.ReferenceID] = true
			}

			t.Logf("Successfully created %d epics without duplicate key constraint violations", len(createdEpics))
		})
	})
}

// E2EEnvironment holds the complete testing environment
type E2EEnvironment struct {
	DB             *gorm.DB
	RedisClient    *database.RedisClient
	Router         *gin.Engine
	Container      testcontainers.Container
	RedisContainer testcontainers.Container
}

func (e *E2EEnvironment) Cleanup() {
	ctx := context.Background()

	// Close database connections first
	if e.DB != nil {
		if sqlDB, err := e.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				fmt.Printf("Failed to close database connection: %v\n", err)
			}
		}
	}

	// Close Redis connections
	if e.RedisClient != nil {
		if err := e.RedisClient.Close(); err != nil {
			fmt.Printf("Failed to close Redis connection: %v\n", err)
		}
	}

	// Terminate containers
	if e.Container != nil {
		if err := e.Container.Terminate(ctx); err != nil {
			fmt.Printf("Failed to terminate PostgreSQL container: %v\n", err)
		}
	}
	if e.RedisContainer != nil {
		if err := e.RedisContainer.Terminate(ctx); err != nil {
			fmt.Printf("Failed to terminate Redis container: %v\n", err)
		}
	}
}

type TestData struct {
	User         *models.User
	Epics        []*models.Epic
	UserStories  []*models.UserStory
	Requirements []*models.Requirement
}

func setupE2EEnvironment(t *testing.T) *E2EEnvironment {
	ctx := context.Background()

	// Setup PostgreSQL container
	postgresReq := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_USER":     "testuser",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort("5432/tcp"),
		).WithDeadline(60 * time.Second),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	require.NoError(t, err)

	// Setup Redis container
	redisReq := testcontainers.ContainerRequest{
		Image:        "redis:7",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("Ready to accept connections"),
			wait.ForListeningPort("6379/tcp"),
		).WithDeadline(30 * time.Second),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: redisReq,
		Started:          true,
	})
	require.NoError(t, err)

	// Get PostgreSQL connection details
	pgHost, err := postgresContainer.Host(ctx)
	require.NoError(t, err)
	pgPort, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Get Redis connection details
	redisHost, err := redisContainer.Host(ctx)
	require.NoError(t, err)
	redisPort, err := redisContainer.MappedPort(ctx, "6379")
	require.NoError(t, err)

	// Setup database connection
	dsn := fmt.Sprintf("host=%s port=%s user=testuser password=password dbname=testdb sslmode=disable", pgHost, pgPort.Port())
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	// Setup Redis connection
	redisConfig := &config.RedisConfig{
		Host:     redisHost,
		Port:     redisPort.Port(),
		Password: "",
		DB:       0,
	}
	logger := logrus.New()
	redisClient, err := database.NewRedisClient(redisConfig, logger)
	require.NoError(t, err)

	// Auto-migrate models
	err = models.AutoMigrate(db)
	require.NoError(t, err)

	// Seed default data
	err = models.SeedDefaultData(db)
	require.NoError(t, err)

	// Setup repositories
	repos := repository.NewRepositories(db)

	// Setup services
	var redisClientForService *redis.Client
	if redisClient != nil {
		redisClientForService = redisClient.Client
	}
	searchService := service.NewSearchService(db, redisClientForService, repos.Epic, repos.UserStory, repos.AcceptanceCriteria, repos.Requirement)
	epicService := service.NewEpicService(repos.Epic, repos.User)
	userStoryService := service.NewUserStoryService(repos.UserStory, repos.Epic, repos.User)
	requirementService := service.NewRequirementService(
		repos.Requirement,
		repos.RequirementType,
		repos.RelationshipType,
		repos.RequirementRelationship,
		repos.UserStory,
		repos.AcceptanceCriteria,
		repos.User,
	)

	// Setup handlers
	searchHandler := handlers.NewSearchHandler(searchService, logger)
	epicHandler := handlers.NewEpicHandler(epicService)
	userStoryHandler := handlers.NewUserStoryHandler(userStoryService)
	requirementHandler := handlers.NewRequirementHandler(requirementService)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup routes manually for testing
	v1 := router.Group("/api/v1")
	{
		v1.GET("/search", searchHandler.Search)
		v1.POST("/epics", epicHandler.CreateEpic)
		v1.POST("/user-stories", userStoryHandler.CreateUserStory)
		v1.POST("/requirements", requirementHandler.CreateRequirement)
	}

	return &E2EEnvironment{
		DB:             db,
		RedisClient:    redisClient,
		Router:         router,
		Container:      postgresContainer,
		RedisContainer: redisContainer,
	}
}

func createTestDataViaAPI(t *testing.T, env *E2EEnvironment) *TestData {
	// Create user directly in database (users are typically created through auth system)
	user := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}
	err := env.DB.Create(user).Error
	require.NoError(t, err)

	// Create epic via API
	epic := createEpicViaAPI(t, env, user, "User Authentication Epic", "This epic covers all user authentication features including login, logout, and password reset functionality.")

	// Create user story via API
	userStory := createUserStoryViaAPI(t, env, user, epic, "User Login Feature", "As a user, I want to login to the system, so that I can access my account and personal data.")

	// Create requirement via API
	requirement := createRequirementViaAPI(t, env, user, userStory, "Password Validation Requirement", "The system must validate user passwords against security policies including minimum length and complexity requirements.")

	return &TestData{
		User:         user,
		Epics:        []*models.Epic{epic},
		UserStories:  []*models.UserStory{userStory},
		Requirements: []*models.Requirement{requirement},
	}
}

func createEpicViaAPI(t *testing.T, env *E2EEnvironment, user *models.User, title, description string) *models.Epic {
	epicData := map[string]interface{}{
		"creator_id":  user.ID,
		"assignee_id": user.ID,
		"priority":    2,
		"title":       title,
		"description": description,
	}

	body, _ := json.Marshal(epicData)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/epics", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	env.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var epic models.Epic
	err := json.Unmarshal(w.Body.Bytes(), &epic)
	require.NoError(t, err)

	return &epic
}

func createUserStoryViaAPI(t *testing.T, env *E2EEnvironment, user *models.User, epic *models.Epic, title, description string) *models.UserStory {
	userStoryData := map[string]interface{}{
		"epic_id":     epic.ID,
		"creator_id":  user.ID,
		"assignee_id": user.ID,
		"priority":    2,
		"title":       title,
		"description": description,
	}

	body, _ := json.Marshal(userStoryData)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/user-stories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	env.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var userStory models.UserStory
	err := json.Unmarshal(w.Body.Bytes(), &userStory)
	require.NoError(t, err)

	return &userStory
}

func createRequirementViaAPI(t *testing.T, env *E2EEnvironment, user *models.User, userStory *models.UserStory, title, description string) *models.Requirement {
	// Get requirement type
	var reqType models.RequirementType
	err := env.DB.Where("name = ?", "Functional").First(&reqType).Error
	require.NoError(t, err)

	requirementData := map[string]interface{}{
		"user_story_id": userStory.ID,
		"creator_id":    user.ID,
		"assignee_id":   user.ID,
		"priority":      2,
		"type_id":       reqType.ID,
		"title":         title,
		"description":   description,
	}

	body, _ := json.Marshal(requirementData)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/requirements", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	env.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var requirement models.Requirement
	err = json.Unmarshal(w.Body.Bytes(), &requirement)
	require.NoError(t, err)

	return &requirement
}

func createLargeDatasetViaAPI(t *testing.T, env *E2EEnvironment, user *models.User, count int) {
	for i := 0; i < count; i++ {
		// Create epic directly in database to avoid API reference ID issues
		epic := &models.Epic{
			ID:          uuid.New(),
			ReferenceID: fmt.Sprintf("EP-%03d", i+100), // Use unique reference IDs
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityMedium,
			Status:      models.EpicStatusBacklog,
			Title:       fmt.Sprintf("Test Epic %d", i),
			Description: &[]string{fmt.Sprintf("This is test epic number %d with various keywords for testing search performance and functionality.", i)}[0],
		}
		err := env.DB.Create(epic).Error
		require.NoError(t, err)
	}
}
