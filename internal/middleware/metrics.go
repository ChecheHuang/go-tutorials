package middleware

// 教學對應：第 35 課（Prometheus 監控）

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP 請求總數",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP 請求處理時間（秒）",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	httpActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_active_connections",
			Help: "目前正在處理的連線數",
		},
	)
)

// Metrics 回傳 Prometheus 指標收集中介層
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath() // 使用路由模板（如 /api/v1/articles/:id），避免高基數
		if path == "" {
			path = "unknown"
		}

		httpActiveConnections.Inc()
		defer httpActiveConnections.Dec()

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}
