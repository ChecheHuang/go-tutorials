# 第三十五課：Prometheus 監控

> **一句話總結**：Prometheus 定期拉取你的應用程式暴露的 metrics，讓你用數字而非猜測來了解系統的健康狀況。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | 了解 metrics 概念，知道 Counter/Gauge/Histogram 的差異 |
| 🔴 資深工程師 | **必備**：能設計監控策略，定義 SLO，撰寫 PromQL 查詢，建立 Grafana Dashboard |
| 🏢 DevOps/SRE | 核心技能之一，與 AlertManager 整合設定告警規則 |

## 你會學到什麼？

- 為什麼需要監控（不是等使用者抱怨才發現問題）
- Google SRE 的四個黃金信號：延遲、流量、錯誤、飽和度
- Prometheus 的 Pull 架構以及為什麼選擇這種設計
- 四種 Metric 類型：Counter、Gauge、Histogram、Summary
- PromQL 基礎查詢：`rate()`、`histogram_quantile()`、`sum by()`
- 在 Go 程式中加入自訂 metrics（使用 `promhttp`）
- Gin middleware 自動收集 HTTP 請求指標
- 告警規則的基本寫法
- Grafana 視覺化整合概覽
- 基數爆炸（Cardinality Explosion）的陷阱與防範

## 執行方式

```bash
go run ./tutorials/35-prometheus

# 訪問 metrics
curl http://localhost:8080/metrics

# 觸發業務 API（產生 metrics 數據）
curl http://localhost:8080/api/posts

# 瀏覽器示範頁面
open http://localhost:8080/demo
```

## 生活比喻：汽車儀表板

```
你開車上高速公路，儀表板上有：

  ┌─────────────────────────────────────────────┐
  │  🔵 速度表（目前 80 km/h）    → Gauge       │
  │  🟢 里程表（累計 45,230 km）  → Counter     │
  │  🟡 油量表（剩餘 60%）        → Gauge       │
  │  🔴 引擎溫度（正常/偏高/過熱）→ Histogram   │
  └─────────────────────────────────────────────┘

沒有儀表板的司機：
  「感覺車子怪怪的...」→ 等到拋錨在路邊才知道沒油了

有儀表板的司機：
  「油量低於 20%」→ 下一個交流道加油
  「引擎溫度偏高」→ 靠邊休息降溫

Prometheus 就是你的伺服器儀表板。
Counter = 里程表（只會往上跑）
Gauge   = 速度表（上上下下）
Histogram = 引擎溫度分布（多少時間在正常範圍、多少時間偏高）

沒有監控 = 盲駕。你不會想在高速公路上盲駕。
```

## 為什麼需要監控？

### 一個真實的故事

```
週五下午 5:00  — 部署新版本，一切看起來正常
週五下午 5:30  — 某個 API 回應時間從 50ms 變成 3s（但沒人知道）
週五下午 6:00  — 資料庫連線池耗盡（但沒人知道）
週五下午 7:00  — 開始有使用者抱怨「網站好慢」
週五下午 8:00  — 客服收到大量投訴，通知開發團隊
週五下午 9:00  — 開發人員開始排查，花了 1 小時才定位問題
週五下午 10:00 — 修復完成，但已經影響了數千名使用者

如果有 Prometheus + 告警：
週五下午 5:31  — 告警：P95 延遲超過 500ms
週五下午 5:32  — 開發人員收到 Slack 通知，開始排查
週五下午 5:45  — 發現是新版本的 N+1 query 問題，回滾部署
週五下午 5:50  — 恢復正常，影響使用者不到 100 人
```

**監控的價值 = 問題發現時間 x 影響範圍**。監控越好，兩個數字都越小。

## 四個黃金信號（Google SRE Book）

這是 Google SRE 團隊提出的監控系統必須涵蓋的四個指標：

```
┌──────────────────────────────────────────────────────────┐
│                    四個黃金信號                             │
│                                                          │
│  ┌─────────────┐  ┌─────────────┐                       │
│  │  Latency    │  │  Traffic    │                       │
│  │  延遲       │  │  流量       │                       │
│  │  請求花多久  │  │  每秒幾個   │                       │
│  └─────────────┘  └─────────────┘                       │
│                                                          │
│  ┌─────────────┐  ┌──────────────┐                      │
│  │  Errors     │  │  Saturation  │                      │
│  │  錯誤率     │  │  飽和度       │                      │
│  │  失敗比例   │  │  還剩多少資源  │                      │
│  └─────────────┘  └──────────────┘                      │
└──────────────────────────────────────────────────────────┘
```

| 信號 | 說明 | 問的問題 | Metric 範例 |
|------|------|---------|-------------|
| **Latency（延遲）** | 請求從進來到回應的時間 | 使用者要等多久？ | `http_request_duration_seconds` |
| **Traffic（流量）** | 系統正在處理多少請求 | 現在多忙？ | `rate(http_requests_total[5m])` |
| **Errors（錯誤）** | 失敗請求佔總請求的比例 | 多少人遇到問題？ | `rate(http_requests_total{status="500"}[5m])` |
| **Saturation（飽和度）** | 系統資源使用了多少 | 還能撐多久？ | `go_goroutines`, `process_resident_memory_bytes` |

> **重要**：Latency 要區分「成功請求的延遲」和「失敗請求的延遲」。一個回傳 500 的請求可能很快（0.1ms），但不代表系統健康。

## Prometheus 架構：Pull 模型

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│  ┌──────────┐    /metrics    ┌────────────────┐                │
│  │ Go App 1 │ ◀──── pull ───│                │                │
│  │ :8080    │                │                │                │
│  └──────────┘                │                │    ┌─────────┐ │
│                              │   Prometheus   │───▶│ Grafana │ │
│  ┌──────────┐    /metrics    │   Server       │    │ :3000   │ │
│  │ Go App 2 │ ◀──── pull ───│   :9090        │    └─────────┘ │
│  │ :8081    │                │                │                │
│  └──────────┘                │                │    ┌─────────┐ │
│                              │                │───▶│ Alert   │ │
│  ┌──────────┐    /metrics    │                │    │ Manager │ │
│  │ Node     │ ◀──── pull ───│                │    └─────────┘ │
│  │ Exporter │                └────────────────┘                │
│  └──────────┘                   │                              │
│                                 │ TSDB（時序資料庫）             │
│                                 │ 預設保留 15 天                 │
│                                 ▼                              │
│                          ┌──────────────┐                      │
│                          │ Local Storage│                      │
│                          └──────────────┘                      │
└─────────────────────────────────────────────────────────────────┘
```

### 為什麼用 Pull 而不是 Push？

| 比較項目 | Pull（Prometheus） | Push（例如 StatsD） |
|---------|-------------------|-------------------|
| 應用程式職責 | 只要暴露 `/metrics` 端點 | 要主動推送到監控伺服器 |
| 服務發現 | Prometheus 知道要監控哪些服務 | 監控伺服器被動接收，不確定有多少服務 |
| 存活偵測 | 拉不到 = 服務掛了（自帶健康檢查） | 沒收到資料 = 掛了？還是沒資料？ |
| 測試/除錯 | 可以直接 `curl /metrics` 查看 | 需要額外工具 |
| 背壓控制 | Prometheus 控制拉取頻率 | 服務可能推太多壓垮監控伺服器 |

## 四種 Metric 類型詳解

### Counter：只增不減的計數器

```go
import "github.com/prometheus/client_golang/prometheus"

// 定義
var httpRequestsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{"method", "path", "status"},  // labels
)

// 註冊
func init() {
    prometheus.MustRegister(httpRequestsTotal)
}

// 使用（在 middleware 中）
httpRequestsTotal.WithLabelValues("GET", "/api/articles", "200").Inc()
```

**適用場景**：請求總數、錯誤總數、處理的 bytes 總數

**PromQL 常見用法**：
```promql
# Counter 本身的值沒什麼意義（永遠往上跑）
# 要搭配 rate() 看每秒增加多少
rate(http_requests_total[5m])                          # 每秒請求數
rate(http_requests_total{status="500"}[5m])            # 每秒錯誤數
```

> **注意**：Counter 只能 `Inc()`（+1）或 `Add(n)`（+n），不能減少。服務重啟後會歸零，但 `rate()` 會自動處理這個情況。

### Gauge：可增可減的量表

```go
var activeConnections = prometheus.NewGauge(
    prometheus.GaugeOpts{
        Name: "active_connections",
        Help: "Number of active connections",
    },
)

var dbPoolSize = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "db_pool_connections",
        Help: "Number of database pool connections by state",
    },
    []string{"state"},  // "active", "idle"
)

// 使用
activeConnections.Inc()    // +1
activeConnections.Dec()    // -1
activeConnections.Set(42)  // 設為特定值

dbPoolSize.WithLabelValues("active").Set(float64(pool.ActiveCount()))
dbPoolSize.WithLabelValues("idle").Set(float64(pool.IdleCount()))
```

**適用場景**：目前連線數、記憶體使用量、佇列長度、Goroutine 數量

### Histogram：分布統計

```go
var requestDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "http_request_duration_seconds",
        Help: "HTTP request duration in seconds",
        // 自訂 bucket 邊界（單位：秒）
        Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
    },
    []string{"method", "path"},
)

// 使用
start := time.Now()
// ... 處理請求 ...
duration := time.Since(start).Seconds()
requestDuration.WithLabelValues("GET", "/api/articles").Observe(duration)

// 或者用 Timer 語法糖
timer := prometheus.NewTimer(requestDuration.WithLabelValues("GET", "/api/articles"))
defer timer.ObserveDuration()
```

**Histogram 的資料結構**：

```
假設請求延遲分別是：12ms, 45ms, 120ms, 350ms, 1.2s

bucket 邊界          落入的請求數（累計）
≤ 0.005s (5ms)     → 0
≤ 0.01s  (10ms)    → 0
≤ 0.025s (25ms)    → 1   (12ms)
≤ 0.05s  (50ms)    → 2   (12ms, 45ms)
≤ 0.1s   (100ms)   → 2
≤ 0.25s  (250ms)   → 3   (+120ms)
≤ 0.5s   (500ms)   → 4   (+350ms)
≤ 1s               → 4
≤ 2.5s             → 5   (+1.2s)
≤ 5s               → 5
≤ 10s              → 5
+Inf               → 5

另外還記錄：
  _sum   = 1.727  （所有觀測值的總和）
  _count = 5      （觀測次數）
```

**PromQL 常見用法**：
```promql
# P95 延遲（95% 的請求在多少秒以內完成）
histogram_quantile(0.95,
  rate(http_request_duration_seconds_bucket[5m])
)

# P50（中位數）
histogram_quantile(0.50,
  rate(http_request_duration_seconds_bucket[5m])
)

# 平均延遲
rate(http_request_duration_seconds_sum[5m])
  / rate(http_request_duration_seconds_count[5m])
```

### Summary：客戶端計算分位數

```go
var requestDurationSummary = prometheus.NewSummaryVec(
    prometheus.SummaryOpts{
        Name: "http_request_duration_summary_seconds",
        Help: "HTTP request duration summary",
        Objectives: map[float64]float64{
            0.5:  0.05,   // P50，誤差 5%
            0.9:  0.01,   // P90，誤差 1%
            0.99: 0.001,  // P99，誤差 0.1%
        },
    },
    []string{"method"},
)
```

### Histogram vs Summary

| 比較項目 | Histogram | Summary |
|---------|-----------|---------|
| 分位數計算位置 | 伺服器端（PromQL） | 客戶端（應用程式） |
| 可聚合 | 是（多台機器的 Histogram 可以合併） | 否（分位數無法跨實例加總） |
| 精確度 | 受 bucket 邊界影響 | 可配置誤差 |
| 效能開銷 | 較低 | 較高（需維護滑動窗口） |
| **建議** | **大多數情況用這個** | 極少使用 |

> **實務建議**：99% 的情況用 Histogram。Summary 只在你確定不需要跨實例聚合、且需要極精確分位數時才用。

## PromQL 基礎

### `rate()`：計算每秒增長率

```promql
# 過去 5 分鐘的平均每秒請求數
rate(http_requests_total[5m])

# 注意：rate() 只能用在 Counter 上
# 對 Gauge 用 rate() 沒有意義
```

### `sum by()`：按 label 分組加總

```promql
# 每個 HTTP method 的每秒請求數
sum by(method) (rate(http_requests_total[5m]))

# 每個 status code 的每秒請求數
sum by(status) (rate(http_requests_total[5m]))

# 錯誤率（百分比）
sum(rate(http_requests_total{status=~"5.."}[5m]))
  / sum(rate(http_requests_total[5m])) * 100
```

### `histogram_quantile()`：計算分位數

```promql
# P95 延遲
histogram_quantile(0.95,
  sum by(le) (rate(http_request_duration_seconds_bucket[5m]))
)

# 按 path 分組的 P99 延遲
histogram_quantile(0.99,
  sum by(le, path) (rate(http_request_duration_seconds_bucket[5m]))
)
```

### 常用 PromQL 查詢速查表

| 需求 | PromQL |
|------|--------|
| 每秒請求數（QPS） | `sum(rate(http_requests_total[5m]))` |
| HTTP 錯誤率 | `sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100` |
| P95 回應時間 | `histogram_quantile(0.95, sum by(le)(rate(http_request_duration_seconds_bucket[5m])))` |
| 平均回應時間 | `rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m])` |
| Goroutine 數量 | `go_goroutines` |
| 記憶體使用 | `process_resident_memory_bytes / 1024 / 1024` |
| 快取命中率 | `rate(cache_hits_total[5m]) / (rate(cache_hits_total[5m]) + rate(cache_misses_total[5m]))` |

## 在 Go 程式中加入自訂 Metrics

### 完整範例：HTTP Metrics Middleware

```go
package middleware

import (
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,  // 預設 bucket
        },
        []string{"method", "path"},
    )

    httpRequestsInFlight = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "http_requests_in_flight",
            Help: "Number of HTTP requests currently being processed",
        },
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
    prometheus.MustRegister(httpRequestsInFlight)
}

// PrometheusMiddleware 自動收集每個請求的 metrics
func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        // 正在處理的請求數 +1
        httpRequestsInFlight.Inc()

        // 執行後續的 handler
        c.Next()

        // 正在處理的請求數 -1
        httpRequestsInFlight.Dec()

        // 記錄延遲
        duration := time.Since(start).Seconds()
        httpRequestDuration.WithLabelValues(
            c.Request.Method,
            c.FullPath(),  // 用路由模板，不是實際 URL
        ).Observe(duration)

        // 記錄請求數
        status := strconv.Itoa(c.Writer.Status())
        httpRequestsTotal.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            status,
        ).Inc()
    }
}

// RegisterRoutes 設定 metrics 端點
func RegisterRoutes(r *gin.Engine) {
    r.Use(PrometheusMiddleware())
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
```

### 重要細節：用 `c.FullPath()` 而非 `c.Request.URL.Path`

```go
// 錯誤：使用實際路徑 → 無限多的 label 值
c.Request.URL.Path  // "/api/articles/1", "/api/articles/2", ...

// 正確：使用路由模板 → 有限的 label 值
c.FullPath()        // "/api/articles/:id" → 所有文章共用一個 label
```

## 搭配 Prometheus + Grafana

### Prometheus 設定

```yaml
# prometheus.yml
global:
  scrape_interval: 15s       # 每 15 秒拉取一次
  evaluation_interval: 15s    # 每 15 秒評估一次告警規則

scrape_configs:
  - job_name: 'my-go-app'
    static_configs:
      - targets: ['localhost:8080']
        labels:
          env: 'production'

  # 如果有多個實例
  - job_name: 'my-go-app-cluster'
    static_configs:
      - targets:
        - 'app1:8080'
        - 'app2:8080'
        - 'app3:8080'
```

### Docker Compose 快速啟動

```yaml
# docker-compose.yml
version: '3'
services:
  app:
    build: .
    ports:
      - "8080:8080"

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.retention.time=15d'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

```bash
docker compose up -d

# Prometheus UI: http://localhost:9090
# Grafana UI:    http://localhost:3000 (admin/admin)
```

### Grafana 設定步驟

```
1. 登入 Grafana (http://localhost:3000)
2. 設定 → Data Sources → Add → Prometheus
3. URL: http://prometheus:9090（Docker 內部網路）
4. 建立 Dashboard → Add Panel
5. 輸入 PromQL 查詢，例如：
   rate(http_requests_total[5m])
```

## 告警規則基礎

```yaml
# alert-rules.yml
groups:
  - name: http-alerts
    rules:
      # P95 延遲超過 500ms
      - alert: HighLatency
        expr: |
          histogram_quantile(0.95,
            sum by(le) (rate(http_request_duration_seconds_bucket[5m]))
          ) > 0.5
        for: 5m          # 持續 5 分鐘才觸發
        labels:
          severity: warning
        annotations:
          summary: "P95 延遲超過 500ms"
          description: "P95 延遲已達 {{ $value | printf \"%.2f\" }}s"

      # 錯誤率超過 5%
      - alert: HighErrorRate
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m]))
          / sum(rate(http_requests_total[5m])) * 100 > 5
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "HTTP 5xx 錯誤率超過 5%"
          description: "目前錯誤率為 {{ $value | printf \"%.1f\" }}%"

      # 服務掛掉（拉不到 metrics）
      - alert: ServiceDown
        expr: up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "服務 {{ $labels.instance }} 無回應"
```

### 告警流程

```
Prometheus 評估規則
    │
    ├── 條件不成立 → 什麼都不做
    │
    └── 條件成立
        │
        ├── 持續時間不足（< for）→ Pending 狀態
        │
        └── 持續時間達標（≥ for）→ Firing 狀態
            │
            ▼
        AlertManager
            │
            ├── 分組（Grouping）：同類告警合併
            ├── 靜默（Silencing）：維護時暫停告警
            ├── 抑制（Inhibition）：高優先級告警抑制低優先級
            │
            ▼
        通知管道
            ├── Slack
            ├── Email
            ├── PagerDuty
            └── Webhook
```

## 部落格專案：該追蹤哪些 Metrics？

| 指標 | 類型 | 用途 |
|------|------|------|
| `http_requests_total` | Counter | API 請求量、按路徑/狀態碼分 |
| `http_request_duration_seconds` | Histogram | 回應延遲分布，算 P50/P95/P99 |
| `http_requests_in_flight` | Gauge | 目前正在處理的請求數 |
| `blog_articles_total` | Gauge | 文章總數 |
| `blog_users_total` | Gauge | 使用者總數 |
| `db_query_duration_seconds` | Histogram | 資料庫查詢延遲 |
| `cache_hits_total` | Counter | 快取命中次數（如有使用 Redis） |
| `cache_misses_total` | Counter | 快取未命中次數 |
| `go_goroutines` | Gauge | Goroutine 數量（自動暴露） |
| `process_resident_memory_bytes` | Gauge | 記憶體使用量（自動暴露） |

## 基數爆炸警告（Cardinality Explosion）

**這是 Prometheus 使用者最常犯的致命錯誤。**

### 什麼是基數爆炸？

```go
// 大錯特錯！user_id 有幾百萬個不同的值
httpRequestsTotal.WithLabelValues(
    "GET",
    "/api/articles",
    "200",
    userID,         // 每個使用者 = 一個新的 time series
).Inc()

// 如果有 100 萬使用者 × 10 個 API × 3 個 method
// = 3000 萬個 time series
// Prometheus 記憶體和磁碟直接爆掉
```

### Label 值數量 vs Time Series 數量

```
Labels: method(3) × path(10) × status(5) = 150 個 time series  ✅ 安全
Labels: method(3) × path(10) × user_id(1M) = 3000 萬           💥 爆炸
Labels: method(3) × path(10) × request_id(∞) = 無限            💀 必死
```

### 什麼可以當 Label？什麼不行？

| 可以當 Label | 不可以當 Label |
|-------------|---------------|
| HTTP method（GET/POST/PUT/DELETE） | user_id |
| 狀態碼（200/400/404/500） | request_id |
| API 路由模板（`/api/articles/:id`） | session_id |
| 服務名稱（auth-service/api-service） | IP 地址 |
| 環境（prod/staging/dev） | email |
| 資料庫操作（SELECT/INSERT/UPDATE） | 實際的 URL path（含 ID） |

### 經驗法則

```
Label 的不同值數量（cardinality）：
  < 10    → 完美
  < 100   → 安全
  < 1000  → 要小心，確認有必要
  > 1000  → 幾乎肯定是錯的，重新設計
  > 10000 → Prometheus 會開始變慢或 OOM
```

## FAQ

### Q1：Prometheus 和 ELK（Elasticsearch + Logstash + Kibana）有什麼不同？

Prometheus 處理的是**數值型的時序資料**（metrics）——例如「每秒多少請求、P95 延遲多少」。ELK 處理的是**文字型的日誌**（logs）——例如「使用者 A 在 3:00 做了什麼操作」。兩者互補：Prometheus 告訴你「系統有問題了」，日誌告訴你「具體是什麼問題」。在可觀測性（Observability）的三大支柱中，Metrics（Prometheus）、Logs（ELK）和 Traces（Jaeger/OpenTelemetry）缺一不可。

### Q2：`rate()` 和 `irate()` 有什麼差別？

`rate()` 計算整個時間範圍的平均每秒增長率（例如 5 分鐘的平均值），適合告警和長期趨勢。`irate()` 只看最後兩個資料點的增長率，對短期突刺更敏感，適合 Dashboard 上看即時變化。一般建議告警規則用 `rate()`（避免短期抖動觸發誤報），Dashboard 上可以兩個都放。

### Q3：為什麼 `http_requests_total` 的 path label 要用路由模板而不是實際路徑？

因為實際路徑包含動態參數（`/api/articles/1`、`/api/articles/2`...），每個不同的 ID 都會產生一個新的 time series，導致基數爆炸。用路由模板（`/api/articles/:id`）就只有一個 time series，大幅減少 Prometheus 的記憶體和儲存壓力。

### Q4：Prometheus 可以監控多久以前的資料？

預設保留 15 天。可以透過 `--storage.tsdb.retention.time` 參數調整（例如設為 `90d`）。如果需要長期保存，建議搭配遠端儲存（Remote Storage），例如 Thanos、Cortex 或 VictoriaMetrics，它們可以將資料存到 S3/GCS 等物件儲存，保留數年的歷史資料。

### Q5：如果我的應用掛了，Prometheus 拉不到 metrics 怎麼辦？

Prometheus 會自動將 `up` 這個 metric 設為 0，代表目標無回應。你可以用 `up == 0` 作為告警條件。當應用重啟後，Counter 會從 0 重新開始，但 `rate()` 函式會自動偵測 Counter reset 並正確計算增長率，所以不用擔心重啟導致圖表異常。

## 練習

1. 為部落格 API 加入自訂指標：`blog_articles_total`（文章總數 Gauge）
2. 寫一個 PromQL 查詢：過去 5 分鐘的平均請求延遲
3. 設定一個告警規則：當 P95 延遲超過 2 秒時觸發
4. 用 Histogram 記錄資料庫查詢時間，並計算 P50/P95/P99
5. 思考：為什麼不應該用 user_id 作為 label？（提示：基數爆炸）

## 下一課預告

下一課我們會學習 **pprof 效能分析**——Prometheus 告訴你「系統變慢了」，但沒告訴你「為什麼慢」。pprof 可以深入程式內部，找出 CPU 和記憶體的瓶頸在哪一行程式碼。監控發現問題，pprof 定位問題，兩者搭配使用。
