package cache

import (
	"context"
	"fmt"
	"github.com/XD/ScholarNet/cmd/follow/domain"
	"github.com/redis/go-redis/v9"
	"strconv"
)

const (
	// 被多少人关注
	fieldFollowerCnt = "follower_cnt"
	// 关注了多少人
	fieldFolloweeCnt = "followee_cnt"
)

type RedisFollowCache struct {
	client redis.Cmdable
}

func NewRedisFollowCache(client redis.Cmdable) FollowCache {
	return &RedisFollowCache{
		client: client,
	}
}

func (r *RedisFollowCache) Follow(ctx context.Context, follower, followee int64) error {
	return r.updateStaticsInfo(ctx, follower, followee, 1)
}

func (r *RedisFollowCache) CancelFollow(ctx context.Context, follower, followee int64) error {
	return r.updateStaticsInfo(ctx, follower, followee, -1)
}

// 我现在要更新数量了
// 这个地方必要根本没有必要非得保持一起成功或者一起失败
func (r *RedisFollowCache) updateStaticsInfo(ctx context.Context, follower, followee int64, delta int64) error {
	tx := r.client.TxPipeline()
	// 理论上你应该要做到一起成功或者一起失败，
	// 用 lua 脚本可以
	// 首先更新 follower 的 followee 数量
	// 我往这个 tx 里面增加了两个指令，Tx 只是记录了，还没发过去 redis 服务端
	tx.HIncrBy(ctx, r.staticsKey(follower), fieldFolloweeCnt, delta)
	// 我是 followee（被关注的人），我的 fieldFollowerCnt（粉丝数量） 增加了
	tx.HIncrBy(ctx, r.staticsKey(followee), fieldFollowerCnt, delta)
	// Exec 的时候，会把两条命令发过去 redis server 上，并且这两条命令会一起执行
	// 中间不会有别的命令执行
	// 问题来了，有没有可能，执行了第一条命令成功，但是没有执行第二条？
	// 当然可能，因为没有 redolog,undolog 什么的，Redis 的事务，不具备 ACID 的特性
	_, err := tx.Exec(ctx)
	return err
}

func (r *RedisFollowCache) StaticsInfo(ctx context.Context, uid int64) (domain.FollowStatics, error) {
	data, err := r.client.HGetAll(ctx, r.staticsKey(uid)).Result()
	if err != nil {
		return domain.FollowStatics{}, err
	}
	if len(data) == 0 {
		return domain.FollowStatics{}, nil
	}
	// 理论上来说，这里不可能有 error
	followerCnt, err := strconv.ParseInt(data[fieldFollowerCnt], 10, 64)
	followeeCnt, _ := strconv.ParseInt(data[fieldFolloweeCnt], 10, 64)
	return domain.FollowStatics{
		Followees: followeeCnt,
		Followers: followerCnt,
	}, nil
}

func (r *RedisFollowCache) SetStaticsInfo(ctx context.Context, uid int64, statics domain.FollowStatics) error {
	key := r.staticsKey(uid)
	return r.client.HMSet(ctx, key, fieldFolloweeCnt, statics.Followees, fieldFollowerCnt, statics.Followers).Err()
}

func (r *RedisFollowCache) staticsKey(uid int64) string {
	return fmt.Sprintf("follow:statics:%d", uid)
}
