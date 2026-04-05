// 第四課：結構體與方法（Structs & Methods）
// 執行方式：go run main.go
package main

import "fmt"

// ========================================
// 1. 定義結構體（Struct）
// ========================================

// User 定義使用者結構體
// 結構體是一組相關資料的集合，類似其他語言的 class
type User struct {
	ID       int
	Username string
	Email    string
	Age      int
}

// ========================================
// 2. 方法（Methods）
// ========================================

// 方法和函式很像，但多了一個「接收者（receiver）」
// 接收者寫在 func 和函式名之間

// DisplayName 是 User 的方法
// (u User) 是「值接收者」—— u 是 User 的副本
func (u User) DisplayName() string {
	return fmt.Sprintf("%s (%s)", u.Username, u.Email)
}

// IsAdult 檢查使用者是否成年
func (u User) IsAdult() bool {
	return u.Age >= 18
}

// SetEmail 使用「指標接收者」修改原始結構體
// (u *User) 中的 * 表示指標 —— u 指向原始資料，不是副本
func (u *User) SetEmail(newEmail string) {
	u.Email = newEmail // 這會修改原始的 User
}

// ========================================
// 3. 結構體嵌套（Embedding）
// ========================================

// Address 定義地址結構體
type Address struct {
	City    string
	Country string
}

// Employee 嵌套了 User 和 Address
type Employee struct {
	User       // 嵌套 User（匿名欄位）
	Address    // 嵌套 Address（匿名欄位）
	Department string
	Salary     float64
}

// ========================================
// 4. 建構函式（Constructor Pattern）
// ========================================

// NewUser 是建構函式（Go 的慣例：用 New + 型別名稱）
// Go 沒有 class 的 constructor，所以用普通函式代替
func NewUser(username, email string, age int) *User {
	return &User{
		Username: username,
		Email:    email,
		Age:      age,
	}
}

// ========================================
// 5. 結構體標籤（Struct Tags）
// ========================================

// Article 示範結構體標籤的用法
// 標籤是附加在欄位上的元資料，供其他套件讀取
type Article struct {
	ID      int    `json:"id"`                    // JSON 輸出時叫 "id"
	Title   string `json:"title"`                 // JSON 輸出時叫 "title"
	Content string `json:"content"`               // JSON 輸出時叫 "content"
	Secret  string `json:"-"`                     // json:"-" 代表不輸出
	UserID  int    `json:"user_id" gorm:"index"`  // 可以有多個標籤
}

func main() {
	// ========================================
	// 建立結構體實例的幾種方式
	// ========================================
	fmt.Println("=== 建立結構體 ===")

	// 方式 1：指定欄位名稱（推薦）
	user1 := User{
		Username: "alice",
		Email:    "alice@example.com",
		Age:      25,
	}

	// 方式 2：按順序賦值（不推薦，容易出錯）
	user2 := User{0, "bob", "bob@example.com", 30}

	// 方式 3：先宣告再賦值
	var user3 User
	user3.Username = "charlie"
	user3.Email = "charlie@example.com"
	user3.Age = 20

	// 方式 4：使用建構函式
	user4 := NewUser("diana", "diana@example.com", 28)

	fmt.Println("user1:", user1)
	fmt.Println("user2:", user2)
	fmt.Println("user3:", user3)
	fmt.Println("user4:", user4)  // 注意：這是指標 *User
	fmt.Println("*user4:", *user4) // 解引用看到實際值

	// ========================================
	// 呼叫方法
	// ========================================
	fmt.Println("\n=== 呼叫方法 ===")

	fmt.Println("顯示名稱:", user1.DisplayName())
	fmt.Println("是否成年:", user1.IsAdult())

	// 使用指標接收者的方法修改值
	fmt.Println("\n修改前 Email:", user1.Email)
	user1.SetEmail("alice.new@example.com")
	fmt.Println("修改後 Email:", user1.Email) // 已被修改

	// ========================================
	// 值接收者 vs 指標接收者的差異
	// ========================================
	fmt.Println("\n=== 值 vs 指標接收者 ===")

	original := User{Username: "test", Email: "old@test.com"}

	// 值接收者：操作的是副本，不會改變原始值
	name := original.DisplayName() // 內部的 u 是副本
	fmt.Println("DisplayName:", name)

	// 指標接收者：操作的是原始值，會改變
	original.SetEmail("new@test.com") // 內部的 u 指向 original
	fmt.Println("Email 已改變:", original.Email)

	// ========================================
	// 結構體嵌套
	// ========================================
	fmt.Println("\n=== 結構體嵌套 ===")

	emp := Employee{
		User: User{
			Username: "王小明",
			Email:    "xiaoming@company.com",
			Age:      30,
		},
		Address: Address{
			City:    "台北",
			Country: "台灣",
		},
		Department: "工程部",
		Salary:     50000,
	}

	// 嵌套的欄位可以直接存取（不需要寫 emp.User.Username）
	fmt.Println("姓名:", emp.Username)       // 直接存取 User 的欄位
	fmt.Println("城市:", emp.City)            // 直接存取 Address 的欄位
	fmt.Println("部門:", emp.Department)
	fmt.Println("顯示:", emp.DisplayName())   // 也能直接呼叫 User 的方法

	// 也可以明確指定
	fmt.Println("完整路徑:", emp.User.Email)

	// ========================================
	// 結構體比較
	// ========================================
	fmt.Println("\n=== 結構體比較 ===")

	a := User{Username: "alice", Email: "alice@test.com"}
	b := User{Username: "alice", Email: "alice@test.com"}
	c := User{Username: "bob", Email: "bob@test.com"}

	fmt.Println("a == b:", a == b) // true（所有欄位相同）
	fmt.Println("a == c:", a == c) // false
}
