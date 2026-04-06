# 第二十四課：Go 泛型（Generics）

> **一句話總結**：泛型讓你寫一個函式或資料結構，能同時處理多種型別，不用再靠 `interface{}` 和型別斷言。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟢 初級工程師 | 了解泛型存在，能讀懂泛型函式 |
| 🟡 中級工程師 | **重點**：能寫泛型函式和資料結構，了解型別約束 |
| 🔴 資深工程師 | 判斷何時該用泛型，設計可重用的泛型函式庫 |

## 你會學到什麼？

- Go 1.18 之前沒有泛型的痛點（`interface{}` 地獄）
- 型別參數（Type Parameter）的語法：`func Foo[T any](x T) T`
- 型別約束（Constraint）：`any`、`comparable`、`cmp.Ordered`、自訂約束
- 泛型函式實作：Map、Filter、Reduce、Contains
- 泛型資料結構實作：Stack[T]、Queue[T]
- 標準庫中的泛型套件：`slices`、`maps`、`cmp`
- 何時該用泛型、何時不該用（過度工程警告）
- 泛型 vs `interface{}` 的效能差異

## 執行方式

```bash
go run ./tutorials/24-generics
```

## 生活比喻：萬用工具組

```
想像你有一組扳手：

沒有泛型的世界（Go 1.18 之前）：
  你只有一把「萬能扳手」（interface{}）
  - 能轉任何螺絲，但每次都要手動調整大小（型別斷言）
  - 調錯了就會滑牙（panic: interface conversion）
  - 沒辦法事先知道會不會出錯（編譯器不幫你檢查）

有泛型的世界（Go 1.18 之後）：
  你有一組「自適應扳手」[T]
  - 你說「這把給 int」，它就只能轉 int 螺絲
  - 你說「這把給 string」，它就只能轉 string 螺絲
  - 用錯型別？編譯器直接報錯，連工具箱都打不開
  - 但底層就是同一個設計，不用重複做 100 把不同扳手
```

## 為什麼需要泛型？Go 1.18 之前的痛

### 問題一：`interface{}` 失去型別安全

```go
// Go 1.18 之前：用 interface{} 實作「通用」的 Contains
func Contains(slice []interface{}, target interface{}) bool {
    for _, item := range slice {
        if item == target {
            return true
        }
    }
    return false
}

// 使用時超級痛苦
nums := []interface{}{1, 2, 3}    // 不能直接用 []int
Contains(nums, "hello")           // 編譯通過！但邏輯上不對（int 和 string 比較）
```

### 問題二：每個型別都要寫一個版本

```go
// Go 1.18 之前：想寫 Min 函式？每個型別寫一次
func MinInt(a, b int) int {
    if a < b { return a }
    return b
}

func MinFloat64(a, b float64) float64 {
    if a < b { return a }
    return b
}

func MinString(a, b string) string {
    if a < b { return a }
    return b
}

// 三個函式的邏輯一模一樣，只是型別不同
```

### 問題三：型別斷言的 runtime panic

```go
// Go 1.18 之前：從 interface{} 取值需要型別斷言
func GetFirst(items []interface{}) interface{} {
    return items[0]
}

result := GetFirst([]interface{}{42})
num := result.(int)      // OK
str := result.(string)   // runtime panic! 編譯器不會擋你
```

### 泛型的解法

```go
// Go 1.18+：一個函式搞定所有可比較的型別
func Contains[T comparable](slice []T, target T) bool {
    for _, item := range slice {
        if item == target {
            return true
        }
    }
    return false
}

// 使用——型別安全，編譯器幫你檢查
Contains([]int{1, 2, 3}, 4)          // T = int（自動推斷）
Contains([]string{"a", "b"}, "c")    // T = string（自動推斷）
// Contains([]int{1, 2, 3}, "hello") // 編譯錯誤！string 不是 int
```

## 型別參數語法

```
func 函式名[T 約束](參數 T) 回傳值 {
         ▲  ▲
         │  └── 約束（Constraint）：T 必須滿足的條件
         └───── 型別參數（Type Parameter）：代表某個型別的佔位符
}
```

### 基本範例

```go
// 單一型別參數
func Print[T any](value T) {
    fmt.Println(value)
}

// 多個型別參數
func Map[T any, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = fn(v)
    }
    return result
}

// 明確指定型別（通常不需要，Go 能自動推斷）
Print[int](42)
Print[string]("hello")

// 自動推斷（推薦）
Print(42)       // T = int
Print("hello")  // T = string
```

## 型別約束（Constraints）

約束決定了型別參數 `T` 可以做什麼操作。

### 內建約束

| 約束 | 說明 | 可用的操作 | 來源 |
|------|------|-----------|------|
| `any` | 任何型別（= `interface{}`） | 賦值、傳遞 | 內建 |
| `comparable` | 可比較型別 | `==`、`!=` | 內建 |
| `cmp.Ordered` | 可排序型別（int、float、string 等） | `<`、`>`、`<=`、`>=`、`==` | `cmp` 套件 |

### 自訂約束

```go
// 方法一：用 interface 限制底層型別
type Number interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~float32 | ~float64
}

// ~ 代表「底層型別（underlying type）」
// 例如 type Age int → Age 的底層型別是 int，符合 ~int

func Sum[T Number](nums []T) T {
    var total T
    for _, n := range nums {
        total += n
    }
    return total
}

// 方法二：限制必須有某個方法
type Stringer interface {
    String() string
}

func PrintAll[T Stringer](items []T) {
    for _, item := range items {
        fmt.Println(item.String())
    }
}

// 方法三：組合（底層型別 + 方法）
type OrderedStringer interface {
    ~int | ~string
    String() string
}
```

### 約束選擇指南

```
需要什麼操作？
  │
  ├── 什麼都不需要（只存取）──────→ any
  │
  ├── 需要 == 比較 ──────────────→ comparable
  │
  ├── 需要 < > 排序 ─────────────→ cmp.Ordered
  │
  ├── 需要特定方法 ──────────────→ 自訂 interface（帶方法）
  │
  └── 需要限制底層型別 + 運算 ───→ 自訂 interface（帶 ~type）
```

## 泛型函式實戰

### Map：轉換切片中的每個元素

```go
func Map[T any, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = fn(v)
    }
    return result
}

// 使用
names := Map(users, func(u User) string { return u.Name })
ids := Map(articles, func(a Article) uint { return a.ID })
```

### Filter：過濾符合條件的元素

```go
func Filter[T any](slice []T, predicate func(T) bool) []T {
    var result []T
    for _, v := range slice {
        if predicate(v) {
            result = append(result, v)
        }
    }
    return result
}

// 使用
admins := Filter(users, func(u User) bool { return u.Role == "admin" })
published := Filter(articles, func(a Article) bool { return a.Published })
```

### Reduce：將切片歸約為單一值

```go
func Reduce[T any, U any](slice []T, initial U, fn func(U, T) U) U {
    result := initial
    for _, v := range slice {
        result = fn(result, v)
    }
    return result
}

// 使用
total := Reduce(prices, 0.0, func(sum float64, p float64) float64 {
    return sum + p
})

// 串接字串
csv := Reduce(names, "", func(acc string, name string) string {
    if acc == "" { return name }
    return acc + "," + name
})
```

### Contains：檢查切片是否包含某值

```go
func Contains[T comparable](slice []T, target T) bool {
    for _, v := range slice {
        if v == target {
            return true
        }
    }
    return false
}

// 使用
Contains([]int{1, 2, 3}, 2)           // true
Contains([]string{"a", "b"}, "c")     // false
```

### Keys / Values：取出 map 的鍵或值

```go
func Keys[K comparable, V any](m map[K]V) []K {
    keys := make([]K, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    return keys
}

func Values[K comparable, V any](m map[K]V) []V {
    vals := make([]V, 0, len(m))
    for _, v := range m {
        vals = append(vals, v)
    }
    return vals
}
```

## 泛型資料結構

### Stack[T]：先進後出

```go
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    last := len(s.items) - 1
    item := s.items[last]
    s.items = s.items[:last]
    return item, true
}

func (s *Stack[T]) Peek() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    return s.items[len(s.items)-1], true
}

func (s *Stack[T]) Size() int {
    return len(s.items)
}

// 使用
var intStack Stack[int]
intStack.Push(1)
intStack.Push(2)
val, ok := intStack.Pop() // val=2, ok=true

var strStack Stack[string]
strStack.Push("hello")
```

### Queue[T]：先進先出

```go
type Queue[T any] struct {
    items []T
}

func (q *Queue[T]) Enqueue(item T) {
    q.items = append(q.items, item)
}

func (q *Queue[T]) Dequeue() (T, bool) {
    if len(q.items) == 0 {
        var zero T
        return zero, false
    }
    item := q.items[0]
    q.items = q.items[1:]
    return item, true
}

func (q *Queue[T]) IsEmpty() bool {
    return len(q.items) == 0
}
```

## 標準庫的泛型套件（Go 1.21+）

Go 1.21 加入了 `slices`、`maps`、`cmp` 等泛型套件，很多工具函式不用自己寫了：

### `slices` 套件

```go
import "slices"

// 排序
nums := []int{3, 1, 4, 1, 5}
slices.Sort(nums)                    // [1, 1, 3, 4, 5]

// 搜尋
idx, found := slices.BinarySearch(nums, 3)  // idx=2, found=true

// 包含
slices.Contains(nums, 4)            // true

// 反轉
slices.Reverse(nums)                // [5, 4, 3, 1, 1]

// 去重（需先排序）
slices.Compact(nums)                // [5, 4, 3, 1]

// 自訂排序
slices.SortFunc(users, func(a, b User) int {
    return cmp.Compare(a.Age, b.Age)
})
```

### `maps` 套件

```go
import "maps"

// 取所有 key
keys := maps.Keys(myMap)            // iter.Seq[K]

// 複製 map
copied := maps.Clone(myMap)

// 比較兩個 map
maps.Equal(map1, map2)              // bool
```

### `cmp` 套件

```go
import "cmp"

// 比較（回傳 -1, 0, 1）
cmp.Compare(3, 5)                   // -1
cmp.Compare("b", "a")              // 1

// 取最小/最大值
cmp.Or(0, 0, 1)                    // 1（回傳第一個非零值）
```

### 標準庫 vs 自己寫

| 需求 | 標準庫 | 自己寫 |
|------|--------|--------|
| 排序切片 | `slices.Sort` | 不需要 |
| 搜尋切片 | `slices.Contains` | 不需要 |
| 複製 map | `maps.Clone` | 不需要 |
| Map/Filter/Reduce | 沒有（截至 Go 1.23） | 需要自己寫 |
| 自訂資料結構 | 沒有 | 需要自己寫 |

## 何時用泛型？何時不用？

### 適合使用泛型

| 場景 | 範例 | 原因 |
|------|------|------|
| 資料結構 | Stack[T]、Queue[T]、LinkedList[T] | 行為與型別無關 |
| 工具函式 | Map、Filter、Contains、Min | 邏輯一樣，只是型別不同 |
| 數學運算 | Sum、Average、Clamp | 需要對多種數值型別操作 |
| 快取/容器 | Cache[K, V]、Pool[T] | 儲存任意型別 |

### 不適合使用泛型

| 場景 | 建議替代方案 | 原因 |
|------|-------------|------|
| 業務邏輯 | 介面（interface） | 行為不同，不是型別不同 |
| 只有一種型別 | 直接寫具體型別 | 泛型只增加複雜度 |
| 需要不同行為 | 多態（介面 + 實作） | 泛型處理的是相同行為、不同型別 |
| JSON 序列化 | `encoding/json`（用 reflection） | 泛型不擅長處理結構不確定的情況 |

### 過度工程警告

```go
// 過度工程：只有 int 會用到，何必泛型？
func AddOne[T ~int](x T) T {
    return x + 1
}

// 直接寫就好
func AddOne(x int) int {
    return x + 1
}

// 過度工程：業務邏輯差異大，不適合泛型
func Process[T Article | Comment](item T) error {
    // Article 和 Comment 的處理邏輯完全不同
    // 這裡最後還是會用 type switch，泛型沒幫上忙
}

// 用介面更適合
type Processable interface {
    Process() error
}
```

## 效能：泛型 vs `interface{}`

```go
// Benchmark 結果（概略數字，供參考）
//
// BenchmarkSumGeneric-8      500000000    2.3 ns/op    0 B/op    0 allocs/op
// BenchmarkSumInterface-8    200000000    8.1 ns/op    16 B/op   1 allocs/op
```

| 比較項目 | 泛型 | `interface{}` |
|---------|------|---------------|
| 型別檢查 | 編譯期 | 執行期 |
| 記憶體分配 | 無額外分配（大多數情況） | 每次裝箱（boxing）分配一次 |
| 效能 | 接近直接呼叫 | 有 interface dispatch 開銷 |
| 可讀性 | 中（需學習語法） | 差（到處型別斷言） |
| 安全性 | 高（編譯期檢查） | 低（runtime panic 風險） |

> **Go 編譯器的泛型策略**：Go 使用「GC Shape Stenciling」——對每種不同的 GC shape（大致對應記憶體佈局）產生一份特化版本，而非像 C++ 那樣對每個型別都生成一份。這在二進位大小和效能之間取得了平衡。

## 部落格專案：哪裡可以用泛型？

```go
// 1. 通用的分頁回應
type PaginatedResponse[T any] struct {
    Data       []T `json:"data"`
    Total      int `json:"total"`
    Page       int `json:"page"`
    PageSize   int `json:"page_size"`
    TotalPages int `json:"total_pages"`
}

// 使用
func (h *articleHandler) ListArticles(c *gin.Context) {
    articles, total := h.usecase.List(page, pageSize)
    c.JSON(200, PaginatedResponse[domain.Article]{
        Data:       articles,
        Total:      total,
        Page:       page,
        PageSize:   pageSize,
        TotalPages: (total + pageSize - 1) / pageSize,
    })
}

// 2. 通用的快取
type Cache[K comparable, V any] struct {
    mu    sync.RWMutex
    items map[K]cacheItem[V]
}

type cacheItem[V any] struct {
    value     V
    expiresAt time.Time
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    item, ok := c.items[key]
    if !ok || time.Now().After(item.expiresAt) {
        var zero V
        return zero, false
    }
    return item.value, true
}

// 3. 通用的 Repository 方法
func FindByID[T any](db *gorm.DB, id uint) (*T, error) {
    var entity T
    if err := db.First(&entity, id).Error; err != nil {
        return nil, err
    }
    return &entity, nil
}

// 使用
article, err := FindByID[domain.Article](db, 42)
user, err := FindByID[domain.User](db, 1)
```

## FAQ

### Q1：Go 的泛型和 Java/C++ 的泛型有什麼不同？

Go 的泛型比較簡潔（有人說簡陋）。不支援泛型方法（method 不能有自己的型別參數，只有 type 可以）、不支援特化（specialization）、不支援泛型的 operator overloading。Go 的設計哲學是「夠用就好」，避免像 C++ template 那樣的複雜度。

### Q2：為什麼 Go 的泛型用方括號 `[]` 而不是尖括號 `<>`？

因為尖括號在 Go 的語法中會和比較運算子 `<` `>` 產生歧義。例如 `f<T>(x)` 可以被解析為 `f < T > (x)`（兩次比較）。方括號不會有這個問題，雖然和切片/map 的語法有點像，但 Go 編譯器能夠正確區分。

### Q3：什麼是 `~int`？波浪號 `~` 代表什麼？

`~int` 代表「底層型別（underlying type）是 int 的所有型別」。例如 `type Age int` 的底層型別就是 `int`，所以 `Age` 符合 `~int` 約束。如果只寫 `int`（不加 `~`），那 `Age` 就不符合，因為 `Age` 和 `int` 是不同的具名型別。

### Q4：泛型函式可以有多個型別參數嗎？

可以。例如 `func Map[T any, U any](s []T, f func(T) U) []U`。型別參數之間用逗號分隔。實務上最常見的是 1-2 個型別參數，超過 3 個通常代表函式設計過於複雜，應該考慮拆分。

### Q5：為什麼標準庫沒有 Map/Filter/Reduce？

Go 團隊在 Go 1.18 時討論過，但認為這些函式和 Go 的 `for range` 迴圈風格有衝突——Go 傾向明確的迴圈而非函式鏈。不過 Go 1.23 引入了 `iter` 套件和 range-over-function，未來可能會加入。目前社群有 `samber/lo` 等第三方套件提供這些函式。

## 練習

1. 實作泛型函式 `Contains[T comparable](slice []T, target T) bool`
2. 實作泛型函式 `Keys[K comparable, V any](m map[K]V) []K`，回傳 map 的所有 key
3. 實作泛型資料結構 `Queue[T any]`，支援 Enqueue、Dequeue、IsEmpty
4. 用 `cmp.Ordered` 約束實作 `Max[T cmp.Ordered](a, b T) T`
5. 思考：為什麼 `[]any` 和泛型 `[]T` 不一樣？各自的優缺點？

## 下一課預告

下一課我們會學習 **Goroutine 與並行處理**——Go 最強大的特色之一。有了泛型，你可以寫出型別安全的並行工具函式（例如泛型的 channel 操作、並行 Map 等），讓並行程式碼更安全、更易讀。
