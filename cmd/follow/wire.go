//go:build wireinject

package main

import (
	grpc2 "github.com/XD/ScholarNet/cmd/follow/grpc"
	"github.com/XD/ScholarNet/cmd/follow/ioc"
	"github.com/XD/ScholarNet/cmd/follow/repository"
	"github.com/XD/ScholarNet/cmd/follow/repository/cache"
	"github.com/XD/ScholarNet/cmd/follow/repository/dao"
	"github.com/XD/ScholarNet/cmd/follow/service"
	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(
	ioc.InitDB,
	ioc.InitRedis,
	ioc.InitLogger,
)

var serviceProvider = wire.NewSet(
	dao.NewFollowRelationDao,
	cache.NewRedisFollowCache,
	repository.NewFollowRepository,
	service.NewFollowRelationService,
	grpc2.NewFollowServiceServer,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		serviceProvider,
		ioc.InitGRPCxServer,
		wire.Struct(new(App), "*"))
	return new(App)
}
