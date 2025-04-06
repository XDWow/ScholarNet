package saramax

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/LXD-c/basic-go/webook/pkg/logger"
	"time"
)

type BatchHandler[T any] struct {
	l  logger.LoggerV1
	fn func(msgs []*sarama.ConsumerMessage, ts []T) error
	// 用 option 模式来设置这个 batchSize 和 duration
	batchSize     int
	batchDuration time.Duration
}

func NewBatchHandler[T any](l logger.LoggerV1, fn func(msgs []*sarama.ConsumerMessage, ts []T) error) *BatchHandler[T] {
	return &BatchHandler[T]{l: l, fn: fn, batchDuration: time.Second, batchSize: 10}
}

func (b *BatchHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// 拿消息给你写好了, 提交也帮你写好了, 都是通用的
// 只需传入你的消费信息业务 fn()
func (b *BatchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgsCh := claim.Messages()
	batchSize := b.batchSize
	for {
		// 批量消息数据，要有个超时 context ，避免一直等待凑齐一批消息
		ctx, cancel := context.WithTimeout(context.Background(), b.batchDuration)
		// 未超时
		done := false
		// 初始化消息切片
		msgs := make([]*sarama.ConsumerMessage, 0, batchSize)
		ts := make([]T, 0, batchSize)
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				done = true
			case msg, ok := <-msgsCh:
				// 通道关闭了
				if !ok {
					cancel()
					return nil
				}
				var t T
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					b.l.Error("反序列化失败",
						logger.Error(err),
						logger.String("topic", msg.Topic),
						logger.Int64("partition", int64(msg.Partition)),
						logger.Int64("offset", msg.Offset))
					// 后面不执行了，跳到下一条消息
					continue
				}
				msgs = append(msgs, msg)
				ts = append(ts, t)
			}
		}
		// 一批数据拿完了
		// 批量消费
		cancel()
		// 一个消息都没拿到，不能执行消耗fn,继续循环等消息吧
		if len(msgs) == 0 {
			continue
		}
		err := b.fn(msgs, ts)
		if err != nil {
			b.l.Error("调用业务批量接口失败",
				logger.Error(err))
			// 你这里整个批次都要记下来

			// 还要继续往前消费
		}
		for _, msg := range msgs {
			// 这样，万无一失
			session.MarkMessage(msg, "")
		}
	}
}
