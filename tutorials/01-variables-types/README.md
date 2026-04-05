# 第一課：變數與型別

## 學習目標

- 了解 Go 的基本資料型別
- 學會三種變數宣告方式
- 理解零值（Zero Value）的概念
- 學會使用 `fmt` 套件輸出資料

## 執行方式

```bash
cd tutorials/01-variables-types
go run main.go
```

## 重點筆記

### Go 的基本型別

| 型別 | 說明 | 範例 | 零值 |
|------|------|------|------|
| `string` | 字串 | `"hello"` | `""` |
| `int` | 整數 | `42` | `0` |
| `float64` | 浮點數 | `3.14` | `0.0` |
| `bool` | 布林 | `true` / `false` | `false` |

### 三種宣告方式

```go
// 1. var + 型別（最完整）
var name string = "Alice"

// 2. var + 自動推導（省略型別）
var name = "Alice"

// 3. := 短變數宣告（最常用，只能在函式內使用）
name := "Alice"
```

### 在專案中的應用

在部落格專案的 `internal/domain/user.go` 中：

```go
type User struct {
    ID       uint   // uint 是無號整數（不會是負數）
    Username string // 字串
    Email    string
    Password string
}
```

每個欄位都有明確的型別，這讓 Go 在編譯時就能抓到型別錯誤。

## 練習

1. 宣告一個變數 `temperature`，型別為 `float64`，值為 `36.5`
2. 使用 `fmt.Printf` 印出：「目前體溫：36.5 度」
3. 試試看把 `int` 直接賦值給 `float64` 變數，觀察錯誤訊息
