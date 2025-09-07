//go:build wireinject

package startup

import (
	"github.com/XD/ScholarNet/cmd/account/grpc"
	"github.com/XD/ScholarNet/cmd/account/repository"
	"github.com/XD/ScholarNet/cmd/account/repository/dao"
	"github.com/XD/ScholarNet/cmd/account/service"
	"github.com/google/wire"
)

func InitAccountService() *grpc.AccountServiceServer {
	wire.Build(InitTestDB,
		dao.NewCreditGORMDAO,
		repository.NewAccountRepository,
		service.NewAccountService,
		grpc.NewAccountServiceServer)
	return new(grpc.AccountServiceServer)
}
