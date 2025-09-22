package integration

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"product-requirements-management/internal/auth"
	"product-requirements-management/internal/models"
)

// TestAuthContext holds authentication context for tests
type TestAuthContext struct {
	AuthService *auth.Service
	TestUser    *models.User
	AdminUser   *models.User
	Token       string
	AdminToken  string
}

// createTestUser creates a test user for integration tests
func createTestUser(t *testing.T, db *gorm.DB) *models.User {
	user := &models.User{
		ID:       uuid.New(),
		Username: fmt.Sprintf("testuser_%s", uuid.New().String()[:8]),
		Email:    fmt.Sprintf("test_%s@example.com", uuid.New().String()[:8]),
		Role:     models.RoleUser,
	}

	err := db.Create(user).Error
	require.NoError(t, err)

	return user
}

// SetupTestAuth creates authentication context for tests
func SetupTestAuth(t *testing.T, db *gorm.DB) *TestAuthContext {
	// Create auth service with test JWT secret
	jwtSecret := "test-jwt-secret-key-for-integration-tests"
	authService := auth.NewService(jwtSecret, 24*time.Hour)

	// Create test user
	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	// Hash password
	hashedPassword, err := authService.HashPassword("testpassword")
	require.NoError(t, err)
	testUser.PasswordHash = hashedPassword

	err = db.Create(testUser).Error
	require.NoError(t, err)

	// Create admin user
	adminUser := &models.User{
		ID:       uuid.New(),
		Username: "adminuser",
		Email:    "admin@example.com",
		Role:     models.RoleAdministrator,
	}

	// Hash admin password
	hashedAdminPassword, err := authService.HashPassword("adminpassword")
	require.NoError(t, err)
	adminUser.PasswordHash = hashedAdminPassword

	err = db.Create(adminUser).Error
	require.NoError(t, err)

	// Generate tokens
	token, err := authService.GenerateToken(testUser)
	require.NoError(t, err)

	adminToken, err := authService.GenerateToken(adminUser)
	require.NoError(t, err)

	return &TestAuthContext{
		AuthService: authService,
		TestUser:    testUser,
		AdminUser:   adminUser,
		Token:       token,
		AdminToken:  adminToken,
	}
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// intPtr returns a pointer to an int
func intPtr(i int) *int {
	return &i
}

// skipIfShort skips the test if running in short mode
func skipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}

// skipIfNoDocker skips the test if Docker is not available
func skipIfNoDocker(t *testing.T) {
	if os.Getenv("SKIP_DOCKER_TESTS") == "true" {
		t.Skip("Skipping Docker-based test")
	}
}

// logTestStart logs the start of a test
func logTestStart(t *testing.T, testName string) {
	log.Printf("🧪 Starting integration test: %s", testName)
}

// logTestEnd logs the end of a test
func logTestEnd(t *testing.T, testName string) {
	log.Printf("✅ Completed integration test: %s", testName)
}

// makeAuthenticatedRequest creates an HTTP request with authentication token
func makeAuthenticatedRequest(method, url, token string) (*http.Request, *httptest.ResponseRecorder) {
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.ResponseRecorder{}
	return req, &w
}
