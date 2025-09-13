package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// AcceptanceCriteriaHandler handles HTTP requests for acceptance criteria operations
type AcceptanceCriteriaHandler struct {
	acceptanceCriteriaService service.AcceptanceCriteriaService
}

// NewAcceptanceCriteriaHandler creates a new acceptance criteria handler instance
func NewAcceptanceCriteriaHandler(acceptanceCriteriaService service.AcceptanceCriteriaService) *AcceptanceCriteriaHandler {
	return &AcceptanceCriteriaHandler{
		acceptanceCriteriaService: acceptanceCriteriaService,
	}
}

// CreateAcceptanceCriteria handles POST /api/v1/user-stories/:id/acceptance-criteria
// @Summary Create acceptance criteria within a user story
// @Description Create new acceptance criteria that belongs to the specified user story. This is a nested resource creation that establishes the parent-child relationship between user story and acceptance criteria. The user story ID from the URL path will be used as the parent.
// @Tags user-stories
// @Accept json
// @Produce json
// @Param id path string true "User story UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param acceptance_criteria body service.CreateAcceptanceCriteriaRequest true "Acceptance criteria creation request (user_story_id will be set from path parameter)"
// @Success 201 {object} models.AcceptanceCriteria "Successfully created acceptance criteria within user story"
// @Failure 400 {object} map[string]interface{} "Invalid user story ID format, request body, user story not found, or author not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user-stories/{id}/acceptance-criteria [post]
func (h *AcceptanceCriteriaHandler) CreateAcceptanceCriteria(c *gin.Context) {
	userStoryIDParam := c.Param("id")

	// Parse user story ID
	userStoryID, err := uuid.Parse(userStoryIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user story ID format",
		})
		return
	}

	var req service.CreateAcceptanceCriteriaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Set user story ID from URL parameter
	req.UserStoryID = userStoryID

	acceptanceCriteria, err := h.acceptanceCriteriaService.CreateAcceptanceCriteria(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserStoryNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "User story not found",
			})
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Author not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create acceptance criteria",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, acceptanceCriteria)
}

// GetAcceptanceCriteria handles GET /api/v1/acceptance-criteria/:id
func (h *AcceptanceCriteriaHandler) GetAcceptanceCriteria(c *gin.Context) {
	idParam := c.Param("id")

	// Try to parse as UUID first, then as reference ID
	var acceptanceCriteria *models.AcceptanceCriteria
	var err error

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		acceptanceCriteria, err = h.acceptanceCriteriaService.GetAcceptanceCriteriaByID(id)
	} else {
		acceptanceCriteria, err = h.acceptanceCriteriaService.GetAcceptanceCriteriaByReferenceID(idParam)
	}

	if err != nil {
		if errors.Is(err, service.ErrAcceptanceCriteriaNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Acceptance criteria not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get acceptance criteria",
			})
		}
		return
	}

	c.JSON(http.StatusOK, acceptanceCriteria)
}

// UpdateAcceptanceCriteria handles PUT /api/v1/acceptance-criteria/:id
func (h *AcceptanceCriteriaHandler) UpdateAcceptanceCriteria(c *gin.Context) {
	idParam := c.Param("id")

	// Parse ID (UUID only for updates)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid acceptance criteria ID format",
		})
		return
	}

	var req service.UpdateAcceptanceCriteriaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	acceptanceCriteria, err := h.acceptanceCriteriaService.UpdateAcceptanceCriteria(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAcceptanceCriteriaNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Acceptance criteria not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update acceptance criteria",
			})
		}
		return
	}

	c.JSON(http.StatusOK, acceptanceCriteria)
}

// DeleteAcceptanceCriteria handles DELETE /api/v1/acceptance-criteria/:id
func (h *AcceptanceCriteriaHandler) DeleteAcceptanceCriteria(c *gin.Context) {
	idParam := c.Param("id")

	// Parse ID (UUID only for deletes)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid acceptance criteria ID format",
		})
		return
	}

	// Check for force parameter
	force := c.Query("force") == "true"

	err = h.acceptanceCriteriaService.DeleteAcceptanceCriteria(id, force)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAcceptanceCriteriaNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Acceptance criteria not found",
			})
		case errors.Is(err, service.ErrAcceptanceCriteriaHasRequirements):
			c.JSON(http.StatusConflict, gin.H{
				"error": "Acceptance criteria has associated requirements and cannot be deleted",
				"hint":  "Use force=true to delete with dependencies",
			})
		case errors.Is(err, service.ErrUserStoryMustHaveAcceptanceCriteria):
			c.JSON(http.StatusConflict, gin.H{
				"error": "User story must have at least one acceptance criteria",
				"hint":  "Create another acceptance criteria before deleting this one, or use force=true",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete acceptance criteria",
			})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListAcceptanceCriteria handles GET /api/v1/acceptance-criteria
func (h *AcceptanceCriteriaHandler) ListAcceptanceCriteria(c *gin.Context) {
	var filters service.AcceptanceCriteriaFilters

	// Parse query parameters
	if userStoryID := c.Query("user_story_id"); userStoryID != "" {
		if id, err := uuid.Parse(userStoryID); err == nil {
			filters.UserStoryID = &id
		}
	}

	if authorID := c.Query("author_id"); authorID != "" {
		if id, err := uuid.Parse(authorID); err == nil {
			filters.AuthorID = &id
		}
	}

	if orderBy := c.Query("order_by"); orderBy != "" {
		filters.OrderBy = orderBy
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filters.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filters.Offset = o
		}
	}

	acceptanceCriteria, err := h.acceptanceCriteriaService.ListAcceptanceCriteria(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list acceptance criteria",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"acceptance_criteria": acceptanceCriteria,
		"count":               len(acceptanceCriteria),
	})
}

// GetAcceptanceCriteriaByUserStory handles GET /api/v1/user-stories/:id/acceptance-criteria
// @Summary Get acceptance criteria for a user story
// @Description Retrieve all acceptance criteria that belong to a specific user story. This endpoint provides a list of testable conditions that define when the user story is considered complete.
// @Tags user-stories
// @Accept json
// @Produce json
// @Param id path string true "User story UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Success 200 {object} map[string]interface{} "Successfully retrieved acceptance criteria list with count"
// @Failure 400 {object} map[string]interface{} "Invalid user story ID format (UUID required)"
// @Failure 404 {object} map[string]interface{} "User story not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user-stories/{id}/acceptance-criteria [get]
func (h *AcceptanceCriteriaHandler) GetAcceptanceCriteriaByUserStory(c *gin.Context) {
	userStoryIDParam := c.Param("id")

	// Try to parse as UUID first, then as reference ID
	var userStoryID uuid.UUID
	var err error

	if id, parseErr := uuid.Parse(userStoryIDParam); parseErr == nil {
		userStoryID = id
	} else {
		// For reference ID, we need to resolve it first
		// This would require the user story service, but for now we'll return an error
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Please use UUID for user story ID in this endpoint",
		})
		return
	}

	acceptanceCriteria, err := h.acceptanceCriteriaService.GetAcceptanceCriteriaByUserStory(userStoryID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserStoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User story not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get acceptance criteria for user story",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"acceptance_criteria": acceptanceCriteria,
		"count":               len(acceptanceCriteria),
	})
}

// GetAcceptanceCriteriaByAuthor handles GET /api/v1/users/:id/acceptance-criteria
func (h *AcceptanceCriteriaHandler) GetAcceptanceCriteriaByAuthor(c *gin.Context) {
	authorIDParam := c.Param("id")

	// Parse author ID (UUID only)
	authorID, err := uuid.Parse(authorIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid author ID format",
		})
		return
	}

	acceptanceCriteria, err := h.acceptanceCriteriaService.GetAcceptanceCriteriaByAuthor(authorID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Author not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get acceptance criteria for author",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"acceptance_criteria": acceptanceCriteria,
		"count":               len(acceptanceCriteria),
	})
}
