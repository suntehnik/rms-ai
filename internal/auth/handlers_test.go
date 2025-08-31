package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"product-requirements-management/internal/models"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	// Migrate the schema
	err = db.AutoMigrate(&models.User{})
	require.NoError(t, err)
	
	return db
}

func setupTestHandlers(t *testing.T) (*Handlers, *gorm.DB, *gin.Engine) {
	db := setupTestDB(t)
	service := NewService("test-secret", time.Hour)
	handlers := NewHandlers(service, db)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	return handlers, db, router
}

func createTestUser(t *testing.T, db *gorm.DB, username, email string, role models.UserRole) *models.User {
	service := NewService("test-secret", time.Hour)
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
	handlers, db, router := setupTestHandlers(t)
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
	handlers, db, router := setupTestHandlers(t)
	service := NewService("test-secret", time.Hour)
	
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
	handlers, db, router := setupTestHandlers(t)
	service := NewService("test-secret", time.Hour)
	
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
	handlers, db, router := setupTestHandlers(t)
	service := NewService("test-secret", time.Hour)
	
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
	handlers, db, router := setupTestHandlers(t)
	service := NewService("test-secret", time.Hour)
	
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