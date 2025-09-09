package setup

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"product-requirements-management/internal/config"
	"product-requirements-management/internal/models"
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
		DBName:   "benchmark_test",
		User:     "benchmark_user",
		Password: "benchmark_pass",
		SSLMode:  "disable",
	}

	// Connect to database
	db, err := initPostgreSQL(*dbConfig)
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
	if err := models.AutoMigrate(dc.DB); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Seed default data
	return models.SeedDefaultData(dc.DB)
}

// SeedTestData populates the database with test data using the data generator
func (dc *DatabaseContainer) SeedTestData(config DataSetConfig) error {
	dataGen := NewDataGenerator(dc.DB)
	return dataGen.GenerateFullDataSet(config)
}

// CleanupTestData removes all test data from the database
func (dc *DatabaseContainer) CleanupTestData() error {
	dataGen := NewDataGenerator(dc.DB)
	return dataGen.CleanupDatabase()
}

// GetConnectionString returns the database connection string
func (dc *DatabaseContainer) GetConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dc.Config.User,
		dc.Config.Password,
		dc.Config.Host,
		dc.Config.Port,
		dc.Config.DBName,
	)
}

// initPostgreSQL initializes PostgreSQL connection with GORM for benchmarks
func initPostgreSQL(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent for benchmarks
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Configure connection pool for benchmarks
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings optimized for benchmarks
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}