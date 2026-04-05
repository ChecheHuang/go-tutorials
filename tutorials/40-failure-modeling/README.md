# 第四十課：分散式系統容錯（Failure Modeling）

> **一句話總結**：資深工程師不是想「系統怎麼跑」，而是先想「系統怎麼壞」。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🔴 資深工程師 | **必備**：理解 Saga、冪等性、WAL 概念 |
| ⚫ Staff / Principal | **核心能力**：設計 failure scenario 和補償機制 |

## 你會學到什麼？

- WAL（Write-Ahead Log）：Redis crash 後如何恢復 queue
- Saga Pattern：分散式交易的自動補償
- Idempotency：MQ 重送時如何避免重複扣款
- Outbox Pattern：確保 DB 和 MQ 的一致性

## 執行方式

```bash
go run ./tutorials/40-failure-modeling/
```

## 真實場景：100 萬人搶 5000 張票

這些問題不是假設，是每個大型售票系統都會遇到的：

| 問題 | 發生時機 | 後果 |
|------|---------|------|
| Redis crash | queue 裡 100 萬人排隊中 | 整個排隊名單消失 |
| Payment 成功但 Order Service crash | 扣款完正要出票 | 錢扣了沒票 |
| MQ 重送訂單 | 網路抖動導致 ACK 失敗 | 重複扣款 |
| DB 寫入成功但 MQ 發送失敗 | 正要通知下游服務 | 下游不知道有新訂單 |

## 四大容錯模式

### 1. WAL（Write-Ahead Log）

```
正常流程：
  使用者排隊 → 寫入 Redis queue

加了 WAL：
  使用者排隊 → 寫入 WAL（DB）→ 寫入 Redis queue → 標記 WAL committed

Redis crash 後：
  讀取 WAL 中 status=pending 的記錄 → 重播到新的 Redis
```

**關鍵原則**：任何操作都先寫日誌，操作成功後才標記完成。

### 2. Saga Pattern

```
正向流程：
  扣庫存 → 建訂單 → 支付 → 出票
    ✅       ✅      ✅     ❌ crash!

補償流程（反向）：
  出票(跳過) ← 退款 ← 取消訂單 ← 回補庫存
               ↩️       ↩️          ↩️

結果：使用者不會「錢扣了沒票」
```

兩種實作方式：
- **Choreography（編舞）**：每個服務自己監聽事件，自己決定要不要補償
- **Orchestrator（指揮）**：有一個中央協調器控制整個流程（本課示範）

### 3. Idempotency（冪等性）

```
第 1 次收到 payment:order-123 → 執行扣款 ✅
第 2 次收到 payment:order-123 → 檢查 key 已存在 → 跳過 ✅
第 3 次收到 payment:order-123 → 檢查 key 已存在 → 跳過 ✅

結果：不管重送幾次，只扣一次款
```

實作方式：
- **Redis SET NX**：`SET idempotency:{key} 1 EX 86400 NX`
- **DB unique constraint**：`INSERT INTO processed_events (key) VALUES (?)`

### 4. Outbox Pattern

```
不用 Outbox（有風險）：
  BEGIN TRANSACTION
    INSERT INTO orders ...    ← ✅ DB 成功
  COMMIT
  publish("order.created")    ← ❌ MQ 掛了，訊息丟失！

用 Outbox：
  BEGIN TRANSACTION
    INSERT INTO orders ...     ← ✅
    INSERT INTO outbox ...     ← ✅ 同一個 transaction
  COMMIT

  背景程式每 5 秒：
    SELECT * FROM outbox WHERE published = false
    → 發送到 MQ
    → UPDATE outbox SET published = true
```

## 面試題對照

| 面試問題 | 考的模式 | 本課位置 |
|---------|---------|---------|
| Redis 掛了 queue 怎麼恢復？ | WAL | Demo 1 |
| 付款成功但出票失敗怎麼辦？ | Saga | Demo 2 |
| MQ 重送怎麼避免重複處理？ | Idempotency | Demo 3 |
| 怎麼確保 DB 和 MQ 一致？ | Outbox | Demo 4 |

## 下一課預告

**第三十九課：高可用架構** — Redis Sentinel/Cluster、Multi-Queue Failover、腦裂問題。
