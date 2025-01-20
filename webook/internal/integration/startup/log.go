package startup

import (
	"github.com/LXD-c/basic-go/webook/pkg/logger"
)

func InitLog() logger.LoggerV1 {
	return &logger.NopLogger{}
}
