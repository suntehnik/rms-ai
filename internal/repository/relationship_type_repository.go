package repository

import (
	"errors"

	"gorm.io/gorm"

	"product-requirements-management/internal/models"
)

// relationshipTypeRepository implements RelationshipTypeRepository interface
type relationshipTypeRepository struct {
	*BaseRepository[models.RelationshipType]
}

// NewRelationshipTypeRepository creates a new relationship type repository instance
func NewRelationshipTypeRepository(db *gorm.DB) RelationshipTypeRepository {
	return &relationshipTypeRepository{
		BaseRepository: NewBaseRepository[models.RelationshipType](db),
	}
}

// GetByName retrieves a relationship type by name
func (r *relationshipTypeRepository) GetByName(name string) (*models.RelationshipType, error) {
	var relType models.RelationshipType
	if err := r.GetDB().Where("name = ?", name).First(&relType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, r.handleDBError(err)
	}
	return &relType, nil
}

// ExistsByName checks if a relationship type exists by name
func (r *relationshipTypeRepository) ExistsByName(name string) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.RelationshipType{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}
	return count > 0, nil
}
