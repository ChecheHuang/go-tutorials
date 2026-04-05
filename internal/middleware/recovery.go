package middleware

import (
	"log/slog"
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
				slog.Error("panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
				)
				response.Error(c, http.StatusInternalServerError, "伺服器內部錯誤")
				c.Abort()
			}
		}()

		c.Next()
	}
}
