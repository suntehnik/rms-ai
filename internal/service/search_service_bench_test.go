package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

// BenchmarkSearchService_Search benchmarks the search functionality
func BenchmarkSearchService_Search(b *testing.B) {
	// Setup in-memory database for benchmarking
	db := setupBenchmarkDB(b)
	defer cleanupBenchmarkDB(db)

	// Setup repositories
	repos := repository.NewRepositories(db)

	// Setup search service
	searchService := NewSearchService(
		db,
		nil, // No Redis for benchmarks
		repos.Epic,
		repos.UserStory,
		repos.AcceptanceCriteria,
		repos.Requirement,
	)

	// Create test data
	user := createBenchmarkUser(b, db)
	createBenchmarkData(b, db, user, 100) // Create 100 entities for meaningful benchmarks

	// Benchmark different search scenarios
	b.Run("simple_word_search", func(b *testing.B) {
		options := SearchOptions{
			Query:     "test",
			Limit:     50,
			Offset:    0,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := searchService.Search(context.Background(), options)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("filtered_search", func(b *testing.B) {
		priority := int(models.PriorityHigh)
		options := SearchOptions{
			Query: "authentication",
			Filters: SearchFilters{
				Priority:  &priority,
				CreatorID: &user.ID,
			},
			Limit:     50,
			Offset:    0,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := searchService.Search(context.Background(), options)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("large_result_set", func(b *testing.B) {
		options := SearchOptions{
			Query:     "", // Empty query to get all results
			Limit:     100,
			Offset:    0,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := searchService.Search(context.Background(), options)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkSearchService_SearchWithPagination benchmarks search with different pagination scenarios
func BenchmarkSearchService_SearchWithPagination(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer cleanupBenchmarkDB(db)

	repos := repository.NewRepositories(db)
	searchService := NewSearchService(db, nil, repos.Epic, repos.UserStory, repos.AcceptanceCriteria, repos.Requirement)

	user := createBenchmarkUser(b, db)
	createBenchmarkData(b, db, user, 500) // Larger dataset for pagination testing

	b.Run("first_page", func(b *testing.B) {
		options := SearchOptions{
			Query:     "test",
			Limit:     20,
			Offset:    0,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := searchService.Search(context.Background(), options)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("middle_page", func(b *testing.B) {
		options := SearchOptions{
			Query:     "test",
			Limit:     20,
			Offset:    100,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := searchService.Search(context.Background(), options)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("last_page", func(b *testing.B) {
		options := SearchOptions{
			Query:     "test",
			Limit:     20,
			Offset:    480,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := searchService.Search(context.Background(), options)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func setupBenchmarkDB(b *testing.B) *gorm.DB {
	ctx := context.Background()

	// Create PostgreSQL container for benchmarks
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "benchdb",
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_USER":     "benchuser",
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
	require.NoError(b, err)

	// Get connection details
	host, err := postgresContainer.Host(ctx)
	require.NoError(b, err)

	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(b, err)

	// Create database connection
	dsn := fmt.Sprintf("host=%s port=%s user=benchuser password=password dbname=benchdb sslmode=disable", host, port.Port())
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(b, err)

	// Verify connection
	sqlDB, err := db.DB()
	require.NoError(b, err)
	
	err = sqlDB.Ping()
	require.NoError(b, err)

	// Auto-migrate models
	err = models.AutoMigrate(db)
	require.NoError(b, err)

	// Seed default data
	err = models.SeedDefaultData(db)
	require.NoError(b, err)

	// Store container reference for cleanup
	b.Cleanup(func() {
		postgresContainer.Terminate(ctx)
	})

	return db
}

func cleanupBenchmarkDB(db *gorm.DB) {
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
}

func createBenchmarkUser(b *testing.B, db *gorm.DB) *models.User {
	user := &models.User{
		ID:           uuid.New(),
		Username:     "benchuser",
		Email:        "bench@example.com",
		PasswordHash: "hashedpassword",
		Role:         models.RoleUser,
	}
	err := db.Create(user).Error
	require.NoError(b, err)
	return user
}

func createBenchmarkData(b *testing.B, db *gorm.DB, user *models.User, count int) {
	// Create epics
	for i := 0; i < count/4; i++ {
		epic := &models.Epic{
			ID:          uuid.New(),
			ReferenceID: fmt.Sprintf("EP-BENCH-%03d", i+1),
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.Priority((i%4)+1),
			Status:      models.EpicStatusBacklog,
			Title:       fmt.Sprintf("Test Epic %d for Authentication", i+1),
			Description: benchStringPtr(fmt.Sprintf("This is benchmark test epic number %d with authentication and various keywords for testing search performance.", i+1)),
		}
		err := db.Create(epic).Error
		require.NoError(b, err)

		// Create user stories for each epic
		for j := 0; j < 2; j++ {
			userStory := &models.UserStory{
				ID:          uuid.New(),
				ReferenceID: fmt.Sprintf("US-BENCH-%03d-%d", i+1, j+1),
				EpicID:      epic.ID,
				CreatorID:   user.ID,
				AssigneeID:  user.ID,
				Priority:    models.Priority((j%4)+1),
				Status:      models.UserStoryStatusBacklog,
				Title:       fmt.Sprintf("Test User Story %d-%d Authentication Feature", i+1, j+1),
				Description: benchStringPtr(fmt.Sprintf("As a user, I want to test authentication feature %d-%d, so that I can benchmark search performance.", i+1, j+1)),
			}
			err := db.Create(userStory).Error
			require.NoError(b, err)
		}
	}
}

func benchStringPtr(s string) *string {
	return &s
}