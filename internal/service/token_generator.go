package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// TokenGenerator defines the interface for generating secure tokens
type TokenGenerator interface {
	GenerateToken(prefix string, secretBytes int) (fullToken string, secretPart string, error error)
}

// SecureTokenGenerator implements TokenGenerator using crypto/rand for secure token generation
type SecureTokenGenerator struct{}

// NewSecureTokenGenerator creates a new instance of SecureTokenGenerator
func NewSecureTokenGenerator() *SecureTokenGenerator {
	return &SecureTokenGenerator{}
}

// GenerateToken generates a cryptographically secure token with the specified prefix
// and number of random bytes for the secret part.
//
// Parameters:
//   - prefix: The prefix to add to the token (e.g., "mcp_pat_")
//   - secretBytes: Number of random bytes to generate for the secret part (recommended: 32)
//
// Returns:
//   - fullToken: The complete token including prefix and secret
//   - secretPart: Just the secret part without prefix (for hashing)
//   - error: Any error that occurred during generation
func (g *SecureTokenGenerator) GenerateToken(prefix string, secretBytes int) (string, string, error) {
	if secretBytes <= 0 {
		return "", "", fmt.Errorf("secretBytes must be positive, got %d", secretBytes)
	}

	// Generate cryptographically secure random bytes
	secretData := make([]byte, secretBytes)
	if _, err := rand.Read(secretData); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode the secret part using URL-safe base64 without padding
	// This ensures the token is safe for use in URLs and HTTP headers
	secretPart := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(secretData)

	// Combine prefix and secret to create the full token
	fullToken := prefix + secretPart

	return fullToken, secretPart, nil
}

// GeneratePATToken is a convenience method that generates a Personal Access Token
// with the standard "mcp_pat_" prefix and 32 bytes of entropy (256 bits)
func (g *SecureTokenGenerator) GeneratePATToken() (string, string, error) {
	return g.GenerateToken("mcp_pat_", 32)
}
