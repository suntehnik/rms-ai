package repository

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
)

// BenchmarkEpicRepository_Create benchmarks epic creation
func BenchmarkEpicRepository_Create(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer cleanupBenchmarkDB(db)

	repo := NewEpicRepository(db)
	user := createBenchmarkUser(b, db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		epic := &models.Epic{
			ID:          uuid.New(),
			ReferenceID: fmt.Sprintf("EP-BENCH-%d", i),
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityHigh,
			Status:      models.EpicStatusBacklog,
			Title:       fmt.Sprintf("Benchmark Epic %d", i),
			Description: benchStringPtr(fmt.Sprintf("Benchmark description %d", i)),
		}

		err := repo.Create(epic)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEpicRepository_GetByID benchmarks epic retrieval by ID
func BenchmarkEpicRepository_GetByID(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer cleanupBenchmarkDB(db)

	repo := NewEpicRepository(db)
	user := createBenchmarkUser(b, db)

	// Create test epics
	var epicIDs []uuid.UUID
	for i := 0; i < 100; i++ {
		epic := &models.Epic{
			ID:          uuid.New(),
			ReferenceID: fmt.Sprintf("EP-BENCH-%d", i),
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityHigh,
			Status:      models.EpicStatusBacklog,
			Title:       fmt.Sprintf("Benchmark Epic %d", i),
			Description: benchStringPtr(fmt.Sprintf("Benchmark description %d", i)),
		}
		err := repo.Create(epic)
		require.NoError(b, err)
		epicIDs = append(epicIDs, epic.ID)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := epicIDs[i%len(epicIDs)]
		_, err := repo.GetByID(id)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEpicRepository_List benchmarks epic listing with filters
func BenchmarkEpicRepository_List(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer cleanupBenchmarkDB(db)

	repo := NewEpicRepository(db)
	user := createBenchmarkUser(b, db)

	// Create test epics
	for i := 0; i < 200; i++ {
		epic := &models.Epic{
			ID:          uuid.New(),
			ReferenceID: fmt.Sprintf("EP-BENCH-%d", i),
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.Priority((i % 4) + 1),
			Status:      models.EpicStatusBacklog,
			Title:       fmt.Sprintf("Benchmark Epic %d", i),
			Description: benchStringPtr(fmt.Sprintf("Benchmark description %d", i)),
		}
		err := repo.Create(epic)
		require.NoError(b, err)
	}

	b.Run("no_filters", func(b *testing.B) {
		filters := make(map[string]interface{})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := repo.List(filters, "created_at", 100, 0)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("with_creator_filter", func(b *testing.B) {
		filters := map[string]interface{}{
			"creator_id": user.ID,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := repo.List(filters, "created_at", 100, 0)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("with_priority_filter", func(b *testing.B) {
		filters := map[string]interface{}{
			"priority": models.PriorityHigh,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := repo.List(filters, "created_at", 100, 0)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkEpicRepository_Update benchmarks epic updates
func BenchmarkEpicRepository_Update(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer cleanupBenchmarkDB(db)

	repo := NewEpicRepository(db)
	user := createBenchmarkUser(b, db)

	// Create test epics
	var epics []*models.Epic
	for i := 0; i < 50; i++ {
		epic := &models.Epic{
			ID:          uuid.New(),
			ReferenceID: fmt.Sprintf("EP-BENCH-%d", i),
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityHigh,
			Status:      models.EpicStatusBacklog,
			Title:       fmt.Sprintf("Benchmark Epic %d", i),
			Description: benchStringPtr(fmt.Sprintf("Benchmark description %d", i)),
		}
		err := repo.Create(epic)
		require.NoError(b, err)
		epics = append(epics, epic)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		epic := epics[i%len(epics)]
		epic.Title = fmt.Sprintf("Updated Benchmark Epic %d", i)
		err := repo.Update(epic)
		if err != nil {
			b.Fatal(err)
		}
	}
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

func benchStringPtr(s string) *string {
	return &s
}
