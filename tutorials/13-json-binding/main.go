// 第十二課：JSON 處理與結構標籤（Struct Tags）
// 了解 Go 如何將 JSON 和結構體互相轉換
// 執行方式：go run main.go
package main

import (
	"encoding/json"
	"fmt"
)

// ========================================
// 1. JSON 標籤基礎
// ========================================

// User 示範各種 JSON 標籤
type User struct {
	ID        int    `json:"id"`                  // JSON 欄位名為 "id"
	Username  string `json:"username"`            // JSON 欄位名為 "username"
	Email     string `json:"email"`               // JSON 欄位名為 "email"
	Password  string `json:"-"`                   // 永遠不會出現在 JSON 中
	Age       int    `json:"age,omitempty"`       // 值為零值時省略
	Nickname  string `json:"nickname,omitempty"`  // 空字串時省略
}

// ========================================
// 2. Binding 標籤（Gin 的驗證）
// ========================================

// CreateUserRequest 示範 Gin 的 binding 驗證標籤
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Age      int    `json:"age"      binding:"omitempty,min=0,max=150"`
}

// ========================================
// 3. GORM 標籤
// ========================================

// Article 示範 GORM 的資料庫標籤
type Article struct {
	ID      int    `json:"id"      gorm:"primaryKey"`
	Title   string `json:"title"   gorm:"size:200;not null"`
	Content string `json:"content" gorm:"type:text"`
	UserID  int    `json:"user_id" gorm:"index;not null"`
}

// ========================================
// 4. 多個標籤共存
// ========================================

// Comment 展示一個欄位可以有多個標籤
type Comment struct {
	ID      int    `json:"id"      gorm:"primaryKey"                   binding:"-"`
	Content string `json:"content" gorm:"type:text;not null"           binding:"required,min=1,max=1000"`
	UserID  int    `json:"user_id" gorm:"index"                       binding:"-"`
}

func main() {
	// ========================================
	// 1. 結構體 → JSON（Marshal / 序列化）
	// ========================================
	fmt.Println("=== 結構體 → JSON ===")

	user := User{
		ID:       1,
		Username: "alice",
		Email:    "alice@example.com",
		Password: "secret123",    // 有 json:"-"，不會出現在 JSON
		Age:      25,
		Nickname: "",             // 空字串 + omitempty → 不會出現
	}

	// json.Marshal：轉為 JSON 位元組
	jsonBytes, _ := json.Marshal(user)
	fmt.Println("Marshal:", string(jsonBytes))
	// 輸出：{"id":1,"username":"alice","email":"alice@example.com","age":25}
	// 注意：Password 和 Nickname 都不在輸出中

	// json.MarshalIndent：格式化的 JSON（適合除錯）
	prettyJSON, _ := json.MarshalIndent(user, "", "  ")
	fmt.Println("\nMarshalIndent:")
	fmt.Println(string(prettyJSON))

	// ========================================
	// 2. JSON → 結構體（Unmarshal / 反序列化）
	// ========================================
	fmt.Println("\n=== JSON → 結構體 ===")

	jsonStr := `{
		"id": 2,
		"username": "bob",
		"email": "bob@example.com",
		"age": 30,
		"password": "hacker_attempt"
	}`

	var user2 User
	err := json.Unmarshal([]byte(jsonStr), &user2)
	if err != nil {
		fmt.Println("解析錯誤:", err)
		return
	}

	fmt.Printf("解析結果: %+v\n", user2)
	fmt.Println("Password:", user2.Password) // 空字串！json:"-" 阻止了寫入

	// ========================================
	// 3. omitempty 的效果
	// ========================================
	fmt.Println("\n=== omitempty 效果 ===")

	// 有值的情況
	userWithAge := User{ID: 1, Username: "test", Age: 25, Nickname: "小T"}
	j1, _ := json.Marshal(userWithAge)
	fmt.Println("有值:", string(j1))

	// 零值的情況
	userNoAge := User{ID: 2, Username: "test2"}
	j2, _ := json.Marshal(userNoAge)
	fmt.Println("零值:", string(j2))
	// age 和 nickname 不會出現（omitempty）

	// ========================================
	// 4. 巢狀結構體的 JSON
	// ========================================
	fmt.Println("\n=== 巢狀結構體 ===")

	type Author struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	type Post struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Author  Author `json:"author"`       // 巢狀結構體
	}

	post := Post{
		Title:   "Go 教學",
		Content: "學習 Go 語言...",
		Author:  Author{Name: "Alice", Email: "alice@test.com"},
	}

	postJSON, _ := json.MarshalIndent(post, "", "  ")
	fmt.Println(string(postJSON))

	// ========================================
	// 5. JSON 陣列
	// ========================================
	fmt.Println("\n=== JSON 陣列 ===")

	users := []User{
		{ID: 1, Username: "alice", Email: "alice@test.com", Age: 25},
		{ID: 2, Username: "bob", Email: "bob@test.com", Age: 30},
	}

	usersJSON, _ := json.MarshalIndent(users, "", "  ")
	fmt.Println(string(usersJSON))

	// ========================================
	// 6. 動態 JSON（使用 map）
	// ========================================
	fmt.Println("\n=== 動態 JSON ===")

	// 不需要預先定義結構體
	response := map[string]interface{}{
		"code":    200,
		"message": "success",
		"data": map[string]interface{}{
			"total": 42,
			"items": []string{"a", "b", "c"},
		},
	}

	respJSON, _ := json.MarshalIndent(response, "", "  ")
	fmt.Println(string(respJSON))

	// ========================================
	// 7. Binding 標籤速查
	// ========================================
	fmt.Println("\n=== Binding 驗證標籤速查 ===")
	fmt.Println("required     → 必填")
	fmt.Println("min=3        → 字串最短 3 / 數字最小 3")
	fmt.Println("max=50       → 字串最長 50 / 數字最大 50")
	fmt.Println("email        → 必須是合法 Email")
	fmt.Println("omitempty    → 空值時跳過驗證")
	fmt.Println("oneof=a b c  → 值必須是 a、b 或 c 之一")
	fmt.Println("gt=0         → 大於 0")
	fmt.Println("gte=0        → 大於等於 0")

	// ========================================
	// 8. GORM 標籤速查
	// ========================================
	fmt.Println("\n=== GORM 標籤速查 ===")
	fmt.Println("primaryKey   → 主鍵")
	fmt.Println("uniqueIndex  → 唯一索引")
	fmt.Println("index        → 一般索引")
	fmt.Println("not null     → 不可為空")
	fmt.Println("size:200     → 欄位長度")
	fmt.Println("type:text    → 資料庫型別")
	fmt.Println("default:0    → 預設值")
	fmt.Println("foreignKey:X → 外鍵")
}
