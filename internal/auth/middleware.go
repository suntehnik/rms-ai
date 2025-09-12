package auth

import (
	"net/http"
	"strings"

	"product-requirements-management/internal/models"

	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserContextKey      = "user"
	ClaimsContextKey    = "claims"
)

// Middleware creates authentication middleware
func (s *Service) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		claims, err := s.ValidateToken(tokenString)
		if err != nil {
			switch err {
			case ErrTokenExpired:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			case ErrInvalidToken:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			default:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
			}
			c.Abort()
			return
		}

		// Store claims in context for use in handlers
		c.Set(ClaimsContextKey, claims)
		c.Next()
	}
}

// RequireRole creates middleware that requires a specific role or higher
func (s *Service) RequireRole(requiredRole models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get(ClaimsContextKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*Claims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid claims"})
			c.Abort()
			return
		}

		if err := s.CheckPermission(userClaims.Role, requiredRole); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdministrator creates middleware that requires administrator role
func (s *Service) RequireAdministrator() gin.HandlerFunc {
	return s.RequireRole(models.RoleAdministrator)
}

// RequireUser creates middleware that requires user role or higher
func (s *Service) RequireUser() gin.HandlerFunc {
	return s.RequireRole(models.RoleUser)
}

// RequireCommenter creates middleware that requires commenter role or higher (any authenticated user)
func (s *Service) RequireCommenter() gin.HandlerFunc {
	return s.RequireRole(models.RoleCommenter)
}

// GetCurrentUser extracts user claims from the Gin context
func GetCurrentUser(c *gin.Context) (*Claims, bool) {
	claims, exists := c.Get(ClaimsContextKey)
	if !exists {
		return nil, false
	}

	userClaims, ok := claims.(*Claims)
	return userClaims, ok
}

// GetCurrentUserID extracts user ID from the Gin context
func GetCurrentUserID(c *gin.Context) (string, bool) {
	claims, ok := GetCurrentUser(c)
	if !ok {
		return "", false
	}
	return claims.UserID, true
}

// GetCurrentUserRole extracts user role from the Gin context
func GetCurrentUserRole(c *gin.Context) (models.UserRole, bool) {
	claims, ok := GetCurrentUser(c)
	if !ok {
		return "", false
	}
	return claims.Role, true
}
