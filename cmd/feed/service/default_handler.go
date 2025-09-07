package service

import (
	"context"
	"github.com/XD/ScholarNet/cmd/feed/domain"
	"github.com/XD/ScholarNet/cmd/feed/repository"
)

type defaultHandler struct {
	repo repository.FeedEventRepo
}

// 默认用推模型
func (d *defaultHandler) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	uid := ext.Get("uid").Int64OrDefault(0)
	return d.repo.CreatePushEvents(ctx, []domain.FeedEvent{{
		Uid:  uid,
		Type: "type 要传下来",
		Ext:  ext,
	}})
}

func (d *defaultHandler) FindFeedEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	//TODO implement me
	panic("implement me")
}
