// Package idempotency 實作冪等性檢查，防止 MQ 重送導致重複處理（第 40 課）
//
// 冪等性的核心概念：
//   - 同一個操作無論執行一次或多次，結果都相同
//   - 在 Message Queue 的「至少一次投遞」(at-least-once) 語意下，
//     消費者可能會收到重複的訊息，必須有機制防止重複處理
//
// 搭配搶票系統的使用情境：
//   - 付款 Worker 從 MQ 收到 "order.created" 訊息
//   - 先用 idempotency.Check(orderID) 檢查是否已處理過
//   - 如果沒處理過，執行付款邏輯，然後 Mark(orderID) 標記為已處理
//   - 如果已處理過，直接跳過（避免重複扣款）
//
// 使用 sync.Map 實現無鎖的高併發安全存取
package idempotency

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// entry 代表一筆冪等性記錄
type entry struct {
	markedAt time.Time
	expireAt time.Time // zero value 表示永不過期
}

func (e *entry) expired() bool {
	return !e.expireAt.IsZero() && time.Now().After(e.expireAt)
}

// Store 冪等性鍵值儲存
//
// 內部使用 sync.Map 來儲存已處理的 key，
// sync.Map 針對「寫入少、讀取多」的場景有更好的效能，
// 非常適合冪等性檢查（大部分請求是首次，只有少數是重複的）
type Store struct {
	processed sync.Map
}

// NewStore 建立一個冪等性儲存，並啟動背景清理 goroutine
//
// cleanupInterval 決定多久掃描一次過期的 key（建議 30 秒 ~ 1 分鐘）
// 當 ctx 取消時，清理 goroutine 會自動停止
//
// 設計要點：用一個 goroutine 定期掃描，取代每筆 key 各開一個 time.AfterFunc goroutine
// 在 1000 req/s + 5min TTL 的場景下，前者只有 1 個 goroutine，後者會有 30 萬個
func NewStore(ctx context.Context, cleanupInterval time.Duration) *Store {
	s := &Store{}
	go s.cleanupLoop(ctx, cleanupInterval)
	return s
}

// cleanupLoop 定期掃描並移除過期的 key
func (s *Store) cleanupLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Debug("冪等性清理 goroutine 已停止")
			return
		case <-ticker.C:
			removed := 0
			s.processed.Range(func(key, value any) bool {
				if e := value.(*entry); e.expired() {
					s.processed.Delete(key)
					removed++
				}
				return true
			})
			if removed > 0 {
				slog.Debug("冪等性清理完成", "removed", removed)
			}
		}
	}
}

// Check 檢查指定的 key 是否已被處理過
//
// 回傳 true 表示「已處理過」，呼叫端應跳過該操作
// 回傳 false 表示「尚未處理」，呼叫端應繼續執行操作
//
// 如果 key 存在但已過期，視為「尚未處理」並自動清除
//
// 範例：
//
//	if store.Check("order:123") {
//	    slog.Info("訂單已處理過，跳過", "order_id", 123)
//	    return nil
//	}
//	// 執行付款邏輯...
//	store.Mark("order:123")
func (s *Store) Check(key string) bool {
	val, exists := s.processed.Load(key)
	if !exists {
		return false
	}
	e := val.(*entry)
	if e.expired() {
		s.processed.Delete(key)
		return false
	}
	slog.Debug("冪等性檢查: 重複的 key", "key", key)
	return true
}

// Mark 將指定的 key 標記為已處理（永久有效）
//
// 在操作成功完成後呼叫，確保同一個 key 不會被重複處理
// 注意：此方法標記的 key 永遠不會過期，適合小規模場景
// 大規模場景請使用 MarkWithTTL 設定自動過期
func (s *Store) Mark(key string) {
	s.processed.Store(key, &entry{markedAt: time.Now()})
	slog.Debug("冪等性標記", "key", key)
}

// MarkWithTTL 將指定的 key 標記為已處理，並在 ttl 時間後自動過期
//
// 過期的 key 會由背景清理 goroutine 定期移除，也會在 Check 時即時清除
// 不再為每筆 key 啟動獨立 goroutine，避免高流量下的 goroutine 爆炸
//
// 建議 TTL 設定：
//   - 付款處理：5~10 分鐘（足夠覆蓋 MQ 重試窗口）
//   - 庫存扣減：1~3 分鐘
//
// 範例：
//
//	store.MarkWithTTL("payment:order:123", 5*time.Minute)
func (s *Store) MarkWithTTL(key string, ttl time.Duration) {
	s.processed.Store(key, &entry{
		markedAt: time.Now(),
		expireAt: time.Now().Add(ttl),
	})
	slog.Debug("冪等性標記（含 TTL）", "key", key, "ttl", ttl)
}

// Size 回傳目前儲存中的 key 數量（主要用於測試和監控）
func (s *Store) Size() int {
	count := 0
	s.processed.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}
