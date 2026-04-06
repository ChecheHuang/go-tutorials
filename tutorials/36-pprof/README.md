# 第三十六課：pprof 效能分析

> **一句話總結**：pprof 就像程式的 X 光片，讓你看透程式內部哪裡在浪費 CPU、哪裡在吃記憶體，用數據而非猜測來優化效能。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | 了解 pprof 存在，知道如何取得基本 profile |
| 🔴 資深工程師 | **必備**：能找出 CPU 熱點、記憶體洩漏、goroutine 洩漏 |
| 🏢 SRE/效能工程師 | 核心技能，負責解決生產環境效能問題 |

## 你會學到什麼？

- 為什麼效能優化要用數據而非直覺
- Go 的 5 種 Profile 類型：CPU、Heap、Goroutine、Block、Mutex
- 如何在程式中啟用 pprof（一行 import 搞定）
- `go tool pprof` 的互動指令：`top`、`list`、`web`、`png`
- 火焰圖（Flame Graph）怎麼看
- CPU Profiling 步驟拆解
- 記憶體分析：`alloc_space` vs `inuse_space`
- Goroutine 洩漏偵測
- 常見效能問題模式與修復方法
- 部落格專案該在什麼時候做 profiling

## 執行方式

```bash
# 啟動範例程式（會開啟 pprof HTTP 端點）
go run ./tutorials/36-pprof

# 另一個終端：取得 CPU Profile（30 秒採樣）
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 取得 Heap Profile
go tool pprof http://localhost:6060/debug/pprof/heap

# 用瀏覽器看火焰圖（需要安裝 graphviz）
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile?seconds=10
```

## 生活比喻：X 光片

```
你的程式跑很慢，你覺得是哪裡有問題？

❌ 用猜的（沒有 pprof）：
  「應該是資料庫查詢太慢吧？」→ 優化了三天
  「還是慢... 可能是 JSON 序列化？」→ 又改了兩天
  「怎麼還是慢！」→ 其實是一個迴圈裡不小心做了 N+1 查詢

✅ 用數據（有 pprof）：
  照一張 X 光片 → 清楚看到 70% CPU 花在 encodeJSON 函式
  → 精準定位問題 → 30 分鐘修好

  pprof 就像醫院的 X 光機：
  - CPU Profile    = 照骨科（看哪裡在「出力」）
  - Heap Profile   = 照內科（看哪裡在「佔空間」）
  - Goroutine      = 驗血報告（看有多少「工人」在忙）
  - Block Profile   = 照心臟（看哪裡在「阻塞」等待）
  - Mutex Profile  = 照關節（看哪裡在「卡住」搶鎖）
```

## 為什麼要用 pprof？不要用猜的

> **Premature optimization is the root of all evil.** — Donald Knuth

但反過來說，**沒有數據的優化更可怕**。你可能花了一週優化一個只佔 0.1% 時間的函式，而真正的瓶頸根本沒碰到。

| 方法 | 做法 | 結果 |
|------|------|------|
| 憑直覺 | 「這個迴圈看起來很慢」 | 可能優化了不重要的部分 |
| 用 pprof | 「數據顯示 72% 時間花在 JSON 編碼」 | 精準打擊瓶頸 |

**優化的正確流程**：

```
1. 先寫正確的程式碼
2. 用 Benchmark 確認「真的慢」（第 42 課）
3. 用 pprof 找到「哪裡慢」
4. 針對瓶頸優化
5. 再用 Benchmark 確認「真的變快了」
```

## Go 的 5 種 Profile 類型

| Profile | 端點 | 用途 | 什麼時候用 |
|---------|------|------|-----------|
| **CPU** | `/debug/pprof/profile` | CPU 時間花在哪些函式 | API 回應慢、CPU 使用率高 |
| **Heap** | `/debug/pprof/heap` | 記憶體分配在哪裡 | 記憶體持續增長、OOM |
| **Goroutine** | `/debug/pprof/goroutine` | 所有 goroutine 的堆疊 | goroutine 數量異常增長 |
| **Block** | `/debug/pprof/block` | goroutine 阻塞等待的地方 | 程式回應慢但 CPU 不高 |
| **Mutex** | `/debug/pprof/mutex` | 鎖競爭的熱點 | 多 goroutine 搶同一把鎖 |

> **注意**：Block 和 Mutex Profile 預設沒有開啟，需要手動設定取樣率：
> ```go
> runtime.SetBlockProfileRate(1)    // 開啟 Block profiling
> runtime.SetMutexProfileFraction(1) // 開啟 Mutex profiling
> ```

## 如何啟用 pprof

### 方法一：HTTP 端點（適合正式環境的 long-running 服務）

```go
import (
    "net/http"
    _ "net/http/pprof"  // 底線 import = 只要它的 init() 副作用
)

func main() {
    // 在獨立的 port 啟動 pprof（不要跟 API 混在一起！）
    go func() {
        // 注意：pprof 掛載在 DefaultServeMux 上
        // 生產環境建議限制存取（防火牆、VPN）
        http.ListenAndServe(":6060", nil)
    }()

    // 你的主要 API server
    startAPIServer(":8080")
}
```

```
import _ "net/http/pprof" 會自動註冊這些路由：

  /debug/pprof/              ← 索引頁面（瀏覽器可以直接看）
  /debug/pprof/profile       ← CPU Profile
  /debug/pprof/heap          ← Heap Profile
  /debug/pprof/goroutine     ← Goroutine 堆疊
  /debug/pprof/block         ← Block Profile
  /debug/pprof/mutex         ← Mutex Profile
  /debug/pprof/threadcreate  ← Thread 建立
  /debug/pprof/trace         ← Execution Trace
```

> **安全提醒**：pprof 端點可以暴露程式內部資訊，生產環境務必不要對外公開！建議用防火牆限制只有內網可以存取，或者只在需要時才啟動。

### 方法二：Benchmark 搭配 Profile（適合開發時期）

```bash
# 跑 Benchmark 時同時生成 CPU 和記憶體 profile
go test -bench=BenchmarkProcess -cpuprofile=cpu.prof -memprofile=mem.prof .

# 分析 CPU profile
go tool pprof cpu.prof

# 分析記憶體 profile
go tool pprof mem.prof
```

### 方法三：程式碼中手動採集（適合 CLI 工具或短期程式）

```go
import "runtime/pprof"

func main() {
    // CPU Profile
    f, _ := os.Create("cpu.prof")
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // 你的程式邏輯
    doWork()

    // Heap Profile（在程式結束前）
    f2, _ := os.Create("mem.prof")
    pprof.WriteHeapProfile(f2)
    f2.Close()
}
```

## go tool pprof 指令完整攻略

```bash
# 啟動互動模式
go tool pprof cpu.prof
```

進入互動模式後：

```
(pprof) top              # 顯示前 10 個最耗資源的函式
(pprof) top 20           # 顯示前 20 個
(pprof) top -cum         # 按累計時間排序（找呼叫鏈的根源）
(pprof) list handleOrder # 顯示 handleOrder 函式的原始碼 + 每行耗時
(pprof) web              # 在瀏覽器開啟呼叫圖（需要 graphviz）
(pprof) png              # 輸出呼叫圖為 PNG 圖片
(pprof) peek handleOrder # 顯示誰呼叫了 handleOrder，handleOrder 又呼叫了誰
(pprof) traces           # 顯示所有取樣的完整堆疊
(pprof) quit             # 離開
```

### top 輸出解讀

```
(pprof) top
Showing nodes accounting for 4.5s, 90% of 5s total
      flat  flat%   sum%        cum   cum%
      2.0s 40.00% 40.00%      3.0s 60.00%  main.encodeJSON
      1.5s 30.00% 70.00%      1.5s 30.00%  runtime.mallocgc
      0.5s 10.00% 80.00%      4.5s 90.00%  main.handleRequest
      0.3s  6.00% 86.00%      0.3s  6.00%  syscall.write
      0.2s  4.00% 90.00%      0.2s  4.00%  runtime.memmove
```

| 欄位 | 意義 | 解讀 |
|------|------|------|
| **flat** | 函式「自己」花的時間 | encodeJSON 本身花了 2 秒 |
| **flat%** | flat 佔全部時間的百分比 | 40% 的 CPU 時間在 encodeJSON |
| **sum%** | flat% 的累計值 | 前兩個函式佔了 70% |
| **cum** | 函式「自己 + 呼叫的函式」花的時間 | handleRequest 含子函式共 4.5 秒 |
| **cum%** | cum 佔全部時間的百分比 | handleRequest 鏈佔 90% |

> **技巧**：`flat` 高 = 這個函式本身很慢，直接優化它。`cum` 高但 `flat` 低 = 這個函式呼叫的子函式很慢，往下找。

## 火焰圖（Flame Graph）怎麼看

```bash
# 啟動 pprof web UI（包含火焰圖）
go tool pprof -http=:8081 cpu.prof
# 瀏覽器會自動開啟，點選 View > Flame Graph
```

```
火焰圖的閱讀方式：

         ┌──────────────────────── handleRequest ────────────────────────┐
         │                                                                │
    ┌─── encodeJSON ───┐    ┌── queryDB ──┐    ┌─ validate ─┐           │
    │                   │    │             │    │            │           │
    │  ┌── marshal ──┐  │    │ ┌─ scan ──┐ │    │            │           │
    │  │█████████████│  │    │ │█████████│ │    │            │           │
    └──┴─────────────┴──┘    └─┴─────────┴─┘    └────────────┘           │
    └────────────────────────────────────────────────────────────────────┘

    ↑ 每一層 = 呼叫堆疊的一層
    ← 寬度 = 佔用的 CPU 時間 →

    越寬 = 越耗 CPU = 越值得優化！

    看到一個「寬又平」的長方形？那就是效能瓶頸！
```

| 特徵 | 意義 | 行動 |
|------|------|------|
| 寬方塊 | 佔用大量 CPU 時間 | 這是優化的重點 |
| 高塔型 | 呼叫堆疊很深 | 考慮減少巢狀呼叫 |
| 鋸齒型 | 很多小函式 | 可能是頻繁的小分配 |

## CPU Profiling 步驟拆解

以部落格 API 為例，假設 `GET /articles` 回應很慢：

```bash
# Step 1：採集 30 秒的 CPU Profile（同時對 API 發請求施壓）
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Step 2：在另一個終端施壓
hey -n 1000 -c 10 http://localhost:8080/articles

# Step 3：進入互動模式後，先看 top
(pprof) top
      flat  flat%   sum%        cum   cum%
      1.2s 35.00% 35.00%      1.5s 44.00%  encoding/json.Marshal
      0.8s 23.00% 58.00%      0.8s 23.00%  runtime.mallocgc
      0.5s 15.00% 73.00%      2.8s 82.00%  handler.GetArticles

# Step 4：看 GetArticles 的原始碼
(pprof) list GetArticles
      0.1s     func GetArticles(c *gin.Context) {
      0.2s         articles := repo.FindAll()  // 資料庫查詢
      0.1s         for _, a := range articles {
      1.5s             json.Marshal(a)          // 瓶頸在這！每筆都序列化一次
      0.1s         }
               }

# Step 5：找到問題！應該一次序列化整個 slice，而不是逐筆
```

## 記憶體分析：alloc_space vs inuse_space

```bash
# 取得 Heap Profile
go tool pprof http://localhost:6060/debug/pprof/heap

# 查看「總共分配了多少」（找出頻繁分配記憶體的地方）
(pprof) top -sample_index=alloc_space

# 查看「目前還在用的」（找出記憶體洩漏）
(pprof) top -sample_index=inuse_space
```

| 指標 | 意義 | 何時使用 |
|------|------|---------|
| **alloc_space** | 程式啟動到現在「總共」分配了多少記憶體 | 找頻繁分配（造成 GC 壓力）|
| **inuse_space** | 程式「目前」正在使用的記憶體 | 找記憶體洩漏（持續增長）|
| **alloc_objects** | 總共分配了多少個物件 | 找小物件頻繁分配 |
| **inuse_objects** | 目前還存在的物件數量 | 找物件洩漏 |

```
alloc_space 高但 inuse_space 低：
  → 頻繁分配+釋放，GC 壓力大
  → 優化方向：減少分配（sync.Pool、預分配 slice）

inuse_space 持續增長：
  → 記憶體洩漏！有東西一直在增加但沒有釋放
  → 常見原因：goroutine 洩漏、全域 map 持續增長、忘記關閉 reader
```

## Goroutine 洩漏偵測

Goroutine 洩漏是 Go 中最常見的「記憶體洩漏」。每個 goroutine 至少佔 2KB 的 stack，洩漏上萬個 goroutine 就會吃掉大量記憶體。

```bash
# 查看所有 goroutine 的堆疊（debug=2 顯示完整資訊）
curl http://localhost:6060/debug/pprof/goroutine?debug=2

# 用 pprof 分析
go tool pprof http://localhost:6060/debug/pprof/goroutine
(pprof) top
```

### 常見的 goroutine 洩漏模式

```go
// ❌ 洩漏模式一：往沒有人讀的 channel 寫入
func leak1() {
    ch := make(chan int)
    go func() {
        ch <- 42  // 永遠阻塞！沒有人讀 ch
    }()
    // 函式返回了，但 goroutine 永遠卡在 ch <- 42
}

// ❌ 洩漏模式二：忘記取消 context
func leak2() {
    ctx, cancel := context.WithCancel(context.Background())
    // 忘記呼叫 cancel()！context 相關的 goroutine 永遠不會結束
    _ = ctx
}

// ❌ 洩漏模式三：HTTP Response Body 忘記關閉
func leak3() {
    resp, _ := http.Get("https://example.com")
    // 忘記 resp.Body.Close()！底層的 goroutine 不會被回收
}

// ✅ 正確做法
func noLeak() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()  // 一定要 defer cancel！

    ch := make(chan int, 1)  // 使用 buffered channel
    go func() {
        select {
        case ch <- 42:
        case <-ctx.Done():  // context 取消時退出
        }
    }()
}
```

### 偵測 goroutine 洩漏的方法

```go
// 方法一：在測試中監控 goroutine 數量
func TestNoGoroutineLeak(t *testing.T) {
    before := runtime.NumGoroutine()

    // 執行你要測試的操作
    doSomething()

    // 等待 goroutine 結束
    time.Sleep(100 * time.Millisecond)

    after := runtime.NumGoroutine()
    if after > before {
        t.Errorf("goroutine leak: before=%d, after=%d", before, after)
    }
}

// 方法二：使用 goleak 套件（Uber 開源）
// go get go.uber.org/goleak
func TestMain(m *testing.M) {
    goleak.VerifyTestMain(m)
}
```

## 常見效能問題模式與修復

| 問題 | pprof 症狀 | 修復方法 |
|------|-----------|---------|
| 頻繁 JSON 序列化 | CPU `flat` 高在 `encoding/json` | 改用 `jsoniter` 或 `sonic` |
| string 拼接 | `runtime.mallocgc` 佔比高 | 改用 `strings.Builder` |
| 頻繁建立小物件 | `alloc_objects` 很高 | 使用 `sync.Pool` |
| Slice 頻繁擴容 | `runtime.growslice` 佔比高 | 預分配 `make([]T, 0, n)` |
| 鎖競爭 | Mutex Profile 某個鎖很高 | 細粒度鎖、讀寫鎖、lock-free |
| Goroutine 洩漏 | goroutine 數量持續增長 | 確保 goroutine 有退出路徑 |
| N+1 查詢 | CPU `cum` 高在 DB 相關函式 | 批次查詢、JOIN、預載入 |

## 部落格專案：什麼時候該做 Profiling？

```
不需要 Profile 的情況：
  - 程式跑得好好的，回應時間符合預期
  - 還在開發初期，功能還沒寫完
  - 「我覺得這裡應該很慢」← 先量測再說！

需要 Profile 的情況：
  ✅ API 回應時間超過 SLA（例如 P99 > 500ms）
  ✅ 記憶體使用量持續增長（懷疑洩漏）
  ✅ goroutine 數量異常（從 Prometheus 監控看到）
  ✅ 上線前的效能基準測試
  ✅ 優化前，先 profile 確認瓶頸在哪
```

## 常見問題診斷速查表

| 問題 | 使用哪個 Profile | 指令 |
|------|-----------------|------|
| CPU 使用率高 | CPU Profile | `go tool pprof .../profile?seconds=30` |
| 記憶體持續增長 | Heap Profile | `go tool pprof .../heap` |
| Goroutine 洩漏 | Goroutine Profile | `go tool pprof .../goroutine` |
| 鎖競爭嚴重 | Mutex Profile | `go tool pprof .../mutex` |
| 程式回應慢但 CPU 不高 | Block Profile | `go tool pprof .../block` |
| 頻繁 GC | Heap + `runtime.ReadMemStats` | 檢查 `alloc_objects` |

## FAQ

### Q1：pprof 會影響正式環境的效能嗎？

HTTP 端點只在被請求時才啟動採樣，平時幾乎沒有開銷。CPU Profile 採樣時會有約 5% 的效能影響（每秒 100 次中斷），但只持續你指定的秒數。Heap Profile 的開銷更低。所以在正式環境中保留 pprof 端點是安全的，只是要注意不要對外公開。

### Q2：火焰圖需要安裝什麼工具？

使用 `go tool pprof -http=:8081 cpu.prof` 需要安裝 [graphviz](https://graphviz.org/)。安裝方式：macOS 用 `brew install graphviz`，Ubuntu 用 `apt install graphviz`，Windows 從官網下載安裝。Go 1.22+ 的 pprof web UI 已經很好用了，火焰圖是內建功能。

### Q3：alloc_space 和 inuse_space 該看哪一個？

如果你在找「記憶體洩漏」（記憶體持續增長不回收），看 `inuse_space`。如果你在找「GC 壓力大」（頻繁分配+回收），看 `alloc_space`。一般建議兩個都看，先用 `inuse_space` 排除洩漏，再用 `alloc_space` 優化分配。

### Q4：pprof 和 Benchmark（第 42 課）有什麼差別？

Benchmark 告訴你「有多慢」（每次操作花多少 ns、分配多少記憶體），pprof 告訴你「為什麼慢」（時間花在哪個函式、哪一行）。正確的優化流程是：先用 Benchmark 量測 → 用 pprof 找瓶頸 → 優化 → 再用 Benchmark 確認改善。

### Q5：看不懂 pprof 輸出中的 runtime 函式怎麼辦？

常見的 runtime 函式：`runtime.mallocgc` = 記憶體分配、`runtime.growslice` = slice 擴容、`runtime.mapaccess` = map 查詢、`runtime.gcBgMarkWorker` = GC 回收。如果這些函式佔比很高，通常代表你的程式有太多記憶體分配操作，可以用 `sync.Pool`、預分配等方式優化。

## 練習

1. 啟動部落格 API，用 `go tool pprof` 抓取 30 秒的 CPU profile，找出最耗 CPU 的函式
2. 用 `go tool pprof` 抓取 heap profile，找出佔用最多記憶體的物件
3. 用 `-http=:8081` 啟動 pprof web UI，查看火焰圖（flame graph）
4. 比較 `alloc_space` 和 `inuse_space` 的差異：前者是總分配量，後者是當前使用量
5. 寫一個會造成 goroutine leak 的範例，用 goroutine profile 偵測它

## 下一課預告

下一課我們會學習 **OpenTelemetry 分散式追蹤**——pprof 讓你看穿單一服務的效能問題，但當你的系統由多個微服務組成時，一個請求跨越多個服務，該怎麼追蹤「時間花在哪個服務的哪個環節」？這就是分散式追蹤要解決的問題。
