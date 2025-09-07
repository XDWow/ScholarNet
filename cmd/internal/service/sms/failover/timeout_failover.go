package failover

import (
	"context"
	"fmt"
	"github.com/XD/ScholarNet/cmd/internal/service/sms"
	"github.com/XD/ScholarNet/cmd/pkg/ratelimit"
	"sync/atomic"
)

type TimeoutFailoverSMSService struct {
	// 你的服务商
	svcs []sms.Service
	idx  int32
	// 连续超时的个数
	cnt int32
	// 阈值
	// 连续超时超过这个数字，就要切换
	threshold int32
	limiter   ratelimit.Limiter
}

func NewTimeoutFailoverSMSService(svcs []sms.Service, cnt int32, threshold int32, limiter ratelimit.Limiter) sms.Service {
	return &TimeoutFailoverSMSService{
		svcs:      svcs,
		idx:       0,
		cnt:       cnt,
		threshold: threshold,
		limiter:   limiter,
	}
}

func (t *TimeoutFailoverSMSService) Name() string {
	return "timeoutFailover"
}

func (t *TimeoutFailoverSMSService) Send(ctx context.Context,
	tpl string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)
	limited, err := t.limiter.Limit(ctx, fmt.Sprintf("sms:%s", t.svcs[idx].Name()))
	if err != nil {
		return fmt.Errorf("短信服务判断是否限流出现问题，%w", err)
	}
	if cnt > t.threshold || limited {
		// 这里要切换，新的下标，往后挪了一个
		newIdx := (idx + 1) % int32(len(t.svcs))
		// 条件更新
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			// 我成功往后挪了一位
			atomic.StoreInt32(&t.cnt, 0)
		}
		// else 就是出现并发，别人换成功了

		//idx = newIdx
		idx = atomic.LoadInt32(&t.idx)
	}

	svc := t.svcs[idx]
	//	带有超时设置的 context
	err = svc.Send(ctx, tpl, args, numbers...)
	switch err {
	case context.DeadlineExceeded:
		atomic.AddInt32(&t.cnt, 1)
		return err
	case nil:
		// 连续超时状态被打断了
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	default:
		// - 非超时，直接换下一个服务
		newIdx := (idx + 1) % int32(len(t.svcs))
		// 无条件更新
		atomic.StoreInt32(&t.idx, newIdx)
		return err
	}
}
