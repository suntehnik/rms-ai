package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
)

func setupUserTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate user model
	err = db.AutoMigrate(&models.User{})
	require.NoError(t, err)

	return db
}

func createTestUser(t *testing.T, repo UserRepository, username, email string, role models.UserRole) *models.User {
	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: "hashed_password",
		Role:         role,
	}

	err := repo.Create(user)
	require.NoError(t, err)

	return user
}

func TestUserRepository_Create(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserRepository(db)

	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}

	err := repo.Create(user)
	assert.NoError(t, err)
	assert.NotNil(t, user.ID)
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserRepository(db)

	// Create test user
	user := createTestUser(t, repo, "testuser", "test@example.com", models.RoleUser)

	// Get by username
	retrieved, err := repo.GetByUsername("testuser")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Username, retrieved.Username)
	assert.Equal(t, user.Email, retrieved.Email)
}

func TestUserRepository_GetByUsername_NotFound(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserRepository(db)

	retrieved, err := repo.GetByUsername("nonexistent")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, retrieved)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserRepository(db)

	// Create test user
	user := createTestUser(t, repo, "testuser", "test@example.com", models.RoleUser)

	// Get by email
	retrieved, err := repo.GetByEmail("test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Username, retrieved.Username)
	assert.Equal(t, user.Email, retrieved.Email)
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserRepository(db)

	retrieved, err := repo.GetByEmail("nonexistent@example.com")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, retrieved)
}

func TestUserRepository_ExistsByUsername(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserRepository(db)

	// Create test user
	createTestUser(t, repo, "testuser", "test@example.com", models.RoleUser)

	// Check existence
	exists, err := repo.ExistsByUsername("testuser")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Check non-existence
	exists, err = repo.ExistsByUsername("nonexistent")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepository_ExistsByEmail(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserRepository(db)

	// Create test user
	createTestUser(t, repo, "testuser", "test@example.com", models.RoleUser)

	// Check existence
	exists, err := repo.ExistsByEmail("test@example.com")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Check non-existence
	exists, err = repo.ExistsByEmail("nonexistent@example.com")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepository_Update(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserRepository(db)

	// Create test user
	user := createTestUser(t, repo, "testuser", "test@example.com", models.RoleUser)

	// Update user
	user.Role = models.RoleAdministrator
	err := repo.Update(user)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.RoleAdministrator, retrieved.Role)
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserRepository(db)

	// Create test user
	user := createTestUser(t, repo, "testuser", "test@example.com", models.RoleUser)

	// Delete user
	err := repo.Delete(user.ID)
	assert.NoError(t, err)

	// Verify deletion
	retrieved, err := repo.GetByID(user.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, retrieved)
}

func TestUserRepository_List(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserRepository(db)

	// Create test users
	createTestUser(t, repo, "user1", "user1@example.com", models.RoleUser)
	createTestUser(t, repo, "user2", "user2@example.com", models.RoleAdministrator)
	createTestUser(t, repo, "user3", "user3@example.com", models.RoleCommenter)

	// List all users
	users, err := repo.List(nil, "", 0, 0)
	assert.NoError(t, err)
	assert.Len(t, users, 3)

	// List users by role
	filters := map[string]interface{}{"role": models.RoleUser}
	users, err = repo.List(filters, "", 0, 0)
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, models.RoleUser, users[0].Role)
}
