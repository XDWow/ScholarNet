package repository

import (
	"context"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository/cache"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao"
)

var (
	ErrDuplicateEmail = dao.ErrDuplicateEmail
	ErrUserNotFound   = dao.ErrRecordNotFound
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UseCache
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

func (repo *UserRepository) Create(ctx context.Context, u domain.User) error {
	return repo.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (repo *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil
}

func (repo *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := repo.cache.Get(ctx, id)
	//三类：
	// 缓存里面有数据
	// 缓存里面没有数据
	// 缓存出错了，不知道有没有数据
	if err == nil {
		return u, nil
	}

	// 缓存出错，加不加载数据库
	// 选加载--做好兜底准备，万一 Redis 崩了，要保护住数据库。面试这样说，制造问题
	// 怎么保护：1、数据库限流 2、备用集群，设置一个垃圾一点的备用Redis 3、本地 4、
	// 选不加载--用户体验差一点
	ue, err := repo.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	u = repo.toDomain(ue)
	err = repo.cache.Set(ctx, u)
	if err != nil {
		// 缓存设置失败不是大问题，可能是偶发性的超时问题
		// 不必返回个空的return domain.User{}, err
		// 打日志，做监控，防止Redis万一是崩了
	}
	return u, err
}

func (repo *UserRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}
}
