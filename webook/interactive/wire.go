//go:build wireinject

package main

import (
	"github.com/LXD-c/basic-go/webook/interactive/events"
	"github.com/LXD-c/basic-go/webook/interactive/grpc"
	"github.com/LXD-c/basic-go/webook/interactive/ioc"
	"github.com/LXD-c/basic-go/webook/interactive/repository"
	"github.com/LXD-c/basic-go/webook/interactive/repository/cache"
	"github.com/LXD-c/basic-go/webook/interactive/repository/dao"
	"github.com/LXD-c/basic-go/webook/interactive/service"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(
	ioc.InitSRC,
	ioc.InitDST,
	ioc.InitBizDB,
	ioc.InitDoubleWritePool,
	ioc.InitRedis,
	ioc.InitKafka,
	ioc.InitSyncProducer,
	ioc.InitLogger)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	cache.NewRedisInteractiveCache,
	dao.NewGORMInteractiveDAO,
)

var migratorProvider = wire.NewSet(
	ioc.InitMigratorProducer,
	ioc.InitFixDataConsumer,
	ioc.InitMigratorWeb)

func InitApp() *App {
	wire.Build(interactiveSvcProvider,
		thirdPartySet,
		migratorProvider,
		events.NewInteractiveReadEventConsumer,
		ioc.NewConsumers,

		grpc.NewInteractiveServiceServer,
		ioc.InitGRPCxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
