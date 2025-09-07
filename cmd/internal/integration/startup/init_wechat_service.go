package startup

import (
	"github.com/XD/ScholarNet/cmd/internal/service/oauth2/wechat"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
)

// InitPhantomWechatService 没啥用的虚拟的 wechatService
func InitPhantomWechatService(l logger.LoggerV1) wechat.Service {
	return wechat.NewService("", "", l)
}
