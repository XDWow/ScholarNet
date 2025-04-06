package repository

import (
	"context"
	"database/sql"
	"github.com/LXD-c/basic-go/webook/internal/domain"
	"github.com/LXD-c/basic-go/webook/internal/repository/cache"
	"github.com/LXD-c/basic-go/webook/internal/repository/dao"
	"time"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
)

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindByWechat(ctx context.Context, openID string) (domain.User, error)
}

type CacheUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDAO, cache cache.UserCache) UserRepository {
	return &CacheUserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (repo *CacheUserRepository) Create(ctx context.Context, u domain.User) error {
	return repo.dao.Insert(ctx, repo.domainToEntity(u))
}

func (repo *CacheUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.entityToDomain(u), nil
}

func (repo *CacheUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := repo.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return repo.entityToDomain(u), nil
}

func (repo *CacheUserRepository) FindByWechat(ctx context.Context, openID string) (domain.User, error) {
	u, err := repo.dao.FindByWechat(ctx, openID)
	if err != nil {
		return domain.User{}, err
	}
	return repo.entityToDomain(u), nil
}

// 查询频率较高，引入缓存机制
func (repo *CacheUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
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
	u = repo.entityToDomain(ue)
	err = repo.cache.Set(ctx, u)
	if err != nil {
		// 缓存设置失败不是大问题，可能是偶发性的超时问题
		// 不必返回个空的return domain.User{}, err
		// 打日志，做监控，防止Redis万一是崩了
	}
	return u, err
}

func (repo *CacheUserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,
		Ctime:    u.Ctime.UnixMilli(),
	}
}

func (repo *CacheUserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:             u.Id,
		Email:          u.Email.String,
		Password:       u.Password,
		Phone:          u.Phone.String,
		Wechat_openID:  u.Wechat_openID.String,
		Wechat_unionID: u.Wechat_unionID.String,
		Ctime:          time.UnixMilli(u.Ctime),
	}
}
