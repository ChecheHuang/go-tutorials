package repository

import (
	"blog-api/internal/domain"

	"gorm.io/gorm"
)

// commentRepository 實作 domain.CommentRepository 介面
type commentRepository struct {
	db *gorm.DB
}

// NewCommentRepository 建立留言 Repository 實例
func NewCommentRepository(db *gorm.DB) domain.CommentRepository {
	return &commentRepository{db: db}
}

// Create 建立新留言
func (r *commentRepository) Create(comment *domain.Comment) error {
	return r.db.Create(comment).Error
}

// FindByID 根據 ID 查詢留言，同時預載入留言者資料
func (r *commentRepository) FindByID(id uint) (*domain.Comment, error) {
	var comment domain.Comment
	if err := r.db.Preload("User").First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

// FindByArticleID 取得某篇文章的所有留言
// 依建立時間升序排列，同時載入留言者資訊
func (r *commentRepository) FindByArticleID(articleID uint) ([]domain.Comment, error) {
	var comments []domain.Comment
	err := r.db.Preload("User").
		Where("article_id = ?", articleID).
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

// Update 更新留言
func (r *commentRepository) Update(comment *domain.Comment) error {
	return r.db.Save(comment).Error
}

// Delete 刪除留言（硬刪除）
func (r *commentRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Comment{}, id).Error
}
