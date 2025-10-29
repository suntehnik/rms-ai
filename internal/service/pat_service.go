package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

var (
	ErrPATNotFound          = errors.New("personal access token not found")
	ErrPATExpired           = errors.New("personal access token has expired")
	ErrPATInvalidToken      = errors.New("invalid token format")
	ErrPATInvalidPrefix     = errors.New("invalid token prefix")
	ErrPATDuplicateName     = errors.New("token name already exists for user")
	ErrPATUserNotFound      = errors.New("user not found")
	ErrPATUnauthorized      = errors.New("unauthorized access to token")
	ErrPATInvalidScopes     = errors.New("invalid scopes specified")
	ErrPATTokenHashMismatch = errors.New("token does not match stored hash")
)

// PATService defines the interface for Personal Access Token business logic
type PATService interface {
	// Token Management
	CreatePAT(ctx context.Context, userID uuid.UUID, req CreatePATRequest) (*PATCreateResponse, error)
	ListUserPATs(ctx context.Context, userID uuid.UUID, limit, offset int) (*ListResponse[models.PersonalAccessToken], error)
	GetPAT(ctx context.Context, patID, userID uuid.UUID) (*models.PersonalAccessToken, error)
	RevokePAT(ctx context.Context, patID, userID uuid.UUID) error

	// Authentication
	ValidateToken(ctx context.Context, token string) (*models.User, error)
	UpdateLastUsed(ctx context.Context, patID uuid.UUID) error

	// Maintenance
	CleanupExpiredTokens(ctx context.Context) (int, error)
}

// CreatePATRequest represents the request to create a personal access token
type CreatePATRequest struct {
	// Name is a descriptive name for the token (required, 1-255 characters)
	Name string `json:"name" binding:"required,min=1,max=255"`
	// ExpiresAt is the optional expiration date for the token
	ExpiresAt *time.Time `json:"expires_at"`
	// Scopes defines the permissions for the token (defaults to ["full_access"])
	Scopes []string `json:"scopes"`
}

// PATCreateResponse represents the response when creating a PAT (includes the full token)
type PATCreateResponse struct {
	// Token is the full PAT token - returned only once during creation
	Token string `json:"token"`
	// PAT contains the token metadata (without the actual token value)
	PAT models.PersonalAccessToken `json:"pat"`
}

// ListResponse represents a paginated list response
type ListResponse[T any] struct {
	// Data contains the list of items for the current page
	Data []T `json:"data"`
	// TotalCount is the total number of items across all pages
	TotalCount int64 `json:"total_count"`
	// Limit is the maximum number of items per page
	Limit int `json:"limit"`
	// Offset is the number of items skipped for pagination
	Offset int `json:"offset"`
}

// patService implements the PATService interface
type patService struct {
	patRepo     repository.PersonalAccessTokenRepository
	userRepo    repository.UserRepository
	tokenGen    TokenGenerator
	hashService HashService
}

// NewPATService creates a new PAT service instance
func NewPATService(
	patRepo repository.PersonalAccessTokenRepository,
	userRepo repository.UserRepository,
	tokenGen TokenGenerator,
	hashService HashService,
) PATService {
	return &patService{
		patRepo:     patRepo,
		userRepo:    userRepo,
		tokenGen:    tokenGen,
		hashService: hashService,
	}
}

// CreatePAT creates a new personal access token for a user
func (s *patService) CreatePAT(ctx context.Context, userID uuid.UUID, req CreatePATRequest) (*PATCreateResponse, error) {
	// Validate user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrPATUserNotFound
	}

	// Check if token name already exists for this user
	exists, err := s.patRepo.ExistsByUserIDAndName(userID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check token name uniqueness: %w", err)
	}
	if exists {
		return nil, ErrPATDuplicateName
	}

	// Validate scopes (for now, only support full_access)
	scopes := req.Scopes
	if len(scopes) == 0 {
		scopes = []string{"full_access"}
	}
	if err := s.validateScopes(scopes); err != nil {
		return nil, err
	}

	// Generate secure token
	fullToken, secretPart, err := s.tokenGen.GenerateToken("mcp_pat_", 32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Hash the secret part for storage
	tokenHash, err := s.hashService.HashToken(secretPart)
	if err != nil {
		return nil, fmt.Errorf("failed to hash token: %w", err)
	}

	// Create PAT model
	pat := &models.PersonalAccessToken{
		UserID:    userID,
		Name:      req.Name,
		TokenHash: tokenHash,
		Prefix:    "mcp_pat_",
		Scopes:    s.scopesToJSON(scopes),
		ExpiresAt: req.ExpiresAt,
	}

	// Save to database
	if err := s.patRepo.Create(pat); err != nil {
		return nil, fmt.Errorf("failed to create PAT: %w", err)
	}

	// Return response with full token (only time it's exposed)
	return &PATCreateResponse{
		Token: fullToken,
		PAT:   *pat,
	}, nil
}

// ListUserPATs retrieves all PATs for a user with pagination
func (s *patService) ListUserPATs(ctx context.Context, userID uuid.UUID, limit, offset int) (*ListResponse[models.PersonalAccessToken], error) {
	// Validate user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrPATUserNotFound
	}

	// Set default pagination limits
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Get paginated tokens with User preloaded
	tokens, total, err := s.patRepo.GetByUserIDWithPaginationAndPreloads(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user PATs: %w", err)
	}

	return &ListResponse[models.PersonalAccessToken]{
		Data:       tokens,
		TotalCount: total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

// GetPAT retrieves a specific PAT by ID for a user
func (s *patService) GetPAT(ctx context.Context, patID, userID uuid.UUID) (*models.PersonalAccessToken, error) {
	pat, err := s.patRepo.GetByID(patID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PAT: %w", err)
	}
	if pat == nil {
		return nil, ErrPATNotFound
	}

	// Ensure the PAT belongs to the requesting user
	if pat.UserID != userID {
		return nil, ErrPATUnauthorized
	}

	return pat, nil
}

// RevokePAT deletes a PAT for a user
func (s *patService) RevokePAT(ctx context.Context, patID, userID uuid.UUID) error {
	// First verify the PAT exists and belongs to the user
	pat, err := s.GetPAT(ctx, patID, userID)
	if err != nil {
		return err
	}

	// Delete the PAT
	if err := s.patRepo.Delete(pat.ID); err != nil {
		return fmt.Errorf("failed to revoke PAT: %w", err)
	}

	return nil
}

// validateScopes validates the provided scopes
func (s *patService) validateScopes(scopes []string) error {
	validScopes := map[string]bool{
		"full_access": true,
		// Future scopes can be added here
		// "read_only": true,
		// "epics:read": true,
		// "epics:write": true,
	}

	for _, scope := range scopes {
		if !validScopes[scope] {
			return fmt.Errorf("%w: invalid scope '%s'", ErrPATInvalidScopes, scope)
		}
	}

	return nil
}

// scopesToJSON converts scopes slice to JSON string
func (s *patService) scopesToJSON(scopes []string) string {
	if len(scopes) == 0 {
		return `["full_access"]`
	}

	// Simple JSON encoding for scopes
	var parts []string
	for _, scope := range scopes {
		parts = append(parts, fmt.Sprintf(`"%s"`, scope))
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, ","))
}

// ValidateToken validates a PAT and returns the associated user
func (s *patService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	// Validate token format
	if token == "" {
		return nil, ErrPATInvalidToken
	}

	// Extract prefix and validate
	const expectedPrefix = "mcp_pat_"
	if !strings.HasPrefix(token, expectedPrefix) {
		return nil, ErrPATInvalidPrefix
	}

	// Extract secret part
	if len(token) <= len(expectedPrefix) {
		return nil, ErrPATInvalidToken
	}
	secretPart := token[len(expectedPrefix):]

	// Get all tokens with this prefix
	tokens, err := s.patRepo.GetHashesByPrefix(expectedPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to get tokens by prefix: %w", err)
	}

	// Try to match the token against stored hashes
	for _, pat := range tokens {
		// Check if token is expired
		if pat.ExpiresAt != nil && time.Now().After(*pat.ExpiresAt) {
			continue // Skip expired tokens
		}

		// Compare token with hash using constant-time comparison
		if err := s.hashService.CompareTokenWithHash(secretPart, pat.TokenHash); err == nil {
			// Token matches, get the user
			user, err := s.userRepo.GetByID(pat.UserID)
			if err != nil {
				return nil, fmt.Errorf("failed to get user for PAT: %w", err)
			}
			if user == nil {
				return nil, ErrPATUserNotFound
			}

			// Update last used timestamp (in production this could be async)
			now := time.Now()
			if updateErr := s.patRepo.UpdateLastUsed(pat.ID, &now); updateErr != nil {
				// Log error but don't fail the authentication
				// In a real application, you'd use a proper logger here
				fmt.Printf("Warning: failed to update last used timestamp for PAT %s: %v\n", pat.ID, updateErr)
			}

			return user, nil
		}
	}

	// No matching token found
	return nil, ErrPATTokenHashMismatch
}

// UpdateLastUsed updates the last used timestamp for a PAT
func (s *patService) UpdateLastUsed(ctx context.Context, patID uuid.UUID) error {
	now := time.Now()
	if err := s.patRepo.UpdateLastUsed(patID, &now); err != nil {
		return fmt.Errorf("failed to update last used timestamp: %w", err)
	}
	return nil
}

// CleanupExpiredTokens removes all expired tokens and returns the count
func (s *patService) CleanupExpiredTokens(ctx context.Context) (int, error) {
	count, err := s.patRepo.DeleteExpired()
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}
	return int(count), nil
}
