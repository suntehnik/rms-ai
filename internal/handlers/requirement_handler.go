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

// CreateRequirement handles POST /api/v1/requirements
func (h *RequirementHandler) CreateRequirement(c *gin.Context) {
	var req service.CreateRequirementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
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

// CreateRequirementInUserStory handles POST /api/v1/user-stories/:id/requirements
func (h *RequirementHandler) CreateRequirementInUserStory(c *gin.Context) {
	userStoryIDParam := c.Param("id")
	
	// Try to parse as UUID first, then as reference ID
	var userStoryID uuid.UUID
	var err error
	
	if id, parseErr := uuid.Parse(userStoryIDParam); parseErr == nil {
		userStoryID = id
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user story ID format",
		})
		return
	}

	// Use a struct without required UserStoryID for this endpoint
	var reqBody struct {
		AcceptanceCriteriaID *uuid.UUID               `json:"acceptance_criteria_id,omitempty"`
		CreatorID            uuid.UUID                `json:"creator_id" binding:"required"`
		AssigneeID           *uuid.UUID               `json:"assignee_id,omitempty"`
		Priority             models.Priority          `json:"priority" binding:"required,min=1,max=4"`
		TypeID               uuid.UUID                `json:"type_id" binding:"required"`
		Title                string                   `json:"title" binding:"required,max=500"`
		Description          *string                  `json:"description,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Create the full request with user story ID from URL
	req := service.CreateRequirementRequest{
		UserStoryID:          userStoryID,
		AcceptanceCriteriaID: reqBody.AcceptanceCriteriaID,
		CreatorID:            reqBody.CreatorID,
		AssigneeID:           reqBody.AssigneeID,
		Priority:             reqBody.Priority,
		TypeID:               reqBody.TypeID,
		Title:                reqBody.Title,
		Description:          reqBody.Description,
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

	requirements, err := h.requirementService.ListRequirements(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list requirements",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"requirements": requirements,
		"count":        len(requirements),
	})
}

// GetRequirementWithRelationships handles GET /api/v1/requirements/:id/relationships
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
func (h *RequirementHandler) GetRelationshipsByRequirement(c *gin.Context) {
	idParam := c.Param("id")
	
	// Try to parse as UUID first, then as reference ID
	var requirementID uuid.UUID
	var err error
	
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

	relationships, err := h.requirementService.GetRelationshipsByRequirement(requirementID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get relationships",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"relationships": relationships,
		"count":         len(relationships),
	})
}

// SearchRequirements handles GET /api/v1/requirements/search
func (h *RequirementHandler) SearchRequirements(c *gin.Context) {
	searchText := c.Query("q")
	if searchText == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query parameter 'q' is required",
		})
		return
	}

	requirements, err := h.requirementService.SearchRequirements(searchText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to search requirements",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"requirements": requirements,
		"count":        len(requirements),
		"query":        searchText,
	})
}