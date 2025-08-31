package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"product-requirements-management/internal/models"
)

// requirementRepository implements RequirementRepository interface
type requirementRepository struct {
	*BaseRepository[models.Requirement]
}

// NewRequirementRepository creates a new requirement repository instance
func NewRequirementRepository(db *gorm.DB) RequirementRepository {
	return &requirementRepository{
		BaseRepository: NewBaseRepository[models.Requirement](db),
	}
}

// GetWithRelationships retrieves a requirement with its relationships
func (r *requirementRepository) GetWithRelationships(id uuid.UUID) (*models.Requirement, error) {
	var requirement models.Requirement
	if err := r.GetDB().Preload("SourceRelationships").Preload("TargetRelationships").Where("id = ?", id).First(&requirement).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, r.handleDBError(err)
	}
	return &requirement, nil
}

// GetByUserStory retrieves requirements by user story ID
func (r *requirementRepository) GetByUserStory(userStoryID uuid.UUID) ([]models.Requirement, error) {
	var requirements []models.Requirement
	if err := r.GetDB().Where("user_story_id = ?", userStoryID).Find(&requirements).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return requirements, nil
}

// GetByAcceptanceCriteria retrieves requirements by acceptance criteria ID
func (r *requirementRepository) GetByAcceptanceCriteria(acceptanceCriteriaID uuid.UUID) ([]models.Requirement, error) {
	var requirements []models.Requirement
	if err := r.GetDB().Where("acceptance_criteria_id = ?", acceptanceCriteriaID).Find(&requirements).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return requirements, nil
}

// GetByCreator retrieves requirements by creator ID
func (r *requirementRepository) GetByCreator(creatorID uuid.UUID) ([]models.Requirement, error) {
	var requirements []models.Requirement
	if err := r.GetDB().Where("creator_id = ?", creatorID).Find(&requirements).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return requirements, nil
}

// GetByAssignee retrieves requirements by assignee ID
func (r *requirementRepository) GetByAssignee(assigneeID uuid.UUID) ([]models.Requirement, error) {
	var requirements []models.Requirement
	if err := r.GetDB().Where("assignee_id = ?", assigneeID).Find(&requirements).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return requirements, nil
}

// GetByStatus retrieves requirements by status
func (r *requirementRepository) GetByStatus(status models.RequirementStatus) ([]models.Requirement, error) {
	var requirements []models.Requirement
	if err := r.GetDB().Where("status = ?", status).Find(&requirements).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return requirements, nil
}

// GetByPriority retrieves requirements by priority
func (r *requirementRepository) GetByPriority(priority models.Priority) ([]models.Requirement, error) {
	var requirements []models.Requirement
	if err := r.GetDB().Where("priority = ?", priority).Find(&requirements).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return requirements, nil
}

// GetByType retrieves requirements by type ID
func (r *requirementRepository) GetByType(typeID uuid.UUID) ([]models.Requirement, error) {
	var requirements []models.Requirement
	if err := r.GetDB().Where("type_id = ?", typeID).Find(&requirements).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return requirements, nil
}

// HasRelationships checks if a requirement has any relationships
func (r *requirementRepository) HasRelationships(id uuid.UUID) (bool, error) {
	var count int64
	
	// Check source relationships
	if err := r.GetDB().Model(&models.RequirementRelationship{}).Where("source_requirement_id = ?", id).Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}
	if count > 0 {
		return true, nil
	}
	
	// Check target relationships
	if err := r.GetDB().Model(&models.RequirementRelationship{}).Where("target_requirement_id = ?", id).Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}
	
	return count > 0, nil
}

// SearchByText performs full-text search on requirements
func (r *requirementRepository) SearchByText(searchText string) ([]models.Requirement, error) {
	var requirements []models.Requirement
	
	// Use PostgreSQL full-text search or simple LIKE for now
	searchPattern := "%" + searchText + "%"
	if err := r.GetDB().Where("title ILIKE ? OR description ILIKE ? OR reference_id ILIKE ?", 
		searchPattern, searchPattern, searchPattern).Find(&requirements).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	
	return requirements, nil
}