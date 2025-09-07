package repository

import (
	"context"
	"github.com/XD/ScholarNet/cmd/search/repository/dao"
)

type anyRepository struct {
	dao dao.AnyDAO
}

func NewAnyRepository(dao dao.AnyDAO) AnyRepository {
	return &anyRepository{dao: dao}
}

func (repo *anyRepository) Input(ctx context.Context, index string, docID string, data string) error {
	return repo.dao.Insert(ctx, index, docID, data)
}
