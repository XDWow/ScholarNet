//go:build wireinject

package startup

import (
	article3 "github.com/XD/ScholarNet/cmd/internal/events/article"
	"github.com/XD/ScholarNet/cmd/internal/repository"
	article2 "github.com/XD/ScholarNet/cmd/internal/repository/article"
	"github.com/XD/ScholarNet/cmd/internal/repository/cache"
	"github.com/XD/ScholarNet/cmd/internal/repository/dao"
	"github.com/XD/ScholarNet/cmd/internal/repository/dao/article"
	"github.com/XD/ScholarNet/cmd/internal/service"
	"github.com/XD/ScholarNet/cmd/internal/web"
	ijwt "github.com/XD/ScholarNet/cmd/internal/web/jwt"
	"github.com/XD/ScholarNet/cmd/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(InitRedis,
	NewSyncProducer,
	InitKafka,
	InitTestDB, InitLog)
var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	cache.NewUserCache,
	repository.NewUserRepository,
	service.NewUserService)
var articlSvcProvider = wire.NewSet(
	article.NewGORMArticleDAO,
	article2.NewArticleRepository,
	service.NewArticleService)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdProvider,
		userSvcProvider,
		articlSvcProvider,

		article3.NewKafkaProducer,
		cache.NewCodeCache,
		repository.NewCodeRepository,
		// service 部分
		// 集成测试我们显式指定使用内存实现
		ioc.InitSMSService,

		// 指定啥也不干的 wechat service
		InitPhantomWechatService,
		service.NewCodeService,
		// handler 部分
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,
		ijwt.NewRedisJWTHandler,

		// gin 的中间件
		ioc.InitMiddlewares,

		// Web 服务器
		ioc.InitWebServer,
	)
	// 随便返回一个
	return gin.Default()
}

func InitArticleHandler(dao article.ArticleDAO) *web.ArticleHandler {
	wire.Build(thirdProvider,
		//userSvcProvider,
		//cache.NewRedisArticleCache,
		//wire.InterfaceValue(new(article.ArticleDAO), dao),
		article3.NewKafkaProducer,
		article2.NewArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler)
	return new(web.ArticleHandler)
}

func InitUserSvc() service.UserService {
	wire.Build(thirdProvider, userSvcProvider)
	return service.NewUserService(nil, nil)
}

func InitJwtHdl() ijwt.Handler {
	wire.Build(thirdProvider, ijwt.NewRedisJWTHandler)
	return ijwt.NewRedisJWTHandler(nil)
}
