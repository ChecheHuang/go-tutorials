// ==========================================================================
// 第十九課：Goroutine 與並發（Goroutines & Concurrency）
// ==========================================================================
//
// 什麼是並發（Concurrency）？
//   並發 = 同時處理多件事情
//   就像一個人可以邊煮飯、邊洗碗、邊聽音樂——不是「同時」做，
//   而是「交替」處理多件事，讓整體效率更高
//
// Go 語言的並發三寶：
//   1. Goroutine — 超輕量的「執行緒」，用 go 關鍵字啟動
//   2. Channel   — Goroutine 之間傳遞資料的「管道」
//   3. Select    — 同時監聽多個 Channel 的「交通指揮」
//
// 還有：
//   4. sync.WaitGroup — 等待一組 Goroutine 全部完成
//   5. sync.Mutex     — 保護共享資料，防止同時寫入造成混亂
//   6. context.Context — 傳遞取消訊號和超時設定
//
// 執行方式：go run ./tutorials/19-goroutines
// ==========================================================================

package main // 宣告這是 main 套件

import ( // 匯入所有需要的套件
	"context"  // 標準庫：傳遞取消訊號、截止時間
	"fmt"      // 標準庫：格式化輸出
	"sync"     // 標準庫：同步原語（WaitGroup、Mutex、Once）
	"time"     // 標準庫：時間相關功能
)

// ==========================================================================
// 1. Goroutine 基礎
// ==========================================================================
//
// Goroutine 就像餐廳的廚師：
//   一個廚師同時能應付多張桌子的訂單
//   Go runtime 會自動調度，讓多個 goroutine 有效利用 CPU
//
// Goroutine 的特點：
//   - 啟動成本極低（只需約 2KB 記憶體，OS Thread 需要 1MB）
//   - 一個程式可以輕鬆跑幾萬個 goroutine
//   - 由 Go runtime 管理（不是 OS），排程更高效

// cookDish 模擬煮一道菜（耗時操作）
func cookDish(dish string, duration time.Duration) { // 接受菜名和烹飪時間
	fmt.Printf("開始煮：%s\n", dish)          // 宣告開始
	time.Sleep(duration)                      // 模擬烹飪耗時（實際上可能是 DB 查詢、API 呼叫）
	fmt.Printf("完成：%s（耗時 %v）\n", dish, duration) // 宣告完成
}

// demonstrateGoroutine 示範 goroutine 的基本用法
func demonstrateGoroutine() { // 函式：示範 goroutine
	fmt.Println("\n=== 1. Goroutine 基礎 ===") // 印出標題

	// 沒有 goroutine：循序執行，一個做完才做下一個
	fmt.Println("\n--- 循序執行（沒有 goroutine）---") // 小標題
	start := time.Now()                              // 記錄開始時間
	cookDish("牛排", 100*time.Millisecond)            // 先煮牛排（100ms）
	cookDish("義大利麵", 80*time.Millisecond)          // 再煮義大利麵（80ms）
	cookDish("沙拉", 50*time.Millisecond)              // 最後做沙拉（50ms）
	fmt.Printf("循序執行總耗時：%v\n", time.Since(start)) // 總耗時應該是 230ms

	// 有 goroutine：並發執行，三道菜同時進行
	fmt.Println("\n--- 並發執行（使用 goroutine）---") // 小標題

	var wg sync.WaitGroup // WaitGroup：用來等待所有 goroutine 完成（見第 2 節）

	start = time.Now() // 重新記錄開始時間

	wg.Add(3)                        // 告訴 WaitGroup：我要等 3 個 goroutine
	go func() {                      // go 關鍵字：在新的 goroutine 中執行
		defer wg.Done()              // defer：函式結束時通知 WaitGroup「我完成了」
		cookDish("牛排", 100*time.Millisecond) // 非同步執行
	}()
	go func() {                      // 第二個 goroutine
		defer wg.Done()              // 完成時通知
		cookDish("義大利麵", 80*time.Millisecond) // 非同步執行
	}()
	go func() {                      // 第三個 goroutine
		defer wg.Done()              // 完成時通知
		cookDish("沙拉", 50*time.Millisecond) // 非同步執行
	}()

	wg.Wait() // 等待所有 3 個 goroutine 都完成（阻塞在這裡）
	fmt.Printf("並發執行總耗時：%v（最慢的那個）\n", time.Since(start)) // 總耗時應該約 100ms
}

// ==========================================================================
// 2. sync.WaitGroup — 等待一組 Goroutine 全部完成
// ==========================================================================
//
// WaitGroup 就像「點名簿」：
//   Add(n)  → 登記要等 n 個人
//   Done()  → 一個人完成了（計數 -1）
//   Wait()  → 等所有人都完成（計數歸零）才繼續

// demonstrateWaitGroup 示範 WaitGroup 的用法
func demonstrateWaitGroup() { // 函式：示範 WaitGroup
	fmt.Println("\n=== 2. WaitGroup ===") // 印出標題

	var wg sync.WaitGroup // 宣告 WaitGroup（零值就可以用，不需要初始化）

	// 用迴圈啟動多個 worker goroutine
	for i := 1; i <= 5; i++ { // 啟動 5 個 worker
		wg.Add(1)  // 每啟動一個 goroutine，計數加 1
		go func(workerID int) { // goroutine 函式，接受 worker ID
			// 注意：不能在 goroutine 裡用外部的 i 變數！
			// 因為迴圈跑完後 i 已經是 6，所有 goroutine 看到的都是 6
			// 正確做法：把 i 當參數傳進去（workerID）
			defer wg.Done() // 函式結束時計數 -1（用 defer 保證一定執行）
			time.Sleep(time.Duration(workerID) * 20 * time.Millisecond) // 模擬不同的工作時間
			fmt.Printf("  Worker %d 完成任務\n", workerID) // 印出完成訊息
		}(i) // 立即呼叫，把 i 的當前值傳進去（避免 closure 陷阱）
	}

	wg.Wait() // 阻塞等待，直到所有 5 個 worker 都呼叫了 Done()
	fmt.Println("所有 worker 都完成了！") // 所有人完成後才會到這裡
}

// ==========================================================================
// 3. Channel — Goroutine 之間的通訊管道
// ==========================================================================
//
// Channel 就像工廠的「傳送帶」：
//   一端放東西（發送），另一端取東西（接收）
//   設計哲學："不要透過共享記憶體來溝通，要透過溝通來共享記憶體"
//
// 兩種 Channel：
//   - 無緩衝（make(chan T)）：發送方等接收方準備好才能送（同步）
//   - 有緩衝（make(chan T, n)）：最多放 n 個，滿了才阻塞（非同步）

// demonstrateChannel 示範 Channel 的基本用法
func demonstrateChannel() { // 函式：示範 Channel
	fmt.Println("\n=== 3. Channel ===") // 印出標題

	// ---- 無緩衝 Channel ----
	fmt.Println("\n--- 無緩衝 Channel（同步）---") // 小標題

	ch := make(chan string) // 建立一個無緩衝的字串 channel

	// 在 goroutine 中發送訊息
	go func() { // 啟動 goroutine
		time.Sleep(50 * time.Millisecond) // 等一下（模擬處理時間）
		ch <- "你好，我是 goroutine！"     // <- 是發送操作：把訊息放入 channel
		// 注意：這裡會「阻塞」，直到有人接收才繼續
	}()

	msg := <-ch // <- 是接收操作：從 channel 取出訊息（阻塞直到有資料）
	fmt.Println("收到:", msg) // 印出收到的訊息

	// ---- 有緩衝 Channel ----
	fmt.Println("\n--- 有緩衝 Channel（非同步）---") // 小標題

	bufferedCh := make(chan int, 3) // 建立容量為 3 的有緩衝 channel

	// 可以連續放入 3 個，不需要接收方就緒
	bufferedCh <- 1  // 放入第 1 個（不阻塞，緩衝還有空間）
	bufferedCh <- 2  // 放入第 2 個（不阻塞）
	bufferedCh <- 3  // 放入第 3 個（不阻塞）
	// bufferedCh <- 4  // 如果放第 4 個就會阻塞，因為緩衝滿了

	fmt.Println("從 bufferedCh 取出:", <-bufferedCh) // 取出第 1 個：1
	fmt.Println("從 bufferedCh 取出:", <-bufferedCh) // 取出第 2 個：2
	fmt.Println("從 bufferedCh 取出:", <-bufferedCh) // 取出第 3 個：3

	// ---- 用 Channel 傳遞計算結果 ----
	fmt.Println("\n--- 用 Channel 收集計算結果 ---") // 小標題

	resultCh := make(chan int, 3) // 有緩衝 channel 來收集結果

	// 啟動 3 個 goroutine 並行計算
	for i := 1; i <= 3; i++ { // 啟動 3 個計算任務
		go func(n int) {              // goroutine 接受數字
			result := n * n           // 計算平方
			resultCh <- result        // 把結果送入 channel
		}(i) // 傳入當前的 i 值
	}

	// 收集 3 個結果
	for i := 0; i < 3; i++ { // 接收 3 次
		fmt.Printf("  計算結果: %d\n", <-resultCh) // 接收並印出
	}
}

// ==========================================================================
// 4. Select — 同時監聽多個 Channel
// ==========================================================================
//
// select 就像餐廳的「服務鈴」：
//   哪個桌子先按鈴，服務員就先去哪桌
//   如果同時多桌按鈴，隨機選一桌
//   加上 default 就變成「非阻塞」：沒有鈴聲就做別的事

// demonstrateSelect 示範 select 語句
func demonstrateSelect() { // 函式：示範 select
	fmt.Println("\n=== 4. Select ===") // 印出標題

	ch1 := make(chan string, 1) // 建立 channel 1
	ch2 := make(chan string, 1) // 建立 channel 2

	// 模擬兩個非同步操作，哪個快就先處理哪個
	go func() { // 第一個操作
		time.Sleep(30 * time.Millisecond) // 慢一點
		ch1 <- "來自 ch1 的訊息"          // 送出結果
	}()
	go func() { // 第二個操作
		time.Sleep(10 * time.Millisecond) // 快一點
		ch2 <- "來自 ch2 的訊息"          // 送出結果
	}()

	// 等待一段時間讓兩個 goroutine 都完成
	time.Sleep(50 * time.Millisecond)

	// select：哪個 channel 有資料就處理哪個（都有的話隨機選一個）
	for i := 0; i < 2; i++ { // 接收兩次（因為兩個 channel 都有資料）
		select { // select 語法，類似 switch 但用於 channel
		case msg1 := <-ch1: // 如果 ch1 有資料
			fmt.Println("收到 ch1:", msg1) // 處理 ch1 的訊息
		case msg2 := <-ch2: // 如果 ch2 有資料
			fmt.Println("收到 ch2:", msg2) // 處理 ch2 的訊息
		}
	}

	// ---- 帶 default 的非阻塞 select ----
	fmt.Println("\n--- 非阻塞 select（帶 default）---") // 小標題

	ch3 := make(chan string) // 建立一個空的 channel（沒有資料）

	select { // select 嘗試從 ch3 接收
	case msg := <-ch3: // 如果有資料
		fmt.Println("收到:", msg) // 處理訊息
	default: // 如果沒有任何 case 就緒
		fmt.Println("目前沒有資料，繼續做其他事") // 不阻塞，直接走 default
	}
}

// ==========================================================================
// 5. sync.Mutex — 保護共享資料
// ==========================================================================
//
// Mutex（互斥鎖）就像公用廚房的刀架：
//   一次只有一個人可以拿到刀（Lock）
//   用完之後要放回去（Unlock）
//   其他人要用就得等待
//
// 什麼時候需要 Mutex？
//   多個 goroutine 同時讀寫同一個變數時
//   如果不加鎖，會發生「競態條件」（Race Condition）—— 結果不可預測！

// SafeCounter 是一個執行緒安全的計數器
type SafeCounter struct { // 定義結構體
	mu    sync.Mutex // 互斥鎖（保護 count 不被同時修改）
	count int        // 計數值（共享狀態，需要保護）
}

// Increment 安全地讓計數器加 1
func (c *SafeCounter) Increment() { // 指標接收者（需要修改 count）
	c.mu.Lock()   // 加鎖：我要開始修改，其他人等一下
	defer c.mu.Unlock() // defer 確保函式結束時一定解鎖（就算 panic 也會解鎖）
	c.count++     // 修改共享狀態（這段是「臨界區」，一次只有一個 goroutine 能執行）
}

// Value 安全地取得計數器的值
func (c *SafeCounter) Value() int { // 取得當前計數值
	c.mu.Lock()         // 加鎖（讀取也要加鎖，避免讀到「中間狀態」）
	defer c.mu.Unlock() // 解鎖
	return c.count      // 回傳計數值
}

// demonstrateMutex 示範 Mutex 的用法
func demonstrateMutex() { // 函式：示範 Mutex
	fmt.Println("\n=== 5. Mutex（互斥鎖）===") // 印出標題

	counter := &SafeCounter{} // 建立安全計數器
	var wg sync.WaitGroup     // WaitGroup 等待所有 goroutine 完成

	// 啟動 100 個 goroutine 同時對計數器加 1
	// 如果沒有 Mutex 保護，最終結果不一定是 100
	for i := 0; i < 100; i++ { // 啟動 100 個
		wg.Add(1)        // 登記一個 goroutine
		go func() {      // 並發執行
			defer wg.Done() // 完成時通知
			counter.Increment() // 安全地加 1
		}()
	}

	wg.Wait() // 等所有 100 個 goroutine 完成
	fmt.Printf("100 個 goroutine 並發累加後的結果：%d（一定是 100！）\n", counter.Value()) // 結果保證正確
}

// ==========================================================================
// 6. Context — 傳遞取消訊號和超時
// ==========================================================================
//
// Context 就像「遙控器」：
//   你可以隨時按停止，告訴所有正在工作的 goroutine「可以停了」
//   或者設定一個計時器：「5 秒後自動取消」
//
// 三種建立 Context 的方式：
//   context.Background()           → 根 Context，不會取消
//   context.WithCancel(parent)     → 可手動取消
//   context.WithTimeout(parent, d) → 超時自動取消
//   context.WithDeadline(parent, t)→ 截止時間自動取消

// longRunningTask 模擬一個長時間執行的任務，支援取消
func longRunningTask(ctx context.Context, name string) { // 接受 Context 和任務名稱
	for i := 1; i <= 10; i++ { // 模擬 10 個步驟
		select { // 同時監聽：取消訊號 和 工作計時器
		case <-ctx.Done(): // 如果收到取消訊號
			fmt.Printf("  [%s] 第 %d 步被取消：%v\n", name, i, ctx.Err()) // 印出取消原因
			return                                                          // 提前結束
		case <-time.After(30 * time.Millisecond): // 每 30ms 執行一步
			fmt.Printf("  [%s] 完成第 %d 步\n", name, i) // 印出進度
		}
	}
	fmt.Printf("  [%s] 全部完成！\n", name) // 所有步驟完成
}

// demonstrateContext 示範 Context 的用法
func demonstrateContext() { // 函式：示範 Context
	fmt.Println("\n=== 6. Context ===") // 印出標題

	// ---- context.WithCancel：手動取消 ----
	fmt.Println("\n--- WithCancel（手動取消）---") // 小標題

	ctx, cancel := context.WithCancel(context.Background()) // 建立可取消的 Context
	// cancel 是一個函式，呼叫它就會取消這個 Context
	// 注意：一定要呼叫 cancel 來釋放資源！

	go longRunningTask(ctx, "任務A") // 在 goroutine 中執行任務

	time.Sleep(80 * time.Millisecond) // 等 80ms（讓任務執行約 2-3 步）
	cancel()                          // 手動取消！發送取消訊號給所有使用這個 ctx 的地方
	time.Sleep(50 * time.Millisecond) // 等任務接收到取消訊號並清理

	// ---- context.WithTimeout：超時自動取消 ----
	fmt.Println("\n--- WithTimeout（超時自動取消）---") // 小標題

	// 建立一個 150ms 後自動取消的 Context
	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancelTimeout() // 即使 timeout 會自動取消，最佳實踐仍要呼叫 cancel 釋放資源

	go longRunningTask(ctxTimeout, "任務B") // 執行任務（超過 150ms 自動取消）

	time.Sleep(200 * time.Millisecond) // 等任務結束
}

// ==========================================================================
// 7. 實際應用：Worker Pool 模式
// ==========================================================================
//
// Worker Pool（工人池）是並發程式設計最常見的模式：
//   - 有一堆「工作」需要處理
//   - 固定 N 個「工人」（goroutine）同時工作
//   - 工作完成一個、接下一個，直到所有工作都完成
//
// 好處：
//   - 控制並發數量（避免開太多 goroutine 耗盡資源）
//   - 工作佇列（Job Queue）自動調度

// Job 代表一個工作任務
type Job struct { // 定義工作結構體
	ID   int // 工作 ID
	Data int // 工作資料（要處理的數字）
}

// Result 代表工作完成的結果
type Result struct { // 定義結果結構體
	JobID  int // 對應的工作 ID
	Output int // 計算結果
}

// worker 是工人函式：從 jobs channel 取工作，把結果放入 results channel
func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	// jobs <-chan Job        → 只能接收（<-chan）的 channel，防止 worker 誤發工作
	// results chan<- Result  → 只能發送（chan<-）的 channel，防止 worker 誤讀結果
	defer wg.Done() // 函式結束時通知 WaitGroup
	for job := range jobs { // 從 jobs channel 持續接收工作（直到 channel 關閉）
		// 模擬工作處理（計算平方）
		time.Sleep(10 * time.Millisecond)                          // 模擬處理耗時
		fmt.Printf("  Worker %d 處理工作 %d\n", id, job.ID)       // 印出哪個 worker 處理哪個工作
		result := Result{                                           // 建立結果
			JobID:  job.ID,                                        // 記錄工作 ID
			Output: job.Data * job.Data,                           // 計算：資料的平方
		}
		results <- result // 把結果放入 results channel
	}
}

// demonstrateWorkerPool 示範 Worker Pool 模式
func demonstrateWorkerPool() { // 函式：示範 Worker Pool
	fmt.Println("\n=== 7. Worker Pool 模式 ===") // 印出標題

	const numWorkers = 3  // 工人數量（固定 3 個 goroutine）
	const numJobs = 10    // 工作總數（10 個工作）

	jobs := make(chan Job, numJobs)        // 工作佇列（有緩衝，容量等於工作總數）
	results := make(chan Result, numJobs)  // 結果佇列（有緩衝）
	var wg sync.WaitGroup                 // 等待所有 worker 完成

	// 啟動固定數量的 worker goroutine
	for i := 1; i <= numWorkers; i++ { // 啟動 3 個 worker
		wg.Add(1)                                  // 登記一個 worker
		go worker(i, jobs, results, &wg)           // 啟動 worker goroutine
		fmt.Printf("  Worker %d 就緒\n", i)       // 印出 worker 就緒訊息
	}

	// 發送所有工作到 jobs channel
	for i := 1; i <= numJobs; i++ { // 發送 10 個工作
		jobs <- Job{ID: i, Data: i} // 把工作放入佇列（工作 i 的資料就是 i）
	}
	close(jobs) // 關閉 jobs channel，通知所有 worker「沒有更多工作了」
	// 注意：worker 的 for range 會在 channel 關閉後自動結束

	// 等所有 worker 完成後，關閉 results channel
	go func() {          // 在 goroutine 中等待，避免阻塞主程式
		wg.Wait()        // 等所有 worker 完成
		close(results)   // 關閉結果 channel，通知收集者「沒有更多結果了」
	}()

	// 收集所有結果
	totalOutput := 0                       // 累計所有結果
	for result := range results {          // 從 results channel 接收所有結果
		fmt.Printf("  工作 %d 完成：%d^2 = %d\n", result.JobID, result.JobID, result.Output) // 印出結果
		totalOutput += result.Output       // 累加
	}
	fmt.Printf("所有工作完成！1^2 + 2^2 + ... + 10^2 = %d\n", totalOutput) // 印出總和
}

// ==========================================================================
// 主程式：執行所有示範
// ==========================================================================

func main() { // 程式進入點
	fmt.Println("==========================================") // 分隔線
	fmt.Println(" 第十九課：Goroutine 與並發")               // 標題
	fmt.Println("==========================================") // 分隔線

	demonstrateGoroutine()  // 示範 1：Goroutine 基礎（循序 vs 並發）
	demonstrateWaitGroup()  // 示範 2：WaitGroup 等待所有 goroutine
	demonstrateChannel()    // 示範 3：Channel 傳遞資料
	demonstrateSelect()     // 示範 4：Select 監聽多個 channel
	demonstrateMutex()      // 示範 5：Mutex 保護共享資料
	demonstrateContext()    // 示範 6：Context 取消和超時
	demonstrateWorkerPool() // 示範 7：Worker Pool 實際應用

	fmt.Println("\n==========================================") // 分隔線
	fmt.Println(" 教學完成！")                                // 結尾
	fmt.Println("==========================================") // 分隔線
	fmt.Println()                                            // 空行
	fmt.Println("💡 進階挑戰：執行 go run -race ./tutorials/19-goroutines") // 提示 race detector
	fmt.Println("   這個旗標會啟用「競態條件偵測器」，幫你找並發 bug")          // 說明用途
}
