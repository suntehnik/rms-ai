package models

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReferenceIDGenerator defines the interface for generating reference IDs
type ReferenceIDGenerator interface {
	Generate(tx *gorm.DB, model interface{}) (string, error)
}

// PostgreSQLReferenceIDGenerator implements reference ID generation for production use
// It uses PostgreSQL advisory locks for atomic generation with UUID fallback
type PostgreSQLReferenceIDGenerator struct {
	lockKey int64  // PostgreSQL advisory lock key
	prefix  string // Entity prefix (EP, US, REQ, AC)
}

// NewPostgreSQLReferenceIDGenerator creates a new PostgreSQL reference ID generator
func NewPostgreSQLReferenceIDGenerator(lockKey int64, prefix string) *PostgreSQLReferenceIDGenerator {
	return &PostgreSQLReferenceIDGenerator{
		lockKey: lockKey,
		prefix:  prefix,
	}
}

// Generate creates a new reference ID using PostgreSQL advisory locks or simple counting for SQLite
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