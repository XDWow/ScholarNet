package repository

import (
	"context"
	"github.com/XD/ScholarNet/cmd/article/domain"
	articleDAO "github.com/XD/ScholarNet/cmd/article/repository/dao"
)

// ArticleAuthorRepository 演示在 service 层面上分流
//
//go:generate mockgen -source=./author.go -package=repomocks -destination=mocks/article_author.mock.go ArticleAuthorRepository
type ArticleAuthorRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
}

// CachedArticleAuthorRepository 按照道理，这里也是可以搞缓存的
type CachedArticleAuthorRepository struct {
	dao articleDAO.ArticleDAO
}

func NewArticleAuthorRepository(dao articleDAO.ArticleDAO) ArticleAuthorRepository {
	return &CachedArticleAuthorRepository{
		dao: dao,
	}
}

func (repo *CachedArticleAuthorRepository) Create(ctx context.Context,
	art domain.Article) (int64, error) {
	return repo.dao.Insert(ctx, repo.toEntity(art))
}

func (repo *CachedArticleAuthorRepository) Update(ctx context.Context,
	art domain.Article) error {
	return repo.dao.UpdateById(ctx, repo.toEntity(art))
}

func (repo *CachedArticleAuthorRepository) toEntity(art domain.Article) articleDAO.Article {
	return articleDAO.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	}
}
