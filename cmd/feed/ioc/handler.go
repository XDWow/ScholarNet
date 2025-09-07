package ioc

import (
	followv1 "github.com/XD/ScholarNet/cmd/api/proto/gen/follow/v1"
	"github.com/XD/ScholarNet/cmd/feed/repository"
	"github.com/XD/ScholarNet/cmd/feed/service"
)

func RegisterHandler(repo repository.FeedEventRepo, followClient followv1.FollowServiceClient) map[string]service.Handler {
	articleHandler := service.NewArticleEventHandler(repo, followClient)
	followHanlder := service.NewFollowEventHandler(repo)
	likeHandler := service.NewLikeEventHandler(repo)
	return map[string]service.Handler{
		service.ArticleEventName: articleHandler,
		service.FollowEventName:  followHanlder,
		service.LikeEventName:    likeHandler,
	}
}
