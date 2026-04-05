// 第二課：控制流程（if、for、switch）
// 執行方式：go run main.go
package main

import "fmt"

func main() {
	// ========================================
	// 1. if / else if / else
	// ========================================
	fmt.Println("=== if 條件判斷 ===")

	score := 85

	if score >= 90 {
		fmt.Println("成績：A")
	} else if score >= 80 {
		fmt.Println("成績：B")
	} else if score >= 70 {
		fmt.Println("成績：C")
	} else {
		fmt.Println("成績：不及格")
	}

	// Go 的 if 不需要小括號 ()，但大括號 {} 是必須的
	// 錯誤：if (score > 90) ← 不需要小括號
	// 錯誤：if score > 90 fmt.Println("A") ← 必須有大括號

	// ========================================
	// 2. if 搭配初始化語句
	// ========================================
	// Go 獨特的語法：可以在 if 中宣告變數
	fmt.Println("\n=== if + 初始化 ===")

	if remainder := 10 % 3; remainder == 0 {
		fmt.Println("10 可以被 3 整除")
	} else {
		fmt.Printf("10 除以 3 的餘數是 %d\n", remainder)
	}
	// 注意：remainder 只在 if/else 區塊內有效

	// ========================================
	// 3. for 迴圈（Go 只有 for，沒有 while）
	// ========================================
	fmt.Println("\n=== for 迴圈 ===")

	// 標準 for（類似 C/Java）
	fmt.Print("標準 for: ")
	for i := 0; i < 5; i++ {
		fmt.Print(i, " ")
	}
	fmt.Println()

	// 類似 while 的 for（只有條件）
	fmt.Print("while 風格: ")
	count := 0
	for count < 5 {
		fmt.Print(count, " ")
		count++
	}
	fmt.Println()

	// 無限迴圈 + break
	fmt.Print("無限迴圈 + break: ")
	n := 0
	for {
		if n >= 5 {
			break // 跳出迴圈
		}
		fmt.Print(n, " ")
		n++
	}
	fmt.Println()

	// continue：跳過當次迭代
	fmt.Print("跳過偶數: ")
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			continue // 跳過，直接進入下一次迭代
		}
		fmt.Print(i, " ")
	}
	fmt.Println()

	// ========================================
	// 4. for range（遍歷）
	// ========================================
	fmt.Println("\n=== for range ===")

	// 遍歷字串
	message := "Hello Go"
	fmt.Println("遍歷字串：")
	for index, char := range message {
		fmt.Printf("  索引 %d: %c\n", index, char)
	}

	// 只需要值，不需要索引：用 _ 忽略
	fmt.Print("只取值: ")
	for _, char := range "ABC" {
		fmt.Printf("%c ", char)
	}
	fmt.Println()

	// 只需要索引：省略第二個變數
	fmt.Print("只取索引: ")
	for i := range "ABC" {
		fmt.Print(i, " ")
	}
	fmt.Println()

	// ========================================
	// 5. switch
	// ========================================
	fmt.Println("\n=== switch ===")

	day := "星期三"

	switch day {
	case "星期一", "星期二", "星期三", "星期四", "星期五":
		fmt.Println(day, "是工作日")
	case "星期六", "星期日":
		fmt.Println(day, "是假日")
	default:
		fmt.Println("未知的日期")
	}
	// Go 的 switch 不需要 break，每個 case 執行完會自動跳出

	// switch 也可以不帶表達式（類似 if-else 鏈）
	temperature := 32

	switch {
	case temperature >= 35:
		fmt.Println("天氣：酷熱")
	case temperature >= 28:
		fmt.Println("天氣：炎熱")
	case temperature >= 20:
		fmt.Println("天氣：舒適")
	default:
		fmt.Println("天氣：寒冷")
	}

	// ========================================
	// 6. 實際範例：判斷 HTTP 狀態碼
	// ========================================
	fmt.Println("\n=== 實際範例 ===")

	statusCode := 404

	switch {
	case statusCode >= 200 && statusCode < 300:
		fmt.Printf("%d: 成功\n", statusCode)
	case statusCode >= 400 && statusCode < 500:
		fmt.Printf("%d: 客戶端錯誤\n", statusCode)
	case statusCode >= 500:
		fmt.Printf("%d: 伺服器錯誤\n", statusCode)
	}
}
