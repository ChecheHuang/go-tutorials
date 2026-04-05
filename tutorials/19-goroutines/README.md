# 第十九課：Goroutine 與並發（Goroutines & Concurrency）

> **一句話總結**：Goroutine 是 Go 最強大的特性——用一個 `go` 關鍵字就能同時處理多件事，比傳統執行緒輕上百倍。

## 你會學到什麼？

- 什麼是並發（Concurrency）以及它解決什麼問題
- `goroutine`：用 `go` 關鍵字啟動，成本極低
- `sync.WaitGroup`：等待一組 goroutine 全部完成
- `channel`：goroutine 之間傳遞資料的管道（無緩衝 vs 有緩衝）
- `select`：同時監聽多個 channel
- `sync.Mutex`：保護共享資料，避免競態條件
- `context.Context`：傳遞取消訊號和超時設定
- **Worker Pool 模式**：最實用的並發設計模式

## 執行方式

```bash
go run ./tutorials/19-goroutines
```

加上競態條件偵測（強烈建議開發時使用）：
```bash
go run -race ./tutorials/19-goroutines
```

## 生活比喻：餐廳廚房

```
沒有並發（一個廚師做所有事）：
  煮牛排（5分鐘）→ 煮義大利麵（4分鐘）→ 做沙拉（2分鐘）= 11 分鐘

有並發（三個廚師同時工作）：
  煮牛排  ───────────── 5 分鐘
  煮義麵  ──────────  4 分鐘
  做沙拉  ──────  2 分鐘
  ────────────────────────── 只需要 5 分鐘！

Goroutine = 廚師（超輕量，可以有幾千個）
Channel   = 傳菜口（廚師之間傳遞訊息）
WaitGroup = 出餐點名（等所有菜都準備好）
Mutex     = 共用刀具（一次只能一人使用）
```

## Goroutine 是什麼？有多輕量？

```
OS Thread（作業系統執行緒）：
  - 記憶體：約 1MB
  - 建立耗時：幾微秒（microseconds）
  - 由作業系統管理

Goroutine：
  - 記憶體：約 2KB（是 OS Thread 的 1/500！）
  - 建立耗時：幾奈秒（nanoseconds）
  - 由 Go runtime 管理

一個普通程式：
  Thread：幾十個就很多了
  Goroutine：輕鬆跑幾萬個
```

## 基本語法

```go
// 啟動 goroutine
go someFunction()    // 在新 goroutine 中執行函式
go func() {          // 匿名函式也可以
    // 做某些事
}()                  // 別忘了最後的 ()

// channel 操作
ch := make(chan int)     // 建立無緩衝 channel
ch := make(chan int, 10) // 建立有緩衝 channel（容量 10）
ch <- 42                 // 發送（可能阻塞）
value := <-ch            // 接收（可能阻塞）
close(ch)               // 關閉（接收方用 for range 會自動結束）
```

## WaitGroup：等待所有 Goroutine 完成

```go
var wg sync.WaitGroup

for i := 0; i < 5; i++ {
    wg.Add(1)           // 登記：我又開始了一個 goroutine
    go func(n int) {
        defer wg.Done() // 完成時：計數 -1
        process(n)
    }(i)                // 傳入 i 的值（避免 closure 陷阱！）
}

wg.Wait() // 阻塞：等所有 goroutine 的 Done() 都呼叫
```

**Closure 陷阱**（很常見的 bug）：

```go
// ❌ 錯誤：所有 goroutine 看到的 i 都是最後的值（5）
for i := 0; i < 5; i++ {
    go func() {
        fmt.Println(i) // 可能印出 5 5 5 5 5
    }()
}

// ✅ 正確：把 i 當參數傳進去，每個 goroutine 有自己的副本
for i := 0; i < 5; i++ {
    go func(n int) {
        fmt.Println(n) // 印出 0 1 2 3 4（順序不定）
    }(i)
}
```

## Channel：goroutine 之間的通訊

```
無緩衝 Channel（同步）：
  發送方 ──── 等待接收方就緒 ────▶ 接收方
  （像面對面交接物品，雙方都要在場）

有緩衝 Channel（非同步）：
  發送方 ──▶ [_ _ _ _] ──▶ 接收方
  （像信箱，可以先放進去，接收方稍後取）
```

```go
// 無緩衝：同步，一發一收
ch := make(chan string)
go func() { ch <- "hello" }()
msg := <-ch // 等 goroutine 發送

// 有緩衝：非同步，最多放 n 個
ch := make(chan int, 3)
ch <- 1  // 不阻塞（緩衝有空間）
ch <- 2  // 不阻塞
ch <- 3  // 不阻塞
// ch <- 4  // 阻塞！緩衝滿了

// 關閉 channel + for range 接收所有資料
close(ch)
for v := range ch {  // 直到 channel 關閉才停止
    fmt.Println(v)
}
```

## Select：同時監聽多個 Channel

```go
select {
case msg := <-ch1:        // 如果 ch1 有資料
    handleCh1(msg)
case msg := <-ch2:        // 如果 ch2 有資料
    handleCh2(msg)
case <-time.After(5 * time.Second): // 超時（5 秒後自動觸發）
    fmt.Println("超時！")
default:                   // 沒有任何 channel 就緒（非阻塞）
    fmt.Println("暫時沒有資料")
}
```

## Mutex：保護共享資料

**什麼是競態條件（Race Condition）？**

```
兩個 goroutine 同時讀寫同一個變數：

  goroutine A: 讀到 count=5 → 計算 5+1=6 → 等一下...
  goroutine B: 讀到 count=5 → 計算 5+1=6 → 寫入 count=6
  goroutine A:                              → 寫入 count=6

  最終 count=6，但應該是 7！（兩次加法只加了一次）
```

**用 Mutex 解決：**

```go
type SafeCounter struct {
    mu    sync.Mutex  // 互斥鎖
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()           // 加鎖：只有我能用
    defer c.mu.Unlock()   // 解鎖：我用完了
    c.count++             // 安全地修改
}
```

**用 race detector 找 bug：**

```bash
go run -race ./你的程式
# 如果有競態條件，race detector 會立刻告訴你！
```

## Context：取消和超時

```go
// 手動取消
ctx, cancel := context.WithCancel(context.Background())
defer cancel()  // 一定要呼叫！否則資源洩漏

go doWork(ctx)  // doWork 要定期檢查 ctx.Done()

time.Sleep(3 * time.Second)
cancel()  // 通知 doWork 停止

// ----

// 超時自動取消
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// doWork 5 秒後自動收到取消訊號

// ----

// doWork 內部如何響應取消：
func doWork(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():  // 收到取消訊號
            return          // 停止工作
        case <-time.After(100 * time.Millisecond):
            // 繼續做工作
        }
    }
}
```

## Worker Pool 模式（最重要的實際應用）

```
                    ┌─ Worker 1 ─┐
jobs channel ──────▶─ Worker 2 ─▶──── results channel
（工作佇列）         └─ Worker 3 ─┘   （結果佇列）

優點：
  - 固定 N 個 goroutine，控制資源使用
  - 工作自動分配，誰空誰處理
  - 無論有多少工作，goroutine 數量不變
```

```go
func worker(id int, jobs <-chan Job, results chan<- Result) {
    for job := range jobs {           // 持續取工作，直到 channel 關閉
        results <- process(job)       // 處理後送結果
    }
}

// 啟動工人池
jobs := make(chan Job, 100)
results := make(chan Result, 100)

for i := 0; i < 5; i++ {  // 固定 5 個 worker
    go worker(i, jobs, results)
}

// 發送工作
for _, j := range allJobs {
    jobs <- j
}
close(jobs)  // 關閉工作 channel，worker 的 range 會結束
```

## 在部落格專案中的對應

| 並發技術 | 使用場景 |
|---------|---------|
| **Goroutine** | 同時處理多個 HTTP 請求（Gin 自動為每個請求開 goroutine）|
| **Context** | 請求超時、客戶端斷線時取消資料庫查詢 |
| **Channel** | 非同步任務（如發送 Email 通知、影像處理）|
| **Worker Pool** | 批量處理（如批次匯出、批次發信）|
| **Mutex** | 記憶體快取的讀寫保護 |

## 常見問題 FAQ

### Q: goroutine 洩漏（Goroutine Leak）是什麼？

如果一個 goroutine 永遠無法結束（例如等待一個永遠沒有資料的 channel），就叫做 goroutine 洩漏。程式會越跑越慢，最終 OOM。

**解法**：用 Context 加上超時，確保 goroutine 一定會結束。

### Q: channel 什麼時候應該關閉？

只有「發送方」才應該關閉 channel。接收方不應該關閉（會 panic）。不是所有 channel 都需要關閉——只有當接收方需要知道「沒有更多資料了」時才需要關閉（如 for range）。

### Q: sync.RWMutex 和 sync.Mutex 差在哪？

- `sync.Mutex`：完全互斥，同一時間只有一人可以操作（讀寫都鎖）
- `sync.RWMutex`：讀共享，寫互斥——多個 goroutine 可以同時讀，但寫的時候完全互斥

讀多寫少的場景用 RWMutex 效能更好。

### Q: 應該用 channel 還是 Mutex？

```
用 channel 的時機：
  - 傳遞所有權（把資料從一個 goroutine 傳到另一個）
  - 協調多個 goroutine 的工作
  - 工作佇列、結果收集

用 Mutex 的時機：
  - 保護共享狀態（快取、計數器）
  - 需要原子性地讀取+修改
```

Go 的設計原則：**優先用 channel，次選 Mutex**。

## 練習

1. **並發計算**：用 goroutine 同時計算 1-100 的費氏數列（Fibonacci），收集所有結果
2. **Pipeline 模式**：建立一個資料處理管道：`生成數字 → 過濾偶數 → 計算平方`，每個步驟一個 goroutine
3. **超時控制**：寫一個函式模擬「慢速 API 呼叫」（sleep 3 秒），用 `context.WithTimeout` 設定 1 秒超時，觀察超時錯誤
4. **Race Detector**：把 `SafeCounter` 的 Mutex 移除，加上 `-race` 旗標執行，觀察 race detector 的輸出

## 下一課預告

**第二十課：Redis 快取** —— 學習如何用 Redis 加速應用程式，包括快取策略、Session 儲存、Rate Limiting，以及如何用 Docker 在本地快速啟動 Redis。
