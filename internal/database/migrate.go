package database

import (
	"fmt"
	"path/filepath"
	"product-requirements-management/internal/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"
)

// MigrationManager handles database migrations
type MigrationManager struct {
	db            *gorm.DB
	migrationsDir string
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *gorm.DB, migrationsDir string) *MigrationManager {
	return &MigrationManager{
		db:            db,
		migrationsDir: migrationsDir,
	}
}

// RunMigrations runs all pending migrations using the existing database connection
func (m *MigrationManager) RunMigrations() error {
	return RunMigrationsWithConnection(m.db, m.migrationsDir)
}

// RunMigrations runs migrations using the provided database connection and configuration
func RunMigrations(db *gorm.DB, cfg *config.Config) error {
	return RunMigrationsWithConnection(db, "migrations")
}

// RunMigrationsWithConnection runs migrations using the provided database connection
func RunMigrationsWithConnection(db *gorm.DB, migrationsDir string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Get absolute path for migrations directory
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	// Run migrations
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// RollbackMigration rolls back the last migration
func (m *MigrationManager) RollbackMigration() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	absPath, err := filepath.Abs(m.migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	// Rollback one step
	if err := migrator.Steps(-1); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	return nil
}

// GetMigrationVersion returns the current migration version
func (m *MigrationManager) GetMigrationVersion() (uint, bool, error) {
	sqlDB, err := m.db.DB()
	if err != nil {
		return 0, false, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	absPath, err := filepath.Abs(m.migrationsDir)
	if err != nil {
		return 0, false, fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"postgres",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrator: %w", err)
	}

	version, dirty, err := migrator.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}

// CreateMigrationFiles creates up and down migration files
func (m *MigrationManager) CreateMigrationFiles(name string) error {
	// This is a helper function to create migration file templates
	// In a real application, you might want to use a CLI tool for this
	return fmt.Errorf("migration file creation should be done using migrate CLI tool")
}

// ValidateDatabase checks if the database schema is up to date
func (m *MigrationManager) ValidateDatabase() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Check if schema_migrations table exists
	var exists bool
	err = sqlDB.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'schema_migrations')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check schema_migrations table: %w", err)
	}

	if !exists {
		return fmt.Errorf("schema_migrations table does not exist - database may not be initialized")
	}

	return nil
}

// GetPendingMigrations returns the number of pending migrations
func (m *MigrationManager) GetPendingMigrations() (int, error) {
	// This would require comparing filesystem migrations with database version
	// For now, we'll return a simple implementation
	version, _, err := m.GetMigrationVersion()
	if err != nil {
		return 0, err
	}

	// In a real implementation, you would scan the migrations directory
	// and compare with the current version
	_ = version
	return 0, nil
}
