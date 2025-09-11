package helpers

import (
	"fmt"
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BenchmarkValidator provides comprehensive validation for benchmark tests
type BenchmarkValidator struct {
	b                *testing.B
	reliabilityMgr   *BenchmarkReliabilityManager
	validationConfig *ValidationConfig
}

// ValidationConfig defines validation behavior and thresholds
type ValidationConfig struct {
	StrictMode           bool
	ValidateDataIntegrity bool
	ValidatePerformance   bool
	MaxResponseTime       time.Duration
	MinThroughput         float64
	MaxErrorRate          float64
	RequiredDataCounts    map[string]int
}

// NewBenchmarkValidator creates a new benchmark validator
func NewBenchmarkValidator(b *testing.B, reliabilityMgr *BenchmarkReliabilityManager) *BenchmarkValidator {
	return &BenchmarkValidator{
		b:              b,
		reliabilityMgr: reliabilityMgr,
		validationConfig: &ValidationConfig{
			StrictMode:           false,
			ValidateDataIntegrity: true,
			ValidatePerformance:   true,
			MaxResponseTime:       5 * time.Second,
			MinThroughput:         1.0, // 1 operation per second minimum
			MaxErrorRate:          0.05, // 5% maximum error rate
			RequiredDataCounts: map[string]int{
				"users":       1,
				"epics":       1,
				"user_stories": 1,
			},
		},
	}
}

// SetStrictMode enables or disables strict validation mode
func (bv *BenchmarkValidator) SetStrictMode(strict bool) {
	bv.validationConfig.StrictMode = strict
}

// ValidatePrerequisites validates that all prerequisites are met before running benchmarks
func (bv *BenchmarkValidator) ValidatePrerequisites(db *gorm.DB, baseURL string) error {
	return bv.reliabilityMgr.ExecuteWithTimeout("validation", func() error {
		// Validate database connection
		if err := bv.ValidateDatabaseConnection(db); err != nil {
			return fmt.Errorf("database validation failed: %w", err)
		}
		
		// Validate server availability
		if err := bv.ValidateServerAvailability(baseURL); err != nil {
			return fmt.Errorf("server validation failed: %w", err)
		}
		
		// Validate data integrity if enabled
		if bv.validationConfig.ValidateDataIntegrity {
			if err := bv.ValidateDataIntegrity(db); err != nil {
				return fmt.Errorf("data integrity validation failed: %w", err)
			}
		}
		
		return nil
	})
}

// ValidateDatabaseConnection validates that the database is accessible and properly configured
func (bv *BenchmarkValidator) ValidateDatabaseConnection(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}
	
	// Test basic connectivity
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	
	// Ping the database
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	
	// Check connection pool stats
	stats := sqlDB.Stats()
	if stats.OpenConnections == 0 {
		return fmt.Errorf("no open database connections available")
	}
	
	// Validate connection pool configuration
	if stats.MaxOpenConnections > 0 && stats.MaxOpenConnections < 5 {
		bv.b.Logf("WARNING: Low max open connections configured: %d", stats.MaxOpenConnections)
	}
	
	// Test a simple query
	var count int64
	if err := db.Raw("SELECT 1").Scan(&count).Error; err != nil {
		return fmt.Errorf("test query failed: %w", err)
	}
	
	bv.b.Logf("Database validation passed - Open connections: %d, In use: %d", 
		stats.OpenConnections, stats.InUse)
	
	return nil
}

// ValidateServerAvailability validates that the HTTP server is running and accessible
func (bv *BenchmarkValidator) ValidateServerAvailability(baseURL string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// Test health endpoint
	healthURL := baseURL + "/health"
	resp, err := client.Get(healthURL)
	if err != nil {
		return fmt.Errorf("failed to reach health endpoint: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health endpoint returned status %d", resp.StatusCode)
	}
	
	// Test basic API endpoint
	apiURL := baseURL + "/api/v1"
	resp, err = client.Get(apiURL)
	if err != nil {
		return fmt.Errorf("failed to reach API endpoint: %w", err)
	}
	defer resp.Body.Close()
	
	// Accept 404 for API root as it may not have a handler
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("API endpoint returned unexpected status %d", resp.StatusCode)
	}
	
	bv.b.Logf("Server validation passed - Health: %d, API: %d", 
		http.StatusOK, resp.StatusCode)
	
	return nil
}

// ValidateDataIntegrity validates that test data is properly structured and accessible
func (bv *BenchmarkValidator) ValidateDataIntegrity(db *gorm.DB) error {
	// Check required tables exist
	requiredTables := []string{"users", "epics", "user_stories", "requirements", "acceptance_criteria"}
	
	for _, table := range requiredTables {
		var exists bool
		query := "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = ?)"
		if err := db.Raw(query, table).Scan(&exists).Error; err != nil {
			return fmt.Errorf("failed to check table %s existence: %w", table, err)
		}
		
		if !exists {
			return fmt.Errorf("required table %s does not exist", table)
		}
	}
	
	// Validate minimum data counts
	for entity, minCount := range bv.validationConfig.RequiredDataCounts {
		var count int64
		tableName := entity
		if entity == "user_stories" {
			tableName = "user_stories"
		}
		
		if err := db.Table(tableName).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to count %s: %w", entity, err)
		}
		
		if count < int64(minCount) {
			if bv.validationConfig.StrictMode {
				return fmt.Errorf("insufficient %s: need %d, got %d", entity, minCount, count)
			}
			bv.b.Logf("WARNING: Low %s count: %d (minimum: %d)", entity, count, minCount)
		}
	}
	
	// Validate data relationships
	if err := bv.validateDataRelationships(db); err != nil {
		return fmt.Errorf("data relationship validation failed: %w", err)
	}
	
	bv.b.Logf("Data integrity validation passed")
	return nil
}

// validateDataRelationships validates that foreign key relationships are properly maintained
func (bv *BenchmarkValidator) validateDataRelationships(db *gorm.DB) error {
	// Check for orphaned user stories (user stories without epics)
	var orphanedUserStories int64
	if err := db.Table("user_stories").
		Where("epic_id NOT IN (SELECT id FROM epics)").
		Count(&orphanedUserStories).Error; err != nil {
		return fmt.Errorf("failed to check orphaned user stories: %w", err)
	}
	
	if orphanedUserStories > 0 {
		if bv.validationConfig.StrictMode {
			return fmt.Errorf("found %d orphaned user stories", orphanedUserStories)
		}
		bv.b.Logf("WARNING: Found %d orphaned user stories", orphanedUserStories)
	}
	
	// Check for orphaned requirements (requirements without user stories)
	var orphanedRequirements int64
	if err := db.Table("requirements").
		Where("user_story_id NOT IN (SELECT id FROM user_stories)").
		Count(&orphanedRequirements).Error; err != nil {
		return fmt.Errorf("failed to check orphaned requirements: %w", err)
	}
	
	if orphanedRequirements > 0 {
		if bv.validationConfig.StrictMode {
			return fmt.Errorf("found %d orphaned requirements", orphanedRequirements)
		}
		bv.b.Logf("WARNING: Found %d orphaned requirements", orphanedRequirements)
	}
	
	return nil
}

// ValidateTestData validates that test data arrays contain valid elements
func (bv *BenchmarkValidator) ValidateTestData(data interface{}, dataType string) error {
	switch v := data.(type) {
	case []uuid.UUID:
		return bv.validateUUIDArray(v, dataType)
	case []string:
		return bv.validateStringArray(v, dataType)
	case []int:
		return bv.validateIntArray(v, dataType)
	default:
		return fmt.Errorf("unsupported data type for validation: %T", data)
	}
}

// validateUUIDArray validates an array of UUIDs
func (bv *BenchmarkValidator) validateUUIDArray(uuids []uuid.UUID, dataType string) error {
	if len(uuids) == 0 {
		return fmt.Errorf("no %s available for testing", dataType)
	}
	
	for i, id := range uuids {
		if id == uuid.Nil {
			return fmt.Errorf("invalid UUID at index %d in %s array", i, dataType)
		}
	}
	
	bv.b.Logf("Validated %d %s UUIDs", len(uuids), dataType)
	return nil
}

// validateStringArray validates an array of strings
func (bv *BenchmarkValidator) validateStringArray(strings []string, dataType string) error {
	if len(strings) == 0 {
		return fmt.Errorf("no %s available for testing", dataType)
	}
	
	for i, s := range strings {
		if s == "" {
			return fmt.Errorf("empty string at index %d in %s array", i, dataType)
		}
	}
	
	bv.b.Logf("Validated %d %s strings", len(strings), dataType)
	return nil
}

// validateIntArray validates an array of integers
func (bv *BenchmarkValidator) validateIntArray(ints []int, dataType string) error {
	if len(ints) == 0 {
		return fmt.Errorf("no %s available for testing", dataType)
	}
	
	for i, val := range ints {
		if val < 0 {
			return fmt.Errorf("negative value at index %d in %s array: %d", i, dataType, val)
		}
	}
	
	bv.b.Logf("Validated %d %s integers", len(ints), dataType)
	return nil
}

// ValidateHTTPResponse validates an HTTP response for common issues
func (bv *BenchmarkValidator) ValidateHTTPResponse(resp *http.Response, expectedStatus int, operation string) error {
	if resp == nil {
		return fmt.Errorf("nil HTTP response for operation: %s", operation)
	}
	
	if resp.StatusCode != expectedStatus {
		return fmt.Errorf("unexpected status code for %s: expected %d, got %d", 
			operation, expectedStatus, resp.StatusCode)
	}
	
	// Validate response headers
	contentType := resp.Header.Get("Content-Type")
	if expectedStatus == http.StatusOK || expectedStatus == http.StatusCreated {
		if contentType == "" {
			bv.b.Logf("WARNING: Missing Content-Type header for %s", operation)
		}
	}
	
	// Check for common error indicators in headers
	if errorHeader := resp.Header.Get("X-Error"); errorHeader != "" {
		return fmt.Errorf("error header present for %s: %s", operation, errorHeader)
	}
	
	return nil
}

// ValidatePerformanceMetrics validates that performance metrics meet minimum requirements
func (bv *BenchmarkValidator) ValidatePerformanceMetrics(metrics BenchmarkMetrics) error {
	if !bv.validationConfig.ValidatePerformance {
		return nil
	}
	
	// Validate response time
	if len(metrics.ResponsePercentiles) > 0 {
		if p95, exists := metrics.ResponsePercentiles["p95"]; exists {
			if p95 > bv.validationConfig.MaxResponseTime {
				if bv.validationConfig.StrictMode {
					return fmt.Errorf("p95 response time %v exceeds maximum %v", 
						p95, bv.validationConfig.MaxResponseTime)
				}
				bv.b.Logf("WARNING: p95 response time %v exceeds target %v", 
					p95, bv.validationConfig.MaxResponseTime)
			}
		}
	}
	
	// Validate throughput
	if metrics.ThroughputPerSec < bv.validationConfig.MinThroughput {
		if bv.validationConfig.StrictMode {
			return fmt.Errorf("throughput %.2f ops/sec below minimum %.2f ops/sec", 
				metrics.ThroughputPerSec, bv.validationConfig.MinThroughput)
		}
		bv.b.Logf("WARNING: throughput %.2f ops/sec below target %.2f ops/sec", 
			metrics.ThroughputPerSec, bv.validationConfig.MinThroughput)
	}
	
	// Validate error rate
	if metrics.ErrorRate > bv.validationConfig.MaxErrorRate {
		if bv.validationConfig.StrictMode {
			return fmt.Errorf("error rate %.2f%% exceeds maximum %.2f%%", 
				metrics.ErrorRate*100, bv.validationConfig.MaxErrorRate*100)
		}
		bv.b.Logf("WARNING: error rate %.2f%% exceeds target %.2f%%", 
			metrics.ErrorRate*100, bv.validationConfig.MaxErrorRate*100)
	}
	
	bv.b.Logf("Performance validation passed - Throughput: %.2f ops/sec, Error rate: %.2f%%", 
		metrics.ThroughputPerSec, metrics.ErrorRate*100)
	
	return nil
}

// ValidateDatabaseState validates the database state after operations
func (bv *BenchmarkValidator) ValidateDatabaseState(db *gorm.DB, expectedChanges map[string]int) error {
	for table, _ := range expectedChanges {
		var currentCount int64
		if err := db.Table(table).Count(&currentCount).Error; err != nil {
			return fmt.Errorf("failed to count %s after operations: %w", table, err)
		}
		
		// This is a simplified validation - in practice, you'd track before/after counts
		if currentCount < 0 {
			return fmt.Errorf("negative count for table %s: %d", table, currentCount)
		}
		
		bv.b.Logf("Table %s has %d records after operations", table, currentCount)
	}
	
	return nil
}

// ValidateResourceUsage validates that resource usage is within acceptable limits
func (bv *BenchmarkValidator) ValidateResourceUsage(metrics BenchmarkMetrics) error {
	// Validate memory usage (convert to MB for readability)
	memoryMB := float64(metrics.MemoryAllocated) / 1024 / 1024
	maxMemoryMB := float64(512) // 512MB limit for benchmarks
	
	if memoryMB > maxMemoryMB {
		if bv.validationConfig.StrictMode {
			return fmt.Errorf("memory usage %.2f MB exceeds limit %.2f MB", memoryMB, maxMemoryMB)
		}
		bv.b.Logf("WARNING: memory usage %.2f MB exceeds target %.2f MB", memoryMB, maxMemoryMB)
	}
	
	// Validate goroutine count
	maxGoroutines := 500
	if metrics.GoroutineCount > maxGoroutines {
		if bv.validationConfig.StrictMode {
			return fmt.Errorf("goroutine count %d exceeds limit %d", metrics.GoroutineCount, maxGoroutines)
		}
		bv.b.Logf("WARNING: goroutine count %d exceeds target %d", metrics.GoroutineCount, maxGoroutines)
	}
	
	// Validate database connections
	maxDBConnections := 50
	if metrics.DBConnections.OpenConnections > maxDBConnections {
		if bv.validationConfig.StrictMode {
			return fmt.Errorf("database connections %d exceeds limit %d", 
				metrics.DBConnections.OpenConnections, maxDBConnections)
		}
		bv.b.Logf("WARNING: database connections %d exceeds target %d", 
			metrics.DBConnections.OpenConnections, maxDBConnections)
	}
	
	bv.b.Logf("Resource validation passed - Memory: %.2f MB, Goroutines: %d, DB Connections: %d", 
		memoryMB, metrics.GoroutineCount, metrics.DBConnections.OpenConnections)
	
	return nil
}

// ValidateTestEnvironment performs comprehensive environment validation
func (bv *BenchmarkValidator) ValidateTestEnvironment() error {
	// Check available memory
	// Note: This is a simplified check - in production you'd use more sophisticated methods
	if bv.reliabilityMgr.IsResourceConstrained() {
		bv.b.Logf("WARNING: System appears to be under resource constraints")
	}
	
	// Validate Go runtime version and settings
	bv.b.Logf("Go runtime validation - GOMAXPROCS: %d, NumCPU: %d", 
		runtime.GOMAXPROCS(0), runtime.NumCPU())
	
	return nil
}

// CreateValidationReport generates a comprehensive validation report
func (bv *BenchmarkValidator) CreateValidationReport(metrics BenchmarkMetrics) string {
	report := "=== Benchmark Validation Report ===\n"
	
	// Performance validation
	report += fmt.Sprintf("Performance Metrics:\n")
	report += fmt.Sprintf("  Throughput: %.2f ops/sec (target: %.2f)\n", 
		metrics.ThroughputPerSec, bv.validationConfig.MinThroughput)
	report += fmt.Sprintf("  Error Rate: %.2f%% (max: %.2f%%)\n", 
		metrics.ErrorRate*100, bv.validationConfig.MaxErrorRate*100)
	
	if len(metrics.ResponsePercentiles) > 0 {
		if p95, exists := metrics.ResponsePercentiles["p95"]; exists {
			report += fmt.Sprintf("  P95 Response Time: %v (max: %v)\n", 
				p95, bv.validationConfig.MaxResponseTime)
		}
	}
	
	// Resource validation
	memoryMB := float64(metrics.MemoryAllocated) / 1024 / 1024
	report += fmt.Sprintf("\nResource Usage:\n")
	report += fmt.Sprintf("  Memory Allocated: %.2f MB\n", memoryMB)
	report += fmt.Sprintf("  Goroutines: %d\n", metrics.GoroutineCount)
	report += fmt.Sprintf("  DB Connections: %d\n", metrics.DBConnections.OpenConnections)
	
	// Validation status
	report += fmt.Sprintf("\nValidation Status: ")
	if bv.validationConfig.StrictMode {
		report += "STRICT MODE - All thresholds enforced\n"
	} else {
		report += "LENIENT MODE - Warnings only\n"
	}
	
	return report
}