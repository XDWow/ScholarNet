package service

import (
	"context"
	"github.com/XD/ScholarNet/cmd/feed/domain"
)

// 组合v1,只需重写发生了变化的方法
type LikeEventHandlerV2 struct {
	LikeEventHandler
}

// CreateFeedEvent 重写发生了变化的方法
func (l *LikeEventHandlerV2) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	// 新版逻辑

	return nil
}
