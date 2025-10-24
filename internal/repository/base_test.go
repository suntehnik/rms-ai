package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestEntity is a simple entity for testing base repository functionality
type TestEntity struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	ReferenceID string    `gorm:"uniqueIndex;not null"`
	Name        string    `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (te *TestEntity) BeforeCreate(tx *gorm.DB) error {
	if te.ID == uuid.Nil {
		te.ID = uuid.New()
	}
	return nil
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate test entity
	err = db.AutoMigrate(&TestEntity{})
	require.NoError(t, err)

	return db
}

func TestBaseRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)

	entity := &TestEntity{
		ReferenceID: "TEST-001",
		Name:        "Test Entity",
	}

	err := repo.Create(entity)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, entity.ID)
}

func TestBaseRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)

	// Create test entity
	entity := &TestEntity{
		ReferenceID: "TEST-001",
		Name:        "Test Entity",
	}
	err := repo.Create(entity)
	require.NoError(t, err)

	// Get by ID
	retrieved, err := repo.GetByID(entity.ID)
	assert.NoError(t, err)
	assert.Equal(t, entity.ID, retrieved.ID)
	assert.Equal(t, entity.Name, retrieved.Name)
}

func TestBaseRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)

	nonExistentID := uuid.New()
	retrieved, err := repo.GetByID(nonExistentID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, retrieved)
}

func TestBaseRepository_GetByReferenceID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)

	// Create test entity
	entity := &TestEntity{
		ReferenceID: "TEST-001",
		Name:        "Test Entity",
	}
	err := repo.Create(entity)
	require.NoError(t, err)

	// Get by reference ID
	retrieved, err := repo.GetByReferenceID("TEST-001")
	assert.NoError(t, err)
	assert.Equal(t, entity.ID, retrieved.ID)
	assert.Equal(t, entity.Name, retrieved.Name)
}

func TestBaseRepository_GetByReferenceID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)

	retrieved, err := repo.GetByReferenceID("NON-EXISTENT")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, retrieved)
}

func TestBaseRepository_GetByReferenceIDCaseInsensitive(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)

	// Create test entity
	entity := &TestEntity{
		ReferenceID: "TEST-001",
		Name:        "Test Entity",
	}
	err := repo.Create(entity)
	require.NoError(t, err)

	// Test case-insensitive retrieval with lowercase
	retrieved, err := repo.GetByReferenceIDCaseInsensitive("test-001")
	assert.NoError(t, err)
	assert.Equal(t, entity.ID, retrieved.ID)
	assert.Equal(t, entity.Name, retrieved.Name)

	// Test case-insensitive retrieval with uppercase
	retrieved, err = repo.GetByReferenceIDCaseInsensitive("TEST-001")
	assert.NoError(t, err)
	assert.Equal(t, entity.ID, retrieved.ID)
	assert.Equal(t, entity.Name, retrieved.Name)

	// Test case-insensitive retrieval with mixed case
	retrieved, err = repo.GetByReferenceIDCaseInsensitive("Test-001")
	assert.NoError(t, err)
	assert.Equal(t, entity.ID, retrieved.ID)
	assert.Equal(t, entity.Name, retrieved.Name)
}

func TestBaseRepository_GetByReferenceIDCaseInsensitive_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)

	retrieved, err := repo.GetByReferenceIDCaseInsensitive("non-existent")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, retrieved)
}

func TestBaseRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)

	// Create test entity
	entity := &TestEntity{
		ReferenceID: "TEST-001",
		Name:        "Test Entity",
	}
	err := repo.Create(entity)
	require.NoError(t, err)

	// Update entity
	entity.Name = "Updated Entity"
	err = repo.Update(entity)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(entity.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Entity", retrieved.Name)
}

func TestBaseRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)

	// Create test entity
	entity := &TestEntity{
		ReferenceID: "TEST-001",
		Name:        "Test Entity",
	}
	err := repo.Create(entity)
	require.NoError(t, err)

	// Delete entity
	err = repo.Delete(entity.ID)
	assert.NoError(t, err)

	// Verify deletion
	retrieved, err := repo.GetByID(entity.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, retrieved)
}

func TestBaseRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)

	// Create test entities
	entities := []*TestEntity{
		{ReferenceID: "TEST-001", Name: "Entity 1"},
		{ReferenceID: "TEST-002", Name: "Entity 2"},
		{ReferenceID: "TEST-003", Name: "Entity 3"},
	}

	for _, entity := range entities {
		err := repo.Create(entity)
		require.NoError(t, err)
	}

	// List all entities
	retrieved, err := repo.List(nil, "", 0, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 3)

	// List with limit
	retrieved, err = repo.List(nil, "", 2, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)

	// List with offset
	retrieved, err = repo.List(nil, "", 0, 1)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)

	// List with filter
	filters := map[string]interface{}{"name": "Entity 1"}
	retrieved, err = repo.List(filters, "", 0, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, "Entity 1", retrieved[0].Name)
}

func TestBaseRepository_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)

	// Create test entities
	entities := []*TestEntity{
		{ReferenceID: "TEST-001", Name: "Entity 1"},
		{ReferenceID: "TEST-002", Name: "Entity 2"},
		{ReferenceID: "TEST-003", Name: "Entity 1"}, // Same name as first
	}

	for _, entity := range entities {
		err := repo.Create(entity)
		require.NoError(t, err)
	}

	// Count all entities
	count, err := repo.Count(nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// Count with filter
	filters := map[string]interface{}{"name": "Entity 1"}
	count, err = repo.Count(filters)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestBaseRepository_Exists(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)

	// Create test entity
	entity := &TestEntity{
		ReferenceID: "TEST-001",
		Name:        "Test Entity",
	}
	err := repo.Create(entity)
	require.NoError(t, err)

	// Check existence
	exists, err := repo.Exists(entity.ID)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Check non-existence
	nonExistentID := uuid.New()
	exists, err = repo.Exists(nonExistentID)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestBaseRepository_ExistsByReferenceID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)

	// Create test entity
	entity := &TestEntity{
		ReferenceID: "TEST-001",
		Name:        "Test Entity",
	}
	err := repo.Create(entity)
	require.NoError(t, err)

	// Check existence
	exists, err := repo.ExistsByReferenceID("TEST-001")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Check non-existence
	exists, err = repo.ExistsByReferenceID("NON-EXISTENT")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestBaseRepository_WithTransaction(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)

	// Test successful transaction
	err := repo.WithTransaction(func(tx *gorm.DB) error {
		entity := &TestEntity{
			ReferenceID: "TEST-001",
			Name:        "Test Entity",
		}
		return tx.Create(entity).Error
	})
	assert.NoError(t, err)

	// Verify entity was created
	count, err := repo.Count(nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}
