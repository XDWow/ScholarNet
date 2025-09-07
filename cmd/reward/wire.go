//go:build wireinject

package main

import (
	"github.com/XD/ScholarNet/cmd/pkg/wego"
	"github.com/XD/ScholarNet/cmd/reward/grpc"
	"github.com/XD/ScholarNet/cmd/reward/ioc"
	"github.com/XD/ScholarNet/cmd/reward/repository"
	"github.com/XD/ScholarNet/cmd/reward/repository/cache"
	"github.com/XD/ScholarNet/cmd/reward/repository/dao"
	"github.com/XD/ScholarNet/cmd/reward/service"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitEtcdClient,
	ioc.InitRedis)

func Init() *wego.App {
	wire.Build(thirdPartySet,
		service.NewWechatNativeRewardService,
		ioc.InitAccountClient,
		ioc.InitGRPCxServer,
		ioc.InitPaymentClient,
		repository.NewRewardRepository,
		cache.NewRewardRedisCache,
		dao.NewRewardGORMDAO,
		grpc.NewRewardServiceServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
