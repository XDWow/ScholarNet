//go:build wireinject

package startup

import (
	pmtv1 "github.com/XD/ScholarNet/cmd/api/proto/gen/payment/v1"
	"github.com/XD/ScholarNet/cmd/reward/repository"
	"github.com/XD/ScholarNet/cmd/reward/repository/cache"
	"github.com/XD/ScholarNet/cmd/reward/repository/dao"
	"github.com/XD/ScholarNet/cmd/reward/service"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(InitTestDB, InitLogger, InitRedis)

func InitWechatNativeSvc(client pmtv1.WechatPaymentServiceClient) *service.WechatNativeRewardService {
	wire.Build(service.NewWechatNativeRewardService,
		thirdPartySet,
		cache.NewRewardRedisCache,
		repository.NewRewardRepository, dao.NewRewardGORMDAO)
	return new(service.WechatNativeRewardService)
}
