// 第十三課：GORM 資料庫操作
// GORM 是 Go 最受歡迎的 ORM 框架，讓你用 Go 程式碼操作資料庫
//
// 執行前安裝：
//   go get gorm.io/gorm
//   go get gorm.io/driver/sqlite
//
// 執行方式：go run main.go
package main

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ========================================
// 1. 定義 Model（對應資料庫的表）
// ========================================

// User 使用者（對應 users 資料表）
type User struct {
	ID       uint      `gorm:"primaryKey"`
	Username string    `gorm:"uniqueIndex;size:50;not null"`
	Email    string    `gorm:"uniqueIndex;size:100;not null"`
	Age      int       `gorm:"default:0"`
	Articles []Article `gorm:"foreignKey:UserID"` // 一對多：一個使用者有多篇文章
}

// Article 文章（對應 articles 資料表）
type Article struct {
	ID      uint   `gorm:"primaryKey"`
	Title   string `gorm:"size:200;not null"`
	Content string `gorm:"type:text"`
	UserID  uint   `gorm:"index;not null"` // 外鍵：指向 User
	User    User   `gorm:"foreignKey:UserID"`
}

func main() {
	// ========================================
	// 2. 連接資料庫
	// ========================================
	fmt.Println("=== 連接資料庫 ===")

	// 使用 SQLite（會在當前目錄建立 tutorial.db 檔案）
	db, err := gorm.Open(sqlite.Open("tutorial.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("連接失敗:", err)
	}
	fmt.Println("連接成功！")

	// ========================================
	// 3. 自動遷移（AutoMigrate）
	// ========================================
	fmt.Println("\n=== 自動遷移 ===")

	// AutoMigrate 會根據結構體自動建立或更新資料表
	// 只會新增欄位，不會刪除或修改現有欄位
	db.AutoMigrate(&User{}, &Article{})
	fmt.Println("資料表建立/更新完成")

	// 清除舊資料（方便重複執行教學）
	db.Where("1 = 1").Delete(&Article{})
	db.Where("1 = 1").Delete(&User{})

	// ========================================
	// 4. Create（建立）
	// ========================================
	fmt.Println("\n=== CREATE ===")

	// 建立單筆
	alice := User{Username: "alice", Email: "alice@test.com", Age: 25}
	result := db.Create(&alice) // 傳入指標，GORM 會回寫自動產生的 ID
	fmt.Printf("建立 alice: ID=%d, 影響 %d 列\n", alice.ID, result.RowsAffected)

	bob := User{Username: "bob", Email: "bob@test.com", Age: 30}
	db.Create(&bob)
	fmt.Printf("建立 bob: ID=%d\n", bob.ID)

	carol := User{Username: "carol", Email: "carol@test.com", Age: 28}
	db.Create(&carol)

	// 建立文章
	articles := []Article{
		{Title: "Go 入門教學", Content: "Go 是一門簡潔的語言", UserID: alice.ID},
		{Title: "GORM 使用指南", Content: "GORM 讓資料庫操作更簡單", UserID: alice.ID},
		{Title: "Python vs Go", Content: "兩者的比較分析", UserID: bob.ID},
	}
	db.Create(&articles)
	fmt.Printf("建立了 %d 篇文章\n", len(articles))

	// ========================================
	// 5. Read（讀取）
	// ========================================
	fmt.Println("\n=== READ ===")

	// 5.1 根據主鍵查詢
	var user User
	db.First(&user, alice.ID) // SELECT * FROM users WHERE id = ?
	fmt.Printf("First: %s (%s)\n", user.Username, user.Email)

	// 5.2 條件查詢
	var foundUser User
	db.Where("email = ?", "bob@test.com").First(&foundUser)
	fmt.Printf("Where: %s\n", foundUser.Username)

	// 5.3 查詢所有
	var allUsers []User
	db.Find(&allUsers) // SELECT * FROM users
	fmt.Printf("Find: 共 %d 位使用者\n", len(allUsers))

	// 5.4 查詢特定欄位
	var names []string
	db.Model(&User{}).Pluck("username", &names)
	fmt.Println("所有使用者名稱:", names)

	// ========================================
	// 6. 進階查詢
	// ========================================
	fmt.Println("\n=== 進階查詢 ===")

	// 6.1 排序
	var sortedUsers []User
	db.Order("age DESC").Find(&sortedUsers)
	fmt.Print("依年齡排序: ")
	for _, u := range sortedUsers {
		fmt.Printf("%s(%d) ", u.Username, u.Age)
	}
	fmt.Println()

	// 6.2 分頁（Limit + Offset）
	var pageUsers []User
	db.Limit(2).Offset(0).Find(&pageUsers) // 第 1 頁，每頁 2 筆
	fmt.Printf("分頁（前 2 筆）: %d 筆\n", len(pageUsers))

	// 6.3 計數
	var count int64
	db.Model(&User{}).Where("age > ?", 25).Count(&count)
	fmt.Printf("年齡 > 25 的使用者: %d 位\n", count)

	// 6.4 LIKE 搜尋
	var searchResults []Article
	db.Where("title LIKE ?", "%Go%").Find(&searchResults)
	fmt.Printf("標題含 'Go' 的文章: %d 篇\n", len(searchResults))

	// ========================================
	// 7. Preload（預載入關聯資料）
	// ========================================
	fmt.Println("\n=== Preload ===")

	// 不用 Preload：user.Articles 會是空的
	var userNoPreload User
	db.First(&userNoPreload, alice.ID)
	fmt.Printf("沒有 Preload: %s 有 %d 篇文章\n",
		userNoPreload.Username, len(userNoPreload.Articles))

	// 使用 Preload：自動載入關聯的文章
	var userWithArticles User
	db.Preload("Articles").First(&userWithArticles, alice.ID)
	fmt.Printf("有 Preload: %s 有 %d 篇文章\n",
		userWithArticles.Username, len(userWithArticles.Articles))
	for _, article := range userWithArticles.Articles {
		fmt.Printf("  - %s\n", article.Title)
	}

	// ========================================
	// 8. Update（更新）
	// ========================================
	fmt.Println("\n=== UPDATE ===")

	// 8.1 Save：更新所有欄位
	user.Age = 26
	db.Save(&user) // UPDATE users SET ... WHERE id = ?
	fmt.Println("Save 後年齡:", user.Age)

	// 8.2 Update：更新單一欄位
	db.Model(&User{}).Where("id = ?", bob.ID).Update("age", 31)

	// 8.3 Updates：更新多個欄位
	db.Model(&User{}).Where("id = ?", carol.ID).Updates(map[string]interface{}{
		"age":   29,
		"email": "carol.new@test.com",
	})

	// 驗證更新結果
	var updatedCarol User
	db.First(&updatedCarol, carol.ID)
	fmt.Printf("Carol 更新後: age=%d, email=%s\n", updatedCarol.Age, updatedCarol.Email)

	// ========================================
	// 9. Delete（刪除）
	// ========================================
	fmt.Println("\n=== DELETE ===")

	// 刪除特定記錄
	db.Delete(&Article{}, articles[2].ID) // 刪除第三篇文章
	fmt.Println("已刪除文章:", articles[2].Title)

	// 驗證
	var remainingArticles []Article
	db.Find(&remainingArticles)
	fmt.Printf("剩餘文章數: %d\n", len(remainingArticles))

	// ========================================
	// 10. 鏈式查詢（Query Chain）
	// ========================================
	fmt.Println("\n=== 鏈式查詢 ===")

	// 可以像 builder 一樣串接多個條件
	var results []Article
	db.Where("user_id = ?", alice.ID).
		Order("title ASC").
		Limit(10).
		Find(&results)

	fmt.Printf("Alice 的文章（排序後）:\n")
	for _, a := range results {
		fmt.Printf("  - %s\n", a.Title)
	}

	fmt.Println("\n教學完成！資料庫檔案: tutorial.db")
}
