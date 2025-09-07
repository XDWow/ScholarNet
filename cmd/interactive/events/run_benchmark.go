package main

import (
	"context"
	"time"

	"github.com/XD/ScholarNet/cmd/interactive/events"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
)

func main() {
	// 创建日志器
	l := logger.NewNoOpLogger()

	// 创建简化压测器
	benchmark := events.NewSimpleBenchmark(l)

	// 创建上下文（5分钟超时）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 运行对比测试
	benchmark.RunComparisonTest(ctx)
}
