package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"product-requirements-management/internal/auth"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

type AcceptanceCriteriaListResponse = ListResponse[models.AcceptanceCriteria]

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

// CreateAcceptanceCriteria handles both POST /api/v1/acceptance-criteria and POST /api/v1/user-stories/:id/acceptance-criteria
// @Summary Create acceptance criteria (standalone or within a user story)
// @Description Create new acceptance criteria. When called via /api/v1/user-stories/:id/acceptance-criteria, the user story ID from the URL path will be used as the parent. When called via /api/v1/acceptance-criteria, the user_story_id must be provided in the request body.
// @Tags acceptance-criteria,user-stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string false "User story UUID (only for nested creation)" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param acceptance_criteria body service.CreateAcceptanceCriteriaRequest true "Acceptance criteria creation request"
// @Success 201 {object} models.AcceptanceCriteria "Successfully created acceptance criteria"
// @Failure 400 {object} map[string]interface{} "Invalid user story ID format, request body, user story not found, or author not found"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/acceptance-criteria [post]
// @Router /api/v1/user-stories/{id}/acceptance-criteria [post]
func (h *AcceptanceCriteriaHandler) CreateAcceptanceCriteria(c *gin.Context) {
	// Check if this is a nested creation (user story ID in path)
	userStoryIDParam := c.Param("id")
	var isNestedCreation bool
	var userStoryID uuid.UUID
	var err error

	if userStoryIDParam != "" {
		// This is nested creation: POST /api/v1/user-stories/:id/acceptance-criteria
		isNestedCreation = true
		if id, parseErr := uuid.Parse(userStoryIDParam); parseErr == nil {
			userStoryID = id
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user story ID format",
			})
			return
		}
	}

	var req service.CreateAcceptanceCriteriaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get current user ID from JWT token context
	authorID, ok := auth.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User authentication required",
		})
		return
	}

	// Set the author ID from the authenticated user
	req.AuthorID = uuid.MustParse(authorID)

	// For nested creation, override the user story ID from the URL path
	if isNestedCreation {
		req.UserStoryID = userStoryID
	}

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
// @Summary Get acceptance criteria by ID or reference ID
// @Description Retrieve specific acceptance criteria by its UUID or human-readable reference ID (e.g., AC-001). Returns the acceptance criteria with all its properties including the testable condition and associated user story.
// @Tags acceptance-criteria
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Acceptance criteria UUID or reference ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Success 200 {object} models.AcceptanceCriteria "Successfully retrieved acceptance criteria"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Acceptance criteria not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/acceptance-criteria/{id} [get]
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
// @Summary Update existing acceptance criteria
// @Description Update acceptance criteria properties including the testable condition text and description. Only provided fields will be updated, maintaining the relationship to the parent user story.
// @Tags acceptance-criteria
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Acceptance criteria UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param acceptance_criteria body service.UpdateAcceptanceCriteriaRequest true "Acceptance criteria update request with optional fields"
// @Success 200 {object} models.AcceptanceCriteria "Successfully updated acceptance criteria"
// @Failure 400 {object} map[string]interface{} "Invalid acceptance criteria ID format or request body"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Acceptance criteria not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/acceptance-criteria/{id} [put]
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
// @Summary Delete acceptance criteria
// @Description Delete acceptance criteria by its UUID. Deletion is prevented if the acceptance criteria has associated requirements or if it's the last acceptance criteria for a user story. Use force=true to override these constraints.
// @Tags acceptance-criteria
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Acceptance criteria UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param force query boolean false "Force delete with dependencies and constraints" example(false)
// @Success 204 "Successfully deleted acceptance criteria"
// @Failure 400 {object} map[string]interface{} "Invalid acceptance criteria ID format"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Acceptance criteria not found"
// @Failure 409 {object} map[string]interface{} "Acceptance criteria has associated requirements or is the last one for user story (use force=true)"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/acceptance-criteria/{id} [delete]
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
// @Summary List acceptance criteria with filtering and pagination
// @Description Retrieve a list of acceptance criteria with optional filtering by user story and author. Supports pagination and custom ordering to help organize testable conditions across the system.
// @Tags acceptance-criteria
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_story_id query string false "Filter by user story UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param author_id query string false "Filter by author UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174001")
// @Param order_by query string false "Order by field (e.g., 'created_at DESC', 'reference_id ASC')" example("created_at DESC")
// @Param limit query integer false "Maximum number of results" minimum(1) maximum(100) example(50)
// @Param offset query integer false "Number of results to skip" minimum(0) example(0)
// @Success 200 {object} map[string]interface{} "Successfully retrieved acceptance criteria list with pagination info"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/acceptance-criteria [get]
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

	acceptanceCriteria, totalCount, err := h.acceptanceCriteriaService.ListAcceptanceCriteria(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to list acceptance criteria",
			},
		})
		return
	}

	// Set default limit if not specified
	limit := 50
	if filters.Limit > 0 {
		limit = filters.Limit
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        acceptanceCriteria,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      filters.Offset,
	})
}

// GetAcceptanceCriteriaByUserStory handles GET /api/v1/user-stories/:id/acceptance-criteria
// @Summary Get acceptance criteria for a user story
// @Description Retrieve all acceptance criteria that belong to a specific user story with pagination support. This endpoint provides a list of testable conditions that define when the user story is considered complete.
// @Tags user-stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User story UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param limit query integer false "Maximum number of results" minimum(1) maximum(100) example(50)
// @Param offset query integer false "Number of results to skip" minimum(0) example(0)
// @Success 200 {object} AcceptanceCriteriaListResponse "Successfully retrieved acceptance criteria list with standardized pagination format"
// @Failure 400 {object} map[string]interface{} "Invalid user story ID format (UUID required)"
// @Failure 401 {object} map[string]interface{} "Authentication required"
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

	// Parse pagination parameters
	var params PaginationParams
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
			params.Limit = l
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid limit parameter (must be between 1 and 100)",
			})
			return
		}
	}
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			params.Offset = o
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid offset parameter (must be >= 0)",
			})
			return
		}
	}
	params.SetDefaults()

	acceptanceCriteria, totalCount, err := h.acceptanceCriteriaService.GetAcceptanceCriteriaByUserStory(userStoryID, params.Limit, params.Offset)
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

	// Use standardized list response format
	SendListResponse(c, acceptanceCriteria, totalCount, params.Limit, params.Offset)
}

// GetAcceptanceCriteriaByAuthor handles GET /api/v1/users/:id/acceptance-criteria
// @Summary Get acceptance criteria by author
// @Description Retrieve all acceptance criteria created by a specific user with pagination support. This endpoint helps track which testable conditions were authored by each team member.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Author UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param limit query integer false "Maximum number of results" minimum(1) maximum(100) example(50)
// @Param offset query integer false "Number of results to skip" minimum(0) example(0)
// @Success 200 {object} AcceptanceCriteriaListResponse "Successfully retrieved acceptance criteria list with standardized pagination format"
// @Failure 400 {object} map[string]interface{} "Invalid author ID format"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Author not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/users/{id}/acceptance-criteria [get]
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

	// Parse pagination parameters
	var params PaginationParams
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
			params.Limit = l
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid limit parameter (must be between 1 and 100)",
			})
			return
		}
	}
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			params.Offset = o
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid offset parameter (must be >= 0)",
			})
			return
		}
	}
	params.SetDefaults()

	acceptanceCriteria, totalCount, err := h.acceptanceCriteriaService.GetAcceptanceCriteriaByAuthor(authorID, params.Limit, params.Offset)
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

	// Use standardized list response format
	SendListResponse(c, acceptanceCriteria, totalCount, params.Limit, params.Offset)
}
