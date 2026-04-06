package middleware

// 教學對應：第 17 課（中介層 Logger）、第 21 課（結構化日誌）

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 回傳結構化日誌中介層
// 使用 slog 記錄每個請求的方法、路徑、狀態碼與回應時間
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path

		c.Next()

		slog.Info("request",
			"method", method,
			"path", path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}
