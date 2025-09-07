package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/XD/ScholarNet/cmd/ranking/domain"
	"github.com/XD/ScholarNet/cmd/ranking/repository/cache"
	"github.com/redis/go-redis/v9"
)

// 示例：在实际项目中使用本地缓存一致性方案

func main() {
	ctx := context.Background()

	// 1. 初始化Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer redisClient.Close()

	// 2. 初始化Kafka客户端
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Consumer.Return.Errors = true

	kafkaClient, err := sarama.NewClient([]string{"localhost:9092"}, kafkaConfig)
	if err != nil {
		log.Fatalf("创建Kafka客户端失败: %v", err)
	}
	defer kafkaClient.Close()

	// 3. 创建本地缓存工厂
	factory := cache.NewLocalCacheFactory(cache.FactoryConfig{
		RedisClient: redisClient,
		KafkaClient: kafkaClient,
		Logger:      &SimpleLogger{},
	})

	// 4. 选择一致性策略
	strategy := cache.StrategyRedisPubSub // 或者 cache.StrategyKafka

	// 5. 创建本地缓存实例
	cacheManager, err := factory.CreateLocalCache(
		ctx,
		strategy,
		"ranking_node_001",
		cache.DefaultLocalCacheConfig(),
	)
	if err != nil {
		log.Fatalf("创建本地缓存失败: %v", err)
	}
	defer cacheManager.Stop()

	// 6. 模拟定时任务更新排行榜数据
	go simulateRankingUpdate(ctx, cacheManager)

	// 7. 模拟多个客户端读取排行榜
	for i := 0; i < 3; i++ {
		go simulateClientRead(ctx, cacheManager, fmt.Sprintf("client_%d", i))
	}

	// 8. 运行一段时间
	time.Sleep(30 * time.Second)
	fmt.Println("示例运行完成")
}

// 模拟定时任务更新排行榜
func simulateRankingUpdate(ctx context.Context, cacheManager *cache.CacheManager) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	articleID := int64(1)

	for {
		select {
		case <-ticker.C:
			// 模拟新的排行榜数据
			articles := []domain.Article{
				{
					Id:      articleID,
					Title:   fmt.Sprintf("热门文章_%d", articleID),
					Content: fmt.Sprintf("这是第%d篇热门文章的内容", articleID),
					Score:   1000 - int64(articleID*10),
				},
				{
					Id:      articleID + 1,
					Title:   fmt.Sprintf("热门文章_%d", articleID+1),
					Content: fmt.Sprintf("这是第%d篇热门文章的内容", articleID+1),
					Score:   1000 - int64((articleID+1)*10),
				},
				{
					Id:      articleID + 2,
					Title:   fmt.Sprintf("热门文章_%d", articleID+2),
					Content: fmt.Sprintf("这是第%d篇热门文章的内容", articleID+2),
					Score:   1000 - int64((articleID+2)*10),
				},
			}

			// 更新本地缓存
			err := cacheManager.Set(ctx, articles)
			if err != nil {
				log.Printf("更新本地缓存失败: %v", err)
			} else {
				log.Printf("成功更新本地缓存，文章ID: %d-%d", articleID, articleID+2)
			}

			articleID += 3
		case <-ctx.Done():
			return
		}
	}
}

// 模拟客户端读取排行榜
func simulateClientRead(ctx context.Context, cacheManager *cache.CacheManager, clientName string) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 从本地缓存读取数据
			articles, err := cacheManager.Get(ctx)
			if err != nil {
				log.Printf("[%s] 读取本地缓存失败: %v", clientName, err)
				continue
			}

			log.Printf("[%s] 成功读取%d篇文章", clientName, len(articles))
			for i, article := range articles {
				if i >= 3 { // 只显示前3篇
					break
				}
				log.Printf("[%s] 第%d名: %s (ID: %d, 分数: %d)",
					clientName, i+1, article.Title, article.Id, article.Score)
			}
		case <-ctx.Done():
			return
		}
	}
}

// 简单日志记录器
type SimpleLogger struct{}

func (l *SimpleLogger) Debug(msg string, args ...interface{}) {
	log.Printf("[DEBUG] "+msg, args...)
}

func (l *SimpleLogger) Info(msg string, args ...interface{}) {
	log.Printf("[INFO] "+msg, args...)
}

func (l *SimpleLogger) Warn(msg string, args ...interface{}) {
	log.Printf("[WARN] "+msg, args...)
}

func (l *SimpleLogger) Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] "+msg, args...)
}

// 示例：比较两种策略的性能
func compareStrategies() {
	ctx := context.Background()

	// 初始化客户端（这里使用模拟客户端）
	redisClient := &MockRedisClient{}
	kafkaClient := &MockKafkaClient{}

	factory := cache.NewLocalCacheFactory(cache.FactoryConfig{
		RedisClient: redisClient,
		KafkaClient: kafkaClient,
		Logger:      &SimpleLogger{},
	})

	// 创建两种策略的缓存
	caches, err := factory.CreateMultipleCaches(ctx, "compare_node", cache.DefaultLocalCacheConfig())
	if err != nil {
		log.Fatalf("创建缓存失败: %v", err)
	}
	defer factory.StopAllCaches(caches)

	// 测试数据
	testData := []domain.Article{
		{Id: 1, Title: "测试文章1", Score: 100},
		{Id: 2, Title: "测试文章2", Score: 90},
		{Id: 3, Title: "测试文章3", Score: 80},
	}

	// 测试Redis发布订阅方案
	start := time.Now()
	redisCache := caches[cache.StrategyRedisPubSub]
	for i := 0; i < 1000; i++ {
		redisCache.Set(ctx, testData)
		redisCache.Get(ctx)
	}
	redisDuration := time.Since(start)

	// 测试Kafka消息队列方案
	start = time.Now()
	kafkaCache := caches[cache.StrategyKafka]
	for i := 0; i < 1000; i++ {
		kafkaCache.Set(ctx, testData)
		kafkaCache.Get(ctx)
	}
	kafkaDuration := time.Since(start)

	fmt.Printf("Redis发布订阅方案耗时: %v\n", redisDuration)
	fmt.Printf("Kafka消息队列方案耗时: %v\n", kafkaDuration)
}

// 模拟客户端（用于测试）
type MockRedisClient struct{}
type MockKafkaClient struct{}

func (m *MockRedisClient) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	return redis.NewIntCmd(ctx, 1)
}

func (m *MockRedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return &redis.PubSub{}
}

func (m *MockRedisClient) ZRevRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return redis.NewStringSliceCmd(ctx, "{}")
}

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

func (m *MockKafkaClient) Close() error {
	return nil
}

func (m *MockKafkaClient) Closed() bool {
	return false
}
