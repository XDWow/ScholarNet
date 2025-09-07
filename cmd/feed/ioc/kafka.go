package ioc

import (
	"github.com/IBM/sarama"
	"github.com/XD/ScholarNet/cmd/feed/events"
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

// 组装 []consumer
func NewConsumers(article *events.ArticleEventConsumer, feed *events.FeedEventConsumer) []events.Consumer {
	return []events.Consumer{
		article,
		feed,
	}
}
