package service

import (
	"errors"

	"github.com/google/uuid"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

// ConfigService defines the interface for configuration management operations
type ConfigService interface {
	// Requirement Type operations
	CreateRequirementType(req CreateRequirementTypeRequest) (*models.RequirementType, error)
	GetRequirementTypeByID(id uuid.UUID) (*models.RequirementType, error)
	GetRequirementTypeByName(name string) (*models.RequirementType, error)
	UpdateRequirementType(id uuid.UUID, req UpdateRequirementTypeRequest) (*models.RequirementType, error)
	DeleteRequirementType(id uuid.UUID, force bool) error
	ListRequirementTypes(filters RequirementTypeFilters) ([]models.RequirementType, error)

	// Relationship Type operations
	CreateRelationshipType(req CreateRelationshipTypeRequest) (*models.RelationshipType, error)
	GetRelationshipTypeByID(id uuid.UUID) (*models.RelationshipType, error)
	GetRelationshipTypeByName(name string) (*models.RelationshipType, error)
	UpdateRelationshipType(id uuid.UUID, req UpdateRelationshipTypeRequest) (*models.RelationshipType, error)
	DeleteRelationshipType(id uuid.UUID, force bool) error
	ListRelationshipTypes(filters RelationshipTypeFilters) ([]models.RelationshipType, error)

	// Validation operations
	ValidateRequirementType(typeID uuid.UUID) error
	ValidateRelationshipType(typeID uuid.UUID) error
}

// configService implements ConfigService interface
type configService struct {
	requirementTypeRepo     repository.RequirementTypeRepository
	relationshipTypeRepo    repository.RelationshipTypeRepository
	requirementRepo         repository.RequirementRepository
	requirementRelationRepo repository.RequirementRelationshipRepository
}

// NewConfigService creates a new configuration service instance
func NewConfigService(
	requirementTypeRepo repository.RequirementTypeRepository,
	relationshipTypeRepo repository.RelationshipTypeRepository,
	requirementRepo repository.RequirementRepository,
	requirementRelationRepo repository.RequirementRelationshipRepository,
) ConfigService {
	return &configService{
		requirementTypeRepo:     requirementTypeRepo,
		relationshipTypeRepo:    relationshipTypeRepo,
		requirementRepo:         requirementRepo,
		requirementRelationRepo: requirementRelationRepo,
	}
}

// Request and response types
type CreateRequirementTypeRequest struct {
	Name        string  `json:"name" binding:"required,max=255"`
	Description *string `json:"description,omitempty"`
}

type UpdateRequirementTypeRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,max=255"`
	Description *string `json:"description,omitempty"`
}

type RequirementTypeFilters struct {
	OrderBy string `json:"order_by,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Offset  int    `json:"offset,omitempty"`
}

type CreateRelationshipTypeRequest struct {
	Name        string  `json:"name" binding:"required,max=255"`
	Description *string `json:"description,omitempty"`
}

type UpdateRelationshipTypeRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,max=255"`
	Description *string `json:"description,omitempty"`
}

type RelationshipTypeFilters struct {
	OrderBy string `json:"order_by,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Offset  int    `json:"offset,omitempty"`
}

// Service errors
var (
	ErrRequirementTypeNameExists    = errors.New("requirement type name already exists")
	ErrRequirementTypeHasRequirements = errors.New("requirement type has associated requirements")
	ErrRelationshipTypeNameExists   = errors.New("relationship type name already exists")
	ErrRelationshipTypeHasRelationships = errors.New("relationship type has associated relationships")
)

// Requirement Type operations

// CreateRequirementType creates a new requirement type
func (s *configService) CreateRequirementType(req CreateRequirementTypeRequest) (*models.RequirementType, error) {
	// Check if name already exists
	exists, err := s.requirementTypeRepo.ExistsByName(req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrRequirementTypeNameExists
	}

	requirementType := &models.RequirementType{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.requirementTypeRepo.Create(requirementType); err != nil {
		return nil, err
	}

	return requirementType, nil
}

// GetRequirementTypeByID retrieves a requirement type by ID
func (s *configService) GetRequirementTypeByID(id uuid.UUID) (*models.RequirementType, error) {
	requirementType, err := s.requirementTypeRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRequirementTypeNotFound
		}
		return nil, err
	}
	return requirementType, nil
}

// GetRequirementTypeByName retrieves a requirement type by name
func (s *configService) GetRequirementTypeByName(name string) (*models.RequirementType, error) {
	requirementType, err := s.requirementTypeRepo.GetByName(name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRequirementTypeNotFound
		}
		return nil, err
	}
	return requirementType, nil
}

// UpdateRequirementType updates an existing requirement type
func (s *configService) UpdateRequirementType(id uuid.UUID, req UpdateRequirementTypeRequest) (*models.RequirementType, error) {
	requirementType, err := s.requirementTypeRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRequirementTypeNotFound
		}
		return nil, err
	}

	// Check if new name already exists (if name is being changed)
	if req.Name != nil && *req.Name != requirementType.Name {
		exists, err := s.requirementTypeRepo.ExistsByName(*req.Name)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrRequirementTypeNameExists
		}
		requirementType.Name = *req.Name
	}

	if req.Description != nil {
		requirementType.Description = req.Description
	}

	if err := s.requirementTypeRepo.Update(requirementType); err != nil {
		return nil, err
	}

	return requirementType, nil
}

// DeleteRequirementType deletes a requirement type
func (s *configService) DeleteRequirementType(id uuid.UUID, force bool) error {
	_, err := s.requirementTypeRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrRequirementTypeNotFound
		}
		return err
	}

	// Check if requirement type has associated requirements
	requirements, err := s.requirementRepo.GetByType(id)
	if err != nil {
		return err
	}

	if len(requirements) > 0 && !force {
		return ErrRequirementTypeHasRequirements
	}

	// If force is true, we need to handle the requirements
	if force && len(requirements) > 0 {
		// For now, we'll prevent deletion even with force if there are requirements
		// This follows the constraint that requirement types should be restricted from deletion
		return ErrRequirementTypeHasRequirements
	}

	return s.requirementTypeRepo.Delete(id)
}

// ListRequirementTypes lists requirement types with optional filtering
func (s *configService) ListRequirementTypes(filters RequirementTypeFilters) ([]models.RequirementType, error) {
	filterMap := make(map[string]interface{})
	
	orderBy := "name"
	if filters.OrderBy != "" {
		orderBy = filters.OrderBy
	}

	limit := 100 // Default limit
	if filters.Limit > 0 {
		limit = filters.Limit
	}

	offset := 0
	if filters.Offset > 0 {
		offset = filters.Offset
	}

	return s.requirementTypeRepo.List(filterMap, orderBy, limit, offset)
}

// Relationship Type operations

// CreateRelationshipType creates a new relationship type
func (s *configService) CreateRelationshipType(req CreateRelationshipTypeRequest) (*models.RelationshipType, error) {
	// Check if name already exists
	exists, err := s.relationshipTypeRepo.ExistsByName(req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrRelationshipTypeNameExists
	}

	relationshipType := &models.RelationshipType{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.relationshipTypeRepo.Create(relationshipType); err != nil {
		return nil, err
	}

	return relationshipType, nil
}

// GetRelationshipTypeByID retrieves a relationship type by ID
func (s *configService) GetRelationshipTypeByID(id uuid.UUID) (*models.RelationshipType, error) {
	relationshipType, err := s.relationshipTypeRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRelationshipTypeNotFound
		}
		return nil, err
	}
	return relationshipType, nil
}

// GetRelationshipTypeByName retrieves a relationship type by name
func (s *configService) GetRelationshipTypeByName(name string) (*models.RelationshipType, error) {
	relationshipType, err := s.relationshipTypeRepo.GetByName(name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRelationshipTypeNotFound
		}
		return nil, err
	}
	return relationshipType, nil
}

// UpdateRelationshipType updates an existing relationship type
func (s *configService) UpdateRelationshipType(id uuid.UUID, req UpdateRelationshipTypeRequest) (*models.RelationshipType, error) {
	relationshipType, err := s.relationshipTypeRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRelationshipTypeNotFound
		}
		return nil, err
	}

	// Check if new name already exists (if name is being changed)
	if req.Name != nil && *req.Name != relationshipType.Name {
		exists, err := s.relationshipTypeRepo.ExistsByName(*req.Name)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrRelationshipTypeNameExists
		}
		relationshipType.Name = *req.Name
	}

	if req.Description != nil {
		relationshipType.Description = req.Description
	}

	if err := s.relationshipTypeRepo.Update(relationshipType); err != nil {
		return nil, err
	}

	return relationshipType, nil
}

// DeleteRelationshipType deletes a relationship type
func (s *configService) DeleteRelationshipType(id uuid.UUID, force bool) error {
	_, err := s.relationshipTypeRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrRelationshipTypeNotFound
		}
		return err
	}

	// Check if relationship type has associated relationships
	relationships, err := s.requirementRelationRepo.GetByType(id)
	if err != nil {
		return err
	}

	if len(relationships) > 0 && !force {
		return ErrRelationshipTypeHasRelationships
	}

	// If force is true, we need to handle the relationships
	if force && len(relationships) > 0 {
		// For now, we'll prevent deletion even with force if there are relationships
		// This follows the constraint that relationship types should be restricted from deletion
		return ErrRelationshipTypeHasRelationships
	}

	return s.relationshipTypeRepo.Delete(id)
}

// ListRelationshipTypes lists relationship types with optional filtering
func (s *configService) ListRelationshipTypes(filters RelationshipTypeFilters) ([]models.RelationshipType, error) {
	filterMap := make(map[string]interface{})
	
	orderBy := "name"
	if filters.OrderBy != "" {
		orderBy = filters.OrderBy
	}

	limit := 100 // Default limit
	if filters.Limit > 0 {
		limit = filters.Limit
	}

	offset := 0
	if filters.Offset > 0 {
		offset = filters.Offset
	}

	return s.relationshipTypeRepo.List(filterMap, orderBy, limit, offset)
}

// Validation operations

// ValidateRequirementType validates that a requirement type exists
func (s *configService) ValidateRequirementType(typeID uuid.UUID) error {
	exists, err := s.requirementTypeRepo.Exists(typeID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrRequirementTypeNotFound
	}
	return nil
}

// ValidateRelationshipType validates that a relationship type exists
func (s *configService) ValidateRelationshipType(typeID uuid.UUID) error {
	exists, err := s.relationshipTypeRepo.Exists(typeID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrRelationshipTypeNotFound
	}
	return nil
}