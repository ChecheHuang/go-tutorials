// =============================================================================
// 第十二課：Gin 框架入門 — Go 最受歡迎的 Web 框架
// =============================================================================
//
// 什麼是 Gin？
// Gin 是一個用 Go 語言寫的「第三方 Web 框架」（不是 Go 內建的）。
// 它幫你處理 HTTP 路由、JSON 解析、參數驗證、中介層等瑣碎工作，
// 讓你專注在業務邏輯上。
//
// 為什麼不用標準庫 net/http？
// 標準庫可以寫 Web 伺服器，但你需要自己處理：
//   - 路由分發（根據 Method + Path 找到對應的處理函式）
//   - JSON 序列化/反序列化
//   - 參數解析和驗證
//   - 錯誤處理和統一回應格式
// Gin 把這些都封裝好了，讓程式碼更簡潔、更不容易出錯。
//
// 什麼是 "github.com/gin-gonic/gin"？
// 這是 Gin 框架的「import path」（匯入路徑）。
// Go 的第三方套件用 GitHub 網址作為唯一識別，所以：
//   - github.com/gin-gonic/gin → GitHub 上 gin-gonic 組織的 gin 專案
//   - 這不代表每次 import 都會連網，go mod 會在本地快取
//   - 使用前要先執行：go get github.com/gin-gonic/gin
//
// 什麼是 gin.Context？（Gin 的「瑞士刀」）
// gin.Context 是 Gin 最核心的結構體，幾乎所有操作都透過它完成：
//   - 讀取請求：取得路由參數、查詢參數、請求 body、Header
//   - 寫入回應：回傳 JSON、設定狀態碼、設定 Header
//   - 流程控制：中止請求鏈（Abort）、繼續下一個 handler（Next）
//   - 資料傳遞：在中介層和 handler 之間傳遞資料（Set/Get）
//
// 執行方式：go run ./tutorials/12-gin-framework
// 注意：這個程式會啟動 HTTP 伺服器，按 Ctrl+C 可以停止伺服器。
// =============================================================================

package main // 每個可執行的 Go 程式都必須是 main 套件

import ( // import 區塊：匯入需要的套件
	"fmt"      // fmt：格式化輸出，用來印東西到終端機
	"net/http" // net/http：HTTP 狀態碼常數（如 200、201、400）
	"strconv"  // strconv：字串和其他型別的互相轉換

	"github.com/gin-gonic/gin" // gin：第三方 Web 框架，這個專案的核心
)

// =============================================================================
// 資料模型（Model）— 定義資料的「形狀」
// =============================================================================

// User 代表一個使用者的資料結構
// 結構體（struct）就像一個「資料模板」，定義了使用者有哪些欄位
type User struct {
	ID   int    `json:"id"`   // 使用者 ID — json:"id" 表示轉成 JSON 時欄位名叫 "id"
	Name string `json:"name"` // 使用者名稱 — json:"name" 表示 JSON 欄位名叫 "name"
	Age  int    `json:"age"`  // 使用者年齡 — json:"age" 表示 JSON 欄位名叫 "age"
}

// CreateUserRequest 定義「建立使用者」時客戶端必須送來的資料
// 這是一個「請求結構體」— 專門用來接收客戶端送來的 JSON
type CreateUserRequest struct {
	Name string `json:"name" binding:"required"`      // binding:"required" 表示這個欄位是必填的
	Age  int    `json:"age"  binding:"required,gt=0"` // binding:"gt=0" 表示年齡必須大於 0
}

// =============================================================================
// 模擬資料庫 — 用切片（slice）暫存資料，重啟後資料會消失
// =============================================================================

var users = []User{ // users 是一個 User 切片，存放所有使用者
	{ID: 1, Name: "Alice", Age: 25}, // 預設使用者 1：Alice
	{ID: 2, Name: "Bob", Age: 30},   // 預設使用者 2：Bob
}

var nextID = 3 // nextID 用來產生下一個使用者的 ID（模擬資料庫的自動遞增）

// =============================================================================
// 統一錯誤回應（Unified Error Response）
// =============================================================================
// 在真實專案中，我們希望所有錯誤都用相同的 JSON 格式回傳，
// 這樣前端只需要一種方式來處理錯誤。
//
// 格式：{"error": "錯誤訊息"}
//
// 我們的部落格專案（internal/handler/router.go）也使用這個模式。

// respondWithError 用統一的格式回傳錯誤 JSON
// 參數：c 是 gin.Context、code 是 HTTP 狀態碼、message 是錯誤訊息
func respondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{ // c.JSON() 回傳 JSON 格式的回應
		"error": message, // gin.H 是 map[string]any 的簡寫，方便快速建立 JSON
	})
}

// respondWithData 用統一的格式回傳成功的 JSON
// 參數：c 是 gin.Context、code 是 HTTP 狀態碼、data 是要回傳的資料
func respondWithData(c *gin.Context, code int, data interface{}) {
	c.JSON(code, gin.H{ // 成功時也用 gin.H 包裝
		"data": data, // 資料放在 "data" 欄位裡
	})
}

// =============================================================================
// main 函式 — 程式的進入點
// =============================================================================

func main() {
	// =====================================================================
	// 第一步：建立 Gin 引擎（Engine）
	// =====================================================================
	// gin.Default() 會建立一個帶有「預設中介層」的引擎：
	//   - Logger 中介層：自動記錄每個請求的日誌（方法、路徑、狀態碼、耗時）
	//   - Recovery 中介層：如果 handler 發生 panic，會自動恢復並回傳 500 錯誤
	// 如果不想要預設中介層，可以用 gin.New() 建立空白引擎
	r := gin.Default() // r 是 Router 的縮寫，代表路由器

	// =====================================================================
	// 第二步：基本路由（Route）— 將 URL 路徑對應到處理函式
	// =====================================================================
	// 路由格式：r.HTTP方法(路徑, 處理函式)
	// HTTP 方法代表不同的操作意圖：
	//   GET    → 讀取資料（不會改變伺服器上的資料）
	//   POST   → 建立新資料
	//   PUT    → 更新現有資料
	//   DELETE → 刪除資料

	// GET /ping — 健康檢查端點（用來確認伺服器是否正常運作）
	r.GET("/ping", func(c *gin.Context) { // c 是 *gin.Context，包含這次請求的所有資訊
		c.JSON(http.StatusOK, gin.H{ // http.StatusOK 就是 200
			"message": "pong", // 回傳一個簡單的 JSON：{"message": "pong"}
		})
	})

	// =====================================================================
	// 第三步：路由群組（Route Groups）— 把相關路由整理在一起
	// =====================================================================
	// 路由群組的好處：
	//   1. 共用路徑前綴：不用每個路由都寫 "/api/v1"
	//   2. 共用中介層：可以一次套用到整個群組（例如認證中介層）
	//   3. 程式碼更整齊：相關的路由放在一起，一目了然
	//
	// 在我們的部落格專案中（internal/handler/router.go），
	// 路由群組是這樣用的：
	//   v1 := r.Group("/api/v1")          → API 版本前綴
	//   auth := v1.Group("/auth")         → 認證相關路由
	//   authenticated := v1.Group("")     → 需要登入的路由
	//   authenticated.Use(middleware.JWTAuth(cfg)) → 套用 JWT 認證中介層

	v1 := r.Group("/api/v1") // 建立 /api/v1 路由群組
	{                        // 大括號只是為了視覺分組，Go 語法上不是必須的
		// ----- 使用者相關路由 -----

		// GET /api/v1/users → 取得所有使用者列表
		v1.GET("/users", getUsers)

		// GET /api/v1/users/:id → 取得特定使用者
		// :id 是「路由參數」（Path Parameter），冒號表示這是一個變數
		// 例如 /api/v1/users/1 中的 "1" 會被捕捉為 id 參數
		v1.GET("/users/:id", getUserByID)

		// POST /api/v1/users → 建立新使用者
		// 客戶端要在 Request Body 中送 JSON 資料
		v1.POST("/users", createUser)

		// PUT /api/v1/users/:id → 更新指定使用者的資料
		v1.PUT("/users/:id", updateUser)

		// DELETE /api/v1/users/:id → 刪除指定使用者
		v1.DELETE("/users/:id", deleteUser)

		// ----- 搜尋相關路由 -----

		// GET /api/v1/search?keyword=alice&page=1 → 搜尋使用者
		// keyword 和 page 是「查詢參數」（Query Parameter），用 ? 開頭、& 分隔
		v1.GET("/search", searchUsers)
	}

	// =====================================================================
	// 路由參數 vs 查詢參數
	// =====================================================================
	//
	// 路由參數（Path Parameter）— 用來「識別特定資源」
	//   路徑定義：/users/:id
	//   實際請求：/users/42
	//   取得方式：c.Param("id") → "42"
	//   適用場景：取得、更新、刪除「某一個」資源
	//
	// 查詢參數（Query Parameter）— 用來「篩選、排序、分頁」
	//   實際請求：/search?keyword=alice&page=2
	//   取得方式：c.Query("keyword") → "alice"
	//            c.DefaultQuery("page", "1") → "2"（如果沒給就用預設值 "1"）
	//   適用場景：搜尋、列表篩選、分頁

	// =====================================================================
	// 第四步：啟動 HTTP 伺服器
	// =====================================================================
	fmt.Println("=== Gin 框架教學伺服器 ===")                                                                                                        // 印出歡迎訊息
	fmt.Println("伺服器啟動在 http://localhost:9090")                                                                                               // 告訴使用者網址
	fmt.Println("按 Ctrl+C 可以停止伺服器")                                                                                                           // 提醒使用者如何停止
	fmt.Println()                                                                                                                             // 印出空行
	fmt.Println("試試這些指令：")                                                                                                                    // 提供測試指令
	fmt.Println("  curl http://localhost:9090/ping")                                                                                          // 健康檢查
	fmt.Println("  curl http://localhost:9090/api/v1/users")                                                                                  // 取得所有使用者
	fmt.Println("  curl http://localhost:9090/api/v1/users/1")                                                                                // 取得特定使用者
	fmt.Println("  curl -X POST http://localhost:9090/api/v1/users -H 'Content-Type: application/json' -d '{\"name\":\"Carol\",\"age\":22}'") // 建立使用者
	fmt.Println("  curl http://localhost:9090/api/v1/search?keyword=alice&page=1")                                                            // 搜尋
	fmt.Println()                                                                                                                             // 印出空行

	r.Run(":9090") // 啟動伺服器，監聽 9090 埠口（預設是 :8080）
	// 注意：Run() 會阻塞（blocking），程式會停在這裡直到你按 Ctrl+C
}

// =============================================================================
// Handler 函式 — 處理每個路由的業務邏輯
// =============================================================================
// 每個 handler 都接收一個 *gin.Context 參數
// gin.Context 是 Gin 的核心，幾乎所有操作都透過它完成：
//   - 讀取請求：c.Param()、c.Query()、c.ShouldBindJSON()
//   - 寫入回應：c.JSON()、c.String()、c.HTML()
//   - 流程控制：c.Abort()（中止）、c.Next()（繼續）
//   - 資料傳遞：c.Set()、c.Get()（在中介層和 handler 間共享資料）

// getUsers 取得所有使用者列表
// 對應路由：GET /api/v1/users
func getUsers(c *gin.Context) {
	respondWithData(c, http.StatusOK, users) // 回傳 200 + 所有使用者資料
}

// getUserByID 根據 ID 取得特定使用者
// 對應路由：GET /api/v1/users/:id
func getUserByID(c *gin.Context) {
	// c.Param("id") 取得路由參數 :id 的值（型別是 string）
	idStr := c.Param("id") // 例如 /users/1 → idStr = "1"

	// strconv.Atoi 把字串轉成整數（Atoi = ASCII to Integer）
	id, err := strconv.Atoi(idStr) // "1" → 1，如果不是數字會回傳 error
	if err != nil {                // 如果轉換失敗（例如 /users/abc）
		respondWithError(c, http.StatusBadRequest, "ID 必須是數字") // 回傳 400 錯誤
		return                                                 // 提早返回，不再往下執行
	}

	// 在切片中尋找符合 ID 的使用者
	for _, user := range users { // range 遍歷 users 切片
		if user.ID == id { // 找到了！
			respondWithData(c, http.StatusOK, user) // 回傳 200 + 使用者資料
			return                                  // 找到就提早返回
		}
	}

	// 跑完整個迴圈都沒找到
	respondWithError(c, http.StatusNotFound, "使用者不存在") // 回傳 404
}

// createUser 建立新使用者
// 對應路由：POST /api/v1/users
func createUser(c *gin.Context) {
	var req CreateUserRequest // 宣告一個請求結構體變數

	// c.ShouldBindJSON(&req) 做兩件事：
	//   1. 從 Request Body 解析 JSON 到 req 結構體
	//   2. 根據 binding 標籤驗證資料（例如 required、gt=0）
	// 如果 JSON 格式錯誤或驗證不通過，會回傳 error
	if err := c.ShouldBindJSON(&req); err != nil { // & 取得指標，讓函式能修改 req 的值
		respondWithError(c, http.StatusBadRequest, "無效的請求: "+err.Error()) // 回傳 400 + 錯誤詳情
		return                                                            // 提早返回
	}

	// 驗證通過，建立新使用者
	newUser := User{ // 建立 User 結構體
		ID:   nextID,   // 分配 ID
		Name: req.Name, // 從請求中取得名稱
		Age:  req.Age,  // 從請求中取得年齡
	}
	nextID++                       // ID 遞增，下一個使用者會用新的 ID
	users = append(users, newUser) // 把新使用者加入切片

	// 回傳 201 Created（表示成功建立了一個新資源）
	respondWithData(c, http.StatusCreated, newUser)
}

// updateUser 更新使用者資料
// 對應路由：PUT /api/v1/users/:id
func updateUser(c *gin.Context) {
	idStr := c.Param("id")         // 取得路由參數 :id
	id, err := strconv.Atoi(idStr) // 字串轉整數
	if err != nil {                // 如果不是數字
		respondWithError(c, http.StatusBadRequest, "ID 必須是數字") // 回傳 400
		return                                                 // 提早返回
	}

	var req CreateUserRequest                      // 用同一個請求結構體接收更新資料
	if err := c.ShouldBindJSON(&req); err != nil { // 解析 + 驗證 JSON
		respondWithError(c, http.StatusBadRequest, "無效的請求: "+err.Error()) // 回傳 400
		return                                                            // 提早返回
	}

	// 尋找並更新使用者
	for i, user := range users { // 遍歷所有使用者
		if user.ID == id { // 找到目標使用者
			users[i].Name = req.Name                    // 更新名稱
			users[i].Age = req.Age                      // 更新年齡
			respondWithData(c, http.StatusOK, users[i]) // 回傳 200 + 更新後的資料
			return                                      // 提早返回
		}
	}

	respondWithError(c, http.StatusNotFound, "使用者不存在") // 沒找到，回傳 404
}

// deleteUser 刪除使用者
// 對應路由：DELETE /api/v1/users/:id
func deleteUser(c *gin.Context) {
	idStr := c.Param("id")         // 取得路由參數 :id
	id, err := strconv.Atoi(idStr) // 字串轉整數
	if err != nil {                // 如果不是數字
		respondWithError(c, http.StatusBadRequest, "ID 必須是數字") // 回傳 400
		return                                                 // 提早返回
	}

	// 尋找並刪除使用者
	for i, user := range users { // 遍歷所有使用者
		if user.ID == id { // 找到目標使用者
			// append(users[:i], users[i+1:]...) 是 Go 刪除切片元素的慣用寫法
			// users[:i] → 索引 0 到 i-1 的元素（不包含 i）
			// users[i+1:] → 索引 i+1 到最後的元素
			// 兩段接在一起，就跳過了索引 i 的元素
			users = append(users[:i], users[i+1:]...) // 從切片中移除使用者
			c.JSON(http.StatusOK, gin.H{              // 回傳 200
				"message": "使用者已刪除", // 告訴客戶端刪除成功
			})
			return // 提早返回
		}
	}

	respondWithError(c, http.StatusNotFound, "使用者不存在") // 沒找到，回傳 404
}

// searchUsers 搜尋使用者（示範查詢參數的用法）
// 對應路由：GET /api/v1/search?keyword=alice&page=1
func searchUsers(c *gin.Context) {
	// c.Query("key") 取得查詢參數的值
	// 如果參數不存在，回傳空字串 ""
	keyword := c.Query("keyword") // 取得 keyword 參數

	// c.DefaultQuery("key", "default") 取得查詢參數，如果不存在就用預設值
	page := c.DefaultQuery("page", "1") // 取得 page 參數，預設為 "1"

	// 簡單的搜尋邏輯：篩選名字包含關鍵字的使用者
	var results []User           // 宣告結果切片
	for _, user := range users { // 遍歷所有使用者
		if keyword == "" || containsIgnoreCase(user.Name, keyword) { // 空關鍵字則回傳全部
			results = append(results, user) // 符合條件的使用者加入結果
		}
	}

	// 回傳搜尋結果
	c.JSON(http.StatusOK, gin.H{ // 回傳 200
		"keyword": keyword,      // 回傳搜尋關鍵字（讓客戶端知道搜了什麼）
		"page":    page,         // 回傳頁碼
		"total":   len(results), // 回傳結果總數
		"data":    results,      // 回傳搜尋結果
	})
}

// containsIgnoreCase 檢查字串 s 是否包含子字串 substr（不區分大小寫）
func containsIgnoreCase(s, substr string) bool {
	// 把兩個字串都轉成小寫後再比較
	// 這是一個簡化的實作，實際專案中可以用 strings.Contains + strings.ToLower
	sLower := toLower(s)        // 將 s 轉為小寫
	subLower := toLower(substr) // 將 substr 轉為小寫
	// 用 Go 標準庫的方式檢查是否包含
	for i := 0; i <= len(sLower)-len(subLower); i++ { // 逐位置比較
		if sLower[i:i+len(subLower)] == subLower { // 截取子字串比較
			return true // 找到了
		}
	}
	return false // 沒找到
}

// toLower 把字串轉為小寫（簡易版，只處理 A-Z）
func toLower(s string) string {
	result := make([]byte, len(s)) // 建立一個同長度的 byte 切片
	for i := 0; i < len(s); i++ {  // 遍歷每個字元
		if s[i] >= 'A' && s[i] <= 'Z' { // 如果是大寫字母
			result[i] = s[i] + 32 // ASCII 中大寫轉小寫就是 +32
		} else {
			result[i] = s[i] // 其他字元不變
		}
	}
	return string(result) // 轉回字串並回傳
}
