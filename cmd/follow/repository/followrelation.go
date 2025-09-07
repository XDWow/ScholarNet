package repository

import (
	"context"
	"github.com/XD/ScholarNet/cmd/follow/domain"
	"github.com/XD/ScholarNet/cmd/follow/repository/cache"
	"github.com/XD/ScholarNet/cmd/follow/repository/dao"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
)

type FollowRepository interface {
	// GetFollowee 获取某人的关注列表
	GetFollowee(ctx context.Context, follower, offset, limit int64) ([]domain.FollowRelation, error)
	// FollowInfo 查看关注的详情
	FollowInfo(ctx context.Context, follower int64, followee int64) (domain.FollowRelation, error)
	// AddFollowRelation 创建关注关系
	AddFollowRelation(ctx context.Context, f domain.FollowRelation) error
	// InactiveFollowRelation 取消关注
	InactiveFollowRelation(ctx context.Context, follower int64, followee int64) error
	GetFollowStatics(ctx context.Context, uid int64) (domain.FollowStatics, error)
}

type CachedRelationRepository struct {
	dao   dao.FollowRelationDao
	cache cache.FollowCache
	l     logger.LoggerV1
}

func NewFollowRepository(dao dao.FollowRelationDao,
	cache cache.FollowCache,
	l logger.LoggerV1) FollowRepository {
	return &CachedRelationRepository{dao: dao, cache: cache, l: l}
}

func (repo *CachedRelationRepository) GetFollowee(ctx context.Context, follower, offset, limit int64) ([]domain.FollowRelation, error) {
	// 你要做缓存，撑死了就是缓存第一页
	// 缓存命中率贼低
	followeeList, err := repo.dao.FollowRelationList(ctx, follower, offset, limit)
	if err != nil {
		return nil, err
	}
	return repo.genFollowRelationList(followeeList), nil
}

func (repo *CachedRelationRepository) genFollowRelationList(followerList []dao.FollowRelation) []domain.FollowRelation {
	res := make([]domain.FollowRelation, 0, len(followerList))
	for _, c := range followerList {
		res = append(res, repo.toDomain(c))
	}
	return res
}

func (repo *CachedRelationRepository) FollowInfo(ctx context.Context, follower int64, followee int64) (domain.FollowRelation, error) {
	// 要比列表有缓存价值
	c, err := repo.dao.FollowRelationDetail(ctx, follower, followee)
	if err != nil {
		return domain.FollowRelation{}, err
	}
	return repo.toDomain(c), nil
}

func (repo *CachedRelationRepository) AddFollowRelation(ctx context.Context, f domain.FollowRelation) error {
	err := repo.dao.CreateFollowRelation(ctx, repo.toEntity(f))
	if err != nil {
		return err
	}
	// 这里要更新在 Redis 上的缓存计数，对于 A 关注了 B 来说，这里要增加 A 的 followee 的数量
	// 同时要增加 B 的 follower 的数量
	return repo.cache.Follow(ctx, f.Follower, f.Followee)
}

func (repo *CachedRelationRepository) InactiveFollowRelation(ctx context.Context, follower int64, followee int64) error {
	err := repo.dao.UpdateStatus(ctx, follower, followee, dao.FollowRelationStatusInactive)
	if err != nil {
		return err
	}
	return repo.cache.CancelFollow(ctx, follower, followee)
}

func (repo *CachedRelationRepository) GetFollowStatics(ctx context.Context, uid int64) (domain.FollowStatics, error) {
	// 这个是经常要用的，可以缓存
	// 快路径
	res, err := repo.cache.StaticsInfo(ctx, uid)
	if err == nil {
		return res, nil
	}
	// 慢路径
	res.Followers, err = repo.dao.CntFollower(ctx, uid)
	if err != nil {
		return domain.FollowStatics{}, err
	}
	res.Followees, err = repo.dao.CntFollowee(ctx, uid)
	if err != nil {
		return domain.FollowStatics{}, err
	}
	err = repo.cache.SetStaticsInfo(ctx, uid, res)
	if err != nil { // 这里记录日志
		repo.l.Error("缓存关注统计信息失败",
			logger.Error(err),
			logger.Int64("uid", uid))
	}
	return res, nil
}

func (repo *CachedRelationRepository) toDomain(fr dao.FollowRelation) domain.FollowRelation {
	return domain.FollowRelation{
		Followee: fr.Followee,
		Follower: fr.Follower,
	}
}

func (repo *CachedRelationRepository) toEntity(c domain.FollowRelation) dao.FollowRelation {
	return dao.FollowRelation{
		Followee: c.Followee,
		Follower: c.Follower,
	}
}
