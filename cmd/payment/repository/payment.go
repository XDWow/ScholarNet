package repository

import (
	"context"
	"github.com/XD/ScholarNet/cmd/payment/domain"
	"github.com/XD/ScholarNet/cmd/payment/repository/dao"
	"time"
)

type paymentRepository struct {
	dao dao.PaymentDAO
}

func NewPaymentRepository(dao dao.PaymentDAO) PaymentRepository {
	return &paymentRepository{
		dao: dao,
	}
}

func (p *paymentRepository) FindExpiredPayment(ctx context.Context, offset int, limit int, t time.Time) ([]domain.Payment, error) {
	pmts, err := p.dao.FindExpiredPayment(ctx, offset, limit, t)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Payment, 0, len(pmts))
	for _, pmt := range pmts {
		res = append(res, p.toDomain(pmt))
	}
	return res, nil
}

func (repo *paymentRepository) UpdatePayment(ctx context.Context, pmt domain.Payment) error {
	return repo.dao.UpdateTxnIDAndStatus(ctx, pmt.BizTradeNo, pmt.TxnID, pmt.Status)
}

func (repo *paymentRepository) AddPayment(ctx context.Context, pmt domain.Payment) error {
	return repo.dao.InsertPayment(ctx, repo.toEntity(pmt))
}

func (repo *paymentRepository) GetPayment(ctx context.Context, bizTradeNo string) (domain.Payment, error) {
	res, err := repo.dao.GetPayment(ctx, bizTradeNo)
	return repo.toDomain(res), err
}

func (repo *paymentRepository) toEntity(pmt domain.Payment) dao.Payment {
	return dao.Payment{
		Amt:         pmt.Amt.Total,
		Currency:    pmt.Amt.Currency,
		BizTradeNO:  pmt.BizTradeNo,
		Description: pmt.Description,
		Status:      domain.PaymentStatusInit,
	}
}

func (repo *paymentRepository) toDomain(pmt dao.Payment) domain.Payment {
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
