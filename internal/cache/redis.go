// Package cache provides Redis client wrapper and caching utilities.
package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"

	"github.com/redis/go-redis/v9"
)

// redisClient wraps the Redis client with connection pool management.
type redisClient struct {
	client *redis.Client
	config config.RedisCacheConfig
	mu     sync.RWMutex
}

// globalRedisClient is the global Redis client instance.
var globalRedisClient *redisClient

// once ensures the Redis client is initialized only once.
var once sync.Once

// InitRedisCache initializes the Redis client with the given configuration.
func InitRedisCache(cfg config.RedisCacheConfig) error {
	if !cfg.Enable {
		return nil
	}

	if cfg.Addr == "" {
		return fmt.Errorf("Redis address is required when caching is enabled")
	}

	var err error
	once.Do(func() {
		globalRedisClient = &redisClient{
			config: cfg,
		}
		err = globalRedisClient.connect()
	})

	if err != nil {
		// Reset once so it can be retried
		once = sync.Once{}
		return err
	}

	return nil
}

// connect establishes the Redis connection.
func (r *redisClient) connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.client != nil {
		return nil
	}

	opts := &redis.Options{
		Addr:     r.config.Addr,
		Password: r.config.Password,
		DB:       r.config.DB,
	}

	r.client = redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}

// GetClient returns the underlying Redis client.
func GetClient() *redis.Client {
	if globalRedisClient == nil {
		return nil
	}
	globalRedisClient.mu.RLock()
	defer globalRedisClient.mu.RUnlock()
	return globalRedisClient.client
}

// IsEnabled returns whether Redis caching is enabled.
func IsEnabled() bool {
	if globalRedisClient == nil {
		return false
	}
	return globalRedisClient.config.Enable
}

// GetConfig returns the Redis configuration.
func GetConfig() config.RedisCacheConfig {
	if globalRedisClient == nil {
		return config.RedisCacheConfig{}
	}
	globalRedisClient.mu.RLock()
	defer globalRedisClient.mu.RUnlock()
	return globalRedisClient.config
}

// Close closes the Redis connection.
func Close() error {
	if globalRedisClient == nil || globalRedisClient.client == nil {
		return nil
	}
	globalRedisClient.mu.Lock()
	defer globalRedisClient.mu.Unlock()
	return globalRedisClient.client.Close()
}

// Ping checks if Redis is reachable.
func Ping(ctx context.Context) error {
	if !IsEnabled() {
		return fmt.Errorf("Redis caching is not enabled")
	}
	client := GetClient()
	if client == nil {
		return fmt.Errorf("Redis client is not initialized")
	}
	return client.Ping(ctx).Err()
}
