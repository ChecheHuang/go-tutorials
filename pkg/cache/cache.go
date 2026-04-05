// Package cache 提供快取抽象層
// 支援 Redis 快取與 NoOp 快取（當 Redis 未啟用時自動降級）
package cache

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache 定義快取介面
type Cache interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// RedisCache 使用 Redis 實作快取
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache 建立 Redis 快取實例
func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	slog.Info("Redis 連線成功", "addr", addr)
	return &RedisCache{client: client}, nil
}

// Client 回傳底層的 redis.Client（給 Health Check 用）
func (c *RedisCache) Client() *redis.Client {
	return c.client
}

func (c *RedisCache) Get(ctx context.Context, key string, dest any) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func (c *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// NoOpCache 不做任何事的快取（Redis 未啟用時使用）
type NoOpCache struct{}

// NewNoOpCache 建立空操作快取
func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

func (c *NoOpCache) Get(_ context.Context, _ string, _ any) error {
	return redis.Nil // 永遠回傳 cache miss
}

func (c *NoOpCache) Set(_ context.Context, _ string, _ any, _ time.Duration) error {
	return nil
}

func (c *NoOpCache) Delete(_ context.Context, _ string) error {
	return nil
}
