// ==========================================================================
// 第二十二課：Benchmark + Load Testing
// ==========================================================================
//
// 這一課教你「怎麼證明你的程式碼夠快」：
//   1. Go Benchmark（go test -bench）— 測量函式效能
//   2. 壓力測試概念 — 模擬多使用者同時請求
//   3. 效能比較 — 用 benchmark 比較不同實作方式
//
// 執行方式：
//   go run ./tutorials/22-benchmark/
//   go test -bench=. -benchmem ./tutorials/22-benchmark/
// ==========================================================================

package main

import (
	"fmt"
	"strings"
	"time"
)

// ==========================================================================
// 1. 字串串接效能比較
// ==========================================================================

// ConcatPlus 用 + 串接（每次都建立新字串，O(n²)）
func ConcatPlus(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "a"
	}
	return s
}

// ConcatBuilder 用 strings.Builder（預分配緩衝區，O(n)）
func ConcatBuilder(n int) string {
	var b strings.Builder
	b.Grow(n)
	for i := 0; i < n; i++ {
		b.WriteString("a")
	}
	return b.String()
}

// ==========================================================================
// 2. Slice 預分配 vs 動態增長
// ==========================================================================

// SliceAppend 動態增長（多次重新分配記憶體）
func SliceAppend(n int) []int {
	var s []int
	for i := 0; i < n; i++ {
		s = append(s, i)
	}
	return s
}

// SlicePrealloc 預分配容量
func SlicePrealloc(n int) []int {
	s := make([]int, 0, n)
	for i := 0; i < n; i++ {
		s = append(s, i)
	}
	return s
}

// ==========================================================================
// 3. Map 查找 vs Slice 線性搜尋
// ==========================================================================

func SearchSlice(data []string, target string) bool {
	for _, v := range data {
		if v == target {
			return true
		}
	}
	return false
}

func SearchMap(data map[string]bool, target string) bool {
	return data[target]
}

// ==========================================================================
// 4. 簡易壓力測試模擬
// ==========================================================================

func simulateLoadTest(name string, fn func(), concurrency, requestsPerWorker int) {
	start := time.Now()
	done := make(chan struct{}, concurrency*requestsPerWorker)

	for w := 0; w < concurrency; w++ {
		go func() {
			for r := 0; r < requestsPerWorker; r++ {
				fn()
				done <- struct{}{}
			}
		}()
	}

	total := concurrency * requestsPerWorker
	for i := 0; i < total; i++ {
		<-done
	}

	elapsed := time.Since(start)
	rps := float64(total) / elapsed.Seconds()

	fmt.Printf("  %-25s │ %5d 請求 │ %3d 並發 │ %8.0f req/s │ %v\n",
		name, total, concurrency, rps, elapsed.Round(time.Millisecond))
}

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║      第 22 課：Benchmark + Load Testing                 ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")

	// === Demo 1: 字串串接 ===
	fmt.Println("\n📊 Demo 1: 字串串接效能比較")
	fmt.Println("─────────────────────────────────────────────")

	n := 100000
	fmt.Printf("  串接 %d 個字元：\n\n", n)

	start := time.Now()
	ConcatPlus(n)
	plusTime := time.Since(start)

	start = time.Now()
	ConcatBuilder(n)
	builderTime := time.Since(start)

	fmt.Printf("  %-20s │ %v\n", "string + （❌ 慢）", plusTime)
	fmt.Printf("  %-20s │ %v\n", "strings.Builder ✅", builderTime)
	if plusTime > 0 {
		fmt.Printf("  Builder 快 %.0fx\n", float64(plusTime)/float64(builderTime))
	}

	// === Demo 2: Slice 預分配 ===
	fmt.Println("\n📊 Demo 2: Slice 預分配 vs 動態增長")
	fmt.Println("─────────────────────────────────────────────")

	n = 1000000
	fmt.Printf("  Append %d 個元素：\n\n", n)

	start = time.Now()
	SliceAppend(n)
	appendTime := time.Since(start)

	start = time.Now()
	SlicePrealloc(n)
	preallocTime := time.Since(start)

	fmt.Printf("  %-20s │ %v\n", "動態 append（❌）", appendTime)
	fmt.Printf("  %-20s │ %v\n", "預分配 make ✅", preallocTime)

	// === Demo 3: Map vs Slice 查找 ===
	fmt.Println("\n📊 Demo 3: Map 查找 vs Slice 線性搜尋")
	fmt.Println("─────────────────────────────────────────────")

	size := 10000
	sliceData := make([]string, size)
	mapData := make(map[string]bool, size)
	for i := 0; i < size; i++ {
		key := fmt.Sprintf("key-%d", i)
		sliceData[i] = key
		mapData[key] = true
	}

	target := fmt.Sprintf("key-%d", size-1) // 最後一個元素（最壞情況）
	iterations := 10000

	start = time.Now()
	for i := 0; i < iterations; i++ {
		SearchSlice(sliceData, target)
	}
	sliceTime := time.Since(start)

	start = time.Now()
	for i := 0; i < iterations; i++ {
		SearchMap(mapData, target)
	}
	mapTime := time.Since(start)

	fmt.Printf("  在 %d 筆資料中搜尋 %d 次：\n\n", size, iterations)
	fmt.Printf("  %-20s │ %v\n", "Slice 線性搜尋 ❌", sliceTime)
	fmt.Printf("  %-20s │ %v\n", "Map O(1) 查找 ✅", mapTime)

	// === Demo 4: 模擬壓力測試 ===
	fmt.Println("\n🔨 Demo 4: 簡易壓力測試")
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println()
	fmt.Printf("  %-25s │ %5s │ %5s │ %12s │ %s\n",
		"測試項目", "總請求", "並發", "吞吐量", "耗時")
	fmt.Println("  ─────────────────────────┼───────┼───────┼──────────────┼────────")

	simulateLoadTest("輕量計算", func() {
		_ = 1 + 1
	}, 10, 10000)

	simulateLoadTest("字串處理", func() {
		ConcatBuilder(100)
	}, 10, 10000)

	simulateLoadTest("Map 查找", func() {
		SearchMap(mapData, "key-5000")
	}, 10, 10000)

	// === 說明 go test -bench ===
	fmt.Println("\n📝 如何用 go test -bench 做正式 Benchmark")
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("  在 _test.go 檔案中寫 Benchmark 函式：")
	fmt.Println()
	fmt.Println("    func BenchmarkConcatPlus(b *testing.B) {")
	fmt.Println("        for i := 0; i < b.N; i++ {")
	fmt.Println("            ConcatPlus(1000)")
	fmt.Println("        }")
	fmt.Println("    }")
	fmt.Println()
	fmt.Println("  執行：go test -bench=. -benchmem ./tutorials/22-benchmark/")
	fmt.Println()
	fmt.Println("  輸出範例：")
	fmt.Println("    BenchmarkConcatPlus-8      1000    1500000 ns/op   5000000 B/op   1000 allocs/op")
	fmt.Println("    BenchmarkConcatBuilder-8  50000      30000 ns/op     65536 B/op      1 allocs/op")
	fmt.Println("                                │          │                │              │")
	fmt.Println("                          執行次數    每次耗時        每次分配記憶體    分配次數")
}
