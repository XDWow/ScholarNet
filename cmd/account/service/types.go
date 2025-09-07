package service

import (
	"context"
	"github.com/XD/ScholarNet/cmd/account/domain"
)

type AccountService interface {
	Credit(ctx context.Context, cr domain.Credit) error
}
