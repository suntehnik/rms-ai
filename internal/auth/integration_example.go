package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"product-requirements-management/internal/config"
)

// IntegrateWithServer demonstrates how to integrate the authentication system
// with the main server. This is an example function that shows the integration pattern.
func IntegrateWithServer(router *gin.Engine, cfg *config.Config, db *gorm.DB) {
	// Initialize authentication service
	// In a real implementation, JWT secret should come from config or environment
	jwtSecret := "your-jwt-secret-here" // Should come from config in real implementation
	tokenDuration := 24 * time.Hour     // Should come from config in real implementation
	
	authService := NewService(jwtSecret, tokenDuration)
	authHandlers := NewHandlers(authService, db)

	// Authentication routes (public)
	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/login", authHandlers.Login)
	}

	// User profile routes (authenticated users only)
	profile := router.Group("/api/v1/profile")
	profile.Use(authService.Middleware())
	{
		profile.GET("", authHandlers.GetProfile)
		profile.PUT("/change-password", authHandlers.ChangePassword)
	}

	// User management routes (admin only)
	users := router.Group("/api/v1/users")
	users.Use(authService.Middleware())
	users.Use(authService.RequireAdministrator())
	{
		users.POST("", authHandlers.CreateUser)
		users.GET("", authHandlers.GetUsers)
		users.GET("/:id", authHandlers.GetUser)
		users.PUT("/:id", authHandlers.UpdateUser)
		users.DELETE("/:id", authHandlers.DeleteUser)
	}

	// Protected API routes examples
	api := router.Group("/api/v1")
	api.Use(authService.Middleware()) // All API routes require authentication
	{
		// Epic routes (User role or higher required)
		epics := api.Group("/epics")
		epics.Use(authService.RequireUser()) // Only Users and Administrators can manage epics
		{
			epics.POST("", func(c *gin.Context) {
				// Epic creation handler - to be implemented in task 6
				c.JSON(200, gin.H{"message": "Create epic - to be implemented"})
			})
			epics.PUT("/:id", func(c *gin.Context) {
				// Epic update handler - to be implemented in task 6
				c.JSON(200, gin.H{"message": "Update epic - to be implemented"})
			})
			epics.DELETE("/:id", func(c *gin.Context) {
				// Epic deletion handler - to be implemented in task 6
				c.JSON(200, gin.H{"message": "Delete epic - to be implemented"})
			})
		}

		// Read-only routes (any authenticated user can access)
		api.GET("/epics", authService.RequireCommenter(), func(c *gin.Context) {
			// List epics handler - to be implemented in task 6
			c.JSON(200, gin.H{"message": "List epics - to be implemented"})
		})
		api.GET("/epics/:id", authService.RequireCommenter(), func(c *gin.Context) {
			// Get epic handler - to be implemented in task 6
			c.JSON(200, gin.H{"message": "Get epic - to be implemented"})
		})

		// Comment routes (any authenticated user can comment)
		comments := api.Group("/comments")
		comments.Use(authService.RequireCommenter())
		{
			comments.POST("", func(c *gin.Context) {
				// Create comment handler - to be implemented in task 13
				c.JSON(200, gin.H{"message": "Create comment - to be implemented"})
			})
		}

		// Configuration routes (admin only)
		config := api.Group("/config")
		config.Use(authService.RequireAdministrator())
		{
			config.GET("/requirement-types", func(c *gin.Context) {
				// Get requirement types handler - to be implemented in task 11
				c.JSON(200, gin.H{"message": "Get requirement types - to be implemented"})
			})
			config.POST("/requirement-types", func(c *gin.Context) {
				// Create requirement type handler - to be implemented in task 11
				c.JSON(200, gin.H{"message": "Create requirement type - to be implemented"})
			})
		}
	}
}

// Example of how to use authentication in handlers
func ExampleHandlerWithAuth(c *gin.Context) {
	// Get current user information
	claims, exists := GetCurrentUser(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Authentication required"})
		return
	}

	// Get specific user information
	userID, _ := GetCurrentUserID(c)
	userRole, _ := GetCurrentUserRole(c)

	// Use the information in your business logic
	c.JSON(200, gin.H{
		"message":  "Success",
		"user_id":  userID,
		"username": claims.Username,
		"role":     userRole,
	})
}

// Example configuration structure that should be added to config package
type AuthConfig struct {
	JWTSecret          string `mapstructure:"jwt_secret"`
	TokenDurationHours int    `mapstructure:"token_duration_hours"`
}

// Example of how the main config structure should be extended
type ConfigWithAuth struct {
	// ... existing config fields
	Auth AuthConfig `mapstructure:"auth"`
}