// ==========================================================================
// 第十七課：進階測試補充
// ==========================================================================
//
// 這個檔案補充三個重要的進階測試技巧：
//
//   1. testify/assert — 讓測試斷言更簡潔易讀
//   2. 整合測試（Integration Test）— 用真實 SQLite 測試 Repository
//   3. 基準測試（Benchmark）— 測量程式碼的效能
//
// 執行方式：
//   go test -v ./tutorials/17-testing/                    # 全部測試
//   go test -v -run TestIntegration ./tutorials/17-testing/ # 只跑整合測試
//   go test -bench=. ./tutorials/17-testing/              # 只跑基準測試
//   go test -race ./tutorials/17-testing/                 # 競爭條件偵測
//   go test -cover ./tutorials/17-testing/                # 顯示覆蓋率
// ==========================================================================

package main // 和 main.go 同 package

import (
	"fmt"     // 格式化輸出
	"testing" // Go 標準測試框架

	"github.com/glebarez/sqlite"          // 純 Go SQLite 驅動
	"github.com/stretchr/testify/assert"  // testify：斷言工具
	"github.com/stretchr/testify/require" // testify：致命斷言（失敗立即停止）
	"gorm.io/gorm"                        // ORM 框架
	"gorm.io/gorm/logger"                 // GORM 日誌設定
)

// ==========================================================================
// 7. testify/assert — 讓斷言更簡潔
// ==========================================================================
//
// Go 標準庫的測試寫法：
//   if result != 5 {
//       t.Errorf("Add(2, 3) = %d，期望 5", result)
//   }
//
// testify/assert 的寫法：
//   assert.Equal(t, 5, result)
//   assert.NoError(t, err)
//   assert.Nil(t, user)
//
// 優點：
//   - 更簡潔（少寫很多 if）
//   - 錯誤訊息更友善（自動顯示「期望 5，但得到 3」）
//   - 提供豐富的斷言函式（EqualError、Contains、Len 等）
//
// assert vs require 的差別：
//   assert.Equal → 失敗後繼續執行（記錄錯誤但不停止）
//   require.Equal → 失敗後立即停止（通常用在後面的測試依賴這個結果時）

// TestWithTestify 示範 testify/assert 的用法
func TestWithTestify(t *testing.T) { // 示範 testify
	t.Run("assert.Equal（基本相等判斷）", func(t *testing.T) {
		result := Add(2, 3)                    // 呼叫被測試的函式
		assert.Equal(t, 5, result)             // 期望 result 等於 5（比 if 簡潔很多）
		assert.Equal(t, 5, result, "Add(2, 3) 應該等於 5") // 可以加自訂訊息
	})

	t.Run("assert.NoError（測試無錯誤）", func(t *testing.T) {
		result, err := Divide(10, 2)  // 除法（不會有錯誤）
		assert.NoError(t, err)         // 期望 err 是 nil
		assert.Equal(t, 5.0, result)  // 期望結果是 5.0
	})

	t.Run("assert.Error（測試有錯誤）", func(t *testing.T) {
		_, err := Divide(10, 0)                  // 除以零（應該有錯誤）
		assert.Error(t, err)                      // 期望 err 不是 nil
		assert.EqualError(t, err, "除數不能為零") // 期望特定的錯誤訊息
	})

	t.Run("assert.Contains（字串包含）", func(t *testing.T) {
		user := User{Name: "Alice", Age: 25}
		str := fmt.Sprintf("%v", user)           // 轉成字串
		assert.Contains(t, str, "Alice")          // 字串應該包含 "Alice"
	})

	t.Run("assert.Len（長度判斷）", func(t *testing.T) {
		items := []int{1, 2, 3}
		assert.Len(t, items, 3) // 切片長度應該是 3
	})

	t.Run("assert.True / assert.False", func(t *testing.T) {
		user := User{Name: "Alice", Age: 25, Email: "alice@example.com"}
		err := user.Validate()
		assert.NoError(t, err)   // 應該沒有驗證錯誤

		emptyUser := User{}
		err = emptyUser.Validate()
		assert.Error(t, err) // 空使用者應該驗證失敗
	})

	t.Run("require.NoError（致命斷言：失敗立即停止）", func(t *testing.T) {
		// require 和 assert 用法一樣，但失敗時立即停止這個子測試
		// 常用在：後面的測試依賴這個結果
		result, err := Divide(10, 2)
		require.NoError(t, err)   // 如果有錯誤，立即停止（下面的 assert 不會執行）
		assert.Equal(t, 5.0, result) // 只有上面成功才執行到這裡
	})
}

// ==========================================================================
// 8. 整合測試（Integration Test）
// ==========================================================================
//
// 什麼是整合測試？
//   單元測試：只測試一個函式，用 Mock 隔離外部依賴
//   整合測試：測試多個元件一起工作，用真實的外部依賴（真實資料庫）
//
// 單元測試 vs 整合測試：
//   ┌──────────────┬─────────────────┬────────────────────────┐
//   │              │ 單元測試         │ 整合測試                │
//   ├──────────────┼─────────────────┼────────────────────────┤
//   │ 速度          │ 極快（毫秒）     │ 較慢（需啟動 DB）       │
//   │ 外部依賴      │ Mock（假的）     │ 真實（SQLite/Redis）    │
//   │ 測試什麼      │ 邏輯是否正確     │ 元件能否一起正常運作     │
//   │ 何時執行      │ 每次提交前       │ CI/CD 或每天一次        │
//   └──────────────┴─────────────────┴────────────────────────┘
//
// 最佳實踐：
//   用 TestMain 或 testing.Short() 讓整合測試可以被略過：
//   go test -short ./...   → 只跑單元測試（略過整合測試）
//   go test ./...          → 跑所有測試（包含整合測試）

// PostModel 整合測試用的資料模型
type PostModel struct {
	ID      uint   `gorm:"primaryKey"` // 主鍵
	Title   string `gorm:"not null"`   // 標題
	Content string                     // 內容
	UserID  uint   `gorm:"not null"`   // 作者 ID
}

// PostRepository 文章 Repository 介面
type PostRepository interface {
	Create(title, content string, userID uint) (*PostModel, error) // 建立文章
	FindByID(id uint) (*PostModel, error)                         // 根據 ID 查詢
	FindByUser(userID uint) ([]*PostModel, error)                 // 查詢使用者的文章
	Delete(id uint) error                                         // 刪除文章
}

// GormPostRepository 用 GORM 實作 PostRepository
type GormPostRepository struct {
	db *gorm.DB // GORM 資料庫連線
}

// Create 建立文章
func (r *GormPostRepository) Create(title, content string, userID uint) (*PostModel, error) {
	post := PostModel{Title: title, Content: content, UserID: userID}
	result := r.db.Create(&post)
	return &post, result.Error
}

// FindByID 根據 ID 查詢文章
func (r *GormPostRepository) FindByID(id uint) (*PostModel, error) {
	var post PostModel
	result := r.db.First(&post, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &post, nil
}

// FindByUser 查詢使用者的所有文章
func (r *GormPostRepository) FindByUser(userID uint) ([]*PostModel, error) {
	var posts []*PostModel
	result := r.db.Where("user_id = ?", userID).Find(&posts)
	return posts, result.Error
}

// Delete 刪除文章
func (r *GormPostRepository) Delete(id uint) error {
	return r.db.Delete(&PostModel{}, id).Error
}

// setupTestDB 建立整合測試用的記憶體資料庫
// 每個測試都呼叫這個，確保每次測試都從乾淨的狀態開始（測試隔離）
func setupTestDB(t *testing.T) *gorm.DB { // 建立測試資料庫
	t.Helper() // 標記為輔助函式（錯誤訊息會顯示呼叫者的位置，不是這裡）

	db, err := gorm.Open(
		sqlite.Open("file::memory:?cache=private"), // 記憶體資料庫（cache=private → 每次都是新的）
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent), // 測試時靜默 SQL 輸出
		},
	)
	require.NoError(t, err, "建立測試資料庫失敗") // 如果失敗，立即停止測試

	err = db.AutoMigrate(&PostModel{}) // 建立表結構
	require.NoError(t, err, "資料庫遷移失敗")

	return db // 回傳測試用資料庫
}

// TestIntegrationPostRepository 整合測試：測試 GormPostRepository
func TestIntegrationPostRepository(t *testing.T) {
	// 如果執行 go test -short，略過整合測試（讓 CI 可以快速跑單元測試）
	if testing.Short() {
		t.Skip("略過整合測試（使用 -short flag）") // 跳過這個測試
	}

	// 建立測試資料庫（記憶體 SQLite，每次測試都是全新的）
	db := setupTestDB(t)
	repo := &GormPostRepository{db: db} // 建立 Repository（注入真實 DB）

	t.Run("建立文章", func(t *testing.T) {
		post, err := repo.Create("測試標題", "測試內容", 1)

		require.NoError(t, err)              // 不應該有錯誤
		assert.NotZero(t, post.ID)           // ID 應該不是 0（資料庫自動設定）
		assert.Equal(t, "測試標題", post.Title) // 標題應該正確
		assert.Equal(t, uint(1), post.UserID) // 作者 ID 應該正確
	})

	t.Run("根據 ID 查詢文章", func(t *testing.T) {
		// 先建立一篇文章
		created, err := repo.Create("查詢測試", "內容", 1)
		require.NoError(t, err)

		// 然後用 ID 查詢
		found, err := repo.FindByID(created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)       // ID 相同
		assert.Equal(t, "查詢測試", found.Title)     // 標題相同
	})

	t.Run("查詢不存在的文章", func(t *testing.T) {
		_, err := repo.FindByID(9999) // ID 9999 不存在
		assert.Error(t, err)          // 應該有錯誤（記錄不存在）
	})

	t.Run("查詢使用者的所有文章", func(t *testing.T) {
		// 建立兩篇屬於 userID=2 的文章
		repo.Create("文章 A", "內容 A", 2)
		repo.Create("文章 B", "內容 B", 2)
		// 建立一篇屬於 userID=3 的文章（不應該出現在結果中）
		repo.Create("文章 C", "內容 C", 3)

		posts, err := repo.FindByUser(2)   // 查詢 userID=2 的文章
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(posts), 2) // 應該至少有 2 篇（可能有更多）
		for _, p := range posts {
			assert.Equal(t, uint(2), p.UserID) // 每篇文章的 UserID 都應該是 2
		}
	})

	t.Run("刪除文章", func(t *testing.T) {
		// 先建立文章
		post, err := repo.Create("要刪除的文章", "內容", 1)
		require.NoError(t, err)

		// 刪除文章
		err = repo.Delete(post.ID)
		assert.NoError(t, err) // 刪除應該成功

		// 再查詢應該找不到
		_, err = repo.FindByID(post.ID)
		assert.Error(t, err) // 應該有錯誤（已被刪除）
	})
}

// ==========================================================================
// 9. 基準測試（Benchmark Test）
// ==========================================================================
//
// 基準測試用來測量程式碼的效能（每秒可以執行幾次）
//
// 執行方式：
//   go test -bench=. ./tutorials/17-testing/              # 跑所有 Benchmark
//   go test -bench=BenchmarkAdd ./tutorials/17-testing/   # 只跑特定 Benchmark
//   go test -bench=. -benchmem ./tutorials/17-testing/    # 顯示記憶體分配
//   go test -bench=. -count=5 ./tutorials/17-testing/     # 重複 5 次取平均值
//
// 輸出格式：
//   BenchmarkAdd-8   1000000000   0.3125 ns/op
//   ↑ 函式名稱     ↑ 執行次數  ↑ 每次操作的奈秒數
//   -8 表示用了 8 個 CPU 核心

// BenchmarkAdd 測試 Add 函式的效能
func BenchmarkAdd(b *testing.B) { // b *testing.B 是基準測試控制器
	for i := 0; i < b.N; i++ { // b.N 是 Go 自動決定的執行次數（確保結果穩定）
		Add(100, 200)          // 執行被測試的函式
	}
	// 輸出範例：BenchmarkAdd-8   1000000000   0.31 ns/op
	// 意思：執行了 10 億次，每次 0.31 奈秒（非常快！）
}

// BenchmarkDivide 測試 Divide 函式的效能（含錯誤處理）
func BenchmarkDivide(b *testing.B) {
	for i := 0; i < b.N; i++ { // 重複執行 b.N 次
		Divide(100.0, 3.0)    // 正常的除法
	}
}

// BenchmarkUserValidate 測試 User.Validate 的效能
func BenchmarkUserValidate(b *testing.B) {
	user := User{Name: "Alice", Age: 25, Email: "alice@example.com"} // 建立測試用使用者
	b.ResetTimer()                                                    // 重置計時器（不計入 user 建立的時間）

	for i := 0; i < b.N; i++ { // 重複執行
		user.Validate()         // 測試 Validate 的效能
	}
}

// ==========================================================================
// 10. 競爭條件偵測（Race Detector）
// ==========================================================================
//
// 什麼是競爭條件（Race Condition）？
//   兩個 goroutine 同時讀寫同一個變數，結果不可預測
//
//   例子：
//     counter := 0
//     go func() { counter++ }()  // goroutine 1
//     go func() { counter++ }()  // goroutine 2
//     // 最後 counter 可能是 1 而不是 2！（兩個同時讀到 0，都加 1，最後都寫 1）
//
// Go 的 Race Detector（競爭偵測器）可以自動找到這類問題：
//   go test -race ./...            # 執行測試時啟用競爭偵測
//   go run -race main.go           # 執行程式時啟用競爭偵測
//
// 注意：啟用 -race 後，程式會慢 5-10 倍（只在開發/CI 環境用）

// SafeCounter 執行緒安全的計數器（使用 channel 避免競爭條件）
type SafeCounter struct {
	ch chan int // 用 channel 傳遞操作（比 Mutex 更 Go-style）
}

// NewSafeCounter 建立 SafeCounter
func NewSafeCounter() *SafeCounter {
	c := &SafeCounter{ch: make(chan int)} // 建立 channel
	go func() {                          // 啟動一個 goroutine 管理 counter 狀態
		count := 0                       // 只有這個 goroutine 修改 count
		for delta := range c.ch {        // 從 channel 讀取操作
			count += delta               // 更新 count
		}
	}()
	return c
}

// Increment 增加計數
func (c *SafeCounter) Increment() {
	c.ch <- 1 // 透過 channel 傳遞操作（不直接修改 count）
}

// TestRaceConditionSafe 測試執行緒安全的計數器
// 執行：go test -race -run TestRaceConditionSafe ./tutorials/17-testing/
func TestRaceConditionSafe(t *testing.T) {
	// 這個測試本身不測功能，主要是讓你知道如何用 -race 偵測
	// 如果程式有 race condition，執行 go test -race 時會報告
	t.Log("💡 用 'go test -race ./tutorials/17-testing/' 偵測競爭條件")
	t.Log("   Race Detector 會在執行時動態偵測所有並發問題")
	t.Log("   建議在 CI/CD 中永遠加上 -race flag")
}
