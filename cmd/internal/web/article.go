package web

import (
	"fmt"
	intrv1 "github.com/XD/ScholarNet/cmd/api/proto/gen/intr/v1"
	"github.com/XD/ScholarNet/cmd/internal/domain"
	"github.com/XD/ScholarNet/cmd/internal/service"
	ijwt "github.com/XD/ScholarNet/cmd/internal/web/jwt"
	"github.com/XD/ScholarNet/cmd/pkg/ginx"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
)

type ArticleHandler struct {
	svc service.ArticleService
	l   logger.LoggerV1

	intrSvc intrv1.InteractiveServiceClient
	biz     string
}

func NewArticleHandler(svc service.ArticleService, l logger.LoggerV1, intrSvc intrv1.InteractiveServiceClient) *ArticleHandler {
	return &ArticleHandler{
		svc:     svc,
		l:       l,
		intrSvc: intrSvc,
		biz:     "article",
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", h.Edit)
	g.POST("/publish", h.Publish)
	g.POST("/withdraw", h.Withdraw)
	// 创作者的查询接口
	// 这个是获取数据的接口，理论上来说（遵循 RESTful 规范），应该是用 GET 方法
	// GET localhost/articles => List 接口
	g.POST("/list", ginx.WrapBodyAndToken[ListReq, ijwt.UserClaims](h.list))
	g.GET("/detail/:id", ginx.Wraptoken[ijwt.UserClaims](h.Detail))

	pub := g.Group("/pub")
	pub.GET("/:id", h.PubDetail)
	// 点赞是这个接口，取消点赞也是这个接口
	// RESTful 风格
	//pub.POST("/like/:id", ginx.WrapBodyAndToken[LikeReq,
	//	ijwt.UserClaims](h.Like))
	pub.POST("/like", ginx.WrapBodyAndToken[LikeReq, ijwt.UserClaims](h.Like))
	// 若不复用接口
	//pub.POST("/cancel_like", ginx.WrapBodyAndToken[LikeReq,
	//	ijwt.UserClaims](h.Like))
}

func (h *ArticleHandler) Publish(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		// 解析错了，就会直接写回一个 400 错误

		// 不用手动写，而且也不会是 200
		//ctx.JSON(http.StatusOK, Result{
		//	Code: 5,
		//	Msg:  "前端请求解析失败",
		//})
		return
	}
	c := ctx.MustGet("users")
	claims, ok := c.(ijwt.UserClaims)
	if !ok {
		// 你可以考虑监控住这里
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的 session 信息")
		return
	}

	id, err := h.svc.Publish(ctx, req.toDomain(claims.Id))
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		// 可以打日志
		h.l.Error("服务层发表帖子返回错误", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg:  "OK",
		Data: id,
	})
}

func (h *ArticleHandler) Edit(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	c := ctx.MustGet("users")
	claims, ok := c.(ijwt.UserClaims)
	if !ok {
		// 你可以考虑监控住这里
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的 session 信息")
		return
	}
	id, err := h.svc.Save(ctx, req.toDomain(claims.Id))
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		// 打日志？
		h.l.Error("保存帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg:  "OK",
		Data: id,
	})
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64 `json:"id"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	c := ctx.MustGet("users")
	claims, ok := c.(ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的 session 信息")
		return
	}

	err := h.svc.Withdraw(ctx, domain.Article{
		Id: req.Id,
		Author: domain.Author{
			Id: claims.Id,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("隐藏帖子失败", logger.Error(err))
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "OK",
	})
}

// 创作者查看自己的文章详情
func (a *ArticleHandler) Detail(ctx *gin.Context, usr ijwt.UserClaims) (ginx.Result, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ginx.Result{
			//ctx.JSON(http.StatusOK, )
			//a.l.Error("前端输入的 ID 不对", logger.Error(err))
			Code: 4,
			Msg:  "参数错误",
		}, err
	}
	art, err := a.svc.GetById(ctx, id)
	if err != nil {
		//ctx.JSON(http.StatusOK, )
		//a.l.Error("获得文章信息失败", logger.Error(err))
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	// 这是不借助数据库查询来判定的方法
	if art.Author.Id != usr.Id {
		//ctx.JSON(http.StatusOK)
		// 如果公司有风控系统，这个时候就要上报这种非法访问的用户了。
		//a.l.Error("非法访问文章，创作者 ID 不匹配",
		//	logger.Int64("uid", usr.Id))
		return ginx.Result{
			Code: 4,
			// 不需要告诉前端（违法操作人员）发生了什么
			Msg: "输入有误",
		}, fmt.Errorf("非法访问文章，创作者 ID 不匹配 %d", usr.Id) // 后端报错，指出谁在捣乱
	}
	return ginx.Result{
		Data: ArticleVo{
			Id:    art.Id,
			Title: art.Title,
			// 不需要这个摘要信息
			//Abstract: art.Abstract(),
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 这个是创作者看自己的文章，也不需要这个字段
			//Author: art.Author
			Ctime: art.Ctime.Format(time.DateTime),
			Utime: art.Utime.Format(time.DateTime),
		},
	}, nil
}

func (h *ArticleHandler) list(ctx *gin.Context, req ListReq, uc ijwt.UserClaims) (ginx.Result, error) {
	res, err := h.svc.List(ctx, uc.Id, req.Offset, req.Limit)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, nil
	}
	// 在列表页，不显示全文，只显示一个"摘要"
	// 比如说，简单的摘要就是前几句话
	// 强大的摘要是 AI 帮你生成的
	return ginx.Result{
		Data: slice.Map[domain.Article, ArticleVo](res, func(idx int, src domain.Article) ArticleVo {
			return ArticleVo{
				Id:       src.Id,
				Title:    src.Title,
				Abstract: src.Abstract(),
				Status:   src.Status.ToUint8(),
				// 这个列表请求，不需要返回内容
				//Content: src.Content,
				// 这个是创作者看自己的文章列表，也不需要这个字段
				//Author: src.Author
				Ctime: src.Ctime.Format(time.DateTime),
				Utime: src.Utime.Format(time.DateTime),
			}
		}),
	}, nil
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 4,
			Msg:  "参数错误",
		})
		h.l.Error("前端输入的 ID 不对", logger.Error(err))
		return
	}
	// errgroup.Group 与 go func() 相比
	// 自动处理错误和同步：其能够统一处理这些任务中的错误。它会在任务中的某个错误发生时终止所有并发任务，并且返回最先遇到的错误
	uc := ctx.MustGet("users").(ijwt.UserClaims)
	var eg errgroup.Group
	var art domain.Article
	eg.Go(func() error {
		art, err = h.svc.GetPublishedById(ctx, id, uc.Id)
		return err
	})
	var resp *intrv1.GetResponse
	eg.Go(func() error {
		// 要在这里获得这篇文章的计数
		// 这个地方可以容忍错误,计数有点偏差影响不大
		resp, err = h.intrSvc.Get(ctx, &intrv1.GetRequest{
			BizId: id,
			Biz:   h.biz,
			Uid:   uc.Id,
		})
		// 这种是容错的写法
		//if err != nil {
		//	// 记录日志
		//}
		//return nil
		return err
	})
	// 这里要等待，要保证前面两个执行完，拿到数据
	err = eg.Wait()
	if err != nil {
		// 代表查询出错了
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	// 增加阅读计数
	go func() {
		// 开一个 goroutine，异步去执行
		_, er := h.intrSvc.IncrReadCnt(ctx, &intrv1.IncrReadCntRequest{
			Biz:   h.biz,
			BizId: art.Id,
		})
		if er != nil {
			h.l.Error("增加阅读计数失败",
				logger.Error(er),
				logger.Int64("aid", art.Id))
		}
	}()

	if err != nil {
		h.l.Error("增加阅读次数失败", logger.Error(err), logger.Int64("aid", art.Id))
	}
	intr := resp.Intr
	ctx.JSON(http.StatusOK, ginx.Result{
		Data: ArticleVo{
			Id:      art.Id,
			Title:   art.Title,
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 要把作者信息带出去
			Author: art.Author.Name,
			Ctime:  art.Ctime.Format(time.DateTime),
			Utime:  art.Utime.Format(time.DateTime),
			// 交互信息
			Liked:      intr.Liked,
			Collected:  intr.Collected,
			ReadCnt:    intr.ReadCnt,
			LikeCnt:    intr.LikeCnt,
			CollectCnt: intr.CollectCnt,
		},
	})
}

func (h *ArticleHandler) Like(ctx *gin.Context, req LikeReq, uc ijwt.UserClaims) (ginx.Result, error) {
	var err error
	if req.Like {
		_, err = h.intrSvc.Like(ctx, &intrv1.LikeRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   uc.Id,
		})
	} else {
		_, err = h.intrSvc.CancelLike(ctx, &intrv1.CancelLikeRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   uc.Id,
		})
	}
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{Msg: "OK"}, nil
}
