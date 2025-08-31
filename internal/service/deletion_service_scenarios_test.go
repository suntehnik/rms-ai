package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

// Test comprehensive deletion scenarios using existing mocks from other test files

// MockCommentRepository is a mock implementation of CommentRepository
type MockCommentRepository struct {
	mock.Mock
}

func (m *MockCommentRepository) Create(comment *models.Comment) error {
	args := m.Called(comment)
	return args.Error(0)
}

func (m *MockCommentRepository) GetByID(id uuid.UUID) (*models.Comment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Comment), args.Error(1)
}

func (m *MockCommentRepository) Update(comment *models.Comment) error {
	args := m.Called(comment)
	return args.Error(0)
}

func (m *MockCommentRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}



func (m *MockCommentRepository) GetByEntity(entityType models.EntityType, entityID uuid.UUID) ([]models.Comment, error) {
	args := m.Called(entityType, entityID)
	return args.Get(0).([]models.Comment), args.Error(1)
}

func (m *MockCommentRepository) GetByAuthor(authorID uuid.UUID) ([]models.Comment, error) {
	args := m.Called(authorID)
	return args.Get(0).([]models.Comment), args.Error(1)
}

func (m *MockCommentRepository) GetThreaded(entityType models.EntityType, entityID uuid.UUID) ([]models.Comment, error) {
	args := m.Called(entityType, entityID)
	return args.Get(0).([]models.Comment), args.Error(1)
}

func (m *MockCommentRepository) GetByReferenceID(referenceID string) (*models.Comment, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Comment), args.Error(1)
}

func (m *MockCommentRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.Comment, error) {
	args := m.Called(filters, orderBy, limit, offset)
	return args.Get(0).([]models.Comment), args.Error(1)
}

func (m *MockCommentRepository) Count(filters map[string]interface{}) (int64, error) {
	args := m.Called(filters)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCommentRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockCommentRepository) ExistsByReferenceID(referenceID string) (bool, error) {
	args := m.Called(referenceID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCommentRepository) WithTransaction(fn func(*gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockCommentRepository) GetDB() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *MockCommentRepository) GetByParent(parentID uuid.UUID) ([]models.Comment, error) {
	args := m.Called(parentID)
	return args.Get(0).([]models.Comment), args.Error(1)
}

func (m *MockCommentRepository) GetByStatus(isResolved bool) ([]models.Comment, error) {
	args := m.Called(isResolved)
	return args.Get(0).([]models.Comment), args.Error(1)
}

func (m *MockCommentRepository) GetInlineComments(entityType models.EntityType, entityID uuid.UUID) ([]models.Comment, error) {
	args := m.Called(entityType, entityID)
	return args.Get(0).([]models.Comment), args.Error(1)
}

// Test Epic Deletion with Dependencies - Validation Scenarios
func TestDeletionScenarios_EpicValidation_WithDependencies(t *testing.T) {
	// Create mocks
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockAcceptanceCriteriaRepo := &MockAcceptanceCriteriaRepository{}
	mockRequirementRepo := &MockRequirementRepository{}
	mockRequirementRelationshipRepo := &MockRequirementRelationshipRepository{}
	mockCommentRepo := &MockCommentRepository{}
	mockUserRepo := &MockUserRepository{}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewDeletionService(
		mockEpicRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockRequirementRepo,
		mockRequirementRelationshipRepo,
		mockCommentRepo,
		mockUserRepo,
		logger,
	)

	epicID := uuid.New()
	userStoryID := uuid.New()
	acceptanceCriteriaID := uuid.New()
	requirementID := uuid.New()

	epic := &models.Epic{
		ID:          epicID,
		ReferenceID: "EP-001",
		Title:       "Test Epic with Dependencies",
	}

	userStory := models.UserStory{
		ID:          userStoryID,
		ReferenceID: "US-001",
		Title:       "Test User Story",
		EpicID:      epicID,
	}

	acceptanceCriteria := models.AcceptanceCriteria{
		ID:          acceptanceCriteriaID,
		ReferenceID: "AC-001",
		Description: "Test Acceptance Criteria",
		UserStoryID: userStoryID,
	}

	requirement := models.Requirement{
		ID:          requirementID,
		ReferenceID: "REQ-001",
		Title:       "Test Requirement",
		UserStoryID: userStoryID,
	}

	// Setup mocks for validation
	mockEpicRepo.On("GetByID", epicID).Return(epic, nil)
	mockUserStoryRepo.On("GetByEpic", epicID).Return([]models.UserStory{userStory}, nil)
	mockAcceptanceCriteriaRepo.On("GetByUserStory", userStoryID).Return([]models.AcceptanceCriteria{acceptanceCriteria}, nil)
	mockRequirementRepo.On("GetByUserStory", userStoryID).Return([]models.Requirement{requirement}, nil)

	// Test validation
	depInfo, err := service.ValidateEpicDeletion(epicID)
	assert.NoError(t, err)
	assert.NotNil(t, depInfo)
	
	// Should not be able to delete without force due to dependencies
	assert.False(t, depInfo.CanDelete)
	assert.Len(t, depInfo.Dependencies, 1)
	assert.Equal(t, "user_story", depInfo.Dependencies[0].EntityType)
	assert.Equal(t, userStoryID, depInfo.Dependencies[0].EntityID)
	assert.Equal(t, "US-001", depInfo.Dependencies[0].ReferenceID)
	assert.Equal(t, "Epic contains user stories", depInfo.Dependencies[0].Reason)
	
	// Should show cascade delete count
	assert.Equal(t, 3, depInfo.CascadeDeleteCount) // user story + acceptance criteria + requirement
	assert.True(t, depInfo.RequiresConfirmation)
	
	// Verify cascade entities
	assert.Len(t, depInfo.CascadeDeleteEntities, 3)
	
	// Find user story in cascade entities
	var foundUserStory, foundAC, foundReq bool
	for _, entity := range depInfo.CascadeDeleteEntities {
		switch entity.EntityType {
		case "user_story":
			assert.Equal(t, userStoryID, entity.EntityID)
			assert.Equal(t, "US-001", entity.ReferenceID)
			foundUserStory = true
		case "acceptance_criteria":
			assert.Equal(t, acceptanceCriteriaID, entity.EntityID)
			assert.Equal(t, "AC-001", entity.ReferenceID)
			foundAC = true
		case "requirement":
			assert.Equal(t, requirementID, entity.EntityID)
			assert.Equal(t, "REQ-001", entity.ReferenceID)
			foundReq = true
		}
	}
	assert.True(t, foundUserStory, "User story should be in cascade delete entities")
	assert.True(t, foundAC, "Acceptance criteria should be in cascade delete entities")
	assert.True(t, foundReq, "Requirement should be in cascade delete entities")

	mockEpicRepo.AssertExpectations(t)
	mockUserStoryRepo.AssertExpectations(t)
	mockAcceptanceCriteriaRepo.AssertExpectations(t)
	mockRequirementRepo.AssertExpectations(t)
}

// Test Epic Deletion without Dependencies
func TestDeletionScenarios_EpicValidation_NoDependencies(t *testing.T) {
	// Create mocks
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockAcceptanceCriteriaRepo := &MockAcceptanceCriteriaRepository{}
	mockRequirementRepo := &MockRequirementRepository{}
	mockRequirementRelationshipRepo := &MockRequirementRelationshipRepository{}
	mockCommentRepo := &MockCommentRepository{}
	mockUserRepo := &MockUserRepository{}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewDeletionService(
		mockEpicRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockRequirementRepo,
		mockRequirementRelationshipRepo,
		mockCommentRepo,
		mockUserRepo,
		logger,
	)

	epicID := uuid.New()

	epic := &models.Epic{
		ID:          epicID,
		ReferenceID: "EP-002",
		Title:       "Test Epic without Dependencies",
	}

	// Setup mocks for validation
	mockEpicRepo.On("GetByID", epicID).Return(epic, nil)
	mockUserStoryRepo.On("GetByEpic", epicID).Return([]models.UserStory{}, nil)

	// Test validation
	depInfo, err := service.ValidateEpicDeletion(epicID)
	assert.NoError(t, err)
	assert.NotNil(t, depInfo)
	
	// Should be able to delete without force
	assert.True(t, depInfo.CanDelete)
	assert.Empty(t, depInfo.Dependencies)
	assert.Equal(t, 0, depInfo.CascadeDeleteCount)
	assert.False(t, depInfo.RequiresConfirmation)
	assert.Empty(t, depInfo.CascadeDeleteEntities)

	mockEpicRepo.AssertExpectations(t)
	mockUserStoryRepo.AssertExpectations(t)
}

// Test User Story Deletion with Dependencies
func TestDeletionScenarios_UserStoryValidation_WithDependencies(t *testing.T) {
	// Create mocks
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockAcceptanceCriteriaRepo := &MockAcceptanceCriteriaRepository{}
	mockRequirementRepo := &MockRequirementRepository{}
	mockRequirementRelationshipRepo := &MockRequirementRelationshipRepository{}
	mockCommentRepo := &MockCommentRepository{}
	mockUserRepo := &MockUserRepository{}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewDeletionService(
		mockEpicRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockRequirementRepo,
		mockRequirementRelationshipRepo,
		mockCommentRepo,
		mockUserRepo,
		logger,
	)

	userStoryID := uuid.New()
	acceptanceCriteriaID := uuid.New()
	requirementID := uuid.New()

	userStory := &models.UserStory{
		ID:          userStoryID,
		ReferenceID: "US-003",
		Title:       "Test User Story with Dependencies",
	}

	acceptanceCriteria := models.AcceptanceCriteria{
		ID:          acceptanceCriteriaID,
		ReferenceID: "AC-002",
		Description: "Test Acceptance Criteria",
		UserStoryID: userStoryID,
	}

	requirement := models.Requirement{
		ID:          requirementID,
		ReferenceID: "REQ-002",
		Title:       "Test Requirement",
		UserStoryID: userStoryID,
	}

	// Setup mocks for validation
	mockUserStoryRepo.On("GetByID", userStoryID).Return(userStory, nil)
	mockAcceptanceCriteriaRepo.On("GetByUserStory", userStoryID).Return([]models.AcceptanceCriteria{acceptanceCriteria}, nil)
	mockRequirementRepo.On("GetByUserStory", userStoryID).Return([]models.Requirement{requirement}, nil)

	// Test validation
	depInfo, err := service.ValidateUserStoryDeletion(userStoryID)
	assert.NoError(t, err)
	assert.NotNil(t, depInfo)
	
	// Should not be able to delete without force due to dependencies
	assert.False(t, depInfo.CanDelete)
	assert.Len(t, depInfo.Dependencies, 2) // acceptance criteria + requirement
	assert.Equal(t, 2, depInfo.CascadeDeleteCount)
	assert.True(t, depInfo.RequiresConfirmation)

	mockUserStoryRepo.AssertExpectations(t)
	mockAcceptanceCriteriaRepo.AssertExpectations(t)
	mockRequirementRepo.AssertExpectations(t)
}

// Test Acceptance Criteria Deletion - Last One Scenario
func TestDeletionScenarios_AcceptanceCriteriaValidation_LastOne(t *testing.T) {
	// Create mocks
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockAcceptanceCriteriaRepo := &MockAcceptanceCriteriaRepository{}
	mockRequirementRepo := &MockRequirementRepository{}
	mockRequirementRelationshipRepo := &MockRequirementRelationshipRepository{}
	mockCommentRepo := &MockCommentRepository{}
	mockUserRepo := &MockUserRepository{}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewDeletionService(
		mockEpicRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockRequirementRepo,
		mockRequirementRelationshipRepo,
		mockCommentRepo,
		mockUserRepo,
		logger,
	)

	acceptanceCriteriaID := uuid.New()
	userStoryID := uuid.New()

	acceptanceCriteria := &models.AcceptanceCriteria{
		ID:          acceptanceCriteriaID,
		ReferenceID: "AC-003",
		Description: "Last Acceptance Criteria",
		UserStoryID: userStoryID,
	}

	// Setup mocks for validation
	mockAcceptanceCriteriaRepo.On("GetByID", acceptanceCriteriaID).Return(acceptanceCriteria, nil)
	mockAcceptanceCriteriaRepo.On("CountByUserStory", userStoryID).Return(int64(1), nil) // Last one
	mockRequirementRepo.On("GetByAcceptanceCriteria", acceptanceCriteriaID).Return([]models.Requirement{}, nil)

	// Test validation
	depInfo, err := service.ValidateAcceptanceCriteriaDeletion(acceptanceCriteriaID)
	assert.NoError(t, err)
	assert.NotNil(t, depInfo)
	
	// Should not be able to delete the last acceptance criteria
	assert.False(t, depInfo.CanDelete)
	assert.Len(t, depInfo.Dependencies, 1)
	assert.Equal(t, "user_story", depInfo.Dependencies[0].EntityType)
	assert.Equal(t, userStoryID, depInfo.Dependencies[0].EntityID)
	assert.Equal(t, "User story must have at least one acceptance criteria", depInfo.Dependencies[0].Reason)
	assert.True(t, depInfo.RequiresConfirmation)

	mockAcceptanceCriteriaRepo.AssertExpectations(t)
	mockRequirementRepo.AssertExpectations(t)
}

// Test Acceptance Criteria Deletion - With Linked Requirements
func TestDeletionScenarios_AcceptanceCriteriaValidation_WithLinkedRequirements(t *testing.T) {
	// Create mocks
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockAcceptanceCriteriaRepo := &MockAcceptanceCriteriaRepository{}
	mockRequirementRepo := &MockRequirementRepository{}
	mockRequirementRelationshipRepo := &MockRequirementRelationshipRepository{}
	mockCommentRepo := &MockCommentRepository{}
	mockUserRepo := &MockUserRepository{}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewDeletionService(
		mockEpicRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockRequirementRepo,
		mockRequirementRelationshipRepo,
		mockCommentRepo,
		mockUserRepo,
		logger,
	)

	acceptanceCriteriaID := uuid.New()
	userStoryID := uuid.New()
	requirementID := uuid.New()

	acceptanceCriteria := &models.AcceptanceCriteria{
		ID:          acceptanceCriteriaID,
		ReferenceID: "AC-004",
		Description: "Acceptance Criteria with Linked Requirements",
		UserStoryID: userStoryID,
	}

	requirement := models.Requirement{
		ID:                   requirementID,
		ReferenceID:          "REQ-003",
		Title:                "Linked Requirement",
		UserStoryID:          userStoryID,
		AcceptanceCriteriaID: &acceptanceCriteriaID,
	}

	// Setup mocks for validation
	mockAcceptanceCriteriaRepo.On("GetByID", acceptanceCriteriaID).Return(acceptanceCriteria, nil)
	mockAcceptanceCriteriaRepo.On("CountByUserStory", userStoryID).Return(int64(2), nil) // Not the last one
	mockRequirementRepo.On("GetByAcceptanceCriteria", acceptanceCriteriaID).Return([]models.Requirement{requirement}, nil)

	// Test validation
	depInfo, err := service.ValidateAcceptanceCriteriaDeletion(acceptanceCriteriaID)
	assert.NoError(t, err)
	assert.NotNil(t, depInfo)
	
	// Should be able to delete (requirements will be unlinked)
	assert.True(t, depInfo.CanDelete)
	assert.Empty(t, depInfo.Dependencies)
	assert.Equal(t, 1, depInfo.CascadeDeleteCount) // requirement will be unlinked
	assert.True(t, depInfo.RequiresConfirmation)
	
	// Verify cascade entities show unlinking
	assert.Len(t, depInfo.CascadeDeleteEntities, 1)
	assert.Equal(t, "requirement_unlink", depInfo.CascadeDeleteEntities[0].EntityType)
	assert.Equal(t, requirementID, depInfo.CascadeDeleteEntities[0].EntityID)
	assert.Equal(t, "REQ-003", depInfo.CascadeDeleteEntities[0].ReferenceID)
	assert.Contains(t, depInfo.CascadeDeleteEntities[0].Title, "Unlink:")

	mockAcceptanceCriteriaRepo.AssertExpectations(t)
	mockRequirementRepo.AssertExpectations(t)
}

// Test Requirement Deletion with Relationships
func TestDeletionScenarios_RequirementValidation_WithRelationships(t *testing.T) {
	// Create mocks
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockAcceptanceCriteriaRepo := &MockAcceptanceCriteriaRepository{}
	mockRequirementRepo := &MockRequirementRepository{}
	mockRequirementRelationshipRepo := &MockRequirementRelationshipRepository{}
	mockCommentRepo := &MockCommentRepository{}
	mockUserRepo := &MockUserRepository{}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewDeletionService(
		mockEpicRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockRequirementRepo,
		mockRequirementRelationshipRepo,
		mockCommentRepo,
		mockUserRepo,
		logger,
	)

	requirementID := uuid.New()
	otherRequirementID := uuid.New()
	relationshipID := uuid.New()

	requirement := &models.Requirement{
		ID:          requirementID,
		ReferenceID: "REQ-004",
		Title:       "Requirement with Relationships",
	}

	otherRequirement := &models.Requirement{
		ID:          otherRequirementID,
		ReferenceID: "REQ-005",
		Title:       "Related Requirement",
	}

	relationship := models.RequirementRelationship{
		ID:                  relationshipID,
		SourceRequirementID: requirementID,
		TargetRequirementID: otherRequirementID,
	}

	// Setup mocks for validation
	mockRequirementRepo.On("GetByID", requirementID).Return(requirement, nil)
	mockRequirementRepo.On("GetByID", otherRequirementID).Return(otherRequirement, nil)
	mockRequirementRelationshipRepo.On("GetByRequirement", requirementID).Return([]models.RequirementRelationship{relationship}, nil)

	// Test validation
	depInfo, err := service.ValidateRequirementDeletion(requirementID)
	assert.NoError(t, err)
	assert.NotNil(t, depInfo)
	
	// Should not be able to delete without force due to relationships
	assert.False(t, depInfo.CanDelete)
	assert.Len(t, depInfo.Dependencies, 1)
	assert.Equal(t, "requirement_relationship", depInfo.Dependencies[0].EntityType)
	assert.Equal(t, relationshipID, depInfo.Dependencies[0].EntityID)
	assert.Equal(t, "Requirement has active relationships", depInfo.Dependencies[0].Reason)
	assert.Contains(t, depInfo.Dependencies[0].Title, "outgoing relationship with REQ-005")
	
	assert.Equal(t, 1, depInfo.CascadeDeleteCount)
	assert.True(t, depInfo.RequiresConfirmation)

	mockRequirementRepo.AssertExpectations(t)
	mockRequirementRelationshipRepo.AssertExpectations(t)
}

// Test Requirement Deletion without Relationships
func TestDeletionScenarios_RequirementValidation_NoRelationships(t *testing.T) {
	// Create mocks
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockAcceptanceCriteriaRepo := &MockAcceptanceCriteriaRepository{}
	mockRequirementRepo := &MockRequirementRepository{}
	mockRequirementRelationshipRepo := &MockRequirementRelationshipRepository{}
	mockCommentRepo := &MockCommentRepository{}
	mockUserRepo := &MockUserRepository{}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewDeletionService(
		mockEpicRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockRequirementRepo,
		mockRequirementRelationshipRepo,
		mockCommentRepo,
		mockUserRepo,
		logger,
	)

	requirementID := uuid.New()

	requirement := &models.Requirement{
		ID:          requirementID,
		ReferenceID: "REQ-006",
		Title:       "Standalone Requirement",
	}

	// Setup mocks for validation
	mockRequirementRepo.On("GetByID", requirementID).Return(requirement, nil)
	mockRequirementRelationshipRepo.On("GetByRequirement", requirementID).Return([]models.RequirementRelationship{}, nil)

	// Test validation
	depInfo, err := service.ValidateRequirementDeletion(requirementID)
	assert.NoError(t, err)
	assert.NotNil(t, depInfo)
	
	// Should be able to delete without force
	assert.True(t, depInfo.CanDelete)
	assert.Empty(t, depInfo.Dependencies)
	assert.Equal(t, 0, depInfo.CascadeDeleteCount)
	assert.False(t, depInfo.RequiresConfirmation)

	mockRequirementRepo.AssertExpectations(t)
	mockRequirementRelationshipRepo.AssertExpectations(t)
}

// Test Entity Not Found Scenarios
func TestDeletionScenarios_EntityNotFound(t *testing.T) {
	// Create mocks
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockAcceptanceCriteriaRepo := &MockAcceptanceCriteriaRepository{}
	mockRequirementRepo := &MockRequirementRepository{}
	mockRequirementRelationshipRepo := &MockRequirementRelationshipRepository{}
	mockCommentRepo := &MockCommentRepository{}
	mockUserRepo := &MockUserRepository{}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewDeletionService(
		mockEpicRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockRequirementRepo,
		mockRequirementRelationshipRepo,
		mockCommentRepo,
		mockUserRepo,
		logger,
	)

	nonExistentID := uuid.New()

	// Test Epic Not Found
	mockEpicRepo.On("GetByID", nonExistentID).Return(nil, repository.ErrNotFound)
	depInfo, err := service.ValidateEpicDeletion(nonExistentID)
	assert.Error(t, err)
	assert.Equal(t, ErrEpicNotFound, err)
	assert.Nil(t, depInfo)

	// Test User Story Not Found
	mockUserStoryRepo.On("GetByID", nonExistentID).Return(nil, repository.ErrNotFound)
	depInfo, err = service.ValidateUserStoryDeletion(nonExistentID)
	assert.Error(t, err)
	assert.Equal(t, ErrUserStoryNotFound, err)
	assert.Nil(t, depInfo)

	// Test Acceptance Criteria Not Found
	mockAcceptanceCriteriaRepo.On("GetByID", nonExistentID).Return(nil, repository.ErrNotFound)
	depInfo, err = service.ValidateAcceptanceCriteriaDeletion(nonExistentID)
	assert.Error(t, err)
	assert.Equal(t, ErrAcceptanceCriteriaNotFound, err)
	assert.Nil(t, depInfo)

	// Test Requirement Not Found
	mockRequirementRepo.On("GetByID", nonExistentID).Return(nil, repository.ErrNotFound)
	depInfo, err = service.ValidateRequirementDeletion(nonExistentID)
	assert.Error(t, err)
	assert.Equal(t, ErrRequirementNotFound, err)
	assert.Nil(t, depInfo)

	mockEpicRepo.AssertExpectations(t)
	mockUserStoryRepo.AssertExpectations(t)
	mockAcceptanceCriteriaRepo.AssertExpectations(t)
	mockRequirementRepo.AssertExpectations(t)
}

// Test Deletion Validation Failed Scenarios
func TestDeletionScenarios_ValidationFailed(t *testing.T) {
	// Create mocks
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockAcceptanceCriteriaRepo := &MockAcceptanceCriteriaRepository{}
	mockRequirementRepo := &MockRequirementRepository{}
	mockRequirementRelationshipRepo := &MockRequirementRelationshipRepository{}
	mockCommentRepo := &MockCommentRepository{}
	mockUserRepo := &MockUserRepository{}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewDeletionService(
		mockEpicRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockRequirementRepo,
		mockRequirementRelationshipRepo,
		mockCommentRepo,
		mockUserRepo,
		logger,
	)

	epicID := uuid.New()
	userStoryID := uuid.New()
	userID := uuid.New()

	epic := &models.Epic{
		ID:          epicID,
		ReferenceID: "EP-003",
		Title:       "Epic with Dependencies",
	}

	userStory := models.UserStory{
		ID:          userStoryID,
		ReferenceID: "US-004",
		Title:       "Dependent User Story",
		EpicID:      epicID,
	}

	// Setup mocks for validation that will fail
	mockEpicRepo.On("GetByID", epicID).Return(epic, nil)
	mockUserStoryRepo.On("GetByEpic", epicID).Return([]models.UserStory{userStory}, nil)
	mockAcceptanceCriteriaRepo.On("GetByUserStory", userStoryID).Return([]models.AcceptanceCriteria{}, nil)
	mockRequirementRepo.On("GetByUserStory", userStoryID).Return([]models.Requirement{}, nil)

	// Test deletion without force should fail validation
	result, err := service.DeleteEpicWithValidation(epicID, userID, false)
	assert.Error(t, err)
	assert.Equal(t, ErrDeletionValidationFailed, err)
	assert.Nil(t, result)

	mockEpicRepo.AssertExpectations(t)
	mockUserStoryRepo.AssertExpectations(t)
	mockAcceptanceCriteriaRepo.AssertExpectations(t)
	mockRequirementRepo.AssertExpectations(t)
}