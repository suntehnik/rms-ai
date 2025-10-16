package benchmarks

import (
	"context"
	"testing"
	"time"

	"product-requirements-management/internal/benchmarks/helpers"
	"product-requirements-management/internal/benchmarks/setup"
)

// BenchmarkMetricsCollectionExample demonstrates the enhanced metrics collection system
func BenchmarkMetricsCollectionExample(b *testing.B) {
	ctx := context.Background()

	// Create PostgreSQL container for database metrics
	dbContainer, err := setup.NewPostgreSQLContainer(ctx)
	if err != nil {
		b.Fatalf("Failed to create PostgreSQL container: %v", err)
	}
	defer dbContainer.Cleanup(ctx)

	b.Run("BasicMetricsCollection", func(b *testing.B) {
		// Create metrics collector with database monitoring
		collector := helpers.NewMetricsCollector(dbContainer.DB)

		// Start measurement
		collector.StartMeasurement()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate some work with concurrent operations
			collector.IncrementConcurrentOps()

			// Simulate HTTP response
			start := time.Now()
			time.Sleep(time.Millisecond * 10) // Simulate processing time
			duration := time.Since(start)

			// Record response metrics
			collector.RecordResponse(duration, 1024) // 1KB response
			collector.RecordSuccess()

			collector.DecrementConcurrentOps()
		}
		b.StopTimer()

		// End measurement and collect metrics
		metrics := collector.EndMeasurement()

		// Report metrics to benchmark
		collector.ReportMetrics(b, metrics)

		// Print detailed report (optional, for debugging)
		if testing.Verbose() {
			b.Logf("Detailed Metrics Report:\n%s", collector.ReportDetailedMetrics(metrics))
		}
	})

	b.Run("ConcurrentOperationsMetrics", func(b *testing.B) {
		collector := helpers.NewMetricsCollector(dbContainer.DB)
		collector.StartMeasurement()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				collector.IncrementConcurrentOps()

				// Simulate concurrent work
				start := time.Now()
				time.Sleep(time.Millisecond * 5)
				duration := time.Since(start)

				// Randomly simulate success or error
				if time.Now().UnixNano()%10 < 9 { // 90% success rate
					collector.RecordSuccess()
					collector.RecordResponse(duration, 512)
				} else {
					collector.RecordError()
				}

				collector.DecrementConcurrentOps()
			}
		})
		b.StopTimer()

		metrics := collector.EndMeasurement()
		collector.ReportMetrics(b, metrics)

		// Verify concurrency metrics
		if metrics.MaxConcurrentOps == 0 {
			b.Errorf("Expected max concurrent operations > 0, got %d", metrics.MaxConcurrentOps)
		}

		if metrics.ErrorRate < 0 || metrics.ErrorRate > 1 {
			b.Errorf("Expected error rate between 0 and 1, got %f", metrics.ErrorRate)
		}
	})

	b.Run("DatabaseConnectionPoolMetrics", func(b *testing.B) {
		collector := helpers.NewMetricsCollector(dbContainer.DB)
		dataGen := setup.NewDataGenerator(dbContainer.DB)

		collector.StartMeasurement()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Perform database operations to test connection pool
			users, err := dataGen.CreateUsers(10)
			if err != nil {
				collector.RecordError()
				continue
			}

			_, err = dataGen.CreateEpics(5, users)
			if err != nil {
				collector.RecordError()
				continue
			}

			collector.RecordSuccess()

			// Cleanup after each iteration
			if err := dataGen.CleanupDatabase(); err != nil {
				b.Logf("Warning: cleanup failed: %v", err)
			}
		}
		b.StopTimer()

		metrics := collector.EndMeasurement()
		collector.ReportMetrics(b, metrics)

		// Verify database metrics are collected
		if metrics.DBConnections.OpenConnections == 0 {
			b.Logf("Warning: No database connections recorded")
		}
	})

	b.Run("ResponseTimePercentiles", func(b *testing.B) {
		collector := helpers.NewMetricsCollector(dbContainer.DB)
		collector.StartMeasurement()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate varying response times
			var duration time.Duration
			switch i % 4 {
			case 0:
				duration = time.Millisecond * 10 // Fast response
			case 1:
				duration = time.Millisecond * 50 // Medium response
			case 2:
				duration = time.Millisecond * 100 // Slow response
			case 3:
				duration = time.Millisecond * 200 // Very slow response
			}

			time.Sleep(duration)
			collector.RecordResponse(duration, int64(100+i%900)) // Varying response sizes
			collector.RecordSuccess()
		}
		b.StopTimer()

		metrics := collector.EndMeasurement()
		collector.ReportMetrics(b, metrics)

		// Verify percentiles are calculated
		if len(metrics.ResponsePercentiles) == 0 {
			b.Errorf("Expected response percentiles to be calculated")
		}

		// Check that percentiles are in ascending order
		if p50, ok := metrics.ResponsePercentiles["p50"]; ok {
			if p95, ok := metrics.ResponsePercentiles["p95"]; ok {
				if p50 > p95 {
					b.Errorf("Expected p50 (%v) <= p95 (%v)", p50, p95)
				}
			}
		}
	})
}

// BenchmarkMetricsFormattingExample demonstrates result formatting capabilities
func BenchmarkMetricsFormattingExample(b *testing.B) {
	collector := helpers.NewMetricsCollectorWithoutDB()
	collector.StartMeasurement()

	// Simulate some operations
	for i := 0; i < 100; i++ {
		collector.RecordResponse(time.Millisecond*time.Duration(10+i%50), int64(500+i%1000))
		collector.RecordSuccess()
	}

	metrics := collector.EndMeasurement()

	// Test different formatting options
	formatter := helpers.NewBenchmarkResultFormatter(metrics)

	// JSON format
	jsonResult, err := formatter.ToJSON()
	if err != nil {
		b.Errorf("Failed to format as JSON: %v", err)
	}

	if testing.Verbose() {
		b.Logf("JSON Format:\n%s", jsonResult)
	}

	// CSV format
	csvResult := formatter.ToCSV()
	if testing.Verbose() {
		b.Logf("CSV Format:\n%s", csvResult)
	}

	// Detailed report
	detailedReport := collector.ReportDetailedMetrics(metrics)
	if testing.Verbose() {
		b.Logf("Detailed Report:\n%s", detailedReport)
	}
}
