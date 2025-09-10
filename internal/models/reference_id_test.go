package models

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
)

// TestReferenceIDGenerator implements a simple reference ID generator for unit tests.
// It uses simple counting without advisory locks for fast test execution.
// 
// This generator is only available in test files (files ending with _test.go) and 
// is not included in production builds. It provides predictable, sequential IDs
// for unit testing scenarios where you need consistent, fast reference ID generation.
//
// Key differences from PostgreSQLReferenceIDGenerator:
// - No database dependency for counting (uses internal counter)
// - No advisory locks (not needed for single-threaded unit tests)
// - Thread-safe with mutex for concurrent test scenarios
// - Provides helper methods for test setup (Reset, SetCounter, GetCounter)
//
// Usage:
//   generator := NewTestReferenceIDGenerator("EP")
//   id, err := generator.Generate(db, &Epic{})
//   // id will be "EP-001"
type TestReferenceIDGenerator struct {
	prefix  string    // Entity prefix (EP, US, REQ, AC)
	counter int64     // Internal counter for generating sequential IDs
	mutex   sync.Mutex // Mutex to ensure thread-safety in tests
}

// NewTestReferenceIDGenerator creates a new test reference ID generator
// This constructor is only available in test files
func NewTestReferenceIDGenerator(prefix string) *TestReferenceIDGenerator {
	return &TestReferenceIDGenerator{
		prefix:  prefix,
		counter: 0,
	}
}

// Generate creates a new reference ID using simple counting logic
// This method provides predictable, sequential IDs for unit testing
func (g *TestReferenceIDGenerator) Generate(tx *gorm.DB, model interface{}) (string, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	// Increment counter and generate sequential ID
	g.counter++
	return fmt.Sprintf("%s-%03d", g.prefix, g.counter), nil
}

// Reset resets the internal counter to 0
// This method is useful for test setup and cleanup
func (g *TestReferenceIDGenerator) Reset() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.counter = 0
}

// SetCounter sets the internal counter to a specific value
// This method is useful for testing specific scenarios
func (g *TestReferenceIDGenerator) SetCounter(count int64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.counter = count
}

// GetCounter returns the current counter value
// This method is useful for test assertions
func (g *TestReferenceIDGenerator) GetCounter() int64 {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	return g.counter
}