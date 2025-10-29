package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
)

// personalAccessTokenRepository implements PersonalAccessTokenRepository interface
type personalAccessTokenRepository struct {
	*BaseRepository[models.PersonalAccessToken]
}

// NewPersonalAccessTokenRepository creates a new personal access token repository instance
func NewPersonalAccessTokenRepository(db *gorm.DB) PersonalAccessTokenRepository {
	return &personalAccessTokenRepository{
		BaseRepository: NewBaseRepository[models.PersonalAccessToken](db),
	}
}

// GetByUserID retrieves all personal access tokens for a specific user
func (r *personalAccessTokenRepository) GetByUserID(userID uuid.UUID) ([]models.PersonalAccessToken, error) {
	var tokens []models.PersonalAccessToken
	if err := r.GetDB().Where("user_id = ?", userID).Order("created_at DESC").Find(&tokens).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return tokens, nil
}

// GetByUserIDWithPagination retrieves personal access tokens for a user with pagination
func (r *personalAccessTokenRepository) GetByUserIDWithPagination(userID uuid.UUID, limit, offset int) ([]models.PersonalAccessToken, int64, error) {
	var tokens []models.PersonalAccessToken
	var total int64

	// Get total count
	if err := r.GetDB().Model(&models.PersonalAccessToken{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, r.handleDBError(err)
	}

	// Get paginated results
	query := r.GetDB().Where("user_id = ?", userID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&tokens).Error; err != nil {
		return nil, 0, r.handleDBError(err)
	}

	return tokens, total, nil
}

// GetByUserIDWithPaginationAndPreloads retrieves personal access tokens for a user with pagination and User preloaded
func (r *personalAccessTokenRepository) GetByUserIDWithPaginationAndPreloads(userID uuid.UUID, limit, offset int) ([]models.PersonalAccessToken, int64, error) {
	var tokens []models.PersonalAccessToken
	var total int64

	// Get total count
	if err := r.GetDB().Model(&models.PersonalAccessToken{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, r.handleDBError(err)
	}

	// Get paginated results with User preloaded
	query := r.GetDB().Preload("User").Where("user_id = ?", userID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&tokens).Error; err != nil {
		return nil, 0, r.handleDBError(err)
	}

	return tokens, total, nil
}

// GetHashesByPrefix retrieves all token hashes that match a given prefix for authentication
func (r *personalAccessTokenRepository) GetHashesByPrefix(prefix string) ([]models.PersonalAccessToken, error) {
	var tokens []models.PersonalAccessToken
	if err := r.GetDB().Where("prefix = ?", prefix).Find(&tokens).Error; err != nil {
		return nil, r.handleDBError(err)
	}
	return tokens, nil
}

// UpdateLastUsed updates the last_used_at timestamp for a token
func (r *personalAccessTokenRepository) UpdateLastUsed(id uuid.UUID, lastUsedAt *time.Time) error {
	if err := r.GetDB().Model(&models.PersonalAccessToken{}).Where("id = ?", id).Update("last_used_at", lastUsedAt).Error; err != nil {
		return r.handleDBError(err)
	}
	return nil
}

// DeleteExpired removes all expired tokens and returns the count of deleted tokens
func (r *personalAccessTokenRepository) DeleteExpired() (int64, error) {
	now := time.Now()
	result := r.GetDB().Where("expires_at IS NOT NULL AND expires_at < ?", now).Delete(&models.PersonalAccessToken{})
	if result.Error != nil {
		return 0, r.handleDBError(result.Error)
	}
	return result.RowsAffected, nil
}

// ExistsByUserIDAndName checks if a token with the given name exists for a user
func (r *personalAccessTokenRepository) ExistsByUserIDAndName(userID uuid.UUID, name string) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.PersonalAccessToken{}).Where("user_id = ? AND name = ?", userID, name).Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}
	return count > 0, nil
}
