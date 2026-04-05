// 第十一課：Gin 框架入門
// Gin 是 Go 最受歡迎的 Web 框架，提供路由、中介層、JSON 綁定等功能
//
// 執行前先安裝：go get github.com/gin-gonic/gin
// 執行方式：go run main.go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// User 使用者結構體
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// 模擬資料庫
var users = []User{
	{ID: 1, Name: "Alice", Age: 25},
	{ID: 2, Name: "Bob", Age: 30},
}
var nextID = 3

func main() {
	// ========================================
	// 1. 建立 Gin 引擎
	// ========================================

	// gin.Default() 包含 Logger 和 Recovery 中介層
	r := gin.Default()

	// ========================================
	// 2. 基本路由
	// ========================================

	// GET /ping → 健康檢查
	r.GET("/ping", func(c *gin.Context) {
		// c 是 gin.Context，包含請求和回應的所有操作
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
		// gin.H 是 map[string]any 的簡寫
	})

	// ========================================
	// 3. 路由群組（Route Groups）
	// ========================================

	// 所有 /api/v1 開頭的路由
	v1 := r.Group("/api/v1")
	{
		// GET /api/v1/users → 取得使用者列表
		v1.GET("/users", getUsers)

		// GET /api/v1/users/1 → 取得特定使用者
		v1.GET("/users/:id", getUserByID)

		// POST /api/v1/users → 建立使用者
		v1.POST("/users", createUser)

		// PUT /api/v1/users/1 → 更新使用者
		v1.PUT("/users/:id", updateUser)

		// DELETE /api/v1/users/1 → 刪除使用者
		v1.DELETE("/users/:id", deleteUser)
	}

	// ========================================
	// 4. 查詢參數範例
	// ========================================

	// GET /api/v1/search?keyword=alice&page=1
	v1.GET("/search", searchUsers)

	// ========================================
	// 啟動伺服器
	// ========================================
	r.Run(":9090") // 預設 :8080
}

// ========================================
// Handler 函式
// ========================================

// getUsers 取得所有使用者
func getUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data": users,
	})
}

// getUserByID 根據 ID 取得使用者
func getUserByID(c *gin.Context) {
	// c.Param("id") 取得路由參數 :id 的值
	idStr := c.Param("id")

	// 手動轉型（第 12 課會教更優雅的方式）
	var id int
	for _, user := range users {
		if idStr == string(rune('0'+user.ID)) || idStr == formatID(user.ID) {
			c.JSON(http.StatusOK, gin.H{"data": user})
			return
		}
	}

	_ = id
	c.JSON(http.StatusNotFound, gin.H{
		"message": "使用者不存在",
	})
}

// createUser 建立新使用者
func createUser(c *gin.Context) {
	var newUser User

	// ShouldBindJSON：從請求 body 解析 JSON 到結構體
	// 比標準庫的 json.Decoder 更簡潔
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "無效的 JSON: " + err.Error(),
		})
		return
	}

	newUser.ID = nextID
	nextID++
	users = append(users, newUser)

	// 回傳 201 Created
	c.JSON(http.StatusCreated, gin.H{
		"data": newUser,
	})
}

// updateUser 更新使用者
func updateUser(c *gin.Context) {
	idStr := c.Param("id")

	var updateData User
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	for i, user := range users {
		if formatID(user.ID) == idStr {
			if updateData.Name != "" {
				users[i].Name = updateData.Name
			}
			if updateData.Age != 0 {
				users[i].Age = updateData.Age
			}
			c.JSON(http.StatusOK, gin.H{"data": users[i]})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "使用者不存在"})
}

// deleteUser 刪除使用者
func deleteUser(c *gin.Context) {
	idStr := c.Param("id")

	for i, user := range users {
		if formatID(user.ID) == idStr {
			// 從切片中移除元素
			users = append(users[:i], users[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "已刪除"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "使用者不存在"})
}

// searchUsers 搜尋使用者（示範查詢參數）
func searchUsers(c *gin.Context) {
	// c.Query("key") 取得 URL 查詢參數
	keyword := c.Query("keyword")
	// c.DefaultQuery 提供預設值
	page := c.DefaultQuery("page", "1")

	c.JSON(http.StatusOK, gin.H{
		"keyword": keyword,
		"page":    page,
		"message": "搜尋功能示範",
	})
}

// formatID 將 int 轉為字串（簡易實作）
func formatID(id int) string {
	return string(rune('0' + id))
}
