package ioc

import (
	"github.com/XD/ScholarNet/cmd/internal/service/sms"
	"github.com/XD/ScholarNet/cmd/internal/service/sms/memory"
	"github.com/redis/go-redis/v9"
)

func InitSMSService(cmd redis.Cmdable) sms.Service {
	// 具体实现随便你换
	return memory.NewService()
}
