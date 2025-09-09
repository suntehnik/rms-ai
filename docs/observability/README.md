# Observability Implementation

This document describes the comprehensive observability implementation for the Product Requirements Management System, including metrics collection, distributed tracing, health checks, and monitoring dashboards.

## Overview

The observability stack includes:

- **Prometheus Metrics**: Custom business and system metrics
- **OpenTelemetry Tracing**: Distributed tracing across all services
- **Health Checks**: Comprehensive health monitoring endpoints
- **Grafana Dashboards**: Visual monitoring and alerting
- **AlertManager**: Intelligent alert routing and notification

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Application   │───▶│ OTEL Collector  │───▶│     Jaeger      │
│                 │    │                 │    │   (Tracing)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │
         │                       ▼
         │              ┌─────────────────┐    ┌─────────────────┐
         │              │   Prometheus    │───▶│    Grafana      │
         │              │   (Metrics)     │    │ (Visualization) │
         │              └─────────────────┘    └─────────────────┘
         │                       │
         ▼                       ▼
┌─────────────────┐    ┌─────────────────┐
│ Health Checks   │    │  AlertManager   │
│   Endpoints     │    │ (Notifications) │
└─────────────────┘    └─────────────────┘
```

## Features

### 1. Prometheus Metrics

#### HTTP Metrics
- `http_requests_total`: Total HTTP requests by method, endpoint, and status code
- `http_request_duration_seconds`: Request duration histogram
- `http_requests_in_flight`: Current number of requests being processed
- `http_response_size_bytes`: Response size histogram

#### Database Metrics
- `database_connections`: Connection pool status (open, idle, in_use)
- `database_queries_total`: Total database queries by operation and table
- `database_query_duration_seconds`: Database query duration histogram

#### Business Metrics
- `entities_created_total`: Total entities created by type and user
- `entities_updated_total`: Total entities updated by type and user
- `entities_deleted_total`: Total entities deleted by type and user
- `comments_total`: Total comments by entity type and status
- `search_queries_total`: Total search queries by type and user
- `search_duration_seconds`: Search query duration histogram

#### System Metrics
- `application_info`: Application metadata (service name, version)
- `application_uptime_seconds_total`: Application uptime

### 2. OpenTelemetry Tracing

#### Automatic Instrumentation
- HTTP requests with method, path, status code, and timing
- Database operations with query type, table, and duration
- Service operations with context propagation

#### Custom Spans
- Business logic operations
- External API calls
- Background job processing

#### Trace Attributes
- Request/correlation IDs for request tracking
- User context for security auditing
- Business context (entity types, operations)

### 3. Health Checks

#### Endpoints
- `GET /health`: Basic liveness check
- `GET /health/live`: Kubernetes liveness probe
- `GET /health/ready`: Kubernetes readiness probe with dependency checks
- `GET /health/deep`: Comprehensive health check with performance metrics

#### Health Check Components
- Application status
- Database connectivity (PostgreSQL and Redis)
- Connection pool health
- Metrics system status
- Memory and performance indicators

### 4. Monitoring Dashboards

#### Grafana Dashboard Panels
- HTTP request rate and response times
- HTTP status code distribution
- Database connection pool status
- Database query performance
- Entity operation rates
- Search performance metrics
- Application uptime and health

### 5. Alerting

#### Alert Categories
- **Critical**: Application down, high error rates, database failures
- **Warning**: High response times, resource usage, slow queries
- **Info**: Business metrics anomalies, unusual activity patterns

#### Alert Routing
- Critical alerts: Immediate notification via email and Slack
- Warning alerts: Team notifications with reduced frequency
- Info alerts: Daily digest for monitoring team

## Configuration

### Environment Variables

```bash
# Observability Configuration
SERVICE_NAME=product-requirements-management
SERVICE_VERSION=1.0.0
ENVIRONMENT=development

# Metrics
METRICS_ENABLED=true

# Tracing
TRACING_ENABLED=true
TRACING_ENDPOINT=http://localhost:4318/v1/traces
```

### Application Configuration

The observability system is configured in `internal/config/config.go`:

```go
type ObservabilityConfig struct {
    ServiceName     string
    ServiceVersion  string
    Environment     string
    MetricsEnabled  bool
    TracingEnabled  bool
    TracingEndpoint string
}
```

## Usage

### Starting the Observability Stack

1. **Start the application with observability enabled**:
   ```bash
   make dev
   ```

2. **Start the monitoring stack**:
   ```bash
   cd docs/observability
   docker-compose -f docker-compose.observability.yml up -d
   ```

### Accessing Monitoring Tools

- **Application Metrics**: http://localhost:8080/metrics
- **Health Checks**: http://localhost:8080/health
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Jaeger**: http://localhost:16686
- **AlertManager**: http://localhost:9093

### Custom Metrics in Code

```go
// Record entity operations
if metrics.AppMetrics != nil {
    metrics.AppMetrics.RecordEntityOperation("create", "epic", userID)
}

// Record database operations
if metrics.AppMetrics != nil {
    metrics.AppMetrics.RecordDatabaseQuery("postgresql", "select", "epics", duration)
}

// Record search operations
if metrics.AppMetrics != nil {
    metrics.AppMetrics.RecordSearch("full_text", userID, duration)
}
```

### Custom Tracing in Code

```go
// Start a service span
ctx, span := tracing.AppTracer.StartServiceSpan(ctx, "epic-service", "create-epic")
defer span.End()

// Add attributes
tracing.AddSpanAttributes(span, map[string]interface{}{
    "epic.id": epic.ID,
    "user.id": userID,
})

// Record errors
if err != nil {
    tracing.RecordError(span, err)
    return err
}
```

## Testing

### Running Observability Tests

```bash
# Run all observability tests
go test ./internal/observability/...

# Run with coverage
go test -cover ./internal/observability/...

# Run benchmarks
go test -bench=. ./internal/observability/...
```

### Test Coverage

The observability implementation includes comprehensive tests for:

- Metrics collection and recording
- Tracing span creation and attributes
- Health check endpoints and responses
- Middleware functionality
- Concurrent access patterns
- Performance benchmarks

## Performance Considerations

### Metrics Collection
- Metrics are collected in-memory with minimal overhead
- Prometheus scraping occurs every 15 seconds by default
- Histogram buckets are optimized for typical response times

### Tracing Overhead
- Tracing can be disabled in production if needed
- Sampling rates can be configured to reduce overhead
- Spans are batched for efficient export

### Health Checks
- Health checks have configurable timeouts
- Deep health checks are more expensive and should be used sparingly
- Basic health checks are optimized for high frequency

## Troubleshooting

### Common Issues

1. **Metrics not appearing in Prometheus**
   - Check that the `/metrics` endpoint is accessible
   - Verify Prometheus configuration and targets
   - Check application logs for metric registration errors

2. **Traces not appearing in Jaeger**
   - Verify OTLP endpoint configuration
   - Check that tracing is enabled in application config
   - Ensure Jaeger is running and accessible

3. **Health checks failing**
   - Check database connectivity
   - Verify Redis connection
   - Review application logs for specific errors

### Debug Commands

```bash
# Check metrics endpoint
curl http://localhost:8080/metrics

# Check health endpoints
curl http://localhost:8080/health
curl http://localhost:8080/health/ready
curl http://localhost:8080/health/deep

# Check Prometheus targets
curl http://localhost:9090/api/v1/targets

# Check Jaeger health
curl http://localhost:16686/api/services
```

## Security Considerations

- Metrics endpoints should be protected in production
- Trace data may contain sensitive information
- Health check endpoints can reveal system architecture
- Consider using authentication for monitoring tools
- Implement proper network segmentation for monitoring stack

## Future Enhancements

- Integration with APM tools (New Relic, Datadog)
- Custom business dashboards
- Automated anomaly detection
- Performance regression testing
- Cost optimization monitoring
- Multi-environment monitoring setup