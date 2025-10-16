package models

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTestReferenceIDGenerator_Interface(t *testing.T) {
	// Verify that TestReferenceIDGenerator implements ReferenceIDGenerator interface
	var _ ReferenceIDGenerator = (*TestReferenceIDGenerator)(nil)
}

func TestNewTestReferenceIDGenerator(t *testing.T) {
	generator := NewTestReferenceIDGenerator("TEST")

	assert.NotNil(t, generator)
	assert.Equal(t, "TEST", generator.prefix)
	assert.Equal(t, int64(0), generator.counter)
}

func TestTestReferenceIDGenerator_Generate(t *testing.T) {
	// Setup in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	generator := NewTestReferenceIDGenerator("TEST")

	// Test sequential ID generation
	id1, err := generator.Generate(db, &Epic{})
	require.NoError(t, err)
	assert.Equal(t, "TEST-001", id1)

	id2, err := generator.Generate(db, &Epic{})
	require.NoError(t, err)
	assert.Equal(t, "TEST-002", id2)

	id3, err := generator.Generate(db, &Epic{})
	require.NoError(t, err)
	assert.Equal(t, "TEST-003", id3)
}

func TestTestReferenceIDGenerator_DifferentPrefixes(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	epicGen := NewTestReferenceIDGenerator("EP")
	userStoryGen := NewTestReferenceIDGenerator("US")
	reqGen := NewTestReferenceIDGenerator("REQ")
	acGen := NewTestReferenceIDGenerator("AC")

	// Test different prefixes generate correctly
	epicID, err := epicGen.Generate(db, &Epic{})
	require.NoError(t, err)
	assert.Equal(t, "EP-001", epicID)

	usID, err := userStoryGen.Generate(db, &UserStory{})
	require.NoError(t, err)
	assert.Equal(t, "US-001", usID)

	reqID, err := reqGen.Generate(db, &Requirement{})
	require.NoError(t, err)
	assert.Equal(t, "REQ-001", reqID)

	acID, err := acGen.Generate(db, &AcceptanceCriteria{})
	require.NoError(t, err)
	assert.Equal(t, "AC-001", acID)
}

func TestTestReferenceIDGenerator_Reset(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	generator := NewTestReferenceIDGenerator("TEST")

	// Generate some IDs
	_, err = generator.Generate(db, &Epic{})
	require.NoError(t, err)
	_, err = generator.Generate(db, &Epic{})
	require.NoError(t, err)

	assert.Equal(t, int64(2), generator.GetCounter())

	// Reset and verify counter is back to 0
	generator.Reset()
	assert.Equal(t, int64(0), generator.GetCounter())

	// Next ID should start from 001 again
	id, err := generator.Generate(db, &Epic{})
	require.NoError(t, err)
	assert.Equal(t, "TEST-001", id)
}

func TestTestReferenceIDGenerator_SetCounter(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	generator := NewTestReferenceIDGenerator("TEST")

	// Set counter to specific value
	generator.SetCounter(10)
	assert.Equal(t, int64(10), generator.GetCounter())

	// Next ID should be 011
	id, err := generator.Generate(db, &Epic{})
	require.NoError(t, err)
	assert.Equal(t, "TEST-011", id)
}

func TestTestReferenceIDGenerator_ConcurrentAccess(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	generator := NewTestReferenceIDGenerator("TEST")

	const numGoroutines = 10
	const idsPerGoroutine = 10

	var wg sync.WaitGroup
	results := make([][]string, numGoroutines)

	// Launch multiple goroutines to generate IDs concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineIndex int) {
			defer wg.Done()

			ids := make([]string, idsPerGoroutine)
			for j := 0; j < idsPerGoroutine; j++ {
				id, err := generator.Generate(db, &Epic{})
				require.NoError(t, err)
				ids[j] = id
			}
			results[goroutineIndex] = ids
		}(i)
	}

	wg.Wait()

	// Collect all generated IDs
	allIDs := make([]string, 0, numGoroutines*idsPerGoroutine)
	for _, goroutineResults := range results {
		allIDs = append(allIDs, goroutineResults...)
	}

	// Verify all IDs are unique
	idSet := make(map[string]bool)
	for _, id := range allIDs {
		assert.False(t, idSet[id], "Duplicate ID found: %s", id)
		idSet[id] = true
	}

	// Verify we have the expected number of unique IDs
	assert.Equal(t, numGoroutines*idsPerGoroutine, len(idSet))

	// Verify final counter value
	assert.Equal(t, int64(numGoroutines*idsPerGoroutine), generator.GetCounter())
}

func TestTestReferenceIDGenerator_CompatibilityWithProductionGenerator(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create the Epic table for testing
	err = db.AutoMigrate(&Epic{})
	require.NoError(t, err)

	// Test that both generators implement the same interface
	var testGen ReferenceIDGenerator = NewTestReferenceIDGenerator("TEST")
	var prodGen ReferenceIDGenerator = NewPostgreSQLReferenceIDGenerator(123456, "PROD")

	// Both should be able to generate IDs
	testID, err := testGen.Generate(db, &Epic{})
	require.NoError(t, err)
	assert.Equal(t, "TEST-001", testID)

	prodID, err := prodGen.Generate(db, &Epic{})
	require.NoError(t, err)
	assert.Equal(t, "PROD-001", prodID) // SQLite mode for production generator
}
