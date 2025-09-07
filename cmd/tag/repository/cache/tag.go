package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/XD/ScholarNet/cmd/tag/domain"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

var ErrKeyNotExist = redis.Nil

type TagCache interface {
	// Tag 业务上只能加，不能更新或者删
	Append(ctx context.Context, uid int64, tags ...domain.Tag) error
	GetTags(ctx context.Context, uid int64) ([]domain.Tag, error)
	// 移除出 cache
	DelTags(ctx context.Context, uid int64) error
}

// Preload 全量加载？
func Preload(ctx context.Context) {
	// 你需要 gorm.DB
}

type RedisTagCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func (r *RedisTagCache) Append(ctx context.Context, uid int64, tags ...domain.Tag) error {
	key := r.userTagsKey(uid)
	pipe := r.client.Pipeline()
	for _, tag := range tags {
		val, err := json.Marshal(tag)
		if err != nil {
			return err
		}
		pipe.HSet(ctx, key, strconv.FormatInt(uid, 10), val)
	}
	pipe.Expire(ctx, key, r.expiration)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisTagCache) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	key := r.userTagsKey(uid)
	m, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	res := make([]domain.Tag, 0, len(m))
	for _, v := range m {
		var tag domain.Tag
		err = json.Unmarshal([]byte(v), &tag)
		if err != nil {
			return nil, err
		}
		res = append(res, tag)
	}
	return res, nil
}

func (r *RedisTagCache) DelTags(ctx context.Context, uid int64) error {
	key := r.userTagsKey(uid)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisTagCache) userTagsKey(uid int64) string {
	return fmt.Sprintf("tag:user_tags:%d", uid)
}

func NewRedisTagCache(client redis.Cmdable) TagCache {
	return &RedisTagCache{
		client: client,
	}
}
