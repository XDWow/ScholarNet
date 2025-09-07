package ioc

import (
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"go.uber.org/zap"
)

// 这里可以选择不同的实现，来初始化
func InitLogger() logger.LoggerV1 {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}
