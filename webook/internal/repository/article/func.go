package article

import (
	"github.com/LXD-c/basic-go/webook/internal/domain"
	"github.com/LXD-c/basic-go/webook/internal/repository/dao/article"
)

func ToEntity(art domain.Article) article.Article {
	return article.Article{
		Title:    art.Title,
		Content:  art.Content,
		Id:       art.Id,
		AuthorId: art.Author.Id,
	}
}
