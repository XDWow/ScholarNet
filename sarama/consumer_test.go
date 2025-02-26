package sarama

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"log"
	"testing"
	"time"
)

func TestConsumer(t *testing.T) {
	cfg := sarama.NewConfig()
	// 正常来说，一个消费者都是归属于一个消费者的组的
	// 消费者组就是你的业务
	consumer, err := sarama.NewConsumerGroup(addrs, "test_group", cfg)
	require.NoError(t, err)

	// 带超时的 context
	start := time.Now()
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(time.Minute*10, func() {
		cancel()
	})
	err = consumer.Consume(ctx, []string{"test_topic"}, testConsumerGroupHandler{})
	// 你消费结束，就会到这里
	t.Log(err, time.Since(start).String())
}

type testConsumerGroupHandler struct{}

func (h testConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	// 执行一些初始化的事情
	partitions := session.Claims()["test_topic"]
	// 遍历所有的分区，每一个分区都有对应的偏移量，都修改
	for _, part := range partitions {
		session.ResetOffset("test_topic", part, sarama.OffsetOldest, "")
		//session.ResetOffset("test_topic", part,
		//	123, "")
		//session.ResetOffset("test_topic", part,
		//	sarama.OffsetNewest, "")
	}
	return nil
}

func (h testConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("Closing session:", session)
	return nil
}

func (t testConsumerGroupHandler) ConsumeClaim(
	// 代表的是你和Kafka 的会话（从建立连接到连接彻底断掉的那一段时间）
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	//for msg := range msgs {
	//	m1 := msg
	//	go func() {
	//		// 消费msg
	//		log.Println(string(m1.Value))
	//		session.MarkMessage(m1, "")
	//	}()
	//}
	const batchsize = 10
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		var eg errgroup.Group
		var last *sarama.ConsumerMessage
		done := false
		for i := 0; i < batchsize && !done; i++ {
			select {
			case msg, ok := <-msgs:
				if !ok {
					// 代表消费者被关闭了
					cancel()
					return nil
				}
				last = msg
				eg.Go(func() error {
					// 异步消费
					time.Sleep(time.Second)
					// 失败了在这里重试
					log.Println(string(msg.Value))
					return nil
				})
			case <-ctx.Done():
				// 超时了，在这里 break 没用
				done = true
			}
		}
		cancel()
		// 要等异步消费完，再批量提交
		err := eg.Wait()
		if err != nil {
			continue
		}
		// 就这样
		if last != nil {
			session.MarkMessage(last, "")
		}
	}
}

func (t testConsumerGroupHandler) ConsumeClaimV1(
	// 代表的是你和Kafka 的会话（从建立连接到连接彻底断掉的那一段时间）
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		//var bizMsg MyBizMsg
		//err := json.Unmarshal(msg.Value, &bizMsg)
		//if err != nil {
		//	// 这就是消费消息出错
		//	// 大多数时候就是重试
		//	// 记录日志
		//	continue
		//}
		log.Println(string(msg.Value))
		session.MarkMessage(msg, "")
	}
	// 什么情况下会到这里
	// msgs 被人关了，也就是要退出消费逻辑
	return nil
}
