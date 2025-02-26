package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strings"
	"time"
)

var (
	AtKey = []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0")
	RtKey = []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvfx")
)

type RedisJwtHandler struct {
	cmd redis.Cmdable
}

func NewRedisJWTHandler(cmd redis.Cmdable) Handler {
	return &RedisJwtHandler{
		cmd: cmd,
	}
}

func (h *RedisJwtHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New()
	err := h.SetJWTToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	err = h.SetRefreshToken(ctx, uid, ssid)
	return err
}

func (h *RedisJwtHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	//要携带数据，要传入一个Claim接口，用它自带的mapClaim麻烦，自己实现一个接口:继承自带的，再加上要放入的信息
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Id:        uid,
		Ssid:      ssid,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//设置好token之后，返回前端
	tokenStr, err := token.SignedString(AtKey)
	if err != nil {
		return err
	}
	//放在 Header 中，返给前端
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (h *RedisJwtHandler) SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	//要携带数据，要传入一个Claim接口，用它自带的mapClaim麻烦，自己实现一个接口:继承自带的，再加上要放入的信息
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		Uid:  uid,
		Ssid: ssid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//设置好token之后，返回前端
	tokenStr, err := token.SignedString(RtKey)
	if err != nil {
		return err
	}
	//放在 Header 中，返给前端
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

func (h *RedisJwtHandler) ClearToken(ctx *gin.Context) error {
	// 前端 token 清除，正常用户退出登录功能已经实现
	// 也就是说在登录校验里面，走不到查 redis 那一步就返回了
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")

	// 后端记录黑名单，以防攻击者拿着 token 访问系统
	claims := ctx.MustGet("claims").(*UserClaims)
	return h.cmd.Set(ctx, fmt.Sprintf("users:ssid:%s", claims.Ssid),
		"", time.Hour*24*7).Err()
}

func (h *RedisJwtHandler) CheckSession(ctx *gin.Context, ssid string) error {
	val, err := h.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	switch err {
	case redis.Nil:
		return nil
	case nil:
		if val == 0 {
			return nil
		}
		return errors.New("session 已经无效了")
	default:
		return err
	}
}

func (h *RedisJwtHandler) ExtractToken(ctx *gin.Context) string {
	tokenHeader := ctx.GetHeader("Authorization")
	segs := strings.Split(tokenHeader, " ")
	//格式不对，Authorization 中的内容是乱传的
	if len(segs) != 2 {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return ""
	}
	return segs[1]
}

func (h *RedisJwtHandler) RefreshToken(ctx *gin.Context) error {
	// 只有这个接口，拿出来的才是 refresh_token，其它地方都是 access token
	refreshToken := h.ExtractToken(ctx)
	var claims RefreshClaims
	token, err := jwt.ParseWithClaims(refreshToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return RtKey, nil
	})
	if err != nil || !token.Valid {
		return err
	}
	err = h.CheckSession(ctx, claims.Ssid)
	if err != nil {
		return err
	}
	// 长 token 校验通过，搞个新的短 token:access_token
	err = h.SetJWTToken(ctx, claims.Uid, claims.Ssid)
	if err != nil {
		return err
	}
	return nil
}

func (h RedisJwtHandler) SetStateJWTToken(ctx *gin.Context, State string) (string, error) {
	//要携带数据，要传入一个Claim接口，用它自带的mapClaim麻烦，自己实现一个接口:继承自带的，再加上要放入的信息
	claims := StateClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		State: State,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//设置好token之后，这里不返回前端，因为后续回来不经过前端，所以直接将 token 放 cookie 中
	return token.SignedString(RtKey)
}
