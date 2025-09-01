package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"product-requirements-management/internal/config"
)

// RedisClient wraps the Redis client with additional functionality
type RedisClient struct {
	*redis.Client
	logger *logrus.Logger
}

// NewRedisClient creates a new Redis client connection
func NewRedisClient(cfg *config.RedisConfig, logger *logrus.Logger) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		
		// Connection pool settings
		PoolSize:     10,
		MinIdleConns: 5,
		
		// Timeouts
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		
		// Retry settings
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Successfully connected to Redis")

	return &RedisClient{
		Client: rdb,
		logger: logger,
	}, nil
}

// HealthCheck performs a health check on the Redis connection
func (r *RedisClient) HealthCheck(ctx context.Context) error {
	return r.Ping(ctx).Err()
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	r.logger.Info("Closing Redis connection")
	return r.Client.Close()
}

// SetWithExpiration sets a key-value pair with expiration
func (r *RedisClient) SetWithExpiration(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Set(ctx, key, value, expiration).Err()
}

// GetString gets a string value by key
func (r *RedisClient) GetString(ctx context.Context, key string) (string, error) {
	return r.Get(ctx, key).Result()
}

// DeleteKeys deletes multiple keys
func (r *RedisClient) DeleteKeys(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return r.Del(ctx, keys...).Err()
}

// DeletePattern deletes all keys matching a pattern
func (r *RedisClient) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := r.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	
	if len(keys) > 0 {
		return r.Del(ctx, keys...).Err()
	}
	
	return nil
}

// Exists checks if a key exists
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.Client.Exists(ctx, key).Result()
	return result > 0, err
}

// TTL gets the time to live for a key
func (r *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.Client.TTL(ctx, key).Result()
}

// Increment increments a key's value
func (r *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	return r.Incr(ctx, key).Result()
}

// IncrementBy increments a key's value by a specific amount
func (r *RedisClient) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.IncrBy(ctx, key, value).Result()
}

// SetNX sets a key only if it doesn't exist (atomic operation)
func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.Client.SetNX(ctx, key, value, expiration).Result()
}