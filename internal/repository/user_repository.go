package repository

import (
	"errors"

	"gorm.io/gorm"

	"product-requirements-management/internal/models"
)

// userRepository implements UserRepository interface
type userRepository struct {
	*BaseRepository[models.User]
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		BaseRepository: NewBaseRepository[models.User](db),
	}
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.GetDB().Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, r.handleDBError(err)
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, r.handleDBError(err)
	}
	return &user, nil
}

// ExistsByUsername checks if a user exists by username
func (r *userRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}
	return count > 0, nil
}

// ExistsByEmail checks if a user exists by email
func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}
	return count > 0, nil
}
