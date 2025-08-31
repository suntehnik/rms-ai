package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"product-requirements-management/internal/models"
)

// epicRepository implements EpicRepository interface
type epicRepository struct {
	*BaseRepository[models.Epic]
}

// NewEpicRepository creates a new epic repository instance
func NewEpicRepository(db *gorm.DB) EpicRepository {
	return &epicRepository{
		BaseRepository: NewBaseRepository[models.Epic](db),
	}
}

// GetWithUserStories retrieves an epic with its user stories
func (r *epicRepository) GetWithUserStories(id uuid.UUID) (*models.Epic, error) {
	var epic models.Epic
	if err := r.GetDB().Preload("UserStories").Where("id = ?", id).First(&epic).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, r.handleDBError(err)
	}
	return &epic, nil
}

// GetByCreator retrieves epics by creator ID
func (r *epicRepository) GetByCreator(creatorID uuid.UUID) ([]models.Epic, error) {
	var epics []models.Epic
	if err := r.GetDB().Where("creator_id = ?", creatorID).Find(&epics).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return epics, nil
}

// GetByAssignee retrieves epics by assignee ID
func (r *epicRepository) GetByAssignee(assigneeID uuid.UUID) ([]models.Epic, error) {
	var epics []models.Epic
	if err := r.GetDB().Where("assignee_id = ?", assigneeID).Find(&epics).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return epics, nil
}

// GetByStatus retrieves epics by status
func (r *epicRepository) GetByStatus(status models.EpicStatus) ([]models.Epic, error) {
	var epics []models.Epic
	if err := r.GetDB().Where("status = ?", status).Find(&epics).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return epics, nil
}

// GetByPriority retrieves epics by priority
func (r *epicRepository) GetByPriority(priority models.Priority) ([]models.Epic, error) {
	var epics []models.Epic
	if err := r.GetDB().Where("priority = ?", priority).Find(&epics).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return epics, nil
}

// HasUserStories checks if an epic has any user stories
func (r *epicRepository) HasUserStories(id uuid.UUID) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.UserStory{}).Where("epic_id = ?", id).Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}
	return count > 0, nil
}