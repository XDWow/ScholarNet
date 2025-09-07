package service

import (
	"context"
	"github.com/XD/ScholarNet/cmd/feed/domain"
	"github.com/XD/ScholarNet/cmd/feed/repository"
)

const (
	LikeEventName = "like_event"
)

type LikeEventHandler struct {
	repo repository.FeedEventRepo
}

// CreateFeedEvent 中的 ext 里面至少需要三个 id，线下协商好
// liked int64: 被点赞的人
// liker int64：点赞的人
// bizId int64: 被点赞的东西
// biz: string
func (h *LikeEventHandler) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	uid, err := ext.Get("liked").AsInt64()
	if err != nil {
		return err
	}
	return h.repo.CreatePushEvents(ctx, []domain.FeedEvent{
		{
			// 收件人
			Uid:  uid,
			Type: LikeEventName,
			Ext:  ext,
		},
	})
}

func (h *LikeEventHandler) FindFeedEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	// 如果你有扩展表的机制
	// 在这里查。你的 repository LikeEventRepository
	// 如果要是你在数据库存储的时候，没有冗余用户的昵称
	// BFF（你的业务方） 又不愿意去聚合（调用用户服务获得昵称）
	// 就得你在这里查,所以这里才不同业务分开查 withTyp
	return h.repo.FindPushEventsWithTyp(ctx, LikeEventName, uid, timestamp, limit)
}

func NewLikeEventHandler(repo repository.FeedEventRepo) *LikeEventHandler {
	return &LikeEventHandler{repo}
}
