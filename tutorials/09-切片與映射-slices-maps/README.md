# 第九課：Slice 與 Map

## 學習目標

- 分辨陣列（Array）和切片（Slice）的差異
- 掌握切片的建立、存取、`append`、切片操作
- 學會 Map 的 CRUD 操作與存在性檢查
- 理解 `make` 的用途

## 執行方式

```bash
cd tutorials/09-slices-maps
go run main.go
```

## 重點筆記

### Slice 速查表

```go
s := []int{1, 2, 3}         // 建立
s = append(s, 4)             // 新增
s[0]                         // 存取
s[1:3]                       // 切片 [2, 3]
len(s)                       // 長度
cap(s)                       // 容量
make([]int, 0, 10)           // 預分配
```

### Map 速查表

```go
m := map[string]int{}        // 建立空 Map
m := make(map[string]int)    // 用 make 建立
m["key"] = 42                // 新增/修改
val := m["key"]              // 存取
val, ok := m["key"]          // 存取 + 存在性檢查
delete(m, "key")             // 刪除
```

### 在專案中的應用

**Slice 在 Domain 層：**
```go
type Article struct {
    Comments []Comment  // 切片：一篇文章有多個留言
}

func FindAll(query ArticleQuery) ([]Article, int64, error)
// 回傳 []Article 切片
```

**Map 在測試的 Mock Repository：**
```go
type mockUserRepository struct {
    users map[string]*domain.User  // 用 Map 模擬資料庫
}
```

## 練習

1. 寫一個 `removeDuplicates(nums []int) []int` 函式（提示：用 Map 去重）
2. 建立一個學生分數的 Map，計算平均分數
3. 實作 `reverse(s []int) []int`，不修改原始切片
