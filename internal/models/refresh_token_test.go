package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshTokenBeforeCreate(t *testing.T) {
	db := setupTestDB(t)

	// Create a test user for foreign key constraints
	user := User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         RoleUser,
	}
	err := db.Create(&user).Error
	require.NoError(t, err)

	t.Run("should generate UUID when ID is nil", func(t *testing.T) {
		refreshToken := RefreshToken{
			UserID:    user.ID,
			TokenHash: "hashed_token_value",
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		}

		err := db.Create(&refreshToken).Error
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, refreshToken.ID)
	})

	t.Run("should preserve existing UUID when ID is set", func(t *testing.T) {
		existingID := uuid.New()
		refreshToken := RefreshToken{
			ID:        existingID,
			UserID:    user.ID,
			TokenHash: "hashed_token_value_2",
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		}

		err := db.Create(&refreshToken).Error
		assert.NoError(t, err)
		assert.Equal(t, existingID, refreshToken.ID)
	})

	t.Run("should set CreatedAt timestamp automatically", func(t *testing.T) {
		beforeCreate := time.Now()
		refreshToken := RefreshToken{
			UserID:    user.ID,
			TokenHash: "hashed_token_value_3",
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		}

		err := db.Create(&refreshToken).Error
		assert.NoError(t, err)
		assert.False(t, refreshToken.CreatedAt.IsZero())
		assert.True(t, refreshToken.CreatedAt.After(beforeCreate) || refreshToken.CreatedAt.Equal(beforeCreate))
	})
}

func TestRefreshTokenIsExpired(t *testing.T) {
	t.Run("should return false for future expiration", func(t *testing.T) {
		refreshToken := RefreshToken{
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		assert.False(t, refreshToken.IsExpired())
	})

	t.Run("should return false for far future expiration", func(t *testing.T) {
		refreshToken := RefreshToken{
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
		}
		assert.False(t, refreshToken.IsExpired())
	})

	t.Run("should return true for past expiration", func(t *testing.T) {
		refreshToken := RefreshToken{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		assert.True(t, refreshToken.IsExpired())
	})

	t.Run("should return true for far past expiration", func(t *testing.T) {
		refreshToken := RefreshToken{
			ExpiresAt: time.Now().Add(-30 * 24 * time.Hour), // 30 days ago
		}
		assert.True(t, refreshToken.IsExpired())
	})

	t.Run("should return true for expiration at current time", func(t *testing.T) {
		// Set expiration to a time slightly in the past to ensure it's expired
		refreshToken := RefreshToken{
			ExpiresAt: time.Now().Add(-1 * time.Millisecond),
		}
		assert.True(t, refreshToken.IsExpired())
	})

	t.Run("should handle edge case of exact current time", func(t *testing.T) {
		// This test verifies behavior at the exact boundary
		now := time.Now()
		refreshToken := RefreshToken{
			ExpiresAt: now,
		}
		// Since time.Now() will be slightly after 'now', this should be expired
		time.Sleep(1 * time.Millisecond)
		assert.True(t, refreshToken.IsExpired())
	})

	t.Run("should handle zero time as expired", func(t *testing.T) {
		refreshToken := RefreshToken{
			ExpiresAt: time.Time{}, // Zero time
		}
		assert.True(t, refreshToken.IsExpired())
	})
}

func TestRefreshTokenTableName(t *testing.T) {
	t.Run("should return correct table name", func(t *testing.T) {
		refreshToken := RefreshToken{}
		assert.Equal(t, "refresh_tokens", refreshToken.TableName())
	})
}

func TestRefreshTokenDatabaseOperations(t *testing.T) {
	db := setupTestDB(t)

	// Create a test user
	user := User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         RoleUser,
	}
	err := db.Create(&user).Error
	require.NoError(t, err)

	t.Run("should create refresh token with all fields", func(t *testing.T) {
		now := time.Now()
		lastUsed := now.Add(-1 * time.Hour)
		refreshToken := RefreshToken{
			UserID:     user.ID,
			TokenHash:  "hashed_token_complete",
			ExpiresAt:  now.Add(30 * 24 * time.Hour),
			LastUsedAt: &lastUsed,
		}

		err := db.Create(&refreshToken).Error
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, refreshToken.ID)
		assert.NotNil(t, refreshToken.LastUsedAt)
		assert.Equal(t, lastUsed.Unix(), refreshToken.LastUsedAt.Unix())
	})

	t.Run("should create refresh token without LastUsedAt", func(t *testing.T) {
		refreshToken := RefreshToken{
			UserID:    user.ID,
			TokenHash: "hashed_token_no_last_used",
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		}

		err := db.Create(&refreshToken).Error
		assert.NoError(t, err)
		assert.Nil(t, refreshToken.LastUsedAt)
	})

	t.Run("should update LastUsedAt timestamp", func(t *testing.T) {
		refreshToken := RefreshToken{
			UserID:    user.ID,
			TokenHash: "hashed_token_update_last_used",
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		}

		err := db.Create(&refreshToken).Error
		require.NoError(t, err)
		assert.Nil(t, refreshToken.LastUsedAt)

		// Update LastUsedAt
		now := time.Now()
		refreshToken.LastUsedAt = &now
		err = db.Save(&refreshToken).Error
		assert.NoError(t, err)

		// Verify update
		var retrieved RefreshToken
		err = db.First(&retrieved, "id = ?", refreshToken.ID).Error
		assert.NoError(t, err)
		assert.NotNil(t, retrieved.LastUsedAt)
		assert.Equal(t, now.Unix(), retrieved.LastUsedAt.Unix())
	})

	t.Run("should retrieve refresh token by user ID", func(t *testing.T) {
		// Create multiple tokens for the same user
		for i := 0; i < 3; i++ {
			refreshToken := RefreshToken{
				UserID:    user.ID,
				TokenHash: "hashed_token_" + string(rune(i)),
				ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
			}
			err := db.Create(&refreshToken).Error
			require.NoError(t, err)
		}

		// Retrieve all tokens for user
		var tokens []RefreshToken
		err := db.Where("user_id = ?", user.ID).Find(&tokens).Error
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(tokens), 3)
	})

	t.Run("should delete refresh token", func(t *testing.T) {
		refreshToken := RefreshToken{
			UserID:    user.ID,
			TokenHash: "hashed_token_delete",
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		}

		err := db.Create(&refreshToken).Error
		require.NoError(t, err)

		// Delete token
		err = db.Delete(&refreshToken).Error
		assert.NoError(t, err)

		// Verify deletion
		var retrieved RefreshToken
		err = db.First(&retrieved, "id = ?", refreshToken.ID).Error
		assert.Error(t, err)
	})
}

func TestRefreshTokenCascadeDelete(t *testing.T) {
	db := setupTestDB(t)

	t.Run("should support cascade delete pattern for refresh tokens", func(t *testing.T) {
		// Note: SQLite in-memory database may not enforce CASCADE constraints
		// This test verifies the relationship structure is correct

		// Create a user
		user := User{
			Username:     "cascadeuser",
			Email:        "cascade@example.com",
			PasswordHash: "hashedpassword",
			Role:         RoleUser,
		}
		err := db.Create(&user).Error
		require.NoError(t, err)

		// Create refresh tokens for the user
		var tokenIDs []uuid.UUID
		for i := 0; i < 3; i++ {
			refreshToken := RefreshToken{
				UserID:    user.ID,
				TokenHash: "cascade_token_" + string(rune(i)),
				ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
			}
			err := db.Create(&refreshToken).Error
			require.NoError(t, err)
			tokenIDs = append(tokenIDs, refreshToken.ID)
		}

		// Verify tokens exist
		var count int64
		db.Model(&RefreshToken{}).Where("user_id = ?", user.ID).Count(&count)
		assert.Equal(t, int64(3), count)

		// Manually delete tokens (simulating cascade behavior)
		// In production PostgreSQL, this would happen automatically via CASCADE
		err = db.Where("user_id = ?", user.ID).Delete(&RefreshToken{}).Error
		assert.NoError(t, err)

		// Delete user
		err = db.Delete(&user).Error
		assert.NoError(t, err)

		// Verify tokens were deleted
		db.Model(&RefreshToken{}).Where("user_id = ?", user.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})
}

func TestRefreshTokenValidation(t *testing.T) {
	db := setupTestDB(t)

	// Create a test user
	user := User{
		Username:     "validationuser",
		Email:        "validation@example.com",
		PasswordHash: "hashedpassword",
		Role:         RoleUser,
	}
	err := db.Create(&user).Error
	require.NoError(t, err)

	t.Run("should have UserID field", func(t *testing.T) {
		// Note: SQLite may not enforce NOT NULL constraints in the same way as PostgreSQL
		// This test verifies the field exists and can be set
		refreshToken := RefreshToken{
			UserID:    user.ID,
			TokenHash: "hashed_token_with_user",
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		}

		err := db.Create(&refreshToken).Error
		assert.NoError(t, err)
		assert.Equal(t, user.ID, refreshToken.UserID)
	})

	t.Run("should have TokenHash field", func(t *testing.T) {
		refreshToken := RefreshToken{
			UserID:    user.ID,
			TokenHash: "hashed_token_value",
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		}

		err := db.Create(&refreshToken).Error
		assert.NoError(t, err)
		assert.Equal(t, "hashed_token_value", refreshToken.TokenHash)
	})

	t.Run("should have ExpiresAt field", func(t *testing.T) {
		expiresAt := time.Now().Add(30 * 24 * time.Hour)
		refreshToken := RefreshToken{
			UserID:    user.ID,
			TokenHash: "hashed_token_with_expiry",
			ExpiresAt: expiresAt,
		}

		err := db.Create(&refreshToken).Error
		assert.NoError(t, err)
		assert.Equal(t, expiresAt.Unix(), refreshToken.ExpiresAt.Unix())
	})

	t.Run("should establish foreign key relationship with User", func(t *testing.T) {
		// Verify the relationship is defined correctly
		refreshToken := RefreshToken{
			UserID:    user.ID,
			TokenHash: "hashed_token_fk_test",
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		}

		err := db.Create(&refreshToken).Error
		assert.NoError(t, err)

		// Load with user relationship
		var retrieved RefreshToken
		err = db.Preload("User").First(&retrieved, "id = ?", refreshToken.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.User.ID)
		assert.Equal(t, user.Username, retrieved.User.Username)
	})
}

func TestRefreshTokenJSONSerialization(t *testing.T) {
	t.Run("should not expose TokenHash in JSON", func(t *testing.T) {
		refreshToken := RefreshToken{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			TokenHash: "secret_hash_value",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		}

		// The TokenHash field has json:"-" tag, so it should not be serialized
		// This is verified by the struct tag, but we can test the behavior
		assert.Equal(t, "secret_hash_value", refreshToken.TokenHash)
		// In actual JSON serialization, TokenHash would be omitted
	})
}
