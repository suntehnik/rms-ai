package init

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/auth"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

func TestAdminCreator_CreateAdminUser_Integration(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Run migrations to create tables
	err := testDB.runSQLMigrations()
	require.NoError(t, err)

	// Create refresh token repository
	refreshTokenRepo := repository.NewRefreshTokenRepository(testDB.DB)

	// Create auth service
	authService := auth.NewService("test-jwt-secret", 24*time.Hour, refreshTokenRepo)

	// Create admin creator
	adminCreator := NewAdminCreator(testDB.DB, authService)

	// Test password
	testPassword := "secure-admin-password-123"

	// Create admin user
	adminUser, err := adminCreator.CreateAdminUser(testPassword)
	assert.NoError(t, err)
	assert.NotNil(t, adminUser)

	// Verify user properties
	assert.Equal(t, "admin", adminUser.Username)
	assert.Equal(t, "admin@localhost", adminUser.Email)
	assert.Equal(t, models.RoleAdministrator, adminUser.Role)
	assert.NotEmpty(t, adminUser.PasswordHash)
	assert.NotEqual(t, testPassword, adminUser.PasswordHash)

	// Verify user exists in database
	var dbUser models.User
	err = testDB.DB.Where("username = ?", "admin").First(&dbUser).Error
	assert.NoError(t, err)
	assert.Equal(t, adminUser.ID, dbUser.ID)
	assert.Equal(t, adminUser.Username, dbUser.Username)
	assert.Equal(t, adminUser.PasswordHash, dbUser.PasswordHash)

	// Verify password can be verified using auth service
	err = authService.VerifyPassword(testPassword, adminUser.PasswordHash)
	assert.NoError(t, err, "Password should be verifiable")
}

func TestAdminCreator_CreateAdminUserFromEnv_Integration(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Run migrations to create tables
	err := testDB.runSQLMigrations()
	require.NoError(t, err)

	// Set environment variable
	originalPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	defer func() {
		if originalPassword != "" {
			os.Setenv("DEFAULT_ADMIN_PASSWORD", originalPassword)
		} else {
			os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
		}
	}()
	testPassword := "env-admin-password-456"
	os.Setenv("DEFAULT_ADMIN_PASSWORD", testPassword)

	// Create refresh token repository
	refreshTokenRepo := repository.NewRefreshTokenRepository(testDB.DB)

	// Create auth service
	authService := auth.NewService("test-jwt-secret", 24*time.Hour, refreshTokenRepo)

	// Create admin creator
	adminCreator := NewAdminCreator(testDB.DB, authService)

	// Create admin user from environment
	adminUser, err := adminCreator.CreateAdminUserFromEnv()
	assert.NoError(t, err)
	assert.NotNil(t, adminUser)

	// Verify user properties
	assert.Equal(t, "admin", adminUser.Username)
	assert.Equal(t, models.RoleAdministrator, adminUser.Role)

	// Verify password can be verified
	err = authService.VerifyPassword(testPassword, adminUser.PasswordHash)
	assert.NoError(t, err, "Password from environment should be verifiable")
}

func TestAdminCreator_CreateAdminUserFromEnv_MissingEnvironmentVariable(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Ensure environment variable is not set
	originalPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	defer func() {
		if originalPassword != "" {
			os.Setenv("DEFAULT_ADMIN_PASSWORD", originalPassword)
		}
	}()
	os.Unsetenv("DEFAULT_ADMIN_PASSWORD")

	// Create refresh token repository
	refreshTokenRepo := repository.NewRefreshTokenRepository(testDB.DB)

	// Create auth service
	authService := auth.NewService("test-jwt-secret", 24*time.Hour, refreshTokenRepo)

	// Create admin creator
	adminCreator := NewAdminCreator(testDB.DB, authService)

	// Attempt to create admin user from environment - should fail
	adminUser, err := adminCreator.CreateAdminUserFromEnv()
	assert.Error(t, err)
	assert.Nil(t, adminUser)
	assert.Contains(t, err.Error(), "DEFAULT_ADMIN_PASSWORD environment variable is required")
}

func TestAdminCreator_CreateAdminUser_DuplicateUser(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Run migrations to create tables
	err := testDB.runSQLMigrations()
	require.NoError(t, err)

	// Create refresh token repository
	refreshTokenRepo := repository.NewRefreshTokenRepository(testDB.DB)

	// Create auth service
	authService := auth.NewService("test-jwt-secret", 24*time.Hour, refreshTokenRepo)

	// Create admin creator
	adminCreator := NewAdminCreator(testDB.DB, authService)

	// Create first admin user
	testPassword := "secure-admin-password-123"
	adminUser1, err := adminCreator.CreateAdminUser(testPassword)
	assert.NoError(t, err)
	assert.NotNil(t, adminUser1)

	// Attempt to create second admin user - should fail
	adminUser2, err := adminCreator.CreateAdminUser(testPassword)
	assert.Error(t, err)
	assert.Nil(t, adminUser2)
	assert.Contains(t, err.Error(), "admin user already exists")

	// Verify only one admin user exists
	var adminCount int64
	err = testDB.DB.Where("username = ?", "admin").Count(&adminCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), adminCount)
}

func TestAdminCreator_CreateAdminUser_WeakPassword(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Create refresh token repository
	refreshTokenRepo := repository.NewRefreshTokenRepository(testDB.DB)

	// Create auth service
	authService := auth.NewService("test-jwt-secret", 24*time.Hour, refreshTokenRepo)

	// Create admin creator
	adminCreator := NewAdminCreator(testDB.DB, authService)

	// Test cases for weak passwords
	weakPasswords := []string{
		"",        // Empty password
		"weak",    // Too short
		"1234567", // 7 characters (less than 8)
	}

	for _, weakPassword := range weakPasswords {
		t.Run("WeakPassword_"+weakPassword, func(t *testing.T) {
			// Attempt to create admin user with weak password
			adminUser, err := adminCreator.CreateAdminUser(weakPassword)
			assert.Error(t, err)
			assert.Nil(t, adminUser)
			assert.Contains(t, err.Error(), "password")
		})
	}
}

func TestAdminCreator_AdminUserExists_Integration(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Run migrations to create tables
	err := testDB.runSQLMigrations()
	require.NoError(t, err)

	// Create refresh token repository
	refreshTokenRepo := repository.NewRefreshTokenRepository(testDB.DB)

	// Create auth service
	authService := auth.NewService("test-jwt-secret", 24*time.Hour, refreshTokenRepo)

	// Create admin creator
	adminCreator := NewAdminCreator(testDB.DB, authService)

	// Initially no admin user should exist
	exists, err := adminCreator.AdminUserExists()
	assert.NoError(t, err)
	assert.False(t, exists)

	// Create admin user
	testPassword := "secure-admin-password-123"
	adminUser, err := adminCreator.CreateAdminUser(testPassword)
	assert.NoError(t, err)
	assert.NotNil(t, adminUser)

	// Now admin user should exist
	exists, err = adminCreator.AdminUserExists()
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestAdminCreator_AdminUserExists_ByRole(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Run migrations to create tables
	err := testDB.runSQLMigrations()
	require.NoError(t, err)

	// Create refresh token repository
	refreshTokenRepo := repository.NewRefreshTokenRepository(testDB.DB)

	// Create auth service
	authService := auth.NewService("test-jwt-secret", 24*time.Hour, refreshTokenRepo)

	// Create admin creator
	adminCreator := NewAdminCreator(testDB.DB, authService)

	// Create a user with administrator role but different username
	adminUser := &models.User{
		Username:     "superadmin",
		Email:        "superadmin@localhost",
		PasswordHash: "hashed_password",
		Role:         models.RoleAdministrator,
	}
	err = testDB.DB.Create(adminUser).Error
	require.NoError(t, err)

	// AdminUserExists should return true because there's a user with Administrator role
	exists, err := adminCreator.AdminUserExists()
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestAdminCreator_CreateAdminUser_DatabaseTransaction(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Run migrations to create tables
	err := testDB.runSQLMigrations()
	require.NoError(t, err)

	// Create refresh token repository
	refreshTokenRepo := repository.NewRefreshTokenRepository(testDB.DB)

	// Create auth service
	authService := auth.NewService("test-jwt-secret", 24*time.Hour, refreshTokenRepo)

	// Create admin creator
	adminCreator := NewAdminCreator(testDB.DB, authService)

	// Verify no users exist initially
	var initialUserCount int64
	err = testDB.DB.Table("users").Count(&initialUserCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), initialUserCount)

	// Create admin user
	testPassword := "secure-admin-password-123"
	adminUser, err := adminCreator.CreateAdminUser(testPassword)
	assert.NoError(t, err)
	assert.NotNil(t, adminUser)

	// Verify exactly one user exists
	var finalUserCount int64
	err = testDB.DB.Table("users").Count(&finalUserCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), finalUserCount)

	// Verify the user is the admin user we created
	var dbUser models.User
	err = testDB.DB.First(&dbUser).Error
	assert.NoError(t, err)
	assert.Equal(t, adminUser.ID, dbUser.ID)
	assert.Equal(t, "admin", dbUser.Username)
	assert.Equal(t, models.RoleAdministrator, dbUser.Role)
}

func TestAdminCreator_PasswordHashing_Integration(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Run migrations to create tables
	err := testDB.runSQLMigrations()
	require.NoError(t, err)

	// Create refresh token repository
	refreshTokenRepo := repository.NewRefreshTokenRepository(testDB.DB)

	// Create auth service
	authService := auth.NewService("test-jwt-secret", 24*time.Hour, refreshTokenRepo)

	// Create admin creator
	adminCreator := NewAdminCreator(testDB.DB, authService)

	// Test different passwords
	testPasswords := []string{
		"simple-password-123",
		"Complex!Password@456",
		"very-long-password-with-many-characters-789",
		"P@ssw0rd!",
	}

	for i, testPassword := range testPasswords {
		t.Run("Password_"+string(rune(i)), func(t *testing.T) {
			// Reset database for each test
			err := testDB.reset()
			require.NoError(t, err)

			// Run migrations again
			err = testDB.runSQLMigrations()
			require.NoError(t, err)

			// Create admin user
			adminUser, err := adminCreator.CreateAdminUser(testPassword)
			assert.NoError(t, err)
			assert.NotNil(t, adminUser)

			// Verify password is hashed (not plain text)
			assert.NotEqual(t, testPassword, adminUser.PasswordHash)
			assert.True(t, len(adminUser.PasswordHash) > len(testPassword))

			// Verify password can be verified
			err = authService.VerifyPassword(testPassword, adminUser.PasswordHash)
			assert.NoError(t, err, "Password should be verifiable for: %s", testPassword)

			// Verify wrong password fails verification
			err = authService.VerifyPassword("wrong-password", adminUser.PasswordHash)
			assert.Error(t, err, "Wrong password should not be verifiable")
		})
	}
}
