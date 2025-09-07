package service

import (
	"context"
	"github.com/XD/ScholarNet/cmd/feed/domain"
)

// FeedService 处理业务公共的部分
// 并且负责找出 Handler 来处理业务的个性部分
type FeedService interface {
	CreateFeedEvent(ctx context.Context, feed domain.FeedEvent) error
	GetFeedEventList(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error)
}

// Handler 具体业务处理逻辑
// 按照 type 来分。因为 type 是天然标记了哪个业务
type Handler interface {
	CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error
	FindFeedEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error)
}
