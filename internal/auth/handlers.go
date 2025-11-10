package auth

import (
	"net/http"
	"time"

	"product-requirements-management/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LoginRequest represents a login request
// @Description Request payload for user authentication
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"john_doe"`    // Username for authentication
	Password string `json:"password" binding:"required" example:"password123"` // Password for authentication
}

// LoginResponse represents a login response
// @Description Response payload for successful authentication containing JWT token, refresh token, and user information
type LoginResponse struct {
	Token        string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`  // JWT authentication token
	RefreshToken string       `json:"refresh_token" example:"dGhpc19pc19hX3JlZnJlc2hfdG9rZW4="` // Refresh token for obtaining new access tokens
	User         UserResponse `json:"user"`                                                     // Authenticated user information
	ExpiresAt    time.Time    `json:"expires_at" example:"2023-01-02T12:30:00Z"`                // Token expiration timestamp
}

// RefreshRequest represents a token refresh request
// @Description Request payload for refreshing an expired access token using a refresh token
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"dGhpc19pc19hX3JlZnJlc2hfdG9rZW4="` // Refresh token obtained from login
}

// RefreshResponse represents a token refresh response
// @Description Response payload for successful token refresh containing new access token and refresh token
type RefreshResponse struct {
	Token        string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // New JWT authentication token
	RefreshToken string    `json:"refresh_token" example:"bmV3X3JlZnJlc2hfdG9rZW4="`        // New refresh token (token rotation)
	ExpiresAt    time.Time `json:"expires_at" example:"2023-01-02T12:30:00Z"`               // New token expiration timestamp
}

// LogoutRequest represents a logout request
// @Description Request payload for logging out and invalidating a refresh token
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"dGhpc19pc19hX3JlZnJlc2hfdG9rZW4="` // Refresh token to invalidate
}

// ErrorResponse represents an error response
// @Description Standard error response format used across authentication endpoints
type ErrorResponse struct {
	Error ErrorDetail `json:"error"` // Error details including code and message
}

// ErrorDetail represents error details
// @Description Detailed error information with code and message
type ErrorDetail struct {
	Code    string `json:"code" example:"INVALID_REFRESH_TOKEN"`    // Error code for programmatic handling
	Message string `json:"message" example:"Invalid refresh token"` // Human-readable error message
}

// UserResponse represents a user in API responses
// @Description User information returned in API responses (password hash excluded for security)
type UserResponse struct {
	ID        string          `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"` // Unique user identifier
	Username  string          `json:"username" example:"john_doe"`                       // Unique username
	Email     string          `json:"email" example:"john.doe@example.com"`              // User email address
	Role      models.UserRole `json:"role" example:"User"`                               // User role determining permissions
	CreatedAt time.Time       `json:"created_at" example:"2023-01-01T00:00:00Z"`         // Account creation timestamp
	UpdatedAt time.Time       `json:"updated_at" example:"2023-01-02T12:30:00Z"`         // Last account update timestamp
}

// CreateUserRequest represents a request to create a new user
// @Description Request payload for creating a new user account (Administrator role required)
type CreateUserRequest struct {
	Username string          `json:"username" binding:"required" example:"jane_doe"`            // Unique username (required)
	Email    string          `json:"email" binding:"required,email" example:"jane@example.com"` // Valid email address (required)
	Password string          `json:"password" binding:"required,min=8" example:"securepass123"` // Password (minimum 8 characters, required)
	Role     models.UserRole `json:"role" binding:"required" example:"User"`                    // User role: Administrator, User, or Commenter (required)
}

// UpdateUserRequest represents a request to update a user
// @Description Request payload for updating user information (Administrator role required)
type UpdateUserRequest struct {
	Username string          `json:"username" example:"jane_smith"`                                    // New username (optional)
	Email    string          `json:"email" binding:"omitempty,email" example:"jane.smith@example.com"` // New email address (optional, must be valid if provided)
	Role     models.UserRole `json:"role" example:"Administrator"`                                     // New user role (optional)
}

// ChangePasswordRequest represents a request to change password
// @Description Request payload for changing user password (authentication required)
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required" example:"oldpassword123"`   // Current password for verification (required)
	NewPassword     string `json:"new_password" binding:"required,min=8" example:"newpassword456"` // New password (minimum 8 characters, required)
}

// Handlers contains authentication handlers
type Handlers struct {
	service *Service
	db      *gorm.DB
}

// NewHandlers creates new authentication handlers
func NewHandlers(service *Service, db *gorm.DB) *Handlers {
	return &Handlers{
		service: service,
		db:      db,
	}
}

// Login handles user login
// @Summary User login
// @Description Authenticate user with username and password to receive JWT token and refresh token
// @Tags authentication
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse "Successful authentication with JWT token and refresh token"
// @Failure 400 {object} map[string]string "Invalid request format"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/login [post]
func (h *Handlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if err := h.service.VerifyPassword(req.Password, user.PasswordHash); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := h.service.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Generate refresh token
	refreshToken, err := h.service.GenerateRefreshToken(c.Request.Context(), &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	response := LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User: UserResponse{
			ID:        user.ID.String(),
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		ExpiresAt: time.Now().Add(h.service.tokenDuration),
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Refresh an expired access token using a refresh token
// @Tags authentication
// @Accept json
// @Produce json
// @Param refresh body RefreshRequest true "Refresh token request"
// @Success 200 {object} RefreshResponse "Successfully refreshed tokens"
// @Failure 400 {object} ErrorResponse "Invalid request format"
// @Failure 401 {object} ErrorResponse "Invalid or expired refresh token"
// @Failure 429 {object} ErrorResponse "Too many refresh attempts"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (h *Handlers) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// Validate refresh token and get user
	user, newRefreshToken, err := h.service.ValidateRefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if err == ErrTokenExpired {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: ErrorDetail{
					Code:    "REFRESH_TOKEN_EXPIRED",
					Message: "Refresh token has expired",
				},
			})
			return
		}
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REFRESH_TOKEN",
				Message: "Invalid or revoked refresh token",
			},
		})
		return
	}

	// Generate new access token
	accessToken, err := h.service.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to generate access token",
			},
		})
		return
	}

	response := RefreshResponse{
		Token:        accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(h.service.tokenDuration),
	}

	c.JSON(http.StatusOK, response)
}

// Logout handles user logout
// @Summary User logout
// @Description Logout user and invalidate refresh token
// @Tags authentication
// @Accept json
// @Produce json
// @Param logout body LogoutRequest true "Logout request"
// @Success 204 "Successfully logged out"
// @Failure 400 {object} ErrorResponse "Invalid request format"
// @Failure 401 {object} ErrorResponse "Invalid refresh token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/logout [post]
func (h *Handlers) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// Revoke refresh token
	if err := h.service.RevokeRefreshToken(c.Request.Context(), req.RefreshToken); err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REFRESH_TOKEN",
				Message: "Session already logged out",
			},
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateUser handles user creation (admin only)
// @Summary Create new user
// @Description Create a new user account (Administrator role required)
// @Tags authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body CreateUserRequest true "User creation request"
// @Success 201 {object} UserResponse "Successfully created user"
// @Failure 400 {object} map[string]string "Invalid request format or role"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 403 {object} map[string]string "Administrator role required"
// @Failure 409 {object} map[string]string "Username or email already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/users [post]
func (h *Handlers) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate role
	if req.Role != models.RoleAdministrator && req.Role != models.RoleUser && req.Role != models.RoleCommenter {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	// Hash password
	passwordHash, err := h.service.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         req.Role,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
		return
	}

	response := UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetUsers handles listing users (admin only)
// @Summary List all users
// @Description Get list of all users (Administrator role required)
// @Tags authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {array} UserResponse "List of users"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 403 {object} map[string]string "Administrator role required"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/users [get]
func (h *Handlers) GetUsers(c *gin.Context) {
	var users []models.User
	if err := h.db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var response []UserResponse
	for _, user := range users {
		response = append(response, UserResponse{
			ID:        user.ID.String(),
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetUser handles getting a specific user (admin only)
// @Summary Get user by ID
// @Description Get specific user details (Administrator role required)
// @Tags authentication
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} UserResponse "User details"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 403 {object} map[string]string "Administrator role required"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/users/{id} [get]
func (h *Handlers) GetUser(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	response := UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateUser handles updating a user (admin only)
// @Summary Update user
// @Description Update user details (Administrator role required)
// @Tags authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param user body UpdateUserRequest true "User update request"
// @Success 200 {object} UserResponse "Updated user details"
// @Failure 400 {object} map[string]string "Invalid request format or role"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 403 {object} map[string]string "Administrator role required"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 409 {object} map[string]string "Username or email already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/users/{id} [put]
func (h *Handlers) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Update fields if provided
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Role != "" {
		// Validate role
		if req.Role != models.RoleAdministrator && req.Role != models.RoleUser && req.Role != models.RoleCommenter {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
			return
		}
		user.Role = req.Role
	}

	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
		return
	}

	response := UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteUser handles deleting a user (admin only)
// @Summary Delete user
// @Description Delete user account (Administrator role required)
// @Tags authentication
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 204 "User successfully deleted"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 403 {object} map[string]string "Administrator role required"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 409 {object} map[string]string "Cannot delete user with associated entities"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/users/{id} [delete]
func (h *Handlers) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if user has created any entities (prevent deletion if they have)
	var count int64
	h.db.Model(&models.Epic{}).Where("creator_id = ? OR assignee_id = ?", userID, userID).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete user with associated entities"})
		return
	}

	if err := h.db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetProfile handles getting current user profile
// @Summary Get current user profile
// @Description Get authenticated user's profile information
// @Tags authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserResponse "Current user profile"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/profile [get]
func (h *Handlers) GetProfile(c *gin.Context) {
	claims, exists := GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var user models.User
	if err := h.db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	response := UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// ChangePassword handles changing current user's password
// @Summary Change password
// @Description Change authenticated user's password
// @Tags authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param password body ChangePasswordRequest true "Password change request"
// @Success 200 {object} map[string]string "Password changed successfully"
// @Failure 400 {object} map[string]string "Invalid request format"
// @Failure 401 {object} map[string]string "Authentication required or invalid current password"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/change-password [post]
func (h *Handlers) ChangePassword(c *gin.Context) {
	claims, exists := GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Verify current password
	if err := h.service.VerifyPassword(req.CurrentPassword, user.PasswordHash); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid current password"})
		return
	}

	// Hash new password
	newPasswordHash, err := h.service.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user.PasswordHash = newPasswordHash
	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
