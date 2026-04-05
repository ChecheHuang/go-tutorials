// ==========================================================================
// 第十五課：PostgreSQL + Schema Design + Index Optimization
// ==========================================================================
//
// 為什麼不能只用 SQLite？
//   - SQLite 是檔案型資料庫，不支援並發寫入
//   - 沒有使用者權限管理
//   - 沒有網路連線（只能本機存取）
//   - 生產環境 99% 用 PostgreSQL 或 MySQL
//
// 本課用純 Go 模擬 PostgreSQL 的核心概念（不需要安裝 PostgreSQL）
// 學習者之後只需要換 driver 就能連真正的 PostgreSQL
//
// 執行方式：
//   go run ./tutorials/15-postgresql/
// ==========================================================================

package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// ==========================================================================
// 1. Schema Design 三大正規化
// ==========================================================================

func demoSchemaDesign() {
	fmt.Println("📐 Demo 1: Schema Design（資料表設計）")
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println()

	// 反正規化（Bad）
	fmt.Println("  ❌ 反正規化（資料重複）：")
	fmt.Println("  ┌──────────────────────────────────────────────────────┐")
	fmt.Println("  │ orders 表                                            │")
	fmt.Println("  │ id │ user_name │ user_email      │ product │ price  │")
	fmt.Println("  │ 1  │ Alice     │ alice@mail.com  │ Go 書   │ 500   │")
	fmt.Println("  │ 2  │ Alice     │ alice@mail.com  │ K8s 書  │ 600   │")
	fmt.Println("  │         ↑ Alice 的資料重複了兩次！改 email 要改兩行   │")
	fmt.Println("  └──────────────────────────────────────────────────────┘")

	fmt.Println()
	fmt.Println("  ✅ 正規化（拆表 + 外鍵關聯）：")
	fmt.Println("  ┌────────────────────┐     ┌────────────────────────┐")
	fmt.Println("  │ users              │     │ orders                  │")
	fmt.Println("  │ id │ name │ email  │     │ id │ user_id │ product │")
	fmt.Println("  │ 1  │ Alice│ a@m.c  │◄────│ 1  │ 1       │ Go 書   │")
	fmt.Println("  └────────────────────┘     │ 2  │ 1       │ K8s 書  │")
	fmt.Println("    改 email 只需改一行       └────────────────────────┘")

	fmt.Println()
	fmt.Println("  三大正規化：")
	fmt.Println("  1NF：每個欄位只存一個值（不要 tags='go,docker,k8s'）")
	fmt.Println("  2NF：非主鍵欄位完全依賴主鍵（不要部分依賴）")
	fmt.Println("  3NF：非主鍵欄位不互相依賴（user_name 不該出現在 orders 表）")
}

// ==========================================================================
// 2. Index Optimization（索引最佳化）
// ==========================================================================

// SimpleTable 模擬資料表（用來展示有無索引的查詢差異）
type SimpleTable struct {
	rows  []map[string]string
	index map[string][]int // column_value → row indices
}

func NewSimpleTable() *SimpleTable {
	return &SimpleTable{
		rows:  make([]map[string]string, 0),
		index: make(map[string][]int),
	}
}

func (t *SimpleTable) Insert(row map[string]string) {
	idx := len(t.rows)
	t.rows = append(t.rows, row)

	// 自動更新索引（模擬 CREATE INDEX）
	for col, val := range row {
		key := col + ":" + val
		t.index[key] = append(t.index[key], idx)
	}
}

// FullScan 全表掃描（沒有索引）
func (t *SimpleTable) FullScan(column, value string) ([]map[string]string, int) {
	var results []map[string]string
	scanned := 0
	for _, row := range t.rows {
		scanned++
		if row[column] == value {
			results = append(results, row)
		}
	}
	return results, scanned
}

// IndexScan 索引掃描（有索引）
func (t *SimpleTable) IndexScan(column, value string) ([]map[string]string, int) {
	key := column + ":" + value
	indices, exists := t.index[key]
	if !exists {
		return nil, 1 // 查索引 1 次就知道不存在
	}

	results := make([]map[string]string, len(indices))
	for i, idx := range indices {
		results[i] = t.rows[idx]
	}
	return results, len(indices) // 只掃描匹配的行
}

func demoIndexOptimization() {
	fmt.Println("\n🔍 Demo 2: Index Optimization（索引最佳化）")
	fmt.Println("─────────────────────────────────────────────")

	table := NewSimpleTable()

	// 插入 10000 筆資料
	emails := []string{"alice", "bob", "carol", "dave", "eve"}
	for i := 0; i < 10000; i++ {
		table.Insert(map[string]string{
			"id":    fmt.Sprintf("%d", i+1),
			"email": fmt.Sprintf("%s@example.com", emails[rand.Intn(len(emails))]),
			"name":  fmt.Sprintf("User-%d", i+1),
		})
	}

	target := "alice@example.com"

	// 全表掃描
	start := time.Now()
	results1, scanned1 := table.FullScan("email", target)
	fullScanTime := time.Since(start)

	// 索引掃描
	start = time.Now()
	results2, scanned2 := table.IndexScan("email", target)
	indexScanTime := time.Since(start)

	fmt.Printf("\n  查詢：WHERE email = '%s'（10000 筆資料）\n\n", target)
	fmt.Printf("  %-15s │ 掃描行數 │ 找到 │ 耗時\n", "方式")
	fmt.Printf("  ─────────────┼─────────┼──────┼──────────\n")
	fmt.Printf("  %-15s │ %7d │ %4d │ %v\n", "全表掃描 ❌", scanned1, len(results1), fullScanTime)
	fmt.Printf("  %-15s │ %7d │ %4d │ %v\n", "索引掃描 ✅", scanned2, len(results2), indexScanTime)

	fmt.Println("\n  索引的代價：")
	fmt.Println("  ✅ 加速 SELECT（讀取快）")
	fmt.Println("  ❌ 減慢 INSERT/UPDATE（每次寫入要更新索引）")
	fmt.Println("  ❌ 佔用額外儲存空間")
}

// ==========================================================================
// 3. PostgreSQL vs SQLite 差異
// ==========================================================================

func demoPGvsSQLite() {
	fmt.Println("\n📊 Demo 3: PostgreSQL vs SQLite")
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("  ┌─────────────────┬──────────────────┬──────────────────┐")
	fmt.Println("  │ 特性            │ SQLite           │ PostgreSQL       │")
	fmt.Println("  ├─────────────────┼──────────────────┼──────────────────┤")
	fmt.Println("  │ 儲存方式        │ 單一檔案          │ 獨立服務（進程）  │")
	fmt.Println("  │ 並發寫入        │ ❌ 不支援         │ ✅ MVCC          │")
	fmt.Println("  │ 網路連線        │ ❌ 只能本機        │ ✅ TCP 連線       │")
	fmt.Println("  │ 使用者權限      │ ❌ 無             │ ✅ RBAC          │")
	fmt.Println("  │ JSON 支援       │ ⚠️ 基本           │ ✅ JSONB 索引    │")
	fmt.Println("  │ 全文搜尋        │ ⚠️ FTS5          │ ✅ tsvector      │")
	fmt.Println("  │ 適合場景        │ 開發/測試/嵌入式   │ 生產環境         │")
	fmt.Println("  │ 安裝難度        │ 零安裝            │ 需要安裝或 Docker │")
	fmt.Println("  └─────────────────┴──────────────────┴──────────────────┘")
}

// ==========================================================================
// 4. 常用索引類型
// ==========================================================================

func demoIndexTypes() {
	fmt.Println("\n📚 Demo 4: 常用索引類型")
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("  ┌────────────────┬──────────────────────┬────────────────────┐")
	fmt.Println("  │ 索引類型       │ SQL 語法              │ 適用場景           │")
	fmt.Println("  ├────────────────┼──────────────────────┼────────────────────┤")
	fmt.Println("  │ B-Tree（預設） │ CREATE INDEX idx     │ =, <, >, BETWEEN   │")
	fmt.Println("  │                │  ON users(email)     │ 大部分場景          │")
	fmt.Println("  ├────────────────┼──────────────────────┼────────────────────┤")
	fmt.Println("  │ Unique Index   │ CREATE UNIQUE INDEX  │ email 不能重複      │")
	fmt.Println("  │                │  ON users(email)     │                    │")
	fmt.Println("  ├────────────────┼──────────────────────┼────────────────────┤")
	fmt.Println("  │ Composite      │ CREATE INDEX idx     │ WHERE a=1 AND b=2  │")
	fmt.Println("  │ （複合索引）    │  ON orders(user, dt) │ 注意欄位順序！     │")
	fmt.Println("  ├────────────────┼──────────────────────┼────────────────────┤")
	fmt.Println("  │ Partial        │ CREATE INDEX idx     │ 只索引部分資料      │")
	fmt.Println("  │ （部分索引）    │  ON orders(status)   │ 例如只索引          │")
	fmt.Println("  │                │  WHERE status='pending'│ 未完成的訂單      │")
	fmt.Println("  ├────────────────┼──────────────────────┼────────────────────┤")
	fmt.Println("  │ GIN            │ CREATE INDEX idx     │ JSONB、全文搜尋     │")
	fmt.Println("  │                │  ON posts USING GIN  │ 陣列包含查詢        │")
	fmt.Println("  │                │  (tags)              │                    │")
	fmt.Println("  └────────────────┴──────────────────────┴────────────────────┘")

	fmt.Println()
	fmt.Println("  複合索引的欄位順序很重要（最左前綴原則）：")
	fmt.Println("  INDEX(a, b, c) 可以加速：")
	fmt.Println("    ✅ WHERE a = 1")
	fmt.Println("    ✅ WHERE a = 1 AND b = 2")
	fmt.Println("    ✅ WHERE a = 1 AND b = 2 AND c = 3")
	fmt.Println("    ❌ WHERE b = 2（跳過了 a，索引無效）")
	fmt.Println("    ❌ WHERE c = 3（跳過了 a 和 b）")
}

// ==========================================================================
// 5. GORM 切換 PostgreSQL 範例
// ==========================================================================

func demoGORMSwitch() {
	fmt.Println("\n🔧 Demo 5: GORM 切換 PostgreSQL（只需改 2 行）")
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("  目前（SQLite）：")
	fmt.Println("  ┌──────────────────────────────────────────────────┐")
	fmt.Println("  │ import \"github.com/glebarez/sqlite\"              │")
	fmt.Println("  │ db, _ := gorm.Open(sqlite.Open(\"blog.db\"))      │")
	fmt.Println("  └──────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  切換 PostgreSQL（只改 import 和 Open）：")
	fmt.Println("  ┌──────────────────────────────────────────────────┐")
	fmt.Println("  │ import \"gorm.io/driver/postgres\"                 │")
	fmt.Println("  │ dsn := \"host=localhost user=postgres\"            │")
	fmt.Println("  │     + \" password=xxx dbname=blog port=5432\"     │")
	fmt.Println("  │ db, _ := gorm.Open(postgres.Open(dsn))          │")
	fmt.Println("  └──────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  Clean Architecture 的好處：")
	fmt.Println("  Repository 層以上的程式碼完全不用改！")
	fmt.Println("  只需要改 main.go 的資料庫初始化。")
}

// ==========================================================================
// 6. EXPLAIN 分析查詢
// ==========================================================================

func demoExplain() {
	fmt.Println("\n🔬 Demo 6: EXPLAIN 分析查詢效能")
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("  查看 SQL 查詢的執行計畫：")
	fmt.Println()
	fmt.Println("  EXPLAIN ANALYZE SELECT * FROM articles WHERE user_id = 1;")
	fmt.Println()
	fmt.Println("  ┌─────────────────────────────────────────────────────────┐")
	fmt.Println("  │ Seq Scan on articles  (cost=0..25 rows=5)              │ ← ❌ 全表掃描")
	fmt.Println("  │   Filter: (user_id = 1)                                │")
	fmt.Println("  │   Rows Removed by Filter: 995                          │ ← 掃了 1000 行只要 5 行")
	fmt.Println("  │   Planning Time: 0.1 ms                                │")
	fmt.Println("  │   Execution Time: 2.5 ms                               │")
	fmt.Println("  └─────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  加上索引後：CREATE INDEX idx_articles_user ON articles(user_id);")
	fmt.Println()
	fmt.Println("  ┌─────────────────────────────────────────────────────────┐")
	fmt.Println("  │ Index Scan using idx_articles_user  (cost=0..8 rows=5) │ ← ✅ 索引掃描")
	fmt.Println("  │   Index Cond: (user_id = 1)                            │")
	fmt.Println("  │   Planning Time: 0.1 ms                                │")
	fmt.Println("  │   Execution Time: 0.05 ms                              │ ← 快 50 倍！")
	fmt.Println("  └─────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  什麼時候該加索引？")
	fmt.Println("  ✅ WHERE 條件常用的欄位（user_id, email, status）")
	fmt.Println("  ✅ JOIN 的外鍵欄位")
	fmt.Println("  ✅ ORDER BY 的欄位")
	fmt.Println("  ❌ 很少查詢的欄位（加了浪費空間）")
	fmt.Println("  ❌ 值很少變化的欄位（如 boolean，選擇性太低）")
}

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║   第 15 課：PostgreSQL + Schema Design + Index          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")

	demoSchemaDesign()
	demoIndexOptimization()
	demoPGvsSQLite()
	demoIndexTypes()
	demoGORMSwitch()
	demoExplain()

	fmt.Println()
	_ = strings.Builder{} // 避免 unused import
}
