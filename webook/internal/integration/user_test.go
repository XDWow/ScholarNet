package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"gitee.com/geekbang/basic-go/webook/internal/web"
	"gitee.com/geekbang/basic-go/webook/ioc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUserHandler_e2e_SendLoginSMSCode(t *testing.T) {
	server := InitWebServer()
	rdb := ioc.InitRedis()
	testCases := []struct {
		name string

		// 要考虑准备数据
		before func(*testing.T)
		// 第三方依赖 Redis ,要验证数据 数据库的数据对不对，你 Redis 的数据对不对
		after   func(*testing.T)
		reqBody string

		wantCode int
		wantBody web.Result
	}{
		{
			name: "发送成功",
			before: func(t *testing.T) {
				// 不需要在 Redis 中准备数据，什么也没有
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				// 验证并删除数据
				val, err := rdb.GetDel(ctx, "phone_code:login:17312345678").Result()
				assert.NoError(t, err)
				// 验证码是 6 位
				assert.True(t, len(val) == 6)
			},
			reqBody: `
{
	"phone":"17312345678"
}
`,
			wantCode: 200,
			wantBody: web.Result{
				Msg: "发送成功",
			},
		},
		{
			name: "Bind 失败",
			before: func(t *testing.T) {
				// 不需要在 Redis 中准备数据，什么也没有
			},
			after: func(t *testing.T) {},
			reqBody: `
{
	"phone":"17312345678"

`,
			wantCode: 400,
			//wantBody: web.Result{
			//	Msg: "bind 失败",
			//},
		},
		{
			name: "手机号输入有误",
			before: func(t *testing.T) {
				// 不需要在 Redis 中准备数据，什么也没有
			},
			after: func(t *testing.T) {},
			reqBody: `
{
	"phone":""
}
`,
			wantCode: 200,
			wantBody: web.Result{
				Code: 4,
				Msg:  "输入有误",
			},
		},
		{
			name: "发送太频繁",
			before: func(t *testing.T) {
				// 提前在 redis 中准备这个手机号码的一个验证码
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				_, err := rdb.Set(ctx, "phone_code:login:17312345678", "123456",
					time.Minute*9+time.Second*30).Result()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				// 验证并删除数据
				val, err := rdb.GetDel(ctx, "phone_code:login:17312345678").Result()
				assert.NoError(t, err)
				// 验证码是 6 位
				assert.True(t, len(val) == 6)
			},
			reqBody: `
{
	"phone":"17312345678"
}
`,
			wantCode: 200,
			wantBody: web.Result{
				Msg: "发送太频繁，请稍后再试",
			},
		},
		{
			name: "系统错误",
			before: func(t *testing.T) {
				// 这个手机号码，已经有一个验证码了，但是没有过期时间
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				_, err := rdb.Set(ctx, "phone_code:login:17312345678", "123456", 0).Result()
				cancel()
				assert.NoError(t, err)

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 你要清理数据
				// "phone_code:%s:%s"
				val, err := rdb.GetDel(ctx, "phone_code:login:17312345678").Result()
				cancel()
				assert.NoError(t, err)
				// 你的验证码是 6 位,没有被覆盖，还是123456
				assert.Equal(t, "123456", val)
			},
			reqBody: `
{
	"phone": "17312345678"
}
`,
			wantCode: 200,
			wantBody: web.Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			req, err := http.NewRequest(http.MethodPost,
				"/users/login_sms/code/send", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			// 数据是 JSON 格式
			req.Header.Set("Content-Type", "application/json")
			// 这里你就可以继续使用 req

			resp := httptest.NewRecorder()
			// 这就是 HTTP 请求进去 GIN 框架的入口。
			// 当你这样调用的时候，GIN 就会处理这个请求
			// 响应写回到 resp 里
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != 200 {
				return
			}
			var webRes web.Result
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBody, webRes)
			tc.after(t)
		})
	}
}
