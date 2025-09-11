package api

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"product-requirements-management/internal/benchmarks/helpers"
)

// stringPtr creates a pointer to a string value
func stringPtr(s string) *string {
	return &s
}

// intPtr creates a pointer to an int value
func intPtr(i int) *int {
	return &i
}

// boolPtr creates a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}

// validateTestData ensures that test data arrays have sufficient elements for benchmark operations
func validateTestData(b *testing.B, dataName string, dataLength, requiredLength int) {
	if dataLength == 0 {
		b.Fatalf("No %s available for benchmark testing", dataName)
	}
	if requiredLength > 0 && dataLength < requiredLength {
		b.Logf("Warning: Only %d %s available, but %d required. Will cycle through available data.", 
			dataLength, dataName, requiredLength)
	}
}

// safeIndex returns a safe index for accessing arrays with bounds checking
func safeIndex(index, arrayLength int) int {
	if arrayLength == 0 {
		return 0
	}
	return index % arrayLength
}

// validateUUIDArray ensures UUID array has valid elements and returns safe index
func validateUUIDArray(b *testing.B, ids []uuid.UUID, index int, arrayName string) int {
	if len(ids) == 0 {
		b.Fatalf("No %s available for benchmark testing", arrayName)
	}
	
	safeIdx := safeIndex(index, len(ids))
	if ids[safeIdx] == uuid.Nil {
		b.Fatalf("Invalid UUID at index %d in %s array", safeIdx, arrayName)
	}
	
	return safeIdx
}

// ensureMinimumTestData checks if we have minimum required test data for benchmarks
func ensureMinimumTestData(b *testing.B, counts map[string]int, minimums map[string]int) {
	for dataType, minimum := range minimums {
		if actual, exists := counts[dataType]; !exists || actual < minimum {
			b.Fatalf("Insufficient %s for benchmark: need at least %d, got %d", 
				dataType, minimum, actual)
		}
	}
}

// BenchmarkDataRequirements defines minimum data requirements for different benchmark types
type BenchmarkDataRequirements struct {
	Users      int
	Epics      int
	UserStories int
	Requirements int
	Comments   int
}

// GetCRUDRequirements returns minimum data requirements for CRUD benchmarks
func GetCRUDRequirements() BenchmarkDataRequirements {
	return BenchmarkDataRequirements{
		Users:      1,
		Epics:      1,
		UserStories: 1,
		Requirements: 0,
		Comments:   0,
	}
}

// GetListingRequirements returns minimum data requirements for listing benchmarks
func GetListingRequirements() BenchmarkDataRequirements {
	return BenchmarkDataRequirements{
		Users:      5,
		Epics:      10,
		UserStories: 20,
		Requirements: 50,
		Comments:   10,
	}
}

// GetConcurrencyRequirements returns minimum data requirements for concurrency benchmarks
func GetConcurrencyRequirements() BenchmarkDataRequirements {
	return BenchmarkDataRequirements{
		Users:      10,
		Epics:      20,
		UserStories: 50,
		Requirements: 100,
		Comments:   20,
	}
}

// ValidateBenchmarkData validates that sufficient test data exists for the benchmark type
func ValidateBenchmarkData(b *testing.B, actualCounts, requirements BenchmarkDataRequirements) {
	validations := map[string][2]int{
		"users":       {actualCounts.Users, requirements.Users},
		"epics":       {actualCounts.Epics, requirements.Epics},
		"user_stories": {actualCounts.UserStories, requirements.UserStories},
		"requirements": {actualCounts.Requirements, requirements.Requirements},
		"comments":    {actualCounts.Comments, requirements.Comments},
	}

	for dataType, counts := range validations {
		actual, required := counts[0], counts[1]
		if required > 0 && actual < required {
			b.Logf("Warning: %s has %d items, benchmark requires %d. Will cycle through available data.", 
				dataType, actual, required)
		}
		if actual == 0 && required > 0 {
			b.Fatalf("No %s available for benchmark testing, but %d required", dataType, required)
		}
	}
}

// BenchmarkErrorHandler provides consistent error handling for benchmark tests
type BenchmarkErrorHandler struct {
	b *testing.B
}

// NewBenchmarkErrorHandler creates a new error handler for benchmarks
func NewBenchmarkErrorHandler(b *testing.B) *BenchmarkErrorHandler {
	return &BenchmarkErrorHandler{b: b}
}

// RequireNoError fails the benchmark if error is not nil
func (beh *BenchmarkErrorHandler) RequireNoError(err error, msgAndArgs ...interface{}) {
	if err != nil {
		if len(msgAndArgs) > 0 {
			beh.b.Fatalf("Benchmark failed with error: %v. %s", err, fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...))
		} else {
			beh.b.Fatalf("Benchmark failed with error: %v", err)
		}
	}
}

// RequireEqual fails the benchmark if values are not equal
func (beh *BenchmarkErrorHandler) RequireEqual(expected, actual interface{}, msgAndArgs ...interface{}) {
	if expected != actual {
		if len(msgAndArgs) > 0 {
			beh.b.Fatalf("Values not equal: expected %v, got %v. %s", expected, actual, fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...))
		} else {
			beh.b.Fatalf("Values not equal: expected %v, got %v", expected, actual)
		}
	}
}

// RequireNotEmpty fails the benchmark if slice/array is empty
func (beh *BenchmarkErrorHandler) RequireNotEmpty(obj interface{}, msgAndArgs ...interface{}) {
	switch v := obj.(type) {
	case []uuid.UUID:
		if len(v) == 0 {
			beh.b.Fatalf("UUID slice is empty. %s", beh.formatMessage(msgAndArgs...))
		}
	case []string:
		if len(v) == 0 {
			beh.b.Fatalf("String slice is empty. %s", beh.formatMessage(msgAndArgs...))
		}
	default:
		beh.b.Fatalf("Unsupported type for RequireNotEmpty: %T", obj)
	}
}

// formatMessage formats optional message arguments
func (beh *BenchmarkErrorHandler) formatMessage(msgAndArgs ...interface{}) string {
	if len(msgAndArgs) > 0 {
		return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
	return ""
}

// BenchmarkDataValidator provides validation utilities for benchmark test data
type BenchmarkDataValidator struct {
	errorHandler *BenchmarkErrorHandler
}

// NewBenchmarkDataValidator creates a new data validator
func NewBenchmarkDataValidator(b *testing.B) *BenchmarkDataValidator {
	return &BenchmarkDataValidator{
		errorHandler: NewBenchmarkErrorHandler(b),
	}
}

// BenchmarkTestRunner provides enhanced test execution with reliability features
type BenchmarkTestRunner struct {
	b                *testing.B
	reliabilityMgr   *helpers.BenchmarkReliabilityManager
	validator        *helpers.BenchmarkValidator
	resourceMgr      *helpers.BenchmarkResourceManager
	timeoutMgr       *helpers.BenchmarkTimeoutManager
}

// NewBenchmarkTestRunner creates a new test runner with reliability features
func NewBenchmarkTestRunner(b *testing.B, db *gorm.DB) *BenchmarkTestRunner {
	reliabilityMgr := helpers.NewBenchmarkReliabilityManager(b)
	validator := helpers.NewBenchmarkValidator(b, reliabilityMgr)
	resourceMgr := helpers.NewBenchmarkResourceManager(b, db)
	timeoutMgr := helpers.NewBenchmarkTimeoutManager()
	
	return &BenchmarkTestRunner{
		b:              b,
		reliabilityMgr: reliabilityMgr,
		validator:      validator,
		resourceMgr:    resourceMgr,
		timeoutMgr:     timeoutMgr,
	}
}

// ExecuteWithReliability executes a benchmark operation with full reliability features
func (btr *BenchmarkTestRunner) ExecuteWithReliability(
	operationName string,
	operation func() error,
) error {
	// Start resource monitoring
	btr.reliabilityMgr.StartResourceMonitoring()
	defer btr.reliabilityMgr.StopResourceMonitoring()
	
	// Execute with retry and timeout
	return btr.reliabilityMgr.ExecuteWithRetry(operationName, operation)
}

// SetupBenchmarkEnvironment performs comprehensive benchmark environment setup
func (btr *BenchmarkTestRunner) SetupBenchmarkEnvironment(db *gorm.DB, baseURL string) error {
	// Add preflight checks
	btr.reliabilityMgr.AddPreflightCheck(helpers.PreflightCheck{
		Name:        "database_connectivity",
		Description: "Verify database is accessible",
		Required:    true,
		CheckFunc: func() error {
			return btr.validator.ValidateDatabaseConnection(db)
		},
	})
	
	btr.reliabilityMgr.AddPreflightCheck(helpers.PreflightCheck{
		Name:        "server_availability",
		Description: "Verify HTTP server is running",
		Required:    true,
		CheckFunc: func() error {
			return btr.validator.ValidateServerAvailability(baseURL)
		},
	})
	
	btr.reliabilityMgr.AddPreflightCheck(helpers.PreflightCheck{
		Name:        "resource_availability",
		Description: "Check system resource availability",
		Required:    false,
		CheckFunc: func() error {
			return btr.validator.ValidateTestEnvironment()
		},
	})
	
	// Run preflight checks
	if err := btr.reliabilityMgr.RunPreflightChecks(); err != nil {
		return fmt.Errorf("preflight checks failed: %w", err)
	}
	
	// Validate prerequisites
	return btr.validator.ValidatePrerequisites(db, baseURL)
}

// Cleanup performs comprehensive cleanup with error handling
func (btr *BenchmarkTestRunner) Cleanup() {
	defer func() {
		if r := recover(); r != nil {
			btr.b.Logf("Panic during cleanup: %v", r)
		}
	}()
	
	btr.resourceMgr.ExecuteCleanup()
	btr.reliabilityMgr.Cleanup()
}

// GetReliabilityManager returns the reliability manager for advanced configuration
func (btr *BenchmarkTestRunner) GetReliabilityManager() *helpers.BenchmarkReliabilityManager {
	return btr.reliabilityMgr
}

// GetValidator returns the validator for custom validation
func (btr *BenchmarkTestRunner) GetValidator() *helpers.BenchmarkValidator {
	return btr.validator
}

// GetResourceManager returns the resource manager for resource tracking
func (btr *BenchmarkTestRunner) GetResourceManager() *helpers.BenchmarkResourceManager {
	return btr.resourceMgr
}

// ValidateUUIDs ensures all UUIDs in the slice are valid (not nil)
func (bdv *BenchmarkDataValidator) ValidateUUIDs(ids []uuid.UUID, dataType string) {
	bdv.errorHandler.RequireNotEmpty(ids, "No %s IDs available", dataType)
	
	for i, id := range ids {
		if id == uuid.Nil {
			bdv.errorHandler.RequireNoError(fmt.Errorf("invalid UUID at index %d", i), 
				"Invalid %s UUID found", dataType)
		}
	}
}

// ValidateMinimumCount ensures we have at least the minimum required items
func (bdv *BenchmarkDataValidator) ValidateMinimumCount(actual, minimum int, dataType string) {
	if actual < minimum {
		bdv.errorHandler.RequireNoError(fmt.Errorf("insufficient data"), 
			"Need at least %d %s, got %d", minimum, dataType, actual)
	}
}

// ValidateNonZeroCount ensures we have at least one item
func (bdv *BenchmarkDataValidator) ValidateNonZeroCount(count int, dataType string) {
	if count == 0 {
		bdv.errorHandler.RequireNoError(fmt.Errorf("no data available"), 
			"No %s available for benchmark", dataType)
	}
}