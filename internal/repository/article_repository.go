package repository

// 教學對應：第 10 課（Clean Architecture Repository 層）、第 14 課（GORM CRUD）、第 28 課（資料庫進階）

import (
	"blog-api/internal/domain"

	"gorm.io/gorm"
)

// articleRepository 實作 domain.ArticleRepository 介面
type articleRepository struct {
	db *gorm.DB
}

// NewArticleRepository 建立文章 Repository 實例
func NewArticleRepository(db *gorm.DB) domain.ArticleRepository {
	return &articleRepository{db: db}
}

// Create 在交易中建立新文章
func (r *articleRepository) Create(article *domain.Article) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(article).Error; err != nil {
			return err
		}
		// 在同一個交易中載入關聯的使用者資料
		return tx.Preload("User").First(article, article.ID).Error
	})
}

// FindByID 根據 ID 查詢文章，同時預載入作者與留言資料
func (r *articleRepository) FindByID(id uint) (*domain.Article, error) {
	var article domain.Article
	err := r.db.Preload("User").Preload("Comments").Preload("Comments.User").First(&article, id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// FindAll 根據查詢條件回傳文章列表與總筆數
func (r *articleRepository) FindAll(query domain.ArticleQuery) ([]domain.Article, int64, error) {
	var articles []domain.Article
	var total int64

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	db := r.db.Model(&domain.Article{})

	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		db = db.Where("title LIKE ? OR content LIKE ?", searchPattern, searchPattern)
	}

	if query.UserID > 0 {
		db = db.Where("user_id = ?", query.UserID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&articles).Error

	return articles, total, err
}

// Update 更新文章
func (r *articleRepository) Update(article *domain.Article) error {
	return r.db.Save(article).Error
}

// Delete 刪除文章（軟刪除，GORM 偵測到 DeletedAt 欄位會自動處理）
func (r *articleRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Article{}, id).Error
}
