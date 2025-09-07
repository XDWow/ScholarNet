package ioc

import (
	"crypto/md5"
	"fmt"
	"os"
	"time"

	"github.com/IBM/sarama"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"github.com/XD/ScholarNet/cmd/ranking/events"
	"github.com/XD/ScholarNet/cmd/ranking/repository/cache"
	"github.com/spf13/viper"
)

func InitKafkaClient() sarama.Client {
	type Config struct {
		Addrs []string `yaml:"addrs"`
	}
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(fmt.Errorf("初始化 Kafka 配置失败: %w", err))
	}

	scfg := sarama.NewConfig()
	scfg.Producer.Return.Successes = true

	// 配置哈希分区器，确保相同key的消息发送到同一分区，保证顺序性
	scfg.Producer.Partitioner = sarama.NewHashPartitioner

	// 设置重试和超时
	scfg.Producer.Retry.Max = 3
	scfg.Producer.Timeout = 10 * time.Second

	// 消费者配置
	scfg.Consumer.Return.Errors = true
	scfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	client, err := sarama.NewClient(cfg.Addrs, scfg)
	if err != nil {
		panic(fmt.Errorf("初始化 Kafka 客户端失败: %w", err))
	}
	return client
}

func InitKafkaProducer(client sarama.Client) events.Producer {
	nodeID := generateNodeID()
	producer, err := events.NewSaramaProducer(client, nodeID)
	if err != nil {
		panic(fmt.Errorf("初始化 Kafka 生产者失败: %w", err))
	}
	return producer
}

func InitKafkaConsumer(
	client sarama.Client,
	l logger.LoggerV1,
	localCache events.LocalCache,
) events.Consumer {
	nodeID := generateNodeID()
	consumer := events.NewArticleEventConsumer(client, l, localCache, nodeID)
	return consumer
}

// 供依赖注入使用：构建带 producer 和 logger 的本地缓存
func InitRankingLocalCache(p events.Producer, l logger.LoggerV1) *cache.RankingLocalCache {
	return cache.NewRankingLocalCache(p, l)
}

// generateNodeID 生成唯一的节点ID
func generateNodeID() string {
	// 优先使用配置文件中的节点ID
	if nodeID := viper.GetString("node.id"); nodeID != "" {
		return nodeID
	}

	// 尝试使用环境变量
	if nodeID := os.Getenv("NODE_ID"); nodeID != "" {
		return nodeID
	}

	// 最后使用主机名+进程ID生成
	hostname, _ := os.Hostname()
	pid := os.Getpid()

	// 使用MD5生成短一点的ID
	data := fmt.Sprintf("%s-%d-%d", hostname, pid, time.Now().UnixNano())
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("node-%x", hash[:8])
}
