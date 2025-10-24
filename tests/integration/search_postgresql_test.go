package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/service"
)

// TestSearchIntegration_PostgreSQL tests the full-text search functionality with real PostgreSQL
func TestSearchIntegration_PostgreSQL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Setup PostgreSQL container
	db := setupPostgreSQLContainer(t)
	defer cleanupDatabase(t, db)

	// Auto-migrate models
	err := models.AutoMigrate(db)
	require.NoError(t, err)

	// Seed default data
	err = models.SeedDefaultData(db)
	require.NoError(t, err)

	// Setup repositories
	repos := repository.NewRepositories(db)

	// Setup real search service with PostgreSQL full-text search
	searchService := service.NewSearchService(
		db,
		nil, // No Redis for integration tests
		repos.Epic,
		repos.UserStory,
		repos.AcceptanceCriteria,
		repos.Requirement,
		repos.SteeringDocument,
	)

	// Create test user
	user := createTestUser(t, db)

	// Create comprehensive test data
	_ = createComprehensiveTestData(t, db, user)

	t.Run("postgresql_full_text_search", func(t *testing.T) {
		t.Run("simple_word_search", func(t *testing.T) {
			options := service.SearchOptions{
				Query:     "authentication",
				Limit:     50,
				Offset:    0,
				SortBy:    "created_at",
				SortOrder: "desc",
			}

			response, err := searchService.Search(context.Background(), options)
			require.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, "authentication", response.Query)

			// Should find entities with "authentication" in title or description
			found := false
			for _, result := range response.Results {
				if result.Type == "epic" && result.Title == "User Authentication Epic" {
					found = true
					break
				}
			}
			assert.True(t, found, "Should find epic with authentication in title using PostgreSQL full-text search")
		})

		t.Run("phrase_search", func(t *testing.T) {
			options := service.SearchOptions{
				Query:     "password reset",
				Limit:     50,
				Offset:    0,
				SortBy:    "created_at",
				SortOrder: "desc",
			}

			response, err := searchService.Search(context.Background(), options)
			require.NoError(t, err)
			assert.NotNil(t, response)

			// PostgreSQL should handle phrase searches better than LIKE
			assert.True(t, len(response.Results) > 0, "Should find results for phrase search")
		})

		t.Run("stemming_and_ranking", func(t *testing.T) {
			// Test PostgreSQL's stemming capabilities
			options := service.SearchOptions{
				Query:     "authenticate", // Should match "authentication"
				Limit:     50,
				Offset:    0,
				SortBy:    "created_at", // Use supported sort field
				SortOrder: "desc",
			}

			response, err := searchService.Search(context.Background(), options)
			require.NoError(t, err)
			assert.NotNil(t, response)

			// Should find results due to stemming
			assert.True(t, len(response.Results) > 0, "Should find results using stemming")
		})

		t.Run("complex_query_with_operators", func(t *testing.T) {
			// Test PostgreSQL-specific query features
			options := service.SearchOptions{
				Query:     "user & login", // AND operator
				Limit:     50,
				Offset:    0,
				SortBy:    "created_at",
				SortOrder: "desc",
			}

			response, err := searchService.Search(context.Background(), options)
			require.NoError(t, err)
			assert.NotNil(t, response)
		})
	})

	t.Run("performance_and_indexing", func(t *testing.T) {
		t.Run("large_dataset_search", func(t *testing.T) {
			// Create a larger dataset for performance testing
			createLargeTestDataset(t, db, user, 100)

			start := time.Now()
			options := service.SearchOptions{
				Query:     "test",
				Limit:     50,
				Offset:    0,
				SortBy:    "created_at",
				SortOrder: "desc",
			}

			response, err := searchService.Search(context.Background(), options)
			duration := time.Since(start)

			require.NoError(t, err)
			assert.NotNil(t, response)

			// Performance assertion - should complete within reasonable time
			assert.Less(t, duration, 5*time.Second, "Search should complete within 5 seconds")

			t.Logf("Search completed in %v with %d results", duration, response.Total)
		})

		t.Run("index_usage_verification", func(t *testing.T) {
			// Verify that PostgreSQL is using full-text search indexes
			var queryPlan string

			// Use EXPLAIN to check if indexes are being used
			err := db.Raw(`
				EXPLAIN (FORMAT TEXT) 
				SELECT * FROM epics 
				WHERE to_tsvector('english', title || ' ' || COALESCE(description, '')) 
				@@ plainto_tsquery('english', 'authentication')
			`).Scan(&queryPlan).Error

			require.NoError(t, err)
			t.Logf("Query plan: %s", queryPlan)

			// The query plan should indicate index usage for optimal performance
			// This is informational for now, but could be made into assertions
		})
	})

	t.Run("advanced_filtering_with_fulltext", func(t *testing.T) {
		t.Run("combined_fulltext_and_filters", func(t *testing.T) {
			priority := int(models.PriorityHigh)
			options := service.SearchOptions{
				Query: "authentication",
				Filters: service.SearchFilters{
					Priority:  &priority,
					CreatorID: &user.ID,
				},
				Limit:     50,
				Offset:    0,
				SortBy:    "created_at",
				SortOrder: "desc",
			}

			response, err := searchService.Search(context.Background(), options)
			require.NoError(t, err)
			assert.NotNil(t, response)

			// Verify results match both full-text search and filters
			for _, result := range response.Results {
				if result.Priority != nil {
					assert.Equal(t, int(models.PriorityHigh), *result.Priority)
				}
				// Should contain search term in title or description
				containsSearchTerm := false
				if result.Title != "" {
					containsSearchTerm = containsSearchTerm || contains(result.Title, "authentication")
				}
				if result.Description != "" {
					containsSearchTerm = containsSearchTerm || contains(result.Description, "authentication")
				}
				assert.True(t, containsSearchTerm, "Result should contain search term")
			}
		})

		t.Run("date_range_filtering", func(t *testing.T) {
			now := time.Now()
			yesterday := now.Add(-24 * time.Hour)

			options := service.SearchOptions{
				Query: "",
				Filters: service.SearchFilters{
					CreatedFrom: &yesterday,
					CreatedTo:   &now,
				},
				Limit:     50,
				Offset:    0,
				SortBy:    "created_at",
				SortOrder: "desc",
			}

			response, err := searchService.Search(context.Background(), options)
			require.NoError(t, err)
			assert.NotNil(t, response)

			// All results should be within the date range
			for _, result := range response.Results {
				assert.True(t, result.CreatedAt.After(yesterday) || result.CreatedAt.Equal(yesterday))
				assert.True(t, result.CreatedAt.Before(now) || result.CreatedAt.Equal(now))
			}
		})
	})

	// Note: Search suggestions functionality not yet implemented
	// This test section is commented out until the GetSearchSuggestions method is added to SearchService
	/*
		t.Run("search_suggestions", func(t *testing.T) {
			t.Run("autocomplete_suggestions", func(t *testing.T) {
				suggestions, err := searchService.GetSearchSuggestions(context.Background(), "auth", 10)
				require.NoError(t, err)
				assert.NotNil(t, suggestions)

				// Should provide relevant suggestions
				found := false
				for _, suggestion := range suggestions {
					if contains(suggestion, "authentication") {
						found = true
						break
					}
				}
				assert.True(t, found, "Should provide authentication as suggestion for 'auth'")
			})

			t.Run("popular_searches", func(t *testing.T) {
				// Test that popular search terms are suggested
				suggestions, err := searchService.GetSearchSuggestions(context.Background(), "", 5)
				require.NoError(t, err)
				assert.NotNil(t, suggestions)
				assert.True(t, len(suggestions) <= 5, "Should respect limit")
			})
		})
	*/
}

func setupPostgreSQLContainer(t *testing.T) *gorm.DB {
	ctx := context.Background()

	// Create PostgreSQL container
	req := testcontainers.ContainerRequest{
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
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get connection details
	host, err := postgresContainer.Host(ctx)
	require.NoError(t, err)

	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Create database connection
	dsn := fmt.Sprintf("host=%s port=%s user=testuser password=password dbname=testdb sslmode=disable", host, port.Port())

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	// Verify connection
	sqlDB, err := db.DB()
	require.NoError(t, err)

	err = sqlDB.Ping()
	require.NoError(t, err)

	// Store container reference for cleanup
	t.Cleanup(func() {
		postgresContainer.Terminate(ctx)
	})

	return db
}

func cleanupDatabase(t *testing.T, db *gorm.DB) {
	if db == nil {
		t.Log("Database connection is nil, skipping cleanup")
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Logf("Failed to get SQL DB for cleanup: %v", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		t.Logf("Failed to close database connection: %v", err)
	} else {
		t.Log("Database connection closed successfully")
	}
}

func createComprehensiveTestData(t *testing.T, db *gorm.DB, user *models.User) map[string]interface{} {
	// Create multiple epics with different content for comprehensive testing
	epics := []*models.Epic{
		{
			ID:          uuid.New(),
			ReferenceID: "EP-001",
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityHigh,
			Status:      models.EpicStatusBacklog,
			Title:       "User Authentication Epic",
			Description: stringPtr("This epic covers all user authentication features including login, logout, and password reset functionality."),
		},
		{
			ID:          uuid.New(),
			ReferenceID: "EP-002",
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityMedium,
			Status:      models.EpicStatusInProgress,
			Title:       "Data Management System",
			Description: stringPtr("Comprehensive data management including CRUD operations, validation, and reporting capabilities."),
		},
		{
			ID:          uuid.New(),
			ReferenceID: "EP-003",
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityLow,
			Status:      models.EpicStatusDone,
			Title:       "User Interface Improvements",
			Description: stringPtr("Enhance user experience through improved UI components and responsive design."),
		},
	}

	for _, epic := range epics {
		err := db.Create(epic).Error
		require.NoError(t, err)
	}

	// Create user stories for each epic
	userStories := []*models.UserStory{
		{
			ID:          uuid.New(),
			ReferenceID: "US-001",
			EpicID:      epics[0].ID,
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityHigh,
			Status:      models.UserStoryStatusDraft,
			Title:       "User Login Feature",
			Description: stringPtr("As a user, I want to login to the system, so that I can access my account and personal data."),
		},
		{
			ID:          uuid.New(),
			ReferenceID: "US-002",
			EpicID:      epics[0].ID,
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityMedium,
			Status:      models.UserStoryStatusInProgress,
			Title:       "Password Reset Functionality",
			Description: stringPtr("As a user, I want to reset my password, so that I can regain access to my account if I forget it."),
		},
		{
			ID:          uuid.New(),
			ReferenceID: "US-003",
			EpicID:      epics[1].ID,
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityHigh,
			Status:      models.UserStoryStatusDone,
			Title:       "Data Export Feature",
			Description: stringPtr("As an admin, I want to export data in various formats, so that I can analyze it externally."),
		},
	}

	for _, us := range userStories {
		err := db.Create(us).Error
		require.NoError(t, err)
	}

	return map[string]interface{}{
		"epics":       epics,
		"userStories": userStories,
	}
}

func createLargeTestDataset(t *testing.T, db *gorm.DB, user *models.User, count int) {
	// Create many entities for performance testing
	for i := 0; i < count; i++ {
		epic := &models.Epic{
			ID:          uuid.New(),
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.Priority((i % 4) + 1),
			Status:      models.EpicStatusBacklog,
			Title:       fmt.Sprintf("Test Epic %d", i),
			Description: stringPtr(fmt.Sprintf("This is test epic number %d with various keywords for testing search performance and functionality.", i)),
		}
		err := db.Create(epic).Error
		require.NoError(t, err)
	}
}

func createTestUser(t *testing.T, db *gorm.DB) *models.User {
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

func stringPtr(s string) *string {
	return &s
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsInMiddle(s, substr))))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
