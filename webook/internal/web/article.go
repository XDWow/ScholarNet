package web

import (
	"github.com/LXD-c/basic-go/webook/internal/service"
	"github.com/LXD-c/basic-go/webook/pkg/logger"
	"github.com/gin-gonic/gin"
)

type ArticleHandler struct {
	svc service.ArticleService
	l   logger.LoggerV1
}

func NewArticleHandler(svc service.ArticleService, l logger.LoggerV1) *ArticleHandler {
	return &ArticleHandler{
		svc: svc,
		l:   l,
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/article")
	g.POST("/edit", h.Edit)
	g.POST("publish", h.Publish)
}

func (h *ArticleHandler) Edit(context *gin.Context) {

}

func (h *ArticleHandler) Publish(context *gin.Context) {

}
