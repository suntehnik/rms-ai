package middleware

import (
	"product-requirements-management/internal/observability/metrics"
	"product-requirements-management/internal/observability/tracing"
	"time"

	"gorm.io/gorm"
)

// DatabaseMetricsPlugin is a GORM plugin that records database metrics
type DatabaseMetricsPlugin struct {
	metrics *metrics.Metrics
	tracer  *tracing.Tracer
}

// NewDatabaseMetricsPlugin creates a new database metrics plugin
func NewDatabaseMetricsPlugin(m *metrics.Metrics, t *tracing.Tracer) *DatabaseMetricsPlugin {
	return &DatabaseMetricsPlugin{
		metrics: m,
		tracer:  t,
	}
}

// Name returns the plugin name
func (p *DatabaseMetricsPlugin) Name() string {
	return "database_metrics"
}

// Initialize initializes the plugin
func (p *DatabaseMetricsPlugin) Initialize(db *gorm.DB) error {
	// Register callbacks for different operations
	if err := p.registerCallbacks(db); err != nil {
		return err
	}

	// Start connection monitoring
	p.startConnectionMonitoring(db)

	return nil
}

// registerCallbacks registers GORM callbacks for metrics collection
func (p *DatabaseMetricsPlugin) registerCallbacks(db *gorm.DB) error {
	// Create callback
	if err := db.Callback().Create().Before("gorm:create").Register("metrics:before_create", p.beforeCallback("create")); err != nil {
		return err
	}
	if err := db.Callback().Create().After("gorm:create").Register("metrics:after_create", p.afterCallback("create")); err != nil {
		return err
	}

	// Query callback
	if err := db.Callback().Query().Before("gorm:query").Register("metrics:before_query", p.beforeCallback("select")); err != nil {
		return err
	}
	if err := db.Callback().Query().After("gorm:query").Register("metrics:after_query", p.afterCallback("select")); err != nil {
		return err
	}

	// Update callback
	if err := db.Callback().Update().Before("gorm:update").Register("metrics:before_update", p.beforeCallback("update")); err != nil {
		return err
	}
	if err := db.Callback().Update().After("gorm:update").Register("metrics:after_update", p.afterCallback("update")); err != nil {
		return err
	}

	// Delete callback
	if err := db.Callback().Delete().Before("gorm:delete").Register("metrics:before_delete", p.beforeCallback("delete")); err != nil {
		return err
	}
	if err := db.Callback().Delete().After("gorm:delete").Register("metrics:after_delete", p.afterCallback("delete")); err != nil {
		return err
	}

	return nil
}

// beforeCallback creates a callback that runs before database operations
func (p *DatabaseMetricsPlugin) beforeCallback(operation string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		startTime := time.Now()
		db.Set("metrics:start_time", startTime)
		db.Set("metrics:operation", operation)

		// Start tracing span if tracer is available
		if p.tracer != nil {
			table := db.Statement.Table
			if table == "" && db.Statement.Schema != nil {
				table = db.Statement.Schema.Table
			}

			ctx, span := p.tracer.StartDatabaseSpan(db.Statement.Context, operation, table)
			db.Statement.Context = ctx
			db.Set("metrics:span", span)
		}
	}
}

// afterCallback creates a callback that runs after database operations
func (p *DatabaseMetricsPlugin) afterCallback(operation string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		// Get start time
		startTimeInterface, exists := db.Get("metrics:start_time")
		if !exists {
			return
		}
		startTime, ok := startTimeInterface.(time.Time)
		if !ok {
			return
		}

		duration := time.Since(startTime)
		table := db.Statement.Table
		if table == "" && db.Statement.Schema != nil {
			table = db.Statement.Schema.Table
		}

		// Record metrics if available
		if p.metrics != nil {
			p.metrics.RecordDatabaseQuery("postgresql", operation, table, duration)
		}

		// End tracing span if available
		if spanInterface, exists := db.Get("metrics:span"); exists {
			if span, ok := spanInterface.(interface{ End() }); ok {
				// Add error information if there was an error
				if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
					if tracingSpan, ok := spanInterface.(interface {
						RecordError(error)
						SetStatus(interface{}, string)
					}); ok {
						tracingSpan.RecordError(db.Error)
						tracingSpan.SetStatus("error", db.Error.Error())
					}
				}
				span.End()
			}
		}
	}
}

// startConnectionMonitoring starts monitoring database connections
func (p *DatabaseMetricsPlugin) startConnectionMonitoring(db *gorm.DB) {
	if p.metrics == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			sqlDB, err := db.DB()
			if err != nil {
				continue
			}

			stats := sqlDB.Stats()
			p.metrics.RecordDatabaseConnection("postgresql", "open", float64(stats.OpenConnections))
			p.metrics.RecordDatabaseConnection("postgresql", "idle", float64(stats.Idle))
			p.metrics.RecordDatabaseConnection("postgresql", "in_use", float64(stats.InUse))
		}
	}()
}

// Note: Custom database logger implementation removed for simplicity
// In production, you might want to implement a custom GORM logger
// that integrates with your observability system
