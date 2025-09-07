package wechat

import (
	"context"
	"errors"
	"github.com/XD/ScholarNet/cmd/payment/domain"
	"github.com/XD/ScholarNet/cmd/payment/repository"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	events2 "github.com/XD/ScholarNet/cmd/ranking/events/events"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"time"
)

var errUnknownTransactionState = errors.New("未知的微信事务状态")

type NativePaymentService struct {
	svc  *native.NativeApiService
	repo repository.PaymentRepository
	l    logger.LoggerV1

	appID     string
	mchID     string
	notifyURL string
	producer  events2.Producer
	// 在微信 native 里面，分别是
	// SUCCESS：支付成功
	// REFUND：转入退款
	// NOTPAY：未支付
	// CLOSED：已关闭
	// REVOKED：已撤销（付款码支付）
	// USERPAYING：用户支付中（付款码支付）
	// PAYERROR：支付失败(其他原因，如银行返回失败)
	nativeCBTypeToStatus map[string]domain.PaymentStatus
}

func NewNativePaymentService(svc *native.NativeApiService,
	repo repository.PaymentRepository,
	l logger.LoggerV1,
	appid, mchid string) *NativePaymentService {
	return &NativePaymentService{
		svc:       svc,
		repo:      repo,
		l:         l,
		appID:     appid,
		mchID:     mchid,
		notifyURL: "http://wechat.meoying.com/pay/callback",
		nativeCBTypeToStatus: map[string]domain.PaymentStatus{
			"SUCCESS":  domain.PaymentStatusSuccess,
			"PAYERROR": domain.PaymentStatusFailed,
			// 这个状态，有些人会考虑映射过去 PaymentStatusFailed
			"NOTPAY":     domain.PaymentStatusInit,
			"USERPAYING": domain.PaymentStatusInit,
			"CLOSED":     domain.PaymentStatusFailed,
			"REVOKED":    domain.PaymentStatusFailed,
			"REFUND":     domain.PaymentStatusRefund,
			// 其它状态你都可以加
		},
	}
}

// Prepay 为了拿到扫码支付的二维码
// 同时在数据库存入支付记录，也就是 domain.Payment
func (s *NativePaymentService) PrePay(ctx context.Context, pmt domain.Payment) (string, error) {
	// 唯一索引冲突
	// 业务方唤起了支付，但是没付，下一次再过来，应该换 BizTradeNO

	// 我要做的：
	// 1、存支付记录
	// 2、调用微信 api 获取 code_url
	err := s.repo.AddPayment(ctx, pmt)
	if err != nil {
		return "", err
	}
	//sn := uuid.New().String()
	resp, result, err := s.svc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(s.appID),
		Mchid:       core.String(s.mchID),
		Description: core.String(pmt.Description),
		// 这个地方是有讲究的
		// 选择1：业务方直接给我，我透传，我啥也不干
		// 选择2：业务方给我它的业务标识，我自己生成一个 - 担忧出现重复
		// 注意，不管你是选择 1 还是选择 2，业务方都一定要传给你（webook payment）一个唯一标识
		// Biz + BizTradeNo 唯一， biz + biz_id
		OutTradeNo: core.String(pmt.BizTradeNo),
		NotifyUrl:  core.String(s.notifyURL),
		// 设置30分钟有效
		TimeExpire: core.Time(time.Now().Add(time.Minute * 30)),
		Amount: &native.Amount{
			Total:    core.Int64(pmt.Amt.Total),
			Currency: core.String(pmt.Amt.Currency),
		},
	})
	// 打印详细信息
	s.l.Debug("微信 prepay 响应",
		logger.Field{Key: "result", Value: result},
		logger.Field{Key: "resp", Value: resp})
	if err != nil {
		return "", err
	}
	return *resp.CodeUrl, nil
}

func (s *NativePaymentService) SyncWechatInfo(ctx context.Context, bizTradeNO string) error {
	txn, _, err := s.svc.QueryOrderByOutTradeNo(ctx, native.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(bizTradeNO),
		Mchid:      core.String(s.mchID),
	})
	if err != nil {
		return err
	}
	return s.updateByTxn(ctx, txn)
}

func (s *NativePaymentService) FindExpiredPayment(ctx context.Context, offset, limit int, t time.Time) ([]domain.Payment, error) {
	return s.repo.FindExpiredPayment(ctx, offset, limit, t)
}

func (s *NativePaymentService) HandleCallback(ctx context.Context, txn *payments.Transaction) error {
	return s.updateByTxn(ctx, txn)
}

func (s *NativePaymentService) GetPayment(ctx context.Context, bizTradeNo string) (domain.Payment, error) {
	// 在这里，我能不能设计一个慢路径？如果要是不知道支付结果，我就去微信里面查一下？
	// 或者异步查一下？
	// 可以，快路径查数据库，慢路径发请求给微信查支付结果，降级限流时只走快路径
	return s.repo.GetPayment(ctx, bizTradeNo)
}

// 回调，基本就是更新 Pyment 状态
func (s *NativePaymentService) HandlerCallback(ctx context.Context, txn *payments.Transaction) error {
	return s.updateByTxn(ctx, txn)
}

func (s *NativePaymentService) updateByTxn(ctx context.Context, txn *payments.Transaction) error {
	// 通过一个map将微信的状态映射为我的状态, switch 太呆了
	status, ok := s.nativeCBTypeToStatus[*txn.TradeState]
	if !ok {
		return errors.New("未知的微信状态，映射失败")
	}
	// 核心就是更新数据库状态
	err := s.repo.UpdatePayment(ctx, domain.Payment{
		// 用这个来找记录
		BizTradeNo: *txn.OutTradeNo,
		// 更新状态，TxnID默认也是没有的，微信回调才有，所以也更新，其他没字段要更新了:别忘记 Utime(猪队友）,db操作时加入
		Status: status,
		TxnID:  *txn.TransactionId,
	})
	if err != nil {
		// 这里有一个小问题，就是如果超时了的话，你都不知道更新成功了没
		return err
	}
	// 处于结束状态
	err1 := s.producer.ProducePaymentEvent(ctx, events2.PaymentEvent{
		BizTradeNO: *txn.TradeState,
		Status:     status.AsUint8(),
	})
	if err1 != nil {
		// 要做好监控和告警
		s.l.Error("发送支付事件失败", logger.Error(err),
			logger.String("biz_trade_no", *txn.TradeState))
	}
	// 虽然发送事件失败，但是数据库记录了，所以可以返回 Nil
	return nil
}
