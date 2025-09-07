package domain

import "time"

// 所有事件的统一表示
type FeedEvent struct {
	// 公共部分
	ID int64
	// 以 A 发表了一篇文章为例
	// 如果是 Pull Event，也就是拉模型，那么 Uid 是 A 的id
	// 如果是 Push Event，也就是推模型，那么 Uid 是 A 的某个粉丝的 id
	Uid int64
	// 用来区分具体业务
	Type string
	// 这里有个 Ctime 是为了聚合排序用
	Ctime time.Time

	// 私有部分，直接 map[string]string
	Ext ExtendFields
}
