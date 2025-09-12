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
// - PostgreSQL advisory locks provide atomic reference ID generation
// - UUID fallback ensures uniqueness when locks are unavailable
// - Database-specific optimizations are beneficial
//
// Key features:
// - Uses PostgreSQL advisory locks for atomic generation
// - Falls back to UUID-based IDs when lock acquisition fails
// - Automatically detects PostgreSQL vs SQLite and adapts behavior
// - Thread-safe for concurrent operations
// - Maintains sequential numbering when possible
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
// - lockKey: Unique PostgreSQL advisory lock key for this entity type
// - prefix: Entity prefix for reference IDs (EP, US, REQ, AC)
//
// Lock key assignments:
// - Epic: 2147483647 (existing, maintained for backward compatibility)
// - UserStory: 2147483646 (existing, maintained for backward compatibility)
// - Requirement: 2147483645 (newly assigned)
// - AcceptanceCriteria: 2147483644 (newly assigned)
func NewPostgreSQLReferenceIDGenerator(lockKey int64, prefix string) *PostgreSQLReferenceIDGenerator {
	return &PostgreSQLReferenceIDGenerator{
		lockKey: lockKey,
		prefix:  prefix,
	}
}

// Generate creates a new reference ID using PostgreSQL advisory locks or simple counting for SQLite.
//
// Behavior by database type:
// - PostgreSQL: Uses advisory locks for atomic generation, falls back to UUID on lock failure
// - SQLite: Uses simple counting (acceptable for integration tests with controlled concurrency)
//
// Reference ID formats:
// - Sequential: "PREFIX-001", "PREFIX-002", etc. (when lock acquired or using SQLite)
// - UUID fallback: "PREFIX-a1b2c3d4" (when PostgreSQL lock acquisition fails)
//
// This method is thread-safe and handles concurrent operations properly in production environments.
func (g *PostgreSQLReferenceIDGenerator) Generate(tx *gorm.DB, model interface{}) (string, error) {
	// Check if we're using PostgreSQL for advisory locks
	if tx.Dialector.Name() == "postgres" {
		// Use PostgreSQL advisory lock for atomic reference ID generation
		var lockAcquired bool
		if err := tx.Raw("SELECT pg_try_advisory_xact_lock(?)", g.lockKey).Scan(&lockAcquired).Error; err != nil {
			return "", fmt.Errorf("failed to acquire advisory lock: %w", err)
		}

		if !lockAcquired {
			// If lock not acquired, fall back to UUID-based ID
			return fmt.Sprintf("%s-%s", g.prefix, uuid.New().String()[:8]), nil
		}

		// Lock acquired, safely generate sequential reference ID
		var count int64
		if err := tx.Model(model).Count(&count).Error; err != nil {
			return "", fmt.Errorf("failed to count records: %w", err)
		}
		return fmt.Sprintf("%s-%03d", g.prefix, count+1), nil
	}

	// For non-PostgreSQL databases (like SQLite in tests), use simple count method
	var count int64
	if err := tx.Model(model).Count(&count).Error; err != nil {
		return "", fmt.Errorf("failed to count records: %w", err)
	}
	return fmt.Sprintf("%s-%03d", g.prefix, count+1), nil
}
