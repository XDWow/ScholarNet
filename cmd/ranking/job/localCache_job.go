package job

import (
	"context"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"github.com/XD/ScholarNet/cmd/ranking/repository"
	"time"
)

// LocalCacheRefreshJob 定时刷新本地缓存
// 每5分钟从Redis拉取前100名
type LocalCacheRefreshJob struct {
	repo   repository.RankingRepository
	logger logger.LoggerV1
	nodeID string
}

func NewLocalCacheRefreshJob(repo repository.RankingRepository, l logger.LoggerV1) *LocalCacheRefreshJob {
	return &LocalCacheRefreshJob{
		repo:   repo,
		logger: l,
	}
}

func (j *LocalCacheRefreshJob) Name() string { return "local_cache_refresh" }

func (j *LocalCacheRefreshJob) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := j.repo.RefreshLocalCache(ctx)
	if err != nil {
		j.logger.Error("刷新本地缓存失败", logger.Error(err))
		return err
	}
	j.logger.Info("本地缓存刷新成功")
	return nil
}
