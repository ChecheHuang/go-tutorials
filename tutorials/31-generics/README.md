# 第三十一課：Go 泛型（Generics）

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟢 初級工程師 | 了解泛型存在，能讀懂泛型函式 |
| 🟡 中級工程師 | **重點**：能寫泛型函式和資料結構，了解型別約束 |
| 🔴 資深工程師 | 判斷何時該用泛型，設計可重用的泛型函式庫 |

## 核心語法

```go
// 型別參數（Type Parameter）
func Min[T cmp.Ordered](a, b T) T {
    if a < b { return a }
    return b
}

// 使用（型別推斷，不需明確寫出 T）
Min(3, 5)          // T = int
Min("a", "b")      // T = string

// 自訂型別約束
type Number interface {
    ~int | ~float64  // ~ 代表底層型別
}
```

## 常用型別約束

| 約束 | 說明 | 來源 |
|------|------|------|
| `any` | 任何型別（= `interface{}`）| 內建 |
| `comparable` | 可用 `==` 比較 | 內建 |
| `cmp.Ordered` | 可用 `<`、`>`、`<=`、`>=` | `cmp` 套件 |
| 自訂 interface | 限制有特定方法或底層型別 | 自己定義 |

## 泛型資料結構

```go
// Stack[T] — 可存任何型別
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) { ... }
func (s *Stack[T]) Pop() (T, bool) { ... }

// 使用
var s Stack[string]
s.Push("hello")

var numStack Stack[int]
numStack.Push(42)
```

## 函式式工具

```go
names := Map(users, func(u User) string { return u.Name })
admins := Filter(users, func(u User) bool { return u.Role == "admin" })
total := Reduce(nums, 0, func(acc, n int) int { return acc + n })
```

## 何時用泛型？

✅ **適合**：資料結構、工具函式（Map/Filter/Reduce）、數學運算
❌ **不適合**：業務邏輯（用介面）、只有一種型別（直接寫）

## 執行方式

```bash
go run ./tutorials/31-generics
```
