package events

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
)

const topicUpdateEvent = "article_localCacheUpdate_event"

type SaramaProducer struct {
	producer sarama.SyncProducer
	topic    string
	nodeID   string
}

func NewSaramaProducer(client sarama.Client, nodeID string) (*SaramaProducer, error) {
	p, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}
	return &SaramaProducer{
		producer: p,
		topic:    topicUpdateEvent,
		nodeID:   nodeID,
	}, nil
}

func (s *SaramaProducer) ProduceUpdateEvent(ctx context.Context, evt LocalCacheUpdateMessage) error {
	evt.NodeID = s.nodeID
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	// 使用固定的key确保所有消息都发送到同一个分区，保证顺序性
	// 这里使用固定的"ranking"作为key，所有排行榜更新消息都会在同一分区
	msg := &sarama.ProducerMessage{
		Topic: s.topic,
		Key:   sarama.StringEncoder("ranking"), // 固定key保证分区一致性和顺序性
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = s.producer.SendMessage(msg)
	return err
}
