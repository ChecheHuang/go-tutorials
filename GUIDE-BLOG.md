# 部落格 API 從零建構指南

> 這份指南帶你理解部落格 API 是怎麼一步一步蓋起來的，每一步都對應具體的教學課程。
> 不只告訴你「怎麼做」，更解釋「為什麼這樣設計」。

---

## 閱讀順序

建議按以下順序閱讀專案程式碼，由內而外、由簡到繁：

### 第一層：理解資料模型（第 4、13、14 課）

1. `internal/domain/article.go` — 文章實體
2. `internal/domain/user.go` — 使用者實體
3. `internal/domain/comment.go` — 留言實體

> **為什麼先讀這裡？** Domain 是系統的核心。在 Clean Architecture 中，domain 層不依賴任何外部套件，
> 它定義了「這個系統在講什麼故事」。先理解資料長什麼樣，後面的程式碼才看得懂。

### 第二層：理解資料存取（第 14、15、28 課）

4. `internal/repository/user_repository.go` — 使用者資料庫操作
5. `internal/repository/article_repository.go` — 文章 CRUD + 分頁搜尋

> **為什麼是第二層？** 有了資料模型後，下一個問題是「資料存在哪裡、怎麼取出來」。
> Repository 層把資料庫的細節封裝起來，上層只需要呼叫介面，不需要知道底層用的是 SQLite 還是 PostgreSQL。

### 第三層：理解商業邏輯（第 10、16 課）

6. `internal/usecase/user_usecase.go` — 註冊 / 登入邏輯
7. `internal/usecase/article_usecase.go` — 文章權限檢查

> **為什麼要分出 Usecase 層？** 「密碼要加密」「只有作者能刪自己的文章」這些是商業規則，
> 不該寫在 HTTP handler 裡（換成 gRPC 就要重寫），也不該寫在 repository 裡（那是資料庫的事）。
> Usecase 層就是專門放這些規則的地方。

### 第四層：理解 HTTP 介面（第 12、13、17、18 課）

8. `internal/handler/router.go` — 路由設定
9. `internal/handler/user_handler.go` — 使用者 API
10. `internal/handler/article_handler.go` — 文章 API
11. `internal/middleware/jwt.go` — JWT 認證中介層
12. `internal/middleware/cors.go` — 跨域設定

> **為什麼 HTTP 放最外層？** 因為 HTTP 只是「遞送方式」之一。如果哪天要改成 gRPC，
> 只需要換掉 handler 層，domain 和 usecase 完全不用動。這就是分層架構的威力。

### 第五層：理解基礎設施（第 20、21、26、29 課）

13. `pkg/config/config.go` — 設定管理
14. `pkg/logger/logger.go` — 結構化日誌
15. `internal/repository/article_cache_repository.go` — Redis 快取裝飾器
16. `cmd/server/main.go` — 程式進入點 + Graceful Shutdown

> **為什麼這些放在 pkg？** `pkg` 目錄放的是「可以被其他專案重用」的通用工具。
> 設定管理和日誌不是部落格專屬的邏輯，任何 Go 專案都需要，所以它們放在 `pkg` 而不是 `internal`。

### 第六層：理解部署與監控（第 23、34、35、36 課）

17. `Dockerfile` — Docker 容器化
18. `docker-compose.yml` — 多服務編排
19. `internal/middleware/metrics.go` — Prometheus 指標

> **為什麼最後才看部署？** 因為部署是把已經寫好的東西打包出去。
> 先理解程式怎麼跑，再理解怎麼把它裝進容器、怎麼監控它的健康狀態。

---

## 從零建構步驟

### 步驟 1：建立領域模型（學完第 4、13 課後）

**先修課程：**
- 第 4 課 — 結構體與方法：學會用 `struct` 定義資料結構
- 第 13 課 — JSON 與 Struct Tags：學會用 tag 控制 JSON 序列化與驗證行為

**要建立的檔案：**
- `internal/domain/user.go`
- `internal/domain/article.go`
- `internal/domain/comment.go`

**關鍵程式碼模式：**

```go
// 實體定義：用 struct tag 同時控制 JSON 輸出和資料庫行為
type User struct {
    ID       uint   `json:"id" gorm:"primaryKey"`
    Username string `json:"username" gorm:"uniqueIndex;size:50;not null"`
    Password string `json:"-" gorm:"not null"` // json:"-" 隱藏密碼
}

// 請求結構：用 binding tag 做輸入驗證
type RegisterRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6,max=100"`
}

// Repository 介面：定義在 domain 層，由外層實作
type UserRepository interface {
    Create(user *User) error
    FindByID(id uint) (*User, error)
    FindByEmail(email string) (*User, error)
}
```

**設計決策解析：**

1. **為什麼 `Password` 的 json tag 是 `"-"`？**
   安全考量。API 回應中絕對不能把密碼回傳給前端，`json:"-"` 讓 `json.Marshal` 自動忽略這個欄位。

2. **為什麼 Repository 介面定義在 domain 層？**
   這是 Clean Architecture 的「依賴反轉原則」（Dependency Inversion Principle）。domain 層定義「我需要什麼能力」，
   由外層決定「怎麼實現這個能力」。這樣 domain 層不需要 import 任何資料庫套件。

3. **為什麼要分開 Entity 和 Request？**
   Entity（如 `User`）代表資料庫裡的完整資料；Request（如 `RegisterRequest`）代表使用者提交的輸入。
   兩者的欄位和驗證規則不同，分開才不會搞混。

**完成後驗證：**
- 確認每個 struct 的 json tag 和 gorm tag 都正確
- 確認 Repository 介面的方法簽名能涵蓋所有 CRUD 需求
- 確認 Request struct 的 binding tag 能擋住不合法的輸入

---

### 步驟 2：建立資料庫層（學完第 14 課後）

**先修課程：**
- 第 14 課 — GORM 資料庫：學會 ORM 的 CRUD 操作、AutoMigrate、Preload

**要建立的檔案：**
- `internal/repository/user_repository.go`
- `internal/repository/article_repository.go`
- `internal/repository/comment_repository.go`

**關鍵程式碼模式：**

```go
// 私有結構體 + 公開建構函式：封裝實作細節
type articleRepository struct {
    db *gorm.DB  // 依賴注入，不自己建立連線
}

func NewArticleRepository(db *gorm.DB) domain.ArticleRepository {
    return &articleRepository{db: db}  // 回傳介面型別，隱藏實作
}

// Preload 預載入關聯資料，避免 N+1 查詢問題
func (r *articleRepository) FindByID(id uint) (*domain.Article, error) {
    var article domain.Article
    err := r.db.Preload("User").Preload("Comments").Preload("Comments.User").
        First(&article, id).Error
    if err != nil {
        return nil, err
    }
    return &article, nil
}

// 分頁查詢：先 Count 總數，再 Offset + Limit 取資料
func (r *articleRepository) FindAll(query domain.ArticleQuery) ([]domain.Article, int64, error) {
    db := r.db.Model(&domain.Article{})
    if query.Search != "" {
        searchPattern := "%" + query.Search + "%"
        db = db.Where("title LIKE ? OR content LIKE ?", searchPattern, searchPattern)
    }

    var total int64
    db.Count(&total)

    offset := (query.Page - 1) * query.PageSize
    var articles []domain.Article
    err := db.Preload("User").Order("created_at DESC").
        Offset(offset).Limit(query.PageSize).Find(&articles).Error
    return articles, total, err
}
```

**設計決策解析：**

1. **為什麼建構函式回傳介面而不是結構體？**
   `NewArticleRepository` 回傳 `domain.ArticleRepository`（介面），而不是 `*articleRepository`（結構體）。
   這樣呼叫者只知道介面，不依賴具體實作。想換成 PostgreSQL 實作？只要寫一個新的結構體實作同一個介面就行。

2. **為什麼用 Preload 而不是手動 JOIN？**
   GORM 的 `Preload` 會自動幫你處理關聯查詢（例如載入文章的作者資料）。
   雖然底層會多發幾次 SQL 查詢，但程式碼更簡潔、更不容易出錯。

3. **為什麼 articleRepository 是小寫開頭（私有）？**
   Go 的命名慣例：小寫開頭代表只能在同一個 package 內存取。
   外部只需要透過 `NewArticleRepository` 拿到介面，不需要直接操作結構體。

**完成後驗證：**
- 寫一個簡單的 `main.go` 呼叫 `gorm.Open` + `AutoMigrate`，確認 table 能建立
- 嘗試 `Create` + `FindByID`，確認資料能存取
- 確認 `Preload` 能正確載入關聯的 User 資料

---

### 步驟 3：建立業務邏輯層（學完第 10 課後）

**先修課程：**
- 第 10 課 — 架構設計：學會 Clean Architecture 的分層思想
- 第 16 課 — Error Wrapping：學會用 `fmt.Errorf %w` 包裝錯誤

**要建立的檔案：**
- `internal/usecase/user_usecase.go`
- `internal/usecase/article_usecase.go`
- `internal/usecase/comment_usecase.go`
- `pkg/apperror/error.go`

**關鍵程式碼模式：**

```go
// Usecase 也定義介面，方便測試時 mock
type ArticleUsecase interface {
    Create(userID uint, req domain.CreateArticleRequest) (*domain.Article, error)
    Update(id, userID uint, req domain.UpdateArticleRequest) (*domain.Article, error)
    Delete(id, userID uint) error
}

// 透過建構函式注入依賴的 Repository
type articleUsecase struct {
    articleRepo domain.ArticleRepository
}

func NewArticleUsecase(articleRepo domain.ArticleRepository) ArticleUsecase {
    return &articleUsecase{articleRepo: articleRepo}
}

// 商業規則：只有作者本人可以更新文章
func (u *articleUsecase) Update(id, userID uint, req domain.UpdateArticleRequest) (*domain.Article, error) {
    article, err := u.articleRepo.FindByID(id)
    if err != nil {
        return nil, apperror.Wrap(apperror.ErrNotFound, "文章 ID=%d", id)
    }
    if article.UserID != userID {
        return nil, apperror.Wrap(apperror.ErrForbidden, "無權限修改文章 ID=%d", id)
    }
    // ... 更新邏輯
}
```

```go
// apperror：用 Sentinel Error + Wrap 模式統一錯誤處理
var (
    ErrNotFound     = errors.New("資源不存在")
    ErrForbidden    = errors.New("無權限")
    ErrUnauthorized = errors.New("未授權")
)

func Wrap(sentinel error, format string, args ...any) error {
    msg := fmt.Sprintf(format, args...)
    return fmt.Errorf("%s: %w", msg, sentinel)
}

// 判斷錯誤類型，對應 HTTP 狀態碼
func HTTPStatus(err error) int {
    switch {
    case errors.Is(err, ErrNotFound):
        return http.StatusNotFound
    case errors.Is(err, ErrForbidden):
        return http.StatusForbidden
    default:
        return http.StatusInternalServerError
    }
}
```

**設計決策解析：**

1. **為什麼權限檢查放在 Usecase 而不是 Handler？**
   「只有作者能刪自己的文章」是商業規則，不是 HTTP 的事。如果你未來加上 gRPC 或 CLI 介面，
   這個規則應該一樣生效。放在 Usecase 層就能確保所有入口都走相同的權限檢查。

2. **為什麼要自定義 apperror 套件？**
   Go 的 error 是一個介面，內容只是字串。如果 handler 拿到 `"文章不存在"` 的 error，
   它怎麼知道該回 404 還是 500？用 Sentinel Error + `errors.Is` 就能判斷錯誤類型，
   再用 `HTTPStatus` 轉成正確的 HTTP 狀態碼。

3. **為什麼密碼加密放在 Usecase 而不是 Repository？**
   密碼加密是商業邏輯（「我們決定要用 bcrypt」），不是資料庫的事。
   Repository 只負責把資料存進去和取出來，不該知道密碼要怎麼處理。

**完成後驗證：**
- 用 `apperror.Wrap` 包裝一個錯誤，用 `errors.Is` 驗證能正確判斷類型
- 確認 `userUsecase.Register` 能把密碼加密後存入資料庫
- 確認 `articleUsecase.Delete` 在非作者操作時回傳 `ErrForbidden`

---

### 步驟 4：建立 HTTP 介面（學完第 12、13 課後）

**先修課程：**
- 第 12 課 — Gin 框架：學會路由設定、路由群組、參數綁定
- 第 13 課 — JSON 與 Struct Tags：學會 `ShouldBindJSON` / `ShouldBindQuery`

**要建立的檔案：**
- `internal/handler/router.go`
- `internal/handler/user_handler.go`
- `internal/handler/article_handler.go`
- `internal/handler/comment_handler.go`
- `internal/handler/error_helper.go`
- `pkg/response/response.go`

**關鍵程式碼模式：**

```go
// Handler 依賴 Usecase 介面，不依賴具體實作
type ArticleHandler struct {
    articleUsecase usecase.ArticleUsecase
}

// Gin handler 的標準寫法：綁定 → 驗證 → 呼叫 usecase → 回應
func (h *ArticleHandler) Create(c *gin.Context) {
    var req domain.CreateArticleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "請求參數驗證失敗："+err.Error())
        return  // Early Return：驗證失敗就直接回傳
    }

    userID := c.GetUint("user_id")  // 從 JWT 中介層取得已驗證的使用者 ID

    article, err := h.articleUsecase.Create(userID, req)
    if err != nil {
        handleError(c, err)  // 統一錯誤處理：根據 apperror 類型回傳對應 HTTP 狀態碼
        return
    }

    response.Created(c, article)
}
```

```go
// 路由群組：公開路由 vs 需要認證的路由
v1 := r.Group("/api/v1")

// 公開路由：任何人都能存取
auth := v1.Group("/auth")
auth.POST("/register", userHandler.Register)
auth.POST("/login", userHandler.Login)

// 受保護路由：需要 JWT Token
authenticated := v1.Group("")
authenticated.Use(middleware.JWTAuth(cfg))
authenticated.POST("/articles", articleHandler.Create)

// 只讀路由：不需要認證
v1.GET("/articles", articleHandler.GetAll)
v1.GET("/articles/:id", articleHandler.GetByID)
```

```go
// 統一回應格式：所有 API 都用相同結構
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

**設計決策解析：**

1. **為什麼用 Early Return 模式？**
   每個 handler 都有多個可能失敗的步驟（綁定、驗證、業務邏輯）。
   用 early return 可以避免 `if-else` 層層巢狀，讓程式碼一目了然：
   「失敗就回傳錯誤，通過所有檢查後才執行正常邏輯。」

2. **為什麼要統一 Response 格式？**
   前端開發者只需要記住一種回應結構。不管是成功還是失敗，`code` 和 `message` 一定有，
   `data` 在成功時才出現。這樣前端可以寫統一的錯誤處理邏輯。

3. **為什麼讀取文章不需要認證，但建立文章需要？**
   這是常見的部落格設計：任何人都能瀏覽文章（推廣內容），
   但只有登入的使用者才能發文和留言（避免垃圾內容）。

**完成後驗證：**
- 用 `curl` 或 Postman 測試 `POST /api/v1/auth/register` 能建立使用者
- 測試 `POST /api/v1/auth/login` 能拿到 JWT Token
- 測試 `GET /api/v1/articles` 不帶 Token 也能正常回傳
- 測試 `POST /api/v1/articles` 不帶 Token 會回傳 401

---

### 步驟 5：加入認證（學完第 17、18 課後）

**先修課程：**
- 第 17 課 — 中介層：學會 Gin Middleware 的運作機制（`c.Next()`、`c.Abort()`）
- 第 18 課 — JWT 認證：學會 Token 的簽發與驗證流程

**要建立的檔案：**
- `internal/middleware/jwt.go`
- `internal/middleware/cors.go`

**關鍵程式碼模式：**

```go
// JWT 中介層：驗證 Token 並把 user_id 存入 Context
func JWTAuth(cfg *config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            response.Unauthorized(c, "缺少認證 Token")
            c.Abort()  // 停止執行後續的 handler
            return
        }

        // 解析 Bearer Token
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            response.Unauthorized(c, "Token 格式錯誤")
            c.Abort()
            return
        }

        // 驗證 JWT 簽名與有效期
        token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrSignatureInvalid  // 防止演算法替換攻擊
            }
            return []byte(cfg.JWT.Secret), nil
        })

        // 提取 user_id 存入 Gin Context，供 handler 使用
        claims := token.Claims.(jwt.MapClaims)
        userID := uint(claims["user_id"].(float64))
        c.Set("user_id", userID)

        c.Next()  // 驗證通過，繼續執行下一個 handler
    }
}
```

**設計決策解析：**

1. **為什麼用中介層而不是在每個 handler 裡檢查 Token？**
   DRY 原則（Don't Repeat Yourself）。如果有 10 個需要認證的 API，
   用中介層只需要寫一次認證邏輯，然後套用到路由群組上。

2. **為什麼要檢查 SigningMethod？**
   這是知名的 JWT 安全漏洞。如果不檢查，攻擊者可以把演算法從 HS256 改成 "none"，
   繞過簽名驗證。這行防禦看似簡單，卻是安全性的關鍵。

3. **為什麼 `c.Abort()` 和 `return` 要同時用？**
   `c.Abort()` 告訴 Gin 不要執行後續的 handler，但它不會停止當前函式的執行。
   如果不 `return`，後面的程式碼還是會跑到，可能造成非預期的行為。

**完成後驗證：**
- 測試不帶 Token 存取受保護路由 → 應回傳 401
- 測試帶錯誤格式的 Token → 應回傳 401
- 測試帶正確 Token → handler 能從 `c.GetUint("user_id")` 拿到正確的使用者 ID
- 測試過期 Token → 應回傳 401

---

### 步驟 6：加入工程品質（學完第 19-21 課後）

**先修課程：**
- 第 19 課 — 單元測試：學會 `testing` 套件、`testify`、表格驅動測試
- 第 20 課 — Config 管理：學會 Viper 讀取設定檔 + 環境變數覆蓋
- 第 21 課 — 結構化日誌：學會 `log/slog`、JSON 日誌格式

**要建立的檔案：**
- `pkg/config/config.go`
- `pkg/logger/logger.go`
- `config.yaml`（設定檔範本）
- `internal/usecase/article_usecase_test.go`
- `internal/usecase/user_usecase_test.go`

**關鍵程式碼模式：**

```go
// Config：用 Viper 支援多來源設定
func Load() *Config {
    v := viper.New()
    v.SetConfigName("config")
    v.SetConfigType("yaml")
    v.AddConfigPath(".")

    // 預設值：確保即使沒有設定檔也能啟動
    v.SetDefault("server.port", "8080")
    v.SetDefault("jwt.secret", "my-secret-key-change-in-production")

    // 環境變數覆蓋：部署時用環境變數覆蓋設定
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    v.AutomaticEnv()

    v.ReadInConfig()  // 找不到檔案也不 panic
    // ...
}
```

```go
// Logger：基於 Go 1.21+ 的 slog，支援 JSON 與 Text 格式
func Init(level, format string) {
    var handler slog.Handler
    if strings.ToLower(format) == "json" {
        handler = slog.NewJSONHandler(os.Stdout, opts)  // 生產環境用 JSON
    } else {
        handler = slog.NewTextHandler(os.Stdout, opts)  // 開發環境用 Text
    }
    slog.SetDefault(slog.New(handler))
}
```

**設計決策解析：**

1. **為什麼設定要支援「設定檔 + 環境變數」兩種來源？**
   本地開發用 `config.yaml` 比較方便（改一個檔案就好），
   但 Docker / Kubernetes 部署時，用環境變數覆蓋設定更靈活（不需要重新打包映像）。

2. **為什麼日誌格式要可切換？**
   開發時 Text 格式方便閱讀，生產環境用 JSON 格式方便 ELK / Grafana Loki 等日誌系統解析。
   透過設定切換，同一份程式碼兩個場景都適用。

3. **為什麼 `ReadInConfig` 失敗不 panic？**
   在 Docker 環境中，可能沒有 `config.yaml`，完全靠環境變數提供設定。
   如果找不到設定檔就 crash，反而會造成部署困難。

**完成後驗證：**
- 跑 `go test ./internal/usecase/...`，確認所有測試通過
- 不放 `config.yaml`，只設環境變數，確認程式能正常啟動
- 切換 `LOG_FORMAT=json`，確認日誌輸出是 JSON 格式

---

### 步驟 7：加入快取（學完第 26 課後）

**先修課程：**
- 第 26 課 — Redis 快取：學會快取策略、Cache Aside Pattern

**要建立的檔案：**
- `pkg/cache/cache.go`
- `internal/repository/article_cache_repository.go`

**關鍵程式碼模式：**

```go
// 快取介面：讓快取實作可替換
type Cache interface {
    Get(ctx context.Context, key string, dest any) error
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}

// NoOpCache：Redis 未啟用時的降級方案
type NoOpCache struct{}
func (c *NoOpCache) Get(_ context.Context, _ string, _ any) error {
    return redis.Nil  // 永遠回傳 cache miss
}
```

```go
// 裝飾器模式：在原有 Repository 外面包一層快取
type CachedArticleRepository struct {
    repo  domain.ArticleRepository  // 被裝飾的原始 Repository
    cache cache.Cache
}

func (r *CachedArticleRepository) FindByID(id uint) (*domain.Article, error) {
    key := fmt.Sprintf("article:%d", id)

    // 1. 先查快取
    var cached domain.Article
    if err := r.cache.Get(ctx, key, &cached); err == nil {
        return &cached, nil  // 快取命中，直接回傳
    }

    // 2. 快取未命中，查資料庫
    article, err := r.repo.FindByID(id)
    if err != nil {
        return nil, err
    }

    // 3. 寫入快取（失敗不影響回傳）
    r.cache.Set(ctx, key, article, 5*time.Minute)
    return article, nil
}

// 更新或刪除時，清除對應的快取
func (r *CachedArticleRepository) Update(article *domain.Article) error {
    if err := r.repo.Update(article); err != nil {
        return err
    }
    r.cache.Delete(ctx, fmt.Sprintf("article:%d", article.ID))
    return nil
}
```

**設計決策解析：**

1. **為什麼用裝飾器模式而不是直接在 Repository 裡加快取邏輯？**
   裝飾器模式讓你可以「選擇性」地加上快取。`NewCachedArticleRepository` 包裝原始的 `articleRepository`，
   兩者都實作同一個介面。不需要快取的時候，直接用原始 Repository 就好。

2. **為什麼列表查詢不做快取？**
   列表有分頁、搜尋、篩選等各種組合，快取的 key 會非常多，命中率很低。
   相比之下，單篇文章的快取（以 ID 為 key）命中率高很多，效益更好。

3. **為什麼 NoOpCache 回傳 `redis.Nil` 而不是其他 error？**
   `redis.Nil` 代表 cache miss。`CachedArticleRepository` 在收到 `redis.Nil` 時會去查資料庫。
   如果 NoOpCache 回傳其他 error，可能會被當成「快取故障」而觸發不必要的警告日誌。

4. **為什麼快取寫入失敗不回傳 error？**
   快取只是加速手段，不是核心邏輯。即使 Redis 掛了，資料庫還是能正常回應。
   這叫做「優雅降級」（Graceful Degradation）。

**完成後驗證：**
- Redis 啟用時，第二次查詢同一篇文章應該走快取（檢查 debug 日誌中的「快取命中」）
- Redis 未啟用時，程式應正常啟動並運作（使用 NoOpCache）
- 更新文章後，下一次查詢應該拿到最新資料（快取已被清除）

---

### 步驟 8：加入監控（學完第 35 課後）

**先修課程：**
- 第 35 課 — Prometheus：學會定義 Counter、Histogram、Gauge 指標

**要建立的檔案：**
- `internal/middleware/metrics.go`

**關鍵程式碼模式：**

```go
// 定義三種指標
var (
    // Counter：只會增加，用來計算請求總數
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "HTTP 請求總數",
        },
        []string{"method", "path", "status"},
    )

    // Histogram：記錄數值分佈，用來測量請求延遲
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP 請求處理時間（秒）",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )

    // Gauge：可增可減，用來追蹤當前活躍連線數
    httpActiveConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "http_active_connections",
            Help: "目前正在處理的連線數",
        },
    )
)

// 中介層：在每個請求前後記錄指標
func Metrics() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.FullPath()  // 用路由模板而非實際路徑，避免高基數問題

        httpActiveConnections.Inc()
        defer httpActiveConnections.Dec()

        c.Next()

        status := strconv.Itoa(c.Writer.Status())
        duration := time.Since(start).Seconds()

        httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
        httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
    }
}
```

**設計決策解析：**

1. **為什麼用 `c.FullPath()` 而不是 `c.Request.URL.Path`？**
   `c.Request.URL.Path` 會是 `/api/v1/articles/42`，每個文章 ID 都不同，會產生無限多的時間序列。
   `c.FullPath()` 回傳 `/api/v1/articles/:id`，所有文章查詢共用一個指標，避免 Prometheus 爆炸。

2. **為什麼 Metrics 中介層要放在最前面？**
   放在最前面才能測量到完整的請求處理時間（包括其他中介層的執行時間）。
   如果放在 Logger 後面，就量不到 Logger 花了多少時間。

3. **三種指標各自的用途是什麼？**
   - **Counter**（`http_requests_total`）：計算 QPS、錯誤率
   - **Histogram**（`http_request_duration_seconds`）：計算 P50/P99 延遲
   - **Gauge**（`http_active_connections`）：即時觀察系統負載

**完成後驗證：**
- 訪問 `/metrics` 端點，確認能看到 `http_requests_total` 等指標
- 打幾個 API 請求後再查 `/metrics`，確認數字有增加
- 確認 path label 是路由模板格式（如 `/api/v1/articles/:id`）

---

### 步驟 9：容器化部署（學完第 23 課後）

**先修課程：**
- 第 23 課 — Docker：學會 Dockerfile、多階段建置、docker-compose

**要建立的檔案：**
- `Dockerfile`
- `docker-compose.yml`

**關鍵程式碼模式：**

```dockerfile
# 多階段建置（Multi-stage Build）
# 階段 1：編譯（使用完整的 Go 映像）
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache gcc musl-dev  # SQLite 需要 CGO
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download                   # 利用 Docker 快取層
COPY . .
RUN CGO_ENABLED=1 go build -o /app/server ./cmd/server/

# 階段 2：執行（使用最小化的 Alpine 映像）
FROM alpine:3.19
RUN apk add --no-cache ca-certificates sqlite-libs
COPY --from=builder /app/server .
COPY --from=builder /app/config.yaml .
ENV SERVER_MODE=release
EXPOSE 8080
CMD ["./server"]
```

```yaml
# docker-compose.yml
services:
  blog-api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - JWT_SECRET=please-change-this-in-production
    volumes:
      - blog-data:/app/data   # 持久化 SQLite 資料
    restart: unless-stopped

volumes:
  blog-data:  # Named Volume，容器刪除後資料不會消失
```

**設計決策解析：**

1. **為什麼用多階段建置？**
   Go 編譯器 + 所有依賴套件的映像可能有好幾百 MB。多階段建置讓你在第一階段完成編譯，
   第二階段只複製編譯好的二進位檔到一個小型 Alpine 映像（最終大小約 20-30 MB）。

2. **為什麼 `COPY go.mod go.sum` 和 `COPY . .` 要分開？**
   Docker 有快取機制：如果某一行的輸入沒變，就不會重新執行。
   `go.mod` 很少改動，所以 `go mod download` 的結果可以被快取。
   只有原始碼改了才需要重新編譯，大幅加速 CI/CD 的建置速度。

3. **為什麼用 Named Volume 而不是 Bind Mount？**
   Named Volume 由 Docker 管理，效能更好、跨平台更穩定。
   Bind Mount 直接掛載主機目錄，適合開發時即時同步原始碼。

4. **為什麼設定 `restart: unless-stopped`？**
   如果容器因為意外 crash，Docker 會自動重啟它。
   只有你手動 `docker-compose stop` 才會真正停止。這是最基本的高可用策略。

**完成後驗證：**
- `docker build -t blog-api .` 能成功建置映像
- `docker-compose up -d` 能啟動服務
- `curl http://localhost:8080/healthz` 回傳成功
- 重啟容器後，之前建立的文章資料還在（Volume 持久化）

---

## 架構圖

```
                    ┌─────────────────────────────────────────┐
                    │              HTTP Request                │
                    └───────────────────┬─────────────────────┘
                                        │
                                        ▼
                    ┌─────────────────────────────────────────┐
                    │            Middleware Layer              │
                    │  ┌─────────┐ ┌────────┐ ┌───────────┐  │
                    │  │ Metrics │ │ Logger │ │ Recovery  │  │
                    │  └─────────┘ └────────┘ └───────────┘  │
                    │  ┌─────────┐ ┌──────────────────────┐  │
                    │  │  CORS   │ │     JWT Auth         │  │
                    │  └─────────┘ └──────────────────────┘  │
                    └───────────────────┬─────────────────────┘
                                        │
                                        ▼
                    ┌─────────────────────────────────────────┐
                    │            Handler Layer                 │
                    │  ┌──────────────┐ ┌──────────────────┐  │
                    │  │ UserHandler  │ │ ArticleHandler   │  │
                    │  └──────────────┘ └──────────────────┘  │
                    │  ┌──────────────┐ ┌──────────────────┐  │
                    │  │CommentHandler│ │ HealthHandler    │  │
                    │  └──────────────┘ └──────────────────┘  │
                    │           + response + error_helper      │
                    └───────────────────┬─────────────────────┘
                                        │
                                        ▼
                    ┌─────────────────────────────────────────┐
                    │            Usecase Layer                 │
                    │  ┌──────────────┐ ┌──────────────────┐  │
                    │  │UserUsecase   │ │ArticleUsecase    │  │
                    │  │ - Register   │ │ - Create         │  │
                    │  │ - Login      │ │ - Update (權限)  │  │
                    │  │ - JWT 簽發   │ │ - Delete (權限)  │  │
                    │  └──────────────┘ └──────────────────┘  │
                    │           + apperror 錯誤包裝            │
                    └───────────────────┬─────────────────────┘
                                        │
                                        ▼
                    ┌─────────────────────────────────────────┐
                    │          Repository Layer                │
                    │  ┌──────────────────────────────────┐   │
                    │  │  CachedArticleRepository         │   │
                    │  │  (裝飾器：快取 + 原始 Repo)       │   │
                    │  └──────────────┬───────────────────┘   │
                    │                 │                        │
                    │  ┌──────────────▼───────────────────┐   │
                    │  │  articleRepository (GORM)        │   │
                    │  │  userRepository (GORM)           │   │
                    │  │  commentRepository (GORM)        │   │
                    │  └──────────────────────────────────┘   │
                    └───────────────────┬─────────────────────┘
                                        │
                          ┌─────────────┼─────────────┐
                          ▼             ▼             ▼
                    ┌──────────┐ ┌──────────┐ ┌──────────┐
                    │  SQLite  │ │  Redis   │ │  Config  │
                    │ (GORM)   │ │ (Cache)  │ │ (Viper)  │
                    └──────────┘ └──────────┘ └──────────┘
```

**請求完整流程範例**（建立文章）：

```
1. POST /api/v1/articles  (帶 JWT Token)
2. → Metrics 中介層記錄開始時間
3. → Logger 中介層記錄請求資訊
4. → CORS 中介層設定跨域 Header
5. → JWT 中介層驗證 Token，提取 user_id
6. → ArticleHandler.Create 綁定 JSON、呼叫 usecase
7. → ArticleUsecase.Create 建立文章實體
8. → CachedArticleRepository.Create 直接委派給底層
9. → articleRepository.Create 用 GORM 寫入 SQLite
10. ← 逐層回傳 Article 資料
11. ← response.Created 回傳 201 JSON
12. ← Metrics 中介層記錄請求時間和狀態碼
```

---

## 每課對應的專案檔案

| 課程 | 主題 | 對應的專案檔案 | 學到的模式 |
|------|------|----------------|-----------|
| 第 4 課 | 結構體與方法 | `internal/domain/*.go` | struct 定義、方法接收器 |
| 第 6 課 | 介面 | `internal/domain/*.go` (Repository 介面) | 隱式實作、依賴反轉 |
| 第 7 課 | 錯誤處理 | `pkg/apperror/error.go` | Sentinel Error、Early Return |
| 第 10 課 | 架構設計 | 整個 `internal/` 目錄結構 | Clean Architecture 分層 |
| 第 12 課 | Gin 框架 | `internal/handler/router.go` | 路由群組、中介層掛載 |
| 第 13 課 | JSON 與 Struct Tags | `internal/domain/*.go` (Request struct) | `json` / `binding` tag |
| 第 14 課 | GORM 資料庫 | `internal/repository/*.go` | CRUD、Preload、分頁 |
| 第 16 課 | Error Wrapping | `pkg/apperror/error.go` | `fmt.Errorf %w`、`errors.Is` |
| 第 17 課 | 中介層 | `internal/middleware/*.go` | Logger、CORS、Recovery |
| 第 18 課 | JWT 認證 | `internal/middleware/jwt.go`、`internal/usecase/user_usecase.go` | Token 簽發 / 驗證 |
| 第 19 課 | 單元測試 | `internal/usecase/*_test.go`、`internal/handler/*_test.go` | 表格驅動測試、Mock |
| 第 20 課 | Config 管理 | `pkg/config/config.go`、`config.yaml` | Viper、環境變數覆蓋 |
| 第 21 課 | 結構化日誌 | `pkg/logger/logger.go`、`internal/middleware/logger.go` | slog、JSON 日誌 |
| 第 23 課 | Docker | `Dockerfile`、`docker-compose.yml` | 多階段建置、Volume |
| 第 26 課 | Redis 快取 | `pkg/cache/cache.go`、`internal/repository/article_cache_repository.go` | 裝飾器、Cache Aside |
| 第 28 課 | 資料庫進階 | `internal/repository/article_repository.go` | 交易、軟刪除、Preload |
| 第 29 課 | Clean Arch 進階 | `cmd/server/main.go` | Graceful Shutdown、Health Check |
| 第 35 課 | Prometheus | `internal/middleware/metrics.go` | Counter、Histogram、Gauge |

---

## 常見問題

### Q: 為什麼不直接用一個檔案寫完所有程式碼？

因為隨著功能增加，單一檔案會變得無法維護。分層架構讓每一層各司其職：
- **改資料庫** → 只動 `repository/`
- **改商業規則** → 只動 `usecase/`
- **改 API 格式** → 只動 `handler/`

### Q: 為什麼不用 `interface{}` 到處傳，而是定義這麼多 struct？

Go 是靜態型別語言，編譯器能在你犯錯的瞬間就告訴你。
用明確的 struct 可以在編譯期就抓到型別錯誤，而不是等到上線後才 panic。

### Q: 專案裡的 `internal/` 和 `pkg/` 差在哪裡？

- `internal/` — Go 語言機制保證：只有同一個模組內的程式碼能 import。放部落格專屬邏輯。
- `pkg/` — 沒有存取限制，理論上可以被其他專案重用。放通用工具（config、logger、cache）。

### Q: 為什麼 main.go 裡面要自己做依賴注入，不用框架？

這個專案的依賴關係夠簡單，手動注入反而最清楚。
等到依賴變得很複雜（像搶票系統 `ticket-system/`），再用 Wire 這種工具幫你自動產生注入程式碼。

### Q: 學完這個專案後下一步該做什麼？

看搶票系統（`ticket-system/`）。它在部落格 API 的基礎上加入了 Goroutine 並發、
gRPC 微服務、Message Queue、CQRS 讀寫分離、Circuit Breaker 等進階模式。
