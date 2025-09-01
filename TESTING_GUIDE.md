# Руководство по тестированию

## 🎯 Стратегия тестирования

### Типы тестов и базы данных:
- **Unit тесты**: SQLite (быстро, изолированно)
- **Integration тесты**: PostgreSQL через testcontainers (реальная среда)
- **E2E тесты**: PostgreSQL через testcontainers (полный стек)

## 🚀 Быстрый старт

### Для разработки:
```bash
# Быстрая проверка во время разработки
make test-fast

# Проверка компиляции
make test-compile

# Unit тесты (всегда работают)
make test-unit

# Отладка тестов
make test-debug
```

### Для полного тестирования:
```bash
# Все тесты по порядку (unit → integration → e2e)
make test

# Только integration тесты
make test-integration

# Только E2E тесты  
make test-e2e

# Тесты для CI/CD (без E2E)
make test-ci
```

### Анализ и покрытие:
```bash
# Покрытие кода всех тестов
make test-coverage

# Покрытие только unit тестов
make test-unit-coverage

# Бенчмарки производительности
make test-bench

# Поиск race conditions
make test-race

# Запуск конкретного теста
make test-run TEST=TestName
```

## 📋 Текущий статус

| Тип теста | База данных | Статус | Примечания |
|-----------|-------------|--------|------------|
| Unit | SQLite | ✅ Работают | ~100+ тестов |
| Integration | PostgreSQL | ⚠️ Миграция | Переводим на testcontainers |
| E2E | PostgreSQL | ❌ Не работают | Нужно исправить API signatures |

## 🛠 Требования

### Для unit тестов:
- Только Go (никаких внешних зависимостей)

### Для integration/E2E тестов:
- Docker (для testcontainers)
- Интернет (для скачивания PostgreSQL образа)

## 📚 Документация

- **Детальные задачи**: `TEST_MIGRATION_TASKS.md`
- **Стратегия**: `INTEGRATION_TEST_STRATEGY.md`
- **Makefile команды**: `MAKEFILE_TEST_STRUCTURE.md`
- **Анализ проблем**: `TEST_FAILURE_ANALYSIS.md`

## 🔧 Разработка тестов

### Unit тесты:
```go
// Используют SQLite, быстрые
func TestSomething(t *testing.T) {
    db := setupTestDB(t) // SQLite in-memory
    // тест логики
}
```

### Integration тесты:
```go
// Используют PostgreSQL testcontainers
func TestIntegration(t *testing.T) {
    testDB := SetupTestDatabase(t) // PostgreSQL container
    defer testDB.Cleanup(t)
    // тест с реальной базой
}
```

## 🚨 Известные проблемы

1. **Integration тесты**: Миграция на PostgreSQL в процессе
2. **E2E тесты**: Не компилируются (API signatures)
3. **Search тесты**: Падают из-за PostgreSQL full-text search

## 📞 Помощь

Если тесты не работают:
1. Проверьте Docker: `docker --version`
2. Запустите: `make test-compile`
3. Проверьте unit тесты: `make test-unit`
4. Посмотрите задачи: `TEST_MIGRATION_TASKS.md`