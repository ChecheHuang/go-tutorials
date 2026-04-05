# 第十二課：Gin 框架入門

> **一句話總結：** Gin 是 Go 最受歡迎的 Web 框架，幫你快速建立 RESTful API。

## 你會學到什麼？

- 什麼是 Gin？什麼是第三方套件？
- 為什麼用 Gin 而不是標準庫 `net/http`？
- `gin.Context` — Gin 的核心（瑞士刀）
- 路由註冊（GET、POST、PUT、DELETE）
- 路由參數（`:id`）vs 查詢參數（`?keyword=alice`）
- JSON 綁定（`c.ShouldBindJSON`）
- 路由群組（Route Groups）
- 統一錯誤回應模式

## 執行方式

```bash
go run ./tutorials/12-gin-framework

# 伺服器啟動後，開另一個終端機測試：
curl http://localhost:9090/ping
curl http://localhost:9090/api/v1/users
curl http://localhost:9090/api/v1/users/1
curl -X POST http://localhost:9090/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Carol","age":22}'
curl http://localhost:9090/api/v1/search?keyword=alice&page=1

# 按 Ctrl+C 停止伺服器
```

---

## 什麼是 Gin？

Gin 是一個 **第三方 Web 框架**（不是 Go 內建的），由社群開發並維護。

### 什麼是第三方套件？

Go 有兩種套件：

| 類型 | 範例 | 取得方式 |
|------|------|---------|
| **標準庫**（Go 內建） | `fmt`、`net/http`、`encoding/json` | 安裝 Go 就有 |
| **第三方套件**（社群開發） | `github.com/gin-gonic/gin` | 用 `go get` 安裝 |

### 第三方套件的 import path

```go
import "github.com/gin-gonic/gin"
//      ^^^^^^^^^^^^^^^^^^^^^^^^
//      這就是 Gin 在 GitHub 上的網址路徑
//      Go 用這個網址作為套件的「唯一識別」
```

**注意：** 這不代表每次編譯都要連網。`go mod` 會在本地快取套件。

### 安裝第三方套件

```bash
# 方法 1：直接安裝
go get github.com/gin-gonic/gin

# 方法 2：寫好 import 後執行 go mod tidy（自動下載缺少的套件）
go mod tidy
```

安裝後，`go.mod` 會自動記錄套件版本，`go.sum` 會記錄校驗碼。

---

## 為什麼用 Gin？net/http vs Gin 對比

### 路由註冊

```go
// === 標準庫 net/http（第 11 課）===
// 一個路徑要自己判斷 HTTP Method
http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":    getUsers(w, r)
    case "POST":   createUser(w, r)
    default:       http.Error(w, "不支援的方法", 405)
    }
})

// === Gin 框架 ===
// 每個 Method 直接對應一個函式，清楚明瞭
r.GET("/users", getUsers)
r.POST("/users", createUser)
```

### 完整對比表

| 功能 | 標準庫 `net/http` | Gin 框架 |
|------|-------------------|---------|
| 路由註冊 | `http.HandleFunc`，要手動判斷 Method | `r.GET()`、`r.POST()` 等，自動分發 |
| 路由參數 | 不支援，要手動解析 URL | `/users/:id` → `c.Param("id")` |
| 查詢參數 | `r.URL.Query().Get("key")` | `c.Query("key")` |
| 讀取 JSON Body | 手動 `json.NewDecoder(r.Body).Decode()` | `c.ShouldBindJSON(&obj)` 一行搞定 |
| 回傳 JSON | 手動設定 Header + `json.Encoder` | `c.JSON(200, data)` 一行搞定 |
| 參數驗證 | 完全手動 | `binding:"required,min=3"` 自動驗證 |
| 路由群組 | 不支援 | `r.Group("/api/v1")` |
| 中介層 | 要自己實作 | `r.Use(middleware)` |

### 總結：什麼時候用什麼？

- **學習 HTTP 原理** → 用 `net/http`（第 11 課）
- **開發真實 API** → 用 Gin（本課開始）

---

## gin.Context — Gin 的「瑞士刀」

`gin.Context` 是 Gin 最重要的結構體。每當一個 HTTP 請求進來，Gin 會建立一個 `*gin.Context` 傳給你的 handler 函式。幾乎所有操作都透過它完成。

### 讀取請求

| 方法 | 用途 | 範例 |
|------|------|------|
| `c.Param("id")` | 取得路由參數 `:id` | `/users/:id` → `c.Param("id")` |
| `c.Query("key")` | 取得查詢參數 | `?page=1` → `c.Query("page")` |
| `c.DefaultQuery("key", "default")` | 查詢參數 + 預設值 | 沒給 page 就用 `"1"` |
| `c.ShouldBindJSON(&obj)` | 解析 Request Body 的 JSON 到結構體 | POST/PUT 的請求資料 |
| `c.ShouldBindQuery(&obj)` | 解析查詢參數到結構體 | 多個查詢參數一次綁定 |
| `c.GetHeader("key")` | 取得 HTTP Header | `c.GetHeader("Authorization")` |

### 寫入回應

| 方法 | 用途 | 範例 |
|------|------|------|
| `c.JSON(code, obj)` | 回傳 JSON | `c.JSON(200, gin.H{"msg": "ok"})` |
| `c.String(code, str)` | 回傳純文字 | `c.String(200, "Hello")` |
| `c.HTML(code, name, obj)` | 回傳 HTML 模板 | 搭配模板引擎使用 |
| `c.Status(code)` | 只設定狀態碼，不回傳 body | `c.Status(204)` |

### 流程控制

| 方法 | 用途 | 使用場景 |
|------|------|---------|
| `c.Next()` | 繼續執行下一個 handler | 中介層中使用，處理完前置邏輯後繼續 |
| `c.Abort()` | 中止請求鏈 | 認證失敗時，阻止請求繼續到 handler |
| `c.AbortWithStatusJSON(code, obj)` | 中止並回傳 JSON 錯誤 | 認證失敗回傳 401 |

### 資料傳遞（中介層 ↔ Handler）

| 方法 | 用途 | 使用場景 |
|------|------|---------|
| `c.Set("key", value)` | 在 Context 中存值 | 中介層解析 JWT 後存入 userID |
| `c.Get("key")` | 從 Context 取值 | Handler 中取出 userID |
| `c.MustGet("key")` | 取值，不存在就 panic | 確定值一定存在時使用 |

---

## 路由註冊模式

### 四種 HTTP 方法 = CRUD 四種操作

```go
r.GET("/users", getUsers)        // Read（讀取）— 取得使用者列表
r.POST("/users", createUser)     // Create（建立）— 建立新使用者
r.PUT("/users/:id", updateUser)  // Update（更新）— 更新指定使用者
r.DELETE("/users/:id", deleteUser) // Delete（刪除）— 刪除指定使用者
```

### RESTful API 設計慣例

| 路由 | 說明 |
|------|------|
| `GET /articles` | 取得文章列表 |
| `GET /articles/:id` | 取得某篇文章 |
| `POST /articles` | 建立新文章 |
| `PUT /articles/:id` | 更新某篇文章 |
| `DELETE /articles/:id` | 刪除某篇文章 |

名詞用**複數**（articles 不是 article），動詞用 HTTP Method 表達。

---

## 路由參數 vs 查詢參數

```
路由參數（Path Parameter）— 識別「哪一個」資源
  定義：/users/:id
  請求：GET /users/42
  取值：c.Param("id") → "42"
  場景：取得、更新、刪除「特定」的資源

查詢參數（Query Parameter）— 「怎麼」取得資源
  請求：GET /users?page=2&search=alice&sort=name
  取值：c.Query("page") → "2"
        c.Query("search") → "alice"
        c.DefaultQuery("sort", "id") → "name"
  場景：篩選、搜尋、排序、分頁
```

**記憶口訣：** 路由參數 = 「你要找誰」，查詢參數 = 「你要怎麼找」。

---

## JSON 綁定（ShouldBindJSON）

`c.ShouldBindJSON` 是 Gin 最強大的功能之一，它同時做兩件事：

1. **解析 JSON** — 把 Request Body 的 JSON 轉成 Go 結構體
2. **驗證資料** — 根據 `binding` 標籤檢查資料是否合法

```go
// 定義請求結構體，binding 標籤控制驗證規則
type CreateUserRequest struct {
    Name string `json:"name" binding:"required"`       // 必填
    Age  int    `json:"age"  binding:"required,gt=0"`  // 必填，且大於 0
}

// Handler 中使用
func createUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // 驗證失敗 → 回傳 400 錯誤
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    // 驗證通過 → 處理業務邏輯
}
```

> `ShouldBindJSON` vs `BindJSON`：
> `ShouldBindJSON` 只回傳 error，讓你自己決定怎麼處理。
> `BindJSON` 會自動回傳 400 錯誤，你無法自訂錯誤格式。
> **建議用 `ShouldBindJSON`**，更靈活。

---

## 路由群組（Route Groups）

路由群組讓你把相關的路由整理在一起，共用路徑前綴和中介層。

```go
// 建立 /api/v1 群組
v1 := r.Group("/api/v1")
{
    v1.GET("/users", getUsers)       // 實際路徑：/api/v1/users
    v1.POST("/users", createUser)    // 實際路徑：/api/v1/users
}

// 群組可以巢狀
auth := v1.Group("/auth")
{
    auth.POST("/register", register) // 實際路徑：/api/v1/auth/register
    auth.POST("/login", login)       // 實際路徑：/api/v1/auth/login
}

// 群組可以套用中介層（下一課會詳細教）
protected := v1.Group("")
protected.Use(authMiddleware)        // 這個群組裡的所有路由都需要認證
{
    protected.POST("/articles", create)  // 需要登入才能建立文章
}
```

---

## 統一錯誤回應模式

在真實專案中，所有 API 的錯誤回應應該用**相同的格式**，方便前端處理。

```go
// 錯誤回應格式
{"error": "使用者不存在"}

// 成功回應格式
{"data": {"id": 1, "name": "Alice"}}
```

```go
// 封裝成函式，所有 handler 都用這個
func respondWithError(c *gin.Context, code int, message string) {
    c.JSON(code, gin.H{"error": message})
}

func respondWithData(c *gin.Context, code int, data interface{}) {
    c.JSON(code, gin.H{"data": data})
}
```

這個模式在我們的部落格專案中也有使用。

---

## 部落格專案對照

本課學到的概念，在部落格專案的 `internal/handler/router.go` 中都有實際應用：

```go
// 路由群組 — 版本化 API
v1 := r.Group("/api/v1")

// 巢狀群組 — 認證路由
auth := v1.Group("/auth")
{
    auth.POST("/register", userHandler.Register)  // 註冊
    auth.POST("/login", userHandler.Login)         // 登入
}

// 群組 + 中介層 — 需要登入的路由
authenticated := v1.Group("")
authenticated.Use(middleware.JWTAuth(cfg))  // 套用 JWT 認證中介層
{
    authenticated.POST("/articles", articleHandler.Create)        // 建立文章
    authenticated.PUT("/articles/:id", articleHandler.Update)     // 路由參數 :id
    authenticated.DELETE("/articles/:id", articleHandler.Delete)  // 路由參數 :id
}

// 公開路由 — 不需要登入
v1.GET("/articles", articleHandler.GetAll)       // 取得文章列表
v1.GET("/articles/:id", articleHandler.GetByID)  // 取得文章詳情
```

---

## 常見問題（FAQ）

### Q: `gin.H` 是什麼？

`gin.H` 就是 `map[string]any` 的別名（type alias）。它讓你快速建立 JSON 回應：

```go
// 這兩行是完全等價的
c.JSON(200, gin.H{"message": "ok"})
c.JSON(200, map[string]any{"message": "ok"})
```

### Q: 為什麼要 `gin.SetMode(gin.ReleaseMode)`？

Gin 有三種模式：
- `gin.DebugMode`（預設）— 會印出路由資訊和除錯日誌
- `gin.ReleaseMode` — 關閉除錯日誌，適合正式環境
- `gin.TestMode` — 測試時使用

在部落格專案中用 `ReleaseMode` 是因為我們自己實作了更好看的路由日誌。

### Q: `c.Abort()` 是什麼？什麼時候用？

`c.Abort()` 會中止目前的請求處理鏈。常見用法是在**中介層**中，當認證失敗時阻止請求繼續到 handler：

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "未登入"})
            return  // 請求到此結束，不會繼續到 handler
        }
        c.Next()  // 認證通過，繼續執行下一個 handler
    }
}
```

### Q: `gin.Default()` 和 `gin.New()` 有什麼不同？

```go
r := gin.Default() // 自帶 Logger + Recovery 中介層
r := gin.New()     // 空白引擎，沒有任何中介層
```

部落格專案用 `gin.New()` 是因為我們用自訂的中介層取代了預設的。

### Q: `ShouldBind` 系列有幾種？

| 方法 | 資料來源 | 使用場景 |
|------|---------|---------|
| `c.ShouldBindJSON(&obj)` | Request Body (JSON) | POST/PUT 的 JSON 資料 |
| `c.ShouldBindQuery(&obj)` | URL 查詢參數 | `?page=1&search=foo` |
| `c.ShouldBindUri(&obj)` | 路由參數 | `/users/:id` |
| `c.ShouldBind(&obj)` | 自動偵測 | 根據 Content-Type 決定 |

---

## 練習

1. **新增健康檢查端點：** 新增 `GET /api/v1/health`，回傳 `{"status": "ok", "version": "1.0.0"}`
2. **新增名字搜尋：** 新增 `GET /api/v1/users/search?name=alice`，根據名字模糊搜尋使用者
3. **巢狀 JSON 回應：** 修改 `getUsers`，回傳包含 `total`（總數）和 `items`（使用者列表）的 JSON
4. **新增路由群組：** 建立一個 `/api/v2` 群組，裡面只放一個測試端點
5. **統一錯誤格式：** 在錯誤回應中加入 `code` 欄位，例如 `{"error": "使用者不存在", "code": "USER_NOT_FOUND"}`

---

## 下一課預告

**第十三課：JSON 處理與結構標籤** — 深入了解 `json`、`binding`、`gorm`、`example` 標籤，學會控制 JSON 的序列化和反序列化行為，這些標籤是建立 API 的基礎。
