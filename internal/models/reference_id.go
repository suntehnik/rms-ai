package models

import (
	"fmt"

	"gorm.io/gorm"
)

// ReferenceIDGenerator defines the interface for generating reference IDs.
type ReferenceIDGenerator interface {
	Generate(tx *gorm.DB, model interface{}) (string, error)
}

// PostgreSQLReferenceIDGenerator implements reference ID generation for production use.
//
// This generator calls PostgreSQL functions that use sequences to generate unique reference IDs.
// The sequences are atomic and guarantee uniqueness even under high concurrency.
//
// Reference ID format:
// - EP-001 to EP-999 (with zero-padding)
// - EP-1000, EP-1001, ... (without padding for numbers >= 1000)
//
// This generator is used in production, integration, and e2e test environments.
// For unit tests, use TestReferenceIDGenerator from reference_id_test.go instead.
type PostgreSQLReferenceIDGenerator struct {
	prefix string // Entity prefix (EP, US, REQ, AC, STD, PROMPT)
}

// NewPostgreSQLReferenceIDGenerator creates a new PostgreSQL reference ID generator.
//
// Parameters:
// - lockKey: Legacy parameter (maintained for backward compatibility, not used)
// - prefix: Entity prefix for reference IDs (EP, US, REQ, AC, STD, PROMPT)
func NewPostgreSQLReferenceIDGenerator(lockKey int64, prefix string) *PostgreSQLReferenceIDGenerator {
	return &PostgreSQLReferenceIDGenerator{
		prefix: prefix,
	}
}

// Generate creates a new reference ID by calling the appropriate PostgreSQL function.
//
// The function uses PostgreSQL sequences which are atomic and guarantee uniqueness.
// This method is called from BeforeCreate hooks in GORM models.
//
// Behavior by database type:
// - PostgreSQL: Calls database function (e.g., get_next_epic_ref_id())
// - SQLite: Uses simple counting (for unit tests only)
func (g *PostgreSQLReferenceIDGenerator) Generate(tx *gorm.DB, model interface{}) (string, error) {
	// Check if we're using PostgreSQL
	if tx.Dialector.Name() == "postgres" {
		// Determine which function to call based on prefix
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

		// Call the PostgreSQL function to get the next reference ID
		var referenceID string
		if err := tx.Raw(fmt.Sprintf("SELECT %s()", functionName)).Scan(&referenceID).Error; err != nil {
			return "", fmt.Errorf("failed to generate reference ID: %w", err)
		}

		return referenceID, nil
	}

	// For non-PostgreSQL databases (like SQLite in unit tests), use simple count method
	var count int64
	if err := tx.Model(model).Count(&count).Error; err != nil {
		return "", fmt.Errorf("failed to count records: %w", err)
	}
	return fmt.Sprintf("%s-%03d", g.prefix, count+1), nil
}
