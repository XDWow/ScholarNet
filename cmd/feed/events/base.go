package events

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/XD/ScholarNet/cmd/feed/domain"
	"github.com/XD/ScholarNet/cmd/feed/service"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"github.com/XD/ScholarNet/cmd/pkg/saramax"
	"strconv"
	"time"
)

const topicFeedEvent = "feed_event"

// FeedEvent 为什么你一开始就确定用消息队列
// 实时性要求不高，肯定用消息队列异步啊，好处多多：削峰、异步、解耦
type FeedEvent struct {
	// Type 是我内部定义，我发给不同业务方
	// like_event
	// like_event_v1
	Type string

	// 加一个 version
	// Version int

	// 你业务方的具体的数据
	// 点赞需要三个 key:
	// liker
	// liked
	// biz + bizId
	Metadata map[string]string
	//Metadata string
	//Metadata []byte
}

type FeedEventConsumer struct {
	client sarama.Client
	l      logger.LoggerV1
	svc    service.FeedService
}

func (f *FeedEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("feed_event", f.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{topicFeedEvent},
			saramax.NewHandler[FeedEvent](f.l, f.Consume))
		if err != nil {
			f.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (f *FeedEventConsumer) Consume(msg *sarama.ConsumerMessage, event FeedEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 我要不要在这里提前校验一下 type 是不是合法的值？
	// 我作为一个懒人，我是不校验的
	// 一旦你校验了，业务方接入的时候就要改代码
	return f.svc.CreateFeedEvent(ctx, domain.FeedEvent{
		Type: event.Type,
		Ext:  event.Metadata,
	})
}

func NewFeedEventConsumer(
	client sarama.Client,
	l logger.LoggerV1,
	svc service.FeedService) *FeedEventConsumer {
	return &FeedEventConsumer{
		svc:    svc,
		client: client,
		l:      l,
	}
}

// LikeFeedEvent 能不能？
// 你就得改代码了
type LikeFeedEvent struct {
	Liker int64
}

func (r *FeedEventConsumer) ConsumeV1(msg *sarama.ConsumerMessage,
	evt LikeFeedEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return r.svc.CreateFeedEvent(ctx, domain.FeedEvent{
		Type: "like_event",
		Ext: map[string]string{
			"liker": strconv.FormatInt(evt.Liker, 10),
			//
		},
	})
}
