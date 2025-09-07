package startup

import (
	"github.com/XD/ScholarNet/cmd/pkg/logger"
)

func InitLog() logger.LoggerV1 {
	return &logger.NopLogger{}
}
