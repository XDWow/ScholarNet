package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"github.com/XD/ScholarNet/cmd/pkg/ratelimit"
	limitmocks "github.com/XD/ScholarNet/cmd/pkg/ratelimit/mocks"
	"github.com/XD/ScholarNet/cmd/sms/service"
	"github.com/XD/ScholarNet/cmd/sms/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestRatelimitSMSService_Send(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (service.Service, ratelimit.Limiter)

		// 因为这边我们测试的是限流，输入是什么不关键，所以不需要定义

		// 输出
		wantErr error
	}{
		{
			name: "正常发送",
			mock: func(ctrl *gomock.Controller) (service.Service, ratelimit.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, nil)
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				return svc, limiter
			},
		},
		{
			name: "触发限流",
			mock: func(ctrl *gomock.Controller) (service.Service, ratelimit.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(true, nil)
				return svc, limiter
			},
			wantErr: errors.New("短信服务触发限流"),
		},
		{
			name: "限流器异常",
			mock: func(ctrl *gomock.Controller) (service.Service, ratelimit.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, errors.New("限流器异常"))
				return svc, limiter
			},
			wantErr: fmt.Errorf("短信服务判断是否限流异常 %w", errors.New("限流器异常")),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc, limiter := tc.mock(ctrl)
			limitSvc := NewRatelimitSMSService(svc, limiter)
			err := limitSvc.Send(context.Background(),
				"mytpl", []string{"123"}, "152xxxx")
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
