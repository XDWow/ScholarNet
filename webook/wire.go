//go:build wireinject

package main

import (
	"github.com/LXD-c/basic-go/webook/internal/events/article"
	"github.com/LXD-c/basic-go/webook/internal/repository"
	article2 "github.com/LXD-c/basic-go/webook/internal/repository/article"
	"github.com/LXD-c/basic-go/webook/internal/repository/cache"
	"github.com/LXD-c/basic-go/webook/internal/repository/dao"
	article3 "github.com/LXD-c/basic-go/webook/internal/repository/dao/article"
	"github.com/LXD-c/basic-go/webook/internal/service"
	"github.com/LXD-c/basic-go/webook/internal/web"
	ijwt "github.com/LXD-c/basic-go/webook/internal/web/jwt"
	"github.com/LXD-c/basic-go/webook/ioc"
	"github.com/google/wire"
)

func InitWebServer() *App {
	wire.Build(
		//最基础的第三方依赖
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLogger,
		ioc.InitKafka,
		ioc.NewConsumers,
		ioc.NewSyncProducer,

		// consumer
		article.NewInteractiveReadEventBatchConsumer,
		article.NewKafkaProducer,

		//初始化 DAO
		dao.NewUserDAO,
		article3.NewGORMArticleDAO,
		dao.NewGORMInteractiveDAO,

		cache.NewUserCache,
		cache.NewCodeCache,
		cache.NewRedisInteractiveCache,
		cache.NewRedisArticleCache,
		repository.NewUserRepository,
		repository.NewCodeRepository,
		article2.NewArticleRepository,
		repository.NewCachedInteractiveRepository,
		//article.NewArticleReaderRepository,
		//article.NewArticleAuthorRepository,

		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,
		service.NewInteractiveServiceImpl,

		// 直接基于内存实现
		ioc.InitSMSService,
		ioc.InitWechatService,

		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		ioc.NewWechatHandlerConfig,
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
