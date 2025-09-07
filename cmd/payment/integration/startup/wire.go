//go:build wireinject

package startup

import (
	"github.com/XD/ScholarNet/cmd/payment/ioc"
	"github.com/XD/ScholarNet/cmd/payment/repository"
	"github.com/XD/ScholarNet/cmd/payment/repository/dao"
	"github.com/XD/ScholarNet/cmd/payment/service/wechat"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(ioc.InitLogger, InitTestDB)

var wechatNativeSvcSet = wire.NewSet(
	ioc.InitWechatClient,
	dao.NewPaymentGORMDAO,
	repository.NewPaymentRepository,
	ioc.InitWechatNativeService,
	ioc.InitWechatConfig)

func InitWechatNativeService() *wechat.NativePaymentService {
	wire.Build(wechatNativeSvcSet, thirdPartySet)
	return new(wechat.NativePaymentService)
}
