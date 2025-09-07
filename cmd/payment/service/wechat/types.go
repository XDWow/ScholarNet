package wechat

import (
	"context"
	"github.com/XD/ScholarNet/cmd/payment/domain"
)

type PaymentService interface {
	// Prepay 预支付，对应于微信创建订单的步骤
	Prepay(ctx context.Context, pmt domain.Payment) (string, error)
}
