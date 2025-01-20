package article

import (
	"context"
	"github.com/LXD-c/basic-go/webook/internal/domain"
	"github.com/LXD-c/basic-go/webook/internal/repository/dao/article"
)

type ArticleAuthorRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
}

type CachedArticleAuthorRepository struct {
	dao article.ArticleDAO
}

func (c *CachedArticleAuthorRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	return c.dao.Insert(ctx, ToEntity(article))
}

func (c *CachedArticleAuthorRepository) Update(ctx context.Context, article domain.Article) error {
	return c.dao.UpdateById(ctx, ToEntity(article))
}

func NewArticleAuthorRepository(dao article.ArticleDAO) ArticleAuthorRepository {
	return &CachedArticleAuthorRepository{
		dao: dao,
	}
}
