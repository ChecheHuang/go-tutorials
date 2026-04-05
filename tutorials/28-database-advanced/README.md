# 第二十八課：資料庫進階（Database Advanced）

> **一句話總結**：學會索引、N+1、Migration 這三個技能，你的資料庫效能可以提升 100 倍，程式也更容易維護。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | **重點**：理解 N+1 問題和索引是後端工程師必備 |
| 🔴 資深工程師 | Query 最佳化、EXPLAIN 分析、資料庫 Migration 策略 |

## 你會學到什麼？

- **索引（Index）**：讓查詢從「全表掃描」變成「精確跳轉」，快 100 倍以上
- **N+1 問題**：最常見的效能殺手，一行 `Preload` 解決
- **原始 SQL（Raw SQL）**：GORM 搞不定的複雜查詢怎麼寫
- **資料庫遷移（Migration）**：用 golang-migrate 管理資料庫版本
- **SavePoint**：交易中的「中間存檔點」
- **查詢優化**：Select、Scopes、FindInBatches、EXPLAIN

## 執行方式

```bash
go run ./tutorials/28-database-advanced
```

## 生活比喻：百科全書的目錄

```
沒有索引的資料庫（沒有目錄的百科全書）：
  找「Go 語言」→ 從第 1 頁翻到最後一頁
  100 萬筆資料 → 掃描 100 萬次

有索引的資料庫（有目錄的百科全書）：
  找「Go 語言」→ 先看目錄，直接翻到第 437 頁
  100 萬筆資料 → 只需幾十次操作（對數時間）

加索引前：SELECT WHERE author_id=42   → 500ms（全表掃描）
加索引後：SELECT WHERE author_id=42   → 0.5ms（索引掃描）
效能提升：1000 倍！
```

## 索引（Index）

### 什麼時候加索引？

| 情況 | 加索引？ | 說明 |
|------|---------|------|
| WHERE 常用的欄位 | ✅ 一定加 | `author_id`、`status`、`country` |
| 外鍵欄位 | ✅ 一定加 | GORM 不會自動加！JOIN 時需要 |
| ORDER BY 常用的欄位 | ✅ 建議加 | `created_at`、`view_count` |
| 唯一性約束 | ✅ 用 uniqueIndex | `email`、`username` |
| 布林值欄位 | ❌ 效果差 | 只有 true/false，選擇性太低 |
| 很少查詢的欄位 | ❌ 不必要 | 佔空間，寫入時還要維護 |

### GORM 中加索引

```go
type Post struct {
    ID       uint   `gorm:"primaryKey"`
    AuthorID uint   `gorm:"not null;index"`           // 普通索引
    Email    string `gorm:"uniqueIndex"`               // 唯一索引
    Status   string `gorm:"index"`                     // 普通索引
    // 複合索引（兩個欄位合用一個索引）
    Category string `gorm:"index:idx_category_status"` // 複合索引（欄位 1）
    Visible  bool   `gorm:"index:idx_category_status"` // 複合索引（欄位 2）
}
```

## N+1 問題（最重要！）

### 問題示範

```go
// ❌ N+1 問題（100 篇文章 = 101 次 SQL）
var posts []Post
db.Limit(100).Find(&posts)               // 1 次：SELECT * FROM posts

for _, post := range posts {
    var author Author
    db.First(&author, post.AuthorID)     // N 次：每篇文章查一次作者
    // SELECT * FROM authors WHERE id = ?
}
// 總共：1 + 100 = 101 次查詢！
```

### Preload 解法

```go
// ✅ Preload（永遠只需要 2 次 SQL）
var posts []Post
db.Preload("Author").Limit(100).Find(&posts)
// SQL 1：SELECT * FROM posts LIMIT 100
// SQL 2：SELECT * FROM authors WHERE id IN (1,2,3,...,100)
// 總共：2 次查詢！

// 多個關聯一起 Preload
db.Preload("Author").Preload("Tags").Find(&posts)
// 3 次 SQL 搞定所有關聯資料

// Preload 加條件（只載入已發布的文章）
db.Preload("Posts", "status = ?", "published").Find(&authors)
```

**N+1 的影響**：

| 資料量 | N+1 次數 | Preload 次數 | 差距 |
|--------|---------|-------------|------|
| 10 筆 | 11 次 | 2 次 | 5.5x |
| 100 筆 | 101 次 | 2 次 | 50x |
| 1000 筆 | 1001 次 | 2 次 | 500x |

## 原始 SQL（Raw SQL）

```go
// 查詢：db.Raw().Scan()
type Stats struct {
    Country     string
    AuthorCount int
}

var stats []Stats
db.Raw(`
    SELECT country, COUNT(*) as author_count
    FROM authors
    GROUP BY country
    ORDER BY author_count DESC
`).Scan(&stats)

// 帶參數（用 ? 防止 SQL Injection）
db.Raw("SELECT * FROM posts WHERE status = ?", "published").Scan(&posts)

// 執行：db.Exec()
db.Exec("UPDATE posts SET view_count = view_count + 1 WHERE id = ?", postID)
```

**SQL Injection 防護**：

```go
// ❌ 危險：直接拼字串
db.Raw("SELECT * FROM users WHERE name = '" + input + "'")
// input = "' OR '1'='1" → 查出所有資料！

// ✅ 安全：用 ? 佔位符
db.Raw("SELECT * FROM users WHERE name = ?", input)
// GORM 自動 escape，完全安全
```

## 資料庫遷移（Migration）

### 為什麼不能只用 AutoMigrate？

| | AutoMigrate | golang-migrate |
|--|-------------|---------------|
| 加欄位 | ✅ | ✅ |
| 刪欄位 | ❌ 不支援 | ✅ |
| 重新命名欄位 | ❌ 不支援 | ✅ |
| 記錄遷移歷史 | ❌ | ✅ |
| 支援 rollback | ❌ | ✅ |
| 適合環境 | 開發 | 正式環境 |

### golang-migrate 使用流程

```bash
# 1. 安裝工具
go install -tags 'sqlite' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 2. 建立遷移檔案（自動產生兩個檔案）
migrate create -ext sql -dir migrations -seq create_posts_table
# → migrations/000001_create_posts_table.up.sql
# → migrations/000001_create_posts_table.down.sql

# 3. 執行遷移（升版到最新）
migrate -database 'sqlite://./blog.db' -path migrations up

# 4. 降版（rollback 1 步）
migrate -database 'sqlite://./blog.db' -path migrations down 1

# 5. 查看當前版本
migrate -database 'sqlite://./blog.db' -path migrations version
```

### 遷移檔案範例

```sql
-- migrations/000001_create_posts_table.up.sql
CREATE TABLE posts (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    title      TEXT NOT NULL,
    author_id  INTEGER NOT NULL,
    status     TEXT DEFAULT 'draft',
    created_at DATETIME,
    FOREIGN KEY (author_id) REFERENCES authors(id)
);

CREATE INDEX idx_posts_author_id ON posts(author_id);
CREATE INDEX idx_posts_status ON posts(status);
```

```sql
-- migrations/000001_create_posts_table.down.sql
DROP TABLE IF EXISTS posts;
```

## 查詢優化技巧

### Select：只取需要的欄位

```go
// ❌ SELECT *（把所有欄位都傳回來，浪費頻寬）
db.Find(&posts)

// ✅ 只取需要的欄位
type PostTitle struct {
    ID    uint
    Title string
}
var titles []PostTitle
db.Model(&Post{}).Select("id, title").Scan(&titles)
```

### 分頁（永遠必須分頁）

```go
func GetPosts(db *gorm.DB, page, pageSize int) ([]Post, int64) {
    var posts []Post
    var total int64

    db.Model(&Post{}).Count(&total)                     // 先算總數
    db.Offset((page-1)*pageSize).Limit(pageSize).       // 分頁
        Order("created_at desc").Find(&posts)

    return posts, total
}
```

### Scopes：可重用的查詢條件

```go
// 定義 Scope
func Published(db *gorm.DB) *gorm.DB {
    return db.Where("status = ?", "published")
}

func RecentFirst(db *gorm.DB) *gorm.DB {
    return db.Order("created_at desc")
}

// 使用 Scope
db.Scopes(Published, RecentFirst).Limit(10).Find(&posts)
```

### FindInBatches：大量資料分批處理

```go
// ❌ 一次查全部（100 萬筆 → 記憶體爆炸）
var allPosts []Post
db.Find(&allPosts)  // 不要這樣！

// ✅ 分批處理（每批 100 筆）
db.FindInBatches(&[]Post{}, 100, func(tx *gorm.DB, batch int) error {
    var batchPosts []Post
    tx.Find(&batchPosts)
    // 處理這批資料...
    return nil  // 繼續下一批
})
```

## EXPLAIN：分析慢查詢

```sql
-- MySQL/PostgreSQL
EXPLAIN SELECT * FROM posts WHERE author_id = 42;
EXPLAIN ANALYZE SELECT * FROM posts WHERE author_id = 42;

-- SQLite
EXPLAIN QUERY PLAN SELECT * FROM posts WHERE author_id = 42;
```

看結果的重點：
- `Using index` → 有用到索引，好！
- `ALL` / `Using filesort` → 全表掃描，需要加索引！

## SavePoint：交易中間存檔

```go
tx := db.Begin()

// 操作 A
tx.Create(&postA)
tx.SavePoint("sp1")  // 建立中間存檔點

// 操作 B（可能失敗）
if err := tx.Create(&badData).Error; err != nil {
    tx.RollbackTo("sp1")  // 回到 sp1（A 保留，B 取消）
}

tx.Commit()  // A 被儲存，B 已取消
```

## 常見問題 FAQ

### Q: 加了索引為什麼查詢還是很慢？

幾個可能原因：
1. **沒有用到索引**：查詢條件沒有以索引欄位開頭（複合索引的最左原則）
2. **資料量太少**：資料庫可能選擇不用索引（小表掃描比用索引還快）
3. **N+1 問題**：索引再快，101 次查詢也比 2 次慢

用 `EXPLAIN` 確認是否有用到索引。

### Q: AutoMigrate 安全嗎？

**開發環境**：安全，很方便。

**正式環境**：小心！AutoMigrate 可能在大表上執行 `ALTER TABLE`，這個操作會鎖表，導致服務暫停。建議：
- 用 golang-migrate 管理遷移
- 大表的 schema 修改在維護視窗執行
- 先在測試環境驗證

### Q: Preload vs Join 哪個比較好？

```go
// Preload：2 次 SQL，結果映射到 struct
db.Preload("Author").Find(&posts)

// Join：1 次 SQL，適合需要 WHERE 篩選關聯欄位的情況
db.Joins("Author").Where("Author.country = ?", "Taiwan").Find(&posts)
```

- 一般情況用 `Preload`（簡單，N 個關聯只多 N 次 SQL）
- 需要 WHERE 篩選關聯欄位時用 `Joins`

## 練習

1. **索引效果實驗**：移除 `Post.AuthorID` 的 `index` tag，用 `EXPLAIN` 比較有無索引的查詢計畫差異
2. **N+1 偵測**：在 GORM 開啟 `Debug()` 模式，數一數 N+1 寫法和 Preload 寫法各執行幾條 SQL
3. **自訂分頁函式**：寫一個 `Paginate(page, pageSize int) func(*gorm.DB) *gorm.DB` 的 Scope 函式，讓分頁可以用 `db.Scopes(Paginate(page, pageSize)).Find(&posts)` 呼叫
4. **Migration 實作**：建立 `migrations/` 目錄，寫三個遷移檔案（建立 authors → 建立 posts → 加 view_count），用 golang-migrate 執行

## 下一課預告

**第二十三課：WebSocket 即時通訊** —— 學習如何用 `gorilla/websocket` 實作聊天室，讓瀏覽器和伺服器保持長連線，實現即時推播。
