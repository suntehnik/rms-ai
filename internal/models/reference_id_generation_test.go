package models

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// mockReferenceIDGenerator is a mock implementation for testing
type mockReferenceIDGenerator struct {
	generateFunc func(tx *gorm.DB, model interface{}) (string, error)
	callCount    int
}

func (m *mockReferenceIDGenerator) Generate(tx *gorm.DB, model interface{}) (string, error) {
	m.callCount++
	if m.generateFunc != nil {
		return m.generateFunc(tx, model)
	}
	return "MOCK-001", nil
}

// setupMockTestDB creates an in-memory SQLite database for testing with mocks
func setupMockTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate all models
	err = db.AutoMigrate(
		&User{},
		&Epic{},
		&UserStory{},
		&Requirement{},
		&AcceptanceCriteria{},
		&SteeringDocument{},
		&RequirementType{},
	)
	require.NoError(t, err)

	return db
}

// TestEpicBeforeCreate_CallsGenerator verifies that Epic.BeforeCreate calls the generator
func TestEpicBeforeCreate_CallsGenerator(t *testing.T) {
	db := setupMockTestDB(t)

	// Create a mock generator for Epic
	mockEpicGen := &mockReferenceIDGenerator{
		generateFunc: func(tx *gorm.DB, model interface{}) (string, error) {
			return "EP-MOCK", nil
		},
	}

	// Replace the epic generator with our mock
	originalGenerator := GetEpicGenerator()
	SetEpicGenerator(mockEpicGen)
	defer func() { SetEpicGenerator(originalGenerator) }()

	// Create a test user first (with manual ID to avoid generator calls)
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     RoleUser,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// Create an epic
	epic := &Epic{
		Title:      "Test Epic",
		Priority:   PriorityHigh,
		CreatorID:  user.ID,
		AssigneeID: user.ID,
	}

	err = db.Create(epic).Error
	require.NoError(t, err)

	// Verify the generator was called
	assert.Equal(t, 1, mockEpicGen.callCount, "Generator should be called once")
	assert.Equal(t, "EP-MOCK", epic.ReferenceID, "ReferenceID should be set by generator")
}

// TestEpicBeforeCreate_GeneratorError verifies error handling when generator fails
func TestEpicBeforeCreate_GeneratorError(t *testing.T) {
	db := setupMockTestDB(t)

	// Create a mock generator that returns an error
	mockEpicGen := &mockReferenceIDGenerator{
		generateFunc: func(tx *gorm.DB, model interface{}) (string, error) {
			return "", errors.New("generator error")
		},
	}

	// Replace the epic generator with our mock
	originalGenerator := GetEpicGenerator()
	SetEpicGenerator(mockEpicGen)
	defer func() { SetEpicGenerator(originalGenerator) }()

	// Create a test user first
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     RoleUser,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// Try to create an epic
	epic := &Epic{
		Title:      "Test Epic",
		Priority:   PriorityHigh,
		CreatorID:  user.ID,
		AssigneeID: user.ID,
	}

	err = db.Create(epic).Error
	assert.Error(t, err, "Should return error when generator fails")
	assert.Contains(t, err.Error(), "generator error")
}

// TestUserStoryBeforeCreate_CallsGenerator verifies that UserStory.BeforeCreate calls the generator
func TestUserStoryBeforeCreate_CallsGenerator(t *testing.T) {
	db := setupMockTestDB(t)

	// Create mock generators for Epic and UserStory
	mockEpicGen := &mockReferenceIDGenerator{
		generateFunc: func(tx *gorm.DB, model interface{}) (string, error) {
			return "EP-001", nil
		},
	}
	mockUSGen := &mockReferenceIDGenerator{
		generateFunc: func(tx *gorm.DB, model interface{}) (string, error) {
			return "US-MOCK", nil
		},
	}

	// Replace generators with our mocks
	originalEpicGen := GetEpicGenerator()
	originalUSGen := GetUserStoryGenerator()
	SetEpicGenerator(mockEpicGen)
	SetUserStoryGenerator(mockUSGen)
	defer func() {
		SetEpicGenerator(originalEpicGen)
		SetUserStoryGenerator(originalUSGen)
	}()

	// Create test data
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     RoleUser,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	epic := &Epic{
		Title:      "Test Epic",
		Priority:   PriorityHigh,
		CreatorID:  user.ID,
		AssigneeID: user.ID,
	}
	err = db.Create(epic).Error
	require.NoError(t, err)

	// Create a user story
	userStory := &UserStory{
		Title:      "Test User Story",
		Priority:   PriorityHigh,
		EpicID:     epic.ID,
		CreatorID:  user.ID,
		AssigneeID: user.ID,
	}

	err = db.Create(userStory).Error
	require.NoError(t, err)

	// Verify the generator was called
	assert.Equal(t, 1, mockUSGen.callCount, "UserStory generator should be called once")
	assert.Equal(t, "US-MOCK", userStory.ReferenceID, "ReferenceID should be set by generator")
}

// TestRequirementBeforeCreate_CallsGenerator verifies that Requirement.BeforeCreate calls the generator
func TestRequirementBeforeCreate_CallsGenerator(t *testing.T) {
	db := setupMockTestDB(t)

	// Create mock generators for all entities
	mockEpicGen := &mockReferenceIDGenerator{
		generateFunc: func(tx *gorm.DB, model interface{}) (string, error) {
			return "EP-001", nil
		},
	}
	mockUSGen := &mockReferenceIDGenerator{
		generateFunc: func(tx *gorm.DB, model interface{}) (string, error) {
			return "US-001", nil
		},
	}
	mockReqGen := &mockReferenceIDGenerator{
		generateFunc: func(tx *gorm.DB, model interface{}) (string, error) {
			return "REQ-MOCK", nil
		},
	}

	// Replace generators with our mocks
	originalEpicGen := GetEpicGenerator()
	originalUSGen := GetUserStoryGenerator()
	originalReqGen := GetRequirementGenerator()
	SetEpicGenerator(mockEpicGen)
	SetUserStoryGenerator(mockUSGen)
	SetRequirementGenerator(mockReqGen)
	defer func() {
		SetEpicGenerator(originalEpicGen)
		SetUserStoryGenerator(originalUSGen)
		SetRequirementGenerator(originalReqGen)
	}()

	// Create test data
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     RoleUser,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	epic := &Epic{
		Title:      "Test Epic",
		Priority:   PriorityHigh,
		CreatorID:  user.ID,
		AssigneeID: user.ID,
	}
	err = db.Create(epic).Error
	require.NoError(t, err)

	userStory := &UserStory{
		Title:      "Test User Story",
		Priority:   PriorityHigh,
		EpicID:     epic.ID,
		CreatorID:  user.ID,
		AssigneeID: user.ID,
	}
	err = db.Create(userStory).Error
	require.NoError(t, err)

	reqType := &RequirementType{
		Name: "Functional",
	}
	err = db.Create(reqType).Error
	require.NoError(t, err)

	// Create a requirement
	requirement := &Requirement{
		Title:       "Test Requirement",
		Priority:    PriorityHigh,
		UserStoryID: userStory.ID,
		TypeID:      reqType.ID,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
	}

	err = db.Create(requirement).Error
	require.NoError(t, err)

	// Verify the generator was called
	assert.Equal(t, 1, mockReqGen.callCount, "Requirement generator should be called once")
	assert.Equal(t, "REQ-MOCK", requirement.ReferenceID, "ReferenceID should be set by generator")
}

// TestAcceptanceCriteriaBeforeCreate_CallsGenerator verifies that AcceptanceCriteria.BeforeCreate calls the generator
func TestAcceptanceCriteriaBeforeCreate_CallsGenerator(t *testing.T) {
	db := setupMockTestDB(t)

	// Create mock generators for all entities
	mockEpicGen := &mockReferenceIDGenerator{
		generateFunc: func(tx *gorm.DB, model interface{}) (string, error) {
			return "EP-001", nil
		},
	}
	mockUSGen := &mockReferenceIDGenerator{
		generateFunc: func(tx *gorm.DB, model interface{}) (string, error) {
			return "US-001", nil
		},
	}
	mockACGen := &mockReferenceIDGenerator{
		generateFunc: func(tx *gorm.DB, model interface{}) (string, error) {
			return "AC-MOCK", nil
		},
	}

	// Replace generators with our mocks
	originalEpicGen := GetEpicGenerator()
	originalUSGen := GetUserStoryGenerator()
	originalACGen := GetAcceptanceCriteriaGenerator()
	SetEpicGenerator(mockEpicGen)
	SetUserStoryGenerator(mockUSGen)
	SetAcceptanceCriteriaGenerator(mockACGen)
	defer func() {
		SetEpicGenerator(originalEpicGen)
		SetUserStoryGenerator(originalUSGen)
		SetAcceptanceCriteriaGenerator(originalACGen)
	}()

	// Create test data
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     RoleUser,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	epic := &Epic{
		Title:      "Test Epic",
		Priority:   PriorityHigh,
		CreatorID:  user.ID,
		AssigneeID: user.ID,
	}
	err = db.Create(epic).Error
	require.NoError(t, err)

	userStory := &UserStory{
		Title:      "Test User Story",
		Priority:   PriorityHigh,
		EpicID:     epic.ID,
		CreatorID:  user.ID,
		AssigneeID: user.ID,
	}
	err = db.Create(userStory).Error
	require.NoError(t, err)

	// Create acceptance criteria
	ac := &AcceptanceCriteria{
		Description: "Test acceptance criteria",
		UserStoryID: userStory.ID,
		AuthorID:    user.ID,
	}

	err = db.Create(ac).Error
	require.NoError(t, err)

	// Verify the generator was called
	assert.Equal(t, 1, mockACGen.callCount, "AcceptanceCriteria generator should be called once")
	assert.Equal(t, "AC-MOCK", ac.ReferenceID, "ReferenceID should be set by generator")
}

// TestSteeringDocumentBeforeCreate_CallsGenerator verifies that SteeringDocument.BeforeCreate calls the generator
func TestSteeringDocumentBeforeCreate_CallsGenerator(t *testing.T) {
	db := setupMockTestDB(t)

	// Create a mock generator
	mockGen := &mockReferenceIDGenerator{
		generateFunc: func(tx *gorm.DB, model interface{}) (string, error) {
			return "STD-MOCK", nil
		},
	}

	// Replace the steering document generator with our mock
	originalGenerator := GetSteeringDocumentGenerator()
	SetSteeringDocumentGenerator(mockGen)
	defer func() { SetSteeringDocumentGenerator(originalGenerator) }()

	// Create test data
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     RoleUser,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// Create steering document
	steeringDoc := &SteeringDocument{
		Title:     "Test Steering Document",
		CreatorID: user.ID,
	}

	err = db.Create(steeringDoc).Error
	require.NoError(t, err)

	// Verify the generator was called
	assert.Equal(t, 1, mockGen.callCount, "Generator should be called once")
	assert.Equal(t, "STD-MOCK", steeringDoc.ReferenceID, "ReferenceID should be set by generator")
}
