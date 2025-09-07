package article

import (
	"context"
	"github.com/XD/ScholarNet/cmd/internal/domain"
	"github.com/XD/ScholarNet/cmd/internal/repository/dao/article"
)

type ArticleReaderRepository interface {
	// Save 有就更新，没有就新建，即 upsert 的语义
	Save(ctx context.Context, art domain.Article) (int64, error)
}

type CachedArticleReaderRepository struct {
	dao article.ArticleDAO
}

func NewArticleReaderRepository(dao article.ArticleDAO) ArticleReaderRepository {
	return &CachedArticleReaderRepository{
		dao: dao,
	}
}

func (r *CachedArticleReaderRepository) Save(ctx context.Context, article domain.Article) (int64, error) {
	var err error
	if article.Id > 0 {
		err = r.dao.UpdateById(ctx, ToEntity(article))
	} else {
		article.Id, err = r.dao.Insert(ctx, ToEntity(article))
	}
	if err != nil {
		return 0, err
	}
	return article.Id, nil
}
