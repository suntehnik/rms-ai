package repository

import (
	"errors"
	"fmt"
	"product-requirements-management/internal/models"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// steeringDocumentRepository implements SteeringDocumentRepository
type steeringDocumentRepository struct {
	db *gorm.DB
}

// NewSteeringDocumentRepository creates a new steering document repository
func NewSteeringDocumentRepository(db *gorm.DB) SteeringDocumentRepository {
	return &steeringDocumentRepository{db: db}
}

// Create creates a new steering document
func (r *steeringDocumentRepository) Create(doc *models.SteeringDocument) error {
	if err := r.db.Create(doc).Error; err != nil {
		return fmt.Errorf("failed to create steering document: %w", err)
	}
	return nil
}

// GetByID retrieves a steering document by ID
func (r *steeringDocumentRepository) GetByID(id uuid.UUID) (*models.SteeringDocument, error) {
	var doc models.SteeringDocument
	err := r.db.Preload("Creator").First(&doc, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get steering document by ID: %w", err)
	}
	return &doc, nil
}

// GetByReferenceID retrieves a steering document by reference ID
func (r *steeringDocumentRepository) GetByReferenceID(referenceID string) (*models.SteeringDocument, error) {
	var doc models.SteeringDocument
	err := r.db.Preload("Creator").First(&doc, "reference_id = ?", referenceID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get steering document by reference ID: %w", err)
	}
	return &doc, nil
}

// GetByReferenceIDCaseInsensitive retrieves a steering document by reference ID (case-insensitive)
func (r *steeringDocumentRepository) GetByReferenceIDCaseInsensitive(referenceID string) (*models.SteeringDocument, error) {
	var doc models.SteeringDocument

	query := r.db.Preload("Creator")
	var err error

	// Use ILIKE for PostgreSQL, LOWER() LIKE for SQLite compatibility
	if r.db.Dialector.Name() == "postgres" {
		err = query.Where("reference_id ILIKE ?", referenceID).First(&doc).Error
	} else {
		// SQLite and other databases - use LOWER() LIKE for case-insensitive matching
		err = query.Where("LOWER(reference_id) LIKE LOWER(?)", referenceID).First(&doc).Error
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get steering document by reference ID (case-insensitive): %w", err)
	}
	return &doc, nil
}

// Update updates a steering document
func (r *steeringDocumentRepository) Update(doc *models.SteeringDocument) error {
	if err := r.db.Save(doc).Error; err != nil {
		return fmt.Errorf("failed to update steering document: %w", err)
	}
	return nil
}

// Delete deletes a steering document
func (r *steeringDocumentRepository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&models.SteeringDocument{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete steering document: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListWithFilters retrieves steering documents with optional filtering
func (r *steeringDocumentRepository) ListWithFilters(filters SteeringDocumentFilters) ([]models.SteeringDocument, int64, error) {
	var docs []models.SteeringDocument
	var totalCount int64

	query := r.db.Model(&models.SteeringDocument{}).Preload("Creator")

	// Apply filters
	if filters.CreatorID != nil {
		query = query.Where("creator_id = ?", *filters.CreatorID)
	}

	if filters.Search != "" {
		searchTerm := "%" + strings.ToLower(filters.Search) + "%"
		query = query.Where(
			"LOWER(title) LIKE ? OR LOWER(description) LIKE ? OR LOWER(reference_id) LIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
	}

	// Count total records
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count steering documents: %w", err)
	}

	// Apply ordering
	orderBy := "created_at DESC"
	if filters.OrderBy != "" {
		switch filters.OrderBy {
		case "title", "title_asc":
			orderBy = "title ASC"
		case "title_desc":
			orderBy = "title DESC"
		case "created_at", "created_at_desc":
			orderBy = "created_at DESC"
		case "created_at_asc":
			orderBy = "created_at ASC"
		case "updated_at", "updated_at_desc":
			orderBy = "updated_at DESC"
		case "updated_at_asc":
			orderBy = "updated_at ASC"
		case "reference_id", "reference_id_asc":
			orderBy = "reference_id ASC"
		case "reference_id_desc":
			orderBy = "reference_id DESC"
		}
	}
	query = query.Order(orderBy)

	// Apply pagination
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if err := query.Find(&docs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list steering documents: %w", err)
	}

	return docs, totalCount, nil
}

// Search performs full-text search on steering documents
func (r *steeringDocumentRepository) Search(query string) ([]models.SteeringDocument, error) {
	var docs []models.SteeringDocument

	searchQuery := r.db.Model(&models.SteeringDocument{}).
		Preload("Creator").
		Where("to_tsvector('english', title || ' ' || COALESCE(description, '')) @@ plainto_tsquery('english', ?)", query).
		Order("ts_rank(to_tsvector('english', title || ' ' || COALESCE(description, '')), plainto_tsquery('english', ?)) DESC")

	if err := searchQuery.Find(&docs).Error; err != nil {
		return nil, fmt.Errorf("failed to search steering documents: %w", err)
	}

	return docs, nil
}

// GetByEpicID retrieves steering documents linked to an epic
func (r *steeringDocumentRepository) GetByEpicID(epicID uuid.UUID) ([]models.SteeringDocument, error) {
	var docs []models.SteeringDocument

	err := r.db.
		Preload("Creator").
		Joins("JOIN epic_steering_documents esd ON steering_documents.id = esd.steering_document_id").
		Where("esd.epic_id = ?", epicID).
		Find(&docs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get steering documents by epic ID: %w", err)
	}

	return docs, nil
}

// GetByEpicIDWithPagination retrieves steering documents linked to an epic with pagination
func (r *steeringDocumentRepository) GetByEpicIDWithPagination(epicID uuid.UUID, limit, offset int) ([]models.SteeringDocument, int64, error) {
	var docs []models.SteeringDocument
	var totalCount int64

	// Count total records
	countQuery := r.db.Model(&models.SteeringDocument{}).
		Joins("JOIN epic_steering_documents esd ON steering_documents.id = esd.steering_document_id").
		Where("esd.epic_id = ?", epicID)

	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count steering documents by epic ID: %w", err)
	}

	// Get paginated results
	query := r.db.
		Preload("Creator").
		Joins("JOIN epic_steering_documents esd ON steering_documents.id = esd.steering_document_id").
		Where("esd.epic_id = ?", epicID).
		Order("steering_documents.created_at DESC")

	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&docs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get steering documents by epic ID with pagination: %w", err)
	}

	return docs, totalCount, nil
}

// LinkToEpic creates a link between a steering document and an epic
func (r *steeringDocumentRepository) LinkToEpic(steeringDocumentID, epicID uuid.UUID) error {
	link := models.EpicSteeringDocument{
		EpicID:             epicID,
		SteeringDocumentID: steeringDocumentID,
	}

	if err := r.db.Create(&link).Error; err != nil {
		// Check if it's a duplicate key error
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint") {
			return fmt.Errorf("steering document is already linked to this epic")
		}
		return fmt.Errorf("failed to link steering document to epic: %w", err)
	}

	return nil
}

// UnlinkFromEpic removes the link between a steering document and an epic
func (r *steeringDocumentRepository) UnlinkFromEpic(steeringDocumentID, epicID uuid.UUID) error {
	result := r.db.Where("steering_document_id = ? AND epic_id = ?", steeringDocumentID, epicID).
		Delete(&models.EpicSteeringDocument{})

	if result.Error != nil {
		return fmt.Errorf("failed to unlink steering document from epic: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("link between steering document and epic not found")
	}

	return nil
}

// WithTransaction executes a function within a database transaction
func (r *steeringDocumentRepository) WithTransaction(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// List retrieves steering documents with basic filtering
func (r *steeringDocumentRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.SteeringDocument, error) {
	var docs []models.SteeringDocument
	query := r.db.Model(&models.SteeringDocument{}).Preload("Creator")

	// Apply filters
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	// Apply ordering
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&docs).Error; err != nil {
		return nil, fmt.Errorf("failed to list steering documents: %w", err)
	}

	return docs, nil
}

// Count returns the total number of steering documents matching the filters
func (r *steeringDocumentRepository) Count(filters map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.Model(&models.SteeringDocument{})

	// Apply filters
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count steering documents: %w", err)
	}

	return count, nil
}

// Exists checks if a steering document exists by ID
func (r *steeringDocumentRepository) Exists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.SteeringDocument{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check steering document existence: %w", err)
	}
	return count > 0, nil
}

// ExistsByReferenceID checks if a steering document exists by reference ID
func (r *steeringDocumentRepository) ExistsByReferenceID(referenceID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.SteeringDocument{}).Where("reference_id = ?", referenceID).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check steering document existence by reference ID: %w", err)
	}
	return count > 0, nil
}

// GetDB returns the underlying GORM database instance
func (r *steeringDocumentRepository) GetDB() *gorm.DB {
	return r.db
}
