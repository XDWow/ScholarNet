//go:build wireinject

package main

import (
	"github.com/XD/ScholarNet/cmd/interactive/events"
	repository2 "github.com/XD/ScholarNet/cmd/interactive/repository"
	cache2 "github.com/XD/ScholarNet/cmd/interactive/repository/cache"
	dao2 "github.com/XD/ScholarNet/cmd/interactive/repository/dao"
	service2 "github.com/XD/ScholarNet/cmd/interactive/service"
	"github.com/XD/ScholarNet/cmd/internal/events/article"
	"github.com/XD/ScholarNet/cmd/internal/repository"
	article2 "github.com/XD/ScholarNet/cmd/internal/repository/article"
	"github.com/XD/ScholarNet/cmd/internal/repository/cache"
	"github.com/XD/ScholarNet/cmd/internal/repository/dao"
	article3 "github.com/XD/ScholarNet/cmd/internal/repository/dao/article"
	"github.com/XD/ScholarNet/cmd/internal/service"
	"github.com/XD/ScholarNet/cmd/internal/web"
	ijwt "github.com/XD/ScholarNet/cmd/internal/web/jwt"
	"github.com/XD/ScholarNet/cmd/ioc"
	"github.com/google/wire"
	rlock "github.com/gotomicro/redis-lock"
)

var interactiveSvcProvider = wire.NewSet(
	service2.NewInteractiveService,
	repository2.NewCachedInteractiveRepository,
	dao2.NewGORMInteractiveDAO,
	cache2.NewRedisInteractiveCache,
)

var rankingSvcProvider = wire.NewSet(
	service.NewBatchRankingService,
	repository.NewCachedRankingRepository,
	cache.NewRankingRedisCache,
	cache.NewRankingLocalCache,
)

func InitWebServer() *App {
	wire.Build(
		//最基础的第三方依赖
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLogger,
		ioc.InitKafka,
		ioc.NewConsumers,
		ioc.NewSyncProducer,
		rlock.NewClient,

		interactiveSvcProvider,
		rankingSvcProvider,
		ioc.InitIntrGRPCClient,
		ioc.InitJobs,
		ioc.InitRankingJob,

		// consumer
		events.NewInteractiveReadEventBatchConsumer,
		article.NewKafkaProducer,

		//初始化 DAO
		dao.NewUserDAO,
		article3.NewGORMArticleDAO,

		cache.NewUserCache,
		cache.NewCodeCache,
		cache.NewRedisArticleCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,
		article2.NewArticleRepository,
		//article.NewArticleReaderRepository,
		//article.NewArticleAuthorRepository,

		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,

		// 直接基于内存实现
		ioc.InitSMSService,
		ioc.InitWechatService,

		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		//ioc.NewWechatHandlerConfig,
		ijwt.NewRedisJWTHandler,
		web.NewArticleHandler,
		// 你中间件呢？
		// 你注册路由呢？
		// 你这个地方没有用到前面的任何东西
		// gin.Default,
		ioc.InitWebServer,
		ioc.InitMiddlewares,
		// 组装我这个结构体的所有字段
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
