# Стратегия тестирования полнотекстового поиска

## Обзор

Мы используем трехуровневую стратегию тестирования для обеспечения качества и производительности системы поиска:

```
tests/
├── unit/           # SQLite, быстрые тесты логики
├── integration/    # PostgreSQL, тесты взаимодействия  
└── e2e/           # Полная среда, сквозные тесты
```

## Уровни тестирования

### 1. Unit тесты (SQLite)

**Цель**: Тестирование бизнес-логики и алгоритмов поиска

**Технологии**:
- SQLite в памяти
- Mock объекты для внешних зависимостей
- Быстрое выполнение (< 1 секунды)

**Что тестируем**:
- Валидация параметров поиска
- Логика фильтрации
- Пагинация
- Сортировка
- Обработка ошибок
- Подготовка поисковых запросов

**Запуск**:
```bash
make test-unit
make test-unit-coverage
```

**Пример теста**:
```go
func TestSearchLogic_Validation(t *testing.T) {
    // Тестирование валидации без реальной БД
    options := service.SearchOptions{Limit: -1}
    err := validateSearchOptions(options)
    assert.Error(t, err)
}
```

### 2. Integration тесты (PostgreSQL)

**Цель**: Тестирование полнотекстового поиска с реальной PostgreSQL

**Технологии**:
- Testcontainers для PostgreSQL
- Реальные full-text search запросы
- Тестирование производительности

**Что тестируем**:
- PostgreSQL full-text search функциональность
- Индексы и производительность
- Сложные поисковые запросы
- Стемминг и ранжирование
- Операторы поиска (`&`, `|`, `!`)

**Запуск**:
```bash
make test-integration
make test-integration-coverage
```

**Пример теста**:
```go
func TestSearchIntegration_PostgreSQL(t *testing.T) {
    db := setupPostgreSQLContainer(t)
    // Тестирование реального PostgreSQL full-text search
    response, err := searchService.Search(ctx, options)
    // Проверка релевантности и ранжирования
}
```

### 3. E2E тесты (Полная среда)

**Цель**: Тестирование полного пользовательского сценария

**Технологии**:
- PostgreSQL + Redis контейнеры
- HTTP API тестирование
- Полная среда выполнения

**Что тестируем**:
- HTTP API endpoints
- Кэширование Redis
- Конкурентные запросы
- Производительность под нагрузкой
- Интеграция всех компонентов

**Запуск**:
```bash
make test-e2e
```

**Пример теста**:
```go
func TestSearchE2E(t *testing.T) {
    env := setupE2EEnvironment(t) // PostgreSQL + Redis
    // HTTP запросы к API
    // Тестирование кэширования
    // Нагрузочное тестирование
}
```

## Стратегия выполнения

### Локальная разработка

```bash
# Быстрая проверка при разработке
make test-unit

# Полная проверка перед коммитом
make test-unit && make test-integration

# Полное тестирование перед релизом
make test-unit && make test-integration && make test-e2e
```

### CI/CD Pipeline

#### Pull Request
- ✅ Unit тесты (всегда)
- ✅ Integration тесты (если изменения в поиске)
- ❌ E2E тесты (не запускаются)

#### Main Branch
- ✅ Unit тесты
- ✅ Integration тесты  
- ✅ E2E тесты
- ✅ Performance benchmarks

#### Scheduled (еженедельно)
- ✅ Все тесты
- ✅ Performance regression тесты
- ✅ Load testing

## Метрики и покрытие

### Целевые показатели

| Тип тестов | Покрытие | Время выполнения |
|------------|----------|------------------|
| Unit | > 90% | < 30 секунд |
| Integration | > 80% | < 5 минут |
| E2E | > 70% | < 15 минут |

### Мониторинг производительности

```bash
# Бенчмарки производительности
make test-bench

# Профилирование памяти
go test -memprofile=mem.prof ./tests/integration/...

# CPU профилирование
go test -cpuprofile=cpu.prof ./tests/integration/...
```

## Конфигурация тестов

### Переменные окружения

```bash
# Пропуск медленных тестов
export TEST_SHORT=true

# Включение debug логов
export TEST_DEBUG=true

# Настройка testcontainers
export TESTCONTAINERS_RYUK_DISABLED=true
```

### Флаги запуска

```bash
# Короткие тесты (только unit)
go test -short ./...

# Параллельное выполнение
go test -parallel 4 ./...

# Verbose вывод
go test -v ./...

# Конкретный тест
go test -run TestSearchLogic ./tests/unit/...
```

## Отладка тестов

### Проблемы с контейнерами

```bash
# Проверка Docker
docker ps
docker logs <container_id>

# Очистка контейнеров
docker system prune -f
```

### Проблемы с PostgreSQL

```bash
# Подключение к тестовой БД
psql -h localhost -p <port> -U testuser -d testdb

# Проверка индексов
\d+ epics
EXPLAIN ANALYZE SELECT ...
```

### Проблемы с Redis

```bash
# Подключение к Redis
redis-cli -h localhost -p <port>

# Проверка кэша
KEYS search:*
GET search:key
```

## Лучшие практики

### Unit тесты
- Используйте table-driven тесты
- Мокайте внешние зависимости
- Тестируйте граничные случаи
- Быстрое выполнение (< 100ms на тест)

### Integration тесты
- Используйте реальные данные
- Тестируйте производительность
- Проверяйте корректность SQL запросов
- Очищайте данные между тестами

### E2E тесты
- Тестируйте реальные сценарии
- Проверяйте интеграцию компонентов
- Мониторьте производительность
- Используйте минимальный набор критичных тестов

## Добавление новых тестов

### Unit тест
```go
// tests/unit/new_feature_test.go
func TestNewFeature(t *testing.T) {
    // SQLite setup
    // Mock dependencies
    // Test business logic
}
```

### Integration тест
```go
// tests/integration/new_feature_postgresql_test.go
func TestNewFeature_PostgreSQL(t *testing.T) {
    db := setupPostgreSQLContainer(t)
    // Test with real PostgreSQL
}
```

### E2E тест
```go
// tests/e2e/new_feature_e2e_test.go
func TestNewFeature_E2E(t *testing.T) {
    env := setupE2EEnvironment(t)
    // Test full user scenario via HTTP API
}
```

## Мониторинг и алерты

### Метрики для отслеживания
- Время выполнения тестов
- Процент успешных тестов
- Покрытие кода
- Производительность поиска
- Использование ресурсов

### Алерты
- Падение тестов на main ветке
- Деградация производительности > 20%
- Снижение покрытия кода > 5%
- Превышение времени выполнения тестов

Эта стратегия обеспечивает надежное тестирование на всех уровнях при оптимальном использовании ресурсов и времени разработки.