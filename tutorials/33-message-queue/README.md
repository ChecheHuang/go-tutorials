# 第三十三課：Message Queue（訊息佇列）

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | 了解 Pub/Sub 概念，能實作基本的生產者/消費者 |
| 🔴 資深工程師 | **必備**：能設計可靠的訊息處理系統，包含 Dead Letter Queue、冪等性 |
| 🏢 架構師 | 評估 Redis/Kafka/RabbitMQ/NATS 的取捨 |

## 核心概念

```
Producer → [Message Queue] → Consumer
              ↓ 失敗時
         [Dead Letter Queue]
```

## 關鍵設計考量

### 冪等性（Idempotency）
消費者可能重複收到同一條訊息（At-least-once），必須設計成「安全可重複執行」：

```go
// ❌ 不冪等
func processOrder(orderID string) {
    db.Exec("INSERT INTO orders ...")  // 重複執行會插入兩筆
}

// ✅ 冪等
func processOrder(orderID string) {
    db.Exec("INSERT INTO orders ... ON CONFLICT DO NOTHING")
}
```

### 死信佇列（Dead Letter Queue）
處理失敗的訊息，避免無限重試阻塞佇列：

```go
if msg.RetryCount >= maxRetry {
    dlq.Publish(msg)  // 移到 DLQ，人工介入
    return
}
```

### 指數退避（Exponential Backoff）
```go
delay := time.Duration(retryCount * retryCount) * time.Second
time.Sleep(delay)  // 1s, 4s, 9s, 16s...
```

## 真實世界的選擇

| 技術 | 吞吐量 | 持久化 | 適用場景 |
|------|--------|--------|----------|
| Go channel | 最高（記憶體） | ❌ | 進程內通訊 |
| Redis Streams | 高 | ✅ | 中小規模 |
| RabbitMQ | 中 | ✅ | 複雜路由 |
| Apache Kafka | 超高 | ✅ | 大數據流 |
| NATS | 極高 | 可選 | 微服務 |

## 執行方式

```bash
go run ./tutorials/33-message-queue
```
