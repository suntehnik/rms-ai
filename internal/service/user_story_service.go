package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

var (
	ErrUserStoryNotFound        = errors.New("user story not found")
	ErrUserStoryHasRequirements = errors.New("user story has associated requirements and cannot be deleted")
	ErrInvalidUserStoryStatus   = errors.New("invalid user story status")
	ErrInvalidUserStoryTemplate = errors.New("user story description must follow template: 'As [role], I want [function], so that [goal]'")
)

// UserStoryService defines the interface for user story business logic
type UserStoryService interface {
	CreateUserStory(req CreateUserStoryRequest) (*models.UserStory, error)
	GetUserStoryByID(id uuid.UUID) (*models.UserStory, error)
	GetUserStoryByReferenceID(referenceID string) (*models.UserStory, error)
	UpdateUserStory(id uuid.UUID, req UpdateUserStoryRequest) (*models.UserStory, error)
	DeleteUserStory(id uuid.UUID, force bool) error
	ListUserStories(filters UserStoryFilters) ([]models.UserStory, error)
	GetUserStoryWithAcceptanceCriteria(id uuid.UUID) (*models.UserStory, error)
	GetUserStoryWithRequirements(id uuid.UUID) (*models.UserStory, error)
	GetUserStoriesByEpic(epicID uuid.UUID) ([]models.UserStory, error)
	ChangeUserStoryStatus(id uuid.UUID, newStatus models.UserStoryStatus) (*models.UserStory, error)
	AssignUserStory(id uuid.UUID, assigneeID uuid.UUID) (*models.UserStory, error)
}

// CreateUserStoryRequest represents the request to create a user story
type CreateUserStoryRequest struct {
	EpicID      uuid.UUID       `json:"epic_id,omitempty"`
	CreatorID   uuid.UUID       `json:"creator_id" binding:"required"`
	AssigneeID  *uuid.UUID      `json:"assignee_id,omitempty"`
	Priority    models.Priority `json:"priority" binding:"required,min=1,max=4"`
	Title       string          `json:"title" binding:"required,max=500"`
	Description *string         `json:"description,omitempty"`
}

// UpdateUserStoryRequest represents the request to update a user story
type UpdateUserStoryRequest struct {
	AssigneeID  *uuid.UUID              `json:"assignee_id,omitempty"`
	Priority    *models.Priority        `json:"priority,omitempty"`
	Status      *models.UserStoryStatus `json:"status,omitempty"`
	Title       *string                 `json:"title,omitempty"`
	Description *string                 `json:"description,omitempty"`
}

// UserStoryFilters represents filters for listing user stories
type UserStoryFilters struct {
	EpicID     *uuid.UUID              `json:"epic_id,omitempty"`
	CreatorID  *uuid.UUID              `json:"creator_id,omitempty"`
	AssigneeID *uuid.UUID              `json:"assignee_id,omitempty"`
	Status     *models.UserStoryStatus `json:"status,omitempty"`
	Priority   *models.Priority        `json:"priority,omitempty"`
	OrderBy    string                  `json:"order_by,omitempty"`
	Limit      int                     `json:"limit,omitempty"`
	Offset     int                     `json:"offset,omitempty"`
}

// userStoryService implements UserStoryService interface
type userStoryService struct {
	userStoryRepo repository.UserStoryRepository
	epicRepo      repository.EpicRepository
	userRepo      repository.UserRepository
}

// NewUserStoryService creates a new user story service instance
func NewUserStoryService(
	userStoryRepo repository.UserStoryRepository,
	epicRepo repository.EpicRepository,
	userRepo repository.UserRepository,
) UserStoryService {
	return &userStoryService{
		userStoryRepo: userStoryRepo,
		epicRepo:      epicRepo,
		userRepo:      userRepo,
	}
}

// validateUserStoryTemplate validates if the description follows the user story template
func (s *userStoryService) validateUserStoryTemplate(description *string) error {
	if description == nil || *description == "" {
		return ErrInvalidUserStoryTemplate
	}

	desc := strings.ToLower(strings.TrimSpace(*description))

	// Check for required components of user story template
	hasAs := strings.Contains(desc, "as ")
	hasIWant := strings.Contains(desc, "i want")
	hasSoThat := strings.Contains(desc, "so that")

	if !hasAs || !hasIWant || !hasSoThat {
		return ErrInvalidUserStoryTemplate
	}

	return nil
}

// CreateUserStory creates a new user story
func (s *userStoryService) CreateUserStory(req CreateUserStoryRequest) (*models.UserStory, error) {
	// Validate required fields
	if req.EpicID == uuid.Nil {
		return nil, fmt.Errorf("epic_id is required")
	}

	// Validate priority
	if req.Priority < models.PriorityCritical || req.Priority > models.PriorityLow {
		return nil, ErrInvalidPriority
	}

	// Validate epic exists
	if exists, err := s.epicRepo.Exists(req.EpicID); err != nil {
		return nil, fmt.Errorf("failed to check epic existence: %w", err)
	} else if !exists {
		return nil, ErrEpicNotFound
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

	// Validate user story template format if description is provided
	if err := s.validateUserStoryTemplate(req.Description); err != nil {
		return nil, err
	}

	userStory := &models.UserStory{
		ID:          uuid.New(),
		EpicID:      req.EpicID,
		CreatorID:   req.CreatorID,
		AssigneeID:  assigneeID,
		Priority:    req.Priority,
		Status:      models.UserStoryStatusBacklog, // Default status
		Title:       req.Title,
		Description: req.Description,
	}

	if err := s.userStoryRepo.Create(userStory); err != nil {
		return nil, fmt.Errorf("failed to create user story: %w", err)
	}

	return userStory, nil
}

// GetUserStoryByID retrieves a user story by its ID
func (s *userStoryService) GetUserStoryByID(id uuid.UUID) (*models.UserStory, error) {
	userStory, err := s.userStoryRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserStoryNotFound
		}
		return nil, fmt.Errorf("failed to get user story: %w", err)
	}
	return userStory, nil
}

// GetUserStoryByReferenceID retrieves a user story by its reference ID
func (s *userStoryService) GetUserStoryByReferenceID(referenceID string) (*models.UserStory, error) {
	userStory, err := s.userStoryRepo.GetByReferenceID(referenceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserStoryNotFound
		}
		return nil, fmt.Errorf("failed to get user story: %w", err)
	}
	return userStory, nil
}

// UpdateUserStory updates an existing user story
func (s *userStoryService) UpdateUserStory(id uuid.UUID, req UpdateUserStoryRequest) (*models.UserStory, error) {
	userStory, err := s.userStoryRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserStoryNotFound
		}
		return nil, fmt.Errorf("failed to get user story: %w", err)
	}

	// Update fields if provided
	if req.AssigneeID != nil {
		// Validate assignee exists
		if exists, err := s.userRepo.Exists(*req.AssigneeID); err != nil {
			return nil, fmt.Errorf("failed to check assignee existence: %w", err)
		} else if !exists {
			return nil, ErrUserNotFound
		}
		userStory.AssigneeID = *req.AssigneeID
	}

	if req.Priority != nil {
		if *req.Priority < models.PriorityCritical || *req.Priority > models.PriorityLow {
			return nil, ErrInvalidPriority
		}
		userStory.Priority = *req.Priority
	}

	if req.Status != nil {
		if !userStory.IsValidStatus(*req.Status) {
			return nil, ErrInvalidUserStoryStatus
		}
		if !userStory.CanTransitionTo(*req.Status) {
			return nil, ErrInvalidStatusTransition
		}
		userStory.Status = *req.Status
	}

	if req.Title != nil {
		userStory.Title = *req.Title
	}

	if req.Description != nil {
		// Validate user story template format
		if err := s.validateUserStoryTemplate(req.Description); err != nil {
			return nil, err
		}
		userStory.Description = req.Description
	}

	if err := s.userStoryRepo.Update(userStory); err != nil {
		return nil, fmt.Errorf("failed to update user story: %w", err)
	}

	return userStory, nil
}

// DeleteUserStory deletes a user story with dependency validation
func (s *userStoryService) DeleteUserStory(id uuid.UUID, force bool) error {
	// Check if user story exists
	_, err := s.userStoryRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserStoryNotFound
		}
		return fmt.Errorf("failed to get user story: %w", err)
	}

	// Check for requirements unless force delete
	if !force {
		hasRequirements, err := s.userStoryRepo.HasRequirements(id)
		if err != nil {
			return fmt.Errorf("failed to check requirements: %w", err)
		}
		if hasRequirements {
			return ErrUserStoryHasRequirements
		}
	}

	// Delete the user story (cascade will handle acceptance criteria and requirements if force=true)
	if err := s.userStoryRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete user story: %w", err)
	}

	return nil
}

// ListUserStories retrieves user stories with optional filtering
func (s *userStoryService) ListUserStories(filters UserStoryFilters) ([]models.UserStory, error) {
	// Build filter map
	filterMap := make(map[string]interface{})

	if filters.EpicID != nil {
		filterMap["epic_id"] = *filters.EpicID
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

	userStories, err := s.userStoryRepo.List(filterMap, orderBy, limit, filters.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user stories: %w", err)
	}

	return userStories, nil
}

// GetUserStoryWithAcceptanceCriteria retrieves a user story with its acceptance criteria
func (s *userStoryService) GetUserStoryWithAcceptanceCriteria(id uuid.UUID) (*models.UserStory, error) {
	userStory, err := s.userStoryRepo.GetWithAcceptanceCriteria(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserStoryNotFound
		}
		return nil, fmt.Errorf("failed to get user story with acceptance criteria: %w", err)
	}
	return userStory, nil
}

// GetUserStoryWithRequirements retrieves a user story with its requirements
func (s *userStoryService) GetUserStoryWithRequirements(id uuid.UUID) (*models.UserStory, error) {
	userStory, err := s.userStoryRepo.GetWithRequirements(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserStoryNotFound
		}
		return nil, fmt.Errorf("failed to get user story with requirements: %w", err)
	}
	return userStory, nil
}

// GetUserStoriesByEpic retrieves user stories by epic ID
func (s *userStoryService) GetUserStoriesByEpic(epicID uuid.UUID) ([]models.UserStory, error) {
	// Validate epic exists
	if exists, err := s.epicRepo.Exists(epicID); err != nil {
		return nil, fmt.Errorf("failed to check epic existence: %w", err)
	} else if !exists {
		return nil, ErrEpicNotFound
	}

	userStories, err := s.userStoryRepo.GetByEpic(epicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stories by epic: %w", err)
	}

	return userStories, nil
}

// ChangeUserStoryStatus changes the status of a user story
func (s *userStoryService) ChangeUserStoryStatus(id uuid.UUID, newStatus models.UserStoryStatus) (*models.UserStory, error) {
	userStory, err := s.userStoryRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserStoryNotFound
		}
		return nil, fmt.Errorf("failed to get user story: %w", err)
	}

	if !userStory.IsValidStatus(newStatus) {
		return nil, ErrInvalidUserStoryStatus
	}

	if !userStory.CanTransitionTo(newStatus) {
		return nil, ErrInvalidStatusTransition
	}

	userStory.Status = newStatus
	if err := s.userStoryRepo.Update(userStory); err != nil {
		return nil, fmt.Errorf("failed to update user story status: %w", err)
	}

	return userStory, nil
}

// AssignUserStory assigns a user story to a user
func (s *userStoryService) AssignUserStory(id uuid.UUID, assigneeID uuid.UUID) (*models.UserStory, error) {
	// Validate assignee exists
	if exists, err := s.userRepo.Exists(assigneeID); err != nil {
		return nil, fmt.Errorf("failed to check assignee existence: %w", err)
	} else if !exists {
		return nil, ErrUserNotFound
	}

	userStory, err := s.userStoryRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserStoryNotFound
		}
		return nil, fmt.Errorf("failed to get user story: %w", err)
	}

	userStory.AssigneeID = assigneeID
	if err := s.userStoryRepo.Update(userStory); err != nil {
		return nil, fmt.Errorf("failed to assign user story: %w", err)
	}

	return userStory, nil
}
