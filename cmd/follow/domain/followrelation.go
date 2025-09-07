package domain

// FollowRelation
type FollowRelation struct {
	// 被关注的人
	Followee int64
	// 关注的人
	Follower int64
	// 根据你的业务需要，你可以在这里加字段
	// 比如说备注啊，标签啊之类的,关注时间
	// Note string

}

// 用户的关注状态
type FollowStatics struct {
	// 被多少人关注
	Followers int64
	// 自己关注了多少人
	Followees int64
}
