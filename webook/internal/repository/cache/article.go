package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/LXD-c/basic-go/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

type ArticleCache interface {
	// GetFirstPage 只缓存第第一页的数据
	// 并且不缓存整个 Content
	SetFirstPage(ctx context.Context, author int64, arts []domain.Article) error
	GetFirstPage(ctx context.Context, author int64) ([]domain.Article, error)
	DelFirstPage(ctx context.Context, author int64) error

	Set(ctx context.Context, art domain.Article) error
	Get(ctx context.Context, id int64) (domain.Article, error)

	// SetPub 正常来说，创作者和读者的 Redis 集群要分开，因为读者是一个核心中的核心
	SetPub(ctx context.Context, article domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
}

type RedisArticleCache struct {
	client redis.Cmdable
}

func NewRedisArticleCache(client redis.Cmdable) ArticleCache {
	return &RedisArticleCache{
		client: client,
	}
}

func (c *RedisArticleCache) SetPub(ctx context.Context, art domain.Article) error {
	data, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.readerArtKey(art.Id), data, time.Minute).Err()
}

func (c *RedisArticleCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	data, err := c.client.Get(ctx, c.readerArtKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var art domain.Article
	err = json.Unmarshal(data, &art)
	return art, err
}

func (c *RedisArticleCache) Set(ctx context.Context, art domain.Article) error {
	data, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.authorArtKey(art.Id), data, time.Minute).Err()
}

func (c *RedisArticleCache) Get(ctx context.Context, id int64) (domain.Article, error) {
	// 可以直接使用 Bytes 方法来获得 []byte
	data, err := c.client.Get(ctx, c.authorArtKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var art domain.Article
	err = json.Unmarshal(data, &art)
	return art, err
}
func (c *RedisArticleCache) SetFirstPage(ctx context.Context, author int64, arts []domain.Article) error {
	for i := range arts {
		// 只缓存摘要部分
		arts[i].Content = arts[i].Abstract()
	}
	bs, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.firstPageKey(author), bs, time.Minute*10).Err()
}

func (c *RedisArticleCache) GetFirstPage(ctx context.Context, author int64) ([]domain.Article, error) {
	bs, err := c.client.Get(ctx, c.firstPageKey(author)).Bytes()
	if err != nil {
		return nil, err
	}
	var arts []domain.Article
	err = json.Unmarshal(bs, &arts)
	return arts, err
}

func (c *RedisArticleCache) DelFirstPage(ctx context.Context, author int64) error {
	return c.client.Del(ctx, c.firstPageKey(author)).Err()
}
func (c *RedisArticleCache) firstPageKey(author int64) string {
	return fmt.Sprintf("article:first_page:%d", author)
}

// 创作端的缓存设置
func (c *RedisArticleCache) authorArtKey(id int64) string {
	return fmt.Sprintf("article:author:%d", id)
}

// 读者端的缓存设置
func (c *RedisArticleCache) readerArtKey(id int64) string {
	return fmt.Sprintf("article:reader:%d", id)
}
