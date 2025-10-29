package service

import (
	"errors"
	"fmt"
	"sort"
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
	ListUserStories(filters UserStoryFilters) ([]models.UserStory, int64, error)
	GetUserStoryWithAcceptanceCriteria(id uuid.UUID) (*models.UserStory, error)
	GetUserStoryWithRequirements(id uuid.UUID) (*models.UserStory, error)
	GetUserStoriesByEpic(epicID uuid.UUID) ([]models.UserStory, error)
	ChangeUserStoryStatus(id uuid.UUID, newStatus models.UserStoryStatus) (*models.UserStory, error)
	AssignUserStory(id uuid.UUID, assigneeID uuid.UUID) (*models.UserStory, error)
}

// CreateUserStoryRequest represents the request to create a user story
// @Description Request structure for creating a new user story
type CreateUserStoryRequest struct {
	// EpicID is the UUID of the epic this user story belongs to
	// @Description UUID of the epic that will contain this user story (required for direct creation)
	// @Example "123e4567-e89b-12d3-a456-426614174000"
	EpicID uuid.UUID `json:"epic_id,omitempty"`

	// CreatorID is the UUID of the user creating the user story
	// @Description UUID of the user who is creating this user story (set automatically from JWT token)
	// @Example "123e4567-e89b-12d3-a456-426614174001"
	CreatorID uuid.UUID `json:"creator_id,omitempty"`

	// AssigneeID is the UUID of the user assigned to the user story
	// @Description UUID of the user to assign this user story to (optional, defaults to creator)
	// @Example "123e4567-e89b-12d3-a456-426614174002"
	AssigneeID *uuid.UUID `json:"assignee_id,omitempty"`

	// Priority indicates the importance level of the user story
	// @Description Priority level of the user story (1=Critical, 2=High, 3=Medium, 4=Low)
	// @Minimum 1
	// @Maximum 4
	// @Example 2
	Priority models.Priority `json:"priority" binding:"required,min=1,max=4"`

	// Title is the name/summary of the user story
	// @Description Title or name of the user story (required, max 500 characters)
	// @MaxLength 500
	// @Example "User Login with Email and Password"
	Title string `json:"title" binding:"required,max=500"`

	// Description provides detailed information about the user story
	// @Description Detailed description following the template 'As [role], I want [function], so that [goal]' (optional, max 50000 characters)
	// @MaxLength 50000
	// @Example "As a registered user, I want to log in with my email and password, so that I can access my personalized dashboard and account features."
	Description *string `json:"description,omitempty"`
}

// UpdateUserStoryRequest represents the request to update a user story
// @Description Request structure for updating an existing user story (all fields are optional)
type UpdateUserStoryRequest struct {
	// AssigneeID is the UUID of the user to assign the user story to
	// @Description UUID of the user to assign this user story to (optional)
	// @Example "123e4567-e89b-12d3-a456-426614174002"
	AssigneeID *uuid.UUID `json:"assignee_id,omitempty"`

	// Priority indicates the importance level of the user story
	// @Description Priority level of the user story (1=Critical, 2=High, 3=Medium, 4=Low) (optional)
	// @Minimum 1
	// @Maximum 4
	// @Example 3
	Priority *models.Priority `json:"priority,omitempty"`

	// Status represents the current workflow state of the user story
	// @Description Current status of the user story in the workflow (optional)
	// @Enum Backlog,Draft,In Progress,Done,Cancelled
	// @Example "In Progress"
	Status *models.UserStoryStatus `json:"status,omitempty"`

	// Title is the name/summary of the user story
	// @Description Title or name of the user story (optional, max 500 characters)
	// @MaxLength 500
	// @Example "Enhanced User Login with Two-Factor Authentication"
	Title *string `json:"title,omitempty"`

	// Description provides detailed information about the user story
	// @Description Detailed description following the template 'As [role], I want [function], so that [goal]' (optional, max 50000 characters)
	// @MaxLength 50000
	// @Example "As a security-conscious user, I want to enable two-factor authentication on my account, so that I can protect my personal information from unauthorized access."
	Description *string `json:"description,omitempty"`
}

// UserStoryFilters represents filters for listing user stories
// @Description Filter and pagination options for listing user stories
type UserStoryFilters struct {
	// EpicID filters user stories by epic
	// @Description Filter user stories by epic UUID (optional)
	// @Example "123e4567-e89b-12d3-a456-426614174000"
	EpicID *uuid.UUID `json:"epic_id,omitempty"`

	// CreatorID filters user stories by creator
	// @Description Filter user stories by creator UUID (optional)
	// @Example "123e4567-e89b-12d3-a456-426614174001"
	CreatorID *uuid.UUID `json:"creator_id,omitempty"`

	// AssigneeID filters user stories by assignee
	// @Description Filter user stories by assignee UUID (optional)
	// @Example "123e4567-e89b-12d3-a456-426614174002"
	AssigneeID *uuid.UUID `json:"assignee_id,omitempty"`

	// Status filters user stories by status
	// @Description Filter user stories by status (optional)
	// @Enum Backlog,Draft,In Progress,Done,Cancelled
	// @Example "Backlog"
	Status *models.UserStoryStatus `json:"status,omitempty"`

	// Priority filters user stories by priority
	// @Description Filter user stories by priority level (optional)
	// @Minimum 1
	// @Maximum 4
	// @Example 2
	Priority *models.Priority `json:"priority,omitempty"`

	// Include specifies which related entities to include
	// @Description Comma-separated list of related entities to include (optional)
	// @Example "epic,creator,assignee,acceptance_criteria,requirements,comments"
	Include []string `json:"include,omitempty"`

	// OrderBy specifies the sort order
	// @Description Sort order for results (optional, default: "created_at DESC")
	// @Example "created_at DESC"
	// @Example "priority ASC"
	// @Example "title ASC"
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

// GetUserStoryByID retrieves a user story by its ID with creator, assignee, and epic populated
func (s *userStoryService) GetUserStoryByID(id uuid.UUID) (*models.UserStory, error) {
	userStory, err := s.userStoryRepo.GetByIDWithUsers(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserStoryNotFound
		}
		return nil, fmt.Errorf("failed to get user story: %w", err)
	}
	return userStory, nil
}

// GetUserStoryByReferenceID retrieves a user story by its reference ID with creator, assignee, and epic populated
func (s *userStoryService) GetUserStoryByReferenceID(referenceID string) (*models.UserStory, error) {
	userStory, err := s.userStoryRepo.GetByReferenceIDWithUsers(referenceID)
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
func (s *userStoryService) ListUserStories(filters UserStoryFilters) ([]models.UserStory, int64, error) {
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

	// Get total count with filters
	totalCount, err := s.userStoryRepo.Count(filterMap)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user stories: %w", err)
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

	// Always include default preloads (Creator, Assignee, and Epic)
	defaultIncludes := []string{"Creator", "Assignee", "Epic"}

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

	// Fix fluky unit tests
	sort.Strings(finalIncludes)

	// Always use the method with includes since we have default preloads
	userStories, err := s.userStoryRepo.ListWithIncludes(filterMap, finalIncludes, orderBy, limit, filters.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list user stories: %w", err)
	}

	return userStories, totalCount, nil
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
