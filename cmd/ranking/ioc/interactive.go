package ioc

import (
	intrv1 "github.com/XD/ScholarNet/cmd/api/proto/gen/intr/v1"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func InitInterActiveRpcClient() intrv1.InteractiveServiceClient {
	type config struct {
		Addr string `yaml:"addr"`
	}
	var cfg config
	err := viper.UnmarshalKey("grpc.client.intr", &cfg)
	if err != nil {
		panic(err)
	}
	conn, err := grpc.Dial(cfg.Addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := intrv1.NewInteractiveServiceClient(conn)
	return client
}
