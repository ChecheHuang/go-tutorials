package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORS 回傳跨來源資源共享（Cross-Origin Resource Sharing）中介層
// 允許前端應用從不同的網域存取此 API
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 允許所有來源（生產環境應設定為特定網域）
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

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
