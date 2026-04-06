// ==========================================================================
// 第二十八課：資料庫進階（Database Advanced）
// ==========================================================================
//
// 第十四課學了 GORM 的基本 CRUD（新增、查詢、更新、刪除）
// 這一課學更進階的主題：
//
//   1. 索引（Index）    → 讓查詢從「全表掃描」變成「精確跳轉」
//   2. N+1 問題        → 最常見的效能殺手，Preload 一行解決
//   3. 原始 SQL        → GORM 做不到的複雜查詢，直接寫 SQL
//   4. 資料庫遷移      → 用版本控制管理資料庫結構變更
//   5. 交易（Transaction）進階 → SavePoint、巢狀交易
//   6. 查詢優化技巧    → EXPLAIN、Batch Insert、分頁
//
// 生活比喻：
//   沒有索引的資料庫 = 一本沒有目錄的百科全書
//   每次找東西都要從第一頁翻到最後一頁
//   有了索引 = 先看目錄，直接翻到正確頁碼
//
// 執行方式：go run ./tutorials/22-database-advanced
// ==========================================================================

package main // 宣告這是 main 套件

import (
	"errors" // 標準錯誤處理
	"fmt"    // 格式化輸出
	"log"    // 日誌輸出
	"time"   // 時間處理

	"github.com/glebarez/sqlite" // 純 Go 的 SQLite 驅動程式
	"gorm.io/gorm"               // GORM 核心框架
	"gorm.io/gorm/logger"        // GORM 日誌設定
)

// ==========================================================================
// 模型定義（帶有索引示範）
// ==========================================================================

// Author 作者模型（示範索引設計）
type Author struct {
	ID        uint      `gorm:"primaryKey"`                    // 主鍵：自動遞增
	Name      string    `gorm:"size:100;not null;index"`       // 姓名：加索引（按名字查詢很常見）
	Email     string    `gorm:"size:100;uniqueIndex;not null"` // Email：唯一索引（每人 email 不重複）
	Country   string    `gorm:"size:50;index"`                 // 國家：加索引（常用來篩選）
	CreatedAt time.Time `gorm:"index"`                         // 建立時間：加索引（常按時間排序）
	Posts     []Post    `gorm:"foreignKey:AuthorID"`           // 一對多：一個作者有多篇文章
}

// Tag 標籤模型
type Tag struct {
	ID    uint   `gorm:"primaryKey"`          // 主鍵
	Name  string `gorm:"size:50;uniqueIndex"` // 標籤名稱：唯一索引
	Posts []Post `gorm:"many2many:post_tags"` // 多對多：一個標籤可以有很多文章
}

// Post 文章模型（示範複合索引）
type Post struct {
	ID        uint      `gorm:"primaryKey"`                    // 主鍵
	Title     string    `gorm:"size:200;not null"`             // 標題
	Content   string    `gorm:"type:text"`                     // 內容（長文字）
	AuthorID  uint      `gorm:"not null;index"`                // 外鍵：加索引（JOIN 時會用到）
	Status    string    `gorm:"size:20;default:'draft';index"` // 狀態：加索引（常按狀態篩選）
	ViewCount int       `gorm:"default:0"`                     // 瀏覽數
	Author    Author    `gorm:"foreignKey:AuthorID"`           // 關聯：文章屬於某個作者
	Tags      []Tag     `gorm:"many2many:post_tags"`           // 多對多：文章有多個標籤
	CreatedAt time.Time `gorm:"index"`                         // 建立時間：加索引
	UpdatedAt time.Time // 更新時間
}

// ==========================================================================
// 1. 索引（Index）
// ==========================================================================
//
// 索引是資料庫最重要的效能優化工具
//
// 沒有索引的查詢（全表掃描）：
//   SELECT * FROM posts WHERE author_id = 42
//   → 資料庫從第 1 筆掃到最後一筆，找到所有 author_id=42 的資料
//   → 100 萬筆資料 = 掃 100 萬次
//
// 有索引的查詢（索引掃描）：
//   → 資料庫用索引（類似 B-tree）直接跳到 author_id=42 的位置
//   → 100 萬筆資料 = 只需幾十次操作（log n）
//
// 索引的代價：
//   好處：SELECT 快很多
//   壞處：INSERT/UPDATE/DELETE 稍慢（因為要同時更新索引）
//   空間：索引本身也佔用磁碟空間
//
// 什麼欄位應該加索引？
//   ✅ WHERE 子句常用的欄位（author_id、status、country）
//   ✅ JOIN 的外鍵欄位（GORM 不會自動加！）
//   ✅ ORDER BY 常用的欄位（created_at）
//   ✅ 唯一性約束的欄位（email、username）
//
//   ❌ 資料種類很少的欄位（is_deleted true/false 只有兩種值，索引效果差）
//   ❌ 很少查詢的欄位（加了也沒用，只佔空間）

// demonstrateIndexes 示範索引的效果
func demonstrateIndexes(db *gorm.DB) { // 示範索引
	fmt.Println("=== 1. 索引（Index）===\n") // 印出標題

	// 在 GORM 的 struct tag 中加索引：
	// `gorm:"index"`              → 普通索引
	// `gorm:"uniqueIndex"`        → 唯一索引（資料不能重複）
	// `gorm:"index:idx_compound"` → 複合索引（多個欄位共用一個索引名稱）

	// 建立測試資料
	authors := []Author{ // 建立多個作者
		{Name: "Alice", Email: "alice@example.com", Country: "Taiwan"},
		{Name: "Bob", Email: "bob@example.com", Country: "Japan"},
		{Name: "Carol", Email: "carol@example.com", Country: "Taiwan"},
		{Name: "Dave", Email: "dave@example.com", Country: "USA"},
		{Name: "Eve", Email: "eve@example.com", Country: "Taiwan"},
	}

	// Batch Insert（批量插入）：一次插入多筆，比逐筆插入快很多
	// GORM 預設批量大小是 100，超過會自動分批
	result := db.CreateInBatches(authors, 100) // 最多每批 100 筆
	if result.Error != nil {                   // 如果插入失敗
		log.Printf("批量插入作者失敗: %v", result.Error) // 印出錯誤
		return                                   // 提前返回
	}
	fmt.Printf("✅ 批量插入 %d 位作者\n", result.RowsAffected) // 印出成功訊息

	// 建立測試文章
	posts := []Post{ // 建立多篇文章
		{Title: "Go 入門指南", AuthorID: authors[0].ID, Status: "published"},
		{Title: "Go 並發教學", AuthorID: authors[0].ID, Status: "published"},
		{Title: "Go 測試最佳實踐", AuthorID: authors[1].ID, Status: "draft"},
		{Title: "Redis 快取策略", AuthorID: authors[2].ID, Status: "published"},
		{Title: "Docker 容器化", AuthorID: authors[3].ID, Status: "published"},
	}
	db.CreateInBatches(posts, 100) // 批量插入文章

	// 用索引欄位查詢（快）
	var taiwanAuthors []Author                                     // 儲存查詢結果
	db.Where("country = ?", "Taiwan").Find(&taiwanAuthors)         // 用 country 索引查詢
	fmt.Printf("✅ 台灣作者（用 country 索引）: %d 人\n", len(taiwanAuthors)) // 印出結果

	// 用索引欄位查詢文章（快）
	var publishedPosts []Post // 儲存查詢結果
	db.Where("status = ? AND author_id = ?", "published", authors[0].ID).Find(&publishedPosts)
	fmt.Printf("✅ Alice 的已發布文章（用 status+author_id 索引）: %d 篇\n", len(publishedPosts))

	fmt.Println("\n💡 索引設計原則：")
	fmt.Println("  WHERE 常用的欄位加索引（status, author_id, country）")
	fmt.Println("  外鍵欄位必加索引（JOIN 時用到，GORM 不會自動加！）")
	fmt.Println("  唯一性欄位用 uniqueIndex（email, username）")
	fmt.Println("  索引不是越多越好：寫入時需要維護索引，消耗效能")
}

// ==========================================================================
// 2. N+1 查詢問題（最常見的效能殺手）
// ==========================================================================
//
// 什麼是 N+1 問題？
//
//   假設你要查「10 篇文章和它們的作者」
//
//   ❌ N+1 寫法（錯誤）：
//     1. SELECT * FROM posts LIMIT 10           → 1 次查詢
//     2. SELECT * FROM authors WHERE id = 1     → 第 1 篇文章的作者
//     3. SELECT * FROM authors WHERE id = 2     → 第 2 篇文章的作者
//     ...（重複 10 次）
//     總共：1 + 10 = 11 次查詢！
//
//   ✅ Preload 寫法（正確）：
//     1. SELECT * FROM posts LIMIT 10                    → 1 次
//     2. SELECT * FROM authors WHERE id IN (1,2,3,...) → 1 次
//     總共：2 次查詢！
//
// N+1 問題在資料量大時非常嚴重：
//   1000 篇文章 → 1001 次查詢 vs 2 次查詢
//   每次查詢耗時 1ms → 1001ms vs 2ms

// demonstrateNPlusOne 示範 N+1 問題和解法
func demonstrateNPlusOne(db *gorm.DB) { // 示範 N+1 問題
	fmt.Println("\n=== 2. N+1 查詢問題 ===\n") // 印出標題

	// 開啟 GORM 的 SQL 日誌，讓你看到實際執行的 SQL 數量
	debugDB := db.Debug() // Debug 模式會印出每一條 SQL

	// ---- 方法 A：會產生 N+1 問題的寫法（DON'T DO THIS！）----
	fmt.Println("❌ 【N+1 問題示範】看看下面執行了幾條 SQL：")

	var posts []Post              // 儲存查詢到的文章
	debugDB.Limit(3).Find(&posts) // 查詢 3 篇文章（1 次 SQL）

	for _, post := range posts { // 遍歷每篇文章
		var author Author                     // 儲存作者
		debugDB.First(&author, post.AuthorID) // 每次迴圈都查一次資料庫！（N 次 SQL）
		fmt.Printf("  文章「%s」→ 作者：%s\n", post.Title, author.Name)
	}
	// ↑ 結果：1 次（文章）+ 3 次（作者）= 4 次 SQL

	// ---- 方法 B：用 Preload 解決（DO THIS！）----
	fmt.Println("\n✅ 【Preload 解法】只需 2 次 SQL：")

	var postsWithAuthors []Post                                // 儲存帶有作者資料的文章
	debugDB.Preload("Author").Limit(3).Find(&postsWithAuthors) // Preload 自動批量查詢
	// ↑ GORM 自動執行：
	//   SELECT * FROM posts LIMIT 3
	//   SELECT * FROM authors WHERE id IN (1,2,3)  ← 只有這一條！

	for _, post := range postsWithAuthors { // 遍歷文章（不需要再查資料庫）
		fmt.Printf("  文章「%s」→ 作者：%s（資料已在記憶體中，不需要查 DB）\n",
			post.Title, post.Author.Name)
	}

	// ---- Preload 進階用法 ----
	fmt.Println("\n✅ 【Preload 進階】載入多個關聯：")

	var fullPosts []Post // 儲存完整文章資料
	debugDB.
		Preload("Author").                // 預載入作者
		Preload("Tags").                  // 預載入標籤（多對多）
		Where("status = ?", "published"). // 只查已發布的
		Limit(2).                         // 最多 2 篇
		Find(&fullPosts)                  // 執行查詢

	for _, post := range fullPosts { // 遍歷結果
		tagNames := make([]string, len(post.Tags)) // 收集標籤名稱
		for i, tag := range post.Tags {
			tagNames[i] = tag.Name
		}
		fmt.Printf("  文章「%s」作者：%s，標籤：%v\n",
			post.Title, post.Author.Name, tagNames)
	}

	// ---- Preload 條件篩選 ----
	fmt.Println("\n✅ 【Preload 條件篩選】只載入特定關聯資料：")

	var authorsWithPosts []Author // 儲存作者（帶文章）
	debugDB.
		Preload("Posts", "status = ?", "published"). // 只預載入已發布的文章
		Find(&authorsWithPosts)                      // 查詢所有作者

	for _, author := range authorsWithPosts { // 遍歷作者
		fmt.Printf("  %s 有 %d 篇已發布文章\n", author.Name, len(author.Posts))
	}
}

// ==========================================================================
// 3. 原始 SQL（Raw SQL）
// ==========================================================================
//
// GORM 的 ORM 語法非常方便，但有時候需要直接寫 SQL：
//   - 複雜的子查詢（Subquery）
//   - 特定資料庫的函式（如 SQLite 的 datetime()）
//   - 效能調優（GORM 生成的 SQL 不夠優化）
//   - 報表查詢（GROUP BY、HAVING、複雜聚合）
//
// GORM 提供三種方式執行原始 SQL：
//   db.Raw(sql, args...).Scan(&result)  → 查詢並掃描到 struct
//   db.Exec(sql, args...)               → 執行（INSERT/UPDATE/DELETE）
//   db.Model(&Post{}).Where(clause)     → 混用 GORM 和原始片段

// demonstrateRawSQL 示範原始 SQL 的用法
func demonstrateRawSQL(db *gorm.DB) { // 示範原始 SQL
	fmt.Println("\n=== 3. 原始 SQL（Raw SQL）===\n") // 印出標題

	// ---- 查詢：db.Raw().Scan() ----
	fmt.Println("【db.Raw()：複雜查詢】")

	// 用 SQL 統計每個國家的作者數量和平均文章數
	type CountryStats struct { // 自訂結果 struct
		Country     string // 國家名稱
		AuthorCount int    // 作者數量
	}

	var stats []CountryStats // 儲存統計結果
	db.Raw(`
		SELECT country, COUNT(*) as author_count
		FROM authors
		GROUP BY country
		ORDER BY author_count DESC
	`).Scan(&stats) // 把查詢結果掃描到 struct slice

	fmt.Println("各國作者統計：")
	for _, s := range stats { // 遍歷統計結果
		fmt.Printf("  %s: %d 位作者\n", s.Country, s.AuthorCount)
	}

	// 帶參數的 Raw SQL（防止 SQL Injection，一定要用 ? 佔位符）
	fmt.Println("\n【帶參數的 Raw SQL（? 佔位符）】")

	type PostSummary struct { // 文章摘要 struct
		Title      string // 文章標題
		AuthorName string // 作者名稱
	}

	var summaries []PostSummary // 儲存結果
	db.Raw(`
		SELECT posts.title, authors.name as author_name
		FROM posts
		JOIN authors ON posts.author_id = authors.id
		WHERE authors.country = ?
		ORDER BY posts.created_at DESC
	`, "Taiwan").Scan(&summaries) // ? 會被 "Taiwan" 替換（安全的方式）

	fmt.Println("台灣作者的文章：")
	for _, s := range summaries { // 遍歷結果
		fmt.Printf("  「%s」by %s\n", s.Title, s.AuthorName)
	}

	// ---- 執行：db.Exec() ----
	fmt.Println("\n【db.Exec()：執行 UPDATE/INSERT/DELETE】")

	// 批量更新：把所有草稿文章的狀態改為 published
	result := db.Exec("UPDATE posts SET status = ? WHERE status = ?",
		"published", "draft") // 兩個 ? 分別對應兩個參數
	fmt.Printf("✅ 更新了 %d 篇草稿 → 已發布\n", result.RowsAffected) // 印出影響筆數

	// ---- 混合使用 GORM + Raw ----
	fmt.Println("\n【混合 GORM + Raw（Where 子句）】")

	var topPosts []Post            // 儲存結果
	db.Where("view_count > ?", 0). // GORM 的 Where
					Order("created_at DESC"). // GORM 的 Order
					Limit(3).                 // GORM 的 Limit
					Find(&topPosts)           // GORM 的 Find

	fmt.Printf("查詢到 %d 篇文章（混合寫法）\n", len(topPosts))

	// ---- 防 SQL Injection 示範 ----
	fmt.Println("\n【SQL Injection 防護】")
	fmt.Println("❌ 危險（千萬不要這樣）：")
	fmt.Println("  db.Raw(\"SELECT * FROM users WHERE name = '\" + input + \"'\")")
	fmt.Println("  → 如果 input = \"' OR '1'='1\"，會查出所有資料！")
	fmt.Println()
	fmt.Println("✅ 安全（用 ? 佔位符）：")
	fmt.Println("  db.Raw(\"SELECT * FROM users WHERE name = ?\", input)")
	fmt.Println("  → GORM 自動 escape，完全安全")
}

// ==========================================================================
// 4. 資料庫遷移管理（Migration）
// ==========================================================================
//
// 什麼是 Migration？
//   隨著專案發展，資料庫結構會不斷改變：
//   - v1.0：建立 users 表
//   - v1.1：users 表加 phone 欄位
//   - v1.2：建立 articles 表
//   - v1.3：articles 表加 view_count 欄位
//
//   如果用 AutoMigrate，每次都要手動追蹤「這個欄位加了嗎？」
//   Migration 工具讓你：
//   - 每個資料庫變更寫成一個 .sql 或 .go 檔案
//   - 記錄哪些變更已經執行過
//   - 可以「升版」也可以「降版」（rollback）
//
// 主流工具：
//   - golang-migrate：純 SQL 檔案，最簡單
//   - goose：Go 程式碼或 SQL，更靈活
//   - GORM 的 AutoMigrate：開發用，不建議在正式環境用
//
// AutoMigrate 的限制（不適合正式環境）：
//   - 只能「加欄位」，不能「刪欄位」或「重新命名欄位」
//   - 不記錄遷移歷史（不知道哪些改動已執行）
//   - 上線前可能需要維護視窗（migration 時間不可控）

// MigrationRecord 模擬遷移記錄表（golang-migrate 會自動管理這個表）
type MigrationRecord struct {
	Version   int64     `gorm:"primaryKey"` // 遷移版本號（通常用時間戳）
	AppliedAt time.Time // 套用時間
	Name      string    `gorm:"size:200"` // 遷移名稱
}

// demonstrateMigration 示範 Migration 概念
func demonstrateMigration(db *gorm.DB) { // 示範遷移管理
	fmt.Println("\n=== 4. 資料庫遷移管理（Migration）===\n") // 印出標題

	// 建立示範用的遷移記錄表
	db.AutoMigrate(&MigrationRecord{}) // 建立 migration_records 表

	// 模擬 golang-migrate 的遷移流程
	// 在真實專案中，這些會是 .sql 檔案：
	// migrations/
	//   000001_create_authors.up.sql     ← 升版
	//   000001_create_authors.down.sql   ← 降版（rollback）
	//   000002_create_posts.up.sql
	//   000002_create_posts.down.sql
	//   000003_add_view_count.up.sql
	//   000003_add_view_count.down.sql

	migrations := []MigrationRecord{ // 模擬遷移歷史
		{Version: 1, Name: "create_authors_table", AppliedAt: time.Now().Add(-72 * time.Hour)},
		{Version: 2, Name: "create_posts_table", AppliedAt: time.Now().Add(-48 * time.Hour)},
		{Version: 3, Name: "add_view_count_to_posts", AppliedAt: time.Now().Add(-24 * time.Hour)},
	}

	for _, m := range migrations { // 插入每個遷移記錄
		db.FirstOrCreate(&m, MigrationRecord{Version: m.Version}) // 已存在則不重複插入
	}

	// 查詢所有已套用的遷移
	var applied []MigrationRecord          // 儲存已套用的遷移
	db.Order("version asc").Find(&applied) // 按版本號排序

	fmt.Println("📋 已套用的資料庫遷移：")
	for _, m := range applied { // 遍歷遷移記錄
		fmt.Printf("  v%03d %-35s 套用於 %s\n",
			m.Version,
			m.Name,
			m.AppliedAt.Format("2006-01-02 15:04"))
	}

	// 說明 golang-migrate 的使用方式
	fmt.Println("\n💡 實際使用 golang-migrate：")
	fmt.Println()
	fmt.Println("  1. 安裝工具：")
	fmt.Println("     go install -tags 'sqlite' github.com/golang-migrate/migrate/v4/cmd/migrate@latest")
	fmt.Println()
	fmt.Println("  2. 建立遷移檔案：")
	fmt.Println("     migrate create -ext sql -dir migrations -seq create_users_table")
	fmt.Println("     → 產生：migrations/000001_create_users_table.up.sql")
	fmt.Println("     →       migrations/000001_create_users_table.down.sql")
	fmt.Println()
	fmt.Println("  3. 執行遷移（升版）：")
	fmt.Println("     migrate -database 'sqlite://./blog.db' -path migrations up")
	fmt.Println()
	fmt.Println("  4. 降版（rollback）：")
	fmt.Println("     migrate -database 'sqlite://./blog.db' -path migrations down 1")
	fmt.Println()
	fmt.Println("  5. 在 Go 程式中使用：")
	fmt.Println("     import \"github.com/golang-migrate/migrate/v4\"")
	fmt.Println("     m, _ := migrate.New(\"file://migrations\", dbURL)")
	fmt.Println("     m.Up()  // 升版到最新")
}

// ==========================================================================
// 5. 交易（Transaction）進階
// ==========================================================================
//
// 第十四課學了基本交易
// 這裡學更進階的：SavePoint（中間存檔點）
//
// SavePoint 像是電玩的「中間存檔」：
//   交易開始 ──▶ 操作 A ──▶ SavePoint ──▶ 操作 B ──▶ 失敗！
//                                                      ↓
//                                              Rollback to SavePoint
//                                              （回到中間存檔點，A 保留，B 取消）
//
// 什麼時候用 SavePoint？
//   - 交易中有些操作失敗是「可以接受的」，不需要全部 rollback
//   - 複雜業務邏輯，需要分段確認

// demonstrateTransaction 示範進階交易用法
func demonstrateTransaction(db *gorm.DB) { // 示範交易進階
	fmt.Println("\n=== 5. 交易（Transaction）進階 ===\n") // 印出標題

	// ---- 基本交易（複習）----
	fmt.Println("【基本交易複習】")

	// 確保 authors[0] 和 authors[1] 存在（先查出來）
	var author1, author2 Author  // 儲存兩個作者
	db.First(&author1)           // 取第一個作者
	db.Offset(1).First(&author2) // 取第二個作者

	err := db.Transaction(func(tx *gorm.DB) error { // 開始交易
		// 在 tx 中的所有操作要嘛全成功，要嘛全失敗
		post1 := Post{Title: "交易測試文章 1", AuthorID: author1.ID, Status: "published"}
		if err := tx.Create(&post1).Error; err != nil { // 建立第一篇文章
			return err // 回傳錯誤 → 交易自動 rollback
		}

		post2 := Post{Title: "交易測試文章 2", AuthorID: author2.ID, Status: "draft"}
		if err := tx.Create(&post2).Error; err != nil { // 建立第二篇文章
			return err // 回傳錯誤 → 交易自動 rollback
		}

		return nil // 回傳 nil → 交易自動 commit
	})

	if err != nil { // 如果交易失敗
		fmt.Printf("❌ 交易失敗: %v\n", err)
	} else {
		fmt.Println("✅ 交易成功：兩篇文章都建立了")
	}

	// ---- SavePoint 示範 ----
	fmt.Println("\n【SavePoint：中間存檔點】")

	tx := db.Begin()     // 手動開始交易（不用 closure 的寫法）
	if tx.Error != nil { // 如果開始交易失敗
		log.Printf("開始交易失敗: %v", tx.Error) // 印出錯誤
		return                             // 提前返回
	}

	// 操作 A：建立文章（這個要保留）
	postA := Post{Title: "SavePoint 測試：A 的文章", AuthorID: author1.ID, Status: "published"}
	tx.Create(&postA) // 在交易中建立
	fmt.Printf("✅ 建立文章 A（ID: %d）\n", postA.ID)

	tx.SavePoint("after_post_a") // 建立 SavePoint，名稱為 "after_post_a"
	fmt.Println("📍 建立 SavePoint: after_post_a")

	// 操作 B：模擬失敗的操作（這個要取消）
	// 嘗試建立一個 email 重複的作者（會失敗）
	badAuthor := Author{Name: "Duplicate", Email: "alice@example.com"} // email 已存在！
	if err := tx.Create(&badAuthor).Error; err != nil {                // 建立失敗
		fmt.Printf("❌ 操作 B 失敗: %v\n", err)
		tx.RollbackTo("after_post_a") // 回到 SavePoint（A 保留）
		fmt.Println("⏪ Rollback 到 SavePoint after_post_a（文章 A 保留）")
	}

	// Commit：A 已保留，B 已取消
	tx.Commit() // 提交交易（只有 A 被儲存）
	fmt.Println("✅ Commit 完成（文章 A 已儲存，壞資料已取消）")

	// ---- 交易中的錯誤處理最佳實踐 ----
	fmt.Println("\n【交易最佳實踐】")
	fmt.Println("  ✅ 用 closure 寫法（db.Transaction）：GORM 自動 commit/rollback")
	fmt.Println("  ✅ 手動寫法（db.Begin/Commit/Rollback）：需要 defer tx.Rollback() 確保 rollback")
	fmt.Println("  ✅ 永遠在 tx 上操作，不要混用 db 和 tx")
	fmt.Println("  ❌ 不要在交易中執行耗時操作（會長時間鎖表）")
}

// ==========================================================================
// 6. 查詢優化技巧
// ==========================================================================
//
// 除了索引和 Preload，還有幾個重要的優化技巧：
//   Select  → 只取需要的欄位，不要 SELECT *
//   Count   → 先算數量再決定要不要查詳細資料
//   Scopes  → 把常用的查詢條件包成函式，提高重用性
//   分頁    → 永遠要分頁，不要一次查全部資料
//   Batch   → 大量操作用 FindInBatches 分批處理

// demonstrateOptimizations 示範查詢優化技巧
func demonstrateOptimizations(db *gorm.DB) { // 示範查詢優化
	fmt.Println("\n=== 6. 查詢優化技巧 ===\n") // 印出標題

	// ---- Select：只取需要的欄位 ----
	fmt.Println("【Select：只取需要的欄位（不要 SELECT *）】")

	type PostTitle struct { // 只需要標題和作者 ID 的輕量 struct
		ID       uint
		Title    string
		AuthorID uint
	}

	var titles []PostTitle                            // 儲存查詢結果
	db.Model(&Post{}).Select("id, title, author_id"). // 只選三個欄位
								Limit(3).     // 最多 3 筆
								Scan(&titles) // 掃描到自訂 struct

	fmt.Printf("✅ 只查詢 3 個欄位（比 SELECT * 更省頻寬和記憶體）：%d 筆\n", len(titles))
	for _, t := range titles {
		fmt.Printf("  ID:%d Title:%s\n", t.ID, t.Title)
	}

	// ---- Count：先算數量 ----
	fmt.Println("\n【Count：先確認有資料再查詳細】")

	var count int64                                                  // 儲存數量
	db.Model(&Post{}).Where("status = ?", "published").Count(&count) // 只計數，不查詳細
	fmt.Printf("✅ 已發布文章：%d 篇\n", count)

	if count > 0 { // 確認有資料才繼續查詢
		fmt.Printf("  → 有資料，繼續查詢詳細資訊\n")
	}

	// ---- 分頁（Pagination）----
	fmt.Println("\n【分頁查詢（生產環境必備）】")

	page := 1                       // 第幾頁（從 1 開始）
	pageSize := 2                   // 每頁幾筆
	offset := (page - 1) * pageSize // 計算跳過幾筆

	var pagedPosts []Post // 儲存分頁結果
	db.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&pagedPosts)

	fmt.Printf("✅ 第 %d 頁（每頁 %d 筆）：查到 %d 筆\n", page, pageSize, len(pagedPosts))
	for _, p := range pagedPosts {
		fmt.Printf("  - %s\n", p.Title)
	}

	// ---- Scopes：可重用的查詢條件 ----
	fmt.Println("\n【Scopes：把查詢條件封裝成函式】")

	// 定義可重用的 Scope 函式
	published := func(db *gorm.DB) *gorm.DB { // published Scope
		return db.Where("status = ?", "published") // 只查已發布的
	}
	recentFirst := func(db *gorm.DB) *gorm.DB { // recentFirst Scope
		return db.Order("created_at desc") // 最新的排前面
	}

	var scopedPosts []Post                                        // 儲存結果
	db.Scopes(published, recentFirst).Limit(3).Find(&scopedPosts) // 組合多個 Scope

	fmt.Printf("✅ 用 Scopes 查詢：%d 篇已發布文章（最新優先）\n", len(scopedPosts))

	// ---- FindInBatches：大量資料分批處理 ----
	fmt.Println("\n【FindInBatches：大量資料分批處理】")

	// 想像有 100 萬筆文章要處理
	// 如果一次全查，記憶體會爆炸：var allPosts []Post; db.Find(&allPosts) ← 不要這樣！
	// 用 FindInBatches 分批處理：每批 2 筆（示範用，實際通常 100-1000）

	totalProcessed := 0 // 記錄處理的總筆數
	result := db.FindInBatches(&[]Post{}, 2, func(tx *gorm.DB, batch int) error {
		var batchPosts []Post             // 這批的文章
		tx.Find(&batchPosts)              // 取這批資料
		totalProcessed += len(batchPosts) // 累計處理數量
		fmt.Printf("  處理第 %d 批：%d 筆\n", batch, len(batchPosts))
		return nil // 繼續下一批
	})

	if result.Error != nil { // 如果處理失敗
		log.Printf("FindInBatches 失敗: %v", result.Error)
	}
	fmt.Printf("✅ FindInBatches 完成，共處理 %d 筆\n", totalProcessed)

	// ---- EXPLAIN（查詢計畫）----
	fmt.Println("\n【EXPLAIN：分析查詢效能】")
	fmt.Println("  在 SQLite 執行：")
	fmt.Println("  EXPLAIN QUERY PLAN SELECT * FROM posts WHERE author_id = 1")
	fmt.Println()
	fmt.Println("  在 MySQL/PostgreSQL 執行：")
	fmt.Println("  EXPLAIN SELECT * FROM posts WHERE author_id = 1")
	fmt.Println("  EXPLAIN ANALYZE SELECT * FROM posts WHERE author_id = 1  ← 實際執行並分析")
	fmt.Println()
	fmt.Println("  看 EXPLAIN 結果的重點：")
	fmt.Println("  ✅ Using index  → 有用到索引，很好")
	fmt.Println("  ❌ Using filesort → 全表掃描後排序，需要加索引")
	fmt.Println("  ❌ ALL type     → 全表掃描（最慢）")

	// 用 db.Raw 執行 EXPLAIN（SQLite 版本）
	var explainResult []map[string]any // 儲存 EXPLAIN 結果
	db.Raw("EXPLAIN QUERY PLAN SELECT * FROM posts WHERE status = 'published'").
		Scan(&explainResult) // 掃描結果

	fmt.Println("\n  EXPLAIN QUERY PLAN 結果（status 欄位有索引）：")
	for _, row := range explainResult { // 遍歷結果
		fmt.Printf("  %v\n", row)
	}
}

// ==========================================================================
// 主程式
// ==========================================================================

// initDB 初始化資料庫連線和自動遷移
func initDB() *gorm.DB { // 初始化資料庫
	// 設定 GORM 日誌（Silent 模式，避免示範時輸出太多 SQL）
	// 在需要看 SQL 的地方，用 db.Debug() 開啟
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 靜默模式（不自動印 SQL）
	}

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), gormConfig)
	// file::memory: → 使用記憶體資料庫，程式結束後資料消失（示範用）
	// ?cache=shared → 允許多個連線共用同一個記憶體資料庫
	if err != nil { // 如果連線失敗
		log.Fatalf("資料庫連線失敗: %v", err) // 印出錯誤並結束程式
	}

	// AutoMigrate：自動建立/更新表結構
	err = db.AutoMigrate(
		&Author{},          // 建立 authors 表
		&Tag{},             // 建立 tags 表
		&Post{},            // 建立 posts 表
		&MigrationRecord{}, // 建立 migration_records 表
	)
	if err != nil { // 如果遷移失敗
		log.Fatalf("AutoMigrate 失敗: %v", err) // 印出錯誤並結束程式
	}

	return db // 回傳資料庫連線
}

// setupTags 建立示範用的標籤資料
func setupTags(db *gorm.DB) { // 建立標籤
	tags := []Tag{ // 準備標籤資料
		{Name: "Go"},
		{Name: "Backend"},
		{Name: "Database"},
	}
	for _, tag := range tags { // 逐一建立標籤
		db.FirstOrCreate(&tag, Tag{Name: tag.Name}) // 不存在才建立
	}
}

// 讓 errors 套件被使用（避免 import 錯誤，實際用在交易的錯誤處理）
var _ = errors.New // 確保 errors 套件被引用

func main() { // 程式進入點
	fmt.Println("==========================================") // 分隔線
	fmt.Println(" 第二十二課：資料庫進階（Database Advanced）")            // 標題
	fmt.Println("==========================================") // 分隔線

	db := initDB() // 初始化資料庫
	setupTags(db)  // 建立測試標籤

	demonstrateIndexes(db)       // 示範 1：索引
	demonstrateNPlusOne(db)      // 示範 2：N+1 問題
	demonstrateRawSQL(db)        // 示範 3：原始 SQL
	demonstrateMigration(db)     // 示範 4：資料庫遷移
	demonstrateTransaction(db)   // 示範 5：交易進階
	demonstrateOptimizations(db) // 示範 6：查詢優化

	fmt.Println("\n==========================================") // 分隔線
	fmt.Println(" 教學完成！")                                       // 結尾
	fmt.Println("==========================================")   // 分隔線
}
