// =============================================================================
// 第十三課：JSON 處理與結構標籤（Struct Tags）
// =============================================================================
//
// 什麼是結構標籤（Struct Tags）？
// 結構標籤是附加在結構體欄位上的「後設資料」（metadata）。
// 不同的套件會讀取不同的標籤，來決定如何處理這個欄位。
// 你可以把它想像成「貼在箱子上的標籤」：
//   - json 標籤 → 告訴 JSON 系統這個欄位叫什麼名字
//   - binding 標籤 → 告訴 Gin 框架要怎麼驗證這個欄位
//   - gorm 標籤 → 告訴 GORM 框架這個欄位對應資料庫哪個欄位
//   - example 標籤 → 告訴 Swagger 文件這個欄位的範例值
//
// 語法：
//   欄位名 型別 `標籤名:"值" 標籤名:"值"`
//   反引號 `` 包住所有標籤，每個標籤之間用空格分隔
//
// 執行方式：go run ./tutorials/13-json-binding
// =============================================================================

package main // 每個可執行的 Go 程式都必須是 main 套件

import (                 // import 區塊：匯入需要的套件
	"encoding/json"      // encoding/json：Go 標準庫的 JSON 編碼/解碼套件
	"fmt"                // fmt：格式化輸出，用來印東西到終端機
)

// =============================================================================
// 第一部分：json 標籤 — 控制 JSON 序列化（Go 結構體 ↔ JSON 字串）
// =============================================================================

// User 示範各種 json 標籤的用法
// 每個欄位的 json 標籤決定了它在 JSON 中的名字和行為
type User struct {
	// json:"id" → 轉成 JSON 時，欄位名稱是 "id"（而不是 Go 的 "ID"）
	// Go 慣例用大寫開頭（PascalCase），但 JSON 慣例用小寫（camelCase 或 snake_case）
	ID int `json:"id"`

	// json:"username" → JSON 欄位名稱是 "username"
	Username string `json:"username"`

	// json:"email" → JSON 欄位名稱是 "email"
	Email string `json:"email"`

	// json:"-" → 這個欄位「永遠不會」出現在 JSON 輸出中
	// 這是安全的關鍵！密碼絕對不能傳給前端
	// 即使你不小心把整個 User 結構體回傳，密碼也不會洩漏
	Password string `json:"-"`

	// json:"age,omitempty" → 如果 age 是零值（0），這個欄位不會出現在 JSON 中
	// omitempty 的意思是「如果是空的就省略」
	// 零值定義：數字是 0、字串是 ""、布林是 false、指標是 nil
	Age int `json:"age,omitempty"`

	// json:"nickname,omitempty" → 如果 nickname 是空字串，就不會出現在 JSON 中
	Nickname string `json:"nickname,omitempty"`
}

// =============================================================================
// 第二部分：binding 標籤 — Gin 框架的請求驗證
// =============================================================================
// binding 標籤由 Gin 框架讀取，在 c.ShouldBindJSON() 時自動驗證
// 如果驗證不通過，ShouldBindJSON 會回傳 error

// CreateUserRequest 示範 Gin 的 binding 驗證標籤
// 這種「Request 結構體」專門用來接收客戶端送來的 JSON
type CreateUserRequest struct {
	// binding:"required,min=3,max=50" 的意思是：
	//   required → 這個欄位必須有值（不能是空字串）
	//   min=3    → 字串最短 3 個字元
	//   max=50   → 字串最長 50 個字元
	Username string `json:"username" binding:"required,min=3,max=50"`

	// binding:"required,email" 的意思是：
	//   required → 必填
	//   email    → 必須是合法的 Email 格式（例如 user@example.com）
	Email string `json:"email" binding:"required,email"`

	// binding:"required,min=6" 的意思是：
	//   required → 必填
	//   min=6    → 密碼最短 6 個字元
	Password string `json:"password" binding:"required,min=6"`

	// binding:"omitempty,min=0,max=150" 的意思是：
	//   omitempty → 如果沒給這個欄位，就跳過驗證（選填）
	//   min=0     → 如果有給值，數字最小是 0
	//   max=150   → 如果有給值，數字最大是 150
	Age int `json:"age" binding:"omitempty,min=0,max=150"`
}

// =============================================================================
// 第三部分：gorm 標籤 — GORM 框架的資料庫欄位定義
// =============================================================================
// gorm 標籤由 GORM 框架讀取，用來定義資料庫表格的欄位屬性
// 例如：主鍵、索引、欄位長度、是否可為空等

// Article 示範 GORM 的資料庫標籤
type Article struct {
	// gorm:"primaryKey" → 這個欄位是資料庫的「主鍵」（Primary Key）
	// 主鍵是每筆資料的唯一識別，通常是自動遞增的數字
	ID int `json:"id" gorm:"primaryKey"`

	// gorm:"size:200;not null" 的意思是：
	//   size:200 → 資料庫欄位長度最多 200 個字元（VARCHAR(200)）
	//   not null → 不可為空（插入資料時一定要有值）
	//   分號 ; 用來分隔多個 gorm 設定
	Title string `json:"title" gorm:"size:200;not null"`

	// gorm:"type:text" → 資料庫欄位型別是 TEXT（可以存很長的文字）
	// TEXT vs VARCHAR：VARCHAR 有長度限制，TEXT 沒有（適合存文章內容）
	Content string `json:"content" gorm:"type:text"`

	// gorm:"index;not null" 的意思是：
	//   index    → 為這個欄位建立「索引」（加速查詢速度）
	//   not null → 不可為空
	UserID int `json:"user_id" gorm:"index;not null"`
}

// =============================================================================
// 第四部分：example 標籤 — Swagger API 文件的範例值
// =============================================================================
// example 標籤由 Swagger 文件產生工具讀取（如 swag）
// 它不會影響程式邏輯，只是讓 API 文件顯示範例值

// RegisterRequest 示範 example 標籤
// 這是部落格專案中「使用者註冊」的請求結構體
type RegisterRequest struct {
	// example:"newuser" → Swagger 文件中會顯示 "newuser" 作為範例
	Username string `json:"username" binding:"required,min=3,max=50" example:"newuser"`

	// example:"newuser@example.com" → Swagger 文件中會顯示這個範例 Email
	Email string `json:"email" binding:"required,email" example:"newuser@example.com"`

	// example:"password123" → Swagger 文件中會顯示這個範例密碼
	Password string `json:"password" binding:"required,min=6,max=100" example:"password123"`
}

// =============================================================================
// 第五部分：多個標籤共存 — 一個欄位可以同時有多種標籤
// =============================================================================
// 在真實專案中，一個欄位通常同時需要 json + gorm + binding + example 標籤

// Comment 展示一個欄位可以有多個標籤
// 每個標籤之間用空格分隔，由不同的套件各自讀取
type Comment struct {
	// json:"id"          → JSON 回應中叫 "id"
	// gorm:"primaryKey"  → 資料庫中是主鍵
	// binding:"-"        → Gin 綁定時「忽略」這個欄位（ID 由系統產生，不是使用者填的）
	ID int `json:"id" gorm:"primaryKey" binding:"-"`

	// json:"content"                  → JSON 欄位名
	// gorm:"type:text;not null"       → 資料庫是 TEXT 型別且不可為空
	// binding:"required,min=1,max=1000" → 必填，1~1000 字元
	Content string `json:"content" gorm:"type:text;not null" binding:"required,min=1,max=1000"`

	// json:"user_id"  → JSON 欄位名
	// gorm:"index"    → 資料庫建立索引（加速依使用者查詢留言的速度）
	// binding:"-"     → 忽略綁定（UserID 從 JWT Token 取得，不是使用者填的）
	UserID int `json:"user_id" gorm:"index" binding:"-"`
}

// =============================================================================
// main 函式 — 示範 JSON 序列化和反序列化
// =============================================================================

func main() {
	// =================================================================
	// 示範 1：結構體 → JSON（Marshal / 序列化）
	// =================================================================
	// Marshal 的意思是「編組」，把 Go 結構體轉成 JSON 字串
	fmt.Println("=== 示範 1：結構體 → JSON（Marshal）===") // 印出標題

	user := User{                          // 建立一個 User 結構體
		ID:       1,                        // ID 設為 1
		Username: "alice",                  // 使用者名稱
		Email:    "alice@example.com",      // Email
		Password: "secret123",              // 密碼（有 json:"-"，不會出現在 JSON）
		Age:      25,                       // 年齡
		Nickname: "",                       // 空字串 + omitempty → 不會出現在 JSON
	}

	// json.Marshal(值) → 把結構體轉成 JSON 位元組切片（[]byte）
	// 回傳兩個值：JSON 位元組和 error
	jsonBytes, err := json.Marshal(user)   // 序列化 user 結構體
	if err != nil {                        // 如果序列化失敗
		fmt.Println("序列化錯誤:", err)      // 印出錯誤訊息
		return                              // 提早返回
	}
	// string(jsonBytes) 把位元組切片轉成字串，方便印出
	fmt.Println("Marshal 結果:", string(jsonBytes)) // 印出 JSON 字串
	// 輸出：{"id":1,"username":"alice","email":"alice@example.com","age":25}
	// 注意：Password 被 json:"-" 隱藏了，Nickname 被 omitempty 省略了

	fmt.Println() // 印出空行分隔

	// json.MarshalIndent(值, 前綴, 縮排) → 產生格式化（好看）的 JSON
	// 第二個參數是每行前綴（通常空字串），第三個參數是縮排字串
	prettyJSON, err := json.MarshalIndent(user, "", "  ") // 用兩個空格縮排
	if err != nil {                        // 如果序列化失敗
		fmt.Println("格式化序列化錯誤:", err) // 印出錯誤
		return                              // 提早返回
	}
	fmt.Println("MarshalIndent 結果:")     // 印出標題
	fmt.Println(string(prettyJSON))        // 印出格式化的 JSON

	fmt.Println() // 印出空行分隔

	// =================================================================
	// 示範 2：JSON → 結構體（Unmarshal / 反序列化）
	// =================================================================
	// Unmarshal 的意思是「解組」，把 JSON 字串轉成 Go 結構體
	fmt.Println("=== 示範 2：JSON → 結構體（Unmarshal）===") // 印出標題

	// 模擬從客戶端收到的 JSON 字串
	// 注意：這裡包含了 password 欄位，但因為 json:"-"，它不會被寫入結構體
	jsonStr := `{
		"id": 2,
		"username": "bob",
		"email": "bob@example.com",
		"age": 30,
		"password": "hacker_attempt"
	}` // 原始 JSON 字串（使用反引號包住，可以跨行）

	var user2 User                                     // 宣告一個空的 User 結構體
	err = json.Unmarshal([]byte(jsonStr), &user2)      // 把 JSON 字串解析到 user2 中
	// []byte(jsonStr) → 把字串轉成位元組切片（Unmarshal 需要 []byte）
	// &user2 → 傳入 user2 的指標（讓 Unmarshal 能修改 user2 的值）
	if err != nil {                                    // 如果解析失敗
		fmt.Println("反序列化錯誤:", err)                // 印出錯誤
		return                                          // 提早返回
	}

	// %+v 格式化列印結構體，會顯示欄位名稱
	fmt.Printf("Unmarshal 結果: %+v\n", user2)         // 印出結構體內容
	fmt.Println("Password 欄位:", user2.Password)       // 印出密碼欄位
	// 密碼會是空字串！json:"-" 阻止了 JSON 中的 password 寫入 Password 欄位

	fmt.Println() // 印出空行分隔

	// =================================================================
	// 示範 3：omitempty 的效果對比
	// =================================================================
	fmt.Println("=== 示範 3：omitempty 效果對比 ===") // 印出標題

	// 情境一：欄位有值 → omitempty 不起作用，正常輸出
	userWithValues := User{                // 建立有值的使用者
		ID:       1,                        // ID
		Username: "test",                   // 使用者名稱
		Age:      25,                       // 年齡有值 → 會出現在 JSON
		Nickname: "小T",                    // 暱稱有值 → 會出現在 JSON
	}
	j1, _ := json.Marshal(userWithValues)  // 序列化（忽略 error）
	fmt.Println("有值:", string(j1))        // 印出結果
	// 輸出：{"id":1,"username":"test","email":"","age":25,"nickname":"小T"}
	// 注意：email 沒有 omitempty，所以即使是空字串也會出現

	// 情境二：欄位是零值 → omitempty 生效，省略這些欄位
	userWithZeros := User{                 // 建立零值的使用者
		ID:       2,                        // ID
		Username: "test2",                  // 使用者名稱
		// Age 沒給 → 預設是 0（零值）→ omitempty 會省略
		// Nickname 沒給 → 預設是 ""（空字串）→ omitempty 會省略
	}
	j2, _ := json.Marshal(userWithZeros)   // 序列化
	fmt.Println("零值:", string(j2))        // 印出結果
	// 輸出：{"id":2,"username":"test2","email":""}
	// age 和 nickname 因為 omitempty 被省略了

	fmt.Println() // 印出空行分隔

	// =================================================================
	// 示範 4：巢狀結構體的 JSON
	// =================================================================
	// 結構體裡面可以包含其他結構體，JSON 也會跟著變成巢狀
	fmt.Println("=== 示範 4：巢狀結構體 ===") // 印出標題

	// Author 代表文章作者（定義在函式內部也是合法的）
	type Author struct {
		Name  string `json:"name"`  // 作者名稱 → JSON 欄位名 "name"
		Email string `json:"email"` // 作者 Email → JSON 欄位名 "email"
	}

	// Post 代表一篇文章
	type Post struct {
		Title   string `json:"title"`   // 文章標題
		Content string `json:"content"` // 文章內容
		Author  Author `json:"author"`  // 作者（巢狀結構體）→ JSON 會變成巢狀物件
	}

	post := Post{                                            // 建立文章
		Title:   "Go 教學",                                   // 設定標題
		Content: "學習 Go 語言...",                            // 設定內容
		Author:  Author{Name: "Alice", Email: "a@test.com"}, // 設定作者
	}

	postJSON, _ := json.MarshalIndent(post, "", "  ") // 格式化序列化
	fmt.Println(string(postJSON))                      // 印出結果
	// 輸出會是巢狀的 JSON：
	// {
	//   "title": "Go 教學",
	//   "content": "學習 Go 語言...",
	//   "author": {
	//     "name": "Alice",
	//     "email": "a@test.com"
	//   }
	// }

	fmt.Println() // 印出空行分隔

	// =================================================================
	// 示範 5：JSON 陣列（切片序列化）
	// =================================================================
	// 切片（slice）序列化後會變成 JSON 陣列（array）
	fmt.Println("=== 示範 5：JSON 陣列 ===") // 印出標題

	users := []User{                                          // 建立 User 切片
		{ID: 1, Username: "alice", Email: "a@test.com", Age: 25}, // 第一個使用者
		{ID: 2, Username: "bob", Email: "b@test.com", Age: 30},   // 第二個使用者
	}

	usersJSON, _ := json.MarshalIndent(users, "", "  ") // 格式化序列化
	fmt.Println(string(usersJSON))                       // 印出 JSON 陣列

	fmt.Println() // 印出空行分隔

	// =================================================================
	// 示範 6：動態 JSON（使用 map，不需要預先定義結構體）
	// =================================================================
	// 有時候 JSON 的結構不固定，可以用 map[string]interface{} 來建立
	// interface{} 表示「任意型別」（Go 1.18+ 可以寫成 any）
	fmt.Println("=== 示範 6：動態 JSON ===") // 印出標題

	response := map[string]interface{}{    // 建立一個 map（鍵是字串，值是任意型別）
		"code":    200,                     // HTTP 狀態碼
		"message": "success",               // 訊息
		"data": map[string]interface{}{     // 巢狀 map（資料內容）
			"total": 42,                    // 總數
			"items": []string{"a", "b", "c"}, // 字串切片
		},
	}

	respJSON, _ := json.MarshalIndent(response, "", "  ") // 格式化序列化
	fmt.Println(string(respJSON))                          // 印出結果
	// 注意：map 的鍵順序不固定（Go 的 map 是無序的）

	fmt.Println() // 印出空行分隔

	// =================================================================
	// 示範 7：JSON → map（解析不確定結構的 JSON）
	// =================================================================
	// 如果你不知道 JSON 長什麼樣，可以解析到 map 裡
	fmt.Println("=== 示範 7：JSON → map ===") // 印出標題

	unknownJSON := `{"name": "Alice", "score": 95.5, "active": true}` // 不確定結構的 JSON
	var result map[string]interface{}                                   // 宣告一個 map 來接收
	err = json.Unmarshal([]byte(unknownJSON), &result)                 // 解析 JSON 到 map
	if err != nil {                                                    // 如果解析失敗
		fmt.Println("解析錯誤:", err)                                    // 印出錯誤
		return                                                          // 提早返回
	}
	fmt.Println("name:", result["name"])       // 取得 name 的值（型別是 interface{}）
	fmt.Println("score:", result["score"])      // 取得 score 的值（JSON 數字會變成 float64）
	fmt.Println("active:", result["active"])    // 取得 active 的值（JSON 布林會變成 bool）

	fmt.Println() // 印出空行分隔

	// =================================================================
	// 附錄：binding 驗證標籤速查表
	// =================================================================
	fmt.Println("=== binding 驗證標籤速查表 ===") // 印出標題
	fmt.Println("required       → 必填（不可為零值）")               // 必填規則
	fmt.Println("min=3          → 字串最短 3 字元 / 數字最小值 3")   // 最小值
	fmt.Println("max=50         → 字串最長 50 字元 / 數字最大值 50") // 最大值
	fmt.Println("email          → 必須是合法的 Email 格式")          // Email 驗證
	fmt.Println("omitempty      → 空值時跳過驗證（選填欄位）")        // 選填
	fmt.Println("oneof=a b c    → 值必須是 a、b 或 c 之一")         // 列舉值
	fmt.Println("gt=0           → 大於 0（greater than）")          // 大於
	fmt.Println("gte=0          → 大於等於 0（greater than or equal）") // 大於等於
	fmt.Println("lt=100         → 小於 100（less than）")           // 小於
	fmt.Println("lte=100        → 小於等於 100（less than or equal）")  // 小於等於
	fmt.Println("len=10         → 長度必須剛好是 10")                // 固定長度
	fmt.Println("url            → 必須是合法的 URL")                 // URL 驗證

	fmt.Println() // 印出空行分隔

	// =================================================================
	// 附錄：gorm 標籤速查表
	// =================================================================
	fmt.Println("=== gorm 標籤速查表 ===")    // 印出標題
	fmt.Println("primaryKey     → 主鍵（每筆資料的唯一識別）")       // 主鍵
	fmt.Println("autoIncrement  → 自動遞增（通常和 primaryKey 搭配）") // 自動遞增
	fmt.Println("uniqueIndex    → 唯一索引（值不可重複，如 Email）")  // 唯一索引
	fmt.Println("index          → 一般索引（加速查詢）")             // 一般索引
	fmt.Println("not null       → 不可為空（必須有值）")             // 非空
	fmt.Println("size:200       → 欄位長度（VARCHAR(200)）")        // 欄位長度
	fmt.Println("type:text      → 資料庫型別（TEXT、BLOB 等）")      // 型別
	fmt.Println("default:0      → 預設值（沒給值時自動填入）")        // 預設值
	fmt.Println("column:user_id → 自訂欄位名稱")                    // 自訂欄位名
	fmt.Println("foreignKey:X   → 外鍵（建立表格之間的關聯）")       // 外鍵

	fmt.Println() // 印出空行分隔

	// =================================================================
	// 附錄：example 標籤速查
	// =================================================================
	fmt.Println("=== example 標籤速查 ===")   // 印出標題
	fmt.Println("example:\"值\"  → Swagger API 文件中顯示的範例值") // 範例值
	fmt.Println("用途：讓 API 文件更清楚，開發者一看就知道要填什麼")  // 用途說明
	fmt.Println("注意：不會影響程式邏輯，只影響文件產生工具")          // 注意事項
}
