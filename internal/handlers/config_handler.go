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
//
//	@Summary		Create a new requirement type
//	@Description	Creates a new requirement type for categorizing requirements. Requirement types help organize and classify different kinds of requirements (functional, non-functional, business rules, etc.).
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			requirement_type	body		service.CreateRequirementTypeRequest	true	"Requirement type creation request"
//	@Success		201					{object}	models.RequirementType					"Successfully created requirement type"
//	@Failure		400					{object}	ErrorResponse							"Invalid request body or validation error"
//	@Failure		409					{object}	ErrorResponse							"Requirement type name already exists"
//	@Failure		500					{object}	ErrorResponse							"Internal server error"
//	@Router			/api/v1/config/requirement-types [post]
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
//
//	@Summary		Get requirement type by ID
//	@Description	Retrieves a specific requirement type by its UUID. Returns complete requirement type information including name, description, and metadata.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Requirement type ID (UUID)"	example("123e4567-e89b-12d3-a456-426614174000")
//	@Success		200	{object}	models.RequirementType	"Successfully retrieved requirement type"
//	@Failure		400	{object}	ErrorResponse			"Invalid UUID format"
//	@Failure		404	{object}	ErrorResponse			"Requirement type not found"
//	@Failure		500	{object}	ErrorResponse			"Internal server error"
//	@Router			/api/v1/config/requirement-types/{id} [get]
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
//
//	@Summary		Update requirement type
//	@Description	Updates an existing requirement type. Only provided fields will be updated. Name must be unique across all requirement types.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string									true	"Requirement type ID (UUID)"	example("123e4567-e89b-12d3-a456-426614174000")
//	@Param			requirement_type	body		service.UpdateRequirementTypeRequest	true	"Requirement type update request"
//	@Success		200					{object}	models.RequirementType					"Successfully updated requirement type"
//	@Failure		400					{object}	ErrorResponse							"Invalid request body or UUID format"
//	@Failure		404					{object}	ErrorResponse							"Requirement type not found"
//	@Failure		409					{object}	ErrorResponse							"Requirement type name already exists"
//	@Failure		500					{object}	ErrorResponse							"Internal server error"
//	@Router			/api/v1/config/requirement-types/{id} [put]
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
//
//	@Summary		Delete requirement type
//	@Description	Deletes a requirement type. By default, deletion is prevented if there are requirements using this type. Use force=true to override this protection (use with caution).
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string	true	"Requirement type ID (UUID)"						example("123e4567-e89b-12d3-a456-426614174000")
//	@Param			force	query	boolean	false	"Force deletion even if requirements exist"			default(false)	example(false)
//	@Success		204		"Successfully deleted requirement type (no content)"
//	@Failure		400		{object}	ErrorResponse	"Invalid UUID format"
//	@Failure		404		{object}	ErrorResponse	"Requirement type not found"
//	@Failure		409		{object}	ErrorResponse	"Requirement type has associated requirements and cannot be deleted"
//	@Failure		500		{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/config/requirement-types/{id} [delete]
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
//
//	@Summary		List requirement types
//	@Description	Retrieves a paginated list of all requirement types with optional sorting. Supports ordering by name, created_at, or updated_at fields.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			order_by	query		string	false	"Sort field (name, created_at, updated_at)"	default(name)		example("name")
//	@Param			limit		query		int		false	"Maximum number of results (1-100)"		default(100)		example(50)
//	@Param			offset		query		int		false	"Number of results to skip"					default(0)			example(0)
//	@Success		200			{object}	RequirementTypeListResponse						"Successfully retrieved requirement types"
//	@Failure		500			{object}	ErrorResponse									"Internal server error"
//	@Router			/api/v1/config/requirement-types [get]
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
//
//	@Summary		Create a new relationship type
//	@Description	Creates a new relationship type for defining how requirements relate to each other. Common types include depends_on, blocks, relates_to, conflicts_with, and derives_from.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			relationship_type	body		service.CreateRelationshipTypeRequest	true	"Relationship type creation request"
//	@Success		201					{object}	models.RelationshipType					"Successfully created relationship type"
//	@Failure		400					{object}	ErrorResponse							"Invalid request body or validation error"
//	@Failure		409					{object}	ErrorResponse							"Relationship type name already exists"
//	@Failure		500					{object}	ErrorResponse							"Internal server error"
//	@Router			/api/v1/config/relationship-types [post]
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
//
//	@Summary		Get relationship type by ID
//	@Description	Retrieves a specific relationship type by its UUID. Returns complete relationship type information including name, description, and usage metadata.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string						true	"Relationship type ID (UUID)"	example("123e4567-e89b-12d3-a456-426614174000")
//	@Success		200	{object}	models.RelationshipType		"Successfully retrieved relationship type"
//	@Failure		400	{object}	ErrorResponse				"Invalid UUID format"
//	@Failure		404	{object}	ErrorResponse				"Relationship type not found"
//	@Failure		500	{object}	ErrorResponse				"Internal server error"
//	@Router			/api/v1/config/relationship-types/{id} [get]
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
//
//	@Summary		Update relationship type
//	@Description	Updates an existing relationship type. Only provided fields will be updated. Name must be unique across all relationship types.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string										true	"Relationship type ID (UUID)"	example("123e4567-e89b-12d3-a456-426614174000")
//	@Param			relationship_type	body		service.UpdateRelationshipTypeRequest	true	"Relationship type update request"
//	@Success		200					{object}	models.RelationshipType					"Successfully updated relationship type"
//	@Failure		400					{object}	ErrorResponse							"Invalid request body or UUID format"
//	@Failure		404					{object}	ErrorResponse							"Relationship type not found"
//	@Failure		409					{object}	ErrorResponse							"Relationship type name already exists"
//	@Failure		500					{object}	ErrorResponse							"Internal server error"
//	@Router			/api/v1/config/relationship-types/{id} [put]
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
//
//	@Summary		Delete relationship type
//	@Description	Deletes a relationship type. By default, deletion is prevented if there are requirement relationships using this type. Use force=true to override this protection (use with caution).
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string	true	"Relationship type ID (UUID)"							example("123e4567-e89b-12d3-a456-426614174000")
//	@Param			force	query	boolean	false	"Force deletion even if relationships exist"			default(false)	example(false)
//	@Success		204		"Successfully deleted relationship type (no content)"
//	@Failure		400		{object}	ErrorResponse	"Invalid UUID format"
//	@Failure		404		{object}	ErrorResponse	"Relationship type not found"
//	@Failure		409		{object}	ErrorResponse	"Relationship type has associated relationships and cannot be deleted"
//	@Failure		500		{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/config/relationship-types/{id} [delete]
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
//
//	@Summary		List relationship types
//	@Description	Retrieves a paginated list of all relationship types with optional sorting. Supports ordering by name, created_at, or updated_at fields.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			order_by	query		string	false	"Sort field (name, created_at, updated_at)"	default(name)		example("name")
//	@Param			limit		query		int		false	"Maximum number of results (1-100)"		default(100)		example(50)
//	@Param			offset		query		int		false	"Number of results to skip"					default(0)			example(0)
//	@Success		200			{object}	RelationshipTypeListResponse					"Successfully retrieved relationship types"
//	@Failure		500			{object}	ErrorResponse									"Internal server error"
//	@Router			/api/v1/config/relationship-types [get]
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
//
//	@Summary		Create a new status model
//	@Description	Creates a new status model for defining status workflows for different entity types (epic, user_story, requirement, acceptance_criteria). Each entity type can have multiple status models with one marked as default.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			status_model	body		service.CreateStatusModelRequest	true	"Status model creation request"
//	@Success		201				{object}	models.StatusModel					"Successfully created status model"
//	@Failure		400				{object}	ErrorResponse						"Invalid request body, validation error, or invalid entity type"
//	@Failure		409				{object}	ErrorResponse						"Status model name already exists for this entity type"
//	@Failure		500				{object}	ErrorResponse						"Internal server error"
//	@Router			/api/v1/config/status-models [post]
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
//
//	@Summary		Get status model by ID
//	@Description	Retrieves a specific status model by its UUID. Returns complete status model information including entity type, name, description, and default flag.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string				true	"Status model ID (UUID)"	example("123e4567-e89b-12d3-a456-426614174000")
//	@Success		200	{object}	models.StatusModel	"Successfully retrieved status model"
//	@Failure		400	{object}	ErrorResponse		"Invalid UUID format"
//	@Failure		404	{object}	ErrorResponse		"Status model not found"
//	@Failure		500	{object}	ErrorResponse		"Internal server error"
//	@Router			/api/v1/config/status-models/{id} [get]
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
//
//	@Summary		Update status model
//	@Description	Updates an existing status model. Only provided fields will be updated. Name must be unique within the same entity type. Setting is_default=true will make other models for the same entity type non-default.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string								true	"Status model ID (UUID)"	example("123e4567-e89b-12d3-a456-426614174000")
//	@Param			status_model	body		service.UpdateStatusModelRequest	true	"Status model update request"
//	@Success		200				{object}	models.StatusModel					"Successfully updated status model"
//	@Failure		400				{object}	ErrorResponse						"Invalid request body or UUID format"
//	@Failure		404				{object}	ErrorResponse						"Status model not found"
//	@Failure		409				{object}	ErrorResponse						"Status model name already exists for this entity type"
//	@Failure		500				{object}	ErrorResponse						"Internal server error"
//	@Router			/api/v1/config/status-models/{id} [put]
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
//
//	@Summary		Delete status model
//	@Description	Deletes a status model and all its associated statuses and transitions. Use with caution as this will affect entities using this status model. Consider setting a different default model before deletion.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string	true	"Status model ID (UUID)"	example("123e4567-e89b-12d3-a456-426614174000")
//	@Param			force	query	boolean	false	"Force deletion (reserved for future use)"	default(false)	example(false)
//	@Success		204		"Successfully deleted status model (no content)"
//	@Failure		400		{object}	ErrorResponse	"Invalid UUID format"
//	@Failure		404		{object}	ErrorResponse	"Status model not found"
//	@Failure		500		{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/config/status-models/{id} [delete]
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
//
//	@Summary		List status models
//	@Description	Retrieves a paginated list of status models with optional filtering by entity type and sorting. Supports ordering by entity_type, name, created_at, or updated_at fields.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			entity_type	query		string	false	"Filter by entity type (epic, user_story, requirement, acceptance_criteria)"	example("epic")
//	@Param			order_by	query		string	false	"Sort field (entity_type, name, created_at, updated_at)"						default("entity_type, name")	example("entity_type")
//	@Param			limit		query		int		false	"Maximum number of results (1-100)"											default(100)					example(50)
//	@Param			offset		query		int		false	"Number of results to skip"														default(0)						example(0)
//	@Success		200			{object}	StatusModelListResponse													"Successfully retrieved status models"
//	@Failure		500			{object}	ErrorResponse															"Internal server error"
//	@Router			/api/v1/config/status-models [get]
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
//
//	@Summary		Get default status model for entity type
//	@Description	Retrieves the default status model for a specific entity type. The default status model is used when creating new entities of that type.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			entity_type	path		string				true	"Entity type (epic, user_story, requirement, acceptance_criteria)"	example("epic")
//	@Success		200			{object}	models.StatusModel	"Successfully retrieved default status model"
//	@Failure		404			{object}	ErrorResponse		"Default status model not found for entity type"
//	@Failure		500			{object}	ErrorResponse		"Internal server error"
//	@Router			/api/v1/config/status-models/default/{entity_type} [get]
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
//
//	@Summary		Create a new status
//	@Description	Creates a new status within a status model. Statuses define the possible states for entities. Each status can be marked as initial (starting state) or final (ending state) and has an order for display purposes.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			status	body		service.CreateStatusRequest	true	"Status creation request"
//	@Success		201		{object}	models.Status				"Successfully created status"
//	@Failure		400		{object}	ErrorResponse				"Invalid request body, validation error, or status model not found"
//	@Failure		409		{object}	ErrorResponse				"Status name already exists in this model"
//	@Failure		500		{object}	ErrorResponse				"Internal server error"
//	@Router			/api/v1/config/statuses [post]
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
//
//	@Summary		Get status by ID
//	@Description	Retrieves a specific status by its UUID. Returns complete status information including name, description, color, flags, and order within the status model.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string			true	"Status ID (UUID)"	example("123e4567-e89b-12d3-a456-426614174000")
//	@Success		200	{object}	models.Status	"Successfully retrieved status"
//	@Failure		400	{object}	ErrorResponse	"Invalid UUID format"
//	@Failure		404	{object}	ErrorResponse	"Status not found"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/config/statuses/{id} [get]
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
//
//	@Summary		Update status
//	@Description	Updates an existing status. Only provided fields will be updated. Name must be unique within the same status model. Changing initial/final flags may affect status transitions.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Status ID (UUID)"	example("123e4567-e89b-12d3-a456-426614174000")
//	@Param			status	body		service.UpdateStatusRequest	true	"Status update request"
//	@Success		200		{object}	models.Status				"Successfully updated status"
//	@Failure		400		{object}	ErrorResponse				"Invalid request body or UUID format"
//	@Failure		404		{object}	ErrorResponse				"Status not found"
//	@Failure		409		{object}	ErrorResponse				"Status name already exists in this model"
//	@Failure		500		{object}	ErrorResponse				"Internal server error"
//	@Router			/api/v1/config/statuses/{id} [put]
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
//
//	@Summary		Delete status
//	@Description	Deletes a status from a status model. This will also remove any status transitions involving this status. Use with caution as entities using this status may be affected.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string	true	"Status ID (UUID)"							example("123e4567-e89b-12d3-a456-426614174000")
//	@Param			force	query	boolean	false	"Force deletion (reserved for future use)"	default(false)	example(false)
//	@Success		204		"Successfully deleted status (no content)"
//	@Failure		400		{object}	ErrorResponse	"Invalid UUID format"
//	@Failure		404		{object}	ErrorResponse	"Status not found"
//	@Failure		500		{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/config/statuses/{id} [delete]
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
//
//	@Summary		List statuses by status model
//	@Description	Retrieves all statuses belonging to a specific status model, ordered by their display order. Includes initial and final status flags for workflow understanding.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string				true	"Status model ID (UUID)"	example("123e4567-e89b-12d3-a456-426614174000")
//	@Success		200	{object}	StatusListResponse	"Successfully retrieved statuses"
//	@Failure		400	{object}	ErrorResponse		"Invalid UUID format"
//	@Failure		500	{object}	ErrorResponse		"Internal server error"
//	@Router			/api/v1/config/status-models/{id}/statuses [get]
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
//
//	@Summary		Create a new status transition
//	@Description	Creates a new status transition rule within a status model. Transitions define which status changes are allowed. Both from_status and to_status must belong to the same status model.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			transition	body		service.CreateStatusTransitionRequest	true	"Status transition creation request"
//	@Success		201			{object}	models.StatusTransition					"Successfully created status transition"
//	@Failure		400			{object}	ErrorResponse							"Invalid request body, validation error, status model not found, or invalid status transition"
//	@Failure		409			{object}	ErrorResponse							"Status transition already exists"
//	@Failure		500			{object}	ErrorResponse							"Internal server error"
//	@Router			/api/v1/config/status-transitions [post]
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
//
//	@Summary		Get status transition by ID
//	@Description	Retrieves a specific status transition by its UUID. Returns complete transition information including from/to statuses, name, and description.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Status transition ID (UUID)"	example("123e4567-e89b-12d3-a456-426614174000")
//	@Success		200	{object}	models.StatusTransition	"Successfully retrieved status transition"
//	@Failure		400	{object}	ErrorResponse			"Invalid UUID format"
//	@Failure		404	{object}	ErrorResponse			"Status transition not found"
//	@Failure		500	{object}	ErrorResponse			"Internal server error"
//	@Router			/api/v1/config/status-transitions/{id} [get]
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
//
//	@Summary		Update status transition
//	@Description	Updates an existing status transition. Only provided fields will be updated. The from_status and to_status cannot be changed after creation.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string									true	"Status transition ID (UUID)"	example("123e4567-e89b-12d3-a456-426614174000")
//	@Param			transition	body		service.UpdateStatusTransitionRequest	true	"Status transition update request"
//	@Success		200			{object}	models.StatusTransition					"Successfully updated status transition"
//	@Failure		400			{object}	ErrorResponse							"Invalid request body or UUID format"
//	@Failure		404			{object}	ErrorResponse							"Status transition not found"
//	@Failure		500			{object}	ErrorResponse							"Internal server error"
//	@Router			/api/v1/config/status-transitions/{id} [put]
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
//
//	@Summary		Delete status transition
//	@Description	Deletes a status transition rule. This will prevent the corresponding status change from being allowed in the future. Existing entities will not be affected.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"Status transition ID (UUID)"	example("123e4567-e89b-12d3-a456-426614174000")
//	@Success		204	"Successfully deleted status transition (no content)"
//	@Failure		400	{object}	ErrorResponse	"Invalid UUID format"
//	@Failure		404	{object}	ErrorResponse	"Status transition not found"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/config/status-transitions/{id} [delete]
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
//
//	@Summary		List status transitions by status model
//	@Description	Retrieves all status transitions belonging to a specific status model. Shows the complete workflow rules including allowed status changes and transition metadata.
//	@Tags			configuration
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string						true	"Status model ID (UUID)"	example("123e4567-e89b-12d3-a456-426614174000")
//	@Success		200	{object}	StatusTransitionListResponse	"Successfully retrieved status transitions"
//	@Failure		400	{object}	ErrorResponse					"Invalid UUID format"
//	@Failure		500	{object}	ErrorResponse					"Internal server error"
//	@Router			/api/v1/config/status-models/{id}/transitions [get]
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
