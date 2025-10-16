package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"product-requirements-management/internal/models"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestAuthMiddleware(t *testing.T) {
	service := NewService("test-secret", time.Hour)
	router := setupTestRouter()

	// Test endpoint that requires authentication
	router.GET("/protected", service.Middleware(), func(c *gin.Context) {
		claims, exists := GetCurrentUser(c)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Claims not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": claims.UserID})
	})

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     models.RoleUser,
	}

	validToken, err := service.GenerateToken(user)
	require.NoError(t, err)

	t.Run("no authorization header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization header required")
	})

	t.Run("invalid authorization header format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Bearer token required")
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid token")
	})

	t.Run("expired token", func(t *testing.T) {
		shortService := NewService("test-secret", time.Nanosecond)
		expiredToken, err := shortService.GenerateToken(user)
		require.NoError(t, err)

		time.Sleep(time.Millisecond)

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Token expired")
	})

	t.Run("valid token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), user.ID.String())
	})
}

func TestRequireRole(t *testing.T) {
	service := NewService("test-secret", time.Hour)
	router := setupTestRouter()

	// Test endpoints with different role requirements
	router.GET("/admin", service.Middleware(), service.RequireAdministrator(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access"})
	})

	router.GET("/user", service.Middleware(), service.RequireUser(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "user access"})
	})

	router.GET("/commenter", service.Middleware(), service.RequireCommenter(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "commenter access"})
	})

	// Create users with different roles
	adminUser := &models.User{
		ID:       uuid.New(),
		Username: "admin",
		Role:     models.RoleAdministrator,
	}

	regularUser := &models.User{
		ID:       uuid.New(),
		Username: "user",
		Role:     models.RoleUser,
	}

	commenterUser := &models.User{
		ID:       uuid.New(),
		Username: "commenter",
		Role:     models.RoleCommenter,
	}

	adminToken, err := service.GenerateToken(adminUser)
	require.NoError(t, err)

	userToken, err := service.GenerateToken(regularUser)
	require.NoError(t, err)

	commenterToken, err := service.GenerateToken(commenterUser)
	require.NoError(t, err)

	testCases := []struct {
		name           string
		endpoint       string
		token          string
		expectedStatus int
	}{
		// Admin endpoint tests
		{"admin can access admin endpoint", "/admin", adminToken, http.StatusOK},
		{"user cannot access admin endpoint", "/admin", userToken, http.StatusForbidden},
		{"commenter cannot access admin endpoint", "/admin", commenterToken, http.StatusForbidden},

		// User endpoint tests
		{"admin can access user endpoint", "/user", adminToken, http.StatusOK},
		{"user can access user endpoint", "/user", userToken, http.StatusOK},
		{"commenter cannot access user endpoint", "/user", commenterToken, http.StatusForbidden},

		// Commenter endpoint tests
		{"admin can access commenter endpoint", "/commenter", adminToken, http.StatusOK},
		{"user can access commenter endpoint", "/commenter", userToken, http.StatusOK},
		{"commenter can access commenter endpoint", "/commenter", commenterToken, http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.endpoint, nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestGetCurrentUser(t *testing.T) {
	service := NewService("test-secret", time.Hour)
	router := setupTestRouter()

	router.GET("/test", service.Middleware(), func(c *gin.Context) {
		claims, exists := GetCurrentUser(c)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Claims not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"user_id":  claims.UserID,
			"username": claims.Username,
			"role":     claims.Role,
		})
	})

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     models.RoleUser,
	}

	token, err := service.GenerateToken(user)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), user.ID.String())
	assert.Contains(t, w.Body.String(), user.Username)
	assert.Contains(t, w.Body.String(), string(user.Role))
}

func TestGetCurrentUserID(t *testing.T) {
	service := NewService("test-secret", time.Hour)
	router := setupTestRouter()

	router.GET("/test", service.Middleware(), func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     models.RoleUser,
	}

	token, err := service.GenerateToken(user)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), user.ID.String())
}

func TestGetCurrentUserRole(t *testing.T) {
	service := NewService("test-secret", time.Hour)
	router := setupTestRouter()

	router.GET("/test", service.Middleware(), func(c *gin.Context) {
		role, exists := GetCurrentUserRole(c)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Role not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"role": role})
	})

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     models.RoleAdministrator,
	}

	token, err := service.GenerateToken(user)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), string(user.Role))
}
