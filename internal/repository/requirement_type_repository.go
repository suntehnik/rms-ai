package repository

import (
	"errors"

	"gorm.io/gorm"
	
	"product-requirements-management/internal/models"
)

// requirementTypeRepository implements RequirementTypeRepository interface
type requirementTypeRepository struct {
	*BaseRepository[models.RequirementType]
}

// NewRequirementTypeRepository creates a new requirement type repository instance
func NewRequirementTypeRepository(db *gorm.DB) RequirementTypeRepository {
	return &requirementTypeRepository{
		BaseRepository: NewBaseRepository[models.RequirementType](db),
	}
}

// GetByName retrieves a requirement type by name
func (r *requirementTypeRepository) GetByName(name string) (*models.RequirementType, error) {
	var reqType models.RequirementType
	if err := r.GetDB().Where("name = ?", name).First(&reqType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, r.handleDBError(err)
	}
	return &reqType, nil
}

// ExistsByName checks if a requirement type exists by name
func (r *requirementTypeRepository) ExistsByName(name string) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.RequirementType{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}
	return count > 0, nil
}