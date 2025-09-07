package test

import (
	feedv1 "github.com/XD/ScholarNet/cmd/api/proto/gen/feed/v1"
	followMocks "github.com/XD/ScholarNet/cmd/api/proto/gen/follow/v1/mocks"
	"github.com/XD/ScholarNet/cmd/feed/grpc"
	"github.com/XD/ScholarNet/cmd/feed/ioc"
	"github.com/XD/ScholarNet/cmd/feed/repository"
	"github.com/XD/ScholarNet/cmd/feed/repository/cache"
	"github.com/XD/ScholarNet/cmd/feed/repository/dao"
	"github.com/XD/ScholarNet/cmd/feed/service"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
	"testing"
)

func InitGrpcServer(t *testing.T) (feedv1.FeedSvcServer, *followMocks.MockFollowServiceClient, *gorm.DB) {
	loggerV1 := ioc.InitLogger()
	db := ioc.InitDB(loggerV1)
	feedPullEventDAO := dao.NewFeedPullEventDAO(db)
	feedPushEventDAO := dao.NewFeedPushEventDAO(db)
	cmdable := ioc.InitRedis()
	feedEventCache := cache.NewFeedEventCache(cmdable)
	feedEventRepo := repository.NewFeedEventRepo(feedPullEventDAO, feedPushEventDAO, feedEventCache)
	mockCtrl := gomock.NewController(t)
	followClient := followMocks.NewMockFollowServiceClient(mockCtrl)
	v := ioc.RegisterHandler(feedEventRepo, followClient)
	feedService := service.NewFeedService(feedEventRepo, v)
	feedEventGrpcSvc := grpc.NewFeedEventGrpcSvc(feedService)
	return feedEventGrpcSvc, followClient, db
}
