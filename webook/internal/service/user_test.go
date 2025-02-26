package service

import (
	"context"
	"errors"
	"github.com/LXD-c/basic-go/webook/internal/domain"
	"github.com/LXD-c/basic-go/webook/internal/repository"
	repomocks "github.com/LXD-c/basic-go/webook/internal/repository/mocks"
	"github.com/LXD-c/basic-go/webook/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
)

func Test_userService_Login(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.UserRepository

		// 方法需要的输入，直接复制过来
		//ctx      context.Context 大家都一样，统一写一个就行
		email    string
		password string

		// 方法的输出,一样复制
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{
						Email:    "123@qq.com",
						Password: "$2a$10$bz5l5B64CxdnNpBL5ymYt.zs2MTtXRjAG1CHQxpAqFaExmRqS9qwS",
						Phone:    "17312345678",
						Ctime:    now,
					}, nil)
				return repo
			},
			email:    "123@qq.com",
			password: "hello#world123",

			wantUser: domain.User{
				Email:    "123@qq.com",
				Password: "$2a$10$bz5l5B64CxdnNpBL5ymYt.zs2MTtXRjAG1CHQxpAqFaExmRqS9qwS",
				Phone:    "17312345678",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "用户不存在",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
			email:    "123@qq.com",
			password: "hello#world123",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "DB错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, errors.New("mock db错误"))
				return repo
			},
			email:    "123@qq.com",
			password: "hello#world123",

			wantUser: domain.User{},
			wantErr:  errors.New("mock db错误"),
		},
		{
			name: "密码错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{
						Email:    "123@qq.com",
						Password: "0",
						Phone:    "17312345678",
						Ctime:    now,
					}, nil)
				return repo
			},
			email:    "123@qq.com",
			password: "hello#world123",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			//具体的测试代码
			svc := NewUserService(testCase.mock(ctrl), &logger.NopLogger{})
			u, err := svc.Login(context.Background(), testCase.email, testCase.password)
			assert.Equal(t, testCase.wantUser, u)
			assert.Equal(t, testCase.wantErr, err)
		})
	}
}

func TestEncrypted(t *testing.T) {
	res, err := bcrypt.GenerateFromPassword([]byte("hello#world123"), bcrypt.DefaultCost)
	if err == nil {
		t.Log(string(res))
	}
}
