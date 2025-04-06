package ioc

import (
	"context"
	"github.com/LXD-c/basic-go/webook/internal/web"
	ijwt "github.com/LXD-c/basic-go/webook/internal/web/jwt"
	"github.com/LXD-c/basic-go/webook/internal/web/middleware"
	"github.com/LXD-c/basic-go/webook/pkg/ginx"
	"github.com/LXD-c/basic-go/webook/pkg/ginx/middlewares/logger"
	"github.com/LXD-c/basic-go/webook/pkg/ginx/middlewares/metric"
	logger2 "github.com/LXD-c/basic-go/webook/pkg/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"strings"
	"time"
)

func InitWebServer(mdls []gin.HandlerFunc, uerHdl *web.UserHandler,
	oauth2WechatHdl *web.OAuth2WechatHandler, articleHdl *web.ArticleHandler) *gin.Engine {
	server := gin.Default()
	//middlewares...将切片解包为多个独立的参数
	server.Use(mdls...)
	uerHdl.RegisterRoutes(server)
	oauth2WechatHdl.RegisterRoutes(server)
	articleHdl.RegisterRoutes(server)
	(&web.ObservabilityHandler{}).RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable, l logger2.LoggerV1, jwthdl ijwt.Handler) []gin.HandlerFunc {
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "basic_go",
		Subsystem: "webook",
		Name:      "http_biz_code",
		Help:      "HTTP 的业务错误码",
	})
	return []gin.HandlerFunc{
		corsHdl(),

		metric.NewBuilder(
			"basic-go",
			"webook",
			"gin_http",
			"统计 GIN 的 HTTP 接口",
			"my-instance-1").Build(),

		logger.NewBuilder(func(ctx context.Context, al *logger.AccessLog) {
			l.Debug("HTTP请求", logger2.Field{Key: "al", Value: al})
		}).AllowReqBody().AllowRespBody().Build(),

		otelgin.Middleware("webook"),
		middleware.NewLoginJWTMiddlewareBuilder(jwthdl).
			IgnorePaths("/users/signup").
			IgnorePaths("/users/refresh_token").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms").
			IgnorePaths("/oauth2/wechat/authurl").
			IgnorePaths("/oauth2/wechat/callback").
			IgnorePaths("/users/login").
			Build(),
		//ratelimit.NewBuilder(redisClient,)
	}
}

func corsHdl() gin.HandlerFunc {
	//使用use方法注册middleware，这个中间件是用于解决 CORS 的 middleware
	return cors.New(cors.Config{
		//是否允许带上用户认证信息（比如 cookie）
		AllowCredentials: true,
		AllowHeaders:     []string{"content-type", "Authorization"},
		ExposeHeaders:    []string{"x-jwt-token", "x-refresh-token"},
		//哪些来源是允许的
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "live.webook.com")
		},
		MaxAge: 12 * time.Hour,
	})
}
