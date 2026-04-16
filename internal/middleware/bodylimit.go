package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BodyLimit 回傳請求 body 大小限制中介層
// 防止惡意的超大 payload 導致記憶體耗盡（OOM）
//
// maxBytes 為允許的最大 body 大小（位元組），建議值：
//   - 一般 API：1 << 20（1MB）
//   - 檔案上傳：10 << 20（10MB）
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}

// DefaultBodyLimit 回傳 1MB 限制的中介層
func DefaultBodyLimit() gin.HandlerFunc {
	return BodyLimit(1 << 20) // 1MB
}
