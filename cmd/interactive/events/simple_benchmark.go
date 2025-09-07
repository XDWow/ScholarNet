package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"github.com/XD/ScholarNet/cmd/pkg/saramax"
)

// SimpleBenchmarkResult 简化压测结果
type SimpleBenchmarkResult struct {
	TestType         string        // 测试类型：single 或 batch
	BatchSize        int           // 批量大小
	TotalMessages    int           // 总消息数
	TotalTime        time.Duration // 总耗时
	Throughput       float64       // 吞吐量（消息/秒）
	TransactionCount int           // 事务提交次数
	MessagesPerTx    float64       // 每条事务处理的消息数
}

// SimpleBenchmark 简化压测器
type SimpleBenchmark struct {
	l       logger.LoggerV1
	results []SimpleBenchmarkResult
	mu      sync.Mutex
}

// NewSimpleBenchmark 创建简化压测器
func NewSimpleBenchmark(l logger.LoggerV1) *SimpleBenchmark {
	return &SimpleBenchmark{
		l:       l,
		results: make([]SimpleBenchmarkResult, 0),
	}
}

// RunSingleInterfaceTest 运行单个接口测试
func (sb *SimpleBenchmark) RunSingleInterfaceTest(ctx context.Context, messageCount int) SimpleBenchmarkResult {
	sb.l.Info("开始单个接口测试", logger.Int("messageCount", messageCount))

	startTime := time.Now()

	// 使用您现有的单个接口：NewAsyncHandler
	handler := saramax.NewAsyncHandler[ReadEvent](sb.l, sb.singleMessageHandler)

	// 模拟消费消息
	for i := 0; i < messageCount; i++ {
		msg := &sarama.ConsumerMessage{}
		event := ReadEvent{Aid: int64(i + 1)}

		// 调用单个接口处理
		if err := sb.singleMessageHandler(ctx, msg, event); err != nil {
			sb.l.Error("单个接口处理失败", logger.Error(err))
		}

		// 检查上下文是否取消
		select {
		case <-ctx.Done():
			return SimpleBenchmarkResult{}
		default:
		}
	}

	totalTime := time.Since(startTime)

	result := SimpleBenchmarkResult{
		TestType:         "single",
		BatchSize:        1,
		TotalMessages:    messageCount,
		TotalTime:        totalTime,
		Throughput:       float64(messageCount) / totalTime.Seconds(),
		TransactionCount: messageCount, // 每条消息一个事务
		MessagesPerTx:    1.0,
	}

	sb.addResult(result)
	sb.l.Info("单个接口测试完成",
		logger.Duration("totalTime", totalTime),
		logger.Float64("throughput", result.Throughput))

	return result
}

// RunBatchInterfaceTest 运行批量接口测试
func (sb *SimpleBenchmark) RunBatchInterfaceTest(ctx context.Context, messageCount, batchSize int) SimpleBenchmarkResult {
	sb.l.Info("开始批量接口测试",
		logger.Int("messageCount", messageCount),
		logger.Int("batchSize", batchSize))

	startTime := time.Now()

	// 使用您现有的批量接口：NewAsyncBatchHandlerSimple
	handler := saramax.NewAsyncBatchHandlerSimple[ReadEvent](sb.l, sb.batchMessageHandler, batchSize)

	// 模拟批量消息处理
	batchCount := (messageCount + batchSize - 1) / batchSize // 向上取整
	for i := 0; i < batchCount; i++ {
		start := i * batchSize
		end := start + batchSize
		if end > messageCount {
			end = messageCount
		}

		// 构建批量消息
		messages := make([]*sarama.ConsumerMessage, 0, end-start)
		events := make([]ReadEvent, 0, end-start)

		for j := start; j < end; j++ {
			messages = append(messages, &sarama.ConsumerMessage{})
			events = append(events, ReadEvent{Aid: int64(j + 1)})
		}

		// 调用批量接口处理
		if err := sb.batchMessageHandler(ctx, messages, events); err != nil {
			sb.l.Error("批量接口处理失败", logger.Error(err))
		}

		// 检查上下文是否取消
		select {
		case <-ctx.Done():
			return SimpleBenchmarkResult{}
		default:
		}
	}

	totalTime := time.Since(startTime)

	result := SimpleBenchmarkResult{
		TestType:         "batch",
		BatchSize:        batchSize,
		TotalMessages:    messageCount,
		TotalTime:        totalTime,
		Throughput:       float64(messageCount) / totalTime.Seconds(),
		TransactionCount: batchCount, // 每个批次一个事务
		MessagesPerTx:    float64(messageCount) / float64(batchCount),
	}

	sb.addResult(result)
	sb.l.Info("批量接口测试完成",
		logger.Duration("totalTime", totalTime),
		logger.Float64("throughput", result.Throughput),
		logger.Int("transactionCount", result.TransactionCount))

	return result
}

// RunComparisonTest 运行对比测试
func (sb *SimpleBenchmark) RunComparisonTest(ctx context.Context) {
	sb.l.Info("开始批量接口 vs 单个接口对比测试")

	messageCount := 100000 // 10万条消息

	// 测试单个接口
	sb.RunSingleInterfaceTest(ctx, messageCount)

	// 测试不同批量大小
	batchSizes := []int{10, 50, 100, 200, 500}
	for _, batchSize := range batchSizes {
		sb.RunBatchInterfaceTest(ctx, messageCount, batchSize)
	}

	// 输出测试结果对比
	sb.printResults()
}

// singleMessageHandler 单个消息处理器（模拟您的 Consume 方法）
func (sb *SimpleBenchmark) singleMessageHandler(ctx context.Context, msg *sarama.ConsumerMessage, evt ReadEvent) error {
	// 模拟您的单个接口处理逻辑
	// 这里应该调用您实际的业务逻辑
	time.Sleep(50 * time.Microsecond)  // 模拟业务处理时间
	time.Sleep(200 * time.Microsecond) // 模拟事务开销
	return nil
}

// batchMessageHandler 批量消息处理器（模拟您的 Consume 方法）
func (sb *SimpleBenchmark) batchMessageHandler(ctx context.Context, msgs []*sarama.ConsumerMessage, events []ReadEvent) error {
	// 模拟您的批量接口处理逻辑
	// 这里应该调用您实际的业务逻辑
	baseTime := 100 * time.Microsecond
	perMessageTime := 50 * time.Microsecond
	totalProcessTime := baseTime + time.Duration(len(events))*perMessageTime
	time.Sleep(totalProcessTime)

	// 模拟批量事务开销
	time.Sleep(300 * time.Microsecond)
	return nil
}

// addResult 添加测试结果
func (sb *SimpleBenchmark) addResult(result SimpleBenchmarkResult) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.results = append(sb.results, result)
}

// printResults 打印测试结果
func (sb *SimpleBenchmark) printResults() {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	fmt.Println("\n=== 批量接口 vs 单个接口性能对比 ===")
	fmt.Printf("%-12s %-8s %-12s %-12s %-12s %-15s %-15s\n",
		"测试类型", "批量大小", "总消息数", "总耗时", "吞吐量(条/秒)", "事务数", "每事务消息数")
	fmt.Println(string(make([]byte, 100, 100)))

	for _, result := range sb.results {
		fmt.Printf("%-12s %-8d %-12d %-12s %-12.0f %-15d %-15.1f\n",
			result.TestType,
			result.BatchSize,
			result.TotalMessages,
			result.TotalTime.String(),
			result.Throughput,
			result.TransactionCount,
			result.MessagesPerTx)
	}

	// 计算性能提升
	if len(sb.results) >= 2 {
		single := sb.results[0]
		bestBatch := sb.results[len(sb.results)-1]

		throughputImprovement := (bestBatch.Throughput - single.Throughput) / single.Throughput * 100
		txReduction := (float64(single.TransactionCount) - float64(bestBatch.TransactionCount)) / float64(single.TransactionCount) * 100

		fmt.Printf("\n=== 性能提升总结 ===\n")
		fmt.Printf("吞吐量提升: %.1f%%\n", throughputImprovement)
		fmt.Printf("事务数减少: %.1f%%\n", txReduction)
		fmt.Printf("每事务处理消息数: %.1f 倍\n", bestBatch.MessagesPerTx)
	}
}
