package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
)

// MockPATRepository is a mock implementation of PersonalAccessTokenRepository
type MockPATRepository struct {
	mock.Mock
}

func (m *MockPATRepository) Create(pat *models.PersonalAccessToken) error {
	args := m.Called(pat)
	return args.Error(0)
}

func (m *MockPATRepository) GetByID(id uuid.UUID) (*models.PersonalAccessToken, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PersonalAccessToken), args.Error(1)
}

func (m *MockPATRepository) GetByReferenceID(referenceID string) (*models.PersonalAccessToken, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PersonalAccessToken), args.Error(1)
}

func (m *MockPATRepository) GetByReferenceIDCaseInsensitive(referenceID string) (*models.PersonalAccessToken, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PersonalAccessToken), args.Error(1)
}

func (m *MockPATRepository) Update(pat *models.PersonalAccessToken) error {
	args := m.Called(pat)
	return args.Error(0)
}

func (m *MockPATRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockPATRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.PersonalAccessToken, error) {
	args := m.Called(filters, orderBy, limit, offset)
	return args.Get(0).([]models.PersonalAccessToken), args.Error(1)
}

func (m *MockPATRepository) Count(filters map[string]interface{}) (int64, error) {
	args := m.Called(filters)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPATRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockPATRepository) ExistsByReferenceID(referenceID string) (bool, error) {
	args := m.Called(referenceID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPATRepository) WithTransaction(fn func(*gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockPATRepository) GetDB() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockPATRepository) GetByUserID(userID uuid.UUID) ([]models.PersonalAccessToken, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.PersonalAccessToken), args.Error(1)
}

func (m *MockPATRepository) GetByUserIDWithPagination(userID uuid.UUID, limit, offset int) ([]models.PersonalAccessToken, int64, error) {
	args := m.Called(userID, limit, offset)
	return args.Get(0).([]models.PersonalAccessToken), args.Get(1).(int64), args.Error(2)
}

func (m *MockPATRepository) GetHashesByPrefix(prefix string) ([]models.PersonalAccessToken, error) {
	args := m.Called(prefix)
	return args.Get(0).([]models.PersonalAccessToken), args.Error(1)
}

func (m *MockPATRepository) UpdateLastUsed(id uuid.UUID, lastUsedAt *time.Time) error {
	args := m.Called(id, lastUsedAt)
	return args.Error(0)
}

func (m *MockPATRepository) DeleteExpired() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPATRepository) ExistsByUserIDAndName(userID uuid.UUID, name string) (bool, error) {
	args := m.Called(userID, name)
	return args.Bool(0), args.Error(1)
}

// Note: MockUserRepository is already defined in epic_service_test.go

// MockTokenGenerator is a mock implementation of TokenGenerator
type MockTokenGenerator struct {
	mock.Mock
}

func (m *MockTokenGenerator) GenerateToken(prefix string, secretBytes int) (string, string, error) {
	args := m.Called(prefix, secretBytes)
	return args.String(0), args.String(1), args.Error(2)
}

// MockHashService is a mock implementation of HashService
type MockHashService struct {
	mock.Mock
}

func (m *MockHashService) HashToken(token string) (string, error) {
	args := m.Called(token)
	return args.String(0), args.Error(1)
}

func (m *MockHashService) CompareTokenWithHash(token, hash string) error {
	args := m.Called(token, hash)
	return args.Error(0)
}

func (m *MockHashService) ValidateToken(fullToken, expectedPrefix, hash string) error {
	args := m.Called(fullToken, expectedPrefix, hash)
	return args.Error(0)
}

// Test helper functions
func setupPATService() (*patService, *MockPATRepository, *MockUserRepository, *MockTokenGenerator, *MockHashService) {
	mockPATRepo := &MockPATRepository{}
	mockUserRepo := &MockUserRepository{}
	mockTokenGen := &MockTokenGenerator{}
	mockHashService := &MockHashService{}

	service := &patService{
		patRepo:     mockPATRepo,
		userRepo:    mockUserRepo,
		tokenGen:    mockTokenGen,
		hashService: mockHashService,
	}

	return service, mockPATRepo, mockUserRepo, mockTokenGen, mockHashService
}

func createTestUser() *models.User {
	return &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "User",
	}
}

func createTestPAT(userID uuid.UUID) *models.PersonalAccessToken {
	return &models.PersonalAccessToken{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      "Test Token",
		TokenHash: "$2a$12$hashedtoken",
		Prefix:    "mcp_pat_",
		Scopes:    `["full_access"]`,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Test CreatePAT
func TestCreatePAT_Success(t *testing.T) {
	service, mockPATRepo, mockUserRepo, mockTokenGen, mockHashService := setupPATService()
	ctx := context.Background()

	user := createTestUser()
	req := CreatePATRequest{
		Name:   "Test Token",
		Scopes: []string{"full_access"},
	}

	// Setup mocks
	mockUserRepo.On("GetByID", user.ID).Return(user, nil)
	mockPATRepo.On("ExistsByUserIDAndName", user.ID, req.Name).Return(false, nil)
	mockTokenGen.On("GenerateToken", "mcp_pat_", 32).Return("mcp_pat_secretpart", "secretpart", nil)
	mockHashService.On("HashToken", "secretpart").Return("$2a$12$hashedtoken", nil)
	mockPATRepo.On("Create", mock.AnythingOfType("*models.PersonalAccessToken")).Return(nil)

	// Execute
	result, err := service.CreatePAT(ctx, user.ID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "mcp_pat_secretpart", result.Token)
	assert.Equal(t, req.Name, result.PAT.Name)
	assert.Equal(t, user.ID, result.PAT.UserID)
	assert.Equal(t, "$2a$12$hashedtoken", result.PAT.TokenHash)

	mockUserRepo.AssertExpectations(t)
	mockPATRepo.AssertExpectations(t)
	mockTokenGen.AssertExpectations(t)
	mockHashService.AssertExpectations(t)
}

func TestCreatePAT_UserNotFound(t *testing.T) {
	service, _, mockUserRepo, _, _ := setupPATService()
	ctx := context.Background()

	userID := uuid.New()
	req := CreatePATRequest{Name: "Test Token"}

	// Setup mocks
	mockUserRepo.On("GetByID", userID).Return(nil, errors.New("user not found"))

	// Execute
	result, err := service.CreatePAT(ctx, userID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get user")

	mockUserRepo.AssertExpectations(t)
}

func TestCreatePAT_DuplicateName(t *testing.T) {
	service, mockPATRepo, mockUserRepo, _, _ := setupPATService()
	ctx := context.Background()

	user := createTestUser()
	req := CreatePATRequest{Name: "Test Token"}

	// Setup mocks
	mockUserRepo.On("GetByID", user.ID).Return(user, nil)
	mockPATRepo.On("ExistsByUserIDAndName", user.ID, req.Name).Return(true, nil)

	// Execute
	result, err := service.CreatePAT(ctx, user.ID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrPATDuplicateName, err)

	mockUserRepo.AssertExpectations(t)
	mockPATRepo.AssertExpectations(t)
}

func TestCreatePAT_TokenGenerationFails(t *testing.T) {
	service, mockPATRepo, mockUserRepo, mockTokenGen, _ := setupPATService()
	ctx := context.Background()

	user := createTestUser()
	req := CreatePATRequest{Name: "Test Token"}

	// Setup mocks
	mockUserRepo.On("GetByID", user.ID).Return(user, nil)
	mockPATRepo.On("ExistsByUserIDAndName", user.ID, req.Name).Return(false, nil)
	mockTokenGen.On("GenerateToken", "mcp_pat_", 32).Return("", "", errors.New("generation failed"))

	// Execute
	result, err := service.CreatePAT(ctx, user.ID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to generate token")

	mockUserRepo.AssertExpectations(t)
	mockPATRepo.AssertExpectations(t)
	mockTokenGen.AssertExpectations(t)
}

// Test ListUserPATs
func TestListUserPATs_Success(t *testing.T) {
	service, mockPATRepo, mockUserRepo, _, _ := setupPATService()
	ctx := context.Background()

	user := createTestUser()
	pat1 := createTestPAT(user.ID)
	pat2 := createTestPAT(user.ID)
	tokens := []models.PersonalAccessToken{*pat1, *pat2}

	// Setup mocks
	mockUserRepo.On("GetByID", user.ID).Return(user, nil)
	mockPATRepo.On("GetByUserIDWithPagination", user.ID, 50, 0).Return(tokens, int64(2), nil)

	// Execute
	result, err := service.ListUserPATs(ctx, user.ID, 0, 0)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, int64(2), result.TotalCount)
	assert.Equal(t, 50, result.Limit)
	assert.Equal(t, 0, result.Offset)

	mockUserRepo.AssertExpectations(t)
	mockPATRepo.AssertExpectations(t)
}

func TestListUserPATs_UserNotFound(t *testing.T) {
	service, _, mockUserRepo, _, _ := setupPATService()
	ctx := context.Background()

	userID := uuid.New()

	// Setup mocks
	mockUserRepo.On("GetByID", userID).Return(nil, errors.New("user not found"))

	// Execute
	result, err := service.ListUserPATs(ctx, userID, 10, 0)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get user")

	mockUserRepo.AssertExpectations(t)
}

// Test GetPAT
func TestGetPAT_Success(t *testing.T) {
	service, mockPATRepo, _, _, _ := setupPATService()
	ctx := context.Background()

	user := createTestUser()
	pat := createTestPAT(user.ID)

	// Setup mocks
	mockPATRepo.On("GetByID", pat.ID).Return(pat, nil)

	// Execute
	result, err := service.GetPAT(ctx, pat.ID, user.ID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, pat.ID, result.ID)
	assert.Equal(t, user.ID, result.UserID)

	mockPATRepo.AssertExpectations(t)
}

func TestGetPAT_NotFound(t *testing.T) {
	service, mockPATRepo, _, _, _ := setupPATService()
	ctx := context.Background()

	patID := uuid.New()
	userID := uuid.New()

	// Setup mocks
	mockPATRepo.On("GetByID", patID).Return(nil, errors.New("not found"))

	// Execute
	result, err := service.GetPAT(ctx, patID, userID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get PAT")

	mockPATRepo.AssertExpectations(t)
}

func TestGetPAT_Unauthorized(t *testing.T) {
	service, mockPATRepo, _, _, _ := setupPATService()
	ctx := context.Background()

	user := createTestUser()
	otherUserID := uuid.New()
	pat := createTestPAT(user.ID)

	// Setup mocks
	mockPATRepo.On("GetByID", pat.ID).Return(pat, nil)

	// Execute
	result, err := service.GetPAT(ctx, pat.ID, otherUserID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrPATUnauthorized, err)

	mockPATRepo.AssertExpectations(t)
}

// Test RevokePAT
func TestRevokePAT_Success(t *testing.T) {
	service, mockPATRepo, _, _, _ := setupPATService()
	ctx := context.Background()

	user := createTestUser()
	pat := createTestPAT(user.ID)

	// Setup mocks
	mockPATRepo.On("GetByID", pat.ID).Return(pat, nil)
	mockPATRepo.On("Delete", pat.ID).Return(nil)

	// Execute
	err := service.RevokePAT(ctx, pat.ID, user.ID)

	// Assert
	assert.NoError(t, err)

	mockPATRepo.AssertExpectations(t)
}

func TestRevokePAT_NotFound(t *testing.T) {
	service, mockPATRepo, _, _, _ := setupPATService()
	ctx := context.Background()

	patID := uuid.New()
	userID := uuid.New()

	// Setup mocks
	mockPATRepo.On("GetByID", patID).Return(nil, errors.New("not found"))

	// Execute
	err := service.RevokePAT(ctx, patID, userID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get PAT")

	mockPATRepo.AssertExpectations(t)
}

// Test ValidateToken
func TestValidateToken_Success(t *testing.T) {
	service, mockPATRepo, mockUserRepo, _, mockHashService := setupPATService()
	ctx := context.Background()

	user := createTestUser()
	pat := createTestPAT(user.ID)
	token := "mcp_pat_secretpart"
	secretPart := "secretpart"

	// Setup mocks
	mockPATRepo.On("GetHashesByPrefix", "mcp_pat_").Return([]models.PersonalAccessToken{*pat}, nil)
	mockHashService.On("CompareTokenWithHash", secretPart, pat.TokenHash).Return(nil)
	mockUserRepo.On("GetByID", user.ID).Return(user, nil)
	mockPATRepo.On("UpdateLastUsed", pat.ID, mock.AnythingOfType("*time.Time")).Return(nil)

	// Execute
	result, err := service.ValidateToken(ctx, token)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, user.Username, result.Username)

	mockPATRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockHashService.AssertExpectations(t)
}

func TestValidateToken_EmptyToken(t *testing.T) {
	service, _, _, _, _ := setupPATService()
	ctx := context.Background()

	// Execute
	result, err := service.ValidateToken(ctx, "")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrPATInvalidToken, err)
}

func TestValidateToken_InvalidPrefix(t *testing.T) {
	service, _, _, _, _ := setupPATService()
	ctx := context.Background()

	// Execute
	result, err := service.ValidateToken(ctx, "invalid_prefix_secretpart")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrPATInvalidPrefix, err)
}

func TestValidateToken_TooShort(t *testing.T) {
	service, _, _, _, _ := setupPATService()
	ctx := context.Background()

	// Execute
	result, err := service.ValidateToken(ctx, "mcp_pat_")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrPATInvalidToken, err)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	service, mockPATRepo, _, _, _ := setupPATService()
	ctx := context.Background()

	user := createTestUser()
	pat := createTestPAT(user.ID)
	// Set token as expired
	expiredTime := time.Now().Add(-1 * time.Hour)
	pat.ExpiresAt = &expiredTime

	token := "mcp_pat_secretpart"

	// Setup mocks
	mockPATRepo.On("GetHashesByPrefix", "mcp_pat_").Return([]models.PersonalAccessToken{*pat}, nil)

	// Execute
	result, err := service.ValidateToken(ctx, token)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrPATTokenHashMismatch, err)

	mockPATRepo.AssertExpectations(t)
}

func TestValidateToken_HashMismatch(t *testing.T) {
	service, mockPATRepo, _, _, mockHashService := setupPATService()
	ctx := context.Background()

	user := createTestUser()
	pat := createTestPAT(user.ID)
	token := "mcp_pat_secretpart"
	secretPart := "secretpart"

	// Setup mocks
	mockPATRepo.On("GetHashesByPrefix", "mcp_pat_").Return([]models.PersonalAccessToken{*pat}, nil)
	mockHashService.On("CompareTokenWithHash", secretPart, pat.TokenHash).Return(errors.New("hash mismatch"))

	// Execute
	result, err := service.ValidateToken(ctx, token)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrPATTokenHashMismatch, err)

	mockPATRepo.AssertExpectations(t)
	mockHashService.AssertExpectations(t)
}

func TestValidateToken_UserNotFound(t *testing.T) {
	service, mockPATRepo, mockUserRepo, _, mockHashService := setupPATService()
	ctx := context.Background()

	user := createTestUser()
	pat := createTestPAT(user.ID)
	token := "mcp_pat_secretpart"
	secretPart := "secretpart"

	// Setup mocks
	mockPATRepo.On("GetHashesByPrefix", "mcp_pat_").Return([]models.PersonalAccessToken{*pat}, nil)
	mockHashService.On("CompareTokenWithHash", secretPart, pat.TokenHash).Return(nil)
	mockUserRepo.On("GetByID", user.ID).Return(nil, errors.New("user not found"))

	// Execute
	result, err := service.ValidateToken(ctx, token)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get user for PAT")

	mockPATRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockHashService.AssertExpectations(t)
}

// Test UpdateLastUsed
func TestUpdateLastUsed_Success(t *testing.T) {
	service, mockPATRepo, _, _, _ := setupPATService()
	ctx := context.Background()

	patID := uuid.New()

	// Setup mocks
	mockPATRepo.On("UpdateLastUsed", patID, mock.AnythingOfType("*time.Time")).Return(nil)

	// Execute
	err := service.UpdateLastUsed(ctx, patID)

	// Assert
	assert.NoError(t, err)

	mockPATRepo.AssertExpectations(t)
}

func TestUpdateLastUsed_Error(t *testing.T) {
	service, mockPATRepo, _, _, _ := setupPATService()
	ctx := context.Background()

	patID := uuid.New()

	// Setup mocks
	mockPATRepo.On("UpdateLastUsed", patID, mock.AnythingOfType("*time.Time")).Return(errors.New("update failed"))

	// Execute
	err := service.UpdateLastUsed(ctx, patID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update last used timestamp")

	mockPATRepo.AssertExpectations(t)
}

// Test CleanupExpiredTokens
func TestCleanupExpiredTokens_Success(t *testing.T) {
	service, mockPATRepo, _, _, _ := setupPATService()
	ctx := context.Background()

	// Setup mocks
	mockPATRepo.On("DeleteExpired").Return(int64(5), nil)

	// Execute
	count, err := service.CleanupExpiredTokens(ctx)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 5, count)

	mockPATRepo.AssertExpectations(t)
}

func TestCleanupExpiredTokens_Error(t *testing.T) {
	service, mockPATRepo, _, _, _ := setupPATService()
	ctx := context.Background()

	// Setup mocks
	mockPATRepo.On("DeleteExpired").Return(int64(0), errors.New("cleanup failed"))

	// Execute
	count, err := service.CleanupExpiredTokens(ctx)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, 0, count)
	assert.Contains(t, err.Error(), "failed to cleanup expired tokens")

	mockPATRepo.AssertExpectations(t)
}

// Test validateScopes
func TestValidateScopes_ValidScopes(t *testing.T) {
	service, _, _, _, _ := setupPATService()

	// Test valid scopes
	validScopes := []string{"full_access"}
	err := service.validateScopes(validScopes)
	assert.NoError(t, err)
}

func TestValidateScopes_InvalidScopes(t *testing.T) {
	service, _, _, _, _ := setupPATService()

	// Test invalid scopes
	invalidScopes := []string{"invalid_scope"}
	err := service.validateScopes(invalidScopes)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid scope 'invalid_scope'")
}

func TestValidateScopes_MixedScopes(t *testing.T) {
	service, _, _, _, _ := setupPATService()

	// Test mixed valid and invalid scopes
	mixedScopes := []string{"full_access", "invalid_scope"}
	err := service.validateScopes(mixedScopes)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid scope 'invalid_scope'")
}

// Test scopesToJSON
func TestScopesToJSON_EmptyScopes(t *testing.T) {
	service, _, _, _, _ := setupPATService()

	result := service.scopesToJSON([]string{})
	assert.Equal(t, `["full_access"]`, result)
}

func TestScopesToJSON_SingleScope(t *testing.T) {
	service, _, _, _, _ := setupPATService()

	result := service.scopesToJSON([]string{"full_access"})
	assert.Equal(t, `["full_access"]`, result)
}

func TestScopesToJSON_MultipleScopes(t *testing.T) {
	service, _, _, _, _ := setupPATService()

	result := service.scopesToJSON([]string{"full_access", "read_only"})
	assert.Equal(t, `["full_access","read_only"]`, result)
}

// Test edge cases and error scenarios
func TestCreatePAT_HashingFails(t *testing.T) {
	service, mockPATRepo, mockUserRepo, mockTokenGen, mockHashService := setupPATService()
	ctx := context.Background()

	user := createTestUser()
	req := CreatePATRequest{Name: "Test Token"}

	// Setup mocks
	mockUserRepo.On("GetByID", user.ID).Return(user, nil)
	mockPATRepo.On("ExistsByUserIDAndName", user.ID, req.Name).Return(false, nil)
	mockTokenGen.On("GenerateToken", "mcp_pat_", 32).Return("mcp_pat_secretpart", "secretpart", nil)
	mockHashService.On("HashToken", "secretpart").Return("", errors.New("hashing failed"))

	// Execute
	result, err := service.CreatePAT(ctx, user.ID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to hash token")

	mockUserRepo.AssertExpectations(t)
	mockPATRepo.AssertExpectations(t)
	mockTokenGen.AssertExpectations(t)
	mockHashService.AssertExpectations(t)
}

func TestCreatePAT_DatabaseCreateFails(t *testing.T) {
	service, mockPATRepo, mockUserRepo, mockTokenGen, mockHashService := setupPATService()
	ctx := context.Background()

	user := createTestUser()
	req := CreatePATRequest{Name: "Test Token"}

	// Setup mocks
	mockUserRepo.On("GetByID", user.ID).Return(user, nil)
	mockPATRepo.On("ExistsByUserIDAndName", user.ID, req.Name).Return(false, nil)
	mockTokenGen.On("GenerateToken", "mcp_pat_", 32).Return("mcp_pat_secretpart", "secretpart", nil)
	mockHashService.On("HashToken", "secretpart").Return("$2a$12$hashedtoken", nil)
	mockPATRepo.On("Create", mock.AnythingOfType("*models.PersonalAccessToken")).Return(errors.New("database error"))

	// Execute
	result, err := service.CreatePAT(ctx, user.ID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create PAT")

	mockUserRepo.AssertExpectations(t)
	mockPATRepo.AssertExpectations(t)
	mockTokenGen.AssertExpectations(t)
	mockHashService.AssertExpectations(t)
}

func TestCreatePAT_WithExpirationDate(t *testing.T) {
	service, mockPATRepo, mockUserRepo, mockTokenGen, mockHashService := setupPATService()
	ctx := context.Background()

	user := createTestUser()
	expiresAt := time.Now().Add(24 * time.Hour)
	req := CreatePATRequest{
		Name:      "Test Token",
		ExpiresAt: &expiresAt,
		Scopes:    []string{"full_access"},
	}

	// Setup mocks
	mockUserRepo.On("GetByID", user.ID).Return(user, nil)
	mockPATRepo.On("ExistsByUserIDAndName", user.ID, req.Name).Return(false, nil)
	mockTokenGen.On("GenerateToken", "mcp_pat_", 32).Return("mcp_pat_secretpart", "secretpart", nil)
	mockHashService.On("HashToken", "secretpart").Return("$2a$12$hashedtoken", nil)
	mockPATRepo.On("Create", mock.AnythingOfType("*models.PersonalAccessToken")).Return(nil)

	// Execute
	result, err := service.CreatePAT(ctx, user.ID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "mcp_pat_secretpart", result.Token)
	assert.NotNil(t, result.PAT.ExpiresAt)
	assert.True(t, result.PAT.ExpiresAt.Equal(expiresAt))

	mockUserRepo.AssertExpectations(t)
	mockPATRepo.AssertExpectations(t)
	mockTokenGen.AssertExpectations(t)
	mockHashService.AssertExpectations(t)
}
