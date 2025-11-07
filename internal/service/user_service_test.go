package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

func TestUserService_GetByName_Username(t *testing.T) {
	mockRepo := new(MockUserRepository)
	expectedUser := &models.User{Username: "john_doe"}

	mockRepo.On("GetByUsername", "john_doe").Return(expectedUser, nil).Once()

	service := NewUserService(mockRepo)

	user, err := service.GetByName("john_doe")

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertCalled(t, "GetByUsername", "john_doe")
	mockRepo.AssertNotCalled(t, "GetByEmail", mock.Anything)
}

func TestUserService_GetByName_Email(t *testing.T) {
	mockRepo := new(MockUserRepository)
	expectedUser := &models.User{Email: "john@example.com"}

	mockRepo.On("GetByEmail", "john@example.com").Return(expectedUser, nil).Once()

	service := NewUserService(mockRepo)

	user, err := service.GetByName("john@example.com")

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertCalled(t, "GetByEmail", "john@example.com")
	mockRepo.AssertNotCalled(t, "GetByUsername", mock.Anything)
}

func TestUserService_GetByName_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)

	mockRepo.On("GetByUsername", "missing").Return((*models.User)(nil), repository.ErrNotFound).Once()

	service := NewUserService(mockRepo)

	user, err := service.GetByName("missing")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrUserNotFound))
	assert.Nil(t, user)
	mockRepo.AssertCalled(t, "GetByUsername", "missing")
}

func TestUserService_GetByName_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	repoErr := errors.New("database unavailable")

	mockRepo.On("GetByUsername", "john_doe").Return((*models.User)(nil), repoErr).Once()

	service := NewUserService(mockRepo)

	user, err := service.GetByName("john_doe")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, repoErr))
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetByName_EmptyName(t *testing.T) {
	mockRepo := new(MockUserRepository)

	service := NewUserService(mockRepo)

	user, err := service.GetByName("   ")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrValidation))
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}
