// ==========================================================================
// 第二十一課：結構化日誌（Structured Logging with zap）
// ==========================================================================
//
// 什麼是日誌（Log）？
//   日誌就像程式的「黑盒子飛行記錄器」
//   當程式出問題時，你可以翻回去看「程式當時做了什麼」
//
// 為什麼不能繼續用 fmt.Println？
//
//   fmt.Println("使用者登入")
//   → 輸出：使用者登入
//   → 問題：幾點？哪個使用者？成功還是失敗？嚴重程度？
//   → 幾百行這樣的輸出，完全看不出問題在哪裡
//
//   zap 的輸出：
//   → {"level":"info","time":"2024-01-15T10:30:00Z","msg":"使用者登入",
//       "user_id":42,"ip":"192.168.1.1","duration_ms":23}
//   → 每個欄位都有名稱，可以用工具快速搜尋和過濾
//
// 結構化日誌的好處：
//   1. 機器可讀（JSON 格式，可以用 grep/jq 快速搜尋）
//   2. 帶有欄位（level、time、caller、plus 自訂欄位）
//   3. 有層級（DEBUG < INFO < WARN < ERROR < FATAL）
//   4. 極快速（zap 是 Go 最快的日誌套件之一）
//
// 什麼是 go.uber.org/zap？
//   zap 是 Uber 開源的 Go 日誌套件，是業界標準
//   特點：零記憶體分配、極快速、類型安全
//   提供兩種介面：
//     zap.Logger      → 快速、類型安全（需要明確指定欄位型別）
//     zap.SugaredLogger → 方便、printf 風格（稍慢但更容易寫）
//
// 執行方式：go run ./tutorials/21-structured-logging
// ==========================================================================

package main // 宣告這是 main 套件

import ( // 匯入所有需要的套件
	"fmt"  // 標準庫：格式化輸出（只用在最開頭的說明）
	"time" // 標準庫：時間相關功能

	"go.uber.org/zap"         // Uber 的結構化日誌套件（核心）
	"go.uber.org/zap/zapcore" // zap 的底層設定元件（用來自訂 logger）
)

// ==========================================================================
// 1. 為什麼需要結構化日誌
// ==========================================================================

// demonstrateWhyStructured 對比傳統日誌和結構化日誌
func demonstrateWhyStructured() { // 示範為什麼要結構化日誌
	fmt.Println("=== 1. 為什麼需要結構化日誌 ===\n") // 印出標題

	// ---- 傳統做法（不好）----
	fmt.Println("【傳統做法 - 純文字】")       // 說明傳統做法
	userID := 42                      // 模擬使用者 ID
	duration := 23 * time.Millisecond // 模擬操作耗時

	// 這樣寫的問題：沒辦法用工具快速過濾「只看 user_id=42 的錯誤」
	fmt.Printf("[INFO] 使用者 %d 登入成功，耗時 %v\n", userID, duration) // 純文字，難以解析
	fmt.Printf("[ERROR] 使用者 %d 密碼錯誤\n", userID)                // 純文字

	// ---- zap 做法（好）----
	fmt.Println("\n【zap 做法 - 結構化 JSON】") // 說明 zap 做法

	// zap.NewExample() 建立一個簡單的 JSON 格式 logger（示範用）
	logger, _ := zap.NewProduction() // 正式環境的 JSON logger
	defer logger.Sync()              // Sync 確保所有日誌都被寫出去（程式結束前很重要！）

	// 每個欄位都有名稱和型別，JSON 格式，工具可以解析
	logger.Info("使用者登入成功", // 日誌訊息
		zap.Int("user_id", userID),         // 整數欄位
		zap.Duration("duration", duration), // 時間長度欄位
		zap.String("ip", "192.168.1.1"),    // 字串欄位
	)
	logger.Error("使用者密碼錯誤", // ERROR 層級
		zap.Int("user_id", userID),               // 整數欄位
		zap.String("reason", "invalid_password"), // 字串欄位
	)

	fmt.Println("\n→ JSON 格式讓你可以執行：")                             // 說明 JSON 的好處
	fmt.Println("  cat app.log | jq 'select(.level==\"error\")'") // 只看錯誤
	fmt.Println("  cat app.log | jq 'select(.user_id==42)'")      // 只看特定使用者
}

// ==========================================================================
// 2. 開發模式 vs 正式模式
// ==========================================================================
//
// zap 提供兩種預設設定：
//
//   zap.NewDevelopment()
//     → 人類可讀的格式（帶顏色）
//     → 顯示所有層級（包含 DEBUG）
//     → 顯示程式碼位置（方便找問題）
//     → 適合：開發時在終端機看
//
//   zap.NewProduction()
//     → JSON 格式（機器可讀）
//     → 只顯示 INFO 以上（忽略 DEBUG）
//     → 適合：正式環境，輸出到日誌系統（ELK、Grafana Loki）

// demonstrateModes 示範開發模式和正式模式的差異
func demonstrateModes() { // 示範兩種模式
	fmt.Println("\n=== 2. 開發模式 vs 正式模式 ===\n") // 印出標題

	// ---- 開發模式：人類可讀 ----
	fmt.Println("【Development 模式（開發用）】") // 說明開發模式

	devLogger, _ := zap.NewDevelopment()     // 建立開發模式的 logger
	defer devLogger.Sync()                   // 確保緩衝都被寫出
	devLogger.Debug("這是 DEBUG 訊息（開發模式才看得到）") // DEBUG 層級
	devLogger.Info("這是 INFO 訊息")             // INFO 層級
	devLogger.Warn("這是 WARN 警告")             // WARN 層級（不嚴重但要注意）
	devLogger.Error("這是 ERROR 錯誤（但程式繼續跑）")   // ERROR 層級

	// ---- 正式模式：JSON ----
	fmt.Println("\n【Production 模式（正式環境）】") // 說明正式模式

	prodLogger, _ := zap.NewProduction() // 建立正式模式的 logger
	defer prodLogger.Sync()              // 確保緩衝都被寫出
	// 注意：Production 模式預設不顯示 DEBUG
	prodLogger.Debug("這行不會出現（Production 忽略 DEBUG）") // 不會被輸出
	prodLogger.Info("正式環境的 INFO 日誌",                // INFO 會輸出
		zap.String("version", "1.0.0"), // 帶欄位
	)
}

// ==========================================================================
// 3. 日誌層級（Log Levels）
// ==========================================================================
//
// 日誌層級從低到高：
//
//   DEBUG → 詳細的除錯資訊（開發時用，正式環境通常關閉）
//   INFO  → 正常的運作資訊（「使用者登入」「文章建立」）
//   WARN  → 警告，不影響運作但需要注意（「回應時間偏長」「快取未命中率高」）
//   ERROR → 錯誤，某個操作失敗（「資料庫查詢失敗」「API 呼叫逾時」）
//   FATAL → 嚴重錯誤，程式無法繼續（呼叫後程式會結束！）
//   PANIC → 呼叫 panic（不常用）
//
// 原則：
//   能 INFO 的不要 DEBUG（讓日誌有意義）
//   能 WARN 的不要 ERROR（ERROR 應該是真的需要處理的問題）
//   FATAL 謹慎使用（程式會直接退出）

// demonstrateLogLevels 示範各個日誌層級
func demonstrateLogLevels() { // 示範日誌層級
	fmt.Println("\n=== 3. 日誌層級 ===\n") // 印出標題

	logger, _ := zap.NewDevelopment() // 開發模式（可以看到所有層級）
	defer logger.Sync()               // 確保緩衝都被寫出

	// 每個層級的典型使用場景
	logger.Debug("開始處理訂單", // DEBUG：詳細除錯資訊
		zap.Int("order_id", 123), // 附上訂單 ID
	)
	logger.Info("訂單建立成功", // INFO：正常業務事件
		zap.Int("order_id", 123),       // 附上訂單 ID
		zap.String("product", "Go 書籍"), // 附上商品名稱
		zap.Float64("amount", 350.00),  // 附上金額
	)
	logger.Warn("庫存量偏低", // WARN：需要注意但不影響服務
		zap.String("product", "Go 書籍"), // 附上商品名稱
		zap.Int("remaining", 5),        // 剩餘數量
	)
	logger.Error("付款失敗", // ERROR：操作失敗，需要處理
		zap.Int("order_id", 123),       // 附上訂單 ID
		zap.String("reason", "卡片餘額不足"), // 失敗原因
	)
	// logger.Fatal("資料庫連線失敗，程式結束")  // FATAL：程式會立即退出！示範時先注解掉
}

// ==========================================================================
// 4. zap 欄位類型（Fields）
// ==========================================================================
//
// zap 的欄位必須明確指定型別，這樣可以避免 reflect 的效能損失：
//   zap.String(key, val)    → 字串
//   zap.Int(key, val)       → 整數（也有 Int64、Uint、Uint64）
//   zap.Float64(key, val)   → 浮點數
//   zap.Bool(key, val)      → 布林值
//   zap.Duration(key, val)  → 時間長度（time.Duration）
//   zap.Time(key, val)      → 時間點（time.Time）
//   zap.Error(err)          → 錯誤（特殊：不需要 key）
//   zap.Any(key, val)       → 任意型別（會用 reflect，稍慢）
//   zap.Strings(key, vals)  → 字串切片

// demonstrateFields 示範各種 zap 欄位類型
func demonstrateFields() { // 示範欄位類型
	fmt.Println("\n=== 4. 欄位類型（Fields）===\n") // 印出標題

	logger, _ := zap.NewDevelopment() // 開發模式
	defer logger.Sync()               // 確保緩衝都被寫出

	// 模擬一個 API 請求的完整日誌
	requestStart := time.Now()           // 記錄請求開始時間
	time.Sleep(15 * time.Millisecond)    // 模擬處理耗時
	duration := time.Since(requestStart) // 計算耗時

	logger.Info("API 請求完成",
		zap.String("method", "GET"),                 // 字串：HTTP 方法
		zap.String("path", "/api/v1/articles"),      // 字串：請求路徑
		zap.Int("status", 200),                      // 整數：HTTP 狀態碼
		zap.Duration("duration", duration),          // 時間長度：處理耗時
		zap.Int64("response_bytes", 1024),           // 64 位元整數：回應大小
		zap.Bool("cache_hit", true),                 // 布林值：是否命中快取
		zap.Strings("tags", []string{"go", "blog"}), // 字串切片
	)

	// 模擬錯誤日誌
	simulatedErr := fmt.Errorf("資料庫連線逾時") // 建立一個模擬錯誤

	logger.Error("資料庫查詢失敗",
		zap.String("query", "SELECT * FROM articles"), // 查詢語句
		zap.Int("user_id", 42),                        // 使用者 ID
		zap.Error(simulatedErr),                       // 錯誤物件（不需要 key）
		zap.Duration("timeout", 5*time.Second),        // 逾時設定
	)

	// zap.Any：任意型別（方便但慢一點）
	type RequestMeta struct { // 定義請求元資料結構
		UserAgent string // 瀏覽器識別
		ClientIP  string // 客戶端 IP
	}
	meta := RequestMeta{UserAgent: "Mozilla/5.0", ClientIP: "192.168.1.1"} // 建立元資料
	logger.Info("請求元資料",
		zap.Any("meta", meta), // 任意型別（會輸出整個 struct）
	)
}

// ==========================================================================
// 5. SugaredLogger — 更方便的介面
// ==========================================================================
//
// SugaredLogger 是 zap.Logger 的「糖衣」版本：
//   - 支援 printf 風格（Infof, Errorf...）
//   - 支援鍵值對（Infow, Errorw...）
//   - 略慢（因為用到 interface{}），但對效能不敏感的地方很方便
//
// 命名規則：
//   sugar.Info(args...)    → 純訊息
//   sugar.Infof("格式 %s", val)  → printf 風格
//   sugar.Infow("訊息", "key", val, "key2", val2)  → 鍵值對

// demonstrateSugared 示範 SugaredLogger 的用法
func demonstrateSugared() { // 示範 SugaredLogger
	fmt.Println("\n=== 5. SugaredLogger（更方便的介面）===\n") // 印出標題

	logger, _ := zap.NewDevelopment() // 建立基礎 logger
	defer logger.Sync()               // 確保緩衝都被寫出

	sugar := logger.Sugar() // 把 Logger 轉成 SugaredLogger（呼叫 .Sugar()）

	userName := "Alice" // 模擬使用者名稱
	userAge := 25       // 模擬年齡

	// Infof：printf 風格（像 fmt.Sprintf）
	sugar.Infof("使用者 %s（%d 歲）已登入", userName, userAge) // 格式化字串

	// Infow：鍵值對風格（key-value pairs，不需要 zap.String 等）
	sugar.Infow("使用者操作", // 訊息
		"action", "create_article", // 鍵值對（偶數個 → key, value, key, value...）
		"user", userName, // 鍵值對
		"article_id", 100, // 鍵值對（自動偵測型別）
	)

	// Warnf：警告，printf 風格
	sugar.Warnf("回應時間偏長：%v（超過閾值 100ms）", 150*time.Millisecond)

	// 從 SugaredLogger 轉回 Logger（需要最高效能的地方用 Logger）
	baseLogger := sugar.Desugar() // 轉回標準 Logger
	baseLogger.Info("這是從 Sugar 轉回來的標準 Logger")
}

// ==========================================================================
// 6. With() — 預設欄位（最重要的功能之一）
// ==========================================================================
//
// logger.With() 建立一個「帶有預設欄位」的子 logger
// 每次用這個子 logger 記錄日誌，都會自動帶上那些欄位
//
// 最常見的用法：在 HTTP 請求進來時，建立一個帶有 request_id 的 logger
// 之後這個請求相關的所有日誌都用這個 logger，方便在幾百萬筆日誌中追蹤同一個請求

// demonstrateWith 示範 With() 的用法
func demonstrateWith() { // 示範 With
	fmt.Println("\n=== 6. With()：預設欄位 ===\n") // 印出標題

	// 全域 logger（應用程式啟動時建立一次）
	baseLogger, _ := zap.NewDevelopment() // 建立基礎 logger
	defer baseLogger.Sync()               // 確保緩衝都被寫出

	// 模擬一個 HTTP 請求進來
	requestID := "req-abc-123" // 模擬的 Request ID（真實場景用 uuid）
	userID := 42               // 模擬的使用者 ID

	// 為這個請求建立一個「子 logger」，帶有 request_id 和 user_id
	// 之後這個請求的所有日誌都用 reqLogger，不需要每次都加這兩個欄位
	reqLogger := baseLogger.With( // 建立子 logger
		zap.String("request_id", requestID), // 預設帶 request_id
		zap.Int("user_id", userID),          // 預設帶 user_id
	)

	// 之後這個請求的所有操作，都用 reqLogger
	// 每行日誌都會自動帶上 request_id 和 user_id，不需要重複寫
	reqLogger.Info("開始處理請求")                                   // 自動帶 request_id 和 user_id
	reqLogger.Info("查詢資料庫", zap.String("table", "articles"))   // 自動帶 + 額外欄位
	reqLogger.Info("快取命中", zap.String("cache_key", "user:42")) // 自動帶 + 額外欄位
	reqLogger.Info("請求完成",                                     // 自動帶 request_id 和 user_id
		zap.Int("status", 200),                     // 額外：狀態碼
		zap.Duration("total", 23*time.Millisecond), // 額外：總耗時
	)

	// 這樣在日誌系統中，只要過濾 request_id="req-abc-123"，就能看到這個請求的完整軌跡
	fmt.Println("\n→ 在日誌系統中執行：")                                       // 說明用法
	fmt.Println("  jq 'select(.request_id==\"req-abc-123\")' app.log") // 過濾特定請求
}

// ==========================================================================
// 7. Named Logger — 帶名稱的子 logger
// ==========================================================================
//
// logger.Named(name) 給 logger 加上名稱
// 日誌輸出中會帶有 "logger": "name"
// 適合用來區分不同模組的日誌（handler、usecase、repository）

// demonstrateNamed 示範 Named Logger
func demonstrateNamed() { // 示範 Named Logger
	fmt.Println("\n=== 7. Named Logger（模組日誌）===\n") // 印出標題

	baseLogger, _ := zap.NewDevelopment() // 建立基礎 logger
	defer baseLogger.Sync()               // 確保緩衝都被寫出

	// 為不同模組建立帶名稱的子 logger
	// 日誌輸出中會顯示 "logger": "handler" 等，方便知道日誌來自哪個模組
	handlerLogger := baseLogger.Named("handler") // Handler 層的 logger
	usecaseLogger := baseLogger.Named("usecase") // Usecase 層的 logger
	repoLogger := baseLogger.Named("repository") // Repository 層的 logger

	// 模擬一個請求流過所有層
	handlerLogger.Info("收到 POST /articles 請求",
		zap.String("user_agent", "Mozilla/5.0"), // 瀏覽器資訊
	)
	usecaseLogger.Info("驗證文章資料",
		zap.String("title", "Go 並發指南"), // 文章標題
	)
	repoLogger.Info("執行 INSERT INTO articles",
		zap.Duration("query_time", 5*time.Millisecond), // 查詢耗時
	)
	handlerLogger.Info("回應 201 Created",
		zap.Int("article_id", 101), // 新建文章 ID
	)
}

// ==========================================================================
// 8. 自訂 Logger 設定
// ==========================================================================
//
// zap 的 NewProduction() 和 NewDevelopment() 是預設設定
// 用 zap.Config 和 zapcore 可以完全自訂：
//   - 日誌輸出到哪裡（stdout、檔案、多個地方）
//   - 使用哪種格式（JSON、console）
//   - 最低日誌層級（只顯示 WARN 以上）
//   - 時間格式
//   - 是否顯示 caller（程式碼位置）

// demonstrateCustomLogger 示範自訂 Logger 設定
func demonstrateCustomLogger() { // 示範自訂設定
	fmt.Println("\n=== 8. 自訂 Logger 設定 ===\n") // 印出標題

	// 使用 zap.Config 建立自訂的 logger
	config := zap.Config{ // 建立 Logger 設定
		Level:    zap.NewAtomicLevelAt(zap.WarnLevel), // 只顯示 WARN 以上（忽略 INFO 和 DEBUG）
		Encoding: "json",                              // 輸出格式：json 或 console

		// OutputPaths：日誌輸出到哪裡（可以同時輸出到多個地方）
		OutputPaths: []string{"stdout"}, // stdout = 終端機
		// 正式環境可以加上檔案：[]string{"stdout", "/var/log/app.log"}

		// ErrorOutputPaths：WARN 以上的日誌輸出到哪裡
		ErrorOutputPaths: []string{"stderr"}, // stderr = 標準錯誤輸出

		// EncoderConfig：設定每個欄位的鍵名和格式
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "msg",                         // 訊息的 key 名稱（預設 "msg"）
			LevelKey:     "level",                       // 層級的 key 名稱（預設 "level"）
			TimeKey:      "time",                        // 時間的 key 名稱
			CallerKey:    "caller",                      // 程式碼位置的 key 名稱
			EncodeLevel:  zapcore.LowercaseLevelEncoder, // 層級格式："info"（小寫）
			EncodeTime:   zapcore.ISO8601TimeEncoder,    // 時間格式："2024-01-15T10:30:00.000Z"
			EncodeCaller: zapcore.ShortCallerEncoder,    // 位置格式："main.go:42"
		},
	}

	customLogger, err := config.Build() // 根據設定建立 Logger
	if err != nil {                     // 如果設定有誤
		fmt.Printf("建立 Logger 失敗: %v\n", err) // 印出錯誤
		return                                // 提前返回
	}
	defer customLogger.Sync() // 確保緩衝都被寫出

	customLogger.Debug("這行不會出現（低於 WARN）") // 不輸出
	customLogger.Info("這行也不會出現（低於 WARN）") // 不輸出
	customLogger.Warn("這行會出現！",           // WARN 以上才輸出
		zap.String("reason", "記憶體使用率超過 80%"), // 附加欄位
	)
	customLogger.Error("這行也會出現",
		zap.String("component", "database"), // 附加欄位
	)
}

// ==========================================================================
// 9. 在 Gin HTTP 中介層使用 zap
// ==========================================================================
//
// 這是 zap 在 Web 服務中最重要的應用：
//   每個 HTTP 請求進來時，建立一個帶有 request_id 的 logger
//   整個請求處理過程的日誌都帶有這個 ID
//   這樣可以在幾百萬筆日誌中，找到同一個請求的完整軌跡
//
// 這只是示範程式碼，展示概念，不實際啟動 HTTP 伺服器

// simulateHTTPRequest 模擬一個完整的 HTTP 請求日誌流程
func simulateHTTPRequest(baseLogger *zap.Logger, requestID string, method string, path string) {
	// 步驟 1：請求進來時，建立帶有 request_id 的子 logger
	logger := baseLogger.With( // 建立帶有固定欄位的子 logger
		zap.String("request_id", requestID), // 每個請求的唯一 ID
		zap.String("method", method),        // HTTP 方法
		zap.String("path", path),            // 請求路徑
	)

	start := time.Now() // 記錄請求開始時間

	// 步驟 2：整個請求處理過程的日誌都用這個 logger
	logger.Info("收到請求") // 自動帶 request_id, method, path

	// 模擬 usecase 和 repository 層的日誌（在真實專案中，logger 會被傳遞進去）
	logger.Info("查詢資料庫",
		zap.String("table", "articles"), // 查詢的表格
	)
	logger.Info("快取命中",
		zap.String("cache_key", "articles:page:1"), // 命中的快取 key
	)

	// 步驟 3：請求結束時記錄總耗時和狀態碼
	logger.Info("請求完成",
		zap.Int("status", 200),                      // HTTP 狀態碼
		zap.Duration("duration", time.Since(start)), // 總耗時
		zap.Int64("bytes", 2048),                    // 回應大小
	)
}

// demonstrateHTTPLogging 示範 HTTP 請求的結構化日誌
func demonstrateHTTPLogging() { // 示範 HTTP 日誌
	fmt.Println("\n=== 9. HTTP 請求日誌（Request ID 追蹤）===\n") // 印出標題

	baseLogger, _ := zap.NewDevelopment() // 建立基礎 logger
	defer baseLogger.Sync()               // 確保緩衝都被寫出

	// 模擬三個不同的 HTTP 請求同時進來（並發，request_id 讓你能分清楚誰是誰）
	simulateHTTPRequest(baseLogger, "req-001", "GET", "/api/articles")  // 請求 1
	simulateHTTPRequest(baseLogger, "req-002", "POST", "/api/articles") // 請求 2
	simulateHTTPRequest(baseLogger, "req-003", "GET", "/api/users/42")  // 請求 3

	fmt.Println("\n→ 在日誌中找同一個請求的所有記錄：")                            // 說明
	fmt.Println("  jq 'select(.request_id==\"req-002\")' app.log") // 過濾指令
}

// ==========================================================================
// 主程式：執行所有示範
// ==========================================================================

func main() { // 程式進入點
	fmt.Println("==========================================") // 分隔線
	fmt.Println(" 第二十一課：結構化日誌（zap）")                          // 標題
	fmt.Println("==========================================") // 分隔線

	demonstrateWhyStructured() // 示範 1：為什麼需要結構化日誌
	demonstrateModes()         // 示範 2：開發模式 vs 正式模式
	demonstrateLogLevels()     // 示範 3：日誌層級
	demonstrateFields()        // 示範 4：欄位類型
	demonstrateSugared()       // 示範 5：SugaredLogger
	demonstrateWith()          // 示範 6：With() 預設欄位
	demonstrateNamed()         // 示範 7：Named Logger
	demonstrateCustomLogger()  // 示範 8：自訂設定
	demonstrateHTTPLogging()   // 示範 9：HTTP 請求日誌

	fmt.Println("\n==========================================") // 分隔線
	fmt.Println(" 教學完成！")                                       // 結尾
	fmt.Println("==========================================")   // 分隔線
}
