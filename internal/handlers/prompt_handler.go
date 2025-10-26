package handlers

import (
	"net/http"
	"product-requirements-management/internal/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"product-requirements-management/internal/service"
)

// PromptHandler handles HTTP requests for prompt management
type PromptHandler struct {
	promptService *service.PromptService
	logger        *logrus.Logger
}

// NewPromptHandler creates a new PromptHandler instance
func NewPromptHandler(promptService *service.PromptService, logger *logrus.Logger) *PromptHandler {
	return &PromptHandler{
		promptService: promptService,
		logger:        logger,
	}
}

// CreatePrompt handles POST /api/v1/prompts
// @Summary Create a new prompt
// @Description Create a new system prompt (requires Administrator role)
// @Tags prompts
// @Accept json
// @Produce json
// @Param prompt body service.CreatePromptRequest true "Prompt creation request"
// @Success 201 {object} models.Prompt
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/prompts [post]
func (ph *PromptHandler) CreatePrompt(c *gin.Context) {
	var req service.CreatePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ph.logger.WithError(err).Error("Invalid request body for create prompt")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get user from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		ph.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	creatorID, ok := userID.(uuid.UUID)
	if !ok {
		ph.logger.Error("Invalid user ID type in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	prompt, err := ph.promptService.Create(c.Request.Context(), &req, creatorID)
	if err != nil {
		if err == service.ErrDuplicateEntry {
			c.JSON(http.StatusConflict, gin.H{"error": "Prompt with this name already exists"})
			return
		}
		ph.logger.WithError(err).Error("Failed to create prompt")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create prompt"})
		return
	}

	c.JSON(http.StatusCreated, prompt)
}

// ListPrompts handles GET /api/v1/prompts
// @Summary List prompts
// @Description List all prompts with pagination
// @Tags prompts
// @Produce json
// @Param limit query int false "Number of items per page (default: 50, max: 100)"
// @Param offset query int false "Number of items to skip (default: 0)"
// @Param creator_id query string false "Filter by creator UUID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/prompts [get]
func (ph *PromptHandler) ListPrompts(c *gin.Context) {
	// Parse pagination parameters
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Parse creator filter
	var creatorID *uuid.UUID
	if creatorIDStr := c.Query("creator_id"); creatorIDStr != "" {
		if parsed, err := uuid.Parse(creatorIDStr); err == nil {
			creatorID = &parsed
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid creator_id format"})
			return
		}
	}

	prompts, total, err := ph.promptService.List(c.Request.Context(), limit, offset, creatorID)
	if err != nil {
		ph.logger.WithError(err).Error("Failed to list prompts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list prompts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        prompts,
		"total_count": total,
		"limit":       limit,
		"offset":      offset,
	})
}

// GetPrompt handles GET /api/v1/prompts/:id
// @Summary Get prompt by ID
// @Description Get a prompt by UUID or reference ID
// @Tags prompts
// @Produce json
// @Param id path string true "Prompt UUID or reference ID (e.g., PROMPT-001)"
// @Success 200 {object} models.Prompt
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/prompts/{id} [get]
func (ph *PromptHandler) GetPrompt(c *gin.Context) {
	idParam := c.Param("id")

	var prompt *models.Prompt
	var err error

	// Try to parse as UUID first
	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		prompt, err = ph.promptService.GetByID(c.Request.Context(), id)
	} else {
		// Try as reference ID
		prompt, err = ph.promptService.GetByReferenceID(c.Request.Context(), idParam)
	}

	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Prompt not found"})
			return
		}
		ph.logger.WithError(err).WithField("id", idParam).Error("Failed to get prompt")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get prompt"})
		return
	}

	c.JSON(http.StatusOK, prompt)
}

// UpdatePrompt handles PUT /api/v1/prompts/:id
// @Summary Update prompt
// @Description Update an existing prompt (requires Administrator role)
// @Tags prompts
// @Accept json
// @Produce json
// @Param id path string true "Prompt UUID or reference ID (e.g., PROMPT-001)"
// @Param prompt body service.UpdatePromptRequest true "Prompt update request"
// @Success 200 {object} models.Prompt
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/prompts/{id} [put]
func (ph *PromptHandler) UpdatePrompt(c *gin.Context) {
	idParam := c.Param("id")

	var req service.UpdatePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ph.logger.WithError(err).Error("Invalid request body for update prompt")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Parse ID (UUID or reference ID)
	var promptID uuid.UUID
	var err error

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		promptID = id
	} else {
		// Get prompt by reference ID to get UUID
		prompt, getErr := ph.promptService.GetByReferenceID(c.Request.Context(), idParam)
		if getErr != nil {
			if getErr == service.ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Prompt not found"})
				return
			}
			ph.logger.WithError(getErr).WithField("reference_id", idParam).Error("Failed to get prompt by reference ID")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get prompt"})
			return
		}
		promptID = prompt.ID
	}

	prompt, err := ph.promptService.Update(c.Request.Context(), promptID, &req)
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Prompt not found"})
			return
		}
		ph.logger.WithError(err).WithField("id", idParam).Error("Failed to update prompt")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update prompt"})
		return
	}

	c.JSON(http.StatusOK, prompt)
}

// DeletePrompt handles DELETE /api/v1/prompts/:id
// @Summary Delete prompt
// @Description Delete a prompt (requires Administrator role)
// @Tags prompts
// @Produce json
// @Param id path string true "Prompt UUID or reference ID (e.g., PROMPT-001)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/prompts/{id} [delete]
func (ph *PromptHandler) DeletePrompt(c *gin.Context) {
	idParam := c.Param("id")

	// Parse ID (UUID or reference ID)
	var promptID uuid.UUID
	var err error

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		promptID = id
	} else {
		// Get prompt by reference ID to get UUID
		prompt, getErr := ph.promptService.GetByReferenceID(c.Request.Context(), idParam)
		if getErr != nil {
			if getErr == service.ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Prompt not found"})
				return
			}
			ph.logger.WithError(getErr).WithField("reference_id", idParam).Error("Failed to get prompt by reference ID")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get prompt"})
			return
		}
		promptID = prompt.ID
	}

	err = ph.promptService.Delete(c.Request.Context(), promptID)
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Prompt not found"})
			return
		}
		ph.logger.WithError(err).WithField("id", idParam).Error("Failed to delete prompt")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete prompt"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ActivatePrompt handles PATCH /api/v1/prompts/:id/activate
// @Summary Activate prompt
// @Description Activate a prompt and deactivate all others (requires Administrator role)
// @Tags prompts
// @Produce json
// @Param id path string true "Prompt UUID or reference ID (e.g., PROMPT-001)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/prompts/{id}/activate [patch]
func (ph *PromptHandler) ActivatePrompt(c *gin.Context) {
	idParam := c.Param("id")

	// Parse ID (UUID or reference ID)
	var promptID uuid.UUID
	var err error

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		promptID = id
	} else {
		// Get prompt by reference ID to get UUID
		prompt, getErr := ph.promptService.GetByReferenceID(c.Request.Context(), idParam)
		if getErr != nil {
			if getErr == service.ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Prompt not found"})
				return
			}
			ph.logger.WithError(getErr).WithField("reference_id", idParam).Error("Failed to get prompt by reference ID")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get prompt"})
			return
		}
		promptID = prompt.ID
	}

	err = ph.promptService.Activate(c.Request.Context(), promptID)
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Prompt not found"})
			return
		}
		ph.logger.WithError(err).WithField("id", idParam).Error("Failed to activate prompt")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate prompt"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Prompt activated successfully"})
}

// GetActivePrompt handles GET /api/v1/prompts/active
// @Summary Get active prompt
// @Description Get the currently active prompt
// @Tags prompts
// @Produce json
// @Success 200 {object} models.Prompt
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/prompts/active [get]
func (ph *PromptHandler) GetActivePrompt(c *gin.Context) {
	prompt, err := ph.promptService.GetActive(c.Request.Context())
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "No active prompt found"})
			return
		}
		ph.logger.WithError(err).Error("Failed to get active prompt")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active prompt"})
		return
	}

	c.JSON(http.StatusOK, prompt)
}
