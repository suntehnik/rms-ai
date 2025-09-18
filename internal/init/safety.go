package init

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// DataSummary contains information about existing data in the database
type DataSummary struct {
	UserCount               int64    `json:"user_count"`
	EpicCount               int64    `json:"epic_count"`
	UserStoryCount          int64    `json:"user_story_count"`
	RequirementCount        int64    `json:"requirement_count"`
	AcceptanceCriteriaCount int64    `json:"acceptance_criteria_count"`
	CommentCount            int64    `json:"comment_count"`
	IsEmpty                 bool     `json:"is_empty"`
	NonEmptyTables          []string `json:"non_empty_tables"`
}

// SafetyChecker provides functionality to check if the database is safe for initialization
type SafetyChecker struct {
	db *gorm.DB
}

// NewSafetyChecker creates a new SafetyChecker instance
func NewSafetyChecker(db *gorm.DB) *SafetyChecker {
	return &SafetyChecker{
		db: db,
	}
}

// IsDatabaseEmpty checks if the database contains any existing data in critical tables
// Returns true if the database is empty and safe for initialization
func (sc *SafetyChecker) IsDatabaseEmpty() (bool, error) {
	summary, err := sc.GetDataSummary()
	if err != nil {
		return false, fmt.Errorf("failed to get data summary: %w", err)
	}

	return summary.IsEmpty, nil
}

// GetDataSummary returns detailed information about existing data in the database
func (sc *SafetyChecker) GetDataSummary() (*DataSummary, error) {
	summary := &DataSummary{
		NonEmptyTables: make([]string, 0),
	}

	// Define tables to check with their corresponding count fields
	tablesToCheck := []struct {
		name     string
		countPtr *int64
		label    string
	}{
		{"users", &summary.UserCount, "users"},
		{"epics", &summary.EpicCount, "epics"},
		{"user_stories", &summary.UserStoryCount, "user_stories"},
		{"requirements", &summary.RequirementCount, "requirements"},
		{"acceptance_criteria", &summary.AcceptanceCriteriaCount, "acceptance_criteria"},
		{"comments", &summary.CommentCount, "comments"},
	}

	// Check each table with enhanced error handling
	for _, table := range tablesToCheck {
		count, err := sc.countTableRecords(table.name)
		if err != nil {
			return nil, fmt.Errorf("failed to check table %s: %w", table.name, err)
		}

		*table.countPtr = count
		if count > 0 {
			summary.NonEmptyTables = append(summary.NonEmptyTables, table.label)
		}
	}

	// Database is empty if all critical tables are empty
	summary.IsEmpty = len(summary.NonEmptyTables) == 0

	return summary, nil
}

// GetNonEmptyTablesReport returns a formatted report of non-empty tables
func (sc *SafetyChecker) GetNonEmptyTablesReport() (string, error) {
	summary, err := sc.GetDataSummary()
	if err != nil {
		return "", fmt.Errorf("failed to get data summary: %w", err)
	}

	if summary.IsEmpty {
		return "Database is empty and safe for initialization", nil
	}

	report := "Database contains existing data in the following tables:\n"

	if summary.UserCount > 0 {
		report += fmt.Sprintf("  - users: %d records\n", summary.UserCount)
	}
	if summary.EpicCount > 0 {
		report += fmt.Sprintf("  - epics: %d records\n", summary.EpicCount)
	}
	if summary.UserStoryCount > 0 {
		report += fmt.Sprintf("  - user_stories: %d records\n", summary.UserStoryCount)
	}
	if summary.RequirementCount > 0 {
		report += fmt.Sprintf("  - requirements: %d records\n", summary.RequirementCount)
	}
	if summary.AcceptanceCriteriaCount > 0 {
		report += fmt.Sprintf("  - acceptance_criteria: %d records\n", summary.AcceptanceCriteriaCount)
	}
	if summary.CommentCount > 0 {
		report += fmt.Sprintf("  - comments: %d records\n", summary.CommentCount)
	}

	report += "\nInitialization cannot proceed on a non-empty database to prevent data corruption."

	return report, nil
}

// ValidateEmptyDatabase performs the safety check and returns an error if the database is not empty
func (sc *SafetyChecker) ValidateEmptyDatabase() error {
	isEmpty, err := sc.IsDatabaseEmpty()
	if err != nil {
		return fmt.Errorf("failed to check database emptiness: %w", err)
	}

	if !isEmpty {
		report, reportErr := sc.GetNonEmptyTablesReport()
		if reportErr != nil {
			return fmt.Errorf("database is not empty and failed to generate report: %w", reportErr)
		}
		return fmt.Errorf("database safety check failed:\n%s", report)
	}

	return nil
}

// countTableRecords safely counts records in a table, handling missing tables gracefully
// Returns zero count for missing tables and propagates other database errors
func (sc *SafetyChecker) countTableRecords(tableName string) (int64, error) {
	var count int64
	err := sc.db.Table(tableName).Count(&count).Error

	if err != nil {
		if isTableNotFoundError(err) {
			// Table doesn't exist - treat as empty (0 records)
			return 0, nil
		}
		// Other database errors should be propagated
		return 0, err
	}

	return count, nil
}

// isTableNotFoundError checks if the given error indicates that a table was not found
// It handles PostgreSQL SQLSTATE 42P01 ("undefined_table") errors and provides
// fallback string matching for generic "table does not exist" messages
func isTableNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Check for PostgreSQL "undefined_table" error (SQLSTATE 42P01)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "42P01"
	}

	// Fallback: check error message for common patterns
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "does not exist") ||
		strings.Contains(errMsg, "no such table") ||
		strings.Contains(errMsg, "undefined_table")
}
