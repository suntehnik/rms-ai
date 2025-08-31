package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"product-requirements-management/internal/models"
)

// userStoryRepository implements UserStoryRepository interface
type userStoryRepository struct {
	*BaseRepository[models.UserStory]
}

// NewUserStoryRepository creates a new user story repository instance
func NewUserStoryRepository(db *gorm.DB) UserStoryRepository {
	return &userStoryRepository{
		BaseRepository: NewBaseRepository[models.UserStory](db),
	}
}

// GetWithAcceptanceCriteria retrieves a user story with its acceptance criteria
func (r *userStoryRepository) GetWithAcceptanceCriteria(id uuid.UUID) (*models.UserStory, error) {
	var userStory models.UserStory
	if err := r.GetDB().Preload("AcceptanceCriteria").Where("id = ?", id).First(&userStory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, r.handleDBError(err)
	}
	return &userStory, nil
}

// GetWithRequirements retrieves a user story with its requirements
func (r *userStoryRepository) GetWithRequirements(id uuid.UUID) (*models.UserStory, error) {
	var userStory models.UserStory
	if err := r.GetDB().Preload("Requirements").Where("id = ?", id).First(&userStory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, r.handleDBError(err)
	}
	return &userStory, nil
}

// GetByEpic retrieves user stories by epic ID
func (r *userStoryRepository) GetByEpic(epicID uuid.UUID) ([]models.UserStory, error) {
	var userStories []models.UserStory
	if err := r.GetDB().Where("epic_id = ?", epicID).Find(&userStories).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return userStories, nil
}

// GetByCreator retrieves user stories by creator ID
func (r *userStoryRepository) GetByCreator(creatorID uuid.UUID) ([]models.UserStory, error) {
	var userStories []models.UserStory
	if err := r.GetDB().Where("creator_id = ?", creatorID).Find(&userStories).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return userStories, nil
}

// GetByAssignee retrieves user stories by assignee ID
func (r *userStoryRepository) GetByAssignee(assigneeID uuid.UUID) ([]models.UserStory, error) {
	var userStories []models.UserStory
	if err := r.GetDB().Where("assignee_id = ?", assigneeID).Find(&userStories).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return userStories, nil
}

// GetByStatus retrieves user stories by status
func (r *userStoryRepository) GetByStatus(status models.UserStoryStatus) ([]models.UserStory, error) {
	var userStories []models.UserStory
	if err := r.GetDB().Where("status = ?", status).Find(&userStories).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return userStories, nil
}

// GetByPriority retrieves user stories by priority
func (r *userStoryRepository) GetByPriority(priority models.Priority) ([]models.UserStory, error) {
	var userStories []models.UserStory
	if err := r.GetDB().Where("priority = ?", priority).Find(&userStories).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return userStories, nil
}

// HasAcceptanceCriteria checks if a user story has any acceptance criteria
func (r *userStoryRepository) HasAcceptanceCriteria(id uuid.UUID) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.AcceptanceCriteria{}).Where("user_story_id = ?", id).Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}
	return count > 0, nil
}

// HasRequirements checks if a user story has any requirements
func (r *userStoryRepository) HasRequirements(id uuid.UUID) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.Requirement{}).Where("user_story_id = ?", id).Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}
	return count > 0, nil
}