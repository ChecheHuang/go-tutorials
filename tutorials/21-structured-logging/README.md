# 第二十一課：結構化日誌（Structured Logging）

> **一句話總結**：結構化日誌就像把 `fmt.Println` 升級成「帶標籤的 JSON」，讓你在幾百萬筆日誌中一秒找到問題。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | **重點**：結構化日誌是生產環境可觀測性的基礎 |
| 🔴 資深工程師 | 日誌策略（採樣、分層）、整合 ELK/Loki |

## 你會學到什麼？

- 為什麼 `fmt.Println` 在正式環境不夠用
- `go.uber.org/zap` 套件：Go 最快的日誌套件
- **日誌層級**：DEBUG / INFO / WARN / ERROR / FATAL
- **zap.Logger vs zap.SugaredLogger**：型別安全 vs 方便
- **With()**：預設欄位，最重要的功能之一
- **Named Logger**：區分不同模組的日誌
- **自訂設定**：控制格式、輸出目標、最低層級
- **Request ID 追蹤**：在 HTTP 請求中追蹤完整軌跡

## 執行方式

```bash
go run ./tutorials/21-structured-logging
```

## 生活比喻：飛機黑盒子

```
沒有結構化日誌（fmt.Println）：
  10:30 使用者登入
  10:30 建立文章
  10:31 使用者登入
  10:31 錯誤！
  ↑ 哪個使用者？什麼錯誤？幾號文章？完全不知道！

有結構化日誌（zap）：
  {"time":"10:30:00","level":"info","msg":"使用者登入","user_id":42,"ip":"1.2.3.4"}
  {"time":"10:30:01","level":"info","msg":"建立文章","user_id":42,"article_id":100}
  {"time":"10:31:00","level":"info","msg":"使用者登入","user_id":99,"ip":"5.6.7.8"}
  {"time":"10:31:01","level":"error","msg":"付款失敗","user_id":99,"reason":"卡片餘額不足"}

  搜尋 user_id=99 的問題：
  cat app.log | jq 'select(.user_id==99)'
  → 一秒找到！
```

## 什麼是 go.uber.org/zap？

`go.uber.org/zap` 是 Uber 開源的 Go 日誌套件，是業界標準：

```bash
go get go.uber.org/zap
```

特點：
- **零記憶體分配**：不像 `fmt.Sprintf` 會產生垃圾，幾乎不影響效能
- **極快速**：比 `log` 標準庫快 10 倍以上
- **類型安全**：欄位必須指定型別，避免執行時錯誤
- **兩種介面**：Logger（快）和 SugaredLogger（方便）

## 快速開始

```go
import "go.uber.org/zap"

// 開發環境（人類可讀）
logger, _ := zap.NewDevelopment()
defer logger.Sync() // 程式結束前要 Sync！

// 正式環境（JSON 格式）
logger, _ := zap.NewProduction()
defer logger.Sync()

// 記錄日誌
logger.Info("使用者登入成功",
    zap.Int("user_id", 42),
    zap.String("ip", "1.2.3.4"),
    zap.Duration("duration", 23*time.Millisecond),
)
```

## 開發模式 vs 正式模式

| 特性 | NewDevelopment() | NewProduction() |
|------|-----------------|-----------------|
| 輸出格式 | 人類可讀（彩色）| JSON |
| 最低層級 | DEBUG（全部顯示）| INFO（忽略 DEBUG）|
| 顯示程式碼位置 | 是 | 是 |
| 適合場景 | 開發時終端機看 | 正式環境輸出到日誌系統 |

## 日誌層級

```
層級（從低到高）：

  DEBUG ── 詳細除錯資訊（開發時用）
    ↓       例：「開始處理第 3 步」
  INFO  ── 正常業務事件
    ↓       例：「使用者登入成功」「訂單建立」
  WARN  ── 警告，不影響服務但需要注意
    ↓       例：「回應時間偏長（200ms）」「快取未命中率高」
  ERROR ── 操作失敗，需要人工處理
    ↓       例：「資料庫查詢失敗」「付款 API 逾時」
  FATAL ── 嚴重錯誤，程式無法繼續（呼叫後程式立即退出！）
            例：「資料庫無法連線，程式結束」
```

**原則**：
- 能 INFO 的不要 DEBUG
- 能 WARN 的不要 ERROR（ERROR 代表「需要有人去修」）
- FATAL 謹慎使用，只用在程式無法恢復的情況

## 欄位類型速查

| 函式 | 範例 | 說明 |
|------|------|------|
| `zap.String` | `zap.String("method", "GET")` | 字串 |
| `zap.Int` | `zap.Int("status", 200)` | 整數 |
| `zap.Int64` | `zap.Int64("bytes", 1024)` | 64 位元整數 |
| `zap.Float64` | `zap.Float64("score", 9.5)` | 浮點數 |
| `zap.Bool` | `zap.Bool("cache_hit", true)` | 布林值 |
| `zap.Duration` | `zap.Duration("latency", 23*time.Millisecond)` | 時間長度 |
| `zap.Error` | `zap.Error(err)` | 錯誤（不需要 key）|
| `zap.Strings` | `zap.Strings("tags", []string{"go"})` | 字串切片 |
| `zap.Any` | `zap.Any("meta", someStruct)` | 任意型別（慢一點）|

## zap.Logger vs zap.SugaredLogger

```go
// zap.Logger — 快速、型別安全（正式場景推薦）
logger.Info("使用者登入",
    zap.String("user", "Alice"),
    zap.Int("id", 42),
)

// zap.SugaredLogger — 方便、printf 風格（快速開發時用）
sugar := logger.Sugar()

sugar.Infof("使用者 %s（id=%d）已登入", "Alice", 42)  // printf 風格

sugar.Infow("使用者登入",                              // 鍵值對風格
    "user", "Alice",
    "id", 42,
)

// 互相轉換
baseLogger := sugar.Desugar()  // Sugar → Logger
sugar := logger.Sugar()         // Logger → Sugar
```

## With()：最重要的功能

```go
// 全域 logger
baseLogger, _ := zap.NewProduction()

// HTTP 請求進來時，建立帶有 request_id 的子 logger
reqLogger := baseLogger.With(
    zap.String("request_id", "req-abc-123"),  // 預設帶這個欄位
    zap.Int("user_id", 42),                   // 預設帶這個欄位
)

// 之後整個請求用 reqLogger，不需要每次都加 request_id
reqLogger.Info("查詢資料庫")    // 自動帶 request_id 和 user_id
reqLogger.Info("快取命中")      // 自動帶 request_id 和 user_id
reqLogger.Info("請求完成")      // 自動帶 request_id 和 user_id

// 在日誌系統中，過濾同一個請求的所有記錄：
// jq 'select(.request_id=="req-abc-123")' app.log
```

## Named Logger：模組區分

```go
baseLogger, _ := zap.NewDevelopment()

// 不同模組用不同名稱，日誌中會顯示 "logger":"handler" 等
handlerLogger := baseLogger.Named("handler")
usecaseLogger := baseLogger.Named("usecase")
repoLogger    := baseLogger.Named("repository")

handlerLogger.Info("收到請求")          // logger: "handler"
usecaseLogger.Info("驗證資料")         // logger: "usecase"
repoLogger.Info("執行 SQL 查詢")       // logger: "repository"
```

## 自訂 Logger 設定

```go
config := zap.Config{
    Level:            zap.NewAtomicLevelAt(zap.WarnLevel), // 只顯示 WARN 以上
    Encoding:         "json",                               // 格式：json 或 console
    OutputPaths:      []string{"stdout"},                   // 輸出到終端機
    ErrorOutputPaths: []string{"stderr"},
    EncoderConfig: zapcore.EncoderConfig{
        MessageKey: "msg",
        LevelKey:   "level",
        TimeKey:    "time",
        EncodeLevel: zapcore.LowercaseLevelEncoder,  // "info"（小寫）
        EncodeTime:  zapcore.ISO8601TimeEncoder,      // "2024-01-15T10:30:00Z"
    },
}

logger, _ := config.Build()
defer logger.Sync()
```

## 在部落格專案中的應用

```go
// main.go 啟動時初始化 logger
var logger *zap.Logger

func init() {
    if os.Getenv("GIN_MODE") == "release" {
        logger, _ = zap.NewProduction()
    } else {
        logger, _ = zap.NewDevelopment()
    }
}

// middleware/logging.go：每個請求建立帶 request_id 的子 logger
func LoggingMiddleware(baseLogger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := uuid.New().String()
        reqLogger := baseLogger.With(
            zap.String("request_id", requestID),
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
        )
        c.Set("logger", reqLogger)  // 存到 context，傳遞給 handler

        c.Next()

        reqLogger.Info("請求完成",
            zap.Int("status", c.Writer.Status()),
            zap.Duration("duration", time.Since(start)),
        )
    }
}

// handler 中取得 logger
func (h *ArticleHandler) GetArticle(c *gin.Context) {
    logger := c.MustGet("logger").(*zap.Logger)
    logger.Info("查詢文章", zap.String("id", c.Param("id")))
}
```

## 為什麼要 defer logger.Sync()？

zap 為了效能，日誌先寫入緩衝區，不是立刻寫到磁碟。

`Sync()` 會把緩衝區的內容強制寫出（flush）。如果程式突然結束，沒有呼叫 `Sync()`，最後幾筆日誌可能會遺失。

```go
// 最佳實踐：每次建立 logger 後立即 defer Sync
logger, _ := zap.NewProduction()
defer logger.Sync()  // 程式結束時自動執行
```

## 常見問題 FAQ

### Q: Development 和 Production 模式輸出長什麼樣？

**Development（開發）：**
```
2024-01-15T10:30:00.123+0800  INFO  main.go:42  使用者登入成功  {"user_id": 42}
```

**Production（正式）：**
```json
{"level":"info","ts":1705282200.123,"caller":"main.go:42","msg":"使用者登入成功","user_id":42}
```

### Q: 日誌要輸出到哪裡？

- **開發**：`stdout`（終端機看）
- **正式**：`stdout` + 用 Docker/K8s 的日誌收集器（如 Fluentd）收集
- **本地檔案**：`OutputPaths: []string{"stdout", "/var/log/app.log"}`

### Q: Sync() 有時候會報錯怎麼辦？

在某些環境（如 /dev/stdout），`Sync()` 可能回傳 "invalid argument" 錯誤，這是正常的，可以忽略：

```go
defer func() { _ = logger.Sync() }()
```

### Q: 能不能動態改變日誌層級？

可以！用 `zap.NewAtomicLevelAt()` 建立的層級可以在執行時修改：

```go
level := zap.NewAtomicLevelAt(zap.InfoLevel)
// 之後在收到 HTTP 請求時動態切換
level.SetLevel(zap.DebugLevel)  // 切換到 DEBUG
```

## 練習

1. **比較輸出差異**：同一段程式碼，分別用 `NewDevelopment()` 和 `NewProduction()` 執行，觀察輸出格式的差異
2. **Request ID 追蹤**：建立一個模擬的 HTTP handler，每次請求用 `uuid.New()` 生成 ID，用 `With()` 建立子 logger，觀察同一請求的日誌如何關聯
3. **自訂層級**：用 `zap.Config` 建立一個「只輸出 ERROR 以上」的 logger，確認 INFO 和 WARN 都被過濾掉
4. **模組分層**：用 `Named()` 為 handler、usecase、repository 三層各建立一個 logger，模擬一個完整請求流過三層的日誌輸出

## 下一課預告

**第二十二課：資料庫進階（Database Advanced）** —— 學習索引優化、N+1 查詢問題、golang-migrate 資料庫遷移管理，以及如何用 `EXPLAIN` 分析慢查詢。
