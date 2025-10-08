package service

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashService defines the interface for secure hashing operations
type HashService interface {
	HashToken(token string) (string, error)
	CompareTokenWithHash(token, hash string) error
}

// BcryptHashService implements HashService using bcrypt for secure hashing
type BcryptHashService struct {
	cost int
}

// NewBcryptHashService creates a new instance of BcryptHashService with the specified cost
// The cost parameter determines the computational cost of hashing (higher = more secure but slower)
// Recommended values:
//   - bcrypt.DefaultCost (10) for general use
//   - 12-14 for high-security applications
//   - bcrypt.MinCost (4) only for testing
func NewBcryptHashService(cost int) (*BcryptHashService, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return nil, fmt.Errorf("invalid bcrypt cost %d, must be between %d and %d",
			cost, bcrypt.MinCost, bcrypt.MaxCost)
	}

	return &BcryptHashService{
		cost: cost,
	}, nil
}

// NewDefaultBcryptHashService creates a new BcryptHashService with the default cost (10)
func NewDefaultBcryptHashService() *BcryptHashService {
	return &BcryptHashService{
		cost: bcrypt.DefaultCost,
	}
}

// HashToken generates a bcrypt hash of the provided token
// The token should be the secret part only (without prefix) for security
//
// Parameters:
//   - token: The token string to hash (typically the secret part without prefix)
//
// Returns:
//   - string: The bcrypt hash of the token
//   - error: Any error that occurred during hashing
func (h *BcryptHashService) HashToken(token string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("token cannot be empty")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(token), h.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash token: %w", err)
	}

	return string(hashedBytes), nil
}

// CompareTokenWithHash performs a constant-time comparison between a token and its hash
// This method is resistant to timing attacks as it uses bcrypt's built-in comparison
//
// Parameters:
//   - token: The plaintext token to verify (typically the secret part without prefix)
//   - hash: The bcrypt hash to compare against
//
// Returns:
//   - error: nil if the token matches the hash, error otherwise
func (h *BcryptHashService) CompareTokenWithHash(token, hash string) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}
	if hash == "" {
		return fmt.Errorf("hash cannot be empty")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(token))
	if err != nil {
		// Don't wrap the error to avoid leaking information about the hash format
		return fmt.Errorf("token does not match hash")
	}

	return nil
}

// ValidateToken is a convenience method that combines token extraction and validation
// It extracts the secret part from a full token and compares it with the provided hash
//
// Parameters:
//   - fullToken: The complete token including prefix (e.g., "mcp_pat_secretpart")
//   - expectedPrefix: The expected prefix (e.g., "mcp_pat_")
//   - hash: The bcrypt hash to compare against
//
// Returns:
//   - error: nil if the token is valid, error otherwise
func (h *BcryptHashService) ValidateToken(fullToken, expectedPrefix, hash string) error {
	if fullToken == "" {
		return fmt.Errorf("token cannot be empty")
	}
	if expectedPrefix == "" {
		return fmt.Errorf("expected prefix cannot be empty")
	}
	if len(fullToken) <= len(expectedPrefix) {
		return fmt.Errorf("token is too short")
	}
	if fullToken[:len(expectedPrefix)] != expectedPrefix {
		return fmt.Errorf("token has invalid prefix")
	}

	// Extract the secret part (everything after the prefix)
	secretPart := fullToken[len(expectedPrefix):]

	return h.CompareTokenWithHash(secretPart, hash)
}
