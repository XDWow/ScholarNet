package service

import "context"

// Service 发送短信的抽象
// 这是一个为了适配不同的短信供应商的抽象
//
//go:generate mockgen -source=./types.go -package=smsmocks -destination=mocks/sms.mock.go Service
type Service interface {
	Send(ctx context.Context, tplId string, args []string, numbers ...string) error
}
