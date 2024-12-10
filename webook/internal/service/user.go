package service

import (
	"context"
	"errors"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail        = repository.ErrUserDuplicate
	ErrInvalidUserOrPassword = errors.New("用户不存在或者密码不对")
)

type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	Login(ctx context.Context, email string, password string) (domain.User, error)
	Profile(ctx context.Context, id int64) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (svc *userService) SignUp(ctx context.Context, u domain.User) error {
	//在服务层加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	//先找，没找到就创建
	//这个叫做快路径
	u, err := svc.repo.FindByPhone(ctx, phone)
	//根据错误类型，判断是系统出错还是没有这个用户
	if err != repository.ErrUserNotFound {
		return u, err
	}

	//到这里是：没有这个用户
	// 这个叫做慢路径，操作数据库，太慢了，且耗费资源多
	// 在系统资源不足，触发降级之后，不执行慢路径了
	//if ctx.Value("降级") == "true" {
	//	return domain.User{}, errors.New("系统降级了")
	//}
	err = svc.repo.Create(ctx, domain.User{Phone: phone})
	if err != nil && err != repository.ErrUserDuplicate {
		return u, err
	}
	// 因为这里会遇到主从延迟的问题
	return svc.repo.FindByPhone(ctx, phone)
}

func (svc *userService) FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error) {
	//先找，没找到就创建
	//这个叫做快路径
	u, err := svc.repo.FindByWechat(ctx, info.OpenID)
	//根据错误类型，判断是系统出错还是没有这个用户
	if err != repository.ErrUserNotFound {
		return u, err
	}

	//到这里是：没有这个用户
	// 这个叫做慢路径，操作数据库，太慢了，且耗费资源多
	// 在系统资源不足，触发降级之后，不执行慢路径了
	//if ctx.Value("降级") == "true" {
	//	return domain.User{}, errors.New("系统降级了")
	//}
	err = svc.repo.Create(ctx, domain.User{Wechat_unionID: info.UnionID, Wechat_openID: info.OpenID})
	if err != nil && err != repository.ErrUserDuplicate {
		return u, err
	}
	// 因为这里会遇到主从延迟的问题
	return svc.repo.FindByWechat(ctx, info.OpenID)
}

func (svc *userService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	// db错误
	if err != nil {
		return domain.User{}, err
	}
	//检查密码是否正确
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (svc *userService) Profile(ctx context.Context, id int64) (domain.User, error) {
	u, err := svc.repo.FindById(ctx, id)
	return u, err
}
