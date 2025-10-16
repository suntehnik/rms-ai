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

// TestAllEntitiesReferenceIDProductionGenerator tests all entity types together
// to verify reference ID generation consistency and uniqueness across all entities
func TestAllEntitiesReferenceIDProductionGenerator(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive integration test in short mode")
	}

	db := setupPostgreSQLForComprehensiveTest(t)
	defer cleanupPostgreSQLComprehensiveTest(t, db)

	// Auto-migrate all required models
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
	testUser, testEpic, testUserStory, testRequirementType := createComprehensiveTestData(t, db)

	t.Run("TestAllEntitiesSequentialGeneration", func(t *testing.T) {
		testAllEntitiesSequentialGeneration(t, db, testUser, testEpic, testUserStory, testRequirementType)
	})

	t.Run("TestAllEntitiesConcurrentGeneration", func(t *testing.T) {
		testAllEntitiesConcurrentGeneration(t, db, testUser, testEpic, testUserStory, testRequirementType)
	})

	t.Run("TestReferenceIDUniquenessAcrossAllEntities", func(t *testing.T) {
		testReferenceIDUniquenessAcrossAllEntities(t, db, testUser, testEpic, testUserStory, testRequirementType)
	})

	t.Run("TestReferenceIDFormatConsistency", func(t *testing.T) {
		testReferenceIDFormatConsistency(t, db, testUser, testEpic, testUserStory, testRequirementType)
	})

	t.Run("TestAllEntitiesUnderLoad", func(t *testing.T) {
		testAllEntitiesUnderLoad(t, db, testUser, testEpic, testUserStory, testRequirementType)
	})

	t.Run("TestProductionGeneratorsDirectly", func(t *testing.T) {
		testAllProductionGeneratorsDirectly(t, db)
	})
}

func testAllEntitiesSequentialGeneration(t *testing.T, db *gorm.DB, testUser *User, testEpic *Epic, testUserStory *UserStory, testRequirementType *RequirementType) {
	// Clean all entity tables
	cleanAllEntityTables(db)

	// Test sequential generation for each entity type
	t.Run("Epics", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			epic := Epic{
				CreatorID:  testUser.ID,
				AssigneeID: testUser.ID,
				Priority:   PriorityMedium,
				Status:     EpicStatusBacklog,
				Title:      fmt.Sprintf("Sequential Epic %d", i+1),
			}

			err := db.Create(&epic).Error
			require.NoError(t, err)

			expectedRefID := fmt.Sprintf("EP-%03d", i+1)
			assert.Equal(t, expectedRefID, epic.ReferenceID, "Epic reference ID should be sequential")
			assert.Regexp(t, `^EP-\d{3}$`, epic.ReferenceID, "Epic reference ID should match EP-XXX format")
		}
	})

	t.Run("UserStories", func(t *testing.T) {
		// Create a fresh epic for user stories
		epic := Epic{
			CreatorID:  testUser.ID,
			AssigneeID: testUser.ID,
			Priority:   PriorityMedium,
			Status:     EpicStatusBacklog,
			Title:      "Sequential Test Epic for UserStories",
		}
		err := db.Create(&epic).Error
		require.NoError(t, err)

		for i := 0; i < 3; i++ {
			userStory := UserStory{
				EpicID:     epic.ID,
				CreatorID:  testUser.ID,
				AssigneeID: testUser.ID,
				Priority:   PriorityMedium,
				Status:     UserStoryStatusBacklog,
				Title:      fmt.Sprintf("Sequential User Story %d", i+1),
			}

			err := db.Create(&userStory).Error
			require.NoError(t, err)

			expectedRefID := fmt.Sprintf("US-%03d", i+1)
			assert.Equal(t, expectedRefID, userStory.ReferenceID, "UserStory reference ID should be sequential")
			assert.Regexp(t, `^US-\d{3}$`, userStory.ReferenceID, "UserStory reference ID should match US-XXX format")
		}
	})

	t.Run("Requirements", func(t *testing.T) {
		// Create a fresh epic and user story for requirements
		epic := Epic{
			CreatorID:  testUser.ID,
			AssigneeID: testUser.ID,
			Priority:   PriorityMedium,
			Status:     EpicStatusBacklog,
			Title:      "Sequential Test Epic for Requirements",
		}
		err := db.Create(&epic).Error
		require.NoError(t, err)

		userStory := UserStory{
			EpicID:     epic.ID,
			CreatorID:  testUser.ID,
			AssigneeID: testUser.ID,
			Priority:   PriorityMedium,
			Status:     UserStoryStatusBacklog,
			Title:      "Sequential Test UserStory for Requirements",
		}
		err = db.Create(&userStory).Error
		require.NoError(t, err)

		for i := 0; i < 3; i++ {
			requirement := Requirement{
				UserStoryID: userStory.ID,
				CreatorID:   testUser.ID,
				AssigneeID:  testUser.ID,
				Priority:    PriorityMedium,
				Status:      RequirementStatusDraft,
				TypeID:      testRequirementType.ID,
				Title:       fmt.Sprintf("Sequential Requirement %d", i+1),
			}

			err := db.Create(&requirement).Error
			require.NoError(t, err)

			expectedRefID := fmt.Sprintf("REQ-%03d", i+1)
			assert.Equal(t, expectedRefID, requirement.ReferenceID, "Requirement reference ID should be sequential")
			assert.Regexp(t, `^REQ-\d{3}$`, requirement.ReferenceID, "Requirement reference ID should match REQ-XXX format")
		}
	})

	t.Run("AcceptanceCriteria", func(t *testing.T) {
		// Create a fresh epic and user story for acceptance criteria
		epic := Epic{
			CreatorID:  testUser.ID,
			AssigneeID: testUser.ID,
			Priority:   PriorityMedium,
			Status:     EpicStatusBacklog,
			Title:      "Sequential Test Epic for AcceptanceCriteria",
		}
		err := db.Create(&epic).Error
		require.NoError(t, err)

		userStory := UserStory{
			EpicID:     epic.ID,
			CreatorID:  testUser.ID,
			AssigneeID: testUser.ID,
			Priority:   PriorityMedium,
			Status:     UserStoryStatusBacklog,
			Title:      "Sequential Test UserStory for AcceptanceCriteria",
		}
		err = db.Create(&userStory).Error
		require.NoError(t, err)

		for i := 0; i < 3; i++ {
			acceptanceCriteria := AcceptanceCriteria{
				UserStoryID: userStory.ID,
				AuthorID:    testUser.ID,
				Description: fmt.Sprintf("WHEN sequential test %d THEN system SHALL respond", i+1),
			}

			err := db.Create(&acceptanceCriteria).Error
			require.NoError(t, err)

			expectedRefID := fmt.Sprintf("AC-%03d", i+1)
			assert.Equal(t, expectedRefID, acceptanceCriteria.ReferenceID, "AcceptanceCriteria reference ID should be sequential")
			assert.Regexp(t, `^AC-\d{3}$`, acceptanceCriteria.ReferenceID, "AcceptanceCriteria reference ID should match AC-XXX format")
		}
	})
}

func testAllEntitiesConcurrentGeneration(t *testing.T, db *gorm.DB, testUser *User, testEpic *Epic, testUserStory *UserStory, testRequirementType *RequirementType) {
	// Clean all entity tables
	cleanAllEntityTables(db)

	const numGoroutines = 5
	const entitiesPerGoroutine = 3

	var wg sync.WaitGroup
	var mu sync.Mutex
	var allRefIDs []string
	var errors []error

	// Create all entity types concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			var localRefIDs []string

			// Create Epics first (they don't depend on other entities)
			var localEpic *Epic
			for j := 0; j < entitiesPerGoroutine; j++ {
				epic := Epic{
					CreatorID:  testUser.ID,
					AssigneeID: testUser.ID,
					Priority:   PriorityMedium,
					Status:     EpicStatusBacklog,
					Title:      fmt.Sprintf("Concurrent Epic G%d-E%d", goroutineID, j),
				}

				err := db.Create(&epic).Error
				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("epic creation error: %w", err))
					mu.Unlock()
					continue
				}
				localRefIDs = append(localRefIDs, epic.ReferenceID)

				// Use the first epic as parent for user stories
				if j == 0 {
					localEpic = &epic
				}
			}

			// Create UserStories using the local epic
			var localUserStory *UserStory
			if localEpic != nil {
				for j := 0; j < entitiesPerGoroutine; j++ {
					userStory := UserStory{
						EpicID:     localEpic.ID,
						CreatorID:  testUser.ID,
						AssigneeID: testUser.ID,
						Priority:   PriorityMedium,
						Status:     UserStoryStatusBacklog,
						Title:      fmt.Sprintf("Concurrent UserStory G%d-US%d", goroutineID, j),
					}

					err := db.Create(&userStory).Error
					if err != nil {
						mu.Lock()
						errors = append(errors, fmt.Errorf("user story creation error: %w", err))
						mu.Unlock()
						continue
					}
					localRefIDs = append(localRefIDs, userStory.ReferenceID)

					// Use the first user story as parent for requirements and acceptance criteria
					if j == 0 {
						localUserStory = &userStory
					}
				}
			}

			// Create Requirements using the local user story
			if localUserStory != nil {
				for j := 0; j < entitiesPerGoroutine; j++ {
					requirement := Requirement{
						UserStoryID: localUserStory.ID,
						CreatorID:   testUser.ID,
						AssigneeID:  testUser.ID,
						Priority:    PriorityMedium,
						Status:      RequirementStatusDraft,
						TypeID:      testRequirementType.ID,
						Title:       fmt.Sprintf("Concurrent Requirement G%d-R%d", goroutineID, j),
					}

					err := db.Create(&requirement).Error
					if err != nil {
						mu.Lock()
						errors = append(errors, fmt.Errorf("requirement creation error: %w", err))
						mu.Unlock()
						continue
					}
					localRefIDs = append(localRefIDs, requirement.ReferenceID)
				}

				// Create AcceptanceCriteria using the local user story
				for j := 0; j < entitiesPerGoroutine; j++ {
					acceptanceCriteria := AcceptanceCriteria{
						UserStoryID: localUserStory.ID,
						AuthorID:    testUser.ID,
						Description: fmt.Sprintf("WHEN concurrent test G%d-AC%d THEN system SHALL handle it", goroutineID, j),
					}

					err := db.Create(&acceptanceCriteria).Error
					if err != nil {
						mu.Lock()
						errors = append(errors, fmt.Errorf("acceptance criteria creation error: %w", err))
						mu.Unlock()
						continue
					}
					localRefIDs = append(localRefIDs, acceptanceCriteria.ReferenceID)
				}
			}

			mu.Lock()
			allRefIDs = append(allRefIDs, localRefIDs...)
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Check for errors
	require.Empty(t, errors, "No errors should occur during concurrent creation")

	// Verify all reference IDs are unique across all entity types
	refIDMap := make(map[string]bool)
	for _, refID := range allRefIDs {
		assert.False(t, refIDMap[refID], "Reference ID %s should be unique across all entities", refID)
		refIDMap[refID] = true
	}

	expectedCount := numGoroutines * entitiesPerGoroutine * 4 // 4 entity types
	assert.Len(t, allRefIDs, expectedCount, "Should create all entities concurrently")
}

func testReferenceIDUniquenessAcrossAllEntities(t *testing.T, db *gorm.DB, testUser *User, testEpic *Epic, testUserStory *UserStory, testRequirementType *RequirementType) {
	// Clean all entity tables
	cleanAllEntityTables(db)

	var allRefIDs []string

	// Create multiple entities of each type with proper hierarchy
	for i := 0; i < 5; i++ {
		// Create Epic
		epic := Epic{
			CreatorID:  testUser.ID,
			AssigneeID: testUser.ID,
			Priority:   PriorityMedium,
			Status:     EpicStatusBacklog,
			Title:      fmt.Sprintf("Uniqueness Test Epic %d", i+1),
		}
		err := db.Create(&epic).Error
		require.NoError(t, err)
		allRefIDs = append(allRefIDs, epic.ReferenceID)

		// Create UserStory using the created epic
		userStory := UserStory{
			EpicID:     epic.ID,
			CreatorID:  testUser.ID,
			AssigneeID: testUser.ID,
			Priority:   PriorityMedium,
			Status:     UserStoryStatusBacklog,
			Title:      fmt.Sprintf("Uniqueness Test UserStory %d", i+1),
		}
		err = db.Create(&userStory).Error
		require.NoError(t, err)
		allRefIDs = append(allRefIDs, userStory.ReferenceID)

		// Create Requirement using the created user story
		requirement := Requirement{
			UserStoryID: userStory.ID,
			CreatorID:   testUser.ID,
			AssigneeID:  testUser.ID,
			Priority:    PriorityMedium,
			Status:      RequirementStatusDraft,
			TypeID:      testRequirementType.ID,
			Title:       fmt.Sprintf("Uniqueness Test Requirement %d", i+1),
		}
		err = db.Create(&requirement).Error
		require.NoError(t, err)
		allRefIDs = append(allRefIDs, requirement.ReferenceID)

		// Create AcceptanceCriteria using the created user story
		acceptanceCriteria := AcceptanceCriteria{
			UserStoryID: userStory.ID,
			AuthorID:    testUser.ID,
			Description: fmt.Sprintf("WHEN uniqueness test %d THEN system SHALL ensure uniqueness", i+1),
		}
		err = db.Create(&acceptanceCriteria).Error
		require.NoError(t, err)
		allRefIDs = append(allRefIDs, acceptanceCriteria.ReferenceID)
	}

	// Verify all reference IDs are unique across all entity types
	refIDMap := make(map[string]bool)
	for _, refID := range allRefIDs {
		assert.False(t, refIDMap[refID], "Reference ID %s should be unique across all entities", refID)
		refIDMap[refID] = true
	}

	// Verify we have the expected number of unique IDs
	assert.Len(t, allRefIDs, 20, "Should have 20 unique reference IDs (5 of each entity type)")
	assert.Len(t, refIDMap, 20, "All reference IDs should be unique")
}

func testReferenceIDFormatConsistency(t *testing.T, db *gorm.DB, testUser *User, testEpic *Epic, testUserStory *UserStory, testRequirementType *RequirementType) {
	// Clean all entity tables
	cleanAllEntityTables(db)

	// Create fresh parent entities for the test
	testEpicForFormat := Epic{
		CreatorID:  testUser.ID,
		AssigneeID: testUser.ID,
		Priority:   PriorityMedium,
		Status:     EpicStatusBacklog,
		Title:      "Format Test Epic Parent",
	}
	err := db.Create(&testEpicForFormat).Error
	require.NoError(t, err)

	testUserStoryForFormat := UserStory{
		EpicID:     testEpicForFormat.ID,
		CreatorID:  testUser.ID,
		AssigneeID: testUser.ID,
		Priority:   PriorityMedium,
		Status:     UserStoryStatusBacklog,
		Title:      "Format Test UserStory Parent",
	}
	err = db.Create(&testUserStoryForFormat).Error
	require.NoError(t, err)

	// Test format consistency for each entity type
	entities := []struct {
		name          string
		expectedRegex string
		createFunc    func() (string, error)
	}{
		{
			name:          "Epic",
			expectedRegex: `^EP-(\d{3}|[a-f0-9]{8})$`,
			createFunc: func() (string, error) {
				epic := Epic{
					CreatorID:  testUser.ID,
					AssigneeID: testUser.ID,
					Priority:   PriorityMedium,
					Status:     EpicStatusBacklog,
					Title:      "Format Test Epic",
				}
				err := db.Create(&epic).Error
				return epic.ReferenceID, err
			},
		},
		{
			name:          "UserStory",
			expectedRegex: `^US-(\d{3}|[a-f0-9]{8})$`,
			createFunc: func() (string, error) {
				userStory := UserStory{
					EpicID:     testEpicForFormat.ID,
					CreatorID:  testUser.ID,
					AssigneeID: testUser.ID,
					Priority:   PriorityMedium,
					Status:     UserStoryStatusBacklog,
					Title:      "Format Test UserStory",
				}
				err := db.Create(&userStory).Error
				return userStory.ReferenceID, err
			},
		},
		{
			name:          "Requirement",
			expectedRegex: `^REQ-(\d{3}|[a-f0-9]{8})$`,
			createFunc: func() (string, error) {
				requirement := Requirement{
					UserStoryID: testUserStoryForFormat.ID,
					CreatorID:   testUser.ID,
					AssigneeID:  testUser.ID,
					Priority:    PriorityMedium,
					Status:      RequirementStatusDraft,
					TypeID:      testRequirementType.ID,
					Title:       "Format Test Requirement",
				}
				err := db.Create(&requirement).Error
				return requirement.ReferenceID, err
			},
		},
		{
			name:          "AcceptanceCriteria",
			expectedRegex: `^AC-(\d{3}|[a-f0-9]{8})$`,
			createFunc: func() (string, error) {
				acceptanceCriteria := AcceptanceCriteria{
					UserStoryID: testUserStoryForFormat.ID,
					AuthorID:    testUser.ID,
					Description: "WHEN format test THEN system SHALL maintain format consistency",
				}
				err := db.Create(&acceptanceCriteria).Error
				return acceptanceCriteria.ReferenceID, err
			},
		},
	}

	for _, entity := range entities {
		t.Run(entity.name, func(t *testing.T) {
			refID, err := entity.createFunc()
			require.NoError(t, err)
			assert.Regexp(t, entity.expectedRegex, refID, "%s reference ID should match expected format", entity.name)
		})
	}
}

func testAllEntitiesUnderLoad(t *testing.T, db *gorm.DB, testUser *User, testEpic *Epic, testUserStory *UserStory, testRequirementType *RequirementType) {
	// Clean all entity tables
	cleanAllEntityTables(db)

	const numWorkers = 10
	const entitiesPerWorker = 2 // Reduced to avoid too many foreign key issues

	var wg sync.WaitGroup
	var mu sync.Mutex
	var allRefIDs []string
	var errors []error

	// Create all entity types under high load
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			var localRefIDs []string

			// Create entities with proper hierarchy
			for j := 0; j < entitiesPerWorker; j++ {
				// Epic
				epic := Epic{
					CreatorID:  testUser.ID,
					AssigneeID: testUser.ID,
					Priority:   PriorityMedium,
					Status:     EpicStatusBacklog,
					Title:      fmt.Sprintf("Load Test Epic W%d-E%d", workerID, j),
				}
				err := db.Create(&epic).Error
				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("epic load test error: %w", err))
					mu.Unlock()
					continue
				}
				localRefIDs = append(localRefIDs, epic.ReferenceID)

				// UserStory using the created epic
				userStory := UserStory{
					EpicID:     epic.ID,
					CreatorID:  testUser.ID,
					AssigneeID: testUser.ID,
					Priority:   PriorityMedium,
					Status:     UserStoryStatusBacklog,
					Title:      fmt.Sprintf("Load Test UserStory W%d-US%d", workerID, j),
				}
				err = db.Create(&userStory).Error
				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("user story load test error: %w", err))
					mu.Unlock()
					continue
				}
				localRefIDs = append(localRefIDs, userStory.ReferenceID)

				// Requirement using the created user story
				requirement := Requirement{
					UserStoryID: userStory.ID,
					CreatorID:   testUser.ID,
					AssigneeID:  testUser.ID,
					Priority:    PriorityMedium,
					Status:      RequirementStatusDraft,
					TypeID:      testRequirementType.ID,
					Title:       fmt.Sprintf("Load Test Requirement W%d-R%d", workerID, j),
				}
				err = db.Create(&requirement).Error
				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("requirement load test error: %w", err))
					mu.Unlock()
					continue
				}
				localRefIDs = append(localRefIDs, requirement.ReferenceID)

				// AcceptanceCriteria using the created user story
				acceptanceCriteria := AcceptanceCriteria{
					UserStoryID: userStory.ID,
					AuthorID:    testUser.ID,
					Description: fmt.Sprintf("WHEN load test W%d-AC%d THEN system SHALL handle load", workerID, j),
				}
				err = db.Create(&acceptanceCriteria).Error
				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("acceptance criteria load test error: %w", err))
					mu.Unlock()
					continue
				}
				localRefIDs = append(localRefIDs, acceptanceCriteria.ReferenceID)
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
		assert.Equal(t, 1, refIDMap[refID], "Reference ID %s should appear only once under load", refID)
	}

	expectedCount := numWorkers * entitiesPerWorker * 4 // 4 entity types
	assert.Len(t, allRefIDs, expectedCount, "Should create all entities under load")
}

func testAllProductionGeneratorsDirectly(t *testing.T, db *gorm.DB) {
	// Test all production generators directly to ensure they use the correct configuration
	generators := []struct {
		name      string
		generator *PostgreSQLReferenceIDGenerator
		lockKey   int64
		prefix    string
		model     interface{}
	}{
		{
			name:      "Epic",
			generator: NewPostgreSQLReferenceIDGenerator(2147483647, "EP"),
			lockKey:   2147483647,
			prefix:    "EP",
			model:     &Epic{},
		},
		{
			name:      "UserStory",
			generator: NewPostgreSQLReferenceIDGenerator(2147483646, "US"),
			lockKey:   2147483646,
			prefix:    "US",
			model:     &UserStory{},
		},
		{
			name:      "Requirement",
			generator: NewPostgreSQLReferenceIDGenerator(2147483645, "REQ"),
			lockKey:   2147483645,
			prefix:    "REQ",
			model:     &Requirement{},
		},
		{
			name:      "AcceptanceCriteria",
			generator: NewPostgreSQLReferenceIDGenerator(2147483644, "AC"),
			lockKey:   2147483644,
			prefix:    "AC",
			model:     &AcceptanceCriteria{},
		},
	}

	for _, gen := range generators {
		t.Run(gen.name, func(t *testing.T) {
			// Verify generator configuration
			assert.Equal(t, gen.lockKey, gen.generator.lockKey, "%s generator should use correct lock key", gen.name)
			assert.Equal(t, gen.prefix, gen.generator.prefix, "%s generator should use correct prefix", gen.name)

			// Test ID generation
			refID, err := gen.generator.Generate(db, gen.model)
			require.NoError(t, err)
			assert.Regexp(t, fmt.Sprintf(`^%s-(\d{3}|[a-f0-9]{8})$`, gen.prefix), refID, "%s generator should produce correct format", gen.name)
		})
	}

	// Verify that the package-level generators match the expected configuration
	assert.Equal(t, int64(2147483647), epicGenerator.lockKey, "Epic package generator should use correct lock key")
	assert.Equal(t, "EP", epicGenerator.prefix, "Epic package generator should use correct prefix")

	assert.Equal(t, int64(2147483646), userStoryGenerator.lockKey, "UserStory package generator should use correct lock key")
	assert.Equal(t, "US", userStoryGenerator.prefix, "UserStory package generator should use correct prefix")

	assert.Equal(t, int64(2147483645), requirementGenerator.lockKey, "Requirement package generator should use correct lock key")
	assert.Equal(t, "REQ", requirementGenerator.prefix, "Requirement package generator should use correct prefix")

	assert.Equal(t, int64(2147483644), acceptanceCriteriaGenerator.lockKey, "AcceptanceCriteria package generator should use correct lock key")
	assert.Equal(t, "AC", acceptanceCriteriaGenerator.prefix, "AcceptanceCriteria package generator should use correct prefix")
}

// Helper functions

func setupPostgreSQLForComprehensiveTest(t *testing.T) *gorm.DB {
	ctx := context.Background()

	// Create PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "comprehensive_ref_test",
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
	dsn := fmt.Sprintf("host=%s port=%s user=testuser password=testpass dbname=comprehensive_ref_test sslmode=disable",
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

func createComprehensiveTestData(t *testing.T, db *gorm.DB) (*User, *Epic, *UserStory, *RequirementType) {
	// Create test user
	testUser := &User{
		ID:           uuid.New(),
		Username:     "comprehensivetestuser",
		Email:        "comprehensive@example.com",
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
		Title:      "Comprehensive Test Epic",
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
		Title:      "Comprehensive Test User Story",
	}
	err = db.Create(testUserStory).Error
	require.NoError(t, err)

	// Create test requirement type
	testRequirementType := &RequirementType{
		ID:          uuid.New(),
		Name:        "ComprehensiveTestType",
		Description: stringPtr("Test requirement type for comprehensive testing"),
	}
	err = db.Create(testRequirementType).Error
	require.NoError(t, err)

	return testUser, testEpic, testUserStory, testRequirementType
}

func cleanAllEntityTables(db *gorm.DB) {
	// Clean all entity tables in dependency order
	tables := []string{
		"acceptance_criteria",
		"requirements",
		"user_stories",
		"epics",
	}

	for _, table := range tables {
		db.Exec("DELETE FROM " + table)
	}
}

func cleanupPostgreSQLComprehensiveTest(t *testing.T, db *gorm.DB) {
	// Clean up test data
	tables := []string{
		"acceptance_criteria",
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
