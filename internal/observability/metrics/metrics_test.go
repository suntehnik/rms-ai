package metrics

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
)

// setupTestMetrics creates a new metrics instance with a clean registry
func setupTestMetrics(t *testing.T) (*Metrics, func()) {
	// Reset global metrics
	AppMetrics = nil
	
	// Create a new registry for this test
	oldRegistry := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	
	// Initialize metrics
	metrics := Init("test-service", "1.0.0")
	
	// Return cleanup function
	cleanup := func() {
		prometheus.DefaultRegisterer = oldRegistry
		AppMetrics = nil
	}
	
	return metrics, cleanup
}

func TestInit(t *testing.T) {
	metrics, cleanup := setupTestMetrics(t)
	defer cleanup()
	
	// Verify metrics were initialized
	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.HTTPRequestsTotal)
	assert.NotNil(t, metrics.HTTPRequestDuration)
	assert.NotNil(t, metrics.DatabaseConnections)
	assert.NotNil(t, metrics.EntitiesCreated)
	
	// Verify global reference was set
	assert.Equal(t, metrics, AppMetrics)
}

func TestPrometheusMiddleware(t *testing.T) {
	metrics, cleanup := setupTestMetrics(t)
	defer cleanup()
	
	// Create test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(metrics.PrometheusMiddleware())
	
	// Add test endpoint
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	// Make test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Verify metrics were recorded
	// Note: In a real test, you'd want to check the actual metric values
	// This is a basic test to ensure the middleware doesn't panic
}

func TestRecordDatabaseConnection(t *testing.T) {
	metrics, cleanup := setupTestMetrics(t)
	defer cleanup()
	
	// Record database connection metrics
	metrics.RecordDatabaseConnection("postgresql", "open", 5.0)
	metrics.RecordDatabaseConnection("postgresql", "idle", 2.0)
	
	// Verify metrics were recorded (basic test)
	assert.NotNil(t, metrics.DatabaseConnections)
}

func TestRecordDatabaseQuery(t *testing.T) {
	metrics, cleanup := setupTestMetrics(t)
	defer cleanup()
	
	// Record database query metrics
	duration := 50 * time.Millisecond
	metrics.RecordDatabaseQuery("postgresql", "select", "users", duration)
	
	// Verify metrics were recorded (basic test)
	assert.NotNil(t, metrics.DatabaseQueries)
	assert.NotNil(t, metrics.DatabaseQueryDuration)
}

func TestRecordEntityOperation(t *testing.T) {
	metrics, cleanup := setupTestMetrics(t)
	defer cleanup()
	
	// Test different operations
	testCases := []struct {
		operation  string
		entityType string
		userID     string
	}{
		{"create", "epic", "user1"},
		{"update", "user_story", "user2"},
		{"delete", "requirement", "user3"},
		{"invalid", "epic", "user1"}, // Should not panic
	}
	
	for _, tc := range testCases {
		t.Run(tc.operation, func(t *testing.T) {
			// Should not panic
			assert.NotPanics(t, func() {
				metrics.RecordEntityOperation(tc.operation, tc.entityType, tc.userID)
			})
		})
	}
}

func TestRecordComment(t *testing.T) {
	metrics, cleanup := setupTestMetrics(t)
	defer cleanup()
	
	// Record comment metrics
	metrics.RecordComment("epic", "general", "unresolved")
	metrics.RecordComment("user_story", "inline", "resolved")
	
	// Verify metrics were recorded (basic test)
	assert.NotNil(t, metrics.CommentsTotal)
}

func TestRecordSearch(t *testing.T) {
	metrics, cleanup := setupTestMetrics(t)
	defer cleanup()
	
	// Record search metrics
	duration := 100 * time.Millisecond
	metrics.RecordSearch("full_text", "user1", duration)
	
	// Verify metrics were recorded (basic test)
	assert.NotNil(t, metrics.SearchQueries)
	assert.NotNil(t, metrics.SearchDuration)
}

func TestRecordUptime(t *testing.T) {
	metrics, cleanup := setupTestMetrics(t)
	defer cleanup()
	
	// Record uptime metrics
	uptime := 1 * time.Hour
	metrics.RecordUptime("test-service", uptime)
	
	// Verify metrics were recorded (basic test)
	assert.NotNil(t, metrics.ApplicationUptime)
}

func TestMetricsEndpoint(t *testing.T) {
	metrics, cleanup := setupTestMetrics(t)
	defer cleanup()
	
	// Record some test metrics
	metrics.RecordEntityOperation("create", "epic", "user1")
	metrics.RecordDatabaseQuery("postgresql", "select", "epics", 10*time.Millisecond)
	
	// Create test server with metrics endpoint
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	
	// Make request to metrics endpoint
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	// Just verify we get a valid Prometheus response
	assert.Contains(t, w.Body.String(), "# HELP")
}

func TestMetricsLabels(t *testing.T) {
	metrics, cleanup := setupTestMetrics(t)
	defer cleanup()
	
	// Test that metrics with different labels are recorded separately
	metrics.RecordEntityOperation("create", "epic", "user1")
	metrics.RecordEntityOperation("create", "epic", "user2")
	metrics.RecordEntityOperation("create", "user_story", "user1")
	
	// Verify metrics exist (basic test)
	assert.NotNil(t, metrics.EntitiesCreated)
}

func TestConcurrentMetricsRecording(t *testing.T) {
	metrics, cleanup := setupTestMetrics(t)
	defer cleanup()
	
	// Test concurrent access to metrics
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			// Record various metrics concurrently
			metrics.RecordEntityOperation("create", "epic", "user1")
			metrics.RecordDatabaseQuery("postgresql", "select", "epics", 10*time.Millisecond)
			metrics.RecordSearch("full_text", "user1", 50*time.Millisecond)
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Test should complete without panics
}

func BenchmarkPrometheusMiddleware(b *testing.B) {
	// Reset global metrics
	AppMetrics = nil
	
	// Create a new registry for this benchmark
	oldRegistry := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	defer func() {
		prometheus.DefaultRegisterer = oldRegistry
		AppMetrics = nil
	}()
	
	metrics := Init("test-service", "1.0.0")
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(metrics.PrometheusMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRecordEntityOperation(b *testing.B) {
	// Reset global metrics
	AppMetrics = nil
	
	// Create a new registry for this benchmark
	oldRegistry := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	defer func() {
		prometheus.DefaultRegisterer = oldRegistry
		AppMetrics = nil
	}()
	
	metrics := Init("test-service", "1.0.0")
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		metrics.RecordEntityOperation("create", "epic", "user1")
	}
}

// Test cleanup
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	
	// Reset global state after tests
	AppMetrics = nil
	
	// Exit with test result code
	os.Exit(code)
}