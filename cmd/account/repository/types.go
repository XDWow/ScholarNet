package repository

import (
	"context"
	"github.com/XD/ScholarNet/cmd/account/domain"
)

type AccountRepository interface {
	AddCredit(ctx context.Context, c domain.Credit) error
}
