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

	// Run migrations using centralized connection management
	// Use the existing database connection instead of creating a separate one
	if err := RunMigrations(db.Postgres, cfg); err != nil {
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
