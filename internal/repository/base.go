package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotFound     = errors.New("record not found")
	ErrInvalidID    = errors.New("invalid ID format")
	ErrDuplicateKey = errors.New("duplicate key violation")
	ErrForeignKey   = errors.New("foreign key constraint violation")
)

// BaseRepository provides common CRUD operations for all entities
type BaseRepository[T any] struct {
	db *gorm.DB
}

// NewBaseRepository creates a new base repository instance
func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{
		db: db,
	}
}

// Create creates a new entity in the database
func (r *BaseRepository[T]) Create(entity *T) error {
	if err := r.db.Create(entity).Error; err != nil {
		return r.handleDBError(err)
	}
	return nil
}

// GetByID retrieves an entity by its UUID
func (r *BaseRepository[T]) GetByID(id uuid.UUID) (*T, error) {
	var entity T
	if err := r.db.Where("id = ?", id).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, r.handleDBError(err)
	}
	return &entity, nil
}

// GetByReferenceID retrieves an entity by its reference ID (case-sensitive)
func (r *BaseRepository[T]) GetByReferenceID(referenceID string) (*T, error) {
	var entity T
	if err := r.db.Where("reference_id = ?", referenceID).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, r.handleDBError(err)
	}
	return &entity, nil
}

// GetByReferenceIDCaseInsensitive retrieves an entity by its reference ID (case-insensitive)
func (r *BaseRepository[T]) GetByReferenceIDCaseInsensitive(referenceID string) (*T, error) {
	var entity T

	// Use ILIKE for PostgreSQL, LOWER() LIKE for SQLite compatibility
	var err error
	if r.db.Dialector.Name() == "postgres" {
		err = r.db.Where("reference_id ILIKE ?", referenceID).First(&entity).Error
	} else {
		// SQLite and other databases - use LOWER() LIKE for case-insensitive matching
		err = r.db.Where("LOWER(reference_id) LIKE LOWER(?)", referenceID).First(&entity).Error
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, r.handleDBError(err)
	}
	return &entity, nil
}

// Update updates an existing entity in the database
func (r *BaseRepository[T]) Update(entity *T) error {
	if err := r.db.Save(entity).Error; err != nil {
		return r.handleDBError(err)
	}
	return nil
}

// Delete deletes an entity by its UUID
func (r *BaseRepository[T]) Delete(id uuid.UUID) error {
	var entity T
	if err := r.db.Where("id = ?", id).Delete(&entity).Error; err != nil {
		return r.handleDBError(err)
	}
	return nil
}

// List retrieves entities with optional filtering, sorting, and pagination
func (r *BaseRepository[T]) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]T, error) {
	var entities []T
	query := r.db.Model(new(T))

	// Apply filters
	for field, value := range filters {
		query = query.Where(fmt.Sprintf("%s = ?", field), value)
	}

	// Apply ordering
	if orderBy != "" {
		query = query.Order(orderBy)
	}

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&entities).Error; err != nil {
		return nil, r.handleDBError(err)
	}

	return entities, nil
}

// Count returns the total number of entities matching the given filters
func (r *BaseRepository[T]) Count(filters map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.Model(new(T))

	// Apply filters
	for field, value := range filters {
		query = query.Where(fmt.Sprintf("%s = ?", field), value)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, r.handleDBError(err)
	}

	return count, nil
}

// Exists checks if an entity exists by its UUID
func (r *BaseRepository[T]) Exists(id uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(new(T)).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}
	return count > 0, nil
}

// ExistsByReferenceID checks if an entity exists by its reference ID
func (r *BaseRepository[T]) ExistsByReferenceID(referenceID string) (bool, error) {
	var count int64
	if err := r.db.Model(new(T)).Where("reference_id = ?", referenceID).Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}
	return count > 0, nil
}

// handleDBError converts GORM errors to repository-specific errors
func (r *BaseRepository[T]) handleDBError(err error) error {
	if err == nil {
		return nil
	}

	// Handle specific GORM/database errors
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return ErrDuplicateKey
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		return ErrForeignKey
	default:
		return err
	}
}

// WithTransaction executes a function within a database transaction
func (r *BaseRepository[T]) WithTransaction(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// GetDB returns the underlying database connection
func (r *BaseRepository[T]) GetDB() *gorm.DB {
	return r.db
}
