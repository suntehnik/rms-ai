package setup

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"

	"product-requirements-management/internal/config"
	"product-requirements-management/internal/database"
)

// DatabaseContainer wraps testcontainer functionality for PostgreSQL
type DatabaseContainer struct {
	Container testcontainers.Container
	DB        *gorm.DB
	Config    *config.DatabaseConfig
}

// NewPostgreSQLContainer creates a new PostgreSQL testcontainer for benchmarks
func NewPostgreSQLContainer(ctx context.Context) (*DatabaseContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:12-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "benchmark_test",
			"POSTGRES_USER":     "benchmark_user",
			"POSTGRES_PASSWORD": "benchmark_pass",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start PostgreSQL container: %w", err)
	}

	// Get connection details
	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	// Create database configuration
	dbConfig := &config.DatabaseConfig{
		Host:     host,
		Port:     mappedPort.Port(),
		Database: "benchmark_test",
		Username: "benchmark_user",
		Password: "benchmark_pass",
	}

	// Connect to database
	db, err := database.NewConnection(dbConfig)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &DatabaseContainer{
		Container: container,
		DB:        db,
		Config:    dbConfig,
	}, nil
}

// Cleanup terminates the container and closes database connections
func (dc *DatabaseContainer) Cleanup(ctx context.Context) error {
	// Close database connection
	if dc.DB != nil {
		if sqlDB, err := dc.DB.DB(); err == nil {
			sqlDB.Close()
		}
	}

	// Terminate container
	if dc.Container != nil {
		return dc.Container.Terminate(ctx)
	}

	return nil
}

// ResetDatabase clears all data and re-runs migrations
func (dc *DatabaseContainer) ResetDatabase() error {
	// Drop all tables
	if err := dc.DB.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;").Error; err != nil {
		return fmt.Errorf("failed to reset database schema: %w", err)
	}

	// Re-run migrations
	return database.RunMigrations(dc.DB)
}

// GetConnectionString returns the database connection string
func (dc *DatabaseContainer) GetConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dc.Config.Username,
		dc.Config.Password,
		dc.Config.Host,
		dc.Config.Port,
		dc.Config.Database,
	)
}