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
	"log/slog"
	"sync"
	"time"
)

// Store 冪等性鍵值儲存
//
// 內部使用 sync.Map 來儲存已處理的 key，
// sync.Map 針對「寫入少、讀取多」的場景有更好的效能，
// 非常適合冪等性檢查（大部分請求是首次，只有少數是重複的）
type Store struct {
	processed sync.Map
}

// NewStore 建立一個新的冪等性儲存
func NewStore() *Store {
	return &Store{}
}

// Check 檢查指定的 key 是否已被處理過
//
// 回傳 true 表示「已處理過」，呼叫端應跳過該操作
// 回傳 false 表示「尚未處理」，呼叫端應繼續執行操作
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
	_, exists := s.processed.Load(key)
	if exists {
		slog.Debug("冪等性檢查: 重複的 key", "key", key)
	}
	return exists
}

// Mark 將指定的 key 標記為已處理（永久有效）
//
// 在操作成功完成後呼叫，確保同一個 key 不會被重複處理
// 注意：此方法標記的 key 永遠不會過期，適合小規模場景
// 大規模場景請使用 MarkWithTTL 設定自動過期
func (s *Store) Mark(key string) {
	s.processed.Store(key, time.Now())
	slog.Debug("冪等性標記", "key", key)
}

// MarkWithTTL 將指定的 key 標記為已處理，並在 ttl 時間後自動過期
//
// 使用 time.AfterFunc 在背景 goroutine 中於到期後自動清除 key
// 這樣可以：
//   - 防止記憶體無限增長
//   - 在合理的時間窗口內防止重複處理
//   - TTL 過期後，允許重新處理（適合重試機制）
//
// 建議 TTL 設定：
//   - 付款處理：5~10 分鐘（足夠覆蓋 MQ 重試窗口）
//   - 庫存扣減：1~3 分鐘
//
// 範例：
//
//	store.MarkWithTTL("payment:order:123", 5*time.Minute)
func (s *Store) MarkWithTTL(key string, ttl time.Duration) {
	s.processed.Store(key, time.Now())
	slog.Debug("冪等性標記（含 TTL）", "key", key, "ttl", ttl)

	// 在 TTL 到期後自動移除 key
	// time.AfterFunc 會啟動一個獨立的 goroutine，在指定時間後執行清除
	time.AfterFunc(ttl, func() {
		s.processed.Delete(key)
		slog.Debug("冪等性 key 已過期移除", "key", key)
	})
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
