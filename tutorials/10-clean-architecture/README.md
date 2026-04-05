# 第十課：架構設計（Clean Architecture）

> **這是整個教學系列最重要的一課。**
> 理解了這一課，你就能理解部落格專案的每一行程式碼為什麼要這樣寫。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | **重要轉折點**：從「寫程式」到「設計系統」 |
| 🔴 資深工程師 | 深化 DDD 概念，理解各層職責的邊界 |

## 學習目標

- 理解「為什麼需要架構」— 沒有架構的程式碼會怎樣
- 完整掌握 Clean Architecture 的四層結構
- 理解依賴方向原則和依賴反轉
- 學會依賴注入（Dependency Injection）
- 能夠在部落格專案中辨識每一層的對應關係

## 執行方式

```bash
go run ./tutorials/10-clean-architecture
```

---

## 用生活來理解：餐廳比喻

在學任何技術名詞之前，先想像一家餐廳：

```
╔══════════════════════════════════════════════════════════════════╗
║                                                                ║
║   🍽️ 服務生（Waiter）= Handler（展示層）                        ║
║   ├─ 接收客人的點餐（接收 HTTP 請求）                            ║
║   ├─ 把菜單交給廚師（呼叫 Usecase）                              ║
║   └─ 把做好的菜端給客人（回傳 HTTP 回應）                        ║
║                                                                ║
║   🧑‍🍳 廚師（Chef）= Usecase（業務邏輯層）                       ║
║   ├─ 知道紅燒肉要先焯水（業務規則：標題不能為空）                 ║
║   ├─ 知道客人的菜只有客人能退（權限檢查）                        ║
║   └─ 不需要知道食材從哪買、客人長什麼樣                          ║
║                                                                ║
║   📖 食譜（Recipe）= Domain（領域層）                            ║
║   ├─ 定義「紅燒肉有哪些成分」（Entity 實體）                     ║
║   ├─ 定義「廚房需要有冰箱」（Repository 介面）                   ║
║   └─ 不管冰箱是什麼牌子（不依賴具體實作）                        ║
║                                                                ║
║   🧊 冰箱（Fridge）= Repository（資料存取層）                    ║
║   ├─ 實際存放和取出食材（資料庫的具體操作）                       ║
║   ├─ 可以換成不同品牌（可以換資料庫）                            ║
║   └─ 廚師只說「給我豬肉」，不管冰箱怎麼運作                     ║
║                                                                ║
║   👔 餐廳老闆（Boss）= main() 函式                              ║
║   ├─ 買冰箱、請廚師、請服務生（建立所有元件）                    ║
║   ├─ 把冰箱交給廚師（依賴注入）                                 ║
║   └─ 是唯一知道所有具體細節的人                                 ║
║                                                                ║
╚══════════════════════════════════════════════════════════════════╝
```

**為什麼這樣分工？**

- 服務生離職了 → 換一個新服務生，廚師完全不受影響
- 冰箱壞了 → 換一台新冰箱，廚師的手藝不變
- 廚師改良食譜 → 服務生和冰箱都不需要改變
- 每個人只做自己擅長的事 → 效率更高、出錯更少

---

## 一、為什麼需要架構？

### 先解釋兩個重要概念

**耦合（Coupling）= 兩個東西綁在一起的程度**

> 想像兩個人三腳綁在一起跑步 — 這就是「高耦合」。一個人摔倒，另一個也跟著倒。
> 如果兩個人各自跑步 — 這就是「低耦合」。互不影響。
>
> 在程式中，如果 A 函式直接使用了 B 函式的具體實作，改 B 就要改 A — 這就是高耦合。

**內聚（Cohesion）= 一個模組裡的東西有多相關**

> 想像一個抽屜：如果裡面全是筷子和湯匙（餐具），找東西很快 — 這就是「高內聚」。
> 如果裡面有筷子、螺絲起子、和襪子，什麼都有 — 這就是「低內聚」。
>
> 好的架構追求：**低耦合、高內聚** — 每個模組做好自己的事，模組之間盡量獨立。

### 沒有架構的程式碼（義大利麵條程式碼）

想像一個沒有架構的 API handler：

```go
// ❌ 所有邏輯混在一起 — 就像一個人同時當服務生、廚師、還管冰箱
func createArticle(c *gin.Context) {
    // 解析請求（服務生的工作）
    var req struct{ Title, Content string }
    c.ShouldBindJSON(&req)

    // 驗證（廚師的工作，但混在服務生的程式碼裡）
    if req.Title == "" {
        c.JSON(400, gin.H{"error": "標題不能為空"})
        return
    }

    // 直接操作資料庫（冰箱的工作，也混在一起）
    db.Exec("INSERT INTO articles (title, content) VALUES (?, ?)", req.Title, req.Content)

    // 回傳（服務生的工作）
    c.JSON(201, gin.H{"message": "ok"})
}
```

**問題在哪？**

| 問題 | 生活比喻 | 技術說明 |
|------|---------|---------|
| 無法測試 | 要測試廚師的手藝，必須整間餐廳都開起來 | 要測試業務邏輯，必須啟動 HTTP 伺服器和資料庫 |
| 無法重用 | 換了一家分店，所有人都要重新訓練 | 如果有 CLI 介面也需要建立文章，邏輯得重寫一遍 |
| 難以維護 | 一個人做三份工，累了就出錯 | 一個函式做三件事，改一個地方可能影響其他地方 |
| 無法替換 | 換冰箱就要重新教廚師做菜 | 想從 MySQL 換成 PostgreSQL？要改每一個 handler |

### 有架構的程式碼

```go
// ✅ 每一層各司其職 — 就像餐廳的專業分工

// Handler 只負責 HTTP（服務生只負責接單和上菜）
func (h *ArticleHandler) Create(c *gin.Context) {
    var req domain.CreateArticleRequest
    c.ShouldBindJSON(&req)                           // 只做 HTTP 的事
    article, err := h.usecase.Create(userID, req)    // 委託給 Usecase
    response.Created(c, article)                     // 只做 HTTP 的事
}

// Usecase 只負責業務邏輯（廚師只負責炒菜）
func (u *articleUsecase) Create(userID uint, req CreateArticleRequest) (*Article, error) {
    // 只寫業務規則，不知道 HTTP 和資料庫的存在
}

// Repository 只負責資料存取（冰箱只負責存取食材）
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
│   Handler / Controller（展示層）🍽️ 服務生              │
│   「使用者怎麼跟系統互動？」                            │
│   接收 HTTP 請求、解析參數、回傳 JSON                   │
│                                                      │
│   ┌──────────────────────────────────────────┐       │
│   │                                          │       │
│   │   Usecase（業務邏輯層）🧑‍🍳 廚師             │       │
│   │   「業務規則是什麼？」                      │       │
│   │   驗證、權限檢查、流程編排                   │       │
│   │                                          │       │
│   │   ┌──────────────────────────────┐       │       │
│   │   │                              │       │       │
│   │   │   Domain（領域層）📖 食譜      │       │       │
│   │   │   「業務世界長什麼樣子？」       │       │       │
│   │   │   Entity + Repository 介面    │       │       │
│   │   │                              │       │       │
│   │   └──────────────────────────────┘       │       │
│   │                                          │       │
│   └──────────────────────────────────────────┘       │
│                                                      │
│   Repository 實作（基礎設施層）🧊 冰箱                  │
│   「資料怎麼存取？」                                   │
│   GORM + SQLite 的具體實作                             │
│                                                      │
└──────────────────────────────────────────────────────┘
```

### 第 1 層：Domain（領域層）— 最內層（📖 食譜）

```
位置：internal/domain/
職責：定義業務實體和資料操作介面
依賴：不依賴任何東西（最純粹）
```

```go
// 實體：描述「文章是什麼」（食譜中的成分清單）
type Article struct {
    ID      uint
    Title   string
    Content string
    UserID  uint
}

// 介面：描述「我需要什麼資料操作」（食譜說「廚房需要有冰箱」）
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
  食譜寫「需要一台 LG 冰箱」→ LG 停產了就完蛋
  換資料庫要改 Usecase

Clean Architecture（正確的依賴方向）：
  Domain:     定義 Repository 介面（食譜寫「需要一台冰箱」）
  Usecase:    依賴 Domain 的介面（廚師照食譜做菜）
  Repository: 實作 Domain 的介面（LG 冰箱滿足「冰箱」的定義）
  ↓
  LG 停產了？換三星的就好，廚師完全不受影響
  換資料庫只要換 Repository 實作，Usecase 完全不改
```

### 第 2 層：Usecase（業務邏輯層）— 中間層（🧑‍🍳 廚師）

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

    // 業務規則 2：只有作者可以修改（只有點餐的客人可以改菜）
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
- 不知道用什麼資料庫（SQLite? PostgreSQL?）— 廚師不知道冰箱是什麼牌子
- 不知道用什麼 HTTP 框架（Gin? Echo?）— 廚師不知道服務生是誰
- 不知道 JSON 怎麼解析 — 廚師不管客人怎麼點的餐
- 只知道業務規則 — 廚師只管怎麼把菜做好

### 第 3 層：Repository 實作（基礎設施層）（🧊 冰箱）

```
位置：internal/repository/
職責：用具體技術實作 Domain 層定義的介面
依賴：依賴 Domain 層的介面和型別 + 具體的 ORM/資料庫
```

```go
type articleRepository struct {
    db *gorm.DB  // 具體的資料庫技術（冰箱的品牌）
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

**Repository 的可替換性（換冰箱的好處）：**

```go
// 開發環境：用 SQLite（家用小冰箱）
repo := repository.NewArticleRepository(sqliteDB)

// 生產環境：用 PostgreSQL（工業大冰櫃）— 只要換這一行
repo := repository.NewArticleRepository(postgresDB)

// 測試環境：用 Mock（紙箱假裝冰箱）
repo := &mockArticleRepository{data: map[uint]*Article{}}

// Usecase 的程式碼完全不需要改！廚師的手藝不因冰箱而改變
```

### 第 4 層：Handler（展示層）— 最外層（🍽️ 服務生）

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
    // 1. HTTP 的事：解析請求（服務生接單）
    var req domain.CreateArticleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }

    // 2. 委託給 Usecase 處理業務邏輯（把菜單交給廚師）
    userID := c.GetUint("user_id")
    article, err := h.articleUsecase.Create(userID, req)

    // 3. HTTP 的事：回傳回應（把菜端給客人）
    if err != nil {
        response.BadRequest(c, err.Error())
        return
    }
    response.Created(c, article)
}
```

**Handler 的三步驟模板（服務生的 SOP）：**
1. 解析請求（接單：客人要什麼？）
2. 呼叫 Usecase（交給廚師做）
3. 回傳回應（端菜上桌）

---

## 三、依賴注入（Dependency Injection）

**依賴注入 = 把依賴從外部傳進來，而非在內部自己建立。**

用餐廳比喻：**不是讓廚師自己去買冰箱，而是老闆買好冰箱後交給廚師。**

所有層的組裝在 `main.go` 中完成，這是**唯一知道所有具體實作**的地方（老闆知道所有事）：

```go
func main() {
    // 基礎設施
    db := connectDatabase()

    // 由內而外組裝（老闆開店的步驟）
    // 第 1 步：買冰箱（Repository 實作 Domain 的介面）
    userRepo    := repository.NewUserRepository(db)
    articleRepo := repository.NewArticleRepository(db)
    commentRepo := repository.NewCommentRepository(db)

    // 第 2 步：請廚師，把冰箱交給他（Usecase 注入 Repository 介面）
    userUsecase    := usecase.NewUserUsecase(userRepo, cfg)
    articleUsecase := usecase.NewArticleUsecase(articleRepo)
    commentUsecase := usecase.NewCommentUsecase(commentRepo, articleRepo)

    // 第 3 步：請服務生，告訴他廚師是誰（Handler 注入 Usecase 介面）
    userHandler    := handler.NewUserHandler(userUsecase)
    articleHandler := handler.NewArticleHandler(articleUsecase)
    commentHandler := handler.NewCommentHandler(commentUsecase)

    // 開門營業
    router := handler.SetupRouter(cfg, userHandler, articleHandler, commentHandler)
    router.Run(":8080")
}
```

**正確 vs 錯誤的依賴注入：**

```go
// ❌ 錯誤：在內部建立依賴（緊耦合）— 廚師自己去買冰箱
func NewArticleUsecase() *articleUsecase {
    db := gorm.Open(sqlite.Open("blog.db"))       // Usecase 知道了資料庫！
    repo := &articleRepository{db: db}              // Usecase 知道了 Repository 實作！
    return &articleUsecase{articleRepo: repo}
}

// ✅ 正確：從外部注入依賴（鬆耦合）— 老闆買好冰箱交給廚師
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
│  🍽️ 服務生     │     │  🧑‍🍳 廚師      │     │  📖 食譜       │
└───────────────┘     └───────────────┘     └───────┬───────┘
                                                     ↑
                       ┌───────────────┐             │
                       │  Repository   │─────────────┘
                       │  (GORM 實作)  │  實作 Domain 定義的介面
                       │  🧊 冰箱      │
                       └───────────────┘

箭頭方向 = 依賴方向
Handler 依賴 Usecase（服務生需要廚師）
Usecase 依賴 Domain 的介面（廚師需要食譜）
Repository 依賴 Domain（冰箱滿足食譜的要求）

Domain 不依賴任何東西 ← 這是架構的核心（食譜不依賴任何人）
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
| **Usecase** | 單元測試 | Mock Repository（假冰箱） | 業務規則正確性 |
| **Handler** | HTTP 測試 | Mock Usecase（假廚師） | 請求解析、回應格式 |
| **Repository** | 整合測試 | 無（用測試 DB） | SQL 查詢正確性 |

```go
// 測試 Usecase 時，不需要資料庫（用假冰箱就夠了）
mockRepo := &mockArticleRepository{}
usecase := NewArticleUsecase(mockRepo)

// 測試業務規則：非作者不能修改（別桌客人不能改你的菜）
_, err := usecase.Update(1, 999, req)  // userID=999 不是作者
assert(err != nil)                      // 應該回傳權限錯誤
```

---

## 七、常見問題

### Q: 這麼多層不會太複雜嗎？

小專案確實可能感覺過度設計。但當專案成長到幾十個 API 端點、多個開發者協作時，清晰的分層會大幅降低維護成本。教學中使用完整架構，是為了讓你在進入真實專案時能立刻上手。

就像一間只有老闆一個人的小吃攤不需要分工，但如果要開連鎖餐廳，專業分工是必須的。

### Q: 什麼時候不需要 Clean Architecture？

- 一次性的腳本或工具（煮泡麵不需要食譜）
- 非常小的 CRUD 應用（< 5 個 API）
- 原型（Prototype）開發（先確認客人喜歡吃什麼）

### Q: 每次新增功能都要改四層嗎？

是的，這是刻意的 — 它迫使你思考每一層的職責。流程是：

1. **Domain**：定義 Entity 和 Repository 介面（寫食譜）
2. **Repository**：實作資料存取（準備冰箱裡的食材）
3. **Usecase**：實作業務規則（廚師研究怎麼炒）
4. **Handler**：接上 HTTP 端點（服務生學新菜名）
5. **main.go**：組裝依賴注入（老闆安排人員）

看似繁瑣，但每一步都有明確的目標，不需要在一個大函式中思考所有事情。

### Q: 什麼是「耦合」和「內聚」？為什麼重要？

- **耦合（Coupling）**：兩個模組之間的依賴程度。耦合越低越好 — 改 A 不需要改 B。
- **內聚（Cohesion）**：一個模組內部的相關程度。內聚越高越好 — 一個模組只做一類事。

Clean Architecture 的目標就是 **低耦合、高內聚**：
- 每一層只做自己的事（高內聚）
- 層與層之間只透過介面溝通（低耦合）

### Q: 資料流是怎麼走的？

以「建立文章」為例：

```
1. 客戶端 POST /api/v1/articles（客人走進餐廳點餐）
       ↓
2. Router 對應到 ArticleHandler.Create（帶位到服務生面前）
       ↓
3. Handler 解析 JSON body → CreateArticleRequest（服務生記下點餐內容）
       ↓
4. Handler 呼叫 usecase.Create(userID, req)（服務生把菜單交給廚師）
       ↓
5. Usecase 驗證業務規則（廚師檢查食材夠不夠、菜色搭不搭）
       ↓
6. Usecase 呼叫 repo.Create(&article)（廚師從冰箱取食材）
       ↓
7. Repository 執行 GORM: db.Create(&article)（冰箱把食材遞出來）
       ↓
8. SQLite 執行 INSERT INTO articles ...（食材從冰箱中取出）
       ↓
9. 回傳路徑：SQLite → Repository → Usecase → Handler → JSON 回應
   （食材 → 冰箱 → 廚師做好菜 → 服務生端上桌 → 客人開吃）
```

## 練習

1. 在 `main.go` 中新增一個 `PrintTodoRepository`（在每次操作時印出 SQL 模擬訊息），不改動任何 Usecase 程式碼
2. 新增業務規則：同一使用者最多只能有 10 個未完成的 Todo
3. 思考：如果要加「標籤（Tag）」功能，每一層分別要加什麼？
