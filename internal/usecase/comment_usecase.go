package usecase

import (
	"errors"

	"blog-api/internal/domain"
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
// 同時需要 CommentRepository 和 ArticleRepository，用於驗證文章是否存在
func NewCommentUsecase(commentRepo domain.CommentRepository, articleRepo domain.ArticleRepository) CommentUsecase {
	return &commentUsecase{
		commentRepo: commentRepo,
		articleRepo: articleRepo,
	}
}

// Create 建立新留言
// 會先驗證目標文章是否存在
func (u *commentUsecase) Create(articleID, userID uint, req domain.CreateCommentRequest) (*domain.Comment, error) {
	// 確認文章存在
	if _, err := u.articleRepo.FindByID(articleID); err != nil {
		return nil, errors.New("文章不存在")
	}

	comment := &domain.Comment{
		Content:   req.Content,
		ArticleID: articleID,
		UserID:    userID,
	}

	if err := u.commentRepo.Create(comment); err != nil {
		return nil, errors.New("建立留言失敗")
	}

	// 重新查詢以載入關聯的使用者資料
	return u.commentRepo.FindByID(comment.ID)
}

// GetByArticleID 取得某篇文章的所有留言
func (u *commentUsecase) GetByArticleID(articleID uint) ([]domain.Comment, error) {
	return u.commentRepo.FindByArticleID(articleID)
}

// Update 更新留言
// 只有留言者本人可以更新
func (u *commentUsecase) Update(id, userID uint, req domain.UpdateCommentRequest) (*domain.Comment, error) {
	comment, err := u.commentRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("留言不存在")
	}

	// 檢查是否為留言者本人
	if comment.UserID != userID {
		return nil, errors.New("無權限修改此留言")
	}

	comment.Content = req.Content

	if err := u.commentRepo.Update(comment); err != nil {
		return nil, errors.New("更新留言失敗")
	}

	return comment, nil
}

// Delete 刪除留言
// 只有留言者本人可以刪除
func (u *commentUsecase) Delete(id, userID uint) error {
	comment, err := u.commentRepo.FindByID(id)
	if err != nil {
		return errors.New("留言不存在")
	}

	// 檢查是否為留言者本人
	if comment.UserID != userID {
		return errors.New("無權限刪除此留言")
	}

	return u.commentRepo.Delete(id)
}
