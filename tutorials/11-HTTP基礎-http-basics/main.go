// 第十課：HTTP 基礎
// 在學 Gin 框架之前，先用 Go 標準庫理解 HTTP 的運作原理
// 執行方式：go run main.go
// 然後在瀏覽器開啟 http://localhost:9090
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// ========================================
// HTTP 基本概念
// ========================================
//
// HTTP 請求由以下部分組成：
// 1. 方法（Method）：GET、POST、PUT、DELETE 等
// 2. 路徑（Path）：/api/users、/api/articles
// 3. 標頭（Headers）：Content-Type、Authorization
// 4. 主體（Body）：POST/PUT 請求攜帶的資料
//
// HTTP 回應由以下部分組成：
// 1. 狀態碼（Status Code）：200、404、500 等
// 2. 標頭（Headers）：Content-Type
// 3. 主體（Body）：回傳的資料

// User 使用者結構體
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// 模擬的使用者資料庫
var users = []User{
	{ID: 1, Name: "Alice", Age: 25},
	{ID: 2, Name: "Bob", Age: 30},
	{ID: 3, Name: "Carol", Age: 28},
}

// ========================================
// Handler 函式
// ========================================

// helloHandler 最簡單的 Handler
func helloHandler(w http.ResponseWriter, r *http.Request) {
	// w = ResponseWriter：用來寫回應
	// r = Request：包含請求的所有資訊

	// 寫入回應
	fmt.Fprintf(w, "Hello! 你的請求方法是 %s，路徑是 %s", r.Method, r.URL.Path)
}

// usersHandler 處理 /api/users 路由
func usersHandler(w http.ResponseWriter, r *http.Request) {
	// 根據 HTTP 方法做不同處理
	switch r.Method {
	case http.MethodGet:
		// GET /api/users → 回傳使用者列表
		getUsers(w, r)

	case http.MethodPost:
		// POST /api/users → 建立新使用者
		createUser(w, r)

	default:
		// 不支援的方法
		http.Error(w, "不支援的方法", http.StatusMethodNotAllowed)
	}
}

// getUsers 取得所有使用者（GET）
func getUsers(w http.ResponseWriter, r *http.Request) {
	// 設定回應的 Content-Type 為 JSON
	w.Header().Set("Content-Type", "application/json")

	// 將使用者列表轉為 JSON 並寫入回應
	json.NewEncoder(w).Encode(users)
}

// createUser 建立新使用者（POST）
func createUser(w http.ResponseWriter, r *http.Request) {
	var newUser User

	// 從請求 body 解析 JSON
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "無效的 JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 設定 ID
	newUser.ID = len(users) + 1
	users = append(users, newUser)

	// 回傳 201 Created
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 設定狀態碼
	json.NewEncoder(w).Encode(newUser)
}

// queryHandler 示範如何讀取 URL 查詢參數
func queryHandler(w http.ResponseWriter, r *http.Request) {
	// URL: /api/search?keyword=hello&page=2
	keyword := r.URL.Query().Get("keyword")
	page := r.URL.Query().Get("page")

	fmt.Fprintf(w, "搜尋關鍵字: %s\n頁碼: %s\n", keyword, page)
}

func main() {
	// ========================================
	// 註冊路由
	// ========================================

	// http.HandleFunc 將 URL 路徑對應到 handler 函式
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/api/users", usersHandler)
	http.HandleFunc("/api/search", queryHandler)

	// ========================================
	// 啟動伺服器
	// ========================================
	port := ":9090"
	fmt.Println("伺服器啟動於 http://localhost" + port)
	fmt.Println()
	fmt.Println("可以嘗試的請求：")
	fmt.Println("  瀏覽器開啟: http://localhost:9090")
	fmt.Println("  瀏覽器開啟: http://localhost:9090/api/users")
	fmt.Println("  瀏覽器開啟: http://localhost:9090/api/search?keyword=Go&page=1")
	fmt.Println()
	fmt.Println("  curl http://localhost:9090/api/users")
	fmt.Println("  curl -X POST http://localhost:9090/api/users -d '{\"name\":\"Dave\",\"age\":35}'")
	fmt.Println()
	fmt.Println("按 Ctrl+C 停止伺服器")

	// ListenAndServe 會阻塞，持續監聽請求
	log.Fatal(http.ListenAndServe(port, nil))
}
