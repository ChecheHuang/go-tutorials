# 第三十五課：Prometheus 監控

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | 了解 metrics 概念，知道 Counter/Gauge/Histogram 的差異 |
| 🔴 資深工程師 | **必備**：能設計監控策略，定義 SLO，撰寫 PromQL 查詢，建立 Grafana Dashboard |
| 🏢 DevOps/SRE | 核心技能之一，與 AlertManager 整合設定告警規則 |

## 核心概念

### 四種 Metric 類型

| 類型 | 特性 | 適用場景 | PromQL 範例 |
|------|------|----------|-------------|
| **Counter** | 只增不減 | 請求數、錯誤數 | `rate(requests_total[5m])` |
| **Gauge** | 可增可減 | 當前連線數、記憶體 | `memory_usage_bytes` |
| **Histogram** | 分 bucket 統計 | 回應時間分布 | `histogram_quantile(0.95, ...)` |
| **Summary** | 客戶端計算分位數 | 精確百分位數 | `request_duration_p99` |

### Prometheus Pull 模型

```
應用程式 ─── /metrics ──▶ Prometheus Server ─── 儲存 ──▶ Grafana 視覺化
              （暴露數據）   （定期拉取）                    （圖表/告警）
```

## 使用方式

```bash
go run ./tutorials/35-prometheus

# 訪問 metrics
curl http://localhost:8080/metrics

# 觸發業務 API（產生 metrics 數據）
curl http://localhost:8080/api/posts

# 瀏覽器示範頁面
open http://localhost:8080/demo
```

## 搭配 Prometheus + Grafana

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'my-go-app'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
```

```bash
# 用 Docker 快速啟動
docker run -d -p 9090:9090 \
  -v ./prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus

docker run -d -p 3000:3000 grafana/grafana
```

## 常用 PromQL 查詢

```promql
# 每秒請求數（5分鐘滾動平均）
rate(http_requests_total[5m])

# HTTP 錯誤率
rate(http_requests_total{status=~"5.."}[5m])
  / rate(http_requests_total[5m]) * 100

# p95 回應時間（Histogram）
histogram_quantile(0.95,
  rate(http_request_duration_seconds_bucket[5m])
)

# 快取命中率
rate(cache_hits_total[5m])
  / (rate(cache_hits_total[5m]) + rate(cache_misses_total[5m]))
```

## 黃金信號（Google SRE Book）

監控系統必須涵蓋這四個指標：

| 信號 | 說明 | Metric 範例 |
|------|------|-------------|
| **Latency（延遲）** | 請求花多少時間 | `http_request_duration_seconds` |
| **Traffic（流量）** | 每秒幾個請求 | `rate(http_requests_total[5m])` |
| **Errors（錯誤）** | 錯誤率是多少 | `rate(http_requests_total{status="500"}[5m])` |
| **Saturation（飽和度）** | 系統還有多少餘量 | `go_goroutines`, `process_resident_memory_bytes` |

## 本課程重點

- `prometheus.NewCounterVec` — 帶 Labels 的 Counter
- `prometheus.NewHistogramVec` — 帶自訂 buckets 的 Histogram
- `prometheus.MustRegister` — 啟動時註冊（失敗立即 panic）
- `promhttp.Handler()` — 暴露 `/metrics` 端點
- Middleware 模式 — 所有 handler 自動被監控
