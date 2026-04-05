package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 回傳自訂的日誌記錄中介層
// 記錄每個請求的方法、路徑、狀態碼與回應時間
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 記錄請求開始時間
		start := time.Now()

		// 取得請求的基本資訊
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		// 執行後續的 handler
		c.Next()

		// 計算回應時間
		latency := time.Since(start)

		// 取得回應狀態碼
		statusCode := c.Writer.Status()

		// 輸出格式化的日誌
		log.Printf("[API] %3d | %13v | %15s | %-7s %s",
			statusCode,  // HTTP 狀態碼
			latency,     // 回應時間
			clientIP,    // 客戶端 IP
			method,      // HTTP 方法
			path,        // 請求路徑
		)
	}
}
