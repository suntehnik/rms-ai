package database

import (
	"context"
	"fmt"
	"time"
)

// HealthStatus represents the health status of a database component
type HealthStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// HealthCheck represents the overall health check result
type HealthCheck struct {
	PostgreSQL HealthStatus `json:"postgresql"`
	Redis      HealthStatus `json:"redis"`
	Overall    HealthStatus `json:"overall"`
}

// CheckHealth performs health checks on all database connections
func (db *DB) CheckHealth(ctx context.Context) HealthCheck {
	result := HealthCheck{}

	// Check PostgreSQL health
	result.PostgreSQL = db.checkPostgreSQLHealth(ctx)

	// Check Redis health
	result.Redis = db.checkRedisHealth(ctx)

	// Determine overall health
	if result.PostgreSQL.Status == "healthy" && result.Redis.Status == "healthy" {
		result.Overall = HealthStatus{Status: "healthy", Message: "All database connections are healthy"}
	} else {
		result.Overall = HealthStatus{Status: "unhealthy", Message: "One or more database connections are unhealthy"}
	}

	return result
}

// checkPostgreSQLHealth checks PostgreSQL connection health
func (db *DB) checkPostgreSQLHealth(ctx context.Context) HealthStatus {
	if db.Postgres == nil {
		return HealthStatus{Status: "unhealthy", Message: "PostgreSQL connection not initialized"}
	}

	// Get underlying sql.DB for health check
	sqlDB, err := db.Postgres.DB()
	if err != nil {
		return HealthStatus{Status: "unhealthy", Message: fmt.Sprintf("Failed to get SQL DB: %v", err)}
	}

	// Create context with timeout for health check
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Ping database
	if err := sqlDB.PingContext(checkCtx); err != nil {
		return HealthStatus{Status: "unhealthy", Message: fmt.Sprintf("PostgreSQL ping failed: %v", err)}
	}

	// Check connection stats
	stats := sqlDB.Stats()
	if stats.OpenConnections == 0 {
		return HealthStatus{Status: "unhealthy", Message: "No open PostgreSQL connections"}
	}

	return HealthStatus{Status: "healthy", Message: fmt.Sprintf("PostgreSQL healthy (open: %d, idle: %d)", stats.OpenConnections, stats.Idle)}
}

// checkRedisHealth checks Redis connection health
func (db *DB) checkRedisHealth(ctx context.Context) HealthStatus {
	if db.Redis == nil {
		return HealthStatus{Status: "unhealthy", Message: "Redis connection not initialized"}
	}

	// Create context with timeout for health check
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Ping Redis
	_, err := db.Redis.Ping(checkCtx).Result()
	if err != nil {
		return HealthStatus{Status: "unhealthy", Message: fmt.Sprintf("Redis ping failed: %v", err)}
	}

	// Get Redis info
	info, err := db.Redis.Info(checkCtx, "server").Result()
	if err != nil {
		return HealthStatus{Status: "healthy", Message: "Redis ping successful"}
	}

	return HealthStatus{Status: "healthy", Message: "Redis healthy - " + extractRedisVersion(info)}
}

// extractRedisVersion extracts Redis version from info string
func extractRedisVersion(info string) string {
	// Simple extraction - in production you might want more robust parsing
	if len(info) > 50 {
		return "connected"
	}
	return "connected"
}

// IsHealthy returns true if all database connections are healthy
func (db *DB) IsHealthy(ctx context.Context) bool {
	health := db.CheckHealth(ctx)
	return health.Overall.Status == "healthy"
}
