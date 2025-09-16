package init

import (
	"context"
	"fmt"
	"os"

	"gorm.io/gorm"

	"product-requirements-management/internal/auth"
	"product-requirements-management/internal/logger"
	"product-requirements-management/internal/models"
)

// AdminCreator handles the creation of the default admin user
type AdminCreator struct {
	db   *gorm.DB
	auth *auth.Service
}

// NewAdminCreator creates a new admin creator instance
func NewAdminCreator(db *gorm.DB, auth *auth.Service) *AdminCreator {
	return &AdminCreator{
		db:   db,
		auth: auth,
	}
}

// CreateAdminUser creates the default admin user with the provided password
func (ac *AdminCreator) CreateAdminUser(password string) (*models.User, error) {
	correlationID := logger.NewCorrelationID()
	ctx := logger.WithCorrelationID(context.Background(), correlationID)

	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "admin_creator",
		"action":    "start_creation",
	}).Info("Creating default admin user")

	// Validate password
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "admin_creator",
		"action":    "validate_password",
	}).Debug("Validating admin password")

	if err := ac.validatePassword(password); err != nil {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"component": "admin_creator",
			"action":    "validation_failed",
			"error":     err.Error(),
		}).Error("Password validation failed")
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Hash the password using auth service
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "admin_creator",
		"action":    "hash_password",
	}).Debug("Hashing admin password")

	hashedPassword, err := ac.auth.HashPassword(password)
	if err != nil {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"component": "admin_creator",
			"action":    "hash_failed",
			"error":     err.Error(),
		}).Error("Failed to hash admin password")
		return nil, fmt.Errorf("failed to hash admin password: %w", err)
	}

	// Create admin user model
	adminUser := &models.User{
		Username:     "admin",
		Email:        "admin@localhost",
		PasswordHash: hashedPassword,
		Role:         models.RoleAdministrator,
	}

	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "admin_creator",
		"action":    "create_user_model",
		"username":  adminUser.Username,
		"email":     adminUser.Email,
		"role":      adminUser.Role,
	}).Debug("Created admin user model")

	// Use transaction for atomic user creation
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "admin_creator",
		"action":    "begin_transaction",
	}).Debug("Beginning database transaction")

	tx := ac.db.Begin()
	if tx.Error != nil {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"component": "admin_creator",
			"action":    "transaction_failed",
			"error":     tx.Error.Error(),
		}).Error("Failed to begin transaction")
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Check if admin user already exists
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "admin_creator",
		"action":    "check_existing_user",
	}).Debug("Checking for existing admin user")

	var existingUser models.User
	if err := tx.Where("username = ?", "admin").First(&existingUser).Error; err == nil {
		tx.Rollback()
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"component":     "admin_creator",
			"action":        "user_exists",
			"existing_user": existingUser.ID,
		}).Error("Admin user already exists")
		return nil, fmt.Errorf("admin user already exists")
	}

	// Create the admin user
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "admin_creator",
		"action":    "insert_user",
	}).Debug("Inserting admin user into database")

	if err := tx.Create(adminUser).Error; err != nil {
		tx.Rollback()
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"component": "admin_creator",
			"action":    "insert_failed",
			"error":     err.Error(),
		}).Error("Failed to create admin user")
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	// Commit transaction
	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "admin_creator",
		"action":    "commit_transaction",
	}).Debug("Committing transaction")

	if err := tx.Commit().Error; err != nil {
		logger.WithContextAndFields(ctx, map[string]interface{}{
			"component": "admin_creator",
			"action":    "commit_failed",
			"error":     err.Error(),
		}).Error("Failed to commit admin user creation")
		return nil, fmt.Errorf("failed to commit admin user creation: %w", err)
	}

	logger.WithContextAndFields(ctx, map[string]interface{}{
		"component": "admin_creator",
		"action":    "creation_completed",
		"username":  adminUser.Username,
		"role":      adminUser.Role,
		"user_id":   adminUser.ID,
	}).Info("Default admin user created successfully")

	return adminUser, nil
}

// CreateAdminUserFromEnv creates the admin user using password from DEFAULT_ADMIN_PASSWORD environment variable
func (ac *AdminCreator) CreateAdminUserFromEnv() (*models.User, error) {
	password := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("DEFAULT_ADMIN_PASSWORD environment variable is required")
	}

	return ac.CreateAdminUser(password)
}

// validatePassword validates the admin password meets security requirements
func (ac *AdminCreator) validatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Additional password strength checks could be added here
	// For now, we keep it simple as per requirements

	return nil
}

// AdminUserExists checks if an admin user already exists in the database
func (ac *AdminCreator) AdminUserExists() (bool, error) {
	var count int64
	if err := ac.db.Model(&models.User{}).Where("username = ? OR role = ?", "admin", models.RoleAdministrator).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check for existing admin user: %w", err)
	}
	return count > 0, nil
}
