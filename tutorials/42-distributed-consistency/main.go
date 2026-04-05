// ==========================================================================
// 第四十二課：分散式一致性（Distributed Consistency）
// ==========================================================================
//
// 這一課回答最難的問題：
//   「庫存扣了、錢也付了，但訂單服務 crash，怎麼保證一致性？」
//
// 核心觀念：
//   1. CAP 定理 — 你不可能同時擁有一致性、可用性、分區容錯
//   2. 最終一致性 — 不追求「立即一致」，而是「最終會一致」
//   3. Inventory Token — 先發 token 再搶票，避免超賣
//   4. 2PC vs Saga — 強一致 vs 最終一致的取捨
//   5. Distributed Lock — 正確使用分散式鎖
//
// 執行方式：
//   go run ./tutorials/40-distributed-consistency/
// ==========================================================================

package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// ==========================================================================
// 1. Inventory Token — 先發 token 再搶票
// ==========================================================================
// 問題：5000 張票，100 萬人同時 SELECT seat → Redis 被打爆
// 解法：只發 5000 個 token，沒有 token 的人根本不能選座位

// TokenBucket 令牌桶（控制進入購票的人數）
type TokenBucket struct {
	mu        sync.Mutex
	tokens    map[string]bool // token → 是否已使用
	remaining int
	total     int
}

func NewTokenBucket(total int) *TokenBucket {
	return &TokenBucket{
		tokens:    make(map[string]bool),
		remaining: total,
		total:     total,
	}
}

// Acquire 嘗試獲取 token（原子操作，並發安全）
func (b *TokenBucket) Acquire(userID string) (string, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.remaining <= 0 {
		return "", false
	}

	token := fmt.Sprintf("TK-%s-%d", userID, time.Now().UnixNano())
	b.tokens[token] = false // false = 未使用
	b.remaining--
	return token, true
}

// Use 使用 token（確認購票）
func (b *TokenBucket) Use(token string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	used, exists := b.tokens[token]
	if !exists || used {
		return false
	}
	b.tokens[token] = true
	return true
}

// Release 釋放 token（超時未使用）
func (b *TokenBucket) Release(token string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if used, exists := b.tokens[token]; exists && !used {
		delete(b.tokens, token)
		b.remaining++
	}
}

// ==========================================================================
// 2. Distributed Lock（分散式鎖，模擬 Redis SET NX）
// ==========================================================================

// DistributedLock 模擬分散式鎖
type DistributedLock struct {
	mu    sync.Mutex
	locks map[string]lockInfo
}

type lockInfo struct {
	owner     string
	expiresAt time.Time
}

func NewDistributedLock() *DistributedLock {
	return &DistributedLock{locks: make(map[string]lockInfo)}
}

// TryLock 嘗試加鎖（等同 Redis SET key owner EX ttl NX）
func (d *DistributedLock) TryLock(key, owner string, ttl time.Duration) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 檢查是否已有鎖且未過期
	if info, exists := d.locks[key]; exists {
		if time.Now().Before(info.expiresAt) {
			return false // 鎖還在，無法獲取
		}
		// 鎖已過期，可以覆蓋
	}

	d.locks[key] = lockInfo{owner: owner, expiresAt: time.Now().Add(ttl)}
	return true
}

// Unlock 釋放鎖（只有 owner 可以釋放，防止誤刪）
func (d *DistributedLock) Unlock(key, owner string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	info, exists := d.locks[key]
	if !exists || info.owner != owner {
		return false // 不是你的鎖，不能釋放
	}

	delete(d.locks, key)
	return true
}

// ==========================================================================
// 3. 最終一致性 — Event Reconciliation
// ==========================================================================

// EventLog 事件日誌（用於最終一致性對帳）
type EventLog struct {
	mu     sync.Mutex
	events []ConsistencyEvent
}

type ConsistencyEvent struct {
	OrderID   string
	Service   string // "inventory", "payment", "order"
	Action    string // "deducted", "charged", "created"
	Timestamp time.Time
}

func NewEventLog() *EventLog {
	return &EventLog{}
}

func (l *EventLog) Record(orderID, service, action string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, ConsistencyEvent{
		OrderID:   orderID,
		Service:   service,
		Action:    action,
		Timestamp: time.Now(),
	})
}

// Reconcile 對帳：檢查每個訂單的三個服務是否都完成
func (l *EventLog) Reconcile() []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 收集每個訂單的服務狀態
	orderServices := make(map[string]map[string]bool)
	for _, e := range l.events {
		if orderServices[e.OrderID] == nil {
			orderServices[e.OrderID] = make(map[string]bool)
		}
		orderServices[e.OrderID][e.Service] = true
	}

	// 找出不一致的訂單
	var inconsistent []string
	required := []string{"inventory", "payment", "order"}
	for orderID, services := range orderServices {
		for _, svc := range required {
			if !services[svc] {
				inconsistent = append(inconsistent, fmt.Sprintf("%s 缺少 %s", orderID, svc))
			}
		}
	}
	return inconsistent
}

// ==========================================================================
// 演示
// ==========================================================================

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║     第 40 課：分散式一致性（Distributed Consistency）    ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")

	// === Demo 1: CAP 定理 ===
	fmt.Println("\n📐 Demo 1: CAP 定理")
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("  C（一致性）  A（可用性）  P（分區容錯）")
	fmt.Println("       ╲          │          ╱")
	fmt.Println("        ╲         │         ╱")
	fmt.Println("         ╲        │        ╱")
	fmt.Println("          ╲       │       ╱")
	fmt.Println("     你只能選兩個，不能三個都要")
	fmt.Println()
	fmt.Println("  搶票系統的選擇：")
	fmt.Println("  ┌─────────────────────────────────────────────────────┐")
	fmt.Println("  │ AP（可用性 + 分區容錯）                              │")
	fmt.Println("  │ → Redis crash 了系統還能跑（A），但庫存可能暫時不準（C）│")
	fmt.Println("  │ → 用「最終一致性」在事後修正                          │")
	fmt.Println("  └─────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  銀行轉帳的選擇：")
	fmt.Println("  ┌─────────────────────────────────────────────────────┐")
	fmt.Println("  │ CP（一致性 + 分區容錯）                              │")
	fmt.Println("  │ → 轉帳必須強一致（C），寧可暫時不可用（犧牲 A）       │")
	fmt.Println("  └─────────────────────────────────────────────────────┘")

	// === Demo 2: Inventory Token ===
	fmt.Println("\n🎫 Demo 2: Inventory Token（先發 token 再搶票）")
	fmt.Println("─────────────────────────────────────────────")

	bucket := NewTokenBucket(5) // 只有 5 張票

	fmt.Println("  只有 5 張票，10 個人同時搶：")
	var wg sync.WaitGroup
	var got, missed atomic.Int32

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			userID := fmt.Sprintf("user-%d", id)
			token, ok := bucket.Acquire(userID)
			if ok {
				got.Add(1)
				fmt.Printf("    %s → ✅ 拿到 token: %s\n", userID, token[:15]+"...")
			} else {
				missed.Add(1)
				fmt.Printf("    %s → ❌ 售罄\n", userID)
			}
		}(i)
	}
	wg.Wait()
	fmt.Printf("\n  結果：%d 人拿到 token，%d 人被擋（不會打到 Redis 選座位）\n", got.Load(), missed.Load())

	// === Demo 3: Distributed Lock ===
	fmt.Println("\n🔒 Demo 3: 分散式鎖（座位鎖定）")
	fmt.Println("─────────────────────────────────────────────")

	lock := NewDistributedLock()
	seatKey := "seat:A-10"

	// User-1 先鎖
	ok := lock.TryLock(seatKey, "user-1", 10*time.Minute)
	fmt.Printf("  user-1 嘗試鎖定 %s → %v\n", seatKey, ok)

	// User-2 也想鎖同一個座位
	ok = lock.TryLock(seatKey, "user-2", 10*time.Minute)
	fmt.Printf("  user-2 嘗試鎖定 %s → %v（已被 user-1 鎖了）\n", seatKey, ok)

	// User-1 付款成功，釋放鎖
	lock.Unlock(seatKey, "user-1")
	fmt.Printf("  user-1 付款成功，釋放 %s\n", seatKey)

	// User-3 嘗試鎖（但座位已賣出，業務層會擋）
	ok = lock.TryLock(seatKey, "user-3", 10*time.Minute)
	fmt.Printf("  user-3 嘗試鎖定 %s → %v（鎖已釋放，可以拿到）\n", seatKey, ok)

	// === Demo 4: 2PC vs Saga ===
	fmt.Println("\n⚖️ Demo 4: 2PC vs Saga 取捨")
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("  ┌─────────────┬───────────────────┬───────────────────┐")
	fmt.Println("  │             │ 2PC（兩階段提交）  │ Saga（補償交易）  │")
	fmt.Println("  ├─────────────┼───────────────────┼───────────────────┤")
	fmt.Println("  │ 一致性      │ 強一致            │ 最終一致          │")
	fmt.Println("  │ 效能        │ 慢（要等所有參與者）│ 快（異步）       │")
	fmt.Println("  │ 可用性      │ 低（任何一方掛就卡）│ 高（可以補償）   │")
	fmt.Println("  │ 複雜度      │ 中                │ 高（要寫補償邏輯）│")
	fmt.Println("  │ 適合場景    │ 銀行轉帳          │ 搶票、電商下單    │")
	fmt.Println("  └─────────────┴───────────────────┴───────────────────┘")
	fmt.Println()
	fmt.Println("  搶票系統選 Saga 的原因：")
	fmt.Println("  → 100 萬人搶票不能用 2PC（太慢、鎖太多）")
	fmt.Println("  → 寧可「先扣再補」（Saga 補償），也不要「等到確認才扣」（2PC 阻塞）")

	// === Demo 5: 最終一致性 — Event Reconciliation ===
	fmt.Println("\n🔍 Demo 5: 最終一致性 — 事件對帳")
	fmt.Println("─────────────────────────────────────────────")

	eventLog := NewEventLog()

	// 訂單 001：三個服務都完成了
	eventLog.Record("order-001", "inventory", "deducted")
	eventLog.Record("order-001", "payment", "charged")
	eventLog.Record("order-001", "order", "created")

	// 訂單 002：payment 成功但 order service crash 了
	eventLog.Record("order-002", "inventory", "deducted")
	eventLog.Record("order-002", "payment", "charged")
	// order service crash！沒有 "order" 事件

	// 訂單 003：只有 inventory 完成
	eventLog.Record("order-003", "inventory", "deducted")

	fmt.Println("  背景對帳程式每 30 秒執行一次：")
	inconsistent := eventLog.Reconcile()
	if len(inconsistent) == 0 {
		fmt.Println("  ✅ 所有訂單一致")
	} else {
		fmt.Println("  ⚠️ 發現不一致：")
		for _, msg := range inconsistent {
			fmt.Printf("    - %s\n", msg)
		}
		fmt.Println("\n  自動修復：")
		fmt.Println("    order-002 缺少 order → 重新建立訂單")
		fmt.Println("    order-003 缺少 payment → 觸發退款 + 回補庫存")
	}

	// === Demo 6: 模擬高併發超賣 ===
	fmt.Println("\n💥 Demo 6: 超賣問題 — 為什麼需要原子操作")
	fmt.Println("─────────────────────────────────────────────")

	// 不安全的計數器（會超賣）
	var unsafeStock int32 = 10
	var unsafeWg sync.WaitGroup
	var oversold atomic.Int32

	for i := 0; i < 100; i++ {
		unsafeWg.Add(1)
		go func() {
			defer unsafeWg.Done()
			// 不安全：先讀再減（race condition）
			if unsafeStock > 0 {
				time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
				unsafeStock-- // 可能多個 goroutine 同時過了 >0 檢查
				oversold.Add(1)
			}
		}()
	}
	unsafeWg.Wait()
	fmt.Printf("  不安全方式：10 張票，賣出 %d 張（超賣 %d 張！）\n",
		oversold.Load(), oversold.Load()-10)

	// 安全的方式（atomic 或 Redis DECR）
	var safeStock atomic.Int32
	safeStock.Store(10)
	var safeSold atomic.Int32

	var safeWg sync.WaitGroup
	for i := 0; i < 100; i++ {
		safeWg.Add(1)
		go func() {
			defer safeWg.Done()
			// 安全：原子扣減（等同 Redis DECRBY）
			if safeStock.Add(-1) >= 0 {
				safeSold.Add(1)
			} else {
				safeStock.Add(1) // 回滾
			}
		}()
	}
	safeWg.Wait()
	fmt.Printf("  原子操作：10 張票，賣出 %d 張（不超賣）\n", safeSold.Load())

	// === 總結 ===
	fmt.Println("\n" + "═══════════════════════════════════════════════════════════")
	fmt.Println("📌 總結：分散式一致性的核心思維")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("  1. 接受不一致是常態 — 用「最終一致」取代「強一致」")
	fmt.Println("  2. 先發 token 再操作 — 從源頭控制流量")
	fmt.Println("  3. 原子操作防超賣   — Redis DECR / atomic.Add")
	fmt.Println("  4. 事件對帳修復     — 背景程式定期檢查 + 補償")
	fmt.Println("  5. 選對取捨         — 搶票用 Saga，轉帳用 2PC")
	fmt.Println()

	_ = context.Background() // 避免 import 未使用
}
