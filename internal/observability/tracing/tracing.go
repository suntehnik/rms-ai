package tracing

import (
	"context"
	"fmt"
	"runtime"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingConfig holds tracing configuration
type TracingConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Endpoint       string
	Enabled        bool
}

// Tracer holds the OpenTelemetry tracer
type Tracer struct {
	tracer   trace.Tracer
	provider *sdktrace.TracerProvider
}

var (
	// Global tracer instance
	AppTracer *Tracer
)

// Init initializes OpenTelemetry tracing
func Init(ctx context.Context, config TracingConfig) (*Tracer, error) {
	if !config.Enabled {
		// Return a no-op tracer if tracing is disabled
		tracer := &Tracer{
			tracer: otel.Tracer(config.ServiceName),
		}

		// Store global reference
		AppTracer = tracer

		return tracer, nil
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
			semconv.TelemetrySDKLanguageGo,
			semconv.TelemetrySDKVersion("1.38.0"),
			semconv.ProcessRuntimeName("go"),
			semconv.ProcessRuntimeVersion(runtime.Version()),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP HTTP exporter
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(config.Endpoint),
		otlptracehttp.WithInsecure(), // Use HTTP instead of HTTPS for local development
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create trace provider
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Sample all traces in development
	)

	// Set global trace provider
	otel.SetTracerProvider(provider)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer
	tracer := &Tracer{
		tracer:   otel.Tracer(config.ServiceName),
		provider: provider,
	}

	// Store global reference
	AppTracer = tracer

	return tracer, nil
}

// Shutdown gracefully shuts down the tracer
func (t *Tracer) Shutdown(ctx context.Context) error {
	if t.provider != nil {
		return t.provider.Shutdown(ctx)
	}
	return nil
}

// StartSpan starts a new span with the given name
func (t *Tracer) StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, spanName, opts...)
}

// StartHTTPSpan starts a new span for HTTP requests
func (t *Tracer) StartHTTPSpan(ctx context.Context, method, path string) (context.Context, trace.Span) {
	spanName := fmt.Sprintf("%s %s", method, path)
	ctx, span := t.tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			semconv.HTTPMethod(method),
			semconv.HTTPRoute(path),
		),
	)
	return ctx, span
}

// StartDatabaseSpan starts a new span for database operations
func (t *Tracer) StartDatabaseSpan(ctx context.Context, operation, table string) (context.Context, trace.Span) {
	spanName := fmt.Sprintf("db.%s.%s", operation, table)
	ctx, span := t.tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.DBSystemPostgreSQL,
			attribute.String("db.operation", operation),
			attribute.String("db.sql.table", table),
		),
	)
	return ctx, span
}

// StartServiceSpan starts a new span for service operations
func (t *Tracer) StartServiceSpan(ctx context.Context, service, operation string) (context.Context, trace.Span) {
	spanName := fmt.Sprintf("%s.%s", service, operation)
	ctx, span := t.tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			semconv.CodeNamespace(service),
			semconv.CodeFunction(operation),
		),
	)
	return ctx, span
}

// AddSpanAttributes adds attributes to the current span
func AddSpanAttributes(span trace.Span, attributes map[string]interface{}) {
	for key, value := range attributes {
		switch v := value.(type) {
		case string:
			span.SetAttributes(attribute.String(key, v))
		case int:
			span.SetAttributes(attribute.Int(key, v))
		case int64:
			span.SetAttributes(attribute.Int64(key, v))
		case float64:
			span.SetAttributes(attribute.Float64(key, v))
		case bool:
			span.SetAttributes(attribute.Bool(key, v))
		default:
			span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", v)))
		}
	}
}

// RecordError records an error in the current span
func RecordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// SetSpanStatus sets the status of the current span
func SetSpanStatus(span trace.Span, code codes.Code, description string) {
	span.SetStatus(code, description)
}
