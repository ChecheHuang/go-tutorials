// ==========================================================================
// 第十五課：中介層（Middleware）
// ==========================================================================
//
// 什麼是中介層（Middleware）？
//   中介層就是「夾在請求和處理器之間的程式碼」
//   就像進電影院的安檢程序：
//     你（客戶端）→ 安檢（Middleware）→ 售票員（Handler）
//
//   安檢可以做很多事：
//     - 記錄你幾點進來（Logger）
//     - 檢查你有沒有票（Auth）
//     - 限制每小時最多進幾次（RateLimiter）
//     - 如果你突然暈倒，叫救護車（Recovery）
//
// 中介層的英文 Middleware 字面意思是「中間的軟體」
// 它讓你把「每個請求都需要執行」的邏輯抽出來，不用每個 Handler 都重複寫
//
// 執行方式：go run ./tutorials/15-middleware
// ==========================================================================

package main // 宣告這是 main 套件（可執行程式）

import (          // 匯入需要的套件
	"fmt"         // 標準庫：格式化輸出
	"log"         // 標準庫：日誌記錄（帶時間戳）
	"net/http"    // 標準庫：HTTP 狀態碼常數
	"time"        // 標準庫：時間相關功能

	"github.com/gin-gonic/gin" // 第三方套件：Gin Web 框架
	// gin.HandlerFunc 型別就是 func(*gin.Context)
	// 中介層和 Handler 的型別完全一樣！差別只在行為上
)

// ==========================================================================
// 中介層的核心概念：洋蔥模型
// ==========================================================================
//
// 請求流程就像剝洋蔥（由外到內，再由內到外）：
//
//   ┌────── Recovery ──────────────────────┐
//   │ ┌──── Logger ─────────────────────┐  │
//   │ │ ┌── Auth ──────────────────┐    │  │
//   │ │ │                          │    │  │
//   │ │ │   Handler（核心處理）     │    │  │
//   │ │ │                          │    │  │
//   │ │ └──────────────────────────┘    │  │
//   │ └────────────────────────────────-┘  │
//   └──────────────────────────────────────┘
//
// 請求進來：Recovery → Logger → Auth → Handler
// 回應出去：Handler → Auth（後半）→ Logger（後半）→ Recovery（後半）
//
// 這就是為什麼 c.Next() 之前的程式碼在「請求進來時」執行
// c.Next() 之後的程式碼在「回應出去時」執行

// ==========================================================================
// 1. Logger 中介層：記錄每個請求
// ==========================================================================

// LoggerMiddleware 回傳一個 gin.HandlerFunc（中介層函式）
// 這種「回傳函式的函式」叫做「工廠函式」，方便你傳入設定參數
func LoggerMiddleware() gin.HandlerFunc { // 函式回傳型別是 gin.HandlerFunc
	return func(c *gin.Context) { // 回傳真正的中介層函式，c 是每個請求的上下文

		// ===== c.Next() 之前：請求「進來」時執行 =====

		start := time.Now()             // 記錄請求開始時間（用來計算耗時）
		method := c.Request.Method      // 取得 HTTP 方法（GET、POST、PUT 等）
		path := c.Request.URL.Path      // 取得請求路徑（如 /api/profile）

		log.Printf("[開始] %s %s", method, path) // 印出請求開始的日誌

		// c.Next() 是關鍵：呼叫後，控制權交給下一個中介層或 Handler
		// 等下一個中介層/Handler 執行完，控制權再回到這裡繼續執行
		c.Next() // 繼續執行後面的 handler 或中介層

		// ===== c.Next() 之後：回應「出去」時執行 =====

		latency := time.Since(start) // 計算耗時 = 當前時間 - 開始時間
		status := c.Writer.Status()  // 取得 HTTP 回應狀態碼（如 200、404）

		log.Printf("[完成] %s %s → %d (%v)", method, path, status, latency) // 印出完整的請求日誌
	}
}

// ==========================================================================
// 2. Auth 中介層：驗證身份
// ==========================================================================

// AuthMiddleware 檢查請求是否帶有合法的 API Key
// 在真實專案中，這裡通常是驗證 JWT Token
func AuthMiddleware() gin.HandlerFunc { // 回傳 gin.HandlerFunc
	return func(c *gin.Context) { // 中介層函式

		// c.GetHeader 從 HTTP 請求的 Header 中取得指定的值
		// Header 就像信封上的標籤，攜帶請求的額外資訊
		apiKey := c.GetHeader("X-API-Key") // 取得名為 X-API-Key 的 Header 值

		// 情況 1：完全沒有帶 API Key
		if apiKey == "" { // 如果 API Key 是空字串
			c.JSON(http.StatusUnauthorized, gin.H{ // 回傳 401 Unauthorized
				"error": "缺少 X-API-Key header", // 錯誤訊息
			})
			c.Abort() // 中止！不執行後面的中介層和 Handler
			return    // 函式提前返回
		}

		// 情況 2：API Key 格式有但內容不對
		if apiKey != "my-secret-key" { // 如果 API Key 不等於預設的密鑰
			c.JSON(http.StatusUnauthorized, gin.H{ // 回傳 401 Unauthorized
				"error": "無效的 API Key", // 錯誤訊息
			})
			c.Abort() // 中止！不執行後面的 handler
			return    // 函式提前返回
		}

		// 情況 3：API Key 正確，通過驗證
		// c.Set 可以在 Context 中儲存任何值，讓後面的 Handler 使用
		// 就像在「購物車」裡放東西，Handler 可以從 Context 取出來
		c.Set("user", "authenticated-user") // 把使用者資訊存入 Context

		c.Next() // 繼續執行後面的 handler（驗證通過了！）
	}
}

// ==========================================================================
// 3. CORS 中介層：處理跨域請求
// ==========================================================================
//
// 什麼是 CORS（Cross-Origin Resource Sharing）？
//   瀏覽器的安全限制：網頁只能請求「同一個網站」的 API
//   例如：https://myapp.com 的網頁，預設不能呼叫 https://api.com 的 API
//   CORS 是讓伺服器「明確允許」跨域請求的機制
//
//   類比：就像大樓門禁，你需要被「白名單」才能進去

// CORSMiddleware 設定允許跨域請求的 HTTP Header
func CORSMiddleware() gin.HandlerFunc { // 回傳 gin.HandlerFunc
	return func(c *gin.Context) { // 中介層函式

		// c.Writer.Header().Set 設定 HTTP 回應的 Header
		// Header 是回傳給瀏覽器的「說明書」，告訴瀏覽器允許什麼

		// Access-Control-Allow-Origin: * 表示允許所有網域存取（開發用）
		// 正式環境應該設定為指定的網域，例如 "https://myapp.com"
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // 允許所有來源

		// Access-Control-Allow-Methods 指定允許的 HTTP 方法
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE") // 允許這些方法

		// Access-Control-Allow-Headers 指定允許的請求 Header
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key") // 允許這些 Header

		// OPTIONS 是瀏覽器在跨域請求前先發出的「預檢請求」（Preflight）
		// 瀏覽器想問：「我可以用 POST 方法請求你嗎？」
		if c.Request.Method == "OPTIONS" { // 如果是 OPTIONS 方法
			c.AbortWithStatus(204) // 回傳 204 No Content，表示允許（沒有回應內容）
			return                 // 函式提前返回
		}

		c.Next() // 繼續處理非 OPTIONS 的請求
	}
}

// ==========================================================================
// 4. Recovery 中介層：攔截 panic 防止伺服器崩潰
// ==========================================================================
//
// 什麼是 panic？
//   panic 是 Go 程式發生嚴重錯誤時的緊急停止機制
//   如果不處理，整個伺服器程式就會崩潰！
//
// Recovery 中介層使用 recover() 函式「接住」panic
// 就像在地板上鋪了軟墊——就算你摔跤，也不會受重傷

// RecoveryMiddleware 防止 panic 讓整個伺服器崩潰
func RecoveryMiddleware() gin.HandlerFunc { // 回傳 gin.HandlerFunc
	return func(c *gin.Context) { // 中介層函式

		// defer 讓這個函式在「離開 RecoveryMiddleware 時」才執行
		// 包括正常離開和 panic 離開都會執行
		defer func() { // 延遲執行的匿名函式
			if err := recover(); err != nil { // recover() 接住 panic，如果有 panic 就回傳它的值
				log.Printf("[PANIC] %v", err)             // 把 panic 的內容印到日誌
				c.JSON(http.StatusInternalServerError, gin.H{ // 回傳 500 Internal Server Error
					"error": "伺服器內部錯誤", // 對外只說「內部錯誤」，不洩漏細節
				})
				c.Abort() // 中止後續的中介層和 handler
			}
		}() // 立即設定好 defer，但實際執行要等函式返回時

		c.Next() // 繼續執行，如果後面的 handler 發生 panic，defer 裡的 recover 會接住
	}
}

// ==========================================================================
// 5. 帶參數的中介層：閉包的威力
// ==========================================================================
//
// 閉包（Closure）：函式可以「記住」建立它時的環境變數
// 就像你設一個鬧鐘，鬧鐘「記住」你設定的時間
//
// 這裡 RateLimiter(3) 建立了一個「記住 maxRequests=3」的中介層
// 每次有請求進來，requestCount 就加 1，超過就拒絕

// RateLimiter 建立一個限流中介層
// 參數 maxRequests：允許的最大請求次數
// 這是一個極簡化的限流器，實際生產環境需要用 Redis 等工具
func RateLimiter(maxRequests int) gin.HandlerFunc { // 接受最大請求數參數，回傳中介層
	requestCount := 0 // 計數器：被閉包捕獲，所有請求共享這個變數

	return func(c *gin.Context) { // 回傳中介層函式（閉包）
		requestCount++ // 每次請求進來，計數器加 1

		if requestCount > maxRequests { // 如果超過最大允許次數
			c.JSON(http.StatusTooManyRequests, gin.H{ // 回傳 429 Too Many Requests
				"error": fmt.Sprintf("已超過限制（%d 次）", maxRequests), // 告知已超過限制
			})
			c.Abort() // 中止，不執行後面的 handler
			return    // 函式提前返回
		}

		c.Next() // 次數還在限制內，繼續執行
	}
}

// ==========================================================================
// 主程式：組裝路由和中介層
// ==========================================================================

func main() { // 程式進入點
	gin.SetMode(gin.ReleaseMode) // 設定 Gin 為正式模式（減少 Gin 自己的輸出）

	// gin.New() 建立一個「乾淨的」Gin 路由器（沒有任何預設中介層）
	// 相比 gin.Default()（預設帶 Logger 和 Recovery），我們自己加中介層更靈活
	r := gin.New() // 建立不帶任何中介層的 Gin 路由器

	// ========================================
	// 全域中介層：對「所有路由」都生效
	// ========================================
	// r.Use 用來加入全域中介層，注意順序！
	// 最外層的中介層最先執行，所以 Recovery 要放最外面

	r.Use(RecoveryMiddleware()) // 最外層：攔截任何 panic，防止伺服器崩潰
	r.Use(LoggerMiddleware())   // 記錄所有請求（放在 Recovery 之後，方便在 panic 時也能記錄）
	r.Use(CORSMiddleware())     // 跨域設定（讓前端可以呼叫 API）

	// ========================================
	// 公開路由：不需要 API Key 就能存取
	// ========================================

	r.GET("/", func(c *gin.Context) { // 定義根路由 GET /
		c.JSON(http.StatusOK, gin.H{ // 回傳 200 OK 和 JSON 資料
			"message": "歡迎！這是公開路由", // 回應訊息
		})
	})

	r.GET("/health", func(c *gin.Context) { // 定義健康檢查路由（常用於監控）
		c.JSON(http.StatusOK, gin.H{"status": "ok"}) // 回傳服務狀態
	})

	// ========================================
	// 受保護的路由群組：需要 API Key 才能存取
	// ========================================
	// r.Group 建立路由群組，群組內的路由都有相同的前綴
	// 群組中介層只對「這個群組」的路由生效

	api := r.Group("/api")       // 建立 /api 路由群組
	api.Use(AuthMiddleware())    // 只對 /api 群組的路由做身份驗證
	{                            // 大括號只是視覺分組，不是語法要求
		api.GET("/profile", func(c *gin.Context) { // 定義 GET /api/profile
			// c.Get 從 Context 取出先前 c.Set 存入的值
			// AuthMiddleware 在驗證通過後把 "user" 存入 Context
			user, _ := c.Get("user") // 取出使用者資訊（_ 忽略第二個回傳值 exists）
			c.JSON(http.StatusOK, gin.H{ // 回傳 200 OK
				"message": "這是受保護的路由",  // 說明這是受保護的
				"user":    user,             // 把使用者資訊一起回傳
			})
		})

		api.GET("/data", func(c *gin.Context) { // 定義 GET /api/data
			c.JSON(http.StatusOK, gin.H{ // 回傳 200 OK
				"data": []string{"機密資料 A", "機密資料 B"}, // 模擬受保護的資料
			})
		})
	}

	// ========================================
	// 帶限流的路由群組：限制請求次數
	// ========================================

	limited := r.Group("/limited")   // 建立 /limited 路由群組
	limited.Use(RateLimiter(3))      // 這個群組最多只允許 3 次請求
	{
		limited.GET("/resource", func(c *gin.Context) { // 定義 GET /limited/resource
			c.JSON(http.StatusOK, gin.H{"message": "限流資源"}) // 在限流內的回應
		})
	}

	// ========================================
	// 測試 panic recovery
	// ========================================

	r.GET("/panic", func(c *gin.Context) { // 定義一個會 panic 的路由（用於測試）
		panic("故意觸發 panic！") // 觸發 panic，RecoveryMiddleware 會接住它
	})

	// ========================================
	// 啟動伺服器
	// ========================================

	fmt.Println("伺服器啟動於 http://localhost:9090") // 印出伺服器位址
	fmt.Println()                                     // 空行
	fmt.Println("測試指令：")                          // 印出測試說明

	// 以下是 curl 指令，可以在終端機測試各個路由
	fmt.Println("  公開路由: curl http://localhost:9090/")                                              // 不需要 Key
	fmt.Println("  無 Key:   curl http://localhost:9090/api/profile")                                  // 回傳 401
	fmt.Println("  有 Key:   curl -H 'X-API-Key: my-secret-key' http://localhost:9090/api/profile")   // 回傳 200
	fmt.Println("  Panic:    curl http://localhost:9090/panic")                                        // 測試 Recovery
	fmt.Println("  限流:     重複執行 curl http://localhost:9090/limited/resource 超過 3 次")           // 測試 RateLimiter

	r.Run(":9090") // 啟動伺服器，監聽 9090 埠（注意：會一直執行，直到按 Ctrl+C）
}
