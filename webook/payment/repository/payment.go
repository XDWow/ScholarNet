package repository

import (
	"context"
	"github.com/LXD-c/basic-go/webook/payment/domain"
	"github.com/LXD-c/basic-go/webook/payment/repository/dao"
)

type PaymentRepository struct {
	dao dao.PaymentGORMDAO
}

func NewPaymentRepository(dao dao.PaymentGORMDAO) *PaymentRepository {
	return &PaymentRepository{
		dao: dao,
	}
}

func (repo *PaymentRepository) UpdatePayment(ctx context.Context, pmt domain.Payment) error {
	return repo.dao.UpdatePayment(ctx, repo.ToEntity(pmt))
}

func (repo *PaymentRepository) AddPayment(ctx context.Context, pmt domain.Payment) error {
	return repo.dao.InsertPayment(ctx, repo.ToEntity(pmt))
}

func (repo *PaymentRepository) GetPayment(ctx context.Context, bizTradeNo string) (domain.Payment, error) {
	res, err := repo.dao.GetPayment(ctx, bizTradeNo)
	return repo.ToDomain(res), err
}

func (repo *PaymentRepository) ToEntity(pmt domain.Payment) dao.Payment {
	return dao.Payment{
		Amt:         pmt.Amt.Total,
		Currency:    pmt.Amt.Currency,
		BizTradeNO:  pmt.BizTradeNo,
		Description: pmt.Description,
		Status:      domain.PaymentStatusInit,
	}
}

func (repo *PaymentRepository) ToDomain(pmt dao.Payment) domain.Payment {
	return domain.Payment{
		Amt: domain.Amount{
			Total:    pmt.Amt,
			Currency: pmt.Currency,
		},
		BizTradeNo:  pmt.BizTradeNO,
		Description: pmt.Description,
		Status:      domain.PaymentStatus(pmt.Status),
		TxnID:       pmt.TxnID.String,
	}
}
