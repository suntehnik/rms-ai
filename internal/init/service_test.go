package init

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"product-requirements-management/internal/config"
	"product-requirements-management/internal/logger"
	"product-requirements-management/internal/models"
)

func TestMain(m *testing.M) {
	// Initialize logger for tests
	logger.Init(&config.LogConfig{
		Level:  "error", // Reduce noise in tests
		Format: "text",
	})

	code := m.Run()
	os.Exit(code)
}

func setupTestConfig() *config.Config {
	return &config.Config{
		Database: config.DatabaseConfig{
			Host:    "localhost",
			Port:    "5432",
			User:    "test_user",
			DBName:  "test_db",
			SSLMode: "disable",
		},
		JWT: config.JWTConfig{
			Secret: "test-jwt-secret-key",
		},
		Log: config.LogConfig{
			Level:  "error",
			Format: "text",
		},
	}
}

func setupServiceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate test models
	err = db.AutoMigrate(
		&models.User{},
		&models.Epic{},
		&models.UserStory{},
		&models.Requirement{},
		&models.AcceptanceCriteria{},
		&models.Comment{},
	)
	require.NoError(t, err)

	return db
}

func TestNewInitService(t *testing.T) {
	cfg := setupTestConfig()

	service, err := NewInitService(cfg)

	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, cfg, service.cfg)
	assert.NotNil(t, service.auth)
	assert.False(t, service.startTime.IsZero())
}

func TestInitService_validateEnvironment(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    func()
		cleanupEnv  func()
		config      *config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid environment",
			setupEnv: func() {
				os.Setenv("DEFAULT_ADMIN_PASSWORD", "test-password-123")
			},
			cleanupEnv: func() {
				os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
			},
			config:      setupTestConfig(),
			expectError: false,
		},
		{
			name: "missing admin password",
			setupEnv: func() {
				os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
			},
			cleanupEnv:  func() {},
			config:      setupTestConfig(),
			expectError: true,
			errorMsg:    "missing required environment variables: [DEFAULT_ADMIN_PASSWORD]",
		},
		{
			name: "short admin password",
			setupEnv: func() {
				os.Setenv("DEFAULT_ADMIN_PASSWORD", "short")
			},
			cleanupEnv: func() {
				os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
			},
			config:      setupTestConfig(),
			expectError: true,
			errorMsg:    "invalid environment variables: [DEFAULT_ADMIN_PASSWORD (must be at least 8 characters)]",
		},
		{
			name: "missing database host",
			setupEnv: func() {
				os.Setenv("DEFAULT_ADMIN_PASSWORD", "test-password-123")
			},
			cleanupEnv: func() {
				os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
			},
			config: &config.Config{
				Database: config.DatabaseConfig{
					Host:    "", // Missing host
					Port:    "5432",
					User:    "test_user",
					DBName:  "test_db",
					SSLMode: "disable",
				},
				JWT: config.JWTConfig{
					Secret: "test-jwt-secret-key",
				},
			},
			expectError: true,
			errorMsg:    "missing required environment variables: [DB_HOST]",
		},
		{
			name: "default JWT secret",
			setupEnv: func() {
				os.Setenv("DEFAULT_ADMIN_PASSWORD", "test-password-123")
			},
			cleanupEnv: func() {
				os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
			},
			config: &config.Config{
				Database: config.DatabaseConfig{
					Host:    "localhost",
					Port:    "5432",
					User:    "test_user",
					DBName:  "test_db",
					SSLMode: "disable",
				},
				JWT: config.JWTConfig{
					Secret: "your-secret-key", // Default value
				},
			},
			expectError: true,
			errorMsg:    "missing required environment variables: [JWT_SECRET]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			service := &InitService{
				cfg:       tt.config,
				startTime: time.Now(),
			}

			err := service.validateEnvironment()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInitService_performSafetyCheck(t *testing.T) {
	tests := []struct {
		name        string
		setupData   func(*gorm.DB)
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty database",
			setupData:   func(db *gorm.DB) {}, // No data
			expectError: false,
		},
		{
			name: "database with users",
			setupData: func(db *gorm.DB) {
				user := &models.User{
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: "hashed",
					Role:         models.RoleUser,
				}
				db.Create(user)
			},
			expectError: true,
			errorMsg:    "database safety check failed",
		},
		{
			name: "database with epics",
			setupData: func(db *gorm.DB) {
				description := "Test Description"
				epic := &models.Epic{
					Title:       "Test Epic",
					Description: &description,
				}
				db.Create(epic)
			},
			expectError: true,
			errorMsg:    "database safety check failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupServiceTestDB(t)
			tt.setupData(db)

			service := &InitService{
				db:            db,
				safetyChecker: NewSafetyChecker(db),
			}

			err := service.performSafetyCheck()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInitService_checkDatabaseHealth(t *testing.T) {
	tests := []struct {
		name        string
		setupDB     func() *gorm.DB
		expectError bool
		errorMsg    string
	}{
		{
			name: "healthy database",
			setupDB: func() *gorm.DB {
				return setupServiceTestDB(t)
			},
			expectError: false,
		},
		{
			name: "closed database",
			setupDB: func() *gorm.DB {
				db := setupServiceTestDB(t)
				sqlDB, _ := db.DB()
				sqlDB.Close()
				return db
			},
			expectError: true,
			errorMsg:    "database ping failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setupDB()

			service := &InitService{
				db: db,
			}

			err := service.checkDatabaseHealth()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInitService_createAdminUser(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    func()
		cleanupEnv  func()
		setupDB     func(*gorm.DB)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful admin user creation",
			setupEnv: func() {
				os.Setenv("DEFAULT_ADMIN_PASSWORD", "test-admin-password")
			},
			cleanupEnv: func() {
				os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
			},
			setupDB:     func(db *gorm.DB) {}, // Empty database
			expectError: false,
		},
		{
			name: "missing admin password",
			setupEnv: func() {
				os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
			},
			cleanupEnv:  func() {},
			setupDB:     func(db *gorm.DB) {},
			expectError: true,
			errorMsg:    "DEFAULT_ADMIN_PASSWORD environment variable is required",
		},
		{
			name: "duplicate admin user",
			setupEnv: func() {
				os.Setenv("DEFAULT_ADMIN_PASSWORD", "test-admin-password")
			},
			cleanupEnv: func() {
				os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
			},
			setupDB: func(db *gorm.DB) {
				// Create existing admin user
				existingAdmin := &models.User{
					Username:     "admin",
					Email:        "admin@localhost",
					PasswordHash: "existing-hash",
					Role:         models.RoleAdministrator,
				}
				db.Create(existingAdmin)
			},
			expectError: true,
			errorMsg:    "failed to create admin user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			db := setupServiceTestDB(t)
			tt.setupDB(db)

			cfg := setupTestConfig()
			service, err := NewInitService(cfg)
			require.NoError(t, err)
			service.db = db
			// Initialize adminCreator since we're bypassing connectDatabase
			service.adminCreator = NewAdminCreator(service.db, service.auth)

			err = service.createAdminUser()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)

				// Verify admin user was created
				var adminUser models.User
				err = db.Where("username = ?", "admin").First(&adminUser).Error
				assert.NoError(t, err)
				assert.Equal(t, "admin", adminUser.Username)
				assert.Equal(t, models.RoleAdministrator, adminUser.Role)
				assert.NotEmpty(t, adminUser.PasswordHash)
			}
		})
	}
}

func TestInitError_Error(t *testing.T) {
	tests := []struct {
		name     string
		initErr  *InitError
		expected string
	}{
		{
			name: "error with cause",
			initErr: &InitError{
				Type:    ErrorTypeConfig,
				Message: "Configuration failed",
				Cause:   assert.AnError,
			},
			expected: "configuration: Configuration failed (caused by: assert.AnError general error for testing)",
		},
		{
			name: "error without cause",
			initErr: &InitError{
				Type:    ErrorTypeDatabase,
				Message: "Database connection failed",
			},
			expected: "database: Database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.initErr.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInitService_Close(t *testing.T) {
	tests := []struct {
		name        string
		setupDB     func() *gorm.DB
		expectError bool
	}{
		{
			name: "successful close",
			setupDB: func() *gorm.DB {
				return setupServiceTestDB(t)
			},
			expectError: false,
		},
		{
			name: "nil database",
			setupDB: func() *gorm.DB {
				return nil
			},
			expectError: false,
		},
		{
			name: "already closed database",
			setupDB: func() *gorm.DB {
				db := setupServiceTestDB(t)
				sqlDB, _ := db.DB()
				sqlDB.Close()
				return db
			},
			expectError: false, // SQLite doesn't error on double close
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &InitService{
				db: tt.setupDB(),
			}

			err := service.Close()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Integration test for the complete initialization flow (without actual database connection)
func TestInitService_Initialize_ValidationSteps(t *testing.T) {
	// Test that Initialize calls validation steps in correct order
	// This test focuses on the orchestration logic rather than individual step implementation

	cfg := setupTestConfig()

	// Test missing admin password scenario
	t.Run("fails on environment validation", func(t *testing.T) {
		os.Unsetenv("DEFAULT_ADMIN_PASSWORD")

		service, err := NewInitService(cfg)
		require.NoError(t, err)

		err = service.Initialize()

		assert.Error(t, err)
		var initErr *InitError
		assert.ErrorAs(t, err, &initErr)
		assert.Equal(t, ErrorTypeConfig, initErr.Type)
		assert.Contains(t, initErr.Message, "Environment validation failed")
	})

	// Test successful environment validation but database connection failure
	t.Run("fails on database connection", func(t *testing.T) {
		os.Setenv("DEFAULT_ADMIN_PASSWORD", "test-password-123")
		defer os.Unsetenv("DEFAULT_ADMIN_PASSWORD")

		// Use invalid database config to force connection failure
		invalidCfg := &config.Config{
			Database: config.DatabaseConfig{
				Host:    "invalid-host-that-does-not-exist",
				Port:    "9999",
				User:    "invalid_user",
				DBName:  "invalid_db",
				SSLMode: "disable",
			},
			JWT: config.JWTConfig{
				Secret: "test-jwt-secret-key",
			},
		}

		service, err := NewInitService(invalidCfg)
		require.NoError(t, err)

		err = service.Initialize()

		assert.Error(t, err)
		var initErr *InitError
		assert.ErrorAs(t, err, &initErr)
		assert.Equal(t, ErrorTypeDatabase, initErr.Type)
		assert.Contains(t, initErr.Message, "Database connection failed")
	})
}

// Benchmark test for initialization service creation
func BenchmarkNewInitService(b *testing.B) {
	cfg := setupTestConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service, err := NewInitService(cfg)
		if err != nil {
			b.Fatal(err)
		}
		_ = service
	}
}

// Test helper functions
func TestInitService_logSuccessAndNextSteps(t *testing.T) {
	service := &InitService{
		startTime: time.Now().Add(-5 * time.Second), // Simulate 5 seconds of execution
	}

	// This test just ensures the method doesn't panic
	// In a real scenario, you might want to capture log output
	assert.NotPanics(t, func() {
		service.logSuccessAndNextSteps()
	})
}
