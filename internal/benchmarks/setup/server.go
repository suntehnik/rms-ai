package setup

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"product-requirements-management/internal/config"
	"product-requirements-management/internal/database"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/server/routes"
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

	// Create a router with full application routes
	router := gin.New()
	router.Use(gin.Recovery())

	// Setup application routes with the benchmark database
	dbWrapper := &database.DB{
		Postgres: db,
		Redis:    nil, // No Redis for benchmarks
	}
	routes.Setup(router, cfg, dbWrapper)
	
	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// Use a fixed port for benchmarks
	cfg.Server.Port = "8081" // Use 8081 to avoid conflicts with development server
	httpServer.Addr = fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	baseURL := "http://localhost:8081"

	return &BenchmarkServer{
		Server:    httpServer,
		DB:        db,
		Container: container,
		BaseURL:   baseURL,
		Config:    cfg,
	}
}

// Start starts the benchmark server with enhanced error handling and validation
func (bs *BenchmarkServer) Start() error {
	// Validate server configuration before starting
	if bs.Server == nil {
		return fmt.Errorf("server is not initialized")
	}
	
	if bs.DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	// Test database connectivity before starting server
	if sqlDB, err := bs.DB.DB(); err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	} else if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Start server in a goroutine with enhanced error handling
	serverErrChan := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				serverErrChan <- fmt.Errorf("server panicked: %v", r)
			}
		}()
		
		if err := bs.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrChan <- fmt.Errorf("server failed to start: %w", err)
		}
	}()

	// Wait for server to be ready with comprehensive health checks
	timeout := 30 * time.Second
	checkInterval := 100 * time.Millisecond
	maxAttempts := int(timeout / checkInterval)
	
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for i := 0; i < maxAttempts; i++ {
		// Check if server encountered an error during startup
		select {
		case err := <-serverErrChan:
			return fmt.Errorf("server startup failed: %w", err)
		default:
		}

		// Test health endpoint
		resp, err := client.Get(bs.BaseURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			
			// Additional validation - test basic API endpoint
			apiResp, apiErr := client.Get(bs.BaseURL + "/api/v1")
			if apiResp != nil {
				apiResp.Body.Close()
			}
			
			// Accept 404 for API root as it may not have a handler
			if apiErr == nil && (apiResp.StatusCode == 200 || apiResp.StatusCode == 404) {
				return nil
			}
		}
		
		if resp != nil {
			resp.Body.Close()
		}
		
		time.Sleep(checkInterval)
	}
	
	return fmt.Errorf("server failed to start within %v timeout", timeout)
}

// Cleanup stops the server and cleans up resources with enhanced error handling
func (bs *BenchmarkServer) Cleanup() {
	// Use a longer timeout for cleanup to ensure proper resource cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Cleanup in reverse order of initialization for proper dependency management
	
	// 1. Shutdown HTTP server gracefully
	if bs.Server != nil {
		fmt.Printf("Shutting down HTTP server...\n")
		if err := bs.Server.Shutdown(ctx); err != nil {
			fmt.Printf("HTTP server shutdown error: %v\n", err)
			// Force close if graceful shutdown fails
			if err := bs.Server.Close(); err != nil {
				fmt.Printf("HTTP server force close error: %v\n", err)
			}
		} else {
			fmt.Printf("HTTP server shutdown completed\n")
		}
	}

	// 2. Close database connections
	if bs.DB != nil {
		fmt.Printf("Closing database connections...\n")
		if sqlDB, err := bs.DB.DB(); err == nil {
			// Get connection stats before closing
			stats := sqlDB.Stats()
			fmt.Printf("Database stats before close - Open: %d, InUse: %d\n", 
				stats.OpenConnections, stats.InUse)
			
			// Set connection limits to 0 to force closure of idle connections
			sqlDB.SetMaxOpenConns(0)
			sqlDB.SetMaxIdleConns(0)
			
			if err := sqlDB.Close(); err != nil {
				fmt.Printf("Database close error: %v\n", err)
			} else {
				fmt.Printf("Database connections closed successfully\n")
			}
		} else {
			fmt.Printf("Failed to get underlying database connection: %v\n", err)
		}
	}

	// 3. Terminate container
	if bs.Container != nil {
		fmt.Printf("Terminating test container...\n")
		
		// Use a separate context for container termination
		containerCtx, containerCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer containerCancel()
		
		if err := bs.Container.Terminate(containerCtx); err != nil {
			fmt.Printf("Container termination error: %v\n", err)
			
			// Try to force terminate if graceful termination fails
			if err := bs.Container.Terminate(context.Background()); err != nil {
				fmt.Printf("Container force termination error: %v\n", err)
			}
		} else {
			fmt.Printf("Test container terminated successfully\n")
		}
	}

	// 4. Force garbage collection to clean up any remaining resources
	fmt.Printf("Running garbage collection...\n")
	runtime.GC()
	runtime.GC() // Run twice for better cleanup
	
	fmt.Printf("Benchmark server cleanup completed\n")
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

	// Create database connection directly for benchmarks
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		host, "benchmark_user", "benchmark_pass", "benchmark_test", mappedPort.Port(), "disable")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent for benchmarks
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return container, db, nil
}

