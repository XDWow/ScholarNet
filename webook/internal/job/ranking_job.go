package job

import (
	"context"
	"github.com/LXD-c/basic-go/webook/internal/service"
	"github.com/LXD-c/basic-go/webook/pkg/logger"
	rlock "github.com/gotomicro/redis-lock"
	"sync"
	"time"
)

type RankingJob struct {
	svc       service.RankingService
	timeout   time.Duration
	client    *rlock.Client
	key       string
	l         logger.LoggerV1
	lock      *rlock.Lock
	localLock *sync.Mutex
}

func NewRankingJob(svc service.RankingService,
	client *rlock.Client,
	l logger.LoggerV1,
	timeout time.Duration) *RankingJob {
	// timeout 根据你的数据量来，如果要是七天内的帖子数量很多，你就要设置长一点
	return &RankingJob{
		svc:       svc,
		timeout:   timeout,
		client:    client,
		key:       "rlock:cron_job:ranking",
		l:         l,
		localLock: &sync.Mutex{},
	}
}

func (j *RankingJob) Name() string { return "ranking" }

// 按时间调度的，三分钟一次
// localLock：为了让本实例中只有一个线程在执行
func (j *RankingJob) Run() error {
	j.localLock.Lock()
	defer j.localLock.Unlock()
	if j.lock == nil {
		// 说明你没拿到锁，你得试着拿锁
		// 拿锁设置超时时间
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// 我可以设置一个比较短的过期时间
		lock, err := j.client.Lock(ctx, j.key, j.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      0,
		}, time.Second)
		if err != nil {
			return nil
		}
		j.lock = lock
		// 我怎么保证我这里，一直拿着这个锁？？？
		go func() {
			// 自动续约机制
			err1 := lock.AutoRefresh(j.timeout/2, time.Second)
			// 这里说明退出了续约机制
			// 续约失败了怎么办？
			if err1 != nil {
				// 不怎么办
				// 争取下一次，继续抢锁
				j.l.Error("续约失败", logger.Error(err))
			}
			j.localLock.Lock()
			j.lock = nil
			defer j.localLock.Unlock()
			// lock.Unlock(ctx)
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), j.timeout)
	defer cancel()
	return j.svc.TopN(ctx)
}

func (j *RankingJob) Close() error {
	j.localLock.Lock()
	lock := j.lock
	j.lock = nil
	j.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}
