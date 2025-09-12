package observability

import (
	"context"
	"fmt"
	"product-requirements-management/internal/observability/metrics"
	"product-requirements-management/internal/observability/tracing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/codes"
)

// Config holds observability configuration
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string

	// Metrics configuration
	MetricsEnabled bool
	MetricsPort    string

	// Tracing configuration
	TracingEnabled  bool
	TracingEndpoint string
}

// Observability holds all observability components
type Observability struct {
	Metrics *metrics.Metrics
	Tracer  *tracing.Tracer
	config  Config
}

// Init initializes all observability components
func Init(ctx context.Context, config Config) (*Observability, error) {
	obs := &Observability{
		config: config,
	}

	// Initialize metrics if enabled
	if config.MetricsEnabled {
		obs.Metrics = metrics.Init(config.ServiceName, config.ServiceVersion)
	}

	// Initialize tracing if enabled
	if config.TracingEnabled {
		tracingConfig := tracing.TracingConfig{
			ServiceName:    config.ServiceName,
			ServiceVersion: config.ServiceVersion,
			Environment:    config.Environment,
			Endpoint:       config.TracingEndpoint,
			Enabled:        config.TracingEnabled,
		}

		tracer, err := tracing.Init(ctx, tracingConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize tracing: %w", err)
		}
		obs.Tracer = tracer
	}

	return obs, nil
}

// Shutdown gracefully shuts down all observability components
func (o *Observability) Shutdown(ctx context.Context) error {
	if o.Tracer != nil {
		if err := o.Tracer.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown tracer: %w", err)
		}
	}
	return nil
}

// SetupMetricsEndpoint sets up the Prometheus metrics endpoint
func (o *Observability) SetupMetricsEndpoint(router *gin.Engine) {
	if o.Metrics != nil {
		router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}
}

// GetMiddleware returns observability middleware for Gin
func (o *Observability) GetMiddleware() []gin.HandlerFunc {
	var middleware []gin.HandlerFunc

	// Add metrics middleware if enabled
	if o.Metrics != nil {
		middleware = append(middleware, o.Metrics.PrometheusMiddleware())
	}

	// Add tracing middleware if enabled
	if o.Tracer != nil {
		middleware = append(middleware, o.tracingMiddleware())
	}

	return middleware
}

// tracingMiddleware returns a Gin middleware for OpenTelemetry tracing
func (o *Observability) tracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start HTTP span
		ctx, span := o.Tracer.StartHTTPSpan(c.Request.Context(), c.Request.Method, c.FullPath())

		// Add request attributes
		tracing.AddSpanAttributes(span, map[string]interface{}{
			"http.method":      c.Request.Method,
			"http.url":         c.Request.URL.String(),
			"http.user_agent":  c.Request.UserAgent(),
			"http.remote_addr": c.ClientIP(),
		})

		// Set context with span
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Add response attributes
		tracing.AddSpanAttributes(span, map[string]interface{}{
			"http.status_code":   c.Writer.Status(),
			"http.response_size": c.Writer.Size(),
		})

		// Set span status based on HTTP status code
		if c.Writer.Status() >= 400 {
			if c.Writer.Status() >= 500 {
				span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", c.Writer.Status()))
			} else {
				span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", c.Writer.Status()))
			}
		} else {
			span.SetStatus(codes.Ok, "")
		}

		// End span
		span.End()
	}
}

// RecordUptime records application uptime metrics
func (o *Observability) RecordUptime(startTime time.Time) {
	if o.Metrics != nil {
		uptime := time.Since(startTime)
		o.Metrics.RecordUptime(o.config.ServiceName, uptime)
	}
}

// StartUptimeRecording starts a goroutine to record uptime metrics periodically
func (o *Observability) StartUptimeRecording(ctx context.Context, startTime time.Time) {
	if o.Metrics == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				o.RecordUptime(startTime)
			}
		}
	}()
}
