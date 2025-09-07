//go:build wireinject

package main

import (
	"github.com/XD/ScholarNet/cmd/pkg/wego"
	"github.com/XD/ScholarNet/cmd/ranking/events"
	"github.com/XD/ScholarNet/cmd/ranking/grpc"
	"github.com/XD/ScholarNet/cmd/ranking/ioc"
	"github.com/XD/ScholarNet/cmd/ranking/repository"
	"github.com/XD/ScholarNet/cmd/ranking/repository/cache"
	"github.com/XD/ScholarNet/cmd/ranking/service"
	"github.com/google/wire"
	rlock "github.com/gotomicro/redis-lock"
)

var serviceProviderSet = wire.NewSet(
	cache.NewRedisRankingCache,
	repository.NewCachedRankingRepository,
	service.NewBatchRankingService,
)

var thirdProvider = wire.NewSet(
	ioc.InitRedis,
	ioc.InitLogger,
	ioc.InitInterActiveRpcClient,
	ioc.InitArticleRpcClient,
	rlock.NewClient,
)

var kafkaProvider = wire.NewSet(
	ioc.InitKafkaClient,
	ioc.InitKafkaProducer,
)

var cronJob = wire.NewSet(
	ioc.InitRankingJob,
	ioc.InitLocalCacheRefreshJob,
	ioc.InitJobs,
)

func Init() *wego.App {
	wire.Build(
		thirdProvider,
		kafkaProvider,
		serviceProviderSet,
		cronJob,
		grpc.NewRankingServiceServer,
		ioc.InitGRPCxServer,
		wire.Struct(new(wego.App), "GRPCServer", "Cron"),
	)
	return new(wego.App)
}

func InitConsumer() events.Consumer {
	wire.Build(
		thirdProvider,
		kafkaProvider,
		ioc.InitKafkaConsumer,
	)
	return nil
}
