package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
)

func setupPATRepositoryTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate models
	err = db.AutoMigrate(&models.User{}, &models.PersonalAccessToken{})
	require.NoError(t, err)

	return db
}

func createTestUserForPATRepo(t *testing.T, db *gorm.DB, username string) *models.User {
	user := &models.User{
		Username:     username,
		Email:        username + "@example.com",
		PasswordHash: "hashedpassword",
		Role:         models.RoleUser,
	}

	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

func createTestPAT(t *testing.T, repo PersonalAccessTokenRepository, userID uuid.UUID, name string) *models.PersonalAccessToken {
	pat := &models.PersonalAccessToken{
		UserID:    userID,
		Name:      name,
		TokenHash: "hashed_token_" + name,
		Prefix:    "mcp_pat_",
		Scopes:    `["full_access"]`,
	}

	err := repo.Create(pat)
	require.NoError(t, err)
	return pat
}

func TestPersonalAccessTokenRepository_Create(t *testing.T) {
	db := setupPATRepositoryTestDB(t)
	repo := NewPersonalAccessTokenRepository(db)
	user := createTestUserForPATRepo(t, db, "testuser")

	pat := &models.PersonalAccessToken{
		UserID:    user.ID,
		Name:      "Test Token",
		TokenHash: "hashed_token_value",
		Prefix:    "mcp_pat_",
		Scopes:    `["full_access"]`,
	}

	err := repo.Create(pat)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, pat.ID)
}

func TestPersonalAccessTokenRepository_GetByID(t *testing.T) {
	db := setupPATRepositoryTestDB(t)
	repo := NewPersonalAccessTokenRepository(db)
	user := createTestUserForPATRepo(t, db, "testuser")

	// Create test PAT
	originalPAT := createTestPAT(t, repo, user.ID, "Test Token")

	// Retrieve by ID
	retrievedPAT, err := repo.GetByID(originalPAT.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedPAT)
	assert.Equal(t, originalPAT.ID, retrievedPAT.ID)
	assert.Equal(t, originalPAT.Name, retrievedPAT.Name)
	assert.Equal(t, originalPAT.UserID, retrievedPAT.UserID)
}

func TestPersonalAccessTokenRepository_GetByUserID(t *testing.T) {
	db := setupPATRepositoryTestDB(t)
	repo := NewPersonalAccessTokenRepository(db)
	user1 := createTestUserForPATRepo(t, db, "user1")
	user2 := createTestUserForPATRepo(t, db, "user2")

	// Create PATs for user1
	createTestPAT(t, repo, user1.ID, "Token 1")
	createTestPAT(t, repo, user1.ID, "Token 2")

	// Create PAT for user2
	createTestPAT(t, repo, user2.ID, "Token 3")

	// Get PATs for user1
	user1PATs, err := repo.GetByUserID(user1.ID)
	assert.NoError(t, err)
	assert.Len(t, user1PATs, 2)

	// Get PATs for user2
	user2PATs, err := repo.GetByUserID(user2.ID)
	assert.NoError(t, err)
	assert.Len(t, user2PATs, 1)
}
func TestPersonalAccessTokenRepository_GetByUserIDWithPagination(t *testing.T) {
	db := setupPATRepositoryTestDB(t)
	repo := NewPersonalAccessTokenRepository(db)
	user := createTestUserForPATRepo(t, db, "testuser")

	// Create 5 PATs
	for i := 1; i <= 5; i++ {
		createTestPAT(t, repo, user.ID, fmt.Sprintf("Token %d", i))
	}

	// Test pagination
	pats, total, err := repo.GetByUserIDWithPagination(user.ID, 2, 0)
	assert.NoError(t, err)
	assert.Len(t, pats, 2)
	assert.Equal(t, int64(5), total)

	// Test second page
	pats, total, err = repo.GetByUserIDWithPagination(user.ID, 2, 2)
	assert.NoError(t, err)
	assert.Len(t, pats, 2)
	assert.Equal(t, int64(5), total)

	// Test last page
	pats, total, err = repo.GetByUserIDWithPagination(user.ID, 2, 4)
	assert.NoError(t, err)
	assert.Len(t, pats, 1)
	assert.Equal(t, int64(5), total)
}

func TestPersonalAccessTokenRepository_GetHashesByPrefix(t *testing.T) {
	db := setupPATRepositoryTestDB(t)
	repo := NewPersonalAccessTokenRepository(db)
	user := createTestUserForPATRepo(t, db, "testuser")

	// Create PATs with different prefixes
	pat1 := createTestPAT(t, repo, user.ID, "Token 1")
	pat1.Prefix = "mcp_pat_"
	repo.Update(pat1)

	pat2 := createTestPAT(t, repo, user.ID, "Token 2")
	pat2.Prefix = "other_"
	repo.Update(pat2)

	pat3 := createTestPAT(t, repo, user.ID, "Token 3")
	pat3.Prefix = "mcp_pat_"
	repo.Update(pat3)

	// Get PATs by prefix
	mcpPATs, err := repo.GetHashesByPrefix("mcp_pat_")
	assert.NoError(t, err)
	assert.Len(t, mcpPATs, 2)

	otherPATs, err := repo.GetHashesByPrefix("other_")
	assert.NoError(t, err)
	assert.Len(t, otherPATs, 1)
}

func TestPersonalAccessTokenRepository_UpdateLastUsed(t *testing.T) {
	db := setupPATRepositoryTestDB(t)
	repo := NewPersonalAccessTokenRepository(db)
	user := createTestUserForPATRepo(t, db, "testuser")

	pat := createTestPAT(t, repo, user.ID, "Test Token")
	assert.Nil(t, pat.LastUsedAt)

	// Update last used timestamp
	now := time.Now()
	err := repo.UpdateLastUsed(pat.ID, &now)
	assert.NoError(t, err)

	// Verify update
	updatedPAT, err := repo.GetByID(pat.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedPAT.LastUsedAt)
	assert.WithinDuration(t, now, *updatedPAT.LastUsedAt, time.Second)
}

func TestPersonalAccessTokenRepository_DeleteExpired(t *testing.T) {
	db := setupPATRepositoryTestDB(t)
	repo := NewPersonalAccessTokenRepository(db)
	user := createTestUserForPATRepo(t, db, "testuser")

	// Create expired PAT
	expiredPAT := createTestPAT(t, repo, user.ID, "Expired Token")
	pastTime := time.Now().Add(-24 * time.Hour)
	expiredPAT.ExpiresAt = &pastTime
	repo.Update(expiredPAT)

	// Create non-expired PAT
	validPAT := createTestPAT(t, repo, user.ID, "Valid Token")
	futureTime := time.Now().Add(24 * time.Hour)
	validPAT.ExpiresAt = &futureTime
	repo.Update(validPAT)

	// Create PAT with no expiration
	createTestPAT(t, repo, user.ID, "No Expiration Token")

	// Delete expired tokens
	deletedCount, err := repo.DeleteExpired()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), deletedCount)

	// Verify only expired token was deleted
	allPATs, err := repo.GetByUserID(user.ID)
	assert.NoError(t, err)
	assert.Len(t, allPATs, 2)
}

func TestPersonalAccessTokenRepository_ExistsByUserIDAndName(t *testing.T) {
	db := setupPATRepositoryTestDB(t)
	repo := NewPersonalAccessTokenRepository(db)
	user := createTestUserForPATRepo(t, db, "testuser")

	// Create PAT
	createTestPAT(t, repo, user.ID, "Existing Token")

	// Test existing token
	exists, err := repo.ExistsByUserIDAndName(user.ID, "Existing Token")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test non-existing token
	exists, err = repo.ExistsByUserIDAndName(user.ID, "Non-existing Token")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Test with different user
	user2 := createTestUserForPATRepo(t, db, "user2")
	exists, err = repo.ExistsByUserIDAndName(user2.ID, "Existing Token")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestPersonalAccessTokenRepository_Delete(t *testing.T) {
	db := setupPATRepositoryTestDB(t)
	repo := NewPersonalAccessTokenRepository(db)
	user := createTestUserForPATRepo(t, db, "testuser")

	pat := createTestPAT(t, repo, user.ID, "Test Token")

	// Delete PAT
	err := repo.Delete(pat.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(pat.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}
