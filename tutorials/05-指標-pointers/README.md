# 第五課：指標

## 學習目標

- 理解指標是什麼、為什麼需要它
- 學會 `&`（取址）和 `*`（解引用）運算子
- 分辨傳值與傳參考的差異
- 理解為什麼專案中大量使用 `*User` 而非 `User`

## 執行方式

```bash
cd tutorials/05-pointers
go run main.go
```

## 重點筆記

### 指標速查表

| 語法 | 意義 | 範例 |
|------|------|------|
| `*int` | 「指向 int 的指標」型別 | `var p *int` |
| `&x` | 取得 x 的位址 | `p := &x` |
| `*p` | 透過指標取得值 | `fmt.Println(*p)` |
| `nil` | 指標的零值（不指向任何東西） | `if p == nil` |

### 記憶口訣

```
& → 取「地址」（& 長得像 A，Address）
* → 取「值」  （穿透指標，取出裡面的東西）
```

### 專案中的三個關鍵場景

**場景 1：Repository 回傳 `*User`**
```go
func FindByID(id uint) (*User, error)
// 找到 → 回傳 &user, nil
// 沒找到 → 回傳 nil, error
```

**場景 2：GORM 需要指標來回寫 ID**
```go
user := &domain.User{Username: "alice"}
db.Create(user)
// GORM 透過指標把自動產生的 ID 寫回 user.ID
```

**場景 3：方法接收者用指標**
```go
func (r *userRepository) Create(user *User) error
// *userRepository：避免複製整個 struct
// *User：讓 GORM 能回寫 ID
```

## 練習

1. 寫一個 `swap(a, b *int)` 函式，交換兩個整數的值
2. 建立一個 `Counter` 結構體，用指標接收者實作 `Increment()` 方法
3. 思考：為什麼 `fmt.Println` 不需要傳指標也能印出值？
