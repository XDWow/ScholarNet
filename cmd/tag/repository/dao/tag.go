package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type Tag struct {
	Id    int64  `gorm:"primary_key;AUTO_INCREMENT"`
	Name  string `gorm:"type:varchar(4096);not null"`
	Uid   int64
	Ctime int64
	Utime int64
}

// 某个人对某个资源打了标签
type TagBiz struct {
	Id    int64 `gorm:"primary_key;AUTO_INCREMENT"`
	Biz   string
	BizId int64
	// 冗余字段，加快查询和删除
	// 这个字段可以删除的
	Uid int64 `gorm:"index"`
	//Tag Tag 没必要，太浪费空间了
	Tid   []int64
	Tag   []*Tag `gorm:"ForeignKey:Tid;AssociationForeignKey:Id;constraint:OnDelete:CASCADE"`
	Ctime int64  `bson:"ctime,omitempty"`
	Utime int64  `bson:"utime,omitempty"`
}

type TagDAO interface {
	CreateTag(ctx context.Context, tag Tag) (int64, error)
	//CreateTagBiz(ctx context.Context, tagBiz []TagBiz) error
	CreateTagBiz(ctx context.Context, tagBiz TagBiz) error
	GetTagsByUid(ctx context.Context, uid int64) ([]Tag, error)
	GetTagsByBiz(ctx context.Context, uid int64, biz string, bizId int64) ([]Tag, error)
	GetTags(ctx context.Context, offset, limit int) ([]Tag, error)
	GetTagsById(ctx context.Context, ids []int64) ([]Tag, error)
	DeleteTagBiz(ctx context.Context, uid int64, biz string, bizId int64) error
}

type GORMTagDAO struct {
	db *gorm.DB
}

func (dao *GORMTagDAO) CreateTag(ctx context.Context, tag Tag) (int64, error) {
	err := dao.db.WithContext(ctx).Create(&tag).Error
	if err != nil {
		return 0, err
	}
	return tag.Id, err
}

func (dao *GORMTagDAO) CreateTagBiz(ctx context.Context, tagBiz TagBiz) error {
	now := time.Now().UnixMilli()
	tagBiz.Ctime = now
	tagBiz.Utime = now
	return dao.db.WithContext(ctx).Create(&tagBiz).Error
}

func (dao *GORMTagDAO) DeleteTagBiz(ctx context.Context, uid int64, biz string, bizId int64) error {
	return dao.db.WithContext(ctx).Model(&TagBiz{}).
		Delete("uid = ? AND biz = ? AND biz_id = ?", uid, biz, bizId).Error
}

func (dao *GORMTagDAO) GetTagsByUid(ctx context.Context, uid int64) ([]Tag, error) {
	var res []Tag
	err := dao.db.WithContext(ctx).
		Where("uid = ?", uid).Find(&res).Error
	return res, err
}

func (dao *GORMTagDAO) GetTagsByBiz(ctx context.Context, uid int64, biz string, bizId int64) ([]Tag, error) {
	var TagBizs TagBiz
	err := dao.db.WithContext(ctx).
		Where("uid = ? AND biz = ? AND biz_id = ?", uid, biz, bizId).
		Find(&TagBizs).Error
	if err != nil {
		return nil, err
	}
	var res []Tag
	err = dao.db.WithContext(ctx).Where("id IN (?)", TagBizs.Tid).Find(&res).Error
	return res, err
}

func (dao *GORMTagDAO) GetTags(ctx context.Context, offset, limit int) ([]Tag, error) {
	var res []Tag
	err := dao.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

func (dao *GORMTagDAO) GetTagsById(ctx context.Context, ids []int64) ([]Tag, error) {
	var res []Tag
	err := dao.db.WithContext(ctx).Where("id IN ?", ids).Find(&res).Error
	return res, err
}

func NewGORMTagDAO(db *gorm.DB) TagDAO {
	return &GORMTagDAO{db: db}
}
