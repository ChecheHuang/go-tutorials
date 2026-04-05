package usecase

import (
	"errors"

	"blog-api/internal/domain"
)

// ArticleUsecase 定義文章業務邏輯的介面
type ArticleUsecase interface {
	Create(userID uint, req domain.CreateArticleRequest) (*domain.Article, error)
	GetByID(id uint) (*domain.Article, error)
	GetAll(query domain.ArticleQuery) ([]domain.Article, int64, error)
	Update(id, userID uint, req domain.UpdateArticleRequest) (*domain.Article, error)
	Delete(id, userID uint) error
}

// articleUsecase 實作 ArticleUsecase 介面
type articleUsecase struct {
	articleRepo domain.ArticleRepository
}

// NewArticleUsecase 建立文章 Usecase 實例
func NewArticleUsecase(articleRepo domain.ArticleRepository) ArticleUsecase {
	return &articleUsecase{articleRepo: articleRepo}
}

// Create 建立新文章
// userID 從 JWT 中介層取得，確保文章歸屬正確的作者
func (u *articleUsecase) Create(userID uint, req domain.CreateArticleRequest) (*domain.Article, error) {
	article := &domain.Article{
		Title:   req.Title,
		Content: req.Content,
		UserID:  userID,
	}

	if err := u.articleRepo.Create(article); err != nil {
		return nil, errors.New("建立文章失敗")
	}

	// 重新查詢以載入關聯的使用者資料
	return u.articleRepo.FindByID(article.ID)
}

// GetByID 根據 ID 取得文章詳情（含作者與留言）
func (u *articleUsecase) GetByID(id uint) (*domain.Article, error) {
	article, err := u.articleRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("文章不存在")
	}
	return article, nil
}

// GetAll 取得文章列表（支援分頁與搜尋）
func (u *articleUsecase) GetAll(query domain.ArticleQuery) ([]domain.Article, int64, error) {
	return u.articleRepo.FindAll(query)
}

// Update 更新文章
// 只有文章作者本人可以更新
func (u *articleUsecase) Update(id, userID uint, req domain.UpdateArticleRequest) (*domain.Article, error) {
	// 先查詢文章是否存在
	article, err := u.articleRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("文章不存在")
	}

	// 檢查是否為文章作者
	if article.UserID != userID {
		return nil, errors.New("無權限修改此文章")
	}

	// 只更新有提供的欄位
	if req.Title != "" {
		article.Title = req.Title
	}
	if req.Content != "" {
		article.Content = req.Content
	}

	if err := u.articleRepo.Update(article); err != nil {
		return nil, errors.New("更新文章失敗")
	}

	return article, nil
}

// Delete 刪除文章
// 只有文章作者本人可以刪除
func (u *articleUsecase) Delete(id, userID uint) error {
	// 先查詢文章是否存在
	article, err := u.articleRepo.FindByID(id)
	if err != nil {
		return errors.New("文章不存在")
	}

	// 檢查是否為文章作者
	if article.UserID != userID {
		return errors.New("無權限刪除此文章")
	}

	return u.articleRepo.Delete(id)
}
