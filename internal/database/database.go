package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"product-requirements-management/internal/config"
	"product-requirements-management/internal/models"
)

// DB holds database connections
type DB struct {
	Postgres *gorm.DB
	Redis    *redis.Client
}

// New creates new database connections
func New(cfg *config.Config) (*DB, error) {
	// Initialize PostgreSQL connection
	pg, err := initPostgreSQL(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}

	// Initialize Redis connection
	rdb, err := initRedis(cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	db := &DB{
		Postgres: pg,
		Redis:    rdb,
	}

	// Initialize models (auto-migrate and seed default data)
	if err := db.InitializeModels(); err != nil {
		return nil, fmt.Errorf("failed to initialize models: %w", err)
	}

	return db, nil
}

func NewPostgresDB(cfg *config.Config) (*gorm.DB, error) {
	pg, err := initPostgreSQL(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}
	return pg, err
}

// initPostgreSQL initializes PostgreSQL connection with GORM
func initPostgreSQL(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// initRedis initializes Redis connection
func initRedis(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return rdb, nil
}

// Close closes all database connections
func (db *DB) Close() error {
	var errs []error

	// Close PostgreSQL connection
	if db.Postgres != nil {
		sqlDB, err := db.Postgres.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close PostgreSQL: %w", err))
			}
		}
	}

	// Close Redis connection
	if db.Redis != nil {
		if err := db.Redis.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Redis: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing database connections: %v", errs)
	}

	return nil
}

// InitializeModels runs auto-migration and seeds default data
func (db *DB) InitializeModels() error {
	// Auto-migrate all models
	if err := models.AutoMigrate(db.Postgres); err != nil {
		return fmt.Errorf("failed to auto-migrate models: %w", err)
	}

	// Seed default data (requirement types and relationship types)
	if err := models.SeedDefaultData(db.Postgres); err != nil {
		return fmt.Errorf("failed to seed default data: %w", err)
	}

	return nil
}
