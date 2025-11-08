package service

import (
	"errors"
	"fmt"
	"strings"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

// UserService defines operations for retrieving users
type UserService interface {
	// GetByName finds a user by username or email
	GetByName(name string) (*models.User, error)
}

// userService provides user-related operations backed by a repository
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// GetByName looks up a user by username or email (if the value contains '@')
func (s *userService) GetByName(name string) (*models.User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("user name cannot be empty: %w", ErrValidation)
	}

	var (
		user *models.User
		err  error
	)

	if strings.Contains(name, "@") {
		user, err = s.userRepo.GetByEmail(name)
	} else {
		user, err = s.userRepo.GetByUsername(name)
	}

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by name %q: %w", name, err)
	}

	return user, nil
}
