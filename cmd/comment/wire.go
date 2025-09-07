//go:build wireinject

package main

import (
	grpc2 "github.com/XD/ScholarNet/cmd/comment/grpc"
	"github.com/XD/ScholarNet/cmd/comment/ioc"
	"github.com/XD/ScholarNet/cmd/comment/repository"
	"github.com/XD/ScholarNet/cmd/comment/repository/dao"
	"github.com/XD/ScholarNet/cmd/comment/service"
	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(
	ioc.InitLogger,
	ioc.InitDB,
)

var serviceProviderSet = wire.NewSet(
	dao.NewCommentDAO,
	repository.NewCommentRepo,
	service.NewCommentSvc,
	grpc2.NewGrpcServer,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		serviceProviderSet,
		ioc.InitGRPCxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
