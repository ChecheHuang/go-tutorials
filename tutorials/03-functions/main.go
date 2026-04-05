// 第三課：函式（Functions）
// 執行方式：go run main.go
package main

import (
	"fmt"
	"strings"
)

// ========================================
// 1. 基本函式
// ========================================

// greet 接收一個字串參數，回傳一個字串
func greet(name string) string {
	return "你好，" + name + "！"
}

// add 接收兩個 int，回傳一個 int
func add(a int, b int) int {
	return a + b
}

// 當連續參數型別相同，可以只寫最後一個的型別
func multiply(a, b int) int {
	return a * b
}

// ========================================
// 2. 多重回傳值
// ========================================

// divide 回傳兩個值：商和錯誤訊息
// 這是 Go 最重要的慣例之一：用多重回傳值處理錯誤
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("除數不能為零")
	}
	return a / b, nil // nil 代表「沒有錯誤」
}

// ========================================
// 3. 命名回傳值
// ========================================

// getMinMax 使用命名回傳值
func getMinMax(numbers []int) (min, max int) {
	min = numbers[0]
	max = numbers[0]

	for _, n := range numbers {
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
	}

	return // 直接 return，會回傳 min 和 max 的當前值
}

// ========================================
// 4. 可變參數（Variadic Functions）
// ========================================

// sum 接收任意數量的 int 參數
// numbers 的型別是 []int（int 切片）
func sum(numbers ...int) int {
	total := 0
	for _, n := range numbers {
		total += n
	}
	return total
}

// ========================================
// 5. 函式作為參數（高階函式）
// ========================================

// apply 接收一個切片和一個「轉換函式」，回傳轉換後的新切片
func apply(items []string, transform func(string) string) []string {
	result := make([]string, len(items))
	for i, item := range items {
		result[i] = transform(item)
	}
	return result
}

// ========================================
// 6. 閉包（Closure）
// ========================================

// makeCounter 回傳一個函式，每次呼叫會遞增計數
func makeCounter() func() int {
	count := 0 // 這個變數被閉包「捕獲」
	return func() int {
		count++ // 每次呼叫都會修改外層的 count
		return count
	}
}

// ========================================
// 7. defer：延遲執行
// ========================================

// processFile 示範 defer 的用法
func processFile(filename string) {
	fmt.Printf("  開啟檔案：%s\n", filename)

	// defer 會在函式結束時執行（無論是正常結束還是發生錯誤）
	// 常用於關閉檔案、釋放資源、解鎖
	defer fmt.Printf("  關閉檔案：%s\n", filename)

	fmt.Printf("  處理檔案：%s\n", filename)
	// 函式結束後，defer 的語句才會執行
}

func main() {
	// 1. 基本函式
	fmt.Println("=== 基本函式 ===")
	fmt.Println(greet("小明"))
	fmt.Println("3 + 5 =", add(3, 5))
	fmt.Println("3 × 5 =", multiply(3, 5))

	// 2. 多重回傳值
	fmt.Println("\n=== 多重回傳值 ===")

	result, err := divide(10, 3)
	if err != nil {
		fmt.Println("錯誤:", err)
	} else {
		fmt.Printf("10 ÷ 3 = %.2f\n", result)
	}

	// 除以零的情況
	result, err = divide(10, 0)
	if err != nil {
		fmt.Println("錯誤:", err)
	}

	// 如果確定不需要某個回傳值，用 _ 忽略
	onlyResult, _ := divide(20, 4)
	fmt.Println("20 ÷ 4 =", onlyResult)

	// 3. 命名回傳值
	fmt.Println("\n=== 命名回傳值 ===")
	numbers := []int{3, 1, 4, 1, 5, 9, 2, 6}
	min, max := getMinMax(numbers)
	fmt.Printf("數列 %v 的最小值=%d，最大值=%d\n", numbers, min, max)

	// 4. 可變參數
	fmt.Println("\n=== 可變參數 ===")
	fmt.Println("sum(1, 2, 3) =", sum(1, 2, 3))
	fmt.Println("sum(10, 20, 30, 40, 50) =", sum(10, 20, 30, 40, 50))

	// 把切片展開傳入：加上 ... 後綴
	nums := []int{1, 2, 3, 4, 5}
	fmt.Println("sum(nums...) =", sum(nums...))

	// 5. 函式作為參數
	fmt.Println("\n=== 函式作為參數 ===")
	fruits := []string{"apple", "banana", "cherry"}

	// 傳入 strings.ToUpper 函式
	upperFruits := apply(fruits, strings.ToUpper)
	fmt.Println("大寫:", upperFruits)

	// 傳入匿名函式（lambda）
	decorated := apply(fruits, func(s string) string {
		return "🍎 " + s
	})
	fmt.Println("裝飾:", decorated)

	// 6. 閉包
	fmt.Println("\n=== 閉包 ===")
	counter := makeCounter()
	fmt.Println("計數:", counter()) // 1
	fmt.Println("計數:", counter()) // 2
	fmt.Println("計數:", counter()) // 3

	// 每個閉包有自己的狀態
	counter2 := makeCounter()
	fmt.Println("新計數器:", counter2()) // 1（獨立的）

	// 7. defer
	fmt.Println("\n=== defer ===")
	processFile("data.txt")

	// 多個 defer 按照「後進先出」(LIFO) 順序執行
	fmt.Println("\n多個 defer（LIFO）：")
	for i := 1; i <= 3; i++ {
		defer fmt.Printf("  defer %d\n", i)
	}
	fmt.Println("  函式主體結束")
	// 輸出順序：函式主體結束 → defer 3 → defer 2 → defer 1
}
