package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal     *prometheus.CounterVec
	HTTPRequestDuration   *prometheus.HistogramVec
	HTTPRequestsInFlight  *prometheus.GaugeVec
	HTTPResponseSize      *prometheus.HistogramVec

	// Database metrics
	DatabaseConnections   *prometheus.GaugeVec
	DatabaseQueries       *prometheus.CounterVec
	DatabaseQueryDuration *prometheus.HistogramVec

	// Business metrics
	EntitiesTotal         *prometheus.CounterVec
	EntitiesCreated       *prometheus.CounterVec
	EntitiesUpdated       *prometheus.CounterVec
	EntitiesDeleted       *prometheus.CounterVec
	CommentsTotal         *prometheus.CounterVec
	SearchQueries         *prometheus.CounterVec
	SearchDuration        *prometheus.HistogramVec

	// System metrics
	ApplicationInfo       *prometheus.GaugeVec
	ApplicationUptime     *prometheus.CounterVec
}

var (
	// Global metrics instance
	AppMetrics *Metrics
)

// Init initializes Prometheus metrics
func Init(serviceName, version string) *Metrics {
	metrics := &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "status_code"},
		),
		HTTPRequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
			[]string{"method", "endpoint"},
		),
		HTTPResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "Size of HTTP responses in bytes",
				Buckets: []float64{100, 1000, 10000, 100000, 1000000},
			},
			[]string{"method", "endpoint", "status_code"},
		),

		// Database metrics
		DatabaseConnections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "database_connections",
				Help: "Number of database connections",
			},
			[]string{"database", "state"},
		),
		DatabaseQueries: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "database_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"database", "operation", "table"},
		),
		DatabaseQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "database_query_duration_seconds",
				Help:    "Duration of database queries in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0},
			},
			[]string{"database", "operation", "table"},
		),

		// Business metrics
		EntitiesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "entities_total",
				Help: "Total number of entities in the system",
			},
			[]string{"entity_type"},
		),
		EntitiesCreated: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "entities_created_total",
				Help: "Total number of entities created",
			},
			[]string{"entity_type", "user_id"},
		),
		EntitiesUpdated: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "entities_updated_total",
				Help: "Total number of entities updated",
			},
			[]string{"entity_type", "user_id"},
		),
		EntitiesDeleted: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "entities_deleted_total",
				Help: "Total number of entities deleted",
			},
			[]string{"entity_type", "user_id"},
		),
		CommentsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "comments_total",
				Help: "Total number of comments",
			},
			[]string{"entity_type", "comment_type", "status"},
		),
		SearchQueries: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "search_queries_total",
				Help: "Total number of search queries",
			},
			[]string{"search_type", "user_id"},
		),
		SearchDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "search_duration_seconds",
				Help:    "Duration of search queries in seconds",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0},
			},
			[]string{"search_type"},
		),

		// System metrics
		ApplicationInfo: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "application_info",
				Help: "Application information",
			},
			[]string{"service_name", "version", "go_version"},
		),
		ApplicationUptime: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "application_uptime_seconds_total",
				Help: "Total application uptime in seconds",
			},
			[]string{"service_name"},
		),
	}

	// Set application info
	metrics.ApplicationInfo.WithLabelValues(serviceName, version, "go1.24.5").Set(1)

	// Store global reference
	AppMetrics = metrics

	return metrics
}

// PrometheusMiddleware returns a Gin middleware for Prometheus metrics collection
func (m *Metrics) PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Increment in-flight requests
		m.HTTPRequestsInFlight.WithLabelValues(c.Request.Method, c.FullPath()).Inc()
		
		// Process request
		c.Next()
		
		// Decrement in-flight requests
		m.HTTPRequestsInFlight.WithLabelValues(c.Request.Method, c.FullPath()).Dec()
		
		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())
		
		m.HTTPRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), statusCode).Inc()
		m.HTTPRequestDuration.WithLabelValues(c.Request.Method, c.FullPath(), statusCode).Observe(duration)
		m.HTTPResponseSize.WithLabelValues(c.Request.Method, c.FullPath(), statusCode).Observe(float64(c.Writer.Size()))
	}
}

// RecordDatabaseConnection records database connection metrics
func (m *Metrics) RecordDatabaseConnection(database, state string, count float64) {
	m.DatabaseConnections.WithLabelValues(database, state).Set(count)
}

// RecordDatabaseQuery records database query metrics
func (m *Metrics) RecordDatabaseQuery(database, operation, table string, duration time.Duration) {
	m.DatabaseQueries.WithLabelValues(database, operation, table).Inc()
	m.DatabaseQueryDuration.WithLabelValues(database, operation, table).Observe(duration.Seconds())
}

// RecordEntityOperation records entity operation metrics
func (m *Metrics) RecordEntityOperation(operation, entityType, userID string) {
	switch operation {
	case "create":
		m.EntitiesCreated.WithLabelValues(entityType, userID).Inc()
	case "update":
		m.EntitiesUpdated.WithLabelValues(entityType, userID).Inc()
	case "delete":
		m.EntitiesDeleted.WithLabelValues(entityType, userID).Inc()
	}
}

// RecordComment records comment metrics
func (m *Metrics) RecordComment(entityType, commentType, status string) {
	m.CommentsTotal.WithLabelValues(entityType, commentType, status).Inc()
}

// RecordSearch records search metrics
func (m *Metrics) RecordSearch(searchType, userID string, duration time.Duration) {
	m.SearchQueries.WithLabelValues(searchType, userID).Inc()
	m.SearchDuration.WithLabelValues(searchType).Observe(duration.Seconds())
}

// RecordUptime records application uptime
func (m *Metrics) RecordUptime(serviceName string, uptime time.Duration) {
	m.ApplicationUptime.WithLabelValues(serviceName).Add(uptime.Seconds())
}