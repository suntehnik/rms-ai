package helpers

import (
	"database/sql"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"
)

// MetricsCollector collects and reports performance metrics for benchmarks
type MetricsCollector struct {
	StartTime     time.Time
	EndTime       time.Time
	StartMemStats runtime.MemStats
	EndMemStats   runtime.MemStats
	StartDBStats  sql.DBStats
	EndDBStats    sql.DBStats
	DB            *gorm.DB

	// Response tracking
	ResponseTracker *ResponseTimeTracker

	// Concurrent access tracking
	mu               sync.RWMutex
	ConcurrentOps    int64
	MaxConcurrentOps int64
	ErrorCount       int64
	SuccessCount     int64
}

// BenchmarkMetrics contains detailed performance metrics
type BenchmarkMetrics struct {
	// Timing metrics
	Duration         time.Duration
	OperationsPerSec float64

	// Memory metrics
	MemoryAllocated   uint64
	MemoryAllocations int64
	MemoryFreed       uint64
	HeapSize          uint64
	GCPauses          []time.Duration
	GCCount           uint32

	// Database metrics
	DBConnections DBConnectionMetrics
	DBQueries     int64

	// Response metrics
	ResponseTimes       []time.Duration
	ResponseSizes       []int64
	ResponsePercentiles map[string]time.Duration

	// Concurrency metrics
	MaxConcurrentOps int64
	ErrorRate        float64
	ThroughputPerSec float64

	// System metrics
	CPUUsage       float64
	GoroutineCount int
}

// DBConnectionMetrics contains database connection pool metrics
type DBConnectionMetrics struct {
	OpenConnections    int
	InUseConnections   int
	IdleConnections    int
	MaxOpenConnections int
	WaitCount          int64
	WaitDuration       time.Duration
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(db *gorm.DB) *MetricsCollector {
	return &MetricsCollector{
		DB:              db,
		ResponseTracker: NewResponseTimeTracker(),
	}
}

// NewMetricsCollectorWithoutDB creates a metrics collector without database monitoring
func NewMetricsCollectorWithoutDB() *MetricsCollector {
	return &MetricsCollector{
		ResponseTracker: NewResponseTimeTracker(),
	}
}

// StartMeasurement begins collecting performance metrics
func (mc *MetricsCollector) StartMeasurement() {
	mc.StartTime = time.Now()
	runtime.ReadMemStats(&mc.StartMemStats)

	// Collect database stats if available
	if mc.DB != nil {
		if sqlDB, err := mc.DB.DB(); err == nil {
			mc.StartDBStats = sqlDB.Stats()
		}
	}

	// Reset counters
	mc.mu.Lock()
	mc.ConcurrentOps = 0
	mc.MaxConcurrentOps = 0
	mc.ErrorCount = 0
	mc.SuccessCount = 0
	mc.mu.Unlock()

	// Force GC to get clean baseline
	runtime.GC()
}

// EndMeasurement stops collecting metrics and returns the results
func (mc *MetricsCollector) EndMeasurement() BenchmarkMetrics {
	mc.EndTime = time.Now()
	runtime.ReadMemStats(&mc.EndMemStats)

	// Collect final database stats
	var dbMetrics DBConnectionMetrics
	if mc.DB != nil {
		if sqlDB, err := mc.DB.DB(); err == nil {
			mc.EndDBStats = sqlDB.Stats()
			dbMetrics = mc.getDBConnectionMetrics()
		}
	}

	duration := mc.EndTime.Sub(mc.StartTime)

	// Calculate response percentiles
	responseTimes, responseSizes := mc.ResponseTracker.GetMetrics()
	percentiles := CalculatePercentiles(responseTimes)

	// Calculate rates
	mc.mu.RLock()
	totalOps := mc.SuccessCount + mc.ErrorCount
	errorRate := float64(0)
	if totalOps > 0 {
		errorRate = float64(mc.ErrorCount) / float64(totalOps)
	}

	opsPerSec := float64(0)
	throughputPerSec := float64(0)
	if duration.Seconds() > 0 {
		opsPerSec = float64(totalOps) / duration.Seconds()
		throughputPerSec = float64(mc.SuccessCount) / duration.Seconds()
	}
	maxConcurrent := mc.MaxConcurrentOps
	mc.mu.RUnlock()

	return BenchmarkMetrics{
		// Timing metrics
		Duration:         duration,
		OperationsPerSec: opsPerSec,

		// Memory metrics
		MemoryAllocated:   mc.EndMemStats.TotalAlloc - mc.StartMemStats.TotalAlloc,
		MemoryAllocations: int64(mc.EndMemStats.Mallocs - mc.StartMemStats.Mallocs),
		MemoryFreed:       mc.EndMemStats.Frees - mc.StartMemStats.Frees,
		HeapSize:          mc.EndMemStats.HeapAlloc,
		GCPauses:          mc.getGCPauses(),
		GCCount:           mc.EndMemStats.NumGC - mc.StartMemStats.NumGC,

		// Database metrics
		DBConnections: dbMetrics,

		// Response metrics
		ResponseTimes:       responseTimes,
		ResponseSizes:       responseSizes,
		ResponsePercentiles: percentiles,

		// Concurrency metrics
		MaxConcurrentOps: maxConcurrent,
		ErrorRate:        errorRate,
		ThroughputPerSec: throughputPerSec,

		// System metrics
		GoroutineCount: runtime.NumGoroutine(),
	}
}

// ReportMetrics reports the collected metrics to the benchmark
func (mc *MetricsCollector) ReportMetrics(b *testing.B, metrics BenchmarkMetrics) {
	// Report memory metrics per operation
	if b.N > 0 {
		b.ReportMetric(float64(metrics.MemoryAllocated)/float64(b.N), "bytes/op")
		b.ReportMetric(float64(metrics.MemoryAllocations)/float64(b.N), "allocs/op")
		b.ReportMetric(float64(metrics.MemoryFreed)/float64(b.N), "frees/op")
	}

	// Report heap size and GC metrics
	b.ReportMetric(float64(metrics.HeapSize)/1024/1024, "heap_mb")
	b.ReportMetric(float64(metrics.GCCount), "gc_count")

	// Report average GC pause time
	if len(metrics.GCPauses) > 0 {
		var totalPause time.Duration
		for _, pause := range metrics.GCPauses {
			totalPause += pause
		}
		avgPause := totalPause / time.Duration(len(metrics.GCPauses))
		b.ReportMetric(float64(avgPause.Nanoseconds())/1e6, "gc_pause_ms")
	}

	// Report database connection metrics
	if metrics.DBConnections.OpenConnections > 0 {
		b.ReportMetric(float64(metrics.DBConnections.OpenConnections), "db_open_conns")
		b.ReportMetric(float64(metrics.DBConnections.InUseConnections), "db_inuse_conns")
		b.ReportMetric(float64(metrics.DBConnections.IdleConnections), "db_idle_conns")
		b.ReportMetric(float64(metrics.DBConnections.WaitCount), "db_wait_count")
		b.ReportMetric(float64(metrics.DBConnections.WaitDuration.Nanoseconds())/1e6, "db_wait_ms")
	}

	// Report response time percentiles
	if len(metrics.ResponsePercentiles) > 0 {
		if p50, ok := metrics.ResponsePercentiles["p50"]; ok {
			b.ReportMetric(float64(p50.Nanoseconds())/1e6, "p50_ms")
		}
		if p90, ok := metrics.ResponsePercentiles["p90"]; ok {
			b.ReportMetric(float64(p90.Nanoseconds())/1e6, "p90_ms")
		}
		if p95, ok := metrics.ResponsePercentiles["p95"]; ok {
			b.ReportMetric(float64(p95.Nanoseconds())/1e6, "p95_ms")
		}
		if p99, ok := metrics.ResponsePercentiles["p99"]; ok {
			b.ReportMetric(float64(p99.Nanoseconds())/1e6, "p99_ms")
		}
	}

	// Report throughput and concurrency metrics
	b.ReportMetric(metrics.OperationsPerSec, "ops/sec")
	b.ReportMetric(metrics.ThroughputPerSec, "success/sec")
	b.ReportMetric(metrics.ErrorRate*100, "error_rate_%")
	b.ReportMetric(float64(metrics.MaxConcurrentOps), "max_concurrent")
	b.ReportMetric(float64(metrics.GoroutineCount), "goroutines")

	// Report average response size
	if len(metrics.ResponseSizes) > 0 {
		var totalSize int64
		for _, size := range metrics.ResponseSizes {
			totalSize += size
		}
		avgSize := float64(totalSize) / float64(len(metrics.ResponseSizes))
		b.ReportMetric(avgSize/1024, "avg_response_kb")
	}
}

// ReportDetailedMetrics provides a detailed text report of all metrics
func (mc *MetricsCollector) ReportDetailedMetrics(metrics BenchmarkMetrics) string {
	report := fmt.Sprintf(`
=== Benchmark Performance Report ===
Duration: %v
Operations/sec: %.2f
Throughput/sec: %.2f
Error Rate: %.2f%%

=== Memory Metrics ===
Memory Allocated: %.2f MB
Memory Allocations: %d
Memory Freed: %d
Heap Size: %.2f MB
GC Count: %d
GC Pauses: %d (avg: %v)

=== Database Metrics ===
Open Connections: %d
In-Use Connections: %d
Idle Connections: %d
Wait Count: %d
Wait Duration: %v

=== Response Metrics ===
Total Responses: %d
`,
		metrics.Duration,
		metrics.OperationsPerSec,
		metrics.ThroughputPerSec,
		metrics.ErrorRate*100,

		float64(metrics.MemoryAllocated)/1024/1024,
		metrics.MemoryAllocations,
		metrics.MemoryFreed,
		float64(metrics.HeapSize)/1024/1024,
		metrics.GCCount,
		len(metrics.GCPauses),
		mc.getAverageGCPause(metrics.GCPauses),

		metrics.DBConnections.OpenConnections,
		metrics.DBConnections.InUseConnections,
		metrics.DBConnections.IdleConnections,
		metrics.DBConnections.WaitCount,
		metrics.DBConnections.WaitDuration,

		len(metrics.ResponseTimes),
	)

	// Add percentiles if available
	if len(metrics.ResponsePercentiles) > 0 {
		report += "Response Time Percentiles:\n"
		for percentile, duration := range metrics.ResponsePercentiles {
			report += fmt.Sprintf("  %s: %v\n", percentile, duration)
		}
	}

	return report
}

// getGCPauses extracts GC pause times from memory stats
func (mc *MetricsCollector) getGCPauses() []time.Duration {
	var pauses []time.Duration

	// Get GC pauses that occurred during measurement
	startGC := mc.StartMemStats.NumGC
	endGC := mc.EndMemStats.NumGC

	if endGC > startGC {
		// Extract pause times for GCs that occurred during measurement
		for i := startGC; i < endGC && i < uint32(len(mc.EndMemStats.PauseNs)); i++ {
			pauseIdx := i % uint32(len(mc.EndMemStats.PauseNs))
			if mc.EndMemStats.PauseNs[pauseIdx] > 0 {
				pauses = append(pauses, time.Duration(mc.EndMemStats.PauseNs[pauseIdx]))
			}
		}
	}

	return pauses
}

// getDBConnectionMetrics extracts database connection pool metrics
func (mc *MetricsCollector) getDBConnectionMetrics() DBConnectionMetrics {
	return DBConnectionMetrics{
		OpenConnections:    mc.EndDBStats.OpenConnections,
		InUseConnections:   mc.EndDBStats.InUse,
		IdleConnections:    mc.EndDBStats.Idle,
		MaxOpenConnections: mc.EndDBStats.MaxOpenConnections,
		WaitCount:          mc.EndDBStats.WaitCount,
		WaitDuration:       mc.EndDBStats.WaitDuration,
		// Note: MaxIdleConnections, MaxLifetime, MaxIdleTime are not available in sql.DBStats
		// These would need to be tracked separately if needed
	}
}

// getAverageGCPause calculates the average GC pause time
func (mc *MetricsCollector) getAverageGCPause(pauses []time.Duration) time.Duration {
	if len(pauses) == 0 {
		return 0
	}

	var total time.Duration
	for _, pause := range pauses {
		total += pause
	}

	return total / time.Duration(len(pauses))
}

// IncrementConcurrentOps safely increments the concurrent operations counter
func (mc *MetricsCollector) IncrementConcurrentOps() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.ConcurrentOps++
	if mc.ConcurrentOps > mc.MaxConcurrentOps {
		mc.MaxConcurrentOps = mc.ConcurrentOps
	}
}

// DecrementConcurrentOps safely decrements the concurrent operations counter
func (mc *MetricsCollector) DecrementConcurrentOps() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.ConcurrentOps > 0 {
		mc.ConcurrentOps--
	}
}

// RecordSuccess records a successful operation
func (mc *MetricsCollector) RecordSuccess() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.SuccessCount++
}

// RecordError records a failed operation
func (mc *MetricsCollector) RecordError() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.ErrorCount++
}

// RecordResponse records a response with timing and size information
func (mc *MetricsCollector) RecordResponse(duration time.Duration, size int64) {
	mc.ResponseTracker.RecordResponse(duration, size)
}

// ResponseTimeTracker tracks HTTP response times for metrics
type ResponseTimeTracker struct {
	mu            sync.RWMutex
	ResponseTimes []time.Duration
	ResponseSizes []int64
}

// NewResponseTimeTracker creates a new response time tracker
func NewResponseTimeTracker() *ResponseTimeTracker {
	return &ResponseTimeTracker{
		ResponseTimes: make([]time.Duration, 0),
		ResponseSizes: make([]int64, 0),
	}
}

// RecordResponse records a response time and size (thread-safe)
func (rtt *ResponseTimeTracker) RecordResponse(duration time.Duration, size int64) {
	rtt.mu.Lock()
	defer rtt.mu.Unlock()

	rtt.ResponseTimes = append(rtt.ResponseTimes, duration)
	rtt.ResponseSizes = append(rtt.ResponseSizes, size)
}

// GetMetrics returns the collected response metrics (thread-safe)
func (rtt *ResponseTimeTracker) GetMetrics() ([]time.Duration, []int64) {
	rtt.mu.RLock()
	defer rtt.mu.RUnlock()

	// Return copies to avoid race conditions
	times := make([]time.Duration, len(rtt.ResponseTimes))
	sizes := make([]int64, len(rtt.ResponseSizes))

	copy(times, rtt.ResponseTimes)
	copy(sizes, rtt.ResponseSizes)

	return times, sizes
}

// Reset clears all recorded metrics (thread-safe)
func (rtt *ResponseTimeTracker) Reset() {
	rtt.mu.Lock()
	defer rtt.mu.Unlock()

	rtt.ResponseTimes = rtt.ResponseTimes[:0]
	rtt.ResponseSizes = rtt.ResponseSizes[:0]
}

// GetCount returns the number of recorded responses (thread-safe)
func (rtt *ResponseTimeTracker) GetCount() int {
	rtt.mu.RLock()
	defer rtt.mu.RUnlock()

	return len(rtt.ResponseTimes)
}

// CalculatePercentiles calculates response time percentiles
func CalculatePercentiles(durations []time.Duration) map[string]time.Duration {
	if len(durations) == 0 {
		return map[string]time.Duration{}
	}

	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	percentiles := map[string]time.Duration{
		"min": sorted[0],
		"p25": sorted[len(sorted)*25/100],
		"p50": sorted[len(sorted)*50/100],
		"p75": sorted[len(sorted)*75/100],
		"p90": sorted[len(sorted)*90/100],
		"p95": sorted[len(sorted)*95/100],
		"p99": sorted[len(sorted)*99/100],
		"max": sorted[len(sorted)-1],
	}

	return percentiles
}

// CalculateStatistics calculates basic statistical measures for durations
func CalculateStatistics(durations []time.Duration) map[string]interface{} {
	if len(durations) == 0 {
		return map[string]interface{}{}
	}

	var total time.Duration
	min := durations[0]
	max := durations[0]

	for _, d := range durations {
		total += d
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
	}

	mean := total / time.Duration(len(durations))

	// Calculate standard deviation
	var variance float64
	for _, d := range durations {
		diff := float64(d - mean)
		variance += diff * diff
	}
	variance /= float64(len(durations))
	stddev := time.Duration(variance)

	return map[string]interface{}{
		"count":  len(durations),
		"total":  total,
		"mean":   mean,
		"min":    min,
		"max":    max,
		"stddev": stddev,
	}
}

// BenchmarkResultFormatter formats benchmark results for different output formats
type BenchmarkResultFormatter struct {
	Metrics BenchmarkMetrics
}

// NewBenchmarkResultFormatter creates a new result formatter
func NewBenchmarkResultFormatter(metrics BenchmarkMetrics) *BenchmarkResultFormatter {
	return &BenchmarkResultFormatter{
		Metrics: metrics,
	}
}

// ToJSON formats the metrics as JSON string
func (brf *BenchmarkResultFormatter) ToJSON() (string, error) {
	// Note: In a real implementation, you'd use json.Marshal
	// For now, return a formatted string representation
	return fmt.Sprintf(`{
  "duration": "%v",
  "operations_per_sec": %.2f,
  "throughput_per_sec": %.2f,
  "error_rate": %.4f,
  "memory_allocated_mb": %.2f,
  "memory_allocations": %d,
  "heap_size_mb": %.2f,
  "gc_count": %d,
  "db_open_connections": %d,
  "db_wait_count": %d,
  "max_concurrent_ops": %d,
  "response_count": %d
}`,
		brf.Metrics.Duration,
		brf.Metrics.OperationsPerSec,
		brf.Metrics.ThroughputPerSec,
		brf.Metrics.ErrorRate,
		float64(brf.Metrics.MemoryAllocated)/1024/1024,
		brf.Metrics.MemoryAllocations,
		float64(brf.Metrics.HeapSize)/1024/1024,
		brf.Metrics.GCCount,
		brf.Metrics.DBConnections.OpenConnections,
		brf.Metrics.DBConnections.WaitCount,
		brf.Metrics.MaxConcurrentOps,
		len(brf.Metrics.ResponseTimes),
	), nil
}

// ToCSV formats the metrics as CSV string
func (brf *BenchmarkResultFormatter) ToCSV() string {
	header := "duration,ops_per_sec,throughput_per_sec,error_rate,memory_mb,allocations,heap_mb,gc_count,db_conns,max_concurrent,responses"
	data := fmt.Sprintf("%v,%.2f,%.2f,%.4f,%.2f,%d,%.2f,%d,%d,%d,%d",
		brf.Metrics.Duration,
		brf.Metrics.OperationsPerSec,
		brf.Metrics.ThroughputPerSec,
		brf.Metrics.ErrorRate,
		float64(brf.Metrics.MemoryAllocated)/1024/1024,
		brf.Metrics.MemoryAllocations,
		float64(brf.Metrics.HeapSize)/1024/1024,
		brf.Metrics.GCCount,
		brf.Metrics.DBConnections.OpenConnections,
		brf.Metrics.MaxConcurrentOps,
		len(brf.Metrics.ResponseTimes),
	)

	return header + "\n" + data
}

// CompareMetrics compares two benchmark metrics and returns a comparison report
func CompareMetrics(baseline, current BenchmarkMetrics) string {
	report := "=== Benchmark Comparison Report ===\n"

	// Duration comparison
	durationChange := float64(current.Duration-baseline.Duration) / float64(baseline.Duration) * 100
	report += fmt.Sprintf("Duration: %v -> %v (%.2f%% change)\n", baseline.Duration, current.Duration, durationChange)

	// Throughput comparison
	throughputChange := (current.ThroughputPerSec - baseline.ThroughputPerSec) / baseline.ThroughputPerSec * 100
	report += fmt.Sprintf("Throughput: %.2f -> %.2f ops/sec (%.2f%% change)\n",
		baseline.ThroughputPerSec, current.ThroughputPerSec, throughputChange)

	// Memory comparison
	memoryChange := float64(int64(current.MemoryAllocated)-int64(baseline.MemoryAllocated)) / float64(baseline.MemoryAllocated) * 100
	report += fmt.Sprintf("Memory: %.2f -> %.2f MB (%.2f%% change)\n",
		float64(baseline.MemoryAllocated)/1024/1024,
		float64(current.MemoryAllocated)/1024/1024,
		memoryChange)

	// Error rate comparison
	errorRateChange := (current.ErrorRate - baseline.ErrorRate) * 100
	report += fmt.Sprintf("Error Rate: %.2f%% -> %.2f%% (%.2f%% point change)\n",
		baseline.ErrorRate*100, current.ErrorRate*100, errorRateChange)

	return report
}
