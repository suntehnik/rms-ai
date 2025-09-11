# Benchmark Test Troubleshooting Guide

## Quick Diagnostic Commands

### Check Benchmark Test Status
```bash
# Compile all benchmark tests
make test-compile

# Run a quick benchmark validation
go test -bench=BenchmarkDataGeneration/SmallDataSet -benchtime=100ms ./internal/benchmarks

# Check for runtime errors
go test -bench=. -benchtime=100ms -timeout=5m ./internal/benchmarks/...
```

### System Resource Check
```bash
# Check Docker status
docker info
docker ps

# Check system resources (macOS)
vm_stat
df -h

# Check system resources (Linux)
free -h
df -h
```

## Common Error Scenarios

### 1. Container Startup Failures

#### Error Messages:
```
Error: failed to start container
Error: could not start resource: API error (500): driver failed programming external connectivity
Error: port already in use
```

#### Root Causes:
- Docker daemon not running
- Port conflicts with existing services
- Insufficient system resources
- Network configuration issues

#### Solutions:
```bash
# Restart Docker daemon
sudo systemctl restart docker  # Linux
# or restart Docker Desktop on macOS/Windows

# Check for port conflicts
lsof -i :5432  # PostgreSQL default port
netstat -tulpn | grep :5432

# Clean up existing containers
docker container prune -f
docker volume prune -f

# Free up system resources
docker system prune -f
```

### 2. Database Connection Issues

#### Error Messages:
```
Error: failed to connect to database
Error: pq: password authentication failed
Error: connection refused
```

#### Root Causes:
- Container not fully initialized
- Database credentials mismatch
- Network connectivity issues
- Container resource constraints

#### Solutions:
```bash
# Wait for container initialization
sleep 10

# Check container logs
docker logs $(docker ps -q --filter "ancestor=postgres")

# Verify container health
docker exec -it <container_id> pg_isready -U postgres

# Test database connection manually
docker exec -it <container_id> psql -U postgres -d testdb -c "SELECT 1;"
```

### 3. Memory and Resource Issues

#### Error Messages:
```
Error: out of memory
Error: cannot allocate memory
runtime: out of memory: cannot allocate
```

#### Root Causes:
- Insufficient system RAM
- Memory leaks in test code
- Large dataset generation
- Concurrent test execution

#### Solutions:
```bash
# Check memory usage
ps aux | grep go
top -p $(pgrep go)

# Run with smaller datasets
go test -bench=BenchmarkDataGeneration/SmallDataSet ./internal/benchmarks

# Reduce concurrency
go test -bench=. -parallel=1 ./internal/benchmarks/...

# Increase system limits (Linux)
ulimit -v unlimited
ulimit -m unlimited
```

### 4. Test Timeout Issues

#### Error Messages:
```
Error: test timeout
panic: test timed out after 10m0s
Error: context deadline exceeded
```

#### Root Causes:
- Slow system performance
- Large dataset processing
- Network latency
- Resource contention

#### Solutions:
```bash
# Increase test timeout
go test -bench=. -timeout=30m ./internal/benchmarks/...

# Reduce benchmark duration
go test -bench=. -benchtime=100ms ./internal/benchmarks/...

# Run tests sequentially
go test -bench=. -parallel=1 ./internal/benchmarks/...

# Check system performance
iostat 1 5  # Linux
top -l 1    # macOS
```

### 5. Compilation Errors

#### Error Messages:
```
Error: undefined: stringPtr
Error: cannot find package
Error: build constraints exclude all Go files
```

#### Root Causes:
- Missing dependencies
- Import path issues
- Build tag conflicts
- Go version compatibility

#### Solutions:
```bash
# Update dependencies
go mod tidy
go mod download

# Check Go version
go version

# Verify build constraints
go list -f '{{.GoFiles}}' ./internal/benchmarks/...

# Clean module cache
go clean -modcache
```

## Performance Issues

### Slow Benchmark Execution

#### Symptoms:
- Benchmarks taking significantly longer than expected
- High CPU or memory usage
- System becoming unresponsive

#### Diagnostic Steps:
```bash
# Monitor system resources during benchmark
top -p $(pgrep go) &
go test -bench=BenchmarkEpicCRUD ./internal/benchmarks/api

# Check database performance
docker exec -it <container_id> psql -U postgres -d testdb -c "
SELECT query, calls, total_time, mean_time 
FROM pg_stat_statements 
ORDER BY total_time DESC LIMIT 10;"

# Profile benchmark execution
go test -bench=BenchmarkEpicCRUD -cpuprofile=cpu.prof ./internal/benchmarks/api
go tool pprof cpu.prof
```

#### Solutions:
- Reduce dataset sizes for development
- Optimize database queries
- Implement proper connection pooling
- Use SSD storage for better I/O performance

### Inconsistent Results

#### Symptoms:
- Large variations in benchmark results
- Unpredictable performance patterns
- Results not reproducible

#### Diagnostic Steps:
```bash
# Run multiple iterations
go test -bench=BenchmarkEpicCRUD -count=5 ./internal/benchmarks/api

# Check for background processes
ps aux | grep -E "(docker|postgres|go)"

# Monitor system stability
vmstat 1 10  # Linux
vm_stat      # macOS
```

#### Solutions:
- Close unnecessary applications
- Run benchmarks on dedicated hardware
- Use consistent test data
- Implement proper warm-up procedures

## Environment-Specific Issues

### macOS Issues

#### Common Problems:
- Docker Desktop resource limits
- File system performance (APFS)
- Network configuration

#### Solutions:
```bash
# Increase Docker Desktop resources
# Docker Desktop > Preferences > Resources > Advanced
# Set Memory: 8GB+, CPUs: 4+

# Use native file system for better performance
# Avoid mounted volumes when possible

# Check network configuration
ping host.docker.internal
```

### Linux Issues

#### Common Problems:
- Docker daemon permissions
- System resource limits
- Container networking

#### Solutions:
```bash
# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Increase system limits
echo "* soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "* hard nofile 65536" | sudo tee -a /etc/security/limits.conf

# Configure container networking
sudo systemctl restart docker
```

### Windows Issues

#### Common Problems:
- WSL2 integration
- File path length limits
- Performance overhead

#### Solutions:
```bash
# Use WSL2 backend for Docker Desktop
# Docker Desktop > Settings > General > Use WSL2 based engine

# Enable long path support
# Windows Settings > Update & Security > For developers > Developer Mode

# Run benchmarks in WSL2 environment
wsl
cd /mnt/c/path/to/project
```

## Debugging Techniques

### Enable Debug Logging

```bash
# Set debug environment variables
export BENCHMARK_DEBUG=true
export TESTCONTAINERS_RYUK_DISABLED=true

# Run with verbose output
go test -bench=. -v ./internal/benchmarks/...
```

### Container Debugging

```bash
# Keep containers running after tests
export TESTCONTAINERS_RYUK_DISABLED=true

# Access container shell
docker exec -it <container_id> /bin/bash

# Check container logs
docker logs -f <container_id>

# Monitor container resources
docker stats <container_id>
```

### Database Debugging

```bash
# Enable PostgreSQL query logging
docker exec -it <container_id> psql -U postgres -c "
ALTER SYSTEM SET log_statement = 'all';
SELECT pg_reload_conf();"

# Monitor database activity
docker exec -it <container_id> psql -U postgres -c "
SELECT pid, usename, application_name, state, query 
FROM pg_stat_activity 
WHERE state = 'active';"
```

## Recovery Procedures

### Clean Environment Reset

```bash
# Stop all containers
docker stop $(docker ps -q)

# Remove all containers and volumes
docker system prune -a -f --volumes

# Clean Go module cache
go clean -modcache

# Rebuild and test
go mod download
make test-compile
```

### Partial Recovery

```bash
# Clean only benchmark-related containers
docker ps --filter "label=testcontainers" -q | xargs docker stop
docker ps --filter "label=testcontainers" -q | xargs docker rm

# Reset test database
go test -bench=BenchmarkDatabaseOperations/DatabaseReset ./internal/benchmarks
```

## Monitoring and Alerting

### Performance Monitoring

```bash
# Continuous monitoring script
#!/bin/bash
while true; do
    echo "$(date): Running benchmark validation..."
    go test -bench=BenchmarkDataGeneration/SmallDataSet -benchtime=100ms ./internal/benchmarks
    if [ $? -ne 0 ]; then
        echo "ALERT: Benchmark test failed!"
        # Send notification
    fi
    sleep 300  # Check every 5 minutes
done
```

### Resource Monitoring

```bash
# Monitor Docker resources
docker stats --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}"

# Monitor system resources
iostat -x 1
sar -u 1
```

## Prevention Strategies

### Development Best Practices

1. **Regular Testing**: Run benchmark tests regularly during development
2. **Resource Management**: Properly clean up resources in test code
3. **Error Handling**: Implement comprehensive error handling
4. **Documentation**: Document known issues and solutions

### Infrastructure Best Practices

1. **Resource Allocation**: Ensure adequate system resources
2. **Monitoring**: Implement continuous monitoring
3. **Backup Plans**: Have fallback procedures for critical issues
4. **Version Control**: Track benchmark performance over time

### Code Quality Best Practices

1. **Code Reviews**: Review benchmark test code for potential issues
2. **Static Analysis**: Use tools to detect potential problems
3. **Performance Profiling**: Regular performance profiling
4. **Dependency Management**: Keep dependencies up to date

## Getting Help

### Internal Resources
- Check project documentation in `docs/` directory
- Review existing benchmark test implementations
- Consult team members familiar with the codebase

### External Resources
- Go testing documentation: https://golang.org/pkg/testing/
- Testcontainers documentation: https://golang.testcontainers.org/
- Docker troubleshooting: https://docs.docker.com/config/troubleshooting/

### Escalation Procedures
1. Document the issue with error messages and steps to reproduce
2. Check if the issue is environment-specific
3. Consult with senior developers or DevOps team
4. Create GitHub issue with detailed information if needed