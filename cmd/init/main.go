package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"product-requirements-management/internal/config"
	initService "product-requirements-management/internal/init"
	"product-requirements-management/internal/logger"
)

// Note: Exit codes are now defined in internal/init/errors.go

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
		os.Exit(initService.ExitConfigError)
	}

	// Show help if requested
	if flags.Help {
		showUsage()
		os.Exit(initService.ExitSuccess)
	}

	// Load configuration
	cfg, err := loadConfiguration()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(initService.ExitConfigError)
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
		os.Exit(initService.ExitConfigError)
	}

	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "init_main",
		"action":    "validation_completed",
	}).Info("Environment validation completed successfully")

	// Run initialization process
	if err := runInitialization(cfg, flags, ctx); err != nil {
		// Enhanced error logging with structured information
		errorFields := map[string]interface{}{
			"component": "init_main",
			"action":    "initialization_failed",
			"error":     err.Error(),
		}

		// Add additional context if it's an InitError
		if initErr, ok := err.(*initService.InitError); ok {
			errorFields["error_type"] = initErr.Type
			errorFields["error_severity"] = initErr.Severity
			errorFields["error_step"] = initErr.Step
			errorFields["error_recoverable"] = initErr.Recoverable
			errorFields["error_context"] = initErr.Context
			errorFields["error_json"] = initErr.JSON()
		}

		logger.WithContextAndFields(ctx, errorFields).Error("Initialization failed")

		// Provide user-friendly error message
		fmt.Fprintf(os.Stderr, "❌ Initialization failed: %v\n", err)

		// If it's a recoverable error, provide guidance
		if initErr, ok := err.(*initService.InitError); ok && initErr.IsRecoverable() {
			fmt.Fprintf(os.Stderr, "\n💡 This error may be recoverable. Please:\n")
			switch initErr.Type {
			case initService.ErrorTypeConfig:
				fmt.Fprintf(os.Stderr, "   - Check your environment variables\n")
				fmt.Fprintf(os.Stderr, "   - Verify configuration values\n")
				fmt.Fprintf(os.Stderr, "   - Run with -dry-run to validate configuration\n")
			case initService.ErrorTypeDatabase:
				fmt.Fprintf(os.Stderr, "   - Verify database server is running\n")
				fmt.Fprintf(os.Stderr, "   - Check database connection parameters\n")
				fmt.Fprintf(os.Stderr, "   - Ensure database user has required permissions\n")
			case initService.ErrorTypeMigration:
				fmt.Fprintf(os.Stderr, "   - Check migration files are present\n")
				fmt.Fprintf(os.Stderr, "   - Verify database schema permissions\n")
				fmt.Fprintf(os.Stderr, "   - Review migration logs for specific errors\n")
			case initService.ErrorTypeCreation:
				fmt.Fprintf(os.Stderr, "   - Verify DEFAULT_ADMIN_PASSWORD is set correctly\n")
				fmt.Fprintf(os.Stderr, "   - Check password meets security requirements\n")
				fmt.Fprintf(os.Stderr, "   - Ensure no conflicting admin user exists\n")
			}
		}

		// Determine appropriate exit code based on error type
		exitCode := initService.DetermineExitCode(err)
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

	os.Exit(initService.ExitSuccess)
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

// Note: determineExitCode and contains functions are now in internal/init/errors.go

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
	fmt.Printf("    %d   Success\n", initService.ExitSuccess)
	fmt.Printf("    %d   Configuration error\n", initService.ExitConfigError)
	fmt.Printf("    %d   Database connection error\n", initService.ExitDatabaseError)
	fmt.Printf("    %d   Safety check failed (database not empty)\n", initService.ExitSafetyError)
	fmt.Printf("    %d   Migration error\n", initService.ExitMigrationError)
	fmt.Printf("    %d   User creation error\n", initService.ExitUserCreationError)
	fmt.Printf("    %d  System error\n", initService.ExitSystemError)
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
