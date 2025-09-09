package setup

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"

	"product-requirements-management/internal/config"
	"product-requirements-management/internal/models"
)

// BenchmarkServer manages HTTP server instances for benchmark testing
type BenchmarkServer struct {
	Server    *http.Server
	DB        *gorm.DB
	Container testcontainers.Container
	BaseURL   string
	Config    *config.Config
}

// NewBenchmarkServer creates a new benchmark server with PostgreSQL testcontainer
func NewBenchmarkServer(b *testing.B) *BenchmarkServer {
	ctx := context.Background()

	// Create PostgreSQL testcontainer
	container, db, err := setupPostgreSQLContainer(ctx)
	if err != nil {
		b.Fatalf("Failed to setup PostgreSQL container: %v", err)
	}

	// Run migrations on the test database
	if err := models.AutoMigrate(db); err != nil {
		container.Terminate(ctx)
		b.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed default data
	if err := models.SeedDefaultData(db); err != nil {
		container.Terminate(ctx)
		b.Fatalf("Failed to seed default data: %v", err)
	}

	// Create configuration for benchmark
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "0", // Let the system assign a free port
		},
		JWT: config.JWTConfig{
			Secret: "benchmark-test-secret-key-for-testing-only",
		},
		Log: config.LogConfig{
			Level:  "error", // Reduce logging noise during benchmarks
			Format: "json",
		},
	}

	// Set up Gin in release mode for benchmarks
	gin.SetMode(gin.ReleaseMode)

	// Create a simple HTTP server for benchmarks
	router := gin.New()
	router.Use(gin.Recovery())
	
	// Add a basic health endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// TODO: Add actual API routes when implementing specific benchmark tests
	
	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// Get the actual port assigned by the system (for now use a fixed port for simplicity)
	baseURL := "http://localhost:8080"

	return &BenchmarkServer{
		Server:    httpServer,
		DB:        db,
		Container: container,
		BaseURL:   baseURL,
		Config:    cfg,
	}
}

// Start starts the benchmark server
func (bs *BenchmarkServer) Start() error {
	go func() {
		if err := bs.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("Failed to start benchmark server: %v", err))
		}
	}()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)
	return nil
}

// Cleanup stops the server and cleans up resources
func (bs *BenchmarkServer) Cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if bs.Server != nil {
		bs.Server.Shutdown(ctx)
	}

	// Close database connection
	if bs.DB != nil {
		if sqlDB, err := bs.DB.DB(); err == nil {
			sqlDB.Close()
		}
	}

	// Terminate container
	if bs.Container != nil {
		bs.Container.Terminate(context.Background())
	}
}

// SeedData populates the database with test data for benchmarking
func (bs *BenchmarkServer) SeedData(entityCounts map[string]int) error {
	dataGen := NewDataGenerator(bs.DB)
	
	// Convert map to DataSetConfig
	config := DataSetConfig{
		Users:              getIntOrDefault(entityCounts, "users", 10),
		Epics:              getIntOrDefault(entityCounts, "epics", 25),
		UserStoriesPerEpic: getIntOrDefault(entityCounts, "user_stories_per_epic", 4),
		RequirementsPerUS:  getIntOrDefault(entityCounts, "requirements_per_us", 3),
		AcceptanceCriteria: getIntOrDefault(entityCounts, "acceptance_criteria", 50),
		Comments:           getIntOrDefault(entityCounts, "comments", 100),
	}
	
	return dataGen.GenerateFullDataSet(config)
}

// SeedSmallDataSet seeds the database with a small dataset for development
func (bs *BenchmarkServer) SeedSmallDataSet() error {
	dataGen := NewDataGenerator(bs.DB)
	return dataGen.GenerateFullDataSet(GetSmallDataSet())
}

// SeedMediumDataSet seeds the database with a medium dataset for CI/CD
func (bs *BenchmarkServer) SeedMediumDataSet() error {
	dataGen := NewDataGenerator(bs.DB)
	return dataGen.GenerateFullDataSet(GetMediumDataSet())
}

// SeedLargeDataSet seeds the database with a large dataset for performance analysis
func (bs *BenchmarkServer) SeedLargeDataSet() error {
	dataGen := NewDataGenerator(bs.DB)
	return dataGen.GenerateFullDataSet(GetLargeDataSet())
}

// CleanupData removes all test data from the database
func (bs *BenchmarkServer) CleanupData() error {
	dataGen := NewDataGenerator(bs.DB)
	return dataGen.CleanupDatabase()
}

// ResetData drops and recreates all tables, then seeds default data
func (bs *BenchmarkServer) ResetData() error {
	dataGen := NewDataGenerator(bs.DB)
	return dataGen.ResetDatabase()
}

// getIntOrDefault returns the value from map or default if not found
func getIntOrDefault(m map[string]int, key string, defaultValue int) int {
	if val, exists := m[key]; exists {
		return val
	}
	return defaultValue
}

// setupPostgreSQLContainer creates and starts a PostgreSQL testcontainer
func setupPostgreSQLContainer(ctx context.Context) (testcontainers.Container, *gorm.DB, error) {
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
		return nil, nil, fmt.Errorf("failed to start PostgreSQL container: %w", err)
	}

	// Get the mapped port
	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	// Get the host
	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to get container host: %w", err)
	}

	// Create database connection
	dbConfig := config.DatabaseConfig{
		Host:     host,
		Port:     mappedPort.Port(),
		DBName:   "benchmark_test",
		User:     "benchmark_user",
		Password: "benchmark_pass",
		SSLMode:  "disable",
	}

	db, err := initPostgreSQL(dbConfig)
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return container, db, nil
}

