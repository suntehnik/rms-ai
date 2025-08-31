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
func (h *EpicHandler) CreateEpic(c *gin.Context) {
	var req service.CreateEpicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	epic, err := h.epicService.CreateEpic(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Creator or assignee not found",
			})
		case errors.Is(err, service.ErrInvalidPriority):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid priority value",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create epic",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, epic)
}

// GetEpic handles GET /api/v1/epics/:id
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
				"error": "Epic not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get epic",
			})
		}
		return
	}

	c.JSON(http.StatusOK, epic)
}

// UpdateEpic handles PUT /api/v1/epics/:id
func (h *EpicHandler) UpdateEpic(c *gin.Context) {
	idParam := c.Param("id")
	
	// Parse ID (UUID only for updates)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid epic ID format",
		})
		return
	}

	var req service.UpdateEpicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	epic, err := h.epicService.UpdateEpic(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEpicNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Epic not found",
			})
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Assignee not found",
			})
		case errors.Is(err, service.ErrInvalidPriority):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid priority value",
			})
		case errors.Is(err, service.ErrInvalidEpicStatus):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid epic status",
			})
		case errors.Is(err, service.ErrInvalidStatusTransition):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid status transition",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update epic",
			})
		}
		return
	}

	c.JSON(http.StatusOK, epic)
}

// DeleteEpic handles DELETE /api/v1/epics/:id
func (h *EpicHandler) DeleteEpic(c *gin.Context) {
	idParam := c.Param("id")
	
	// Parse ID (UUID only for deletes)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid epic ID format",
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
				"error": "Epic not found",
			})
		case errors.Is(err, service.ErrEpicHasUserStories):
			c.JSON(http.StatusConflict, gin.H{
				"error": "Epic has associated user stories and cannot be deleted",
				"hint": "Use force=true to delete with dependencies",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete epic",
			})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListEpics handles GET /api/v1/epics
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

	epics, err := h.epicService.ListEpics(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list epics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"epics": epics,
		"count": len(epics),
	})
}

// GetEpicWithUserStories handles GET /api/v1/epics/:id/user-stories
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
				"error": "Epic not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get epic with user stories",
			})
		}
		return
	}

	c.JSON(http.StatusOK, epic)
}

// ChangeEpicStatus handles PATCH /api/v1/epics/:id/status
func (h *EpicHandler) ChangeEpicStatus(c *gin.Context) {
	idParam := c.Param("id")
	
	// Parse ID (UUID only for status changes)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid epic ID format",
		})
		return
	}

	var req struct {
		Status models.EpicStatus `json:"status" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	epic, err := h.epicService.ChangeEpicStatus(id, req.Status)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEpicNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Epic not found",
			})
		case errors.Is(err, service.ErrInvalidEpicStatus):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid epic status",
			})
		case errors.Is(err, service.ErrInvalidStatusTransition):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid status transition",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to change epic status",
			})
		}
		return
	}

	c.JSON(http.StatusOK, epic)
}

// AssignEpic handles PATCH /api/v1/epics/:id/assign
func (h *EpicHandler) AssignEpic(c *gin.Context) {
	idParam := c.Param("id")
	
	// Parse ID (UUID only for assignments)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid epic ID format",
		})
		return
	}

	var req struct {
		AssigneeID uuid.UUID `json:"assignee_id" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	epic, err := h.epicService.AssignEpic(id, req.AssigneeID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEpicNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Epic not found",
			})
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Assignee not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to assign epic",
			})
		}
		return
	}

	c.JSON(http.StatusOK, epic)
}