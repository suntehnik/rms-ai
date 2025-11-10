package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"product-requirements-management/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&models.User{}, &models.RefreshToken{})
	require.NoError(t, err)

	return db
}

func setupTestHandlers(t *testing.T) (*Handlers, *gorm.DB, *gin.Engine, *Service) {
	db := setupTestDB(t)

	// Import repository package to avoid import cycle
	refreshTokenRepo := &mockRefreshTokenRepository{db: db}
	service := NewService("test-secret", time.Hour, refreshTokenRepo)
	handlers := NewHandlers(service, db)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	return handlers, db, router, service
}

// mockRefreshTokenRepository is a simple mock for testing
type mockRefreshTokenRepository struct {
	db *gorm.DB
}

func (m *mockRefreshTokenRepository) Create(token *models.RefreshToken) error {
	return m.db.Create(token).Error
}

func (m *mockRefreshTokenRepository) FindByTokenHash(tokenHash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := m.db.Where("token_hash = ?", tokenHash).First(&token).Error
	return &token, err
}

func (m *mockRefreshTokenRepository) FindByUserID(userID uuid.UUID) ([]*models.RefreshToken, error) {
	var tokens []*models.RefreshToken
	err := m.db.Where("user_id = ?", userID).Find(&tokens).Error
	return tokens, err
}

func (m *mockRefreshTokenRepository) FindAll() ([]*models.RefreshToken, error) {
	var tokens []*models.RefreshToken
	err := m.db.Find(&tokens).Error
	return tokens, err
}

func (m *mockRefreshTokenRepository) Update(token *models.RefreshToken) error {
	return m.db.Save(token).Error
}

func (m *mockRefreshTokenRepository) Delete(id uuid.UUID) error {
	return m.db.Delete(&models.RefreshToken{}, "id = ?", id).Error
}

func (m *mockRefreshTokenRepository) DeleteByUserID(userID uuid.UUID) error {
	return m.db.Delete(&models.RefreshToken{}, "user_id = ?", userID).Error
}

func (m *mockRefreshTokenRepository) DeleteExpired() (int64, error) {
	result := m.db.Where("expires_at < ?", time.Now()).Delete(&models.RefreshToken{})
	return result.RowsAffected, result.Error
}

func (m *mockRefreshTokenRepository) GetDB() *gorm.DB {
	return m.db
}

func createTestUser(t *testing.T, db *gorm.DB, username, email string, role models.UserRole) *models.User {
	service := NewService("test-secret", time.Hour, nil)
	passwordHash, err := service.HashPassword("testpassword123")
	require.NoError(t, err)

	user := &models.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
	}

	err = db.Create(user).Error
	require.NoError(t, err)

	return user
}

func TestLogin(t *testing.T) {
	handlers, db, router, _ := setupTestHandlers(t)
	router.POST("/login", handlers.Login)

	// Create test user
	user := createTestUser(t, db, "testuser", "test@example.com", models.RoleUser)

	t.Run("successful login", func(t *testing.T) {
		loginReq := LoginRequest{
			Username: "testuser",
			Password: "testpassword123",
		}

		body, err := json.Marshal(loginReq)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response LoginResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.Token)
		assert.NotEmpty(t, response.RefreshToken)
		assert.Equal(t, user.ID.String(), response.User.ID)
		assert.Equal(t, user.Username, response.User.Username)
		assert.Equal(t, user.Email, response.User.Email)
		assert.Equal(t, user.Role, response.User.Role)
	})

	t.Run("invalid username", func(t *testing.T) {
		loginReq := LoginRequest{
			Username: "nonexistent",
			Password: "testpassword123",
		}

		body, err := json.Marshal(loginReq)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid credentials")
	})

	t.Run("invalid password", func(t *testing.T) {
		loginReq := LoginRequest{
			Username: "testuser",
			Password: "wrongpassword",
		}

		body, err := json.Marshal(loginReq)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid credentials")
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCreateUser(t *testing.T) {
	handlers, db, router, service := setupTestHandlers(t)

	router.POST("/users", service.Middleware(), service.RequireAdministrator(), handlers.CreateUser)

	// Create admin user for authentication
	adminUser := createTestUser(t, db, "admin", "admin@example.com", models.RoleAdministrator)
	adminToken, err := service.GenerateToken(adminUser)
	require.NoError(t, err)

	t.Run("successful user creation", func(t *testing.T) {
		createReq := CreateUserRequest{
			Username: "newuser",
			Email:    "newuser@example.com",
			Password: "newpassword123",
			Role:     models.RoleUser,
		}

		body, err := json.Marshal(createReq)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response UserResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "newuser", response.Username)
		assert.Equal(t, "newuser@example.com", response.Email)
		assert.Equal(t, models.RoleUser, response.Role)

		// Verify user was created in database
		var user models.User
		err = db.Where("username = ?", "newuser").First(&user).Error
		assert.NoError(t, err)
	})

	t.Run("duplicate username", func(t *testing.T) {
		createReq := CreateUserRequest{
			Username: "admin", // Already exists
			Email:    "different@example.com",
			Password: "newpassword123",
			Role:     models.RoleUser,
		}

		body, err := json.Marshal(createReq)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "already exists")
	})

	t.Run("invalid role", func(t *testing.T) {
		createReq := CreateUserRequest{
			Username: "invalidrole",
			Email:    "invalidrole@example.com",
			Password: "newpassword123",
			Role:     "InvalidRole",
		}

		body, err := json.Marshal(createReq)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid role")
	})
}

func TestGetUsers(t *testing.T) {
	handlers, db, router, service := setupTestHandlers(t)

	router.GET("/users", service.Middleware(), service.RequireAdministrator(), handlers.GetUsers)

	// Create admin user for authentication
	adminUser := createTestUser(t, db, "admin", "admin@example.com", models.RoleAdministrator)
	adminToken, err := service.GenerateToken(adminUser)
	require.NoError(t, err)

	// Create additional test users
	createTestUser(t, db, "user1", "user1@example.com", models.RoleUser)
	createTestUser(t, db, "user2", "user2@example.com", models.RoleCommenter)

	req := httptest.NewRequest("GET", "/users", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []UserResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response, 3) // admin + user1 + user2
}

func TestGetProfile(t *testing.T) {
	handlers, db, router, service := setupTestHandlers(t)

	router.GET("/profile", service.Middleware(), handlers.GetProfile)

	// Create test user
	user := createTestUser(t, db, "testuser", "test@example.com", models.RoleUser)
	token, err := service.GenerateToken(user)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response UserResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, user.ID.String(), response.ID)
	assert.Equal(t, user.Username, response.Username)
	assert.Equal(t, user.Email, response.Email)
	assert.Equal(t, user.Role, response.Role)
}

func TestChangePassword(t *testing.T) {
	handlers, db, router, service := setupTestHandlers(t)

	router.PUT("/change-password", service.Middleware(), handlers.ChangePassword)

	// Create test user
	user := createTestUser(t, db, "testuser", "test@example.com", models.RoleUser)
	token, err := service.GenerateToken(user)
	require.NoError(t, err)

	t.Run("successful password change", func(t *testing.T) {
		changeReq := ChangePasswordRequest{
			CurrentPassword: "testpassword123",
			NewPassword:     "newpassword456",
		}

		body, err := json.Marshal(changeReq)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/change-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Password changed successfully")

		// Verify password was changed
		var updatedUser models.User
		err = db.Where("id = ?", user.ID).First(&updatedUser).Error
		require.NoError(t, err)

		err = service.VerifyPassword("newpassword456", updatedUser.PasswordHash)
		assert.NoError(t, err)
	})

	t.Run("invalid current password", func(t *testing.T) {
		changeReq := ChangePasswordRequest{
			CurrentPassword: "wrongpassword",
			NewPassword:     "newpassword456",
		}

		body, err := json.Marshal(changeReq)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/change-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid current password")
	})
}

func TestRefreshToken(t *testing.T) {
	handlers, db, router, service := setupTestHandlers(t)
	router.POST("/refresh", handlers.RefreshToken)

	// Create test user
	user := createTestUser(t, db, "testuser", "test@example.com", models.RoleUser)

	t.Run("successful token refresh", func(t *testing.T) {
		// Generate initial refresh token
		refreshToken, err := service.GenerateRefreshToken(nil, user)
		require.NoError(t, err)

		refreshReq := RefreshRequest{
			RefreshToken: refreshToken,
		}

		body, err := json.Marshal(refreshReq)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response RefreshResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.Token)
		assert.NotEmpty(t, response.RefreshToken)
		assert.NotEqual(t, refreshToken, response.RefreshToken) // Token rotation
	})

	t.Run("expired refresh token", func(t *testing.T) {
		// Create an expired refresh token
		expiredToken := &models.RefreshToken{
			ID:        uuid.New(),
			UserID:    user.ID,
			TokenHash: "expired_hash",
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		}
		err := db.Create(expiredToken).Error
		require.NoError(t, err)

		refreshReq := RefreshRequest{
			RefreshToken: "expired_token",
		}

		body, err := json.Marshal(refreshReq)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "INVALID_REFRESH_TOKEN")
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		refreshReq := RefreshRequest{
			RefreshToken: "invalid_token_that_does_not_exist",
		}

		body, err := json.Marshal(refreshReq)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "INVALID_REFRESH_TOKEN")
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
	})
}

func TestLogout(t *testing.T) {
	handlers, db, router, service := setupTestHandlers(t)
	router.POST("/logout", handlers.Logout)

	// Create test user
	user := createTestUser(t, db, "testuser", "test@example.com", models.RoleUser)

	t.Run("successful logout", func(t *testing.T) {
		// Generate refresh token
		refreshToken, err := service.GenerateRefreshToken(nil, user)
		require.NoError(t, err)

		logoutReq := LogoutRequest{
			RefreshToken: refreshToken,
		}

		body, err := json.Marshal(logoutReq)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/logout", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify token was deleted
		var count int64
		db.Model(&models.RefreshToken{}).Where("user_id = ?", user.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		logoutReq := LogoutRequest{
			RefreshToken: "invalid_token_that_does_not_exist",
		}

		body, err := json.Marshal(logoutReq)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/logout", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "INVALID_REFRESH_TOKEN")
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/logout", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
	})
}
