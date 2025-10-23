package init

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgresContainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"product-requirements-management/internal/config"
	"product-requirements-management/internal/database"
	"product-requirements-management/internal/models"
)

// TestDatabase represents a test PostgreSQL database for integration tests
type TestDatabase struct {
	DB        *gorm.DB
	Container *postgresContainer.PostgresContainer
	DSN       string
	Config    *config.Config
}

// setupTestDatabase creates a new PostgreSQL container for integration tests
func setupTestDatabase(t *testing.T) *TestDatabase {
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

	// Create config for the test database
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost", // Will be overridden by DSN
			Port:     "5432",
			User:     "testuser",
			Password: "testpass",
			DBName:   "testdb",
			SSLMode:  "disable",
		},
		JWT: config.JWTConfig{
			Secret: "test-jwt-secret-for-integration-tests",
		},
		Log: config.LogConfig{
			Level:  "error", // Reduce log noise in tests
			Format: "json",
		},
	}

	return &TestDatabase{
		DB:        db,
		Container: container,
		DSN:       dsn,
		Config:    cfg,
	}
}

// cleanup stops and removes the PostgreSQL container
func (td *TestDatabase) cleanup(t *testing.T) {
	if td.Container != nil {
		ctx := context.Background()
		if err := td.Container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate PostgreSQL container: %v", err)
		}
	}
}

// reset clears all data in the database for test isolation
func (td *TestDatabase) reset() error {
	// List of tables in dependency order (dependent tables first)
	tables := []string{
		"comments",
		"requirement_relationships",
		"requirements",
		"acceptance_criteria",
		"user_stories",
		"epics",
		"users",
		"schema_migrations", // Also clear migration state for clean tests
	}

	// Disable foreign key checks
	if err := td.DB.Exec("SET session_replication_role = replica").Error; err != nil {
		return err
	}

	// Clear tables
	for _, table := range tables {
		// Check if table exists first
		var exists bool
		err := td.DB.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = ?)", table).Scan(&exists).Error
		if err != nil {
			continue // Skip if we can't check
		}
		if exists {
			if err := td.DB.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE").Error; err != nil {
				// Continue with other tables even if one fails
				continue
			}
		}
	}

	// Re-enable foreign key checks
	if err := td.DB.Exec("SET session_replication_role = DEFAULT").Error; err != nil {
		return err
	}

	return nil
}

// runSQLMigrations executes SQL migrations for the test database
func (td *TestDatabase) runSQLMigrations() error {
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
		td.DSN,
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
	if err := models.SeedDefaultData(td.DB); err != nil {
		return fmt.Errorf("failed to seed default data: %w", err)
	}

	return nil
}

// createTestDataForSafetyCheck creates test data to make database non-empty
func (td *TestDatabase) createTestDataForSafetyCheck(t *testing.T) {
	// First run SQL migrations to create tables with proper PostgreSQL functions
	if err := td.runSQLMigrations(); err != nil {
		require.NoError(t, err)
	}

	// Create a test user to make database non-empty
	user := &models.User{
		Username:     "existing_user",
		Email:        "existing@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}

	err := td.DB.Create(user).Error
	require.NoError(t, err)
}

func TestInitService_CompleteInitializationFlow_EmptyDatabase(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Set required environment variable
	originalPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	defer func() {
		if originalPassword != "" {
			os.Setenv("DEFAULT_ADMIN_PASSWORD", originalPassword)
		} else {
			os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
		}
	}()
	os.Setenv("DEFAULT_ADMIN_PASSWORD", "test-admin-password-123")

	// Create initialization service
	service, err := NewInitService(testDB.Config)
	require.NoError(t, err)
	defer service.Close()

	// Override database connection to use test database and initialize components
	service.db = testDB.DB
	service.safetyChecker = NewSafetyChecker(service.db)
	service.migrator = database.NewMigrationManager(service.db, "migrations")
	service.adminCreator = NewAdminCreator(service.db, service.auth)

	// Run complete initialization
	err = service.Initialize()
	assert.NoError(t, err)

	// Verify migrations were applied
	var migrationCount int64
	err = testDB.DB.Table("schema_migrations").Count(&migrationCount).Error
	assert.NoError(t, err)
	assert.Greater(t, migrationCount, int64(0), "Migrations should have been applied")

	// Verify admin user was created
	var adminUser models.User
	err = testDB.DB.Where("username = ?", "admin").First(&adminUser).Error
	assert.NoError(t, err)
	assert.Equal(t, "admin", adminUser.Username)
	assert.Equal(t, "admin@localhost", adminUser.Email)
	assert.Equal(t, models.RoleAdministrator, adminUser.Role)
	assert.NotEmpty(t, adminUser.PasswordHash)

	// Verify database tables exist
	tables := []string{"users", "epics", "user_stories", "requirements", "acceptance_criteria", "comments"}
	for _, table := range tables {
		var exists bool
		err = testDB.DB.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = ?)", table).Scan(&exists).Error
		assert.NoError(t, err)
		assert.True(t, exists, "Table %s should exist after initialization", table)
	}

	// Verify reference types were seeded
	var reqTypeCount int64
	err = testDB.DB.Table("requirement_types").Count(&reqTypeCount).Error
	assert.NoError(t, err)
	assert.Greater(t, reqTypeCount, int64(0), "Requirement types should be seeded")

	var relTypeCount int64
	err = testDB.DB.Table("relationship_types").Count(&relTypeCount).Error
	assert.NoError(t, err)
	assert.Greater(t, relTypeCount, int64(0), "Relationship types should be seeded")
}

func TestInitService_SafetyCheck_PreventInitializationOnNonEmptyDatabase(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Create test data to make database non-empty
	testDB.createTestDataForSafetyCheck(t)

	// Set required environment variable
	originalPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	defer func() {
		if originalPassword != "" {
			os.Setenv("DEFAULT_ADMIN_PASSWORD", originalPassword)
		} else {
			os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
		}
	}()
	os.Setenv("DEFAULT_ADMIN_PASSWORD", "test-admin-password-123")

	// Create initialization service
	service, err := NewInitService(testDB.Config)
	require.NoError(t, err)
	defer service.Close()

	// Override database connection to use test database and initialize components
	service.db = testDB.DB
	service.safetyChecker = NewSafetyChecker(service.db)
	service.migrator = database.NewMigrationManager(service.db, "migrations")
	service.adminCreator = NewAdminCreator(service.db, service.auth)

	// Attempt initialization - should fail due to safety check
	err = service.Initialize()
	assert.Error(t, err)

	// Verify error is of safety type
	var initErr *InitError
	assert.ErrorAs(t, err, &initErr)
	assert.Equal(t, ErrorTypeSafety, initErr.Type)
	assert.Contains(t, initErr.Message, "Database safety check failed")

	// Verify no additional admin user was created
	var userCount int64
	err = testDB.DB.Table("users").Count(&userCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), userCount, "Should still have only the original test user")

	// Verify no admin user exists
	var adminUser models.User
	err = testDB.DB.Where("username = ?", "admin").First(&adminUser).Error
	assert.Error(t, err) // Should not find admin user
}

func TestInitService_PartialFailure_MissingEnvironmentVariable(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Ensure DEFAULT_ADMIN_PASSWORD is not set
	originalPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	defer func() {
		if originalPassword != "" {
			os.Setenv("DEFAULT_ADMIN_PASSWORD", originalPassword)
		}
	}()
	os.Unsetenv("DEFAULT_ADMIN_PASSWORD")

	// Create initialization service
	service, err := NewInitService(testDB.Config)
	require.NoError(t, err)
	defer service.Close()

	// Override database connection to use test database and initialize components
	service.db = testDB.DB
	service.safetyChecker = NewSafetyChecker(service.db)
	service.migrator = database.NewMigrationManager(service.db, "migrations")
	service.adminCreator = NewAdminCreator(service.db, service.auth)

	// Attempt initialization - should fail during environment validation
	err = service.Initialize()
	assert.Error(t, err)

	// Verify error is of configuration type
	var initErr *InitError
	assert.ErrorAs(t, err, &initErr)
	assert.Equal(t, ErrorTypeConfig, initErr.Type)
	assert.Contains(t, initErr.Message, "Environment validation failed")

	// Verify no admin user was created
	var userCount int64
	err = testDB.DB.Table("users").Count(&userCount).Error
	if err == nil { // Table might not exist yet
		assert.Equal(t, int64(0), userCount, "No users should be created on environment validation failure")
	}
}

func TestInitService_PartialFailure_InvalidJWTSecret(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Set invalid JWT secret (default value)
	testDB.Config.JWT.Secret = "your-secret-key"

	// Set required environment variable
	originalPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	defer func() {
		if originalPassword != "" {
			os.Setenv("DEFAULT_ADMIN_PASSWORD", originalPassword)
		} else {
			os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
		}
	}()
	os.Setenv("DEFAULT_ADMIN_PASSWORD", "test-admin-password-123")

	// Create initialization service
	service, err := NewInitService(testDB.Config)
	require.NoError(t, err)
	defer service.Close()

	// Override database connection to use test database and initialize components
	service.db = testDB.DB
	service.safetyChecker = NewSafetyChecker(service.db)
	service.migrator = database.NewMigrationManager(service.db, "migrations")
	service.adminCreator = NewAdminCreator(service.db, service.auth)

	// Attempt initialization - should fail during environment validation
	err = service.Initialize()
	assert.Error(t, err)

	// Verify error is of configuration type
	var initErr *InitError
	assert.ErrorAs(t, err, &initErr)
	assert.Equal(t, ErrorTypeConfig, initErr.Type)
	assert.Contains(t, initErr.Message, "Environment validation failed")
}

func TestInitService_PartialFailure_WeakPassword(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Set weak password (less than 8 characters)
	originalPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	defer func() {
		if originalPassword != "" {
			os.Setenv("DEFAULT_ADMIN_PASSWORD", originalPassword)
		} else {
			os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
		}
	}()
	os.Setenv("DEFAULT_ADMIN_PASSWORD", "weak")

	// Create initialization service
	service, err := NewInitService(testDB.Config)
	require.NoError(t, err)
	defer service.Close()

	// Override database connection to use test database and initialize components
	service.db = testDB.DB
	service.safetyChecker = NewSafetyChecker(service.db)
	service.migrator = database.NewMigrationManager(service.db, "migrations")
	service.adminCreator = NewAdminCreator(service.db, service.auth)

	// Attempt initialization - should fail during environment validation
	err = service.Initialize()
	assert.Error(t, err)

	// Verify error is of configuration type
	var initErr *InitError
	assert.ErrorAs(t, err, &initErr)
	assert.Equal(t, ErrorTypeConfig, initErr.Type)
	assert.Contains(t, initErr.Message, "Environment validation failed")
}

func TestInitService_MigrationExecution_EndToEnd(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Set required environment variable
	originalPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	defer func() {
		if originalPassword != "" {
			os.Setenv("DEFAULT_ADMIN_PASSWORD", originalPassword)
		} else {
			os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
		}
	}()
	os.Setenv("DEFAULT_ADMIN_PASSWORD", "test-admin-password-123")

	// Create initialization service
	service, err := NewInitService(testDB.Config)
	require.NoError(t, err)
	defer service.Close()

	// Override database connection to use test database and initialize components
	service.db = testDB.DB
	service.safetyChecker = NewSafetyChecker(service.db)
	service.migrator = database.NewMigrationManager(service.db, "migrations")
	service.adminCreator = NewAdminCreator(service.db, service.auth)

	// Verify no migrations exist initially
	var initialMigrationCount int64
	err = testDB.DB.Table("schema_migrations").Count(&initialMigrationCount).Error
	if err != nil {
		// Table doesn't exist yet, which is expected
		initialMigrationCount = 0
	}

	// Run initialization
	err = service.Initialize()
	assert.NoError(t, err)

	// Verify migrations were applied
	var finalMigrationCount int64
	err = testDB.DB.Table("schema_migrations").Count(&finalMigrationCount).Error
	assert.NoError(t, err)
	assert.Greater(t, finalMigrationCount, initialMigrationCount, "Migrations should have been applied")

	// Verify migration version is not dirty
	migrator := service.migrator
	version, dirty, err := migrator.GetMigrationVersion()
	assert.NoError(t, err)
	assert.False(t, dirty, "Migration should not be in dirty state")
	assert.Greater(t, version, uint(0), "Migration version should be greater than 0")
}

func TestInitService_AdminUserCreation_EndToEnd(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Set required environment variable
	originalPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	defer func() {
		if originalPassword != "" {
			os.Setenv("DEFAULT_ADMIN_PASSWORD", originalPassword)
		} else {
			os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
		}
	}()
	testPassword := "secure-admin-password-123"
	os.Setenv("DEFAULT_ADMIN_PASSWORD", testPassword)

	// Create initialization service
	service, err := NewInitService(testDB.Config)
	require.NoError(t, err)
	defer service.Close()

	// Override database connection to use test database and initialize components
	service.db = testDB.DB
	service.safetyChecker = NewSafetyChecker(service.db)
	service.migrator = database.NewMigrationManager(service.db, "migrations")
	service.adminCreator = NewAdminCreator(service.db, service.auth)

	// Run initialization
	err = service.Initialize()
	assert.NoError(t, err)

	// Verify admin user was created with correct properties
	var adminUser models.User
	err = testDB.DB.Where("username = ?", "admin").First(&adminUser).Error
	assert.NoError(t, err)

	// Verify user properties
	assert.Equal(t, "admin", adminUser.Username)
	assert.Equal(t, "admin@localhost", adminUser.Email)
	assert.Equal(t, models.RoleAdministrator, adminUser.Role)
	assert.NotEmpty(t, adminUser.PasswordHash)
	assert.NotEqual(t, testPassword, adminUser.PasswordHash, "Password should be hashed")

	// Verify password hash is valid by checking it can be verified
	// (We can't directly test password verification without importing auth service)
	assert.True(t, len(adminUser.PasswordHash) > 20, "Password hash should be properly generated")

	// Verify only one admin user exists
	var adminCount int64
	err = testDB.DB.Where("role = ?", models.RoleAdministrator).Count(&adminCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), adminCount, "Should have exactly one admin user")

	// Verify total user count
	var totalUserCount int64
	err = testDB.DB.Table("users").Count(&totalUserCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), totalUserCount, "Should have exactly one user total")
}

func TestInitService_DatabaseConnection_FailureHandling(t *testing.T) {
	// Create config with invalid database connection
	invalidConfig := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "nonexistent-host",
			Port:     "5432",
			User:     "testuser",
			Password: "testpass",
			DBName:   "testdb",
			SSLMode:  "disable",
		},
		JWT: config.JWTConfig{
			Secret: "test-jwt-secret",
		},
	}

	// Set required environment variable
	originalPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	defer func() {
		if originalPassword != "" {
			os.Setenv("DEFAULT_ADMIN_PASSWORD", originalPassword)
		} else {
			os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
		}
	}()
	os.Setenv("DEFAULT_ADMIN_PASSWORD", "test-admin-password-123")

	// Create initialization service
	service, err := NewInitService(invalidConfig)
	require.NoError(t, err)
	defer service.Close()

	// Attempt initialization - should fail during database connection
	err = service.Initialize()
	assert.Error(t, err)

	// Verify error is of database type
	var initErr *InitError
	assert.ErrorAs(t, err, &initErr)
	assert.Equal(t, ErrorTypeDatabase, initErr.Type)
	assert.Contains(t, initErr.Message, "Database connection failed")
}

func TestInitService_SafetyChecker_DetailedReporting(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Create comprehensive test data
	err := testDB.runSQLMigrations()
	require.NoError(t, err)

	// Seed default data
	err = models.SeedDefaultData(testDB.DB)
	require.NoError(t, err)

	// Create test users
	users := []*models.User{
		{Username: "user1", Email: "user1@example.com", PasswordHash: "hash1", Role: models.RoleUser},
		{Username: "user2", Email: "user2@example.com", PasswordHash: "hash2", Role: models.RoleCommenter},
	}
	for _, user := range users {
		err = testDB.DB.Create(user).Error
		require.NoError(t, err)
	}

	// Create test epics
	epics := []*models.Epic{
		{CreatorID: users[0].ID, Title: "Test Epic 1", Priority: models.PriorityHigh, Status: models.EpicStatusBacklog},
		{CreatorID: users[1].ID, Title: "Test Epic 2", Priority: models.PriorityMedium, Status: models.EpicStatusInProgress},
	}
	for _, epic := range epics {
		err = testDB.DB.Create(epic).Error
		require.NoError(t, err)
	}

	// Create safety checker
	safetyChecker := NewSafetyChecker(testDB.DB)

	// Test detailed reporting
	summary, err := safetyChecker.GetDataSummary()
	assert.NoError(t, err)
	assert.False(t, summary.IsEmpty)
	assert.Equal(t, int64(2), summary.UserCount)
	assert.Equal(t, int64(2), summary.EpicCount)
	assert.Contains(t, summary.NonEmptyTables, "users")
	assert.Contains(t, summary.NonEmptyTables, "epics")

	// Test formatted report
	report, err := safetyChecker.GetNonEmptyTablesReport()
	assert.NoError(t, err)
	assert.Contains(t, report, "users: 2 records")
	assert.Contains(t, report, "epics: 2 records")
	assert.Contains(t, report, "Initialization cannot proceed")
}

func TestInitService_ErrorContext_CorrelationID(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Create test data to trigger safety check failure
	testDB.createTestDataForSafetyCheck(t)

	// Set required environment variable
	originalPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	defer func() {
		if originalPassword != "" {
			os.Setenv("DEFAULT_ADMIN_PASSWORD", originalPassword)
		} else {
			os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
		}
	}()
	os.Setenv("DEFAULT_ADMIN_PASSWORD", "test-admin-password-123")

	// Create initialization service
	service, err := NewInitService(testDB.Config)
	require.NoError(t, err)
	defer service.Close()

	// Override database connection to use test database and initialize components
	service.db = testDB.DB
	service.safetyChecker = NewSafetyChecker(service.db)
	service.migrator = database.NewMigrationManager(service.db, "migrations")
	service.adminCreator = NewAdminCreator(service.db, service.auth)

	// Attempt initialization - should fail
	err = service.Initialize()
	assert.Error(t, err)

	// Verify error contains correlation ID context
	var initErr *InitError
	assert.ErrorAs(t, err, &initErr)
	assert.NotNil(t, initErr.Context)
	assert.Contains(t, initErr.Context, "correlation_id")
	assert.NotEmpty(t, initErr.Context["correlation_id"])
	assert.Contains(t, initErr.Context, "step")
}
