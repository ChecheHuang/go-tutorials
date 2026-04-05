# 第十一課：Gin 框架入門

## 學習目標

- 了解為什麼需要 Web 框架（對比第 10 課的標準庫）
- 學會 Gin 的路由設定、路由群組
- 掌握 `gin.Context` 的核心 API
- 學會取得路由參數和查詢參數

## 執行方式

```bash
cd tutorials/11-gin-framework
go mod init gin-demo && go mod tidy
go run main.go

# 測試：
curl http://localhost:9090/ping
curl http://localhost:9090/api/v1/users
curl -X POST http://localhost:9090/api/v1/users -H "Content-Type: application/json" -d '{"name":"Carol","age":22}'
```

## 重點筆記

### gin.Context 常用 API

| 方法 | 用途 | 範例 |
|------|------|------|
| `c.JSON(code, obj)` | 回傳 JSON | `c.JSON(200, gin.H{"msg": "ok"})` |
| `c.Param("id")` | 取得路由參數 | `/users/:id` → `c.Param("id")` |
| `c.Query("key")` | 取得查詢參數 | `?page=1` → `c.Query("page")` |
| `c.DefaultQuery("key", "default")` | 查詢參數 + 預設值 | |
| `c.ShouldBindJSON(&obj)` | 解析 JSON body | POST/PUT 的 body |
| `c.ShouldBindQuery(&obj)` | 解析查詢參數到結構體 | |
| `c.GetHeader("key")` | 取得 Header | `c.GetHeader("Authorization")` |
| `c.Set("key", value)` | 存值到 Context | 中介層傳遞資料 |
| `c.Get("key")` | 從 Context 取值 | |
| `c.Abort()` | 中止請求鏈 | 認證失敗時使用 |
| `c.Next()` | 繼續下一個 handler | 中介層中使用 |

### 路由參數 vs 查詢參數

```
路由參數（Path Parameter）：
  GET /api/v1/users/:id    →  /api/v1/users/42
  c.Param("id") = "42"
  用於：識別特定資源

查詢參數（Query Parameter）：
  GET /api/v1/users?page=2&search=alice
  c.Query("page") = "2"
  c.Query("search") = "alice"
  用於：篩選、排序、分頁
```

### 標準庫 vs Gin 對比

```go
// 標準庫
http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":  getUsers(w, r)
    case "POST": createUser(w, r)
    }
})

// Gin（更簡潔，自動路由分發）
r.GET("/users", getUsers)
r.POST("/users", createUser)
```

## 練習

1. 新增一個 `GET /api/v1/health` 端點，回傳 `{"status": "ok"}`
2. 新增 `GET /api/v1/users/search?name=alice` 端點，根據名字搜尋
3. 嘗試用 `gin.H` 建立巢狀的 JSON 回應
