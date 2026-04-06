# 第二十二課：Benchmark + Load Testing（效能測試）

> **一句話總結**：「我覺得很快」不算數——用數據證明你的程式碼夠快。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | **重點**：學會用 `go test -bench` 測量效能，讀懂 benchmark 結果 |
| 🔴 資深工程師 | **必備**：效能瓶頸分析、壓力測試、容量規劃、benchstat 統計比較 |

## 你會學到什麼？

- 為什麼需要 Benchmark（不是「感覺快」，而是「數據證明快」）
- `go test -bench` 語法：如何撰寫 Benchmark 函式
- `b.N`、`b.ResetTimer()`、`b.ReportAllocs()` 的用途
- 讀懂 Benchmark 輸出：iterations、ns/op、B/op、allocs/op
- `benchstat`：統計比較兩次 benchmark 結果
- 記憶體分析基礎：`-benchmem` 旗標
- 常見效能優化模式：strings.Builder、slice 預分配、sync.Pool、map vs slice
- 壓力測試工具：wrk、vegeta、k6
- `pprof` 效能剖析預覽（第 36 課詳細教學）
- 部落格專案中該 benchmark 什麼

## 執行方式

```bash
# 執行 Demo（互動式效能比較展示）
go run ./tutorials/22-benchmark/

# 執行正式 Benchmark（這才是標準做法）
go test -bench=. -benchmem ./tutorials/22-benchmark/

# 只跑特定 Benchmark（名稱包含 Concat）
go test -bench=Concat -benchmem ./tutorials/22-benchmark/

# 跑 Benchmark 並輸出到檔案（供 benchstat 比較用）
go test -bench=. -benchmem -count=10 ./tutorials/22-benchmark/ > old.txt
```

## 生活比喻：Benchmark = 體檢報告

```
沒有 Benchmark 的開發：
  工程師 A：「我改了字串處理，感覺快了不少」
  工程師 B：「真的嗎？我覺得沒什麼差」
  主管：「到底快了多少？能上線嗎？」
  → 沒人知道，憑感覺做決定

有 Benchmark 的開發：
  工程師 A：「我改了字串處理，Benchmark 顯示：
    舊版：1,500,000 ns/op、5,000,000 B/op
    新版：   30,000 ns/op、   65,536 B/op
    速度提升 50 倍、記憶體減少 98%」
  主管：「數據很明確，合併！」
  → 用數據說話，專業的工程決策

Benchmark 就像年度體檢報告：
  你「感覺」自己很健康 ≠ 你真的健康
  體檢報告上的數字才是客觀事實
  程式碼也一樣——跑 Benchmark 才知道真正的效能
```

## 為什麼需要 Benchmark？

很多工程師會犯一個錯誤：**憑直覺優化程式碼**。

```
常見錯誤思路：
  「用 + 串接字串一定很慢吧？」 → 不一定，要看串接幾次
  「map 一定比 slice 快吧？」   → 不一定，要看資料量大小
  「預分配記憶體一定值得吧？」  → 不一定，要看使用場景

正確思路：
  1. 先寫出正確的程式碼
  2. 用 Benchmark 量測效能
  3. 找到瓶頸後，才開始優化
  4. 優化後再次 Benchmark，確認真的變快了
```

> "Premature optimization is the root of all evil." — Donald Knuth
>
> 過早優化是萬惡之源。先量測，再優化。

## 如何撰寫 Benchmark 函式

Go 的 benchmark 寫在 `_test.go` 檔案中，函式名稱必須以 `Benchmark` 開頭：

```go
// benchmark_test.go
package main

import "testing"

// 函式名稱：Benchmark + 要測的功能
// 參數：*testing.B（不是 *testing.T）
func BenchmarkConcatPlus(b *testing.B) {
    // b.N 由 Go 自動決定，會自動增加直到結果穩定
    for i := 0; i < b.N; i++ {
        ConcatPlus(1000)
    }
}

func BenchmarkConcatBuilder(b *testing.B) {
    for i := 0; i < b.N; i++ {
        ConcatBuilder(1000)
    }
}
```

### b.N：Go 幫你決定跑幾次

`b.N` 不是你設定的數字。Go 的測試框架會自動調整：

```
第 1 輪：b.N = 1        → 測一次，太快了，不準
第 2 輪：b.N = 100      → 測 100 次，還是太快
第 3 輪：b.N = 10000    → 測 10000 次，結果穩定了
→ 回報：10000 次、每次 XXX ns
```

### b.ResetTimer()：排除初始化時間

當 benchmark 需要前置準備（比如建立測試資料），用 `b.ResetTimer()` 排除準備時間：

```go
func BenchmarkSearchMap(b *testing.B) {
    // 準備測試資料（不計入 benchmark 時間）
    data := make(map[string]bool, 10000)
    for i := 0; i < 10000; i++ {
        data[fmt.Sprintf("key-%d", i)] = true
    }

    b.ResetTimer() // 重設計時器！從這裡開始才計時

    for i := 0; i < b.N; i++ {
        _ = data["key-9999"]
    }
}
```

### b.ReportAllocs()：強制回報記憶體分配

即使沒有加 `-benchmem` 旗標，也能在程式碼中強制回報：

```go
func BenchmarkConcatPlus(b *testing.B) {
    b.ReportAllocs() // 強制顯示記憶體分配資訊
    for i := 0; i < b.N; i++ {
        ConcatPlus(1000)
    }
}
```

## 讀懂 Benchmark 輸出

執行 `go test -bench=. -benchmem ./tutorials/22-benchmark/` 後的輸出：

```
BenchmarkConcatPlus-8      1000    1500000 ns/op   5000000 B/op   1000 allocs/op
BenchmarkConcatBuilder-8  50000      30000 ns/op     65536 B/op      1 allocs/op
```

每個欄位的含義：

```
BenchmarkConcatPlus-8      1000    1500000 ns/op   5000000 B/op   1000 allocs/op
        │              │     │          │                │              │
     函式名稱      CPU 核心數  │     每次耗時        每次分配記憶體    分配次數
                          執行次數  (奈秒/次操作)    (位元組/次操作)  (次/次操作)
```

### 欄位詳解

| 欄位 | 範例值 | 含義 |
|------|--------|------|
| 函式名稱-N | `BenchmarkConcatPlus-8` | Benchmark 函式名稱，`-8` 表示使用 8 個 CPU 核心 |
| 執行次數 | `1000` | Go 自動決定的迭代次數（`b.N`），越大代表單次操作越快 |
| ns/op | `1500000` | 每次操作的耗時（奈秒），**越小越好** |
| B/op | `5000000` | 每次操作分配的記憶體（位元組），**越小越好** |
| allocs/op | `1000` | 每次操作的記憶體分配次數，**越小越好** |

### 結果解讀範例

```
ConcatPlus:    1,500,000 ns/op = 1.5 毫秒/次
ConcatBuilder:    30,000 ns/op = 0.03 毫秒/次
→ Builder 快 50 倍！

ConcatPlus:    5,000,000 B/op = 每次分配 ~5MB
ConcatBuilder:    65,536 B/op = 每次分配 ~64KB
→ Builder 省 98% 記憶體！

ConcatPlus:    1000 allocs/op = 每次操作分配 1000 次記憶體
ConcatBuilder:    1 allocs/op  = 每次操作只分配 1 次
→ Builder 的 GC 壓力幾乎為零！
```

## benchstat：統計比較兩次結果

單次 benchmark 結果可能受到系統負載影響，不夠準確。`benchstat` 能做統計分析：

```bash
# 安裝 benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# 跑多次 benchmark，得到統計上有意義的結果
go test -bench=. -benchmem -count=10 ./tutorials/22-benchmark/ > old.txt

# 優化程式碼後，再跑一次
go test -bench=. -benchmem -count=10 ./tutorials/22-benchmark/ > new.txt

# 比較兩次結果
benchstat old.txt new.txt
```

輸出範例：

```
goos: linux
goarch: amd64
pkg: github.com/user/project/tutorials/22-benchmark
                     │  old.txt   │              new.txt               │
                     │   sec/op   │   sec/op     vs base               │
ConcatPlus-8           1.500m ± 2%   1.480m ± 1%       ~ (p=0.108 n=10)
ConcatBuilder-8        30.00µ ± 3%   28.50µ ± 2%  -5.00% (p=0.001 n=10)
```

| 符號 | 含義 |
|------|------|
| `± 2%` | 標準差，表示結果的波動幅度 |
| `~` | 沒有統計顯著差異（可能只是噪音） |
| `-5.00%` | 快了 5%，且統計上有意義 |
| `p=0.001` | p 值越小，結果越可靠（通常 p < 0.05 就算顯著） |

## 記憶體分析基礎：-benchmem

`-benchmem` 旗標讓你看到每次操作的記憶體分配情況：

```bash
# 沒有 -benchmem：只看到時間
go test -bench=. ./tutorials/22-benchmark/
# BenchmarkConcatPlus-8      1000    1500000 ns/op

# 有 -benchmem：看到時間 + 記憶體
go test -bench=. -benchmem ./tutorials/22-benchmark/
# BenchmarkConcatPlus-8      1000    1500000 ns/op   5000000 B/op   1000 allocs/op
```

### B/op 和 allocs/op 為什麼重要？

```
Go 有垃圾回收器（GC）。每次分配記憶體，GC 以後都要回收。

allocs/op 高 → GC 要做的工作多 → 程式會有短暫停頓（STW pause）
B/op 高     → 消耗更多記憶體   → 可能觸發更頻繁的 GC

所以好的效能優化 = 減少 allocs/op + 減少 B/op
```

## 常見效能優化模式

### 1. strings.Builder 取代 + 串接

```go
// ❌ 慢：每次 + 都建立新字串，O(n²)
func slow(n int) string {
    s := ""
    for i := 0; i < n; i++ {
        s += "a"
    }
    return s
}

// ✅ 快：預分配緩衝區，O(n)
func fast(n int) string {
    var b strings.Builder
    b.Grow(n) // 預分配容量
    for i := 0; i < n; i++ {
        b.WriteString("a")
    }
    return b.String()
}
```

### 2. Slice 預分配

```go
// ❌ 慢：動態增長，多次重新分配
func slow(n int) []int {
    var s []int // len=0, cap=0
    for i := 0; i < n; i++ {
        s = append(s, i) // cap 不夠就要擴容
    }
    return s
}

// ✅ 快：一次分配到位
func fast(n int) []int {
    s := make([]int, 0, n) // len=0, cap=n
    for i := 0; i < n; i++ {
        s = append(s, i) // 不需要擴容
    }
    return s
}
```

### 3. sync.Pool：重複使用物件

```go
import "sync"

// 建立一個 Pool（物件池）
var bufPool = sync.Pool{
    New: func() any {
        return new(bytes.Buffer)
    },
}

func processRequest() {
    // 從 Pool 借一個 Buffer（而不是每次 new 一個）
    buf := bufPool.Get().(*bytes.Buffer)
    buf.Reset()
    defer bufPool.Put(buf) // 用完歸還 Pool

    buf.WriteString("處理結果...")
    // 使用 buf
}

// 適用場景：高頻率建立/銷毀的短生命週期物件
// 例如：HTTP handler 中的 buffer、JSON encoder
```

### 4. Map vs Slice 查找

```go
// Slice 線性搜尋：O(n) — 資料量小時可以
func SearchSlice(data []string, target string) bool {
    for _, v := range data {
        if v == target {
            return true
        }
    }
    return false
}

// Map 查找：O(1) — 資料量大時明顯勝出
func SearchMap(data map[string]bool, target string) bool {
    return data[target]
}
```

| 資料量 | Slice 線性搜尋 | Map O(1) 查找 | 建議 |
|--------|---------------|---------------|------|
| < 10 | 極快 | 極快 | 都可以，slice 更簡單 |
| 10 ~ 100 | 快 | 極快 | 看使用頻率決定 |
| 100 ~ 1000 | 開始變慢 | 極快 | 建議用 map |
| > 1000 | 明顯瓶頸 | 極快 | **一定要用 map** |

## 壓力測試（Load Testing）

Benchmark 測的是單一函式的效能。壓力測試測的是**整個系統在大量請求下的表現**。

```
Benchmark = 測單一零件的品質
壓力測試  = 測整台車在高速公路上的表現
```

### 工具一覽

| 工具 | 語言 | 特點 | 適合場景 |
|------|------|------|----------|
| **wrk** | C | 極高效能，簡單易用 | 快速壓測 HTTP API |
| **vegeta** | Go | 精確控制 QPS，輸出詳細報告 | 精準負載測試 |
| **k6** | Go/JS | 用 JavaScript 寫測試腳本 | 複雜情境壓測 |
| **ab** | C | Apache 內建，簡單 | 快速驗證 |

### wrk 範例

```bash
# 安裝（macOS）
brew install wrk

# 10 個連線、2 個執行緒、持續 30 秒
wrk -t2 -c10 -d30s http://localhost:8080/api/v1/articles

# 輸出範例：
# Running 30s test @ http://localhost:8080/api/v1/articles
#   2 threads and 10 connections
#   Thread Stats   Avg      Stdev     Max   +/- Stdev
#     Latency     2.35ms    1.12ms   28.32ms   89.21%
#     Req/Sec     2.15k   198.32     2.89k     72.33%
#   128456 requests in 30.01s, 245.67MB read
# Requests/sec:   4280.12
# Transfer/sec:      8.19MB
```

| wrk 參數 | 含義 |
|-----------|------|
| `-t2` | 2 個執行緒 |
| `-c10` | 10 個同時連線 |
| `-d30s` | 持續 30 秒 |
| `Latency` | 回應延遲（越低越好） |
| `Req/Sec` | 每秒請求數（越高越好） |

### vegeta 範例

```bash
# 安裝
go install github.com/tsenart/vegeta@latest

# 每秒 100 個請求、持續 10 秒
echo "GET http://localhost:8080/api/v1/articles" | vegeta attack -rate=100 -duration=10s | vegeta report

# 輸出範例：
# Requests      [total, rate, throughput]  1000, 100.10, 99.85
# Duration      [total, attack, wait]      10.015s, 9.99s, 25.123ms
# Latencies     [min, mean, 50, 90, 95, 99, max]  1.234ms, 5.678ms, 4.321ms, 9.876ms, 15.432ms, 28.765ms, 45.678ms
# Success       [ratio]  100.00%
# Status Codes  [code:count]  200:1000

# 輸出延遲分佈圖
echo "GET http://localhost:8080/api/v1/articles" | vegeta attack -rate=100 -duration=10s | vegeta plot > plot.html
```

| vegeta 參數 | 含義 |
|-------------|------|
| `-rate=100` | 每秒發送 100 個請求（精準控制 QPS） |
| `-duration=10s` | 持續 10 秒 |
| `vegeta report` | 產生文字報告 |
| `vegeta plot` | 產生 HTML 延遲圖表 |

### k6 範例

```bash
# 安裝
brew install k6  # 或 go install go.k6.io/k6@latest
```

```javascript
// load_test.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '10s', target: 10 },   // 10 秒內增加到 10 個虛擬使用者
    { duration: '30s', target: 10 },   // 維持 10 個虛擬使用者 30 秒
    { duration: '10s', target: 0 },    // 10 秒內降到 0
  ],
};

export default function () {
  const res = http.get('http://localhost:8080/api/v1/articles');
  check(res, {
    '狀態碼為 200': (r) => r.status === 200,
    '回應時間 < 200ms': (r) => r.timings.duration < 200,
  });
  sleep(1);
}
```

```bash
k6 run load_test.js
```

## 效能剖析預覽：pprof

`pprof` 是 Go 內建的效能剖析工具，可以找出程式碼中最耗時、最耗記憶體的地方。

```bash
# 在 benchmark 中產生 CPU profile
go test -bench=. -cpuprofile=cpu.prof ./tutorials/22-benchmark/

# 在 benchmark 中產生記憶體 profile
go test -bench=. -memprofile=mem.prof ./tutorials/22-benchmark/

# 用 pprof 分析（會開啟互動式介面）
go tool pprof cpu.prof
```

> **注意**：pprof 的詳細用法會在**第 36 課**深入教學，這裡只是讓你知道有這個工具存在。目前只要會用 `-bench` 和 `-benchmem` 就夠了。

## 在部落格專案中該 Benchmark 什麼？

```
值得 Benchmark 的地方：

  1. API 端點回應時間
     → 用 wrk 或 vegeta 壓測 GET /api/v1/articles

  2. 資料庫查詢
     → Benchmark ORM 查詢 vs 原生 SQL
     → Benchmark 有索引 vs 無索引

  3. JSON 序列化/反序列化
     → Benchmark encoding/json vs easyjson vs sonic

  4. 中介層（Middleware）開銷
     → Benchmark 加了 3 層 middleware 後的效能影響

  5. 快取命中 vs 未命中
     → Benchmark Redis 快取 vs 直接查 DB

不值得 Benchmark 的地方：

  ✗ 簡單的 CRUD 邏輯（瓶頸不在這裡）
  ✗ 只跑一次的初始化程式碼
  ✗ 還沒有使用者的功能（先求正確，再求快）
```

## 什麼時候不該優化？

```
                 ┌─────────────────────────────────────┐
                 │         過早優化的危險               │
                 ├─────────────────────────────────────┤
                 │                                     │
                 │  1. 程式碼還沒寫完 → 先求正確       │
                 │  2. 沒有 Benchmark 數據 → 先量測    │
                 │  3. 沒有效能問題 → 別動它            │
                 │  4. 只快了 5% → 不值得犧牲可讀性    │
                 │                                     │
                 │  正確的順序：                        │
                 │  Make it work → Make it right →     │
                 │  Make it fast                        │
                 │                                     │
                 └─────────────────────────────────────┘
```

**優化的黃金法則**：

1. **先寫正確的程式碼**（Make it work）
2. **再寫乾淨的程式碼**（Make it right）
3. **有效能問題時，先 Benchmark 找瓶頸**
4. **只優化瓶頸，不要優化「感覺慢」的地方**
5. **優化後再 Benchmark，確認真的變快了**

## 常見問題 FAQ

### Q: Benchmark 結果每次都不一樣，正常嗎？

正常。受到 CPU 頻率、背景程式、溫度等因素影響。解決方法：

```bash
# 多跑幾次取平均（-count=10 表示跑 10 次）
go test -bench=. -benchmem -count=10 ./tutorials/22-benchmark/

# 用 benchstat 做統計分析
go test -bench=. -count=10 ./tutorials/22-benchmark/ | benchstat
```

跑 benchmark 時建議：關掉瀏覽器、音樂播放器等吃 CPU 的程式。

### Q: b.N 能自己設定嗎？

不能，也不應該。`b.N` 是 Go 測試框架自動調整的。如果你想控制迭代次數，說明你對 benchmark 的理解有誤——重點不是「跑幾次」，而是「每次操作花多少時間」。

### Q: -bench=. 的 . 是什麼意思？

`.` 是正則表達式，匹配所有 Benchmark 函式。其他範例：

```bash
-bench=.              # 跑所有 Benchmark
-bench=Concat         # 只跑名稱含 "Concat" 的
-bench=^BenchmarkSlice  # 只跑以 "BenchmarkSlice" 開頭的
-bench=Builder$       # 只跑以 "Builder" 結尾的
```

### Q: Benchmark 和壓力測試有什麼差別？

| 比較 | Benchmark (`go test -bench`) | 壓力測試（wrk/vegeta） |
|------|---------------------------|----------------------|
| 測試對象 | 單一函式 | 整個 HTTP 服務 |
| 目的 | 微觀效能（ns 級別） | 巨觀效能（RPS、延遲） |
| 環境 | 不需要啟動服務 | 需要啟動服務 |
| 結果 | ns/op、B/op、allocs/op | RPS、延遲百分位、錯誤率 |
| 用途 | 比較演算法、資料結構 | 容量規劃、找系統瓶頸 |

### Q: 什麼時候該用 sync.Pool？

當你的程式碼**高頻率建立和銷毀短生命週期物件**時，例如：

- HTTP handler 中每個請求都要建立的 buffer
- JSON encoder/decoder
- 臨時的 byte slice

**不適合**用 sync.Pool 的情況：

- 長生命週期物件（應該用全域變數）
- 低頻率操作（Pool 的開銷可能比直接 new 還大）
- 物件很小（GC 回收小物件的成本很低）

## 練習

1. **撰寫 Map vs Slice Benchmark**：在 `benchmark_test.go` 中新增 `BenchmarkSearchSlice` 和 `BenchmarkSearchMap`，使用 `b.ResetTimer()` 排除資料準備時間，比較在 10,000 筆資料中搜尋最後一個元素的效能差異

2. **benchstat 統計比較**：分別把目前的 benchmark 結果存成 `old.txt`，然後把 `ConcatPlus` 中的 `n` 改成 500，重新跑 benchmark 存成 `new.txt`，用 `benchstat old.txt new.txt` 比較差異

3. **寫一個 JSON 序列化 Benchmark**：建立一個包含 10 個欄位的 struct，benchmark `encoding/json.Marshal` 的效能，觀察 B/op 和 allocs/op

4. **模擬壓力測試**：啟動部落格專案的 HTTP 伺服器，用 wrk 或 vegeta 壓測 `GET /api/v1/articles`，記錄 RPS 和平均延遲，嘗試調整連線數觀察變化

5. **sync.Pool 實驗**：寫一個 Benchmark 比較「每次 `new(bytes.Buffer)`」vs「用 `sync.Pool` 重複使用 Buffer」的效能差異，觀察 allocs/op 的變化

## 下一課預告

**第二十三課：Docker 容器化** — 把應用程式打包成 Docker 映像，用 docker-compose 一鍵啟動整個開發環境（Go 服務 + PostgreSQL + Redis）。
