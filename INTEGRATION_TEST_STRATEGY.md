# Стратегия исправления интеграционных тестов с PostgreSQL

## 🎯 Цель
Перевести интеграционные тесты с SQLite на PostgreSQL для полной совместимости с production окружением и устранения ошибок full-text search.

## 📋 План реализации

### Этап 1: Настройка PostgreSQL для тестов

#### 1.1 Создание тестовой конфигурации базы данных
```go
// internal/database/test_database.go
func NewTestDatabase() (*DB, error) {
    // Подключение к тестовой PostgreSQL базе
    // Использование testcontainers или локального PostgreSQL
}
```

#### 1.2 Варианты подключения к PostgreSQL:

**Вариант A: Docker Testcontainers (Рекомендуемый)**
- Автоматический запуск PostgreSQL контейнера для каждого теста
- Изолированная среда
- Не требует предустановленного PostgreSQL

**Вариант B: Локальный PostgreSQL**
- Использование локально установленного PostgreSQL
- Быстрее запуск, но требует настройки
- Отдельная тестовая база данных

**Вариант C: Docker Compose для тестов**
- Предварительный запуск PostgreSQL через docker-compose
- Переиспользование контейнера между тестами

### Этап 2: Обновление тестовой инфраструктуры

#### 2.1 Создание тестовых утилит
```go
// internal/integration/test_database.go
type TestDatabase struct {
    DB *gorm.DB
    Container testcontainers.Container // если используем testcontainers
}

func SetupTestDatabase(t *testing.T) *TestDatabase
func (td *TestDatabase) Cleanup()
func (td *TestDatabase) Reset() // очистка данных между тестами
```

#### 2.2 Обновление существующих тестов
- Замена `setupTestDB()` на `SetupTestDatabase()`
- Добавление proper cleanup в каждый тест
- Обеспечение изоляции данных между тестами

### Этап 3: Конфигурация и переменные окружения

#### 3.1 Тестовые переменные окружения
```bash
# .env.test
TEST_DB_HOST=localhost
TEST_DB_PORT=5432
TEST_DB_NAME=product_requirements_test
TEST_DB_USER=test_user
TEST_DB_PASSWORD=test_password
TEST_DB_SSL_MODE=disable
```

#### 3.2 Конфигурация для разных сред
```go
// internal/config/test_config.go
func NewTestConfig() *Config {
    // Специальная конфигурация для тестов
    // Переопределение database настроек
}
```

### Этап 4: Makefile интеграция

#### 4.1 Обновление test targets
```makefile
# Интеграционные тесты с PostgreSQL (testcontainers)
test-integration:
	@echo "🔧 Running integration tests..."
	go test -v ./internal/integration/...
	@echo "✅ Integration tests completed"

# E2E тесты с PostgreSQL (testcontainers)  
test-e2e:
	@echo "🌐 Running E2E tests..."
	go test -v ./tests/e2e/...
	@echo "✅ E2E tests completed"

# Все тесты по порядку
test: test-unit test-integration test-e2e
	@echo "✅ All tests completed successfully!"

# Дополнительные команды
test-coverage    # Покрытие кода
test-debug       # Отладочный режим
test-race        # Поиск race conditions
test-compile     # Проверка компиляции
```

## 🛠 Детальная реализация

### Шаг 1: Установка testcontainers

```bash
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/postgres
```

### Шаг 2: Создание тестовой базы данных

```go
// internal/integration/test_database.go
package integration

import (
    "context"
    "fmt"
    "testing"
    "time"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    
    "product-requirements-management/internal/models"
)

type TestDatabase struct {
    DB        *gorm.DB
    Container *postgres.PostgresContainer
    DSN       string
}

func SetupTestDatabase(t *testing.T) *TestDatabase {
    ctx := context.Background()
    
    // Создаем PostgreSQL контейнер
    container, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:15"),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("testuser"),
        postgres.WithPassword("testpass"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).WithStartupTimeout(30*time.Second)),
    )
    if err != nil {
        t.Fatalf("Failed to start PostgreSQL container: %v", err)
    }

    // Получаем connection string
    dsn, err := container.ConnectionString(ctx, "sslmode=disable")
    if err != nil {
        t.Fatalf("Failed to get connection string: %v", err)
    }

    // Подключаемся к базе
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        t.Fatalf("Failed to connect to test database: %v", err)
    }

    // Выполняем миграции
    err = models.AutoMigrate(db)
    if err != nil {
        t.Fatalf("Failed to migrate test database: %v", err)
    }

    // Заполняем тестовые данные
    err = models.SeedDefaultData(db)
    if err != nil {
        t.Fatalf("Failed to seed test database: %v", err)
    }

    return &TestDatabase{
        DB:        db,
        Container: container,
        DSN:       dsn,
    }
}

func (td *TestDatabase) Cleanup(t *testing.T) {
    if td.Container != nil {
        ctx := context.Background()
        if err := td.Container.Terminate(ctx); err != nil {
            t.Logf("Failed to terminate container: %v", err)
        }
    }
}

func (td *TestDatabase) Reset() error {
    // Очищаем все таблицы для изоляции тестов
    tables := []string{
        "comments", "requirements", "acceptance_criteria", 
        "user_stories", "epics", "users",
        "requirement_relationships", "requirement_types", "relationship_types",
    }
    
    for _, table := range tables {
        if err := td.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
            return err
        }
    }
    
    // Пересоздаем default data
    return models.SeedDefaultData(td.DB)
}
```

### Шаг 3: Обновление существующих тестов

```go
// internal/integration/search_comprehensive_test.go
func TestSearchIntegration_ComprehensiveSearch(t *testing.T) {
    // Настройка PostgreSQL тестовой базы
    testDB := SetupTestDatabase(t)
    defer testDB.Cleanup(t)

    // Создаем сервисы с PostgreSQL
    repos := repository.NewRepositories(testDB.DB)
    searchService := service.NewSearchService(testDB.DB, nil) // Redis не нужен для тестов

    // Создаем тестовые данные
    user := createTestUser(t, testDB.DB)
    
    // Тест поиска - теперь будет работать с PostgreSQL full-text search
    t.Run("search_by_title", func(t *testing.T) {
        // Создаем epic с заголовком содержащим "authentication"
        epic := &models.Epic{
            CreatorID:   user.ID,
            Title:       "User Authentication System",
            Description: stringPtr("System for user login and registration"),
            Priority:    models.PriorityHigh,
            Status:      models.EpicStatusBacklog,
        }
        err := repos.Epic.Create(epic)
        require.NoError(t, err)

        // Поиск должен найти epic
        results, err := searchService.Search(context.Background(), service.SearchOptions{
            Query: "authentication",
            Limit: 10,
        })
        
        require.NoError(t, err) // Теперь не будет ошибки "@" token
        assert.Len(t, results.Results, 1)
        assert.Equal(t, epic.ID, results.Results[0].ID)
    })
}
```

### Шаг 4: Обновление Makefile

```makefile
# PostgreSQL тесты с testcontainers
test-integration:
	@echo "🔧 Running integration tests with PostgreSQL (testcontainers)..."
	go test -v ./internal/integration/...
	@echo "✅ Integration tests completed"

# Альтернативный вариант с внешним PostgreSQL
test-integration-external:
	@echo "🔧 Running integration tests with external PostgreSQL..."
	@if ! docker ps | grep -q test-postgres; then \
		echo "Starting test PostgreSQL container..."; \
		docker run -d --name test-postgres \
			-e POSTGRES_DB=product_requirements_test \
			-e POSTGRES_USER=test_user \
			-e POSTGRES_PASSWORD=test_password \
			-p 5433:5432 postgres:15; \
		sleep 5; \
	fi
	TEST_DB_HOST=localhost TEST_DB_PORT=5433 TEST_DB_NAME=product_requirements_test \
	TEST_DB_USER=test_user TEST_DB_PASSWORD=test_password TEST_DB_SSL_MODE=disable \
	go test -v ./internal/integration/...
	@echo "✅ Integration tests completed"

# Очистка тестовой базы
test-db-clean:
	docker stop test-postgres || true
	docker rm test-postgres || true
```

## 🚀 Преимущества этого подхода

### ✅ Решаемые проблемы:
1. **PostgreSQL Full-Text Search**: Полная совместимость с production
2. **Реальная среда**: Тесты выполняются в условиях, идентичных production
3. **Изоляция**: Каждый тест получает чистую базу данных
4. **CI/CD готовность**: Testcontainers работает в любой среде с Docker

### ✅ Дополнительные преимущества:
1. **Автоматическая очистка**: Контейнеры удаляются после тестов
2. **Параллельное выполнение**: Каждый тест может иметь свою базу
3. **Версионирование**: Можно тестировать с разными версиями PostgreSQL
4. **Отладка**: Можно подключиться к тестовой базе для отладки

## 📅 План выполнения

### Неделя 1: Инфраструктура
- [ ] Установка testcontainers
- [ ] Создание TestDatabase утилит
- [ ] Обновление конфигурации

### Неделя 2: Миграция тестов
- [ ] Обновление search тестов
- [ ] Обновление integration тестов
- [ ] Тестирование на CI/CD

### Неделя 3: Оптимизация
- [ ] Оптимизация скорости тестов
- [ ] Добавление параллельного выполнения
- [ ] Документация

## 🎯 Ожидаемый результат

После реализации:
- ✅ **Unit тесты**: SQLite (быстро, изолированно) - 100% проходят
- ✅ **Integration тесты**: PostgreSQL (testcontainers) - полная совместимость с production
- ✅ **E2E тесты**: PostgreSQL (testcontainers) - реальная среда тестирования
- ✅ **Full-text search**: Работает корректно во всех тестах
- ✅ **CI/CD готовность**: Автоматическое управление контейнерами
- ✅ **Изоляция тестов**: Каждый тест получает чистую базу данных

## 📋 Конкретные задачи для выполнения

Детальный список задач с приоритетами см. в файле: **`TEST_MIGRATION_TASKS.md`**

### Следующие шаги:
1. Выполнить задачи высокого приоритета (Неделя 1)
2. Проверить результаты: `make test-compile && make test-unit`
3. Продолжить с задачами среднего приоритета (Неделя 2)