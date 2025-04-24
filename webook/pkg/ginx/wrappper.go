package ginx

import (
	"github.com/LXD-c/basic-go/webook/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
)

// 这个东西，放到你们的 ginx 插件库里面去
// 技术含量不是很高，但是绝对有技巧
// L 使用包变量
var L logger.LoggerV1

var vector *prometheus.CounterVec

// 这里创建并注册
// 在具体代码中，调用 vector.WithLabelValues("200").Inc() 来增加对应状态码的计数
func InitCounter(opt prometheus.CounterOpts) {
	vector = prometheus.NewCounterVec(opt, []string{"code"})
	prometheus.MustRegister(vector)
	// 你还可以考虑使用 code, method, 命中路由，HTTP 状态码
}

func WrapBodyAndToken[Req any, C jwt.Claims](fn func(ctx *gin.Context, req Req, uc C) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			return
		}

		val, ok := ctx.Get("users")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c, ok := val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 下半段的业务逻辑从哪里来？
		// 我的业务逻辑有可能要操作 ctx
		// 你要读取 HTTP HEADER
		res, err := fn(ctx, req, c)
		if err != nil {
			// 开始处理 error，其实就是记录一下日志
			L.Error("处理业务逻辑出错",
				logger.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func Wraptoken[C jwt.Claims](fn func(ctx *gin.Context, c C) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		val, ok := ctx.Get("users")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c, ok := val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		res, err := fn(ctx, c)
		if err != nil {
			// 开始处理 error，其实就是记录一下日志
			L.Error("处理业务逻辑出错",
				logger.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
		}
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		ctx.JSON(http.StatusOK, res)
	}
}

func Wrap(fn func(ctx *gin.Context) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res, err := fn(ctx)
		if err != nil {
			// 开始处理 error，其实就是记录一下日志
			L.Error("处理业务逻辑出错",
				logger.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
		}
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBodyV1[T any](fn func(ctx *gin.Context, req T) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}

		// 下半段的业务逻辑从哪里来？
		// 我的业务逻辑有可能要操作 ctx
		// 你要读取 HTTP HEADER
		res, err := fn(ctx, req)
		if err != nil {
			// 开始处理 error，其实就是记录一下日志
			L.Error("处理业务逻辑出错",
				logger.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}
