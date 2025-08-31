package repository

import (
	"product-requirements-management/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)



// statusModelRepository implements StatusModelRepository interface
type statusModelRepository struct {
	db *gorm.DB
}

// NewStatusModelRepository creates a new status model repository instance
func NewStatusModelRepository(db *gorm.DB) StatusModelRepository {
	return &statusModelRepository{db: db}
}

// Create creates a new status model
func (r *statusModelRepository) Create(statusModel *models.StatusModel) error {
	return r.db.Create(statusModel).Error
}

// GetByID retrieves a status model by ID with its statuses and transitions
func (r *statusModelRepository) GetByID(id uuid.UUID) (*models.StatusModel, error) {
	var statusModel models.StatusModel
	err := r.db.Preload("Statuses").Preload("Transitions").Where("id = ?", id).First(&statusModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &statusModel, nil
}

// GetByEntityTypeAndName retrieves a status model by entity type and name
func (r *statusModelRepository) GetByEntityTypeAndName(entityType models.EntityType, name string) (*models.StatusModel, error) {
	var statusModel models.StatusModel
	err := r.db.Preload("Statuses").Preload("Transitions").
		Where("entity_type = ? AND name = ?", entityType, name).First(&statusModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &statusModel, nil
}

// GetDefaultByEntityType retrieves the default status model for an entity type
func (r *statusModelRepository) GetDefaultByEntityType(entityType models.EntityType) (*models.StatusModel, error) {
	var statusModel models.StatusModel
	err := r.db.Preload("Statuses").Preload("Transitions").
		Where("entity_type = ? AND is_default = ?", entityType, true).First(&statusModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &statusModel, nil
}

// Update updates an existing status model
func (r *statusModelRepository) Update(statusModel *models.StatusModel) error {
	return r.db.Save(statusModel).Error
}

// Delete deletes a status model by ID
func (r *statusModelRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.StatusModel{}, id).Error
}

// List retrieves status models with filtering, ordering, and pagination
func (r *statusModelRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.StatusModel, error) {
	var statusModels []models.StatusModel
	
	query := r.db.Preload("Statuses").Preload("Transitions")
	
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
	
	err := query.Find(&statusModels).Error
	return statusModels, err
}

// ListByEntityType retrieves all status models for a specific entity type
func (r *statusModelRepository) ListByEntityType(entityType models.EntityType) ([]models.StatusModel, error) {
	var statusModels []models.StatusModel
	err := r.db.Preload("Statuses").Preload("Transitions").
		Where("entity_type = ?", entityType).Find(&statusModels).Error
	return statusModels, err
}

// Exists checks if a status model exists by ID
func (r *statusModelRepository) Exists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.StatusModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// ExistsByEntityTypeAndName checks if a status model exists by entity type and name
func (r *statusModelRepository) ExistsByEntityTypeAndName(entityType models.EntityType, name string) (bool, error) {
	var count int64
	err := r.db.Model(&models.StatusModel{}).Where("entity_type = ? AND name = ?", entityType, name).Count(&count).Error
	return count > 0, err
}