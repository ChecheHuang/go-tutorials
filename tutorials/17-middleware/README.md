# 第十七課：中介層（Middleware）

> **一句話總結**：中介層就是「夾在請求和 Handler 之間的程式碼」，讓你把每個請求都需要的邏輯（日誌、驗證、跨域）抽出來，不用每個 Handler 重複寫。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | **重點**：中介層是橫切關注點的標準解法 |
| 🔴 資深工程師 | 設計可組合的中介層鏈，避免重複程式碼 |

## 你會學到什麼？

- 什麼是中介層（Middleware）以及它解決什麼問題
- Gin 中介層的洋蔥模型（Onion Model）
- `c.Next()` 和 `c.Abort()` 的差異和使用時機
- 全域中介層 vs 路由群組中介層
- 實作 4 種常見中介層：Logger、Auth、CORS、Recovery
- 閉包（Closure）如何讓中介層帶參數

## 執行方式

```bash
go run ./tutorials/17-middleware
```

然後打開另一個終端機，用 curl 測試：

```bash
# 公開路由（不需要 Key）
curl http://localhost:9090/

# 沒有 API Key → 401 Unauthorized
curl http://localhost:9090/api/profile

# 有正確 API Key → 200 OK
curl -H "X-API-Key: my-secret-key" http://localhost:9090/api/profile

# 測試 panic recovery（伺服器不會崩潰！）
curl http://localhost:9090/panic

# 測試限流（第 4 次開始會被拒絕）
curl http://localhost:9090/limited/resource
curl http://localhost:9090/limited/resource
curl http://localhost:9090/limited/resource
curl http://localhost:9090/limited/resource  # 這次會被拒絕
```

## 生活比喻：電影院安檢

```
你（請求）→ 安檢門（中介層）→ 售票員（Handler）→ 你（回應）

安檢門的工作：
  1. 記錄幾點進來、幾點離開  → Logger 中介層
  2. 檢查有沒有票（身份驗證）→ Auth 中介層
  3. 允許外地人進入（跨域）  → CORS 中介層
  4. 有人暈倒叫救護車（panic）→ Recovery 中介層
```

## 洋蔥模型（最重要的概念）

```
請求 ──────────────────────────────▶
                                    │
  ┌────── Recovery ──────────────────┤
  │ ┌──── Logger ─────────────────┐  │
  │ │ ┌── Auth ────────────────┐  │  │
  │ │ │                        │  │  │
  │ │ │   Handler（核心邏輯）    │  │  │
  │ │ │                        │  │  │
  │ │ └────────────────────────┘  │  │
  │ └──────────────────────────-──┘  │
  └──────────────────────────────────┘
                                    │
回應 ◀──────────────────────────────

進入順序：Recovery → Logger → Auth → Handler
離開順序：Handler → Auth → Logger → Recovery
```

這就是為什麼叫「洋蔥模型」：一層包一層，請求從外到內，回應從內到外。

## c.Next() vs c.Abort()

這是中介層最核心的兩個函式：

| 函式 | 作用 | 類比 |
|------|------|------|
| `c.Next()` | 繼續往內層執行 | 安檢通過，往前走 |
| `c.Abort()` | 中止！後面都不執行 | 安檢攔下，退出去 |

```go
// c.Next() 之前的程式碼 = 請求「進來」時執行
// c.Next() 呼叫 → 控制權交給下一層
// c.Next() 之後的程式碼 = 回應「出去」時執行

func LoggerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()          // 請求進來：記錄開始時間
        c.Next()                     // 等待 Handler 執行完
        duration := time.Since(start) // 回應出去：計算耗時
    }
}
```

## 中介層的作用範圍

```go
// 全域中介層：所有路由都生效
r.Use(RecoveryMiddleware())
r.Use(LoggerMiddleware())

// 路由群組中介層：只對 /api 底下的路由生效
api := r.Group("/api")
api.Use(AuthMiddleware())

// 單一路由中介層：只對這一個路由生效
r.GET("/special", SpecialMiddleware(), handler)
```

## 4 種常見中介層說明

### Logger 中介層

記錄每個請求的方法、路徑、狀態碼、耗時：

```
[開始] GET /api/profile
[完成] GET /api/profile → 200 (1.23ms)
```

### Auth 中介層

```go
apiKey := c.GetHeader("X-API-Key")  // 從 Header 取得 Key
if apiKey == "" {
    c.JSON(401, gin.H{"error": "缺少 API Key"})
    c.Abort()   // 中止！不繼續執行
    return
}
c.Set("user", "verified-user")  // 通過驗證，把資訊存入 Context
c.Next()
```

### CORS 中介層

```
瀏覽器安全限制：https://frontend.com 不能呼叫 https://api.com
CORS 的作用：讓 https://api.com 說「我允許 https://frontend.com 呼叫我」
```

### Recovery 中介層

```go
defer func() {
    if err := recover(); err != nil {
        // 接住 panic，回傳 500，伺服器不崩潰
    }
}()
c.Next()
```

## 帶參數的中介層（閉包）

```go
// 工廠函式：傳入參數，回傳中介層
func RateLimiter(maxRequests int) gin.HandlerFunc {
    requestCount := 0  // 閉包變數：被所有請求共享

    return func(c *gin.Context) {
        requestCount++
        if requestCount > maxRequests {
            c.JSON(429, gin.H{"error": "超過限制"})
            c.Abort()
            return
        }
        c.Next()
    }
}

// 使用：限制最多 3 次
limited.Use(RateLimiter(3))
```

**閉包的魔法**：`requestCount` 在 `RateLimiter(3)` 被呼叫時建立，之後每次請求都會修改同一個 `requestCount`。這個變數「活在」閉包裡，不會消失。

## c.Set / c.Get：在中介層和 Handler 之間傳遞資料

```go
// 在中介層中設定
c.Set("user_id", 42)
c.Set("role", "admin")

// 在 Handler 中取得
userID, exists := c.Get("user_id")  // exists 表示有沒有這個 key
if exists {
    fmt.Println(userID)  // 42
}
```

就像「購物車」——中介層把驗證好的使用者資訊放進去，Handler 直接拿來用。

## 在部落格專案中的對應

`internal/middleware/` 目錄：

```go
// middleware/logger.go
func Logger() gin.HandlerFunc { ... }

// middleware/recovery.go
func Recovery() gin.HandlerFunc { ... }

// middleware/cors.go
func CORS() gin.HandlerFunc { ... }

// middleware/jwt.go
func JWTAuth(cfg *config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        // 驗證 JWT Token（不是 API Key）
        // 把 user_id 存入 Context
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
```

`internal/handler/router.go`：
```go
r.Use(middleware.Logger())    // 全域
r.Use(middleware.Recovery())  // 全域

authenticated := v1.Group("")
authenticated.Use(middleware.JWTAuth(cfg))  // 只保護需要認證的路由
```

## 常見問題 FAQ

### Q: 為什麼 Recovery 要放最外層？

因為如果 Logger 或 Auth 發生 panic，只有放在它們「外面」的 Recovery 才能攔截到。中間層越外，保護範圍越大。

### Q: `c.Abort()` 後面還能寫程式碼嗎？

可以，但要手動 `return`。`c.Abort()` 只是設定一個旗標，讓 Gin 知道不要繼續往下，但當前函式還是會繼續執行到 `return`。

```go
c.Abort()  // 設定「中止」旗標
return      // 函式提前返回（別忘了！）
```

### Q: 全域 `r.Use()` 和群組 `api.Use()` 有什麼差？

- `r.Use()`：全部路由都執行這個中介層
- `api.Use()`：只有 `/api` 底下的路由執行這個中介層

通常 Logger 和 Recovery 是全域的（所有請求都需要），Auth 是群組的（只有部分路由需要認證）。

### Q: gin.Default() 和 gin.New() 差別是什麼？

```go
gin.Default()  // = gin.New() + 預設的 Logger + 預設的 Recovery
gin.New()      // 完全空的，你自己加中介層（更靈活）
```

## 練習

1. **RequestID 中介層**：為每個請求產生一個唯一 ID（用 `time.Now().UnixNano()`），加到回應 Header 中，格式：`X-Request-ID: 1234567890`
2. **AdminOnly 中介層**：檢查 Header 的 `X-Role` 是否為 `admin`，不是就回傳 403 Forbidden
3. **觀察 Logger**：在伺服器日誌中找到 `[開始]` 和 `[完成]` 的輸出，理解 c.Next() 前後的執行順序
4. **測試 Recovery**：執行 `curl http://localhost:9090/panic`，確認伺服器沒有崩潰，繼續回應後續請求

## 下一課預告

**第十六課：JWT 認證（JWT Authentication）** —— 學習如何用 JWT Token 做更完整的身份驗證，包括 bcrypt 密碼雜湊、Token 產生、Token 驗證，以及完整的登入流程。
