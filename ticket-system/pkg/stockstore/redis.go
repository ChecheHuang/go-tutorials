// Package stockstore 提供庫存儲存的不同實作
package stockstore

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// RedisStore 使用 Redis 的庫存儲存
type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

func (s *RedisStore) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return s.client.DecrBy(ctx, key, value).Result()
}

func (s *RedisStore) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return s.client.IncrBy(ctx, key, value).Result()
}

func (s *RedisStore) Get(ctx context.Context, key string) (int, error) {
	return s.client.Get(ctx, key).Int()
}

func (s *RedisStore) Set(ctx context.Context, key string, value int) error {
	return s.client.Set(ctx, key, value, 0).Err()
}
