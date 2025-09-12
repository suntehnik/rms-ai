package auth

import (
	"errors"
	"time"

	"product-requirements-management/internal/models"

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
	jwtSecret     []byte
	tokenDuration time.Duration
}

// NewService creates a new authentication service
func NewService(jwtSecret string, tokenDuration time.Duration) *Service {
	return &Service{
		jwtSecret:     []byte(jwtSecret),
		tokenDuration: tokenDuration,
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
