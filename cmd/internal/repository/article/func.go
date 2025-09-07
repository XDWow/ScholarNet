package article

import (
	"github.com/XD/ScholarNet/cmd/internal/domain"
	"github.com/XD/ScholarNet/cmd/internal/repository/dao/article"
)

func ToEntity(art domain.Article) article.Article {
	return article.Article{
		Title:    art.Title,
		Content:  art.Content,
		Id:       art.Id,
		AuthorId: art.Author.Id,
	}
}
