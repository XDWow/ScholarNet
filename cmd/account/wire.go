//go:build wireinject

package main

import (
	"github.com/XD/ScholarNet/cmd/account/grpc"
	"github.com/XD/ScholarNet/cmd/account/ioc"
	"github.com/XD/ScholarNet/cmd/account/repository"
	"github.com/XD/ScholarNet/cmd/account/repository/dao"
	"github.com/XD/ScholarNet/cmd/account/service"
	"github.com/XD/ScholarNet/cmd/pkg/wego"
	"github.com/google/wire"
)

func Init() *wego.App {
	wire.Build(
		ioc.InitDB,
		ioc.InitLogger,
		ioc.InitEtcdClient,
		ioc.InitGRPCxServer,
		ioc.Initjob,
		dao.NewCreditGORMDAO,
		repository.NewAccountRepository,
		service.NewAccountService,
		grpc.NewAccountServiceServer,
		wire.Struct(new(wego.App), "GRPCServer", ""))
	return new(wego.App)
}
