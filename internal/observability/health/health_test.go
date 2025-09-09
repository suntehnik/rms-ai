package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"product-requirements-management/internal/observability/metrics"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHealthChecker(t *testing.T) {
	metrics := metrics.Init("test-service", "1.0.0")
	
	hc := NewHealthChecker(nil, metrics)
	assert.NotNil(t, hc)
	assert.Nil(t, hc.db)
	assert.Equal(t, metrics, hc.metrics)
}

func TestBasicHealth(t *testing.T) {
	hc := NewHealthChecker(nil, nil)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", hc.BasicHealth)
	
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "healthy", response.Status)
	assert.NotEmpty(t, response.Timestamp)
	assert.NotEmpty(t, response.Version)
	assert.Contains(t, response.Checks, "application")
	assert.Equal(t, "healthy", response.Checks["application"].Status)
}

func TestReadinessHealth_NoDatabase(t *testing.T) {
	// Create a new registry for this test
	oldRegistry := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	defer func() {
		prometheus.DefaultRegisterer = oldRegistry
	}()
	
	metrics := metrics.Init("test-service", "1.0.0")
	hc := NewHealthChecker(nil, metrics)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ready", hc.ReadinessHealth)
	
	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "healthy", response.Status)
	assert.Contains(t, response.Checks, "metrics")
	assert.Equal(t, "healthy", response.Checks["metrics"].Status)
}

func TestReadinessHealth_NoMetrics(t *testing.T) {
	hc := NewHealthChecker(nil, nil)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ready", hc.ReadinessHealth)
	
	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "healthy", response.Status)
	assert.Contains(t, response.Checks, "metrics")
	assert.Equal(t, "disabled", response.Checks["metrics"].Status)
}

func TestDeepHealth_NoDatabase(t *testing.T) {
	metrics := metrics.Init("test-service", "1.0.0")
	hc := NewHealthChecker(nil, metrics)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health/deep", hc.DeepHealth)
	
	req := httptest.NewRequest("GET", "/health/deep", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "healthy", response.Status)
	assert.Contains(t, response.Checks, "application")
	assert.Contains(t, response.Checks, "metrics")
	assert.Contains(t, response.Checks, "memory")
	assert.Equal(t, "healthy", response.Checks["application"].Status)
	assert.Equal(t, "healthy", response.Checks["metrics"].Status)
	assert.Equal(t, "healthy", response.Checks["memory"].Status)
}

func TestSetupHealthRoutes(t *testing.T) {
	hc := NewHealthChecker(nil, nil)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	hc.SetupHealthRoutes(router)
	
	// Test all health endpoints
	endpoints := []struct {
		path           string
		expectedStatus int
	}{
		{"/health", http.StatusOK},
		{"/health/live", http.StatusOK},
		{"/health/ready", http.StatusOK},
		{"/health/deep", http.StatusOK},
	}
	
	for _, endpoint := range endpoints {
		t.Run(endpoint.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", endpoint.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, endpoint.expectedStatus, w.Code)
			
			var response HealthResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			
			assert.NotEmpty(t, response.Status)
			assert.NotEmpty(t, response.Timestamp)
			assert.NotEmpty(t, response.Checks)
		})
	}
}

func TestHealthResponseStructure(t *testing.T) {
	hc := NewHealthChecker(nil, nil)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", hc.BasicHealth)
	
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	// Verify response structure
	assert.NotEmpty(t, response.Status)
	assert.NotEmpty(t, response.Timestamp)
	assert.NotEmpty(t, response.Version)
	assert.NotNil(t, response.Checks)
	
	// Verify timestamp format
	_, err = time.Parse(time.RFC3339, response.Timestamp)
	assert.NoError(t, err, "Timestamp should be in RFC3339 format")
}

func TestCheckResultStructure(t *testing.T) {
	hc := NewHealthChecker(nil, nil)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", hc.BasicHealth)
	
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	// Verify check result structure
	for checkName, checkResult := range response.Checks {
		t.Run(checkName, func(t *testing.T) {
			assert.NotEmpty(t, checkResult.Status)
			// Message and Latency are optional
		})
	}
}

func TestHealthTimeout(t *testing.T) {
	hc := NewHealthChecker(nil, nil)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ready", hc.ReadinessHealth)
	
	// Create a request with a very short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	
	req := httptest.NewRequest("GET", "/ready", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	
	// This should still complete successfully since we don't have a real database
	router.ServeHTTP(w, req)
	
	// The health check should still work even with a cancelled context
	// because it creates its own timeout context internally
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConcurrentHealthChecks(t *testing.T) {
	hc := NewHealthChecker(nil, nil)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	hc.SetupHealthRoutes(router)
	
	// Test concurrent health check requests
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			
			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, http.StatusOK, w.Code)
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func BenchmarkBasicHealth(b *testing.B) {
	hc := NewHealthChecker(nil, nil)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", hc.BasicHealth)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkReadinessHealth(b *testing.B) {
	hc := NewHealthChecker(nil, nil)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ready", hc.ReadinessHealth)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}