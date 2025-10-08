package auth

import (
	"net/http"
	"strings"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	PATPrefix            = "mcp_pat_"
	AuthMethodContextKey = "auth_method"
	UserIDContextKey     = "user_id"
)

// PATMiddleware creates authentication middleware that supports both PAT and JWT tokens
func PATMiddleware(authService *Service, patService service.PATService) gin.HandlerFunc {
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

		// Try PAT authentication first if token has PAT prefix
		if strings.HasPrefix(tokenString, PATPrefix) {
			if err := authenticateWithPAT(c, patService, tokenString); err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// Fall back to JWT authentication
		if err := authenticateWithJWT(c, authService, tokenString); err != nil {
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

		c.Next()
	}
}

// authenticateWithPAT handles PAT token authentication
func authenticateWithPAT(c *gin.Context, patService service.PATService, token string) error {
	// Validate PAT token and get associated user
	user, err := patService.ValidateToken(c.Request.Context(), token)
	if err != nil {
		return err
	}

	// Create claims-like structure for PAT authentication
	patClaims := &Claims{
		UserID:   user.ID.String(),
		Username: user.Username,
		Role:     user.Role,
	}

	// Set context values for compatibility with existing handlers
	c.Set(ClaimsContextKey, patClaims)
	c.Set(UserContextKey, user)
	c.Set(UserIDContextKey, user.ID.String())
	c.Set(AuthMethodContextKey, "pat")

	return nil
}

// authenticateWithJWT handles JWT token authentication
func authenticateWithJWT(c *gin.Context, authService *Service, tokenString string) error {
	claims, err := authService.ValidateToken(tokenString)
	if err != nil {
		return err
	}

	// Set context values (existing JWT behavior)
	c.Set(ClaimsContextKey, claims)
	c.Set(UserIDContextKey, claims.UserID)
	c.Set(AuthMethodContextKey, "jwt")

	return nil
}

// GetAuthMethod extracts the authentication method from the Gin context
func GetAuthMethod(c *gin.Context) (string, bool) {
	method, exists := c.Get(AuthMethodContextKey)
	if !exists {
		return "", false
	}

	authMethod, ok := method.(string)
	return authMethod, ok
}

// IsPATAuthenticated checks if the current request was authenticated using a PAT
func IsPATAuthenticated(c *gin.Context) bool {
	method, ok := GetAuthMethod(c)
	return ok && method == "pat"
}

// IsJWTAuthenticated checks if the current request was authenticated using JWT
func IsJWTAuthenticated(c *gin.Context) bool {
	method, ok := GetAuthMethod(c)
	return ok && method == "jwt"
}

// GetUserFromContext extracts the user from context (works for both PAT and JWT auth)
func GetUserFromContext(c *gin.Context) (*models.User, bool) {
	// Try to get user directly (set by PAT auth)
	if user, exists := c.Get(UserContextKey); exists {
		if u, ok := user.(*models.User); ok {
			return u, true
		}
	}

	// Fall back to claims-based approach (JWT auth)
	claims, ok := GetCurrentUser(c)
	if !ok {
		return nil, false
	}

	// Create a user object from claims for compatibility
	user := &models.User{
		Username: claims.Username,
		Role:     claims.Role,
	}

	// Parse UUID from string
	if userID, err := parseUUID(claims.UserID); err == nil {
		user.ID = userID
	}

	return user, true
}

// parseUUID is a helper function to parse UUID from string
func parseUUID(s string) (uuid.UUID, error) {
	// Import uuid package at the top if not already imported
	return uuid.Parse(s)
}
