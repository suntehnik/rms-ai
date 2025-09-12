package repository

import (
	"product-requirements-management/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// statusRepository implements StatusRepository interface
type statusRepository struct {
	db *gorm.DB
}

// NewStatusRepository creates a new status repository instance
func NewStatusRepository(db *gorm.DB) StatusRepository {
	return &statusRepository{db: db}
}

// Create creates a new status
func (r *statusRepository) Create(status *models.Status) error {
	return r.db.Create(status).Error
}

// GetByID retrieves a status by ID
func (r *statusRepository) GetByID(id uuid.UUID) (*models.Status, error) {
	var status models.Status
	err := r.db.Preload("StatusModel").Where("id = ?", id).First(&status).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &status, nil
}

// GetByStatusModelID retrieves all statuses for a status model
func (r *statusRepository) GetByStatusModelID(statusModelID uuid.UUID) ([]models.Status, error) {
	var statuses []models.Status
	err := r.db.Where("status_model_id = ?", statusModelID).Order("\"order\" ASC").Find(&statuses).Error
	return statuses, err
}

// GetByName retrieves a status by name within a status model
func (r *statusRepository) GetByName(statusModelID uuid.UUID, name string) (*models.Status, error) {
	var status models.Status
	err := r.db.Where("status_model_id = ? AND name = ?", statusModelID, name).First(&status).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &status, nil
}

// Update updates an existing status
func (r *statusRepository) Update(status *models.Status) error {
	return r.db.Save(status).Error
}

// Delete deletes a status by ID
func (r *statusRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Status{}, id).Error
}

// List retrieves statuses with filtering, ordering, and pagination
func (r *statusRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.Status, error) {
	var statuses []models.Status

	query := r.db.Preload("StatusModel")

	// Apply filters
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	// Apply ordering
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("\"order\" ASC")
	}

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&statuses).Error
	return statuses, err
}

// Exists checks if a status exists by ID
func (r *statusRepository) Exists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Status{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// ExistsByName checks if a status exists by name within a status model
func (r *statusRepository) ExistsByName(statusModelID uuid.UUID, name string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Status{}).Where("status_model_id = ? AND name = ?", statusModelID, name).Count(&count).Error
	return count > 0, err
}
