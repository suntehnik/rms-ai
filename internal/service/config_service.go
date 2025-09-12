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

	// Status Model operations
	CreateStatusModel(req CreateStatusModelRequest) (*models.StatusModel, error)
	GetStatusModelByID(id uuid.UUID) (*models.StatusModel, error)
	GetStatusModelByEntityTypeAndName(entityType models.EntityType, name string) (*models.StatusModel, error)
	GetDefaultStatusModelByEntityType(entityType models.EntityType) (*models.StatusModel, error)
	UpdateStatusModel(id uuid.UUID, req UpdateStatusModelRequest) (*models.StatusModel, error)
	DeleteStatusModel(id uuid.UUID, force bool) error
	ListStatusModels(filters StatusModelFilters) ([]models.StatusModel, error)
	ListStatusModelsByEntityType(entityType models.EntityType) ([]models.StatusModel, error)

	// Status operations
	CreateStatus(req CreateStatusRequest) (*models.Status, error)
	GetStatusByID(id uuid.UUID) (*models.Status, error)
	UpdateStatus(id uuid.UUID, req UpdateStatusRequest) (*models.Status, error)
	DeleteStatus(id uuid.UUID, force bool) error
	ListStatusesByModel(statusModelID uuid.UUID) ([]models.Status, error)

	// Status Transition operations
	CreateStatusTransition(req CreateStatusTransitionRequest) (*models.StatusTransition, error)
	GetStatusTransitionByID(id uuid.UUID) (*models.StatusTransition, error)
	UpdateStatusTransition(id uuid.UUID, req UpdateStatusTransitionRequest) (*models.StatusTransition, error)
	DeleteStatusTransition(id uuid.UUID) error
	ListStatusTransitionsByModel(statusModelID uuid.UUID) ([]models.StatusTransition, error)

	// Validation operations
	ValidateRequirementType(typeID uuid.UUID) error
	ValidateRelationshipType(typeID uuid.UUID) error
	ValidateStatusTransition(entityType models.EntityType, fromStatus, toStatus string) error
}

// configService implements ConfigService interface
type configService struct {
	requirementTypeRepo     repository.RequirementTypeRepository
	relationshipTypeRepo    repository.RelationshipTypeRepository
	requirementRepo         repository.RequirementRepository
	requirementRelationRepo repository.RequirementRelationshipRepository
	statusModelRepo         repository.StatusModelRepository
	statusRepo              repository.StatusRepository
	statusTransitionRepo    repository.StatusTransitionRepository
}

// NewConfigService creates a new configuration service instance
func NewConfigService(
	requirementTypeRepo repository.RequirementTypeRepository,
	relationshipTypeRepo repository.RelationshipTypeRepository,
	requirementRepo repository.RequirementRepository,
	requirementRelationRepo repository.RequirementRelationshipRepository,
	statusModelRepo repository.StatusModelRepository,
	statusRepo repository.StatusRepository,
	statusTransitionRepo repository.StatusTransitionRepository,
) ConfigService {
	return &configService{
		requirementTypeRepo:     requirementTypeRepo,
		relationshipTypeRepo:    relationshipTypeRepo,
		requirementRepo:         requirementRepo,
		requirementRelationRepo: requirementRelationRepo,
		statusModelRepo:         statusModelRepo,
		statusRepo:              statusRepo,
		statusTransitionRepo:    statusTransitionRepo,
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

type CreateStatusModelRequest struct {
	EntityType  models.EntityType `json:"entity_type" binding:"required"`
	Name        string            `json:"name" binding:"required,max=255"`
	Description *string           `json:"description,omitempty"`
	IsDefault   bool              `json:"is_default,omitempty"`
}

type UpdateStatusModelRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,max=255"`
	Description *string `json:"description,omitempty"`
	IsDefault   *bool   `json:"is_default,omitempty"`
}

type StatusModelFilters struct {
	EntityType models.EntityType `json:"entity_type,omitempty"`
	OrderBy    string            `json:"order_by,omitempty"`
	Limit      int               `json:"limit,omitempty"`
	Offset     int               `json:"offset,omitempty"`
}

type CreateStatusRequest struct {
	StatusModelID uuid.UUID `json:"status_model_id" binding:"required"`
	Name          string    `json:"name" binding:"required,max=255"`
	Description   *string   `json:"description,omitempty"`
	Color         *string   `json:"color,omitempty"`
	IsInitial     bool      `json:"is_initial,omitempty"`
	IsFinal       bool      `json:"is_final,omitempty"`
	Order         int       `json:"order,omitempty"`
}

type UpdateStatusRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,max=255"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
	IsInitial   *bool   `json:"is_initial,omitempty"`
	IsFinal     *bool   `json:"is_final,omitempty"`
	Order       *int    `json:"order,omitempty"`
}

type CreateStatusTransitionRequest struct {
	StatusModelID uuid.UUID `json:"status_model_id" binding:"required"`
	FromStatusID  uuid.UUID `json:"from_status_id" binding:"required"`
	ToStatusID    uuid.UUID `json:"to_status_id" binding:"required"`
	Name          *string   `json:"name,omitempty"`
	Description   *string   `json:"description,omitempty"`
}

type UpdateStatusTransitionRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// Config service specific errors
var (
	ErrRequirementTypeNameExists        = errors.New("requirement type name already exists")
	ErrRequirementTypeHasRequirements   = errors.New("requirement type has associated requirements")
	ErrRelationshipTypeNameExists       = errors.New("relationship type name already exists")
	ErrRelationshipTypeHasRelationships = errors.New("relationship type has associated relationships")
	ErrStatusModelNameExists            = errors.New("status model name already exists for this entity type")
	ErrStatusModelNotFound              = errors.New("status model not found")
	ErrStatusNotFound                   = errors.New("status not found")
	ErrStatusTransitionNotFound         = errors.New("status transition not found")
	ErrStatusNameExists                 = errors.New("status name already exists in this model")
	ErrTransitionExists                 = errors.New("status transition already exists")
	ErrInvalidEntityType                = errors.New("invalid entity type")
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

// Status Model operations

// CreateStatusModel creates a new status model
func (s *configService) CreateStatusModel(req CreateStatusModelRequest) (*models.StatusModel, error) {
	// Validate entity type
	if !models.IsValidEntityType(req.EntityType) {
		return nil, ErrInvalidEntityType
	}

	// Check if name already exists for this entity type
	exists, err := s.statusModelRepo.ExistsByEntityTypeAndName(req.EntityType, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrStatusModelNameExists
	}

	statusModel := &models.StatusModel{
		EntityType:  req.EntityType,
		Name:        req.Name,
		Description: req.Description,
		IsDefault:   req.IsDefault,
	}

	if err := s.statusModelRepo.Create(statusModel); err != nil {
		return nil, err
	}

	return statusModel, nil
}

// GetStatusModelByID retrieves a status model by ID
func (s *configService) GetStatusModelByID(id uuid.UUID) (*models.StatusModel, error) {
	statusModel, err := s.statusModelRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrStatusModelNotFound
		}
		return nil, err
	}
	return statusModel, nil
}

// GetStatusModelByEntityTypeAndName retrieves a status model by entity type and name
func (s *configService) GetStatusModelByEntityTypeAndName(entityType models.EntityType, name string) (*models.StatusModel, error) {
	statusModel, err := s.statusModelRepo.GetByEntityTypeAndName(entityType, name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrStatusModelNotFound
		}
		return nil, err
	}
	return statusModel, nil
}

// GetDefaultStatusModelByEntityType retrieves the default status model for an entity type
func (s *configService) GetDefaultStatusModelByEntityType(entityType models.EntityType) (*models.StatusModel, error) {
	statusModel, err := s.statusModelRepo.GetDefaultByEntityType(entityType)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrStatusModelNotFound
		}
		return nil, err
	}
	return statusModel, nil
}

// UpdateStatusModel updates an existing status model
func (s *configService) UpdateStatusModel(id uuid.UUID, req UpdateStatusModelRequest) (*models.StatusModel, error) {
	statusModel, err := s.statusModelRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrStatusModelNotFound
		}
		return nil, err
	}

	// Check if new name already exists (if name is being changed)
	if req.Name != nil && *req.Name != statusModel.Name {
		exists, err := s.statusModelRepo.ExistsByEntityTypeAndName(statusModel.EntityType, *req.Name)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrStatusModelNameExists
		}
		statusModel.Name = *req.Name
	}

	if req.Description != nil {
		statusModel.Description = req.Description
	}

	if req.IsDefault != nil {
		statusModel.IsDefault = *req.IsDefault
	}

	if err := s.statusModelRepo.Update(statusModel); err != nil {
		return nil, err
	}

	return statusModel, nil
}

// DeleteStatusModel deletes a status model
func (s *configService) DeleteStatusModel(id uuid.UUID, force bool) error {
	_, err := s.statusModelRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrStatusModelNotFound
		}
		return err
	}

	// For now, we'll allow deletion of status models
	// In a production system, you might want to check if entities are using this status model
	return s.statusModelRepo.Delete(id)
}

// ListStatusModels lists status models with optional filtering
func (s *configService) ListStatusModels(filters StatusModelFilters) ([]models.StatusModel, error) {
	filterMap := make(map[string]interface{})

	if filters.EntityType != "" {
		filterMap["entity_type"] = filters.EntityType
	}

	orderBy := "entity_type, name"
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

	return s.statusModelRepo.List(filterMap, orderBy, limit, offset)
}

// ListStatusModelsByEntityType lists status models for a specific entity type
func (s *configService) ListStatusModelsByEntityType(entityType models.EntityType) ([]models.StatusModel, error) {
	return s.statusModelRepo.ListByEntityType(entityType)
}

// Status operations

// CreateStatus creates a new status
func (s *configService) CreateStatus(req CreateStatusRequest) (*models.Status, error) {
	// Validate that the status model exists
	_, err := s.statusModelRepo.GetByID(req.StatusModelID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrStatusModelNotFound
		}
		return nil, err
	}

	// Check if name already exists in this status model
	exists, err := s.statusRepo.ExistsByName(req.StatusModelID, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrStatusNameExists
	}

	status := &models.Status{
		StatusModelID: req.StatusModelID,
		Name:          req.Name,
		Description:   req.Description,
		Color:         req.Color,
		IsInitial:     req.IsInitial,
		IsFinal:       req.IsFinal,
		Order:         req.Order,
	}

	if err := s.statusRepo.Create(status); err != nil {
		return nil, err
	}

	return status, nil
}

// GetStatusByID retrieves a status by ID
func (s *configService) GetStatusByID(id uuid.UUID) (*models.Status, error) {
	status, err := s.statusRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrStatusNotFound
		}
		return nil, err
	}
	return status, nil
}

// UpdateStatus updates an existing status
func (s *configService) UpdateStatus(id uuid.UUID, req UpdateStatusRequest) (*models.Status, error) {
	status, err := s.statusRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrStatusNotFound
		}
		return nil, err
	}

	// Check if new name already exists (if name is being changed)
	if req.Name != nil && *req.Name != status.Name {
		exists, err := s.statusRepo.ExistsByName(status.StatusModelID, *req.Name)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrStatusNameExists
		}
		status.Name = *req.Name
	}

	if req.Description != nil {
		status.Description = req.Description
	}

	if req.Color != nil {
		status.Color = req.Color
	}

	if req.IsInitial != nil {
		status.IsInitial = *req.IsInitial
	}

	if req.IsFinal != nil {
		status.IsFinal = *req.IsFinal
	}

	if req.Order != nil {
		status.Order = *req.Order
	}

	if err := s.statusRepo.Update(status); err != nil {
		return nil, err
	}

	return status, nil
}

// DeleteStatus deletes a status
func (s *configService) DeleteStatus(id uuid.UUID, force bool) error {
	_, err := s.statusRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrStatusNotFound
		}
		return err
	}

	// For now, we'll allow deletion of statuses
	// In a production system, you might want to check if entities are using this status
	return s.statusRepo.Delete(id)
}

// ListStatusesByModel lists statuses for a specific status model
func (s *configService) ListStatusesByModel(statusModelID uuid.UUID) ([]models.Status, error) {
	return s.statusRepo.GetByStatusModelID(statusModelID)
}

// Status Transition operations

// CreateStatusTransition creates a new status transition
func (s *configService) CreateStatusTransition(req CreateStatusTransitionRequest) (*models.StatusTransition, error) {
	// Validate that the status model exists
	_, err := s.statusModelRepo.GetByID(req.StatusModelID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrStatusModelNotFound
		}
		return nil, err
	}

	// Validate that both statuses exist and belong to the same status model
	fromStatus, err := s.statusRepo.GetByID(req.FromStatusID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrStatusNotFound
		}
		return nil, err
	}

	toStatus, err := s.statusRepo.GetByID(req.ToStatusID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrStatusNotFound
		}
		return nil, err
	}

	if fromStatus.StatusModelID != req.StatusModelID || toStatus.StatusModelID != req.StatusModelID {
		return nil, ErrInvalidStatusTransition
	}

	// Check if transition already exists
	exists, err := s.statusTransitionRepo.ExistsByTransition(req.StatusModelID, req.FromStatusID, req.ToStatusID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrTransitionExists
	}

	transition := &models.StatusTransition{
		StatusModelID: req.StatusModelID,
		FromStatusID:  req.FromStatusID,
		ToStatusID:    req.ToStatusID,
		Name:          req.Name,
		Description:   req.Description,
	}

	if err := s.statusTransitionRepo.Create(transition); err != nil {
		return nil, err
	}

	return transition, nil
}

// GetStatusTransitionByID retrieves a status transition by ID
func (s *configService) GetStatusTransitionByID(id uuid.UUID) (*models.StatusTransition, error) {
	transition, err := s.statusTransitionRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrStatusTransitionNotFound
		}
		return nil, err
	}
	return transition, nil
}

// UpdateStatusTransition updates an existing status transition
func (s *configService) UpdateStatusTransition(id uuid.UUID, req UpdateStatusTransitionRequest) (*models.StatusTransition, error) {
	transition, err := s.statusTransitionRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrStatusTransitionNotFound
		}
		return nil, err
	}

	if req.Name != nil {
		transition.Name = req.Name
	}

	if req.Description != nil {
		transition.Description = req.Description
	}

	if err := s.statusTransitionRepo.Update(transition); err != nil {
		return nil, err
	}

	return transition, nil
}

// DeleteStatusTransition deletes a status transition
func (s *configService) DeleteStatusTransition(id uuid.UUID) error {
	_, err := s.statusTransitionRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrStatusTransitionNotFound
		}
		return err
	}

	return s.statusTransitionRepo.Delete(id)
}

// ListStatusTransitionsByModel lists status transitions for a specific status model
func (s *configService) ListStatusTransitionsByModel(statusModelID uuid.UUID) ([]models.StatusTransition, error) {
	return s.statusTransitionRepo.GetByStatusModelID(statusModelID)
}

// ValidateStatusTransition validates that a status transition is allowed
func (s *configService) ValidateStatusTransition(entityType models.EntityType, fromStatus, toStatus string) error {
	// Get the default status model for the entity type
	statusModel, err := s.statusModelRepo.GetDefaultByEntityType(entityType)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// If no status model is found, allow all transitions (default behavior)
			return nil
		}
		return err
	}

	// Get the from and to status objects
	fromStatusObj, err := s.statusRepo.GetByName(statusModel.ID, fromStatus)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInvalidStatusTransition
		}
		return err
	}

	toStatusObj, err := s.statusRepo.GetByName(statusModel.ID, toStatus)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInvalidStatusTransition
		}
		return err
	}

	// Check if the transition is allowed
	if !statusModel.CanTransitionTo(fromStatusObj.ID, toStatusObj.ID) {
		return ErrInvalidStatusTransition
	}

	return nil
}
