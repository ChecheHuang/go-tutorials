# 第四課：結構體與方法（Structs & Methods）

> **一句話總結**：結構體就像「身分證」——把一個人的所有資料寫在同一張卡片上；方法就像「技能」——這個人能做什麼事。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟢 初學者 | **入門必修**：Go 用 struct 取代 class |
| 🟡 中級工程師 | 值接收者 vs 指標接收者的取捨 |

## 你會學到什麼？

- 結構體（Struct）是什麼——Go 版的「物件」
- 四種建立結構體的方式
- 方法（Method）——和函式的差別
- 值接收者 vs 指標接收者——深度理解「什麼時候用哪個」
- 結構體嵌套（Embedding）——Go 的「組合」哲學
- 建構函式模式（Constructor Pattern）——`NewXxx()` 慣例
- 結構體標籤（Struct Tags）——`json`、`binding`、`gorm` 的作用

## 執行方式

```bash
go run ./tutorials/04-structs-methods
```

## 用生活來理解

### 結構體 = 身分證

一張身分證上面有：姓名、生日、地址、照片、身分證字號。

這些資料**彼此相關**——它們都在描述同一個人。
結構體就是把這些相關的資料「打包」在一起。

```
身分證（struct）
┌─────────────────────┐
│ 字號：A123456789    │  ← ID int
│ 姓名：王小明        │  ← Username string
│ 信箱：ming@mail.com │  ← Email string
│ 年齡：25            │  ← Age int
│ 狀態：有效          │  ← IsActive bool
└─────────────────────┘
```

```go
type User struct {
    ID       int
    Username string
    Email    string
    Age      int
    IsActive bool
}
```

### 方法 = 技能

「人」是結構體，「人能做的事」就是方法：

```
王小明（User）的技能：
  ├── 自我介紹（DisplayName）→ 只需要「讀取」自己的名字
  ├── 換信箱（SetEmail）     → 需要「修改」自己的資料
  └── 驗證年齡（IsAdult）    → 只需要「讀取」自己的年齡
```

「只讀取」用**值接收者**，「要修改」用**指標接收者**。

### 嵌套 = 合體

想像身分證 + 名片 = 員工證：

```
員工證（Employee）
┌─────────────────────────────┐
│ ┌── 身分證（User）──────┐   │
│ │ 姓名：王小明          │   │
│ │ 信箱：ming@mail.com   │   │
│ └───────────────────────┘   │
│ ┌── 地址卡（Address）───┐   │
│ │ 城市：台北            │   │
│ │ 國家：台灣            │   │
│ └───────────────────────┘   │
│ 部門：工程部                │
│ 薪資：85000                 │
└─────────────────────────────┘
```

## Go 和 OOP 的對應關係

Go 沒有 `class`，但你可以用結構體和方法達成 OOP 的效果：

| OOP 概念      | Go 的對應方式                     | 範例                                    |
|---------------|----------------------------------|-----------------------------------------|
| Class         | `type Xxx struct`                | `type User struct { ... }`              |
| Constructor   | `NewXxx()` 函式                  | `func NewUser(...) *User`               |
| Method        | `func (receiver) Method()`       | `func (u User) Name() string`          |
| Property      | struct 欄位                      | `u.Username`                            |
| 繼承          | 結構體嵌套（Embedding）           | `type Employee struct { User }`         |
| 介面          | `type Xxx interface`（第六課）   | `type Reader interface { Read() }`      |
| this / self   | 接收者變數（`u`、`a`、`h`）       | `func (u *User) SetEmail(e string)`    |
| 封裝          | 大寫 = 公開、小寫 = 私有          | `Username`（公開） vs `password`（私有）|

### 關鍵差異

- Go **沒有**繼承鏈（不會有「爺爺 → 爸爸 → 兒子」的關係）
- Go 用**組合**（把小零件拼在一起）代替繼承（從上到下的分類）
- Go 的介面是**隱式實作**——不需要寫 `implements`（第六課詳解）

## 值接收者 vs 指標接收者——深度解析

這是 Go 結構體中**最重要**的概念，值得花時間理解。

### 基本差異

```go
// 值接收者：u 是原始結構體的「副本」（影印本）
func (u User) DisplayName() string {
    // 在這裡修改 u，不會影響原始值
    return u.Username
}

// 指標接收者：u 指向原始結構體（拿到正本）
func (u *User) SetEmail(email string) {
    // 在這裡修改 u，會直接改到原始值
    u.Email = email
}
```

### 什麼時候用指標接收者？

| 使用情境           | 用哪個？     | 為什麼？                                           |
|-------------------|-------------|---------------------------------------------------|
| 需要修改接收者的值  | `*User` ✅  | 值接收者操作的是副本，修改了也是白改                   |
| 結構體很大         | `*User` ✅  | 指標只有 8 bytes，避免每次呼叫都複製整個結構體          |
| 一致性原則         | `*User` ✅  | 同一個型別的方法應該統一，不要混用                     |
| 方法會被介面使用    | `*User` ✅  | `*User` 實作介面，`User` 不一定能（第六課詳解）        |
| 小型、唯讀結構體    | `User` 可以  | 像只有 2-3 個欄位的小結構體，讀取時用值接收者也行       |

### 判斷流程

```
需要修改接收者嗎？
  ├── 是 → 用指標接收者 (*T)
  └── 否 → 這個型別的其他方法有用指標接收者嗎？
              ├── 是 → 為了一致性，也用指標接收者 (*T)
              └── 否 → 結構體很大嗎？
                        ├── 是 → 用指標接收者 (*T)
                        └── 否 → 值接收者 (T) 可以
```

### 業界實務建議

> **簡單法則：如果你不確定，就用指標接收者。**
>
> 在正式專案中，絕大多數的方法都使用指標接收者。
> 值接收者通常只用在非常小的、不可變的結構體上（例如座標 `Point{X, Y}`）。

## 結構體標籤（Struct Tags）詳解

標籤是附在欄位上的元資料，不同的套件會讀取不同的標籤：

### 常見標籤

| 標籤              | 套件    | 用途                              | 範例                                |
|-------------------|--------|-----------------------------------|-------------------------------------|
| `json:"name"`     | 標準庫  | 控制 JSON 欄位名稱                 | `json:"user_name"`                  |
| `json:"-"`        | 標準庫  | 從 JSON 中隱藏此欄位               | `json:"-"`（密碼、Token 等機密資料） |
| `json:",omitempty"` | 標準庫 | 零值時不輸出到 JSON                 | `json:"bio,omitempty"`              |
| `binding:"required"` | Gin  | 標記為必填欄位（請求驗證用）         | `binding:"required"`                |
| `binding:"email"` | Gin    | 驗證是否為合法的 email 格式         | `binding:"required,email"`          |
| `binding:"min=6"` | Gin    | 最小長度驗證                        | `binding:"required,min=6,max=50"`   |
| `gorm:"primaryKey"` | GORM | 標記為資料庫主鍵                    | `gorm:"primaryKey"`                 |
| `gorm:"index"`    | GORM   | 為欄位建立資料庫索引                 | `gorm:"index"`                      |
| `gorm:"unique"`   | GORM   | 唯一性約束                          | `gorm:"unique"`                     |
| `gorm:"not null"` | GORM   | 非空約束                            | `gorm:"not null"`                   |
| `gorm:"column:xxx"` | GORM | 自訂資料庫欄位名稱                  | `gorm:"column:user_name"`           |

### 多個標籤的寫法

一個欄位可以同時有多個標籤，用空格分隔：

```go
type User struct {
    Email string `json:"email" binding:"required,email" gorm:"unique;not null"`
}
```

### 為什麼需要 json 標籤？

Go 的欄位名稱是大寫開頭（`Username`），但 JSON 慣例是小寫蛇形命名（`user_name`）。
標籤讓你控制輸出格式，不需要改變 Go 程式碼的命名慣例：

```go
type User struct {
    Username string `json:"user_name"`  // Go 裡叫 Username，JSON 裡叫 user_name
}
```

## 在部落格專案中的應用

在真實的部落格專案中，結構體和方法無處不在：

### Domain 層——定義資料結構

```go
// domain/user.go
type User struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Username  string    `json:"username" gorm:"unique;not null"`
    Email     string    `json:"email" gorm:"unique;not null"`
    Password  string    `json:"-"`   // 密碼永遠不輸出到 JSON！
    CreatedAt time.Time `json:"created_at"`
}

// domain/article.go
type Article struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Title     string    `json:"title" binding:"required" gorm:"not null"`
    Content   string    `json:"content" binding:"required"`
    AuthorID  uint      `json:"author_id" gorm:"index"`
    Author    User      `json:"author" gorm:"foreignKey:AuthorID"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Repository 層——結構體持有資料庫連線

```go
// repository/user_repository.go
type userRepository struct {
    db *gorm.DB   // 結構體欄位：持有資料庫連線
}

// 建構函式
func NewUserRepository(db *gorm.DB) *userRepository {
    return &userRepository{db: db}
}

// 方法：用指標接收者（一致性原則 + 避免複製 gorm.DB）
func (r *userRepository) Create(user *User) error {
    return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uint) (*User, error) {
    var user User
    err := r.db.First(&user, id).Error
    return &user, err
}
```

### Handler 層——結構體持有 UseCase

```go
// handler/user_handler.go
type UserHandler struct {
    userUsecase usecase.UserUsecase   // 持有 usecase 介面
}

func NewUserHandler(uc usecase.UserUsecase) *UserHandler {
    return &UserHandler{userUsecase: uc}
}

func (h *UserHandler) Register(c *gin.Context) {
    // 處理使用者註冊的 HTTP 請求
}

func (h *UserHandler) Login(c *gin.Context) {
    // 處理使用者登入的 HTTP 請求
}
```

### 請求/回應結構體

```go
// 註冊請求
type RegisterRequest struct {
    Username string `json:"username" binding:"required,min=3,max=30"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

// API 回應
type Response struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

## 常見問題（FAQ）

### Q1：Go 真的沒有 class 嗎？

是的。Go 故意不加入 class，因為 Go 的設計者（包括 C 語言之父 Ken Thompson）認為繼承帶來的複雜性大於它的好處。Go 用結構體 + 方法 + 介面 + 嵌套，提供了一套更簡單、更靈活的方式來組織程式碼。

### Q2：為什麼建構函式回傳指標 `*User` 而不是值 `User`？

三個原因：
1. **避免複製**：回傳值會複製整個結構體，回傳指標只複製一個地址（8 bytes）
2. **一致性**：指標可以直接呼叫指標接收者的方法
3. **慣例**：Go 社群的標準做法，大家都這樣寫

### Q3：什麼時候用值接收者就好？

當結構體**很小**（例如只有 2-3 個基本型別的欄位）且**不需要修改**時，值接收者是可以的。例如：

```go
type Point struct { X, Y float64 }
func (p Point) Distance() float64 { ... }  // 值接收者 OK
```

但即使是小結構體，如果有任何一個方法需要指標接收者，建議全部統一用指標接收者。

### Q4：`json:"-"` 和 `json:",omitempty"` 有什麼不同？

- `json:"-"`：**永遠不會**出現在 JSON 中，不管有沒有值
- `json:",omitempty"`：只有在**零值**時才不出現；有值的話還是會出現

### Q5：嵌套（Embedding）和「有個欄位」有什麼不同？

```go
// 嵌套（匿名欄位）——可以直接存取 User 的欄位
type Employee struct {
    User         // emp.Username 直接存取
    Department string
}

// 具名欄位——必須通過欄位名存取
type Employee2 struct {
    UserInfo   User   // emp.UserInfo.Username 才能存取
    Department string
}
```

嵌套讓你可以「提升」被嵌套結構體的欄位和方法，用起來更方便。

## 練習題

1. **基礎練習**：定義一個 `Rectangle` 結構體（Width、Height float64），加上 `Area()`（面積）和 `Perimeter()`（周長）方法，並寫一個 `NewRectangle` 建構函式
2. **理解差異**：把 `SetEmail` 的接收者改成值接收者 `(u User)`，執行看看有什麼不同
3. **嵌套練習**：定義 `Shape` 結構體（Color string），讓 `Rectangle` 嵌套 `Shape`，使得 `rect.Color` 可以直接存取
4. **JSON 練習**：給 `Rectangle` 加上 `json` 標籤，把它轉成 JSON 並印出來
5. **挑戰題**：設計一個 `BlogPost` 結構體，包含標題、內容、作者（嵌套 User）、標籤（`[]string`），並寫 `NewBlogPost` 建構函式和 `Summary()` 方法

## 下一課預告

**第五課：指標（Pointers）**

在這一課中你已經看到了 `*User` 和 `&User{...}`。下一課我們會深入理解：
- 什麼是記憶體地址？
- `*` 和 `&` 到底在做什麼？
- 為什麼 Go 需要指標？
- 指標和值傳遞的完整真相
