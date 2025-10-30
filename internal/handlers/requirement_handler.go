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

type RequirementRelationshipListResponse = ListResponse[models.RequirementRelationship]
type RequirementListResponse = ListResponse[models.Requirement]

// RequirementHandler handles HTTP requests for requirement operations
type RequirementHandler struct {
	requirementService service.RequirementService
}

// NewRequirementHandler creates a new requirement handler instance
func NewRequirementHandler(requirementService service.RequirementService) *RequirementHandler {
	return &RequirementHandler{
		requirementService: requirementService,
	}
}

// CreateRequirement handles both POST /api/v1/requirements and POST /api/v1/user-stories/:id/requirements
// @Summary Create a requirement (standalone or within a user story)
// @Description Create a new detailed requirement. When called via /api/v1/user-stories/:id/requirements, the user story ID from the URL path will be used as the parent. When called via /api/v1/requirements, the user_story_id must be provided in the request body.
// @Tags requirements,user-stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string false "User story UUID (only for nested creation)" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param requirement body service.CreateRequirementRequest true "Requirement creation request"
// @Success 201 {object} models.Requirement "Successfully created requirement"
// @Failure 400 {object} map[string]interface{} "Invalid user story ID format, request body, creator/assignee not found, user story not found, requirement type not found, acceptance criteria not found, or invalid priority"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/requirements [post]
// @Router /api/v1/user-stories/{id}/requirements [post]
func (h *RequirementHandler) CreateRequirement(c *gin.Context) {
	// Check if this is a nested creation (user story ID in path)
	userStoryIDParam := c.Param("id")
	var isNestedCreation bool
	var userStoryID uuid.UUID
	var err error

	if userStoryIDParam != "" {
		// This is nested creation: POST /api/v1/user-stories/:id/requirements
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

	var req service.CreateRequirementRequest
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

	// Set the creator ID from the authenticated user
	req.CreatorID = uuid.MustParse(creatorID)

	// For nested creation, override the user story ID from the URL path
	if isNestedCreation {
		req.UserStoryID = userStoryID
	}

	requirement, err := h.requirementService.CreateRequirement(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Creator or assignee not found",
			})
		case errors.Is(err, service.ErrUserStoryNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "User story not found",
			})
		case errors.Is(err, service.ErrRequirementTypeNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Requirement type not found",
			})
		case errors.Is(err, service.ErrAcceptanceCriteriaNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Acceptance criteria not found",
			})
		case errors.Is(err, service.ErrInvalidPriority):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid priority value",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create requirement",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, requirement)
}

// GetRequirement handles GET /api/v1/requirements/:id
// @Summary Get a requirement by ID or reference ID
// @Description Retrieve a specific requirement by its UUID or human-readable reference ID (e.g., REQ-001). Returns the requirement with all its properties and relationships.
// @Tags requirements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Requirement UUID or reference ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Success 200 {object} models.Requirement "Successfully retrieved requirement"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Requirement not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/requirements/{id} [get]
func (h *RequirementHandler) GetRequirement(c *gin.Context) {
	idParam := c.Param("id")

	// Try to parse as UUID first, then as reference ID
	var requirement *models.Requirement
	var err error

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		requirement, err = h.requirementService.GetRequirementByID(id)
	} else {
		requirement, err = h.requirementService.GetRequirementByReferenceID(idParam)
	}

	if err != nil {
		if errors.Is(err, service.ErrRequirementNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Requirement not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get requirement",
			})
		}
		return
	}

	c.JSON(http.StatusOK, requirement)
}

// UpdateRequirement handles PUT /api/v1/requirements/:id
// @Summary Update an existing requirement
// @Description Update a requirement's properties including acceptance criteria, assignee, priority, status, type, title, and description. Only provided fields will be updated. Status transitions are validated according to business rules.
// @Tags requirements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Requirement UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param requirement body service.UpdateRequirementRequest true "Requirement update request with optional fields"
// @Success 200 {object} models.Requirement "Successfully updated requirement"
// @Failure 400 {object} map[string]interface{} "Invalid requirement ID format, request body, assignee not found, requirement type not found, acceptance criteria not found, invalid priority, invalid requirement status, or invalid status transition"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Requirement not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/requirements/{id} [put]
func (h *RequirementHandler) UpdateRequirement(c *gin.Context) {
	idParam := c.Param("id")

	// Parse ID (UUID only for updates)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid requirement ID format",
		})
		return
	}

	var req service.UpdateRequirementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	requirement, err := h.requirementService.UpdateRequirement(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRequirementNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Requirement not found",
			})
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Assignee not found",
			})
		case errors.Is(err, service.ErrRequirementTypeNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Requirement type not found",
			})
		case errors.Is(err, service.ErrAcceptanceCriteriaNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Acceptance criteria not found",
			})
		case errors.Is(err, service.ErrInvalidPriority):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid priority value",
			})
		case errors.Is(err, service.ErrInvalidRequirementStatus):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid requirement status",
			})
		case errors.Is(err, service.ErrInvalidStatusTransition):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid status transition",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update requirement",
			})
		}
		return
	}

	c.JSON(http.StatusOK, requirement)
}

// DeleteRequirement handles DELETE /api/v1/requirements/:id
// @Summary Delete a requirement
// @Description Delete a requirement by its UUID. By default, deletion is prevented if the requirement has associated relationships. Use force=true query parameter to delete with all dependencies.
// @Tags requirements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Requirement UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param force query boolean false "Force delete with dependencies" example(false)
// @Success 204 "Successfully deleted requirement"
// @Failure 400 {object} map[string]interface{} "Invalid requirement ID format"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Requirement not found"
// @Failure 409 {object} map[string]interface{} "Requirement has associated relationships and cannot be deleted (use force=true)"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/requirements/{id} [delete]
func (h *RequirementHandler) DeleteRequirement(c *gin.Context) {
	idParam := c.Param("id")

	// Parse ID (UUID only for deletes)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid requirement ID format",
		})
		return
	}

	// Check for force parameter
	force := c.Query("force") == "true"

	err = h.requirementService.DeleteRequirement(id, force)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRequirementNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Requirement not found",
			})
		case errors.Is(err, service.ErrRequirementHasRelationships):
			c.JSON(http.StatusConflict, gin.H{
				"error": "Requirement has associated relationships and cannot be deleted",
				"hint":  "Use force=true to delete with dependencies",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete requirement",
			})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListRequirements handles GET /api/v1/requirements
// @Summary List requirements with filtering and pagination
// @Description Retrieve a list of requirements with optional filtering by user story, acceptance criteria, creator, assignee, status, priority, and type. Supports pagination and custom ordering.
// @Tags requirements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_story_id query string false "Filter by user story UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param acceptance_criteria_id query string false "Filter by acceptance criteria UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174001")
// @Param creator_id query string false "Filter by creator UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174002")
// @Param assignee_id query string false "Filter by assignee UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174003")
// @Param status query string false "Filter by requirement status" Enums(draft, in_review, approved, implemented, tested, rejected) example("draft")
// @Param priority query integer false "Filter by priority level" minimum(1) maximum(4) example(2)
// @Param type_id query string false "Filter by requirement type UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174004")
// @Param order_by query string false "Order by field (e.g., 'created_at DESC', 'priority ASC')" example("created_at DESC")
// @Param limit query integer false "Maximum number of results" minimum(1) maximum(100) example(50)
// @Param offset query integer false "Number of results to skip" minimum(0) example(0)
// @Success 200 {object} map[string]interface{} "Successfully retrieved requirements list with pagination info"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/requirements [get]
func (h *RequirementHandler) ListRequirements(c *gin.Context) {
	var filters service.RequirementFilters

	// Parse query parameters
	if userStoryID := c.Query("user_story_id"); userStoryID != "" {
		if id, err := uuid.Parse(userStoryID); err == nil {
			filters.UserStoryID = &id
		}
	}

	if acceptanceCriteriaID := c.Query("acceptance_criteria_id"); acceptanceCriteriaID != "" {
		if id, err := uuid.Parse(acceptanceCriteriaID); err == nil {
			filters.AcceptanceCriteriaID = &id
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
		requirementStatus := models.RequirementStatus(status)
		filters.Status = &requirementStatus
	}

	if priority := c.Query("priority"); priority != "" {
		if p, err := strconv.Atoi(priority); err == nil && p >= 1 && p <= 4 {
			prio := models.Priority(p)
			filters.Priority = &prio
		}
	}

	if typeID := c.Query("type_id"); typeID != "" {
		if id, err := uuid.Parse(typeID); err == nil {
			filters.TypeID = &id
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

	requirements, totalCount, err := h.requirementService.ListRequirements(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to list requirements",
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
		"data":        requirements,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      filters.Offset,
	})
}

// GetRequirementWithRelationships handles GET /api/v1/requirements/:id/relationships
// @Summary Get a requirement with all its relationships
// @Description Retrieve a specific requirement by its UUID or reference ID, including all incoming and outgoing relationships with other requirements. This provides a complete view of requirement dependencies.
// @Tags requirements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Requirement UUID or reference ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Success 200 {object} models.Requirement "Successfully retrieved requirement with relationships"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Requirement not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/requirements/{id}/relationships [get]
func (h *RequirementHandler) GetRequirementWithRelationships(c *gin.Context) {
	idParam := c.Param("id")

	// Try to parse as UUID first, then as reference ID
	var requirement *models.Requirement
	var err error

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		requirement, err = h.requirementService.GetRequirementWithRelationships(id)
	} else {
		// For reference ID, first get the requirement, then get with relationships
		if tempRequirement, tempErr := h.requirementService.GetRequirementByReferenceID(idParam); tempErr == nil {
			requirement, err = h.requirementService.GetRequirementWithRelationships(tempRequirement.ID)
		} else {
			err = tempErr
		}
	}

	if err != nil {
		if errors.Is(err, service.ErrRequirementNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Requirement not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get requirement with relationships",
			})
		}
		return
	}

	c.JSON(http.StatusOK, requirement)
}

// ChangeRequirementStatus handles PATCH /api/v1/requirements/:id/status
// @Summary Change requirement status
// @Description Update the status of a requirement. Status transitions are validated according to business rules to ensure proper workflow progression (e.g., draft → in_review → approved → implemented → tested).
// @Tags requirements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Requirement UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param status body object true "Status change request" example({"status":"in_review"})
// @Success 200 {object} models.Requirement "Successfully changed requirement status"
// @Failure 400 {object} map[string]interface{} "Invalid requirement ID format, request body, invalid requirement status, or invalid status transition"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Requirement not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/requirements/{id}/status [patch]
func (h *RequirementHandler) ChangeRequirementStatus(c *gin.Context) {
	idParam := c.Param("id")

	// Parse ID (UUID only for status changes)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid requirement ID format",
		})
		return
	}

	var req struct {
		Status models.RequirementStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	requirement, err := h.requirementService.ChangeRequirementStatus(id, req.Status)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRequirementNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Requirement not found",
			})
		case errors.Is(err, service.ErrInvalidRequirementStatus):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid requirement status",
			})
		case errors.Is(err, service.ErrInvalidStatusTransition):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid status transition",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to change requirement status",
			})
		}
		return
	}

	c.JSON(http.StatusOK, requirement)
}

// AssignRequirement handles PATCH /api/v1/requirements/:id/assign
// @Summary Assign requirement to a user
// @Description Assign a requirement to a specific user by updating the assignee field. The assignee must be a valid user in the system.
// @Tags requirements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Requirement UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Param assignment body object true "Assignment request" example({"assignee_id":"123e4567-e89b-12d3-a456-426614174001"})
// @Success 200 {object} models.Requirement "Successfully assigned requirement"
// @Failure 400 {object} map[string]interface{} "Invalid requirement ID format, request body, or assignee not found"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Requirement not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/requirements/{id}/assign [patch]
func (h *RequirementHandler) AssignRequirement(c *gin.Context) {
	idParam := c.Param("id")

	// Parse ID (UUID only for assignments)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid requirement ID format",
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

	requirement, err := h.requirementService.AssignRequirement(id, req.AssigneeID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRequirementNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Requirement not found",
			})
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Assignee not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to assign requirement",
			})
		}
		return
	}

	c.JSON(http.StatusOK, requirement)
}

// CreateRelationship handles POST /api/v1/requirements/:id/relationships
// @Summary Create a relationship between requirements
// @Description Create a typed relationship between two requirements (e.g., depends_on, blocks, relates_to, conflicts_with, derives_from). Prevents circular relationships and duplicate relationships. Validates that both requirements and the relationship type exist.
// @Tags requirements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param relationship body service.CreateRelationshipRequest true "Relationship creation request with source, target, type, and creator"
// @Success 201 {object} models.RequirementRelationship "Successfully created requirement relationship"
// @Failure 400 {object} map[string]interface{} "Invalid request body, source/target requirement not found, relationship type not found, creator not found, or circular relationship detected"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 409 {object} map[string]interface{} "Relationship already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/requirements/relationships [post]
func (h *RequirementHandler) CreateRelationship(c *gin.Context) {
	var req service.CreateRelationshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	relationship, err := h.requirementService.CreateRelationship(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRequirementNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Source or target requirement not found",
			})
		case errors.Is(err, service.ErrRelationshipTypeNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Relationship type not found",
			})
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Creator not found",
			})
		case errors.Is(err, service.ErrCircularRelationship):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Cannot create relationship between the same requirement",
			})
		case errors.Is(err, service.ErrDuplicateRelationship):
			c.JSON(http.StatusConflict, gin.H{
				"error": "Relationship already exists",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create relationship",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, relationship)
}

// DeleteRelationship handles DELETE /api/v1/requirement-relationships/:id
// @Summary Delete a requirement relationship
// @Description Delete a specific relationship between requirements by its UUID. This removes the dependency or association between the two requirements.
// @Tags requirements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Relationship UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174000")
// @Success 204 "Successfully deleted relationship"
// @Failure 400 {object} map[string]interface{} "Invalid relationship ID format"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Relationship not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/requirement-relationships/{id} [delete]
func (h *RequirementHandler) DeleteRelationship(c *gin.Context) {
	idParam := c.Param("id")

	// Parse ID (UUID only for deletes)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid relationship ID format",
		})
		return
	}

	err = h.requirementService.DeleteRelationship(id)
	if err != nil {
		if errors.Is(err, service.ErrRequirementNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Relationship not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete relationship",
			})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetRelationshipsByRequirement handles GET /api/v1/requirements/:id/relationships
// @Summary Get all relationships for a requirement
// @Description Retrieve all incoming and outgoing relationships for a specific requirement by its UUID or reference ID. Returns both relationships where the requirement is the source and where it is the target.
// @Tags requirements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Requirement UUID or reference ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Param limit query integer false "Maximum number of results" minimum(1) maximum(100) example(50)
// @Param offset query integer false "Number of results to skip" minimum(0) example(0)
// @Success 200 {object} RequirementRelationshipListResponse "Successfully retrieved relationships list with pagination"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Requirement not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/requirements/{id}/relationships [get]
func (h *RequirementHandler) GetRelationshipsByRequirement(c *gin.Context) {
	idParam := c.Param("id")

	// Parse pagination parameters
	var limit, offset int
	var err error

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err = strconv.Atoi(limitStr); err != nil || limit < 1 || limit > 100 {
			limit = 50 // Default limit
		}
	} else {
		limit = 50 // Default limit
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err = strconv.Atoi(offsetStr); err != nil || offset < 0 {
			offset = 0 // Default offset
		}
	} else {
		offset = 0 // Default offset
	}

	// Try to parse as UUID first, then as reference ID
	var requirementID uuid.UUID

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		requirementID = id
	} else {
		// For reference ID, first get the requirement ID
		if requirement, tempErr := h.requirementService.GetRequirementByReferenceID(idParam); tempErr == nil {
			requirementID = requirement.ID
		} else {
			err = tempErr
		}
	}

	if err != nil {
		if errors.Is(err, service.ErrRequirementNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Requirement not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get requirement",
			})
		}
		return
	}

	relationships, totalCount, err := h.requirementService.GetRelationshipsByRequirementWithPagination(requirementID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get relationships",
		})
		return
	}

	SendListResponse(c, relationships, totalCount, limit, offset)
}

// SearchRequirements handles GET /api/v1/requirements/search
// @Summary Search requirements by text
// @Description Perform full-text search across requirement titles and descriptions using PostgreSQL's text search capabilities. Returns requirements that match the search query with relevance ranking.
// @Tags requirements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query text" example("authentication validation")
// @Param limit query integer false "Maximum number of results" minimum(1) maximum(100) example(50)
// @Param offset query integer false "Number of results to skip" minimum(0) example(0)
// @Success 200 {object} RequirementListResponse "Successfully retrieved search results with pagination"
// @Failure 400 {object} map[string]interface{} "Search query parameter 'q' is required"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/requirements/search [get]
func (h *RequirementHandler) SearchRequirements(c *gin.Context) {
	searchText := c.Query("q")
	if searchText == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query parameter 'q' is required",
		})
		return
	}

	// Parse pagination parameters
	var limit, offset int
	var err error

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err = strconv.Atoi(limitStr); err != nil || limit < 1 || limit > 100 {
			limit = 50 // Default limit
		}
	} else {
		limit = 50 // Default limit
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err = strconv.Atoi(offsetStr); err != nil || offset < 0 {
			offset = 0 // Default offset
		}
	} else {
		offset = 0 // Default offset
	}

	requirements, totalCount, err := h.requirementService.SearchRequirementsWithPagination(searchText, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to search requirements",
		})
		return
	}

	SendListResponse(c, requirements, totalCount, limit, offset)
}
