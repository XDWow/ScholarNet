package events

import (
	"encoding/json"
	"github.com/IBM/sarama"
)

const topicReadEvent = "article_read_event"

type ReadEvent struct {
	Aid int64
	Uid int64
}

type Producer interface {
	ProduceReadEvent(evt ReadEvent) error
}

type SaramaSyncProducer struct {
	producer sarama.SyncProducer
}

func NewSaramaSyncProducer(producer sarama.SyncProducer) Producer {
	return &SaramaSyncProducer{
		producer: producer,
	}
}

func (s *SaramaSyncProducer) ProduceReadEvent(evt ReadEvent) error {
	val, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = s.producer.
		SendMessage(&sarama.ProducerMessage{
			Topic: topicReadEvent,
			Value: sarama.ByteEncoder(val),
		})
	return err
}
