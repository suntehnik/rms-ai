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

// UserStoryHandler handles HTTP requests for user story operations
type UserStoryHandler struct {
	userStoryService service.UserStoryService
}

// NewUserStoryHandler creates a new user story handler instance
func NewUserStoryHandler(userStoryService service.UserStoryService) *UserStoryHandler {
	return &UserStoryHandler{
		userStoryService: userStoryService,
	}
}

// CreateUserStory handles POST /api/v1/user-stories
func (h *UserStoryHandler) CreateUserStory(c *gin.Context) {
	var req service.CreateUserStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate that EpicID is provided for direct user story creation
	if req.EpicID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "epic_id is required",
		})
		return
	}

	userStory, err := h.userStoryService.CreateUserStory(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Creator or assignee not found",
			})
		case errors.Is(err, service.ErrEpicNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Epic not found",
			})
		case errors.Is(err, service.ErrInvalidPriority):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid priority value",
			})
		case errors.Is(err, service.ErrInvalidUserStoryTemplate):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "User story description must follow template: 'As [role], I want [function], so that [goal]'",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create user story",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, userStory)
}

// CreateUserStoryInEpic handles POST /api/v1/epics/:id/user-stories
func (h *UserStoryHandler) CreateUserStoryInEpic(c *gin.Context) {
	epicIDParam := c.Param("id")
	
	// Try to parse as UUID first, then as reference ID
	var epicID uuid.UUID
	var err error
	
	if id, parseErr := uuid.Parse(epicIDParam); parseErr == nil {
		epicID = id
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid epic ID format",
		})
		return
	}

	var req service.CreateUserStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Override epic ID from URL parameter
	req.EpicID = epicID

	userStory, err := h.userStoryService.CreateUserStory(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Creator or assignee not found",
			})
		case errors.Is(err, service.ErrEpicNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Epic not found",
			})
		case errors.Is(err, service.ErrInvalidPriority):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid priority value",
			})
		case errors.Is(err, service.ErrInvalidUserStoryTemplate):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "User story description must follow template: 'As [role], I want [function], so that [goal]'",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create user story",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, userStory)
}

// GetUserStory handles GET /api/v1/user-stories/:id
func (h *UserStoryHandler) GetUserStory(c *gin.Context) {
	idParam := c.Param("id")
	
	// Try to parse as UUID first, then as reference ID
	var userStory *models.UserStory
	var err error
	
	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		userStory, err = h.userStoryService.GetUserStoryByID(id)
	} else {
		userStory, err = h.userStoryService.GetUserStoryByReferenceID(idParam)
	}

	if err != nil {
		if errors.Is(err, service.ErrUserStoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User story not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get user story",
			})
		}
		return
	}

	c.JSON(http.StatusOK, userStory)
}

// UpdateUserStory handles PUT /api/v1/user-stories/:id
func (h *UserStoryHandler) UpdateUserStory(c *gin.Context) {
	idParam := c.Param("id")
	
	// Parse ID (UUID only for updates)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user story ID format",
		})
		return
	}

	var req service.UpdateUserStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	userStory, err := h.userStoryService.UpdateUserStory(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserStoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User story not found",
			})
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Assignee not found",
			})
		case errors.Is(err, service.ErrInvalidPriority):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid priority value",
			})
		case errors.Is(err, service.ErrInvalidUserStoryStatus):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user story status",
			})
		case errors.Is(err, service.ErrInvalidStatusTransition):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid status transition",
			})
		case errors.Is(err, service.ErrInvalidUserStoryTemplate):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "User story description must follow template: 'As [role], I want [function], so that [goal]'",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update user story",
			})
		}
		return
	}

	c.JSON(http.StatusOK, userStory)
}

// DeleteUserStory handles DELETE /api/v1/user-stories/:id
func (h *UserStoryHandler) DeleteUserStory(c *gin.Context) {
	idParam := c.Param("id")
	
	// Parse ID (UUID only for deletes)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user story ID format",
		})
		return
	}

	// Check for force parameter
	force := c.Query("force") == "true"

	err = h.userStoryService.DeleteUserStory(id, force)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserStoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User story not found",
			})
		case errors.Is(err, service.ErrUserStoryHasRequirements):
			c.JSON(http.StatusConflict, gin.H{
				"error": "User story has associated requirements and cannot be deleted",
				"hint":  "Use force=true to delete with dependencies",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete user story",
			})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListUserStories handles GET /api/v1/user-stories
func (h *UserStoryHandler) ListUserStories(c *gin.Context) {
	var filters service.UserStoryFilters

	// Parse query parameters
	if epicID := c.Query("epic_id"); epicID != "" {
		if id, err := uuid.Parse(epicID); err == nil {
			filters.EpicID = &id
		}
	}

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
		userStoryStatus := models.UserStoryStatus(status)
		filters.Status = &userStoryStatus
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

	userStories, err := h.userStoryService.ListUserStories(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list user stories",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_stories": userStories,
		"count":        len(userStories),
	})
}

// GetUserStoryWithAcceptanceCriteria handles GET /api/v1/user-stories/:id/acceptance-criteria
func (h *UserStoryHandler) GetUserStoryWithAcceptanceCriteria(c *gin.Context) {
	idParam := c.Param("id")
	
	// Try to parse as UUID first, then as reference ID
	var userStory *models.UserStory
	var err error
	
	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		userStory, err = h.userStoryService.GetUserStoryWithAcceptanceCriteria(id)
	} else {
		// For reference ID, first get the user story, then get with acceptance criteria
		if tempUserStory, tempErr := h.userStoryService.GetUserStoryByReferenceID(idParam); tempErr == nil {
			userStory, err = h.userStoryService.GetUserStoryWithAcceptanceCriteria(tempUserStory.ID)
		} else {
			err = tempErr
		}
	}

	if err != nil {
		if errors.Is(err, service.ErrUserStoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User story not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get user story with acceptance criteria",
			})
		}
		return
	}

	c.JSON(http.StatusOK, userStory)
}

// GetUserStoryWithRequirements handles GET /api/v1/user-stories/:id/requirements
func (h *UserStoryHandler) GetUserStoryWithRequirements(c *gin.Context) {
	idParam := c.Param("id")
	
	// Try to parse as UUID first, then as reference ID
	var userStory *models.UserStory
	var err error
	
	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		userStory, err = h.userStoryService.GetUserStoryWithRequirements(id)
	} else {
		// For reference ID, first get the user story, then get with requirements
		if tempUserStory, tempErr := h.userStoryService.GetUserStoryByReferenceID(idParam); tempErr == nil {
			userStory, err = h.userStoryService.GetUserStoryWithRequirements(tempUserStory.ID)
		} else {
			err = tempErr
		}
	}

	if err != nil {
		if errors.Is(err, service.ErrUserStoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User story not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get user story with requirements",
			})
		}
		return
	}

	c.JSON(http.StatusOK, userStory)
}

// ChangeUserStoryStatus handles PATCH /api/v1/user-stories/:id/status
func (h *UserStoryHandler) ChangeUserStoryStatus(c *gin.Context) {
	idParam := c.Param("id")
	
	// Parse ID (UUID only for status changes)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user story ID format",
		})
		return
	}

	var req struct {
		Status models.UserStoryStatus `json:"status" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	userStory, err := h.userStoryService.ChangeUserStoryStatus(id, req.Status)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserStoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User story not found",
			})
		case errors.Is(err, service.ErrInvalidUserStoryStatus):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user story status",
			})
		case errors.Is(err, service.ErrInvalidStatusTransition):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid status transition",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to change user story status",
			})
		}
		return
	}

	c.JSON(http.StatusOK, userStory)
}

// AssignUserStory handles PATCH /api/v1/user-stories/:id/assign
func (h *UserStoryHandler) AssignUserStory(c *gin.Context) {
	idParam := c.Param("id")
	
	// Parse ID (UUID only for assignments)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user story ID format",
		})
		return
	}

	var req struct {
		AssigneeID uuid.UUID `json:"assignee_id" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	userStory, err := h.userStoryService.AssignUserStory(id, req.AssigneeID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserStoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User story not found",
			})
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Assignee not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to assign user story",
			})
		}
		return
	}

	c.JSON(http.StatusOK, userStory)
}