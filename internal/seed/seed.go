// Package seed 提供資料庫初始資料填充功能
// 當資料庫為空時，自動建立測試用的使用者、文章與留言
package seed

import (
	"log/slog"

	"blog-api/internal/domain"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Run 檢查資料庫是否為空，若為空則填入初始資料
func Run(db *gorm.DB) {
	var count int64
	db.Model(&domain.User{}).Count(&count)
	if count > 0 {
		return
	}

	slog.Info("資料庫為空，開始填入初始資料")

	password, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("密碼加密失敗，跳過初始資料填入", "error", err)
		return
	}
	users := []domain.User{
		{Username: "alice", Email: "alice@example.com", Password: string(password)},
		{Username: "bob", Email: "bob@example.com", Password: string(password)},
		{Username: "carol", Email: "carol@example.com", Password: string(password)},
	}
	for i := range users {
		db.Create(&users[i])
	}
	slog.Info("建立使用者", "count", len(users))

	articles := []domain.Article{
		{Title: "Go 語言入門：為什麼選擇 Go？", Content: "Go 語言由 Google 開發，具有簡潔的語法、強大的並行處理能力，以及極快的編譯速度。本文將帶你了解 Go 的核心優勢與適用場景。", UserID: users[0].ID},
		{Title: "使用 Gin 框架建立 REST API", Content: "Gin 是 Go 語言最受歡迎的 Web 框架之一。本文將示範如何用 Gin 建立一個完整的 RESTful API，包含路由設定、中介層與錯誤處理。", UserID: users[0].ID},
		{Title: "GORM 資料庫操作完全指南", Content: "GORM 是 Go 語言的 ORM 框架，讓你用結構體操作資料庫。本文涵蓋 CRUD、關聯查詢、分頁、交易等常見操作。", UserID: users[0].ID},
		{Title: "JWT 認證機制詳解", Content: "JSON Web Token 是現代 API 最常用的認證方式。本文解釋 JWT 的結構、產生流程、驗證原理，以及安全注意事項。", UserID: users[1].ID},
		{Title: "Clean Architecture 實戰心得", Content: "在實際專案中導入 Clean Architecture 的經驗分享。探討分層的好處、依賴注入的實作方式，以及常見的踩坑經驗。", UserID: users[1].ID},
		{Title: "Docker 容器化部署入門", Content: "從零開始學習 Docker，包含 Dockerfile 撰寫、多階段建置、docker-compose 編排，以及生產環境的最佳實踐。", UserID: users[2].ID},
	}
	for i := range articles {
		db.Create(&articles[i])
	}
	slog.Info("建立文章", "count", len(articles))

	comments := []domain.Comment{
		{Content: "寫得很清楚，對初學者很有幫助！", ArticleID: articles[0].ID, UserID: users[1].ID},
		{Content: "請問 Go 和 Rust 相比有什麼優劣？", ArticleID: articles[0].ID, UserID: users[2].ID},
		{Content: "感謝分享，已經照著做出來了。", ArticleID: articles[1].ID, UserID: users[2].ID},
		{Content: "Gin 的效能真的很好，推薦！", ArticleID: articles[1].ID, UserID: users[1].ID},
		{Content: "GORM 的 Preload 功能太方便了。", ArticleID: articles[2].ID, UserID: users[1].ID},
		{Content: "JWT 的安全性要注意密鑰管理。", ArticleID: articles[3].ID, UserID: users[0].ID},
		{Content: "Clean Architecture 真的讓程式碼好維護很多。", ArticleID: articles[4].ID, UserID: users[2].ID},
		{Content: "Docker 多階段建置可以大幅縮小映像大小。", ArticleID: articles[5].ID, UserID: users[0].ID},
	}
	for i := range comments {
		db.Create(&comments[i])
	}
	slog.Info("建立留言", "count", len(comments))

	slog.Info("初始資料填入完成")
}
