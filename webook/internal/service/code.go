package service

import (
	"context"
	"fmt"
	"gitee.com/geekbang/basic-go/webook/internal/repository"
	"gitee.com/geekbang/basic-go/webook/internal/service/sms"
	"math/rand"
)

var (
	ErrCodeSendTooMany        = repository.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
	ErrUnknownForCode         = repository.ErrUnknownForCode
)

const codeTplId = "1877556"

type CodeService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error)
}

type codeService struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &codeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

// Send 发验证码，需要什么参数
func (svc *codeService) Send(ctx context.Context,
	// 区别业务场景
	biz string,
	phone string) error {
	//生成一个验证码
	code := svc.generateCode()
	// 塞进 Redis,单线程的，解决了并发问题
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 到这里算验证码的业务流程都通过了

	//可以发送出去
	err = svc.smsSvc.Send(ctx, codeTplId, []string{code}, phone)
	//if err != nil {
	//	这个地方怎么办？
	//	这意味着，Redis 有这个验证码，但是不好意思，发送失败
	//	我能不能删掉这个验证码？
	//	不能，你这个 err 可能是超时的 err，你都不知道，发出了没
	//	在这里重试？
	//	要重试的话，初始化的时候，传入一个自己就会重试的 smsSvc
	//}
	return err
}

func (svc *codeService) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *codeService) generateCode() string {
	// 六位数，num 在0，999999 之间，左右包含
	num := rand.Intn(1000000)
	//不够六位的，加上前导0：000001
	return fmt.Sprintf("%06d", num)
}

//func (svc *codeService) VerifyV1(ctx context.Context, phone string, biz string, intputCode string) error {}
