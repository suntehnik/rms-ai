package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
)

func setupRefreshTokenTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate models
	err = db.AutoMigrate(&models.User{}, &models.RefreshToken{})
	require.NoError(t, err)

	return db
}

func createTestUserForRefreshToken(t *testing.T, db *gorm.DB) *models.User {
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}

	err := db.Create(user).Error
	require.NoError(t, err)

	return user
}

func createTestRefreshToken(t *testing.T, repo RefreshTokenRepository, userID uuid.UUID, tokenHash string, expiresAt time.Time) *models.RefreshToken {
	token := &models.RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	err := repo.Create(token)
	require.NoError(t, err)

	return token
}

func TestRefreshTokenRepository_Create(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)
	user := createTestUserForRefreshToken(t, db)

	token := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: "hashed_token_value",
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}

	err := repo.Create(token)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, token.ID)
	assert.NotZero(t, token.CreatedAt)
}

func TestRefreshTokenRepository_FindByTokenHash_Existing(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)
	user := createTestUserForRefreshToken(t, db)

	// Create test token
	tokenHash := "hashed_token_value"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	created := createTestRefreshToken(t, repo, user.ID, tokenHash, expiresAt)

	// Find by token hash
	retrieved, err := repo.FindByTokenHash(tokenHash)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.UserID, retrieved.UserID)
	assert.Equal(t, created.TokenHash, retrieved.TokenHash)
}

func TestRefreshTokenRepository_FindByTokenHash_NotFound(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)

	// Try to find non-existent token
	retrieved, err := repo.FindByTokenHash("nonexistent_hash")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, retrieved)
}

func TestRefreshTokenRepository_FindByUserID_MultipleTokens(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)
	user := createTestUserForRefreshToken(t, db)

	// Create multiple tokens for the same user
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	token1 := createTestRefreshToken(t, repo, user.ID, "hash1", expiresAt)
	time.Sleep(10 * time.Millisecond) // Ensure different created_at timestamps
	token2 := createTestRefreshToken(t, repo, user.ID, "hash2", expiresAt)
	time.Sleep(10 * time.Millisecond)
	token3 := createTestRefreshToken(t, repo, user.ID, "hash3", expiresAt)

	// Find all tokens for user
	tokens, err := repo.FindByUserID(user.ID)
	assert.NoError(t, err)
	assert.Len(t, tokens, 3)

	// Verify tokens are ordered by created_at DESC (newest first)
	assert.Equal(t, token3.ID, tokens[0].ID)
	assert.Equal(t, token2.ID, tokens[1].ID)
	assert.Equal(t, token1.ID, tokens[2].ID)
}

func TestRefreshTokenRepository_FindByUserID_NoTokens(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)
	user := createTestUserForRefreshToken(t, db)

	// Find tokens for user with no tokens
	tokens, err := repo.FindByUserID(user.ID)
	assert.NoError(t, err)
	assert.Empty(t, tokens)
}

func TestRefreshTokenRepository_Update(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)
	user := createTestUserForRefreshToken(t, db)

	// Create test token
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	token := createTestRefreshToken(t, repo, user.ID, "hash1", expiresAt)

	// Update last_used_at
	now := time.Now()
	token.LastUsedAt = &now

	err := repo.Update(token)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.FindByTokenHash("hash1")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved.LastUsedAt)
	assert.WithinDuration(t, now, *retrieved.LastUsedAt, time.Second)
}

func TestRefreshTokenRepository_Delete(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)
	user := createTestUserForRefreshToken(t, db)

	// Create test token
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	token := createTestRefreshToken(t, repo, user.ID, "hash1", expiresAt)

	// Delete token
	err := repo.Delete(token.ID)
	assert.NoError(t, err)

	// Verify deletion
	retrieved, err := repo.FindByTokenHash("hash1")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, retrieved)
}

func TestRefreshTokenRepository_Delete_NotFound(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)

	// Try to delete non-existent token
	err := repo.Delete(uuid.New())
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

func TestRefreshTokenRepository_DeleteByUserID(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)
	user := createTestUserForRefreshToken(t, db)

	// Create multiple tokens for the user
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	createTestRefreshToken(t, repo, user.ID, "hash1", expiresAt)
	createTestRefreshToken(t, repo, user.ID, "hash2", expiresAt)
	createTestRefreshToken(t, repo, user.ID, "hash3", expiresAt)

	// Delete all tokens for user
	err := repo.DeleteByUserID(user.ID)
	assert.NoError(t, err)

	// Verify all tokens are deleted
	tokens, err := repo.FindByUserID(user.ID)
	assert.NoError(t, err)
	assert.Empty(t, tokens)
}

func TestRefreshTokenRepository_DeleteExpired(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)
	user := createTestUserForRefreshToken(t, db)

	// Create expired tokens
	expiredTime := time.Now().Add(-1 * time.Hour)
	createTestRefreshToken(t, repo, user.ID, "expired1", expiredTime)
	createTestRefreshToken(t, repo, user.ID, "expired2", expiredTime)

	// Create valid tokens
	validTime := time.Now().Add(30 * 24 * time.Hour)
	validToken1 := createTestRefreshToken(t, repo, user.ID, "valid1", validTime)
	validToken2 := createTestRefreshToken(t, repo, user.ID, "valid2", validTime)

	// Delete expired tokens
	count, err := repo.DeleteExpired()
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Verify only valid tokens remain
	tokens, err := repo.FindByUserID(user.ID)
	assert.NoError(t, err)
	assert.Len(t, tokens, 2)

	// Verify the remaining tokens are the valid ones
	tokenIDs := []uuid.UUID{tokens[0].ID, tokens[1].ID}
	assert.Contains(t, tokenIDs, validToken1.ID)
	assert.Contains(t, tokenIDs, validToken2.ID)
}

func TestRefreshTokenRepository_DeleteExpired_NoExpiredTokens(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)
	user := createTestUserForRefreshToken(t, db)

	// Create only valid tokens
	validTime := time.Now().Add(30 * 24 * time.Hour)
	createTestRefreshToken(t, repo, user.ID, "valid1", validTime)
	createTestRefreshToken(t, repo, user.ID, "valid2", validTime)

	// Delete expired tokens
	count, err := repo.DeleteExpired()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Verify all tokens remain
	tokens, err := repo.FindByUserID(user.ID)
	assert.NoError(t, err)
	assert.Len(t, tokens, 2)
}
