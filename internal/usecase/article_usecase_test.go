package usecase

import (
	"errors"
	"strings"
	"testing"

	"blog-api/internal/domain"
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
		// 模擬搜尋功能
		if query.Search != "" {
			if !strings.Contains(article.Title, query.Search) && !strings.Contains(article.Content, query.Search) {
				continue
			}
		}
		// 模擬依作者篩選
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

// TestCreateArticle_Success 測試建立文章
func TestCreateArticle_Success(t *testing.T) {
	repo := newMockArticleRepository()
	uc := NewArticleUsecase(repo)

	article, err := uc.Create(1, domain.CreateArticleRequest{
		Title:   "測試文章",
		Content: "這是測試內容",
	})
	if err != nil {
		t.Fatalf("預期建立成功，但得到錯誤：%v", err)
	}

	if article.Title != "測試文章" {
		t.Errorf("標題不符：預期「測試文章」，得到「%s」", article.Title)
	}

	if article.UserID != 1 {
		t.Errorf("作者 ID 不符：預期 1，得到 %d", article.UserID)
	}
}

// TestUpdateArticle_OwnerOnly 測試只有作者可以更新文章
func TestUpdateArticle_OwnerOnly(t *testing.T) {
	repo := newMockArticleRepository()
	uc := NewArticleUsecase(repo)

	// 使用者 1 建立文章
	article, _ := uc.Create(1, domain.CreateArticleRequest{
		Title:   "原始標題",
		Content: "原始內容",
	})

	// 使用者 2 嘗試更新（應該失敗）
	_, err := uc.Update(article.ID, 2, domain.UpdateArticleRequest{
		Title: "新標題",
	})
	if err == nil {
		t.Error("非作者更新文章應該失敗")
	}

	// 使用者 1 更新（應該成功）
	updated, err := uc.Update(article.ID, 1, domain.UpdateArticleRequest{
		Title: "新標題",
	})
	if err != nil {
		t.Fatalf("作者更新文章應該成功：%v", err)
	}

	if updated.Title != "新標題" {
		t.Errorf("標題未更新：預期「新標題」，得到「%s」", updated.Title)
	}
}

// TestDeleteArticle_OwnerOnly 測試只有作者可以刪除文章
func TestDeleteArticle_OwnerOnly(t *testing.T) {
	repo := newMockArticleRepository()
	uc := NewArticleUsecase(repo)

	// 使用者 1 建立文章
	article, _ := uc.Create(1, domain.CreateArticleRequest{
		Title:   "待刪除文章",
		Content: "內容",
	})

	// 使用者 2 嘗試刪除（應該失敗）
	err := uc.Delete(article.ID, 2)
	if err == nil {
		t.Error("非作者刪除文章應該失敗")
	}

	// 使用者 1 刪除（應該成功）
	err = uc.Delete(article.ID, 1)
	if err != nil {
		t.Fatalf("作者刪除文章應該成功：%v", err)
	}

	// 確認文章已被刪除
	_, err = uc.GetByID(article.ID)
	if err == nil {
		t.Error("已刪除的文章不應該被找到")
	}
}

// TestGetAllArticles_WithSearch 測試文章搜尋功能
func TestGetAllArticles_WithSearch(t *testing.T) {
	repo := newMockArticleRepository()
	uc := NewArticleUsecase(repo)

	// 建立多篇文章
	uc.Create(1, domain.CreateArticleRequest{Title: "Go 教學", Content: "學習 Go 語言"})
	uc.Create(1, domain.CreateArticleRequest{Title: "Python 入門", Content: "學習 Python"})
	uc.Create(2, domain.CreateArticleRequest{Title: "Go 進階", Content: "Go 的進階技巧"})

	// 搜尋 "Go"
	articles, total, err := uc.GetAll(domain.ArticleQuery{Search: "Go"})
	if err != nil {
		t.Fatalf("搜尋失敗：%v", err)
	}

	if total != 2 {
		t.Errorf("搜尋 'Go' 應該找到 2 篇文章，但找到 %d 篇", total)
	}

	_ = articles // 使用變數避免警告
}
