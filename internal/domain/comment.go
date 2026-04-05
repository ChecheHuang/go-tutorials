package domain

import "time"

// Comment 定義留言實體
type Comment struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	ArticleID uint      `json:"article_id" gorm:"index;not null"` // 所屬文章 ID（外鍵）
	UserID    uint      `json:"user_id" gorm:"index;not null"`    // 留言者 ID（外鍵）
	User      User      `json:"user" gorm:"foreignKey:UserID"`    // 關聯的使用者資料
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateCommentRequest 定義建立留言的請求結構
type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000" example:"寫得很好，感謝分享！"`
}

// UpdateCommentRequest 定義更新留言的請求結構
type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000" example:"更新後的留言內容"`
}

// CommentRepository 定義留言的資料存取介面
type CommentRepository interface {
	Create(comment *Comment) error
	FindByID(id uint) (*Comment, error)
	FindByArticleID(articleID uint) ([]Comment, error) // 取得某篇文章的所有留言
	Update(comment *Comment) error
	Delete(id uint) error
}
