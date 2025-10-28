package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

// MockRequirementRepository is a mock implementation of RequirementRepository
type MockRequirementRepository struct {
	mock.Mock
}

func (m *MockRequirementRepository) Create(entity *models.Requirement) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockRequirementRepository) GetByID(id uuid.UUID) (*models.Requirement, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementRepository) GetByReferenceID(referenceID string) (*models.Requirement, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementRepository) GetByReferenceIDCaseInsensitive(referenceID string) (*models.Requirement, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementRepository) Update(entity *models.Requirement) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockRequirementRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRequirementRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.Requirement, error) {
	args := m.Called(filters, orderBy, limit, offset)
	return args.Get(0).([]models.Requirement), args.Error(1)
}

func (m *MockRequirementRepository) Count(filters map[string]interface{}) (int64, error) {
	args := m.Called(filters)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRequirementRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockRequirementRepository) ExistsByReferenceID(referenceID string) (bool, error) {
	args := m.Called(referenceID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRequirementRepository) WithTransaction(fn func(*gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockRequirementRepository) GetDB() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockRequirementRepository) GetWithRelationships(id uuid.UUID) (*models.Requirement, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementRepository) GetByUserStory(userStoryID uuid.UUID) ([]models.Requirement, error) {
	args := m.Called(userStoryID)
	return args.Get(0).([]models.Requirement), args.Error(1)
}

func (m *MockRequirementRepository) GetByAcceptanceCriteria(acceptanceCriteriaID uuid.UUID) ([]models.Requirement, error) {
	args := m.Called(acceptanceCriteriaID)
	return args.Get(0).([]models.Requirement), args.Error(1)
}

func (m *MockRequirementRepository) GetByCreator(creatorID uuid.UUID) ([]models.Requirement, error) {
	args := m.Called(creatorID)
	return args.Get(0).([]models.Requirement), args.Error(1)
}

func (m *MockRequirementRepository) GetByAssignee(assigneeID uuid.UUID) ([]models.Requirement, error) {
	args := m.Called(assigneeID)
	return args.Get(0).([]models.Requirement), args.Error(1)
}

func (m *MockRequirementRepository) GetByStatus(status models.RequirementStatus) ([]models.Requirement, error) {
	args := m.Called(status)
	return args.Get(0).([]models.Requirement), args.Error(1)
}

func (m *MockRequirementRepository) GetByPriority(priority models.Priority) ([]models.Requirement, error) {
	args := m.Called(priority)
	return args.Get(0).([]models.Requirement), args.Error(1)
}

func (m *MockRequirementRepository) GetByType(typeID uuid.UUID) ([]models.Requirement, error) {
	args := m.Called(typeID)
	return args.Get(0).([]models.Requirement), args.Error(1)
}

func (m *MockRequirementRepository) HasRelationships(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockRequirementRepository) SearchByText(searchText string) ([]models.Requirement, error) {
	args := m.Called(searchText)
	return args.Get(0).([]models.Requirement), args.Error(1)
}

func (m *MockRequirementRepository) SearchByTextWithPagination(searchText string, limit, offset int) ([]models.Requirement, int64, error) {
	args := m.Called(searchText, limit, offset)
	return args.Get(0).([]models.Requirement), args.Get(1).(int64), args.Error(2)
}

// MockRequirementTypeRepository is a mock implementation of RequirementTypeRepository
type MockRequirementTypeRepository struct {
	mock.Mock
}

func (m *MockRequirementTypeRepository) Create(entity *models.RequirementType) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockRequirementTypeRepository) GetByID(id uuid.UUID) (*models.RequirementType, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockRequirementTypeRepository) GetByReferenceID(referenceID string) (*models.RequirementType, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockRequirementTypeRepository) GetByReferenceIDCaseInsensitive(referenceID string) (*models.RequirementType, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockRequirementTypeRepository) Update(entity *models.RequirementType) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockRequirementTypeRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRequirementTypeRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.RequirementType, error) {
	args := m.Called(filters, orderBy, limit, offset)
	return args.Get(0).([]models.RequirementType), args.Error(1)
}

func (m *MockRequirementTypeRepository) Count(filters map[string]interface{}) (int64, error) {
	args := m.Called(filters)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRequirementTypeRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockRequirementTypeRepository) ExistsByReferenceID(referenceID string) (bool, error) {
	args := m.Called(referenceID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRequirementTypeRepository) WithTransaction(fn func(*gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockRequirementTypeRepository) GetDB() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockRequirementTypeRepository) GetByName(name string) (*models.RequirementType, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockRequirementTypeRepository) ExistsByName(name string) (bool, error) {
	args := m.Called(name)
	return args.Bool(0), args.Error(1)
}

// MockRelationshipTypeRepository is a mock implementation of RelationshipTypeRepository
type MockRelationshipTypeRepository struct {
	mock.Mock
}

func (m *MockRelationshipTypeRepository) Create(entity *models.RelationshipType) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockRelationshipTypeRepository) GetByID(id uuid.UUID) (*models.RelationshipType, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RelationshipType), args.Error(1)
}

func (m *MockRelationshipTypeRepository) GetByReferenceID(referenceID string) (*models.RelationshipType, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RelationshipType), args.Error(1)
}

func (m *MockRelationshipTypeRepository) GetByReferenceIDCaseInsensitive(referenceID string) (*models.RelationshipType, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RelationshipType), args.Error(1)
}

func (m *MockRelationshipTypeRepository) Update(entity *models.RelationshipType) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockRelationshipTypeRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRelationshipTypeRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.RelationshipType, error) {
	args := m.Called(filters, orderBy, limit, offset)
	return args.Get(0).([]models.RelationshipType), args.Error(1)
}

func (m *MockRelationshipTypeRepository) Count(filters map[string]interface{}) (int64, error) {
	args := m.Called(filters)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRelationshipTypeRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockRelationshipTypeRepository) ExistsByReferenceID(referenceID string) (bool, error) {
	args := m.Called(referenceID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRelationshipTypeRepository) WithTransaction(fn func(*gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockRelationshipTypeRepository) GetDB() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockRelationshipTypeRepository) GetByName(name string) (*models.RelationshipType, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RelationshipType), args.Error(1)
}

func (m *MockRelationshipTypeRepository) ExistsByName(name string) (bool, error) {
	args := m.Called(name)
	return args.Bool(0), args.Error(1)
}

// MockRequirementRelationshipRepository is a mock implementation of RequirementRelationshipRepository
type MockRequirementRelationshipRepository struct {
	mock.Mock
}

func (m *MockRequirementRelationshipRepository) Create(entity *models.RequirementRelationship) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockRequirementRelationshipRepository) GetByID(id uuid.UUID) (*models.RequirementRelationship, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RequirementRelationship), args.Error(1)
}

func (m *MockRequirementRelationshipRepository) GetByReferenceID(referenceID string) (*models.RequirementRelationship, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RequirementRelationship), args.Error(1)
}

func (m *MockRequirementRelationshipRepository) GetByReferenceIDCaseInsensitive(referenceID string) (*models.RequirementRelationship, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RequirementRelationship), args.Error(1)
}

func (m *MockRequirementRelationshipRepository) Update(entity *models.RequirementRelationship) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockRequirementRelationshipRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRequirementRelationshipRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.RequirementRelationship, error) {
	args := m.Called(filters, orderBy, limit, offset)
	return args.Get(0).([]models.RequirementRelationship), args.Error(1)
}

func (m *MockRequirementRelationshipRepository) Count(filters map[string]interface{}) (int64, error) {
	args := m.Called(filters)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRequirementRelationshipRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockRequirementRelationshipRepository) ExistsByReferenceID(referenceID string) (bool, error) {
	args := m.Called(referenceID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRequirementRelationshipRepository) WithTransaction(fn func(*gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockRequirementRelationshipRepository) GetDB() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockRequirementRelationshipRepository) GetBySourceRequirement(sourceID uuid.UUID) ([]models.RequirementRelationship, error) {
	args := m.Called(sourceID)
	return args.Get(0).([]models.RequirementRelationship), args.Error(1)
}

func (m *MockRequirementRelationshipRepository) GetByTargetRequirement(targetID uuid.UUID) ([]models.RequirementRelationship, error) {
	args := m.Called(targetID)
	return args.Get(0).([]models.RequirementRelationship), args.Error(1)
}

func (m *MockRequirementRelationshipRepository) GetByRequirement(requirementID uuid.UUID) ([]models.RequirementRelationship, error) {
	args := m.Called(requirementID)
	return args.Get(0).([]models.RequirementRelationship), args.Error(1)
}

func (m *MockRequirementRelationshipRepository) GetByType(typeID uuid.UUID) ([]models.RequirementRelationship, error) {
	args := m.Called(typeID)
	return args.Get(0).([]models.RequirementRelationship), args.Error(1)
}

func (m *MockRequirementRelationshipRepository) ExistsRelationship(sourceID, targetID, typeID uuid.UUID) (bool, error) {
	args := m.Called(sourceID, targetID, typeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRequirementRelationshipRepository) GetByRequirementWithPagination(requirementID uuid.UUID, limit, offset int) ([]models.RequirementRelationship, int64, error) {
	args := m.Called(requirementID, limit, offset)
	return args.Get(0).([]models.RequirementRelationship), args.Get(1).(int64), args.Error(2)
}

func TestRequirementService_CreateRequirement(t *testing.T) {
	mockRequirementRepo := new(MockRequirementRepository)
	mockRequirementTypeRepo := new(MockRequirementTypeRepository)
	mockRelationshipTypeRepo := new(MockRelationshipTypeRepository)
	mockRequirementRelationshipRepo := new(MockRequirementRelationshipRepository)
	mockUserStoryRepo := new(MockUserStoryRepository)
	mockAcceptanceCriteriaRepo := new(MockAcceptanceCriteriaRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewRequirementService(
		mockRequirementRepo,
		mockRequirementTypeRepo,
		mockRelationshipTypeRepo,
		mockRequirementRelationshipRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockUserRepo,
	)

	t.Run("successful creation", func(t *testing.T) {
		userStoryID := uuid.New()
		creatorID := uuid.New()
		typeID := uuid.New()

		req := CreateRequirementRequest{
			UserStoryID: userStoryID,
			CreatorID:   creatorID,
			Priority:    models.PriorityHigh,
			TypeID:      typeID,
			Title:       "Test Requirement",
			Description: stringPtr("Test requirement description"),
		}

		// Mock expectations
		mockUserStoryRepo.On("Exists", userStoryID).Return(true, nil)
		mockRequirementTypeRepo.On("Exists", typeID).Return(true, nil)
		mockUserRepo.On("Exists", creatorID).Return(true, nil)
		mockRequirementRepo.On("Create", mock.AnythingOfType("*models.Requirement")).Return(nil)

		result, err := service.CreateRequirement(req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userStoryID, result.UserStoryID)
		assert.Equal(t, creatorID, result.CreatorID)
		assert.Equal(t, creatorID, result.AssigneeID) // Should default to creator
		assert.Equal(t, models.PriorityHigh, result.Priority)
		assert.Equal(t, models.RequirementStatusDraft, result.Status) // Should default to Draft
		assert.Equal(t, typeID, result.TypeID)
		assert.Equal(t, "Test Requirement", result.Title)

		mockUserStoryRepo.AssertExpectations(t)
		mockRequirementTypeRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockRequirementRepo.AssertExpectations(t)
	})

	t.Run("invalid priority", func(t *testing.T) {
		req := CreateRequirementRequest{
			UserStoryID: uuid.New(),
			CreatorID:   uuid.New(),
			Priority:    models.Priority(5), // Invalid priority
			TypeID:      uuid.New(),
			Title:       "Test Requirement",
		}

		result, err := service.CreateRequirement(req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrInvalidPriority, err)
	})

	t.Run("user story not found", func(t *testing.T) {
		userStoryID := uuid.New()

		req := CreateRequirementRequest{
			UserStoryID: userStoryID,
			CreatorID:   uuid.New(),
			Priority:    models.PriorityHigh,
			TypeID:      uuid.New(),
			Title:       "Test Requirement",
		}

		mockUserStoryRepo.On("Exists", userStoryID).Return(false, nil)

		result, err := service.CreateRequirement(req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrUserStoryNotFound, err)

		mockUserStoryRepo.AssertExpectations(t)
	})

	t.Run("requirement type not found", func(t *testing.T) {
		userStoryID := uuid.New()
		typeID := uuid.New()

		req := CreateRequirementRequest{
			UserStoryID: userStoryID,
			CreatorID:   uuid.New(),
			Priority:    models.PriorityHigh,
			TypeID:      typeID,
			Title:       "Test Requirement",
		}

		mockUserStoryRepo.On("Exists", userStoryID).Return(true, nil)
		mockRequirementTypeRepo.On("Exists", typeID).Return(false, nil)

		result, err := service.CreateRequirement(req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrRequirementTypeNotFound, err)

		mockUserStoryRepo.AssertExpectations(t)
		mockRequirementTypeRepo.AssertExpectations(t)
	})
}

func TestRequirementService_GetRequirementByID(t *testing.T) {
	mockRequirementRepo := new(MockRequirementRepository)
	mockRequirementTypeRepo := new(MockRequirementTypeRepository)
	mockRelationshipTypeRepo := new(MockRelationshipTypeRepository)
	mockRequirementRelationshipRepo := new(MockRequirementRelationshipRepository)
	mockUserStoryRepo := new(MockUserStoryRepository)
	mockAcceptanceCriteriaRepo := new(MockAcceptanceCriteriaRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewRequirementService(
		mockRequirementRepo,
		mockRequirementTypeRepo,
		mockRelationshipTypeRepo,
		mockRequirementRelationshipRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockUserRepo,
	)

	t.Run("successful retrieval", func(t *testing.T) {
		requirementID := uuid.New()
		expectedRequirement := &models.Requirement{
			ID:    requirementID,
			Title: "Test Requirement",
		}

		mockRequirementRepo.On("GetByID", requirementID).Return(expectedRequirement, nil)

		result, err := service.GetRequirementByID(requirementID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedRequirement.ID, result.ID)
		assert.Equal(t, expectedRequirement.Title, result.Title)

		mockRequirementRepo.AssertExpectations(t)
	})

	t.Run("requirement not found", func(t *testing.T) {
		requirementID := uuid.New()

		mockRequirementRepo.On("GetByID", requirementID).Return(nil, repository.ErrNotFound)

		result, err := service.GetRequirementByID(requirementID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrRequirementNotFound, err)

		mockRequirementRepo.AssertExpectations(t)
	})
}

func TestRequirementService_CreateRelationship(t *testing.T) {
	mockRequirementRepo := new(MockRequirementRepository)
	mockRequirementTypeRepo := new(MockRequirementTypeRepository)
	mockRelationshipTypeRepo := new(MockRelationshipTypeRepository)
	mockRequirementRelationshipRepo := new(MockRequirementRelationshipRepository)
	mockUserStoryRepo := new(MockUserStoryRepository)
	mockAcceptanceCriteriaRepo := new(MockAcceptanceCriteriaRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewRequirementService(
		mockRequirementRepo,
		mockRequirementTypeRepo,
		mockRelationshipTypeRepo,
		mockRequirementRelationshipRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockUserRepo,
	)

	t.Run("successful creation", func(t *testing.T) {
		sourceID := uuid.New()
		targetID := uuid.New()
		relationshipTypeID := uuid.New()
		creatorID := uuid.New()

		req := CreateRelationshipRequest{
			SourceRequirementID: sourceID,
			TargetRequirementID: targetID,
			RelationshipTypeID:  relationshipTypeID,
			CreatedBy:           creatorID,
		}

		// Mock expectations
		mockRequirementRepo.On("Exists", sourceID).Return(true, nil)
		mockRequirementRepo.On("Exists", targetID).Return(true, nil)
		mockRelationshipTypeRepo.On("Exists", relationshipTypeID).Return(true, nil)
		mockUserRepo.On("Exists", creatorID).Return(true, nil)
		mockRequirementRelationshipRepo.On("ExistsRelationship", sourceID, targetID, relationshipTypeID).Return(false, nil)
		mockRequirementRelationshipRepo.On("Create", mock.AnythingOfType("*models.RequirementRelationship")).Return(nil)

		result, err := service.CreateRelationship(req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, sourceID, result.SourceRequirementID)
		assert.Equal(t, targetID, result.TargetRequirementID)
		assert.Equal(t, relationshipTypeID, result.RelationshipTypeID)
		assert.Equal(t, creatorID, result.CreatedBy)

		mockRequirementRepo.AssertExpectations(t)
		mockRelationshipTypeRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockRequirementRelationshipRepo.AssertExpectations(t)
	})

	t.Run("circular relationship", func(t *testing.T) {
		requirementID := uuid.New()

		req := CreateRelationshipRequest{
			SourceRequirementID: requirementID,
			TargetRequirementID: requirementID, // Same as source
			RelationshipTypeID:  uuid.New(),
			CreatedBy:           uuid.New(),
		}

		result, err := service.CreateRelationship(req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrCircularRelationship, err)
	})

	t.Run("duplicate relationship", func(t *testing.T) {
		sourceID := uuid.New()
		targetID := uuid.New()
		relationshipTypeID := uuid.New()
		creatorID := uuid.New()

		req := CreateRelationshipRequest{
			SourceRequirementID: sourceID,
			TargetRequirementID: targetID,
			RelationshipTypeID:  relationshipTypeID,
			CreatedBy:           creatorID,
		}

		// Mock expectations
		mockRequirementRepo.On("Exists", sourceID).Return(true, nil)
		mockRequirementRepo.On("Exists", targetID).Return(true, nil)
		mockRelationshipTypeRepo.On("Exists", relationshipTypeID).Return(true, nil)
		mockUserRepo.On("Exists", creatorID).Return(true, nil)
		mockRequirementRelationshipRepo.On("ExistsRelationship", sourceID, targetID, relationshipTypeID).Return(true, nil)

		result, err := service.CreateRelationship(req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrDuplicateRelationship, err)

		mockRequirementRepo.AssertExpectations(t)
		mockRelationshipTypeRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockRequirementRelationshipRepo.AssertExpectations(t)
	})
}

func TestRequirementService_DeleteRequirement(t *testing.T) {
	mockRequirementRepo := new(MockRequirementRepository)
	mockRequirementTypeRepo := new(MockRequirementTypeRepository)
	mockRelationshipTypeRepo := new(MockRelationshipTypeRepository)
	mockRequirementRelationshipRepo := new(MockRequirementRelationshipRepository)
	mockUserStoryRepo := new(MockUserStoryRepository)
	mockAcceptanceCriteriaRepo := new(MockAcceptanceCriteriaRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewRequirementService(
		mockRequirementRepo,
		mockRequirementTypeRepo,
		mockRelationshipTypeRepo,
		mockRequirementRelationshipRepo,
		mockUserStoryRepo,
		mockAcceptanceCriteriaRepo,
		mockUserRepo,
	)

	t.Run("successful deletion without relationships", func(t *testing.T) {
		requirementID := uuid.New()
		requirement := &models.Requirement{ID: requirementID}

		mockRequirementRepo.On("GetByID", requirementID).Return(requirement, nil)
		mockRequirementRepo.On("HasRelationships", requirementID).Return(false, nil)
		mockRequirementRepo.On("Delete", requirementID).Return(nil)

		err := service.DeleteRequirement(requirementID, false)

		assert.NoError(t, err)

		mockRequirementRepo.AssertExpectations(t)
	})

	t.Run("deletion blocked by relationships", func(t *testing.T) {
		requirementID := uuid.New()
		requirement := &models.Requirement{ID: requirementID}

		mockRequirementRepo.On("GetByID", requirementID).Return(requirement, nil)
		mockRequirementRepo.On("HasRelationships", requirementID).Return(true, nil)

		err := service.DeleteRequirement(requirementID, false)

		assert.Error(t, err)
		assert.Equal(t, ErrRequirementHasRelationships, err)

		mockRequirementRepo.AssertExpectations(t)
	})

	t.Run("force deletion with relationships", func(t *testing.T) {
		requirementID := uuid.New()
		requirement := &models.Requirement{ID: requirementID}

		mockRequirementRepo.On("GetByID", requirementID).Return(requirement, nil)
		mockRequirementRepo.On("Delete", requirementID).Return(nil)

		err := service.DeleteRequirement(requirementID, true)

		assert.NoError(t, err)

		mockRequirementRepo.AssertExpectations(t)
	})

	t.Run("requirement not found", func(t *testing.T) {
		requirementID := uuid.New()

		mockRequirementRepo.On("GetByID", requirementID).Return(nil, repository.ErrNotFound)

		err := service.DeleteRequirement(requirementID, false)

		assert.Error(t, err)
		assert.Equal(t, ErrRequirementNotFound, err)

		mockRequirementRepo.AssertExpectations(t)
	})
}
