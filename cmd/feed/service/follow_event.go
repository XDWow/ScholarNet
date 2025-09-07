package service

import (
	"context"
	"github.com/XD/ScholarNet/cmd/feed/domain"
	"github.com/XD/ScholarNet/cmd/feed/repository"
)

const (
	FollowEventName = "follow_event"
)

type FollowEventHandler struct {
	repo repository.FeedEventRepo
}

func (f *FollowEventHandler) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	followee, err := ext.Get("followee").AsInt64()
	if err != nil {
		return err
	}
	return f.repo.CreatePushEvents(ctx, []domain.FeedEvent{
		{
			Uid:  followee,
			Type: FollowEventName,
			Ext:  ext,
		},
	})
}

func (f *FollowEventHandler) FindFeedEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	return f.repo.FindPushEventsWithTyp(ctx, FollowEventName, uid, timestamp, limit)
}

func NewFollowEventHandler(repo repository.FeedEventRepo) *FollowEventHandler {
	return &FollowEventHandler{repo: repo}
}
