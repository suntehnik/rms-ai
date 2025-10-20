package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"product-requirements-management/internal/auth"
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
// @Summary Create a new user story
// @Description Create a new user story with the provided details. The epic_id must be specified in the request body to establish the parent-child relationship. The user story description should follow the template format: 'As [role], I want [function], so that [goal]'.
// @Tags user-stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_story body service.CreateUserStoryRequest true "User story creation request"
// @Success 201 {object} models.UserStory "Successfully created user story"
// @Failure 400 {object} map[string]interface{} "Invalid request body, epic_id required, creator/assignee not found, epic not found, invalid priority, or invalid user story template"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user-stories [post]
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

	// Get current user ID from JWT token context
	creatorID, ok := auth.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User authentication required",
		})
		return
	}

	// Set the creator ID from the authenticated user
	req.CreatorID = uuid.MustParse(creatorID)

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
// @Summary Create a new user story within an epic
// @Description Create a new user story that belongs to the specified epic. This is a nested resource creation that establishes the parent-child relationship between epic and user story. The epic ID from the URL path will override any epic_id in the request body.
// @Tags epics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Epic UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param user_story body service.CreateUserStoryRequest true "User story creation request (epic_id will be overridden by path parameter)"
// @Success 201 {object} models.UserStory "Successfully created user story within epic"
// @Failure 400 {object} map[string]interface{} "Invalid epic ID format, request body, creator/assignee not found, epic not found, invalid priority, or invalid user story template"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/epics/{id}/user-stories [post]
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

	// Get current user ID from JWT token context
	creatorID, ok := auth.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User authentication required",
		})
		return
	}

	// Set the creator ID from the authenticated user and override epic ID from URL parameter
	req.CreatorID = uuid.MustParse(creatorID)
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
// @Summary Get a user story by ID or reference ID
// @Description Retrieve a specific user story by its UUID or human-readable reference ID (e.g., US-001). Returns the user story with basic information excluding related entities.
// @Tags user-stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User story UUID or reference ID" example("123e4567-e89b-12d3-a456-426614174000") example("US-001")
// @Success 200 {object} models.UserStory "Successfully retrieved user story"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "User story not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user-stories/{id} [get]
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
// @Summary Update a user story
// @Description Update an existing user story by its UUID. All fields in the request body are optional and will only update the provided fields. The user story description should follow the template format if provided.
// @Tags user-stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User story UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param user_story body service.UpdateUserStoryRequest true "User story update request"
// @Success 200 {object} models.UserStory "Successfully updated user story"
// @Failure 400 {object} map[string]interface{} "Invalid user story ID format, request body, assignee not found, invalid priority, invalid status, invalid status transition, or invalid user story template"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "User story not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user-stories/{id} [put]
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
// @Summary Delete a user story
// @Description Delete a user story by its UUID. By default, deletion will fail if the user story has associated requirements or acceptance criteria. Use the force=true query parameter to delete with all dependencies.
// @Tags user-stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User story UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param force query boolean false "Force delete with all dependencies" example(false)
// @Success 204 "Successfully deleted user story"
// @Failure 400 {object} map[string]interface{} "Invalid user story ID format"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "User story not found"
// @Failure 409 {object} map[string]interface{} "User story has associated requirements and cannot be deleted (use force=true)"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user-stories/{id} [delete]
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
// @Summary List user stories with filtering and pagination
// @Description Retrieve a list of user stories with optional filtering by epic, creator, assignee, status, and priority. Supports pagination and custom sorting. Use the include parameter to load related entities.
// @Tags user-stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param epic_id query string false "Filter by epic UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param creator_id query string false "Filter by creator UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174001")
// @Param assignee_id query string false "Filter by assignee UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174002")
// @Param status query string false "Filter by user story status" Enums(Backlog,Draft,In Progress,Done,Cancelled) example("Backlog")
// @Param priority query integer false "Filter by priority level" minimum(1) maximum(4) example(2)
// @Param include query string false "Include related entities (comma-separated)" example("epic,creator,assignee") example("acceptance_criteria,requirements,comments")
// @Param order_by query string false "Sort order for results" example("created_at DESC") example("priority ASC") example("title ASC")
// @Param limit query integer false "Maximum number of results to return" minimum(1) maximum(100) default(50) example(20)
// @Param offset query integer false "Number of results to skip for pagination" minimum(0) default(0) example(0)
// @Success 200 {object} map[string]interface{} "Successfully retrieved user stories list with pagination info"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user-stories [get]
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

	if include := c.Query("include"); include != "" {
		// Split comma-separated includes and trim whitespace
		includes := make([]string, 0)
		for _, inc := range strings.Split(include, ",") {
			trimmed := strings.TrimSpace(inc)
			if trimmed != "" {
				includes = append(includes, trimmed)
			}
		}
		filters.Include = includes
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

	userStories, totalCount, err := h.userStoryService.ListUserStories(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to list user stories",
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
		"data":        userStories,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      filters.Offset,
	})
}

// GetUserStoryWithAcceptanceCriteria handles GET /api/v1/user-stories/:id/acceptance-criteria
// @Summary Get user story with acceptance criteria
// @Description Retrieve a specific user story by its UUID or reference ID, including all associated acceptance criteria. This endpoint provides hierarchical data showing the user story and its testable conditions.
// @Tags user-stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User story UUID or reference ID" example("123e4567-e89b-12d3-a456-426614174000") example("US-001")
// @Success 200 {object} models.UserStory "Successfully retrieved user story with acceptance criteria"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "User story not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user-stories/{id}/acceptance-criteria [get]
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
// @Summary Get user story with requirements
// @Description Retrieve a specific user story by its UUID or reference ID, including all associated detailed requirements. This endpoint provides hierarchical data showing the user story and its technical requirements.
// @Tags user-stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User story UUID or reference ID" example("123e4567-e89b-12d3-a456-426614174000") example("US-001")
// @Success 200 {object} models.UserStory "Successfully retrieved user story with requirements"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "User story not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user-stories/{id}/requirements [get]
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
// @Summary Change user story status
// @Description Update the status of a user story. All status transitions are allowed by default. Valid statuses are: Backlog, Draft, In Progress, Done, Cancelled.
// @Tags user-stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User story UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param status body object true "Status change request" example({"status": "In Progress"})
// @Success 200 {object} models.UserStory "Successfully changed user story status"
// @Failure 400 {object} map[string]interface{} "Invalid user story ID format, request body, invalid status, or invalid status transition"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "User story not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user-stories/{id}/status [patch]
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
// @Summary Assign user story to a user
// @Description Assign a user story to a specific user by updating the assignee_id. The assignee must be a valid user in the system.
// @Tags user-stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User story UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param assignment body object true "Assignment request" example({"assignee_id": "123e4567-e89b-12d3-a456-426614174003"})
// @Success 200 {object} models.UserStory "Successfully assigned user story"
// @Failure 400 {object} map[string]interface{} "Invalid user story ID format, request body, or assignee not found"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "User story not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user-stories/{id}/assign [patch]
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
