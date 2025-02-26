package article

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/LXD-c/basic-go/webook/internal/repository"
	"github.com/LXD-c/basic-go/webook/pkg/logger"
	"github.com/LXD-c/basic-go/webook/pkg/saramax"
	"time"
)

type InteractiveReadEventBatchConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.LoggerV1
}

func NewInteractiveReadEventBatchConsumer(client sarama.Client, repo repository.InteractiveRepository, l logger.LoggerV1) *InteractiveReadEventBatchConsumer {
	return &InteractiveReadEventBatchConsumer{client: client, repo: repo, l: l}
}

func (r *InteractiveReadEventBatchConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", r.client)
	if err != nil {
		return err
	}
	// 开一个线程专门守着消费信息：for ()
	go func() {
		err := cg.Consume(context.Background(),
			[]string{"read_article"},
			saramax.NewBatchHandler[ReadEvent](r.l, r.Consume))
		if err != nil {
			r.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (r *InteractiveReadEventBatchConsumer) Consume(msgs []*sarama.ConsumerMessage, ts []ReadEvent) error {
	ids := make([]int64, 0, len(ts))
	bizs := make([]string, 0, len(ts))
	// 第一个参数是索引
	for _, evt := range ts {
		ids = append(ids, evt.Aid)
		bizs = append(bizs, "article")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := r.repo.BatchIncrReadCnt(ctx, bizs, ids)
	if err != nil {
		r.l.Error("批量增加阅读计数失败",
			logger.Field{Key: "ids", Value: ids},
			logger.Error(err))
	} else {
		r.l.Info("消费成功")
	}
	return nil
}
