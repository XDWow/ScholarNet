package time

import (
	"context"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	tm := time.NewTicker(time.Second)
	// 这一句不要忘了
	// 避免潜在的 goroutine 泄露的问题
	defer tm.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			t.Log("超时了，或者被取消了")
			goto end
		case now := <-tm.C:
			t.Log(now.Unix())
		}
	}
end:
	t.Log("退出循环")
}
