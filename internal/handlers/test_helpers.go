package handlers

import (
	"product-requirements-management/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// mockRefreshTokenRepository is a minimal mock for testing
type mockRefreshTokenRepository struct{}

func (m *mockRefreshTokenRepository) Create(token *models.RefreshToken) error {
	return nil
}

func (m *mockRefreshTokenRepository) FindByTokenHash(tokenHash string) (*models.RefreshToken, error) {
	return nil, nil
}

func (m *mockRefreshTokenRepository) FindByUserID(userID uuid.UUID) ([]*models.RefreshToken, error) {
	return nil, nil
}

func (m *mockRefreshTokenRepository) FindAll() ([]*models.RefreshToken, error) {
	return nil, nil
}

func (m *mockRefreshTokenRepository) Update(token *models.RefreshToken) error {
	return nil
}

func (m *mockRefreshTokenRepository) Delete(id uuid.UUID) error {
	return nil
}

func (m *mockRefreshTokenRepository) DeleteByUserID(userID uuid.UUID) error {
	return nil
}

func (m *mockRefreshTokenRepository) DeleteExpired() (int64, error) {
	return 0, nil
}

func (m *mockRefreshTokenRepository) GetDB() *gorm.DB {
	return nil
}
