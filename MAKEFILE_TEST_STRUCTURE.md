# Makefile Test Structure Documentation

## Overview

The Makefile has been updated with a comprehensive test structure that separates different types of tests and provides various testing options for different scenarios.

## Test Hierarchy

### Main Test Target
```bash
make test
```
**Behavior**: Runs tests in sequence: unit → integration → e2e
**Output**: Provides clear progress indicators and stops on first failure
**Example**: `✅ All tests completed successfully!`

### Individual Test Types

#### Unit Tests
```bash
make test-unit
```
- **Scope**: `./internal/models/...`, `./internal/repository/...`, `./internal/service/...`, `./tests/unit/...`
- **Database**: SQLite (in-memory)
- **Speed**: Fast (~1-2 seconds)
- **Status**: ✅ All passing

#### Integration Tests  
```bash
make test-integration
```
- **Scope**: `./internal/integration/...`
- **Database**: PostgreSQL (testcontainers) - NEW STRATEGY
- **Speed**: Medium (~5-10 seconds with container startup)
- **Status**: ⚠️ Migrating to PostgreSQL testcontainers

#### E2E Tests
```bash
make test-e2e
```
- **Scope**: `./tests/e2e/...`
- **Database**: PostgreSQL (testcontainers) - NEW STRATEGY
- **Environment**: Full application stack with real database
- **Speed**: Slow (~10-15 seconds with container startup)
- **Status**: ❌ Build failures + needs PostgreSQL migration

## Specialized Test Targets

### Fast Tests
```bash
make test-fast
```
- Runs unit tests with `-short` flag
- Skips slow operations
- Perfect for development workflow

### CI/CD Tests
```bash
make test-ci
```
- Runs unit + integration tests
- Skips E2E tests (suitable for automated environments)
- Current status: Passes unit tests, fails on integration PostgreSQL issues

### Compilation Check
```bash
make test-compile
```
- Checks if all tests compile without running them
- Useful for catching syntax errors quickly
- Shows exact compilation errors

## Coverage Targets

### Combined Coverage
```bash
make test-coverage
```
Generates coverage for unit and integration tests

### Individual Coverage
```bash
make test-unit-coverage      # Unit test coverage
make test-integration-coverage # Integration test coverage  
make test-e2e-coverage       # E2E test coverage
```

## Advanced Testing

### Race Detection
```bash
make test-race
```
Runs tests with Go's race detector

### Benchmarks
```bash
make test-bench
```
Runs performance benchmarks

### Parallel Execution
```bash
make test-parallel
```
Runs tests in parallel for faster execution

### Specific Test Execution
```bash
make test-run TEST=TestName
```
Runs a specific test by name

### Debug Mode
```bash
make test-debug
```
Runs tests with verbose output and no caching

## Current Test Status Summary

| Test Type | Database | Status | Count | Issues |
|-----------|----------|--------|-------|--------|
| Unit Tests | SQLite | ✅ Pass | ~100+ | None |
| Integration Tests | PostgreSQL* | ⚠️ Migration | ~20 | Migrating to testcontainers |
| E2E Tests | PostgreSQL* | ❌ Build Issues | ~5 | API signatures + migration needed |

*\* Новая стратегия: PostgreSQL через testcontainers*

## Key Features

### Progress Indicators
All test targets include emoji indicators and clear messaging:
- 🧪 Unit tests
- 🔧 Integration tests  
- 🌐 E2E tests
- ⚡ Fast tests
- 🤖 CI tests

### Error Handling
- Tests stop on first failure in sequence
- Clear error messages with file locations
- Compilation errors shown before test execution

### Flexibility
- Run all tests or specific types
- Coverage reports for different scopes
- Debug and development-friendly options

## Usage Examples

### Development Workflow
```bash
# Quick check during development
make test-fast

# Check compilation first
make test-compile

# Run unit tests (always work)
make test-unit

# Debug failing tests
make test-debug

# Full local testing (when integration tests work)
make test-unit test-integration

# Before committing
make test-compile && make test-unit
```

### CI/CD Pipeline
```bash
# Automated testing (unit + integration, no E2E)
make test-ci

# With coverage reports
make test-coverage

# Individual coverage reports
make test-unit-coverage
make test-integration-coverage
make test-e2e-coverage

# Performance analysis
make test-bench
```

### Debugging
```bash
# Check compilation first
make test-compile

# Run specific failing test
make test-run TEST=TestSearchIntegration_ComprehensiveSearch

# Debug mode with verbose output
make test-debug

# Find race conditions
make test-race

# Run tests in parallel
make test-parallel

# Show all available commands
make help
```

## Help System

```bash
make help
```
Shows comprehensive help with all available targets organized by category.

## Next Steps - New Testing Strategy

### ✅ **Completed**:
1. **Testing Strategy Defined**: Unit (SQLite) + Integration/E2E (PostgreSQL)
2. **Infrastructure Created**: testcontainers setup in `internal/integration/test_database.go`
3. **Makefile Updated**: Separate targets for different test types

### 🚀 **In Progress**:
1. **Migration Tasks**: See detailed plan in `TEST_MIGRATION_TASKS.md`
2. **Priority 1**: Fix search integration tests (PostgreSQL full-text search)
3. **Priority 2**: Fix E2E build issues (API signatures)
4. **Priority 3**: Complete migration of all integration tests

### 📋 **Task Tracking**:
- **High Priority**: Week 1 - Core functionality fixes
- **Medium Priority**: Week 2 - Complete migration  
- **Low Priority**: Week 3 - Documentation and CI/CD
#
# 📋 Краткая справка по командам

### Ежедневная разработка:
```bash
make test-fast        # Быстрая проверка
make test-unit        # Unit тесты  
make test-compile     # Проверка компиляции
make test-debug       # Отладка
```

### Полное тестирование:
```bash
make test             # Все тесты по порядку
make test-ci          # Тесты для CI/CD
make test-coverage    # С покрытием кода
```

### Анализ и оптимизация:
```bash
make test-race        # Race conditions
make test-bench       # Производительность
make test-parallel    # Параллельное выполнение
```

### Помощь:
```bash
make help             # Полный список команд
```

> **Совет**: Начинайте всегда с `make test-compile` для проверки компиляции, затем `make test-unit` для быстрой проверки логики.