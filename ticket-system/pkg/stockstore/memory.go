package stockstore

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"ticket-system/internal/domain"
)

// MemoryStore 使用 in-memory 的庫存儲存（不需要 Redis）
type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]*atomic.Int64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string]*atomic.Int64)}
}

func (s *MemoryStore) getOrCreate(key string) *atomic.Int64 {
	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()
	if ok {
		return v
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.data[key]; ok {
		return v
	}
	v = &atomic.Int64{}
	s.data[key] = v
	return v
}

// DecrIfSufficient 原子扣減：使用 CAS（Compare-And-Swap）迴圈確保不會超賣
// 只有當前庫存 >= 扣減數量時才會成功，否則回傳 ErrInsufficientStock
func (s *MemoryStore) DecrIfSufficient(_ context.Context, key string, value int64) (int64, error) {
	counter := s.getOrCreate(key)
	for {
		current := counter.Load()
		if current < value {
			return current, domain.ErrInsufficientStock
		}
		// CAS：只有在值未被其他 goroutine 修改時才寫入
		// 如果 CAS 失敗，代表有其他 goroutine 同時修改了庫存，重新嘗試
		if counter.CompareAndSwap(current, current-value) {
			return current - value, nil
		}
	}
}

func (s *MemoryStore) IncrBy(_ context.Context, key string, value int64) (int64, error) {
	return s.getOrCreate(key).Add(value), nil
}

func (s *MemoryStore) Get(_ context.Context, key string) (int, error) {
	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()
	if !ok {
		return 0, fmt.Errorf("key not found: %s", key)
	}
	return int(v.Load()), nil
}

func (s *MemoryStore) Set(_ context.Context, key string, value int) error {
	s.getOrCreate(key).Store(int64(value))
	return nil
}
