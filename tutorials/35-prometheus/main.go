// ==========================================================================
// 第三十五課：Prometheus 監控
// ==========================================================================
//
// 為什麼需要監控？
//   程式上線後，你怎麼知道它運作正常？
//   - 每秒處理多少請求？
//   - 有多少請求失敗？
//   - 回應時間是否變長？
//   - 記憶體用量有沒有洩漏？
//
// Prometheus 是什麼？
//   Google 開源的監控系統（現為 CNCF 項目）：
//   - 應用程式暴露 /metrics 端點
//   - Prometheus Server 定期「拉取」（Pull）這些數據
//   - Grafana 把數據畫成漂亮的圖表
//
// 四種 Metric 類型：
//   Counter   → 只增不減（請求數、錯誤數）
//   Gauge     → 可增可減（當前連線數、記憶體用量）
//   Histogram → 分布統計（回應時間分布到各 bucket）
//   Summary   → 百分位數（p50, p95, p99 回應時間）
//
// 執行方式：go run ./tutorials/29-prometheus
// 然後訪問：http://localhost:8080/metrics  ← 看 Prometheus 格式的數據
//           http://localhost:8080/demo     ← 觸發一些請求示範
// ==========================================================================

package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ==========================================================================
// 1. 定義自訂 Metrics
// ==========================================================================
//
// Prometheus 內建的 metrics（runtime、GC 等）會自動收集
// 我們還需要定義「業務 metrics」來追蹤應用程式行為

// httpRequestsTotal HTTP 請求計數器（Counter）
// Counter 適合：請求次數、錯誤次數、處理的訊息數
// Labels（標籤）讓你可以按維度過濾：method="GET", path="/api/posts", status="200"
var httpRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",    // Metric 名稱（慣例：小寫、底線分隔）
		Help: "HTTP 請求總數，按方法、路徑、狀態碼分類", // 說明文字（在 /metrics 顯示）
	},
	[]string{"method", "path", "status"}, // Label 名稱
)

// httpRequestDuration HTTP 請求時間（Histogram）
// Histogram 適合：回應時間、請求大小分布
// Buckets 定義時間範圍（秒）：< 10ms, < 25ms, < 50ms, ..., < 10s
var httpRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP 請求處理時間（秒）",
		Buckets: prometheus.DefBuckets, // 預設 bucket：.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
	},
	[]string{"method", "path"},
)

// activeConnections 當前連線數（Gauge）
// Gauge 適合：當前值（可增可減）：連線數、佇列長度、快取大小
var activeConnections = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "http_active_connections",
		Help: "當前活躍的 HTTP 連線數",
	},
)

// dbQueryDuration 資料庫查詢時間（Histogram）
var dbQueryDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "db_query_duration_seconds",
		Help:    "資料庫查詢時間（秒）",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1}, // 自訂 bucket（ms 級別）
	},
	[]string{"operation", "table"},
)

// businessErrors 業務邏輯錯誤計數（Counter）
var businessErrors = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "business_errors_total",
		Help: "業務邏輯錯誤總數，按錯誤類型分類",
	},
	[]string{"error_type"},
)

// cacheHitRatio 快取命中數（Counter，用兩個 counter 計算比率）
// 注意：比率通常用兩個 counter 計算，不用 Gauge
// 在 Prometheus 查詢語言（PromQL）：rate(cache_hits[5m]) / rate(cache_requests[5m])
var cacheHits = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "cache_hits_total",
	Help: "快取命中次數",
})
var cacheMisses = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "cache_misses_total",
	Help: "快取未命中次數",
})

// ==========================================================================
// 2. 註冊 Metrics
// ==========================================================================

func init() {
	// 向 Prometheus 預設 Registry 註冊所有自訂 metrics
	// 如果名稱重複，這裡會 panic（在啟動時就能發現錯誤）
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		activeConnections,
		dbQueryDuration,
		businessErrors,
		cacheHits,
		cacheMisses,
	)
}

// ==========================================================================
// 3. Metrics 中介層（Middleware）
// ==========================================================================
//
// 把 metrics 收集包成 middleware，讓所有 handler 自動被監控
// 這是「橫切關注點」（Cross-cutting Concern）的標準做法

// metricsMiddleware 自動收集 HTTP metrics 的中介層
func metricsMiddleware(next http.HandlerFunc, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now() // 記錄開始時間

		activeConnections.Inc()       // Gauge +1（新連線進來）
		defer activeConnections.Dec() // Gauge -1（請求結束，defer 確保一定執行）

		// 包裝 ResponseWriter 以攔截狀態碼
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next(wrapped, r) // 執行實際的 handler

		duration := time.Since(start).Seconds() // 計算耗時

		// 記錄請求數（Counter）
		httpRequestsTotal.WithLabelValues(
			r.Method,
			path,
			fmt.Sprintf("%d", wrapped.statusCode),
		).Inc()

		// 記錄請求時間（Histogram）
		httpRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
	}
}

// responseWriter 包裝標準 ResponseWriter，用於攔截狀態碼
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// ==========================================================================
// 4. 模擬業務邏輯（用來產生 metrics 數據）
// ==========================================================================

// simulateDBQuery 模擬資料庫查詢（帶 metrics 記錄）
func simulateDBQuery(operation, table string) error {
	start := time.Now()

	// 模擬查詢耗時（50-200ms）
	time.Sleep(time.Duration(50+rand.Intn(150)) * time.Millisecond)

	// 模擬 5% 查詢失敗
	if rand.Float64() < 0.05 {
		businessErrors.WithLabelValues("db_error").Inc()
		return fmt.Errorf("資料庫查詢失敗: %s %s", operation, table)
	}

	// 記錄查詢時間
	dbQueryDuration.WithLabelValues(operation, table).Observe(time.Since(start).Seconds())
	return nil
}

// simulateCache 模擬快取查詢（帶 metrics 記錄）
func simulateCache() bool {
	// 模擬 70% 命中率
	if rand.Float64() < 0.7 {
		cacheHits.Inc()
		return true
	}
	cacheMisses.Inc()
	return false
}

// ==========================================================================
// 5. HTTP Handlers
// ==========================================================================

// postsHandler 文章列表 API（示範 metrics）
func postsHandler(w http.ResponseWriter, r *http.Request) {
	// 先查快取
	if hit := simulateCache(); !hit {
		// 快取未命中，查資料庫
		if err := simulateDBQuery("SELECT", "posts"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			businessErrors.WithLabelValues("api_error").Inc()
			return
		}
	}

	// 模擬慢請求（10% 機率）
	if rand.Float64() < 0.1 {
		time.Sleep(500 * time.Millisecond) // 模擬慢查詢
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"posts": [{"id": 1, "title": "Prometheus 監控教學"}]}`)
}

// createPostHandler 建立文章 API
func createPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		businessErrors.WithLabelValues("method_not_allowed").Inc()
		return
	}

	// 模擬資料庫寫入
	if err := simulateDBQuery("INSERT", "posts"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"id": 42, "status": "created"}`)
}

// demoHandler 示範頁面（用瀏覽器訪問）
func demoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Prometheus 監控示範</title></head>
<body>
<h1>第二十九課：Prometheus 監控</h1>

<h2>可用端點</h2>
<ul>
  <li><a href="/metrics">/metrics</a> — Prometheus 格式的所有 metrics（給 Prometheus Server 抓取）</li>
  <li><a href="/api/posts">/api/posts</a> — 模擬文章 API（會產生 metrics）</li>
  <li><a href="/demo">/demo</a> — 此頁面</li>
</ul>

<h2>如何使用 Prometheus</h2>
<pre>
# 1. 安裝 Prometheus（Docker 方式）
docker run -d -p 9090:9090 \
  -v ./prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus

# 2. prometheus.yml（最簡單的設定）
scrape_configs:
  - job_name: 'my-app'
    static_configs:
      - targets: ['host.docker.internal:8080']

# 3. 訪問 Prometheus UI：http://localhost:9090
# 4. 安裝 Grafana 做視覺化
</pre>

<h2>有用的 PromQL 查詢</h2>
<pre>
# 每秒請求數（5分鐘滾動平均）
rate(http_requests_total[5m])

# 錯誤率
rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])

# p95 回應時間
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# 快取命中率
rate(cache_hits_total[5m]) / (rate(cache_hits_total[5m]) + rate(cache_misses_total[5m]))
</pre>

<script>
// 每秒自動觸發 API 請求，產生 metrics 數據
setInterval(async () => {
  try { await fetch('/api/posts'); } catch(e) {}
}, 1000);
</script>
<p><em>（頁面每秒自動呼叫 /api/posts，在 /metrics 可以看到數字增加）</em></p>
</body>
</html>`)
}

// ==========================================================================
// 主程式
// ==========================================================================

func main() {
	fmt.Println("==========================================")
	fmt.Println(" 第二十九課：Prometheus 監控")
	fmt.Println("==========================================")
	fmt.Println()

	// ──── 說明四種 Metric 類型 ────
	fmt.Println("=== Prometheus 四種 Metric 類型 ===")
	fmt.Println()
	fmt.Println("1. Counter（計數器）")
	fmt.Println("   - 只增不減，重啟後歸零")
	fmt.Println("   - 用途：請求數、錯誤數、處理的訊息數")
	fmt.Println("   - PromQL：rate(http_requests_total[5m]) → 每秒請求數")
	fmt.Println()
	fmt.Println("2. Gauge（儀表盤）")
	fmt.Println("   - 可增可減，當前值")
	fmt.Println("   - 用途：連線數、記憶體用量、佇列長度")
	fmt.Println("   - PromQL：http_active_connections → 直接看當前值")
	fmt.Println()
	fmt.Println("3. Histogram（直方圖）")
	fmt.Println("   - 把觀測值分到預定義的 bucket")
	fmt.Println("   - 用途：回應時間分布、請求大小分布")
	fmt.Println("   - PromQL：histogram_quantile(0.95, rate(...bucket[5m])) → p95 延遲")
	fmt.Println()
	fmt.Println("4. Summary（摘要）")
	fmt.Println("   - 客戶端計算百分位數")
	fmt.Println("   - 用途：需要精確百分位數（但比 Histogram 佔更多資源）")
	fmt.Println("   - 現代做法：優先用 Histogram，用 PromQL 計算百分位")
	fmt.Println()

	// ──── 示範 Counter 和 Gauge 的基本用法 ────
	fmt.Println("=== 模擬業務操作，產生 Metrics ===")
	fmt.Println()

	// 模擬一些操作
	for i := 0; i < 5; i++ {
		if err := simulateDBQuery("SELECT", "posts"); err != nil {
			fmt.Printf("  查詢失敗: %v\n", err)
		} else {
			fmt.Printf("  查詢成功 #%d\n", i+1)
		}
	}

	hit := simulateCache()
	fmt.Printf("  快取查詢: %s\n", map[bool]string{true: "命中 ✅", false: "未命中 ❌"}[hit])
	fmt.Println()

	// ──── 啟動 HTTP 伺服器 ────
	fmt.Println("=== 啟動 HTTP 伺服器 ===")
	fmt.Println()
	fmt.Println("監控端點：")
	fmt.Println("  http://localhost:8080/metrics   ← Prometheus 格式的 metrics")
	fmt.Println("  http://localhost:8080/api/posts ← 業務 API（自動收集 metrics）")
	fmt.Println("  http://localhost:8080/demo      ← 示範頁面")
	fmt.Println()
	fmt.Println("在 /metrics 你會看到：")
	fmt.Println("  http_requests_total{method=\"GET\",path=\"/api/posts\",status=\"200\"} 42")
	fmt.Println("  http_request_duration_seconds_bucket{le=\"0.1\"} 38")
	fmt.Println("  cache_hits_total 29")
	fmt.Println()

	mux := http.NewServeMux()

	// /metrics 端點：Prometheus Server 會定期來這裡抓取數據
	mux.Handle("/metrics", promhttp.Handler())

	// 業務 API（用 middleware 包裝，自動收集 metrics）
	mux.HandleFunc("/api/posts", metricsMiddleware(postsHandler, "/api/posts"))
	mux.HandleFunc("/api/posts/create", metricsMiddleware(createPostHandler, "/api/posts/create"))
	mux.HandleFunc("/demo", metricsMiddleware(demoHandler, "/demo"))

	log.Println("伺服器啟動：http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
