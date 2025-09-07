package events

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"github.com/XD/ScholarNet/cmd/pkg/saramax"
	"github.com/XD/ScholarNet/cmd/search/service"
	"time"
)

type AnyEvent struct {
	IndexName string
	DocID     string
	Data      string
	// 假如说用于同步 user
	// IndexName = user_index
	// DocID = "123"
	// Data = {"id": 123, "email":xx, nickname: ""}
}

func NewAnyConsumer(client sarama.Client,
	l logger.LoggerV1,
	svc service.SyncService) *AnyConsumer {
	return &AnyConsumer{
		svc:    svc,
		client: client,
		l:      l,
	}
}

type AnyConsumer struct {
	svc    service.SyncService
	client sarama.Client
	l      logger.LoggerV1
}

func (a *AnyConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("search_sync_data",
		a.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{topicSyncArticle},
			saramax.NewHandler[AnyEvent](a.l, a.Consume))
		if err != nil {
			a.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (a *AnyConsumer) Consume(sg *sarama.ConsumerMessage,
	evt AnyEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 在这里执行转化
	return a.svc.InputAny(ctx, evt.IndexName, evt.DocID, evt.Data)
}
