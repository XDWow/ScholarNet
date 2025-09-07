package web

import (
	articlev1 "github.com/XD/ScholarNet/cmd/api/proto/gen/article/v1"
	rewardv1 "github.com/XD/ScholarNet/cmd/api/proto/gen/reward/v1"
	"github.com/XD/ScholarNet/cmd/pkg/ginx"
	"github.com/gin-gonic/gin"
)

type RewardHandler struct {
	client    rewardv1.RewardServiceClient
	artClient articlev1.ArticleServiceClient
}

func NewRewardHandler(client rewardv1.RewardServiceClient, artClient articlev1.ArticleServiceClient) *RewardHandler {
	return &RewardHandler{client: client, artClient: artClient}
}

func (h *RewardHandler) RegisterRoutes(server *gin.Engine) {
	rg := server.Group("/reward")
	rg.POST("/detail",
		ginx.WrapClaimsAndReq[GetRewardReq](h.GetReward))
}

type GetRewardReq struct {
	Rid int64
}

func (h *RewardHandler) GetReward(
	ctx *gin.Context,
	req GetRewardReq,
	claims ginx.UserClaims) (ginx.Result, error) {
	resp, err := h.client.GetReward(ctx, &rewardv1.GetRewardRequest{
		Rid: req.Rid,
		Uid: claims.Id,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		// 暂时也就是只需要状态
		Data: resp.Status.String(),
	}, nil
}

type RewardArticleReq struct {
	Aid int64 `json:"aid"`
	Amt int64 `json:"amt"`
}
