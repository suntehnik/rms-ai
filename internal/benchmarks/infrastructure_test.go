package benchmarks

import (
	"net/http"
	"testing"
)

// BenchmarkInfrastructure tests the benchmark infrastructure setup
func BenchmarkInfrastructure(b *testing.B) {
	// Create benchmark suite
	suite := NewBenchmarkSuite(b)
	defer suite.Cleanup()

	// Seed with small dataset for infrastructure testing
	if err := suite.SeedTestData("small"); err != nil {
		b.Fatalf("Failed to seed test data: %v", err)
	}

	b.ResetTimer()
	suite.StartMetrics()

	// Run benchmark iterations
	for i := 0; i < b.N; i++ {
		// Test basic HTTP connectivity
		resp, err := suite.Client.GET("/health")
		if err != nil {
			b.Fatalf("Failed to make health check request: %v", err)
		}
		
		if resp.StatusCode != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}
		
		resp.Body.Close()
	}

	suite.EndMetrics(b)
}

// BenchmarkDatabaseConnection tests database connectivity performance
func BenchmarkDatabaseConnection(b *testing.B) {
	suite := NewBenchmarkSuite(b)
	defer suite.Cleanup()

	b.ResetTimer()
	suite.StartMetrics()

	for i := 0; i < b.N; i++ {
		// Test database connectivity through a simple query
		var count int64
		if err := suite.Server.DB.Raw("SELECT COUNT(*) FROM users").Scan(&count).Error; err != nil {
			b.Fatalf("Failed to execute database query: %v", err)
		}
	}

	suite.EndMetrics(b)
}