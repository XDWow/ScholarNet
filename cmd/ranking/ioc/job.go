package ioc

import (
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"github.com/XD/ScholarNet/cmd/ranking/job"
	"github.com/XD/ScholarNet/cmd/ranking/repository"
	"github.com/XD/ScholarNet/cmd/ranking/service"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/robfig/cron/v3"
	"time"
)

func InitRankingJob(svc service.RankingService, rlockClient *rlock.Client, l logger.LoggerV1) *job.RankingJob {
	return job.NewRankingJob(svc, rlockClient, l, time.Second*30)
}

func InitLocalCacheRefreshJob(repo repository.RankingRepository, l logger.LoggerV1) *job.LocalCacheRefreshJob {
	return job.NewLocalCacheRefreshJob(repo, l)
}

// RankingJob 实现了一个分布式任务，现在用 cron 实现定时
// 所有任务都在这里初始化
func InitJobs(l logger.LoggerV1, rankingJob *job.RankingJob, localCacheJob *job.LocalCacheRefreshJob) *cron.Cron {
	res := cron.New(cron.WithSeconds())
	cbd := job.NewCronJobBuilder(l)
	// 三分钟执行一次，装饰器封装好的 rankingJob
	//_, err := res.AddJob("@every 3min", cbd.Build(rankingJob))
	_, err := res.AddJob("0 */3 * * * ?", cbd.Build(rankingJob))
	if err != nil {
		panic(err)
	}
	// 每5分钟刷新一次本地缓存
	_, err = res.AddJob("0 */5 * * * ?", cbd.Build(localCacheJob))
	if err != nil {
		panic(err)
	}

	// 缓存预热
	go func() {
		if err = localCacheJob.Run(); err != nil {
			l.Warn("启动时预热失败，已注册定时任务会继续刷新", logger.Error(err))
		}
	}()

	return res
}
