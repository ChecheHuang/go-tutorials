// 第九課：Slice（切片）與 Map（映射）
// 這兩個是 Go 中最常用的集合型別
// 執行方式：go run main.go
package main

import (
	"fmt"
	"sort"
	"strings"
)

func main() {
	// ========================================
	// 1. 陣列（Array）vs 切片（Slice）
	// ========================================
	fmt.Println("=== 陣列 vs 切片 ===")

	// 陣列：長度固定，很少直接使用
	var arr [3]int = [3]int{1, 2, 3}
	fmt.Println("陣列:", arr)

	// 切片：長度可變，日常幾乎都用切片
	slice := []int{1, 2, 3} // 注意：沒有寫長度
	fmt.Println("切片:", slice)

	// ========================================
	// 2. 切片操作
	// ========================================
	fmt.Println("\n=== 切片操作 ===")

	// 建立切片
	fruits := []string{"蘋果", "香蕉", "櫻桃", "葡萄", "芒果"}
	fmt.Println("原始:", fruits)

	// 存取元素
	fmt.Println("第一個:", fruits[0])
	fmt.Println("最後一個:", fruits[len(fruits)-1])

	// 切片操作：[start:end]（包含 start，不包含 end）
	fmt.Println("fruits[1:3]:", fruits[1:3]) // [香蕉, 櫻桃]
	fmt.Println("fruits[:2]:", fruits[:2])    // [蘋果, 香蕉]
	fmt.Println("fruits[3:]:", fruits[3:])    // [葡萄, 芒果]

	// 長度和容量
	fmt.Printf("長度: %d, 容量: %d\n", len(fruits), cap(fruits))

	// ========================================
	// 3. append：新增元素（不可變方式）
	// ========================================
	fmt.Println("\n=== append ===")

	numbers := []int{1, 2, 3}
	fmt.Println("原始:", numbers)

	// append 回傳新的切片，不修改原始切片
	numbers = append(numbers, 4)
	fmt.Println("新增 4:", numbers)

	// 一次新增多個
	numbers = append(numbers, 5, 6, 7)
	fmt.Println("新增 5,6,7:", numbers)

	// 合併兩個切片
	more := []int{8, 9}
	numbers = append(numbers, more...) // ... 展開切片
	fmt.Println("合併:", numbers)

	// ========================================
	// 4. make：預先分配容量
	// ========================================
	fmt.Println("\n=== make ===")

	// make(型別, 長度, 容量)
	// 當你知道大約需要多少元素時，預先分配可以避免多次擴容
	scores := make([]int, 0, 10) // 長度 0，容量 10
	fmt.Printf("scores: len=%d, cap=%d\n", len(scores), cap(scores))

	for i := 0; i < 5; i++ {
		scores = append(scores, i*10)
	}
	fmt.Println("填充後:", scores)

	// ========================================
	// 5. 遍歷切片
	// ========================================
	fmt.Println("\n=== 遍歷切片 ===")

	colors := []string{"紅", "綠", "藍"}

	// for range 取得 index 和 value
	for i, color := range colors {
		fmt.Printf("  [%d] %s\n", i, color)
	}

	// 只要 value
	fmt.Print("只要值: ")
	for _, color := range colors {
		fmt.Print(color, " ")
	}
	fmt.Println()

	// ========================================
	// 6. 函式式操作：filter、map
	// ========================================
	fmt.Println("\n=== filter / map ===")

	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Filter：篩選偶數
	evens := filter(nums, func(n int) bool { return n%2 == 0 })
	fmt.Println("偶數:", evens)

	// Map：每個元素乘以 2
	doubled := mapInts(nums, func(n int) int { return n * 2 })
	fmt.Println("乘以 2:", doubled)

	// ========================================
	// 7. Map（映射 / 字典）
	// ========================================
	fmt.Println("\n=== Map ===")

	// 建立 Map
	ages := map[string]int{
		"Alice": 25,
		"Bob":   30,
		"Carol": 28,
	}
	fmt.Println("Map:", ages)

	// 存取值
	fmt.Println("Alice 的年齡:", ages["Alice"])

	// 新增/修改
	ages["Dave"] = 35
	fmt.Println("新增 Dave:", ages)

	// 刪除
	delete(ages, "Bob")
	fmt.Println("刪除 Bob:", ages)

	// 檢查 key 是否存在
	age, exists := ages["Bob"]
	if exists {
		fmt.Println("Bob 的年齡:", age)
	} else {
		fmt.Println("Bob 不存在")
	}

	// ========================================
	// 8. 遍歷 Map
	// ========================================
	fmt.Println("\n=== 遍歷 Map ===")

	for name, age := range ages {
		fmt.Printf("  %s: %d 歲\n", name, age)
	}
	// 注意：Map 的遍歷順序是隨機的！

	// ========================================
	// 9. Map 的實際應用
	// ========================================
	fmt.Println("\n=== 實際應用：單字計數 ===")

	text := "the quick brown fox jumps over the lazy dog the fox"
	wordCount := countWords(text)

	for word, count := range wordCount {
		if count > 1 {
			fmt.Printf("  '%s' 出現了 %d 次\n", word, count)
		}
	}

	// ========================================
	// 10. 排序
	// ========================================
	fmt.Println("\n=== 排序 ===")

	unsorted := []int{5, 3, 8, 1, 9, 2}
	sort.Ints(unsorted)
	fmt.Println("排序後:", unsorted)

	names := []string{"Charlie", "Alice", "Bob"}
	sort.Strings(names)
	fmt.Println("字串排序:", names)
}

// filter 篩選切片中符合條件的元素
func filter(nums []int, predicate func(int) bool) []int {
	result := make([]int, 0)
	for _, n := range nums {
		if predicate(n) {
			result = append(result, n)
		}
	}
	return result
}

// mapInts 將切片中的每個元素進行轉換
func mapInts(nums []int, transform func(int) int) []int {
	result := make([]int, len(nums))
	for i, n := range nums {
		result[i] = transform(n)
	}
	return result
}

// countWords 計算每個單字出現的次數
func countWords(text string) map[string]int {
	counts := make(map[string]int)
	words := strings.Fields(text) // 用空白分割
	for _, word := range words {
		counts[word]++ // map 的零值是 0，所以可以直接 ++
	}
	return counts
}
