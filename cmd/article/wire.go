//go:build wireinject

package main

import (
	"github.com/XD/ScholarNet/cmd/article/events"
	"github.com/XD/ScholarNet/cmd/article/grpc"
	"github.com/XD/ScholarNet/cmd/article/ioc"
	"github.com/XD/ScholarNet/cmd/article/repository"
	"github.com/XD/ScholarNet/cmd/article/repository/cache"
	"github.com/XD/ScholarNet/cmd/article/repository/dao"
	"github.com/XD/ScholarNet/cmd/article/service"
	"github.com/XD/ScholarNet/cmd/pkg/wego"
	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(
	ioc.InitRedis,
	ioc.InitLogger,
	ioc.InitUserRpcClient,
	ioc.InitProducer,
	ioc.InitEtcdClient,
	ioc.InitDB,
)

func Init() *wego.App {
	wire.Build(
		thirdProvider,
		events.NewSaramaSyncProducer,
		cache.NewRedisArticleCache,
		dao.NewGORMArticleDAO,
		repository.NewArticleRepository,
		repository.NewGrpcAuthorRepository,
		service.NewArticleService,
		grpc.NewArticleServiceServer,
		ioc.InitGRPCxServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
