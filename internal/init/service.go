package init

import (
	"context"
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"

	"product-requirements-management/internal/auth"
	"product-requirements-management/internal/config"
	"product-requirements-management/internal/database"
	"product-requirements-management/internal/logger"
	"product-requirements-management/internal/models"
)

// InitError represents different types of initialization errors
type InitError struct {
	Type    ErrorType              `json:"type"`
	Message string                 `json:"message"`
	Cause   error                  `json:"cause,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

func (e *InitError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// ErrorType represents the category of initialization error
type ErrorType string

const (
	ErrorTypeConfig   ErrorType = "configuration"
	ErrorTypeDatabase ErrorType = "database"
	ErrorTypeSafety   ErrorType = "safety"
	ErrorTypeCreation ErrorType = "creation"
	ErrorTypeSystem   ErrorType = "system"
)

// InitService coordinates the initialization process
type InitService struct {
	cfg           *config.Config
	db            *gorm.DB
	auth          *auth.Service
	safetyChecker *SafetyChecker
	migrator      *database.MigrationManager
	startTime     time.Time
}

// NewInitService creates a new initialization service
func NewInitService(cfg *config.Config) (*InitService, error) {
	logger.WithField("component", "init_service").Info("Creating initialization service")

	// Create auth service for password hashing
	authService := auth.NewService(cfg.JWT.Secret, 24*time.Hour)

	service := &InitService{
		cfg:       cfg,
		auth:      authService,
		startTime: time.Now(),
	}

	return service, nil
}

// Initialize runs the complete initialization process
func (s *InitService) Initialize() error {
	logger.WithField("component", "init_service").Info("Starting initialization process")

	// Step 1: Validate environment
	if err := s.validateEnvironment(); err != nil {
		return &InitError{
			Type:    ErrorTypeConfig,
			Message: "Environment validation failed",
			Cause:   err,
		}
	}

	// Step 2: Establish database connection
	if err := s.connectDatabase(); err != nil {
		return &InitError{
			Type:    ErrorTypeDatabase,
			Message: "Database connection failed",
			Cause:   err,
		}
	}

	// Step 3: Check database health
	if err := s.checkDatabaseHealth(); err != nil {
		return &InitError{
			Type:    ErrorTypeDatabase,
			Message: "Database health check failed",
			Cause:   err,
		}
	}

	// Step 4: Safety check - ensure database is empty
	if err := s.performSafetyCheck(); err != nil {
		return &InitError{
			Type:    ErrorTypeSafety,
			Message: "Database safety check failed",
			Cause:   err,
		}
	}

	// Step 5: Run migrations
	if err := s.runMigrations(); err != nil {
		return &InitError{
			Type:    ErrorTypeDatabase,
			Message: "Migration execution failed",
			Cause:   err,
		}
	}

	// Step 6: Create admin user
	if err := s.createAdminUser(); err != nil {
		return &InitError{
			Type:    ErrorTypeCreation,
			Message: "Admin user creation failed",
			Cause:   err,
		}
	}

	// Step 7: Log success and next steps
	s.logSuccessAndNextSteps()

	return nil
}

// validateEnvironment validates all required environment variables
func (s *InitService) validateEnvironment() error {
	stepStart := time.Now()
	logger.WithField("step", "environment_validation").Info("Validating environment configuration")

	var missingVars []string
	var invalidVars []string

	// Check required database configuration
	if s.cfg.Database.Host == "" {
		missingVars = append(missingVars, "DB_HOST")
	}
	if s.cfg.Database.Port == "" {
		missingVars = append(missingVars, "DB_PORT")
	}
	if s.cfg.Database.User == "" {
		missingVars = append(missingVars, "DB_USER")
	}
	if s.cfg.Database.DBName == "" {
		missingVars = append(missingVars, "DB_NAME")
	}

	// Check JWT secret (should not be default value)
	if s.cfg.JWT.Secret == "your-secret-key" || s.cfg.JWT.Secret == "" {
		missingVars = append(missingVars, "JWT_SECRET")
	}

	// Check admin password
	adminPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	if adminPassword == "" {
		missingVars = append(missingVars, "DEFAULT_ADMIN_PASSWORD")
	} else if len(adminPassword) < 8 {
		invalidVars = append(invalidVars, "DEFAULT_ADMIN_PASSWORD (must be at least 8 characters)")
	}

	// Report missing variables
	if len(missingVars) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missingVars)
	}

	// Report invalid variables
	if len(invalidVars) > 0 {
		return fmt.Errorf("invalid environment variables: %v", invalidVars)
	}

	duration := time.Since(stepStart)
	logger.WithFields(map[string]interface{}{
		"step":     "environment_validation",
		"duration": duration.String(),
		"status":   "success",
	}).Info("Environment validation completed successfully")

	return nil
}

// connectDatabase establishes database connection
func (s *InitService) connectDatabase() error {
	stepStart := time.Now()
	logger.WithField("step", "database_connection").Info("Establishing database connection")

	// Create PostgreSQL connection
	db, err := database.NewPostgresDB(s.cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	s.db = db

	// Initialize safety checker
	s.safetyChecker = NewSafetyChecker(s.db)

	// Initialize migration manager
	s.migrator = database.NewMigrationManager(s.db, "migrations")

	duration := time.Since(stepStart)
	logger.WithFields(map[string]interface{}{
		"step":     "database_connection",
		"duration": duration.String(),
		"status":   "success",
		"host":     s.cfg.Database.Host,
		"port":     s.cfg.Database.Port,
		"database": s.cfg.Database.DBName,
	}).Info("Database connection established successfully")

	return nil
}

// checkDatabaseHealth verifies database is accessible and responsive
func (s *InitService) checkDatabaseHealth() error {
	stepStart := time.Now()
	logger.WithField("step", "database_health_check").Info("Checking database health")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get underlying sql.DB for health check
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Ping database
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check connection stats
	stats := sqlDB.Stats()
	if stats.OpenConnections == 0 {
		return fmt.Errorf("no open database connections")
	}

	duration := time.Since(stepStart)
	logger.WithFields(map[string]interface{}{
		"step":             "database_health_check",
		"duration":         duration.String(),
		"status":           "success",
		"open_connections": stats.OpenConnections,
		"idle_connections": stats.Idle,
	}).Info("Database health check completed successfully")

	return nil
}

// performSafetyCheck ensures database is empty before initialization
func (s *InitService) performSafetyCheck() error {
	stepStart := time.Now()
	logger.WithField("step", "safety_check").Info("Performing database safety check")

	// Check if database is empty
	isEmpty, err := s.safetyChecker.IsDatabaseEmpty()
	if err != nil {
		return fmt.Errorf("failed to check database emptiness: %w", err)
	}

	if !isEmpty {
		// Get detailed report of existing data
		report, reportErr := s.safetyChecker.GetNonEmptyTablesReport()
		if reportErr != nil {
			return fmt.Errorf("database is not empty and failed to generate report: %w", reportErr)
		}

		logger.WithField("step", "safety_check").Error("Database safety check failed - database contains existing data")
		return fmt.Errorf("database safety check failed:\n%s", report)
	}

	duration := time.Since(stepStart)
	logger.WithFields(map[string]interface{}{
		"step":     "safety_check",
		"duration": duration.String(),
		"status":   "success",
	}).Info("Database safety check completed - database is empty and safe for initialization")

	return nil
}

// runMigrations executes all pending database migrations
func (s *InitService) runMigrations() error {
	stepStart := time.Now()
	logger.WithField("step", "migration_execution").Info("Running database migrations")

	// Check current migration version
	version, dirty, err := s.migrator.GetMigrationVersion()
	if err != nil {
		logger.WithField("step", "migration_execution").Warn("Could not get current migration version, proceeding with migration")
	} else {
		logger.WithFields(map[string]interface{}{
			"current_version": version,
			"dirty":           dirty,
		}).Info("Current migration status")
	}

	// Run migrations
	if err := s.migrator.RunMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Verify migrations completed successfully
	newVersion, newDirty, err := s.migrator.GetMigrationVersion()
	if err != nil {
		logger.WithField("step", "migration_execution").Warn("Could not verify migration completion")
	} else if newDirty {
		return fmt.Errorf("migrations completed but database is in dirty state (version: %d)", newVersion)
	}

	duration := time.Since(stepStart)
	logger.WithFields(map[string]interface{}{
		"step":        "migration_execution",
		"duration":    duration.String(),
		"status":      "success",
		"new_version": newVersion,
	}).Info("Database migrations completed successfully")

	return nil
}

// createAdminUser creates the default admin user
func (s *InitService) createAdminUser() error {
	stepStart := time.Now()
	logger.WithField("step", "admin_user_creation").Info("Creating default admin user")

	// Get admin password from environment
	adminPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	if adminPassword == "" {
		return fmt.Errorf("DEFAULT_ADMIN_PASSWORD environment variable is required")
	}

	// Hash the password
	hashedPassword, err := s.auth.HashPassword(adminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	// Create admin user
	adminUser := &models.User{
		Username:     "admin",
		Email:        "admin@localhost",
		PasswordHash: hashedPassword,
		Role:         models.RoleAdministrator,
	}

	// Use transaction for user creation
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	if err := tx.Create(adminUser).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit admin user creation: %w", err)
	}

	duration := time.Since(stepStart)
	logger.WithFields(map[string]interface{}{
		"step":     "admin_user_creation",
		"duration": duration.String(),
		"status":   "success",
		"username": adminUser.Username,
		"role":     adminUser.Role,
		"user_id":  adminUser.ID,
	}).Info("Default admin user created successfully")

	return nil
}

// logSuccessAndNextSteps logs successful completion and provides next steps
func (s *InitService) logSuccessAndNextSteps() {
	totalDuration := time.Since(s.startTime)

	logger.WithFields(map[string]interface{}{
		"component":      "init_service",
		"status":         "completed",
		"total_duration": totalDuration.String(),
	}).Info("Initialization process completed successfully")

	// Log next steps for operators
	logger.Info("=== INITIALIZATION COMPLETE ===")
	logger.Info("Next steps:")
	logger.Info("1. Start the main application server")
	logger.Info("2. Login with username 'admin' and the password you provided")
	logger.Info("3. Create additional users and configure the system as needed")
	logger.Info("4. Review the application logs for any additional configuration requirements")
	logger.WithField("admin_username", "admin").Info("Default admin user is ready for use")
}

// Close cleans up resources used by the initialization service
func (s *InitService) Close() error {
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err == nil {
			return sqlDB.Close()
		}
	}
	return nil
}
