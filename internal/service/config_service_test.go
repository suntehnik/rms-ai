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

// Mock repositories for testing (only the ones not already defined)
type MockConfigRequirementTypeRepository struct {
	mock.Mock
}

func (m *MockConfigRequirementTypeRepository) Create(entity *models.RequirementType) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockConfigRequirementTypeRepository) GetByID(id uuid.UUID) (*models.RequirementType, error) {
	args := m.Called(id)
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockConfigRequirementTypeRepository) GetByReferenceID(referenceID string) (*models.RequirementType, error) {
	args := m.Called(referenceID)
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockConfigRequirementTypeRepository) GetByReferenceIDCaseInsensitive(referenceID string) (*models.RequirementType, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockConfigRequirementTypeRepository) Update(entity *models.RequirementType) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockConfigRequirementTypeRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockConfigRequirementTypeRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.RequirementType, error) {
	args := m.Called(filters, orderBy, limit, offset)
	return args.Get(0).([]models.RequirementType), args.Error(1)
}

func (m *MockConfigRequirementTypeRepository) Count(filters map[string]interface{}) (int64, error) {
	args := m.Called(filters)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockConfigRequirementTypeRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockConfigRequirementTypeRepository) ExistsByReferenceID(referenceID string) (bool, error) {
	args := m.Called(referenceID)
	return args.Bool(0), args.Error(1)
}

func (m *MockConfigRequirementTypeRepository) WithTransaction(fn func(db *gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockConfigRequirementTypeRepository) GetDB() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *MockConfigRequirementTypeRepository) GetByName(name string) (*models.RequirementType, error) {
	args := m.Called(name)
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockConfigRequirementTypeRepository) ExistsByName(name string) (bool, error) {
	args := m.Called(name)
	return args.Bool(0), args.Error(1)
}

type MockConfigRelationshipTypeRepository struct {
	mock.Mock
}

func (m *MockConfigRelationshipTypeRepository) Create(entity *models.RelationshipType) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockConfigRelationshipTypeRepository) GetByID(id uuid.UUID) (*models.RelationshipType, error) {
	args := m.Called(id)
	return args.Get(0).(*models.RelationshipType), args.Error(1)
}

func (m *MockConfigRelationshipTypeRepository) GetByReferenceID(referenceID string) (*models.RelationshipType, error) {
	args := m.Called(referenceID)
	return args.Get(0).(*models.RelationshipType), args.Error(1)
}

func (m *MockConfigRelationshipTypeRepository) GetByReferenceIDCaseInsensitive(referenceID string) (*models.RelationshipType, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RelationshipType), args.Error(1)
}

func (m *MockConfigRelationshipTypeRepository) Update(entity *models.RelationshipType) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockConfigRelationshipTypeRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockConfigRelationshipTypeRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.RelationshipType, error) {
	args := m.Called(filters, orderBy, limit, offset)
	return args.Get(0).([]models.RelationshipType), args.Error(1)
}

func (m *MockConfigRelationshipTypeRepository) Count(filters map[string]interface{}) (int64, error) {
	args := m.Called(filters)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockConfigRelationshipTypeRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockConfigRelationshipTypeRepository) ExistsByReferenceID(referenceID string) (bool, error) {
	args := m.Called(referenceID)
	return args.Bool(0), args.Error(1)
}

func (m *MockConfigRelationshipTypeRepository) WithTransaction(fn func(db *gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockConfigRelationshipTypeRepository) GetDB() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *MockConfigRelationshipTypeRepository) GetByName(name string) (*models.RelationshipType, error) {
	args := m.Called(name)
	return args.Get(0).(*models.RelationshipType), args.Error(1)
}

func (m *MockConfigRelationshipTypeRepository) ExistsByName(name string) (bool, error) {
	args := m.Called(name)
	return args.Bool(0), args.Error(1)
}

type MockConfigRequirementRepository struct {
	mock.Mock
}

func (m *MockConfigRequirementRepository) GetByType(typeID uuid.UUID) ([]models.Requirement, error) {
	args := m.Called(typeID)
	return args.Get(0).([]models.Requirement), args.Error(1)
}

// Add other required methods (simplified for brevity)
func (m *MockConfigRequirementRepository) Create(entity *models.Requirement) error { return nil }
func (m *MockConfigRequirementRepository) GetByID(id uuid.UUID) (*models.Requirement, error) {
	return nil, nil
}
func (m *MockConfigRequirementRepository) GetByReferenceID(referenceID string) (*models.Requirement, error) {
	return nil, nil
}

func (m *MockConfigRequirementRepository) GetByReferenceIDCaseInsensitive(referenceID string) (*models.Requirement, error) {
	return nil, nil
}
func (m *MockConfigRequirementRepository) Update(entity *models.Requirement) error { return nil }
func (m *MockConfigRequirementRepository) Delete(id uuid.UUID) error               { return nil }
func (m *MockConfigRequirementRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.Requirement, error) {
	return nil, nil
}
func (m *MockConfigRequirementRepository) Count(filters map[string]interface{}) (int64, error) {
	return 0, nil
}
func (m *MockConfigRequirementRepository) Exists(id uuid.UUID) (bool, error) { return false, nil }
func (m *MockConfigRequirementRepository) ExistsByReferenceID(referenceID string) (bool, error) {
	return false, nil
}
func (m *MockConfigRequirementRepository) WithTransaction(fn func(db *gorm.DB) error) error {
	return nil
}
func (m *MockConfigRequirementRepository) GetDB() *gorm.DB { return nil }
func (m *MockConfigRequirementRepository) GetWithRelationships(id uuid.UUID) (*models.Requirement, error) {
	return nil, nil
}
func (m *MockConfigRequirementRepository) GetByUserStory(userStoryID uuid.UUID) ([]models.Requirement, error) {
	return nil, nil
}
func (m *MockConfigRequirementRepository) GetByAcceptanceCriteria(acceptanceCriteriaID uuid.UUID) ([]models.Requirement, error) {
	return nil, nil
}
func (m *MockConfigRequirementRepository) GetByCreator(creatorID uuid.UUID) ([]models.Requirement, error) {
	return nil, nil
}
func (m *MockConfigRequirementRepository) GetByAssignee(assigneeID uuid.UUID) ([]models.Requirement, error) {
	return nil, nil
}
func (m *MockConfigRequirementRepository) GetByStatus(status models.RequirementStatus) ([]models.Requirement, error) {
	return nil, nil
}
func (m *MockConfigRequirementRepository) GetByPriority(priority models.Priority) ([]models.Requirement, error) {
	return nil, nil
}
func (m *MockConfigRequirementRepository) HasRelationships(id uuid.UUID) (bool, error) {
	return false, nil
}
func (m *MockConfigRequirementRepository) SearchByText(searchText string) ([]models.Requirement, error) {
	return nil, nil
}
func (m *MockConfigRequirementRepository) SearchByTextWithPagination(searchText string, limit, offset int) ([]models.Requirement, int64, error) {
	return nil, 0, nil
}
func (m *MockConfigRequirementRepository) GetByIDWithPreloads(id uuid.UUID) (*models.Requirement, error) {
	return nil, nil
}
func (m *MockConfigRequirementRepository) GetByReferenceIDWithPreloads(referenceID string) (*models.Requirement, error) {
	return nil, nil
}
func (m *MockConfigRequirementRepository) ListWithPreloads(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.Requirement, error) {
	return nil, nil
}

type MockConfigRequirementRelationshipRepository struct {
	mock.Mock
}

func (m *MockConfigRequirementRelationshipRepository) GetByType(typeID uuid.UUID) ([]models.RequirementRelationship, error) {
	args := m.Called(typeID)
	return args.Get(0).([]models.RequirementRelationship), args.Error(1)
}

// Add other required methods (simplified for brevity)
func (m *MockConfigRequirementRelationshipRepository) Create(entity *models.RequirementRelationship) error {
	return nil
}
func (m *MockConfigRequirementRelationshipRepository) GetByID(id uuid.UUID) (*models.RequirementRelationship, error) {
	return nil, nil
}
func (m *MockConfigRequirementRelationshipRepository) GetByReferenceID(referenceID string) (*models.RequirementRelationship, error) {
	return nil, nil
}

func (m *MockConfigRequirementRelationshipRepository) GetByReferenceIDCaseInsensitive(referenceID string) (*models.RequirementRelationship, error) {
	return nil, nil
}
func (m *MockConfigRequirementRelationshipRepository) Update(entity *models.RequirementRelationship) error {
	return nil
}
func (m *MockConfigRequirementRelationshipRepository) Delete(id uuid.UUID) error { return nil }
func (m *MockConfigRequirementRelationshipRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.RequirementRelationship, error) {
	return nil, nil
}
func (m *MockConfigRequirementRelationshipRepository) Count(filters map[string]interface{}) (int64, error) {
	return 0, nil
}
func (m *MockConfigRequirementRelationshipRepository) Exists(id uuid.UUID) (bool, error) {
	return false, nil
}
func (m *MockConfigRequirementRelationshipRepository) ExistsByReferenceID(referenceID string) (bool, error) {
	return false, nil
}
func (m *MockConfigRequirementRelationshipRepository) WithTransaction(fn func(db *gorm.DB) error) error {
	return nil
}
func (m *MockConfigRequirementRelationshipRepository) GetDB() *gorm.DB { return nil }
func (m *MockConfigRequirementRelationshipRepository) GetBySourceRequirement(sourceID uuid.UUID) ([]models.RequirementRelationship, error) {
	return nil, nil
}
func (m *MockConfigRequirementRelationshipRepository) GetByTargetRequirement(targetID uuid.UUID) ([]models.RequirementRelationship, error) {
	return nil, nil
}
func (m *MockConfigRequirementRelationshipRepository) GetByRequirement(requirementID uuid.UUID) ([]models.RequirementRelationship, error) {
	return nil, nil
}
func (m *MockConfigRequirementRelationshipRepository) ExistsRelationship(sourceID, targetID, typeID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *MockConfigRequirementRelationshipRepository) GetByRequirementWithPagination(requirementID uuid.UUID, limit, offset int) ([]models.RequirementRelationship, int64, error) {
	return nil, 0, nil
}

func TestConfigService_CreateRequirementType(t *testing.T) {
	mockRequirementTypeRepo := &MockConfigRequirementTypeRepository{}
	mockRelationshipTypeRepo := &MockConfigRelationshipTypeRepository{}
	mockRequirementRepo := &MockConfigRequirementRepository{}
	mockRequirementRelationRepo := &MockConfigRequirementRelationshipRepository{}

	service := NewConfigService(
		mockRequirementTypeRepo,
		mockRelationshipTypeRepo,
		mockRequirementRepo,
		mockRequirementRelationRepo,
		nil, // statusModelRepo - not needed for this test
		nil, // statusRepo - not needed for this test
		nil, // statusTransitionRepo - not needed for this test
	)

	t.Run("successful creation", func(t *testing.T) {
		req := CreateRequirementTypeRequest{
			Name:        "Test Type",
			Description: stringPtrConfig("Test description"),
		}

		mockRequirementTypeRepo.On("ExistsByName", "Test Type").Return(false, nil)
		mockRequirementTypeRepo.On("Create", mock.AnythingOfType("*models.RequirementType")).Return(nil)

		result, err := service.CreateRequirementType(req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Type", result.Name)
		assert.Equal(t, "Test description", *result.Description)
		mockRequirementTypeRepo.AssertExpectations(t)
	})

	t.Run("name already exists", func(t *testing.T) {
		req := CreateRequirementTypeRequest{
			Name: "Existing Type",
		}

		mockRequirementTypeRepo.On("ExistsByName", "Existing Type").Return(true, nil)

		result, err := service.CreateRequirementType(req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrRequirementTypeNameExists, err)
		mockRequirementTypeRepo.AssertExpectations(t)
	})
}

func TestConfigService_GetRequirementTypeByID(t *testing.T) {
	mockRequirementTypeRepo := &MockConfigRequirementTypeRepository{}
	mockRelationshipTypeRepo := &MockConfigRelationshipTypeRepository{}
	mockRequirementRepo := &MockConfigRequirementRepository{}
	mockRequirementRelationRepo := &MockConfigRequirementRelationshipRepository{}

	service := NewConfigService(
		mockRequirementTypeRepo,
		mockRelationshipTypeRepo,
		mockRequirementRepo,
		mockRequirementRelationRepo,
		nil, // statusModelRepo - not needed for this test
		nil, // statusRepo - not needed for this test
		nil, // statusTransitionRepo - not needed for this test
	)

	t.Run("successful retrieval", func(t *testing.T) {
		id := uuid.New()
		expectedType := &models.RequirementType{
			ID:   id,
			Name: "Test Type",
		}

		mockRequirementTypeRepo.On("GetByID", id).Return(expectedType, nil)

		result, err := service.GetRequirementTypeByID(id)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedType.ID, result.ID)
		assert.Equal(t, expectedType.Name, result.Name)
		mockRequirementTypeRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		id := uuid.New()

		mockRequirementTypeRepo.On("GetByID", id).Return((*models.RequirementType)(nil), repository.ErrNotFound)

		result, err := service.GetRequirementTypeByID(id)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrRequirementTypeNotFound, err)
		mockRequirementTypeRepo.AssertExpectations(t)
	})
}

func TestConfigService_DeleteRequirementType(t *testing.T) {
	mockRequirementTypeRepo := &MockConfigRequirementTypeRepository{}
	mockRelationshipTypeRepo := &MockConfigRelationshipTypeRepository{}
	mockRequirementRepo := &MockConfigRequirementRepository{}
	mockRequirementRelationRepo := &MockConfigRequirementRelationshipRepository{}

	service := NewConfigService(
		mockRequirementTypeRepo,
		mockRelationshipTypeRepo,
		mockRequirementRepo,
		mockRequirementRelationRepo,
		nil, // statusModelRepo - not needed for this test
		nil, // statusRepo - not needed for this test
		nil, // statusTransitionRepo - not needed for this test
	)

	t.Run("successful deletion with no requirements", func(t *testing.T) {
		id := uuid.New()
		requirementType := &models.RequirementType{
			ID:   id,
			Name: "Test Type",
		}

		mockRequirementTypeRepo.On("GetByID", id).Return(requirementType, nil)
		mockRequirementRepo.On("GetByType", id).Return([]models.Requirement{}, nil)
		mockRequirementTypeRepo.On("Delete", id).Return(nil)

		err := service.DeleteRequirementType(id, false)

		assert.NoError(t, err)
		mockRequirementTypeRepo.AssertExpectations(t)
		mockRequirementRepo.AssertExpectations(t)
	})

	t.Run("deletion blocked by existing requirements", func(t *testing.T) {
		id := uuid.New()
		requirementType := &models.RequirementType{
			ID:   id,
			Name: "Test Type",
		}
		requirements := []models.Requirement{
			{ID: uuid.New(), TypeID: id},
		}

		mockRequirementTypeRepo.On("GetByID", id).Return(requirementType, nil)
		mockRequirementRepo.On("GetByType", id).Return(requirements, nil)

		err := service.DeleteRequirementType(id, false)

		assert.Error(t, err)
		assert.Equal(t, ErrRequirementTypeHasRequirements, err)
		mockRequirementTypeRepo.AssertExpectations(t)
		mockRequirementRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		id := uuid.New()

		mockRequirementTypeRepo.On("GetByID", id).Return((*models.RequirementType)(nil), repository.ErrNotFound)

		err := service.DeleteRequirementType(id, false)

		assert.Error(t, err)
		assert.Equal(t, ErrRequirementTypeNotFound, err)
		mockRequirementTypeRepo.AssertExpectations(t)
	})
}

func TestConfigService_ValidateRequirementType(t *testing.T) {
	mockRequirementTypeRepo := &MockConfigRequirementTypeRepository{}
	mockRelationshipTypeRepo := &MockConfigRelationshipTypeRepository{}
	mockRequirementRepo := &MockConfigRequirementRepository{}
	mockRequirementRelationRepo := &MockConfigRequirementRelationshipRepository{}

	service := NewConfigService(
		mockRequirementTypeRepo,
		mockRelationshipTypeRepo,
		mockRequirementRepo,
		mockRequirementRelationRepo,
		nil, // statusModelRepo - not needed for this test
		nil, // statusRepo - not needed for this test
		nil, // statusTransitionRepo - not needed for this test
	)

	t.Run("valid requirement type", func(t *testing.T) {
		id := uuid.New()

		mockRequirementTypeRepo.On("Exists", id).Return(true, nil)

		err := service.ValidateRequirementType(id)

		assert.NoError(t, err)
		mockRequirementTypeRepo.AssertExpectations(t)
	})

	t.Run("invalid requirement type", func(t *testing.T) {
		id := uuid.New()

		mockRequirementTypeRepo.On("Exists", id).Return(false, nil)

		err := service.ValidateRequirementType(id)

		assert.Error(t, err)
		assert.Equal(t, ErrRequirementTypeNotFound, err)
		mockRequirementTypeRepo.AssertExpectations(t)
	})
}

func TestConfigService_ValidateRelationshipType(t *testing.T) {
	mockRequirementTypeRepo := &MockConfigRequirementTypeRepository{}
	mockRelationshipTypeRepo := &MockConfigRelationshipTypeRepository{}
	mockRequirementRepo := &MockConfigRequirementRepository{}
	mockRequirementRelationRepo := &MockConfigRequirementRelationshipRepository{}

	service := NewConfigService(
		mockRequirementTypeRepo,
		mockRelationshipTypeRepo,
		mockRequirementRepo,
		mockRequirementRelationRepo,
		nil, // statusModelRepo - not needed for this test
		nil, // statusRepo - not needed for this test
		nil, // statusTransitionRepo - not needed for this test
	)

	t.Run("valid relationship type", func(t *testing.T) {
		id := uuid.New()

		mockRelationshipTypeRepo.On("Exists", id).Return(true, nil)

		err := service.ValidateRelationshipType(id)

		assert.NoError(t, err)
		mockRelationshipTypeRepo.AssertExpectations(t)
	})

	t.Run("invalid relationship type", func(t *testing.T) {
		id := uuid.New()

		mockRelationshipTypeRepo.On("Exists", id).Return(false, nil)

		err := service.ValidateRelationshipType(id)

		assert.Error(t, err)
		assert.Equal(t, ErrRelationshipTypeNotFound, err)
		mockRelationshipTypeRepo.AssertExpectations(t)
	})
}

// Helper function for string pointers (with unique name to avoid conflicts)
func stringPtrConfig(s string) *string {
	return &s
}
