# Задачи по миграции тестов на новую стратегию

## 🎯 Цель
Реализовать новую стратегию тестирования:
- **Unit тесты**: SQLite (быстро, изолированно) 
- **Integration тесты**: PostgreSQL с testcontainers
- **E2E тесты**: PostgreSQL с testcontainers

## 📋 Список задач

### Этап 1: Инфраструктура PostgreSQL тестов

- [x] **1.1** Добавить testcontainers зависимости
- [x] **1.2** Создать `internal/integration/test_database.go` с PostgreSQL поддержкой
- [ ] **1.3** Создать утилиты для E2E тестов с PostgreSQL
- [ ] **1.4** Обновить Makefile targets для новой стратегии

### Этап 2: Миграция Integration тестов

- [ ] **2.1** Обновить `search_comprehensive_test.go` для PostgreSQL
  - Заменить `setupTestDB()` на `SetupTestDatabase()`
  - Исправить full-text search тесты (устранить ошибки "@" token)
  - Добавить proper cleanup

- [ ] **2.2** Обновить `config_integration_test.go` для PostgreSQL
  - Перевести на testcontainers
  - Обеспечить изоляцию данных между тестами

- [ ] **2.3** Обновить `epic_integration_test.go` для PostgreSQL
  - Заменить SQLite setup на PostgreSQL
  - Проверить все CRUD операции

- [ ] **2.4** Обновить `inline_comment_integration_test.go` для PostgreSQL
  - Перевести на testcontainers
  - Проверить позиционирование комментариев

- [ ] **2.5** Обновить `requirement_integration_test.go` для PostgreSQL
  - Заменить database setup
  - Проверить связи между требованиями

- [ ] **2.6** Обновить `user_story_integration_test.go` для PostgreSQL
  - Перевести на testcontainers
  - Проверить валидацию user story templates

- [ ] **2.7** Обновить `status_model_integration_test.go` для PostgreSQL
  - Убрать skip и реализовать с PostgreSQL

### Этап 3: Исправление E2E тестов

- [ ] **3.1** Исправить `tests/e2e/search_e2e_test.go`
  - Обновить `database.NewRedisClient` signature (строка 360)
  - Исправить `service.NewEpicService` arguments (строка 375)
  - Исправить `service.NewUserStoryService` arguments (строка 376)
  - Исправить `service.NewRequirementService` arguments (строка 377)
  - Исправить `handlers.NewSearchHandler` arguments (строка 380)
  - Добавить `routes.SetupRoutes` и `routes.Handlers` (строка 390)
  - Удалить поле `FullName` из `models.User` (строка 412)

- [ ] **3.2** Перевести E2E тесты на PostgreSQL testcontainers
  - Создать E2E test database utilities
  - Заменить database setup в E2E тестах

### Этап 4: Исправление Integration тестов (build issues)

- [ ] **4.1** Исправить `tests/integration/search_postgresql_test.go`
  - Удалить неиспользуемый import `database/sql` (строка 5)
  - Исправить поля `SearchFilters`: `CreatedAfter`, `CreatedBefore` (строки 228-229)
  - Добавить метод `GetSearchSuggestions` или удалить использование (строки 251, 268)
  - Добавить константы: `models.EpicStatusCompleted`, `models.UserStoryStatusReady` (строки 363, 402)
  - Удалить поле `FullName` из `models.User` (строка 441)

### Этап 5: Обновление конфигурации и документации

- [x] **5.1** Обновить `.kiro/steering/tech.md` с новой стратегией тестирования
- [x] **5.2** Обновить `TESTING_RESULTS.md` с текущим статусом
- [ ] **5.3** Обновить `README.md` с инструкциями по тестированию
- [ ] **5.4** Создать документацию по запуску тестов для разработчиков

### Этап 6: CI/CD интеграция

- [ ] **6.1** Обновить GitHub Actions (если есть) для новой стратегии тестирования
- [ ] **6.2** Добавить Docker requirements для CI окружения
- [ ] **6.3** Создать скрипты для автоматического тестирования

## 🎯 Приоритеты выполнения

### Высокий приоритет (Неделя 1)
- Задачи 2.1 (search тесты) - устраняет основные падения
- Задачи 3.1 (E2E build issues) - позволяет компилировать E2E тесты
- Задачи 4.1 (integration build issues) - позволяет компилировать integration тесты

### Средний приоритет (Неделя 2)  
- Задачи 2.2-2.7 (остальные integration тесты)
- Задачи 3.2 (E2E PostgreSQL migration)
- Задачи 1.3-1.4 (инфраструктура)

### Низкий приоритет (Неделя 3)
- Задачи 5.3-5.4 (документация)
- Задачи 6.1-6.3 (CI/CD)

## 📊 Ожидаемые результаты

После выполнения всех задач:
- ✅ **Unit тесты**: 100% проходят на SQLite (быстро)
- ✅ **Integration тесты**: 100% проходят на PostgreSQL (реальная среда)
- ✅ **E2E тесты**: 100% проходят на PostgreSQL (полный стек)
- ✅ **Full-text search**: Работает корректно в тестах
- ✅ **CI/CD готовность**: Тесты работают в любой среде с Docker
- ✅ **Изоляция**: Каждый тест получает чистую базу данных

## 🚀 Команды для проверки прогресса

### Базовые команды:
```bash
# Проверка компиляции всех тестов
make test-compile

# Запуск unit тестов (должны работать)
make test-unit

# Запуск integration тестов (после миграции)
make test-integration  

# Запуск E2E тестов (после исправления)
make test-e2e

# Полный цикл тестирования (unit → integration → e2e)
make test
```

### Команды для отладки:
```bash
# Быстрые тесты для разработки
make test-fast

# Отладочный режим с подробным выводом
make test-debug

# Поиск race conditions
make test-race

# Запуск конкретного теста
make test-run TEST=TestSearchIntegration_ComprehensiveSearch
```

### Команды для анализа:
```bash
# Покрытие кода
make test-coverage
make test-unit-coverage
make test-integration-coverage

# Бенчмарки производительности
make test-bench

# Тесты для CI/CD (без E2E)
make test-ci
```

### Полный список команд:
```bash
# Показать все доступные команды
make help
```