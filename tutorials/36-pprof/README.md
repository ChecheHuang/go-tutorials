# 第三十六課：pprof 效能分析

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | 了解 pprof 存在，知道如何取得基本 profile |
| 🔴 資深工程師 | **必備**：能找出 CPU 熱點、記憶體洩漏、goroutine 洩漏 |
| 🏢 SRE/效能工程師 | 核心技能，負責解決生產環境效能問題 |

## 核心用法

### 方法一：HTTP 端點（生產環境）

```go
import _ "net/http/pprof"  // 自動掛載到 DefaultServeMux

// 啟動 HTTP Server
http.ListenAndServe(":6060", nil)
```

```bash
# 取得 CPU Profile（30 秒採樣）
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 取得 Heap Profile
go tool pprof http://localhost:6060/debug/pprof/heap

# 查看 goroutine
curl http://localhost:6060/debug/pprof/goroutine?debug=2
```

### 方法二：Benchmark（最常用）

```bash
# 生成 CPU 和記憶體 profile
go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof .

# 分析（互動模式）
go tool pprof cpu.prof

# 瀏覽器火焰圖（需要 graphviz）
go tool pprof -http=:8081 cpu.prof
```

## pprof 互動指令

```
(pprof) top         # 最耗資源的前 10 個函式
(pprof) top -cum    # 按累計時間排序（找呼叫鏈的根源）
(pprof) list Foo    # 顯示 Foo 函式的原始碼+耗時
(pprof) web         # 在瀏覽器開啟調用圖（需要 graphviz）
```

## 常見問題診斷

| 問題 | 使用哪個 Profile |
|------|-----------------|
| CPU 使用率高 | `/debug/pprof/profile` |
| 記憶體持續增長 | `/debug/pprof/heap` |
| Goroutine 洩漏 | `/debug/pprof/goroutine` |
| 鎖競爭嚴重 | `/debug/pprof/mutex` |
| 頻繁 GC | `runtime.ReadMemStats` |

## 執行方式

```bash
go run ./tutorials/36-pprof

# 另一個終端
go tool pprof http://localhost:6060/debug/pprof/heap
```
