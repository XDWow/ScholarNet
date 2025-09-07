package cache

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"github.com/XD/ScholarNet/cmd/ranking/domain"
	"github.com/XD/ScholarNet/cmd/ranking/events"
	"github.com/ecodeclub/ekit/syncx/atomicx"
)

// RankingLocalCache 因为本身数据只有一份，所以不需要借助真正的本地缓存
type RankingLocalCache struct {
	topN       *atomicx.Value[[]domain.Article]
	ddl        *atomicx.Value[time.Time]
	hash       *atomicx.Value[string] // 本地缓存中热榜的文章 id 排成的 hash 值
	producer   events.Producer
	expiration time.Duration
	l          logger.LoggerV1
}

func NewRankingLocalCache(producer events.Producer, l logger.LoggerV1) *RankingLocalCache {
	return &RankingLocalCache{
		topN:       atomicx.NewValue[[]domain.Article](),
		ddl:        atomicx.NewValueOf[time.Time](time.Now()),
		hash:       atomicx.NewValueOf[string](""), // 初始为空
		producer:   producer,
		expiration: time.Minute * 5,
		l:          l,
	}
}

// 计算热榜文章的哈希值，如果与上次相同则不更新，否则更新成功，发送消息
func (r *RankingLocalCache) Set(_ context.Context, arts []domain.Article) error {
	n := len(arts)
	if n == 0 {
		return errors.New("没有文章数据")
	}
	ids := make([]string, n)
	for i := 0; i < n; i++ {
		ids[i] = strconv.FormatInt(arts[i].Id, 10)
	}
	newHash := hashIDs(ids)
	// 本地缓存变化？：我从 redis 拿到的数据，跟我本地缓存数据是否一样
	if newHash != r.hash.Load() {
		r.ddl.Store(time.Now().Add(r.expiration))
		r.topN.Store(arts)
		r.hash.Store(newHash)
		err := r.producer.ProduceUpdateEvent(context.Background(), events.LocalCacheUpdateMessage{
			Timestamp: time.Now().Unix(),
			Articles:  arts,
		})
		if err != nil {
			r.l.Error("发送新热榜数据失败")
		}
	}
	return nil
}

func (r *RankingLocalCache) Get(_ context.Context) ([]domain.Article, error) {
	arts := r.topN.Load()
	if len(arts) == 0 || r.ddl.Load().Before(time.Now()) {
		return nil, errors.New("本地缓存失效了")
	}
	return arts, nil
}

func (r *RankingLocalCache) ForceGet(_ context.Context) ([]domain.Article, error) {
	return r.topN.Load(), nil
}

// UpdateFromMessage 从消息更新本地缓存（消费者专用，不发送消息）
func (r *RankingLocalCache) UpdateFromMessage(ctx context.Context, arts []domain.Article) error {
	n := len(arts)
	if n == 0 {
		return errors.New("没有文章数据")
	}

	// 直接更新缓存，不检查哈希值，不发送消息
	r.ddl.Store(time.Now().Add(r.expiration))
	r.topN.Store(arts)

	// 更新哈希值用于后续比较
	ids := make([]string, n)
	for i := 0; i < n; i++ {
		ids[i] = strconv.FormatInt(arts[i].Id, 10)
	}
	r.hash.Store(hashIDs(ids))

	return nil
}

func hashIDs(ids []string) string {
	str := strings.Join(ids, ",")
	h := sha1.Sum([]byte(str))
	return hex.EncodeToString(h[:])
}
