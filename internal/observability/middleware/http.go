package middleware

import (
	"product-requirements-management/internal/logger"
	"product-requirements-management/internal/observability/metrics"
	"product-requirements-management/internal/observability/tracing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

// ObservabilityMiddleware creates a comprehensive observability middleware
func ObservabilityMiddleware(m *metrics.Metrics, t *tracing.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Generate or extract correlation ID
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}
		c.Header("X-Correlation-ID", correlationID)

		// Generate request ID if not present
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Header("X-Request-ID", requestID)

		// Store IDs in context for use by other middleware/handlers
		c.Set("correlation_id", correlationID)
		c.Set("request_id", requestID)

		// Start tracing span if tracer is available
		var span trace.Span
		if t != nil {
			ctx, tracingSpan := t.StartHTTPSpan(c.Request.Context(), c.Request.Method, c.FullPath())
			c.Request = c.Request.WithContext(ctx)
			span = tracingSpan

			// Add correlation and request IDs to span
			tracing.AddSpanAttributes(tracingSpan, map[string]interface{}{
				"correlation_id": correlationID,
				"request_id":     requestID,
				"user_agent":     c.Request.UserAgent(),
				"remote_addr":    c.ClientIP(),
			})
		}

		// Increment in-flight requests metric
		if m != nil {
			m.HTTPRequestsInFlight.WithLabelValues(c.Request.Method, c.FullPath()).Inc()
		}

		// Create structured logger with correlation context
		logEntry := logger.WithFields(logrus.Fields{
			"correlation_id": correlationID,
			"request_id":     requestID,
			"method":         c.Request.Method,
			"path":           c.Request.URL.Path,
			"client_ip":      c.ClientIP(),
			"user_agent":     c.Request.UserAgent(),
		})

		// Store logger in context
		c.Set("logger", logEntry)

		// Log request start
		logEntry.Info("HTTP request started")

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Decrement in-flight requests metric
		if m != nil {
			m.HTTPRequestsInFlight.WithLabelValues(c.Request.Method, c.FullPath()).Dec()
		}

		// Record metrics
		if m != nil {
			statusCode := c.Writer.Status()
			m.HTTPRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), string(rune(statusCode))).Inc()
			m.HTTPRequestDuration.WithLabelValues(c.Request.Method, c.FullPath(), string(rune(statusCode))).Observe(duration.Seconds())
			m.HTTPResponseSize.WithLabelValues(c.Request.Method, c.FullPath(), string(rune(statusCode))).Observe(float64(c.Writer.Size()))
		}

		// Update tracing span
		if span != nil && t != nil {
			// This is a bit hacky, but we need to cast back to the proper span type
			// In a real implementation, you'd want to store the span in the context properly
		}

		// Log request completion
		logEntry.WithFields(logrus.Fields{
			"status":        c.Writer.Status(),
			"duration_ms":   duration.Milliseconds(),
			"response_size": c.Writer.Size(),
		}).Info("HTTP request completed")

		// End tracing span
		if span != nil {
			span.End()
		}
	}
}

// GetCorrelationID extracts correlation ID from Gin context
func GetCorrelationID(c *gin.Context) string {
	if id, exists := c.Get("correlation_id"); exists {
		if correlationID, ok := id.(string); ok {
			return correlationID
		}
	}
	return ""
}

// GetRequestID extracts request ID from Gin context
func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get("request_id"); exists {
		if requestID, ok := id.(string); ok {
			return requestID
		}
	}
	return ""
}

// GetLogger extracts the structured logger from Gin context
func GetLogger(c *gin.Context) *logrus.Entry {
	if loggerInterface, exists := c.Get("logger"); exists {
		if logger, ok := loggerInterface.(*logrus.Entry); ok {
			return logger
		}
	}
	// Fallback to default logger
	return logger.WithField("context", "unknown")
}