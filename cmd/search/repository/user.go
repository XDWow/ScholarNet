package repository

import (
	"context"
	"github.com/XD/ScholarNet/cmd/search/domain"
	"github.com/XD/ScholarNet/cmd/search/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type userRepository struct {
	dao dao.UserDAO
}

func NewUserRepository(dao dao.UserDAO) UserRepository {
	return &userRepository{dao}
}

func (repo *userRepository) InputUser(ctx context.Context, user domain.User) error {
	return repo.dao.InputUser(ctx, repo.ToEntity(user))
}

func (repo *userRepository) SearchUser(ctx context.Context, keywords []string) ([]domain.User, error) {
	users, err := repo.dao.Search(ctx, keywords)
	if err != nil {
		return nil, err
	}
	return slice.Map(users, func(idx int, src dao.User) domain.User {
		return domain.User{
			Id:       src.Id,
			Nickname: src.Nickname,
			Email:    src.Email,
			Phone:    src.Phone,
		}
	}), err
}

func (repo *userRepository) ToEntity(u domain.User) dao.User {
	return dao.User{
		Id:       u.Id,
		Nickname: u.Nickname,
		Email:    u.Email,
		Phone:    u.Phone,
	}
}
