package models

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
)

// TestReferenceIDGenerator implements a simple reference ID generator for unit tests
// It uses simple counting without advisory locks for fast test execution
type TestReferenceIDGenerator struct {
	prefix  string
	counter int64
	mutex   sync.Mutex // Protects counter for concurrent test scenarios
}

// NewTestReferenceIDGenerator creates a new test reference ID generator
func NewTestReferenceIDGenerator(prefix string) *TestReferenceIDGenerator {
	return &TestReferenceIDGenerator{
		prefix:  prefix,
		counter: 0,
	}
}

// Generate creates a new reference ID using simple counting logic
// This is optimized for fast unit tests and doesn't use database queries
func (g *TestReferenceIDGenerator) Generate(tx *gorm.DB, model interface{}) (string, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	g.counter++
	return fmt.Sprintf("%s-%03d", g.prefix, g.counter), nil
}

// Reset resets the counter to 0 - useful for test isolation
func (g *TestReferenceIDGenerator) Reset() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.counter = 0
}

// GetCount returns the current counter value - useful for test assertions
func (g *TestReferenceIDGenerator) GetCount() int64 {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	return g.counter
}