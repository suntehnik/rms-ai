package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
)

// acceptanceCriteriaRepository implements AcceptanceCriteriaRepository interface
type acceptanceCriteriaRepository struct {
	*BaseRepository[models.AcceptanceCriteria]
}

// NewAcceptanceCriteriaRepository creates a new acceptance criteria repository instance
func NewAcceptanceCriteriaRepository(db *gorm.DB) AcceptanceCriteriaRepository {
	return &acceptanceCriteriaRepository{
		BaseRepository: NewBaseRepository[models.AcceptanceCriteria](db),
	}
}

// GetByUserStory retrieves acceptance criteria by user story ID
func (r *acceptanceCriteriaRepository) GetByUserStory(userStoryID uuid.UUID) ([]models.AcceptanceCriteria, error) {
	var criteria []models.AcceptanceCriteria
	if err := r.GetDB().Where("user_story_id = ?", userStoryID).Find(&criteria).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return criteria, nil
}

// GetByAuthor retrieves acceptance criteria by author ID
func (r *acceptanceCriteriaRepository) GetByAuthor(authorID uuid.UUID) ([]models.AcceptanceCriteria, error) {
	var criteria []models.AcceptanceCriteria
	if err := r.GetDB().Where("author_id = ?", authorID).Find(&criteria).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return criteria, nil
}

// HasRequirements checks if acceptance criteria has any associated requirements
func (r *acceptanceCriteriaRepository) HasRequirements(id uuid.UUID) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.Requirement{}).Where("acceptance_criteria_id = ?", id).Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}
	return count > 0, nil
}

// CountByUserStory returns the count of acceptance criteria for a user story
func (r *acceptanceCriteriaRepository) CountByUserStory(userStoryID uuid.UUID) (int64, error) {
	var count int64
	if err := r.GetDB().Model(&models.AcceptanceCriteria{}).Where("user_story_id = ?", userStoryID).Count(&count).Error; err != nil {
		return 0, r.handleDBError(err)
	}
	return count, nil
}
