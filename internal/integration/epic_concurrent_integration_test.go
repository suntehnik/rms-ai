package integration

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"product-requirements-management/internal/models"
)

func TestEpicConcurrentCreation_PostgreSQL(t *testing.T) {
	// Setup PostgreSQL test database
	testDB := SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	// Reset database to ensure clean state
	err := testDB.Reset()
	require.NoError(t, err)

	// Create test user
	user := testDB.CreateTestUser(t)

	t.Run("concurrent_epic_creation_with_advisory_locks", func(t *testing.T) {
		const numGoroutines = 10
		const numEpicsPerGoroutine = 5
		
		var wg sync.WaitGroup
		var mu sync.Mutex
		var createdEpics []models.Epic
		var errors []error

		// Channel to collect results
		results := make(chan struct {
			epic models.Epic
			err  error
		}, numGoroutines*numEpicsPerGoroutine)

		wg.Add(numGoroutines)

		// Launch concurrent goroutines
		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer wg.Done()

				// Create separate database connection for each goroutine
				// This simulates real concurrent access from multiple application instances
				db, err := gorm.Open(postgres.Open(testDB.DSN), &gorm.Config{
					Logger: logger.Default.LogMode(logger.Silent),
					NowFunc: func() time.Time {
						return time.Now().UTC()
					},
				})
				if err != nil {
					results <- struct {
						epic models.Epic
						err  error
					}{models.Epic{}, fmt.Errorf("failed to create DB connection: %w", err)}
					return
				}

				// Create multiple epics in this goroutine
				for j := 0; j < numEpicsPerGoroutine; j++ {
					epic := models.Epic{
						CreatorID:   user.ID,
						AssigneeID:  user.ID,
						Priority:    models.PriorityHigh,
						Title:       fmt.Sprintf("Concurrent Epic G%d-E%d", goroutineID, j),
						Description: stringPtr(fmt.Sprintf("Epic created by goroutine %d, epic %d", goroutineID, j)),
					}

					// Create epic - this should trigger BeforeCreate hook with advisory lock
					err := db.Create(&epic).Error
					results <- struct {
						epic models.Epic
						err  error
					}{epic, err}
				}
			}(i)
		}

		// Wait for all goroutines to complete
		wg.Wait()
		close(results)

		// Collect results
		for result := range results {
			mu.Lock()
			if result.err != nil {
				errors = append(errors, result.err)
			} else {
				createdEpics = append(createdEpics, result.epic)
			}
			mu.Unlock()
		}

		// Verify results
		t.Logf("Created %d epics, encountered %d errors", len(createdEpics), len(errors))
		
		// All epics should be created successfully
		assert.Empty(t, errors, "No errors should occur during concurrent creation")
		assert.Len(t, createdEpics, numGoroutines*numEpicsPerGoroutine, "All epics should be created")

		// Verify all reference IDs are unique
		referenceIDs := make(map[string]bool)
		sequentialCount := 0
		uuidBasedCount := 0

		for _, epic := range createdEpics {
			assert.NotEmpty(t, epic.ReferenceID, "Reference ID should not be empty")
			assert.False(t, referenceIDs[epic.ReferenceID], "Reference ID should be unique: %s", epic.ReferenceID)
			referenceIDs[epic.ReferenceID] = true

			// Check if it's sequential (EP-001, EP-002, etc.) or UUID-based (EP-xxxxxxxx)
			if len(epic.ReferenceID) == 6 && epic.ReferenceID[:3] == "EP-" {
				sequentialCount++
			} else if len(epic.ReferenceID) == 11 && epic.ReferenceID[:3] == "EP-" {
				uuidBasedCount++
			}
		}

		t.Logf("Sequential reference IDs: %d, UUID-based reference IDs: %d", sequentialCount, uuidBasedCount)
		
		// We expect most to be sequential (when advisory lock is acquired)
		// Some might be UUID-based (when lock is not acquired due to contention)
		assert.True(t, sequentialCount > 0, "Should have some sequential reference IDs")
		
		// Verify database consistency
		var dbCount int64
		err = testDB.DB.Model(&models.Epic{}).Count(&dbCount).Error
		require.NoError(t, err)
		assert.Equal(t, int64(len(createdEpics)), dbCount, "Database count should match created epics")
	})

	t.Run("concurrent_epic_creation_stress_test", func(t *testing.T) {
		// Reset database
		err := testDB.Reset()
		require.NoError(t, err)

		// Create fresh test user
		user := testDB.CreateTestUser(t)

		const numGoroutines = 20
		const numEpicsPerGoroutine = 3
		
		var wg sync.WaitGroup
		var mu sync.Mutex
		var allErrors []error
		var successCount int

		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer wg.Done()

				// Create separate database connection
				db, err := gorm.Open(postgres.Open(testDB.DSN), &gorm.Config{
					Logger: logger.Default.LogMode(logger.Silent),
				})
				if err != nil {
					mu.Lock()
					allErrors = append(allErrors, fmt.Errorf("goroutine %d: failed to create DB connection: %w", goroutineID, err))
					mu.Unlock()
					return
				}

				for j := 0; j < numEpicsPerGoroutine; j++ {
					epic := models.Epic{
						CreatorID:   user.ID,
						AssigneeID:  user.ID,
						Priority:    models.PriorityMedium,
						Title:       fmt.Sprintf("Stress Test Epic G%d-E%d", goroutineID, j),
						Description: stringPtr("Stress test epic"),
					}

					err := db.Create(&epic).Error
					mu.Lock()
					if err != nil {
						allErrors = append(allErrors, fmt.Errorf("goroutine %d, epic %d: %w", goroutineID, j, err))
					} else {
						successCount++
					}
					mu.Unlock()

					// Small delay to increase contention
					time.Sleep(1 * time.Millisecond)
				}
			}(i)
		}

		wg.Wait()

		t.Logf("Stress test completed: %d successes, %d errors", successCount, len(allErrors))
		
		// Log errors for debugging
		for _, err := range allErrors {
			t.Logf("Error: %v", err)
		}

		// All operations should succeed
		assert.Empty(t, allErrors, "No errors should occur during stress test")
		assert.Equal(t, numGoroutines*numEpicsPerGoroutine, successCount, "All epics should be created successfully")

		// Verify final database state
		var finalCount int64
		err = testDB.DB.Model(&models.Epic{}).Count(&finalCount).Error
		require.NoError(t, err)
		assert.Equal(t, int64(successCount), finalCount, "Database count should match successful creations")
	})

	t.Run("concurrent_epic_creation_with_transaction_rollback", func(t *testing.T) {
		// Reset database
		err := testDB.Reset()
		require.NoError(t, err)

		// Create fresh test user
		user := testDB.CreateTestUser(t)

		const numGoroutines = 5
		
		var wg sync.WaitGroup
		var mu sync.Mutex
		var results []struct {
			success bool
			err     error
		}

		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer wg.Done()

				// Create separate database connection
				db, err := gorm.Open(postgres.Open(testDB.DSN), &gorm.Config{
					Logger: logger.Default.LogMode(logger.Silent),
				})
				if err != nil {
					mu.Lock()
					results = append(results, struct {
						success bool
						err     error
					}{false, err})
					mu.Unlock()
					return
				}

				// Start transaction
				tx := db.Begin()
				defer func() {
					if r := recover(); r != nil {
						tx.Rollback()
					}
				}()

				epic := models.Epic{
					CreatorID:   user.ID,
					AssigneeID:  user.ID,
					Priority:    models.PriorityLow,
					Title:       fmt.Sprintf("Transaction Test Epic %d", goroutineID),
					Description: stringPtr("Transaction test epic"),
				}

				// Create epic within transaction
				err = tx.Create(&epic).Error
				if err != nil {
					tx.Rollback()
					mu.Lock()
					results = append(results, struct {
						success bool
						err     error
					}{false, err})
					mu.Unlock()
					return
				}

				// Simulate some processing time
				time.Sleep(10 * time.Millisecond)

				// Commit transaction
				err = tx.Commit().Error
				mu.Lock()
				results = append(results, struct {
					success bool
					err     error
				}{err == nil, err})
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		// Verify results
		successCount := 0
		for _, result := range results {
			if result.success {
				successCount++
			} else if result.err != nil {
				t.Logf("Transaction error: %v", result.err)
			}
		}

		t.Logf("Transaction test: %d successes out of %d attempts", successCount, len(results))
		
		// All transactions should succeed
		assert.Equal(t, numGoroutines, successCount, "All transactions should succeed")

		// Verify database state
		var dbCount int64
		err = testDB.DB.Model(&models.Epic{}).Count(&dbCount).Error
		require.NoError(t, err)
		assert.Equal(t, int64(successCount), dbCount, "Database should contain all committed epics")
	})
}

func TestEpicConcurrentCreation_ReferenceIDGeneration(t *testing.T) {
	// Setup PostgreSQL test database
	testDB := SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	// Reset database to ensure clean state
	err := testDB.Reset()
	require.NoError(t, err)

	// Create test user
	user := testDB.CreateTestUser(t)

	t.Run("reference_id_uniqueness_under_high_concurrency", func(t *testing.T) {
		const numGoroutines = 15
		const numEpicsPerGoroutine = 2
		
		var wg sync.WaitGroup
		referenceIDs := make(chan string, numGoroutines*numEpicsPerGoroutine)

		wg.Add(numGoroutines)

		// Launch concurrent epic creation
		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer wg.Done()

				// Create separate database connection
				db, err := gorm.Open(postgres.Open(testDB.DSN), &gorm.Config{
					Logger: logger.Default.LogMode(logger.Silent),
				})
				if err != nil {
					t.Errorf("Failed to create DB connection: %v", err)
					return
				}

				for j := 0; j < numEpicsPerGoroutine; j++ {
					epic := models.Epic{
						CreatorID:   user.ID,
						AssigneeID:  user.ID,
						Priority:    models.PriorityCritical,
						Title:       fmt.Sprintf("Uniqueness Test Epic G%d-E%d", goroutineID, j),
						Description: stringPtr("Reference ID uniqueness test"),
					}

					err := db.Create(&epic).Error
					if err != nil {
						t.Errorf("Failed to create epic: %v", err)
						continue
					}

					referenceIDs <- epic.ReferenceID
				}
			}(i)
		}

		wg.Wait()
		close(referenceIDs)

		// Collect and verify uniqueness
		seenIDs := make(map[string]bool)
		duplicates := make([]string, 0)
		totalCount := 0

		for refID := range referenceIDs {
			totalCount++
			if seenIDs[refID] {
				duplicates = append(duplicates, refID)
			} else {
				seenIDs[refID] = true
			}
		}

		t.Logf("Total reference IDs generated: %d", totalCount)
		t.Logf("Unique reference IDs: %d", len(seenIDs))
		t.Logf("Duplicates found: %v", duplicates)

		assert.Empty(t, duplicates, "No duplicate reference IDs should be generated")
		assert.Equal(t, numGoroutines*numEpicsPerGoroutine, len(seenIDs), "All reference IDs should be unique")
		assert.Equal(t, numGoroutines*numEpicsPerGoroutine, totalCount, "All epics should be created")
	})

	t.Run("reference_id_format_validation", func(t *testing.T) {
		// Reset database
		err := testDB.Reset()
		require.NoError(t, err)

		// Create fresh test user
		user := testDB.CreateTestUser(t)

		const numEpics = 10
		var createdEpics []models.Epic

		// Create epics sequentially first to establish baseline
		for i := 0; i < numEpics; i++ {
			epic := models.Epic{
				CreatorID:   user.ID,
				AssigneeID:  user.ID,
				Priority:    models.PriorityHigh,
				Title:       fmt.Sprintf("Format Test Epic %d", i+1),
				Description: stringPtr("Format validation test"),
			}

			err := testDB.DB.Create(&epic).Error
			require.NoError(t, err)
			createdEpics = append(createdEpics, epic)
		}

		// Verify reference ID formats
		for i, epic := range createdEpics {
			assert.NotEmpty(t, epic.ReferenceID, "Reference ID should not be empty")
			assert.True(t, 
				len(epic.ReferenceID) == 6 || len(epic.ReferenceID) == 11, 
				"Reference ID should be either EP-XXX (6 chars) or EP-XXXXXXXX (11 chars), got: %s", 
				epic.ReferenceID)
			assert.True(t, 
				epic.ReferenceID[:3] == "EP-", 
				"Reference ID should start with 'EP-', got: %s", 
				epic.ReferenceID)

			t.Logf("Epic %d: ID=%s, ReferenceID=%s", i+1, epic.ID, epic.ReferenceID)
		}
	})
}

func TestEpicConcurrentCreation_ErrorHandling(t *testing.T) {
	// Setup PostgreSQL test database
	testDB := SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	// Reset database to ensure clean state
	err := testDB.Reset()
	require.NoError(t, err)

	// Create test user
	user := testDB.CreateTestUser(t)

	t.Run("database_connection_failure_handling", func(t *testing.T) {
		// Test with invalid DSN to simulate connection failures
		invalidDSN := "postgres://invalid:invalid@localhost:9999/invalid?sslmode=disable"
		
		db, err := gorm.Open(postgres.Open(invalidDSN), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		
		// Connection might be established but queries will fail
		if err == nil {
			epic := models.Epic{
				CreatorID:   user.ID,
				AssigneeID:  user.ID,
				Priority:    models.PriorityHigh,
				Title:       "Connection Test Epic",
				Description: stringPtr("Testing connection failure"),
			}

			err = db.Create(&epic).Error
			assert.Error(t, err, "Should fail with invalid database connection")
			t.Logf("Expected connection error: %v", err)
		} else {
			t.Logf("Expected connection establishment error: %v", err)
		}
	})

	t.Run("constraint_violation_handling", func(t *testing.T) {
		// Create epic with specific reference ID
		epic1 := models.Epic{
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityHigh,
			Title:       "First Epic",
			Description: stringPtr("First epic"),
			ReferenceID: "EP-DUPLICATE", // Set explicit reference ID
		}

		err := testDB.DB.Create(&epic1).Error
		require.NoError(t, err)

		// Try to create another epic with the same reference ID
		epic2 := models.Epic{
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityHigh,
			Title:       "Second Epic",
			Description: stringPtr("Second epic"),
			ReferenceID: "EP-DUPLICATE", // Same reference ID - should fail
		}

		err = testDB.DB.Create(&epic2).Error
		assert.Error(t, err, "Should fail with duplicate reference ID constraint violation")
		t.Logf("Expected constraint violation error: %v", err)

		// Verify only one epic exists
		var count int64
		err = testDB.DB.Model(&models.Epic{}).Where("reference_id = ?", "EP-DUPLICATE").Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "Only one epic should exist with the duplicate reference ID")
	})
}

