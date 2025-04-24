package ioc

import (
	"github.com/IBM/sarama"
	events2 "github.com/LXD-c/basic-go/webook/interactive/events"
	"github.com/LXD-c/basic-go/webook/interactive/repository/dao"
	"github.com/LXD-c/basic-go/webook/pkg/migrator/events/fixer"
	"github.com/LXD-c/basic-go/webook/pkg/saramax"
	"github.com/spf13/viper"
)

func InitKafka() sarama.Client {
	type Config struct {
		Addrs []string `yaml:"addrs"`
	}
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	client, err := sarama.NewClient(cfg.Addrs, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

func InitSyncProducer(client sarama.Client) sarama.SyncProducer {
	res, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return res
}

// NewConsumers 面临的问题依旧是所有的 Consumer 在这里注册一下
// 加上fix.Consumer
func NewConsumers(c1 *events2.InteractiveReadEventConsumer,
	fix *fixer.Consumer[dao.Interactive]) []saramax.Consumer {
	return []saramax.Consumer{c1, fix}
}
