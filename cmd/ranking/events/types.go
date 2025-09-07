package events

import (
	"context"

	"github.com/XD/ScholarNet/cmd/ranking/domain"
)

type Consumer interface {
	Start() error
}

type Producer interface {
	ProduceUpdateEvent(ctx context.Context, evt LocalCacheUpdateMessage) error
}

// LocalCache 抽象本地缓存能力，供消费者在收到消息时更新本地缓存使用
// 将接口定义放在 events 包，避免 events 依赖具体的 cache 实现
type LocalCache interface {
	UpdateFromMessage(ctx context.Context, arts []domain.Article) error
}

type LocalCacheUpdateMessage struct {
	NodeID    string           `json:"node_id"`
	Timestamp int64            `json:"timestamp"`
	Articles  []domain.Article `json:"articles"`
}
