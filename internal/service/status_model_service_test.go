package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

// Mock repositories for testing
type MockStatusModelRepository struct {
	mock.Mock
}

func (m *MockStatusModelRepository) Create(statusModel *models.StatusModel) error {
	args := m.Called(statusModel)
	return args.Error(0)
}

func (m *MockStatusModelRepository) GetByID(id uuid.UUID) (*models.StatusModel, error) {
	args := m.Called(id)
	return args.Get(0).(*models.StatusModel), args.Error(1)
}

func (m *MockStatusModelRepository) GetByEntityTypeAndName(entityType models.EntityType, name string) (*models.StatusModel, error) {
	args := m.Called(entityType, name)
	return args.Get(0).(*models.StatusModel), args.Error(1)
}

func (m *MockStatusModelRepository) GetDefaultByEntityType(entityType models.EntityType) (*models.StatusModel, error) {
	args := m.Called(entityType)
	return args.Get(0).(*models.StatusModel), args.Error(1)
}

func (m *MockStatusModelRepository) Update(statusModel *models.StatusModel) error {
	args := m.Called(statusModel)
	return args.Error(0)
}

func (m *MockStatusModelRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockStatusModelRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.StatusModel, error) {
	args := m.Called(filters, orderBy, limit, offset)
	return args.Get(0).([]models.StatusModel), args.Error(1)
}

func (m *MockStatusModelRepository) ListByEntityType(entityType models.EntityType) ([]models.StatusModel, error) {
	args := m.Called(entityType)
	return args.Get(0).([]models.StatusModel), args.Error(1)
}

func (m *MockStatusModelRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockStatusModelRepository) ExistsByEntityTypeAndName(entityType models.EntityType, name string) (bool, error) {
	args := m.Called(entityType, name)
	return args.Bool(0), args.Error(1)
}

type MockStatusRepository struct {
	mock.Mock
}

func (m *MockStatusRepository) Create(status *models.Status) error {
	args := m.Called(status)
	return args.Error(0)
}

func (m *MockStatusRepository) GetByID(id uuid.UUID) (*models.Status, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Status), args.Error(1)
}

func (m *MockStatusRepository) GetByStatusModelID(statusModelID uuid.UUID) ([]models.Status, error) {
	args := m.Called(statusModelID)
	return args.Get(0).([]models.Status), args.Error(1)
}

func (m *MockStatusRepository) GetByName(statusModelID uuid.UUID, name string) (*models.Status, error) {
	args := m.Called(statusModelID, name)
	return args.Get(0).(*models.Status), args.Error(1)
}

func (m *MockStatusRepository) Update(status *models.Status) error {
	args := m.Called(status)
	return args.Error(0)
}

func (m *MockStatusRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockStatusRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.Status, error) {
	args := m.Called(filters, orderBy, limit, offset)
	return args.Get(0).([]models.Status), args.Error(1)
}

func (m *MockStatusRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockStatusRepository) ExistsByName(statusModelID uuid.UUID, name string) (bool, error) {
	args := m.Called(statusModelID, name)
	return args.Bool(0), args.Error(1)
}

type MockStatusTransitionRepository struct {
	mock.Mock
}

func (m *MockStatusTransitionRepository) Create(transition *models.StatusTransition) error {
	args := m.Called(transition)
	return args.Error(0)
}

func (m *MockStatusTransitionRepository) GetByID(id uuid.UUID) (*models.StatusTransition, error) {
	args := m.Called(id)
	return args.Get(0).(*models.StatusTransition), args.Error(1)
}

func (m *MockStatusTransitionRepository) GetByStatusModelID(statusModelID uuid.UUID) ([]models.StatusTransition, error) {
	args := m.Called(statusModelID)
	return args.Get(0).([]models.StatusTransition), args.Error(1)
}

func (m *MockStatusTransitionRepository) GetByFromStatus(fromStatusID uuid.UUID) ([]models.StatusTransition, error) {
	args := m.Called(fromStatusID)
	return args.Get(0).([]models.StatusTransition), args.Error(1)
}

func (m *MockStatusTransitionRepository) Update(transition *models.StatusTransition) error {
	args := m.Called(transition)
	return args.Error(0)
}

func (m *MockStatusTransitionRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockStatusTransitionRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.StatusTransition, error) {
	args := m.Called(filters, orderBy, limit, offset)
	return args.Get(0).([]models.StatusTransition), args.Error(1)
}

func (m *MockStatusTransitionRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockStatusTransitionRepository) ExistsByTransition(statusModelID, fromStatusID, toStatusID uuid.UUID) (bool, error) {
	args := m.Called(statusModelID, fromStatusID, toStatusID)
	return args.Bool(0), args.Error(1)
}

func TestCreateStatusModel(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockStatusModelRepo := new(MockStatusModelRepository)
		mockStatusRepo := new(MockStatusRepository)
		mockStatusTransitionRepo := new(MockStatusTransitionRepository)
		mockRequirementTypeRepo := new(MockRequirementTypeRepository)
		mockRelationshipTypeRepo := new(MockRelationshipTypeRepository)
		mockRequirementRepo := new(MockRequirementRepository)
		mockRequirementRelationRepo := new(MockRequirementRelationshipRepository)

		service := NewConfigService(
			mockRequirementTypeRepo,
			mockRelationshipTypeRepo,
			mockRequirementRepo,
			mockRequirementRelationRepo,
			mockStatusModelRepo,
			mockStatusRepo,
			mockStatusTransitionRepo,
		)

		req := CreateStatusModelRequest{
			EntityType:  models.EntityTypeEpic,
			Name:        "Test Status Model",
			Description: stringPtr("Test description"),
			IsDefault:   true,
		}

		mockStatusModelRepo.On("ExistsByEntityTypeAndName", models.EntityTypeEpic, "Test Status Model").Return(false, nil)
		mockStatusModelRepo.On("Create", mock.AnythingOfType("*models.StatusModel")).Return(nil)

		result, err := service.CreateStatusModel(req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, models.EntityTypeEpic, result.EntityType)
		assert.Equal(t, "Test Status Model", result.Name)
		assert.True(t, result.IsDefault)

		mockStatusModelRepo.AssertExpectations(t)
	})

	t.Run("invalid entity type", func(t *testing.T) {
		mockStatusModelRepo := new(MockStatusModelRepository)
		mockStatusRepo := new(MockStatusRepository)
		mockStatusTransitionRepo := new(MockStatusTransitionRepository)
		mockRequirementTypeRepo := new(MockRequirementTypeRepository)
		mockRelationshipTypeRepo := new(MockRelationshipTypeRepository)
		mockRequirementRepo := new(MockRequirementRepository)
		mockRequirementRelationRepo := new(MockRequirementRelationshipRepository)

		service := NewConfigService(
			mockRequirementTypeRepo,
			mockRelationshipTypeRepo,
			mockRequirementRepo,
			mockRequirementRelationRepo,
			mockStatusModelRepo,
			mockStatusRepo,
			mockStatusTransitionRepo,
		)

		req := CreateStatusModelRequest{
			EntityType: models.EntityType("invalid"),
			Name:       "Test Status Model",
		}

		result, err := service.CreateStatusModel(req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrInvalidEntityType, err)
	})

	t.Run("name already exists", func(t *testing.T) {
		mockStatusModelRepo := new(MockStatusModelRepository)
		mockStatusRepo := new(MockStatusRepository)
		mockStatusTransitionRepo := new(MockStatusTransitionRepository)
		mockRequirementTypeRepo := new(MockRequirementTypeRepository)
		mockRelationshipTypeRepo := new(MockRelationshipTypeRepository)
		mockRequirementRepo := new(MockRequirementRepository)
		mockRequirementRelationRepo := new(MockRequirementRelationshipRepository)

		service := NewConfigService(
			mockRequirementTypeRepo,
			mockRelationshipTypeRepo,
			mockRequirementRepo,
			mockRequirementRelationRepo,
			mockStatusModelRepo,
			mockStatusRepo,
			mockStatusTransitionRepo,
		)

		req := CreateStatusModelRequest{
			EntityType: models.EntityTypeEpic,
			Name:       "Test Status Model",
		}

		mockStatusModelRepo.On("ExistsByEntityTypeAndName", models.EntityTypeEpic, "Test Status Model").Return(true, nil)

		result, err := service.CreateStatusModel(req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrStatusModelNameExists, err)

		mockStatusModelRepo.AssertExpectations(t)
	})
}

func TestGetStatusModelByID(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		mockStatusModelRepo := new(MockStatusModelRepository)
		mockStatusRepo := new(MockStatusRepository)
		mockStatusTransitionRepo := new(MockStatusTransitionRepository)
		mockRequirementTypeRepo := new(MockRequirementTypeRepository)
		mockRelationshipTypeRepo := new(MockRelationshipTypeRepository)
		mockRequirementRepo := new(MockRequirementRepository)
		mockRequirementRelationRepo := new(MockRequirementRelationshipRepository)

		service := NewConfigService(
			mockRequirementTypeRepo,
			mockRelationshipTypeRepo,
			mockRequirementRepo,
			mockRequirementRelationRepo,
			mockStatusModelRepo,
			mockStatusRepo,
			mockStatusTransitionRepo,
		)

		id := uuid.New()
		expectedStatusModel := &models.StatusModel{
			ID:         id,
			EntityType: models.EntityTypeEpic,
			Name:       "Test Status Model",
		}

		mockStatusModelRepo.On("GetByID", id).Return(expectedStatusModel, nil)

		result, err := service.GetStatusModelByID(id)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedStatusModel.ID, result.ID)
		assert.Equal(t, expectedStatusModel.Name, result.Name)

		mockStatusModelRepo.AssertExpectations(t)
	})

	t.Run("status model not found", func(t *testing.T) {
		mockStatusModelRepo := new(MockStatusModelRepository)
		mockStatusRepo := new(MockStatusRepository)
		mockStatusTransitionRepo := new(MockStatusTransitionRepository)
		mockRequirementTypeRepo := new(MockRequirementTypeRepository)
		mockRelationshipTypeRepo := new(MockRelationshipTypeRepository)
		mockRequirementRepo := new(MockRequirementRepository)
		mockRequirementRelationRepo := new(MockRequirementRelationshipRepository)

		service := NewConfigService(
			mockRequirementTypeRepo,
			mockRelationshipTypeRepo,
			mockRequirementRepo,
			mockRequirementRelationRepo,
			mockStatusModelRepo,
			mockStatusRepo,
			mockStatusTransitionRepo,
		)

		id := uuid.New()

		mockStatusModelRepo.On("GetByID", id).Return((*models.StatusModel)(nil), repository.ErrNotFound)

		result, err := service.GetStatusModelByID(id)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrStatusModelNotFound, err)

		mockStatusModelRepo.AssertExpectations(t)
	})
}

func TestValidateStatusTransition(t *testing.T) {
	t.Run("valid transition with no explicit transitions defined", func(t *testing.T) {
		mockStatusModelRepo := new(MockStatusModelRepository)
		mockStatusRepo := new(MockStatusRepository)
		mockStatusTransitionRepo := new(MockStatusTransitionRepository)
		mockRequirementTypeRepo := new(MockRequirementTypeRepository)
		mockRelationshipTypeRepo := new(MockRelationshipTypeRepository)
		mockRequirementRepo := new(MockRequirementRepository)
		mockRequirementRelationRepo := new(MockRequirementRelationshipRepository)

		service := NewConfigService(
			mockRequirementTypeRepo,
			mockRelationshipTypeRepo,
			mockRequirementRepo,
			mockRequirementRelationRepo,
			mockStatusModelRepo,
			mockStatusRepo,
			mockStatusTransitionRepo,
		)

		statusModel := &models.StatusModel{
			ID:         uuid.New(),
			EntityType: models.EntityTypeEpic,
			Name:       "Default Epic Workflow",
			Transitions: []models.StatusTransition{}, // No explicit transitions = all allowed
		}

		fromStatus := &models.Status{
			ID:            uuid.New(),
			StatusModelID: statusModel.ID,
			Name:          "Backlog",
		}

		toStatus := &models.Status{
			ID:            uuid.New(),
			StatusModelID: statusModel.ID,
			Name:          "In Progress",
		}

		mockStatusModelRepo.On("GetDefaultByEntityType", models.EntityTypeEpic).Return(statusModel, nil)
		mockStatusRepo.On("GetByName", statusModel.ID, "Backlog").Return(fromStatus, nil)
		mockStatusRepo.On("GetByName", statusModel.ID, "In Progress").Return(toStatus, nil)

		err := service.ValidateStatusTransition(models.EntityTypeEpic, "Backlog", "In Progress")

		assert.NoError(t, err)

		mockStatusModelRepo.AssertExpectations(t)
		mockStatusRepo.AssertExpectations(t)
	})

	t.Run("no status model found - allow all transitions", func(t *testing.T) {
		mockStatusModelRepo := new(MockStatusModelRepository)
		mockStatusRepo := new(MockStatusRepository)
		mockStatusTransitionRepo := new(MockStatusTransitionRepository)
		mockRequirementTypeRepo := new(MockRequirementTypeRepository)
		mockRelationshipTypeRepo := new(MockRelationshipTypeRepository)
		mockRequirementRepo := new(MockRequirementRepository)
		mockRequirementRelationRepo := new(MockRequirementRelationshipRepository)

		service := NewConfigService(
			mockRequirementTypeRepo,
			mockRelationshipTypeRepo,
			mockRequirementRepo,
			mockRequirementRelationRepo,
			mockStatusModelRepo,
			mockStatusRepo,
			mockStatusTransitionRepo,
		)

		mockStatusModelRepo.On("GetDefaultByEntityType", models.EntityTypeEpic).Return((*models.StatusModel)(nil), repository.ErrNotFound)

		err := service.ValidateStatusTransition(models.EntityTypeEpic, "Backlog", "In Progress")

		assert.NoError(t, err)

		mockStatusModelRepo.AssertExpectations(t)
	})

	t.Run("invalid status name", func(t *testing.T) {
		mockStatusModelRepo := new(MockStatusModelRepository)
		mockStatusRepo := new(MockStatusRepository)
		mockStatusTransitionRepo := new(MockStatusTransitionRepository)
		mockRequirementTypeRepo := new(MockRequirementTypeRepository)
		mockRelationshipTypeRepo := new(MockRelationshipTypeRepository)
		mockRequirementRepo := new(MockRequirementRepository)
		mockRequirementRelationRepo := new(MockRequirementRelationshipRepository)

		service := NewConfigService(
			mockRequirementTypeRepo,
			mockRelationshipTypeRepo,
			mockRequirementRepo,
			mockRequirementRelationRepo,
			mockStatusModelRepo,
			mockStatusRepo,
			mockStatusTransitionRepo,
		)

		statusModel := &models.StatusModel{
			ID:         uuid.New(),
			EntityType: models.EntityTypeEpic,
			Name:       "Default Epic Workflow",
		}

		mockStatusModelRepo.On("GetDefaultByEntityType", models.EntityTypeEpic).Return(statusModel, nil)
		mockStatusRepo.On("GetByName", statusModel.ID, "InvalidStatus").Return((*models.Status)(nil), repository.ErrNotFound)

		err := service.ValidateStatusTransition(models.EntityTypeEpic, "InvalidStatus", "In Progress")

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidStatusTransition, err)

		mockStatusModelRepo.AssertExpectations(t)
		mockStatusRepo.AssertExpectations(t)
	})
}

