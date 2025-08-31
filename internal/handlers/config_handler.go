package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

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
		"count":            len(requirementTypes),
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
		"count":             len(relationshipTypes),
	})
}