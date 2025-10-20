package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

var (
	ErrInvalidNavigationEntityType = errors.New("invalid entity type")
)

// NavigationService defines the interface for hierarchical navigation business logic
type NavigationService interface {
	GetHierarchy(filters HierarchyFilters) (*HierarchyResponse, error)
	GetEpicHierarchy(epicID uuid.UUID, expand, orderBy, orderDirection string) (*EpicHierarchy, error)
	GetUserStoryHierarchy(userStoryID uuid.UUID, expand, orderBy, orderDirection string) (*UserStoryHierarchy, error)
	GetEntityPath(entityType string, entityID uuid.UUID) ([]PathElement, error)
	GetEpicByReferenceID(referenceID string) (*models.Epic, error)
	GetUserStoryByReferenceID(referenceID string) (*models.UserStory, error)
	ResolveReferenceID(entityType, referenceID string) (uuid.UUID, error)
}

// HierarchyFilters represents filters for hierarchy queries
type HierarchyFilters struct {
	CreatorID      *uuid.UUID       `json:"creator_id,omitempty"`
	AssigneeID     *uuid.UUID       `json:"assignee_id,omitempty"`
	Status         *string          `json:"status,omitempty"`
	Priority       *models.Priority `json:"priority,omitempty"`
	OrderBy        string           `json:"order_by,omitempty"`
	OrderDirection string           `json:"order_direction,omitempty"`
	Limit          int              `json:"limit,omitempty"`
	Offset         int              `json:"offset,omitempty"`
	Expand         string           `json:"expand,omitempty"`
}

// HierarchyResponse represents the complete hierarchy response
type HierarchyResponse struct {
	Epics []EpicHierarchy `json:"epics"`
	Total int             `json:"total"`
	Count int             `json:"count"`
}

// EpicHierarchy represents an epic with its hierarchical children
type EpicHierarchy struct {
	models.Epic
	UserStories []UserStoryHierarchy `json:"user_stories"`
}

// UserStoryHierarchy represents a user story with its hierarchical children
type UserStoryHierarchy struct {
	models.UserStory
	AcceptanceCriteria []models.AcceptanceCriteria `json:"acceptance_criteria"`
	Requirements       []RequirementHierarchy      `json:"requirements"`
}

// RequirementHierarchy represents a requirement with its relationships
type RequirementHierarchy struct {
	models.Requirement
	Relationships []models.RequirementRelationship `json:"relationships"`
}

// UnmarshalJSON implements custom JSON unmarshaling for EpicHierarchy
func (eh *EpicHierarchy) UnmarshalJSON(data []byte) error {
	type Alias EpicHierarchy
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(eh),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Ensure UserStories is never nil
	if eh.UserStories == nil {
		eh.UserStories = make([]UserStoryHierarchy, 0)
	}

	return nil
}

// UnmarshalJSON implements custom JSON unmarshaling for UserStoryHierarchy
func (ush *UserStoryHierarchy) UnmarshalJSON(data []byte) error {
	type Alias UserStoryHierarchy
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(ush),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Ensure slices are never nil
	if ush.Requirements == nil {
		ush.Requirements = make([]RequirementHierarchy, 0)
	}
	if ush.AcceptanceCriteria == nil {
		ush.AcceptanceCriteria = make([]models.AcceptanceCriteria, 0)
	}

	return nil
}

// UnmarshalJSON implements custom JSON unmarshaling for RequirementHierarchy
func (rh *RequirementHierarchy) UnmarshalJSON(data []byte) error {
	type Alias RequirementHierarchy
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(rh),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Ensure Relationships is never nil
	if rh.Relationships == nil {
		rh.Relationships = make([]models.RequirementRelationship, 0)
	}

	return nil
}

// PathElement represents an element in the entity path
type PathElement struct {
	ID          uuid.UUID `json:"id"`
	ReferenceID string    `json:"reference_id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
}

// navigationService implements NavigationService
type navigationService struct {
	epicRepo               repository.EpicRepository
	userStoryRepo          repository.UserStoryRepository
	acceptanceCriteriaRepo repository.AcceptanceCriteriaRepository
	requirementRepo        repository.RequirementRepository
	relationshipRepo       repository.RequirementRelationshipRepository
	userRepo               repository.UserRepository
}

// NewNavigationService creates a new navigation service instance
func NewNavigationService(
	epicRepo repository.EpicRepository,
	userStoryRepo repository.UserStoryRepository,
	acceptanceCriteriaRepo repository.AcceptanceCriteriaRepository,
	requirementRepo repository.RequirementRepository,
	relationshipRepo repository.RequirementRelationshipRepository,
	userRepo repository.UserRepository,
) NavigationService {
	return &navigationService{
		epicRepo:               epicRepo,
		userStoryRepo:          userStoryRepo,
		acceptanceCriteriaRepo: acceptanceCriteriaRepo,
		requirementRepo:        requirementRepo,
		relationshipRepo:       relationshipRepo,
		userRepo:               userRepo,
	}
}

// GetHierarchy returns the complete hierarchy with filtering and sorting
func (s *navigationService) GetHierarchy(filters HierarchyFilters) (*HierarchyResponse, error) {
	// Convert filters to epic filters
	epicFilters := EpicFilters{
		CreatorID:  filters.CreatorID,
		AssigneeID: filters.AssigneeID,
		Priority:   filters.Priority,
		OrderBy:    filters.OrderBy,
		Limit:      filters.Limit,
		Offset:     filters.Offset,
	}

	// Handle status filtering for epics
	if filters.Status != nil {
		epicStatus := models.EpicStatus(*filters.Status)
		epicFilters.Status = &epicStatus
	}

	// Build filter map for repository
	filterMap := make(map[string]interface{})

	if epicFilters.CreatorID != nil {
		filterMap["creator_id"] = *epicFilters.CreatorID
	}
	if epicFilters.AssigneeID != nil {
		filterMap["assignee_id"] = *epicFilters.AssigneeID
	}
	if epicFilters.Status != nil {
		filterMap["status"] = *epicFilters.Status
	}
	if epicFilters.Priority != nil {
		filterMap["priority"] = *epicFilters.Priority
	}

	// Set default ordering
	orderBy := "created_at DESC"
	if epicFilters.OrderBy != "" {
		orderBy = epicFilters.OrderBy
	}

	// Set default limit
	limit := 50
	if epicFilters.Limit > 0 {
		limit = epicFilters.Limit
	}

	// Get epics with the filters
	epics, err := s.epicRepo.List(filterMap, orderBy, limit, epicFilters.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list epics: %w", err)
	}

	// Build hierarchy response
	hierarchyEpics := make([]EpicHierarchy, 0, len(epics))

	for _, epic := range epics {
		epicHierarchy := EpicHierarchy{
			Epic:        epic,
			UserStories: make([]UserStoryHierarchy, 0), // Initialize empty slice
		}

		// Check if we should expand user stories
		if shouldExpand(filters.Expand, "user_stories") {
			userStories, err := s.getUserStoriesForEpic(epic.ID, filters.OrderBy, filters.OrderDirection)
			if err != nil {
				return nil, fmt.Errorf("failed to get user stories for epic %s: %w", epic.ID, err)
			}

			// Build user story hierarchies
			for _, userStory := range userStories {
				userStoryHierarchy := UserStoryHierarchy{
					UserStory:          userStory,
					Requirements:       make([]RequirementHierarchy, 0),      // Initialize empty slice
					AcceptanceCriteria: make([]models.AcceptanceCriteria, 0), // Initialize empty slice
				}

				// Check if we should expand requirements and acceptance criteria
				if shouldExpand(filters.Expand, "requirements") {
					requirements, err := s.getRequirementsForUserStory(userStory.ID, filters.OrderBy, filters.OrderDirection)
					if err != nil {
						return nil, fmt.Errorf("failed to get requirements for user story %s: %w", userStory.ID, err)
					}

					// Build requirement hierarchies
					for _, requirement := range requirements {
						reqHierarchy := RequirementHierarchy{
							Requirement:   requirement,
							Relationships: make([]models.RequirementRelationship, 0), // Initialize empty slice
						}

						if shouldExpand(filters.Expand, "relationships") {
							relationships, err := s.relationshipRepo.GetByRequirement(requirement.ID)
							if err != nil {
								return nil, fmt.Errorf("failed to get relationships for requirement %s: %w", requirement.ID, err)
							}
							reqHierarchy.Relationships = relationships
						}

						userStoryHierarchy.Requirements = append(userStoryHierarchy.Requirements, reqHierarchy)
					}
				}

				if shouldExpand(filters.Expand, "acceptance_criteria") {
					acceptanceCriteria, err := s.acceptanceCriteriaRepo.GetByUserStory(userStory.ID)
					if err != nil {
						return nil, fmt.Errorf("failed to get acceptance criteria for user story %s: %w", userStory.ID, err)
					}
					userStoryHierarchy.AcceptanceCriteria = acceptanceCriteria
				}

				epicHierarchy.UserStories = append(epicHierarchy.UserStories, userStoryHierarchy)
			}
		}

		hierarchyEpics = append(hierarchyEpics, epicHierarchy)
	}

	return &HierarchyResponse{
		Epics: hierarchyEpics,
		Total: len(hierarchyEpics), // In a real implementation, this would be the total count without pagination
		Count: len(hierarchyEpics),
	}, nil
}

// GetEpicHierarchy returns a single epic with its complete hierarchy
func (s *navigationService) GetEpicHierarchy(epicID uuid.UUID, expand, orderBy, orderDirection string) (*EpicHierarchy, error) {
	epic, err := s.epicRepo.GetByIDWithUsers(epicID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEpicNotFound
		}
		return nil, fmt.Errorf("failed to get epic: %w", err)
	}

	epicHierarchy := &EpicHierarchy{
		Epic:        *epic,
		UserStories: make([]UserStoryHierarchy, 0), // Initialize empty slice
	}

	// Always expand user stories for single epic view
	userStories, err := s.getUserStoriesForEpic(epicID, orderBy, orderDirection)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stories: %w", err)
	}

	for _, userStory := range userStories {
		userStoryHierarchy := UserStoryHierarchy{
			UserStory:          userStory,
			Requirements:       make([]RequirementHierarchy, 0),      // Initialize empty slice
			AcceptanceCriteria: make([]models.AcceptanceCriteria, 0), // Initialize empty slice
		}

		// Expand requirements if requested
		if shouldExpand(expand, "requirements") {
			requirements, err := s.getRequirementsForUserStory(userStory.ID, orderBy, orderDirection)
			if err != nil {
				return nil, fmt.Errorf("failed to get requirements: %w", err)
			}

			for _, requirement := range requirements {
				reqHierarchy := RequirementHierarchy{
					Requirement:   requirement,
					Relationships: make([]models.RequirementRelationship, 0), // Initialize empty slice
				}

				if shouldExpand(expand, "relationships") {
					relationships, err := s.relationshipRepo.GetByRequirement(requirement.ID)
					if err != nil {
						return nil, fmt.Errorf("failed to get relationships: %w", err)
					}
					reqHierarchy.Relationships = relationships
				}

				userStoryHierarchy.Requirements = append(userStoryHierarchy.Requirements, reqHierarchy)
			}
		}

		// Expand acceptance criteria if requested
		if shouldExpand(expand, "acceptance_criteria") {
			acceptanceCriteria, err := s.acceptanceCriteriaRepo.GetByUserStory(userStory.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get acceptance criteria: %w", err)
			}
			userStoryHierarchy.AcceptanceCriteria = acceptanceCriteria
		}

		epicHierarchy.UserStories = append(epicHierarchy.UserStories, userStoryHierarchy)
	}

	return epicHierarchy, nil
}

// GetUserStoryHierarchy returns a single user story with its complete hierarchy
func (s *navigationService) GetUserStoryHierarchy(userStoryID uuid.UUID, expand, orderBy, orderDirection string) (*UserStoryHierarchy, error) {
	userStory, err := s.userStoryRepo.GetByIDWithUsers(userStoryID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserStoryNotFound
		}
		return nil, fmt.Errorf("failed to get user story: %w", err)
	}

	userStoryHierarchy := &UserStoryHierarchy{
		UserStory:          *userStory,
		Requirements:       make([]RequirementHierarchy, 0),      // Initialize empty slice
		AcceptanceCriteria: make([]models.AcceptanceCriteria, 0), // Initialize empty slice
	}

	// Expand requirements if requested
	if shouldExpand(expand, "requirements") {
		requirements, err := s.getRequirementsForUserStory(userStoryID, orderBy, orderDirection)
		if err != nil {
			return nil, fmt.Errorf("failed to get requirements: %w", err)
		}

		for _, requirement := range requirements {
			reqHierarchy := RequirementHierarchy{
				Requirement:   requirement,
				Relationships: make([]models.RequirementRelationship, 0), // Initialize empty slice
			}

			if shouldExpand(expand, "relationships") {
				relationships, err := s.relationshipRepo.GetByRequirement(requirement.ID)
				if err != nil {
					return nil, fmt.Errorf("failed to get relationships: %w", err)
				}
				reqHierarchy.Relationships = relationships
			}

			userStoryHierarchy.Requirements = append(userStoryHierarchy.Requirements, reqHierarchy)
		}
	}

	// Expand acceptance criteria if requested
	if shouldExpand(expand, "acceptance_criteria") {
		acceptanceCriteria, err := s.acceptanceCriteriaRepo.GetByUserStory(userStoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to get acceptance criteria: %w", err)
		}
		userStoryHierarchy.AcceptanceCriteria = acceptanceCriteria
	}

	return userStoryHierarchy, nil
}

// GetEntityPath returns the hierarchical path to an entity
func (s *navigationService) GetEntityPath(entityType string, entityID uuid.UUID) ([]PathElement, error) {
	var path []PathElement

	switch entityType {
	case "requirement":
		requirement, err := s.requirementRepo.GetByID(entityID)
		if err != nil {
			return nil, fmt.Errorf("failed to get requirement: %w", err)
		}

		// Add requirement to path
		path = append(path, PathElement{
			ID:          requirement.ID,
			ReferenceID: requirement.ReferenceID,
			Type:        "requirement",
			Title:       requirement.Title,
		})

		// Get user story
		userStory, err := s.userStoryRepo.GetByID(requirement.UserStoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user story: %w", err)
		}

		// Add user story to path
		path = append([]PathElement{{
			ID:          userStory.ID,
			ReferenceID: userStory.ReferenceID,
			Type:        "user_story",
			Title:       userStory.Title,
		}}, path...)

		// Get epic
		epic, err := s.epicRepo.GetByID(userStory.EpicID)
		if err != nil {
			return nil, fmt.Errorf("failed to get epic: %w", err)
		}

		// Add epic to path
		path = append([]PathElement{{
			ID:          epic.ID,
			ReferenceID: epic.ReferenceID,
			Type:        "epic",
			Title:       epic.Title,
		}}, path...)

	case "acceptance_criteria":
		acceptanceCriteria, err := s.acceptanceCriteriaRepo.GetByID(entityID)
		if err != nil {
			return nil, fmt.Errorf("failed to get acceptance criteria: %w", err)
		}

		// Add acceptance criteria to path
		path = append(path, PathElement{
			ID:          acceptanceCriteria.ID,
			ReferenceID: acceptanceCriteria.ReferenceID,
			Type:        "acceptance_criteria",
			Title:       acceptanceCriteria.Description[:minInt(50, len(acceptanceCriteria.Description))] + "...",
		})

		// Get user story
		userStory, err := s.userStoryRepo.GetByID(acceptanceCriteria.UserStoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user story: %w", err)
		}

		// Add user story to path
		path = append([]PathElement{{
			ID:          userStory.ID,
			ReferenceID: userStory.ReferenceID,
			Type:        "user_story",
			Title:       userStory.Title,
		}}, path...)

		// Get epic
		epic, err := s.epicRepo.GetByID(userStory.EpicID)
		if err != nil {
			return nil, fmt.Errorf("failed to get epic: %w", err)
		}

		// Add epic to path
		path = append([]PathElement{{
			ID:          epic.ID,
			ReferenceID: epic.ReferenceID,
			Type:        "epic",
			Title:       epic.Title,
		}}, path...)

	case "user_story":
		userStory, err := s.userStoryRepo.GetByID(entityID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user story: %w", err)
		}

		// Add user story to path
		path = append(path, PathElement{
			ID:          userStory.ID,
			ReferenceID: userStory.ReferenceID,
			Type:        "user_story",
			Title:       userStory.Title,
		})

		// Get epic
		epic, err := s.epicRepo.GetByID(userStory.EpicID)
		if err != nil {
			return nil, fmt.Errorf("failed to get epic: %w", err)
		}

		// Add epic to path
		path = append([]PathElement{{
			ID:          epic.ID,
			ReferenceID: epic.ReferenceID,
			Type:        "epic",
			Title:       epic.Title,
		}}, path...)

	case "epic":
		epic, err := s.epicRepo.GetByID(entityID)
		if err != nil {
			return nil, fmt.Errorf("failed to get epic: %w", err)
		}

		// Add epic to path
		path = append(path, PathElement{
			ID:          epic.ID,
			ReferenceID: epic.ReferenceID,
			Type:        "epic",
			Title:       epic.Title,
		})

	default:
		return nil, ErrInvalidNavigationEntityType
	}

	return path, nil
}

// GetEpicByReferenceID gets an epic by its reference ID
func (s *navigationService) GetEpicByReferenceID(referenceID string) (*models.Epic, error) {
	epic, err := s.epicRepo.GetByReferenceID(referenceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEpicNotFound
		}
		return nil, err
	}
	return epic, nil
}

// GetUserStoryByReferenceID gets a user story by its reference ID
func (s *navigationService) GetUserStoryByReferenceID(referenceID string) (*models.UserStory, error) {
	userStory, err := s.userStoryRepo.GetByReferenceID(referenceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserStoryNotFound
		}
		return nil, err
	}
	return userStory, nil
}

// ResolveReferenceID resolves a reference ID to UUID based on entity type
func (s *navigationService) ResolveReferenceID(entityType, referenceID string) (uuid.UUID, error) {
	switch entityType {
	case "epic":
		epic, err := s.epicRepo.GetByReferenceID(referenceID)
		if err != nil {
			return uuid.Nil, err
		}
		return epic.ID, nil
	case "user_story":
		userStory, err := s.userStoryRepo.GetByReferenceID(referenceID)
		if err != nil {
			return uuid.Nil, err
		}
		return userStory.ID, nil
	case "acceptance_criteria":
		acceptanceCriteria, err := s.acceptanceCriteriaRepo.GetByReferenceID(referenceID)
		if err != nil {
			return uuid.Nil, err
		}
		return acceptanceCriteria.ID, nil
	case "requirement":
		requirement, err := s.requirementRepo.GetByReferenceID(referenceID)
		if err != nil {
			return uuid.Nil, err
		}
		return requirement.ID, nil
	default:
		return uuid.Nil, ErrInvalidNavigationEntityType
	}
}

// Helper functions

// getUserStoriesForEpic gets user stories for an epic with sorting
func (s *navigationService) getUserStoriesForEpic(epicID uuid.UUID, _, _ string) ([]models.UserStory, error) {
	// Get user stories by epic ID - this will return user stories with populated users
	// since they're being used in hierarchy context where user info is expected
	return s.userStoryRepo.GetByEpic(epicID)
}

// getRequirementsForUserStory gets requirements for a user story with sorting
func (s *navigationService) getRequirementsForUserStory(userStoryID uuid.UUID, orderBy, orderDirection string) ([]models.Requirement, error) {
	// Use the repository method directly
	filterMap := make(map[string]interface{})
	filterMap["user_story_id"] = userStoryID

	orderByClause := "created_at DESC"
	if orderBy != "" {
		orderByClause = orderBy
		if orderDirection == "desc" {
			orderByClause += " DESC"
		}
	}

	return s.requirementRepo.List(filterMap, orderByClause, 100, 0)
}

// shouldExpand checks if a specific field should be expanded
func shouldExpand(expand, field string) bool {
	if expand == "" {
		return false
	}

	// Handle comma-separated expansion fields
	fields := strings.Split(expand, ",")
	for _, f := range fields {
		if strings.TrimSpace(f) == field {
			return true
		}
	}

	return false
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
