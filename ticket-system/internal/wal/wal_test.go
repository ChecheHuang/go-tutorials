package wal

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWAL_WriteAndCommit(t *testing.T) {
	w := New()

	entry := w.Write("deduct_stock", `{"event_id":1,"quantity":2}`)
	assert.Equal(t, uint64(1), entry.ID)
	assert.Equal(t, StatusPending, entry.Status)
	assert.Equal(t, "deduct_stock", entry.Operation)

	// Commit 後不應出現在 Recover 中
	require.NoError(t, w.Commit(entry.ID))
	pending := w.Recover()
	assert.Empty(t, pending)
}

func TestWAL_WriteAndFail(t *testing.T) {
	w := New()

	entry := w.Write("deduct_stock", `{"event_id":1}`)
	require.NoError(t, w.Fail(entry.ID))

	pending := w.Recover()
	assert.Empty(t, pending, "failed 的記錄不應出現在 Recover 中")
}

func TestWAL_Recover_OnlyReturnsPending(t *testing.T) {
	w := New()

	e1 := w.Write("op1", "data1") // 會 commit
	e2 := w.Write("op2", "data2") // 維持 pending
	e3 := w.Write("op3", "data3") // 會 fail

	w.Commit(e1.ID)
	w.Fail(e3.ID)

	pending := w.Recover()
	require.Len(t, pending, 1)
	assert.Equal(t, e2.ID, pending[0].ID)
	assert.Equal(t, "op2", pending[0].Operation)
	assert.Equal(t, "data2", pending[0].Payload)
}

func TestWAL_CommitNonexistent_ReturnsError(t *testing.T) {
	w := New()
	assert.Error(t, w.Commit(999))
}

func TestWAL_FailNonexistent_ReturnsError(t *testing.T) {
	w := New()
	assert.Error(t, w.Fail(999))
}

func TestWAL_IDsAreIncremental(t *testing.T) {
	w := New()

	e1 := w.Write("op1", "")
	e2 := w.Write("op2", "")
	e3 := w.Write("op3", "")

	assert.Equal(t, uint64(1), e1.ID)
	assert.Equal(t, uint64(2), e2.ID)
	assert.Equal(t, uint64(3), e3.ID)
}

func TestWAL_Len(t *testing.T) {
	w := New()
	assert.Equal(t, 0, w.Len())

	w.Write("op1", "")
	w.Write("op2", "")
	assert.Equal(t, 2, w.Len())
}

func TestWAL_ConcurrentWriteAndCommit(t *testing.T) {
	w := New()

	var wg sync.WaitGroup
	const goroutines = 100

	entries := make([]Entry, goroutines)
	var mu sync.Mutex

	// 併發寫入
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			entry := w.Write("op", "data")
			mu.Lock()
			entries[n] = entry
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	assert.Equal(t, goroutines, w.Len())

	// 併發 Commit 一半、Fail 一半
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			if n%2 == 0 {
				w.Commit(entries[n].ID)
			} else {
				w.Fail(entries[n].ID)
			}
		}(i)
	}
	wg.Wait()

	// 沒有 pending 了（全部都被 commit 或 fail）
	pending := w.Recover()
	assert.Empty(t, pending)
}

func TestWAL_RecoverReturnsCopy(t *testing.T) {
	w := New()
	w.Write("op1", "data1")

	pending1 := w.Recover()
	require.Len(t, pending1, 1)

	// 修改回傳值不影響 WAL 內部
	pending1[0].Status = StatusCommitted

	pending2 := w.Recover()
	require.Len(t, pending2, 1)
	assert.Equal(t, StatusPending, pending2[0].Status, "修改回傳值不應影響 WAL 內部")
}
