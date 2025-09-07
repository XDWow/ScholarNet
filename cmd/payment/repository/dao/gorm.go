package dao

import (
	"context"
	"github.com/XD/ScholarNet/cmd/payment/domain"
	"gorm.io/gorm"
	"time"
)

type paymentGORMDAO struct {
	db *gorm.DB
}

func NewPaymentGORMDAO(db *gorm.DB) PaymentDAO {
	return &paymentGORMDAO{
		db: db,
	}
}

func (p *paymentGORMDAO) FindExpiredPayment(
	ctx context.Context,
	offset int, limit int, t time.Time) ([]Payment, error) {
	var res []Payment
	err := p.db.WithContext(ctx).Where("status = ? AND utime < ?",
		// 我的 IDE 有问题，AsUint8 会报错
		uint8(domain.PaymentStatusInit), t.UnixMilli()).
		Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

func (p *paymentGORMDAO) UpdateTxnIDAndStatus(ctx context.Context,
	bizTradeNo string,
	txnID string, status domain.PaymentStatus) error {
	return p.db.WithContext(ctx).Model(&Payment{}).
		Where("biz_trade_no = ?", bizTradeNo).
		Updates(map[string]interface{}{
			"status": status,
			"txnID":  txnID,
			"utime":  time.Now().UnixMilli(),
		}).Error
}

func (p *paymentGORMDAO) InsertPayment(ctx context.Context, pmt Payment) error {
	now := time.Now().UnixMilli()
	pmt.Utime = now
	pmt.Ctime = now
	return p.db.Create(&pmt).Error
}

func (p *paymentGORMDAO) GetPayment(ctx context.Context, bizTradeNo string) (Payment, error) {
	var res Payment
	err := p.db.Where("biz_trade_no = ?", bizTradeNo).First(&res).Error
	return res, err
}
