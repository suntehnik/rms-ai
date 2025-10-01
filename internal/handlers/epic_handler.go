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

// EpicHandler handles HTTP requests for epic operations
type EpicHandler struct {
	epicService service.EpicService
}

// NewEpicHandler creates a new epic handler instance
func NewEpicHandler(epicService service.EpicService) *EpicHandler {
	return &EpicHandler{
		epicService: epicService,
	}
}

// CreateEpic handles POST /api/v1/epics
// @Summary Create a new epic
// @Description Create a new epic with the provided details. The epic will be assigned a unique reference ID (EP-XXX format) and default status of "Backlog". Requires User or Administrator role.
// @Tags epics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param epic body service.CreateEpicRequest true "Epic creation request"
// @Success 201 {object} models.Epic "Successfully created epic"
// @Failure 400 {object} map[string]interface{} "Invalid request body, creator/assignee not found, or invalid priority"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 403 {object} map[string]interface{} "User or Administrator role required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/epics [post]
func (h *EpicHandler) CreateEpic(c *gin.Context) {
	var req service.CreateEpicRequest
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
	creatorID, ok := auth.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "AUTHENTICATION_REQUIRED",
				"message": "User authentication required",
			},
		})
		return
	}

	// Set the creator ID from the authenticated user
	req.CreatorID = uuid.MustParse(creatorID)

	epic, err := h.epicService.CreateEpic(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "Creator or assignee not found",
				},
			})
		case errors.Is(err, service.ErrInvalidPriority):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": "Invalid priority value",
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to create epic",
				},
			})
		}
		return
	}

	c.JSON(http.StatusCreated, epic)
}

// GetEpic handles GET /api/v1/epics/:id
// @Summary Get an epic by ID or reference ID
// @Description Retrieve a single epic by its UUID or reference ID (e.g., EP-001). Supports both formats for flexible access. Requires authentication.
// @Tags epics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Epic ID (UUID) or reference ID (EP-XXX)" example("123e4567-e89b-12d3-a456-426614174000")
// @Success 200 {object} models.Epic "Epic found successfully"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Epic not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/epics/{id} [get]
func (h *EpicHandler) GetEpic(c *gin.Context) {
	idParam := c.Param("id")

	// Try to parse as UUID first, then as reference ID
	var epic *models.Epic
	var err error

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		epic, err = h.epicService.GetEpicByID(id)
	} else {
		epic, err = h.epicService.GetEpicByReferenceID(idParam)
	}

	if err != nil {
		if errors.Is(err, service.ErrEpicNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "Epic not found",
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to get epic",
				},
			})
		}
		return
	}

	c.JSON(http.StatusOK, epic)
}

// UpdateEpic handles PUT /api/v1/epics/:id
// @Summary Update an existing epic
// @Description Update an epic's properties. Only provided fields will be updated. Supports partial updates with validation for status transitions and priority values. Requires User or Administrator role.
// @Tags epics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Epic UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param epic body service.UpdateEpicRequest true "Epic update request"
// @Success 200 {object} models.Epic "Epic updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body, epic ID format, assignee not found, invalid priority, invalid status, or invalid status transition"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 403 {object} map[string]interface{} "User or Administrator role required"
// @Failure 404 {object} map[string]interface{} "Epic not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/epics/{id} [put]
func (h *EpicHandler) UpdateEpic(c *gin.Context) {
	idParam := c.Param("id")

	// Parse ID (UUID only for updates)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid epic ID format",
			},
		})
		return
	}

	var req service.UpdateEpicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	epic, err := h.epicService.UpdateEpic(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEpicNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "Epic not found",
				},
			})
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "Assignee not found",
				},
			})
		case errors.Is(err, service.ErrInvalidPriority):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": "Invalid priority value",
				},
			})
		case errors.Is(err, service.ErrInvalidEpicStatus):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": "Invalid epic status",
				},
			})
		case errors.Is(err, service.ErrInvalidStatusTransition):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": "Invalid status transition",
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to update epic",
				},
			})
		}
		return
	}

	c.JSON(http.StatusOK, epic)
}

// DeleteEpic handles DELETE /api/v1/epics/:id
// @Summary Delete an epic
// @Description Delete an epic by UUID. By default, epics with associated user stories cannot be deleted unless force=true is specified. Requires User or Administrator role.
// @Tags epics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Epic UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param force query boolean false "Force delete even if epic has user stories" example(false)
// @Success 204 "Epic deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid epic ID format"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 403 {object} map[string]interface{} "User or Administrator role required"
// @Failure 404 {object} map[string]interface{} "Epic not found"
// @Failure 409 {object} map[string]interface{} "Epic has associated user stories and cannot be deleted (use force=true to override)"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/epics/{id} [delete]
func (h *EpicHandler) DeleteEpic(c *gin.Context) {
	idParam := c.Param("id")

	// Parse ID (UUID only for deletes)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid epic ID format",
			},
		})
		return
	}

	// Check for force parameter
	force := c.Query("force") == "true"

	err = h.epicService.DeleteEpic(id, force)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEpicNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "Epic not found",
				},
			})
		case errors.Is(err, service.ErrEpicHasUserStories):
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{
					"code":    "DELETION_CONFLICT",
					"message": "Epic has associated user stories and cannot be deleted. Use force=true to delete with dependencies",
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to delete epic",
				},
			})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// ListEpics handles GET /api/v1/epics
// @Summary List epics with filtering and pagination
// @Description Retrieve a list of epics with optional filtering by creator, assignee, status, and priority. Supports pagination and custom ordering. Requires authentication.
// @Tags epics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param creator_id query string false "Filter by creator UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174001")
// @Param assignee_id query string false "Filter by assignee UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174002")
// @Param status query string false "Filter by epic status" Enums(Backlog,Draft,In Progress,Done,Cancelled) example("Backlog")
// @Param priority query integer false "Filter by priority level" minimum(1) maximum(4) example(1)
// @Param order_by query string false "Order results by field" example("created_at DESC")
// @Param limit query integer false "Maximum number of results to return" minimum(1) maximum(100) default(50) example(20)
// @Param offset query integer false "Number of results to skip for pagination" minimum(0) default(0) example(0)
// @Success 200 {object} map[string]interface{} "List of epics with count"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/epics [get]
func (h *EpicHandler) ListEpics(c *gin.Context) {
	var filters service.EpicFilters

	// Parse query parameters
	if creatorID := c.Query("creator_id"); creatorID != "" {
		if id, err := uuid.Parse(creatorID); err == nil {
			filters.CreatorID = &id
		}
	}

	if assigneeID := c.Query("assignee_id"); assigneeID != "" {
		if id, err := uuid.Parse(assigneeID); err == nil {
			filters.AssigneeID = &id
		}
	}

	if status := c.Query("status"); status != "" {
		epicStatus := models.EpicStatus(status)
		filters.Status = &epicStatus
	}

	if priority := c.Query("priority"); priority != "" {
		if p, err := strconv.Atoi(priority); err == nil && p >= 1 && p <= 4 {
			prio := models.Priority(p)
			filters.Priority = &prio
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

	epics, totalCount, err := h.epicService.ListEpics(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to list epics",
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
		"data":        epics,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      filters.Offset,
	})
}

// GetEpicWithUserStories handles GET /api/v1/epics/:id/user-stories
// @Summary Get an epic with its user stories
// @Description Retrieve an epic along with all its associated user stories. This endpoint provides a hierarchical view of the epic and its child user stories, including their acceptance criteria and requirements if available.
// @Tags epics
// @Accept json
// @Produce json
// @Param id path string true "Epic ID (UUID) or reference ID (EP-XXX)" example("123e4567-e89b-12d3-a456-426614174000")
// @Success 200 {object} models.Epic "Epic with user stories retrieved successfully"
// @Failure 404 {object} map[string]interface{} "Epic not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/epics/{id}/user-stories [get]
func (h *EpicHandler) GetEpicWithUserStories(c *gin.Context) {
	idParam := c.Param("id")

	// Try to parse as UUID first, then as reference ID
	var epic *models.Epic
	var err error

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		epic, err = h.epicService.GetEpicWithUserStories(id)
	} else {
		// For reference ID, first get the epic, then get with user stories
		if tempEpic, tempErr := h.epicService.GetEpicByReferenceID(idParam); tempErr == nil {
			epic, err = h.epicService.GetEpicWithUserStories(tempEpic.ID)
		} else {
			err = tempErr
		}
	}

	if err != nil {
		if errors.Is(err, service.ErrEpicNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "Epic not found",
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to get epic with user stories",
				},
			})
		}
		return
	}

	c.JSON(http.StatusOK, epic)
}

// ChangeEpicStatus handles PATCH /api/v1/epics/:id/status
// @Summary Change the status of an epic
// @Description Update the workflow status of an epic. The system validates status transitions and ensures only valid status changes are allowed. All status transitions are currently permitted as per business requirements.
// @Tags epics
// @Accept json
// @Produce json
// @Param id path string true "Epic UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param status body service.ChangeEpicStatusRequest true "Status change request"
// @Success 200 {object} models.Epic "Epic status updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid epic ID format, request body, epic status, or status transition"
// @Failure 404 {object} map[string]interface{} "Epic not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/epics/{id}/status [patch]
func (h *EpicHandler) ChangeEpicStatus(c *gin.Context) {
	idParam := c.Param("id")

	// Parse ID (UUID only for status changes)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid epic ID format",
			},
		})
		return
	}

	var req service.ChangeEpicStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	epic, err := h.epicService.ChangeEpicStatus(id, req.Status)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEpicNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "Epic not found",
				},
			})
		case errors.Is(err, service.ErrInvalidEpicStatus):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": "Invalid epic status",
				},
			})
		case errors.Is(err, service.ErrInvalidStatusTransition):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": "Invalid status transition",
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to change epic status",
				},
			})
		}
		return
	}

	c.JSON(http.StatusOK, epic)
}

// AssignEpic handles PATCH /api/v1/epics/:id/assign
// @Summary Assign an epic to a user
// @Description Assign an epic to a specific user by updating the assignee_id. The system validates that the assignee user exists before making the assignment.
// @Tags epics
// @Accept json
// @Produce json
// @Param id path string true "Epic UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param assignment body service.AssignEpicRequest true "Assignment request"
// @Success 200 {object} models.Epic "Epic assigned successfully"
// @Failure 400 {object} map[string]interface{} "Invalid epic ID format, request body, or assignee not found"
// @Failure 404 {object} map[string]interface{} "Epic not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/epics/{id}/assign [patch]
func (h *EpicHandler) AssignEpic(c *gin.Context) {
	idParam := c.Param("id")

	// Parse ID (UUID only for assignments)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid epic ID format",
			},
		})
		return
	}

	var req service.AssignEpicRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	epic, err := h.epicService.AssignEpic(id, req.AssigneeID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEpicNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "Epic not found",
				},
			})
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "Assignee not found",
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to assign epic",
				},
			})
		}
		return
	}

	c.JSON(http.StatusOK, epic)
}
