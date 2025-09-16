package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"product-requirements-management/internal/config"
	"product-requirements-management/internal/logger"
)

// Exit codes for different failure scenarios
const (
	ExitSuccess           = 0
	ExitConfigError       = 1
	ExitDatabaseError     = 2
	ExitSafetyError       = 3
	ExitMigrationError    = 4
	ExitUserCreationError = 5
	ExitSystemError       = 10
)

// InitFlags holds command-line flags for initialization
type InitFlags struct {
	DryRun  bool
	Verbose bool
	Help    bool
}

func main() {
	// Parse command-line flags
	flags, err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Show help if requested
	if flags.Help {
		showUsage()
		os.Exit(ExitSuccess)
	}

	// Load configuration
	cfg, err := loadConfiguration()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Initialize logger
	logger.Init(&cfg.Log)

	// Create correlation ID for this initialization run
	correlationID := logger.NewCorrelationID()
	ctx := logger.WithCorrelationID(context.Background(), correlationID)

	// Log initialization start
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "init_main",
		"action":    "start_initialization",
		"version":   "1.0.0", // TODO: Get from build info
	}).Info("Starting production initialization service")

	// Validate environment before proceeding
	if err := validateEnvironment(cfg, ctx); err != nil {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"component": "init_main",
			"action":    "validation_failed",
			"error":     err.Error(),
		}).Error("Environment validation failed")
		fmt.Fprintf(os.Stderr, "Environment validation failed: %v\n", err)
		os.Exit(ExitConfigError)
	}

	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "init_main",
		"action":    "validation_completed",
	}).Info("Environment validation completed successfully")

	// Run initialization process
	if err := runInitialization(cfg, flags, ctx); err != nil {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"component": "init_main",
			"action":    "initialization_failed",
			"error":     err.Error(),
		}).Error("Initialization failed")
		fmt.Fprintf(os.Stderr, "Initialization failed: %v\n", err)

		// Determine appropriate exit code based on error type
		exitCode := determineExitCode(err)
		os.Exit(exitCode)
	}

	// Log successful completion
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "init_main",
		"action":    "initialization_completed",
	}).Info("Production initialization completed successfully")
	fmt.Println("✓ Production initialization completed successfully")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Start the main application server")
	fmt.Println("2. Login with username 'admin' and the configured password")
	fmt.Println("3. Configure additional users and system settings as needed")

	os.Exit(ExitSuccess)
}

// parseFlags parses and validates command-line flags
func parseFlags() (*InitFlags, error) {
	flags := &InitFlags{}

	flag.BoolVar(&flags.DryRun, "dry-run", false, "Perform validation checks without making changes")
	flag.BoolVar(&flags.Verbose, "verbose", false, "Enable verbose logging output")
	flag.BoolVar(&flags.Help, "help", false, "Show usage information")
	flag.BoolVar(&flags.Help, "h", false, "Show usage information (shorthand)")

	flag.Parse()

	return flags, nil
}

// loadConfiguration loads and validates the application configuration
func loadConfiguration() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return cfg, nil
}

// validateEnvironment validates required environment variables and configuration
func validateEnvironment(cfg *config.Config, ctx context.Context) error {
	var missingVars []string

	// Check required database configuration
	if cfg.Database.Host == "" {
		missingVars = append(missingVars, "DB_HOST")
	}
	if cfg.Database.User == "" {
		missingVars = append(missingVars, "DB_USER")
	}
	if cfg.Database.DBName == "" {
		missingVars = append(missingVars, "DB_NAME")
	}

	// Check JWT secret (required for production)
	if cfg.JWT.Secret == "" || cfg.JWT.Secret == "your-secret-key" {
		missingVars = append(missingVars, "JWT_SECRET")
	}

	// Check for DEFAULT_ADMIN_PASSWORD
	adminPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	if adminPassword == "" {
		missingVars = append(missingVars, "DEFAULT_ADMIN_PASSWORD")
	}

	// Report missing variables
	if len(missingVars) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missingVars)
	}

	// Validate admin password strength (basic check)
	if len(adminPassword) < 8 {
		return fmt.Errorf("DEFAULT_ADMIN_PASSWORD must be at least 8 characters long")
	}

	return nil
}

// runInitialization orchestrates the initialization process
func runInitialization(cfg *config.Config, flags *InitFlags, ctx context.Context) error {
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "init_main",
		"action":    "start_orchestration",
		"dry_run":   flags.DryRun,
	}).Info("Starting initialization process")

	if flags.DryRun {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"component": "init_main",
			"action":    "dry_run_mode",
		}).Info("Dry run mode: validation only, no changes will be made")
		fmt.Println("🔍 Dry run mode: performing validation checks only")

		// In dry run mode, just validate that we can create the service
		service, err := initService.NewInitService(cfg)
		if err != nil {
			return fmt.Errorf("failed to create initialization service: %w", err)
		}
		defer service.Close()

		fmt.Println("✓ Dry run validation completed - no issues found")
		return nil
	}

	// Create and run the initialization service
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "init_main",
		"action":    "create_service",
	}).Debug("Creating initialization service")

	service, err := initService.NewInitService(cfg)
	if err != nil {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"component": "init_main",
			"action":    "service_creation_failed",
			"error":     err.Error(),
		}).Error("Failed to create initialization service")
		return fmt.Errorf("failed to create initialization service: %w", err)
	}
	defer service.Close()

	// Run the initialization
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "init_main",
		"action":    "execute_initialization",
	}).Info("Executing initialization process")

	if err := service.Initialize(); err != nil {
		return err
	}

	return nil
}

// determineExitCode determines the appropriate exit code based on error type
func determineExitCode(err error) int {
	// This is a basic implementation - will be enhanced as error types are defined
	// in subsequent tasks

	errStr := err.Error()

	// Configuration-related errors
	if contains(errStr, "configuration", "environment", "missing") {
		return ExitConfigError
	}

	// Database-related errors
	if contains(errStr, "database", "connection", "postgres") {
		return ExitDatabaseError
	}

	// Safety-related errors
	if contains(errStr, "safety", "not empty", "existing data") {
		return ExitSafetyError
	}

	// Migration-related errors
	if contains(errStr, "migration", "schema") {
		return ExitMigrationError
	}

	// User creation errors
	if contains(errStr, "user", "admin", "password") {
		return ExitUserCreationError
	}

	// Default to system error
	return ExitSystemError
}

// contains checks if any of the substrings are present in the main string
func contains(str string, substrings ...string) bool {
	for _, substr := range substrings {
		if len(str) >= len(substr) {
			for i := 0; i <= len(str)-len(substr); i++ {
				if str[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// showUsage displays usage information
func showUsage() {
	fmt.Println("Product Requirements Management - Initialization Service")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("    init [OPTIONS]")
	fmt.Println()
	fmt.Println("DESCRIPTION:")
	fmt.Println("    Safely initializes a fresh installation of the product requirements")
	fmt.Println("    management system. Performs database setup, runs migrations, and")
	fmt.Println("    creates a default admin user.")
	fmt.Println()
	fmt.Println("    SAFETY: This service will only run on completely empty databases.")
	fmt.Println("    It will exit with an error if any existing data is found.")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("    -dry-run     Perform validation checks without making changes")
	fmt.Println("    -verbose     Enable verbose logging output")
	fmt.Println("    -help, -h    Show this usage information")
	fmt.Println()
	fmt.Println("REQUIRED ENVIRONMENT VARIABLES:")
	fmt.Println("    DB_HOST                 Database host")
	fmt.Println("    DB_USER                 Database user")
	fmt.Println("    DB_PASSWORD             Database password")
	fmt.Println("    DB_NAME                 Database name")
	fmt.Println("    JWT_SECRET              JWT signing secret")
	fmt.Println("    DEFAULT_ADMIN_PASSWORD  Password for default admin user")
	fmt.Println()
	fmt.Println("OPTIONAL ENVIRONMENT VARIABLES:")
	fmt.Println("    DB_PORT                 Database port (default: 5432)")
	fmt.Println("    DB_SSLMODE              SSL mode (default: disable)")
	fmt.Println("    LOG_LEVEL               Log level (default: info)")
	fmt.Println("    LOG_FORMAT              Log format: json|text (default: json)")
	fmt.Println()
	fmt.Println("EXIT CODES:")
	fmt.Println("    0   Success")
	fmt.Println("    1   Configuration error")
	fmt.Println("    2   Database connection error")
	fmt.Println("    3   Safety check failed (database not empty)")
	fmt.Println("    4   Migration error")
	fmt.Println("    5   User creation error")
	fmt.Println("    10  System error")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("    # Validate configuration and environment")
	fmt.Println("    init -dry-run")
	fmt.Println()
	fmt.Println("    # Run full initialization")
	fmt.Println("    init")
	fmt.Println()
	fmt.Println("    # Run with verbose logging")
	fmt.Println("    init -verbose")
}
