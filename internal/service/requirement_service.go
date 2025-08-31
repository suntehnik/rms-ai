package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

var (
	ErrRequirementNotFound           = errors.New("requirement not found")
	ErrRequirementHasRelationships   = errors.New("requirement has associated relationships and cannot be deleted")
	ErrInvalidRequirementStatus      = errors.New("invalid requirement status")

	ErrCircularRelationship          = errors.New("circular relationship detected")
	ErrDuplicateRelationship         = errors.New("relationship already exists")
)

// RequirementService defines the interface for requirement business logic
type RequirementService interface {
	CreateRequirement(req CreateRequirementRequest) (*models.Requirement, error)
	GetRequirementByID(id uuid.UUID) (*models.Requirement, error)
	GetRequirementByReferenceID(referenceID string) (*models.Requirement, error)
	UpdateRequirement(id uuid.UUID, req UpdateRequirementRequest) (*models.Requirement, error)
	DeleteRequirement(id uuid.UUID, force bool) error
	ListRequirements(filters RequirementFilters) ([]models.Requirement, error)
	GetRequirementWithRelationships(id uuid.UUID) (*models.Requirement, error)
	GetRequirementsByUserStory(userStoryID uuid.UUID) ([]models.Requirement, error)
	ChangeRequirementStatus(id uuid.UUID, newStatus models.RequirementStatus) (*models.Requirement, error)
	AssignRequirement(id uuid.UUID, assigneeID uuid.UUID) (*models.Requirement, error)
	CreateRelationship(req CreateRelationshipRequest) (*models.RequirementRelationship, error)
	DeleteRelationship(id uuid.UUID) error
	GetRelationshipsByRequirement(requirementID uuid.UUID) ([]models.RequirementRelationship, error)
	SearchRequirements(searchText string) ([]models.Requirement, error)
}

// CreateRequirementRequest represents the request to create a requirement
type CreateRequirementRequest struct {
	UserStoryID          uuid.UUID                `json:"user_story_id" binding:"required"`
	AcceptanceCriteriaID *uuid.UUID               `json:"acceptance_criteria_id,omitempty"`
	CreatorID            uuid.UUID                `json:"creator_id" binding:"required"`
	AssigneeID           *uuid.UUID               `json:"assignee_id,omitempty"`
	Priority             models.Priority          `json:"priority" binding:"required,min=1,max=4"`
	TypeID               uuid.UUID                `json:"type_id" binding:"required"`
	Title                string                   `json:"title" binding:"required,max=500"`
	Description          *string                  `json:"description,omitempty"`
}

// UpdateRequirementRequest represents the request to update a requirement
type UpdateRequirementRequest struct {
	AcceptanceCriteriaID *uuid.UUID                `json:"acceptance_criteria_id,omitempty"`
	AssigneeID           *uuid.UUID                `json:"assignee_id,omitempty"`
	Priority             *models.Priority          `json:"priority,omitempty"`
	Status               *models.RequirementStatus `json:"status,omitempty"`
	TypeID               *uuid.UUID                `json:"type_id,omitempty"`
	Title                *string                   `json:"title,omitempty"`
	Description          *string                   `json:"description,omitempty"`
}

// RequirementFilters represents filters for listing requirements
type RequirementFilters struct {
	UserStoryID          *uuid.UUID                `json:"user_story_id,omitempty"`
	AcceptanceCriteriaID *uuid.UUID                `json:"acceptance_criteria_id,omitempty"`
	CreatorID            *uuid.UUID                `json:"creator_id,omitempty"`
	AssigneeID           *uuid.UUID                `json:"assignee_id,omitempty"`
	Status               *models.RequirementStatus `json:"status,omitempty"`
	Priority             *models.Priority          `json:"priority,omitempty"`
	TypeID               *uuid.UUID                `json:"type_id,omitempty"`
	OrderBy              string                    `json:"order_by,omitempty"`
	Limit                int                       `json:"limit,omitempty"`
	Offset               int                       `json:"offset,omitempty"`
}

// CreateRelationshipRequest represents the request to create a requirement relationship
type CreateRelationshipRequest struct {
	SourceRequirementID  uuid.UUID `json:"source_requirement_id" binding:"required"`
	TargetRequirementID  uuid.UUID `json:"target_requirement_id" binding:"required"`
	RelationshipTypeID   uuid.UUID `json:"relationship_type_id" binding:"required"`
	CreatedBy            uuid.UUID `json:"created_by" binding:"required"`
}

// requirementService implements RequirementService interface
type requirementService struct {
	requirementRepo             repository.RequirementRepository
	requirementTypeRepo         repository.RequirementTypeRepository
	relationshipTypeRepo        repository.RelationshipTypeRepository
	requirementRelationshipRepo repository.RequirementRelationshipRepository
	userStoryRepo               repository.UserStoryRepository
	acceptanceCriteriaRepo      repository.AcceptanceCriteriaRepository
	userRepo                    repository.UserRepository
}

// NewRequirementService creates a new requirement service instance
func NewRequirementService(
	requirementRepo repository.RequirementRepository,
	requirementTypeRepo repository.RequirementTypeRepository,
	relationshipTypeRepo repository.RelationshipTypeRepository,
	requirementRelationshipRepo repository.RequirementRelationshipRepository,
	userStoryRepo repository.UserStoryRepository,
	acceptanceCriteriaRepo repository.AcceptanceCriteriaRepository,
	userRepo repository.UserRepository,
) RequirementService {
	return &requirementService{
		requirementRepo:             requirementRepo,
		requirementTypeRepo:         requirementTypeRepo,
		relationshipTypeRepo:        relationshipTypeRepo,
		requirementRelationshipRepo: requirementRelationshipRepo,
		userStoryRepo:               userStoryRepo,
		acceptanceCriteriaRepo:      acceptanceCriteriaRepo,
		userRepo:                    userRepo,
	}
}

// CreateRequirement creates a new requirement
func (s *requirementService) CreateRequirement(req CreateRequirementRequest) (*models.Requirement, error) {
	// Validate priority
	if req.Priority < models.PriorityCritical || req.Priority > models.PriorityLow {
		return nil, ErrInvalidPriority
	}

	// Validate user story exists
	if exists, err := s.userStoryRepo.Exists(req.UserStoryID); err != nil {
		return nil, fmt.Errorf("failed to check user story existence: %w", err)
	} else if !exists {
		return nil, ErrUserStoryNotFound
	}

	// Validate requirement type exists
	if exists, err := s.requirementTypeRepo.Exists(req.TypeID); err != nil {
		return nil, fmt.Errorf("failed to check requirement type existence: %w", err)
	} else if !exists {
		return nil, ErrRequirementTypeNotFound
	}

	// Validate creator exists
	if exists, err := s.userRepo.Exists(req.CreatorID); err != nil {
		return nil, fmt.Errorf("failed to check creator existence: %w", err)
	} else if !exists {
		return nil, ErrUserNotFound
	}

	// Set assignee to creator if not specified
	assigneeID := req.CreatorID
	if req.AssigneeID != nil {
		assigneeID = *req.AssigneeID
		// Validate assignee exists
		if exists, err := s.userRepo.Exists(assigneeID); err != nil {
			return nil, fmt.Errorf("failed to check assignee existence: %w", err)
		} else if !exists {
			return nil, ErrUserNotFound
		}
	}

	// Validate acceptance criteria if provided
	if req.AcceptanceCriteriaID != nil {
		if exists, err := s.acceptanceCriteriaRepo.Exists(*req.AcceptanceCriteriaID); err != nil {
			return nil, fmt.Errorf("failed to check acceptance criteria existence: %w", err)
		} else if !exists {
			return nil, ErrAcceptanceCriteriaNotFound
		}
	}

	requirement := &models.Requirement{
		ID:                   uuid.New(),
		UserStoryID:          req.UserStoryID,
		AcceptanceCriteriaID: req.AcceptanceCriteriaID,
		CreatorID:            req.CreatorID,
		AssigneeID:           assigneeID,
		Priority:             req.Priority,
		Status:               models.RequirementStatusDraft, // Default status
		TypeID:               req.TypeID,
		Title:                req.Title,
		Description:          req.Description,
	}

	if err := s.requirementRepo.Create(requirement); err != nil {
		return nil, fmt.Errorf("failed to create requirement: %w", err)
	}

	return requirement, nil
}

// GetRequirementByID retrieves a requirement by its ID
func (s *requirementService) GetRequirementByID(id uuid.UUID) (*models.Requirement, error) {
	requirement, err := s.requirementRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRequirementNotFound
		}
		return nil, fmt.Errorf("failed to get requirement: %w", err)
	}
	return requirement, nil
}

// GetRequirementByReferenceID retrieves a requirement by its reference ID
func (s *requirementService) GetRequirementByReferenceID(referenceID string) (*models.Requirement, error) {
	requirement, err := s.requirementRepo.GetByReferenceID(referenceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRequirementNotFound
		}
		return nil, fmt.Errorf("failed to get requirement: %w", err)
	}
	return requirement, nil
}

// UpdateRequirement updates an existing requirement
func (s *requirementService) UpdateRequirement(id uuid.UUID, req UpdateRequirementRequest) (*models.Requirement, error) {
	requirement, err := s.requirementRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRequirementNotFound
		}
		return nil, fmt.Errorf("failed to get requirement: %w", err)
	}

	// Update fields if provided
	if req.AcceptanceCriteriaID != nil {
		// Validate acceptance criteria exists
		if exists, err := s.acceptanceCriteriaRepo.Exists(*req.AcceptanceCriteriaID); err != nil {
			return nil, fmt.Errorf("failed to check acceptance criteria existence: %w", err)
		} else if !exists {
			return nil, ErrAcceptanceCriteriaNotFound
		}
		requirement.AcceptanceCriteriaID = req.AcceptanceCriteriaID
	}

	if req.AssigneeID != nil {
		// Validate assignee exists
		if exists, err := s.userRepo.Exists(*req.AssigneeID); err != nil {
			return nil, fmt.Errorf("failed to check assignee existence: %w", err)
		} else if !exists {
			return nil, ErrUserNotFound
		}
		requirement.AssigneeID = *req.AssigneeID
	}

	if req.Priority != nil {
		if *req.Priority < models.PriorityCritical || *req.Priority > models.PriorityLow {
			return nil, ErrInvalidPriority
		}
		requirement.Priority = *req.Priority
	}

	if req.Status != nil {
		if !requirement.IsValidStatus(*req.Status) {
			return nil, ErrInvalidRequirementStatus
		}
		if !requirement.CanTransitionTo(*req.Status) {
			return nil, ErrInvalidStatusTransition
		}
		requirement.Status = *req.Status
	}

	if req.TypeID != nil {
		// Validate requirement type exists
		if exists, err := s.requirementTypeRepo.Exists(*req.TypeID); err != nil {
			return nil, fmt.Errorf("failed to check requirement type existence: %w", err)
		} else if !exists {
			return nil, ErrRequirementTypeNotFound
		}
		requirement.TypeID = *req.TypeID
	}

	if req.Title != nil {
		requirement.Title = *req.Title
	}

	if req.Description != nil {
		requirement.Description = req.Description
	}

	if err := s.requirementRepo.Update(requirement); err != nil {
		return nil, fmt.Errorf("failed to update requirement: %w", err)
	}

	return requirement, nil
}

// DeleteRequirement deletes a requirement with dependency validation
func (s *requirementService) DeleteRequirement(id uuid.UUID, force bool) error {
	// Check if requirement exists
	_, err := s.requirementRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrRequirementNotFound
		}
		return fmt.Errorf("failed to get requirement: %w", err)
	}

	// Check for relationships unless force delete
	if !force {
		hasRelationships, err := s.requirementRepo.HasRelationships(id)
		if err != nil {
			return fmt.Errorf("failed to check relationships: %w", err)
		}
		if hasRelationships {
			return ErrRequirementHasRelationships
		}
	}

	// Delete the requirement (cascade will handle relationships if force=true)
	if err := s.requirementRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete requirement: %w", err)
	}

	return nil
}

// ListRequirements retrieves requirements with optional filtering
func (s *requirementService) ListRequirements(filters RequirementFilters) ([]models.Requirement, error) {
	// Build filter map
	filterMap := make(map[string]interface{})
	
	if filters.UserStoryID != nil {
		filterMap["user_story_id"] = *filters.UserStoryID
	}
	if filters.AcceptanceCriteriaID != nil {
		filterMap["acceptance_criteria_id"] = *filters.AcceptanceCriteriaID
	}
	if filters.CreatorID != nil {
		filterMap["creator_id"] = *filters.CreatorID
	}
	if filters.AssigneeID != nil {
		filterMap["assignee_id"] = *filters.AssigneeID
	}
	if filters.Status != nil {
		filterMap["status"] = *filters.Status
	}
	if filters.Priority != nil {
		filterMap["priority"] = *filters.Priority
	}
	if filters.TypeID != nil {
		filterMap["type_id"] = *filters.TypeID
	}

	// Set default ordering
	orderBy := "created_at DESC"
	if filters.OrderBy != "" {
		orderBy = filters.OrderBy
	}

	// Set default limit
	limit := 50
	if filters.Limit > 0 {
		limit = filters.Limit
	}

	requirements, err := s.requirementRepo.List(filterMap, orderBy, limit, filters.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list requirements: %w", err)
	}

	return requirements, nil
}

// GetRequirementWithRelationships retrieves a requirement with its relationships
func (s *requirementService) GetRequirementWithRelationships(id uuid.UUID) (*models.Requirement, error) {
	requirement, err := s.requirementRepo.GetWithRelationships(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRequirementNotFound
		}
		return nil, fmt.Errorf("failed to get requirement with relationships: %w", err)
	}
	return requirement, nil
}

// GetRequirementsByUserStory retrieves requirements by user story ID
func (s *requirementService) GetRequirementsByUserStory(userStoryID uuid.UUID) ([]models.Requirement, error) {
	// Validate user story exists
	if exists, err := s.userStoryRepo.Exists(userStoryID); err != nil {
		return nil, fmt.Errorf("failed to check user story existence: %w", err)
	} else if !exists {
		return nil, ErrUserStoryNotFound
	}

	requirements, err := s.requirementRepo.GetByUserStory(userStoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get requirements by user story: %w", err)
	}

	return requirements, nil
}

// ChangeRequirementStatus changes the status of a requirement
func (s *requirementService) ChangeRequirementStatus(id uuid.UUID, newStatus models.RequirementStatus) (*models.Requirement, error) {
	requirement, err := s.requirementRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRequirementNotFound
		}
		return nil, fmt.Errorf("failed to get requirement: %w", err)
	}

	if !requirement.IsValidStatus(newStatus) {
		return nil, ErrInvalidRequirementStatus
	}

	if !requirement.CanTransitionTo(newStatus) {
		return nil, ErrInvalidStatusTransition
	}

	requirement.Status = newStatus
	if err := s.requirementRepo.Update(requirement); err != nil {
		return nil, fmt.Errorf("failed to update requirement status: %w", err)
	}

	return requirement, nil
}

// AssignRequirement assigns a requirement to a user
func (s *requirementService) AssignRequirement(id uuid.UUID, assigneeID uuid.UUID) (*models.Requirement, error) {
	// Validate assignee exists
	if exists, err := s.userRepo.Exists(assigneeID); err != nil {
		return nil, fmt.Errorf("failed to check assignee existence: %w", err)
	} else if !exists {
		return nil, ErrUserNotFound
	}

	requirement, err := s.requirementRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRequirementNotFound
		}
		return nil, fmt.Errorf("failed to get requirement: %w", err)
	}

	requirement.AssigneeID = assigneeID
	if err := s.requirementRepo.Update(requirement); err != nil {
		return nil, fmt.Errorf("failed to assign requirement: %w", err)
	}

	return requirement, nil
}

// CreateRelationship creates a relationship between two requirements
func (s *requirementService) CreateRelationship(req CreateRelationshipRequest) (*models.RequirementRelationship, error) {
	// Validate that source and target are different
	if req.SourceRequirementID == req.TargetRequirementID {
		return nil, ErrCircularRelationship
	}

	// Validate source requirement exists
	if exists, err := s.requirementRepo.Exists(req.SourceRequirementID); err != nil {
		return nil, fmt.Errorf("failed to check source requirement existence: %w", err)
	} else if !exists {
		return nil, ErrRequirementNotFound
	}

	// Validate target requirement exists
	if exists, err := s.requirementRepo.Exists(req.TargetRequirementID); err != nil {
		return nil, fmt.Errorf("failed to check target requirement existence: %w", err)
	} else if !exists {
		return nil, ErrRequirementNotFound
	}

	// Validate relationship type exists
	if exists, err := s.relationshipTypeRepo.Exists(req.RelationshipTypeID); err != nil {
		return nil, fmt.Errorf("failed to check relationship type existence: %w", err)
	} else if !exists {
		return nil, ErrRelationshipTypeNotFound
	}

	// Validate creator exists
	if exists, err := s.userRepo.Exists(req.CreatedBy); err != nil {
		return nil, fmt.Errorf("failed to check creator existence: %w", err)
	} else if !exists {
		return nil, ErrUserNotFound
	}

	// Check if relationship already exists
	exists, err := s.requirementRelationshipRepo.ExistsRelationship(
		req.SourceRequirementID, req.TargetRequirementID, req.RelationshipTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing relationship: %w", err)
	}
	if exists {
		return nil, ErrDuplicateRelationship
	}

	relationship := &models.RequirementRelationship{
		ID:                  uuid.New(),
		SourceRequirementID: req.SourceRequirementID,
		TargetRequirementID: req.TargetRequirementID,
		RelationshipTypeID:  req.RelationshipTypeID,
		CreatedBy:           req.CreatedBy,
	}

	if err := s.requirementRelationshipRepo.Create(relationship); err != nil {
		return nil, fmt.Errorf("failed to create relationship: %w", err)
	}

	return relationship, nil
}

// DeleteRelationship deletes a requirement relationship
func (s *requirementService) DeleteRelationship(id uuid.UUID) error {
	// Check if relationship exists
	_, err := s.requirementRelationshipRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrRequirementNotFound
		}
		return fmt.Errorf("failed to get relationship: %w", err)
	}

	if err := s.requirementRelationshipRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete relationship: %w", err)
	}

	return nil
}

// GetRelationshipsByRequirement retrieves all relationships for a requirement
func (s *requirementService) GetRelationshipsByRequirement(requirementID uuid.UUID) ([]models.RequirementRelationship, error) {
	// Validate requirement exists
	if exists, err := s.requirementRepo.Exists(requirementID); err != nil {
		return nil, fmt.Errorf("failed to check requirement existence: %w", err)
	} else if !exists {
		return nil, ErrRequirementNotFound
	}

	relationships, err := s.requirementRelationshipRepo.GetByRequirement(requirementID)
	if err != nil {
		return nil, fmt.Errorf("failed to get relationships: %w", err)
	}

	return relationships, nil
}

// SearchRequirements performs full-text search on requirements
func (s *requirementService) SearchRequirements(searchText string) ([]models.Requirement, error) {
	if searchText == "" {
		return []models.Requirement{}, nil
	}

	requirements, err := s.requirementRepo.SearchByText(searchText)
	if err != nil {
		return nil, fmt.Errorf("failed to search requirements: %w", err)
	}

	return requirements, nil
}