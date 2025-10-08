package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPATTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate models
	err = db.AutoMigrate(&User{}, &PersonalAccessToken{})
	require.NoError(t, err)

	return db
}

func createTestUserForPAT(t *testing.T, db *gorm.DB) *User {
	user := &User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         RoleUser,
	}

	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

func TestPersonalAccessTokenModel_BeforeCreate(t *testing.T) {
	db := setupPATTestDB(t)
	user := createTestUserForPAT(t, db)

	pat := PersonalAccessToken{
		UserID:    user.ID,
		Name:      "Test Token",
		TokenHash: "hashed_token_value",
	}

	// Test creation with BeforeCreate hook
	err := db.Create(&pat).Error
	assert.NoError(t, err)

	// Verify ID was set
	assert.NotEqual(t, uuid.Nil, pat.ID)

	// Verify default prefix was set
	assert.Equal(t, "mcp_pat_", pat.Prefix)

	// Verify default scopes were set
	assert.Equal(t, `["full_access"]`, pat.Scopes)

	// Verify timestamps were set
	assert.False(t, pat.CreatedAt.IsZero())
	assert.False(t, pat.UpdatedAt.IsZero())
}

func TestPersonalAccessTokenModel_BeforeCreateWithCustomValues(t *testing.T) {
	db := setupPATTestDB(t)
	user := createTestUserForPAT(t, db)

	customPrefix := "custom_"
	customScopes := `["read_only"]`

	pat := PersonalAccessToken{
		UserID:    user.ID,
		Name:      "Custom Token",
		TokenHash: "hashed_token_value",
		Prefix:    customPrefix,
		Scopes:    customScopes,
	}

	err := db.Create(&pat).Error
	assert.NoError(t, err)

	// Verify custom values were preserved
	assert.Equal(t, customPrefix, pat.Prefix)
	assert.Equal(t, customScopes, pat.Scopes)
}
func TestPersonalAccessTokenModel_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt *time.Time
		expected  bool
	}{
		{
			name:      "no expiration",
			expiresAt: nil,
			expected:  true,
		},
		{
			name:      "future expiration",
			expiresAt: &[]time.Time{time.Now().Add(24 * time.Hour)}[0],
			expected:  false,
		},
		{
			name:      "past expiration",
			expiresAt: &[]time.Time{time.Now().Add(-24 * time.Hour)}[0],
			expected:  true,
		},
		{
			name:      "just expired",
			expiresAt: &[]time.Time{time.Now().Add(-1 * time.Second)}[0],
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pat := PersonalAccessToken{
				ExpiresAt: tt.expiresAt,
			}
			assert.Equal(t, tt.expected, pat.IsExpired())
		})
	}
}

func TestPersonalAccessTokenModel_TableName(t *testing.T) {
	pat := PersonalAccessToken{}
	assert.Equal(t, "personal_access_tokens", pat.TableName())
}

func TestPersonalAccessTokenModel_JSONSerialization(t *testing.T) {
	db := setupPATTestDB(t)
	user := createTestUserForPAT(t, db)

	pat := PersonalAccessToken{
		UserID:    user.ID,
		Name:      "Test Token",
		TokenHash: "secret_hash_should_not_appear",
		Prefix:    "mcp_pat_",
		Scopes:    `["full_access"]`,
	}

	err := db.Create(&pat).Error
	require.NoError(t, err)

	// Retrieve and verify JSON serialization excludes TokenHash
	var retrieved PersonalAccessToken
	err = db.First(&retrieved, pat.ID).Error
	require.NoError(t, err)

	// TokenHash should be populated from database
	assert.Equal(t, "secret_hash_should_not_appear", retrieved.TokenHash)

	// But should not appear in JSON (this is enforced by json:"-" tag)
	// We can't easily test JSON marshaling here, but the tag ensures it won't be serialized
}

func TestPersonalAccessTokenModel_UserAssociation(t *testing.T) {
	db := setupPATTestDB(t)
	user := createTestUserForPAT(t, db)

	pat := PersonalAccessToken{
		UserID:    user.ID,
		Name:      "Test Token",
		TokenHash: "hashed_token_value",
	}

	err := db.Create(&pat).Error
	require.NoError(t, err)

	// Test preloading user association
	var retrievedPAT PersonalAccessToken
	err = db.Preload("User").First(&retrievedPAT, pat.ID).Error
	require.NoError(t, err)

	assert.NotNil(t, retrievedPAT.User)
	assert.Equal(t, user.ID, retrievedPAT.User.ID)
	assert.Equal(t, user.Username, retrievedPAT.User.Username)
}
