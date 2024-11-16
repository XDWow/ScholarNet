package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

var ErrUserNotFound = redis.Nil

type UseCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

// NewUserCache
// A 用到了 B, B 一定是接口 => 保证面向接口
// A 用到了 B, B 一定是A的字段 => 避免包变量、包方法，都非常缺乏扩展性
// A 用到了 B, A 绝对不初始化 B, 而是外面注入 => 保持依赖注入（DI）和依赖反转（IOC)
func NewUserCache(client *redis.Client) *UseCache {
	return &UseCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

func (cache *UseCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.key(id)
	val, err := cache.client.Get(ctx, key).Bytes()
	var u domain.User
	err = json.Unmarshal(val, &u)
	return u, err
}

func (cache *UseCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := cache.key(u.Id)
	return cache.client.Set(ctx, key, val, cache.expiration).Err()
}

func (cache *UseCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
