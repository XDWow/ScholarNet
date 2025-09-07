package service

import (
	"context"
	"github.com/XD/ScholarNet/cmd/search/domain"
	"github.com/XD/ScholarNet/cmd/search/repository"
)

// 实际就是 写
type SyncService interface {
	InputArticle(ctx context.Context, article domain.Article) error
	InputUser(ctx context.Context, user domain.User) error
	// ...
	// 通用的，考虑一些通用元素
	InputAny(ctx context.Context, index, docID, data string) error
}

type syncService struct {
	userRepo    repository.UserRepository
	articleRepo repository.ArticleRepository
	anyRepo     repository.AnyRepository
}

func NewSyncService(userRepo repository.UserRepository,
	anyRepo repository.AnyRepository,
	articleRepo repository.ArticleRepository) SyncService {
	return &syncService{userRepo: userRepo, anyRepo: anyRepo, articleRepo: articleRepo}
}

func (s *syncService) InputArticle(ctx context.Context, article domain.Article) error {
	return s.articleRepo.InputArticle(ctx, article)
}

func (s *syncService) InputUser(ctx context.Context, user domain.User) error {
	return s.userRepo.InputUser(ctx, user)
}

func (s *syncService) InputAny(ctx context.Context, index, docID, data string) error {
	return s.anyRepo.Input(ctx, index, docID, data)
}
