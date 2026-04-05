package repository

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

// Create 建立新文章
func (r *articleRepository) Create(article *domain.Article) error {
	return r.db.Create(article).Error
}

// FindByID 根據 ID 查詢文章，同時預載入作者與留言資料
func (r *articleRepository) FindByID(id uint) (*domain.Article, error) {
	var article domain.Article
	// Preload 會自動載入關聯的 User 和 Comments 資料
	err := r.db.Preload("User").Preload("Comments").Preload("Comments.User").First(&article, id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// FindAll 根據查詢條件回傳文章列表與總筆數
// 支援分頁、關鍵字搜尋與依作者篩選
func (r *articleRepository) FindAll(query domain.ArticleQuery) ([]domain.Article, int64, error) {
	var articles []domain.Article
	var total int64

	// 設定分頁預設值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	// 建立查詢 builder
	db := r.db.Model(&domain.Article{})

	// 關鍵字搜尋：搜尋標題與內容
	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		db = db.Where("title LIKE ? OR content LIKE ?", searchPattern, searchPattern)
	}

	// 依作者 ID 篩選
	if query.UserID > 0 {
		db = db.Where("user_id = ?", query.UserID)
	}

	// 先計算符合條件的總筆數（分頁前）
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 計算偏移量並執行分頁查詢
	offset := (query.Page - 1) * query.PageSize
	err := db.Preload("User").
		Order("created_at DESC"). // 依建立時間降序排列
		Offset(offset).
		Limit(query.PageSize).
		Find(&articles).Error

	return articles, total, err
}

// Update 更新文章
func (r *articleRepository) Update(article *domain.Article) error {
	return r.db.Save(article).Error
}

// Delete 刪除文章（硬刪除）
func (r *articleRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Article{}, id).Error
}
