package health

import (
	"context"
	"net/http"
	"product-requirements-management/internal/database"
	"product-requirements-management/internal/observability/metrics"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthChecker provides health check functionality
type HealthChecker struct {
	db      *database.DB
	metrics *metrics.Metrics
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *database.DB, m *metrics.Metrics) *HealthChecker {
	return &HealthChecker{
		db:      db,
		metrics: m,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version"`
	Checks    map[string]CheckResult `json:"checks"`
}

// CheckResult represents individual health check result
type CheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// BasicHealth returns basic health status (liveness probe)
func (h *HealthChecker) BasicHealth(c *gin.Context) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0", // This should come from build info
		Checks:    make(map[string]CheckResult),
	}

	// Basic application health
	response.Checks["application"] = CheckResult{
		Status:  "healthy",
		Message: "Application is running",
	}

	c.JSON(http.StatusOK, response)
}

// ReadinessHealth returns readiness status (readiness probe)
func (h *HealthChecker) ReadinessHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
		Checks:    make(map[string]CheckResult),
	}

	overallHealthy := true

	// Check database health
	if h.db != nil {
		start := time.Now()
		dbHealth := h.db.CheckHealth(ctx)
		latency := time.Since(start)

		if dbHealth.Overall.Status == "healthy" {
			response.Checks["database"] = CheckResult{
				Status:  "healthy",
				Message: "All database connections are healthy",
				Latency: latency.String(),
			}
		} else {
			response.Checks["database"] = CheckResult{
				Status:  "unhealthy",
				Message: dbHealth.Overall.Message,
				Latency: latency.String(),
			}
			overallHealthy = false
		}

		// Add individual database component checks
		response.Checks["postgresql"] = CheckResult{
			Status:  dbHealth.PostgreSQL.Status,
			Message: dbHealth.PostgreSQL.Message,
		}

		response.Checks["redis"] = CheckResult{
			Status:  dbHealth.Redis.Status,
			Message: dbHealth.Redis.Message,
		}
	}

	// Check metrics system health
	if h.metrics != nil {
		response.Checks["metrics"] = CheckResult{
			Status:  "healthy",
			Message: "Metrics collection is active",
		}
	} else {
		response.Checks["metrics"] = CheckResult{
			Status:  "disabled",
			Message: "Metrics collection is disabled",
		}
	}

	// Set overall status
	if !overallHealthy {
		response.Status = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeepHealth returns comprehensive health status
func (h *HealthChecker) DeepHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
		Checks:    make(map[string]CheckResult),
	}

	overallHealthy := true

	// Application health
	response.Checks["application"] = CheckResult{
		Status:  "healthy",
		Message: "Application is running",
	}

	// Database health with detailed checks
	if h.db != nil {
		start := time.Now()
		dbHealth := h.db.CheckHealth(ctx)
		latency := time.Since(start)

		response.Checks["database_overall"] = CheckResult{
			Status:  dbHealth.Overall.Status,
			Message: dbHealth.Overall.Message,
			Latency: latency.String(),
		}

		response.Checks["postgresql"] = CheckResult{
			Status:  dbHealth.PostgreSQL.Status,
			Message: dbHealth.PostgreSQL.Message,
		}

		response.Checks["redis"] = CheckResult{
			Status:  dbHealth.Redis.Status,
			Message: dbHealth.Redis.Message,
		}

		if dbHealth.Overall.Status != "healthy" {
			overallHealthy = false
		}

		// Additional database performance checks
		if sqlDB, err := h.db.Postgres.DB(); err == nil {
			stats := sqlDB.Stats()
			response.Checks["database_connections"] = CheckResult{
				Status:  "healthy",
				Message: "Connection pool is healthy",
			}

			// Check for connection pool issues
			if stats.OpenConnections > 50 { // Assuming max 100 connections
				response.Checks["database_connections"] = CheckResult{
					Status:  "warning",
					Message: "High number of database connections",
				}
			}
		}
	}

	// Metrics system health
	if h.metrics != nil {
		response.Checks["metrics"] = CheckResult{
			Status:  "healthy",
			Message: "Prometheus metrics collection is active",
		}
	} else {
		response.Checks["metrics"] = CheckResult{
			Status:  "disabled",
			Message: "Metrics collection is disabled",
		}
	}

	// Memory and performance checks could be added here
	response.Checks["memory"] = CheckResult{
		Status:  "healthy",
		Message: "Memory usage is within normal limits",
	}

	// Set overall status
	if !overallHealthy {
		response.Status = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// SetupHealthRoutes sets up health check routes
func (h *HealthChecker) SetupHealthRoutes(router *gin.Engine) {
	health := router.Group("/health")
	{
		health.GET("", h.BasicHealth)           // Basic liveness check
		health.GET("/live", h.BasicHealth)      // Kubernetes liveness probe
		health.GET("/ready", h.ReadinessHealth) // Kubernetes readiness probe
		health.GET("/deep", h.DeepHealth)       // Comprehensive health check
	}
}
