//go:build wireinject

package main

import (
	"github.com/XD/ScholarNet/cmd/pkg/wego"
	"github.com/XD/ScholarNet/cmd/sms/grpc"
	"github.com/XD/ScholarNet/cmd/sms/ioc"
	"github.com/google/wire"
)

func Init() *wego.App {
	wire.Build(
		ioc.InitLogger,
		ioc.InitEtcdClient,
		ioc.InitSmsTencentService,
		grpc.NewSmsServiceServer,
		ioc.InitGRPCxServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
