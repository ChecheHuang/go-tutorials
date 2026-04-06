// ============================================================
// 第四課：結構體與方法（Structs & Methods）
// ============================================================
// 結構體就像「身分證」——把一個人的所有相關資料放在一張卡片上。
// 方法就像「技能」——這張卡片的主人能做什麼事。
//
// 在其他語言（Java、Python、C#）中，你會用 class 來定義物件。
// Go 沒有 class，而是用 struct（結構體）+ method（方法）來達成。
// 這種方式更簡單、更直接，沒有繼承的複雜性。
//
// 想像一下：
//   class Person     →  type Person struct { ... }
//   person.sayHi()   →  func (p Person) SayHi() { ... }
//   new Person(...)  →  NewPerson(...)
//   extends          →  結構體嵌套（embedding）
//
// 你會學到：
//   1. 結構體的定義與建立（4 種方式）
//   2. 方法（值接收者 vs 指標接收者）——深度理解
//   3. 結構體嵌套（Embedding）——Go 的「繼承」替代方案
//   4. 建構函式模式（Constructor Pattern）
//   5. 結構體標籤（Struct Tags）——JSON、資料庫映射
//   6. 結構體比較
//
// 執行方式：go run ./tutorials/04-structs-methods
// ============================================================

package main // 每個可執行程式都必須屬於 main 套件

import ( // import 區塊：引入需要的套件
	"encoding/json" // 處理 JSON 序列化與反序列化
	"fmt"           // 格式化輸出，印東西到螢幕上
	"time"          // 時間相關功能
)

// ============================================================
// 1. 結構體的定義（Struct Definition）
// ============================================================
// 結構體就是「把相關的資料打包在一起」。
//
// 想像你要描述一個部落格的使用者：
//   - 使用者有 ID、帳號、Email、年齡、註冊時間
//   - 這些資料彼此相關，應該放在一起
//   - 就像身分證上面有姓名、生日、地址——都是「描述同一個人」的資料
//
// 如果不用結構體，你得用一堆零散的變數：
//   var userName string
//   var userEmail string
//   var userAge int
//   ... 如果有 100 個使用者呢？這會變得超級混亂！
//
// 結構體讓你把這些資料「打包」成一個自訂型別。

// User 代表部落格系統中的一個使用者
// type → 定義新型別的關鍵字
// User → 型別名稱（大寫開頭 = 公開，其他套件也能使用）
// struct → 表示這是一個結構體
type User struct {
	ID        int       // 使用者的唯一識別碼（就像身分證字號）
	Username  string    // 使用者帳號（就像身分證上的姓名）
	Email     string    // 電子信箱（就像身分證上的地址）
	Age       int       // 年齡
	IsActive  bool      // 帳號是否啟用（true = 啟用，false = 停用）
	CreatedAt time.Time // 註冊時間（使用 time 套件的 Time 型別）
}

// ============================================================
// 2. 方法（Methods）——值接收者（Value Receiver）
// ============================================================
// 方法和函式很像，唯一的差別是方法有一個「接收者（receiver）」。
// 接收者寫在 func 和方法名之間，用小括號包住。
//
// 語法拆解：
//   func (u User) DisplayName() string
//   │     │  │     │              │
//   │     │  │     │              └── 回傳值型別
//   │     │  │     └────────────── 方法名稱
//   │     │  └──────────────────── 接收者型別
//   │     └─────────────────────── 接收者變數名（慣例用型別首字母小寫）
//   └───────────────────────────── func 關鍵字
//
// 「值接收者」(u User) 的意思：
//   呼叫這個方法時，Go 會把原始結構體「複製一份」給 u
//   你對 u 做的任何修改，都不會影響原始的結構體
//   就像你影印了一份身分證——在影印本上塗改，正本不受影響

// DisplayName 回傳使用者的顯示名稱
// 這是「值接收者」方法——只讀取資料，不修改
func (u User) DisplayName() string {
	return fmt.Sprintf("%s (%s)", u.Username, u.Email) // 格式化成「帳號 (信箱)」
}

// IsAdult 檢查使用者是否年滿 18 歲
// 這也是「值接收者」——只讀取 u.Age，不需要修改任何東西
func (u User) IsAdult() bool {
	return u.Age >= 18 // 如果年齡 >= 18 就回傳 true
}

// Summary 回傳使用者的摘要資訊
// 值接收者適合用在「只需要讀取、不需要修改」的場景
func (u User) Summary() string {
	status := "停用"  // 先假設帳號是停用的
	if u.IsActive { // 如果 IsActive 是 true
		status = "啟用" // 改成「啟用」
	}
	// Sprintf 類似其他語言的字串格式化（template literal）
	return fmt.Sprintf("[%d] %s | %s | 狀態：%s", u.ID, u.Username, u.Email, status)
}

// ============================================================
// 3. 方法——指標接收者（Pointer Receiver）
// ============================================================
// 「指標接收者」(u *User) 的意思：
//   u 不是副本，而是指向原始結構體的「地址」
//   透過 u 修改的東西，會直接反映在原始結構體上
//   就像你拿到的是身分證的「正本」——你改了什麼，就真的改了
//
// ★★★ 什麼時候該用指標接收者？★★★
//
// 1. 需要修改接收者的值
//    → 這是最明顯的理由。值接收者操作的是副本，改了也白改。
//
// 2. 結構體很大（有很多欄位或包含大型切片/映射）
//    → 每次呼叫值接收者方法都會複製整個結構體，浪費記憶體和效能。
//    → 指標只是一個記憶體地址（8 bytes），複製起來幾乎零成本。
//
// 3. 一致性原則（最重要的實務考量！）
//    → 如果一個型別的「任何一個」方法需要指標接收者，
//      那「所有」方法都應該用指標接收者。
//    → 為什麼？因為混用會讓使用者困惑，也可能導致介面無法實作。
//
// 4. 方法可能會被介面使用
//    → *User 實作了某個介面，User 不一定實作了（第六課會詳細講）
//
// ★ 簡單的判斷法則 ★
//   → 如果是「讀取資料」的小結構體 → 值接收者 OK
//   → 其他情況 → 一律用指標接收者（這是業界最常見的做法）

// SetEmail 修改使用者的信箱
// (u *User) 是「指標接收者」——u 指向原始結構體
func (u *User) SetEmail(newEmail string) {
	u.Email = newEmail // 直接修改原始結構體的 Email 欄位
}

// Activate 啟用使用者帳號
// 需要修改 IsActive 的值，所以用指標接收者
func (u *User) Activate() {
	u.IsActive = true                          // 把帳號狀態改為啟用
	fmt.Printf("  ✓ 使用者 %s 已啟用\n", u.Username) // 印出確認訊息
}

// Deactivate 停用使用者帳號
// 和 Activate 配對，也用指標接收者
func (u *User) Deactivate() {
	u.IsActive = false                         // 把帳號狀態改為停用
	fmt.Printf("  ✗ 使用者 %s 已停用\n", u.Username) // 印出確認訊息
}

// UpdateAge 更新年齡並印出變更紀錄
func (u *User) UpdateAge(newAge int) {
	oldAge := u.Age                              // 先記住舊的年齡
	u.Age = newAge                               // 更新為新的年齡
	fmt.Printf("  年齡：%d → %d\n", oldAge, newAge) // 印出變更紀錄
}

// ============================================================
// 4. 值接收者 vs 指標接收者——為什麼這麼重要？
// ============================================================
// 這邊用一個特別的方法來「證明」兩者的差異。
// 在實務中你不會寫這種方法，但這是理解的關鍵。

// TryChangeNameValue 嘗試用「值接收者」修改名稱
// 劇透：這不會成功！因為 u 只是副本
func (u User) TryChangeNameValue(newName string) {
	u.Username = newName // 修改的是副本的 Username
	// 方法結束後，副本被丟棄，原始值完全不受影響
	fmt.Printf("  [值接收者內部] Username = %s（這是副本）\n", u.Username)
}

// TryChangeNamePointer 用「指標接收者」修改名稱
// 這會成功！因為 u 指向原始結構體
func (u *User) TryChangeNamePointer(newName string) {
	u.Username = newName // 修改的是原始結構體的 Username
	fmt.Printf("  [指標接收者內部] Username = %s（這是原始值）\n", u.Username)
}

// ============================================================
// 5. 結構體嵌套（Embedding）——Go 的「組合」哲學
// ============================================================
// Go 沒有繼承（inheritance），而是用「組合」（composition）。
// Go 的哲學：「組合優於繼承」（Composition over Inheritance）
//
// 想像一下：
//   「員工」不是一種特殊的「人」（繼承的思維）
//   「員工」是一個「人」加上「工作資訊」（組合的思維）
//
// 嵌套讓你可以直接存取被嵌套結構體的欄位和方法，
// 就像它們是自己的一樣，非常方便！

// Address 代表地址資訊
type Address struct {
	City    string // 城市
	Country string // 國家
}

// FullAddress 回傳完整地址字串
// 這是 Address 的方法
func (a Address) FullAddress() string {
	return fmt.Sprintf("%s, %s", a.City, a.Country) // 格式化成「城市, 國家」
}

// Employee 代表員工，「嵌套」了 User 和 Address
// 注意：User 和 Address 沒有寫欄位名稱，這叫「匿名欄位」
// 匿名欄位的魔法：Employee 可以直接使用 User 和 Address 的欄位和方法
type Employee struct {
	User               // 嵌套 User（匿名欄位）——Employee「擁有」User 的所有欄位和方法
	Address            // 嵌套 Address（匿名欄位）——Employee「擁有」Address 的所有欄位和方法
	Department string  // 部門（Employee 自己的欄位）
	Salary     float64 // 薪資（Employee 自己的欄位）
}

// ============================================================
// 6. 建構函式模式（Constructor Pattern）
// ============================================================
// Go 沒有 class，所以沒有 constructor（建構子）。
// 取而代之的是一個「普通的函式」，慣例命名為 NewXxx。
//
// 為什麼需要建構函式？
//   1. 設定預設值（例如自動填入建立時間）
//   2. 驗證參數（例如檢查 email 不能是空的）
//   3. 確保結構體在「一致的狀態」下被建立
//
// 慣例：
//   - 函式名稱以 New 開頭：NewUser、NewArticle
//   - 回傳指標（*User）而不是值（User），原因：
//     a. 避免複製大型結構體
//     b. 讓呼叫者可以直接使用指標接收者的方法
//     c. 這是 Go 社群的慣例

// NewUser 是 User 的建構函式
// 接收必要的參數，回傳一個初始化好的 *User（User 的指標）
func NewUser(id int, username, email string, age int) *User {
	return &User{ // &User{...} 建立 User 並取得它的指標
		ID:        id,         // 設定 ID
		Username:  username,   // 設定帳號
		Email:     email,      // 設定信箱
		Age:       age,        // 設定年齡
		IsActive:  true,       // 預設為啟用狀態
		CreatedAt: time.Now(), // 自動填入當前時間
	}
}

// NewEmployee 是 Employee 的建構函式
func NewEmployee(id int, name, email string, age int, city, country, dept string, salary float64) *Employee {
	return &Employee{ // 建立 Employee 並回傳指標
		User: User{ // 初始化嵌套的 User
			ID:        id,         // 設定使用者 ID
			Username:  name,       // 設定使用者名稱
			Email:     email,      // 設定使用者信箱
			Age:       age,        // 設定年齡
			IsActive:  true,       // 預設啟用
			CreatedAt: time.Now(), // 自動填入當前時間
		},
		Address: Address{ // 初始化嵌套的 Address
			City:    city,    // 設定城市
			Country: country, // 設定國家
		},
		Department: dept,   // 設定部門
		Salary:     salary, // 設定薪資
	}
}

// ============================================================
// 7. 結構體標籤（Struct Tags）
// ============================================================
// 標籤是附加在結構體欄位上的「元資料」（metadata）。
// 它們不會影響程式邏輯，而是給其他套件（如 json、gorm）讀取用的。
//
// 語法：`key:"value" key2:"value2"`（用反引號包住）
//
// 常見的標籤：
//   json:"name"      → 控制 JSON 序列化時的欄位名稱
//   json:"-"         → 這個欄位不會出現在 JSON 中（隱藏它）
//   json:",omitempty" → 如果是零值就不輸出
//   binding:"required" → Gin 框架用：此欄位為必填
//   gorm:"primaryKey" → GORM 用：標記為主鍵
//   gorm:"index"      → GORM 用：建立資料庫索引

// Article 代表部落格文章——示範結構體標籤的用法
type Article struct {
	ID        int       `json:"id"`                       // JSON 輸出時欄位名叫 "id"（小寫）
	Title     string    `json:"title" binding:"required"` // JSON 叫 "title"，且為必填欄位
	Content   string    `json:"content"`                  // JSON 叫 "content"
	AuthorID  int       `json:"author_id" gorm:"index"`   // JSON 叫 "author_id"，資料庫建索引
	Views     int       `json:"views,omitempty"`          // 如果 Views 是 0，JSON 中不顯示
	Secret    string    `json:"-"`                        // json:"-" → 永遠不會出現在 JSON 中
	CreatedAt time.Time `json:"created_at"`               // JSON 叫 "created_at"
}

// NewArticle 是 Article 的建構函式
func NewArticle(id int, title, content string, authorID int) *Article {
	return &Article{ // 建立 Article 並回傳指標
		ID:        id,         // 設定文章 ID
		Title:     title,      // 設定標題
		Content:   content,    // 設定內容
		AuthorID:  authorID,   // 設定作者 ID
		Views:     0,          // 瀏覽次數初始為 0
		Secret:    "內部備註：草稿",  // 這個不會出現在 JSON 中
		CreatedAt: time.Now(), // 自動填入建立時間
	}
}

// ============================================================
// 主程式
// ============================================================

func main() { // main 是程式的進入點

	// ========================================
	// 示範 1：建立結構體的四種方式
	// ========================================
	fmt.Println("========================================") // 分隔線
	fmt.Println("1. 建立結構體的四種方式")                            // 標題
	fmt.Println("========================================") // 分隔線

	// 方式一：指定欄位名稱（★ 最推薦！★）
	// 優點：清楚、不怕欄位順序改變、可以只填部分欄位
	user1 := User{ // 用 := 建立並賦值
		ID:       1,                // 指定 ID
		Username: "alice",          // 指定帳號
		Email:    "alice@blog.com", // 指定信箱
		Age:      25,               // 指定年齡
		IsActive: true,             // 指定為啟用
	} // CreatedAt 沒寫，自動是零值（time.Time 的零值）
	fmt.Println("方式一（指定欄位）:", user1) // 印出 user1 的所有欄位

	// 方式二：按順序賦值（⚠ 不推薦！）
	// 缺點：必須記住所有欄位的順序，新增欄位時容易出錯
	user2 := User{2, "bob", "bob@blog.com", 30, true, time.Now()} // 按 ID, Username, Email, Age, IsActive, CreatedAt 的順序
	fmt.Println("方式二（按順序）:  ", user2)                             // 印出 user2

	// 方式三：先宣告再逐一賦值
	// 適合需要根據條件決定值的場景
	var user3 User                   // 宣告一個 User 變數（所有欄位都是零值）
	user3.ID = 3                     // 設定 ID（零值是 0，現在改成 3）
	user3.Username = "charlie"       // 設定帳號（零值是 ""，現在改成 "charlie"）
	user3.Email = "charlie@blog.com" // 設定信箱
	user3.Age = 17                   // 設定年齡
	user3.IsActive = false           // 設定為停用（bool 的零值就是 false，其實不寫也一樣）
	fmt.Println("方式三（逐一賦值）:", user3) // 印出 user3

	// 方式四：使用建構函式（★ 正式專案中最推薦！★）
	// 優點：封裝初始化邏輯、設定預設值、可以加驗證
	user4 := NewUser(4, "diana", "diana@blog.com", 28) // 呼叫建構函式
	fmt.Println("方式四（建構函式）:", user4)                   // 注意：印出來是記憶體地址（因為是指標 *User）
	fmt.Println("  解引用後的值:   ", *user4)                // 用 * 解引用，看到實際內容

	// ========================================
	// 示範 2：呼叫方法
	// ========================================
	fmt.Println()                                           // 空行，讓輸出更美觀
	fmt.Println("========================================") // 分隔線
	fmt.Println("2. 呼叫方法")                                  // 標題
	fmt.Println("========================================") // 分隔線

	fmt.Println("顯示名稱:", user1.DisplayName())     // 呼叫值接收者方法
	fmt.Println("是否成年:", user1.IsAdult())         // alice 25 歲，回傳 true
	fmt.Println("摘要:    ", user1.Summary())       // 呼叫 Summary 方法
	fmt.Println()                                 // 空行
	fmt.Println("charlie 是否成年:", user3.IsAdult()) // charlie 17 歲，回傳 false

	// ========================================
	// 示範 3：值接收者 vs 指標接收者——親眼見證差異
	// ========================================
	fmt.Println()                                           // 空行
	fmt.Println("========================================") // 分隔線
	fmt.Println("3. 值接收者 vs 指標接收者")                         // 標題
	fmt.Println("========================================") // 分隔線

	// 建立一個測試用的使用者
	testUser := User{ // 建立測試用的 User
		ID:       99,                  // ID = 99
		Username: "original",          // 帳號 = "original"
		Email:    "original@test.com", // 信箱 = "original@test.com"
	}

	fmt.Println("原始值:", testUser.Username) // 印出原始的 Username

	// 嘗試用「值接收者」修改——不會成功
	fmt.Println("\n▶ 用值接收者嘗試修改：")                    // 提示說明
	testUser.TryChangeNameValue("modified_by_value") // 呼叫值接收者方法
	fmt.Println("  修改後的原始值:", testUser.Username)     // 原始值沒變！還是 "original"

	// 嘗試用「指標接收者」修改——會成功
	fmt.Println("\n▶ 用指標接收者嘗試修改：")                   // 提示說明
	testUser.TryChangeNamePointer("modified_by_ptr") // 呼叫指標接收者方法
	fmt.Println("  修改後的原始值:", testUser.Username)     // 原始值變了！變成 "modified_by_ptr"

	// ========================================
	// 示範 4：指標接收者的實際應用
	// ========================================
	fmt.Println()                                           // 空行
	fmt.Println("========================================") // 分隔線
	fmt.Println("4. 指標接收者的實際應用")                            // 標題
	fmt.Println("========================================") // 分隔線

	fmt.Println("修改前 Email:", user1.Email) // 印出修改前的 Email
	user1.SetEmail("alice.new@blog.com")   // 用指標接收者方法修改 Email
	fmt.Println("修改後 Email:", user1.Email) // 確認 Email 已被修改

	fmt.Println()                        // 空行
	user1.Deactivate()                   // 停用帳號
	fmt.Println("帳號狀態:", user1.IsActive) // 印出 false
	user1.Activate()                     // 啟用帳號
	fmt.Println("帳號狀態:", user1.IsActive) // 印出 true

	// ========================================
	// 示範 5：結構體嵌套
	// ========================================
	fmt.Println()                                           // 空行
	fmt.Println("========================================") // 分隔線
	fmt.Println("5. 結構體嵌套（Embedding）")                      // 標題
	fmt.Println("========================================") // 分隔線

	// 使用建構函式建立 Employee
	emp := NewEmployee( // 呼叫 Employee 的建構函式
		10,                     // ID
		"王小明",                  // 使用者名稱
		"xiaoming@company.com", // 信箱
		30,                     // 年齡
		"台北",                   // 城市
		"台灣",                   // 國家
		"工程部",                  // 部門
		85000,                  // 薪資
	)

	// 嵌套的魔法：可以直接存取 User 和 Address 的欄位！
	fmt.Println("帳號:", emp.Username)   // 直接存取！不用寫 emp.User.Username
	fmt.Println("信箱:", emp.Email)      // 直接存取 User 的 Email
	fmt.Println("城市:", emp.City)       // 直接存取 Address 的 City
	fmt.Println("部門:", emp.Department) // Employee 自己的欄位
	fmt.Println("薪資:", emp.Salary)     // Employee 自己的欄位

	// 嵌套的方法也可以直接呼叫！
	fmt.Println("顯示名稱:", emp.DisplayName()) // 直接呼叫 User 的方法！
	fmt.Println("完整地址:", emp.FullAddress()) // 直接呼叫 Address 的方法！

	// 如果需要明確指定是哪個嵌套結構體的欄位，也可以寫完整路徑
	fmt.Println("完整路徑:", emp.User.Email) // 用完整路徑存取也行

	// ========================================
	// 示範 6：結構體標籤與 JSON 序列化
	// ========================================
	fmt.Println()                                           // 空行
	fmt.Println("========================================") // 分隔線
	fmt.Println("6. 結構體標籤與 JSON")                           // 標題
	fmt.Println("========================================") // 分隔線

	// 建立一篇文章
	article := NewArticle(1, "Go 入門教學", "這是一篇入門文章...", user1.ID) // 呼叫建構函式

	// 把結構體轉成 JSON（序列化）
	// json.MarshalIndent 會把結構體轉成格式化的 JSON 字串
	// 參數：要轉的值, 前綴（通常空字串）, 縮排（用兩個空格）
	jsonData, err := json.MarshalIndent(article, "", "  ") // 轉成美化的 JSON
	if err != nil {                                        // 如果轉換出錯
		fmt.Println("JSON 轉換錯誤:", err) // 印出錯誤
		return                         // 提早結束程式
	}

	fmt.Println("JSON 輸出：")       // 提示
	fmt.Println(string(jsonData)) // 把 []byte 轉成 string 印出

	// 注意觀察 JSON 輸出：
	// - "id" 不是 "ID"（因為標籤寫了 json:"id"）
	// - "Secret" 完全不會出現（因為標籤寫了 json:"-"）
	// - "views" 不會出現（因為值是 0，且標籤有 omitempty）
	// - "author_id" 不是 "AuthorID"（標籤的命名轉換）
	fmt.Println("\n注意：")                                 // 提示
	fmt.Println("  - Secret 欄位沒有出現在 JSON 中（json:\"-\"）") // 解說
	fmt.Println("  - Views 為 0 所以沒出現（omitempty）")        // 解說
	fmt.Println("  - 欄位名都是小寫蛇形命名（json tag 控制）")          // 解說

	// ========================================
	// 示範 7：結構體比較
	// ========================================
	fmt.Println()                                           // 空行
	fmt.Println("========================================") // 分隔線
	fmt.Println("7. 結構體比較")                                 // 標題
	fmt.Println("========================================") // 分隔線

	// 如果所有欄位都是「可比較的型別」，結構體就可以用 == 比較
	a := User{ID: 1, Username: "alice", Email: "alice@test.com"} // 建立 User a
	b := User{ID: 1, Username: "alice", Email: "alice@test.com"} // 建立 User b（和 a 相同）
	c := User{ID: 2, Username: "bob", Email: "bob@test.com"}     // 建立 User c（和 a 不同）

	fmt.Println("a == b:", a == b) // true：所有欄位都一樣
	fmt.Println("a == c:", a == c) // false：ID、Username、Email 都不同

	// ========================================
	// 總結
	// ========================================
	fmt.Println()                                           // 空行
	fmt.Println("========================================") // 分隔線
	fmt.Println("總結")                                       // 標題
	fmt.Println("========================================") // 分隔線
	fmt.Println("1. 結構體 = 把相關資料打包在一起（像身分證）")                // 重點 1
	fmt.Println("2. 方法 = 結構體的技能（像一個人能做的事）")                 // 重點 2
	fmt.Println("3. 值接收者 → 操作副本 → 不改原始值")                   // 重點 3
	fmt.Println("4. 指標接收者 → 操作原始值 → 會改原始值")                 // 重點 4
	fmt.Println("5. 嵌套 = Go 版的組合（不是繼承！）")                   // 重點 5
	fmt.Println("6. 建構函式 = NewXxx()，設定預設值和驗證")              // 重點 6
	fmt.Println("7. 結構體標籤 = 給其他套件的元資料（json、gorm）")          // 重點 7
	fmt.Println()                                           // 空行
	fmt.Println("下一課：指標（Pointers）——理解記憶體地址的秘密！")            // 預告下一課
}
