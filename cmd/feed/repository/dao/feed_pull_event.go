package dao

import (
	"context"
	"gorm.io/gorm"
)

type FeedPullEventDAO interface {
	CreatePullEvent(ctx context.Context, event FeedPullEvent) error
	FindPullEventList(ctx context.Context, uids []int64, timestamp, limit int64) ([]FeedPullEvent, error)
	FindPullEventListWithType(ctx context.Context, typ string, uids []int64, timestamp, limit int64) ([]FeedPullEvent, error)
}

// FeedPullEvent 拉模型
// 目前我们的业务里面没明显区别
// 在实践中很可能会有区别
type FeedPullEvent struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 发件人
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

type GORMFeedPullEventDAO struct {
	db *gorm.DB
}

func (dao *GORMFeedPullEventDAO) FindPullEventListWithType(ctx context.Context, typ string,
	uids []int64, timestamp, limit int64) ([]FeedPullEvent, error) {
	var events []FeedPullEvent
	err := dao.db.WithContext(ctx).
		Where("type = ? AND uid in (?) AND ctime < ?", typ, uids, timestamp).Find(&events).
		Order("ctime desc").
		Limit(int(limit)).
		Find(&events).Error
	return events, err
}

func (dao *GORMFeedPullEventDAO) CreatePullEvent(ctx context.Context, event FeedPullEvent) error {
	return dao.db.WithContext(ctx).Create(&event).Error
}

func (dao *GORMFeedPullEventDAO) FindPullEventList(ctx context.Context, uids []int64, timestamp, limit int64) ([]FeedPullEvent, error) {
	var events []FeedPullEvent
	err := dao.db.WithContext(ctx).
		Where("uid in (?) AND ctime < ?", uids, timestamp).Find(&events).
		Order("ctime desc").
		Limit(int(limit)).
		Find(&events).Error
	return events, err
}

func NewGORMFeedPullEventDAO(db *gorm.DB) *GORMFeedPullEventDAO {
	return &GORMFeedPullEventDAO{db: db}
}
