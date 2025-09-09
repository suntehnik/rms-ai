package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
)

func TestInit_Disabled(t *testing.T) {
	ctx := context.Background()
	config := TracingConfig{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Endpoint:       "http://localhost:4318/v1/traces",
		Enabled:        false,
	}

	tracer, err := Init(ctx, config)
	require.NoError(t, err)
	assert.NotNil(t, tracer)
	assert.NotNil(t, tracer.tracer)
	assert.Nil(t, tracer.provider) // Should be nil when disabled
}

func TestInit_Enabled(t *testing.T) {
	ctx := context.Background()
	config := TracingConfig{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Endpoint:       "http://localhost:4318/v1/traces",
		Enabled:        true,
	}

	tracer, err := Init(ctx, config)
	
	// Note: This test might fail if no OTLP receiver is running
	// In a real environment, you'd mock the exporter or use a test receiver
	if err != nil {
		t.Skipf("Skipping test due to OTLP endpoint not available: %v", err)
	}
	
	require.NoError(t, err)
	assert.NotNil(t, tracer)
	assert.NotNil(t, tracer.tracer)
	assert.NotNil(t, tracer.provider)
	
	// Clean up
	if tracer.provider != nil {
		_ = tracer.Shutdown(ctx)
	}
}

func TestStartSpan(t *testing.T) {
	// Use disabled tracing for this test
	ctx := context.Background()
	config := TracingConfig{
		ServiceName: "test-service",
		Enabled:     false,
	}

	tracer, err := Init(ctx, config)
	require.NoError(t, err)

	// Start a span
	spanCtx, span := tracer.StartSpan(ctx, "test-span")
	assert.NotNil(t, spanCtx)
	assert.NotNil(t, span)
	
	// End the span
	span.End()
}

func TestStartHTTPSpan(t *testing.T) {
	ctx := context.Background()
	config := TracingConfig{
		ServiceName: "test-service",
		Enabled:     false,
	}

	tracer, err := Init(ctx, config)
	require.NoError(t, err)

	// Start HTTP span
	spanCtx, span := tracer.StartHTTPSpan(ctx, "GET", "/api/test")
	assert.NotNil(t, spanCtx)
	assert.NotNil(t, span)
	
	// End the span
	span.End()
}

func TestStartDatabaseSpan(t *testing.T) {
	ctx := context.Background()
	config := TracingConfig{
		ServiceName: "test-service",
		Enabled:     false,
	}

	tracer, err := Init(ctx, config)
	require.NoError(t, err)

	// Start database span
	spanCtx, span := tracer.StartDatabaseSpan(ctx, "select", "users")
	assert.NotNil(t, spanCtx)
	assert.NotNil(t, span)
	
	// End the span
	span.End()
}

func TestStartServiceSpan(t *testing.T) {
	ctx := context.Background()
	config := TracingConfig{
		ServiceName: "test-service",
		Enabled:     false,
	}

	tracer, err := Init(ctx, config)
	require.NoError(t, err)

	// Start service span
	spanCtx, span := tracer.StartServiceSpan(ctx, "user-service", "create-user")
	assert.NotNil(t, spanCtx)
	assert.NotNil(t, span)
	
	// End the span
	span.End()
}

func TestAddSpanAttributes(t *testing.T) {
	ctx := context.Background()
	config := TracingConfig{
		ServiceName: "test-service",
		Enabled:     false,
	}

	tracer, err := Init(ctx, config)
	require.NoError(t, err)

	// Start a span
	_, span := tracer.StartSpan(ctx, "test-span")
	
	// Add various attribute types
	attributes := map[string]interface{}{
		"string_attr":  "test-value",
		"int_attr":     42,
		"int64_attr":   int64(123),
		"float64_attr": 3.14,
		"bool_attr":    true,
		"other_attr":   []string{"test"}, // Should be converted to string
	}
	
	// Should not panic
	assert.NotPanics(t, func() {
		AddSpanAttributes(span, attributes)
	})
	
	span.End()
}

func TestRecordError(t *testing.T) {
	ctx := context.Background()
	config := TracingConfig{
		ServiceName: "test-service",
		Enabled:     false,
	}

	tracer, err := Init(ctx, config)
	require.NoError(t, err)

	// Start a span
	_, span := tracer.StartSpan(ctx, "test-span")
	
	// Test with nil error (should not panic)
	assert.NotPanics(t, func() {
		RecordError(span, nil)
	})
	
	// Test with actual error
	testErr := assert.AnError
	assert.NotPanics(t, func() {
		RecordError(span, testErr)
	})
	
	span.End()
}

func TestSetSpanStatus(t *testing.T) {
	ctx := context.Background()
	config := TracingConfig{
		ServiceName: "test-service",
		Enabled:     false,
	}

	tracer, err := Init(ctx, config)
	require.NoError(t, err)

	// Start a span
	_, span := tracer.StartSpan(ctx, "test-span")
	
	// Set different status codes
	assert.NotPanics(t, func() {
		SetSpanStatus(span, codes.Ok, "Success")
	})
	
	assert.NotPanics(t, func() {
		SetSpanStatus(span, codes.Error, "Error occurred")
	})
	
	span.End()
}

func TestShutdown(t *testing.T) {
	ctx := context.Background()
	
	// Test shutdown with nil provider (disabled tracing)
	tracer := &Tracer{provider: nil}
	err := tracer.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestGlobalTracer(t *testing.T) {
	// Reset global tracer
	AppTracer = nil
	
	ctx := context.Background()
	config := TracingConfig{
		ServiceName: "test-service",
		Enabled:     false,
	}

	tracer, err := Init(ctx, config)
	require.NoError(t, err)
	require.NotNil(t, tracer)
	
	// Verify global reference was set
	assert.NotNil(t, AppTracer)
	assert.Equal(t, tracer, AppTracer)
	
	// Reset for other tests
	AppTracer = nil
}

func TestConcurrentSpanCreation(t *testing.T) {
	ctx := context.Background()
	config := TracingConfig{
		ServiceName: "test-service",
		Enabled:     false,
	}

	tracer, err := Init(ctx, config)
	require.NoError(t, err)

	// Test concurrent span creation
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			// Create various types of spans concurrently
			_, span1 := tracer.StartSpan(ctx, "concurrent-span")
			_, span2 := tracer.StartHTTPSpan(ctx, "GET", "/test")
			_, span3 := tracer.StartDatabaseSpan(ctx, "select", "users")
			_, span4 := tracer.StartServiceSpan(ctx, "service", "operation")
			
			// End all spans
			span1.End()
			span2.End()
			span3.End()
			span4.End()
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Test should complete without panics
}

func BenchmarkStartSpan(b *testing.B) {
	ctx := context.Background()
	config := TracingConfig{
		ServiceName: "test-service",
		Enabled:     false,
	}

	tracer, _ := Init(ctx, config)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, span := tracer.StartSpan(ctx, "benchmark-span")
		span.End()
	}
}

func BenchmarkAddSpanAttributes(b *testing.B) {
	ctx := context.Background()
	config := TracingConfig{
		ServiceName: "test-service",
		Enabled:     false,
	}

	tracer, _ := Init(ctx, config)
	_, span := tracer.StartSpan(ctx, "benchmark-span")
	defer span.End()
	
	attributes := map[string]interface{}{
		"string_attr": "test-value",
		"int_attr":    42,
		"bool_attr":   true,
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		AddSpanAttributes(span, attributes)
	}
}