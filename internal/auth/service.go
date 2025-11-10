package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrInsufficientRole   = errors.New("insufficient role permissions")
)

// Claims represents the JWT claims
type Claims struct {
	UserID   string          `json:"user_id"`
	Username string          `json:"username"`
	Role     models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// Service handles authentication operations
type Service struct {
	jwtSecret          []byte
	tokenDuration      time.Duration
	refreshTokenRepo   repository.RefreshTokenRepository
	refreshTokenExpiry time.Duration
}

// NewService creates a new authentication service
func NewService(jwtSecret string, tokenDuration time.Duration, refreshTokenRepo repository.RefreshTokenRepository) *Service {
	return &Service{
		jwtSecret:          []byte(jwtSecret),
		tokenDuration:      tokenDuration,
		refreshTokenRepo:   refreshTokenRepo,
		refreshTokenExpiry: 30 * 24 * time.Hour, // 30 days
	}
}

// HashPassword hashes a password using bcrypt
func (s *Service) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against its hash
func (s *Service) VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// GenerateToken generates a JWT token for a user
func (s *Service) GenerateToken(user *models.User) (string, error) {
	claims := Claims{
		UserID:   user.ID.String(),
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateToken validates a JWT token and returns the claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// CheckPermission checks if a user has the required permission level
func (s *Service) CheckPermission(userRole models.UserRole, requiredRole models.UserRole) error {
	// Define role hierarchy: Administrator > User > Commenter
	roleHierarchy := map[models.UserRole]int{
		models.RoleAdministrator: 3,
		models.RoleUser:          2,
		models.RoleCommenter:     1,
	}

	userLevel, userExists := roleHierarchy[userRole]
	requiredLevel, requiredExists := roleHierarchy[requiredRole]

	if !userExists || !requiredExists {
		return ErrInsufficientRole
	}

	if userLevel < requiredLevel {
		return ErrInsufficientRole
	}

	return nil
}

// CanEdit checks if a user can edit entities
func (s *Service) CanEdit(userRole models.UserRole) bool {
	return userRole == models.RoleAdministrator || userRole == models.RoleUser
}

// CanDelete checks if a user can delete entities
func (s *Service) CanDelete(userRole models.UserRole) bool {
	return userRole == models.RoleAdministrator || userRole == models.RoleUser
}

// CanManageUsers checks if a user can manage other users
func (s *Service) CanManageUsers(userRole models.UserRole) bool {
	return userRole == models.RoleAdministrator
}

// CanManageConfig checks if a user can manage system configuration
func (s *Service) CanManageConfig(userRole models.UserRole) bool {
	return userRole == models.RoleAdministrator
}

// GenerateRefreshToken creates a new refresh token for a user
func (s *Service) GenerateRefreshToken(ctx context.Context, user *models.User) (string, error) {
	// Generate secure random token (32 bytes = 256 bits)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Encode to base64 URL-safe string
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Hash the token for storage
	tokenHash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash token: %w", err)
	}

	// Create refresh token record
	refreshToken := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: string(tokenHash),
		ExpiresAt: time.Now().Add(s.refreshTokenExpiry),
	}

	if err := s.refreshTokenRepo.Create(refreshToken); err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return token, nil
}

// ValidateRefreshToken validates and returns user for refresh token
func (s *Service) ValidateRefreshToken(ctx context.Context, token string) (*models.User, string, error) {
	// Find all refresh tokens (we need to check hashes)
	// In production, consider adding a token prefix/identifier to optimize this
	allTokens, err := s.refreshTokenRepo.FindAll()
	if err != nil {
		return nil, "", ErrInvalidToken
	}

	// Find matching token by comparing hashes
	var matchedToken *models.RefreshToken
	for _, rt := range allTokens {
		if err := bcrypt.CompareHashAndPassword([]byte(rt.TokenHash), []byte(token)); err == nil {
			matchedToken = rt
			break
		}
	}

	if matchedToken == nil {
		return nil, "", ErrInvalidToken
	}

	// Check expiration
	if matchedToken.IsExpired() {
		// Clean up expired token
		s.refreshTokenRepo.Delete(matchedToken.ID)
		return nil, "", ErrTokenExpired
	}

	// Update last used timestamp
	now := time.Now()
	matchedToken.LastUsedAt = &now
	s.refreshTokenRepo.Update(matchedToken)

	// Get user from database
	var user models.User
	if err := s.refreshTokenRepo.GetDB().First(&user, "id = ?", matchedToken.UserID).Error; err != nil {
		return nil, "", ErrInvalidToken
	}

	// Generate new refresh token (token rotation)
	newRefreshToken, err := s.GenerateRefreshToken(ctx, &user)
	if err != nil {
		return nil, "", err
	}

	// Revoke old token
	s.refreshTokenRepo.Delete(matchedToken.ID)

	return &user, newRefreshToken, nil
}

// RevokeRefreshToken invalidates a refresh token
func (s *Service) RevokeRefreshToken(ctx context.Context, token string) error {
	// Find and delete the token by comparing hashes
	allTokens, err := s.refreshTokenRepo.FindAll()
	if err != nil {
		return ErrInvalidToken
	}

	for _, rt := range allTokens {
		if err := bcrypt.CompareHashAndPassword([]byte(rt.TokenHash), []byte(token)); err == nil {
			return s.refreshTokenRepo.Delete(rt.ID)
		}
	}

	return ErrInvalidToken
}

// CleanupExpiredTokens removes expired refresh tokens
func (s *Service) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	return s.refreshTokenRepo.DeleteExpired()
}
