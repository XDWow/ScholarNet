//go:build wireinject

package main

import (
	"github.com/XD/ScholarNet/cmd/feed/events"
	"github.com/XD/ScholarNet/cmd/feed/grpc"
	"github.com/XD/ScholarNet/cmd/feed/ioc"
	"github.com/XD/ScholarNet/cmd/feed/repository"
	"github.com/XD/ScholarNet/cmd/feed/repository/cache"
	"github.com/XD/ScholarNet/cmd/feed/repository/dao"
	"github.com/XD/ScholarNet/cmd/feed/service"
	"github.com/google/wire"
)

var serviceProviderSet = wire.NewSet(
	dao.NewFeedPushEventDAO,
	dao.NewFeedPullEventDAO,
	cache.NewFeedEventCache,
	repository.NewFeedEventRepo,
)

var thirdProvider = wire.NewSet(
	ioc.InitEtcdClient,
	ioc.InitLogger,
	ioc.InitRedis,
	ioc.InitKafka,
	ioc.InitDB,
	ioc.InitFollowClient,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		serviceProviderSet,
		ioc.RegisterHandler,
		service.NewFeedService,
		grpc.NewFeedEventGrpcSvc,
		events.NewArticleEventConsumer,
		events.NewFeedEventConsumer,
		ioc.InitGRPCxServer,
		ioc.NewConsumers,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
