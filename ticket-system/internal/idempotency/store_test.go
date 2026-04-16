package idempotency

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newTestStore(t *testing.T) (*Store, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	store := NewStore(ctx, 50*time.Millisecond) // 50ms 清理間隔（測試用）
	return store, cancel
}

func TestStore_MarkAndCheck(t *testing.T) {
	store, cancel := newTestStore(t)
	defer cancel()

	assert.False(t, store.Check("order:1"), "新 key 不應存在")

	store.Mark("order:1")
	assert.True(t, store.Check("order:1"), "Mark 後應存在")

	// 重複 Mark 不應 panic
	store.Mark("order:1")
	assert.True(t, store.Check("order:1"))
}

func TestStore_MarkWithTTL_ExpiresAfterTTL(t *testing.T) {
	store, cancel := newTestStore(t)
	defer cancel()

	store.MarkWithTTL("order:1", 100*time.Millisecond)
	assert.True(t, store.Check("order:1"), "TTL 內應存在")

	// 等待 TTL + 清理間隔
	time.Sleep(200 * time.Millisecond)
	assert.False(t, store.Check("order:1"), "TTL 過期後不應存在")
}

func TestStore_Mark_NeverExpires(t *testing.T) {
	store, cancel := newTestStore(t)
	defer cancel()

	store.Mark("permanent:1")
	time.Sleep(200 * time.Millisecond) // 多個清理週期後
	assert.True(t, store.Check("permanent:1"), "永久 key 不應過期")
}

func TestStore_MarkWithTTL_CheckBeforeCleanup_StillExpires(t *testing.T) {
	store, cancel := newTestStore(t)
	defer cancel()

	store.MarkWithTTL("order:1", 50*time.Millisecond)
	time.Sleep(80 * time.Millisecond)

	// 即使 cleanup goroutine 還沒跑，Check 本身也會偵測到過期
	assert.False(t, store.Check("order:1"), "Check 應即時偵測過期")
}

func TestStore_Size(t *testing.T) {
	store, cancel := newTestStore(t)
	defer cancel()

	assert.Equal(t, 0, store.Size())

	store.Mark("a")
	store.Mark("b")
	store.MarkWithTTL("c", 1*time.Minute)
	assert.Equal(t, 3, store.Size())
}

func TestStore_ConcurrentMarkAndCheck(t *testing.T) {
	store, cancel := newTestStore(t)
	defer cancel()

	var wg sync.WaitGroup
	const goroutines = 100

	// 併發寫入不同的 key
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "order:" + string(rune('A'+n%26))
			store.MarkWithTTL(key, 5*time.Second)
			store.Check(key)
		}(i)
	}
	wg.Wait()
}

func TestStore_CleanupStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	store := NewStore(ctx, 50*time.Millisecond)

	store.MarkWithTTL("order:1", 10*time.Millisecond)
	cancel() // 停止 cleanup goroutine

	// 即使 cleanup 停了，Check 仍可正確偵測過期
	time.Sleep(50 * time.Millisecond)
	assert.False(t, store.Check("order:1"))
}
