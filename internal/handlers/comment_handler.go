package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// CommentHandler handles HTTP requests for comment operations
type CommentHandler struct {
	commentService service.CommentService
}

// NewCommentHandler creates a new comment handler instance
func NewCommentHandler(commentService service.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

// CreateComment handles POST /api/v1/:entityType/:id/comments
func (h *CommentHandler) CreateComment(c *gin.Context) {
	entityTypeParam := c.Param("entityType")
	entityIDParam := c.Param("id")

	// Parse entity type
	entityType := models.EntityType(entityTypeParam)

	// Parse entity ID
	entityID, err := uuid.Parse(entityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid entity ID format",
		})
		return
	}

	var req service.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Set entity type and ID from URL parameters
	req.EntityType = entityType
	req.EntityID = entityID

	comment, err := h.commentService.CreateComment(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentInvalidEntityType):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid entity type",
			})
		case errors.Is(err, service.ErrCommentEntityNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Entity not found",
			})
		case errors.Is(err, service.ErrCommentAuthorNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Author not found",
			})
		case errors.Is(err, service.ErrParentCommentNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Parent comment not found",
			})
		case errors.Is(err, service.ErrParentCommentWrongEntity):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Parent comment must be on the same entity",
			})
		case errors.Is(err, service.ErrEmptyContent):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Content cannot be empty",
			})
		case errors.Is(err, service.ErrInvalidInlineCommentData):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Inline comments require linked_text, text_position_start, and text_position_end",
			})
		case errors.Is(err, service.ErrInvalidTextPosition):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid text position: start must be >= 0 and end must be >= start",
			})
		case errors.Is(err, service.ErrEmptyLinkedText):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Linked text cannot be empty for inline comments",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create comment",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// GetCommentsByEntity handles GET /api/v1/:entityType/:id/comments
func (h *CommentHandler) GetCommentsByEntity(c *gin.Context) {
	entityTypeParam := c.Param("entityType")
	entityIDParam := c.Param("id")

	// Parse entity type
	entityType := models.EntityType(entityTypeParam)

	// Parse entity ID
	entityID, err := uuid.Parse(entityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid entity ID format",
		})
		return
	}

	// Check for threaded parameter
	threaded := c.Query("threaded") == "true"

	// Check for inline parameter
	inlineOnly := c.Query("inline") == "true"

	// Check for status filter
	statusFilter := c.Query("status")

	var comments []service.CommentResponse

	if inlineOnly {
		// Use visible inline comments to exclude hidden ones
		comments, err = h.commentService.GetVisibleInlineComments(entityType, entityID)
	} else if threaded {
		comments, err = h.commentService.GetThreadedComments(entityType, entityID)
	} else {
		comments, err = h.commentService.GetCommentsByEntity(entityType, entityID)
	}

	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentInvalidEntityType):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid entity type",
			})
		case errors.Is(err, service.ErrCommentEntityNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Entity not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get comments",
			})
		}
		return
	}

	// Apply status filter if specified
	if statusFilter != "" {
		filteredComments := make([]service.CommentResponse, 0)
		for _, comment := range comments {
			switch statusFilter {
			case "resolved":
				if comment.IsResolved {
					filteredComments = append(filteredComments, comment)
				}
			case "unresolved":
				if !comment.IsResolved {
					filteredComments = append(filteredComments, comment)
				}
			default:
				// Invalid status filter, return all comments
				filteredComments = comments
				break
			}
		}
		comments = filteredComments
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": comments,
		"count":    len(comments),
	})
}

// GetComment handles GET /api/v1/comments/:id
func (h *CommentHandler) GetComment(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid comment ID format",
		})
		return
	}

	comment, err := h.commentService.GetComment(id)
	if err != nil {
		if errors.Is(err, service.ErrCommentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Comment not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get comment",
			})
		}
		return
	}

	c.JSON(http.StatusOK, comment)
}

// UpdateComment handles PUT /api/v1/comments/:id
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid comment ID format",
		})
		return
	}

	var req service.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	comment, err := h.commentService.UpdateComment(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Comment not found",
			})
		case errors.Is(err, service.ErrEmptyContent):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Content cannot be empty",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update comment",
			})
		}
		return
	}

	c.JSON(http.StatusOK, comment)
}

// DeleteComment handles DELETE /api/v1/comments/:id
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid comment ID format",
		})
		return
	}

	err = h.commentService.DeleteComment(id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Comment not found",
			})
		case errors.Is(err, service.ErrCommentHasReplies):
			c.JSON(http.StatusConflict, gin.H{
				"error": "Comment has replies and cannot be deleted",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete comment",
			})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ResolveComment handles POST /api/v1/comments/:id/resolve
func (h *CommentHandler) ResolveComment(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid comment ID format",
		})
		return
	}

	comment, err := h.commentService.ResolveComment(id)
	if err != nil {
		if errors.Is(err, service.ErrCommentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Comment not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to resolve comment",
			})
		}
		return
	}

	c.JSON(http.StatusOK, comment)
}

// UnresolveComment handles POST /api/v1/comments/:id/unresolve
func (h *CommentHandler) UnresolveComment(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid comment ID format",
		})
		return
	}

	comment, err := h.commentService.UnresolveComment(id)
	if err != nil {
		if errors.Is(err, service.ErrCommentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Comment not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to unresolve comment",
			})
		}
		return
	}

	c.JSON(http.StatusOK, comment)
}

// GetCommentsByStatus handles GET /api/v1/comments/status/:status
func (h *CommentHandler) GetCommentsByStatus(c *gin.Context) {
	statusParam := c.Param("status")

	var isResolved bool
	switch statusParam {
	case "resolved":
		isResolved = true
	case "unresolved":
		isResolved = false
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status. Use 'resolved' or 'unresolved'",
		})
		return
	}

	comments, err := h.commentService.GetCommentsByStatus(isResolved)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get comments by status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": comments,
		"count":    len(comments),
		"status":   statusParam,
	})
}

// GetCommentReplies handles GET /api/v1/comments/:id/replies
func (h *CommentHandler) GetCommentReplies(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid comment ID format",
		})
		return
	}

	// First verify the parent comment exists
	_, err = h.commentService.GetComment(id)
	if err != nil {
		if errors.Is(err, service.ErrCommentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Comment not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get comment",
			})
		}
		return
	}

	// Get replies through the comment repository
	// Note: This would need to be implemented in the service layer
	// For now, we'll return an empty array as a placeholder
	c.JSON(http.StatusOK, gin.H{
		"replies": []service.CommentResponse{},
		"count":   0,
	})
}

// CreateCommentReply handles POST /api/v1/comments/:id/replies
func (h *CommentHandler) CreateCommentReply(c *gin.Context) {
	parentIDParam := c.Param("id")

	parentID, err := uuid.Parse(parentIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid parent comment ID format",
		})
		return
	}

	// Get parent comment to extract entity type and ID
	parentComment, err := h.commentService.GetComment(parentID)
	if err != nil {
		if errors.Is(err, service.ErrCommentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Parent comment not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get parent comment",
			})
		}
		return
	}

	var req service.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Set entity information from parent comment
	req.EntityType = parentComment.EntityType
	req.EntityID = parentComment.EntityID
	req.ParentCommentID = &parentID

	comment, err := h.commentService.CreateComment(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentAuthorNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Author not found",
			})
		case errors.Is(err, service.ErrEmptyContent):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Content cannot be empty",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create reply",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// CreateInlineComment handles POST /api/v1/:entityType/:id/comments/inline
func (h *CommentHandler) CreateInlineComment(c *gin.Context) {
	entityTypeParam := c.Param("entityType")
	entityIDParam := c.Param("id")

	// Parse entity type
	entityType := models.EntityType(entityTypeParam)

	// Parse entity ID
	entityID, err := uuid.Parse(entityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid entity ID format",
		})
		return
	}

	var req service.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Set entity type and ID from URL parameters
	req.EntityType = entityType
	req.EntityID = entityID

	// Validate that this is an inline comment request
	if req.LinkedText == nil || req.TextPositionStart == nil || req.TextPositionEnd == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Inline comments require linked_text, text_position_start, and text_position_end",
		})
		return
	}

	comment, err := h.commentService.CreateComment(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentInvalidEntityType):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid entity type",
			})
		case errors.Is(err, service.ErrCommentEntityNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Entity not found",
			})
		case errors.Is(err, service.ErrCommentAuthorNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Author not found",
			})
		case errors.Is(err, service.ErrEmptyContent):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Content cannot be empty",
			})
		case errors.Is(err, service.ErrInvalidInlineCommentData):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Inline comments require linked_text, text_position_start, and text_position_end",
			})
		case errors.Is(err, service.ErrInvalidTextPosition):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid text position: start must be >= 0 and end must be >= start",
			})
		case errors.Is(err, service.ErrEmptyLinkedText):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Linked text cannot be empty for inline comments",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create inline comment",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// CreateEpicInlineComment handles POST /api/v1/epics/:id/comments/inline
func (h *CommentHandler) CreateEpicInlineComment(c *gin.Context) {
	h.createInlineCommentForEntity(c, models.EntityTypeEpic)
}

// CreateUserStoryInlineComment handles POST /api/v1/user-stories/:id/comments/inline
func (h *CommentHandler) CreateUserStoryInlineComment(c *gin.Context) {
	h.createInlineCommentForEntity(c, models.EntityTypeUserStory)
}

// CreateAcceptanceCriteriaInlineComment handles POST /api/v1/acceptance-criteria/:id/comments/inline
func (h *CommentHandler) CreateAcceptanceCriteriaInlineComment(c *gin.Context) {
	h.createInlineCommentForEntity(c, models.EntityTypeAcceptanceCriteria)
}

// CreateRequirementInlineComment handles POST /api/v1/requirements/:id/comments/inline
func (h *CommentHandler) CreateRequirementInlineComment(c *gin.Context) {
	h.createInlineCommentForEntity(c, models.EntityTypeRequirement)
}

// createInlineCommentForEntity is a helper function for entity-specific inline comment creation
func (h *CommentHandler) createInlineCommentForEntity(c *gin.Context, entityType models.EntityType) {
	entityIDParam := c.Param("id")

	// Parse entity ID
	entityID, err := uuid.Parse(entityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid entity ID format",
		})
		return
	}

	var req service.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Set entity type and ID
	req.EntityType = entityType
	req.EntityID = entityID

	// Validate that this is an inline comment request
	if req.LinkedText == nil || req.TextPositionStart == nil || req.TextPositionEnd == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Inline comments require linked_text, text_position_start, and text_position_end",
		})
		return
	}

	comment, err := h.commentService.CreateComment(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentInvalidEntityType):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid entity type",
			})
		case errors.Is(err, service.ErrCommentEntityNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Entity not found",
			})
		case errors.Is(err, service.ErrCommentAuthorNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Author not found",
			})
		case errors.Is(err, service.ErrEmptyContent):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Content cannot be empty",
			})
		case errors.Is(err, service.ErrInvalidInlineCommentData):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Inline comments require linked_text, text_position_start, and text_position_end",
			})
		case errors.Is(err, service.ErrInvalidTextPosition):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid text position: start must be >= 0 and end must be >= start",
			})
		case errors.Is(err, service.ErrEmptyLinkedText):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Linked text cannot be empty for inline comments",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create inline comment",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// GetVisibleInlineComments handles GET /api/v1/:entityType/:id/comments/inline/visible
func (h *CommentHandler) GetVisibleInlineComments(c *gin.Context) {
	entityTypeParam := c.Param("entityType")
	entityIDParam := c.Param("id")

	// Parse entity type
	entityType := models.EntityType(entityTypeParam)

	// Parse entity ID
	entityID, err := uuid.Parse(entityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid entity ID format",
		})
		return
	}

	comments, err := h.commentService.GetVisibleInlineComments(entityType, entityID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentInvalidEntityType):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid entity type",
			})
		case errors.Is(err, service.ErrCommentEntityNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Entity not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get visible inline comments",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": comments,
		"count":    len(comments),
	})
}

// GetEpicVisibleInlineComments handles GET /api/v1/epics/:id/comments/inline/visible
func (h *CommentHandler) GetEpicVisibleInlineComments(c *gin.Context) {
	h.getVisibleInlineCommentsForEntity(c, models.EntityTypeEpic)
}

// GetUserStoryVisibleInlineComments handles GET /api/v1/user-stories/:id/comments/inline/visible
func (h *CommentHandler) GetUserStoryVisibleInlineComments(c *gin.Context) {
	h.getVisibleInlineCommentsForEntity(c, models.EntityTypeUserStory)
}

// GetAcceptanceCriteriaVisibleInlineComments handles GET /api/v1/acceptance-criteria/:id/comments/inline/visible
func (h *CommentHandler) GetAcceptanceCriteriaVisibleInlineComments(c *gin.Context) {
	h.getVisibleInlineCommentsForEntity(c, models.EntityTypeAcceptanceCriteria)
}

// GetRequirementVisibleInlineComments handles GET /api/v1/requirements/:id/comments/inline/visible
func (h *CommentHandler) GetRequirementVisibleInlineComments(c *gin.Context) {
	h.getVisibleInlineCommentsForEntity(c, models.EntityTypeRequirement)
}

// getVisibleInlineCommentsForEntity is a helper function for entity-specific visible inline comments retrieval
func (h *CommentHandler) getVisibleInlineCommentsForEntity(c *gin.Context, entityType models.EntityType) {
	entityIDParam := c.Param("id")

	// Parse entity ID
	entityID, err := uuid.Parse(entityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid entity ID format",
		})
		return
	}

	comments, err := h.commentService.GetVisibleInlineComments(entityType, entityID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentInvalidEntityType):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid entity type",
			})
		case errors.Is(err, service.ErrCommentEntityNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Entity not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get visible inline comments",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": comments,
		"count":    len(comments),
	})
}

// ValidateInlineComments handles POST /api/v1/:entityType/:id/comments/inline/validate
func (h *CommentHandler) ValidateInlineComments(c *gin.Context) {
	entityTypeParam := c.Param("entityType")
	entityIDParam := c.Param("id")

	// Parse entity type
	entityType := models.EntityType(entityTypeParam)

	// Parse entity ID
	entityID, err := uuid.Parse(entityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid entity ID format",
		})
		return
	}

	var req struct {
		NewDescription string `json:"new_description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	err = h.commentService.ValidateInlineCommentsAfterTextChange(entityType, entityID, req.NewDescription)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to validate inline comments",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Inline comments validated successfully",
	})
}

// ValidateEpicInlineComments handles POST /api/v1/epics/:id/comments/inline/validate
func (h *CommentHandler) ValidateEpicInlineComments(c *gin.Context) {
	h.validateInlineCommentsForEntity(c, models.EntityTypeEpic)
}

// ValidateUserStoryInlineComments handles POST /api/v1/user-stories/:id/comments/inline/validate
func (h *CommentHandler) ValidateUserStoryInlineComments(c *gin.Context) {
	h.validateInlineCommentsForEntity(c, models.EntityTypeUserStory)
}

// ValidateAcceptanceCriteriaInlineComments handles POST /api/v1/acceptance-criteria/:id/comments/inline/validate
func (h *CommentHandler) ValidateAcceptanceCriteriaInlineComments(c *gin.Context) {
	h.validateInlineCommentsForEntity(c, models.EntityTypeAcceptanceCriteria)
}

// ValidateRequirementInlineComments handles POST /api/v1/requirements/:id/comments/inline/validate
func (h *CommentHandler) ValidateRequirementInlineComments(c *gin.Context) {
	h.validateInlineCommentsForEntity(c, models.EntityTypeRequirement)
}

// validateInlineCommentsForEntity is a helper function for entity-specific inline comment validation
func (h *CommentHandler) validateInlineCommentsForEntity(c *gin.Context, entityType models.EntityType) {
	entityIDParam := c.Param("id")

	// Parse entity ID
	entityID, err := uuid.Parse(entityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid entity ID format",
		})
		return
	}

	var req struct {
		NewDescription string `json:"new_description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	err = h.commentService.ValidateInlineCommentsAfterTextChange(entityType, entityID, req.NewDescription)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to validate inline comments",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Inline comments validated successfully",
	})
}

// Entity-specific comment handlers that determine entity type from route context

// CreateEpicComment handles POST /api/v1/epics/:id/comments
func (h *CommentHandler) CreateEpicComment(c *gin.Context) {
	h.createCommentForEntity(c, models.EntityTypeEpic)
}

// CreateUserStoryComment handles POST /api/v1/user-stories/:id/comments
func (h *CommentHandler) CreateUserStoryComment(c *gin.Context) {
	h.createCommentForEntity(c, models.EntityTypeUserStory)
}

// CreateAcceptanceCriteriaComment handles POST /api/v1/acceptance-criteria/:id/comments
func (h *CommentHandler) CreateAcceptanceCriteriaComment(c *gin.Context) {
	h.createCommentForEntity(c, models.EntityTypeAcceptanceCriteria)
}

// CreateRequirementComment handles POST /api/v1/requirements/:id/comments
func (h *CommentHandler) CreateRequirementComment(c *gin.Context) {
	h.createCommentForEntity(c, models.EntityTypeRequirement)
}

// createCommentForEntity is a helper function for entity-specific comment creation
func (h *CommentHandler) createCommentForEntity(c *gin.Context, entityType models.EntityType) {
	entityIDParam := c.Param("id")

	// Parse entity ID
	entityID, err := uuid.Parse(entityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid entity ID format",
		})
		return
	}

	var req service.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Set entity type and ID
	req.EntityType = entityType
	req.EntityID = entityID

	comment, err := h.commentService.CreateComment(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentInvalidEntityType):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid entity type",
			})
		case errors.Is(err, service.ErrCommentEntityNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Entity not found",
			})
		case errors.Is(err, service.ErrCommentAuthorNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Author not found",
			})
		case errors.Is(err, service.ErrParentCommentNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Parent comment not found",
			})
		case errors.Is(err, service.ErrParentCommentWrongEntity):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Parent comment must be on the same entity",
			})
		case errors.Is(err, service.ErrEmptyContent):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Content cannot be empty",
			})
		case errors.Is(err, service.ErrInvalidInlineCommentData):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Inline comments require linked_text, text_position_start, and text_position_end",
			})
		case errors.Is(err, service.ErrInvalidTextPosition):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid text position: start must be >= 0 and end must be >= start",
			})
		case errors.Is(err, service.ErrEmptyLinkedText):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Linked text cannot be empty for inline comments",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create comment",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// GetEpicComments handles GET /api/v1/epics/:id/comments
func (h *CommentHandler) GetEpicComments(c *gin.Context) {
	h.getCommentsForEntity(c, models.EntityTypeEpic)
}

// GetUserStoryComments handles GET /api/v1/user-stories/:id/comments
func (h *CommentHandler) GetUserStoryComments(c *gin.Context) {
	h.getCommentsForEntity(c, models.EntityTypeUserStory)
}

// GetAcceptanceCriteriaComments handles GET /api/v1/acceptance-criteria/:id/comments
func (h *CommentHandler) GetAcceptanceCriteriaComments(c *gin.Context) {
	h.getCommentsForEntity(c, models.EntityTypeAcceptanceCriteria)
}

// GetRequirementComments handles GET /api/v1/requirements/:id/comments
func (h *CommentHandler) GetRequirementComments(c *gin.Context) {
	h.getCommentsForEntity(c, models.EntityTypeRequirement)
}

// getCommentsForEntity is a helper function for entity-specific comment retrieval
func (h *CommentHandler) getCommentsForEntity(c *gin.Context, entityType models.EntityType) {
	entityIDParam := c.Param("id")

	// Parse entity ID
	entityID, err := uuid.Parse(entityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid entity ID format",
		})
		return
	}

	// Check for threaded parameter
	threaded := c.Query("threaded") == "true"

	// Check for inline parameter
	inlineOnly := c.Query("inline") == "true"

	// Check for status filter
	statusFilter := c.Query("status")

	var comments []service.CommentResponse

	if inlineOnly {
		// Use visible inline comments to exclude hidden ones
		comments, err = h.commentService.GetVisibleInlineComments(entityType, entityID)
	} else if threaded {
		comments, err = h.commentService.GetThreadedComments(entityType, entityID)
	} else {
		comments, err = h.commentService.GetCommentsByEntity(entityType, entityID)
	}

	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentInvalidEntityType):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid entity type",
			})
		case errors.Is(err, service.ErrCommentEntityNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Entity not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get comments",
			})
		}
		return
	}

	// Apply status filter if specified
	if statusFilter != "" {
		filteredComments := make([]service.CommentResponse, 0)
		for _, comment := range comments {
			switch statusFilter {
			case "resolved":
				if comment.IsResolved {
					filteredComments = append(filteredComments, comment)
				}
			case "unresolved":
				if !comment.IsResolved {
					filteredComments = append(filteredComments, comment)
				}
			default:
				// Invalid status filter, return all comments
				filteredComments = comments
				break
			}
		}
		comments = filteredComments
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": comments,
		"count":    len(comments),
	})
}
