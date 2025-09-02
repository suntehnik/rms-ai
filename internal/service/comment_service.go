package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

var (
	ErrCommentNotFound           = errors.New("comment not found")
	ErrCommentHasReplies         = errors.New("comment has replies and cannot be deleted")
	ErrCommentInvalidEntityType  = errors.New("invalid entity type")
	ErrCommentEntityNotFound     = errors.New("entity not found")
	ErrCommentAuthorNotFound     = errors.New("author not found")
	ErrParentCommentNotFound     = errors.New("parent comment not found")
	ErrParentCommentWrongEntity  = errors.New("parent comment must be on the same entity")
	ErrEmptyContent              = errors.New("content cannot be empty")
	ErrInvalidInlineCommentData  = errors.New("inline comments require linked_text, text_position_start, and text_position_end")
	ErrInvalidTextPosition       = errors.New("invalid text position: start must be >= 0 and end must be >= start")
	ErrEmptyLinkedText           = errors.New("linked_text cannot be empty for inline comments")
)

// CommentService defines the interface for comment business logic
type CommentService interface {
	CreateComment(req CreateCommentRequest) (*CommentResponse, error)
	GetComment(id uuid.UUID) (*CommentResponse, error)
	UpdateComment(id uuid.UUID, req UpdateCommentRequest) (*CommentResponse, error)
	DeleteComment(id uuid.UUID) error
	GetCommentsByEntity(entityType models.EntityType, entityID uuid.UUID) ([]CommentResponse, error)
	GetThreadedComments(entityType models.EntityType, entityID uuid.UUID) ([]CommentResponse, error)
	GetCommentsByStatus(isResolved bool) ([]CommentResponse, error)
	GetInlineComments(entityType models.EntityType, entityID uuid.UUID) ([]CommentResponse, error)
	GetVisibleInlineComments(entityType models.EntityType, entityID uuid.UUID) ([]CommentResponse, error)
	ValidateInlineCommentsAfterTextChange(entityType models.EntityType, entityID uuid.UUID, newDescription string) error
	ResolveComment(id uuid.UUID) (*CommentResponse, error)
	UnresolveComment(id uuid.UUID) (*CommentResponse, error)
}

// commentService implements CommentService interface
type commentService struct {
	commentRepo repository.CommentRepository
	userRepo    repository.UserRepository
	repos       *repository.Repositories
}

// NewCommentService creates a new comment service instance
func NewCommentService(repos *repository.Repositories) CommentService {
	return &commentService{
		commentRepo: repos.Comment,
		userRepo:    repos.User,
		repos:       repos,
	}
}

// CreateCommentRequest represents the request to create a comment
type CreateCommentRequest struct {
	EntityType        models.EntityType `json:"entity_type"`
	EntityID          uuid.UUID         `json:"entity_id"`
	ParentCommentID   *uuid.UUID        `json:"parent_comment_id"`
	AuthorID          uuid.UUID         `json:"author_id" binding:"required"`
	Content           string            `json:"content"`
	LinkedText        *string           `json:"linked_text"`
	TextPositionStart *int              `json:"text_position_start"`
	TextPositionEnd   *int              `json:"text_position_end"`
}

// UpdateCommentRequest represents the request to update a comment
type UpdateCommentRequest struct {
	Content string `json:"content"`
}

// CommentResponse represents a comment in API responses
type CommentResponse struct {
	ID                uuid.UUID                `json:"id"`
	EntityType        models.EntityType        `json:"entity_type"`
	EntityID          uuid.UUID                `json:"entity_id"`
	ParentCommentID   *uuid.UUID               `json:"parent_comment_id"`
	AuthorID          uuid.UUID                `json:"author_id"`
	Author            *models.User             `json:"author,omitempty"`
	CreatedAt         string                   `json:"created_at"`
	UpdatedAt         string                   `json:"updated_at"`
	Content           string                   `json:"content"`
	IsResolved        bool                     `json:"is_resolved"`
	LinkedText        *string                  `json:"linked_text"`
	TextPositionStart *int                     `json:"text_position_start"`
	TextPositionEnd   *int                     `json:"text_position_end"`
	Replies           []CommentResponse        `json:"replies,omitempty"`
	IsInline          bool                     `json:"is_inline"`
	IsReply           bool                     `json:"is_reply"`
	Depth             int                      `json:"depth"`
}

// CreateComment creates a new comment
func (s *commentService) CreateComment(req CreateCommentRequest) (*CommentResponse, error) {
	// Validate entity type
	if !isValidEntityType(req.EntityType) {
		return nil, ErrCommentInvalidEntityType
	}

	// Validate entity exists
	if err := s.validateEntityExists(req.EntityType, req.EntityID); err != nil {
		return nil, err
	}

	// Validate author exists
	if _, err := s.userRepo.GetByID(req.AuthorID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCommentAuthorNotFound
		}
		return nil, fmt.Errorf("failed to validate author: %w", err)
	}

	// Validate parent comment if specified
	if req.ParentCommentID != nil {
		parentComment, err := s.commentRepo.GetByID(*req.ParentCommentID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, ErrParentCommentNotFound
			}
			return nil, fmt.Errorf("failed to validate parent comment: %w", err)
		}

		// Ensure parent comment is on the same entity
		if parentComment.EntityType != req.EntityType || parentComment.EntityID != req.EntityID {
			return nil, ErrParentCommentWrongEntity
		}
	}

	// Validate inline comment data
	if err := s.validateInlineCommentData(req.LinkedText, req.TextPositionStart, req.TextPositionEnd); err != nil {
		return nil, err
	}

	// For inline comments, validate that the linked text matches the actual text fragment
	if req.LinkedText != nil && req.TextPositionStart != nil && req.TextPositionEnd != nil {
		if err := s.validateTextFragment(req.EntityType, req.EntityID, *req.LinkedText, *req.TextPositionStart, *req.TextPositionEnd); err != nil {
			return nil, fmt.Errorf("text fragment validation failed: %w", err)
		}
	}

	// Validate content
	if strings.TrimSpace(req.Content) == "" {
		return nil, ErrEmptyContent
	}

	// Create comment
	comment := &models.Comment{
		EntityType:        req.EntityType,
		EntityID:          req.EntityID,
		ParentCommentID:   req.ParentCommentID,
		AuthorID:          req.AuthorID,
		Content:           strings.TrimSpace(req.Content),
		IsResolved:        false,
		LinkedText:        req.LinkedText,
		TextPositionStart: req.TextPositionStart,
		TextPositionEnd:   req.TextPositionEnd,
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return s.toCommentResponse(comment), nil
}

// GetComment retrieves a comment by ID
func (s *commentService) GetComment(id uuid.UUID) (*CommentResponse, error) {
	comment, err := s.commentRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	return s.toCommentResponse(comment), nil
}

// UpdateComment updates an existing comment
func (s *commentService) UpdateComment(id uuid.UUID, req UpdateCommentRequest) (*CommentResponse, error) {
	// Validate content
	if strings.TrimSpace(req.Content) == "" {
		return nil, ErrEmptyContent
	}

	comment, err := s.commentRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	// Update comment
	comment.Content = strings.TrimSpace(req.Content)

	if err := s.commentRepo.Update(comment); err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	return s.toCommentResponse(comment), nil
}

// DeleteComment deletes a comment
func (s *commentService) DeleteComment(id uuid.UUID) error {
	_, err := s.commentRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrCommentNotFound
		}
		return fmt.Errorf("failed to get comment: %w", err)
	}

	// Check if comment has replies
	replies, err := s.commentRepo.GetByParent(id)
	if err != nil {
		return fmt.Errorf("failed to check for replies: %w", err)
	}

	if len(replies) > 0 {
		return ErrCommentHasReplies
	}

	if err := s.commentRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	return nil
}

// GetCommentsByEntity retrieves all comments for an entity
func (s *commentService) GetCommentsByEntity(entityType models.EntityType, entityID uuid.UUID) ([]CommentResponse, error) {
	// Validate entity type
	if !isValidEntityType(entityType) {
		return nil, ErrCommentInvalidEntityType
	}

	// Validate entity exists
	if err := s.validateEntityExists(entityType, entityID); err != nil {
		return nil, err
	}

	comments, err := s.commentRepo.GetByEntity(entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	responses := make([]CommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = *s.toCommentResponse(&comment)
	}

	return responses, nil
}

// GetThreadedComments retrieves comments in threaded format for an entity
func (s *commentService) GetThreadedComments(entityType models.EntityType, entityID uuid.UUID) ([]CommentResponse, error) {
	// Validate entity type
	if !isValidEntityType(entityType) {
		return nil, ErrCommentInvalidEntityType
	}

	// Validate entity exists
	if err := s.validateEntityExists(entityType, entityID); err != nil {
		return nil, err
	}

	comments, err := s.commentRepo.GetThreaded(entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get threaded comments: %w", err)
	}

	responses := make([]CommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = *s.toCommentResponseWithReplies(&comment)
	}

	return responses, nil
}

// GetCommentsByStatus retrieves comments by resolution status
func (s *commentService) GetCommentsByStatus(isResolved bool) ([]CommentResponse, error) {
	comments, err := s.commentRepo.GetByStatus(isResolved)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by status: %w", err)
	}

	responses := make([]CommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = *s.toCommentResponse(&comment)
	}

	return responses, nil
}

// GetInlineComments retrieves inline comments for an entity
func (s *commentService) GetInlineComments(entityType models.EntityType, entityID uuid.UUID) ([]CommentResponse, error) {
	// Validate entity type
	if !isValidEntityType(entityType) {
		return nil, ErrCommentInvalidEntityType
	}

	// Validate entity exists
	if err := s.validateEntityExists(entityType, entityID); err != nil {
		return nil, err
	}

	comments, err := s.commentRepo.GetInlineComments(entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inline comments: %w", err)
	}

	responses := make([]CommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = *s.toCommentResponse(&comment)
	}

	return responses, nil
}

// ResolveComment marks a comment as resolved
func (s *commentService) ResolveComment(id uuid.UUID) (*CommentResponse, error) {
	comment, err := s.commentRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	comment.MarkAsResolved()

	if err := s.commentRepo.Update(comment); err != nil {
		return nil, fmt.Errorf("failed to resolve comment: %w", err)
	}

	return s.toCommentResponse(comment), nil
}

// UnresolveComment marks a comment as unresolved
func (s *commentService) UnresolveComment(id uuid.UUID) (*CommentResponse, error) {
	comment, err := s.commentRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	comment.MarkAsUnresolved()

	if err := s.commentRepo.Update(comment); err != nil {
		return nil, fmt.Errorf("failed to unresolve comment: %w", err)
	}

	return s.toCommentResponse(comment), nil
}

// ValidateInlineCommentsAfterTextChange validates inline comments after entity description changes
// This method should be called whenever an entity's description is updated
func (s *commentService) ValidateInlineCommentsAfterTextChange(entityType models.EntityType, entityID uuid.UUID, newDescription string) error {
	// Get all inline comments for this entity
	inlineComments, err := s.commentRepo.GetInlineComments(entityType, entityID)
	if err != nil {
		return fmt.Errorf("failed to get inline comments: %w", err)
	}

	for _, comment := range inlineComments {
		if comment.LinkedText == nil || comment.TextPositionStart == nil || comment.TextPositionEnd == nil {
			continue // Skip non-inline comments
		}

		start := *comment.TextPositionStart
		end := *comment.TextPositionEnd
		linkedText := *comment.LinkedText

		// Check if the text positions are still valid
		if start < 0 || end > len(newDescription) || start > end {
			// Hide the comment by marking it as invalid
			if err := s.hideInlineComment(&comment); err != nil {
				return fmt.Errorf("failed to hide invalid comment %s: %w", comment.ID, err)
			}
			continue
		}

		// Check if the linked text still matches
		actualFragment := newDescription[start:end]
		if actualFragment != linkedText {
			// Hide the comment because the text has changed
			if err := s.hideInlineComment(&comment); err != nil {
				return fmt.Errorf("failed to hide changed comment %s: %w", comment.ID, err)
			}
		}
	}

	return nil
}

// hideInlineComment hides an inline comment by setting special markers
// We don't delete the comment to preserve audit trail, but mark it as hidden
func (s *commentService) hideInlineComment(comment *models.Comment) error {
	// Mark the comment as hidden by setting linked text to null
	// This effectively makes it invisible in inline comment queries
	comment.LinkedText = nil
	comment.TextPositionStart = nil
	comment.TextPositionEnd = nil
	
	// Add a note to the content indicating why it was hidden
	comment.Content = "[HIDDEN: Linked text was modified or deleted] " + comment.Content

	return s.commentRepo.Update(comment)
}

// GetVisibleInlineComments retrieves only visible inline comments for an entity
func (s *commentService) GetVisibleInlineComments(entityType models.EntityType, entityID uuid.UUID) ([]CommentResponse, error) {
	// Validate entity type
	if !isValidEntityType(entityType) {
		return nil, ErrCommentInvalidEntityType
	}

	// Validate entity exists
	if err := s.validateEntityExists(entityType, entityID); err != nil {
		return nil, err
	}

	// Get inline comments (this already filters for non-null linked_text)
	comments, err := s.commentRepo.GetInlineComments(entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inline comments: %w", err)
	}

	responses := make([]CommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = *s.toCommentResponse(&comment)
	}

	return responses, nil
}

// validateEntityExists validates that the specified entity exists
func (s *commentService) validateEntityExists(entityType models.EntityType, entityID uuid.UUID) error {
	switch entityType {
	case models.EntityTypeEpic:
		if exists, err := s.repos.Epic.Exists(entityID); err != nil {
			return fmt.Errorf("failed to validate epic: %w", err)
		} else if !exists {
			return ErrCommentEntityNotFound
		}
	case models.EntityTypeUserStory:
		if exists, err := s.repos.UserStory.Exists(entityID); err != nil {
			return fmt.Errorf("failed to validate user story: %w", err)
		} else if !exists {
			return ErrCommentEntityNotFound
		}
	case models.EntityTypeAcceptanceCriteria:
		if exists, err := s.repos.AcceptanceCriteria.Exists(entityID); err != nil {
			return fmt.Errorf("failed to validate acceptance criteria: %w", err)
		} else if !exists {
			return ErrCommentEntityNotFound
		}
	case models.EntityTypeRequirement:
		if exists, err := s.repos.Requirement.Exists(entityID); err != nil {
			return fmt.Errorf("failed to validate requirement: %w", err)
		} else if !exists {
			return ErrCommentEntityNotFound
		}
	default:
		return ErrCommentInvalidEntityType
	}
	return nil
}

// validateInlineCommentData validates inline comment data consistency
func (s *commentService) validateInlineCommentData(linkedText *string, start *int, end *int) error {
	// If any inline comment field is provided, all must be provided
	hasLinkedText := linkedText != nil && *linkedText != ""
	hasStart := start != nil
	hasEnd := end != nil

	if hasLinkedText || hasStart || hasEnd {
		if !hasLinkedText || !hasStart || !hasEnd {
			return ErrInvalidInlineCommentData
		}

		if *start < 0 || *end < *start {
			return ErrInvalidTextPosition
		}

		if strings.TrimSpace(*linkedText) == "" {
			return ErrEmptyLinkedText
		}
	}

	return nil
}

// validateTextFragment validates that the linked text matches the actual text in the entity's description
func (s *commentService) validateTextFragment(entityType models.EntityType, entityID uuid.UUID, linkedText string, start, end int) error {
	var description string

	// Get the entity's description based on type
	switch entityType {
	case models.EntityTypeEpic:
		epic, err := s.repos.Epic.GetByID(entityID)
		if err != nil {
			return fmt.Errorf("failed to get epic: %w", err)
		}
		if epic.Description != nil {
			description = *epic.Description
		}
	case models.EntityTypeUserStory:
		userStory, err := s.repos.UserStory.GetByID(entityID)
		if err != nil {
			return fmt.Errorf("failed to get user story: %w", err)
		}
		if userStory.Description != nil {
			description = *userStory.Description
		}
	case models.EntityTypeAcceptanceCriteria:
		acceptanceCriteria, err := s.repos.AcceptanceCriteria.GetByID(entityID)
		if err != nil {
			return fmt.Errorf("failed to get acceptance criteria: %w", err)
		}
		// AcceptanceCriteria.Description is a string, not a pointer
		description = acceptanceCriteria.Description
	case models.EntityTypeRequirement:
		requirement, err := s.repos.Requirement.GetByID(entityID)
		if err != nil {
			return fmt.Errorf("failed to get requirement: %w", err)
		}
		if requirement.Description != nil {
			description = *requirement.Description
		}
	default:
		return ErrCommentInvalidEntityType
	}

	// Validate text positions are within bounds
	if start < 0 || end > len(description) || start > end {
		return ErrInvalidTextPosition
	}

	// Extract the actual text fragment from the description
	actualFragment := description[start:end]

	// Compare with the provided linked text
	if actualFragment != linkedText {
		return errors.New("linked text does not match the actual text fragment in the description")
	}

	return nil
}

// isValidEntityType checks if the entity type is valid
func isValidEntityType(entityType models.EntityType) bool {
	validTypes := []models.EntityType{
		models.EntityTypeEpic,
		models.EntityTypeUserStory,
		models.EntityTypeAcceptanceCriteria,
		models.EntityTypeRequirement,
	}

	for _, validType := range validTypes {
		if entityType == validType {
			return true
		}
	}
	return false
}

// toCommentResponse converts a comment model to response format
func (s *commentService) toCommentResponse(comment *models.Comment) *CommentResponse {
	response := &CommentResponse{
		ID:                comment.ID,
		EntityType:        comment.EntityType,
		EntityID:          comment.EntityID,
		ParentCommentID:   comment.ParentCommentID,
		AuthorID:          comment.AuthorID,
		CreatedAt:         comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         comment.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Content:           comment.Content,
		IsResolved:        comment.IsResolved,
		LinkedText:        comment.LinkedText,
		TextPositionStart: comment.TextPositionStart,
		TextPositionEnd:   comment.TextPositionEnd,
		IsInline:          comment.IsInlineComment(),
		IsReply:           comment.IsReply(),
		Depth:             comment.GetDepth(),
	}

	// Load author if available
	if comment.Author.ID != uuid.Nil {
		response.Author = &comment.Author
	}

	return response
}

// toCommentResponseWithReplies converts a comment model to response format including replies
func (s *commentService) toCommentResponseWithReplies(comment *models.Comment) *CommentResponse {
	response := s.toCommentResponse(comment)

	// Convert replies
	if len(comment.Replies) > 0 {
		response.Replies = make([]CommentResponse, len(comment.Replies))
		for i, reply := range comment.Replies {
			response.Replies[i] = *s.toCommentResponseWithReplies(&reply)
		}
	}

	return response
}