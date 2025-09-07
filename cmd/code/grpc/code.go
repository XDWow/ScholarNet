package grpc

import (
	"context"
	codev1 "github.com/XD/ScholarNet/cmd/api/proto/gen/code/v1"
	"github.com/XD/ScholarNet/cmd/code/service"
	"google.golang.org/grpc"
)

type CodeServiceServer struct {
	codev1.UnimplementedCodeServiceServer
	service service.CodeService
}

func NewCodeServiceServer(svc service.CodeService) *CodeServiceServer {
	return &CodeServiceServer{
		service: svc,
	}
}

func (c *CodeServiceServer) Register(server grpc.ServiceRegistrar) {
	codev1.RegisterCodeServiceServer(server, c)
}

func (c *CodeServiceServer) Send(ctx context.Context, req *codev1.CodeSendRequest) (*codev1.CodeSendResponse, error) {
	err := c.service.Send(ctx, req.Biz, req.Phone)
	return &codev1.CodeSendResponse{}, err
}

func (c *CodeServiceServer) Verify(ctx context.Context, req *codev1.VerifyRequest) (*codev1.VerifyResponse, error) {
	ans, err := c.service.Verify(ctx, req.Biz, req.Phone, req.InputCode)
	return &codev1.VerifyResponse{
		Answer: ans,
	}, err
}
