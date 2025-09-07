package service

import (
	"context"
	followv1 "github.com/XD/ScholarNet/cmd/api/proto/gen/follow/v1"
	"github.com/XD/ScholarNet/cmd/feed/domain"
	"github.com/XD/ScholarNet/cmd/feed/repository"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"sort"
	"sync"
)

const (
	ArticleEventName = "article_event"
	// 你可以调大或者调小
	// 调大，数据量大，但是用户体验好
	// 调小，数据量小，但是用户体验差
	threshold = 32
)

type ArticleEventHandler struct {
	repo         repository.FeedEventRepo
	followClient followv1.FollowServiceClient
}

func NewArticleEventHandler(repo repository.FeedEventRepo, followClient followv1.FollowServiceClient) Handler {
	return &ArticleEventHandler{repo: repo, followClient: followClient}
}

func (a *ArticleEventHandler) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	authorId, err := ext.Get("uid").AsInt64()
	if err != nil {
		return err
	}
	resp, err := a.followClient.GetFollowStatics(ctx, &followv1.GetFollowStaticsRequest{
		Uid: authorId,
	})
	if err != nil {
		return err
	}
	// 根据粉丝数量决定使用推/拉模型
	// 粉丝不多，推模型，写扩散，收件箱
	if resp.GetFollowers() < threshold {
		followers, err := a.followClient.GetFollower(ctx, &followv1.GetFollowerRequest{
			Followee: authorId,
		})
		if err != nil {
			return err
		}
		// 要综合考虑什么活跃用户，是不是铁粉，
		// 在这里判定
		events := slice.Map(followers.GetFollowRelations(),
			func(idx int, src *followv1.FollowRelation) domain.FeedEvent {
				return domain.FeedEvent{
					Uid:  src.Follower,
					Type: ArticleEventName,
					Ext:  ext,
				}
			})
		return a.repo.CreatePushEvents(ctx, events)
	} else {
		return a.repo.CreatePullEvent(ctx, domain.FeedEvent{
			Uid:  authorId,
			Type: ArticleEventName,
			Ext:  ext,
		})
	}
}

func (a *ArticleEventHandler) FindFeedEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	// 我关注的人，可能使用推模型，也可能使用拉模型，所以全都得查
	var (
		eg errgroup.Group
		mu sync.Mutex
	)
	events := make([]domain.FeedEvent, 0, limit*2)
	// Push Event
	eg.Go(func() error {
		pushEvents, err := a.repo.FindPushEventsWithTyp(ctx, ArticleEventName, uid, timestamp, limit)
		if err != nil {
			return err
		}
		mu.Lock()
		events = append(events, pushEvents...)
		mu.Unlock()
		return nil
	})

	// Pull Event
	eg.Go(func() error {
		// 首先要查，我关注了哪些人，拿到他们的 id
		resp, rerr := a.followClient.GetFollowee(ctx, &followv1.GetFolloweeRequest{
			Follower: uid,
			Offset:   0,
			Limit:    200,
		})
		if rerr != nil {
			return rerr
		}
		followeeIds := slice.Map(resp.FollowRelations, func(idx int, src *followv1.FollowRelation) int64 {
			return src.Followee
		})
		pullEvents, err := a.repo.FindPullEventsWithTyp(ctx, ArticleEventName, followeeIds, timestamp, limit)
		if err != nil {
			return err
		}
		mu.Lock()
		events = append(events, pullEvents...)
		mu.Unlock()
		return nil
	})
	err := eg.Wait()
	if err != nil {
		return nil, err
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].Ctime.Unix() > events[j].Ctime.Unix()
	})
	return events[:min[int](int(limit), len(events))], nil
}
