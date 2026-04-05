package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"blog-api/internal/domain"
	"blog-api/pkg/cache"

	"github.com/redis/go-redis/v9"
)

const articleCacheTTL = 5 * time.Minute

// CachedArticleRepository 在 ArticleRepository 外層加上快取
type CachedArticleRepository struct {
	repo  domain.ArticleRepository
	cache cache.Cache
}

// NewCachedArticleRepository 建立帶快取的文章 Repository
func NewCachedArticleRepository(repo domain.ArticleRepository, c cache.Cache) domain.ArticleRepository {
	return &CachedArticleRepository{repo: repo, cache: c}
}

func (r *CachedArticleRepository) Create(article *domain.Article) error {
	return r.repo.Create(article)
}

func (r *CachedArticleRepository) FindByID(id uint) (*domain.Article, error) {
	ctx := context.Background()
	key := fmt.Sprintf("article:%d", id)

	// 嘗試從快取讀取
	var cached domain.Article
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		slog.Debug("快取命中", "key", key)
		return &cached, nil
	} else if err != redis.Nil {
		slog.Warn("快取讀取失敗，改從資料庫查詢", "key", key, "error", err)
	}

	// 快取未命中，從資料庫查詢
	article, err := r.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// 寫入快取（失敗不影響回傳）
	if err := r.cache.Set(ctx, key, article, articleCacheTTL); err != nil {
		slog.Warn("快取寫入失敗", "key", key, "error", err)
	}

	return article, nil
}

func (r *CachedArticleRepository) FindAll(query domain.ArticleQuery) ([]domain.Article, int64, error) {
	// 列表查詢不做快取（條件太多，快取效益低）
	return r.repo.FindAll(query)
}

func (r *CachedArticleRepository) Update(article *domain.Article) error {
	if err := r.repo.Update(article); err != nil {
		return err
	}
	// 更新後清除快取
	ctx := context.Background()
	key := fmt.Sprintf("article:%d", article.ID)
	r.cache.Delete(ctx, key)
	return nil
}

func (r *CachedArticleRepository) Delete(id uint) error {
	if err := r.repo.Delete(id); err != nil {
		return err
	}
	// 刪除後清除快取
	ctx := context.Background()
	key := fmt.Sprintf("article:%d", id)
	r.cache.Delete(ctx, key)
	return nil
}
