package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/validation"
)

var (
	ErrEpicNotFound       = errors.New("epic not found")
	ErrEpicHasUserStories = errors.New("epic has associated user stories and cannot be deleted")
	ErrInvalidEpicStatus  = errors.New("invalid epic status")
	ErrInvalidPriority    = errors.New("invalid priority")
	ErrUserNotFound       = errors.New("user not found")
)

// EpicService defines the interface for epic business logic
type EpicService interface {
	CreateEpic(req CreateEpicRequest) (*models.Epic, error)
	GetEpicByID(id uuid.UUID) (*models.Epic, error)
	GetEpicByReferenceID(referenceID string) (*models.Epic, error)
	UpdateEpic(id uuid.UUID, req UpdateEpicRequest) (*models.Epic, error)
	DeleteEpic(id uuid.UUID, force bool) error
	ListEpics(filters EpicFilters) ([]models.Epic, int64, error)
	GetEpicWithUserStories(id uuid.UUID) (*models.Epic, error)
	GetEpicWithCompleteHierarchy(id uuid.UUID) (*models.Epic, error)
	ChangeEpicStatus(id uuid.UUID, newStatus models.EpicStatus) (*models.Epic, error)
	AssignEpic(id uuid.UUID, assigneeID *uuid.UUID) (*models.Epic, error)
}

// CreateEpicRequest represents the request to create an epic
// @Description Request payload for creating a new epic
type CreateEpicRequest struct {
	// CreatorID is the UUID of the user creating the epic
	// @Description UUID of the user who is creating this epic (required)
	// @Example "123e4567-e89b-12d3-a456-426614174001"
	CreatorID uuid.UUID `json:"creator_id"`

	// AssigneeID is the UUID of the user to assign the epic to (optional, defaults to creator)
	// @Description UUID of the user to assign this epic to (optional, defaults to creator if not provided)
	// @Example "123e4567-e89b-12d3-a456-426614174002"
	AssigneeID *uuid.UUID `json:"assignee_id,omitempty"`

	// Priority is the importance level of the epic
	// @Description Priority level of the epic (1=Critical, 2=High, 3=Medium, 4=Low)
	// @Minimum 1
	// @Maximum 4
	// @Example 1
	Priority models.Priority `json:"priority" binding:"required,min=1,max=4"`

	// Title is the name/summary of the epic
	// @Description Title or name of the epic (required, max 500 characters)
	// @MaxLength 500
	// @Example "User Authentication System"
	Title string `json:"title" binding:"required,max=500"`

	// Description provides detailed information about the epic
	// @Description Detailed description of the epic's purpose and scope (optional, max 50000 characters)
	// @MaxLength 50000
	// @Example "Implement a comprehensive user authentication and authorization system with JWT tokens, role-based access control, and secure password management."
	Description *string `json:"description,omitempty"`
}

// UpdateEpicRequest represents the request to update an epic
// @Description Request payload for updating an existing epic (all fields are optional)
type UpdateEpicRequest struct {
	// AssigneeID is the UUID of the user to assign the epic to
	// @Description UUID of the user to assign this epic to (optional)
	// @Example "123e4567-e89b-12d3-a456-426614174002"
	AssigneeID *uuid.UUID `json:"assignee_id,omitempty"`

	// Priority is the importance level of the epic
	// @Description Priority level of the epic (1=Critical, 2=High, 3=Medium, 4=Low) (optional)
	// @Minimum 1
	// @Maximum 4
	// @Example 2
	Priority *models.Priority `json:"priority,omitempty"`

	// Status is the workflow state of the epic
	// @Description Current status of the epic in the workflow (optional)
	// @Enum Backlog,Draft,In Progress,Done,Cancelled
	// @Example "In Progress"
	Status *models.EpicStatus `json:"status,omitempty"`

	// Title is the name/summary of the epic
	// @Description Title or name of the epic (optional, max 500 characters)
	// @MaxLength 500
	// @Example "Enhanced User Authentication System"
	Title *string `json:"title,omitempty"`

	// Description provides detailed information about the epic
	// @Description Detailed description of the epic's purpose and scope (optional, max 50000 characters)
	// @MaxLength 50000
	// @Example "Enhanced implementation with multi-factor authentication and advanced security features."
	Description *string `json:"description,omitempty"`
}

// EpicFilters represents filters for listing epics
// @Description Filters and pagination options for listing epics
type EpicFilters struct {
	// CreatorID filters epics by creator
	// @Description Filter epics by creator UUID (optional)
	// @Example "123e4567-e89b-12d3-a456-426614174001"
	CreatorID *uuid.UUID `json:"creator_id,omitempty"`

	// AssigneeID filters epics by assignee
	// @Description Filter epics by assignee UUID (optional)
	// @Example "123e4567-e89b-12d3-a456-426614174002"
	AssigneeID *uuid.UUID `json:"assignee_id,omitempty"`

	// Status filters epics by status
	// @Description Filter epics by status (optional)
	// @Enum Backlog,Draft,In Progress,Done,Cancelled
	// @Example "Backlog"
	Status *models.EpicStatus `json:"status,omitempty"`

	// Priority filters epics by priority level
	// @Description Filter epics by priority level (optional)
	// @Minimum 1
	// @Maximum 4
	// @Example 1
	Priority *models.Priority `json:"priority,omitempty"`

	// Include specifies which related entities to include
	// @Description Comma-separated list of related entities to include (optional)
	// @Example "creator,assignee,user_stories,comments"
	Include []string `json:"include,omitempty"`

	// OrderBy specifies the field and direction for sorting
	// @Description Order results by field and direction (optional, default: "created_at DESC")
	// @Example "created_at DESC"
	OrderBy string `json:"order_by,omitempty"`

	// Limit specifies the maximum number of results
	// @Description Maximum number of results to return (optional, default: 50, max: 100)
	// @Minimum 1
	// @Maximum 100
	// @Example 20
	Limit int `json:"limit,omitempty"`

	// Offset specifies the number of results to skip
	// @Description Number of results to skip for pagination (optional, default: 0)
	// @Minimum 0
	// @Example 0
	Offset int `json:"offset,omitempty"`
}

// ChangeEpicStatusRequest represents the request to change an epic's status
// @Description Request payload for changing an epic's status
type ChangeEpicStatusRequest struct {
	// Status is the new workflow state for the epic
	// @Description New status for the epic
	// @Enum Backlog,Draft,In Progress,Done,Cancelled
	// @Example "In Progress"
	Status models.EpicStatus `json:"status" binding:"required"`
}

// AssignEpicRequest represents the request to assign an epic to a user
// @Description Request payload for assigning an epic to a user
type AssignEpicRequest struct {
	// AssigneeID is the UUID of the user to assign the epic to (nullable for unassignment)
	// @Description UUID of the user to assign this epic to (null to unassign)
	// @Example "123e4567-e89b-12d3-a456-426614174002"
	AssigneeID *uuid.UUID `json:"assignee_id"`
}

// epicService implements EpicService interface
type epicService struct {
	epicRepo        repository.EpicRepository
	userRepo        repository.UserRepository
	statusValidator validation.StatusValidator
}

// NewEpicService creates a new epic service instance
func NewEpicService(epicRepo repository.EpicRepository, userRepo repository.UserRepository) EpicService {
	return &epicService{
		epicRepo:        epicRepo,
		userRepo:        userRepo,
		statusValidator: validation.NewStatusValidator(),
	}
}

// CreateEpic creates a new epic
func (s *epicService) CreateEpic(req CreateEpicRequest) (*models.Epic, error) {
	// Validate priority first
	if req.Priority < models.PriorityCritical || req.Priority > models.PriorityLow {
		return nil, ErrInvalidPriority
	}

	// Validate creator exists (creator ID is set from authenticated context)
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
		ID:          uuid.New(),
		CreatorID:   req.CreatorID,
		AssigneeID:  assigneeID,
		Priority:    req.Priority,
		Status:      models.EpicStatusBacklog, // Default status
		Title:       req.Title,
		Description: req.Description,
	}

	if err := s.epicRepo.Create(epic); err != nil {
		return nil, fmt.Errorf("failed to create epic: %w", err)
	}

	return epic, nil
}

// GetEpicByID retrieves an epic by its ID with creator and assignee preloaded
func (s *epicService) GetEpicByID(id uuid.UUID) (*models.Epic, error) {
	epic, err := s.epicRepo.GetByIDWithUsers(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEpicNotFound
		}
		return nil, fmt.Errorf("failed to get epic: %w", err)
	}
	return epic, nil
}

// GetEpicByReferenceID retrieves an epic by its reference ID with creator and assignee preloaded
func (s *epicService) GetEpicByReferenceID(referenceID string) (*models.Epic, error) {
	epic, err := s.epicRepo.GetByReferenceIDWithUsers(referenceID)
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
		// Validate status using centralized validator
		if err := s.statusValidator.ValidateEpicStatus(string(*req.Status)); err != nil {
			return nil, err
		}

		// Check status transition rules
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
func (s *epicService) ListEpics(filters EpicFilters) ([]models.Epic, int64, error) {
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

	// Get total count with filters
	totalCount, err := s.epicRepo.Count(filterMap)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count epics: %w", err)
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

	// Always include default preloads (Creator and Assignee)
	defaultIncludes := []string{"Creator", "Assignee"}

	// Merge with any additional includes specified in filters
	includes := make(map[string]bool)
	for _, include := range defaultIncludes {
		includes[include] = true
	}
	for _, include := range filters.Include {
		includes[include] = true
	}

	// Convert back to slice
	finalIncludes := make([]string, 0, len(includes))
	for include := range includes {
		finalIncludes = append(finalIncludes, include)
	}

	// Always use the method with includes since we have default preloads
	epics, err := s.epicRepo.ListWithIncludes(filterMap, finalIncludes, orderBy, limit, filters.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list epics: %w", err)
	}

	return epics, totalCount, nil
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

// GetEpicWithCompleteHierarchy retrieves an epic with complete hierarchy
// This includes: Epic → UserStories → [Requirements, AcceptanceCriteria]
// Requirements and AcceptanceCriteria are loaded at the same level under each UserStory
func (s *epicService) GetEpicWithCompleteHierarchy(id uuid.UUID) (*models.Epic, error) {
	epic, err := s.epicRepo.GetCompleteHierarchy(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEpicNotFound
		}
		return nil, fmt.Errorf("failed to get epic hierarchy: %w", err)
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

	// Validate status using centralized validator
	if err := s.statusValidator.ValidateEpicStatus(string(newStatus)); err != nil {
		return nil, err
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

// AssignEpic assigns an epic to a user or unassigns it
func (s *epicService) AssignEpic(id uuid.UUID, assigneeID *uuid.UUID) (*models.Epic, error) {
	// Validate assignee exists if provided
	if assigneeID != nil {
		if exists, err := s.userRepo.Exists(*assigneeID); err != nil {
			return nil, fmt.Errorf("failed to check assignee existence: %w", err)
		} else if !exists {
			return nil, ErrUserNotFound
		}
	}

	epic, err := s.epicRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEpicNotFound
		}
		return nil, fmt.Errorf("failed to get epic: %w", err)
	}

	// Assign or unassign based on whether assigneeID is provided
	if assigneeID != nil {
		epic.AssigneeID = *assigneeID
	} else {
		// Unassign by setting to creator (default behavior)
		epic.AssigneeID = epic.CreatorID
	}

	if err := s.epicRepo.Update(epic); err != nil {
		return nil, fmt.Errorf("failed to assign epic: %w", err)
	}

	return epic, nil
}
