//go:build wireinject

package main

import (
	"github.com/XD/ScholarNet/cmd/pkg/wego"
	"github.com/XD/ScholarNet/cmd/user/grpc"
	"github.com/XD/ScholarNet/cmd/user/ioc"
	"github.com/XD/ScholarNet/cmd/user/repository"
	"github.com/XD/ScholarNet/cmd/user/repository/cache"
	"github.com/XD/ScholarNet/cmd/user/repository/dao"
	"github.com/XD/ScholarNet/cmd/user/service"

	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(
	ioc.InitLogger,
	ioc.InitDB,
	ioc.InitRedis,
	ioc.InitEtcdClient,
)

func Init() *wego.App {
	wire.Build(
		thirdProvider,
		cache.NewRedisUserCache,
		dao.NewGORMUserDAO,
		repository.NewCachedUserRepository,
		service.NewUserService,
		grpc.NewUserServiceServer,
		ioc.InitGRPCxServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
