package auth

import (
	"context"
	"errors"
	"github.com/LXD-c/basic-go/webook/internal/service/sms"
	"github.com/golang-jwt/jwt/v5"
)

type SMSService struct {
	svc sms.Service
	key string
}

// key 是 JWT 签发时使用的密钥，用于验证 JWT
func NewSMSService(svc sms.Service, key string) sms.Service {
	return &SMSService{
		svc: svc,
		key: key,
	}
}

// 或者想用一个方法产生 Token 也可以
//func (s *SMSService) GenerateToken(ctx context.Context, tplId string) (string, error) {
//
//}

// Send 发送，其中 biz 必须是线下申请的一个代表业务方的 token
func (s SMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	var tc Claims
	// 是不是就在这？
	// 如果我这里能解析成功，说明就是对应的业务方
	// 没有 error 就说明，token 是我发的
	token, err := jwt.ParseWithClaims(biz, &tc, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.key), nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("invalid token")
	}
	// 到这就是通过所有验证，可以准备发送了
	return s.svc.Send(ctx, tc.Tpl, args, numbers...)
}

// 定义一个能放入 jwtToken 的数据结构体,同时也能将 jwtToken 中的数据解析到这个结构体
type Claims struct {
	jwt.RegisteredClaims
	Tpl string
}
