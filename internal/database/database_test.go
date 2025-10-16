package database

import (
	"context"
	"testing"
	"time"

	"product-requirements-management/internal/config"
)

func TestDatabaseConnection(t *testing.T) {
	// Skip if no test database is configured
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "test",
			DBName:   "test_db",
			SSLMode:  "disable",
		},
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
			DB:       1, // Use different DB for tests
		},
	}

	// Try to create database connections
	db, err := New(cfg)
	if err != nil {
		t.Skipf("Skipping database test - no test database available: %v", err)
		return
	}
	defer db.Close()

	// Test health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health := db.CheckHealth(ctx)
	if health.Overall.Status != "healthy" {
		t.Errorf("Expected healthy database, got: %s - %s", health.Overall.Status, health.Overall.Message)
	}

	// Test individual components
	if health.PostgreSQL.Status != "healthy" {
		t.Errorf("PostgreSQL not healthy: %s", health.PostgreSQL.Message)
	}

	if health.Redis.Status != "healthy" {
		t.Errorf("Redis not healthy: %s", health.Redis.Message)
	}
}

func TestMigrationManager(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "test",
			DBName:   "test_db",
			SSLMode:  "disable",
		},
	}

	db, err := New(cfg)
	if err != nil {
		t.Skipf("Skipping migration test - no test database available: %v", err)
		return
	}
	defer db.Close()

	// Test migration manager creation
	migrationManager := NewMigrationManager(db.Postgres, "../../migrations")
	if migrationManager == nil {
		t.Error("Failed to create migration manager")
	}

	// Test getting migration version (should work even if no migrations are run)
	_, _, err = migrationManager.GetMigrationVersion()
	if err != nil {
		// This is expected if no migrations table exists yet
		t.Logf("No migration version found (expected for fresh database): %v", err)
	}
}
