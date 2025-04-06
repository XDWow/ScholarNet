package grpc

import (
	"context"
	intrv1 "github.com/LXD-c/basic-go/webook/api/proto/gen/intr/v1"
	"github.com/LXD-c/basic-go/webook/interactive/domain"
	"github.com/LXD-c/basic-go/webook/interactive/service"
)

// InteractiveServiceServer 我这里只是把 service 包装成一个 grpc 而已
// 和 grpc 有关的操作，就限定在这里
type InteractiveServiceServer struct {
	intrv1.UnimplementedInteractiveServiceServer
	// 注意，核心业务逻辑一定是在 service 里面的
	svc service.InteractiveService
}

func (i *InteractiveServiceServer) IncrReadCnt(ctx context.Context, request *intrv1.IncrReadCntRequest) (*intrv1.IncrReadCntResponse, error) {
	err := i.svc.IncrReadCnt(ctx, request.GetBiz(), request.GetBizId())
	return &intrv1.IncrReadCntResponse{}, err
}

func (i *InteractiveServiceServer) Like(ctx context.Context, request *intrv1.LikeRequest) (*intrv1.LikeResponse, error) {
	err := i.svc.Like(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	return &intrv1.LikeResponse{}, err
}

func (i *InteractiveServiceServer) CancelLike(ctx context.Context, request *intrv1.CancelLikeRequest) (*intrv1.CancelLikeResponse, error) {
	err := i.svc.CancelLike(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	return &intrv1.CancelLikeResponse{}, err
}

func (i *InteractiveServiceServer) Collect(ctx context.Context, request *intrv1.CollectRequest) (*intrv1.CollectResponse, error) {
	err := i.svc.Collect(ctx, request.GetBiz(), request.GetBizId(), request.GetCid(), request.GetUid())
	return &intrv1.CollectResponse{}, err
}

func (i *InteractiveServiceServer) Get(ctx context.Context, request *intrv1.GetRequest) (*intrv1.GetResponse, error) {
	interactive, err := i.svc.Get(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	return &intrv1.GetResponse{
		Intr: i.toDTO(interactive),
	}, err
}

func (i *InteractiveServiceServer) GetByIds(ctx context.Context, request *intrv1.GetByIdsRequest) (*intrv1.GetByIdsResponse, error) {
	mp, err := i.svc.GetByIds(ctx, request.GetBiz(), request.GetBizIds())
	if err != nil {
		return nil, err
	}
	res := make(map[int64]*intrv1.Interactive, len(mp))
	for k, v := range mp {
		res[k] = i.toDTO(v)
	}
	return &intrv1.GetByIdsResponse{
		Intrs: res,
	}, nil
}

// DTO data transfer object
func (i *InteractiveServiceServer) toDTO(interactive domain.Interactive) *intrv1.Interactive {
	return &intrv1.Interactive{
		Biz:        interactive.Biz,
		BizId:      interactive.BizId,
		CollectCnt: interactive.CollectCnt,
		ReadCnt:    interactive.ReadCnt,
		LikeCnt:    interactive.LikeCnt,
		Liked:      interactive.Liked,
		Collected:  interactive.Collected,
	}
}
