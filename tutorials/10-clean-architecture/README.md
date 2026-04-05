# 第十課：架構設計（Clean Architecture）

> **這是整個教學系列最重要的一課。**
> 理解了這一課，你就能理解部落格專案的每一行程式碼為什麼要這樣寫。

## 學習目標

- 理解「為什麼需要架構」— 沒有架構的程式碼會怎樣
- 完整掌握 Clean Architecture 的四層結構
- 理解依賴方向原則和依賴反轉
- 學會依賴注入（Dependency Injection）
- 能夠在部落格專案中辨識每一層的對應關係

## 執行方式

```bash
cd tutorials/10-架構設計-clean-architecture
go run main.go
```

---

## 一、為什麼需要架構？

### 沒有架構的程式碼（義大利麵條程式碼）

想像一個沒有架構的 API handler：

```go
// 所有邏輯混在一起
func createArticle(c *gin.Context) {
    // 解析請求
    var req struct{ Title, Content string }
    c.ShouldBindJSON(&req)

    // 驗證（業務邏輯混在 handler 裡）
    if req.Title == "" {
        c.JSON(400, gin.H{"error": "標題不能為空"})
        return
    }

    // 直接操作資料庫（資料存取混在 handler 裡）
    db.Exec("INSERT INTO articles (title, content) VALUES (?, ?)", req.Title, req.Content)

    // 回傳
    c.JSON(201, gin.H{"message": "ok"})
}
```

**問題在哪？**

| 問題 | 說明 |
|------|------|
| 無法測試 | 要測試業務邏輯，必須啟動 HTTP 伺服器和資料庫 |
| 無法重用 | 如果有 CLI 介面也需要建立文章，邏輯得重寫一遍 |
| 難以維護 | 一個函式做三件事，改一個地方可能影響其他地方 |
| 無法替換 | 想從 MySQL 換成 PostgreSQL？要改每一個 handler |

### 有架構的程式碼

```go
// Handler 只負責 HTTP
func (h *ArticleHandler) Create(c *gin.Context) {
    var req domain.CreateArticleRequest
    c.ShouldBindJSON(&req)                           // 只做 HTTP 的事
    article, err := h.usecase.Create(userID, req)    // 委託給 Usecase
    response.Created(c, article)                     // 只做 HTTP 的事
}

// Usecase 只負責業務邏輯
func (u *articleUsecase) Create(userID uint, req CreateArticleRequest) (*Article, error) {
    // 只寫業務規則，不知道 HTTP 和資料庫的存在
}

// Repository 只負責資料存取
func (r *articleRepository) Create(article *Article) error {
    return r.db.Create(article).Error    // 只做資料庫的事
}
```

---

## 二、Clean Architecture 四層結構

### 總覽

```
┌──────────────────────────────────────────────────────┐
│                                                      │
│   Handler / Controller（展示層）                      │
│   「使用者怎麼跟系統互動？」                            │
│   接收 HTTP 請求、解析參數、回傳 JSON                   │
│                                                      │
│   ┌──────────────────────────────────────────┐       │
│   │                                          │       │
│   │   Usecase（業務邏輯層）                    │       │
│   │   「業務規則是什麼？」                      │       │
│   │   驗證、權限檢查、流程編排                   │       │
│   │                                          │       │
│   │   ┌──────────────────────────────┐       │       │
│   │   │                              │       │       │
│   │   │   Domain（領域層）             │       │       │
│   │   │   「業務世界長什麼樣子？」       │       │       │
│   │   │   Entity + Repository 介面    │       │       │
│   │   │                              │       │       │
│   │   └──────────────────────────────┘       │       │
│   │                                          │       │
│   └──────────────────────────────────────────┘       │
│                                                      │
│   Repository 實作（基礎設施層）                        │
│   「資料怎麼存取？」                                   │
│   GORM + SQLite 的具體實作                             │
│                                                      │
└──────────────────────────────────────────────────────┘
```

### 第 1 層：Domain（領域層）— 最內層

```
位置：internal/domain/
職責：定義業務實體和資料操作介面
依賴：不依賴任何東西（最純粹）
```

```go
// 實體：描述「文章是什麼」
type Article struct {
    ID      uint
    Title   string
    Content string
    UserID  uint
}

// 介面：描述「我需要什麼資料操作」
type ArticleRepository interface {
    Create(article *Article) error
    FindByID(id uint) (*Article, error)
    FindAll(query ArticleQuery) ([]Article, int64, error)
    Update(article *Article) error
    Delete(id uint) error
}
```

**為什麼介面定義在 Domain 層？**

這是整個架構最精妙的設計 — **依賴反轉原則（Dependency Inversion Principle）**：

```
傳統做法（錯誤的依賴方向）：
  Usecase → 具體的 MySQLRepository
  ↓
  換資料庫要改 Usecase

Clean Architecture（正確的依賴方向）：
  Domain:     定義 Repository 介面
  Usecase:    依賴 Domain 的介面
  Repository: 實作 Domain 的介面
  ↓
  換資料庫只要換 Repository 實作，Usecase 完全不改
```

### 第 2 層：Usecase（業務邏輯層）— 中間層

```
位置：internal/usecase/
職責：實作具體的業務規則
依賴：只依賴 Domain 層的介面
```

```go
type articleUsecase struct {
    articleRepo domain.ArticleRepository  // 依賴「介面」不是「實作」
}

func (u *articleUsecase) Update(id, userID uint, req UpdateArticleRequest) (*Article, error) {
    // 業務規則 1：文章必須存在
    article, err := u.articleRepo.FindByID(id)
    if err != nil {
        return nil, errors.New("文章不存在")
    }

    // 業務規則 2：只有作者可以修改
    if article.UserID != userID {
        return nil, errors.New("無權限修改此文章")
    }

    // 業務規則 3：部分更新
    if req.Title != "" {
        article.Title = req.Title
    }

    // 儲存
    u.articleRepo.Update(article)
    return article, nil
}
```

**Usecase 不知道的事：**
- 不知道用什麼資料庫（SQLite? PostgreSQL?）
- 不知道用什麼 HTTP 框架（Gin? Echo?）
- 不知道 JSON 怎麼解析
- 只知道業務規則

### 第 3 層：Repository 實作（基礎設施層）

```
位置：internal/repository/
職責：用具體技術實作 Domain 層定義的介面
依賴：依賴 Domain 層的介面和型別 + 具體的 ORM/資料庫
```

```go
type articleRepository struct {
    db *gorm.DB  // 具體的資料庫技術
}

// 實作 domain.ArticleRepository 介面
func (r *articleRepository) Create(article *domain.Article) error {
    return r.db.Create(article).Error
}

func (r *articleRepository) FindByID(id uint) (*domain.Article, error) {
    var article domain.Article
    err := r.db.Preload("User").First(&article, id).Error
    return &article, err
}
```

**Repository 的可替換性：**

```go
// 開發環境：用 SQLite
repo := repository.NewArticleRepository(sqliteDB)

// 生產環境：用 PostgreSQL（只要換這一行）
repo := repository.NewArticleRepository(postgresDB)

// 測試環境：用 Mock（記憶體 map）
repo := &mockArticleRepository{data: map[uint]*Article{}}

// Usecase 的程式碼完全不需要改！
```

### 第 4 層：Handler（展示層）— 最外層

```
位置：internal/handler/
職責：接收外部請求、呼叫 Usecase、回傳回應
依賴：依賴 Usecase 介面 + HTTP 框架
```

```go
type ArticleHandler struct {
    articleUsecase usecase.ArticleUsecase  // 依賴 Usecase 介面
}

func (h *ArticleHandler) Create(c *gin.Context) {
    // 1. HTTP 的事：解析請求
    var req domain.CreateArticleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }

    // 2. 委託給 Usecase 處理業務邏輯
    userID := c.GetUint("user_id")
    article, err := h.articleUsecase.Create(userID, req)

    // 3. HTTP 的事：回傳回應
    if err != nil {
        response.BadRequest(c, err.Error())
        return
    }
    response.Created(c, article)
}
```

**Handler 的三步驟模板：**
1. 解析請求（HTTP 的事）
2. 呼叫 Usecase（業務邏輯）
3. 回傳回應（HTTP 的事）

---

## 三、依賴注入（Dependency Injection）

所有層的組裝在 `main.go` 中完成，這是**唯一知道所有具體實作**的地方：

```go
func main() {
    // 基礎設施
    db := connectDatabase()

    // 由內而外組裝
    // Repository（實作 Domain 的介面）
    userRepo    := repository.NewUserRepository(db)
    articleRepo := repository.NewArticleRepository(db)
    commentRepo := repository.NewCommentRepository(db)

    // Usecase（注入 Repository 介面）
    userUsecase    := usecase.NewUserUsecase(userRepo, cfg)
    articleUsecase := usecase.NewArticleUsecase(articleRepo)
    commentUsecase := usecase.NewCommentUsecase(commentRepo, articleRepo)

    // Handler（注入 Usecase 介面）
    userHandler    := handler.NewUserHandler(userUsecase)
    articleHandler := handler.NewArticleHandler(articleUsecase)
    commentHandler := handler.NewCommentHandler(commentUsecase)

    // 路由
    router := handler.SetupRouter(cfg, userHandler, articleHandler, commentHandler)
    router.Run(":8080")
}
```

**依賴注入 = 把依賴從外部傳進來，而非在內部自己建立。**

```go
// 錯誤：在內部建立依賴（緊耦合）
func NewArticleUsecase() *articleUsecase {
    db := gorm.Open(sqlite.Open("blog.db"))       // Usecase 知道了資料庫！
    repo := &articleRepository{db: db}              // Usecase 知道了 Repository 實作！
    return &articleUsecase{articleRepo: repo}
}

// 正確：從外部注入依賴（鬆耦合）
func NewArticleUsecase(repo domain.ArticleRepository) ArticleUsecase {
    return &articleUsecase{articleRepo: repo}        // 只依賴介面
}
```

---

## 四、依賴方向圖

```
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│    Handler     │────>│    Usecase     │────>│    Domain      │
│  (Gin HTTP)    │     │  (業務邏輯)    │     │  (Entity+介面) │
└───────────────┘     └───────────────┘     └───────┬───────┘
                                                     ↑
                       ┌───────────────┐             │
                       │  Repository   │─────────────┘
                       │  (GORM 實作)  │  實作 Domain 定義的介面
                       └───────────────┘

箭頭方向 = 依賴方向
Handler 依賴 Usecase
Usecase 依賴 Domain（的介面）
Repository 依賴 Domain（實作其介面）

Domain 不依賴任何東西 ← 這是架構的核心
```

---

## 五、部落格專案完整對照表

| 架構層 | 目錄 | 檔案 | 職責 |
|--------|------|------|------|
| **Domain** | `internal/domain/` | `user.go` | User Entity + UserRepository 介面 |
| | | `article.go` | Article Entity + ArticleRepository 介面 |
| | | `comment.go` | Comment Entity + CommentRepository 介面 |
| **Repository** | `internal/repository/` | `user_repository.go` | 用 GORM 實作 UserRepository |
| | | `article_repository.go` | 用 GORM 實作 ArticleRepository |
| | | `comment_repository.go` | 用 GORM 實作 CommentRepository |
| **Usecase** | `internal/usecase/` | `user_usecase.go` | 註冊、登入、JWT 產生 |
| | | `article_usecase.go` | 文章 CRUD + 權限檢查 |
| | | `comment_usecase.go` | 留言 CRUD + 權限檢查 |
| **Handler** | `internal/handler/` | `user_handler.go` | 認證 API 端點 |
| | | `article_handler.go` | 文章 API 端點 |
| | | `comment_handler.go` | 留言 API 端點 |
| | | `router.go` | 路由設定 |
| **Middleware** | `internal/middleware/` | `jwt.go` 等 | 跨切面關注點 |
| **組裝** | `cmd/server/` | `main.go` | 依賴注入、啟動伺服器 |

---

## 六、每一層的測試策略

Clean Architecture 讓測試變得非常容易：

| 層 | 測試方式 | Mock 什麼 | 測試什麼 |
|----|---------|----------|---------|
| **Usecase** | 單元測試 | Mock Repository | 業務規則正確性 |
| **Handler** | HTTP 測試 | Mock Usecase | 請求解析、回應格式 |
| **Repository** | 整合測試 | 無（用測試 DB） | SQL 查詢正確性 |

```go
// 測試 Usecase 時，不需要資料庫
mockRepo := &mockArticleRepository{}
usecase := NewArticleUsecase(mockRepo)

// 測試業務規則：非作者不能修改
_, err := usecase.Update(1, 999, req)  // userID=999 不是作者
assert(err != nil)                      // 應該回傳權限錯誤
```

---

## 七、常見問題

### Q: 這麼多層不會太複雜嗎？

小專案確實可能感覺過度設計。但當專案成長到幾十個 API 端點、多個開發者協作時，清晰的分層會大幅降低維護成本。教學中使用完整架構，是為了讓你在進入真實專案時能立刻上手。

### Q: 什麼時候不需要 Clean Architecture？

- 一次性的腳本或工具
- 非常小的 CRUD 應用（< 5 個 API）
- 原型（Prototype）開發

### Q: 每次新增功能都要改四層嗎？

是的，這是刻意的 —— 它迫使你思考每一層的職責。流程是：

1. **Domain**：定義 Entity 和 Repository 介面
2. **Repository**：實作資料存取
3. **Usecase**：實作業務規則
4. **Handler**：接上 HTTP 端點
5. **main.go**：組裝依賴注入

看似繁瑣，但每一步都有明確的目標，不需要在一個大函式中思考所有事情。

### Q: 資料流是怎麼走的？

以「建立文章」為例：

```
1. 客戶端 POST /api/v1/articles
       ↓
2. Router 對應到 ArticleHandler.Create
       ↓
3. Handler 解析 JSON body → CreateArticleRequest
       ↓
4. Handler 呼叫 usecase.Create(userID, req)
       ↓
5. Usecase 驗證業務規則（標題不能空等）
       ↓
6. Usecase 呼叫 repo.Create(&article)
       ↓
7. Repository 執行 GORM: db.Create(&article)
       ↓
8. SQLite 執行 INSERT INTO articles ...
       ↓
9. 回傳路徑：SQLite → Repository → Usecase → Handler → JSON 回應
```

## 練習

1. 在 `main.go` 中新增一個 `PrintTodoRepository`（在每次操作時印出 SQL 模擬訊息），不改動任何 Usecase 程式碼
2. 新增業務規則：同一使用者最多只能有 10 個未完成的 Todo
3. 思考：如果要加「標籤（Tag）」功能，每一層分別要加什麼？
