package repository

import (
	"product-requirements-management/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// statusTransitionRepository implements StatusTransitionRepository interface
type statusTransitionRepository struct {
	db *gorm.DB
}

// NewStatusTransitionRepository creates a new status transition repository instance
func NewStatusTransitionRepository(db *gorm.DB) StatusTransitionRepository {
	return &statusTransitionRepository{db: db}
}

// Create creates a new status transition
func (r *statusTransitionRepository) Create(transition *models.StatusTransition) error {
	return r.db.Create(transition).Error
}

// GetByID retrieves a status transition by ID
func (r *statusTransitionRepository) GetByID(id uuid.UUID) (*models.StatusTransition, error) {
	var transition models.StatusTransition
	err := r.db.Preload("StatusModel").Preload("FromStatus").Preload("ToStatus").
		Where("id = ?", id).First(&transition).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &transition, nil
}

// GetByStatusModelID retrieves all transitions for a status model
func (r *statusTransitionRepository) GetByStatusModelID(statusModelID uuid.UUID) ([]models.StatusTransition, error) {
	var transitions []models.StatusTransition
	err := r.db.Preload("FromStatus").Preload("ToStatus").
		Where("status_model_id = ?", statusModelID).Find(&transitions).Error
	return transitions, err
}

// GetByFromStatus retrieves all transitions from a specific status
func (r *statusTransitionRepository) GetByFromStatus(fromStatusID uuid.UUID) ([]models.StatusTransition, error) {
	var transitions []models.StatusTransition
	err := r.db.Preload("ToStatus").Where("from_status_id = ?", fromStatusID).Find(&transitions).Error
	return transitions, err
}

// Update updates an existing status transition
func (r *statusTransitionRepository) Update(transition *models.StatusTransition) error {
	return r.db.Save(transition).Error
}

// Delete deletes a status transition by ID
func (r *statusTransitionRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.StatusTransition{}, id).Error
}

// List retrieves status transitions with filtering, ordering, and pagination
func (r *statusTransitionRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.StatusTransition, error) {
	var transitions []models.StatusTransition

	query := r.db.Preload("StatusModel").Preload("FromStatus").Preload("ToStatus")

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

	err := query.Find(&transitions).Error
	return transitions, err
}

// Exists checks if a status transition exists by ID
func (r *statusTransitionRepository) Exists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.StatusTransition{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// ExistsByTransition checks if a specific transition exists
func (r *statusTransitionRepository) ExistsByTransition(statusModelID, fromStatusID, toStatusID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.StatusTransition{}).
		Where("status_model_id = ? AND from_status_id = ? AND to_status_id = ?", statusModelID, fromStatusID, toStatusID).
		Count(&count).Error
	return count > 0, err
}
