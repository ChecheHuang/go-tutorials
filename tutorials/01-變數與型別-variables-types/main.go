// 第一課：變數與型別
// 執行方式：go run main.go
package main

import "fmt"

func main() {
	// ========================================
	// 1. 變數宣告：使用 var 關鍵字
	// ========================================
	var name string = "Alice"    // 明確指定型別
	var age int = 25             // 整數型別
	var height float64 = 165.5   // 浮點數型別
	var isStudent bool = true    // 布林型別

	fmt.Println("=== var 宣告 ===")
	fmt.Println("姓名:", name)
	fmt.Println("年齡:", age)
	fmt.Println("身高:", height)
	fmt.Println("是學生:", isStudent)

	// ========================================
	// 2. 短變數宣告：使用 := （最常用）
	// ========================================
	// Go 會自動推導型別，不需要寫 var 和型別
	city := "台北"          // 自動推導為 string
	score := 95             // 自動推導為 int
	pi := 3.14159           // 自動推導為 float64
	passed := true          // 自動推導為 bool

	fmt.Println("\n=== := 短變數宣告 ===")
	fmt.Println("城市:", city)
	fmt.Println("分數:", score)
	fmt.Println("圓周率:", pi)
	fmt.Println("通過:", passed)

	// ========================================
	// 3. 零值（Zero Value）
	// ========================================
	// Go 的變數如果沒有賦值，會有預設的「零值」
	var defaultInt int        // 零值：0
	var defaultFloat float64  // 零值：0.0
	var defaultString string  // 零值：""（空字串）
	var defaultBool bool      // 零值：false

	fmt.Println("\n=== 零值 ===")
	fmt.Println("int 零值:", defaultInt)
	fmt.Println("float64 零值:", defaultFloat)
	fmt.Printf("string 零值: \"%s\"\n", defaultString)
	fmt.Println("bool 零值:", defaultBool)

	// ========================================
	// 4. 常數（Constants）
	// ========================================
	// 常數使用 const 宣告，值在編譯時確定，之後不能修改
	const maxRetries = 3
	const appName = "Blog API"
	const version = 1.0

	fmt.Println("\n=== 常數 ===")
	fmt.Println("應用名稱:", appName)
	fmt.Println("版本:", version)
	fmt.Println("最大重試次數:", maxRetries)

	// maxRetries = 5  // ← 這行會編譯錯誤！常數不能修改

	// ========================================
	// 5. 多重賦值
	// ========================================
	// Go 可以同時宣告或賦值多個變數
	var (
		firstName = "王"
		lastName  = "小明"
		email     = "xiaoming@example.com"
	)

	fmt.Println("\n=== 多重賦值 ===")
	fmt.Println("名:", firstName)
	fmt.Println("姓:", lastName)
	fmt.Println("信箱:", email)

	// 也可以一行宣告多個
	x, y, z := 1, 2, 3
	fmt.Println("x, y, z =", x, y, z)

	// ========================================
	// 6. 型別轉換
	// ========================================
	// Go 不會自動轉換型別，必須明確轉換
	intValue := 42
	floatValue := float64(intValue) // int → float64
	stringValue := fmt.Sprintf("%d", intValue) // int → string

	fmt.Println("\n=== 型別轉換 ===")
	fmt.Println("int:", intValue)
	fmt.Println("轉為 float64:", floatValue)
	fmt.Println("轉為 string:", stringValue)

	// ========================================
	// 7. fmt 套件常用函式
	// ========================================
	fmt.Println("\n=== fmt 常用函式 ===")

	// Println：印出後換行
	fmt.Println("這是 Println")

	// Printf：格式化輸出（不會自動換行）
	fmt.Printf("姓名：%s，年齡：%d，身高：%.1f\n", name, age, height)

	// Sprintf：格式化後回傳字串（不印出）
	greeting := fmt.Sprintf("你好，%s！", name)
	fmt.Println(greeting)

	// 常用格式化動詞：
	// %s = 字串
	// %d = 整數
	// %f = 浮點數（%.1f = 小數點後 1 位）
	// %v = 任何型別的預設格式
	// %T = 印出型別名稱
	// %t = 布林值

	fmt.Printf("型別檢查：%T, %T, %T, %T\n", name, age, height, isStudent)
}
