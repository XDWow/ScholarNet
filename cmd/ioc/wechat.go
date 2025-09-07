package ioc

import (
	"github.com/XD/ScholarNet/cmd/internal/service/oauth2/wechat"
	"github.com/XD/ScholarNet/cmd/internal/web"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
)

func InitWechatService(l logger.LoggerV1) wechat.Service {
	//appId, ok := os.LookupEnv("WECHAT_APP_ID")
	//if !ok {
	//	panic("没有找到环境变量 WECHAT_APP_ID ")
	//}
	//appKey, ok := os.LookupEnv("WECHAT_APP_SECRET")
	//if !ok {
	//	panic("没有找到环境变量 WECHAT_APP_SECRET")
	//}

	return wechat.NewService("1", "2", l)
}

func NewWechatHandlerConfig() web.WechatHandlerConfig {
	return web.WechatHandlerConfig{
		Secure: false,
	}
}
