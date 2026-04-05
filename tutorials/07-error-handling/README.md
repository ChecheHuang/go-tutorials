# 第七課：錯誤處理（Error Handling）

> **一句話總結**：在 Go 中，錯誤是正常的回傳值，不是例外（exception）—— 就像餐廳跟你說「牛排賣完了」一樣自然。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟢 初學者 | **入門必修**：Go 的錯誤處理哲學（不用例外） |
| 🟡 中級工程師 | 自訂錯誤型別，錯誤處理的最佳實踐 |
| 🔴 資深工程師 | 搭配第 25 課 Error Wrapping，設計完整的錯誤策略 |

## 你會學到什麼？

- Go 的「錯誤是值」（errors are values）哲學
- error 介面的本質（只有一個方法！）
- 建立錯誤的三種方式：`errors.New`、`fmt.Errorf`、自訂錯誤型別
- 哨兵錯誤（Sentinel Error）模式和命名慣例 `ErrXxx`
- 自訂錯誤型別：攜帶結構化資訊
- 錯誤包裝 `%w` 與解包 `errors.Unwrap`
- `errors.Is`（比較值）vs `errors.As`（比較型別）
- `panic` vs `error` 的使用時機
- 錯誤邊界：在哪裡處理，在哪裡傳遞

## 執行方式

```bash
go run ./tutorials/07-error-handling
```

## 生活比喻：餐廳點餐

```
你在餐廳點餐：

  「我要一份牛排」
       │
       ▼
  ┌─────────────────────────────────────────────┐
  │ 服務生：「抱歉，牛排賣完了」                    │  ← error（正常的錯誤）
  │ 你：「那我改點雞排」                           │  ← 處理錯誤，選擇替代方案
  └─────────────────────────────────────────────┘

  ┌─────────────────────────────────────────────┐
  │ 服務生：「廚房著火了！全部撤離！」               │  ← panic（嚴重的意外）
  │ 你：逃跑                                     │  ← 無法正常處理
  └─────────────────────────────────────────────┘
```

- **error** = 「牛排賣完了」—— 可預期的、能處理的錯誤
- **panic** = 「廚房著火了」—— 不可預期的、嚴重的意外

## 建立錯誤的方式比較

| 方式 | 語法 | 適用場景 | 範例 |
|------|------|---------|------|
| `errors.New` | `errors.New("訊息")` | 簡單固定的錯誤（哨兵錯誤） | `var ErrNotFound = errors.New("找不到")` |
| `fmt.Errorf` | `fmt.Errorf("ID=%d 失敗", id)` | 需要包含動態變數的錯誤 | `fmt.Errorf("使用者 %d 不存在", id)` |
| `fmt.Errorf + %w` | `fmt.Errorf("失敗: %w", err)` | 包裝底層錯誤，保留錯誤鏈 | `fmt.Errorf("查詢失敗: %w", err)` |
| 自訂錯誤型別 | `&MyError{Field: "age"}` | 需要攜帶結構化的額外資訊 | `&ValidationError{Field: "age"}` |

## 哨兵錯誤（Sentinel Error）

「哨兵」就像站崗的衛兵 —— 是一個**固定的、已知的**錯誤值。

### 命名慣例

```go
// 永遠以 Err 開頭，後面接描述
var ErrNotFound     = errors.New("找不到資源")
var ErrUnauthorized = errors.New("未經授權")
var ErrTimeout      = errors.New("操作逾時")
var ErrInvalidInput = errors.New("輸入不合法")
```

### 使用時機

```
什麼時候用哨兵錯誤？
  ✅ 錯誤是「可預期的」且「全域統一的」
  ✅ 呼叫方需要用 errors.Is() 判斷錯誤類型
  ✅ 多個函式可能回傳同一種錯誤

什麼時候不用？
  ❌ 錯誤需要攜帶動態資訊 → 用 fmt.Errorf 或自訂型別
  ❌ 錯誤只在一個函式內使用 → 直接回傳即可
```

## 自訂錯誤型別

當你需要在錯誤中攜帶「更多資訊」時：

```go
// 定義自訂錯誤型別
type ValidationError struct {
    Field   string  // 哪個欄位出錯
    Message string  // 具體的錯誤訊息
}

// 實作 error 介面（只需要一個方法！）
func (e *ValidationError) Error() string {
    return fmt.Sprintf("驗證錯誤 [%s]: %s", e.Field, e.Message)
}

// 使用
return &ValidationError{Field: "age", Message: "不能為負數"}
```

## 錯誤包裝（%w）與解包

```
  原始錯誤                     包裝後的錯誤
┌──────────────┐           ┌──────────────────────────────────┐
│ 資料庫連線失敗 │  ── %w ──▶ │ 取得使用者失敗: 資料庫連線失敗     │
└──────────────┘           └──────────────────────────────────┘
                                       │
                              errors.Unwrap()
                                       │
                                       ▼
                           ┌──────────────────┐
                           │ 資料庫連線失敗      │ ← 取回原始錯誤
                           └──────────────────┘
```

```go
// 包裝：用 %w 動詞
return fmt.Errorf("取得使用者失敗: %w", originalErr)

// 解包：取回原始錯誤
inner := errors.Unwrap(wrappedErr)
```

## errors.Is vs errors.As

| 函式 | 用途 | 比較什麼 | 類比 |
|------|------|---------|------|
| `errors.Is(err, target)` | 是不是「這個」錯誤？ | 比較**值** | 「你是不是小明？」 |
| `errors.As(err, &target)` | 是不是「這種」錯誤？ | 比較**型別** | 「你是不是學生？」 |

### errors.Is —— 比較錯誤值

```go
var ErrNotFound = errors.New("找不到")

_, err := findUser(999)
if errors.Is(err, ErrNotFound) {  // 即使 err 被包裝過，也能穿透找到！
    // 回傳 HTTP 404
}
```

### errors.As —— 取出錯誤型別

```go
var valErr *ValidationError
if errors.As(err, &valErr) {     // 從錯誤鏈中取出 *ValidationError
    fmt.Println(valErr.Field)    // 可以存取額外的結構化資訊
    fmt.Println(valErr.Message)
}
```

## panic vs error 決策指南

```
遇到錯誤時，問自己：

  「程式還能繼續正常運作嗎？」
       │
       ├── 可以 → 用 error（99% 的情況）
       │   例如：檔案不存在、網路逾時、使用者輸入錯誤、查詢無結果
       │
       └── 不行 → 用 panic（1% 的情況）
           例如：設定檔損壞（無法啟動）、程式邏輯 bug（不應該發生的情況）
```

**經驗法則**：如果你不確定，就用 `error`。用 `panic` 之前先想三次。

## 在部落格專案中的應用

### Usecase 層的錯誤處理

```go
// internal/usecase/user_usecase.go
func (u *userUsecase) Login(req domain.LoginRequest) (*domain.LoginResponse, error) {
    user, err := u.userRepo.FindByEmail(req.Email)
    if err != nil {
        return nil, errors.New("信箱或密碼錯誤")  // 不洩漏「是信箱不存在」
    }

    if err := bcrypt.CompareHashAndPassword(...); err != nil {
        return nil, errors.New("信箱或密碼錯誤")  // 統一錯誤訊息（安全考量）
    }
    // ...
}
```

### Middleware 用 recover 攔截 panic

```go
// internal/middleware/recovery.go
defer func() {
    if err := recover(); err != nil {
        // 攔截 panic，防止整個伺服器崩潰
        // 回傳 HTTP 500 Internal Server Error
    }
}()
```

### Handler 層是錯誤邊界

```go
// internal/handler/user_handler.go
func (h *userHandler) GetUser(c *gin.Context) {
    user, err := h.usecase.GetUserByID(id)
    if err != nil {
        // 這裡是「錯誤邊界」—— 把 error 轉換成 HTTP 回應
        c.JSON(http.StatusNotFound, gin.H{"error": "使用者不存在"})
        return
    }
    c.JSON(http.StatusOK, user)
}
```

## 常見問題 (FAQ)

### Q: 為什麼 Go 不用 try/catch？

Go 的設計者認為 try/catch 容易讓人忽略錯誤（隨便 catch 一下就算了）。用回傳值強迫你面對每一個可能的錯誤，讓程式更可靠。

### Q: 到處寫 `if err != nil` 不會很煩嗎？

確實比較冗長，但好處是：
1. 每個錯誤都「看得見」—— 不會有隱藏的例外
2. 錯誤處理邏輯就在出錯的地方 —— 不用跳到 catch 區塊
3. 控制流程非常清楚 —— 讀程式碼時一目了然

### Q: `%w` 和 `%v` 有什麼差別？

```go
// %w（wrap）：包裝錯誤，保留錯誤鏈（可以用 errors.Is / errors.As 穿透）
fmt.Errorf("操作失敗: %w", err)

// %v（value）：只是把錯誤訊息嵌入字串（錯誤鏈斷裂！無法穿透）
fmt.Errorf("操作失敗: %v", err)
```

**建議**：永遠用 `%w`，除非你刻意想要隱藏底層錯誤。

### Q: 什麼時候該用自訂錯誤型別而不是哨兵錯誤？

- **哨兵錯誤**：只需要知道「是什麼錯」，不需要額外資訊
- **自訂型別**：需要知道「哪裡錯」「錯了什麼」等結構化資訊

```go
// 哨兵錯誤：只知道「找不到」
var ErrNotFound = errors.New("找不到資源")

// 自訂型別：知道「哪個欄位」「什麼問題」
type ValidationError struct {
    Field   string
    Message string
}
```

## 練習

1. **基礎**：寫一個 `Divide(a, b float64) (float64, error)` 函式，當 `b == 0` 時回傳錯誤
2. **哨兵錯誤**：定義 `var ErrDivideByZero = errors.New("除以零")` 並在 Divide 中使用它
3. **自訂型別**：定義 `NotFoundError` 自訂錯誤型別，包含 `Resource` 和 `ID` 欄位，用 `errors.As` 取出資訊
4. **錯誤包裝**：寫一個呼叫 Divide 的函式，用 `%w` 包裝錯誤，然後用 `errors.Is` 驗證能穿透包裝
5. **思考**：在部落格 API 中，登入失敗時為什麼不分開回傳「信箱不存在」和「密碼錯誤」？（提示：安全性）

## 下一課預告

**第八課：套件與模組（Packages & Modules）** —— 當程式越來越大，你需要把程式碼拆分到不同的檔案和資料夾。Go 的模組系統讓你可以組織程式碼、管理依賴、和全世界分享你的套件。
