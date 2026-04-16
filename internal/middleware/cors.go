package middleware

// 教學對應：第 17 課（中介層 CORS）

import (
	"slices"

	"github.com/gin-gonic/gin"
)

// CORS 回傳跨來源資源共享（Cross-Origin Resource Sharing）中介層
// allowedOrigins 為允許的來源白名單；空陣列或包含 "*" 時允許所有來源（僅限開發環境使用）
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowAll := len(allowedOrigins) == 0 || slices.Contains(allowedOrigins, "*")

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if allowAll {
			// 開發模式：允許所有來源
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin != "" && slices.Contains(allowedOrigins, origin) {
			// 生產模式：僅允許白名單中的來源，並回傳具體的 Origin（而非 *）
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
		} else {
			// 來源不在白名單中，不設定 CORS header，瀏覽器會自動拒絕
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(403)
				return
			}
			c.Next()
			return
		}

		// 允許的 HTTP 方法
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// 允許的請求標頭
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// 預檢請求的快取時間（秒）
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		// 處理 OPTIONS 預檢請求
		// 瀏覽器在發送跨域請求前，會先發送 OPTIONS 請求確認伺服器是否允許
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204) // 回傳 204 No Content
			return
		}

		c.Next()
	}
}
