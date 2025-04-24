package grpc

import (
	"context"
	_ "github.com/LXD-c/basic-go/webook/pkg/grpcx/balancer/wrr"
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

type EtcdTestSuite struct {
	suite.Suite
	client *etcdv3.Client
}

func (s *EtcdTestSuite) SetupSuite() {
	client, err := etcdv3.New(etcdv3.Config{
		Endpoints: []string{"localhost:12379"},
	})
	require.NoError(s.T(), err)
	s.client = client
}

func (s *EtcdTestSuite) TestCustomRoundRobinClient() {
	bd, err := resolver.NewBuilder(s.client)
	require.NoError(s.T(), err)
	// URL 的规范 scheme:///xxxxx
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(bd),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"custom_wrr"}`),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	client := NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		resp, err := client.GetById(ctx, &GetByIdRequest{
			Id: 123,
		})
		require.NoError(s.T(), err)
		s.T().Log(resp.User)
	}
}

func (s *EtcdTestSuite) TestWeightedRoundRobinClient() {
	bd, err := resolver.NewBuilder(s.client)
	require.NoError(s.T(), err)
	// URL 的规范 scheme:///xxxxx
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(bd),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"weighted_round_robin"}`),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	client := NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		resp, err := client.GetById(ctx, &GetByIdRequest{
			Id: 123,
		})
		require.NoError(s.T(), err)
		s.T().Log(resp.User)
	}
}

func (s *EtcdTestSuite) TestRoundRobinClient() {
	bd, err := resolver.NewBuilder(s.client)
	require.NoError(s.T(), err)
	// URL 的规范 scheme:///xxxxx
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(bd),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	client := NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		resp, err := client.GetById(ctx, &GetByIdRequest{
			Id: 123,
		})
		require.NoError(s.T(), err)
		s.T().Log(resp)
	}
}

func (s *EtcdTestSuite) TestClient() {
	bd, err := resolver.NewBuilder(s.client)
	require.NoError(s.T(), err)
	// URL 的规范 scheme:///xxxxx
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(bd),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	client := NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := client.GetById(ctx, &GetByIdRequest{
		Id: 123,
	})
	require.NoError(s.T(), err)
	s.T().Log(resp)
	time.Sleep(time.Minute)
}

func (s *EtcdTestSuite) TestServer() {
	go func() {
		s.startServer(":8090", 20)
	}()

	s.startServer(":8091", 10)
}

func (s *EtcdTestSuite) startServer(addr string, weight int) {
	l, err := net.Listen("tcp", addr)
	require.NoError(s.T(), err)

	// endpoint 以服务为维度。一个服务一个 Manager
	em, err := endpoints.NewManager(s.client, "service/user")
	require.NoError(s.T(), err)
	addr = netx.GetOutboundIP() + addr
	// key 是指这个实例的 key
	// 如果有 instance id，用 instance id，如果没有，本机 IP + 端口
	// 端口一般是从配置文件里面读
	key := "service/user/" + addr
	//... 在这一步之前完成所有的启动的准备工作，包括缓存预加载之类的事情,因为要注册到etcd了

	// 这个 ctx 是控制创建租约的超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// ttl 是租期
	// 秒作为单位
	// 过了 1/3（还剩下 2/3 的时候）就续约
	var ttl int64 = 30
	leaseResp, err := s.client.Grant(ctx, ttl)
	cancel()
	require.NoError(s.T(), err)

	// 注册到 etcd
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 带上第三个就代表我是租房,后需修改的 Add 也要带上租约，不然不认
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr: addr,
		Metadata: map[string]any{
			"weight": weight,
			//"cpu":	90,
		},
	}, etcdv3.WithLease(leaseResp.ID))
	require.NoError(s.T(), err)

	// 续约
	// 为什么要搞租+续约，不直接长期，为了应对服务器突然宕机
	// 不租，长期的话：服务器突然宕机，etcd会一直有注册信息，客户端会一直发请求，没人理
	// 租+续约：服务器宕机，退出续约，到期自动删除注册信息
	// 如何退出续约？通过 context
	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		_, err1 := s.client.KeepAlive(kaCtx, leaseResp.ID)
		require.NoError(s.T(), err1)
		//for kaResp := range ch {
		//	// 正常就是打印一下 DEBUG 日志啥的
		//	s.T().Log(kaResp.String(), time.Now().String())
		//}
	}()

	//go func() {
	//	ticker := time.NewTicker(time.Second)
	//	// 万一，我的注册信息有变动，怎么办？
	//	for now := range ticker.C {
	//		ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
	//		// AddEndpoint 是一个覆盖的语义。也就是说，如果你这边已经有这个 key 了，就覆盖
	//		// upsert，set
	//		err = em.AddEndpoint(ctx1, key, endpoints.Endpoint{
	//			Addr: addr,
	//			// 你们的分组信息，权重信息，机房信息
	//			// 以及动态判定负载的时候，可以把你的负载信息也写到这里
	//			Metadata: now.String(),
	//		}, etcdv3.WithLease(leaseResp.ID))
	//		if err != nil {
	//			s.T().Log(err)
	//		}
	//		cancel1()
	//	}
	//}()

	// 注册好之后，也就是初始化结束，启动 grpc 服务
	server := grpc.NewServer()
	// 传入业务的 server
	RegisterUserServiceServer(server, &Server{
		// 用地址来标识
		Name: addr,
	})
	err = server.Serve(l)

	// 正常退出
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 先取消续约
	kaCancel()
	// 再从注册中心里面删了自己
	err = em.DeleteEndpoint(ctx, key)
	// 关闭客户端，每个服务和客户端都会有个etcd.Client，通过其于etcd服务器交互
	s.client.Close()
	// 服务优雅退出：还未执行完的执行完，不再接受新请求
	server.GracefulStop()
}

func TestEtcd(t *testing.T) {
	suite.Run(t, new(EtcdTestSuite))
}
