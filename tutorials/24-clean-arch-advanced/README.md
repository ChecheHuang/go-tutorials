# 第二十四課：Clean Architecture 進階 + 依賴注入

> **一句話總結**：依賴注入讓每個元件只知道「介面（契約）」，不管實作細節，這樣可以獨立測試，也可以輕易替換底層技術。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | **重點**：完整的 Clean Architecture 實作，含 DI 鏈 |
| 🔴 資深工程師 | **必備**：Graceful Shutdown、Health Check、API 版本控制 |

## 你會學到什麼？

- **依賴注入（Dependency Injection）**：元件不自己建立依賴，從外部傳入
- **介面解耦**：層與層之間只透過介面溝通
- **Domain 錯誤 vs HTTP 狀態碼**：Usecase 回傳業務錯誤，Handler 轉換成 HTTP 狀態碼
- **Model 和 Entity 分開**：GORM Model（基礎設施）vs Domain Entity（業務核心）
- 把 Gin + GORM + zap + JWT 整合在一個架構裡
- 如何在 `main()` 組裝整個系統

## 執行方式

```bash
go run ./tutorials/24-clean-arch-advanced
```

然後用 curl 測試 API：

```bash
# 1. 取得文章列表（公開，不需要 Token）
curl http://localhost:8080/api/v1/articles

# 2. 登入取得 Token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"username":"Alice"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

# 3. 建立文章（需要 Token）
curl -X POST http://localhost:8080/api/v1/articles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"新文章","content":"內容"}'
```

## 架構全貌

```
┌─────────────────────────────────────────────────────┐
│                   main()                            │
│   「餐廳老闆」：買好所有設備，組裝整間餐廳            │
│                                                     │
│   db → Repository → Usecase → Handler → Router     │
│   （依賴注入鏈，每個箭頭都是「注入」）                │
└─────────────────────────────────────────────────────┘
          ↓ 組裝後啟動
┌─────────────────────────────────────────────────────┐
│  HTTP 請求                                          │
│  curl POST /api/v1/articles                         │
│           ↓                                         │
│  Middleware（LoggingMiddleware → JWTMiddleware）     │
│           ↓                                         │
│  ArticleHandler.CreateArticle()                     │
│    解析 JSON → 取得 user_id → 呼叫 Usecase           │
│           ↓                                         │
│  articleUsecase.CreateArticle()                     │
│    驗證輸入 → 業務規則 → 呼叫 Repository             │
│           ↓                                         │
│  GormArticleRepository.Create()                     │
│    建立 ArticleModel → db.Create() → 轉換成 Entity  │
│           ↓                                         │
│  回傳 Article → Usecase → Handler → JSON 回應       │
└─────────────────────────────────────────────────────┘
```

## 依賴注入（DI）詳解

### 錯誤做法（元件自己建立依賴）

```go
// ❌ 這樣寫，Usecase 跟 GORM 死死綁在一起
type articleUsecase struct{}

func (u *articleUsecase) CreateArticle(...) (*Article, error) {
    db, _ := gorm.Open(sqlite.Open("blog.db"), &gorm.Config{})  // 自己建立 DB！
    // 問題 1：每次呼叫都建立新的 DB 連線（效能爆炸）
    // 問題 2：想測試怎麼辦？只能用真實資料庫
    // 問題 3：想換成 PostgreSQL 要改這裡的程式碼
}
```

### 正確做法（依賴從外部注入）

```go
// ✅ Usecase 只持有介面，不知道底層是什麼
type articleUsecase struct {
    repo   ArticleRepository  // 介面！不是具體型別
    logger *zap.Logger
}

// 建構子：依賴從外部傳進來
func NewArticleUsecase(repo ArticleRepository, logger *zap.Logger) ArticleUsecase {
    return &articleUsecase{repo: repo, logger: logger}
}
```

### 組裝點（main.go）

```go
// 所有的「new」發生在 main()
db := initDB()

// 注入鏈：每個元件都接收自己需要的依賴
repo    := NewGormArticleRepository(db, logger)    // db 注入進 repo
uc      := NewArticleUsecase(repo, logger)         // repo 注入進 uc
handler := NewArticleHandler(uc, logger)           // uc 注入進 handler
```

## 各層的職責

| 層 | 職責 | 知道什麼 | 不知道什麼 |
|---|------|---------|-----------|
| Domain | 定義 Entity 和介面 | 業務概念 | HTTP、DB、框架 |
| Repository | 資料存取 | GORM、SQL | HTTP、業務邏輯 |
| Usecase | 業務邏輯 | Domain 介面 | HTTP、DB 細節 |
| Handler | HTTP 處理 | Gin、Usecase 介面 | DB 細節 |
| Middleware | 橫切關注點 | Gin | 業務邏輯 |
| main() | 組裝所有元件 | 全部 | — |

## Domain 錯誤 vs HTTP 狀態碼

```
業務錯誤的流向：

  Repository: gorm.ErrRecordNotFound → ErrArticleNotFound（領域錯誤）
      ↓
  Usecase: 回傳 ErrArticleNotFound（不知道 HTTP！）
      ↓
  Handler: handleError(c, err) → 轉換成 HTTP 404
```

```go
// Domain 層：定義業務錯誤
var ErrArticleNotFound  = errors.New("文章不存在")    // 不包含 HTTP 狀態碼
var ErrArticleForbidden = errors.New("沒有權限操作")

// Handler 層：轉換成 HTTP 狀態碼
func handleError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, ErrArticleNotFound):
        c.JSON(404, ...)
    case errors.Is(err, ErrArticleForbidden):
        c.JSON(403, ...)
    default:
        c.JSON(500, ...)
    }
}
```

## Model vs Entity 分開

```go
// Domain Entity（業務核心，不含框架 tag）
type Article struct {
    ID       uint
    Title    string
    AuthorID uint
    Status   string
    // 沒有 gorm tag！
}

// Repository Model（基礎設施層，含 GORM tag）
type ArticleModel struct {
    ID       uint   `gorm:"primaryKey"`
    Title    string `gorm:"size:200;not null;index"`
    AuthorID uint   `gorm:"not null;index"`
    Status   string `gorm:"size:20;default:'draft'"`
}

// 轉換函式：Model → Entity（在 Repository 層做）
func (m *ArticleModel) toArticle() *Article {
    return &Article{ID: m.ID, Title: m.Title, ...}
}
```

**為什麼要分開？**

如果 Domain Entity 直接用 GORM tag，那麼 Domain 層就依賴了 GORM，違反了「內層不依賴外層」的原則。換成 PostgreSQL 時要改 Domain，不合理。

## 如何測試（依賴注入的好處）

```go
// 建立假的 Repository（只在記憶體中操作）
type MockArticleRepository struct {
    articles map[uint]*Article
}

func (m *MockArticleRepository) FindByID(id uint) (*Article, error) {
    if a, ok := m.articles[id]; ok {
        return a, nil
    }
    return nil, ErrArticleNotFound
}
// 實作其他方法...

// 測試 Usecase（不需要真實資料庫！）
func TestCreateArticle(t *testing.T) {
    mockRepo := &MockArticleRepository{articles: map[uint]*Article{}}
    logger, _ := zap.NewDevelopment()
    uc := NewArticleUsecase(mockRepo, logger)

    article, err := uc.CreateArticle(1, CreateArticleInput{
        Title: "測試文章", Content: "測試內容",
    })
    assert.NoError(t, err)
    assert.Equal(t, "測試文章", article.Title)
}
```

## 真實專案的目錄結構

```
project/
├── cmd/
│   └── server/
│       └── main.go           ← 組裝點（DI Container）
│
├── internal/
│   ├── domain/
│   │   ├── article.go        ← Entity、介面定義
│   │   └── errors.go         ← 領域錯誤
│   │
│   ├── repository/
│   │   ├── article_repo.go   ← GormArticleRepository
│   │   └── models.go         ← ArticleModel 等 DB 模型
│   │
│   ├── usecase/
│   │   └── article_usecase.go← 業務邏輯
│   │
│   ├── handler/
│   │   ├── article_handler.go← HTTP 處理器
│   │   └── response.go       ← 共用的回應格式
│   │
│   └── middleware/
│       ├── jwt.go            ← JWT 中介層
│       └── logging.go        ← 請求日誌中介層
│
└── go.mod
```

## API 文件

| 方法 | 路徑 | 認證 | 說明 |
|------|------|------|------|
| POST | /api/v1/login | 無 | 取得 JWT Token |
| GET | /api/v1/articles | 無 | 取得已發布文章列表 |
| GET | /api/v1/articles/:id | 無 | 取得單篇文章 |
| POST | /api/v1/articles | Bearer Token | 建立文章 |
| PUT | /api/v1/articles/:id | Bearer Token | 更新文章（只有作者）|
| DELETE | /api/v1/articles/:id | Bearer Token | 刪除文章（只有作者）|

## 常見問題 FAQ

### Q: 依賴注入一定要自己手動寫嗎？

手動寫就夠了。如果專案很大（幾十個元件），可以用 Wire（Google 出的 DI 工具）自動生成注入程式碼。

### Q: Repository 要回傳 Domain Entity 還是 GORM Model？

**永遠回傳 Domain Entity**（或介面定義的型別）。GORM Model 只在 Repository 層內部使用，不應該傳出去給 Usecase 或 Handler。

### Q: 如果 Usecase 需要多個 Repository 怎麼辦？

```go
type articleUsecase struct {
    articleRepo ArticleRepository  // 文章 Repository
    userRepo    UserRepository     // 使用者 Repository（新增）
    cache       CacheRepository    // 快取 Repository（新增）
    logger      *zap.Logger
}
```

全部注入進來就好。

## 練習

1. **新增 UserRepository**：建立 `UserRepository` 介面和 `GormUserRepository` 實作，然後在 `articleUsecase.CreateArticle` 中用它驗證 `authorID` 是否存在
2. **加入 Redis 快取層**：建立 `CacheRepository` 介面，在 `GetArticle` 中先查快取，再查 DB（Cache-Aside 模式）
3. **撰寫測試**：用 `MockArticleRepository`（只存 in-memory）測試 `articleUsecase.UpdateArticle` 的權限邏輯（確認非作者會被拒絕）
4. **加入分頁到 Header**：在 `ListArticles` 回應中加上 `X-Total-Count` 和 `X-Page` HTTP Header

## 下一課預告

**第二十五課：gRPC 基礎** —— 學習 gRPC 和 Protocol Buffers，用更高效率的二進位格式在服務之間通訊（微服務架構的標配）。
