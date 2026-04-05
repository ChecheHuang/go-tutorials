# 第三課：函式（Functions）

> **一句話總結**：函式就是「把程式碼包裝起來，取個名字，需要時隨時呼叫」。

## 你會學到什麼？

- 怎麼建立函式、傳入參數、接收回傳值
- Go 的多重回傳值——回傳 `(結果, error)` 的錯誤處理模式
- 可變參數——讓函式接收任意數量的參數
- 函式也是「值」——可以存到變數、當參數傳遞
- 函式型別——幫函式簽名取名字（像 Gin 的 HandlerFunc）
- 閉包——函式「記住」外部變數
- defer——「等函式結束時再執行」

## 執行方式

```bash
go run ./tutorials/03-functions
```

## 用生活來理解

### 函式 = 食譜

```
食譜名稱：煎蛋（材料：雞蛋、鹽）→ 成品：荷包蛋
  1. 熱鍋
  2. 打蛋
  3. 加鹽
  4. 翻面
  5. 完成 → 回傳荷包蛋
```

```go
func 煎蛋(雞蛋 Egg, 鹽 Salt) 荷包蛋 {
    // 一系列步驟
    return 荷包蛋
}
```

你不需要每次煎蛋都重寫步驟，只要**呼叫食譜**就好。

## 函式語法拆解

```go
func add(a int, b int) int {
│     │    │         │    │
│     │    │         │    └── 回傳值型別
│     │    │         └────── 第二個參數
│     │    └──────────────── 第一個參數
│     └───────────────────── 函式名稱
└─────────────────────────── func 關鍵字
    return a + b
}
```

## 多重回傳值——Go 的錯誤處理核心

這是 Go 最重要的設計模式之一！

### 為什麼需要多重回傳值？

很多語言用 `try/catch` 處理錯誤。Go 認為錯誤是「正常的事」，不應該用例外機制，所以用多重回傳值：

```go
// 回傳 (結果, 錯誤)
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("除數不能為零")  // 有錯誤
    }
    return a / b, nil  // nil = 沒有錯誤
}

// 使用時：
result, err := divide(10, 3)
if err != nil {       // 先檢查錯誤
    fmt.Println("出錯了:", err)
    return
}
fmt.Println("結果:", result)  // 確認沒錯才用結果
```

> **nil 是什麼？** `nil` 代表「空的、什麼都沒有」。當 `error` 是 `nil` 時，表示「沒有錯誤」。在部落格專案中，幾乎每個函式都回傳 `error`。

### 在部落格專案中的應用

```go
// internal/usecase/user_usecase.go
func (u *userUsecase) Login(req domain.LoginRequest) (*domain.LoginResponse, error) {
    //                                                  ↑ 結果           ↑ 錯誤
    user, err := u.userRepo.GetByEmail(req.Email)
    if err != nil {
        return nil, fmt.Errorf("帳號或密碼錯誤")
    }
    // ...
}
```

## 函式也是「值」

在 Go 中，函式不只是程式碼區塊——它也是一種「值」，就像 `int`、`string` 一樣。

### 存到變數

```go
// 把函式存到變數 f 裡
f := func(x int) int { return x * 2 }
fmt.Println(f(5))  // 10
```

### 當參數傳遞

```go
func apply(items []string, transform func(string) string) []string {
    //                       ↑ 參數型別是「一個函式」
}
```

### 用 type 幫函式型別取名字

```go
// 定義一個函式型別
type HandlerFunc func(request string)

// 這跟 Gin 框架裡的概念一模一樣：
// type HandlerFunc func(*gin.Context)
// 每個路由處理器的型別就是 gin.HandlerFunc
```

> **為什麼要知道這個？** 因為在部落格專案中，`router.GET("/articles", handler.GetAll)` 裡面的 `handler.GetAll` 就是一個「函式值」被當作參數傳給 `GET`。

## 閉包（Closure）

閉包 = 函式 + 它「記住」的外部變數。

```go
func makeCounter() func() int {
    count := 0              // 外部變數
    return func() int {     // 回傳一個函式
        count++             // 這個函式「記住」了 count
        return count
    }
}

counter := makeCounter()
counter()  // 1
counter()  // 2
counter()  // 3  ← count 一直被記住，不會消失
```

> **類比**：閉包像一個人帶著背包。背包裡的東西（count）走到哪帶到哪，別人看不到也拿不走。

### 在部落格專案中的應用

```go
// internal/middleware/jwt.go — 中介層就是閉包的典型應用
func JWTAuth(cfg *config.Config) gin.HandlerFunc {
    // cfg 被回傳的函式「記住」了（閉包捕獲）
    return func(c *gin.Context) {
        // 這裡可以存取外層的 cfg
        // 用 cfg.JWTSecret 來驗證 JWT Token
    }
}
```

## defer——延遲執行

```go
func processFile(filename string) {
    file := open(filename)      // 1. 開啟檔案
    defer file.Close()          // 排隊：等函式結束時才關閉
    // ... 處理檔案 ...         // 2. 做事
}                               // 3. 函式結束，自動執行 defer 關閉檔案
```

### defer 的執行順序：後進先出（LIFO）

```go
defer fmt.Println("第一個 defer")   // 排在最底下
defer fmt.Println("第二個 defer")   // 排在中間
defer fmt.Println("第三個 defer")   // 排在最上面
// 執行順序：第三 → 第二 → 第一（像疊盤子，最後放的最先拿）
```

> **為什麼要後進先出？** 因為資源的取得和釋放通常是「相反順序」。先開 A 再開 B，關的時候應該先關 B 再關 A（像穿脫衣服）。

### 在部落格專案中

```go
// internal/middleware/recovery.go — defer + recover 攔截 panic
func Recovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // 即使程式 panic 了，這裡也一定會執行
                // 攔截 panic，回傳 500 錯誤，不讓伺服器崩潰
            }
        }()
        c.Next()
    }
}
```

## strings 套件——內建的字串工具箱

本課用到了 `strings.ToUpper`。`strings` 是 Go 內建的字串處理套件：

| 函式 | 功能 | 範例 |
|------|------|------|
| `strings.ToUpper(s)` | 轉大寫 | `"hello"` → `"HELLO"` |
| `strings.ToLower(s)` | 轉小寫 | `"HELLO"` → `"hello"` |
| `strings.Contains(s, sub)` | 是否包含子字串 | `"hello", "ell"` → `true` |
| `strings.Split(s, sep)` | 切割字串 | `"a,b,c", ","` → `["a","b","c"]` |
| `strings.Join(arr, sep)` | 合併字串 | `["a","b"], "-"` → `"a-b"` |
| `strings.TrimSpace(s)` | 去掉前後空白 | `" hello "` → `"hello"` |

## 常見問題

### Q: 什麼時候用命名回傳值？
當函式回傳多個同型別的值時，命名回傳值讓程式碼更清楚。但一般簡短函式用普通回傳就好。

### Q: 匿名函式是什麼？
就是「沒有名字的函式」，直接寫在需要的地方。像 `func(x int) int { return x * 2 }`。其他語言叫 lambda 或箭頭函式。

### Q: 為什麼 Go 沒有 try/catch？
Go 的設計者認為 `try/catch` 會讓錯誤處理變得「隱形」——你可能會忘記處理錯誤。多重回傳值強迫你在**每個可能出錯的地方**都明確處理錯誤，雖然寫起來多幾行，但程式碼更可靠。

## 練習題

1. **基礎題**：寫一個函式 `isEven(n int) bool`，判斷數字是不是偶數
2. **多重回傳**：寫一個 `safeSqrt(n float64) (float64, error)`，負數時回傳錯誤
3. **高階函式**：寫一個 `filter(nums []int, fn func(int) bool) []int`，回傳所有讓 fn 回傳 true 的數字
4. **閉包題**：寫一個 `makeAdder(x int) func(int) int`，回傳一個「加上 x」的函式

## 下一課預告

學會了函式，接下來要學**結構體與方法**——Go 的「物件導向」方式！
