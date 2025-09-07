package repository

import (
	"context"
	"github.com/XD/ScholarNet/cmd/code/repository/cache"
)

var (
	ErrCodeVerifyTooManyTimes = cache.ErrCodeVerifyTooManyTimes
	ErrCodeSendTooMany        = cache.ErrCodeSendTooMany
)

//go:generate mockgen -source=./code.go -package=repomocks -destination=mocks/code.mock.go CodeRepository
type CodeRepository interface {
	Store(ctx context.Context, biz string,
		phone string, code string) error

	Verify(ctx context.Context, biz string,
		phone string, inputCode string) (bool, error)
}

type CachedCodeRepository struct {
	cache cache.CodeCache
}

func NewCachedCodeRepository(c cache.CodeCache) CodeRepository {
	return &CachedCodeRepository{
		cache: c,
	}
}

func (repo *CachedCodeRepository) Store(ctx context.Context,
	biz string,
	phone string,
	code string) error {
	err := repo.cache.Set(ctx, biz, phone, code)
	return err
}

// Verify 比较验证码。如果验证码相等，那么删除；
func (repo *CachedCodeRepository) Verify(ctx context.Context,
	biz string, phone string, inputCode string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phone, inputCode)
}
