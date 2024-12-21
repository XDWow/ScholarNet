package jwt

import "github.com/gin-gonic/gin"

type RedisJwtHandler struct {
}

func (h *RedisJwtHandler) SetJWTToken(ctx *gin.Context, uid int64) error {

}
