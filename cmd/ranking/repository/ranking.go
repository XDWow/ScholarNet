package repository

import (
	"context"
	"github.com/XD/ScholarNet/cmd/ranking/domain"
	"github.com/XD/ScholarNet/cmd/ranking/repository/cache"
	"github.com/ecodeclub/ekit/syncx/atomicx"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
	RefreshLocalCache(ctx context.Context) error
	RefreshLocalCacheV1(ctx context.Context) error
}

type CachedRankingRepository struct {
	redisCache *cache.RedisRankingCache
	localCache *cache.RankingLocalCache
	// 你也可以考虑将这个本地缓存塞进去 RankingCache 里面，作为一个实现
	topN atomicx.Value[[]domain.Article]
}

func NewCachedRankingRepository(
	redisCache *cache.RedisRankingCache,
	localCache *cache.RankingLocalCache) RankingRepository {
	repo := &CachedRankingRepository{
		redisCache: redisCache,
		localCache: localCache,
	}
	return repo
}

func (c *CachedRankingRepository) ReplaceTopN(ctx context.Context,
	arts []domain.Article) error {
	// 这一步必然不会出错
	_ = c.localCache.Set(ctx, arts)
	return c.redisCache.Set(ctx, arts)
}

func (c *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	arts, err := c.localCache.Get(ctx)
	if err == nil {
		return arts, nil
	}
	arts, err = c.redisCache.Get(ctx)
	if err == nil {
		_ = c.localCache.Set(ctx, arts)
	} else {
		// 降级兜底策略，拿本地缓存的老数据
		// 这里，我们没有进一步区分是什么原因导致的 Redis 错误
		return c.localCache.ForceGet(ctx)
	}
	return arts, err
}

// 定时任务用的
func (c *CachedRankingRepository) RefreshLocalCache(ctx context.Context) error {
	arts, err := c.redisCache.Get(ctx)
	if err != nil {
		return err
	}
	return c.localCache.Set(ctx, arts)
}

// redis 订阅专用
func (c *CachedRankingRepository) RefreshLocalCacheV1(ctx context.Context) error {
	panic("不想实现你！")
}
