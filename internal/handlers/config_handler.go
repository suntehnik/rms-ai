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

// ConfigHandler handles HTTP requests for configuration operations
type ConfigHandler struct {
	configService service.ConfigService
}

// NewConfigHandler creates a new configuration handler instance
func NewConfigHandler(configService service.ConfigService) *ConfigHandler {
	return &ConfigHandler{
		configService: configService,
	}
}

// Requirement Type handlers

// CreateRequirementType handles POST /api/v1/config/requirement-types
func (h *ConfigHandler) CreateRequirementType(c *gin.Context) {
	var req service.CreateRequirementTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	requirementType, err := h.configService.CreateRequirementType(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRequirementTypeNameExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": "Requirement type name already exists",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create requirement type",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, requirementType)
}

// GetRequirementType handles GET /api/v1/config/requirement-types/:id
func (h *ConfigHandler) GetRequirementType(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid requirement type ID format",
		})
		return
	}

	requirementType, err := h.configService.GetRequirementTypeByID(id)
	if err != nil {
		if errors.Is(err, service.ErrRequirementTypeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Requirement type not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get requirement type",
			})
		}
		return
	}

	c.JSON(http.StatusOK, requirementType)
}

// UpdateRequirementType handles PUT /api/v1/config/requirement-types/:id
func (h *ConfigHandler) UpdateRequirementType(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid requirement type ID format",
		})
		return
	}

	var req service.UpdateRequirementTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	requirementType, err := h.configService.UpdateRequirementType(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRequirementTypeNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Requirement type not found",
			})
		case errors.Is(err, service.ErrRequirementTypeNameExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": "Requirement type name already exists",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update requirement type",
			})
		}
		return
	}

	c.JSON(http.StatusOK, requirementType)
}

// DeleteRequirementType handles DELETE /api/v1/config/requirement-types/:id
func (h *ConfigHandler) DeleteRequirementType(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid requirement type ID format",
		})
		return
	}

	// Check for force parameter
	force := c.Query("force") == "true"

	err = h.configService.DeleteRequirementType(id, force)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRequirementTypeNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Requirement type not found",
			})
		case errors.Is(err, service.ErrRequirementTypeHasRequirements):
			c.JSON(http.StatusConflict, gin.H{
				"error": "Requirement type has associated requirements and cannot be deleted",
				"hint":  "Remove all requirements using this type first",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete requirement type",
			})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListRequirementTypes handles GET /api/v1/config/requirement-types
func (h *ConfigHandler) ListRequirementTypes(c *gin.Context) {
	var filters service.RequirementTypeFilters

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

	requirementTypes, err := h.configService.ListRequirementTypes(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list requirement types",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"requirement_types": requirementTypes,
		"count":             len(requirementTypes),
	})
}

// Relationship Type handlers

// CreateRelationshipType handles POST /api/v1/config/relationship-types
func (h *ConfigHandler) CreateRelationshipType(c *gin.Context) {
	var req service.CreateRelationshipTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	relationshipType, err := h.configService.CreateRelationshipType(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRelationshipTypeNameExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": "Relationship type name already exists",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create relationship type",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, relationshipType)
}

// GetRelationshipType handles GET /api/v1/config/relationship-types/:id
func (h *ConfigHandler) GetRelationshipType(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid relationship type ID format",
		})
		return
	}

	relationshipType, err := h.configService.GetRelationshipTypeByID(id)
	if err != nil {
		if errors.Is(err, service.ErrRelationshipTypeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Relationship type not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get relationship type",
			})
		}
		return
	}

	c.JSON(http.StatusOK, relationshipType)
}

// UpdateRelationshipType handles PUT /api/v1/config/relationship-types/:id
func (h *ConfigHandler) UpdateRelationshipType(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid relationship type ID format",
		})
		return
	}

	var req service.UpdateRelationshipTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	relationshipType, err := h.configService.UpdateRelationshipType(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRelationshipTypeNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Relationship type not found",
			})
		case errors.Is(err, service.ErrRelationshipTypeNameExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": "Relationship type name already exists",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update relationship type",
			})
		}
		return
	}

	c.JSON(http.StatusOK, relationshipType)
}

// DeleteRelationshipType handles DELETE /api/v1/config/relationship-types/:id
func (h *ConfigHandler) DeleteRelationshipType(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid relationship type ID format",
		})
		return
	}

	// Check for force parameter
	force := c.Query("force") == "true"

	err = h.configService.DeleteRelationshipType(id, force)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRelationshipTypeNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Relationship type not found",
			})
		case errors.Is(err, service.ErrRelationshipTypeHasRelationships):
			c.JSON(http.StatusConflict, gin.H{
				"error": "Relationship type has associated relationships and cannot be deleted",
				"hint":  "Remove all relationships using this type first",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete relationship type",
			})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListRelationshipTypes handles GET /api/v1/config/relationship-types
func (h *ConfigHandler) ListRelationshipTypes(c *gin.Context) {
	var filters service.RelationshipTypeFilters

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

	relationshipTypes, err := h.configService.ListRelationshipTypes(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list relationship types",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"relationship_types": relationshipTypes,
		"count":              len(relationshipTypes),
	})
}

// Status Model handlers

// CreateStatusModel handles POST /api/v1/config/status-models
func (h *ConfigHandler) CreateStatusModel(c *gin.Context) {
	var req service.CreateStatusModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	statusModel, err := h.configService.CreateStatusModel(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrStatusModelNameExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": "Status model name already exists for this entity type",
			})
		case errors.Is(err, service.ErrInvalidEntityType):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid entity type",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create status model",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, statusModel)
}

// GetStatusModel handles GET /api/v1/config/status-models/:id
func (h *ConfigHandler) GetStatusModel(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status model ID format",
		})
		return
	}

	statusModel, err := h.configService.GetStatusModelByID(id)
	if err != nil {
		if errors.Is(err, service.ErrStatusModelNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Status model not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get status model",
			})
		}
		return
	}

	c.JSON(http.StatusOK, statusModel)
}

// UpdateStatusModel handles PUT /api/v1/config/status-models/:id
func (h *ConfigHandler) UpdateStatusModel(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status model ID format",
		})
		return
	}

	var req service.UpdateStatusModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	statusModel, err := h.configService.UpdateStatusModel(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrStatusModelNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Status model not found",
			})
		case errors.Is(err, service.ErrStatusModelNameExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": "Status model name already exists for this entity type",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update status model",
			})
		}
		return
	}

	c.JSON(http.StatusOK, statusModel)
}

// DeleteStatusModel handles DELETE /api/v1/config/status-models/:id
func (h *ConfigHandler) DeleteStatusModel(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status model ID format",
		})
		return
	}

	// Check for force parameter
	force := c.Query("force") == "true"

	err = h.configService.DeleteStatusModel(id, force)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrStatusModelNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Status model not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete status model",
			})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListStatusModels handles GET /api/v1/config/status-models
func (h *ConfigHandler) ListStatusModels(c *gin.Context) {
	var filters service.StatusModelFilters

	if entityType := c.Query("entity_type"); entityType != "" {
		filters.EntityType = models.EntityType(entityType)
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

	statusModels, err := h.configService.ListStatusModels(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list status models",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status_models": statusModels,
		"count":         len(statusModels),
	})
}

// GetDefaultStatusModel handles GET /api/v1/config/status-models/default/:entity_type
func (h *ConfigHandler) GetDefaultStatusModel(c *gin.Context) {
	entityTypeParam := c.Param("entity_type")

	statusModel, err := h.configService.GetDefaultStatusModelByEntityType(models.EntityType(entityTypeParam))
	if err != nil {
		if errors.Is(err, service.ErrStatusModelNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Default status model not found for entity type",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get default status model",
			})
		}
		return
	}

	c.JSON(http.StatusOK, statusModel)
}

// Status handlers

// CreateStatus handles POST /api/v1/config/statuses
func (h *ConfigHandler) CreateStatus(c *gin.Context) {
	var req service.CreateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	status, err := h.configService.CreateStatus(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrStatusModelNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Status model not found",
			})
		case errors.Is(err, service.ErrStatusNameExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": "Status name already exists in this model",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create status",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, status)
}

// GetStatus handles GET /api/v1/config/statuses/:id
func (h *ConfigHandler) GetStatus(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status ID format",
		})
		return
	}

	status, err := h.configService.GetStatusByID(id)
	if err != nil {
		if errors.Is(err, service.ErrStatusNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Status not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get status",
			})
		}
		return
	}

	c.JSON(http.StatusOK, status)
}

// UpdateStatus handles PUT /api/v1/config/statuses/:id
func (h *ConfigHandler) UpdateStatus(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status ID format",
		})
		return
	}

	var req service.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	status, err := h.configService.UpdateStatus(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrStatusNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Status not found",
			})
		case errors.Is(err, service.ErrStatusNameExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": "Status name already exists in this model",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update status",
			})
		}
		return
	}

	c.JSON(http.StatusOK, status)
}

// DeleteStatus handles DELETE /api/v1/config/statuses/:id
func (h *ConfigHandler) DeleteStatus(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status ID format",
		})
		return
	}

	// Check for force parameter
	force := c.Query("force") == "true"

	err = h.configService.DeleteStatus(id, force)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrStatusNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Status not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete status",
			})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListStatusesByModel handles GET /api/v1/config/status-models/:id/statuses
func (h *ConfigHandler) ListStatusesByModel(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status model ID format",
		})
		return
	}

	statuses, err := h.configService.ListStatusesByModel(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list statuses",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statuses": statuses,
		"count":    len(statuses),
	})
}

// Status Transition handlers

// CreateStatusTransition handles POST /api/v1/config/status-transitions
func (h *ConfigHandler) CreateStatusTransition(c *gin.Context) {
	var req service.CreateStatusTransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	transition, err := h.configService.CreateStatusTransition(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrStatusModelNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Status model not found",
			})
		case errors.Is(err, service.ErrStatusNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Status not found",
			})
		case errors.Is(err, service.ErrInvalidStatusTransition):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid status transition",
			})
		case errors.Is(err, service.ErrTransitionExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": "Status transition already exists",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create status transition",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, transition)
}

// GetStatusTransition handles GET /api/v1/config/status-transitions/:id
func (h *ConfigHandler) GetStatusTransition(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status transition ID format",
		})
		return
	}

	transition, err := h.configService.GetStatusTransitionByID(id)
	if err != nil {
		if errors.Is(err, service.ErrStatusTransitionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Status transition not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get status transition",
			})
		}
		return
	}

	c.JSON(http.StatusOK, transition)
}

// UpdateStatusTransition handles PUT /api/v1/config/status-transitions/:id
func (h *ConfigHandler) UpdateStatusTransition(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status transition ID format",
		})
		return
	}

	var req service.UpdateStatusTransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	transition, err := h.configService.UpdateStatusTransition(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrStatusTransitionNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Status transition not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update status transition",
			})
		}
		return
	}

	c.JSON(http.StatusOK, transition)
}

// DeleteStatusTransition handles DELETE /api/v1/config/status-transitions/:id
func (h *ConfigHandler) DeleteStatusTransition(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status transition ID format",
		})
		return
	}

	err = h.configService.DeleteStatusTransition(id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrStatusTransitionNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Status transition not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete status transition",
			})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListStatusTransitionsByModel handles GET /api/v1/config/status-models/:id/transitions
func (h *ConfigHandler) ListStatusTransitionsByModel(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status model ID format",
		})
		return
	}

	transitions, err := h.configService.ListStatusTransitionsByModel(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list status transitions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transitions": transitions,
		"count":       len(transitions),
	})
}
