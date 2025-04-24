package grpc

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	// cc 是一个连接池的池子，就是 cc 里面放了很多个连接池，一个 IP+端口 一个连接池
	cc, err := grpc.Dial("localhost:8090",
		grpc.WithChainUnaryInterceptor(clientFirst, clientSecond),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 那我怎么知道是 vip？？
	// shadow, ab test
	ctx = context.WithValue(ctx, "label", "vip")
	resp, err := client.GetById(ctx, &GetByIdRequest{
		Id: 456,
	})
	assert.NoError(t, err)
	t.Log(resp.User)
}

var clientFirst grpc.UnaryClientInterceptor = func(ctx context.Context,
	method string, req, reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	log.Println("客户端第一个前")
	err := invoker(ctx, method, req, reply, cc, opts...)
	log.Println("客户端第一个后")
	return err
}

var clientSecond grpc.UnaryClientInterceptor = func(ctx context.Context,
	method string, req, reply any,
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {
	log.Println("客户端第二个前")
	err := invoker(ctx, method, req, reply, cc, opts...)
	log.Println("客户端第二个后")
	return err
}
