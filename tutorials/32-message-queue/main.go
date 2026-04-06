// ==========================================================================
// 第三十二課：Message Queue（訊息佇列）
// ==========================================================================
//
// 為什麼需要訊息佇列？
//   同步呼叫：服務 A 直接呼叫服務 B → A 必須等 B 回應
//   非同步佇列：A 把訊息放進佇列 → A 繼續跑，B 稍後處理
//
// 訊息佇列解決的問題：
//   - 解耦（Decoupling）：生產者和消費者互不知道對方
//   - 削峰（Load Leveling）：流量突增時，訊息排隊，消費者按自己的速度處理
//   - 可靠性（Reliability）：訊息持久化，消費者掛掉重啟後繼續處理
//
// 本課示範：用 channel 模擬 Message Queue 的核心概念
//   （不需要安裝 Redis/Kafka，純 Go 實作）
//   真實專案：Redis Streams、RabbitMQ、Kafka、NATS
//
// 執行方式：go run ./tutorials/33-message-queue
// ==========================================================================

package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ==========================================================================
// 1. 訊息定義
// ==========================================================================

// MessageType 訊息類型
type MessageType string

const (
	MsgTypeOrderCreated    MessageType = "order.created"    // 訂單建立
	MsgTypeOrderPaid       MessageType = "order.paid"       // 訂單付款
	MsgTypeEmailSend       MessageType = "email.send"       // 寄送 Email
	MsgTypeInventoryDeduct MessageType = "inventory.deduct" // 庫存扣除
)

// Message 訊息結構
type Message struct {
	ID         string      // 訊息唯一 ID
	Type       MessageType // 訊息類型
	Payload    any         // 訊息內容
	Timestamp  time.Time   // 發送時間
	RetryCount int         // 重試次數
}

func (m Message) String() string {
	return fmt.Sprintf("Message{ID: %s, Type: %s, Retry: %d}", m.ID, m.Type, m.RetryCount)
}

// ==========================================================================
// 2. 簡單的 Channel-based Queue
// ==========================================================================

// Queue 基於 channel 的簡單訊息佇列
type Queue struct {
	name  string
	ch    chan Message
	mu    sync.RWMutex
	stats QueueStats
}

// QueueStats 佇列統計
type QueueStats struct {
	Published int64 // 發布總數
	Consumed  int64 // 消費總數
	Failed    int64 // 失敗總數
	Retried   int64 // 重試總數
}

// NewQueue 建立新佇列
func NewQueue(name string, bufferSize int) *Queue {
	return &Queue{
		name: name,
		ch:   make(chan Message, bufferSize),
	}
}

// Publish 發布訊息
func (q *Queue) Publish(msg Message) error {
	select {
	case q.ch <- msg:
		atomic.AddInt64(&q.stats.Published, 1)
		return nil
	default:
		// 佇列已滿
		return fmt.Errorf("佇列 [%s] 已滿，無法發布訊息 %s", q.name, msg.ID)
	}
}

// Subscribe 訂閱訊息（返回 channel，呼叫者自行讀取）
func (q *Queue) Subscribe() <-chan Message {
	return q.ch
}

// Len 取得佇列長度
func (q *Queue) Len() int { return len(q.ch) }

// ==========================================================================
// 3. 消費者（Consumer）
// ==========================================================================

// Handler 訊息處理函式
type Handler func(msg Message) error

// Consumer 訊息消費者
type Consumer struct {
	name     string
	queue    *Queue
	handler  Handler
	maxRetry int
}

// NewConsumer 建立消費者
func NewConsumer(name string, queue *Queue, handler Handler, maxRetry int) *Consumer {
	return &Consumer{name: name, queue: queue, handler: handler, maxRetry: maxRetry}
}

// Start 開始消費（在背景 goroutine 中執行）
func (c *Consumer) Start(ctx context.Context) {
	go func() {
		fmt.Printf("  消費者 [%s] 啟動，監聽佇列 [%s]\n", c.name, c.queue.name)
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("  消費者 [%s] 收到停止信號，退出\n", c.name)
				return
			case msg := <-c.queue.Subscribe():
				c.process(msg)
			}
		}
	}()
}

// process 處理訊息（含重試邏輯）
func (c *Consumer) process(msg Message) {
	err := c.handler(msg)
	if err == nil {
		atomic.AddInt64(&c.queue.stats.Consumed, 1)
		return
	}

	// 處理失敗，判斷是否重試
	if msg.RetryCount < c.maxRetry {
		msg.RetryCount++
		atomic.AddInt64(&c.queue.stats.Retried, 1)
		fmt.Printf("  ⚠️  [%s] 處理失敗，重試 %d/%d: %s\n",
			c.name, msg.RetryCount, c.maxRetry, err)

		// 延遲重試（指數退避 Exponential Backoff）
		delay := time.Duration(msg.RetryCount*100) * time.Millisecond
		time.Sleep(delay)

		// 重新放入佇列
		if putErr := c.queue.Publish(msg); putErr != nil {
			fmt.Printf("  ❌ [%s] 重試入佇列失敗: %v\n", c.name, putErr)
			atomic.AddInt64(&c.queue.stats.Failed, 1)
		}
	} else {
		// 超過重試上限，進入 Dead Letter Queue（死信佇列）
		atomic.AddInt64(&c.queue.stats.Failed, 1)
		fmt.Printf("  💀 [%s] 已達最大重試次數，訊息進入 DLQ: %s\n", c.name, msg)
	}
}

// ==========================================================================
// 4. 訊息內容定義
// ==========================================================================

// Order 訂單
type Order struct {
	ID     string
	UserID string
	Amount float64
	Items  []string
}

// EmailTask Email 任務
type EmailTask struct {
	To      string
	Subject string
	Body    string
}

// ==========================================================================
// 5. 各種服務的 Handler
// ==========================================================================

// retryCount 用於示範重試（模擬第一次失敗）
var retryCount = make(map[string]int)
var retryMu sync.Mutex

// emailHandler 處理 Email 發送
func emailHandler(msg Message) error {
	task, ok := msg.Payload.(EmailTask)
	if !ok {
		return errors.New("無效的 Email 任務格式")
	}

	// 模擬第一次失敗（示範重試機制）
	retryMu.Lock()
	count := retryCount[msg.ID]
	retryCount[msg.ID]++
	retryMu.Unlock()

	if count == 0 && msg.Type == MsgTypeEmailSend {
		return errors.New("SMTP 連線失敗（模擬）")
	}

	fmt.Printf("  📧 [Email 服務] 寄送郵件給 %s: %s\n", task.To, task.Subject)
	time.Sleep(50 * time.Millisecond) // 模擬發送時間
	return nil
}

// inventoryHandler 處理庫存扣除
func inventoryHandler(msg Message) error {
	order, ok := msg.Payload.(Order)
	if !ok {
		return errors.New("無效的訂單格式")
	}
	fmt.Printf("  📦 [庫存服務] 扣除訂單 %s 的庫存（%d 個商品）\n",
		order.ID, len(order.Items))
	time.Sleep(30 * time.Millisecond)
	return nil
}

// analyticsHandler 處理分析數據（示範 Fan-out：多個消費者訂閱同一佇列）
func analyticsHandler(msg Message) error {
	order, ok := msg.Payload.(Order)
	if !ok {
		return nil // 分析服務忽略非訂單訊息
	}
	fmt.Printf("  📊 [分析服務] 記錄訂單 %s，金額 %.2f\n", order.ID, order.Amount)
	return nil
}

// ==========================================================================
// 示範
// ==========================================================================

func demonstrateBasicQueue() {
	fmt.Println("=== 1. 基本發布/訂閱模式 ===")
	fmt.Println()

	// 建立佇列
	orderQueue := NewQueue("orders", 100)
	emailQueue := NewQueue("emails", 100)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 啟動消費者
	emailConsumer := NewConsumer("email-worker", emailQueue, emailHandler, 3)
	inventoryConsumer := NewConsumer("inventory-worker", orderQueue, inventoryHandler, 2)
	analyticsConsumer := NewConsumer("analytics-worker", orderQueue, analyticsHandler, 1)

	emailConsumer.Start(ctx)
	inventoryConsumer.Start(ctx)
	analyticsConsumer.Start(ctx) // Fan-out：同一個 orderQueue 有兩個消費者

	time.Sleep(50 * time.Millisecond) // 等消費者啟動

	// 生產者：建立訂單
	orders := []Order{
		{ID: "ORD-001", UserID: "user-1", Amount: 299.0, Items: []string{"Go 書籍", "鍵盤"}},
		{ID: "ORD-002", UserID: "user-2", Amount: 1299.0, Items: []string{"顯示器"}},
		{ID: "ORD-003", UserID: "user-3", Amount: 49.0, Items: []string{"滑鼠墊"}},
	}

	fmt.Println("發布訂單訊息...")
	for i, order := range orders {
		// 發布到訂單佇列（庫存、分析服務會收到）
		err := orderQueue.Publish(Message{
			ID:        fmt.Sprintf("msg-%d", i+1),
			Type:      MsgTypeOrderCreated,
			Payload:   order,
			Timestamp: time.Now(),
		})
		if err != nil {
			fmt.Printf("  發布失敗: %v\n", err)
			continue
		}
		fmt.Printf("  ✉️  發布訂單 %s\n", order.ID)

		// 同時發布 Email 訊息
		_ = emailQueue.Publish(Message{
			ID:   fmt.Sprintf("email-%d", i+1),
			Type: MsgTypeEmailSend,
			Payload: EmailTask{
				To:      fmt.Sprintf("user%d@example.com", i+1),
				Subject: fmt.Sprintf("訂單確認：%s", order.ID),
				Body:    fmt.Sprintf("您的訂單 %s 已建立", order.ID),
			},
			Timestamp: time.Now(),
		})
	}

	// 等待消費者處理
	time.Sleep(800 * time.Millisecond)

	fmt.Println()
	fmt.Println("佇列統計:")
	fmt.Printf("  [訂單佇列] 已發布: %d, 已消費: %d（庫存+分析各計一次）\n",
		orderQueue.stats.Published, orderQueue.stats.Consumed)
	fmt.Printf("  [Email佇列] 已發布: %d, 已消費: %d, 重試: %d\n",
		emailQueue.stats.Published, emailQueue.stats.Consumed, emailQueue.stats.Retried)
}

func demonstrateLoadLeveling() {
	fmt.Println("\n=== 2. 削峰（Load Leveling）===")
	fmt.Println()
	fmt.Println("場景：瞬間收到 50 個訂單，但庫存服務每次只能處理一個")
	fmt.Println("沒有佇列：系統直接崩潰或大量請求超時")
	fmt.Println("有佇列：訊息排隊，消費者按自己的速度處理")
	fmt.Println()

	queue := NewQueue("spike-test", 100)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	processed := int64(0)
	consumer := NewConsumer("slow-worker", queue, func(msg Message) error {
		time.Sleep(20 * time.Millisecond) // 模擬慢速消費者
		atomic.AddInt64(&processed, 1)
		return nil
	}, 0)
	consumer.Start(ctx)

	// 瞬間發布 20 個訊息（模擬流量突增）
	start := time.Now()
	for i := range 20 {
		_ = queue.Publish(Message{
			ID:        fmt.Sprintf("spike-%d", i),
			Type:      MsgTypeOrderCreated,
			Payload:   Order{ID: fmt.Sprintf("ORD-%d", i)},
			Timestamp: time.Now(),
		})
	}
	fmt.Printf("20 個訊息在 %v 內全部放入佇列（生產者不等待）\n", time.Since(start))
	fmt.Println("消費者慢慢處理中...")

	// 等待消費者處理完
	<-ctx.Done()
	fmt.Printf("消費者處理了 %d 個訊息（剩餘 %d 個在佇列中）\n",
		atomic.LoadInt64(&processed), queue.Len())
}

func main() {
	fmt.Println("==========================================")
	fmt.Println(" 第三十三課：Message Queue（訊息佇列）")
	fmt.Println("==========================================")
	fmt.Println()

	demonstrateBasicQueue()
	demonstrateLoadLeveling()

	// ──── 真實世界的選擇 ────
	fmt.Println("\n=== 真實世界的 Message Queue 選擇 ===")
	fmt.Println()
	fmt.Println("| 技術          | 特點                          | 適用場景              |")
	fmt.Println("|---------------|-------------------------------|----------------------|")
	fmt.Println("| Redis Streams | 輕量、低延遲、持久化            | 中小規模、速度優先    |")
	fmt.Println("| RabbitMQ      | 完整 AMQP、複雜路由            | 企業級、需要 ACK 確認 |")
	fmt.Println("| Apache Kafka  | 超高吞吐、可重播、分散式        | 大數據、事件溯源      |")
	fmt.Println("| NATS          | 超低延遲、Go 原生              | 微服務通訊、IoT       |")
	fmt.Println("| Go channel    | 單機、無需外部依賴（本課方式）  | 學習概念、簡單場景    |")
	fmt.Println()
	fmt.Println("核心概念（與使用什麼技術無關）：")
	fmt.Println("  - Publisher/Producer  → 發送訊息")
	fmt.Println("  - Subscriber/Consumer → 接收訊息")
	fmt.Println("  - Dead Letter Queue   → 處理失敗的訊息")
	fmt.Println("  - Idempotency         → 消費者要能安全地重複處理同一訊息")
	fmt.Println("  - At-least-once       → 訊息至少被處理一次（可能重複）")
	fmt.Println("  - Exactly-once        → 訊息剛好被處理一次（很難實作）")

	fmt.Println("\n==========================================")
	fmt.Println(" 教學完成！")
	fmt.Println("==========================================")
}
