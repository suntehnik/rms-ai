package helpers

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"
)

// BenchmarkCleanupManager provides comprehensive cleanup and resource management
type BenchmarkCleanupManager struct {
	b                *testing.B
	cleanupTasks     []CleanupTask
	resourceTrackers []ResourceTracker
	mu               sync.RWMutex
	cleanupTimeout   time.Duration
	forceCleanup     bool
}

// CleanupTask represents a single cleanup operation
type CleanupTask struct {
	Name        string
	Description string
	Priority    int // Lower numbers = higher priority
	Timeout     time.Duration
	CleanupFunc func() error
	Required    bool // If true, failure will be logged but won't stop other cleanup
}

// ResourceTracker tracks and manages specific types of resources
type ResourceTracker interface {
	GetResourceType() string
	GetResourceCount() int
	Cleanup() error
	ForceCleanup() error
}

// DatabaseResourceTracker tracks database connections and transactions
type DatabaseResourceTracker struct {
	db           *gorm.DB
	transactions []*gorm.DB
	mu           sync.RWMutex
}

// HTTPResourceTracker tracks HTTP connections and clients
type HTTPResourceTracker struct {
	clients []interface{} // Store HTTP clients or connections
	mu      sync.RWMutex
}

// MemoryResourceTracker tracks memory allocations and triggers cleanup
type MemoryResourceTracker struct {
	initialMemStats runtime.MemStats
	thresholdMB     int64
	mu              sync.RWMutex
}

// NewBenchmarkCleanupManager creates a new cleanup manager
func NewBenchmarkCleanupManager(b *testing.B) *BenchmarkCleanupManager {
	return &BenchmarkCleanupManager{
		b:              b,
		cleanupTasks:   make([]CleanupTask, 0),
		resourceTrackers: make([]ResourceTracker, 0),
		cleanupTimeout: 30 * time.Second,
		forceCleanup:   false,
	}
}

// SetCleanupTimeout sets the maximum time allowed for cleanup operations
func (bcm *BenchmarkCleanupManager) SetCleanupTimeout(timeout time.Duration) {
	bcm.mu.Lock()
	defer bcm.mu.Unlock()
	bcm.cleanupTimeout = timeout
}

// SetForceCleanup enables aggressive cleanup that may impact performance
func (bcm *BenchmarkCleanupManager) SetForceCleanup(force bool) {
	bcm.mu.Lock()
	defer bcm.mu.Unlock()
	bcm.forceCleanup = force
}

// AddCleanupTask registers a cleanup task to be executed during cleanup
func (bcm *BenchmarkCleanupManager) AddCleanupTask(task CleanupTask) {
	bcm.mu.Lock()
	defer bcm.mu.Unlock()
	
	// Set default timeout if not specified
	if task.Timeout == 0 {
		task.Timeout = 10 * time.Second
	}
	
	bcm.cleanupTasks = append(bcm.cleanupTasks, task)
}

// AddResourceTracker registers a resource tracker
func (bcm *BenchmarkCleanupManager) AddResourceTracker(tracker ResourceTracker) {
	bcm.mu.Lock()
	defer bcm.mu.Unlock()
	bcm.resourceTrackers = append(bcm.resourceTrackers, tracker)
}

// ExecuteCleanup performs all registered cleanup operations
func (bcm *BenchmarkCleanupManager) ExecuteCleanup() {
	bcm.b.Logf("Starting benchmark cleanup...")
	
	// Sort cleanup tasks by priority (lower number = higher priority)
	bcm.mu.RLock()
	tasks := make([]CleanupTask, len(bcm.cleanupTasks))
	copy(tasks, bcm.cleanupTasks)
	trackers := make([]ResourceTracker, len(bcm.resourceTrackers))
	copy(trackers, bcm.resourceTrackers)
	bcm.mu.RUnlock()
	
	// Sort tasks by priority
	for i := 0; i < len(tasks)-1; i++ {
		for j := i + 1; j < len(tasks); j++ {
			if tasks[i].Priority > tasks[j].Priority {
				tasks[i], tasks[j] = tasks[j], tasks[i]
			}
		}
	}
	
	// Execute cleanup tasks
	for _, task := range tasks {
		bcm.executeCleanupTask(task)
	}
	
	// Clean up tracked resources
	for _, tracker := range trackers {
		bcm.cleanupResourceTracker(tracker)
	}
	
	// Force garbage collection if requested
	if bcm.forceCleanup {
		bcm.forceGarbageCollection()
	}
	
	bcm.b.Logf("Benchmark cleanup completed")
}

// executeCleanupTask executes a single cleanup task with timeout and error handling
func (bcm *BenchmarkCleanupManager) executeCleanupTask(task CleanupTask) {
	bcm.b.Logf("Executing cleanup task: %s", task.Name)
	
	ctx, cancel := context.WithTimeout(context.Background(), task.Timeout)
	defer cancel()
	
	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("cleanup task panicked: %v", r)
			}
		}()
		done <- task.CleanupFunc()
	}()
	
	select {
	case err := <-done:
		if err != nil {
			if task.Required {
				bcm.b.Logf("Required cleanup task '%s' failed: %v", task.Name, err)
			} else {
				bcm.b.Logf("Optional cleanup task '%s' failed: %v", task.Name, err)
			}
		} else {
			bcm.b.Logf("Cleanup task '%s' completed successfully", task.Name)
		}
	case <-ctx.Done():
		bcm.b.Logf("Cleanup task '%s' timed out after %v", task.Name, task.Timeout)
	}
}

// cleanupResourceTracker cleans up resources managed by a tracker
func (bcm *BenchmarkCleanupManager) cleanupResourceTracker(tracker ResourceTracker) {
	resourceType := tracker.GetResourceType()
	resourceCount := tracker.GetResourceCount()
	
	bcm.b.Logf("Cleaning up %d %s resources", resourceCount, resourceType)
	
	var err error
	if bcm.forceCleanup {
		err = tracker.ForceCleanup()
	} else {
		err = tracker.Cleanup()
	}
	
	if err != nil {
		bcm.b.Logf("Failed to cleanup %s resources: %v", resourceType, err)
	} else {
		bcm.b.Logf("Successfully cleaned up %s resources", resourceType)
	}
}

// forceGarbageCollection performs aggressive garbage collection
func (bcm *BenchmarkCleanupManager) forceGarbageCollection() {
	bcm.b.Logf("Forcing garbage collection...")
	
	var beforeStats, afterStats runtime.MemStats
	runtime.ReadMemStats(&beforeStats)
	
	// Run GC multiple times for thorough cleanup
	runtime.GC()
	runtime.GC()
	runtime.GC()
	
	// Force finalization
	runtime.GC()
	runtime.GC()
	
	runtime.ReadMemStats(&afterStats)
	
	freedMB := float64(beforeStats.HeapAlloc-afterStats.HeapAlloc) / 1024 / 1024
	bcm.b.Logf("Garbage collection freed %.2f MB", freedMB)
}

// GetResourceSummary returns a summary of all tracked resources
func (bcm *BenchmarkCleanupManager) GetResourceSummary() map[string]int {
	bcm.mu.RLock()
	defer bcm.mu.RUnlock()
	
	summary := make(map[string]int)
	for _, tracker := range bcm.resourceTrackers {
		summary[tracker.GetResourceType()] = tracker.GetResourceCount()
	}
	
	return summary
}

// NewDatabaseResourceTracker creates a new database resource tracker
func NewDatabaseResourceTracker(db *gorm.DB) *DatabaseResourceTracker {
	return &DatabaseResourceTracker{
		db:           db,
		transactions: make([]*gorm.DB, 0),
	}
}

// GetResourceType returns the resource type name
func (drt *DatabaseResourceTracker) GetResourceType() string {
	return "database"
}

// GetResourceCount returns the number of tracked database resources
func (drt *DatabaseResourceTracker) GetResourceCount() int {
	drt.mu.RLock()
	defer drt.mu.RUnlock()
	
	count := 0
	if drt.db != nil {
		count++
	}
	count += len(drt.transactions)
	
	return count
}

// TrackTransaction adds a transaction to be tracked and cleaned up
func (drt *DatabaseResourceTracker) TrackTransaction(tx *gorm.DB) {
	drt.mu.Lock()
	defer drt.mu.Unlock()
	drt.transactions = append(drt.transactions, tx)
}

// Cleanup performs graceful cleanup of database resources
func (drt *DatabaseResourceTracker) Cleanup() error {
	drt.mu.Lock()
	defer drt.mu.Unlock()
	
	// Rollback any open transactions
	for i, tx := range drt.transactions {
		if tx != nil {
			if err := tx.Rollback().Error; err != nil {
				// Log but continue with other transactions
				fmt.Printf("Failed to rollback transaction %d: %v\n", i, err)
			}
		}
	}
	
	// Clear transaction list
	drt.transactions = drt.transactions[:0]
	
	// Close database connection if we own it
	if drt.db != nil {
		if sqlDB, err := drt.db.DB(); err == nil {
			return sqlDB.Close()
		}
	}
	
	return nil
}

// ForceCleanup performs aggressive cleanup of database resources
func (drt *DatabaseResourceTracker) ForceCleanup() error {
	// First try graceful cleanup
	if err := drt.Cleanup(); err != nil {
		return err
	}
	
	// Force close any remaining connections
	if drt.db != nil {
		if sqlDB, err := drt.db.DB(); err == nil {
			// Set connection limits to 0 to force closure
			sqlDB.SetMaxOpenConns(0)
			sqlDB.SetMaxIdleConns(0)
			return sqlDB.Close()
		}
	}
	
	return nil
}

// NewHTTPResourceTracker creates a new HTTP resource tracker
func NewHTTPResourceTracker() *HTTPResourceTracker {
	return &HTTPResourceTracker{
		clients: make([]interface{}, 0),
	}
}

// GetResourceType returns the resource type name
func (hrt *HTTPResourceTracker) GetResourceType() string {
	return "http"
}

// GetResourceCount returns the number of tracked HTTP resources
func (hrt *HTTPResourceTracker) GetResourceCount() int {
	hrt.mu.RLock()
	defer hrt.mu.RUnlock()
	return len(hrt.clients)
}

// TrackClient adds an HTTP client to be tracked and cleaned up
func (hrt *HTTPResourceTracker) TrackClient(client interface{}) {
	hrt.mu.Lock()
	defer hrt.mu.Unlock()
	hrt.clients = append(hrt.clients, client)
}

// Cleanup performs graceful cleanup of HTTP resources
func (hrt *HTTPResourceTracker) Cleanup() error {
	hrt.mu.Lock()
	defer hrt.mu.Unlock()
	
	// Clear client list (HTTP clients don't need explicit cleanup in Go)
	hrt.clients = hrt.clients[:0]
	
	return nil
}

// ForceCleanup performs aggressive cleanup of HTTP resources
func (hrt *HTTPResourceTracker) ForceCleanup() error {
	return hrt.Cleanup()
}

// NewMemoryResourceTracker creates a new memory resource tracker
func NewMemoryResourceTracker(thresholdMB int64) *MemoryResourceTracker {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	return &MemoryResourceTracker{
		initialMemStats: memStats,
		thresholdMB:     thresholdMB,
	}
}

// GetResourceType returns the resource type name
func (mrt *MemoryResourceTracker) GetResourceType() string {
	return "memory"
}

// GetResourceCount returns current memory usage in MB
func (mrt *MemoryResourceTracker) GetResourceCount() int {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return int(memStats.Alloc / 1024 / 1024)
}

// Cleanup performs graceful memory cleanup
func (mrt *MemoryResourceTracker) Cleanup() error {
	runtime.GC()
	return nil
}

// ForceCleanup performs aggressive memory cleanup
func (mrt *MemoryResourceTracker) ForceCleanup() error {
	// Multiple GC cycles for thorough cleanup
	for i := 0; i < 3; i++ {
		runtime.GC()
	}
	
	// Force finalization
	runtime.GC()
	
	return nil
}

// IsMemoryThresholdExceeded checks if memory usage exceeds the threshold
func (mrt *MemoryResourceTracker) IsMemoryThresholdExceeded() bool {
	currentMB := int64(mrt.GetResourceCount())
	return currentMB > mrt.thresholdMB
}

// GetMemoryUsageDelta returns the change in memory usage since initialization
func (mrt *MemoryResourceTracker) GetMemoryUsageDelta() int64 {
	var currentStats runtime.MemStats
	runtime.ReadMemStats(&currentStats)
	
	deltaMB := int64(currentStats.Alloc-mrt.initialMemStats.Alloc) / 1024 / 1024
	return deltaMB
}

// BenchmarkResourceManager provides high-level resource management for benchmarks
type BenchmarkResourceManager struct {
	cleanupManager *BenchmarkCleanupManager
	dbTracker      *DatabaseResourceTracker
	httpTracker    *HTTPResourceTracker
	memoryTracker  *MemoryResourceTracker
	b              *testing.B
}

// NewBenchmarkResourceManager creates a comprehensive resource manager
func NewBenchmarkResourceManager(b *testing.B, db *gorm.DB) *BenchmarkResourceManager {
	cleanupManager := NewBenchmarkCleanupManager(b)
	dbTracker := NewDatabaseResourceTracker(db)
	httpTracker := NewHTTPResourceTracker()
	memoryTracker := NewMemoryResourceTracker(512) // 512MB threshold
	
	// Register trackers with cleanup manager
	cleanupManager.AddResourceTracker(dbTracker)
	cleanupManager.AddResourceTracker(httpTracker)
	cleanupManager.AddResourceTracker(memoryTracker)
	
	return &BenchmarkResourceManager{
		cleanupManager: cleanupManager,
		dbTracker:      dbTracker,
		httpTracker:    httpTracker,
		memoryTracker:  memoryTracker,
		b:              b,
	}
}

// RegisterDatabaseCleanup registers database-specific cleanup tasks
func (brm *BenchmarkResourceManager) RegisterDatabaseCleanup(db *gorm.DB) {
	// Clean up test data
	brm.cleanupManager.AddCleanupTask(CleanupTask{
		Name:        "database_data_cleanup",
		Description: "Clean up test data from database",
		Priority:    1,
		Timeout:     30 * time.Second,
		Required:    false,
		CleanupFunc: func() error {
			return brm.cleanupTestData(db)
		},
	})
	
	// Reset sequences
	brm.cleanupManager.AddCleanupTask(CleanupTask{
		Name:        "database_sequence_reset",
		Description: "Reset database sequences",
		Priority:    2,
		Timeout:     10 * time.Second,
		Required:    false,
		CleanupFunc: func() error {
			return brm.resetDatabaseSequences(db)
		},
	})
}

// RegisterServerCleanup registers server-specific cleanup tasks
func (brm *BenchmarkResourceManager) RegisterServerCleanup(server interface{}) {
	brm.cleanupManager.AddCleanupTask(CleanupTask{
		Name:        "server_shutdown",
		Description: "Shutdown benchmark server",
		Priority:    0, // Highest priority
		Timeout:     15 * time.Second,
		Required:    true,
		CleanupFunc: func() error {
			return brm.shutdownServer(server)
		},
	})
}

// RegisterContainerCleanup registers container-specific cleanup tasks
func (brm *BenchmarkResourceManager) RegisterContainerCleanup(container interface{}) {
	brm.cleanupManager.AddCleanupTask(CleanupTask{
		Name:        "container_cleanup",
		Description: "Terminate test containers",
		Priority:    3,
		Timeout:     20 * time.Second,
		Required:    false,
		CleanupFunc: func() error {
			return brm.terminateContainer(container)
		},
	})
}

// ExecuteCleanup performs all registered cleanup operations
func (brm *BenchmarkResourceManager) ExecuteCleanup() {
	brm.cleanupManager.ExecuteCleanup()
}

// GetResourceSummary returns a summary of all managed resources
func (brm *BenchmarkResourceManager) GetResourceSummary() map[string]interface{} {
	summary := make(map[string]interface{})
	
	// Basic resource counts
	resourceCounts := brm.cleanupManager.GetResourceSummary()
	summary["resource_counts"] = resourceCounts
	
	// Memory usage details
	summary["memory_usage_mb"] = brm.memoryTracker.GetResourceCount()
	summary["memory_delta_mb"] = brm.memoryTracker.GetMemoryUsageDelta()
	summary["memory_threshold_exceeded"] = brm.memoryTracker.IsMemoryThresholdExceeded()
	
	// Runtime information
	summary["goroutine_count"] = runtime.NumGoroutine()
	summary["gc_count"] = getGCCount()
	
	return summary
}

// cleanupTestData removes test data from the database
func (brm *BenchmarkResourceManager) cleanupTestData(db *gorm.DB) error {
	// Delete in reverse dependency order to avoid foreign key constraints
	tables := []string{
		"comments",
		"acceptance_criteria",
		"requirements",
		"user_stories",
		"epics",
		// Don't delete users as they might be needed for other tests
	}
	
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE created_at > NOW() - INTERVAL '1 hour'", table)).Error; err != nil {
			brm.b.Logf("Failed to cleanup table %s: %v", table, err)
		}
	}
	
	return nil
}

// resetDatabaseSequences resets auto-increment sequences
func (brm *BenchmarkResourceManager) resetDatabaseSequences(db *gorm.DB) error {
	// This is PostgreSQL-specific - adjust for other databases
	sequences := []string{
		"epics_id_seq",
		"user_stories_id_seq",
		"requirements_id_seq",
		"acceptance_criteria_id_seq",
	}
	
	for _, seq := range sequences {
		query := fmt.Sprintf("SELECT setval('%s', 1, false)", seq)
		if err := db.Exec(query).Error; err != nil {
			// Log but don't fail - sequences might not exist
			brm.b.Logf("Failed to reset sequence %s: %v", seq, err)
		}
	}
	
	return nil
}

// shutdownServer gracefully shuts down the benchmark server
func (brm *BenchmarkResourceManager) shutdownServer(server interface{}) error {
	// Type assertion and shutdown logic would go here
	// This is a placeholder for the actual server shutdown
	brm.b.Logf("Shutting down benchmark server")
	return nil
}

// terminateContainer terminates test containers
func (brm *BenchmarkResourceManager) terminateContainer(container interface{}) error {
	// Container termination logic would go here
	// This is a placeholder for the actual container termination
	brm.b.Logf("Terminating test container")
	return nil
}

// getGCCount returns the current garbage collection count
func getGCCount() uint32 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats.NumGC
}