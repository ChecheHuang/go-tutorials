# 第十二課：JSON 處理與結構標籤

## 學習目標

- 理解結構標籤（Struct Tags）的語法和用途
- 學會 JSON 序列化（Marshal）與反序列化（Unmarshal）
- 掌握 `json`、`binding`、`gorm` 三種標籤
- 了解 `omitempty` 和 `"-"` 的作用

## 執行方式

```bash
cd tutorials/12-json-binding
go run main.go
```

## 重點筆記

### 三種標籤的分工

```go
type User struct {
    ID       uint   `json:"id"       gorm:"primaryKey"          binding:"-"`
    Username string `json:"username" gorm:"uniqueIndex;size:50" binding:"required,min=3"`
    Password string `json:"-"        gorm:"not null"            binding:"required,min=6"`
}
```

| 標籤 | 負責 | 由誰讀取 |
|------|------|---------|
| `json` | JSON 序列化/反序列化 | `encoding/json` 套件 |
| `binding` | 請求參數驗證 | Gin 框架 |
| `gorm` | 資料庫欄位定義 | GORM 框架 |

### `json:"-"` 是安全的關鍵

```go
Password string `json:"-"`
```

這行確保密碼永遠不會出現在 API 回應的 JSON 中，即使你不小心把整個 User 物件回傳。

### 在專案中的對應

`internal/domain/user.go`：
```go
type User struct {
    ID        uint   `json:"id"       gorm:"primaryKey"`
    Username  string `json:"username" gorm:"uniqueIndex;size:50;not null"`
    Email     string `json:"email"    gorm:"uniqueIndex;size:100;not null"`
    Password  string `json:"-"        gorm:"not null"`
}
```

一個欄位同時被三個系統使用：JSON 輸出、資料庫定義、輸入驗證。

## 練習

1. 定義一個 `Product` 結構體，包含 name、price、secret_code，確保 secret_code 不會出現在 JSON
2. 用 `json.Marshal` 和 `json.Unmarshal` 互相轉換
3. 嘗試不同的 `omitempty` 組合，觀察輸出差異
