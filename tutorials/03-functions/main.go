// ============================================================
// 第三課：函式（Functions）
// ============================================================
// 函式就是「把一段程式碼包裝起來，取個名字，之後可以重複呼叫」
//
// 你會學到：
//   1. 怎麼建立和呼叫函式
//   2. 參數和回傳值
//   3. 多重回傳值（Go 的招牌功能！）
//   4. 可變參數（接收任意數量的參數）
//   5. 函式也是「值」——可以當參數傳來傳去
//   6. 閉包（Closure）——函式記住外面的變數
//   7. defer——「等一下再做」
//
// 執行方式：go run ./tutorials/03-functions
// ============================================================

package main // 可執行程式的套件名稱

import (
	"fmt"     // 印東西到螢幕上
	"strings" // Go 內建的字串處理套件（提供 ToUpper 等函式）
)

// ========================================
// 1. 基本函式
// ========================================

// greet 是一個函式，接收一個 string 參數，回傳一個 string
//
// 語法拆解：
//   func    → 宣告函式的關鍵字
//   greet   → 函式名稱
//   (name string) → 參數列表（參數名稱 型別）
//   string  → 回傳值的型別
func greet(name string) string {
	return "你好，" + name + "！" // return 把值回傳給呼叫者
}

// add 接收兩個 int 參數，回傳一個 int
func add(a int, b int) int {
	return a + b // 把 a + b 的結果回傳
}

// 當「連續的參數型別相同」時，可以只寫最後一個的型別
// multiply(a, b int) 等於 multiply(a int, b int)
func multiply(a, b int) int {
	return a * b
}

// ========================================
// 2. 多重回傳值 —— Go 的招牌功能！
// ========================================
// 大多數語言的函式只能回傳「一個值」
// Go 可以回傳「多個值」，最常見的用法是：回傳（結果, 錯誤）
//
// 這是 Go 處理錯誤的核心模式：
//   func 做某事() (結果型別, error) {
//       if 出錯了 {
//           return 零值, fmt.Errorf("錯誤訊息")
//       }
//       return 結果, nil  // nil = 沒有錯誤
//   }

// divide 計算 a ÷ b，回傳兩個值：(商, 錯誤)
func divide(a, b float64) (float64, error) {
	if b == 0 { // 除數不能是 0
		// fmt.Errorf() 建立一個「錯誤物件」
		return 0, fmt.Errorf("除數不能為零")
	}
	return a / b, nil // nil 代表「沒有錯誤，一切正常」
}

// ========================================
// 3. 命名回傳值（Named Return Values）
// ========================================
// 你可以幫回傳值取名字，它們會自動變成函式內的變數
// 最後用「裸 return」（不寫值）就會回傳這些命名的變數

// getMinMax 找出切片中的最小值和最大值
// (min, max int) 是命名回傳值：min 和 max 會自動宣告為 int 變數
func getMinMax(numbers []int) (min, max int) {
	min = numbers[0] // 先假設第一個是最小的
	max = numbers[0] // 先假設第一個是最大的

	for _, n := range numbers { // 遍歷每個數字
		if n < min { // 如果找到更小的
			min = n // 更新最小值
		}
		if n > max { // 如果找到更大的
			max = n // 更新最大值
		}
	}

	return // 裸 return：自動回傳 min 和 max 的當前值
}

// ========================================
// 4. 可變參數（Variadic Functions）
// ========================================
// 有時候你不知道會收到幾個參數
// 用 ...型別 表示「可以接收 0 個或任意多個」
// 在函式內部，它會變成一個切片（slice）

// sum 可以接收任意數量的 int 參數
// numbers 在函式內部的型別是 []int（int 的切片）
func sum(numbers ...int) int {
	total := 0                    // 初始化總和為 0
	for _, n := range numbers {   // 遍歷所有傳入的數字
		total += n                // total = total + n
	}
	return total                  // 回傳總和
}

// ========================================
// 5. 函式作為參數（Higher-Order Functions）
// ========================================
// 在 Go 中，函式也是一種「值」，可以：
//   - 存到變數裡
//   - 當成參數傳給另一個函式
//   - 從函式中回傳
//
// 「接收函式作為參數」或「回傳函式」的函式，叫做「高階函式」

// apply 接收一個字串切片和一個「轉換函式」
// 它會對切片中的每個元素套用轉換函式，回傳新的切片
//
// transform 參數的型別是 func(string) string
// 意思是：「一個接收 string、回傳 string 的函式」
func apply(items []string, transform func(string) string) []string {
	result := make([]string, len(items)) // 建立一個跟原切片一樣長的新切片
	for i, item := range items {         // 遍歷每個元素
		result[i] = transform(item)      // 對每個元素套用轉換函式
	}
	return result // 回傳轉換後的新切片
}

// ========================================
// 6. 函式型別（Function Types）
// ========================================
// 你可以用 type 幫函式簽名取一個名字
// 這在部落格專案的中介層（middleware）和路由處理中非常常見

// HandlerFunc 是一個函式型別：接收 string，不回傳值
// 這跟 Gin 框架的 gin.HandlerFunc 概念一樣
type HandlerFunc func(request string)

// useHandler 接收一個 HandlerFunc 型別的參數
func useHandler(path string, handler HandlerFunc) {
	fmt.Printf("  收到請求：%s\n", path) // 印出路徑
	handler(path)                        // 呼叫傳入的處理函式
}

// ========================================
// 7. 閉包（Closure）
// ========================================
// 閉包 = 函式 + 它「記住」的外部變數
// 就像一個人帶著自己的背包（外部變數），走到哪都帶著

// makeCounter 回傳一個函式
// 回傳型別 func() int 表示「一個不接收參數、回傳 int 的函式」
func makeCounter() func() int {
	count := 0 // 這個變數會被回傳的函式「記住」（捕獲）

	// 回傳一個匿名函式（沒有名字的函式）
	return func() int {
		count++ // 每次呼叫都會修改 count（它記住了外層的 count）
		return count
	}
	// count 不會消失！因為回傳的函式還在用它
}

// ========================================
// 8. defer —— 「等一下再做」
// ========================================
// defer 會把後面的函式呼叫「排隊」，等到目前函式結束時才執行
// 最常用來：關閉檔案、關閉資料庫連線、解除鎖定
// 好處：開啟資源後「馬上」寫 defer 關閉，就不怕忘記

func processFile(filename string) {
	fmt.Printf("  1. 開啟檔案：%s\n", filename)

	// defer 不是「現在」執行，而是「函式結束時」才執行
	// 就像你進教室後在門上貼便利貼：「離開時記得關燈」
	defer fmt.Printf("  3. 關閉檔案：%s（defer 延遲執行）\n", filename)

	fmt.Printf("  2. 處理檔案：%s\n", filename)
	// 函式結束 → 執行 defer 的內容
}

// ========================================
// main 函式——程式入口
// ========================================
func main() {

	// --- 1. 呼叫基本函式 ---
	fmt.Println("=== 1. 基本函式 ===")
	fmt.Println(greet("小明"))          // 呼叫 greet，印出：你好，小明！
	fmt.Println("3 + 5 =", add(3, 5))  // 呼叫 add，印出：3 + 5 = 8
	fmt.Println("3 × 5 =", multiply(3, 5)) // 呼叫 multiply，印出：3 × 5 = 15

	// --- 2. 多重回傳值 ---
	fmt.Println("\n=== 2. 多重回傳值（Go 的錯誤處理模式）===")

	// 接收兩個回傳值：result 和 err
	result, err := divide(10, 3)
	if err != nil { // 先檢查有沒有錯誤
		fmt.Println("錯誤:", err)
	} else {
		fmt.Printf("10 ÷ 3 = %.2f\n", result) // 沒有錯誤才使用結果
	}

	// 故意除以零，觸發錯誤
	result, err = divide(10, 0) // 注意：第二次賦值用 = 不用 :=
	if err != nil {
		fmt.Println("錯誤:", err) // 印出：錯誤: 除數不能為零
	}

	// 如果你「確定」不需要 error（不建議，但有時會這樣做）
	// 用 _ 忽略它
	onlyResult, _ := divide(20, 4) // 忽略 error
	fmt.Println("20 ÷ 4 =", onlyResult)

	// --- 3. 命名回傳值 ---
	fmt.Println("\n=== 3. 命名回傳值 ===")
	numbers := []int{3, 1, 4, 1, 5, 9, 2, 6} // 一個數字切片
	min, max := getMinMax(numbers)             // 接收兩個回傳值
	fmt.Printf("數列 %v 的最小值=%d，最大值=%d\n", numbers, min, max)

	// --- 4. 可變參數 ---
	fmt.Println("\n=== 4. 可變參數 ===")
	fmt.Println("sum(1, 2, 3) =", sum(1, 2, 3))             // 傳 3 個參數
	fmt.Println("sum(10, 20, 30, 40, 50) =", sum(10, 20, 30, 40, 50)) // 傳 5 個

	// 如果你已經有一個切片，要展開傳入，加上 ... 後綴
	nums := []int{1, 2, 3, 4, 5}
	fmt.Println("sum(nums...) =", sum(nums...)) // 把切片展開成 sum(1,2,3,4,5)

	// --- 5. 函式作為參數 ---
	fmt.Println("\n=== 5. 函式作為參數 ===")
	fruits := []string{"apple", "banana", "cherry"}

	// 傳入 strings.ToUpper（Go 內建的「轉大寫」函式）
	// strings.ToUpper 的簽名是 func(string) string，剛好匹配
	upperFruits := apply(fruits, strings.ToUpper)
	fmt.Println("轉大寫:", upperFruits) // [APPLE BANANA CHERRY]

	// 傳入「匿名函式」（沒有名字的函式，也叫 lambda）
	// 就地定義一個函式，直接傳進去
	decorated := apply(fruits, func(s string) string {
		return "[ " + s + " ]" // 在每個水果名稱前後加上方括號
	})
	fmt.Println("裝飾:", decorated) // [[ apple ] [ banana ] [ cherry ]]

	// --- 6. 函式型別 ---
	fmt.Println("\n=== 6. 函式型別 ===")
	// 定義一個 HandlerFunc 並傳給 useHandler
	useHandler("/api/v1/articles", func(request string) {
		fmt.Printf("  處理文章請求：%s\n", request)
	})
	// 這跟部落格專案中 Gin 的路由處理非常類似：
	//   router.GET("/articles", articleHandler.GetAll)
	//   ↑ GET 接收的就是一個 HandlerFunc 型別的參數

	// --- 7. 閉包 ---
	fmt.Println("\n=== 7. 閉包（函式 + 記憶）===")
	counter := makeCounter()         // 取得一個計數器函式
	fmt.Println("計數:", counter())   // 1（內部 count 變成 1）
	fmt.Println("計數:", counter())   // 2（內部 count 變成 2）
	fmt.Println("計數:", counter())   // 3（內部 count 變成 3）

	// 每個閉包有「自己的」count，互不影響
	counter2 := makeCounter()          // 新的計數器
	fmt.Println("新計數器:", counter2()) // 1（獨立的 count，從 0 開始）

	// --- 8. defer ---
	fmt.Println("\n=== 8. defer（延遲執行）===")
	processFile("data.txt")
	// 輸出順序：開啟 → 處理 → 關閉（defer 最後才執行）

	// 多個 defer 的執行順序：後進先出（LIFO，像疊盤子）
	// 最後 defer 的最先執行
	fmt.Println("\n多個 defer 的順序（後進先出）：")
	fmt.Println("  開始 defer 排隊⋯")
	defer fmt.Println("  defer 第 1 個排的（最後執行）")
	defer fmt.Println("  defer 第 2 個排的")
	defer fmt.Println("  defer 第 3 個排的（最先執行）")
	fmt.Println("  main 函式主體結束")
	// 實際輸出順序：
	//   main 函式主體結束
	//   defer 第 3 個排的（最先執行）
	//   defer 第 2 個排的
	//   defer 第 1 個排的（最後執行）
}
