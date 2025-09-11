package models

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestAcceptanceCriteriaReferenceIDProductionGenerator tests the production reference ID generator
// with PostgreSQL using testcontainers for concurrent operations and uniqueness
func TestAcceptanceCriteriaReferenceIDProductionGenerator(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupPostgreSQLForAcceptanceCriteriaReferenceIDTest(t)
	defer cleanupPostgreSQLAcceptanceCriteriaTest(t, db)

	// Auto-migrate required models
	err := db.AutoMigrate(
		&User{},
		&Epic{},
		&UserStory{},
		&AcceptanceCriteria{},
	)
	require.NoError(t, err)

	// Create test data
	testUser, testUserStory := createTestDataForAcceptanceCriteria(t, db)

	t.Run("TestSequentialReferenceIDGeneration", func(t *testing.T) {
		testAcceptanceCriteriaSequentialReferenceIDGeneration(t, db, testUser, testUserStory)
	})

	t.Run("TestConcurrentReferenceIDGeneration", func(t *testing.T) {
		testAcceptanceCriteriaConcurrentReferenceIDGeneration(t, db, testUser, testUserStory)
	})

	t.Run("TestReferenceIDFormatAndUniqueness", func(t *testing.T) {
		testAcceptanceCriteriaReferenceIDFormatAndUniqueness(t, db, testUser, testUserStory)
	})

	t.Run("TestReferenceIDUnderLoad", func(t *testing.T) {
		testAcceptanceCriteriaReferenceIDUnderLoad(t, db, testUser, testUserStory)
	})

	t.Run("TestProductionGeneratorDirectly", func(t *testing.T) {
		testAcceptanceCriteriaProductionGeneratorDirectly(t, db)
	})
}

func testAcceptanceCriteriaSequentialReferenceIDGeneration(t *testing.T, db *gorm.DB, testUser *User, testUserStory *UserStory) {
	// Clean acceptance_criteria table
	db.Exec("DELETE FROM acceptance_criteria")

	// Create acceptance criteria sequentially and verify reference IDs
	var acceptanceCriteria []AcceptanceCriteria
	for i := 0; i < 5; i++ {
		ac := AcceptanceCriteria{
			UserStoryID: testUserStory.ID,
			AuthorID:    testUser.ID,
			Description: fmt.Sprintf("WHEN user performs action %d THEN system SHALL respond accordingly", i+1),
		}

		err := db.Create(&ac).Error
		require.NoError(t, err)
		acceptanceCriteria = append(acceptanceCriteria, ac)

		// Verify reference ID format
		assert.Regexp(t, `^AC-\d{3}$`, ac.ReferenceID, "Reference ID should match AC-XXX format")
		
		// Verify sequential numbering
		expectedRefID := fmt.Sprintf("AC-%03d", i+1)
		assert.Equal(t, expectedRefID, ac.ReferenceID, "Reference ID should be sequential")
	}

	// Verify all reference IDs are unique
	refIDs := make(map[string]bool)
	for _, ac := range acceptanceCriteria {
		assert.False(t, refIDs[ac.ReferenceID], "Reference ID %s should be unique", ac.ReferenceID)
		refIDs[ac.ReferenceID] = true
	}
}

func testAcceptanceCriteriaConcurrentReferenceIDGeneration(t *testing.T, db *gorm.DB, testUser *User, testUserStory *UserStory) {
	// Clean acceptance_criteria table
	db.Exec("DELETE FROM acceptance_criteria")

	const numGoroutines = 10
	const acceptanceCriteriaPerGoroutine = 5
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allAcceptanceCriteria []AcceptanceCriteria
	var errors []error

	// Create acceptance criteria concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			var localAcceptanceCriteria []AcceptanceCriteria
			for j := 0; j < acceptanceCriteriaPerGoroutine; j++ {
				ac := AcceptanceCriteria{
					UserStoryID: testUserStory.ID,
					AuthorID:    testUser.ID,
					Description: fmt.Sprintf("WHEN concurrent action G%d-AC%d THEN system SHALL handle it", goroutineID, j),
				}

				err := db.Create(&ac).Error
				if err != nil {
					mu.Lock()
					errors = append(errors, err)
					mu.Unlock()
					continue
				}
				localAcceptanceCriteria = append(localAcceptanceCriteria, ac)
			}

			mu.Lock()
			allAcceptanceCriteria = append(allAcceptanceCriteria, localAcceptanceCriteria...)
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Check for errors
	require.Empty(t, errors, "No errors should occur during concurrent creation")

	// Verify we created the expected number of acceptance criteria
	expectedCount := numGoroutines * acceptanceCriteriaPerGoroutine
	assert.Len(t, allAcceptanceCriteria, expectedCount, "Should create all acceptance criteria")

	// Verify all reference IDs are unique
	refIDs := make(map[string]bool)
	for _, ac := range allAcceptanceCriteria {
		assert.False(t, refIDs[ac.ReferenceID], "Reference ID %s should be unique", ac.ReferenceID)
		refIDs[ac.ReferenceID] = true
		
		// Verify format (sequential AC-XXX or fallback AC-xxxxxxxx)
		assert.Regexp(t, `^AC-(\d{3}|[a-f0-9]{8})$`, ac.ReferenceID, "Reference ID should match AC-XXX or AC-xxxxxxxx format")
	}

	// Verify database consistency
	var dbCount int64
	err := db.Model(&AcceptanceCriteria{}).Count(&dbCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(expectedCount), dbCount, "Database should contain all created acceptance criteria")
}

func testAcceptanceCriteriaReferenceIDFormatAndUniqueness(t *testing.T, db *gorm.DB, testUser *User, testUserStory *UserStory) {
	// Clean acceptance_criteria table
	db.Exec("DELETE FROM acceptance_criteria")

	// Test that reference IDs follow the correct format
	ac := AcceptanceCriteria{
		UserStoryID: testUserStory.ID,
		AuthorID:    testUser.ID,
		Description: "WHEN user submits form THEN system SHALL validate all fields",
	}

	err := db.Create(&ac).Error
	require.NoError(t, err)

	// Test format
	assert.Regexp(t, `^AC-\d{3}$`, ac.ReferenceID, "First acceptance criteria should have sequential format AC-001")
	assert.Equal(t, "AC-001", ac.ReferenceID, "First acceptance criteria should be AC-001")

	// Test that manually set reference IDs are preserved
	manualAC := AcceptanceCriteria{
		ReferenceID: "AC-MANUAL",
		UserStoryID: testUserStory.ID,
		AuthorID:    testUser.ID,
		Description: "WHEN manual test THEN system SHALL preserve reference ID",
	}

	err = db.Create(&manualAC).Error
	require.NoError(t, err)
	assert.Equal(t, "AC-MANUAL", manualAC.ReferenceID, "Manual reference ID should be preserved")

	// Test that next auto-generated ID continues sequence
	nextAC := AcceptanceCriteria{
		UserStoryID: testUserStory.ID,
		AuthorID:    testUser.ID,
		Description: "WHEN next sequential test THEN system SHALL continue sequence",
	}

	err = db.Create(&nextAC).Error
	require.NoError(t, err)
	assert.Equal(t, "AC-003", nextAC.ReferenceID, "Should continue sequence after manual ID")
}

func testAcceptanceCriteriaReferenceIDUnderLoad(t *testing.T, db *gorm.DB, testUser *User, testUserStory *UserStory) {
	// Clean acceptance_criteria table
	db.Exec("DELETE FROM acceptance_criteria")

	const numWorkers = 20
	const acceptanceCriteriaPerWorker = 10
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allRefIDs []string
	var errors []error

	// Create acceptance criteria under high concurrency load
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			var localRefIDs []string
			for j := 0; j < acceptanceCriteriaPerWorker; j++ {
				ac := AcceptanceCriteria{
					UserStoryID: testUserStory.ID,
					AuthorID:    testUser.ID,
					Description: fmt.Sprintf("WHEN load test W%d-AC%d THEN system SHALL handle load", workerID, j),
				}

				err := db.Create(&ac).Error
				if err != nil {
					mu.Lock()
					errors = append(errors, err)
					mu.Unlock()
					continue
				}
				localRefIDs = append(localRefIDs, ac.ReferenceID)
			}

			mu.Lock()
			allRefIDs = append(allRefIDs, localRefIDs...)
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Check for errors
	require.Empty(t, errors, "No errors should occur under load")

	// Verify all reference IDs are unique
	refIDMap := make(map[string]int)
	for _, refID := range allRefIDs {
		refIDMap[refID]++
		assert.Equal(t, 1, refIDMap[refID], "Reference ID %s should appear only once", refID)
	}

	expectedCount := numWorkers * acceptanceCriteriaPerWorker
	assert.Len(t, allRefIDs, expectedCount, "Should create all acceptance criteria under load")
}

func testAcceptanceCriteriaProductionGeneratorDirectly(t *testing.T, db *gorm.DB) {
	// Test the production generator directly
	generator := NewPostgreSQLReferenceIDGenerator(2147483644, "AC")
	
	// Clean acceptance_criteria table
	db.Exec("DELETE FROM acceptance_criteria")

	// Create unique test data for this test
	testUser := &User{
		ID:           uuid.New(),
		Username:     "acdirectgentestuser",
		Email:        "acdirectgentest@example.com",
		PasswordHash: "hashedpassword",
		Role:         RoleUser,
	}
	err := db.Create(testUser).Error
	require.NoError(t, err)

	testEpic := &Epic{
		ID:         uuid.New(),
		CreatorID:  testUser.ID,
		AssigneeID: testUser.ID,
		Priority:   PriorityMedium,
		Status:     EpicStatusBacklog,
		Title:      "AC Direct Generator Test Epic",
	}
	err = db.Create(testEpic).Error
	require.NoError(t, err)

	testUserStory := &UserStory{
		ID:         uuid.New(),
		EpicID:     testEpic.ID,
		CreatorID:  testUser.ID,
		AssigneeID: testUser.ID,
		Priority:   PriorityMedium,
		Status:     UserStoryStatusBacklog,
		Title:      "AC Direct Generator Test User Story",
	}
	err = db.Create(testUserStory).Error
	require.NoError(t, err)
	
	for i := 0; i < 3; i++ {
		// Generate reference ID
		refID, err := generator.Generate(db, &AcceptanceCriteria{})
		require.NoError(t, err)
		
		expectedRefID := fmt.Sprintf("AC-%03d", i+1)
		assert.Equal(t, expectedRefID, refID, "Generator should produce sequential IDs")
		
		// Create an acceptance criteria with this reference ID to maintain count for next iteration
		ac := AcceptanceCriteria{
			ReferenceID: refID,
			UserStoryID: testUserStory.ID,
			AuthorID:    testUser.ID,
			Description: fmt.Sprintf("WHEN direct generator test %d THEN system SHALL work", i+1),
		}
		err = db.Create(&ac).Error
		require.NoError(t, err)
	}

	// Test that the generator is the same one used by the model
	assert.Equal(t, int64(2147483644), acceptanceCriteriaGenerator.lockKey, "Model should use correct lock key")
	assert.Equal(t, "AC", acceptanceCriteriaGenerator.prefix, "Model should use correct prefix")
}

func setupPostgreSQLForAcceptanceCriteriaReferenceIDTest(t *testing.T) *gorm.DB {
	ctx := context.Background()

	// Create PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "acceptance_criteria_ref_test",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_USER":     "testuser",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort("5432/tcp"),
		).WithDeadline(60 * time.Second),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Cleanup container when test finishes
	t.Cleanup(func() {
		err := postgresContainer.Terminate(ctx)
		if err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	})

	// Get connection details
	host, err := postgresContainer.Host(ctx)
	require.NoError(t, err)

	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Create database connection
	dsn := fmt.Sprintf("host=%s port=%s user=testuser password=testpass dbname=acceptance_criteria_ref_test sslmode=disable", 
		host, port.Port())
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	// Verify connection
	sqlDB, err := db.DB()
	require.NoError(t, err)
	
	err = sqlDB.Ping()
	require.NoError(t, err)

	return db
}

func createTestDataForAcceptanceCriteria(t *testing.T, db *gorm.DB) (*User, *UserStory) {
	// Create test user
	testUser := &User{
		ID:           uuid.New(),
		Username:     "acrefidtestuser",
		Email:        "acrefidtest@example.com",
		PasswordHash: "hashedpassword",
		Role:         RoleUser,
	}
	err := db.Create(testUser).Error
	require.NoError(t, err)

	// Create test epic
	testEpic := &Epic{
		ID:         uuid.New(),
		CreatorID:  testUser.ID,
		AssigneeID: testUser.ID,
		Priority:   PriorityMedium,
		Status:     EpicStatusBacklog,
		Title:      "AC Reference ID Test Epic",
	}
	err = db.Create(testEpic).Error
	require.NoError(t, err)

	// Create test user story
	testUserStory := &UserStory{
		ID:         uuid.New(),
		EpicID:     testEpic.ID,
		CreatorID:  testUser.ID,
		AssigneeID: testUser.ID,
		Priority:   PriorityMedium,
		Status:     UserStoryStatusBacklog,
		Title:      "AC Reference ID Test User Story",
	}
	err = db.Create(testUserStory).Error
	require.NoError(t, err)

	return testUser, testUserStory
}

func cleanupPostgreSQLAcceptanceCriteriaTest(t *testing.T, db *gorm.DB) {
	// Clean up test data
	tables := []string{
		"acceptance_criteria",
		"user_stories",
		"epics",
		"users",
	}

	for _, table := range tables {
		err := db.Exec("DELETE FROM " + table).Error
		if err != nil {
			t.Logf("Warning: Could not clean table %s: %v", table, err)
		}
	}
}