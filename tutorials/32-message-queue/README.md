# 第三十二課：Message Queue（訊息佇列）

> **一句話總結**：訊息佇列就像郵局 — 寄信的人不需要等收信的人在家，信放進郵箱就好。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | 了解 Pub/Sub 概念，能實作基本的生產者/消費者 |
| 🔴 資深工程師 | **必備**：能設計可靠的訊息處理系統，包含 Dead Letter Queue、冪等性 |
| 🏢 架構師 | 評估 Redis/Kafka/RabbitMQ/NATS 的取捨 |

## 你會學到什麼？

- 為什麼需要訊息佇列（解耦、非同步、削峰）
- Pub/Sub vs Point-to-Point 兩種模式
- 訊息傳遞保證：At-most-once、At-least-once、Exactly-once
- Dead Letter Queue（死信佇列）處理失敗訊息
- Fan-out 模式：一個訊息廣播給多個消費者
- Backpressure（背壓）處理策略
- In-Memory MQ vs Redis Streams vs Kafka 的選擇

## 執行方式

```bash
go run ./tutorials/32-message-queue
```

---

## 郵局比喻：為什麼需要訊息佇列？

```
【沒有訊息佇列（同步）】              【有訊息佇列（非同步）】

  寄件人 ──電話──▶ 收件人              寄件人 ──信──▶ 郵箱 ──▶ 收件人
    │                │                    │           │          │
    │   必須同時在線  │                   │   放進去   │  稍後取  │
    │   等對方接電話  │                   │   就走了   │          │
    ▼                ▼                    ▼           ▼          ▼
  如果收件人不在 → 失敗            如果收件人不在 → 信在郵箱等著
  如果收件人很忙 → 等很久          如果收件人很忙 → 信排隊，慢慢處理
  如果一次打給 100 人 → 崩潰       如果一次寄 100 封 → 郵箱排隊，正常運作
```

### 訊息佇列解決的三大問題

| 問題 | 沒有 MQ | 有 MQ |
|------|---------|-------|
| **耦合** | 服務 A 直接呼叫服務 B，A 必須知道 B 的位址 | A 只管發訊息，B 只管收訊息，互不知道 |
| **削峰** | 流量突增時，下游服務直接被打爆 | 訊息排隊，消費者按自己的速度處理 |
| **可靠性** | B 掛了 → A 的請求直接失敗 | B 掛了 → 訊息在佇列中等 B 重啟 |

---

## Pub/Sub vs Point-to-Point

訊息佇列有兩種基本模式：

### Point-to-Point（點對點）

```
一個訊息只被一個消費者處理

  Producer ──msg──▶ [Queue] ──▶ Consumer A  ✅ 收到
                              ──▶ Consumer B  ❌ 沒收到（已被 A 消費）

適用場景：工作佇列（Task Queue）
  - 每個任務只需要處理一次
  - 例如：發送 Email、處理訂單
```

### Pub/Sub（發布/訂閱）

```
一個訊息被所有訂閱者收到

  Publisher ──msg──▶ [Topic] ──▶ Subscriber A  ✅ 收到
                               ──▶ Subscriber B  ✅ 收到
                               ──▶ Subscriber C  ✅ 收到

適用場景：事件廣播
  - 多個服務需要知道同一件事
  - 例如：「訂單建立」→ 庫存扣除 + Email 通知 + 數據分析
```

### 搶票系統中的實例

```
搶票成功 → Broker.Publish("order.created", order)
                            │
              ┌─────────────┼─────────────┐
              ▼             ▼             ▼
        PaymentWorker  StockBroadcaster  (可再加更多)
        (處理支付)     (推播庫存變更)
```

---

## 訊息傳遞保證（Delivery Guarantee）

這是訊息佇列最重要的概念之一：訊息到底會被處理幾次？

| 保證等級 | 說明 | 實作難度 | 適用場景 |
|---------|------|---------|---------|
| **At-most-once** | 最多一次（可能丟失） | 最簡單 | 日誌、監控指標 |
| **At-least-once** | 至少一次（可能重複） | 中等 | 大部分業務場景 |
| **Exactly-once** | 剛好一次（不丟不重） | 非常困難 | 金融交易 |

### At-most-once（最多一次）

```
Producer ──msg──▶ Queue ──▶ Consumer
                              │
                         處理失敗？
                              │
                         丟棄，不重試

實作：發了就忘（fire-and-forget）
缺點：訊息可能丟失
適合：不重要的通知、日誌
```

### At-least-once（至少一次）— 最常用

```
Producer ──msg──▶ Queue ──▶ Consumer
                              │
                         處理失敗？
                              │
                         重新放回 Queue ──▶ 重試
                              │
                         重試 3 次？
                              │
                         移到 DLQ（死信佇列）

實作：Consumer 處理成功後才 ACK
缺點：同一訊息可能被處理多次 → 必須設計冪等性
適合：大部分業務場景（搶票、Email、庫存）
```

### Exactly-once（剛好一次）

```
理論上不可能在分散式系統中完美實現（CAP 定理限制）
實務上的做法：At-least-once + 冪等性消費者

Consumer 端保證冪等：
  INSERT INTO orders (id, ...) ON CONFLICT (id) DO NOTHING
  → 即使重複處理，結果也一樣
```

---

## Dead Letter Queue（死信佇列）

處理失敗且超過重試上限的訊息，不能無限重試，也不能直接丟棄。

```
正常流程：
  Queue ──▶ Consumer ──▶ 成功 ✅

失敗流程：
  Queue ──▶ Consumer ──▶ 失敗 ❌
    ↑                       │
    └── 重試（指數退避）────┘
                            │
                  超過 maxRetry？
                            │
                            ▼
                   [Dead Letter Queue]
                            │
                   人工檢視 / 告警
                   修復後重新投遞
```

### 指數退避（Exponential Backoff）

```go
// 不要固定等待時間，用指數退避避免衝擊
// 重試 1: 等 1 秒
// 重試 2: 等 2 秒
// 重試 3: 等 4 秒
// 重試 4: 等 8 秒
delay := time.Duration(1<<retryCount) * time.Second
```

### 為什麼需要 DLQ？

| 沒有 DLQ | 有 DLQ |
|----------|--------|
| 壞訊息無限重試，阻塞其他訊息 | 壞訊息移走，佇列繼續運作 |
| 無法追蹤失敗原因 | 集中檢視所有失敗訊息 |
| 可能造成雪崩效應 | 系統穩定，人工介入修復 |

---

## Fan-out 模式（一個訊息，多個消費者）

```
                         ┌──▶ [Email Queue]  ──▶ Email Worker
                         │
  Order Created ──▶ MQ ──┼──▶ [Inventory Queue] ──▶ Inventory Worker
                         │
                         └──▶ [Analytics Queue] ──▶ Analytics Worker
```

本課的程式碼示範了 Fan-out：同一個 `orderQueue` 有 `inventoryConsumer` 和 `analyticsConsumer` 兩個消費者。

**注意**：用 Go channel 實作的 Fan-out 有個問題 — channel 只能被一個 goroutine 讀取。本課用**同一個 channel 被多個 goroutine 讀**的方式，實際上是 **competing consumers**（競爭消費），不是真正的 Fan-out。

真正的 Fan-out 需要 Broker 把同一訊息**複製**給每個訂閱者：

```go
// 真正的 Fan-out（像搶票系統的 Broker）
func (b *Broker) Publish(topic string, payload any) {
    for _, ch := range b.subscribers[topic] {
        ch <- msg    // 每個訂閱者都收到一份
    }
}
```

---

## Backpressure 處理（背壓）

當生產者太快、消費者太慢時，佇列會滿，這就是 backpressure。

### 處理策略

| 策略 | 做法 | 適用場景 |
|------|------|---------|
| **丟棄** | 佇列滿了就丟掉新訊息 | 監控指標、日誌 |
| **阻塞** | 生產者等到佇列有空間 | 重要但不緊急的訊息 |
| **回壓** | 通知上游降速 | API Rate Limiting |
| **擴容** | 增加消費者數量 | 自動擴展架構 |
| **緩衝** | 增加佇列容量 | 短暫流量突增 |

```go
// 本課使用「丟棄」策略
func (q *Queue) Publish(msg Message) error {
    select {
    case q.ch <- msg:
        return nil          // 成功
    default:
        return fmt.Errorf("佇列已滿")  // 丟棄
    }
}

// 「阻塞」策略（生產者等待）
func (q *Queue) PublishBlocking(ctx context.Context, msg Message) error {
    select {
    case q.ch <- msg:
        return nil
    case <-ctx.Done():
        return ctx.Err()    // 等太久，超時
    }
}
```

---

## In-Memory MQ vs Redis Streams vs Kafka

| 比較 | Go channel（本課） | Redis Streams | Apache Kafka |
|------|-------------------|---------------|-------------|
| **部署** | 不需要 | 需要 Redis | 需要 Kafka + ZooKeeper |
| **持久化** | 無（程式重啟就沒了） | 有（Redis AOF/RDB） | 有（磁碟日誌） |
| **吞吐量** | 極高（記憶體內） | 高（10 萬+/s） | 超高（百萬+/s） |
| **消費者群組** | 不支援 | 支援（Consumer Group） | 支援（Consumer Group） |
| **訊息重播** | 不支援 | 支援（XRANGE） | 支援（offset reset） |
| **跨程序** | 不支援（只能同一程序） | 支援 | 支援 |
| **順序保證** | 保證（channel FIFO） | 同一 stream 保證 | 同一 partition 保證 |
| **適用場景** | 學習概念、單機簡單場景 | 中小規模、速度優先 | 大數據、事件溯源 |
| **運維複雜度** | 無 | 低 | 高 |

### 選擇建議

```
你的專案是什麼規模？

  小型（單機、< 1000 QPS）
    → Go channel 就夠了

  中型（多台機器、1000-100000 QPS）
    → Redis Streams（簡單、快速、你可能已經有 Redis）

  大型（高吞吐、需要持久化、事件重播）
    → Apache Kafka（複雜但強大）

  微服務通訊（低延遲、輕量）
    → NATS（Go 原生，超低延遲）

  企業級（複雜路由、多協定支援）
    → RabbitMQ（完整的 AMQP 協定）
```

---

## 冪等性（Idempotency）— 最重要的概念

因為 At-least-once 是最常見的保證，消費者**必須**設計成可以安全地重複處理同一訊息。

```go
// ❌ 不冪等：重複處理會造成問題
func processPayment(orderID string, amount float64) error {
    db.Exec("UPDATE accounts SET balance = balance - ?", amount)  // 重複扣款！
    return nil
}

// ✅ 冪等：重複處理也安全
func processPayment(orderID string, amount float64) error {
    // 用 orderID 作為冪等鍵
    result := db.Exec(`
        INSERT INTO payments (order_id, amount, status)
        VALUES (?, ?, 'completed')
        ON CONFLICT (order_id) DO NOTHING
    `, orderID, amount)

    if result.RowsAffected == 0 {
        log.Printf("訂單 %s 已處理過，跳過", orderID)
    }
    return nil
}
```

---

## 與搶票系統的連結

搶票系統的 `ticket-system/internal/mq/broker.go` 就是一個 In-Memory Pub/Sub Broker：

```go
// Broker 訊息代理（用 channel 實作的 Pub/Sub）
type Broker struct {
    mu          sync.RWMutex
    subscribers map[string][]chan Message    // topic → 多個訂閱者
    bufferSize  int
}

// 發布：每個訂閱者都收到一份（Fan-out）
func (b *Broker) Publish(topic string, payload any) {
    for _, ch := range b.subscribers[topic] {
        select {
        case ch <- msg:        // 成功送出
        default:
            slog.Warn("佇列已滿，丟棄")  // Backpressure：丟棄策略
        }
    }
}
```

搶票流程中，MQ 的角色：

```
1. 使用者搶票 → TicketUsecase.PlaceOrder()
2. 扣庫存成功 → broker.Publish("order.created", order)
3. PaymentWorker 收到 → 呼叫 gRPC 支付服務
4. 支付成功 → broker.Publish("order.paid", order)
5. StockBroadcaster 收到 → 透過 WebSocket 推播庫存變更給前端
```

---

## 練習題

### 練習 1：實作 Dead Letter Queue
在本課的 `Consumer` 中：
- 建立一個 `DLQ` 佇列
- 當訊息重試超過 `maxRetry` 次，把它移到 DLQ
- 新增一個 `DLQConsumer`，每 5 秒檢查 DLQ 並印出所有失敗訊息
- 思考：DLQ 中的訊息如何重新投遞？

### 練習 2：實作冪等消費者
修改 `inventoryHandler`：
- 用 `sync.Map` 紀錄已處理的訊息 ID
- 如果同一個訊息 ID 已經處理過，直接跳過
- 故意發送重複的訊息，驗證冪等性

### 練習 3：實作 Pub/Sub Broker
參考搶票系統的 `mq/broker.go`，實作一個真正的 Pub/Sub Broker：
- `Subscribe(topic)` 回傳獨立的 channel
- `Publish(topic, msg)` 把訊息發給所有該 topic 的訂閱者
- 同一訊息每個訂閱者都收到一份（真正的 Fan-out）
- 加上 `Unsubscribe` 功能

### 練習 4：實作 Backpressure 阻塞策略
修改 `Queue.Publish`：
- 新增 `PublishWithTimeout(ctx context.Context, msg Message) error`
- 當佇列滿時，等待直到有空間或 context 超時
- 用一個超級慢的消費者（1 秒處理一個）測試效果

### 練習 5：訊息優先級佇列
實作一個 `PriorityQueue`：
- 訊息有 `Priority` 欄位（1=高、2=中、3=低）
- 高優先級的訊息先被消費
- 提示：可以用多個 channel，消費者先檢查高優先級的 channel
- 思考：為什麼 Kafka 不支援訊息優先級？
