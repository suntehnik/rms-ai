package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
)

// requirementRelationshipRepository implements RequirementRelationshipRepository interface
type requirementRelationshipRepository struct {
	*BaseRepository[models.RequirementRelationship]
}

// NewRequirementRelationshipRepository creates a new requirement relationship repository instance
func NewRequirementRelationshipRepository(db *gorm.DB) RequirementRelationshipRepository {
	return &requirementRelationshipRepository{
		BaseRepository: NewBaseRepository[models.RequirementRelationship](db),
	}
}

// GetBySourceRequirement retrieves relationships by source requirement ID
func (r *requirementRelationshipRepository) GetBySourceRequirement(sourceID uuid.UUID) ([]models.RequirementRelationship, error) {
	var relationships []models.RequirementRelationship
	if err := r.GetDB().Where("source_requirement_id = ?", sourceID).Find(&relationships).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return relationships, nil
}

// GetByTargetRequirement retrieves relationships by target requirement ID
func (r *requirementRelationshipRepository) GetByTargetRequirement(targetID uuid.UUID) ([]models.RequirementRelationship, error) {
	var relationships []models.RequirementRelationship
	if err := r.GetDB().Where("target_requirement_id = ?", targetID).Find(&relationships).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return relationships, nil
}

// GetByRequirement retrieves all relationships for a requirement (both source and target)
func (r *requirementRelationshipRepository) GetByRequirement(requirementID uuid.UUID) ([]models.RequirementRelationship, error) {
	var relationships []models.RequirementRelationship
	if err := r.GetDB().Where("source_requirement_id = ? OR target_requirement_id = ?",
		requirementID, requirementID).Find(&relationships).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return relationships, nil
}

// GetByRequirementWithPagination retrieves relationships for a requirement with pagination
func (r *requirementRelationshipRepository) GetByRequirementWithPagination(requirementID uuid.UUID, limit, offset int) ([]models.RequirementRelationship, int64, error) {
	var relationships []models.RequirementRelationship
	var totalCount int64

	// Get total count
	if err := r.GetDB().Model(&models.RequirementRelationship{}).Where("source_requirement_id = ? OR target_requirement_id = ?",
		requirementID, requirementID).Count(&totalCount).Error; err != nil {
		return nil, 0, r.handleDBError(err)
	}

	// Get paginated results
	if err := r.GetDB().Where("source_requirement_id = ? OR target_requirement_id = ?",
		requirementID, requirementID).Limit(limit).Offset(offset).Find(&relationships).Error; err != nil {
		return nil, 0, r.handleDBError(err)
	}

	return relationships, totalCount, nil
}

// GetByType retrieves relationships by relationship type ID
func (r *requirementRelationshipRepository) GetByType(typeID uuid.UUID) ([]models.RequirementRelationship, error) {
	var relationships []models.RequirementRelationship
	if err := r.GetDB().Where("relationship_type_id = ?", typeID).Find(&relationships).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return relationships, nil
}

// ExistsRelationship checks if a specific relationship exists
func (r *requirementRelationshipRepository) ExistsRelationship(sourceID, targetID, typeID uuid.UUID) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.RequirementRelationship{}).
		Where("source_requirement_id = ? AND target_requirement_id = ? AND relationship_type_id = ?",
			sourceID, targetID, typeID).Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}
	return count > 0, nil
}
