package init

import (
	"fmt"

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

	// Check users table
	if err := sc.db.Table("users").Count(&summary.UserCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}
	if summary.UserCount > 0 {
		summary.NonEmptyTables = append(summary.NonEmptyTables, "users")
	}

	// Check epics table
	if err := sc.db.Table("epics").Count(&summary.EpicCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count epics: %w", err)
	}
	if summary.EpicCount > 0 {
		summary.NonEmptyTables = append(summary.NonEmptyTables, "epics")
	}

	// Check user_stories table
	if err := sc.db.Table("user_stories").Count(&summary.UserStoryCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count user stories: %w", err)
	}
	if summary.UserStoryCount > 0 {
		summary.NonEmptyTables = append(summary.NonEmptyTables, "user_stories")
	}

	// Check requirements table
	if err := sc.db.Table("requirements").Count(&summary.RequirementCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count requirements: %w", err)
	}
	if summary.RequirementCount > 0 {
		summary.NonEmptyTables = append(summary.NonEmptyTables, "requirements")
	}

	// Check acceptance_criteria table
	if err := sc.db.Table("acceptance_criteria").Count(&summary.AcceptanceCriteriaCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count acceptance criteria: %w", err)
	}
	if summary.AcceptanceCriteriaCount > 0 {
		summary.NonEmptyTables = append(summary.NonEmptyTables, "acceptance_criteria")
	}

	// Check comments table
	if err := sc.db.Table("comments").Count(&summary.CommentCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count comments: %w", err)
	}
	if summary.CommentCount > 0 {
		summary.NonEmptyTables = append(summary.NonEmptyTables, "comments")
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
