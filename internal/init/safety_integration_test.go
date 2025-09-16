package init

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/models"
)

func TestSafetyChecker_Integration_IsDatabaseEmpty_EmptyDatabase(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Run migrations to create tables but don't add data
	err := models.AutoMigrate(testDB.DB)
	require.NoError(t, err)

	// Create safety checker
	safetyChecker := NewSafetyChecker(testDB.DB)

	// Check if database is empty
	isEmpty, err := safetyChecker.IsDatabaseEmpty()
	assert.NoError(t, err)
	assert.True(t, isEmpty, "Database should be empty after migrations only")
}

func TestSafetyChecker_Integration_IsDatabaseEmpty_WithUsers(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Run migrations and seed data
	err := models.AutoMigrate(testDB.DB)
	require.NoError(t, err)

	// Create test user
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}
	err = testDB.DB.Create(user).Error
	require.NoError(t, err)

	// Create safety checker
	safetyChecker := NewSafetyChecker(testDB.DB)

	// Check if database is empty
	isEmpty, err := safetyChecker.IsDatabaseEmpty()
	assert.NoError(t, err)
	assert.False(t, isEmpty, "Database should not be empty with users")
}

func TestSafetyChecker_Integration_IsDatabaseEmpty_WithEpics(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Run migrations and seed data
	err := models.AutoMigrate(testDB.DB)
	require.NoError(t, err)

	// Seed default data to get requirement types
	err = models.SeedDefaultData(testDB.DB)
	require.NoError(t, err)

	// Create test user first (required for epic)
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}
	err = testDB.DB.Create(user).Error
	require.NoError(t, err)

	// Create test epic
	epic := &models.Epic{
		CreatorID:  user.ID,
		AssigneeID: user.ID,
		Title:      "Test Epic",
		Priority:   models.PriorityHigh,
		Status:     models.EpicStatusBacklog,
	}
	err = testDB.DB.Create(epic).Error
	require.NoError(t, err)

	// Create safety checker
	safetyChecker := NewSafetyChecker(testDB.DB)

	// Check if database is empty
	isEmpty, err := safetyChecker.IsDatabaseEmpty()
	assert.NoError(t, err)
	assert.False(t, isEmpty, "Database should not be empty with epics")
}

func TestSafetyChecker_Integration_GetDataSummary_EmptyDatabase(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Run migrations to create tables
	err := models.AutoMigrate(testDB.DB)
	require.NoError(t, err)

	// Create safety checker
	safetyChecker := NewSafetyChecker(testDB.DB)

	// Get data summary
	summary, err := safetyChecker.GetDataSummary()
	assert.NoError(t, err)
	assert.NotNil(t, summary)

	// Verify all counts are zero
	assert.Equal(t, int64(0), summary.UserCount)
	assert.Equal(t, int64(0), summary.EpicCount)
	assert.Equal(t, int64(0), summary.UserStoryCount)
	assert.Equal(t, int64(0), summary.RequirementCount)
	assert.Equal(t, int64(0), summary.AcceptanceCriteriaCount)
	assert.Equal(t, int64(0), summary.CommentCount)
	assert.True(t, summary.IsEmpty)
	assert.Empty(t, summary.NonEmptyTables)
}

func TestSafetyChecker_Integration_GetDataSummary_WithData(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Run migrations and seed data
	err := models.AutoMigrate(testDB.DB)
	require.NoError(t, err)

	err = models.SeedDefaultData(testDB.DB)
	require.NoError(t, err)

	// Create test users
	users := []*models.User{
		{Username: "user1", Email: "user1@example.com", PasswordHash: "hash1", Role: models.RoleUser},
		{Username: "user2", Email: "user2@example.com", PasswordHash: "hash2", Role: models.RoleCommenter},
		{Username: "user3", Email: "user3@example.com", PasswordHash: "hash3", Role: models.RoleAdministrator},
	}
	for _, user := range users {
		err = testDB.DB.Create(user).Error
		require.NoError(t, err)
	}

	// Create test epics
	epics := []*models.Epic{
		{CreatorID: users[0].ID, AssigneeID: users[0].ID, Title: "Epic 1", Priority: models.PriorityHigh, Status: models.EpicStatusBacklog},
		{CreatorID: users[1].ID, AssigneeID: users[1].ID, Title: "Epic 2", Priority: models.PriorityMedium, Status: models.EpicStatusInProgress},
	}
	for _, epic := range epics {
		err = testDB.DB.Create(epic).Error
		require.NoError(t, err)
	}

	// Create test user stories
	userStories := []*models.UserStory{
		{EpicID: epics[0].ID, CreatorID: users[0].ID, AssigneeID: users[0].ID, Title: "User Story 1", Priority: models.PriorityHigh, Status: models.UserStoryStatusBacklog},
		{EpicID: epics[1].ID, CreatorID: users[1].ID, AssigneeID: users[1].ID, Title: "User Story 2", Priority: models.PriorityMedium, Status: models.UserStoryStatusInProgress},
		{EpicID: epics[0].ID, CreatorID: users[2].ID, AssigneeID: users[2].ID, Title: "User Story 3", Priority: models.PriorityLow, Status: models.UserStoryStatusDone},
	}
	for _, userStory := range userStories {
		err = testDB.DB.Create(userStory).Error
		require.NoError(t, err)
	}

	// Get requirement type for requirements
	var reqType models.RequirementType
	err = testDB.DB.Where("name = ?", "Functional").First(&reqType).Error
	require.NoError(t, err)

	// Create test requirements
	requirements := []*models.Requirement{
		{UserStoryID: userStories[0].ID, CreatorID: users[0].ID, AssigneeID: users[0].ID, TypeID: reqType.ID, Title: "Requirement 1", Priority: models.PriorityHigh, Status: models.RequirementStatusDraft},
		{UserStoryID: userStories[1].ID, CreatorID: users[1].ID, AssigneeID: users[1].ID, TypeID: reqType.ID, Title: "Requirement 2", Priority: models.PriorityMedium, Status: models.RequirementStatusActive},
	}
	for _, requirement := range requirements {
		err = testDB.DB.Create(requirement).Error
		require.NoError(t, err)
	}

	// Create test acceptance criteria
	acceptanceCriteria := []*models.AcceptanceCriteria{
		{UserStoryID: userStories[0].ID, AuthorID: users[0].ID, Description: "WHEN user clicks button THEN system SHALL respond"},
		{UserStoryID: userStories[1].ID, AuthorID: users[1].ID, Description: "WHEN data is valid THEN system SHALL save record"},
		{UserStoryID: userStories[2].ID, AuthorID: users[2].ID, Description: "WHEN error occurs THEN system SHALL display message"},
	}
	for _, ac := range acceptanceCriteria {
		err = testDB.DB.Create(ac).Error
		require.NoError(t, err)
	}

	// Create test comments
	comments := []*models.Comment{
		{AuthorID: users[0].ID, EntityType: "epic", EntityID: epics[0].ID, Content: "Comment 1"},
		{AuthorID: users[1].ID, EntityType: "user_story", EntityID: userStories[0].ID, Content: "Comment 2"},
		{AuthorID: users[2].ID, EntityType: "requirement", EntityID: requirements[0].ID, Content: "Comment 3"},
		{AuthorID: users[0].ID, EntityType: "acceptance_criteria", EntityID: acceptanceCriteria[0].ID, Content: "Comment 4"},
	}
	for _, comment := range comments {
		err = testDB.DB.Create(comment).Error
		require.NoError(t, err)
	}

	// Create safety checker
	safetyChecker := NewSafetyChecker(testDB.DB)

	// Get data summary
	summary, err := safetyChecker.GetDataSummary()
	assert.NoError(t, err)
	assert.NotNil(t, summary)

	// Verify counts
	assert.Equal(t, int64(3), summary.UserCount)
	assert.Equal(t, int64(2), summary.EpicCount)
	assert.Equal(t, int64(3), summary.UserStoryCount)
	assert.Equal(t, int64(2), summary.RequirementCount)
	assert.Equal(t, int64(3), summary.AcceptanceCriteriaCount)
	assert.Equal(t, int64(4), summary.CommentCount)
	assert.False(t, summary.IsEmpty)

	// Verify non-empty tables
	expectedTables := []string{"users", "epics", "user_stories", "requirements", "acceptance_criteria", "comments"}
	for _, table := range expectedTables {
		assert.Contains(t, summary.NonEmptyTables, table, "Table %s should be in non-empty tables list", table)
	}
}

func TestSafetyChecker_Integration_GetNonEmptyTablesReport_EmptyDatabase(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Run migrations to create tables
	err := models.AutoMigrate(testDB.DB)
	require.NoError(t, err)

	// Create safety checker
	safetyChecker := NewSafetyChecker(testDB.DB)

	// Get report
	report, err := safetyChecker.GetNonEmptyTablesReport()
	assert.NoError(t, err)
	assert.Contains(t, report, "Database is empty and safe for initialization")
}

func TestSafetyChecker_Integration_GetNonEmptyTablesReport_WithData(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Run migrations and seed data
	err := models.AutoMigrate(testDB.DB)
	require.NoError(t, err)

	// Create test users
	users := []*models.User{
		{Username: "user1", Email: "user1@example.com", PasswordHash: "hash1", Role: models.RoleUser},
		{Username: "user2", Email: "user2@example.com", PasswordHash: "hash2", Role: models.RoleCommenter},
	}
	for _, user := range users {
		err = testDB.DB.Create(user).Error
		require.NoError(t, err)
	}

	// Create test epics
	epics := []*models.Epic{
		{CreatorID: users[0].ID, AssigneeID: users[0].ID, Title: "Epic 1", Priority: models.PriorityHigh, Status: models.EpicStatusBacklog},
	}
	for _, epic := range epics {
		err = testDB.DB.Create(epic).Error
		require.NoError(t, err)
	}

	// Create safety checker
	safetyChecker := NewSafetyChecker(testDB.DB)

	// Get report
	report, err := safetyChecker.GetNonEmptyTablesReport()
	assert.NoError(t, err)
	assert.Contains(t, report, "Database contains existing data")
	assert.Contains(t, report, "users: 2 records")
	assert.Contains(t, report, "epics: 1 records")
	assert.Contains(t, report, "Initialization cannot proceed")
}

func TestSafetyChecker_Integration_ValidateEmptyDatabase_EmptyDatabase(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Run migrations to create tables
	err := models.AutoMigrate(testDB.DB)
	require.NoError(t, err)

	// Create safety checker
	safetyChecker := NewSafetyChecker(testDB.DB)

	// Validate empty database
	err = safetyChecker.ValidateEmptyDatabase()
	assert.NoError(t, err, "Empty database should pass validation")
}

func TestSafetyChecker_Integration_ValidateEmptyDatabase_NonEmptyDatabase(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Run migrations and create test data
	err := models.AutoMigrate(testDB.DB)
	require.NoError(t, err)

	// Create test user
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}
	err = testDB.DB.Create(user).Error
	require.NoError(t, err)

	// Create safety checker
	safetyChecker := NewSafetyChecker(testDB.DB)

	// Validate empty database - should fail
	err = safetyChecker.ValidateEmptyDatabase()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database safety check failed")
	assert.Contains(t, err.Error(), "users: 1 records")
}

func TestSafetyChecker_PartialData_OnlyComments(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Run migrations and seed data
	err := models.AutoMigrate(testDB.DB)
	require.NoError(t, err)

	err = models.SeedDefaultData(testDB.DB)
	require.NoError(t, err)

	// Create minimal data structure for comment
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}
	err = testDB.DB.Create(user).Error
	require.NoError(t, err)

	epic := &models.Epic{
		CreatorID:  user.ID,
		AssigneeID: user.ID,
		Title:      "Test Epic",
		Priority:   models.PriorityHigh,
		Status:     models.EpicStatusBacklog,
	}
	err = testDB.DB.Create(epic).Error
	require.NoError(t, err)

	// Create only comment (to test partial data detection)
	comment := &models.Comment{
		AuthorID:   user.ID,
		EntityType: "epic",
		EntityID:   epic.ID,
		Content:    "Test comment",
	}
	err = testDB.DB.Create(comment).Error
	require.NoError(t, err)

	// Create safety checker
	safetyChecker := NewSafetyChecker(testDB.DB)

	// Get data summary
	summary, err := safetyChecker.GetDataSummary()
	assert.NoError(t, err)
	assert.False(t, summary.IsEmpty)
	assert.Equal(t, int64(1), summary.UserCount)
	assert.Equal(t, int64(1), summary.EpicCount)
	assert.Equal(t, int64(1), summary.CommentCount)
	assert.Contains(t, summary.NonEmptyTables, "users")
	assert.Contains(t, summary.NonEmptyTables, "epics")
	assert.Contains(t, summary.NonEmptyTables, "comments")
}

func TestSafetyChecker_ErrorHandling_DatabaseError(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Create safety checker
	safetyChecker := NewSafetyChecker(testDB.DB)

	// Close database connection to simulate error
	sqlDB, err := testDB.DB.DB()
	require.NoError(t, err)
	sqlDB.Close()

	// Attempt to check database - should fail
	isEmpty, err := safetyChecker.IsDatabaseEmpty()
	assert.Error(t, err)
	assert.False(t, isEmpty)

	// Attempt to get summary - should fail
	summary, err := safetyChecker.GetDataSummary()
	assert.Error(t, err)
	assert.Nil(t, summary)

	// Attempt to get report - should fail
	report, err := safetyChecker.GetNonEmptyTablesReport()
	assert.Error(t, err)
	assert.Empty(t, report)
}

func TestSafetyChecker_LargeDataset_Performance(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Run migrations and seed data
	err := models.AutoMigrate(testDB.DB)
	require.NoError(t, err)

	err = models.SeedDefaultData(testDB.DB)
	require.NoError(t, err)

	// Create many users to test performance with larger dataset
	users := make([]*models.User, 100)
	for i := 0; i < 100; i++ {
		users[i] = &models.User{
			Username:     "user" + string(rune(i)),
			Email:        "user" + string(rune(i)) + "@example.com",
			PasswordHash: "hashed_password",
			Role:         models.RoleUser,
		}
	}

	// Batch create users
	err = testDB.DB.CreateInBatches(users, 10).Error
	require.NoError(t, err)

	// Create safety checker
	safetyChecker := NewSafetyChecker(testDB.DB)

	// Test performance of safety check with larger dataset
	isEmpty, err := safetyChecker.IsDatabaseEmpty()
	assert.NoError(t, err)
	assert.False(t, isEmpty)

	// Get data summary
	summary, err := safetyChecker.GetDataSummary()
	assert.NoError(t, err)
	assert.Equal(t, int64(100), summary.UserCount)
	assert.False(t, summary.IsEmpty)
	assert.Contains(t, summary.NonEmptyTables, "users")
}
