package service

import (
	"context"
	"github.com/LXD-c/basic-go/webook/internal/domain"
	"github.com/LXD-c/basic-go/webook/internal/repository"
	"github.com/LXD-c/basic-go/webook/pkg/logger"
	"golang.org/x/sync/errgroup"
)
//go:generate mockgen -source=./interactive.go -package=svcmocks -destination=mocks/interactive.mock.go InteractiveService
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

type InteractiveServiceImpl struct {
	repo repository.InteractiveRepository
	l    logger.LoggerV1
}

func NewInteractiveServiceImpl(repo repository.InteractiveRepository, l logger.LoggerV1) InteractiveService {
	return &InteractiveServiceImpl{
		repo: repo,
		l:    l,
	}
}

func (s *InteractiveServiceImpl) GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error) {
	//TODO implement me
	panic("implement me")
}

func (s *InteractiveServiceImpl) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return s.repo.IncrReadCnt(ctx, biz, bizId)
}

func (s *InteractiveServiceImpl) Like(ctx context.Context, biz string, bizId int64, uid int64) error {
	return s.repo.IncrLike(ctx, biz, bizId, uid)
}

func (s *InteractiveServiceImpl) CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	return s.repo.DecrLike(ctx, biz, bizId, uid)
}

func (s *InteractiveServiceImpl) Collect(ctx context.Context, biz string, bizId int64, cid, uid int64) error {
	return s.repo.AddCollectionItem(ctx, biz, bizId, cid, uid)
}

func (s *InteractiveServiceImpl) Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error) {
	// 按照 repository 的语义(完成 domain.Interactive 的完整构造)，你这里拿到的就应该是包含全部字段的
	var (
		eg        errgroup.Group
		intr      domain.Interactive
		liked     bool
		collected bool
	)
	eg.Go(func() error {
		var err error
		intr, err = s.repo.Get(ctx, biz, bizId)
		return err
	})
	eg.Go(func() error {
		var err error
		liked, err = s.repo.Liked(ctx, biz, bizId, uid)
		return err
	})
	eg.Go(func() error {
		var err error
		collected, err = s.repo.Collected(ctx, biz, bizId, uid)
		return err
	})
	if err := eg.Wait(); err != nil {
		return domain.Interactive{}, err
	}
	intr.Liked = liked
	intr.Collected = collected
	return intr, nil
}
