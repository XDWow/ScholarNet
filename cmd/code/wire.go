//go:build wireinject

package main

import (
	"github.com/XD/ScholarNet/cmd/code/grpc"
	"github.com/XD/ScholarNet/cmd/code/ioc"
	"github.com/XD/ScholarNet/cmd/code/repository"
	"github.com/XD/ScholarNet/cmd/code/repository/cache"
	"github.com/XD/ScholarNet/cmd/code/service"
	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(
	ioc.InitRedis,
	ioc.InitEtcdClient,
	ioc.InitLogger,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		ioc.InitSmsRpcClient,
		cache.NewRedisCodeCache,
		repository.NewCachedCodeRepository,
		service.NewSMSCodeService,
		grpc.NewCodeServiceServer,
		ioc.InitGRPCxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
