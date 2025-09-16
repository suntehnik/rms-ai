package init

import (
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
	logger.WithField("component", "admin_creator").Info("Creating default admin user")

	// Validate password
	if err := ac.validatePassword(password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Hash the password using auth service
	hashedPassword, err := ac.auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash admin password: %w", err)
	}

	// Create admin user model
	adminUser := &models.User{
		Username:     "admin",
		Email:        "admin@localhost",
		PasswordHash: hashedPassword,
		Role:         models.RoleAdministrator,
	}

	// Use transaction for atomic user creation
	tx := ac.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Check if admin user already exists
	var existingUser models.User
	if err := tx.Where("username = ?", "admin").First(&existingUser).Error; err == nil {
		tx.Rollback()
		return nil, fmt.Errorf("admin user already exists")
	}

	// Create the admin user
	if err := tx.Create(adminUser).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit admin user creation: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"component": "admin_creator",
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
