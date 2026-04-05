# 第八課：套件與模組

## 學習目標

- 理解 Go Module（`go.mod`）的作用
- 學會建立和匯入自己的套件
- 了解套件的可見性規則（大小寫）
- 理解 `internal` 目錄的特殊性

## 動手做

這一課沒有單一的 `main.go`，而是一個完整的小專案。請按照下面的步驟操作。

### 步驟 1：建立新專案

```bash
mkdir myproject && cd myproject
go mod init myproject
```

`go mod init` 會建立 `go.mod` 檔案，這是 Go Module 的核心：

```
module myproject    ← 模組名稱
go 1.21             ← Go 版本
```

### 步驟 2：建立自己的套件

```bash
mkdir -p math/calculator
```

建立 `math/calculator/calculator.go`：

```go
// package 名稱通常和目錄名稱相同
package calculator

// Add 是匯出的函式（大寫開頭 = public）
func Add(a, b int) int {
    return a + b
}

// Multiply 也是匯出的
func Multiply(a, b int) int {
    return a * b
}

// helper 是未匯出的函式（小寫開頭 = private）
// 只能在 calculator 套件內部使用
func helper() string {
    return "我是內部函式"
}
```

### 步驟 3：在 main 中匯入

建立 `main.go`：

```go
package main

import (
    "fmt"
    "myproject/math/calculator"  // 匯入自己的套件
)

func main() {
    result := calculator.Add(3, 5)
    fmt.Println("3 + 5 =", result)

    // calculator.helper()  // ← 編譯錯誤！小寫開頭無法存取
}
```

### 步驟 4：執行

```bash
go run main.go
```

## 重點筆記

### 可見性規則（超重要！）

Go 使用命名慣例來控制可見性，**不需要** `public` / `private` 關鍵字：

| 命名 | 可見性 | 範例 |
|------|--------|------|
| **大寫**開頭 | 匯出的（public） | `User`、`Create`、`ErrNotFound` |
| **小寫**開頭 | 未匯出的（private） | `userRepository`、`helper` |

### 在部落格專案中的體現

```go
// domain/user.go
type User struct { ... }           // 大寫 → 其他套件可以使用
type RegisterRequest struct { ... } // 大寫 → handler 可以使用
type UserRepository interface { ... } // 大寫 → repository 可以實作

// repository/user_repository.go
type userRepository struct { ... }  // 小寫 → 外部不能直接建立
func NewUserRepository(db *gorm.DB) domain.UserRepository { ... }
// ↑ 大寫 → 外部透過這個函式取得實例
// 回傳介面型別，隱藏具體實作
```

### `internal` 目錄

Go 語言規定 `internal/` 目錄下的套件只能被同一模組的程式碼匯入：

```
blog-api/
├── internal/          ← 只有 blog-api 模組能匯入
│   ├── domain/
│   ├── handler/
│   └── repository/
├── pkg/               ← 任何外部專案都能匯入
│   ├── config/
│   └── response/
└── cmd/server/main.go ← 可以匯入 internal 和 pkg
```

### 安裝第三方套件

```bash
# 安裝單一套件
go get github.com/gin-gonic/gin

# 安裝所有缺少的依賴
go mod tidy

# 查看所有依賴
go list -m all
```

安裝後會更新兩個檔案：
- `go.mod`：記錄依賴和版本
- `go.sum`：記錄依賴的雜湊值（確保完整性）

### 匯入的排列慣例

```go
import (
    // 第一組：標準庫
    "fmt"
    "net/http"

    // 第二組：自己的套件
    "blog-api/internal/domain"
    "blog-api/pkg/config"

    // 第三組：第三方套件
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)
```

## 練習

1. 建立一個新的 Go 模組，包含 `stringutil` 套件，提供 `Reverse(s string) string` 函式
2. 試試把函式改成小寫開頭，觀察匯入時的錯誤訊息
3. 用 `go get` 安裝 `github.com/fatih/color` 套件，在 main 中使用它
