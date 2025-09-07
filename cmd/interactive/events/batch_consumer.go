package events

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/XD/ScholarNet/cmd/interactive/repository"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"github.com/XD/ScholarNet/cmd/pkg/saramax"
)

type InteractiveReadEventBatchConsumer struct {
	client    sarama.Client
	repo      repository.InteractiveRepository
	l         logger.LoggerV1
	batchsize int
}

func NewInteractiveReadEventBatchConsumer(client sarama.Client,
	repo repository.InteractiveRepository,
	l logger.LoggerV1, batchsize int) *InteractiveReadEventBatchConsumer {
	return &InteractiveReadEventBatchConsumer{client: client, repo: repo, l: l, batchsize: batchsize}
}

func (r *InteractiveReadEventBatchConsumer) Start() error {
	// 创建了消费者组 interactive 的一个消费者实列
	cg, err := sarama.NewConsumerGroupFromClient("interactive", r.client)
	if err != nil {
		return err
	}
	// 开一个线程，使用 Consume 启动一个消费者
	go func() {
		// 启动消费，指定 topic，一个消费者可以消费多个 topic，所以是 []string ：切片
		// 分区分配：这个 cg.Consume 会向 Kafka 的 Group Coordinator 报到，Kafka 会根据 Group 内活跃的消费者实例数量，把分区分配下去。
		// 后续有新的消费者加入，Kafka 会触发 Rebalance，重新分配分区。
		// 拿到分区列表，Sarama 调用 Setup(session)初始化，然后为每个分区启动一个 goroutine，调用 ConsumerClaim，
		// 会话结束时（rebalacne或者退出),Sarama 会调用 Cleanup(session) 释放资源
		err := cg.Consume(context.Background(),
			[]string{"read_article"},
			saramax.NewAsyncBatchHandlerSimple[ReadEvent](r.l, r.Consume, r.batchsize))
		if err != nil {
			r.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return nil
}

func (r *InteractiveReadEventBatchConsumer) Consume(ctx context.Context, msgs []*sarama.ConsumerMessage, ts []ReadEvent) error {
	ids := make([]int64, 0, len(ts))
	bizs := make([]string, 0, len(ts))
	// 第一个参数是索引
	for _, evt := range ts {
		ids = append(ids, evt.Aid)
		bizs = append(bizs, "article")
	}
	c, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	err := r.repo.BatchIncrReadCnt(c, bizs, ids)
	if err != nil {
		r.l.Error("批量增加阅读计数失败",
			logger.Field{Key: "ids", Value: ids},
			logger.Error(err))
	} else {
		r.l.Info("消费成功")
	}
	return nil
}
