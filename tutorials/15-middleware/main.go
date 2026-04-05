// 第十四課：中介層（Middleware）
// 中介層是在請求到達 Handler 之前/之後執行的程式碼
// 用於日誌記錄、認證、跨域設定等「每個請求都需要的」通用邏輯
//
// 執行方式：go run main.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ========================================
// 1. 中介層的本質
// ========================================
//
// 中介層就是一個 gin.HandlerFunc，和 handler 的型別完全一樣
// 唯一的差別是：中介層會呼叫 c.Next() 繼續執行後續的 handler
//
// 請求流程：
//   Client → Logger → Auth → Handler → Auth(後半) → Logger(後半) → Client
//
// 這個結構像洋蔥：外層中介層包裹內層，請求進去、回應出來

// ========================================
// 2. 自訂 Logger 中介層
// ========================================

// LoggerMiddleware 記錄每個請求的資訊
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ===== c.Next() 之前：請求進來時執行 =====
		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path

		log.Printf("[開始] %s %s", method, path)

		// 繼續執行後面的 handler 或中介層
		c.Next()

		// ===== c.Next() 之後：回應出去時執行 =====
		latency := time.Since(start)
		status := c.Writer.Status()

		log.Printf("[完成] %s %s → %d (%v)", method, path, status, latency)
	}
}

// ========================================
// 3. 認證中介層
// ========================================

// AuthMiddleware 檢查是否有合法的 API Key
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 從 Header 取得 API Key
		apiKey := c.GetHeader("X-API-Key")

		// 驗證 API Key
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "缺少 X-API-Key header",
			})
			c.Abort() // 中止！不執行後續的 handler
			return
		}

		if apiKey != "my-secret-key" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "無效的 API Key",
			})
			c.Abort()
			return
		}

		// 驗證通過，將使用者資訊存入 Context
		c.Set("user", "authenticated-user")

		// 繼續執行
		c.Next()
	}
}

// ========================================
// 4. CORS 中介層
// ========================================

// CORSMiddleware 處理跨域請求
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")

		// OPTIONS 預檢請求直接回傳
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// ========================================
// 5. Recovery 中介層（攔截 panic）
// ========================================

// RecoveryMiddleware 防止 panic 讓伺服器崩潰
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC] %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "伺服器內部錯誤",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}

// ========================================
// 6. 帶參數的中介層（閉包）
// ========================================

// RateLimiter 建立一個請求限流中介層
// maxRequests 是允許的最大請求數（簡化版本）
func RateLimiter(maxRequests int) gin.HandlerFunc {
	requestCount := 0 // 被閉包捕獲

	return func(c *gin.Context) {
		requestCount++

		if requestCount > maxRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": fmt.Sprintf("已超過限制（%d 次）", maxRequests),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode) // 減少 Gin 預設輸出

	r := gin.New() // 不用 Default()，自己加中介層

	// ========================================
	// 全域中介層（對所有路由生效）
	// ========================================
	r.Use(RecoveryMiddleware()) // 最外層：攔截 panic
	r.Use(LoggerMiddleware())   // 記錄所有請求
	r.Use(CORSMiddleware())     // 跨域設定

	// ========================================
	// 公開路由（不需要認證）
	// ========================================
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "歡迎！這是公開路由",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// ========================================
	// 受保護的路由群組（需要認證）
	// ========================================
	api := r.Group("/api")
	api.Use(AuthMiddleware()) // 只對 /api 路由群組生效
	{
		api.GET("/profile", func(c *gin.Context) {
			// 從中介層設定的 Context 中取得使用者資訊
			user, _ := c.Get("user")
			c.JSON(http.StatusOK, gin.H{
				"message": "這是受保護的路由",
				"user":    user,
			})
		})

		api.GET("/data", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"data": []string{"機密資料 A", "機密資料 B"},
			})
		})
	}

	// ========================================
	// 帶限流的路由
	// ========================================
	limited := r.Group("/limited")
	limited.Use(RateLimiter(3)) // 最多 3 次請求
	{
		limited.GET("/resource", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "限流資源"})
		})
	}

	// ========================================
	// 測試 panic recovery
	// ========================================
	r.GET("/panic", func(c *gin.Context) {
		panic("故意觸發 panic！")
	})

	// ========================================
	// 啟動
	// ========================================
	fmt.Println("伺服器啟動於 http://localhost:9090")
	fmt.Println()
	fmt.Println("測試指令：")
	fmt.Println("  公開路由:   curl http://localhost:9090/")
	fmt.Println("  無 Key:     curl http://localhost:9090/api/profile")
	fmt.Println("  有 Key:     curl -H 'X-API-Key: my-secret-key' http://localhost:9090/api/profile")
	fmt.Println("  Panic:      curl http://localhost:9090/panic")
	fmt.Println("  限流:       重複執行 curl http://localhost:9090/limited/resource 超過 3 次")

	r.Run(":9090")
}
