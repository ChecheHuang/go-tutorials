# 第十三課：JSON 處理與結構標籤（Struct Tags）

> **一句話總結：** 結構標籤就像「貼在箱子上的標籤」，告訴不同的系統（JSON、資料庫、驗證器）如何處理每個欄位。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟢 初學者 | API 開發必備，JSON 序列化/反序列化 |
| 🟡 中級工程師 | struct tag、自訂驗證、omitempty 的使用 |

## 你會學到什麼？

- 什麼是結構標籤？為什麼需要它們？
- `json` 標籤 — 控制 JSON 的欄位名稱和行為
- `binding` 標籤 — Gin 框架的自動驗證
- `gorm` 標籤 — GORM 框架的資料庫定義
- `example` 標籤 — Swagger API 文件的範例值
- 一個欄位上的多個標籤如何協同工作
- `Marshal`（序列化）和 `Unmarshal`（反序列化）的實際操作

## 執行方式

```bash
go run ./tutorials/13-json-binding
```

---

## 什麼是結構標籤（Struct Tags）？

結構標籤是用反引號（`` ` ``）包住的後設資料（metadata），附加在結構體的欄位上。

**比喻：** 想像你有一個箱子（結構體欄位），上面可以貼多張標籤：
- **紅色標籤**（json）→ 告訴搬運工（JSON 系統）：「這個箱子叫什麼名字」
- **藍色標籤**（binding）→ 告訴檢查員（Gin 框架）：「這個箱子裡的東西要符合什麼條件」
- **綠色標籤**（gorm）→ 告訴倉庫管理員（GORM 框架）：「這個箱子要放在哪個貨架」
- **黃色標籤**（example）→ 告訴文件編輯（Swagger）：「這個箱子裡通常放什麼」

### 語法

```go
type User struct {
    欄位名 型別 `標籤名:"值" 標籤名:"值"`
}

// 實際範例
type User struct {
    ID       uint   `json:"id"       gorm:"primaryKey"`
    Username string `json:"username" gorm:"uniqueIndex;size:50" binding:"required,min=3"`
    Password string `json:"-"        gorm:"not null"            binding:"required,min=6"`
}
```

**注意事項：**
- 所有標籤都用一對反引號 `` ` `` 包住
- 標籤之間用**空格**分隔
- 每個標籤內部用**逗號**分隔選項（如 `binding:"required,min=3"`）
- gorm 標籤內部用**分號**分隔選項（如 `gorm:"size:200;not null"`）

---

## json 標籤 — 控制 JSON 序列化

`json` 標籤由 Go 標準庫的 `encoding/json` 套件讀取，控制結構體和 JSON 之間的轉換。

### 完整選項表

| 標籤寫法 | 意思 | 範例 |
|---------|------|------|
| `json:"name"` | JSON 欄位名稱是 `name` | `Username string \`json:"name"\`` → `{"name": "alice"}` |
| `json:"name,omitempty"` | 零值時省略此欄位 | 空字串 → 不出現在 JSON |
| `json:"-"` | 永遠不出現在 JSON 中 | 密碼欄位必用 |
| `json:",omitempty"` | 保持原始欄位名，但零值時省略 | 較少使用 |
| `json:"-,"` | JSON 欄位名稱就叫 `-`（很少用） | 特殊情況 |

### 零值定義（omitempty 何時生效）

| 型別 | 零值 |
|------|------|
| `int`、`float64` 等數字 | `0` |
| `string` | `""` （空字串） |
| `bool` | `false` |
| 指標（`*Type`） | `nil` |
| 切片（`[]Type`） | `nil`（注意：空切片 `[]` 不是 nil） |
| map | `nil` |

### json:"-" 是安全的關鍵

```go
type User struct {
    Password string `json:"-"` // 密碼永遠不會出現在 JSON 回應中
}
```

即使你不小心寫了 `c.JSON(200, user)`，密碼也不會洩漏。這是 API 開發的**必備安全措施**。

---

## binding 標籤 — Gin 的自動驗證

`binding` 標籤由 Gin 框架讀取（底層使用 `go-playground/validator`），在 `c.ShouldBindJSON()` 時自動驗證。

### 驗證規則表

| 規則 | 用途 | 範例 |
|------|------|------|
| `required` | 必填（不可為零值） | `binding:"required"` |
| `min=N` | 字串最短 N 字元 / 數字最小值 N | `binding:"min=3"` |
| `max=N` | 字串最長 N 字元 / 數字最大值 N | `binding:"max=50"` |
| `len=N` | 長度必須剛好是 N | `binding:"len=10"` |
| `email` | 合法 Email 格式 | `binding:"email"` |
| `url` | 合法 URL 格式 | `binding:"url"` |
| `omitempty` | 空值時跳過驗證（選填） | `binding:"omitempty,min=3"` |
| `oneof=a b c` | 值必須是列舉值之一 | `binding:"oneof=admin user guest"` |
| `gt=N` | 大於 N | `binding:"gt=0"` |
| `gte=N` | 大於等於 N | `binding:"gte=0"` |
| `lt=N` | 小於 N | `binding:"lt=100"` |
| `lte=N` | 小於等於 N | `binding:"lte=100"` |
| `-` | 忽略此欄位 | `binding:"-"` |

### 組合使用

```go
// 必填 + 長度限制
Username string `binding:"required,min=3,max=50"`

// 必填 + Email 格式
Email string `binding:"required,email"`

// 選填，但如果有值就要驗證
Age int `binding:"omitempty,min=0,max=150"`
```

---

## gorm 標籤 — 資料庫欄位定義

`gorm` 標籤由 GORM 框架讀取，用來定義資料庫表格的結構。

### 選項表

| 選項 | 用途 | 範例 |
|------|------|------|
| `primaryKey` | 主鍵 | `gorm:"primaryKey"` |
| `autoIncrement` | 自動遞增 | `gorm:"autoIncrement"` |
| `uniqueIndex` | 唯一索引（值不可重複） | `gorm:"uniqueIndex"` |
| `index` | 一般索引（加速查詢） | `gorm:"index"` |
| `not null` | 不可為空 | `gorm:"not null"` |
| `size:N` | 欄位長度 | `gorm:"size:200"` → VARCHAR(200) |
| `type:TYPE` | 指定資料庫型別 | `gorm:"type:text"` |
| `default:VALUE` | 預設值 | `gorm:"default:0"` |
| `column:NAME` | 自訂欄位名稱 | `gorm:"column:user_name"` |
| `foreignKey:FIELD` | 外鍵 | `gorm:"foreignKey:UserID"` |
| `-` | 忽略此欄位（不建表） | `gorm:"-"` |

### 多個選項用分號分隔

```go
Title string `gorm:"size:200;not null"`     // VARCHAR(200) NOT NULL
Email string `gorm:"uniqueIndex;size:100"`   // VARCHAR(100) UNIQUE INDEX
```

---

## example 標籤 — Swagger API 文件

`example` 標籤由 Swagger 文件產生工具（如 `swag`）讀取，用來在 API 文件中顯示範例值。

```go
type RegisterRequest struct {
    Username string `json:"username" example:"newuser"`
    Email    string `json:"email"    example:"newuser@example.com"`
    Password string `json:"password" example:"password123"`
}
```

**注意：** `example` 標籤不會影響程式的執行邏輯，它只用於文件產生。

---

## 多個標籤如何協同工作

一個欄位可以同時有多個標籤，各自由不同的套件讀取：

```go
type User struct {
    Username string `json:"username" gorm:"uniqueIndex;size:50;not null" binding:"required,min=3" example:"alice"`
}
```

| 標籤 | 由誰讀取 | 做什麼 |
|------|---------|--------|
| `json:"username"` | `encoding/json` 套件 | JSON 欄位名稱叫 "username" |
| `gorm:"uniqueIndex;size:50;not null"` | GORM 框架 | 資料庫建唯一索引、最多 50 字元、不可為空 |
| `binding:"required,min=3"` | Gin 框架 | API 請求時必填，最少 3 字元 |
| `example:"alice"` | Swagger 工具 | 文件中顯示 "alice" 作為範例 |

**每個套件只讀自己認識的標籤，其他標籤會被忽略。**

---

## 部落格專案實戰解析

### User 結構體（`internal/domain/user.go`）

```go
type User struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    //                   ^^^^^^^^   ^^^^^^^^^^^^^^^^^
    //                   JSON 叫 id  資料庫主鍵

    Username  string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
    //                   ^^^^^^^^^^^^^^   ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
    //                   JSON 叫 username  唯一索引 + 50字元 + 不可空

    Email     string    `json:"email" gorm:"uniqueIndex;size:100;not null"`
    //                   ^^^^^^^^^^^^  ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
    //                   JSON 叫 email  唯一索引（Email 不可重複）

    Password  string    `json:"-" gorm:"not null"`
    //                   ^^^^^^^^  ^^^^^^^^^^^^^^^^
    //                   JSON 隱藏  資料庫不可空（一定要有密碼）

    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    // GORM 會自動管理 CreatedAt 和 UpdatedAt
}
```

### RegisterRequest 結構體（`internal/domain/user.go`）

```go
type RegisterRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50"  example:"newuser"`
    //               ^^^^^^^^^^^^^^  ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^   ^^^^^^^^^^^^^^^^^
    //               JSON 欄位名      必填+3~50字元                     Swagger 範例

    Email    string `json:"email"    binding:"required,email"         example:"newuser@example.com"`
    //               ^^^^^^^^^^^^    ^^^^^^^^^^^^^^^^^^^^^^           ^^^^^^^^^^^^^^^^^^^^^^^^^^
    //               JSON 欄位名      必填+Email格式                     Swagger 範例

    Password string `json:"password" binding:"required,min=6,max=100" example:"password123"`
    //               ^^^^^^^^^^^^^^  ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^  ^^^^^^^^^^^^^^^^^^^^^
    //               JSON 欄位名      必填+6~100字元                      Swagger 範例
}
```

### Article 結構體（`internal/domain/article.go`）

```go
type Article struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Title     string    `json:"title" gorm:"size:200;not null"`
    Content   string    `json:"content" gorm:"type:text;not null"`
    UserID    uint      `json:"user_id" gorm:"index;not null"`
    //                                   ^^^^^^^^
    //                                   建立索引，加速「查某個使用者的所有文章」

    User      User      `json:"user" gorm:"foreignKey:UserID"`
    //                                ^^^^^^^^^^^^^^^^^^^^^^
    //                                外鍵關聯：Article.UserID → User.ID

    Comments  []Comment `json:"comments,omitempty" gorm:"foreignKey:ArticleID"`
    //                   ^^^^^^^^^^^^^^^^^^^^^       ^^^^^^^^^^^^^^^^^^^^^^^^^^
    //                   沒有留言時不顯示              外鍵關聯
}
```

### CreateArticleRequest（`internal/domain/article.go`）

```go
type CreateArticleRequest struct {
    Title   string `json:"title"   binding:"required,min=1,max=200" example:"我的第一篇文章"`
    Content string `json:"content" binding:"required,min=1"         example:"這是文章的內容..."`
}
```

### UpdateArticleRequest — 選填欄位的技巧

```go
type UpdateArticleRequest struct {
    Title   string `json:"title"   binding:"omitempty,min=1,max=200"`
    Content string `json:"content" binding:"omitempty,min=1"`
}
// omitempty → 欄位選填，但如果有值就要通過驗證
// 這樣使用者可以只更新標題或只更新內容
```

### ArticleQuery — form 標籤（查詢參數綁定）

```go
type ArticleQuery struct {
    Page     int    `form:"page"      binding:"omitempty,min=1"`
    PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
    Search   string `form:"search"`
    UserID   uint   `form:"user_id"`
}
// form 標籤和 json 標籤類似，但用於 URL 查詢參數
// 搭配 c.ShouldBindQuery(&query) 使用
```

---

## 常見問題（FAQ）

### Q: 為什麼不直接用 Go 的欄位名（如 ID、Username）當 JSON 名？

因為 Go 的命名慣例是 PascalCase（大寫開頭），但 JSON/API 的慣例是 snake_case 或 camelCase。如果不寫 json 標籤，JSON 會用 Go 的欄位名（`ID` 而不是 `id`），不符合 API 慣例。

### Q: `json:"-"` 和 `json:",omitempty"` 的差別？

- `json:"-"` → **永遠不會**出現在 JSON 中，不管有沒有值
- `json:",omitempty"` → 有值就會出現，零值才省略

密碼用 `json:"-"`（絕對不能出現），可選欄位用 `omitempty`。

### Q: binding:"-" 是什麼意思？

`binding:"-"` 表示 Gin 在綁定 JSON 時「忽略」這個欄位。常用於：
- ID 欄位（由系統產生，不是使用者填的）
- UserID 欄位（從 JWT Token 取得，不是使用者填的）

### Q: 為什麼 gorm 用分號，binding 用逗號？

這是各自套件的設計決定：
- `gorm:"size:200;not null"` → 分號分隔
- `binding:"required,min=3"` → 逗號分隔

沒有特別的原因，就是不同套件的語法不同，記住就好。

### Q: 結構標籤寫錯了會怎樣？

Go 編譯器**不會**檢查標籤內容。如果你寫了 `json:"usernmae"`（拼錯），程式不會報錯，但 JSON 欄位名就會是 "usernmae"。建議用 `go vet` 工具檢查常見的標籤錯誤。

### Q: Marshal 和 Unmarshal 是什麼意思？

- **Marshal**（序列化）→ Go 結構體 → JSON 字串（「把資料打包送出去」）
- **Unmarshal**（反序列化）→ JSON 字串 → Go 結構體（「把收到的包裹拆開」）

---

## 練習

1. **定義 Product 結構體：** 包含 name、price、secret_code 欄位，確保 secret_code 不會出現在 JSON 中
2. **序列化和反序列化：** 用 `json.Marshal` 把 Product 轉成 JSON，再用 `json.Unmarshal` 轉回來
3. **omitempty 實驗：** 在 Product 上加 omitempty，觀察有值和零值時的 JSON 差異
4. **完整的請求結構體：** 定義 `CreateProductRequest`，加上 `json` + `binding` + `example` 標籤
5. **GORM 結構體：** 定義一個 `Order` 結構體，包含 `json` + `gorm` 標籤，想想哪些欄位需要索引

---

## 下一課預告

**第十四課：GORM 資料庫** — 學習如何用 GORM 框架操作資料庫。本課的 `gorm` 標籤會在那裡真正派上用場：建表、查詢、新增、更新、刪除，全部用 Go 程式碼搞定，不用寫 SQL。
