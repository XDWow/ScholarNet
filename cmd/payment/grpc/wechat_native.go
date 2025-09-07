package grpc

import (
	"context"
	pmtv1 "github.com/XD/ScholarNet/cmd/api/proto/gen/payment/v1"
	"github.com/XD/ScholarNet/cmd/payment/domain"
	"github.com/XD/ScholarNet/cmd/payment/service/wechat"
	"google.golang.org/grpc"
)

// 定义一个 rpc 服务: 实现 WechatPaymentServiceServer 接口
type WechatServiceServer struct {
	pmtv1.UnimplementedWechatPaymentServiceServer
	svc *wechat.NativePaymentService
}

func (s *WechatServiceServer) NativePrepay(ctx context.Context, request *pmtv1.PrePayRequest) (*pmtv1.NativePrePayResponse, error) {
	codeURL, err := s.svc.PrePay(ctx, domain.Payment{
		// 微信支付 api 需要的参数
		Amt: domain.Amount{
			Currency: request.Amt.Currency,
			Total:    request.Amt.Total,
		},
		BizTradeNo:  request.BizTradeNo,
		Description: request.Description,
	})
	if err != nil {
		return nil, err
	}
	return &pmtv1.NativePrePayResponse{
		CodeUrl: codeURL,
	}, nil
}

func (s *WechatServiceServer) GetPayment(ctx context.Context, request *pmtv1.GetPaymentRequest) (*pmtv1.GetPaymentResponse, error) {
	p, err := s.svc.GetPayment(ctx, request.BizTradeNo)
	if err != nil {
		return nil, err
	}
	return &pmtv1.GetPaymentResponse{
		Status: pmtv1.PaymentStatus(p.Status),
	}, nil
}

func NewWechatServiceServer(svc *wechat.NativePaymentService) *WechatServiceServer {
	return &WechatServiceServer{svc: svc}
}

func (s *WechatServiceServer) Register(server *grpc.Server) {
	pmtv1.RegisterWechatPaymentServiceServer(server, s)
}
