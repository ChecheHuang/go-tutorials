// ==========================================================================
// 第十四課：GORM 資料庫操作（完整版）
// ==========================================================================
//
// 什麼是 ORM？
//   ORM（Object-Relational Mapping，物件關聯對映）是程式語言和資料庫之間的「翻譯員」
//   它讓你用 Go 的 struct 來操作資料庫，不需要手寫 SQL 語句
//   Go struct ←→ 資料庫表格（Table）
//   struct 欄位 ←→ 表格欄位（Column）
//
// 什麼是 GORM？
//   GORM 是 Go 語言最受歡迎的 ORM 框架
//   官網：https://gorm.io
//   它幫你把 Go 程式碼自動轉換成 SQL，大幅減少重複程式碼
//
// 為什麼用 github.com/glebarez/sqlite？
//   這是純 Go 寫的 SQLite 驅動程式，不需要安裝 C 編譯器（CGO）
//   相比 gorm.io/driver/sqlite 需要 gcc，這個套件開箱即用
//
// gorm.io/gorm 是什麼？
//   這是 GORM 的核心框架，提供所有 ORM 功能
//   包括：自動遷移、CRUD 操作、關聯查詢、交易處理等
//
// 執行方式：go run ./tutorials/14-gorm-database
// ==========================================================================

package main // 每個可執行程式都必須是 main 套件

import (         // 匯入需要的套件
	"errors"   // 標準錯誤處理套件
	"fmt"      // 格式化輸出套件
	"log"      // 日誌套件
	"time"     // 時間套件

	"github.com/glebarez/sqlite" // 純 Go 的 SQLite 驅動程式，不需要 CGO
	"gorm.io/gorm"               // GORM 核心框架，提供所有 ORM 功能
)

// ==========================================================================
// 1. 定義 Model（模型）—— 每個 struct 對應資料庫中的一張表
// ==========================================================================

// User 使用者模型，對應資料庫中的 users 表
type User struct {
	ID        uint           `gorm:"primaryKey"`                   // 主鍵，自動遞增的唯一識別碼
	Username  string         `gorm:"uniqueIndex;size:50;not null"` // 使用者名稱，唯一索引、最大 50 字元、不可為空
	Email     string         `gorm:"uniqueIndex;size:100;not null"`// 電子郵件，唯一索引、最大 100 字元、不可為空
	Age       int            `gorm:"default:0"`                    // 年齡，預設值為 0
	Balance   float64        `gorm:"default:0"`                    // 餘額，用於示範交易（Transaction）
	Articles  []Article      `gorm:"foreignKey:UserID"`            // 一對多關聯：一個使用者有多篇文章
	CreatedAt time.Time      // GORM 自動管理：記錄建立時間
	UpdatedAt time.Time      // GORM 自動管理：記錄更新時間
	DeletedAt gorm.DeletedAt `gorm:"index"`                        // 軟刪除：刪除時不真正刪除，只標記時間
}

// Article 文章模型，對應資料庫中的 articles 表
type Article struct {
	ID        uint           `gorm:"primaryKey"`        // 主鍵
	Title     string         `gorm:"size:200;not null"` // 標題，最大 200 字元
	Content   string         `gorm:"type:text"`         // 內容，使用 text 型別以支援長文
	UserID    uint           `gorm:"index;not null"`    // 外鍵，指向 User 表的 ID
	User      User           `gorm:"foreignKey:UserID"` // 屬於關聯：每篇文章屬於一個使用者
	CreatedAt time.Time      // 建立時間
	UpdatedAt time.Time      // 更新時間
	DeletedAt gorm.DeletedAt `gorm:"index"` // 軟刪除欄位
}

// ==========================================================================
// 軟刪除（Soft Delete）說明
// ==========================================================================
//
// 什麼是軟刪除？
//   一般刪除（硬刪除）：資料從資料庫中永久消失
//   軟刪除：資料還在，但被標記為「已刪除」，查詢時自動忽略
//
// 為什麼要軟刪除？
//   - 資料可以恢復（像資源回收桶）
//   - 保留歷史紀錄，方便稽核
//   - 防止誤刪導致資料遺失
//
// 怎麼用？
//   在 struct 中加入 `gorm.DeletedAt` 欄位
//   GORM 會自動：
//   - 刪除時：填入當前時間（而不是真的刪除）
//   - 查詢時：自動加上 WHERE deleted_at IS NULL（只顯示未刪除的）
//   - 需要查已刪除的資料：db.Unscoped().Find(...)

func main() { // 程式進入點
	// ======================================================================
	// 2. 連接資料庫
	// ======================================================================
	fmt.Println("=== 連接資料庫 ===") // 輸出段落標題

	// 使用 SQLite，會在當前目錄建立 tutorial.db 檔案
	// sqlite.Open("tutorial.db") 建立一個 SQLite 連線設定
	// &gorm.Config{} 是 GORM 的設定，這裡使用預設值
	db, err := gorm.Open(sqlite.Open("tutorial.db"), &gorm.Config{}) // 開啟資料庫連線
	if err != nil {                                                    // 如果連線失敗
		log.Fatal("連接失敗:", err) // 印出錯誤並結束程式
	}
	fmt.Println("連接成功！") // 連線成功的提示

	// ======================================================================
	// 3. 自動遷移（AutoMigrate）
	// ======================================================================
	//
	// 什麼是遷移（Migration）？
	//   遷移就是讓資料庫的表格結構與你的 Go struct 保持同步
	//   當你新增或修改 struct 的欄位時，資料庫也需要跟著改變
	//
	// AutoMigrate 會做什麼？
	//   ✅ 建立不存在的表格
	//   ✅ 新增缺少的欄位
	//   ✅ 建立索引
	//   ❌ 不會刪除欄位（怕你誤刪資料）
	//   ❌ 不會修改欄位型別（怕破壞現有資料）
	//   ❌ 不會刪除索引
	//
	// 如果需要更精細的遷移控制，可以使用：
	//   - golang-migrate/migrate（專業遷移工具）
	//   - 手動寫 SQL 遷移腳本
	//
	// 在正式產品中的建議：
	//   開發環境：用 AutoMigrate，方便快速
	//   正式環境：用遷移工具，更安全可控
	// ======================================================================
	fmt.Println("\n=== 自動遷移（AutoMigrate）===") // 輸出段落標題

	// 傳入所有 Model 的指標，GORM 會自動建立或更新對應的資料表
	err = db.AutoMigrate(&User{}, &Article{}) // 執行自動遷移
	if err != nil {                            // 如果遷移失敗
		log.Fatal("遷移失敗:", err) // 印出錯誤並結束程式
	}
	fmt.Println("資料表建立/更新完成") // 遷移成功的提示

	// 清除舊資料，方便重複執行教學範例
	db.Unscoped().Where("1 = 1").Delete(&Article{}) // Unscoped() 真正刪除（包含軟刪除的資料）
	db.Unscoped().Where("1 = 1").Delete(&User{})    // 清除所有使用者資料

	// ======================================================================
	// 4. Create（建立）—— 把 Go struct 存入資料庫
	// ======================================================================
	fmt.Println("\n=== CREATE（建立）===") // 輸出段落標題

	// 建立單筆使用者
	alice := User{Username: "alice", Email: "alice@test.com", Age: 25, Balance: 1000} // 建立 alice 的資料
	result := db.Create(&alice)                                                        // 傳入指標，GORM 會回寫自動產生的 ID
	fmt.Printf("建立 alice: ID=%d, 影響 %d 列\n", alice.ID, result.RowsAffected)      // 顯示結果

	bob := User{Username: "bob", Email: "bob@test.com", Age: 30, Balance: 500} // 建立 bob 的資料
	db.Create(&bob)                                                             // 存入資料庫
	fmt.Printf("建立 bob: ID=%d\n", bob.ID)                                    // 顯示 bob 的 ID

	carol := User{Username: "carol", Email: "carol@test.com", Age: 28, Balance: 800} // 建立 carol 的資料
	db.Create(&carol)                                                                  // 存入資料庫

	// 批量建立文章（一次建立多筆）
	articles := []Article{ // 建立文章切片
		{Title: "Go 入門教學", Content: "Go 是一門簡潔的語言", UserID: alice.ID},       // alice 的文章
		{Title: "GORM 使用指南", Content: "GORM 讓資料庫操作更簡單", UserID: alice.ID},  // alice 的另一篇
		{Title: "Python vs Go", Content: "兩者的比較分析", UserID: bob.ID},              // bob 的文章
	}
	db.Create(&articles)                               // 批量存入資料庫
	fmt.Printf("建立了 %d 篇文章\n", len(articles))    // 顯示建立的數量

	// ======================================================================
	// 5. Read（讀取）—— 從資料庫查詢資料
	// ======================================================================
	fmt.Println("\n=== READ（讀取）===") // 輸出段落標題

	// 5.1 根據主鍵查詢（First）
	var user User                        // 宣告一個空的 User 變數來接收結果
	db.First(&user, alice.ID)            // 相當於 SELECT * FROM users WHERE id = ? LIMIT 1
	fmt.Printf("First: %s (%s)\n", user.Username, user.Email) // 印出查詢結果

	// 5.2 條件查詢（Where）
	var foundUser User                                         // 宣告變數接收結果
	db.Where("email = ?", "bob@test.com").First(&foundUser)    // 用 email 查詢，? 是防 SQL 注入的佔位符
	fmt.Printf("Where: %s\n", foundUser.Username)              // 印出找到的使用者

	// 5.3 查詢所有（Find）
	var allUsers []User                                       // 宣告切片接收多筆結果
	db.Find(&allUsers)                                         // 相當於 SELECT * FROM users
	fmt.Printf("Find: 共 %d 位使用者\n", len(allUsers))       // 印出總數

	// 5.4 查詢特定欄位（Pluck）
	var names []string                                         // 宣告字串切片接收使用者名稱
	db.Model(&User{}).Pluck("username", &names)                // 只取 username 欄位
	fmt.Println("所有使用者名稱:", names)                       // 印出所有名稱

	// 5.5 檢查記錄是否存在
	var count int64                                             // 宣告計數變數
	result = db.Where("username = ?", "dave").First(&User{})    // 查詢不存在的使用者
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {        // 判斷是否為「找不到記錄」錯誤
		fmt.Println("dave 不存在（ErrRecordNotFound）")          // 正確處理找不到的情況
	}

	// ======================================================================
	// 6. 進階查詢
	// ======================================================================
	fmt.Println("\n=== 進階查詢 ===") // 輸出段落標題

	// 6.1 排序（Order）
	var sortedUsers []User                    // 宣告切片接收排序結果
	db.Order("age DESC").Find(&sortedUsers)   // 依年齡降序排列
	fmt.Print("依年齡排序: ")                  // 印出提示
	for _, u := range sortedUsers {            // 走訪每個使用者
		fmt.Printf("%s(%d) ", u.Username, u.Age) // 印出名稱和年齡
	}
	fmt.Println() // 換行

	// 6.2 分頁（Limit + Offset）
	var pageUsers []User                                // 宣告切片接收分頁結果
	db.Limit(2).Offset(0).Find(&pageUsers)              // 第 1 頁，每頁 2 筆（Offset=0 表示從第一筆開始）
	fmt.Printf("分頁（前 2 筆）: %d 筆\n", len(pageUsers)) // 印出分頁結果數量

	// 6.3 計數（Count）
	db.Model(&User{}).Where("age > ?", 25).Count(&count) // 計算年齡大於 25 的使用者數量
	fmt.Printf("年齡 > 25 的使用者: %d 位\n", count)      // 印出計數結果

	// 6.4 LIKE 模糊搜尋
	var searchResults []Article                                 // 宣告切片接收搜尋結果
	db.Where("title LIKE ?", "%Go%").Find(&searchResults)       // 搜尋標題包含 "Go" 的文章
	fmt.Printf("標題含 'Go' 的文章: %d 篇\n", len(searchResults)) // 印出搜尋結果數量

	// ======================================================================
	// 7. Preload（預載入關聯資料）—— 解決 N+1 問題
	// ======================================================================
	//
	// 什麼是 N+1 問題？
	//   假設你有 10 個使用者，每個使用者有文章
	//   不用 Preload：1 次查使用者 + 10 次查文章 = 11 次查詢（效率差）
	//   用 Preload：1 次查使用者 + 1 次查所有文章 = 2 次查詢（效率好）
	//
	// Preload 就是「預先載入」關聯資料，避免重複查詢
	// ======================================================================
	fmt.Println("\n=== Preload（預載入）===") // 輸出段落標題

	// 不用 Preload：關聯資料是空的
	var userNoPreload User                                   // 宣告變數
	db.First(&userNoPreload, alice.ID)                       // 只查使用者，不查文章
	fmt.Printf("沒有 Preload: %s 有 %d 篇文章\n",            // 文章數量會是 0
		userNoPreload.Username, len(userNoPreload.Articles))

	// 使用 Preload：自動載入關聯的文章
	var userWithArticles User                                       // 宣告變數
	db.Preload("Articles").First(&userWithArticles, alice.ID)       // 同時查使用者和文章
	fmt.Printf("有 Preload: %s 有 %d 篇文章\n",                    // 文章數量會正確顯示
		userWithArticles.Username, len(userWithArticles.Articles))
	for _, article := range userWithArticles.Articles { // 走訪每篇文章
		fmt.Printf("  - %s\n", article.Title) // 印出文章標題
	}

	// ======================================================================
	// 8. Update（更新）—— 修改資料庫中的資料
	// ======================================================================
	fmt.Println("\n=== UPDATE（更新）===") // 輸出段落標題

	// 8.1 Save：更新所有欄位（會更新整筆記錄）
	user.Age = 26                                     // 修改年齡
	db.Save(&user)                                    // 相當於 UPDATE users SET ... WHERE id = ?
	fmt.Println("Save 後年齡:", user.Age)              // 印出更新後的年齡

	// 8.2 Update：只更新單一欄位（效率更好）
	db.Model(&User{}).Where("id = ?", bob.ID).Update("age", 31) // 只更新 bob 的年齡

	// 8.3 Updates：更新多個欄位（用 map）
	db.Model(&User{}).Where("id = ?", carol.ID).Updates(map[string]interface{}{ // 用 map 指定多個欄位
		"age":   29,                 // 更新年齡
		"email": "carol.new@test.com", // 更新信箱
	})

	// 驗證更新結果
	var updatedCarol User                                                             // 宣告變數
	db.First(&updatedCarol, carol.ID)                                                 // 重新查詢
	fmt.Printf("Carol 更新後: age=%d, email=%s\n", updatedCarol.Age, updatedCarol.Email) // 印出結果

	// ======================================================================
	// 9. Delete（刪除）
	// ======================================================================
	fmt.Println("\n=== DELETE（刪除）===") // 輸出段落標題

	// 9.1 軟刪除（因為 Article 有 DeletedAt 欄位）
	db.Delete(&Article{}, articles[2].ID)                   // 軟刪除第三篇文章（只標記 deleted_at）
	fmt.Println("已軟刪除文章:", articles[2].Title)           // 提示已刪除

	// 驗證：一般查詢看不到軟刪除的資料
	var remainingArticles []Article                          // 宣告切片
	db.Find(&remainingArticles)                              // 查詢所有文章（自動排除已軟刪除的）
	fmt.Printf("剩餘文章數（排除軟刪除）: %d\n", len(remainingArticles)) // 印出數量

	// 9.2 查看包含軟刪除的所有資料
	var allArticles []Article                                // 宣告切片
	db.Unscoped().Find(&allArticles)                         // Unscoped() 忽略軟刪除條件
	fmt.Printf("所有文章數（包含軟刪除）: %d\n", len(allArticles)) // 印出包含軟刪除的數量

	// ======================================================================
	// 10. Transaction（交易）—— 確保多個操作要嘛全部成功，要嘛全部失敗
	// ======================================================================
	//
	// 什麼是交易（Transaction）？
	//   想像銀行轉帳：A 轉 100 元給 B
	//   步驟 1：A 的帳戶扣 100 元
	//   步驟 2：B 的帳戶加 100 元
	//
	//   如果步驟 1 成功但步驟 2 失敗怎麼辦？
	//   A 的錢扣了，B 卻沒收到 → 錢憑空消失！
	//
	//   交易的作用：把多個步驟包成「一個操作」
	//   全部成功 → 提交（Commit）：所有變更生效
	//   任一失敗 → 回滾（Rollback）：所有變更取消，恢復原狀
	//
	// GORM 的 db.Transaction() 方法：
	//   - 傳入一個函式，函式內的所有操作都在同一個交易中
	//   - 函式回傳 nil → 自動 Commit（提交）
	//   - 函式回傳 error → 自動 Rollback（回滾）
	// ======================================================================
	fmt.Println("\n=== Transaction（交易）===") // 輸出段落標題

	// 先查看轉帳前的餘額
	var aliceBefore, bobBefore User                                           // 宣告兩個變數
	db.First(&aliceBefore, alice.ID)                                          // 查詢 alice 的資料
	db.First(&bobBefore, bob.ID)                                              // 查詢 bob 的資料
	fmt.Printf("轉帳前: Alice=%.0f, Bob=%.0f\n", aliceBefore.Balance, bobBefore.Balance) // 顯示轉帳前餘額

	// 10.1 成功的交易：Alice 轉 200 元給 Bob
	transferAmount := 200.0 // 轉帳金額

	err = db.Transaction(func(tx *gorm.DB) error { // 開始交易，tx 是交易專用的資料庫連線
		// 步驟 1：扣除 Alice 的餘額
		if err := tx.Model(&User{}).Where("id = ?", alice.ID). // 找到 alice
			Update("balance", gorm.Expr("balance - ?", transferAmount)).Error; err != nil { // 扣除金額
			return err // 回傳錯誤 → 自動回滾
		}

		// 步驟 2：增加 Bob 的餘額
		if err := tx.Model(&User{}).Where("id = ?", bob.ID). // 找到 bob
			Update("balance", gorm.Expr("balance + ?", transferAmount)).Error; err != nil { // 增加金額
			return err // 回傳錯誤 → 自動回滾
		}

		return nil // 回傳 nil → 自動提交，兩個步驟都生效
	})

	if err != nil { // 檢查交易是否成功
		fmt.Println("交易失敗:", err) // 印出錯誤
	} else { // 交易成功
		fmt.Println("轉帳成功！") // 印出成功訊息
	}

	// 查看轉帳後的餘額
	var aliceAfter, bobAfter User                                          // 宣告兩個變數
	db.First(&aliceAfter, alice.ID)                                        // 重新查詢 alice
	db.First(&bobAfter, bob.ID)                                            // 重新查詢 bob
	fmt.Printf("轉帳後: Alice=%.0f, Bob=%.0f\n", aliceAfter.Balance, bobAfter.Balance) // 顯示轉帳後餘額

	// 10.2 失敗的交易：餘額不足時自動回滾
	fmt.Println("\n--- 測試交易回滾 ---") // 輸出小標題

	err = db.Transaction(func(tx *gorm.DB) error { // 開始另一個交易
		// 步驟 1：扣除 Alice 的餘額（大額轉帳）
		if err := tx.Model(&User{}).Where("id = ?", alice.ID). // 找到 alice
			Update("balance", gorm.Expr("balance - ?", 99999)).Error; err != nil { // 扣除超大金額
			return err // 有錯誤就回滾
		}

		// 步驟 2：檢查餘額是否足夠
		var checkUser User                                 // 宣告變數來檢查
		tx.First(&checkUser, alice.ID)                     // 查詢最新餘額
		if checkUser.Balance < 0 {                         // 如果餘額變成負數
			return fmt.Errorf("餘額不足！目前餘額: %.0f", checkUser.Balance) // 回傳錯誤 → 自動回滾
		}

		return nil // 這行不會執行到，因為上面已經 return error 了
	})

	if err != nil { // 檢查交易結果
		fmt.Println("交易被回滾:", err) // 交易失敗，所有變更都被取消
	}

	// 驗證回滾後餘額沒有變化
	var aliceCheck User                                              // 宣告變數
	db.First(&aliceCheck, alice.ID)                                  // 重新查詢
	fmt.Printf("回滾後 Alice 餘額: %.0f（沒有變化）\n", aliceCheck.Balance) // 餘額應該跟轉帳後一樣

	// ======================================================================
	// 11. 鏈式查詢（Query Chain）
	// ======================================================================
	fmt.Println("\n=== 鏈式查詢 ===") // 輸出段落標題

	// 可以像建造者模式一樣串接多個條件
	var chainResults []Article                  // 宣告切片接收結果
	db.Where("user_id = ?", alice.ID).          // 條件：alice 的文章
		Order("title ASC").                     // 排序：按標題升序
		Limit(10).                              // 限制：最多 10 筆
		Find(&chainResults)                     // 執行查詢

	fmt.Printf("Alice 的文章（排序後）:\n") // 印出標題
	for _, a := range chainResults {        // 走訪每篇文章
		fmt.Printf("  - %s\n", a.Title)    // 印出文章標題
	}

	// ======================================================================
	// 教學完成
	// ======================================================================
	fmt.Println("\n=== 教學完成 ===")                     // 輸出結尾標題
	fmt.Println("資料庫檔案: tutorial.db")                // 提示資料庫檔案位置
	fmt.Println("你可以用 DB Browser for SQLite 開啟查看") // 推薦的 GUI 工具
}
