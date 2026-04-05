# 第四課：結構體與方法

## 學習目標

- 理解結構體（Struct）—— Go 版的「物件」
- 學會定義方法（Methods）
- 分辨值接收者與指標接收者的差異
- 了解結構體嵌套與建構函式模式

## 執行方式

```bash
cd tutorials/04-structs-methods
go run main.go
```

## 重點筆記

### Go 沒有 Class，用 Struct + Method 代替

| OOP 概念 | Go 的對應方式 |
|----------|-------------|
| Class | Struct |
| Constructor | `NewXxx()` 函式 |
| Method | `func (receiver) Method()` |
| 繼承 | 結構體嵌套（Embedding） |
| 介面 | Interface（下一課） |

### 值接收者 vs 指標接收者

```go
func (u User) Name() string    // 值接收者：不修改原始值
func (u *User) SetName(n string) // 指標接收者：會修改原始值
```

**何時用指標接收者？**
- 需要修改接收者的值
- 結構體很大，避免複製的開銷
- 一般建議：**同一個型別的所有方法統一使用指標接收者**

### 在專案中的應用

部落格專案的每一層都大量使用結構體和方法：

```go
// domain 層定義結構體
type User struct { ... }

// repository 層的結構體持有資料庫連線
type userRepository struct {
    db *gorm.DB
}

// 方法實作資料操作
func (r *userRepository) Create(user *User) error { ... }

// handler 層的結構體持有 usecase
type UserHandler struct {
    userUsecase usecase.UserUsecase
}

func (h *UserHandler) Register(c *gin.Context) { ... }
```

## 練習

1. 定義一個 `Rectangle` 結構體（Width, Height），加上 `Area()` 和 `Perimeter()` 方法
2. 寫一個 `NewRectangle` 建構函式
3. 試試把 `SetEmail` 的接收者改成值接收者 `(u User)`，觀察行為差異
