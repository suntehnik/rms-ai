package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

var (
	ErrAcceptanceCriteriaNotFound          = errors.New("acceptance criteria not found")
	ErrAcceptanceCriteriaHasRequirements   = errors.New("acceptance criteria has associated requirements and cannot be deleted")
	ErrUserStoryMustHaveAcceptanceCriteria = errors.New("user story must have at least one acceptance criteria")
)

// AcceptanceCriteriaService defines the interface for acceptance criteria business logic
type AcceptanceCriteriaService interface {
	CreateAcceptanceCriteria(req CreateAcceptanceCriteriaRequest) (*models.AcceptanceCriteria, error)
	GetAcceptanceCriteriaByID(id uuid.UUID) (*models.AcceptanceCriteria, error)
	GetAcceptanceCriteriaByReferenceID(referenceID string) (*models.AcceptanceCriteria, error)
	UpdateAcceptanceCriteria(id uuid.UUID, req UpdateAcceptanceCriteriaRequest) (*models.AcceptanceCriteria, error)
	DeleteAcceptanceCriteria(id uuid.UUID, force bool) error
	ListAcceptanceCriteria(filters AcceptanceCriteriaFilters) ([]models.AcceptanceCriteria, int64, error)
	GetAcceptanceCriteriaByUserStory(userStoryID uuid.UUID, limit, offset int) ([]models.AcceptanceCriteria, int64, error)
	GetAcceptanceCriteriaByAuthor(authorID uuid.UUID, limit, offset int) ([]models.AcceptanceCriteria, int64, error)
	ValidateUserStoryHasAcceptanceCriteria(userStoryID uuid.UUID) error
}

// CreateAcceptanceCriteriaRequest represents the request to create acceptance criteria
type CreateAcceptanceCriteriaRequest struct {
	UserStoryID uuid.UUID `json:"user_story_id,omitempty"`
	AuthorID    uuid.UUID `json:"author_id"`
	Description string    `json:"description" binding:"required"`
}

// UpdateAcceptanceCriteriaRequest represents the request to update acceptance criteria
type UpdateAcceptanceCriteriaRequest struct {
	Description *string `json:"description,omitempty"`
}

// AcceptanceCriteriaFilters represents filters for listing acceptance criteria
type AcceptanceCriteriaFilters struct {
	UserStoryID *uuid.UUID `json:"user_story_id,omitempty"`
	AuthorID    *uuid.UUID `json:"author_id,omitempty"`
	OrderBy     string     `json:"order_by,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	Offset      int        `json:"offset,omitempty"`
}

// acceptanceCriteriaService implements AcceptanceCriteriaService interface
type acceptanceCriteriaService struct {
	acceptanceCriteriaRepo repository.AcceptanceCriteriaRepository
	userStoryRepo          repository.UserStoryRepository
	userRepo               repository.UserRepository
}

// NewAcceptanceCriteriaService creates a new acceptance criteria service instance
func NewAcceptanceCriteriaService(
	acceptanceCriteriaRepo repository.AcceptanceCriteriaRepository,
	userStoryRepo repository.UserStoryRepository,
	userRepo repository.UserRepository,
) AcceptanceCriteriaService {
	return &acceptanceCriteriaService{
		acceptanceCriteriaRepo: acceptanceCriteriaRepo,
		userStoryRepo:          userStoryRepo,
		userRepo:               userRepo,
	}
}

// CreateAcceptanceCriteria creates new acceptance criteria
func (s *acceptanceCriteriaService) CreateAcceptanceCriteria(req CreateAcceptanceCriteriaRequest) (*models.AcceptanceCriteria, error) {
	// Validate user story exists
	if exists, err := s.userStoryRepo.Exists(req.UserStoryID); err != nil {
		return nil, fmt.Errorf("failed to check user story existence: %w", err)
	} else if !exists {
		return nil, ErrUserStoryNotFound
	}

	// Validate author exists
	if exists, err := s.userRepo.Exists(req.AuthorID); err != nil {
		return nil, fmt.Errorf("failed to check author existence: %w", err)
	} else if !exists {
		return nil, ErrUserNotFound
	}

	acceptanceCriteria := &models.AcceptanceCriteria{
		ID:          uuid.New(),
		UserStoryID: req.UserStoryID,
		AuthorID:    req.AuthorID,
		Description: req.Description,
	}

	if err := s.acceptanceCriteriaRepo.Create(acceptanceCriteria); err != nil {
		return nil, fmt.Errorf("failed to create acceptance criteria: %w", err)
	}

	return acceptanceCriteria, nil
}

// GetAcceptanceCriteriaByID retrieves acceptance criteria by its ID with UserStory and Author preloaded
func (s *acceptanceCriteriaService) GetAcceptanceCriteriaByID(id uuid.UUID) (*models.AcceptanceCriteria, error) {
	acceptanceCriteria, err := s.acceptanceCriteriaRepo.GetByIDWithPreloads(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrAcceptanceCriteriaNotFound
		}
		return nil, fmt.Errorf("failed to get acceptance criteria: %w", err)
	}
	return acceptanceCriteria, nil
}

// GetAcceptanceCriteriaByReferenceID retrieves acceptance criteria by its reference ID
func (s *acceptanceCriteriaService) GetAcceptanceCriteriaByReferenceID(referenceID string) (*models.AcceptanceCriteria, error) {
	acceptanceCriteria, err := s.acceptanceCriteriaRepo.GetByReferenceID(referenceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrAcceptanceCriteriaNotFound
		}
		return nil, fmt.Errorf("failed to get acceptance criteria: %w", err)
	}
	return acceptanceCriteria, nil
}

// UpdateAcceptanceCriteria updates existing acceptance criteria
func (s *acceptanceCriteriaService) UpdateAcceptanceCriteria(id uuid.UUID, req UpdateAcceptanceCriteriaRequest) (*models.AcceptanceCriteria, error) {
	acceptanceCriteria, err := s.acceptanceCriteriaRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrAcceptanceCriteriaNotFound
		}
		return nil, fmt.Errorf("failed to get acceptance criteria: %w", err)
	}

	// Update fields if provided
	if req.Description != nil {
		acceptanceCriteria.Description = *req.Description
	}

	if err := s.acceptanceCriteriaRepo.Update(acceptanceCriteria); err != nil {
		return nil, fmt.Errorf("failed to update acceptance criteria: %w", err)
	}

	return acceptanceCriteria, nil
}

// DeleteAcceptanceCriteria deletes acceptance criteria with dependency validation
func (s *acceptanceCriteriaService) DeleteAcceptanceCriteria(id uuid.UUID, force bool) error {
	// Check if acceptance criteria exists
	acceptanceCriteria, err := s.acceptanceCriteriaRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrAcceptanceCriteriaNotFound
		}
		return fmt.Errorf("failed to get acceptance criteria: %w", err)
	}

	// Check for requirements unless force delete
	if !force {
		hasRequirements, err := s.acceptanceCriteriaRepo.HasRequirements(id)
		if err != nil {
			return fmt.Errorf("failed to check requirements: %w", err)
		}
		if hasRequirements {
			return ErrAcceptanceCriteriaHasRequirements
		}
	}

	// Check if this is the last acceptance criteria for the user story
	if !force {
		count, err := s.acceptanceCriteriaRepo.CountByUserStory(acceptanceCriteria.UserStoryID)
		if err != nil {
			return fmt.Errorf("failed to count acceptance criteria: %w", err)
		}
		if count <= 1 {
			return ErrUserStoryMustHaveAcceptanceCriteria
		}
	}

	// Delete the acceptance criteria (cascade will handle requirements if force=true)
	if err := s.acceptanceCriteriaRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete acceptance criteria: %w", err)
	}

	return nil
}

// ListAcceptanceCriteria retrieves acceptance criteria with optional filtering and relationships preloaded
func (s *acceptanceCriteriaService) ListAcceptanceCriteria(filters AcceptanceCriteriaFilters) ([]models.AcceptanceCriteria, int64, error) {
	// Build filter map
	filterMap := make(map[string]interface{})

	if filters.UserStoryID != nil {
		filterMap["user_story_id"] = *filters.UserStoryID
	}
	if filters.AuthorID != nil {
		filterMap["author_id"] = *filters.AuthorID
	}

	// Get total count with filters
	totalCount, err := s.acceptanceCriteriaRepo.Count(filterMap)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count acceptance criteria: %w", err)
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

	// Always use the method with preloads to include UserStory and Author by default
	acceptanceCriteria, err := s.acceptanceCriteriaRepo.ListWithPreloads(filterMap, orderBy, limit, filters.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list acceptance criteria: %w", err)
	}

	return acceptanceCriteria, totalCount, nil
}

// GetAcceptanceCriteriaByUserStory retrieves acceptance criteria by user story ID with pagination
func (s *acceptanceCriteriaService) GetAcceptanceCriteriaByUserStory(userStoryID uuid.UUID, limit, offset int) ([]models.AcceptanceCriteria, int64, error) {
	// Validate user story exists
	if exists, err := s.userStoryRepo.Exists(userStoryID); err != nil {
		return nil, 0, fmt.Errorf("failed to check user story existence: %w", err)
	} else if !exists {
		return nil, 0, ErrUserStoryNotFound
	}

	// Set default limit if not provided
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Build filter for user story
	filterMap := map[string]interface{}{
		"user_story_id": userStoryID,
	}

	// Get total count for this user story
	totalCount, err := s.acceptanceCriteriaRepo.Count(filterMap)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count acceptance criteria by user story: %w", err)
	}

	// Get paginated results
	acceptanceCriteria, err := s.acceptanceCriteriaRepo.List(filterMap, "created_at DESC", limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get acceptance criteria by user story: %w", err)
	}

	return acceptanceCriteria, totalCount, nil
}

// GetAcceptanceCriteriaByAuthor retrieves acceptance criteria by author ID with pagination
func (s *acceptanceCriteriaService) GetAcceptanceCriteriaByAuthor(authorID uuid.UUID, limit, offset int) ([]models.AcceptanceCriteria, int64, error) {
	// Validate author exists
	if exists, err := s.userRepo.Exists(authorID); err != nil {
		return nil, 0, fmt.Errorf("failed to check author existence: %w", err)
	} else if !exists {
		return nil, 0, ErrUserNotFound
	}

	// Set default limit if not provided
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Build filter for author
	filterMap := map[string]interface{}{
		"author_id": authorID,
	}

	// Get total count for this author
	totalCount, err := s.acceptanceCriteriaRepo.Count(filterMap)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count acceptance criteria by author: %w", err)
	}

	// Get paginated results
	acceptanceCriteria, err := s.acceptanceCriteriaRepo.List(filterMap, "created_at DESC", limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get acceptance criteria by author: %w", err)
	}

	return acceptanceCriteria, totalCount, nil
}

// ValidateUserStoryHasAcceptanceCriteria validates that a user story has at least one acceptance criteria
func (s *acceptanceCriteriaService) ValidateUserStoryHasAcceptanceCriteria(userStoryID uuid.UUID) error {
	count, err := s.acceptanceCriteriaRepo.CountByUserStory(userStoryID)
	if err != nil {
		return fmt.Errorf("failed to count acceptance criteria: %w", err)
	}

	if count == 0 {
		return ErrUserStoryMustHaveAcceptanceCriteria
	}

	return nil
}
