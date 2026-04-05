package domain

import "time"

// Article 定義文章實體
type Article struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"size:200;not null"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`       // 作者 ID（外鍵）
	User      User      `json:"user" gorm:"foreignKey:UserID"`       // 關聯的使用者資料
	Comments  []Comment `json:"comments,omitempty" gorm:"foreignKey:ArticleID"` // 文章的留言列表
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateArticleRequest 定義建立文章的請求結構
type CreateArticleRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=200" example:"我的第一篇文章"`
	Content string `json:"content" binding:"required,min=1" example:"這是文章的內容，支援很長的文字..."`
}

// UpdateArticleRequest 定義更新文章的請求結構
type UpdateArticleRequest struct {
	Title   string `json:"title" binding:"omitempty,min=1,max=200" example:"更新後的標題"`
	Content string `json:"content" binding:"omitempty,min=1" example:"更新後的內容"`
}

// ArticleQuery 定義文章查詢參數
type ArticleQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`          // 頁碼（預設 1）
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"` // 每頁筆數（預設 10，最大 100）
	Search   string `form:"search"`                                  // 搜尋關鍵字（搜尋標題與內容）
	UserID   uint   `form:"user_id"`                                 // 依作者 ID 篩選
}

// ArticleRepository 定義文章的資料存取介面
type ArticleRepository interface {
	Create(article *Article) error
	FindByID(id uint) (*Article, error)
	FindAll(query ArticleQuery) ([]Article, int64, error) // 回傳文章列表、總筆數、錯誤
	Update(article *Article) error
	Delete(id uint) error
}
