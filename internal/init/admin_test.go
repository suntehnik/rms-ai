package init

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"product-requirements-management/internal/auth"
	"product-requirements-management/internal/models"
)

func setupAdminTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the User model
	err = db.AutoMigrate(&models.User{})
	require.NoError(t, err)

	return db
}

func setupAdminCreator(t *testing.T) (*AdminCreator, *gorm.DB) {
	db := setupAdminTestDB(t)
	authService := auth.NewService("test-secret", 24*time.Hour)
	adminCreator := NewAdminCreator(db, authService)
	return adminCreator, db
}

func TestNewAdminCreator(t *testing.T) {
	db := setupAdminTestDB(t)
	authService := auth.NewService("test-secret", 24*time.Hour)

	adminCreator := NewAdminCreator(db, authService)

	assert.NotNil(t, adminCreator)
	assert.Equal(t, db, adminCreator.db)
	assert.Equal(t, authService, adminCreator.auth)
}

func TestCreateAdminUser_Success(t *testing.T) {
	adminCreator, db := setupAdminCreator(t)
	password := "testpassword123"

	user, err := adminCreator.CreateAdminUser(password)

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "admin", user.Username)
	assert.Equal(t, "admin@localhost", user.Email)
	assert.Equal(t, models.RoleAdministrator, user.Role)
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotEqual(t, password, user.PasswordHash) // Password should be hashed

	// Verify user was saved to database
	var savedUser models.User
	err = db.Where("username = ?", "admin").First(&savedUser).Error
	require.NoError(t, err)
	assert.Equal(t, user.ID, savedUser.ID)
	assert.Equal(t, user.Username, savedUser.Username)
	assert.Equal(t, user.Role, savedUser.Role)
}

func TestCreateAdminUser_EmptyPassword(t *testing.T) {
	adminCreator, _ := setupAdminCreator(t)

	user, err := adminCreator.CreateAdminUser("")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "password cannot be empty")
}

func TestCreateAdminUser_ShortPassword(t *testing.T) {
	adminCreator, _ := setupAdminCreator(t)

	user, err := adminCreator.CreateAdminUser("short")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "password must be at least 8 characters long")
}

func TestCreateAdminUser_MinimumValidPassword(t *testing.T) {
	adminCreator, _ := setupAdminCreator(t)
	password := "12345678" // Exactly 8 characters

	user, err := adminCreator.CreateAdminUser(password)

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "admin", user.Username)
	assert.Equal(t, models.RoleAdministrator, user.Role)
}

func TestCreateAdminUser_AlreadyExists(t *testing.T) {
	adminCreator, db := setupAdminCreator(t)

	// Create first admin user
	existingUser := &models.User{
		Username:     "admin",
		Email:        "admin@localhost",
		PasswordHash: "hashedpassword",
		Role:         models.RoleAdministrator,
	}
	err := db.Create(existingUser).Error
	require.NoError(t, err)

	// Try to create another admin user
	user, err := adminCreator.CreateAdminUser("testpassword123")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "admin user already exists")
}

func TestCreateAdminUser_PasswordHashing(t *testing.T) {
	adminCreator, _ := setupAdminCreator(t)
	password := "testpassword123"

	user, err := adminCreator.CreateAdminUser(password)

	require.NoError(t, err)
	assert.NotNil(t, user)

	// Verify password was hashed correctly by trying to verify it
	err = adminCreator.auth.VerifyPassword(password, user.PasswordHash)
	assert.NoError(t, err)

	// Verify wrong password fails
	err = adminCreator.auth.VerifyPassword("wrongpassword", user.PasswordHash)
	assert.Error(t, err)
}

func TestCreateAdminUserFromEnv_Success(t *testing.T) {
	adminCreator, _ := setupAdminCreator(t)
	password := "envpassword123"

	// Set environment variable
	os.Setenv("DEFAULT_ADMIN_PASSWORD", password)
	defer os.Unsetenv("DEFAULT_ADMIN_PASSWORD")

	user, err := adminCreator.CreateAdminUserFromEnv()

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "admin", user.Username)
	assert.Equal(t, models.RoleAdministrator, user.Role)

	// Verify password was set correctly
	err = adminCreator.auth.VerifyPassword(password, user.PasswordHash)
	assert.NoError(t, err)
}

func TestCreateAdminUserFromEnv_MissingEnvVar(t *testing.T) {
	adminCreator, _ := setupAdminCreator(t)

	// Ensure environment variable is not set
	os.Unsetenv("DEFAULT_ADMIN_PASSWORD")

	user, err := adminCreator.CreateAdminUserFromEnv()

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "DEFAULT_ADMIN_PASSWORD environment variable is required")
}

func TestCreateAdminUserFromEnv_InvalidPassword(t *testing.T) {
	adminCreator, _ := setupAdminCreator(t)

	// Set invalid password (too short)
	os.Setenv("DEFAULT_ADMIN_PASSWORD", "short")
	defer os.Unsetenv("DEFAULT_ADMIN_PASSWORD")

	user, err := adminCreator.CreateAdminUserFromEnv()

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "password must be at least 8 characters long")
}

func TestValidatePassword_ValidPasswords(t *testing.T) {
	adminCreator, _ := setupAdminCreator(t)

	validPasswords := []string{
		"12345678",            // Minimum length
		"longerpassword123",   // Longer password
		"P@ssw0rd!",           // Complex password
		"simple_password_123", // Underscore and numbers
	}

	for _, password := range validPasswords {
		t.Run("password_"+password, func(t *testing.T) {
			err := adminCreator.validatePassword(password)
			assert.NoError(t, err)
		})
	}
}

func TestValidatePassword_InvalidPasswords(t *testing.T) {
	adminCreator, _ := setupAdminCreator(t)

	invalidPasswords := []struct {
		password string
		errorMsg string
	}{
		{"", "password cannot be empty"},
		{"short", "password must be at least 8 characters long"},
		{"1234567", "password must be at least 8 characters long"},
	}

	for _, tc := range invalidPasswords {
		t.Run("password_"+tc.password, func(t *testing.T) {
			err := adminCreator.validatePassword(tc.password)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.errorMsg)
		})
	}
}

func TestAdminUserExists_NoUsers(t *testing.T) {
	adminCreator, _ := setupAdminCreator(t)

	exists, err := adminCreator.AdminUserExists()

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestAdminUserExists_AdminUserExists(t *testing.T) {
	adminCreator, db := setupAdminCreator(t)

	// Create admin user
	adminUser := &models.User{
		Username:     "admin",
		Email:        "admin@localhost",
		PasswordHash: "hashedpassword",
		Role:         models.RoleAdministrator,
	}
	err := db.Create(adminUser).Error
	require.NoError(t, err)

	exists, err := adminCreator.AdminUserExists()

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestAdminUserExists_AdministratorRoleExists(t *testing.T) {
	adminCreator, db := setupAdminCreator(t)

	// Create user with Administrator role but different username
	adminUser := &models.User{
		Username:     "superuser",
		Email:        "superuser@localhost",
		PasswordHash: "hashedpassword",
		Role:         models.RoleAdministrator,
	}
	err := db.Create(adminUser).Error
	require.NoError(t, err)

	exists, err := adminCreator.AdminUserExists()

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestAdminUserExists_OnlyRegularUsers(t *testing.T) {
	adminCreator, db := setupAdminCreator(t)

	// Create regular users
	users := []*models.User{
		{
			Username:     "user1",
			Email:        "user1@localhost",
			PasswordHash: "hashedpassword",
			Role:         models.RoleUser,
		},
		{
			Username:     "commenter1",
			Email:        "commenter1@localhost",
			PasswordHash: "hashedpassword",
			Role:         models.RoleCommenter,
		},
	}

	for _, user := range users {
		err := db.Create(user).Error
		require.NoError(t, err)
	}

	exists, err := adminCreator.AdminUserExists()

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestCreateAdminUser_TransactionRollback(t *testing.T) {
	adminCreator, db := setupAdminCreator(t)

	// Create a user with the same username to force a conflict
	existingUser := &models.User{
		Username:     "admin",
		Email:        "existing@localhost",
		PasswordHash: "hashedpassword",
		Role:         models.RoleUser,
	}
	err := db.Create(existingUser).Error
	require.NoError(t, err)

	// Try to create admin user (should fail due to username conflict)
	user, err := adminCreator.CreateAdminUser("testpassword123")

	assert.Error(t, err)
	assert.Nil(t, user)

	// Verify the existing user is still there and unchanged
	var savedUser models.User
	err = db.Where("username = ?", "admin").First(&savedUser).Error
	require.NoError(t, err)
	assert.Equal(t, existingUser.ID, savedUser.ID)
	assert.Equal(t, existingUser.Email, savedUser.Email)
	assert.Equal(t, existingUser.Role, savedUser.Role)
}
