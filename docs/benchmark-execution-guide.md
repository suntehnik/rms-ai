# Benchmark Test Execution Guide

## Overview

This guide provides comprehensive instructions for executing benchmark tests in the product requirements management system. The benchmark tests measure API endpoint performance using PostgreSQL databases via testcontainers, providing realistic performance measurements that reflect production conditions.

## Quick Start

### Running All Benchmarks

```bash
# Run all benchmark tests
make test-bench

# Run only benchmark tests in the benchmarks package
go test -bench=. -benchmem ./internal/benchmarks/...

# Run benchmarks with specific duration
go test -bench=. -benchtime=5s ./internal/benchmarks/...
```

### Running Specific Benchmark Categories

```bash
# Data generation benchmarks
go test -bench=BenchmarkDataGeneration -benchmem ./internal/benchmarks

# Database operation benchmarks
go test -bench=BenchmarkDatabaseOperations -benchmem ./internal/benchmarks

# API endpoint benchmarks
go test -bench=. -benchmem ./internal/benchmarks/api

# Epic API benchmarks only
go test -bench=BenchmarkEpic -benchmem ./internal/benchmarks/api

# User Story API benchmarks only
go test -bench=BenchmarkUserStory -benchmem ./internal/benchmarks/api
```

## Benchmark Test Categories

### 1. Infrastructure Benchmarks (`internal/benchmarks`)

#### Data Generation Benchmarks
- **SmallDataSet**: Tests creation of 10 users, 25 epics, 100 user stories
- **MediumDataSet**: Tests creation of 50 users, 100 epics, 500 user stories
- **UserCreation**: Measures user entity creation performance
- **EpicCreation**: Measures epic entity creation performance
- **BulkDataInsertion**: Tests bulk insertion of multiple entities

#### Database Operation Benchmarks
- **DatabaseCleanup**: Measures database cleanup and reset performance
- **DatabaseReset**: Tests complete database reset with migrations

#### Metrics Collection Benchmarks
- **BasicMetricsCollection**: Tests basic performance metrics gathering
- **ConcurrentOperationsMetrics**: Measures metrics collection under concurrent load
- **DatabaseConnectionPoolMetrics**: Tests database connection pool monitoring
- **ResponseTimePercentiles**: Measures response time percentile calculations

### 2. API Endpoint Benchmarks (`internal/benchmarks/api`)

#### Epic API Benchmarks
- **BenchmarkEpicCRUD**: Create, Read, Update, Delete operations
- **BenchmarkEpicListing**: List operations with filtering and pagination
- **BenchmarkEpicStatusChange**: Status transition operations
- **BenchmarkEpicAssignment**: Epic assignment operations
- **BenchmarkEpicWithUserStories**: Epic retrieval with related user stories
- **BenchmarkEpicConcurrentOperations**: Concurrent epic operations

#### User Story API Benchmarks
- **BenchmarkUserStoryCRUD**: Create, Read, Update, Delete operations
- **BenchmarkUserStoryListing**: List operations with filtering and pagination
- **BenchmarkUserStoryStatusTransition**: Status change operations
- **BenchmarkUserStoryAssignment**: User story assignment operations
- **BenchmarkUserStoryRelationshipManagement**: Relationship operations
- **BenchmarkUserStoryConcurrentOperations**: Concurrent user story operations

#### Reliability and Error Handling Benchmarks
- **BenchmarkEpicCRUDWithReliability**: CRUD operations with reliability testing
- **BenchmarkConcurrentOperationsWithReliability**: Concurrent operations with reliability
- **BenchmarkResourceConstrainedOperations**: Operations under resource constraints
- **BenchmarkErrorRecoveryScenarios**: Error recovery and timeout handling

## Performance Metrics

### Standard Go Benchmark Metrics
- **ns/op**: Nanoseconds per operation
- **B/op**: Bytes allocated per operation
- **allocs/op**: Number of allocations per operation

### Custom Performance Metrics
- **avg_response_kb**: Average response size in kilobytes
- **bytes/op**: Custom byte allocation tracking
- **db_idle_conns**: Database idle connections
- **db_inuse_conns**: Database connections in use
- **db_open_conns**: Total open database connections
- **db_wait_count**: Database connection wait count
- **db_wait_ms**: Database connection wait time in milliseconds
- **error_rate_%**: Error rate percentage
- **frees/op**: Memory frees per operation
- **gc_count**: Garbage collection count
- **goroutines**: Number of goroutines
- **heap_mb**: Heap memory usage in megabytes
- **max_concurrent**: Maximum concurrent operations
- **ops/sec**: Operations per second
- **p50_ms**, **p90_ms**, **p95_ms**, **p99_ms**: Response time percentiles
- **success/sec**: Successful operations per second

## Expected Performance Targets

### Response Time Targets
- **CRUD operations**: < 100ms (95th percentile)
- **Simple search**: < 200ms (95th percentile)
- **Complex search**: < 500ms (95th percentile)
- **Bulk operations**: < 1000ms for 100 entities (95th percentile)

### Throughput Targets
- **CRUD operations**: > 100 ops/second
- **Search operations**: > 50 ops/second
- **Concurrent requests**: > 200 requests/second

### Resource Usage Targets
- **Memory allocation**: < 10MB per 1000 operations
- **Database connections**: < 20 concurrent connections
- **CPU usage**: < 80% during benchmark execution

## Troubleshooting Guide

### Common Issues and Solutions

#### 1. Container Startup Failures

**Symptoms:**
```
Error: failed to start container
Error: could not start resource
```

**Solutions:**
- Ensure Docker is running and accessible
- Check available system resources (memory, disk space)
- Verify network connectivity for container image downloads
- Restart Docker daemon if necessary

**Commands:**
```bash
# Check Docker status
docker info

# Restart Docker (macOS)
sudo systemctl restart docker

# Clean up Docker resources
docker system prune -f
```

#### 2. Database Connection Issues

**Symptoms:**
```
Error: failed to connect to database
Error: connection refused
```

**Solutions:**
- Wait for container initialization to complete
- Check PostgreSQL container logs
- Verify database credentials and connection parameters
- Ensure proper container networking

**Commands:**
```bash
# Check container logs
docker logs <container_id>

# Verify container networking
docker network ls
docker inspect <container_id>
```

#### 3. Memory Issues

**Symptoms:**
```
Error: out of memory
Error: cannot allocate memory
```

**Solutions:**
- Increase available system memory
- Reduce benchmark dataset sizes
- Run benchmarks sequentially instead of in parallel
- Clean up resources between benchmark runs

**Commands:**
```bash
# Check system memory
free -h  # Linux
vm_stat  # macOS

# Run with smaller datasets
go test -bench=BenchmarkDataGeneration/SmallDataSet ./internal/benchmarks
```

#### 4. Timeout Issues

**Symptoms:**
```
Error: test timeout
Error: context deadline exceeded
```

**Solutions:**
- Increase test timeout duration
- Reduce benchmark execution time
- Check system performance and resource availability

**Commands:**
```bash
# Increase timeout
go test -bench=. -timeout=60m ./internal/benchmarks/...

# Reduce benchmark time
go test -bench=. -benchtime=100ms ./internal/benchmarks/...
```

#### 5. Compilation Errors

**Symptoms:**
```
Error: undefined function
Error: cannot find package
```

**Solutions:**
- Ensure all dependencies are installed
- Run `go mod tidy` to clean up dependencies
- Check import statements and package paths

**Commands:**
```bash
# Update dependencies
go mod tidy
go mod download

# Check compilation
make test-compile
```

### Performance Troubleshooting

#### Slow Benchmark Performance

**Diagnostic Steps:**
1. Check system resource usage during benchmarks
2. Monitor database connection pool metrics
3. Analyze memory allocation patterns
4. Review garbage collection frequency

**Optimization Strategies:**
- Reduce dataset sizes for development testing
- Use connection pooling optimization
- Implement proper resource cleanup
- Consider running benchmarks on dedicated hardware

#### Inconsistent Results

**Diagnostic Steps:**
1. Run benchmarks multiple times with `-count` flag
2. Check for background processes affecting performance
3. Verify system stability and resource availability
4. Monitor external factors (network, disk I/O)

**Commands:**
```bash
# Run multiple iterations
go test -bench=BenchmarkEpicCRUD -count=5 ./internal/benchmarks/api

# Check result consistency
go test -bench=BenchmarkEpicCRUD -count=3 -benchtime=1s ./internal/benchmarks/api
```

## Best Practices

### Development Environment

1. **Resource Allocation**: Ensure adequate system resources (8GB+ RAM recommended)
2. **Docker Configuration**: Allocate sufficient resources to Docker
3. **Background Processes**: Minimize background applications during benchmarking
4. **Network Stability**: Ensure stable network connection for container operations

### Benchmark Execution

1. **Warm-up Runs**: Run benchmarks multiple times to account for JIT optimization
2. **Consistent Environment**: Use consistent hardware and software configurations
3. **Isolation**: Run benchmarks in isolation from other intensive operations
4. **Documentation**: Record system configuration and benchmark parameters

### Result Analysis

1. **Statistical Significance**: Run multiple iterations to ensure statistical validity
2. **Trend Analysis**: Track performance trends over time
3. **Regression Detection**: Compare results against baseline measurements
4. **Context Documentation**: Document system state and configuration during benchmarks

## Integration with CI/CD

### GitHub Actions Integration

The benchmark tests are integrated with GitHub Actions for continuous performance monitoring:

```yaml
# Example GitHub Actions workflow
- name: Run Benchmark Tests
  run: make test-bench
  timeout-minutes: 30

- name: Upload Benchmark Results
  uses: actions/upload-artifact@v3
  with:
    name: benchmark-results
    path: benchmark-results.txt
```

### Makefile Integration

The benchmark tests are integrated with the existing Makefile:

```bash
# Run all benchmarks
make test-bench

# Run benchmarks with coverage
make test-bench-coverage

# Run API-specific benchmarks
make test-bench-api
```

## Monitoring and Alerting

### Performance Regression Detection

Monitor key performance indicators:
- Response time percentiles (P95, P99)
- Throughput metrics (ops/sec)
- Resource utilization (memory, CPU)
- Error rates

### Alerting Thresholds

Set up alerts for:
- Response time increases > 20%
- Throughput decreases > 15%
- Memory usage increases > 30%
- Error rate increases > 1%

## Conclusion

This benchmark test suite provides comprehensive performance monitoring for the product requirements management system. Regular execution of these benchmarks helps ensure system performance remains within acceptable limits as the codebase evolves.

For additional support or questions about benchmark execution, refer to the project documentation or contact the development team.