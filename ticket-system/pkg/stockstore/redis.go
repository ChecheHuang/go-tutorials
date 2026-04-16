// Package stockstore 提供庫存儲存的不同實作
package stockstore

import (
	"context"

	"ticket-system/internal/domain"

	"github.com/redis/go-redis/v9"
)

// decrIfSufficientScript 原子扣減庫存的 Lua script
// 只有當庫存 >= 扣減數量時才執行 DECRBY，否則回傳 -1 表示不足
// 整個 script 在 Redis 中是單執行緒原子執行的，不存在 race condition
var decrIfSufficientScript = redis.NewScript(`
local current = tonumber(redis.call('GET', KEYS[1]) or '0')
local deduct = tonumber(ARGV[1])
if current >= deduct then
    return redis.call('DECRBY', KEYS[1], deduct)
else
    return -1
end
`)

// RedisStore 使用 Redis 的庫存儲存
type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

// DecrIfSufficient 原子扣減：透過 Lua script 確保「檢查 + 扣減」在同一個 Redis 命令中完成
func (s *RedisStore) DecrIfSufficient(ctx context.Context, key string, value int64) (int64, error) {
	result, err := decrIfSufficientScript.Run(ctx, s.client, []string{key}, value).Int64()
	if err != nil {
		return 0, err
	}
	if result < 0 {
		return 0, domain.ErrInsufficientStock
	}
	return result, nil
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
