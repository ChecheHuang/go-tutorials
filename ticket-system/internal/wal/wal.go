// Package wal 實作 Write-Ahead Log，確保 Redis crash 後能恢復排隊資料（第 40 課）
//
// WAL 的核心概念：
//   - 在執行任何副作用操作（寫入 Redis、扣庫存等）之前，先將操作寫入日誌
//   - 操作成功後，將日誌標記為已提交（committed）
//   - 如果系統崩潰，重啟時可以透過 Recover() 找出所有 pending 的操作並重新執行
//
// 這裡用記憶體模擬持久化儲存（實際場景會用檔案或 DB）
// 搭配搶票系統的使用情境：
//   - 使用者搶票 → 先寫 WAL（pending）→ 扣 Redis 庫存 → 標記 WAL（committed）
//   - Redis 崩潰重啟 → 讀取所有 pending 的 WAL 記錄 → 重新執行扣庫存
package wal

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// EntryStatus 代表 WAL 記錄的狀態
type EntryStatus string

const (
	StatusPending   EntryStatus = "pending"   // 已寫入但尚未完成
	StatusCommitted EntryStatus = "committed" // 操作已成功完成
	StatusFailed    EntryStatus = "failed"    // 操作失敗
)

// Entry 代表一筆 WAL 記錄
//
// 每次對外部系統的寫入操作，都會先產生一筆 Entry：
//   - Operation: 操作類型（例如 "deduct_stock"、"create_order"）
//   - Payload:   操作的資料（JSON 序列化的內容）
//   - Status:    記錄目前的狀態（pending → committed 或 failed）
type Entry struct {
	ID        uint64      `json:"id"`
	Operation string      `json:"operation"`
	Payload   string      `json:"payload"`
	Status    EntryStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

// WAL 是 Write-Ahead Log 的實作
//
// 使用 sync.Mutex 確保多個 goroutine 同時寫入時的執行緒安全
// entries 模擬持久化儲存（實際場景可替換為檔案 I/O 或資料庫）
type WAL struct {
	mu      sync.Mutex
	entries []Entry
	nextID  uint64
}

// New 建立一個新的 WAL 實例
func New() *WAL {
	return &WAL{
		entries: make([]Entry, 0),
		nextID:  1,
	}
}

// Write 寫入一筆 pending 狀態的 WAL 記錄
//
// 在執行實際操作之前呼叫，確保即使後續操作失敗或系統崩潰，
// 我們仍然有記錄可以知道「有一個操作需要被執行」
//
// 範例：
//
//	entry := wal.Write("deduct_stock", `{"event_id":1,"quantity":2}`)
//	// 接著執行 Redis DECRBY ...
//	wal.Commit(entry.ID)
func (w *WAL) Write(operation, payload string) Entry {
	w.mu.Lock()
	defer w.mu.Unlock()

	entry := Entry{
		ID:        w.nextID,
		Operation: operation,
		Payload:   payload,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}
	w.nextID++
	w.entries = append(w.entries, entry)

	slog.Debug("WAL 寫入",
		"id", entry.ID,
		"operation", operation,
		"status", StatusPending,
	)

	return entry
}

// Commit 將指定 ID 的記錄標記為已提交
//
// 在實際操作（例如 Redis 扣庫存）成功後呼叫，
// 表示這筆操作已經完成，不需要在崩潰恢復時重新執行
func (w *WAL) Commit(id uint64) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i := range w.entries {
		if w.entries[i].ID == id {
			w.entries[i].Status = StatusCommitted
			slog.Debug("WAL 提交", "id", id)
			return nil
		}
	}

	return fmt.Errorf("WAL 記錄不存在: id=%d", id)
}

// Fail 將指定 ID 的記錄標記為失敗
//
// 在操作確定失敗且不需要重試時呼叫
func (w *WAL) Fail(id uint64) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i := range w.entries {
		if w.entries[i].ID == id {
			w.entries[i].Status = StatusFailed
			slog.Debug("WAL 標記失敗", "id", id)
			return nil
		}
	}

	return fmt.Errorf("WAL 記錄不存在: id=%d", id)
}

// Recover 回傳所有 pending 狀態的記錄，用於崩潰後的恢復
//
// 系統重啟時呼叫此方法，取得所有「已寫入但尚未完成」的操作，
// 然後逐一重新執行這些操作
//
// 範例：
//
//	pendingEntries := wal.Recover()
//	for _, entry := range pendingEntries {
//	    // 根據 entry.Operation 和 entry.Payload 重新執行操作
//	    replayOperation(entry)
//	    wal.Commit(entry.ID)
//	}
func (w *WAL) Recover() []Entry {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 收集所有 pending 狀態的記錄（不修改原始 slice）
	pending := make([]Entry, 0)
	for _, entry := range w.entries {
		if entry.Status == StatusPending {
			pending = append(pending, entry)
		}
	}

	slog.Info("WAL 恢復", "pending_count", len(pending))
	return pending
}

// Len 回傳目前 WAL 中的總記錄數（含所有狀態）
func (w *WAL) Len() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.entries)
}
