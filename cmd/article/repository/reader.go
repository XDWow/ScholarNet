package repository

import (
	"context"
	"github.com/XD/ScholarNet/cmd/article/domain"
	articleDao "github.com/XD/ScholarNet/cmd/article/repository/dao"
)

//go:generate mockgen -source=./reader.go -package=repomocks -destination=mocks/article_reader.mock.go ArticleReaderRepository
type ArticleReaderRepository interface {
	Save(ctx context.Context, art domain.Article) error
}

func NewCachedArticleReaderRepository(dao articleDao.ArticleReaderDAO) ArticleReaderRepository {
	return &CachedArticleReaderRepository{
		dao: dao,
	}
}

type CachedArticleReaderRepository struct {
	dao articleDao.ArticleReaderDAO
}

func (repo *CachedArticleReaderRepository) Save(ctx context.Context, art domain.Article) error {
	return repo.dao.Upsert(ctx, repo.toEntity(art))
}

// toEntity 理论上来说各个 repository 都有差异，所以复制粘贴也没关系。
// 做成一个包方法也可以，看你喜好。
func (repo *CachedArticleReaderRepository) toEntity(art domain.Article) articleDao.Article {
	return articleDao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	}
}
