package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ExampleTestReferenceIDGenerator demonstrates how to use the test generator in unit tests
func ExampleTestReferenceIDGenerator() {
	// Create an in-memory SQLite database for testing
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// Create a test generator for Epic entities
	epicGenerator := NewTestReferenceIDGenerator("EP")

	// Generate some reference IDs
	id1, _ := epicGenerator.Generate(db, &Epic{})
	id2, _ := epicGenerator.Generate(db, &Epic{})

	// IDs will be EP-001, EP-002, etc.
	_ = id1 // EP-001
	_ = id2 // EP-002

	// Reset counter for clean test state
	epicGenerator.Reset()

	// Next ID will start from EP-001 again
	id3, _ := epicGenerator.Generate(db, &Epic{})
	_ = id3 // EP-001
}

// TestExampleUsageInUnitTest shows how to use the test generator in actual unit tests
func TestExampleUsageInUnitTest(t *testing.T) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create test generators for different entity types
	epicGen := NewTestReferenceIDGenerator("EP")
	userStoryGen := NewTestReferenceIDGenerator("US")
	reqGen := NewTestReferenceIDGenerator("REQ")
	acGen := NewTestReferenceIDGenerator("AC")

	// Test Epic reference ID generation
	epicID1, err := epicGen.Generate(db, &Epic{})
	require.NoError(t, err)
	assert.Equal(t, "EP-001", epicID1)

	epicID2, err := epicGen.Generate(db, &Epic{})
	require.NoError(t, err)
	assert.Equal(t, "EP-002", epicID2)

	// Test UserStory reference ID generation (independent counter)
	usID1, err := userStoryGen.Generate(db, &UserStory{})
	require.NoError(t, err)
	assert.Equal(t, "US-001", usID1)

	// Test Requirement reference ID generation (independent counter)
	reqID1, err := reqGen.Generate(db, &Requirement{})
	require.NoError(t, err)
	assert.Equal(t, "REQ-001", reqID1)

	// Test AcceptanceCriteria reference ID generation (independent counter)
	acID1, err := acGen.Generate(db, &AcceptanceCriteria{})
	require.NoError(t, err)
	assert.Equal(t, "AC-001", acID1)

	// Verify each generator maintains its own counter
	assert.Equal(t, int64(2), epicGen.GetCounter())
	assert.Equal(t, int64(1), userStoryGen.GetCounter())
	assert.Equal(t, int64(1), reqGen.GetCounter())
	assert.Equal(t, int64(1), acGen.GetCounter())
}

// TestTestGeneratorIsolation demonstrates that test generators are isolated from each other
func TestTestGeneratorIsolation(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create two separate generators with the same prefix
	gen1 := NewTestReferenceIDGenerator("TEST")
	gen2 := NewTestReferenceIDGenerator("TEST")

	// Each generator maintains its own counter
	id1, err := gen1.Generate(db, &Epic{})
	require.NoError(t, err)
	assert.Equal(t, "TEST-001", id1)

	id2, err := gen2.Generate(db, &Epic{})
	require.NoError(t, err)
	assert.Equal(t, "TEST-001", id2) // Same ID because separate counters

	// Verify counters are independent
	assert.Equal(t, int64(1), gen1.GetCounter())
	assert.Equal(t, int64(1), gen2.GetCounter())

	// Reset one generator doesn't affect the other
	gen1.Reset()
	assert.Equal(t, int64(0), gen1.GetCounter())
	assert.Equal(t, int64(1), gen2.GetCounter())
}

// TestTestGeneratorForMocking shows how to use the test generator for mocking scenarios
func TestTestGeneratorForMocking(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	generator := NewTestReferenceIDGenerator("MOCK")

	// Set up specific counter value for testing edge cases
	generator.SetCounter(999)

	// Next ID will be MOCK-1000
	id, err := generator.Generate(db, &Epic{})
	require.NoError(t, err)
	assert.Equal(t, "MOCK-1000", id)

	// Verify counter incremented
	assert.Equal(t, int64(1000), generator.GetCounter())
}
