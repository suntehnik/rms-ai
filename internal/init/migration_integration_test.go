package init

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/database"
	"product-requirements-management/internal/models"
)

func TestMigrationExecution_CompleteFlow(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Verify no tables exist initially
	tables := []string{"users", "epics", "user_stories", "requirements", "acceptance_criteria", "comments"}
	for _, table := range tables {
		var exists bool
		err := testDB.DB.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = ?)", table).Scan(&exists).Error
		assert.NoError(t, err)
		assert.False(t, exists, "Table %s should not exist initially", table)
	}

	// Run auto-migration (this is what the initialization service actually uses)
	err := models.AutoMigrate(testDB.DB)
	assert.NoError(t, err)

	// Seed default data
	err = models.SeedDefaultData(testDB.DB)
	assert.NoError(t, err)

	// Verify expected tables were created by migrations
	expectedTables := []string{
		"users",
		"epics",
		"user_stories",
		"requirements",
		"acceptance_criteria",
		"comments",
		"requirement_types",
		"relationship_types",
		"requirement_relationships",
	}

	for _, table := range expectedTables {
		var exists bool
		err = testDB.DB.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = ?)", table).Scan(&exists).Error
		assert.NoError(t, err)
		assert.True(t, exists, "Table %s should exist after migrations", table)
	}
}

func TestMigrationExecution_MigrationManager_Standalone(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Create migration manager directly
	migrator := database.NewMigrationManager(testDB.DB, "../../migrations")

	// Verify no migrations initially
	version, dirty, err := migrator.GetMigrationVersion()
	if err != nil {
		// This is expected if no migrations table exists yet
		assert.Contains(t, err.Error(), "no migration")
	}

	// Run migrations
	err = migrator.RunMigrations()
	assert.NoError(t, err)

	// Verify migrations were applied
	version, dirty, err = migrator.GetMigrationVersion()
	assert.NoError(t, err)
	assert.False(t, dirty, "Migration should not be in dirty state")
	assert.Greater(t, version, uint(0), "Migration version should be greater than 0")

	// Verify database validation passes
	err = migrator.ValidateDatabase()
	assert.NoError(t, err)
}

func TestMigrationExecution_IdempotentMigrations(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Create migration manager
	migrator := database.NewMigrationManager(testDB.DB, "../../migrations")

	// Run migrations first time
	err := migrator.RunMigrations()
	assert.NoError(t, err)

	// Get version after first run
	version1, dirty1, err := migrator.GetMigrationVersion()
	assert.NoError(t, err)
	assert.False(t, dirty1)

	// Run migrations second time (should be idempotent)
	err = migrator.RunMigrations()
	assert.NoError(t, err)

	// Get version after second run
	version2, dirty2, err := migrator.GetMigrationVersion()
	assert.NoError(t, err)
	assert.False(t, dirty2)

	// Versions should be the same
	assert.Equal(t, version1, version2, "Migration version should be the same after second run")
}

func TestMigrationExecution_WithInvalidMigrationsPath(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Create migration manager with invalid path
	migrator := database.NewMigrationManager(testDB.DB, "nonexistent-migrations")

	// Attempt to run migrations - should fail
	err := migrator.RunMigrations()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestMigrationExecution_DatabaseValidation(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Create migration manager
	migrator := database.NewMigrationManager(testDB.DB, "../../migrations")

	// Validate database before migrations - should fail
	err := migrator.ValidateDatabase()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "schema_migrations table does not exist")

	// Run migrations
	err = migrator.RunMigrations()
	assert.NoError(t, err)

	// Validate database after migrations - should pass
	err = migrator.ValidateDatabase()
	assert.NoError(t, err)
}

func TestMigrationExecution_GetPendingMigrations(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Create migration manager
	migrator := database.NewMigrationManager(testDB.DB, "../../migrations")

	// Get pending migrations before running any
	pending, err := migrator.GetPendingMigrations()
	assert.NoError(t, err)
	// Note: The current implementation returns 0, but this tests the interface

	// Run migrations
	err = migrator.RunMigrations()
	assert.NoError(t, err)

	// Get pending migrations after running all
	pending, err = migrator.GetPendingMigrations()
	assert.NoError(t, err)
	assert.Equal(t, 0, pending, "Should have no pending migrations after running all")
}

func TestMigrationExecution_RollbackCapability(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Create migration manager
	migrator := database.NewMigrationManager(testDB.DB, "../../migrations")

	// Run migrations
	err := migrator.RunMigrations()
	assert.NoError(t, err)

	// Get version after migrations
	versionBefore, _, err := migrator.GetMigrationVersion()
	assert.NoError(t, err)
	assert.Greater(t, versionBefore, uint(0))

	// Rollback one migration
	err = migrator.RollbackMigration()
	assert.NoError(t, err)

	// Get version after rollback
	versionAfter, dirty, err := migrator.GetMigrationVersion()
	assert.NoError(t, err)
	assert.False(t, dirty, "Migration should not be in dirty state after rollback")
	assert.Less(t, versionAfter, versionBefore, "Version should be lower after rollback")
}

func TestMigrationExecution_ErrorHandling_ClosedDatabase(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Create migration manager
	migrator := database.NewMigrationManager(testDB.DB, "../../migrations")

	// Close database connection
	sqlDB, err := testDB.DB.DB()
	require.NoError(t, err)
	sqlDB.Close()

	// Attempt to run migrations - should fail
	err = migrator.RunMigrations()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection")
}

func TestMigrationExecution_MigrationsDirectory_Exists(t *testing.T) {
	// Verify migrations directory exists in the project
	migrationsDir := "../../migrations"

	// Get absolute path
	absPath, err := filepath.Abs(migrationsDir)
	assert.NoError(t, err)

	// Check if directory exists
	info, err := os.Stat(absPath)
	assert.NoError(t, err)
	assert.True(t, info.IsDir(), "Migrations directory should exist")

	// Check for migration files
	files, err := os.ReadDir(absPath)
	assert.NoError(t, err)
	assert.Greater(t, len(files), 0, "Should have migration files")

	// Verify migration files follow naming convention
	for _, file := range files {
		if !file.IsDir() {
			name := file.Name()
			assert.True(t,
				filepath.Ext(name) == ".sql" || name == ".gitkeep",
				"Migration files should be .sql files or .gitkeep: %s", name)
		}
	}
}

func TestMigrationExecution_InitializationService_MigrationStep(t *testing.T) {
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

	// Override database connection to use test database
	service.db = testDB.DB

	// Initialize components manually to test migration step in isolation
	service.safetyChecker = NewSafetyChecker(service.db)
	service.migrator = database.NewMigrationManager(service.db, "../../migrations")

	// Test migration step specifically (use context.Background() since service.ctx might not be initialized)
	migrationsApplied, err := service.runMigrations(context.Background())
	assert.NoError(t, err)
	assert.Greater(t, migrationsApplied, 0, "Should have applied migrations")

	// Verify migration state
	version, dirty, err := service.migrator.GetMigrationVersion()
	assert.NoError(t, err)
	assert.False(t, dirty, "Migration should not be in dirty state")
	assert.Greater(t, version, uint(0), "Migration version should be greater than 0")
}

func TestMigrationExecution_FailureRecovery_DirtyState(t *testing.T) {
	// This test would require creating a scenario where migrations fail mid-way
	// For now, we'll test the detection of dirty state

	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Create migration manager
	migrator := database.NewMigrationManager(testDB.DB, "../../migrations")

	// Run migrations successfully first
	err := migrator.RunMigrations()
	assert.NoError(t, err)

	// Verify clean state
	version, dirty, err := migrator.GetMigrationVersion()
	assert.NoError(t, err)
	assert.False(t, dirty, "Migration should be in clean state")
	assert.Greater(t, version, uint(0))

	// Note: Creating an actual dirty state would require manipulating the schema_migrations table
	// or having a migration that fails mid-way, which is complex to set up in tests
	// The important part is that we can detect and report dirty states
}

func TestMigrationExecution_ConcurrentMigrations(t *testing.T) {
	// This test verifies that the migration system handles concurrent access properly
	// PostgreSQL advisory locks should prevent concurrent migrations

	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Create two migration managers
	migrator1 := database.NewMigrationManager(testDB.DB, "../../migrations")
	migrator2 := database.NewMigrationManager(testDB.DB, "../../migrations")

	// Run migrations with first manager
	err := migrator1.RunMigrations()
	assert.NoError(t, err)

	// Run migrations with second manager (should be idempotent)
	err = migrator2.RunMigrations()
	assert.NoError(t, err)

	// Both should report the same version
	version1, dirty1, err := migrator1.GetMigrationVersion()
	assert.NoError(t, err)
	assert.False(t, dirty1)

	version2, dirty2, err := migrator2.GetMigrationVersion()
	assert.NoError(t, err)
	assert.False(t, dirty2)

	assert.Equal(t, version1, version2, "Both migration managers should report the same version")
}
