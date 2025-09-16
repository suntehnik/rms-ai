package logger

import (
	"context"
	"os"
	"product-requirements-management/internal/config"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

// CorrelationIDKey is the context key for correlation IDs
type CorrelationIDKey struct{}

// InitializationStepKey is the context key for initialization steps
type InitializationStepKey struct{}

// Init initializes the logger with the given configuration
func Init(cfg *config.LogConfig) {
	Logger = logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		Logger.Warn("Invalid log level, defaulting to info")
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)

	// Set log format
	if cfg.Format == "json" {
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	} else {
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	}

	// Set output to stdout
	Logger.SetOutput(os.Stdout)
}

// WithFields creates a new logger entry with the given fields
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Logger.WithFields(fields)
}

// WithField creates a new logger entry with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	return Logger.WithField(key, value)
}

// Info logs an info message
func Info(args ...interface{}) {
	Logger.Info(args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	Logger.Error(args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}

// NewCorrelationID generates a new correlation ID for tracking operations
func NewCorrelationID() string {
	return uuid.New().String()
}

// WithCorrelationID creates a context with a correlation ID
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey{}, correlationID)
}

// GetCorrelationID retrieves the correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(CorrelationIDKey{}).(string); ok {
		return id
	}
	return ""
}

// WithInitializationStep creates a context with an initialization step
func WithInitializationStep(ctx context.Context, step string) context.Context {
	return context.WithValue(ctx, InitializationStepKey{}, step)
}

// GetInitializationStep retrieves the initialization step from context
func GetInitializationStep(ctx context.Context) string {
	if step, ok := ctx.Value(InitializationStepKey{}).(string); ok {
		return step
	}
	return ""
}

// WithContext creates a logger entry with context information (correlation ID, step)
func WithContext(ctx context.Context) *logrus.Entry {
	entry := Logger.WithFields(logrus.Fields{})

	if correlationID := GetCorrelationID(ctx); correlationID != "" {
		entry = entry.WithField("correlation_id", correlationID)
	}

	if step := GetInitializationStep(ctx); step != "" {
		entry = entry.WithField("step", step)
	}

	return entry
}

// WithContextAndFields creates a logger entry with context and additional fields
func WithContextAndFields(ctx context.Context, fields logrus.Fields) *logrus.Entry {
	entry := WithContext(ctx)
	return entry.WithFields(fields)
}
