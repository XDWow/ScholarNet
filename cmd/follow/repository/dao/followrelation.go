package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type GORMFollowRelationDAO struct {
	db *gorm.DB
}

func NewFollowRelationDao(db *gorm.DB) FollowRelationDao {
	return &GORMFollowRelationDAO{db: db}
}

func (dao *GORMFollowRelationDAO) FollowRelationList(ctx context.Context, follower, offset, limit int64) ([]FollowRelation, error) {
	var res []FollowRelation
	err := dao.db.WithContext(ctx).
		// 这个查询要求我们要在 follower 上创建一个索引，或者 <follower, followee> 联合唯一索引
		// 进一步考虑，将 status 也加入索引:<follower, status, followee>
		Where("follower = ? AND status = ?", follower, FollowRelationStatusActive).
		Offset(int(offset)).Limit(int(limit)).Find(&res).Error
	return res, err
}

func (dao *GORMFollowRelationDAO) FollowRelationDetail(ctx context.Context, follower int64, followee int64) (FollowRelation, error) {
	var res FollowRelation
	err := dao.db.WithContext(ctx).
		Where("follower = ? AND followee = ? AND status = ?", follower, followee, FollowRelationStatusActive).
		First(&res).Error
	return res, err
}

func (dao *GORMFollowRelationDAO) CreateFollowRelation(ctx context.Context, f FollowRelation) error {
	now := time.Now().UnixMilli()
	f.Utime = now
	f.Ctime = now
	f.Status = FollowRelationStatusActive
	return dao.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"utime":  now,
				"status": FollowRelationStatusActive,
			}),
		}).Create(&f).Error
}

func (dao *GORMFollowRelationDAO) UpdateStatus(ctx context.Context, follower, followee int64, status uint8) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).
		Where("follower = ? AND followee = ?", follower, followee).
		Updates(map[string]interface{}{
			"status": status,
			"utime":  now,
		}).Error
}

func (dao *GORMFollowRelationDAO) CntFollower(ctx context.Context, uid int64) (int64, error) {
	var res int64
	// 这样你就慢了
	//err := dao.db.WithContext(ctx).
	//	Where("follower = ? AND status = ?", uid, FollowRelationStatusActive).
	//	Count(&res).Error
	err := dao.db.WithContext(ctx).
		Select("count(follower)").
		// 我这个怎么办？
		// 考虑在 followee 上创建一个索引（followee, status)
		// <followee, follower> 行不行？没必要，<follower, followee>的作用不仅加快查找，还有唯一限制的作用，
		// 你已经有唯一限制了，没必要再 <followee, follower>
		Where("followee = ? AND status = ?", uid, FollowRelationStatusActive).Count(&res).Error
	return res, err
}

func (dao *GORMFollowRelationDAO) CntFollowee(ctx context.Context, uid int64) (int64, error) {
	var res int64
	err := dao.db.WithContext(ctx).
		Select("count(followee)").
		// 我在这里，能利用到 <follower, followee> 的联合唯一索引
		Where("follower = ? AND status = ?",
			uid, FollowRelationStatusActive).Count(&res).Error
	return res, err
}
