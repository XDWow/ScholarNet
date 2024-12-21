package jwt

import "github.com/gin-gonic/gin"

type Handler interface {
	SetJWTToken(ctx *gin.Context, uid int64) error
}
