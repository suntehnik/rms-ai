package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
)

// commentRepository implements CommentRepository interface
type commentRepository struct {
	*BaseRepository[models.Comment]
}

// NewCommentRepository creates a new comment repository instance
func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{
		BaseRepository: NewBaseRepository[models.Comment](db),
	}
}

// GetByEntity retrieves comments by entity type and ID
func (r *commentRepository) GetByEntity(entityType models.EntityType, entityID uuid.UUID) ([]models.Comment, error) {
	var comments []models.Comment
	if err := r.GetDB().Preload("Author").Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at ASC").Find(&comments).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return comments, nil
}

// GetByAuthor retrieves comments by author ID
func (r *commentRepository) GetByAuthor(authorID uuid.UUID) ([]models.Comment, error) {
	var comments []models.Comment
	if err := r.GetDB().Preload("Author").Where("author_id = ?", authorID).
		Order("created_at DESC").Find(&comments).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return comments, nil
}

// GetByParent retrieves replies to a parent comment
func (r *commentRepository) GetByParent(parentID uuid.UUID) ([]models.Comment, error) {
	var comments []models.Comment
	if err := r.GetDB().Preload("Author").Where("parent_comment_id = ?", parentID).
		Order("created_at ASC").Find(&comments).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return comments, nil
}

// GetByParentWithPagination retrieves replies to a parent comment with pagination
func (r *commentRepository) GetByParentWithPagination(parentID uuid.UUID, limit, offset int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var totalCount int64

	// Get total count
	if err := r.GetDB().Model(&models.Comment{}).Where("parent_comment_id = ?", parentID).Count(&totalCount).Error; err != nil {
		return nil, 0, r.handleDBError(err)
	}

	// Get paginated results
	if err := r.GetDB().Preload("Author").Where("parent_comment_id = ?", parentID).
		Order("created_at ASC").Limit(limit).Offset(offset).Find(&comments).Error; err != nil {
		return nil, 0, r.handleDBError(err)
	}

	return comments, totalCount, nil
}

// GetThreaded retrieves comments in threaded format for an entity
func (r *commentRepository) GetThreaded(entityType models.EntityType, entityID uuid.UUID) ([]models.Comment, error) {
	var comments []models.Comment
	if err := r.GetDB().Preload("Author").Preload("Replies").Preload("Replies.Author").
		Where("entity_type = ? AND entity_id = ? AND parent_comment_id IS NULL", entityType, entityID).
		Order("created_at ASC").Find(&comments).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return comments, nil
}

// GetByStatus retrieves comments by resolution status
func (r *commentRepository) GetByStatus(isResolved bool) ([]models.Comment, error) {
	var comments []models.Comment
	if err := r.GetDB().Preload("Author").Where("is_resolved = ?", isResolved).
		Order("created_at DESC").Find(&comments).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return comments, nil
}

// GetInlineComments retrieves inline comments for an entity
func (r *commentRepository) GetInlineComments(entityType models.EntityType, entityID uuid.UUID) ([]models.Comment, error) {
	var comments []models.Comment
	if err := r.GetDB().Preload("Author").Where("entity_type = ? AND entity_id = ? AND linked_text IS NOT NULL",
		entityType, entityID).Order("text_position_start ASC").Find(&comments).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return comments, nil
}

// GetByID retrieves a comment by ID with author information
func (r *commentRepository) GetByID(id uuid.UUID) (*models.Comment, error) {
	var comment models.Comment
	if err := r.GetDB().Preload("Author").Where("id = ?", id).First(&comment).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return &comment, nil
}
