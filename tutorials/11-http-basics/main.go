// ==========================================================================
// 第十一課：HTTP 基礎
// ==========================================================================
//
// 在學 Gin 框架之前，先用 Go 標準庫理解 HTTP 的運作原理
//
// 執行方式：go run ./tutorials/11-http-basics
// 然後在瀏覽器開啟 http://localhost:9090
//
// ╔══════════════════════════════════════════════════════════════════════╗
// ║                                                                    ║
// ║   HTTP 是什麼？— 用「寄信」來理解                                    ║
// ║                                                                    ║
// ║   HTTP（HyperText Transfer Protocol）= 超文本傳輸協定               ║
// ║   它是瀏覽器和伺服器之間溝通的「語言」                                ║
// ║                                                                    ║
// ║   想像你寄一封信：                                                   ║
// ║                                                                    ║
// ║   📨 請求（Request）= 你寄出的信                                     ║
// ║   ├─ 信封上寫著：寄去哪裡（URL）、做什麼事（Method）                  ║
// ║   ├─ 信封資訊：附加說明（Headers）                                   ║
// ║   └─ 信的內容：你要傳的資料（Body）                                  ║
// ║                                                                    ║
// ║   📬 回應（Response）= 對方寄回的回信                                ║
// ║   ├─ 信封上寫著：處理結果（Status Code）                             ║
// ║   ├─ 信封資訊：附加說明（Headers）                                   ║
// ║   └─ 回信內容：回傳的資料（Body）                                    ║
// ║                                                                    ║
// ╠══════════════════════════════════════════════════════════════════════╣
// ║                                                                    ║
// ║   伺服器（Server）是什麼？                                           ║
// ║   就是一台「一直在等信的郵局」                                        ║
// ║   它 24 小時待命，只要收到信（請求），就會處理並寄回回信（回應）         ║
// ║   我們這個程式就是一台伺服器，在 localhost:9090 等待請求               ║
// ║                                                                    ║
// ╠══════════════════════════════════════════════════════════════════════╣
// ║                                                                    ║
// ║   HTTP 方法（Method）— 信上寫的「目的」                               ║
// ║                                                                    ║
// ║   GET    = 「我要看東西」（看菜單）                                   ║
// ║            用於取得資料，不會改變伺服器的資料                          ║
// ║   POST   = 「我要新增東西」（點一道新菜）                             ║
// ║            用於建立新資料，會在伺服器新增一筆記錄                      ║
// ║   PUT    = 「我要修改東西」（改菜單上的菜名）                         ║
// ║            用於更新現有資料                                          ║
// ║   DELETE = 「我要刪除東西」（取消一道菜）                             ║
// ║            用於刪除現有資料                                          ║
// ║                                                                    ║
// ╠══════════════════════════════════════════════════════════════════════╣
// ║                                                                    ║
// ║   HTTP 狀態碼（Status Code）— 回信上的「處理結果」                    ║
// ║                                                                    ║
// ║   200 OK              = 成功！你要的東西在這裡                        ║
// ║   201 Created         = 成功！已經幫你建立好了                        ║
// ║   400 Bad Request     = 你的信寫錯了，我看不懂                        ║
// ║   404 Not Found       = 你要的東西不存在                              ║
// ║   405 Method Not Allowed = 這個地址不接受你的請求方式                  ║
// ║   500 Internal Error  = 郵局內部出問題了，不是你的錯                   ║
// ║                                                                    ║
// ╠══════════════════════════════════════════════════════════════════════╣
// ║                                                                    ║
// ║   JSON 是什麼？— API 的「共通語言」                                   ║
// ║                                                                    ║
// ║   JSON = JavaScript Object Notation                                ║
// ║   它是一種資料格式，長得像這樣：                                      ║
// ║   {"name": "Alice", "age": 25}                                     ║
// ║                                                                    ║
// ║   為什麼 API 都用 JSON？                                             ║
// ║   1. 人類讀得懂（不像二進位亂碼）                                     ║
// ║   2. 機器也讀得懂（所有程式語言都能解析）                              ║
// ║   3. 結構簡單明確（key-value 配對）                                   ║
// ║   4. 是目前 Web API 的事實標準                                       ║
// ║                                                                    ║
// ╚══════════════════════════════════════════════════════════════════════╝
package main // 宣告這是主程式套件

import (                // 匯入需要的標準函式庫
	"encoding/json" // encoding/json：用來將 Go 物件轉成 JSON，或將 JSON 轉成 Go 物件
	"fmt"           // fmt：格式化輸出，用來印文字到終端機或回應中
	"log"           // log：記錄日誌，這裡用來在伺服器啟動失敗時印出錯誤
	"net/http"      // net/http：Go 標準庫的 HTTP 套件，提供伺服器和客戶端功能
)

// ========================================
// 資料結構定義
// ========================================

// User 使用者結構體 — 定義一個使用者有哪些欄位
// json:"..." 是「結構體標籤」（struct tag），告訴 JSON 編碼器用什麼名稱
// 例如 json:"id" 表示轉成 JSON 時，欄位名稱會是 "id" 而不是 "ID"
type User struct {           // 定義 User 結構體
	ID   int    `json:"id"`   // 使用者 ID，JSON 中顯示為 "id"
	Name string `json:"name"` // 使用者名稱，JSON 中顯示為 "name"
	Age  int    `json:"age"`  // 使用者年齡，JSON 中顯示為 "age"
}

// 模擬的使用者資料庫（用切片代替真實資料庫）
// 在真實專案中，這些資料會存在 SQLite 或 PostgreSQL 等資料庫裡
var users = []User{                        // 宣告一個 User 切片並初始化
	{ID: 1, Name: "Alice", Age: 25}, // 第一個使用者
	{ID: 2, Name: "Bob", Age: 30},   // 第二個使用者
	{ID: 3, Name: "Carol", Age: 28}, // 第三個使用者
}

// ========================================
// Handler 函式
// ========================================
//
// Handler 函式是什麼？
// 就是「郵局裡負責處理信件的員工」
// 每當有請求（信件）送到特定的路徑（地址），對應的 Handler 就會處理它
//
// 每個 Handler 都接收兩個參數：
// - w http.ResponseWriter：用來寫回應（回信的信紙）
// - r *http.Request：包含請求的所有資訊（收到的信）

// helloHandler 最簡單的 Handler — 收到任何請求都回一句問候
func helloHandler(w http.ResponseWriter, r *http.Request) {
	// w = ResponseWriter：這是你的「回信信紙」，往上面寫什麼，客戶端就收到什麼
	// r = Request：這是客戶端寄來的「信」，包含方法、路徑、標頭、內容等所有資訊

	// r.Method 是 HTTP 方法（GET、POST 等），r.URL.Path 是請求的路徑
	// Fprintf 將格式化的字串寫入 w（回應給客戶端）
	fmt.Fprintf(w, "Hello! 你的請求方法是 %s，路徑是 %s", r.Method, r.URL.Path)
}

// usersHandler 處理 /api/users 路由
// 這個 Handler 根據 HTTP 方法（GET 或 POST）做不同的處理
// 就像同一個窗口，你說「我要查詢」和「我要寄件」會走不同流程
func usersHandler(w http.ResponseWriter, r *http.Request) {
	// switch 根據 HTTP 方法分流處理
	switch r.Method { // 檢查請求的 HTTP 方法
	case http.MethodGet: // 如果是 GET 方法（http.MethodGet = "GET"）
		// GET /api/users → 回傳使用者列表（客人說「我要看菜單」）
		getUsers(w, r) // 呼叫 getUsers 函式處理

	case http.MethodPost: // 如果是 POST 方法（http.MethodPost = "POST"）
		// POST /api/users → 建立新使用者（客人說「我要加一道新菜」）
		createUser(w, r) // 呼叫 createUser 函式處理

	default: // 如果是其他方法（PUT、DELETE 等）
		// 回傳 405 Method Not Allowed（這個窗口不辦這種業務）
		// http.Error 是一個便捷函式，設定狀態碼並寫入錯誤訊息
		http.Error(w, "不支援的方法", http.StatusMethodNotAllowed)
	}
}

// getUsers 取得所有使用者（處理 GET 請求）
// 回傳 JSON 格式的使用者列表
func getUsers(w http.ResponseWriter, r *http.Request) {
	// 設定回應的 Content-Type 標頭為 JSON
	// 這告訴客戶端：「我回傳的資料格式是 JSON」
	// 就像在回信的信封上寫「裡面是中文信」
	w.Header().Set("Content-Type", "application/json") // 設定回應標頭

	// json.NewEncoder(w) 建立一個 JSON 編碼器，輸出目標是 w（回應）
	// .Encode(users) 將 users 切片轉成 JSON 格式並寫入回應
	// 例如：[{"id":1,"name":"Alice","age":25}, ...]
	json.NewEncoder(w).Encode(users) // 將使用者列表轉為 JSON 並寫入回應
}

// createUser 建立新使用者（處理 POST 請求）
// 從請求的 body 中讀取 JSON 資料，建立新使用者
func createUser(w http.ResponseWriter, r *http.Request) {
	var newUser User // 宣告一個空的 User 變數，準備接收客戶端傳來的資料

	// json.NewDecoder(r.Body) 建立一個 JSON 解碼器，讀取來源是請求的 body
	// .Decode(&newUser) 將 JSON 解析到 newUser 變數中
	// &newUser 是指標，這樣 Decode 才能修改 newUser 的值
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil { // 如果 JSON 解析失敗
		// 回傳 400 Bad Request（你的信我看不懂）
		http.Error(w, "無效的 JSON: "+err.Error(), http.StatusBadRequest) // 回傳錯誤
		return // 提早返回，不繼續執行
	}

	// 設定 ID 為目前使用者數量 + 1（模擬資料庫的自動遞增 ID）
	newUser.ID = len(users) + 1 // 計算新 ID
	// 將新使用者加入切片（模擬寫入資料庫）
	users = append(users, newUser) // 附加到 users 切片

	// 設定回應的 Content-Type 為 JSON
	w.Header().Set("Content-Type", "application/json") // 設定回應標頭
	// 設定 HTTP 狀態碼為 201 Created（建立成功！）
	// 注意：WriteHeader 必須在 Encode 之前呼叫，否則會預設 200
	w.WriteHeader(http.StatusCreated) // 設定狀態碼為 201
	// 將新建立的使用者轉為 JSON 回傳給客戶端
	json.NewEncoder(w).Encode(newUser) // 回傳新使用者的 JSON
}

// queryHandler 示範如何讀取 URL 查詢參數（Query Parameters）
// 查詢參數是 URL 中 ? 後面的部分
// 例如：/api/search?keyword=hello&page=2
// keyword 和 page 就是查詢參數
func queryHandler(w http.ResponseWriter, r *http.Request) {
	// r.URL.Query() 回傳一個 map，包含所有查詢參數
	// .Get("keyword") 取得 keyword 參數的值
	// 如果參數不存在，回傳空字串 ""
	keyword := r.URL.Query().Get("keyword") // 取得 "keyword" 參數的值
	page := r.URL.Query().Get("page")       // 取得 "page" 參數的值

	// 將查詢結果寫入回應
	fmt.Fprintf(w, "搜尋關鍵字: %s\n頁碼: %s\n", keyword, page) // 回傳參數值
}

// ========================================
// 主程式：啟動 HTTP 伺服器
// ========================================

func main() { // 程式進入點
	// ========================================
	// 註冊路由（告訴郵局哪個地址由誰負責）
	// ========================================
	//
	// http.HandleFunc 將 URL 路徑「對應」到 Handler 函式
	// 就像郵局的分信系統：
	// - 寄到 "/" 的信 → 交給 helloHandler 處理
	// - 寄到 "/api/users" 的信 → 交給 usersHandler 處理
	// - 寄到 "/api/search" 的信 → 交給 queryHandler 處理

	http.HandleFunc("/", helloHandler)           // 根路徑：任何未匹配的路徑都會到這裡
	http.HandleFunc("/api/users", usersHandler)  // 使用者 API：處理 GET 和 POST
	http.HandleFunc("/api/search", queryHandler) // 搜尋 API：示範查詢參數

	// ========================================
	// 啟動伺服器
	// ========================================

	port := ":9090" // 定義伺服器監聽的埠號（port）

	// 印出啟動訊息和使用說明
	fmt.Println("伺服器啟動於 http://localhost" + port) // 顯示伺服器地址
	fmt.Println()                                       // 空行
	fmt.Println("可以嘗試的請求：")                       // 提示標題
	fmt.Println("  瀏覽器開啟: http://localhost:9090")                                    // 測試根路徑
	fmt.Println("  瀏覽器開啟: http://localhost:9090/api/users")                          // 測試 GET 使用者
	fmt.Println("  瀏覽器開啟: http://localhost:9090/api/search?keyword=Go&page=1")      // 測試查詢參數
	fmt.Println()                                                                         // 空行
	fmt.Println("  curl http://localhost:9090/api/users")                                 // 用 curl 測試 GET
	fmt.Println("  curl -X POST http://localhost:9090/api/users -d '{\"name\":\"Dave\",\"age\":35}'") // 用 curl 測試 POST
	fmt.Println()                                                                         // 空行
	fmt.Println("按 Ctrl+C 停止伺服器")                                                    // 如何停止

	// http.ListenAndServe 啟動 HTTP 伺服器
	// 參數 1：port = ":9090"，表示監聽所有網路介面的 9090 埠
	// 參數 2：nil 表示使用預設的路由器（DefaultServeMux）
	//        就是我們上面用 http.HandleFunc 註冊的那些路由
	//
	// 這個函式會「阻塞」— 程式會停在這裡，持續等待並處理請求
	// 只有在伺服器出錯時才會回傳 error
	// log.Fatal 會印出錯誤訊息並結束程式（如果 ListenAndServe 回傳錯誤的話）
	log.Fatal(http.ListenAndServe(port, nil)) // 啟動伺服器，失敗則印出錯誤並結束
}
