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

// setupRequirementTestDB creates an in-memory SQLite database for testing
func setupRequirementTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate all required models
	err = db.AutoMigrate(
		&models.User{},
		&models.Epic{},
		&models.UserStory{},
		&models.Requirement{},
		&models.RequirementType{},
		&models.AcceptanceCriteria{},
		&models.RequirementRelationship{},
		&models.RelationshipType{},
	)
	require.NoError(t, err)

	return db
}

// Helper functions for Requirement tests

func createTestRequirementType(t *testing.T, db *gorm.DB, name string) *models.RequirementType {
	desc := name + " requirements"
	reqType := &models.RequirementType{
		ID:          uuid.New(),
		Name:        name,
		Description: &desc,
	}
	err := db.Create(reqType).Error
	require.NoError(t, err)
	return reqType
}

func createTestRequirement(t *testing.T, db *gorm.DB, userStory *models.UserStory, creator *models.User, reqType *models.RequirementType, refID string) *models.Requirement {
	requirement := &models.Requirement{
		UserStoryID: userStory.ID,
		CreatorID:   creator.ID,
		AssigneeID:  creator.ID,
		Title:       "Test Requirement",
		Priority:    models.PriorityMedium,
		Status:      models.RequirementStatusDraft,
		TypeID:      reqType.ID,
		ReferenceID: refID, // Set manually to avoid PostgreSQL function call
	}
	err := db.Create(requirement).Error
	require.NoError(t, err)
	return requirement
}

// CRUD Tests

// TestRequirementRepository_Create tests the Create method
// References: AC-739 - Create returns requirement ID and commits; insert errors surface without committing
func TestRequirementRepository_Create(t *testing.T) {
	db := setupRequirementTestDB(t)
	repo := NewRequirementRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	reqType := createTestRequirementType(t, db, "Functional")

	requirement := &models.Requirement{
		UserStoryID: userStory.ID,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Title:       "Test Requirement",
		Priority:    models.PriorityMedium,
		TypeID:      reqType.ID,
		ReferenceID: "REQ-001", // Set manually to avoid PostgreSQL function call
	}

	err := repo.Create(requirement)
	require.NoError(t, err)
	assert.NotEmpty(t, requirement.ID)
	assert.Equal(t, "REQ-001", requirement.ReferenceID)
	assert.Equal(t, models.RequirementStatusDraft, requirement.Status) // Default status
}

// TestRequirementRepository_GetByID tests the GetByID method
// References: AC-740 - Get by ID returns full requirement; missing record yields not-found
func TestRequirementRepository_GetByID(t *testing.T) {
	db := setupRequirementTestDB(t)
	repo := NewRequirementRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	reqType := createTestRequirementType(t, db, "Functional")
	requirement := createTestRequirement(t, db, userStory, user, reqType, "REQ-001")

	// Test successful retrieval
	retrieved, err := repo.GetByID(requirement.ID)
	assert.NoError(t, err)
	assert.Equal(t, requirement.ID, retrieved.ID)
	assert.Equal(t, requirement.ReferenceID, retrieved.ReferenceID)
	assert.Equal(t, requirement.Title, retrieved.Title)
}

// TestRequirementRepository_GetByID_NotFound tests GetByID with non-existent ID
// References: AC-740 - Missing record yields not-found
func TestRequirementRepository_GetByID_NotFound(t *testing.T) {
	db := setupRequirementTestDB(t)
	repo := NewRequirementRepository(db)

	nonExistentID := uuid.New()
	req, err := repo.GetByID(nonExistentID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, req)
}

// TestRequirementRepository_GetByReferenceID tests the GetByReferenceID method
// References: AC-740 - Get by ReferenceID returns full requirement; missing record yields not-found
func TestRequirementRepository_GetByReferenceID(t *testing.T) {
	db := setupRequirementTestDB(t)
	repo := NewRequirementRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	reqType := createTestRequirementType(t, db, "Functional")
	requirement := createTestRequirement(t, db, userStory, user, reqType, "REQ-001")

	// Test successful retrieval
	retrieved, err := repo.GetByReferenceID(requirement.ReferenceID)
	assert.NoError(t, err)
	assert.Equal(t, requirement.ID, retrieved.ID)
	assert.Equal(t, requirement.ReferenceID, retrieved.ReferenceID)
	assert.Equal(t, requirement.Title, retrieved.Title)
}

// TestRequirementRepository_GetByReferenceID_NotFound tests GetByReferenceID with non-existent reference ID
// References: AC-740 - Missing record yields not-found
func TestRequirementRepository_GetByReferenceID_NotFound(t *testing.T) {
	db := setupRequirementTestDB(t)
	repo := NewRequirementRepository(db)

	req, err := repo.GetByReferenceID("REQ-999")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, req)
}

// TestRequirementRepository_Update tests the Update method
// References: AC-741 - Update performs a single UPDATE; conflicts/errors propagate without committing
func TestRequirementRepository_Update(t *testing.T) {
	db := setupRequirementTestDB(t)
	repo := NewRequirementRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	reqType := createTestRequirementType(t, db, "Functional")
	requirement := createTestRequirement(t, db, userStory, user, reqType, "REQ-001")

	// Update the requirement
	newTitle := "Updated Requirement"
	requirement.Title = newTitle
	requirement.Status = models.RequirementStatusActive

	err := repo.Update(requirement)
	require.NoError(t, err)

	// Verify the update
	retrieved, err := repo.GetByID(requirement.ID)
	require.NoError(t, err)
	assert.Equal(t, newTitle, retrieved.Title)
	assert.Equal(t, models.RequirementStatusActive, retrieved.Status)
}

// TestRequirementRepository_Delete tests the Delete method
// References: AC-742 - Delete removes requirement; FK/relationship conflicts surface and record remains
func TestRequirementRepository_Delete(t *testing.T) {
	db := setupRequirementTestDB(t)
	repo := NewRequirementRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	reqType := createTestRequirementType(t, db, "Functional")
	requirement := createTestRequirement(t, db, userStory, user, reqType, "REQ-001")

	// Delete the requirement
	err := repo.Delete(requirement.ID)
	require.NoError(t, err)

	// Verify deletion
	retrieved, err := repo.GetByID(requirement.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, retrieved)
}

// TestRequirementRepository_GetWithRelationships tests GetWithRelationships method
// References: AC-743 - CRUD tests are isolated on mocked or in-memory DB, idempotent and independent
func TestRequirementRepository_GetWithRelationships(t *testing.T) {
	db := setupRequirementTestDB(t)
	repo := NewRequirementRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	reqType := createTestRequirementType(t, db, "Functional")

	// Create source and target requirements
	sourceReq := createTestRequirement(t, db, userStory, user, reqType, "REQ-001")
	targetReq := createTestRequirement(t, db, userStory, user, reqType, "REQ-002")

	// Create relationship type
	relType := &models.RelationshipType{
		ID:   uuid.New(),
		Name: "depends_on",
	}
	err := db.Create(relType).Error
	require.NoError(t, err)

	// Create relationship
	relationship := &models.RequirementRelationship{
		ID:                     uuid.New(),
		SourceRequirementID:    sourceReq.ID,
		TargetRequirementID:    targetReq.ID,
		RelationshipTypeID:     relType.ID,
		CreatedBy:              user.ID,
	}
	err = db.Create(relationship).Error
	require.NoError(t, err)

	// Get requirement with relationships
	retrieved, err := repo.GetWithRelationships(sourceReq.ID)
	assert.NoError(t, err)
	assert.Equal(t, sourceReq.ID, retrieved.ID)
	assert.Len(t, retrieved.SourceRelationships, 1)
	assert.Equal(t, targetReq.ID, retrieved.SourceRelationships[0].TargetRequirementID)
}

// Filtering and Preload Tests

// TestRequirementRepository_GetByUserStory tests filtering by user story ID
// References: AC-744 - Filter by user_story_id returns only matching requirements; missing story yields empty result
func TestRequirementRepository_GetByUserStory(t *testing.T) {
	db := setupRequirementTestDB(t)
	repo := NewRequirementRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory1 := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	userStory2 := createUserStoryTestUserStory(t, db, epic, user, user, "US-002")
	reqType := createTestRequirementType(t, db, "Functional")

	// Create requirements for different user stories
	createTestRequirement(t, db, userStory1, user, reqType, "REQ-001")
	createTestRequirement(t, db, userStory1, user, reqType, "REQ-002")
	createTestRequirement(t, db, userStory2, user, reqType, "REQ-003")

	// Test filter by userStory1
	result, err := repo.GetByUserStory(userStory1.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Test filter by userStory2
	result, err = repo.GetByUserStory(userStory2.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	// Test filter by non-existent user story
	nonExistentID := uuid.New()
	result, err = repo.GetByUserStory(nonExistentID)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

// TestRequirementRepository_GetByType tests filtering by type ID
// References: AC-745 - Filter by type_id returns only that type; unknown type_id yields empty result
func TestRequirementRepository_GetByType(t *testing.T) {
	db := setupRequirementTestDB(t)
	repo := NewRequirementRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")

	reqTypeFunctional := createTestRequirementType(t, db, "Functional")
	reqTypeNonFunctional := createTestRequirementType(t, db, "Non-Functional")

	// Create requirements with different types
	createTestRequirement(t, db, userStory, user, reqTypeFunctional, "REQ-001")
	createTestRequirement(t, db, userStory, user, reqTypeFunctional, "REQ-002")
	createTestRequirement(t, db, userStory, user, reqTypeNonFunctional, "REQ-003")

	// Test filter by Functional type
	result, err := repo.GetByType(reqTypeFunctional.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Test filter by Non-Functional type
	result, err = repo.GetByType(reqTypeNonFunctional.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	// Test filter by non-existent type
	nonExistentID := uuid.New()
	result, err = repo.GetByType(nonExistentID)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

// TestRequirementRepository_GetByStatus tests filtering by status
// References: AC-746 - Filter by status returns only matching requirements; invalid status yields empty result
func TestRequirementRepository_GetByStatus(t *testing.T) {
	db := setupRequirementTestDB(t)
	repo := NewRequirementRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	reqType := createTestRequirementType(t, db, "Functional")

	// Create requirements with different statuses
	req1 := createTestRequirement(t, db, userStory, user, reqType, "REQ-001")
	req1.Status = models.RequirementStatusDraft
	db.Save(req1)

	req2 := createTestRequirement(t, db, userStory, user, reqType, "REQ-002")
	req2.Status = models.RequirementStatusActive
	db.Save(req2)

	req3 := createTestRequirement(t, db, userStory, user, reqType, "REQ-003")
	req3.Status = models.RequirementStatusDraft
	db.Save(req3)

	// Test filter by Draft status
	result, err := repo.GetByStatus(models.RequirementStatusDraft)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Test filter by Active status
	result, err = repo.GetByStatus(models.RequirementStatusActive)
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	// Test filter by Obsolete status (should be empty)
	result, err = repo.GetByStatus(models.RequirementStatusObsolete)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

// TestRequirementRepository_GetByIDWithPreloads tests preloading of related entities
// References: AC-747 - With preload enabled, acceptance_criteria and relationships load in minimal queries
func TestRequirementRepository_GetByIDWithPreloads(t *testing.T) {
	db := setupRequirementTestDB(t)
	repo := NewRequirementRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	reqType := createTestRequirementType(t, db, "Functional")

	// Create acceptance criteria
	ac := &models.AcceptanceCriteria{
		UserStoryID: userStory.ID,
		AuthorID:    user.ID,
		Description: "WHEN user logs in THEN the system SHALL show dashboard",
		ReferenceID: "AC-001",
	}
	err := db.Create(ac).Error
	require.NoError(t, err)

	// Create requirement with acceptance criteria link
	requirement := &models.Requirement{
		UserStoryID:          userStory.ID,
		CreatorID:            user.ID,
		AssigneeID:           user.ID,
		Title:                "Test Requirement",
		Priority:             models.PriorityMedium,
		TypeID:               reqType.ID,
		AcceptanceCriteriaID: &ac.ID,
		ReferenceID:          "REQ-001",
	}
	err = repo.Create(requirement)
	require.NoError(t, err)

	// Get requirement with preloads
	retrieved, err := repo.GetByIDWithPreloads(requirement.ID)
	assert.NoError(t, err)
	assert.Equal(t, requirement.ID, retrieved.ID)
	// Verify preloads
	assert.NotEmpty(t, retrieved.Creator.Username)
	assert.NotEmpty(t, retrieved.Assignee.Username)
	assert.NotEmpty(t, retrieved.UserStory.Title)
	assert.NotNil(t, retrieved.AcceptanceCriteria)
	assert.NotEmpty(t, retrieved.AcceptanceCriteria.Description)
	assert.NotEmpty(t, retrieved.Type.Name)
}

// TestRequirementRepository_ListWithPreloads tests listing with preloads
// References: AC-748 - DB interactions in filter/preload tests are mocked/verified; no external dependencies
func TestRequirementRepository_ListWithPreloads(t *testing.T) {
	db := setupRequirementTestDB(t)
	repo := NewRequirementRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	reqType := createTestRequirementType(t, db, "Functional")

	// Create requirements
	createTestRequirement(t, db, userStory, user, reqType, "REQ-001")
	createTestRequirement(t, db, userStory, user, reqType, "REQ-002")

	// Test list with preloads and filters
	filters := map[string]interface{}{
		"user_story_id": userStory.ID,
	}
	result, err := repo.ListWithPreloads(filters, "created_at ASC", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	// Verify preloads
	assert.NotEmpty(t, result[0].Creator.Username)
	assert.NotEmpty(t, result[0].Type.Name)
	assert.NotEmpty(t, result[0].UserStory.Title)
}

// TestRequirementRepository_GetByAcceptanceCriteria tests filtering by acceptance criteria ID
// References: AC-748 - DB interactions in filter/preload tests are mocked/verified; no external dependencies
func TestRequirementRepository_GetByAcceptanceCriteria(t *testing.T) {
	db := setupRequirementTestDB(t)
	repo := NewRequirementRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	reqType := createTestRequirementType(t, db, "Functional")

	// Create acceptance criteria
	ac1 := &models.AcceptanceCriteria{
		UserStoryID: userStory.ID,
		AuthorID:    user.ID,
		Description: "WHEN user logs in THEN system SHALL authenticate",
		ReferenceID: "AC-001",
	}
	err := db.Create(ac1).Error
	require.NoError(t, err)

	ac2 := &models.AcceptanceCriteria{
		UserStoryID: userStory.ID,
		AuthorID:    user.ID,
		Description: "WHEN credentials are invalid THEN system SHALL reject",
		ReferenceID: "AC-002",
	}
	err = db.Create(ac2).Error
	require.NoError(t, err)

	// Create requirements linked to acceptance criteria
	req1 := &models.Requirement{
		UserStoryID:          userStory.ID,
		CreatorID:            user.ID,
		AssigneeID:           user.ID,
		Title:                "Requirement 1",
		Priority:             models.PriorityMedium,
		TypeID:               reqType.ID,
		AcceptanceCriteriaID: &ac1.ID,
		ReferenceID:          "REQ-001",
	}
	err = repo.Create(req1)
	require.NoError(t, err)

	req2 := &models.Requirement{
		UserStoryID:          userStory.ID,
		CreatorID:            user.ID,
		AssigneeID:           user.ID,
		Title:                "Requirement 2",
		Priority:             models.PriorityMedium,
		TypeID:               reqType.ID,
		AcceptanceCriteriaID: &ac1.ID,
		ReferenceID:          "REQ-002",
	}
	err = repo.Create(req2)
	require.NoError(t, err)

	// Test filter by ac1
	result, err := repo.GetByAcceptanceCriteria(ac1.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Test filter by ac2 (should be empty)
	result, err = repo.GetByAcceptanceCriteria(ac2.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}
