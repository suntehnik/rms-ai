package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"product-requirements-management/internal/models"
)

func TestNewService(t *testing.T) {
	secret := "test-secret"
	duration := time.Hour

	service := NewService(secret, duration)

	assert.NotNil(t, service)
	assert.Equal(t, []byte(secret), service.jwtSecret)
	assert.Equal(t, duration, service.tokenDuration)
}

func TestHashPassword(t *testing.T) {
	service := NewService("test-secret", time.Hour)
	password := "testpassword123"

	hash, err := service.HashPassword(password)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestVerifyPassword(t *testing.T) {
	service := NewService("test-secret", time.Hour)
	password := "testpassword123"

	hash, err := service.HashPassword(password)
	require.NoError(t, err)

	t.Run("valid password", func(t *testing.T) {
		err := service.VerifyPassword(password, hash)
		assert.NoError(t, err)
	})

	t.Run("invalid password", func(t *testing.T) {
		err := service.VerifyPassword("wrongpassword", hash)
		assert.Error(t, err)
	})
}

func TestGenerateToken(t *testing.T) {
	service := NewService("test-secret", time.Hour)
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     models.RoleUser,
	}

	token, err := service.GenerateToken(user)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateToken(t *testing.T) {
	service := NewService("test-secret", time.Hour)
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     models.RoleUser,
	}

	token, err := service.GenerateToken(user)
	require.NoError(t, err)

	t.Run("valid token", func(t *testing.T) {
		claims, err := service.ValidateToken(token)

		require.NoError(t, err)
		assert.Equal(t, user.ID.String(), claims.UserID)
		assert.Equal(t, user.Username, claims.Username)
		assert.Equal(t, user.Role, claims.Role)
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := service.ValidateToken("invalid-token")
		assert.Equal(t, ErrInvalidToken, err)
	})

	t.Run("expired token", func(t *testing.T) {
		// Create service with very short duration
		shortService := NewService("test-secret", time.Nanosecond)
		expiredToken, err := shortService.GenerateToken(user)
		require.NoError(t, err)

		// Wait for token to expire
		time.Sleep(time.Millisecond)

		_, err = shortService.ValidateToken(expiredToken)
		assert.Equal(t, ErrTokenExpired, err)
	})

	t.Run("token with different secret", func(t *testing.T) {
		differentService := NewService("different-secret", time.Hour)
		_, err := differentService.ValidateToken(token)
		assert.Equal(t, ErrInvalidToken, err)
	})
}

func TestCheckPermission(t *testing.T) {
	service := NewService("test-secret", time.Hour)

	testCases := []struct {
		name         string
		userRole     models.UserRole
		requiredRole models.UserRole
		expectError  bool
	}{
		{
			name:         "administrator can access administrator",
			userRole:     models.RoleAdministrator,
			requiredRole: models.RoleAdministrator,
			expectError:  false,
		},
		{
			name:         "administrator can access user",
			userRole:     models.RoleAdministrator,
			requiredRole: models.RoleUser,
			expectError:  false,
		},
		{
			name:         "administrator can access commenter",
			userRole:     models.RoleAdministrator,
			requiredRole: models.RoleCommenter,
			expectError:  false,
		},
		{
			name:         "user cannot access administrator",
			userRole:     models.RoleUser,
			requiredRole: models.RoleAdministrator,
			expectError:  true,
		},
		{
			name:         "user can access user",
			userRole:     models.RoleUser,
			requiredRole: models.RoleUser,
			expectError:  false,
		},
		{
			name:         "user can access commenter",
			userRole:     models.RoleUser,
			requiredRole: models.RoleCommenter,
			expectError:  false,
		},
		{
			name:         "commenter cannot access administrator",
			userRole:     models.RoleCommenter,
			requiredRole: models.RoleAdministrator,
			expectError:  true,
		},
		{
			name:         "commenter cannot access user",
			userRole:     models.RoleCommenter,
			requiredRole: models.RoleUser,
			expectError:  true,
		},
		{
			name:         "commenter can access commenter",
			userRole:     models.RoleCommenter,
			requiredRole: models.RoleCommenter,
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.CheckPermission(tc.userRole, tc.requiredRole)
			if tc.expectError {
				assert.Equal(t, ErrInsufficientRole, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCanEdit(t *testing.T) {
	service := NewService("test-secret", time.Hour)

	testCases := []struct {
		role     models.UserRole
		expected bool
	}{
		{models.RoleAdministrator, true},
		{models.RoleUser, true},
		{models.RoleCommenter, false},
	}

	for _, tc := range testCases {
		t.Run(string(tc.role), func(t *testing.T) {
			result := service.CanEdit(tc.role)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCanDelete(t *testing.T) {
	service := NewService("test-secret", time.Hour)

	testCases := []struct {
		role     models.UserRole
		expected bool
	}{
		{models.RoleAdministrator, true},
		{models.RoleUser, true},
		{models.RoleCommenter, false},
	}

	for _, tc := range testCases {
		t.Run(string(tc.role), func(t *testing.T) {
			result := service.CanDelete(tc.role)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCanManageUsers(t *testing.T) {
	service := NewService("test-secret", time.Hour)

	testCases := []struct {
		role     models.UserRole
		expected bool
	}{
		{models.RoleAdministrator, true},
		{models.RoleUser, false},
		{models.RoleCommenter, false},
	}

	for _, tc := range testCases {
		t.Run(string(tc.role), func(t *testing.T) {
			result := service.CanManageUsers(tc.role)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCanManageConfig(t *testing.T) {
	service := NewService("test-secret", time.Hour)

	testCases := []struct {
		role     models.UserRole
		expected bool
	}{
		{models.RoleAdministrator, true},
		{models.RoleUser, false},
		{models.RoleCommenter, false},
	}

	for _, tc := range testCases {
		t.Run(string(tc.role), func(t *testing.T) {
			result := service.CanManageConfig(tc.role)
			assert.Equal(t, tc.expected, result)
		})
	}
}
