package events_test

import (
	"context"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/XD/ScholarNet/cmd/ranking/domain"
	"github.com/XD/ScholarNet/cmd/ranking/events"
	"github.com/XD/ScholarNet/cmd/ranking/repository/cache"
)

// 模拟Kafka客户端
type MockKafkaClient struct{}

func (m *MockKafkaClient) Config() *sarama.Config {
	return sarama.NewConfig()
}

func (m *MockKafkaClient) Brokers() []string {
	return []string{"localhost:9092"}
}

func (m *MockKafkaClient) Controller() (*sarama.Broker, error) {
	return nil, nil
}

func (m *MockKafkaClient) Topics() ([]string, error) {
	return []string{"local_cache_updates"}, nil
}

func (m *MockKafkaClient) Partitions(topic string) ([]int32, error) {
	return []int32{0}, nil
}

func (m *MockKafkaClient) WritablePartitions(topic string) ([]int32, error) {
	return []int32{0}, nil
}

func (m *MockKafkaClient) Leader(topic string, partitionID int32) (*sarama.Broker, error) {
	return nil, nil
}

func (m *MockKafkaClient) Replicas(topic string, partitionID int32) ([]int32, error) {
	return []int32{0}, nil
}

func (m *MockKafkaClient) InSyncReplicas(topic string, partitionID int32) ([]int32, error) {
	return []int32{0}, nil
}

func (m *MockKafkaClient) RefreshMetadata(topics ...string) error {
	return nil
}

func (m *MockKafkaClient) GetOffset(topic string, partitionID int32, time int64) (int64, error) {
	return 0, nil
}

func (m *MockKafkaClient) Coordinator(consumerGroup string) (*sarama.Broker, error) {
	return nil, nil
}

func (m *MockKafkaClient) RefreshCoordinator(consumerGroup string) error {
	return nil
}

func (m *MockKafkaClient) InitProducerID() (*sarama.InitProducerIDResponse, error) {
	return nil, nil
}

func (m *MockKafkaClient) Close() error { return nil }
func (m *MockKafkaClient) Closed() bool { return false }

// 模拟日志记录器
type MockLogger struct{}

func (m *MockLogger) Debug(msg string, args ...interface{}) {}
func (m *MockLogger) Info(msg string, args ...interface{})  {}
func (m *MockLogger) Warn(msg string, args ...interface{})  {}
func (m *MockLogger) Error(msg string, args ...interface{}) {}

// 测试Kafka消费者逻辑
func TestSaramaConsumer_Consume(t *testing.T) {
	// 创建模拟客户端和本地缓存
	kafkaClient := &MockKafkaClient{}
	localCache := cache.NewRankingLocalCache()
	logger := &MockLogger{}

	// 创建消费者
	consumer := events.NewArticleEventConsumer(kafkaClient, logger, localCache, "node_001")

	// 测试数据
	testArticles := []domain.Article{
		{Id: 1, Title: "测试文章1", Content: "内容1", Score: 100},
		{Id: 2, Title: "测试文章2", Content: "内容2", Score: 90},
	}

	// 测试1：自己发送的消息应该被跳过
	t.Run("跳过自己发送的消息", func(t *testing.T) {
		msg := events.LocalCacheUpdateMessage{
			NodeID:    "node_001", // 自己的节点ID
			Timestamp: time.Now().Unix(),
			Articles:  testArticles,
		}

		err := consumer.Consume(nil, msg)
		if err != nil {
			t.Errorf("应该成功跳过自己发送的消息，但得到错误: %v", err)
		}

		// 验证本地缓存没有被更新
		articles, err := localCache.Get(context.Background())
		if err == nil {
			t.Errorf("本地缓存不应该有数据，但得到了%d篇文章", len(articles))
		}
	})

	// 测试2：其他节点发送的消息应该被处理
	t.Run("处理其他节点的消息", func(t *testing.T) {
		msg := events.LocalCacheUpdateMessage{
			NodeID:    "node_002", // 其他节点的ID
			Timestamp: time.Now().Unix(),
			Articles:  testArticles,
		}

		err := consumer.Consume(nil, msg)
		if err != nil {
			t.Errorf("应该成功处理其他节点的消息，但得到错误: %v", err)
		}

		// 验证本地缓存被更新
		articles, err := localCache.Get(context.Background())
		if err != nil {
			t.Errorf("本地缓存应该有数据，但得到错误: %v", err)
		}
		if len(articles) != 2 {
			t.Errorf("期望2篇文章，实际得到%d篇", len(articles))
		}
	})

	// 测试3：冷却时间内的消息应该被跳过
	t.Run("跳过冷却时间内的消息", func(t *testing.T) {
		// 立即发送另一条消息
		msg := events.LocalCacheUpdateMessage{
			NodeID:    "node_003",
			Timestamp: time.Now().Unix(),
			Articles:  testArticles,
		}

		err := consumer.Consume(nil, msg)
		if err != nil {
			t.Errorf("应该成功跳过冷却时间内的消息，但得到错误: %v", err)
		}

		// 验证本地缓存没有被重复更新（文章数量应该还是2）
		articles, err := localCache.Get(context.Background())
		if err != nil {
			t.Errorf("本地缓存应该有数据，但得到错误: %v", err)
		}
		if len(articles) != 2 {
			t.Errorf("期望2篇文章，实际得到%d篇", len(articles))
		}
	})
}
