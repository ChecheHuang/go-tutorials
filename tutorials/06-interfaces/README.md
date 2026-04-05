# 第六課：介面（Interfaces）

## 學習目標

- 理解介面的概念：「一組方法的約定」
- 體會 Go 的隱式介面實作（不需要 `implements`）
- 學會依賴注入（Dependency Injection）模式
- 理解為什麼 Clean Architecture 能輕鬆切換資料庫

## 執行方式

```bash
cd tutorials/06-interfaces
go run main.go
```

## 重點筆記

### 核心概念

```
介面 = 一份「合約」
實作介面 = 履行合約中的所有條款（方法）

只要你有 Area() 和 Perimeter()，你就是 Shape
不需要宣告「我實作了 Shape」
```

### Go vs Java/C# 的介面

| | Go | Java/C# |
|---|---|---|
| 宣告 | `type Shape interface { ... }` | `interface Shape { ... }` |
| 實作 | **隱式**（自動滿足） | **顯式**（`implements Shape`） |
| 位置 | 介面定義在使用方 | 介面定義在提供方 |

### 這就是 Clean Architecture 的核心

```
domain/user.go:
    type UserRepository interface {    ← 定義「我需要什麼」
        Create(user *User) error
        FindByID(id uint) (*User, error)
    }

repository/user_repository.go:
    type userRepository struct { db *gorm.DB }
    func (r *userRepository) Create(...) error { ... }    ← 用 GORM 實作
    func (r *userRepository) FindByID(...) (*User, error) { ... }

usecase/user_usecase.go:
    type userUsecase struct {
        userRepo domain.UserRepository    ← 依賴介面，不依賴具體實作
    }
```

**好處：** 想把 SQLite 換成 PostgreSQL？只需要寫一個新的 Repository 實作，Usecase 完全不用改。測試時？注入 Mock Repository 就好。

## 練習

1. 定義一個 `Stringer` 介面（有 `String() string` 方法），讓 `User` 實作它
2. 寫一個 `MockUserRepository`（把資料存在 map 中），確認它滿足 `UserRepository` 介面
3. 思考：為什麼 Go 選擇隱式介面實作？這對程式碼解耦有什麼好處？
