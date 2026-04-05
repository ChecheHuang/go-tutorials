# 第二課：控制流程

## 學習目標

- 學會 `if` / `else if` / `else` 條件判斷
- 掌握 `for` 迴圈的各種寫法
- 了解 `for range` 遍歷語法
- 學會 `switch` 多條件分支

## 執行方式

```bash
cd tutorials/02-control-flow
go run main.go
```

## 重點筆記

### Go 與其他語言的差異

| 特性 | Go | JavaScript / Java |
|------|-----|-------------------|
| if 條件 | 不需要 `()` | 需要 `()` |
| 大括號 | 必須 `{}` | 有時可省略 |
| while 迴圈 | 沒有，用 `for` 代替 | 有 `while` |
| switch break | 自動 break | 需要手動 `break` |
| for range | `for i, v := range` | `for...of` / `forEach` |

### `_` 空白識別符

```go
// 當你不需要某個回傳值時，用 _ 忽略它
for _, value := range items {
    // 不需要 index，用 _ 忽略
}
```

Go 不允許有「宣告了但沒使用」的變數，否則會編譯錯誤。`_` 是告訴編譯器「我知道有這個值，但我不需要」。

### 在專案中的應用

在 `internal/repository/article_repository.go` 中，我們用條件判斷來組合查詢：

```go
if query.Search != "" {
    db = db.Where("title LIKE ?", "%"+query.Search+"%")
}
if query.UserID > 0 {
    db = db.Where("user_id = ?", query.UserID)
}
```

## 練習

1. 寫一個 `for` 迴圈，印出 1 到 100 中所有 3 的倍數
2. 用 `switch` 判斷一個月份屬於哪個季節
3. 用 `for range` 遍歷字串 `"你好世界"` 並印出每個字元
