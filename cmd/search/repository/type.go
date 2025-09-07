package repository

import (
	"context"
	"github.com/XD/ScholarNet/cmd/search/domain"
)

type UserRepository interface {
	InputUser(ctx context.Context, msg domain.User) error
	SearchUser(ctx context.Context, keywords []string) ([]domain.User, error)
}

type ArticleRepository interface {
	InputArticle(ctx context.Context, msg domain.Article) error
	SearchArticle(ctx context.Context, uid int64, keywords []string) ([]domain.Article, error)
}

type AnyRepository interface {
	Input(ctx context.Context, index string, docID string, data string) error
}
