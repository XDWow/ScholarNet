package ratelimit

import (
	"context"
	"fmt"
	"github.com/XD/ScholarNet/cmd/internal/service/sms"
	"github.com/XD/ScholarNet/cmd/pkg/ratelimit"
)

// 这里定义错误的原因是：可以通过改变首字母大小写，随时决定是否将该错误暴露出去
var errLimited = fmt.Errorf("触发了限流")

type RatelimitSMSService struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewRatelimitSMSService(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &RatelimitSMSService{
		svc:     svc,
		limiter: limiter,
	}
}

// 装饰器模式
func (s *RatelimitSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	limited, err := s.limiter.Limit(ctx, "sms:tencent")
	if err != nil {
		return fmt.Errorf("短信服务判断是否限流出现问题，%w", err)
	}
	if limited {
		return errLimited
	}
	// 在这前面加一些代码，新特性
	err = s.svc.Send(ctx, tpl, args, numbers...)
	// 在这里也可以加一些代码，新特性
	return err
}

func (s *RatelimitSMSService) Name() string {
	return s.svc.Name()
}
