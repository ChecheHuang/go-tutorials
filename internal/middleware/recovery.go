package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"blog-api/pkg/response"

	"github.com/gin-gonic/gin"
)

// Recovery 回傳 panic 恢復中介層
// 當 handler 發生 panic 時，攔截並回傳 500 錯誤，防止伺服器崩潰
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 記錄 panic 的堆疊追蹤資訊，方便除錯
				log.Printf("[Recovery] panic recovered: %v\n%s", err, debug.Stack())

				// 回傳 500 錯誤回應
				response.Error(c, http.StatusInternalServerError, "伺服器內部錯誤")
				c.Abort()
			}
		}()

		c.Next()
	}
}
