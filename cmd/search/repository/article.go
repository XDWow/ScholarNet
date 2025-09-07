package repository

import (
	"context"
	"github.com/XD/ScholarNet/cmd/search/domain"
	"github.com/XD/ScholarNet/cmd/search/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type articleRepository struct {
	dao  dao.ArticleDAO
	tags dao.TagDAO
}

func NewArticleRepository(dao dao.ArticleDAO, tags dao.TagDAO) ArticleRepository {
	return &articleRepository{dao: dao, tags: tags}
}

func (repo *articleRepository) InputArticle(ctx context.Context, art domain.Article) error {
	return repo.dao.InputArticle(ctx, dao.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Status:  art.Status,
		Tags:    art.Tags,
	})
}

func (repo *articleRepository) SearchArticle(ctx context.Context, uid int64, keywords []string) ([]domain.Article, error) {
	// keywords 命中文章的标签，传 uid 才知道哪些标签
	ids, err := repo.tags.Search(ctx, uid, "article", keywords)
	if err != nil {
		return nil, err
	}
	// 加一个 bizids 的输入，这个 bizid 是标签含有关键字的 biz_id
	arts, err := repo.dao.Search(ctx, ids, keywords)
	if err != nil {
		return nil, err
	}
	return slice.Map(arts, func(idx int, src dao.Article) domain.Article {
		return domain.Article{
			Id:      src.Id,
			Title:   src.Title,
			Content: src.Content,
			Status:  src.Status,
			Tags:    src.Tags,
		}
	}), nil
}
