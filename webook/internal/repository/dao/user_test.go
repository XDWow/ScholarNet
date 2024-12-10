package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestGORMUserDAO_Insert(t *testing.T) {
	testCases := []struct {
		name string

		// 这里为什么不用 ctrl ?
		// 因为 sqlmock 不是 gomock
		mock func(t *testing.T) *sql.DB
		user User

		wantErr error
	}{
		{
			name: "插入成功",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				res := sqlmock.NewResult(3, 1)
				// 这边预期的是正则表达式
				// 这个写法的意思就是，只要是 INSERT 到 users 的语句
				mock.ExpectExec("INSERT INTO `users` .*").WillReturnResult(res)
				return mockDB
			},
			user: User{
				Email: sql.NullString{
					String: "123@qq.com",
					Valid:  true,
				},
			},
			wantErr: nil,
		},
		{
			name: "重复插入",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				// 这边预期的是正则表达式
				// 这个写法的意思就是，只要是 INSERT 到 users 的语句
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnError(&mysql.MySQLError{
						Number: 1062,
					})
				return mockDB
			},
			user: User{
				Email: sql.NullString{
					String: "123@qq.com",
					Valid:  true,
				},
			},
			wantErr: ErrUserDuplicate,
		},
		{
			name: "数据库错误",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				// 这边预期的是正则表达式
				// 这个写法的意思就是，只要是 INSERT 到 users 的语句
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnError(errors.New("数据库错误"))
				return mockDB
			},
			user: User{
				Email: sql.NullString{
					String: "123@qq.com",
					Valid:  true,
				},
			},
			wantErr: errors.New("数据库错误"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			db, err := gorm.Open(gormMysql.New(gormMysql.Config{
				Conn: testCase.mock(t),
				// SELECT VERSION;
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				// 你 mock DB 不需要 ping
				DisableAutomaticPing: true,
				// 这个是什么呢？
				SkipDefaultTransaction: true,
			})
			assert.Nil(t, err)
			d := NewUserDAO(db)
			err = d.Insert(context.Background(), testCase.user)
			assert.Equal(t, testCase.wantErr, err)
		})
	}
}
