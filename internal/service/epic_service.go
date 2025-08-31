package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

var (
	ErrEpicNotFound           = errors.New("epic not found")
	ErrEpicHasUserStories     = errors.New("epic has associated user stories and cannot be deleted")
	ErrInvalidEpicStatus      = errors.New("invalid epic status")
	ErrInvalidPriority        = errors.New("invalid priority")
	ErrUserNotFound           = errors.New("user not found")
)

// EpicService defines the interface for epic business logic
type EpicService interface {
	CreateEpic(req CreateEpicRequest) (*models.Epic, error)
	GetEpicByID(id uuid.UUID) (*models.Epic, error)
	GetEpicByReferenceID(referenceID string) (*models.Epic, error)
	UpdateEpic(id uuid.UUID, req UpdateEpicRequest) (*models.Epic, error)
	DeleteEpic(id uuid.UUID, force bool) error
	ListEpics(filters EpicFilters) ([]models.Epic, error)
	GetEpicWithUserStories(id uuid.UUID) (*models.Epic, error)
	ChangeEpicStatus(id uuid.UUID, newStatus models.EpicStatus) (*models.Epic, error)
	AssignEpic(id uuid.UUID, assigneeID uuid.UUID) (*models.Epic, error)
}

// CreateEpicRequest represents the request to create an epic
type CreateEpicRequest struct {
	CreatorID   uuid.UUID       `json:"creator_id" binding:"required"`
	AssigneeID  *uuid.UUID      `json:"assignee_id,omitempty"`
	Priority    models.Priority `json:"priority" binding:"required,min=1,max=4"`
	Title       string          `json:"title" binding:"required,max=500"`
	Description *string         `json:"description,omitempty"`
}

// UpdateEpicRequest represents the request to update an epic
type UpdateEpicRequest struct {
	AssigneeID  *uuid.UUID      `json:"assignee_id,omitempty"`
	Priority    *models.Priority `json:"priority,omitempty"`
	Status      *models.EpicStatus `json:"status,omitempty"`
	Title       *string         `json:"title,omitempty"`
	Description *string         `json:"description,omitempty"`
}

// EpicFilters represents filters for listing epics
type EpicFilters struct {
	CreatorID  *uuid.UUID         `json:"creator_id,omitempty"`
	AssigneeID *uuid.UUID         `json:"assignee_id,omitempty"`
	Status     *models.EpicStatus `json:"status,omitempty"`
	Priority   *models.Priority   `json:"priority,omitempty"`
	OrderBy    string             `json:"order_by,omitempty"`
	Limit      int                `json:"limit,omitempty"`
	Offset     int                `json:"offset,omitempty"`
}

// epicService implements EpicService interface
type epicService struct {
	epicRepo repository.EpicRepository
	userRepo repository.UserRepository
}

// NewEpicService creates a new epic service instance
func NewEpicService(epicRepo repository.EpicRepository, userRepo repository.UserRepository) EpicService {
	return &epicService{
		epicRepo: epicRepo,
		userRepo: userRepo,
	}
}

// CreateEpic creates a new epic
func (s *epicService) CreateEpic(req CreateEpicRequest) (*models.Epic, error) {
	// Validate priority first
	if req.Priority < models.PriorityCritical || req.Priority > models.PriorityLow {
		return nil, ErrInvalidPriority
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

	epic := &models.Epic{
		ID:           uuid.New(),
		CreatorID:    req.CreatorID,
		AssigneeID:   assigneeID,
		Priority:     req.Priority,
		Status:       models.EpicStatusBacklog, // Default status
		Title:        req.Title,
		Description:  req.Description,
	}

	if err := s.epicRepo.Create(epic); err != nil {
		return nil, fmt.Errorf("failed to create epic: %w", err)
	}

	return epic, nil
}

// GetEpicByID retrieves an epic by its ID
func (s *epicService) GetEpicByID(id uuid.UUID) (*models.Epic, error) {
	epic, err := s.epicRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEpicNotFound
		}
		return nil, fmt.Errorf("failed to get epic: %w", err)
	}
	return epic, nil
}

// GetEpicByReferenceID retrieves an epic by its reference ID
func (s *epicService) GetEpicByReferenceID(referenceID string) (*models.Epic, error) {
	epic, err := s.epicRepo.GetByReferenceID(referenceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEpicNotFound
		}
		return nil, fmt.Errorf("failed to get epic: %w", err)
	}
	return epic, nil
}

// UpdateEpic updates an existing epic
func (s *epicService) UpdateEpic(id uuid.UUID, req UpdateEpicRequest) (*models.Epic, error) {
	epic, err := s.epicRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEpicNotFound
		}
		return nil, fmt.Errorf("failed to get epic: %w", err)
	}

	// Update fields if provided
	if req.AssigneeID != nil {
		// Validate assignee exists
		if exists, err := s.userRepo.Exists(*req.AssigneeID); err != nil {
			return nil, fmt.Errorf("failed to check assignee existence: %w", err)
		} else if !exists {
			return nil, ErrUserNotFound
		}
		epic.AssigneeID = *req.AssigneeID
	}

	if req.Priority != nil {
		if *req.Priority < models.PriorityCritical || *req.Priority > models.PriorityLow {
			return nil, ErrInvalidPriority
		}
		epic.Priority = *req.Priority
	}

	if req.Status != nil {
		if !epic.IsValidStatus(*req.Status) {
			return nil, ErrInvalidEpicStatus
		}
		if !epic.CanTransitionTo(*req.Status) {
			return nil, ErrInvalidStatusTransition
		}
		epic.Status = *req.Status
	}

	if req.Title != nil {
		epic.Title = *req.Title
	}

	if req.Description != nil {
		epic.Description = req.Description
	}

	if err := s.epicRepo.Update(epic); err != nil {
		return nil, fmt.Errorf("failed to update epic: %w", err)
	}

	return epic, nil
}

// DeleteEpic deletes an epic with dependency validation
func (s *epicService) DeleteEpic(id uuid.UUID, force bool) error {
	// Check if epic exists
	_, err := s.epicRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrEpicNotFound
		}
		return fmt.Errorf("failed to get epic: %w", err)
	}

	// Check for user stories unless force delete
	if !force {
		hasUserStories, err := s.epicRepo.HasUserStories(id)
		if err != nil {
			return fmt.Errorf("failed to check user stories: %w", err)
		}
		if hasUserStories {
			return ErrEpicHasUserStories
		}
	}

	// Delete the epic (cascade will handle user stories if force=true)
	if err := s.epicRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete epic: %w", err)
	}

	return nil
}

// ListEpics retrieves epics with optional filtering
func (s *epicService) ListEpics(filters EpicFilters) ([]models.Epic, error) {
	// Build filter map
	filterMap := make(map[string]interface{})
	
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

	epics, err := s.epicRepo.List(filterMap, orderBy, limit, filters.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list epics: %w", err)
	}

	return epics, nil
}

// GetEpicWithUserStories retrieves an epic with its user stories
func (s *epicService) GetEpicWithUserStories(id uuid.UUID) (*models.Epic, error) {
	epic, err := s.epicRepo.GetWithUserStories(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEpicNotFound
		}
		return nil, fmt.Errorf("failed to get epic with user stories: %w", err)
	}
	return epic, nil
}

// ChangeEpicStatus changes the status of an epic
func (s *epicService) ChangeEpicStatus(id uuid.UUID, newStatus models.EpicStatus) (*models.Epic, error) {
	epic, err := s.epicRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEpicNotFound
		}
		return nil, fmt.Errorf("failed to get epic: %w", err)
	}

	if !epic.IsValidStatus(newStatus) {
		return nil, ErrInvalidEpicStatus
	}

	if !epic.CanTransitionTo(newStatus) {
		return nil, ErrInvalidStatusTransition
	}

	epic.Status = newStatus
	if err := s.epicRepo.Update(epic); err != nil {
		return nil, fmt.Errorf("failed to update epic status: %w", err)
	}

	return epic, nil
}

// AssignEpic assigns an epic to a user
func (s *epicService) AssignEpic(id uuid.UUID, assigneeID uuid.UUID) (*models.Epic, error) {
	// Validate assignee exists
	if exists, err := s.userRepo.Exists(assigneeID); err != nil {
		return nil, fmt.Errorf("failed to check assignee existence: %w", err)
	} else if !exists {
		return nil, ErrUserNotFound
	}

	epic, err := s.epicRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEpicNotFound
		}
		return nil, fmt.Errorf("failed to get epic: %w", err)
	}

	epic.AssigneeID = assigneeID
	if err := s.epicRepo.Update(epic); err != nil {
		return nil, fmt.Errorf("failed to assign epic: %w", err)
	}

	return epic, nil
}