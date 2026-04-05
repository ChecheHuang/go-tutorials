# 第六課：介面（Interfaces）

> **一句話總結**：介面就是「合約」—— 它定義了「你必須會做什麼」，但不管你「怎麼做」。

## 你會學到什麼？

- 介面的概念：一組方法的合約
- Go 的隱式實作（不需要寫 `implements`！）
- 多型：同一個函式處理不同型別
- 空介面 `any`（`interface{}`）：接受任何型別
- 型別斷言（Type Assertion）和 comma-ok 安全模式
- 型別 Switch：根據型別做不同處理
- 依賴注入（Dependency Injection）和 Clean Architecture 的 Repository 模式

## 執行方式

```bash
go run ./tutorials/06-interfaces
```

## 生活比喻：USB 接口

想像你電腦上的 **USB 接口**：

```
  USB 接口（介面）         各種裝置（實作）
┌─────────────────┐     ┌──────────────┐
│                 │     │  隨身碟       │ ← 能傳檔案
│  規格：         │     ├──────────────┤
│  - 能傳輸資料   │ ◄── │  鍵盤         │ ← 能輸入文字
│  - 能供電       │     ├──────────────┤
│                 │     │  滑鼠         │ ← 能移動游標
└─────────────────┘     ├──────────────┤
                        │  手機充電線   │ ← 能充電
                        └──────────────┘
```

- **USB 接口** = 介面（Interface）：定義了「必須能傳輸資料、能供電」
- **各種 USB 裝置** = 實作（Implementation）：每種裝置用不同方式滿足這個規格
- **你不需要知道裝置內部怎麼運作**，只要它符合 USB 規格就能插上去用

在 Go 中也一樣：

```go
// USB 接口 = Shape 介面
type Shape interface {
    Area() float64       // 規格 1：能算面積
    Perimeter() float64  // 規格 2：能算周長
}

// 隨身碟 = Circle（用自己的方式滿足規格）
// 鍵盤   = Rectangle（用另一種方式滿足規格）
```

## Go vs Java/C# 的介面比較

| 特性 | Go | Java / C# |
|------|------|------|
| 宣告語法 | `type Shape interface { ... }` | `interface Shape { ... }` |
| 實作方式 | **隱式**（自動滿足） | **顯式**（`implements Shape`） |
| 介面定義位置 | 通常在「使用方」 | 通常在「提供方」 |
| 是否需要匯入介面 | 不需要！ | 必須匯入 |

### 隱式 vs 顯式實作

```
Java（顯式）：
  class Circle implements Shape {  ← 必須明確宣告「我實作了 Shape」
      ...
  }

Go（隱式）：
  type Circle struct { ... }
  func (c Circle) Area() float64 { ... }       ← 只要有這兩個方法
  func (c Circle) Perimeter() float64 { ... }  ← Circle 就自動是 Shape！
```

**為什麼 Go 選擇隱式？**
- 不需要事先知道介面的存在 —— 你可以「事後」定義介面來匹配既有的型別
- 降低套件之間的耦合 —— 實作方不需要匯入介面方的套件
- 更靈活的組合 —— 第三方套件的型別也可以滿足你定義的介面

## Clean Architecture 中的依賴注入

```
┌─────────────────────────────────────────────┐
│                 main.go                      │
│  （在這裡決定要注入哪個實作）                   │
│                                              │
│  repo := repository.NewUserRepository(db)    │
│  usecase := usecase.NewUserUsecase(repo)     │
│  handler := handler.NewUserHandler(usecase)  │
└──────────────────┬──────────────────────────┘
                   │ 注入
                   ▼
┌──────────────────────────────────────┐
│         usecase（業務邏輯層）          │
│                                      │
│  type UserUsecase struct {           │
│      repo domain.UserRepository ←─── 依賴「介面」
│  }                                   │
└──────────────────┬───────────────────┘
                   │ 呼叫介面方法
                   ▼
┌──────────────────────────────────────┐
│   domain（定義介面 — 合約）            │
│                                      │
│  type UserRepository interface {     │
│      Create(user *User) error        │
│      FindByID(id uint) (*User, error)│
│  }                                   │
└──────────────────────────────────────┘
                   ▲ 實作介面
                   │
┌──────────────────────────────────────┐
│   repository（資料存取層 — 實作合約）   │
│                                      │
│  type userRepository struct {        │
│      db *gorm.DB                     │
│  }                                   │
│  func (r *userRepository) Create ... │
│  func (r *userRepository) FindByID.. │
└──────────────────────────────────────┘
```

**核心概念**：Usecase 依賴「介面」（合約），不依賴「具體實作」。這樣換資料庫或寫測試時，Usecase 完全不用改。

## 型別斷言（Type Assertion）

當你有一個 `any`（空介面）變數，想取出裡面的具體型別：

```go
var something any = "Hello"

// ✅ 安全寫法：comma-ok 模式（推薦！）
str, ok := something.(string)  // ok == true，str == "Hello"
num, ok := something.(int)     // ok == false，num == 0（零值）

// ❌ 危險寫法：如果型別不對，直接 panic！
str := something.(string)  // 可以，因為確實是 string
num := something.(int)     // panic: interface conversion 錯誤！
```

### 型別 Switch

需要判斷多種型別時，用型別 switch 更優雅：

```go
switch val := v.(type) {
case int:        // val 自動是 int 型別
case string:     // val 自動是 string 型別
case Shape:      // 也可以判斷是否滿足某個介面！
default:         // 以上都不是
}
```

## 空介面 `any`（`interface{}`）

```go
// 空介面 = 沒有任何方法要求的介面
// 所有型別都自動滿足它（因為不需要滿足任何條件）

func printAnything(v any) {  // any == interface{}
    fmt.Println(v)
}

printAnything(42)       // int → 可以！
printAnything("hello")  // string → 可以！
printAnything(true)     // bool → 可以！
```

**什麼時候用 `any`？**
- JSON 解析（結構未知時）
- 通用的日誌/除錯函式
- 需要容器裝不同型別的值時

**注意**：盡量少用 `any`，因為會失去型別安全性。能用具體介面就用具體介面。

## 在部落格專案中的應用

### UserRepository 介面

```go
// domain/user.go — 定義合約
type UserRepository interface {
    Create(user *User) error
    FindByEmail(email string) (*User, error)
    FindByID(id uint) (*User, error)
}
```

### ArticleRepository 介面

```go
// domain/article.go — 定義合約
type ArticleRepository interface {
    Create(article *Article) error
    FindAll() ([]Article, error)
    FindByID(id uint) (*Article, error)
    Update(article *Article) error
    Delete(id uint) error
}
```

### 為什麼用介面？

| 情境 | 不用介面 | 用介面 |
|------|---------|--------|
| 換資料庫 | 改 Usecase + Repository | 只改 Repository |
| 單元測試 | 需要真資料庫 | 注入 Mock，不需要資料庫 |
| 新增快取 | 改現有程式碼 | 寫一個新的帶快取 Repository |

## 常見問題 (FAQ)

### Q: 為什麼 Go 不用 `implements` 關鍵字？

Go 的設計哲學是「少即是多」。隱式介面讓你：
1. 不需要事先知道介面存在 —— 可以事後為既有型別定義介面
2. 不需要修改第三方套件的程式碼 —— 它們的型別自動滿足你的介面
3. 降低套件之間的依賴 —— 實作方不需要匯入定義介面的套件

### Q: 什麼時候該用空介面 `any`？

盡量少用！`any` 會失去編譯時的型別檢查。只在以下情況使用：
- 解析未知結構的 JSON
- 寫通用的日誌/除錯工具
- 標準函式庫要求（如 `fmt.Println` 的參數就是 `any`）

### Q: 介面應該定義多大？

Go 社群的慣例是「小介面」：
- 標準函式庫的 `io.Reader` 只有一個方法
- `io.Writer` 也只有一個方法
- 只在需要時才把介面變大
- 經驗法則：1-3 個方法的介面最好用

### Q: 介面變數是 nil 會怎樣？

```go
var s Shape           // s 是 nil
s.Area()              // panic: nil pointer dereference！
```

使用介面變數之前，永遠要檢查是否為 nil。

## 練習

1. **基礎**：定義一個 `Stringer` 介面（有 `String() string` 方法），讓 `User` 結構體實作它
2. **進階**：寫一個 `MockUserRepository`（用 map 儲存資料），確認它滿足 `UserRepository` 介面
3. **挑戰**：寫一個 `describeShape(s Shape)` 函式，用型別 switch 判斷是 Circle 還是 Rectangle，印出不同的描述
4. **思考**：為什麼 Go 把介面定義在「使用方」而不是「提供方」？這對程式架構有什麼影響？

## 下一課預告

**第七課：錯誤處理（Error Handling）** —— Go 沒有 try/catch，那怎麼處理錯誤？答案是：錯誤就是普通的值，用回傳值傳遞。這是 Go 最獨特的設計之一。
