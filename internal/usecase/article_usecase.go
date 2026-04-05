package usecase

import (
	"blog-api/internal/domain"
	"blog-api/pkg/apperror"
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
func (u *articleUsecase) Create(userID uint, req domain.CreateArticleRequest) (*domain.Article, error) {
	article := &domain.Article{
		Title:   req.Title,
		Content: req.Content,
		UserID:  userID,
	}

	if err := u.articleRepo.Create(article); err != nil {
		return nil, apperror.Wrap(apperror.ErrInternal, "建立文章失敗")
	}

	return u.articleRepo.FindByID(article.ID)
}

// GetByID 根據 ID 取得文章詳情（含作者與留言）
func (u *articleUsecase) GetByID(id uint) (*domain.Article, error) {
	article, err := u.articleRepo.FindByID(id)
	if err != nil {
		return nil, apperror.Wrap(apperror.ErrNotFound, "文章 ID=%d", id)
	}
	return article, nil
}

// GetAll 取得文章列表（支援分頁與搜尋）
func (u *articleUsecase) GetAll(query domain.ArticleQuery) ([]domain.Article, int64, error) {
	return u.articleRepo.FindAll(query)
}

// Update 更新文章（只有作者本人可以更新）
func (u *articleUsecase) Update(id, userID uint, req domain.UpdateArticleRequest) (*domain.Article, error) {
	article, err := u.articleRepo.FindByID(id)
	if err != nil {
		return nil, apperror.Wrap(apperror.ErrNotFound, "文章 ID=%d", id)
	}

	if article.UserID != userID {
		return nil, apperror.Wrap(apperror.ErrForbidden, "無權限修改文章 ID=%d", id)
	}

	if req.Title != "" {
		article.Title = req.Title
	}
	if req.Content != "" {
		article.Content = req.Content
	}

	if err := u.articleRepo.Update(article); err != nil {
		return nil, apperror.Wrap(apperror.ErrInternal, "更新文章失敗")
	}

	return article, nil
}

// Delete 刪除文章（只有作者本人可以刪除）
func (u *articleUsecase) Delete(id, userID uint) error {
	article, err := u.articleRepo.FindByID(id)
	if err != nil {
		return apperror.Wrap(apperror.ErrNotFound, "文章 ID=%d", id)
	}

	if article.UserID != userID {
		return apperror.Wrap(apperror.ErrForbidden, "無權限刪除文章 ID=%d", id)
	}

	return u.articleRepo.Delete(id)
}
