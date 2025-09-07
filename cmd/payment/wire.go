//go:build wireinject

package main

import (
	"github.com/XD/ScholarNet/cmd/payment/grpc"
	"github.com/XD/ScholarNet/cmd/payment/ioc"
	"github.com/XD/ScholarNet/cmd/payment/repository"
	"github.com/XD/ScholarNet/cmd/payment/repository/dao"
	"github.com/XD/ScholarNet/cmd/payment/web"
	"github.com/XD/ScholarNet/cmd/pkg/wego"
	"github.com/google/wire"
)

func InitApp() *wego.App {
	wire.Build(
		ioc.InitEtcdClient,
		ioc.InitKafka,
		ioc.InitProducer,
		ioc.InitWechatClient,
		dao.NewPaymentGORMDAO,
		ioc.InitDB,
		repository.NewPaymentRepository,
		grpc.NewWechatServiceServer,
		ioc.InitWechatNativeService,
		ioc.InitWechatConfig,
		ioc.InitWechatNotifyHandler,
		ioc.InitGRPCServer,
		web.NewWechatHandler,
		ioc.InitGinServer,
		ioc.InitLogger,
		wire.Struct(new(wego.App), "WebServer", "GRPCServer"))
	return new(wego.App)
}
