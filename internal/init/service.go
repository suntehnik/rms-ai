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

// Note: Error types and InitError are now defined in errors.go

// InitService coordinates the initialization process
type InitService struct {
	cfg           *config.Config
	db            *gorm.DB
	auth          *auth.Service
	safetyChecker *SafetyChecker
	migrator      *database.MigrationManager
	adminCreator  *AdminCreator
	startTime     time.Time
	correlationID string
	ctx           context.Context
	errorReporter *ErrorReporter
}

// InitializationSummary contains information about the completed initialization
type InitializationSummary struct {
	CorrelationID     string        `json:"correlation_id"`
	StartTime         time.Time     `json:"start_time"`
	EndTime           time.Time     `json:"end_time"`
	TotalDuration     time.Duration `json:"total_duration"`
	StepsCompleted    []StepSummary `json:"steps_completed"`
	AdminUserCreated  bool          `json:"admin_user_created"`
	AdminUsername     string        `json:"admin_username"`
	MigrationsApplied int           `json:"migrations_applied"`
	DatabaseHost      string        `json:"database_host"`
	DatabaseName      string        `json:"database_name"`
}

// StepSummary contains information about a completed initialization step
type StepSummary struct {
	Name      string        `json:"name"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
	Status    string        `json:"status"`
	Details   interface{}   `json:"details,omitempty"`
}

// NewInitService creates a new initialization service
func NewInitService(cfg *config.Config) (*InitService, error) {
	correlationID := logger.NewCorrelationID()
	ctx := logger.WithCorrelationID(context.Background(), correlationID)

	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "init_service",
		"action":    "create_service",
	}).Info("Creating initialization service")

	// Create auth service for password hashing
	authService := auth.NewService(cfg.JWT.Secret, 24*time.Hour)

	service := &InitService{
		cfg:           cfg,
		auth:          authService,
		startTime:     time.Now(),
		correlationID: correlationID,
		ctx:           ctx,
		errorReporter: NewErrorReporter(correlationID),
	}

	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component":      "init_service",
		"action":         "service_created",
		"correlation_id": correlationID,
	}).Info("Initialization service created successfully")

	return service, nil
}

// Initialize runs the complete initialization process
func (s *InitService) Initialize() error {
	var stepSummaries []StepSummary

	logger.WithContextAndFields(s.ctx, map[string]interface{}{
		"component": "init_service",
		"action":    "start_initialization",
	}).Info("Starting initialization process")

	// Step 1: Validate environment
	stepCtx := logger.WithInitializationStep(s.ctx, "environment_validation")
	stepStart := time.Now()
	if err := s.validateEnvironment(stepCtx); err != nil {
		s.logStepFailure("environment_validation", stepStart, err)
		initErr := NewConfigError("Environment validation failed", err).
			WithStep("environment_validation").
			WithCorrelationID(s.correlationID).
			WithContext("duration", time.Since(stepStart).String())
		s.errorReporter.ReportError(initErr)
		return initErr
	}
	stepSummaries = append(stepSummaries, s.createStepSummary("environment_validation", stepStart, time.Now(), "success", nil))

	// Step 2: Establish database connection
	stepCtx = logger.WithInitializationStep(s.ctx, "database_connection")
	stepStart = time.Now()
	if err := s.connectDatabase(stepCtx); err != nil {
		s.logStepFailure("database_connection", stepStart, err)
		initErr := NewDatabaseError("Database connection failed", err).
			WithStep("database_connection").
			WithCorrelationID(s.correlationID).
			WithContext("duration", time.Since(stepStart).String()).
			WithContext("host", s.cfg.Database.Host).
			WithContext("database", s.cfg.Database.DBName)
		s.errorReporter.ReportError(initErr)
		return initErr
	}
	stepSummaries = append(stepSummaries, s.createStepSummary("database_connection", stepStart, time.Now(), "success", map[string]interface{}{
		"host":     s.cfg.Database.Host,
		"database": s.cfg.Database.DBName,
	}))

	// Step 3: Check database health
	stepCtx = logger.WithInitializationStep(s.ctx, "database_health_check")
	stepStart = time.Now()
	if err := s.checkDatabaseHealth(stepCtx); err != nil {
		s.logStepFailure("database_health_check", stepStart, err)
		initErr := NewDatabaseError("Database health check failed", err).
			WithStep("database_health_check").
			WithCorrelationID(s.correlationID).
			WithContext("duration", time.Since(stepStart).String())
		s.errorReporter.ReportError(initErr)
		return initErr
	}
	stepSummaries = append(stepSummaries, s.createStepSummary("database_health_check", stepStart, time.Now(), "success", nil))

	// Step 4: Safety check - ensure database is empty
	stepCtx = logger.WithInitializationStep(s.ctx, "safety_check")
	stepStart = time.Now()
	if err := s.performSafetyCheck(stepCtx); err != nil {
		s.logStepFailure("safety_check", stepStart, err)
		// Get additional context for safety errors
		summary, _ := s.safetyChecker.GetDataSummary()
		initErr := NewSafetyError("Database safety check failed", err).
			WithStep("safety_check").
			WithCorrelationID(s.correlationID).
			WithContext("duration", time.Since(stepStart).String())
		if summary != nil {
			initErr.WithContext("user_count", summary.UserCount).
				WithContext("epic_count", summary.EpicCount).
				WithContext("user_story_count", summary.UserStoryCount).
				WithContext("requirement_count", summary.RequirementCount).
				WithContext("non_empty_tables", summary.NonEmptyTables)
		}
		s.errorReporter.ReportError(initErr)
		return initErr
	}
	stepSummaries = append(stepSummaries, s.createStepSummary("safety_check", stepStart, time.Now(), "success", nil))

	// Step 5: Run migrations
	stepCtx = logger.WithInitializationStep(s.ctx, "migration_execution")
	stepStart = time.Now()
	migrationsApplied, err := s.runMigrations(stepCtx)
	if err != nil {
		s.logStepFailure("migration_execution", stepStart, err)
		initErr := NewMigrationError("Migration execution failed", err).
			WithStep("migration_execution").
			WithCorrelationID(s.correlationID).
			WithContext("duration", time.Since(stepStart).String()).
			WithContext("migrations_path", "migrations")
		s.errorReporter.ReportError(initErr)
		return initErr
	}
	stepSummaries = append(stepSummaries, s.createStepSummary("migration_execution", stepStart, time.Now(), "success", map[string]interface{}{
		"migrations_applied": migrationsApplied,
	}))

	// Step 6: Create admin user
	stepCtx = logger.WithInitializationStep(s.ctx, "admin_user_creation")
	stepStart = time.Now()
	adminUser, err := s.createAdminUser(stepCtx)
	if err != nil {
		s.logStepFailure("admin_user_creation", stepStart, err)
		initErr := NewCreationError("Admin user creation failed", err).
			WithStep("admin_user_creation").
			WithCorrelationID(s.correlationID).
			WithContext("duration", time.Since(stepStart).String()).
			WithContext("username", "admin")
		s.errorReporter.ReportError(initErr)
		return initErr
	}
	stepSummaries = append(stepSummaries, s.createStepSummary("admin_user_creation", stepStart, time.Now(), "success", map[string]interface{}{
		"username": adminUser.Username,
		"role":     adminUser.Role,
	}))

	// Step 7: Log success and next steps
	s.logSuccessAndNextSteps(stepSummaries, adminUser.Username, migrationsApplied)

	return nil
}

// validateEnvironment validates all required environment variables
func (s *InitService) validateEnvironment(ctx context.Context) error {
	stepStart := time.Now()
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action": "start_validation",
	}).Info("Validating environment configuration")

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

	// Log validation progress
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action":            "validation_progress",
		"variables_checked": []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_NAME", "JWT_SECRET", "DEFAULT_ADMIN_PASSWORD"},
		"missing_variables": missingVars,
		"invalid_variables": invalidVars,
	}).Info("Environment variable validation progress")

	// Report missing variables
	if len(missingVars) > 0 {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"action":            "validation_failed",
			"missing_variables": missingVars,
		}).Error("Environment validation failed - missing required variables")
		return fmt.Errorf("missing required environment variables: %v", missingVars)
	}

	// Report invalid variables
	if len(invalidVars) > 0 {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"action":            "validation_failed",
			"invalid_variables": invalidVars,
		}).Error("Environment validation failed - invalid variables")
		return fmt.Errorf("invalid environment variables: %v", invalidVars)
	}

	duration := time.Since(stepStart)
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action":   "validation_completed",
		"duration": duration.String(),
		"status":   "success",
	}).Info("Environment validation completed successfully")

	return nil
}

// connectDatabase establishes database connection
func (s *InitService) connectDatabase(ctx context.Context) error {
	stepStart := time.Now()
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action":        "start_connection",
		"database_host": s.cfg.Database.Host,
		"database_port": s.cfg.Database.Port,
		"database_name": s.cfg.Database.DBName,
		"database_user": s.cfg.Database.User,
	}).Info("Establishing database connection")

	// Create PostgreSQL connection
	db, err := database.NewPostgresDB(s.cfg)
	if err != nil {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"action": "connection_failed",
			"error":  err.Error(),
		}).Error("Failed to establish database connection")
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	s.db = db

	// Initialize safety checker
	s.safetyChecker = NewSafetyChecker(s.db)
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action": "safety_checker_initialized",
	}).Debug("Safety checker initialized")

	// Initialize migration manager
	s.migrator = database.NewMigrationManager(s.db, "migrations")
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action":          "migration_manager_initialized",
		"migrations_path": "migrations",
	}).Debug("Migration manager initialized")

	// Initialize admin creator
	s.adminCreator = NewAdminCreator(s.db, s.auth)
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action": "admin_creator_initialized",
	}).Debug("Admin creator initialized")

	duration := time.Since(stepStart)
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action":   "connection_completed",
		"duration": duration.String(),
		"status":   "success",
		"host":     s.cfg.Database.Host,
		"port":     s.cfg.Database.Port,
		"database": s.cfg.Database.DBName,
	}).Info("Database connection established successfully")

	return nil
}

// checkDatabaseHealth verifies database is accessible and responsive
func (s *InitService) checkDatabaseHealth(ctx context.Context) error {
	stepStart := time.Now()
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action": "start_health_check",
	}).Info("Checking database health")

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Get underlying sql.DB for health check
	sqlDB, err := s.db.DB()
	if err != nil {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"action": "health_check_failed",
			"error":  err.Error(),
			"reason": "failed_to_get_sql_db",
		}).Error("Failed to get underlying sql.DB for health check")
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Ping database
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action":  "ping_database",
		"timeout": "10s",
	}).Debug("Pinging database")

	if err := sqlDB.PingContext(pingCtx); err != nil {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"action": "health_check_failed",
			"error":  err.Error(),
			"reason": "ping_failed",
		}).Error("Database ping failed")
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check connection stats
	stats := sqlDB.Stats()
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action":               "connection_stats",
		"open_connections":     stats.OpenConnections,
		"idle_connections":     stats.Idle,
		"in_use_connections":   stats.InUse,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}).Debug("Database connection statistics")

	if stats.OpenConnections == 0 {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"action": "health_check_failed",
			"reason": "no_open_connections",
		}).Error("No open database connections")
		return fmt.Errorf("no open database connections")
	}

	duration := time.Since(stepStart)
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action":           "health_check_completed",
		"duration":         duration.String(),
		"status":           "success",
		"open_connections": stats.OpenConnections,
		"idle_connections": stats.Idle,
	}).Info("Database health check completed successfully")

	return nil
}

// performSafetyCheck ensures database is empty before initialization
func (s *InitService) performSafetyCheck(ctx context.Context) error {
	stepStart := time.Now()
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action": "start_safety_check",
	}).Info("Performing database safety check")

	// Check if database is empty
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action": "check_database_emptiness",
	}).Debug("Checking if database is empty")

	isEmpty, err := s.safetyChecker.IsDatabaseEmpty()
	if err != nil {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"action": "safety_check_failed",
			"error":  err.Error(),
			"reason": "failed_to_check_emptiness",
		}).Error("Failed to check database emptiness")
		return fmt.Errorf("failed to check database emptiness: %w", err)
	}

	if !isEmpty {
		// Get detailed report of existing data
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"action": "generate_data_report",
		}).Debug("Generating detailed report of existing data")

		report, reportErr := s.safetyChecker.GetNonEmptyTablesReport()
		if reportErr != nil {
			logger.WithContextAndFields(ctx, map[string]interface{}{
				"action": "safety_check_failed",
				"error":  reportErr.Error(),
				"reason": "failed_to_generate_report",
			}).Error("Database is not empty and failed to generate report")
			return fmt.Errorf("database is not empty and failed to generate report: %w", reportErr)
		}

		// Get data summary for structured logging
		summary, _ := s.safetyChecker.GetDataSummary()
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"action":                    "safety_check_failed",
			"reason":                    "database_not_empty",
			"user_count":                summary.UserCount,
			"epic_count":                summary.EpicCount,
			"user_story_count":          summary.UserStoryCount,
			"requirement_count":         summary.RequirementCount,
			"acceptance_criteria_count": summary.AcceptanceCriteriaCount,
			"comment_count":             summary.CommentCount,
			"non_empty_tables":          summary.NonEmptyTables,
		}).Error("Database safety check failed - database contains existing data")

		return fmt.Errorf("database safety check failed:\n%s", report)
	}

	duration := time.Since(stepStart)
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action":   "safety_check_completed",
		"duration": duration.String(),
		"status":   "success",
	}).Info("Database safety check completed - database is empty and safe for initialization")

	return nil
}

// runMigrations executes all pending database migrations using centralized connection management
func (s *InitService) runMigrations(ctx context.Context) (int, error) {
	stepStart := time.Now()
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action": "start_migrations",
	}).Info("Running database migrations")

	// Check current migration version using the existing migrator
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action": "check_current_version",
	}).Debug("Checking current migration version")

	version, dirty, err := s.migrator.GetMigrationVersion()
	if err != nil {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"action": "version_check_warning",
			"error":  err.Error(),
		}).Warn("Could not get current migration version, proceeding with migration")
	} else {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"action":          "current_migration_status",
			"current_version": version,
			"dirty":           dirty,
		}).Info("Current migration status")
	}

	// Run migrations using the centralized connection management approach
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action": "execute_migrations",
	}).Info("Executing database migrations")

	if err := database.RunMigrations(s.db, s.cfg); err != nil {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"action": "migration_failed",
			"error":  err.Error(),
		}).Error("Failed to run migrations")
		return 0, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Verify migrations completed successfully
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action": "verify_migrations",
	}).Debug("Verifying migration completion")

	newVersion, newDirty, err := s.migrator.GetMigrationVersion()
	if err != nil {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"action": "verification_warning",
			"error":  err.Error(),
		}).Warn("Could not verify migration completion")
	} else if newDirty {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"action":      "migration_failed",
			"reason":      "dirty_state",
			"new_version": newVersion,
		}).Error("Migrations completed but database is in dirty state")
		return 0, fmt.Errorf("migrations completed but database is in dirty state (version: %d)", newVersion)
	}

	// Calculate migrations applied (assuming we started from 0 or version)
	migrationsApplied := int(newVersion)
	if version > 0 {
		migrationsApplied = int(newVersion - version)
	}

	duration := time.Since(stepStart)
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action":             "migrations_completed",
		"duration":           duration.String(),
		"status":             "success",
		"previous_version":   version,
		"new_version":        newVersion,
		"migrations_applied": migrationsApplied,
	}).Info("Database migrations completed successfully")

	return migrationsApplied, nil
}

// createAdminUser creates the default admin user using AdminCreator
func (s *InitService) createAdminUser(ctx context.Context) (*models.User, error) {
	stepStart := time.Now()
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action": "start_admin_creation",
	}).Info("Creating default admin user")

	// Create admin user using AdminCreator
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action": "execute_admin_creation",
	}).Debug("Executing admin user creation")

	adminUser, err := s.adminCreator.CreateAdminUserFromEnv()
	if err != nil {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"action": "admin_creation_failed",
			"error":  err.Error(),
		}).Error("Failed to create admin user")
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	duration := time.Since(stepStart)
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"action":   "admin_creation_completed",
		"duration": duration.String(),
		"status":   "success",
		"username": adminUser.Username,
		"role":     adminUser.Role,
		"user_id":  adminUser.ID,
	}).Info("Default admin user created successfully")

	return adminUser, nil
}

// logSuccessAndNextSteps logs successful completion and provides next steps
func (s *InitService) logSuccessAndNextSteps(stepSummaries []StepSummary, adminUsername string, migrationsApplied int) {
	endTime := time.Now()
	totalDuration := endTime.Sub(s.startTime)

	// Create comprehensive initialization summary
	summary := InitializationSummary{
		CorrelationID:     s.correlationID,
		StartTime:         s.startTime,
		EndTime:           endTime,
		TotalDuration:     totalDuration,
		StepsCompleted:    stepSummaries,
		AdminUserCreated:  true,
		AdminUsername:     adminUsername,
		MigrationsApplied: migrationsApplied,
		DatabaseHost:      s.cfg.Database.Host,
		DatabaseName:      s.cfg.Database.DBName,
	}

	// Log comprehensive summary
	logger.WithContextAndFields(s.ctx, map[string]interface{}{
		"action":             "initialization_completed",
		"status":             "success",
		"total_duration":     totalDuration.String(),
		"steps_completed":    len(stepSummaries),
		"admin_username":     adminUsername,
		"migrations_applied": migrationsApplied,
		"database_host":      s.cfg.Database.Host,
		"database_name":      s.cfg.Database.DBName,
		"summary":            summary,
	}).Info("Initialization process completed successfully")

	// Log step-by-step summary
	logger.WithContextAndFields(s.ctx, map[string]interface{}{
		"action": "step_summary",
	}).Info("=== INITIALIZATION STEP SUMMARY ===")

	for i, step := range stepSummaries {
		logger.WithContextAndFields(s.ctx, map[string]interface{}{
			"action":      "step_detail",
			"step_number": i + 1,
			"step_name":   step.Name,
			"duration":    step.Duration.String(),
			"status":      step.Status,
			"details":     step.Details,
		}).Info(fmt.Sprintf("Step %d: %s (%s)", i+1, step.Name, step.Duration.String()))
	}

	// Log next steps for operators
	logger.WithContextAndFields(s.ctx, map[string]interface{}{
		"action": "next_steps",
	}).Info("=== INITIALIZATION COMPLETE ===")

	nextSteps := []string{
		"Start the main application server",
		fmt.Sprintf("Login with username '%s' and the password you provided", adminUsername),
		"Create additional users and configure the system as needed",
		"Review the application logs for any additional configuration requirements",
	}

	for i, step := range nextSteps {
		logger.WithContextAndFields(s.ctx, map[string]interface{}{
			"action":      "next_step",
			"step_number": i + 1,
			"instruction": step,
		}).Info(fmt.Sprintf("%d. %s", i+1, step))
	}

	logger.WithContextAndFields(s.ctx, map[string]interface{}{
		"action":         "admin_ready",
		"admin_username": adminUsername,
	}).Info("Default admin user is ready for use")
}

// createStepSummary creates a summary for a completed step
func (s *InitService) createStepSummary(name string, startTime, endTime time.Time, status string, details interface{}) StepSummary {
	return StepSummary{
		Name:      name,
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
		Status:    status,
		Details:   details,
	}
}

// logStepFailure logs detailed information about a failed step
func (s *InitService) logStepFailure(stepName string, stepStart time.Time, err error) {
	duration := time.Since(stepStart)
	stepCtx := logger.WithInitializationStep(s.ctx, stepName)

	logger.WithContextAndFields(stepCtx, map[string]interface{}{
		"action":   "step_failed",
		"duration": duration.String(),
		"error":    err.Error(),
		"status":   "failed",
	}).Error(fmt.Sprintf("Step '%s' failed after %s", stepName, duration.String()))
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
