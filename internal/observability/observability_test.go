package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit_AllDisabled(t *testing.T) {
	ctx := context.Background()
	config := Config{
		ServiceName:     "test-service",
		ServiceVersion:  "1.0.0",
		Environment:     "test",
		MetricsEnabled:  false,
		TracingEnabled:  false,
		TracingEndpoint: "http://localhost:4318/v1/traces",
	}

	obs, err := Init(ctx, config)
	require.NoError(t, err)
	assert.NotNil(t, obs)
	assert.Nil(t, obs.Metrics)
	assert.Nil(t, obs.Tracer)
}

func TestInit_MetricsOnly(t *testing.T) {
	// Create a new registry for this test
	oldRegistry := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	defer func() {
		prometheus.DefaultRegisterer = oldRegistry
	}()
	
	ctx := context.Background()
	config := Config{
		ServiceName:     "test-service",
		ServiceVersion:  "1.0.0",
		Environment:     "test",
		MetricsEnabled:  true,
		TracingEnabled:  false,
		TracingEndpoint: "http://localhost:4318/v1/traces",
	}

	obs, err := Init(ctx, config)
	require.NoError(t, err)
	assert.NotNil(t, obs)
	assert.NotNil(t, obs.Metrics)
	assert.Nil(t, obs.Tracer)
}

func TestInit_TracingOnly(t *testing.T) {
	ctx := context.Background()
	config := Config{
		ServiceName:     "test-service",
		ServiceVersion:  "1.0.0",
		Environment:     "test",
		MetricsEnabled:  false,
		TracingEnabled:  true,
		TracingEndpoint: "http://localhost:4318/v1/traces",
	}

	obs, err := Init(ctx, config)
	
	// Note: This test might fail if no OTLP receiver is running
	if err != nil {
		t.Skipf("Skipping test due to OTLP endpoint not available: %v", err)
	}
	
	require.NoError(t, err)
	assert.NotNil(t, obs)
	assert.Nil(t, obs.Metrics)
	assert.NotNil(t, obs.Tracer)
	
	// Clean up
	_ = obs.Shutdown(ctx)
}

func TestInit_AllEnabled(t *testing.T) {
	// Create a new registry for this test
	oldRegistry := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	defer func() {
		prometheus.DefaultRegisterer = oldRegistry
	}()
	
	ctx := context.Background()
	config := Config{
		ServiceName:     "test-service",
		ServiceVersion:  "1.0.0",
		Environment:     "test",
		MetricsEnabled:  true,
		TracingEnabled:  true,
		TracingEndpoint: "http://localhost:4318/v1/traces",
	}

	obs, err := Init(ctx, config)
	
	// Note: This test might fail if no OTLP receiver is running
	if err != nil {
		t.Skipf("Skipping test due to OTLP endpoint not available: %v", err)
	}
	
	require.NoError(t, err)
	assert.NotNil(t, obs)
	assert.NotNil(t, obs.Metrics)
	assert.NotNil(t, obs.Tracer)
	
	// Clean up
	_ = obs.Shutdown(ctx)
}

func TestShutdown(t *testing.T) {
	ctx := context.Background()
	
	// Test shutdown with no components
	obs := &Observability{}
	err := obs.Shutdown(ctx)
	assert.NoError(t, err)
	
	// Test shutdown with metrics only
	// Create a new registry for this test
	oldRegistry := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	defer func() {
		prometheus.DefaultRegisterer = oldRegistry
	}()
	
	config := Config{
		ServiceName:    "test-service",
		MetricsEnabled: true,
		TracingEnabled: false,
	}
	
	obs, err = Init(ctx, config)
	require.NoError(t, err)
	
	err = obs.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestSetupMetricsEndpoint(t *testing.T) {
	// Create a new registry for this test
	oldRegistry := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	defer func() {
		prometheus.DefaultRegisterer = oldRegistry
	}()
	
	ctx := context.Background()
	config := Config{
		ServiceName:    "test-service",
		MetricsEnabled: true,
		TracingEnabled: false,
	}

	obs, err := Init(ctx, config)
	require.NoError(t, err)

	// Create test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Setup metrics endpoint
	obs.SetupMetricsEndpoint(router)
	
	// Test metrics endpoint
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	// Just verify we get a valid Prometheus response
	assert.Contains(t, w.Body.String(), "# HELP")
}

func TestSetupMetricsEndpoint_Disabled(t *testing.T) {
	ctx := context.Background()
	config := Config{
		ServiceName:    "test-service",
		MetricsEnabled: false,
		TracingEnabled: false,
	}

	obs, err := Init(ctx, config)
	require.NoError(t, err)

	// Create test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Setup metrics endpoint (should not add route when disabled)
	obs.SetupMetricsEndpoint(router)
	
	// Test metrics endpoint (should return 404)
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetMiddleware_AllDisabled(t *testing.T) {
	ctx := context.Background()
	config := Config{
		ServiceName:    "test-service",
		MetricsEnabled: false,
		TracingEnabled: false,
	}

	obs, err := Init(ctx, config)
	require.NoError(t, err)

	middleware := obs.GetMiddleware()
	assert.Empty(t, middleware)
}

func TestGetMiddleware_MetricsOnly(t *testing.T) {
	// Create a new registry for this test
	oldRegistry := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	defer func() {
		prometheus.DefaultRegisterer = oldRegistry
	}()
	
	ctx := context.Background()
	config := Config{
		ServiceName:    "test-service",
		MetricsEnabled: true,
		TracingEnabled: false,
	}

	obs, err := Init(ctx, config)
	require.NoError(t, err)

	middleware := obs.GetMiddleware()
	assert.Len(t, middleware, 1) // Only metrics middleware
}

func TestGetMiddleware_TracingOnly(t *testing.T) {
	ctx := context.Background()
	config := Config{
		ServiceName:     "test-service",
		MetricsEnabled:  false,
		TracingEnabled:  true,
		TracingEndpoint: "http://localhost:4318/v1/traces",
	}

	obs, err := Init(ctx, config)
	
	// Skip if tracing initialization fails
	if err != nil {
		t.Skipf("Skipping test due to OTLP endpoint not available: %v", err)
	}
	
	require.NoError(t, err)

	middleware := obs.GetMiddleware()
	assert.Len(t, middleware, 1) // Only tracing middleware
	
	// Clean up
	_ = obs.Shutdown(ctx)
}

func TestGetMiddleware_AllEnabled(t *testing.T) {
	// Create a new registry for this test
	oldRegistry := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	defer func() {
		prometheus.DefaultRegisterer = oldRegistry
	}()
	
	ctx := context.Background()
	config := Config{
		ServiceName:     "test-service",
		MetricsEnabled:  true,
		TracingEnabled:  true,
		TracingEndpoint: "http://localhost:4318/v1/traces",
	}

	obs, err := Init(ctx, config)
	
	// Skip if tracing initialization fails
	if err != nil {
		t.Skipf("Skipping test due to OTLP endpoint not available: %v", err)
	}
	
	require.NoError(t, err)

	middleware := obs.GetMiddleware()
	assert.Len(t, middleware, 2) // Both metrics and tracing middleware
	
	// Clean up
	_ = obs.Shutdown(ctx)
}

func TestTracingMiddleware(t *testing.T) {
	ctx := context.Background()
	config := Config{
		ServiceName:     "test-service",
		MetricsEnabled:  false,
		TracingEnabled:  true,
		TracingEndpoint: "http://localhost:4318/v1/traces",
	}

	obs, err := Init(ctx, config)
	
	// Skip if tracing initialization fails
	if err != nil {
		t.Skipf("Skipping test due to OTLP endpoint not available: %v", err)
	}
	
	require.NoError(t, err)
	defer obs.Shutdown(ctx)

	// Create test router with tracing middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	middleware := obs.GetMiddleware()
	for _, mw := range middleware {
		router.Use(mw)
	}
	
	// Add test endpoint
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	// Test successful request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Add test endpoint that returns error
	router.GET("/error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "test error"})
	})
	
	// Test error request
	req = httptest.NewRequest("GET", "/error", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRecordUptime(t *testing.T) {
	// Create a new registry for this test
	oldRegistry := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	defer func() {
		prometheus.DefaultRegisterer = oldRegistry
	}()
	
	ctx := context.Background()
	config := Config{
		ServiceName:    "test-service",
		MetricsEnabled: true,
		TracingEnabled: false,
	}

	obs, err := Init(ctx, config)
	require.NoError(t, err)

	startTime := time.Now().Add(-1 * time.Hour)
	
	// Should not panic
	assert.NotPanics(t, func() {
		obs.RecordUptime(startTime)
	})
}

func TestRecordUptime_Disabled(t *testing.T) {
	ctx := context.Background()
	config := Config{
		ServiceName:    "test-service",
		MetricsEnabled: false,
		TracingEnabled: false,
	}

	obs, err := Init(ctx, config)
	require.NoError(t, err)

	startTime := time.Now().Add(-1 * time.Hour)
	
	// Should not panic even when metrics are disabled
	assert.NotPanics(t, func() {
		obs.RecordUptime(startTime)
	})
}

func TestStartUptimeRecording(t *testing.T) {
	// Create a new registry for this test
	oldRegistry := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	defer func() {
		prometheus.DefaultRegisterer = oldRegistry
	}()
	
	ctx := context.Background()
	config := Config{
		ServiceName:    "test-service",
		MetricsEnabled: true,
		TracingEnabled: false,
	}

	obs, err := Init(ctx, config)
	require.NoError(t, err)

	startTime := time.Now()
	
	// Create a context that will be cancelled
	uptimeCtx, cancel := context.WithCancel(ctx)
	
	// Start uptime recording
	obs.StartUptimeRecording(uptimeCtx, startTime)
	
	// Let it run briefly
	time.Sleep(100 * time.Millisecond)
	
	// Cancel the context to stop recording
	cancel()
	
	// Give it time to stop
	time.Sleep(100 * time.Millisecond)
	
	// Test should complete without hanging
}

func TestStartUptimeRecording_Disabled(t *testing.T) {
	ctx := context.Background()
	config := Config{
		ServiceName:    "test-service",
		MetricsEnabled: false,
		TracingEnabled: false,
	}

	obs, err := Init(ctx, config)
	require.NoError(t, err)

	startTime := time.Now()
	
	// Should not panic even when metrics are disabled
	assert.NotPanics(t, func() {
		obs.StartUptimeRecording(ctx, startTime)
	})
}

func BenchmarkTracingMiddleware(b *testing.B) {
	ctx := context.Background()
	config := Config{
		ServiceName:    "test-service",
		MetricsEnabled: false,
		TracingEnabled: false, // Use disabled tracing for consistent benchmarks
	}

	obs, _ := Init(ctx, config)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	middleware := obs.GetMiddleware()
	for _, mw := range middleware {
		router.Use(mw)
	}
	
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