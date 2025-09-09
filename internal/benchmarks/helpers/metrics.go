package helpers

import (
	"database/sql"
	"runtime"
	"sort"
	"testing"
	"time"
)

// MetricsCollector collects and reports performance metrics for benchmarks
type MetricsCollector struct {
	StartTime    time.Time
	StartMemStats runtime.MemStats
	EndMemStats   runtime.MemStats
	DBStats      sql.DBStats
}

// BenchmarkMetrics contains detailed performance metrics
type BenchmarkMetrics struct {
	Duration          time.Duration
	MemoryAllocated   uint64
	MemoryAllocations int64
	GCPauses          []time.Duration
	DBConnections     int
	DBQueries         int64
	ResponseSizes     []int64
	ResponseTimes     []time.Duration
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

// StartMeasurement begins collecting performance metrics
func (mc *MetricsCollector) StartMeasurement() {
	mc.StartTime = time.Now()
	runtime.ReadMemStats(&mc.StartMemStats)
	runtime.GC() // Force GC to get clean baseline
}

// EndMeasurement stops collecting metrics and returns the results
func (mc *MetricsCollector) EndMeasurement() BenchmarkMetrics {
	endTime := time.Now()
	runtime.ReadMemStats(&mc.EndMemStats)

	return BenchmarkMetrics{
		Duration:          endTime.Sub(mc.StartTime),
		MemoryAllocated:   mc.EndMemStats.TotalAlloc - mc.StartMemStats.TotalAlloc,
		MemoryAllocations: int64(mc.EndMemStats.Mallocs - mc.StartMemStats.Mallocs),
		GCPauses:          mc.getGCPauses(),
		DBConnections:     mc.DBStats.OpenConnections,
		DBQueries:         0, // Will be populated by database middleware
	}
}

// ReportMetrics reports the collected metrics to the benchmark
func (mc *MetricsCollector) ReportMetrics(b *testing.B, metrics BenchmarkMetrics) {
	// Report memory allocations per operation
	b.ReportMetric(float64(metrics.MemoryAllocated)/float64(b.N), "bytes/op")
	b.ReportMetric(float64(metrics.MemoryAllocations)/float64(b.N), "allocs/op")
	
	// Report database connections if available
	if metrics.DBConnections > 0 {
		b.ReportMetric(float64(metrics.DBConnections), "db_conns")
	}
	
	// Report response time percentiles if available
	if len(metrics.ResponseTimes) > 0 {
		sort.Slice(metrics.ResponseTimes, func(i, j int) bool {
			return metrics.ResponseTimes[i] < metrics.ResponseTimes[j]
		})
		
		p50 := metrics.ResponseTimes[len(metrics.ResponseTimes)*50/100]
		p95 := metrics.ResponseTimes[len(metrics.ResponseTimes)*95/100]
		p99 := metrics.ResponseTimes[len(metrics.ResponseTimes)*99/100]
		
		b.ReportMetric(float64(p50.Nanoseconds())/1e6, "p50_ms")
		b.ReportMetric(float64(p95.Nanoseconds())/1e6, "p95_ms")
		b.ReportMetric(float64(p99.Nanoseconds())/1e6, "p99_ms")
	}
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

// ResponseTimeTracker tracks HTTP response times for metrics
type ResponseTimeTracker struct {
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

// RecordResponse records a response time and size
func (rtt *ResponseTimeTracker) RecordResponse(duration time.Duration, size int64) {
	rtt.ResponseTimes = append(rtt.ResponseTimes, duration)
	rtt.ResponseSizes = append(rtt.ResponseSizes, size)
}

// GetMetrics returns the collected response metrics
func (rtt *ResponseTimeTracker) GetMetrics() ([]time.Duration, []int64) {
	return rtt.ResponseTimes, rtt.ResponseSizes
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
		"p50": sorted[len(sorted)*50/100],
		"p90": sorted[len(sorted)*90/100],
		"p95": sorted[len(sorted)*95/100],
		"p99": sorted[len(sorted)*99/100],
	}

	return percentiles
}