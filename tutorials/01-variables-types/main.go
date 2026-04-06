// ============================================================
// 第一課：變數與型別（Variables & Types）
// ============================================================
// 這是你學 Go 語言的第一步！
// 在這堂課你會學到：
//   - 什麼是變數（variable）？ → 用來存放資料的「盒子」
//   - 什麼是型別（type）？    → 告訴電腦這個盒子裡放的是什麼種類的東西
//   - 怎麼建立變數、怎麼印出來看
//
// 執行方式：在終端機輸入
//   go run ./tutorials/01-variables-types
// ============================================================

package main // 每個 Go 檔案的第一行都要宣告自己屬於哪個「套件（package）」
// "main" 是特殊的套件名稱，代表「這是一個可以執行的程式」

import "fmt" // 引入 fmt 套件（format 的縮寫）
// fmt 是 Go 內建的套件，專門用來「印東西到螢幕上」
// 就像你用 console.log()（JavaScript）或 print()（Python）一樣

// main() 是程式的「入口點」——Go 執行程式時，永遠從這裡開始
func main() {

	// ========================================
	// 1. 用 var 宣告變數（最基本的方式）
	// ========================================
	// 語法：var 變數名稱 型別 = 值
	// 就像在說：「我要一個盒子，名字叫 name，裡面只能放文字，先放入 Alice」

	var name string = "Alice"  // string = 文字（字串），用雙引號包起來
	var age int = 25           // int = 整數（integer），沒有小數點的數字
	var height float64 = 165.5 // float64 = 浮點數，有小數點的數字（64 代表精度）
	var isStudent bool = true  // bool = 布林值，只有 true（是）或 false（否）

	// fmt.Println() 會把括號裡的東西印到螢幕上，印完會自動換行
	fmt.Println("=== 1. var 宣告 ===")
	fmt.Println("姓名:", name)       // 印出：姓名: Alice
	fmt.Println("年齡:", age)        // 印出：年齡: 25
	fmt.Println("身高:", height)     // 印出：身高: 165.5
	fmt.Println("是學生:", isStudent) // 印出：是學生: true

	// ========================================
	// 2. 短變數宣告 :=（最常用！）
	// ========================================
	// 語法：變數名稱 := 值
	// Go 會自動判斷「值是什麼型別」，不需要你寫出來
	// 注意：:= 只能在函式裡面用，不能在函式外面用

	city := "台北"   // Go 看到 "台北" 是文字，自動判斷為 string
	score := 95    // Go 看到 95 是整數，自動判斷為 int
	pi := 3.14159  // Go 看到有小數點，自動判斷為 float64
	passed := true // Go 看到 true，自動判斷為 bool

	fmt.Println("\n=== 2. := 短變數宣告（最常用）===")
	fmt.Println("城市:", city)   // 印出：城市: 台北
	fmt.Println("分數:", score)  // 印出：分數: 95
	fmt.Println("圓周率:", pi)    // 印出：圓周率: 3.14159
	fmt.Println("通過:", passed) // 印出：通過: true

	// ========================================
	// 3. 零值（Zero Value）——Go 的超棒設計
	// ========================================
	// 如果你宣告了變數但「沒有給值」，Go 不會報錯
	// 它會自動給一個「零值」（就像考試沒寫答案，老師給你 0 分）
	//
	// 不同型別的零值：
	//   int    → 0
	//   float64 → 0.0
	//   string → ""（空字串，什麼都沒有）
	//   bool   → false

	var defaultInt int       // 沒給值 → 自動是 0
	var defaultFloat float64 // 沒給值 → 自動是 0
	var defaultString string // 沒給值 → 自動是 ""
	var defaultBool bool     // 沒給值 → 自動是 false

	fmt.Println("\n=== 3. 零值（Zero Value）===")
	fmt.Println("int 零值:", defaultInt)               // 0
	fmt.Println("float64 零值:", defaultFloat)         // 0
	fmt.Printf("string 零值: \"%s\"\n", defaultString) // ""
	fmt.Println("bool 零值:", defaultBool)             // false

	// ========================================
	// 4. 常數（Constants）——不能改的值
	// ========================================
	// 用 const 宣告的值，一旦設定就「永遠不能修改」
	// 適合用在不會變的設定值，像是 API 版本號、最大重試次數等

	const maxRetries = 3       // 最大重試次數（不能改）
	const appName = "Blog API" // 應用程式名稱（不能改）
	const version = 1.0        // 版本號（不能改）

	fmt.Println("\n=== 4. 常數（Constants）===")
	fmt.Println("應用名稱:", appName)      // Blog API
	fmt.Println("版本:", version)        // 1
	fmt.Println("最大重試次數:", maxRetries) // 3

	// 如果你取消下面這行的註解，Go 會報錯：「cannot assign to maxRetries」
	// maxRetries = 5  // ← 編譯錯誤！常數不能修改

	// ========================================
	// 5. 多重賦值——一次宣告好幾個
	// ========================================
	// 用小括號 () 把多個 var 包在一起，比較整齊

	var (
		firstName = "小明" // Go 自動推導為 string
		lastName  = "王"  // Go 自動推導為 string
		email     = "xiaoming@example.com"
	)

	fmt.Println("\n=== 5. 多重賦值 ===")
	fmt.Println("名:", firstName) // 小明
	fmt.Println("姓:", lastName)  // 王
	fmt.Println("信箱:", email)    // xiaoming@example.com

	// 也可以一行同時宣告多個變數（用逗號分隔）
	x, y, z := 1, 2, 3                // 同時建立三個 int 變數
	fmt.Println("x, y, z =", x, y, z) // x, y, z = 1 2 3

	// ========================================
	// 6. 型別轉換——Go 不會自動幫你轉
	// ========================================
	// 很多語言會「自動」把 int 變成 float（隱式轉換）
	// 但 Go 不會！你必須「明確告訴 Go」要轉換
	// 這是為了避免意外的精度遺失

	intValue := 42                             // 這是 int
	floatValue := float64(intValue)            // 把 int 轉成 float64，語法：目標型別(值)
	stringValue := fmt.Sprintf("%d", intValue) // 把 int 轉成 string（用 Sprintf 格式化）

	fmt.Println("\n=== 6. 型別轉換 ===")
	fmt.Println("int:", intValue)          // 42
	fmt.Println("轉為 float64:", floatValue) // 42
	fmt.Println("轉為 string:", stringValue) // "42"

	// 注意：下面這行如果取消註解會報錯
	// var f float64 = intValue  // ← 錯誤！不能把 int 直接放進 float64 的盒子

	// ========================================
	// 7. Go 的整數家族——不只有 int
	// ========================================
	// Go 有很多種整數型別，差別在於「能存多大的數字」和「能不能存負數」
	//
	// 有號整數（可以是負數）：
	//   int8   → -128 ~ 127
	//   int16  → -32768 ~ 32767
	//   int32  → 約 -21 億 ~ 21 億
	//   int64  → 超級大的數字
	//   int    → 在 64 位元電腦上等於 int64
	//
	// 無號整數（只能是 0 或正數）：
	//   uint8  → 0 ~ 255（也叫 byte）
	//   uint16 → 0 ~ 65535
	//   uint32 → 0 ~ 約 42 億
	//   uint64 → 超級大的正數
	//   uint   → 在 64 位元電腦上等於 uint64
	//
	// 在我們的部落格專案中，ID 用的是 uint（無號整數）
	// 因為 ID 不可能是負數

	var id uint = 1            // uint：無號整數，只能是 0 或正數
	var smallNumber int8 = 127 // int8 最大只能存 127

	fmt.Println("\n=== 7. 整數家族 ===")
	fmt.Println("使用者 ID (uint):", id)
	fmt.Println("int8 最大值:", smallNumber)
	fmt.Printf("int 佔幾個位元: 看你的電腦，通常是 64 位元\n")

	// ========================================
	// 8. fmt 套件常用函式
	// ========================================
	// fmt 是你最常用的套件，以下是三個最重要的函式：

	fmt.Println("\n=== 8. fmt 套件常用函式 ===")

	// (1) Println：印出來 + 自動換行
	//     最簡單，日常最常用
	fmt.Println("這是 Println，印完會自動換行")

	// (2) Printf：格式化輸出（不會自動換行，要自己加 \n）
	//     用 %s、%d 等「佔位符」來插入變數的值
	//     %s = 字串（string）
	//     %d = 整數（decimal）
	//     %f = 浮點數（float），%.1f 表示小數點後 1 位
	//     %v = 任何型別都可以用（value）
	//     %T = 印出「型別名稱」（Type）
	//     %t = 布林值（true/false）
	//     \n = 換行符號
	fmt.Printf("姓名：%s，年齡：%d，身高：%.1f\n", name, age, height)

	// (3) Sprintf：跟 Printf 一樣格式化，但「不會印出來」
	//     它會把結果「存成字串」回傳給你
	greeting := fmt.Sprintf("你好，%s！歡迎來到 %s", name, appName)
	fmt.Println(greeting) // 你好，Alice！歡迎來到 Blog API

	// 超實用：%T 可以查看任何變數的型別
	fmt.Println("\n--- 型別檢查 ---")
	fmt.Printf("name 的型別是 %T\n", name)           // string
	fmt.Printf("age 的型別是 %T\n", age)             // int
	fmt.Printf("height 的型別是 %T\n", height)       // float64
	fmt.Printf("isStudent 的型別是 %T\n", isStudent) // bool
	fmt.Printf("id 的型別是 %T\n", id)               // uint
}
