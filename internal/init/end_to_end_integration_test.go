package init

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/auth"
	"product-requirements-management/internal/database"
	"product-requirements-management/internal/models"
)

func TestEndToEnd_CompleteInitializationWorkflow(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Set required environment variables
	originalPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	defer func() {
		if originalPassword != "" {
			os.Setenv("DEFAULT_ADMIN_PASSWORD", originalPassword)
		} else {
			os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
		}
	}()
	testPassword := "secure-admin-password-e2e-test"
	os.Setenv("DEFAULT_ADMIN_PASSWORD", testPassword)

	// Create initialization service
	service, err := NewInitService(testDB.Config)
	require.NoError(t, err)
	defer service.Close()

	// Override database connection to use test database and initialize components
	service.db = testDB.DB
	service.safetyChecker = NewSafetyChecker(service.db)
	service.migrator = database.NewMigrationManager(service.db, "migrations")
	service.adminCreator = NewAdminCreator(service.db, service.auth)

	// Record start time
	startTime := time.Now()

	// Run complete initialization
	err = service.Initialize()
	assert.NoError(t, err)

	// Record end time
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// Verify initialization completed in reasonable time (should be fast for empty DB)
	assert.Less(t, duration, 30*time.Second, "Initialization should complete within 30 seconds")

	// === VERIFY DATABASE SCHEMA ===

	// Verify schema_migrations table exists and has records
	var migrationCount int64
	err = testDB.DB.Table("schema_migrations").Count(&migrationCount).Error
	assert.NoError(t, err)
	assert.Greater(t, migrationCount, int64(0), "Should have migration records")

	// Verify all expected tables exist
	expectedTables := []string{
		"users", "epics", "user_stories", "requirements",
		"acceptance_criteria", "comments", "requirement_types",
		"relationship_types", "requirement_relationships",
	}

	for _, table := range expectedTables {
		var exists bool
		err = testDB.DB.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = ?)", table).Scan(&exists).Error
		assert.NoError(t, err)
		assert.True(t, exists, "Table %s should exist", table)
	}

	// === VERIFY REFERENCE DATA ===

	// Verify requirement types were seeded
	var reqTypeCount int64
	err = testDB.DB.Table("requirement_types").Count(&reqTypeCount).Error
	assert.NoError(t, err)
	assert.Greater(t, reqTypeCount, int64(0), "Should have requirement types")

	// Verify relationship types were seeded
	var relTypeCount int64
	err = testDB.DB.Table("relationship_types").Count(&relTypeCount).Error
	assert.NoError(t, err)
	assert.Greater(t, relTypeCount, int64(0), "Should have relationship types")

	// Verify specific requirement types exist
	expectedReqTypes := []string{"Functional", "Non-Functional", "Business", "Technical"}
	for _, reqType := range expectedReqTypes {
		var count int64
		err = testDB.DB.Table("requirement_types").Where("name = ?", reqType).Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count, "Requirement type %s should exist", reqType)
	}

	// Verify specific relationship types exist
	expectedRelTypes := []string{"depends_on", "blocks", "relates_to", "conflicts_with", "derives_from"}
	for _, relType := range expectedRelTypes {
		var count int64
		err = testDB.DB.Table("relationship_types").Where("name = ?", relType).Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count, "Relationship type %s should exist", relType)
	}

	// === VERIFY ADMIN USER ===

	// Verify admin user was created
	var adminUser models.User
	err = testDB.DB.Where("username = ?", "admin").First(&adminUser).Error
	assert.NoError(t, err)

	// Verify admin user properties
	assert.Equal(t, "admin", adminUser.Username)
	assert.Equal(t, "admin@localhost", adminUser.Email)
	assert.Equal(t, models.RoleAdministrator, adminUser.Role)
	assert.NotEmpty(t, adminUser.PasswordHash)
	assert.NotEqual(t, testPassword, adminUser.PasswordHash, "Password should be hashed")

	// Verify admin user is the only user
	var totalUserCount int64
	err = testDB.DB.Table("users").Count(&totalUserCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), totalUserCount, "Should have exactly one user")

	// === VERIFY PASSWORD AUTHENTICATION ===

	// Create auth service to verify password
	authService := auth.NewService(testDB.Config.JWT.Secret, 24*time.Hour)

	// Verify password can be authenticated
	err = authService.VerifyPassword(testPassword, adminUser.PasswordHash)
	assert.NoError(t, err, "Admin password should be verifiable")

	// Verify wrong password fails
	err = authService.VerifyPassword("wrong-password", adminUser.PasswordHash)
	assert.Error(t, err, "Wrong password should not be verifiable")

	// === VERIFY DATABASE CONSTRAINTS ===

	// Test that we can create entities using the admin user

	// Create an epic
	epicDescription := "Test epic created during E2E test"
	epic := &models.Epic{
		CreatorID:   adminUser.ID,
		AssigneeID:  adminUser.ID,
		Title:       "Test Epic",
		Description: &epicDescription,
		Priority:    models.PriorityHigh,
		Status:      models.EpicStatusBacklog,
	}
	err = testDB.DB.Create(epic).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, epic.ReferenceID, "Epic should have reference ID")
	assert.Contains(t, epic.ReferenceID, "EP-", "Epic reference ID should have EP- prefix")

	// Create a user story
	userStoryDescription := "Test user story created during E2E test"
	userStory := &models.UserStory{
		EpicID:      epic.ID,
		CreatorID:   adminUser.ID,
		AssigneeID:  adminUser.ID,
		Title:       "Test User Story",
		Description: &userStoryDescription,
		Priority:    models.PriorityMedium,
		Status:      models.UserStoryStatusBacklog,
	}
	err = testDB.DB.Create(userStory).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, userStory.ReferenceID, "User story should have reference ID")
	assert.Contains(t, userStory.ReferenceID, "US-", "User story reference ID should have US- prefix")

	// Get requirement type for requirement creation
	var reqType models.RequirementType
	err = testDB.DB.Where("name = ?", "Functional").First(&reqType).Error
	assert.NoError(t, err)

	// Create a requirement
	requirementDescription := "Test requirement created during E2E test"
	requirement := &models.Requirement{
		UserStoryID: userStory.ID,
		CreatorID:   adminUser.ID,
		AssigneeID:  adminUser.ID,
		TypeID:      reqType.ID,
		Title:       "Test Requirement",
		Description: &requirementDescription,
		Priority:    models.PriorityHigh,
		Status:      models.RequirementStatusDraft,
	}
	err = testDB.DB.Create(requirement).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, requirement.ReferenceID, "Requirement should have reference ID")
	assert.Contains(t, requirement.ReferenceID, "REQ-", "Requirement reference ID should have REQ- prefix")

	// Create acceptance criteria
	acceptanceCriteria := &models.AcceptanceCriteria{
		UserStoryID: userStory.ID,
		AuthorID:    adminUser.ID,
		Description: "WHEN user submits valid data THEN system SHALL save the record",
	}
	err = testDB.DB.Create(acceptanceCriteria).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, acceptanceCriteria.ReferenceID, "Acceptance criteria should have reference ID")
	assert.Contains(t, acceptanceCriteria.ReferenceID, "AC-", "AC reference ID should have AC- prefix")

	// Create a comment
	comment := &models.Comment{
		AuthorID:   adminUser.ID,
		EntityType: "requirement",
		EntityID:   requirement.ID,
		Content:    "Test comment created during E2E test",
	}
	err = testDB.DB.Create(comment).Error
	assert.NoError(t, err)

	// === VERIFY RELATIONSHIPS ===

	// Verify epic-user story relationship
	var epicWithStories models.Epic
	err = testDB.DB.Preload("UserStories").Where("id = ?", epic.ID).First(&epicWithStories).Error
	assert.NoError(t, err)
	assert.Len(t, epicWithStories.UserStories, 1, "Epic should have one user story")
	assert.Equal(t, userStory.ID, epicWithStories.UserStories[0].ID)

	// Verify user story-requirement relationship
	var storyWithRequirements models.UserStory
	err = testDB.DB.Preload("Requirements").Where("id = ?", userStory.ID).First(&storyWithRequirements).Error
	assert.NoError(t, err)
	assert.Len(t, storyWithRequirements.Requirements, 1, "User story should have one requirement")
	assert.Equal(t, requirement.ID, storyWithRequirements.Requirements[0].ID)

	// Verify user story-acceptance criteria relationship
	var storyWithAC models.UserStory
	err = testDB.DB.Preload("AcceptanceCriteria").Where("id = ?", userStory.ID).First(&storyWithAC).Error
	assert.NoError(t, err)
	assert.Len(t, storyWithAC.AcceptanceCriteria, 1, "User story should have one acceptance criteria")
	assert.Equal(t, acceptanceCriteria.ID, storyWithAC.AcceptanceCriteria[0].ID)

	// === VERIFY FULL-TEXT SEARCH CAPABILITIES ===

	// Test search functionality (basic check that search indexes work)
	var searchResults []models.Epic
	err = testDB.DB.Where("title ILIKE ?", "%Test%").Find(&searchResults).Error
	assert.NoError(t, err)
	assert.Len(t, searchResults, 1, "Should find the test epic")
	assert.Equal(t, epic.ID, searchResults[0].ID)
}

func TestEndToEnd_InitializationFailure_Recovery(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Create some existing data to trigger safety check failure
	err := testDB.runSQLMigrations()
	require.NoError(t, err)

	existingUser := &models.User{
		Username:     "existing_user",
		Email:        "existing@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}
	err = testDB.DB.Create(existingUser).Error
	require.NoError(t, err)

	// Set required environment variable
	originalPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	defer func() {
		if originalPassword != "" {
			os.Setenv("DEFAULT_ADMIN_PASSWORD", originalPassword)
		} else {
			os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
		}
	}()
	os.Setenv("DEFAULT_ADMIN_PASSWORD", "test-admin-password-123")

	// Create initialization service
	service, err := NewInitService(testDB.Config)
	require.NoError(t, err)
	defer service.Close()

	// Override database connection to use test database and initialize components
	service.db = testDB.DB
	service.safetyChecker = NewSafetyChecker(service.db)
	service.migrator = database.NewMigrationManager(service.db, "migrations")
	service.adminCreator = NewAdminCreator(service.db, service.auth)

	// Attempt initialization - should fail due to existing data
	err = service.Initialize()
	assert.Error(t, err)

	// Verify error type and message
	var initErr *InitError
	assert.ErrorAs(t, err, &initErr)
	assert.Equal(t, ErrorTypeSafety, initErr.Type)
	assert.Contains(t, initErr.Message, "Database safety check failed")

	// Verify no admin user was created (only existing user remains)
	var userCount int64
	err = testDB.DB.Table("users").Count(&userCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), userCount, "Should still have only the existing user")

	// Verify no admin user exists
	var adminUser models.User
	err = testDB.DB.Where("username = ?", "admin").First(&adminUser).Error
	assert.Error(t, err, "Admin user should not exist")

	// === RECOVERY TEST ===

	// Clear the database to simulate recovery
	err = testDB.reset()
	require.NoError(t, err)

	// Now initialization should succeed
	err = service.Initialize()
	assert.NoError(t, err)

	// Verify admin user was created after recovery
	err = testDB.DB.Where("username = ?", "admin").First(&adminUser).Error
	assert.NoError(t, err)
	assert.Equal(t, "admin", adminUser.Username)
	assert.Equal(t, models.RoleAdministrator, adminUser.Role)
}

func TestEndToEnd_MultipleInitializationAttempts(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Set required environment variable
	originalPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	defer func() {
		if originalPassword != "" {
			os.Setenv("DEFAULT_ADMIN_PASSWORD", originalPassword)
		} else {
			os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
		}
	}()
	os.Setenv("DEFAULT_ADMIN_PASSWORD", "test-admin-password-123")

	// First initialization - should succeed
	service1, err := NewInitService(testDB.Config)
	require.NoError(t, err)
	defer service1.Close()
	service1.db = testDB.DB
	service1.safetyChecker = NewSafetyChecker(service1.db)
	service1.migrator = database.NewMigrationManager(service1.db, "migrations")
	service1.adminCreator = NewAdminCreator(service1.db, service1.auth)

	err = service1.Initialize()
	assert.NoError(t, err)

	// Verify admin user was created
	var adminUser models.User
	err = testDB.DB.Where("username = ?", "admin").First(&adminUser).Error
	assert.NoError(t, err)
	firstAdminID := adminUser.ID

	// Second initialization attempt - should fail due to existing data
	service2, err := NewInitService(testDB.Config)
	require.NoError(t, err)
	defer service2.Close()
	service2.db = testDB.DB
	service2.safetyChecker = NewSafetyChecker(service2.db)
	service2.migrator = database.NewMigrationManager(service2.db, "migrations")
	service2.adminCreator = NewAdminCreator(service2.db, service2.auth)

	err = service2.Initialize()
	assert.Error(t, err)

	// Verify error is safety check failure
	var initErr *InitError
	assert.ErrorAs(t, err, &initErr)
	assert.Equal(t, ErrorTypeSafety, initErr.Type)

	// Verify original admin user is unchanged
	err = testDB.DB.Where("username = ?", "admin").First(&adminUser).Error
	assert.NoError(t, err)
	assert.Equal(t, firstAdminID, adminUser.ID, "Admin user should be unchanged")

	// Verify only one admin user exists
	var adminCount int64
	err = testDB.DB.Where("username = ?", "admin").Count(&adminCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), adminCount, "Should have exactly one admin user")
}

func TestEndToEnd_InitializationWithCustomConfiguration(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Customize configuration
	testDB.Config.JWT.Secret = "custom-jwt-secret-for-e2e-test"
	testDB.Config.Log.Level = "debug"
	testDB.Config.Log.Format = "text"

	// Set required environment variable
	originalPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	defer func() {
		if originalPassword != "" {
			os.Setenv("DEFAULT_ADMIN_PASSWORD", originalPassword)
		} else {
			os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
		}
	}()
	customPassword := "custom-secure-password-e2e"
	os.Setenv("DEFAULT_ADMIN_PASSWORD", customPassword)

	// Create initialization service with custom config
	service, err := NewInitService(testDB.Config)
	require.NoError(t, err)
	defer service.Close()

	// Override database connection to use test database and initialize components
	service.db = testDB.DB
	service.safetyChecker = NewSafetyChecker(service.db)
	service.migrator = database.NewMigrationManager(service.db, "migrations")
	service.adminCreator = NewAdminCreator(service.db, service.auth)

	// Run initialization
	err = service.Initialize()
	assert.NoError(t, err)

	// Verify admin user was created with custom password
	var adminUser models.User
	err = testDB.DB.Where("username = ?", "admin").First(&adminUser).Error
	assert.NoError(t, err)

	// Verify password works with custom JWT secret
	authService := auth.NewService(testDB.Config.JWT.Secret, 24*time.Hour)
	err = authService.VerifyPassword(customPassword, adminUser.PasswordHash)
	assert.NoError(t, err, "Custom password should be verifiable")

	// Verify JWT token generation works
	token, err := authService.GenerateToken(&adminUser)
	assert.NoError(t, err)
	assert.NotEmpty(t, token, "Should be able to generate JWT token")

	// Verify token validation works
	claims, err := authService.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, adminUser.ID.String(), claims.UserID)
	assert.Equal(t, adminUser.Username, claims.Username)
	assert.Equal(t, adminUser.Role, claims.Role)
}

func TestEndToEnd_InitializationPerformance_Benchmarking(t *testing.T) {
	// Set up test database
	testDB := setupTestDatabase(t)
	defer testDB.cleanup(t)

	// Set required environment variable
	originalPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	defer func() {
		if originalPassword != "" {
			os.Setenv("DEFAULT_ADMIN_PASSWORD", originalPassword)
		} else {
			os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
		}
	}()
	os.Setenv("DEFAULT_ADMIN_PASSWORD", "test-admin-password-123")

	// Create initialization service
	service, err := NewInitService(testDB.Config)
	require.NoError(t, err)
	defer service.Close()

	// Override database connection to use test database and initialize components
	service.db = testDB.DB
	service.safetyChecker = NewSafetyChecker(service.db)
	service.migrator = database.NewMigrationManager(service.db, "migrations")
	service.adminCreator = NewAdminCreator(service.db, service.auth)

	// Measure initialization time
	startTime := time.Now()
	err = service.Initialize()
	duration := time.Since(startTime)

	assert.NoError(t, err)

	// Performance assertions
	assert.Less(t, duration, 10*time.Second, "Initialization should complete within 10 seconds")

	// Log performance metrics for monitoring
	t.Logf("Initialization completed in: %v", duration)
	t.Logf("Performance benchmark: %v", duration.Milliseconds())

	// Verify initialization was complete and correct
	var adminUser models.User
	err = testDB.DB.Where("username = ?", "admin").First(&adminUser).Error
	assert.NoError(t, err)

	var migrationCount int64
	err = testDB.DB.Table("schema_migrations").Count(&migrationCount).Error
	assert.NoError(t, err)
	assert.Greater(t, migrationCount, int64(0))

	t.Logf("Migrations applied: %d", migrationCount)
	t.Logf("Admin user created: %s (ID: %s)", adminUser.Username, adminUser.ID)
}
