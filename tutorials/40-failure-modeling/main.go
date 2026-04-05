// ==========================================================================
// 第四十課：分散式系統容錯（Failure Modeling）
// ==========================================================================
//
// 這一課回答一個資深工程師面試題：
//   「Payment 成功但 Order Service crash，錢扣了卻沒票，怎麼辦？」
//
// 核心觀念：
//   1. WAL（Write-Ahead Log）— 先寫日誌再執行操作，crash 後可重播
//   2. Saga Pattern — 分散式交易的補償機制
//   3. Idempotency（冪等性）— 重試不會重複扣款/出票
//   4. Outbox Pattern — 確保 DB 寫入和訊息發送的一致性
//
// 執行方式：
//   go run ./tutorials/38-failure-modeling/
// ==========================================================================

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

// ==========================================================================
// 1. WAL（Write-Ahead Log）— 先寫日誌再做事
// ==========================================================================
// 問題：Redis queue 裡有 100 萬人排隊，Redis crash 了，queue 全部消失
// 解法：每次操作前先寫一條日誌到持久化儲存，crash 後從日誌重播

// WALEntry 日誌條目
type WALEntry struct {
	ID        string    `json:"id"`
	Operation string    `json:"operation"` // "enqueue", "dequeue", "lock_seat", "payment"
	Payload   string    `json:"payload"`
	Status    string    `json:"status"` // "pending", "committed", "rolled_back"
	CreatedAt time.Time `json:"created_at"`
}

// WAL Write-Ahead Log 實作（生產環境用 DB 或 Kafka）
type WAL struct {
	mu      sync.Mutex
	entries []WALEntry
}

func NewWAL() *WAL {
	return &WAL{}
}

// Append 寫入日誌（在執行操作之前呼叫）
func (w *WAL) Append(entry WALEntry) {
	w.mu.Lock()
	defer w.mu.Unlock()
	entry.CreatedAt = time.Now()
	w.entries = append(w.entries, entry)
	slog.Info("[WAL] 寫入日誌", "id", entry.ID, "op", entry.Operation, "status", entry.Status)
}

// MarkCommitted 標記操作已完成
func (w *WAL) MarkCommitted(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for i := range w.entries {
		if w.entries[i].ID == id {
			w.entries[i].Status = "committed"
			break
		}
	}
}

// GetPending 取得所有未完成的操作（crash recovery 用）
func (w *WAL) GetPending() []WALEntry {
	w.mu.Lock()
	defer w.mu.Unlock()
	var pending []WALEntry
	for _, e := range w.entries {
		if e.Status == "pending" {
			pending = append(pending, e)
		}
	}
	return pending
}

// ==========================================================================
// 2. Saga Pattern — 分散式交易補償
// ==========================================================================
// 問題：搶票涉及多個服務（庫存、訂單、支付），任何一步失敗都要回滾前面的步驟
// 解法：每個步驟都有對應的「補償動作」

// SagaStep 定義 Saga 的一個步驟
type SagaStep struct {
	Name       string
	Execute    func(ctx context.Context) error // 正向操作
	Compensate func(ctx context.Context) error // 補償操作（回滾）
}

// Saga 協調器
type Saga struct {
	steps     []SagaStep
	completed []int // 已完成的步驟索引
}

func NewSaga() *Saga {
	return &Saga{}
}

func (s *Saga) AddStep(step SagaStep) {
	s.steps = append(s.steps, step)
}

// Run 執行 Saga，失敗時自動補償已完成的步驟
func (s *Saga) Run(ctx context.Context) error {
	for i, step := range s.steps {
		slog.Info("[Saga] 執行步驟", "step", step.Name, "index", i)

		if err := step.Execute(ctx); err != nil {
			slog.Error("[Saga] 步驟失敗，開始補償", "step", step.Name, "error", err)
			s.compensate(ctx)
			return fmt.Errorf("saga failed at step %s: %w", step.Name, err)
		}

		s.completed = append(s.completed, i)
		slog.Info("[Saga] 步驟完成", "step", step.Name)
	}

	slog.Info("[Saga] 所有步驟完成")
	return nil
}

// compensate 反向補償已完成的步驟
func (s *Saga) compensate(ctx context.Context) {
	// 反向執行補償（後完成的先補償）
	for i := len(s.completed) - 1; i >= 0; i-- {
		idx := s.completed[i]
		step := s.steps[idx]
		slog.Warn("[Saga] 補償步驟", "step", step.Name)

		if err := step.Compensate(ctx); err != nil {
			slog.Error("[Saga] 補償失敗（需要人工介入）", "step", step.Name, "error", err)
		}
	}
}

// ==========================================================================
// 3. Idempotency（冪等性）— 重試安全
// ==========================================================================
// 問題：Message Queue 重送了兩次相同的訂單，會不會重複扣款？
// 解法：用 idempotency key 確保同一個操作只執行一次

// IdempotencyStore 冪等性檢查器
type IdempotencyStore struct {
	mu   sync.Mutex
	keys map[string]bool
}

func NewIdempotencyStore() *IdempotencyStore {
	return &IdempotencyStore{keys: make(map[string]bool)}
}

// TryExecute 嘗試執行操作，如果已經執行過則跳過
func (s *IdempotencyStore) TryExecute(key string, fn func() error) error {
	s.mu.Lock()
	if s.keys[key] {
		s.mu.Unlock()
		slog.Warn("[冪等] 操作已執行過，跳過", "key", key)
		return nil // 已執行過，安全跳過
	}
	s.keys[key] = true
	s.mu.Unlock()

	return fn()
}

// ==========================================================================
// 4. Outbox Pattern — DB + Message 一致性
// ==========================================================================
// 問題：訂單寫入 DB 成功，但發送 MQ 訊息失敗，下游服務不知道有新訂單
// 解法：把訊息寫入 DB 的 outbox 表（同一個 transaction），再由背景程式發送

// OutboxEntry 待發送的訊息
type OutboxEntry struct {
	ID        string `json:"id"`
	Topic     string `json:"topic"`
	Payload   string `json:"payload"`
	Published bool   `json:"published"`
}

// OutboxPublisher 模擬 Outbox 發送器
type OutboxPublisher struct {
	mu      sync.Mutex
	entries []OutboxEntry
}

func NewOutboxPublisher() *OutboxPublisher {
	return &OutboxPublisher{}
}

// SaveWithMessage 在同一個「交易」中儲存資料和訊息
func (p *OutboxPublisher) SaveWithMessage(orderID, topic, payload string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 真實環境：在同一個 DB transaction 中 INSERT order + INSERT outbox
	p.entries = append(p.entries, OutboxEntry{
		ID:      orderID,
		Topic:   topic,
		Payload: payload,
	})
	slog.Info("[Outbox] 訊息已寫入 outbox 表", "order_id", orderID, "topic", topic)
}

// PublishPending 背景程式定期發送未發布的訊息
func (p *OutboxPublisher) PublishPending() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := range p.entries {
		if !p.entries[i].Published {
			// 真實環境：發送到 Kafka/RabbitMQ
			slog.Info("[Outbox] 發送訊息", "id", p.entries[i].ID, "topic", p.entries[i].Topic)
			p.entries[i].Published = true
		}
	}
}

// ==========================================================================
// 演示
// ==========================================================================

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	ctx := context.Background()

	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║      第 38 課：分散式系統容錯（Failure Modeling）        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")

	// === Demo 1: WAL ===
	fmt.Println("\n📝 Demo 1: Write-Ahead Log（防止 Redis crash 丟失 queue）")
	fmt.Println("─────────────────────────────────────────────")

	wal := NewWAL()

	// 模擬排隊：先寫 WAL，再寫 Redis
	for i := 1; i <= 5; i++ {
		id := fmt.Sprintf("user-%d", i)
		wal.Append(WALEntry{ID: id, Operation: "enqueue", Payload: id, Status: "pending"})
	}
	// 模擬 3 個成功入隊
	wal.MarkCommitted("user-1")
	wal.MarkCommitted("user-2")
	wal.MarkCommitted("user-3")

	// 模擬 Redis crash！user-4, user-5 還沒入隊
	fmt.Println("\n💥 Redis crash！正在從 WAL 恢復...")
	pending := wal.GetPending()
	fmt.Printf("   需要重播 %d 筆操作：\n", len(pending))
	for _, e := range pending {
		fmt.Printf("   - 重新入隊：%s\n", e.ID)
		wal.MarkCommitted(e.ID)
	}
	fmt.Println("   ✅ WAL 恢復完成，沒有丟失任何排隊使用者")

	// === Demo 2: Saga ===
	fmt.Println("\n🔄 Demo 2: Saga Pattern（Payment 成功但 Order crash）")
	fmt.Println("─────────────────────────────────────────────")

	saga := NewSaga()

	saga.AddStep(SagaStep{
		Name:       "1. 扣庫存",
		Execute:    func(ctx context.Context) error { fmt.Println("   ✅ 庫存 -1"); return nil },
		Compensate: func(ctx context.Context) error { fmt.Println("   ↩️  庫存 +1（回滾）"); return nil },
	})
	saga.AddStep(SagaStep{
		Name:       "2. 建立訂單",
		Execute:    func(ctx context.Context) error { fmt.Println("   ✅ 訂單已建立"); return nil },
		Compensate: func(ctx context.Context) error { fmt.Println("   ↩️  訂單已取消（回滾）"); return nil },
	})
	saga.AddStep(SagaStep{
		Name: "3. 支付",
		Execute: func(ctx context.Context) error {
			fmt.Println("   ✅ 支付成功")
			return nil
		},
		Compensate: func(ctx context.Context) error { fmt.Println("   ↩️  退款（回滾）"); return nil },
	})
	saga.AddStep(SagaStep{
		Name: "4. 確認出票",
		Execute: func(ctx context.Context) error {
			return fmt.Errorf("Order Service crash！") // 模擬 crash
		},
		Compensate: func(ctx context.Context) error { return nil },
	})

	err := saga.Run(ctx)
	if err != nil {
		fmt.Printf("\n   結果：Saga 自動補償完成，使用者不會「錢扣了沒票」\n")
	}

	// === Demo 3: Idempotency ===
	fmt.Println("\n🔑 Demo 3: Idempotency（MQ 重送不會重複扣款）")
	fmt.Println("─────────────────────────────────────────────")

	idem := NewIdempotencyStore()
	paymentKey := "payment:order-123"

	// 第一次執行
	idem.TryExecute(paymentKey, func() error {
		fmt.Println("   ✅ 第 1 次：扣款 $2800")
		return nil
	})

	// MQ 重送！第二次執行
	idem.TryExecute(paymentKey, func() error {
		fmt.Println("   ❌ 第 2 次：扣款 $2800（不應該出現）")
		return nil
	})

	// MQ 又重送！第三次
	idem.TryExecute(paymentKey, func() error {
		fmt.Println("   ❌ 第 3 次：扣款 $2800（不應該出現）")
		return nil
	})

	fmt.Println("   ✅ 結果：只扣了一次款")

	// === Demo 4: Outbox ===
	fmt.Println("\n📤 Demo 4: Outbox Pattern（DB + MQ 一致性）")
	fmt.Println("─────────────────────────────────────────────")

	outbox := NewOutboxPublisher()

	// 在同一個 DB transaction 中儲存訂單 + outbox 訊息
	payload, _ := json.Marshal(map[string]any{"order_id": 123, "amount": 2800})
	outbox.SaveWithMessage("order-123", "order.confirmed", string(payload))
	outbox.SaveWithMessage("order-456", "order.confirmed", string(payload))

	fmt.Println("\n   背景程式定期發送 outbox 中的訊息：")
	outbox.PublishPending()
	fmt.Println("   ✅ 即使 MQ 當時不可用，訊息也不會丟失（因為在 DB 裡）")

	// === 總結 ===
	fmt.Println("\n" + "═══════════════════════════════════════════════════════════")
	fmt.Println("📌 總結：資深工程師的容錯思維")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("  問題                          │ 解法")
	fmt.Println("  ─────────────────────────────┼──────────────────────────")
	fmt.Println("  Redis crash queue 消失       │ WAL：先寫日誌再操作")
	fmt.Println("  Payment 成功但 Order crash   │ Saga：自動補償回滾")
	fmt.Println("  MQ 重送導致重複扣款          │ Idempotency：冪等 key")
	fmt.Println("  DB 成功但 MQ 發送失敗        │ Outbox：DB + MQ 同交易")
	fmt.Println()
}
