# 第三十九課：Circuit Breaker（熔斷器）

> **一句話總結**：熔斷器就像家裡的保險絲——當電流異常時自動斷電，避免整棟房子燒起來。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | 了解熔斷器概念，知道為什麼需要它 |
| 🔴 資深工程師 | **必備**：能設計容錯策略，包含熔斷器、重試、超時、Fallback |
| 🏢 SRE/架構師 | 設計整體系統韌性（Resilience）策略 |

## 你會學到什麼？

- 為什麼需要熔斷器：串聯失敗（Cascading Failure）的災難
- 三個狀態：Closed → Open → Half-Open 的完整流程
- gobreaker v2 的 API 與設定調校
- 搭配 Retry + Exponential Backoff
- 搭配 context.WithTimeout 控制超時
- Fallback 降級策略：快取、預設值、降級服務
- 將熔斷器狀態暴露為 Prometheus 指標
- 搶票系統中的付款服務保護
- Bulkhead（隔艙模式）比較

## 執行方式

```bash
go run ./tutorials/39-circuit-breaker
```

---

## 生活比喻：熔斷器 = 家裡的保險絲

```
正常用電（Closed 狀態）：
  電流正常通過 → 冰箱、冷氣、電燈正常運作
  ✅ 每個請求都正常送到下游服務

電流異常（故障累積）：
  微波爐 + 烤箱 + 電暖器同時開 → 電流過載
  ⚠️ 下游服務回應越來越慢、錯誤越來越多

保險絲跳掉（Open 狀態）：
  啪！斷電了 → 所有電器都停了，但電線沒燒掉
  🛑 熔斷器開啟 → 所有請求立即失敗，不再打下游

試著復電（Half-Open 狀態）：
  推上開關 → 先試一下會不會又跳掉
  🟡 放幾個試探請求過去 → 成功就恢復，失敗就繼續斷電
```

---

## 為什麼需要熔斷器？

### 串聯失敗的災難

想像搶票系統的付款流程：

```
使用者 → API Server → 付款服務 → 銀行 API
                         │
                         ╳ 銀行 API 回應超時（10 秒）

沒有熔斷器：
  1000 個使用者同時付款
  → 1000 個 goroutine 都卡在等銀行回應
  → API Server 的 goroutine 全部被佔滿
  → 新的請求（包括不需要付款的）全部排隊
  → 整個系統掛了 💀

有熔斷器：
  前 5 個請求失敗 → 熔斷器開啟
  → 第 6-1000 個請求立即返回 "付款暫時不可用，請稍後再試"
  → API Server 的 goroutine 沒被佔滿
  → 查票、瀏覽等其他功能正常運作 ✅
```

**核心原則**：快速失敗（Fail Fast）比慢慢等死好。

---

## 三個狀態詳解

```
         ┌──────────────────────────────────────────┐
         │                                          │
         ▼                                          │
    ┌─────────┐    失敗率 ≥ 閾值     ┌──────────┐   │
    │ Closed  │ ──────────────────→ │   Open   │   │
    │（正常）  │                      │（熔斷）   │   │
    └─────────┘                      └────┬─────┘   │
         ▲                                │         │
         │ 試探成功                  等待 Timeout    │
         │                                │         │
    ┌────┴──────┐                   ┌─────▼──────┐  │
    │           │ ←──── 成功 ────── │ Half-Open  │  │
    │           │                   │（試探）     │  │
    │           │                   └─────┬──────┘  │
    └───────────┘                         │         │
                                    試探失敗 ───────┘
```

| 狀態 | 行為 | 轉換條件 |
|------|------|---------|
| **Closed（正常）** | 所有請求正常通過，統計失敗率 | 失敗率 ≥ 閾值 → Open |
| **Open（熔斷）** | 所有請求立即失敗，返回 `ErrOpenState` | 等待 Timeout 後 → Half-Open |
| **Half-Open（試探）** | 放 N 個試探請求通過 | 成功 → Closed；失敗 → Open |

### 狀態轉換的時間軸

```
時間：  0s ────── 10s ────── 15s ────── 20s ────── 25s
狀態：  Closed     │         Open        │       Half-Open
                   │                     │          │
        失敗率到達 60%             5 秒 Timeout    放 3 個請求試試
        → 切到 Open              → 切到 Half-Open  → 成功？回 Closed
                                                    → 失敗？回 Open
```

---

## gobreaker v2 完整設定

```go
import "github.com/sony/gobreaker/v2"

settings := gobreaker.Settings{
    Name:        "payment-service",    // 名稱（用於日誌和指標）
    MaxRequests: 3,                    // Half-Open 時允許的試探請求數
    Interval:    10 * time.Second,     // Closed 狀態的統計週期（歸零週期）
    Timeout:     5 * time.Second,      // Open → Half-Open 的等待時間
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        // 自訂觸發條件：5 次以上請求且失敗率 ≥ 60%
        return counts.Requests >= 5 &&
            float64(counts.TotalFailures)/float64(counts.Requests) >= 0.6
    },
    OnStateChange: func(name string, from, to gobreaker.State) {
        // 狀態變化時的回調（記日誌、發告警）
        log.Printf("Circuit breaker %s: %s → %s", name, from, to)
    },
    IsSuccessful: func(err error) bool {
        // 自訂什麼算「成功」（例如 4xx 不算熔斷器失敗）
        if err == nil {
            return true
        }
        var httpErr *HTTPError
        if errors.As(err, &httpErr) && httpErr.StatusCode < 500 {
            return true // 4xx 是客戶端錯誤，不算服務失敗
        }
        return false
    },
}

cb := gobreaker.NewCircuitBreaker[[]byte](settings)
```

### 設定參數調校指南

| 參數 | 建議值 | 設太小 | 設太大 |
|------|-------|--------|--------|
| **MaxRequests** | 3-5 | 只有 1 次機會，不容易恢復 | 太多試探請求打到還沒恢復的服務 |
| **Interval** | 10-30s | 短暫錯誤就觸發熔斷 | 反應太慢，已經大量失敗了才熔斷 |
| **Timeout** | 5-30s | 服務還沒恢復就試探 | 可以恢復了但還在等待 |
| **ReadyToTrip 閾值** | 50-70% | 太敏感，偶爾一兩個錯就熔斷 | 太遲鈍，大量失敗才熔斷 |

---

## 包裝外部呼叫

```go
// 建立熔斷器
cb := gobreaker.NewCircuitBreaker[*PaymentResult](settings)

// 包裝付款 API 呼叫
func (s *PaymentService) Pay(ctx context.Context, req *PaymentRequest) (*PaymentResult, error) {
    result, err := cb.Execute(func() (*PaymentResult, error) {
        return s.client.Charge(ctx, req)
    })

    // 判斷錯誤類型
    if err != nil {
        if errors.Is(err, gobreaker.ErrOpenState) {
            // 熔斷器開啟 → 快速失敗
            return nil, fmt.Errorf("付款服務暫時不可用，請稍後再試")
        }
        if errors.Is(err, gobreaker.ErrTooManyRequests) {
            // Half-Open 狀態，試探請求已滿
            return nil, fmt.Errorf("付款服務恢復中，請稍後再試")
        }
        // 其他錯誤（下游真的回錯誤了）
        return nil, fmt.Errorf("付款失敗: %w", err)
    }

    return result, nil
}
```

---

## 搭配 Retry + Exponential Backoff

重試和熔斷器不衝突：先重試幾次（處理短暫抖動），仍然失敗就算熔斷器的一次失敗。

```go
func callWithRetry(ctx context.Context, fn func() (string, error)) (string, error) {
    const maxRetries = 3
    var lastErr error

    for i := range maxRetries {
        result, err := fn()
        if err == nil {
            return result, nil
        }
        lastErr = err

        // Exponential Backoff: 100ms, 200ms, 400ms
        backoff := time.Duration(100<<i) * time.Millisecond
        select {
        case <-time.After(backoff):
        case <-ctx.Done():
            return "", ctx.Err()
        }
    }
    return "", lastErr
}

// 組合：Retry 包在 Circuit Breaker 裡面
result, err := cb.Execute(func() (string, error) {
    return callWithRetry(ctx, func() (string, error) {
        return callExternalAPI(ctx, req)
    })
})
```

```
請求的旅程：

 ┌─ Circuit Breaker ─────────────────────────────────┐
 │                                                    │
 │  ┌─ Retry（最多 3 次）──────────────────────────┐  │
 │  │                                               │  │
 │  │  ┌─ Timeout（每次最多 3 秒）──────────────┐  │  │
 │  │  │                                         │  │  │
 │  │  │   呼叫外部 API                          │  │  │
 │  │  │                                         │  │  │
 │  │  └─────────────────────────────────────────┘  │  │
 │  │                                               │  │
 │  └───────────────────────────────────────────────┘  │
 │                                                    │
 │  整個過程算 CB 的一次「成功」或「失敗」            │
 └────────────────────────────────────────────────────┘
```

---

## 搭配 context.WithTimeout

```go
func callWithTimeout(ctx context.Context, url string) ([]byte, error) {
    // 每次請求最多等 3 秒
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err // context 超時也會走到這裡
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)
}

// 組合使用
result, err := cb.Execute(func() ([]byte, error) {
    return callWithTimeout(ctx, "https://payment.example.com/charge")
})
```

---

## Fallback 降級策略

熔斷器開啟時，不一定要直接回傳錯誤。可以有更優雅的降級：

```go
func (s *PriceService) GetPrice(ctx context.Context, itemID string) (Price, error) {
    result, err := s.cb.Execute(func() (Price, error) {
        return s.remoteClient.FetchPrice(ctx, itemID)
    })

    if err == nil {
        // 成功 → 更新快取
        s.cache.Set(itemID, result)
        return result, nil
    }

    // 策略 1：返回快取的上次成功結果
    if cached, ok := s.cache.Get(itemID); ok {
        log.Warn("使用快取價格", "itemID", itemID)
        return cached, nil
    }

    // 策略 2：返回預設值
    if errors.Is(err, gobreaker.ErrOpenState) {
        log.Warn("返回預設價格", "itemID", itemID)
        return Price{Amount: 0, Currency: "TWD", IsEstimate: true}, nil
    }

    // 策略 3：呼叫備用服務
    return s.backupClient.FetchPrice(ctx, itemID)
}
```

| 策略 | 適用場景 | 範例 |
|------|---------|------|
| **快取結果** | 資料不常變動 | 商品價格、匯率 |
| **預設值** | 可以接受粗略結果 | 預估運費、推薦列表 |
| **降級服務** | 有備用資料源 | 從 DB 讀取而非 API |
| **空結果** | 非核心功能 | 推薦系統掛了 → 不顯示推薦 |
| **排入佇列** | 可以延遲處理 | 付款失敗 → 放入重試佇列 |

---

## 暴露為 Prometheus 指標

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    cbStateGauge = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "circuit_breaker_state",
            Help: "Circuit breaker state: 0=closed, 1=half-open, 2=open",
        },
        []string{"service"},
    )
    cbRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "circuit_breaker_requests_total",
            Help: "Total requests through circuit breaker",
        },
        []string{"service", "result"}, // result: success, failure, rejected
    )
)

// 在 OnStateChange 回調中更新
OnStateChange: func(name string, from, to gobreaker.State) {
    stateValue := map[gobreaker.State]float64{
        gobreaker.StateClosed:   0,
        gobreaker.StateHalfOpen: 1,
        gobreaker.StateOpen:     2,
    }
    cbStateGauge.WithLabelValues(name).Set(stateValue[to])
    log.Printf("Circuit breaker %s: %s → %s", name, from, to)
},
```

搭配 Grafana 告警：當 `circuit_breaker_state == 2`（Open）持續超過 1 分鐘，發 Slack 通知。

---

## 搶票系統中的付款服務保護

```
使用者搶到票 → 進入付款流程
                  │
          ┌───────▼───────┐
          │ Circuit Breaker│
          │ (payment-svc) │
          └───┬───────┬───┘
              │       │
         Closed?    Open?
              │       │
              ▼       ▼
        呼叫付款 API   立即返回：
              │        "付款暫時不可用，
              │         您的票已保留 5 分鐘，
              ▼         請稍後再試"
         成功/失敗
              │
              ▼
        更新訂單狀態
```

**重點**：熔斷時不要釋放使用者搶到的票。保留座位鎖，等付款服務恢復後讓使用者重試。

---

## Bulkhead（隔艙模式）比較

| 模式 | 保護目標 | 做法 |
|------|---------|------|
| **Circuit Breaker** | 保護下游服務 | 失敗率過高時切斷連線 |
| **Bulkhead** | 保護自己 | 限制對每個下游的並發數 |

```go
// Bulkhead 簡單實作：用 semaphore 限制並發
type Bulkhead struct {
    sem chan struct{}
}

func NewBulkhead(maxConcurrent int) *Bulkhead {
    return &Bulkhead{sem: make(chan struct{}, maxConcurrent)}
}

func (b *Bulkhead) Execute(fn func() error) error {
    select {
    case b.sem <- struct{}{}:
        defer func() { <-b.sem }()
        return fn()
    default:
        return errors.New("bulkhead: too many concurrent requests")
    }
}
```

**最佳實踐**：Circuit Breaker + Bulkhead 一起用。Bulkhead 限制並發數，Circuit Breaker 監測失敗率。

---

## FAQ

### Q1：熔斷器和 Rate Limiter 的差別？

Rate Limiter 保護自己不被打爆（限制進來的流量），Circuit Breaker 保護下游不被打爆（限制出去的流量）。兩者互補，不是替代。

### Q2：如果下游服務「很慢」但不是「完全掛掉」，熔斷器會怎麼反應？

單純的慢不會觸發熔斷器（因為沒有錯誤）。你需要搭配 `context.WithTimeout`：超時 = 錯誤 → 累積到熔斷閾值 → 開啟熔斷。

### Q3：每個下游服務需要一個獨立的熔斷器嗎？

是的。付款服務和通知服務應該有各自的熔斷器。付款服務掛了不應該影響通知服務——這就是 Bulkhead 的概念。

### Q4：熔斷器的狀態是存在記憶體中嗎？重啟會怎樣？

gobreaker 的狀態在記憶體中，Pod 重啟會重置為 Closed。這通常是合理的——重啟後應該重新評估下游狀態。如果需要跨 Pod 共享狀態，可以用 Redis 存。

### Q5：在微服務架構中，熔斷器應該放在 Client 端還是 Server 端？

放在 Client 端（呼叫方）。因為目的是保護呼叫方不被拖垮。Server 端用 Rate Limiter 保護自己。

---

## 練習

1. 用 gobreaker 包裝一個 HTTP client，設定 5 次失敗後開啟熔斷
2. 實作 fallback 機制：熔斷器開啟時回傳快取的上次成功結果
3. 把熔斷器狀態（Closed/Open/HalfOpen）暴露為 Prometheus 指標
4. 結合 context.WithTimeout，設定每次請求最多等 3 秒
5. 思考：如果下游服務「很慢」但不是「完全掛掉」，熔斷器會怎麼反應？

---

## 下一課預告

**第四十課：分散式一致性** — CAP 定理實戰、Saga vs 2PC、最終一致性、Inventory Token。
