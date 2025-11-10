package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"product-requirements-management/internal/config"
	"product-requirements-management/internal/logger"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPATService is a mock implementation of the PATService interface
type MockPATService struct {
	mock.Mock
}

func (m *MockPATService) CreatePAT(ctx context.Context, userID uuid.UUID, req service.CreatePATRequest) (*service.PATCreateResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.PATCreateResponse), args.Error(1)
}

func (m *MockPATService) ListUserPATs(ctx context.Context, userID uuid.UUID, limit, offset int) (*service.ListResponse[models.PersonalAccessToken], error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ListResponse[models.PersonalAccessToken]), args.Error(1)
}

func (m *MockPATService) GetPAT(ctx context.Context, patID, userID uuid.UUID) (*models.PersonalAccessToken, error) {
	args := m.Called(ctx, patID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PersonalAccessToken), args.Error(1)
}

func (m *MockPATService) RevokePAT(ctx context.Context, patID, userID uuid.UUID) error {
	args := m.Called(ctx, patID, userID)
	return args.Error(0)
}

func (m *MockPATService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockPATService) UpdateLastUsed(ctx context.Context, patID uuid.UUID) error {
	args := m.Called(ctx, patID)
	return args.Error(0)
}

func (m *MockPATService) CleanupExpiredTokens(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func TestPATMiddleware_MissingAuthorizationHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	authService := NewService("test-secret", time.Hour, nil)
	mockPATService := &MockPATService{}

	router := gin.New()
	router.Use(PATMiddleware(authService, mockPATService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization header required")
}

func TestPATMiddleware_InvalidBearerFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	authService := NewService("test-secret", time.Hour, nil)
	mockPATService := &MockPATService{}

	router := gin.New()
	router.Use(PATMiddleware(authService, mockPATService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic invalid")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Bearer token required")
}

func TestPATMiddleware_ValidPATToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Initialize logger for tests
	logger.Init(&config.LogConfig{
		Level:  "error", // Use error level to reduce test noise
		Format: "text",
	})

	// Setup
	authService := NewService("test-secret", time.Hour, nil)
	mockPATService := &MockPATService{}

	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	mockPATService.On("ValidateToken", mock.Anything, "mcp_pat_validtoken123").Return(testUser, nil)

	router := gin.New()
	router.Use(PATMiddleware(authService, mockPATService))
	router.GET("/test", func(c *gin.Context) {
		// Verify context is set correctly
		claims, exists := GetCurrentUser(c)
		assert.True(t, exists)
		assert.Equal(t, testUser.ID.String(), claims.UserID)
		assert.Equal(t, testUser.Username, claims.Username)
		assert.Equal(t, testUser.Role, claims.Role)

		// Verify auth method
		method, exists := GetAuthMethod(c)
		assert.True(t, exists)
		assert.Equal(t, "pat", method)

		// Verify user in context
		user, exists := GetUserFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, testUser.ID, user.ID)

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer mcp_pat_validtoken123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	mockPATService.AssertExpectations(t)
}

func TestPATMiddleware_InvalidPATToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	authService := NewService("test-secret", time.Hour, nil)
	mockPATService := &MockPATService{}

	mockPATService.On("ValidateToken", mock.Anything, "mcp_pat_invalidtoken").Return(nil, errors.New("invalid token"))

	router := gin.New()
	router.Use(PATMiddleware(authService, mockPATService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer mcp_pat_invalidtoken")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid token")
	mockPATService.AssertExpectations(t)
}

func TestPATMiddleware_JWTFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	authService := NewService("test-secret", time.Hour, nil)
	mockPATService := &MockPATService{}

	// Create a valid JWT token
	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	jwtToken, err := authService.GenerateToken(testUser)
	assert.NoError(t, err)

	router := gin.New()
	router.Use(PATMiddleware(authService, mockPATService))
	router.GET("/test", func(c *gin.Context) {
		// Verify context is set correctly for JWT
		claims, exists := GetCurrentUser(c)
		assert.True(t, exists)
		assert.Equal(t, testUser.ID.String(), claims.UserID)
		assert.Equal(t, testUser.Username, claims.Username)
		assert.Equal(t, testUser.Role, claims.Role)

		// Verify auth method
		method, exists := GetAuthMethod(c)
		assert.True(t, exists)
		assert.Equal(t, "jwt", method)

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

func TestPATMiddleware_InvalidJWTToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	authService := NewService("test-secret", time.Hour, nil)
	mockPATService := &MockPATService{}

	router := gin.New()
	router.Use(PATMiddleware(authService, mockPATService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test with invalid JWT token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid token")
}

func TestGetAuthMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		contextValue   interface{}
		expectedMethod string
		expectedExists bool
	}{
		{
			name:           "PAT method",
			contextValue:   "pat",
			expectedMethod: "pat",
			expectedExists: true,
		},
		{
			name:           "JWT method",
			contextValue:   "jwt",
			expectedMethod: "jwt",
			expectedExists: true,
		},
		{
			name:           "No method set",
			contextValue:   nil,
			expectedMethod: "",
			expectedExists: false,
		},
		{
			name:           "Invalid type",
			contextValue:   123,
			expectedMethod: "",
			expectedExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())

			if tt.contextValue != nil {
				c.Set(AuthMethodContextKey, tt.contextValue)
			}

			method, exists := GetAuthMethod(c)
			assert.Equal(t, tt.expectedExists, exists)
			assert.Equal(t, tt.expectedMethod, method)
		})
	}
}

func TestIsPATAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		authMethod string
		expected   bool
	}{
		{
			name:       "PAT authenticated",
			authMethod: "pat",
			expected:   true,
		},
		{
			name:       "JWT authenticated",
			authMethod: "jwt",
			expected:   false,
		},
		{
			name:       "No auth method",
			authMethod: "",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())

			if tt.authMethod != "" {
				c.Set(AuthMethodContextKey, tt.authMethod)
			}

			result := IsPATAuthenticated(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsJWTAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		authMethod string
		expected   bool
	}{
		{
			name:       "JWT authenticated",
			authMethod: "jwt",
			expected:   true,
		},
		{
			name:       "PAT authenticated",
			authMethod: "pat",
			expected:   false,
		},
		{
			name:       "No auth method",
			authMethod: "",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())

			if tt.authMethod != "" {
				c.Set(AuthMethodContextKey, tt.authMethod)
			}

			result := IsJWTAuthenticated(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetUserFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     models.RoleUser,
	}

	t.Run("User from PAT context", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(UserContextKey, testUser)

		user, exists := GetUserFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.Username, user.Username)
		assert.Equal(t, testUser.Role, user.Role)
	})

	t.Run("User from JWT claims", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		claims := &Claims{
			UserID:   testUser.ID.String(),
			Username: testUser.Username,
			Role:     testUser.Role,
		}
		c.Set(ClaimsContextKey, claims)

		user, exists := GetUserFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.Username, user.Username)
		assert.Equal(t, testUser.Role, user.Role)
	})

	t.Run("No user context", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		user, exists := GetUserFromContext(c)
		assert.False(t, exists)
		assert.Nil(t, user)
	})
}
