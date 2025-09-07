package repository

import (
	"context"
	"encoding/json"
	"github.com/XD/ScholarNet/cmd/feed/domain"
	"github.com/XD/ScholarNet/cmd/feed/repository/dao"
	"github.com/ecodeclub/ekit/slice"
	"time"
)

type FeedEventRepo interface {
	// CreatePushEvents 批量推事件
	CreatePushEvents(ctx context.Context, events []domain.FeedEvent) error
	// CreatePullEvent 创建拉事件
	CreatePullEvent(ctx context.Context, event domain.FeedEvent) error
	FindPullEvents(ctx context.Context, uids []int64, timestamp, limit int64) ([]domain.FeedEvent, error)
	// FindPushEvents 获取推事件，也就是自己收件箱里面的事件
	FindPushEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error)
	// FindPullEventsWithTyp 获取某个类型的拉事件，
	FindPullEventsWithTyp(ctx context.Context, typ string, uids []int64, timestamp, limit int64) ([]domain.FeedEvent, error)
	// FindPushEvents 获取某个类型的推事件，也就是发件箱里面的事件
	FindPushEventsWithTyp(ctx context.Context, typ string, uid, timestamp, limit int64) ([]domain.FeedEvent, error)
}

type feedEventRepo struct {
	pullDao dao.FeedPullEventDAO
	pushDao dao.FeedPushEventDAO
}

func (f *feedEventRepo) FindPushEventsWithTyp(ctx context.Context,
	typ string, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	events, err := f.pushDao.GetPushEventsWithTyp(ctx, typ, uid, timestamp, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map(events, func(idx int, src dao.FeedPushEvent) domain.FeedEvent {
		return convertToPushEventDomain(src)
	}), nil
}

func (f *feedEventRepo) FindPullEventsWithTyp(ctx context.Context,
	typ string, uids []int64, timestamp, limit int64) ([]domain.FeedEvent, error) {
	events, err := f.pullDao.FindPullEventListWithType(ctx, typ, uids, timestamp, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map(events, func(idx int, src dao.FeedPullEvent) domain.FeedEvent {
		return convertToPullEventDomain(src)
	}), nil
}

func (f *feedEventRepo) CreatePushEvents(ctx context.Context, events []domain.FeedEvent) error {
	return f.pushDao.CreatePushEvents(ctx, slice.Map(events, func(idx int, src domain.FeedEvent) dao.FeedPushEvent {
		return convertToPushEvent(src)
	}))
}

func (f *feedEventRepo) CreatePullEvent(ctx context.Context, event domain.FeedEvent) error {
	return f.pullDao.CreatePullEvent(ctx, convertToPullEvent(event))
}

func (f *feedEventRepo) FindPullEvents(ctx context.Context, uids []int64, timestamp, limit int64) ([]domain.FeedEvent, error) {
	events, err := f.pullDao.FindPullEventList(ctx, uids, timestamp, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map(events, func(idx int, src dao.FeedPullEvent) domain.FeedEvent {
		return convertToPullEventDomain(src)
	}), nil
}

func (f *feedEventRepo) FindPushEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	events, err := f.pushDao.GetPushEvents(ctx, uid, timestamp, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map(events, func(idx int, src dao.FeedPushEvent) domain.FeedEvent {
		return convertToPushEventDomain(src)
	}), nil
}

func convertToPushEvent(event domain.FeedEvent) dao.FeedPushEvent {
	val, _ := json.Marshal(event.Ext)
	return dao.FeedPushEvent{
		UID:     event.Uid,
		Type:    event.Type,
		Content: string(val),
	}
}

func convertToPullEvent(event domain.FeedEvent) dao.FeedPullEvent {
	val, _ := json.Marshal(event.Ext)
	return dao.FeedPullEvent{
		UID:     event.Uid,
		Type:    event.Type,
		Content: string(val),
	}
}

func convertToPushEventDomain(event dao.FeedPushEvent) domain.FeedEvent {
	var ext map[string]string
	_ = json.Unmarshal([]byte(event.Content), &ext)
	return domain.FeedEvent{
		ID:    event.Id,
		Uid:   event.UID,
		Type:  event.Type,
		Ctime: time.Unix(event.Ctime, 0),
		Ext:   ext,
	}
}

func convertToPullEventDomain(event dao.FeedPullEvent) domain.FeedEvent {
	var ext map[string]string
	_ = json.Unmarshal([]byte(event.Content), &ext)
	return domain.FeedEvent{
		ID:    event.Id,
		Uid:   event.UID,
		Type:  event.Type,
		Ctime: time.Unix(event.Ctime, 0),
		Ext:   ext,
	}
}
