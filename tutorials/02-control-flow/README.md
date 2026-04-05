# 第二課：控制流程（Control Flow）

> **一句話總結**：控制流程就是讓程式能「做判斷」和「重複做事」。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟢 初學者 | **入門必修**：Go 的 for/if/switch 語法 |
| 🟡 中級工程師 | 注意 Go 沒有 while，以及 for range 的用法 |

## 你會學到什麼？

- `if / else if / else`：讓程式做判斷
- `for` 迴圈：讓程式重複做某件事
- `for range`：一個一個走訪集合中的元素
- `switch`：多重選擇（像選擇題）
- `break` 和 `continue`：控制迴圈的行為
- 標籤式 `break`：跳出巢狀迴圈

## 執行方式

```bash
go run ./tutorials/02-control-flow
```

## 用生活來理解

### if/else = 做決定

```
早上出門前：
  如果（下雨）{
      帶雨傘
  } 否則如果（太陽很大）{
      帶陽傘
  } 否則 {
      什麼都不帶
  }
```

### for = 重複做事

```
寫作業：
  從第 1 題開始，只要還沒寫到第 10 題，就繼續寫：
      寫這一題
      翻到下一題
```

### switch = 選擇題

```
今天星期幾？
  星期一~五 → 上班
  星期六、日 → 放假
```

## Go 與其他語言的重要差異

| 特性 | Go | JavaScript / Java / C |
|------|-----|----------------------|
| if 的小括號 | **不需要** `()` | 需要 `()` |
| if 的大括號 | **必須** `{}` | 有時可以省略 |
| while 迴圈 | **沒有** `while`，用 `for` 代替 | 有 `while` |
| switch 的 break | **不需要**，自動 break | 需要手動寫 `break` |
| for range | `for i, v := range ...` | `for...of` / `forEach` |

## if 條件判斷

### 基本語法

```go
if 條件 {
    // 條件成立時執行
} else if 另一個條件 {
    // 第一個不成立、這個成立時執行
} else {
    // 全部不成立時執行
}
```

### if + 初始化語句（Go 獨有）

```go
// 分號前是「初始化」，分號後是「條件」
if remainder := 10 % 3; remainder == 0 {
    fmt.Println("整除")
} else {
    fmt.Println("餘數是", remainder)
}
// remainder 出了 if/else 就不能用了（作用域限制）
```

> **為什麼要這樣設計？** 因為 `remainder` 只在 if 裡面有用，Go 不希望它「汙染」外面的程式碼。這是 Go「作用域越小越好」的設計哲學。

## for 迴圈

Go **只有 `for`**，沒有 `while`、`do-while`。但 `for` 可以寫成各種形式：

### 四種 for 的寫法

```go
// 形式 1：標準 for（最常見）
for i := 0; i < 10; i++ {
    // i 從 0 開始，每次加 1，到 9 結束
}

// 形式 2：類似 while（只寫條件）
for count < 10 {
    count++
}

// 形式 3：無限迴圈（什麼都不寫）
for {
    // 永遠不會自己停，必須用 break 跳出
    break
}

// 形式 4：for range（遍歷集合）
for index, value := range collection {
    // 一個一個拜訪
}
```

### break vs continue

```
for 迴圈在跑 [1, 2, 3, 4, 5]：

  遇到 break    → 整個迴圈立刻結束，不再繼續
  遇到 continue → 跳過「這一次」，直接進入下一次
```

```go
// break：找到 3 就停
for i := 1; i <= 5; i++ {
    if i == 3 {
        break  // 印出 1, 2 後停止
    }
    fmt.Println(i)
}

// continue：跳過 3
for i := 1; i <= 5; i++ {
    if i == 3 {
        continue  // 跳過 3，印出 1, 2, 4, 5
    }
    fmt.Println(i)
}
```

### 標籤式 break（跳出巢狀迴圈）

```go
outer:  // ← 這是一個標籤
    for i := 0; i < 3; i++ {
        for j := 0; j < 3; j++ {
            if i == 1 && j == 2 {
                break outer  // 跳出「外層」迴圈
            }
        }
    }
// 如果只寫 break（沒有 outer），只會跳出內層迴圈
```

## `_` 空白識別符

Go 規定：**宣告了變數就一定要用，否則編譯錯誤**。

如果 `for range` 回傳的 index 你不需要，用 `_`（底線）告訴 Go「我知道有這個值，但我不要」：

```go
// ❌ 編譯錯誤：index 宣告了但沒用
for index, value := range items {
    fmt.Println(value)
}

// ✅ 用 _ 忽略 index
for _, value := range items {
    fmt.Println(value)
}
```

## 在部落格專案中的應用

### 1. if 條件判斷（handler 層的參數驗證）

```go
// internal/handler/article_handler.go
if err := c.ShouldBindJSON(&req); err != nil {
    response.BadRequest(c, "請求參數驗證失敗")
    return  // 提早返回，不繼續往下
}
```

### 2. if 條件組合查詢（repository 層）

```go
// internal/repository/article_repository.go
if query.Search != "" {
    db = db.Where("title LIKE ?", "%"+query.Search+"%")
}
if query.UserID > 0 {
    db = db.Where("user_id = ?", query.UserID)
}
```

### 3. switch 判斷 HTTP 狀態碼（response 套件）

```go
// pkg/response/response.go 的概念
// 根據不同情況回傳不同的 HTTP 狀態碼：
//   Success     → 200
//   Created     → 201
//   BadRequest  → 400
//   NotFound    → 404
```

## 常見問題

### Q: 為什麼 Go 沒有 while？
Go 的設計哲學是「**少即是多**」。既然 `for` 能做到 `while` 的所有事，就不需要多一個關鍵字。減少語言的複雜度，讓每個人寫出來的程式碼長得差不多。

### Q: switch 不寫 break 不會「穿透」到下一個 case 嗎？
不會！Go 的 switch 預設就是「每個 case 執行完自動跳出」。如果你真的想穿透（很少用），要寫 `fallthrough`。

### Q: for range 的 index 和 value 是複製還是引用？
是**複製**。修改 value 不會影響原始資料。這個概念在第 5 課（指標）和第 9 課（切片）會更詳細說明。

## 練習題

1. **基礎題**：寫一個 `for` 迴圈，印出 1 到 100 中所有 3 的倍數
2. **switch 題**：寫一個 `switch`，根據月份（1-12）印出它屬於哪個季節
3. **range 題**：用 `for range` 遍歷字串 `"你好世界"`，印出每個中文字元和它的索引
4. **進階題**：用巢狀 for 迴圈印出九九乘法表

## 下一課預告

學會了控制流程，接下來要學**函式**——把程式碼「包裝」起來重複使用！
