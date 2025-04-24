package grpc

import (
	"context"
	_ "embed"
	"github.com/LXD-c/basic-go/webook/pkg/netx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

type FailoverTestSuite struct {
	suite.Suite
	client *etcdv3.Client
}

func (s *FailoverTestSuite) SetupSuite() {
	client, err := etcdv3.New(etcdv3.Config{
		Endpoints: []string{"localhost:12379"},
	})
	require.NoError(s.T(), err)
	s.client = client
}

func (s *FailoverTestSuite) TestAllInOrder() {
	go s.TestServer()
	time.Sleep(3 * time.Second) // 或 etcd 检查服务注册状态
	s.TestRoundRobinClient()
}

//go:embed failover.json
var svcCfg string

func (s *FailoverTestSuite) TestRoundRobinClient() {
	bd, err := resolver.NewBuilder(s.client)
	require.NoError(s.T(), err)
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(bd),
		// 负载均衡+重试实现 failover
		// 在这里使用负载均衡，并设置重试
		grpc.WithDefaultServiceConfig(svcCfg),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	client := NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetById(ctx, &GetByIdRequest{
			Id: 123,
		})
		cancel()
		require.NoError(s.T(), err)
		s.T().Log(resp.User)
	}
}

func (s *FailoverTestSuite) TestServer() {
	go func() {
		s.startServer(":8091", &AlwaysFailedServer{
			Name: "failed",
		})
	}()
	go func() {
		s.startServer(":8092", &Server{
			Name: "normal",
		})
	}()
	s.startServer(":8090", &Server{
		Name: "normal",
	})
}

func (s *FailoverTestSuite) startServer(addr string, svc UserServiceServer) {
	l, err := net.Listen("tcp", addr)
	require.NoError(s.T(), err)

	// endpoint 以服务为维度。一个服务一个 Manager
	em, err := endpoints.NewManager(s.client, "service/user")
	require.NoError(s.T(), err)
	addr = netx.GetOutboundIP() + addr
	key := "service/user/" + addr
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	var ttl int64 = 30
	leaseResp, err := s.client.Grant(ctx, ttl)
	require.NoError(s.T(), err)

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr:     addr,
		Metadata: map[string]any{},
	}, etcdv3.WithLease(leaseResp.ID))
	require.NoError(s.T(), err)

	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		// 在这里操作续约
		_, err1 := s.client.KeepAlive(kaCtx, leaseResp.ID)
		require.NoError(s.T(), err1)
		//for kaResp := range ch {
		//	// 正常就是打印一下 DEBUG 日志啥的
		//	s.T().Log(kaResp.String(), time.Now().String())
		//}
	}()

	server := grpc.NewServer()
	// 要注册那个健康检查的服务

	RegisterUserServiceServer(server, svc)
	//health_v1.Regist health.Server{}
	err = server.Serve(l)
	s.T().Log(err)
	// 你要退出了，正常退出
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 我要先取消续约
	kaCancel()
	// 退出阶段，先从注册中心里面删了自己
	err = em.DeleteEndpoint(ctx, key)
	// 关掉客户端
	//s.client.Close()
	server.GracefulStop()
}

func TestFailover(t *testing.T) {
	suite.Run(t, new(FailoverTestSuite))
}
