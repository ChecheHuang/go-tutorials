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
- 四大容錯模式之間如何協同運作
- 搶票系統中各容錯元件的對應程式碼

## 執行方式

```bash
go run ./tutorials/40-failure-modeling/
```

---

## 生活比喻：醫院急診室

想像一家醫院的急診流程：

```
╔══════════════════════════════════════════════════════════════════╗
║                                                                ║
║   📋 病歷表 = WAL（Write-Ahead Log）                            ║
║   ├─ 病人一進來就先寫病歷（先記錄，再治療）                      ║
║   ├─ 就算醫生換班，新醫生看病歷就能接手                          ║
║   └─ 沒有病歷 → 換班後沒人知道病人在治什麼                       ║
║                                                                ║
║   🏥 治療流程 = Saga Pattern                                    ║
║   ├─ 掛號 → 看診 → 檢驗 → 開藥 → 領藥                          ║
║   ├─ 如果檢驗發現不需要用藥 → 退掛號費、取消處方（補償）          ║
║   └─ 每一步都有「怎麼退回去」的 SOP                             ║
║                                                                ║
║   🏷️ 看診號碼牌 = Idempotency（冪等性）                         ║
║   ├─ 同一張號碼牌只能看一次診                                   ║
║   ├─ 不會因為叫了兩次號就看兩次、收兩次費                        ║
║   └─ 號碼牌就是 idempotency key                                ║
║                                                                ║
║   📬 院內公文系統 = Outbox Pattern                              ║
║   ├─ 醫生開完處方，同時放進「待通知藥局」信箱                    ║
║   ├─ 行政人員定期檢查信箱，通知藥局配藥                         ║
║   └─ 就算通知系統當掉，處方還在信箱裡不會丟                     ║
║                                                                ║
╚══════════════════════════════════════════════════════════════════╝
```

**為什麼不能只靠一種模式？**

- 只有病歷（WAL）→ 流程出錯不知怎麼善後
- 只有流程 SOP（Saga）→ 系統重啟後不知道做到哪一步
- 只有號碼牌（冪等）→ 不能防止通知丟失
- 四個模式組合起來 → 完整的容錯體系

---

## 真實場景：100 萬人搶 5000 張票

這些問題不是假設，是每個大型售票系統都會遇到的：

| 問題 | 發生時機 | 後果 | 需要哪個模式 |
|------|---------|------|-------------|
| Redis crash | queue 裡 100 萬人排隊中 | 整個排隊名單消失 | WAL |
| Payment 成功但 Order Service crash | 扣款完正要出票 | 錢扣了沒票 | Saga |
| MQ 重送訂單 | 網路抖動導致 ACK 失敗 | 重複扣款 | Idempotency |
| DB 寫入成功但 MQ 發送失敗 | 正要通知下游服務 | 下游不知道有新訂單 | Outbox |
| Saga 補償到一半又 crash | 退款退到一半 | 部分退款、資料不一致 | WAL + Idempotency |

---

## 一、WAL（Write-Ahead Log）

### 核心原則

**任何操作都先寫日誌，操作成功後才標記完成。**

這和資料庫的 WAL 完全一樣 — PostgreSQL 的每一筆寫入都是先寫 WAL 再寫 data page。

### 基本流程

```
正常流程：
  使用者排隊 → 寫入 Redis queue

加了 WAL：
  使用者排隊 → 寫入 WAL（DB）→ 寫入 Redis queue → 標記 WAL committed
```

### Redis Crash 後的恢復流程

```
  ┌─────────────────────────────────────────────────────┐
  │                  Redis Crash 發生                     │
  └──────────────────────┬──────────────────────────────┘
                         ▼
  ┌─────────────────────────────────────────────────────┐
  │  Step 1: 偵測到 Redis 不可用                          │
  │  （Health check 失敗 / Connection refused）           │
  └──────────────────────┬──────────────────────────────┘
                         ▼
  ┌─────────────────────────────────────────────────────┐
  │  Step 2: 啟動新的 Redis 或等待 Redis 恢復             │
  └──────────────────────┬──────────────────────────────┘
                         ▼
  ┌─────────────────────────────────────────────────────┐
  │  Step 3: 掃描 WAL 表                                 │
  │  SELECT * FROM wal_entries                           │
  │  WHERE status = 'pending'                            │
  │  ORDER BY created_at ASC                             │
  └──────────────────────┬──────────────────────────────┘
                         ▼
  ┌─────────────────────────────────────────────────────┐
  │  Step 4: 逐筆重播到新 Redis                           │
  │  FOR EACH entry:                                     │
  │    ├─ LPUSH queue entry.data                         │
  │    ├─ UPDATE wal SET status='committed'              │
  │    └─ 如果重播失敗 → 記錄錯誤，跳過，稍後重試          │
  └──────────────────────┬──────────────────────────────┘
                         ▼
  ┌─────────────────────────────────────────────────────┐
  │  Step 5: 恢復完成                                     │
  │  ├─ 記錄恢復統計（共重播 N 筆、失敗 M 筆）             │
  │  └─ 恢復正常服務                                     │
  └─────────────────────────────────────────────────────┘
```

### WAL 在搶票系統中的程式碼

對應檔案：`wal.go`

```go
// WAL Entry 的狀態機
// pending → committed（正常完成）
// pending → replayed（crash 後重播完成）
// pending → failed（重試多次仍失敗）

type WALEntry struct {
    ID        string
    Data      []byte
    Status    string    // "pending", "committed", "replayed", "failed"
    CreatedAt time.Time
    Retries   int
}

// 每次操作都先寫 WAL
func (w *WAL) WriteAndExecute(data []byte, action func() error) error {
    // Step 1: 先寫 WAL
    entry := w.Write(data)

    // Step 2: 執行實際操作
    if err := action(); err != nil {
        return err
    }

    // Step 3: 標記完成
    w.MarkCommitted(entry.ID)
    return nil
}
```

### WAL 清理策略

WAL 不能無限增長，需要定期清理：

| 策略 | 說明 | 適用場景 |
|------|------|---------|
| 時間清理 | 刪除 7 天前 committed 的記錄 | 大部分場景 |
| 數量清理 | 只保留最新 10 萬筆 | 高流量場景 |
| Checkpoint | 記錄最後成功位置，之前全刪 | 類似 DB 的 checkpoint |

---

## 二、Saga Pattern

### 核心概念

Saga 把一個「大交易」拆成多個「小步驟」，每個步驟都有對應的「補償動作」。如果某一步失敗，就反向執行所有已完成步驟的補償。

### 完整 4 步驟範例：搶票 Saga

```
                        正向流程（Forward）
  ═══════════════════════════════════════════════════════════════

  Step 1           Step 2           Step 3           Step 4
  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
  │ 扣庫存    │───▶│ 建訂單    │───▶│ 扣款     │───▶│ 出票     │
  │          │    │          │    │          │    │          │
  │ stock-1  │    │ order:   │    │ charge   │    │ send     │
  │          │    │ created  │    │ $500     │    │ ticket   │
  └──────────┘    └──────────┘    └──────────┘    └──────────┘
       ✅               ✅              ✅              ❌
                                                   crash!

  ═══════════════════════════════════════════════════════════════
                        補償流程（Compensation）

                  ◀─────────────────────────────────────────
  Step 1'          Step 2'          Step 3'          Step 4'
  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
  │ 回補庫存  │◀───│ 取消訂單  │◀───│ 退款     │◀───│ （跳過）  │
  │          │    │          │    │          │    │          │
  │ stock+1  │    │ order:   │    │ refund   │    │ 未成功    │
  │          │    │ cancelled│    │ $500     │    │ 不需補償  │
  └──────────┘    └──────────┘    └──────────┘    └──────────┘
       ↩️               ↩️              ↩️
```

### 每一步的正向 + 補償對照

| 步驟 | 正向操作 | 補償操作 | 冪等保護 |
|------|---------|---------|---------|
| Step 1 | `stock -= 1` | `stock += 1` | 用 order_id 確認是否已扣過 |
| Step 2 | `INSERT order (status=created)` | `UPDATE order SET status=cancelled` | 檢查 order 是否存在 |
| Step 3 | 呼叫支付 API 扣款 | 呼叫支付 API 退款 | 用 payment_id 冪等退款 |
| Step 4 | 發送電子票券 | 作廢票券 | 檢查票是否已發送 |

### Orchestrator vs Choreography

```
  Orchestrator（指揮者模式）— 本課示範

  ┌─────────────┐
  │  Saga        │    「扣庫存」     ┌──────────┐
  │  Orchestrator│──────────────────▶│ 庫存服務  │
  │              │◀─────── OK ──────│          │
  │              │                   └──────────┘
  │              │    「建訂單」     ┌──────────┐
  │              │──────────────────▶│ 訂單服務  │
  │              │◀─────── OK ──────│          │
  │              │                   └──────────┘
  │              │    「扣款」       ┌──────────┐
  │              │──────────────────▶│ 支付服務  │
  │              │◀─────── FAIL ────│          │
  │              │                   └──────────┘
  │              │
  │   開始補償！  │    「取消訂單」   ┌──────────┐
  │              │──────────────────▶│ 訂單服務  │
  │              │◀─────── OK ──────│          │
  │              │    「回補庫存」   ┌──────────┐
  │              │──────────────────▶│ 庫存服務  │
  │              │◀─────── OK ──────│          │
  └─────────────┘                   └──────────┘


  Choreography（編舞模式）

  ┌──────────┐  event   ┌──────────┐  event   ┌──────────┐
  │ 庫存服務  │────────▶│ 訂單服務  │────────▶│ 支付服務  │
  │          │         │          │         │          │
  │ 監聽：    │         │ 監聽：    │         │ 監聽：    │
  │ order.   │         │ stock.   │         │ order.   │
  │ cancelled│         │ deducted │         │ created  │
  └──────────┘         └──────────┘         └──────────┘
       ▲                                         │
       └─────────── compensation event ──────────┘
```

**怎麼選？**

| 比較 | Orchestrator | Choreography |
|------|-------------|-------------|
| 流程可見性 | 高（集中控制） | 低（分散在各服務） |
| 耦合度 | 中（依賴 orchestrator） | 低（只依賴事件） |
| 除錯難度 | 低（看 orchestrator 日誌） | 高（需追蹤多個服務） |
| 適合場景 | 步驟多、流程複雜 | 步驟少、服務獨立 |
| 搶票系統 | **推薦** | 不建議 |

### Saga 在搶票系統中的程式碼

對應檔案：`saga.go`

```go
type SagaStep struct {
    Name       string
    Forward    func(ctx context.Context) error  // 正向操作
    Compensate func(ctx context.Context) error  // 補償操作
}

type Saga struct {
    Steps     []SagaStep
    Completed []int  // 已完成的步驟 index
}

func (s *Saga) Execute(ctx context.Context) error {
    for i, step := range s.Steps {
        if err := step.Forward(ctx); err != nil {
            // 失敗！開始反向補償
            return s.compensate(ctx, i-1)
        }
        s.Completed = append(s.Completed, i)
    }
    return nil
}

func (s *Saga) compensate(ctx context.Context, fromStep int) error {
    for i := fromStep; i >= 0; i-- {
        if err := s.Steps[i].Compensate(ctx); err != nil {
            // 補償也失敗 → 記錄到 dead letter，人工處理
            log.Printf("compensation failed at step %d: %v", i, err)
        }
    }
    return ErrSagaCompensated
}
```

---

## 三、Idempotency（冪等性）

### 核心概念

**同一個操作執行 1 次和執行 N 次，結果完全一樣。**

```
第 1 次收到 payment:order-123 → 執行扣款 ✅
第 2 次收到 payment:order-123 → 檢查 key 已存在 → 跳過 ✅
第 3 次收到 payment:order-123 → 檢查 key 已存在 → 跳過 ✅

結果：不管重送幾次，只扣一次款
```

### Idempotency Key 的生成策略

| 生成方式 | 格式 | 適用場景 | 範例 |
|---------|------|---------|------|
| 業務欄位組合 | `{service}:{entity}:{id}` | 訂單處理 | `payment:order:order-123` |
| 客戶端生成 UUID | `{client-uuid}` | API 請求 | `550e8400-e29b-41d4-a716-446655440000` |
| 內容 Hash | `sha256({body})` | 訊息去重 | `sha256("{"order_id":"123","amount":500}")` |
| 時間窗口 | `{user}:{action}:{minute}` | 防重複提交 | `user-456:submit:202601061430` |

**最佳實踐**：搶票系統使用「業務欄位組合」，因為 order_id 天然唯一。

### Idempotency Key 的儲存方案

```
方案一：Redis（推薦，高效能）

  SET idempotency:payment:order-123 "{result_json}" EX 86400 NX
  │                                  │                │       │
  │                                  │                │       └─ 只在 key 不存在時設定
  │                                  │                └─ TTL: 24 小時
  │                                  └─ 儲存上次的執行結果
  └─ key 格式

方案二：DB（推薦，持久化）

  CREATE TABLE idempotency_keys (
      key         VARCHAR(255) PRIMARY KEY,
      result      JSONB,
      created_at  TIMESTAMP DEFAULT NOW(),
      expires_at  TIMESTAMP
  );

  -- 利用 unique constraint 防止重複
  INSERT INTO idempotency_keys (key, result, expires_at)
  VALUES ('payment:order-123', '{"status":"ok"}', NOW() + INTERVAL '24 hours')
  ON CONFLICT (key) DO NOTHING;
```

### TTL 策略

| 場景 | TTL | 原因 |
|------|-----|------|
| 支付扣款 | 24 小時 | MQ 重送通常在幾分鐘內 |
| 票券發送 | 7 天 | 可能有延遲重送 |
| API 請求去重 | 1 小時 | 使用者不太可能 1 小時後重試同樣請求 |
| Saga 補償 | 30 天 | 補償可能很晚才觸發 |

### 冪等性檢查流程

```
  收到請求（key = "payment:order-123"）
       │
       ▼
  ┌──────────────────────┐
  │ 查詢 idempotency key │
  │ GET key / SELECT key │
  └──────────┬───────────┘
             │
       ┌─────┴─────┐
       │           │
    key 存在    key 不存在
       │           │
       ▼           ▼
  ┌─────────┐  ┌─────────────────────┐
  │ 直接返回 │  │ SET key NX + 執行操作 │
  │ 上次結果 │  │                     │
  └─────────┘  └──────────┬──────────┘
                          │
                    ┌─────┴─────┐
                    │           │
                 SET 成功    SET 失敗
                 （我先搶到）  （別人先搶到）
                    │           │
                    ▼           ▼
               ┌─────────┐  ┌─────────┐
               │ 執行操作 │  │ 等待結果 │
               │ 儲存結果 │  │ 或重試   │
               └─────────┘  └─────────┘
```

### 在搶票系統中的程式碼

對應檔案：`idempotency/store.go`

```go
type IdempotencyStore struct {
    redis *redis.Client
}

func (s *IdempotencyStore) CheckAndSet(ctx context.Context, key string, ttl time.Duration) (bool, error) {
    // SET NX: 只在 key 不存在時設定
    ok, err := s.redis.SetNX(ctx, "idempotency:"+key, "processing", ttl).Result()
    if err != nil {
        return false, err
    }
    return ok, nil // true = 第一次，false = 重複
}

func (s *IdempotencyStore) SaveResult(ctx context.Context, key string, result []byte, ttl time.Duration) error {
    return s.redis.Set(ctx, "idempotency:"+key, result, ttl).Err()
}
```

---

## 四、Outbox Pattern

### 核心問題

DB 和 MQ 是兩個不同的系統，沒辦法放在同一個 transaction 裡：

```
不用 Outbox（有風險）：

  BEGIN TRANSACTION
    INSERT INTO orders ...    ← ✅ DB 成功
  COMMIT
  publish("order.created")    ← ❌ MQ 掛了，訊息丟失！

  結果：DB 有訂單，但下游服務永遠不知道
```

### Outbox 解決方案

```
  BEGIN TRANSACTION
    INSERT INTO orders (id, user_id, amount)
      VALUES ('order-123', 'user-456', 500);

    INSERT INTO outbox (id, aggregate_type, aggregate_id, event_type, payload)
      VALUES ('evt-001', 'order', 'order-123', 'order.created',
              '{"order_id":"order-123","amount":500}');
  COMMIT                        ← 同一個 DB transaction，要嘛都成功，要嘛都失敗

  ┌──── 背景程式（Outbox Relay）每 N 秒 ────┐
  │                                          │
  │  SELECT * FROM outbox                    │
  │  WHERE published = false                 │
  │  ORDER BY created_at ASC                 │
  │  LIMIT 100                               │
  │          │                               │
  │          ▼                               │
  │  FOR EACH event:                         │
  │    ├─ publish(event) → MQ                │
  │    └─ UPDATE outbox SET published = true │
  │                                          │
  └──────────────────────────────────────────┘
```

### DB + MQ 一致性保證

```
  ┌───────────────────────────────────────────────────────────┐
  │                   Application Server                      │
  │                                                           │
  │   ┌─────────────┐          ┌─────────────────────┐        │
  │   │  Business    │          │  Outbox Relay        │       │
  │   │  Logic       │          │  (Background Worker) │       │
  │   └──────┬──────┘          └──────────┬──────────┘        │
  │          │                            │                   │
  └──────────┼────────────────────────────┼───────────────────┘
             │                            │
             │ 1. 同一個 TX               │ 3. 定期掃描
             │    寫 orders + outbox      │    未發布的事件
             ▼                            ▼
  ┌─────────────────────┐      ┌───────────────────┐
  │     PostgreSQL       │      │     PostgreSQL     │
  │  ┌───────┬────────┐ │      │  ┌──────────────┐  │
  │  │orders │ outbox │ │      │  │   outbox      │  │
  │  │ table │ table  │ │      │  │ published=f   │  │
  │  └───────┴────────┘ │      │  └──────┬───────┘  │
  └─────────────────────┘      └─────────┼─────────┘
                                         │
                                         │ 4. 發布事件
                                         ▼
                                ┌───────────────────┐
                                │   Message Queue    │
                                │   (RabbitMQ/Kafka) │
                                └───────────────────┘
                                         │
                                         │ 5. 下游消費
                                         ▼
                                ┌───────────────────┐
                                │  下游服務           │
                                │  （支付、通知...）   │
                                └───────────────────┘
```

### Outbox 表結構

```sql
CREATE TABLE outbox (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(50) NOT NULL,    -- 'order', 'payment'
    aggregate_id   VARCHAR(100) NOT NULL,   -- 'order-123'
    event_type     VARCHAR(100) NOT NULL,   -- 'order.created'
    payload        JSONB NOT NULL,          -- 事件內容
    published      BOOLEAN DEFAULT false,
    created_at     TIMESTAMP DEFAULT NOW(),
    published_at   TIMESTAMP
);

-- 加 index 加速 Relay 查詢
CREATE INDEX idx_outbox_unpublished ON outbox (created_at)
    WHERE published = false;
```

### Outbox 的注意事項

| 問題 | 解法 |
|------|------|
| Relay 掛了怎麼辦？ | 重啟後從 `published=false` 繼續，不會丟 |
| 事件重複發布？ | 下游用 idempotency key 去重 |
| Outbox 表太大？ | 定期刪除 `published=true` 的舊記錄 |
| 順序重要嗎？ | 用 `created_at` 排序保證順序 |

---

## 五、容錯模式比較

### 每個模式解決什麼問題？

| 問題場景 | WAL | Saga | Idempotency | Outbox |
|---------|-----|------|-------------|--------|
| Redis crash 後恢復資料 | ✅ | - | - | - |
| 多步驟操作失敗要回滾 | - | ✅ | - | - |
| MQ 重送導致重複執行 | - | - | ✅ | - |
| DB 和 MQ 資料不一致 | - | - | - | ✅ |
| Crash 後不知道做到哪步 | ✅ | ✅ | - | - |
| 補償操作被重複執行 | - | - | ✅ | - |
| 分散式交易一致性 | - | ✅ | - | ✅ |

### 模式之間的依賴關係

```
  ┌─────────────────────────────────────────────────────┐
  │                                                     │
  │   WAL 保護 Saga 的狀態                               │
  │   ┌─────┐         ┌──────┐                          │
  │   │ WAL │────────▶│ Saga │   Saga 的每一步           │
  │   └─────┘         └──┬───┘   都需要冪等保護           │
  │                      │       ┌─────────────┐        │
  │                      └──────▶│ Idempotency │        │
  │                              └─────────────┘        │
  │                                    ▲                │
  │   Outbox 發出的事件                 │ 下游收到事件     │
  │   ┌────────┐                       │ 需要冪等去重     │
  │   │ Outbox │───────────────────────┘                │
  │   └────────┘                                        │
  │                                                     │
  └─────────────────────────────────────────────────────┘
```

---

## 六、四大模式協同運作

在搶票系統中，這四個模式不是各自獨立的，而是一起保護整個流程：

```
使用者搶票
    │
    ▼
┌───────────────────────────────────────────────────────────────┐
│ Step 1: 排隊入 queue                                          │
│                                                               │
│   WAL: 先寫 WAL → 再 LPUSH Redis → 標記 committed             │
│   Idempotency: 同一 user 同一場次只能排一次                     │
└──────────────────────────┬────────────────────────────────────┘
                           ▼
┌───────────────────────────────────────────────────────────────┐
│ Step 2: 排到你了，開始購票 Saga                                │
│                                                               │
│   Saga Orchestrator 執行：                                     │
│     1) 扣庫存（idempotent: order_id）                          │
│     2) 建訂單（idempotent: order_id）                          │
│     3) 扣款  （idempotent: payment_id）                        │
│     4) 出票  （idempotent: ticket_id）                         │
│                                                               │
│   WAL: Saga 每完成一步都記錄到 WAL                              │
│   Idempotency: Saga 每一步都有冪等保護                          │
└──────────────────────────┬────────────────────────────────────┘
                           ▼
┌───────────────────────────────────────────────────────────────┐
│ Step 3: 通知下游（座位確認、寄 email）                          │
│                                                               │
│   Outbox: 訂單寫入 DB 的同時寫入 outbox                        │
│   Relay: 背景程式把 outbox 事件發布到 MQ                        │
│   Idempotency: 下游服務用 event_id 去重                        │
└───────────────────────────────────────────────────────────────┘
```

### 搶票系統程式碼對照

| 模式 | 檔案 | 關鍵函式 |
|------|------|---------|
| WAL | `wal.go` | `WAL.Write()`, `WAL.Recover()` |
| Saga | `saga.go` | `Saga.Execute()`, `Saga.Compensate()` |
| Idempotency | `idempotency/store.go` | `Store.CheckAndSet()`, `Store.SaveResult()` |
| Outbox | `outbox.go` (如有) | `Outbox.Write()`, `OutboxRelay.Run()` |

---

## 面試題對照

| 面試問題 | 考的模式 | 本課位置 |
|---------|---------|---------|
| Redis 掛了 queue 怎麼恢復？ | WAL | Demo 1 |
| 付款成功但出票失敗怎麼辦？ | Saga | Demo 2 |
| MQ 重送怎麼避免重複處理？ | Idempotency | Demo 3 |
| 怎麼確保 DB 和 MQ 一致？ | Outbox | Demo 4 |
| 這四個模式怎麼一起用？ | 全部 | 第六節 |

---

## FAQ

### Q1: WAL 寫入 DB 不會成為效能瓶頸嗎？

**不會，因為 WAL 是 append-only 的順序寫入。**

順序寫入是資料庫最快的操作之一（比隨機寫入快 10-100 倍）。而且 WAL 只記錄最少的必要資訊（操作類型 + payload），不做複雜的索引更新。如果真的擔心效能，可以用批次寫入（batch insert）每 100ms 批量寫一次。

### Q2: Saga 的補償也失敗了怎麼辦？

**進入 Dead Letter Queue，人工介入。**

這是分散式系統中的現實：不存在 100% 自動化的容錯。做法是：
1. 補償失敗 → 重試 3 次（加 exponential backoff）
2. 仍然失敗 → 寫入 dead letter table
3. 告警通知 on-call 工程師
4. 工程師用管理後台手動處理

實務上，補償失敗率通常低於 0.01%。

### Q3: Idempotency key 過期後重送怎麼辦？

**這代表 TTL 設太短。**

TTL 應該設為「最長可能重送間隔」的 2-3 倍。例如 MQ 的最大重試時間是 1 小時，TTL 就設 3 小時。如果是支付場景，建議直接用 DB 永久存儲（定期歸檔而非刪除）。

### Q4: Outbox 和直接用 MQ 的 transaction 功能有什麼差別？

**Outbox 不依賴 MQ 的 transaction 功能。**

有些 MQ（如 RabbitMQ）的 transaction 效能很差，或功能有限。Outbox 的優勢是：
- 只依賴 DB 的 transaction（成熟、可靠）
- 換 MQ 不需要改邏輯
- 事件天然有持久化（DB 裡的 outbox 表）

### Q5: 這些模式會讓程式碼變很複雜，值得嗎？

**看流量和業務重要性。**

| 場景 | 需要嗎？ |
|------|---------|
| 內部管理系統（10 人用） | 不需要，KISS 原則 |
| 電商網站（日訂單 1000） | Idempotency + 基本重試就好 |
| 搶票系統（100 萬人搶） | **全部都需要** |
| 金融系統（不能錯一毛） | **全部都需要 + 更嚴格的對帳** |

---

## 練習

1. 實作一個簡單的 WAL：先寫日誌再執行操作，程式重啟後重播未完成的日誌
2. 設計一個 3 步驟的 Saga：扣庫存 → 建訂單 → 扣款，並為每步寫補償函式
3. 實作冪等性檢查：用 idempotency key 確保同一個請求只處理一次
4. 實作 Outbox Pattern：寫入 DB 時同時寫入 outbox 表，背景程式讀取 outbox 發送 MQ
5. 模擬 Redis crash：停掉 Redis 後觀察 WAL 如何恢復排隊資料

---

## 下一課預告

**第四十一課：高可用架構** — Redis Sentinel/Cluster、Multi-Queue Failover、腦裂問題。
