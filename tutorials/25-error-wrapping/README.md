# 第二十五課：Error Wrapping 完整示範

> **一句話總結**：用 `%w` 包裝錯誤，保留完整錯誤鏈；用 `errors.Is/As` 穿透任意層數判斷錯誤類型。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | **重點**：%w 包裝、errors.Is、errors.As 是必備技能 |
| 🔴 資深工程師 | 設計完整的錯誤體系，包含 AppError、多重錯誤 |

## 執行方式

```bash
go run ./tutorials/25-error-wrapping
```

## 核心概念

```go
// 包裝（每層加上下文）
return fmt.Errorf("createUser(email=%s): %w", email, err)

// 判斷（沿整條鏈向內找）
errors.Is(err, ErrNotFound)   // true，即使包了 10 層

// 取出（找到特定類型）
var appErr *AppError
if errors.As(err, &appErr) {
    fmt.Println(appErr.Code, appErr.Status)
}
```

## 錯誤流向

```
DB 錯誤（gorm.ErrRecordNotFound）
    ↓ Repository 包裝
領域錯誤（ErrArticleNotFound）
    ↓ Usecase 加上下文
業務錯誤（"getArticle: 文章 42 不存在: 文章不存在"）
    ↓ Handler 判斷
HTTP 回應（404 Not Found）
```

## 最佳實踐

| 做法 | 正確 | 錯誤 |
|------|------|------|
| 包裝 | `fmt.Errorf("...: %w", err)` | `fmt.Errorf("...: %v", err)` |
| 判斷 | `errors.Is(err, ErrFoo)` | `err.Error() == "foo"` |
| 自訂型別 | 實作 `Unwrap() error` | 不實作（errors.Is 無法穿透）|
| 記錄日誌 | 只在最頂層記錄一次 | 每層都 log（重複）|
