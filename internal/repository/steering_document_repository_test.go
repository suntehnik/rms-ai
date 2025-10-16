package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
)

func setupSteeringDocumentTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema including the junction table model
	err = db.AutoMigrate(&models.User{}, &models.Epic{}, &models.SteeringDocument{}, &models.EpicSteeringDocument{})
	require.NoError(t, err)

	// Ensure the UNIQUE constraint exists on the junction table for testing duplicate links
	err = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_epic_steering_unique ON epic_steering_documents(epic_id, steering_document_id)`).Error
	require.NoError(t, err)

	return db
}

func TestSteeringDocumentRepository_Create(t *testing.T) {
	db := setupSteeringDocumentTestDB(t)
	repo := NewSteeringDocumentRepository(db)

	// Create a test user first
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// Create a steering document
	title := "Test Steering Document"
	description := "This is a test steering document"
	doc := &models.SteeringDocument{
		Title:       title,
		Description: &description,
		CreatorID:   user.ID,
	}

	err = repo.Create(doc)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, doc.ID)
	assert.NotEmpty(t, doc.ReferenceID)
	assert.Contains(t, doc.ReferenceID, "STD-")
}

func TestSteeringDocumentRepository_GetByID(t *testing.T) {
	db := setupSteeringDocumentTestDB(t)
	repo := NewSteeringDocumentRepository(db)

	// Create a test user first
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// Create a steering document
	title := "Test Steering Document"
	description := "This is a test steering document"
	doc := &models.SteeringDocument{
		Title:       title,
		Description: &description,
		CreatorID:   user.ID,
	}

	err = repo.Create(doc)
	require.NoError(t, err)

	// Retrieve the document
	retrieved, err := repo.GetByID(doc.ID)
	assert.NoError(t, err)
	assert.Equal(t, doc.ID, retrieved.ID)
	assert.Equal(t, doc.Title, retrieved.Title)
	assert.Equal(t, doc.Description, retrieved.Description)
	assert.Equal(t, doc.CreatorID, retrieved.CreatorID)
	assert.Equal(t, user.Username, retrieved.Creator.Username) // Creator should be preloaded
}

func TestSteeringDocumentRepository_GetByReferenceID(t *testing.T) {
	db := setupSteeringDocumentTestDB(t)
	repo := NewSteeringDocumentRepository(db)

	// Create a test user first
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// Create a steering document
	title := "Test Steering Document"
	description := "This is a test steering document"
	doc := &models.SteeringDocument{
		Title:       title,
		Description: &description,
		CreatorID:   user.ID,
	}

	err = repo.Create(doc)
	require.NoError(t, err)

	// Retrieve the document by reference ID
	retrieved, err := repo.GetByReferenceID(doc.ReferenceID)
	assert.NoError(t, err)
	assert.Equal(t, doc.ID, retrieved.ID)
	assert.Equal(t, doc.ReferenceID, retrieved.ReferenceID)
	assert.Equal(t, user.Username, retrieved.Creator.Username) // Creator should be preloaded
}

func TestSteeringDocumentRepository_ListWithFilters(t *testing.T) {
	db := setupSteeringDocumentTestDB(t)
	repo := NewSteeringDocumentRepository(db)

	// Create test users
	user1 := &models.User{
		ID:       uuid.New(),
		Username: "user1",
		Email:    "user1@example.com",
		Role:     models.RoleUser,
	}
	user2 := &models.User{
		ID:       uuid.New(),
		Username: "user2",
		Email:    "user2@example.com",
		Role:     models.RoleUser,
	}
	err := db.Create([]*models.User{user1, user2}).Error
	require.NoError(t, err)

	// Create test steering documents
	docs := []*models.SteeringDocument{
		{
			Title:     "First Document",
			CreatorID: user1.ID,
		},
		{
			Title:     "Second Document",
			CreatorID: user2.ID,
		},
		{
			Title:     "Third Document",
			CreatorID: user1.ID,
		},
	}

	for _, doc := range docs {
		err = repo.Create(doc)
		require.NoError(t, err)
	}

	// Test listing all documents
	filters := SteeringDocumentFilters{
		Limit:  10,
		Offset: 0,
	}
	result, total, err := repo.ListWithFilters(filters)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, result, 3)

	// Test filtering by creator
	filters.CreatorID = &user1.ID
	result, total, err = repo.ListWithFilters(filters)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)

	// Test pagination
	filters.CreatorID = nil
	filters.Limit = 2
	filters.Offset = 1
	result, total, err = repo.ListWithFilters(filters)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, result, 2)
}

func TestSteeringDocumentRepository_LinkToEpic(t *testing.T) {
	db := setupSteeringDocumentTestDB(t)
	repo := NewSteeringDocumentRepository(db)

	// Create test user
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// Create test epic
	epic := &models.Epic{
		ID:        uuid.New(),
		Title:     "Test Epic",
		Priority:  models.PriorityHigh,
		CreatorID: user.ID,
		Status:    models.EpicStatusBacklog,
	}
	err = db.Create(epic).Error
	require.NoError(t, err)

	// Create test steering document
	doc := &models.SteeringDocument{
		Title:     "Test Document",
		CreatorID: user.ID,
	}
	err = repo.Create(doc)
	require.NoError(t, err)

	// Link document to epic
	err = repo.LinkToEpic(doc.ID, epic.ID)
	assert.NoError(t, err)

	// Verify the link exists
	docs, err := repo.GetByEpicID(epic.ID)
	assert.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, doc.ID, docs[0].ID)

	// Test duplicate link (should fail)
	err = repo.LinkToEpic(doc.ID, epic.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already linked")
}

func TestSteeringDocumentRepository_UnlinkFromEpic(t *testing.T) {
	db := setupSteeringDocumentTestDB(t)
	repo := NewSteeringDocumentRepository(db)

	// Create test user
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// Create test epic
	epic := &models.Epic{
		ID:        uuid.New(),
		Title:     "Test Epic",
		Priority:  models.PriorityHigh,
		CreatorID: user.ID,
		Status:    models.EpicStatusBacklog,
	}
	err = db.Create(epic).Error
	require.NoError(t, err)

	// Create test steering document
	doc := &models.SteeringDocument{
		Title:     "Test Document",
		CreatorID: user.ID,
	}
	err = repo.Create(doc)
	require.NoError(t, err)

	// Link document to epic
	err = repo.LinkToEpic(doc.ID, epic.ID)
	require.NoError(t, err)

	// Verify the link exists
	docs, err := repo.GetByEpicID(epic.ID)
	require.NoError(t, err)
	assert.Len(t, docs, 1)

	// Unlink document from epic
	err = repo.UnlinkFromEpic(doc.ID, epic.ID)
	assert.NoError(t, err)

	// Verify the link is removed
	docs, err = repo.GetByEpicID(epic.ID)
	assert.NoError(t, err)
	assert.Len(t, docs, 0)
}

func TestSteeringDocumentRepository_GetByID_NotFound(t *testing.T) {
	db := setupSteeringDocumentTestDB(t)
	repo := NewSteeringDocumentRepository(db)

	nonExistentID := uuid.New()
	doc, err := repo.GetByID(nonExistentID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, doc)
}

func TestSteeringDocumentRepository_GetByReferenceID_NotFound(t *testing.T) {
	db := setupSteeringDocumentTestDB(t)
	repo := NewSteeringDocumentRepository(db)

	doc, err := repo.GetByReferenceID("STD-999")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, doc)
}
