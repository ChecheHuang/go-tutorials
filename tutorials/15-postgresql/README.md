# 第十五課：PostgreSQL + Schema Design + Index Optimization

> **一句話總結**：生產環境不用 SQLite——學會 PostgreSQL、設計好的 Schema、和正確的索引，是後端工程師的基本功。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟢 初學者 | 了解為什麼生產環境要用 PostgreSQL，認識資料表正規化 |
| 🟡 中級工程師 | **重點**：理解三大正規化、學會設計資料表、建立正確的索引 |
| 🔴 資深工程師 | **必備**：EXPLAIN 分析、Connection Pooling、Transaction Isolation、效能調校 |

## 學習目標

- 理解 PostgreSQL 與 SQLite 的差異，知道何時該切換
- 掌握 Schema 三大正規化（1NF、2NF、3NF）的設計原則
- 學會建立正確的索引，理解複合索引最左前綴原則
- 讀懂 EXPLAIN ANALYZE 的輸出，判斷查詢效能瓶頸
- 了解 Connection Pooling 的設定與重要性
- 認識 PostgreSQL 特有功能（JSONB、Array、Enum）
- 理解 Transaction Isolation Level 的差異

## 執行方式

```bash
go run ./tutorials/15-postgresql/
```

本課用純 Go 模擬 PostgreSQL 的核心概念（不需要安裝 PostgreSQL），學完後只需換 driver 就能連上真正的 PostgreSQL。

## 重點筆記

### 為什麼生產環境不用 SQLite？

用一個比喻來理解：

```
SQLite   = 個人筆記本 📓
             一個人用很方便，但你不會拿筆記本管理整間公司的帳本

PostgreSQL = 企業級資料庫系統 🏢
             多人同時存取、權限管理、資料安全、備份還原全都有
```

**具體差異：**

| 特性 | SQLite | PostgreSQL |
|------|--------|------------|
| 儲存方式 | 單一檔案 | 獨立服務（進程） |
| 並發寫入 | 不支援 | MVCC 多版本並發控制 |
| 網路連線 | 只能本機 | TCP 遠端連線 |
| 使用者權限 | 無 | 完整 RBAC 權限控制 |
| JSON 支援 | 基本 | JSONB 可建索引 |
| 全文搜尋 | FTS5（基本） | tsvector（強大） |
| 適合場景 | 開發/測試/嵌入式 | 生產環境 |
| 安裝難度 | 零安裝 | 需要安裝或 Docker |
| 資料量 | 適合 < 1GB | 輕鬆處理 TB 級別 |
| 備份還原 | 複製檔案 | pg_dump / pg_restore |

**什麼時候該從 SQLite 切換到 PostgreSQL？**
- 需要多人（多個 API server）同時寫入
- 需要使用者權限控制（不同角色不同權限）
- 資料量超過幾百 MB
- 需要進階功能（JSONB、全文搜尋、Pub/Sub）
- 準備上線到生產環境

### Schema Design 三大正規化

正規化（Normalization）是資料表設計的核心原則，目標是**消除資料重複、避免更新異常**。

#### 反正規化的問題（Bad）

假設部落格系統把所有資料塞在一張表：

```
┌──────────────────────────────────────────────────────────────────┐
│ articles 表                                                      │
│ id │ title    │ user_name │ user_email     │ tag1  │ tag2       │
│ 1  │ Go 入門  │ Alice     │ alice@mail.com │ go    │ tutorial   │
│ 2  │ Docker   │ Alice     │ alice@mail.com │ docker│ devops     │
│        ↑ Alice 資料重複！改 email 要改多行                        │
│        ↑ tag 欄位有上限！想加第三個 tag 怎麼辦？                  │
└──────────────────────────────────────────────────────────────────┘
```

這會造成三種異常：
- **更新異常**：Alice 改 email，每一筆文章都要改
- **插入異常**：想新增使用者但他還沒寫文章？放不進去
- **刪除異常**：刪掉 Alice 唯一的文章，使用者資料也跟著消失

#### 第一正規化（1NF）：每個欄位只存一個值

```
❌ 違反 1NF：
│ id │ title   │ tags              │
│ 1  │ Go 入門 │ go,tutorial,basic │  ← 一個欄位塞了三個值！

✅ 符合 1NF（拆成獨立的 article_tags 表）：
│ article_id │ tag      │
│ 1          │ go       │
│ 1          │ tutorial │
│ 1          │ basic    │
```

**Blog 專案對應**：`article_tags` 多對多關聯表，每個 tag 獨立一行。

#### 第二正規化（2NF）：非主鍵欄位完全依賴主鍵

```
❌ 違反 2NF（複合主鍵 = article_id + tag_id）：
│ article_id │ tag_id │ article_title │ tag_name │
│ 1          │ 1      │ Go 入門       │ go       │
         article_title 只依賴 article_id，不依賴 tag_id → 部分依賴！

✅ 符合 2NF（拆表）：
articles:     │ id │ title   │
tags:         │ id │ name    │
article_tags: │ article_id │ tag_id │
```

#### 第三正規化（3NF）：非主鍵欄位不互相依賴

```
❌ 違反 3NF：
│ id │ title   │ user_id │ user_name │ user_email     │
│ 1  │ Go 入門 │ 1       │ Alice     │ alice@mail.com │
         user_name 和 user_email 依賴 user_id，不是直接依賴主鍵 id

✅ 符合 3NF（拆表 + 外鍵關聯）：
users:    │ id │ name  │ email          │
articles: │ id │ title │ user_id（FK）   │
```

**Blog 專案對應**：`articles.user_id` 是外鍵，關聯到 `users.id`，不在 articles 表存 user_name。

#### 什麼時候可以反正規化？

正規化不是絕對，有時候為了效能可以**故意**反正規化：

```
場景：首頁顯示文章列表 + 作者名稱
正規化：每次都要 JOIN users 表 → 讀取慢
反正規化：在 articles 表冗餘存一個 author_name → 讀取快，但寫入要多維護

原則：先正規化設計，確認有效能瓶頸後再考慮反正規化
```

### Index Deep Dive（索引深入理解）

索引就像書的**目錄**。沒有目錄，找一個詞要從第一頁翻到最後一頁（全表掃描）；有目錄，直接翻到對應頁碼（索引掃描）。

#### 什麼時候該加索引？

```
✅ 該加索引的欄位：
  - WHERE 條件常用的欄位（user_id, email, status）
  - JOIN 的外鍵欄位（articles.user_id）
  - ORDER BY 的排序欄位（created_at）
  - UNIQUE 約束的欄位（email）

❌ 不該加索引的欄位：
  - 很少出現在查詢條件的欄位
  - 值的選擇性太低（如 boolean，只有 true/false）
  - 頻繁更新的欄位（索引維護成本高）
  - 小表（幾百筆以下，全表掃描可能更快）
```

#### 索引的代價

```
✅ 加速 SELECT（讀取快）
❌ 減慢 INSERT / UPDATE / DELETE（每次寫入要更新索引）
❌ 佔用額外儲存空間
❌ 過多索引會讓查詢最佳化器選擇困難
```

#### 常用索引類型

| 索引類型 | SQL 語法 | 適用場景 |
|---------|---------|---------|
| B-Tree（預設） | `CREATE INDEX idx ON users(email)` | `=`, `<`, `>`, `BETWEEN`，大部分場景 |
| Unique Index | `CREATE UNIQUE INDEX ON users(email)` | 確保欄位值不重複 |
| Composite（複合） | `CREATE INDEX ON orders(user_id, created_at)` | 多欄位組合查詢 |
| Partial（部分） | `CREATE INDEX ON orders(status) WHERE status='pending'` | 只索引部分資料，省空間 |
| GIN | `CREATE INDEX ON posts USING GIN(tags)` | JSONB、全文搜尋、陣列查詢 |
| GiST | `CREATE INDEX ON locations USING GiST(geom)` | 地理空間資料、範圍查詢 |

#### 複合索引的最左前綴原則

這是面試常考題。複合索引 `INDEX(a, b, c)` 就像電話簿先按**姓氏**排，再按**名字**排，最後按**地址**排：

```
INDEX(a, b, c) 可以加速：
  ✅ WHERE a = 1                          （查姓氏）
  ✅ WHERE a = 1 AND b = 2                （查姓氏 + 名字）
  ✅ WHERE a = 1 AND b = 2 AND c = 3      （查姓氏 + 名字 + 地址）
  ✅ WHERE a = 1 ORDER BY b               （查姓氏，按名字排序）

  ❌ WHERE b = 2                          （跳過姓氏直接查名字？找不到！）
  ❌ WHERE c = 3                          （跳過姓氏和名字？更找不到！）
  ❌ WHERE b = 2 AND c = 3                （沒有最左欄位 a，索引無效）
```

**設計複合索引的原則：**
- 把**等值查詢**的欄位放前面（`WHERE status = 'active'`）
- 把**範圍查詢**的欄位放後面（`WHERE created_at > '2024-01-01'`）
- 把**選擇性高**的欄位放前面（email 比 status 選擇性高）

#### Covering Index（覆蓋索引）

當索引包含了查詢需要的所有欄位，PostgreSQL 可以**只讀索引不讀資料表**，這叫 Index Only Scan：

```sql
-- 建立覆蓋索引
CREATE INDEX idx_covering ON articles(user_id) INCLUDE (title, created_at);

-- 這個查詢可以直接從索引取得所有資料，不需要回表
SELECT title, created_at FROM articles WHERE user_id = 1;
-- → Index Only Scan（最快！）
```

### EXPLAIN ANALYZE 讀懂查詢計畫

`EXPLAIN ANALYZE` 是 PostgreSQL 效能調校的核心工具，它告訴你資料庫**怎麼執行**你的 SQL：

```sql
EXPLAIN ANALYZE SELECT * FROM articles WHERE user_id = 1;
```

#### 掃描類型對照表

| 掃描類型 | 說明 | 效能 | 何時出現 |
|---------|------|------|---------|
| Seq Scan | 全表掃描，逐行檢查 | 慢 | 沒有索引、或資料量小 |
| Index Scan | 用索引找到行，再回表取完整資料 | 快 | 有索引，且結果集小 |
| Index Only Scan | 只讀索引，不回表 | 最快 | 索引包含所有需要的欄位 |
| Bitmap Index Scan | 先用索引建立 bitmap，再批次讀取 | 中等偏快 | 結果集中等大小 |
| Bitmap Heap Scan | 配合 Bitmap Index Scan 讀取資料頁 | 中等偏快 | 通常和 Bitmap Index Scan 一起出現 |

#### 實際範例解讀

```
沒有索引：
┌─────────────────────────────────────────────────────────┐
│ Seq Scan on articles  (cost=0.00..25.00 rows=5)        │ ← 全表掃描
│   Filter: (user_id = 1)                                │
│   Rows Removed by Filter: 995                          │ ← 掃了 1000 行只取 5 行
│   Planning Time: 0.1 ms                                │
│   Execution Time: 2.5 ms                               │
└─────────────────────────────────────────────────────────┘

加上索引後：
┌─────────────────────────────────────────────────────────┐
│ Index Scan using idx_articles_user  (cost=0.00..8.00)  │ ← 索引掃描
│   Index Cond: (user_id = 1)                            │
│   Planning Time: 0.1 ms                                │
│   Execution Time: 0.05 ms                              │ ← 快 50 倍！
└─────────────────────────────────────────────────────────┘
```

**如何解讀 cost：**
- `cost=0.00..25.00`：第一個數字是啟動成本，第二個是總成本
- `rows=5`：預估回傳的行數
- `actual time=0.01..2.5`：實際執行時間（毫秒）
- `Planning Time`：查詢計畫的規劃時間
- `Execution Time`：實際執行時間

**效能調校流程：**
1. 先用 `EXPLAIN ANALYZE` 看查詢計畫
2. 如果是 Seq Scan → 考慮加索引
3. 加索引後再跑一次 `EXPLAIN ANALYZE` 確認變成 Index Scan
4. 注意 `Rows Removed by Filter` 如果很大，代表索引不夠精準

### GORM 切換 PostgreSQL（只需改 2 行）

這就是 Clean Architecture 的好處——Repository 層以上的程式碼完全不用改：

```go
// ── 目前（SQLite）──────────────────────────
import "github.com/glebarez/sqlite"

db, _ := gorm.Open(sqlite.Open("blog.db"))

// ── 切換 PostgreSQL（只改 import 和 Open）──
import "gorm.io/driver/postgres"

dsn := "host=localhost user=postgres password=xxx dbname=blog port=5432 sslmode=disable"
db, _ := gorm.Open(postgres.Open(dsn))
```

### Connection Pooling（連線池）

資料庫連線是**昂貴的資源**。每次查詢都開新連線就像每次打電話都重新撥號——慢且浪費。連線池預先建好一批連線，需要時借出，用完歸還。

```go
sqlDB, err := db.DB() // 取得底層 *sql.DB

// 最大開啟連線數（同時最多幾條連線）
sqlDB.SetMaxOpenConns(25)

// 最大閒置連線數（保持幾條連線待命）
sqlDB.SetMaxIdleConns(5)

// 連線最大存活時間（避免使用過期的連線）
sqlDB.SetConnMaxLifetime(5 * time.Minute)

// 閒置連線最大存活時間
sqlDB.SetConnMaxIdleTime(1 * time.Minute)
```

**參數怎麼設定？**

| 參數 | 建議值 | 說明 |
|------|-------|------|
| `MaxOpenConns` | 25-50 | 取決於 PostgreSQL 的 `max_connections`（預設 100） |
| `MaxIdleConns` | 5-10 | 太少會頻繁建立新連線，太多浪費記憶體 |
| `ConnMaxLifetime` | 5 分鐘 | 避免使用被 PostgreSQL 關閉的連線 |
| `ConnMaxIdleTime` | 1 分鐘 | 回收長時間閒置的連線 |

**常見問題：**
- `MaxOpenConns` 設太高 → PostgreSQL 連線耗盡，其他服務連不上
- `MaxOpenConns` 設太低 → 請求排隊等待連線，回應變慢
- 沒設 `ConnMaxLifetime` → 可能用到被資料庫關閉的連線，導致隨機錯誤

### Migration 工具：AutoMigrate vs golang-migrate

| 特性 | AutoMigrate | golang-migrate |
|------|------------|----------------|
| 建立新表 | ✅ | ✅ |
| 新增欄位 | ✅ | ✅ |
| 刪除欄位 | ❌ | ✅ |
| 修改欄位型別 | ❌ | ✅ |
| 版本控制 | ❌ | ✅ 每次變更一個版本 |
| 回滾（Rollback） | ❌ | ✅ 可以回到任一版本 |
| 團隊協作 | ⚠️ 容易衝突 | ✅ 每人一個 migration 檔 |
| 適合場景 | 開發/原型 | 正式環境 |

**golang-migrate 範例：**

```bash
# 安裝
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 建立 migration 檔案
migrate create -ext sql -dir migrations -seq add_index_to_articles

# 會產生兩個檔案：
# migrations/000001_add_index_to_articles.up.sql   ← 升級
# migrations/000001_add_index_to_articles.down.sql ← 回滾

# 執行 migration
migrate -path migrations -database "postgres://localhost/blog?sslmode=disable" up

# 回滾一個版本
migrate -path migrations -database "postgres://localhost/blog?sslmode=disable" down 1
```

### Transaction Isolation Levels（交易隔離等級）

PostgreSQL 支援多種隔離等級，解決不同的並發問題：

| 隔離等級 | Dirty Read | Non-Repeatable Read | Phantom Read | 效能 |
|---------|-----------|-------------------|-------------|------|
| Read Uncommitted | PostgreSQL 實際等同 Read Committed | - | - | - |
| **Read Committed**（預設） | ✅ 防止 | ❌ 可能 | ❌ 可能 | 快 |
| Repeatable Read | ✅ 防止 | ✅ 防止 | ❌ 可能 | 中等 |
| **Serializable** | ✅ 防止 | ✅ 防止 | ✅ 防止 | 慢 |

**用白話解釋：**

```
Read Committed（預設）：
  每次 SELECT 都看到「最新已提交」的資料
  適合大部分場景

Serializable：
  交易執行起來就像「一個接一個」，完全不互相干擾
  適合金融轉帳、庫存扣減等需要強一致性的場景
  代價：衝突時會自動 rollback，需要應用層重試
```

**GORM 設定隔離等級：**

```go
// 使用 Serializable 隔離等級
db.Transaction(func(tx *gorm.DB) error {
    // 交易邏輯
    return nil
}, &sql.TxOptions{Isolation: sql.LevelSerializable})
```

### PostgreSQL 特有功能

#### JSONB 欄位

PostgreSQL 的 JSONB 可以存結構化資料，還能建索引：

```go
// GORM Model
type Article struct {
    ID       uint
    Title    string
    Metadata datatypes.JSON `gorm:"type:jsonb"` // 儲存額外資訊
}

// 插入
db.Create(&Article{
    Title:    "Go 入門",
    Metadata: datatypes.JSON(`{"views": 100, "likes": 25}`),
})

// 查詢 JSONB 欄位
db.Where("metadata->>'views' > ?", "50").Find(&articles)
```

#### Array 欄位

```go
// PostgreSQL 原生支援陣列
type Article struct {
    ID   uint
    Tags pq.StringArray `gorm:"type:text[]"` // PostgreSQL Array
}

// 查詢包含某個 tag 的文章
db.Where("? = ANY(tags)", "go").Find(&articles)
```

#### Enum 型別

```sql
-- PostgreSQL 可以建立自定義型別
CREATE TYPE article_status AS ENUM ('draft', 'published', 'archived');

-- 使用 Enum 欄位
ALTER TABLE articles ADD COLUMN status article_status DEFAULT 'draft';
```

### Blog 專案如何對應 PostgreSQL 概念

```
Blog 專案結構 → PostgreSQL 概念對應：

users 表          → 3NF 設計，email 有 Unique Index
articles 表       → user_id 外鍵有 B-Tree Index
article_tags 表   → 多對多關聯表，符合 1NF
                  → (article_id, tag_id) Composite Index

查詢最佳化：
  文章列表頁  → articles.created_at 索引（ORDER BY 排序）
  作者篩選    → articles.user_id 索引（WHERE 條件）
  搜尋功能    → articles.title 可考慮 GIN 索引（全文搜尋）

連線設定：
  開發環境 → SQLite，零設定
  生產環境 → PostgreSQL + Connection Pooling
  切換方式 → 只改 main.go 的 DB 初始化（Clean Architecture 好處）
```

## 常見問題 FAQ

### Q: PostgreSQL 一定要裝在本機嗎？

不用。最簡單的方式是用 Docker：
```bash
docker run --name postgres -e POSTGRES_PASSWORD=mysecret -p 5432:5432 -d postgres:16
```
一行指令就有一個可用的 PostgreSQL。雲端服務也可以用 Supabase、Neon、Railway 等免費方案。

### Q: 索引加越多越好嗎？

不是。每個索引都會：
- 佔用磁碟空間
- 每次 INSERT/UPDATE/DELETE 都要同步更新
- 過多索引會讓查詢最佳化器「選擇困難」

原則：只為**常用的查詢條件**加索引，用 `EXPLAIN ANALYZE` 驗證效果。

### Q: 複合索引 (a, b) 和分別建 (a) + (b) 兩個索引有什麼差別？

複合索引 `(a, b)` 可以同時滿足 `WHERE a = 1 AND b = 2` 的查詢，用一次索引掃描完成。
分開的兩個索引，PostgreSQL 可能用 Bitmap Index Scan 合併，但效率通常不如複合索引。

### Q: 什麼是 N+1 問題？跟索引有關嗎？

N+1 是 ORM 的問題（第十四課有詳細說明），跟索引是不同層面。索引解決的是「單次查詢的效率」，N+1 解決的是「查詢次數太多」。兩者都要注意。

### Q: GORM 的 AutoMigrate 可以用在正式環境嗎？

不建議。AutoMigrate 只能新增，不能刪除或修改欄位，也沒有版本控制。正式環境建議用 `golang-migrate` 或 `goose` 等專業遷移工具。

## 練習

1. **設計 Schema**：為一個電商系統設計資料表（users、products、orders、order_items），確保符合 3NF，畫出 ER 關係圖
2. **索引設計**：針對以下查詢，設計合適的索引：
   - `SELECT * FROM orders WHERE user_id = ? AND status = 'pending' ORDER BY created_at DESC`
   - `SELECT * FROM products WHERE category = ? AND price BETWEEN ? AND ?`
3. **EXPLAIN 分析**：在 PostgreSQL 中建立一張 10 萬筆資料的表，分別在有索引和無索引的情況下跑 `EXPLAIN ANALYZE`，比較 Seq Scan 和 Index Scan 的差異
4. **Connection Pooling**：在 Blog 專案中加入 Connection Pooling 設定，寫一個壓力測試驗證連線池的效果
5. **Migration 實作**：用 `golang-migrate` 為 Blog 專案建立第一個 migration 檔案，包含建表和建索引的 SQL

## 下一課預告

**第十六課：Error Wrapping** — 用 `fmt.Errorf %w` 和 `errors.Is/As` 建立完整的錯誤鏈，讓錯誤訊息既有上下文又能精準判斷類型。
