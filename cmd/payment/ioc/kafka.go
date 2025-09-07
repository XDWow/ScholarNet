package ioc

import (
	"github.com/IBM/sarama"
	events2 "github.com/XD/ScholarNet/cmd/ranking/events/events"
	"github.com/spf13/viper"
)

func InitKafka() sarama.Client {
	type Config struct {
		Addrs []string `yaml:"addrs"`
	}
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	saramaCfg.Producer.Partitioner = sarama.NewConsistentCRCHashPartitioner
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

func InitProducer(client sarama.Client) events2.Producer {
	res, err := events2.NewSaramaProducer(client)
	if err != nil {
		panic(err)
	}
	return res
}
