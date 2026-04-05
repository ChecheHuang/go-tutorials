// ============================================================
// 第六課：介面（Interfaces）
// ============================================================
// 介面是什麼？想像一份「合約」或「協議」：
// 你跟外送平台簽約，合約上寫著「你必須能接單、能送餐」。
// 不管你是騎機車的、騎腳踏車的、還是走路的，
// 只要你能「接單」和「送餐」，你就是合格的外送員。
//
// 在程式中，介面就是這樣的合約：
// 它定義了「你必須有哪些方法」，但不管你是怎麼實作的。
//
// 你會學到：
//   1. 定義介面：一組方法的合約
//   2. 隱式實作：Go 不需要寫 implements（自動滿足）
//   3. 多型：同一個函式處理不同型別
//   4. 空介面 any（interface{}）：接受任何型別
//   5. 型別斷言（Type Assertion）：從 any 中取出具體型別
//   6. 型別 Switch：根據型別做不同處理
//   7. 依賴注入：Clean Architecture 的 Repository 模式
//
// 執行方式：go run ./tutorials/06-interfaces
// ============================================================

package main // 可執行程式的套件名稱，每個可執行程式都必須是 main

import (      // 匯入需要的標準函式庫
	"fmt"  // fmt 套件：用來格式化輸出到螢幕
	"math" // math 套件：提供數學常數和函式（例如 math.Pi）
)

// ========================================
// 1. 定義介面（合約）
// ========================================

// Shape 是一個介面，定義了「形狀」必須遵守的合約
// 任何型別只要擁有 Area() 和 Perimeter() 這兩個方法，
// 就自動被視為 Shape —— 不需要額外宣告！
type Shape interface {       // 用 interface 關鍵字定義介面
	Area() float64       // 合約條款 1：必須有計算面積的方法，回傳 float64
	Perimeter() float64  // 合約條款 2：必須有計算周長的方法，回傳 float64
}

// ========================================
// 2. 實作介面（隱式實作 — 不需要寫 implements！）
// ========================================
// Go 的介面是「隱式」的：
// 你不需要宣告「我實作了 Shape」，
// 只要你的型別有 Area() 和 Perimeter() 方法，你就是 Shape。
// 這就像：你不需要跟人說「我是外送員」，
// 只要你能接單、能送餐，你就是外送員。

// Circle 是圓形結構體
type Circle struct {   // 定義一個圓形的結構體
	Radius float64 // 半徑，型別是 float64（小數）
}

// Area 計算圓形的面積（公式：π × r²）
// 這是 Circle 實作 Shape 介面的第一個方法
func (c Circle) Area() float64 { // (c Circle) 是值接收者，代表「圓形的」Area 方法
	return math.Pi * c.Radius * c.Radius // π × 半徑 × 半徑
}

// Perimeter 計算圓形的周長（公式：2 × π × r）
// 這是 Circle 實作 Shape 介面的第二個方法
// 兩個方法都有了，Circle 就自動滿足 Shape 介面！
func (c Circle) Perimeter() float64 { // 圓形的周長方法
	return 2 * math.Pi * c.Radius // 2 × π × 半徑
}

// Rectangle 是矩形結構體
type Rectangle struct {      // 定義一個矩形的結構體
	Width, Height float64 // 寬和高，都是 float64 型別（可以一行宣告多個同型別欄位）
}

// Area 計算矩形的面積（公式：寬 × 高）
// Rectangle 也實作了 Shape 介面的 Area() 方法
func (r Rectangle) Area() float64 { // 矩形的面積方法
	return r.Width * r.Height // 寬 × 高
}

// Perimeter 計算矩形的周長（公式：2 × (寬 + 高)）
// Rectangle 也實作了 Shape 介面的 Perimeter() 方法
// 有了 Area() 和 Perimeter()，Rectangle 也自動滿足 Shape！
func (r Rectangle) Perimeter() float64 { // 矩形的周長方法
	return 2 * (r.Width + r.Height) // 2 × (寬 + 高)
}

// ========================================
// 3. 使用介面（多型 — 同一函式，處理不同型別）
// ========================================

// printShapeInfo 接收任何滿足 Shape 介面的型別
// 不管傳入圓形、矩形還是其他形狀，只要有 Area() 和 Perimeter() 就能用
// 這就是「多型」：同一個函式，因傳入型別不同而有不同行為
func printShapeInfo(s Shape) { // 參數型別是 Shape「介面」，不是某個具體的結構體
	fmt.Printf("  面積: %.2f\n", s.Area())      // 呼叫該形狀的面積方法（具體行為取決於型別）
	fmt.Printf("  周長: %.2f\n", s.Perimeter()) // 呼叫該形狀的周長方法（具體行為取決於型別）
}

// ========================================
// 4. 空介面 interface{} 和 any
// ========================================
// 空介面沒有任何方法要求，所以「所有型別」都自動滿足它。
// 就像一份「什麼條件都不要求」的合約，人人都合格。
//
// interface{} 是 Go 1.18 之前的寫法
// any 是 Go 1.18 之後的別名，兩者完全等價
// 建議使用 any（更簡潔易讀）

// printAnything 接收任何型別的值
// 用 any（空介面）作為參數型別，什麼都能傳進來
func printAnything(value any) { // any 等於 interface{}，接受任何型別
	fmt.Printf("  型別: %-10T 值: %v\n", value, value) // %T 印出型別名稱，%v 印出值
}

// ========================================
// 5. 型別斷言（Type Assertion）— 從 any 取出具體型別
// ========================================
// 當你有一個 any（空介面）變數，但你想取出裡面的具體型別時，
// 就需要用「型別斷言」。
//
// 安全寫法（comma-ok 模式）：
//   value, ok := something.(string)
//   ok 是 bool，告訴你斷言是否成功
//
// 危險寫法（不建議）：
//   value := something.(string)
//   如果型別不對，直接 panic！

// describeValue 示範安全的型別斷言（comma-ok 模式）
func describeValue(v any) { // 接收 any 型別的參數
	// 嘗試斷言為 string
	str, ok := v.(string) // comma-ok 模式：str 是值，ok 是是否成功
	if ok {               // ok 為 true 代表 v 確實是 string
		fmt.Printf("  是字串！值為: %q，長度: %d\n", str, len(str)) // %q 會加引號
		return // 提早返回，不需要繼續檢查
	}

	// 嘗試斷言為 int
	num, ok := v.(int) // 嘗試把 v 當作 int 取出
	if ok {            // ok 為 true 代表 v 確實是 int
		fmt.Printf("  是整數！值為: %d，平方: %d\n", num, num*num) // 印出整數資訊
		return // 提早返回
	}

	// 以上都不是
	fmt.Printf("  是其他型別: %T，值: %v\n", v, v) // 用 %T 印出實際型別名稱
}

// ========================================
// 6. 型別 Switch — 根據型別做不同處理
// ========================================
// 當你需要判斷多種型別時，型別 switch 比連續的型別斷言更優雅。
// 語法：switch val := v.(type) { ... }
// 注意：v.(type) 是特殊語法，只能在 switch 裡面用！

// classifyValue 使用型別 switch 來分類不同型別的值
func classifyValue(v any) string { // 接收 any，回傳字串描述
	switch val := v.(type) { // v.(type) 取出具體型別，val 在每個 case 中是對應的型別
	case int:     // 如果 v 是 int
		return fmt.Sprintf("整數 %d", val)     // val 在這裡自動是 int 型別
	case string:  // 如果 v 是 string
		return fmt.Sprintf("字串 %q", val)     // val 在這裡自動是 string 型別
	case bool:    // 如果 v 是 bool
		return fmt.Sprintf("布林 %t", val)     // val 在這裡自動是 bool 型別
	case float64: // 如果 v 是 float64
		return fmt.Sprintf("浮點數 %.2f", val) // val 在這裡自動是 float64 型別
	case Shape:   // 也可以判斷是否滿足某個介面！
		return fmt.Sprintf("形狀（面積=%.2f）", val.Area()) // val 在這裡是 Shape 介面型別
	default:      // 以上全都不符合
		return fmt.Sprintf("未知型別 %T", val) // %T 印出實際的型別名稱
	}
}

// ========================================
// 7. 實際範例：Blog 專案的 Repository 模式
// ========================================
// 在 Clean Architecture 中，業務邏輯（Usecase）不應該直接依賴資料庫。
// 而是透過「介面」來定義需要的功能，
// 然後由 Repository 來決定怎麼實作。
//
// 這就像餐廳（Usecase）跟供應商（Repository）簽合約：
// 合約上寫著「你要能提供雞肉和蔬菜」，
// 至於供應商是去市場買的還是自己養的，餐廳不管。

// User 是使用者實體（模擬部落格專案的 domain.User）
type User struct { // 定義使用者結構體
	ID    int    // 使用者 ID（主鍵）
	Name  string // 使用者名稱
	Email string // 電子郵件
}

// UserRepository 定義了「使用者資料操作」的合約（介面）
// 對應部落格專案的 domain.UserRepository
// 重要：這個介面定義在「使用方」（Usecase 層），不是「實作方」（Repository 層）
type UserRepository interface { // 使用者倉庫介面
	Create(user *User) error            // 建立使用者，失敗回傳 error
	FindByID(id int) (*User, error)     // 用 ID 查找使用者，找不到回傳 error
}

// --- 實作 1：記憶體儲存（用於測試）---
// MemoryUserRepository 把資料存在 map 中，程式結束就沒了
// 適合用於單元測試，不需要真正的資料庫

// MemoryUserRepository 用 map 模擬資料庫（記憶體版本）
type MemoryUserRepository struct { // 記憶體版本的 Repository 結構體
	users  map[int]*User // 用 map 儲存使用者資料，key 是使用者 ID
	nextID int           // 下一個可用的 ID（模擬自動遞增）
}

// NewMemoryUserRepository 是工廠函式，建立並初始化 MemoryUserRepository
func NewMemoryUserRepository() *MemoryUserRepository { // 回傳指標型別
	return &MemoryUserRepository{   // 建立結構體並回傳其指標
		users:  make(map[int]*User), // 用 make 初始化 map（map 一定要初始化！）
		nextID: 1,                   // ID 從 1 開始
	}
}

// Create 在記憶體中建立使用者（實作 UserRepository 介面的 Create 方法）
func (r *MemoryUserRepository) Create(user *User) error { // 指標接收者，因為要修改 r 的狀態
	user.ID = r.nextID             // 分配 ID 給使用者
	r.nextID++                     // ID 遞增，下次用新的
	r.users[user.ID] = user        // 把使用者存進 map
	fmt.Printf("  [Memory] 已儲存使用者：%s (ID=%d)\n", user.Name, user.ID) // 印出確認
	return nil                     // nil 代表沒有錯誤
}

// FindByID 在記憶體中查找使用者（實作 UserRepository 介面的 FindByID 方法）
func (r *MemoryUserRepository) FindByID(id int) (*User, error) { // 回傳使用者指標和錯誤
	user, exists := r.users[id] // 從 map 取值，exists 是 bool（是否存在）
	if !exists {                // 如果找不到
		return nil, fmt.Errorf("使用者 ID=%d 不存在", id) // 回傳 nil 使用者和錯誤訊息
	}
	fmt.Printf("  [Memory] 找到使用者：%s\n", user.Name) // 印出找到的使用者
	return user, nil // 回傳使用者，錯誤為 nil（沒問題）
}

// --- 實作 2：模擬資料庫儲存（用於正式環境）---
// DatabaseUserRepository 模擬真正的資料庫操作
// 在實際專案中，這裡會有一個 *gorm.DB 欄位來連接資料庫

// DatabaseUserRepository 模擬資料庫版本的 Repository
type DatabaseUserRepository struct { // 資料庫版本的結構體
	// 實際專案中會有：db *gorm.DB
	// 這裡為了教學簡化，不放任何欄位
}

// Create 模擬用 SQL 在資料庫中建立使用者（實作 UserRepository 介面）
func (r *DatabaseUserRepository) Create(user *User) error { // 資料庫版本的 Create
	fmt.Printf("  [Database] INSERT INTO users ... (%s)\n", user.Name) // 模擬 SQL INSERT
	user.ID = 1 // 模擬資料庫自動分配的 ID
	return nil   // 假設操作成功，回傳 nil
}

// FindByID 模擬用 SQL 在資料庫中查找使用者（實作 UserRepository 介面）
func (r *DatabaseUserRepository) FindByID(id int) (*User, error) { // 資料庫版本的 FindByID
	fmt.Printf("  [Database] SELECT * FROM users WHERE id = %d\n", id) // 模擬 SQL SELECT
	return &User{ID: id, Name: "來自資料庫的使用者"}, nil // 回傳模擬的使用者資料
}

// ========================================
// 8. UserService 依賴「介面」而非「具體實作」
// ========================================
// 這就是「依賴反轉原則」（Dependency Inversion Principle）：
// 高層模組（Service）不依賴低層模組（具體的 Repository），
// 兩者都依賴「抽象」（介面）。
//
// 好處：
// - 測試時注入 MemoryRepository（不需要真資料庫）
// - 生產環境注入 DatabaseRepository（連接真資料庫）
// - Service 的程式碼完全不用改！

// UserService 代表業務邏輯層（對應部落格專案的 Usecase 層）
type UserService struct {       // 使用者服務結構體
	repo UserRepository // 欄位型別是「介面」！不是 *MemoryUserRepository 或 *DatabaseUserRepository
}

// NewUserService 建立 UserService，注入任何滿足 UserRepository 介面的實作
// 這就是「依賴注入」（Dependency Injection）：由外部決定注入哪個實作
func NewUserService(repo UserRepository) *UserService { // 參數是介面型別
	return &UserService{repo: repo} // 把注入的 repo 存到結構體欄位中
}

// RegisterUser 是業務邏輯：註冊新使用者
func (s *UserService) RegisterUser(name, email string) (*User, error) { // 業務方法
	user := &User{Name: name, Email: email} // 建立使用者物件（指標）
	if err := s.repo.Create(user); err != nil { // 呼叫 repo 的 Create（不管是哪種實作）
		return nil, err // 如果出錯，向上傳遞錯誤
	}
	return user, nil // 成功，回傳使用者
}

// GetUser 是業務邏輯：根據 ID 取得使用者
func (s *UserService) GetUser(id int) (*User, error) { // 業務方法
	return s.repo.FindByID(id) // 直接呼叫 repo 的 FindByID（具體行為取決於注入的實作）
}

// ========================================
// 主程式入口
// ========================================

func main() { // 程式的入口點

	// ========================================
	// 示範 1：介面與多型
	// ========================================
	fmt.Println("=== 1. 介面與多型 ===") // 印出區塊標題

	// 建立一個 Shape 介面的切片（slice），可以混合裝不同形狀
	shapes := []Shape{                       // Shape 切片，裡面可以放任何滿足 Shape 的型別
		Circle{Radius: 5},               // 圓形，半徑 5
		Rectangle{Width: 4, Height: 6},  // 矩形，寬 4 高 6
	}

	// 用 for range 遍歷所有形狀
	for _, shape := range shapes { // shape 的型別是 Shape 介面
		fmt.Printf("\n%T:\n", shape) // %T 印出實際的具體型別（Circle 或 Rectangle）
		printShapeInfo(shape)        // 同一個函式處理不同型別 — 這就是多型！
	}

	// ========================================
	// 示範 2：空介面 (any) — 接受任何型別
	// ========================================
	fmt.Println("\n=== 2. 空介面 (any) ===") // 印出區塊標題

	printAnything(42)                // 傳入整數 → 可以！
	printAnything("hello")           // 傳入字串 → 可以！
	printAnything(true)              // 傳入布林值 → 可以！
	printAnything(Circle{Radius: 3}) // 傳入結構體 → 也可以！any 接受一切

	// ========================================
	// 示範 3：型別斷言（安全的 comma-ok 模式）
	// ========================================
	fmt.Println("\n=== 3. 型別斷言（comma-ok 模式）===") // 印出區塊標題

	var something any = "Hello, Go!" // 宣告一個 any 變數，裡面裝了字串

	// 安全的型別斷言：使用兩個回傳值（comma-ok 模式）
	str, ok := something.(string) // 嘗試取出 string，ok 告訴你是否成功
	if ok {                       // ok == true → 斷言成功，v 確實是 string
		fmt.Println("  斷言成功！是字串:", str) // 安全地使用取出的字串
	}

	// 斷言失敗的情況
	num, ok := something.(int) // 嘗試取出 int（但 something 是 string）
	if !ok {                   // ok == false → 斷言失敗，something 不是 int
		fmt.Println("  斷言失敗！不是整數，num 的零值:", num) // num 會是 0（int 的零值）
	}

	// ⚠️ 危險寫法（不建議）：
	// num := something.(int)  // 如果型別不對，程式直接 panic 崩潰！
	// 所以永遠建議使用 comma-ok 模式來做型別斷言

	// 更多型別斷言範例
	fmt.Println("\n  更多型別斷言範例:") // 小標題
	describeValue("Go 語言")          // 傳入字串 → 會走 string 分支
	describeValue(42)                 // 傳入整數 → 會走 int 分支
	describeValue(3.14)               // 傳入浮點數 → 會走「其他型別」分支

	// ========================================
	// 示範 4：型別 Switch
	// ========================================
	fmt.Println("\n=== 4. 型別 Switch ===") // 印出區塊標題

	// 建立一個 any 切片，裝各種不同型別的值
	values := []any{42, "test", true, 3.14, Circle{Radius: 2}} // any 切片

	for _, v := range values { // 遍歷每個值
		result := classifyValue(v)            // 用型別 switch 分類
		fmt.Printf("  %v → %s\n", v, result) // 印出值和分類結果
	}

	// ========================================
	// 示範 5：依賴注入 — 使用記憶體 Repository（測試用）
	// ========================================
	fmt.Println("\n=== 5. 依賴注入：Memory Repository ===") // 印出區塊標題

	memoryRepo := NewMemoryUserRepository()  // 建立記憶體版本的 Repository
	service := NewUserService(memoryRepo)    // 注入記憶體實作到 UserService

	service.RegisterUser("Alice", "alice@example.com") // 用 Service 註冊使用者
	service.GetUser(1)                                  // 用 Service 查找使用者

	// ========================================
	// 示範 6：依賴注入 — 切換為資料庫 Repository（正式環境）
	// ========================================
	fmt.Println("\n=== 6. 依賴注入：Database Repository ===") // 印出區塊標題

	dbRepo := &DatabaseUserRepository{}   // 建立資料庫版本的 Repository
	service2 := NewUserService(dbRepo)    // 注入資料庫實作到 UserService

	service2.RegisterUser("Bob", "bob@example.com") // 同樣的 Service 方法
	service2.GetUser(1)                              // 同樣的 Service 方法

	// ✨ 關鍵重點：UserService 的程式碼一行都沒改！
	// 只是在 NewUserService() 時注入了不同的 Repository 實作。
	// 測試時用 MemoryRepository，正式環境用 DatabaseRepository。

	// ========================================
	// 示範 7：介面的零值是 nil
	// ========================================
	fmt.Println("\n=== 7. 介面的零值 ===") // 印出區塊標題

	var s Shape                  // 宣告一個 Shape 介面變數，沒有賦值
	fmt.Println("  Shape 的零值:", s) // 印出 <nil>（介面的零值是 nil）

	if s == nil { // 介面變數可以跟 nil 比較
		fmt.Println("  介面變數是 nil，呼叫方法會 panic！") // 提醒：nil 介面不能呼叫方法
	}
	// s.Area() ← 這行如果取消註解會 panic: nil pointer dereference

	// ========================================
	// 結語
	// ========================================
	fmt.Println("\n=== 學習完成！ ===")                                       // 結束訊息
	fmt.Println("介面是 Go 最強大的特性之一，也是 Clean Architecture 的基礎。") // 總結
	fmt.Println("下一課：錯誤處理（Error Handling）— Go 獨特的錯誤哲學。")     // 預告下一課
}
