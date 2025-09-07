//go:build wireinject

package startup

import (
	grpc2 "github.com/XD/ScholarNet/cmd/comment/grpc"
	"github.com/XD/ScholarNet/cmd/comment/repository"
	"github.com/XD/ScholarNet/cmd/comment/repository/dao"
	"github.com/XD/ScholarNet/cmd/comment/service"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"github.com/google/wire"
)

var serviceProviderSet = wire.NewSet(
	dao.NewCommentDAO,
	repository.NewCommentRepo,
	service.NewCommentSvc,
	grpc2.NewGrpcServer,
)

var thirdProvider = wire.NewSet(
	logger.NewNoOpLogger,
	InitTestDB,
)

func InitGRPCServer() *grpc2.CommentServiceServer {
	wire.Build(thirdProvider, serviceProviderSet)
	return new(grpc2.CommentServiceServer)
}
