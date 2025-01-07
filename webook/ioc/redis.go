package ioc

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() redis.Cmdable {
	addr := viper.GetString("redis.addr")
	redisClient := redis.NewClient(&redis.Options{
		// 包变量--万恶之源
		//Addr: config.Config.Redis.Addr,
		Addr: addr,
	})
	return redisClient
}
