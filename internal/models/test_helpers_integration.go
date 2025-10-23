package models

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgresContainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// testDatabase represents a test PostgreSQL database
type testDatabase struct {
	db        *gorm.DB
	container *postgresContainer.PostgresContainer
	dsn       string
}

// setupPostgreSQLWithMigrations creates a new PostgreSQL container with SQL migrations
func setupPostgreSQLWithMigrations(t *testing.T) *testDatabase {
	ctx := context.Background()

	// Create PostgreSQL container
	container, err := postgresContainer.Run(ctx,
		"postgres:15-alpine",
		postgresContainer.WithDatabase("testdb"),
		postgresContainer.WithUsername("testuser"),
		postgresContainer.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)

	// Get connection string
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Disable logs for tests
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	require.NoError(t, err)

	testDB := &testDatabase{
		db:        db,
		container: container,
		dsn:       dsn,
	}

	// Run SQL migrations
	err = testDB.runSQLMigrations()
	require.NoError(t, err)

	return testDB
}

// cleanup stops and removes the PostgreSQL container
func (td *testDatabase) cleanup(t *testing.T) {
	// Close database connection first
	if td.db != nil {
		if sqlDB, err := td.db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				t.Logf("Failed to close database connection: %v", err)
			}
		}
	}

	// Terminate container
	if td.container != nil {
		ctx := context.Background()
		if err := td.container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate PostgreSQL container: %v", err)
		}
	}
}

// runSQLMigrations executes SQL migrations for the test database
func (td *testDatabase) runSQLMigrations() error {
	// Get absolute path to migrations directory relative to project root
	migrationsDir := "../../migrations"
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	// Check that migrations directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory does not exist: %s", absPath)
	}

	// Create migrator
	migrator, err := migrate.New(
		fmt.Sprintf("file://%s", absPath),
		td.dsn,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	// Run migrations
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// After running SQL migrations, seed default data
	if err := SeedDefaultData(td.db); err != nil {
		return fmt.Errorf("failed to seed default data: %w", err)
	}

	// Reset sequences to ensure tests start with predictable reference IDs
	if err := td.resetSequences(); err != nil {
		return fmt.Errorf("failed to reset sequences: %w", err)
	}

	return nil
}

// resetSequences resets all reference ID sequences to start from 1
func (td *testDatabase) resetSequences() error {
	sequences := []string{
		"epic_ref_seq",
		"user_story_ref_seq",
		"acceptance_criteria_ref_seq",
		"requirement_ref_seq",
		"steering_document_ref_seq",
	}

	for _, seq := range sequences {
		if err := td.db.Exec(fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH 1", seq)).Error; err != nil {
			return fmt.Errorf("failed to reset sequence %s: %w", seq, err)
		}
	}

	return nil
}
