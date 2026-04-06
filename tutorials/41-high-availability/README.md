# 第四十一課：高可用架構（High Availability）

> **一句話總結**：100 萬人排隊中 Redis crash，系統不能停——這就是高可用。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🔴 資深工程師 | **重點**：理解 Replication、Failover 概念 |
| ⚫ Staff / Principal | **核心能力**：設計多層防線、處理腦裂問題 |

## 你會學到什麼？

- 什麼是高可用？99.9% vs 99.99% vs 99.999% 的真實差距
- 單點故障（SPOF）辨識與消除
- 主從複製（Replication）：資料不只一份
- Redis Sentinel vs Cluster：選擇與比較
- Sentinel 自動故障轉移的完整流程
- Multi-Queue Failover：Redis 掛了自動切 DB
- 六層流量防線：100 萬 QPS 降到 100
- 資料庫 HA：主從複製與讀寫分離
- 多組件系統的健康檢查
- 優雅降級策略
- 腦裂（Split Brain）：兩個 Primary 同時存在的災難

## 執行方式

```bash
go run ./tutorials/41-high-availability/
```

---

## 生活比喻：高可用 = 醫院的急診室

```
急診室永遠不能關門——這就是高可用。

1. 人員冗餘：不只一個醫生值班（Replication）
2. 自動替補：醫生 A 倒下，醫生 B 立刻接手（Failover）
3. 分級看診：輕症到一般門診，重症才進急診（流量分層）
4. 備用電源：市電斷了，發電機 3 秒內啟動（備援機制）
5. 損壞控制：一間手術室漏水，其他手術室繼續運作（隔離故障）
```

| 醫院概念 | 系統對應 | 說明 |
|---------|---------|------|
| 多位醫生值班 | Redis Replica | 資料有多份副本 |
| 護理站監控 | Sentinel | 持續監控健康狀態 |
| 醫生倒下自動替補 | Automatic Failover | 主節點掛了自動切換 |
| 分級看診 | 六層流量防線 | 不是所有請求都打到核心 |
| 備用發電機 | Multi-Queue Fallback | Redis 掛了切 DB |
| 檢傷分類 | Rate Limiter | 過載時優先處理重要請求 |

---

## 什麼是「高可用」？

高可用（High Availability, HA）用「幾個 9」衡量：

| SLA | 可用性 | 每年允許停機 | 每月允許停機 | 每天允許停機 | 適合場景 |
|-----|-------|------------|------------|------------|---------|
| 99% | 兩個 9 | 3.65 天 | 7.3 小時 | 14.4 分鐘 | 內部工具 |
| 99.9% | 三個 9 | 8.76 小時 | 43.8 分鐘 | 1.44 分鐘 | 一般 Web 應用 |
| 99.99% | 四個 9 | 52.6 分鐘 | 4.38 分鐘 | 8.6 秒 | 電商、搶票系統 |
| 99.999% | 五個 9 | 5.26 分鐘 | 26.3 秒 | 0.86 秒 | 金融、醫療 |

**串聯可用性公式**：如果系統有 5 個元件串聯，每個 99.9%：

```
總可用性 = 0.999 ^ 5 = 0.995 = 99.5%
每年停機 = 365 * 24 * (1 - 0.995) = 43.8 小時 😱
```

**結論**：每多一個串聯元件，可用性就下降。所以每個元件都要有冗餘。

---

## 單點故障（SPOF）辨識

**SPOF = Single Point of Failure**：只要它掛了，整個系統就掛了。

```
找出 SPOF 的方法：逐一問「如果這個掛了會怎樣？」

┌─────────────────────────────────────────────────────────┐
│  搶票系統架構                                            │
│                                                          │
│  Load Balancer  ← SPOF？用雲端 LB（AWS ALB）→ HA       │
│       │                                                  │
│  API Server x3  ← 3 台，掛一台沒事 → ✅                 │
│       │                                                  │
│  Redis（單台）  ← SPOF！掛了 queue 就沒了 → ❌          │
│       │                                                  │
│  PostgreSQL     ← SPOF！掛了資料就讀不到 → ❌           │
│       │                                                  │
│  付款服務       ← 外部依賴 → 用 Circuit Breaker 保護    │
└─────────────────────────────────────────────────────────┘
```

**消除 SPOF 的方法**：

| 元件 | SPOF 風險 | 解法 |
|------|----------|------|
| Load Balancer | 中 | 使用雲端 LB（本身就是 HA） |
| API Server | 低 | 多台 + HPA 自動擴展 |
| Redis | **高** | Sentinel 或 Cluster |
| PostgreSQL | **高** | Primary-Replica + 自動 Failover |
| 付款服務 | 外部 | Circuit Breaker + Fallback |
| DNS | 低 | 多 DNS 提供商（Route 53 + Cloudflare） |

---

## Redis HA：Sentinel vs Cluster

| 方案 | 自動 Failover | 寫入擴展 | 資料分片 | 最少節點 | 適合場景 |
|------|:---:|:---:|:---:|:---:|---------|
| 主從複製 | ❌ 手動 | ❌ | ❌ | 2 | 小型系統、讀多寫少 |
| **Sentinel** | ✅ 自動 | ❌ | ❌ | 5（1M+2R+3S）| 中型系統（搶票 queue）|
| **Cluster** | ✅ 自動 | ✅ 分片 | ✅ | 6（3M+3R）| 大型系統（百萬 QPS）|

### Sentinel 架構與 Failover 流程

```
正常狀態：
                    ┌─────────────────┐
                    │  Sentinel-1 (S) │
                    │  Sentinel-2 (S) │  ← 3 個哨兵持續監控
                    │  Sentinel-3 (S) │
                    └────────┬────────┘
                             │ 監控
                             ▼
    ┌──────────┐     ┌──────────┐     ┌──────────┐
    │ Primary  │────→│ Replica-1│     │ Replica-2│
    │ (寫入)   │────→│ (唯讀)   │     │ (唯讀)   │
    └──────────┘     └──────────┘     └──────────┘
         │                 ▲                ▲
         └─── 非同步複製 ──┘────────────────┘

Failover 流程：
  Step 1: Sentinel-1 偵測到 Primary 無回應
          → 標記 SDOWN（Subjectively Down，主觀下線）

  Step 2: Sentinel-2, 3 也確認無回應
          → 標記 ODOWN（Objectively Down，客觀下線）
          → 需要 Quorum（多數決）同意

  Step 3: Sentinel 之間選出 Leader
          → Leader 負責執行 Failover

  Step 4: 選出新的 Primary（通常選同步最快的 Replica）
          → Replica-1 升級為 Primary

  Step 5: 通知其他 Replica 改跟新 Primary 同步
          → 通知應用程式切換連線

                    ┌─────────────────┐
                    │  Sentinel-1 (S) │
                    │  Sentinel-2 (S) │
                    │  Sentinel-3 (S) │
                    └────────┬────────┘
                             │
                             ▼
    ┌──────────┐     ┌──────────┐     ┌──────────┐
    │ 舊Primary│     │ Replica-1│     │ Replica-2│
    │  (已掛)  │     │→新Primary│     │ (唯讀)   │
    └──────────┘     └──────────┘     └──────────┘
                           │                ▲
                           └── 非同步複製 ──┘
```

---

## 六層流量防線（詳解）

這是搶票系統最重要的架構——不是「怎麼處理 100 萬請求」，而是「怎麼讓 100 萬變成 100」：

```
100 萬請求
    ↓
┌─ 1. CDN / 靜態快取 ───────────── 擋掉 70%（靜態資源不打後端）
│   30 萬
├─ 2. WAF / Bot 過濾 ───────────── 擋掉 50%（機器人、黃牛）
│   15 萬
├─ 3. API Gateway ───────────────── 擋掉 20%（認證失敗、格式錯誤）
│   12 萬
├─ 4. Rate Limiter ──────────────── 擋掉 90%（每人每秒 1 次）
│   1.2 萬
├─ 5. Waiting Room Queue ───────── 每秒只放 120 人進入
│   120
└─ 6. Seat Lock + Circuit Breaker ─ 鎖座位 + 保護付款服務
    120
```

### 每一層的作用

| 層 | 名稱 | 擋什麼 | 實作方式 | 擋掉比例 |
|----|------|--------|---------|---------|
| 1 | **CDN 靜態快取** | 靜態資源請求（CSS/JS/圖片） | Cloudflare / CloudFront | ~70% |
| 2 | **WAF / Bot 過濾** | 機器人、爬蟲、DDoS 攻擊 | WAF 規則 + 行為分析 | ~50% |
| 3 | **API Gateway** | 無效請求（認證失敗、格式錯誤） | JWT 驗證、Schema 檢查 | ~20% |
| 4 | **Rate Limiter** | 超頻請求（每人每秒限 1 次） | Token Bucket / Sliding Window | ~90% |
| 5 | **Waiting Room** | 超過處理能力的請求 | Redis Queue + 排隊機制 | ~99% |
| 6 | **Seat Lock + CB** | 超賣 + 下游故障 | Redis Lock + gobreaker | 最終關卡 |

### 為什麼要分這麼多層？

```
如果只有 Rate Limiter：
  100 萬請求 → 全部打到 Rate Limiter → Rate Limiter 本身先掛了

分六層的好處：
  每一層只需要處理上一層過濾後的流量
  第 4 層 Rate Limiter 只需要處理 12 萬，不是 100 萬
  就算 Rate Limiter 掛了，前面 3 層已經擋掉 88% 了
```

---

## 資料庫 HA：Primary-Replica 與讀寫分離

```
                    寫入請求
                       │
                ┌──────▼──────┐
                │   Primary   │  ← 只有一台能寫
                │ (PostgreSQL) │
                └──┬──────┬───┘
                   │      │
            WAL 串流複製  WAL 串流複製
                   │      │
            ┌──────▼──┐ ┌─▼─────────┐
            │ Replica-1│ │ Replica-2 │  ← 多台可讀
            │ (唯讀)   │ │ (唯讀)    │
            └──────────┘ └───────────┘
                   ▲           ▲
                   │           │
                讀取請求     讀取請求
```

### 讀寫分離策略

```go
type DBRouter struct {
    primary  *sql.DB   // 寫入用
    replicas []*sql.DB // 讀取用
    mu       sync.Mutex
    next     int
}

// 寫入 → 一定走 Primary
func (r *DBRouter) Writer() *sql.DB {
    return r.primary
}

// 讀取 → Round-Robin 分散到 Replica
func (r *DBRouter) Reader() *sql.DB {
    r.mu.Lock()
    defer r.mu.Unlock()
    db := r.replicas[r.next%len(r.replicas)]
    r.next++
    return db
}
```

**注意**：Replica 有複製延遲（通常 < 1 秒），剛寫入的資料可能在 Replica 上讀不到。關鍵讀取（例如剛付款完查訂單）要走 Primary。

---

## 多組件健康檢查

```go
type HealthStatus struct {
    Status     string                 `json:"status"`
    Components map[string]ComponentHealth `json:"components"`
}

type ComponentHealth struct {
    Status  string `json:"status"`
    Latency string `json:"latency,omitempty"`
    Error   string `json:"error,omitempty"`
}

func healthCheckHandler(c *gin.Context) {
    status := HealthStatus{
        Status:     "healthy",
        Components: make(map[string]ComponentHealth),
    }

    // 檢查 PostgreSQL
    start := time.Now()
    if err := db.PingContext(c.Request.Context()); err != nil {
        status.Components["postgres"] = ComponentHealth{
            Status: "unhealthy", Error: err.Error(),
        }
        status.Status = "degraded"
    } else {
        status.Components["postgres"] = ComponentHealth{
            Status: "healthy", Latency: time.Since(start).String(),
        }
    }

    // 檢查 Redis
    start = time.Now()
    if err := rdb.Ping(c.Request.Context()).Err(); err != nil {
        status.Components["redis"] = ComponentHealth{
            Status: "unhealthy", Error: err.Error(),
        }
        status.Status = "degraded"
    } else {
        status.Components["redis"] = ComponentHealth{
            Status: "healthy", Latency: time.Since(start).String(),
        }
    }

    code := 200
    if status.Status != "healthy" {
        code = 503
    }
    c.JSON(code, status)
}
```

回應範例：

```json
{
  "status": "degraded",
  "components": {
    "postgres": { "status": "healthy", "latency": "1.2ms" },
    "redis":    { "status": "unhealthy", "error": "connection refused" }
  }
}
```

---

## 優雅降級策略

當某個元件出問題時，不是整個系統掛掉，而是降低部分功能：

| 故障元件 | 降級策略 | 使用者體驗 |
|---------|---------|-----------|
| Redis Cache 掛了 | 直接打 DB（變慢但能用） | 頁面載入 2s → 5s |
| Redis Queue 掛了 | 切換到 DB-based Queue | 排隊功能正常，但較慢 |
| 付款服務掛了 | 熔斷 + 保留座位 5 分鐘 | 「付款暫時不可用，票已保留」 |
| 推薦服務掛了 | 顯示熱門票券（靜態資料） | 推薦不精準但有東西看 |
| 通知服務掛了 | 放入重試佇列 | 通知延遲但不丟失 |

```go
// Redis 掛了的降級範例
func GetTicketCount(ctx context.Context, eventID string) (int, error) {
    // 嘗試從 Redis 讀取（快）
    count, err := rdb.Get(ctx, "ticket:count:"+eventID).Int()
    if err == nil {
        return count, nil
    }

    // Redis 不可用 → 降級到 DB（慢但正確）
    log.Warn("Redis unavailable, falling back to DB", "error", err)
    row := db.QueryRowContext(ctx,
        "SELECT COUNT(*) FROM tickets WHERE event_id = $1 AND status = 'available'",
        eventID,
    )
    if err := row.Scan(&count); err != nil {
        return 0, fmt.Errorf("both Redis and DB failed: %w", err)
    }
    return count, nil
}
```

---

## 腦裂（Split Brain）

```
正常：
  [Primary] ←→ [Replica-1] ←→ [Replica-2]

網路分區：
  [Primary]     ╳     [Replica-1] ←→ [Replica-2]
   寫入 A                 ↑ 被 Sentinel 選為新 Primary
                         寫入 B

網路恢復：
  A 和 B 衝突了！哪個是對的？
```

**解法**：

1. **Quorum（多數決）**：需要 N/2+1 個 Sentinel 同意才能 failover
2. **min-replicas-to-write**：Primary 如果連不到足夠 Replica，拒絕寫入

```
# redis.conf — 防止腦裂
min-replicas-to-write 1     # 至少有 1 個 Replica 在線才接受寫入
min-replicas-max-lag 10     # Replica 延遲超過 10 秒算離線
```

---

## 搶票系統：生產環境就緒清單

上線前逐項確認：

| 類別 | 檢查項目 | 狀態 |
|------|---------|------|
| **冗餘** | API Server ≥ 2 台 | ☐ |
| **冗餘** | Redis Sentinel（1M+2R+3S） | ☐ |
| **冗餘** | PostgreSQL Primary + Replica | ☐ |
| **流量** | CDN 設定靜態資源快取 | ☐ |
| **流量** | WAF 規則啟用 | ☐ |
| **流量** | Rate Limiter 設定並壓測 | ☐ |
| **流量** | Waiting Room 佇列機制就緒 | ☐ |
| **監控** | Health Check 端點（/healthz, /readyz） | ☐ |
| **監控** | Prometheus + Grafana 看板 | ☐ |
| **監控** | 告警規則（Slack/PagerDuty） | ☐ |
| **韌性** | Circuit Breaker 保護付款服務 | ☐ |
| **韌性** | Graceful Shutdown 實作 | ☐ |
| **韌性** | Multi-Queue Fallback（Redis → DB） | ☐ |
| **部署** | Rolling Update 策略（maxUnavailable=0） | ☐ |
| **部署** | Resource Requests/Limits 設定 | ☐ |
| **測試** | 壓力測試（模擬搶票高峰） | ☐ |
| **測試** | 故障演練（Chaos Engineering） | ☐ |

---

## 面試題對照

| 面試問題 | 考的概念 | 本課位置 |
|---------|---------|---------|
| Redis 掛了系統怎麼不崩？ | Replication + Failover | Sentinel 架構 |
| Queue 掛了排隊資料怎麼辦？ | Multi-Queue Fallback | 優雅降級 |
| 100 萬人怎麼不打爆後端？ | 六層流量防線 | 流量防線詳解 |
| 兩個 Primary 同時寫入？ | Split Brain + Quorum | 腦裂 |
| 99.99% 可用性怎麼算？ | SLA 計算 | SLA 表 |
| 怎麼做讀寫分離？ | Primary-Replica | 資料庫 HA |

---

## FAQ

### Q1：99.9% 和 99.99% 差很多嗎？

差 10 倍。99.9% 每月可以停 43 分鐘，99.99% 只能停 4 分鐘。搶票系統如果在開賣那 10 分鐘掛了，43 分鐘的額度直接爆。

### Q2：Sentinel 和 Cluster 怎麼選？

看寫入量。如果單台 Redis 寫入扛得住（大部分場景），用 Sentinel 就夠了，維運簡單。如果需要水平擴展寫入能力（百萬 QPS 以上），才需要 Cluster。

### Q3：資料庫複製延遲怎麼處理？

關鍵讀取（剛寫入的資料）走 Primary，一般讀取走 Replica。可以在寫入後設一個短暫的 flag，表示「接下來 2 秒讀取走 Primary」。

### Q4：故障演練（Chaos Engineering）怎麼做？

從小開始：手動停一台 Redis Replica，看 Sentinel 會不會 failover。進階可以用 Chaos Monkey 隨機殺 Pod，確認系統能自動恢復。

### Q5：每個元件 99.9%，串聯後只有 99.5%，怎麼提高？

兩種方法：(1) 提高單一元件可用性（99.9% → 99.99%）；(2) 並聯冗餘——兩台並聯的可用性 = 1 - (1-0.999)^2 = 99.9999%。

---

## 練習

1. 用 Docker Compose 設定 Redis Sentinel（1 master + 2 replica + 3 sentinel）
2. 實作健康檢查端點：回傳 Redis、DB、MQ 的連線狀態
3. 模擬 master 故障：停掉 Redis master，觀察 Sentinel 自動 failover
4. 設計六層流量防線中的 Rate Limiter 層：用 Token Bucket 演算法實作
5. 計算可用性：如果每個元件 99.9% 可用，5 個元件串聯的總可用性是多少？

---

## 下一課預告

**第四十二課：系統回顧** — 回顧整個搶票系統的架構演進，從單體到微服務、從單機到高可用。
