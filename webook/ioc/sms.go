package ioc

import (
	"gitee.com/geekbang/basic-go/webook/internal/service/sms"
	"gitee.com/geekbang/basic-go/webook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	// 具体实现随便你换
	return memory.NewService()
}
