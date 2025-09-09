package helpers

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthHelper provides authentication utilities for benchmark testing
type AuthHelper struct {
	JWTSecret string
}

// NewAuthHelper creates a new authentication helper
func NewAuthHelper(jwtSecret string) *AuthHelper {
	return &AuthHelper{
		JWTSecret: jwtSecret,
	}
}

// GenerateTestToken creates a JWT token for benchmark testing
func (ah *AuthHelper) GenerateTestToken(userID string, username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24 hour expiration
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(ah.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	return tokenString, nil
}

// AuthenticateClient sets up authentication for a benchmark client
func (ah *AuthHelper) AuthenticateClient(client *BenchmarkClient, userID string, username string) error {
	token, err := ah.GenerateTestToken(userID, username)
	if err != nil {
		return fmt.Errorf("failed to generate test token: %w", err)
	}

	client.SetAuthToken(token)
	return nil
}

// CreateTestUser represents a test user for authentication
type TestUser struct {
	ID       string
	Username string
	Email    string
	FullName string
}

// GetDefaultTestUser returns a default test user for benchmarks
func GetDefaultTestUser() TestUser {
	return TestUser{
		ID:       "benchmark-user-1",
		Username: "benchmark_user",
		Email:    "benchmark@example.com",
		FullName: "Benchmark User",
	}
}