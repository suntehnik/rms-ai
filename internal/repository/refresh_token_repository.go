package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
)

// refreshTokenRepository implements RefreshTokenRepository interface
type refreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new refresh token repository instance
func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

// Create creates a new refresh token
func (r *refreshTokenRepository) Create(token *models.RefreshToken) error {
	if err := r.db.Create(token).Error; err != nil {
		return handleDBError(err)
	}
	return nil
}

// FindByTokenHash finds a refresh token by its hash
func (r *refreshTokenRepository) FindByTokenHash(tokenHash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	if err := r.db.Where("token_hash = ?", tokenHash).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, handleDBError(err)
	}
	return &token, nil
}

// FindByUserID finds all refresh tokens for a user
func (r *refreshTokenRepository) FindByUserID(userID uuid.UUID) ([]*models.RefreshToken, error) {
	var tokens []*models.RefreshToken
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&tokens).Error; err != nil {
		return nil, handleDBError(err)
	}
	return tokens, nil
}

// FindAll finds all refresh tokens
func (r *refreshTokenRepository) FindAll() ([]*models.RefreshToken, error) {
	var tokens []*models.RefreshToken
	if err := r.db.Find(&tokens).Error; err != nil {
		return nil, handleDBError(err)
	}
	return tokens, nil
}

// Update updates a refresh token
func (r *refreshTokenRepository) Update(token *models.RefreshToken) error {
	if err := r.db.Save(token).Error; err != nil {
		return handleDBError(err)
	}
	return nil
}

// Delete deletes a refresh token by ID
func (r *refreshTokenRepository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&models.RefreshToken{}, "id = ?", id)
	if result.Error != nil {
		return handleDBError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteByUserID deletes all refresh tokens for a user
func (r *refreshTokenRepository) DeleteByUserID(userID uuid.UUID) error {
	if err := r.db.Delete(&models.RefreshToken{}, "user_id = ?", userID).Error; err != nil {
		return handleDBError(err)
	}
	return nil
}

// DeleteExpired deletes all expired refresh tokens
func (r *refreshTokenRepository) DeleteExpired() (int64, error) {
	result := r.db.Delete(&models.RefreshToken{}, "expires_at < ?", time.Now())
	if result.Error != nil {
		return 0, handleDBError(result.Error)
	}
	return result.RowsAffected, nil
}

// GetDB returns the underlying database connection
func (r *refreshTokenRepository) GetDB() *gorm.DB {
	return r.db
}

// handleDBError converts database errors to repository errors
func handleDBError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrDuplicateKey
	}
	return err
}
