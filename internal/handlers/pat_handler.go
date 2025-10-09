package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"product-requirements-management/internal/auth"
	"product-requirements-management/internal/service"
)

// PATHandler handles HTTP requests for Personal Access Token operations
type PATHandler struct {
	patService service.PATService
}

// NewPATHandler creates a new PAT handler instance
func NewPATHandler(patService service.PATService) *PATHandler {
	return &PATHandler{
		patService: patService,
	}
}

// CreatePAT handles POST /api/v1/pats
// @Summary Create a new Personal Access Token
// @Description Create a new Personal Access Token for the authenticated user. The token will be returned only once in the response and cannot be retrieved again. Requires User or Administrator role.
// @Tags personal-access-tokens
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param pat body service.CreatePATRequest true "PAT creation request"
// @Success 201 {object} service.PATCreateResponse "Successfully created PAT with token"
// @Failure 400 {object} map[string]interface{} "Invalid request body, duplicate token name, or invalid scopes"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 403 {object} map[string]interface{} "User or Administrator role required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/pats [post]
func (h *PATHandler) CreatePAT(c *gin.Context) {
	var req service.CreatePATRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	// Get current user ID from JWT token context
	userID, ok := auth.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "AUTHENTICATION_REQUIRED",
				"message": "User authentication required",
			},
		})
		return
	}

	// Parse user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Invalid user ID format",
			},
		})
		return
	}

	response, err := h.patService.CreatePAT(c.Request.Context(), userUUID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPATUserNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "User not found",
				},
			})
		case errors.Is(err, service.ErrPATDuplicateName):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": "Token name already exists",
				},
			})
		case errors.Is(err, service.ErrPATInvalidScopes):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": "Invalid scopes specified",
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to create PAT",
				},
			})
		}
		return
	}

	c.JSON(http.StatusCreated, response)
}

// ListPATs handles GET /api/v1/pats
// @Summary List Personal Access Tokens for the authenticated user
// @Description Retrieve a paginated list of Personal Access Tokens for the authenticated user. Token values are not included in the response for security. Requires User or Administrator role.
// @Tags personal-access-tokens
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query integer false "Maximum number of results to return" minimum(1) maximum(100) default(50) example(20)
// @Param offset query integer false "Number of results to skip for pagination" minimum(0) default(0) example(0)
// @Success 200 {object} service.ListResponse[models.PersonalAccessToken] "List of PATs with pagination info"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 403 {object} map[string]interface{} "User or Administrator role required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/pats [get]
func (h *PATHandler) ListPATs(c *gin.Context) {
	// Get current user ID from JWT token context
	userID, ok := auth.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "AUTHENTICATION_REQUIRED",
				"message": "User authentication required",
			},
		})
		return
	}

	// Parse user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Invalid user ID format",
			},
		})
		return
	}

	// Parse pagination parameters
	limit := 50 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	response, err := h.patService.ListUserPATs(c.Request.Context(), userUUID, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPATUserNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "User not found",
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to list PATs",
				},
			})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// RevokePAT handles DELETE /api/v1/pats/:id
// @Summary Revoke a Personal Access Token
// @Description Revoke (delete) a Personal Access Token by its ID. Only the owner of the token can revoke it. Once revoked, the token cannot be used for authentication. Requires User or Administrator role.
// @Tags personal-access-tokens
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "PAT UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Success 204 "PAT revoked successfully"
// @Failure 400 {object} map[string]interface{} "Invalid PAT ID format"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 403 {object} map[string]interface{} "User or Administrator role required, or unauthorized access to token"
// @Failure 404 {object} map[string]interface{} "PAT not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/pats/{id} [delete]
func (h *PATHandler) RevokePAT(c *gin.Context) {
	idParam := c.Param("id")

	// Parse PAT ID (UUID only)
	patID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid PAT ID format",
			},
		})
		return
	}

	// Get current user ID from JWT token context
	userID, ok := auth.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "AUTHENTICATION_REQUIRED",
				"message": "User authentication required",
			},
		})
		return
	}

	// Parse user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Invalid user ID format",
			},
		})
		return
	}

	err = h.patService.RevokePAT(c.Request.Context(), patID, userUUID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPATNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "PAT not found",
				},
			})
		case errors.Is(err, service.ErrPATUnauthorized):
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "INSUFFICIENT_PERMISSIONS",
					"message": "Unauthorized access to token",
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to revoke PAT",
				},
			})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
