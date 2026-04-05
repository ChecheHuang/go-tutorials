# 第十四課：中介層（Middleware）

## 學習目標

- 理解中介層的「洋蔥模型」
- 學會寫自訂中介層（Logger、Auth、CORS、Recovery）
- 分辨 `c.Next()` 和 `c.Abort()` 的差異
- 了解全域中介層 vs 群組中介層

## 執行方式

```bash
cd tutorials/14-middleware
go mod init middleware-demo && go mod tidy
go run main.go
```

## 重點筆記

### 洋蔥模型

```
請求進來 ─→ Logger ─→ Auth ─→ Handler
                                  │
回應出去 ←─ Logger ←─ Auth ←─────┘

Logger 中的 c.Next() 前 = 請求進來時
Logger 中的 c.Next() 後 = 回應出去時
```

### c.Next() vs c.Abort()

```go
c.Next()   // 繼續執行下一個 handler → 最終到達目標 Handler
c.Abort()  // 中止！後面的 handler 都不會執行
```

### 中介層的應用範圍

```go
// 全域：對所有路由生效
r.Use(LoggerMiddleware())

// 群組：只對特定路由群組生效
api := r.Group("/api")
api.Use(AuthMiddleware())

// 單一路由：只對一個路由生效
r.GET("/special", SpecialMiddleware(), handler)
```

### 在專案中的對應

`internal/handler/router.go`：
```go
r.Use(middleware.Logger())    // 全域：記錄所有請求
r.Use(middleware.Recovery())  // 全域：攔截 panic
r.Use(middleware.CORS())      // 全域：跨域設定

authenticated := v1.Group("")
authenticated.Use(middleware.JWTAuth(cfg))  // 群組：只保護需要認證的路由
```

## 練習

1. 寫一個 `RequestID` 中介層，為每個請求產生唯一 ID 並加到 Header
2. 寫一個 `AdminOnly` 中介層，檢查 Header 中的 `X-Role` 是否為 `admin`
3. 觀察 panic 路由的行為：Recovery 中介層如何防止伺服器崩潰？
