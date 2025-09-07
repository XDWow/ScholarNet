package repository

import (
	"context"
	"github.com/XD/ScholarNet/cmd/interactive/domain"
	"github.com/XD/ScholarNet/cmd/interactive/repository/cache"
	"github.com/XD/ScholarNet/cmd/interactive/repository/dao"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
)

//go:generate mockgen -source=./interactive.go -package=repomocks -destination=mocks/interactive.mock.go InteractiveRepository
type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error
	IncrLike(ctx context.Context, biz string, bizId int64, uid int64) error
	DecrLike(ctx context.Context, biz string, bizId int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, bizId int64, cid, uid int64) error
	GetByIds(ctx context.Context, biz string, bizIds []int64) ([]domain.Interactive, error)
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, bizId int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, bizId int64, uid int64) (bool, error)
	AddRecord(ctx context.Context, aid int64, uid int64) error
}

type CachedInteractiveRepository struct {
	cache cache.InteractiveCache
	dao   dao.InteractiveDAO
	l     logger.LoggerV1
}

func NewCachedInteractiveRepository(cache cache.InteractiveCache, dao dao.InteractiveDAO, l logger.LoggerV1) InteractiveRepository {
	return &CachedInteractiveRepository{
		cache: cache,
		dao:   dao,
		l:     l,
	}
}

func (c *CachedInteractiveRepository) AddRecord(ctx context.Context, aid int64, uid int64) error {
	//TODO implement me
	panic("implement me")
}

// BatchIncrReadCnt bizs 和 ids 的长度必须相等
func (repo *CachedInteractiveRepository) BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error {
	err := repo.dao.BatchIncrReadCnt(ctx, bizs, bizIds)
	// 你也要批量的去修改 redis，所以就要去改 lua 脚本
	// c.cache.IncrReadCntIfPresent()
	// TODO, 等我写新的 lua 脚本/或者用 pipeline
	return err
}

func (repo *CachedInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	// 要考虑缓存方案了
	// 这两个操作能不能换顺序？ —— 不能,因为数据库更重要，放前面，就算后面的缓存未更新，影响也不大，以数据库为准
	err := repo.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	//go func() {
	//	c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
	//}()
	//return err

	return repo.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (repo *CachedInteractiveRepository) IncrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	// 先插入点赞，然后更新点赞计数，更新缓存
	err := repo.dao.InsertLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	// 这种做法，你需要在 repository 层面上维持住事务，来保证插入点赞和更新点赞计数同时成功或失败
	//c.dao.IncrLikeCnt()
	return repo.cache.IncrLikeCntIfPresent(ctx, biz, bizId)
}

func (repo *CachedInteractiveRepository) DecrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	err := repo.dao.DeleteLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return repo.cache.DecrLikeCntIfPresent(ctx, biz, bizId)
}

func (repo *CachedInteractiveRepository) AddCollectionItem(ctx context.Context, biz string, bizId, cid, uid int64) error {
	// 这个地方，你要不要考虑缓存收藏夹？
	// 以及收藏夹里面的内容
	// 用户会频繁访问他的收藏夹，那么你就应该缓存，不然你就不需要
	// 一个东西要不要缓存，你就看用户会不会频繁访问（反复访问）
	err := repo.dao.InsertCollectionBiz(ctx, dao.UserCollectionBiz{
		Cid:   cid,
		Biz:   biz,
		BizId: bizId,
		Uid:   uid,
	})
	if err != nil {
		return err
	}
	return repo.cache.IncrCollectCntIfPresent(ctx, biz, bizId)
}

func (repo *CachedInteractiveRepository) GetByIds(ctx context.Context, biz string, bizIds []int64) ([]domain.Interactive, error) {
	vals, err := repo.dao.GetByIds(ctx, biz, bizIds)
	if err != nil {
		return nil, err
	}
	return slice.Map[dao.Interactive, domain.Interactive](vals,
		func(idx int, src dao.Interactive) domain.Interactive { return repo.toDomain(src) }), nil
}

func (repo *CachedInteractiveRepository) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	// 拿阅读数，点赞数和收藏数
	// 先从缓存拿
	intr, err := repo.cache.Get(ctx, biz, bizId)
	if err == nil {
		return intr, nil
	}
	// 但不是所有的结构体都是可比较的
	//if intr == (domain.Interactive{}) {
	//
	//}

	// 缓存没有，去数据库拿，并写回缓存
	daoIntr, err := repo.dao.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}
	intr = repo.toDomain(daoIntr)
	go func() {
		er := repo.cache.Set(ctx, biz, bizId, intr)
		if er != nil {
			repo.l.Error("回写缓存失败",
				logger.String("biz", biz),
				logger.Int64("bizId", bizId))
		}
	}()
	return intr, nil
}

func (repo *CachedInteractiveRepository) Liked(ctx context.Context, biz string, bizId int64, uid int64) (bool, error) {
	_, err := repo.dao.GetLikeInfo(ctx, biz, bizId, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrDataNotFound:
		return false, nil // 从来没点赞过，或者软删除了（取消点赞）
	// 这才是真正的 错误
	default:
		return false, err
	}
}

func (repo *CachedInteractiveRepository) Collected(ctx context.Context, biz string, bizId int64, uid int64) (bool, error) {
	_, err := repo.dao.GetCollectInfo(ctx, biz, bizId, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrDataNotFound:
		return false, nil
	// 这才是真正的 错误
	default:
		return false, err
	}
}

// 最简原则：
// 1. 接收器永远用指针
// 2. 输入输出都用结构体
func (repo *CachedInteractiveRepository) toDomain(intr dao.Interactive) domain.Interactive {
	return domain.Interactive{
		Biz:        intr.Biz,
		BizId:      intr.BizId,
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
		ReadCnt:    intr.ReadCnt,
	}
}
