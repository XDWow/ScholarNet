package dao

import (
	"context"
	"gorm.io/gorm"
)

type FeedPushEventDAO interface {
	CreatePushEvents(ctx context.Context, events []FeedPushEvent) error
	GetPushEvents(ctx context.Context, uid int64, timestamp, limit int64) ([]FeedPushEvent, error)
	GetPushEventsWithTyp(ctx context.Context, typ string, uid int64, timestamp, limit int64) ([]FeedPushEvent, error)
}

// FeedPushEvent 写扩散，推模型，收件箱
// 这个表理论上是只插入，不更新，也不删除的
// 但是可以归档
type FeedPushEvent struct {
	Id  int64 `gorm:"primary_key;auto_increment"`
	UID int64 `gorm:"index"`
	// Type 用来标记是什么类型的事件
	// 这边决定了 Content 怎么解读
	Type string
	// 大的 json 串
	Content string
	Ctime   int64 `gorm:"index"`
	// 这个表理论上来说，是没有 Update 操作的
	Utime int64
}

type feedPushEventDAO struct {
	db *gorm.DB
}

func (dao *feedPushEventDAO) GetPushEventsWithTyp(ctx context.Context,
	typ string, uid int64, timestamp, limit int64) ([]FeedPushEvent, error) {
	var res []FeedPushEvent
	err := dao.db.WithContext(ctx).Where("type = ? AND uid = ? AND ctime < ?", typ, uid, timestamp).
		Order("ctime desc").Limit(int(limit)).Find(&res).Error
	return res, err
}

func (dao *feedPushEventDAO) CreatePushEvents(ctx context.Context, events []FeedPushEvent) error {
	return dao.db.WithContext(ctx).Create(&events).Error
}

func (dao *feedPushEventDAO) GetPushEvents(ctx context.Context, uid int64, timestamp, limit int64) ([]FeedPushEvent, error) {
	var res []FeedPushEvent
	err := dao.db.Where("uid = ? AND ctime < ?", uid, timestamp).
		Order("ctime desc").Limit(int(limit)).Find(&res).Error
	return res, err
}

func NewFeedPushEventDAO(db *gorm.DB) FeedPushEventDAO {
	return &feedPushEventDAO{
		db: db,
	}
}
