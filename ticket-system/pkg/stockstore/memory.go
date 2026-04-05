package stockstore

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
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

func (s *MemoryStore) DecrBy(_ context.Context, key string, value int64) (int64, error) {
	return s.getOrCreate(key).Add(-value), nil
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
