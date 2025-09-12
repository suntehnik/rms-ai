package repository

import (
	"errors"
	"fmt"
	"strings"

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

// Create creates a new requirement with proper concurrent reference ID generation
func (r *requirementRepository) Create(requirement *models.Requirement) error {
	maxRetries := 10

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Generate reference ID if not set
		if requirement.ReferenceID == "" {
			if err := r.generateReferenceID(requirement, attempt); err != nil {
				return err
			}
		}

		// Try to create the requirement
		err := r.BaseRepository.Create(requirement)
		if err == nil {
			// Success
			return nil
		}

		// Check if it's a duplicate key error on reference_id
		if errors.Is(err, ErrDuplicateKey) ||
			(err != nil && (strings.Contains(strings.ToLower(err.Error()), "unique constraint") ||
				strings.Contains(strings.ToLower(err.Error()), "duplicate key") ||
				strings.Contains(strings.ToLower(err.Error()), "reference_id"))) {
			// Clear the reference ID and retry
			requirement.ReferenceID = ""
			continue
		}

		// Non-retryable error
		return err
	}

	// If we exhausted retries, use UUID-based reference ID as fallback
	requirement.ReferenceID = fmt.Sprintf("REQ-%s", uuid.New().String()[:8])
	return r.BaseRepository.Create(requirement)
}

// generateReferenceID generates a reference ID for the requirement
func (r *requirementRepository) generateReferenceID(requirement *models.Requirement, attempt int) error {
	// Get the current maximum reference number
	var maxRefNum int
	var maxRef string

	err := r.GetDB().Model(&models.Requirement{}).
		Select("reference_id").
		Where("reference_id LIKE 'REQ-%'").
		Order("reference_id DESC").
		Limit(1).
		Pluck("reference_id", &maxRef).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return r.handleDBError(err)
	}

	// Parse the maximum reference number
	if maxRef != "" {
		if _, scanErr := fmt.Sscanf(maxRef, "REQ-%d", &maxRefNum); scanErr != nil {
			// If parsing fails, fall back to counting all records
			var count int64
			if countErr := r.GetDB().Model(&models.Requirement{}).Count(&count).Error; countErr != nil {
				return r.handleDBError(countErr)
			}
			maxRefNum = int(count)
		}
	}

	// Generate next reference ID with attempt offset to reduce collisions
	nextNum := maxRefNum + 1 + attempt
	requirement.ReferenceID = fmt.Sprintf("REQ-%03d", nextNum)

	return nil
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
