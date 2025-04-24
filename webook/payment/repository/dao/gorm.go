package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type PaymentGORMDAO struct {
	db *gorm.DB
}

func NewPaymentGORMDAO(db *gorm.DB) *PaymentGORMDAO {
	return &PaymentGORMDAO{
		db: db,
	}
}

func (p *PaymentGORMDAO) UpdatePayment(ctx context.Context, pmt Payment) error {
	return p.db.WithContext(ctx).Where("biz_trade_no = ?", pmt.BizTradeNO).
		Updates(map[string]any{
			"txn_id": pmt.TxnID,
			"status": pmt.Status,
			"utime":  time.Now().UnixMilli(),
		}).Error
}

func (p *PaymentGORMDAO) InsertPayment(ctx context.Context, pmt Payment) error {
	now := time.Now().UnixMilli()
	pmt.Utime = now
	pmt.Ctime = now
	return p.db.Create(&pmt).Error
}

func (p *PaymentGORMDAO) GetPayment(ctx context.Context, bizTradeNo string) (Payment, error) {
	var res Payment
	err := p.db.Where("biz_trade_no = ?", bizTradeNo).First(&res).Error
	return res, err
}
