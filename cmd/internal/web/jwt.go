package web

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type jwtHandler struct {
}

func (h *jwtHandler) setJWTToken(ctx *gin.Context, uid int64) error {
	//要携带数据，要传入一个Claim接口，用它自带的mapClaim麻烦，自己实现一个接口:继承自带的，再加上要放入的信息
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid:       uid,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//设置好token之后，返回前端
	tokenStr, err := token.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"))
	if err != nil {
		return err
	}
	//放在 Header 中，返给前端
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

type UserClaims struct {
	//RegisteredClaims 实现了这个接口
	jwt.RegisteredClaims
	//你自己要放进去token的数据
	Uid       int64
	UserAgent string
}
