package events

import (
	"context"
	"fmt"
	"time"

	"github.com/XD/ScholarNet/cmd/pkg/saramax"

	"github.com/IBM/sarama"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
)

type SaramaConsumer struct {
	client     sarama.Client
	l          logger.LoggerV1
	localCache LocalCache
	nodeID     string
	lastUpdate time.Time
	cooldown   time.Duration
}

func NewArticleEventConsumer(
	client sarama.Client,
	l logger.LoggerV1,
	localCache LocalCache,
	nodeID string) *SaramaConsumer {
	c := &SaramaConsumer{
		localCache: localCache,
		client:     client,
		l:          l,
		nodeID:     nodeID,
		cooldown:   10 * time.Second, // 10秒冷却时间
	}
	return c
}

// Start 这边就是自己启动 goroutine 了
func (r *SaramaConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("articleRanking",
		r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{topicUpdateEvent},
			saramax.NewHandler[LocalCacheUpdateMessage](r.l, r.Consume))
		if err != nil {
			r.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (r *SaramaConsumer) Consume(msg *sarama.ConsumerMessage,
	evt LocalCacheUpdateMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 1. 检查是否是自己发送的消息
	if evt.NodeID == r.nodeID {
		r.l.Info("收到自己发送的消息，跳过更新",
			logger.String("nodeID", r.nodeID),
			logger.Int64("timestamp", evt.Timestamp))
		return nil
	}

	// 2. 冷却时间（10秒内不重复更新）
	if time.Since(r.lastUpdate) < r.cooldown {
		r.l.Info("冷却时间内，跳过更新",
			logger.String("nodeID", r.nodeID),
			logger.String("cooldown", r.cooldown.String()),
			logger.String("lastUpdate", r.lastUpdate.Format(time.RFC3339)))
		return nil
	}

	// 3. 直接从消息体中获取 articles 切片并更新本地缓存
	articles := evt.Articles
	if len(articles) > 0 {
		// 使用专门的方法更新本地缓存（不触发消息发送）
		if err := r.localCache.UpdateFromMessage(ctx, articles); err != nil {
			return fmt.Errorf("更新本地缓存失败: %w", err)
		}

		r.lastUpdate = time.Now()
		r.l.Info("成功更新本地缓存",
			logger.String("nodeID", r.nodeID),
			logger.Int("articleCount", len(articles)),
			logger.String("fromNodeID", evt.NodeID))
	}

	return nil
}
