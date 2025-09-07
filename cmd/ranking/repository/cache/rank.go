package cache

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/XD/ScholarNet/cmd/ranking/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

const key = "ranking:article"

type RankingCache interface {
	Set(ctx context.Context, arts []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}

type RedisRankingCache struct {
	client     redis.Cmdable
	key        string
	expiration time.Duration
	pubsubKey  string
}

func NewRedisRankingCache(client redis.Cmdable) *RedisRankingCache {
	return &RedisRankingCache{
		key:        key,
		client:     client,
		expiration: time.Minute * 10,
		pubsubKey:  key + ":notify",
	}
}

// 更新排行榜
func (r *RedisRankingCache) Set(ctx context.Context, arts []domain.Article) error {
	tempKey := r.key + ":temp"

	// 构建ZSET
	zs := make([]redis.Z, 0, len(arts))
	for _, art := range arts {
		art.Content = art.Abstract()
		data, err := json.Marshal(art)
		if err != nil {
			return err
		}
		zs = append(zs, redis.Z{
			Score:  art.Score,
			Member: data,
		})
	}

	if len(zs) == 0 {
		return errors.New("文章列表为空")
	}

	// 写 Redis
	pipe := r.client.TxPipeline()
	pipe.Del(ctx, tempKey)
	pipe.ZAdd(ctx, tempKey, zs...)
	pipe.Expire(ctx, tempKey, r.expiration+time.Minute)
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	pipe = r.client.TxPipeline()
	pipe.Rename(ctx, tempKey, r.key)
	pipe.Expire(ctx, r.key, r.expiration)
	_, err := pipe.Exec(ctx)
	return err
}

//func (r *RedisRankingCache) Set(ctx context.Context, arts []domain.Article) error {
//	// 使用临时键构建新排行榜
//	tempKey := r.key + ":temp"
//	pipe := r.client.Pipeline()
//
//	// 1. 删除可能残留的临时键
//	pipe.Del(ctx, tempKey)
//
//	// 2. 准备文章数据
//	zs := make([]redis.Z, 0, len(arts))
//	for _, art := range arts {
//		// 仅存储摘要内容
//		art.Content = art.Abstract()
//		data, err := json.Marshal(art)
//		if err != nil {
//			pipe.Discard()
//			return err
//		}
//		zs = append(zs, redis.Z{
//			Score:  art.Score,
//			Member: data,
//		})
//	}
//
//	// 3. 添加数据到临时键
//	if len(zs) > 0 {
//		pipe.ZAdd(ctx, tempKey, zs...)
//	} else {
//		return errors.New("文章列表为空")
//	}
//
//	// 4. 防御性编程，设置临时键过期时间，避免这里突然中断，临时键成为”僵尸键“一直存在（增加缓冲防止重名前过期）
//	pipe.Expire(ctx, tempKey, r.expiration+time.Minute)
//
//	// 5. 执行第一组命令
//	if _, err := pipe.Exec(ctx); err != nil {
//		return err
//	}
//
//	// 6. 原子操作：重命名键+设置过期时间
//	pipe = r.client.Pipeline()
//	// Rename 原子操作：如果目标键已经存在，自动删除它
//	pipe.Rename(ctx, tempKey, r.key)
//	pipe.Expire(ctx, r.key, r.expiration)
//	_, err := pipe.Exec(ctx)
//	return err
//}

// 会出现空窗，并且不具备原子性，导致删除了，但没有写入的情况
//func (r *RedisRankingCache) Set(ctx context.Context, arts []domain.Article) error {
//	pipe := r.client.Pipeline()
//	// 先删除原有ZSet
//	pipe.Del(ctx, r.key)
//	// 只存摘要
//	for _, art := range arts {
//		art.Content = art.Abstract()
//		val, _ := json.Marshal(art)
//		pipe.ZAdd(ctx, r.key, redis.Z{
//			Score:  art.Score, // 直接用Article的Score字段
//			Member: val,
//		})
//	}
//	_, err := pipe.Exec(ctx)
//	return err
//}

func (r *RedisRankingCache) Get(ctx context.Context) ([]domain.Article, error) {
	// 取前100名
	vals, err := r.client.ZRevRange(ctx, r.key, 0, 99).Result()
	if err != nil {
		return nil, err
	}
	res := make([]domain.Article, 0, len(vals))
	for _, v := range vals {
		var art domain.Article
		_ = json.Unmarshal([]byte(v), &art)
		res = append(res, art)
	}
	return res, nil
}
