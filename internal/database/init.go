package database

import (
	"context"
	"fmt"
	"time"

	"product-requirements-management/internal/config"
)

// Initialize sets up database connections and runs migrations
func Initialize(cfg *config.Config) (*DB, error) {
	// Create database connections
	db, err := New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connections: %w", err)
	}

	// Test connections
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if !db.IsHealthy(ctx) {
		db.Close()
		return nil, fmt.Errorf("database health check failed")
	}

	// Run migrations
	// for migrations need to create a separate instance of the DB,
	// because migrations close DB after performing a migration task
	migrationDb, err := NewPostgresDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection for migration: %w", err)
	}

	migrationManager := NewMigrationManager(migrationDb, "migrations")
	if err := migrationManager.RunMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// InitializeWithoutMigrations sets up database connections without running migrations
func InitializeWithoutMigrations(cfg *config.Config) (*DB, error) {
	// Create database connections
	db, err := New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connections: %w", err)
	}

	// Test connections
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if !db.IsHealthy(ctx) {
		db.Close()
		return nil, fmt.Errorf("database health check failed")
	}

	return db, nil
}
