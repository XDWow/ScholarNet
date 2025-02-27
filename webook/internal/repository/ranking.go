package repository

import (
	"context"
	"github.com/LXD-c/basic-go/webook/internal/domain"
	"github.com/LXD-c/basic-go/webook/internal/repository/cache"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
}

type CachedRankingRepository struct {
	// 使用具体实现，可读性更好，对测试不友好，因为咩有面向接口编程
	redis cache.RankingCache
	local *cache.RankingLocalCache
}

func NewCachedRankingRepository(redis cache.RankingCache, local *cache.RankingLocalCache) RankingRepository {
	return &CachedRankingRepository{
		redis: redis,
		local: local,
	}
}

func (c *CachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	_ = c.local.Set(arts)
	return c.redis.Set(ctx, arts)
}

func (c *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	data, err := c.local.Get(ctx)
	if err == nil {
		return data, nil
	}
	data, err = c.redis.Get(ctx)
	if err == nil {
		// 本地缓存没找到，redis找到了，那就需要写进本地缓存
		_ = c.local.Set(data)
	} else {
		// 热榜数据没有存数据库的，两个缓存都没找到，要求放低一点，本地缓存过期的数据拿出来用算了，总比没有好
		return c.local.ForceGet(ctx)
	}
	return data, err
}
