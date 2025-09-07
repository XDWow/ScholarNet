//go:build wireinject

package main

import (
	"github.com/XD/ScholarNet/cmd/search/events"
	"github.com/XD/ScholarNet/cmd/search/grpc"
	"github.com/XD/ScholarNet/cmd/search/ioc"
	"github.com/XD/ScholarNet/cmd/search/repository"
	"github.com/XD/ScholarNet/cmd/search/repository/dao"
	"github.com/XD/ScholarNet/cmd/search/service"
	"github.com/google/wire"
)

var thirdProvicer = wire.NewSet(
	ioc.InitESClient,
	ioc.InitLogger,
	ioc.InitKafka,
	ioc.InitEtcdClient,
)

var serviceProviderSet = wire.NewSet(
	dao.NewUserElasticDAO,
	dao.NewArticleElasticDAO,
	dao.NewAnyElasticDAO,
	dao.NewTagElasticDAO,
	repository.NewUserRepository,
	repository.NewArticleRepository,
	repository.NewAnyRepository,
	service.NewSearchService,
	service.NewSyncService,
)

func Init() *App {
	wire.Build(
		serviceProviderSet,
		thirdProvicer,
		// 接下来就是 grpc,消费者这些额外的东西
		grpc.NewSyncServiceServer,
		grpc.NewSearchService,
		ioc.InitGRPCxServer,
		events.NewUserConsumer,
		events.NewArticleConsumer,
		events.NewAnyConsumer,
		ioc.NewConsumers,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
