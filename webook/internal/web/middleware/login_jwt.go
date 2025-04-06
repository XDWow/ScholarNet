package middleware

import (
	ijwt "github.com/LXD-c/basic-go/webook/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

// LoginJWTMiddlewareBuilder JWT 登录校验
type LoginJWTMiddlewareBuilder struct {
	paths []string
	ijwt.Handler
}

func NewLoginJWTMiddlewareBuilder(jwthdl ijwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		Handler: jwthdl,
	}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		//不需要登录校验的
		//path := ctx.Request.URL.Path
		//if path == "/users/login" || path == "/users/signup" {
		//	return
		//}
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		tokenStr := l.ExtractToken(ctx)
		claims := &ijwt.UserClaims{}
		//ParseWithClaims 带Claim的解析，会将Claim解析出来放在参数的Claim里面，所以这里用指针（赋值）
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return ijwt.AtKey, nil
		})
		if err != nil {
			// token不对，比如伪造的:Bearer xxx1234,就会解析失败返回错误
			// token 过期
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if claims.UserAgent != ctx.GetHeader("User-Agent") {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// 检查是否已经退出登录，redis 里面查黑名单
		err = l.CheckSession(ctx, claims.Ssid)
		if err != nil {
			// 要么 redis 出问题了，要么已经退出登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ctx.Set("users", claims)

		// 你以为的退出登录，没有用的,别人拿到你的 token，valid 随便改
		//token.Valid = false
		//// tokenStr 是一个新的字符串
		//tokenStr, err = token.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"))
		//if err != nil {
		//	// 记录日志
		//	log.Println("jwt 续约失败", err)
		//}
		//ctx.Header("x-jwt-token", tokenStr)

		// 短的 token 过期了，搞个新的
		//now := time.Now()
		// 每十秒钟刷新一次
		//if claims.ExpiresAt.Sub(now) < time.Second*50 {
		//	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute))
		//	tokenStr, err = token.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"))
		//	if err != nil {
		//		// 记录日志
		//		log.Println("jwt 续约失败", err)
		//	}
		//	ctx.Header("x-jwt-token", tokenStr)
		//}
		//ctx.Set("userId", claims.Uid)
	}
}
