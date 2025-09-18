package init

import (
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create tables that match our models
	err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE epics (
			id TEXT PRIMARY KEY,
			reference_id TEXT UNIQUE NOT NULL,
			creator_id TEXT NOT NULL,
			assignee_id TEXT NOT NULL,
			created_at DATETIME,
			last_modified DATETIME,
			priority INTEGER NOT NULL,
			status TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE user_stories (
			id TEXT PRIMARY KEY,
			reference_id TEXT UNIQUE NOT NULL,
			epic_id TEXT NOT NULL,
			creator_id TEXT NOT NULL,
			assignee_id TEXT NOT NULL,
			created_at DATETIME,
			last_modified DATETIME,
			priority INTEGER NOT NULL,
			status TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE requirements (
			id TEXT PRIMARY KEY,
			reference_id TEXT UNIQUE NOT NULL,
			user_story_id TEXT NOT NULL,
			acceptance_criteria_id TEXT,
			creator_id TEXT NOT NULL,
			assignee_id TEXT NOT NULL,
			created_at DATETIME,
			last_modified DATETIME,
			priority INTEGER NOT NULL,
			status TEXT NOT NULL,
			type_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE acceptance_criteria (
			id TEXT PRIMARY KEY,
			reference_id TEXT UNIQUE NOT NULL,
			user_story_id TEXT NOT NULL,
			author_id TEXT NOT NULL,
			created_at DATETIME,
			last_modified DATETIME,
			description TEXT NOT NULL
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE comments (
			id TEXT PRIMARY KEY,
			entity_type TEXT NOT NULL,
			entity_id TEXT NOT NULL,
			parent_comment_id TEXT,
			author_id TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			content TEXT NOT NULL,
			is_resolved BOOLEAN,
			linked_text TEXT,
			text_position_start INTEGER,
			text_position_end INTEGER
		)
	`).Error
	require.NoError(t, err)

	return db
}

func TestNewSafetyChecker(t *testing.T) {
	db := setupTestDB(t)

	checker := NewSafetyChecker(db)

	assert.NotNil(t, checker)
	assert.Equal(t, db, checker.db)
}

func TestSafetyChecker_IsDatabaseEmpty_EmptyDatabase(t *testing.T) {
	db := setupTestDB(t)
	checker := NewSafetyChecker(db)

	isEmpty, err := checker.IsDatabaseEmpty()

	assert.NoError(t, err)
	assert.True(t, isEmpty)
}

func TestSafetyChecker_IsDatabaseEmpty_WithUsers(t *testing.T) {
	db := setupTestDB(t)
	checker := NewSafetyChecker(db)

	// Insert a user
	err := db.Exec(`
		INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at)
		VALUES ('user-1', 'testuser', 'test@example.com', 'hashedpassword', 'User', datetime('now'), datetime('now'))
	`).Error
	require.NoError(t, err)

	isEmpty, err := checker.IsDatabaseEmpty()

	assert.NoError(t, err)
	assert.False(t, isEmpty)
}

func TestSafetyChecker_IsDatabaseEmpty_WithEpics(t *testing.T) {
	db := setupTestDB(t)
	checker := NewSafetyChecker(db)

	// Insert an epic
	err := db.Exec(`
		INSERT INTO epics (id, reference_id, creator_id, assignee_id, created_at, last_modified, priority, status, title)
		VALUES ('epic-1', 'EP-001', 'user-1', 'user-1', datetime('now'), datetime('now'), 1, 'Backlog', 'Test Epic')
	`).Error
	require.NoError(t, err)

	isEmpty, err := checker.IsDatabaseEmpty()

	assert.NoError(t, err)
	assert.False(t, isEmpty)
}

func TestSafetyChecker_IsDatabaseEmpty_WithUserStories(t *testing.T) {
	db := setupTestDB(t)
	checker := NewSafetyChecker(db)

	// Insert a user story
	err := db.Exec(`
		INSERT INTO user_stories (id, reference_id, epic_id, creator_id, assignee_id, created_at, last_modified, priority, status, title)
		VALUES ('us-1', 'US-001', 'epic-1', 'user-1', 'user-1', datetime('now'), datetime('now'), 1, 'Backlog', 'Test User Story')
	`).Error
	require.NoError(t, err)

	isEmpty, err := checker.IsDatabaseEmpty()

	assert.NoError(t, err)
	assert.False(t, isEmpty)
}

func TestSafetyChecker_IsDatabaseEmpty_WithRequirements(t *testing.T) {
	db := setupTestDB(t)
	checker := NewSafetyChecker(db)

	// Insert a requirement
	err := db.Exec(`
		INSERT INTO requirements (id, reference_id, user_story_id, creator_id, assignee_id, created_at, last_modified, priority, status, type_id, title)
		VALUES ('req-1', 'REQ-001', 'us-1', 'user-1', 'user-1', datetime('now'), datetime('now'), 1, 'Draft', 'type-1', 'Test Requirement')
	`).Error
	require.NoError(t, err)

	isEmpty, err := checker.IsDatabaseEmpty()

	assert.NoError(t, err)
	assert.False(t, isEmpty)
}

func TestSafetyChecker_IsDatabaseEmpty_WithAcceptanceCriteria(t *testing.T) {
	db := setupTestDB(t)
	checker := NewSafetyChecker(db)

	// Insert acceptance criteria
	err := db.Exec(`
		INSERT INTO acceptance_criteria (id, reference_id, user_story_id, author_id, created_at, last_modified, description)
		VALUES ('ac-1', 'AC-001', 'us-1', 'user-1', datetime('now'), datetime('now'), 'WHEN user logs in THEN system SHALL authenticate')
	`).Error
	require.NoError(t, err)

	isEmpty, err := checker.IsDatabaseEmpty()

	assert.NoError(t, err)
	assert.False(t, isEmpty)
}

func TestSafetyChecker_IsDatabaseEmpty_WithComments(t *testing.T) {
	db := setupTestDB(t)
	checker := NewSafetyChecker(db)

	// Insert a comment
	err := db.Exec(`
		INSERT INTO comments (id, entity_type, entity_id, author_id, created_at, updated_at, content, is_resolved)
		VALUES ('comment-1', 'epic', 'epic-1', 'user-1', datetime('now'), datetime('now'), 'Test comment', false)
	`).Error
	require.NoError(t, err)

	isEmpty, err := checker.IsDatabaseEmpty()

	assert.NoError(t, err)
	assert.False(t, isEmpty)
}

func TestSafetyChecker_GetDataSummary_EmptyDatabase(t *testing.T) {
	db := setupTestDB(t)
	checker := NewSafetyChecker(db)

	summary, err := checker.GetDataSummary()

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, int64(0), summary.UserCount)
	assert.Equal(t, int64(0), summary.EpicCount)
	assert.Equal(t, int64(0), summary.UserStoryCount)
	assert.Equal(t, int64(0), summary.RequirementCount)
	assert.Equal(t, int64(0), summary.AcceptanceCriteriaCount)
	assert.Equal(t, int64(0), summary.CommentCount)
	assert.True(t, summary.IsEmpty)
	assert.Empty(t, summary.NonEmptyTables)
}

func TestSafetyChecker_GetDataSummary_WithData(t *testing.T) {
	db := setupTestDB(t)
	checker := NewSafetyChecker(db)

	// Insert test data
	err := db.Exec(`
		INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at)
		VALUES ('user-1', 'testuser', 'test@example.com', 'hashedpassword', 'User', datetime('now'), datetime('now'))
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		INSERT INTO epics (id, reference_id, creator_id, assignee_id, created_at, last_modified, priority, status, title)
		VALUES 
		('epic-1', 'EP-001', 'user-1', 'user-1', datetime('now'), datetime('now'), 1, 'Backlog', 'Test Epic 1'),
		('epic-2', 'EP-002', 'user-1', 'user-1', datetime('now'), datetime('now'), 2, 'Draft', 'Test Epic 2')
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		INSERT INTO comments (id, entity_type, entity_id, author_id, created_at, updated_at, content, is_resolved)
		VALUES ('comment-1', 'epic', 'epic-1', 'user-1', datetime('now'), datetime('now'), 'Test comment', false)
	`).Error
	require.NoError(t, err)

	summary, err := checker.GetDataSummary()

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, int64(1), summary.UserCount)
	assert.Equal(t, int64(2), summary.EpicCount)
	assert.Equal(t, int64(0), summary.UserStoryCount)
	assert.Equal(t, int64(0), summary.RequirementCount)
	assert.Equal(t, int64(0), summary.AcceptanceCriteriaCount)
	assert.Equal(t, int64(1), summary.CommentCount)
	assert.False(t, summary.IsEmpty)
	assert.Contains(t, summary.NonEmptyTables, "users")
	assert.Contains(t, summary.NonEmptyTables, "epics")
	assert.Contains(t, summary.NonEmptyTables, "comments")
	assert.Len(t, summary.NonEmptyTables, 3)
}

func TestSafetyChecker_GetNonEmptyTablesReport_EmptyDatabase(t *testing.T) {
	db := setupTestDB(t)
	checker := NewSafetyChecker(db)

	report, err := checker.GetNonEmptyTablesReport()

	assert.NoError(t, err)
	assert.Equal(t, "Database is empty and safe for initialization", report)
}

func TestSafetyChecker_GetNonEmptyTablesReport_WithData(t *testing.T) {
	db := setupTestDB(t)
	checker := NewSafetyChecker(db)

	// Insert test data
	err := db.Exec(`
		INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at)
		VALUES ('user-1', 'testuser', 'test@example.com', 'hashedpassword', 'User', datetime('now'), datetime('now'))
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		INSERT INTO epics (id, reference_id, creator_id, assignee_id, created_at, last_modified, priority, status, title)
		VALUES ('epic-1', 'EP-001', 'user-1', 'user-1', datetime('now'), datetime('now'), 1, 'Backlog', 'Test Epic')
	`).Error
	require.NoError(t, err)

	report, err := checker.GetNonEmptyTablesReport()

	assert.NoError(t, err)
	assert.Contains(t, report, "Database contains existing data in the following tables:")
	assert.Contains(t, report, "users: 1 records")
	assert.Contains(t, report, "epics: 1 records")
	assert.Contains(t, report, "Initialization cannot proceed on a non-empty database to prevent data corruption.")
}

func TestSafetyChecker_ValidateEmptyDatabase_EmptyDatabase(t *testing.T) {
	db := setupTestDB(t)
	checker := NewSafetyChecker(db)

	err := checker.ValidateEmptyDatabase()

	assert.NoError(t, err)
}

func TestSafetyChecker_ValidateEmptyDatabase_NonEmptyDatabase(t *testing.T) {
	db := setupTestDB(t)
	checker := NewSafetyChecker(db)

	// Insert test data
	err := db.Exec(`
		INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at)
		VALUES ('user-1', 'testuser', 'test@example.com', 'hashedpassword', 'User', datetime('now'), datetime('now'))
	`).Error
	require.NoError(t, err)

	err = checker.ValidateEmptyDatabase()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database safety check failed")
	assert.Contains(t, err.Error(), "users: 1 records")
}

func TestSafetyChecker_DatabaseConnectionError(t *testing.T) {
	// Create a database connection that will fail
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Close the database to simulate connection error
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()

	checker := NewSafetyChecker(db)

	// Test IsDatabaseEmpty with connection error
	isEmpty, err := checker.IsDatabaseEmpty()
	assert.Error(t, err)
	assert.False(t, isEmpty)
	assert.Contains(t, err.Error(), "failed to get data summary")

	// Test GetDataSummary with connection error
	summary, err := checker.GetDataSummary()
	assert.Error(t, err)
	assert.Nil(t, summary)

	// Test ValidateEmptyDatabase with connection error
	err = checker.ValidateEmptyDatabase()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check database emptiness")
}

func TestSafetyChecker_AllTablesWithData(t *testing.T) {
	db := setupTestDB(t)
	checker := NewSafetyChecker(db)

	// Insert data in all tables
	err := db.Exec(`
		INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at)
		VALUES ('user-1', 'testuser', 'test@example.com', 'hashedpassword', 'User', datetime('now'), datetime('now'))
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		INSERT INTO epics (id, reference_id, creator_id, assignee_id, created_at, last_modified, priority, status, title)
		VALUES ('epic-1', 'EP-001', 'user-1', 'user-1', datetime('now'), datetime('now'), 1, 'Backlog', 'Test Epic')
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		INSERT INTO user_stories (id, reference_id, epic_id, creator_id, assignee_id, created_at, last_modified, priority, status, title)
		VALUES ('us-1', 'US-001', 'epic-1', 'user-1', 'user-1', datetime('now'), datetime('now'), 1, 'Backlog', 'Test User Story')
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		INSERT INTO requirements (id, reference_id, user_story_id, creator_id, assignee_id, created_at, last_modified, priority, status, type_id, title)
		VALUES ('req-1', 'REQ-001', 'us-1', 'user-1', 'user-1', datetime('now'), datetime('now'), 1, 'Draft', 'type-1', 'Test Requirement')
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		INSERT INTO acceptance_criteria (id, reference_id, user_story_id, author_id, created_at, last_modified, description)
		VALUES ('ac-1', 'AC-001', 'us-1', 'user-1', datetime('now'), datetime('now'), 'WHEN user logs in THEN system SHALL authenticate')
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		INSERT INTO comments (id, entity_type, entity_id, author_id, created_at, updated_at, content, is_resolved)
		VALUES ('comment-1', 'epic', 'epic-1', 'user-1', datetime('now'), datetime('now'), 'Test comment', false)
	`).Error
	require.NoError(t, err)

	summary, err := checker.GetDataSummary()

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, int64(1), summary.UserCount)
	assert.Equal(t, int64(1), summary.EpicCount)
	assert.Equal(t, int64(1), summary.UserStoryCount)
	assert.Equal(t, int64(1), summary.RequirementCount)
	assert.Equal(t, int64(1), summary.AcceptanceCriteriaCount)
	assert.Equal(t, int64(1), summary.CommentCount)
	assert.False(t, summary.IsEmpty)
	assert.Len(t, summary.NonEmptyTables, 6)

	// Test the report contains all tables
	report, err := checker.GetNonEmptyTablesReport()
	assert.NoError(t, err)
	assert.Contains(t, report, "users: 1 records")
	assert.Contains(t, report, "epics: 1 records")
	assert.Contains(t, report, "user_stories: 1 records")
	assert.Contains(t, report, "requirements: 1 records")
	assert.Contains(t, report, "acceptance_criteria: 1 records")
	assert.Contains(t, report, "comments: 1 records")
}

func TestSafetyChecker_MultipleRecordsInTables(t *testing.T) {
	db := setupTestDB(t)
	checker := NewSafetyChecker(db)

	// Insert multiple users
	err := db.Exec(`
		INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at)
		VALUES 
		('user-1', 'testuser1', 'test1@example.com', 'hashedpassword', 'User', datetime('now'), datetime('now')),
		('user-2', 'testuser2', 'test2@example.com', 'hashedpassword', 'Admin', datetime('now'), datetime('now')),
		('user-3', 'testuser3', 'test3@example.com', 'hashedpassword', 'User', datetime('now'), datetime('now'))
	`).Error
	require.NoError(t, err)

	// Insert multiple epics
	err = db.Exec(`
		INSERT INTO epics (id, reference_id, creator_id, assignee_id, created_at, last_modified, priority, status, title)
		VALUES 
		('epic-1', 'EP-001', 'user-1', 'user-1', datetime('now'), datetime('now'), 1, 'Backlog', 'Test Epic 1'),
		('epic-2', 'EP-002', 'user-2', 'user-2', datetime('now'), datetime('now'), 2, 'Draft', 'Test Epic 2')
	`).Error
	require.NoError(t, err)

	summary, err := checker.GetDataSummary()

	assert.NoError(t, err)
	assert.Equal(t, int64(3), summary.UserCount)
	assert.Equal(t, int64(2), summary.EpicCount)
	assert.False(t, summary.IsEmpty)

	report, err := checker.GetNonEmptyTablesReport()
	assert.NoError(t, err)
	assert.Contains(t, report, "users: 3 records")
	assert.Contains(t, report, "epics: 2 records")
}

// Tests for new helper methods and enhanced error handling

func TestIsTableNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "PostgreSQL undefined_table error",
			err:      &pgconn.PgError{Code: "42P01", Message: "relation \"nonexistent_table\" does not exist"},
			expected: true,
		},
		{
			name:     "PostgreSQL other error",
			err:      &pgconn.PgError{Code: "23505", Message: "duplicate key value violates unique constraint"},
			expected: false,
		},
		{
			name:     "generic does not exist error",
			err:      fmt.Errorf("table \"users\" does not exist"),
			expected: true,
		},
		{
			name:     "SQLite no such table error",
			err:      fmt.Errorf("no such table: users"),
			expected: true,
		},
		{
			name:     "undefined_table in message",
			err:      fmt.Errorf("database error: undefined_table"),
			expected: true,
		},
		{
			name:     "case insensitive matching",
			err:      fmt.Errorf("Table DOES NOT EXIST"),
			expected: true,
		},
		{
			name:     "unrelated error",
			err:      fmt.Errorf("connection refused"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTableNotFoundError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSafetyChecker_countTableRecords_MissingTable(t *testing.T) {
	// Create a database without creating all tables
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Only create users table, leave others missing
	err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	checker := NewSafetyChecker(db)

	// Test counting records in existing table
	count, err := checker.countTableRecords("users")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Test counting records in missing table - should return 0, not error
	count, err = checker.countTableRecords("nonexistent_table")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestSafetyChecker_countTableRecords_WithData(t *testing.T) {
	db := setupTestDB(t)
	checker := NewSafetyChecker(db)

	// Insert test data
	err := db.Exec(`
		INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at)
		VALUES 
		('user-1', 'testuser1', 'test1@example.com', 'hashedpassword', 'User', datetime('now'), datetime('now')),
		('user-2', 'testuser2', 'test2@example.com', 'hashedpassword', 'Admin', datetime('now'), datetime('now'))
	`).Error
	require.NoError(t, err)

	count, err := checker.countTableRecords("users")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestSafetyChecker_GetDataSummary_MixedScenario(t *testing.T) {
	// Create a database with only some tables
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create only users and epics tables, leave others missing
	err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE epics (
			id TEXT PRIMARY KEY,
			reference_id TEXT UNIQUE NOT NULL,
			creator_id TEXT NOT NULL,
			assignee_id TEXT NOT NULL,
			created_at DATETIME,
			last_modified DATETIME,
			priority INTEGER NOT NULL,
			status TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT
		)
	`).Error
	require.NoError(t, err)

	// Insert data only in users table
	err = db.Exec(`
		INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at)
		VALUES ('user-1', 'testuser', 'test@example.com', 'hashedpassword', 'User', datetime('now'), datetime('now'))
	`).Error
	require.NoError(t, err)

	checker := NewSafetyChecker(db)
	summary, err := checker.GetDataSummary()

	assert.NoError(t, err)
	assert.NotNil(t, summary)

	// Should count existing tables correctly
	assert.Equal(t, int64(1), summary.UserCount)
	assert.Equal(t, int64(0), summary.EpicCount)

	// Should treat missing tables as empty (0 count)
	assert.Equal(t, int64(0), summary.UserStoryCount)
	assert.Equal(t, int64(0), summary.RequirementCount)
	assert.Equal(t, int64(0), summary.AcceptanceCriteriaCount)
	assert.Equal(t, int64(0), summary.CommentCount)

	// Database is not empty because users table has data
	assert.False(t, summary.IsEmpty)
	assert.Contains(t, summary.NonEmptyTables, "users")
	assert.Len(t, summary.NonEmptyTables, 1)
}

func TestSafetyChecker_GetDataSummary_AllTablesMissing(t *testing.T) {
	// Create a completely empty database (no tables at all)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	checker := NewSafetyChecker(db)
	summary, err := checker.GetDataSummary()

	assert.NoError(t, err)
	assert.NotNil(t, summary)

	// All counts should be zero
	assert.Equal(t, int64(0), summary.UserCount)
	assert.Equal(t, int64(0), summary.EpicCount)
	assert.Equal(t, int64(0), summary.UserStoryCount)
	assert.Equal(t, int64(0), summary.RequirementCount)
	assert.Equal(t, int64(0), summary.AcceptanceCriteriaCount)
	assert.Equal(t, int64(0), summary.CommentCount)

	// Database should be considered empty
	assert.True(t, summary.IsEmpty)
	assert.Empty(t, summary.NonEmptyTables)
}

func TestSafetyChecker_countTableRecords_DatabaseError(t *testing.T) {
	// Create a database connection that will fail
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Close the database to simulate connection error
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()

	checker := NewSafetyChecker(db)

	// Should propagate database errors (not table not found errors)
	count, err := checker.countTableRecords("users")
	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
	assert.NotContains(t, err.Error(), "does not exist") // Should not be treated as table not found
}

func TestSafetyChecker_GetDataSummary_PropagatesNonTableErrors(t *testing.T) {
	// Create a database connection that will fail
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Close the database to simulate connection error
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()

	checker := NewSafetyChecker(db)

	// Should propagate database connection errors
	summary, err := checker.GetDataSummary()
	assert.Error(t, err)
	assert.Nil(t, summary)
	assert.Contains(t, err.Error(), "failed to check table")
}

// Benchmark tests for performance validation
func BenchmarkSafetyChecker_IsDatabaseEmpty(b *testing.B) {
	db := setupTestDB(&testing.T{})
	checker := NewSafetyChecker(db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = checker.IsDatabaseEmpty()
	}
}

func BenchmarkSafetyChecker_GetDataSummary(b *testing.B) {
	db := setupTestDB(&testing.T{})
	checker := NewSafetyChecker(db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = checker.GetDataSummary()
	}
}

func BenchmarkSafetyChecker_countTableRecords(b *testing.B) {
	db := setupTestDB(&testing.T{})
	checker := NewSafetyChecker(db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = checker.countTableRecords("users")
	}
}
