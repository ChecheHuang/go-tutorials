# 第七課：錯誤處理

## 學習目標

- 理解 Go「沒有 try/catch」的設計哲學
- 學會 `error` 介面與多重回傳值的錯誤處理模式
- 掌握自訂錯誤、錯誤包裝、`errors.Is` / `errors.As`
- 了解 `panic` / `recover` 的適用時機

## 執行方式

```bash
cd tutorials/07-error-handling
go run main.go
```

## 重點筆記

### Go 錯誤處理的核心模式

```go
result, err := someFunction()
if err != nil {
    return fmt.Errorf("操作失敗: %w", err) // 包裝並向上傳遞
}
// 使用 result
```

這個模式在專案中出現了幾十次。

### errors.Is vs errors.As

| 函式 | 用途 | 比較什麼 |
|------|------|---------|
| `errors.Is(err, target)` | 判斷是不是某個特定錯誤 | 比較值 |
| `errors.As(err, &target)` | 取出錯誤中的特定型別 | 比較型別 |

### 在專案中的應用

`internal/usecase/user_usecase.go` 中的登入流程：
```go
func (u *userUsecase) Login(req domain.LoginRequest) (*domain.LoginResponse, error) {
    user, err := u.userRepo.FindByEmail(req.Email)
    if err != nil {
        return nil, errors.New("信箱或密碼錯誤")  // 不洩漏具體原因
    }

    if err := bcrypt.CompareHashAndPassword(...); err != nil {
        return nil, errors.New("信箱或密碼錯誤")  // 統一錯誤訊息
    }
    // ...
}
```

`internal/middleware/recovery.go` 中用 `recover` 攔截 panic：
```go
defer func() {
    if err := recover(); err != nil {
        // 攔截 panic，防止整個伺服器崩潰
    }
}()
```

## 練習

1. 寫一個 `Divide(a, b float64) (float64, error)` 函式
2. 定義 `NotFoundError` 自訂錯誤型別，包含 `Resource` 和 `ID` 欄位
3. 用 `errors.As` 判斷錯誤是否為 `NotFoundError` 並取出 Resource 名稱
