# 第十課：HTTP 基礎

## 學習目標

- 理解 HTTP 請求/回應的基本結構
- 用 Go 標準庫 `net/http` 建立簡單的 HTTP 伺服器
- 了解 Handler 函式的運作方式
- 為學習 Gin 框架打下基礎

## 執行方式

```bash
cd tutorials/10-http-basics
go run main.go

# 然後在另一個終端測試：
curl http://localhost:9090/api/users
curl -X POST http://localhost:9090/api/users -d '{"name":"Dave","age":35}'
```

## 重點筆記

### HTTP 方法與 CRUD 的對應

| HTTP 方法 | CRUD 操作 | 範例 |
|-----------|----------|------|
| `GET` | Read（讀取） | 取得使用者列表 |
| `POST` | Create（建立） | 建立新使用者 |
| `PUT` | Update（更新） | 更新使用者資料 |
| `DELETE` | Delete（刪除） | 刪除使用者 |

### 常見 HTTP 狀態碼

| 狀態碼 | 意義 | 使用時機 |
|--------|------|---------|
| `200 OK` | 成功 | GET/PUT/DELETE 成功 |
| `201 Created` | 建立成功 | POST 建立資源成功 |
| `400 Bad Request` | 請求錯誤 | 參數格式錯誤 |
| `401 Unauthorized` | 未認證 | 缺少或無效的 Token |
| `403 Forbidden` | 無權限 | 有 Token 但權限不足 |
| `404 Not Found` | 找不到 | 資源不存在 |
| `500 Internal Server Error` | 伺服器錯誤 | 伺服器內部出錯 |

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

## 練習

1. 新增一個 `DELETE /api/users?id=1` 端點
2. 在回應中加上自訂的 Header `X-Server: MyGoServer`
3. 觀察：直接用瀏覽器存取 `/api/users` 時，瀏覽器發送的是什麼方法？
