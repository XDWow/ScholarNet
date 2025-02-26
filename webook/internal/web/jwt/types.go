package jwt

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Handler interface {
	SetLoginToken(ctx *gin.Context, uid int64) error
	SetJWTToken(ctx *gin.Context, uid int64, ssid string) error
	SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error
	ClearToken(ctx *gin.Context) error
	CheckSession(ctx *gin.Context, ssid string) error
	ExtractToken(ctx *gin.Context) string
	SetStateJWTToken(ctx *gin.Context, State string) (string, error)
	RefreshToken(ctx *gin.Context) error
}

type UserClaims struct {
	//RegisteredClaims 实现了这个接口
	jwt.RegisteredClaims
	//你自己要放进去token的数据
	Id   int64
	Ssid string
	// 自己随便加,这里短 token 容易泄露，可以加一个校验机制，提升安全性
	UserAgent string
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string
}

type StateClaims struct {
	//RegisteredClaims 实现了这个接口
	jwt.RegisteredClaims
	//你自己要放进去token的数据
	State string
}
