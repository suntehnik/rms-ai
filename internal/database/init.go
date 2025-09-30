package database

import (
	"context"
	"fmt"
	"time"

	"product-requirements-management/internal/config"
	"product-requirements-management/internal/models"
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
	// Initialize PostgreSQL connection
	pg, err := initPostgreSQL(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}

	// Initialize Redis connection
	rdb, err := initRedis(cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	db := &DB{
		Postgres: pg,
		Redis:    rdb,
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

// InitializeForProduction sets up database connections for production use without any migrations
// This assumes the database has already been initialized with proper migrations
func InitializeForProduction(cfg *config.Config) (*DB, error) {
	// Initialize PostgreSQL connection without auto-migrations
	pg, err := initPostgreSQL(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}

	// Initialize Redis connection
	rdb, err := initRedis(cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	db := &DB{
		Postgres: pg,
		Redis:    rdb,
	}

	// Test connections
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if !db.IsHealthy(ctx) {
		db.Close()
		return nil, fmt.Errorf("database health check failed")
	}

	// Only seed default data if it doesn't exist (safe for production)
	if err := models.SeedDefaultData(db.Postgres); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to seed default data: %w", err)
	}

	return db, nil
}
