# 第三課：函式

## 學習目標

- 學會定義和呼叫函式
- 理解多重回傳值（Go 最重要的慣例之一）
- 學會可變參數、閉包、匿名函式
- 理解 `defer` 延遲執行的用途

## 執行方式

```bash
cd tutorials/03-functions
go run main.go
```

## 重點筆記

### 多重回傳值 — Go 的錯誤處理基石

Go 沒有 try/catch，而是使用多重回傳值來處理錯誤：

```go
result, err := someFunction()
if err != nil {
    // 處理錯誤
}
// 使用 result
```

這個模式在整個部落格專案中無處不在。

### defer 的實際用途

在專案的 `internal/middleware/recovery.go` 中：

```go
func Recovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // 攔截 panic，防止伺服器崩潰
            }
        }()
        c.Next()
    }
}
```

`defer` 確保即使發生 panic，恢復邏輯也一定會執行。

### 閉包的實際用途

在專案中，中介層就是閉包的應用：

```go
func JWTAuth(cfg *config.Config) gin.HandlerFunc {
    // cfg 被閉包捕獲
    return func(c *gin.Context) {
        // 這裡可以存取外層的 cfg
    }
}
```

## 練習

1. 寫一個函式 `safeDivide(a, b int) (int, error)`，除以零時回傳錯誤
2. 寫一個 `makeGreeter(prefix string) func(string) string` 閉包
3. 用 `defer` 印出「程式結束」，確保它在 main 的最後才印出
