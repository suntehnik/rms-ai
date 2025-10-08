package service

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecureTokenGenerator_GenerateToken(t *testing.T) {
	generator := NewSecureTokenGenerator()

	t.Run("generates token with correct format", func(t *testing.T) {
		prefix := "mcp_pat_"
		secretBytes := 32

		fullToken, secretPart, err := generator.GenerateToken(prefix, secretBytes)

		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(fullToken, prefix), "Token should start with prefix")
		assert.Equal(t, fullToken, prefix+secretPart, "Full token should be prefix + secret")

		// Verify the secret part is valid base64url
		_, err = base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(secretPart)
		assert.NoError(t, err, "Secret part should be valid base64url")
	})

	t.Run("generates tokens with correct length", func(t *testing.T) {
		testCases := []struct {
			name        string
			secretBytes int
			// base64url encoding of N bytes results in ceil(N*4/3) characters without padding
			expectedSecretLength int
		}{
			{"16 bytes", 16, 22}, // 16*4/3 = 21.33 -> 22
			{"32 bytes", 32, 43}, // 32*4/3 = 42.67 -> 43
			{"64 bytes", 64, 86}, // 64*4/3 = 85.33 -> 86
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				fullToken, secretPart, err := generator.GenerateToken("test_", tc.secretBytes)

				require.NoError(t, err)
				assert.Len(t, secretPart, tc.expectedSecretLength)
				assert.Len(t, fullToken, 5+tc.expectedSecretLength) // "test_" + secret
			})
		}
	})

	t.Run("generates unique tokens", func(t *testing.T) {
		const numTokens = 100
		tokens := make(map[string]bool)

		for i := 0; i < numTokens; i++ {
			fullToken, _, err := generator.GenerateToken("test_", 32)
			require.NoError(t, err)

			assert.False(t, tokens[fullToken], "Token should be unique")
			tokens[fullToken] = true
		}

		assert.Len(t, tokens, numTokens, "All tokens should be unique")
	})

	t.Run("validates entropy requirements", func(t *testing.T) {
		// Generate multiple tokens and verify they have sufficient entropy
		// by checking that they don't have obvious patterns
		tokens := make([]string, 10)

		for i := 0; i < 10; i++ {
			_, secretPart, err := generator.GenerateToken("test_", 32)
			require.NoError(t, err)
			tokens[i] = secretPart
		}

		// Check that tokens don't start with the same characters
		// (this would indicate poor entropy)
		firstChars := make(map[byte]int)
		for _, token := range tokens {
			firstChars[token[0]]++
		}

		// With good entropy, we shouldn't see the same first character too often
		for char, count := range firstChars {
			assert.LessOrEqual(t, count, 5, "First character '%c' appears too frequently (%d times), indicating poor entropy", char, count)
		}
	})

	t.Run("handles invalid input", func(t *testing.T) {
		testCases := []struct {
			name        string
			secretBytes int
			expectError bool
		}{
			{"zero bytes", 0, true},
			{"negative bytes", -1, true},
			{"valid bytes", 32, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, _, err := generator.GenerateToken("test_", tc.secretBytes)

				if tc.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("works with different prefixes", func(t *testing.T) {
		prefixes := []string{"", "pat_", "mcp_pat_", "very_long_prefix_"}

		for _, prefix := range prefixes {
			t.Run("prefix: "+prefix, func(t *testing.T) {
				fullToken, secretPart, err := generator.GenerateToken(prefix, 16)

				require.NoError(t, err)
				assert.True(t, strings.HasPrefix(fullToken, prefix))
				assert.Equal(t, fullToken, prefix+secretPart)
			})
		}
	})
}

func TestSecureTokenGenerator_GeneratePATToken(t *testing.T) {
	generator := NewSecureTokenGenerator()

	t.Run("generates PAT token with correct format", func(t *testing.T) {
		fullToken, secretPart, err := generator.GeneratePATToken()

		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(fullToken, "mcp_pat_"), "PAT token should start with mcp_pat_")
		assert.Equal(t, fullToken, "mcp_pat_"+secretPart)

		// 32 bytes encoded in base64url without padding should be 43 characters
		assert.Len(t, secretPart, 43, "Secret part should be 43 characters for 32 bytes")
	})

	t.Run("generates unique PAT tokens", func(t *testing.T) {
		tokens := make(map[string]bool)

		for i := 0; i < 50; i++ {
			fullToken, _, err := generator.GeneratePATToken()
			require.NoError(t, err)

			assert.False(t, tokens[fullToken], "PAT token should be unique")
			tokens[fullToken] = true
		}
	})
}

// Benchmark tests to ensure token generation performance
func BenchmarkSecureTokenGenerator_GenerateToken(b *testing.B) {
	generator := NewSecureTokenGenerator()

	b.Run("32 bytes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := generator.GenerateToken("mcp_pat_", 32)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("64 bytes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := generator.GenerateToken("mcp_pat_", 64)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkSecureTokenGenerator_GeneratePATToken(b *testing.B) {
	generator := NewSecureTokenGenerator()

	for i := 0; i < b.N; i++ {
		_, _, err := generator.GeneratePATToken()
		if err != nil {
			b.Fatal(err)
		}
	}
}
