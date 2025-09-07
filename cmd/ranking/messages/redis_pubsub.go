package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/XD/ScholarNet/cmd/ranking/repository"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// Redis通知频道
	LocalCacheNotifyChannel = "local_cache:notify"
)

// RedisNotifyMessage Redis通知消息结构
type RedisNotifyMessage struct {
	NodeID    string `json:"node_id"`   // 发送节点ID
	Timestamp int64  `json:"timestamp"` // 时间戳
}

// RedisPublisher Redis发布者
type RedisPublisher struct {
	client redis.UniversalClient
}

func NewRedisPublisher(client redis.UniversalClient) *RedisPublisher {
	return &RedisPublisher{
		client: client,
	}
}

// PublishUpdate 发布更新通知
func (r *RedisPublisher) PublishUpdate(ctx context.Context, nodeID string) error {
	msg := RedisNotifyMessage{
		NodeID:    nodeID,
		Timestamp: time.Now().Unix(),
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化通知消息失败: %w", err)
	}

	return r.client.Publish(ctx, LocalCacheNotifyChannel, msgBytes).Err()
}

// PublishDelete 发布删除通知
func (r *RedisPublisher) PublishDelete(ctx context.Context, nodeID string) error {
	msg := RedisNotifyMessage{
		NodeID:    nodeID,
		Timestamp: time.Now().Unix(),
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化通知消息失败: %w", err)
	}

	return r.client.Publish(ctx, LocalCacheNotifyChannel, msgBytes).Err()
}

// RedisSubscriber Redis订阅者
type RedisSubscriber struct {
	client     redis.UniversalClient
	subscriber *redis.PubSub
	repo       repository.RankingRepository
	nodeID     string
	lastUpdate time.Time
	cooldown   time.Duration // 冷却时间，防止频繁更新
}

func NewRedisSubscriber(client redis.UniversalClient, repo repository.RankingRepository, nodeID string) *RedisSubscriber {
	return &RedisSubscriber{
		client:   client,
		repo:     repo,
		nodeID:   nodeID,
		cooldown: 10 * time.Second, // 10秒冷却时间
	}
}

// Start 开始订阅
func (r *RedisSubscriber) Start(ctx context.Context) error {
	r.subscriber = r.client.Subscribe(ctx, LocalCacheNotifyChannel)

	// 启动消息处理协程
	go r.handleMessages(ctx)

	return nil
}

// Stop 停止订阅
func (r *RedisSubscriber) Stop() error {
	if r.subscriber != nil {
		return r.subscriber.Close()
	}
	return nil
}

// handleMessages 处理订阅消息
func (r *RedisSubscriber) handleMessages(ctx context.Context) {
	ch := r.subscriber.Channel()

	for {
		select {
		case msg := <-ch:
			r.processMessage(ctx, msg)
		case <-ctx.Done():
			return
		}
	}
}

// processMessage 处理单条消息
func (r *RedisSubscriber) processMessage(ctx context.Context, msg *redis.Message) {
	var notifyMsg RedisNotifyMessage
	if err := json.Unmarshal([]byte(msg.Payload), &notifyMsg); err != nil {
		return
	}

	// 忽略自己发送的消息
	if notifyMsg.NodeID == r.nodeID {
		return
	}

	// 检查冷却时间
	if notifyMsg.Timestamp-r.lastUpdate.Unix() < int64(r.cooldown) {
		return
	}

	if err := r.repo.RefreshLocalCacheV1(ctx); err == nil {
		r.lastUpdate = time.Now()
	}
}
