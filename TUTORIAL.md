# Go REST API 教學：使用 Gin 框架建立部落格系統

## 目錄

- [專案簡介](#專案簡介)
- [技術棧](#技術棧)
- [Clean Architecture 架構說明](#clean-architecture-架構說明)
- [專案結構](#專案結構)
- [快速啟動](#快速啟動)
- [逐層解析](#逐層解析)
  - [第一層：Domain（領域層）](#第一層domain領域層)
  - [第二層：Repository（資料存取層）](#第二層repository資料存取層)
  - [第三層：Usecase（業務邏輯層）](#第三層usecase業務邏輯層)
  - [第四層：Handler（HTTP 處理層）](#第四層handler-http-處理層)
  - [中介層（Middleware）](#中介層middleware)
  - [工具層（pkg）](#工具層pkg)
  - [主程式入口](#主程式入口)
- [API 端點一覽](#api-端點一覽)
- [功能詳細說明](#功能詳細說明)
  - [使用者認證（JWT）](#使用者認證jwt)
  - [文章 CRUD 與分頁搜尋](#文章-crud-與分頁搜尋)
  - [留言系統](#留言系統)
  - [輸入驗證](#輸入驗證)
  - [統一回應格式](#統一回應格式)
- [單元測試](#單元測試)
- [Swagger API 文件](#swagger-api-文件)
- [Docker 部署](#docker-部署)
- [實際操作範例（cURL）](#實際操作範例curl)
- [常見問題](#常見問題)

---

## 專案簡介

這是一個**教學用的部落格系統 REST API**，使用 Go 語言的 Gin 框架開發，採用 Clean Architecture 架構。專案涵蓋了實際開發中最常見的功能：

- 使用者註冊與登入（JWT 認證）
- 文章的建立、讀取、更新、刪除（CRUD）
- 留言系統（巢狀在文章之下）
- 分頁與關鍵字搜尋
- 輸入驗證
- 中介層（日誌、CORS、錯誤恢復、認證）
- Swagger API 文件自動產生
- 單元測試
- Docker 容器化部署

## 技術棧

| 技術 | 用途 | 說明 |
|------|------|------|
| [Go](https://go.dev/) | 程式語言 | Google 開發的靜態型別、編譯式語言 |
| [Gin](https://gin-gonic.com/) | Web 框架 | 高效能的 Go HTTP 框架 |
| [GORM](https://gorm.io/) | ORM | Go 語言最受歡迎的 ORM 框架 |
| [SQLite](https://www.sqlite.org/) | 資料庫 | 輕量級嵌入式資料庫，免安裝 |
| [JWT](https://jwt.io/) | 認證 | JSON Web Token，無狀態的身份驗證機制 |
| [bcrypt](https://en.wikipedia.org/wiki/Bcrypt) | 密碼雜湊 | 業界標準的密碼加密演算法 |
| [Swaggo](https://github.com/swaggo/swag) | API 文件 | 從 Go 註解自動產生 Swagger 文件 |
| [Docker](https://www.docker.com/) | 容器化 | 應用程式容器化部署 |

---

## Clean Architecture 架構說明

本專案採用 **Clean Architecture（整潔架構）**，這是 Robert C. Martin（Uncle Bob）提出的軟體架構設計原則。核心思想是**依賴方向由外向內**，讓業務邏輯不依賴於框架、資料庫或外部服務。

```
┌─────────────────────────────────────────────┐
│              Handler（HTTP 處理層）           │  ← 最外層：接收 HTTP 請求
│  ┌─────────────────────────────────────┐    │
│  │         Usecase（業務邏輯層）         │    │  ← 中間層：處理業務規則
│  │  ┌─────────────────────────────┐    │    │
│  │  │    Domain（領域層）           │    │    │  ← 最內層：定義實體與介面
│  │  │                             │    │    │
│  │  │  - User / Article / Comment │    │    │
│  │  │  - Repository 介面          │    │    │
│  │  └─────────────────────────────┘    │    │
│  │                                     │    │
│  │  Usecase 依賴 Domain 的介面          │    │
│  └─────────────────────────────────────┘    │
│                                             │
│  Handler 依賴 Usecase                        │
│  Repository 實作 Domain 的介面               │
└─────────────────────────────────────────────┘
```

### 為什麼要用 Clean Architecture？

1. **可測試性**：業務邏輯不依賴資料庫，可以用 Mock 輕鬆測試
2. **可替換性**：想把 SQLite 換成 PostgreSQL？只需要改 Repository 層
3. **關注點分離**：每一層只負責自己的事，程式碼更容易理解
4. **依賴反轉**：內層定義介面，外層實作介面，而非內層依賴外層

### 依賴方向

```
Handler → Usecase → Domain ← Repository
                      ↑
               （Domain 定義介面，
                Repository 實作介面）
```

- `Domain` 不依賴任何其他層（最純粹的業務定義）
- `Usecase` 只依賴 `Domain` 的介面
- `Handler` 依賴 `Usecase`
- `Repository` 實作 `Domain` 中定義的介面

---

## 專案結構

```
blog-api/
├── cmd/
│   └── server/
│       └── main.go                  # 應用程式入口點
├── internal/                        # 內部套件（不可被外部專案匯入）
│   ├── domain/                      # 領域層：實體定義 + Repository 介面
│   │   ├── user.go
│   │   ├── article.go
│   │   └── comment.go
│   ├── repository/                  # 資料存取層：GORM 實作
│   │   ├── user_repository.go
│   │   ├── article_repository.go
│   │   └── comment_repository.go
│   ├── usecase/                     # 業務邏輯層
│   │   ├── user_usecase.go
│   │   ├── article_usecase.go
│   │   ├── comment_usecase.go
│   │   ├── user_usecase_test.go     # 單元測試
│   │   └── article_usecase_test.go  # 單元測試
│   ├── handler/                     # HTTP 處理層：Gin 路由與控制器
│   │   ├── user_handler.go
│   │   ├── article_handler.go
│   │   ├── comment_handler.go
│   │   ├── router.go                # 路由設定
│   │   └── user_handler_test.go     # Handler 測試
│   └── middleware/                  # 中介層
│       ├── jwt.go                   # JWT 認證
│       ├── logger.go                # 請求日誌
│       ├── cors.go                  # 跨域設定
│       └── recovery.go              # Panic 恢復
├── pkg/                             # 公用套件（可被外部專案匯入）
│   ├── config/
│   │   └── config.go                # 環境變數設定管理
│   └── response/
│       └── response.go              # 統一 API 回應格式
├── docs/                            # Swagger 自動產生的文件
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── Dockerfile                       # Docker 多階段建置
├── docker-compose.yml               # Docker Compose 編排
├── .gitignore
├── go.mod                           # Go 模組定義
└── go.sum                           # 依賴版本鎖定
```

### `internal/` vs `pkg/` 的差異

Go 語言有一個特殊約定：

- **`internal/`**：只有本專案內部可以匯入，Go 編譯器會強制執行此限制。適合放專案的核心業務邏輯。
- **`pkg/`**：可以被其他專案匯入使用。適合放通用的工具函式。

---

## 快速啟動

### 前置需求

- [Go 1.21+](https://go.dev/dl/)
- [Git](https://git-scm.com/)
- （選用）[Docker](https://www.docker.com/)

### 本地開發

```bash
# 1. 安裝依賴
go mod download

# 2. 產生 Swagger 文件（首次或 API 變更後需要執行）
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go -o docs

# 3. 啟動伺服器
go run ./cmd/server/

# 4. 伺服器啟動後，可以存取：
#    - API:     http://localhost:8080/api/v1/
#    - Swagger: http://localhost:8080/swagger/index.html
```

### 使用 Docker

```bash
# 建置並啟動
docker-compose up -d

# 查看日誌
docker-compose logs -f

# 停止
docker-compose down
```

### 執行測試

```bash
# 執行所有測試
go test ./...

# 顯示詳細測試輸出
go test -v ./...

# 查看測試覆蓋率
go test -cover ./...
```

---

## 逐層解析

### 第一層：Domain（領域層）

Domain 層是整個應用程式的核心，定義了**業務實體（Entity）**和**資料存取介面（Repository Interface）**。這一層完全不依賴任何外部套件。

#### User 實體（`internal/domain/user.go`）

```go
// User 定義使用者實體
type User struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Username  string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
    Email     string    `json:"email" gorm:"uniqueIndex;size:100;not null"`
    Password  string    `json:"-" gorm:"not null"`   // json:"-" 確保密碼不會出現在 API 回應中
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**重點說明：**

- **Struct Tags（結構標籤）**：Go 使用反引號中的標籤來控制序列化行為
  - `json:"id"` — 控制 JSON 輸出時的欄位名稱
  - `json:"-"` — **排除此欄位**，密碼絕對不能出現在 API 回應中
  - `gorm:"primaryKey"` — 告訴 GORM 這是主鍵
  - `gorm:"uniqueIndex"` — 建立唯一索引，防止重複
- **時間欄位**：`CreatedAt` 和 `UpdatedAt` 是 GORM 的約定欄位，會自動維護

#### 請求/回應結構

```go
// RegisterRequest 定義使用者註冊的請求結構
type RegisterRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6,max=100"`
}
```

**`binding` 標籤**是 Gin 框架提供的輸入驗證機制：

| 標籤 | 說明 |
|------|------|
| `required` | 必填欄位 |
| `min=3` | 最小長度 3 |
| `max=50` | 最大長度 50 |
| `email` | 必須是合法的 Email 格式 |
| `omitempty` | 欄位為空時跳過驗證（用於可選欄位） |

#### Repository 介面

```go
// UserRepository 定義使用者的資料存取介面
type UserRepository interface {
    Create(user *User) error
    FindByID(id uint) (*User, error)
    FindByEmail(email string) (*User, error)
}
```

**為什麼在 Domain 層定義介面？**

這是 Clean Architecture 的核心 —— **依賴反轉原則（Dependency Inversion Principle）**：

- Domain 層定義「我需要什麼功能」（介面）
- Repository 層實作「具體怎麼做」（實作）
- Usecase 層只依賴介面，不關心具體實作

這樣做的好處是：測試時可以用 Mock 替換真實的資料庫實作。

#### Article 實體（`internal/domain/article.go`）

```go
type Article struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Title     string    `json:"title" gorm:"size:200;not null"`
    Content   string    `json:"content" gorm:"type:text;not null"`
    UserID    uint      `json:"user_id" gorm:"index;not null"`
    User      User      `json:"user" gorm:"foreignKey:UserID"`
    Comments  []Comment `json:"comments,omitempty" gorm:"foreignKey:ArticleID"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**關聯關係說明：**

- `User User` — 文章**屬於**一個使用者（多對一），`foreignKey:UserID` 指定外鍵欄位
- `Comments []Comment` — 文章**擁有多個**留言（一對多）
- `json:"comments,omitempty"` — 如果沒有留言就不輸出此欄位

```go
// ArticleQuery 定義文章查詢參數
type ArticleQuery struct {
    Page     int    `form:"page" binding:"omitempty,min=1"`
    PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
    Search   string `form:"search"`
    UserID   uint   `form:"user_id"`
}
```

注意這裡用的是 **`form` 標籤**而非 `json`，因為查詢參數來自 URL query string（`?page=1&search=go`），不是 JSON body。

#### Comment 實體（`internal/domain/comment.go`）

```go
type Comment struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Content   string    `json:"content" gorm:"type:text;not null"`
    ArticleID uint      `json:"article_id" gorm:"index;not null"`
    UserID    uint      `json:"user_id" gorm:"index;not null"`
    User      User      `json:"user" gorm:"foreignKey:UserID"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

留言同時關聯到文章（`ArticleID`）和使用者（`UserID`），形成多對一的關係。

---

### 第二層：Repository（資料存取層）

Repository 層負責與資料庫互動，實作 Domain 層定義的介面。本專案使用 **GORM** 作為 ORM 框架，搭配 **SQLite** 資料庫。

#### User Repository（`internal/repository/user_repository.go`）

```go
// userRepository 實作 domain.UserRepository 介面
type userRepository struct {
    db *gorm.DB   // GORM 資料庫連線
}

// NewUserRepository 建立使用者 Repository 實例
// 回傳介面型別而非具體型別 —— 這是依賴反轉的關鍵
func NewUserRepository(db *gorm.DB) domain.UserRepository {
    return &userRepository{db: db}
}
```

**Go 介面實作的特色：**

Go 語言使用**隱式介面實作**，不需要像 Java 那樣寫 `implements`。只要 `userRepository` 實作了 `domain.UserRepository` 定義的所有方法，它就自動滿足該介面。

```go
// FindByEmail 根據 Email 查詢使用者
func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
    var user domain.User
    if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
        return nil, err
    }
    return &user, nil
}
```

**GORM 查詢方法說明：**

| 方法 | 說明 |
|------|------|
| `db.Create(&entity)` | INSERT 新記錄 |
| `db.First(&entity, id)` | 依主鍵查詢第一筆 |
| `db.Where("欄位 = ?", 值).First(&entity)` | 條件查詢 |
| `db.Find(&entities)` | 查詢多筆 |
| `db.Save(&entity)` | UPDATE 更新 |
| `db.Delete(&entity, id)` | DELETE 刪除 |

> **安全提醒：** `Where("email = ?", email)` 使用參數化查詢（`?` 佔位符），GORM 會自動處理 SQL 跳脫，防止 SQL Injection 攻擊。**永遠不要**用字串拼接的方式組合 SQL。

#### Article Repository（`internal/repository/article_repository.go`）

文章的 Repository 比較複雜，因為需要處理**分頁**、**搜尋**和**關聯載入**。

```go
func (r *articleRepository) FindAll(query domain.ArticleQuery) ([]domain.Article, int64, error) {
    var articles []domain.Article
    var total int64

    // 設定分頁預設值
    if query.Page <= 0 {
        query.Page = 1
    }
    if query.PageSize <= 0 {
        query.PageSize = 10
    }

    // 建立查詢 builder
    db := r.db.Model(&domain.Article{})

    // 關鍵字搜尋：搜尋標題與內容
    if query.Search != "" {
        searchPattern := "%" + query.Search + "%"
        db = db.Where("title LIKE ? OR content LIKE ?", searchPattern, searchPattern)
    }

    // 依作者 ID 篩選
    if query.UserID > 0 {
        db = db.Where("user_id = ?", query.UserID)
    }

    // 先計算符合條件的總筆數（用於分頁資訊）
    if err := db.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // 計算偏移量並執行分頁查詢
    offset := (query.Page - 1) * query.PageSize
    err := db.Preload("User").
        Order("created_at DESC").
        Offset(offset).
        Limit(query.PageSize).
        Find(&articles).Error

    return articles, total, err
}
```

**分頁原理：**

```
假設總共 25 篇文章，每頁 10 篇：

第 1 頁：Offset(0).Limit(10)   → 取第 1~10 筆
第 2 頁：Offset(10).Limit(10)  → 取第 11~20 筆
第 3 頁：Offset(20).Limit(10)  → 取第 21~25 筆

公式：Offset = (Page - 1) × PageSize
```

**Preload（預載入）：**

```go
db.Preload("User").Find(&articles)
```

`Preload("User")` 告訴 GORM 自動載入每篇文章的作者資訊。沒有 Preload 的話，`article.User` 會是空值。這對應到 SQL 的 JOIN 操作，但 GORM 會用更有效率的方式（分開查詢再合併）。

---

### 第三層：Usecase（業務邏輯層）

Usecase 層包含所有的業務規則，是整個應用程式最重要的一層。

#### User Usecase（`internal/usecase/user_usecase.go`）

```go
type UserUsecase interface {
    Register(req domain.RegisterRequest) (*domain.User, error)
    Login(req domain.LoginRequest) (*domain.LoginResponse, error)
    GetByID(id uint) (*domain.User, error)
}
```

**註冊流程：**

```go
func (u *userUsecase) Register(req domain.RegisterRequest) (*domain.User, error) {
    // 1. 使用 bcrypt 雜湊密碼
    hashedPassword, err := bcrypt.GenerateFromPassword(
        []byte(req.Password),
        bcrypt.DefaultCost,  // cost = 10，越高越安全但越慢
    )
    if err != nil {
        return nil, errors.New("密碼加密失敗")
    }

    // 2. 建立使用者物件
    user := &domain.User{
        Username: req.Username,
        Email:    req.Email,
        Password: string(hashedPassword),  // 儲存雜湊後的密碼
    }

    // 3. 存入資料庫
    if err := u.userRepo.Create(user); err != nil {
        return nil, errors.New("使用者名稱或信箱已被使用")
    }

    return user, nil
}
```

> **安全重點：** 密碼永遠不能以明文儲存。`bcrypt` 會產生包含鹽值（salt）的雜湊值，即使兩個使用者的密碼相同，雜湊結果也不同。

**登入流程：**

```go
func (u *userUsecase) Login(req domain.LoginRequest) (*domain.LoginResponse, error) {
    // 1. 根據 Email 查找使用者
    user, err := u.userRepo.FindByEmail(req.Email)
    if err != nil {
        return nil, errors.New("信箱或密碼錯誤")  // 刻意使用模糊的錯誤訊息
    }

    // 2. 比對密碼
    if err := bcrypt.CompareHashAndPassword(
        []byte(user.Password),
        []byte(req.Password),
    ); err != nil {
        return nil, errors.New("信箱或密碼錯誤")  // 不透露是哪個欄位錯誤
    }

    // 3. 產生 JWT Token
    token, err := u.generateToken(user.ID)
    if err != nil {
        return nil, errors.New("產生 Token 失敗")
    }

    return &domain.LoginResponse{Token: token, User: *user}, nil
}
```

> **安全重點：** 無論是 Email 不存在還是密碼錯誤，都回傳相同的錯誤訊息「信箱或密碼錯誤」。這是為了防止攻擊者透過不同的錯誤訊息來列舉有效的 Email 地址。

**JWT Token 產生：**

```go
func (u *userUsecase) generateToken(userID uint) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID,                                          // 使用者 ID
        "exp":     time.Now().Add(u.cfg.JWT.Expiration).Unix(),    // 過期時間
        "iat":     time.Now().Unix(),                               // 簽發時間
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)     // 使用 HS256 演算法
    return token.SignedString([]byte(u.cfg.JWT.Secret))            // 用密鑰簽名
}
```

**JWT 的組成：**

```
eyJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoxLCJleHAiOjE3MTIzNTc2MDB9.xxxxx
└──── Header ────────┘ └──────── Payload ──────────────────────────┘ └Sig─┘

Header:  {"alg": "HS256"}           — 使用的演算法
Payload: {"user_id": 1, "exp": ...} — 攜帶的資料（Claims）
Signature: HMAC-SHA256 簽名          — 確保 Token 未被竄改
```

#### Article Usecase（`internal/usecase/article_usecase.go`）

文章 Usecase 包含**權限檢查**邏輯：

```go
// 只有文章作者本人可以更新
func (u *articleUsecase) Update(id, userID uint, req domain.UpdateArticleRequest) (*domain.Article, error) {
    // 1. 先查詢文章是否存在
    article, err := u.articleRepo.FindByID(id)
    if err != nil {
        return nil, errors.New("文章不存在")
    }

    // 2. 檢查是否為文章作者
    if article.UserID != userID {
        return nil, errors.New("無權限修改此文章")
    }

    // 3. 只更新有提供的欄位（部分更新）
    if req.Title != "" {
        article.Title = req.Title
    }
    if req.Content != "" {
        article.Content = req.Content
    }

    // 4. 儲存更新
    if err := u.articleRepo.Update(article); err != nil {
        return nil, errors.New("更新文章失敗")
    }

    return article, nil
}
```

#### Comment Usecase（`internal/usecase/comment_usecase.go`）

留言 Usecase 需要同時注入兩個 Repository：

```go
func NewCommentUsecase(
    commentRepo domain.CommentRepository,
    articleRepo domain.ArticleRepository,  // 需要 ArticleRepo 來驗證文章是否存在
) CommentUsecase {
    return &commentUsecase{
        commentRepo: commentRepo,
        articleRepo: articleRepo,
    }
}

func (u *commentUsecase) Create(articleID, userID uint, req domain.CreateCommentRequest) (*domain.Comment, error) {
    // 先確認文章存在，才允許建立留言
    if _, err := u.articleRepo.FindByID(articleID); err != nil {
        return nil, errors.New("文章不存在")
    }
    // ...建立留言
}
```

---

### 第四層：Handler（HTTP 處理層）

Handler 層負責：接收 HTTP 請求 → 解析參數 → 呼叫 Usecase → 回傳回應。

#### User Handler（`internal/handler/user_handler.go`）

```go
func (h *UserHandler) Register(c *gin.Context) {
    var req domain.RegisterRequest

    // ShouldBindJSON 會：
    // 1. 解析 JSON request body
    // 2. 根據 binding tag 驗證欄位
    // 3. 驗證失敗回傳錯誤
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "請求參數驗證失敗："+err.Error())
        return
    }

    user, err := h.userUsecase.Register(req)
    if err != nil {
        response.BadRequest(c, err.Error())
        return
    }

    response.Created(c, user)   // 回傳 201 Created
}
```

**Handler 的三個步驟（幾乎所有 Handler 都遵循）：**

1. **綁定與驗證**：從請求中提取資料並驗證
2. **呼叫 Usecase**：執行業務邏輯
3. **回傳回應**：使用統一的回應格式

#### Article Handler 的分頁查詢（`internal/handler/article_handler.go`）

```go
func (h *ArticleHandler) GetAll(c *gin.Context) {
    var query domain.ArticleQuery

    // ShouldBindQuery 綁定 URL 查詢參數
    // 例如：GET /api/v1/articles?page=2&page_size=5&search=Go
    if err := c.ShouldBindQuery(&query); err != nil {
        response.BadRequest(c, "查詢參數驗證失敗："+err.Error())
        return
    }

    // 設定預設值
    if query.Page <= 0 {
        query.Page = 1
    }
    if query.PageSize <= 0 {
        query.PageSize = 10
    }

    articles, total, err := h.articleUsecase.GetAll(query)
    if err != nil {
        response.InternalServerError(c, "取得文章列表失敗")
        return
    }

    // 回傳分頁格式的回應
    response.Paginated(c, articles, total, query.Page, query.PageSize)
}
```

#### 路由設定（`internal/handler/router.go`）

路由設定將 URL 路徑與 Handler 對應起來：

```go
func SetupRouter(
    cfg *config.Config,
    userHandler *UserHandler,
    articleHandler *ArticleHandler,
    commentHandler *CommentHandler,
) *gin.Engine {
    r := gin.New()

    // 全域中介層
    r.Use(middleware.Logger())
    r.Use(middleware.Recovery())
    r.Use(middleware.CORS())

    // Swagger 文件路由
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    v1 := r.Group("/api/v1")
    {
        // 公開路由（不需要登入）
        auth := v1.Group("/auth")
        {
            auth.POST("/register", userHandler.Register)
            auth.POST("/login", userHandler.Login)
        }

        // 需要 JWT 認證的路由
        authenticated := v1.Group("")
        authenticated.Use(middleware.JWTAuth(cfg))
        {
            authenticated.POST("/articles", articleHandler.Create)
            authenticated.PUT("/articles/:id", articleHandler.Update)
            authenticated.DELETE("/articles/:id", articleHandler.Delete)
            // ...
        }

        // 公開的讀取路由
        v1.GET("/articles", articleHandler.GetAll)
        v1.GET("/articles/:id", articleHandler.GetByID)
    }

    return r
}
```

**路由設計重點：**

- 使用 `v1.Group` 建立路由群組，實現 API 版本控制
- 公開路由（GET 讀取）不需要認證
- 需要認證的路由套用 `JWTAuth` 中介層
- RESTful 慣例：用 HTTP 方法區分操作（GET 讀取、POST 建立、PUT 更新、DELETE 刪除）

---

### 中介層（Middleware）

中介層在請求到達 Handler 之前（或之後）執行，形成一個**處理鏈**：

```
客戶端請求 → Logger → Recovery → CORS → [JWTAuth] → Handler → 回應
```

#### JWT 認證中介層（`internal/middleware/jwt.go`）

```go
func JWTAuth(cfg *config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 取得 Authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            response.Unauthorized(c, "缺少認證 Token")
            c.Abort()   // 中止請求鏈，不執行後續 Handler
            return
        }

        // 2. 解析 Bearer Token 格式
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            response.Unauthorized(c, "Token 格式錯誤")
            c.Abort()
            return
        }

        // 3. 驗證 JWT Token
        token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
            // 確認使用 HMAC 演算法（防止演算法替換攻擊）
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrSignatureInvalid
            }
            return []byte(cfg.JWT.Secret), nil
        })

        if err != nil || !token.Valid {
            response.Unauthorized(c, "Token 無效或已過期")
            c.Abort()
            return
        }

        // 4. 提取使用者 ID 並存入 Context
        claims := token.Claims.(jwt.MapClaims)
        userID := uint(claims["user_id"].(float64))
        c.Set("user_id", userID)

        c.Next()   // 繼續執行下一個 Handler
    }
}
```

**`c.Abort()` vs `c.Next()`：**

- `c.Abort()` — 中止後續所有 Handler，直接回傳回應
- `c.Next()` — 繼續執行下一個 Handler

**`c.Set()` / `c.Get()`：**

中介層透過 `c.Set("user_id", userID)` 將資料存入 Gin Context，後續的 Handler 可以透過 `c.GetUint("user_id")` 取出。這是中介層與 Handler 之間傳遞資料的標準方式。

#### Logger 中介層（`internal/middleware/logger.go`）

```go
func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()              // 記錄開始時間

        c.Next()                          // 執行 Handler

        latency := time.Since(start)     // 計算回應時間
        statusCode := c.Writer.Status()  // 取得狀態碼

        log.Printf("[API] %3d | %13v | %15s | %-7s %s",
            statusCode, latency, c.ClientIP(), c.Request.Method, c.Request.URL.Path)
    }
}
```

輸出範例：
```
[API] 200 |    1.234567ms |       127.0.0.1 | GET     /api/v1/articles
[API] 201 |    5.678901ms |       127.0.0.1 | POST    /api/v1/articles
[API] 401 |      234.56µs |       127.0.0.1 | POST    /api/v1/articles
```

#### CORS 中介層（`internal/middleware/cors.go`）

CORS（Cross-Origin Resource Sharing）允許瀏覽器從不同網域存取 API：

```go
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        // 處理預檢請求（Preflight Request）
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

> 如果你的前端（例如 `http://localhost:3000`）和後端（`http://localhost:8080`）不在同一個網域，瀏覽器會因為同源政策（Same-Origin Policy）而阻擋請求。CORS 中介層透過設定回應標頭來允許跨域存取。

#### Recovery 中介層（`internal/middleware/recovery.go`）

```go
func Recovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("[Recovery] panic recovered: %v\n%s", err, debug.Stack())
                response.Error(c, http.StatusInternalServerError, "伺服器內部錯誤")
                c.Abort()
            }
        }()
        c.Next()
    }
}
```

當 Handler 中發生未預期的 panic 時（例如空指標存取），Recovery 中介層會攔截它，回傳 500 錯誤而不是讓整個伺服器崩潰。`debug.Stack()` 會記錄完整的堆疊追蹤，方便除錯。

---

### 工具層（pkg）

#### 設定管理（`pkg/config/config.go`）

使用環境變數設定應用程式，未設定時使用預設值：

```go
func Load() *Config {
    return &Config{
        Server: ServerConfig{
            Port: getEnv("SERVER_PORT", "8080"),     // 預設 8080
            Mode: getEnv("GIN_MODE", "debug"),       // 預設 debug 模式
        },
        Database: DatabaseConfig{
            DSN: getEnv("DB_DSN", "blog.db"),        // 預設 blog.db
        },
        JWT: JWTConfig{
            Secret:     getEnv("JWT_SECRET", "my-secret-key-change-in-production"),
            Expiration: parseDuration(getEnv("JWT_EXPIRATION", "24h")),
        },
    }
}

func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}
```

**環境變數一覽：**

| 變數 | 預設值 | 說明 |
|------|--------|------|
| `SERVER_PORT` | `8080` | 伺服器埠號 |
| `GIN_MODE` | `debug` | Gin 模式（debug/release/test） |
| `DB_DSN` | `blog.db` | SQLite 資料庫檔案路徑 |
| `JWT_SECRET` | `my-secret-key...` | JWT 簽名密鑰（生產環境必須更換） |
| `JWT_EXPIRATION` | `24h` | Token 有效期 |

#### 統一回應格式（`pkg/response/response.go`）

所有 API 回傳相同的 JSON 結構，讓前端更容易處理：

```go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

**成功回應範例：**

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": 1,
        "title": "第一篇文章",
        "content": "..."
    }
}
```

**分頁回應範例：**

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "items": [...],
        "total": 25,
        "page": 1,
        "page_size": 10,
        "total_pages": 3
    }
}
```

**錯誤回應範例：**

```json
{
    "code": 400,
    "message": "請求參數驗證失敗：..."
}
```

---

### 主程式入口

`cmd/server/main.go` 是整個應用程式的啟動點，負責**依賴注入**與**組裝**：

```go
func main() {
    // 1. 載入設定
    cfg := config.Load()

    // 2. 連接資料庫
    db, err := gorm.Open(sqlite.Open(cfg.Database.DSN), &gorm.Config{})

    // 3. 自動遷移（建立資料表）
    db.AutoMigrate(&domain.User{}, &domain.Article{}, &domain.Comment{})

    // 4. 依賴注入：由內而外逐層建立
    //    Repository → Usecase → Handler

    // Repository 層
    userRepo := repository.NewUserRepository(db)
    articleRepo := repository.NewArticleRepository(db)
    commentRepo := repository.NewCommentRepository(db)

    // Usecase 層（注入 Repository）
    userUsecase := usecase.NewUserUsecase(userRepo, cfg)
    articleUsecase := usecase.NewArticleUsecase(articleRepo)
    commentUsecase := usecase.NewCommentUsecase(commentRepo, articleRepo)

    // Handler 層（注入 Usecase）
    userHandler := handler.NewUserHandler(userUsecase)
    articleHandler := handler.NewArticleHandler(articleUsecase)
    commentHandler := handler.NewCommentHandler(commentUsecase)

    // 5. 設定路由並啟動伺服器
    router := handler.SetupRouter(cfg, userHandler, articleHandler, commentHandler)
    router.Run(":" + cfg.Server.Port)
}
```

**依賴注入的好處：**

所有依賴都在 `main.go` 中明確組裝，沒有使用任何 DI 框架。這種方式稱為**手動依賴注入（Manual DI）**，好處是：
- 依賴關係一目瞭然
- 編譯時期就能發現缺少的依賴
- 不需要學習額外的 DI 框架

---

## API 端點一覽

### 認證相關

| 方法 | 路徑 | 說明 | 認證 |
|------|------|------|------|
| `POST` | `/api/v1/auth/register` | 使用者註冊 | 不需要 |
| `POST` | `/api/v1/auth/login` | 使用者登入 | 不需要 |
| `GET` | `/api/v1/auth/profile` | 取得個人資料 | 需要 JWT |

### 文章相關

| 方法 | 路徑 | 說明 | 認證 |
|------|------|------|------|
| `GET` | `/api/v1/articles` | 取得文章列表（分頁+搜尋） | 不需要 |
| `GET` | `/api/v1/articles/:id` | 取得文章詳情 | 不需要 |
| `POST` | `/api/v1/articles` | 建立文章 | 需要 JWT |
| `PUT` | `/api/v1/articles/:id` | 更新文章（僅作者） | 需要 JWT |
| `DELETE` | `/api/v1/articles/:id` | 刪除文章（僅作者） | 需要 JWT |

### 留言相關

| 方法 | 路徑 | 說明 | 認證 |
|------|------|------|------|
| `GET` | `/api/v1/articles/:articleId/comments` | 取得文章留言 | 不需要 |
| `POST` | `/api/v1/articles/:articleId/comments` | 建立留言 | 需要 JWT |
| `PUT` | `/api/v1/articles/:articleId/comments/:id` | 更新留言（僅留言者） | 需要 JWT |
| `DELETE` | `/api/v1/articles/:articleId/comments/:id` | 刪除留言（僅留言者） | 需要 JWT |

---

## 功能詳細說明

### 使用者認證（JWT）

#### 認證流程

```
1. 使用者註冊
   POST /api/v1/auth/register
   → 密碼經 bcrypt 雜湊後存入資料庫

2. 使用者登入
   POST /api/v1/auth/login
   → 驗證密碼 → 產生 JWT Token → 回傳給客戶端

3. 存取受保護的 API
   GET /api/v1/auth/profile
   Header: Authorization: Bearer <JWT Token>
   → JWT 中介層驗證 Token → 提取 user_id → Handler 處理請求
```

#### JWT Token 生命週期

```
產生 Token（登入時）
    ↓
客戶端儲存 Token（localStorage / Cookie）
    ↓
每次請求帶上 Token（Authorization header）
    ↓
伺服器驗證 Token（JWT 中介層）
    ↓
Token 過期後需重新登入（預設 24 小時）
```

### 文章 CRUD 與分頁搜尋

#### 分頁參數

| 參數 | 型別 | 預設值 | 說明 |
|------|------|--------|------|
| `page` | int | 1 | 頁碼 |
| `page_size` | int | 10 | 每頁筆數（最大 100） |
| `search` | string | - | 搜尋標題與內容 |
| `user_id` | int | - | 篩選特定作者的文章 |

#### 範例

```
GET /api/v1/articles?page=2&page_size=5&search=Go
```

### 留言系統

留言採用**巢狀路由**設計，路徑中包含文章 ID：

```
POST   /api/v1/articles/1/comments      → 在文章 1 下建立留言
GET    /api/v1/articles/1/comments      → 取得文章 1 的所有留言
PUT    /api/v1/articles/1/comments/3    → 更新留言 3
DELETE /api/v1/articles/1/comments/3    → 刪除留言 3
```

### 輸入驗證

Gin 框架提供 `binding` 標籤進行自動驗證：

```go
type RegisterRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6,max=100"`
}
```

驗證失敗時，Gin 會自動回傳描述性的錯誤訊息，例如：

```json
{
    "code": 400,
    "message": "請求參數驗證失敗：Key: 'RegisterRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"
}
```

### 統一回應格式

所有 API 端點都使用 `pkg/response` 套件回傳統一格式：

| 函式 | HTTP 狀態碼 | 用途 |
|------|-------------|------|
| `response.Success()` | 200 | 成功 |
| `response.Created()` | 201 | 建立成功 |
| `response.BadRequest()` | 400 | 請求參數錯誤 |
| `response.Unauthorized()` | 401 | 未認證 |
| `response.Forbidden()` | 403 | 無權限 |
| `response.NotFound()` | 404 | 找不到資源 |
| `response.InternalServerError()` | 500 | 伺服器內部錯誤 |
| `response.Paginated()` | 200 | 分頁資料 |

---

## 單元測試

本專案包含三個測試檔案，示範不同層級的測試方式。

### Usecase 層測試（Mock Repository）

使用手寫的 Mock Repository 來測試業務邏輯，不需要真實的資料庫：

```go
// Mock Repository 實作
type mockUserRepository struct {
    users  map[string]*domain.User
    nextID uint
}

func (m *mockUserRepository) Create(user *domain.User) error {
    if _, exists := m.users[user.Email]; exists {
        return errors.New("duplicate email")
    }
    user.ID = m.nextID
    m.nextID++
    m.users[user.Email] = user
    return nil
}
```

**測試案例：**

| 檔案 | 測試案例 | 驗證內容 |
|------|---------|---------|
| `user_usecase_test.go` | `TestRegister_Success` | 正常註冊流程、密碼被雜湊 |
| | `TestRegister_DuplicateEmail` | 重複 Email 被拒絕 |
| | `TestLogin_Success` | 正常登入、回傳 Token |
| | `TestLogin_WrongPassword` | 錯誤密碼被拒絕 |
| | `TestLogin_NonExistentUser` | 不存在的帳號被拒絕 |
| | `TestGetByID_Success` | 根據 ID 查詢成功 |
| `article_usecase_test.go` | `TestCreateArticle_Success` | 建立文章成功 |
| | `TestUpdateArticle_OwnerOnly` | 只有作者能更新 |
| | `TestDeleteArticle_OwnerOnly` | 只有作者能刪除 |
| | `TestGetAllArticles_WithSearch` | 搜尋功能正確 |

### Handler 層測試（Mock Usecase + httptest）

使用 Go 標準庫的 `httptest` 套件模擬 HTTP 請求：

```go
func TestRegisterHandler_Success(t *testing.T) {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    h := NewUserHandler(&mockUserUsecase{})
    r.POST("/api/v1/auth/register", h.Register)

    // 準備請求
    body := domain.RegisterRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "password123",
    }
    jsonBody, _ := json.Marshal(body)

    // 模擬 HTTP 請求
    req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
    req.Header.Set("Content-Type", "application/json")

    // 執行並驗證
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    if w.Code != http.StatusCreated {
        t.Errorf("預期狀態碼 %d，得到 %d", http.StatusCreated, w.Code)
    }
}
```

### 執行測試

```bash
# 執行所有測試
go test ./...

# 顯示詳細輸出
go test -v ./...

# 查看覆蓋率
go test -cover ./...

# 產生覆蓋率報告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

---

## Swagger API 文件

本專案使用 [Swaggo](https://github.com/swaggo/swag) 從 Go 程式碼的註解自動產生 Swagger（OpenAPI）文件。

### Swagger 註解格式

在 Handler 函式上方加上特殊格式的註解：

```go
// Register 處理使用者註冊請求
// @Summary     使用者註冊
// @Description 建立新的使用者帳號
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body domain.RegisterRequest true "註冊資訊"
// @Success     201 {object} response.Response{data=domain.User}
// @Failure     400 {object} response.Response
// @Router      /api/v1/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) { ... }
```

| 註解 | 說明 |
|------|------|
| `@Summary` | API 簡短說明 |
| `@Description` | API 詳細說明 |
| `@Tags` | API 分類標籤 |
| `@Accept` | 接受的內容類型 |
| `@Produce` | 回傳的內容類型 |
| `@Param` | 參數定義（名稱、位置、型別、是否必填、說明） |
| `@Success` | 成功回應的狀態碼與結構 |
| `@Failure` | 失敗回應的狀態碼與結構 |
| `@Security` | 需要的認證方式 |
| `@Router` | 路由路徑與 HTTP 方法 |

### 產生與存取

```bash
# 安裝 swag CLI
go install github.com/swaggo/swag/cmd/swag@latest

# 產生 Swagger 文件
swag init -g cmd/server/main.go -o docs

# 啟動伺服器後，瀏覽器開啟：
# http://localhost:8080/swagger/index.html
```

Swagger UI 提供互動式介面，可以直接在瀏覽器中測試 API。

---

## Docker 部署

### Dockerfile（多階段建置）

```dockerfile
# === 階段 1：建置 ===
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache gcc musl-dev     # SQLite 需要 C 編譯器
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download                      # 先下載依賴（利用 Docker 快取）
COPY . .
RUN CGO_ENABLED=1 go build -o /app/server ./cmd/server/

# === 階段 2：執行 ===
FROM alpine:3.19
RUN apk add --no-cache ca-certificates sqlite-libs
WORKDIR /app
COPY --from=builder /app/server .        # 只複製編譯好的二進位檔
EXPOSE 8080
CMD ["./server"]
```

**多階段建置的好處：**

| 階段 | 映像大小 | 內容 |
|------|----------|------|
| builder | ~800MB | Go 編譯器 + 原始碼 + 依賴 |
| 最終映像 | ~15MB | 只有二進位檔 + 執行環境 |

### docker-compose.yml

```yaml
services:
  blog-api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - GIN_MODE=release
      - JWT_SECRET=please-change-this-in-production
    volumes:
      - blog-data:/app/data   # 持久化 SQLite 資料庫
    restart: unless-stopped

volumes:
  blog-data:
```

### Docker 常用指令

```bash
# 建置並啟動（背景執行）
docker-compose up -d --build

# 查看日誌
docker-compose logs -f blog-api

# 停止並移除容器
docker-compose down

# 停止並移除容器 + 資料卷（會刪除資料庫）
docker-compose down -v
```

---

## 實際操作範例（cURL）

### 1. 註冊使用者

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "email": "alice@example.com",
    "password": "password123"
  }'
```

回應：
```json
{
  "code": 201,
  "message": "created",
  "data": {
    "id": 1,
    "username": "alice",
    "email": "alice@example.com",
    "created_at": "2026-04-05T16:30:00Z",
    "updated_at": "2026-04-05T16:30:00Z"
  }
}
```

### 2. 登入取得 Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "password123"
  }'
```

回應：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "alice",
      "email": "alice@example.com"
    }
  }
}
```

### 3. 建立文章（需要 Token）

```bash
curl -X POST http://localhost:8080/api/v1/articles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "title": "Go 語言入門教學",
    "content": "Go 是 Google 開發的程式語言..."
  }'
```

### 4. 取得文章列表（分頁 + 搜尋）

```bash
# 第一頁，每頁 5 筆
curl "http://localhost:8080/api/v1/articles?page=1&page_size=5"

# 搜尋關鍵字
curl "http://localhost:8080/api/v1/articles?search=Go"

# 篩選特定作者
curl "http://localhost:8080/api/v1/articles?user_id=1"
```

### 5. 取得文章詳情

```bash
curl http://localhost:8080/api/v1/articles/1
```

### 6. 更新文章（僅作者可操作）

```bash
curl -X PUT http://localhost:8080/api/v1/articles/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{
    "title": "Go 語言入門教學（更新版）"
  }'
```

### 7. 建立留言

```bash
curl -X POST http://localhost:8080/api/v1/articles/1/comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{
    "content": "寫得很好，謝謝分享！"
  }'
```

### 8. 取得文章留言

```bash
curl http://localhost:8080/api/v1/articles/1/comments
```

### 9. 刪除文章

```bash
curl -X DELETE http://localhost:8080/api/v1/articles/1 \
  -H "Authorization: Bearer <TOKEN>"
```

---

## 常見問題

### Q: 為什麼選擇 SQLite 而不是 PostgreSQL/MySQL？

SQLite 是嵌入式資料庫，不需要額外安裝和設定資料庫伺服器。對於教學和本地開發來說非常方便。如果要用於生產環境，只需要：

1. 安裝對應的 GORM driver（例如 `gorm.io/driver/postgres`）
2. 修改 `main.go` 中的資料庫連線程式碼
3. 更新環境變數

Repository 層和其他所有程式碼都不需要修改 —— 這就是 Clean Architecture 的好處。

### Q: JWT Secret 應該怎麼設定？

預設的 Secret 只適合本地開發。在生產環境中，應該：

```bash
# 產生隨機密鑰
openssl rand -hex 32

# 設定為環境變數
export JWT_SECRET="你產生的隨機密鑰"
```

### Q: 為什麼 CORS 設定為允許所有來源（`*`）？

這是為了簡化教學。在生產環境中，應該設定為你的前端網域：

```go
c.Writer.Header().Set("Access-Control-Allow-Origin", "https://your-frontend.com")
```

### Q: 如何新增更多的 API 端點？

按照 Clean Architecture 的分層，依序進行：

1. 在 `internal/domain/` 定義實體與 Repository 介面
2. 在 `internal/repository/` 實作 Repository
3. 在 `internal/usecase/` 撰寫業務邏輯
4. 在 `internal/handler/` 建立 HTTP Handler
5. 在 `internal/handler/router.go` 註冊路由
6. 在 `cmd/server/main.go` 組裝依賴注入

### Q: AutoMigrate 在生產環境安全嗎？

`db.AutoMigrate()` 只會新增欄位和建立資料表，不會刪除或修改現有的欄位。在小型專案中可以直接使用。大型專案建議使用專門的資料庫遷移工具，例如 [golang-migrate](https://github.com/golang-migrate/migrate)。

### Q: 如何擴展到使用者角色與權限？

可以在 User 實體中加入 `Role` 欄位：

```go
type User struct {
    // ...
    Role string `json:"role" gorm:"default:user"` // user / admin / editor
}
```

然後建立一個角色檢查的中介層，在需要特定權限的路由上使用。
