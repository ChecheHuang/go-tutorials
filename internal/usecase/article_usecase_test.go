package usecase

import (
	"errors"
	"strings"
	"testing"

	"blog-api/internal/domain"
	"blog-api/pkg/apperror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// === Mock Article Repository ===
type mockArticleRepository struct {
	articles map[uint]*domain.Article
	nextID   uint
}

func newMockArticleRepository() *mockArticleRepository {
	return &mockArticleRepository{
		articles: make(map[uint]*domain.Article),
		nextID:   1,
	}
}

func (m *mockArticleRepository) Create(article *domain.Article) error {
	article.ID = m.nextID
	m.nextID++
	m.articles[article.ID] = article
	return nil
}

func (m *mockArticleRepository) FindByID(id uint) (*domain.Article, error) {
	article, exists := m.articles[id]
	if !exists {
		return nil, errors.New("article not found")
	}
	return article, nil
}

func (m *mockArticleRepository) FindAll(query domain.ArticleQuery) ([]domain.Article, int64, error) {
	var result []domain.Article
	for _, article := range m.articles {
		if query.Search != "" {
			if !strings.Contains(article.Title, query.Search) && !strings.Contains(article.Content, query.Search) {
				continue
			}
		}
		if query.UserID > 0 && article.UserID != query.UserID {
			continue
		}
		result = append(result, *article)
	}
	return result, int64(len(result)), nil
}

func (m *mockArticleRepository) Update(article *domain.Article) error {
	m.articles[article.ID] = article
	return nil
}

func (m *mockArticleRepository) Delete(id uint) error {
	delete(m.articles, id)
	return nil
}

// TestArticleUsecase 使用表格驅動測試
func TestArticleUsecase_Create(t *testing.T) {
	tests := []struct {
		name    string
		userID  uint
		req     domain.CreateArticleRequest
		wantErr bool
	}{
		{
			name:   "建立文章成功",
			userID: 1,
			req:    domain.CreateArticleRequest{Title: "測試文章", Content: "這是測試內容"},
		},
		{
			name:   "不同使用者建立文章",
			userID: 2,
			req:    domain.CreateArticleRequest{Title: "另一篇文章", Content: "另一篇內容"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockArticleRepository()
			uc := NewArticleUsecase(repo)

			article, err := uc.Create(tt.userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.req.Title, article.Title)
			assert.Equal(t, tt.req.Content, article.Content)
			assert.Equal(t, tt.userID, article.UserID)
		})
	}
}

func TestArticleUsecase_GetByID(t *testing.T) {
	repo := newMockArticleRepository()
	uc := NewArticleUsecase(repo)

	// 建立一篇文章
	created, err := uc.Create(1, domain.CreateArticleRequest{Title: "測試", Content: "內容"})
	require.NoError(t, err)

	t.Run("找到文章", func(t *testing.T) {
		article, err := uc.GetByID(created.ID)
		require.NoError(t, err)
		assert.Equal(t, "測試", article.Title)
	})

	t.Run("文章不存在回傳 ErrNotFound", func(t *testing.T) {
		_, err := uc.GetByID(999)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperror.ErrNotFound), "應該是 ErrNotFound，但得到：%v", err)
	})
}

func TestArticleUsecase_Update(t *testing.T) {
	repo := newMockArticleRepository()
	uc := NewArticleUsecase(repo)

	article, _ := uc.Create(1, domain.CreateArticleRequest{Title: "原始標題", Content: "原始內容"})

	t.Run("作者更新成功", func(t *testing.T) {
		updated, err := uc.Update(article.ID, 1, domain.UpdateArticleRequest{Title: "新標題"})
		require.NoError(t, err)
		assert.Equal(t, "新標題", updated.Title)
	})

	t.Run("非作者更新回傳 ErrForbidden", func(t *testing.T) {
		_, err := uc.Update(article.ID, 2, domain.UpdateArticleRequest{Title: "駭客標題"})
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperror.ErrForbidden), "應該是 ErrForbidden，但得到：%v", err)
	})

	t.Run("文章不存在回傳 ErrNotFound", func(t *testing.T) {
		_, err := uc.Update(999, 1, domain.UpdateArticleRequest{Title: "不存在"})
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperror.ErrNotFound))
	})
}

func TestArticleUsecase_Delete(t *testing.T) {
	repo := newMockArticleRepository()
	uc := NewArticleUsecase(repo)

	article, _ := uc.Create(1, domain.CreateArticleRequest{Title: "待刪除", Content: "內容"})

	t.Run("非作者刪除回傳 ErrForbidden", func(t *testing.T) {
		err := uc.Delete(article.ID, 2)
		assert.True(t, errors.Is(err, apperror.ErrForbidden))
	})

	t.Run("作者刪除成功", func(t *testing.T) {
		err := uc.Delete(article.ID, 1)
		require.NoError(t, err)
	})

	t.Run("已刪除的文章不存在", func(t *testing.T) {
		_, err := uc.GetByID(article.ID)
		assert.True(t, errors.Is(err, apperror.ErrNotFound))
	})
}

func TestArticleUsecase_GetAll_WithSearch(t *testing.T) {
	repo := newMockArticleRepository()
	uc := NewArticleUsecase(repo)

	uc.Create(1, domain.CreateArticleRequest{Title: "Go 教學", Content: "學習 Go 語言"})
	uc.Create(1, domain.CreateArticleRequest{Title: "Python 入門", Content: "學習 Python"})
	uc.Create(2, domain.CreateArticleRequest{Title: "Go 進階", Content: "Go 的進階技巧"})

	t.Run("搜尋 Go 找到 2 篇", func(t *testing.T) {
		articles, total, err := uc.GetAll(domain.ArticleQuery{Search: "Go"})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, articles, 2)
	})

	t.Run("依作者篩選", func(t *testing.T) {
		_, total, err := uc.GetAll(domain.ArticleQuery{UserID: 2})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
	})
}
