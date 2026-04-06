# 第十六課：Error Wrapping 完整示範

> **一句話總結**：用 `%w` 包裝錯誤，保留完整錯誤鏈；用 `errors.Is/As` 穿透任意層數判斷錯誤類型。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | **重點**：%w 包裝、errors.Is、errors.As 是必備技能 |
| 🔴 資深工程師 | 設計完整的錯誤體系，包含 AppError、多重錯誤 |

## 你會學到什麼？

- `fmt.Errorf("%w")` 包裝錯誤的語法與原理
- `%w` 和 `%v` 的關鍵差異以及各自的使用時機
- `errors.Is` 如何穿透多層包裝比對錯誤值
- `errors.As` 如何從錯誤鏈中提取特定型別
- Clean Architecture 中錯誤從 DB → Repository → Usecase → Handler 的流向
- 實作 `Unwrap()` 讓自訂錯誤型別支援錯誤鏈
- 什麼時候該包裝、什麼時候不該包裝的判斷準則

## 執行方式

```bash
go run ./tutorials/16-error-wrapping
```

## 生活比喻：包裹追蹤單

```
想像你從淘寶買了一個東西，結果包裹出問題了。

沒有追蹤單的世界：
  客服：「您的包裹出問題了。」
  你：「什麼問題？在哪裡出的？」
  客服：「不知道。」

有追蹤單的世界：
  客服：「您的包裹出問題了。」
  追蹤紀錄：
    [配送站] 配送失敗：地址不存在
      └── [分揀中心] 轉運異常：標籤模糊
           └── [倉庫] 出貨錯誤：商品缺貨

每一站都在追蹤單上加了一行「我這邊發生了什麼」，
最後你看到的是完整的錯誤路徑。

Error Wrapping 就是這個追蹤單——
每一層函式用 %w 包裝，就像在追蹤單上加一行紀錄。
```

## 為什麼 Error Wrapping 很重要？

### 一個真實的除錯故事

假設你在生產環境看到這個錯誤日誌：

```
ERROR: query failed
```

你能知道什麼？幾乎什麼都不知道。哪個查詢？在哪個函式？為什麼失敗？

現在換成這個：

```
ERROR: handler.GetArticle(id=42) → usecase.GetArticle → repo.FindByID: query failed: connection refused (host=db:5432)
```

現在你馬上知道：
1. 是 `GetArticle` API 出問題
2. 錯誤從 Repository 層的 `FindByID` 發起
3. 根本原因是資料庫連線被拒絕
4. 連接的是 `db:5432` 這個主機

**這就是 Error Wrapping 的價值：讓錯誤攜帶完整的上下文鏈路。**

## `%w` vs `%v` 深入比較

這是 Error Wrapping 最核心也最容易搞混的概念。

### `%w`：保留錯誤鏈（Wrap）

```go
var ErrNotFound = errors.New("not found")

func findUser(id int) error {
    return ErrNotFound
}

func getUser(id int) error {
    err := findUser(id)
    if err != nil {
        // 用 %w 包裝——保留了 ErrNotFound 在鏈中
        return fmt.Errorf("getUser(id=%d): %w", id, err)
    }
    return nil
}

func main() {
    err := getUser(42)
    fmt.Println(err)
    // 輸出：getUser(id=42): not found

    fmt.Println(errors.Is(err, ErrNotFound))
    // 輸出：true ← 可以穿透包裝找到 ErrNotFound
}
```

### `%v`：斷開錯誤鏈（Format only）

```go
var ErrNotFound = errors.New("not found")

func findUser(id int) error {
    return ErrNotFound
}

func getUser(id int) error {
    err := findUser(id)
    if err != nil {
        // 用 %v 格式化——只保留了錯誤「文字」，斷開了鏈
        return fmt.Errorf("getUser(id=%d): %v", id, err)
    }
    return nil
}

func main() {
    err := getUser(42)
    fmt.Println(err)
    // 輸出：getUser(id=42): not found  ← 看起來一模一樣！

    fmt.Println(errors.Is(err, ErrNotFound))
    // 輸出：false ← 無法穿透！鏈已斷開
}
```

### 差異對照表

| 特性 | `%w`（Wrap） | `%v`（Format） |
|------|-------------|----------------|
| 保留原始錯誤的 reference | 是 | 否 |
| `errors.Is` 可以穿透 | 是 | 否 |
| `errors.As` 可以穿透 | 是 | 否 |
| 輸出的錯誤訊息 | 一樣 | 一樣 |
| 適用場景 | 呼叫者需要判斷原始錯誤 | 故意隱藏實作細節 |

### 什麼時候用 `%v` 而不是 `%w`？

```go
// 場景：你不希望呼叫者依賴你的內部實作
// 例如：你的套件內部用了 MySQL，但不想暴露 mysql.ErrNoRows
func (r *UserRepo) FindByEmail(email string) (*User, error) {
    var user User
    err := r.db.Where("email = ?", email).First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        // 轉換為自己的 domain error，用 %w
        return nil, fmt.Errorf("FindByEmail(%s): %w", email, ErrUserNotFound)
    }
    if err != nil {
        // 不想暴露 GORM 的內部錯誤型別給上層，用 %v
        return nil, fmt.Errorf("FindByEmail(%s): database error: %v", email, err)
    }
    return &user, nil
}
```

## `errors.Is`：值比較穿透錯誤鏈

`errors.Is` 會沿著錯誤鏈一層一層往內找，直到找到匹配的錯誤值。

```go
var (
    ErrNotFound    = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
)

// 三層包裝
err1 := fmt.Errorf("handler: %w",
    fmt.Errorf("usecase: %w",
        fmt.Errorf("repo: %w", ErrNotFound)))

errors.Is(err1, ErrNotFound)     // true  — 穿透 3 層找到了
errors.Is(err1, ErrUnauthorized) // false — 鏈中沒有這個錯誤
```

### 錯誤鏈展開過程

```
errors.Is(err1, ErrNotFound)

第 1 層：「handler: usecase: repo: not found」 == ErrNotFound?  → 否
  ↓ Unwrap()
第 2 層：「usecase: repo: not found」 == ErrNotFound?  → 否
  ↓ Unwrap()
第 3 層：「repo: not found」 == ErrNotFound?  → 否
  ↓ Unwrap()
第 4 層：「not found」 == ErrNotFound?  → 是！回傳 true
```

### 自訂錯誤型別支援 `errors.Is`

```go
type AppError struct {
    Code    int
    Message string
    Err     error  // 內層錯誤
}

func (e *AppError) Error() string {
    return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
}

// 實作 Unwrap() 讓 errors.Is 可以穿透
func (e *AppError) Unwrap() error {
    return e.Err
}

// 使用
appErr := &AppError{Code: 404, Message: "article not found", Err: ErrNotFound}
wrapped := fmt.Errorf("handler: %w", appErr)

errors.Is(wrapped, ErrNotFound) // true — 穿透 AppError 再找到 ErrNotFound
```

## `errors.As`：型別提取穿透錯誤鏈

`errors.As` 沿著錯誤鏈找到第一個符合目標型別的錯誤，並把它的值填入目標變數。

```go
type AppError struct {
    Code    int
    Message string
    Err     error
}

func (e *AppError) Error() string {
    return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// 使用
original := &AppError{Code: 404, Message: "article not found", Err: ErrNotFound}
wrapped := fmt.Errorf("handler: %w", original)

var appErr *AppError
if errors.As(wrapped, &appErr) {
    fmt.Println(appErr.Code)    // 404
    fmt.Println(appErr.Message) // article not found
}
```

### `errors.Is` vs `errors.As` 對照

| 特性 | `errors.Is` | `errors.As` |
|------|-------------|-------------|
| 比對方式 | 值相等（`==`） | 型別匹配 |
| 常見用途 | 判斷是否為特定 sentinel error | 取出結構化錯誤的欄位 |
| 搭配使用 | `var ErrNotFound = errors.New(...)` | `type AppError struct{...}` |
| 目標 | `errors.Is(err, ErrNotFound)` | `errors.As(err, &appErr)` |
| 回傳值 | `bool` | `bool`（且填入目標變數） |

## Clean Architecture 中的錯誤流向

在分層架構中，錯誤從最底層（資料庫）往上流動，每一層都加上自己的上下文：

```
┌─────────────────────────────────────────────────────────────┐
│                        Handler 層                            │
│                                                              │
│  收到錯誤：                                                    │
│  "GetArticle(id=42): getArticle: findByID: record not found" │
│                                                              │
│  判斷：errors.Is(err, ErrNotFound) → true → 回 404           │
│        errors.Is(err, ErrForbidden) → true → 回 403          │
│        其他 → 回 500                                          │
└──────────────────────────┬──────────────────────────────────┘
                           │ ▲
          fmt.Errorf(      │ │    errors.Is / errors.As
          "GetArticle: %w")│ │
                           ▼ │
┌─────────────────────────────────────────────────────────────┐
│                       Usecase 層                             │
│                                                              │
│  收到錯誤："findByID: record not found"                       │
│  包裝：fmt.Errorf("getArticle: %w", err)                     │
│                                                              │
│  可以在這裡做業務判斷：                                         │
│  if errors.Is(err, ErrNotFound) {                            │
│      return nil, fmt.Errorf("getArticle: %w", ErrNotFound)   │
│  }                                                           │
└──────────────────────────┬──────────────────────────────────┘
                           │ ▲
          fmt.Errorf(      │ │
          "findByID: %w")  │ │
                           ▼ │
┌─────────────────────────────────────────────────────────────┐
│                     Repository 層                            │
│                                                              │
│  收到錯誤：gorm.ErrRecordNotFound                             │
│  轉換為 domain error：                                        │
│  if errors.Is(err, gorm.ErrRecordNotFound) {                 │
│      return fmt.Errorf("findByID: %w", ErrNotFound)          │
│  }                                                           │
└──────────────────────────┬──────────────────────────────────┘
                           │ ▲
                           ▼ │
┌─────────────────────────────────────────────────────────────┐
│                      Database（GORM）                        │
│                                                              │
│  gorm.ErrRecordNotFound                                      │
│  或其他 SQL 錯誤                                               │
└─────────────────────────────────────────────────────────────┘
```

### 關鍵原則

Repository 層是**錯誤轉換的邊界**：

```go
// repository/article.go
func (r *articleRepo) FindByID(id uint) (*domain.Article, error) {
    var article domain.Article
    err := r.db.First(&article, id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // 把 GORM 錯誤轉換為 domain 錯誤
            return nil, fmt.Errorf("findByID(id=%d): %w", id, domain.ErrNotFound)
        }
        // 未知的資料庫錯誤，不用 %w 避免洩漏 GORM 實作
        return nil, fmt.Errorf("findByID(id=%d): database error: %v", id, err)
    }
    return &article, nil
}
```

Usecase 層加上**業務上下文**：

```go
// usecase/article.go
func (uc *articleUsecase) GetArticle(id uint) (*domain.Article, error) {
    article, err := uc.repo.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("getArticle: %w", err)
    }
    return article, nil
}
```

Handler 層做**最終判斷和回應**：

```go
// handler/article.go
func (h *articleHandler) GetArticle(c *gin.Context) {
    id := parseID(c.Param("id"))
    article, err := h.usecase.GetArticle(id)
    if err != nil {
        if errors.Is(err, domain.ErrNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
            return
        }
        // 記錄完整錯誤鏈到日誌（只在最頂層記錄！）
        log.Printf("GetArticle failed: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusOK, article)
}
```

## 最佳實踐

| 做法 | 正確 | 錯誤 |
|------|------|------|
| 包裝錯誤 | `fmt.Errorf("funcName: %w", err)` | `fmt.Errorf("error: %v", err)` |
| 判斷錯誤 | `errors.Is(err, ErrFoo)` | `err.Error() == "foo"` |
| 自訂型別 | 實作 `Unwrap() error` | 不實作（errors.Is 無法穿透）|
| 記錄日誌 | 只在最頂層記錄一次 | 每層都 log（日誌重複爆量）|
| Repository 層 | 把第三方錯誤轉換為 domain error | 直接往上丟 gorm 錯誤 |
| 跨套件邊界 | 用 `%v` 隱藏內部實作 | 用 `%w` 暴露內部依賴型別 |

### 何時該包裝？何時不該？

| 情境 | 建議 | 原因 |
|------|------|------|
| 呼叫下一層函式失敗 | `%w` 包裝 | 保留錯誤鏈方便上層判斷 |
| 轉換為 domain error | `%w` 包裝新的 domain error | 統一錯誤語義 |
| 跨套件邊界的第三方錯誤 | `%v` 格式化 | 避免上層依賴你的實作細節 |
| 錯誤已經有足夠上下文 | 直接 `return err` | 避免冗餘包裝 |
| 需要新增函式名/參數等上下文 | `%w` 包裝 | 方便除錯 |

## 常見錯誤：過度包裝

每一層都無腦包裝，會導致錯誤訊息又臭又長：

```
// 過度包裝（每層都加了函式名但沒有新資訊）
handler.GetArticle: usecase.GetArticle: repo.FindByID: db.First: sql: no rows in result set

// 適度包裝（每層加上有意義的上下文）
GetArticle(id=42): article not found
```

### 原則：每層只加「新資訊」

```go
// Repository：加上「是什麼操作」和「轉換為 domain error」
return fmt.Errorf("findByID(id=%d): %w", id, ErrNotFound)

// Usecase：如果沒有新資訊，可以直接 return
return nil, err  // 不需要再包裝

// Usecase：如果有額外業務上下文，才包裝
return nil, fmt.Errorf("getArticle for user %d: %w", userID, err)

// Handler：不包裝，直接判斷和回應
if errors.Is(err, ErrNotFound) {
    c.JSON(404, gin.H{"error": "not found"})
}
```

## 部落格專案實戰範例

### 定義 domain errors

```go
// domain/errors.go
package domain

import "errors"

var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrForbidden    = errors.New("forbidden")
    ErrConflict     = errors.New("conflict")
)
```

### Repository 層：轉換 GORM 錯誤

```go
// repository/article.go
func (r *articleRepo) Create(article *domain.Article) error {
    err := r.db.Create(article).Error
    if err != nil {
        // 檢查是否是唯一約束違反（例如 slug 重複）
        if strings.Contains(err.Error(), "duplicate") {
            return fmt.Errorf("create article(slug=%s): %w", article.Slug, domain.ErrConflict)
        }
        return fmt.Errorf("create article: database error: %v", err)
    }
    return nil
}

func (r *articleRepo) FindBySlug(slug string) (*domain.Article, error) {
    var article domain.Article
    err := r.db.Where("slug = ?", slug).First(&article).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("findBySlug(%s): %w", slug, domain.ErrNotFound)
        }
        return nil, fmt.Errorf("findBySlug(%s): %v", slug, err)
    }
    return &article, nil
}
```

### Handler 層：統一錯誤回應

```go
// handler/error.go
func handleError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, domain.ErrNotFound):
        c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
    case errors.Is(err, domain.ErrUnauthorized):
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
    case errors.Is(err, domain.ErrForbidden):
        c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
    case errors.Is(err, domain.ErrConflict):
        c.JSON(http.StatusConflict, gin.H{"error": "resource already exists"})
    default:
        log.Printf("unexpected error: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
    }
}
```

## FAQ

### Q1：`%w` 是什麼時候加入 Go 的？

Go 1.13（2019 年 9 月）。在此之前，社群普遍使用 `github.com/pkg/errors` 套件來做 Error Wrapping。Go 1.13 之後官方原生支援，不再需要第三方套件。如果你的專案還在用 `pkg/errors`，建議逐步遷移到標準庫。

### Q2：一個錯誤可以被 `%w` 包裝多次嗎？

可以。Go 1.20 之後，`fmt.Errorf` 甚至支援多個 `%w`，建立多重錯誤鏈。例如 `fmt.Errorf("both failed: %w and %w", err1, err2)`。在這種情況下，`errors.Is` 會檢查所有分支。不過在大部分業務場景中，單一 `%w` 就夠了。

### Q3：`errors.Is` 和 `==` 有什麼差別？

`==` 只比較最外層的錯誤值，無法穿透包裝。`errors.Is` 會遍歷整條錯誤鏈（透過 `Unwrap()`），只要鏈中任何一層匹配就回傳 `true`。另外，`errors.Is` 還支援自訂的比較邏輯——如果錯誤型別實作了 `Is(target error) bool` 方法，會優先使用該方法。

### Q4：什麼時候用 sentinel error（`var ErrNotFound = errors.New(...)`），什麼時候用自訂型別（`type AppError struct{...}`）？

Sentinel error 適合簡單的分類判斷（「是不是找不到？」），用 `errors.Is` 檢查。自訂型別適合需要攜帶額外資訊的場景（HTTP 狀態碼、錯誤代碼、使用者可見訊息），用 `errors.As` 提取。很多專案會兩者混用：自訂型別的 `Err` 欄位存放 sentinel error，這樣 `errors.Is` 和 `errors.As` 都能用。

### Q5：為什麼不建議在每一層都 `log.Printf` 錯誤？

因為錯誤會往上傳遞，如果每一層都記錄日誌，同一個錯誤會被記錄 3-4 次，日誌量暴增且難以追蹤。正確的做法是：**只在最頂層（通常是 Handler 或 main 函式）記錄一次完整的錯誤鏈**。中間層只負責包裝和傳遞。

## 練習

1. 寫一個三層函式呼叫（handler → usecase → repository），每層用 `%w` 包裝錯誤，最後在 handler 層用 `errors.Is` 判斷是否為 `ErrNotFound`
2. 建立一個 `AppError` 自訂錯誤型別，包含 `Code`（HTTP 狀態碼）和 `Message`，用 `errors.As` 取出
3. 寫一個函式鏈，故意用 `%v` 包裝錯誤，驗證 `errors.Is` 無法穿透（理解 `%w` vs `%v` 的差異）
4. 在部落格專案中找到所有使用 `fmt.Errorf` 的地方，判斷哪些應該用 `%w`、哪些不應該
5. 實作一個 `handleError` 函式，接收 `error` 參數，根據 `errors.Is` 和 `errors.As` 的結果回傳不同的 HTTP 狀態碼（404、403、409、500），並寫測試驗證每個分支

## 下一課預告

下一課我們會學習**中介層（Middleware）**——如何把日誌、驗證、跨域等「每個請求都需要做的事」抽出來，不用在每個 Handler 重複寫。Error Wrapping 在 Middleware 中也很有用，例如 Recovery 中介層需要捕捉 panic 並回傳適當的錯誤。
