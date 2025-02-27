package ioc

import (
	"github.com/LXD-c/basic-go/webook/internal/job"
	"github.com/LXD-c/basic-go/webook/internal/service"
	"github.com/LXD-c/basic-go/webook/pkg/logger"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/robfig/cron/v3"
	"time"
)

func InitRankingJob(svc service.RankingService, rlockClient *rlock.Client, l logger.LoggerV1) *job.RankingJob {
	return job.NewRankingJob(svc, rlockClient, l, time.Second*30)
}

// 之前的所有实现了一个分布式任务，现在用 cron 实现定时
func InitJobs(l logger.LoggerV1, rankingJob *job.RankingJob) *cron.Cron {
	res := cron.New(cron.WithSeconds())
	cbd := job.NewCronJobBuilder(l)
	// 三分钟执行一次，装饰器封装好的 rankingJob
	_, err := res.AddJob("@every 3min", cbd.Build(rankingJob))
	if err != nil {
		panic(err)
	}
	return res
}
