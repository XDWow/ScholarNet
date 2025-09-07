package main

import (
	"github.com/XD/ScholarNet/cmd/pkg/wego"
	"github.com/XD/ScholarNet/cmd/tag/grpc"
	"github.com/XD/ScholarNet/cmd/tag/ioc"
	"github.com/XD/ScholarNet/cmd/tag/repository/cache"
	"github.com/XD/ScholarNet/cmd/tag/repository/dao"
	"github.com/XD/ScholarNet/cmd/tag/service"
	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(
	ioc.InitRedis,
	ioc.InitLogger,
	ioc.InitDB,
	ioc.InitEtcdClient,
)

func Init() *wego.App {
	wire.Build(
		thirdProvider,
		cache.NewRedisTagCache,
		dao.NewGORMTagDAO,
		ioc.InitRepository,
		service.NewTagService,
		grpc.NewTagServiceServer,
		ioc.InitGRPCxServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
