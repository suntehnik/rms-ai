package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"product-requirements-management/internal/auth"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/service"
)

// SteeringDocumentHandler handles HTTP requests for steering document operations
type SteeringDocumentHandler struct {
	steeringDocumentService service.SteeringDocumentService
	epicService             service.EpicService
	userRepo                repository.UserRepository
}

// NewSteeringDocumentHandler creates a new steering document handler instance
func NewSteeringDocumentHandler(steeringDocumentService service.SteeringDocumentService, epicService service.EpicService, userRepo repository.UserRepository) *SteeringDocumentHandler {
	return &SteeringDocumentHandler{
		steeringDocumentService: steeringDocumentService,
		epicService:             epicService,
		userRepo:                userRepo,
	}
}

// getCurrentUser is a helper function to get the current user object from JWT claims
func (h *SteeringDocumentHandler) getCurrentUser(c *gin.Context) (*models.User, error) {
	userID, ok := auth.GetCurrentUserID(c)
	if !ok {
		return nil, errors.New("user authentication required")
	}

	user, err := h.userRepo.GetByID(uuid.MustParse(userID))
	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// CreateSteeringDocument handles POST /api/v1/steering-documents
// @Summary Create a new steering document
// @Description Create a new steering document with the provided details. The steering document will be assigned a unique reference ID (STD-XXX format). Requires authentication with JWT token. Users with User or Administrator role can create documents.
// @Tags steering-documents
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param steering_document body service.CreateSteeringDocumentRequest true "Steering document creation request"
// @Success 201 {object} models.SteeringDocument "Successfully created steering document"
// @Failure 400 {object} map[string]interface{} "Invalid request body or validation error"
// @Failure 401 {object} map[string]interface{} "Authentication required - missing or invalid JWT token"
// @Failure 403 {object} map[string]interface{} "User or Administrator role required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/steering-documents [post]
func (h *SteeringDocumentHandler) CreateSteeringDocument(c *gin.Context) {
	var req service.CreateSteeringDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	// Get current user
	currentUser, err := h.getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "AUTHENTICATION_REQUIRED",
				"message": "User authentication required",
			},
		})
		return
	}

	doc, err := h.steeringDocumentService.CreateSteeringDocument(req, currentUser)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUnauthorizedAccess):
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "INSUFFICIENT_PERMISSIONS",
					"message": "User or Administrator role required",
				},
			})
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "Creator not found",
				},
			})
		case errors.Is(err, service.ErrEpicNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "Epic not found",
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to create steering document",
				},
			})
		}
		return
	}

	c.JSON(http.StatusCreated, doc)
}

// GetSteeringDocument handles GET /api/v1/steering-documents/:id
// @Summary Get a steering document by ID or reference ID
// @Description Retrieve a single steering document by its UUID or reference ID (e.g., STD-001). Supports both formats for flexible access. Requires authentication with JWT token.
// @Tags steering-documents
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Steering document ID (UUID) or reference ID (STD-XXX)" example("123e4567-e89b-12d3-a456-426614174000")
// @Success 200 {object} models.SteeringDocument "Steering document found successfully"
// @Failure 401 {object} map[string]interface{} "Authentication required - missing or invalid JWT token"
// @Failure 404 {object} map[string]interface{} "Steering document not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/steering-documents/{id} [get]
func (h *SteeringDocumentHandler) GetSteeringDocument(c *gin.Context) {
	idParam := c.Param("id")

	// Get current user
	currentUser, err := h.getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "AUTHENTICATION_REQUIRED",
				"message": "User authentication required",
			},
		})
		return
	}

	// Try to parse as UUID first, then as reference ID
	var doc *models.SteeringDocument

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		doc, err = h.steeringDocumentService.GetSteeringDocumentByID(id, currentUser)
	} else {
		doc, err = h.steeringDocumentService.GetSteeringDocumentByReferenceID(idParam, currentUser)
	}

	if err != nil {
		if errors.Is(err, service.ErrSteeringDocumentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "Steering document not found",
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to get steering document",
				},
			})
		}
		return
	}

	c.JSON(http.StatusOK, doc)
}

// UpdateSteeringDocument handles PUT /api/v1/steering-documents/:id
// @Summary Update an existing steering document
// @Description Update a steering document's properties. Only provided fields will be updated. Requires authentication with JWT token. Administrators can update any document, Users can only update their own documents.
// @Tags steering-documents
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Steering document UUID or reference ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Param steering_document body service.UpdateSteeringDocumentRequest true "Steering document update request"
// @Success 200 {object} models.SteeringDocument "Steering document updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or steering document ID format"
// @Failure 401 {object} map[string]interface{} "Authentication required - missing or invalid JWT token"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions - can only update own documents"
// @Failure 404 {object} map[string]interface{} "Steering document not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/steering-documents/{id} [put]
func (h *SteeringDocumentHandler) UpdateSteeringDocument(c *gin.Context) {
	idParam := c.Param("id")

	// Get current user
	currentUser, err := h.getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "AUTHENTICATION_REQUIRED",
				"message": "User authentication required",
			},
		})
		return
	}

	// Parse ID (try UUID first, then reference ID)
	var id uuid.UUID
	if parsedID, parseErr := uuid.Parse(idParam); parseErr == nil {
		id = parsedID
	} else {
		// For reference ID, first get the document to get its UUID
		doc, err := h.steeringDocumentService.GetSteeringDocumentByReferenceID(idParam, currentUser)
		if err != nil {
			if errors.Is(err, service.ErrSteeringDocumentNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{
						"code":    "ENTITY_NOT_FOUND",
						"message": "Steering document not found",
					},
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "Failed to get steering document",
					},
				})
			}
			return
		}
		id = doc.ID
	}

	var req service.UpdateSteeringDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	doc, err := h.steeringDocumentService.UpdateSteeringDocument(id, req, currentUser)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSteeringDocumentNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "Steering document not found",
				},
			})
		case errors.Is(err, service.ErrUnauthorizedAccess):
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "INSUFFICIENT_PERMISSIONS",
					"message": "You can only update your own steering documents",
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to update steering document",
				},
			})
		}
		return
	}

	c.JSON(http.StatusOK, doc)
}

// DeleteSteeringDocument handles DELETE /api/v1/steering-documents/:id
// @Summary Delete a steering document
// @Description Delete a steering document by UUID or reference ID. Requires authentication with JWT token. Administrators can delete any document, Users can only delete their own documents.
// @Tags steering-documents
// @Accept json
// @Produce json
// @Security		BearerAuth
// @Param id path string true "Steering document UUID or reference ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Success 204 "Steering document deleted successfully"
// @Failure 401 {object} map[string]interface{} "Authentication required - missing or invalid JWT token"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions - can only delete own documents"
// @Failure 404 {object} map[string]interface{} "Steering document not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/steering-documents/{id} [delete]
func (h *SteeringDocumentHandler) DeleteSteeringDocument(c *gin.Context) {
	idParam := c.Param("id")

	// Get current user
	currentUser, err := h.getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "AUTHENTICATION_REQUIRED",
				"message": "User authentication required",
			},
		})
		return
	}

	// Parse ID (try UUID first, then reference ID)
	var id uuid.UUID
	if parsedID, parseErr := uuid.Parse(idParam); parseErr == nil {
		id = parsedID
	} else {
		// For reference ID, first get the document to get its UUID
		doc, err := h.steeringDocumentService.GetSteeringDocumentByReferenceID(idParam, currentUser)
		if err != nil {
			if errors.Is(err, service.ErrSteeringDocumentNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{
						"code":    "ENTITY_NOT_FOUND",
						"message": "Steering document not found",
					},
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "Failed to get steering document",
					},
				})
			}
			return
		}
		id = doc.ID
	}

	err = h.steeringDocumentService.DeleteSteeringDocument(id, currentUser)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSteeringDocumentNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "Steering document not found",
				},
			})
		case errors.Is(err, service.ErrUnauthorizedAccess):
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "INSUFFICIENT_PERMISSIONS",
					"message": "You can only delete your own steering documents",
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to delete steering document",
				},
			})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// ListSteeringDocuments handles GET /api/v1/steering-documents
// @Summary List steering documents with filtering and pagination
// @Description Retrieve a list of steering documents with optional filtering by creator and search query. Supports pagination and custom ordering. Requires authentication with JWT token.
// @Tags steering-documents
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param creator_id query string false "Filter by creator UUID" format(uuid) example("123e4567-e89b-12d3-a456-426614174001")
// @Param search query string false "Search query for full-text search in title and description" example("code review")
// @Param order_by query string false "Order results by field" example("created_at DESC")
// @Param limit query integer false "Maximum number of results to return" minimum(1) maximum(100) default(50) example(20)
// @Param offset query integer false "Number of results to skip for pagination" minimum(0) default(0) example(0)
// @Success 200 {object} map[string]interface{} "List of steering documents with count"
// @Failure 401 {object} map[string]interface{} "Authentication required - missing or invalid JWT token"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/steering-documents [get]
func (h *SteeringDocumentHandler) ListSteeringDocuments(c *gin.Context) {
	// Get current user
	currentUser, err := h.getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "AUTHENTICATION_REQUIRED",
				"message": "User authentication required",
			},
		})
		return
	}

	var filters service.SteeringDocumentFilters

	// Parse query parameters
	if creatorID := c.Query("creator_id"); creatorID != "" {
		if id, err := uuid.Parse(creatorID); err == nil {
			filters.CreatorID = &id
		}
	}

	filters.Search = c.Query("search")
	filters.OrderBy = c.Query("order_by")

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

	docs, totalCount, err := h.steeringDocumentService.ListSteeringDocuments(filters, currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to list steering documents",
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
		"data":        docs,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      filters.Offset,
	})
}

// LinkSteeringDocumentToEpic handles POST /api/v1/epics/:epic_id/steering-documents/:doc_id
// @Summary Link a steering document to an epic
// @Description Create a link between a steering document and an epic. Both entities must exist. Requires authentication with JWT token. Administrators can link any document, Users can only link their own documents.
// @Tags epics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param epic_id path string true "Epic UUID or reference ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Param doc_id path string true "Steering document UUID or reference ID" example("123e4567-e89b-12d3-a456-426614174001")
// @Success 201 {object} map[string]interface{} "Successfully linked steering document to epic"
// @Failure 400 {object} map[string]interface{} "Invalid epic or document ID format"
// @Failure 401 {object} map[string]interface{} "Authentication required - missing or invalid JWT token"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions - can only link own documents"
// @Failure 404 {object} map[string]interface{} "Epic or steering document not found"
// @Failure 409 {object} map[string]interface{} "Link already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/epics/{epic_id}/steering-documents/{doc_id} [post]
func (h *SteeringDocumentHandler) LinkSteeringDocumentToEpic(c *gin.Context) {
	epicIDParam := c.Param("id")
	docIDParam := c.Param("doc_id")

	// Get current user
	currentUser, err := h.getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "AUTHENTICATION_REQUIRED",
				"message": "User authentication required",
			},
		})
		return
	}

	// Parse epic_id (try UUID first, then reference ID)
	var epicID uuid.UUID
	if id, parseErr := uuid.Parse(epicIDParam); parseErr == nil {
		epicID = id
	} else {
		// For reference ID, first get the epic to get its UUID
		epic, err := h.epicService.GetEpicByReferenceID(epicIDParam)
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
		epicID = epic.ID
	}

	// Parse doc_id (try UUID first, then reference ID)
	var docID uuid.UUID
	if id, parseErr := uuid.Parse(docIDParam); parseErr == nil {
		docID = id
	} else {
		// For reference ID, first get the document to get its UUID
		doc, err := h.steeringDocumentService.GetSteeringDocumentByReferenceID(docIDParam, currentUser)
		if err != nil {
			if errors.Is(err, service.ErrSteeringDocumentNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{
						"code":    "ENTITY_NOT_FOUND",
						"message": "Steering document not found",
					},
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "Failed to get steering document",
					},
				})
			}
			return
		}
		docID = doc.ID
	}

	err = h.steeringDocumentService.LinkSteeringDocumentToEpic(docID, epicID, currentUser)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSteeringDocumentNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "Steering document not found",
				},
			})
		case errors.Is(err, service.ErrEpicNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "Epic not found",
				},
			})
		case errors.Is(err, service.ErrLinkAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{
					"code":    "CONFLICT",
					"message": "Link already exists",
				},
			})
		case errors.Is(err, service.ErrUnauthorizedAccess):
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "INSUFFICIENT_PERMISSIONS",
					"message": "You can only link your own steering documents",
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to link steering document to epic",
				},
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Successfully linked steering document to epic",
	})
}

// UnlinkSteeringDocumentFromEpic handles DELETE /api/v1/epics/:epic_id/steering-documents/:doc_id
// @Summary Unlink a steering document from an epic
// @Description Remove the link between a steering document and an epic. Requires authentication with JWT token. Administrators can unlink any document, Users can only unlink their own documents.
// @Tags epics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param epic_id path string true "Epic UUID or reference ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Param doc_id path string true "Steering document UUID or reference ID" example("123e4567-e89b-12d3-a456-426614174001")
// @Success 204 "Successfully unlinked steering document from epic"
// @Failure 400 {object} map[string]interface{} "Invalid epic or document ID format"
// @Failure 401 {object} map[string]interface{} "Authentication required - missing or invalid JWT token"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions - can only unlink own documents"
// @Failure 404 {object} map[string]interface{} "Epic or steering document not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/epics/{epic_id}/steering-documents/{doc_id} [delete]
func (h *SteeringDocumentHandler) UnlinkSteeringDocumentFromEpic(c *gin.Context) {
	epicIDParam := c.Param("id")
	docIDParam := c.Param("doc_id")

	// Get current user
	currentUser, err := h.getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "AUTHENTICATION_REQUIRED",
				"message": "User authentication required",
			},
		})
		return
	}

	// Parse epic_id (try UUID first, then reference ID)
	var epicID uuid.UUID
	if id, parseErr := uuid.Parse(epicIDParam); parseErr == nil {
		epicID = id
	} else {
		// For reference ID, first get the epic to get its UUID
		epic, err := h.epicService.GetEpicByReferenceID(epicIDParam)
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
		epicID = epic.ID
	}

	// Parse doc_id (try UUID first, then reference ID)
	var docID uuid.UUID
	if id, parseErr := uuid.Parse(docIDParam); parseErr == nil {
		docID = id
	} else {
		// For reference ID, first get the document to get its UUID
		doc, err := h.steeringDocumentService.GetSteeringDocumentByReferenceID(docIDParam, currentUser)
		if err != nil {
			if errors.Is(err, service.ErrSteeringDocumentNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{
						"code":    "ENTITY_NOT_FOUND",
						"message": "Steering document not found",
					},
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "Failed to get steering document",
					},
				})
			}
			return
		}
		docID = doc.ID
	}

	err = h.steeringDocumentService.UnlinkSteeringDocumentFromEpic(docID, epicID, currentUser)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSteeringDocumentNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "Steering document not found",
				},
			})
		case errors.Is(err, service.ErrEpicNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "Epic not found",
				},
			})
		case errors.Is(err, service.ErrUnauthorizedAccess):
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "INSUFFICIENT_PERMISSIONS",
					"message": "You can only unlink your own steering documents",
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to unlink steering document from epic",
				},
			})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// GetEpicSteeringDocuments handles GET /api/v1/epics/:id/steering-documents
// @Summary Get steering documents linked to an epic
// @Description Retrieve all steering documents that are linked to a specific epic. Returns an array of steering documents associated with the epic. Requires authentication with JWT token.
// @Tags epics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Epic UUID or reference ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Success 200 {array} models.SteeringDocument "Successfully retrieved steering documents for epic"
// @Failure 400 {object} map[string]interface{} "Invalid epic ID format"
// @Failure 401 {object} map[string]interface{} "Authentication required - missing or invalid JWT token"
// @Failure 404 {object} map[string]interface{} "Epic not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/epics/{id}/steering-documents [get]
func (h *SteeringDocumentHandler) GetEpicSteeringDocuments(c *gin.Context) {
	idParam := c.Param("id")

	// Get current user
	currentUser, err := h.getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "AUTHENTICATION_REQUIRED",
				"message": "User authentication required",
			},
		})
		return
	}

	// Parse epic_id (try UUID first, then reference ID)
	var epicID uuid.UUID
	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		epicID = id
	} else {
		// For reference ID, first get the epic to get its UUID
		epic, err := h.epicService.GetEpicByReferenceID(idParam)
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
		epicID = epic.ID
	}

	docs, err := h.steeringDocumentService.GetSteeringDocumentsByEpicID(epicID, currentUser)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEpicNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ENTITY_NOT_FOUND",
					"message": "Epic not found",
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to get steering documents for epic",
				},
			})
		}
		return
	}

	c.JSON(http.StatusOK, docs)
}
