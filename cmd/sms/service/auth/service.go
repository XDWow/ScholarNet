package auth

import (
	"context"
	"github.com/XD/ScholarNet/cmd/sms/service"
	"github.com/golang-jwt/jwt/v5"
)

type SMSService struct {
	svc service.Service
	key []byte
}

func (s *SMSService) Send(ctx context.Context,
	// 改变了语义
	tplToken string, args []string, numbers ...string) error {
	var c Claims
	_, err := jwt.ParseWithClaims(tplToken, &c, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	return s.svc.Send(ctx, c.Tpl, args, numbers...)
}

type Claims struct {
	jwt.RegisteredClaims
	Tpl string
}
