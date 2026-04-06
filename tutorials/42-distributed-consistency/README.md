# 第四十二課：分散式一致性（Distributed Consistency）

> **一句話總結**：分散式系統不追求「絕對一致」，而是在一致性、可用性、效能之間做出正確的取捨。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🔴 資深工程師 | **必備**：理解 CAP、最終一致性、分散式鎖 |
| ⚫ Staff / Principal | **核心能力**：設計 inventory token、事件對帳、選擇一致性策略 |

## 你會學到什麼？

- CAP 定理：為什麼不能同時要一致性、可用性、分區容錯
- Inventory Token：先發 token 再搶票，從源頭控制流量
- 分散式鎖：正確使用 `SET key owner EX ttl NX`，自動續約與 fencing token
- 樂觀鎖 vs 悲觀鎖：什麼場景用哪種
- 2PC vs Saga：兩種分散式交易方案的取捨
- 最終一致性：事件對帳 + 自動修復
- 超賣問題：為什麼需要原子操作
- 對帳（Reconciliation）：如何偵測和修復不一致

## 執行方式

```bash
go run ./tutorials/42-distributed-consistency/
```

---

## 生活比喻：超商限量商品促銷

想像全台 5000 家超商同時賣限量 5000 個公仔：

```
╔══════════════════════════════════════════════════════════════════╗
║                                                                ║
║   📊 CAP 定理 = 總部和門市的溝通                                ║
║   ├─ 一致性（C）：每家門市的庫存數字都和總部一樣                  ║
║   ├─ 可用性（A）：客人來了一定能結帳（不能說系統維護中）           ║
║   ├─ 分區容錯（P）：總部和門市斷網了系統還能運作                  ║
║   └─ 三者不能全拿 → 斷網時選「庫存準確」或「能繼續賣」            ║
║                                                                ║
║   🎫 Inventory Token = 限量號碼牌                               ║
║   ├─ 總部只發 5000 張號碼牌給門市                                ║
║   ├─ 沒號碼牌 → 直接跟客人說賣完了（不用查庫存）                  ║
║   └─ 從源頭控制，根本不會超賣                                   ║
║                                                                ║
║   🔒 分散式鎖 = 保險箱裡的最後一個公仔                          ║
║   ├─ 誰先拿到鑰匙誰才能打開保險箱                               ║
║   ├─ 拿鑰匙有時間限制（10 分鐘沒結帳就還鑰匙）                   ║
║   └─ 防止兩個人同時拿到最後一個公仔                              ║
║                                                                ║
║   📋 對帳 = 每天打烊後盤點                                     ║
║   ├─ 比對「賣了幾個」和「收了多少錢」                            ║
║   ├─ 有差異 → 找原因、修正                                     ║
║   └─ 最終一致性的保障                                          ║
║                                                                ║
╚══════════════════════════════════════════════════════════════════╝
```

---

## 一、CAP 定理實戰

### CAP 三角形

```
                      C（Consistency）
                     一致性
                    ╱      ╲
                   ╱        ╲
                  ╱   在網路   ╲
                 ╱   分區時     ╲
                ╱    你只能      ╲
               ╱     選 C 或 A   ╲
              ╱                    ╲
             ╱                      ╲
            A ────────────────────── P
      （Availability）        （Partition
         可用性               Tolerance）
                             分區容錯
```

**重要澄清**：CAP 定理說的不是「永遠只能選兩個」，而是**當網路分區發生時**，你必須在 C 和 A 之間做選擇。正常運作時三者可以同時滿足。

### 真實系統的 CAP 選擇

| 系統 | 選擇 | 犧牲了什麼 | 具體表現 |
|------|------|-----------|---------|
| PostgreSQL（單機） | **CP** | 可用性 | 主機掛了，寫入不可用，直到 failover 完成 |
| PostgreSQL（同步複製） | **CP** | 可用性 | 備機掛了，主機也不能寫（等備機回來） |
| DynamoDB | **AP** | 一致性 | 跨區域複製有延遲，讀到舊資料 |
| Cassandra | **AP** | 一致性 | 多節點寫入可能衝突，用 last-write-wins |
| etcd / ZooKeeper | **CP** | 可用性 | 超過半數節點掛了就拒絕服務 |
| DNS | **AP** | 一致性 | TTL 內拿到舊記錄，不影響解析 |
| 搶票系統 | **AP** | 一致性 | 寧可庫存暫時不準，也不能讓系統掛掉 |
| 銀行轉帳 | **CP** | 可用性 | 寧可暫時不能用，也不能帳對不上 |

### 搶票系統為什麼選 AP？

```
  場景：Redis 主從之間發生網路分區

  選 CP：
    ┌──────────┐    X    ┌──────────┐
    │ Redis 主  │────X────│ Redis 從  │
    └──────────┘    X    └──────────┘
         │
         ▼
    拒絕所有寫入，等待網路恢復
    → 100 萬人全部看到「系統維護中」
    → 搶票活動泡湯 💀

  選 AP：
    ┌──────────┐    X    ┌──────────┐
    │ Redis 主  │────X────│ Redis 從  │
    └──────────┘    X    └──────────┘
         │                     │
         ▼                     ▼
    繼續接受請求            繼續接受請求
    庫存可能暫時不準         靠事件對帳修復
    → 使用者體驗正常
    → 0.1% 的訂單需要事後補償 ✅
```

---

## 二、Inventory Token 模式

### 核心思想

**不讓 100 萬人打 Redis，先從源頭過濾到 5000 人。**

### 完整流程

```
  100 萬人同時請求
         │
         ▼
  ┌──────────────────────────────────────┐
  │        Token Bucket Service           │
  │  （只有 5000 個 token）                │
  │                                      │
  │  atomic counter: remaining = 5000    │
  │                                      │
  │  if atomic.Add(&remaining, -1) >= 0  │
  │    → 發 token                        │
  │  else                                │
  │    → 回傳「售罄」                     │
  └─────────────┬────────────────────────┘
                │
       ┌────────┴─────────┐
       │                  │
   有 token            沒 token
   (5,000 人)         (995,000 人)
       │                  │
       ▼                  ▼
  進入選座頁面         直接回傳「售罄」
       │              （不打 Redis）
       ▼
  ┌──────────────────────────────────────┐
  │          Token 驗證 + 座位鎖定        │
  │                                      │
  │  1. 驗證 token 有效性（簽名 + 過期）   │
  │  2. SET seat:{id} {user} EX 600 NX   │
  │  3. token 標記已使用                   │
  └─────────────┬────────────────────────┘
                │
       ┌────────┴─────────┐
       │                  │
    鎖定成功           鎖定失敗
       │              （座位被搶）
       ▼                  │
  進入付款流程            ▼
       │              選其他座位
       ▼              或放棄
  ┌──────────────────────────────────────┐
  │          付款 + 出票                   │
  │                                      │
  │  Saga: 扣庫存 → 扣款 → 出票           │
  │  成功 → 座位永久鎖定                   │
  │  失敗 → 座位釋放 + 退款               │
  └──────────────────────────────────────┘
```

### Token 的設計

```go
type InventoryToken struct {
    TokenID   string    // UUID
    EventID   string    // 場次 ID
    UserID    string    // 使用者 ID
    IssuedAt  time.Time // 發放時間
    ExpiresAt time.Time // 過期時間（通常 10-15 分鐘）
    Signature string    // HMAC 簽名防偽造
}
```

| 設計考量 | 決策 | 原因 |
|---------|------|------|
| Token 數量 | 庫存的 1.2 倍（6000 張） | 預留部分人放棄購買 |
| 有效期 | 10-15 分鐘 | 太短 → 來不及付款；太長 → 佔著不買 |
| 驗證方式 | HMAC 簽名 | 防止偽造，不需要查 DB |
| 使用次數 | 一次性 | 用 Redis SET NX 標記已使用 |

### 為什麼不直接讓 100 萬人選座位？

| 方案 | Redis 操作次數 | Redis 是否撐得住 |
|------|--------------|-----------------|
| 直接搶座 | 100 萬次 `SET seat:xxx NX` | 很可能被打爆 |
| Inventory Token | 5,000 次 `SET seat:xxx NX` | 輕鬆承受 |

---

## 三、分散式鎖深入

### 基本用法：Redis SET NX EX

```go
// 嘗試取得鎖
func (l *DistLock) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
    // SET key owner EX ttl NX
    // owner = 唯一識別碼（用來確認是自己加的鎖）
    owner := uuid.New().String()
    ok, err := l.redis.SetNX(ctx, "lock:"+key, owner, ttl).Result()
    if err != nil {
        return false, err
    }
    if ok {
        l.owner = owner
    }
    return ok, nil
}

// 釋放鎖（只能釋放自己加的鎖）
// 必須用 Lua script 保證原子性
func (l *DistLock) Unlock(ctx context.Context, key string) error {
    script := `
        if redis.call("GET", KEYS[1]) == ARGV[1] then
            return redis.call("DEL", KEYS[1])
        else
            return 0
        end
    `
    return l.redis.Eval(ctx, script, []string{"lock:" + key}, l.owner).Err()
}
```

### 自動續約（Auto-renewal）

鎖有 TTL，但操作可能比 TTL 久。如果鎖過期了但操作還沒完成，別人就能拿到鎖 — 資料會亂。

```
  不續約的問題：

  時間軸 ─────────────────────────────────────────▶

  Client A:  ├── 取得鎖 ──── TTL 到期 ────── 還在操作！─── 寫入 DB ──┤
  Client B:                    ├── 取得鎖（A 過期了）─── 寫入 DB ──┤

  結果：A 和 B 同時操作 → 資料不一致 💀


  有續約：

  時間軸 ─────────────────────────────────────────▶

  Client A:  ├── 取得鎖 ── 續約 ── 續約 ── 完成 ── 釋放鎖 ──┤
                              ↑       ↑
                           每 TTL/3 續約一次
  Client B:                                        ├── 取得鎖 ──┤

  結果：A 完成後 B 才開始 ✅
```

```go
// 自動續約 goroutine
func (l *DistLock) startRenewal(ctx context.Context, key string, ttl time.Duration) {
    ticker := time.NewTicker(ttl / 3) // 每 TTL 的 1/3 續約一次
    go func() {
        defer ticker.Stop()
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                // 只續約自己的鎖
                script := `
                    if redis.call("GET", KEYS[1]) == ARGV[1] then
                        return redis.call("PEXPIRE", KEYS[1], ARGV[2])
                    else
                        return 0
                    end
                `
                l.redis.Eval(ctx, script, []string{"lock:" + key},
                    l.owner, ttl.Milliseconds())
            }
        }
    }()
}
```

### Fencing Token（防護令牌）

即使有續約，在極端情況下（GC pause、網路延遲）仍可能出現鎖失效的問題。Fencing token 是最後一道防線：

```
  ┌──────────┐     取得鎖        ┌──────────┐
  │ Client A │──────────────────▶│  Redis    │
  │          │◀── token = 33 ───│          │
  └────┬─────┘                   └──────────┘
       │
       │ GC pause 30 秒...鎖過期了
       │
  ┌──────────┐     取得鎖        ┌──────────┐
  │ Client B │──────────────────▶│  Redis    │
  │          │◀── token = 34 ───│          │
  └────┬─────┘                   └──────────┘
       │
       │ 寫入 DB (token=34)       ┌──────────┐
       └─────────────────────────▶│    DB     │
                                  │ 記錄:     │
  Client A GC 結束，              │ token=34  │
  嘗試寫入 DB (token=33)          └────┬─────┘
       │                               │
       └──────── token 33 < 34 ────────┘
                 → 拒絕寫入！✅

  DB 只接受 token >= 目前最大 token 的寫入
```

```go
// Fencing token 通常是一個自增的計數器
func (l *DistLock) LockWithFencing(ctx context.Context, key string, ttl time.Duration) (int64, error) {
    // 取得鎖的同時，取得一個遞增的 fencing token
    script := `
        if redis.call("SET", KEYS[1], ARGV[1], "EX", ARGV[2], "NX") then
            return redis.call("INCR", KEYS[2])
        else
            return 0
        end
    `
    token, err := l.redis.Eval(ctx, script,
        []string{"lock:" + key, "fencing:" + key},
        l.owner, int(ttl.Seconds())).Int64()
    return token, err
}
```

---

## 四、樂觀鎖 vs 悲觀鎖

### 悲觀鎖（Pessimistic Locking）

「我覺得一定會有衝突，所以先鎖起來再操作。」

```go
// 悲觀鎖：SELECT ... FOR UPDATE
func (r *SeatRepo) LockSeatPessimistic(ctx context.Context, tx *sql.Tx, seatID string) (*Seat, error) {
    row := tx.QueryRowContext(ctx, `
        SELECT id, event_id, status, locked_by, version
        FROM seats
        WHERE id = $1
        FOR UPDATE              -- 鎖定這一列，其他 transaction 要等
    `, seatID)

    seat := &Seat{}
    if err := row.Scan(&seat.ID, &seat.EventID, &seat.Status, &seat.LockedBy, &seat.Version); err != nil {
        return nil, err
    }
    return seat, nil
}

// 其他 transaction 嘗試鎖同一列 → 被阻塞，直到第一個 transaction commit
```

### 樂觀鎖（Optimistic Locking）

「我覺得通常不會衝突，如果衝突了就重試。」

```go
// 樂觀鎖：用 version 欄位
func (r *SeatRepo) LockSeatOptimistic(ctx context.Context, seatID string, userID string) error {
    // Step 1: 讀取當前 version
    seat, err := r.GetSeat(ctx, seatID)
    if err != nil {
        return err
    }

    // Step 2: 帶著 version 去更新
    result, err := r.db.ExecContext(ctx, `
        UPDATE seats
        SET status = 'locked', locked_by = $1, version = version + 1
        WHERE id = $2
        AND version = $3       -- 只有 version 沒變才能更新
        AND status = 'available'
    `, userID, seatID, seat.Version)
    if err != nil {
        return err
    }

    // Step 3: 檢查是否真的更新了
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return ErrConflict // version 變了 = 被別人搶先 → 重試或放棄
    }
    return nil
}
```

### 比較表

| 比較 | 悲觀鎖 | 樂觀鎖 |
|------|--------|--------|
| 策略 | 先鎖再操作 | 先操作，提交時檢查 |
| 實作 | `SELECT FOR UPDATE` | `WHERE version = ?` |
| 衝突頻率高時 | 效能好（減少重試） | 效能差（大量重試） |
| 衝突頻率低時 | 效能差（鎖的開銷） | 效能好（無鎖） |
| 死鎖風險 | 有（需小心順序） | 無 |
| 適用場景 | 搶票座位鎖定 | 購物車更新 |
| 搶票系統中 | 座位最終確認 | 庫存預扣 |

```
  衝突率低（購物車）：
    樂觀鎖 ✅  ──── 幾乎不重試，效能高
    悲觀鎖 ❌  ──── 每次都加鎖，浪費

  衝突率高（搶座位）：
    樂觀鎖 ❌  ──── 不斷重試，CPU 浪費
    悲觀鎖 ✅  ──── 排隊等鎖，有序處理
```

---

## 五、超賣問題

```go
// ❌ 不安全（race condition）
if stock > 0 {
    stock--  // 多個 goroutine 同時通過 >0 檢查
}

// ✅ 安全（原子操作，等同 Redis DECRBY）
if atomic.AddInt32(&stock, -1) >= 0 {
    // 成功扣庫存
} else {
    atomic.AddInt32(&stock, 1) // 回滾
}
```

### Redis 原子扣庫存（Lua script）

```lua
-- KEYS[1] = stock:{event_id}
-- ARGV[1] = 要扣的數量
local stock = tonumber(redis.call('GET', KEYS[1]))
if stock == nil then
    return -1  -- key 不存在
end
if stock >= tonumber(ARGV[1]) then
    return redis.call('DECRBY', KEYS[1], ARGV[1])
else
    return -2  -- 庫存不足
end
```

**為什麼用 Lua script？** Redis 執行 Lua script 時是原子的（single-threaded），不會有其他命令插進來。

---

## 六、2PC vs Saga

```
  2PC（兩階段提交）：

  ┌─────────────┐
  │  協調者       │
  └──────┬──────┘
         │
    Phase 1: Prepare（投票階段）
         │
         ├──── 準備? ────▶ 庫存服務: OK ✅
         ├──── 準備? ────▶ 支付服務: OK ✅
         └──── 準備? ────▶ 訂單服務: OK ✅
         │
    Phase 2: Commit（提交階段）
         │
         ├──── 提交! ────▶ 庫存服務: committed
         ├──── 提交! ────▶ 支付服務: committed
         └──── 提交! ────▶ 訂單服務: committed

  問題：任何一方在 Phase 2 掛了 → 全部卡住等恢復


  Saga（補償交易）：

  扣庫存 ✅ → 建訂單 ✅ → 支付 ✅ → 出票 ❌
                                       ↓
  回補庫存 ← 取消訂單 ← 退款 ← 補償開始

  結果：不會卡住，但需要寫補償邏輯
```

| 比較 | 2PC | Saga |
|------|-----|------|
| 一致性 | 強一致 | 最終一致 |
| 效能 | 差（全程加鎖） | 好（無全域鎖） |
| 可用性 | 低（協調者 SPOF） | 高（各服務獨立） |
| 複雜度 | 中（協調者邏輯） | 高（每步要寫補償） |
| 搶票適用 | **不適合** | **適合** |

---

## 七、對帳（Reconciliation）— 最終一致性的保障

### 為什麼需要對帳？

即使有 Saga、Outbox、Idempotency，仍然可能出現不一致：
- 網路超時後其實執行成功了（幽靈操作）
- Bug 導致補償邏輯有漏洞
- 第三方服務（支付閘道）狀態和本地不同步

**對帳是最終一致性的最後防線。**

### 對帳流程

```
  每 30 秒 / 每分鐘 / 每天（根據業務需求）

  ┌──────────────────────────────────────────────────────┐
  │                  對帳程式（Reconciler）                │
  └────────────────────────┬─────────────────────────────┘
                           │
              ┌────────────┼────────────────┐
              ▼            ▼                ▼
  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
  │  Redis 庫存   │  │  DB 訂單      │  │  支付閘道     │
  │  remaining:  │  │  orders:     │  │  payments:   │
  │  4997        │  │  3 筆成功     │  │  3 筆成功     │
  └──────────────┘  └──────────────┘  └──────────────┘
              │            │                │
              └────────────┼────────────────┘
                           ▼
  ┌──────────────────────────────────────────────────────┐
  │  比對結果：                                           │
  │                                                      │
  │  order-001: 庫存 ✅  訂單 ✅  支付 ✅  → 一致 ✅       │
  │  order-002: 庫存 ✅  訂單 ✅  支付 ✅  → 一致 ✅       │
  │  order-003: 庫存 ✅  訂單 ✅  支付 ❌  → 不一致！      │
  │             ↳ 訂單已建立但支付未完成 → 取消訂單+回補庫存│
  │  order-004: 庫存 ✅  訂單 ❌  支付 ❌  → 不一致！      │
  │             ↳ 庫存已扣但沒訂單 → 回補庫存              │
  └──────────────────────────────────────────────────────┘
```

### 不一致情境的修復策略

| 狀態 | 庫存 | 訂單 | 支付 | 修復方式 |
|------|------|------|------|---------|
| 正常完成 | 已扣 | 已建立 | 已支付 | 不需修復 |
| 支付失敗 | 已扣 | 已建立 | 未支付 | 取消訂單 + 回補庫存 |
| 訂單丟失 | 已扣 | 不存在 | 不存在 | 回補庫存 |
| 幽靈支付 | 未扣 | 不存在 | 已支付 | 呼叫支付閘道退款 |
| 重複扣款 | 已扣 | 已建立 | 多筆支付 | 退掉多餘的支付 |

### 對帳程式碼骨架

```go
type Reconciler struct {
    inventoryDB *sql.DB
    orderDB     *sql.DB
    paymentAPI  PaymentGateway
}

func (r *Reconciler) Run(ctx context.Context, eventID string) ([]Discrepancy, error) {
    // 1. 取得各系統的資料
    inventory := r.getInventoryRecords(ctx, eventID)
    orders := r.getOrderRecords(ctx, eventID)
    payments := r.getPaymentRecords(ctx, eventID)

    // 2. 比對
    discrepancies := r.compare(inventory, orders, payments)

    // 3. 自動修復可修復的
    for _, d := range discrepancies {
        if d.AutoFixable {
            r.fix(ctx, d)
        } else {
            r.alertOnCall(d) // 不能自動修的通知人工
        }
    }

    return discrepancies, nil
}
```

---

## 八、最終一致性在實務中的意義

### 對使用者來說「最終一致」是什麼體驗？

```
  使用者視角                          系統內部
  ──────────                          ──────────

  「我搶到票了！」                    訂單 created ✅
       │                              庫存 deducted ✅
       │                              支付... pending
       ▼
  「等待付款確認...」                  支付閘道回覆中...
       │                              （3-5 秒）
       ▼
  「付款成功！」                      支付 completed ✅
       │                              出票... pending
       ▼
  「查看我的票券」                    ticket sent ✅
                                      所有系統一致 ✅
                                      ↑
                                      這個「一致」可能是
                                      在 30 秒到 5 分鐘內
                                      逐步達成的
```

### 設計最終一致性系統的原則

| 原則 | 說明 | 搶票系統範例 |
|------|------|------------|
| 告知使用者狀態 | 不要假裝是即時的 | 「訂單處理中...」 |
| 設定預期 | 告訴使用者大概要多久 | 「預計 3-5 分鐘內出票」 |
| 提供查詢管道 | 讓使用者能查進度 | 訂單狀態頁面 |
| 有明確的超時 | 超過時間就走補償流程 | 15 分鐘未完成 → 自動取消 |
| 對帳兜底 | 系統自動偵測修復 | 每分鐘跑一次 reconciler |

---

## 九、搶票系統完整流程

從 Token 到出票的完整一致性保護流程：

```
  使用者點擊「搶票」
         │
         ▼
  ┌─────────────────────────────────────┐
  │  1. Token 發放                       │
  │     atomic.Add(&remaining, -1)       │
  │     >= 0 → 發 token（HMAC 簽名）     │
  │     < 0  → 「售罄」                  │
  │     一致性保護：原子操作防超發          │
  └──────────────┬──────────────────────┘
                 ▼
  ┌─────────────────────────────────────┐
  │  2. 座位鎖定                         │
  │     驗證 token 簽名                   │
  │     SET seat:{id} {user} EX 600 NX  │
  │     一致性保護：分散式鎖 + TTL         │
  │     Fencing token：防止過期鎖寫入     │
  └──────────────┬──────────────────────┘
                 ▼
  ┌─────────────────────────────────────┐
  │  3. 購票 Saga                        │
  │     扣庫存 → 建訂單 → 扣款 → 出票    │
  │     一致性保護：                      │
  │       Saga 補償（失敗自動回滾）        │
  │       Idempotency（每步冪等）         │
  │       WAL（crash 後恢復進度）          │
  └──────────────┬──────────────────────┘
                 ▼
  ┌─────────────────────────────────────┐
  │  4. 通知 + 對帳                      │
  │     Outbox：DB + MQ 一致             │
  │     Reconciler：定期比對修復          │
  │     一致性保護：最終一致性             │
  └─────────────────────────────────────┘
```

---

## 面試題對照

| 面試問題 | 考的概念 | 本課位置 |
|---------|---------|---------|
| 搶票系統怎麼選 CAP？ | CAP 取捨 | 第一節 |
| 100 萬人搶 5000 票怎麼設計？ | Inventory Token | 第二節 |
| 座位鎖怎麼實作？ | 分散式鎖 | 第三節 |
| 樂觀鎖和悲觀鎖差在哪？ | 鎖策略 | 第四節 |
| 為什麼不用 2PC？ | 2PC vs Saga | 第六節 |
| 怎麼防止超賣？ | 原子操作 | 第五節 |
| 錢扣了沒票怎麼辦？ | 最終一致性 + 對帳 | 第七節 |
| 最終一致性使用者會察覺嗎？ | 使用者體驗設計 | 第八節 |

---

## FAQ

### Q1: CAP 是說系統只能選兩個嗎？平常三個不能同時滿足？

**不是。** CAP 定理是說「當網路分區發生時（P 必然存在），你只能在 C 和 A 之間選一個」。在正常運作時，三者可以同時滿足。很多人誤解 CAP 是三選二，其實更精確的說法是：**P 是前提，在 P 發生時選 C 或 A。**

現實中網路分區雖然不常見，但一定會發生（海底電纜斷裂、機房網路異常），所以系統設計必須預先決定分區時的行為。

### Q2: 分散式鎖用 Redis 安全嗎？Redlock 呢？

**單機 Redis 鎖在大部分場景夠用。** 如果 Redis 是主從架構且發生 failover，鎖可能丟失（主機掛了、從機沒收到鎖就升為主機）。

Martin Kleppmann（《Designing Data-Intensive Applications》作者）指出 Redlock 也有問題（時鐘漂移、GC pause）。更安全的做法是：
1. 用 Redis 鎖 + fencing token
2. DB 端用 fencing token 做最終檢查
3. 如果要強一致，用 etcd 或 ZooKeeper 的分散式鎖

### Q3: 樂觀鎖在高衝突場景會不會一直重試到超時？

**會。** 這就是為什麼搶票座位要用悲觀鎖。解法：

1. 設定重試上限（最多 3 次）
2. 加 exponential backoff + jitter
3. 超過上限就告訴使用者「座位已被搶走」
4. 或者改用悲觀鎖 — 讓請求排隊而不是重試

經驗法則：**衝突率 > 20% 就該用悲觀鎖**。

### Q4: 對帳跑很久會影響線上服務嗎？

**不會，只要設計正確。** 最佳實踐：

1. 對帳用 read replica（唯讀副本），不打主庫
2. 分批查詢，每批加 `LIMIT` 和 `WHERE created_at > last_checkpoint`
3. 修復動作走正常的 API 流程（不直接改 DB），自然有限流保護
4. 核心對帳（金額）每分鐘跑一次，非核心（統計）每天跑一次

### Q5: 最終一致性的「最終」是多久？能保證嗎？

**取決於你的設計，不是「某天會一致」的空話。** 搶票系統的做法：

| 一致性層級 | 時間 | 機制 |
|-----------|------|------|
| Saga 內部 | 秒級（3-10 秒） | Saga orchestrator 自動補償 |
| Outbox 事件 | 秒級（5-30 秒） | Relay 每 5 秒掃描一次 |
| 對帳修復 | 分鐘級（1-5 分鐘） | Reconciler 每分鐘跑一次 |
| 人工介入 | 小時級 | Dead letter → on-call 處理 |

設定 SLA：「99.9% 的不一致在 5 分鐘內自動修復。」

---

## 練習

1. 用 Redis `SET NX EX` 實作分散式鎖，並加上自動續約機制
2. 實作 Inventory Token 模式：發放 N 個 token，搶到 token 才能購票
3. 設計一個對帳程式：比對 Redis 庫存和 DB 訂單數量，找出差異
4. 思考 CAP 定理：搶票系統在 Redis 分區時應該選擇 CP 還是 AP？為什麼？
5. 實作樂觀鎖：用 version 欄位防止並發更新衝突

---

## 三課總結（40-42）

| 課程 | 核心問題 | 一句話答案 |
|------|---------|-----------|
| 40 容錯 | 系統壞了怎麼辦？ | WAL 恢復、Saga 補償、冪等重試、Outbox 保一致 |
| 41 高可用 | 怎麼讓系統不停擺？ | 複製、Sentinel、多層降級、六層流量防線 |
| 42 一致性 | 怎麼保證資料對得上？ | Token 控流量、原子防超賣、鎖定座位、事件對帳修復 |

---

## 下一課預告

恭喜你完成了整個 42 課的旅程！回顧 `ROADMAP.md` 規劃你的進階學習方向。
