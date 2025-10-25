package models

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReferenceIDGenerator defines the interface for generating reference IDs.
// This interface is implemented by both production and test generators to ensure
// consistent behavior across different environments.
type ReferenceIDGenerator interface {
	Generate(tx *gorm.DB, model interface{}) (string, error)
}

// PostgreSQLReferenceIDGenerator implements reference ID generation for production use.
//
// This generator is designed for production, integration, and e2e test environments where:
// - Thread-safety and concurrency handling are critical
// - PostgreSQL sequences provide atomic reference ID generation
// - UUID fallback ensures uniqueness when sequence functions fail
// - Database-specific optimizations are beneficial
//
// Key features:
// - Uses PostgreSQL sequence functions for atomic generation
// - Falls back to UUID-based IDs when sequence functions fail
// - Automatically detects PostgreSQL vs SQLite and adapts behavior
// - Thread-safe for concurrent operations
// - Maintains sequential numbering using database sequences
//
// This generator is statically selected at compile time for all non-unit-test scenarios.
// For unit tests, use TestReferenceIDGenerator from reference_id_test.go instead.
type PostgreSQLReferenceIDGenerator struct {
	lockKey int64  // PostgreSQL advisory lock key (unique per entity type)
	prefix  string // Entity prefix (EP, US, REQ, AC)
}

// NewPostgreSQLReferenceIDGenerator creates a new PostgreSQL reference ID generator.
//
// Parameters:
// - lockKey: Legacy parameter (maintained for backward compatibility, not used in sequence-based approach)
// - prefix: Entity prefix for reference IDs (EP, US, REQ, AC)
//
// Supported prefixes and their corresponding PostgreSQL functions:
// - EP: get_next_epic_ref_id()
// - US: get_next_user_story_ref_id()
// - REQ: get_next_requirement_ref_id()
// - AC: get_next_acceptance_criteria_ref_id()
// - STD: get_next_steering_document_ref_id()
func NewPostgreSQLReferenceIDGenerator(lockKey int64, prefix string) *PostgreSQLReferenceIDGenerator {
	return &PostgreSQLReferenceIDGenerator{
		lockKey: lockKey,
		prefix:  prefix,
	}
}

// Generate creates a new reference ID using PostgreSQL sequences or simple counting for SQLite.
//
// Behavior by database type:
// - PostgreSQL: Uses database sequences via helper functions for atomic generation
// - SQLite: Uses simple counting (acceptable for integration tests with controlled concurrency)
//
// Reference ID formats:
// - Sequential: "PREFIX-001", "PREFIX-002", etc. (using PostgreSQL sequences or SQLite counting)
// - UUID fallback: "PREFIX-a1b2c3d4" (when PostgreSQL function call fails)
//
// This method is thread-safe and handles concurrent operations properly in production environments.
func (g *PostgreSQLReferenceIDGenerator) Generate(tx *gorm.DB, model interface{}) (string, error) {
	// Check if we're using PostgreSQL for sequence-based generation
	if tx.Dialector.Name() == "postgres" {
		// Use PostgreSQL sequence functions for atomic reference ID generation
		var functionName string
		switch g.prefix {
		case "EP":
			functionName = "get_next_epic_ref_id"
		case "US":
			functionName = "get_next_user_story_ref_id"
		case "AC":
			functionName = "get_next_acceptance_criteria_ref_id"
		case "REQ":
			functionName = "get_next_requirement_ref_id"
		case "STD":
			functionName = "get_next_steering_document_ref_id"
		case "PROMPT":
			functionName = "get_next_prompt_ref_id"
		default:
			return "", fmt.Errorf("unknown prefix: %s", g.prefix)
		}

		var referenceID string
		if err := tx.Raw(fmt.Sprintf("SELECT %s()", functionName)).Scan(&referenceID).Error; err != nil {
			// If sequence function fails, fall back to UUID-based ID
			return fmt.Sprintf("%s-%s", g.prefix, uuid.New().String()[:8]), nil
		}
		return referenceID, nil
	}

	// For non-PostgreSQL databases (like SQLite in tests), use simple count method
	var count int64
	if err := tx.Model(model).Count(&count).Error; err != nil {
		return "", fmt.Errorf("failed to count records: %w", err)
	}
	return fmt.Sprintf("%s-%03d", g.prefix, count+1), nil
}
