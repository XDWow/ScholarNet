package service

import (
	"context"
	"github.com/LXD-c/basic-go/webook/interactive/domain"
	"github.com/LXD-c/basic-go/webook/interactive/repository"
	"github.com/LXD-c/basic-go/webook/pkg/logger"
	"golang.org/x/sync/errgroup"
)

type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, bizId int64, uid int64) error
	CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error
	// Collect 收藏, cid 是收藏夹的 ID
	// cid 不一定有，或者说 0 对应的是该用户的默认收藏夹
	Collect(ctx context.Context, biz string, bizId, cid, uid int64) error
	// 获取与交互相关的全部信息：点赞收藏浏览数量，以及某用户是否点赞收藏
	Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error)
	GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error)
}

type interactiveService struct {
	repo repository.InteractiveRepository
	l    logger.LoggerV1
}

func NewInteractiveService(repo repository.InteractiveRepository, l logger.LoggerV1) InteractiveService {
	return &interactiveService{
		repo: repo,
		l:    l,
	}
}

func (s *interactiveService) GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error) {
	intrs, err := s.repo.GetByIds(ctx, biz, bizIds)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]domain.Interactive, len(intrs))
	for _, intr := range intrs {
		res[intr.BizId] = intr
	}
	return res, nil
}

func (s *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return s.repo.IncrReadCnt(ctx, biz, bizId)
}

func (s *interactiveService) Like(ctx context.Context, biz string, bizId int64, uid int64) error {
	return s.repo.IncrLike(ctx, biz, bizId, uid)
}

func (s *interactiveService) CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	return s.repo.DecrLike(ctx, biz, bizId, uid)
}

func (s *interactiveService) Collect(ctx context.Context, biz string, bizId int64, cid, uid int64) error {
	return s.repo.AddCollectionItem(ctx, biz, bizId, cid, uid)
}

//func (s *interactiveService) Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error) {
//	// 按照 repository 的语义(完成 domain.Interactive 的完整构造)，你这里拿到的就应该是包含全部字段的
//	intr, err := s.repo.Get(ctx, biz, bizId)
//	if err != nil {
//		return domain.Interactive{}, err
//	}
//	//eg.Go(func() error {
//	//	var err error
//	//	intr, err = s.repo.Get(ctx, biz, bizId)
//	//	return err
//	//})
//	var eg errgroup.Group
//	eg.Go(func() error {
//		intr.Liked, err = s.repo.Liked(ctx, biz, bizId, uid)
//		return err
//	})
//	eg.Go(func() error {
//		intr.Collected, err = s.repo.Collected(ctx, biz, bizId, uid)
//		return err
//	})
//	if err = eg.Wait(); err != nil {
//		// 这个查询失败只需要记录日志就可以，不需要中断执行
//		s.l.Error("查询用户是否点赞的信息失败",
//			logger.String("biz", biz),
//			logger.Int64("bizId", bizId),
//			logger.Int64("uid", uid),
//			logger.Error(err))
//	}
//	return intr, nil
//}

func (s *interactiveService) Get(
	ctx context.Context, biz string,
	bizId, uid int64) (domain.Interactive, error) {
	// 你也可以考虑将分发的逻辑也下沉到 repository 里面
	intr, err := s.repo.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}
	var eg errgroup.Group
	eg.Go(func() error {
		intr.Liked, err = s.repo.Liked(ctx, biz, bizId, uid)
		return err
	})
	eg.Go(func() error {
		intr.Collected, err = s.repo.Collected(ctx, biz, bizId, uid)
		return err
	})
	// 说明是登录过的，补充用户是否点赞或者
	// 新的打印日志的形态 zap 本身就有这种用法
	err = eg.Wait()
	if err != nil {
		// 这个查询失败只需要记录日志就可以，不需要中断执行
		s.l.Error("查询用户是否点赞的信息失败",
			logger.String("biz", biz),
			logger.Int64("bizId", bizId),
			logger.Int64("uid", uid),
			logger.Error(err))
	}
	return intr, nil
}
