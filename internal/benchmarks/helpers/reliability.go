package helpers

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

// BenchmarkReliabilityManager provides comprehensive error handling and recovery mechanisms
type BenchmarkReliabilityManager struct {
	b                   *testing.B
	timeouts            map[string]time.Duration
	retryConfig         RetryConfig
	resourceMonitor     *ResourceMonitor
	cleanupFunctions    []func() error
	preflightChecks     []PreflightCheck
	gracefulDegradation *GracefulDegradationConfig
	mu                  sync.RWMutex
}

// RetryConfig defines retry behavior for failed operations
type RetryConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors []string
}

// ResourceMonitor tracks system resource usage and constraints
type ResourceMonitor struct {
	MaxMemoryMB      int64
	MaxGoroutines    int
	MaxDBConnections int
	CheckInterval    time.Duration
	AlertThresholds  ResourceThresholds
	mu               sync.RWMutex
	monitoring       bool
	stopChan         chan struct{}
}

// ResourceThresholds defines when to trigger alerts or degradation
type ResourceThresholds struct {
	MemoryWarningMB   int64
	MemoryCriticalMB  int64
	GoroutineWarning  int
	GoroutineCritical int
	DBConnWarning     int
	DBConnCritical    int
}

// PreflightCheck represents a validation that must pass before benchmark execution
type PreflightCheck struct {
	Name        string
	Description string
	CheckFunc   func() error
	Required    bool
}

// GracefulDegradationConfig defines how to handle resource constraints
type GracefulDegradationConfig struct {
	ReduceConcurrency    bool
	SkipNonEssentialOps  bool
	UseSimplifiedData    bool
	EnableFastMode       bool
	MaxOperationsPerTest int
}

// NewBenchmarkReliabilityManager creates a new reliability manager
func NewBenchmarkReliabilityManager(b *testing.B) *BenchmarkReliabilityManager {
	return &BenchmarkReliabilityManager{
		b: b,
		timeouts: map[string]time.Duration{
			"default":   30 * time.Second,
			"database":  60 * time.Second,
			"http":      30 * time.Second,
			"cleanup":   10 * time.Second,
			"preflight": 5 * time.Second,
		},
		retryConfig: RetryConfig{
			MaxRetries:    3,
			InitialDelay:  100 * time.Millisecond,
			MaxDelay:      5 * time.Second,
			BackoffFactor: 2.0,
			RetryableErrors: []string{
				"connection refused",
				"timeout",
				"temporary failure",
				"resource temporarily unavailable",
			},
		},
		resourceMonitor: &ResourceMonitor{
			MaxMemoryMB:      1024, // 1GB default limit
			MaxGoroutines:    1000,
			MaxDBConnections: 50,
			CheckInterval:    1 * time.Second,
			AlertThresholds: ResourceThresholds{
				MemoryWarningMB:   512, // 512MB warning
				MemoryCriticalMB:  896, // 896MB critical (leave 128MB buffer)
				GoroutineWarning:  500,
				GoroutineCritical: 800,
				DBConnWarning:     30,
				DBConnCritical:    45,
			},
		},
		gracefulDegradation: &GracefulDegradationConfig{
			ReduceConcurrency:    true,
			SkipNonEssentialOps:  true,
			UseSimplifiedData:    true,
			EnableFastMode:       false,
			MaxOperationsPerTest: 10000,
		},
		cleanupFunctions: make([]func() error, 0),
		preflightChecks:  make([]PreflightCheck, 0),
	}
}

// SetTimeout configures timeout for specific operation types
func (brm *BenchmarkReliabilityManager) SetTimeout(operation string, timeout time.Duration) {
	brm.mu.Lock()
	defer brm.mu.Unlock()
	brm.timeouts[operation] = timeout
}

// GetTimeout returns the timeout for a specific operation type
func (brm *BenchmarkReliabilityManager) GetTimeout(operation string) time.Duration {
	brm.mu.RLock()
	defer brm.mu.RUnlock()

	if timeout, exists := brm.timeouts[operation]; exists {
		return timeout
	}
	return brm.timeouts["default"]
}

// AddCleanupFunction registers a cleanup function to be called on benchmark completion
func (brm *BenchmarkReliabilityManager) AddCleanupFunction(cleanup func() error) {
	brm.mu.Lock()
	defer brm.mu.Unlock()
	brm.cleanupFunctions = append(brm.cleanupFunctions, cleanup)
}

// AddPreflightCheck registers a preflight check to be run before benchmark execution
func (brm *BenchmarkReliabilityManager) AddPreflightCheck(check PreflightCheck) {
	brm.mu.Lock()
	defer brm.mu.Unlock()
	brm.preflightChecks = append(brm.preflightChecks, check)
}

// RunPreflightChecks executes all registered preflight checks
func (brm *BenchmarkReliabilityManager) RunPreflightChecks() error {
	brm.mu.RLock()
	checks := make([]PreflightCheck, len(brm.preflightChecks))
	copy(checks, brm.preflightChecks)
	brm.mu.RUnlock()

	for _, check := range checks {
		brm.b.Logf("Running preflight check: %s", check.Name)

		ctx, cancel := context.WithTimeout(context.Background(), brm.GetTimeout("preflight"))
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- check.CheckFunc()
		}()

		select {
		case err := <-done:
			if err != nil {
				if check.Required {
					return fmt.Errorf("required preflight check '%s' failed: %w", check.Name, err)
				}
				brm.b.Logf("Optional preflight check '%s' failed: %v", check.Name, err)
			} else {
				brm.b.Logf("Preflight check '%s' passed", check.Name)
			}
		case <-ctx.Done():
			if check.Required {
				return fmt.Errorf("required preflight check '%s' timed out", check.Name)
			}
			brm.b.Logf("Optional preflight check '%s' timed out", check.Name)
		}
	}

	return nil
}

// StartResourceMonitoring begins monitoring system resources
func (brm *BenchmarkReliabilityManager) StartResourceMonitoring() {
	brm.resourceMonitor.mu.Lock()
	defer brm.resourceMonitor.mu.Unlock()

	if brm.resourceMonitor.monitoring {
		return // Already monitoring
	}

	brm.resourceMonitor.monitoring = true
	brm.resourceMonitor.stopChan = make(chan struct{})

	go brm.monitorResources()
}

// StopResourceMonitoring stops monitoring system resources
func (brm *BenchmarkReliabilityManager) StopResourceMonitoring() {
	brm.resourceMonitor.mu.Lock()
	defer brm.resourceMonitor.mu.Unlock()

	if !brm.resourceMonitor.monitoring {
		return
	}

	brm.resourceMonitor.monitoring = false
	close(brm.resourceMonitor.stopChan)
}

// monitorResources continuously monitors system resources
func (brm *BenchmarkReliabilityManager) monitorResources() {
	ticker := time.NewTicker(brm.resourceMonitor.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			brm.checkResourceUsage()
		case <-brm.resourceMonitor.stopChan:
			return
		}
	}
}

// checkResourceUsage checks current resource usage against thresholds
func (brm *BenchmarkReliabilityManager) checkResourceUsage() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	currentMemoryMB := int64(memStats.Alloc / 1024 / 1024)
	currentGoroutines := runtime.NumGoroutine()

	thresholds := brm.resourceMonitor.AlertThresholds

	// Check memory usage
	if currentMemoryMB >= thresholds.MemoryCriticalMB {
		brm.b.Logf("CRITICAL: Memory usage at %d MB (threshold: %d MB)",
			currentMemoryMB, thresholds.MemoryCriticalMB)
		brm.triggerGracefulDegradation("memory_critical")
	} else if currentMemoryMB >= thresholds.MemoryWarningMB {
		brm.b.Logf("WARNING: Memory usage at %d MB (threshold: %d MB)",
			currentMemoryMB, thresholds.MemoryWarningMB)
	}

	// Check goroutine count
	if currentGoroutines >= thresholds.GoroutineCritical {
		brm.b.Logf("CRITICAL: Goroutine count at %d (threshold: %d)",
			currentGoroutines, thresholds.GoroutineCritical)
		brm.triggerGracefulDegradation("goroutine_critical")
	} else if currentGoroutines >= thresholds.GoroutineWarning {
		brm.b.Logf("WARNING: Goroutine count at %d (threshold: %d)",
			currentGoroutines, thresholds.GoroutineWarning)
	}
}

// triggerGracefulDegradation activates degradation measures based on resource constraints
func (brm *BenchmarkReliabilityManager) triggerGracefulDegradation(reason string) {
	brm.b.Logf("Triggering graceful degradation due to: %s", reason)

	config := brm.gracefulDegradation

	switch reason {
	case "memory_critical":
		// Force garbage collection
		runtime.GC()
		runtime.GC() // Run twice for better cleanup

		// Enable fast mode to reduce memory allocations
		config.EnableFastMode = true
		config.UseSimplifiedData = true

	case "goroutine_critical":
		// Reduce concurrency to limit goroutine creation
		config.ReduceConcurrency = true

	case "db_critical":
		// Skip non-essential database operations
		config.SkipNonEssentialOps = true
	}
}

// ExecuteWithRetry executes a function with retry logic and timeout
func (brm *BenchmarkReliabilityManager) ExecuteWithRetry(
	operation string,
	fn func() error,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), brm.GetTimeout(operation))
	defer cancel()

	config := brm.retryConfig
	delay := config.InitialDelay

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return fmt.Errorf("operation '%s' timed out during retry %d", operation, attempt)
			}

			// Increase delay for next attempt
			delay = time.Duration(float64(delay) * config.BackoffFactor)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}

		// Execute function with timeout
		done := make(chan error, 1)
		go func() {
			done <- fn()
		}()

		select {
		case err := <-done:
			if err == nil {
				if attempt > 0 {
					brm.b.Logf("Operation '%s' succeeded on attempt %d", operation, attempt+1)
				}
				return nil
			}

			// Check if error is retryable
			if attempt < config.MaxRetries && brm.isRetryableError(err) {
				brm.b.Logf("Operation '%s' failed on attempt %d, retrying: %v", operation, attempt+1, err)
				continue
			}

			return fmt.Errorf("operation '%s' failed after %d attempts: %w", operation, attempt+1, err)

		case <-ctx.Done():
			return fmt.Errorf("operation '%s' timed out on attempt %d", operation, attempt+1)
		}
	}

	return fmt.Errorf("operation '%s' exhausted all retry attempts", operation)
}

// isRetryableError checks if an error should trigger a retry
func (brm *BenchmarkReliabilityManager) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	for _, retryableErr := range brm.retryConfig.RetryableErrors {
		if contains(errStr, retryableErr) {
			return true
		}
	}

	return false
}

// ExecuteWithTimeout executes a function with a timeout
func (brm *BenchmarkReliabilityManager) ExecuteWithTimeout(
	operation string,
	fn func() error,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), brm.GetTimeout(operation))
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation '%s' timed out after %v", operation, brm.GetTimeout(operation))
	}
}

// Cleanup executes all registered cleanup functions
func (brm *BenchmarkReliabilityManager) Cleanup() {
	brm.StopResourceMonitoring()

	brm.mu.RLock()
	cleanupFuncs := make([]func() error, len(brm.cleanupFunctions))
	copy(cleanupFuncs, brm.cleanupFunctions)
	brm.mu.RUnlock()

	for i, cleanup := range cleanupFuncs {
		err := brm.ExecuteWithTimeout("cleanup", cleanup)
		if err != nil {
			brm.b.Logf("Cleanup function %d failed: %v", i, err)
		}
	}
}

// GetGracefulDegradationConfig returns the current graceful degradation configuration
func (brm *BenchmarkReliabilityManager) GetGracefulDegradationConfig() *GracefulDegradationConfig {
	return brm.gracefulDegradation
}

// IsResourceConstrained checks if the system is currently under resource constraints
func (brm *BenchmarkReliabilityManager) IsResourceConstrained() bool {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	currentMemoryMB := int64(memStats.Alloc / 1024 / 1024)
	currentGoroutines := runtime.NumGoroutine()

	thresholds := brm.resourceMonitor.AlertThresholds

	return currentMemoryMB >= thresholds.MemoryWarningMB ||
		currentGoroutines >= thresholds.GoroutineWarning
}

// AdjustConcurrencyForConstraints adjusts concurrency based on current resource constraints
func (brm *BenchmarkReliabilityManager) AdjustConcurrencyForConstraints(requestedConcurrency int) int {
	if !brm.gracefulDegradation.ReduceConcurrency {
		return requestedConcurrency
	}

	if brm.IsResourceConstrained() {
		// Reduce concurrency by 50% under resource constraints
		adjusted := requestedConcurrency / 2
		if adjusted < 1 {
			adjusted = 1
		}

		brm.b.Logf("Reducing concurrency from %d to %d due to resource constraints",
			requestedConcurrency, adjusted)
		return adjusted
	}

	return requestedConcurrency
}

// ShouldSkipOperation determines if an operation should be skipped due to constraints
func (brm *BenchmarkReliabilityManager) ShouldSkipOperation(operationType string) bool {
	if !brm.gracefulDegradation.SkipNonEssentialOps {
		return false
	}

	nonEssentialOps := []string{
		"detailed_metrics",
		"complex_queries",
		"bulk_operations",
		"relationship_traversal",
	}

	if brm.IsResourceConstrained() {
		for _, nonEssential := range nonEssentialOps {
			if operationType == nonEssential {
				brm.b.Logf("Skipping non-essential operation '%s' due to resource constraints", operationType)
				return true
			}
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr)))
}

// containsSubstring performs a simple substring search
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// BenchmarkTimeoutManager manages timeouts for different benchmark operations
type BenchmarkTimeoutManager struct {
	defaultTimeout    time.Duration
	operationTimeouts map[string]time.Duration
	mu                sync.RWMutex
}

// NewBenchmarkTimeoutManager creates a new timeout manager
func NewBenchmarkTimeoutManager() *BenchmarkTimeoutManager {
	return &BenchmarkTimeoutManager{
		defaultTimeout: 30 * time.Second,
		operationTimeouts: map[string]time.Duration{
			"server_start":    60 * time.Second,
			"database_setup":  120 * time.Second,
			"data_generation": 180 * time.Second,
			"http_request":    10 * time.Second,
			"bulk_operation":  300 * time.Second,
			"cleanup":         30 * time.Second,
		},
	}
}

// SetOperationTimeout sets a timeout for a specific operation
func (btm *BenchmarkTimeoutManager) SetOperationTimeout(operation string, timeout time.Duration) {
	btm.mu.Lock()
	defer btm.mu.Unlock()
	btm.operationTimeouts[operation] = timeout
}

// GetOperationTimeout gets the timeout for a specific operation
func (btm *BenchmarkTimeoutManager) GetOperationTimeout(operation string) time.Duration {
	btm.mu.RLock()
	defer btm.mu.RUnlock()

	if timeout, exists := btm.operationTimeouts[operation]; exists {
		return timeout
	}
	return btm.defaultTimeout
}

// WithTimeout executes a function with the appropriate timeout for the operation
func (btm *BenchmarkTimeoutManager) WithTimeout(operation string, fn func() error) error {
	timeout := btm.GetOperationTimeout(operation)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation '%s' timed out after %v", operation, timeout)
	}
}
