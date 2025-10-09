package service

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestNewBcryptHashService(t *testing.T) {
	t.Run("creates service with valid cost", func(t *testing.T) {
		validCosts := []int{bcrypt.MinCost, bcrypt.DefaultCost, 12, bcrypt.MaxCost}

		for _, cost := range validCosts {
			service, err := NewBcryptHashService(cost)
			require.NoError(t, err)
			assert.Equal(t, cost, service.cost)
		}
	})

	t.Run("rejects invalid cost", func(t *testing.T) {
		invalidCosts := []int{bcrypt.MinCost - 1, bcrypt.MaxCost + 1, -1, 0, 100}

		for _, cost := range invalidCosts {
			service, err := NewBcryptHashService(cost)
			assert.Error(t, err)
			assert.Nil(t, service)
			assert.Contains(t, err.Error(), "invalid bcrypt cost")
		}
	})
}

func TestNewDefaultBcryptHashService(t *testing.T) {
	service := NewDefaultBcryptHashService()
	assert.Equal(t, bcrypt.DefaultCost, service.cost)
}

func TestBcryptHashService_HashToken(t *testing.T) {
	service := NewDefaultBcryptHashService()

	t.Run("hashes token successfully", func(t *testing.T) {
		token := "test_secret_token_12345"

		hash, err := service.HashToken(token)

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.True(t, strings.HasPrefix(hash, "$2a$"), "Hash should start with bcrypt identifier")
		assert.NotEqual(t, token, hash, "Hash should not equal original token")
	})

	t.Run("generates different hashes for same token", func(t *testing.T) {
		token := "same_token"

		hash1, err1 := service.HashToken(token)
		hash2, err2 := service.HashToken(token)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2, "Different hashes should be generated due to salt")
	})

	t.Run("rejects empty token", func(t *testing.T) {
		hash, err := service.HashToken("")

		assert.Error(t, err)
		assert.Empty(t, hash)
		assert.Contains(t, err.Error(), "token cannot be empty")
	})

	t.Run("handles various token lengths", func(t *testing.T) {
		testTokens := []string{
			"a",                                    // very short
			"short_token",                          // short
			"medium_length_token_with_numbers_123", // medium
			strings.Repeat("long_", 10),            // long but within bcrypt 72-byte limit (50 chars)
		}

		for _, token := range testTokens {
			hash, err := service.HashToken(token)
			require.NoError(t, err, "Should hash token of length %d", len(token))
			assert.NotEmpty(t, hash)
		}
	})

	t.Run("handles bcrypt length limit", func(t *testing.T) {
		// bcrypt has a 72-byte limit, test tokens at and beyond this limit
		token72 := strings.Repeat("a", 72) // exactly 72 bytes
		token73 := strings.Repeat("a", 73) // exceeds limit

		// 72 bytes should work
		hash, err := service.HashToken(token72)
		require.NoError(t, err, "Should hash token of exactly 72 bytes")
		assert.NotEmpty(t, hash)

		// 73 bytes should fail
		hash, err = service.HashToken(token73)
		assert.Error(t, err, "Should reject token exceeding 72 bytes")
		assert.Empty(t, hash)
		assert.Contains(t, err.Error(), "failed to hash token")
	})
}

func TestBcryptHashService_CompareTokenWithHash(t *testing.T) {
	service := NewDefaultBcryptHashService()

	t.Run("validates correct token", func(t *testing.T) {
		token := "correct_token_12345"
		hash, err := service.HashToken(token)
		require.NoError(t, err)

		err = service.CompareTokenWithHash(token, hash)
		assert.NoError(t, err, "Correct token should validate successfully")
	})

	t.Run("rejects incorrect token", func(t *testing.T) {
		correctToken := "correct_token"
		incorrectToken := "incorrect_token"

		hash, err := service.HashToken(correctToken)
		require.NoError(t, err)

		err = service.CompareTokenWithHash(incorrectToken, hash)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token does not match hash")
	})

	t.Run("rejects empty inputs", func(t *testing.T) {
		validToken := "valid_token"
		validHash, err := service.HashToken(validToken)
		require.NoError(t, err)

		// Empty token
		err = service.CompareTokenWithHash("", validHash)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token cannot be empty")

		// Empty hash
		err = service.CompareTokenWithHash(validToken, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "hash cannot be empty")
	})

	t.Run("rejects malformed hash", func(t *testing.T) {
		token := "test_token"
		malformedHashes := []string{
			"not_a_hash",
			"$2a$10$invalid",
			"plain_text_password",
		}

		for _, hash := range malformedHashes {
			err := service.CompareTokenWithHash(token, hash)
			assert.Error(t, err, "Should reject malformed hash: %s", hash)
		}
	})

	t.Run("is resistant to timing attacks", func(t *testing.T) {
		// This test verifies that comparison time doesn't vary significantly
		// based on how "close" the wrong token is to the correct one
		correctToken := "correct_token_with_specific_pattern"
		hash, err := service.HashToken(correctToken)
		require.NoError(t, err)

		wrongTokens := []string{
			"a",                                  // very different
			"correct",                            // partial match
			"correct_token",                      // closer match
			"correct_token_with_specific_patter", // very close
			"correct_token_with_specific_pattern_extra", // extra chars
		}

		// Measure comparison times
		times := make([]time.Duration, len(wrongTokens))
		for i, wrongToken := range wrongTokens {
			start := time.Now()
			service.CompareTokenWithHash(wrongToken, hash)
			times[i] = time.Since(start)
		}

		// All comparisons should take roughly the same time (within an order of magnitude)
		// This is a basic check - bcrypt's constant-time comparison should handle this
		minTime := times[0]
		maxTime := times[0]
		for _, t := range times[1:] {
			if t < minTime {
				minTime = t
			}
			if t > maxTime {
				maxTime = t
			}
		}

		// Allow up to 10x difference (bcrypt should be much more consistent than this)
		ratio := float64(maxTime) / float64(minTime)
		assert.Less(t, ratio, 10.0, "Comparison times should not vary significantly")
	})
}

func TestBcryptHashService_ValidateToken(t *testing.T) {
	service := NewDefaultBcryptHashService()
	prefix := "mcp_pat_"
	secretPart := "K7gNU3sdo-OL0wNhqoVWhr3g6s1xYv72ol_pe_Unols"
	fullToken := prefix + secretPart

	t.Run("validates correct full token", func(t *testing.T) {
		hash, err := service.HashToken(secretPart)
		require.NoError(t, err)

		err = service.ValidateToken(fullToken, prefix, hash)
		assert.NoError(t, err, "Valid full token should validate successfully")
	})

	t.Run("rejects token with wrong prefix", func(t *testing.T) {
		wrongPrefix := "wrong_prefix_"
		hash, err := service.HashToken(secretPart)
		require.NoError(t, err)

		err = service.ValidateToken(fullToken, wrongPrefix, hash)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid prefix")
	})

	t.Run("rejects token with wrong secret", func(t *testing.T) {
		wrongSecretPart := "wrong_secret_part"
		wrongFullToken := prefix + wrongSecretPart
		hash, err := service.HashToken(secretPart) // Hash of correct secret
		require.NoError(t, err)

		err = service.ValidateToken(wrongFullToken, prefix, hash)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token does not match hash")
	})

	t.Run("validates input parameters", func(t *testing.T) {
		hash, err := service.HashToken(secretPart)
		require.NoError(t, err)

		testCases := []struct {
			name          string
			token         string
			prefix        string
			hash          string
			expectedError string
		}{
			{"empty token", "", prefix, hash, "token cannot be empty"},
			{"empty prefix", fullToken, "", hash, "expected prefix cannot be empty"},
			{"token too short", "abc", prefix, hash, "token is too short"},
			{"token shorter than prefix", "mcp", prefix, hash, "token is too short"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := service.ValidateToken(tc.token, tc.prefix, tc.hash)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			})
		}
	})
}

func TestBcryptHashService_SecurityProperties(t *testing.T) {
	t.Run("different costs produce different hashes", func(t *testing.T) {
		token := "test_token_for_cost_comparison"

		lowCostService, err := NewBcryptHashService(bcrypt.MinCost)
		require.NoError(t, err)

		highCostService, err := NewBcryptHashService(12)
		require.NoError(t, err)

		lowHash, err := lowCostService.HashToken(token)
		require.NoError(t, err)

		highHash, err := highCostService.HashToken(token)
		require.NoError(t, err)

		assert.NotEqual(t, lowHash, highHash, "Different costs should produce different hashes")

		// Both should validate the same token
		assert.NoError(t, lowCostService.CompareTokenWithHash(token, lowHash))
		assert.NoError(t, highCostService.CompareTokenWithHash(token, highHash))
	})

	t.Run("hash contains cost information", func(t *testing.T) {
		token := "test_token"

		testCosts := []int{bcrypt.MinCost, bcrypt.DefaultCost, 12}
		for _, cost := range testCosts {
			service, err := NewBcryptHashService(cost)
			require.NoError(t, err)

			hash, err := service.HashToken(token)
			require.NoError(t, err)

			// bcrypt hash format: $2a$cost$salt+hash
			parts := strings.Split(hash, "$")
			require.Len(t, parts, 4, "Hash should have 4 parts")
			assert.Equal(t, "2a", parts[1], "Should use bcrypt 2a variant")

			// The cost should be encoded in the hash
			expectedCostStr := ""
			if cost < 10 {
				expectedCostStr = "0" + string(rune('0'+cost))
			} else {
				expectedCostStr = string(rune('0'+cost/10)) + string(rune('0'+cost%10))
			}
			assert.Equal(t, expectedCostStr, parts[2], "Cost should be encoded in hash")
		}
	})
}

// Benchmark tests to ensure hashing performance is acceptable
func BenchmarkBcryptHashService_HashToken(b *testing.B) {
	service := NewDefaultBcryptHashService()
	token := "benchmark_token_with_reasonable_length"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.HashToken(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBcryptHashService_CompareTokenWithHash(b *testing.B) {
	service := NewDefaultBcryptHashService()
	token := "benchmark_token_for_comparison"
	hash, err := service.HashToken(token)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.CompareTokenWithHash(token, hash)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBcryptHashService_ValidateToken(b *testing.B) {
	service := NewDefaultBcryptHashService()
	prefix := "mcp_pat_"
	secretPart := "benchmark_secret_part"
	fullToken := prefix + secretPart
	hash, err := service.HashToken(secretPart)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.ValidateToken(fullToken, prefix, hash)
		if err != nil {
			b.Fatal(err)
		}
	}
}
