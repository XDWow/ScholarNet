//go:build wireinject

package main

import (
	"github.com/XD/ScholarNet/cmd/interactive/events"
	"github.com/XD/ScholarNet/cmd/interactive/grpc"
	"github.com/XD/ScholarNet/cmd/interactive/ioc"
	"github.com/XD/ScholarNet/cmd/interactive/repository"
	"github.com/XD/ScholarNet/cmd/interactive/repository/cache"
	"github.com/XD/ScholarNet/cmd/interactive/repository/dao"
	"github.com/XD/ScholarNet/cmd/interactive/service"
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
