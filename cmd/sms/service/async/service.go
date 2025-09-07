package async

import (
	"context"
	"fmt"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"github.com/XD/ScholarNet/cmd/pkg/ratelimit"
	"github.com/XD/ScholarNet/cmd/sms/domain"
	"github.com/XD/ScholarNet/cmd/sms/repository"
	"github.com/XD/ScholarNet/cmd/sms/service"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
	"time"
)

const key = "sms"

// 业务方需要线下申请 token，可以指定服务商重试策略：放入metadata
type Service struct {
	svcs map[string]service.Service
	// 转异步，存储发短信请求的 repository
	repo    repository.AsyncSmsRepository
	l       logger.LoggerV1
	key     []byte
	limiter ratelimit.Limiter
}

func NewService(svcs map[string]service.Service,
	repo repository.AsyncSmsRepository,
	l logger.LoggerV1, key []byte, limiter ratelimit.Limiter) *Service {
	res := &Service{
		svcs:    svcs,
		repo:    repo,
		l:       l,
		key:     key,
		limiter: limiter,
	}
	go func() {
		res.StartAsyncCycle()
	}()
	return res
}

// StartAsyncCycle 异步发送消息
// 这里我们没有设计退出机制，是因为没啥必要
// 因为程序停止的时候，它自然就停止了
// 原理：这是最简单的抢占式调度
func (s *Service) StartAsyncCycle() {
	for {
		s.AsyncSend()
	}
}

func (s *Service) AsyncSend() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// 抢占一个异步发送的消息，确保在非常多个实例
	// 比如 k8s 部署了三个 pod，一个请求，只有一个实例能拿到
	as, err := s.repo.PreemptWaitingSMS(ctx)
	cancel()
	switch err {
	case nil:
		// 执行发送
		svc, ok := s.svcs[as.Strategy]
		if !ok {
			svc = s.svcs["error"]
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err = svc.Send(ctx, as.TplId, as.Args, as.Numbers...)
		if err != nil {
			// 啥也不需要干
			s.l.Error("执行异步发送短信失败",
				logger.Error(err),
				logger.Int64("id", as.Id))
		}
		res := err == nil
		// 通知 repository 我这一次的执行结果
		err = s.repo.ReportScheduleResult(ctx, as.Id, res)
		if err != nil {
			s.l.Error("执行异步发送短信成功，但是标记数据库失败",
				logger.Error(err),
				logger.Bool("res", res),
				logger.Int64("id", as.Id))
		}
	case repository.ErrWaitingSMSNotFound:
		// 睡一秒。这个你可以自己决定
		time.Sleep(time.Second)
	default:
		// 正常来说应该是数据库那边出了问题，
		// 但是为了尽量运行，还是要继续的
		// 你可以稍微睡眠，也可以不睡眠
		// 睡眠的话可以帮你规避掉短时间的网络抖动问题
		s.l.Error("抢占异步发送短信任务失败",
			logger.Error(err))
		time.Sleep(time.Second)
	}
}

func (s *Service) Send(ctx context.Context, tplToken string, args []string, numbers ...string) error {
	// 鉴权
	var c Claims
	_, err := jwt.ParseWithClaims(tplToken, &c, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	// 这里可以拿到biz去数据库中查，这个业务部门还剩多少条短信，没了就返回错误：短信数量不足

	// 从 Metadata 拿数据
	bizType := getHeader(ctx, "x-biz-type", "")
	bizID := getHeader(ctx, "x-biz-id", "")
	if bizType == "" || bizID == "" {
		return fmt.Errorf("biz_type or biz_id not provided")
	}
	strategy := getHeader(ctx, "x-retry-strategy", "error")
	err = s.repo.Add(ctx, domain.AsyncSms{
		TplId:   c.Tpl,
		Args:    args,
		Numbers: numbers,
		// 设置可以重试三次
		RetryMax: 3,
		Strategy: strategy,
		BizType:  bizType,
		BizID:    bizID,
	})

	flag, _ := s.needAsync(ctx)
	if !flag {
		svc, ok := s.svcs[strategy]
		if !ok {
			//return errors.New("服务商重试策略不存在")，没必要返回错误，换个策略问题不大
			svc = s.svcs["error"]
		}
		return svc.Send(ctx, c.Tpl, args, numbers...)
	}
	return nil
}

// 提前引导你们，开始思考系统容错问题
// 你们面试装逼，赢得竞争优势就靠这一类的东西
// 在这里判断负载，负载高就返回 true,否则返回 false
func (s *Service) needAsync(ctx context.Context) (bool, error) {
	// 这边就是你要设计的，各种判定要不要触发异步的方案
	// 1. 基于响应时间的，平均响应时间
	// 1.1 使用绝对阈值，比如说直接发送的时候，（连续一段时间，或者连续N个请求）响应时间超过了 500ms，然后后续请求转异步
	// 1.2 变化趋势，比如说当前一秒钟内的所有请求的响应时间比上一秒钟增长了 X%，就转异步
	// 2. 基于错误率：一段时间内，收到 err 的请求比率大于 X%，转异步

	// 什么时候退出异步
	// 1. 进入异步 N 分钟后
	// 2. 保留 1% 的流量（或者更少），继续同步发送，判定响应时间/错误率

	// 限流
	limited, err := s.limiter.Limit(ctx, key)
	if err != nil {
		return false, fmt.Errorf("短信服务判断是否限流异常 %w", err)
	}
	if limited {
		return true, nil
	}
	// 降级
	if ctx.Value("") == true {
		return true, nil
	}
	return false, nil
}

// getHeader 从 ctx.Metadata 里拿第一个值，
// 如果不存在则返回 defaultVal
func getHeader(ctx context.Context, key, defaultVal string) string {
	md, _ := metadata.FromIncomingContext(ctx)
	if vals := md.Get(key); len(vals) > 0 {
		return vals[0]
	}
	return defaultVal
}

type Claims struct {
	jwt.RegisteredClaims
	Tpl string
	biz string
}
