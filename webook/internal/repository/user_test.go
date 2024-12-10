package repository

import (
	"context"
	"database/sql"
	"errors"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository/cache"
	cachemocks "gitee.com/geekbang/basic-go/webook/internal/repository/cache/mocks"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao"
	daomocks "gitee.com/geekbang/basic-go/webook/internal/repository/dao/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestCacheUserRepository_FindById(t *testing.T) {
	// 111ms.11111ns
	now := time.Now()
	// 你要去掉毫秒以外的部分
	// 111ms
	now = time.UnixMilli(now.UnixMilli())
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache)

		id int64

		wantUser domain.User
		wantErr  error
	}{
		{
			name: "缓存命中",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				ud := daomocks.NewMockUserDAO(ctrl)
				uc := cachemocks.NewMockUserCache(ctrl)
				// 这里一定要 int64(1) ,写 1 默认是 int ，类型错误，测试不通过
				uc.EXPECT().Get(gomock.Any(), int64(1)).
					Return(domain.User{
						Id:       1,
						Email:    "123@qq.com",
						Password: "$2a$10$bz5l5B64CxdnNpBL5ymYt.zs2MTtXRjAG1CHQxpAqFaExmRqS9qwS",
						Phone:    "17312345678",
						Ctime:    now,
					}, nil)
				return ud, uc
			},
			id: 1,
			wantUser: domain.User{
				Id:       1,
				Email:    "123@qq.com",
				Password: "$2a$10$bz5l5B64CxdnNpBL5ymYt.zs2MTtXRjAG1CHQxpAqFaExmRqS9qwS",
				Phone:    "17312345678",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "缓存未命中，数据库查询失败",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				ud := daomocks.NewMockUserDAO(ctrl)
				uc := cachemocks.NewMockUserCache(ctrl)
				// 这里一定要 int64(1) ,写 1 默认是 int ，类型错误，测试不通过
				uc.EXPECT().Get(gomock.Any(), int64(1)).
					Return(domain.User{}, cache.ErrUserNotFound)
				ud.EXPECT().FindById(gomock.Any(), int64(1)).
					Return(dao.User{}, errors.New("随便什么错误"))
				return ud, uc
			},
			id:       1,
			wantUser: domain.User{},
			wantErr:  errors.New("随便什么错误"),
		},
		{
			name: "缓存未命中，数据库查询成功",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				ud := daomocks.NewMockUserDAO(ctrl)
				uc := cachemocks.NewMockUserCache(ctrl)
				// 这里一定要 int64(1) ,写 1 默认是 int ，类型错误，测试不通过
				uc.EXPECT().Get(gomock.Any(), int64(1)).
					Return(domain.User{}, cache.ErrUserNotFound)
				uc.EXPECT().Set(gomock.Any(), domain.User{
					Id:       1,
					Email:    "123@qq.com",
					Password: "密码",
					Phone:    "17312345678",
					Ctime:    now,
				},
				).Return(nil)
				ud.EXPECT().FindById(gomock.Any(), int64(1)).
					Return(dao.User{
						Id: 1,
						Email: sql.NullString{
							"123@qq.com",
							true,
						},
						Password: "密码",
						Phone: sql.NullString{
							"17312345678",
							true,
						},
						Ctime: now.UnixMilli(),
						Utime: now.UnixMilli(),
					}, nil)
				return ud, uc
			},
			id: 1,
			wantUser: domain.User{
				Id:       1,
				Email:    "123@qq.com",
				Password: "密码",
				Phone:    "17312345678",
				Ctime:    now,
			},
			wantErr: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := NewUserRepository(testCase.mock(ctrl))
			u, err := repo.FindById(context.Background(), testCase.id)
			assert.Equal(t, testCase.wantUser, u)
			assert.Equal(t, testCase.wantErr, err)
		})
	}
}
