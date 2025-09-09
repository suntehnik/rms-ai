// Package benchmarks provides comprehensive performance testing utilities
// for the product requirements management system API endpoints.
//
// This package implements benchmark tests that measure API endpoint performance
// using PostgreSQL databases via testcontainers, providing realistic performance
// measurements that reflect production conditions.
//
// The benchmark implementation follows ADR-1 requirement to test actual service
// API endpoints rather than isolated database operations, ensuring measurements
// capture the full request/response cycle including middleware overhead,
// JSON serialization, and authentication.
package benchmarks

import (
	"testing"

	"product-requirements-management/internal/benchmarks/helpers"
	"product-requirements-management/internal/benchmarks/setup"
)

// BenchmarkSuite provides a complete benchmark testing environment
type BenchmarkSuite struct {
	Server    *setup.BenchmarkServer
	Client    *helpers.BenchmarkClient
	Auth      *helpers.AuthHelper
	Metrics   *helpers.MetricsCollector
	DataGen   *setup.DataGenerator
}

// NewBenchmarkSuite creates a new benchmark testing suite
func NewBenchmarkSuite(b *testing.B) *BenchmarkSuite {
	// Create benchmark server with PostgreSQL container
	server := setup.NewBenchmarkServer(b)
	
	// Start the server
	if err := server.Start(); err != nil {
		b.Fatalf("Failed to start benchmark server: %v", err)
	}

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Create authentication helper
	auth := helpers.NewAuthHelper(server.Config.JWT.Secret)

	// Set up authentication for the client
	testUser := helpers.GetDefaultTestUser()
	if err := auth.AuthenticateClient(client, testUser.ID, testUser.Username); err != nil {
		b.Fatalf("Failed to authenticate benchmark client: %v", err)
	}

	// Create metrics collector
	metrics := helpers.NewMetricsCollector()

	// Create data generator
	dataGen := setup.NewDataGenerator(server.DB)

	return &BenchmarkSuite{
		Server:  server,
		Client:  client,
		Auth:    auth,
		Metrics: metrics,
		DataGen: dataGen,
	}
}

// Cleanup cleans up all benchmark resources
func (bs *BenchmarkSuite) Cleanup() {
	if bs.Server != nil {
		bs.Server.Cleanup()
	}
}

// SeedTestData populates the database with test data for benchmarking
func (bs *BenchmarkSuite) SeedTestData(dataSetName string) error {
	var config setup.DataSetConfig
	
	switch dataSetName {
	case "small":
		config = setup.GetSmallDataSet()
	case "medium":
		config = setup.GetMediumDataSet()
	case "large":
		config = setup.GetLargeDataSet()
	default:
		config = setup.GetSmallDataSet() // Default to small dataset
	}

	return bs.DataGen.GenerateFullDataSet(config)
}

// ResetDatabase clears all data and re-runs migrations
func (bs *BenchmarkSuite) ResetDatabase() error {
	return bs.DataGen.ResetDatabase()
}

// CleanupTestData removes all test data from the database
func (bs *BenchmarkSuite) CleanupTestData() error {
	return bs.DataGen.CleanupDatabase()
}

// StartMetrics begins collecting performance metrics
func (bs *BenchmarkSuite) StartMetrics() {
	bs.Metrics.StartMeasurement()
}

// EndMetrics stops collecting metrics and reports them to the benchmark
func (bs *BenchmarkSuite) EndMetrics(b *testing.B) {
	metrics := bs.Metrics.EndMeasurement()
	bs.Metrics.ReportMetrics(b, metrics)
}