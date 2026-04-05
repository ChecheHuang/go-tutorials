package usecase

import (
	"blog-api/internal/domain"
	"blog-api/pkg/apperror"
)

// CommentUsecase 定義留言業務邏輯的介面
type CommentUsecase interface {
	Create(articleID, userID uint, req domain.CreateCommentRequest) (*domain.Comment, error)
	GetByArticleID(articleID uint) ([]domain.Comment, error)
	Update(id, userID uint, req domain.UpdateCommentRequest) (*domain.Comment, error)
	Delete(id, userID uint) error
}

// commentUsecase 實作 CommentUsecase 介面
type commentUsecase struct {
	commentRepo domain.CommentRepository
	articleRepo domain.ArticleRepository
}

// NewCommentUsecase 建立留言 Usecase 實例
func NewCommentUsecase(commentRepo domain.CommentRepository, articleRepo domain.ArticleRepository) CommentUsecase {
	return &commentUsecase{
		commentRepo: commentRepo,
		articleRepo: articleRepo,
	}
}

// Create 建立新留言（會先驗證目標文章是否存在）
func (u *commentUsecase) Create(articleID, userID uint, req domain.CreateCommentRequest) (*domain.Comment, error) {
	if _, err := u.articleRepo.FindByID(articleID); err != nil {
		return nil, apperror.Wrap(apperror.ErrNotFound, "文章 ID=%d", articleID)
	}

	comment := &domain.Comment{
		Content:   req.Content,
		ArticleID: articleID,
		UserID:    userID,
	}

	if err := u.commentRepo.Create(comment); err != nil {
		return nil, apperror.Wrap(apperror.ErrInternal, "建立留言失敗")
	}

	return u.commentRepo.FindByID(comment.ID)
}

// GetByArticleID 取得某篇文章的所有留言
func (u *commentUsecase) GetByArticleID(articleID uint) ([]domain.Comment, error) {
	return u.commentRepo.FindByArticleID(articleID)
}

// Update 更新留言（只有留言者本人可以更新）
func (u *commentUsecase) Update(id, userID uint, req domain.UpdateCommentRequest) (*domain.Comment, error) {
	comment, err := u.commentRepo.FindByID(id)
	if err != nil {
		return nil, apperror.Wrap(apperror.ErrNotFound, "留言 ID=%d", id)
	}

	if comment.UserID != userID {
		return nil, apperror.Wrap(apperror.ErrForbidden, "無權限修改留言 ID=%d", id)
	}

	comment.Content = req.Content

	if err := u.commentRepo.Update(comment); err != nil {
		return nil, apperror.Wrap(apperror.ErrInternal, "更新留言失敗")
	}

	return comment, nil
}

// Delete 刪除留言（只有留言者本人可以刪除）
func (u *commentUsecase) Delete(id, userID uint) error {
	comment, err := u.commentRepo.FindByID(id)
	if err != nil {
		return apperror.Wrap(apperror.ErrNotFound, "留言 ID=%d", id)
	}

	if comment.UserID != userID {
		return apperror.Wrap(apperror.ErrForbidden, "無權限刪除留言 ID=%d", id)
	}

	return u.commentRepo.Delete(id)
}
