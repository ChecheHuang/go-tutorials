# 第八課：套件與模組（Packages & Modules）

> **一句話總結**：套件（Package）就像「工具箱」——把相關的工具放在同一個箱子裡；模組（Module）就像「整間工廠」——一間工廠裡有很多不同的工具箱。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟢 初學者 | **入門必修**：Go modules、import、套件組織 |
| 🟡 中級工程師 | 了解 go.mod、版本管理、private module |

## 你會學到什麼？

- 套件（Package）是什麼——Go 組織程式碼的方式
- 模組（Module）是什麼——`go.mod` 的作用
- 公開 vs 私有——大寫/小寫命名規則
- `import` 語句的三種寫法
- 標準庫常用套件（fmt、strings、strconv、time、math）
- 第三方套件是什麼（`go get` 安裝）
- 部落格專案的套件結構（internal/、pkg/、cmd/）

## 執行方式

```bash
go run ./tutorials/08-packages-modules
```

## 用生活來理解

### 套件 = 公司裡的部門

想像一間公司：

```
公司（Module = 模組）
├── 行銷部（Package marketing）
│   ├── 做廣告（公開函式 = 大寫開頭）
│   └── 內部報表（私有函式 = 小寫開頭）
├── 工程部（Package engineering）
│   ├── 做產品（公開函式）
│   └── 程式碼審查（私有函式）
└── 人事部（Package hr）
    ├── 招募員工（公開函式）
    └── 薪資計算（私有函式）
```

- **每個部門**就是一個套件（Package），負責一組相關的工作
- **整間公司**就是一個模組（Module），由 `go.mod` 定義
- **公開函式**就像「對外服務」——其他部門可以來找你幫忙
- **私有函式**就像「內部流程」——只有自己部門的人知道怎麼做

### 為什麼需要套件？

如果所有程式碼都寫在同一個檔案裡：
- 檔案會越來越大，難以維護
- 不同功能混在一起，找東西很痛苦
- 多人協作時容易衝突

用套件把程式碼分類，就像把文件分門別類放進不同的資料夾。

## 公開 vs 私有命名規則（超重要！）

Go 使用「名稱的第一個字母」來決定可見性，**不需要** `public` / `private` 關鍵字：

| 命名方式 | 可見性 | 說明 | 範例 |
|---------|--------|------|------|
| **大寫**開頭 | 公開（Exported） | 任何套件都能使用 | `User`、`Create`、`NewServer` |
| **小寫**開頭 | 私有（Unexported） | 只有同一個套件能使用 | `userRepo`、`validate`、`helper` |

```go
// domain/user.go 套件中

type User struct { ... }              // ✅ 大寫 → 其他套件可以使用
type userValidator struct { ... }     // ❌ 小寫 → 只有 domain 套件內部能用

func NewUser(name string) User { ... } // ✅ 大寫 → 外部可以呼叫
func validate(u User) error { ... }    // ❌ 小寫 → 外部無法呼叫
```

> **記憶口訣**：大寫公開，小寫私有。就像寫信——「大」聲說的話大家都聽得到，「小」聲說的話只有身邊的人聽到。

## go.mod 逐行解釋

`go.mod` 是模組的身分證，位於專案的根目錄：

```
module blog-api              ← 模組名稱（你的專案叫什麼）
                                通常是 GitHub 路徑，如 github.com/username/blog-api

go 1.26.1                   ← 使用的 Go 版本（最低要求）

require (                    ← 依賴清單：你的專案需要哪些第三方套件
    github.com/gin-gonic/gin v1.12.0    ← 套件路徑 + 版本號
    gorm.io/gorm v1.31.1                ← 另一個依賴
)
```

| 欄位 | 用途 | 範例 |
|------|------|------|
| `module` | 定義模組名稱 | `module blog-api` |
| `go` | 指定 Go 最低版本 | `go 1.26.1` |
| `require` | 列出所有依賴及版本 | `github.com/gin-gonic/gin v1.12.0` |
| `// indirect` | 間接依賴（你的依賴的依賴） | 自動管理，不用手動改 |

## import 語句詳解

### 三種寫法

```go
// 寫法一：單行匯入（只匯入一個套件時）
import "fmt"

// 寫法二：群組匯入（最常用 ✅ 推薦）
import (
    "fmt"
    "strings"
)

// 寫法三：別名匯入（套件名稱衝突時使用）
import (
    "fmt"
    str "strings"    // 用 str 代替 strings
    _ "net/http/pprof"  // _ 表示只執行 init()，不使用套件的其他功能
)
```

### 匯入分組慣例

在專案中，匯入應該按照三組排列，每組之間空一行：

```go
import (
    // 第一組：標準庫（Go 內建的套件）
    "fmt"
    "net/http"
    "time"

    // 第二組：自己專案的套件
    "blog-api/internal/domain"
    "blog-api/internal/handler"
    "blog-api/pkg/config"

    // 第三組：第三方套件
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)
```

> **注意**：如果你匯入了套件但沒有使用，Go 編譯器會報錯！這是為了保持程式碼整潔。

## 標準庫常用套件

Go 的標準庫非常豐富，以下是最常用的套件：

| 套件 | 用途 | 常用函式 |
|------|------|---------|
| `fmt` | 格式化輸出 | `Println`、`Printf`、`Sprintf`、`Errorf` |
| `strings` | 字串處理 | `Contains`、`Split`、`Join`、`TrimSpace`、`ToLower` |
| `strconv` | 型別轉換 | `Atoi`、`Itoa`、`ParseBool`、`ParseFloat` |
| `time` | 時間處理 | `Now`、`Format`、`Parse`、`Since`、`Sleep` |
| `math` | 數學運算 | `Sqrt`、`Pow`、`Abs`、`Round`、`Pi` |
| `os` | 作業系統操作 | `Getenv`、`ReadFile`、`Exit` |
| `net/http` | HTTP 伺服器/客戶端 | `ListenAndServe`、`Get`、`HandleFunc` |
| `encoding/json` | JSON 處理 | `Marshal`、`Unmarshal` |
| `errors` | 錯誤處理 | `New`、`Is`、`As`、`Unwrap` |
| `log` | 日誌記錄 | `Println`、`Fatal`、`Printf` |
| `sort` | 排序 | `Ints`、`Strings`、`Slice` |
| `io` | 輸入/輸出 | `ReadAll`、`Copy` |

## 第三方套件

### 什麼是第三方套件？

第三方套件是其他開發者寫好並分享在網路上的程式碼。Go 的套件通常託管在 GitHub 上。

例如 `github.com/gin-gonic/gin`：
- `github.com` → 託管在 GitHub 上
- `gin-gonic` → 開發者或組織名稱
- `gin` → 套件名稱
- 這是一個 Web 框架，幫你快速建立 HTTP API

### 部落格專案用到的第三方套件

| 套件 | 用途 |
|------|------|
| `github.com/gin-gonic/gin` | Web 框架（處理 HTTP 請求） |
| `gorm.io/gorm` | ORM（用 Go 操作資料庫） |
| `github.com/golang-jwt/jwt/v5` | JWT 驗證（使用者登入） |
| `golang.org/x/crypto` | 密碼加密 |
| `github.com/fatih/color` | 彩色終端輸出 |

## 部落格專案的套件結構

```
blog-api/                          ← 專案根目錄（go.mod 所在位置）
├── cmd/                           ← 程式進入點
│   └── server/
│       └── main.go                ← main 函式在這裡，啟動伺服器
│
├── internal/                      ← 🔒 內部套件（只有本專案能匯入）
│   ├── domain/                    ← 領域模型：定義核心資料結構
│   │   ├── article.go             ←   Article、ArticleQuery 等型別
│   │   └── user.go                ←   User、RegisterRequest 等型別
│   ├── handler/                   ← 處理器：接收 HTTP 請求、回傳 HTTP 回應
│   │   ├── article_handler.go     ←   處理文章相關的 API
│   │   └── user_handler.go        ←   處理使用者相關的 API
│   └── repository/                ← 儲存庫：和資料庫溝通
│       ├── article_repository.go  ←   文章的 CRUD 操作
│       └── user_repository.go     ←   使用者的 CRUD 操作
│
├── pkg/                           ← 📦 公開套件（其他專案也能匯入）
│   ├── config/                    ← 設定檔讀取
│   └── response/                  ← 統一的 API 回應格式
│
├── go.mod                         ← 模組定義檔
└── go.sum                         ← 依賴的校驗碼（自動產生）
```

### internal/ vs pkg/ 的差別

| 特性 | `internal/` | `pkg/` |
|------|-------------|--------|
| 誰能匯入？ | 只有同一個模組內的程式碼 | 任何外部專案都可以 |
| 是否強制？ | **Go 編譯器強制執行** | 只是慣例（convention） |
| 適合放什麼？ | 商業邏輯、核心功能 | 通用工具、共用函式 |
| 範例 | domain、handler、repository | config、response、logger |

> **為什麼 `internal/` 很重要？** 它保護你的核心程式碼不被外部專案直接依賴。這樣你可以自由修改內部實作，不怕破壞別人的程式碼。

### cmd/ 目錄

`cmd/` 存放程式的進入點（`main` 套件）。如果一個專案有多個可執行程式，每個都放在 `cmd/` 的子目錄裡：

```
cmd/
├── server/main.go    ← API 伺服器
├── migrate/main.go   ← 資料庫遷移工具
└── seed/main.go      ← 種子資料工具
```

## 常用指令

| 指令 | 用途 |
|------|------|
| `go mod init <模組名>` | 建立新模組（產生 go.mod） |
| `go mod tidy` | 自動整理依賴（加入缺少的、移除多餘的） |
| `go get <套件路徑>` | 安裝第三方套件 |
| `go get <套件路徑>@latest` | 安裝最新版本 |
| `go get <套件路徑>@v1.2.3` | 安裝指定版本 |
| `go list -m all` | 列出所有依賴 |
| `go mod vendor` | 把依賴複製到 vendor/ 目錄 |

## 常見問題（FAQ）

### Q1: 什麼是循環匯入（Circular Import）？

如果套件 A 匯入套件 B，同時套件 B 也匯入套件 A，Go 編譯器會報錯。這叫做循環匯入。

```
❌ 不允許：
  handler 匯入 domain  ←─┐
  domain 匯入 handler  ───┘  循環了！
```

**解決方法**：
- 把共用的型別提取到第三個套件（例如 `domain`）
- 使用介面（interface）來解耦
- 重新思考套件的職責劃分

### Q2: init() 函式是什麼？

每個套件可以有一個 `init()` 函式，它會在程式啟動時自動執行（在 `main()` 之前）：

```go
package config

func init() {
    // 這裡的程式碼會在程式啟動時自動執行
    // 常用於初始化設定、註冊驅動程式等
}
```

執行順序：所有匯入套件的 `init()` → `main` 套件的 `init()` → `main()`

> **注意**：不要過度使用 `init()`，它會讓程式的初始化順序變得難以追蹤。

### Q3: go.sum 是什麼？

`go.sum` 記錄每個依賴的雜湊值（hash），用來確保：
- 你下載的套件和別人下載的是完全一樣的
- 套件在傳輸過程中沒有被篡改

**你不需要手動編輯 `go.sum`**，它會被 `go mod tidy` 和 `go get` 自動管理。但要記得把它一起提交到版本控制（git）。

### Q4: 匯入了套件但沒使用會怎樣？

Go 編譯器會報錯！這是 Go 的設計哲學：不要留無用的程式碼。

```go
import "fmt"      // 如果沒有用到 fmt，編譯會失敗
import _ "fmt"    // 加 _ 前綴可以避免報錯（但很少這樣做）
```

## 練習

1. **認識標準庫**：用 `strings` 套件寫一個函式，接受一個句子，回傳每個單字首字母大寫的結果（例如 "hello world" → "Hello World"）
2. **型別轉換**：模擬處理 URL 查詢參數，接受 `page`（字串）和 `limit`（字串），轉成整數後計算 `offset = (page - 1) * limit`
3. **時間格式化**：建立一個函式，接受 `time.Time`，回傳「3 分鐘前」、「2 小時前」、「昨天」這樣的相對時間字串

## 下一課預告

[第九課：Slice 與 Map](../09-slices-maps/) — 學習 Go 中最常用的兩個集合型別。切片（Slice）就像可以伸縮的陣列，Map 就像字典。這是處理「一組資料」的必備工具！
