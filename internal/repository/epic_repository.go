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

// GetByIDWithUsers retrieves an epic by its ID with creator and assignee preloaded
func (r *epicRepository) GetByIDWithUsers(id uuid.UUID) (*models.Epic, error) {
	var epic models.Epic
	if err := r.GetDB().Preload("Creator").Preload("Assignee").Where("id = ?", id).First(&epic).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, r.handleDBError(err)
	}
	return &epic, nil
}

// GetByReferenceIDWithUsers retrieves an epic by its reference ID with creator and assignee preloaded
func (r *epicRepository) GetByReferenceIDWithUsers(referenceID string) (*models.Epic, error) {
	var epic models.Epic
	if err := r.GetDB().Preload("Creator").Preload("Assignee").Where("reference_id = ?", referenceID).First(&epic).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, r.handleDBError(err)
	}
	return &epic, nil
}

// GetByReferenceIDWithUsersCaseInsensitive retrieves an epic by its reference ID (case-insensitive) with creator and assignee preloaded
func (r *epicRepository) GetByReferenceIDWithUsersCaseInsensitive(referenceID string) (*models.Epic, error) {
	var epic models.Epic

	query := r.GetDB().Preload("Creator").Preload("Assignee")
	var err error

	// Use ILIKE for PostgreSQL, LOWER() LIKE for SQLite compatibility
	if r.GetDB().Dialector.Name() == "postgres" {
		err = query.Where("reference_id ILIKE ?", referenceID).First(&epic).Error
	} else {
		// SQLite and other databases - use LOWER() LIKE for case-insensitive matching
		err = query.Where("LOWER(reference_id) LIKE LOWER(?)", referenceID).First(&epic).Error
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, r.handleDBError(err)
	}
	return &epic, nil
}

// ListWithIncludes retrieves epics with optional related entity preloading
func (r *epicRepository) ListWithIncludes(filters map[string]interface{}, includes []string, orderBy string, limit, offset int) ([]models.Epic, error) {
	var epics []models.Epic

	query := r.GetDB().Model(&models.Epic{})

	// Apply includes (preloads)
	for _, include := range includes {
		switch include {
		case "creator":
			query = query.Preload("Creator")
		case "assignee":
			query = query.Preload("Assignee")
		case "user_stories":
			query = query.Preload("UserStories")
		case "comments":
			query = query.Preload("Comments")
		}
	}

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

	if err := query.Find(&epics).Error; err != nil {
		return nil, r.handleDBError(err)
	}

	return epics, nil
}

// GetCompleteHierarchy retrieves an epic with all nested entities preloaded
// This method loads the complete hierarchy: Epic → [SteeringDocuments, UserStories] → [Requirements, AcceptanceCriteria]
// SteeringDocuments and UserStories are loaded at the same level under Epic
// Requirements and AcceptanceCriteria are loaded at the same level under each UserStory
func (r *epicRepository) GetCompleteHierarchy(id uuid.UUID) (*models.Epic, error) {
	var epic models.Epic
	err := r.GetDB().
		// Load steering documents with ordering
		Preload("SteeringDocuments", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		// Load user stories with ordering
		Preload("UserStories", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		// Load requirements for each user story
		Preload("UserStories.Requirements", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		// Load requirement types
		Preload("UserStories.Requirements.Type").
		// Load acceptance criteria for each user story
		Preload("UserStories.AcceptanceCriteria", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Where("id = ?", id).
		First(&epic).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, r.handleDBError(err)
	}
	return &epic, nil
}
