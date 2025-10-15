package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

// Steering document specific errors (additional to common errors in errors.go)
var (
	ErrInvalidCreator = errors.New("invalid creator")
)

// SteeringDocumentService defines the interface for steering document business logic
type SteeringDocumentService interface {
	CreateSteeringDocument(req CreateSteeringDocumentRequest, currentUser *models.User) (*models.SteeringDocument, error)
	GetSteeringDocumentByID(id uuid.UUID, currentUser *models.User) (*models.SteeringDocument, error)
	GetSteeringDocumentByReferenceID(referenceID string, currentUser *models.User) (*models.SteeringDocument, error)
	UpdateSteeringDocument(id uuid.UUID, req UpdateSteeringDocumentRequest, currentUser *models.User) (*models.SteeringDocument, error)
	DeleteSteeringDocument(id uuid.UUID, currentUser *models.User) error
	ListSteeringDocuments(filters SteeringDocumentFilters, currentUser *models.User) ([]models.SteeringDocument, int64, error)
	SearchSteeringDocuments(query string, currentUser *models.User) ([]models.SteeringDocument, error)
	GetSteeringDocumentsByEpicID(epicID uuid.UUID, currentUser *models.User) ([]models.SteeringDocument, error)
	LinkSteeringDocumentToEpic(steeringDocumentID, epicID uuid.UUID, currentUser *models.User) error
	UnlinkSteeringDocumentFromEpic(steeringDocumentID, epicID uuid.UUID, currentUser *models.User) error
}

// CreateSteeringDocumentRequest represents the request to create a steering document
// @Description Request payload for creating a new steering document
type CreateSteeringDocumentRequest struct {
	// Title is the name/summary of the steering document
	// @Description Title or name of the steering document (required, max 500 characters)
	// @MaxLength 500
	// @Example "Code Review Standards"
	Title string `json:"title" binding:"required,max=500"`

	// Description provides detailed information about the steering document
	// @Description Detailed description of the steering document content (optional, max 50000 characters)
	// @MaxLength 50000
	// @Example "This document outlines the code review standards and practices for the development team..."
	Description *string `json:"description,omitempty" binding:"omitempty,max=50000"`
}

// UpdateSteeringDocumentRequest represents the request to update a steering document
// @Description Request payload for updating an existing steering document (all fields are optional)
type UpdateSteeringDocumentRequest struct {
	// Title is the name/summary of the steering document
	// @Description Title or name of the steering document (optional, max 500 characters)
	// @MaxLength 500
	// @Example "Enhanced Code Review Standards"
	Title *string `json:"title,omitempty" binding:"omitempty,max=500"`

	// Description provides detailed information about the steering document
	// @Description Detailed description of the steering document content (optional, max 50000 characters)
	// @MaxLength 50000
	// @Example "Enhanced document with additional security review requirements..."
	Description *string `json:"description,omitempty" binding:"omitempty,max=50000"`
}

// SteeringDocumentFilters represents filters for listing steering documents
// @Description Filters and pagination options for listing steering documents
type SteeringDocumentFilters struct {
	// CreatorID filters steering documents by creator
	// @Description Filter steering documents by creator UUID (optional)
	// @Example "123e4567-e89b-12d3-a456-426614174001"
	CreatorID *uuid.UUID `json:"creator_id,omitempty"`

	// Search filters steering documents by text search
	// @Description Search query for full-text search in title and description (optional)
	// @Example "code review"
	Search string `json:"search,omitempty"`

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

// steeringDocumentService implements SteeringDocumentService interface
type steeringDocumentService struct {
	steeringDocumentRepo repository.SteeringDocumentRepository
	epicRepo             repository.EpicRepository
	userRepo             repository.UserRepository
}

// NewSteeringDocumentService creates a new steering document service instance
func NewSteeringDocumentService(
	steeringDocumentRepo repository.SteeringDocumentRepository,
	epicRepo repository.EpicRepository,
	userRepo repository.UserRepository,
) SteeringDocumentService {
	return &steeringDocumentService{
		steeringDocumentRepo: steeringDocumentRepo,
		epicRepo:             epicRepo,
		userRepo:             userRepo,
	}
}

// CreateSteeringDocument creates a new steering document
func (s *steeringDocumentService) CreateSteeringDocument(req CreateSteeringDocumentRequest, currentUser *models.User) (*models.SteeringDocument, error) {
	// Authorization check: Only Administrator and User roles can create steering documents
	if !currentUser.CanEdit() {
		return nil, ErrUnauthorizedAccess
	}

	// Validate input
	if err := s.validateCreateSteeringDocumentRequest(req); err != nil {
		return nil, err
	}

	// Validate creator exists (should be the current user)
	if exists, err := s.userRepo.Exists(currentUser.ID); err != nil {
		return nil, fmt.Errorf("failed to check creator existence: %w", err)
	} else if !exists {
		return nil, ErrUserNotFound
	}

	doc := &models.SteeringDocument{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		CreatorID:   currentUser.ID,
	}

	if err := s.steeringDocumentRepo.Create(doc); err != nil {
		return nil, fmt.Errorf("failed to create steering document: %w", err)
	}

	return s.steeringDocumentRepo.GetByID(doc.ID)
}

// GetSteeringDocumentByID retrieves a steering document by its ID
func (s *steeringDocumentService) GetSteeringDocumentByID(id uuid.UUID, currentUser *models.User) (*models.SteeringDocument, error) {
	// Authorization check: Only Administrator and User roles can read steering documents
	if !currentUser.CanRead() {
		return nil, ErrUnauthorizedAccess
	}

	doc, err := s.steeringDocumentRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSteeringDocumentNotFound
		}
		return nil, fmt.Errorf("failed to get steering document: %w", err)
	}
	return doc, nil
}

// GetSteeringDocumentByReferenceID retrieves a steering document by its reference ID
func (s *steeringDocumentService) GetSteeringDocumentByReferenceID(referenceID string, currentUser *models.User) (*models.SteeringDocument, error) {
	// Authorization check: Only Administrator and User roles can read steering documents
	if !currentUser.CanRead() {
		return nil, ErrUnauthorizedAccess
	}

	doc, err := s.steeringDocumentRepo.GetByReferenceID(referenceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSteeringDocumentNotFound
		}
		return nil, fmt.Errorf("failed to get steering document: %w", err)
	}
	return doc, nil
}

// UpdateSteeringDocument updates an existing steering document
func (s *steeringDocumentService) UpdateSteeringDocument(id uuid.UUID, req UpdateSteeringDocumentRequest, currentUser *models.User) (*models.SteeringDocument, error) {
	// Get the existing document
	doc, err := s.steeringDocumentRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSteeringDocumentNotFound
		}
		return nil, fmt.Errorf("failed to get steering document: %w", err)
	}

	// Authorization check: Administrator can edit any document, User can only edit their own documents
	if !currentUser.IsAdministrator() && doc.CreatorID != currentUser.ID {
		return nil, ErrUnauthorizedAccess
	}

	// Update fields if provided
	if req.Title != nil {
		doc.Title = *req.Title
	}

	if req.Description != nil {
		doc.Description = req.Description
	}

	if err := s.steeringDocumentRepo.Update(doc); err != nil {
		return nil, fmt.Errorf("failed to update steering document: %w", err)
	}

	return s.steeringDocumentRepo.GetByID(id)
}

// DeleteSteeringDocument deletes a steering document
func (s *steeringDocumentService) DeleteSteeringDocument(id uuid.UUID, currentUser *models.User) error {
	// Get the existing document
	doc, err := s.steeringDocumentRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrSteeringDocumentNotFound
		}
		return fmt.Errorf("failed to get steering document: %w", err)
	}

	// Authorization check: Administrator can delete any document, User can only delete their own documents
	if !currentUser.IsAdministrator() && doc.CreatorID != currentUser.ID {
		return ErrUnauthorizedAccess
	}

	if err := s.steeringDocumentRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete steering document: %w", err)
	}

	return nil
}

// ListSteeringDocuments retrieves steering documents with optional filtering
func (s *steeringDocumentService) ListSteeringDocuments(filters SteeringDocumentFilters, currentUser *models.User) ([]models.SteeringDocument, int64, error) {
	// Authorization check: Only Administrator and User roles can list steering documents
	if !currentUser.CanRead() {
		return nil, 0, ErrUnauthorizedAccess
	}

	// Convert service filters to repository filters
	repoFilters := repository.SteeringDocumentFilters{
		CreatorID: filters.CreatorID,
		Search:    filters.Search,
		OrderBy:   filters.OrderBy,
		Limit:     filters.Limit,
		Offset:    filters.Offset,
	}

	// Non-administrators can only see their own documents
	if !currentUser.IsAdministrator() {
		repoFilters.CreatorID = &currentUser.ID
	}

	// Set default limit if not specified
	if repoFilters.Limit <= 0 {
		repoFilters.Limit = 50
	}

	// Set maximum limit
	if repoFilters.Limit > 100 {
		repoFilters.Limit = 100
	}

	return s.steeringDocumentRepo.ListWithFilters(repoFilters)
}

// SearchSteeringDocuments performs full-text search on steering documents
func (s *steeringDocumentService) SearchSteeringDocuments(query string, currentUser *models.User) ([]models.SteeringDocument, error) {
	// All authenticated users can search steering documents
	return s.steeringDocumentRepo.Search(query)
}

// GetSteeringDocumentsByEpicID retrieves all steering documents linked to a specific epic
func (s *steeringDocumentService) GetSteeringDocumentsByEpicID(epicID uuid.UUID, currentUser *models.User) ([]models.SteeringDocument, error) {
	// All authenticated users can view steering documents linked to epics

	// Verify epic exists
	_, err := s.epicRepo.GetByID(epicID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEpicNotFound
		}
		return nil, fmt.Errorf("failed to get epic: %w", err)
	}

	return s.steeringDocumentRepo.GetByEpicID(epicID)
}

// LinkSteeringDocumentToEpic creates a link between a steering document and an epic
func (s *steeringDocumentService) LinkSteeringDocumentToEpic(steeringDocumentID, epicID uuid.UUID, currentUser *models.User) error {
	// Authorization check: Only Administrator and User roles can link documents to epics
	if !currentUser.CanEdit() {
		return ErrUnauthorizedAccess
	}

	// Verify steering document exists
	doc, err := s.steeringDocumentRepo.GetByID(steeringDocumentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrSteeringDocumentNotFound
		}
		return fmt.Errorf("failed to get steering document: %w", err)
	}

	// Additional authorization check: User can only link their own documents (Administrator can link any)
	if !currentUser.IsAdministrator() && doc.CreatorID != currentUser.ID {
		return ErrUnauthorizedAccess
	}

	// Verify epic exists
	if exists, err := s.epicRepo.Exists(epicID); err != nil {
		return fmt.Errorf("failed to check epic existence: %w", err)
	} else if !exists {
		return ErrEpicNotFound
	}

	// Create the link
	if err := s.steeringDocumentRepo.LinkToEpic(steeringDocumentID, epicID); err != nil {
		// Check if it's a "link already exists" error
		if err.Error() == fmt.Sprintf("link already exists between steering document %s and epic %s", steeringDocumentID, epicID) {
			return ErrLinkAlreadyExists
		}
		return fmt.Errorf("failed to link steering document to epic: %w", err)
	}

	return nil
}

// UnlinkSteeringDocumentFromEpic removes the link between a steering document and an epic
func (s *steeringDocumentService) UnlinkSteeringDocumentFromEpic(steeringDocumentID, epicID uuid.UUID, currentUser *models.User) error {
	// Authorization check: Only Administrator and User roles can unlink documents from epics
	if !currentUser.CanEdit() {
		return ErrUnauthorizedAccess
	}

	// Verify steering document exists
	doc, err := s.steeringDocumentRepo.GetByID(steeringDocumentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrSteeringDocumentNotFound
		}
		return fmt.Errorf("failed to get steering document: %w", err)
	}

	// Additional authorization check: User can only unlink their own documents (Administrator can unlink any)
	if !currentUser.IsAdministrator() && doc.CreatorID != currentUser.ID {
		return ErrUnauthorizedAccess
	}

	// Verify epic exists (optional check, but good for consistency)
	_, err = s.epicRepo.GetByID(epicID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrEpicNotFound
		}
		return fmt.Errorf("failed to get epic: %w", err)
	}

	// Remove the link
	if err := s.steeringDocumentRepo.UnlinkFromEpic(steeringDocumentID, epicID); err != nil {
		return fmt.Errorf("failed to unlink steering document from epic: %w", err)
	}

	return nil
}

// validateCreateSteeringDocumentRequest validates the create steering document request
func (s *steeringDocumentService) validateCreateSteeringDocumentRequest(req CreateSteeringDocumentRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return fmt.Errorf("title is required")
	}

	if len(req.Title) > 500 {
		return fmt.Errorf("title must be at most 500 characters")
	}

	if req.Description != nil && len(*req.Description) > 50000 {
		return fmt.Errorf("description must be at most 50000 characters")
	}

	return nil
}
