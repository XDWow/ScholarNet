//go:build wireinject

package startup

import (
	repository2 "github.com/XD/ScholarNet/cmd/interactive/repository"
	cache2 "github.com/XD/ScholarNet/cmd/interactive/repository/cache"
	dao2 "github.com/XD/ScholarNet/cmd/interactive/repository/dao"
	"github.com/XD/ScholarNet/cmd/interactive/service"
	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(InitRedis,
	InitTestDB, InitLogger)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository2.NewCachedInteractiveRepository,
	dao2.NewGORMInteractiveDAO,
	cache2.NewRedisInteractiveCache,
)

func InitInteractiveService() service.InteractiveService {
	wire.Build(thirdProvider, interactiveSvcProvider)
	return service.NewInteractiveService(nil, nil)
}
