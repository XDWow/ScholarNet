package ioc

import (
	"context"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"github.com/XD/ScholarNet/cmd/tag/repository"
	"github.com/XD/ScholarNet/cmd/tag/repository/cache"
	"github.com/XD/ScholarNet/cmd/tag/repository/dao"
	"time"
)

func InitRepository(d dao.TagDAO, c cache.TagCache, l logger.LoggerV1) repository.TagRepository {
	repo := repository.NewTagRepository(d, c, l)
	go func() {
		// 执行缓存预加载
		// 或者启动的环境变量
		// 启动参数控制
		// 或者借助配置中心的开关
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		// 也可以同步执行。但是在一些场景下，同步执行会占用很长的时间，所以可以考虑异步执行。
		repo.PreloadUserTags(ctx)
	}()
	return repo
}
