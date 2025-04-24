package grpc

import (
	"context"
	"errors"
	"github.com/ecodeclub/ekit/queue"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"sync/atomic"
	"time"
)

type CounterLimiter struct {
	cnt       *atomic.Int32
	threshold int32
}

func (l *CounterLimiter) NewServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		cnt := l.cnt.Add(1)
		defer func() {
			l.cnt.Add(-1)
		}()
		if cnt > l.threshold {
			// 这里就是拒绝
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
		}
		return handler(ctx, req)
	}
}

type FixedWindowLimiter struct {
	// 窗口大小
	window time.Duration
	// 上一个窗口的起始时间
	lastStart time.Time

	// 当前窗口的请求数量
	cnt int
	// 窗口允许的最大的请求数量
	threshold int

	lock sync.Mutex
}

func (l *FixedWindowLimiter) NewServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		l.lock.Lock()
		now := time.Now()
		// 要换窗口了
		if now.After(l.lastStart.Add(l.window)) {
			l.lastStart = now
			l.cnt = 0
		}
		l.cnt++
		if l.cnt <= l.threshold {
			l.lock.Unlock()
			res, err := handler(ctx, req)
			return res, err
		}
		l.lock.Unlock()
		return nil, status.Errorf(codes.ResourceExhausted, "限流了")
	}
}

type SlideWindowLimiter struct {
	window time.Duration
	// 请求的时间戳
	queue     queue.PriorityQueue[time.Time]
	lock      sync.Mutex
	threshold int
}

func (l *SlideWindowLimiter) NewServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		l.lock.Lock()
		now := time.Now()
		if l.queue.Len() < l.threshold {
			_ = l.queue.Enqueue(time.Now())
			l.lock.Unlock()
			return handler(ctx, req)
		}
		windowStart := now.Add(-l.window)
		for {
			// 最早的请求
			first, _ := l.queue.Peek()
			if first.Before(windowStart) {
				// 就是删了 first
				_, _ = l.queue.Dequeue()
			} else {
				// 退出循环
				break
			}
		}
		if l.queue.Len() < l.threshold {
			_ = l.queue.Enqueue(time.Now())
			l.lock.Unlock()
			return handler(ctx, req)
		}
		l.lock.Unlock()
		return nil, status.Errorf(codes.ResourceExhausted, "限流了")
	}
}

type TokenBucketLimiter struct {
	//ch      *time.Ticker
	buckets chan struct{}
	// 每隔多久一个令牌
	interval time.Duration

	closeCh chan struct{}
}

// 把 capacity 设置成0，就是漏桶算法
// 但是，代码可以简化
func NewTokenBucketLimiter(interval time.Duration, capacity int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		interval: interval,
		buckets:  make(chan struct{}, capacity),
	}
}

func (l *TokenBucketLimiter) NewServerInterceptor() grpc.UnaryServerInterceptor {
	ticker := time.NewTicker(l.interval)
	go func() {
		for {
			select {
			// 通过 ch 退出, 也可以通过传入一个 ctx 退出
			case <-l.closeCh:
				return
			case <-ticker.C:
				select {
				case l.buckets <- struct{}{}:
				default:

				}
			}
		}
		//for _ = range ticker.C {
		//	select {
		//	case l.buckets <- struct{}{}:
		//	// 发到了桶里面
		//	default:
		//		// 桶满了
		//	}
		//}
	}()
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		select {
		case <-l.buckets:
			// 拿到了令牌
			return handler(ctx, req)
			//default:
		// 就意味着你认为，没有令牌不应阻塞，直接返回
		//return nil, status.Errorf(codes.ResourceExhausted, "限流了")
		case <-ctx.Done():
			// 没有令牌就等令牌，直到超时
			return nil, ctx.Err()
		}
	}
}

// 向外暴露 Close 方法：关闭 closeCh，监听这个通道的就能收到消息
// 不能反复调用 Close: 不能二次关闭通道，会 panic
func (l *TokenBucketLimiter) Close() error {
	close(l.closeCh)
	return nil
}

// LeakyBucket 漏桶算法
type LeakyBuketLimiter struct {
	// 每隔多久一个令牌
	interval  time.Duration
	closeCh   chan struct{}
	closeOnce sync.Once
}

func (l *LeakyBuketLimiter) NewLeakyBuketLimiter(interval time.Duration) *LeakyBuketLimiter {
	return &LeakyBuketLimiter{
		interval: interval,
		closeCh:  make(chan struct{}),
		// 零值有效
		closeOnce: sync.Once{},
	}
}

func (l *LeakyBuketLimiter) NewServerInterceptor() grpc.UnaryServerInterceptor {
	ticker := time.NewTicker(l.interval)
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		select {
		case <-l.closeCh:
			return nil, errors.New("限流器关闭")
		case <-ticker.C:
			// 拿到了令牌
			return handler(ctx, req)
		case <-ctx.Done():
			return nil, ctx.Err()
			// 这样意味着你认为：没拿到令牌不应该阻塞，直接返回
			//default:
			//	return nil, status.Errorf(codes.ResourceExhausted, "限流了")
		}
	}
}

func (l *LeakyBuketLimiter) Close() error {
	l.closeOnce.Do(func() {
		close(l.closeCh)
	})
	return nil
}
