package ioc

import (
	"gitee.com/geekbang/basic-go/webook/config"
	"github.com/redis/go-redis/v9"
)

func initRedis() redis.Cmdable {
	redisClient := redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
	return redisClient
}
