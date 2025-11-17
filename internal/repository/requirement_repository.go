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

	// Use LIKE for compatibility with SQLite (tests) and PostgreSQL
	searchPattern := "%" + searchText + "%"
	if err := r.GetDB().Where("title LIKE ? OR description LIKE ? OR reference_id LIKE ?",
		searchPattern, searchPattern, searchPattern).Find(&requirements).Error; err != nil {
		return nil, r.handleDBError(err)
	}

	return requirements, nil
}

// SearchByTextWithPagination performs full-text search on requirements with pagination
func (r *requirementRepository) SearchByTextWithPagination(searchText string, limit, offset int) ([]models.Requirement, int64, error) {
	var requirements []models.Requirement
	var totalCount int64

	// Use LIKE for compatibility with SQLite (tests) and PostgreSQL
	searchPattern := "%" + searchText + "%"

	// Get total count
	if err := r.GetDB().Model(&models.Requirement{}).Where("title LIKE ? OR description LIKE ? OR reference_id LIKE ?",
		searchPattern, searchPattern, searchPattern).Count(&totalCount).Error; err != nil {
		return nil, 0, r.handleDBError(err)
	}

	// Get paginated results
	if err := r.GetDB().Where("title LIKE ? OR description LIKE ? OR reference_id LIKE ?",
		searchPattern, searchPattern, searchPattern).Limit(limit).Offset(offset).Find(&requirements).Error; err != nil {
		return nil, 0, r.handleDBError(err)
	}

	return requirements, totalCount, nil
}

// GetByIDWithPreloads retrieves a requirement by its ID with all relationships preloaded
func (r *requirementRepository) GetByIDWithPreloads(id uuid.UUID) (*models.Requirement, error) {
	var requirement models.Requirement
	if err := r.GetDB().
		Preload("Creator").
		Preload("Assignee").
		Preload("UserStory").
		Preload("AcceptanceCriteria").
		Preload("Type").
		Where("id = ?", id).First(&requirement).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, r.handleDBError(err)
	}
	return &requirement, nil
}

// GetByReferenceIDWithPreloads retrieves a requirement by its reference ID with all relationships preloaded
func (r *requirementRepository) GetByReferenceIDWithPreloads(referenceID string) (*models.Requirement, error) {
	var requirement models.Requirement
	if err := r.GetDB().
		Preload("Creator").
		Preload("Assignee").
		Preload("UserStory").
		Preload("AcceptanceCriteria").
		Preload("Type").
		Where("reference_id = ?", referenceID).First(&requirement).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, r.handleDBError(err)
	}
	return &requirement, nil
}

// ListWithPreloads retrieves requirements with all relationships preloaded
func (r *requirementRepository) ListWithPreloads(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.Requirement, error) {
	var requirements []models.Requirement

	query := r.GetDB().
		Preload("Creator").
		Preload("Assignee").
		Preload("UserStory").
		Preload("AcceptanceCriteria").
		Preload("Type")

	// Apply filters
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	// Apply ordering
	if orderBy != "" {
		query = query.Order(orderBy)
	}

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&requirements).Error; err != nil {
		return nil, r.handleDBError(err)
	}

	return requirements, nil
}
