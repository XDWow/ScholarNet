package grpc

import (
	"context"
	"fmt"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"github.com/XD/ScholarNet/cmd/pkg/ratelimit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InterceptorBuilder struct {
	limiter ratelimit.Limiter
	key     string
	l       logger.LoggerV1
}

// NewInterceptorBuilder key: user-service
// 整个应用、集群限流
func NewInterceptorBuilder(limiter ratelimit.Limiter, key string, l logger.LoggerV1) *InterceptorBuilder {
	return &InterceptorBuilder{limiter: limiter, key: key, l: l}
}

func (b *InterceptorBuilder) BuildServerInterceptorServiceBiz() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if idReq, ok := req.(*GetByIdRequest); ok {
			// 限流
			limited, err := b.limiter.Limit(ctx,
				// limiter:user:456
				fmt.Sprintf("limiter:user:%s:%d", info.FullMethod, idReq.Id))
			if err != nil {
				// err 不为nil，你要考虑你用保守的，还是用激进的策略
				// 这是保守的策略，可能导致：只因 Redis 限流器出现了问题， 相关业务都不可用，影响过大
				b.l.Error("判定限流出现问题", logger.Error(err))
				return nil, err
				// 这是激进的策略，可能导致：错误原因是用于限流 redis 崩了，还去处理只会加剧其压力
				// return handler(ctx, req)
			}
			if limited {
				return nil, status.Error(codes.ResourceExhausted, "触发限流")
			}
		}
		// handler 可能是下一个 Interceptor, 也可能是业务
		return handler(ctx, req)
	}
}
