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

// TestRequirementReferenceIDProductionGenerator tests the production reference ID generator
// with PostgreSQL using testcontainers for concurrent operations and uniqueness
func TestRequirementReferenceIDProductionGenerator(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupPostgreSQLForReferenceIDTest(t)
	defer cleanupPostgreSQLTest(t, db)

	// Auto-migrate required models
	err := db.AutoMigrate(
		&User{},
		&Epic{},
		&UserStory{},
		&AcceptanceCriteria{},
		&Requirement{},
		&RequirementType{},
	)
	require.NoError(t, err)

	// Create test data
	testUser, testUserStory, testRequirementType := createTestDataForRequirement(t, db)

	t.Run("TestSequentialReferenceIDGeneration", func(t *testing.T) {
		testSequentialReferenceIDGeneration(t, db, testUser, testUserStory, testRequirementType)
	})

	t.Run("TestConcurrentReferenceIDGeneration", func(t *testing.T) {
		testConcurrentReferenceIDGeneration(t, db, testUser, testUserStory, testRequirementType)
	})

	t.Run("TestReferenceIDFormatAndUniqueness", func(t *testing.T) {
		testReferenceIDFormatAndUniqueness(t, db, testUser, testUserStory, testRequirementType)
	})

	t.Run("TestReferenceIDUnderLoad", func(t *testing.T) {
		testReferenceIDUnderLoad(t, db, testUser, testUserStory, testRequirementType)
	})

	t.Run("TestProductionGeneratorDirectly", func(t *testing.T) {
		testProductionGeneratorDirectly(t, db)
	})
}

func testSequentialReferenceIDGeneration(t *testing.T, db *gorm.DB, testUser *User, testUserStory *UserStory, testRequirementType *RequirementType) {
	// Clean requirements table
	db.Exec("DELETE FROM requirements")

	// Create requirements sequentially and verify reference IDs
	var requirements []Requirement
	for i := 0; i < 5; i++ {
		req := Requirement{
			UserStoryID: testUserStory.ID,
			CreatorID:   testUser.ID,
			AssigneeID:  testUser.ID,
			Priority:    PriorityMedium,
			Status:      RequirementStatusDraft,
			TypeID:      testRequirementType.ID,
			Title:       fmt.Sprintf("Sequential Requirement %d", i+1),
		}

		err := db.Create(&req).Error
		require.NoError(t, err)
		requirements = append(requirements, req)

		// Verify reference ID format
		assert.Regexp(t, `^REQ-\d{3}$`, req.ReferenceID, "Reference ID should match REQ-XXX format")

		// Verify sequential numbering
		expectedRefID := fmt.Sprintf("REQ-%03d", i+1)
		assert.Equal(t, expectedRefID, req.ReferenceID, "Reference ID should be sequential")
	}

	// Verify all reference IDs are unique
	refIDs := make(map[string]bool)
	for _, req := range requirements {
		assert.False(t, refIDs[req.ReferenceID], "Reference ID %s should be unique", req.ReferenceID)
		refIDs[req.ReferenceID] = true
	}
}

func testConcurrentReferenceIDGeneration(t *testing.T, db *gorm.DB, testUser *User, testUserStory *UserStory, testRequirementType *RequirementType) {
	// Clean requirements table
	db.Exec("DELETE FROM requirements")

	const numGoroutines = 10
	const requirementsPerGoroutine = 5

	var wg sync.WaitGroup
	var mu sync.Mutex
	var allRequirements []Requirement
	var errors []error

	// Create requirements concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			var localRequirements []Requirement
			for j := 0; j < requirementsPerGoroutine; j++ {
				req := Requirement{
					UserStoryID: testUserStory.ID,
					CreatorID:   testUser.ID,
					AssigneeID:  testUser.ID,
					Priority:    PriorityMedium,
					Status:      RequirementStatusDraft,
					TypeID:      testRequirementType.ID,
					Title:       fmt.Sprintf("Concurrent Requirement G%d-R%d", goroutineID, j),
				}

				err := db.Create(&req).Error
				if err != nil {
					mu.Lock()
					errors = append(errors, err)
					mu.Unlock()
					continue
				}
				localRequirements = append(localRequirements, req)
			}

			mu.Lock()
			allRequirements = append(allRequirements, localRequirements...)
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Check for errors
	require.Empty(t, errors, "No errors should occur during concurrent creation")

	// Verify we created the expected number of requirements
	expectedCount := numGoroutines * requirementsPerGoroutine
	assert.Len(t, allRequirements, expectedCount, "Should create all requirements")

	// Verify all reference IDs are unique
	refIDs := make(map[string]bool)
	for _, req := range allRequirements {
		assert.False(t, refIDs[req.ReferenceID], "Reference ID %s should be unique", req.ReferenceID)
		refIDs[req.ReferenceID] = true

		// Verify format
		assert.Regexp(t, `^REQ-(\d{3}|[a-f0-9]{8})$`, req.ReferenceID, "Reference ID should match REQ-XXX or REQ-xxxxxxxx format")
	}

	// Verify database consistency
	var dbCount int64
	err := db.Model(&Requirement{}).Count(&dbCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(expectedCount), dbCount, "Database should contain all created requirements")
}

func testReferenceIDFormatAndUniqueness(t *testing.T, db *gorm.DB, testUser *User, testUserStory *UserStory, testRequirementType *RequirementType) {
	// Clean requirements table
	db.Exec("DELETE FROM requirements")

	// Test that reference IDs follow the correct format
	req := Requirement{
		UserStoryID: testUserStory.ID,
		CreatorID:   testUser.ID,
		AssigneeID:  testUser.ID,
		Priority:    PriorityHigh,
		Status:      RequirementStatusDraft,
		TypeID:      testRequirementType.ID,
		Title:       "Format Test Requirement",
	}

	err := db.Create(&req).Error
	require.NoError(t, err)

	// Test format
	assert.Regexp(t, `^REQ-\d{3}$`, req.ReferenceID, "First requirement should have sequential format REQ-001")
	assert.Equal(t, "REQ-001", req.ReferenceID, "First requirement should be REQ-001")

	// Test that manually set reference IDs are preserved
	manualReq := Requirement{
		ReferenceID: "REQ-MANUAL",
		UserStoryID: testUserStory.ID,
		CreatorID:   testUser.ID,
		AssigneeID:  testUser.ID,
		Priority:    PriorityHigh,
		Status:      RequirementStatusDraft,
		TypeID:      testRequirementType.ID,
		Title:       "Manual Reference ID Test",
	}

	err = db.Create(&manualReq).Error
	require.NoError(t, err)
	assert.Equal(t, "REQ-MANUAL", manualReq.ReferenceID, "Manual reference ID should be preserved")

	// Test that next auto-generated ID continues sequence
	nextReq := Requirement{
		UserStoryID: testUserStory.ID,
		CreatorID:   testUser.ID,
		AssigneeID:  testUser.ID,
		Priority:    PriorityHigh,
		Status:      RequirementStatusDraft,
		TypeID:      testRequirementType.ID,
		Title:       "Next Sequential Requirement",
	}

	err = db.Create(&nextReq).Error
	require.NoError(t, err)
	assert.Equal(t, "REQ-003", nextReq.ReferenceID, "Should continue sequence after manual ID")
}

func testReferenceIDUnderLoad(t *testing.T, db *gorm.DB, testUser *User, testUserStory *UserStory, testRequirementType *RequirementType) {
	// Clean requirements table
	db.Exec("DELETE FROM requirements")

	const numWorkers = 20
	const requirementsPerWorker = 10

	var wg sync.WaitGroup
	var mu sync.Mutex
	var allRefIDs []string
	var errors []error

	// Create requirements under high concurrency load
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			var localRefIDs []string
			for j := 0; j < requirementsPerWorker; j++ {
				req := Requirement{
					UserStoryID: testUserStory.ID,
					CreatorID:   testUser.ID,
					AssigneeID:  testUser.ID,
					Priority:    PriorityLow,
					Status:      RequirementStatusDraft,
					TypeID:      testRequirementType.ID,
					Title:       fmt.Sprintf("Load Test Req W%d-R%d", workerID, j),
				}

				err := db.Create(&req).Error
				if err != nil {
					mu.Lock()
					errors = append(errors, err)
					mu.Unlock()
					continue
				}
				localRefIDs = append(localRefIDs, req.ReferenceID)
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

	expectedCount := numWorkers * requirementsPerWorker
	assert.Len(t, allRefIDs, expectedCount, "Should create all requirements under load")
}

func testProductionGeneratorDirectly(t *testing.T, db *gorm.DB) {
	// Test the production generator directly
	generator := NewPostgreSQLReferenceIDGenerator(2147483645, "REQ")

	// Clean requirements table
	db.Exec("DELETE FROM requirements")

	// Create unique test data for this test
	testUser := &User{
		ID:           uuid.New(),
		Username:     "directgentestuser",
		Email:        "directgentest@example.com",
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
		Title:      "Direct Generator Test Epic",
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
		Title:      "Direct Generator Test User Story",
	}
	err = db.Create(testUserStory).Error
	require.NoError(t, err)

	testRequirementType := &RequirementType{
		ID:          uuid.New(),
		Name:        "DirectGenTestType",
		Description: stringPtr("Test requirement type for direct generator testing"),
	}
	err = db.Create(testRequirementType).Error
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		// Generate reference ID
		refID, err := generator.Generate(db, &Requirement{})
		require.NoError(t, err)

		expectedRefID := fmt.Sprintf("REQ-%03d", i+1)
		assert.Equal(t, expectedRefID, refID, "Generator should produce sequential IDs")

		// Create a requirement with this reference ID to maintain count for next iteration
		req := Requirement{
			ReferenceID: refID,
			UserStoryID: testUserStory.ID,
			CreatorID:   testUser.ID,
			AssigneeID:  testUser.ID,
			Priority:    PriorityLow,
			Status:      RequirementStatusDraft,
			TypeID:      testRequirementType.ID,
			Title:       fmt.Sprintf("Direct Generator Test %d", i+1),
		}
		err = db.Create(&req).Error
		require.NoError(t, err)
	}

	// Test that the generator is the same one used by the model
	assert.Equal(t, int64(2147483645), requirementGenerator.lockKey, "Model should use correct lock key")
	assert.Equal(t, "REQ", requirementGenerator.prefix, "Model should use correct prefix")
}

func setupPostgreSQLForReferenceIDTest(t *testing.T) *gorm.DB {
	ctx := context.Background()

	// Create PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "requirement_ref_test",
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
	dsn := fmt.Sprintf("host=%s port=%s user=testuser password=testpass dbname=requirement_ref_test sslmode=disable",
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

func createTestDataForRequirement(t *testing.T, db *gorm.DB) (*User, *UserStory, *RequirementType) {
	// Create test user
	testUser := &User{
		ID:           uuid.New(),
		Username:     "refidtestuser",
		Email:        "refidtest@example.com",
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
		Title:      "Reference ID Test Epic",
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
		Title:      "Reference ID Test User Story",
	}
	err = db.Create(testUserStory).Error
	require.NoError(t, err)

	// Create test requirement type
	testRequirementType := &RequirementType{
		ID:          uuid.New(),
		Name:        "RefIDTestType",
		Description: stringPtr("Test requirement type for reference ID testing"),
	}
	err = db.Create(testRequirementType).Error
	require.NoError(t, err)

	return testUser, testUserStory, testRequirementType
}

func cleanupPostgreSQLTest(t *testing.T, db *gorm.DB) {
	// Clean up test data
	tables := []string{
		"requirements",
		"requirement_types",
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
