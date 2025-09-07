package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"github.com/XD/ScholarNet/cmd/pkg/ratelimit"
	"github.com/XD/ScholarNet/cmd/sms/service"
)

const key = "sms_tencent"

// 不到逼不得已，不要做成公开的
var errLimited = errors.New("短信服务触发限流")

type RatelimitSMSService struct {
	svc     service.Service
	limiter ratelimit.Limiter
}

func NewRatelimitSMSService(svc service.Service, limiter ratelimit.Limiter) *RatelimitSMSService {
	return &RatelimitSMSService{
		svc:     svc,
		limiter: limiter,
	}
}

func (r *RatelimitSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	limited, err := r.limiter.Limit(ctx, key)
	if err != nil {
		return fmt.Errorf("短信服务判断是否限流异常 %w", err)
	}
	if limited {
		return errLimited
	}
	// 最终业务逻辑交给了被装饰实现
	return r.svc.Send(ctx, tplId, args, numbers...)
}

//type RatelimitSMSServiceV1 struct {
//	Service
//	limiter ratelimit.Limiter
//}
