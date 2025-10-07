package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"product-requirements-management/internal/auth"
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
// @Summary Create a new comment on an entity
// @Description Create a new comment (general or inline) on any entity type (epic, user_story, acceptance_criteria, requirement). Supports threaded discussions through parent_comment_id.
// @Tags comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param entityType path string true "Entity type" Enums(epic,user_story,acceptance_criteria,requirement)
// @Param id path string true "Entity ID" format(uuid)
// @Param comment body service.CreateCommentRequest true "Comment creation request"
// @Success 201 {object} service.CommentResponse "Successfully created comment"
// @Failure 400 {object} map[string]string "Invalid request - malformed entity ID, invalid entity type, missing required fields, or invalid inline comment data"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Entity not found or parent comment not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/{entityType}/{id}/comments [post]
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
// @Summary Get all comments for an entity
// @Description Retrieve all comments for a specific entity with optional filtering by status and threading. Supports both flat and threaded comment structures.
// @Tags comments
// @Produce json
// @Security BearerAuth
// @Param entityType path string true "Entity type" Enums(epic,user_story,acceptance_criteria,requirement)
// @Param id path string true "Entity ID" format(uuid)
// @Param threaded query boolean false "Return comments in threaded structure"
// @Param inline query boolean false "Return only inline comments"
// @Param status query string false "Filter by resolution status" Enums(resolved,unresolved)
// @Success 200 {object} map[string]interface{} "Successfully retrieved comments" example({"comments": [{"id": "123e4567-e89b-12d3-a456-426614174000", "content": "This needs clarification", "is_resolved": false}], "count": 1})
// @Failure 400 {object} map[string]string "Invalid entity type or malformed entity ID"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Entity not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/{entityType}/{id}/comments [get]
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
// @Summary Get a specific comment by ID
// @Description Retrieve a single comment by its unique identifier, including author information and thread context.
// @Tags comments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Comment ID" format(uuid)
// @Success 200 {object} service.CommentResponse "Successfully retrieved comment"
// @Failure 400 {object} map[string]string "Invalid comment ID format"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Comment not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/comments/{id} [get]
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
// @Summary Update an existing comment
// @Description Update the content of an existing comment. Only the comment content can be modified after creation.
// @Tags comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Comment ID" format(uuid)
// @Param comment body service.UpdateCommentRequest true "Comment update request"
// @Success 200 {object} service.CommentResponse "Successfully updated comment"
// @Failure 400 {object} map[string]string "Invalid comment ID format, invalid request body, or empty content"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Comment not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/comments/{id} [put]
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
// @Summary Delete a comment
// @Description Delete a comment by ID. Comments with replies cannot be deleted to maintain thread integrity.
// @Tags comments
// @Security BearerAuth
// @Param id path string true "Comment ID" format(uuid)
// @Success 204 "Successfully deleted comment"
// @Failure 400 {object} map[string]string "Invalid comment ID format"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Comment not found"
// @Failure 409 {object} map[string]string "Comment has replies and cannot be deleted"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/comments/{id} [delete]
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
// @Summary Mark a comment as resolved
// @Description Mark a comment as resolved to indicate that the issue or question has been addressed.
// @Tags comments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Comment ID" format(uuid)
// @Success 200 {object} service.CommentResponse "Successfully resolved comment"
// @Failure 400 {object} map[string]string "Invalid comment ID format"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Comment not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/comments/{id}/resolve [post]
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
// @Summary Mark a comment as unresolved
// @Description Mark a previously resolved comment as unresolved to reopen the discussion or issue.
// @Tags comments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Comment ID" format(uuid)
// @Success 200 {object} service.CommentResponse "Successfully unresolved comment"
// @Failure 400 {object} map[string]string "Invalid comment ID format"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Comment not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/comments/{id}/unresolve [post]
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
// @Summary Get comments by resolution status
// @Description Retrieve all comments filtered by their resolution status (resolved or unresolved) across all entities.
// @Tags comments
// @Produce json
// @Security BearerAuth
// @Param status path string true "Resolution status" Enums(resolved,unresolved)
// @Success 200 {object} map[string]interface{} "Successfully retrieved comments by status" example({"comments": [{"id": "123e4567-e89b-12d3-a456-426614174000", "content": "This needs clarification", "is_resolved": false}], "count": 1, "status": "unresolved"})
// @Failure 400 {object} map[string]string "Invalid status parameter"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/comments/status/{status} [get]
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
// @Summary Get replies to a specific comment
// @Description Retrieve all direct replies to a specific comment with pagination support. Returns replies in chronological order (oldest first) to maintain conversation flow. Each reply includes author information and metadata for building threaded comment interfaces.
// @Tags comments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Parent comment ID" format(uuid)
// @Param limit query int false "Maximum number of replies to return (1-100)" minimum(1) maximum(100) default(50)
// @Param offset query int false "Number of replies to skip for pagination" minimum(0) default(0)
// @Success 200 {object} map[string]interface{} "Successfully retrieved comment replies" example({"data": [{"id": "123e4567-e89b-12d3-a456-426614174001", "content": "I agree with this point", "parent_comment_id": "123e4567-e89b-12d3-a456-426614174000", "author_id": "456e7890-e89b-12d3-a456-426614174002", "created_at": "2024-01-15T10:30:00Z", "is_resolved": false, "is_reply": true, "depth": 1}], "total_count": 1, "limit": 50, "offset": 0})
// @Failure 400 {object} map[string]string "Invalid comment ID format"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Parent comment not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/comments/{id}/replies [get]
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

	// Get replies through the comment service
	replies, err := h.commentService.GetCommentReplies(id)
	if err != nil {
		if errors.Is(err, service.ErrCommentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Comment not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get comment replies",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        replies,
		"total_count": len(replies),
		"limit":       len(replies), // For now, return all replies without pagination
		"offset":      0,
	})
}

// CreateCommentReply handles POST /api/v1/comments/:id/replies
// @Summary Create a reply to a comment
// @Description Create a new reply to an existing comment, automatically inheriting the parent's entity context for threaded discussions. The reply will be linked to the same entity (epic, user story, etc.) as the parent comment and establish a parent-child relationship for proper threading.
// @Tags comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Parent comment ID" format(uuid)
// @Param reply body service.CreateCommentRequest true "Reply creation request (only content and author_id required - entity context inherited from parent)"
// @Success 201 {object} service.CommentResponse "Successfully created reply with parent-child relationship established"
// @Failure 400 {object} map[string]string "Invalid parent comment ID format, invalid request body, empty content, or author not found"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Parent comment not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/comments/{id}/replies [post]
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
// @Summary Create an inline comment on specific text
// @Description Create an inline comment linked to specific text positions within an entity's content. Requires linked_text, text_position_start, and text_position_end fields.
// @Tags comments,inline-comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param entityType path string true "Entity type" Enums(epic,user_story,acceptance_criteria,requirement)
// @Param id path string true "Entity ID" format(uuid)
// @Param comment body service.CreateCommentRequest true "Inline comment creation request with text position data"
// @Success 201 {object} service.CommentResponse "Successfully created inline comment"
// @Failure 400 {object} map[string]string "Invalid request - missing inline comment data, invalid text positions, or empty linked text"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Entity not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/{entityType}/{id}/comments/inline [post]
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
// @Summary Create an inline comment on an epic's text
// @Description Create an inline comment linked to specific text positions within an epic's description or title.
// @Tags epics,comments,inline-comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Epic ID" format(uuid)
// @Param comment body service.CreateCommentRequest true "Inline comment creation request with text position data"
// @Success 201 {object} service.CommentResponse "Successfully created epic inline comment"
// @Failure 400 {object} map[string]string "Invalid request - missing inline comment data or invalid text positions"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Epic not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/epics/{id}/comments/inline [post]
func (h *CommentHandler) CreateEpicInlineComment(c *gin.Context) {
	h.createInlineCommentForEntity(c, models.EntityTypeEpic)
}

// CreateUserStoryInlineComment handles POST /api/v1/user-stories/:id/comments/inline
// @Summary Create an inline comment on a user story's text
// @Description Create an inline comment linked to specific text positions within a user story's description or acceptance criteria. Requires authentication.
// @Tags user-stories,comments,inline-comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User Story ID" format(uuid)
// @Param comment body service.CreateCommentRequest true "Inline comment creation request with text position data"
// @Success 201 {object} service.CommentResponse "Successfully created user story inline comment"
// @Failure 400 {object} map[string]string "Invalid request - missing inline comment data or invalid text positions"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "User story not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/user-stories/{id}/comments/inline [post]
func (h *CommentHandler) CreateUserStoryInlineComment(c *gin.Context) {
	h.createInlineCommentForEntity(c, models.EntityTypeUserStory)
}

// CreateAcceptanceCriteriaInlineComment handles POST /api/v1/acceptance-criteria/:id/comments/inline
// @Summary Create an inline comment on acceptance criteria text
// @Description Create an inline comment linked to specific text positions within acceptance criteria description or conditions. Requires authentication.
// @Tags acceptance-criteria,comments,inline-comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Acceptance Criteria ID" format(uuid)
// @Param comment body service.CreateCommentRequest true "Inline comment creation request with text position data"
// @Success 201 {object} service.CommentResponse "Successfully created acceptance criteria inline comment"
// @Failure 400 {object} map[string]string "Invalid request - missing inline comment data or invalid text positions"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Acceptance criteria not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/acceptance-criteria/{id}/comments/inline [post]
func (h *CommentHandler) CreateAcceptanceCriteriaInlineComment(c *gin.Context) {
	h.createInlineCommentForEntity(c, models.EntityTypeAcceptanceCriteria)
}

// CreateRequirementInlineComment handles POST /api/v1/requirements/:id/comments/inline
// @Summary Create an inline comment on a requirement's text
// @Description Create an inline comment linked to specific text positions within a requirement's description or specification. Requires authentication.
// @Tags requirements,comments,inline-comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Requirement ID" format(uuid)
// @Param comment body service.CreateCommentRequest true "Inline comment creation request with text position data"
// @Success 201 {object} service.CommentResponse "Successfully created requirement inline comment"
// @Failure 400 {object} map[string]string "Invalid request - missing inline comment data or invalid text positions"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Requirement not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/requirements/{id}/comments/inline [post]
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

	// Get current user ID from JWT token context
	authorID, ok := auth.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User authentication required",
		})
		return
	}

	// Set entity type, ID, and author ID from context
	req.EntityType = entityType
	req.EntityID = entityID
	req.AuthorID = uuid.MustParse(authorID)

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
// @Summary Get visible inline comments for an entity
// @Description Retrieve all inline comments that are still valid (visible) for an entity, excluding those that may have become invalid due to text changes. Requires authentication.
// @Tags comments,inline-comments
// @Produce json
// @Security BearerAuth
// @Param entityType path string true "Entity type" Enums(epic,user_story,acceptance_criteria,requirement)
// @Param id path string true "Entity ID" format(uuid)
// @Success 200 {object} map[string]interface{} "Successfully retrieved visible inline comments" example({"comments": [{"id": "123e4567-e89b-12d3-a456-426614174000", "linked_text": "OAuth 2.0 authentication", "text_position_start": 45, "text_position_end": 67, "content": "Need to clarify which OAuth flow to use"}], "count": 1})
// @Failure 400 {object} map[string]string "Invalid entity type or malformed entity ID"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Entity not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/{entityType}/{id}/comments/inline/visible [get]
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
// @Summary Get visible inline comments for an epic
// @Description Retrieve all visible inline comments for a specific epic, excluding those invalidated by text changes. Requires authentication.
// @Tags epics,comments,inline-comments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Epic ID" format(uuid)
// @Success 200 {object} map[string]interface{} "Successfully retrieved epic inline comments"
// @Failure 400 {object} map[string]string "Invalid epic ID format"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Epic not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/epics/{id}/comments/inline/visible [get]
func (h *CommentHandler) GetEpicVisibleInlineComments(c *gin.Context) {
	h.getVisibleInlineCommentsForEntity(c, models.EntityTypeEpic)
}

// GetUserStoryVisibleInlineComments handles GET /api/v1/user-stories/:id/comments/inline/visible
// @Summary Get visible inline comments for a user story
// @Description Retrieve all visible inline comments for a specific user story, excluding those invalidated by text changes. Requires authentication.
// @Tags user-stories,comments,inline-comments
// @Produce json
// @Security BearerAuth
// @Param id path string true "User Story ID" format(uuid)
// @Success 200 {object} map[string]interface{} "Successfully retrieved user story inline comments"
// @Failure 400 {object} map[string]string "Invalid user story ID format"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "User story not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/user-stories/{id}/comments/inline/visible [get]
func (h *CommentHandler) GetUserStoryVisibleInlineComments(c *gin.Context) {
	h.getVisibleInlineCommentsForEntity(c, models.EntityTypeUserStory)
}

// GetAcceptanceCriteriaVisibleInlineComments handles GET /api/v1/acceptance-criteria/:id/comments/inline/visible
// @Summary Get visible inline comments for acceptance criteria
// @Description Retrieve all visible inline comments for specific acceptance criteria, excluding those invalidated by text changes. Requires authentication.
// @Tags acceptance-criteria,comments,inline-comments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Acceptance Criteria ID" format(uuid)
// @Success 200 {object} map[string]interface{} "Successfully retrieved acceptance criteria inline comments"
// @Failure 400 {object} map[string]string "Invalid acceptance criteria ID format"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Acceptance criteria not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/acceptance-criteria/{id}/comments/inline/visible [get]
func (h *CommentHandler) GetAcceptanceCriteriaVisibleInlineComments(c *gin.Context) {
	h.getVisibleInlineCommentsForEntity(c, models.EntityTypeAcceptanceCriteria)
}

// GetRequirementVisibleInlineComments handles GET /api/v1/requirements/:id/comments/inline/visible
// @Summary Get visible inline comments for a requirement
// @Description Retrieve all visible inline comments for a specific requirement, excluding those invalidated by text changes. Requires authentication.
// @Tags requirements,comments,inline-comments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Requirement ID" format(uuid)
// @Success 200 {object} map[string]interface{} "Successfully retrieved requirement inline comments"
// @Failure 400 {object} map[string]string "Invalid requirement ID format"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Requirement not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/requirements/{id}/comments/inline/visible [get]
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
// @Summary Validate inline comments after text changes
// @Description Validate and update inline comment positions after entity text content has been modified. This ensures inline comments remain accurately positioned. Requires authentication.
// @Tags comments,inline-comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param entityType path string true "Entity type" Enums(epic,user_story,acceptance_criteria,requirement)
// @Param id path string true "Entity ID" format(uuid)
// @Param validation body object true "Text validation request" example({"new_description": "Updated entity description with modified text content"})
// @Success 200 {object} map[string]string "Successfully validated inline comments" example({"message": "Inline comments validated successfully"})
// @Failure 400 {object} map[string]string "Invalid entity ID format or missing new_description"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 500 {object} map[string]string "Internal server error during validation"
// @Router /api/v1/{entityType}/{id}/comments/inline/validate [post]
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
// @Summary Validate epic inline comments after text changes
// @Description Validate and update inline comment positions after an epic's text content has been modified. Requires authentication.
// @Tags epics,comments,inline-comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Epic ID" format(uuid)
// @Param validation body object true "Text validation request" example({"new_description": "Updated epic description with modified text content"})
// @Success 200 {object} map[string]string "Successfully validated epic inline comments"
// @Failure 400 {object} map[string]string "Invalid epic ID format or missing new_description"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 500 {object} map[string]string "Internal server error during validation"
// @Router /api/v1/epics/{id}/comments/inline/validate [post]
func (h *CommentHandler) ValidateEpicInlineComments(c *gin.Context) {
	h.validateInlineCommentsForEntity(c, models.EntityTypeEpic)
}

// ValidateUserStoryInlineComments handles POST /api/v1/user-stories/:id/comments/inline/validate
// @Summary Validate user story inline comments after text changes
// @Description Validate and update inline comment positions after a user story's text content has been modified. Requires authentication.
// @Tags user-stories,comments,inline-comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User Story ID" format(uuid)
// @Param validation body object true "Text validation request" example({"new_description": "Updated user story description with modified text content"})
// @Success 200 {object} map[string]string "Successfully validated user story inline comments"
// @Failure 400 {object} map[string]string "Invalid user story ID format or missing new_description"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 500 {object} map[string]string "Internal server error during validation"
// @Router /api/v1/user-stories/{id}/comments/inline/validate [post]
func (h *CommentHandler) ValidateUserStoryInlineComments(c *gin.Context) {
	h.validateInlineCommentsForEntity(c, models.EntityTypeUserStory)
}

// ValidateAcceptanceCriteriaInlineComments handles POST /api/v1/acceptance-criteria/:id/comments/inline/validate
// @Summary Validate acceptance criteria inline comments after text changes
// @Description Validate and update inline comment positions after acceptance criteria text content has been modified. Requires authentication.
// @Tags acceptance-criteria,comments,inline-comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Acceptance Criteria ID" format(uuid)
// @Param validation body object true "Text validation request" example({"new_description": "Updated acceptance criteria description with modified text content"})
// @Success 200 {object} map[string]string "Successfully validated acceptance criteria inline comments"
// @Failure 400 {object} map[string]string "Invalid acceptance criteria ID format or missing new_description"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 500 {object} map[string]string "Internal server error during validation"
// @Router /api/v1/acceptance-criteria/{id}/comments/inline/validate [post]
func (h *CommentHandler) ValidateAcceptanceCriteriaInlineComments(c *gin.Context) {
	h.validateInlineCommentsForEntity(c, models.EntityTypeAcceptanceCriteria)
}

// ValidateRequirementInlineComments handles POST /api/v1/requirements/:id/comments/inline/validate
// @Summary Validate requirement inline comments after text changes
// @Description Validate and update inline comment positions after a requirement's text content has been modified. Requires authentication.
// @Tags requirements,comments,inline-comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Requirement ID" format(uuid)
// @Param validation body object true "Text validation request" example({"new_description": "Updated requirement description with modified text content"})
// @Success 200 {object} map[string]string "Successfully validated requirement inline comments"
// @Failure 400 {object} map[string]string "Invalid requirement ID format or missing new_description"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 500 {object} map[string]string "Internal server error during validation"
// @Router /api/v1/requirements/{id}/comments/inline/validate [post]
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
// @Summary Create a new comment on an epic
// @Description Create a new comment (general or inline) on a specific epic. Supports threaded discussions through parent_comment_id.
// @Tags epics,comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Epic ID" format(uuid)
// @Param comment body service.CreateCommentRequest true "Comment creation request"
// @Success 201 {object} service.CommentResponse "Successfully created epic comment"
// @Failure 400 {object} map[string]string "Invalid request - malformed epic ID, missing required fields, or invalid inline comment data"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Epic not found or parent comment not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/epics/{id}/comments [post]
func (h *CommentHandler) CreateEpicComment(c *gin.Context) {
	h.createCommentForEntity(c, models.EntityTypeEpic)
}

// CreateUserStoryComment handles POST /api/v1/user-stories/:id/comments
// @Summary Create a new comment on a user story
// @Description Create a new comment (general or inline) on a specific user story. Supports threaded discussions through parent_comment_id.
// @Tags user-stories,comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User Story ID" format(uuid)
// @Param comment body service.CreateCommentRequest true "Comment creation request"
// @Success 201 {object} service.CommentResponse "Successfully created user story comment"
// @Failure 400 {object} map[string]string "Invalid request - malformed user story ID, missing required fields, or invalid inline comment data"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "User story not found or parent comment not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/user-stories/{id}/comments [post]
func (h *CommentHandler) CreateUserStoryComment(c *gin.Context) {
	h.createCommentForEntity(c, models.EntityTypeUserStory)
}

// CreateAcceptanceCriteriaComment handles POST /api/v1/acceptance-criteria/:id/comments
// @Summary Create a new comment on acceptance criteria
// @Description Create a new comment (general or inline) on specific acceptance criteria. Supports threaded discussions through parent_comment_id.
// @Tags acceptance-criteria,comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Acceptance Criteria ID" format(uuid)
// @Param comment body service.CreateCommentRequest true "Comment creation request"
// @Success 201 {object} service.CommentResponse "Successfully created acceptance criteria comment"
// @Failure 400 {object} map[string]string "Invalid request - malformed acceptance criteria ID, missing required fields, or invalid inline comment data"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Acceptance criteria not found or parent comment not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/acceptance-criteria/{id}/comments [post]
func (h *CommentHandler) CreateAcceptanceCriteriaComment(c *gin.Context) {
	h.createCommentForEntity(c, models.EntityTypeAcceptanceCriteria)
}

// CreateRequirementComment handles POST /api/v1/requirements/:id/comments
// @Summary Create a new comment on a requirement
// @Description Create a new comment (general or inline) on a specific requirement. Supports threaded discussions through parent_comment_id.
// @Tags requirements,comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Requirement ID" format(uuid)
// @Param comment body service.CreateCommentRequest true "Comment creation request"
// @Success 201 {object} service.CommentResponse "Successfully created requirement comment"
// @Failure 400 {object} map[string]string "Invalid request - malformed requirement ID, missing required fields, or invalid inline comment data"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Requirement not found or parent comment not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/requirements/{id}/comments [post]
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

	// Get current user ID from JWT token context
	authorID, ok := auth.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User authentication required",
		})
		return
	}

	// Set entity type, ID, and author ID from context
	req.EntityType = entityType
	req.EntityID = entityID
	req.AuthorID = uuid.MustParse(authorID)

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
// @Summary Get all comments for an epic
// @Description Retrieve all comments for a specific epic with optional filtering by status and threading. Supports both flat and threaded comment structures.
// @Tags epics,comments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Epic ID" format(uuid)
// @Param threaded query boolean false "Return comments in threaded structure"
// @Param inline query boolean false "Return only inline comments"
// @Param status query string false "Filter by resolution status" Enums(resolved,unresolved)
// @Success 200 {object} map[string]interface{} "Successfully retrieved epic comments" example({"comments": [{"id": "123e4567-e89b-12d3-a456-426614174000", "content": "This needs clarification", "is_resolved": false}], "count": 1})
// @Failure 400 {object} map[string]string "Invalid epic ID format"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Epic not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/epics/{id}/comments [get]
func (h *CommentHandler) GetEpicComments(c *gin.Context) {
	h.getCommentsForEntity(c, models.EntityTypeEpic)
}

// GetUserStoryComments handles GET /api/v1/user-stories/:id/comments
// @Summary Get all comments for a user story
// @Description Retrieve all comments for a specific user story with optional filtering by status and threading. Supports both flat and threaded comment structures.
// @Tags user-stories,comments
// @Produce json
// @Security BearerAuth
// @Param id path string true "User Story ID" format(uuid)
// @Param threaded query boolean false "Return comments in threaded structure"
// @Param inline query boolean false "Return only inline comments"
// @Param status query string false "Filter by resolution status" Enums(resolved,unresolved)
// @Success 200 {object} map[string]interface{} "Successfully retrieved user story comments" example({"comments": [{"id": "123e4567-e89b-12d3-a456-426614174000", "content": "This needs clarification", "is_resolved": false}], "count": 1})
// @Failure 400 {object} map[string]string "Invalid user story ID format"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "User story not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/user-stories/{id}/comments [get]
func (h *CommentHandler) GetUserStoryComments(c *gin.Context) {
	h.getCommentsForEntity(c, models.EntityTypeUserStory)
}

// GetAcceptanceCriteriaComments handles GET /api/v1/acceptance-criteria/:id/comments
// @Summary Get all comments for acceptance criteria
// @Description Retrieve all comments for specific acceptance criteria with optional filtering by status and threading. Supports both flat and threaded comment structures.
// @Tags acceptance-criteria,comments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Acceptance Criteria ID" format(uuid)
// @Param threaded query boolean false "Return comments in threaded structure"
// @Param inline query boolean false "Return only inline comments"
// @Param status query string false "Filter by resolution status" Enums(resolved,unresolved)
// @Success 200 {object} map[string]interface{} "Successfully retrieved acceptance criteria comments" example({"comments": [{"id": "123e4567-e89b-12d3-a456-426614174000", "content": "This needs clarification", "is_resolved": false}], "count": 1})
// @Failure 400 {object} map[string]string "Invalid acceptance criteria ID format"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Acceptance criteria not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/acceptance-criteria/{id}/comments [get]
func (h *CommentHandler) GetAcceptanceCriteriaComments(c *gin.Context) {
	h.getCommentsForEntity(c, models.EntityTypeAcceptanceCriteria)
}

// GetRequirementComments handles GET /api/v1/requirements/:id/comments
// @Summary Get all comments for a requirement
// @Description Retrieve all comments for a specific requirement with optional filtering by status and threading. Supports both flat and threaded comment structures.
// @Tags requirements,comments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Requirement ID" format(uuid)
// @Param threaded query boolean false "Return comments in threaded structure"
// @Param inline query boolean false "Return only inline comments"
// @Param status query string false "Filter by resolution status" Enums(resolved,unresolved)
// @Success 200 {object} map[string]interface{} "Successfully retrieved requirement comments" example({"comments": [{"id": "123e4567-e89b-12d3-a456-426614174000", "content": "This needs clarification", "is_resolved": false}], "count": 1})
// @Failure 400 {object} map[string]string "Invalid requirement ID format"
// @Failure 401 {object} map[string]string "Authentication required"
// @Failure 404 {object} map[string]string "Requirement not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/requirements/{id}/comments [get]
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
			}
		}
		comments = filteredComments
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": comments,
		"count":    len(comments),
	})
}
