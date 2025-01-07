package service

import (
	"context"
	"github.com/LXD-c/basic-go/webook/internal/domain"
)

type ArticleService interface {
	Save(ctx context.Context, article domain.Article) error
	Publish(ctx context.Context, article domain.Article) error
}
