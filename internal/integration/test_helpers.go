package integration

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
)

// createTestUser creates a test user for integration tests
func createTestUser(t *testing.T, db *gorm.DB) *models.User {
	user := &models.User{
		ID:       uuid.New(),
		Username: fmt.Sprintf("testuser_%s", uuid.New().String()[:8]),
		Email:    fmt.Sprintf("test_%s@example.com", uuid.New().String()[:8]),
		Role:     models.RoleUser,
	}

	err := db.Create(user).Error
	require.NoError(t, err)

	return user
}
