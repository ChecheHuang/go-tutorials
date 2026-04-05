# 第十一課：HTTP 基礎

> **HTTP = 瀏覽器和伺服器之間溝通的語言。**
> 你在瀏覽器輸入網址按下 Enter，到網頁出現在螢幕上 — 中間就是 HTTP 在運作。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟢 初學者 | **入門必修**：HTTP 是所有後端開發的基礎 |
| 🟡 中級工程師 | 理解 Request/Response 生命週期，handler 設計 |

## 學習目標

- 理解 HTTP 是什麼（用生活比喻）
- 理解 HTTP 請求/回應的基本結構
- 認識 HTTP 方法（GET、POST、PUT、DELETE）
- 認識 HTTP 狀態碼（200、404、500 等）
- 了解 JSON 是什麼、為什麼 API 都用它
- 用 Go 標準庫 `net/http` 建立簡單的 HTTP 伺服器
- 為學習 Gin 框架打下基礎

## 執行方式

```bash
go run ./tutorials/11-http-basics

# 然後在另一個終端測試：
curl http://localhost:9090/api/users
curl -X POST http://localhost:9090/api/users -d '{"name":"Dave","age":35}'
```

---

## 用生活來理解 HTTP

### 寄信的比喻

HTTP 就像寄信：

```
你（瀏覽器/客戶端）                     郵局（伺服器）
     │                                    │
     │  📨 寄出一封信（HTTP 請求）          │
     │  ┌─────────────────────────┐       │
     │  │ 寄件地址：/api/users     │       │
     │  │ 目的：我要看資料（GET）   │       │
     │  │ 附加說明：我要 JSON 格式  │       │
     │  │ 信的內容：（空）          │       │
     │  └─────────────────────────┘       │
     │ ──────────────────────────────────> │
     │                                    │  處理中...
     │  📬 收到回信（HTTP 回應）            │
     │  ┌─────────────────────────┐       │
     │  │ 處理結果：200 OK         │       │
     │  │ 附加說明：回傳 JSON 格式  │       │
     │  │ 回信內容：使用者列表      │       │
     │  └─────────────────────────┘       │
     │ <────────────────────────────────── │
     │                                    │
```

### 餐廳點餐的比喻

HTTP 方法就像餐廳裡客人說的話：

| HTTP 方法 | 餐廳比喻 | 實際用途 | CRUD |
|-----------|---------|---------|------|
| `GET` | 「讓我看看菜單」 | 取得資料，不會改變伺服器的資料 | Read（讀取） |
| `POST` | 「我要點一道新菜」 | 建立新資料 | Create（建立） |
| `PUT` | 「幫我換一道菜」 | 更新現有資料 | Update（更新） |
| `DELETE` | 「取消那道菜」 | 刪除現有資料 | Delete（刪除） |

---

## HTTP 請求的結構

一個 HTTP 請求由四個部分組成：

```
┌─────────────────────────────────────────────────┐
│  HTTP 請求（Request）                             │
├─────────────────────────────────────────────────┤
│                                                 │
│  1. 方法 + 路徑（第一行）                         │
│     POST /api/users HTTP/1.1                    │
│     ^^^^  ^^^^^^^^^^                            │
│     方法   路徑                                  │
│                                                 │
│  2. 標頭（Headers）— 附加說明                     │
│     Content-Type: application/json              │
│     Authorization: Bearer eyJhbGciOi...         │
│                                                 │
│  3. 空行（分隔標頭和內容）                         │
│                                                 │
│  4. 主體（Body）— 信的內容（GET 通常沒有）          │
│     {"name": "Dave", "age": 35}                 │
│                                                 │
└─────────────────────────────────────────────────┘
```

## HTTP 回應的結構

```
┌─────────────────────────────────────────────────┐
│  HTTP 回應（Response）                            │
├─────────────────────────────────────────────────┤
│                                                 │
│  1. 狀態行（第一行）                               │
│     HTTP/1.1 201 Created                        │
│              ^^^^^^^^^^^                        │
│              狀態碼 + 說明                        │
│                                                 │
│  2. 標頭（Headers）— 附加說明                     │
│     Content-Type: application/json              │
│                                                 │
│  3. 空行                                         │
│                                                 │
│  4. 主體（Body）— 回傳的資料                       │
│     {"id": 4, "name": "Dave", "age": 35}        │
│                                                 │
└─────────────────────────────────────────────────┘
```

---

## 常見 HTTP 狀態碼

狀態碼是「回信上的處理結果」，告訴你這封信被怎麼處理了：

### 2xx — 成功

| 狀態碼 | 名稱 | 意義 | 使用時機 |
|--------|------|------|---------|
| `200 OK` | 成功 | 「你要的東西在這裡」 | GET/PUT/DELETE 成功 |
| `201 Created` | 建立成功 | 「已經幫你建立好了」 | POST 建立資源成功 |

### 4xx — 客戶端錯誤（你的信有問題）

| 狀態碼 | 名稱 | 意義 | 使用時機 |
|--------|------|------|---------|
| `400 Bad Request` | 請求錯誤 | 「你的信我看不懂」 | 參數格式錯誤、JSON 無效 |
| `401 Unauthorized` | 未認證 | 「你沒有出示身分證」 | 缺少或無效的 Token |
| `403 Forbidden` | 無權限 | 「你有身分證但不能進」 | 有 Token 但權限不足 |
| `404 Not Found` | 找不到 | 「你要的東西不存在」 | 資源不存在 |
| `405 Method Not Allowed` | 方法不允許 | 「這個窗口不辦這業務」 | 用了不支援的 HTTP 方法 |

### 5xx — 伺服器錯誤（郵局的問題）

| 狀態碼 | 名稱 | 意義 | 使用時機 |
|--------|------|------|---------|
| `500 Internal Server Error` | 伺服器錯誤 | 「郵局內部出問題了」 | 伺服器程式出錯 |

**記憶技巧：**
- **2xx** = 一切順利（200 多分 = 成功）
- **4xx** = 是你的問題（404 = 你走錯路了）
- **5xx** = 是伺服器的問題（500 = 不是你的錯）

---

## JSON 是什麼？

**JSON**（JavaScript Object Notation）是一種「資料格式」，用來在不同系統之間交換資料。

### JSON 長什麼樣子？

```json
{
    "name": "Alice",
    "age": 25,
    "isStudent": true,
    "hobbies": ["reading", "coding"],
    "address": {
        "city": "Taipei",
        "country": "Taiwan"
    }
}
```

### 為什麼 API 都用 JSON？

| 原因 | 說明 |
|------|------|
| 人類讀得懂 | 比起二進位的資料格式（如 Protobuf），JSON 是純文字 |
| 所有語言都支援 | Go、Python、JavaScript、Java... 都能解析 JSON |
| 結構簡單 | 只有物件 `{}`、陣列 `[]`、字串、數字、布林值、null |
| Web 標準 | 幾乎所有現代 Web API 都用 JSON |

### Go 如何處理 JSON？

Go 使用 `encoding/json` 套件來處理 JSON：

```go
// Go struct → JSON（編碼）
user := User{ID: 1, Name: "Alice", Age: 25}
jsonBytes, _ := json.Marshal(user)
// 結果：{"id":1,"name":"Alice","age":25}

// JSON → Go struct（解碼）
var user User
json.Unmarshal(jsonBytes, &user)
// user.Name == "Alice"
```

**struct tag**（結構體標籤）控制 JSON 的欄位名稱：
```go
type User struct {
    ID   int    `json:"id"`    // JSON 中的 key 是 "id"
    Name string `json:"name"`  // JSON 中的 key 是 "name"
}
```

---

## Go 的 net/http 套件

Go 標準庫自帶 HTTP 功能，不需要安裝任何第三方套件。

### 核心概念

```
net/http 的三個核心：

1. Handler（處理器）= 處理請求的函式
   func(w http.ResponseWriter, r *http.Request)
   - w：回應的「信紙」，往上面寫東西
   - r：收到的「信」，從裡面讀資訊

2. ServeMux（路由器）= 郵局的分信系統
   將不同的 URL 路徑對應到不同的 Handler
   "/" → helloHandler
   "/api/users" → usersHandler

3. Server（伺服器）= 郵局本身
   http.ListenAndServe(":9090", nil)
   在 9090 埠持續等待並處理請求
```

### 標準庫 vs Gin 框架

標準庫能做一切，但 Gin 讓程式碼更簡潔：

```go
// 標準庫：手動判斷方法、手動解析 JSON
func handler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET": ...
    case "POST": ...
    }
}

// Gin：路由自動分發、內建 JSON 綁定
router.GET("/users", getUsers)
router.POST("/users", createUsers)
```

Gin 還提供：
- 更好的路由匹配（支援路徑參數如 `/users/:id`）
- 內建 JSON 綁定和驗證
- 中介軟體（Middleware）支援
- 更好的錯誤處理
- 效能更高

---

## 與部落格專案的關係

這一課學的 HTTP 基礎，是部落格專案 Handler 層的基石：

| 本課概念 | 部落格專案對應 |
|---------|--------------|
| `http.HandleFunc` 路由 | Gin 的 `router.GET`、`router.POST` |
| `r.Method` 判斷方法 | Gin 自動根據方法分發 |
| `json.NewEncoder(w).Encode(data)` | `c.JSON(200, data)` |
| `json.NewDecoder(r.Body).Decode(&req)` | `c.ShouldBindJSON(&req)` |
| `w.WriteHeader(201)` | 由 response helper 處理 |
| `r.URL.Query().Get("key")` | `c.Query("key")` |

**理解標準庫的 HTTP，就能理解 Gin 在背後幫你做了什麼。**

---

## 常見問題

### Q: 什麼是 localhost:9090？

- `localhost` = 本機（就是你自己的電腦）
- `9090` = 埠號（port），就像公寓的門牌號碼
- 你的電腦可以同時開很多服務，靠不同的埠號區分

### Q: 為什麼 GET 請求沒有 Body？

GET 的語意是「取得資料」，所以只需要告訴伺服器「我要什麼」（透過 URL 和查詢參數），不需要附帶資料。POST 的語意是「建立資料」，所以需要附帶要建立的資料（放在 Body 裡）。

### Q: 為什麼要設定 Content-Type？

就像信封上標明「裡面是中文信」或「裡面是英文信」，Content-Type 告訴接收方「這份資料的格式是什麼」。`application/json` 表示內容是 JSON 格式。

### Q: curl 是什麼？

curl 是一個命令列工具，可以用來發送 HTTP 請求。就像一個「虛擬的瀏覽器」，但在終端機裡操作。

```bash
# GET 請求（預設就是 GET）
curl http://localhost:9090/api/users

# POST 請求（-X 指定方法，-d 指定 body）
curl -X POST http://localhost:9090/api/users -d '{"name":"Dave","age":35}'
```

### Q: 標準庫夠用嗎？為什麼還要學 Gin？

標準庫完全可以建一個完整的 API，但 Gin 幫你省去很多重複的工作（路由分發、JSON 處理、中介軟體等）。就像你可以用記事本寫程式，但 VS Code 會讓你更有效率。

---

## 練習

1. 新增一個 `DELETE /api/users?id=1` 端點，刪除指定 ID 的使用者
2. 在回應中加上自訂的 Header `X-Server: MyGoServer`
3. 觀察：直接用瀏覽器存取 `/api/users` 時，瀏覽器發送的是什麼方法？
4. 試著用 `curl -v http://localhost:9090/api/users` 查看完整的 HTTP 請求和回應（`-v` = verbose，顯示詳細資訊）
5. 故意發送錯誤的 JSON（例如 `curl -X POST /api/users -d 'not json'`），觀察回應

---

## 下一課預告：Gin 框架

下一課我們將學習 **Gin** — Go 最受歡迎的 HTTP 框架。你會發現 Gin 把這一課手動做的事情都自動化了：

```go
// 這一課（標準庫）：手動判斷方法、手動解析 JSON、手動設定狀態碼
func usersHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(users)
    case "POST":
        var user User
        json.NewDecoder(r.Body).Decode(&user)
        w.WriteHeader(201)
        json.NewEncoder(w).Encode(user)
    }
}

// 下一課（Gin）：一切都變得簡潔
router.GET("/api/users", func(c *gin.Context) {
    c.JSON(200, users)
})
router.POST("/api/users", func(c *gin.Context) {
    var user User
    c.ShouldBindJSON(&user)
    c.JSON(201, user)
})
```

有了這一課的基礎，你將完全理解 Gin 在幕後做了什麼！
