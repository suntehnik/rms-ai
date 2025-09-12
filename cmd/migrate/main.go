package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"product-requirements-management/internal/config"
	"product-requirements-management/internal/database"
)

func main() {
	var (
		up      = flag.Bool("up", false, "Run migrations up")
		down    = flag.Bool("down", false, "Rollback one migration")
		version = flag.Bool("version", false, "Show current migration version")
	)
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection without running migrations
	db, err := database.InitializeWithoutMigrations(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create migration manager
	migrationManager := database.NewMigrationManager(db.Postgres, "migrations")

	switch {
	case *up:
		fmt.Println("Running migrations...")
		if err := migrationManager.RunMigrations(); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		fmt.Println("Migrations completed successfully")

	case *down:
		fmt.Println("Rolling back migration...")
		if err := migrationManager.RollbackMigration(); err != nil {
			log.Fatalf("Failed to rollback migration: %v", err)
		}
		fmt.Println("Migration rollback completed successfully")

	case *version:
		version, dirty, err := migrationManager.GetMigrationVersion()
		if err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}
		fmt.Printf("Current migration version: %d (dirty: %t)\n", version, dirty)

	default:
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/migrate/main.go -up      # Run migrations")
		fmt.Println("  go run cmd/migrate/main.go -down    # Rollback one migration")
		fmt.Println("  go run cmd/migrate/main.go -version # Show current version")
		os.Exit(1)
	}
}
