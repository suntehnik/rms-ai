package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	postgresContainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"product-requirements-management/internal/models"
)

// TestDatabase представляет тестовую PostgreSQL базу данных
type TestDatabase struct {
	DB        *gorm.DB
	Container *postgresContainer.PostgresContainer
	DSN       string
}

// SetupTestDatabase создает новый PostgreSQL контейнер для тестов
func SetupTestDatabase(t *testing.T) *TestDatabase {
	ctx := context.Background()

	// Создаем PostgreSQL контейнер
	container, err := postgresContainer.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgresContainer.WithDatabase("testdb"),
		postgresContainer.WithUsername("testuser"),
		postgresContainer.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	// Получаем connection string
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Подключаемся к базе данных
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Отключаем логи для тестов
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Выполняем миграции
	if err := models.AutoMigrate(db); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Заполняем базовые данные
	if err := models.SeedDefaultData(db); err != nil {
		t.Fatalf("Failed to seed test database: %v", err)
	}

	return &TestDatabase{
		DB:        db,
		Container: container,
		DSN:       dsn,
	}
}

// Cleanup останавливает и удаляет PostgreSQL контейнер
func (td *TestDatabase) Cleanup(t *testing.T) {
	if td.Container != nil {
		ctx := context.Background()
		if err := td.Container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate PostgreSQL container: %v", err)
		}
	}
}

// Reset очищает все данные в базе для изоляции тестов
func (td *TestDatabase) Reset() error {
	// Список таблиц в порядке зависимостей (сначала зависимые, потом основные)
	tables := []string{
		"comments",
		"requirement_relationships", 
		"requirements",
		"acceptance_criteria",
		"user_stories",
		"epics",
		"users",
		// Справочники не очищаем, они нужны для тестов
	}

	// Отключаем проверку внешних ключей
	if err := td.DB.Exec("SET session_replication_role = replica").Error; err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}

	// Очищаем таблицы
	for _, table := range tables {
		if err := td.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)).Error; err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	// Включаем обратно проверку внешних ключей
	if err := td.DB.Exec("SET session_replication_role = DEFAULT").Error; err != nil {
		return fmt.Errorf("failed to enable foreign key checks: %w", err)
	}

	return nil
}

// CreateTestUser создает тестового пользователя
func (td *TestDatabase) CreateTestUser(t *testing.T) *models.User {
	user := &models.User{
		Username:     fmt.Sprintf("testuser_%d", time.Now().UnixNano()),
		Email:        fmt.Sprintf("test_%d@example.com", time.Now().UnixNano()),
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}

	if err := td.DB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

// GetRequirementType получает тип требования по имени
func (td *TestDatabase) GetRequirementType(name string) (*models.RequirementType, error) {
	var reqType models.RequirementType
	err := td.DB.Where("name = ?", name).First(&reqType).Error
	return &reqType, err
}

// GetRelationshipType получает тип связи по имени  
func (td *TestDatabase) GetRelationshipType(name string) (*models.RelationshipType, error) {
	var relType models.RelationshipType
	err := td.DB.Where("name = ?", name).First(&relType).Error
	return &relType, err
}