# 第十四課：GORM 資料庫操作

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟢 初學者 | ORM 入門，不需要寫 SQL 就能操作資料庫 |
| 🟡 中級工程師 | **重點**：GORM 是 Go 最常用的 ORM，後端必備 |
| 🔴 資深工程師 | 搭配第 22 課，理解 N+1 問題、索引最佳化 |

## 學習目標

- 了解資料庫、SQL、ORM 是什麼
- 學會用 GORM 做完整的 CRUD 操作
- 理解 Preload 和 N+1 問題
- 掌握 Transaction（交易）確保資料一致性
- 了解 AutoMigrate 的能力與限制
- 認識軟刪除（Soft Delete）的概念

## 執行方式

```bash
go run ./tutorials/14-gorm-database
```

## 重點筆記

### 什麼是資料庫？

資料庫就是一個「有組織的資料倉庫」。想像一個 Excel 試算表：
- **資料庫（Database）** = 整個 Excel 檔案
- **資料表（Table）** = 一個工作表（Sheet）
- **欄位（Column）** = 欄（如：姓名、年齡、信箱）
- **記錄（Row）** = 列（如：Alice, 25, alice@test.com）

### 什麼是 SQL？

SQL（Structured Query Language）是和資料庫溝通的語言：
```sql
-- 新增一筆資料
INSERT INTO users (username, email) VALUES ('alice', 'alice@test.com');

-- 查詢所有資料
SELECT * FROM users;

-- 更新資料
UPDATE users SET age = 26 WHERE username = 'alice';

-- 刪除資料
DELETE FROM users WHERE username = 'alice';
```

### 什麼是 ORM？翻譯員比喻

ORM（Object-Relational Mapping）是程式語言和資料庫之間的**翻譯員**：

```
你（Go 程式碼）          翻譯員（GORM）         資料庫（SQLite）

db.Create(&user)    →   翻譯成 SQL    →   INSERT INTO users ...
db.First(&user, 1)  →   翻譯成 SQL    →   SELECT * FROM users WHERE id = 1
db.Save(&user)      →   翻譯成 SQL    →   UPDATE users SET ...
db.Delete(&user)    →   翻譯成 SQL    →   DELETE FROM users WHERE ...
```

**好處：**
- 不需要手寫 SQL（減少出錯機會）
- Go struct 就是資料表的定義（程式碼即文件）
- 自動防止 SQL 注入攻擊
- 換資料庫時不需要改程式碼（例如從 SQLite 換到 PostgreSQL）

### 什麼是 GORM？

GORM 是 Go 語言**最受歡迎**的 ORM 框架：
- GitHub 超過 36,000 顆星
- 支援多種資料庫（SQLite、MySQL、PostgreSQL、SQL Server）
- 功能完整：CRUD、關聯、交易、遷移、Hook 等
- 文件齊全，社群活躍

### 什麼是 SQLite？為什麼選它？

SQLite 是一個**嵌入式資料庫**：
- **不需要安裝**：不像 MySQL/PostgreSQL 需要安裝伺服器
- **檔案型資料庫**：整個資料庫就是一個 `.db` 檔案
- **零設定**：不需要帳號密碼、不需要啟動服務
- **適合學習**：簡單直覺，專注於學習 GORM 操作

我們使用 `github.com/glebarez/sqlite` 而不是 `gorm.io/driver/sqlite`，因為前者是**純 Go 實作**，不需要安裝 C 編譯器（CGO）。

### GORM Model 說明

```go
type User struct {
    ID        uint           `gorm:"primaryKey"`    // 主鍵，自動遞增
    Username  string         `gorm:"uniqueIndex"`   // 唯一索引
    Email     string         `gorm:"not null"`      // 不可為空
    CreatedAt time.Time      // GORM 自動填入建立時間
    UpdatedAt time.Time      // GORM 自動填入更新時間
    DeletedAt gorm.DeletedAt // 軟刪除欄位
}
```

GORM 會自動管理這些特殊欄位：

| 欄位 | 型別 | 說明 |
|------|------|------|
| `ID` | `uint` | 主鍵，自動遞增 |
| `CreatedAt` | `time.Time` | 記錄建立時自動填入當前時間 |
| `UpdatedAt` | `time.Time` | 記錄更新時自動填入當前時間 |
| `DeletedAt` | `gorm.DeletedAt` | 軟刪除時填入時間，查詢時自動過濾 |

### GORM CRUD 速查表

| 操作 | 方法 | 對應 SQL |
|------|------|----------|
| 建立單筆 | `db.Create(&user)` | `INSERT INTO users ...` |
| 建立多筆 | `db.Create(&users)` | `INSERT INTO users ... (多筆)` |
| 主鍵查詢 | `db.First(&user, id)` | `SELECT ... WHERE id = ? LIMIT 1` |
| 條件查詢 | `db.Where("x = ?", v).First(&u)` | `SELECT ... WHERE x = ? LIMIT 1` |
| 查詢全部 | `db.Find(&users)` | `SELECT *` |
| 模糊搜尋 | `db.Where("title LIKE ?", "%Go%")` | `SELECT ... WHERE title LIKE '%Go%'` |
| 排序 | `db.Order("age DESC").Find(&users)` | `SELECT ... ORDER BY age DESC` |
| 分頁 | `db.Limit(10).Offset(20).Find(&users)` | `SELECT ... LIMIT 10 OFFSET 20` |
| 計數 | `db.Model(&User{}).Count(&count)` | `SELECT COUNT(*)` |
| 更新全部 | `db.Save(&user)` | `UPDATE ... SET (所有欄位)` |
| 更新單一 | `db.Model(&u).Update("col", val)` | `UPDATE ... SET col = ?` |
| 更新多個 | `db.Model(&u).Updates(map)` | `UPDATE ... SET col1 = ?, col2 = ?` |
| 刪除 | `db.Delete(&user, id)` | `DELETE ... WHERE id = ?` |

### Preload 和 N+1 問題

**N+1 問題是什麼？** 用餐廳點餐來比喻：

```
不好的方式（N+1）：
  「給我所有桌子的編號」→ 1 次查詢
  「1 號桌點了什麼？」  → 第 2 次查詢
  「2 號桌點了什麼？」  → 第 3 次查詢
  「3 號桌點了什麼？」  → 第 4 次查詢
  ... 共 N+1 次查詢

好的方式（Preload）：
  「給我所有桌子的編號和他們各自點了什麼」→ 2 次查詢搞定
```

**程式碼對比：**

```go
// ❌ 沒有 Preload → user.Articles 是空的
db.First(&user, 1)

// ✅ 有 Preload → user.Articles 有資料
db.Preload("Articles").First(&user, 1)

// ✅ 多層 Preload → 同時載入文章和文章的留言
db.Preload("Articles").Preload("Articles.Comments").First(&user, 1)
```

### Transaction（交易）—— 銀行轉帳比喻

**什麼是交易？** 用銀行轉帳來說明：

```
Alice 要轉 100 元給 Bob

步驟 1：Alice 帳戶 -100 元
步驟 2：Bob 帳戶 +100 元

如果步驟 1 成功、步驟 2 失敗？
→ Alice 的錢扣了，Bob 沒收到 → 錢消失了！💸

交易的作用：
→ 全部成功 → Commit（提交）：兩邊都生效
→ 任一失敗 → Rollback（回滾）：恢復原狀，當作什麼都沒發生
```

**程式碼：**

```go
err := db.Transaction(func(tx *gorm.DB) error {
    // 步驟 1：Alice 扣款
    if err := tx.Model(&User{}).Where("id = ?", aliceID).
        Update("balance", gorm.Expr("balance - ?", 100)).Error; err != nil {
        return err  // 回傳 error → 自動 Rollback
    }

    // 步驟 2：Bob 收款
    if err := tx.Model(&User{}).Where("id = ?", bobID).
        Update("balance", gorm.Expr("balance + ?", 100)).Error; err != nil {
        return err  // 回傳 error → 自動 Rollback
    }

    return nil  // 回傳 nil → 自動 Commit
})
```

**重要規則：**
- 交易函式內必須用 `tx`（不是 `db`），否則操作不在交易範圍內
- 回傳 `nil` → 自動 Commit
- 回傳 `error` → 自動 Rollback

### Migration（遷移）—— AutoMigrate 的能力與限制

**AutoMigrate 會做什麼？**

| 能做 ✅ | 不能做 ❌ |
|---------|----------|
| 建立不存在的表格 | 刪除欄位 |
| 新增缺少的欄位 | 修改欄位型別 |
| 建立索引 | 刪除索引 |
| | 重新命名欄位 |

**為什麼有這些限制？**
- 防止誤刪資料（安全第一）
- 修改欄位型別可能破壞現有資料

**開發 vs 正式環境：**

```
開發環境：AutoMigrate 很方便，快速建表
正式環境：建議用專業遷移工具（如 golang-migrate）
         → 可以追蹤每次變更
         → 可以回滾到之前的版本
         → 更安全可控
```

### 軟刪除（Soft Delete）

```go
// 加入 gorm.DeletedAt 欄位就啟用軟刪除
type Article struct {
    ID        uint
    Title     string
    DeletedAt gorm.DeletedAt  // 有這個欄位就自動啟用軟刪除
}

// 刪除 → 只是標記 deleted_at 時間，資料還在
db.Delete(&article, 1)

// 一般查詢 → 自動排除已刪除的資料
db.Find(&articles)  // WHERE deleted_at IS NULL

// 包含已刪除的資料
db.Unscoped().Find(&articles)  // 不加 WHERE deleted_at IS NULL

// 真正刪除（永久）
db.Unscoped().Delete(&article, 1)
```

### 在 Blog 專案中的對應

`internal/repository/article_repository.go`：
```go
func (r *articleRepository) FindAll(query domain.ArticleQuery) ([]domain.Article, int64, error) {
    db := r.db.Model(&domain.Article{})

    // Where 條件查詢
    if query.Search != "" {
        db = db.Where("title LIKE ?", "%"+query.Search+"%")
    }

    // Count 計算總數
    db.Count(&total)

    // Preload + Order + Offset + Limit = 完整的分頁查詢
    db.Preload("User").
       Order("created_at DESC").
       Offset(offset).
       Limit(query.PageSize).
       Find(&articles)
}
```

## 常見問題 FAQ

### Q: GORM 的 `?` 佔位符和直接拼字串有什麼差別？

```go
// ✅ 安全：用 ? 佔位符，GORM 會自動處理，防止 SQL 注入
db.Where("username = ?", userInput)

// ❌ 危險：直接拼字串，容易被 SQL 注入攻擊
db.Where("username = '" + userInput + "'")
```

### Q: Save 和 Update 差在哪？

- `Save`：更新**所有欄位**（即使沒改的也會寫入）
- `Update/Updates`：只更新**指定的欄位**（效率更好）

### Q: 什麼時候該用 First，什麼時候用 Find？

- `First`：查詢**一筆**記錄（找不到會回傳 error）
- `Find`：查詢**多筆**記錄（找不到回傳空切片，不會 error）

### Q: 為什麼 Create 要傳指標（&user）？

因為 GORM 需要回寫自動產生的 ID。如果傳值（user），GORM 改不到你的變數。

### Q: 軟刪除的資料怎麼恢復？

```go
// 用 Unscoped 找到資料，然後清除 DeletedAt
db.Unscoped().Model(&article).Where("id = ?", id).Update("deleted_at", nil)
```

## 練習

1. **新增 Comment 模型**：建立一個 `Comment` 結構體，包含 `Content`、`ArticleID`、`UserID`，建立與 Article 的一對多關係
2. **多層 Preload**：用 Preload 同時載入文章的作者和留言
3. **分頁查詢**：實作第 2 頁、每頁 5 筆的查詢
4. **交易練習**：寫一個交易，同時建立使用者和他的第一篇文章，如果文章建立失敗，使用者也不會被建立
5. **軟刪除實驗**：刪除一篇文章，然後用 `Unscoped()` 找回來

## 下一課預告

**第十五課：中介層（Middleware）** —— 學習如何在請求到達 Handler 之前加入通用邏輯，像是日誌記錄、身份驗證、跨域設定等。
