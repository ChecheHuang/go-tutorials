package stockstore

import (
	"context"
	"sync"
	"testing"

	"ticket-system/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore_DecrIfSufficient_Basic(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	require.NoError(t, store.Set(ctx, "stock:1", 10))

	// 扣 3，剩 7
	remaining, err := store.DecrIfSufficient(ctx, "stock:1", 3)
	require.NoError(t, err)
	assert.Equal(t, int64(7), remaining)

	// 再扣 7，剩 0
	remaining, err = store.DecrIfSufficient(ctx, "stock:1", 7)
	require.NoError(t, err)
	assert.Equal(t, int64(0), remaining)

	// 再扣 1，庫存不足
	_, err = store.DecrIfSufficient(ctx, "stock:1", 1)
	assert.ErrorIs(t, err, domain.ErrInsufficientStock)

	// 確認庫存沒被錯誤扣減
	val, err := store.Get(ctx, "stock:1")
	require.NoError(t, err)
	assert.Equal(t, 0, val)
}

func TestMemoryStore_DecrIfSufficient_Concurrent_NeverOversell(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	const totalStock = 50
	const goroutines = 200 // 200 人搶 50 張票，每人搶 1 張
	require.NoError(t, store.Set(ctx, "stock:1", totalStock))

	var (
		wg        sync.WaitGroup
		successMu sync.Mutex
		successes int
	)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.DecrIfSufficient(ctx, "stock:1", 1)
			if err == nil {
				successMu.Lock()
				successes++
				successMu.Unlock()
			}
		}()
	}
	wg.Wait()

	// 成功數量必須剛好等於總庫存
	assert.Equal(t, totalStock, successes, "成功搶到的數量應該剛好等於庫存")

	// 最終庫存必須為 0（不是負數）
	val, err := store.Get(ctx, "stock:1")
	require.NoError(t, err)
	assert.Equal(t, 0, val, "最終庫存必須為 0，不能超賣")
}

func TestMemoryStore_DecrIfSufficient_Concurrent_MultiQuantity(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	const totalStock = 100
	const goroutines = 80 // 80 人各搶 2 張 = 需要 160 張，但只有 100 張
	require.NoError(t, store.Set(ctx, "stock:1", totalStock))

	var (
		wg        sync.WaitGroup
		successMu sync.Mutex
		successes int
	)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.DecrIfSufficient(ctx, "stock:1", 2)
			if err == nil {
				successMu.Lock()
				successes++
				successMu.Unlock()
			}
		}()
	}
	wg.Wait()

	// 最多 50 人成功（100 張 / 每人 2 張）
	assert.Equal(t, totalStock/2, successes, "成功數量應為 50")

	val, err := store.Get(ctx, "stock:1")
	require.NoError(t, err)
	assert.Equal(t, 0, val, "最終庫存必須為 0")
}

func TestMemoryStore_IncrBy(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	require.NoError(t, store.Set(ctx, "stock:1", 5))

	result, err := store.IncrBy(ctx, "stock:1", 3)
	require.NoError(t, err)
	assert.Equal(t, int64(8), result)
}

func TestMemoryStore_GetNotFound(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	_, err := store.Get(ctx, "nonexistent")
	assert.Error(t, err)
}
