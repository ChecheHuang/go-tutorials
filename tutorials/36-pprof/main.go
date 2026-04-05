// ==========================================================================
// 第三十六課：pprof 效能分析（Performance Profiling）
// ==========================================================================
//
// 什麼是 pprof？
//   Go 內建的效能分析工具，可以分析：
//   - CPU 使用率：哪個函式最耗 CPU
//   - 記憶體分配：哪裡在不停分配記憶體（可能導致 GC 壓力）
//   - Goroutine：有多少 goroutine、它們在做什麼
//   - 阻塞分析：goroutine 在哪裡等待
//
// 使用流程：
//   1. 在程式裡加入 _ "net/http/pprof"（自動掛載 /debug/pprof）
//   2. 執行程式
//   3. go tool pprof http://localhost:6060/debug/pprof/profile
//   4. 在 pprof 工具裡分析（top、list、web 指令）
//
// 執行方式：go run ./tutorials/32-pprof
// 然後在另一個終端：go tool pprof http://localhost:6060/debug/pprof/heap
// ==========================================================================

package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof" // 這一行自動把 pprof handler 掛載到預設 ServeMux
	"runtime"
	"strings"
	"sync"
	"time"
)

// ==========================================================================
// 1. 模擬效能問題（用來示範 pprof 能找到什麼）
// ==========================================================================

// inefficientStringConcat 低效的字串拼接（會大量記憶體分配）
// 每次 += 都建立新字串，O(n²) 記憶體分配
func inefficientStringConcat(n int) string {
	result := ""
	for i := range n {
		result += fmt.Sprintf("item-%d,", i) // ❌ 低效：每次都重新分配
	}
	return result
}

// efficientStringConcat 高效的字串拼接
func efficientStringConcat(n int) string {
	var sb strings.Builder
	sb.Grow(n * 10) // 預先分配足夠空間
	for i := range n {
		sb.WriteString(fmt.Sprintf("item-%d,", i)) // ✅ 高效：在既有 buffer 上附加
	}
	return sb.String()
}

// cpuIntensiveWork 模擬 CPU 密集工作
func cpuIntensiveWork(n int) float64 {
	result := 0.0
	for i := range n {
		result += math.Sqrt(float64(i)) * math.Sin(float64(i))
	}
	return result
}

// memoryLeak 模擬記憶體洩漏（只增不減的 slice）
type memoryLeakSimulator struct {
	mu   sync.Mutex
	data [][]byte
}

func (m *memoryLeakSimulator) add(size int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// 模擬忘記釋放記憶體
	m.data = append(m.data, make([]byte, size))
}

func (m *memoryLeakSimulator) size() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	total := 0
	for _, d := range m.data {
		total += len(d)
	}
	return total
}

// goroutineLeak 模擬 goroutine 洩漏
func goroutineLeak() {
	ch := make(chan int) // 無緩衝 channel，沒有人讀取
	go func() {
		<-ch // goroutine 永遠阻塞在這裡！（洩漏）
	}()
	// ch 和 goroutine 都無法被 GC，因為 goroutine 還在等待
}

// ==========================================================================
// 2. 效能基準測試（與 pprof 配合使用）
// ==========================================================================
//
// 雖然這裡示範 HTTP pprof，但 pprof 最常見的用法是搭配 Benchmark：
//
//   // 在 *_test.go 中：
//   func BenchmarkStringConcat(b *testing.B) {
//       for range b.N {
//           inefficientStringConcat(1000)
//       }
//   }
//
//   // 執行並生成 CPU profile：
//   go test -bench=. -cpuprofile=cpu.prof ./tutorials/32-pprof/
//   go tool pprof cpu.prof
//
//   // 生成 Memory profile：
//   go test -bench=. -memprofile=mem.prof ./tutorials/32-pprof/
//   go tool pprof mem.prof

// ==========================================================================
// 3. 記憶體統計
// ==========================================================================

// printMemStats 印出當前記憶體使用狀況
func printMemStats(label string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("\n[%s] 記憶體統計：\n", label)
	fmt.Printf("  Alloc（當前使用）:   %7.2f KB\n", float64(m.Alloc)/1024)
	fmt.Printf("  TotalAlloc（累計）:  %7.2f KB\n", float64(m.TotalAlloc)/1024)
	fmt.Printf("  HeapObjects（物件數）: %d\n", m.HeapObjects)
	fmt.Printf("  NumGC（GC 次數）:     %d\n", m.NumGC)
	fmt.Printf("  Goroutines:           %d\n", runtime.NumGoroutine())
}

// ==========================================================================
// 4. HTTP Handlers（示範 pprof 端點）
// ==========================================================================

// benchmarkHandler 讓外部可以觸發工作負載
func benchmarkHandler(w http.ResponseWriter, r *http.Request) {
	taskType := r.URL.Query().Get("type")

	switch taskType {
	case "cpu":
		// 觸發 CPU 密集工作
		result := cpuIntensiveWork(1_000_000)
		fmt.Fprintf(w, "CPU 工作完成，結果: %.4f\n", result)

	case "memory-bad":
		// 示範低效記憶體使用
		_ = inefficientStringConcat(5000)
		fmt.Fprintf(w, "低效字串拼接完成（注意 /debug/pprof/heap）\n")

	case "memory-good":
		// 示範高效記憶體使用
		_ = efficientStringConcat(5000)
		fmt.Fprintf(w, "高效字串拼接完成\n")

	default:
		fmt.Fprintf(w, `可用的 type 參數：
  ?type=cpu          觸發 CPU 密集工作
  ?type=memory-bad   低效字串拼接（大量記憶體分配）
  ?type=memory-good  高效字串拼接（減少分配）
`)
	}
}

// ==========================================================================
// 主程式
// ==========================================================================

func main() {
	fmt.Println("==========================================")
	fmt.Println(" 第三十二課：pprof 效能分析")
	fmt.Println("==========================================")

	// ──── 1. 記憶體分配對比 ────
	fmt.Println("\n=== 1. 字串拼接效能對比 ===")
	printMemStats("開始前")

	start := time.Now()
	_ = inefficientStringConcat(10000)
	fmt.Printf("\n低效拼接耗時: %v\n", time.Since(start))
	printMemStats("低效拼接後")

	runtime.GC() // 強制 GC，清理上一個測試的垃圾

	start = time.Now()
	_ = efficientStringConcat(10000)
	fmt.Printf("\n高效拼接耗時: %v\n", time.Since(start))
	printMemStats("高效拼接後")

	// ──── 2. Goroutine 洩漏示範 ────
	fmt.Println("\n=== 2. Goroutine 洩漏示範 ===")
	fmt.Printf("洩漏前 goroutine 數: %d\n", runtime.NumGoroutine())

	for range 5 {
		goroutineLeak() // 建立 5 個洩漏的 goroutine
	}

	time.Sleep(10 * time.Millisecond) // 讓 goroutine 啟動
	fmt.Printf("洩漏後 goroutine 數: %d（增加了 5 個！）\n", runtime.NumGoroutine())
	fmt.Println("→ 用 /debug/pprof/goroutine 可以看到它們在哪裡阻塞")

	// ──── 3. pprof 端點說明 ────
	fmt.Println("\n=== 3. pprof 端點（import _ \"net/http/pprof\" 自動掛載）===")
	fmt.Println()
	fmt.Println("  /debug/pprof/           → 首頁，列出所有 profile 類型")
	fmt.Println("  /debug/pprof/profile    → CPU profile（預設 30 秒採樣）")
	fmt.Println("  /debug/pprof/heap       → 記憶體 heap 快照")
	fmt.Println("  /debug/pprof/goroutine  → 所有 goroutine 的堆疊追蹤")
	fmt.Println("  /debug/pprof/block      → goroutine 阻塞分析")
	fmt.Println("  /debug/pprof/mutex      → 互斥鎖競爭分析")
	fmt.Println("  /debug/pprof/allocs     → 記憶體分配採樣")
	fmt.Println()

	// ──── 4. 使用說明 ────
	fmt.Println("=== 4. 如何使用 go tool pprof ===")
	fmt.Println()
	fmt.Println("# 取得 CPU profile（在伺服器執行時，開另一個終端）：")
	fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/profile?seconds=10")
	fmt.Println()
	fmt.Println("# 取得記憶體 heap profile：")
	fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/heap")
	fmt.Println()
	fmt.Println("# pprof 互動指令：")
	fmt.Println("  top          → 列出最耗資源的函式")
	fmt.Println("  top -cum     → 按累計時間排序")
	fmt.Println("  list <func>  → 顯示特定函式的原始碼（帶效能數據）")
	fmt.Println("  web          → 用瀏覽器開啟火焰圖（需要 graphviz）")
	fmt.Println()
	fmt.Println("# 搭配 Benchmark 使用（最常見）：")
	fmt.Println("  go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof .")
	fmt.Println("  go tool pprof cpu.prof")
	fmt.Println("  go tool pprof -http=:8081 mem.prof  ← 用瀏覽器看火焰圖")
	fmt.Println()

	// ──── 5. 啟動 HTTP 伺服器（含 pprof 端點）────
	fmt.Println("=== 啟動 pprof 伺服器 ===")
	fmt.Println()
	fmt.Println("pprof 端點: http://localhost:6060/debug/pprof/")
	fmt.Println("業務 API:   http://localhost:6060/benchmark?type=cpu")
	fmt.Println()
	fmt.Println("試試看（開另一個終端）：")
	fmt.Println("  curl http://localhost:6060/benchmark?type=cpu")
	fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/heap")
	fmt.Println()

	// 加入業務 handler
	http.HandleFunc("/benchmark", benchmarkHandler)

	// 啟動在 :6060（不佔用 :8080，與其他課程共存）
	// 注意：_ "net/http/pprof" 已自動把 pprof handler 掛載到 http.DefaultServeMux
	log.Fatal(http.ListenAndServe(":6060", nil))
}
